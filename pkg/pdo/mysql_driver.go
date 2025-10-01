package pdo

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/wudi/hey/values"
)

// MySQLDriver implements the Driver interface for MySQL
type MySQLDriver struct{}

// Open creates a new MySQL connection
func (d *MySQLDriver) Open(dsn string) (Conn, error) {
	// Parse PDO DSN
	dsnInfo, err := ParseDSN(dsn)
	if err != nil {
		return nil, err
	}

	// Build MySQL-specific DSN (handled separately with username/password)
	// This will be set when calling PDO constructor with credentials
	return &MySQLConn{
		dsnInfo: dsnInfo,
		db:      nil, // Will be set by Connect()
	}, nil
}

// Name returns the driver name
func (d *MySQLDriver) Name() string {
	return "mysql"
}

// MySQLConn implements the Conn interface for MySQL
type MySQLConn struct {
	dsnInfo  *DSN
	db       *sql.DB
	lastInsertId int64
	username string
	password string
}

// Connect establishes the actual database connection
func (c *MySQLConn) Connect(username, password string) error {
	c.username = username
	c.password = password

	dsn := BuildMySQLDSN(c.dsnInfo, username, password)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return NewPDOError("HY000", 2002, fmt.Sprintf("Failed to connect: %v", err))
	}

	// Verify connection
	if err := db.Ping(); err != nil {
		db.Close()
		return NewPDOError("HY000", 2002, fmt.Sprintf("Failed to ping: %v", err))
	}

	c.db = db
	return nil
}

// Prepare creates a prepared statement
func (c *MySQLConn) Prepare(query string) (Stmt, error) {
	if c.db == nil {
		return nil, NewPDOError("HY000", 2006, "MySQL server has gone away")
	}

	stmt, err := c.db.Prepare(query)
	if err != nil {
		return nil, NewPDOError("HY000", 1064, fmt.Sprintf("Prepare error: %v", err))
	}

	return &MySQLStmt{
		stmt:        stmt,
		query:       query,
		params:      make(map[interface{}]*values.Value),
		paramOrder:  []interface{}{},
	}, nil
}

// Query executes a query that returns rows
func (c *MySQLConn) Query(query string) (Rows, error) {
	if c.db == nil {
		return nil, NewPDOError("HY000", 2006, "MySQL server has gone away")
	}

	rows, err := c.db.Query(query)
	if err != nil {
		return nil, NewPDOError("HY000", 1064, fmt.Sprintf("Query error: %v", err))
	}

	return &MySQLRows{rows: rows}, nil
}

// Exec executes a query that doesn't return rows
func (c *MySQLConn) Exec(query string) (Result, error) {
	if c.db == nil {
		return nil, NewPDOError("HY000", 2006, "MySQL server has gone away")
	}

	result, err := c.db.Exec(query)
	if err != nil {
		return nil, NewPDOError("HY000", 1064, fmt.Sprintf("Exec error: %v", err))
	}

	// Store last insert ID
	if lastID, err := result.LastInsertId(); err == nil {
		c.lastInsertId = lastID
	}

	return result, nil
}

// Begin starts a transaction
func (c *MySQLConn) Begin() (Tx, error) {
	if c.db == nil {
		return nil, NewPDOError("HY000", 2006, "MySQL server has gone away")
	}

	tx, err := c.db.Begin()
	if err != nil {
		return nil, NewPDOError("HY000", 1064, fmt.Sprintf("Begin error: %v", err))
	}

	return &MySQLTx{tx: tx}, nil
}

// Close closes the connection
func (c *MySQLConn) Close() error {
	if c.db == nil {
		return nil
	}
	return c.db.Close()
}

// Ping verifies the connection is alive
func (c *MySQLConn) Ping() error {
	if c.db == nil {
		return NewPDOError("HY000", 2006, "MySQL server has gone away")
	}
	return c.db.Ping()
}

// LastInsertId returns the last inserted ID
func (c *MySQLConn) LastInsertId() (int64, error) {
	return c.lastInsertId, nil
}

// GetUnderlyingDB returns the underlying *sql.DB
func (c *MySQLConn) GetUnderlyingDB() *sql.DB {
	return c.db
}

// MySQLStmt implements the Stmt interface for MySQL
type MySQLStmt struct {
	stmt       *sql.Stmt
	query      string
	params     map[interface{}]*values.Value
	paramOrder []interface{}
	rowCount   int64
}

// BindValue binds a value to a parameter
func (s *MySQLStmt) BindValue(param interface{}, value *values.Value, dataType int) error {
	s.params[param] = value

	// Track parameter order for positional binding
	found := false
	for _, p := range s.paramOrder {
		if p == param {
			found = true
			break
		}
	}
	if !found {
		s.paramOrder = append(s.paramOrder, param)
	}

	return nil
}

// Execute executes a prepared statement
func (s *MySQLStmt) Execute() (Result, error) {
	args, err := s.buildArgs()
	if err != nil {
		return nil, err
	}

	result, err := s.stmt.Exec(args...)
	if err != nil {
		return nil, NewPDOError("HY000", 1064, fmt.Sprintf("Execute error: %v", err))
	}

	// Update row count
	if affected, err := result.RowsAffected(); err == nil {
		s.rowCount = affected
	}

	return result, nil
}

// Query executes a prepared statement and returns rows
func (s *MySQLStmt) Query() (Rows, error) {
	args, err := s.buildArgs()
	if err != nil {
		return nil, err
	}

	rows, err := s.stmt.Query(args...)
	if err != nil {
		return nil, NewPDOError("HY000", 1064, fmt.Sprintf("Query error: %v", err))
	}

	return &MySQLRows{rows: rows}, nil
}

// Close closes the statement
func (s *MySQLStmt) Close() error {
	return s.stmt.Close()
}

// RowCount returns the number of rows affected
func (s *MySQLStmt) RowCount() int64 {
	return s.rowCount
}

// buildArgs builds the argument slice from bound parameters
func (s *MySQLStmt) buildArgs() ([]interface{}, error) {
	// Count placeholders in query
	placeholderCount := strings.Count(s.query, "?")

	args := make([]interface{}, placeholderCount)

	// Fill args in order
	for i := 0; i < placeholderCount; i++ {
		param := i + 1 // PDO uses 1-based indexing
		val, ok := s.params[param]
		if !ok {
			return nil, NewPDOError("HY000", 1064, fmt.Sprintf("Missing parameter: %d", param))
		}
		args[i] = convertValueToInterface(val)
	}

	return args, nil
}

// MySQLRows implements the Rows interface for MySQL
type MySQLRows struct {
	rows    *sql.Rows
	columns []string
}

// Next advances to the next row
func (r *MySQLRows) Next() bool {
	return r.rows.Next()
}

// Scan scans the current row
func (r *MySQLRows) Scan(dest ...interface{}) error {
	return r.rows.Scan(dest...)
}

// Columns returns column names
func (r *MySQLRows) Columns() ([]string, error) {
	if r.columns == nil {
		cols, err := r.rows.Columns()
		if err != nil {
			return nil, err
		}
		r.columns = cols
	}
	return r.columns, nil
}

// Close closes the rows
func (r *MySQLRows) Close() error {
	return r.rows.Close()
}

// Err returns any error encountered
func (r *MySQLRows) Err() error {
	return r.rows.Err()
}

// FetchAssoc fetches the next row as associative array
func (r *MySQLRows) FetchAssoc() (map[string]*values.Value, error) {
	if !r.Next() {
		return nil, nil
	}

	columns, err := r.Columns()
	if err != nil {
		return nil, err
	}

	rowValues := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range rowValues {
		valuePtrs[i] = &rowValues[i]
	}

	if err := r.Scan(valuePtrs...); err != nil {
		return nil, err
	}

	result := make(map[string]*values.Value)
	for i, col := range columns {
		result[col] = convertInterfaceToValue(rowValues[i])
	}

	return result, nil
}

// FetchNum fetches the next row as numeric array
func (r *MySQLRows) FetchNum() ([]*values.Value, error) {
	if !r.Next() {
		return nil, nil
	}

	columns, err := r.Columns()
	if err != nil {
		return nil, err
	}

	rowValues := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range rowValues {
		valuePtrs[i] = &rowValues[i]
	}

	if err := r.Scan(valuePtrs...); err != nil {
		return nil, err
	}

	result := make([]*values.Value, len(rowValues))
	for i := range rowValues {
		result[i] = convertInterfaceToValue(rowValues[i])
	}

	return result, nil
}

// FetchBoth fetches the next row as both associative and numeric
func (r *MySQLRows) FetchBoth() (map[string]*values.Value, []*values.Value, error) {
	if !r.Next() {
		return nil, nil, nil
	}

	columns, err := r.Columns()
	if err != nil {
		return nil, nil, err
	}

	rowValues := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range rowValues {
		valuePtrs[i] = &rowValues[i]
	}

	if err := r.Scan(valuePtrs...); err != nil {
		return nil, nil, err
	}

	assoc := make(map[string]*values.Value)
	numeric := make([]*values.Value, len(rowValues))

	for i, col := range columns {
		val := convertInterfaceToValue(rowValues[i])
		assoc[col] = val
		numeric[i] = val
	}

	return assoc, numeric, nil
}

// MySQLTx implements the Tx interface for MySQL
type MySQLTx struct {
	tx *sql.Tx
}

// Commit commits the transaction
func (t *MySQLTx) Commit() error {
	return t.tx.Commit()
}

// Rollback rolls back the transaction
func (t *MySQLTx) Rollback() error {
	return t.tx.Rollback()
}

// Prepare creates a prepared statement in transaction context
func (t *MySQLTx) Prepare(query string) (Stmt, error) {
	stmt, err := t.tx.Prepare(query)
	if err != nil {
		return nil, NewPDOError("HY000", 1064, fmt.Sprintf("Prepare error: %v", err))
	}

	return &MySQLStmt{
		stmt:       stmt,
		query:      query,
		params:     make(map[interface{}]*values.Value),
		paramOrder: []interface{}{},
	}, nil
}

// Query executes a query in transaction context
func (t *MySQLTx) Query(query string) (Rows, error) {
	rows, err := t.tx.Query(query)
	if err != nil {
		return nil, NewPDOError("HY000", 1064, fmt.Sprintf("Query error: %v", err))
	}

	return &MySQLRows{rows: rows}, nil
}

// Exec executes a statement in transaction context
func (t *MySQLTx) Exec(query string) (Result, error) {
	result, err := t.tx.Exec(query)
	if err != nil {
		return nil, NewPDOError("HY000", 1064, fmt.Sprintf("Exec error: %v", err))
	}

	return result, nil
}

// Helper functions for value conversion

func convertValueToInterface(v *values.Value) interface{} {
	switch v.Type {
	case values.TypeNull:
		return nil
	case values.TypeInt:
		return v.Data.(int64)
	case values.TypeFloat:
		return v.Data.(float64)
	case values.TypeString:
		return v.Data.(string)
	case values.TypeBool:
		if v.Data.(bool) {
			return int64(1)
		}
		return int64(0)
	default:
		return fmt.Sprintf("%v", v.Data)
	}
}

func convertInterfaceToValue(i interface{}) *values.Value {
	if i == nil {
		return values.NewNull()
	}

	switch v := i.(type) {
	case int64:
		return values.NewInt(v)
	case float64:
		return values.NewFloat(v)
	case []byte:
		return values.NewString(string(v))
	case string:
		return values.NewString(v)
	case bool:
		return values.NewBool(v)
	default:
		return values.NewString(fmt.Sprintf("%v", i))
	}
}

// Register MySQL driver on package initialization
func init() {
	RegisterDriver("mysql", &MySQLDriver{})
}
