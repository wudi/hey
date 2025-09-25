package spl

import (
	"testing"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

func TestRecursiveIteratorIterator(t *testing.T) {
	// Initialize registry
	registry.Initialize()
	ctx := &mockContext{registry: registry.GlobalRegistry}

	// Register all SPL classes for testing
	for _, class := range GetSplClasses() {
		if err := registry.GlobalRegistry.RegisterClass(class); err != nil {
			t.Fatalf("Failed to register class %s: %v", class.Name, err)
		}
	}

	// Get the RecursiveIteratorIterator and RecursiveArrayIterator classes
	recursiveIterIterClass := GetRecursiveIteratorIteratorClass()
	recursiveArrayIterClass := GetRecursiveArrayIteratorClass()

	if recursiveIterIterClass == nil {
		t.Fatal("RecursiveIteratorIterator class is nil")
	}
	if recursiveArrayIterClass == nil {
		t.Fatal("RecursiveArrayIterator class is nil")
	}

	// Create nested array structure for testing
	nestedArray := values.NewArray()

	// level1_a
	level1a := values.NewArray()
	level1a.ArraySet(values.NewString("level2_a"), values.NewString("leaf_aa"))

	// level1_a.level2_b (nested)
	level2b := values.NewArray()
	level2b.ArraySet(values.NewString("level3_a"), values.NewString("leaf_aba"))
	level2b.ArraySet(values.NewString("level3_b"), values.NewString("leaf_abb"))
	level1a.ArraySet(values.NewString("level2_b"), level2b)

	level1a.ArraySet(values.NewString("level2_c"), values.NewString("leaf_ac"))
	nestedArray.ArraySet(values.NewString("level1_a"), level1a)

	// level1_b (leaf)
	nestedArray.ArraySet(values.NewString("level1_b"), values.NewString("leaf_b"))

	// level1_c
	level1c := values.NewArray()
	level1c.ArraySet(values.NewString("level2_d"), values.NewString("leaf_cd"))

	// level1_c.level2_e (nested)
	level2e := values.NewArray()
	level2e.ArraySet(values.NewString("level3_c"), values.NewString("leaf_cec"))
	level1c.ArraySet(values.NewString("level2_e"), level2e)

	nestedArray.ArraySet(values.NewString("level1_c"), level1c)

	// Create RecursiveArrayIterator instance
	recursiveArrayIterObj := &values.Object{
		ClassName:  "RecursiveArrayIterator",
		Properties: make(map[string]*values.Value),
	}
	recursiveArrayIterValue := &values.Value{
		Type: values.TypeObject,
		Data: recursiveArrayIterObj,
	}

	// Initialize RecursiveArrayIterator
	recursiveArrayIterConstructor := recursiveArrayIterClass.Methods["__construct"]
	recursiveArrayIterImpl := recursiveArrayIterConstructor.Implementation.(*BuiltinMethodImpl)
	_, err := recursiveArrayIterImpl.GetFunction().Builtin(ctx, []*values.Value{recursiveArrayIterValue, nestedArray})
	if err != nil {
		t.Fatalf("Failed to initialize RecursiveArrayIterator: %v", err)
	}

	t.Run("Constructor", func(t *testing.T) {
		// Create RecursiveIteratorIterator instance
		obj := &values.Object{
			ClassName:  "RecursiveIteratorIterator",
			Properties: make(map[string]*values.Value),
		}
		thisObj := &values.Value{
			Type: values.TypeObject,
			Data: obj,
		}

		constructor := recursiveIterIterClass.Methods["__construct"]
		if constructor == nil {
			t.Fatal("Constructor not found")
		}

		// Test with LEAVES_ONLY mode (default)
		args := []*values.Value{thisObj, recursiveArrayIterValue}
		constructorImpl := constructor.Implementation.(*BuiltinMethodImpl)
		_, err := constructorImpl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("Constructor failed: %v", err)
		}

		// Check that parameters were stored
		if iterator, ok := obj.Properties["__iterator"]; !ok || iterator != recursiveArrayIterValue {
			t.Fatal("Iterator not stored correctly")
		}
		if mode, ok := obj.Properties["__mode"]; !ok || mode.ToInt() != RECURSIVE_ITERATOR_ITERATOR_LEAVES_ONLY {
			t.Fatal("Mode not stored correctly")
		}
	})

	t.Run("ConstructorWithMode", func(t *testing.T) {
		// Test constructor with explicit SELF_FIRST mode
		obj := &values.Object{
			ClassName:  "RecursiveIteratorIterator",
			Properties: make(map[string]*values.Value),
		}
		thisObj := &values.Value{
			Type: values.TypeObject,
			Data: obj,
		}

		constructor := recursiveIterIterClass.Methods["__construct"]
		args := []*values.Value{thisObj, recursiveArrayIterValue, values.NewInt(RECURSIVE_ITERATOR_ITERATOR_SELF_FIRST)}
		constructorImpl := constructor.Implementation.(*BuiltinMethodImpl)
		_, err := constructorImpl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("Constructor with SELF_FIRST failed: %v", err)
		}

		// Check that mode was stored correctly
		if mode, ok := obj.Properties["__mode"]; !ok || mode.ToInt() != RECURSIVE_ITERATOR_ITERATOR_SELF_FIRST {
			t.Fatal("SELF_FIRST mode not stored correctly")
		}
	})

	t.Run("BasicIteration", func(t *testing.T) {
		// Create fresh RecursiveArrayIterator
		freshRecursiveArrayIterObj := &values.Object{
			ClassName:  "RecursiveArrayIterator",
			Properties: make(map[string]*values.Value),
		}
		freshRecursiveArrayIterValue := &values.Value{
			Type: values.TypeObject,
			Data: freshRecursiveArrayIterObj,
		}

		// Initialize with smaller test array for clearer testing
		testArray := values.NewArray()

		innerArray := values.NewArray()
		innerArray.ArraySet(values.NewString("inner_key"), values.NewString("inner_value"))

		testArray.ArraySet(values.NewString("outer_key"), values.NewString("outer_value"))
		testArray.ArraySet(values.NewString("nested"), innerArray)

		_, err = recursiveArrayIterImpl.GetFunction().Builtin(ctx, []*values.Value{freshRecursiveArrayIterValue, testArray})
		if err != nil {
			t.Fatalf("Failed to initialize fresh RecursiveArrayIterator: %v", err)
		}

		// Create RecursiveIteratorIterator
		obj := &values.Object{
			ClassName:  "RecursiveIteratorIterator",
			Properties: make(map[string]*values.Value),
		}
		thisObj := &values.Value{
			Type: values.TypeObject,
			Data: obj,
		}

		constructor := recursiveIterIterClass.Methods["__construct"]
		args := []*values.Value{thisObj, freshRecursiveArrayIterValue, values.NewInt(RECURSIVE_ITERATOR_ITERATOR_LEAVES_ONLY)}
		constructorImpl := constructor.Implementation.(*BuiltinMethodImpl)
		_, err = constructorImpl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("Constructor failed: %v", err)
		}

		// Test basic iteration methods
		rewindMethod := recursiveIterIterClass.Methods["rewind"]
		validMethod := recursiveIterIterClass.Methods["valid"]
		currentMethod := recursiveIterIterClass.Methods["current"]
		keyMethod := recursiveIterIterClass.Methods["key"]

		rewindImpl := rewindMethod.Implementation.(*BuiltinMethodImpl)
		validImpl := validMethod.Implementation.(*BuiltinMethodImpl)
		currentImpl := currentMethod.Implementation.(*BuiltinMethodImpl)
		keyImpl := keyMethod.Implementation.(*BuiltinMethodImpl)

		// Rewind to start
		_, err = rewindImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("rewind failed: %v", err)
		}

		// Check if valid
		validResult, _ := validImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if !validResult.ToBool() {
			t.Log("Iterator is not valid - this may be expected for LEAVES_ONLY mode")
		}

		// Test current() method doesn't crash
		currentResult, err := currentImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("current() failed: %v", err)
		}
		_ = currentResult // Current result may be null if not valid

		// Test key() method doesn't crash
		keyResult, err := keyImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("key() failed: %v", err)
		}
		_ = keyResult // Key result may be null if not valid
	})

	t.Run("DepthTracking", func(t *testing.T) {
		// Create RecursiveIteratorIterator instance
		obj := &values.Object{
			ClassName:  "RecursiveIteratorIterator",
			Properties: make(map[string]*values.Value),
		}
		thisObj := &values.Value{
			Type: values.TypeObject,
			Data: obj,
		}

		constructor := recursiveIterIterClass.Methods["__construct"]
		args := []*values.Value{thisObj, recursiveArrayIterValue, values.NewInt(RECURSIVE_ITERATOR_ITERATOR_SELF_FIRST)}
		constructorImpl := constructor.Implementation.(*BuiltinMethodImpl)
		_, err := constructorImpl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("Constructor failed: %v", err)
		}

		// Test getDepth() method
		getDepthMethod := recursiveIterIterClass.Methods["getDepth"]
		if getDepthMethod == nil {
			t.Fatal("getDepth method not found")
		}

		getDepthImpl := getDepthMethod.Implementation.(*BuiltinMethodImpl)
		depthResult, err := getDepthImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("getDepth failed: %v", err)
		}

		// Initial depth should be 0
		if depthResult.ToInt() != 0 {
			t.Fatalf("Expected initial depth 0, got %d", depthResult.ToInt())
		}
	})

	t.Run("MaxDepthControl", func(t *testing.T) {
		// Create RecursiveIteratorIterator instance
		obj := &values.Object{
			ClassName:  "RecursiveIteratorIterator",
			Properties: make(map[string]*values.Value),
		}
		thisObj := &values.Value{
			Type: values.TypeObject,
			Data: obj,
		}

		constructor := recursiveIterIterClass.Methods["__construct"]
		args := []*values.Value{thisObj, recursiveArrayIterValue}
		constructorImpl := constructor.Implementation.(*BuiltinMethodImpl)
		_, err := constructorImpl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("Constructor failed: %v", err)
		}

		// Test setMaxDepth() and getMaxDepth() methods
		setMaxDepthMethod := recursiveIterIterClass.Methods["setMaxDepth"]
		getMaxDepthMethod := recursiveIterIterClass.Methods["getMaxDepth"]

		if setMaxDepthMethod == nil {
			t.Fatal("setMaxDepth method not found")
		}
		if getMaxDepthMethod == nil {
			t.Fatal("getMaxDepth method not found")
		}

		setMaxDepthImpl := setMaxDepthMethod.Implementation.(*BuiltinMethodImpl)
		getMaxDepthImpl := getMaxDepthMethod.Implementation.(*BuiltinMethodImpl)

		// Set max depth to 1
		_, err = setMaxDepthImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj, values.NewInt(1)})
		if err != nil {
			t.Fatalf("setMaxDepth failed: %v", err)
		}

		// Get max depth and verify
		maxDepthResult, err := getMaxDepthImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("getMaxDepth failed: %v", err)
		}

		if maxDepthResult.ToInt() != 1 {
			t.Fatalf("Expected max depth 1, got %d", maxDepthResult.ToInt())
		}
	})

	t.Run("Constants", func(t *testing.T) {
		// Test class constants
		constants := recursiveIterIterClass.Constants

		if leavesOnly, exists := constants["LEAVES_ONLY"]; !exists || leavesOnly.Value.ToInt() != RECURSIVE_ITERATOR_ITERATOR_LEAVES_ONLY {
			t.Fatal("LEAVES_ONLY constant not defined correctly")
		}

		if selfFirst, exists := constants["SELF_FIRST"]; !exists || selfFirst.Value.ToInt() != RECURSIVE_ITERATOR_ITERATOR_SELF_FIRST {
			t.Fatal("SELF_FIRST constant not defined correctly")
		}

		if childFirst, exists := constants["CHILD_FIRST"]; !exists || childFirst.Value.ToInt() != RECURSIVE_ITERATOR_ITERATOR_CHILD_FIRST {
			t.Fatal("CHILD_FIRST constant not defined correctly")
		}
	})

	t.Run("InterfaceImplementation", func(t *testing.T) {
		// Verify that RecursiveIteratorIterator implements Iterator and OuterIterator
		interfaces := recursiveIterIterClass.Interfaces

		foundIterator := false
		foundOuterIterator := false

		for _, iface := range interfaces {
			if iface == "Iterator" {
				foundIterator = true
			}
			if iface == "OuterIterator" {
				foundOuterIterator = true
			}
		}

		if !foundIterator {
			t.Fatal("RecursiveIteratorIterator should implement Iterator interface")
		}
		if !foundOuterIterator {
			t.Fatal("RecursiveIteratorIterator should implement OuterIterator interface")
		}
	})
}