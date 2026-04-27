
////////////////////////////////
package database

//#include "rocksdb/c.h"
import "C"
import (
    "sync"
    "strconv"
    "encoding/hex"
    "encoding/binary"
    "google.golang.org/protobuf/proto"
    "kaspabook/config"
    "kaspabook/proto/protowire"
    "kaspabook/proto/protobook"
)

////////////////////////////////
const KeyPrefixVspc = "vspc_"
const KeyPrefixBlue = "blue_"
const KeyPrefixBlock = "block_"
const KeyPrefixBlockSt = "blockst_"
const KeyPrefixTransaction = "txn_"
const KeyPrefixTransactionSt = "txnst_"
const KeyPrefixAddress = "addr_"

////////////////////////////////
const maxCountIndexVspc = 50

////////////////////////////////
func ProcessIndexVspc(daaScoreListByRemoved []uint64, acceptedList []*protowire.RpcChainBlockAcceptedTransactions, status *DbRuntimeStatusType) (*DbRuntimeStatusType, error) {
    // iddkeysToRemove
    lenAdded := len(acceptedList)
    iddkeysToRemove, err := getIddkeysByDaaScoreList(daaScoreListByRemoved)
    if err != nil {
        return status, err
    }
    // loop - accepted
    txRocks := txBegin()
    lenTransactionTotal := 0
    iddkeysToAdd := make(map[string][]string, 256)
    for _, accepted := range acceptedList {
        // index block
        acceptedBlockHash := *accepted.ChainBlockHeader.Hash
        acceptedBlockHashBin, _ := hex.DecodeString(acceptedBlockHash)
        acceptedIdMerkleRootBin, _ := hex.DecodeString(*accepted.ChainBlockHeader.AcceptedIdMerkleRoot)
        block := &protobook.Block{
            Hash: acceptedBlockHashBin,
            DaaScore: *accepted.ChainBlockHeader.DaaScore,
            BlueScore: *accepted.ChainBlockHeader.BlueScore,
            Timestamp: uint64(*accepted.ChainBlockHeader.Timestamp),
            AcceptedIdMerkleRoot: acceptedIdMerkleRootBin,
        }
        keyBlock := append([]byte(KeyPrefixBlock), acceptedBlockHashBin...)
        val, err := proto.Marshal(block)
        if err != nil {
            txRollback(txRocks)
            return status, err
        }
        err = putCF(txRocks, cfIndex, keyBlock, val, block.DaaScore)
        if err != nil {
            txRollback(txRocks)
            return status, err
        }
        // keyIddkeys
        daaScoreBe := make([]byte, 8)
        binary.BigEndian.PutUint64(daaScoreBe, block.DaaScore)
        daaScoreBeString := string(daaScoreBe)
        blueScoreBe := make([]byte, 8)
        binary.BigEndian.PutUint64(blueScoreBe, block.BlueScore)
        keyIddkeys := keyPrefixIddkeys + daaScoreBeString
        iddkeysToAdd[keyIddkeys] = make([]string, 0, 32)
        // index block-st
        keyBlockSt := KeyPrefixBlockSt + string(acceptedBlockHashBin)
        iddkeysToAdd[keyIddkeys] = append(iddkeysToAdd[keyIddkeys], keyBlockSt)
        _, exists := iddkeysToRemove[keyBlockSt]
        if exists {
            delete(iddkeysToRemove, keyBlockSt)
        } else {
            err = putCF(txRocks, cfIndex, []byte(keyBlockSt), nil, block.DaaScore)
            if err != nil {
                txRollback(txRocks)
                return status, err
            }
        }
        // loop - transaction
        lenTransaction := len(accepted.AcceptedTransactions)
        lenTransactionTotal += lenTransaction
        txIdList := make([][]byte, 0, lenTransaction)
        for _, txAccepted := range accepted.AcceptedTransactions {
            // index transaction
            addressMap := make(map[string]struct{})
            txId := *txAccepted.VerboseData.TransactionId
            txIdBin, _ := hex.DecodeString(txId)
            txHashBin, _ := hex.DecodeString(*txAccepted.VerboseData.Hash)
            blockHashBin, _ := hex.DecodeString(*txAccepted.VerboseData.BlockHash)
            txInputs := make([]*protobook.TransactionInput, len(txAccepted.Inputs))
            for i, input := range txAccepted.Inputs {
                prevTxIdBin, _ := hex.DecodeString(*input.PreviousOutpoint.TransactionId)
                address := *input.VerboseData.UtxoEntry.VerboseData.ScriptPublicKeyAddress
                txInputs[i] = &protobook.TransactionInput{
                    PrevTxId: prevTxIdBin,
                    PrevTxIndex: *input.PreviousOutpoint.Index,
                    Address: address,
                    Amount: *input.VerboseData.UtxoEntry.Amount,
                }
                addressMap[address] = struct{}{}
            }
            txOutputs := make([]*protobook.TransactionOutput, len(txAccepted.Outputs))
            for i, output := range txAccepted.Outputs {
                address := *output.VerboseData.ScriptPublicKeyAddress
                txOutputs[i] = &protobook.TransactionOutput{
                    Address: address,
                    Amount: *output.Amount,
                }
                addressMap[address] = struct{}{}
            }
            tx := &protobook.Transaction{
                TxId: txIdBin,
                TxHash: txHashBin,
                BlockHash: blockHashBin,
                BlockTime: *txAccepted.VerboseData.BlockTime,
                Inputs: txInputs,
                Outputs: txOutputs,
            }
            keyTransaction := append([]byte(KeyPrefixTransaction), txIdBin...)
            val, err := proto.Marshal(tx)
            if err != nil {
                txRollback(txRocks)
                return status, err
            }
            err = putCF(txRocks, cfIndex, keyTransaction, val, block.DaaScore)
            if err != nil {
                txRollback(txRocks)
                return status, err
            }
            txIdList = append(txIdList, txIdBin)
            txIdString := string(txIdBin)
            // index transaction-st
            keyTransactionSt := KeyPrefixTransactionSt + txIdString
            iddkeysToAdd[keyIddkeys] = append(iddkeysToAdd[keyIddkeys], keyTransactionSt)
            delete(iddkeysToRemove, keyTransactionSt)
            err = putCF(txRocks, cfIndex, []byte(keyTransactionSt), acceptedBlockHashBin, block.DaaScore)
            if err != nil {
                txRollback(txRocks)
                return status, err
            }
            // index address
            for address := range addressMap {
                keyAddress := KeyPrefixAddress + address + "_" + daaScoreBeString + "_" + txIdString
                iddkeysToAdd[keyIddkeys] = append(iddkeysToAdd[keyIddkeys], keyAddress)
                _, exists := iddkeysToRemove[keyAddress]
                if exists {
                    delete(iddkeysToRemove, keyAddress)
                } else {
                    err = putCF(txRocks, cfIndex, []byte(keyAddress), nil, block.DaaScore)
                    if err != nil {
                        txRollback(txRocks)
                        return status, err
                    }
                }
            }
        }
        // index vspc
        vspc := &protobook.Vspc{
            Hash: acceptedBlockHashBin,
            TxIds: txIdList,
        }
        keyVspc := KeyPrefixVspc + daaScoreBeString
        iddkeysToAdd[keyIddkeys] = append(iddkeysToAdd[keyIddkeys], keyVspc)
        delete(iddkeysToRemove, keyVspc)
        val, err = proto.Marshal(vspc)
        if err != nil {
            txRollback(txRocks)
            return status, err
        }
        err = putCF(txRocks, cfIndex, []byte(keyVspc), val, block.DaaScore)
        if err != nil {
            txRollback(txRocks)
            return status, err
        }
        // index blue
        keyBlue := KeyPrefixBlue + string(blueScoreBe)
        iddkeysToAdd[keyIddkeys] = append(iddkeysToAdd[keyIddkeys], keyBlue)
        delete(iddkeysToRemove, keyBlue)
        err = putCF(txRocks, cfIndex, []byte(keyBlue), daaScoreBe, block.DaaScore)
        if err != nil {
            txRollback(txRocks)
            return status, err
        }
    }
    // iddkeys
    err = delIddkeys(txRocks, iddkeysToRemove)
    if err != nil {
        txRollback(txRocks)
        return status, err
    }
    err = setIddkeys(txRocks, iddkeysToAdd)
    if err != nil {
        txRollback(txRocks)
        return status, err
    }
    // runtime status
    blockLast := acceptedList[lenAdded-1].ChainBlockHeader
    status.LenTransaction = lenTransactionTotal
    status.DaaScoreBookInt = *blockLast.DaaScore
    status.DaaScoreBook = strconv.FormatUint(status.DaaScoreBookInt, 10)
    status.BlueScoreBook = strconv.FormatUint(*blockLast.BlueScore, 10)
    status.ScannedBook = *blockLast.Hash
    gap := uint64(0)
    if status.DaaScoreKaspadInt > status.DaaScoreBookInt {
        gap = status.DaaScoreKaspadInt - status.DaaScoreBookInt
    }
    status.GapBook = strconv.FormatUint(gap, 10)
    err = SetRuntimeStatus(txRocks, status)
    if err != nil {
        txRollback(txRocks)
        return status, err
    }
    // commit
    err = txCommit(txRocks, true)
    if err != nil {
        return status, err
    }
    SetDaaScoreLastRocks(status.DaaScoreBookInt)
    return status, nil
}

////////////////////////////////
func getIndexBlockByHashBin(hashBin []byte) (*protobook.Block, error) {
    var block *protobook.Block
    key := append([]byte(KeyPrefixBlock), hashBin...)
    _, err := getCF(nil, cfIndex, key, func(val []byte) (error) {
        if len(val) == 0 {
            return nil
        }
        block = &protobook.Block{}
        err := proto.Unmarshal(val, block)
        return err
    })
    if err != nil {
        return nil, err
    }
    return block, nil
}

////////////////////////////////
func getIndexTransactionByTxIdBin(txIdBin []byte) (*protobook.Transaction, error) {
    var tx *protobook.Transaction
    key := append([]byte(KeyPrefixTransaction), txIdBin...)
    _, err := getCF(nil, cfIndex, key, func(val []byte) (error) {
        if len(val) == 0 {
            return nil
        }
        tx = &protobook.Transaction{}
        err := proto.Unmarshal(val, tx)
        return err
    })
    if err != nil {
        return nil, err
    }
    return tx, nil
}

////////////////////////////////
func GetIndexChainBlock(hash string) (*protobook.Block, error) {
    hashBin, _ := hex.DecodeString(hash)
    isChainBlock := false
    key := append([]byte(KeyPrefixBlockSt), hashBin...)
    _, err := getCF(nil, cfIndex, key, func(val []byte) (error) {
        isChainBlock = true
        return nil
    })
    if err != nil || !isChainBlock {
        return nil, err
    }
    block, err := getIndexBlockByHashBin(hashBin)
    if err != nil {
        return nil, err
    }
    return block, nil
}

////////////////////////////////
func GetIndexAcceptedTransaction(txId string) (*protobook.Transaction, *protobook.Block, error) {
    txIdBin, _ := hex.DecodeString(txId)
    acceptedBlockBin := ""
    key := append([]byte(KeyPrefixTransactionSt), txIdBin...)
    _, err := getCF(nil, cfIndex, key, func(val []byte) (error) {
        if len(val) == 0 {
            return nil
        }
        acceptedBlockBin = string(val)
        return nil
    })
    if err != nil || acceptedBlockBin == "" {
        return nil, nil, err
    }
    block, err := getIndexBlockByHashBin([]byte(acceptedBlockBin))
    if err != nil {
        return nil, nil, err
    }
    transaction, err := getIndexTransactionByTxIdBin(txIdBin)
    if err != nil {
        return nil, nil, err
    }
    return transaction, block, nil
}

////////////////////////////////
func GetIndexVspcListByDaaScore(score uint64, maxCount int, prev bool) ([]*protobook.Vspc, map[string]*protobook.Block, map[string]*protobook.Transaction, error) {
    daaScoreStart := uint64(0)
    daaScoreEnd := uint64(18446744073709551615)
    daaScoreExpired := uint64(0)
    if config.Rocksdb.DtlIndex > 0 {
        daaScoreExpired = GetDaaScoreLastRocks() - config.Rocksdb.DtlIndex
    }
    if !prev {
        daaScoreStart = score
    } else {
        daaScoreEnd = score + 1
    }
    if daaScoreStart < daaScoreExpired {
        daaScoreStart = daaScoreExpired
    }
    if maxCount > maxCountIndexVspc {
        maxCount = maxCountIndexVspc
    } else if maxCount <= 0 {
        maxCount = 1
    }
    daaScoreStartBe := make([]byte, 8)
    binary.BigEndian.PutUint64(daaScoreStartBe, daaScoreStart)
    daaScoreEndBe := make([]byte, 8)
    binary.BigEndian.PutUint64(daaScoreEndBe, daaScoreEnd)
    var keyStart []byte
    var keyEnd []byte
    keyStart = append([]byte(KeyPrefixVspc), daaScoreStartBe...)
    keyEnd = append([]byte(KeyPrefixVspc), daaScoreEndBe...)
    vspcList := make([]*protobook.Vspc, 0, maxCount)
    blockKeyList := make([]string, 0, maxCount)
    txKeyList := make([]string, 0, 256)
    err := seekCF(nil, cfIndex, keyStart, keyEnd, maxCount, prev, nil, func(i int, key []byte, val []byte) (bool, error) {
        if len(val) == 0 {
            return true, nil
        }
        vspc := &protobook.Vspc{}
        err := proto.Unmarshal(val, vspc)
        if err != nil {
            return false, err
        }
        vspcList = append(vspcList, vspc)
        blockKeyList = append(blockKeyList, KeyPrefixBlock+string(vspc.Hash))
        for _, txIdBin := range vspc.TxIds {
            txKeyList = append(txKeyList, KeyPrefixTransaction+string(txIdBin))
        }
        return true, nil
    })
    if err != nil {
        return nil, nil, nil, err
    }
    if len(vspcList) == 0 {
        return nil, nil, nil, nil
    }
    blockMap := make(map[string]*protobook.Block, len(vspcList))
    mutex := new(sync.RWMutex)
    _, err = doGetBatchCF(nil, cfIndex, blockKeyList, func(i int, val []byte) (error) {
        if len(val) == 0 {
            return nil
        }
        block := &protobook.Block{}
        err := proto.Unmarshal(val, block)
        if err == nil {
            mutex.Lock()
            blockMap[string(block.Hash)] = block
            mutex.Unlock()
        }
        return err
    })
    if err != nil {
        return nil, nil, nil, err
    }
    transactionMap := make(map[string]*protobook.Transaction, len(txKeyList))
    _, err = doGetBatchCF(nil, cfIndex, txKeyList, func(i int, val []byte) (error) {
        if len(val) == 0 {
            return nil
        }
        tx := &protobook.Transaction{}
        err := proto.Unmarshal(val, tx)
        if err == nil {
            mutex.Lock()
            transactionMap[string(tx.TxId)] = tx
            mutex.Unlock()
        }
        return err
    })
    if err != nil {
        return nil, nil, nil, err
    }
    return vspcList, blockMap, transactionMap, nil
}

////////////////////////////////
func GetIndexVspcListByBlueScore(score uint64, maxCount int, prev bool) ([]*protobook.Vspc, map[string]*protobook.Block, map[string]*protobook.Transaction, error) {
    blueScoreStart := uint64(0)
    blueScoreEnd := uint64(18446744073709551615)
    if !prev {
        blueScoreStart = score
    } else {
        blueScoreEnd = score + 1
    }
    blueScoreStartBe := make([]byte, 8)
    binary.BigEndian.PutUint64(blueScoreStartBe, blueScoreStart)
    blueScoreEndBe := make([]byte, 8)
    binary.BigEndian.PutUint64(blueScoreEndBe, blueScoreEnd)
    var keyStart []byte
    var keyEnd []byte
    keyStart = append([]byte(KeyPrefixBlue), blueScoreStartBe...)
    keyEnd = append([]byte(KeyPrefixBlue), blueScoreEndBe...)
    daaScore := uint64(0)
    err := seekCF(nil, cfIndex, keyStart, keyEnd, 1, prev, nil, func(i int, key []byte, val []byte) (bool, error) {
        if len(val) == 0 {
            return true, nil
        }
        daaScore = binary.BigEndian.Uint64(val)
        return false, nil
    })
    if err != nil {
        return nil, nil, nil, err
    }
    if daaScore == 0 {
        return nil, nil, nil, nil
    }
    return GetIndexVspcListByDaaScore(daaScore, maxCount, prev)
}

// ...
