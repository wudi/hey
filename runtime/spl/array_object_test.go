package spl

import (
	"testing"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

func TestArrayObject(t *testing.T) {
	// Initialize registry
	registry.Initialize()
	ctx := &mockContext{registry: registry.GlobalRegistry}

	// Get the ArrayObject class
	class := GetArrayObjectClass()
	if class == nil {
		t.Fatal("ArrayObject class is nil")
	}

	// Create a new ArrayObject instance
	obj := &values.Object{
		ClassName:  "ArrayObject",
		Properties: make(map[string]*values.Value),
	}
	thisObj := &values.Value{
		Type: values.TypeObject,
		Data: obj,
	}

	// Test data
	testArray := values.NewArray()
	testArray.ArraySet(values.NewString("x"), values.NewInt(10))
	testArray.ArraySet(values.NewString("y"), values.NewInt(20))

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

		if result.ToInt() != 2 {
			t.Fatalf("Expected count 2, got %d", result.ToInt())
		}
	})

	t.Run("ArrayAccess", func(t *testing.T) {
		// Test offsetGet
		offsetGetMethod := class.Methods["offsetGet"]
		if offsetGetMethod == nil {
			t.Fatal("offsetGet method not found")
		}

		args := []*values.Value{thisObj, values.NewString("y")}
		offsetGetImpl := offsetGetMethod.Implementation.(*BuiltinMethodImpl)
		result, err := offsetGetImpl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("offsetGet failed: %v", err)
		}

		if result.ToInt() != 20 {
			t.Fatalf("Expected 20, got %d", result.ToInt())
		}

		// Test offsetSet
		offsetSetMethod := class.Methods["offsetSet"]
		offsetSetImpl := offsetSetMethod.Implementation.(*BuiltinMethodImpl)
		args = []*values.Value{thisObj, values.NewString("z"), values.NewInt(30)}
		_, err = offsetSetImpl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("offsetSet failed: %v", err)
		}

		// Verify set worked
		args = []*values.Value{thisObj, values.NewString("z")}
		result, _ = offsetGetImpl.GetFunction().Builtin(ctx, args)
		if result.ToInt() != 30 {
			t.Fatalf("Expected 30 after set, got %d", result.ToInt())
		}

		// Test offsetExists
		offsetExistsMethod := class.Methods["offsetExists"]
		offsetExistsImpl := offsetExistsMethod.Implementation.(*BuiltinMethodImpl)
		args = []*values.Value{thisObj, values.NewString("z")}
		result, _ = offsetExistsImpl.GetFunction().Builtin(ctx, args)
		if !result.ToBool() {
			t.Fatal("Expected offsetExists to return true for 'z'")
		}

		// Test offsetUnset
		offsetUnsetMethod := class.Methods["offsetUnset"]
		offsetUnsetImpl := offsetUnsetMethod.Implementation.(*BuiltinMethodImpl)
		args = []*values.Value{thisObj, values.NewString("z")}
		_, _ = offsetUnsetImpl.GetFunction().Builtin(ctx, args)

		// Verify unset worked
		args = []*values.Value{thisObj, values.NewString("z")}
		result, _ = offsetExistsImpl.GetFunction().Builtin(ctx, args)
		if result.ToBool() {
			t.Fatal("Expected offsetExists to return false after unset")
		}
	})

	t.Run("Iterator", func(t *testing.T) {
		// Test getIterator
		getIteratorMethod := class.Methods["getIterator"]
		if getIteratorMethod == nil {
			t.Fatal("getIterator method not found")
		}

		args := []*values.Value{thisObj}
		impl := getIteratorMethod.Implementation.(*BuiltinMethodImpl)
		result, err := impl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("getIterator failed: %v", err)
		}

		// Should return an ArrayIterator
		if !result.IsObject() {
			t.Fatal("Expected getIterator to return an object")
		}

		iterObj := result.Data.(*values.Object)
		if iterObj.ClassName != "ArrayIterator" {
			t.Fatalf("Expected ArrayIterator, got %s", iterObj.ClassName)
		}
	})

	t.Run("GetArrayCopy", func(t *testing.T) {
		getArrayCopyMethod := class.Methods["getArrayCopy"]
		if getArrayCopyMethod == nil {
			t.Fatal("getArrayCopy method not found")
		}

		args := []*values.Value{thisObj}
		impl := getArrayCopyMethod.Implementation.(*BuiltinMethodImpl)
		result, err := impl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("getArrayCopy failed: %v", err)
		}

		if !result.IsArray() {
			t.Fatal("Expected getArrayCopy to return an array")
		}

		// Should have same values
		if result.ArrayCount() != 2 {
			t.Fatalf("Expected array count 2, got %d", result.ArrayCount())
		}

		val := result.ArrayGet(values.NewString("x"))
		if val.ToInt() != 10 {
			t.Fatalf("Expected x=10, got %d", val.ToInt())
		}
	})

	t.Run("ExchangeArray", func(t *testing.T) {
		exchangeArrayMethod := class.Methods["exchangeArray"]
		if exchangeArrayMethod == nil {
			t.Fatal("exchangeArray method not found")
		}

		newArray := values.NewArray()
		newArray.ArraySet(values.NewString("a"), values.NewString("apple"))
		newArray.ArraySet(values.NewString("b"), values.NewString("banana"))

		args := []*values.Value{thisObj, newArray}
		impl := exchangeArrayMethod.Implementation.(*BuiltinMethodImpl)
		result, err := impl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("exchangeArray failed: %v", err)
		}

		// Should return old array
		if !result.IsArray() {
			t.Fatal("Expected exchangeArray to return an array")
		}

		// New array should be set
		countMethod := class.Methods["count"]
		countImpl := countMethod.Implementation.(*BuiltinMethodImpl)
		args = []*values.Value{thisObj}
		countResult, _ := countImpl.GetFunction().Builtin(ctx, args)
		if countResult.ToInt() != 2 {
			t.Fatalf("Expected count 2 after exchange, got %d", countResult.ToInt())
		}

		// Test new values
		offsetGetMethod := class.Methods["offsetGet"]
		offsetGetImpl := offsetGetMethod.Implementation.(*BuiltinMethodImpl)
		args = []*values.Value{thisObj, values.NewString("a")}
		val, _ := offsetGetImpl.GetFunction().Builtin(ctx, args)
		if val.ToString() != "apple" {
			t.Fatalf("Expected 'apple', got '%s'", val.ToString())
		}
	})
}