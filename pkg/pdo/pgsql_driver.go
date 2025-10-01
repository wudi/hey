package pdo

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/lib/pq"
	"github.com/wudi/hey/values"
)

// PgSQLDriver implements the Driver interface for PostgreSQL
type PgSQLDriver struct{}

// Open creates a new PostgreSQL connection
func (d *PgSQLDriver) Open(dsn string) (Conn, error) {
	// Parse PDO DSN
	dsnInfo, err := ParseDSN(dsn)
	if err != nil {
		return nil, err
	}

	return &PgSQLConn{
		dsnInfo: dsnInfo,
		db:      nil, // Will be set by Connect()
	}, nil
}

// Name returns the driver name
func (d *PgSQLDriver) Name() string {
	return "pgsql"
}

// PgSQLConn implements the Conn interface for PostgreSQL
type PgSQLConn struct {
	dsnInfo      *DSN
	db           *sql.DB
	lastInsertId int64
	username     string
	password     string
}

// Connect establishes the actual database connection
func (c *PgSQLConn) Connect(username, password string) error {
	c.username = username
	c.password = password

	dsn := BuildPostgreSQLDSN(c.dsnInfo, username, password)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return NewPDOError("HY000", 7, fmt.Sprintf("Failed to connect: %v", err))
	}

	// Verify connection
	if err := db.Ping(); err != nil {
		db.Close()
		return NewPDOError("HY000", 7, fmt.Sprintf("Failed to ping: %v", err))
	}

	c.db = db
	return nil
}

// Prepare creates a prepared statement
func (c *PgSQLConn) Prepare(query string) (Stmt, error) {
	if c.db == nil {
		return nil, NewPDOError("HY000", 8, "Database not connected")
	}

	// Convert ? placeholders to $1, $2, etc for PostgreSQL
	pgQuery := convertPlaceholders(query)

	stmt, err := c.db.Prepare(pgQuery)
	if err != nil {
		return nil, NewPDOError("42601", 1, fmt.Sprintf("Prepare error: %v", err))
	}

	return &PgSQLStmt{
		stmt:       stmt,
		query:      pgQuery,
		params:     make(map[interface{}]*values.Value),
		paramOrder: []interface{}{},
	}, nil
}

// Query executes a query that returns rows
func (c *PgSQLConn) Query(query string) (Rows, error) {
	if c.db == nil {
		return nil, NewPDOError("HY000", 8, "Database not connected")
	}

	rows, err := c.db.Query(query)
	if err != nil {
		return nil, NewPDOError("42601", 1, fmt.Sprintf("Query error: %v", err))
	}

	return &PgSQLRows{rows: rows}, nil
}

// Exec executes a query that doesn't return rows
func (c *PgSQLConn) Exec(query string) (Result, error) {
	if c.db == nil {
		return nil, NewPDOError("HY000", 8, "Database not connected")
	}

	result, err := c.db.Exec(query)
	if err != nil {
		return nil, NewPDOError("42601", 1, fmt.Sprintf("Exec error: %v", err))
	}

	// PostgreSQL doesn't support LastInsertId directly
	// Would need RETURNING clause for this
	return result, nil
}

// Begin starts a transaction
func (c *PgSQLConn) Begin() (Tx, error) {
	if c.db == nil {
		return nil, NewPDOError("HY000", 8, "Database not connected")
	}

	tx, err := c.db.Begin()
	if err != nil {
		return nil, NewPDOError("HY000", 1, fmt.Sprintf("Begin error: %v", err))
	}

	return &PgSQLTx{tx: tx}, nil
}

// Close closes the connection
func (c *PgSQLConn) Close() error {
	if c.db == nil {
		return nil
	}
	return c.db.Close()
}

// Ping verifies the connection is alive
func (c *PgSQLConn) Ping() error {
	if c.db == nil {
		return NewPDOError("HY000", 8, "Database not connected")
	}
	return c.db.Ping()
}

// LastInsertId returns the last inserted ID
func (c *PgSQLConn) LastInsertId() (int64, error) {
	return c.lastInsertId, nil
}

// GetUnderlyingDB returns the underlying *sql.DB
func (c *PgSQLConn) GetUnderlyingDB() *sql.DB {
	return c.db
}

// convertPlaceholders converts ? placeholders to PostgreSQL $1, $2, etc.
func convertPlaceholders(query string) string {
	count := 1
	result := strings.Builder{}

	for _, ch := range query {
		if ch == '?' {
			result.WriteString(fmt.Sprintf("$%d", count))
			count++
		} else {
			result.WriteRune(ch)
		}
	}

	return result.String()
}

// PgSQLStmt implements the Stmt interface for PostgreSQL
type PgSQLStmt struct {
	stmt       *sql.Stmt
	query      string
	params     map[interface{}]*values.Value
	paramOrder []interface{}
	rowCount   int64
}

// BindValue binds a value to a parameter
func (s *PgSQLStmt) BindValue(param interface{}, value *values.Value, dataType int) error {
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
func (s *PgSQLStmt) Execute() (Result, error) {
	args, err := s.buildArgs()
	if err != nil {
		return nil, err
	}

	result, err := s.stmt.Exec(args...)
	if err != nil {
		return nil, NewPDOError("42601", 1, fmt.Sprintf("Execute error: %v", err))
	}

	// Update row count
	if affected, err := result.RowsAffected(); err == nil {
		s.rowCount = affected
	}

	return result, nil
}

// Query executes a prepared statement and returns rows
func (s *PgSQLStmt) Query() (Rows, error) {
	args, err := s.buildArgs()
	if err != nil {
		return nil, err
	}

	rows, err := s.stmt.Query(args...)
	if err != nil {
		return nil, NewPDOError("42601", 1, fmt.Sprintf("Query error: %v", err))
	}

	return &PgSQLRows{rows: rows}, nil
}

// Close closes the statement
func (s *PgSQLStmt) Close() error {
	return s.stmt.Close()
}

// RowCount returns the number of rows affected
func (s *PgSQLStmt) RowCount() int64 {
	return s.rowCount
}

// buildArgs builds the argument slice from bound parameters
func (s *PgSQLStmt) buildArgs() ([]interface{}, error) {
	// Count $N placeholders in query
	count := 0
	for i := 0; i < len(s.query); i++ {
		if s.query[i] == '$' {
			count++
		}
	}

	args := make([]interface{}, count)

	// Fill args in order
	for i := 0; i < count; i++ {
		param := i + 1 // PDO uses 1-based indexing
		val, ok := s.params[param]
		if !ok {
			return nil, NewPDOError("HY000", 1, fmt.Sprintf("Missing parameter: %d", param))
		}
		args[i] = convertValueToInterface(val)
	}

	return args, nil
}

// PgSQLRows implements the Rows interface for PostgreSQL
type PgSQLRows struct {
	rows    *sql.Rows
	columns []string
}

// Next advances to the next row
func (r *PgSQLRows) Next() bool {
	return r.rows.Next()
}

// Scan scans the current row
func (r *PgSQLRows) Scan(dest ...interface{}) error {
	return r.rows.Scan(dest...)
}

// Columns returns column names
func (r *PgSQLRows) Columns() ([]string, error) {
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
func (r *PgSQLRows) Close() error {
	return r.rows.Close()
}

// Err returns any error encountered
func (r *PgSQLRows) Err() error {
	return r.rows.Err()
}

// FetchAssoc fetches the next row as associative array
func (r *PgSQLRows) FetchAssoc() (map[string]*values.Value, error) {
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
func (r *PgSQLRows) FetchNum() ([]*values.Value, error) {
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
func (r *PgSQLRows) FetchBoth() (map[string]*values.Value, []*values.Value, error) {
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

// PgSQLTx implements the Tx interface for PostgreSQL
type PgSQLTx struct {
	tx *sql.Tx
}

// Commit commits the transaction
func (t *PgSQLTx) Commit() error {
	return t.tx.Commit()
}

// Rollback rolls back the transaction
func (t *PgSQLTx) Rollback() error {
	return t.tx.Rollback()
}

// Prepare creates a prepared statement in transaction context
func (t *PgSQLTx) Prepare(query string) (Stmt, error) {
	// Convert ? placeholders to PostgreSQL $1, $2, etc.
	pgQuery := convertPlaceholders(query)

	stmt, err := t.tx.Prepare(pgQuery)
	if err != nil {
		return nil, NewPDOError("42601", 1, fmt.Sprintf("Prepare error: %v", err))
	}

	return &PgSQLStmt{
		stmt:       stmt,
		query:      pgQuery,
		params:     make(map[interface{}]*values.Value),
		paramOrder: []interface{}{},
	}, nil
}

// Query executes a query in transaction context
func (t *PgSQLTx) Query(query string) (Rows, error) {
	rows, err := t.tx.Query(query)
	if err != nil {
		return nil, NewPDOError("42601", 1, fmt.Sprintf("Query error: %v", err))
	}

	return &PgSQLRows{rows: rows}, nil
}

// Exec executes a statement in transaction context
func (t *PgSQLTx) Exec(query string) (Result, error) {
	result, err := t.tx.Exec(query)
	if err != nil {
		return nil, NewPDOError("42601", 1, fmt.Sprintf("Exec error: %v", err))
	}

	return result, nil
}

// Register PostgreSQL driver on package initialization
func init() {
	RegisterDriver("pgsql", &PgSQLDriver{})
}
