# MySQLi Examples

This directory contains practical examples demonstrating mysqli usage in Hey-Codex.

## Prerequisites

1. **Start MySQL Docker container**:
   ```bash
   docker-compose up -d
   ```

2. **Verify MySQL is running**:
   ```bash
   docker ps | grep hey-mysql
   ```

3. **Build Hey-Codex**:
   ```bash
   make build
   ```

## Examples

### 1. CRUD Demo (`mysqli_crud_demo.php`)

Comprehensive demonstration of Create, Read, Update, Delete operations using:
- **Part 1**: OOP style (recommended)
- **Part 2**: Procedural style
- **Part 3**: Prepared statements
- **Part 4**: Connection info

**Run**:
```bash
MYSQL_HOST=localhost MYSQL_USER=testuser MYSQL_PASS=testpass MYSQL_DB=testdb \
  ./build/hey examples/mysqli_crud_demo.php
```

**Expected Output**:
```
=== MySQLi CRUD Demo ===

=== PART 1: OOP Style ===
1. Connecting to database...
   ✓ Connected successfully
2. Creating new record...
   ✓ Record created with ID: 20
   Affected rows: 1
3. Reading records...
   Found 5 records:
   - ID: 1, Name: Alice, Email: alice@example.com, Age: 30
   ...
```

### Key Features Demonstrated

#### OOP Style
```php
$mysqli = new mysqli('localhost', 'user', 'pass', 'db');
$result = $mysqli->query("SELECT * FROM users");
$row = $result->fetch_assoc();
echo $mysqli->insert_id;
echo $mysqli->affected_rows;
$mysqli->close();
```

#### Procedural Style
```php
$conn = mysqli_connect('localhost', 'user', 'pass', 'db');
$result = mysqli_query($conn, "SELECT * FROM users");
$row = mysqli_fetch_assoc($result);
echo mysqli_insert_id($conn);
echo mysqli_affected_rows($conn);
mysqli_close($conn);
```

#### Prepared Statements
```php
$stmt = $mysqli->prepare("INSERT INTO users (name, email) VALUES (?, ?)");
echo $stmt->param_count;  // Number of placeholders
```

#### Error Handling
```php
$result = $mysqli->query("SELECT * FROM nonexistent");
if (!$result) {
    echo "Error: " . $mysqli->error . " (Code: " . $mysqli->errno . ")";
}
```

## Environment Variables

Set these variables to configure database connection:
- `MYSQL_HOST` - Database host (default: localhost)
- `MYSQL_USER` - Database user (default: testuser)
- `MYSQL_PASS` - Database password (default: testpass)
- `MYSQL_DB` - Database name (default: testdb)

## Test Database Schema

The examples use the `users` table:
```sql
CREATE TABLE users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100) NOT NULL UNIQUE,
    age INT DEFAULT 0
);
```

This schema is automatically created when you run `docker-compose up`.

## API Coverage

These examples demonstrate:
- ✅ Connection management (connect, close)
- ✅ Query execution (SELECT, INSERT, UPDATE, DELETE)
- ✅ Result fetching (fetch_assoc, fetch_row)
- ✅ Insert ID and affected rows
- ✅ Error handling (errno, error)
- ✅ Character set management
- ✅ Prepared statements (structure)
- ✅ Real escape string
- ✅ Connection info (server_info, host_info, etc.)

## Next Steps

1. **Try the demo**:
   ```bash
   ./build/hey examples/mysqli_crud_demo.php
   ```

2. **Modify the examples** to suit your use case

3. **Build your own application** using mysqli in Hey-Codex

## Notes

- **Prepared statement binding** (bind_param, execute) are stub implementations
- All other features are **fully functional** with real MySQL
- Examples use **real MySQL database** via Docker
- Compatible with both **OOP and procedural** styles

## Troubleshooting

### Connection Failed
```bash
# Check if MySQL container is running
docker ps | grep hey-mysql

# View MySQL logs
docker logs hey-mysql

# Restart MySQL
docker-compose down
docker-compose up -d
```

### Permission Denied
```bash
# Ensure testuser has privileges
docker exec -it hey-mysql mysql -uroot -prootpass \
  -e "GRANT ALL ON testdb.* TO 'testuser'@'%'; FLUSH PRIVILEGES;"
```

### Build Issues
```bash
# Clean rebuild
make clean
make build
```
