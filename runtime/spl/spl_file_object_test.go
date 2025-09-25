package spl

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

func TestSplFileObject(t *testing.T) {
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

	t.Run("BasicFileReading", func(t *testing.T) {
		testBasicFileReading(t, ctx)
	})

	t.Run("FileIteration", func(t *testing.T) {
		testFileIteration(t, ctx)
	})

	t.Run("FileFlags", func(t *testing.T) {
		testFileFlags(t, ctx)
	})

	t.Run("FileOperations", func(t *testing.T) {
		testFileOperations(t, ctx)
	})

	t.Run("CSVFunctionality", func(t *testing.T) {
		testCSVFunctionality(t, ctx)
	})

	t.Run("FileWriting", func(t *testing.T) {
		testFileWriting(t, ctx)
	})

	t.Run("SplFileObjectConstants", func(t *testing.T) {
		testSplFileObjectConstants(t, ctx)
	})

	t.Run("EdgeCases", func(t *testing.T) {
		testEdgeCases(t, ctx)
	})
}

func testBasicFileReading(t *testing.T, ctx *mockContext) {
	// Create test file
	testFile := "/tmp/test_spl_file_object_basic.txt"
	content := "line 1\nline 2\nline 3\n"
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(testFile)

	// Get the SplFileObject class
	class, err := ctx.registry.GetClass("SplFileObject")
	if err != nil {
		t.Fatalf("SplFileObject class not found: %v", err)
	}

	// Create SplFileObject instance
	obj := &values.Object{
		ClassName:  "SplFileObject",
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
	_, err = constructorImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj, values.NewString(testFile), values.NewString("r")})
	if err != nil {
		t.Fatalf("Constructor failed: %v", err)
	}

	// Test file information methods (inherited from SplFileInfo)
	getBasenameMethod := class.Methods["getBasename"]
	if getBasenameMethod != nil {
		getBasenameImpl := getBasenameMethod.Implementation.(*BuiltinMethodImpl)
		basenameResult, err := getBasenameImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("getBasename failed: %v", err)
		}
		expectedBasename := filepath.Base(testFile)
		if basenameResult.ToString() != expectedBasename {
			t.Errorf("Expected basename %s, got %s", expectedBasename, basenameResult.ToString())
		}
	}

	// Test basic file operations
	currentMethod := class.Methods["current"]
	currentImpl := currentMethod.Implementation.(*BuiltinMethodImpl)

	// Test first line
	currentResult, err := currentImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
	if err != nil {
		t.Fatalf("Current failed: %v", err)
	}
	if currentResult.ToString() != "line 1\n" {
		t.Errorf("Expected 'line 1\\n', got '%s'", currentResult.ToString())
	}

	// Test key
	keyMethod := class.Methods["key"]
	keyImpl := keyMethod.Implementation.(*BuiltinMethodImpl)
	keyResult, err := keyImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
	if err != nil {
		t.Fatalf("Key failed: %v", err)
	}
	if keyResult.ToInt() != 0 {
		t.Errorf("Expected key 0, got %d", keyResult.ToInt())
	}

	// Test next and current
	nextMethod := class.Methods["next"]
	nextImpl := nextMethod.Implementation.(*BuiltinMethodImpl)
	_, err = nextImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
	if err != nil {
		t.Fatalf("Next failed: %v", err)
	}

	currentResult, err = currentImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
	if err != nil {
		t.Fatalf("Current failed: %v", err)
	}
	if currentResult.ToString() != "line 2\n" {
		t.Errorf("Expected 'line 2\\n', got '%s'", currentResult.ToString())
	}
}

func testFileIteration(t *testing.T, ctx *mockContext) {
	// Create test file
	testFile := "/tmp/test_spl_file_iteration.txt"
	content := "first\nsecond\nthird\n"
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(testFile)

	// Get the SplFileObject class
	class, err := ctx.registry.GetClass("SplFileObject")
	if err != nil {
		t.Fatalf("SplFileObject class not found: %v", err)
	}

	// Create SplFileObject instance
	obj := &values.Object{
		ClassName:  "SplFileObject",
		Properties: make(map[string]*values.Value),
	}
	thisObj := &values.Value{
		Type: values.TypeObject,
		Data: obj,
	}

	// Initialize
	constructorMethod := class.Methods["__construct"]
	constructorImpl := constructorMethod.Implementation.(*BuiltinMethodImpl)
	_, err = constructorImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj, values.NewString(testFile), values.NewString("r")})
	if err != nil {
		t.Fatalf("Constructor failed: %v", err)
	}

	// Test full iteration
	methods := map[string]*BuiltinMethodImpl{
		"valid":   class.Methods["valid"].Implementation.(*BuiltinMethodImpl),
		"key":     class.Methods["key"].Implementation.(*BuiltinMethodImpl),
		"current": class.Methods["current"].Implementation.(*BuiltinMethodImpl),
		"next":    class.Methods["next"].Implementation.(*BuiltinMethodImpl),
		"rewind":  class.Methods["rewind"].Implementation.(*BuiltinMethodImpl),
	}

	// Start from beginning
	_, err = methods["rewind"].GetFunction().Builtin(ctx, []*values.Value{thisObj})
	if err != nil {
		t.Fatalf("Rewind failed: %v", err)
	}

	expectedLines := []string{"first\n", "second\n", "third\n", ""}
	lineIndex := 0

	for {
		validResult, err := methods["valid"].GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("Valid check failed: %v", err)
		}
		if !validResult.ToBool() {
			break
		}

		keyResult, err := methods["key"].GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("Key failed: %v", err)
		}
		if keyResult.ToInt() != int64(lineIndex) {
			t.Errorf("Expected key %d, got %d", lineIndex, keyResult.ToInt())
		}

		currentResult, err := methods["current"].GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("Current failed: %v", err)
		}

		if lineIndex < len(expectedLines) {
			if currentResult.ToString() != expectedLines[lineIndex] {
				t.Errorf("Line %d: expected '%s', got '%s'", lineIndex, expectedLines[lineIndex], currentResult.ToString())
			}
		}

		_, err = methods["next"].GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("Next failed: %v", err)
		}

		lineIndex++
	}

	if lineIndex != len(expectedLines) {
		t.Errorf("Expected %d lines, got %d", len(expectedLines), lineIndex)
	}
}

func testFileFlags(t *testing.T, ctx *mockContext) {
	// Create test file
	testFile := "/tmp/test_spl_file_flags.txt"
	content := "line1\n\nline3\n\nline5\n"
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(testFile)

	// Get the SplFileObject class
	class, err := ctx.registry.GetClass("SplFileObject")
	if err != nil {
		t.Fatalf("SplFileObject class not found: %v", err)
	}

	// Test DROP_NEW_LINE flag
	t.Run("DROP_NEW_LINE", func(t *testing.T) {
		obj := &values.Object{
			ClassName:  "SplFileObject",
			Properties: make(map[string]*values.Value),
		}
		thisObj := &values.Value{
			Type: values.TypeObject,
			Data: obj,
		}

		constructorMethod := class.Methods["__construct"]
		constructorImpl := constructorMethod.Implementation.(*BuiltinMethodImpl)
		_, err = constructorImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj, values.NewString(testFile), values.NewString("r")})
		if err != nil {
			t.Fatalf("Constructor failed: %v", err)
		}

		setFlagsMethod := class.Methods["setFlags"]
		setFlagsImpl := setFlagsMethod.Implementation.(*BuiltinMethodImpl)
		_, err = setFlagsImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj, values.NewInt(DROP_NEW_LINE)})
		if err != nil {
			t.Fatalf("setFlags failed: %v", err)
		}

		currentMethod := class.Methods["current"]
		currentImpl := currentMethod.Implementation.(*BuiltinMethodImpl)
		currentResult, err := currentImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("Current failed: %v", err)
		}

		// Should not have newline
		if currentResult.ToString() != "line1" {
			t.Errorf("Expected 'line1' without newline, got '%s'", currentResult.ToString())
		}
	})

	// Test SKIP_EMPTY flag
	t.Run("SKIP_EMPTY", func(t *testing.T) {
		obj := &values.Object{
			ClassName:  "SplFileObject",
			Properties: make(map[string]*values.Value),
		}
		thisObj := &values.Value{
			Type: values.TypeObject,
			Data: obj,
		}

		constructorMethod := class.Methods["__construct"]
		constructorImpl := constructorMethod.Implementation.(*BuiltinMethodImpl)
		_, err = constructorImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj, values.NewString(testFile), values.NewString("r")})
		if err != nil {
			t.Fatalf("Constructor failed: %v", err)
		}

		setFlagsMethod := class.Methods["setFlags"]
		setFlagsImpl := setFlagsMethod.Implementation.(*BuiltinMethodImpl)
		_, err = setFlagsImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj, values.NewInt(SKIP_EMPTY)})
		if err != nil {
			t.Fatalf("setFlags failed: %v", err)
		}

		// Collect non-empty lines
		methods := map[string]*BuiltinMethodImpl{
			"valid":   class.Methods["valid"].Implementation.(*BuiltinMethodImpl),
			"current": class.Methods["current"].Implementation.(*BuiltinMethodImpl),
			"next":    class.Methods["next"].Implementation.(*BuiltinMethodImpl),
			"rewind":  class.Methods["rewind"].Implementation.(*BuiltinMethodImpl),
		}

		_, err = methods["rewind"].GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("Rewind failed: %v", err)
		}

		var nonEmptyLines []string
		for {
			validResult, err := methods["valid"].GetFunction().Builtin(ctx, []*values.Value{thisObj})
			if err != nil {
				t.Fatalf("Valid check failed: %v", err)
			}
			if !validResult.ToBool() {
				break
			}

			currentResult, err := methods["current"].GetFunction().Builtin(ctx, []*values.Value{thisObj})
			if err != nil {
				t.Fatalf("Current failed: %v", err)
			}

			line := currentResult.ToString()
			if line != "" {
				nonEmptyLines = append(nonEmptyLines, line)
			}

			_, err = methods["next"].GetFunction().Builtin(ctx, []*values.Value{thisObj})
			if err != nil {
				t.Fatalf("Next failed: %v", err)
			}
		}

		// Should have skipped empty lines
		expectedNonEmpty := []string{"line1\n", "line3\n", "line5\n"}
		if len(nonEmptyLines) != len(expectedNonEmpty) {
			t.Errorf("Expected %d non-empty lines, got %d", len(expectedNonEmpty), len(nonEmptyLines))
		}
	})
}

func testFileOperations(t *testing.T, ctx *mockContext) {
	// Create test file
	testFile := "/tmp/test_spl_file_operations.txt"
	content := "line 1\nline 2\nline 3\n"
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(testFile)

	// Get the SplFileObject class
	class, err := ctx.registry.GetClass("SplFileObject")
	if err != nil {
		t.Fatalf("SplFileObject class not found: %v", err)
	}

	// Create SplFileObject instance
	obj := &values.Object{
		ClassName:  "SplFileObject",
		Properties: make(map[string]*values.Value),
	}
	thisObj := &values.Value{
		Type: values.TypeObject,
		Data: obj,
	}

	// Initialize
	constructorMethod := class.Methods["__construct"]
	constructorImpl := constructorMethod.Implementation.(*BuiltinMethodImpl)
	_, err = constructorImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj, values.NewString(testFile), values.NewString("r")})
	if err != nil {
		t.Fatalf("Constructor failed: %v", err)
	}

	// Test seek
	seekMethod := class.Methods["seek"]
	seekImpl := seekMethod.Implementation.(*BuiltinMethodImpl)
	_, err = seekImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj, values.NewInt(1)})
	if err != nil {
		t.Fatalf("Seek failed: %v", err)
	}

	currentMethod := class.Methods["current"]
	currentImpl := currentMethod.Implementation.(*BuiltinMethodImpl)
	currentResult, err := currentImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
	if err != nil {
		t.Fatalf("Current failed: %v", err)
	}

	if currentResult.ToString() != "line 2\n" {
		t.Errorf("Expected 'line 2\\n' after seek, got '%s'", currentResult.ToString())
	}

	// Test eof
	eofMethod := class.Methods["eof"]
	eofImpl := eofMethod.Implementation.(*BuiltinMethodImpl)
	eofResult, err := eofImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
	if err != nil {
		t.Fatalf("EOF failed: %v", err)
	}

	if eofResult.ToBool() {
		t.Error("Should not be at EOF")
	}

	// Test fgets
	fgetsMethod := class.Methods["fgets"]
	fgetsImpl := fgetsMethod.Implementation.(*BuiltinMethodImpl)
	fgetsResult, err := fgetsImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
	if err != nil {
		t.Fatalf("fgets failed: %v", err)
	}

	if fgetsResult.ToString() != "line 2\n" {
		t.Errorf("Expected 'line 2\\n' from fgets, got '%s'", fgetsResult.ToString())
	}
}

func testCSVFunctionality(t *testing.T, ctx *mockContext) {
	// Create CSV test file
	csvFile := "/tmp/test_csv_functionality.csv"
	csvContent := "name,age,city\nJohn,25,\"New York\"\nJane,30,London\n"
	err := os.WriteFile(csvFile, []byte(csvContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create CSV file: %v", err)
	}
	defer os.Remove(csvFile)

	// Get the SplFileObject class
	class, err := ctx.registry.GetClass("SplFileObject")
	if err != nil {
		t.Fatalf("SplFileObject class not found: %v", err)
	}

	// Create SplFileObject instance
	obj := &values.Object{
		ClassName:  "SplFileObject",
		Properties: make(map[string]*values.Value),
	}
	thisObj := &values.Value{
		Type: values.TypeObject,
		Data: obj,
	}

	// Initialize
	constructorMethod := class.Methods["__construct"]
	constructorImpl := constructorMethod.Implementation.(*BuiltinMethodImpl)
	_, err = constructorImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj, values.NewString(csvFile), values.NewString("r")})
	if err != nil {
		t.Fatalf("Constructor failed: %v", err)
	}

	// Set READ_CSV flag
	setFlagsMethod := class.Methods["setFlags"]
	setFlagsImpl := setFlagsMethod.Implementation.(*BuiltinMethodImpl)
	_, err = setFlagsImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj, values.NewInt(READ_CSV)})
	if err != nil {
		t.Fatalf("setFlags failed: %v", err)
	}

	// Test CSV parsing
	methods := map[string]*BuiltinMethodImpl{
		"valid":   class.Methods["valid"].Implementation.(*BuiltinMethodImpl),
		"current": class.Methods["current"].Implementation.(*BuiltinMethodImpl),
		"next":    class.Methods["next"].Implementation.(*BuiltinMethodImpl),
		"rewind":  class.Methods["rewind"].Implementation.(*BuiltinMethodImpl),
	}

	_, err = methods["rewind"].GetFunction().Builtin(ctx, []*values.Value{thisObj})
	if err != nil {
		t.Fatalf("Rewind failed: %v", err)
	}

	expectedCSVData := [][]string{
		{"name", "age", "city"},
		{"John", "25", "New York"},
		{"Jane", "30", "London"},
	}

	lineIndex := 0
	for {
		validResult, err := methods["valid"].GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("Valid check failed: %v", err)
		}
		if !validResult.ToBool() {
			break
		}

		currentResult, err := methods["current"].GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("Current failed: %v", err)
		}

		// Skip the final empty line
		if lineIndex >= len(expectedCSVData) {
			break
		}

		if currentResult.Type == values.TypeArray {
			expected := expectedCSVData[lineIndex]
			arrayCount := currentResult.ArrayCount()

			if arrayCount != len(expected) {
				t.Errorf("Line %d: expected %d CSV fields, got %d", lineIndex, len(expected), arrayCount)
			} else {
				for i, expectedField := range expected {
					field := currentResult.ArrayGet(values.NewInt(int64(i)))
					if field == nil {
						t.Errorf("Line %d: missing field at index %d", lineIndex, i)
						continue
					}
					if field.ToString() != expectedField {
						t.Errorf("Line %d, field %d: expected '%s', got '%s'", lineIndex, i, expectedField, field.ToString())
					}
				}
			}
		} else {
			t.Errorf("Line %d: expected array for CSV data, got %s", lineIndex, currentResult.Type)
		}

		_, err = methods["next"].GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("Next failed: %v", err)
		}

		lineIndex++
	}
}

func testFileWriting(t *testing.T, ctx *mockContext) {
	// Create test file for writing
	writeFile := "/tmp/test_spl_file_write.txt"

	// Get the SplFileObject class
	class, err := ctx.registry.GetClass("SplFileObject")
	if err != nil {
		t.Fatalf("SplFileObject class not found: %v", err)
	}

	// Create SplFileObject instance for writing
	obj := &values.Object{
		ClassName:  "SplFileObject",
		Properties: make(map[string]*values.Value),
	}
	thisObj := &values.Value{
		Type: values.TypeObject,
		Data: obj,
	}

	// Initialize for writing
	constructorMethod := class.Methods["__construct"]
	constructorImpl := constructorMethod.Implementation.(*BuiltinMethodImpl)
	_, err = constructorImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj, values.NewString(writeFile), values.NewString("w")})
	if err != nil {
		t.Fatalf("Constructor failed: %v", err)
	}
	defer os.Remove(writeFile)

	// Test fwrite
	fwriteMethod := class.Methods["fwrite"]
	fwriteImpl := fwriteMethod.Implementation.(*BuiltinMethodImpl)
	bytesResult, err := fwriteImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj, values.NewString("Hello World\n")})
	if err != nil {
		t.Fatalf("fwrite failed: %v", err)
	}

	expectedBytes := int64(12) // "Hello World\n"
	if bytesResult.ToInt() != expectedBytes {
		t.Errorf("Expected %d bytes written, got %d", expectedBytes, bytesResult.ToInt())
	}

	// Write more data
	_, err = fwriteImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj, values.NewString("Second line\n")})
	if err != nil {
		t.Fatalf("Second fwrite failed: %v", err)
	}

	// Need to create a new instance to read the file (close the writer)
	// In real usage, the file would be closed properly
}

func testSplFileObjectConstants(t *testing.T, ctx *mockContext) {
	// Get the SplFileObject class
	class, err := ctx.registry.GetClass("SplFileObject")
	if err != nil {
		t.Fatalf("SplFileObject class not found: %v", err)
	}

	// Test all constants
	expectedConstants := map[string]int{
		"DROP_NEW_LINE": DROP_NEW_LINE,
		"READ_AHEAD":    READ_AHEAD,
		"SKIP_EMPTY":    SKIP_EMPTY,
		"READ_CSV":      READ_CSV,
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

func testEdgeCases(t *testing.T, ctx *mockContext) {
	// Get the SplFileObject class
	class, err := ctx.registry.GetClass("SplFileObject")
	if err != nil {
		t.Fatalf("SplFileObject class not found: %v", err)
	}

	// Test non-existent file
	t.Run("NonExistentFile", func(t *testing.T) {
		obj := &values.Object{
			ClassName:  "SplFileObject",
			Properties: make(map[string]*values.Value),
		}
		thisObj := &values.Value{
			Type: values.TypeObject,
			Data: obj,
		}

		constructorMethod := class.Methods["__construct"]
		constructorImpl := constructorMethod.Implementation.(*BuiltinMethodImpl)
		_, err := constructorImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj, values.NewString("/non/existent/file.txt"), values.NewString("r")})
		if err == nil {
			t.Error("Expected error when opening non-existent file")
		}
	})

	// Test inheritance from SplFileInfo
	t.Run("InheritanceFromSplFileInfo", func(t *testing.T) {
		// SplFileObject should have methods from SplFileInfo
		requiredMethods := []string{"getBasename", "getExtension", "getPath", "getPathname", "getFilename"}
		for _, methodName := range requiredMethods {
			if _, exists := class.Methods[methodName]; !exists {
				t.Errorf("SplFileObject should have method %s", methodName)
			}
		}

		// Check parent class
		if class.Parent != "SplFileInfo" {
			t.Errorf("SplFileObject parent should be SplFileInfo, got %s", class.Parent)
		}
	})

	// Test interfaces
	t.Run("InterfaceImplementation", func(t *testing.T) {
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
				t.Errorf("SplFileObject should implement interface %s", interfaceName)
			}
		}
	})
}