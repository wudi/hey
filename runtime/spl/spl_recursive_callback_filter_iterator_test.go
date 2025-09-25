package spl

import (
	"testing"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

func TestRecursiveCallbackFilterIterator(t *testing.T) {
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
		class := GetRecursiveCallbackFilterIteratorClass()

		// Test class properties
		if class.Name != "RecursiveCallbackFilterIterator" {
			t.Errorf("Expected class name 'RecursiveCallbackFilterIterator', got '%s'", class.Name)
		}

		if class.Parent != "CallbackFilterIterator" {
			t.Errorf("Expected parent 'CallbackFilterIterator', got '%s'", class.Parent)
		}

		if class.IsAbstract {
			t.Error("RecursiveCallbackFilterIterator should not be abstract")
		}

		// Test interfaces
		expectedInterfaces := []string{"Iterator", "OuterIterator", "RecursiveIterator"}
		if len(class.Interfaces) != len(expectedInterfaces) {
			t.Errorf("Expected %d interfaces, got %d", len(expectedInterfaces), len(class.Interfaces))
		}

		for i, expected := range expectedInterfaces {
			if i >= len(class.Interfaces) || class.Interfaces[i] != expected {
				t.Errorf("Expected interface '%s' at position %d", expected, i)
			}
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

		// Create a mock callback
		callback := values.NewString("callback") // Simplified for testing

		// Test valid constructor
		recursiveCallbackFilterObj := &values.Object{
			ClassName:  "RecursiveCallbackFilterIterator",
			Properties: make(map[string]*values.Value),
		}
		recursiveCallbackFilterThis := &values.Value{
			Type: values.TypeObject,
			Data: recursiveCallbackFilterObj,
		}

		class := GetRecursiveCallbackFilterIteratorClass()
		constructMethod := class.Methods["__construct"]
		constructImpl := constructMethod.Implementation.(*BuiltinMethodImpl)

		// Test with RecursiveArrayIterator and callback
		result, err := constructImpl.GetFunction().Builtin(ctx, []*values.Value{
			recursiveCallbackFilterThis,
			recursiveArrayIter,
			callback,
		})

		if err != nil {
			t.Fatalf("Constructor failed: %v", err)
		}

		if result.Type != values.TypeNull {
			t.Errorf("Expected constructor to return null, got %s", result.Type)
		}

		// Check properties were set
		innerIter := recursiveCallbackFilterObj.Properties["__iterator"]
		if innerIter == nil {
			t.Error("__iterator property should be set")
		}

		callbackProp := recursiveCallbackFilterObj.Properties["__callback"]
		if callbackProp == nil {
			t.Error("__callback property should be set")
		}
	})

	t.Run("HasChildren", func(t *testing.T) {
		// Create test iterator with nested data
		testData := values.NewArray()
		testData.ArraySet(values.NewString("level1"), createTestArray()) // has children
		testData.ArraySet(values.NewString("level2"), values.NewString("simple")) // no children

		recursiveArrayIter := createRecursiveArrayIterator(testData)
		callback := values.NewString("callback")
		recursiveCallbackFilterIter := createRecursiveCallbackFilterIterator(recursiveArrayIter, callback)

		// Get hasChildren method
		class := GetRecursiveCallbackFilterIteratorClass()
		hasChildrenMethod := class.Methods["hasChildren"]
		hasChildrenImpl := hasChildrenMethod.Implementation.(*BuiltinMethodImpl)

		// Test hasChildren
		result, err := hasChildrenImpl.GetFunction().Builtin(ctx, []*values.Value{recursiveCallbackFilterIter})
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
		callback := values.NewString("callback")
		recursiveCallbackFilterIter := createRecursiveCallbackFilterIterator(recursiveArrayIter, callback)

		// Get getChildren method
		class := GetRecursiveCallbackFilterIteratorClass()
		getChildrenMethod := class.Methods["getChildren"]
		getChildrenImpl := getChildrenMethod.Implementation.(*BuiltinMethodImpl)

		// Test getChildren
		result, err := getChildrenImpl.GetFunction().Builtin(ctx, []*values.Value{recursiveCallbackFilterIter})
		if err != nil {
			t.Fatalf("getChildren failed: %v", err)
		}

		if !result.IsObject() {
			t.Fatalf("Expected object result, got %s", result.Type)
		}

		resultObj := result.Data.(*values.Object)
		if resultObj.ClassName != "RecursiveCallbackFilterIterator" {
			t.Errorf("Expected RecursiveCallbackFilterIterator, got %s", resultObj.ClassName)
		}

		// Check that child has same callback
		childCallback := resultObj.Properties["__callback"]
		if childCallback == nil {
			t.Error("Child should inherit callback from parent")
		}
	})

	t.Run("InheritedMethods", func(t *testing.T) {
		class := GetRecursiveCallbackFilterIteratorClass()

		// Test inherited methods exist from CallbackFilterIterator
		inheritedMethods := []string{"accept", "current", "key", "next", "rewind", "valid", "getInnerIterator"}
		for _, methodName := range inheritedMethods {
			if _, exists := class.Methods[methodName]; !exists {
				t.Errorf("Inherited method '%s' should exist", methodName)
			}
		}
	})
}

// Helper function to create RecursiveCallbackFilterIterator for testing
func createRecursiveCallbackFilterIterator(innerIter *values.Value, callback *values.Value) *values.Value {
	recursiveCallbackFilterObj := &values.Object{
		ClassName:  "RecursiveCallbackFilterIterator",
		Properties: make(map[string]*values.Value),
	}
	recursiveCallbackFilterObj.Properties["__iterator"] = innerIter
	recursiveCallbackFilterObj.Properties["__callback"] = callback

	return &values.Value{
		Type: values.TypeObject,
		Data: recursiveCallbackFilterObj,
	}
}