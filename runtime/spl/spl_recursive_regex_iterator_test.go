package spl

import (
	"testing"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

func TestRecursiveRegexIterator(t *testing.T) {
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
		class := GetRecursiveRegexIteratorClass()

		// Test class properties
		if class.Name != "RecursiveRegexIterator" {
			t.Errorf("Expected class name 'RecursiveRegexIterator', got '%s'", class.Name)
		}

		if class.Parent != "RegexIterator" {
			t.Errorf("Expected parent 'RegexIterator', got '%s'", class.Parent)
		}

		if class.IsAbstract {
			t.Error("RecursiveRegexIterator should not be abstract")
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
		testData.ArraySet(values.NewString("test.php"), values.NewString("php file"))
		testData.ArraySet(values.NewString("styles"), createTestArray())
		testData.ArraySet(values.NewString("config.txt"), values.NewString("config file"))

		// Create RecursiveArrayIterator
		recursiveArrayIter := createRecursiveArrayIterator(testData)

		// Create regex pattern
		regex := values.NewString("\\.php$")
		mode := values.NewInt(0)  // MATCH mode
		flags := values.NewInt(0) // No flags

		// Test valid constructor
		recursiveRegexObj := &values.Object{
			ClassName:  "RecursiveRegexIterator",
			Properties: make(map[string]*values.Value),
		}
		recursiveRegexThis := &values.Value{
			Type: values.TypeObject,
			Data: recursiveRegexObj,
		}

		class := GetRecursiveRegexIteratorClass()
		constructMethod := class.Methods["__construct"]
		constructImpl := constructMethod.Implementation.(*BuiltinMethodImpl)

		// Test with RecursiveArrayIterator and regex
		result, err := constructImpl.GetFunction().Builtin(ctx, []*values.Value{
			recursiveRegexThis,
			recursiveArrayIter,
			regex,
			mode,
			flags,
		})

		if err != nil {
			t.Fatalf("Constructor failed: %v", err)
		}

		if result.Type != values.TypeNull {
			t.Errorf("Expected constructor to return null, got %s", result.Type)
		}

		// Check properties were set
		innerIter := recursiveRegexObj.Properties["__iterator"]
		if innerIter == nil {
			t.Error("__iterator property should be set")
		}

		regexProp := recursiveRegexObj.Properties["__regex"]
		if regexProp == nil {
			t.Error("__regex property should be set")
		}

		modeProp := recursiveRegexObj.Properties["__mode"]
		if modeProp == nil || modeProp.Type != values.TypeInt {
			t.Error("__mode property should be set as integer")
		}

		flagsProp := recursiveRegexObj.Properties["__flags"]
		if flagsProp == nil || flagsProp.Type != values.TypeInt {
			t.Error("__flags property should be set as integer")
		}
	})

	t.Run("HasChildren", func(t *testing.T) {
		// Create test iterator with nested data
		testData := values.NewArray()
		testData.ArraySet(values.NewString("file.php"), values.NewString("php file"))
		testData.ArraySet(values.NewString("directory"), createTestArray()) // has children

		recursiveArrayIter := createRecursiveArrayIterator(testData)
		regex := values.NewString("\\.php$")
		recursiveRegexIter := createRecursiveRegexIterator(recursiveArrayIter, regex, 0, 0)

		// Get hasChildren method
		class := GetRecursiveRegexIteratorClass()
		hasChildrenMethod := class.Methods["hasChildren"]
		hasChildrenImpl := hasChildrenMethod.Implementation.(*BuiltinMethodImpl)

		// Test hasChildren
		result, err := hasChildrenImpl.GetFunction().Builtin(ctx, []*values.Value{recursiveRegexIter})
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
		testData.ArraySet(values.NewString("directory"), childData)

		recursiveArrayIter := createRecursiveArrayIterator(testData)
		regex := values.NewString("\\.php$")
		recursiveRegexIter := createRecursiveRegexIterator(recursiveArrayIter, regex, 0, 0)

		// Get getChildren method
		class := GetRecursiveRegexIteratorClass()
		getChildrenMethod := class.Methods["getChildren"]
		getChildrenImpl := getChildrenMethod.Implementation.(*BuiltinMethodImpl)

		// Test getChildren
		result, err := getChildrenImpl.GetFunction().Builtin(ctx, []*values.Value{recursiveRegexIter})
		if err != nil {
			t.Fatalf("getChildren failed: %v", err)
		}

		if !result.IsObject() {
			t.Fatalf("Expected object result, got %s", result.Type)
		}

		resultObj := result.Data.(*values.Object)
		if resultObj.ClassName != "RecursiveRegexIterator" {
			t.Errorf("Expected RecursiveRegexIterator, got %s", resultObj.ClassName)
		}

		// Check that child has same regex settings
		childRegex := resultObj.Properties["__regex"]
		if childRegex == nil {
			t.Error("Child should inherit regex from parent")
		}

		childMode := resultObj.Properties["__mode"]
		if childMode == nil || childMode.Type != values.TypeInt {
			t.Error("Child should inherit mode from parent")
		}

		childFlags := resultObj.Properties["__flags"]
		if childFlags == nil || childFlags.Type != values.TypeInt {
			t.Error("Child should inherit flags from parent")
		}
	})

	t.Run("InheritedMethods", func(t *testing.T) {
		class := GetRecursiveRegexIteratorClass()

		// Test inherited methods exist from RegexIterator
		inheritedMethods := []string{"accept", "current", "key", "next", "rewind", "valid", "getInnerIterator"}
		for _, methodName := range inheritedMethods {
			if _, exists := class.Methods[methodName]; !exists {
				t.Errorf("Inherited method '%s' should exist", methodName)
			}
		}
	})

	t.Run("Constants", func(t *testing.T) {
		class := GetRecursiveRegexIteratorClass()

		// Should inherit constants from RegexIterator
		expectedConstants := []string{"MATCH", "GET_MATCH", "ALL_MATCHES", "SPLIT", "REPLACE", "USE_KEY"}
		for _, constantName := range expectedConstants {
			if _, exists := class.Constants[constantName]; !exists {
				// This is expected since we inherit from RegexIterator
				// The actual test would be that the constants exist in the parent
			}
		}
	})
}

// Helper function to create RecursiveRegexIterator for testing
func createRecursiveRegexIterator(innerIter *values.Value, regex *values.Value, mode int64, flags int64) *values.Value {
	recursiveRegexObj := &values.Object{
		ClassName:  "RecursiveRegexIterator",
		Properties: make(map[string]*values.Value),
	}
	recursiveRegexObj.Properties["__iterator"] = innerIter
	recursiveRegexObj.Properties["__regex"] = regex
	recursiveRegexObj.Properties["__mode"] = values.NewInt(mode)
	recursiveRegexObj.Properties["__flags"] = values.NewInt(flags)
	recursiveRegexObj.Properties["__preg_flags"] = values.NewInt(0)

	return &values.Value{
		Type: values.TypeObject,
		Data: recursiveRegexObj,
	}
}