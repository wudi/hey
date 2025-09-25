package spl

import (
	"testing"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

func TestInfiniteIterator(t *testing.T) {
	// Initialize registry
	registry.Initialize()
	ctx := &mockContext{registry: registry.GlobalRegistry}

	// Register all SPL classes for testing
	for _, class := range GetSplClasses() {
		if err := registry.GlobalRegistry.RegisterClass(class); err != nil {
			t.Fatalf("Failed to register class %s: %v", class.Name, err)
		}
	}

	// Get the InfiniteIterator class
	class := GetInfiniteIteratorClass()
	if class == nil {
		t.Fatal("InfiniteIterator class is nil")
	}

	// Create test ArrayIterator
	arrayIterClass := GetArrayIteratorClass()
	arrayIterObj := &values.Object{
		ClassName:  "ArrayIterator",
		Properties: make(map[string]*values.Value),
	}
	arrayIterValue := &values.Value{
		Type: values.TypeObject,
		Data: arrayIterObj,
	}

	// Initialize ArrayIterator with test data [a, b]
	testArray := values.NewArray()
	testArray.ArraySet(values.NewInt(0), values.NewString("a"))
	testArray.ArraySet(values.NewInt(1), values.NewString("b"))

	constructorArgs := []*values.Value{arrayIterValue, testArray}
	arrayIterConstructor := arrayIterClass.Methods["__construct"]
	impl := arrayIterConstructor.Implementation.(*BuiltinMethodImpl)
	_, err := impl.GetFunction().Builtin(ctx, constructorArgs)
	if err != nil {
		t.Fatalf("Failed to initialize ArrayIterator: %v", err)
	}

	// Create InfiniteIterator instance
	obj := &values.Object{
		ClassName:  "InfiniteIterator",
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

		args := []*values.Value{thisObj, arrayIterValue}
		impl := constructor.Implementation.(*BuiltinMethodImpl)
		_, err := impl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("Constructor failed: %v", err)
		}

		// Check that iterator was stored
		if iterator, ok := obj.Properties["__iterator"]; !ok || iterator != arrayIterValue {
			t.Fatal("Iterator not stored correctly")
		}
	})

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

	t.Run("InfiniteCycling", func(t *testing.T) {
		// Test that iterator cycles through elements infinitely
		rewindMethod := class.Methods["rewind"]
		validMethod := class.Methods["valid"]
		currentMethod := class.Methods["current"]
		keyMethod := class.Methods["key"]
		nextMethod := class.Methods["next"]

		rewindImpl := rewindMethod.Implementation.(*BuiltinMethodImpl)
		validImpl := validMethod.Implementation.(*BuiltinMethodImpl)
		currentImpl := currentMethod.Implementation.(*BuiltinMethodImpl)
		keyImpl := keyMethod.Implementation.(*BuiltinMethodImpl)
		nextImpl := nextMethod.Implementation.(*BuiltinMethodImpl)

		// Rewind to start
		_, err := rewindImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("rewind failed: %v", err)
		}

		expectedValues := []string{"a", "b", "a", "b", "a", "b"} // Should cycle
		expectedKeys := []int64{0, 1, 0, 1, 0, 1}

		for i := 0; i < 6; i++ { // Test 3 full cycles
			// Check valid
			validResult, err := validImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
			if err != nil {
				t.Fatalf("valid failed at iteration %d: %v", i, err)
			}

			if !validResult.ToBool() {
				t.Fatalf("Should always be valid in infinite iterator, failed at iteration %d", i)
			}

			// Check current value
			currentResult, err := currentImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
			if err != nil {
				t.Fatalf("current failed at iteration %d: %v", i, err)
			}

			if !currentResult.IsString() || currentResult.ToString() != expectedValues[i] {
				t.Fatalf("Expected value '%s' at iteration %d, got %v", expectedValues[i], i, currentResult)
			}

			// Check key
			keyResult, err := keyImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
			if err != nil {
				t.Fatalf("key failed at iteration %d: %v", i, err)
			}

			if !keyResult.IsInt() || keyResult.ToInt() != expectedKeys[i] {
				t.Fatalf("Expected key %d at iteration %d, got %v", expectedKeys[i], i, keyResult)
			}

			// Move to next
			_, err = nextImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
			if err != nil {
				t.Fatalf("next failed at iteration %d: %v", i, err)
			}
		}
	})

	t.Run("RewindAlwaysWorks", func(t *testing.T) {
		rewindMethod := class.Methods["rewind"]
		currentMethod := class.Methods["current"]
		nextMethod := class.Methods["next"]

		rewindImpl := rewindMethod.Implementation.(*BuiltinMethodImpl)
		currentImpl := currentMethod.Implementation.(*BuiltinMethodImpl)
		nextImpl := nextMethod.Implementation.(*BuiltinMethodImpl)

		// Advance several positions
		for i := 0; i < 5; i++ {
			nextImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		}

		// Should be at some position (likely 'a' since we cycled)
		_, _ = currentImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})

		// Rewind should always reset to first element
		rewindImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})

		// Should now be back at first element
		newCurrentResult, err := currentImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("current after rewind failed: %v", err)
		}

		if !newCurrentResult.IsString() || newCurrentResult.ToString() != "a" {
			t.Fatalf("Expected 'a' after rewind, got %v", newCurrentResult)
		}
	})

	t.Run("OuterIteratorInterface", func(t *testing.T) {
		// Should implement OuterIterator interface
		if _, exists := class.Methods["getInnerIterator"]; !exists {
			t.Fatal("Should implement OuterIterator interface")
		}
	})

	t.Run("BasicIteratorMethods", func(t *testing.T) {
		// Test that all basic Iterator methods exist
		requiredMethods := []string{"current", "key", "valid", "next", "rewind"}

		for _, methodName := range requiredMethods {
			method := class.Methods[methodName]
			if method == nil {
				t.Fatalf("Required method %s not found", methodName)
			}

			// Test that method can be called
			args := []*values.Value{thisObj}
			impl := method.Implementation.(*BuiltinMethodImpl)
			_, err := impl.GetFunction().Builtin(ctx, args)
			if err != nil {
				t.Fatalf("Method %s failed: %v", methodName, err)
			}
		}
	})

	t.Run("EmptyIteratorHandling", func(t *testing.T) {
		// Test with empty iterator - should remain invalid
		emptyArrayIter := &values.Object{
			ClassName:  "ArrayIterator",
			Properties: make(map[string]*values.Value),
		}
		emptyArrayValue := &values.Value{
			Type: values.TypeObject,
			Data: emptyArrayIter,
		}

		// Initialize with empty array
		emptyArray := values.NewArray()
		emptyConstructorArgs := []*values.Value{emptyArrayValue, emptyArray}
		arrayIterConstructor := arrayIterClass.Methods["__construct"]
		arrayImplConstructor := arrayIterConstructor.Implementation.(*BuiltinMethodImpl)
		_, err := arrayImplConstructor.GetFunction().Builtin(ctx, emptyConstructorArgs)
		if err != nil {
			t.Fatalf("Failed to initialize empty ArrayIterator: %v", err)
		}

		// Create InfiniteIterator with empty inner iterator
		emptyInfiniteObj := &values.Object{
			ClassName:  "InfiniteIterator",
			Properties: make(map[string]*values.Value),
		}
		emptyInfiniteValue := &values.Value{
			Type: values.TypeObject,
			Data: emptyInfiniteObj,
		}

		constructor := class.Methods["__construct"]
		constructorImpl := constructor.Implementation.(*BuiltinMethodImpl)
		_, err = constructorImpl.GetFunction().Builtin(ctx, []*values.Value{emptyInfiniteValue, emptyArrayValue})
		if err != nil {
			t.Fatalf("Constructor with empty iterator failed: %v", err)
		}

		// Should not be valid (empty iterator stays empty)
		validMethod := class.Methods["valid"]
		validImpl := validMethod.Implementation.(*BuiltinMethodImpl)
		validResult, err := validImpl.GetFunction().Builtin(ctx, []*values.Value{emptyInfiniteValue})
		if err != nil {
			t.Fatalf("valid on empty infinite iterator failed: %v", err)
		}

		if validResult.ToBool() {
			t.Fatal("Empty InfiniteIterator should not be valid")
		}
	})
}