
////////////////////////////////
package kaspa

import (
    //"fmt"
    "sync"
    "time"
    "context"
    //"strconv"
    "log/slog"
    //"kaspabook/config"
)

////////////////////////////////
    //const
    
////////////////////////////////
var ctx context.Context
var grpcKaspa *GrpcConnectionType

////////////////////////////////
func Init(ctxRaw context.Context, wgRaw *sync.WaitGroup) (error) {
    var err error
    ctx = ctxRaw
    
    // ...
    
    grpcKaspa, err = GrpcNewConnection()
    if err != nil {
        
        // ...
        
    }
    
    // ...
    
    initCache()
    return nil
}

////////////////////////////////
func Run() {
    defer func() {
        grpcKaspa.Close()
        // ...
    }()
    for {
        select {
        case <-ctx.Done():
            slog.Info("kaspa.Scan stopped.")
            return
        default:
            scan()
            time.Sleep(50*time.Millisecond)
        }
    }
}
