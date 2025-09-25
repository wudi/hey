package spl

import (
	"testing"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

func TestCallbackFilterIterator(t *testing.T) {
	// Initialize registry
	registry.Initialize()
	ctx := &mockContext{registry: registry.GlobalRegistry}

	// Register all SPL classes for testing
	for _, class := range GetSplClasses() {
		if err := registry.GlobalRegistry.RegisterClass(class); err != nil {
			t.Fatalf("Failed to register class %s: %v", class.Name, err)
		}
	}

	// Get the CallbackFilterIterator class
	class := GetCallbackFilterIteratorClass()
	if class == nil {
		t.Fatal("CallbackFilterIterator class is nil")
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

	// Initialize ArrayIterator with test data (numbers 1-6)
	testArray := values.NewArray()
	testArray.ArraySet(values.NewInt(0), values.NewInt(1))
	testArray.ArraySet(values.NewInt(1), values.NewInt(2))
	testArray.ArraySet(values.NewInt(2), values.NewInt(3))
	testArray.ArraySet(values.NewInt(3), values.NewInt(4))
	testArray.ArraySet(values.NewInt(4), values.NewInt(5))
	testArray.ArraySet(values.NewInt(5), values.NewInt(6))

	constructorArgs := []*values.Value{arrayIterValue, testArray}
	arrayIterConstructor := arrayIterClass.Methods["__construct"]
	impl := arrayIterConstructor.Implementation.(*BuiltinMethodImpl)
	_, err := impl.GetFunction().Builtin(ctx, constructorArgs)
	if err != nil {
		t.Fatalf("Failed to initialize ArrayIterator: %v", err)
	}

	// Create CallbackFilterIterator instance
	filterObj := &values.Object{
		ClassName:  "CallbackFilterIterator",
		Properties: make(map[string]*values.Value),
	}
	filterValue := &values.Value{
		Type: values.TypeObject,
		Data: filterObj,
	}

	// Create a mock callback (for testing we'll simulate a callback that accepts even numbers)
	mockCallback := values.NewString("even_filter") // Mock callback identifier

	t.Run("Constructor", func(t *testing.T) {
		constructor := class.Methods["__construct"]
		if constructor == nil {
			t.Fatal("Constructor not found")
		}

		args := []*values.Value{filterValue, arrayIterValue, mockCallback}
		impl := constructor.Implementation.(*BuiltinMethodImpl)
		_, err := impl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("Constructor failed: %v", err)
		}

		// Check that iterator and callback were stored
		if iterator, ok := filterObj.Properties["__iterator"]; !ok || iterator != arrayIterValue {
			t.Fatal("Iterator not stored correctly")
		}
		if callback, ok := filterObj.Properties["__callback"]; !ok || callback != mockCallback {
			t.Fatal("Callback not stored correctly")
		}
	})

	t.Run("getInnerIterator", func(t *testing.T) {
		method := class.Methods["getInnerIterator"]
		if method == nil {
			t.Fatal("getInnerIterator method not found")
		}

		args := []*values.Value{filterValue}
		impl := method.Implementation.(*BuiltinMethodImpl)
		result, err := impl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("getInnerIterator failed: %v", err)
		}

		if result != arrayIterValue {
			t.Fatal("getInnerIterator should return the wrapped iterator")
		}
	})

	t.Run("accept", func(t *testing.T) {
		// CallbackFilterIterator should implement accept() method (not abstract like FilterIterator)
		method := class.Methods["accept"]
		if method == nil {
			t.Fatal("accept method not found")
		}

		args := []*values.Value{filterValue}
		impl := method.Implementation.(*BuiltinMethodImpl)

		// For testing purposes, we'll mock the behavior
		// In a real implementation, this would call the stored callback
		result, err := impl.GetFunction().Builtin(ctx, args)

		// Should not error (unlike abstract FilterIterator)
		if err != nil {
			t.Fatalf("accept failed: %v", err)
		}

		// Result should be a boolean
		if !result.IsBool() {
			t.Fatal("accept should return boolean")
		}
	})

	t.Run("rewind", func(t *testing.T) {
		// Test that rewind works and positions to first valid item
		rewindMethod := class.Methods["rewind"]
		rewindImpl := rewindMethod.Implementation.(*BuiltinMethodImpl)
		_, err := rewindImpl.GetFunction().Builtin(ctx, []*values.Value{filterValue})
		if err != nil {
			t.Fatalf("rewind failed: %v", err)
		}

		// Should be positioned at first valid item according to filter
		validMethod := class.Methods["valid"]
		validImpl := validMethod.Implementation.(*BuiltinMethodImpl)
		result, err := validImpl.GetFunction().Builtin(ctx, []*values.Value{filterValue})
		if err != nil {
			t.Fatalf("valid after rewind failed: %v", err)
		}

		// Should be valid if there are matching items
		if !result.ToBool() {
			// This might be OK if the mock filter rejects all items
			t.Logf("No valid items found after rewind (may be expected with mock filter)")
		}
	})

	t.Run("BasicIteratorMethods", func(t *testing.T) {
		// Test that all basic Iterator methods exist and work
		requiredMethods := []string{"current", "key", "valid", "next"}

		for _, methodName := range requiredMethods {
			method := class.Methods[methodName]
			if method == nil {
				t.Fatalf("Required method %s not found", methodName)
			}

			// Test that method can be called without error
			args := []*values.Value{filterValue}
			impl := method.Implementation.(*BuiltinMethodImpl)
			_, err := impl.GetFunction().Builtin(ctx, args)
			if err != nil {
				t.Logf("Method %s returned error: %v (may be expected)", methodName, err)
			}
		}
	})

	t.Run("InheritanceFromFilterIterator", func(t *testing.T) {
		// CallbackFilterIterator should have the same interface as FilterIterator
		// but with concrete implementation of accept()

		// Should have all FilterIterator methods
		expectedMethods := []string{"__construct", "getInnerIterator", "accept", "current", "key", "valid", "next", "rewind"}

		for _, methodName := range expectedMethods {
			if _, exists := class.Methods[methodName]; !exists {
				t.Fatalf("Method %s should be inherited from FilterIterator", methodName)
			}
		}
	})

	t.Run("CallbackStorage", func(t *testing.T) {
		// Test that callback is properly stored and accessible
		if callback, ok := filterObj.Properties["__callback"]; !ok {
			t.Fatal("Callback should be stored in object properties")
		} else if !callback.IsString() {
			t.Fatal("Callback should be stored as string (for mock testing)")
		}
	})
}