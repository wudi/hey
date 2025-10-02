# WordPress Compatibility Issues

This document tracks known compatibility issues between hey and WordPress.

## Output Encoding Differences

### Charset Case Difference (Minor)

**Status**: Known, Low Priority

**Description**:
WordPress outputs `charset=UTF-8` (uppercase) while hey outputs `charset=utf-8` (lowercase).

**Location**: HTTP Content-Type header generation

**Impact**: None - both are valid and equivalent per HTTP/HTML standards

**Example**:
```
PHP:  <meta http-equiv="Content-Type" content="text/html; charset=UTF-8" />
hey:  <meta http-equiv="Content-Type" content="text/html; charset=utf-8" />
```

**Root Cause**: Different header generation logic - need to investigate where charset is set

**Priority**: Low - purely cosmetic difference with no functional impact

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
