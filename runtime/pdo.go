package runtime

import (
	"fmt"
	"strings"

	"github.com/wudi/hey/pkg/pdo"
	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// setPDOError sets the error state for a PDO object
func setPDOError(obj *values.Object, sqlState string, driverCode int, message string) {
	obj.Properties["__pdo_error_code"] = values.NewString(sqlState)

	errorInfo := values.NewArray()
	errorInfo.ArraySet(values.NewInt(0), values.NewString(sqlState))
	errorInfo.ArraySet(values.NewInt(1), values.NewInt(int64(driverCode)))
	errorInfo.ArraySet(values.NewInt(2), values.NewString(message))

	obj.Properties["__pdo_error_info"] = errorInfo
}

// clearPDOError clears the error state for a PDO object
func clearPDOError(obj *values.Object) {
	obj.Properties["__pdo_error_code"] = values.NewString("00000")
	obj.Properties["__pdo_error_info"] = values.NewNull()
}

// extractSQLState extracts SQLSTATE from error message
func extractSQLState(err error) string {
	if err == nil {
		return "00000"
	}

	errMsg := err.Error()

	// Try to extract SQLSTATE from common patterns
	// MySQL: "Error 1146: Table 'test.foo' doesn't exist"
	// PostgreSQL: "pq: error code 42P01: relation does not exist"
	// SQLite: "SQL logic error: no such table: foo (1)"

	if strings.Contains(errMsg, "no such table") || strings.Contains(errMsg, "doesn't exist") {
		return "42S02" // Base table or view not found
	}
	if strings.Contains(errMsg, "syntax error") {
		return "42000" // Syntax error
	}
	if strings.Contains(errMsg, "Access denied") || strings.Contains(errMsg, "authentication failed") {
		return "28000" // Invalid authorization
	}
	if strings.Contains(errMsg, "duplicate") || strings.Contains(errMsg, "unique constraint") {
		return "23000" // Integrity constraint violation
	}

	// Default to general error
	return "HY000"
}

// pdoConstruct implements new PDO($dsn, $username, $password, $options)
// args[0] = $this, args[1] = dsn, args[2] = username, args[3] = password, args[4] = options
func pdoConstruct(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("PDO::__construct() expects at least 1 parameter")
	}

	thisObj := args[0]
	dsn := args[1].Data.(string)
	username := ""
	password := ""

	if len(args) > 2 && args[2].Type != values.TypeNull {
		username = args[2].Data.(string)
	}
	if len(args) > 3 && args[3].Type != values.TypeNull {
		password = args[3].Data.(string)
	}

	// Parse DSN to get driver
	dsnInfo, err := pdo.ParseDSN(dsn)
	if err != nil {
		return nil, fmt.Errorf("invalid DSN: %v", err)
	}

	// Get driver
	driver, ok := pdo.GetDriver(dsnInfo.Driver)
	if !ok {
		return nil, fmt.Errorf("could not find driver: %s", dsnInfo.Driver)
	}

	// Open connection
	conn, err := driver.Open(dsn)
	if err != nil {
		return nil, fmt.Errorf("connection failed: %v", err)
	}

	// Driver-specific connection setup
	switch dsnInfo.Driver {
	case "mysql":
		if mysqlConn, ok := conn.(*pdo.MySQLConn); ok {
			if err := mysqlConn.Connect(username, password); err != nil {
				return nil, fmt.Errorf("connection failed: %v", err)
			}
		}
	case "sqlite":
		if sqliteConn, ok := conn.(*pdo.SQLiteConn); ok {
			if err := sqliteConn.Connect(); err != nil {
				return nil, fmt.Errorf("connection failed: %v", err)
			}
		}
	case "pgsql":
		if pgsqlConn, ok := conn.(*pdo.PgSQLConn); ok {
			if err := pgsqlConn.Connect(username, password); err != nil {
				return nil, fmt.Errorf("connection failed: %v", err)
			}
		}
	}

	// Store connection in object properties
	obj := thisObj.Data.(*values.Object)
	if obj.Properties == nil {
		obj.Properties = make(map[string]*values.Value)
	}

	obj.Properties["__pdo_conn"] = values.NewResource(conn)
	obj.Properties["__pdo_driver"] = values.NewString(dsnInfo.Driver)
	obj.Properties["__pdo_in_tx"] = values.NewBool(false)
	obj.Properties["__pdo_tx"] = values.NewNull()
	obj.Properties["__pdo_error_code"] = values.NewString("00000") // Success SQLSTATE
	obj.Properties["__pdo_error_info"] = values.NewNull()

	// Initialize attributes with defaults
	attributes := values.NewArray()
	// PDO::ATTR_ERRMODE = 3, default is ERRMODE_SILENT = 0
	attributes.ArraySet(values.NewInt(3), values.NewInt(0))
	// PDO::ATTR_DEFAULT_FETCH_MODE = 19, default is FETCH_BOTH = 4
	attributes.ArraySet(values.NewInt(19), values.NewInt(4))
	// PDO::ATTR_CASE = 8, default is CASE_NATURAL = 0
	attributes.ArraySet(values.NewInt(8), values.NewInt(0))
	// PDO::ATTR_AUTOCOMMIT = 0, default is true (1)
	attributes.ArraySet(values.NewInt(0), values.NewInt(1))
	obj.Properties["__pdo_attributes"] = attributes

	return values.NewNull(), nil
}

// convertNamedToPositionalParams converts named parameters (:name) to positional (?)
// Returns the converted query and a map of parameter names to positions
func convertNamedToPositionalParams(query string) (string, map[string]int) {
	paramMap := make(map[string]int)
	result := strings.Builder{}
	paramIndex := 1
	i := 0

	for i < len(query) {
		// Skip strings to avoid replacing : inside quotes
		if query[i] == '\'' || query[i] == '"' {
			quote := query[i]
			result.WriteByte(query[i])
			i++
			for i < len(query) {
				result.WriteByte(query[i])
				if query[i] == quote && (i == 0 || query[i-1] != '\\') {
					i++
					break
				}
				i++
			}
			continue
		}

		// Check for named parameter
		if query[i] == ':' && i+1 < len(query) {
			// Extract parameter name
			start := i + 1
			end := start
			for end < len(query) && (isAlphaNumeric(query[end]) || query[end] == '_') {
				end++
			}

			if end > start {
				paramName := query[start:end]
				if _, exists := paramMap[paramName]; !exists {
					paramMap[paramName] = paramIndex
					paramIndex++
				}
				result.WriteByte('?')
				i = end
				continue
			}
		}

		result.WriteByte(query[i])
		i++
	}

	return result.String(), paramMap
}

func isAlphaNumeric(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}

// pdoPrepare implements $pdo->prepare($query)
// args[0] = $this, args[1] = query, args[2] = options
func pdoPrepare(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 2 {
		return values.NewBool(false), fmt.Errorf("PDO::prepare() expects at least 1 parameter")
	}

	thisObj := args[0]
	originalQuery := args[1].Data.(string)

	// Convert named parameters to positional if needed
	query, paramMap := convertNamedToPositionalParams(originalQuery)

	// Get connection from object properties
	obj := thisObj.Data.(*values.Object)
	connVal, ok := obj.Properties["__pdo_conn"]
	if !ok {
		return values.NewBool(false), fmt.Errorf("invalid PDO object: no connection")
	}

	conn := connVal.Data.(pdo.Conn)

	stmt, err := conn.Prepare(query)
	if err != nil {
		return values.NewBool(false), nil
	}

	// Create PDOStatement object
	stmtObj := values.NewObject("PDOStatement")
	stmtObjData := stmtObj.Data.(*values.Object)
	stmtObjData.Properties["queryString"] = values.NewString(originalQuery)
	stmtObjData.Properties["__pdo_stmt"] = values.NewResource(stmt)
	stmtObjData.Properties["__pdo_rows"] = values.NewNull()

	// Store parameter map for named parameter binding
	if len(paramMap) > 0 {
		paramMapValue := values.NewArray()
		for name, pos := range paramMap {
			paramMapValue.ArraySet(values.NewString(name), values.NewInt(int64(pos)))
		}
		stmtObjData.Properties["__pdo_param_map"] = paramMapValue
	}

	return stmtObj, nil
}

// pdoQuery implements $pdo->query($query)
// args[0] = $this, args[1] = query
func pdoQuery(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 2 {
		return values.NewBool(false), fmt.Errorf("PDO::query() expects at least 1 parameter")
	}

	thisObj := args[0]
	query := args[1].Data.(string)

	// Get connection
	obj := thisObj.Data.(*values.Object)

	var rows pdo.Rows
	var err error

	// Check if we're in a transaction
	inTxVal, hasInTx := obj.Properties["__pdo_in_tx"]
	inTx := hasInTx && inTxVal.Type == values.TypeBool && inTxVal.Data.(bool)

	if inTx {
		txVal := obj.Properties["__pdo_tx"]
		if txVal.Type != values.TypeNull {
			tx := txVal.Data.(pdo.Tx)
			rows, err = tx.Query(query)
		} else {
			return values.NewBool(false), fmt.Errorf("invalid transaction state")
		}
	} else {
		connVal, ok := obj.Properties["__pdo_conn"]
		if !ok {
			return values.NewBool(false), fmt.Errorf("invalid PDO object")
		}
		conn := connVal.Data.(pdo.Conn)
		rows, err = conn.Query(query)
	}

	if err != nil {
		// Set error state
		sqlState := extractSQLState(err)
		setPDOError(obj, sqlState, 1, err.Error())
		return values.NewBool(false), nil
	}

	if rows == nil {
		setPDOError(obj, "HY000", 1, "query failed to return rows")
		return values.NewBool(false), fmt.Errorf("query failed to return rows")
	}

	// Clear error state on success
	clearPDOError(obj)

	// Create PDOStatement object with active result set
	stmtObj := values.NewObject("PDOStatement")
	stmtObjData := stmtObj.Data.(*values.Object)
	stmtObjData.Properties["queryString"] = values.NewString(query)
	stmtObjData.Properties["__pdo_rows"] = values.NewResource(rows)

	return stmtObj, nil
}

// pdoExec implements $pdo->exec($statement)
// args[0] = $this, args[1] = statement
func pdoExec(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 2 {
		return values.NewBool(false), fmt.Errorf("PDO::exec() expects 1 parameter")
	}

	thisObj := args[0]
	query := args[1].Data.(string)

	// Get connection
	obj := thisObj.Data.(*values.Object)

	var result pdo.Result
	var err error

	// Check if we're in a transaction
	inTxVal, hasInTx := obj.Properties["__pdo_in_tx"]
	inTx := hasInTx && inTxVal.Type == values.TypeBool && inTxVal.Data.(bool)

	if inTx {
		txVal := obj.Properties["__pdo_tx"]
		if txVal.Type != values.TypeNull {
			tx := txVal.Data.(pdo.Tx)
			result, err = tx.Exec(query)
		} else {
			return values.NewBool(false), fmt.Errorf("invalid transaction state")
		}
	} else {
		connVal, ok := obj.Properties["__pdo_conn"]
		if !ok {
			return values.NewBool(false), fmt.Errorf("invalid PDO object")
		}
		conn := connVal.Data.(pdo.Conn)
		result, err = conn.Exec(query)
	}

	if err != nil {
		// Set error state
		sqlState := extractSQLState(err)
		setPDOError(obj, sqlState, 1, err.Error())
		return values.NewBool(false), nil
	}

	if result == nil {
		setPDOError(obj, "HY000", 1, "exec failed to return result")
		return values.NewBool(false), fmt.Errorf("exec failed to return result")
	}

	// Clear error state on success
	clearPDOError(obj)

	affected, _ := result.RowsAffected()
	return values.NewInt(affected), nil
}

// pdoLastInsertId implements $pdo->lastInsertId()
// args[0] = $this, args[1] = name (optional)
func pdoLastInsertId(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	thisObj := args[0]

	// Get connection
	obj := thisObj.Data.(*values.Object)
	connVal, ok := obj.Properties["__pdo_conn"]
	if !ok {
		return values.NewString("0"), fmt.Errorf("invalid PDO object")
	}

	conn := connVal.Data.(pdo.Conn)

	id, err := conn.LastInsertId()
	if err != nil {
		return values.NewString("0"), nil
	}

	return values.NewString(fmt.Sprintf("%d", id)), nil
}

// pdoBeginTransaction implements $pdo->beginTransaction()
// args[0] = $this
func pdoBeginTransaction(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	thisObj := args[0]

	// Get connection
	obj := thisObj.Data.(*values.Object)
	connVal, ok := obj.Properties["__pdo_conn"]
	if !ok {
		return values.NewBool(false), fmt.Errorf("invalid PDO object")
	}

	// Check if already in transaction
	if obj.Properties["__pdo_in_tx"].Data.(bool) {
		return values.NewBool(false), fmt.Errorf("already in transaction")
	}

	conn := connVal.Data.(pdo.Conn)

	tx, err := conn.Begin()
	if err != nil {
		return values.NewBool(false), nil
	}

	obj.Properties["__pdo_tx"] = values.NewResource(tx)
	obj.Properties["__pdo_in_tx"] = values.NewBool(true)

	return values.NewBool(true), nil
}

// pdoCommit implements $pdo->commit()
// args[0] = $this
func pdoCommit(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	thisObj := args[0]

	// Get transaction
	obj := thisObj.Data.(*values.Object)

	if !obj.Properties["__pdo_in_tx"].Data.(bool) {
		return values.NewBool(false), fmt.Errorf("no active transaction")
	}

	txVal := obj.Properties["__pdo_tx"]
	if txVal.Type == values.TypeNull {
		return values.NewBool(false), fmt.Errorf("no active transaction")
	}

	tx := txVal.Data.(pdo.Tx)

	if err := tx.Commit(); err != nil {
		return values.NewBool(false), nil
	}

	obj.Properties["__pdo_in_tx"] = values.NewBool(false)
	obj.Properties["__pdo_tx"] = values.NewNull()

	return values.NewBool(true), nil
}

// pdoRollBack implements $pdo->rollBack()
// args[0] = $this
func pdoRollBack(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	thisObj := args[0]

	// Get transaction
	obj := thisObj.Data.(*values.Object)

	if !obj.Properties["__pdo_in_tx"].Data.(bool) {
		return values.NewBool(false), fmt.Errorf("no active transaction")
	}

	txVal := obj.Properties["__pdo_tx"]
	if txVal.Type == values.TypeNull {
		return values.NewBool(false), fmt.Errorf("no active transaction")
	}

	tx := txVal.Data.(pdo.Tx)

	if err := tx.Rollback(); err != nil {
		return values.NewBool(false), nil
	}

	obj.Properties["__pdo_in_tx"] = values.NewBool(false)
	obj.Properties["__pdo_tx"] = values.NewNull()

	return values.NewBool(true), nil
}

// pdoInTransaction implements $pdo->inTransaction()
// args[0] = $this
func pdoInTransaction(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	thisObj := args[0]

	obj := thisObj.Data.(*values.Object)
	inTx := obj.Properties["__pdo_in_tx"]

	if inTx == nil {
		return values.NewBool(false), nil
	}

	return values.NewBool(inTx.Data.(bool)), nil
}

// Placeholder implementations for remaining methods
// pdoGetAttribute implements $pdo->getAttribute($attribute)
// args[0] = $this, args[1] = attribute
func pdoGetAttribute(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 2 {
		return values.NewNull(), fmt.Errorf("PDO::getAttribute() expects 1 parameter")
	}

	thisObj := args[0]
	attribute := args[1].Data.(int64)

	obj := thisObj.Data.(*values.Object)
	attributesVal, ok := obj.Properties["__pdo_attributes"]
	if !ok || attributesVal.Type != values.TypeArray {
		return values.NewNull(), nil
	}

	attributes := attributesVal.Data.(*values.Array)

	// Look up the attribute value
	if val, exists := attributes.Elements[attribute]; exists {
		return val, nil
	}

	// Special attributes that return driver-specific info
	switch attribute {
	case 4: // PDO::ATTR_SERVER_VERSION
		// Return a placeholder version string
		return values.NewString("1.0.0"), nil
	case 5: // PDO::ATTR_CLIENT_VERSION
		return values.NewString("hey-codex-1.0"), nil
	case 6: // PDO::ATTR_SERVER_INFO
		driverVal := obj.Properties["__pdo_driver"]
		if driverVal != nil && driverVal.Type == values.TypeString {
			return values.NewString(driverVal.Data.(string) + " driver"), nil
		}
		return values.NewString("PDO driver"), nil
	}

	return values.NewNull(), nil
}

// pdoSetAttribute implements $pdo->setAttribute($attribute, $value)
// args[0] = $this, args[1] = attribute, args[2] = value
func pdoSetAttribute(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 3 {
		return values.NewBool(false), fmt.Errorf("PDO::setAttribute() expects 2 parameters")
	}

	thisObj := args[0]
	attribute := args[1].Data.(int64)
	value := args[2]

	obj := thisObj.Data.(*values.Object)
	attributesVal, ok := obj.Properties["__pdo_attributes"]
	if !ok {
		// Initialize if not present
		attributesVal = values.NewArray()
		obj.Properties["__pdo_attributes"] = attributesVal
	}

	if attributesVal.Type != values.TypeArray {
		return values.NewBool(false), nil
	}

	// Validate certain attributes
	switch attribute {
	case 3: // PDO::ATTR_ERRMODE
		// Must be 0 (SILENT), 1 (WARNING), or 2 (EXCEPTION)
		if value.Type == values.TypeInt {
			intVal := value.Data.(int64)
			if intVal < 0 || intVal > 2 {
				// Invalid value, return false
				return values.NewBool(false), nil
			}
		}
	case 19: // PDO::ATTR_DEFAULT_FETCH_MODE
		// Must be valid fetch mode
		if value.Type == values.TypeInt {
			intVal := value.Data.(int64)
			if intVal < 1 || intVal > 12 {
				// Invalid value, return false
				return values.NewBool(false), nil
			}
		}
	}

	// Set the attribute
	attributesVal.ArraySet(values.NewInt(attribute), value)

	return values.NewBool(true), nil
}

// pdoErrorCode implements $pdo->errorCode()
// args[0] = $this
func pdoErrorCode(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	thisObj := args[0]
	obj := thisObj.Data.(*values.Object)

	errorCode, ok := obj.Properties["__pdo_error_code"]
	if !ok || errorCode.Type != values.TypeString {
		return values.NewNull(), nil
	}

	return errorCode, nil
}

// pdoErrorInfo implements $pdo->errorInfo()
// args[0] = $this
func pdoErrorInfo(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
	thisObj := args[0]
	obj := thisObj.Data.(*values.Object)

	errorInfo, ok := obj.Properties["__pdo_error_info"]
	if !ok || errorInfo.Type == values.TypeNull {
		// Return default error info array [SQLSTATE, driver_code, message]
		arr := values.NewArray()
		arr.ArraySet(values.NewInt(0), values.NewString("00000"))
		arr.ArraySet(values.NewInt(1), values.NewNull())
		arr.ArraySet(values.NewInt(2), values.NewNull())
		return arr, nil
	}

	return errorInfo, nil
}
