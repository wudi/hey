package spl

import (
	"os"
	"testing"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

func TestParentIterator(t *testing.T) {
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

	t.Run("ParentIteratorBasic", func(t *testing.T) {
		testParentIteratorBasic(t, ctx)
	})

	t.Run("ParentIteratorFiltering", func(t *testing.T) {
		testParentIteratorFiltering(t, ctx)
	})

	t.Run("ParentIteratorRecursive", func(t *testing.T) {
		testParentIteratorRecursive(t, ctx)
	})

	t.Run("ParentIteratorInheritance", func(t *testing.T) {
		testParentIteratorInheritance(t, ctx)
	})

	t.Run("ParentIteratorErrorCases", func(t *testing.T) {
		testParentIteratorErrorCases(t, ctx)
	})
}

func testParentIteratorBasic(t *testing.T, ctx *mockContext) {
	// Create test directory structure
	testDir := "/tmp/test_parent_iterator_basic"
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	// Create directories and files
	subDir := testDir + "/parent_dir"
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	file1, err := os.Create(testDir + "/file1.txt")
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	file1.Close()

	file2, err := os.Create(subDir + "/child_file.txt")
	if err != nil {
		t.Fatalf("Failed to create child file: %v", err)
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
	skipDotsFlag := int64(4096) // SKIP_DOTS
	_, err = recursiveDirConstructorImpl.GetFunction().Builtin(ctx, []*values.Value{recursiveDirThis, values.NewString(testDir), values.NewInt(skipDotsFlag)})
	if err != nil {
		t.Fatalf("Failed to create RecursiveDirectoryIterator: %v", err)
	}

	// Create ParentIterator
	parentClass, err := ctx.registry.GetClass("ParentIterator")
	if err != nil {
		t.Fatalf("ParentIterator class not found: %v", err)
	}

	parentObj := &values.Object{
		ClassName:  "ParentIterator",
		Properties: make(map[string]*values.Value),
	}
	parentThis := &values.Value{
		Type: values.TypeObject,
		Data: parentObj,
	}

	parentConstructor := parentClass.Methods["__construct"]
	parentConstructorImpl := parentConstructor.Implementation.(*BuiltinMethodImpl)
	_, err = parentConstructorImpl.GetFunction().Builtin(ctx, []*values.Value{parentThis, recursiveDirThis})
	if err != nil {
		t.Fatalf("ParentIterator constructor failed: %v", err)
	}

	// Test that ParentIterator filters correctly
	rewindMethod := parentClass.Methods["rewind"]
	validMethod := parentClass.Methods["valid"]
	currentMethod := parentClass.Methods["current"]
	nextMethod := parentClass.Methods["next"]

	rewindImpl := rewindMethod.Implementation.(*BuiltinMethodImpl)
	validImpl := validMethod.Implementation.(*BuiltinMethodImpl)
	currentImpl := currentMethod.Implementation.(*BuiltinMethodImpl)
	nextImpl := nextMethod.Implementation.(*BuiltinMethodImpl)

	_, err = rewindImpl.GetFunction().Builtin(ctx, []*values.Value{parentThis})
	if err != nil {
		t.Fatalf("Rewind failed: %v", err)
	}

	foundParentDir := false
	foundFile := false
	itemCount := 0

	for {
		validResult, err := validImpl.GetFunction().Builtin(ctx, []*values.Value{parentThis})
		if err != nil {
			t.Fatalf("Valid failed: %v", err)
		}

		if !validResult.ToBool() {
			break
		}

		currentResult, err := currentImpl.GetFunction().Builtin(ctx, []*values.Value{parentThis})
		if err != nil {
			t.Fatalf("Current failed: %v", err)
		}

		itemCount++

		// Analyze what we got
		if currentResult.Type == values.TypeObject {
			currentObj := currentResult.Data.(*values.Object)
			if filepath, ok := currentObj.Properties["__filepath"]; ok {
				path := filepath.ToString()
				if containsSubstring(path, "parent_dir") {
					foundParentDir = true
				}
				if containsSubstring(path, "file1.txt") {
					foundFile = true
				}
			}
		}

		_, err = nextImpl.GetFunction().Builtin(ctx, []*values.Value{parentThis})
		if err != nil {
			t.Fatalf("Next failed: %v", err)
		}
	}

	// ParentIterator should show directories that have children, not files
	if foundFile {
		t.Error("ParentIterator should not show files")
	}

	if !foundParentDir {
		t.Error("ParentIterator should show directories with children")
	}

	if itemCount == 0 {
		t.Error("ParentIterator should find at least some parent directories")
	}

	t.Logf("ParentIterator found %d items", itemCount)
}

func testParentIteratorFiltering(t *testing.T, ctx *mockContext) {
	// Create test structure with specific pattern
	testDir := "/tmp/test_parent_filtering"
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	// Create empty directory (should appear in ParentIterator in PHP behavior)
	emptyDir := testDir + "/empty_dir"
	if err := os.MkdirAll(emptyDir, 0755); err != nil {
		t.Fatalf("Failed to create empty directory: %v", err)
	}

	// Create directory with children
	parentDir := testDir + "/parent_with_children"
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		t.Fatalf("Failed to create parent directory: %v", err)
	}

	childFile, err := os.Create(parentDir + "/child.txt")
	if err != nil {
		t.Fatalf("Failed to create child file: %v", err)
	}
	childFile.Close()

	// Create regular file
	regularFile, err := os.Create(testDir + "/regular.txt")
	if err != nil {
		t.Fatalf("Failed to create regular file: %v", err)
	}
	regularFile.Close()

	// Test filtering behavior
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
	skipDotsFlag := int64(4096) // SKIP_DOTS
	_, err = recursiveDirConstructorImpl.GetFunction().Builtin(ctx, []*values.Value{recursiveDirThis, values.NewString(testDir), values.NewInt(skipDotsFlag)})
	if err != nil {
		t.Fatalf("Failed to create RecursiveDirectoryIterator: %v", err)
	}

	parentClass, err := ctx.registry.GetClass("ParentIterator")
	if err != nil {
		t.Fatalf("ParentIterator class not found: %v", err)
	}

	parentObj := &values.Object{
		ClassName:  "ParentIterator",
		Properties: make(map[string]*values.Value),
	}
	parentThis := &values.Value{
		Type: values.TypeObject,
		Data: parentObj,
	}

	parentConstructor := parentClass.Methods["__construct"]
	parentConstructorImpl := parentConstructor.Implementation.(*BuiltinMethodImpl)
	_, err = parentConstructorImpl.GetFunction().Builtin(ctx, []*values.Value{parentThis, recursiveDirThis})
	if err != nil {
		t.Fatalf("ParentIterator constructor failed: %v", err)
	}

	// Count results and check types
	rewindMethod := parentClass.Methods["rewind"]
	validMethod := parentClass.Methods["valid"]
	currentMethod := parentClass.Methods["current"]
	nextMethod := parentClass.Methods["next"]

	rewindImpl := rewindMethod.Implementation.(*BuiltinMethodImpl)
	validImpl := validMethod.Implementation.(*BuiltinMethodImpl)
	currentImpl := currentMethod.Implementation.(*BuiltinMethodImpl)
	nextImpl := nextMethod.Implementation.(*BuiltinMethodImpl)

	_, err = rewindImpl.GetFunction().Builtin(ctx, []*values.Value{parentThis})
	if err != nil {
		t.Fatalf("Rewind failed: %v", err)
	}

	directoryCount := 0
	fileCount := 0

	for {
		validResult, err := validImpl.GetFunction().Builtin(ctx, []*values.Value{parentThis})
		if err != nil {
			t.Fatalf("Valid failed: %v", err)
		}

		if !validResult.ToBool() {
			break
		}

		currentResult, err := currentImpl.GetFunction().Builtin(ctx, []*values.Value{parentThis})
		if err != nil {
			t.Fatalf("Current failed: %v", err)
		}

		// Check what type of item we got
		if currentResult.Type == values.TypeObject {
			currentObj := currentResult.Data.(*values.Object)
			if filepath, ok := currentObj.Properties["__filepath"]; ok {
				path := filepath.ToString()

				// Check if it's a directory or file by looking at the actual filesystem
				if info, err := os.Stat(path); err == nil {
					if info.IsDir() {
						directoryCount++
					} else {
						fileCount++
					}
				}
			}
		}

		_, err = nextImpl.GetFunction().Builtin(ctx, []*values.Value{parentThis})
		if err != nil {
			t.Fatalf("Next failed: %v", err)
		}
	}

	// ParentIterator should only show directories
	if fileCount > 0 {
		t.Errorf("ParentIterator should not show files, but found %d files", fileCount)
	}

	if directoryCount == 0 {
		t.Error("ParentIterator should find at least some directories")
	}

	t.Logf("ParentIterator found %d directories, %d files", directoryCount, fileCount)
}

func testParentIteratorRecursive(t *testing.T, ctx *mockContext) {
	// Test hasChildren and getChildren methods
	testDir := "/tmp/test_parent_recursive"
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	// Create nested structure
	level1 := testDir + "/level1"
	level2 := level1 + "/level2"
	if err := os.MkdirAll(level2, 0755); err != nil {
		t.Fatalf("Failed to create nested directories: %v", err)
	}

	file, err := os.Create(level2 + "/deep.txt")
	if err != nil {
		t.Fatalf("Failed to create deep file: %v", err)
	}
	file.Close()

	// Create ParentIterator
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
	skipDotsFlag := int64(4096) // SKIP_DOTS
	_, err = recursiveDirConstructorImpl.GetFunction().Builtin(ctx, []*values.Value{recursiveDirThis, values.NewString(testDir), values.NewInt(skipDotsFlag)})
	if err != nil {
		t.Fatalf("Failed to create RecursiveDirectoryIterator: %v", err)
	}

	parentClass, err := ctx.registry.GetClass("ParentIterator")
	if err != nil {
		t.Fatalf("ParentIterator class not found: %v", err)
	}

	parentObj := &values.Object{
		ClassName:  "ParentIterator",
		Properties: make(map[string]*values.Value),
	}
	parentThis := &values.Value{
		Type: values.TypeObject,
		Data: parentObj,
	}

	parentConstructor := parentClass.Methods["__construct"]
	parentConstructorImpl := parentConstructor.Implementation.(*BuiltinMethodImpl)
	_, err = parentConstructorImpl.GetFunction().Builtin(ctx, []*values.Value{parentThis, recursiveDirThis})
	if err != nil {
		t.Fatalf("ParentIterator constructor failed: %v", err)
	}

	// Test hasChildren and getChildren
	hasChildrenMethod := parentClass.Methods["hasChildren"]
	getChildrenMethod := parentClass.Methods["getChildren"]
	rewindMethod := parentClass.Methods["rewind"]

	hasChildrenImpl := hasChildrenMethod.Implementation.(*BuiltinMethodImpl)
	getChildrenImpl := getChildrenMethod.Implementation.(*BuiltinMethodImpl)
	rewindImpl := rewindMethod.Implementation.(*BuiltinMethodImpl)

	_, err = rewindImpl.GetFunction().Builtin(ctx, []*values.Value{parentThis})
	if err != nil {
		t.Fatalf("Rewind failed: %v", err)
	}

	// Test hasChildren
	hasChildrenResult, err := hasChildrenImpl.GetFunction().Builtin(ctx, []*values.Value{parentThis})
	if err != nil {
		t.Fatalf("hasChildren failed: %v", err)
	}

	if hasChildrenResult.Type != values.TypeBool {
		t.Errorf("hasChildren should return bool, got %v", hasChildrenResult.Type)
	}

	// Test getChildren
	if hasChildrenResult.ToBool() {
		childrenResult, err := getChildrenImpl.GetFunction().Builtin(ctx, []*values.Value{parentThis})
		if err != nil {
			t.Fatalf("getChildren failed: %v", err)
		}

		if childrenResult.Type != values.TypeObject {
			t.Errorf("getChildren should return object, got %v", childrenResult.Type)
		} else {
			childObj := childrenResult.Data.(*values.Object)
			if childObj.ClassName != "ParentIterator" {
				t.Errorf("getChildren should return ParentIterator, got %s", childObj.ClassName)
			}
		}
	}
}

func testParentIteratorInheritance(t *testing.T, ctx *mockContext) {
	class, err := ctx.registry.GetClass("ParentIterator")
	if err != nil {
		t.Fatalf("ParentIterator class not found: %v", err)
	}

	// Test parent class
	if class.Parent != "RecursiveFilterIterator" {
		t.Errorf("ParentIterator parent should be RecursiveFilterIterator, got %s", class.Parent)
	}

	// Test that it's not abstract
	if class.IsAbstract {
		t.Error("ParentIterator should not be abstract")
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
			t.Errorf("ParentIterator should implement interface %s", interfaceName)
		}
	}

	// Test that it has all required methods
	requiredMethods := []string{"__construct", "accept", "hasChildren", "getChildren", "getInnerIterator", "valid", "key", "current", "next", "rewind"}
	for _, methodName := range requiredMethods {
		if _, exists := class.Methods[methodName]; !exists {
			t.Errorf("ParentIterator should have method %s", methodName)
		}
	}
}

func testParentIteratorErrorCases(t *testing.T, ctx *mockContext) {
	parentClass, err := ctx.registry.GetClass("ParentIterator")
	if err != nil {
		t.Fatalf("ParentIterator class not found: %v", err)
	}

	// Test constructor with insufficient parameters
	obj := &values.Object{
		ClassName:  "ParentIterator",
		Properties: make(map[string]*values.Value),
	}
	thisObj := &values.Value{
		Type: values.TypeObject,
		Data: obj,
	}

	constructorMethod := parentClass.Methods["__construct"]
	constructorImpl := constructorMethod.Implementation.(*BuiltinMethodImpl)

	_, err = constructorImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
	if err == nil {
		t.Error("Expected error for constructor with insufficient parameters, but got none")
	}

	// Test constructor with non-object parameter
	_, err = constructorImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj, values.NewString("not an iterator")})
	if err == nil {
		t.Error("Expected error for constructor with non-object parameter, but got none")
	}

	if !contains(err.Error(), "must be of type RecursiveIterator") {
		t.Errorf("Expected error about RecursiveIterator type, got: %s", err.Error())
	}
}