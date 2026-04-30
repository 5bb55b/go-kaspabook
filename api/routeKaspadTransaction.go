////////////////////////////////
package api

import (
    "github.com/gofiber/fiber/v3"
    "kaspabook/proto/protowire"
)

////////////////////////////////
type responseTransactionSubmitType struct {
    Message string `json:"message"`
    Result *protowire.SubmitTransactionResponseMessage `json:"result"`
}

type responseTransactionSubmitRbfType struct {
    Message string `json:"message"`
    Result *protowire.SubmitTransactionReplacementResponseMessage `json:"result"`
}

////////////////////////////////
func routeKaspadTransactionSubmit(c fiber.Ctx) (error) {
    r := &responseTransactionSubmitType{}
    var txData protowire.RpcTransaction
    err := c.Bind().Body(&txData)
    if err != nil {
        r.Message = "data invalid"
        return c.Status(400).JSON(r)
    }
    response, err := grpcKaspa.SubmitTransaction(&txData)
    if err != nil {
        r.Message = msgKaspadError
        return c.Status(503).JSON(r)
    }
    r.Result = response
    if response.Error != nil {
        r.Message = msgFailed
    } else {
        r.Message = msgSuccessful
    }
    return c.JSON(r)
}

////////////////////////////////
func routeKaspadTransactionSubmitRbf(c fiber.Ctx) (error) {
    r := &responseTransactionSubmitRbfType{}
    var txData protowire.RpcTransaction
    err := c.Bind().Body(&txData)
    if err != nil {
        r.Message = "data invalid"
        return c.Status(400).JSON(r)
    }
    response, err := grpcKaspa.SubmitTransactionReplacement(&txData)
    if err != nil {
        r.Message = msgKaspadError
        return c.Status(503).JSON(r)
    }
    r.Result = response
    if response.Error != nil {
        r.Message = msgFailed
    } else {
        r.Message = msgSuccessful
    }
    return c.JSON(r)
}
