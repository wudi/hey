package runtime

import (
	"fmt"
	"testing"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// mockBuiltinContext is a simple mock for BuiltinCallContext
type mockBuiltinContext struct {
	registry *registry.Registry
}

func (m *mockBuiltinContext) WriteOutput(val *values.Value) error {
	return nil
}

func (m *mockBuiltinContext) GetGlobal(name string) (*values.Value, bool) {
	return nil, false
}

func (m *mockBuiltinContext) SetGlobal(name string, val *values.Value) {}

func (m *mockBuiltinContext) SymbolRegistry() *registry.Registry {
	if m.registry == nil {
		// Initialize the global registry to ensure string functions are available
		registry.Initialize()
		m.registry = registry.GlobalRegistry

		// Register functions for testing
		for _, fn := range GetStringFunctions() {
			m.registry.RegisterFunction(fn)
		}
		for _, fn := range GetTypeFunctions() {
			m.registry.RegisterFunction(fn)
		}
	}
	return m.registry
}

func (m *mockBuiltinContext) LookupUserFunction(name string) (*registry.Function, bool) {
	return nil, false
}

func (m *mockBuiltinContext) CallUserFunction(function *registry.Function, args []*values.Value) (*values.Value, error) {
	return nil, fmt.Errorf("user function calls not supported in test mock")
}

func (m *mockBuiltinContext) ThrowException(exception *values.Value) error {
	return fmt.Errorf("exception thrown in test mock: %v", exception)
}

func (m *mockBuiltinContext) SimpleCallUserFunction(function *registry.Function, args []*values.Value) (*values.Value, error) {
	return nil, fmt.Errorf("user function calls not supported in test mock")
}

func (m *mockBuiltinContext) LookupUserClass(name string) (*registry.Class, bool) {
	return nil, false
}

func (m *mockBuiltinContext) Halt(exitCode int, message string) error {
	return nil
}

func (m *mockBuiltinContext) GetExecutionContext() registry.ExecutionContextInterface {
	return nil
}

func (m *mockBuiltinContext) GetOutputBufferStack() registry.OutputBufferStackInterface {
	return nil
}

func (m *mockBuiltinContext) GetCurrentFunctionArgCount() (int, error) {
	return 0, fmt.Errorf("cannot be called from the global scope")
}

func (m *mockBuiltinContext) GetCurrentFunctionArg(index int) (*values.Value, error) {
	return nil, fmt.Errorf("cannot be called from the global scope")
}

func (m *mockBuiltinContext) GetCurrentFunctionArgs() ([]*values.Value, error) {
	return nil, fmt.Errorf("cannot be called from the global scope")
}

// TestArrayFunctions tests all array functions using TDD approach
func TestArrayFunctions(t *testing.T) {
	functions := GetArrayFunctions()
	functionMap := make(map[string]*registry.Function)
	for _, fn := range functions {
		functionMap[fn.Name] = fn
	}

	ctx := &mockBuiltinContext{}

	t.Run("array_map", func(t *testing.T) {
		fn := functionMap["array_map"]
		if fn == nil {
			t.Fatal("array_map function not found")
		}

		t.Run("null callback creates array of arrays", func(t *testing.T) {
			// Create test arrays: [1,2,3] and [4,5,6]
			arr1 := values.NewArray()
			arr1Ptr := arr1.Data.(*values.Array)
			arr1Ptr.Elements[int64(0)] = values.NewInt(1)
			arr1Ptr.Elements[int64(1)] = values.NewInt(2)
			arr1Ptr.Elements[int64(2)] = values.NewInt(3)
			arr1Ptr.NextIndex = 3

			arr2 := values.NewArray()
			arr2Ptr := arr2.Data.(*values.Array)
			arr2Ptr.Elements[int64(0)] = values.NewInt(4)
			arr2Ptr.Elements[int64(1)] = values.NewInt(5)
			arr2Ptr.Elements[int64(2)] = values.NewInt(6)
			arr2Ptr.NextIndex = 3

			result, err := fn.Builtin(ctx, []*values.Value{
				values.NewNull(), // null callback
				arr1,
				arr2,
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !result.IsArray() {
				t.Fatal("result should be an array")
			}

			resultArr := result.Data.(*values.Array)
			if result.ArrayCount() != 3 {
				t.Errorf("expected 3 elements, got %d", result.ArrayCount())
			}

			// Check first element: [1,4]
			firstRow := resultArr.Elements[int64(0)]
			if !firstRow.IsArray() {
				t.Fatal("first row should be an array")
			}
			firstRowArr := firstRow.Data.(*values.Array)
			if firstRow.ArrayCount() != 2 {
				t.Errorf("first row should have 2 elements, got %d", firstRow.ArrayCount())
			}
			if firstRowArr.Elements[int64(0)].ToInt() != 1 {
				t.Errorf("expected 1, got %d", firstRowArr.Elements[int64(0)].ToInt())
			}
			if firstRowArr.Elements[int64(1)].ToInt() != 4 {
				t.Errorf("expected 4, got %d", firstRowArr.Elements[int64(1)].ToInt())
			}
		})

		t.Run("string function name callback", func(t *testing.T) {
			// Create test array: ["hello", "world"]
			arr := values.NewArray()
			arrPtr := arr.Data.(*values.Array)
			arrPtr.Elements[int64(0)] = values.NewString("hello")
			arrPtr.Elements[int64(1)] = values.NewString("world")
			arrPtr.NextIndex = 2

			result, err := fn.Builtin(ctx, []*values.Value{
				values.NewString("strtoupper"), // builtin function name
				arr,
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !result.IsArray() {
				t.Fatal("result should be an array")
			}

			if result.ArrayCount() != 2 {
				t.Errorf("expected 2 elements, got %d", result.ArrayCount())
			}

			resultArr := result.Data.(*values.Array)
			if resultArr.Elements[int64(0)].ToString() != "HELLO" {
				t.Errorf("expected 'HELLO', got '%s'", resultArr.Elements[int64(0)].ToString())
			}
			if resultArr.Elements[int64(1)].ToString() != "WORLD" {
				t.Errorf("expected 'WORLD', got '%s'", resultArr.Elements[int64(1)].ToString())
			}
		})

		t.Run("associative array preserves keys", func(t *testing.T) {
			// Create test array: ["a" => 1, "b" => 2, "c" => 3]
			arr := values.NewArray()
			arrPtr := arr.Data.(*values.Array)
			arrPtr.Elements["a"] = values.NewInt(1)
			arrPtr.Elements["b"] = values.NewInt(2)
			arrPtr.Elements["c"] = values.NewInt(3)

			result, err := fn.Builtin(ctx, []*values.Value{
				values.NewString("strlen"), // get string length
				arr,
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !result.IsArray() {
				t.Fatal("result should be an array")
			}

			resultArr := result.Data.(*values.Array)
			if resultArr.Elements["a"].ToInt() != 1 {
				t.Errorf("expected 1, got %d", resultArr.Elements["a"].ToInt())
			}
			if resultArr.Elements["b"].ToInt() != 1 {
				t.Errorf("expected 1, got %d", resultArr.Elements["b"].ToInt())
			}
			if resultArr.Elements["c"].ToInt() != 1 {
				t.Errorf("expected 1, got %d", resultArr.Elements["c"].ToInt())
			}
		})

		t.Run("empty array returns empty", func(t *testing.T) {
			arr := values.NewArray()

			result, err := fn.Builtin(ctx, []*values.Value{
				values.NewString("strtoupper"),
				arr,
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !result.IsArray() {
				t.Fatal("result should be an array")
			}

			if result.ArrayCount() != 0 {
				t.Errorf("expected 0 elements, got %d", result.ArrayCount())
			}
		})

		t.Run("invalid function name returns error", func(t *testing.T) {
			arr := values.NewArray()
			arrPtr := arr.Data.(*values.Array)
			arrPtr.Elements[int64(0)] = values.NewString("test")

			_, err := fn.Builtin(ctx, []*values.Value{
				values.NewString("nonexistent_function"),
				arr,
			})

			if err == nil {
				t.Fatal("expected error for nonexistent function")
			}
		})

		t.Run("too few arguments", func(t *testing.T) {
			result, err := fn.Builtin(ctx, []*values.Value{
				values.NewString("strtoupper"),
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !result.IsArray() {
				t.Fatal("result should be an array")
			}

			if result.ArrayCount() != 0 {
				t.Errorf("expected empty array, got %d elements", result.ArrayCount())
			}
		})
	})

	t.Run("array_slice", func(t *testing.T) {
		fn := functionMap["array_slice"]
		if fn == nil {
			t.Fatal("array_slice function not found")
		}

		t.Run("basic slice from offset", func(t *testing.T) {
			// Create test array: [1,2,3,4,5]
			arr := values.NewArray()
			arrPtr := arr.Data.(*values.Array)
			for i := int64(0); i < 5; i++ {
				arrPtr.Elements[i] = values.NewInt(i + 1)
			}
			arrPtr.NextIndex = 5

			result, err := fn.Builtin(ctx, []*values.Value{
				arr,
				values.NewInt(2), // offset 2
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !result.IsArray() {
				t.Fatal("result should be an array")
			}

			if result.ArrayCount() != 3 {
				t.Errorf("expected 3 elements, got %d", result.ArrayCount())
			}

			resultArr := result.Data.(*values.Array)
			if resultArr.Elements[int64(0)].ToInt() != 3 {
				t.Errorf("expected 3, got %d", resultArr.Elements[int64(0)].ToInt())
			}
			if resultArr.Elements[int64(1)].ToInt() != 4 {
				t.Errorf("expected 4, got %d", resultArr.Elements[int64(1)].ToInt())
			}
			if resultArr.Elements[int64(2)].ToInt() != 5 {
				t.Errorf("expected 5, got %d", resultArr.Elements[int64(2)].ToInt())
			}
		})

		t.Run("slice with length", func(t *testing.T) {
			// Create test array: [1,2,3,4,5]
			arr := values.NewArray()
			arrPtr := arr.Data.(*values.Array)
			for i := int64(0); i < 5; i++ {
				arrPtr.Elements[i] = values.NewInt(i + 1)
			}
			arrPtr.NextIndex = 5

			result, err := fn.Builtin(ctx, []*values.Value{
				arr,
				values.NewInt(1), // offset 1
				values.NewInt(3), // length 3
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.ArrayCount() != 3 {
				t.Errorf("expected 3 elements, got %d", result.ArrayCount())
			}

			resultArr := result.Data.(*values.Array)
			if resultArr.Elements[int64(0)].ToInt() != 2 {
				t.Errorf("expected 2, got %d", resultArr.Elements[int64(0)].ToInt())
			}
			if resultArr.Elements[int64(1)].ToInt() != 3 {
				t.Errorf("expected 3, got %d", resultArr.Elements[int64(1)].ToInt())
			}
			if resultArr.Elements[int64(2)].ToInt() != 4 {
				t.Errorf("expected 4, got %d", resultArr.Elements[int64(2)].ToInt())
			}
		})

		t.Run("negative offset", func(t *testing.T) {
			// Create test array: [1,2,3,4,5]
			arr := values.NewArray()
			arrPtr := arr.Data.(*values.Array)
			for i := int64(0); i < 5; i++ {
				arrPtr.Elements[i] = values.NewInt(i + 1)
			}
			arrPtr.NextIndex = 5

			result, err := fn.Builtin(ctx, []*values.Value{
				arr,
				values.NewInt(-2), // offset -2
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.ArrayCount() != 2 {
				t.Errorf("expected 2 elements, got %d", result.ArrayCount())
			}

			resultArr := result.Data.(*values.Array)
			if resultArr.Elements[int64(0)].ToInt() != 4 {
				t.Errorf("expected 4, got %d", resultArr.Elements[int64(0)].ToInt())
			}
			if resultArr.Elements[int64(1)].ToInt() != 5 {
				t.Errorf("expected 5, got %d", resultArr.Elements[int64(1)].ToInt())
			}
		})

		t.Run("associative array with preserve_keys", func(t *testing.T) {
			// Create test array: ["a" => 1, "b" => 2, "c" => 3, "d" => 4]
			arr := values.NewArray()
			arrPtr := arr.Data.(*values.Array)
			arrPtr.Elements["a"] = values.NewInt(1)
			arrPtr.Elements["b"] = values.NewInt(2)
			arrPtr.Elements["c"] = values.NewInt(3)
			arrPtr.Elements["d"] = values.NewInt(4)

			result, err := fn.Builtin(ctx, []*values.Value{
				arr,
				values.NewInt(1),     // offset 1
				values.NewInt(2),     // length 2
				values.NewBool(true), // preserve_keys = true
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.ArrayCount() != 2 {
				t.Errorf("expected 2 elements, got %d", result.ArrayCount())
			}

			resultArr := result.Data.(*values.Array)
			// Note: the order depends on iteration order of string keys
			// We'll just check that we have the right number of elements
			if len(resultArr.Elements) != 2 {
				t.Errorf("expected 2 elements in result map, got %d", len(resultArr.Elements))
			}
		})

		t.Run("zero length returns empty", func(t *testing.T) {
			arr := values.NewArray()
			arrPtr := arr.Data.(*values.Array)
			arrPtr.Elements[int64(0)] = values.NewInt(1)
			arrPtr.Elements[int64(1)] = values.NewInt(2)
			arrPtr.NextIndex = 2

			result, err := fn.Builtin(ctx, []*values.Value{
				arr,
				values.NewInt(1),
				values.NewInt(0), // length 0
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.ArrayCount() != 0 {
				t.Errorf("expected 0 elements, got %d", result.ArrayCount())
			}
		})
	})

	t.Run("array_search", func(t *testing.T) {
		fn := functionMap["array_search"]
		if fn == nil {
			t.Fatal("array_search function not found")
		}

		t.Run("basic search found", func(t *testing.T) {
			// Create test array: [1,2,3,4,5]
			arr := values.NewArray()
			arrPtr := arr.Data.(*values.Array)
			for i := int64(0); i < 5; i++ {
				arrPtr.Elements[i] = values.NewInt(i + 1)
			}
			arrPtr.NextIndex = 5

			result, err := fn.Builtin(ctx, []*values.Value{
				values.NewInt(3), // search for 3
				arr,
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !result.IsInt() {
				t.Fatalf("expected int result, got %s", result.Type)
			}

			if result.ToInt() != 2 {
				t.Errorf("expected index 2, got %d", result.ToInt())
			}
		})

		t.Run("search not found", func(t *testing.T) {
			arr := values.NewArray()
			arrPtr := arr.Data.(*values.Array)
			arrPtr.Elements[int64(0)] = values.NewInt(1)
			arrPtr.Elements[int64(1)] = values.NewInt(2)
			arrPtr.NextIndex = 2

			result, err := fn.Builtin(ctx, []*values.Value{
				values.NewInt(10), // search for 10 (not found)
				arr,
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !result.IsBool() {
				t.Fatalf("expected bool result, got %s", result.Type)
			}

			if result.ToBool() != false {
				t.Errorf("expected false, got %v", result.ToBool())
			}
		})

		t.Run("associative array search", func(t *testing.T) {
			arr := values.NewArray()
			arrPtr := arr.Data.(*values.Array)
			arrPtr.Elements["a"] = values.NewString("apple")
			arrPtr.Elements["b"] = values.NewString("banana")
			arrPtr.Elements["c"] = values.NewString("cherry")

			result, err := fn.Builtin(ctx, []*values.Value{
				values.NewString("banana"), // search for banana
				arr,
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !result.IsString() {
				t.Fatalf("expected string result, got %s", result.Type)
			}

			if result.ToString() != "b" {
				t.Errorf("expected key 'b', got '%s'", result.ToString())
			}
		})

		t.Run("strict vs loose comparison", func(t *testing.T) {
			arr := values.NewArray()
			arrPtr := arr.Data.(*values.Array)
			arrPtr.Elements[int64(0)] = values.NewString("1")
			arrPtr.Elements[int64(1)] = values.NewInt(2)
			arrPtr.NextIndex = 2

			// Loose comparison (default)
			result, err := fn.Builtin(ctx, []*values.Value{
				values.NewInt(1), // search for 1
				arr,
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.ToInt() != 0 {
				t.Errorf("loose comparison: expected index 0, got %d", result.ToInt())
			}

			// Strict comparison
			result, err = fn.Builtin(ctx, []*values.Value{
				values.NewInt(1), // search for 1
				arr,
				values.NewBool(true), // strict = true
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !result.IsBool() || result.ToBool() != false {
				t.Errorf("strict comparison: expected false, got %v", result)
			}
		})
	})

	t.Run("array_pop", func(t *testing.T) {
		fn := functionMap["array_pop"]
		if fn == nil {
			t.Fatal("array_pop function not found")
		}

		t.Run("pop from numeric array", func(t *testing.T) {
			// Create test array: [1,2,3,4,5]
			arr := values.NewArray()
			arrPtr := arr.Data.(*values.Array)
			for i := int64(0); i < 5; i++ {
				arrPtr.Elements[i] = values.NewInt(i + 1)
			}
			arrPtr.NextIndex = 5

			result, err := fn.Builtin(ctx, []*values.Value{arr})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.ToInt() != 5 {
				t.Errorf("expected 5, got %d", result.ToInt())
			}

			if arr.ArrayCount() != 4 {
				t.Errorf("expected array length 4, got %d", arr.ArrayCount())
			}
		})

		t.Run("pop from empty array", func(t *testing.T) {
			arr := values.NewArray()
			result, err := fn.Builtin(ctx, []*values.Value{arr})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !result.IsNull() {
				t.Errorf("expected null, got %v", result)
			}
		})
	})

	t.Run("array_shift", func(t *testing.T) {
		fn := functionMap["array_shift"]
		if fn == nil {
			t.Fatal("array_shift function not found")
		}

		t.Run("shift from numeric array", func(t *testing.T) {
			// Create test array: [1,2,3,4,5]
			arr := values.NewArray()
			arrPtr := arr.Data.(*values.Array)
			for i := int64(0); i < 5; i++ {
				arrPtr.Elements[i] = values.NewInt(i + 1)
			}
			arrPtr.NextIndex = 5

			result, err := fn.Builtin(ctx, []*values.Value{arr})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.ToInt() != 1 {
				t.Errorf("expected 1, got %d", result.ToInt())
			}

			if arr.ArrayCount() != 4 {
				t.Errorf("expected array length 4, got %d", arr.ArrayCount())
			}

			// Check that elements were shifted
			if arrPtr.Elements[int64(0)].ToInt() != 2 {
				t.Errorf("expected first element to be 2, got %d", arrPtr.Elements[int64(0)].ToInt())
			}
		})

		t.Run("shift from empty array", func(t *testing.T) {
			arr := values.NewArray()
			result, err := fn.Builtin(ctx, []*values.Value{arr})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !result.IsNull() {
				t.Errorf("expected null, got %v", result)
			}
		})
	})

	t.Run("array_unshift", func(t *testing.T) {
		fn := functionMap["array_unshift"]
		if fn == nil {
			t.Fatal("array_unshift function not found")
		}

		t.Run("unshift single value", func(t *testing.T) {
			// Create test array: [2,3,4]
			arr := values.NewArray()
			arrPtr := arr.Data.(*values.Array)
			arrPtr.Elements[int64(0)] = values.NewInt(2)
			arrPtr.Elements[int64(1)] = values.NewInt(3)
			arrPtr.Elements[int64(2)] = values.NewInt(4)
			arrPtr.NextIndex = 3

			result, err := fn.Builtin(ctx, []*values.Value{
				arr,
				values.NewInt(1), // unshift 1
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.ToInt() != 4 {
				t.Errorf("expected length 4, got %d", result.ToInt())
			}

			if arr.ArrayCount() != 4 {
				t.Errorf("expected array length 4, got %d", arr.ArrayCount())
			}

			// Check that values were shifted
			if arrPtr.Elements[int64(0)].ToInt() != 1 {
				t.Errorf("expected first element to be 1, got %d", arrPtr.Elements[int64(0)].ToInt())
			}
			if arrPtr.Elements[int64(1)].ToInt() != 2 {
				t.Errorf("expected second element to be 2, got %d", arrPtr.Elements[int64(1)].ToInt())
			}
		})

		t.Run("unshift multiple values", func(t *testing.T) {
			arr := values.NewArray()
			arrPtr := arr.Data.(*values.Array)
			arrPtr.Elements[int64(0)] = values.NewInt(4)
			arrPtr.Elements[int64(1)] = values.NewInt(5)
			arrPtr.NextIndex = 2

			result, err := fn.Builtin(ctx, []*values.Value{
				arr,
				values.NewInt(1), // unshift multiple values
				values.NewInt(2),
				values.NewInt(3),
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.ToInt() != 5 {
				t.Errorf("expected length 5, got %d", result.ToInt())
			}

			// Check order: [1,2,3,4,5]
			expected := []int64{1, 2, 3, 4, 5}
			for i, exp := range expected {
				if arrPtr.Elements[int64(i)].ToInt() != exp {
					t.Errorf("expected element %d to be %d, got %d", i, exp, arrPtr.Elements[int64(i)].ToInt())
				}
			}
		})

		t.Run("unshift to empty array", func(t *testing.T) {
			arr := values.NewArray()
			result, err := fn.Builtin(ctx, []*values.Value{
				arr,
				values.NewInt(1),
				values.NewInt(2),
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.ToInt() != 2 {
				t.Errorf("expected length 2, got %d", result.ToInt())
			}

			arrPtr := arr.Data.(*values.Array)
			if arrPtr.Elements[int64(0)].ToInt() != 1 {
				t.Errorf("expected first element to be 1, got %d", arrPtr.Elements[int64(0)].ToInt())
			}
			if arrPtr.Elements[int64(1)].ToInt() != 2 {
				t.Errorf("expected second element to be 2, got %d", arrPtr.Elements[int64(1)].ToInt())
			}
		})
	})

	t.Run("array_pad", func(t *testing.T) {
		fn := functionMap["array_pad"]
		if fn == nil {
			t.Fatal("array_pad function not found")
		}

		t.Run("pad to the right", func(t *testing.T) {
			// Create test array: [1,2,3]
			arr := values.NewArray()
			arrPtr := arr.Data.(*values.Array)
			arrPtr.Elements[int64(0)] = values.NewInt(1)
			arrPtr.Elements[int64(1)] = values.NewInt(2)
			arrPtr.Elements[int64(2)] = values.NewInt(3)
			arrPtr.NextIndex = 3

			result, err := fn.Builtin(ctx, []*values.Value{
				arr,
				values.NewInt(5),  // size 5
				values.NewInt(0),  // pad with 0
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.ArrayCount() != 5 {
				t.Errorf("expected length 5, got %d", result.ArrayCount())
			}

			resultArr := result.Data.(*values.Array)
			// Check original elements
			if resultArr.Elements[int64(0)].ToInt() != 1 {
				t.Errorf("expected element 0 to be 1, got %d", resultArr.Elements[int64(0)].ToInt())
			}
			if resultArr.Elements[int64(1)].ToInt() != 2 {
				t.Errorf("expected element 1 to be 2, got %d", resultArr.Elements[int64(1)].ToInt())
			}
			// Check padded elements
			if resultArr.Elements[int64(3)].ToInt() != 0 {
				t.Errorf("expected element 3 to be 0, got %d", resultArr.Elements[int64(3)].ToInt())
			}
			if resultArr.Elements[int64(4)].ToInt() != 0 {
				t.Errorf("expected element 4 to be 0, got %d", resultArr.Elements[int64(4)].ToInt())
			}
		})

		t.Run("pad to the left", func(t *testing.T) {
			arr := values.NewArray()
			arrPtr := arr.Data.(*values.Array)
			arrPtr.Elements[int64(0)] = values.NewInt(1)
			arrPtr.Elements[int64(1)] = values.NewInt(2)
			arrPtr.NextIndex = 2

			result, err := fn.Builtin(ctx, []*values.Value{
				arr,
				values.NewInt(-4),         // size -4 (pad to left)
				values.NewString("pad"),   // pad with "pad"
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.ArrayCount() != 4 {
				t.Errorf("expected length 4, got %d", result.ArrayCount())
			}

			resultArr := result.Data.(*values.Array)
			// Check padded elements at beginning
			if resultArr.Elements[int64(0)].ToString() != "pad" {
				t.Errorf("expected element 0 to be 'pad', got '%s'", resultArr.Elements[int64(0)].ToString())
			}
			if resultArr.Elements[int64(1)].ToString() != "pad" {
				t.Errorf("expected element 1 to be 'pad', got '%s'", resultArr.Elements[int64(1)].ToString())
			}
			// Check original elements shifted
			if resultArr.Elements[int64(2)].ToInt() != 1 {
				t.Errorf("expected element 2 to be 1, got %d", resultArr.Elements[int64(2)].ToInt())
			}
			if resultArr.Elements[int64(3)].ToInt() != 2 {
				t.Errorf("expected element 3 to be 2, got %d", resultArr.Elements[int64(3)].ToInt())
			}
		})

		t.Run("no padding needed", func(t *testing.T) {
			arr := values.NewArray()
			arrPtr := arr.Data.(*values.Array)
			arrPtr.Elements[int64(0)] = values.NewInt(1)
			arrPtr.Elements[int64(1)] = values.NewInt(2)
			arrPtr.NextIndex = 2

			result, err := fn.Builtin(ctx, []*values.Value{
				arr,
				values.NewInt(2),  // same size as current
				values.NewInt(0),
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.ArrayCount() != 2 {
				t.Errorf("expected length 2, got %d", result.ArrayCount())
			}

			// Should be a copy with same values
			resultArr := result.Data.(*values.Array)
			if resultArr.Elements[int64(0)].ToInt() != 1 {
				t.Errorf("expected element 0 to be 1, got %d", resultArr.Elements[int64(0)].ToInt())
			}
			if resultArr.Elements[int64(1)].ToInt() != 2 {
				t.Errorf("expected element 1 to be 2, got %d", resultArr.Elements[int64(1)].ToInt())
			}
		})

		t.Run("pad empty array", func(t *testing.T) {
			arr := values.NewArray()

			result, err := fn.Builtin(ctx, []*values.Value{
				arr,
				values.NewInt(3),
				values.NewString("fill"),
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.ArrayCount() != 3 {
				t.Errorf("expected length 3, got %d", result.ArrayCount())
			}

			resultArr := result.Data.(*values.Array)
			for i := int64(0); i < 3; i++ {
				if resultArr.Elements[i].ToString() != "fill" {
					t.Errorf("expected element %d to be 'fill', got '%s'", i, resultArr.Elements[i].ToString())
				}
			}
		})
	})

	t.Run("array_fill", func(t *testing.T) {
		fn := functionMap["array_fill"]
		if fn == nil {
			t.Fatal("array_fill function not found")
		}

		t.Run("basic fill", func(t *testing.T) {
			result, err := fn.Builtin(ctx, []*values.Value{
				values.NewInt(0),       // start_index
				values.NewInt(3),       // count
				values.NewString("hi"), // value
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.ArrayCount() != 3 {
				t.Errorf("expected length 3, got %d", result.ArrayCount())
			}

			resultArr := result.Data.(*values.Array)
			for i := int64(0); i < 3; i++ {
				if resultArr.Elements[i].ToString() != "hi" {
					t.Errorf("expected element %d to be 'hi', got '%s'", i, resultArr.Elements[i].ToString())
				}
			}
		})

		t.Run("fill with different start index", func(t *testing.T) {
			result, err := fn.Builtin(ctx, []*values.Value{
				values.NewInt(5),   // start_index
				values.NewInt(2),   // count
				values.NewInt(42),  // value
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.ArrayCount() != 2 {
				t.Errorf("expected length 2, got %d", result.ArrayCount())
			}

			resultArr := result.Data.(*values.Array)
			if resultArr.Elements[int64(5)].ToInt() != 42 {
				t.Errorf("expected element 5 to be 42, got %d", resultArr.Elements[int64(5)].ToInt())
			}
			if resultArr.Elements[int64(6)].ToInt() != 42 {
				t.Errorf("expected element 6 to be 42, got %d", resultArr.Elements[int64(6)].ToInt())
			}
		})

		t.Run("zero count returns empty", func(t *testing.T) {
			result, err := fn.Builtin(ctx, []*values.Value{
				values.NewInt(0),
				values.NewInt(0), // count 0
				values.NewString("test"),
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.ArrayCount() != 0 {
				t.Errorf("expected empty array, got length %d", result.ArrayCount())
			}
		})
	})

	t.Run("array_fill_keys", func(t *testing.T) {
		fn := functionMap["array_fill_keys"]
		if fn == nil {
			t.Fatal("array_fill_keys function not found")
		}

		t.Run("basic fill keys", func(t *testing.T) {
			// Create keys array: ['a', 'b', 'c']
			keys := values.NewArray()
			keysPtr := keys.Data.(*values.Array)
			keysPtr.Elements[int64(0)] = values.NewString("a")
			keysPtr.Elements[int64(1)] = values.NewString("b")
			keysPtr.Elements[int64(2)] = values.NewString("c")
			keysPtr.NextIndex = 3

			result, err := fn.Builtin(ctx, []*values.Value{
				keys,
				values.NewString("value"),
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.ArrayCount() != 3 {
				t.Errorf("expected length 3, got %d", result.ArrayCount())
			}

			resultArr := result.Data.(*values.Array)
			if resultArr.Elements["a"].ToString() != "value" {
				t.Errorf("expected key 'a' to have 'value', got '%s'", resultArr.Elements["a"].ToString())
			}
			if resultArr.Elements["b"].ToString() != "value" {
				t.Errorf("expected key 'b' to have 'value', got '%s'", resultArr.Elements["b"].ToString())
			}
			if resultArr.Elements["c"].ToString() != "value" {
				t.Errorf("expected key 'c' to have 'value', got '%s'", resultArr.Elements["c"].ToString())
			}
		})

		t.Run("numeric keys", func(t *testing.T) {
			// Create keys array: [1, 2, 3]
			keys := values.NewArray()
			keysPtr := keys.Data.(*values.Array)
			keysPtr.Elements[int64(0)] = values.NewInt(1)
			keysPtr.Elements[int64(1)] = values.NewInt(2)
			keysPtr.Elements[int64(2)] = values.NewInt(3)
			keysPtr.NextIndex = 3

			result, err := fn.Builtin(ctx, []*values.Value{
				keys,
				values.NewInt(0),
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			resultArr := result.Data.(*values.Array)
			if resultArr.Elements[int64(1)].ToInt() != 0 {
				t.Errorf("expected key 1 to have 0, got %d", resultArr.Elements[int64(1)].ToInt())
			}
			if resultArr.Elements[int64(2)].ToInt() != 0 {
				t.Errorf("expected key 2 to have 0, got %d", resultArr.Elements[int64(2)].ToInt())
			}
			if resultArr.Elements[int64(3)].ToInt() != 0 {
				t.Errorf("expected key 3 to have 0, got %d", resultArr.Elements[int64(3)].ToInt())
			}
		})

		t.Run("empty keys array", func(t *testing.T) {
			keys := values.NewArray()
			result, err := fn.Builtin(ctx, []*values.Value{
				keys,
				values.NewString("value"),
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.ArrayCount() != 0 {
				t.Errorf("expected empty array, got length %d", result.ArrayCount())
			}
		})
	})

	t.Run("range", func(t *testing.T) {
		fn := functionMap["range"]
		if fn == nil {
			t.Fatal("range function not found")
		}

		t.Run("basic numeric range", func(t *testing.T) {
			result, err := fn.Builtin(ctx, []*values.Value{
				values.NewInt(1),
				values.NewInt(5),
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.ArrayCount() != 5 {
				t.Errorf("expected length 5, got %d", result.ArrayCount())
			}

			resultArr := result.Data.(*values.Array)
			expected := []int64{1, 2, 3, 4, 5}
			for i, exp := range expected {
				if resultArr.Elements[int64(i)].ToInt() != exp {
					t.Errorf("expected element %d to be %d, got %d", i, exp, resultArr.Elements[int64(i)].ToInt())
				}
			}
		})

		t.Run("range with step", func(t *testing.T) {
			result, err := fn.Builtin(ctx, []*values.Value{
				values.NewInt(0),
				values.NewInt(10),
				values.NewInt(2),
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.ArrayCount() != 6 {
				t.Errorf("expected length 6, got %d", result.ArrayCount())
			}

			resultArr := result.Data.(*values.Array)
			expected := []int64{0, 2, 4, 6, 8, 10}
			for i, exp := range expected {
				if resultArr.Elements[int64(i)].ToInt() != exp {
					t.Errorf("expected element %d to be %d, got %d", i, exp, resultArr.Elements[int64(i)].ToInt())
				}
			}
		})

		t.Run("descending range", func(t *testing.T) {
			result, err := fn.Builtin(ctx, []*values.Value{
				values.NewInt(5),
				values.NewInt(1),
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.ArrayCount() != 5 {
				t.Errorf("expected length 5, got %d", result.ArrayCount())
			}

			resultArr := result.Data.(*values.Array)
			expected := []int64{5, 4, 3, 2, 1}
			for i, exp := range expected {
				if resultArr.Elements[int64(i)].ToInt() != exp {
					t.Errorf("expected element %d to be %d, got %d", i, exp, resultArr.Elements[int64(i)].ToInt())
				}
			}
		})

		t.Run("character range", func(t *testing.T) {
			result, err := fn.Builtin(ctx, []*values.Value{
				values.NewString("a"),
				values.NewString("e"),
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.ArrayCount() != 5 {
				t.Errorf("expected length 5, got %d", result.ArrayCount())
			}

			resultArr := result.Data.(*values.Array)
			expected := []string{"a", "b", "c", "d", "e"}
			for i, exp := range expected {
				if resultArr.Elements[int64(i)].ToString() != exp {
					t.Errorf("expected element %d to be '%s', got '%s'", i, exp, resultArr.Elements[int64(i)].ToString())
				}
			}
		})

		t.Run("single element range", func(t *testing.T) {
			result, err := fn.Builtin(ctx, []*values.Value{
				values.NewInt(5),
				values.NewInt(5),
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.ArrayCount() != 1 {
				t.Errorf("expected length 1, got %d", result.ArrayCount())
			}

			resultArr := result.Data.(*values.Array)
			if resultArr.Elements[int64(0)].ToInt() != 5 {
				t.Errorf("expected element 0 to be 5, got %d", resultArr.Elements[int64(0)].ToInt())
			}
		})

		t.Run("zero step returns error", func(t *testing.T) {
			_, err := fn.Builtin(ctx, []*values.Value{
				values.NewInt(1),
				values.NewInt(5),
				values.NewInt(0), // zero step
			})

			if err == nil {
				t.Fatal("expected error for zero step")
			}
		})
	})

	t.Run("array_splice", func(t *testing.T) {
		fn := functionMap["array_splice"]
		if fn == nil {
			t.Fatal("array_splice function not found")
		}

		t.Run("basic removal", func(t *testing.T) {
			// Create test array: [1,2,3,4,5]
			arr := values.NewArray()
			arrPtr := arr.Data.(*values.Array)
			for i := int64(0); i < 5; i++ {
				arrPtr.Elements[i] = values.NewInt(i + 1)
			}
			arrPtr.NextIndex = 5

			result, err := fn.Builtin(ctx, []*values.Value{
				arr,
				values.NewInt(2), // offset
				values.NewInt(2), // length
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Check removed elements
			if result.ArrayCount() != 2 {
				t.Errorf("expected 2 removed elements, got %d", result.ArrayCount())
			}

			resultArr := result.Data.(*values.Array)
			if resultArr.Elements[int64(0)].ToInt() != 3 {
				t.Errorf("expected removed[0] to be 3, got %d", resultArr.Elements[int64(0)].ToInt())
			}
			if resultArr.Elements[int64(1)].ToInt() != 4 {
				t.Errorf("expected removed[1] to be 4, got %d", resultArr.Elements[int64(1)].ToInt())
			}

			// Check modified original array
			if arr.ArrayCount() != 3 {
				t.Errorf("expected array length 3 after splice, got %d", arr.ArrayCount())
			}

			if arrPtr.Elements[int64(0)].ToInt() != 1 {
				t.Errorf("expected element 0 to be 1, got %d", arrPtr.Elements[int64(0)].ToInt())
			}
			if arrPtr.Elements[int64(2)].ToInt() != 5 {
				t.Errorf("expected element 2 to be 5, got %d", arrPtr.Elements[int64(2)].ToInt())
			}
		})

		t.Run("splice with replacement", func(t *testing.T) {
			arr := values.NewArray()
			arrPtr := arr.Data.(*values.Array)
			for i := int64(0); i < 5; i++ {
				arrPtr.Elements[i] = values.NewInt(i + 1)
			}
			arrPtr.NextIndex = 5

			// Create replacement array: ['a', 'b']
			replacement := values.NewArray()
			replacementPtr := replacement.Data.(*values.Array)
			replacementPtr.Elements[int64(0)] = values.NewString("a")
			replacementPtr.Elements[int64(1)] = values.NewString("b")
			replacementPtr.NextIndex = 2

			removed, err := fn.Builtin(ctx, []*values.Value{
				arr,
				values.NewInt(1), // offset
				values.NewInt(2), // length
				replacement,      // replacement
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Check that elements were removed
			if removed.ArrayCount() != 2 {
				t.Errorf("expected 2 removed elements, got %d", removed.ArrayCount())
			}

			// Check that array now has replacements
			if arrPtr.Elements[int64(1)].ToString() != "a" {
				t.Errorf("expected element 1 to be 'a', got '%s'", arrPtr.Elements[int64(1)].ToString())
			}
			if arrPtr.Elements[int64(2)].ToString() != "b" {
				t.Errorf("expected element 2 to be 'b', got '%s'", arrPtr.Elements[int64(2)].ToString())
			}
		})

		t.Run("insert without removal", func(t *testing.T) {
			arr := values.NewArray()
			arrPtr := arr.Data.(*values.Array)
			arrPtr.Elements[int64(0)] = values.NewInt(1)
			arrPtr.Elements[int64(1)] = values.NewInt(2)
			arrPtr.NextIndex = 2

			// Create replacement array: ['x']
			replacement := values.NewArray()
			replacementPtr := replacement.Data.(*values.Array)
			replacementPtr.Elements[int64(0)] = values.NewString("x")
			replacementPtr.NextIndex = 1

			result, err := fn.Builtin(ctx, []*values.Value{
				arr,
				values.NewInt(1), // offset
				values.NewInt(0), // length 0 = no removal
				replacement,      // insert 'x'
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Should have removed nothing
			if result.ArrayCount() != 0 {
				t.Errorf("expected 0 removed elements, got %d", result.ArrayCount())
			}

			// Array should now be [1, 'x', 2]
			if arr.ArrayCount() != 3 {
				t.Errorf("expected array length 3 after insert, got %d", arr.ArrayCount())
			}
		})

		t.Run("remove all from offset", func(t *testing.T) {
			arr := values.NewArray()
			arrPtr := arr.Data.(*values.Array)
			for i := int64(0); i < 5; i++ {
				arrPtr.Elements[i] = values.NewInt(i + 1)
			}
			arrPtr.NextIndex = 5

			removed, err := fn.Builtin(ctx, []*values.Value{
				arr,
				values.NewInt(2), // offset
				// no length = remove all from offset
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Should have removed 3 elements (3,4,5)
			if removed.ArrayCount() != 3 {
				t.Errorf("expected 3 removed elements, got %d", removed.ArrayCount())
			}

			// Array should now have 2 elements (1,2)
			if arr.ArrayCount() != 2 {
				t.Errorf("expected array length 2 after splice, got %d", arr.ArrayCount())
			}
		})
	})

	t.Run("array_column", func(t *testing.T) {
		fn := functionMap["array_column"]
		if fn == nil {
			t.Fatal("array_column function not found")
		}

		t.Run("basic column extraction", func(t *testing.T) {
			// Create test data: [['id'=>1,'name'=>'John'],['id'=>2,'name'=>'Jane']]
			data := values.NewArray()
			dataPtr := data.Data.(*values.Array)

			// First record
			record1 := values.NewArray()
			record1Ptr := record1.Data.(*values.Array)
			record1Ptr.Elements["id"] = values.NewInt(1)
			record1Ptr.Elements["name"] = values.NewString("John")

			// Second record
			record2 := values.NewArray()
			record2Ptr := record2.Data.(*values.Array)
			record2Ptr.Elements["id"] = values.NewInt(2)
			record2Ptr.Elements["name"] = values.NewString("Jane")

			dataPtr.Elements[int64(0)] = record1
			dataPtr.Elements[int64(1)] = record2
			dataPtr.NextIndex = 2

			result, err := fn.Builtin(ctx, []*values.Value{
				data,
				values.NewString("name"),
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.ArrayCount() != 2 {
				t.Errorf("expected 2 elements, got %d", result.ArrayCount())
			}

			names := result.ArrayGet(values.NewInt(0))
			if names.Data.(string) != "John" {
				t.Errorf("expected 'John', got %s", names.Data.(string))
			}

			names = result.ArrayGet(values.NewInt(1))
			if names.Data.(string) != "Jane" {
				t.Errorf("expected 'Jane', got %s", names.Data.(string))
			}
		})

		t.Run("with index key", func(t *testing.T) {
			// Create test data
			data := values.NewArray()
			dataPtr := data.Data.(*values.Array)

			record1 := values.NewArray()
			record1Ptr := record1.Data.(*values.Array)
			record1Ptr.Elements["id"] = values.NewInt(10)
			record1Ptr.Elements["name"] = values.NewString("Alice")

			record2 := values.NewArray()
			record2Ptr := record2.Data.(*values.Array)
			record2Ptr.Elements["id"] = values.NewInt(20)
			record2Ptr.Elements["name"] = values.NewString("Bob")

			dataPtr.Elements[int64(0)] = record1
			dataPtr.Elements[int64(1)] = record2
			dataPtr.NextIndex = 2

			result, err := fn.Builtin(ctx, []*values.Value{
				data,
				values.NewString("name"), // column key
				values.NewString("id"),   // index key
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Should be indexed by id values
			resultArr := result.Data.(*values.Array)
			alice := resultArr.Elements[int64(10)]
			if alice == nil || alice.Data.(string) != "Alice" {
				t.Errorf("expected 'Alice' at key 10, got %v", alice)
			}

			bob := resultArr.Elements[int64(20)]
			if bob == nil || bob.Data.(string) != "Bob" {
				t.Errorf("expected 'Bob' at key 20, got %v", bob)
			}
		})

		t.Run("null column key returns whole elements", func(t *testing.T) {
			data := values.NewArray()
			dataPtr := data.Data.(*values.Array)

			record1 := values.NewArray()
			record1Ptr := record1.Data.(*values.Array)
			record1Ptr.Elements["id"] = values.NewInt(1)
			record1Ptr.Elements["name"] = values.NewString("Test")

			dataPtr.Elements[int64(0)] = record1
			dataPtr.NextIndex = 1

			result, err := fn.Builtin(ctx, []*values.Value{
				data,
				values.NewNull(),         // null column key
				values.NewString("id"),   // index key
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Should return whole record indexed by id
			resultArr := result.Data.(*values.Array)
			wholeRecord := resultArr.Elements[int64(1)]
			if wholeRecord == nil || wholeRecord.Type != values.TypeArray {
				t.Errorf("expected array at key 1, got %v", wholeRecord)
			}
		})

		t.Run("numeric array keys", func(t *testing.T) {
			// Test with numeric arrays like [[1,"John"],[2,"Jane"]]
			data := values.NewArray()
			dataPtr := data.Data.(*values.Array)

			record1 := values.NewArray()
			record1Ptr := record1.Data.(*values.Array)
			record1Ptr.Elements[int64(0)] = values.NewInt(1)
			record1Ptr.Elements[int64(1)] = values.NewString("John")

			record2 := values.NewArray()
			record2Ptr := record2.Data.(*values.Array)
			record2Ptr.Elements[int64(0)] = values.NewInt(2)
			record2Ptr.Elements[int64(1)] = values.NewString("Jane")

			dataPtr.Elements[int64(0)] = record1
			dataPtr.Elements[int64(1)] = record2
			dataPtr.NextIndex = 2

			result, err := fn.Builtin(ctx, []*values.Value{
				data,
				values.NewInt(1), // column 1 (names)
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.ArrayCount() != 2 {
				t.Errorf("expected 2 elements, got %d", result.ArrayCount())
			}

			name1 := result.ArrayGet(values.NewInt(0))
			if name1.Data.(string) != "John" {
				t.Errorf("expected 'John', got %s", name1.Data.(string))
			}
		})

		t.Run("missing column", func(t *testing.T) {
			data := values.NewArray()
			dataPtr := data.Data.(*values.Array)

			record1 := values.NewArray()
			record1Ptr := record1.Data.(*values.Array)
			record1Ptr.Elements["id"] = values.NewInt(1)

			dataPtr.Elements[int64(0)] = record1
			dataPtr.NextIndex = 1

			result, err := fn.Builtin(ctx, []*values.Value{
				data,
				values.NewString("nonexistent"),
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Should return empty array
			if result.ArrayCount() != 0 {
				t.Errorf("expected empty array for missing column, got %d elements", result.ArrayCount())
			}
		})

		t.Run("empty input array", func(t *testing.T) {
			data := values.NewArray()

			result, err := fn.Builtin(ctx, []*values.Value{
				data,
				values.NewString("name"),
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.ArrayCount() != 0 {
				t.Errorf("expected empty result for empty input, got %d", result.ArrayCount())
			}
		})
	})

	t.Run("array_reverse", func(t *testing.T) {
		fn := functionMap["array_reverse"]
		if fn == nil {
			t.Fatal("array_reverse function not found")
		}

		t.Run("basic reverse without preserving keys", func(t *testing.T) {
			arr := values.NewArray()
			arrPtr := arr.Data.(*values.Array)
			for i := int64(0); i < 5; i++ {
				arrPtr.Elements[i] = values.NewInt(i + 1)
			}
			arrPtr.NextIndex = 5

			result, err := fn.Builtin(ctx, []*values.Value{arr})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.ArrayCount() != 5 {
				t.Errorf("expected 5 elements, got %d", result.ArrayCount())
			}

			// Should be [5,4,3,2,1] with keys [0,1,2,3,4]
			expected := []int{5, 4, 3, 2, 1}
			for i, exp := range expected {
				val := result.ArrayGet(values.NewInt(int64(i)))
				if val.Data.(int64) != int64(exp) {
					t.Errorf("expected %d at index %d, got %d", exp, i, val.Data.(int64))
				}
			}
		})

		t.Run("reverse preserving keys", func(t *testing.T) {
			arr := values.NewArray()
			arrPtr := arr.Data.(*values.Array)
			for i := int64(0); i < 3; i++ {
				arrPtr.Elements[i] = values.NewInt(i + 1)
			}
			arrPtr.NextIndex = 3

			result, err := fn.Builtin(ctx, []*values.Value{
				arr,
				values.NewBool(true), // preserve_keys
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Should preserve original keys: [2=>3, 1=>2, 0=>1]
			resultArr := result.Data.(*values.Array)
			if resultArr.Elements[int64(2)].Data.(int64) != 3 {
				t.Errorf("expected 3 at key 2, got %v", resultArr.Elements[int64(2)])
			}
			if resultArr.Elements[int64(1)].Data.(int64) != 2 {
				t.Errorf("expected 2 at key 1, got %v", resultArr.Elements[int64(1)])
			}
			if resultArr.Elements[int64(0)].Data.(int64) != 1 {
				t.Errorf("expected 1 at key 0, got %v", resultArr.Elements[int64(0)])
			}
		})

		t.Run("associative array", func(t *testing.T) {
			arr := values.NewArray()
			arrPtr := arr.Data.(*values.Array)
			arrPtr.Elements["a"] = values.NewString("first")
			arrPtr.Elements["b"] = values.NewString("second")
			arrPtr.Elements["c"] = values.NewString("third")

			result, err := fn.Builtin(ctx, []*values.Value{arr})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// String keys are always preserved
			resultArr := result.Data.(*values.Array)
			if resultArr.Elements["c"].Data.(string) != "third" {
				t.Errorf("expected 'third' for key 'c'")
			}
		})

		t.Run("empty array", func(t *testing.T) {
			arr := values.NewArray()

			result, err := fn.Builtin(ctx, []*values.Value{arr})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.ArrayCount() != 0 {
				t.Errorf("expected empty array, got %d elements", result.ArrayCount())
			}
		})
	})

	t.Run("array_keys", func(t *testing.T) {
		fn := functionMap["array_keys"]
		if fn == nil {
			t.Fatal("array_keys function not found")
		}

		t.Run("basic keys extraction", func(t *testing.T) {
			arr := values.NewArray()
			arrPtr := arr.Data.(*values.Array)
			arrPtr.Elements["a"] = values.NewString("first")
			arrPtr.Elements["b"] = values.NewString("second")
			arrPtr.Elements["c"] = values.NewString("third")

			result, err := fn.Builtin(ctx, []*values.Value{arr})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.ArrayCount() != 3 {
				t.Errorf("expected 3 keys, got %d", result.ArrayCount())
			}

			// Should return string keys: ["a", "b", "c"]
			key0 := result.ArrayGet(values.NewInt(0))
			if key0.Data.(string) != "a" {
				t.Errorf("expected 'a', got %s", key0.Data.(string))
			}
		})

		t.Run("numeric keys", func(t *testing.T) {
			arr := values.NewArray()
			arrPtr := arr.Data.(*values.Array)
			for i := int64(0); i < 3; i++ {
				arrPtr.Elements[i] = values.NewInt((i + 1) * 10)
			}
			arrPtr.NextIndex = 3

			result, err := fn.Builtin(ctx, []*values.Value{arr})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.ArrayCount() != 3 {
				t.Errorf("expected 3 keys, got %d", result.ArrayCount())
			}

			// Should return numeric keys: [0, 1, 2]
			key0 := result.ArrayGet(values.NewInt(0))
			if key0.Data.(int64) != 0 {
				t.Errorf("expected 0, got %d", key0.Data.(int64))
			}
		})

		t.Run("search for specific value", func(t *testing.T) {
			arr := values.NewArray()
			arrPtr := arr.Data.(*values.Array)
			arrPtr.Elements["a"] = values.NewInt(1)
			arrPtr.Elements["b"] = values.NewInt(2)
			arrPtr.Elements["c"] = values.NewInt(3)

			result, err := fn.Builtin(ctx, []*values.Value{
				arr,
				values.NewInt(2), // search value
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.ArrayCount() != 1 {
				t.Errorf("expected 1 matching key, got %d", result.ArrayCount())
			}

			// Should return ["b"]
			key0 := result.ArrayGet(values.NewInt(0))
			if key0.Data.(string) != "b" {
				t.Errorf("expected 'b', got %s", key0.Data.(string))
			}
		})

		t.Run("strict search", func(t *testing.T) {
			arr := values.NewArray()
			arrPtr := arr.Data.(*values.Array)
			arrPtr.Elements[int64(0)] = values.NewInt(1)
			arrPtr.Elements[int64(1)] = values.NewString("1")

			// Strict search for int(1)
			result, err := fn.Builtin(ctx, []*values.Value{
				arr,
				values.NewInt(1),  // search value
				values.NewBool(true), // strict
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.ArrayCount() != 1 {
				t.Errorf("expected 1 matching key, got %d", result.ArrayCount())
			}

			// Should return [0] only
			key0 := result.ArrayGet(values.NewInt(0))
			if key0.Data.(int64) != 0 {
				t.Errorf("expected 0, got %d", key0.Data.(int64))
			}
		})

		t.Run("loose search", func(t *testing.T) {
			arr := values.NewArray()
			arrPtr := arr.Data.(*values.Array)
			arrPtr.Elements[int64(0)] = values.NewInt(1)
			arrPtr.Elements[int64(1)] = values.NewString("1")

			// Loose search for string("1")
			result, err := fn.Builtin(ctx, []*values.Value{
				arr,
				values.NewString("1"), // search value
				values.NewBool(false),   // not strict
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.ArrayCount() != 2 {
				t.Errorf("expected 2 matching keys, got %d", result.ArrayCount())
			}

			// Should return [0, 1]
			key0 := result.ArrayGet(values.NewInt(0))
			key1 := result.ArrayGet(values.NewInt(1))
			if key0.Data.(int64) != 0 || key1.Data.(int64) != 1 {
				t.Errorf("expected keys [0,1], got [%d,%d]", key0.Data.(int64), key1.Data.(int64))
			}
		})

		t.Run("empty array", func(t *testing.T) {
			arr := values.NewArray()

			result, err := fn.Builtin(ctx, []*values.Value{arr})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.ArrayCount() != 0 {
				t.Errorf("expected empty result, got %d keys", result.ArrayCount())
			}
		})
	})

	t.Run("array_values", func(t *testing.T) {
		fn := functionMap["array_values"]
		if fn == nil {
			t.Fatal("array_values function not found")
		}

		t.Run("associative array", func(t *testing.T) {
			arr := values.NewArray()
			arrPtr := arr.Data.(*values.Array)
			arrPtr.Elements["a"] = values.NewInt(1)
			arrPtr.Elements["b"] = values.NewInt(2)
			arrPtr.Elements["c"] = values.NewInt(3)

			result, err := fn.Builtin(ctx, []*values.Value{arr})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.ArrayCount() != 3 {
				t.Errorf("expected 3 values, got %d", result.ArrayCount())
			}

			// Should be reindexed as [0=>1, 1=>2, 2=>3]
			val0 := result.ArrayGet(values.NewInt(0))
			val1 := result.ArrayGet(values.NewInt(1))
			val2 := result.ArrayGet(values.NewInt(2))

			if val0.Data.(int64) != 1 || val1.Data.(int64) != 2 || val2.Data.(int64) != 3 {
				t.Errorf("expected values [1,2,3], got [%d,%d,%d]",
					val0.Data.(int64), val1.Data.(int64), val2.Data.(int64))
			}
		})

		t.Run("numeric array unchanged", func(t *testing.T) {
			arr := values.NewArray()
			arrPtr := arr.Data.(*values.Array)
			for i := int64(0); i < 3; i++ {
				arrPtr.Elements[i] = values.NewInt((i + 1) * 10)
			}
			arrPtr.NextIndex = 3

			result, err := fn.Builtin(ctx, []*values.Value{arr})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.ArrayCount() != 3 {
				t.Errorf("expected 3 values, got %d", result.ArrayCount())
			}

			// Should remain [0=>10, 1=>20, 2=>30]
			val0 := result.ArrayGet(values.NewInt(0))
			val1 := result.ArrayGet(values.NewInt(1))
			val2 := result.ArrayGet(values.NewInt(2))

			if val0.Data.(int64) != 10 || val1.Data.(int64) != 20 || val2.Data.(int64) != 30 {
				t.Errorf("expected values [10,20,30], got [%d,%d,%d]",
					val0.Data.(int64), val1.Data.(int64), val2.Data.(int64))
			}
		})

		t.Run("mixed keys array", func(t *testing.T) {
			arr := values.NewArray()
			arrPtr := arr.Data.(*values.Array)
			arrPtr.Elements[int64(0)] = values.NewString("a")
			arrPtr.Elements["x"] = values.NewString("b")
			arrPtr.Elements[int64(1)] = values.NewString("c")
			arrPtr.Elements["y"] = values.NewString("d")

			result, err := fn.Builtin(ctx, []*values.Value{arr})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.ArrayCount() != 4 {
				t.Errorf("expected 4 values, got %d", result.ArrayCount())
			}

			// Should be reindexed with values in order
			val0 := result.ArrayGet(values.NewInt(0))
			val1 := result.ArrayGet(values.NewInt(1))
			val2 := result.ArrayGet(values.NewInt(2))
			val3 := result.ArrayGet(values.NewInt(3))

			// Numeric keys come first (0, 1), then string keys (x, y)
			if val0.Data.(string) != "a" || val1.Data.(string) != "c" ||
			   val2.Data.(string) != "b" || val3.Data.(string) != "d" {
				t.Errorf("expected values [a,c,b,d], got [%s,%s,%s,%s]",
					val0.Data.(string), val1.Data.(string), val2.Data.(string), val3.Data.(string))
			}
		})

		t.Run("sparse array", func(t *testing.T) {
			arr := values.NewArray()
			arrPtr := arr.Data.(*values.Array)
			arrPtr.Elements[int64(0)] = values.NewString("first")
			arrPtr.Elements[int64(5)] = values.NewString("second")
			arrPtr.Elements[int64(10)] = values.NewString("third")

			result, err := fn.Builtin(ctx, []*values.Value{arr})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.ArrayCount() != 3 {
				t.Errorf("expected 3 values, got %d", result.ArrayCount())
			}

			// Should be reindexed as [0,1,2]
			val0 := result.ArrayGet(values.NewInt(0))
			val1 := result.ArrayGet(values.NewInt(1))
			val2 := result.ArrayGet(values.NewInt(2))

			if val0.Data.(string) != "first" || val1.Data.(string) != "second" || val2.Data.(string) != "third" {
				t.Errorf("expected values [first,second,third], got [%s,%s,%s]",
					val0.Data.(string), val1.Data.(string), val2.Data.(string))
			}
		})

		t.Run("array with null values", func(t *testing.T) {
			arr := values.NewArray()
			arrPtr := arr.Data.(*values.Array)
			arrPtr.Elements["a"] = values.NewNull()
			arrPtr.Elements["b"] = values.NewString("test")
			arrPtr.Elements["c"] = values.NewNull()

			result, err := fn.Builtin(ctx, []*values.Value{arr})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.ArrayCount() != 3 {
				t.Errorf("expected 3 values, got %d", result.ArrayCount())
			}

			// Should preserve null values
			val0 := result.ArrayGet(values.NewInt(0))
			val1 := result.ArrayGet(values.NewInt(1))
			val2 := result.ArrayGet(values.NewInt(2))

			if !val0.IsNull() || val1.Data.(string) != "test" || !val2.IsNull() {
				t.Errorf("expected [null,test,null], got [%v,%v,%v]", val0, val1, val2)
			}
		})

		t.Run("empty array", func(t *testing.T) {
			arr := values.NewArray()

			result, err := fn.Builtin(ctx, []*values.Value{arr})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.ArrayCount() != 0 {
				t.Errorf("expected empty result, got %d values", result.ArrayCount())
			}
		})
	})

	t.Run("array_merge", func(t *testing.T) {
		fn := functionMap["array_merge"]
		if fn == nil {
			t.Fatal("array_merge function not found")
		}

		t.Run("merge two numeric arrays", func(t *testing.T) {
			arr1 := values.NewArray()
			arr1Ptr := arr1.Data.(*values.Array)
			for i := int64(0); i < 3; i++ {
				arr1Ptr.Elements[i] = values.NewInt(i + 1)
			}
			arr1Ptr.NextIndex = 3

			arr2 := values.NewArray()
			arr2Ptr := arr2.Data.(*values.Array)
			for i := int64(0); i < 3; i++ {
				arr2Ptr.Elements[i] = values.NewInt(i + 4)
			}
			arr2Ptr.NextIndex = 3

			result, err := fn.Builtin(ctx, []*values.Value{arr1, arr2})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.ArrayCount() != 6 {
				t.Errorf("expected 6 elements, got %d", result.ArrayCount())
			}

			// Should be [1,2,3,4,5,6] with keys [0,1,2,3,4,5]
			for i := int64(0); i < 6; i++ {
				val := result.ArrayGet(values.NewInt(i))
				expected := i + 1
				if val.Data.(int64) != expected {
					t.Errorf("expected %d at index %d, got %d", expected, i, val.Data.(int64))
				}
			}
		})

		t.Run("merge associative arrays", func(t *testing.T) {
			arr1 := values.NewArray()
			arr1Ptr := arr1.Data.(*values.Array)
			arr1Ptr.Elements["a"] = values.NewInt(1)
			arr1Ptr.Elements["b"] = values.NewInt(2)

			arr2 := values.NewArray()
			arr2Ptr := arr2.Data.(*values.Array)
			arr2Ptr.Elements["c"] = values.NewInt(3)
			arr2Ptr.Elements["d"] = values.NewInt(4)

			result, err := fn.Builtin(ctx, []*values.Value{arr1, arr2})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.ArrayCount() != 4 {
				t.Errorf("expected 4 elements, got %d", result.ArrayCount())
			}

			// Should preserve string keys
			resultArr := result.Data.(*values.Array)
			if resultArr.Elements["a"].Data.(int64) != 1 ||
			   resultArr.Elements["b"].Data.(int64) != 2 ||
			   resultArr.Elements["c"].Data.(int64) != 3 ||
			   resultArr.Elements["d"].Data.(int64) != 4 {
				t.Errorf("associative keys not preserved correctly")
			}
		})

		t.Run("overlapping string keys", func(t *testing.T) {
			arr1 := values.NewArray()
			arr1Ptr := arr1.Data.(*values.Array)
			arr1Ptr.Elements["a"] = values.NewInt(1)
			arr1Ptr.Elements["b"] = values.NewInt(2)

			arr2 := values.NewArray()
			arr2Ptr := arr2.Data.(*values.Array)
			arr2Ptr.Elements["b"] = values.NewInt(20) // Should overwrite
			arr2Ptr.Elements["c"] = values.NewInt(3)

			result, err := fn.Builtin(ctx, []*values.Value{arr1, arr2})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.ArrayCount() != 3 {
				t.Errorf("expected 3 elements, got %d", result.ArrayCount())
			}

			// b should be overwritten to 20
			resultArr := result.Data.(*values.Array)
			if resultArr.Elements["b"].Data.(int64) != 20 {
				t.Errorf("expected b to be overwritten to 20, got %d", resultArr.Elements["b"].Data.(int64))
			}
		})

		t.Run("numeric keys get reindexed", func(t *testing.T) {
			arr1 := values.NewArray()
			arr1Ptr := arr1.Data.(*values.Array)
			arr1Ptr.Elements[int64(10)] = values.NewString("a")
			arr1Ptr.Elements[int64(20)] = values.NewString("b")

			arr2 := values.NewArray()
			arr2Ptr := arr2.Data.(*values.Array)
			arr2Ptr.Elements[int64(30)] = values.NewString("c")
			arr2Ptr.Elements[int64(40)] = values.NewString("d")

			result, err := fn.Builtin(ctx, []*values.Value{arr1, arr2})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.ArrayCount() != 4 {
				t.Errorf("expected 4 elements, got %d", result.ArrayCount())
			}

			// Should be reindexed as [0,1,2,3]
			expected := []string{"a", "b", "c", "d"}
			for i, exp := range expected {
				val := result.ArrayGet(values.NewInt(int64(i)))
				if val.Data.(string) != exp {
					t.Errorf("expected %s at index %d, got %s", exp, i, val.Data.(string))
				}
			}
		})

		t.Run("multiple arrays", func(t *testing.T) {
			arr1 := values.NewArray()
			arr1Ptr := arr1.Data.(*values.Array)
			arr1Ptr.Elements[int64(0)] = values.NewInt(1)
			arr1Ptr.Elements[int64(1)] = values.NewInt(2)

			arr2 := values.NewArray()
			arr2Ptr := arr2.Data.(*values.Array)
			arr2Ptr.Elements[int64(0)] = values.NewInt(3)
			arr2Ptr.Elements[int64(1)] = values.NewInt(4)

			arr3 := values.NewArray()
			arr3Ptr := arr3.Data.(*values.Array)
			arr3Ptr.Elements[int64(0)] = values.NewInt(5)
			arr3Ptr.Elements[int64(1)] = values.NewInt(6)

			result, err := fn.Builtin(ctx, []*values.Value{arr1, arr2, arr3})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.ArrayCount() != 6 {
				t.Errorf("expected 6 elements, got %d", result.ArrayCount())
			}

			// Should be [1,2,3,4,5,6]
			expected := []int{1, 2, 3, 4, 5, 6}
			for i, exp := range expected {
				val := result.ArrayGet(values.NewInt(int64(i)))
				if val.Data.(int64) != int64(exp) {
					t.Errorf("expected %d at index %d, got %d", exp, i, val.Data.(int64))
				}
			}
		})

		t.Run("empty arrays", func(t *testing.T) {
			arr1 := values.NewArray() // empty
			arr2 := values.NewArray()
			arr2Ptr := arr2.Data.(*values.Array)
			arr2Ptr.Elements[int64(0)] = values.NewInt(1)
			arr2Ptr.Elements[int64(1)] = values.NewInt(2)

			arr3 := values.NewArray() // empty

			result, err := fn.Builtin(ctx, []*values.Value{arr1, arr2, arr3})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.ArrayCount() != 2 {
				t.Errorf("expected 2 elements, got %d", result.ArrayCount())
			}
		})

		t.Run("no arguments", func(t *testing.T) {
			result, err := fn.Builtin(ctx, []*values.Value{})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.ArrayCount() != 0 {
				t.Errorf("expected empty result, got %d elements", result.ArrayCount())
			}
		})
	})

	t.Run("array_unique", func(t *testing.T) {
		fn := functionMap["array_unique"]
		if fn == nil {
			t.Fatal("array_unique function not found")
		}

		t.Run("numeric duplicates", func(t *testing.T) {
			arr := values.NewArray()
			arrPtr := arr.Data.(*values.Array)

			// [1,2,2,3,1,4,3]
			values_list := []int64{1, 2, 2, 3, 1, 4, 3}
			for i, val := range values_list {
				arrPtr.Elements[int64(i)] = values.NewInt(val)
			}
			arrPtr.NextIndex = int64(len(values_list))

			result, err := fn.Builtin(ctx, []*values.Value{arr})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Should preserve keys of first occurrences: [0=>1, 1=>2, 3=>3, 5=>4]
			if result.ArrayCount() != 4 {
				t.Errorf("expected 4 unique elements, got %d", result.ArrayCount())
			}

			resultArr := result.Data.(*values.Array)
			if resultArr.Elements[int64(0)].Data.(int64) != 1 ||
			   resultArr.Elements[int64(1)].Data.(int64) != 2 ||
			   resultArr.Elements[int64(3)].Data.(int64) != 3 ||
			   resultArr.Elements[int64(5)].Data.(int64) != 4 {
				t.Errorf("unique values not correct or keys not preserved")
			}
		})

		t.Run("string duplicates", func(t *testing.T) {
			arr := values.NewArray()
			arrPtr := arr.Data.(*values.Array)

			// ['a','b','b','c','a']
			strings_list := []string{"a", "b", "b", "c", "a"}
			for i, val := range strings_list {
				arrPtr.Elements[int64(i)] = values.NewString(val)
			}
			arrPtr.NextIndex = int64(len(strings_list))

			result, err := fn.Builtin(ctx, []*values.Value{arr})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Should preserve keys: [0=>'a', 1=>'b', 3=>'c']
			if result.ArrayCount() != 3 {
				t.Errorf("expected 3 unique elements, got %d", result.ArrayCount())
			}

			resultArr := result.Data.(*values.Array)
			if resultArr.Elements[int64(0)].Data.(string) != "a" ||
			   resultArr.Elements[int64(1)].Data.(string) != "b" ||
			   resultArr.Elements[int64(3)].Data.(string) != "c" {
				t.Errorf("unique string values not correct or keys not preserved")
			}
		})

		t.Run("associative array with duplicate values", func(t *testing.T) {
			arr := values.NewArray()
			arrPtr := arr.Data.(*values.Array)
			arrPtr.Elements["x"] = values.NewString("a")
			arrPtr.Elements["y"] = values.NewString("b")
			arrPtr.Elements["z"] = values.NewString("b") // duplicate
			arrPtr.Elements["w"] = values.NewString("c")

			result, err := fn.Builtin(ctx, []*values.Value{arr})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Should preserve first keys: [x=>'a', y=>'b', w=>'c']
			if result.ArrayCount() != 3 {
				t.Errorf("expected 3 unique elements, got %d", result.ArrayCount())
			}

			resultArr := result.Data.(*values.Array)
			if resultArr.Elements["x"].Data.(string) != "a" ||
			   resultArr.Elements["y"].Data.(string) != "b" ||
			   resultArr.Elements["w"].Data.(string) != "c" {
				t.Errorf("unique associative values not correct")
			}

			// z key should be removed (duplicate value)
			if resultArr.Elements["z"] != nil {
				t.Errorf("duplicate key 'z' should have been removed")
			}
		})

		t.Run("with null values", func(t *testing.T) {
			arr := values.NewArray()
			arrPtr := arr.Data.(*values.Array)
			arrPtr.Elements[int64(0)] = values.NewNull()
			arrPtr.Elements[int64(1)] = values.NewString("test")
			arrPtr.Elements[int64(2)] = values.NewNull() // duplicate
			arrPtr.Elements[int64(3)] = values.NewString("other")

			result, err := fn.Builtin(ctx, []*values.Value{arr})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Should be [0=>null, 1=>'test', 3=>'other']
			if result.ArrayCount() != 3 {
				t.Errorf("expected 3 unique elements, got %d", result.ArrayCount())
			}

			resultArr := result.Data.(*values.Array)
			if !resultArr.Elements[int64(0)].IsNull() ||
			   resultArr.Elements[int64(1)].Data.(string) != "test" ||
			   resultArr.Elements[int64(3)].Data.(string) != "other" {
				t.Errorf("unique null values not handled correctly")
			}

			// Index 2 should be removed (duplicate null)
			if resultArr.Elements[int64(2)] != nil {
				t.Errorf("duplicate null at index 2 should have been removed")
			}
		})

		t.Run("empty array", func(t *testing.T) {
			arr := values.NewArray()

			result, err := fn.Builtin(ctx, []*values.Value{arr})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.ArrayCount() != 0 {
				t.Errorf("expected empty result, got %d elements", result.ArrayCount())
			}
		})

		t.Run("single element", func(t *testing.T) {
			arr := values.NewArray()
			arrPtr := arr.Data.(*values.Array)
			arrPtr.Elements[int64(0)] = values.NewString("test")

			result, err := fn.Builtin(ctx, []*values.Value{arr})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.ArrayCount() != 1 {
				t.Errorf("expected 1 element, got %d", result.ArrayCount())
			}

			val := result.ArrayGet(values.NewInt(0))
			if val.Data.(string) != "test" {
				t.Errorf("expected 'test', got %s", val.Data.(string))
			}
		})
	})
}