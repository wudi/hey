package spl

import (
	"strings"
	"testing"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

func TestEmptyIterator(t *testing.T) {
	// Initialize registry
	registry.Initialize()
	ctx := &mockContext{registry: registry.GlobalRegistry}

	// Get the EmptyIterator class
	class := GetEmptyIteratorClass()
	if class == nil {
		t.Fatal("EmptyIterator class is nil")
	}

	// Create a new EmptyIterator instance
	obj := &values.Object{
		ClassName:  "EmptyIterator",
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
	})

	t.Run("Current", func(t *testing.T) {
		currentMethod := class.Methods["current"]
		if currentMethod == nil {
			t.Fatal("current method not found")
		}

		args := []*values.Value{thisObj}
		impl := currentMethod.Implementation.(*BuiltinMethodImpl)
		_, err := impl.GetFunction().Builtin(ctx, args)
		if err == nil {
			t.Fatal("Expected current() to throw BadMethodCallException")
		}

		// Check that it's a BadMethodCallException
		if !strings.Contains(err.Error(), "BadMethodCallException") {
			t.Fatalf("Expected BadMethodCallException, got: %v", err)
		}
	})

	t.Run("Key", func(t *testing.T) {
		keyMethod := class.Methods["key"]
		if keyMethod == nil {
			t.Fatal("key method not found")
		}

		args := []*values.Value{thisObj}
		impl := keyMethod.Implementation.(*BuiltinMethodImpl)
		_, err := impl.GetFunction().Builtin(ctx, args)
		if err == nil {
			t.Fatal("Expected key() to throw BadMethodCallException")
		}

		// Check that it's a BadMethodCallException
		if !strings.Contains(err.Error(), "BadMethodCallException") {
			t.Fatalf("Expected BadMethodCallException, got: %v", err)
		}
	})

	t.Run("Valid", func(t *testing.T) {
		validMethod := class.Methods["valid"]
		if validMethod == nil {
			t.Fatal("valid method not found")
		}

		args := []*values.Value{thisObj}
		impl := validMethod.Implementation.(*BuiltinMethodImpl)
		result, err := impl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("valid failed: %v", err)
		}

		if result.ToBool() {
			t.Fatal("Expected valid() to return false for empty iterator")
		}
	})

	t.Run("Next", func(t *testing.T) {
		nextMethod := class.Methods["next"]
		if nextMethod == nil {
			t.Fatal("next method not found")
		}

		args := []*values.Value{thisObj}
		impl := nextMethod.Implementation.(*BuiltinMethodImpl)
		_, err := impl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("next failed: %v", err)
		}

		// After next, should still be invalid
		validMethod := class.Methods["valid"]
		validImpl := validMethod.Implementation.(*BuiltinMethodImpl)
		args = []*values.Value{thisObj}
		result, _ := validImpl.GetFunction().Builtin(ctx, args)
		if result.ToBool() {
			t.Fatal("Expected valid() to return false after next()")
		}
	})

	t.Run("Rewind", func(t *testing.T) {
		rewindMethod := class.Methods["rewind"]
		if rewindMethod == nil {
			t.Fatal("rewind method not found")
		}

		args := []*values.Value{thisObj}
		impl := rewindMethod.Implementation.(*BuiltinMethodImpl)
		_, err := impl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("rewind failed: %v", err)
		}

		// After rewind, should still be invalid
		validMethod := class.Methods["valid"]
		validImpl := validMethod.Implementation.(*BuiltinMethodImpl)
		args = []*values.Value{thisObj}
		result, _ := validImpl.GetFunction().Builtin(ctx, args)
		if result.ToBool() {
			t.Fatal("Expected valid() to return false after rewind()")
		}
	})

	t.Run("Iteration", func(t *testing.T) {
		// Test that iteration over EmptyIterator produces no elements
		rewindMethod := class.Methods["rewind"]
		validMethod := class.Methods["valid"]
		nextMethod := class.Methods["next"]
		currentMethod := class.Methods["current"]

		rewindImpl := rewindMethod.Implementation.(*BuiltinMethodImpl)
		validImpl := validMethod.Implementation.(*BuiltinMethodImpl)
		nextImpl := nextMethod.Implementation.(*BuiltinMethodImpl)
		currentImpl := currentMethod.Implementation.(*BuiltinMethodImpl)

		// Start iteration
		args := []*values.Value{thisObj}
		rewindImpl.GetFunction().Builtin(ctx, args)

		count := 0
		for {
			result, _ := validImpl.GetFunction().Builtin(ctx, args)
			if !result.ToBool() {
				break
			}
			// This should never execute since the iterator is always invalid
			currentImpl.GetFunction().Builtin(ctx, args)
			nextImpl.GetFunction().Builtin(ctx, args)
			count++
			if count > 10 { // Safety check
				break
			}
		}

		if count != 0 {
			t.Fatalf("Expected 0 iterations, got %d", count)
		}
	})
}