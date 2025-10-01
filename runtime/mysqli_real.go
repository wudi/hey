package runtime

import (
	"database/sql"
	"fmt"
	"sync"

	_ "github.com/go-sql-driver/mysql"
	"github.com/wudi/hey/values"
)

// MySQLi real database connection pool
var (
	connectionPool  = make(map[*MySQLiConnection]*sql.DB)
	poolMutex       sync.RWMutex
	lastConnectErr  error
	lastConnectErrNo int
)

// RealMySQLiConnect establishes a real MySQL connection
func RealMySQLiConnect(conn *MySQLiConnection) error {
	// Build DSN: user:password@tcp(host:port)/database
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		conn.Username,
		conn.Password,
		conn.Host,
		conn.Port,
		conn.Database,
	)

	// Open database connection
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		conn.Connected = false
		conn.ErrorNo = 2002
		conn.Error = fmt.Sprintf("Failed to connect: %v", err)
		lastConnectErr = err
		lastConnectErrNo = 2002
		return err
	}

	// Verify connection
	if err := db.Ping(); err != nil {
		conn.Connected = false
		conn.ErrorNo = 2002
		conn.Error = fmt.Sprintf("Failed to ping: %v", err)
		lastConnectErr = err
		lastConnectErrNo = 2002
		db.Close()
		return err
	}

	// Store connection in pool
	poolMutex.Lock()
	connectionPool[conn] = db
	poolMutex.Unlock()

	conn.Connected = true
	conn.ErrorNo = 0
	conn.Error = ""
	lastConnectErr = nil
	lastConnectErrNo = 0
	return nil
}

// GetLastConnectError returns the last connection error message
func GetLastConnectError() string {
	if lastConnectErr != nil {
		return lastConnectErr.Error()
	}
	return ""
}

// GetLastConnectErrno returns the last connection error number
func GetLastConnectErrno() int {
	return lastConnectErrNo
}

// GetRealConnection retrieves the real SQL connection from pool
func GetRealConnection(conn *MySQLiConnection) (*sql.DB, bool) {
	poolMutex.RLock()
	defer poolMutex.RUnlock()
	db, ok := connectionPool[conn]
	return db, ok
}

// RealMySQLiClose closes a real MySQL connection
func RealMySQLiClose(conn *MySQLiConnection) error {
	poolMutex.Lock()
	defer poolMutex.Unlock()

	db, ok := connectionPool[conn]
	if !ok {
		return nil
	}

	err := db.Close()
	delete(connectionPool, conn)
	conn.Connected = false
	return err
}

// RealMySQLiQuery executes a SQL query and returns a result
func RealMySQLiQuery(conn *MySQLiConnection, query string) (*MySQLiResult, error) {
	db, ok := GetRealConnection(conn)
	if !ok {
		conn.ErrorNo = 2006
		conn.Error = "MySQL server has gone away"
		return nil, fmt.Errorf("no active connection")
	}

	// Execute query
	rows, err := db.Query(query)
	if err != nil {
		conn.ErrorNo = 1064
		conn.Error = fmt.Sprintf("Query error: %v", err)
		return nil, err
	}
	defer rows.Close()

	// Get column information
	columns, err := rows.Columns()
	if err != nil {
		conn.ErrorNo = 1054
		conn.Error = fmt.Sprintf("Column error: %v", err)
		return nil, err
	}

	// Get column types
	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		conn.ErrorNo = 1054
		conn.Error = fmt.Sprintf("Column type error: %v", err)
		return nil, err
	}

	// Create result structure
	result := &MySQLiResult{
		FieldCount: len(columns),
		CurrentRow: 0,
		Rows:       make([]map[string]*values.Value, 0),
		Fields:     make([]MySQLiField, len(columns)),
	}

	// Populate field metadata
	for i, col := range columns {
		colType := columnTypes[i]
		result.Fields[i] = MySQLiField{
			Name:     col,
			OrgName:  col,
			Table:    colType.Name(),
			Database: conn.Database,
			Type:     mapGoTypeToMySQLiType(colType.DatabaseTypeName()),
		}
	}

	// Fetch all rows
	for rows.Next() {
		// Create slice to hold column values
		scanValues := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range scanValues {
			valuePtrs[i] = &scanValues[i]
		}

		// Scan row into value pointers
		if err := rows.Scan(valuePtrs...); err != nil {
			conn.ErrorNo = 1054
			conn.Error = fmt.Sprintf("Scan error: %v", err)
			return nil, err
		}

		// Create row map with values.Value
		row := make(map[string]*values.Value)
		for i, col := range columns {
			val := scanValues[i]
			if val == nil {
				row[col] = values.NewNull()
			} else if b, ok := val.([]byte); ok {
				row[col] = values.NewString(string(b))
			} else if intVal, ok := val.(int64); ok {
				row[col] = values.NewInt(intVal)
			} else if floatVal, ok := val.(float64); ok {
				row[col] = values.NewFloat(floatVal)
			} else if strVal, ok := val.(string); ok {
				row[col] = values.NewString(strVal)
			} else {
				row[col] = values.NewString(fmt.Sprintf("%v", val))
			}
		}

		result.Rows = append(result.Rows, row)
	}

	// Update result metadata
	result.NumRows = int64(len(result.Rows))
	conn.FieldCount = result.FieldCount
	conn.ErrorNo = 0
	conn.Error = ""

	return result, nil
}

// RealMySQLiExecute executes a non-query SQL statement (INSERT, UPDATE, DELETE)
func RealMySQLiExecute(conn *MySQLiConnection, query string) (int64, error) {
	db, ok := GetRealConnection(conn)
	if !ok {
		conn.ErrorNo = 2006
		conn.Error = "MySQL server has gone away"
		return 0, fmt.Errorf("no active connection")
	}

	// Execute statement
	result, err := db.Exec(query)
	if err != nil {
		conn.ErrorNo = 1064
		conn.Error = fmt.Sprintf("Execute error: %v", err)
		return 0, err
	}

	// Get affected rows
	affectedRows, err := result.RowsAffected()
	if err != nil {
		affectedRows = 0
	}
	conn.AffectedRows = affectedRows

	// Get last insert ID
	lastID, err := result.LastInsertId()
	if err != nil {
		lastID = 0
	}
	conn.InsertID = lastID

	conn.ErrorNo = 0
	conn.Error = ""

	return affectedRows, nil
}

// mapGoTypeToMySQLiType maps Go SQL types to MySQLi type constants
func mapGoTypeToMySQLiType(typeName string) int {
	switch typeName {
	case "TINYINT":
		return 1 // MYSQLI_TYPE_TINY
	case "SMALLINT":
		return 2 // MYSQLI_TYPE_SHORT
	case "INT", "INTEGER":
		return 3 // MYSQLI_TYPE_LONG
	case "BIGINT":
		return 8 // MYSQLI_TYPE_LONGLONG
	case "FLOAT":
		return 4 // MYSQLI_TYPE_FLOAT
	case "DOUBLE", "REAL":
		return 5 // MYSQLI_TYPE_DOUBLE
	case "DECIMAL", "NUMERIC":
		return 246 // MYSQLI_TYPE_DECIMAL
	case "DATE":
		return 10 // MYSQLI_TYPE_DATE
	case "TIME":
		return 11 // MYSQLI_TYPE_TIME
	case "DATETIME":
		return 12 // MYSQLI_TYPE_DATETIME
	case "TIMESTAMP":
		return 7 // MYSQLI_TYPE_TIMESTAMP
	case "YEAR":
		return 13 // MYSQLI_TYPE_YEAR
	case "VARCHAR", "CHAR":
		return 253 // MYSQLI_TYPE_VAR_STRING
	case "TEXT", "TINYTEXT", "MEDIUMTEXT", "LONGTEXT":
		return 252 // MYSQLI_TYPE_BLOB
	case "BLOB", "TINYBLOB", "MEDIUMBLOB", "LONGBLOB":
		return 252 // MYSQLI_TYPE_BLOB
	case "JSON":
		return 245 // MYSQLI_TYPE_JSON
	default:
		return 253 // MYSQLI_TYPE_VAR_STRING (default)
	}
}
