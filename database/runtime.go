
////////////////////////////////
package database

//#include "rocksdb/c.h"
import "C"
import (
    "sync"
    "encoding/binary"
)

////////////////////////////////
type DbRuntimeStatusType struct {
    // kaspad
    VersionKaspad string `json:"versionKaspad,omitempty"`
    DaaScoreKaspad string `json:"daaScoreKaspad,omitempty"`
    StatusKaspad string `json:"statusKaspad,omitempty"`
    // book
    VersionBook string `json:"versionBook,omitempty"`
    DaaScoreBook string `json:"daaScoreBook,omitempty"`
    BlueScoreBook string `json:"blueScoreBook,omitempty"`
    ScannedBook string `json:"scannedBook,omitempty"`
    GapBook string `json:"gapBook,omitempty"`
    SizeBook string `json:"sizeBook,omitempty"`
    // config
    Network string `json:"network,omitempty"`
    Hysteresis string `json:"hysteresis,omitempty"`
    DtlIndex string `json:"dtlIndex,omitempty"`
    // runtime
    DaaScoreKaspadInt uint64
    DaaScoreBookInt uint64
    LenTransaction int
}

////////////////////////////////
const dtlIddKeys = 864000
const keyRuntimeStatus = "runtime_status"
const keyPrefixIddkeys = "iddkeys_"

////////////////////////////////
func GetRuntimeStatus() (*DbRuntimeStatusType, error) {
    val, err := getCF(nil, cfState, []byte(keyRuntimeStatus), nil)
    if err != nil {
        return nil, err
    }
    data := &DbRuntimeStatusType{}
    if len(val) == 0 {
        return data, nil
    }
    err = json.Unmarshal(val, data)
    if err != nil {
        return nil, err
    }
    return data, nil
}

////////////////////////////////
func SetRuntimeStatus(tx *C.rocksdb_transaction_t, data *DbRuntimeStatusType) (error) {
    val, err := json.Marshal(data)
    if err != nil {
        return err
    }
    return putCF(tx, cfState, []byte(keyRuntimeStatus), val, 0)
}

////////////////////////////////
func getIddkeysByDaaScoreList(daaScoreList []uint64) (map[string]struct{}, error) {
    lenDaaScore := len(daaScoreList)
    iddkeys := make(map[string]struct{}, lenDaaScore)
    if lenDaaScore == 0 {
        return iddkeys, nil
    }
    keyList := make([]string, lenDaaScore)
    for i, daaScore := range daaScoreList {
        daaScoreBe := make([]byte, 8)
        binary.BigEndian.PutUint64(daaScoreBe, daaScore)
        keyList[i] = keyPrefixIddkeys + string(daaScoreBe)
    }
    mutex := new(sync.RWMutex)
    _, err := doGetBatchCF(nil, cfState, keyList, func(i int, val []byte) (error) {
        if len(val) == 0 {
            return nil
        }
        decoded := [][]byte{}
        err := json.Unmarshal(val, &decoded)
        if err != nil {
            return err
        }
        for _, key := range decoded {
            mutex.Lock()
            iddkeys[string(key)] = struct{}{}
            mutex.Unlock()
        }
        return nil
    })
    if err != nil {
        return nil, err
    }
    return iddkeys, nil
}

////////////////////////////////
func delIddkeys(tx *C.rocksdb_transaction_t, iddkeys map[string]struct{}) (error) {
    for key := range iddkeys {
        err := deleteCF(tx, cfState, []byte(key))
        if err != nil {
            return err
        }
    }
    return nil
}

////////////////////////////////
func setIddkeys(tx *C.rocksdb_transaction_t, iddkeysList map[string][]string) (error) {
    for key, list := range iddkeysList {
        listBin := make([][]byte, len(list))
        for i, iddKey := range list {
            listBin[i] = []byte(iddKey)
        }
        val, err := json.Marshal(&listBin)
        if err != nil {
            return err
        }
        err = putCF(tx, cfState, []byte(key), val, 0)
        if err != nil {
            return err
        }
    }
    return nil
}

////////////////////////////////
func ExpireIddkeys(daaScore uint64) (error) {
    daaScore -= dtlIddKeys
    daaScoreBe := make([]byte, 8)
    binary.BigEndian.PutUint64(daaScoreBe, daaScore)
    keyStart := keyPrefixIddkeys
    keyEnd := keyPrefixIddkeys + string(daaScoreBe)
    err := deleteRangeCF(cfState, []byte(keyStart), []byte(keyEnd))
    if err != nil {
        return err
    }
    return nil
}
