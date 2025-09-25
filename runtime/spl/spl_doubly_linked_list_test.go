package spl

import (
	"testing"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

func TestSplDoublyLinkedList(t *testing.T) {
	// Initialize registry
	registry.Initialize()
	ctx := &mockContext{registry: registry.GlobalRegistry}

	// Get the SplDoublyLinkedList class
	class := GetSplDoublyLinkedListClass()
	if class == nil {
		t.Fatal("SplDoublyLinkedList class is nil")
	}

	// Create a new SplDoublyLinkedList instance
	obj := &values.Object{
		ClassName:  "SplDoublyLinkedList",
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

		args := []*values.Value{thisObj}
		impl := constructor.Implementation.(*BuiltinMethodImpl)
		_, err := impl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("Constructor failed: %v", err)
		}

		// Check if internal properties are set
		if obj.Properties["__count"] == nil {
			t.Fatal("Internal count not set")
		}
	})

	t.Run("IsEmpty", func(t *testing.T) {
		isEmptyMethod := class.Methods["isEmpty"]
		if isEmptyMethod == nil {
			t.Fatal("isEmpty method not found")
		}

		args := []*values.Value{thisObj}
		impl := isEmptyMethod.Implementation.(*BuiltinMethodImpl)
		result, err := impl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("isEmpty failed: %v", err)
		}

		if !result.ToBool() {
			t.Fatal("Expected isEmpty to return true for empty list")
		}
	})

	t.Run("Push", func(t *testing.T) {
		pushMethod := class.Methods["push"]
		if pushMethod == nil {
			t.Fatal("push method not found")
		}

		// Push some values
		impl := pushMethod.Implementation.(*BuiltinMethodImpl)
		args := []*values.Value{thisObj, values.NewInt(1)}
		_, err := impl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("push failed: %v", err)
		}

		args = []*values.Value{thisObj, values.NewInt(2)}
		_, err = impl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("push failed: %v", err)
		}

		args = []*values.Value{thisObj, values.NewInt(3)}
		_, err = impl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("push failed: %v", err)
		}

		// Check count
		countMethod := class.Methods["count"]
		countImpl := countMethod.Implementation.(*BuiltinMethodImpl)
		args = []*values.Value{thisObj}
		result, _ := countImpl.GetFunction().Builtin(ctx, args)
		if result.ToInt() != 3 {
			t.Fatalf("Expected count 3, got %d", result.ToInt())
		}
	})

	t.Run("Top", func(t *testing.T) {
		topMethod := class.Methods["top"]
		if topMethod == nil {
			t.Fatal("top method not found")
		}

		args := []*values.Value{thisObj}
		impl := topMethod.Implementation.(*BuiltinMethodImpl)
		result, err := impl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("top failed: %v", err)
		}

		if result.ToInt() != 3 {
			t.Fatalf("Expected top to be 3, got %d", result.ToInt())
		}
	})

	t.Run("Bottom", func(t *testing.T) {
		bottomMethod := class.Methods["bottom"]
		if bottomMethod == nil {
			t.Fatal("bottom method not found")
		}

		args := []*values.Value{thisObj}
		impl := bottomMethod.Implementation.(*BuiltinMethodImpl)
		result, err := impl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("bottom failed: %v", err)
		}

		if result.ToInt() != 1 {
			t.Fatalf("Expected bottom to be 1, got %d", result.ToInt())
		}
	})

	t.Run("Pop", func(t *testing.T) {
		popMethod := class.Methods["pop"]
		if popMethod == nil {
			t.Fatal("pop method not found")
		}

		args := []*values.Value{thisObj}
		impl := popMethod.Implementation.(*BuiltinMethodImpl)
		result, err := impl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("pop failed: %v", err)
		}

		if result.ToInt() != 3 {
			t.Fatalf("Expected pop to return 3, got %d", result.ToInt())
		}

		// Check count
		countMethod := class.Methods["count"]
		countImpl := countMethod.Implementation.(*BuiltinMethodImpl)
		args = []*values.Value{thisObj}
		countResult, _ := countImpl.GetFunction().Builtin(ctx, args)
		if countResult.ToInt() != 2 {
			t.Fatalf("Expected count 2 after pop, got %d", countResult.ToInt())
		}
	})

	t.Run("Unshift", func(t *testing.T) {
		unshiftMethod := class.Methods["unshift"]
		if unshiftMethod == nil {
			t.Fatal("unshift method not found")
		}

		args := []*values.Value{thisObj, values.NewInt(0)}
		impl := unshiftMethod.Implementation.(*BuiltinMethodImpl)
		_, err := impl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("unshift failed: %v", err)
		}

		// Check bottom now
		bottomMethod := class.Methods["bottom"]
		bottomImpl := bottomMethod.Implementation.(*BuiltinMethodImpl)
		args = []*values.Value{thisObj}
		result, _ := bottomImpl.GetFunction().Builtin(ctx, args)
		if result.ToInt() != 0 {
			t.Fatalf("Expected bottom to be 0 after unshift, got %d", result.ToInt())
		}

		// Check count
		countMethod := class.Methods["count"]
		countImpl := countMethod.Implementation.(*BuiltinMethodImpl)
		args = []*values.Value{thisObj}
		countResult, _ := countImpl.GetFunction().Builtin(ctx, args)
		if countResult.ToInt() != 3 {
			t.Fatalf("Expected count 3 after unshift, got %d", countResult.ToInt())
		}
	})

	t.Run("Shift", func(t *testing.T) {
		shiftMethod := class.Methods["shift"]
		if shiftMethod == nil {
			t.Fatal("shift method not found")
		}

		args := []*values.Value{thisObj}
		impl := shiftMethod.Implementation.(*BuiltinMethodImpl)
		result, err := impl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("shift failed: %v", err)
		}

		if result.ToInt() != 0 {
			t.Fatalf("Expected shift to return 0, got %d", result.ToInt())
		}

		// Check bottom now
		bottomMethod := class.Methods["bottom"]
		bottomImpl := bottomMethod.Implementation.(*BuiltinMethodImpl)
		args = []*values.Value{thisObj}
		bottomResult, _ := bottomImpl.GetFunction().Builtin(ctx, args)
		if bottomResult.ToInt() != 1 {
			t.Fatalf("Expected bottom to be 1 after shift, got %d", bottomResult.ToInt())
		}

		// Check count
		countMethod := class.Methods["count"]
		countImpl := countMethod.Implementation.(*BuiltinMethodImpl)
		args = []*values.Value{thisObj}
		countResult, _ := countImpl.GetFunction().Builtin(ctx, args)
		if countResult.ToInt() != 2 {
			t.Fatalf("Expected count 2 after shift, got %d", countResult.ToInt())
		}
	})

	t.Run("EmptyList", func(t *testing.T) {
		// Pop remaining items
		popMethod := class.Methods["pop"]
		popImpl := popMethod.Implementation.(*BuiltinMethodImpl)

		args := []*values.Value{thisObj}
		popImpl.GetFunction().Builtin(ctx, args) // Remove 2
		popImpl.GetFunction().Builtin(ctx, args) // Remove 1

		// Should be empty now
		isEmptyMethod := class.Methods["isEmpty"]
		isEmptyImpl := isEmptyMethod.Implementation.(*BuiltinMethodImpl)
		args = []*values.Value{thisObj}
		result, _ := isEmptyImpl.GetFunction().Builtin(ctx, args)
		if !result.ToBool() {
			t.Fatal("Expected isEmpty to return true after removing all items")
		}
	})
}