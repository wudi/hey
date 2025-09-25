package spl

import (
	"os"
	"testing"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

func TestSplTempFileObject(t *testing.T) {
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

	t.Run("BasicTempFileCreation", func(t *testing.T) {
		testBasicTempFileCreation(t, ctx)
	})

	t.Run("TempFileWriteAndRead", func(t *testing.T) {
		testTempFileWriteAndRead(t, ctx)
	})

	t.Run("TempFileIteration", func(t *testing.T) {
		testTempFileIteration(t, ctx)
	})

	t.Run("InheritanceFromSplFileObject", func(t *testing.T) {
		testInheritanceFromSplFileObject(t, ctx)
	})

	t.Run("TempFileCleanup", func(t *testing.T) {
		testTempFileCleanup(t, ctx)
	})
}

func testBasicTempFileCreation(t *testing.T, ctx *mockContext) {
	// Get the SplTempFileObject class
	class, err := ctx.registry.GetClass("SplTempFileObject")
	if err != nil {
		t.Fatalf("SplTempFileObject class not found: %v", err)
	}

	// Create SplTempFileObject instance
	obj := &values.Object{
		ClassName:  "SplTempFileObject",
		Properties: make(map[string]*values.Value),
	}
	thisObj := &values.Value{
		Type: values.TypeObject,
		Data: obj,
	}

	// Test constructor with default max memory
	constructorMethod := class.Methods["__construct"]
	if constructorMethod == nil {
		t.Fatal("__construct method not found")
	}
	constructorImpl := constructorMethod.Implementation.(*BuiltinMethodImpl)
	_, err = constructorImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
	if err != nil {
		t.Fatalf("Constructor failed: %v", err)
	}

	// Test that file was created
	getPathnameMethod := class.Methods["getPathname"]
	if getPathnameMethod == nil {
		t.Fatal("getPathname method not found")
	}
	getPathnameImpl := getPathnameMethod.Implementation.(*BuiltinMethodImpl)
	pathnameResult, err := getPathnameImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
	if err != nil {
		t.Fatalf("getPathname failed: %v", err)
	}

	tempPath := pathnameResult.ToString()
	if tempPath == "" {
		t.Error("Expected non-empty temporary file path")
	}

	// Verify the temporary file exists
	if _, err := os.Stat(tempPath); os.IsNotExist(err) {
		t.Errorf("Temporary file does not exist at path: %s", tempPath)
	}

	// Test that the file is writable
	isReadableMethod := class.Methods["isReadable"]
	isReadableImpl := isReadableMethod.Implementation.(*BuiltinMethodImpl)
	readableResult, err := isReadableImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
	if err != nil {
		t.Fatalf("isReadable failed: %v", err)
	}
	if !readableResult.ToBool() {
		t.Error("Expected temporary file to be readable")
	}
}

func testTempFileWriteAndRead(t *testing.T, ctx *mockContext) {
	// Get the SplTempFileObject class
	class, err := ctx.registry.GetClass("SplTempFileObject")
	if err != nil {
		t.Fatalf("SplTempFileObject class not found: %v", err)
	}

	// Create SplTempFileObject instance with custom max memory
	obj := &values.Object{
		ClassName:  "SplTempFileObject",
		Properties: make(map[string]*values.Value),
	}
	thisObj := &values.Value{
		Type: values.TypeObject,
		Data: obj,
	}

	constructorMethod := class.Methods["__construct"]
	constructorImpl := constructorMethod.Implementation.(*BuiltinMethodImpl)
	maxMemory := 1024 * 1024 // 1MB
	_, err = constructorImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj, values.NewInt(int64(maxMemory))})
	if err != nil {
		t.Fatalf("Constructor failed: %v", err)
	}

	// Test writing to the temporary file
	fwriteMethod := class.Methods["fwrite"]
	fwriteImpl := fwriteMethod.Implementation.(*BuiltinMethodImpl)
	testData := "Hello, World!\nSecond line\nThird line\n"
	bytesResult, err := fwriteImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj, values.NewString(testData)})
	if err != nil {
		t.Fatalf("fwrite failed: %v", err)
	}

	expectedBytes := int64(len(testData))
	if bytesResult.ToInt() != expectedBytes {
		t.Errorf("Expected %d bytes written, got %d", expectedBytes, bytesResult.ToInt())
	}

	// Test reading from the file by creating a new SplTempFileObject instance with the same file
	// First get the pathname
	getPathnameMethod := class.Methods["getPathname"]
	getPathnameImpl := getPathnameMethod.Implementation.(*BuiltinMethodImpl)
	pathnameResult, err := getPathnameImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
	if err != nil {
		t.Fatalf("getPathname failed: %v", err)
	}
	tempPath := pathnameResult.ToString()

	// Create a new SplFileObject to read the written data
	splFileObjectClass, err := ctx.registry.GetClass("SplFileObject")
	if err != nil {
		t.Fatalf("SplFileObject class not found: %v", err)
	}

	readObj := &values.Object{
		ClassName:  "SplFileObject",
		Properties: make(map[string]*values.Value),
	}
	readThisObj := &values.Value{
		Type: values.TypeObject,
		Data: readObj,
	}

	readConstructorMethod := splFileObjectClass.Methods["__construct"]
	readConstructorImpl := readConstructorMethod.Implementation.(*BuiltinMethodImpl)
	_, err = readConstructorImpl.GetFunction().Builtin(ctx, []*values.Value{readThisObj, values.NewString(tempPath), values.NewString("r")})
	if err != nil {
		t.Fatalf("Read constructor failed: %v", err)
	}

	// Read and verify the content
	currentMethod := splFileObjectClass.Methods["current"]
	currentImpl := currentMethod.Implementation.(*BuiltinMethodImpl)
	currentResult, err := currentImpl.GetFunction().Builtin(ctx, []*values.Value{readThisObj})
	if err != nil {
		t.Fatalf("Current failed: %v", err)
	}

	if currentResult.ToString() != "Hello, World!\n" {
		t.Errorf("Expected 'Hello, World!\\n', got '%s'", currentResult.ToString())
	}
}

func testTempFileIteration(t *testing.T, ctx *mockContext) {
	// Get the SplTempFileObject class
	class, err := ctx.registry.GetClass("SplTempFileObject")
	if err != nil {
		t.Fatalf("SplTempFileObject class not found: %v", err)
	}

	// Create and write to temp file
	obj := &values.Object{
		ClassName:  "SplTempFileObject",
		Properties: make(map[string]*values.Value),
	}
	thisObj := &values.Value{
		Type: values.TypeObject,
		Data: obj,
	}

	constructorMethod := class.Methods["__construct"]
	constructorImpl := constructorMethod.Implementation.(*BuiltinMethodImpl)
	_, err = constructorImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
	if err != nil {
		t.Fatalf("Constructor failed: %v", err)
	}

	// Write test data
	fwriteMethod := class.Methods["fwrite"]
	fwriteImpl := fwriteMethod.Implementation.(*BuiltinMethodImpl)
	testLines := "line1\nline2\nline3\n"
	_, err = fwriteImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj, values.NewString(testLines)})
	if err != nil {
		t.Fatalf("fwrite failed: %v", err)
	}

	// Test iteration capabilities (inherited from SplFileObject)
	validMethod := class.Methods["valid"]
	nextMethod := class.Methods["next"]
	rewindMethod := class.Methods["rewind"]

	if validMethod == nil || nextMethod == nil || rewindMethod == nil {
		t.Fatal("Required iterator methods not found (should inherit from SplFileObject)")
	}

	// The iteration would work the same as SplFileObject since it inherits all methods
	// This test mainly verifies that the inheritance is working correctly
}

func testInheritanceFromSplFileObject(t *testing.T, ctx *mockContext) {
	// Get the SplTempFileObject class
	class, err := ctx.registry.GetClass("SplTempFileObject")
	if err != nil {
		t.Fatalf("SplTempFileObject class not found: %v", err)
	}

	// Test parent class
	if class.Parent != "SplFileObject" {
		t.Errorf("SplTempFileObject parent should be SplFileObject, got %s", class.Parent)
	}

	// Test that it inherits methods from SplFileObject
	inheritedMethods := []string{"fread", "fwrite", "fgets", "fseek", "ftell", "eof", "getFlags", "setFlags", "current", "next", "valid", "rewind"}
	for _, methodName := range inheritedMethods {
		if _, exists := class.Methods[methodName]; !exists {
			t.Errorf("SplTempFileObject should inherit method %s from SplFileObject", methodName)
		}
	}

	// Test that it inherits constants from SplFileObject
	inheritedConstants := []string{"DROP_NEW_LINE", "READ_AHEAD", "SKIP_EMPTY", "READ_CSV"}
	for _, constantName := range inheritedConstants {
		if _, exists := class.Constants[constantName]; !exists {
			t.Errorf("SplTempFileObject should inherit constant %s from SplFileObject", constantName)
		}
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
			t.Errorf("SplTempFileObject should implement interface %s", interfaceName)
		}
	}
}

func testTempFileCleanup(t *testing.T, ctx *mockContext) {
	// Get the SplTempFileObject class
	class, err := ctx.registry.GetClass("SplTempFileObject")
	if err != nil {
		t.Fatalf("SplTempFileObject class not found: %v", err)
	}

	// Create SplTempFileObject instance
	obj := &values.Object{
		ClassName:  "SplTempFileObject",
		Properties: make(map[string]*values.Value),
	}
	thisObj := &values.Value{
		Type: values.TypeObject,
		Data: obj,
	}

	constructorMethod := class.Methods["__construct"]
	constructorImpl := constructorMethod.Implementation.(*BuiltinMethodImpl)
	_, err = constructorImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
	if err != nil {
		t.Fatalf("Constructor failed: %v", err)
	}

	// Get the temporary file path
	getPathnameMethod := class.Methods["getPathname"]
	getPathnameImpl := getPathnameMethod.Implementation.(*BuiltinMethodImpl)
	pathnameResult, err := getPathnameImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
	if err != nil {
		t.Fatalf("getPathname failed: %v", err)
	}

	tempPath := pathnameResult.ToString()

	// Write some data to make sure the file is created
	fwriteMethod := class.Methods["fwrite"]
	fwriteImpl := fwriteMethod.Implementation.(*BuiltinMethodImpl)
	_, err = fwriteImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj, values.NewString("test data")})
	if err != nil {
		t.Fatalf("fwrite failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(tempPath); os.IsNotExist(err) {
		t.Errorf("Temporary file does not exist at path: %s", tempPath)
	}

	// In a real scenario, the temporary file would be cleaned up when the object is destroyed
	// For testing, let's manually clean it up
	if err := os.Remove(tempPath); err != nil {
		t.Errorf("Failed to clean up temporary file: %v", err)
	}

	// Verify file was removed
	if _, err := os.Stat(tempPath); !os.IsNotExist(err) {
		t.Error("Temporary file should have been removed")
	}
}