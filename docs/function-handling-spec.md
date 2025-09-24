# Function Handling Functions Specification

This document describes the implementation status of PHP's Function Handling functions in hey-codex.

## Overview

Function handling functions provide runtime introspection, dynamic function calling, and function lifecycle management capabilities. These functions are essential for implementing callbacks, dynamic dispatching, and reflection-based programming patterns.

## Functions List and Implementation Status

### Argument Introspection Functions
- [ ] `func_num_args()` - Returns the number of arguments passed to the function
- [ ] `func_get_arg($arg_num)` - Returns an item from the argument list
- [ ] `func_get_args()` - Returns an array comprising a function's argument list

### Function Introspection Functions
- [ ] `function_exists($function_name)` - Returns TRUE if the given function has been defined
- [ ] `get_defined_functions()` - Returns an array of all defined functions

### Dynamic Function Calling Functions
- [ ] `call_user_func($callback, ...$args)` - Call the callback given by the first parameter
- [ ] `call_user_func_array($callback, $args)` - Call a callback with an array of parameters

### Static Method Calling Functions
- [ ] `forward_static_call($callback, ...$args)` - Call a static method
- [ ] `forward_static_call_array($callback, $args)` - Call a static method and pass the arguments as array

### Dynamic Function Creation Functions
- [ ] `create_function($args, $code)` - Create a function dynamically (deprecated in PHP 7.2, removed in PHP 8.0)

### Lifecycle Management Functions
- [ ] `register_shutdown_function($callback, ...$args)` - Register a function for execution on shutdown
- [ ] `register_tick_function($callback, ...$args)` - Register a function for execution on each tick
- [ ] `unregister_tick_function($callback)` - De-register a function for execution on each tick

## Detailed Specifications

### func_num_args()

**Signature:** `int func_num_args()`

**Description:** Returns the number of arguments passed to the user-defined function.

**Behavior:**
- Can only be called from within a user-defined function
- Returns total count of arguments passed, including optional arguments
- Throws `Error` if called from global scope

**Example:**
```php
function test($a, $b = 'default') {
    return func_num_args(); // Returns actual number of args passed
}
test('hello'); // Returns 1
test('hello', 'world'); // Returns 2
```

### func_get_arg($arg_num)

**Signature:** `mixed func_get_arg(int $arg_num)`

**Description:** Returns the specified argument from a user-defined function's argument list.

**Parameters:**
- `$arg_num`: The argument number (0-based index)

**Behavior:**
- Can only be called from within a user-defined function
- Returns the argument at the specified index
- Throws `Error` if called from global scope
- Throws `Warning` if `$arg_num` is greater than the number of arguments

### func_get_args()

**Signature:** `array func_get_args()`

**Description:** Returns an array comprising a function's argument list.

**Behavior:**
- Can only be called from within a user-defined function
- Returns indexed array of all arguments passed to the function
- Throws `Error` if called from global scope

### function_exists($function_name)

**Signature:** `bool function_exists(string $function_name)`

**Description:** Checks whether a function with the given name has been defined.

**Parameters:**
- `$function_name`: The function name (case-insensitive)

**Behavior:**
- Returns `true` if function exists (built-in or user-defined)
- Returns `false` if function does not exist
- Function name comparison is case-insensitive

### get_defined_functions()

**Signature:** `array get_defined_functions()`

**Description:** Returns an associative array of all currently defined functions.

**Return Value:**
- Array with two keys:
  - `'internal'`: Array of built-in function names
  - `'user'`: Array of user-defined function names

### call_user_func($callback, ...$args)

**Signature:** `mixed call_user_func(callable $callback, mixed ...$args)`

**Description:** Calls a user function given by the first parameter.

**Parameters:**
- `$callback`: The callable to call (string function name or array [class, method])
- `...$args`: Zero or more parameters to be passed to the callback

**Behavior:**
- Supports string function names
- Supports array callbacks for static methods: `['ClassName', 'methodName']`
- Returns the return value of the callback
- Throws `Error` if callback is not callable

### call_user_func_array($callback, $args)

**Signature:** `mixed call_user_func_array(callable $callback, array $args)`

**Description:** Calls a user function with an array of parameters.

**Parameters:**
- `$callback`: The callable to call
- `$args`: Array of parameters to pass to the callback

**Behavior:**
- Similar to `call_user_func` but takes arguments as array
- Array values are passed as individual arguments to the callback

### forward_static_call($callback, ...$args)

**Signature:** `mixed forward_static_call(callable $callback, mixed ...$args)`

**Description:** Calls a static method from within another static method, forwarding the called class context.

**Parameters:**
- `$callback`: The static method to call
- `...$args`: Arguments to pass

**Behavior:**
- Must be called from within a static method context
- Forwards late static binding context

### forward_static_call_array($callback, $args)

**Signature:** `mixed forward_static_call_array(callable $callback, array $args)`

**Description:** Array version of `forward_static_call()`.

### create_function($args, $code)

**Signature:** `string create_function(string $args, string $code)`

**Description:** Creates an anonymous function from the parameters passed, and returns a unique name for it.

**Status:** DEPRECATED (PHP 7.2+), REMOVED (PHP 8.0+)
**Implementation Priority:** LOW (modern PHP uses closures instead)

### register_shutdown_function($callback, ...$args)

**Signature:** `bool register_shutdown_function(callable $callback, mixed ...$args)`

**Description:** Registers a function to be executed on script termination.

**Behavior:**
- Functions are called in FIFO order
- Called on normal termination, fatal errors, or explicit exit()
- Cannot be unregistered once registered

### register_tick_function($callback, ...$args)

**Signature:** `bool register_tick_function(callable $callback, mixed ...$args)`

**Description:** Registers a function to be executed on each tick.

**Behavior:**
- Only executes within `declare(ticks=n)` blocks
- Called after every n low-level statements

### unregister_tick_function($callback)

**Signature:** `void unregister_tick_function(callable $callback)`

**Description:** De-registers a function for execution on each tick.

## Implementation Priority

1. **HIGH PRIORITY** (Core functionality):
   - `func_num_args()`, `func_get_arg()`, `func_get_args()` - Essential for variadic functions
   - `function_exists()` - Common in conditional loading patterns
   - `call_user_func()`, `call_user_func_array()` - Dynamic function calling

2. **MEDIUM PRIORITY** (Advanced features):
   - `get_defined_functions()` - Reflection and debugging
   - `forward_static_call()`, `forward_static_call_array()` - OOP patterns

3. **LOW PRIORITY** (Lifecycle/Legacy):
   - `register_shutdown_function()` - Process lifecycle
   - `register_tick_function()`, `unregister_tick_function()` - Rarely used
   - `create_function()` - Deprecated/removed

## Implementation Notes

### VM Integration Requirements

1. **Execution Context Access**: Functions like `func_num_args()` need access to the current call frame
2. **Function Registry**: `function_exists()` and `get_defined_functions()` need access to the global registry
3. **Dynamic Calling**: `call_user_func()` needs ability to invoke both built-in and user-defined functions
4. **Shutdown Hooks**: Need VM lifecycle integration for shutdown functions

### Error Handling

- Functions that can only be called within function context must validate call frame
- Invalid callbacks must throw proper `Error` exceptions
- Argument validation must match PHP's behavior exactly

### Testing Strategy

- Test argument introspection with various parameter counts
- Test dynamic calling with built-in and user-defined functions
- Test error conditions (invalid callbacks, wrong context)
- Test edge cases (empty arrays, null callbacks)