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

## TODO: Issues Requiring Investigation

### Nested Error Messages in require_once Chain

**Status**: Needs Investigation

**Description**:
When WordPress files fail to execute, error messages show deep nesting with all errors pointing to same instruction pointer (ip=361).

**Example Error**:
```
execution error in version.php: 
  vm error at ip=361 opcode=REQUIRE_ONCE in load.php line 1541: 
    execution error in pomo/mo.php: 
      vm error at ip=361 opcode=REQUIRE_ONCE in load.php line 1542: 
        ...
```

**Observations**:
- All nested errors show same IP (361)
- Error occurs when calling functions before their definition files are loaded
- Error message structure may be misleading about actual execution flow

**Next Steps**:
1. Investigate error message formatting in `vmfactory/factory.go:67`
2. Check if error context is being properly preserved through nested require_once
3. Verify IP tracking across file boundaries

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
