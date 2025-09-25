package spl

import (
	"testing"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

func TestFilterIterator(t *testing.T) {
	// Initialize registry
	registry.Initialize()
	ctx := &mockContext{registry: registry.GlobalRegistry}

	// Register all SPL classes for testing
	for _, class := range GetSplClasses() {
		if err := registry.GlobalRegistry.RegisterClass(class); err != nil {
			t.Fatalf("Failed to register class %s: %v", class.Name, err)
		}
	}

	// Get the FilterIterator class
	class := GetFilterIteratorClass()
	if class == nil {
		t.Fatal("FilterIterator class is nil")
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

	// Since FilterIterator is abstract, we create a mock concrete implementation
	// that only accepts even numbers
	mockFilterObj := &values.Object{
		ClassName:  "FilterIterator",
		Properties: make(map[string]*values.Value),
	}
	mockFilterValue := &values.Value{
		Type: values.TypeObject,
		Data: mockFilterObj,
	}

	t.Run("Constructor", func(t *testing.T) {
		constructor := class.Methods["__construct"]
		if constructor == nil {
			t.Fatal("Constructor not found")
		}

		args := []*values.Value{mockFilterValue, arrayIterValue}
		impl := constructor.Implementation.(*BuiltinMethodImpl)
		_, err := impl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("Constructor failed: %v", err)
		}

		// Check that iterator was stored
		if iterator, ok := mockFilterObj.Properties["__iterator"]; !ok || iterator != arrayIterValue {
			t.Fatal("Iterator not stored correctly")
		}
	})

	t.Run("getInnerIterator", func(t *testing.T) {
		method := class.Methods["getInnerIterator"]
		if method == nil {
			t.Fatal("getInnerIterator method not found")
		}

		args := []*values.Value{mockFilterValue}
		impl := method.Implementation.(*BuiltinMethodImpl)
		result, err := impl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("getInnerIterator failed: %v", err)
		}

		if result != arrayIterValue {
			t.Fatal("getInnerIterator should return the wrapped iterator")
		}
	})

	t.Run("accept_abstract", func(t *testing.T) {
		// FilterIterator.accept() should be abstract and throw an error
		method := class.Methods["accept"]
		if method == nil {
			t.Fatal("accept method not found")
		}

		args := []*values.Value{mockFilterValue}
		impl := method.Implementation.(*BuiltinMethodImpl)
		_, err := impl.GetFunction().Builtin(ctx, args)

		// Should return error since accept() is abstract
		if err == nil {
			t.Fatal("Expected error from abstract accept() method")
		}
	})

	// Test with a mock concrete subclass that implements accept()
	t.Run("ConcreteFilterTest", func(t *testing.T) {
		// Create a mock subclass that filters even numbers
		// We'll simulate this by overriding the accept method behavior

		// Create a new filter object for testing filtering behavior
		concreteFilterObj := &values.Object{
			ClassName:  "FilterIterator",
			Properties: make(map[string]*values.Value),
		}
		concreteFilterValue := &values.Value{
			Type: values.TypeObject,
			Data: concreteFilterObj,
		}

		// Initialize with constructor
		constructor := class.Methods["__construct"]
		args := []*values.Value{concreteFilterValue, arrayIterValue}
		impl := constructor.Implementation.(*BuiltinMethodImpl)
		_, err := impl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("Constructor failed: %v", err)
		}

		// Add a mock accept implementation that returns true for even numbers
		concreteFilterObj.Properties["__accept_impl"] = values.NewString("even_numbers")

		// Test rewind
		rewindMethod := class.Methods["rewind"]
		rewindImpl := rewindMethod.Implementation.(*BuiltinMethodImpl)
		_, err = rewindImpl.GetFunction().Builtin(ctx, []*values.Value{concreteFilterValue})
		if err != nil {
			t.Fatalf("rewind failed: %v", err)
		}

		// Test that basic iterator methods work (they should delegate to inner iterator after filtering)
		validMethod := class.Methods["valid"]
		validImpl := validMethod.Implementation.(*BuiltinMethodImpl)
		result, err := validImpl.GetFunction().Builtin(ctx, []*values.Value{concreteFilterValue})
		if err != nil {
			t.Fatalf("valid failed: %v", err)
		}

		// Since we have numbers 1-6 and filtering for evens, should be valid (2,4,6 exist)
		if !result.ToBool() {
			t.Fatal("Expected valid() to return true for filtered iterator")
		}
	})

	t.Run("BasicMethods", func(t *testing.T) {
		// Test that FilterIterator has all the basic Iterator methods
		requiredMethods := []string{"current", "key", "valid", "next", "rewind"}

		for _, methodName := range requiredMethods {
			method := class.Methods[methodName]
			if method == nil {
				t.Fatalf("Required method %s not found", methodName)
			}
		}
	})
}