////////////////////////////////
package api

import (
    "os"
    "time"
    "sync"
    "strconv"
    "log/slog"
    jsoniter "github.com/json-iterator/go"
    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/limiter"
    "github.com/gofiber/fiber/v2/middleware/timeout"
    "github.com/gofiber/fiber/v2/middleware/recover"
    "kaspabook/config"
)

////////////////////////////////
var json = jsoniter.ConfigCompatibleWithStandardLibrary

////////////////////////////////
type cacheStateType struct {
    sync.RWMutex
    mtsUpdate int64
}

////////////////////////////////
const (
    msgSynced = "synced"
    msgUnsynced = "unsynced"
    msgSuccessful = "successful"
    msgInternalError = "internal error"
    msgDataExpired = "data expired"
    msgNotReached = "not reached"
)

////////////////////////////////
var server *fiber.App

////////////////////////////////
func Init(c chan os.Signal) {
    slog.Info("api server starting.", "host", config.Api.Host, "port", config.Api.Port)
    server = fiber.New(fiber.Config{DisableStartupMessage:true})
    server.Use(limiter.New(limiter.Config{ Max: config.Api.ConnMax }))
    server.Use(timeout.NewWithContext(func(c *fiber.Ctx) error { return c.Next() }, time.Duration(config.Api.Timeout)*time.Second))
    server.Use(recover.New())
    server.Use(func(c *fiber.Ctx) error {
        c.Set("Access-Control-Allow-Origin", "*")
        c.Set("Access-Control-Allow-Methods", "GET")
        c.Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
        c.Set("Access-Control-Max-Age", "1")
        c.Set("X-Content-Type-Options", "nosniff")
        _, err := getBookStatus()
        if err != nil {
            return c.SendStatus(500)
        }
        return c.Next()
    })
    server.Get("/book/status", routeBookStatus)
    server.Get("/book/block/:hash", routeBookBlock)
    server.Get("/book/transaction/:txid", routeBookTransaction)
    server.Get("/book/vspc/daascore/:score", routeBookVspcDaaScore)
    server.Get("/book/vspc/bluescore/:score", routeBookVspcBlueScore)
/*    server.Get("/book/address/:address/txns", routeBookAddressTransaction)
    // ...
    server.Get("/kaspad/address/:address/balance", routeKaspadAddressBalance)
    server.Get("/kaspad/address/:address/utxos", routeKaspadAddressUtxos)
    server.Post("/kaspad/transaction/broadcast", routeKaspadTransactionBroadcast)
    // ...
*/    server.All("*", func(c *fiber.Ctx) (error) {
        return c.SendStatus(404)
    })
    go func() {
        err := server.Listen(config.Api.Host + ":" + strconv.Itoa(config.Api.Port))
        if err != nil {
            slog.Warn("api server down.", "error", err.Error())
        } else {
            slog.Info("api server down.")
        }
        c <- os.Interrupt
    }()
    time.Sleep(345*time.Millisecond)
}

////////////////////////////////
func Shutdown() {
    if server != nil {
        server.Shutdown()
    }
}
