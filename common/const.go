package common

import "time"

const (
	MaxTaosSqlLen   = 1048576
	DefaultUser     = "root"
	DefaultPassword = "taosdata"
)

const (
	PrecisionMilliSecond = 0
	PrecisionMicroSecond = 1
	PrecisionNanoSecond  = 2
)

const (
	DefaultHandlerRecycleCheckInterval   = 3 * time.Second
	DefaultHandlerRecycleElemMaxLifeTime = 1 * time.Minute
)

const (
	TSDB_OPTION_LOCALE = iota
	TSDB_OPTION_CHARSET
	TSDB_OPTION_TIMEZONE
	TSDB_OPTION_CONFIGDIR
	TSDB_OPTION_SHELL_ACTIVITY_TIMER
	TSDB_MAX_OPTIONS
)

const (
	TSDB_DATA_TYPE_NULL      = 0  // 1 bytes
	TSDB_DATA_TYPE_BOOL      = 1  // 1 bytes
	TSDB_DATA_TYPE_TINYINT   = 2  // 1 byte
	TSDB_DATA_TYPE_SMALLINT  = 3  // 2 bytes
	TSDB_DATA_TYPE_INT       = 4  // 4 bytes
	TSDB_DATA_TYPE_BIGINT    = 5  // 8 bytes
	TSDB_DATA_TYPE_FLOAT     = 6  // 4 bytes
	TSDB_DATA_TYPE_DOUBLE    = 7  // 8 bytes
	TSDB_DATA_TYPE_BINARY    = 8  // string
	TSDB_DATA_TYPE_TIMESTAMP = 9  // 8 bytes
	TSDB_DATA_TYPE_NCHAR     = 10 // unicode string
	TSDB_DATA_TYPE_UTINYINT  = 11 // 1 byte
	TSDB_DATA_TYPE_USMALLINT = 12 // 2 bytes
	TSDB_DATA_TYPE_UINT      = 13 // 4 bytes
	TSDB_DATA_TYPE_UBIGINT   = 14 // 8 bytes
	TSDB_DATA_TYPE_JSON      = 15
)

var TypeNameMap = map[int]string{
	TSDB_DATA_TYPE_NULL:      "NULL",
	TSDB_DATA_TYPE_BOOL:      "BOOL",
	TSDB_DATA_TYPE_TINYINT:   "TINYINT",
	TSDB_DATA_TYPE_SMALLINT:  "SMALLINT",
	TSDB_DATA_TYPE_INT:       "INT",
	TSDB_DATA_TYPE_BIGINT:    "BIGINT",
	TSDB_DATA_TYPE_FLOAT:     "FLOAT",
	TSDB_DATA_TYPE_DOUBLE:    "DOUBLE",
	TSDB_DATA_TYPE_BINARY:    "BINARY",
	TSDB_DATA_TYPE_TIMESTAMP: "TIMESTAMP",
	TSDB_DATA_TYPE_NCHAR:     "NCHAR",
	TSDB_DATA_TYPE_UTINYINT:  "TINYINT UNSIGNED",
	TSDB_DATA_TYPE_USMALLINT: "SMALLINT UNSIGNED",
	TSDB_DATA_TYPE_UINT:      "INT UNSIGNED",
	TSDB_DATA_TYPE_UBIGINT:   "BIGINT UNSIGNED",
	TSDB_DATA_TYPE_JSON:      "JSON",
}
