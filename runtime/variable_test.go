package runtime

import (
	"fmt"
	"testing"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// Test isset function
func TestIsset(t *testing.T) {
	functions := GetVariableFunctions()

	var issetFunc *registry.Function
	for _, f := range functions {
		if f.Name == "isset" {
			issetFunc = f
			break
		}
	}

	if issetFunc == nil {
		t.Fatal("isset function not found")
	}

	// Create mock builtin call context
	ctx := &mockBuiltinCallContext{}

	tests := []struct {
		name string
		args []*values.Value
		want bool
	}{
		{"isset with value", []*values.Value{values.NewString("test")}, true},
		{"isset with null", []*values.Value{values.NewNull()}, false},
		{"isset with zero", []*values.Value{values.NewInt(0)}, true},
		{"isset with false", []*values.Value{values.NewBool(false)}, true},
		{"isset multiple values", []*values.Value{values.NewString("test"), values.NewInt(42)}, true},
		{"isset multiple with null", []*values.Value{values.NewString("test"), values.NewNull()}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := issetFunc.Builtin(ctx, tt.args)
			if err != nil {
				t.Fatalf("isset failed: %v", err)
			}
			if result.ToBool() != tt.want {
				t.Errorf("isset() = %v, want %v", result.ToBool(), tt.want)
			}
		})
	}
}

// Test define and defined functions
func TestDefineAndDefined(t *testing.T) {
	functions := GetVariableFunctions()

	var defineFunc, definedFunc *registry.Function
	for _, f := range functions {
		if f.Name == "define" {
			defineFunc = f
		} else if f.Name == "defined" {
			definedFunc = f
		}
	}

	if defineFunc == nil || definedFunc == nil {
		t.Fatal("define or defined function not found")
	}

	// Initialize registry
	registry.Initialize()
	ctx := &mockBuiltinCallContext{registry: registry.GlobalRegistry}

	// Test define
	result, err := defineFunc.Builtin(ctx, []*values.Value{
		values.NewString("TEST_CONSTANT"),
		values.NewString("test_value"),
	})
	if err != nil {
		t.Fatalf("define failed: %v", err)
	}
	if !result.ToBool() {
		t.Error("define() should return true on success")
	}

	// Test defined - should return true for existing constant
	result, err = definedFunc.Builtin(ctx, []*values.Value{
		values.NewString("TEST_CONSTANT"),
	})
	if err != nil {
		t.Fatalf("defined failed: %v", err)
	}
	if !result.ToBool() {
		t.Error("defined() should return true for existing constant")
	}

	// Test defined - should return false for non-existing constant
	result, err = definedFunc.Builtin(ctx, []*values.Value{
		values.NewString("NONEXISTENT_CONSTANT"),
	})
	if err != nil {
		t.Fatalf("defined failed: %v", err)
	}
	if result.ToBool() {
		t.Error("defined() should return false for non-existing constant")
	}
}

// Test constant function
func TestConstant(t *testing.T) {
	functions := GetVariableFunctions()

	var constantFunc *registry.Function
	for _, f := range functions {
		if f.Name == "constant" {
			constantFunc = f
			break
		}
	}

	if constantFunc == nil {
		t.Fatal("constant function not found")
	}

	// Initialize registry and add a test constant
	registry.Initialize()
	ctx := &mockBuiltinCallContext{registry: registry.GlobalRegistry}

	testConstant := &registry.ConstantDescriptor{
		Name:       "TEST_CONSTANT",
		Visibility: "public",
		Value:      values.NewString("test_value"),
		IsFinal:    true,
	}
	ctx.registry.RegisterConstant(testConstant)

	// Test getting existing constant
	result, err := constantFunc.Builtin(ctx, []*values.Value{
		values.NewString("TEST_CONSTANT"),
	})
	if err != nil {
		t.Fatalf("constant failed: %v", err)
	}
	if result.ToString() != "test_value" {
		t.Errorf("constant() = %v, want %v", result.ToString(), "test_value")
	}

	// Test getting non-existing constant - should return null
	result, err = constantFunc.Builtin(ctx, []*values.Value{
		values.NewString("NONEXISTENT"),
	})
	if err != nil {
		t.Fatalf("constant failed: %v", err)
	}
	if !result.IsNull() {
		t.Error("constant() should return null for non-existing constant")
	}
}

// Test get_defined_constants function
func TestGetDefinedConstants(t *testing.T) {
	functions := GetVariableFunctions()

	var getConstantsFunc *registry.Function
	for _, f := range functions {
		if f.Name == "get_defined_constants" {
			getConstantsFunc = f
			break
		}
	}

	if getConstantsFunc == nil {
		t.Fatal("get_defined_constants function not found")
	}

	// Initialize registry with builtin constants
	registry.Initialize()
	ctx := &mockBuiltinCallContext{registry: registry.GlobalRegistry}

	// Add builtin constants
	builtinConstants := GetAllBuiltinConstants()
	for _, constant := range builtinConstants {
		ctx.registry.RegisterConstant(constant)
	}

	// Test flat constants
	result, err := getConstantsFunc.Builtin(ctx, []*values.Value{})
	if err != nil {
		t.Fatalf("get_defined_constants failed: %v", err)
	}

	if !result.IsArray() {
		t.Error("get_defined_constants() should return array")
	}

	// Test grouped constants
	result, err = getConstantsFunc.Builtin(ctx, []*values.Value{values.NewBool(true)})
	if err != nil {
		t.Fatalf("get_defined_constants with grouping failed: %v", err)
	}

	if !result.IsArray() {
		t.Error("get_defined_constants(true) should return array")
	}

	// Should have Core group
	coreGroup := result.ArrayGet(values.NewString("Core"))
	if coreGroup.IsNull() {
		t.Error("get_defined_constants(true) should have 'Core' group")
	}
	if !coreGroup.IsArray() {
		t.Error("Core group should be an array")
	}
}

// Test get_defined_vars function
func TestGetDefinedVars(t *testing.T) {
	functions := GetVariableFunctions()

	var getVarsFunc *registry.Function
	for _, f := range functions {
		if f.Name == "get_defined_vars" {
			getVarsFunc = f
			break
		}
	}

	if getVarsFunc == nil {
		t.Fatal("get_defined_vars function not found")
	}

	ctx := &mockBuiltinCallContext{}

	// Test basic functionality - should return empty array for now
	result, err := getVarsFunc.Builtin(ctx, []*values.Value{})
	if err != nil {
		t.Fatalf("get_defined_vars failed: %v", err)
	}

	if !result.IsArray() {
		t.Error("get_defined_vars() should return array")
	}
}

// Mock builtin call context for testing
type mockBuiltinCallContext struct {
	registry *registry.Registry
}

func (m *mockBuiltinCallContext) WriteOutput(val *values.Value) error { return nil }
func (m *mockBuiltinCallContext) GetGlobal(name string) (*values.Value, bool) { return nil, false }
func (m *mockBuiltinCallContext) SetGlobal(name string, val *values.Value) {}
func (m *mockBuiltinCallContext) SymbolRegistry() *registry.Registry { return m.registry }
func (m *mockBuiltinCallContext) LookupUserFunction(name string) (*registry.Function, bool) { return nil, false }
func (m *mockBuiltinCallContext) CallUserFunction(function *registry.Function, args []*values.Value) (*values.Value, error) { return nil, nil }
func (m *mockBuiltinCallContext) SimpleCallUserFunction(function *registry.Function, args []*values.Value) (*values.Value, error) { return nil, nil }
func (m *mockBuiltinCallContext) LookupUserClass(name string) (*registry.Class, bool) { return nil, false }
func (m *mockBuiltinCallContext) Halt(exitCode int, message string) error { return nil }
func (m *mockBuiltinCallContext) GetExecutionContext() registry.ExecutionContextInterface { return nil }
func (m *mockBuiltinCallContext) GetOutputBufferStack() registry.OutputBufferStackInterface { return nil }
func (m *mockBuiltinCallContext) GetCurrentFunctionArgCount() (int, error) { return 0, nil }
func (m *mockBuiltinCallContext) GetCurrentFunctionArg(index int) (*values.Value, error) { return nil, nil }
func (m *mockBuiltinCallContext) GetCurrentFunctionArgs() ([]*values.Value, error) { return nil, nil }
func (m *mockBuiltinCallContext) ThrowException(exception *values.Value) error { return fmt.Errorf("exception thrown in test mock: %v", exception) }
func (m *mockBuiltinCallContext) GetHTTPContext() registry.HTTPContext { return &mockHTTPContext{} }
func (m *mockBuiltinCallContext) ResetHTTPContext() {}
func (m *mockBuiltinCallContext) RemoveHTTPHeader(name string) {}

func TestIsCallable(t *testing.T) {
	functions := GetVariableFunctions()

	var isCallableFunc *registry.Function
	for _, f := range functions {
		if f.Name == "is_callable" {
			isCallableFunc = f
			break
		}
	}

	if isCallableFunc == nil {
		t.Fatal("is_callable function not found")
	}

	registry.Initialize()
	ctx := &mockBuiltinCallContext{registry: registry.GlobalRegistry}

	strlenFunc := &registry.Function{
		Name:       "strlen",
		ReturnType: "int",
		IsBuiltin:  true,
	}
	ctx.registry.RegisterFunction(strlenFunc)

	tests := []struct {
		name string
		args []*values.Value
		want bool
	}{
		{"function name", []*values.Value{values.NewString("strlen")}, true},
		{"non-existent function", []*values.Value{values.NewString("nonexistent")}, false},
		{"empty string", []*values.Value{values.NewString("")}, false},
		{"integer", []*values.Value{values.NewInt(123)}, false},
		{"null", []*values.Value{values.NewNull()}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := isCallableFunc.Builtin(ctx, tt.args)
			if err != nil {
				t.Fatalf("is_callable failed: %v", err)
			}
			if result.ToBool() != tt.want {
				t.Errorf("is_callable() = %v, want %v", result.ToBool(), tt.want)
			}
		})
	}
}

func TestExtensionLoaded(t *testing.T) {
	functions := GetVariableFunctions()

	var extLoadedFunc *registry.Function
	for _, f := range functions {
		if f.Name == "extension_loaded" {
			extLoadedFunc = f
			break
		}
	}

	if extLoadedFunc == nil {
		t.Fatal("extension_loaded function not found")
	}

	ctx := &mockBuiltinCallContext{}

	tests := []struct {
		name string
		args []*values.Value
		want bool
	}{
		{"json", []*values.Value{values.NewString("json")}, true},
		{"JSON uppercase", []*values.Value{values.NewString("JSON")}, true},
		{"Json mixed", []*values.Value{values.NewString("Json")}, true},
		{"mbstring", []*values.Value{values.NewString("mbstring")}, true},
		{"standard", []*values.Value{values.NewString("standard")}, true},
		{"core", []*values.Value{values.NewString("core")}, true},
		{"date", []*values.Value{values.NewString("date")}, true},
		{"pcre", []*values.Value{values.NewString("pcre")}, true},
		{"ctype", []*values.Value{values.NewString("ctype")}, true},
		{"nonexistent", []*values.Value{values.NewString("nonexistent")}, false},
		{"empty string", []*values.Value{values.NewString("")}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := extLoadedFunc.Builtin(ctx, tt.args)
			if err != nil {
				t.Fatalf("extension_loaded failed: %v", err)
			}
			if result.ToBool() != tt.want {
				t.Errorf("extension_loaded(%q) = %v, want %v", tt.args[0].ToString(), result.ToBool(), tt.want)
			}
		})
	}
}