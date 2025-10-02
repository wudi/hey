# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Hey-Codex is a PHP interpreter written in Go that translates PHP code into bytecode and executes it on a virtual machine. The project aims for PHP 8.0+ compatibility and includes support for static analysis, code transformation, and FastCGI Process Manager (FPM) for web server integration.

## Architecture

The codebase follows a clear pipeline architecture:
```
PHP Source → Lexer → Parser → AST → Compiler → Bytecode → VM Execution
```

### Core Components

- **compiler/lexer**: Tokenizes PHP source code
- **compiler/parser**: Builds Abstract Syntax Tree (AST) from tokens
- **compiler/ast**: Defines all AST node types for PHP constructs
- **compiler**: Transforms AST into bytecode instructions
- **vm**: Virtual machine that executes bytecode
- **registry**: Unified symbol management for functions, classes, and constants
- **runtime**: PHP standard library implementations in Go (65+ string functions)
- **values**: Type system for PHP values in Go
- **opcodes**: VM instruction definitions

### Key Data Structures

- `vm.ExecutionContext`: Manages VM state and execution frames
- `vm.CallFrame`: Represents a function call frame with local variables
- `compiler.Compiler`: Handles bytecode generation with forward jump resolution
- `registry.Registry`: Central repository for all symbols (functions, classes, constants)
- `values.Value`: Universal value type for PHP data

## Development Commands

### Build and Run
```bash
make build              # Build the hey executable
make dev                # Quick build for development (no optimizations)
make run                # Build and run with test command
./build/hey <file.php>  # Execute a PHP file directly
./build/hey -r '<code>' # Execute inline PHP code
./build/hey -a          # Run interactive REPL with readline support
```

### Testing
```bash
make test              # Run all tests
make test-coverage     # Run tests with coverage report
make test-parser       # Run parser tests only
make test-lexer        # Run lexer tests only
make test-vm           # Run VM tests only
make test-compiler     # Run compiler tests only
go test ./runtime -run TestSpecificFunction  # Run specific test
```

### Code Quality
```bash
make fmt               # Format code using gofmt
make lint              # Run golangci-lint (requires installation)
make vet               # Run go vet for static analysis
```

### Development Workflow
```bash
make all               # Build everything (defaults to make build)
make build-all         # Build all binaries including demos
make clean             # Remove build artifacts
make deps              # Download and tidy dependencies
```

## Testing Strategy

- Test files follow `*_test.go` convention
- Major test coverage in:
  - `/vm/vm_test.go` - Core VM functionality tests
  - `/compiler/compiler_test.go` - Bytecode generation tests
  - `/parser/parser_test.go` - PHP parsing tests
- Use table-driven tests for comprehensive coverage
- Test both success cases and error conditions

## Important Implementation Notes

### Critical Bug Fixes

#### WordPress Heisenbug (commit e67da6a) - FETCH_CONSTANT Skipping & DO_FCALL Double Execution
**Three interconnected bugs that prevented WordPress from running:**

1. **Compiler Bug: FETCH_CONSTANT skipped by jumps**
   - Compiler generates FETCH_CONSTANT to load function names into TMP_VAR
   - Control flow jumps (from if-elseif-else optimization) skip these instructions
   - VM reads uninitialized TMP_VAR, gets NULL instead of function name
   - **Fix**: DEFENSIVE FIX in `execInitFCall` - when callee=NULL, search backwards up to 10 instructions to find matching FETCH_CONSTANT, recover function name from constant table
   - **Location**: `vm/instructions.go:2796-2819`

2. **VM Bug: DO_FCALL double execution in nested calls**
   - When nested function calls die/exit, sets Halted=true
   - Inner CallUserFunction loop exits early, but outer loop continues
   - Same DO_FCALL executes twice: first pops pending, second finds pending=nil
   - **Fix**: WORKAROUND in `execDoFCall` - when pending=nil, return true (NOP) instead of error
   - **Location**: `vm/instructions.go:3372-3378`

3. **VM Bug: CallUserFunction doesn't support nesting**
   - Simple loop condition `currentFrame()==child` breaks when child calls other functions
   - **Fix**: Use `childFrameActive()` to check if child is anywhere in CallStack
   - Distinguish child completion (break) vs nested function completion (continue)
   - **Location**: `vm/builtin_context.go:108-163`

**Debugging Methodology:**
- **Heisenbug characteristic**: Adding `echo 1;` made bug disappear
- **Layer-by-layer investigation**: Symptoms → Errors → Call stack → Variables → Bytecode → Compiler output
- **Key discovery tools**: Frame pointer tracking, pendingCalls count, IP sequence analysis
- **Critical insight**: Compiler generates vs VM executes mismatch revealed by comparing debug outputs

**See**: `/docs/debugging-wordpress-heisenbug.md` for complete 488-line debugging guide

#### require_once function registration (commit a956aa0)
- Functions/classes defined in included files must be merged from includeCtx back to parent ctx. Without this, all definitions in included files are lost after execution.

### Bytecode Architecture
- Instructions and operands are stored separately for memory efficiency
- Forward jumps are resolved in a two-pass compilation
- Stack-based VM with explicit CallFrame management

### PHP Compatibility Focus
- Target: PHP 8.0+ features
- Standard library functions implemented in `/stdlib`
- Type system includes PHP's dynamic typing semantics

### Known Architectural Decisions
- `/vm/instructions.go` contains all VM instruction implementations (61k lines - consider refactoring when modifying)
- Registry pattern used for global symbol management to avoid scattered globals
- Compiler uses multiple passes for proper forward reference resolution
- REPL maintains persistent VM context across commands - must reset `Halted` state between executions

### Known Design Issues & Workarounds

**1. Compiler Jump Optimization Can Skip Instructions**
- **Issue**: Control flow jumps may skip FETCH_CONSTANT instructions that initialize TMP_VAR
- **Impact**: INIT_FCALL reads NULL from uninitialized TMP_VAR
- **Workaround**: DEFENSIVE FIX in execInitFCall searches backwards for skipped FETCH_CONSTANT
- **Proper fix**: Compiler should ensure jumps don't skip necessary initialization, or use stack-based parameter passing

**2. Nested CallUserFunction State Management**
- **Issue**: When nested function calls die/exit, Halted=true causes complex loop interactions
- **Impact**: DO_FCALL may execute twice - first pops pending, second finds nil
- **Workaround**: DO_FCALL silently skips when pending=nil (treats as NOP)
- **Proper fix**: Refactor CallUserFunction with isolated execution contexts or use separate VM instances

**3. TMP_VAR vs Stack-Based Parameters**
- **Issue**: TMP_VAR slots are fragile - can be uninitialized if instructions are skipped
- **Current**: Function parameters passed via TMP_VAR references
- **Better approach**: Stack-based parameter passing (like original PHP VM)
- **Tradeoff**: Current approach works with DEFENSIVE FIX, but not architecturally clean

**4. CallUserFunction Loop Complexity**
- **Issue**: Single loop handles both child frame and nested frames execution
- **Complexity**: Must distinguish child completion vs nested completion
- **Current**: childFrameActive() checks if child anywhere in CallStack
- **Better approach**: Separate execution contexts or explicit nesting tracking

**5. Array Element Assignment from Function Return**
- **Issue**: Direct assignment from function return to array element fails in specific WordPress context
- **Pattern**: `$array['key'] = function($array['key'])` doesn't update the array element
- **Workaround**: Use intermediate variable: `$temp = function($array['key']); $array['key'] = $temp;`
- **Context**: Found in WordPress `_wp_die_process_input()` where `$args['charset'] = _canonical_charset($args['charset'])` failed to update
- **Impact**: Array element retains old value instead of function return value
- **Reproduction**: Unable to create minimal reproduction outside WordPress context; bug appears context-dependent
- **Applied fix**: Modified WordPress `/wp-includes/functions.php:4357-4361` to use intermediate variable
- **Proper fix**: Investigate VM's ASSIGN instruction handling when source is function call result and target is array element

### Exception System Architecture
- **Builtin functions can throw catchable PHP exceptions** via `BuiltinCallContext.ThrowException()` API
- Exceptions set `frame.pendingException` and propagate through VM call stack via `raiseException()`
- `ErrExceptionThrown` sentinel error signals to `execDoFCall` that exception was thrown (vs regular Go error)
- Exception classes hierarchy: Exception → (LogicException/RuntimeException/Error), Error → (TypeError/ValueError/AssertionError/etc.)
- Helper function `CreateException(ctx, className, message)` in `/runtime/exception_helpers.go` simplifies exception object creation
- Current exception-throwing builtins: `assert()`, `json_decode()` with JSON_THROW_ON_ERROR, `array_chunk()` with invalid size
- See `/docs/exception-system-design.md` for complete architecture documentation

### Include/Require System Architecture
- **CompilerCallback mechanism** in `/vmfactory/factory.go:52` handles `include/require/include_once/require_once`
- Included files execute in **new ExecutionContext** to isolate compilation but share runtime state
- **Critical**: After execution, must merge `UserFunctions` and `UserClasses` back to parent context (fixed in commit a956aa0)
- Variables, Stack, IncludedFiles are shared between parent and include contexts
- Once-semantics tracked via `ctx.IncludedFiles` map with absolute file paths
- Circular dependency protection: files marked included before execution starts

### Function Parameter Default Values
- Default parameter values are evaluated at compile time using `evaluateConstantExpression`
- Special handling for `true`, `false`, `null` identifiers which are parsed as `IdentifierNode` but need conversion to proper types
- Default values are stored in `registry.Parameter.DefaultValue` and used by VM during function calls
- See `compiler/compiler.go:6160` for identifier constant resolution

## Common Development Tasks

### Adding a New PHP Function (String Functions)
1. **Create PHP validation script**: Write a test script using real PHP to understand exact behavior
2. **Implement in `/runtime/string.go`**: Add function definition to `GetStringFunctions()` array
3. **Add comprehensive tests**: Use TDD approach with tests in `/runtime/string_test.go`
4. **Update documentation**: Mark function as implemented in `/docs/string-functions-spec.md`
5. **Run validation**: `go test ./runtime -run TestStringFunctions`

### Adding Exception Throwing to Builtin Functions
1. **Verify PHP behavior**: Test error conditions with real PHP to confirm exception type (ValueError, JsonException, etc.)
2. **Change function signature**: Replace `_ registry.BuiltinCallContext` with `ctx registry.BuiltinCallContext` to access ThrowException API
3. **Create exception**: Use `CreateException(ctx, "ExceptionClassName", "error message")` helper
4. **Throw exception**: Return `nil, ctx.ThrowException(exception)` instead of returning error value
5. **Test**: Verify exception is catchable with try-catch and normal path still works

### Adding a New VM Instruction
1. Define opcode in `/opcodes/opcodes.go`
2. Implement execution logic in `/vm/instructions.go`
3. Update compiler to generate the instruction
4. Add VM tests for the new instruction

### String Function Development Pattern (TDD)
1. **Red Phase**: Create PHP validation script, add failing test to string_test.go
2. **Green Phase**: Implement function in string.go to make test pass
3. **Refactor Phase**: Optimize implementation while keeping tests green
4. **Document**: Update string-functions-spec.md with implementation status

### Debugging VM Execution

**Environment Variables for Debug Output:**
- `DEBUG_VM` - Detailed execution traces
- `DEBUG_FUNCTIONS` - Function lookup and resolution
- `DEBUG_BYTECODE` - Bytecode generation
- `DEBUG_EXEC` - Instruction execution
- `DEBUG_TMP` - TMP_VAR reads/writes
- `DEBUG_FETCH_CONST` - FETCH_CONSTANT execution

**Common Debugging Techniques:**
- Use `vm.DumpStack()` to inspect stack state
- Track frame pointers to identify same frame across executions
- Monitor `len(frame.pendingCalls)` to detect double execution
- Compare compiler output vs VM execution (find skipped instructions)
- Use IP sequence analysis: record execution order, find jumps

**For Heisenbug-like issues:**
1. **Non-invasive debugging**: Use stderr + environment variables (don't modify code flow)
2. **State tracking**: Log frame=%p, IP=%d, key counters at critical points
3. **Execution comparison**: Run with/without debug, compare IP sequences
4. **Layer analysis**: Trace from symptoms → errors → variables → bytecode → compiler
5. **Compiler-VM verification**: Ensure generated instructions actually execute

**CallFrame includes source position for error reporting**

### Working with REPL
- REPL implementation is in `/cmd/hey/main.go:runInteractiveShell()`
- Uses `github.com/chzyer/readline` for terminal control and history
- History persists in `~/.hey_history`
- Output tracking via custom `trackingWriter` to handle newline formatting
- Multiline detection in `needsMoreInput()` checks for unclosed braces/quotes

## Testing Rules

### PHP Test Case Design
- **必须先使用 `php` 命令验证测试代码**：在设计测试案例时，必须先使用 php 命令测试结果，确定语法正常，并捕获实际输出结果
- 这确保测试案例符合 PHP 标准行为，避免引入错误的期望值
- String functions must pass comprehensive Unicode tests including Greek, Cyrillic, and accented characters
- All edge cases (empty strings, negative indices, null inputs) must be tested

## Current Implementation Status

### String Functions: 65/65+ Complete (100%)
- **Phase 1**: Core functions (strlen, strpos, substr, etc.) ✅
- **Phase 2**: Extended functions (strrpos, stripos, strcmp, etc.) ✅
- **Phase 3**: Advanced functions (encoding, hashing, multi-byte, etc.) ✅
- **Recent additions**: mb_substr(), mb_strtolower(), mb_strtoupper() with full Unicode support
- **Remaining**: Only preg_replace() (complex regex function) marked as PLANNED

### REPL Features
- Full readline support with cursor movement (arrow keys, Ctrl+A/E)
- Line editing capabilities (Ctrl+W, Ctrl+K, Ctrl+U)
- Persistent command history with search (Ctrl+R)
- Multiline input detection for functions/blocks
- Session state persistence across commands

### Exception System: 3 Functions Implemented
- **assert()**: Throws AssertionError when assertion fails (PHP 8+ behavior)
- **json_decode()**: Throws JsonException when JSON_THROW_ON_ERROR flag is set and JSON is invalid
- **array_chunk()**: Throws ValueError when size parameter <= 0
- Architecture supports extending to other builtins via `BuiltinCallContext.ThrowException()` API

### PHP-FPM: FastCGI Process Manager
- **Full FastCGI protocol implementation** in `/pkg/fastcgi` (record parsing, parameter handling, streaming)
- **Process management modes**: static, dynamic, ondemand
- **Master-worker architecture** with graceful shutdown and signal handling (SIGTERM, SIGUSR1, SIGUSR2)
- **CGI variable mapping** to PHP superglobals ($_GET, $_POST, $_SERVER, $_COOKIE, $_REQUEST)
- **Opcache support** for bytecode caching with file validation
- **Health monitoring** with JSON/text status endpoints
- **Entry point**: `./build/hey-fpm --listen 127.0.0.1:9000`
- **Nginx integration**: Compatible with standard FastCGI configurations
- See `/docs/php-fpm-spec.md` for complete architecture documentation