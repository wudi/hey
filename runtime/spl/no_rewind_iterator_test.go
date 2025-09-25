package spl

import (
	"testing"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

func TestNoRewindIterator(t *testing.T) {
	// Initialize registry
	registry.Initialize()
	ctx := &mockContext{registry: registry.GlobalRegistry}

	// Register all SPL classes for testing
	for _, class := range GetSplClasses() {
		if err := registry.GlobalRegistry.RegisterClass(class); err != nil {
			t.Fatalf("Failed to register class %s: %v", class.Name, err)
		}
	}

	// Get the NoRewindIterator class
	class := GetNoRewindIteratorClass()
	if class == nil {
		t.Fatal("NoRewindIterator class is nil")
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

	// Initialize ArrayIterator with test data
	testArray := values.NewArray()
	testArray.ArraySet(values.NewInt(0), values.NewString("a"))
	testArray.ArraySet(values.NewInt(1), values.NewString("b"))
	testArray.ArraySet(values.NewInt(2), values.NewString("c"))

	constructorArgs := []*values.Value{arrayIterValue, testArray}
	arrayIterConstructor := arrayIterClass.Methods["__construct"]
	impl := arrayIterConstructor.Implementation.(*BuiltinMethodImpl)
	_, err := impl.GetFunction().Builtin(ctx, constructorArgs)
	if err != nil {
		t.Fatalf("Failed to initialize ArrayIterator: %v", err)
	}

	// Create NoRewindIterator instance
	obj := &values.Object{
		ClassName:  "NoRewindIterator",
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

		// Should track rewind state
		if rewindCalled, ok := obj.Properties["__rewind_called"]; !ok || !rewindCalled.IsBool() {
			t.Fatal("Rewind state not initialized correctly")
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

	t.Run("FirstRewindWorks", func(t *testing.T) {
		// First rewind should work
		rewindMethod := class.Methods["rewind"]
		rewindImpl := rewindMethod.Implementation.(*BuiltinMethodImpl)
		_, err := rewindImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("First rewind failed: %v", err)
		}

		// Should be valid and positioned at first element
		validMethod := class.Methods["valid"]
		validImpl := validMethod.Implementation.(*BuiltinMethodImpl)
		result, err := validImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("valid after first rewind failed: %v", err)
		}

		if !result.ToBool() {
			t.Fatal("Should be valid after first rewind")
		}

		// Check current value
		currentMethod := class.Methods["current"]
		currentImpl := currentMethod.Implementation.(*BuiltinMethodImpl)
		currentResult, err := currentImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("current after rewind failed: %v", err)
		}

		if !currentResult.IsString() || currentResult.ToString() != "a" {
			t.Fatalf("Expected 'a', got %v", currentResult)
		}
	})

	t.Run("SubsequentRewindsIgnored", func(t *testing.T) {
		// Advance the iterator
		nextMethod := class.Methods["next"]
		nextImpl := nextMethod.Implementation.(*BuiltinMethodImpl)
		nextImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})

		// Should now be at second element
		currentMethod := class.Methods["current"]
		currentImpl := currentMethod.Implementation.(*BuiltinMethodImpl)
		currentResult, _ := currentImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})

		if !currentResult.IsString() || currentResult.ToString() != "b" {
			t.Fatalf("Expected 'b' after next(), got %v", currentResult)
		}

		// Try to rewind again - should be ignored
		rewindMethod := class.Methods["rewind"]
		rewindImpl := rewindMethod.Implementation.(*BuiltinMethodImpl)
		rewindImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})

		// Should still be at second element
		currentResult2, _ := currentImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if !currentResult2.IsString() || currentResult2.ToString() != "b" {
			t.Fatal("Rewind should have been ignored - should still be at 'b'")
		}
	})

	t.Run("BasicIteratorMethods", func(t *testing.T) {
		// Test that all basic Iterator methods exist and delegate correctly
		requiredMethods := []string{"current", "key", "valid", "next"}

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

	t.Run("OuterIteratorInterface", func(t *testing.T) {
		// Should implement OuterIterator interface
		if _, exists := class.Methods["getInnerIterator"]; !exists {
			t.Fatal("Should implement OuterIterator interface")
		}
	})

	t.Run("OneTimeIteration", func(t *testing.T) {
		// Create a fresh NoRewindIterator for this test
		freshObj := &values.Object{
			ClassName:  "NoRewindIterator",
			Properties: make(map[string]*values.Value),
		}
		freshThisObj := &values.Value{
			Type: values.TypeObject,
			Data: freshObj,
		}

		constructor := class.Methods["__construct"]
		constructorImpl := constructor.Implementation.(*BuiltinMethodImpl)
		_, err := constructorImpl.GetFunction().Builtin(ctx, []*values.Value{freshThisObj, arrayIterValue})
		if err != nil {
			t.Fatalf("Constructor failed: %v", err)
		}

		rewindMethod := class.Methods["rewind"]
		validMethod := class.Methods["valid"]
		nextMethod := class.Methods["next"]

		rewindImpl := rewindMethod.Implementation.(*BuiltinMethodImpl)
		validImpl := validMethod.Implementation.(*BuiltinMethodImpl)
		nextImpl := nextMethod.Implementation.(*BuiltinMethodImpl)

		// First iteration - should work
		rewindImpl.GetFunction().Builtin(ctx, []*values.Value{freshThisObj})

		firstIterationCount := 0
		for {
			validResult, _ := validImpl.GetFunction().Builtin(ctx, []*values.Value{freshThisObj})
			if !validResult.ToBool() {
				break
			}
			nextImpl.GetFunction().Builtin(ctx, []*values.Value{freshThisObj})
			firstIterationCount++
			if firstIterationCount > 10 {
				break // Safety
			}
		}

		if firstIterationCount != 3 {
			t.Fatalf("Expected 3 elements in first iteration, got %d", firstIterationCount)
		}

		// Second iteration - should show no elements (iterator exhausted)
		secondIterationCount := 0
		for {
			validResult, _ := validImpl.GetFunction().Builtin(ctx, []*values.Value{freshThisObj})
			if !validResult.ToBool() {
				break
			}
			nextImpl.GetFunction().Builtin(ctx, []*values.Value{freshThisObj})
			secondIterationCount++
			if secondIterationCount > 10 {
				break // Safety
			}
		}

		if secondIterationCount != 0 {
			t.Fatalf("Expected 0 elements in second iteration, got %d", secondIterationCount)
		}
	})
}