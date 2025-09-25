package spl

import (
	"testing"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

func TestCachingIterator(t *testing.T) {
	// Initialize registry
	registry.Initialize()
	ctx := &mockContext{registry: registry.GlobalRegistry}

	// Register all SPL classes for testing
	for _, class := range GetSplClasses() {
		if err := registry.GlobalRegistry.RegisterClass(class); err != nil {
			t.Fatalf("Failed to register class %s: %v", class.Name, err)
		}
	}

	// Get the CachingIterator class
	class := GetCachingIteratorClass()
	if class == nil {
		t.Fatal("CachingIterator class is nil")
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

	testArray := values.NewArray()
	testArray.ArraySet(values.NewInt(0), values.NewString("hello"))
	testArray.ArraySet(values.NewInt(1), values.NewString("world"))
	testArray.ArraySet(values.NewInt(2), values.NewString("test"))

	constructorArgs := []*values.Value{arrayIterValue, testArray}
	arrayIterConstructor := arrayIterClass.Methods["__construct"]
	impl := arrayIterConstructor.Implementation.(*BuiltinMethodImpl)
	_, err := impl.GetFunction().Builtin(ctx, constructorArgs)
	if err != nil {
		t.Fatalf("Failed to initialize ArrayIterator: %v", err)
	}

	// Create CachingIterator instance
	obj := &values.Object{
		ClassName:  "CachingIterator",
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

		// Test with CALL_TOSTRING flag (default)
		args := []*values.Value{thisObj, arrayIterValue, values.NewInt(1)} // CALL_TOSTRING = 1
		impl := constructor.Implementation.(*BuiltinMethodImpl)
		_, err := impl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("Constructor failed: %v", err)
		}

		// Check that iterator was stored
		if iterator, ok := obj.Properties["__iterator"]; !ok || iterator != arrayIterValue {
			t.Fatal("Iterator not stored correctly")
		}

		// Check that flags were stored
		if flags, ok := obj.Properties["__flags"]; !ok || !flags.IsInt() || flags.ToInt() != 1 {
			t.Fatal("Flags not stored correctly")
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

	t.Run("BasicIterationWithCaching", func(t *testing.T) {
		// Create fresh ArrayIterator for this test
		freshArrayIterObj := &values.Object{
			ClassName:  "ArrayIterator",
			Properties: make(map[string]*values.Value),
		}
		freshArrayIterValue := &values.Value{
			Type: values.TypeObject,
			Data: freshArrayIterObj,
		}

		freshTestArray := values.NewArray()
		freshTestArray.ArraySet(values.NewInt(0), values.NewString("hello"))
		freshTestArray.ArraySet(values.NewInt(1), values.NewString("world"))
		freshTestArray.ArraySet(values.NewInt(2), values.NewString("test"))

		constructorArgs := []*values.Value{freshArrayIterValue, freshTestArray}
		arrayIterConstructor := arrayIterClass.Methods["__construct"]
		impl := arrayIterConstructor.Implementation.(*BuiltinMethodImpl)
		_, err := impl.GetFunction().Builtin(ctx, constructorArgs)
		if err != nil {
			t.Fatalf("Failed to initialize fresh ArrayIterator: %v", err)
		}

		// Ensure ArrayIterator starts at position 0
		arrayIterRewind := arrayIterClass.Methods["rewind"]
		arrayIterRewindImpl := arrayIterRewind.Implementation.(*BuiltinMethodImpl)
		_, err = arrayIterRewindImpl.GetFunction().Builtin(ctx, []*values.Value{freshArrayIterValue})
		if err != nil {
			t.Fatalf("Failed to rewind fresh ArrayIterator: %v", err)
		}

		// Create fresh CachingIterator object
		freshCachingObj := &values.Object{
			ClassName:  "CachingIterator",
			Properties: make(map[string]*values.Value),
		}
		freshCachingValue := &values.Value{
			Type: values.TypeObject,
			Data: freshCachingObj,
		}

		// Initialize CachingIterator with fresh ArrayIterator
		constructor := class.Methods["__construct"]
		args := []*values.Value{freshCachingValue, freshArrayIterValue, values.NewInt(CACHING_ITERATOR_CALL_TOSTRING)}
		constructorImpl := constructor.Implementation.(*BuiltinMethodImpl)
		_, err = constructorImpl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("CachingIterator constructor failed: %v", err)
		}

		// Get iteration methods
		rewindMethod := class.Methods["rewind"]
		validMethod := class.Methods["valid"]
		currentMethod := class.Methods["current"]
		keyMethod := class.Methods["key"]
		nextMethod := class.Methods["next"]
		toStringMethod := class.Methods["__toString"]

		rewindImpl := rewindMethod.Implementation.(*BuiltinMethodImpl)
		validImpl := validMethod.Implementation.(*BuiltinMethodImpl)
		currentImpl := currentMethod.Implementation.(*BuiltinMethodImpl)
		keyImpl := keyMethod.Implementation.(*BuiltinMethodImpl)
		nextImpl := nextMethod.Implementation.(*BuiltinMethodImpl)
		toStringImpl := toStringMethod.Implementation.(*BuiltinMethodImpl)

		// Rewind to start
		_, err = rewindImpl.GetFunction().Builtin(ctx, []*values.Value{freshCachingValue})
		if err != nil {
			t.Fatalf("rewind failed: %v", err)
		}

		expectedValues := []string{"hello", "world", "test"}

		for i := 0; i < 3; i++ {
			// Check valid
			validResult, err := validImpl.GetFunction().Builtin(ctx, []*values.Value{freshCachingValue})
			if err != nil {
				t.Fatalf("valid failed at iteration %d: %v", i, err)
			}

			if !validResult.ToBool() {
				t.Fatalf("Should be valid at iteration %d", i)
			}

			// Check current value
			currentResult, err := currentImpl.GetFunction().Builtin(ctx, []*values.Value{freshCachingValue})
			if err != nil {
				t.Fatalf("current failed at iteration %d: %v", i, err)
			}

			if !currentResult.IsString() || currentResult.ToString() != expectedValues[i] {
				t.Fatalf("Expected value '%s' at iteration %d, got %v", expectedValues[i], i, currentResult)
			}

			// Check key
			keyResult, err := keyImpl.GetFunction().Builtin(ctx, []*values.Value{freshCachingValue})
			if err != nil {
				t.Fatalf("key failed at iteration %d: %v", i, err)
			}

			if !keyResult.IsInt() || keyResult.ToInt() != int64(i) {
				t.Fatalf("Expected key %d at iteration %d, got %v", i, i, keyResult)
			}

			// Check __toString (should return current value for CALL_TOSTRING flag)
			toStringResult, err := toStringImpl.GetFunction().Builtin(ctx, []*values.Value{freshCachingValue})
			if err != nil {
				t.Fatalf("__toString failed at iteration %d: %v", i, err)
			}

			if !toStringResult.IsString() || toStringResult.ToString() != expectedValues[i] {
				t.Fatalf("Expected cached string '%s' at iteration %d, got %v", expectedValues[i], i, toStringResult)
			}

			// Move to next (but not on the last iteration to avoid advancing past end)
			if i < 2 {
				_, err = nextImpl.GetFunction().Builtin(ctx, []*values.Value{freshCachingValue})
				if err != nil {
					t.Fatalf("next failed at iteration %d: %v", i, err)
				}
			}
		}

		// Move past the end to make iterator invalid
		_, err = nextImpl.GetFunction().Builtin(ctx, []*values.Value{freshCachingValue})
		if err != nil {
			t.Fatalf("final next failed: %v", err)
		}

		// Should be invalid after all elements consumed
		validResult, err := validImpl.GetFunction().Builtin(ctx, []*values.Value{freshCachingValue})
		if err != nil {
			t.Fatalf("final valid check failed: %v", err)
		}

		if validResult.ToBool() {
			t.Fatal("Should be invalid after all elements consumed")
		}
	})

	t.Run("hasNext", func(t *testing.T) {
		// Create a fresh CachingIterator for this test
		freshObj := &values.Object{
			ClassName:  "CachingIterator",
			Properties: make(map[string]*values.Value),
		}
		freshValue := &values.Value{
			Type: values.TypeObject,
			Data: freshObj,
		}

		// Create fresh ArrayIterator with small data
		freshArrayIterObj := &values.Object{
			ClassName:  "ArrayIterator",
			Properties: make(map[string]*values.Value),
		}
		freshArrayValue := &values.Value{
			Type: values.TypeObject,
			Data: freshArrayIterObj,
		}

		smallArray := values.NewArray()
		smallArray.ArraySet(values.NewInt(0), values.NewString("first"))
		smallArray.ArraySet(values.NewInt(1), values.NewString("second"))

		_, err := impl.GetFunction().Builtin(ctx, []*values.Value{freshArrayValue, smallArray})
		if err != nil {
			t.Fatalf("Failed to initialize fresh ArrayIterator: %v", err)
		}

		// Initialize CachingIterator
		constructor := class.Methods["__construct"]
		constructorImpl := constructor.Implementation.(*BuiltinMethodImpl)
		_, err = constructorImpl.GetFunction().Builtin(ctx, []*values.Value{freshValue, freshArrayValue, values.NewInt(1)})
		if err != nil {
			t.Fatalf("Constructor failed: %v", err)
		}

		// Get hasNext method
		hasNextMethod := class.Methods["hasNext"]
		hasNextImpl := hasNextMethod.Implementation.(*BuiltinMethodImpl)
		rewindMethod := class.Methods["rewind"]
		rewindImpl := rewindMethod.Implementation.(*BuiltinMethodImpl)
		nextMethod := class.Methods["next"]
		nextImpl := nextMethod.Implementation.(*BuiltinMethodImpl)

		// Rewind to start
		_, err = rewindImpl.GetFunction().Builtin(ctx, []*values.Value{freshValue})
		if err != nil {
			t.Fatalf("rewind failed: %v", err)
		}

		// At first element, should have next
		hasNextResult, err := hasNextImpl.GetFunction().Builtin(ctx, []*values.Value{freshValue})
		if err != nil {
			t.Fatalf("hasNext failed at first element: %v", err)
		}

		if !hasNextResult.ToBool() {
			t.Fatal("Should have next at first element")
		}

		// Move to second element
		_, err = nextImpl.GetFunction().Builtin(ctx, []*values.Value{freshValue})
		if err != nil {
			t.Fatalf("next failed: %v", err)
		}

		// At second (last) element, should not have next
		hasNextResult, err = hasNextImpl.GetFunction().Builtin(ctx, []*values.Value{freshValue})
		if err != nil {
			t.Fatalf("hasNext failed at last element: %v", err)
		}

		if hasNextResult.ToBool() {
			t.Fatal("Should not have next at last element")
		}
	})

	t.Run("FullCacheMode", func(t *testing.T) {
		// Create CachingIterator with FULL_CACHE flag
		fullCacheObj := &values.Object{
			ClassName:  "CachingIterator",
			Properties: make(map[string]*values.Value),
		}
		fullCacheValue := &values.Value{
			Type: values.TypeObject,
			Data: fullCacheObj,
		}

		// Initialize with FULL_CACHE flag (value = 2)
		constructor := class.Methods["__construct"]
		constructorImpl := constructor.Implementation.(*BuiltinMethodImpl)
		_, err := constructorImpl.GetFunction().Builtin(ctx, []*values.Value{fullCacheValue, arrayIterValue, values.NewInt(2)})
		if err != nil {
			t.Fatalf("Constructor with FULL_CACHE failed: %v", err)
		}

		// Iterate through all elements to build cache
		rewindMethod := class.Methods["rewind"]
		validMethod := class.Methods["valid"]
		nextMethod := class.Methods["next"]

		rewindImpl := rewindMethod.Implementation.(*BuiltinMethodImpl)
		validImpl := validMethod.Implementation.(*BuiltinMethodImpl)
		nextImpl := nextMethod.Implementation.(*BuiltinMethodImpl)

		_, err = rewindImpl.GetFunction().Builtin(ctx, []*values.Value{fullCacheValue})
		if err != nil {
			t.Fatalf("rewind failed: %v", err)
		}

		// Iterate through all elements
		for {
			validResult, err := validImpl.GetFunction().Builtin(ctx, []*values.Value{fullCacheValue})
			if err != nil {
				t.Fatalf("valid failed: %v", err)
			}

			if !validResult.ToBool() {
				break
			}

			_, err = nextImpl.GetFunction().Builtin(ctx, []*values.Value{fullCacheValue})
			if err != nil {
				t.Fatalf("next failed: %v", err)
			}
		}

		// Check getCache method
		getCacheMethod := class.Methods["getCache"]
		if getCacheMethod != nil {
			getCacheImpl := getCacheMethod.Implementation.(*BuiltinMethodImpl)
			cacheResult, err := getCacheImpl.GetFunction().Builtin(ctx, []*values.Value{fullCacheValue})
			if err != nil {
				t.Fatalf("getCache failed: %v", err)
			}

			if !cacheResult.IsArray() {
				t.Fatal("getCache should return an array")
			}

			cacheArray := cacheResult.Data.(*values.Array)
			if len(cacheArray.Elements) != 3 {
				t.Fatalf("Expected cache size 3, got %d", len(cacheArray.Elements))
			}
		}
	})

	t.Run("EmptyIteratorHandling", func(t *testing.T) {
		// Create empty ArrayIterator
		emptyArrayIterObj := &values.Object{
			ClassName:  "ArrayIterator",
			Properties: make(map[string]*values.Value),
		}
		emptyArrayValue := &values.Value{
			Type: values.TypeObject,
			Data: emptyArrayIterObj,
		}

		emptyArray := values.NewArray()
		_, err := impl.GetFunction().Builtin(ctx, []*values.Value{emptyArrayValue, emptyArray})
		if err != nil {
			t.Fatalf("Failed to initialize empty ArrayIterator: %v", err)
		}

		// Create CachingIterator with empty iterator
		emptyCacheObj := &values.Object{
			ClassName:  "CachingIterator",
			Properties: make(map[string]*values.Value),
		}
		emptyCacheValue := &values.Value{
			Type: values.TypeObject,
			Data: emptyCacheObj,
		}

		constructor := class.Methods["__construct"]
		constructorImpl := constructor.Implementation.(*BuiltinMethodImpl)
		_, err = constructorImpl.GetFunction().Builtin(ctx, []*values.Value{emptyCacheValue, emptyArrayValue, values.NewInt(1)})
		if err != nil {
			t.Fatalf("Constructor with empty iterator failed: %v", err)
		}

		// Should not be valid
		validMethod := class.Methods["valid"]
		validImpl := validMethod.Implementation.(*BuiltinMethodImpl)
		validResult, err := validImpl.GetFunction().Builtin(ctx, []*values.Value{emptyCacheValue})
		if err != nil {
			t.Fatalf("valid on empty caching iterator failed: %v", err)
		}

		if validResult.ToBool() {
			t.Fatal("CachingIterator with empty iterator should not be valid")
		}

		// hasNext should also be false
		hasNextMethod := class.Methods["hasNext"]
		hasNextImpl := hasNextMethod.Implementation.(*BuiltinMethodImpl)
		hasNextResult, err := hasNextImpl.GetFunction().Builtin(ctx, []*values.Value{emptyCacheValue})
		if err != nil {
			t.Fatalf("hasNext on empty caching iterator failed: %v", err)
		}

		if hasNextResult.ToBool() {
			t.Fatal("Empty CachingIterator should not have next")
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
		requiredMethods := []string{"current", "key", "valid", "next", "rewind", "__toString", "hasNext"}

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
}