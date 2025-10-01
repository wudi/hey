# PDO (PHP Data Objects) Implementation Specification

## Overview

PDO provides a unified interface for database access across multiple database backends (MySQL, SQLite, PostgreSQL). This implementation brings production-ready database connectivity to hey-codex.

## Architecture

### Core Components

```
PHP Layer (runtime/)
├─ pdo.go              → PDO class methods (__construct, prepare, query, exec, etc.)
├─ pdo_statement.go    → PDOStatement methods (execute, fetch, fetchAll, etc.)
├─ pdo_constants.go    → PDO::FETCH_*, PDO::PARAM_*, etc.
├─ pdo_classes.go      → Class descriptors for registration
└─ pdo_helpers.go      → Helper functions (newPDOMethod, param conversion)

Driver Layer (pkg/pdo/)
├─ driver.go           → Core interfaces (Driver, Conn, Stmt, Rows, Tx)
├─ dsn.go              → DSN parsing and building
├─ mysql_driver.go     → MySQL implementation (go-sql-driver/mysql)
├─ sqlite_driver.go    → SQLite implementation (modernc.org/sqlite) [TODO]
└─ pgsql_driver.go     → PostgreSQL implementation (lib/pq) [TODO]
```

### Design Principles

1. **Driver Abstraction**: All database-specific logic hidden behind interfaces
2. **Zero Special Cases**: No `if driver == "mysql"` branches in PHP layer
3. **SPL Pattern**: Methods receive `$this` as first parameter for object state
4. **Real Connections**: Uses production-ready Go SQL drivers

## Interfaces

### Driver Interface

```go
type Driver interface {
    Open(dsn string) (Conn, error)
    Name() string
}
```

### Connection Interface

```go
type Conn interface {
    Prepare(query string) (Stmt, error)
    Query(query string) (Rows, error)
    Exec(query string) (Result, error)
    Begin() (Tx, error)
    Close() error
    Ping() error
    LastInsertId() (int64, error)
    GetUnderlyingDB() *sql.DB
}
```

### Statement Interface

```go
type Stmt interface {
    BindValue(param interface{}, value *values.Value, dataType int) error
    Execute() (Result, error)
    Query() (Rows, error)
    Close() error
    RowCount() int64
}
```

### Rows Interface

```go
type Rows interface {
    Next() bool
    Scan(dest ...interface{}) error
    Columns() ([]string, error)
    Close() error
    Err() error
    FetchAssoc() (map[string]*values.Value, error)
    FetchNum() ([]*values.Value, error)
    FetchBoth() (map[string]*values.Value, []*values.Value, error)
}
```

## DSN Formats

### MySQL
```
mysql:host=localhost;port=3306;dbname=testdb
```

### SQLite
```
sqlite:/path/to/database.db
sqlite::memory:
```

### PostgreSQL
```
pgsql:host=localhost;port=5432;dbname=testdb
```

## PHP Usage Examples

### Basic Connection

```php
<?php
// MySQL
$pdo = new PDO('mysql:host=localhost;dbname=testdb', 'user', 'pass');

// SQLite
$pdo = new PDO('sqlite:/tmp/test.db');

// PostgreSQL
$pdo = new PDO('pgsql:host=localhost;dbname=testdb', 'user', 'pass');
?>
```

### Prepared Statements

```php
<?php
$stmt = $pdo->prepare('SELECT * FROM users WHERE age > ?');
$stmt->execute([25]);

while ($row = $stmt->fetch(PDO::FETCH_ASSOC)) {
    echo $row['username'] . "\n";
}
?>
```

### Named Parameters

```php
<?php
$stmt = $pdo->prepare('INSERT INTO users (username, email) VALUES (:name, :email)');
$stmt->bindValue(':name', 'john_doe');
$stmt->bindValue(':email', 'john@example.com');
$stmt->execute();

echo "Last insert ID: " . $pdo->lastInsertId() . "\n";
?>
```

### Transactions

```php
<?php
try {
    $pdo->beginTransaction();

    $pdo->exec('INSERT INTO users (username) VALUES ("alice")');
    $pdo->exec('INSERT INTO posts (user_id, title) VALUES (1, "First Post")');

    $pdo->commit();
    echo "Transaction committed\n";
} catch (PDOException $e) {
    $pdo->rollBack();
    echo "Transaction rolled back: " . $e->getMessage() . "\n";
}
?>
```

### Fetch Modes

```php
<?php
$stmt = $pdo->query('SELECT id, username, email FROM users LIMIT 3');

// Associative array
$row = $stmt->fetch(PDO::FETCH_ASSOC);
// ['id' => 1, 'username' => 'john', 'email' => 'john@example.com']

// Numeric array
$row = $stmt->fetch(PDO::FETCH_NUM);
// [0 => 1, 1 => 'john', 2 => 'john@example.com']

// Both (default)
$row = $stmt->fetch(PDO::FETCH_BOTH);
// [0 => 1, 'id' => 1, 1 => 'john', 'username' => 'john', ...]

// Fetch all rows
$rows = $stmt->fetchAll(PDO::FETCH_ASSOC);
?>
```

## PDO Class Methods

### Constructor
```php
__construct(string $dsn, ?string $username = null, ?string $password = null, ?array $options = null): void
```

### Query Execution
```php
prepare(string $query, array $options = []): PDOStatement|false
query(string $query): PDOStatement|false
exec(string $statement): int|false
```

### Transaction Management
```php
beginTransaction(): bool
commit(): bool
rollBack(): bool
inTransaction(): bool
```

### Metadata
```php
lastInsertId(?string $name = null): string|false
getAttribute(int $attribute): mixed
setAttribute(int $attribute, mixed $value): bool
```

### Error Handling
```php
errorCode(): ?string
errorInfo(): array  // [sqlstate, driver_code, message]
```

## PDOStatement Class Methods

### Execution
```php
execute(?array $params = null): bool
closeCursor(): bool
```

### Parameter Binding
```php
bindValue(mixed $param, mixed $value, int $type = PDO::PARAM_STR): bool
bindParam(mixed $param, mixed &$var, int $type = PDO::PARAM_STR): bool
```

### Fetching Results
```php
fetch(int $mode = PDO::FETCH_BOTH): mixed
fetchAll(int $mode = PDO::FETCH_BOTH): array
fetchColumn(int $column = 0): mixed
```

### Metadata
```php
rowCount(): int
columnCount(): int
```

### Error Handling
```php
errorCode(): ?string
errorInfo(): array
```

## PDO Constants

### Fetch Modes
```php
PDO::FETCH_ASSOC      // Associative array indexed by column name
PDO::FETCH_NUM        // Numeric array indexed by column number
PDO::FETCH_BOTH       // Both associative and numeric (default)
PDO::FETCH_OBJ        // Anonymous object with properties
PDO::FETCH_LAZY       // Fetch on demand
PDO::FETCH_CLASS      // Fetch into class instance
PDO::FETCH_INTO       // Update existing object
PDO::FETCH_COLUMN     // Return single column
PDO::FETCH_KEY_PAIR   // Two-column result as key-value pairs
```

### Parameter Types
```php
PDO::PARAM_NULL       // NULL type
PDO::PARAM_INT        // Integer type
PDO::PARAM_STR        // String type (default)
PDO::PARAM_LOB        // Large object (blob)
PDO::PARAM_BOOL       // Boolean type
```

### Error Modes
```php
PDO::ERRMODE_SILENT    // Silent mode (default)
PDO::ERRMODE_WARNING   // PHP warnings
PDO::ERRMODE_EXCEPTION // Throw PDOException
```

### Attributes
```php
PDO::ATTR_ERRMODE            // Error reporting mode
PDO::ATTR_DEFAULT_FETCH_MODE // Default fetch mode
PDO::ATTR_AUTOCOMMIT         // Auto-commit mode
PDO::ATTR_PERSISTENT         // Persistent connection
```

## Development Workflow

### Start Database Containers

```bash
make -f Makefile.pdo pdo-start
```

This starts:
- MySQL 8.0 on port 3306
- PostgreSQL 15 on port 5432
- phpMyAdmin on port 8080
- pgAdmin on port 8081

### Run Tests

```bash
# All PDO tests
make -f Makefile.pdo pdo-test

# MySQL only
make -f Makefile.pdo pdo-test-mysql

# PostgreSQL only
make -f Makefile.pdo pdo-test-postgres

# SQLite only
make -f Makefile.pdo pdo-test-sqlite
```

### Database CLI Access

```bash
# MySQL shell
make -f Makefile.pdo pdo-mysql-cli

# PostgreSQL shell
make -f Makefile.pdo pdo-postgres-cli
```

### Stop Containers

```bash
make -f Makefile.pdo pdo-stop
```

### Clean Everything

```bash
make -f Makefile.pdo pdo-clean
```

## Test Database Schema

### Users Table
```sql
CREATE TABLE users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(100) NOT NULL,
    age INT,
    balance DECIMAL(10, 2) DEFAULT 0.00,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Posts Table
```sql
CREATE TABLE posts (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    title VARCHAR(200) NOT NULL,
    content TEXT,
    published BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

### Tags Table
```sql
CREATE TABLE tags (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE
);
```

### Sample Data

5 users (john_doe, jane_smith, bob_wilson, alice_brown, charlie_davis)
5 posts (mix of published and draft)
5 tags (php, mysql, programming, web-development, tutorial)

## Implementation Status

### ✅ Phase 1: Core Architecture + MySQL (COMPLETED)
- [x] Driver interface design
- [x] DSN parsing
- [x] MySQL driver implementation
- [x] PDO class methods
- [x] PDOStatement methods
- [x] Transaction support
- [x] Docker test environment
- [x] Sample data

### 🔧 Phase 1.5: Bug Fixes (IN PROGRESS)
- [ ] Fix Value API calls (use v.Data instead of methods)
- [ ] Update methods to accept $this as first parameter
- [ ] Test basic MySQL connectivity
- [ ] Fix compilation errors

### 📋 Phase 2: SQLite Driver (TODO)
- [ ] Implement SQLiteDriver
- [ ] Implement SQLiteConn
- [ ] Implement SQLiteStmt
- [ ] Memory database support (:memory:)
- [ ] File database support
- [ ] Tests

### 📋 Phase 3: PostgreSQL Driver (TODO)
- [ ] Implement PgSQLDriver
- [ ] Implement PgSQLConn
- [ ] Implement PgSQLStmt
- [ ] RETURNING clause support
- [ ] Array type handling
- [ ] Tests

### 📋 Phase 4: Advanced Features (TODO)
- [ ] PDO::FETCH_CLASS support
- [ ] Named parameter binding (:name)
- [ ] LOB support
- [ ] Error mode handling
- [ ] Prepared statement caching
- [ ] Connection pooling

## Technical Notes

### Object State Management

PDO methods follow the SPL pattern where `$this` is passed as the first argument:

```go
func pdoConstruct(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
    thisObj := args[0]  // $this object
    dsn := args[1].Data.(string)
    // ...
}
```

Object-specific state is stored in `obj.Properties`:

```go
obj := thisObj.Data.(*values.Object)
obj.Properties["__pdo_conn"] = connValue
obj.Properties["__pdo_driver"] = driverValue
```

### Value API

Values use direct field access, not methods:

```go
// ✅ Correct
if v.Type == values.TypeInt {
    intVal := v.Data.(int64)
}

// ❌ Incorrect
if v.IsInt() {
    intVal := v.AsInt()
}
```

### Error Handling

PDO errors use the PDOError type:

```go
return nil, &pdo.PDOError{
    SQLState: "HY000",
    Code:     2002,
    Message:  "Connection failed",
}
```

### Driver Registration

Drivers register themselves in `init()`:

```go
func init() {
    pdo.RegisterDriver("mysql", &MySQLDriver{})
}
```

## References

- PHP PDO Documentation: https://www.php.net/manual/en/book.pdo.php
- go-sql-driver/mysql: https://github.com/go-sql-driver/mysql
- modernc.org/sqlite: https://gitlab.com/cznic/sqlite
- lib/pq (PostgreSQL): https://github.com/lib/pq

## Future Enhancements

1. **Connection Pooling**: Reuse connections across requests
2. **Prepared Statement Caching**: Cache parsed statements
3. **Query Builder**: Optional fluent interface
4. **Migration Support**: Schema versioning
5. **ORM Integration**: Active Record pattern
6. **Async Support**: Non-blocking queries

---

**Status**: Phase 1 architecture complete, bug fixes in progress.
**Next Steps**: Fix compilation errors, test MySQL connectivity, implement SQLite driver.
