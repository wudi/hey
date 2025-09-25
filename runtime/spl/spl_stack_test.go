package spl

import (
	"testing"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

func TestSplStack(t *testing.T) {
	// Initialize registry
	registry.Initialize()
	ctx := &mockContext{registry: registry.GlobalRegistry}

	// Get the SplStack class
	class := GetSplStackClass()
	if class == nil {
		t.Fatal("SplStack class is nil")
	}

	// Create a new SplStack instance
	obj := &values.Object{
		ClassName:  "SplStack",
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

	t.Run("StackBehavior", func(t *testing.T) {
		// Push items (uses push from parent)
		pushMethod := class.Methods["push"]
		pushImpl := pushMethod.Implementation.(*BuiltinMethodImpl)

		args := []*values.Value{thisObj, values.NewString("first")}
		_, _ = pushImpl.GetFunction().Builtin(ctx, args)

		args = []*values.Value{thisObj, values.NewString("second")}
		_, _ = pushImpl.GetFunction().Builtin(ctx, args)

		args = []*values.Value{thisObj, values.NewString("third")}
		_, _ = pushImpl.GetFunction().Builtin(ctx, args)

		// Check count
		countMethod := class.Methods["count"]
		countImpl := countMethod.Implementation.(*BuiltinMethodImpl)
		args = []*values.Value{thisObj}
		result, _ := countImpl.GetFunction().Builtin(ctx, args)
		if result.ToInt() != 3 {
			t.Fatalf("Expected count 3, got %d", result.ToInt())
		}

		// Check top (should be last pushed - "third")
		topMethod := class.Methods["top"]
		topImpl := topMethod.Implementation.(*BuiltinMethodImpl)
		args = []*values.Value{thisObj}
		result, _ = topImpl.GetFunction().Builtin(ctx, args)
		if result.ToString() != "third" {
			t.Fatalf("Expected top 'third', got '%s'", result.ToString())
		}

		// Pop items (LIFO - Last In First Out)
		popMethod := class.Methods["pop"]
		popImpl := popMethod.Implementation.(*BuiltinMethodImpl)

		args = []*values.Value{thisObj}
		result, _ = popImpl.GetFunction().Builtin(ctx, args)
		if result.ToString() != "third" {
			t.Fatalf("Expected pop 'third', got '%s'", result.ToString())
		}

		result, _ = popImpl.GetFunction().Builtin(ctx, args)
		if result.ToString() != "second" {
			t.Fatalf("Expected pop 'second', got '%s'", result.ToString())
		}

		result, _ = popImpl.GetFunction().Builtin(ctx, args)
		if result.ToString() != "first" {
			t.Fatalf("Expected pop 'first', got '%s'", result.ToString())
		}

		// Check empty
		isEmptyMethod := class.Methods["isEmpty"]
		isEmptyImpl := isEmptyMethod.Implementation.(*BuiltinMethodImpl)
		args = []*values.Value{thisObj}
		result, _ = isEmptyImpl.GetFunction().Builtin(ctx, args)
		if !result.ToBool() {
			t.Fatal("Expected isEmpty to return true after popping all items")
		}
	})
}

func TestSplQueue(t *testing.T) {
	// Initialize registry
	registry.Initialize()
	ctx := &mockContext{registry: registry.GlobalRegistry}

	// Get the SplQueue class
	class := GetSplQueueClass()
	if class == nil {
		t.Fatal("SplQueue class is nil")
	}

	// Create a new SplQueue instance
	obj := &values.Object{
		ClassName:  "SplQueue",
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

	t.Run("QueueBehavior", func(t *testing.T) {
		// Enqueue items
		enqueueMethod := class.Methods["enqueue"]
		if enqueueMethod == nil {
			t.Fatal("enqueue method not found")
		}
		enqueueImpl := enqueueMethod.Implementation.(*BuiltinMethodImpl)

		args := []*values.Value{thisObj, values.NewString("first")}
		_, _ = enqueueImpl.GetFunction().Builtin(ctx, args)

		args = []*values.Value{thisObj, values.NewString("second")}
		_, _ = enqueueImpl.GetFunction().Builtin(ctx, args)

		args = []*values.Value{thisObj, values.NewString("third")}
		_, _ = enqueueImpl.GetFunction().Builtin(ctx, args)

		// Check count
		countMethod := class.Methods["count"]
		countImpl := countMethod.Implementation.(*BuiltinMethodImpl)
		args = []*values.Value{thisObj}
		result, _ := countImpl.GetFunction().Builtin(ctx, args)
		if result.ToInt() != 3 {
			t.Fatalf("Expected count 3, got %d", result.ToInt())
		}

		// Dequeue items (FIFO - First In First Out)
		dequeueMethod := class.Methods["dequeue"]
		if dequeueMethod == nil {
			t.Fatal("dequeue method not found")
		}
		dequeueImpl := dequeueMethod.Implementation.(*BuiltinMethodImpl)

		args = []*values.Value{thisObj}
		result, _ = dequeueImpl.GetFunction().Builtin(ctx, args)
		if result.ToString() != "first" {
			t.Fatalf("Expected dequeue 'first', got '%s'", result.ToString())
		}

		result, _ = dequeueImpl.GetFunction().Builtin(ctx, args)
		if result.ToString() != "second" {
			t.Fatalf("Expected dequeue 'second', got '%s'", result.ToString())
		}

		result, _ = dequeueImpl.GetFunction().Builtin(ctx, args)
		if result.ToString() != "third" {
			t.Fatalf("Expected dequeue 'third', got '%s'", result.ToString())
		}

		// Check empty
		isEmptyMethod := class.Methods["isEmpty"]
		isEmptyImpl := isEmptyMethod.Implementation.(*BuiltinMethodImpl)
		args = []*values.Value{thisObj}
		result, _ = isEmptyImpl.GetFunction().Builtin(ctx, args)
		if !result.ToBool() {
			t.Fatal("Expected isEmpty to return true after dequeuing all items")
		}
	})
}