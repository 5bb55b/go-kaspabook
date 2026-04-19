
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
    "kaspabook/protokaspa"
)

////////////////////////////////
type GrpcConnectionType struct {
    sync.RWMutex
    conn *grpc.ClientConn
    client protokaspa.RPCClient
    stream protokaspa.RPC_MessageStreamClient
    pending map[uint64]chan *protokaspa.KaspadResponse
    retry int
    index int
    closed chan struct{}
}

////////////////////////////////
func GrpcNewConnection() (*GrpcConnectionType, error) {
    g := &GrpcConnectionType{}
    g.pending = make(map[uint64]chan *protokaspa.KaspadResponse, 256)
    g.closed = make(chan struct{})
    err := g.connect()
    if err != nil {
        return nil, err
    }
    return g, nil
}

////////////////////////////////
func (g *GrpcConnectionType) Close() {
    close(g.closed)
    g.disconnect()
}

////////////////////////////////
func (g *GrpcConnectionType) connect() (error) {
    g.Lock()
    defer g.Unlock()
    if g.conn != nil {
        return nil
    }
    ctxTimeout, cancelTimeout := context.WithTimeout(context.Background(), 13*time.Second)
    defer cancelTimeout()
    g.retry ++
    if g.retry > 55 {
        // ...
    }
    slog.Info("kaspa.connect dialing ..", "index", g.index, "retry", g.retry, "grpc", config.Kaspad.Grpc[g.index])
    var err error
    g.conn, err = grpc.DialContext(
        ctxTimeout, 
        config.Kaspad.Grpc[g.index], 
        grpc.WithTransportCredentials(insecure.NewCredentials()), 
        grpc.WithBlock(), 
        grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(512*1024*1024)), 
        grpc.WithKeepaliveParams(keepalive.ClientParameters{
            Time: 8 * time.Second,
            Timeout: 5 * time.Second,
            PermitWithoutStream: true,
        }),
    )
    if err != nil {
        return err
    }
    g.client = protokaspa.NewRPCClient(g.conn)
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
                slog.Warn("kaspa.connect stream/Recv failed.", "error", err.Error())
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
                slog.Error("kaspa.connect disconnected.", "error", err.Error())
                for {
                    time.Sleep(3 * time.Second)
                    err := g.connect()
                    if err != nil {
                        slog.Error("kaspa.connect failed.", "error", err.Error())
                        continue
                    }
                    break
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
func (g *GrpcConnectionType) request(message *protokaspa.KaspadRequest) (*protokaspa.KaspadResponse, error) {
    id := rand.Uint64()
    message.Id = id
    ch := make(chan *protokaspa.KaspadResponse, 1)
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
func (g *GrpcConnectionType) GetBlock(hash string) (*protokaspa.GetBlockResponseMessage, error) {
    r, err := g.request(&protokaspa.KaspadRequest{Id:0, Payload:&protokaspa.KaspadRequest_GetBlockRequest{GetBlockRequest:&protokaspa.GetBlockRequestMessage{
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

// ...
