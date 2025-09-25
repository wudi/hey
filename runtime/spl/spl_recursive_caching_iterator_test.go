package spl

import (
	"testing"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)


func TestRecursiveCachingIterator(t *testing.T) {
	registry.Initialize()

	// Register SPL classes for testing
	for _, class := range GetSplClasses() {
		err := registry.GlobalRegistry.RegisterClass(class)
		if err != nil {
			t.Fatalf("Failed to register class %s: %v", class.Name, err)
		}
	}

	// Register SPL interfaces for testing
	for _, iface := range GetSplInterfaces() {
		err := registry.GlobalRegistry.RegisterInterface(iface)
		if err != nil {
			t.Fatalf("Failed to register interface %s: %v", iface.Name, err)
		}
	}

	ctx := &mockContext{registry: registry.GlobalRegistry}

	t.Run("ClassStructure", func(t *testing.T) {
		class := GetRecursiveCachingIteratorClass()

		// Test class properties
		if class.Name != "RecursiveCachingIterator" {
			t.Errorf("Expected class name 'RecursiveCachingIterator', got '%s'", class.Name)
		}

		if class.Parent != "CachingIterator" {
			t.Errorf("Expected parent 'CachingIterator', got '%s'", class.Parent)
		}

		if class.IsAbstract {
			t.Error("RecursiveCachingIterator should not be abstract")
		}

		// Test interfaces
		expectedInterfaces := []string{"Iterator", "OuterIterator", "RecursiveIterator", "Countable", "ArrayAccess", "Stringable"}
		if len(class.Interfaces) != len(expectedInterfaces) {
			t.Errorf("Expected %d interfaces, got %d", len(expectedInterfaces), len(class.Interfaces))
		}

		// Test required methods exist
		requiredMethods := []string{"__construct", "hasChildren", "getChildren"}
		for _, methodName := range requiredMethods {
			if _, exists := class.Methods[methodName]; !exists {
				t.Errorf("Method '%s' should exist", methodName)
			}
		}
	})

	t.Run("Constructor", func(t *testing.T) {
		// Create test data
		testData := values.NewArray()
		testData.ArraySet(values.NewString("level1"), createTestArray())
		testData.ArraySet(values.NewString("level2"), values.NewString("simple"))

		// Create RecursiveArrayIterator
		recursiveArrayIter := createRecursiveArrayIterator(testData)

		// Test valid constructor
		recursiveCachingObj := &values.Object{
			ClassName:  "RecursiveCachingIterator",
			Properties: make(map[string]*values.Value),
		}
		recursiveCachingThis := &values.Value{
			Type: values.TypeObject,
			Data: recursiveCachingObj,
		}

		class := GetRecursiveCachingIteratorClass()
		constructMethod := class.Methods["__construct"]
		constructImpl := constructMethod.Implementation.(*BuiltinMethodImpl)

		// Test with RecursiveArrayIterator
		result, err := constructImpl.GetFunction().Builtin(ctx, []*values.Value{
			recursiveCachingThis,
			recursiveArrayIter,
			values.NewInt(2), // FULL_CACHE
		})

		if err != nil {
			t.Fatalf("Constructor failed: %v", err)
		}

		if result.Type != values.TypeNull {
			t.Errorf("Expected constructor to return null, got %s", result.Type)
		}

		// Check properties were set
		innerIter := recursiveCachingObj.Properties["__iterator"]
		if innerIter == nil {
			t.Error("__iterator property should be set")
		}

		flags := recursiveCachingObj.Properties["__flags"]
		if flags == nil || flags.Type != values.TypeInt || flags.Data.(int64) != 2 {
			t.Error("__flags property should be set to 2")
		}

		cache := recursiveCachingObj.Properties["__cache"]
		if cache == nil || cache.Type != values.TypeArray {
			t.Error("__cache property should be set as array")
		}
	})

	t.Run("HasChildren", func(t *testing.T) {
		// Create test iterator with nested data
		testData := values.NewArray()
		testData.ArraySet(values.NewString("level1"), createTestArray()) // has children
		testData.ArraySet(values.NewString("level2"), values.NewString("simple")) // no children

		recursiveArrayIter := createRecursiveArrayIterator(testData)
		recursiveCachingIter := createRecursiveCachingIterator(recursiveArrayIter, 1)

		// Get hasChildren method
		class := GetRecursiveCachingIteratorClass()
		hasChildrenMethod := class.Methods["hasChildren"]
		hasChildrenImpl := hasChildrenMethod.Implementation.(*BuiltinMethodImpl)

		// Test hasChildren
		result, err := hasChildrenImpl.GetFunction().Builtin(ctx, []*values.Value{recursiveCachingIter})
		if err != nil {
			t.Fatalf("hasChildren failed: %v", err)
		}

		if result.Type != values.TypeBool {
			t.Errorf("Expected bool result, got %s", result.Type)
		}
	})

	t.Run("GetChildren", func(t *testing.T) {
		// Create test iterator with nested data
		testData := values.NewArray()
		childData := createTestArray()
		testData.ArraySet(values.NewString("level1"), childData)

		recursiveArrayIter := createRecursiveArrayIterator(testData)
		recursiveCachingIter := createRecursiveCachingIterator(recursiveArrayIter, 2) // FULL_CACHE

		// Get getChildren method
		class := GetRecursiveCachingIteratorClass()
		getChildrenMethod := class.Methods["getChildren"]
		getChildrenImpl := getChildrenMethod.Implementation.(*BuiltinMethodImpl)

		// Test getChildren
		result, err := getChildrenImpl.GetFunction().Builtin(ctx, []*values.Value{recursiveCachingIter})
		if err != nil {
			t.Fatalf("getChildren failed: %v", err)
		}

		if !result.IsObject() {
			t.Fatalf("Expected object result, got %s", result.Type)
		}

		resultObj := result.Data.(*values.Object)
		if resultObj.ClassName != "RecursiveCachingIterator" {
			t.Errorf("Expected RecursiveCachingIterator, got %s", resultObj.ClassName)
		}

		// Check that child has same flags
		childFlags := resultObj.Properties["__flags"]
		if childFlags == nil || childFlags.Type != values.TypeInt || childFlags.Data.(int64) != 2 {
			t.Error("Child should inherit flags from parent")
		}
	})
}

// Helper functions
func createTestArray() *values.Value {
	arr := values.NewArray()
	arr.ArraySet(values.NewString("item1"), values.NewString("value1"))
	arr.ArraySet(values.NewString("item2"), values.NewString("value2"))
	return arr
}

func createRecursiveArrayIterator(data *values.Value) *values.Value {
	recursiveArrayObj := &values.Object{
		ClassName:  "RecursiveArrayIterator",
		Properties: make(map[string]*values.Value),
	}
	recursiveArrayObj.Properties["__array"] = data
	recursiveArrayObj.Properties["__flags"] = values.NewInt(0)
	recursiveArrayObj.Properties["__position"] = values.NewInt(0)

	// Build keys array for iteration in sorted order (same as ArrayIterator)
	keys := values.NewArray()
	if data.IsArray() {
		arr := data.Data.(*values.Array)

		// Collect all keys first
		var stringKeys []string

		for k := range arr.Elements {
			switch v := k.(type) {
			case string:
				stringKeys = append(stringKeys, v)
			}
		}

		// Add sorted keys to keys array
		index := 0
		for _, key := range stringKeys {
			keys.ArraySet(values.NewInt(int64(index)), values.NewString(key))
			index++
		}
	}
	recursiveArrayObj.Properties["__keys"] = keys

	return &values.Value{
		Type: values.TypeObject,
		Data: recursiveArrayObj,
	}
}

func createRecursiveCachingIterator(innerIter *values.Value, flags int64) *values.Value {
	recursiveCachingObj := &values.Object{
		ClassName:  "RecursiveCachingIterator",
		Properties: make(map[string]*values.Value),
	}
	recursiveCachingObj.Properties["__iterator"] = innerIter
	recursiveCachingObj.Properties["__flags"] = values.NewInt(flags)
	recursiveCachingObj.Properties["__cache"] = values.NewArray()

	return &values.Value{
		Type: values.TypeObject,
		Data: recursiveCachingObj,
	}
}