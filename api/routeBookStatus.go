////////////////////////////////
package api

import (
    "time"
    "github.com/gofiber/fiber/v2"
    "kaspabook/database"
)

////////////////////////////////
type responseInfoType struct {
    Message string `json:"message"`
    Result *database.DbRuntimeStatusType `json:"result"`
}

////////////////////////////////
const cacheTimeoutInfo = 1000

////////////////////////////////
var dataStatus database.DbRuntimeStatusType
var cacheStateInfo cacheStateType

////////////////////////////////
func routeBookStatus(c *fiber.Ctx) (error) {
    r := &responseInfoType{}
    status, err := getBookStatus()
    if err != nil {
        r.Message = msgInternalError
        return c.Status(503).JSON(r)
    }
    r.Result = status
    if status.StatusKaspad != "synced" {
        r.Message = msgUnsynced
        return c.Status(503).JSON(r)
    }
    r.Message = msgSynced
    return c.JSON(r)
}

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
