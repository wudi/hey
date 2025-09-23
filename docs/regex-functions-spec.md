# PHP Regex Functions Implementation Specification

This document tracks the implementation status of PHP PCRE (Perl Compatible Regular Expression) functions in the Hey-Codex interpreter.

## Implementation Status Legend
- ✅ **IMPLEMENTED** - Fully implemented and tested with PHP compatibility
- 🚧 **PARTIAL** - Basic implementation, missing advanced features
- 📝 **PLANNED** - Specified but not yet implemented
- ❌ **NOT_PLANNED** - Not planned for current iteration

## Core Regex Functions

### Pattern Matching
| Function | Status | Description | Test Coverage |
|----------|--------|-------------|---------------|
| `preg_match()` | ✅ IMPLEMENTED | Find first match of pattern | Basic match, no match, capture groups, case-insensitive, invalid patterns |
| `preg_match_all()` | ✅ IMPLEMENTED | Find all matches of pattern | Multiple matches, capture groups, empty results, PHP array structure |

### Search and Replace
| Function | Status | Description | Test Coverage |
|----------|--------|-------------|---------------|
| `preg_replace()` | ✅ IMPLEMENTED | Replace matches with replacement | Basic replacement, multiple patterns |
| `preg_filter()` | ✅ IMPLEMENTED | Like preg_replace but filters arrays | Array filtering, key preservation, string replacement |
| `preg_replace_callback()` | 📝 PLANNED | Replace with callback function | Callback support needs implementation |
| `preg_replace_callback_array()` | 📝 PLANNED | Replace with multiple callbacks | Advanced callback support needed |

### String Utilities
| Function | Status | Description | Test Coverage |
|----------|--------|-------------|---------------|
| `preg_split()` | ✅ IMPLEMENTED | Split string by regex pattern | Pattern splitting, limit parameter |
| `preg_quote()` | ✅ IMPLEMENTED | Quote regex metacharacters | Meta character escaping, delimiter escaping |

### Array Utilities
| Function | Status | Description | Test Coverage |
|----------|--------|-------------|---------------|
| `preg_grep()` | ✅ IMPLEMENTED | Return array entries that match pattern | Array filtering with pattern matching |

### Error Handling
| Function | Status | Description | Test Coverage |
|----------|--------|-------------|---------------|
| `preg_last_error()` | ✅ IMPLEMENTED | Get last PCRE error code | Error code constants, error state tracking |
| `preg_last_error_msg()` | ✅ IMPLEMENTED | Get last PCRE error message | Error message retrieval |

## Key Features Implemented

### Pattern Support
- ✅ **PHP Delimiter Syntax**: `/pattern/flags` format fully supported
- ✅ **Common Delimiters**: `/`, `#`, `~`, `@` and other delimiter characters
- ✅ **Regex Flags**: Case-insensitive (`i`), multiline (`m`), dotall (`s`)
- ✅ **Capture Groups**: Full support for parenthetical groups
- ✅ **Meta Character Escaping**: Automatic escaping via `preg_quote()`

### Reference Parameter Handling
- ✅ **Matches Arrays**: Proper PHP reference semantics for match results
- ✅ **Array Structure**: Correct PHP match array format with capture groups
- ✅ **Key Preservation**: Original array keys maintained in filter operations

### Error Management
- ✅ **PCRE Error Codes**: Standard PHP error constants implemented
- ✅ **Error Tracking**: Thread-safe error state management
- ✅ **Error Recovery**: Proper error clearing between operations

## Implementation Details

### Architecture
```
PHP Pattern → Pattern Parser → Go Regexp → Result Formatter → PHP Value
```

### Core Components
1. **Pattern Parser**: Converts PHP regex patterns (`/pattern/flags`) to Go format
2. **Error Manager**: Tracks PCRE-compatible error states with thread safety
3. **Result Formatter**: Creates PHP-compatible result arrays and structures
4. **Reference Handler**: Manages PHP reference parameters for output arrays

### PHP Compatibility
- **Array Formats**: Match arrays exactly follow PHP structure
- **Key Preservation**: Original array keys maintained in filter operations
- **Error Behavior**: Error codes and messages match PHP PCRE behavior
- **Type Coercion**: Proper string conversion for all input types

## Test Coverage

### Validation Methodology
- **PHP Reference Tests**: All behavior validated against real PHP 8.0+
- **Unit Test Coverage**: Comprehensive Go test suite with 40+ test cases
- **Integration Tests**: End-to-end testing via Hey interpreter
- **Edge Case Testing**: Invalid patterns, empty inputs, unicode strings

### Test Categories
1. **Basic Functionality**: Core pattern matching and replacement
2. **Error Handling**: Invalid patterns, compilation failures
3. **Reference Parameters**: Output array population and modification
4. **PHP Compatibility**: Exact behavior matching with native PHP
5. **Performance**: Large string and array processing

## Usage Examples

### Basic Pattern Matching
```php
// Simple match
preg_match('/hello/', 'hello world', $matches);
// $matches = ['hello']

// Capture groups
preg_match('/(\w+)\s+(\w+)/', 'hello world', $matches);
// $matches = ['hello world', 'hello', 'world']
```

### Array Processing
```php
// Filter matching elements
$input = ['apple', 'banana', 'cherry', 'apricot'];
$result = preg_filter('/^ap/', 'fruit: $0', $input);
// $result = [0 => 'fruit: apple', 3 => 'fruit: apricot']

// Grep matching elements
$result = preg_grep('/^ap/', $input);
// $result = [0 => 'apple', 3 => 'apricot']
```

### Error Handling
```php
// Invalid pattern
preg_match('/[/', 'test');
echo preg_last_error(); // 1 (PREG_INTERNAL_ERROR)
echo preg_last_error_msg(); // "missing terminating ] for character class"
```

## Performance Characteristics

- **Pattern Compilation**: Cached per pattern for multiple uses
- **Memory Efficiency**: Minimal copying for large strings/arrays
- **Thread Safety**: All operations thread-safe with proper locking
- **Error Overhead**: Minimal performance impact for error tracking

## Future Enhancements

### Planned Callback Support
- `preg_replace_callback()` - Single callback function support
- `preg_replace_callback_array()` - Multiple callback pattern support

### Advanced Features (Future)
- Named capture groups support
- Advanced PCRE flags and modifiers
- Recursive pattern support
- Performance optimizations for large datasets

## Version History

- **v1.0**: Core functions implemented (preg_match, preg_match_all, preg_replace)
- **v1.1**: Added utility functions (preg_split, preg_quote, preg_grep)
- **v1.2**: Added filtering and error handling (preg_filter, error functions)
- **v1.3**: Enhanced reference parameter support and PHP compatibility

---

*This specification reflects the current implementation status as of the regex system development. All implemented functions pass comprehensive PHP compatibility tests.*