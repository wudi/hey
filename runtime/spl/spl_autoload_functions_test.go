package spl

import (
	"testing"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

func TestSplAutoloadFunctions(t *testing.T) {
	registry.Initialize()

	// Reset autoload state before each test
	autoloadMutex.Lock()
	autoloadFunctions = make([]*values.Value, 0)
	autoloadExtensions = ".inc,.php"
	autoloadMutex.Unlock()

	ctx := &mockContext{registry: registry.GlobalRegistry}

	t.Run("SplAutoloadExtensions", func(t *testing.T) {
		function := getSplAutoloadExtensionsFunction()

		// Test getting default extensions
		result, err := function.Builtin(ctx, []*values.Value{})
		if err != nil {
			t.Fatalf("spl_autoload_extensions() failed: %v", err)
		}

		if result.Type != values.TypeString {
			t.Errorf("Expected string result, got %s", result.Type)
		}

		if result.ToString() != ".inc,.php" {
			t.Errorf("Expected '.inc,.php', got '%s'", result.ToString())
		}

		// Test setting new extensions
		newExtensions := values.NewString(".php,.class.php,.inc")
		result, err = function.Builtin(ctx, []*values.Value{newExtensions})
		if err != nil {
			t.Fatalf("Setting extensions failed: %v", err)
		}

		// Should return old extensions
		if result.ToString() != ".inc,.php" {
			t.Errorf("Expected old extensions '.inc,.php', got '%s'", result.ToString())
		}

		// Verify new extensions are set
		result, err = function.Builtin(ctx, []*values.Value{})
		if err != nil {
			t.Fatalf("Getting new extensions failed: %v", err)
		}

		if result.ToString() != ".php,.class.php,.inc" {
			t.Errorf("Expected '.php,.class.php,.inc', got '%s'", result.ToString())
		}
	})

	t.Run("SplAutoloadFunctions", func(t *testing.T) {
		function := getSplAutoloadFunctionsFunction()

		// Test with no registered functions
		result, err := function.Builtin(ctx, []*values.Value{})
		if err != nil {
			t.Fatalf("spl_autoload_functions() failed: %v", err)
		}

		if result.Type != values.TypeBool || result.ToBool() != false {
			t.Errorf("Expected false when no functions registered, got %v", result)
		}

		// Register a function and test again
		autoloadMutex.Lock()
		autoloadFunctions = append(autoloadFunctions, values.NewString("test_function"))
		autoloadMutex.Unlock()

		result, err = function.Builtin(ctx, []*values.Value{})
		if err != nil {
			t.Fatalf("spl_autoload_functions() with registered function failed: %v", err)
		}

		if !result.IsArray() {
			t.Errorf("Expected array result, got %s", result.Type)
		}

		if result.ArrayCount() != 1 {
			t.Errorf("Expected 1 function, got %d", result.ArrayCount())
		}

		firstFunc := result.ArrayGet(values.NewInt(0))
		if firstFunc == nil || firstFunc.ToString() != "test_function" {
			t.Errorf("Expected 'test_function', got %v", firstFunc)
		}
	})

	t.Run("SplAutoloadRegister", func(t *testing.T) {
		function := getSplAutoloadRegisterFunction()

		// Test registering a function name
		callback := values.NewString("myAutoloader")
		result, err := function.Builtin(ctx, []*values.Value{callback})
		if err != nil {
			t.Fatalf("spl_autoload_register() failed: %v", err)
		}

		if result.Type != values.TypeBool || !result.ToBool() {
			t.Errorf("Expected true result, got %v", result)
		}

		// Verify function was registered
		functionsResult, err := getSplAutoloadFunctionsFunction().Builtin(ctx, []*values.Value{})
		if err != nil {
			t.Fatalf("Getting functions after registration failed: %v", err)
		}

		if !functionsResult.IsArray() || functionsResult.ArrayCount() == 0 {
			t.Error("Function should be registered")
		}

		// Test registering the same function again (should still return true)
		result, err = function.Builtin(ctx, []*values.Value{callback})
		if err != nil {
			t.Fatalf("Re-registering same function failed: %v", err)
		}

		if !result.ToBool() {
			t.Error("Re-registering same function should return true")
		}

		// Test prepend option
		prependCallback := values.NewString("prependAutoloader")
		result, err = function.Builtin(ctx, []*values.Value{
			prependCallback,
			values.NewBool(true),  // throw
			values.NewBool(true),  // prepend
		})
		if err != nil {
			t.Fatalf("Prepend registration failed: %v", err)
		}

		if !result.ToBool() {
			t.Error("Prepend registration should return true")
		}

		// Verify prepend worked (should be first in array)
		functionsResult, err = getSplAutoloadFunctionsFunction().Builtin(ctx, []*values.Value{})
		if err != nil {
			t.Fatalf("Getting functions after prepend failed: %v", err)
		}

		firstFunc := functionsResult.ArrayGet(values.NewInt(0))
		if firstFunc == nil || firstFunc.ToString() != "prependAutoloader" {
			t.Errorf("Expected 'prependAutoloader' as first function, got %v", firstFunc)
		}
	})

	t.Run("SplAutoloadUnregister", func(t *testing.T) {
		function := getSplAutoloadUnregisterFunction()

		// First register a function
		registerFunc := getSplAutoloadRegisterFunction()
		callback := values.NewString("testFunction")
		_, err := registerFunc.Builtin(ctx, []*values.Value{callback})
		if err != nil {
			t.Fatalf("Registration for unregister test failed: %v", err)
		}

		// Test unregistering the function
		result, err := function.Builtin(ctx, []*values.Value{callback})
		if err != nil {
			t.Fatalf("spl_autoload_unregister() failed: %v", err)
		}

		if result.Type != values.TypeBool || !result.ToBool() {
			t.Errorf("Expected true result, got %v", result)
		}

		// Verify function was removed
		functionsResult, err := getSplAutoloadFunctionsFunction().Builtin(ctx, []*values.Value{})
		if err != nil {
			t.Fatalf("Getting functions after unregister failed: %v", err)
		}

		// Should be false (no functions) or empty array
		if functionsResult.IsArray() && functionsResult.ArrayCount() > 0 {
			// Check if our function is still there
			for i := 0; i < functionsResult.ArrayCount(); i++ {
				funcVal := functionsResult.ArrayGet(values.NewInt(int64(i)))
				if funcVal != nil && funcVal.ToString() == "testFunction" {
					t.Error("Function should have been removed")
				}
			}
		}

		// Test unregistering non-existent function
		nonExistent := values.NewString("nonExistentFunction")
		result, err = function.Builtin(ctx, []*values.Value{nonExistent})
		if err != nil {
			t.Fatalf("Unregistering non-existent function failed: %v", err)
		}

		if result.ToBool() {
			t.Error("Unregistering non-existent function should return false")
		}
	})

	t.Run("SplAutoload", func(t *testing.T) {
		function := getSplAutoloadFunction()

		// Test basic autoload call
		className := values.NewString("TestClass")
		result, err := function.Builtin(ctx, []*values.Value{className})
		if err != nil {
			t.Fatalf("spl_autoload() failed: %v", err)
		}

		if result.Type != values.TypeNull {
			t.Errorf("Expected null result, got %s", result.Type)
		}

		// Test with custom extensions
		customExtensions := values.NewString(".class.php,.inc")
		result, err = function.Builtin(ctx, []*values.Value{className, customExtensions})
		if err != nil {
			t.Fatalf("spl_autoload() with custom extensions failed: %v", err)
		}

		if result.Type != values.TypeNull {
			t.Errorf("Expected null result, got %s", result.Type)
		}
	})

	t.Run("SplAutoloadCall", func(t *testing.T) {
		function := getSplAutoloadCallFunction()

		// Test calling with no registered functions
		className := values.NewString("TestClass")
		result, err := function.Builtin(ctx, []*values.Value{className})
		if err != nil {
			t.Fatalf("spl_autoload_call() failed: %v", err)
		}

		if result.Type != values.TypeNull {
			t.Errorf("Expected null result, got %s", result.Type)
		}

		// Test with registered functions
		registerFunc := getSplAutoloadRegisterFunction()
		callback := values.NewString("testAutoloader")
		_, err = registerFunc.Builtin(ctx, []*values.Value{callback})
		if err != nil {
			t.Fatalf("Registration for autoload_call test failed: %v", err)
		}

		result, err = function.Builtin(ctx, []*values.Value{className})
		if err != nil {
			t.Fatalf("spl_autoload_call() with registered functions failed: %v", err)
		}

		if result.Type != values.TypeNull {
			t.Errorf("Expected null result, got %s", result.Type)
		}
	})

	t.Run("ErrorHandling", func(t *testing.T) {
		// Test spl_autoload_register with too few arguments (should work - callback is optional)
		registerFunc := getSplAutoloadRegisterFunction()
		result, err := registerFunc.Builtin(ctx, []*values.Value{})
		if err != nil {
			t.Fatalf("spl_autoload_register() with no args failed: %v", err)
		}

		if !result.ToBool() {
			t.Error("spl_autoload_register() with no args should succeed (registers default)")
		}

		// Test spl_autoload_unregister with too few arguments
		unregisterFunc := getSplAutoloadUnregisterFunction()
		_, err = unregisterFunc.Builtin(ctx, []*values.Value{})
		if err == nil {
			t.Error("spl_autoload_unregister() should require exactly 1 parameter")
		}

		// Test spl_autoload with too few arguments
		autoloadFunc := getSplAutoloadFunction()
		_, err = autoloadFunc.Builtin(ctx, []*values.Value{})
		if err == nil {
			t.Error("spl_autoload() should require at least 1 parameter")
		}
	})
}