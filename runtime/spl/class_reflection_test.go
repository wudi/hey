package spl

import (
	"testing"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

func TestClassReflectionFunctions(t *testing.T) {
	// Initialize registry
	registry.Initialize()
	ctx := &mockContext{registry: registry.GlobalRegistry}

	// Register SPL classes for testing
	for _, class := range GetSplClasses() {
		if err := registry.GlobalRegistry.RegisterClass(class); err != nil {
			t.Fatalf("Failed to register class %s: %v", class.Name, err)
		}
	}

	// Register SPL interfaces for testing
	for _, iface := range GetSplInterfaces() {
		if err := registry.GlobalRegistry.RegisterInterface(iface); err != nil {
			t.Fatalf("Failed to register interface %s: %v", iface.Name, err)
		}
	}

	t.Run("class_implements", func(t *testing.T) {
		function := getClassImplementsFunction()

		// Test with ArrayIterator class name
		args := []*values.Value{values.NewString("ArrayIterator")}
		result, err := function.Builtin(ctx, args)
		if err != nil {
			t.Fatalf("class_implements failed: %v", err)
		}

		if !result.IsArray() {
			t.Fatal("Expected array result")
		}

		// Check that Iterator interface is present
		arr := result.Data.(*values.Array)
		found := false
		for _, v := range arr.Elements {
			if v.IsString() && v.ToString() == "Iterator" {
				found = true
				break
			}
		}
		if !found {
			t.Fatal("Expected Iterator interface to be present")
		}
	})

	t.Run("class_implements_with_object", func(t *testing.T) {
		function := getClassImplementsFunction()

		// Create ArrayIterator instance
		obj := &values.Object{
			ClassName:  "ArrayIterator",
			Properties: make(map[string]*values.Value),
		}
		instance := &values.Value{
			Type: values.TypeObject,
			Data: obj,
		}

		args := []*values.Value{instance}
		result, err := function.Builtin(ctx, args)
		if err != nil {
			t.Fatalf("class_implements with object failed: %v", err)
		}

		if !result.IsArray() {
			t.Fatal("Expected array result")
		}
	})

	t.Run("class_implements_nonexistent", func(t *testing.T) {
		function := getClassImplementsFunction()

		args := []*values.Value{values.NewString("NonExistentClass")}
		result, err := function.Builtin(ctx, args)
		if err != nil {
			t.Fatalf("class_implements with non-existent class failed: %v", err)
		}

		if !result.IsBool() || result.ToBool() {
			t.Fatal("Expected false for non-existent class")
		}
	})

	t.Run("class_parents", func(t *testing.T) {
		function := getClassParentsFunction()

		// Test with ArrayIterator (should have no parents)
		args := []*values.Value{values.NewString("ArrayIterator")}
		result, err := function.Builtin(ctx, args)
		if err != nil {
			t.Fatalf("class_parents failed: %v", err)
		}

		if !result.IsArray() {
			t.Fatal("Expected array result")
		}

		// Should be empty array
		arr := result.Data.(*values.Array)
		if len(arr.Elements) != 0 {
			t.Fatal("Expected empty array for ArrayIterator parents")
		}
	})

	t.Run("class_parents_nonexistent", func(t *testing.T) {
		function := getClassParentsFunction()

		args := []*values.Value{values.NewString("NonExistentClass")}
		result, err := function.Builtin(ctx, args)
		if err != nil {
			t.Fatalf("class_parents with non-existent class failed: %v", err)
		}

		if !result.IsBool() || result.ToBool() {
			t.Fatal("Expected false for non-existent class")
		}
	})

	t.Run("class_uses", func(t *testing.T) {
		function := getClassUsesFunction()

		// Test with ArrayIterator (should have no traits)
		args := []*values.Value{values.NewString("ArrayIterator")}
		result, err := function.Builtin(ctx, args)
		if err != nil {
			t.Fatalf("class_uses failed: %v", err)
		}

		if !result.IsArray() {
			t.Fatal("Expected array result")
		}

		// Should be empty array
		arr := result.Data.(*values.Array)
		if len(arr.Elements) != 0 {
			t.Fatal("Expected empty array for ArrayIterator traits")
		}
	})

	t.Run("class_uses_nonexistent", func(t *testing.T) {
		function := getClassUsesFunction()

		args := []*values.Value{values.NewString("NonExistentClass")}
		result, err := function.Builtin(ctx, args)
		if err != nil {
			t.Fatalf("class_uses with non-existent class failed: %v", err)
		}

		if !result.IsBool() || result.ToBool() {
			t.Fatal("Expected false for non-existent class")
		}
	})
}