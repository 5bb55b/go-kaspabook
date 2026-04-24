
////////////////////////////////
package kaspa

import (
    "time"
    "strconv"
    "context"
    "log/slog"
    "kaspabook/config"
    "kaspabook/database"
)
    
////////////////////////////////
var ctx context.Context
var grpcKaspa *GrpcConnectionType
var dataRuntimeStatus *database.DbRuntimeStatusType

////////////////////////////////
func Init(ctxRaw context.Context) (error) {
    var err error
    ctx = ctxRaw
    grpcKaspa, err = GrpcNewConnection()
    if err != nil {
        return err
    }
    dataRuntimeStatus, err = database.GetRuntimeStatus()
    if err != nil {
        return err
    }
    if dataRuntimeStatus.ScannedBook == "" {
        dagInfo, err := grpcKaspa.GetBlockDagInfo()
        if err != nil {
            return err
        }
        dataRuntimeStatus.ScannedBook = dagInfo.PruningPointHash
    }
    dataRuntimeStatus.VersionBook = config.Version
    dataRuntimeStatus.Hysteresis = strconv.Itoa(config.Startup.Hysteresis)
    dataRuntimeStatus.DtlIndex = strconv.FormatUint(config.Rocksdb.DtlIndex, 10)
    initCache()
    return nil
}

////////////////////////////////
func Run() {
    defer func() {
        grpcKaspa.Close()
    }()
    if config.Rocksdb.GcLoop {
        
        // go func - compaction ...
        
    }
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
