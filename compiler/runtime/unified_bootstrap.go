package runtime

import (
	"fmt"

	"github.com/wudi/php-parser/compiler/registry"
	"github.com/wudi/php-parser/compiler/values"
)

// UnifiedBootstrap initializes the runtime with the unified class system
func UnifiedBootstrap() error {
	// Initialize unified registry
	registry.Initialize()

	// Register built-in constants (same as before)
	if err := registerBuiltinConstants(); err != nil {
		return fmt.Errorf("failed to register built-in constants: %v", err)
	}

	// Register built-in variables (same as before)
	if err := registerBuiltinVariables(); err != nil {
		return fmt.Errorf("failed to register built-in variables: %v", err)
	}

	// Register built-in functions (same as before)
	if err := registerBuiltinFunctions(); err != nil {
		return fmt.Errorf("failed to register built-in functions: %v", err)
	}

	// Register unified built-in classes
	if err := registerUnifiedBuiltinClasses(); err != nil {
		return fmt.Errorf("failed to register unified built-in classes: %v", err)
	}

	return nil
}

// registerUnifiedBuiltinClasses registers all built-in classes using the unified system
func registerUnifiedBuiltinClasses() error {
	// Exception class
	_, err := registry.NewClass("Exception").
		BuiltinClass("core").
		AddProperty("message", "protected", "string", values.NewString("")).
		AddProperty("code", "protected", "int", values.NewInt(0)).
		AddProperty("file", "protected", "string", values.NewString("")).
		AddProperty("line", "protected", "int", values.NewInt(0)).
		Constructor().
		WithOptionalParameter("message", "string", values.NewString("")).
		WithOptionalParameter("code", "int", values.NewInt(0)).
		WithNativeHandler(exceptionConstructHandler).
		Done().
		AddMethod("getMessage", "public").
		WithNativeHandler(exceptionGetMessageHandler).
		Done().
		AddMethod("getCode", "public").
		WithNativeHandler(exceptionGetCodeHandler).
		Done().
		AddMethod("getFile", "public").
		WithNativeHandler(exceptionGetFileHandler).
		Done().
		AddMethod("getLine", "public").
		WithNativeHandler(exceptionGetLineHandler).
		Done().
		AddMethod("getTrace", "public").
		WithNativeHandler(exceptionGetTraceHandler).
		Done().
		AddMethod("getTraceAsString", "public").
		WithNativeHandler(exceptionGetTraceAsStringHandler).
		Done().
		AddMethod("getPrevious", "public").
		WithNativeHandler(exceptionGetPreviousHandler).
		Done().
		AddMethod("__toString", "public").
		WithNativeHandler(exceptionToStringHandler).
		Done().
		BuildAndRegister()

	if err != nil {
		return fmt.Errorf("failed to register Exception class: %v", err)
	}

	// stdClass
	_, err = registry.NewClass("stdClass").
		BuiltinClass("core").
		BuildAndRegister()

	if err != nil {
		return fmt.Errorf("failed to register stdClass: %v", err)
	}

	// WaitGroup class for concurrency
	_, err = registry.NewClass("WaitGroup").
		BuiltinClass("core").
		Constructor().
		WithNativeHandler(unifiedWaitGroupConstructHandler).
		Done().
		AddMethod("Add", "public").
		WithParameter("delta", "int").
		WithNativeHandler(unifiedWaitGroupAddHandler).
		Done().
		AddMethod("Done", "public").
		WithNativeHandler(unifiedWaitGroupDoneHandler).
		Done().
		AddMethod("Wait", "public").
		WithNativeHandler(unifiedWaitGroupWaitHandler).
		Done().
		BuildAndRegister()

	if err != nil {
		return fmt.Errorf("failed to register WaitGroup class: %v", err)
	}

	return nil
}

// Updated method handlers to work with the unified system

// Exception method handlers - these use the unified ExecutionContext interface
func exceptionConstructHandler(ctx registry.ExecutionContext, args []*values.Value) (*values.Value, error) {
	// In a full implementation, this would set the object properties
	// For now, we just return null
	return values.NewNull(), nil
}

func exceptionGetMessageHandler(ctx registry.ExecutionContext, args []*values.Value) (*values.Value, error) {
	// Simplified implementation - in a full system would access object properties
	return values.NewString("Exception"), nil
}

func exceptionGetCodeHandler(ctx registry.ExecutionContext, args []*values.Value) (*values.Value, error) {
	return values.NewInt(0), nil
}

func exceptionGetFileHandler(ctx registry.ExecutionContext, args []*values.Value) (*values.Value, error) {
	return values.NewString(""), nil
}

func exceptionGetLineHandler(ctx registry.ExecutionContext, args []*values.Value) (*values.Value, error) {
	return values.NewInt(0), nil
}

func exceptionGetTraceHandler(ctx registry.ExecutionContext, args []*values.Value) (*values.Value, error) {
	return values.NewArray(), nil
}

func exceptionGetTraceAsStringHandler(ctx registry.ExecutionContext, args []*values.Value) (*values.Value, error) {
	return values.NewString(""), nil
}

func exceptionGetPreviousHandler(ctx registry.ExecutionContext, args []*values.Value) (*values.Value, error) {
	return values.NewNull(), nil
}

func exceptionToStringHandler(ctx registry.ExecutionContext, args []*values.Value) (*values.Value, error) {
	return values.NewString("Exception"), nil
}

// Unified WaitGroup method handlers
func unifiedWaitGroupConstructHandler(ctx registry.ExecutionContext, args []*values.Value) (*values.Value, error) {
	return values.NewNull(), nil
}

func unifiedWaitGroupAddHandler(ctx registry.ExecutionContext, args []*values.Value) (*values.Value, error) {
	return values.NewNull(), nil
}

func unifiedWaitGroupDoneHandler(ctx registry.ExecutionContext, args []*values.Value) (*values.Value, error) {
	return values.NewNull(), nil
}

func unifiedWaitGroupWaitHandler(ctx registry.ExecutionContext, args []*values.Value) (*values.Value, error) {
	return values.NewNull(), nil
}

// UnifiedBootstrapLegacy provides a bootstrap function that also calls unified system
func UnifiedBootstrapLegacy() error {
	// Initialize the old system first for backward compatibility
	Initialize()

	if err := registerBuiltinConstants(); err != nil {
		return fmt.Errorf("failed to register built-in constants: %v", err)
	}

	if err := registerBuiltinVariables(); err != nil {
		return fmt.Errorf("failed to register built-in variables: %v", err)
	}

	if err := registerBuiltinFunctions(); err != nil {
		return fmt.Errorf("failed to register built-in functions: %v", err)
	}

	// Register old-style classes
	if err := registerBuiltinClasses(); err != nil {
		return fmt.Errorf("failed to register built-in classes: %v", err)
	}

	// Also initialize the unified system
	return UnifiedBootstrap()
}
