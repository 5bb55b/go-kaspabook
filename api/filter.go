////////////////////////////////
package api

import (
    "fmt"
    "strings"
    "strconv"
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

////////////////////////////////
func filterUint(value string) (uint64, error) {
    value = strings.TrimSpace(value)
    if value == "" {
        return 0, fmt.Errorf("invalid")
    }
    valueUint, err := strconv.ParseUint(value, 10, 64)
    if err != nil {
        return 0, err
    }
    valueString := strconv.FormatUint(valueUint, 10)
    if (valueString != value) {
        return 0, fmt.Errorf("invalid")
    }
    return valueUint, nil
}
