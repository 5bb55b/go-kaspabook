//go:build linux && amd64
////////////////////////////////
package main

import (
    "fmt"
    
    // ...
    
    "kaspabook/config"
    
    // ...
    
)

////////////////////////////////
func main() {
    fmt.Println("KaspaBOOK v"+config.Version)
    
    // ...
    
}
