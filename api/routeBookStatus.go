////////////////////////////////
package api

import (
    "time"
    "kaspabook/database"
)

////////////////////////////////
type responseStatusType struct {
    Message string `json:"message"`
    Result *database.DbRuntimeStatusType `json:"result"`
}

////////////////////////////////
const cacheTimeoutInfo = 1000

////////////////////////////////
var dataStatus database.DbRuntimeStatusType
var cacheStateInfo cacheStateType

////////////////////////////////
func getBookStatus() (*database.DbRuntimeStatusType, error) {
    status := &database.DbRuntimeStatusType{}
    mtsNow := time.Now().UnixMilli()
    cacheAvailable := false
    cacheStateInfo.RLock()
    if mtsNow - cacheStateInfo.mtsUpdate <= cacheTimeoutInfo {
        *status = dataStatus
        cacheAvailable = true
    }
    cacheStateInfo.RUnlock()
    if cacheAvailable {
        return status, nil
    }
    var err error
    status, err = database.GetRuntimeStatus()
    if err != nil {
        return nil, err
    }
    cacheStateInfo.Lock()
    defer cacheStateInfo.Unlock()
    dataStatus = *status
    cacheStateInfo.mtsUpdate = time.Now().UnixMilli()
    return status, nil
}
