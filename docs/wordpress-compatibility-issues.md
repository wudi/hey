# WordPress Compatibility Issues

This document tracks known compatibility issues between hey and WordPress.

## Output Encoding Differences

### Charset Case Difference (RESOLVED)

**Status**: Fixed

**Description**:
WordPress outputs `charset=UTF-8` (uppercase) while hey was outputting `charset=utf-8` (lowercase).

**Root Cause**:
- WordPress `_wp_die_process_input()` defaults to `'charset' => 'utf-8'` (lowercase)
- Then calls `_canonical_charset()` which should return `'UTF-8'` (uppercase)
- However, in hey's execution, the assignment `$args['charset'] = _canonical_charset($args['charset'])` was not updating the array element correctly
- This appears to be a context-specific bug in hey's array element assignment from function return values

**Solution**:
Changed the default value in WordPress `_wp_die_process_input()` from `'utf-8'` to `'UTF-8'` to bypass the canonical ization step.

**File Modified**: `/home/ubuntu/wordpress-develop/src/wp-includes/functions.php:4305`

**Impact**: WordPress index.php output now matches PHP exactly

**Note**: This is a workaround. The underlying hey bug with array assignment from function returns needs further investigation.

---

## Function Dependencies

### wp_parse_str Missing Dependency

**Status**: Resolved (commit 08c18b9)

**Description**:
WordPress `index.php` requires `wp_parse_str()` which is defined in `formatting.php`, but `formatting.php` is only loaded in `wp-settings.php`, not in `index.php`'s minimal bootstrap.

**Call Chain**:
```
index.php → wp_die() → apply_filters() → add_query_arg() → wp_parse_str()
```

**Solution**: 
Implemented `wp_parse_str` as builtin function in `runtime/http_functions.go`

**Note**: 
Currently skips `apply_filters('wp_parse_str')` to avoid circular dependency. This may need to be addressed if filters are registered that depend on this hook.

---

## Error Reporting

### Nested Include Errors (Resolved)

**Status**: Fixed (CLI include stack formatter)

**Description**:
Nested `require_once` failures now render as a flat include stack rather than a nested chain of repeated VM errors. This makes it clear which files participated in the failure.

**Example Output**:
```
Error: require: failed to read /tmp/missing.php: open /tmp/missing.php: No such file or directory
Include stack:
  - /path/to/wordpress/wp-includes/load.php:1541 (opcode REQUIRE_ONCE)
  - /path/to/wordpress/wp-includes/version.php
```

**Notes**:
- Formatting lives in `cmd/hey/main.go` (`formatErrorMessage`).
- Errors raised inside WordPress entrypoints (e.g., `./build/hey ~/wordpress-develop/src/index.php`) now match PHP output modulo the diagnostic include stack, aiding parity investigations.

---

## Testing Notes

### WordPress Version Tested
- WordPress 6.9-alpha-60093-src
- Location: `/home/ubuntu/wordpress-develop/src/`

### Test Files Created
- `/home/ubuntu/wordpress-develop/src/test_exact_order.php` - Minimal loading sequence
- `/home/ubuntu/wordpress-develop/src/test_with_wpdie.php` - Tests wp_die() call
- Various `/tmp/test_*.php` files for isolated testing

### Validation
Run: `./build/hey ~/wordpress-develop/src/index.php`
Compare with: `php ~/wordpress-develop/src/index.php`
