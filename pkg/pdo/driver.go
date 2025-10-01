package pdo

import (
	"context"
	"database/sql"

	"github.com/wudi/hey/values"
)

// Driver defines the interface that all PDO drivers must implement
type Driver interface {
	// Open creates a new database connection
	Open(dsn string) (Conn, error)

	// Name returns the driver name (mysql, sqlite, pgsql)
	Name() string
}

// Conn represents a database connection
type Conn interface {
	// Prepare creates a prepared statement
	Prepare(query string) (Stmt, error)

	// Query executes a query that returns rows
	Query(query string) (Rows, error)

	// Exec executes a query that doesn't return rows
	Exec(query string) (Result, error)

	// Begin starts a transaction
	Begin() (Tx, error)

	// Close closes the connection
	Close() error

	// Ping verifies connection is alive
	Ping() error

	// LastInsertId returns the last inserted ID
	LastInsertId() (int64, error)

	// GetUnderlyingDB returns the underlying *sql.DB for advanced operations
	GetUnderlyingDB() *sql.DB
}

// Stmt represents a prepared statement
type Stmt interface {
	// BindValue binds a value to a parameter (by position or name)
	BindValue(param interface{}, value *values.Value, dataType int) error

	// Execute executes a prepared statement
	Execute() (Result, error)

	// Query executes a prepared statement and returns rows
	Query() (Rows, error)

	// Close closes the statement
	Close() error

	// RowCount returns the number of rows affected by the last statement
	RowCount() int64
}

// Rows represents a result set from a query
type Rows interface {
	// Next advances to the next row
	Next() bool

	// Scan scans the current row into destination values
	Scan(dest ...interface{}) error

	// Columns returns column names
	Columns() ([]string, error)

	// Close closes the rows
	Close() error

	// Err returns any error encountered during iteration
	Err() error

	// FetchAssoc fetches the next row as an associative array
	FetchAssoc() (map[string]*values.Value, error)

	// FetchNum fetches the next row as a numeric array
	FetchNum() ([]*values.Value, error)

	// FetchBoth fetches the next row as both associative and numeric
	FetchBoth() (map[string]*values.Value, []*values.Value, error)
}

// Result represents the result of a statement execution
type Result interface {
	// LastInsertId returns the last inserted ID
	LastInsertId() (int64, error)

	// RowsAffected returns the number of rows affected
	RowsAffected() (int64, error)
}

// Tx represents a database transaction
type Tx interface {
	// Commit commits the transaction
	Commit() error

	// Rollback rolls back the transaction
	Rollback() error

	// Prepare creates a prepared statement in the transaction context
	Prepare(query string) (Stmt, error)

	// Query executes a query in the transaction context
	Query(query string) (Rows, error)

	// Exec executes a statement in the transaction context
	Exec(query string) (Result, error)
}

// DSN represents a parsed data source name
type DSN struct {
	Driver   string            // mysql, sqlite, pgsql
	Host     string            // hostname or path
	Port     int               // port number
	Database string            // database name
	Username string            // username
	Password string            // password
	Options  map[string]string // additional options
}

// DriverRegistry holds all registered PDO drivers
var driverRegistry = make(map[string]Driver)

// RegisterDriver registers a PDO driver
func RegisterDriver(name string, driver Driver) {
	driverRegistry[name] = driver
}

// GetDriver retrieves a registered driver by name
func GetDriver(name string) (Driver, bool) {
	driver, ok := driverRegistry[name]
	return driver, ok
}

// PDOError represents a PDO error with SQL state
type PDOError struct {
	SQLState string // SQL state code (e.g., "42S02")
	Code     int    // Driver-specific error code
	Message  string // Error message
}

func (e *PDOError) Error() string {
	return e.Message
}

// NewPDOError creates a new PDO error
func NewPDOError(sqlState string, code int, message string) *PDOError {
	return &PDOError{
		SQLState: sqlState,
		Code:     code,
		Message:  message,
	}
}

// ParamType represents PDO parameter types
type ParamType int

const (
	ParamNull ParamType = iota
	ParamInt
	ParamStr
	ParamLOB
	ParamBool
)

// FetchMode represents PDO fetch modes
// Must match PHP PDO constants exactly
type FetchMode int

const (
	FetchLazy      FetchMode = 1
	FetchAssoc     FetchMode = 2
	FetchNum       FetchMode = 3
	FetchBoth      FetchMode = 4
	FetchObj       FetchMode = 5
	FetchBound     FetchMode = 6
	FetchColumn    FetchMode = 7
	FetchClass     FetchMode = 8
	FetchInto      FetchMode = 9
	FetchFunc      FetchMode = 10
	FetchKeyPair   FetchMode = 12
)

// ErrorMode represents PDO error modes
type ErrorMode int

const (
	ErrModeSilent ErrorMode = iota
	ErrModeWarning
	ErrModeException
)

// Context-aware wrapper types for cancellation support
type contextConn struct {
	Conn
	ctx context.Context
}

func (c *contextConn) Prepare(query string) (Stmt, error) {
	return c.Conn.Prepare(query)
}

func (c *contextConn) Query(query string) (Rows, error) {
	return c.Conn.Query(query)
}

func (c *contextConn) Exec(query string) (Result, error) {
	return c.Conn.Exec(query)
}
