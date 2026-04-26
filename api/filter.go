////////////////////////////////
package api

import (
    "fmt"
    "strings"
    "encoding/hex"
)

////////////////////////////////
func filterHash(hash string) (string, error) {
    hash = strings.TrimSpace(hash)
    hash = strings.ToLower(hash)
    if len(hash) != 64 {
        return "", fmt.Errorf("invalid")
    }
    _, err := hex.DecodeString(hash)
    if err != nil {
        return "", err
    }
    return hash, nil
}
