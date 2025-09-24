package runtime

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// Process represents an open process handle
type Process struct {
	cmd       *exec.Cmd
	stdin     io.WriteCloser
	stdout    io.ReadCloser
	stderr    io.ReadCloser
	pipes     []io.Closer
	startTime time.Time
	exitCode  int
	signaled  bool
	stopped   bool
	termsig   int
	stopsig   int
	closed    atomic.Bool
}

// Global process registry
var (
	processRegistry = make(map[int64]*Process)
	processCounter  int64
	processMutex    sync.RWMutex
)

// GetSystemFunctions returns system-related PHP functions
func GetSystemFunctions() []*registry.Function {
	return []*registry.Function{
		{
			Name: "getenv",
			Parameters: []*registry.Parameter{
				{Name: "varname", Type: "string", HasDefault: true, DefaultValue: values.NewNull()},
			},
			ReturnType: "string|array|false",
			MinArgs:    0,
			MaxArgs:    1,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || (len(args) > 0 && args[0].IsNull()) {
					// Return all environment variables as an associative array
					envVars := os.Environ()
					result := values.NewArray()
					arr := result.Data.(*values.Array)

					for _, env := range envVars {
						parts := strings.SplitN(env, "=", 2)
						if len(parts) == 2 {
							arr.Elements[parts[0]] = values.NewString(parts[1])
						}
					}
					return result, nil
				}

				// Get specific environment variable
				varname := args[0].ToString()
				value, exists := os.LookupEnv(varname)

				if !exists {
					return values.NewBool(false), nil
				}

				return values.NewString(value), nil
			},
		},
		{
			Name: "putenv",
			Parameters: []*registry.Parameter{
				{Name: "setting", Type: "string"},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 {
					return values.NewBool(false), nil
				}

				setting := args[0].ToString()

				// Parse the setting string (format: "NAME=value")
				parts := strings.SplitN(setting, "=", 2)
				if len(parts) != 2 {
					return values.NewBool(false), nil
				}

				name := parts[0]
				value := parts[1]

				// Set the environment variable
				err := os.Setenv(name, value)
				if err != nil {
					return values.NewBool(false), nil
				}

				return values.NewBool(true), nil
			},
		},
		{
			Name:       "getcwd",
			Parameters: []*registry.Parameter{},
			ReturnType: "string|false",
			MinArgs:    0,
			MaxArgs:    0,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// Get the current working directory
				cwd, err := os.Getwd()
				if err != nil {
					// Return false on error (PHP behavior)
					return values.NewBool(false), nil
				}
				return values.NewString(cwd), nil
			},
		},
		{
			Name: "exit",
			Parameters: []*registry.Parameter{
				{Name: "status", Type: "mixed", HasDefault: true, DefaultValue: values.NewInt(0)},
			},
			ReturnType: "void",
			MinArgs:    0,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				var exitCode int
				var message string

				if len(args) > 0 && args[0] != nil {
					if args[0].IsString() {
						// String argument: print message and exit with code 0
						message = args[0].ToString()
						exitCode = 0
					} else {
						// Numeric argument: exit with this code
						exitCode = int(args[0].ToInt())
					}
				} else {
					// No argument: exit with code 0
					exitCode = 0
				}

				// Halt execution
				ctx.Halt(exitCode, message)
				return values.NewNull(), nil
			},
		},
		{
			Name: "die",
			Parameters: []*registry.Parameter{
				{Name: "status", Type: "mixed", HasDefault: true, DefaultValue: values.NewInt(0)},
			},
			ReturnType: "void",
			MinArgs:    0,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				var exitCode int
				var message string

				if len(args) > 0 && args[0] != nil {
					if args[0].IsString() {
						// String argument: print message and exit with code 0
						message = args[0].ToString()
						exitCode = 0
					} else {
						// Numeric argument: exit with this code
						exitCode = int(args[0].ToInt())
					}
				} else {
					// No argument: exit with code 0
					exitCode = 0
				}

				// Halt execution (die is an alias of exit)
				ctx.Halt(exitCode, message)
				return values.NewNull(), nil
			},
		},
		{
			Name: "escapeshellarg",
			Parameters: []*registry.Parameter{
				{Name: "arg", Type: "string"},
			},
			ReturnType: "string",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 {
					return values.NewString("''"), nil
				}

				arg := args[0].ToString()

				// PHP's escapeshellarg wraps the argument in single quotes
				// and escapes any single quotes by ending the quote, adding
				// a backslash-escaped single quote, then starting a new quote
				escaped := "'" + strings.ReplaceAll(arg, "'", "'\\''") + "'"
				return values.NewString(escaped), nil
			},
		},
		{
			Name: "escapeshellcmd",
			Parameters: []*registry.Parameter{
				{Name: "command", Type: "string"},
			},
			ReturnType: "string",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 {
					return values.NewString(""), nil
				}

				cmd := args[0].ToString()

				// Characters that need to be escaped according to PHP escapeshellcmd behavior
				// Based on PHP source, these are the dangerous metacharacters
				metaChars := "#&;`|*?~<>^()[]{}$\\'\"\n\t"

				var escaped strings.Builder
				for _, char := range cmd {
					if strings.ContainsRune(metaChars, char) {
						escaped.WriteByte('\\')
					}
					escaped.WriteRune(char)
				}

				return values.NewString(escaped.String()), nil
			},
		},
		{
			Name: "shell_exec",
			Parameters: []*registry.Parameter{
				{Name: "command", Type: "string"},
			},
			ReturnType: "string|null",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 {
					return values.NewNull(), nil
				}

				command := args[0].ToString()
				if command == "" {
					return values.NewNull(), nil
				}

				// Execute the command using shell
				cmd := exec.Command("sh", "-c", command)

				var stdout, stderr bytes.Buffer
				cmd.Stdout = &stdout
				cmd.Stderr = &stderr

				err := cmd.Run()
				if err != nil {
					// If there's stderr output, include it
					if stderr.Len() > 0 {
						return values.NewString(stderr.String()), nil
					}
					// Command failed but no stderr - return empty string
					return values.NewString(""), nil
				}

				return values.NewString(stdout.String()), nil
			},
		},
		{
			Name: "exec",
			Parameters: []*registry.Parameter{
				{Name: "command", Type: "string"},
				{Name: "output", Type: "array", HasDefault: true, DefaultValue: values.NewArray()},
				{Name: "result_code", Type: "int", HasDefault: true, DefaultValue: values.NewInt(0)},
			},
			ReturnType: "string|false",
			MinArgs:    1,
			MaxArgs:    3,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 {
					return values.NewBool(false), nil
				}

				command := args[0].ToString()
				if command == "" {
					return values.NewBool(false), nil
				}

				// Execute the command
				cmd := exec.Command("sh", "-c", command)

				var stdout, stderr bytes.Buffer
				cmd.Stdout = &stdout
				cmd.Stderr = &stderr

				err := cmd.Run()

				// Get exit code
				exitCode := 0
				if err != nil {
					if exitError, ok := err.(*exec.ExitError); ok {
						exitCode = exitError.ExitCode()
					} else {
						exitCode = 1
					}
				}

				// Prepare output lines
				outputLines := strings.Split(strings.TrimSuffix(stdout.String(), "\n"), "\n")
				if len(outputLines) == 1 && outputLines[0] == "" {
					outputLines = []string{}
				}

				// If stderr exists and stdout is empty, use stderr
				if stderr.Len() > 0 && stdout.Len() == 0 {
					outputLines = strings.Split(strings.TrimSuffix(stderr.String(), "\n"), "\n")
				} else if stderr.Len() > 0 {
					// Append stderr to stdout
					stderrLines := strings.Split(strings.TrimSuffix(stderr.String(), "\n"), "\n")
					outputLines = append(outputLines, stderrLines...)
				}

				// Set output array if provided
				if len(args) > 1 && args[1] != nil {
					outputArray := args[1]
					if outputArray.IsArray() {
						arr := outputArray.Data.(*values.Array)
						// Clear the array
						arr.Elements = make(map[interface{}]*values.Value)
						arr.NextIndex = 0

						// Add output lines
						for i, line := range outputLines {
							arr.Elements[int64(i)] = values.NewString(line)
						}
						arr.NextIndex = int64(len(outputLines))
					}
				}

				// Set return code if provided
				if len(args) > 2 && args[2] != nil {
					// In PHP, this would modify the variable by reference
					// For now, we'll just set the value
					args[2].Data = int64(exitCode)
					args[2].Type = values.TypeInt
				}

				// Return last line or empty string
				lastLine := ""
				if len(outputLines) > 0 {
					lastLine = outputLines[len(outputLines)-1]
				}

				return values.NewString(lastLine), nil
			},
		},
		{
			Name: "system",
			Parameters: []*registry.Parameter{
				{Name: "command", Type: "string"},
				{Name: "result_code", Type: "int", HasDefault: true, DefaultValue: values.NewInt(0)},
			},
			ReturnType: "string|false",
			MinArgs:    1,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 {
					return values.NewBool(false), nil
				}

				command := args[0].ToString()
				if command == "" {
					return values.NewBool(false), nil
				}

				// Execute command
				cmd := exec.Command("sh", "-c", command)

				// system() outputs directly to stdout/stderr, so we inherit them
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr

				err := cmd.Run()

				// Get exit code
				exitCode := 0
				if err != nil {
					if exitError, ok := err.(*exec.ExitError); ok {
						exitCode = exitError.ExitCode()
					} else {
						exitCode = 1
					}
				}

				// Set return code if provided
				if len(args) > 1 && args[1] != nil {
					args[1].Data = int64(exitCode)
					args[1].Type = values.TypeInt
				}

				// In PHP, system() returns the last line of output, but since we're
				// directing output to stdout, we can't capture it. We'll return empty
				// string for now (this matches some PHP behavior when output is redirected)
				return values.NewString(""), nil
			},
		},
		{
			Name: "passthru",
			Parameters: []*registry.Parameter{
				{Name: "command", Type: "string"},
				{Name: "result_code", Type: "int", HasDefault: true, DefaultValue: values.NewInt(0)},
			},
			ReturnType: "void",
			MinArgs:    1,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 {
					return values.NewNull(), nil
				}

				command := args[0].ToString()
				if command == "" {
					return values.NewNull(), nil
				}

				// Execute command - passthru is like system but returns void
				cmd := exec.Command("sh", "-c", command)

				// passthru outputs directly to stdout/stderr
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr

				err := cmd.Run()

				// Get exit code
				exitCode := 0
				if err != nil {
					if exitError, ok := err.(*exec.ExitError); ok {
						exitCode = exitError.ExitCode()
					} else {
						exitCode = 1
					}
				}

				// Set return code if provided
				if len(args) > 1 && args[1] != nil {
					args[1].Data = int64(exitCode)
					args[1].Type = values.TypeInt
				}

				return values.NewNull(), nil
			},
		},
		{
			Name: "proc_open",
			Parameters: []*registry.Parameter{
				{Name: "command", Type: "string|array"},
				{Name: "descriptorspec", Type: "array"},
				{Name: "pipes", Type: "array"},
				{Name: "cwd", Type: "string", HasDefault: true, DefaultValue: values.NewNull()},
				{Name: "env", Type: "array", HasDefault: true, DefaultValue: values.NewNull()},
				{Name: "other_options", Type: "array", HasDefault: true, DefaultValue: values.NewNull()},
			},
			ReturnType: "resource|false",
			MinArgs:    3,
			MaxArgs:    6,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 3 {
					return values.NewBool(false), nil
				}

				// Get command
				var command string
				if args[0].IsArray() {
					// TODO: Handle array command format
					return values.NewBool(false), nil
				} else {
					command = args[0].ToString()
				}

				// Parse descriptor spec
				descriptorSpec := args[1]
				if !descriptorSpec.IsArray() {
					return values.NewBool(false), nil
				}

				// Create command
				cmd := exec.Command("sh", "-c", command)

				// Set working directory if provided
				if len(args) > 3 && !args[3].IsNull() {
					cmd.Dir = args[3].ToString()
				}

				// Set environment if provided
				if len(args) > 4 && !args[4].IsNull() && args[4].IsArray() {
					envArray := args[4].Data.(*values.Array)
					var envVars []string
					for key, val := range envArray.Elements {
						envVars = append(envVars, fmt.Sprintf("%v=%s", key, val.ToString()))
					}
					cmd.Env = envVars
				} else {
					cmd.Env = os.Environ()
				}

				// Create process handle
				proc := &Process{
					cmd:       cmd,
					startTime: time.Now(),
					exitCode:  -1,
				}

				// Setup pipes based on descriptor spec
				pipesArray := args[2]
				if !pipesArray.IsArray() {
					return values.NewBool(false), nil
				}
				pipeArr := pipesArray.Data.(*values.Array)

				// Clear pipes array
				pipeArr.Elements = make(map[interface{}]*values.Value)
				pipeArr.NextIndex = 0

				descArr := descriptorSpec.Data.(*values.Array)

				// Process each descriptor
				for fdNum, descriptor := range descArr.Elements {
					if !descriptor.IsArray() {
						continue
					}

					descArray := descriptor.Data.(*values.Array)
					if len(descArray.Elements) < 2 {
						continue
					}

					// Get descriptor type
					descType := ""
					if typeVal, ok := descArray.Elements[int64(0)]; ok {
						descType = typeVal.ToString()
					}

					switch fdNum {
					case int64(0): // stdin
						if descType == "pipe" {
							stdin, err := cmd.StdinPipe()
							if err != nil {
								return values.NewBool(false), nil
							}
							proc.stdin = stdin
							// Store pipe handle (simplified - would need proper resource type)
							pipeArr.Elements[int64(0)] = values.NewResource(stdin)
						}
					case int64(1): // stdout
						if descType == "pipe" {
							stdout, err := cmd.StdoutPipe()
							if err != nil {
								return values.NewBool(false), nil
							}
							proc.stdout = stdout
							pipeArr.Elements[int64(1)] = values.NewResource(stdout)
						} else if descType == "file" {
							// Handle file output
							if fileVal, ok := descArray.Elements[int64(1)]; ok {
								filename := fileVal.ToString()
								file, err := os.Create(filename)
								if err != nil {
									return values.NewBool(false), nil
								}
								cmd.Stdout = file
								proc.pipes = append(proc.pipes, file)
							}
						}
					case int64(2): // stderr
						if descType == "pipe" {
							stderr, err := cmd.StderrPipe()
							if err != nil {
								return values.NewBool(false), nil
							}
							proc.stderr = stderr
							pipeArr.Elements[int64(2)] = values.NewResource(stderr)
						} else if descType == "file" {
							// Handle file output
							if fileVal, ok := descArray.Elements[int64(1)]; ok {
								filename := fileVal.ToString()
								file, err := os.Create(filename)
								if err != nil {
									return values.NewBool(false), nil
								}
								cmd.Stderr = file
								proc.pipes = append(proc.pipes, file)
							}
						}
					}
				}

				// Start the process
				if err := cmd.Start(); err != nil {
					return values.NewBool(false), nil
				}

				// Register process
				processMutex.Lock()
				processCounter++
				procID := processCounter
				processRegistry[procID] = proc
				processMutex.Unlock()

				// Return resource handle
				return values.NewResource(procID), nil
			},
		},
		{
			Name: "proc_close",
			Parameters: []*registry.Parameter{
				{Name: "process", Type: "resource"},
			},
			ReturnType: "int",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || !args[0].IsResource() {
					return values.NewInt(-1), nil
				}

				procID, ok := args[0].Data.(int64)
				if !ok {
					return values.NewInt(-1), nil
				}

				processMutex.Lock()
				proc, exists := processRegistry[procID]
				if !exists {
					processMutex.Unlock()
					return values.NewInt(-1), nil
				}
				delete(processRegistry, procID)
				processMutex.Unlock()

				// Mark as closed
				proc.closed.Store(true)

				// Close all pipes
				if proc.stdin != nil {
					proc.stdin.Close()
				}
				if proc.stdout != nil {
					proc.stdout.Close()
				}
				if proc.stderr != nil {
					proc.stderr.Close()
				}
				for _, pipe := range proc.pipes {
					pipe.Close()
				}

				// Wait for process to finish
				err := proc.cmd.Wait()
				if err != nil {
					if exitError, ok := err.(*exec.ExitError); ok {
						return values.NewInt(int64(exitError.ExitCode())), nil
					}
					return values.NewInt(1), nil
				}

				return values.NewInt(0), nil
			},
		},
		{
			Name: "proc_get_status",
			Parameters: []*registry.Parameter{
				{Name: "process", Type: "resource"},
			},
			ReturnType: "array",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || !args[0].IsResource() {
					return values.NewBool(false), nil
				}

				procID, ok := args[0].Data.(int64)
				if !ok {
					return values.NewBool(false), nil
				}

				processMutex.RLock()
				proc, exists := processRegistry[procID]
				processMutex.RUnlock()

				if !exists {
					return values.NewBool(false), nil
				}

				// Create status array
				status := values.NewArray()
				arr := status.Data.(*values.Array)

				// Set command
				arr.Elements["command"] = values.NewString(strings.Join(proc.cmd.Args, " "))

				// Set PID
				pid := 0
				if proc.cmd.Process != nil {
					pid = proc.cmd.Process.Pid
				}
				arr.Elements["pid"] = values.NewInt(int64(pid))

				// Check if running
				running := false
				if proc.cmd.Process != nil && !proc.closed.Load() {
					// Check if process is still running
					if proc.cmd.ProcessState == nil {
						running = true
					}
				}
				arr.Elements["running"] = values.NewBool(running)

				// Set exit status
				if proc.cmd.ProcessState != nil {
					arr.Elements["exitcode"] = values.NewInt(int64(proc.cmd.ProcessState.ExitCode()))
				} else {
					arr.Elements["exitcode"] = values.NewInt(-1)
				}

				// Set signal info
				arr.Elements["signaled"] = values.NewBool(proc.signaled)
				arr.Elements["stopped"] = values.NewBool(proc.stopped)
				arr.Elements["termsig"] = values.NewInt(int64(proc.termsig))
				arr.Elements["stopsig"] = values.NewInt(int64(proc.stopsig))

				return status, nil
			},
		},
		{
			Name: "proc_terminate",
			Parameters: []*registry.Parameter{
				{Name: "process", Type: "resource"},
				{Name: "signal", Type: "int", HasDefault: true, DefaultValue: values.NewInt(15)}, // SIGTERM
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || !args[0].IsResource() {
					return values.NewBool(false), nil
				}

				procID, ok := args[0].Data.(int64)
				if !ok {
					return values.NewBool(false), nil
				}

				signal := syscall.SIGTERM
				if len(args) > 1 {
					sigNum := int(args[1].ToInt())
					signal = syscall.Signal(sigNum)
				}

				processMutex.RLock()
				proc, exists := processRegistry[procID]
				processMutex.RUnlock()

				if !exists || proc.cmd.Process == nil {
					return values.NewBool(false), nil
				}

				// Send signal to process
				err := proc.cmd.Process.Signal(signal)
				if err != nil {
					return values.NewBool(false), nil
				}

				// Mark as signaled if terminated
				if signal == syscall.SIGTERM || signal == syscall.SIGKILL {
					proc.signaled = true
					proc.termsig = int(signal)
				}

				return values.NewBool(true), nil
			},
		},
		{
			Name: "proc_nice",
			Parameters: []*registry.Parameter{
				{Name: "increment", Type: "int"},
			},
			ReturnType: "bool|int",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 {
					return values.NewBool(false), nil
				}

				increment := int(args[0].ToInt())

				// Get current priority
				currentPrio, err := syscall.Getpriority(syscall.PRIO_PROCESS, 0)
				if err != nil {
					return values.NewBool(false), nil
				}

				// If increment is 0, return current priority
				if increment == 0 {
					// The priority value needs adjustment as getpriority returns 20-nice
					return values.NewInt(int64(20 - currentPrio)), nil
				}

				// Set new priority (nice value)
				// Note: In Unix, lower numeric values mean higher priority
				// nice value range is typically -20 to 19
				newPrio := currentPrio - increment
				err = syscall.Setpriority(syscall.PRIO_PROCESS, 0, newPrio)
				if err != nil {
					return values.NewBool(false), nil
				}

				return values.NewBool(true), nil
			},
		},
		{
			Name: "set_time_limit",
			Parameters: []*registry.Parameter{
				{Name: "seconds", Type: "int"},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 {
					return values.NewBool(false), nil
				}

				seconds := int(args[0].ToInt())

				// Get the execution context from the builtin context
				if execCtx := ctx.GetExecutionContext(); execCtx != nil {
					// SetTimeLimit returns true on success, consistent with PHP behavior
					success := execCtx.SetTimeLimit(seconds)

					if success {
						// Also update the ini setting to keep it in sync
						UpdateIniMaxExecutionTime(seconds)
					}

					return values.NewBool(success), nil
				}

				return values.NewBool(false), nil
			},
		},
	}
}