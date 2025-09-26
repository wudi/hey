package spl

import (
	"fmt"
	"testing"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// MockContext for testing
type mockContext struct {
	registry *registry.Registry
}

func (m *mockContext) WriteOutput(val *values.Value) error                      { return nil }
func (m *mockContext) GetGlobal(name string) (*values.Value, bool)              { return nil, false }
func (m *mockContext) SetGlobal(name string, val *values.Value)                 {}
func (m *mockContext) SymbolRegistry() *registry.Registry                       { return m.registry }
func (m *mockContext) LookupUserFunction(name string) (*registry.Function, bool) { return nil, false }
func (m *mockContext) CallUserFunction(function *registry.Function, args []*values.Value) (*values.Value, error) {
	return nil, nil
}
func (m *mockContext) SimpleCallUserFunction(function *registry.Function, args []*values.Value) (*values.Value, error) {
	return nil, nil
}
func (m *mockContext) LookupUserClass(name string) (*registry.Class, bool)      { return nil, false }
func (m *mockContext) Halt(exitCode int, message string) error                  { return nil }
func (m *mockContext) GetExecutionContext() registry.ExecutionContextInterface  { return nil }
func (m *mockContext) GetOutputBufferStack() registry.OutputBufferStackInterface { return nil }
func (m *mockContext) GetCurrentFunctionArgCount() (int, error)                 { return 0, nil }
func (m *mockContext) GetCurrentFunctionArg(index int) (*values.Value, error)   { return nil, nil }
func (m *mockContext) GetCurrentFunctionArgs() ([]*values.Value, error)         { return nil, nil }
func (m *mockContext) ThrowException(exception *values.Value) error { return fmt.Errorf("exception thrown in test mock: %v", exception) }

func TestArrayIterator(t *testing.T) {
	// Initialize registry
	registry.Initialize()
	ctx := &mockContext{registry: registry.GlobalRegistry}

	// Get the ArrayIterator class
	class := GetArrayIteratorClass()
	if class == nil {
		t.Fatal("ArrayIterator class is nil")
	}

	// Create a new ArrayIterator instance
	obj := &values.Object{
		ClassName:  "ArrayIterator",
		Properties: make(map[string]*values.Value),
	}
	thisObj := &values.Value{
		Type: values.TypeObject,
		Data: obj,
	}

	// Test data
	testArray := values.NewArray()
	testArray.ArraySet(values.NewString("a"), values.NewString("apple"))
	testArray.ArraySet(values.NewString("b"), values.NewString("banana"))
	testArray.ArraySet(values.NewString("c"), values.NewString("cherry"))

	t.Run("Constructor", func(t *testing.T) {
		constructor := class.Methods["__construct"]
		if constructor == nil {
			t.Fatal("Constructor not found")
		}

		args := []*values.Value{thisObj, testArray}
		impl := constructor.Implementation.(*BuiltinMethodImpl)
		_, err := impl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("Constructor failed: %v", err)
		}

		// Check if internal array is set
		if obj.Properties["__array"] == nil {
			t.Fatal("Internal array not set")
		}
	})

	t.Run("Count", func(t *testing.T) {
		countMethod := class.Methods["count"]
		if countMethod == nil {
			t.Fatal("count method not found")
		}

		args := []*values.Value{thisObj}
		impl := countMethod.Implementation.(*BuiltinMethodImpl)
		result, err := impl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("count failed: %v", err)
		}

		if result.ToInt() != 3 {
			t.Fatalf("Expected count 3, got %d", result.ToInt())
		}
	})

	t.Run("Current", func(t *testing.T) {
		currentMethod := class.Methods["current"]
		if currentMethod == nil {
			t.Fatal("current method not found")
		}

		// First rewind
		rewindMethod := class.Methods["rewind"]
		args := []*values.Value{thisObj}
		rewindImpl := rewindMethod.Implementation.(*BuiltinMethodImpl)
		_, _ = rewindImpl.GetFunction().Builtin(ctx, args)

		currentImpl := currentMethod.Implementation.(*BuiltinMethodImpl)
		result, err := currentImpl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("current failed: %v", err)
		}

		if result.ToString() != "apple" {
			t.Fatalf("Expected 'apple', got '%s'", result.ToString())
		}
	})

	t.Run("Key", func(t *testing.T) {
		// First rewind to ensure clean state
		rewindMethod := class.Methods["rewind"]
		args := []*values.Value{thisObj}
		rewindImpl := rewindMethod.Implementation.(*BuiltinMethodImpl)
		_, _ = rewindImpl.GetFunction().Builtin(ctx, args)

		keyMethod := class.Methods["key"]
		if keyMethod == nil {
			t.Fatal("key method not found")
		}

		keyImpl := keyMethod.Implementation.(*BuiltinMethodImpl)
		result, err := keyImpl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("key failed: %v", err)
		}

		if result.ToString() != "a" {
			t.Fatalf("Expected key 'a', got '%s'", result.ToString())
		}
	})

	t.Run("Next", func(t *testing.T) {
		// First rewind to ensure clean state
		rewindMethod := class.Methods["rewind"]
		args := []*values.Value{thisObj}
		rewindImpl := rewindMethod.Implementation.(*BuiltinMethodImpl)
		_, _ = rewindImpl.GetFunction().Builtin(ctx, args)

		nextMethod := class.Methods["next"]
		currentMethod := class.Methods["current"]
		nextImpl := nextMethod.Implementation.(*BuiltinMethodImpl)
		_, err := nextImpl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("next failed: %v", err)
		}

		currentImpl := currentMethod.Implementation.(*BuiltinMethodImpl)
		result, _ := currentImpl.GetFunction().Builtin(ctx, args)
		if result.ToString() != "banana" {
			t.Fatalf("Expected 'banana' after next, got '%s'", result.ToString())
		}
	})

	t.Run("Valid", func(t *testing.T) {
		validMethod := class.Methods["valid"]
		if validMethod == nil {
			t.Fatal("valid method not found")
		}

		args := []*values.Value{thisObj}
		validImpl := validMethod.Implementation.(*BuiltinMethodImpl)
		result, err := validImpl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("valid failed: %v", err)
		}

		if !result.ToBool() {
			t.Fatal("Expected valid() to return true")
		}

		// Move to end
		nextMethod := class.Methods["next"]
		nextImpl := nextMethod.Implementation.(*BuiltinMethodImpl)
		nextImpl.GetFunction().Builtin(ctx, args) // Move to 'c'
		nextImpl.GetFunction().Builtin(ctx, args) // Move past end

		result, _ = validImpl.GetFunction().Builtin(ctx, args)
		if result.ToBool() {
			t.Fatal("Expected valid() to return false at end")
		}
	})

	t.Run("ArrayAccess", func(t *testing.T) {
		// Test offsetGet
		offsetGetMethod := class.Methods["offsetGet"]
		if offsetGetMethod == nil {
			t.Fatal("offsetGet method not found")
		}

		args := []*values.Value{thisObj, values.NewString("b")}
		offsetGetImpl := offsetGetMethod.Implementation.(*BuiltinMethodImpl)
		result, err := offsetGetImpl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("offsetGet failed: %v", err)
		}

		if result.ToString() != "banana" {
			t.Fatalf("Expected 'banana', got '%s'", result.ToString())
		}

		// Test offsetSet
		offsetSetMethod := class.Methods["offsetSet"]
		offsetSetImpl := offsetSetMethod.Implementation.(*BuiltinMethodImpl)
		args = []*values.Value{thisObj, values.NewString("d"), values.NewString("date")}
		_, err = offsetSetImpl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("offsetSet failed: %v", err)
		}

		// Verify set worked
		args = []*values.Value{thisObj, values.NewString("d")}
		result, _ = offsetGetImpl.GetFunction().Builtin(ctx, args)
		if result.ToString() != "date" {
			t.Fatalf("Expected 'date' after set, got '%s'", result.ToString())
		}

		// Test offsetExists
		offsetExistsMethod := class.Methods["offsetExists"]
		offsetExistsImpl := offsetExistsMethod.Implementation.(*BuiltinMethodImpl)
		args = []*values.Value{thisObj, values.NewString("d")}
		result, _ = offsetExistsImpl.GetFunction().Builtin(ctx, args)
		if !result.ToBool() {
			t.Fatal("Expected offsetExists to return true for 'd'")
		}

		// Test offsetUnset
		offsetUnsetMethod := class.Methods["offsetUnset"]
		offsetUnsetImpl := offsetUnsetMethod.Implementation.(*BuiltinMethodImpl)
		args = []*values.Value{thisObj, values.NewString("d")}
		_, _ = offsetUnsetImpl.GetFunction().Builtin(ctx, args)

		// Verify unset worked
		args = []*values.Value{thisObj, values.NewString("d")}
		result, _ = offsetExistsImpl.GetFunction().Builtin(ctx, args)
		if result.ToBool() {
			t.Fatal("Expected offsetExists to return false after unset")
		}
	})
}