package spl

import (
	"testing"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

func TestLimitIterator(t *testing.T) {
	// Initialize registry
	registry.Initialize()
	ctx := &mockContext{registry: registry.GlobalRegistry}

	// Register all SPL classes for testing
	for _, class := range GetSplClasses() {
		if err := registry.GlobalRegistry.RegisterClass(class); err != nil {
			t.Fatalf("Failed to register class %s: %v", class.Name, err)
		}
	}

	// Get the LimitIterator class
	class := GetLimitIteratorClass()
	if class == nil {
		t.Fatal("LimitIterator class is nil")
	}

	// Create an ArrayIterator to wrap
	arrayIterClass := GetArrayIteratorClass()
	arrayIterObj := &values.Object{
		ClassName:  "ArrayIterator",
		Properties: make(map[string]*values.Value),
	}
	arrayIterValue := &values.Value{
		Type: values.TypeObject,
		Data: arrayIterObj,
	}

	// Initialize ArrayIterator with test data
	testArray := values.NewArray()
	testArray.ArraySet(values.NewString("a"), values.NewInt(1))
	testArray.ArraySet(values.NewString("b"), values.NewInt(2))
	testArray.ArraySet(values.NewString("c"), values.NewInt(3))
	testArray.ArraySet(values.NewString("d"), values.NewInt(4))
	testArray.ArraySet(values.NewString("e"), values.NewInt(5))

	constructorArgs := []*values.Value{arrayIterValue, testArray}
	arrayIterConstructor := arrayIterClass.Methods["__construct"]
	impl := arrayIterConstructor.Implementation.(*BuiltinMethodImpl)
	_, err := impl.GetFunction().Builtin(ctx, constructorArgs)
	if err != nil {
		t.Fatalf("Failed to initialize ArrayIterator: %v", err)
	}

	// Rewind the ArrayIterator
	rewindMethod := arrayIterClass.Methods["rewind"]
	rewindImpl := rewindMethod.Implementation.(*BuiltinMethodImpl)
	_, err = rewindImpl.GetFunction().Builtin(ctx, []*values.Value{arrayIterValue})
	if err != nil {
		t.Fatalf("Failed to rewind ArrayIterator: %v", err)
	}

	t.Run("ConstructorBasicLimit", func(t *testing.T) {
		// Create LimitIterator with offset=0, limit=2
		obj := &values.Object{
			ClassName:  "LimitIterator",
			Properties: make(map[string]*values.Value),
		}
		thisObj := &values.Value{
			Type: values.TypeObject,
			Data: obj,
		}

		constructor := class.Methods["__construct"]
		if constructor == nil {
			t.Fatal("Constructor not found")
		}

		args := []*values.Value{thisObj, arrayIterValue, values.NewInt(0), values.NewInt(2)}
		impl := constructor.Implementation.(*BuiltinMethodImpl)
		_, err := impl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("Constructor failed: %v", err)
		}

		// Check that properties were stored
		if iterator, ok := obj.Properties["__iterator"]; !ok || iterator != arrayIterValue {
			t.Fatal("Iterator not stored correctly")
		}
		if offset, ok := obj.Properties["__offset"]; !ok || !offset.IsInt() || offset.ToInt() != 0 {
			t.Fatal("Offset not stored correctly")
		}
		if limit, ok := obj.Properties["__limit"]; !ok || !limit.IsInt() || limit.ToInt() != 2 {
			t.Fatal("Limit not stored correctly")
		}
	})

	// Use a fresh LimitIterator for remaining tests
	obj := &values.Object{
		ClassName:  "LimitIterator",
		Properties: make(map[string]*values.Value),
	}
	thisObj := &values.Value{
		Type: values.TypeObject,
		Data: obj,
	}

	// Initialize with offset=1, limit=2
	constructor := class.Methods["__construct"]
	args := []*values.Value{thisObj, arrayIterValue, values.NewInt(1), values.NewInt(2)}
	impl = constructor.Implementation.(*BuiltinMethodImpl)
	_, err = impl.GetFunction().Builtin(ctx, args)
	if err != nil {
		t.Fatalf("Constructor failed: %v", err)
	}

	t.Run("getInnerIterator", func(t *testing.T) {
		method := class.Methods["getInnerIterator"]
		if method == nil {
			t.Fatal("getInnerIterator method not found")
		}

		args := []*values.Value{thisObj}
		impl := method.Implementation.(*BuiltinMethodImpl)
		result, err := impl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("getInnerIterator failed: %v", err)
		}

		if result != arrayIterValue {
			t.Fatal("getInnerIterator should return the wrapped iterator")
		}
	})

	t.Run("rewindAndIteration", func(t *testing.T) {
		// Rewind the LimitIterator
		rewindMethod := class.Methods["rewind"]
		rewindImpl := rewindMethod.Implementation.(*BuiltinMethodImpl)
		_, err := rewindImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("rewind failed: %v", err)
		}

		// Should be positioned at offset=1 (second element)
		currentMethod := class.Methods["current"]
		currentImpl := currentMethod.Implementation.(*BuiltinMethodImpl)
		result, err := currentImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("current failed: %v", err)
		}

		if !result.IsInt() || result.ToInt() != 2 {
			t.Fatalf("Expected current value to be 2, got: %v", result)
		}

		// Check key
		keyMethod := class.Methods["key"]
		keyImpl := keyMethod.Implementation.(*BuiltinMethodImpl)
		keyResult, err := keyImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("key failed: %v", err)
		}

		if !keyResult.IsString() || keyResult.ToString() != "b" {
			t.Fatalf("Expected key to be 'b', got: %v", keyResult.ToString())
		}
	})

	t.Run("valid", func(t *testing.T) {
		method := class.Methods["valid"]
		if method == nil {
			t.Fatal("valid method not found")
		}

		args := []*values.Value{thisObj}
		impl := method.Implementation.(*BuiltinMethodImpl)
		result, err := impl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("valid failed: %v", err)
		}

		if !result.IsBool() || !result.ToBool() {
			t.Fatal("Expected valid() to return true")
		}
	})

	t.Run("getPosition", func(t *testing.T) {
		method := class.Methods["getPosition"]
		if method == nil {
			t.Fatal("getPosition method not found")
		}

		args := []*values.Value{thisObj}
		impl := method.Implementation.(*BuiltinMethodImpl)
		result, err := impl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("getPosition failed: %v", err)
		}

		if !result.IsInt() || result.ToInt() != 0 {
			t.Fatalf("Expected position to be 0, got: %v", result.ToInt())
		}
	})

	t.Run("next", func(t *testing.T) {
		method := class.Methods["next"]
		if method == nil {
			t.Fatal("next method not found")
		}

		args := []*values.Value{thisObj}
		impl := method.Implementation.(*BuiltinMethodImpl)
		_, err := impl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("next failed: %v", err)
		}

		// After next(), current should be the third element (value 3)
		currentMethod := class.Methods["current"]
		currentImpl := currentMethod.Implementation.(*BuiltinMethodImpl)
		result, err := currentImpl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("current after next failed: %v", err)
		}

		if !result.IsInt() || result.ToInt() != 3 {
			t.Fatalf("Expected current value after next to be 3, got: %v", result)
		}

		// Position should be 1
		posMethod := class.Methods["getPosition"]
		posImpl := posMethod.Implementation.(*BuiltinMethodImpl)
		posResult, err := posImpl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("getPosition after next failed: %v", err)
		}

		if !posResult.IsInt() || posResult.ToInt() != 1 {
			t.Fatalf("Expected position after next to be 1, got: %v", posResult.ToInt())
		}
	})

	t.Run("limitReached", func(t *testing.T) {
		// Move to next position (should be at limit now)
		nextMethod := class.Methods["next"]
		nextImpl := nextMethod.Implementation.(*BuiltinMethodImpl)
		_, err := nextImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("next failed: %v", err)
		}

		// Should now be invalid (limit reached)
		validMethod := class.Methods["valid"]
		validImpl := validMethod.Implementation.(*BuiltinMethodImpl)
		result, err := validImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("valid failed: %v", err)
		}

		if !result.IsBool() || result.ToBool() {
			t.Fatal("Expected valid() to return false after limit reached")
		}
	})

	t.Run("seek", func(t *testing.T) {
		// Create fresh LimitIterator for seek test
		seekObj := &values.Object{
			ClassName:  "LimitIterator",
			Properties: make(map[string]*values.Value),
		}
		seekThisObj := &values.Value{
			Type: values.TypeObject,
			Data: seekObj,
		}

		// Initialize with offset=0, limit=3
		constructor := class.Methods["__construct"]
		args := []*values.Value{seekThisObj, arrayIterValue, values.NewInt(0), values.NewInt(3)}
		impl := constructor.Implementation.(*BuiltinMethodImpl)
		_, err := impl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("Constructor failed: %v", err)
		}

		// Rewind
		rewindMethod := class.Methods["rewind"]
		rewindImpl := rewindMethod.Implementation.(*BuiltinMethodImpl)
		_, err = rewindImpl.GetFunction().Builtin(ctx, []*values.Value{seekThisObj})
		if err != nil {
			t.Fatalf("rewind failed: %v", err)
		}

		// Seek to position 1
		seekMethod := class.Methods["seek"]
		if seekMethod == nil {
			t.Fatal("seek method not found")
		}

		seekImpl := seekMethod.Implementation.(*BuiltinMethodImpl)
		_, err = seekImpl.GetFunction().Builtin(ctx, []*values.Value{seekThisObj, values.NewInt(1)})
		if err != nil {
			t.Fatalf("seek failed: %v", err)
		}

		// Current should be the second element (value 2)
		currentMethod := class.Methods["current"]
		currentImpl := currentMethod.Implementation.(*BuiltinMethodImpl)
		result, err := currentImpl.GetFunction().Builtin(ctx, []*values.Value{seekThisObj})
		if err != nil {
			t.Fatalf("current after seek failed: %v", err)
		}

		if !result.IsInt() || result.ToInt() != 2 {
			t.Fatalf("Expected current value after seek to be 2, got: %v", result)
		}
	})
}