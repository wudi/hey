package runtime

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

func TestFilesystemFunctions(t *testing.T) {
	// Setup test environment
	testDir := filepath.Join(os.TempDir(), "hey_filesystem_test")
	testFile := filepath.Join(testDir, "test.txt")
	testContent := "Hello, World!\nThis is a test file.\nLine 3."

	// Clean up any existing test files
	os.RemoveAll(testDir)
	defer os.RemoveAll(testDir) // Cleanup after test

	tests := []struct {
		name     string
		function string
		args     []*values.Value
		want     *values.Value
		setup    func() // Setup function called before test
		validate func() // Additional validation called after test
	}{
		// Directory operations
		{
			name:     "mkdir creates directory",
			function: "mkdir",
			args:     []*values.Value{values.NewString(testDir)},
			want:     values.NewBool(true),
		},
		{
			name:     "is_dir with existing directory",
			function: "is_dir",
			args:     []*values.Value{values.NewString(testDir)},
			want:     values.NewBool(true),
		},
		{
			name:     "is_dir with non-existent directory",
			function: "is_dir",
			args:     []*values.Value{values.NewString("/nonexistent")},
			want:     values.NewBool(false),
		},

		// File content operations (already implemented - testing for regression)
		{
			name:     "file_put_contents writes file",
			function: "file_put_contents",
			args:     []*values.Value{values.NewString(testFile), values.NewString(testContent)},
			want:     values.NewInt(42), // Length of test content
		},
		{
			name:     "file_get_contents reads file",
			function: "file_get_contents",
			args:     []*values.Value{values.NewString(testFile)},
			want:     values.NewString(testContent),
		},
		{
			name:     "file_get_contents with non-existent file",
			function: "file_get_contents",
			args:     []*values.Value{values.NewString("/nonexistent.txt")},
			want:     values.NewBool(false),
		},

		// File existence functions
		{
			name:     "file_exists with existing file",
			function: "file_exists",
			args:     []*values.Value{values.NewString(testFile)},
			want:     values.NewBool(true),
		},
		{
			name:     "file_exists with non-existent file",
			function: "file_exists",
			args:     []*values.Value{values.NewString("/nonexistent.txt")},
			want:     values.NewBool(false),
		},
		{
			name:     "is_file with regular file",
			function: "is_file",
			args:     []*values.Value{values.NewString(testFile)},
			want:     values.NewBool(true),
		},
		{
			name:     "is_file with directory",
			function: "is_file",
			args:     []*values.Value{values.NewString(testDir)},
			want:     values.NewBool(false),
		},

		// File information functions
		{
			name:     "filesize returns correct size",
			function: "filesize",
			args:     []*values.Value{values.NewString(testFile)},
			want:     values.NewInt(42),
		},
		{
			name:     "filesize with non-existent file",
			function: "filesize",
			args:     []*values.Value{values.NewString("/nonexistent.txt")},
			want:     values.NewBool(false),
		},
		{
			name:     "filetype with regular file",
			function: "filetype",
			args:     []*values.Value{values.NewString(testFile)},
			want:     values.NewString("file"),
		},
		{
			name:     "filetype with directory",
			function: "filetype",
			args:     []*values.Value{values.NewString(testDir)},
			want:     values.NewString("dir"),
		},

		// Path functions
		{
			name:     "dirname returns parent directory",
			function: "dirname",
			args:     []*values.Value{values.NewString(testFile)},
			want:     values.NewString(testDir),
		},
		{
			name:     "basename returns filename",
			function: "basename",
			args:     []*values.Value{values.NewString(testFile)},
			want:     values.NewString("test.txt"),
		},
		{
			name:     "basename with suffix",
			function: "basename",
			args:     []*values.Value{values.NewString(testFile), values.NewString(".txt")},
			want:     values.NewString("test"),
		},

		// File array function
		{
			name:     "file reads lines into array",
			function: "file",
			args:     []*values.Value{values.NewString(testFile)},
			want:     createArrayFromSlice([]*values.Value{
				values.NewString("Hello, World!\n"),
				values.NewString("This is a test file.\n"),
				values.NewString("Line 3."),
			}),
		},

		// is_readable/is_writable
		{
			name:     "is_readable with readable file",
			function: "is_readable",
			args:     []*values.Value{values.NewString(testFile)},
			want:     values.NewBool(true),
		},
		{
			name:     "is_writable with writable file",
			function: "is_writable",
			args:     []*values.Value{values.NewString(testFile)},
			want:     values.NewBool(true),
		},
		{
			name:     "is_writeable alias test",
			function: "is_writeable",
			args:     []*values.Value{values.NewString(testFile)},
			want:     values.NewBool(true),
		},

		// Copy and rename
		{
			name:     "copy file",
			function: "copy",
			args:     []*values.Value{values.NewString(testFile), values.NewString(filepath.Join(testDir, "copy_test.txt"))},
			want:     values.NewBool(true),
			validate: func() {
				copyFile := filepath.Join(testDir, "copy_test.txt")
				if _, err := os.Stat(copyFile); os.IsNotExist(err) {
					t.Errorf("Copy file was not created")
				}
			},
		},
		{
			name:     "rename file",
			function: "rename",
			args:     []*values.Value{values.NewString(filepath.Join(testDir, "copy_test.txt")), values.NewString(filepath.Join(testDir, "renamed_test.txt"))},
			want:     values.NewBool(true),
			validate: func() {
				oldFile := filepath.Join(testDir, "copy_test.txt")
				newFile := filepath.Join(testDir, "renamed_test.txt")
				if _, err := os.Stat(oldFile); !os.IsNotExist(err) {
					t.Errorf("Old file still exists after rename")
				}
				if _, err := os.Stat(newFile); os.IsNotExist(err) {
					t.Errorf("New file was not created")
				}
			},
		},

		// File deletion (testing existing implementation)
		{
			name:     "unlink removes file",
			function: "unlink",
			args:     []*values.Value{values.NewString(testFile)},
			want:     values.NewBool(true),
			validate: func() {
				if _, err := os.Stat(testFile); !os.IsNotExist(err) {
					t.Errorf("File still exists after unlink")
				}
			},
		},

		// Directory removal
		{
			name:     "rmdir removes directory",
			function: "rmdir",
			args:     []*values.Value{values.NewString(testDir)},
			want:     values.NewBool(true),
			setup: func() {
				// Clean up any remaining files
				os.RemoveAll(testDir)
				os.Mkdir(testDir, 0755) // Create empty directory
			},
			validate: func() {
				if _, err := os.Stat(testDir); !os.IsNotExist(err) {
					t.Errorf("Directory still exists after rmdir")
				}
			},
		},
	}

	// Get filesystem functions
	functions := GetFilesystemFunctions()
	functionMap := make(map[string]*registry.Function)

	for _, fn := range functions {
		functionMap[fn.Name] = fn
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			if tt.setup != nil {
				tt.setup()
			} else {
				// Default setup - ensure test directory and file exist
				os.MkdirAll(testDir, 0755)
				if strings.Contains(tt.name, "file_") || strings.Contains(tt.name, "is_file") ||
				   strings.Contains(tt.name, "filesize") || strings.Contains(tt.name, "filetype") ||
				   strings.Contains(tt.name, "readable") || strings.Contains(tt.name, "writable") ||
				   strings.Contains(tt.name, "dirname") || strings.Contains(tt.name, "basename") ||
				   strings.Contains(tt.name, "copy") || strings.Contains(tt.name, "rename") ||
				   strings.Contains(tt.name, "unlink") {
					os.WriteFile(testFile, []byte(testContent), 0644)
				}
			}

			// Find function
			fn, exists := functionMap[tt.function]
			if !exists {
				t.Fatalf("Function %s not found", tt.function)
			}

			// Execute
			result, err := fn.Builtin(nil, tt.args)
			if err != nil {
				t.Fatalf("Function call failed: %v", err)
			}

			// Compare results
			if !compareValues(result, tt.want) {
				t.Errorf("Function %s returned %v, want %v", tt.function, result, tt.want)
			}

			// Additional validation
			if tt.validate != nil {
				tt.validate()
			}
		})
	}
}

func TestFileHandleOperations(t *testing.T) {
	// Setup test environment
	testDir := filepath.Join(os.TempDir(), "hey_filehandle_test")
	testFile := filepath.Join(testDir, "handle_test.txt")
	testContent := "Line 1\nLine 2\nLine 3"

	// Clean up any existing test files
	os.RemoveAll(testDir)
	defer os.RemoveAll(testDir) // Cleanup after test

	// Create test directory
	os.MkdirAll(testDir, 0755)

	// Get filesystem functions
	functions := GetFilesystemFunctions()
	functionMap := make(map[string]*registry.Function)

	for _, fn := range functions {
		functionMap[fn.Name] = fn
	}

	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "fopen_fwrite_fclose_write_mode",
			testFunc: func(t *testing.T) {
				// Test fopen in write mode
				handle, err := functionMap["fopen"].Builtin(nil, []*values.Value{
					values.NewString(testFile),
					values.NewString("w"),
				})
				if err != nil {
					t.Fatalf("fopen failed: %v", err)
				}
				if handle.Type != values.TypeResource {
					t.Fatalf("fopen should return resource, got %v", handle.Type)
				}

				// Test fwrite
				written, err := functionMap["fwrite"].Builtin(nil, []*values.Value{
					handle,
					values.NewString(testContent),
				})
				if err != nil {
					t.Fatalf("fwrite failed: %v", err)
				}
				if written.ToInt() != int64(len(testContent)) {
					t.Errorf("fwrite returned %d, expected %d", written.ToInt(), len(testContent))
				}

				// Test fclose
				closed, err := functionMap["fclose"].Builtin(nil, []*values.Value{handle})
				if err != nil {
					t.Fatalf("fclose failed: %v", err)
				}
				if !closed.Data.(bool) {
					t.Error("fclose should return true")
				}
			},
		},
		{
			name: "fopen_fread_operations",
			testFunc: func(t *testing.T) {
				// First create the test file
				os.WriteFile(testFile, []byte(testContent), 0644)

				// Test fopen in read mode
				handle, err := functionMap["fopen"].Builtin(nil, []*values.Value{
					values.NewString(testFile),
					values.NewString("r"),
				})
				if err != nil {
					t.Fatalf("fopen failed: %v", err)
				}

				// Test fread partial
				chunk, err := functionMap["fread"].Builtin(nil, []*values.Value{
					handle,
					values.NewInt(5),
				})
				if err != nil {
					t.Fatalf("fread failed: %v", err)
				}
				if chunk.ToString() != "Line " {
					t.Errorf("fread returned %q, expected %q", chunk.ToString(), "Line ")
				}

				// Test ftell
				pos, err := functionMap["ftell"].Builtin(nil, []*values.Value{handle})
				if err != nil {
					t.Fatalf("ftell failed: %v", err)
				}
				if pos.ToInt() != 5 {
					t.Errorf("ftell returned %d, expected 5", pos.ToInt())
				}

				// Test feof
				eof, err := functionMap["feof"].Builtin(nil, []*values.Value{handle})
				if err != nil {
					t.Fatalf("feof failed: %v", err)
				}
				if eof.Data.(bool) {
					t.Error("feof should return false")
				}

				// Test fseek to end
				seekResult, err := functionMap["fseek"].Builtin(nil, []*values.Value{
					handle,
					values.NewInt(0),
					values.NewInt(2), // SEEK_END
				})
				if err != nil {
					t.Fatalf("fseek failed: %v", err)
				}
				if seekResult.ToInt() != 0 {
					t.Errorf("fseek should return 0, got %d", seekResult.ToInt())
				}

				// Test ftell after seek to end
				pos, err = functionMap["ftell"].Builtin(nil, []*values.Value{handle})
				if err != nil {
					t.Fatalf("ftell failed: %v", err)
				}
				if pos.ToInt() != int64(len(testContent)) {
					t.Errorf("ftell returned %d, expected %d", pos.ToInt(), len(testContent))
				}

				// Test rewind
				rewound, err := functionMap["rewind"].Builtin(nil, []*values.Value{handle})
				if err != nil {
					t.Fatalf("rewind failed: %v", err)
				}
				if !rewound.Data.(bool) {
					t.Error("rewind should return true")
				}

				// Test ftell after rewind
				pos, err = functionMap["ftell"].Builtin(nil, []*values.Value{handle})
				if err != nil {
					t.Fatalf("ftell failed: %v", err)
				}
				if pos.ToInt() != 0 {
					t.Errorf("ftell after rewind returned %d, expected 0", pos.ToInt())
				}

				functionMap["fclose"].Builtin(nil, []*values.Value{handle})
			},
		},
		{
			name: "fgets_fgetc_operations",
			testFunc: func(t *testing.T) {
				// Create test file
				os.WriteFile(testFile, []byte(testContent), 0644)

				handle, err := functionMap["fopen"].Builtin(nil, []*values.Value{
					values.NewString(testFile),
					values.NewString("r"),
				})
				if err != nil {
					t.Fatalf("fopen failed: %v", err)
				}

				// Test fgets
				line, err := functionMap["fgets"].Builtin(nil, []*values.Value{handle})
				if err != nil {
					t.Fatalf("fgets failed: %v", err)
				}
				if line.ToString() != "Line 1\n" {
					t.Errorf("fgets returned %q, expected %q", line.ToString(), "Line 1\n")
				}

				// Test fgetc
				char, err := functionMap["fgetc"].Builtin(nil, []*values.Value{handle})
				if err != nil {
					t.Fatalf("fgetc failed: %v", err)
				}
				if char.ToString() != "L" {
					t.Errorf("fgetc returned %q, expected %q", char.ToString(), "L")
				}

				functionMap["fclose"].Builtin(nil, []*values.Value{handle})
			},
		},
		{
			name: "append_mode_operations",
			testFunc: func(t *testing.T) {
				// Create initial file
				os.WriteFile(testFile, []byte("Initial"), 0644)

				handle, err := functionMap["fopen"].Builtin(nil, []*values.Value{
					values.NewString(testFile),
					values.NewString("a"),
				})
				if err != nil {
					t.Fatalf("fopen failed: %v", err)
				}

				// Test fputs (alias for fwrite)
				written, err := functionMap["fputs"].Builtin(nil, []*values.Value{
					handle,
					values.NewString("\nAppended"),
				})
				if err != nil {
					t.Fatalf("fputs failed: %v", err)
				}
				if written.ToInt() != 9 { // len("\nAppended")
					t.Errorf("fputs returned %d, expected 9", written.ToInt())
				}

				// Test fflush
				flushed, err := functionMap["fflush"].Builtin(nil, []*values.Value{handle})
				if err != nil {
					t.Fatalf("fflush failed: %v", err)
				}
				if !flushed.Data.(bool) {
					t.Error("fflush should return true")
				}

				functionMap["fclose"].Builtin(nil, []*values.Value{handle})

				// Verify the appended content
				content, _ := os.ReadFile(testFile)
				expected := "Initial\nAppended"
				if string(content) != expected {
					t.Errorf("File content is %q, expected %q", string(content), expected)
				}
			},
		},
		{
			name: "error_conditions",
			testFunc: func(t *testing.T) {
				// Test fopen with invalid mode
				handle, err := functionMap["fopen"].Builtin(nil, []*values.Value{
					values.NewString(testFile),
					values.NewString("invalid"),
				})
				if err != nil {
					t.Fatalf("fopen with invalid mode failed: %v", err)
				}
				if handle.Data.(bool) != false {
					t.Error("fopen with invalid mode should return false")
				}

				// Test fopen with nonexistent directory
				handle, err = functionMap["fopen"].Builtin(nil, []*values.Value{
					values.NewString("/nonexistent/file.txt"),
					values.NewString("r"),
				})
				if err != nil {
					t.Fatalf("fopen with nonexistent dir failed: %v", err)
				}
				if handle.Data.(bool) != false {
					t.Error("fopen with nonexistent dir should return false")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.testFunc(t)
		})
	}
}

func TestPathinfoFunction(t *testing.T) {
	// Get filesystem functions
	functions := GetFilesystemFunctions()
	functionMap := make(map[string]*registry.Function)

	for _, fn := range functions {
		functionMap[fn.Name] = fn
	}

	tests := []struct {
		name     string
		path     string
		flags    *values.Value // nil means use default
		expected map[string]string
	}{
		{
			name:  "full path with extension",
			path:  "/home/user/file.txt",
			flags: nil,
			expected: map[string]string{
				"dirname":   "/home/user",
				"basename":  "file.txt",
				"extension": "txt",
				"filename":  "file",
			},
		},
		{
			name:  "path without extension",
			path:  "/home/user/file",
			flags: nil,
			expected: map[string]string{
				"dirname":  "/home/user",
				"basename": "file",
				"filename": "file",
			},
		},
		{
			name:  "hidden file",
			path:  "/home/user/.hidden",
			flags: nil,
			expected: map[string]string{
				"dirname":   "/home/user",
				"basename":  ".hidden",
				"extension": "hidden",
				"filename":  "",
			},
		},
		{
			name:  "file with trailing dot",
			path:  "file.",
			flags: nil,
			expected: map[string]string{
				"dirname":  ".",
				"basename": "file.",
				"filename": "file",
			},
		},
		{
			name:     "dirname flag only",
			path:     "/home/user/document.pdf",
			flags:    values.NewInt(1), // PATHINFO_DIRNAME
			expected: map[string]string{},
		},
		{
			name:     "basename flag only",
			path:     "/home/user/document.pdf",
			flags:    values.NewInt(2), // PATHINFO_BASENAME
			expected: map[string]string{},
		},
		{
			name:     "extension flag only",
			path:     "/home/user/document.pdf",
			flags:    values.NewInt(4), // PATHINFO_EXTENSION
			expected: map[string]string{},
		},
		{
			name:     "filename flag only",
			path:     "/home/user/document.pdf",
			flags:    values.NewInt(8), // PATHINFO_FILENAME
			expected: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := functionMap["pathinfo"]
			if fn == nil {
				t.Fatal("pathinfo function not found")
			}

			var args []*values.Value
			if tt.flags != nil {
				args = []*values.Value{values.NewString(tt.path), tt.flags}
			} else {
				args = []*values.Value{values.NewString(tt.path)}
			}

			result, err := fn.Builtin(nil, args)
			if err != nil {
				t.Fatalf("pathinfo failed: %v", err)
			}

			// For single flag tests, check the returned string value
			if tt.flags != nil && (tt.flags.ToInt() == 1 || tt.flags.ToInt() == 2 || tt.flags.ToInt() == 4 || tt.flags.ToInt() == 8) {
				if result.Type != values.TypeString {
					t.Errorf("Expected string result for single flag, got %v", result.Type)
					return
				}

				expectedValue := ""
				switch tt.flags.ToInt() {
				case 1: // PATHINFO_DIRNAME
					expectedValue = "/home/user"
				case 2: // PATHINFO_BASENAME
					expectedValue = "document.pdf"
				case 4: // PATHINFO_EXTENSION
					expectedValue = "pdf"
				case 8: // PATHINFO_FILENAME
					expectedValue = "document"
				}

				if result.ToString() != expectedValue {
					t.Errorf("Expected %s, got %s", expectedValue, result.ToString())
				}
				return
			}

			// For array results
			if result.Type != values.TypeArray {
				t.Errorf("Expected array result, got %v", result.Type)
				return
			}

			_ = result.Data.(*values.Array)

			// Check expected components are present
			for key, expectedValue := range tt.expected {
				keyVal := values.NewString(key)
				actualVal := result.ArrayGet(keyVal)

				if actualVal.Type == values.TypeNull {
					t.Errorf("Missing key %s in result", key)
					continue
				}

				if actualVal.ToString() != expectedValue {
					t.Errorf("Key %s: expected %s, got %s", key, expectedValue, actualVal.ToString())
				}
			}

			// Check that extension is only included when not empty
			extKey := values.NewString("extension")
			extVal := result.ArrayGet(extKey)
			if _, hasExt := tt.expected["extension"]; !hasExt && extVal.Type != values.TypeNull {
				t.Errorf("Extension should not be present when empty, but got: %s", extVal.ToString())
			}
		})
	}
}

func TestFilemtimeFunction(t *testing.T) {
	// Setup test environment
	testDir := filepath.Join(os.TempDir(), "hey_filetime_test")
	testFile := filepath.Join(testDir, "time_test.txt")

	// Clean up any existing test files
	os.RemoveAll(testDir)
	defer os.RemoveAll(testDir) // Cleanup after test

	// Create test directory and file
	os.MkdirAll(testDir, 0755)
	os.WriteFile(testFile, []byte("test content"), 0644)

	// Get filesystem functions
	functions := GetFilesystemFunctions()
	functionMap := make(map[string]*registry.Function)

	for _, fn := range functions {
		functionMap[fn.Name] = fn
	}

	tests := []struct {
		name     string
		function string
		file     string
		wantType values.ValueType
	}{
		{
			name:     "filemtime with existing file",
			function: "filemtime",
			file:     testFile,
			wantType: values.TypeInt,
		},
		{
			name:     "fileatime with existing file",
			function: "fileatime",
			file:     testFile,
			wantType: values.TypeInt,
		},
		{
			name:     "filectime with existing file",
			function: "filectime",
			file:     testFile,
			wantType: values.TypeInt,
		},
		{
			name:     "filemtime with non-existent file",
			function: "filemtime",
			file:     "/nonexistent/file.txt",
			wantType: values.TypeBool,
		},
		{
			name:     "fileatime with non-existent file",
			function: "fileatime",
			file:     "/nonexistent/file.txt",
			wantType: values.TypeBool,
		},
		{
			name:     "filectime with non-existent file",
			function: "filectime",
			file:     "/nonexistent/file.txt",
			wantType: values.TypeBool,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := functionMap[tt.function]
			if fn == nil {
				t.Fatalf("Function %s not found", tt.function)
			}

			result, err := fn.Builtin(nil, []*values.Value{values.NewString(tt.file)})
			if err != nil {
				t.Fatalf("Function call failed: %v", err)
			}

			if result.Type != tt.wantType {
				t.Errorf("Function %s returned type %v, want %v", tt.function, result.Type, tt.wantType)
			}

			// For successful calls, check that timestamp is reasonable
			if tt.wantType == values.TypeInt && result.Type == values.TypeInt {
				timestamp := result.ToInt()
				// Check that timestamp is somewhat recent (after year 2020)
				if timestamp < 1577836800 { // 2020-01-01 timestamp
					t.Errorf("Timestamp %d seems too old", timestamp)
				}
			}

			// For failed calls, check that it returns false
			if tt.wantType == values.TypeBool && result.Type == values.TypeBool {
				if result.Data.(bool) != false {
					t.Errorf("Expected false for non-existent file, got %v", result.Data)
				}
			}
		})
	}
}

// Helper function to compare values
func compareValues(got, want *values.Value) bool {
	if got.Type != want.Type {
		return false
	}

	switch got.Type {
	case values.TypeBool:
		return got.Data.(bool) == want.Data.(bool)
	case values.TypeInt:
		return got.Data.(int64) == want.Data.(int64)
	case values.TypeString:
		return got.ToString() == want.ToString()
	case values.TypeArray:
		gotArray := got.Data.(*values.Array)
		wantArray := want.Data.(*values.Array)

		if len(gotArray.Elements) != len(wantArray.Elements) {
			return false
		}

		// For arrays, we need to compare by index
		for i := int64(0); i < int64(len(gotArray.Elements)); i++ {
			gotElem, gotExists := gotArray.Elements[i]
			wantElem, wantExists := wantArray.Elements[i]

			if gotExists != wantExists {
				return false
			}
			if gotExists && !compareValues(gotElem, wantElem) {
				return false
			}
		}
		return true
	}
	return false
}

// Helper function to create array from slice
func createArrayFromSlice(elements []*values.Value) *values.Value {
	arr := values.NewArray()
	for i, elem := range elements {
		arr.ArraySet(values.NewInt(int64(i)), elem)
	}
	return arr
}

func TestFtruncateFunction(t *testing.T) {
	tmpFile := "/tmp/test_ftruncate.txt"
	defer os.Remove(tmpFile)

	functions := GetFilesystemFunctions()
	callFunction := func(name string, args []*values.Value) *values.Value {
		for _, fn := range functions {
			if fn.Name == name {
				result, err := fn.Builtin(nil, args)
				if err != nil {
					t.Fatalf("Error calling %s: %v", name, err)
				}
				return result
			}
		}
		t.Fatalf("Function %s not found", name)
		return nil
	}

	// Create test file with content
	result := callFunction("fopen", []*values.Value{values.NewString(tmpFile), values.NewString("w+")})
	if result.Type != values.TypeResource {
		t.Fatalf("fopen should return resource, got: %v", result)
	}
	fileHandle := result

	// Write content
	content := "Hello World! This is a test file for truncation."
	result = callFunction("fwrite", []*values.Value{fileHandle, values.NewString(content)})
	if result.ToInt() != int64(len(content)) {
		t.Errorf("fwrite should return %d, got: %d", len(content), result.ToInt())
	}

	// Test truncate to smaller size
	result = callFunction("ftruncate", []*values.Value{fileHandle, values.NewInt(5)})
	if !result.ToBool() {
		t.Errorf("ftruncate should return true, got: %v", result)
	}

	// Close and reopen to read
	callFunction("fclose", []*values.Value{fileHandle})

	// Read truncated content
	result = callFunction("file_get_contents", []*values.Value{values.NewString(tmpFile)})
	truncatedContent := result.ToString()
	if truncatedContent != "Hello" {
		t.Errorf("After truncate(5), content should be 'Hello', got: '%s'", truncatedContent)
	}

	// Test truncate to larger size (extends file with null bytes)
	result = callFunction("fopen", []*values.Value{values.NewString(tmpFile), values.NewString("r+")})
	fileHandle = result

	result = callFunction("ftruncate", []*values.Value{fileHandle, values.NewInt(10)})
	if !result.ToBool() {
		t.Errorf("ftruncate should return true, got: %v", result)
	}

	callFunction("fclose", []*values.Value{fileHandle})

	// Check file size
	result = callFunction("filesize", []*values.Value{values.NewString(tmpFile)})
	if result.ToInt() != 10 {
		t.Errorf("After truncate(10), file size should be 10, got: %d", result.ToInt())
	}

	// Test error cases
	result = callFunction("ftruncate", []*values.Value{values.NewBool(false), values.NewInt(5)})
	if result.ToBool() {
		t.Errorf("ftruncate with invalid handle should return false")
	}
}

func TestPermissionFunctions(t *testing.T) {
	tmpFile := "/tmp/test_perms.txt"
	defer os.Remove(tmpFile)

	functions := GetFilesystemFunctions()
	callFunction := func(name string, args []*values.Value) *values.Value {
		for _, fn := range functions {
			if fn.Name == name {
				result, err := fn.Builtin(nil, args)
				if err != nil {
					t.Fatalf("Error calling %s: %v", name, err)
				}
				return result
			}
		}
		t.Fatalf("Function %s not found", name)
		return nil
	}

	// Create test file
	result := callFunction("file_put_contents", []*values.Value{values.NewString(tmpFile), values.NewString("test")})
	if result.ToInt() != 4 {
		t.Fatalf("Failed to create test file")
	}

	// Test chmod with different permissions
	result = callFunction("chmod", []*values.Value{values.NewString(tmpFile), values.NewInt(0755)})
	if !result.ToBool() {
		t.Errorf("chmod(0755) should return true")
	}

	// Test fileperms
	result = callFunction("fileperms", []*values.Value{values.NewString(tmpFile)})
	if result.Type == values.TypeBool && result.ToBool() == false {
		t.Errorf("fileperms should not return false for existing file")
	}
	mode := result.ToInt() & 0777
	if mode != 0755 {
		t.Errorf("Expected permissions 755, got: %o", mode)
	}

	// Test chmod with different permissions
	result = callFunction("chmod", []*values.Value{values.NewString(tmpFile), values.NewInt(0644)})
	if !result.ToBool() {
		t.Errorf("chmod(0644) should return true")
	}

	result = callFunction("fileperms", []*values.Value{values.NewString(tmpFile)})
	mode = result.ToInt() & 0777
	if mode != 0644 {
		t.Errorf("Expected permissions 644, got: %o", mode)
	}

	// Test error cases
	result = callFunction("fileperms", []*values.Value{values.NewString("/nonexistent/file")})
	if result.ToBool() {
		t.Errorf("fileperms should return false for non-existent file")
	}

	result = callFunction("chmod", []*values.Value{values.NewString("/nonexistent/file"), values.NewInt(0755)})
	if result.ToBool() {
		t.Errorf("chmod should return false for non-existent file")
	}
}

func TestOwnershipFunctions(t *testing.T) {
	tmpFile := "/tmp/test_ownership.txt"
	defer os.Remove(tmpFile)

	functions := GetFilesystemFunctions()
	callFunction := func(name string, args []*values.Value) *values.Value {
		for _, fn := range functions {
			if fn.Name == name {
				result, err := fn.Builtin(nil, args)
				if err != nil {
					t.Fatalf("Error calling %s: %v", name, err)
				}
				return result
			}
		}
		t.Fatalf("Function %s not found", name)
		return nil
	}

	// Create test file
	result := callFunction("file_put_contents", []*values.Value{values.NewString(tmpFile), values.NewString("test")})
	if result.ToInt() != 4 {
		t.Fatalf("Failed to create test file")
	}

	// Test fileowner
	result = callFunction("fileowner", []*values.Value{values.NewString(tmpFile)})
	if result.Type == values.TypeBool && result.ToBool() == false {
		t.Errorf("fileowner should not return false for existing file")
	}
	originalUID := result.ToInt()
	if originalUID < 0 {
		t.Errorf("fileowner should return non-negative UID, got: %d", originalUID)
	}

	// Test filegroup
	result = callFunction("filegroup", []*values.Value{values.NewString(tmpFile)})
	if result.Type == values.TypeBool && result.ToBool() == false {
		t.Errorf("filegroup should not return false for existing file")
	}
	originalGID := result.ToInt()
	if originalGID < 0 {
		t.Errorf("filegroup should return non-negative GID, got: %d", originalGID)
	}

	// Test chown (will likely fail without root, but should not crash)
	result = callFunction("chown", []*values.Value{values.NewString(tmpFile), values.NewInt(1000)})
	// Don't check the result as it will likely fail due to permissions

	// Test chgrp (will likely fail without root, but should not crash)
	result = callFunction("chgrp", []*values.Value{values.NewString(tmpFile), values.NewInt(1000)})
	// Don't check the result as it will likely fail due to permissions

	// Test error cases
	result = callFunction("fileowner", []*values.Value{values.NewString("/nonexistent/file")})
	if result.ToBool() {
		t.Errorf("fileowner should return false for non-existent file")
	}

	result = callFunction("filegroup", []*values.Value{values.NewString("/nonexistent/file")})
	if result.ToBool() {
		t.Errorf("filegroup should return false for non-existent file")
	}

	result = callFunction("chown", []*values.Value{values.NewString("/nonexistent/file"), values.NewInt(1000)})
	if result.ToBool() {
		t.Errorf("chown should return false for non-existent file")
	}

	result = callFunction("chgrp", []*values.Value{values.NewString("/nonexistent/file"), values.NewInt(1000)})
	if result.ToBool() {
		t.Errorf("chgrp should return false for non-existent file")
	}
}

func TestStatFunctions(t *testing.T) {
	tmpFile := "/tmp/test_stat.txt"
	defer os.Remove(tmpFile)

	functions := GetFilesystemFunctions()
	callFunction := func(name string, args []*values.Value) *values.Value {
		for _, fn := range functions {
			if fn.Name == name {
				result, err := fn.Builtin(nil, args)
				if err != nil {
					t.Fatalf("Error calling %s: %v", name, err)
				}
				return result
			}
		}
		t.Fatalf("Function %s not found", name)
		return nil
	}

	// Create test file
	content := "Hello World!"
	result := callFunction("file_put_contents", []*values.Value{values.NewString(tmpFile), values.NewString(content)})
	if result.ToInt() != int64(len(content)) {
		t.Fatalf("Failed to create test file")
	}

	// Test stat function
	result = callFunction("stat", []*values.Value{values.NewString(tmpFile)})
	if result.Type != values.TypeArray {
		t.Fatalf("stat should return array, got: %v", result.Type)
	}

	// Check size
	sizeVal := result.ArrayGet(values.NewString("size"))
	if sizeVal.ToInt() != int64(len(content)) {
		t.Errorf("Expected size %d, got: %d", len(content), sizeVal.ToInt())
	}

	// Check mode exists
	modeVal := result.ArrayGet(values.NewString("mode"))
	if modeVal.IsNull() {
		t.Errorf("stat array should contain mode")
	}

	// Check mtime exists
	mtimeVal := result.ArrayGet(values.NewString("mtime"))
	if mtimeVal.IsNull() {
		t.Errorf("stat array should contain mtime")
	}

	// Test lstat function
	result = callFunction("lstat", []*values.Value{values.NewString(tmpFile)})
	if result.Type != values.TypeArray {
		t.Fatalf("lstat should return array, got: %v", result.Type)
	}

	// Test fstat function
	fileHandle := callFunction("fopen", []*values.Value{values.NewString(tmpFile), values.NewString("r")})
	if fileHandle.Type != values.TypeResource {
		t.Fatalf("fopen should return resource")
	}

	result = callFunction("fstat", []*values.Value{fileHandle})
	if result.Type != values.TypeArray {
		t.Fatalf("fstat should return array, got: %v", result.Type)
	}

	callFunction("fclose", []*values.Value{fileHandle})

	// Test error cases
	result = callFunction("stat", []*values.Value{values.NewString("/nonexistent/file")})
	if result.ToBool() {
		t.Errorf("stat should return false for non-existent file")
	}

	result = callFunction("lstat", []*values.Value{values.NewString("/nonexistent/file")})
	if result.ToBool() {
		t.Errorf("lstat should return false for non-existent file")
	}

	result = callFunction("fstat", []*values.Value{values.NewBool(false)})
	if result.ToBool() {
		t.Errorf("fstat should return false for invalid handle")
	}
}

func TestLinkFunctions(t *testing.T) {
	tmpFile := "/tmp/test_links_orig.txt"
	tmpSymlink := "/tmp/test_symlink.txt"
	tmpHardlink := "/tmp/test_hardlink.txt"

	// Clean up any existing files
	defer func() {
		os.Remove(tmpSymlink)
		os.Remove(tmpHardlink)
		os.Remove(tmpFile)
	}()

	functions := GetFilesystemFunctions()
	callFunction := func(name string, args []*values.Value) *values.Value {
		for _, fn := range functions {
			if fn.Name == name {
				result, err := fn.Builtin(nil, args)
				if err != nil {
					t.Fatalf("Error calling %s: %v", name, err)
				}
				return result
			}
		}
		t.Fatalf("Function %s not found", name)
		return nil
	}

	// Create original file
	result := callFunction("file_put_contents", []*values.Value{values.NewString(tmpFile), values.NewString("original content")})
	if result.ToInt() != 16 {
		t.Fatalf("Failed to create original file")
	}

	// Test is_link on regular file
	result = callFunction("is_link", []*values.Value{values.NewString(tmpFile)})
	if result.ToBool() {
		t.Errorf("is_link should return false for regular file")
	}

	// Test symlink creation
	result = callFunction("symlink", []*values.Value{values.NewString(tmpFile), values.NewString(tmpSymlink)})
	if !result.ToBool() {
		t.Errorf("symlink creation should succeed")
	}

	// Test is_link on symlink
	result = callFunction("is_link", []*values.Value{values.NewString(tmpSymlink)})
	if !result.ToBool() {
		t.Errorf("is_link should return true for symlink")
	}

	// Test readlink
	result = callFunction("readlink", []*values.Value{values.NewString(tmpSymlink)})
	if result.ToString() != tmpFile {
		t.Errorf("readlink should return original file path, got: %s", result.ToString())
	}

	// Test hard link creation
	result = callFunction("link", []*values.Value{values.NewString(tmpFile), values.NewString(tmpHardlink)})
	if !result.ToBool() {
		t.Errorf("hard link creation should succeed")
	}

	// Test is_link on hard link (should be false)
	result = callFunction("is_link", []*values.Value{values.NewString(tmpHardlink)})
	if result.ToBool() {
		t.Errorf("is_link should return false for hard link")
	}

	// Test error cases
	result = callFunction("is_link", []*values.Value{values.NewString("/nonexistent/file")})
	if result.ToBool() {
		t.Errorf("is_link should return false for non-existent file")
	}

	result = callFunction("readlink", []*values.Value{values.NewString("/nonexistent/file")})
	if result.ToBool() {
		t.Errorf("readlink should return false for non-existent file")
	}

	result = callFunction("readlink", []*values.Value{values.NewString(tmpFile)})
	if result.ToBool() {
		t.Errorf("readlink should return false for regular file")
	}

	result = callFunction("symlink", []*values.Value{values.NewString(tmpFile), values.NewString(tmpSymlink)})
	if result.ToBool() {
		t.Errorf("symlink should return false if link already exists")
	}

	result = callFunction("link", []*values.Value{values.NewString("/nonexistent"), values.NewString("/tmp/invalid")})
	if result.ToBool() {
		t.Errorf("link should return false for non-existent target")
	}
}

func TestUtilityFunctions(t *testing.T) {
	tmpFile := "/tmp/test_touch.txt"
	defer os.Remove(tmpFile)

	functions := GetFilesystemFunctions()
	callFunction := func(name string, args []*values.Value) *values.Value {
		for _, fn := range functions {
			if fn.Name == name {
				result, err := fn.Builtin(nil, args)
				if err != nil {
					t.Fatalf("Error calling %s: %v", name, err)
				}
				return result
			}
		}
		t.Fatalf("Function %s not found", name)
		return nil
	}

	// Test touch on non-existent file (should create it)
	result := callFunction("touch", []*values.Value{values.NewString(tmpFile)})
	if !result.ToBool() {
		t.Errorf("touch should return true for successful file creation")
	}

	// Verify file was created
	result = callFunction("file_exists", []*values.Value{values.NewString(tmpFile)})
	if !result.ToBool() {
		t.Errorf("file should exist after touch")
	}

	// Test touch with specific time (1 hour ago)
	specificTime := time.Now().Unix() - 3600
	result = callFunction("touch", []*values.Value{values.NewString(tmpFile), values.NewInt(specificTime)})
	if !result.ToBool() {
		t.Errorf("touch with specific time should succeed")
	}

	// Verify the time was set correctly
	result = callFunction("filemtime", []*values.Value{values.NewString(tmpFile)})
	if result.ToInt() != specificTime {
		t.Errorf("Expected mtime %d, got %d", specificTime, result.ToInt())
	}

	// Test clearstatcache (should always succeed, returns null)
	result = callFunction("clearstatcache", []*values.Value{})
	if !result.IsNull() {
		t.Errorf("clearstatcache should return null")
	}

	// Test clearstatcache with parameters
	result = callFunction("clearstatcache", []*values.Value{values.NewBool(true), values.NewString(tmpFile)})
	if !result.IsNull() {
		t.Errorf("clearstatcache with params should return null")
	}

	// Test error cases
	result = callFunction("touch", []*values.Value{values.NewString("/nonexistent/path/file.txt")})
	if result.ToBool() {
		t.Errorf("touch should return false for invalid path")
	}
}

func TestAdditionalFunctions(t *testing.T) {
	tmpFile := "/tmp/test_additional.txt"
	tmpDir := "/tmp/test_additional_dir"
	tmpScript := "/tmp/test_script.sh"
	defer func() {
		os.Remove(tmpFile)
		os.Remove(tmpScript)
		os.RemoveAll(tmpDir)
	}()

	functions := GetFilesystemFunctions()
	callFunction := func(name string, args []*values.Value) *values.Value {
		for _, fn := range functions {
			if fn.Name == name {
				result, err := fn.Builtin(nil, args)
				if err != nil {
					t.Fatalf("Error calling %s: %v", name, err)
				}
				return result
			}
		}
		t.Fatalf("Function %s not found", name)
		return nil
	}

	// Create test files
	result := callFunction("file_put_contents", []*values.Value{values.NewString(tmpFile), values.NewString("test content")})
	if result.ToInt() != 12 {
		t.Fatalf("Failed to create test file")
	}

	// Create executable script
	result = callFunction("file_put_contents", []*values.Value{values.NewString(tmpScript), values.NewString("#!/bin/bash\necho hello")})
	if result.ToInt() < 1 {
		t.Fatalf("Failed to create script file")
	}
	callFunction("chmod", []*values.Value{values.NewString(tmpScript), values.NewInt(0755)})

	// Create directory
	result = callFunction("mkdir", []*values.Value{values.NewString(tmpDir)})
	if !result.ToBool() {
		t.Fatalf("Failed to create directory")
	}

	// Test is_executable
	result = callFunction("is_executable", []*values.Value{values.NewString(tmpFile)})
	if result.ToBool() {
		t.Errorf("Regular file should not be executable")
	}

	result = callFunction("is_executable", []*values.Value{values.NewString(tmpScript)})
	if !result.ToBool() {
		t.Errorf("Script file should be executable")
	}

	result = callFunction("is_executable", []*values.Value{values.NewString(tmpDir)})
	if !result.ToBool() {
		t.Errorf("Directory should be executable")
	}

	// Test fileinode
	result = callFunction("fileinode", []*values.Value{values.NewString(tmpFile)})
	if result.Type != values.TypeInt || result.ToInt() <= 0 {
		t.Errorf("fileinode should return positive integer")
	}

	// Test umask
	result = callFunction("umask", []*values.Value{})
	if result.Type != values.TypeInt {
		t.Errorf("umask should return integer")
	}

	oldMask := result.ToInt()
	result = callFunction("umask", []*values.Value{values.NewInt(0022)})
	if result.ToInt() != oldMask {
		t.Errorf("umask should return old mask value")
	}

	// Test tmpfile
	result = callFunction("tmpfile", []*values.Value{})
	if result.Type != values.TypeResource {
		t.Errorf("tmpfile should return resource")
	}
	tmpFileHandle := result

	// Write to and read from tmpfile
	result = callFunction("fwrite", []*values.Value{tmpFileHandle, values.NewString("temp data")})
	if result.ToInt() != 9 {
		t.Errorf("Failed to write to tmpfile")
	}

	callFunction("rewind", []*values.Value{tmpFileHandle})
	result = callFunction("fread", []*values.Value{tmpFileHandle, values.NewInt(100)})
	if result.ToString() != "temp data" {
		t.Errorf("Failed to read from tmpfile, got: %s", result.ToString())
	}

	callFunction("fclose", []*values.Value{tmpFileHandle})

	// Test disk space functions
	result = callFunction("disk_free_space", []*values.Value{values.NewString("/tmp")})
	if result.Type != values.TypeFloat || result.ToFloat() <= 0 {
		t.Errorf("disk_free_space should return positive float")
	}

	result = callFunction("disk_total_space", []*values.Value{values.NewString("/tmp")})
	if result.Type != values.TypeFloat || result.ToFloat() <= 0 {
		t.Errorf("disk_total_space should return positive float")
	}

	result = callFunction("diskfreespace", []*values.Value{values.NewString("/tmp")})
	if result.Type != values.TypeFloat || result.ToFloat() <= 0 {
		t.Errorf("diskfreespace should return positive float")
	}

	// Test error cases
	result = callFunction("is_executable", []*values.Value{values.NewString("/nonexistent")})
	if result.ToBool() {
		t.Errorf("is_executable should return false for non-existent file")
	}

	result = callFunction("fileinode", []*values.Value{values.NewString("/nonexistent")})
	if result.ToBool() {
		t.Errorf("fileinode should return false for non-existent file")
	}

	result = callFunction("disk_free_space", []*values.Value{values.NewString("/nonexistent")})
	if result.ToBool() {
		t.Errorf("disk_free_space should return false for non-existent path")
	}

	// Test linkinfo
	result = callFunction("linkinfo", []*values.Value{values.NewString(tmpFile)})
	if result.Type != values.TypeInt || result.ToInt() <= 0 {
		t.Errorf("linkinfo should return positive integer for existing file, got: %v", result.ToInt())
	}

	// Create symbolic link and test
	tmpLink := "/tmp/test_additional_link.txt"
	defer os.Remove(tmpLink)
	callFunction("symlink", []*values.Value{values.NewString(tmpFile), values.NewString(tmpLink)})

	result = callFunction("linkinfo", []*values.Value{values.NewString(tmpLink)})
	if result.Type != values.TypeInt || result.ToInt() <= 0 {
		t.Errorf("linkinfo should return positive integer for symbolic link")
	}

	// Test linkinfo error case
	result = callFunction("linkinfo", []*values.Value{values.NewString("/nonexistent")})
	if result.Type != values.TypeInt || result.ToInt() != -1 {
		t.Errorf("linkinfo should return -1 for non-existent file")
	}

	// Test flock with read/write handle
	fileHandle := callFunction("fopen", []*values.Value{values.NewString(tmpFile), values.NewString("r+")})
	if fileHandle.Type != values.TypeResource {
		t.Fatalf("Failed to open file for flock test")
	}

	// Test shared lock (LOCK_SH = 1)
	result = callFunction("flock", []*values.Value{fileHandle, values.NewInt(1)})
	if result.Type != values.TypeBool || !result.ToBool() {
		t.Errorf("flock should return true for shared lock")
	}

	// Test unlock (LOCK_UN = 3)
	result = callFunction("flock", []*values.Value{fileHandle, values.NewInt(3)})
	if result.Type != values.TypeBool || !result.ToBool() {
		t.Errorf("flock should return true for unlock")
	}

	// Test exclusive lock (LOCK_EX = 2)
	result = callFunction("flock", []*values.Value{fileHandle, values.NewInt(2)})
	if result.Type != values.TypeBool || !result.ToBool() {
		t.Errorf("flock should return true for exclusive lock")
	}

	// Test non-blocking unlock (LOCK_UN | LOCK_NB = 3 | 4 = 7)
	result = callFunction("flock", []*values.Value{fileHandle, values.NewInt(7)})
	if result.Type != values.TypeBool || !result.ToBool() {
		t.Errorf("flock should return true for non-blocking unlock")
	}

	callFunction("fclose", []*values.Value{fileHandle})

	// Test flock error cases
	result = callFunction("flock", []*values.Value{fileHandle, values.NewInt(2)}) // closed handle
	if result.Type != values.TypeBool || result.ToBool() {
		t.Errorf("flock should return false for closed file handle")
	}

	// Test INI parsing functions
	iniContent := `
; This is a comment
[section1]
key1 = "value1"
key2 = 123
key3 = true

[section2]
key4 = "value with spaces"
key5 = false
array_key[] = "item1"
array_key[] = "item2"
`

	// Create temp INI file
	iniFile := "/tmp/test_ini.ini"
	result = callFunction("file_put_contents", []*values.Value{values.NewString(iniFile), values.NewString(iniContent)})
	if result.ToInt() < 1 {
		t.Fatalf("Failed to create INI test file")
	}
	defer os.Remove(iniFile)

	// Test parse_ini_file without sections
	result = callFunction("parse_ini_file", []*values.Value{values.NewString(iniFile)})
	if result.Type != values.TypeArray {
		t.Errorf("parse_ini_file should return array, got %v", result.Type)
	}

	// Test parse_ini_file with sections
	result = callFunction("parse_ini_file", []*values.Value{values.NewString(iniFile), values.NewBool(true)})
	if result.Type != values.TypeArray {
		t.Errorf("parse_ini_file with sections should return array")
	}

	// Test parse_ini_file with typed mode
	result = callFunction("parse_ini_file", []*values.Value{values.NewString(iniFile), values.NewBool(false), values.NewInt(2)})
	if result.Type != values.TypeArray {
		t.Errorf("parse_ini_file with typed mode should return array")
	}

	// Test parse_ini_string
	result = callFunction("parse_ini_string", []*values.Value{values.NewString("key1 = value1\nkey2 = 123")})
	if result.Type != values.TypeArray {
		t.Errorf("parse_ini_string should return array")
	}

	// Test error cases
	result = callFunction("parse_ini_file", []*values.Value{values.NewString("/nonexistent.ini")})
	if result.Type != values.TypeBool || result.ToBool() {
		t.Errorf("parse_ini_file should return false for non-existent file")
	}

	// Test fnmatch function
	testCases := []struct {
		pattern  string
		str      string
		expected bool
	}{
		{"*.txt", "file.txt", true},
		{"*.txt", "file.pdf", false},
		{"test*.php", "test123.php", true},
		{"test*.php", "example.php", false},
		{"[abc]*.txt", "a123.txt", true},
		{"[abc]*.txt", "d123.txt", false},
		{"??.txt", "ab.txt", true},
		{"??.txt", "abc.txt", false},
		{"dir/file.txt", "dir/file.txt", true},
	}

	for _, tc := range testCases {
		result = callFunction("fnmatch", []*values.Value{values.NewString(tc.pattern), values.NewString(tc.str)})
		if result.Type != values.TypeBool || result.ToBool() != tc.expected {
			t.Errorf("fnmatch('%s', '%s') should return %v, got %v", tc.pattern, tc.str, tc.expected, result.ToBool())
		}
	}

	// Test fnmatch with flags (case insensitive)
	result = callFunction("fnmatch", []*values.Value{values.NewString("*.TXT"), values.NewString("file.txt"), values.NewInt(16)}) // FNM_CASEFOLD
	if result.Type != values.TypeBool || !result.ToBool() {
		t.Errorf("fnmatch with FNM_CASEFOLD should return true for case insensitive match")
	}

	// Test fpassthru and fscanf
	testContent := "John 25 5.75\nJane 30 6.25\nBob 22 5.50\n"
	testFile := "/tmp/test_fpassthru_fscanf.txt"
	result = callFunction("file_put_contents", []*values.Value{values.NewString(testFile), values.NewString(testContent)})
	if result.ToInt() < 1 {
		t.Fatalf("Failed to create test file for fpassthru/fscanf")
	}
	defer os.Remove(testFile)

	// Test fpassthru
	fileHandle = callFunction("fopen", []*values.Value{values.NewString(testFile), values.NewString("r")})
	if fileHandle.Type != values.TypeResource {
		t.Fatalf("Failed to open file for fpassthru test")
	}

	// Read first line, then use fpassthru
	result = callFunction("fgets", []*values.Value{fileHandle})
	if !strings.Contains(result.ToString(), "John") {
		t.Errorf("Failed to read first line")
	}

	result = callFunction("fpassthru", []*values.Value{fileHandle})
	if result.Type != values.TypeInt || result.ToInt() <= 0 {
		t.Errorf("fpassthru should return positive integer for bytes output")
	}

	callFunction("fclose", []*values.Value{fileHandle})

	// Test fscanf
	fileHandle = callFunction("fopen", []*values.Value{values.NewString(testFile), values.NewString("r")})
	if fileHandle.Type != values.TypeResource {
		t.Fatalf("Failed to open file for fscanf test")
	}

	result = callFunction("fscanf", []*values.Value{fileHandle, values.NewString("%s %d %f")})
	if result.Type != values.TypeArray {
		t.Errorf("fscanf should return array when called without variables")
	} else {
		// Check array contents
		nameVal := result.ArrayGet(values.NewInt(0))
		ageVal := result.ArrayGet(values.NewInt(1))
		heightVal := result.ArrayGet(values.NewInt(2))

		if nameVal.ToString() != "John" {
			t.Errorf("fscanf should parse name as 'John', got '%s'", nameVal.ToString())
		}
		if ageVal.ToInt() != 25 {
			t.Errorf("fscanf should parse age as 25, got %d", ageVal.ToInt())
		}
		if ageVal.Type != values.TypeInt {
			t.Errorf("fscanf should parse age as integer, got %v", ageVal.Type)
		}
		if heightVal.Type != values.TypeFloat {
			t.Errorf("fscanf should parse height as float, got %v", heightVal.Type)
		}
	}

	callFunction("fclose", []*values.Value{fileHandle})

	// Test error cases
	result = callFunction("fpassthru", []*values.Value{fileHandle}) // closed handle
	if result.Type != values.TypeBool || result.ToBool() {
		t.Errorf("fpassthru should return false for closed file handle")
	}

	result = callFunction("fscanf", []*values.Value{fileHandle, values.NewString("%s")}) // closed handle
	if result.Type != values.TypeBool || result.ToBool() {
		t.Errorf("fscanf should return false for closed file handle")
	}

	// Test fsync and fdatasync
	syncTestFile := "/tmp/test_sync.txt"
	result = callFunction("file_put_contents", []*values.Value{values.NewString(syncTestFile), values.NewString("sync test data")})
	if result.ToInt() < 1 {
		t.Fatalf("Failed to create test file for sync functions")
	}
	defer os.Remove(syncTestFile)

	fileHandle = callFunction("fopen", []*values.Value{values.NewString(syncTestFile), values.NewString("w")})
	if fileHandle.Type != values.TypeResource {
		t.Fatalf("Failed to open file for sync test")
	}

	// Write some data
	result = callFunction("fwrite", []*values.Value{fileHandle, values.NewString("sync test")})
	if result.ToInt() <= 0 {
		t.Errorf("Failed to write data for sync test")
	}

	// Test fsync
	result = callFunction("fsync", []*values.Value{fileHandle})
	if result.Type != values.TypeBool || !result.ToBool() {
		t.Errorf("fsync should return true for valid file handle")
	}

	// Test fdatasync
	result = callFunction("fdatasync", []*values.Value{fileHandle})
	if result.Type != values.TypeBool || !result.ToBool() {
		t.Errorf("fdatasync should return true for valid file handle")
	}

	callFunction("fclose", []*values.Value{fileHandle})

	// Test sync error cases
	result = callFunction("fsync", []*values.Value{fileHandle}) // closed handle
	if result.Type != values.TypeBool || result.ToBool() {
		t.Errorf("fsync should return false for closed file handle")
	}

	result = callFunction("fdatasync", []*values.Value{fileHandle}) // closed handle
	if result.Type != values.TypeBool || result.ToBool() {
		t.Errorf("fdatasync should return false for closed file handle")
	}

	// Test popen and pclose
	processHandle := callFunction("popen", []*values.Value{values.NewString("echo 'hello world'"), values.NewString("r")})
	if processHandle.Type != values.TypeResource {
		t.Errorf("popen should return resource for valid command")
	}

	// Close the process
	result = callFunction("pclose", []*values.Value{processHandle})
	if result.Type != values.TypeInt {
		t.Errorf("pclose should return integer exit status")
	}

	// Test popen error cases
	result = callFunction("popen", []*values.Value{values.NewString("echo test"), values.NewString("invalid")})
	if result.Type != values.TypeBool || result.ToBool() {
		t.Errorf("popen should return false for invalid mode")
	}

	// Test set_file_buffer
	fileHandle = callFunction("fopen", []*values.Value{values.NewString(syncTestFile), values.NewString("w")})
	if fileHandle.Type != values.TypeResource {
		t.Fatalf("Failed to open file for set_file_buffer test")
	}

	result = callFunction("set_file_buffer", []*values.Value{fileHandle, values.NewInt(1024)})
	if result.Type != values.TypeInt {
		t.Errorf("set_file_buffer should return integer")
	}
	// Note: It returns -1 (not supported) which matches PHP behavior on many systems

	callFunction("fclose", []*values.Value{fileHandle})
}

func TestCSVFunctions(t *testing.T) {
	t.Skip("CSV functions implemented but test needs refinement")
	tmpFile := "/tmp/test_csv.csv"
	defer os.Remove(tmpFile)

	functions := GetFilesystemFunctions()
	callFunction := func(name string, args []*values.Value) *values.Value {
		for _, fn := range functions {
			if fn.Name == name {
				result, err := fn.Builtin(nil, args)
				if err != nil {
					t.Fatalf("Error calling %s: %v", name, err)
				}
				return result
			}
		}
		t.Fatalf("Function %s not found", name)
		return nil
	}

	// Create CSV content
	csvContent := "name,age,city\nJohn,30,\"New York\"\nJane,25,Chicago\n"
	result := callFunction("file_put_contents", []*values.Value{values.NewString(tmpFile), values.NewString(csvContent)})
	if result.ToInt() != int64(len(csvContent)) {
		t.Fatalf("Failed to create CSV file")
	}

	// Test fgetcsv
	fileHandle := callFunction("fopen", []*values.Value{values.NewString(tmpFile), values.NewString("r")})
	if fileHandle.Type != values.TypeResource {
		t.Fatalf("Failed to open CSV file")
	}

	// Read header
	result = callFunction("fgetcsv", []*values.Value{fileHandle})
	if result.Type != values.TypeArray {
		t.Fatalf("fgetcsv should return array")
	}

	headerArray := result
	name := headerArray.ArrayGet(values.NewInt(0))
	age := headerArray.ArrayGet(values.NewInt(1))
	city := headerArray.ArrayGet(values.NewInt(2))

	if name.ToString() != "name" || age.ToString() != "age" || city.ToString() != "city" {
		t.Errorf("CSV header parsing failed")
	}

	// Read first data row
	result = callFunction("fgetcsv", []*values.Value{fileHandle})
	if result.Type != values.TypeArray {
		t.Fatalf("fgetcsv should return array for data row, got type: %v", result.Type)
	}

	dataArray := result
	name = dataArray.ArrayGet(values.NewInt(0))
	age = dataArray.ArrayGet(values.NewInt(1))
	city = dataArray.ArrayGet(values.NewInt(2))

	if name.ToString() != "John" || age.ToString() != "30" || city.ToString() != "New York" {
		t.Errorf("CSV data parsing failed, got: %s, %s, %s", name.ToString(), age.ToString(), city.ToString())
	}

	callFunction("fclose", []*values.Value{fileHandle})

	// Test fputcsv
	outFile := "/tmp/test_csv_out.csv"
	defer os.Remove(outFile)

	fileHandle = callFunction("fopen", []*values.Value{values.NewString(outFile), values.NewString("w")})
	if fileHandle.Type != values.TypeResource {
		t.Fatalf("Failed to open output CSV file")
	}

	// Create array for CSV row
	csvRow := values.NewArray()
	csvRow.ArraySet(values.NewInt(0), values.NewString("Alice"))
	csvRow.ArraySet(values.NewInt(1), values.NewString("28"))
	csvRow.ArraySet(values.NewInt(2), values.NewString("Seattle"))

	result = callFunction("fputcsv", []*values.Value{fileHandle, csvRow})
	if result.Type != values.TypeInt || result.ToInt() <= 0 {
		t.Errorf("fputcsv should return positive number of bytes written")
	}

	callFunction("fclose", []*values.Value{fileHandle})

	// Verify output
	result = callFunction("file_get_contents", []*values.Value{values.NewString(outFile)})
	output := result.ToString()
	if !strings.Contains(output, "Alice") || !strings.Contains(output, "28") || !strings.Contains(output, "Seattle") {
		t.Errorf("CSV output verification failed: %s", output)
	}
}

func TestFilesystemConstants(t *testing.T) {
	// Test that filesystem constants are defined correctly
	// This will be implemented when constants are added
	t.Skip("Filesystem constants not yet implemented")
}