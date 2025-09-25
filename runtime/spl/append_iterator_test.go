package spl

import (
	"testing"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

func TestAppendIterator(t *testing.T) {
	// Initialize registry
	registry.Initialize()
	ctx := &mockContext{registry: registry.GlobalRegistry}

	// Register all SPL classes for testing
	for _, class := range GetSplClasses() {
		if err := registry.GlobalRegistry.RegisterClass(class); err != nil {
			t.Fatalf("Failed to register class %s: %v", class.Name, err)
		}
	}

	// Get the AppendIterator class
	class := GetAppendIteratorClass()
	if class == nil {
		t.Fatal("AppendIterator class is nil")
	}

	// Create test ArrayIterators
	arrayIterClass := GetArrayIteratorClass()

	// Create first ArrayIterator
	arrayIterObj1 := &values.Object{
		ClassName:  "ArrayIterator",
		Properties: make(map[string]*values.Value),
	}
	arrayIterValue1 := &values.Value{
		Type: values.TypeObject,
		Data: arrayIterObj1,
	}

	testArray1 := values.NewArray()
	testArray1.ArraySet(values.NewString("a"), values.NewInt(1))
	testArray1.ArraySet(values.NewString("b"), values.NewInt(2))

	constructorArgs1 := []*values.Value{arrayIterValue1, testArray1}
	arrayIterConstructor := arrayIterClass.Methods["__construct"]
	impl := arrayIterConstructor.Implementation.(*BuiltinMethodImpl)
	_, err := impl.GetFunction().Builtin(ctx, constructorArgs1)
	if err != nil {
		t.Fatalf("Failed to initialize ArrayIterator1: %v", err)
	}

	// Create second ArrayIterator
	arrayIterObj2 := &values.Object{
		ClassName:  "ArrayIterator",
		Properties: make(map[string]*values.Value),
	}
	arrayIterValue2 := &values.Value{
		Type: values.TypeObject,
		Data: arrayIterObj2,
	}

	testArray2 := values.NewArray()
	testArray2.ArraySet(values.NewString("c"), values.NewInt(3))
	testArray2.ArraySet(values.NewString("d"), values.NewInt(4))

	constructorArgs2 := []*values.Value{arrayIterValue2, testArray2}
	_, err = impl.GetFunction().Builtin(ctx, constructorArgs2)
	if err != nil {
		t.Fatalf("Failed to initialize ArrayIterator2: %v", err)
	}

	// Create AppendIterator instance
	obj := &values.Object{
		ClassName:  "AppendIterator",
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

		args := []*values.Value{thisObj}
		impl := constructor.Implementation.(*BuiltinMethodImpl)
		_, err := impl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("Constructor failed: %v", err)
		}

		// Check that iterators array was initialized
		if iterators, ok := obj.Properties["__iterators"]; !ok || !iterators.IsArray() {
			t.Fatal("Iterators array not initialized correctly")
		}
		if currentIndex, ok := obj.Properties["__current_index"]; !ok || !currentIndex.IsInt() || currentIndex.ToInt() != 0 {
			t.Fatal("Current index not initialized correctly")
		}
	})

	t.Run("append", func(t *testing.T) {
		appendMethod := class.Methods["append"]
		if appendMethod == nil {
			t.Fatal("append method not found")
		}

		// Append first iterator
		args := []*values.Value{thisObj, arrayIterValue1}
		impl := appendMethod.Implementation.(*BuiltinMethodImpl)
		_, err := impl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("append first iterator failed: %v", err)
		}

		// Append second iterator
		args = []*values.Value{thisObj, arrayIterValue2}
		_, err = impl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("append second iterator failed: %v", err)
		}

		// Check that iterators were stored
		iterators := obj.Properties["__iterators"]
		if iterators.ArrayCount() != 2 {
			t.Fatalf("Expected 2 iterators, got %d", iterators.ArrayCount())
		}
	})

	t.Run("getIteratorIndex", func(t *testing.T) {
		method := class.Methods["getIteratorIndex"]
		if method == nil {
			t.Fatal("getIteratorIndex method not found")
		}

		args := []*values.Value{thisObj}
		impl := method.Implementation.(*BuiltinMethodImpl)
		result, err := impl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("getIteratorIndex failed: %v", err)
		}

		if !result.IsInt() || result.ToInt() != 0 {
			t.Fatalf("Expected iterator index 0, got %v", result)
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

		if !result.IsObject() {
			t.Fatal("Expected object result from getInnerIterator")
		}

		// Should return the first iterator since we haven't moved yet
		if result != arrayIterValue1 {
			t.Fatal("getInnerIterator should return the current active iterator")
		}
	})

	t.Run("rewindAndIteration", func(t *testing.T) {
		// Rewind the AppendIterator
		rewindMethod := class.Methods["rewind"]
		rewindImpl := rewindMethod.Implementation.(*BuiltinMethodImpl)
		_, err := rewindImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("rewind failed: %v", err)
		}

		// Should be at the first element of first iterator
		currentMethod := class.Methods["current"]
		currentImpl := currentMethod.Implementation.(*BuiltinMethodImpl)
		result, err := currentImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("current failed: %v", err)
		}

		if !result.IsInt() || result.ToInt() != 1 {
			t.Fatalf("Expected current value to be 1, got: %v", result)
		}

		// Check key
		keyMethod := class.Methods["key"]
		keyImpl := keyMethod.Implementation.(*BuiltinMethodImpl)
		keyResult, err := keyImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("key failed: %v", err)
		}

		if !keyResult.IsString() || keyResult.ToString() != "a" {
			t.Fatalf("Expected key to be 'a', got: %v", keyResult.ToString())
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

	t.Run("FullIteration", func(t *testing.T) {
		// Rewind to start
		rewindMethod := class.Methods["rewind"]
		rewindImpl := rewindMethod.Implementation.(*BuiltinMethodImpl)
		_, err := rewindImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("rewind failed: %v", err)
		}

		validMethod := class.Methods["valid"]
		currentMethod := class.Methods["current"]
		keyMethod := class.Methods["key"]
		nextMethod := class.Methods["next"]

		validImpl := validMethod.Implementation.(*BuiltinMethodImpl)
		currentImpl := currentMethod.Implementation.(*BuiltinMethodImpl)
		keyImpl := keyMethod.Implementation.(*BuiltinMethodImpl)
		nextImpl := nextMethod.Implementation.(*BuiltinMethodImpl)

		expectedKeys := []string{"a", "b", "c", "d"}
		expectedValues := []int64{1, 2, 3, 4}
		count := 0

		for {
			// Check if valid
			args := []*values.Value{thisObj}
			validResult, _ := validImpl.GetFunction().Builtin(ctx, args)
			if !validResult.ToBool() {
				break
			}

			// Get current key and value
			keyResult, _ := keyImpl.GetFunction().Builtin(ctx, args)
			currentResult, _ := currentImpl.GetFunction().Builtin(ctx, args)

			if count >= len(expectedKeys) {
				t.Fatalf("Too many iterations, expected %d", len(expectedKeys))
			}

			if keyResult.ToString() != expectedKeys[count] {
				t.Fatalf("Expected key '%s', got '%s'", expectedKeys[count], keyResult.ToString())
			}

			if currentResult.ToInt() != expectedValues[count] {
				t.Fatalf("Expected value %d, got %d", expectedValues[count], currentResult.ToInt())
			}

			// Move to next
			nextImpl.GetFunction().Builtin(ctx, args)
			count++

			if count > 10 { // Safety check
				break
			}
		}

		if count != 4 {
			t.Fatalf("Expected 4 iterations, got %d", count)
		}
	})
}