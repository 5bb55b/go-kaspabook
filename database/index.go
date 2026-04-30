
////////////////////////////////
package database

//#include "rocksdb/c.h"
import "C"
import (
    "sync"
    "time"
    "bytes"
    "context"
    "strconv"
    "log/slog"
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
const maxCountIndexTransaction = 100

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
func getIndexBlockMap(keyList []string, keyByScore bool) (map[string]*protobook.Block, error) {
    blockMap := make(map[string]*protobook.Block, len(keyList))
    mutex := new(sync.RWMutex)
    _, err := doGetBatchCF(nil, cfIndex, keyList, func(i int, val []byte) (error) {
        if len(val) == 0 {
            return nil
        }
        block := &protobook.Block{}
        err := proto.Unmarshal(val, block)
        if err == nil {
            var key string
            if keyByScore {
                daaScoreBe := make([]byte, 8)
                binary.BigEndian.PutUint64(daaScoreBe, block.DaaScore)
                key = string(daaScoreBe)
            } else {
                key = string(block.Hash)
            }
            mutex.Lock()
            blockMap[key] = block
            mutex.Unlock()
        }
        return err
    })
    return blockMap, err
}

////////////////////////////////
func getIndexTransactionMap(keyList []string) (map[string]*protobook.Transaction, error) {
    transactionMap := make(map[string]*protobook.Transaction, len(keyList))
    mutex := new(sync.RWMutex)
    _, err := doGetBatchCF(nil, cfIndex, keyList, func(i int, val []byte) (error) {
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
    return transactionMap, err
}

////////////////////////////////
func seekIndexDaaScoreByBlueScore(blueScore uint64, prev bool) (uint64, error) {
    blueScoreStart := uint64(0)
    blueScoreEnd := uint64(18446744073709551615)
    if !prev {
        blueScoreStart = blueScore
    } else {
        blueScoreEnd = blueScore + 1
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
        return 0, err
    }
    return daaScore, nil
}

////////////////////////////////
func seekIndexDaaScoreByTimestamp(timestamp uint64, prev bool) (uint64, error) {
    
    // ...
    
    return 0, nil
}

////////////////////////////////
func GetIndexVspcListByDaaScore(daaScore uint64, maxCount int, prev bool) ([]*protobook.Vspc, map[string]*protobook.Block, map[string]*protobook.Transaction, error) {
    daaScoreStart := uint64(0)
    daaScoreEnd := uint64(18446744073709551615)
    daaScoreExpired := uint64(0)
    if config.Rocksdb.DtlIndex > 0 {
        daaScoreExpired = GetDaaScoreLastRocks() - config.Rocksdb.DtlIndex
    }
    if !prev {
        daaScoreStart = daaScore
    } else {
        daaScoreEnd = daaScore + 1
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
    blockMap, err := getIndexBlockMap(blockKeyList, false)
    if err != nil {
        return nil, nil, nil, err
    }
    transactionMap, err := getIndexTransactionMap(txKeyList)
    if err != nil {
        return nil, nil, nil, err
    }
    return vspcList, blockMap, transactionMap, nil
}

////////////////////////////////
func GetIndexVspcListByBlueScore(blueScore uint64, maxCount int, prev bool) ([]*protobook.Vspc, map[string]*protobook.Block, map[string]*protobook.Transaction, error) {
    daaScore, err := seekIndexDaaScoreByBlueScore(blueScore, prev)
    if err != nil {
        return nil, nil, nil, err
    }
    return GetIndexVspcListByDaaScore(daaScore, maxCount, prev)
}

////////////////////////////////
func GetIndexAcceptedTransactionListByAddressDaaScore(address string, score uint64, maxCount int, prev bool) ([]*protobook.Transaction, []string, map[string]*protobook.Block, error) {
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
    if maxCount > maxCountIndexTransaction {
        maxCount = maxCountIndexTransaction
    } else if maxCount <= 0 {
        maxCount = 1
    }
    addressByte := []byte(address)
    daaScoreStartBe := make([]byte, 8)
    binary.BigEndian.PutUint64(daaScoreStartBe, daaScoreStart)
    daaScoreEndBe := make([]byte, 8)
    binary.BigEndian.PutUint64(daaScoreEndBe, daaScoreEnd)
    var keyStart []byte
    var keyEnd []byte
    keyStart = append([]byte(KeyPrefixAddress), addressByte...)
    keyStart = append(keyStart, 95)
    keyStart = append(keyStart, daaScoreStartBe...)
    keyEnd = append([]byte(KeyPrefixAddress), addressByte...)
    keyEnd = append(keyEnd, 95)
    keyEnd = append(keyEnd, daaScoreEndBe...)
    daaScoreBeMap := make(map[string]struct{}, maxCount)
    daaScoreBeList := make([]string, 0, maxCount)
    txIdList := make([]string, 0, maxCount)
    txKeyList := make([]string, 0, maxCount)
    err := seekCF(nil, cfIndex, keyStart, keyEnd, maxCount, prev, nil, func(i int, key []byte, val []byte) (bool, error) {
        keyList := bytes.SplitN(key, []byte{95}, 3)
        daaScoreBe := string(keyList[2][:8])
        daaScoreBeMap[daaScoreBe] = struct{}{}
        daaScoreBeList = append(daaScoreBeList, daaScoreBe)
        txId := string(keyList[2][9:])
        txIdList = append(txIdList, txId)
        txKeyList = append(txKeyList, KeyPrefixTransaction+txId)
        return true, nil
    })
    if err != nil {
        return nil, nil, nil, err
    }
    if len(txKeyList) == 0 {
        return nil, nil, nil, nil
    }
    vspcKeyList := make([]string, 0, len(daaScoreBeMap))
    for daaScoreBe := range daaScoreBeMap {
        vspcKeyList = append(vspcKeyList, KeyPrefixVspc+daaScoreBe)
    }
    blockKeyList := make([]string, len(vspcKeyList))
    _, err = doGetBatchCF(nil, cfIndex, vspcKeyList, func(i int, val []byte) (error) {
        if len(val) == 0 {
            return nil
        }
        vspc := &protobook.Vspc{}
        err := proto.Unmarshal(val, vspc)
        if err == nil {
            blockKeyList[i] = KeyPrefixBlock + string(vspc.Hash)
        }
        return err
    })
    if err != nil {
        return nil, nil, nil, err
    }
    blockMap, err := getIndexBlockMap(blockKeyList, true)
    if err != nil {
        return nil, nil, nil, err
    }
    transactionMap, err := getIndexTransactionMap(txKeyList)
    if err != nil {
        return nil, nil, nil, err
    }
    transactionList := make([]*protobook.Transaction, len(txIdList))
    for i, txId := range txIdList {
        transactionList[i] = transactionMap[txId]
    }
    return transactionList, daaScoreBeList, blockMap, nil
}

////////////////////////////////
func GetIndexAcceptedTransactionListByAddressBlueScore(address string, blueScore uint64, maxCount int, prev bool) ([]*protobook.Transaction, []string, map[string]*protobook.Block, error) {
    daaScore, err := seekIndexDaaScoreByBlueScore(blueScore, prev)
    if err != nil {
        return nil, nil, nil, err
    }
    return GetIndexAcceptedTransactionListByAddressDaaScore(address, daaScore, maxCount, prev)
}

////////////////////////////////
func GetIndexAcceptedTransactionListByAddressTimestamp(address string, timestamp uint64, maxCount int, prev bool) ([]*protobook.Transaction, []string, map[string]*protobook.Block, error) {
    daaScore, err := seekIndexDaaScoreByTimestamp(timestamp, prev)
    if err != nil {
        return nil, nil, nil, err
    }
    return GetIndexAcceptedTransactionListByAddressDaaScore(address, daaScore, maxCount, prev)
}

////////////////////////////////
func RunIndexCompactionLoop(ctx context.Context) {
    const interval = 28800
    go func() {
        ticker := interval - 30
        for {
            select {
            case <-ctx.Done():
                return
            default:
                time.Sleep(1 * time.Second)
                ticker ++
                if ticker >= interval {
                    mtss := time.Now().UnixMilli()
                    slog.Debug("database.RunIndexCompactionLoop starting.")
                    CompactCF(cfIndex)
                    slog.Debug("database.RunIndexCompactionLoop ended.", "mSecond", time.Now().UnixMilli()-mtss)
                    ticker = 0
                }
            }
        }
    }()
}
