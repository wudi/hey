package spl

import (
	"testing"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

func TestRecursiveTreeIterator(t *testing.T) {
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
		class := GetRecursiveTreeIteratorClass()

		// Test class properties
		if class.Name != "RecursiveTreeIterator" {
			t.Errorf("Expected class name 'RecursiveTreeIterator', got '%s'", class.Name)
		}

		if class.Parent != "RecursiveIteratorIterator" {
			t.Errorf("Expected parent 'RecursiveIteratorIterator', got '%s'", class.Parent)
		}

		if class.IsAbstract {
			t.Error("RecursiveTreeIterator should not be abstract")
		}

		// Test interfaces
		expectedInterfaces := []string{"Iterator", "OuterIterator"}
		if len(class.Interfaces) != len(expectedInterfaces) {
			t.Errorf("Expected %d interfaces, got %d", len(expectedInterfaces), len(class.Interfaces))
		}

		for i, expected := range expectedInterfaces {
			if i >= len(class.Interfaces) || class.Interfaces[i] != expected {
				t.Errorf("Expected interface '%s' at position %d", expected, i)
			}
		}

		// Test required methods exist
		requiredMethods := []string{"__construct", "current", "getEntry", "getPrefix", "getPostfix", "setPostfix", "setPrefixPart"}
		for _, methodName := range requiredMethods {
			if _, exists := class.Methods[methodName]; !exists {
				t.Errorf("Method '%s' should exist", methodName)
			}
		}
	})

	t.Run("Constructor", func(t *testing.T) {
		// Create test data
		testData := values.NewArray()
		testData.ArraySet(values.NewString("folder1"), createTestArray())
		testData.ArraySet(values.NewString("file.txt"), values.NewString("content"))

		// Create RecursiveArrayIterator
		recursiveArrayIter := createRecursiveArrayIterator(testData)

		// Test valid constructor
		recursiveTreeObj := &values.Object{
			ClassName:  "RecursiveTreeIterator",
			Properties: make(map[string]*values.Value),
		}
		recursiveTreeThis := &values.Value{
			Type: values.TypeObject,
			Data: recursiveTreeObj,
		}

		class := GetRecursiveTreeIteratorClass()
		constructMethod := class.Methods["__construct"]
		constructImpl := constructMethod.Implementation.(*BuiltinMethodImpl)

		// Test with RecursiveArrayIterator
		result, err := constructImpl.GetFunction().Builtin(ctx, []*values.Value{
			recursiveTreeThis,
			recursiveArrayIter,
			values.NewInt(1), // SELF_FIRST
		})

		if err != nil {
			t.Fatalf("Constructor failed: %v", err)
		}

		if result.Type != values.TypeNull {
			t.Errorf("Expected constructor to return null, got %s", result.Type)
		}

		// Check properties were set
		innerIter := recursiveTreeObj.Properties["__iterator"]
		if innerIter == nil {
			t.Error("__iterator property should be set")
		}

		flags := recursiveTreeObj.Properties["__flags"]
		if flags == nil || flags.Type != values.TypeInt {
			t.Error("__flags property should be set as integer")
		}

		prefixes := recursiveTreeObj.Properties["__prefixes"]
		if prefixes == nil || !prefixes.IsArray() {
			t.Error("__prefixes property should be set as array")
		}
	})

	t.Run("GetEntry", func(t *testing.T) {
		// Create test iterator
		testData := values.NewArray()
		testData.ArraySet(values.NewString("test_file"), values.NewString("content"))

		recursiveArrayIter := createRecursiveArrayIterator(testData)
		recursiveTreeIter := createRecursiveTreeIterator(recursiveArrayIter, 1)

		// Get getEntry method
		class := GetRecursiveTreeIteratorClass()
		getEntryMethod := class.Methods["getEntry"]
		getEntryImpl := getEntryMethod.Implementation.(*BuiltinMethodImpl)

		// Test getEntry
		result, err := getEntryImpl.GetFunction().Builtin(ctx, []*values.Value{recursiveTreeIter})
		if err != nil {
			t.Fatalf("getEntry failed: %v", err)
		}

		if result.Type != values.TypeString {
			t.Errorf("Expected string result, got %s", result.Type)
		}
	})

	t.Run("GetPrefix", func(t *testing.T) {
		// Create test iterator
		testData := values.NewArray()
		testData.ArraySet(values.NewString("test_file"), values.NewString("content"))

		recursiveArrayIter := createRecursiveArrayIterator(testData)
		recursiveTreeIter := createRecursiveTreeIterator(recursiveArrayIter, 1)

		// Get getPrefix method
		class := GetRecursiveTreeIteratorClass()
		getPrefixMethod := class.Methods["getPrefix"]
		getPrefixImpl := getPrefixMethod.Implementation.(*BuiltinMethodImpl)

		// Test getPrefix
		result, err := getPrefixImpl.GetFunction().Builtin(ctx, []*values.Value{recursiveTreeIter})
		if err != nil {
			t.Fatalf("getPrefix failed: %v", err)
		}

		if result.Type != values.TypeString {
			t.Errorf("Expected string result, got %s", result.Type)
		}

		// Should return some tree prefix like "|-"
		prefix := result.ToString()
		if prefix == "" {
			t.Error("Prefix should not be empty")
		}
	})

	t.Run("SetPostfix", func(t *testing.T) {
		// Create test iterator
		testData := values.NewArray()
		recursiveArrayIter := createRecursiveArrayIterator(testData)
		recursiveTreeIter := createRecursiveTreeIterator(recursiveArrayIter, 1)

		// Get setPostfix method
		class := GetRecursiveTreeIteratorClass()
		setPostfixMethod := class.Methods["setPostfix"]
		setPostfixImpl := setPostfixMethod.Implementation.(*BuiltinMethodImpl)

		// Test setPostfix
		result, err := setPostfixImpl.GetFunction().Builtin(ctx, []*values.Value{
			recursiveTreeIter,
			values.NewString(" <--"),
		})

		if err != nil {
			t.Fatalf("setPostfix failed: %v", err)
		}

		if result.Type != values.TypeNull {
			t.Errorf("Expected null result, got %s", result.Type)
		}

		// Verify postfix was set
		objData := recursiveTreeIter.Data.(*values.Object)
		postfix := objData.Properties["__postfix"]
		if postfix == nil || postfix.ToString() != " <--" {
			t.Error("Postfix should be set to ' <--'")
		}
	})

	t.Run("SetPrefixPart", func(t *testing.T) {
		// Create test iterator
		testData := values.NewArray()
		recursiveArrayIter := createRecursiveArrayIterator(testData)
		recursiveTreeIter := createRecursiveTreeIterator(recursiveArrayIter, 1)

		// Get setPrefixPart method
		class := GetRecursiveTreeIteratorClass()
		setPrefixPartMethod := class.Methods["setPrefixPart"]
		setPrefixPartImpl := setPrefixPartMethod.Implementation.(*BuiltinMethodImpl)

		// Test setPrefixPart
		result, err := setPrefixPartImpl.GetFunction().Builtin(ctx, []*values.Value{
			recursiveTreeIter,
			values.NewInt(1), // PREFIX_MID_HAS_NEXT
			values.NewString(">> "),
		})

		if err != nil {
			t.Fatalf("setPrefixPart failed: %v", err)
		}

		if result.Type != values.TypeNull {
			t.Errorf("Expected null result, got %s", result.Type)
		}

		// Verify prefix part was set
		objData := recursiveTreeIter.Data.(*values.Object)
		prefixes := objData.Properties["__prefixes"]
		if prefixes != nil && prefixes.IsArray() {
			updatedPrefix := prefixes.ArrayGet(values.NewInt(1))
			if updatedPrefix == nil || updatedPrefix.ToString() != ">> " {
				t.Error("Prefix part should be updated to '>> '")
			}
		}
	})

	t.Run("Constants", func(t *testing.T) {
		class := GetRecursiveTreeIteratorClass()

		// Test tree-specific constants exist
		expectedConstants := []string{"BYPASS_CURRENT", "BYPASS_KEY", "SHOW_TREE"}
		for _, constantName := range expectedConstants {
			if _, exists := class.Constants[constantName]; !exists {
				t.Errorf("Constant '%s' should exist", constantName)
			}
		}

		// Verify constant values
		if bypassCurrent, exists := class.Constants["BYPASS_CURRENT"]; exists {
			if bypassCurrent.Value.Data.(int64) != RECURSIVE_TREE_ITERATOR_BYPASS_CURRENT {
				t.Errorf("BYPASS_CURRENT should equal %d", RECURSIVE_TREE_ITERATOR_BYPASS_CURRENT)
			}
		}
	})

	t.Run("InheritedMethods", func(t *testing.T) {
		class := GetRecursiveTreeIteratorClass()

		// Test inherited methods exist from RecursiveIteratorIterator
		inheritedMethods := []string{"key", "next", "rewind", "valid", "getDepth"}
		for _, methodName := range inheritedMethods {
			if _, exists := class.Methods[methodName]; !exists {
				t.Errorf("Inherited method '%s' should exist", methodName)
			}
		}
	})
}

// Helper function to create RecursiveTreeIterator for testing
func createRecursiveTreeIterator(innerIter *values.Value, flags int64) *values.Value {
	recursiveTreeObj := &values.Object{
		ClassName:  "RecursiveTreeIterator",
		Properties: make(map[string]*values.Value),
	}
	recursiveTreeObj.Properties["__iterator"] = innerIter
	recursiveTreeObj.Properties["__flags"] = values.NewInt(flags)
	recursiveTreeObj.Properties["__prefixes"] = createDefaultPrefixes()
	recursiveTreeObj.Properties["__postfix"] = values.NewString("")
	recursiveTreeObj.Properties["__depth"] = values.NewInt(0)
	recursiveTreeObj.Properties["__max_depth"] = values.NewInt(-1)

	return &values.Value{
		Type: values.TypeObject,
		Data: recursiveTreeObj,
	}
}