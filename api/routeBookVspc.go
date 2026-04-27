////////////////////////////////
package api

import (
    "strconv"
    "encoding/hex"
    "github.com/gofiber/fiber/v2"
    "kaspabook/proto/protobook"
    "kaspabook/database"
    "kaspabook/misc"
)

////////////////////////////////
type resultBlockType struct {
	Hash string `json:"hash"`
	DaaScore string `json:"daaScore"`
	BlueScore string `json:"blueScore"`
	Timestamp string `json:"timestamp"`
	AcceptedIdMerkleRoot string `json:"acceptedIdMerkleRoot"`
    IsChainBlock string `json:"isChainBlock"`
}
type responseBlockType struct {
    Message string `json:"message"`
    Result *resultBlockType `json:"result"`
}

////////////////////////////////
type resultTransactionInputType struct {
	PrevTxId string `json:"prevTxId"`
	PrevTxIndex string `json:"prevTxIndex"`
	Address string `json:"address"`
	Amount string `json:"amount"`
    Spk string `json:"spk"`
    SpkType string `json:"spkType"`
}
type resultTransactionOutputType struct {
	Address string `json:"address"`
	Amount string `json:"amount"`
    Spk string `json:"spk"`
    SpkType string `json:"spkType"`
}
type resultTransactionType struct {
	TxId string `json:"txId"`
	TxHash string `json:"txHash"`
	Inputs []*resultTransactionInputType `json:"inputs"`
	Outputs []*resultTransactionOutputType `json:"outputs"`
    Fee string `json:"fee"`
	BlockHash string `json:"blockHash"`
	BlockTime string `json:"blockTime"`
	AcceptedBlock string `json:"acceptedBlock"`
	AcceptedTime string `json:"acceptedTime"`
    IsAccepted string `json:"isAccepted"`
}
type responseTransactionType struct {
    Message string `json:"message"`
    Result *resultTransactionType `json:"result"`
}

////////////////////////////////
type resultVspcType struct {
    Block *resultBlockType `json:"block"`
    Transactions []*resultTransactionType `json:"transactions"`
}
type responseVspcType struct {
    Message string `json:"message"`
    Result []*resultVspcType `json:"result"`
}

////////////////////////////////
func getBookVspc(c *fiber.Ctx, isBlueScore bool) (error) {
    r := &responseVspcType{}
    status, err := getBookStatus()
    if err != nil {
        r.Message = msgInternalError
        return c.Status(503).JSON(r)
    }
    if status.StatusKaspad != "synced" {
        r.Message = msgUnsynced
        return c.Status(503).JSON(r)
    }
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
func routeBookVspcDaaScore(c *fiber.Ctx) (error) {
    return getBookVspc(c, false)
}

////////////////////////////////
func routeBookVspcBlueScore(c *fiber.Ctx) (error) {
    return getBookVspc(c, true)
}

////////////////////////////////
func routeBookBlock(c *fiber.Ctx) (error) {
    r := &responseBlockType{}
    status, err := getBookStatus()
    if err != nil {
        r.Message = msgInternalError
        return c.Status(503).JSON(r)
    }
    if status.StatusKaspad != "synced" {
        r.Message = msgUnsynced
        return c.Status(503).JSON(r)
    }
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
func routeBookTransaction(c *fiber.Ctx) (error) {
    r := &responseTransactionType{}
    status, err := getBookStatus()
    if err != nil {
        r.Message = msgInternalError
        return c.Status(503).JSON(r)
    }
    if status.StatusKaspad != "synced" {
        r.Message = msgUnsynced
        return c.Status(503).JSON(r)
    }
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

////////////////////////////////
func formatBookVspcList(vspcDataList []*protobook.Vspc, blockDataMap map[string]*protobook.Block, txDataMap map[string]*protobook.Transaction, prev bool) ([]*resultVspcType) {
    result := make([]*resultVspcType, len(vspcDataList))
    for i, vspc := range vspcDataList {
        block := blockDataMap[string(vspc.Hash)]
        if block == nil {
            continue
        }
        result[i] = &resultVspcType{
            Block: formatBookBlock(block),
            Transactions: make([]*resultTransactionType, len(vspc.TxIds)),
        }
        for j, txIdBin := range vspc.TxIds {
            txData := txDataMap[string(txIdBin)]
            if txData == nil {
                continue
            }
            result[i].Transactions[j] = formatBookTransaction(txData, block)
        }
    }
    return result
}

////////////////////////////////
func formatBookBlock(blockData *protobook.Block) (*resultBlockType) {
    result := &resultBlockType{
        Hash: hex.EncodeToString(blockData.Hash),
        DaaScore: strconv.FormatUint(blockData.DaaScore,10),
        BlueScore: strconv.FormatUint(blockData.BlueScore,10),
        Timestamp: strconv.FormatUint(blockData.Timestamp,10),
        AcceptedIdMerkleRoot: hex.EncodeToString(blockData.AcceptedIdMerkleRoot),
        IsChainBlock: "true",
    }
    return result
}

////////////////////////////////
func formatBookTransaction(txData *protobook.Transaction, blockData *protobook.Block) (*resultTransactionType) {
    result := &resultTransactionType{
        TxId: hex.EncodeToString(txData.TxId),
        TxHash: hex.EncodeToString(txData.TxHash),
        BlockHash: hex.EncodeToString(txData.BlockHash),
        BlockTime: strconv.FormatUint(txData.BlockTime,10),
        AcceptedBlock: hex.EncodeToString(blockData.Hash),
        AcceptedTime: strconv.FormatUint(blockData.Timestamp,10),
        IsAccepted: "true",
        Fee: "0",
    }
    amountIn := uint64(0)
    amountOut := uint64(0)
    result.Inputs = make([]*resultTransactionInputType, len(txData.Inputs))
    for i, input := range txData.Inputs {
        result.Inputs[i] = &resultTransactionInputType{
            PrevTxId: hex.EncodeToString(input.PrevTxId),
            PrevTxIndex: strconv.FormatUint(uint64(input.PrevTxIndex),10),
            Address: input.Address,
            Amount: strconv.FormatUint(input.Amount,10),
        }
        amountIn += input.Amount
        result.Inputs[i].Spk, result.Inputs[i].SpkType = misc.ConvAddressToSpk(input.Address)
    }
    result.Outputs = make([]*resultTransactionOutputType, len(txData.Outputs))
    for i, output := range txData.Outputs {
        result.Outputs[i] = &resultTransactionOutputType{
            Address: output.Address,
            Amount: strconv.FormatUint(output.Amount,10),
        }
        amountOut += output.Amount
        result.Outputs[i].Spk, result.Outputs[i].SpkType = misc.ConvAddressToSpk(output.Address)
    }
    if amountIn >= amountOut {
        result.Fee = strconv.FormatUint(amountIn-amountOut, 10)
    }
    return result
}
