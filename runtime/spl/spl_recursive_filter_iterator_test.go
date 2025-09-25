package spl

import (
	"os"
	"testing"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

func TestRecursiveFilterIterator(t *testing.T) {
	registry.Initialize()

	// Manually register the SPL classes for testing
	for _, class := range GetSplClasses() {
		err := registry.GlobalRegistry.RegisterClass(class)
		if err != nil {
			t.Fatalf("Failed to register class %s: %v", class.Name, err)
		}
	}

	// Manually register the SPL interfaces for testing
	for _, iface := range GetSplInterfaces() {
		err := registry.GlobalRegistry.RegisterInterface(iface)
		if err != nil {
			t.Fatalf("Failed to register interface %s: %v", iface.Name, err)
		}
	}

	ctx := &mockContext{registry: registry.GlobalRegistry}

	t.Run("RecursiveFilterIteratorAbstract", func(t *testing.T) {
		testRecursiveFilterIteratorAbstract(t, ctx)
	})

	t.Run("RecursiveFilterIteratorInheritance", func(t *testing.T) {
		testRecursiveFilterIteratorInheritance(t, ctx)
	})

	t.Run("RecursiveFilterIteratorConstructor", func(t *testing.T) {
		testRecursiveFilterIteratorConstructor(t, ctx)
	})

	t.Run("RecursiveFilterIteratorMethods", func(t *testing.T) {
		testRecursiveFilterIteratorMethods(t, ctx)
	})
}

func testRecursiveFilterIteratorAbstract(t *testing.T, ctx *mockContext) {
	// Test that RecursiveFilterIterator is abstract and cannot be instantiated directly
	class, err := ctx.registry.GetClass("RecursiveFilterIterator")
	if err != nil {
		t.Fatalf("RecursiveFilterIterator class not found: %v", err)
	}

	// Check that it's marked as abstract
	if !class.IsAbstract {
		t.Error("RecursiveFilterIterator should be marked as abstract")
	}

	// Test that accept method is abstract (calling it should fail)
	obj := &values.Object{
		ClassName:  "RecursiveFilterIterator",
		Properties: make(map[string]*values.Value),
	}
	thisObj := &values.Value{
		Type: values.TypeObject,
		Data: obj,
	}

	acceptMethod := class.Methods["accept"]
	acceptImpl := acceptMethod.Implementation.(*BuiltinMethodImpl)
	_, err = acceptImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
	if err == nil {
		t.Error("Expected error when calling abstract accept method, but got none")
	}

	if !contains(err.Error(), "abstract") || !contains(err.Error(), "must be implemented") {
		t.Errorf("Expected error about abstract method, got: %s", err.Error())
	}
}

func testRecursiveFilterIteratorInheritance(t *testing.T, ctx *mockContext) {
	class, err := ctx.registry.GetClass("RecursiveFilterIterator")
	if err != nil {
		t.Fatalf("RecursiveFilterIterator class not found: %v", err)
	}

	// Test parent class
	if class.Parent != "FilterIterator" {
		t.Errorf("RecursiveFilterIterator parent should be FilterIterator, got %s", class.Parent)
	}

	// Test interfaces
	expectedInterfaces := []string{"Iterator", "OuterIterator", "RecursiveIterator"}
	for _, interfaceName := range expectedInterfaces {
		found := false
		for _, iface := range class.Interfaces {
			if iface == interfaceName {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("RecursiveFilterIterator should implement interface %s", interfaceName)
		}
	}

	// Test that it has inherited methods from FilterIterator
	inheritedMethods := []string{"getInnerIterator", "valid", "key", "current", "next", "rewind"}
	for _, methodName := range inheritedMethods {
		if _, exists := class.Methods[methodName]; !exists {
			t.Errorf("RecursiveFilterIterator should inherit method %s from FilterIterator", methodName)
		}
	}

	// Test that it has RecursiveIterator methods
	recursiveMethods := []string{"hasChildren", "getChildren"}
	for _, methodName := range recursiveMethods {
		if _, exists := class.Methods[methodName]; !exists {
			t.Errorf("RecursiveFilterIterator should have RecursiveIterator method %s", methodName)
		}
	}
}

func testRecursiveFilterIteratorConstructor(t *testing.T, ctx *mockContext) {
	class, err := ctx.registry.GetClass("RecursiveFilterIterator")
	if err != nil {
		t.Fatalf("RecursiveFilterIterator class not found: %v", err)
	}

	// Test constructor with valid RecursiveIterator
	testDir := "/tmp/test_recursive_filter_constructor"
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	file, err := os.Create(testDir + "/test.txt")
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	file.Close()

	// Create RecursiveDirectoryIterator
	recursiveDirClass, err := ctx.registry.GetClass("RecursiveDirectoryIterator")
	if err != nil {
		t.Fatalf("RecursiveDirectoryIterator class not found: %v", err)
	}

	recursiveDirObj := &values.Object{
		ClassName:  "RecursiveDirectoryIterator",
		Properties: make(map[string]*values.Value),
	}
	recursiveDirThis := &values.Value{
		Type: values.TypeObject,
		Data: recursiveDirObj,
	}

	recursiveDirConstructor := recursiveDirClass.Methods["__construct"]
	recursiveDirConstructorImpl := recursiveDirConstructor.Implementation.(*BuiltinMethodImpl)
	_, err = recursiveDirConstructorImpl.GetFunction().Builtin(ctx, []*values.Value{recursiveDirThis, values.NewString(testDir)})
	if err != nil {
		t.Fatalf("Failed to create RecursiveDirectoryIterator: %v", err)
	}

	// Test RecursiveFilterIterator constructor with valid RecursiveIterator
	obj := &values.Object{
		ClassName:  "RecursiveFilterIterator",
		Properties: make(map[string]*values.Value),
	}
	thisObj := &values.Value{
		Type: values.TypeObject,
		Data: obj,
	}

	constructorMethod := class.Methods["__construct"]
	constructorImpl := constructorMethod.Implementation.(*BuiltinMethodImpl)
	_, err = constructorImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj, recursiveDirThis})
	if err != nil {
		t.Fatalf("Constructor failed with valid RecursiveIterator: %v", err)
	}

	// Verify inner iterator was stored
	if _, exists := obj.Properties["__iterator"]; !exists {
		t.Error("Constructor should store inner iterator in __iterator property")
	}

	// Test constructor with insufficient parameters
	obj2 := &values.Object{
		ClassName:  "RecursiveFilterIterator",
		Properties: make(map[string]*values.Value),
	}
	thisObj2 := &values.Value{
		Type: values.TypeObject,
		Data: obj2,
	}

	_, err = constructorImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj2})
	if err == nil {
		t.Error("Expected error for constructor with insufficient parameters, but got none")
	}

	// Test constructor with non-object parameter
	_, err = constructorImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj2, values.NewString("not an iterator")})
	if err == nil {
		t.Error("Expected error for constructor with non-object parameter, but got none")
	}
}

func testRecursiveFilterIteratorMethods(t *testing.T, ctx *mockContext) {
	// Create test directory structure
	testDir := "/tmp/test_recursive_filter_methods"
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	// Create subdirectory with files
	subDir := testDir + "/subdir"
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	// Create files
	file1, err := os.Create(testDir + "/file1.txt")
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	file1.Close()

	file2, err := os.Create(subDir + "/file2.txt")
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	file2.Close()

	// Create RecursiveDirectoryIterator
	recursiveDirClass, err := ctx.registry.GetClass("RecursiveDirectoryIterator")
	if err != nil {
		t.Fatalf("RecursiveDirectoryIterator class not found: %v", err)
	}

	recursiveDirObj := &values.Object{
		ClassName:  "RecursiveDirectoryIterator",
		Properties: make(map[string]*values.Value),
	}
	recursiveDirThis := &values.Value{
		Type: values.TypeObject,
		Data: recursiveDirObj,
	}

	recursiveDirConstructor := recursiveDirClass.Methods["__construct"]
	recursiveDirConstructorImpl := recursiveDirConstructor.Implementation.(*BuiltinMethodImpl)
	_, err = recursiveDirConstructorImpl.GetFunction().Builtin(ctx, []*values.Value{recursiveDirThis, values.NewString(testDir)})
	if err != nil {
		t.Fatalf("Failed to create RecursiveDirectoryIterator: %v", err)
	}

	// Create RecursiveFilterIterator
	class, err := ctx.registry.GetClass("RecursiveFilterIterator")
	if err != nil {
		t.Fatalf("RecursiveFilterIterator class not found: %v", err)
	}

	obj := &values.Object{
		ClassName:  "RecursiveFilterIterator",
		Properties: make(map[string]*values.Value),
	}
	thisObj := &values.Value{
		Type: values.TypeObject,
		Data: obj,
	}

	constructorMethod := class.Methods["__construct"]
	constructorImpl := constructorMethod.Implementation.(*BuiltinMethodImpl)
	_, err = constructorImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj, recursiveDirThis})
	if err != nil {
		t.Fatalf("Constructor failed: %v", err)
	}

	// Test hasChildren method
	hasChildrenMethod := class.Methods["hasChildren"]
	hasChildrenImpl := hasChildrenMethod.Implementation.(*BuiltinMethodImpl)

	// First rewind the inner iterator
	rewindMethod := class.Methods["rewind"]
	rewindImpl := rewindMethod.Implementation.(*BuiltinMethodImpl)
	_, err = rewindImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
	if err != nil {
		t.Fatalf("Rewind failed: %v", err)
	}

	// Test that hasChildren method works
	hasChildrenResult, err := hasChildrenImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
	if err != nil {
		t.Fatalf("hasChildren failed: %v", err)
	}

	if hasChildrenResult.Type != values.TypeBool {
		t.Errorf("hasChildren should return bool, got %v", hasChildrenResult.Type)
	}

	// Test getChildren method (this might fail because we need proper iteration state)
	getChildrenMethod := class.Methods["getChildren"]
	getChildrenImpl := getChildrenMethod.Implementation.(*BuiltinMethodImpl)

	// Move to a directory entry that has children
	validMethod := class.Methods["valid"]
	validImpl := validMethod.Implementation.(*BuiltinMethodImpl)
	nextMethod := class.Methods["next"]
	nextImpl := nextMethod.Implementation.(*BuiltinMethodImpl)

	// Find an entry with children
	maxAttempts := 10
	for i := 0; i < maxAttempts; i++ {
		validResult, err := validImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil || !validResult.ToBool() {
			break
		}

		hasChildrenResult, err := hasChildrenImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err == nil && hasChildrenResult.ToBool() {
			// Try getChildren
			childrenResult, err := getChildrenImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
			if err != nil {
				t.Logf("getChildren failed (expected with abstract class): %v", err)
			} else {
				if childrenResult.Type == values.TypeObject {
					childObj := childrenResult.Data.(*values.Object)
					if childObj.ClassName == "RecursiveFilterIterator" {
						t.Logf("getChildren returned RecursiveFilterIterator as expected")
					}
				}
			}
			break
		}

		_, err = nextImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			break
		}
	}

	// Test getInnerIterator method (inherited from FilterIterator)
	getInnerIteratorMethod := class.Methods["getInnerIterator"]
	if getInnerIteratorMethod == nil {
		t.Error("RecursiveFilterIterator should have getInnerIterator method")
	} else {
		getInnerIteratorImpl := getInnerIteratorMethod.Implementation.(*BuiltinMethodImpl)
		innerResult, err := getInnerIteratorImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("getInnerIterator failed: %v", err)
		}

		if innerResult.Type != values.TypeObject {
			t.Errorf("getInnerIterator should return object, got %v", innerResult.Type)
		}
	}
}