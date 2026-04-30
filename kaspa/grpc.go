
////////////////////////////////
package kaspa

import (
    "fmt"
    "sync"
    "time"
    "context"
    "log/slog"
    "math/rand/v2"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    "google.golang.org/grpc/keepalive"
    "kaspabook/config"
    "kaspabook/proto/protowire"
)

////////////////////////////////
type GrpcConnectionType struct {
    sync.RWMutex
    conn *grpc.ClientConn
    client protowire.RPCClient
    stream protowire.RPC_MessageStreamClient
    pending map[uint64]chan *protowire.KaspadResponse
    retry int
    index int
    closed chan struct{}
}

////////////////////////////////
func GrpcNewConnection(maxCallRecvMsgSize int) (*GrpcConnectionType, error) {
    g := &GrpcConnectionType{}
    g.pending = make(map[uint64]chan *protowire.KaspadResponse, 256)
    g.closed = make(chan struct{})
    err := g.connect(maxCallRecvMsgSize)
    if err != nil {
        return nil, err
    }
    return g, nil
}

////////////////////////////////
func (g *GrpcConnectionType) Close() {
    close(g.closed)
    g.disconnect()
    slog.Info("kaspag.GrpcConnection.Close")
}

////////////////////////////////
func (g *GrpcConnectionType) connect(maxCallRecvMsgSize int) (error) {
    g.Lock()
    defer g.Unlock()
    if g.conn != nil {
        return nil
    }
    ctxTimeout, cancelTimeout := context.WithTimeout(context.Background(), 13*time.Second)
    defer cancelTimeout()
    slog.Info("kaspa.GrpcConnection.connect dialing ..", "index", g.index, "retry", g.retry, "grpc", config.Kaspad.Grpc[g.index])
    var err error
    g.conn, err = grpc.DialContext(
        ctxTimeout, 
        config.Kaspad.Grpc[g.index], 
        grpc.WithTransportCredentials(insecure.NewCredentials()), 
        grpc.WithBlock(), 
        grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxCallRecvMsgSize)),
        grpc.WithKeepaliveParams(keepalive.ClientParameters{
            Time: 8 * time.Second,
            Timeout: 5 * time.Second,
            PermitWithoutStream: true,
        }),
    )
    if err != nil {
        return err
    }
    g.client = protowire.NewRPCClient(g.conn)
    g.stream, err = g.client.MessageStream(context.Background())
    if err != nil {
        g.conn.Close()
        g.conn = nil
        g.client = nil
        g.stream = nil
        return err
    }
    go func() {
        for {
            r, err := g.stream.Recv()
            if err != nil {
                slog.Warn("kaspa.GrpcConnection.connect stream/Recv ended.", "reason", err.Error())
                return
            }
            if r.Id == 0 {
                continue
            }
            g.Lock()
            ch, exists := g.pending[r.Id]
            if exists {
                ch <- r
            }
            g.Unlock()
        }
    }()
    go func() {
        for {
            select {
            case <-g.closed:
                return
            case <-g.stream.Context().Done():
                err := g.stream.Context().Err()
                g.disconnect()
                slog.Info("kaspa.GrpcConnection.connect disconnected.", "reason", err.Error())
                for {
                    select {
                    case <-ctx.Done():
                        return
                    default:
                        g.retry ++
                        if g.retry > 55 {
                            slog.Error("kaspa.GrpcConnection.connect failed.", "error", "retries exceeded")
                            return
                        }
                        time.Sleep(3 * time.Second)
                        err := g.connect(maxCallRecvMsgSize)
                        if err != nil {
                            slog.Error("kaspa.GrpcConnection.connect failed.", "error", err.Error())
                            break
                        }
                        return
                    }
                }
                return
            }
        }
    }()
    return nil
}

////////////////////////////////
func (g *GrpcConnectionType) disconnect() {
    g.Lock()
    defer g.Unlock()
    if g.conn == nil {
        return
    }
    g.index ++
    if g.index >= len(config.Kaspad.Grpc) {
        g.index = 0
    }
    for id, ch := range g.pending {
        delete(g.pending, id)
        close(ch)
    }
    if g.stream != nil {
        g.stream.CloseSend()
    }
    g.conn.Close()
    g.conn = nil
    g.client = nil
    g.stream = nil
}

////////////////////////////////
func (g *GrpcConnectionType) request(message *protowire.KaspadRequest) (*protowire.KaspadResponse, error) {
    id := rand.Uint64()
    message.Id = id
    ch := make(chan *protowire.KaspadResponse, 1)
    g.Lock()
    g.pending[id] = ch
    g.Unlock()
    defer func() {
        g.Lock()
        _, exists := g.pending[id]
        if exists {
            delete(g.pending, id)
            close(ch)
        }
        g.Unlock()
    }()
    g.RLock()
    if g.stream == nil {
        g.RUnlock()
        return nil, fmt.Errorf("nil stream")
    }
    err := g.stream.Send(message)
    g.RUnlock()
    if err != nil {
        return nil, err
    }
    ctxTimeout, cancelTimeout := context.WithTimeout(context.Background(), 37*time.Second)
    defer cancelTimeout()
    select {
    case r := <-ch:
        if r == nil {
            return nil, fmt.Errorf("nil response")
        }
        return r, nil
    case <-ctxTimeout.Done():
        return nil, ctxTimeout.Err()
    }
    return nil, nil
}

////////////////////////////////
func (g *GrpcConnectionType) GetInfo() (*protowire.GetInfoResponseMessage, error) {
    r, err := g.request(&protowire.KaspadRequest{Payload:&protowire.KaspadRequest_GetInfoRequest{GetInfoRequest:&protowire.GetInfoRequestMessage{}}})
    if err != nil {
        return nil, err
    }
    response := r.GetGetInfoResponse()
    if response == nil {
        return nil, fmt.Errorf("nil info")
    } else if response.Error != nil && response.Error.Message != "" {
        return nil, fmt.Errorf("%s", response.Error.Message)
    }
    return response, nil
}

////////////////////////////////
func (g *GrpcConnectionType) GetServerInfo() (*protowire.GetServerInfoResponseMessage, error) {
    r, err := g.request(&protowire.KaspadRequest{Payload:&protowire.KaspadRequest_GetServerInfoRequest{GetServerInfoRequest:&protowire.GetServerInfoRequestMessage{}}})
    if err != nil {
        return nil, err
    }
    response := r.GetGetServerInfoResponse()
    if response == nil {
        return nil, fmt.Errorf("nil serverInfo")
    } else if response.Error != nil && response.Error.Message != "" {
        return nil, fmt.Errorf("%s", response.Error.Message)
    }
    return response, nil
}

////////////////////////////////
func (g *GrpcConnectionType) GetBlockDagInfo() (*protowire.GetBlockDagInfoResponseMessage, error) {
    r, err := g.request(&protowire.KaspadRequest{Payload:&protowire.KaspadRequest_GetBlockDagInfoRequest{GetBlockDagInfoRequest:&protowire.GetBlockDagInfoRequestMessage{}}})
    if err != nil {
        return nil, err
    }
    response := r.GetGetBlockDagInfoResponse()
    if response == nil {
        return nil, fmt.Errorf("nil dagInfo")
    } else if response.Error != nil && response.Error.Message != "" {
        return nil, fmt.Errorf("%s", response.Error.Message)
    }
    return response, nil
}

////////////////////////////////
func (g *GrpcConnectionType) GetBlock(hash string) (*protowire.GetBlockResponseMessage, error) {
    r, err := g.request(&protowire.KaspadRequest{Payload:&protowire.KaspadRequest_GetBlockRequest{GetBlockRequest:&protowire.GetBlockRequestMessage{
        Hash: hash,
        IncludeTransactions: false,
    }}})
    if err != nil {
        return nil, err
    }
    response := r.GetGetBlockResponse()
    if response == nil {
        return nil, fmt.Errorf("nil block")
    } else if response.Error != nil && response.Error.Message != "" {
        return nil, fmt.Errorf("%s", response.Error.Message)
    }
    return response, nil
}

////////////////////////////////
func (g *GrpcConnectionType) GetVirtualChainFromBlockV2(startHash string, count uint64) (*protowire.GetVirtualChainFromBlockV2ResponseMessage, error) {
    level := protowire.RpcDataVerbosityLevel_HIGH
    r, err := g.request(&protowire.KaspadRequest{Payload:&protowire.KaspadRequest_GetVirtualChainFromBlockV2Request{GetVirtualChainFromBlockV2Request:&protowire.GetVirtualChainFromBlockV2RequestMessage{
        StartHash: startHash,
        DataVerbosityLevel: &level,
        MinConfirmationCount: &count,
    }}})
    if err != nil {
        return nil, err
    }
    response := r.GetGetVirtualChainFromBlockV2Response()
    if response == nil {
        return nil, fmt.Errorf("nil vspc")
    } else if response.Error != nil && response.Error.Message != "" {
        return nil, fmt.Errorf("%s", response.Error.Message)
    }
    return response, nil
}

////////////////////////////////
func (g *GrpcConnectionType) GetBalancesByAddresses(addressList []string) (*protowire.GetBalancesByAddressesResponseMessage, error) {
    r, err := g.request(&protowire.KaspadRequest{Payload:&protowire.KaspadRequest_GetBalancesByAddressesRequest{GetBalancesByAddressesRequest:&protowire.GetBalancesByAddressesRequestMessage{
        Addresses: addressList,
    }}})
    if err != nil {
        return nil, err
    }
    response := r.GetGetBalancesByAddressesResponse()
    if response == nil {
        return nil, fmt.Errorf("nil balance")
    } else if response.Error != nil && response.Error.Message != "" {
        return nil, fmt.Errorf("%s", response.Error.Message)
    }
    return response, nil
}

////////////////////////////////
func (g *GrpcConnectionType) GetUtxosByAddresses(addressList []string) (*protowire.GetUtxosByAddressesResponseMessage, error) {
    r, err := g.request(&protowire.KaspadRequest{Payload:&protowire.KaspadRequest_GetUtxosByAddressesRequest{GetUtxosByAddressesRequest:&protowire.GetUtxosByAddressesRequestMessage{
        Addresses: addressList,
    }}})
    if err != nil {
        return nil, err
    }
    response := r.GetGetUtxosByAddressesResponse()
    if response == nil {
        return nil, fmt.Errorf("nil utxo")
    } else if response.Error != nil && response.Error.Message != "" {
        return nil, fmt.Errorf("%s", response.Error.Message)
    }
    return response, nil
}

////////////////////////////////
func (g *GrpcConnectionType) SubmitTransaction(txData *protowire.RpcTransaction) (*protowire.SubmitTransactionResponseMessage, error) {
    r, err := g.request(&protowire.KaspadRequest{Payload:&protowire.KaspadRequest_SubmitTransactionRequest{SubmitTransactionRequest:&protowire.SubmitTransactionRequestMessage{
        Transaction: txData,
    }}})
    if err != nil {
        return nil, err
    }
    response := r.GetSubmitTransactionResponse()
    if response == nil {
        return nil, fmt.Errorf("nil response")
    }
    return response, nil
}

////////////////////////////////
func (g *GrpcConnectionType) SubmitTransactionReplacement(txData *protowire.RpcTransaction) (*protowire.SubmitTransactionReplacementResponseMessage, error) {
    r, err := g.request(&protowire.KaspadRequest{Payload:&protowire.KaspadRequest_SubmitTransactionReplacementRequest{SubmitTransactionReplacementRequest:&protowire.SubmitTransactionReplacementRequestMessage{
        Transaction: txData,
    }}})
    if err != nil {
        return nil, err
    }
    response := r.GetSubmitTransactionReplacementResponse()
    if response == nil {
        return nil, fmt.Errorf("nil response")
    }
    return response, nil
}
