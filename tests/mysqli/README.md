# MySQLi Test Suite

This directory contains all test files for the MySQLi implementation.

## Test Files

### Integration Tests
- **test_mysqli_complete_integration.php** - Main integration test suite (33 tests, 100% passing)
- **test_mysqli_features.php** - Feature detection test (222 features)

### WordPress Compatibility
- **test_wordpress_compat.php** - WordPress compatibility test (29 tests, 100% passing)
- **test_wordpress*.php** - Various WordPress integration tests

### Debug & Development
- **test_debug_properties.php** - Property synchronization debugging
- **test_debug_insert_only.php** - INSERT operation debugging
- **test_insert_isolated.php** - Isolated INSERT test
- **test_oop_property.php** - OOP property access debugging

## Running Tests

### Prerequisites
```bash
# Start MySQL container
docker-compose up -d

# Verify MySQL is running
docker ps | grep hey-mysql
```

### Run Integration Tests
```bash
export MYSQL_HOST=localhost
export MYSQL_USER=testuser
export MYSQL_PASS=testpass
export MYSQL_DB=testdb

./build/hey tests/mysqli/test_mysqli_complete_integration.php
```

### Run WordPress Compatibility Test
```bash
./build/hey tests/mysqli/test_wordpress_compat.php
```

### Run All Tests
```bash
for test in tests/mysqli/test_*.php; do
    echo "Running $test..."
    ./build/hey "$test"
done
```

## Test Results

### Integration Tests (33/33 ✓)
- Connection Tests: 3/3
- Query Execution: 3/3
- Fetch Methods: 4/4
- Result Info: 3/3
- Data Modification: 2/2
- Prepared Statements: 4/4
- Error Handling: 4/4
- Character Set: 4/4
- Connection Info: 3/3
- Advanced Functions: 3/3

### WordPress Compatibility (29/29 ✓)
- Core Connection: 3/3
- Character Set: 3/3
- Query Execution: 4/4
- Result Fetching: 3/3
- Result Metadata: 4/4
- Error Handling: 2/2
- Data Sanitization: 2/2
- Connection Info: 3/3
- Prepared Statements: 2/2
- WordPress Patterns: 3/3

## Success Rate
- **Overall**: 100% (62/62 tests passing)
- **Integration**: 100% (33/33)
- **WordPress**: 100% (29/29)

## Environment Variables
- `MYSQL_HOST` - Database host (default: localhost)
- `MYSQL_USER` - Database user (default: testuser)
- `MYSQL_PASS` - Database password (default: testpass)
- `MYSQL_DB` - Database name (default: testdb)
