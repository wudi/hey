package spl

import (
	"fmt"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// PriorityQueueItem represents an item with priority in the priority queue
type PriorityQueueItem struct {
	Value    *values.Value
	Priority *values.Value
}

// SplPriorityQueue represents the SplPriorityQueue class
type SplPriorityQueue struct {
	elements []*PriorityQueueItem
	position int
	extractFlags int // For setExtractFlags/getExtractFlags
}

// Constants for extract flags
const (
	EXTR_DATA     = 0x00000001
	EXTR_PRIORITY = 0x00000002
	EXTR_BOTH     = 0x00000003
)

// GetSplPriorityQueueClass returns the SplPriorityQueue class descriptor
func GetSplPriorityQueueClass() *registry.ClassDescriptor {
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
			pq, err := getSplPriorityQueueFromObject(thisObj)
			if err != nil {
				return nil, err
			}
			return values.NewInt(int64(len(pq.elements))), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// extract() method
	extractImpl := &registry.Function{
		Name:      "extract",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			pq, err := getSplPriorityQueueFromObject(thisObj)
			if err != nil {
				return nil, err
			}

			if len(pq.elements) == 0 {
				return nil, fmt.Errorf("can't extract from an empty heap")
			}

			// Get the root element
			root := pq.elements[0]

			// Move the last element to the root and remove it
			lastIndex := len(pq.elements) - 1
			pq.elements[0] = pq.elements[lastIndex]
			pq.elements = pq.elements[:lastIndex]

			// Restore heap property
			if len(pq.elements) > 0 {
				pq.heapifyDown(0)
			}

			// Reset iterator position
			pq.position = 0

			// Return based on extract flags
			switch pq.extractFlags {
			case EXTR_DATA:
				return root.Value, nil
			case EXTR_PRIORITY:
				return root.Priority, nil
			case EXTR_BOTH:
				result := values.NewArray()
				result.ArraySet(values.NewString("data"), root.Value)
				result.ArraySet(values.NewString("priority"), root.Priority)
				return result, nil
			default:
				return root.Value, nil
			}
		},
		Parameters: []*registry.Parameter{},
	}

	// insert() method
	insertImpl := &registry.Function{
		Name:      "insert",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 3 {
				return nil, fmt.Errorf("SplPriorityQueue::insert() expects 2 parameters, %d given", len(args)-1)
			}

			thisObj := args[0]
			value := args[1]
			priority := args[2]

			pq, err := getSplPriorityQueueFromObject(thisObj)
			if err != nil {
				return nil, err
			}

			// Create new item
			item := &PriorityQueueItem{
				Value:    value,
				Priority: priority,
			}

			// Add the item to the end
			pq.elements = append(pq.elements, item)

			// Restore heap property
			pq.heapifyUp(len(pq.elements) - 1)

			return values.NewBool(true), nil
		},
		Parameters: []*registry.Parameter{
			{Name: "value", Type: "mixed"},
			{Name: "priority", Type: "mixed"},
		},
	}

	// isEmpty() method
	isEmptyImpl := &registry.Function{
		Name:      "isEmpty",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			pq, err := getSplPriorityQueueFromObject(thisObj)
			if err != nil {
				return nil, err
			}
			return values.NewBool(len(pq.elements) == 0), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// top() method
	topImpl := &registry.Function{
		Name:      "top",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			pq, err := getSplPriorityQueueFromObject(thisObj)
			if err != nil {
				return nil, err
			}

			if len(pq.elements) == 0 {
				return nil, fmt.Errorf("can't peek at an empty heap")
			}

			item := pq.elements[0]

			// Return based on extract flags
			switch pq.extractFlags {
			case EXTR_DATA:
				return item.Value, nil
			case EXTR_PRIORITY:
				return item.Priority, nil
			case EXTR_BOTH:
				result := values.NewArray()
				result.ArraySet(values.NewString("data"), item.Value)
				result.ArraySet(values.NewString("priority"), item.Priority)
				return result, nil
			default:
				return item.Value, nil
			}
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
				{Name: "priority", Type: "mixed"},
			},
			Implementation: NewBuiltinMethodImpl(insertImpl),
		},
		"isEmpty": {
			Name:           "isEmpty",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(isEmptyImpl),
		},
		"top": {
			Name:           "top",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(topImpl),
		},
	}

	// Add constants
	constants := map[string]*registry.ConstantDescriptor{
		"EXTR_DATA": {
			Name:       "EXTR_DATA",
			Visibility: "public",
			Value:      values.NewInt(EXTR_DATA),
			IsFinal:    true,
		},
		"EXTR_PRIORITY": {
			Name:       "EXTR_PRIORITY",
			Visibility: "public",
			Value:      values.NewInt(EXTR_PRIORITY),
			IsFinal:    true,
		},
		"EXTR_BOTH": {
			Name:       "EXTR_BOTH",
			Visibility: "public",
			Value:      values.NewInt(EXTR_BOTH),
			IsFinal:    true,
		},
	}

	return &registry.ClassDescriptor{
		Name:       "SplPriorityQueue",
		Parent:     "SplHeap",
		Interfaces: []string{"Iterator", "Countable"},
		Traits:     []string{},
		IsAbstract: false,
		IsFinal:    false,
		Methods:    methods,
		Properties: map[string]*registry.PropertyDescriptor{},
		Constants:  constants,
	}
}

// Helper function to get the priority queue from object properties
func getSplPriorityQueueFromObject(obj *values.Value) (*SplPriorityQueue, error) {
	if obj.Type != values.TypeObject {
		return nil, fmt.Errorf("expected object, got %s", obj.Type)
	}

	objData := obj.Data.(*values.Object)
	pqValue, exists := objData.Properties["_priorityQueue"]
	if !exists {
		// Create new priority queue
		pq := &SplPriorityQueue{
			elements:     make([]*PriorityQueueItem, 0),
			position:     0,
			extractFlags: EXTR_DATA, // Default to extracting data only
		}
		objData.Properties["_priorityQueue"] = &values.Value{
			Type: values.TypeResource,
			Data: pq,
		}
		return pq, nil
	}

	if pqValue.Type != values.TypeResource {
		return nil, fmt.Errorf("invalid priority queue property type")
	}

	pq, ok := pqValue.Data.(*SplPriorityQueue)
	if !ok {
		return nil, fmt.Errorf("invalid priority queue property data")
	}

	return pq, nil
}

// Priority queue heap operations
func (pq *SplPriorityQueue) heapifyUp(index int) {
	parent := (index - 1) / 2
	if index > 0 && pq.comparePriorities(pq.elements[index].Priority, pq.elements[parent].Priority) > 0 {
		pq.elements[index], pq.elements[parent] = pq.elements[parent], pq.elements[index]
		pq.heapifyUp(parent)
	}
}

func (pq *SplPriorityQueue) heapifyDown(index int) {
	size := len(pq.elements)
	leftChild := 2*index + 1
	rightChild := 2*index + 2
	largest := index

	if leftChild < size && pq.comparePriorities(pq.elements[leftChild].Priority, pq.elements[largest].Priority) > 0 {
		largest = leftChild
	}

	if rightChild < size && pq.comparePriorities(pq.elements[rightChild].Priority, pq.elements[largest].Priority) > 0 {
		largest = rightChild
	}

	if largest != index {
		pq.elements[index], pq.elements[largest] = pq.elements[largest], pq.elements[index]
		pq.heapifyDown(largest)
	}
}

func (pq *SplPriorityQueue) comparePriorities(p1, p2 *values.Value) int {
	return compareValues(p1, p2)
}