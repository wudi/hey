package spl

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

func TestSplFileInfo(t *testing.T) {
	// Initialize registry
	registry.Initialize()
	ctx := &mockContext{registry: registry.GlobalRegistry}

	// Create test directory and file
	testDir := "/tmp/claude_test"
	testFile := filepath.Join(testDir, "test_file.txt")

	err := os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	err = os.WriteFile(testFile, []byte("Hello, SPL FileInfo test!"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Get the SplFileInfo class
	class := GetSplFileInfoClass()
	if class == nil {
		t.Fatal("SplFileInfo class is nil")
	}

	// Create a new SplFileInfo instance
	obj := &values.Object{
		ClassName:  "SplFileInfo",
		Properties: make(map[string]*values.Value),
	}
	thisObj := &values.Value{
		Type: values.TypeObject,
		Data: obj,
	}

	t.Run("Constructor", func(t *testing.T) {
		constructor := class.Methods["__construct"]
		if constructor == nil {
			t.Fatal("Constructor not found")
		}

		args := []*values.Value{thisObj, values.NewString(testFile)}
		impl := constructor.Implementation.(*BuiltinMethodImpl)
		_, err := impl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("Constructor failed: %v", err)
		}

		// Check that filepath was stored
		if filepath, ok := obj.Properties["__filepath"]; !ok || !filepath.IsString() {
			t.Fatal("Filepath not stored correctly")
		}
	})

	t.Run("getFilename", func(t *testing.T) {
		method := class.Methods["getFilename"]
		if method == nil {
			t.Fatal("getFilename method not found")
		}

		args := []*values.Value{thisObj}
		impl := method.Implementation.(*BuiltinMethodImpl)
		result, err := impl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("getFilename failed: %v", err)
		}

		if !result.IsString() || result.ToString() != "test_file.txt" {
			t.Fatalf("Expected 'test_file.txt', got: %v", result.ToString())
		}
	})

	t.Run("getBasename", func(t *testing.T) {
		method := class.Methods["getBasename"]
		if method == nil {
			t.Fatal("getBasename method not found")
		}

		args := []*values.Value{thisObj}
		impl := method.Implementation.(*BuiltinMethodImpl)
		result, err := impl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("getBasename failed: %v", err)
		}

		if !result.IsString() || result.ToString() != "test_file.txt" {
			t.Fatalf("Expected 'test_file.txt', got: %v", result.ToString())
		}
	})

	t.Run("getExtension", func(t *testing.T) {
		method := class.Methods["getExtension"]
		if method == nil {
			t.Fatal("getExtension method not found")
		}

		args := []*values.Value{thisObj}
		impl := method.Implementation.(*BuiltinMethodImpl)
		result, err := impl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("getExtension failed: %v", err)
		}

		if !result.IsString() || result.ToString() != "txt" {
			t.Fatalf("Expected 'txt', got: %v", result.ToString())
		}
	})

	t.Run("getPath", func(t *testing.T) {
		method := class.Methods["getPath"]
		if method == nil {
			t.Fatal("getPath method not found")
		}

		args := []*values.Value{thisObj}
		impl := method.Implementation.(*BuiltinMethodImpl)
		result, err := impl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("getPath failed: %v", err)
		}

		expected := filepath.Dir(testFile)
		if !result.IsString() || result.ToString() != expected {
			t.Fatalf("Expected '%s', got: %v", expected, result.ToString())
		}
	})

	t.Run("getPathname", func(t *testing.T) {
		method := class.Methods["getPathname"]
		if method == nil {
			t.Fatal("getPathname method not found")
		}

		args := []*values.Value{thisObj}
		impl := method.Implementation.(*BuiltinMethodImpl)
		result, err := impl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("getPathname failed: %v", err)
		}

		if !result.IsString() || result.ToString() != testFile {
			t.Fatalf("Expected '%s', got: %v", testFile, result.ToString())
		}
	})

	t.Run("isFile", func(t *testing.T) {
		method := class.Methods["isFile"]
		if method == nil {
			t.Fatal("isFile method not found")
		}

		args := []*values.Value{thisObj}
		impl := method.Implementation.(*BuiltinMethodImpl)
		result, err := impl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("isFile failed: %v", err)
		}

		if !result.IsBool() || !result.ToBool() {
			t.Fatal("Expected isFile() to return true for a file")
		}
	})

	t.Run("isDir", func(t *testing.T) {
		method := class.Methods["isDir"]
		if method == nil {
			t.Fatal("isDir method not found")
		}

		args := []*values.Value{thisObj}
		impl := method.Implementation.(*BuiltinMethodImpl)
		result, err := impl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("isDir failed: %v", err)
		}

		if !result.IsBool() || result.ToBool() {
			t.Fatal("Expected isDir() to return false for a file")
		}
	})

	t.Run("getSize", func(t *testing.T) {
		method := class.Methods["getSize"]
		if method == nil {
			t.Fatal("getSize method not found")
		}

		args := []*values.Value{thisObj}
		impl := method.Implementation.(*BuiltinMethodImpl)
		result, err := impl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("getSize failed: %v", err)
		}

		if !result.IsInt() || result.ToInt() != 25 {
			t.Fatalf("Expected size 25, got: %v", result.ToInt())
		}
	})

	t.Run("DirectoryTest", func(t *testing.T) {
		// Test with directory
		dirObj := &values.Object{
			ClassName:  "SplFileInfo",
			Properties: make(map[string]*values.Value),
		}
		dirThisObj := &values.Value{
			Type: values.TypeObject,
			Data: dirObj,
		}

		// Initialize with directory
		constructor := class.Methods["__construct"]
		args := []*values.Value{dirThisObj, values.NewString(testDir)}
		impl := constructor.Implementation.(*BuiltinMethodImpl)
		_, err := impl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("Constructor failed for directory: %v", err)
		}

		// Test isDir
		isDirMethod := class.Methods["isDir"]
		args = []*values.Value{dirThisObj}
		impl = isDirMethod.Implementation.(*BuiltinMethodImpl)
		result, err := impl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("isDir failed for directory: %v", err)
		}

		if !result.IsBool() || !result.ToBool() {
			t.Fatal("Expected isDir() to return true for a directory")
		}

		// Test isFile
		isFileMethod := class.Methods["isFile"]
		args = []*values.Value{dirThisObj}
		impl = isFileMethod.Implementation.(*BuiltinMethodImpl)
		result, err = impl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("isFile failed for directory: %v", err)
		}

		if !result.IsBool() || result.ToBool() {
			t.Fatal("Expected isFile() to return false for a directory")
		}
	})

	t.Run("NonExistentFile", func(t *testing.T) {
		// Test with non-existent file
		nonExistentFile := "/tmp/claude_test/nonexistent.txt"
		neObj := &values.Object{
			ClassName:  "SplFileInfo",
			Properties: make(map[string]*values.Value),
		}
		neThisObj := &values.Value{
			Type: values.TypeObject,
			Data: neObj,
		}

		// Initialize with non-existent file
		constructor := class.Methods["__construct"]
		args := []*values.Value{neThisObj, values.NewString(nonExistentFile)}
		impl := constructor.Implementation.(*BuiltinMethodImpl)
		_, err := impl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("Constructor failed for non-existent file: %v", err)
		}

		// Test getFilename still works
		getFilenameMethod := class.Methods["getFilename"]
		args = []*values.Value{neThisObj}
		impl = getFilenameMethod.Implementation.(*BuiltinMethodImpl)
		result, err := impl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("getFilename failed for non-existent file: %v", err)
		}

		if !result.IsString() || result.ToString() != "nonexistent.txt" {
			t.Fatalf("Expected 'nonexistent.txt', got: %v", result.ToString())
		}

		// Test isFile returns false
		isFileMethod := class.Methods["isFile"]
		args = []*values.Value{neThisObj}
		impl = isFileMethod.Implementation.(*BuiltinMethodImpl)
		result, err = impl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("isFile failed for non-existent file: %v", err)
		}

		if !result.IsBool() || result.ToBool() {
			t.Fatal("Expected isFile() to return false for non-existent file")
		}

		// Test getRealPath returns false
		getRealPathMethod := class.Methods["getRealPath"]
		args = []*values.Value{neThisObj}
		impl = getRealPathMethod.Implementation.(*BuiltinMethodImpl)
		result, err = impl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("getRealPath failed for non-existent file: %v", err)
		}

		if !result.IsBool() || result.ToBool() {
			t.Fatal("Expected getRealPath() to return false for non-existent file")
		}
	})
}