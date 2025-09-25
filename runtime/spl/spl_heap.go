package spl

import (
	"fmt"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// SplHeap represents the abstract SplHeap class
type SplHeap struct {
	elements []*values.Value
	position int
}

// GetSplHeapClass returns the SplHeap class descriptor
func GetSplHeapClass() *registry.ClassDescriptor {
	// Constructor
	constructorImpl := &registry.Function{
		Name:      "__construct",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// count() method
	countImpl := &registry.Function{
		Name:      "count",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			heap, err := getSplHeapFromObject(thisObj)
			if err != nil {
				return nil, err
			}
			return values.NewInt(int64(len(heap.elements))), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// current() method
	currentImpl := &registry.Function{
		Name:      "current",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			heap, err := getSplHeapFromObject(thisObj)
			if err != nil {
				return nil, err
			}
			if heap.position >= 0 && heap.position < len(heap.elements) {
				return heap.elements[heap.position], nil
			}
			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// extract() method
	extractImpl := &registry.Function{
		Name:      "extract",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			heap, err := getSplHeapFromObject(thisObj)
			if err != nil {
				return nil, err
			}

			if len(heap.elements) == 0 {
				return nil, fmt.Errorf("can't extract from an empty heap")
			}

			// Get the root element
			root := heap.elements[0]

			// Move the last element to the root and remove it
			lastIndex := len(heap.elements) - 1
			heap.elements[0] = heap.elements[lastIndex]
			heap.elements = heap.elements[:lastIndex]

			// Get the compare function based on the class type
			objData := thisObj.Data.(*values.Object)
			var compareFn func(a, b *values.Value) int

			switch objData.ClassName {
			case "SplMaxHeap":
				compareFn = func(a, b *values.Value) int {
					return compareValues(a, b)
				}
			case "SplMinHeap":
				compareFn = func(a, b *values.Value) int {
					return -compareValues(a, b) // Reverse comparison for min heap
				}
			default:
				return nil, fmt.Errorf("unknown heap class: %s", objData.ClassName)
			}

			// Restore heap property
			if len(heap.elements) > 0 {
				heap.heapifyDown(0, compareFn)
			}

			// Reset iterator position
			heap.position = 0

			return root, nil
		},
		Parameters: []*registry.Parameter{},
	}

	// insert() method
	insertImpl := &registry.Function{
		Name:      "insert",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 2 {
				return nil, fmt.Errorf("SplHeap::insert() expects at least 1 parameter, %d given", len(args)-1)
			}

			thisObj := args[0]
			value := args[1]

			heap, err := getSplHeapFromObject(thisObj)
			if err != nil {
				return nil, err
			}

			// Add the value to the end
			heap.elements = append(heap.elements, value)

			// Get the compare function based on the class type
			objData := thisObj.Data.(*values.Object)
			var compareFn func(a, b *values.Value) int

			switch objData.ClassName {
			case "SplMaxHeap":
				compareFn = func(a, b *values.Value) int {
					return compareValues(a, b)
				}
			case "SplMinHeap":
				compareFn = func(a, b *values.Value) int {
					return -compareValues(a, b) // Reverse comparison for min heap
				}
			default:
				return nil, fmt.Errorf("unknown heap class: %s", objData.ClassName)
			}

			// Restore heap property
			heap.heapifyUp(len(heap.elements)-1, compareFn)

			return values.NewBool(true), nil
		},
		Parameters: []*registry.Parameter{
			{Name: "value", Type: "mixed"},
		},
	}

	// isEmpty() method
	isEmptyImpl := &registry.Function{
		Name:      "isEmpty",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			heap, err := getSplHeapFromObject(thisObj)
			if err != nil {
				return nil, err
			}
			return values.NewBool(len(heap.elements) == 0), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// key() method
	keyImpl := &registry.Function{
		Name:      "key",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			heap, err := getSplHeapFromObject(thisObj)
			if err != nil {
				return nil, err
			}
			return values.NewInt(int64(heap.position)), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// next() method
	nextImpl := &registry.Function{
		Name:      "next",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			heap, err := getSplHeapFromObject(thisObj)
			if err != nil {
				return nil, err
			}
			heap.position++
			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// recoverFromCorruption() method
	recoverFromCorruptionImpl := &registry.Function{
		Name:      "recoverFromCorruption",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			// For now, just return true as we don't track corruption
			return values.NewBool(true), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// rewind() method
	rewindImpl := &registry.Function{
		Name:      "rewind",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			heap, err := getSplHeapFromObject(thisObj)
			if err != nil {
				return nil, err
			}
			heap.position = 0
			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// top() method
	topImpl := &registry.Function{
		Name:      "top",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			heap, err := getSplHeapFromObject(thisObj)
			if err != nil {
				return nil, err
			}

			if len(heap.elements) == 0 {
				return nil, fmt.Errorf("can't peek at an empty heap")
			}

			return heap.elements[0], nil
		},
		Parameters: []*registry.Parameter{},
	}

	// valid() method
	validImpl := &registry.Function{
		Name:      "valid",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			heap, err := getSplHeapFromObject(thisObj)
			if err != nil {
				return nil, err
			}
			return values.NewBool(heap.position >= 0 && heap.position < len(heap.elements)), nil
		},
		Parameters: []*registry.Parameter{},
	}

	methods := map[string]*registry.MethodDescriptor{
		"__construct": {
			Name:           "__construct",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(constructorImpl),
		},
		"count": {
			Name:           "count",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(countImpl),
		},
		"current": {
			Name:           "current",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(currentImpl),
		},
		"extract": {
			Name:           "extract",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(extractImpl),
		},
		"insert": {
			Name:       "insert",
			Visibility: "public",
			Parameters: []*registry.ParameterDescriptor{
				{Name: "value", Type: "mixed"},
			},
			Implementation: NewBuiltinMethodImpl(insertImpl),
		},
		"isEmpty": {
			Name:           "isEmpty",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(isEmptyImpl),
		},
		"key": {
			Name:           "key",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(keyImpl),
		},
		"next": {
			Name:           "next",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(nextImpl),
		},
		"recoverFromCorruption": {
			Name:           "recoverFromCorruption",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(recoverFromCorruptionImpl),
		},
		"rewind": {
			Name:           "rewind",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(rewindImpl),
		},
		"top": {
			Name:           "top",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(topImpl),
		},
		"valid": {
			Name:           "valid",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(validImpl),
		},
	}

	return &registry.ClassDescriptor{
		Name:       "SplHeap",
		Parent:     "",
		Interfaces: []string{"Iterator", "Countable"},
		Traits:     []string{},
		IsAbstract: true, // SplHeap is abstract
		IsFinal:    false,
		Methods:    methods,
		Properties: map[string]*registry.PropertyDescriptor{},
		Constants:  map[string]*registry.ConstantDescriptor{},
	}
}

// Helper function to get the heap from object properties
func getSplHeapFromObject(obj *values.Value) (*SplHeap, error) {
	if obj.Type != values.TypeObject {
		return nil, fmt.Errorf("expected object, got %s", obj.Type)
	}

	objData := obj.Data.(*values.Object)
	heapValue, exists := objData.Properties["_heap"]
	if !exists {
		// Create new heap
		heap := &SplHeap{
			elements: make([]*values.Value, 0),
			position: 0,
		}
		objData.Properties["_heap"] = &values.Value{
			Type: values.TypeResource,
			Data: heap,
		}
		return heap, nil
	}

	if heapValue.Type != values.TypeResource {
		return nil, fmt.Errorf("invalid heap property type")
	}

	heap, ok := heapValue.Data.(*SplHeap)
	if !ok {
		return nil, fmt.Errorf("invalid heap property data")
	}

	return heap, nil
}

// Heap operations
func (h *SplHeap) heapifyUp(index int, compareFn func(a, b *values.Value) int) {
	parent := (index - 1) / 2
	if index > 0 && compareFn(h.elements[index], h.elements[parent]) > 0 {
		h.elements[index], h.elements[parent] = h.elements[parent], h.elements[index]
		h.heapifyUp(parent, compareFn)
	}
}

func (h *SplHeap) heapifyDown(index int, compareFn func(a, b *values.Value) int) {
	size := len(h.elements)
	leftChild := 2*index + 1
	rightChild := 2*index + 2
	largest := index

	if leftChild < size && compareFn(h.elements[leftChild], h.elements[largest]) > 0 {
		largest = leftChild
	}

	if rightChild < size && compareFn(h.elements[rightChild], h.elements[largest]) > 0 {
		largest = rightChild
	}

	if largest != index {
		h.elements[index], h.elements[largest] = h.elements[largest], h.elements[index]
		h.heapifyDown(largest, compareFn)
	}
}

// Helper function to compare values for heap operations
func compareValues(a, b *values.Value) int {
	// Simple numeric comparison for now
	aNum := a.ToFloat()
	bNum := b.ToFloat()

	if aNum > bNum {
		return 1
	} else if aNum < bNum {
		return -1
	} else {
		return 0
	}
}