package spl

import (
	"os"
	"testing"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

func TestRecursiveDirectoryIterator(t *testing.T) {
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

	t.Run("BasicRecursiveDirectoryIterator", func(t *testing.T) {
		testBasicRecursiveDirectoryIterator(t, ctx)
	})

	t.Run("RecursiveDirectoryIteratorHasChildren", func(t *testing.T) {
		testRecursiveDirectoryIteratorHasChildren(t, ctx)
	})

	t.Run("RecursiveDirectoryIteratorGetChildren", func(t *testing.T) {
		testRecursiveDirectoryIteratorGetChildren(t, ctx)
	})

	t.Run("RecursiveDirectoryIteratorFlags", func(t *testing.T) {
		testRecursiveDirectoryIteratorFlags(t, ctx)
	})

	t.Run("RecursiveDirectoryIteratorIteration", func(t *testing.T) {
		testRecursiveDirectoryIteratorIteration(t, ctx)
	})

	t.Run("RecursiveDirectoryIteratorErrorCases", func(t *testing.T) {
		testRecursiveDirectoryIteratorErrorCases(t, ctx)
	})

	t.Run("RecursiveDirectoryIteratorInheritance", func(t *testing.T) {
		testRecursiveDirectoryIteratorInheritance(t, ctx)
	})

	t.Run("RecursiveDirectoryIteratorWithRecursiveIteratorIterator", func(t *testing.T) {
		testRecursiveDirectoryIteratorWithRecursiveIteratorIterator(t, ctx)
	})
}

func testBasicRecursiveDirectoryIterator(t *testing.T, ctx *mockContext) {
	// Create test directory structure
	testDir := "/tmp/test_recursive_basic"
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	// Create subdirectory and files
	subDir := testDir + "/subdir"
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	// Create test files
	testFiles := []string{
		testDir + "/file1.txt",
		subDir + "/file2.txt",
	}
	for _, file := range testFiles {
		f, err := os.Create(file)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
		f.Close()
	}

	// Get the RecursiveDirectoryIterator class
	class, err := ctx.registry.GetClass("RecursiveDirectoryIterator")
	if err != nil {
		t.Fatalf("RecursiveDirectoryIterator class not found: %v", err)
	}

	// Create RecursiveDirectoryIterator instance
	obj := &values.Object{
		ClassName:  "RecursiveDirectoryIterator",
		Properties: make(map[string]*values.Value),
	}
	thisObj := &values.Value{
		Type: values.TypeObject,
		Data: obj,
	}

	// Test constructor
	constructorMethod := class.Methods["__construct"]
	constructorImpl := constructorMethod.Implementation.(*BuiltinMethodImpl)
	_, err = constructorImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj, values.NewString(testDir)})
	if err != nil {
		t.Fatalf("Constructor failed: %v", err)
	}

	// Test that it implements RecursiveIterator
	// This is tested by checking the interface in the class descriptor
	expectedInterfaces := []string{"Iterator", "RecursiveIterator", "SeekableIterator"}
	for _, interfaceName := range expectedInterfaces {
		found := false
		for _, iface := range class.Interfaces {
			if iface == interfaceName {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("RecursiveDirectoryIterator should implement interface %s", interfaceName)
		}
	}
}

func testRecursiveDirectoryIteratorHasChildren(t *testing.T, ctx *mockContext) {
	// Create test directory structure
	testDir := "/tmp/test_recursive_has_children"
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	// Create subdirectory and files
	subDir := testDir + "/subdir"
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	// Create a file
	file, err := os.Create(testDir + "/file.txt")
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	file.Close()

	class, err := ctx.registry.GetClass("RecursiveDirectoryIterator")
	if err != nil {
		t.Fatalf("RecursiveDirectoryIterator class not found: %v", err)
	}

	obj := &values.Object{
		ClassName:  "RecursiveDirectoryIterator",
		Properties: make(map[string]*values.Value),
	}
	thisObj := &values.Value{
		Type: values.TypeObject,
		Data: obj,
	}

	constructorMethod := class.Methods["__construct"]
	constructorImpl := constructorMethod.Implementation.(*BuiltinMethodImpl)
	_, err = constructorImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj, values.NewString(testDir)})
	if err != nil {
		t.Fatalf("Constructor failed: %v", err)
	}

	// Get methods
	rewindMethod := class.Methods["rewind"]
	validMethod := class.Methods["valid"]
	currentMethod := class.Methods["current"]
	hasChildrenMethod := class.Methods["hasChildren"]
	nextMethod := class.Methods["next"]

	rewindImpl := rewindMethod.Implementation.(*BuiltinMethodImpl)
	validImpl := validMethod.Implementation.(*BuiltinMethodImpl)
	currentImpl := currentMethod.Implementation.(*BuiltinMethodImpl)
	hasChildrenImpl := hasChildrenMethod.Implementation.(*BuiltinMethodImpl)
	nextImpl := nextMethod.Implementation.(*BuiltinMethodImpl)

	// Start iteration
	_, err = rewindImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
	if err != nil {
		t.Fatalf("Rewind failed: %v", err)
	}

	foundDir := false
	foundFile := false

	for {
		validResult, err := validImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("Valid failed: %v", err)
		}

		if !validResult.ToBool() {
			break
		}

		currentResult, err := currentImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("Current failed: %v", err)
		}

		hasChildrenResult, err := hasChildrenImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("HasChildren failed: %v", err)
		}

		// Check if we found a directory with children or a file without children
		if currentResult.Type == values.TypeObject {
			currentObj := currentResult.Data.(*values.Object)
			if filepath, ok := currentObj.Properties["__filepath"]; ok {
				path := filepath.ToString()
				if containsSubstring(path, "subdir") {
					foundDir = true
					if !hasChildrenResult.ToBool() {
						t.Error("Directory should have children")
					}
				} else if containsSubstring(path, "file.txt") {
					foundFile = true
					if hasChildrenResult.ToBool() {
						t.Error("File should not have children")
					}
				}
			}
		}

		_, err = nextImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("Next failed: %v", err)
		}
	}

	if !foundDir {
		t.Error("Should have found a directory")
	}
	if !foundFile {
		t.Error("Should have found a file")
	}
}

func testRecursiveDirectoryIteratorGetChildren(t *testing.T, ctx *mockContext) {
	// Create test directory structure
	testDir := "/tmp/test_recursive_get_children"
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	// Create subdirectory with file
	subDir := testDir + "/subdir"
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	file, err := os.Create(subDir + "/child.txt")
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	file.Close()

	class, err := ctx.registry.GetClass("RecursiveDirectoryIterator")
	if err != nil {
		t.Fatalf("RecursiveDirectoryIterator class not found: %v", err)
	}

	obj := &values.Object{
		ClassName:  "RecursiveDirectoryIterator",
		Properties: make(map[string]*values.Value),
	}
	thisObj := &values.Value{
		Type: values.TypeObject,
		Data: obj,
	}

	constructorMethod := class.Methods["__construct"]
	constructorImpl := constructorMethod.Implementation.(*BuiltinMethodImpl)
	_, err = constructorImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj, values.NewString(testDir)})
	if err != nil {
		t.Fatalf("Constructor failed: %v", err)
	}

	// Find the subdirectory and get its children
	methods := map[string]*BuiltinMethodImpl{
		"rewind":      class.Methods["rewind"].Implementation.(*BuiltinMethodImpl),
		"valid":       class.Methods["valid"].Implementation.(*BuiltinMethodImpl),
		"current":     class.Methods["current"].Implementation.(*BuiltinMethodImpl),
		"hasChildren": class.Methods["hasChildren"].Implementation.(*BuiltinMethodImpl),
		"getChildren": class.Methods["getChildren"].Implementation.(*BuiltinMethodImpl),
		"next":        class.Methods["next"].Implementation.(*BuiltinMethodImpl),
	}

	_, err = methods["rewind"].GetFunction().Builtin(ctx, []*values.Value{thisObj})
	if err != nil {
		t.Fatalf("Rewind failed: %v", err)
	}

	for {
		validResult, err := methods["valid"].GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("Valid failed: %v", err)
		}

		if !validResult.ToBool() {
			break
		}

		hasChildrenResult, err := methods["hasChildren"].GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("HasChildren failed: %v", err)
		}

		if hasChildrenResult.ToBool() {
			// Found a directory with children - get the children iterator
			childrenResult, err := methods["getChildren"].GetFunction().Builtin(ctx, []*values.Value{thisObj})
			if err != nil {
				t.Fatalf("GetChildren failed: %v", err)
			}

			if childrenResult.Type != values.TypeObject {
				t.Errorf("getChildren should return an object, got %v", childrenResult.Type)
			} else {
				childObj := childrenResult.Data.(*values.Object)
				if childObj.ClassName != "RecursiveDirectoryIterator" {
					t.Errorf("getChildren should return RecursiveDirectoryIterator, got %s", childObj.ClassName)
				}
			}

			// Test that we can iterate over the children
			childCurrentMethod := class.Methods["current"]
			childCurrentImpl := childCurrentMethod.Implementation.(*BuiltinMethodImpl)
			childCurrentResult, err := childCurrentImpl.GetFunction().Builtin(ctx, []*values.Value{childrenResult})
			if err != nil {
				t.Fatalf("Child current failed: %v", err)
			}

			if childCurrentResult.Type == values.TypeObject {
				// Successfully got a child - that's what we wanted to test
				t.Logf("Successfully got children iterator and child object")
			}

			break // Found what we were testing for
		}

		_, err = methods["next"].GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("Next failed: %v", err)
		}
	}
}

func testRecursiveDirectoryIteratorFlags(t *testing.T, ctx *mockContext) {
	// Create test directory
	testDir := "/tmp/test_recursive_flags"
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	file, err := os.Create(testDir + "/test.txt")
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	file.Close()

	class, err := ctx.registry.GetClass("RecursiveDirectoryIterator")
	if err != nil {
		t.Fatalf("RecursiveDirectoryIterator class not found: %v", err)
	}

	// Test SKIP_DOTS flag
	t.Run("SKIP_DOTS_Flag", func(t *testing.T) {
		obj := &values.Object{
			ClassName:  "RecursiveDirectoryIterator",
			Properties: make(map[string]*values.Value),
		}
		thisObj := &values.Value{
			Type: values.TypeObject,
			Data: obj,
		}

		constructorMethod := class.Methods["__construct"]
		constructorImpl := constructorMethod.Implementation.(*BuiltinMethodImpl)
		skipDotsFlag := int64(4096) // SKIP_DOTS
		_, err = constructorImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj, values.NewString(testDir), values.NewInt(skipDotsFlag)})
		if err != nil {
			t.Fatalf("Constructor with SKIP_DOTS failed: %v", err)
		}

		// Test that iteration skips '.' and '..' entries
		rewindMethod := class.Methods["rewind"]
		validMethod := class.Methods["valid"]
		currentMethod := class.Methods["current"]
		nextMethod := class.Methods["next"]

		rewindImpl := rewindMethod.Implementation.(*BuiltinMethodImpl)
		validImpl := validMethod.Implementation.(*BuiltinMethodImpl)
		currentImpl := currentMethod.Implementation.(*BuiltinMethodImpl)
		nextImpl := nextMethod.Implementation.(*BuiltinMethodImpl)

		_, err = rewindImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("Rewind failed: %v", err)
		}

		foundDotEntry := false

		for {
			validResult, err := validImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
			if err != nil {
				t.Fatalf("Valid failed: %v", err)
			}

			if !validResult.ToBool() {
				break
			}

			currentResult, err := currentImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
			if err != nil {
				t.Fatalf("Current failed: %v", err)
			}

			// Check if we got a dot entry (shouldn't happen with SKIP_DOTS)
			if currentResult.Type == values.TypeObject {
				currentObj := currentResult.Data.(*values.Object)
				if filepath, ok := currentObj.Properties["__filepath"]; ok {
					path := filepath.ToString()
					if containsSubstring(path, "/.") && (containsSubstring(path, "/.") || containsSubstring(path, "/..")) {
						foundDotEntry = true
					}
				}
			}

			_, err = nextImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
			if err != nil {
				t.Fatalf("Next failed: %v", err)
			}
		}

		if foundDotEntry {
			t.Error("SKIP_DOTS flag should prevent dot entries from appearing")
		}
	})

	// Test flags inheritance from FilesystemIterator
	t.Run("InheritedFlags", func(t *testing.T) {
		obj := &values.Object{
			ClassName:  "RecursiveDirectoryIterator",
			Properties: make(map[string]*values.Value),
		}
		thisObj := &values.Value{
			Type: values.TypeObject,
			Data: obj,
		}

		constructorMethod := class.Methods["__construct"]
		constructorImpl := constructorMethod.Implementation.(*BuiltinMethodImpl)
		flags := int64(32) // CURRENT_AS_PATHNAME
		_, err = constructorImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj, values.NewString(testDir), values.NewInt(flags)})
		if err != nil {
			t.Fatalf("Constructor with flags failed: %v", err)
		}

		// Test getFlags method
		getFlagsMethod := class.Methods["getFlags"]
		getFlagsImpl := getFlagsMethod.Implementation.(*BuiltinMethodImpl)
		flagsResult, err := getFlagsImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("getFlags failed: %v", err)
		}

		if flagsResult.ToInt() != flags {
			t.Errorf("Expected flags %d, got %d", flags, flagsResult.ToInt())
		}
	})
}

func testRecursiveDirectoryIteratorIteration(t *testing.T, ctx *mockContext) {
	// Create test directory structure
	testDir := "/tmp/test_recursive_iteration"
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	// Create files
	files := []string{"file1.txt", "file2.txt"}
	for _, filename := range files {
		file, err := os.Create(testDir + "/" + filename)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
		file.Close()
	}

	class, err := ctx.registry.GetClass("RecursiveDirectoryIterator")
	if err != nil {
		t.Fatalf("RecursiveDirectoryIterator class not found: %v", err)
	}

	obj := &values.Object{
		ClassName:  "RecursiveDirectoryIterator",
		Properties: make(map[string]*values.Value),
	}
	thisObj := &values.Value{
		Type: values.TypeObject,
		Data: obj,
	}

	constructorMethod := class.Methods["__construct"]
	constructorImpl := constructorMethod.Implementation.(*BuiltinMethodImpl)
	_, err = constructorImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj, values.NewString(testDir)})
	if err != nil {
		t.Fatalf("Constructor failed: %v", err)
	}

	// Test basic iterator methods
	methods := map[string]*BuiltinMethodImpl{
		"rewind": class.Methods["rewind"].Implementation.(*BuiltinMethodImpl),
		"valid":  class.Methods["valid"].Implementation.(*BuiltinMethodImpl),
		"key":    class.Methods["key"].Implementation.(*BuiltinMethodImpl),
		"next":   class.Methods["next"].Implementation.(*BuiltinMethodImpl),
	}

	_, err = methods["rewind"].GetFunction().Builtin(ctx, []*values.Value{thisObj})
	if err != nil {
		t.Fatalf("Rewind failed: %v", err)
	}

	iterationCount := 0
	for {
		validResult, err := methods["valid"].GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("Valid failed: %v", err)
		}

		if !validResult.ToBool() {
			break
		}

		keyResult, err := methods["key"].GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("Key failed: %v", err)
		}

		// Key should be a string
		if keyResult.Type != values.TypeString {
			t.Errorf("Key should be string, got %v", keyResult.Type)
		}

		iterationCount++

		_, err = methods["next"].GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("Next failed: %v", err)
		}
	}

	// Should have at least the files we created plus potentially '.' and '..'
	if iterationCount < len(files) {
		t.Errorf("Expected at least %d iterations, got %d", len(files), iterationCount)
	}
}

func testRecursiveDirectoryIteratorErrorCases(t *testing.T, ctx *mockContext) {
	class, err := ctx.registry.GetClass("RecursiveDirectoryIterator")
	if err != nil {
		t.Fatalf("RecursiveDirectoryIterator class not found: %v", err)
	}

	constructorMethod := class.Methods["__construct"]
	constructorImpl := constructorMethod.Implementation.(*BuiltinMethodImpl)

	// Test non-existent directory
	t.Run("NonExistentDirectory", func(t *testing.T) {
		obj := &values.Object{
			ClassName:  "RecursiveDirectoryIterator",
			Properties: make(map[string]*values.Value),
		}
		thisObj := &values.Value{
			Type: values.TypeObject,
			Data: obj,
		}

		_, err = constructorImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj, values.NewString("/nonexistent/path")})
		if err == nil {
			t.Error("Expected error for non-existent directory, but got none")
		}

		if !contains(err.Error(), "No such file or directory") {
			t.Errorf("Expected 'No such file or directory' error, got: %s", err.Error())
		}
	})

	// Test file instead of directory
	t.Run("FileInsteadOfDirectory", func(t *testing.T) {
		// Create a test file
		testFile := "/tmp/test_file_not_dir.txt"
		file, err := os.Create(testFile)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		file.Close()
		defer os.Remove(testFile)

		obj := &values.Object{
			ClassName:  "RecursiveDirectoryIterator",
			Properties: make(map[string]*values.Value),
		}
		thisObj := &values.Value{
			Type: values.TypeObject,
			Data: obj,
		}

		_, err = constructorImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj, values.NewString(testFile)})
		if err == nil {
			t.Error("Expected error for file instead of directory, but got none")
		}

		if !contains(err.Error(), "Not a directory") {
			t.Errorf("Expected 'Not a directory' error, got: %s", err.Error())
		}
	})

	// Test empty path
	t.Run("EmptyPath", func(t *testing.T) {
		obj := &values.Object{
			ClassName:  "RecursiveDirectoryIterator",
			Properties: make(map[string]*values.Value),
		}
		thisObj := &values.Value{
			Type: values.TypeObject,
			Data: obj,
		}

		_, err = constructorImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj, values.NewString("")})
		if err == nil {
			t.Error("Expected error for empty path, but got none")
		}

		if !contains(err.Error(), "Path cannot be empty") {
			t.Errorf("Expected 'Path cannot be empty' error, got: %s", err.Error())
		}
	})
}

func testRecursiveDirectoryIteratorInheritance(t *testing.T, ctx *mockContext) {
	class, err := ctx.registry.GetClass("RecursiveDirectoryIterator")
	if err != nil {
		t.Fatalf("RecursiveDirectoryIterator class not found: %v", err)
	}

	// Test parent class
	if class.Parent != "FilesystemIterator" {
		t.Errorf("RecursiveDirectoryIterator parent should be FilesystemIterator, got %s", class.Parent)
	}

	// Test interfaces
	expectedInterfaces := []string{"Iterator", "RecursiveIterator", "SeekableIterator"}
	for _, interfaceName := range expectedInterfaces {
		found := false
		for _, iface := range class.Interfaces {
			if iface == interfaceName {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("RecursiveDirectoryIterator should implement interface %s", interfaceName)
		}
	}

	// Test that it has inherited methods
	inheritedMethods := []string{"getFlags", "setFlags", "current", "key", "next", "valid", "rewind"}
	for _, methodName := range inheritedMethods {
		if _, exists := class.Methods[methodName]; !exists {
			t.Errorf("RecursiveDirectoryIterator should have inherited method %s", methodName)
		}
	}

	// Test that it has RecursiveIterator methods
	recursiveMethods := []string{"hasChildren", "getChildren"}
	for _, methodName := range recursiveMethods {
		if _, exists := class.Methods[methodName]; !exists {
			t.Errorf("RecursiveDirectoryIterator should have RecursiveIterator method %s", methodName)
		}
	}

	// Test that it has constants from FilesystemIterator
	if len(class.Constants) == 0 {
		t.Error("RecursiveDirectoryIterator should inherit constants from FilesystemIterator")
	}
}

func testRecursiveDirectoryIteratorWithRecursiveIteratorIterator(t *testing.T, ctx *mockContext) {
	// Create test directory structure for recursive traversal
	testDir := "/tmp/test_recursive_full"
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	// Create nested structure
	subDir := testDir + "/level1"
	subSubDir := subDir + "/level2"
	if err := os.MkdirAll(subSubDir, 0755); err != nil {
		t.Fatalf("Failed to create nested directories: %v", err)
	}

	// Create files at each level
	files := map[string]string{
		testDir + "/root.txt":         "root",
		subDir + "/level1.txt":        "level1",
		subSubDir + "/level2.txt":     "level2",
	}

	for file, content := range files {
		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}

	// Test that RecursiveDirectoryIterator can be used with RecursiveIteratorIterator
	// This is more of an integration test to verify the recursive behavior works

	recursiveDirClass, err := ctx.registry.GetClass("RecursiveDirectoryIterator")
	if err != nil {
		t.Fatalf("RecursiveDirectoryIterator class not found: %v", err)
	}

	recursiveIterClass, err := ctx.registry.GetClass("RecursiveIteratorIterator")
	if err != nil {
		t.Fatalf("RecursiveIteratorIterator class not found: %v", err)
	}

	// Create RecursiveDirectoryIterator
	dirIterObj := &values.Object{
		ClassName:  "RecursiveDirectoryIterator",
		Properties: make(map[string]*values.Value),
	}
	dirIterThis := &values.Value{
		Type: values.TypeObject,
		Data: dirIterObj,
	}

	dirConstructorMethod := recursiveDirClass.Methods["__construct"]
	dirConstructorImpl := dirConstructorMethod.Implementation.(*BuiltinMethodImpl)
	skipDotsFlag := int64(4096) // SKIP_DOTS to make output cleaner
	_, err = dirConstructorImpl.GetFunction().Builtin(ctx, []*values.Value{dirIterThis, values.NewString(testDir), values.NewInt(skipDotsFlag)})
	if err != nil {
		t.Fatalf("RecursiveDirectoryIterator constructor failed: %v", err)
	}

	// Create RecursiveIteratorIterator
	recursiveIterObj := &values.Object{
		ClassName:  "RecursiveIteratorIterator",
		Properties: make(map[string]*values.Value),
	}
	recursiveIterThis := &values.Value{
		Type: values.TypeObject,
		Data: recursiveIterObj,
	}

	recursiveConstructorMethod := recursiveIterClass.Methods["__construct"]
	recursiveConstructorImpl := recursiveConstructorMethod.Implementation.(*BuiltinMethodImpl)
	_, err = recursiveConstructorImpl.GetFunction().Builtin(ctx, []*values.Value{recursiveIterThis, dirIterThis})
	if err != nil {
		t.Fatalf("RecursiveIteratorIterator constructor failed: %v", err)
	}

	// Test that we can iterate and find files at multiple levels
	rewindMethod := recursiveIterClass.Methods["rewind"]
	validMethod := recursiveIterClass.Methods["valid"]
	currentMethod := recursiveIterClass.Methods["current"]
	nextMethod := recursiveIterClass.Methods["next"]

	rewindImpl := rewindMethod.Implementation.(*BuiltinMethodImpl)
	validImpl := validMethod.Implementation.(*BuiltinMethodImpl)
	currentImpl := currentMethod.Implementation.(*BuiltinMethodImpl)
	nextImpl := nextMethod.Implementation.(*BuiltinMethodImpl)

	_, err = rewindImpl.GetFunction().Builtin(ctx, []*values.Value{recursiveIterThis})
	if err != nil {
		t.Fatalf("Rewind failed: %v", err)
	}

	foundFiles := make(map[string]bool)
	iterationCount := 0

	for {
		validResult, err := validImpl.GetFunction().Builtin(ctx, []*values.Value{recursiveIterThis})
		if err != nil {
			t.Fatalf("Valid failed: %v", err)
		}

		if !validResult.ToBool() {
			break
		}

		currentResult, err := currentImpl.GetFunction().Builtin(ctx, []*values.Value{recursiveIterThis})
		if err != nil {
			t.Fatalf("Current failed: %v", err)
		}

		if currentResult.Type == values.TypeObject {
			currentObj := currentResult.Data.(*values.Object)
			if filepath, ok := currentObj.Properties["__filepath"]; ok {
				path := filepath.ToString()
				for expectedFile := range files {
					if containsSubstring(path, expectedFile[len(expectedFile)-10:]) { // Check last part of filename
						foundFiles[expectedFile] = true
					}
				}
			}
		}

		iterationCount++

		_, err = nextImpl.GetFunction().Builtin(ctx, []*values.Value{recursiveIterThis})
		if err != nil {
			t.Fatalf("Next failed: %v", err)
		}
	}

	// Verify we found files at multiple levels
	if len(foundFiles) < 2 {
		t.Errorf("Expected to find files at multiple levels, found %d files", len(foundFiles))
	}

	if iterationCount < 3 { // At least the directories and files we created
		t.Errorf("Expected more iterations for recursive traversal, got %d", iterationCount)
	}
}