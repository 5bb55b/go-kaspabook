
////////////////////////////////
package database

//#include "rocksdb/c.h"
import "C"
import (
    "log/slog"
    jsoniter "github.com/json-iterator/go"
)

////////////////////////////////
var json = jsoniter.ConfigCompatibleWithStandardLibrary

////////////////////////////////
func Init() {
    initRocks()
    slog.Info("database ready.")
}

////////////////////////////////
func Close() {
    destroyRocks()
    slog.Info("database released.")
}
