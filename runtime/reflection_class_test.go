package runtime

import (
	"fmt"
	"testing"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// Test property_exists function
func TestPropertyExistsFunction(t *testing.T) {
	functions := GetReflectionFunctions()

	var propertyExistsFunc *registry.Function
	for _, f := range functions {
		if f.Name == "property_exists" {
			propertyExistsFunc = f
			break
		}
	}

	if propertyExistsFunc == nil {
		t.Fatal("property_exists function not found")
	}

	// Initialize registry
	registry.Initialize()
	ctx := &mockReflectionBuiltinCallContext{registry: registry.GlobalRegistry}

	tests := []struct {
		name     string
		args     []*values.Value
		expected bool
	}{
		{
			"object with existing property",
			[]*values.Value{createTestObject(), values.NewString("testProp")},
			true,
		},
		{
			"object with non-existing property",
			[]*values.Value{createTestObject(), values.NewString("nonexistent")},
			false,
		},
		{
			"class string with property",
			[]*values.Value{values.NewString("TestClass"), values.NewString("testProp")},
			false, // Will be false in our simplified implementation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := propertyExistsFunc.Builtin(ctx, tt.args)
			if err != nil {
				t.Fatalf("property_exists failed: %v", err)
			}
			if result.ToBool() != tt.expected {
				t.Errorf("property_exists() = %v, want %v", result.ToBool(), tt.expected)
			}
		})
	}
}

// Test is_a function
func TestIsAFunction(t *testing.T) {
	functions := GetReflectionFunctions()

	var isAFunc *registry.Function
	for _, f := range functions {
		if f.Name == "is_a" {
			isAFunc = f
			break
		}
	}

	if isAFunc == nil {
		t.Fatal("is_a function not found")
	}

	ctx := &mockReflectionBuiltinCallContext{}

	tests := []struct {
		name     string
		args     []*values.Value
		expected bool
	}{
		{
			"object is own class",
			[]*values.Value{createTestObject(), values.NewString("TestClass")},
			true,
		},
		{
			"object is not different class",
			[]*values.Value{createTestObject(), values.NewString("OtherClass")},
			false,
		},
		{
			"string class with allow_string true",
			[]*values.Value{values.NewString("TestClass"), values.NewString("TestClass"), values.NewBool(true)},
			true,
		},
		{
			"string class with allow_string false",
			[]*values.Value{values.NewString("TestClass"), values.NewString("TestClass"), values.NewBool(false)},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := isAFunc.Builtin(ctx, tt.args)
			if err != nil {
				t.Fatalf("is_a failed: %v", err)
			}
			if result.ToBool() != tt.expected {
				t.Errorf("is_a() = %v, want %v", result.ToBool(), tt.expected)
			}
		})
	}
}

// Helper function to create test object
func createTestObject() *values.Value {
	obj := values.NewObject("TestClass")
	obj.ObjectSet("testProp", values.NewString("testValue"))
	return obj
}

// Mock builtin call context for testing
type mockReflectionBuiltinCallContext struct {
	registry *registry.Registry
}

func (m *mockReflectionBuiltinCallContext) WriteOutput(val *values.Value) error { return nil }
func (m *mockReflectionBuiltinCallContext) GetGlobal(name string) (*values.Value, bool) { return nil, false }
func (m *mockReflectionBuiltinCallContext) SetGlobal(name string, val *values.Value) {}
func (m *mockReflectionBuiltinCallContext) SymbolRegistry() *registry.Registry { return m.registry }
func (m *mockReflectionBuiltinCallContext) LookupUserFunction(name string) (*registry.Function, bool) { return nil, false }
func (m *mockReflectionBuiltinCallContext) CallUserFunction(function *registry.Function, args []*values.Value) (*values.Value, error) { return nil, nil }
func (m *mockReflectionBuiltinCallContext) SimpleCallUserFunction(function *registry.Function, args []*values.Value) (*values.Value, error) { return nil, nil }
func (m *mockReflectionBuiltinCallContext) LookupUserClass(name string) (*registry.Class, bool) { return nil, false }
func (m *mockReflectionBuiltinCallContext) Halt(exitCode int, message string) error { return nil }
func (m *mockReflectionBuiltinCallContext) GetExecutionContext() registry.ExecutionContextInterface { return nil }
func (m *mockReflectionBuiltinCallContext) GetOutputBufferStack() registry.OutputBufferStackInterface { return nil }
func (m *mockReflectionBuiltinCallContext) GetCurrentFunctionArgCount() (int, error) { return 0, nil }
func (m *mockReflectionBuiltinCallContext) GetCurrentFunctionArg(index int) (*values.Value, error) { return nil, nil }
func (m *mockReflectionBuiltinCallContext) GetCurrentFunctionArgs() ([]*values.Value, error) { return nil, nil }
func (m *mockReflectionBuiltinCallContext) ThrowException(exception *values.Value) error { return fmt.Errorf("exception thrown in test mock: %v", exception) }
func (m *mockReflectionBuiltinCallContext) GetHTTPContext() registry.HTTPContext { return &mockHTTPContext{} }
func (m *mockReflectionBuiltinCallContext) ResetHTTPContext() {}
func (m *mockReflectionBuiltinCallContext) RemoveHTTPHeader(name string) {}