
////////////////////////////////
package database

//#include "rocksdb/c.h"
import "C"
import (
    "log/slog"
)

////////////////////////////////
func Init() {
    initRocks()
    slog.Info("storage ready.")
}

////////////////////////////////
func Close() {
    destroyRocks()
    slog.Info("storage released.")
}
