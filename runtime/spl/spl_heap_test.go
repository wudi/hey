package spl

import (
	"testing"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

func TestSplHeapFamily(t *testing.T) {
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

	t.Run("SplMaxHeap", func(t *testing.T) {
		testMaxHeap(t, ctx)
	})

	t.Run("SplMinHeap", func(t *testing.T) {
		testMinHeap(t, ctx)
	})

	t.Run("SplPriorityQueue", func(t *testing.T) {
		testPriorityQueue(t, ctx)
	})
}

func testMaxHeap(t *testing.T, ctx *mockContext) {
	// Get the MaxHeap class
	class, err := ctx.registry.GetClass("SplMaxHeap")
	if err != nil {
		t.Fatalf("SplMaxHeap class not found: %v", err)
	}

	// Create MaxHeap instance
	obj := &values.Object{
		ClassName:  "SplMaxHeap",
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
	_, err = constructorImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
	if err != nil {
		t.Fatalf("Constructor failed: %v", err)
	}

	// Test isEmpty on empty heap
	isEmptyMethod := class.Methods["isEmpty"]
	if isEmptyMethod == nil {
		t.Fatal("isEmpty method not found")
	}
	isEmptyImpl := isEmptyMethod.Implementation.(*BuiltinMethodImpl)
	result, err := isEmptyImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
	if err != nil {
		t.Fatalf("isEmpty failed: %v", err)
	}
	if !result.ToBool() {
		t.Error("Expected empty heap to return true")
	}

	// Test count on empty heap
	countMethod := class.Methods["count"]
	if countMethod == nil {
		t.Fatal("count method not found")
	}
	countImpl := countMethod.Implementation.(*BuiltinMethodImpl)
	result, err = countImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
	if err != nil {
		t.Fatalf("count failed: %v", err)
	}
	if result.ToInt() != 0 {
		t.Error("Expected empty heap count to be 0")
	}

	// Test insertions
	insertMethod := class.Methods["insert"]
	if insertMethod == nil {
		t.Fatal("insert method not found")
	}
	insertImpl := insertMethod.Implementation.(*BuiltinMethodImpl)
	values_to_insert := []int64{10, 5, 15, 2, 8}
	for _, val := range values_to_insert {
		_, err := insertImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj, values.NewInt(val)})
		if err != nil {
			t.Fatalf("Insert failed for %d: %v", val, err)
		}
	}

	// Test count after insertions
	result, err = countImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
	if err != nil {
		t.Fatalf("count failed: %v", err)
	}
	if result.ToInt() != 5 {
		t.Errorf("Expected count to be 5, got %d", result.ToInt())
	}

	// Test isEmpty after insertions
	result, err = isEmptyImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
	if err != nil {
		t.Fatalf("isEmpty failed: %v", err)
	}
	if result.ToBool() {
		t.Error("Expected non-empty heap to return false")
	}

	// Test top element (should be max: 15)
	topMethod := class.Methods["top"]
	if topMethod == nil {
		t.Fatal("top method not found")
	}
	topImpl := topMethod.Implementation.(*BuiltinMethodImpl)
	result, err = topImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
	if err != nil {
		t.Fatalf("top failed: %v", err)
	}
	if result.ToInt() != 15 {
		t.Errorf("Expected top element to be 15, got %d", result.ToInt())
	}

	// Test extraction (should be in descending order)
	extractMethod := class.Methods["extract"]
	if extractMethod == nil {
		t.Fatal("extract method not found")
	}
	extractImpl := extractMethod.Implementation.(*BuiltinMethodImpl)
	expectedOrder := []int64{15, 10, 8, 5, 2}
	for i, expected := range expectedOrder {
		result, err := extractImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("Extract failed at position %d: %v", i, err)
		}
		if result.ToInt() != expected {
			t.Errorf("Expected extracted value %d to be %d, got %d", i, expected, result.ToInt())
		}
	}

	// Test isEmpty after extraction
	result, err = isEmptyImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
	if err != nil {
		t.Fatalf("isEmpty failed: %v", err)
	}
	if !result.ToBool() {
		t.Error("Expected heap to be empty after extraction")
	}
}

func testMinHeap(t *testing.T, ctx *mockContext) {
	// Get the MinHeap class
	class, err := ctx.registry.GetClass("SplMinHeap")
	if err != nil {
		t.Fatalf("SplMinHeap class not found: %v", err)
	}

	// Create MinHeap instance
	obj := &values.Object{
		ClassName:  "SplMinHeap",
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
	_, err = constructorImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
	if err != nil {
		t.Fatalf("Constructor failed: %v", err)
	}

	// Test insertions
	insertMethod := class.Methods["insert"]
	if insertMethod == nil {
		t.Fatal("insert method not found")
	}
	insertImpl := insertMethod.Implementation.(*BuiltinMethodImpl)
	values_to_insert := []int64{10, 5, 15, 2, 8}
	for _, val := range values_to_insert {
		_, err := insertImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj, values.NewInt(val)})
		if err != nil {
			t.Fatalf("Insert failed for %d: %v", val, err)
		}
	}

	// Test top element (should be min: 2)
	topMethod := class.Methods["top"]
	if topMethod == nil {
		t.Fatal("top method not found")
	}
	topImpl := topMethod.Implementation.(*BuiltinMethodImpl)
	result, err := topImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
	if err != nil {
		t.Fatalf("top failed: %v", err)
	}
	if result.ToInt() != 2 {
		t.Errorf("Expected top element to be 2, got %d", result.ToInt())
	}

	// Test extraction (should be in ascending order)
	extractMethod := class.Methods["extract"]
	if extractMethod == nil {
		t.Fatal("extract method not found")
	}
	extractImpl := extractMethod.Implementation.(*BuiltinMethodImpl)
	expectedOrder := []int64{2, 5, 8, 10, 15}
	for i, expected := range expectedOrder {
		result, err := extractImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("Extract failed at position %d: %v", i, err)
		}
		if result.ToInt() != expected {
			t.Errorf("Expected extracted value %d to be %d, got %d", i, expected, result.ToInt())
		}
	}
}

func testPriorityQueue(t *testing.T, ctx *mockContext) {
	// Get the PriorityQueue class
	class, err := ctx.registry.GetClass("SplPriorityQueue")
	if err != nil {
		t.Fatalf("SplPriorityQueue class not found: %v", err)
	}

	// Create PriorityQueue instance
	obj := &values.Object{
		ClassName:  "SplPriorityQueue",
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
	_, err = constructorImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
	if err != nil {
		t.Fatalf("Constructor failed: %v", err)
	}

	// Test isEmpty on empty queue
	isEmptyMethod := class.Methods["isEmpty"]
	if isEmptyMethod == nil {
		t.Fatal("isEmpty method not found")
	}
	isEmptyImpl := isEmptyMethod.Implementation.(*BuiltinMethodImpl)
	result, err := isEmptyImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
	if err != nil {
		t.Fatalf("isEmpty failed: %v", err)
	}
	if !result.ToBool() {
		t.Error("Expected empty queue to return true")
	}

	// Test insertions with priorities
	insertMethod := class.Methods["insert"]
	if insertMethod == nil {
		t.Fatal("insert method not found")
	}
	insertImpl := insertMethod.Implementation.(*BuiltinMethodImpl)
	insertions := []struct {
		value    string
		priority int64
	}{
		{"Task A", 3},
		{"Task B", 1},
		{"Task C", 5},
		{"Task D", 2},
		{"Task E", 4},
	}

	for _, insertion := range insertions {
		_, err := insertImpl.GetFunction().Builtin(ctx, []*values.Value{
			thisObj,
			values.NewString(insertion.value),
			values.NewInt(insertion.priority),
		})
		if err != nil {
			t.Fatalf("Insert failed for %s with priority %d: %v", insertion.value, insertion.priority, err)
		}
	}

	// Test count
	countMethod := class.Methods["count"]
	if countMethod == nil {
		t.Fatal("count method not found")
	}
	countImpl := countMethod.Implementation.(*BuiltinMethodImpl)
	result, err = countImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
	if err != nil {
		t.Fatalf("count failed: %v", err)
	}
	if result.ToInt() != 5 {
		t.Errorf("Expected count to be 5, got %d", result.ToInt())
	}

	// Test top element (should be highest priority: "Task C")
	topMethod := class.Methods["top"]
	if topMethod == nil {
		t.Fatal("top method not found")
	}
	topImpl := topMethod.Implementation.(*BuiltinMethodImpl)
	result, err = topImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
	if err != nil {
		t.Fatalf("top failed: %v", err)
	}
	if result.ToString() != "Task C" {
		t.Errorf("Expected top element to be 'Task C', got '%s'", result.ToString())
	}

	// Test extraction (should be in priority order: highest first)
	extractMethod := class.Methods["extract"]
	if extractMethod == nil {
		t.Fatal("extract method not found")
	}
	extractImpl := extractMethod.Implementation.(*BuiltinMethodImpl)
	expectedOrder := []string{"Task C", "Task E", "Task A", "Task D", "Task B"}
	for i, expected := range expectedOrder {
		result, err := extractImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("Extract failed at position %d: %v", i, err)
		}
		if result.ToString() != expected {
			t.Errorf("Expected extracted value %d to be %s, got %s", i, expected, result.ToString())
		}
	}

	// Test isEmpty after extraction
	result, err = isEmptyImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
	if err != nil {
		t.Fatalf("isEmpty failed: %v", err)
	}
	if !result.ToBool() {
		t.Error("Expected queue to be empty after extraction")
	}
}

func TestSplHeapEdgeCases(t *testing.T) {
	registry.Initialize()

	// Manually register the SPL classes for testing
	for _, class := range GetSplClasses() {
		err := registry.GlobalRegistry.RegisterClass(class)
		if err != nil {
			t.Fatalf("Failed to register class %s: %v", class.Name, err)
		}
	}

	ctx := &mockContext{registry: registry.GlobalRegistry}

	t.Run("EmptyHeapOperations", func(t *testing.T) {
		// Get the MaxHeap class
		class, err := ctx.registry.GetClass("SplMaxHeap")
		if err != nil {
			t.Fatalf("SplMaxHeap class not found: %v", err)
		}

		// Create empty MaxHeap
		obj := &values.Object{
			ClassName:  "SplMaxHeap",
			Properties: make(map[string]*values.Value),
		}
		thisObj := &values.Value{
			Type: values.TypeObject,
			Data: obj,
		}

		// Test top() on empty heap - should fail
		topMethod := class.Methods["top"]
		if topMethod == nil {
			t.Fatal("top method not found")
		}
		topImpl := topMethod.Implementation.(*BuiltinMethodImpl)
		_, err = topImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err == nil {
			t.Error("Expected error when calling top() on empty heap")
		}

		// Test extract() on empty heap - should fail
		extractMethod := class.Methods["extract"]
		if extractMethod == nil {
			t.Fatal("extract method not found")
		}
		extractImpl := extractMethod.Implementation.(*BuiltinMethodImpl)
		_, err = extractImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err == nil {
			t.Error("Expected error when calling extract() on empty heap")
		}
	})

	t.Run("DuplicateValues", func(t *testing.T) {
		// Get the MaxHeap class
		class, err := ctx.registry.GetClass("SplMaxHeap")
		if err != nil {
			t.Fatalf("SplMaxHeap class not found: %v", err)
		}

		// Create MaxHeap
		obj := &values.Object{
			ClassName:  "SplMaxHeap",
			Properties: make(map[string]*values.Value),
		}
		thisObj := &values.Value{
			Type: values.TypeObject,
			Data: obj,
		}

		// Insert duplicate values
		insertMethod := class.Methods["insert"]
		if insertMethod == nil {
			t.Fatal("insert method not found")
		}
		insertImpl := insertMethod.Implementation.(*BuiltinMethodImpl)
		for i := 0; i < 3; i++ {
			_, err := insertImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj, values.NewInt(5)})
			if err != nil {
				t.Fatalf("Insert failed: %v", err)
			}
		}

		// Test count
		countMethod := class.Methods["count"]
		if countMethod == nil {
			t.Fatal("count method not found")
		}
		countImpl := countMethod.Implementation.(*BuiltinMethodImpl)
		result, err := countImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("count failed: %v", err)
		}
		if result.ToInt() != 3 {
			t.Errorf("Expected count to be 3, got %d", result.ToInt())
		}

		// Extract all - should all be 5
		extractMethod := class.Methods["extract"]
		if extractMethod == nil {
			t.Fatal("extract method not found")
		}
		extractImpl := extractMethod.Implementation.(*BuiltinMethodImpl)
		for i := 0; i < 3; i++ {
			result, err := extractImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
			if err != nil {
				t.Fatalf("Extract failed: %v", err)
			}
			if result.ToInt() != 5 {
				t.Errorf("Expected extracted value to be 5, got %d", result.ToInt())
			}
		}
	})
}