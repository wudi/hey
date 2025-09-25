package spl

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

func TestDirectoryIterator(t *testing.T) {
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

	t.Run("BasicDirectoryIteration", func(t *testing.T) {
		testBasicDirectoryIteration(t, ctx)
	})

	t.Run("DirectoryIteratorMethods", func(t *testing.T) {
		testDirectoryIteratorMethods(t, ctx)
	})

	t.Run("DirectoryIteratorEdgeCases", func(t *testing.T) {
		testDirectoryIteratorEdgeCases(t, ctx)
	})
}

func testBasicDirectoryIteration(t *testing.T, ctx *mockContext) {
	// Create test directory structure
	testDir := "/tmp/test_directory_iterator_go"
	err := os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	// Create test files and subdirectory
	testFiles := []string{
		"file1.txt",
		"file2.php",
	}
	testContent := []string{
		"content1",
		"<?php echo \"test\"; ?>",
	}

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

	// Get the DirectoryIterator class
	class, err := ctx.registry.GetClass("DirectoryIterator")
	if err != nil {
		t.Fatalf("DirectoryIterator class not found: %v", err)
	}

	// Create DirectoryIterator instance
	obj := &values.Object{
		ClassName:  "DirectoryIterator",
		Properties: make(map[string]*values.Value),
	}
	thisObj := &values.Value{
		Type: values.TypeObject,
		Data: obj,
	}

	// Test constructor
	constructorMethod := class.Methods["__construct"]
	if constructorMethod == nil {
		t.Fatal("__construct method not found")
	}
	constructorImpl := constructorMethod.Implementation.(*BuiltinMethodImpl)
	_, err = constructorImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj, values.NewString(testDir)})
	if err != nil {
		t.Fatalf("Constructor failed: %v", err)
	}

	// Test iteration
	validMethod := class.Methods["valid"]
	nextMethod := class.Methods["next"]
	rewindMethod := class.Methods["rewind"]
	getFilenameMethod := class.Methods["getFilename"]

	if validMethod == nil || nextMethod == nil || rewindMethod == nil || getFilenameMethod == nil {
		t.Fatal("Required iterator methods not found")
	}

	validImpl := validMethod.Implementation.(*BuiltinMethodImpl)
	nextImpl := nextMethod.Implementation.(*BuiltinMethodImpl)
	rewindImpl := rewindMethod.Implementation.(*BuiltinMethodImpl)
	getFilenameImpl := getFilenameMethod.Implementation.(*BuiltinMethodImpl)

	// Rewind to start
	_, err = rewindImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
	if err != nil {
		t.Fatalf("Rewind failed: %v", err)
	}

	// Collect all entries
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

	// Should have at least: ., .., file1.txt, file2.php, subdir
	if len(entries) < 5 {
		t.Errorf("Expected at least 5 entries, got %d: %v", len(entries), entries)
	}

	// Check for expected entries
	expectedEntries := []string{".", "..", "file1.txt", "file2.php", "subdir"}
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
}

func testDirectoryIteratorMethods(t *testing.T, ctx *mockContext) {
	// Create test directory structure
	testDir := "/tmp/test_directory_iterator_methods"
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

	// Get the DirectoryIterator class
	class, err := ctx.registry.GetClass("DirectoryIterator")
	if err != nil {
		t.Fatalf("DirectoryIterator class not found: %v", err)
	}

	// Create DirectoryIterator instance
	obj := &values.Object{
		ClassName:  "DirectoryIterator",
		Properties: make(map[string]*values.Value),
	}
	thisObj := &values.Value{
		Type: values.TypeObject,
		Data: obj,
	}

	// Initialize iterator
	constructorMethod := class.Methods["__construct"]
	constructorImpl := constructorMethod.Implementation.(*BuiltinMethodImpl)
	_, err = constructorImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj, values.NewString(testDir)})
	if err != nil {
		t.Fatalf("Constructor failed: %v", err)
	}

	// Test various methods on each entry
	methods := map[string]*BuiltinMethodImpl{
		"valid":       class.Methods["valid"].Implementation.(*BuiltinMethodImpl),
		"key":         class.Methods["key"].Implementation.(*BuiltinMethodImpl),
		"current":     class.Methods["current"].Implementation.(*BuiltinMethodImpl),
		"next":        class.Methods["next"].Implementation.(*BuiltinMethodImpl),
		"rewind":      class.Methods["rewind"].Implementation.(*BuiltinMethodImpl),
		"getFilename": class.Methods["getFilename"].Implementation.(*BuiltinMethodImpl),
		"getPathname": class.Methods["getPathname"].Implementation.(*BuiltinMethodImpl),
		"getSize":     class.Methods["getSize"].Implementation.(*BuiltinMethodImpl),
		"getType":     class.Methods["getType"].Implementation.(*BuiltinMethodImpl),
		"getPerms":    class.Methods["getPerms"].Implementation.(*BuiltinMethodImpl),
		"isDir":       class.Methods["isDir"].Implementation.(*BuiltinMethodImpl),
		"isFile":      class.Methods["isFile"].Implementation.(*BuiltinMethodImpl),
		"isDot":       class.Methods["isDot"].Implementation.(*BuiltinMethodImpl),
		"isReadable":  class.Methods["isReadable"].Implementation.(*BuiltinMethodImpl),
	}

	// Start iteration
	_, err = methods["rewind"].GetFunction().Builtin(ctx, []*values.Value{thisObj})
	if err != nil {
		t.Fatalf("Rewind failed: %v", err)
	}

	iterationCount := 0
	for {
		validResult, err := methods["valid"].GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("Valid check failed: %v", err)
		}
		if !validResult.ToBool() {
			break
		}

		// Test key() method
		keyResult, err := methods["key"].GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("Key failed: %v", err)
		}
		if keyResult.ToInt() != int64(iterationCount) {
			t.Errorf("Expected key %d, got %d", iterationCount, keyResult.ToInt())
		}

		// Test current() method
		currentResult, err := methods["current"].GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("Current failed: %v", err)
		}
		if currentResult != thisObj {
			t.Error("Current should return the DirectoryIterator object itself")
		}

		// Test getFilename()
		filenameResult, err := methods["getFilename"].GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("GetFilename failed: %v", err)
		}
		filename := filenameResult.ToString()

		// Test getPathname()
		pathnameResult, err := methods["getPathname"].GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("GetPathname failed: %v", err)
		}
		pathname := pathnameResult.ToString()

		// Pathname should be testDir + filename
		expectedPathname := filepath.Join(testDir, filename)
		if pathname != expectedPathname {
			t.Errorf("Expected pathname %s, got %s", expectedPathname, pathname)
		}

		// Test isDot()
		isDotResult, err := methods["isDot"].GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("IsDot failed: %v", err)
		}
		isDot := isDotResult.ToBool()

		if filename == "." || filename == ".." {
			if !isDot {
				t.Errorf("Expected isDot() to be true for %s", filename)
			}
		} else {
			if isDot {
				t.Errorf("Expected isDot() to be false for %s", filename)
			}
		}

		// Test isDir() and isFile()
		isDirResult, err := methods["isDir"].GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("IsDir failed: %v", err)
		}
		isFileResult, err := methods["isFile"].GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("IsFile failed: %v", err)
		}

		isDir := isDirResult.ToBool()
		isFile := isFileResult.ToBool()

		// Should be either dir or file, not both
		if isDir && isFile {
			t.Error("Entry cannot be both directory and file")
		}
		if !isDir && !isFile {
			t.Error("Entry must be either directory or file")
		}

		// Test getType()
		getTypeResult, err := methods["getType"].GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("GetType failed: %v", err)
		}
		entryType := getTypeResult.ToString()

		if isDir && entryType != "dir" {
			t.Errorf("Expected type 'dir' for directory, got %s", entryType)
		}
		if isFile && entryType != "file" {
			t.Errorf("Expected type 'file' for file, got %s", entryType)
		}

		// Test isReadable()
		isReadableResult, err := methods["isReadable"].GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("IsReadable failed: %v", err)
		}
		// Should generally be readable
		if !isReadableResult.ToBool() {
			t.Logf("Warning: Entry %s is not readable", filename)
		}

		// Test getSize() and getPerms() for non-dot entries
		if !isDot {
			getSizeResult, err := methods["getSize"].GetFunction().Builtin(ctx, []*values.Value{thisObj})
			if err != nil {
				t.Fatalf("GetSize failed: %v", err)
			}
			size := getSizeResult.ToInt()

			if filename == "test.txt" && size != 12 { // "test content" is 12 bytes
				t.Errorf("Expected size 12 for test.txt, got %d", size)
			}

			getPermsResult, err := methods["getPerms"].GetFunction().Builtin(ctx, []*values.Value{thisObj})
			if err != nil {
				t.Fatalf("GetPerms failed: %v", err)
			}
			perms := getPermsResult.ToInt()

			if perms == 0 {
				t.Error("Expected non-zero permissions")
			}
		}

		_, err = methods["next"].GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("Next failed: %v", err)
		}

		iterationCount++
	}

	if iterationCount == 0 {
		t.Error("No entries found during iteration")
	}
}

func testDirectoryIteratorEdgeCases(t *testing.T, ctx *mockContext) {
	// Get the DirectoryIterator class
	class, err := ctx.registry.GetClass("DirectoryIterator")
	if err != nil {
		t.Fatalf("DirectoryIterator class not found: %v", err)
	}

	// Test non-existent directory
	t.Run("NonExistentDirectory", func(t *testing.T) {
		obj := &values.Object{
			ClassName:  "DirectoryIterator",
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

	// Test invalid path (file instead of directory)
	t.Run("FileInsteadOfDirectory", func(t *testing.T) {
		// Create a temporary file
		testFile := "/tmp/test_file_not_dir"
		err := os.WriteFile(testFile, []byte("content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		defer os.Remove(testFile)

		obj := &values.Object{
			ClassName:  "DirectoryIterator",
			Properties: make(map[string]*values.Value),
		}
		thisObj := &values.Value{
			Type: values.TypeObject,
			Data: obj,
		}

		constructorMethod := class.Methods["__construct"]
		constructorImpl := constructorMethod.Implementation.(*BuiltinMethodImpl)
		_, err = constructorImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj, values.NewString(testFile)})
		if err == nil {
			t.Error("Expected error when opening file as directory")
		}
	})

	// Test empty directory
	t.Run("EmptyDirectory", func(t *testing.T) {
		emptyDir := "/tmp/empty_dir_test"
		err := os.MkdirAll(emptyDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create empty directory: %v", err)
		}
		defer os.RemoveAll(emptyDir)

		obj := &values.Object{
			ClassName:  "DirectoryIterator",
			Properties: make(map[string]*values.Value),
		}
		thisObj := &values.Value{
			Type: values.TypeObject,
			Data: obj,
		}

		constructorMethod := class.Methods["__construct"]
		constructorImpl := constructorMethod.Implementation.(*BuiltinMethodImpl)
		_, err = constructorImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj, values.NewString(emptyDir)})
		if err != nil {
			t.Fatalf("Constructor failed for empty directory: %v", err)
		}

		// Should still have . and .. entries
		validMethod := class.Methods["valid"]
		validImpl := validMethod.Implementation.(*BuiltinMethodImpl)

		validResult, err := validImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("Valid check failed: %v", err)
		}
		if !validResult.ToBool() {
			t.Error("Expected valid to be true for empty directory (should have . and .. entries)")
		}
	})
}