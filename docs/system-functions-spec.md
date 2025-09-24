# PHP System Functions Specification

This document tracks the implementation status of PHP's system program execution functions in Hey-Codex.

## Functions Overview

PHP provides several functions for executing system commands and interacting with the operating system. This implementation focuses on the core program execution functions.

## Implementation Status

### ✅ Completed Functions (13/13) 🎯

| Function | Status | Description | Test Coverage |
|----------|--------|-------------|---------------|
| `escapeshellarg` | ✅ | Escape a string to be used as a shell argument | ✅ |
| `escapeshellcmd` | ✅ | Escape shell metacharacters | ✅ |
| `shell_exec` | ✅ | Execute command via shell and return output | ✅ |
| `exec` | ✅ | Execute external program, return last line | ✅ |
| `system` | ✅ | Execute external program and display output | ✅ |
| `passthru` | ✅ | Execute external program and display raw output | ✅ |
| `getenv` | ✅ | Get environment variable value | ✅ |
| `putenv` | ✅ | Set environment variable | ✅ |
| `proc_open` | ✅ | Execute command and open file pointers | ✅ |
| `proc_close` | ✅ | Close a process opened by proc_open | ✅ |
| `proc_get_status` | ✅ | Get information about a process | ✅ |
| `proc_terminate` | ✅ | Kill a process opened by proc_open | ✅ |
| `proc_nice` | ✅ | Change process priority | ✅ |

**🎉 All system execution functions have been successfully implemented!**

## Function Details

### escapeshellarg(string $arg): string

Escapes a string to be used as a shell argument by wrapping it in single quotes and escaping any embedded single quotes.

**Implementation Notes:**
- Wraps the argument in single quotes
- Escapes single quotes using the pattern: `'arg'` → `'arg'\\''more'`
- Handles empty strings by returning `''`

**Test Cases:**
- Basic strings with spaces
- Strings containing single quotes, double quotes, and special characters
- Empty strings
- Strings with tabs and newlines

### escapeshellcmd(string $command): string

Escapes shell metacharacters to prevent command injection.

**Implementation Notes:**
- Escapes dangerous characters: `#&;`|*?~<>^()[]{}$\\'"\n\t`
- Does NOT escape spaces (unlike some other implementations)
- Uses backslash escaping for each metacharacter

**Test Cases:**
- Commands with semicolons and pipes
- Commands with various metacharacters
- Empty commands
- Normal commands without metacharacters

### shell_exec(string $command): string|null

Executes a command via shell and returns the complete output as a string.

**Implementation Notes:**
- Uses `sh -c` for command execution
- Returns stdout as string, including stderr if stdout is empty
- Returns null for empty commands
- Preserves newlines in output

**Test Cases:**
- Basic echo commands
- Commands with stderr output
- Empty commands
- Multi-line output commands

### exec(string $command, array &$output = null, int &$result_code = null): string|false

Executes an external program and returns the last line of output.

**Implementation Notes:**
- Populates `$output` array with all output lines (by reference)
- Sets `$result_code` with the command's exit status (by reference)
- Returns the last line of output as a string
- Handles stderr by including it in output when stdout is empty

**Test Cases:**
- Single-line output commands
- Multi-line output commands
- Commands with non-zero exit codes
- Commands with stderr output

### system(string $command, int &$result_code = null): string|false

Executes an external program and displays output directly.

**Implementation Notes:**
- Outputs directly to stdout/stderr (inherits current process streams)
- Sets `$result_code` with exit status (by reference)
- Returns empty string (since output goes directly to terminal)
- Cannot capture output for return value due to direct output

**Test Cases:**
- Commands with successful execution
- Commands with error conditions
- Return code verification

### passthru(string $command, int &$result_code = null): void

Executes an external program and displays raw output. Similar to `system()` but returns void.

**Implementation Notes:**
- Identical to `system()` but returns null instead of string
- Direct output to stdout/stderr
- Sets result code by reference
- Useful when you need raw binary output

**Test Cases:**
- Basic command execution
- Return code verification
- Void return value verification

### proc_open(cmd, descriptorspec, &$pipes, $cwd, $env, $other): resource|false

Execute a command and open file pointers for input/output with full control over process I/O.

**Implementation Notes:**
- Creates process with configurable stdin/stdout/stderr
- Supports pipe, file, and inherited I/O descriptors
- Returns resource handle for process management
- Maintains global process registry for handle tracking
- Supports custom working directory and environment variables
- Pipes array populated with I/O handles by reference

**Test Cases:**
- Basic process creation with pipes
- File descriptor redirection
- Custom environment variables
- Working directory configuration
- Process handle validation

### proc_close(resource $process): int

Close a process opened by proc_open and return its exit code.

**Implementation Notes:**
- Closes all associated pipes and file handles
- Waits for process completion
- Returns process exit code
- Removes process from global registry
- Handles both normal and error exits

**Test Cases:**
- Normal process closure
- Exit code capture
- Resource cleanup verification

### proc_get_status(resource $process): array

Get detailed information about a process opened by proc_open.

**Implementation Notes:**
- Returns associative array with process information
- Fields: command, pid, running, signaled, stopped, exitcode, termsig, stopsig
- Real-time status checking
- Non-blocking operation
- Works with active and terminated processes

**Test Cases:**
- Running process status
- Terminated process status
- Field validation
- PID retrieval

### proc_terminate(resource $process, int $signal = 15): bool

Sends a signal to a process opened by proc_open.

**Implementation Notes:**
- Default signal is SIGTERM (15)
- Supports custom signal numbers
- Returns true on successful signal delivery
- Updates process signaled/termsig fields
- Non-blocking signal delivery

**Test Cases:**
- SIGTERM delivery
- Custom signal support
- Process termination verification
- Invalid process handling

### proc_nice(int $increment): bool|int

Change the priority of the current process.

**Implementation Notes:**
- Returns current priority when increment is 0
- Positive values decrease priority (nice)
- Negative values increase priority (requires privileges)
- Uses Unix nice values (-20 to 19)
- May fail without appropriate permissions

**Test Cases:**
- Current priority retrieval
- Priority increment
- Permission handling
- Return value validation

## Security Considerations

All execution functions in this implementation:
- Execute commands using `sh -c` on Unix systems
- Do NOT perform automatic escaping - use `escapeshellarg()` and `escapeshellcmd()`
- Inherit the parent process's environment and permissions
- Can execute any command the PHP process has permission to run

**Security Best Practices:**
1. Always validate and sanitize user input
2. Use `escapeshellarg()` for user-provided arguments
3. Use `escapeshellcmd()` for user-provided commands
4. Consider using allow-lists for permitted commands
5. Run PHP processes with minimal necessary privileges

## Testing

All functions have comprehensive test coverage including:
- Happy path scenarios
- Error conditions
- Edge cases (empty inputs, special characters)
- Return value validation
- Reference parameter behavior (where applicable)
- Cross-platform compatibility considerations

## Performance Notes

- `shell_exec()` and `exec()` capture output, requiring memory allocation
- `system()` and `passthru()` stream output directly, using less memory
- All functions block until command completion
- No timeout mechanism implemented (relies on OS-level timeouts)

## Future Enhancements

Potential improvements for future versions:
1. Implement `proc_*` functions for advanced process management
2. Add timeout support for long-running commands
3. Implement Windows-specific command execution paths
4. Add async execution capabilities
5. Enhanced error reporting and logging