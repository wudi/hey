package runtime

import "github.com/wudi/hey/values"

// getPDOConstants returns all PDO class constants
func getPDOConstants() map[string]*values.Value {
	return map[string]*values.Value{
		// Fetch modes
		"FETCH_LAZY":      values.NewInt(1),
		"FETCH_ASSOC":     values.NewInt(2),
		"FETCH_NUM":       values.NewInt(3),
		"FETCH_BOTH":      values.NewInt(4),
		"FETCH_OBJ":       values.NewInt(5),
		"FETCH_BOUND":     values.NewInt(6),
		"FETCH_COLUMN":    values.NewInt(7),
		"FETCH_CLASS":     values.NewInt(8),
		"FETCH_INTO":      values.NewInt(9),
		"FETCH_FUNC":      values.NewInt(10),
		"FETCH_KEY_PAIR":  values.NewInt(12),

		// Parameter types
		"PARAM_NULL":      values.NewInt(0),
		"PARAM_INT":       values.NewInt(1),
		"PARAM_STR":       values.NewInt(2),
		"PARAM_LOB":       values.NewInt(3),
		"PARAM_BOOL":      values.NewInt(5),

		// Error modes
		"ERRMODE_SILENT":    values.NewInt(0),
		"ERRMODE_WARNING":   values.NewInt(1),
		"ERRMODE_EXCEPTION": values.NewInt(2),

		// Attributes
		"ATTR_AUTOCOMMIT":         values.NewInt(0),
		"ATTR_PREFETCH":           values.NewInt(1),
		"ATTR_TIMEOUT":            values.NewInt(2),
		"ATTR_ERRMODE":            values.NewInt(3),
		"ATTR_SERVER_VERSION":     values.NewInt(4),
		"ATTR_CLIENT_VERSION":     values.NewInt(5),
		"ATTR_SERVER_INFO":        values.NewInt(6),
		"ATTR_CONNECTION_STATUS":  values.NewInt(7),
		"ATTR_CASE":               values.NewInt(8),
		"ATTR_CURSOR_NAME":        values.NewInt(9),
		"ATTR_CURSOR":             values.NewInt(10),
		"ATTR_ORACLE_NULLS":       values.NewInt(11),
		"ATTR_PERSISTENT":         values.NewInt(12),
		"ATTR_STATEMENT_CLASS":    values.NewInt(13),
		"ATTR_FETCH_TABLE_NAMES":  values.NewInt(14),
		"ATTR_FETCH_CATALOG_NAMES": values.NewInt(15),
		"ATTR_DRIVER_NAME":        values.NewInt(16),
		"ATTR_STRINGIFY_FETCHES":  values.NewInt(17),
		"ATTR_MAX_COLUMN_LEN":     values.NewInt(18),
		"ATTR_EMULATE_PREPARES":   values.NewInt(20),
		"ATTR_DEFAULT_FETCH_MODE": values.NewInt(19),

		// Cursor types
		"CURSOR_FWDONLY": values.NewInt(0),
		"CURSOR_SCROLL":  values.NewInt(1),

		// NULL handling
		"NULL_NATURAL":     values.NewInt(0),
		"NULL_EMPTY_STRING": values.NewInt(1),
		"NULL_TO_STRING":   values.NewInt(2),

		// Case conversion
		"CASE_NATURAL": values.NewInt(0),
		"CASE_LOWER":   values.NewInt(2),
		"CASE_UPPER":   values.NewInt(1),

		// Fetch orientation
		"FETCH_ORI_NEXT":     values.NewInt(0),
		"FETCH_ORI_PRIOR":    values.NewInt(1),
		"FETCH_ORI_FIRST":    values.NewInt(2),
		"FETCH_ORI_LAST":     values.NewInt(3),
		"FETCH_ORI_ABS":      values.NewInt(4),
		"FETCH_ORI_REL":      values.NewInt(5),
	}
}

// GetPDOGlobalConstants returns PDO constants for global scope
func GetPDOGlobalConstants() map[string]*values.Value {
	constants := make(map[string]*values.Value)

	// Add all PDO class constants to global scope with PDO:: prefix
	pdoConstants := getPDOConstants()
	for name, value := range pdoConstants {
		// These will be accessible as PDO::FETCH_ASSOC
		constants["PDO::"+name] = value
	}

	return constants
}
