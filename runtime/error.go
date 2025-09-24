package runtime

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// ErrorState manages global error handling state
type ErrorState struct {
	mu             sync.RWMutex
	errorReporting int64
	lastError      *values.Value
	errorHandler   *values.Value
	exceptionHandler *values.Value
}

// Global error state
var globalErrorState = &ErrorState{
	errorReporting: 30719, // E_ALL by default
	lastError:      values.NewNull(),
	errorHandler:   values.NewNull(),
	exceptionHandler: values.NewNull(),
}

// GetErrorFunctions returns all error handling PHP functions
func GetErrorFunctions() []*registry.Function {
	return []*registry.Function{
		{
			Name:       "error_reporting",
			Parameters: []*registry.Parameter{{Name: "level", Type: "int", HasDefault: true, DefaultValue: values.NewNull()}},
			ReturnType: "int",
			MinArgs:    0,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				globalErrorState.mu.Lock()
				defer globalErrorState.mu.Unlock()

				current := globalErrorState.errorReporting
				if len(args) > 0 && !args[0].IsNull() {
					globalErrorState.errorReporting = args[0].ToInt()
				}
				return values.NewInt(current), nil
			},
		},
		{
			Name: "trigger_error",
			Parameters: []*registry.Parameter{
				{Name: "message", Type: "string"},
				{Name: "error_level", Type: "int", HasDefault: true, DefaultValue: values.NewInt(1024)}, // E_USER_NOTICE
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 {
					return values.NewBool(false), nil
				}

				message := args[0].ToString()
				errorLevel := int64(1024) // E_USER_NOTICE default
				if len(args) > 1 {
					errorLevel = args[1].ToInt()
				}

				// Store as last error
				globalErrorState.mu.Lock()
				defer globalErrorState.mu.Unlock()

				_, file, line, ok := runtime.Caller(1)
				if !ok {
					file = ""
					line = 0
				}

				errorArray := values.NewArray()
				errorArray.ArraySet(values.NewString("message"), values.NewString(message))
				errorArray.ArraySet(values.NewString("type"), values.NewInt(errorLevel))
				errorArray.ArraySet(values.NewString("file"), values.NewString(file))
				errorArray.ArraySet(values.NewString("line"), values.NewInt(int64(line)))

				globalErrorState.lastError = errorArray

				// Check if error should be displayed based on error_reporting
				if globalErrorState.errorReporting&errorLevel != 0 {
					// Output error message to stderr like PHP does
					errorTypeName := getErrorTypeName(errorLevel)
					fmt.Fprintf(os.Stderr, "PHP %s: %s in %s on line %d\n", errorTypeName, message, file, line)
				}

				return values.NewBool(true), nil
			},
		},
		{
			Name: "user_error",
			Parameters: []*registry.Parameter{
				{Name: "message", Type: "string"},
				{Name: "error_level", Type: "int", HasDefault: true, DefaultValue: values.NewInt(1024)}, // E_USER_NOTICE
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// user_error is just an alias for trigger_error
				return GetErrorFunctions()[1].Builtin(ctx, args) // Call trigger_error
			},
		},
		{
			Name:       "error_get_last",
			Parameters: []*registry.Parameter{},
			ReturnType: "array|null",
			MinArgs:    0,
			MaxArgs:    0,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				globalErrorState.mu.RLock()
				defer globalErrorState.mu.RUnlock()

				return globalErrorState.lastError, nil
			},
		},
		{
			Name:       "error_clear_last",
			Parameters: []*registry.Parameter{},
			ReturnType: "void",
			MinArgs:    0,
			MaxArgs:    0,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				globalErrorState.mu.Lock()
				defer globalErrorState.mu.Unlock()

				globalErrorState.lastError = values.NewNull()
				return values.NewNull(), nil
			},
		},
		{
			Name:       "set_error_handler",
			Parameters: []*registry.Parameter{
				{Name: "callback", Type: "callable"},
				{Name: "error_levels", Type: "int", HasDefault: true, DefaultValue: values.NewInt(30719)}, // E_ALL
			},
			ReturnType: "callable|null",
			MinArgs:    1,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 {
					return values.NewNull(), nil
				}

				globalErrorState.mu.Lock()
				defer globalErrorState.mu.Unlock()

				previous := globalErrorState.errorHandler
				globalErrorState.errorHandler = args[0]

				return previous, nil
			},
		},
		{
			Name:       "restore_error_handler",
			Parameters: []*registry.Parameter{},
			ReturnType: "bool",
			MinArgs:    0,
			MaxArgs:    0,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				globalErrorState.mu.Lock()
				defer globalErrorState.mu.Unlock()

				globalErrorState.errorHandler = values.NewNull()
				return values.NewBool(true), nil
			},
		},
		{
			Name:       "set_exception_handler",
			Parameters: []*registry.Parameter{
				{Name: "callback", Type: "callable"},
			},
			ReturnType: "callable|null",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 {
					return values.NewNull(), nil
				}

				globalErrorState.mu.Lock()
				defer globalErrorState.mu.Unlock()

				previous := globalErrorState.exceptionHandler
				globalErrorState.exceptionHandler = args[0]

				return previous, nil
			},
		},
		{
			Name:       "restore_exception_handler",
			Parameters: []*registry.Parameter{},
			ReturnType: "bool",
			MinArgs:    0,
			MaxArgs:    0,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				globalErrorState.mu.Lock()
				defer globalErrorState.mu.Unlock()

				globalErrorState.exceptionHandler = values.NewNull()
				return values.NewBool(true), nil
			},
		},
		{
			Name:       "debug_backtrace",
			Parameters: []*registry.Parameter{
				{Name: "options", Type: "int", HasDefault: true, DefaultValue: values.NewInt(1)}, // DEBUG_BACKTRACE_PROVIDE_OBJECT
				{Name: "limit", Type: "int", HasDefault: true, DefaultValue: values.NewInt(0)},
			},
			ReturnType: "array",
			MinArgs:    0,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				limit := 0
				if len(args) > 1 && !args[1].IsNull() {
					limit = int(args[1].ToInt())
				}

				// Get Go runtime stack trace
				pc := make([]uintptr, 50)
				n := runtime.Callers(2, pc) // Skip current frame and Callers itself
				if limit > 0 && limit < n {
					n = limit
				}

				result := values.NewArray()
				frames := runtime.CallersFrames(pc[:n])

				index := int64(0)
				for {
					frame, more := frames.Next()
					if !more {
						break
					}

					frameArray := values.NewArray()
					frameArray.ArraySet(values.NewString("function"), values.NewString(extractFunctionName(frame.Function)))
					frameArray.ArraySet(values.NewString("file"), values.NewString(frame.File))
					frameArray.ArraySet(values.NewString("line"), values.NewInt(int64(frame.Line)))

					result.ArraySet(values.NewInt(index), frameArray)
					index++

					if !more {
						break
					}
				}

				return result, nil
			},
		},
		{
			Name:       "debug_print_backtrace",
			Parameters: []*registry.Parameter{
				{Name: "options", Type: "int", HasDefault: true, DefaultValue: values.NewInt(0)},
				{Name: "limit", Type: "int", HasDefault: true, DefaultValue: values.NewInt(0)},
			},
			ReturnType: "void",
			MinArgs:    0,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				limit := 0
				if len(args) > 1 && !args[1].IsNull() {
					limit = int(args[1].ToInt())
				}

				// Get Go runtime stack trace
				pc := make([]uintptr, 50)
				n := runtime.Callers(2, pc) // Skip current frame and Callers itself
				if limit > 0 && limit < n {
					n = limit
				}

				frames := runtime.CallersFrames(pc[:n])
				index := 0
				for {
					frame, more := frames.Next()
					if !more {
						break
					}

					fmt.Printf("#%d %s:%d: %s()\n", index, frame.File, frame.Line, extractFunctionName(frame.Function))
					index++

					if !more {
						break
					}
				}

				return values.NewNull(), nil
			},
		},
		{
			Name:       "error_log",
			Parameters: []*registry.Parameter{
				{Name: "message", Type: "string"},
				{Name: "message_type", Type: "int", HasDefault: true, DefaultValue: values.NewInt(0)},
				{Name: "destination", Type: "string", HasDefault: true, DefaultValue: values.NewNull()},
				{Name: "extra_headers", Type: "string", HasDefault: true, DefaultValue: values.NewNull()},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    4,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 {
					return values.NewBool(false), nil
				}

				message := args[0].ToString()
				messageType := int64(0)
				if len(args) > 1 {
					messageType = args[1].ToInt()
				}

				switch messageType {
				case 0: // Send to system logger
					log.Println("[PHP]", message)
				case 1: // Send by email (not implemented, just log)
					log.Println("[PHP Email]", message)
				case 3: // Write to file (not implemented, just log)
					log.Println("[PHP File]", message)
				case 4: // Write to SAPI log (just log)
					log.Println("[PHP SAPI]", message)
				default:
					log.Println("[PHP]", message)
				}

				return values.NewBool(true), nil
			},
		},
	}
}

// getErrorTypeName converts error level to human readable name
func getErrorTypeName(level int64) string {
	switch level {
	case 1:
		return "Fatal error"
	case 2:
		return "Warning"
	case 4:
		return "Parse error"
	case 8:
		return "Notice"
	case 16:
		return "Core error"
	case 32:
		return "Core warning"
	case 64:
		return "Compile error"
	case 128:
		return "Compile warning"
	case 256:
		return "Fatal error" // E_USER_ERROR
	case 512:
		return "Warning"    // E_USER_WARNING
	case 1024:
		return "Notice" // E_USER_NOTICE
	case 2048:
		return "Strict Standards"
	case 4096:
		return "Catchable fatal error"
	case 8192:
		return "Deprecated"
	case 16384:
		return "Deprecated" // E_USER_DEPRECATED
	default:
		return "Unknown error"
	}
}

// extractFunctionName extracts the function name from a full Go function path
func extractFunctionName(fullName string) string {
	parts := strings.Split(fullName, ".")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return fullName
}