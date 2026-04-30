//go:build linux && amd64
////////////////////////////////
package main

import (
    "os"
    "fmt"
    "log"
    "sync"
    "strings"
    "syscall"
    "context"
    "log/slog"
    "os/signal"
    "path/filepath"
    "kaspabook/config"
    "kaspabook/database"
    "kaspabook/api"
    "kaspabook/kaspa"
)

////////////////////////////////
func main() {
    fmt.Println("KaspaBOOK v"+config.Version)
    
    // Set the correct working directory.
    arg0 := os.Args[0]
    if strings.Index(arg0, "go-build") < 0 {
        dir, err := filepath.Abs(filepath.Dir(arg0))
        if err != nil {
            log.Fatalln("main fatal:", err.Error())
        }
        os.Chdir(dir)
    }

    // Load config.
    config.Load()
    
    // Set the log level.
    logOpt := &slog.HandlerOptions{Level: slog.LevelError,}
    if config.Startup.Debug == 3 {
        logOpt = &slog.HandlerOptions{Level: slog.LevelDebug,}
    } else if config.Startup.Debug == 2 {
        logOpt = &slog.HandlerOptions{Level: slog.LevelInfo,}
    } else if config.Startup.Debug == 1 {
        logOpt = &slog.HandlerOptions{Level: slog.LevelWarn,}
    }
    logHandler := slog.NewTextHandler(os.Stdout, logOpt)
    slog.SetDefault(slog.New(logHandler))
    
    // Set exit signal.
    ctx, cancel := context.WithCancel(context.Background())
    wg := &sync.WaitGroup{}
    c := make(chan os.Signal, 8)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    down := false
    go func() {
        wg.Add(1)
        defer wg.Done()
        <-c
        slog.Info("main stopping ..")
        cancel()
        down = true
    }()
    
    // Init database driver.
    database.Init()
    
    // Init api server
    api.Init(c)
    
    // Init scanner if api server up.
    if (!down) {
        err := kaspa.Init(ctx)
        if err != nil {
            slog.Error("kaspa.Init fatal.", "error", err.Error())
            c <- syscall.SIGTERM
        } else {
            go func() {
                wg.Add(1)
                defer wg.Done()
                kaspa.Run()
            }()
        }
    }
    
    // Waiting and exit.
    wg.Wait()
    api.Shutdown()
    database.Close()
    slog.Info("main exited.")
}
