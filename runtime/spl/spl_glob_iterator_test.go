package spl

import (
	"fmt"
	"os"
	"testing"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

func TestGlobIterator(t *testing.T) {
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

	t.Run("BasicGlobIterator", func(t *testing.T) {
		testBasicGlobIterator(t, ctx)
	})

	t.Run("GlobPatternMatching", func(t *testing.T) {
		testGlobPatternMatching(t, ctx)
	})

	t.Run("GlobIteratorCount", func(t *testing.T) {
		testGlobIteratorCount(t, ctx)
	})

	t.Run("GlobIteratorIteration", func(t *testing.T) {
		testGlobIteratorIteration(t, ctx)
	})

	t.Run("GlobIteratorFlags", func(t *testing.T) {
		testGlobIteratorFlags(t, ctx)
	})

	t.Run("GlobIteratorEmptyPattern", func(t *testing.T) {
		testGlobIteratorEmptyPattern(t, ctx)
	})

	t.Run("GlobIteratorNoMatches", func(t *testing.T) {
		testGlobIteratorNoMatches(t, ctx)
	})

	t.Run("GlobIteratorInheritance", func(t *testing.T) {
		testGlobIteratorInheritance(t, ctx)
	})
}

func testBasicGlobIterator(t *testing.T, ctx *mockContext) {
	// Create test directory structure
	testDir := "/tmp/test_glob_basic"
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	// Create test files
	testFiles := []string{"test1.txt", "test2.txt", "data.dat", "script.php"}
	for _, filename := range testFiles {
		file, err := os.Create(testDir + "/" + filename)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
		file.Close()
	}

	// Get the GlobIterator class
	class, err := ctx.registry.GetClass("GlobIterator")
	if err != nil {
		t.Fatalf("GlobIterator class not found: %v", err)
	}

	// Create GlobIterator instance
	obj := &values.Object{
		ClassName:  "GlobIterator",
		Properties: make(map[string]*values.Value),
	}
	thisObj := &values.Value{
		Type: values.TypeObject,
		Data: obj,
	}

	// Test constructor with *.txt pattern
	constructorMethod := class.Methods["__construct"]
	constructorImpl := constructorMethod.Implementation.(*BuiltinMethodImpl)
	pattern := testDir + "/*.txt"
	_, err = constructorImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj, values.NewString(pattern)})
	if err != nil {
		t.Fatalf("Constructor failed: %v", err)
	}

	// Test count method
	countMethod := class.Methods["count"]
	countImpl := countMethod.Implementation.(*BuiltinMethodImpl)
	countResult, err := countImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}

	if countResult.ToInt() != 2 {
		t.Errorf("Expected count of 2 for *.txt files, got %d", countResult.ToInt())
	}
}

func testGlobPatternMatching(t *testing.T, ctx *mockContext) {
	// Create test directory structure
	testDir := "/tmp/test_glob_patterns"
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	// Create test files with different extensions
	testFiles := map[string]int{
		"*.txt": 3,
		"*.dat": 2,
		"*.php": 1,
		"*":     6, // All files
	}

	// Create the actual files
	for i := 1; i <= 3; i++ {
		file, _ := os.Create(testDir + fmt.Sprintf("/file%d.txt", i))
		file.Close()
	}
	for i := 1; i <= 2; i++ {
		file, _ := os.Create(testDir + fmt.Sprintf("/data%d.dat", i))
		file.Close()
	}
	file, _ := os.Create(testDir + "/script.php")
	file.Close()

	class, err := ctx.registry.GetClass("GlobIterator")
	if err != nil {
		t.Fatalf("GlobIterator class not found: %v", err)
	}

	for pattern, expectedCount := range testFiles {
		t.Run(pattern, func(t *testing.T) {
			obj := &values.Object{
				ClassName:  "GlobIterator",
				Properties: make(map[string]*values.Value),
			}
			thisObj := &values.Value{
				Type: values.TypeObject,
				Data: obj,
			}

			// Create iterator with pattern
			constructorMethod := class.Methods["__construct"]
			constructorImpl := constructorMethod.Implementation.(*BuiltinMethodImpl)
			fullPattern := testDir + "/" + pattern
			_, err = constructorImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj, values.NewString(fullPattern)})
			if err != nil {
				t.Fatalf("Constructor failed for pattern %s: %v", pattern, err)
			}

			// Test count
			countMethod := class.Methods["count"]
			countImpl := countMethod.Implementation.(*BuiltinMethodImpl)
			countResult, err := countImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
			if err != nil {
				t.Fatalf("Count failed for pattern %s: %v", pattern, err)
			}

			if countResult.ToInt() != int64(expectedCount) {
				t.Errorf("Pattern %s: expected count %d, got %d", pattern, expectedCount, countResult.ToInt())
			}
		})
	}
}

func testGlobIteratorCount(t *testing.T, ctx *mockContext) {
	// Create test directory
	testDir := "/tmp/test_glob_count"
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	// Create exactly 5 .txt files
	for i := 1; i <= 5; i++ {
		file, err := os.Create(testDir + fmt.Sprintf("/test%d.txt", i))
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		file.Close()
	}

	class, err := ctx.registry.GetClass("GlobIterator")
	if err != nil {
		t.Fatalf("GlobIterator class not found: %v", err)
	}

	obj := &values.Object{
		ClassName:  "GlobIterator",
		Properties: make(map[string]*values.Value),
	}
	thisObj := &values.Value{
		Type: values.TypeObject,
		Data: obj,
	}

	constructorMethod := class.Methods["__construct"]
	constructorImpl := constructorMethod.Implementation.(*BuiltinMethodImpl)
	pattern := testDir + "/*.txt"
	_, err = constructorImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj, values.NewString(pattern)})
	if err != nil {
		t.Fatalf("Constructor failed: %v", err)
	}

	countMethod := class.Methods["count"]
	countImpl := countMethod.Implementation.(*BuiltinMethodImpl)
	countResult, err := countImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}

	if countResult.ToInt() != 5 {
		t.Errorf("Expected count of 5, got %d", countResult.ToInt())
	}

	// Test Countable interface - count should be accessible
	// This tests that GlobIterator implements Countable correctly
	if countResult.ToInt() <= 0 {
		t.Error("Count should be positive when files match")
	}
}

func testGlobIteratorIteration(t *testing.T, ctx *mockContext) {
	// Create test directory
	testDir := "/tmp/test_glob_iteration"
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	// Create test files in specific order
	testFiles := []string{"a.txt", "b.txt", "c.txt"}
	for _, filename := range testFiles {
		file, err := os.Create(testDir + "/" + filename)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
		file.Close()
	}

	class, err := ctx.registry.GetClass("GlobIterator")
	if err != nil {
		t.Fatalf("GlobIterator class not found: %v", err)
	}

	obj := &values.Object{
		ClassName:  "GlobIterator",
		Properties: make(map[string]*values.Value),
	}
	thisObj := &values.Value{
		Type: values.TypeObject,
		Data: obj,
	}

	constructorMethod := class.Methods["__construct"]
	constructorImpl := constructorMethod.Implementation.(*BuiltinMethodImpl)
	pattern := testDir + "/*.txt"
	_, err = constructorImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj, values.NewString(pattern)})
	if err != nil {
		t.Fatalf("Constructor failed: %v", err)
	}

	// Get methods
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

	// Test iteration
	_, err = rewindImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
	if err != nil {
		t.Fatalf("Rewind failed: %v", err)
	}

	iterationCount := 0
	expectedFiles := []string{"a.txt", "b.txt", "c.txt"}

	for {
		validResult, err := validImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("Valid failed: %v", err)
		}

		if !validResult.ToBool() {
			break
		}

		// Test current() returns SplFileInfo object
		currentResult, err := currentImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("Current failed: %v", err)
		}

		if currentResult.Type != values.TypeObject {
			t.Errorf("Current should return SplFileInfo object, got %v", currentResult.Type)
		}

		// Test key() returns file path
		keyResult, err := keyImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("Key failed: %v", err)
		}

		// Key should contain the expected filename
		keyStr := keyResult.ToString()
		if iterationCount < len(expectedFiles) {
			if !contains(keyStr, expectedFiles[iterationCount]) {
				t.Errorf("Expected key to contain %s, got %s", expectedFiles[iterationCount], keyStr)
			}
		}

		iterationCount++

		_, err = nextImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("Next failed: %v", err)
		}
	}

	if iterationCount != len(expectedFiles) {
		t.Errorf("Expected %d iterations, got %d", len(expectedFiles), iterationCount)
	}
}

func testGlobIteratorFlags(t *testing.T, ctx *mockContext) {
	// Create test directory
	testDir := "/tmp/test_glob_flags"
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	file, _ := os.Create(testDir + "/test.txt")
	file.Close()

	class, err := ctx.registry.GetClass("GlobIterator")
	if err != nil {
		t.Fatalf("GlobIterator class not found: %v", err)
	}

	obj := &values.Object{
		ClassName:  "GlobIterator",
		Properties: make(map[string]*values.Value),
	}
	thisObj := &values.Value{
		Type: values.TypeObject,
		Data: obj,
	}

	// Test constructor with flags
	constructorMethod := class.Methods["__construct"]
	constructorImpl := constructorMethod.Implementation.(*BuiltinMethodImpl)
	pattern := testDir + "/*.txt"
	flags := int64(1) // Some flag value
	_, err = constructorImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj, values.NewString(pattern), values.NewInt(flags)})
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

	// Test setFlags method
	setFlagsMethod := class.Methods["setFlags"]
	setFlagsImpl := setFlagsMethod.Implementation.(*BuiltinMethodImpl)
	newFlags := int64(42)
	_, err = setFlagsImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj, values.NewInt(newFlags)})
	if err != nil {
		t.Fatalf("setFlags failed: %v", err)
	}

	// Verify flags were set
	flagsResult, err = getFlagsImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
	if err != nil {
		t.Fatalf("getFlags after setFlags failed: %v", err)
	}

	if flagsResult.ToInt() != newFlags {
		t.Errorf("Expected flags %d after setFlags, got %d", newFlags, flagsResult.ToInt())
	}
}

func testGlobIteratorEmptyPattern(t *testing.T, ctx *mockContext) {
	class, err := ctx.registry.GetClass("GlobIterator")
	if err != nil {
		t.Fatalf("GlobIterator class not found: %v", err)
	}

	obj := &values.Object{
		ClassName:  "GlobIterator",
		Properties: make(map[string]*values.Value),
	}
	thisObj := &values.Value{
		Type: values.TypeObject,
		Data: obj,
	}

	constructorMethod := class.Methods["__construct"]
	constructorImpl := constructorMethod.Implementation.(*BuiltinMethodImpl)

	// Test empty pattern should fail
	_, err = constructorImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj, values.NewString("")})
	if err == nil {
		t.Error("Expected error for empty pattern, but got none")
	}

	if !contains(err.Error(), "must not be empty") {
		t.Errorf("Expected error about empty pattern, got: %s", err.Error())
	}
}

func testGlobIteratorNoMatches(t *testing.T, ctx *mockContext) {
	class, err := ctx.registry.GetClass("GlobIterator")
	if err != nil {
		t.Fatalf("GlobIterator class not found: %v", err)
	}

	obj := &values.Object{
		ClassName:  "GlobIterator",
		Properties: make(map[string]*values.Value),
	}
	thisObj := &values.Value{
		Type: values.TypeObject,
		Data: obj,
	}

	constructorMethod := class.Methods["__construct"]
	constructorImpl := constructorMethod.Implementation.(*BuiltinMethodImpl)

	// Use pattern that won't match anything
	pattern := "/nonexistent/path/*.xyz"
	_, err = constructorImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj, values.NewString(pattern)})
	if err != nil {
		t.Fatalf("Constructor failed for non-matching pattern: %v", err)
	}

	// Test count should be 0
	countMethod := class.Methods["count"]
	countImpl := countMethod.Implementation.(*BuiltinMethodImpl)
	countResult, err := countImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}

	if countResult.ToInt() != 0 {
		t.Errorf("Expected count of 0 for non-matching pattern, got %d", countResult.ToInt())
	}

	// Test valid should be false
	validMethod := class.Methods["valid"]
	validImpl := validMethod.Implementation.(*BuiltinMethodImpl)
	validResult, err := validImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
	if err != nil {
		t.Fatalf("Valid failed: %v", err)
	}

	if validResult.ToBool() {
		t.Error("Expected valid to be false for non-matching pattern")
	}
}

func testGlobIteratorInheritance(t *testing.T, ctx *mockContext) {
	class, err := ctx.registry.GetClass("GlobIterator")
	if err != nil {
		t.Fatalf("GlobIterator class not found: %v", err)
	}

	// Test parent class
	if class.Parent != "FilesystemIterator" {
		t.Errorf("GlobIterator parent should be FilesystemIterator, got %s", class.Parent)
	}

	// Test interfaces
	expectedInterfaces := []string{"Iterator", "SeekableIterator", "Countable"}
	for _, interfaceName := range expectedInterfaces {
		found := false
		for _, iface := range class.Interfaces {
			if iface == interfaceName {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("GlobIterator should implement interface %s", interfaceName)
		}
	}

	// Test that it has required methods
	requiredMethods := []string{"__construct", "count", "current", "key", "next", "rewind", "valid", "getFlags", "setFlags"}
	for _, methodName := range requiredMethods {
		if _, exists := class.Methods[methodName]; !exists {
			t.Errorf("GlobIterator should have method %s", methodName)
		}
	}

	// Test that it has constants from FilesystemIterator
	if len(class.Constants) == 0 {
		t.Error("GlobIterator should inherit constants from FilesystemIterator")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}