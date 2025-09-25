package spl

import (
	"testing"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

func TestRecursiveArrayIterator(t *testing.T) {
	// Initialize registry
	registry.Initialize()
	ctx := &mockContext{registry: registry.GlobalRegistry}

	// Register all SPL classes for testing
	for _, class := range GetSplClasses() {
		if err := registry.GlobalRegistry.RegisterClass(class); err != nil {
			t.Fatalf("Failed to register class %s: %v", class.Name, err)
		}
	}

	// Get the RecursiveArrayIterator class
	class := GetRecursiveArrayIteratorClass()
	if class == nil {
		t.Fatal("RecursiveArrayIterator class is nil")
	}

	// Create nested test array
	// [a => 1, b => [b1 => 2, b2 => 3], c => 4]
	nestedArray := values.NewArray()
	nestedArray.ArraySet(values.NewString("a"), values.NewInt(1))

	nestedSubArray := values.NewArray()
	nestedSubArray.ArraySet(values.NewString("b1"), values.NewInt(2))
	nestedSubArray.ArraySet(values.NewString("b2"), values.NewInt(3))
	nestedArray.ArraySet(values.NewString("b"), nestedSubArray)

	nestedArray.ArraySet(values.NewString("c"), values.NewInt(4))

	// Create RecursiveArrayIterator instance
	obj := &values.Object{
		ClassName:  "RecursiveArrayIterator",
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

		args := []*values.Value{thisObj, nestedArray}
		impl := constructor.Implementation.(*BuiltinMethodImpl)
		_, err := impl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("Constructor failed: %v", err)
		}

		// Should store the array like ArrayIterator
		if storedArray, ok := obj.Properties["__array"]; !ok || storedArray != nestedArray {
			t.Fatal("Array not stored correctly")
		}
	})

	t.Run("BasicIteration", func(t *testing.T) {
		// Should be able to iterate like ArrayIterator
		rewindMethod := class.Methods["rewind"]
		rewindImpl := rewindMethod.Implementation.(*BuiltinMethodImpl)
		_, err := rewindImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("rewind failed: %v", err)
		}

		validMethod := class.Methods["valid"]
		currentMethod := class.Methods["current"]
		keyMethod := class.Methods["key"]

		validImpl := validMethod.Implementation.(*BuiltinMethodImpl)
		currentImpl := currentMethod.Implementation.(*BuiltinMethodImpl)
		keyImpl := keyMethod.Implementation.(*BuiltinMethodImpl)

		// Should be at first element
		validResult, _ := validImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if !validResult.ToBool() {
			t.Fatal("Should be valid after rewind")
		}

		// First element should be "a" => 1
		keyResult, _ := keyImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		currentResult, _ := currentImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})

		if !keyResult.IsString() || keyResult.ToString() != "a" {
			t.Fatalf("Expected key 'a', got %v", keyResult)
		}

		if !currentResult.IsInt() || currentResult.ToInt() != 1 {
			t.Fatalf("Expected value 1, got %v", currentResult)
		}
	})

	t.Run("hasChildren", func(t *testing.T) {
		hasChildrenMethod := class.Methods["hasChildren"]
		if hasChildrenMethod == nil {
			t.Fatal("hasChildren method not found")
		}

		impl := hasChildrenMethod.Implementation.(*BuiltinMethodImpl)

		// Rewind to start
		rewindMethod := class.Methods["rewind"]
		rewindImpl := rewindMethod.Implementation.(*BuiltinMethodImpl)
		rewindImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})

		// First element ("a" => 1) should not have children
		result, err := impl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("hasChildren failed: %v", err)
		}

		if !result.IsBool() || result.ToBool() {
			t.Fatal("Element 'a' should not have children")
		}

		// Move to second element ("b" => [array])
		nextMethod := class.Methods["next"]
		nextImpl := nextMethod.Implementation.(*BuiltinMethodImpl)
		nextImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})

		// Second element should have children (it's an array)
		result, err = impl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("hasChildren failed on second element: %v", err)
		}

		if !result.IsBool() || !result.ToBool() {
			t.Fatal("Element 'b' should have children (it's an array)")
		}
	})

	t.Run("getChildren", func(t *testing.T) {
		getChildrenMethod := class.Methods["getChildren"]
		if getChildrenMethod == nil {
			t.Fatal("getChildren method not found")
		}

		impl := getChildrenMethod.Implementation.(*BuiltinMethodImpl)

		// Position at second element (the nested array)
		rewindMethod := class.Methods["rewind"]
		nextMethod := class.Methods["next"]
		rewindImpl := rewindMethod.Implementation.(*BuiltinMethodImpl)
		nextImpl := nextMethod.Implementation.(*BuiltinMethodImpl)

		rewindImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		nextImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj}) // Move to "b"

		// Get children
		result, err := impl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("getChildren failed: %v", err)
		}

		if !result.IsObject() {
			t.Fatal("getChildren should return an object (RecursiveArrayIterator)")
		}

		childIterator := result.Data.(*values.Object)
		if childIterator.ClassName != "RecursiveArrayIterator" {
			t.Fatalf("Expected RecursiveArrayIterator, got %s", childIterator.ClassName)
		}

		// The child iterator should contain the nested array
		if childArray, ok := childIterator.Properties["__array"]; !ok || childArray != nestedSubArray {
			t.Fatal("Child iterator should contain the nested sub-array")
		}
	})

	t.Run("RecursiveIteratorInterface", func(t *testing.T) {
		// Should implement RecursiveIterator interface methods
		requiredMethods := []string{"hasChildren", "getChildren"}

		for _, methodName := range requiredMethods {
			if _, exists := class.Methods[methodName]; !exists {
				t.Fatalf("RecursiveIterator method %s not found", methodName)
			}
		}
	})

	t.Run("ArrayIteratorInheritance", func(t *testing.T) {
		// Should have all ArrayIterator methods
		arrayIteratorMethods := []string{"current", "key", "valid", "next", "rewind"}

		for _, methodName := range arrayIteratorMethods {
			if _, exists := class.Methods[methodName]; !exists {
				t.Fatalf("ArrayIterator method %s not found", methodName)
			}
		}
	})

	t.Run("EmptyArrayHasChildren", func(t *testing.T) {
		// Test edge case: empty array should still have children (but count 0)
		emptyNestedArray := values.NewArray()
		emptyNestedArray.ArraySet(values.NewString("empty"), values.NewArray()) // empty sub-array

		emptyObj := &values.Object{
			ClassName:  "RecursiveArrayIterator",
			Properties: make(map[string]*values.Value),
		}
		emptyThisObj := &values.Value{
			Type: values.TypeObject,
			Data: emptyObj,
		}

		// Initialize with empty nested array
		constructor := class.Methods["__construct"]
		constructorImpl := constructor.Implementation.(*BuiltinMethodImpl)
		_, err := constructorImpl.GetFunction().Builtin(ctx, []*values.Value{emptyThisObj, emptyNestedArray})
		if err != nil {
			t.Fatalf("Constructor failed: %v", err)
		}

		// Rewind and check hasChildren for empty array
		rewindMethod := class.Methods["rewind"]
		rewindImpl := rewindMethod.Implementation.(*BuiltinMethodImpl)
		rewindImpl.GetFunction().Builtin(ctx, []*values.Value{emptyThisObj})

		hasChildrenMethod := class.Methods["hasChildren"]
		hasChildrenImpl := hasChildrenMethod.Implementation.(*BuiltinMethodImpl)
		result, err := hasChildrenImpl.GetFunction().Builtin(ctx, []*values.Value{emptyThisObj})
		if err != nil {
			t.Fatalf("hasChildren failed: %v", err)
		}

		// Empty array should still report hasChildren = true
		if !result.ToBool() {
			t.Fatal("Empty array should still have children (even if count is 0)")
		}
	})
}