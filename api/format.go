////////////////////////////////
package api

import (
    "strconv"
    "encoding/hex"
    "kaspabook/proto/protobook"
    "kaspabook/misc"
)

////////////////////////////////
type formatBlockType struct {
	Hash string `json:"hash"`
	DaaScore string `json:"daaScore"`
	BlueScore string `json:"blueScore"`
	Timestamp string `json:"timestamp"`
	AcceptedIdMerkleRoot string `json:"acceptedIdMerkleRoot"`
    IsChainBlock string `json:"isChainBlock"`
}

type formatTransactionInputType struct {
	PrevTxId string `json:"prevTxId"`
	PrevTxIndex string `json:"prevTxIndex"`
	Address string `json:"address"`
	Amount string `json:"amount"`
    Spk string `json:"spk"`
    SpkType string `json:"spkType"`
}

type formatTransactionOutputType struct {
	Address string `json:"address"`
	Amount string `json:"amount"`
    Spk string `json:"spk"`
    SpkType string `json:"spkType"`
}

type formatTransactionType struct {
	TxId string `json:"txId"`
	TxHash string `json:"txHash"`
	Inputs []*formatTransactionInputType `json:"inputs"`
	Outputs []*formatTransactionOutputType `json:"outputs"`
    Fee string `json:"fee"`
	BlockHash string `json:"blockHash"`
	BlockTime string `json:"blockTime"`
	AcceptedBlock string `json:"acceptedBlock"`
	AcceptedDaaScore string `json:"acceptedDaaScore"`
	AcceptedBlueScore string `json:"acceptedBlueScore"`
	AcceptedTime string `json:"acceptedTime"`
    IsAccepted string `json:"isAccepted"`
}

type formatVspcType struct {
    Block *formatBlockType `json:"block"`
    Transactions []*formatTransactionType `json:"transactions"`
}

////////////////////////////////
func formatBookVspcList(vspcDataList []*protobook.Vspc, blockDataMap map[string]*protobook.Block, txDataMap map[string]*protobook.Transaction, prev bool) ([]*formatVspcType) {
    result := make([]*formatVspcType, len(vspcDataList))
    for i, vspc := range vspcDataList {
        block := blockDataMap[string(vspc.Hash)]
        if block == nil {
            continue
        }
        result[i] = &formatVspcType{
            Block: formatBookBlock(block),
            Transactions: make([]*formatTransactionType, len(vspc.TxIds)),
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
func formatBookBlock(blockData *protobook.Block) (*formatBlockType) {
    
    // if nil ...
    
    result := &formatBlockType{
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
func formatBookTransaction(txData *protobook.Transaction, blockData *protobook.Block) (*formatTransactionType) {
    
    // if nil ...
    
    result := &formatTransactionType{
        TxId: hex.EncodeToString(txData.TxId),
        TxHash: hex.EncodeToString(txData.TxHash),
        BlockHash: hex.EncodeToString(txData.BlockHash),
        BlockTime: strconv.FormatUint(txData.BlockTime,10),
        AcceptedBlock: hex.EncodeToString(blockData.Hash),
        AcceptedDaaScore: strconv.FormatUint(blockData.DaaScore,10),
        AcceptedBlueScore: strconv.FormatUint(blockData.BlueScore,10),
        AcceptedTime: strconv.FormatUint(blockData.Timestamp,10),
        IsAccepted: "true",
        Fee: "0",
    }
    amountIn := uint64(0)
    amountOut := uint64(0)
    result.Inputs = make([]*formatTransactionInputType, len(txData.Inputs))
    for i, input := range txData.Inputs {
        result.Inputs[i] = &formatTransactionInputType{
            PrevTxId: hex.EncodeToString(input.PrevTxId),
            PrevTxIndex: strconv.FormatUint(uint64(input.PrevTxIndex),10),
            Address: input.Address,
            Amount: strconv.FormatUint(input.Amount,10),
        }
        amountIn += input.Amount
        result.Inputs[i].Spk, result.Inputs[i].SpkType, _, _ = misc.ConvAddressToSpk(input.Address)
    }
    result.Outputs = make([]*formatTransactionOutputType, len(txData.Outputs))
    for i, output := range txData.Outputs {
        result.Outputs[i] = &formatTransactionOutputType{
            Address: output.Address,
            Amount: strconv.FormatUint(output.Amount,10),
        }
        amountOut += output.Amount
        result.Outputs[i].Spk, result.Outputs[i].SpkType, _, _ = misc.ConvAddressToSpk(output.Address)
    }
    if amountIn >= amountOut {
        result.Fee = strconv.FormatUint(amountIn-amountOut, 10)
    }
    return result
}
