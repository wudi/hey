package runtime

import (
	"testing"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// TestSystemFunctions tests system-related functions
func TestSystemFunctions(t *testing.T) {
	functions := GetSystemFunctions()
	functionMap := make(map[string]*registry.Function)
	for _, fn := range functions {
		functionMap[fn.Name] = fn
	}

	t.Run("getenv", func(t *testing.T) {
		fn := functionMap["getenv"]
		if fn == nil {
			t.Fatal("getenv function not found")
		}

		// Test getting a specific environment variable
		result, err := fn.Builtin(nil, []*values.Value{values.NewString("PATH")})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}

		if result == nil {
			t.Error("result is nil")
			return
		}

		// PATH should exist and be a string
		if result.IsBool() && !result.ToBool() {
			t.Error("PATH environment variable should exist")
		}
	})

	t.Run("getcwd", func(t *testing.T) {
		fn := functionMap["getcwd"]
		if fn == nil {
			t.Fatal("getcwd function not found")
		}

		result, err := fn.Builtin(nil, []*values.Value{})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}

		if result == nil {
			t.Error("result is nil")
			return
		}

		// Should return a string path
		if !result.IsString() {
			t.Errorf("expected string result, got %T", result.Data)
		}

		cwd := result.ToString()
		if len(cwd) == 0 {
			t.Error("current working directory should not be empty")
		}
	})
}