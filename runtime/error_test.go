package runtime

import (
	"strings"
	"testing"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// TestErrorFunctions tests all error handling functions using TDD approach
func TestErrorFunctions(t *testing.T) {
	functions := GetErrorFunctions()
	functionMap := make(map[string]*registry.Function)
	for _, fn := range functions {
		functionMap[fn.Name] = fn
	}

	t.Run("error_reporting", func(t *testing.T) {
		fn := functionMap["error_reporting"]
		if fn == nil {
			t.Fatal("error_reporting function not found")
		}

		tests := []struct {
			name     string
			args     []*values.Value
			expected int64
		}{
			{
				name:     "get current level",
				args:     []*values.Value{},
				expected: 30719, // E_ALL
			},
			{
				name:     "set new level",
				args:     []*values.Value{values.NewInt(3)}, // E_ERROR | E_WARNING
				expected: 30719, // Returns previous value
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := fn.Builtin(nil, tt.args)
				if err != nil {
					t.Fatalf("error_reporting() error = %v", err)
				}
				if result.ToInt() != tt.expected {
					t.Errorf("error_reporting() = %v, want %v", result.ToInt(), tt.expected)
				}
			})
		}
	})

	t.Run("trigger_error", func(t *testing.T) {
		fn := functionMap["trigger_error"]
		if fn == nil {
			t.Fatal("trigger_error function not found")
		}

		tests := []struct {
			name     string
			args     []*values.Value
			expected bool
		}{
			{
				name: "trigger user notice",
				args: []*values.Value{
					values.NewString("Test notice"),
					values.NewInt(1024), // E_USER_NOTICE
				},
				expected: true,
			},
			{
				name: "trigger user warning",
				args: []*values.Value{
					values.NewString("Test warning"),
					values.NewInt(512), // E_USER_WARNING
				},
				expected: true,
			},
			{
				name: "default error level",
				args: []*values.Value{
					values.NewString("Default test"),
				},
				expected: true,
			},
			{
				name:     "no arguments",
				args:     []*values.Value{},
				expected: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := fn.Builtin(nil, tt.args)
				if err != nil {
					t.Fatalf("trigger_error() error = %v", err)
				}
				if result.ToBool() != tt.expected {
					t.Errorf("trigger_error() = %v, want %v", result.ToBool(), tt.expected)
				}
			})
		}
	})

	t.Run("user_error", func(t *testing.T) {
		fn := functionMap["user_error"]
		if fn == nil {
			t.Fatal("user_error function not found")
		}

		// user_error is alias for trigger_error, just test basic functionality
		result, err := fn.Builtin(nil, []*values.Value{values.NewString("Alias test")})
		if err != nil {
			t.Fatalf("user_error() error = %v", err)
		}
		if !result.ToBool() {
			t.Errorf("user_error() should return true")
		}
	})

	t.Run("error_get_last_and_clear", func(t *testing.T) {
		clearFn := functionMap["error_clear_last"]
		getFn := functionMap["error_get_last"]
		triggerFn := functionMap["trigger_error"]

		if clearFn == nil || getFn == nil || triggerFn == nil {
			t.Fatal("Required error functions not found")
		}

		// Clear any existing errors
		_, err := clearFn.Builtin(nil, []*values.Value{})
		if err != nil {
			t.Fatalf("error_clear_last() error = %v", err)
		}

		// Get last error after clear (should be null)
		result, err := getFn.Builtin(nil, []*values.Value{})
		if err != nil {
			t.Fatalf("error_get_last() error = %v", err)
		}
		if !result.IsNull() {
			t.Errorf("error_get_last() after clear should return null, got %v", result)
		}

		// Trigger an error
		_, err = triggerFn.Builtin(nil, []*values.Value{
			values.NewString("Test error for get_last"),
			values.NewInt(512), // E_USER_WARNING
		})
		if err != nil {
			t.Fatalf("trigger_error() error = %v", err)
		}

		// Get last error (should be array)
		result, err = getFn.Builtin(nil, []*values.Value{})
		if err != nil {
			t.Fatalf("error_get_last() error = %v", err)
		}
		if !result.IsArray() {
			t.Fatalf("error_get_last() should return array, got %T", result)
		}

		// Check error array structure
		errorArray := result
		message := errorArray.ArrayGet(values.NewString("message"))
		if message == nil || message.ToString() != "Test error for get_last" {
			t.Errorf("error message not stored correctly")
		}

		errorType := errorArray.ArrayGet(values.NewString("type"))
		if errorType == nil || errorType.ToInt() != 512 {
			t.Errorf("error type not stored correctly")
		}

		file := errorArray.ArrayGet(values.NewString("file"))
		if file == nil || !strings.Contains(file.ToString(), "error_test.go") {
			t.Errorf("error file not stored correctly")
		}

		line := errorArray.ArrayGet(values.NewString("line"))
		if line == nil || line.ToInt() == 0 {
			t.Errorf("error line not stored correctly")
		}

		// Clear and verify again
		_, err = clearFn.Builtin(nil, []*values.Value{})
		if err != nil {
			t.Fatalf("error_clear_last() error = %v", err)
		}

		result, err = getFn.Builtin(nil, []*values.Value{})
		if err != nil {
			t.Fatalf("error_get_last() after second clear error = %v", err)
		}
		if !result.IsNull() {
			t.Errorf("error_get_last() after second clear should return null")
		}
	})

	t.Run("set_and_restore_error_handler", func(t *testing.T) {
		setFn := functionMap["set_error_handler"]
		restoreFn := functionMap["restore_error_handler"]

		if setFn == nil || restoreFn == nil {
			t.Fatal("Error handler functions not found")
		}

		// Test setting error handler
		handler := values.NewString("myErrorHandler")
		result, err := setFn.Builtin(nil, []*values.Value{handler})
		if err != nil {
			t.Fatalf("set_error_handler() error = %v", err)
		}
		// Previous handler should be null initially
		if !result.IsNull() {
			t.Errorf("set_error_handler() initial previous should be null, got %v", result)
		}

		// Test restoring error handler
		result, err = restoreFn.Builtin(nil, []*values.Value{})
		if err != nil {
			t.Fatalf("restore_error_handler() error = %v", err)
		}
		if !result.ToBool() {
			t.Errorf("restore_error_handler() should return true")
		}
	})

	t.Run("set_and_restore_exception_handler", func(t *testing.T) {
		setFn := functionMap["set_exception_handler"]
		restoreFn := functionMap["restore_exception_handler"]

		if setFn == nil || restoreFn == nil {
			t.Fatal("Exception handler functions not found")
		}

		// Test setting exception handler
		handler := values.NewString("myExceptionHandler")
		result, err := setFn.Builtin(nil, []*values.Value{handler})
		if err != nil {
			t.Fatalf("set_exception_handler() error = %v", err)
		}
		// Previous handler should be null initially
		if !result.IsNull() {
			t.Errorf("set_exception_handler() initial previous should be null, got %v", result)
		}

		// Test restoring exception handler
		result, err = restoreFn.Builtin(nil, []*values.Value{})
		if err != nil {
			t.Fatalf("restore_exception_handler() error = %v", err)
		}
		if !result.ToBool() {
			t.Errorf("restore_exception_handler() should return true")
		}
	})

	t.Run("debug_backtrace", func(t *testing.T) {
		fn := functionMap["debug_backtrace"]
		if fn == nil {
			t.Fatal("debug_backtrace function not found")
		}

		tests := []struct {
			name string
			args []*values.Value
		}{
			{
				name: "default parameters",
				args: []*values.Value{},
			},
			{
				name: "with options",
				args: []*values.Value{values.NewInt(1)},
			},
			{
				name: "with options and limit",
				args: []*values.Value{values.NewInt(1), values.NewInt(5)},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := fn.Builtin(nil, tt.args)
				if err != nil {
					t.Fatalf("debug_backtrace() error = %v", err)
				}
				if !result.IsArray() {
					t.Errorf("debug_backtrace() should return array, got %T", result)
				}
				// Should have at least one frame
				if result.ArrayCount() == 0 {
					t.Errorf("debug_backtrace() should return non-empty array")
				}
			})
		}
	})

	t.Run("debug_print_backtrace", func(t *testing.T) {
		fn := functionMap["debug_print_backtrace"]
		if fn == nil {
			t.Fatal("debug_print_backtrace function not found")
		}

		// Just test that it doesn't error
		result, err := fn.Builtin(nil, []*values.Value{})
		if err != nil {
			t.Fatalf("debug_print_backtrace() error = %v", err)
		}
		if !result.IsNull() {
			t.Errorf("debug_print_backtrace() should return null")
		}
	})

	t.Run("error_log", func(t *testing.T) {
		fn := functionMap["error_log"]
		if fn == nil {
			t.Fatal("error_log function not found")
		}

		tests := []struct {
			name     string
			args     []*values.Value
			expected bool
		}{
			{
				name: "basic error log",
				args: []*values.Value{
					values.NewString("Test log message"),
				},
				expected: true,
			},
			{
				name: "error log with type",
				args: []*values.Value{
					values.NewString("Test log message with type"),
					values.NewInt(0), // System log
				},
				expected: true,
			},
			{
				name: "error log email type",
				args: []*values.Value{
					values.NewString("Test email log"),
					values.NewInt(1), // Email
				},
				expected: true,
			},
			{
				name:     "no arguments",
				args:     []*values.Value{},
				expected: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := fn.Builtin(nil, tt.args)
				if err != nil {
					t.Fatalf("error_log() error = %v", err)
				}
				if result.ToBool() != tt.expected {
					t.Errorf("error_log() = %v, want %v", result.ToBool(), tt.expected)
				}
			})
		}
	})
}

// TestErrorConstants tests that error constants are properly defined
func TestErrorConstants(t *testing.T) {
	constants := GetAllBuiltinConstants()
	constantMap := make(map[string]*registry.ConstantDescriptor)
	for _, c := range constants {
		constantMap[c.Name] = c
	}

	expectedConstants := map[string]int64{
		"E_ERROR":             1,
		"E_WARNING":           2,
		"E_PARSE":             4,
		"E_NOTICE":            8,
		"E_CORE_ERROR":        16,
		"E_CORE_WARNING":      32,
		"E_COMPILE_ERROR":     64,
		"E_COMPILE_WARNING":   128,
		"E_USER_ERROR":        256,
		"E_USER_WARNING":      512,
		"E_USER_NOTICE":       1024,
		"E_STRICT":            2048,
		"E_RECOVERABLE_ERROR": 4096,
		"E_DEPRECATED":        8192,
		"E_USER_DEPRECATED":   16384,
		"E_ALL":               30719,
	}

	for name, expectedValue := range expectedConstants {
		t.Run(name, func(t *testing.T) {
			constant := constantMap[name]
			if constant == nil {
				t.Fatalf("Constant %s not found", name)
			}
			if constant.Value.ToInt() != expectedValue {
				t.Errorf("Constant %s = %v, want %v", name, constant.Value.ToInt(), expectedValue)
			}
		})
	}
}