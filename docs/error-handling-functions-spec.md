# PHP Error Handling Functions Implementation Specification

This document tracks the implementation status of PHP error handling and logging functions in the Hey-Codex interpreter.

## Implementation Status Legend
- ✅ **IMPLEMENTED** - Fully implemented and tested
- 🚧 **IN_PROGRESS** - Currently being implemented
- 📝 **PLANNED** - Specified but not yet implemented
- ❌ **NOT_PLANNED** - Not planned for current iteration

## Error Handling Functions

### Error Reporting and Management
| Function | Status | Description | Test Cases |
|----------|--------|-------------|------------|
| `error_reporting()` | ✅ IMPLEMENTED | Get or set error reporting level | Get current level, set new level, restore |
| `error_get_last()` | ✅ IMPLEMENTED | Get last occurred error | After clear, after error, error structure |
| `error_clear_last()` | ✅ IMPLEMENTED | Clear the most recent error | Basic clear, verify null return |

### Error Generation
| Function | Status | Description | Test Cases |
|----------|--------|-------------|------------|
| `trigger_error()` | ✅ IMPLEMENTED | Generate user-level error/warning/notice | User notice, user warning, default level |
| `user_error()` | ✅ IMPLEMENTED | Alias of trigger_error | Basic functionality as alias |

### Error Handler Management
| Function | Status | Description | Test Cases |
|----------|--------|-------------|------------|
| `set_error_handler()` | ✅ IMPLEMENTED | Set user-defined error handler | Set handler, return previous |
| `restore_error_handler()` | ✅ IMPLEMENTED | Restore previous error handler | Restore after set, return value |
| `get_error_handler()` | ❌ NOT_PLANNED | Get user-defined error handler | Not available in PHP 8.4+ |

### Exception Handler Management
| Function | Status | Description | Test Cases |
|----------|--------|-------------|------------|
| `set_exception_handler()` | ✅ IMPLEMENTED | Set user-defined exception handler | Set handler, return previous |
| `restore_exception_handler()` | ✅ IMPLEMENTED | Restore previous exception handler | Restore after set, return value |
| `get_exception_handler()` | ❌ NOT_PLANNED | Get user-defined exception handler | Not available in PHP 8.4+ |

### Debug and Backtrace Functions
| Function | Status | Description | Test Cases |
|----------|--------|-------------|------------|
| `debug_backtrace()` | ✅ IMPLEMENTED | Generate a backtrace | Default params, with options, with limit |
| `debug_print_backtrace()` | ✅ IMPLEMENTED | Print a backtrace | Basic output, with options, with limit |

### Logging Functions
| Function | Status | Description | Test Cases |
|----------|--------|-------------|------------|
| `error_log()` | ✅ IMPLEMENTED | Send error message to defined error handling routines | System log, email type, file type, no args |

## Error Constants

### Core Error Types
| Constant | Status | Value | Description |
|----------|--------|-------|-------------|
| `E_ERROR` | ✅ IMPLEMENTED | 1 | Fatal run-time errors |
| `E_WARNING` | ✅ IMPLEMENTED | 2 | Run-time warnings (non-fatal errors) |
| `E_PARSE` | ✅ IMPLEMENTED | 4 | Compile-time parse errors |
| `E_NOTICE` | ✅ IMPLEMENTED | 8 | Run-time notices |

### Core System Error Types
| Constant | Status | Value | Description |
|----------|--------|-------|-------------|
| `E_CORE_ERROR` | ✅ IMPLEMENTED | 16 | Fatal errors during PHP's initial startup |
| `E_CORE_WARNING` | ✅ IMPLEMENTED | 32 | Warnings during PHP's initial startup |
| `E_COMPILE_ERROR` | ✅ IMPLEMENTED | 64 | Fatal compile-time errors |
| `E_COMPILE_WARNING` | ✅ IMPLEMENTED | 128 | Compile-time warnings |

### User Error Types
| Constant | Status | Value | Description |
|----------|--------|-------|-------------|
| `E_USER_ERROR` | ✅ IMPLEMENTED | 256 | User-generated error message |
| `E_USER_WARNING` | ✅ IMPLEMENTED | 512 | User-generated warning message |
| `E_USER_NOTICE` | ✅ IMPLEMENTED | 1024 | User-generated notice message |
| `E_USER_DEPRECATED` | ✅ IMPLEMENTED | 16384 | User-generated deprecation message |

### Special Error Types
| Constant | Status | Value | Description |
|----------|--------|-------|-------------|
| `E_STRICT` | ✅ IMPLEMENTED | 2048 | Run-time suggestions for forward compatibility |
| `E_RECOVERABLE_ERROR` | ✅ IMPLEMENTED | 4096 | Catchable fatal error |
| `E_DEPRECATED` | ✅ IMPLEMENTED | 8192 | Run-time deprecation notices |
| `E_ALL` | ✅ IMPLEMENTED | 30719 | All errors, warnings, and notices |

## Implementation Details

### Error State Management
- Global error state maintained with thread-safe mutex protection
- Last error stored as array with message, type, file, and line information
- Error reporting level stored as integer bitmask

### Error Handler Integration
- Custom error handlers stored as callable values
- Handler stack management for restore functionality
- Exception handlers managed separately from error handlers

### Backtrace Implementation
- Uses Go runtime.Callers() for stack trace generation
- Provides file, line, and function information
- Supports limit parameter for depth control

### Compatibility Notes
- `get_error_handler()` and `get_exception_handler()` functions are not implemented as they don't exist in PHP 8.4+
- Error output formatting matches PHP standard output format
- Error reporting bitmask behavior matches PHP specifications

## Test Coverage

### Unit Tests
- All functions have comprehensive unit tests in `runtime/error_test.go`
- Constants tested for correct values and availability
- Edge cases covered including null arguments, empty strings

### Integration Tests
- Error state persistence across function calls
- Handler setting and restoration workflow
- Error generation and retrieval cycle

## Usage Examples

### Basic Error Handling
```php
// Set error reporting level
$old_level = error_reporting(E_ERROR | E_WARNING);

// Trigger a user error
trigger_error("Custom error message", E_USER_WARNING);

// Get last error details
$last_error = error_get_last();
print_r($last_error);

// Clear last error
error_clear_last();
```

### Custom Error Handler
```php
function myErrorHandler($errno, $errstr, $errfile, $errline) {
    echo "Error [$errno]: $errstr in $errfile on line $errline\n";
    return true;
}

$old_handler = set_error_handler('myErrorHandler');
trigger_error("Test error", E_USER_NOTICE);
restore_error_handler();
```

### Debug Information
```php
function level1() { level2(); }
function level2() { level3(); }
function level3() {
    $trace = debug_backtrace();
    debug_print_backtrace();
}
level1();
```

## Performance Considerations
- Minimal overhead for error reporting level checks
- Backtrace generation uses efficient Go runtime functions
- Thread-safe implementation with read-write mutexes for optimal performance