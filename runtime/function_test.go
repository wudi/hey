package runtime

import (
	"strings"
	"testing"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// Use existing mockBuiltinContext from array_test.go
// Extended mockBuiltinContext with function-specific fields
type functionTestContext struct {
	mockBuiltinContext
	userFunctions map[string]*registry.Function
}

func (f *functionTestContext) LookupUserFunction(name string) (*registry.Function, bool) {
	if f.userFunctions == nil {
		return nil, false
	}
	fn, ok := f.userFunctions[strings.ToLower(name)]
	return fn, ok
}

func TestFunctionExists(t *testing.T) {
	functions := GetFunctionFunctions()
	var functionExistsFn *registry.Function

	for _, fn := range functions {
		if fn.Name == "function_exists" {
			functionExistsFn = fn
			break
		}
	}

	if functionExistsFn == nil {
		t.Fatal("function_exists function not found")
	}

	ctx := &functionTestContext{}

	// Test with empty/null function name
	result, err := functionExistsFn.Builtin(ctx, []*values.Value{values.NewNull()})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result.Type != values.TypeBool || result.Data.(bool) != false {
		t.Errorf("Expected false for null function name, got %v", result)
	}

	// Test with empty string
	result, err = functionExistsFn.Builtin(ctx, []*values.Value{values.NewString("")})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result.Type != values.TypeBool || result.Data.(bool) != false {
		t.Errorf("Expected false for empty function name, got %v", result)
	}

	// Test with non-existent function
	result, err = functionExistsFn.Builtin(ctx, []*values.Value{values.NewString("nonexistent_function")})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result.Type != values.TypeBool || result.Data.(bool) != false {
		t.Errorf("Expected false for non-existent function, got %v", result)
	}

	// Test with user-defined function
	ctx.userFunctions = map[string]*registry.Function{
		"test_function": {Name: "test_function", IsBuiltin: false},
	}
	result, err = functionExistsFn.Builtin(ctx, []*values.Value{values.NewString("test_function")})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result.Type != values.TypeBool || result.Data.(bool) != true {
		t.Errorf("Expected true for existing user function, got %v", result)
	}
}

func TestCallUserFunc(t *testing.T) {
	functions := GetFunctionFunctions()
	var callUserFuncFn *registry.Function

	for _, fn := range functions {
		if fn.Name == "call_user_func" {
			callUserFuncFn = fn
			break
		}
	}

	if callUserFuncFn == nil {
		t.Fatal("call_user_func function not found")
	}

	ctx := &functionTestContext{}

	// Test with no arguments
	_, err := callUserFuncFn.Builtin(ctx, []*values.Value{})
	if err == nil {
		t.Error("Expected error for no arguments")
	}
	if !strings.Contains(err.Error(), "expects at least 1 parameter") {
		t.Errorf("Expected 'expects at least 1 parameter' error, got %v", err)
	}

	// Test with invalid callback type
	_, err = callUserFuncFn.Builtin(ctx, []*values.Value{values.NewInt(123)})
	if err == nil {
		t.Error("Expected error for invalid callback type")
	}
	if !strings.Contains(err.Error(), "must be a valid callback") {
		t.Errorf("Expected 'must be a valid callback' error, got %v", err)
	}

	// Test with non-existent function
	_, err = callUserFuncFn.Builtin(ctx, []*values.Value{values.NewString("nonexistent_func")})
	if err == nil {
		t.Error("Expected error for non-existent function")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected 'not found' error, got %v", err)
	}
}

func TestCallUserFuncArray(t *testing.T) {
	functions := GetFunctionFunctions()
	var callUserFuncArrayFn *registry.Function

	for _, fn := range functions {
		if fn.Name == "call_user_func_array" {
			callUserFuncArrayFn = fn
			break
		}
	}

	if callUserFuncArrayFn == nil {
		t.Fatal("call_user_func_array function not found")
	}

	ctx := &functionTestContext{}

	// Test with wrong number of arguments
	_, err := callUserFuncArrayFn.Builtin(ctx, []*values.Value{values.NewString("test")})
	if err == nil {
		t.Error("Expected error for wrong number of arguments")
	}
	if !strings.Contains(err.Error(), "expects exactly 2 parameters") {
		t.Errorf("Expected 'expects exactly 2 parameters' error, got %v", err)
	}

	// Test with valid arguments structure (even if function doesn't exist)
	argsArray := values.NewArray()
	_, err = callUserFuncArrayFn.Builtin(ctx, []*values.Value{
		values.NewString("nonexistent_func"),
		argsArray,
	})
	if err == nil {
		t.Error("Expected error for non-existent function")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected 'not found' error, got %v", err)
	}
}

func TestGetDefinedFunctions(t *testing.T) {
	functions := GetFunctionFunctions()
	var getDefinedFunctionsFn *registry.Function

	for _, fn := range functions {
		if fn.Name == "get_defined_functions" {
			getDefinedFunctionsFn = fn
			break
		}
	}

	if getDefinedFunctionsFn == nil {
		t.Fatal("get_defined_functions function not found")
	}

	ctx := &functionTestContext{}

	result, err := getDefinedFunctionsFn.Builtin(ctx, []*values.Value{})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result.Type != values.TypeArray {
		t.Errorf("Expected array result, got %v", result.Type)
	}

	// Verify structure: should have 'internal' and 'user' keys
	resultData := result.Data.(*values.Array)
	if _, hasInternal := resultData.Elements["internal"]; !hasInternal {
		t.Error("Expected 'internal' key in result")
	}
	if _, hasUser := resultData.Elements["user"]; !hasUser {
		t.Error("Expected 'user' key in result")
	}
}

func TestArgIntrospectionFunctions(t *testing.T) {
	functions := GetFunctionFunctions()

	// Test func_num_args
	var funcNumArgsFn *registry.Function
	for _, fn := range functions {
		if fn.Name == "func_num_args" {
			funcNumArgsFn = fn
			break
		}
	}

	if funcNumArgsFn == nil {
		t.Fatal("func_num_args function not found")
	}

	ctx := &functionTestContext{}

	_, err := funcNumArgsFn.Builtin(ctx, []*values.Value{})
	if err == nil {
		t.Error("Expected error when called from global scope")
	}
	if !strings.Contains(err.Error(), "cannot be called from the global scope") {
		t.Errorf("Expected global scope error, got %v", err)
	}

	// Test func_get_arg
	var funcGetArgFn *registry.Function
	for _, fn := range functions {
		if fn.Name == "func_get_arg" {
			funcGetArgFn = fn
			break
		}
	}

	if funcGetArgFn == nil {
		t.Fatal("func_get_arg function not found")
	}

	_, err = funcGetArgFn.Builtin(ctx, []*values.Value{values.NewInt(0)})
	if err == nil {
		t.Error("Expected error when called from global scope")
	}
	if !strings.Contains(err.Error(), "cannot be called from the global scope") {
		t.Errorf("Expected global scope error, got %v", err)
	}

	// Test func_get_args
	var funcGetArgsFn *registry.Function
	for _, fn := range functions {
		if fn.Name == "func_get_args" {
			funcGetArgsFn = fn
			break
		}
	}

	if funcGetArgsFn == nil {
		t.Fatal("func_get_args function not found")
	}

	_, err = funcGetArgsFn.Builtin(ctx, []*values.Value{})
	if err == nil {
		t.Error("Expected error when called from global scope")
	}
	if !strings.Contains(err.Error(), "cannot be called from the global scope") {
		t.Errorf("Expected global scope error, got %v", err)
	}
}

func TestCreateFunction(t *testing.T) {
	functions := GetFunctionFunctions()
	var createFunctionFn *registry.Function

	for _, fn := range functions {
		if fn.Name == "create_function" {
			createFunctionFn = fn
			break
		}
	}

	if createFunctionFn == nil {
		t.Fatal("create_function function not found")
	}

	ctx := &functionTestContext{}

	_, err := createFunctionFn.Builtin(ctx, []*values.Value{
		values.NewString("$a,$b"),
		values.NewString("return $a + $b;"),
	})
	if err == nil {
		t.Error("Expected error for deprecated create_function")
	}
	if !strings.Contains(err.Error(), "deprecated") {
		t.Errorf("Expected deprecation error, got %v", err)
	}
}

func TestShutdownAndTickFunctions(t *testing.T) {
	functions := GetFunctionFunctions()

	// Test register_shutdown_function (should return true for now)
	var registerShutdownFn *registry.Function
	for _, fn := range functions {
		if fn.Name == "register_shutdown_function" {
			registerShutdownFn = fn
			break
		}
	}

	if registerShutdownFn == nil {
		t.Fatal("register_shutdown_function function not found")
	}

	ctx := &functionTestContext{}

	result, err := registerShutdownFn.Builtin(ctx, []*values.Value{values.NewString("test_function")})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result.Type != values.TypeBool || result.Data.(bool) != true {
		t.Errorf("Expected true result, got %v", result)
	}

	// Test register_tick_function (should return true for now)
	var registerTickFn *registry.Function
	for _, fn := range functions {
		if fn.Name == "register_tick_function" {
			registerTickFn = fn
			break
		}
	}

	if registerTickFn == nil {
		t.Fatal("register_tick_function function not found")
	}

	result, err = registerTickFn.Builtin(ctx, []*values.Value{values.NewString("test_function")})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result.Type != values.TypeBool || result.Data.(bool) != true {
		t.Errorf("Expected true result, got %v", result)
	}

	// Test unregister_tick_function (should return null)
	var unregisterTickFn *registry.Function
	for _, fn := range functions {
		if fn.Name == "unregister_tick_function" {
			unregisterTickFn = fn
			break
		}
	}

	if unregisterTickFn == nil {
		t.Fatal("unregister_tick_function function not found")
	}

	result, err = unregisterTickFn.Builtin(ctx, []*values.Value{values.NewString("test_function")})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result.Type != values.TypeNull {
		t.Errorf("Expected null result, got %v", result)
	}
}