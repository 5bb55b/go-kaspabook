
////////////////////////////////
package kaspa

import (
    "fmt"
    "log/slog"
)

////////////////////////////////
const cacheBlockScoreMax = uint64(2048)

////////////////////////////////
type blockScoreType struct {
    Daa uint64
    Blue uint64
}
type cacheBlockScoreType struct {
    Index []string
    Score map[string]*blockScoreType
}
var cacheBlockScore cacheBlockScoreType

////////////////////////////////
func initCache() {
    cacheBlockScore.Index = make([]string, 0, cacheBlockScoreMax*2)
    cacheBlockScore.Score = make(map[string]*blockScoreType)
    // ...
}

////////////////////////////////
func getBlockScore(hash string) (*blockScoreType, error) {
    score, exists := cacheBlockScore.Score[hash]
    if exists && score != nil {
        return score, nil
    }
    r, err := grpcKaspa.GetBlock(hash)
    if err != nil {
        return nil, err
    }
    score = &blockScoreType{
        Daa: r.Block.Header.DaaScore,
        Blue: r.Block.Header.BlueScore,
    }
    slog.Info("kaspa.grpcGetBlock", "hash", hash, "daaScore", score.Daa, "blueScore", score.Blue)
    if score.Daa == 0 || score.Blue == 0 {
        return nil, fmt.Errorf("nil block")
    }
    addCacheBlockScore(hash, score)
    return score, nil
}

////////////////////////////////
func addCacheBlockScore(hash string, score *blockScoreType) {
    if (hash == "" || score == nil || score.Daa == 0 || score.Blue == 0) {
        return
    }
    if cacheBlockScore.Score[hash] == nil {
        cacheBlockScore.Score[hash] = score
        cacheBlockScore.Index = append(cacheBlockScore.Index, hash)
    }
}

////////////////////////////////
func expireCacheBlockScore(daaScore uint64) {
    if daaScore <= cacheBlockScoreMax {
        return
    }
    daaScore -= cacheBlockScoreMax
    s := 0
    lenCache := len(cacheBlockScore.Index)
    for i := 0; i < lenCache; i++ {
        hash := cacheBlockScore.Index[i]
        if cacheBlockScore.Score[hash].Daa > daaScore {
            s = i
            break
        }
        delete(cacheBlockScore.Score, hash)
    }
    if s > 0 {
        cacheBlockScore.Index = cacheBlockScore.Index[s:]
    }
}
