package spl

import (
	"fmt"
	"strings"
	"sync"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// Global state for autoload functionality
var (
	autoloadMutex      sync.RWMutex
	autoloadFunctions  []*values.Value // Registered autoload functions
	autoloadExtensions string          // File extensions for default autoloader
)

func init() {
	// Initialize with default PHP autoload extensions
	autoloadExtensions = ".inc,.php"
	autoloadFunctions = make([]*values.Value, 0)
}

// GetSplAutoloadFunctions returns all SPL autoload functions
func GetSplAutoloadFunctions() []*registry.Function {
	return []*registry.Function{
		getSplAutoloadFunction(),
		getSplAutoloadCallFunction(),
		getSplAutoloadExtensionsFunction(),
		getSplAutoloadFunctionsFunction(),
		getSplAutoloadRegisterFunction(),
		getSplAutoloadUnregisterFunction(),
	}
}

// spl_autoload() - Default autoload implementation
func getSplAutoloadFunction() *registry.Function {
	return &registry.Function{
		Name:      "spl_autoload",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 {
				return values.NewNull(), fmt.Errorf("spl_autoload() expects at least 1 parameter, %d given", len(args))
			}

			className := args[0].ToString()

			// Get file extensions (optional second parameter)
			var fileExtensions string
			if len(args) > 1 && !args[1].IsNull() {
				fileExtensions = args[1].ToString()
			} else {
				autoloadMutex.RLock()
				fileExtensions = autoloadExtensions
				autoloadMutex.RUnlock()
			}

			// Default autoloader attempts to include files based on class name + extensions
			// In a real implementation, this would attempt to include files
			// For now, we'll just indicate the attempt was made

			extensions := strings.Split(fileExtensions, ",")
			for _, ext := range extensions {
				fileName := strings.ToLower(className) + strings.TrimSpace(ext)
				// In real implementation: attempt include/require fileName
				// For simulation: just indicate attempt
				_ = fileName
			}

			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{
			{Name: "class_name", Type: "string"},
			{Name: "file_extensions", Type: "string", DefaultValue: values.NewNull()},
		},
	}
}

// spl_autoload_call() - Call all registered autoload functions
func getSplAutoloadCallFunction() *registry.Function {
	return &registry.Function{
		Name:      "spl_autoload_call",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 {
				return values.NewNull(), fmt.Errorf("spl_autoload_call() expects exactly 1 parameter, %d given", len(args))
			}

			className := args[0].ToString()

			autoloadMutex.RLock()
			functions := make([]*values.Value, len(autoloadFunctions))
			copy(functions, autoloadFunctions)
			autoloadMutex.RUnlock()

			// Call each registered autoload function
			for _, autoloadFunc := range functions {
				if autoloadFunc.IsCallable() {
					// Call the autoload function with the class name
					// In a full implementation, this would use CallUserFunction
					// For now, we'll simulate the call by using the className
					_ = className // Use className to avoid unused variable error
					continue
				}
			}

			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{
			{Name: "class_name", Type: "string"},
		},
	}
}

// spl_autoload_extensions() - Get or set autoload file extensions
func getSplAutoloadExtensionsFunction() *registry.Function {
	return &registry.Function{
		Name:      "spl_autoload_extensions",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) == 0 {
				// Get current extensions
				autoloadMutex.RLock()
				current := autoloadExtensions
				autoloadMutex.RUnlock()
				return values.NewString(current), nil
			}

			// Set new extensions
			if args[0].IsNull() {
				return values.NewNull(), fmt.Errorf("spl_autoload_extensions() expects parameter 1 to be string")
			}

			newExtensions := args[0].ToString()
			autoloadMutex.Lock()
			old := autoloadExtensions
			autoloadExtensions = newExtensions
			autoloadMutex.Unlock()

			return values.NewString(old), nil
		},
		Parameters: []*registry.Parameter{
			{Name: "file_extensions", Type: "string", DefaultValue: values.NewNull()},
		},
	}
}

// spl_autoload_functions() - Return all registered autoload functions
func getSplAutoloadFunctionsFunction() *registry.Function {
	return &registry.Function{
		Name:      "spl_autoload_functions",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			autoloadMutex.RLock()
			defer autoloadMutex.RUnlock()

			if len(autoloadFunctions) == 0 {
				return values.NewBool(false), nil
			}

			// Create array of registered functions
			result := values.NewArray()
			for i, autoloadFunc := range autoloadFunctions {
				result.ArraySet(values.NewInt(int64(i)), autoloadFunc)
			}

			return result, nil
		},
		Parameters: []*registry.Parameter{},
	}
}

// spl_autoload_register() - Register an autoload function
func getSplAutoloadRegisterFunction() *registry.Function {
	return &registry.Function{
		Name:      "spl_autoload_register",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			var autoloadFunc *values.Value
			var throw bool = true
			var prepend bool = false

			// Parse arguments
			if len(args) > 0 && !args[0].IsNull() {
				autoloadFunc = args[0]
			} else {
				// If no function provided, register default spl_autoload
				autoloadFunc = values.NewString("spl_autoload")
			}

			if len(args) > 1 {
				throw = args[1].ToBool()
			}

			if len(args) > 2 {
				prepend = args[2].ToBool()
			}

			// Validate that the function is callable
			// For simplicity, we'll accept strings (function names) and closures
			if !autoloadFunc.IsCallable() && autoloadFunc.Type != values.TypeString {
				if throw {
					return values.NewNull(), fmt.Errorf("spl_autoload_register(): Argument #1 ($callback) must be a valid callback or null")
				}
				return values.NewBool(false), nil
			}

			autoloadMutex.Lock()
			defer autoloadMutex.Unlock()

			// Check if already registered
			for _, existing := range autoloadFunctions {
				if equalCallables(existing, autoloadFunc) {
					return values.NewBool(true), nil // Already registered
				}
			}

			// Add to list
			if prepend {
				// Add to beginning
				autoloadFunctions = append([]*values.Value{autoloadFunc}, autoloadFunctions...)
			} else {
				// Add to end
				autoloadFunctions = append(autoloadFunctions, autoloadFunc)
			}

			return values.NewBool(true), nil
		},
		Parameters: []*registry.Parameter{
			{Name: "callback", Type: "callable", DefaultValue: values.NewNull()},
			{Name: "throw", Type: "bool", DefaultValue: values.NewBool(true)},
			{Name: "prepend", Type: "bool", DefaultValue: values.NewBool(false)},
		},
	}
}

// spl_autoload_unregister() - Unregister an autoload function
func getSplAutoloadUnregisterFunction() *registry.Function {
	return &registry.Function{
		Name:      "spl_autoload_unregister",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 {
				return values.NewNull(), fmt.Errorf("spl_autoload_unregister() expects exactly 1 parameter, %d given", len(args))
			}

			autoloadFunc := args[0]

			autoloadMutex.Lock()
			defer autoloadMutex.Unlock()

			// Find and remove the function
			for i, existing := range autoloadFunctions {
				if equalCallables(existing, autoloadFunc) {
					// Remove from slice
					autoloadFunctions = append(autoloadFunctions[:i], autoloadFunctions[i+1:]...)
					return values.NewBool(true), nil
				}
			}

			// Function not found
			return values.NewBool(false), nil
		},
		Parameters: []*registry.Parameter{
			{Name: "callback", Type: "callable"},
		},
	}
}

// Helper function to compare callables for equality
func equalCallables(a, b *values.Value) bool {
	if a.Type != b.Type {
		return false
	}

	switch a.Type {
	case values.TypeString:
		return a.ToString() == b.ToString()
	case values.TypeObject:
		// For closures/objects, compare by memory address (simplified)
		return a.Data == b.Data
	default:
		return false
	}
}