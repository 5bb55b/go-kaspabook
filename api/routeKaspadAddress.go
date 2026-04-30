////////////////////////////////
package api

import (
    "github.com/gofiber/fiber/v3"
    "kaspabook/proto/protowire"
)

////////////////////////////////
type responseAddressBalanceType struct {
    Message string `json:"message"`
    Result *protowire.RpcBalancesByAddressesEntry `json:"result"`
}

type responseAddressUtxosType struct {
    Message string `json:"message"`
    Result []*protowire.RpcUtxosByAddressesEntry `json:"result"`
}

////////////////////////////////
func routeKaspadAddressBalance(c fiber.Ctx) (error) {
    r := &responseAddressBalanceType{}
    address, err := filterAddress(c.Params("address"))
    if err != nil {
        r.Message = "address invalid"
        return c.Status(400).JSON(r)
    }
    balanceData, err := grpcKaspa.GetBalancesByAddresses([]string{address})
    if err != nil {
        r.Message = msgKaspadError
        return c.Status(503).JSON(r)
    }
    if balanceData == nil || len(balanceData.Entries) != 1 {
        r.Message = msgSuccessful
        return c.JSON(r)
    }
    r.Result = balanceData.Entries[0]
    r.Message = msgSuccessful
    return c.JSON(r)
}

////////////////////////////////
func routeKaspadAddressUtxos(c fiber.Ctx) (error) {
    r := &responseAddressUtxosType{}
    address, err := filterAddress(c.Params("address"))
    if err != nil {
        r.Message = "address invalid"
        return c.Status(400).JSON(r)
    }
    utxoData, err := grpcKaspa.GetUtxosByAddresses([]string{address})
    if err != nil {
        r.Message = msgKaspadError
        return c.Status(503).JSON(r)
    }
    if utxoData == nil {
        r.Message = msgSuccessful
        return c.JSON(r)
    }
    r.Result = utxoData.Entries
    r.Message = msgSuccessful
    return c.JSON(r)
}
