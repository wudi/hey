package spl

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

func TestFilesystemIterator(t *testing.T) {
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

	t.Run("BasicFilesystemIteration", func(t *testing.T) {
		testBasicFilesystemIteration(t, ctx)
	})

	t.Run("FilesystemIteratorFlags", func(t *testing.T) {
		testFilesystemIteratorFlags(t, ctx)
	})

	t.Run("FilesystemIteratorMethods", func(t *testing.T) {
		testFilesystemIteratorMethods(t, ctx)
	})

	t.Run("FilesystemIteratorConstants", func(t *testing.T) {
		testFilesystemIteratorConstants(t, ctx)
	})

	t.Run("FilesystemIteratorEdgeCases", func(t *testing.T) {
		testFilesystemIteratorEdgeCases(t, ctx)
	})
}

func testBasicFilesystemIteration(t *testing.T, ctx *mockContext) {
	// Create test directory structure
	testDir := "/tmp/test_filesystem_iterator_go"
	err := os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	// Create test files and subdirectory
	testFiles := []string{"file1.txt", "file2.php", ".hidden"}
	testContent := []string{"content1", "<?php echo \"test\"; ?>", "hidden"}

	for i, filename := range testFiles {
		err := os.WriteFile(filepath.Join(testDir, filename), []byte(testContent[i]), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	// Create subdirectory
	subdir := filepath.Join(testDir, "subdir")
	err = os.Mkdir(subdir, 0755)
	if err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	// Get the FilesystemIterator class
	class, err := ctx.registry.GetClass("FilesystemIterator")
	if err != nil {
		t.Fatalf("FilesystemIterator class not found: %v", err)
	}

	// Create FilesystemIterator instance (default flags)
	obj := &values.Object{
		ClassName:  "FilesystemIterator",
		Properties: make(map[string]*values.Value),
	}
	thisObj := &values.Value{
		Type: values.TypeObject,
		Data: obj,
	}

	// Test constructor with default flags (SKIP_DOTS)
	constructorMethod := class.Methods["__construct"]
	if constructorMethod == nil {
		t.Fatal("__construct method not found")
	}
	constructorImpl := constructorMethod.Implementation.(*BuiltinMethodImpl)
	_, err = constructorImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj, values.NewString(testDir)})
	if err != nil {
		t.Fatalf("Constructor failed: %v", err)
	}

	// Test iteration (should skip dots by default)
	validMethod := class.Methods["valid"]
	nextMethod := class.Methods["next"]
	rewindMethod := class.Methods["rewind"]
	keyMethod := class.Methods["key"]
	currentMethod := class.Methods["current"]
	getFilenameMethod := class.Methods["getFilename"]

	if validMethod == nil || nextMethod == nil || rewindMethod == nil || keyMethod == nil || currentMethod == nil || getFilenameMethod == nil {
		t.Fatal("Required iterator methods not found")
	}

	validImpl := validMethod.Implementation.(*BuiltinMethodImpl)
	nextImpl := nextMethod.Implementation.(*BuiltinMethodImpl)
	rewindImpl := rewindMethod.Implementation.(*BuiltinMethodImpl)
	keyImpl := keyMethod.Implementation.(*BuiltinMethodImpl)
	currentImpl := currentMethod.Implementation.(*BuiltinMethodImpl)
	getFilenameImpl := getFilenameMethod.Implementation.(*BuiltinMethodImpl)

	// Rewind to start
	_, err = rewindImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
	if err != nil {
		t.Fatalf("Rewind failed: %v", err)
	}

	// Collect all entries
	var entries []string
	var keys []string
	for {
		validResult, err := validImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("Valid check failed: %v", err)
		}
		if !validResult.ToBool() {
			break
		}

		// Get key (should be pathname by default)
		keyResult, err := keyImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("Key failed: %v", err)
		}
		keys = append(keys, keyResult.ToString())

		// Get current (should be fileinfo by default)
		currentResult, err := currentImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("Current failed: %v", err)
		}

		// Current should return the FilesystemIterator object itself
		if currentResult != thisObj {
			t.Error("Current should return the FilesystemIterator object itself")
		}

		// Get filename
		filenameResult, err := getFilenameImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("GetFilename failed: %v", err)
		}
		entries = append(entries, filenameResult.ToString())

		_, err = nextImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("Next failed: %v", err)
		}
	}

	// Should NOT include . and .. (SKIP_DOTS is default)
	// Should include: file1.txt, file2.php, .hidden, subdir
	if len(entries) != 4 {
		t.Errorf("Expected 4 entries (no dots), got %d: %v", len(entries), entries)
	}

	// Check for expected entries
	expectedEntries := []string{"file1.txt", "file2.php", ".hidden", "subdir"}
	for _, expected := range expectedEntries {
		found := false
		for _, entry := range entries {
			if entry == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected entry %s not found in entries: %v", expected, entries)
		}
	}

	// Check that . and .. are NOT included
	for _, entry := range entries {
		if entry == "." || entry == ".." {
			t.Errorf("Unexpected dot entry %s found (should be skipped by default)", entry)
		}
	}

	// Keys should be full pathnames
	for i, key := range keys {
		expectedKey := filepath.Join(testDir, entries[i])
		if key != expectedKey {
			t.Errorf("Expected key %s, got %s", expectedKey, key)
		}
	}
}

func testFilesystemIteratorFlags(t *testing.T, ctx *mockContext) {
	// Create test directory
	testDir := "/tmp/test_filesystem_iterator_flags"
	err := os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	// Create a test file
	testFile := filepath.Join(testDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Get the FilesystemIterator class
	class, err := ctx.registry.GetClass("FilesystemIterator")
	if err != nil {
		t.Fatalf("FilesystemIterator class not found: %v", err)
	}

	// Test with CURRENT_AS_PATHNAME flag
	t.Run("CURRENT_AS_PATHNAME", func(t *testing.T) {
		obj := &values.Object{
			ClassName:  "FilesystemIterator",
			Properties: make(map[string]*values.Value),
		}
		thisObj := &values.Value{
			Type: values.TypeObject,
			Data: obj,
		}

		constructorMethod := class.Methods["__construct"]
		constructorImpl := constructorMethod.Implementation.(*BuiltinMethodImpl)
		flags := CURRENT_AS_PATHNAME | SKIP_DOTS
		_, err = constructorImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj, values.NewString(testDir), values.NewInt(int64(flags))})
		if err != nil {
			t.Fatalf("Constructor failed: %v", err)
		}

		currentMethod := class.Methods["current"]
		currentImpl := currentMethod.Implementation.(*BuiltinMethodImpl)

		currentResult, err := currentImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("Current failed: %v", err)
		}

		// Should return pathname string
		expectedPath := filepath.Join(testDir, "test.txt")
		if currentResult.ToString() != expectedPath {
			t.Errorf("Expected current to return pathname %s, got %s", expectedPath, currentResult.ToString())
		}
	})

	// Test with KEY_AS_FILENAME flag
	t.Run("KEY_AS_FILENAME", func(t *testing.T) {
		obj := &values.Object{
			ClassName:  "FilesystemIterator",
			Properties: make(map[string]*values.Value),
		}
		thisObj := &values.Value{
			Type: values.TypeObject,
			Data: obj,
		}

		constructorMethod := class.Methods["__construct"]
		constructorImpl := constructorMethod.Implementation.(*BuiltinMethodImpl)
		flags := KEY_AS_FILENAME | SKIP_DOTS
		_, err = constructorImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj, values.NewString(testDir), values.NewInt(int64(flags))})
		if err != nil {
			t.Fatalf("Constructor failed: %v", err)
		}

		keyMethod := class.Methods["key"]
		keyImpl := keyMethod.Implementation.(*BuiltinMethodImpl)

		keyResult, err := keyImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("Key failed: %v", err)
		}

		// Should return filename only
		if keyResult.ToString() != "test.txt" {
			t.Errorf("Expected key to return filename 'test.txt', got %s", keyResult.ToString())
		}
	})

	// Test without SKIP_DOTS flag
	t.Run("Without_SKIP_DOTS", func(t *testing.T) {
		obj := &values.Object{
			ClassName:  "FilesystemIterator",
			Properties: make(map[string]*values.Value),
		}
		thisObj := &values.Value{
			Type: values.TypeObject,
			Data: obj,
		}

		constructorMethod := class.Methods["__construct"]
		constructorImpl := constructorMethod.Implementation.(*BuiltinMethodImpl)
		flags := 0 // No SKIP_DOTS
		_, err = constructorImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj, values.NewString(testDir), values.NewInt(int64(flags))})
		if err != nil {
			t.Fatalf("Constructor failed: %v", err)
		}

		// Count entries - should include . and .. now
		validMethod := class.Methods["valid"]
		nextMethod := class.Methods["next"]
		getFilenameMethod := class.Methods["getFilename"]
		rewindMethod := class.Methods["rewind"]

		validImpl := validMethod.Implementation.(*BuiltinMethodImpl)
		nextImpl := nextMethod.Implementation.(*BuiltinMethodImpl)
		getFilenameImpl := getFilenameMethod.Implementation.(*BuiltinMethodImpl)
		rewindImpl := rewindMethod.Implementation.(*BuiltinMethodImpl)

		_, err = rewindImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("Rewind failed: %v", err)
		}

		var entries []string
		for {
			validResult, err := validImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
			if err != nil {
				t.Fatalf("Valid check failed: %v", err)
			}
			if !validResult.ToBool() {
				break
			}

			filenameResult, err := getFilenameImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
			if err != nil {
				t.Fatalf("GetFilename failed: %v", err)
			}
			entries = append(entries, filenameResult.ToString())

			_, err = nextImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
			if err != nil {
				t.Fatalf("Next failed: %v", err)
			}
		}

		// Should include ., .., and test.txt
		if len(entries) < 3 {
			t.Errorf("Expected at least 3 entries (including dots), got %d: %v", len(entries), entries)
		}

		// Check for . and .. entries
		foundDot := false
		foundDotDot := false
		for _, entry := range entries {
			if entry == "." {
				foundDot = true
			}
			if entry == ".." {
				foundDotDot = true
			}
		}
		if !foundDot {
			t.Error("Expected '.' entry not found")
		}
		if !foundDotDot {
			t.Error("Expected '..' entry not found")
		}
	})
}

func testFilesystemIteratorMethods(t *testing.T, ctx *mockContext) {
	// Create test directory
	testDir := "/tmp/test_filesystem_iterator_methods"
	err := os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	// Get the FilesystemIterator class
	class, err := ctx.registry.GetClass("FilesystemIterator")
	if err != nil {
		t.Fatalf("FilesystemIterator class not found: %v", err)
	}

	// Create FilesystemIterator instance
	obj := &values.Object{
		ClassName:  "FilesystemIterator",
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

	// Test getFlags() and setFlags()
	getFlagsMethod := class.Methods["getFlags"]
	setFlagsMethod := class.Methods["setFlags"]

	if getFlagsMethod == nil || setFlagsMethod == nil {
		t.Fatal("getFlags or setFlags method not found")
	}

	getFlagsImpl := getFlagsMethod.Implementation.(*BuiltinMethodImpl)
	setFlagsImpl := setFlagsMethod.Implementation.(*BuiltinMethodImpl)

	// Test default flags
	flagsResult, err := getFlagsImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
	if err != nil {
		t.Fatalf("GetFlags failed: %v", err)
	}
	defaultFlags := flagsResult.ToInt()
	if defaultFlags != SKIP_DOTS {
		t.Errorf("Expected default flags to be SKIP_DOTS (%d), got %d", SKIP_DOTS, defaultFlags)
	}

	// Test setFlags
	newFlags := CURRENT_AS_PATHNAME | KEY_AS_FILENAME
	_, err = setFlagsImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj, values.NewInt(int64(newFlags))})
	if err != nil {
		t.Fatalf("SetFlags failed: %v", err)
	}

	// Verify flags were set
	flagsResult, err = getFlagsImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
	if err != nil {
		t.Fatalf("GetFlags failed: %v", err)
	}
	if flagsResult.ToInt() != int64(newFlags) {
		t.Errorf("Expected flags to be %d, got %d", newFlags, flagsResult.ToInt())
	}
}

func testFilesystemIteratorConstants(t *testing.T, ctx *mockContext) {
	// Get the FilesystemIterator class
	class, err := ctx.registry.GetClass("FilesystemIterator")
	if err != nil {
		t.Fatalf("FilesystemIterator class not found: %v", err)
	}

	// Test all constants
	expectedConstants := map[string]int{
		"CURRENT_AS_PATHNAME": CURRENT_AS_PATHNAME,
		"CURRENT_AS_FILEINFO": CURRENT_AS_FILEINFO,
		"CURRENT_AS_SELF":     CURRENT_AS_SELF,
		"CURRENT_MODE_MASK":   CURRENT_MODE_MASK,
		"KEY_AS_PATHNAME":     KEY_AS_PATHNAME,
		"KEY_AS_FILENAME":     KEY_AS_FILENAME,
		"FOLLOW_SYMLINKS":     FOLLOW_SYMLINKS,
		"KEY_MODE_MASK":       KEY_MODE_MASK,
		"NEW_CURRENT_AND_KEY": NEW_CURRENT_AND_KEY,
		"SKIP_DOTS":           SKIP_DOTS,
		"UNIX_PATHS":          UNIX_PATHS,
	}

	for name, expectedValue := range expectedConstants {
		constant, exists := class.Constants[name]
		if !exists {
			t.Errorf("Constant %s not found", name)
			continue
		}

		if constant.Value.ToInt() != int64(expectedValue) {
			t.Errorf("Constant %s has wrong value: expected %d, got %d",
				name, expectedValue, constant.Value.ToInt())
		}

		if !constant.IsFinal {
			t.Errorf("Constant %s should be final", name)
		}

		if constant.Visibility != "public" {
			t.Errorf("Constant %s should be public", name)
		}
	}
}

func testFilesystemIteratorEdgeCases(t *testing.T, ctx *mockContext) {
	// Get the FilesystemIterator class
	class, err := ctx.registry.GetClass("FilesystemIterator")
	if err != nil {
		t.Fatalf("FilesystemIterator class not found: %v", err)
	}

	// Test non-existent directory
	t.Run("NonExistentDirectory", func(t *testing.T) {
		obj := &values.Object{
			ClassName:  "FilesystemIterator",
			Properties: make(map[string]*values.Value),
		}
		thisObj := &values.Value{
			Type: values.TypeObject,
			Data: obj,
		}

		constructorMethod := class.Methods["__construct"]
		constructorImpl := constructorMethod.Implementation.(*BuiltinMethodImpl)
		_, err := constructorImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj, values.NewString("/non/existent/directory")})
		if err == nil {
			t.Error("Expected error when opening non-existent directory")
		}
	})

	// Test inheritance from DirectoryIterator
	t.Run("InheritanceFromDirectoryIterator", func(t *testing.T) {
		// FilesystemIterator should inherit methods from DirectoryIterator
		requiredMethods := []string{"getFilename", "getPathname", "getSize", "getType", "isDir", "isFile", "isDot", "isReadable"}
		for _, methodName := range requiredMethods {
			if _, exists := class.Methods[methodName]; !exists {
				t.Errorf("FilesystemIterator should inherit method %s from DirectoryIterator", methodName)
			}
		}

		// Check parent class
		if class.Parent != "DirectoryIterator" {
			t.Errorf("FilesystemIterator parent should be DirectoryIterator, got %s", class.Parent)
		}
	})
}