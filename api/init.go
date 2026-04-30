////////////////////////////////
package api

import (
    "os"
    "time"
    "sync"
    "strconv"
    "log/slog"
    jsoniter "github.com/json-iterator/go"
    "github.com/gofiber/fiber/v3"
    "github.com/gofiber/fiber/v3/middleware/limiter"
    "github.com/gofiber/fiber/v3/middleware/timeout"
    "github.com/gofiber/fiber/v3/middleware/recover"
    "kaspabook/config"
    "kaspabook/kaspa"
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
    msgFailed = "failed"
    msgInternalError = "internal error"
    msgKaspadError = "kaspad error"
    msgDataExpired = "data expired"
    msgNotReached = "not reached"
)

////////////////////////////////
var grpcKaspa *kaspa.GrpcConnectionType
var server *fiber.App

////////////////////////////////
func Init(c chan os.Signal) {
    var err error
    slog.Info("api server starting.", "host", config.Api.Host, "port", config.Api.Port)
    grpcKaspa, err = kaspa.GrpcNewConnection(32*1024*1024)
    if err != nil {
        slog.Error("kaspa.GrpcConnection.connect failed.", "error", err.Error())
        c <- os.Interrupt
        return
    }
    server = fiber.New()
    server.Use(limiter.New(limiter.Config{ Max: config.Api.ConnMax }))
    server.Use(timeout.New(func(c fiber.Ctx) error { return c.Next() }, timeout.Config{ Timeout: time.Duration(config.Api.Timeout)*time.Second }))
    server.Use(recover.New(recover.Config{ EnableStackTrace: false }))
    server.Use(func(c fiber.Ctx) error {
        c.Set("Access-Control-Allow-Origin", "*")
        c.Set("Access-Control-Allow-Methods", "GET")
        c.Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
        c.Set("Access-Control-Max-Age", "1")
        c.Set("X-Content-Type-Options", "nosniff")
        status, err := getBookStatus()
        if err != nil {
            return c.SendStatus(500)
        }
        if c.Method() == "GET" && c.Path() == "/book/status" {
            r := &responseStatusType{ Result: status }
            if status.StatusKaspad != "synced" {
                r.Message = msgUnsynced
                return c.Status(503).JSON(r)
            }
            r.Message = msgSynced
            return c.JSON(r)
        } else if status.StatusKaspad != "synced" {
            r := &responseStatusType{ Message: msgUnsynced }
            return c.Status(503).JSON(r)
        }
        return c.Next()
    })
    server.Get("/book/blocks/:hash", routeBookBlock)
    server.Get("/book/transactions/:txid", routeBookTransaction)
    server.Get("/book/vspcs/daascore/:score", routeBookVspcsByDaaScore)
    server.Get("/book/vspcs/bluescore/:score", routeBookVspcsByBlueScore)
    server.Get("/book/addresses/:address/transactions", routeBookAddressTransactions)
    // ...
    server.Get("/kaspad/addresses/:address/balance", routeKaspadAddressBalance)
    server.Get("/kaspad/addresses/:address/utxos", routeKaspadAddressUtxos)
    server.Post("/kaspad/transactions", routeKaspadTransactionSubmit)
    server.Post("/kaspad/transactions/rbf", routeKaspadTransactionSubmitRbf)
    // ...
    server.All("*", func(c fiber.Ctx) (error) {
        return c.SendStatus(404)
    })
    go func() {
        err := server.Listen(config.Api.Host + ":" + strconv.Itoa(config.Api.Port), fiber.ListenConfig{ DisableStartupMessage: true })
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
    grpcKaspa.Close()
}
