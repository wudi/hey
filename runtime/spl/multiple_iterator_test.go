package spl

import (
	"testing"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

func TestMultipleIterator(t *testing.T) {
	// Initialize registry
	registry.Initialize()
	ctx := &mockContext{registry: registry.GlobalRegistry}

	// Register all SPL classes for testing
	for _, class := range GetSplClasses() {
		if err := registry.GlobalRegistry.RegisterClass(class); err != nil {
			t.Fatalf("Failed to register class %s: %v", class.Name, err)
		}
	}

	// Get the MultipleIterator class
	class := GetMultipleIteratorClass()
	if class == nil {
		t.Fatal("MultipleIterator class is nil")
	}

	// Create MultipleIterator instance
	obj := &values.Object{
		ClassName:  "MultipleIterator",
		Properties: make(map[string]*values.Value),
	}
	thisObj := &values.Value{
		Type: values.TypeObject,
		Data: obj,
	}

	// Create test ArrayIterators
	arrayIterClass := GetArrayIteratorClass()

	// First array iterator [a, b, c]
	arrayIterObj1 := &values.Object{
		ClassName:  "ArrayIterator",
		Properties: make(map[string]*values.Value),
	}
	arrayIterValue1 := &values.Value{
		Type: values.TypeObject,
		Data: arrayIterObj1,
	}

	testArray1 := values.NewArray()
	testArray1.ArraySet(values.NewInt(0), values.NewString("a"))
	testArray1.ArraySet(values.NewInt(1), values.NewString("b"))
	testArray1.ArraySet(values.NewInt(2), values.NewString("c"))

	constructorArgs1 := []*values.Value{arrayIterValue1, testArray1}
	arrayIterConstructor := arrayIterClass.Methods["__construct"]
	impl1 := arrayIterConstructor.Implementation.(*BuiltinMethodImpl)
	_, err := impl1.GetFunction().Builtin(ctx, constructorArgs1)
	if err != nil {
		t.Fatalf("Failed to initialize ArrayIterator1: %v", err)
	}

	// Second array iterator [1, 2, 3]
	arrayIterObj2 := &values.Object{
		ClassName:  "ArrayIterator",
		Properties: make(map[string]*values.Value),
	}
	arrayIterValue2 := &values.Value{
		Type: values.TypeObject,
		Data: arrayIterObj2,
	}

	testArray2 := values.NewArray()
	testArray2.ArraySet(values.NewInt(0), values.NewInt(1))
	testArray2.ArraySet(values.NewInt(1), values.NewInt(2))
	testArray2.ArraySet(values.NewInt(2), values.NewInt(3))

	constructorArgs2 := []*values.Value{arrayIterValue2, testArray2}
	impl2 := arrayIterConstructor.Implementation.(*BuiltinMethodImpl)
	_, err = impl2.GetFunction().Builtin(ctx, constructorArgs2)
	if err != nil {
		t.Fatalf("Failed to initialize ArrayIterator2: %v", err)
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
		if iterators, ok := obj.Properties["__iterators"]; !ok || iterators == nil {
			t.Fatal("Iterators array not initialized")
		}
	})

	t.Run("countIterators", func(t *testing.T) {
		method := class.Methods["countIterators"]
		if method == nil {
			t.Fatal("countIterators method not found")
		}

		// Initially should be 0
		args := []*values.Value{thisObj}
		impl := method.Implementation.(*BuiltinMethodImpl)
		result, err := impl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("countIterators failed: %v", err)
		}

		if !result.IsInt() || result.ToInt() != 0 {
			t.Fatalf("Expected count 0, got %v", result)
		}
	})

	t.Run("attachIterator", func(t *testing.T) {
		method := class.Methods["attachIterator"]
		if method == nil {
			t.Fatal("attachIterator method not found")
		}

		// Attach first iterator
		args := []*values.Value{thisObj, arrayIterValue1}
		impl := method.Implementation.(*BuiltinMethodImpl)
		_, err := impl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("attachIterator failed: %v", err)
		}

		// Check count increased
		countMethod := class.Methods["countIterators"]
		countImpl := countMethod.Implementation.(*BuiltinMethodImpl)
		countResult, err := countImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("countIterators after attach failed: %v", err)
		}

		if !countResult.IsInt() || countResult.ToInt() != 1 {
			t.Fatalf("Expected count 1 after attach, got %v", countResult)
		}
	})

	t.Run("containsIterator", func(t *testing.T) {
		method := class.Methods["containsIterator"]
		if method == nil {
			t.Fatal("containsIterator method not found")
		}

		// Should contain first iterator
		args := []*values.Value{thisObj, arrayIterValue1}
		impl := method.Implementation.(*BuiltinMethodImpl)
		result, err := impl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("containsIterator failed: %v", err)
		}

		if !result.ToBool() {
			t.Fatal("Should contain attached iterator")
		}

		// Should not contain second iterator yet
		args = []*values.Value{thisObj, arrayIterValue2}
		result, err = impl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("containsIterator failed: %v", err)
		}

		if result.ToBool() {
			t.Fatal("Should not contain unattached iterator")
		}
	})

	t.Run("ParallelIteration", func(t *testing.T) {
		// First rewind both ArrayIterators to ensure clean state
		arrayRewindMethod := arrayIterClass.Methods["rewind"]
		arrayRewindImpl := arrayRewindMethod.Implementation.(*BuiltinMethodImpl)
		_, err := arrayRewindImpl.GetFunction().Builtin(ctx, []*values.Value{arrayIterValue1})
		if err != nil {
			t.Fatalf("Failed to rewind first iterator: %v", err)
		}
		_, err = arrayRewindImpl.GetFunction().Builtin(ctx, []*values.Value{arrayIterValue2})
		if err != nil {
			t.Fatalf("Failed to rewind second iterator: %v", err)
		}

		// Attach second iterator for parallel iteration
		attachMethod := class.Methods["attachIterator"]
		attachImpl := attachMethod.Implementation.(*BuiltinMethodImpl)
		_, err = attachImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj, arrayIterValue2})
		if err != nil {
			t.Fatalf("Failed to attach second iterator: %v", err)
		}

		// Get iteration methods
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

		// Rewind MultipleIterator to start
		_, err = rewindImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("rewind failed: %v", err)
		}

		expectedValues := [][]string{
			{"a", "1"}, {"b", "2"}, {"c", "3"},
		}

		for i := 0; i < 3; i++ {
			// Check valid
			validResult, err := validImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
			if err != nil {
				t.Fatalf("valid failed at iteration %d: %v", i, err)
			}

			if !validResult.ToBool() {
				t.Fatalf("Should be valid at iteration %d", i)
			}

			// Check current values (should be array)
			currentResult, err := currentImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
			if err != nil {
				t.Fatalf("current failed at iteration %d: %v", i, err)
			}

			if !currentResult.IsArray() {
				t.Fatalf("Current should return array at iteration %d, got %v", i, currentResult)
			}

			currentArray := currentResult.Data.(*values.Array)

			// Check first value
			val0, exists := currentArray.Elements[int64(0)]
			if !exists || !val0.IsString() || val0.ToString() != expectedValues[i][0] {
				t.Fatalf("Expected first value '%s' at iteration %d, got %v", expectedValues[i][0], i, val0)
			}

			// Check second value
			val1, exists := currentArray.Elements[int64(1)]
			if !exists || !val1.IsInt() || val1.ToString() != expectedValues[i][1] {
				t.Fatalf("Expected second value '%s' at iteration %d, got %v", expectedValues[i][1], i, val1)
			}

			// Check keys (should also be array)
			keyResult, err := keyImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
			if err != nil {
				t.Fatalf("key failed at iteration %d: %v", i, err)
			}

			if !keyResult.IsArray() {
				t.Fatalf("Key should return array at iteration %d", i)
			}

			// Move to next
			_, err = nextImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
			if err != nil {
				t.Fatalf("next failed at iteration %d: %v", i, err)
			}
		}

		// Should be invalid after 3 iterations
		validResult, err := validImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("final valid check failed: %v", err)
		}

		if validResult.ToBool() {
			t.Fatal("Should be invalid after all elements consumed")
		}
	})

	t.Run("detachIterator", func(t *testing.T) {
		method := class.Methods["detachIterator"]
		if method == nil {
			t.Fatal("detachIterator method not found")
		}

		// Detach first iterator
		args := []*values.Value{thisObj, arrayIterValue1}
		impl := method.Implementation.(*BuiltinMethodImpl)
		_, err := impl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("detachIterator failed: %v", err)
		}

		// Check count decreased
		countMethod := class.Methods["countIterators"]
		countImpl := countMethod.Implementation.(*BuiltinMethodImpl)
		countResult, err := countImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("countIterators after detach failed: %v", err)
		}

		if !countResult.IsInt() || countResult.ToInt() != 1 {
			t.Fatalf("Expected count 1 after detach, got %v", countResult)
		}

		// Should no longer contain first iterator
		containsMethod := class.Methods["containsIterator"]
		containsImpl := containsMethod.Implementation.(*BuiltinMethodImpl)
		containsResult, err := containsImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj, arrayIterValue1})
		if err != nil {
			t.Fatalf("containsIterator after detach failed: %v", err)
		}

		if containsResult.ToBool() {
			t.Fatal("Should not contain detached iterator")
		}
	})

	t.Run("EmptyIteratorHandling", func(t *testing.T) {
		// Create new MultipleIterator for this test
		emptyMultiObj := &values.Object{
			ClassName:  "MultipleIterator",
			Properties: make(map[string]*values.Value),
		}
		emptyMultiValue := &values.Value{
			Type: values.TypeObject,
			Data: emptyMultiObj,
		}

		// Initialize
		constructor := class.Methods["__construct"]
		constructorImpl := constructor.Implementation.(*BuiltinMethodImpl)
		_, err := constructorImpl.GetFunction().Builtin(ctx, []*values.Value{emptyMultiValue})
		if err != nil {
			t.Fatalf("Constructor failed: %v", err)
		}

		// Create empty ArrayIterator
		emptyArrayIterObj := &values.Object{
			ClassName:  "ArrayIterator",
			Properties: make(map[string]*values.Value),
		}
		emptyArrayIterValue := &values.Value{
			Type: values.TypeObject,
			Data: emptyArrayIterObj,
		}

		emptyArray := values.NewArray()
		emptyConstructorArgs := []*values.Value{emptyArrayIterValue, emptyArray}
		_, err = impl1.GetFunction().Builtin(ctx, emptyConstructorArgs)
		if err != nil {
			t.Fatalf("Failed to initialize empty ArrayIterator: %v", err)
		}

		// Attach empty iterator
		attachMethod := class.Methods["attachIterator"]
		attachImpl := attachMethod.Implementation.(*BuiltinMethodImpl)
		_, err = attachImpl.GetFunction().Builtin(ctx, []*values.Value{emptyMultiValue, emptyArrayIterValue})
		if err != nil {
			t.Fatalf("Failed to attach empty iterator: %v", err)
		}

		// Should not be valid (empty iterator)
		validMethod := class.Methods["valid"]
		validImpl := validMethod.Implementation.(*BuiltinMethodImpl)
		validResult, err := validImpl.GetFunction().Builtin(ctx, []*values.Value{emptyMultiValue})
		if err != nil {
			t.Fatalf("valid on empty multiple iterator failed: %v", err)
		}

		if validResult.ToBool() {
			t.Fatal("MultipleIterator with empty iterator should not be valid")
		}
	})

	t.Run("BasicIteratorMethods", func(t *testing.T) {
		// Test that all basic Iterator methods exist and can be called
		requiredMethods := []string{"current", "key", "valid", "next", "rewind"}

		for _, methodName := range requiredMethods {
			method := class.Methods[methodName]
			if method == nil {
				t.Fatalf("Required method %s not found", methodName)
			}

			// Test that method can be called (even if result might not be meaningful)
			args := []*values.Value{thisObj}
			impl := method.Implementation.(*BuiltinMethodImpl)
			_, err := impl.GetFunction().Builtin(ctx, args)
			if err != nil {
				t.Fatalf("Method %s failed: %v", methodName, err)
			}
		}
	})
}