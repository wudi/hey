package spl

import (
	"testing"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

func TestSplObjectStorage(t *testing.T) {
	// Initialize registry
	registry.Initialize()
	ctx := &mockContext{registry: registry.GlobalRegistry}

	// Get the SplObjectStorage class
	class := GetSplObjectStorageClass()
	if class == nil {
		t.Fatal("SplObjectStorage class is nil")
	}

	// Create a new SplObjectStorage instance
	obj := &values.Object{
		ClassName:  "SplObjectStorage",
		Properties: make(map[string]*values.Value),
	}
	thisObj := &values.Value{
		Type: values.TypeObject,
		Data: obj,
	}

	// Create test objects
	testObj1 := &values.Object{
		ClassName:  "TestObject",
		Properties: map[string]*values.Value{"value": values.NewString("obj1")},
	}
	objVal1 := &values.Value{
		Type: values.TypeObject,
		Data: testObj1,
	}

	testObj2 := &values.Object{
		ClassName:  "TestObject",
		Properties: map[string]*values.Value{"value": values.NewString("obj2")},
	}
	objVal2 := &values.Value{
		Type: values.TypeObject,
		Data: testObj2,
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
		if obj.Properties["__storage"] == nil {
			t.Fatal("Internal storage not set")
		}
	})

	t.Run("AttachAndContains", func(t *testing.T) {
		// Test attach
		attachMethod := class.Methods["attach"]
		if attachMethod == nil {
			t.Fatal("attach method not found")
		}

		attachImpl := attachMethod.Implementation.(*BuiltinMethodImpl)
		args := []*values.Value{thisObj, objVal1, values.NewString("data1")}
		_, err := attachImpl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("attach failed: %v", err)
		}

		// Test contains
		containsMethod := class.Methods["contains"]
		if containsMethod == nil {
			t.Fatal("contains method not found")
		}

		containsImpl := containsMethod.Implementation.(*BuiltinMethodImpl)
		args = []*values.Value{thisObj, objVal1}
		result, err := containsImpl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("contains failed: %v", err)
		}

		if !result.ToBool() {
			t.Fatal("Expected contains to return true for attached object")
		}

		// Test contains for non-attached object
		args = []*values.Value{thisObj, objVal2}
		result, _ = containsImpl.GetFunction().Builtin(ctx, args)
		if result.ToBool() {
			t.Fatal("Expected contains to return false for non-attached object")
		}
	})

	t.Run("Count", func(t *testing.T) {
		countMethod := class.Methods["count"]
		if countMethod == nil {
			t.Fatal("count method not found")
		}

		countImpl := countMethod.Implementation.(*BuiltinMethodImpl)
		args := []*values.Value{thisObj}
		result, err := countImpl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("count failed: %v", err)
		}

		if result.ToInt() != 1 {
			t.Fatalf("Expected count 1, got %d", result.ToInt())
		}
	})

	t.Run("AttachMultiple", func(t *testing.T) {
		// Attach second object
		attachMethod := class.Methods["attach"]
		attachImpl := attachMethod.Implementation.(*BuiltinMethodImpl)

		args := []*values.Value{thisObj, objVal2, values.NewString("data2")}
		_, _ = attachImpl.GetFunction().Builtin(ctx, args)

		// Check count
		countMethod := class.Methods["count"]
		countImpl := countMethod.Implementation.(*BuiltinMethodImpl)
		args = []*values.Value{thisObj}
		result, _ := countImpl.GetFunction().Builtin(ctx, args)
		if result.ToInt() != 2 {
			t.Fatalf("Expected count 2 after attaching second object, got %d", result.ToInt())
		}
	})

	t.Run("GetInfoAndSetInfo", func(t *testing.T) {
		// Test getInfo
		getInfoMethod := class.Methods["getInfo"]
		if getInfoMethod == nil {
			t.Fatal("getInfo method not found")
		}

		// First, set current to objVal1
		args := []*values.Value{thisObj, objVal1}
		// Simulate setting current position by calling contains which should set it
		containsMethod := class.Methods["contains"]
		containsImpl := containsMethod.Implementation.(*BuiltinMethodImpl)
		containsImpl.GetFunction().Builtin(ctx, args)

		getInfoImpl := getInfoMethod.Implementation.(*BuiltinMethodImpl)
		args = []*values.Value{thisObj}
		result, err := getInfoImpl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("getInfo failed: %v", err)
		}

		if result.ToString() != "data1" {
			t.Fatalf("Expected 'data1', got '%s'", result.ToString())
		}

		// Test setInfo
		setInfoMethod := class.Methods["setInfo"]
		if setInfoMethod == nil {
			t.Fatal("setInfo method not found")
		}

		setInfoImpl := setInfoMethod.Implementation.(*BuiltinMethodImpl)
		args = []*values.Value{thisObj, values.NewString("updated_data1")}
		_, err = setInfoImpl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("setInfo failed: %v", err)
		}

		// Check updated info
		args = []*values.Value{thisObj}
		result, _ = getInfoImpl.GetFunction().Builtin(ctx, args)
		if result.ToString() != "updated_data1" {
			t.Fatalf("Expected 'updated_data1', got '%s'", result.ToString())
		}
	})

	t.Run("Detach", func(t *testing.T) {
		detachMethod := class.Methods["detach"]
		if detachMethod == nil {
			t.Fatal("detach method not found")
		}

		detachImpl := detachMethod.Implementation.(*BuiltinMethodImpl)
		args := []*values.Value{thisObj, objVal1}
		_, err := detachImpl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("detach failed: %v", err)
		}

		// Check contains after detach
		containsMethod := class.Methods["contains"]
		containsImpl := containsMethod.Implementation.(*BuiltinMethodImpl)
		args = []*values.Value{thisObj, objVal1}
		result, _ := containsImpl.GetFunction().Builtin(ctx, args)
		if result.ToBool() {
			t.Fatal("Expected contains to return false after detach")
		}

		// Check count after detach
		countMethod := class.Methods["count"]
		countImpl := countMethod.Implementation.(*BuiltinMethodImpl)
		args = []*values.Value{thisObj}
		result, _ = countImpl.GetFunction().Builtin(ctx, args)
		if result.ToInt() != 1 {
			t.Fatalf("Expected count 1 after detach, got %d", result.ToInt())
		}
	})

	t.Run("Iterator", func(t *testing.T) {
		// Test rewind
		rewindMethod := class.Methods["rewind"]
		if rewindMethod == nil {
			t.Fatal("rewind method not found")
		}

		rewindImpl := rewindMethod.Implementation.(*BuiltinMethodImpl)
		args := []*values.Value{thisObj}
		_, err := rewindImpl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("rewind failed: %v", err)
		}

		// Test valid
		validMethod := class.Methods["valid"]
		if validMethod == nil {
			t.Fatal("valid method not found")
		}

		validImpl := validMethod.Implementation.(*BuiltinMethodImpl)
		args = []*values.Value{thisObj}
		result, _ := validImpl.GetFunction().Builtin(ctx, args)
		if !result.ToBool() {
			t.Fatal("Expected valid to return true after rewind")
		}

		// Test current
		currentMethod := class.Methods["current"]
		if currentMethod == nil {
			t.Fatal("current method not found")
		}

		currentImpl := currentMethod.Implementation.(*BuiltinMethodImpl)
		args = []*values.Value{thisObj}
		result, _ = currentImpl.GetFunction().Builtin(ctx, args)
		if !result.IsObject() {
			t.Fatal("Expected current to return an object")
		}

		// Test next
		nextMethod := class.Methods["next"]
		if nextMethod == nil {
			t.Fatal("next method not found")
		}

		nextImpl := nextMethod.Implementation.(*BuiltinMethodImpl)
		args = []*values.Value{thisObj}
		_, _ = nextImpl.GetFunction().Builtin(ctx, args)

		// After next, should be at end (only one object left)
		args = []*values.Value{thisObj}
		result, _ = validImpl.GetFunction().Builtin(ctx, args)
		if result.ToBool() {
			t.Fatal("Expected valid to return false after next (at end)")
		}
	})
}