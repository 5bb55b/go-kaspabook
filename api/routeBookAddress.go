////////////////////////////////
package api

import (
    "strconv"
    "github.com/gofiber/fiber/v3"
    "kaspabook/database"
)

////////////////////////////////
type responseAddressTransactionsType struct {
    Message string `json:"message"`
    Result []*formatTransactionType `json:"result"`
}

////////////////////////////////
func routeBookAddressTransactions(c fiber.Ctx) (error) {
    r := &responseAddressTransactionsType{}
    address, err := filterAddress(c.Params("address"))
    if err != nil {
        r.Message = "address invalid"
        return c.Status(400).JSON(r)
    }
    count, _ := strconv.Atoi(c.Query("count", "10"))
    prev := c.Query("prev", "")
    rangeBy := "daascore"
    rangeByDaascore := c.Query("daascore")
    rangeByBluescore := c.Query("bluescore")
    //rangeByTimestamp := c.Query("timestamp")
    if rangeByDaascore == "" && rangeByBluescore != "" {
        rangeBy = "bluescore"
    //} else if rangeByDaascore == "" && rangeByTimestamp != "" {
    //    rangeBy = "timestamp"
    }
    rangeStart := uint64(0)
    if prev == "1" {
        rangeStart, err = filterUint(c.Query(rangeBy, "18446744073709551615"))
    } else {
        rangeStart, err = filterUint(c.Query(rangeBy, "0"))
    }
    if err != nil {
        r.Message = "range invalid"
        return c.Status(400).JSON(r)
    }
    getIndexAcceptedTransactionListByAddress := database.GetIndexAcceptedTransactionListByAddressDaaScore
    if rangeBy == "bluescore" {
        getIndexAcceptedTransactionListByAddress = database.GetIndexAcceptedTransactionListByAddressBlueScore
    //} else if rangeBy == "timestamp" {
    //    getIndexAcceptedTransactionListByAddress = database.GetIndexAcceptedTransactionListByAddressTimestamp
    }
    txDataList, daaScoreBeList, blockDataMap, err := getIndexAcceptedTransactionListByAddress(address, rangeStart, count, prev=="1")
    if err != nil {
        r.Message = msgInternalError
        return c.Status(503).JSON(r)
    }
    if len(txDataList) == 0 {
        r.Message = msgSuccessful
        return c.JSON(r)
    }
    r.Result = make([]*formatTransactionType, len(txDataList))
    for i, txData := range txDataList {
        r.Result[i] = formatBookTransaction(txData, blockDataMap[daaScoreBeList[i]])
    }
    r.Message = msgSuccessful
    return c.JSON(r)
}
