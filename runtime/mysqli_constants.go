package runtime

import (
	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// GetMySQLiConstants returns all MySQLi predefined constants
func GetMySQLiConstants() []*registry.Constant {
	return []*registry.Constant{
		// Read options
		{Name: "MYSQLI_READ_DEFAULT_GROUP", Value: values.NewInt(5)},
		{Name: "MYSQLI_READ_DEFAULT_FILE", Value: values.NewInt(4)},
		{Name: "MYSQLI_OPT_CONNECT_TIMEOUT", Value: values.NewInt(0)},
		{Name: "MYSQLI_OPT_LOCAL_INFILE", Value: values.NewInt(8)},
		{Name: "MYSQLI_INIT_COMMAND", Value: values.NewInt(3)},
		{Name: "MYSQLI_OPT_READ_TIMEOUT", Value: values.NewInt(11)},
		{Name: "MYSQLI_OPT_NET_CMD_BUFFER_SIZE", Value: values.NewInt(202)},
		{Name: "MYSQLI_OPT_NET_READ_BUFFER_SIZE", Value: values.NewInt(203)},
		{Name: "MYSQLI_OPT_INT_AND_FLOAT_NATIVE", Value: values.NewInt(201)},
		{Name: "MYSQLI_OPT_SSL_VERIFY_SERVER_CERT", Value: values.NewInt(21)},

		// Client flags
		{Name: "MYSQLI_CLIENT_COMPRESS", Value: values.NewInt(32)},
		{Name: "MYSQLI_CLIENT_FOUND_ROWS", Value: values.NewInt(2)},
		{Name: "MYSQLI_CLIENT_IGNORE_SPACE", Value: values.NewInt(256)},
		{Name: "MYSQLI_CLIENT_INTERACTIVE", Value: values.NewInt(1024)},
		{Name: "MYSQLI_CLIENT_SSL", Value: values.NewInt(2048)},
		{Name: "MYSQLI_CLIENT_SSL_DONT_VERIFY_SERVER_CERT", Value: values.NewInt(64)},
		{Name: "MYSQLI_CLIENT_CAN_HANDLE_EXPIRED_PASSWORDS", Value: values.NewInt(4194304)},

		// Fetch modes
		{Name: "MYSQLI_ASSOC", Value: values.NewInt(1)},
		{Name: "MYSQLI_NUM", Value: values.NewInt(2)},
		{Name: "MYSQLI_BOTH", Value: values.NewInt(3)},

		// Store/Use result modes
		{Name: "MYSQLI_STORE_RESULT", Value: values.NewInt(0)},
		{Name: "MYSQLI_USE_RESULT", Value: values.NewInt(1)},
		{Name: "MYSQLI_ASYNC", Value: values.NewInt(8)},

		// Field types
		{Name: "MYSQLI_TYPE_DECIMAL", Value: values.NewInt(0)},
		{Name: "MYSQLI_TYPE_TINY", Value: values.NewInt(1)},
		{Name: "MYSQLI_TYPE_SHORT", Value: values.NewInt(2)},
		{Name: "MYSQLI_TYPE_LONG", Value: values.NewInt(3)},
		{Name: "MYSQLI_TYPE_FLOAT", Value: values.NewInt(4)},
		{Name: "MYSQLI_TYPE_DOUBLE", Value: values.NewInt(5)},
		{Name: "MYSQLI_TYPE_NULL", Value: values.NewInt(6)},
		{Name: "MYSQLI_TYPE_TIMESTAMP", Value: values.NewInt(7)},
		{Name: "MYSQLI_TYPE_LONGLONG", Value: values.NewInt(8)},
		{Name: "MYSQLI_TYPE_INT24", Value: values.NewInt(9)},
		{Name: "MYSQLI_TYPE_DATE", Value: values.NewInt(10)},
		{Name: "MYSQLI_TYPE_TIME", Value: values.NewInt(11)},
		{Name: "MYSQLI_TYPE_DATETIME", Value: values.NewInt(12)},
		{Name: "MYSQLI_TYPE_YEAR", Value: values.NewInt(13)},
		{Name: "MYSQLI_TYPE_NEWDATE", Value: values.NewInt(14)},
		{Name: "MYSQLI_TYPE_VARCHAR", Value: values.NewInt(15)},
		{Name: "MYSQLI_TYPE_ENUM", Value: values.NewInt(247)},
		{Name: "MYSQLI_TYPE_SET", Value: values.NewInt(248)},
		{Name: "MYSQLI_TYPE_TINY_BLOB", Value: values.NewInt(249)},
		{Name: "MYSQLI_TYPE_MEDIUM_BLOB", Value: values.NewInt(250)},
		{Name: "MYSQLI_TYPE_LONG_BLOB", Value: values.NewInt(251)},
		{Name: "MYSQLI_TYPE_BLOB", Value: values.NewInt(252)},
		{Name: "MYSQLI_TYPE_VAR_STRING", Value: values.NewInt(253)},
		{Name: "MYSQLI_TYPE_STRING", Value: values.NewInt(254)},
		{Name: "MYSQLI_TYPE_CHAR", Value: values.NewInt(1)},
		{Name: "MYSQLI_TYPE_INTERVAL", Value: values.NewInt(247)},
		{Name: "MYSQLI_TYPE_GEOMETRY", Value: values.NewInt(255)},
		{Name: "MYSQLI_TYPE_JSON", Value: values.NewInt(245)},
		{Name: "MYSQLI_TYPE_NEWDECIMAL", Value: values.NewInt(246)},
		{Name: "MYSQLI_TYPE_BIT", Value: values.NewInt(16)},

		// Field flags
		{Name: "MYSQLI_NOT_NULL_FLAG", Value: values.NewInt(1)},
		{Name: "MYSQLI_PRI_KEY_FLAG", Value: values.NewInt(2)},
		{Name: "MYSQLI_UNIQUE_KEY_FLAG", Value: values.NewInt(4)},
		{Name: "MYSQLI_MULTIPLE_KEY_FLAG", Value: values.NewInt(8)},
		{Name: "MYSQLI_BLOB_FLAG", Value: values.NewInt(16)},
		{Name: "MYSQLI_UNSIGNED_FLAG", Value: values.NewInt(32)},
		{Name: "MYSQLI_ZEROFILL_FLAG", Value: values.NewInt(64)},
		{Name: "MYSQLI_AUTO_INCREMENT_FLAG", Value: values.NewInt(512)},
		{Name: "MYSQLI_TIMESTAMP_FLAG", Value: values.NewInt(1024)},
		{Name: "MYSQLI_SET_FLAG", Value: values.NewInt(2048)},
		{Name: "MYSQLI_NUM_FLAG", Value: values.NewInt(32768)},
		{Name: "MYSQLI_PART_KEY_FLAG", Value: values.NewInt(16384)},
		{Name: "MYSQLI_GROUP_FLAG", Value: values.NewInt(32768)},
		{Name: "MYSQLI_ENUM_FLAG", Value: values.NewInt(256)},
		{Name: "MYSQLI_BINARY_FLAG", Value: values.NewInt(128)},
		{Name: "MYSQLI_NO_DEFAULT_VALUE_FLAG", Value: values.NewInt(4096)},
		{Name: "MYSQLI_ON_UPDATE_NOW_FLAG", Value: values.NewInt(8192)},

		// Cursor types
		{Name: "MYSQLI_CURSOR_TYPE_NO_CURSOR", Value: values.NewInt(0)},
		{Name: "MYSQLI_CURSOR_TYPE_READ_ONLY", Value: values.NewInt(1)},
		{Name: "MYSQLI_CURSOR_TYPE_FOR_UPDATE", Value: values.NewInt(2)},
		{Name: "MYSQLI_CURSOR_TYPE_SCROLLABLE", Value: values.NewInt(4)},

		// Statement attributes
		{Name: "MYSQLI_STMT_ATTR_UPDATE_MAX_LENGTH", Value: values.NewInt(0)},
		{Name: "MYSQLI_STMT_ATTR_CURSOR_TYPE", Value: values.NewInt(1)},
		{Name: "MYSQLI_STMT_ATTR_PREFETCH_ROWS", Value: values.NewInt(2)},

		// Report modes
		{Name: "MYSQLI_REPORT_OFF", Value: values.NewInt(0)},
		{Name: "MYSQLI_REPORT_ERROR", Value: values.NewInt(1)},
		{Name: "MYSQLI_REPORT_STRICT", Value: values.NewInt(2)},
		{Name: "MYSQLI_REPORT_INDEX", Value: values.NewInt(4)},
		{Name: "MYSQLI_REPORT_ALL", Value: values.NewInt(255)},

		// Transaction flags
		{Name: "MYSQLI_TRANS_START_READ_ONLY", Value: values.NewInt(4)},
		{Name: "MYSQLI_TRANS_START_READ_WRITE", Value: values.NewInt(2)},
		{Name: "MYSQLI_TRANS_START_CONSISTENT_SNAPSHOT", Value: values.NewInt(1)},

		// Refresh options
		{Name: "MYSQLI_REFRESH_GRANT", Value: values.NewInt(1)},
		{Name: "MYSQLI_REFRESH_LOG", Value: values.NewInt(2)},
		{Name: "MYSQLI_REFRESH_TABLES", Value: values.NewInt(4)},
		{Name: "MYSQLI_REFRESH_HOSTS", Value: values.NewInt(8)},
		{Name: "MYSQLI_REFRESH_STATUS", Value: values.NewInt(16)},
		{Name: "MYSQLI_REFRESH_THREADS", Value: values.NewInt(32)},
		{Name: "MYSQLI_REFRESH_SLAVE", Value: values.NewInt(64)},
		{Name: "MYSQLI_REFRESH_MASTER", Value: values.NewInt(128)},
		{Name: "MYSQLI_REFRESH_BACKUP_LOG", Value: values.NewInt(2097152)},

		// Server options
		{Name: "MYSQLI_SERVER_QUERY_NO_GOOD_INDEX_USED", Value: values.NewInt(16)},
		{Name: "MYSQLI_SERVER_QUERY_NO_INDEX_USED", Value: values.NewInt(32)},
		{Name: "MYSQLI_SERVER_QUERY_WAS_SLOW", Value: values.NewInt(2048)},
		{Name: "MYSQLI_SERVER_PS_OUT_PARAMS", Value: values.NewInt(4096)},

		// Set charset directory (deprecated)
		{Name: "MYSQLI_SET_CHARSET_DIR", Value: values.NewInt(6)},
		{Name: "MYSQLI_SET_CHARSET_NAME", Value: values.NewInt(7)},

		// Debug options
		{Name: "MYSQLI_DEBUG_TRACE_ENABLED", Value: values.NewInt(0)},

		// Data truncation
		{Name: "MYSQLI_DATA_TRUNCATED", Value: values.NewInt(101)},

		// No data
		{Name: "MYSQLI_NO_DATA", Value: values.NewInt(100)},
	}
}
