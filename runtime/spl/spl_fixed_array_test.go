package spl

import (
	"testing"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

func TestSplFixedArray(t *testing.T) {
	// Initialize registry
	registry.Initialize()
	ctx := &mockContext{registry: registry.GlobalRegistry}

	// Get the SplFixedArray class
	class := GetSplFixedArrayClass()
	if class == nil {
		t.Fatal("SplFixedArray class is nil")
	}

	// Create a new SplFixedArray instance
	obj := &values.Object{
		ClassName:  "SplFixedArray",
		Properties: make(map[string]*values.Value),
	}
	thisObj := &values.Value{
		Type: values.TypeObject,
		Data: obj,
	}

	t.Run("Constructor", func(t *testing.T) {
		constructor := class.Methods["__construct"]
		if constructor == nil {
			t.Fatal("Constructor not found")
		}

		// Create fixed array of size 5
		args := []*values.Value{thisObj, values.NewInt(5)}
		impl := constructor.Implementation.(*BuiltinMethodImpl)
		_, err := impl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("Constructor failed: %v", err)
		}

		// Check if internal properties are set
		if obj.Properties["__size"] == nil {
			t.Fatal("Internal size not set")
		}
		if obj.Properties["__data"] == nil {
			t.Fatal("Internal data not set")
		}
	})

	t.Run("GetSize", func(t *testing.T) {
		getSizeMethod := class.Methods["getSize"]
		if getSizeMethod == nil {
			t.Fatal("getSize method not found")
		}

		args := []*values.Value{thisObj}
		impl := getSizeMethod.Implementation.(*BuiltinMethodImpl)
		result, err := impl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("getSize failed: %v", err)
		}

		if result.ToInt() != 5 {
			t.Fatalf("Expected size 5, got %d", result.ToInt())
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

		if result.ToInt() != 5 {
			t.Fatalf("Expected count 5, got %d", result.ToInt())
		}
	})

	t.Run("ArrayAccess", func(t *testing.T) {
		// Test offsetSet
		offsetSetMethod := class.Methods["offsetSet"]
		offsetSetImpl := offsetSetMethod.Implementation.(*BuiltinMethodImpl)

		args := []*values.Value{thisObj, values.NewInt(0), values.NewString("zero")}
		_, err := offsetSetImpl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("offsetSet failed: %v", err)
		}

		args = []*values.Value{thisObj, values.NewInt(1), values.NewString("one")}
		_, _ = offsetSetImpl.GetFunction().Builtin(ctx, args)

		args = []*values.Value{thisObj, values.NewInt(4), values.NewString("four")}
		_, _ = offsetSetImpl.GetFunction().Builtin(ctx, args)

		// Test offsetGet
		offsetGetMethod := class.Methods["offsetGet"]
		offsetGetImpl := offsetGetMethod.Implementation.(*BuiltinMethodImpl)

		args = []*values.Value{thisObj, values.NewInt(0)}
		result, err := offsetGetImpl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("offsetGet failed: %v", err)
		}

		if result.ToString() != "zero" {
			t.Fatalf("Expected 'zero', got '%s'", result.ToString())
		}

		// Test offsetExists
		offsetExistsMethod := class.Methods["offsetExists"]
		offsetExistsImpl := offsetExistsMethod.Implementation.(*BuiltinMethodImpl)

		args = []*values.Value{thisObj, values.NewInt(1)}
		result, _ = offsetExistsImpl.GetFunction().Builtin(ctx, args)
		if !result.ToBool() {
			t.Fatal("Expected offsetExists to return true for index 1")
		}

		// Test out of bounds
		args = []*values.Value{thisObj, values.NewInt(10)}
		result, _ = offsetExistsImpl.GetFunction().Builtin(ctx, args)
		if result.ToBool() {
			t.Fatal("Expected offsetExists to return false for out of bounds index")
		}

		// Test offsetUnset
		offsetUnsetMethod := class.Methods["offsetUnset"]
		offsetUnsetImpl := offsetUnsetMethod.Implementation.(*BuiltinMethodImpl)

		args = []*values.Value{thisObj, values.NewInt(1)}
		_, _ = offsetUnsetImpl.GetFunction().Builtin(ctx, args)

		// Check if unset worked
		args = []*values.Value{thisObj, values.NewInt(1)}
		result, _ = offsetGetImpl.GetFunction().Builtin(ctx, args)
		if !result.IsNull() {
			t.Fatal("Expected null after unset")
		}
	})

	t.Run("SetSize", func(t *testing.T) {
		setSizeMethod := class.Methods["setSize"]
		if setSizeMethod == nil {
			t.Fatal("setSize method not found")
		}

		// Resize to 3
		args := []*values.Value{thisObj, values.NewInt(3)}
		impl := setSizeMethod.Implementation.(*BuiltinMethodImpl)
		_, err := impl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("setSize failed: %v", err)
		}

		// Check new size
		getSizeMethod := class.Methods["getSize"]
		getSizeImpl := getSizeMethod.Implementation.(*BuiltinMethodImpl)
		args = []*values.Value{thisObj}
		result, _ := getSizeImpl.GetFunction().Builtin(ctx, args)
		if result.ToInt() != 3 {
			t.Fatalf("Expected size 3 after resize, got %d", result.ToInt())
		}

		// Check that data was preserved for valid indices
		offsetGetMethod := class.Methods["offsetGet"]
		offsetGetImpl := offsetGetMethod.Implementation.(*BuiltinMethodImpl)

		args = []*values.Value{thisObj, values.NewInt(0)}
		result, _ = offsetGetImpl.GetFunction().Builtin(ctx, args)
		if result.ToString() != "zero" {
			t.Fatalf("Expected 'zero' after resize, got '%s'", result.ToString())
		}

		// Check that out of bounds index is no longer accessible
		offsetExistsMethod := class.Methods["offsetExists"]
		offsetExistsImpl := offsetExistsMethod.Implementation.(*BuiltinMethodImpl)

		args = []*values.Value{thisObj, values.NewInt(4)}
		result, _ = offsetExistsImpl.GetFunction().Builtin(ctx, args)
		if result.ToBool() {
			t.Fatal("Expected offsetExists to return false for index 4 after resize to 3")
		}
	})

	t.Run("ToArray", func(t *testing.T) {
		toArrayMethod := class.Methods["toArray"]
		if toArrayMethod == nil {
			t.Fatal("toArray method not found")
		}

		args := []*values.Value{thisObj}
		impl := toArrayMethod.Implementation.(*BuiltinMethodImpl)
		result, err := impl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("toArray failed: %v", err)
		}

		if !result.IsArray() {
			t.Fatal("Expected toArray to return an array")
		}

		// Check array contents
		if result.ArrayCount() != 3 {
			t.Fatalf("Expected array count 3, got %d", result.ArrayCount())
		}

		val := result.ArrayGet(values.NewInt(0))
		if val.ToString() != "zero" {
			t.Fatalf("Expected 'zero' at index 0, got '%s'", val.ToString())
		}
	})
}