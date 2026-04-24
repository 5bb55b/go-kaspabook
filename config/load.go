
////////////////////////////////
package config

import (
    "os"
    "fmt"
    "log"
    "strings"
    jsoniter "github.com/json-iterator/go"
    "github.com/jessevdk/go-flags"
)

////////////////////////////////
var json = jsoniter.ConfigCompatibleWithStandardLibrary

////////////////////////////////
const Version = "1.01.260420"

////////////////////////////////
type cmdConfig struct {
    // startup
    ShowConfig bool `long:"showconfig" description:"Show all configuration parameters."`
    Hysteresis int `long:"hysteresis" description:"Number of DAA Scores hysteresis for data scanning."`
    Concurrency int `long:"concurrency" description:"Number of concurrent workers."`
    Debug int `long:"debug" description:"Debug information level; [0-3] available."`
    // kaspad
    KaspadGrpc string `long:"kaspad-grpc" description:"Kaspa node gRPC endpoint (comma-separated for multiple)."`
    // rocksdb
    RocksPath string `long:"rocks-path" description:"RocksDB data path."`
    RocksDtl uint64 `long:"rocks-dtl" description:"Maximum DAA Score lifetime for indexed data."`
    RocksGcLoop bool `long:"rocks-gcloop" description:"Enable proactive compaction loop."`
    // Api
    ApiHost string `long:"api-host" description:"Listen host for the API server."`
    ApiPort int `long:"api-port" description:"Listen port for the API server."`
    ApiTimeout int `long:"api-timeout" description:"Processing timeout for the API server in seconds."`
    ApiConnMax int `long:"api-connmax" description:"Maximum number of concurrent connections for the API server."`
}

////////////////////////////////
type StartupConfig struct {
    Hysteresis int `json:"hysteresis"`
    Debug int `json:"debug"`
    Concurrency int `json:"concurrency"`
}
type KaspadConfig struct {
    Grpc []string `json:"grpc"`
}
type RocksConfig struct {
    Path string `json:"path"`
    DtlIndex uint64 `json:"dtlIndex"`
    GcLoop bool `json:"gcLoop"`
}
type ApiConfig struct {
    Host string `json:"host"`
    Port int `json:"port"`
    Timeout int `json:"timeout"`
    ConnMax int `json:"connMax"`
}

////////////////////////////////
var Startup StartupConfig
var Kaspad KaspadConfig
var Rocksdb RocksConfig
var Api ApiConfig

////////////////////////////////
var args = &cmdConfig{  // default
    Hysteresis: 100,
    Concurrency: 8,
    Debug: 2,
    KaspadGrpc: "127.0.0.1:16110",
    RocksPath: "./data",
    RocksDtl: 86400000,
    RocksGcLoop: true,
    ApiHost: "0.0.0.0",
    ApiPort: 8003,
    ApiTimeout: 15,
    ApiConnMax: 1000,
}

////////////////////////////////
func Load() {
    var err error
    parser := flags.NewParser(args, flags.Default)
    _, err = parser.Parse()
    if err != nil {
        errFlags, ok := err.(*flags.Error)
        if ok && errFlags.Type == flags.ErrHelp {
            os.Exit(0)
        }
        log.Fatalln("config.Load fatal:", err.Error())
    }
    if args.ShowConfig {
        defer func() {
            fmt.Println("")
            cfgStartup, _ := json.MarshalIndent(Startup, "", "    ")
            fmt.Println(`"startup": ` + string(cfgStartup))
            fmt.Println("")
            cfgKaspad, _ := json.MarshalIndent(Kaspad, "", "    ")
            fmt.Println(`"kaspad": ` + string(cfgKaspad))
            fmt.Println("")
            cfgRocksdb, _ := json.MarshalIndent(Rocksdb, "", "    ")
            fmt.Println(`"rocksdb": ` + string(cfgRocksdb))
            fmt.Println("")
            cfgApi, _ := json.MarshalIndent(Api, "", "    ")
            fmt.Println(`"api": ` + string(cfgApi))
            fmt.Println("")
            os.Exit(0)
        }()
    }
    Startup.Hysteresis = args.Hysteresis
    Startup.Debug = args.Debug
    kaspadGrpc := []string{}
    if args.KaspadGrpc != "" {
        kaspadGrpc = strings.Split(args.KaspadGrpc, ",")
    }
    Kaspad.Grpc = kaspadGrpc
    Rocksdb.Path = args.RocksPath
    Rocksdb.DtlIndex = args.RocksDtl
    Rocksdb.GcLoop = args.RocksGcLoop
    Api.Host = args.ApiHost
    Api.Port = args.ApiPort
    Api.Timeout = args.ApiTimeout
    Api.ConnMax = args.ApiConnMax
    if Startup.Hysteresis < 0 {
        Startup.Hysteresis = 0
    } else if Startup.Hysteresis > 36000 {
        Startup.Hysteresis = 36000
    }
    if Startup.Concurrency < 2 {
        Startup.Concurrency = 2
    } else if Startup.Concurrency > 32 {
        Startup.Concurrency = 32
    }
    if Rocksdb.DtlIndex > 0 && Rocksdb.DtlIndex < 2592000 {
        Rocksdb.DtlIndex = 2592000
    }
}
