////////////////////////////////
package api

import (
    "strconv"
    "github.com/gofiber/fiber/v3"
    "kaspabook/database"
)

////////////////////////////////
type responseBlockType struct {
    Message string `json:"message"`
    Result *formatBlockType `json:"result"`
}

type responseTransactionType struct {
    Message string `json:"message"`
    Result *formatTransactionType `json:"result"`
}

type responseVspcType struct {
    Message string `json:"message"`
    Result []*formatVspcType `json:"result"`
}

////////////////////////////////
func getBookVspcs(c fiber.Ctx, isBlueScore bool) (error) {
    r := &responseVspcType{}
    score, err := filterUint(c.Params("score"))
    if err != nil {
        r.Message = "score invalid"
        return c.Status(400).JSON(r)
    }
    count, _ := strconv.Atoi(c.Query("count", "10"))
    prev := c.Query("prev", "")
    getIndexVspcListByDaaScore := database.GetIndexVspcListByDaaScore
    if isBlueScore {
        getIndexVspcListByDaaScore = database.GetIndexVspcListByBlueScore
    }
    vspcList, blockDataMap, txDataMap, err := getIndexVspcListByDaaScore(score, count, prev=="1")
    if err != nil {
        r.Message = msgInternalError
        return c.Status(503).JSON(r)
    }
    if len(vspcList) == 0 {
        r.Message = msgSuccessful
        return c.JSON(r)
    }
    r.Result = formatBookVspcList(vspcList, blockDataMap, txDataMap, prev=="1")
    r.Message = msgSuccessful
    return c.JSON(r)
}

////////////////////////////////
func routeBookVspcsByDaaScore(c fiber.Ctx) (error) {
    return getBookVspcs(c, false)
}

////////////////////////////////
func routeBookVspcsByBlueScore(c fiber.Ctx) (error) {
    return getBookVspcs(c, true)
}

////////////////////////////////
func routeBookBlock(c fiber.Ctx) (error) {
    r := &responseBlockType{}
    hash, _ := filterHash(c.Params("hash"))
    if hash == "" {
        r.Message = "hash invalid"
        return c.Status(400).JSON(r)
    }
    blockData, err := database.GetIndexChainBlock(hash)
    if err != nil {
        r.Message = msgInternalError
        return c.Status(503).JSON(r)
    }
    if blockData == nil {
        r.Message = msgSuccessful
        return c.JSON(r)
    }
    r.Result = formatBookBlock(blockData)
    r.Message = msgSuccessful
    return c.JSON(r)
}

////////////////////////////////
func routeBookTransaction(c fiber.Ctx) (error) {
    r := &responseTransactionType{}
    txId, _ := filterHash(c.Params("txid"))
    if txId == "" {
        r.Message = "txId invalid"
        return c.Status(400).JSON(r)
    }
    txData, blockData, err := database.GetIndexAcceptedTransaction(txId)
    if err != nil {
        r.Message = msgInternalError
        return c.Status(503).JSON(r)
    }
    if txData == nil {
        r.Message = msgSuccessful
        return c.JSON(r)
    }
    r.Result = formatBookTransaction(txData, blockData)
    r.Message = msgSuccessful
    return c.JSON(r)
}
