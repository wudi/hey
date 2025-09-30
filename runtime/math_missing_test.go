package runtime

import (
	"fmt"
	"testing"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// Test max function
func TestMaxFunction(t *testing.T) {
	functions := GetMathFunctions()

	var maxFunc *registry.Function
	for _, f := range functions {
		if f.Name == "max" {
			maxFunc = f
			break
		}
	}

	if maxFunc == nil {
		t.Fatal("max function not found")
	}

	ctx := &mockMathBuiltinCallContext{}

	tests := []struct {
		name string
		args []*values.Value
		want string // Use string for easier comparison
	}{
		{
			"max with integers",
			[]*values.Value{values.NewInt(1), values.NewInt(3), values.NewInt(2)},
			"3",
		},
		{
			"max with single array",
			[]*values.Value{createIntArray([]int64{1, 2, 3})},
			"3",
		},
		{
			"max with strings",
			[]*values.Value{values.NewString("a"), values.NewString("c"), values.NewString("b")},
			"c",
		},
		{
			"max with floats",
			[]*values.Value{values.NewFloat(1.5), values.NewFloat(2.7)},
			"2.7",
		},
		{
			"max with mixed types",
			[]*values.Value{values.NewInt(1), values.NewFloat(2.5), values.NewInt(2)},
			"2.5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := maxFunc.Builtin(ctx, tt.args)
			if err != nil {
				t.Fatalf("max failed: %v", err)
			}
			if result.ToString() != tt.want {
				t.Errorf("max() = %v, want %v", result.ToString(), tt.want)
			}
		})
	}
}

// Test min function
func TestMinFunction(t *testing.T) {
	functions := GetMathFunctions()

	var minFunc *registry.Function
	for _, f := range functions {
		if f.Name == "min" {
			minFunc = f
			break
		}
	}

	if minFunc == nil {
		t.Fatal("min function not found")
	}

	ctx := &mockMathBuiltinCallContext{}

	tests := []struct {
		name string
		args []*values.Value
		want string
	}{
		{
			"min with integers",
			[]*values.Value{values.NewInt(3), values.NewInt(1), values.NewInt(2)},
			"1",
		},
		{
			"min with single array",
			[]*values.Value{createIntArray([]int64{3, 1, 2})},
			"1",
		},
		{
			"min with strings",
			[]*values.Value{values.NewString("c"), values.NewString("a"), values.NewString("b")},
			"a",
		},
		{
			"min with floats",
			[]*values.Value{values.NewFloat(2.7), values.NewFloat(1.5)},
			"1.5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := minFunc.Builtin(ctx, tt.args)
			if err != nil {
				t.Fatalf("min failed: %v", err)
			}
			if result.ToString() != tt.want {
				t.Errorf("min() = %v, want %v", result.ToString(), tt.want)
			}
		})
	}
}

// Test ceil function
func TestCeilFunction(t *testing.T) {
	functions := GetMathFunctions()

	var ceilFunc *registry.Function
	for _, f := range functions {
		if f.Name == "ceil" {
			ceilFunc = f
			break
		}
	}

	if ceilFunc == nil {
		t.Fatal("ceil function not found")
	}

	ctx := &mockMathBuiltinCallContext{}

	tests := []struct {
		name string
		arg  *values.Value
		want int64
	}{
		{"ceil(1.2)", values.NewFloat(1.2), 2},
		{"ceil(-1.2)", values.NewFloat(-1.2), -1},
		{"ceil(5)", values.NewInt(5), 5},
		{"ceil(0)", values.NewInt(0), 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ceilFunc.Builtin(ctx, []*values.Value{tt.arg})
			if err != nil {
				t.Fatalf("ceil failed: %v", err)
			}
			if result.ToInt() != tt.want {
				t.Errorf("ceil() = %v, want %v", result.ToInt(), tt.want)
			}
		})
	}
}

// Test floor function
func TestFloorFunction(t *testing.T) {
	functions := GetMathFunctions()

	var floorFunc *registry.Function
	for _, f := range functions {
		if f.Name == "floor" {
			floorFunc = f
			break
		}
	}

	if floorFunc == nil {
		t.Fatal("floor function not found")
	}

	ctx := &mockMathBuiltinCallContext{}

	tests := []struct {
		name string
		arg  *values.Value
		want int64
	}{
		{"floor(1.8)", values.NewFloat(1.8), 1},
		{"floor(-1.8)", values.NewFloat(-1.8), -2},
		{"floor(5)", values.NewInt(5), 5},
		{"floor(0)", values.NewInt(0), 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := floorFunc.Builtin(ctx, []*values.Value{tt.arg})
			if err != nil {
				t.Fatalf("floor failed: %v", err)
			}
			if result.ToInt() != tt.want {
				t.Errorf("floor() = %v, want %v", result.ToInt(), tt.want)
			}
		})
	}
}

// Helper function to create integer array
func createIntArray(nums []int64) *values.Value {
	array := values.NewArray()
	for i, v := range nums {
		array.ArraySet(values.NewInt(int64(i)), values.NewInt(v))
	}
	return array
}

// Mock builtin call context for testing
type mockMathBuiltinCallContext struct{}

func (m *mockMathBuiltinCallContext) WriteOutput(val *values.Value) error { return nil }
func (m *mockMathBuiltinCallContext) GetGlobal(name string) (*values.Value, bool) { return nil, false }
func (m *mockMathBuiltinCallContext) SetGlobal(name string, val *values.Value) {}
func (m *mockMathBuiltinCallContext) SymbolRegistry() *registry.Registry { return nil }
func (m *mockMathBuiltinCallContext) LookupUserFunction(name string) (*registry.Function, bool) { return nil, false }
func (m *mockMathBuiltinCallContext) CallUserFunction(function *registry.Function, args []*values.Value) (*values.Value, error) { return nil, nil }
func (m *mockMathBuiltinCallContext) SimpleCallUserFunction(function *registry.Function, args []*values.Value) (*values.Value, error) { return nil, nil }
func (m *mockMathBuiltinCallContext) LookupUserClass(name string) (*registry.Class, bool) { return nil, false }
func (m *mockMathBuiltinCallContext) Halt(exitCode int, message string) error { return nil }
func (m *mockMathBuiltinCallContext) GetExecutionContext() registry.ExecutionContextInterface { return nil }
func (m *mockMathBuiltinCallContext) GetOutputBufferStack() registry.OutputBufferStackInterface { return nil }
func (m *mockMathBuiltinCallContext) GetCurrentFunctionArgCount() (int, error) { return 0, nil }
func (m *mockMathBuiltinCallContext) GetCurrentFunctionArg(index int) (*values.Value, error) { return nil, nil }
func (m *mockMathBuiltinCallContext) GetCurrentFunctionArgs() ([]*values.Value, error) { return nil, nil }
func (m *mockMathBuiltinCallContext) ThrowException(exception *values.Value) error { return fmt.Errorf("exception thrown in test mock: %v", exception) }
func (m *mockMathBuiltinCallContext) GetHTTPContext() registry.HTTPContext { return &mockHTTPContext{} }
func (m *mockMathBuiltinCallContext) ResetHTTPContext() {}
func (m *mockMathBuiltinCallContext) RemoveHTTPHeader(name string) {}