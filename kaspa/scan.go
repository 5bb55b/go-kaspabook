
////////////////////////////////
package kaspa

import (
    "time"
    "strconv"
    "log/slog"
    "kaspabook/config"
    "kaspabook/database"
)

////////////////////////////////
var countLoopSynced = int64(1)
var countReorg = int64(0)
var loopScan = 300

////////////////////////////////
func scan() (bool) {
    mtss := time.Now().UnixMilli()
    // Some things to clean up.
    loopScan ++
    if loopScan > 300 {
        loopScan = 0
        database.ExpireIddkeys(dataRuntimeStatus.DaaScoreBookInt)

        // dataRuntimeStatus.SizeBook ...
        
    }
    // Get the synced info.
    serverInfo, err := grpcKaspa.GetServerInfo()
    if err != nil {
        return sleepLog(3000, slog.Warn, "kaspa.GetServerInfo failed, sleep 3s.", "error", err.Error())
    }
    if !serverInfo.IsSynced {
        dataRuntimeStatus.StatusKaspad = "unsynced"
        database.SetRuntimeStatus(nil, dataRuntimeStatus)
        return sleepLog(3000, slog.Warn, "kaspa.GetServerInfo unsynced, sleep 3s.")
    }
    dataRuntimeStatus.VersionKaspad = serverInfo.ServerVersion
    dataRuntimeStatus.DaaScoreKaspad = strconv.FormatUint(serverInfo.VirtualDaaScore, 10)
    dataRuntimeStatus.DaaScoreKaspadInt = serverInfo.VirtualDaaScore
    dataRuntimeStatus.StatusKaspad = "synced"
    dataRuntimeStatus.Network = serverInfo.NetworkId
    // Get the vspc/tx data list.
    vspc, err := grpcKaspa.GetVirtualChainFromBlockV2(dataRuntimeStatus.ScannedBook, uint64(config.Startup.Hysteresis))
    if err != nil {
        return sleepLog(3000, slog.Warn, "kaspa.GetVirtualChainFromBlockV2 failed, sleep 3s.", "error", err.Error())
    }
    lenAdded := len(vspc.AddedChainBlockHashes)
    lenRemoved := len(vspc.RemovedChainBlockHashes)
    if lenAdded == 0 || lenAdded != len(vspc.ChainBlockAcceptedTransactions) {
        return sleepLog(550, slog.Warn, "kaspa.GetVirtualChainFromBlockV2 empty/mismatch, sleep 0.55s.")
    }
    // Process the vspc/tx data list and update the runtime status.
    daaScoreListByRemoved := make([]uint64, 0, lenRemoved)
    for _, hash := range vspc.RemovedChainBlockHashes {
        score, err := getBlockScore(hash)
        if err != nil {
            return sleepLog(3000, slog.Warn, "kaspa.getBlockScore failed, sleep 3s.", "error", err.Error())
        }
        daaScoreListByRemoved = append(daaScoreListByRemoved, score.Daa)
    }
    for _, accepted := range vspc.ChainBlockAcceptedTransactions {
        hash := *accepted.ChainBlockHeader.Hash
        addCacheBlockScore(hash, &blockScoreType{
            Daa: *accepted.ChainBlockHeader.DaaScore,
            Blue: *accepted.ChainBlockHeader.BlueScore,
        })
    }
    dataRuntimeStatus, err = database.ProcessIndexVspc(daaScoreListByRemoved, vspc.ChainBlockAcceptedTransactions, dataRuntimeStatus)
    if err != nil {
        return sleepLog(3000, slog.Warn, "database.ProcessVspc failed, sleep 3s.", "error", err.Error())
    }
    expireCacheBlockScore(dataRuntimeStatus.DaaScoreBookInt)
    // Additional delay if synced.
    mtsLoop := time.Now().UnixMilli() - mtss
    slog.Info("explorer.scan", "lenAdded", lenAdded, "lenRemoved", lenRemoved, "lenTransaction", dataRuntimeStatus.LenTransaction, "mSecondLoop", mtsLoop, "rateReorg", strconv.FormatInt(countReorg*1000/countLoopSynced,10)+"pt", "statusKaspad", dataRuntimeStatus.StatusKaspad)
    gapInt := uint64(0)
    if dataRuntimeStatus.DaaScoreKaspadInt > dataRuntimeStatus.DaaScoreBookInt {
        gapInt = dataRuntimeStatus.DaaScoreKaspadInt - dataRuntimeStatus.DaaScoreBookInt
    }
    if gapInt < 50 {
        countLoopSynced ++
        mtsLoop = 550 - mtsLoop
        if mtsLoop <= 0 {
            return true
        }
        time.Sleep(time.Duration(mtsLoop)*time.Millisecond)
    }
    return true
}

////////////////////////////////
func sleepLog(ms time.Duration, f func(string, ...any), msg string, args ...any) (bool) {
    if ms > 0 {
        time.Sleep(ms * time.Millisecond)
    }
    f(msg, args...)
    return false
}
