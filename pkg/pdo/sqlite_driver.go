package pdo

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
	"github.com/wudi/hey/values"
)

// SQLiteDriver implements the Driver interface for SQLite
type SQLiteDriver struct{}

// Open creates a new SQLite connection
func (d *SQLiteDriver) Open(dsn string) (Conn, error) {
	// Parse PDO DSN
	dsnInfo, err := ParseDSN(dsn)
	if err != nil {
		return nil, err
	}

	// Build SQLite DSN (just the file path or :memory:)
	sqliteDSN := BuildSQLiteDSN(dsnInfo)

	return &SQLiteConn{
		dsnInfo: dsnInfo,
		dsn:     sqliteDSN,
		db:      nil, // Will be set by Connect()
	}, nil
}

// Name returns the driver name
func (d *SQLiteDriver) Name() string {
	return "sqlite"
}

// SQLiteConn implements the Conn interface for SQLite
type SQLiteConn struct {
	dsnInfo      *DSN
	dsn          string
	db           *sql.DB
	lastInsertId int64
}

// Connect establishes the actual database connection
func (c *SQLiteConn) Connect() error {
	db, err := sql.Open("sqlite", c.dsn)
	if err != nil {
		return NewPDOError("HY000", 1, fmt.Sprintf("Failed to connect: %v", err))
	}

	// Verify connection
	if err := db.Ping(); err != nil {
		db.Close()
		return NewPDOError("HY000", 1, fmt.Sprintf("Failed to ping: %v", err))
	}

	c.db = db
	return nil
}

// Prepare creates a prepared statement
func (c *SQLiteConn) Prepare(query string) (Stmt, error) {
	if c.db == nil {
		return nil, NewPDOError("HY000", 1, "Database not connected")
	}

	stmt, err := c.db.Prepare(query)
	if err != nil {
		return nil, NewPDOError("HY000", 1, fmt.Sprintf("Prepare error: %v", err))
	}

	return &SQLiteStmt{
		stmt:       stmt,
		query:      query,
		params:     make(map[interface{}]*values.Value),
		paramOrder: []interface{}{},
	}, nil
}

// Query executes a query that returns rows
func (c *SQLiteConn) Query(query string) (Rows, error) {
	if c.db == nil {
		return nil, NewPDOError("HY000", 1, "Database not connected")
	}

	rows, err := c.db.Query(query)
	if err != nil {
		return nil, NewPDOError("HY000", 1, fmt.Sprintf("Query error: %v", err))
	}

	return &SQLiteRows{rows: rows}, nil
}

// Exec executes a query that doesn't return rows
func (c *SQLiteConn) Exec(query string) (Result, error) {
	if c.db == nil {
		return nil, NewPDOError("HY000", 1, "Database not connected")
	}

	result, err := c.db.Exec(query)
	if err != nil {
		return nil, NewPDOError("HY000", 1, fmt.Sprintf("Exec error: %v", err))
	}

	// Store last insert ID
	if lastID, err := result.LastInsertId(); err == nil {
		c.lastInsertId = lastID
	}

	return result, nil
}

// Begin starts a transaction
func (c *SQLiteConn) Begin() (Tx, error) {
	if c.db == nil {
		return nil, NewPDOError("HY000", 1, "Database not connected")
	}

	tx, err := c.db.Begin()
	if err != nil {
		return nil, NewPDOError("HY000", 1, fmt.Sprintf("Begin error: %v", err))
	}

	return &SQLiteTx{tx: tx}, nil
}

// Close closes the connection
func (c *SQLiteConn) Close() error {
	if c.db == nil {
		return nil
	}
	return c.db.Close()
}

// Ping verifies the connection is alive
func (c *SQLiteConn) Ping() error {
	if c.db == nil {
		return NewPDOError("HY000", 1, "Database not connected")
	}
	return c.db.Ping()
}

// LastInsertId returns the last inserted ID
func (c *SQLiteConn) LastInsertId() (int64, error) {
	return c.lastInsertId, nil
}

// GetUnderlyingDB returns the underlying *sql.DB
func (c *SQLiteConn) GetUnderlyingDB() *sql.DB {
	return c.db
}

// SQLiteStmt implements the Stmt interface for SQLite
type SQLiteStmt struct {
	stmt       *sql.Stmt
	query      string
	params     map[interface{}]*values.Value
	paramOrder []interface{}
	rowCount   int64
}

// BindValue binds a value to a parameter
func (s *SQLiteStmt) BindValue(param interface{}, value *values.Value, dataType int) error {
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
func (s *SQLiteStmt) Execute() (Result, error) {
	args, err := s.buildArgs()
	if err != nil {
		return nil, err
	}

	result, err := s.stmt.Exec(args...)
	if err != nil {
		return nil, NewPDOError("HY000", 1, fmt.Sprintf("Execute error: %v", err))
	}

	// Update row count
	if affected, err := result.RowsAffected(); err == nil {
		s.rowCount = affected
	}

	return result, nil
}

// Query executes a prepared statement and returns rows
func (s *SQLiteStmt) Query() (Rows, error) {
	args, err := s.buildArgs()
	if err != nil {
		return nil, err
	}

	rows, err := s.stmt.Query(args...)
	if err != nil {
		return nil, NewPDOError("HY000", 1, fmt.Sprintf("Query error: %v", err))
	}

	return &SQLiteRows{rows: rows}, nil
}

// Close closes the statement
func (s *SQLiteStmt) Close() error {
	return s.stmt.Close()
}

// RowCount returns the number of rows affected
func (s *SQLiteStmt) RowCount() int64 {
	return s.rowCount
}

// buildArgs builds the argument slice from bound parameters
func (s *SQLiteStmt) buildArgs() ([]interface{}, error) {
	// Count placeholders in query
	placeholderCount := 0
	for _, c := range s.query {
		if c == '?' {
			placeholderCount++
		}
	}

	args := make([]interface{}, placeholderCount)

	// Fill args in order
	for i := 0; i < placeholderCount; i++ {
		param := i + 1 // PDO uses 1-based indexing
		val, ok := s.params[param]
		if !ok {
			return nil, NewPDOError("HY000", 1, fmt.Sprintf("Missing parameter: %d", param))
		}
		args[i] = convertValueToInterface(val)
	}

	return args, nil
}

// SQLiteRows implements the Rows interface for SQLite
type SQLiteRows struct {
	rows    *sql.Rows
	columns []string
}

// Next advances to the next row
func (r *SQLiteRows) Next() bool {
	return r.rows.Next()
}

// Scan scans the current row
func (r *SQLiteRows) Scan(dest ...interface{}) error {
	return r.rows.Scan(dest...)
}

// Columns returns column names
func (r *SQLiteRows) Columns() ([]string, error) {
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
func (r *SQLiteRows) Close() error {
	return r.rows.Close()
}

// Err returns any error encountered
func (r *SQLiteRows) Err() error {
	return r.rows.Err()
}

// FetchAssoc fetches the next row as associative array
func (r *SQLiteRows) FetchAssoc() (map[string]*values.Value, error) {
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
func (r *SQLiteRows) FetchNum() ([]*values.Value, error) {
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
func (r *SQLiteRows) FetchBoth() (map[string]*values.Value, []*values.Value, error) {
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

// SQLiteTx implements the Tx interface for SQLite
type SQLiteTx struct {
	tx *sql.Tx
}

// Commit commits the transaction
func (t *SQLiteTx) Commit() error {
	return t.tx.Commit()
}

// Rollback rolls back the transaction
func (t *SQLiteTx) Rollback() error {
	return t.tx.Rollback()
}

// Prepare creates a prepared statement in transaction context
func (t *SQLiteTx) Prepare(query string) (Stmt, error) {
	stmt, err := t.tx.Prepare(query)
	if err != nil {
		return nil, NewPDOError("HY000", 1, fmt.Sprintf("Prepare error: %v", err))
	}

	return &SQLiteStmt{
		stmt:       stmt,
		query:      query,
		params:     make(map[interface{}]*values.Value),
		paramOrder: []interface{}{},
	}, nil
}

// Query executes a query in transaction context
func (t *SQLiteTx) Query(query string) (Rows, error) {
	rows, err := t.tx.Query(query)
	if err != nil {
		return nil, NewPDOError("HY000", 1, fmt.Sprintf("Query error: %v", err))
	}

	return &SQLiteRows{rows: rows}, nil
}

// Exec executes a statement in transaction context
func (t *SQLiteTx) Exec(query string) (Result, error) {
	result, err := t.tx.Exec(query)
	if err != nil {
		return nil, NewPDOError("HY000", 1, fmt.Sprintf("Exec error: %v", err))
	}

	return result, nil
}

// Register SQLite driver on package initialization
func init() {
	RegisterDriver("sqlite", &SQLiteDriver{})
}
