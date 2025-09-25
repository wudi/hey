package spl

import (
	"fmt"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// SplDoublyLinkedListNode represents a node in the doubly linked list
type SplDoublyLinkedListNode struct {
	Value *values.Value
	Next  *SplDoublyLinkedListNode
	Prev  *SplDoublyLinkedListNode
}

// GetSplDoublyLinkedListClass returns the SplDoublyLinkedList class descriptor
func GetSplDoublyLinkedListClass() *registry.ClassDescriptor {
	// Constructor
	constructorImpl := &registry.Function{
		Name:      "__construct",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 {
				return values.NewNull(), fmt.Errorf("SplDoublyLinkedList::__construct() expects at least 1 argument, %d given", len(args))
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("__construct called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			if obj.Properties == nil {
				obj.Properties = make(map[string]*values.Value)
			}

			// Initialize empty list
			obj.Properties["__head"] = values.NewNull()
			obj.Properties["__tail"] = values.NewNull()
			obj.Properties["__count"] = values.NewInt(0)

			return thisObj, nil
		},
		Parameters: []*registry.Parameter{},
	}

	// count() method
	countImpl := &registry.Function{
		Name:      "count",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 || !args[0].IsObject() {
				return values.NewInt(0), nil
			}
			obj := args[0].Data.(*values.Object)
			if obj.Properties != nil {
				if count, ok := obj.Properties["__count"]; ok && count != nil {
					return count, nil
				}
			}
			return values.NewInt(0), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// push() method - adds to end of list
	pushImpl := &registry.Function{
		Name:      "push",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 2 {
				return values.NewNull(), fmt.Errorf("push expects 2 arguments")
			}
			if !args[0].IsObject() {
				return values.NewNull(), fmt.Errorf("push called on non-object")
			}
			obj := args[0].Data.(*values.Object)
			if obj.Properties == nil {
				obj.Properties = make(map[string]*values.Value)
			}

			value := args[1]

			// Create new node - store as Go data in a Value wrapper
			nodeData := &SplDoublyLinkedListNode{
				Value: value,
				Next:  nil,
				Prev:  nil,
			}
			newNode := &values.Value{
				Type: values.TypeObject,
				Data: nodeData,
			}

			count := int64(0)
			if countVal, ok := obj.Properties["__count"]; ok && countVal != nil {
				count = countVal.ToInt()
			}

			if count == 0 {
				// First node
				obj.Properties["__head"] = newNode
				obj.Properties["__tail"] = newNode
			} else {
				// Add to end
				if tailVal, ok := obj.Properties["__tail"]; ok && tailVal != nil && !tailVal.IsNull() {
					tailNode := tailVal.Data.(*SplDoublyLinkedListNode)
					tailNode.Next = nodeData
					nodeData.Prev = tailNode
					obj.Properties["__tail"] = newNode
				}
			}

			obj.Properties["__count"] = values.NewInt(count + 1)
			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{
			{Name: "value", Type: "mixed"},
		},
	}

	// unshift() method - adds to beginning of list
	unshiftImpl := &registry.Function{
		Name:      "unshift",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 2 {
				return values.NewNull(), fmt.Errorf("unshift expects 2 arguments")
			}
			if !args[0].IsObject() {
				return values.NewNull(), fmt.Errorf("unshift called on non-object")
			}
			obj := args[0].Data.(*values.Object)
			if obj.Properties == nil {
				obj.Properties = make(map[string]*values.Value)
			}

			value := args[1]

			// Create new node
			nodeData := &SplDoublyLinkedListNode{
				Value: value,
				Next:  nil,
				Prev:  nil,
			}
			newNode := &values.Value{
				Type: values.TypeObject,
				Data: nodeData,
			}

			count := int64(0)
			if countVal, ok := obj.Properties["__count"]; ok && countVal != nil {
				count = countVal.ToInt()
			}

			if count == 0 {
				// First node
				obj.Properties["__head"] = newNode
				obj.Properties["__tail"] = newNode
			} else {
				// Add to beginning
				if headVal, ok := obj.Properties["__head"]; ok && headVal != nil && !headVal.IsNull() {
					headNode := headVal.Data.(*SplDoublyLinkedListNode)
					headNode.Prev = nodeData
					nodeData.Next = headNode
					obj.Properties["__head"] = newNode
				}
			}

			obj.Properties["__count"] = values.NewInt(count + 1)
			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{
			{Name: "value", Type: "mixed"},
		},
	}

	// pop() method - removes from end
	popImpl := &registry.Function{
		Name:      "pop",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 || !args[0].IsObject() {
				return values.NewNull(), fmt.Errorf("pop called on non-object")
			}
			obj := args[0].Data.(*values.Object)
			if obj.Properties == nil {
				return values.NewNull(), fmt.Errorf("list is empty")
			}

			count := int64(0)
			if countVal, ok := obj.Properties["__count"]; ok && countVal != nil {
				count = countVal.ToInt()
			}

			if count == 0 {
				return values.NewNull(), fmt.Errorf("list is empty")
			}

			tailVal, ok := obj.Properties["__tail"]
			if !ok || tailVal == nil || tailVal.IsNull() {
				return values.NewNull(), fmt.Errorf("list corruption")
			}

			tailNode := tailVal.Data.(*SplDoublyLinkedListNode)
			value := tailNode.Value

			if count == 1 {
				// Last node
				obj.Properties["__head"] = values.NewNull()
				obj.Properties["__tail"] = values.NewNull()
			} else {
				// Update tail
				if tailNode.Prev != nil {
					tailNode.Prev.Next = nil
					newTail := &values.Value{
						Type: values.TypeObject,
						Data: tailNode.Prev,
					}
					obj.Properties["__tail"] = newTail
				}
			}

			obj.Properties["__count"] = values.NewInt(count - 1)
			return value, nil
		},
		Parameters: []*registry.Parameter{},
	}

	// shift() method - removes from beginning
	shiftImpl := &registry.Function{
		Name:      "shift",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 || !args[0].IsObject() {
				return values.NewNull(), fmt.Errorf("shift called on non-object")
			}
			obj := args[0].Data.(*values.Object)
			if obj.Properties == nil {
				return values.NewNull(), fmt.Errorf("list is empty")
			}

			count := int64(0)
			if countVal, ok := obj.Properties["__count"]; ok && countVal != nil {
				count = countVal.ToInt()
			}

			if count == 0 {
				return values.NewNull(), fmt.Errorf("list is empty")
			}

			headVal, ok := obj.Properties["__head"]
			if !ok || headVal == nil || headVal.IsNull() {
				return values.NewNull(), fmt.Errorf("list corruption")
			}

			headNode := headVal.Data.(*SplDoublyLinkedListNode)
			value := headNode.Value

			if count == 1 {
				// Last node
				obj.Properties["__head"] = values.NewNull()
				obj.Properties["__tail"] = values.NewNull()
			} else {
				// Update head
				if headNode.Next != nil {
					headNode.Next.Prev = nil
					newHead := &values.Value{
						Type: values.TypeObject,
						Data: headNode.Next,
					}
					obj.Properties["__head"] = newHead
				}
			}

			obj.Properties["__count"] = values.NewInt(count - 1)
			return value, nil
		},
		Parameters: []*registry.Parameter{},
	}

	// top() method - peek at end without removing
	topImpl := &registry.Function{
		Name:      "top",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 || !args[0].IsObject() {
				return values.NewNull(), fmt.Errorf("top called on non-object")
			}
			obj := args[0].Data.(*values.Object)
			if obj.Properties == nil {
				return values.NewNull(), fmt.Errorf("list is empty")
			}

			count := int64(0)
			if countVal, ok := obj.Properties["__count"]; ok && countVal != nil {
				count = countVal.ToInt()
			}

			if count == 0 {
				return values.NewNull(), fmt.Errorf("list is empty")
			}

			tailVal, ok := obj.Properties["__tail"]
			if !ok || tailVal == nil || tailVal.IsNull() {
				return values.NewNull(), fmt.Errorf("list corruption")
			}

			tailNode := tailVal.Data.(*SplDoublyLinkedListNode)
			return tailNode.Value, nil
		},
		Parameters: []*registry.Parameter{},
	}

	// bottom() method - peek at beginning without removing
	bottomImpl := &registry.Function{
		Name:      "bottom",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 || !args[0].IsObject() {
				return values.NewNull(), fmt.Errorf("bottom called on non-object")
			}
			obj := args[0].Data.(*values.Object)
			if obj.Properties == nil {
				return values.NewNull(), fmt.Errorf("list is empty")
			}

			count := int64(0)
			if countVal, ok := obj.Properties["__count"]; ok && countVal != nil {
				count = countVal.ToInt()
			}

			if count == 0 {
				return values.NewNull(), fmt.Errorf("list is empty")
			}

			headVal, ok := obj.Properties["__head"]
			if !ok || headVal == nil || headVal.IsNull() {
				return values.NewNull(), fmt.Errorf("list corruption")
			}

			headNode := headVal.Data.(*SplDoublyLinkedListNode)
			return headNode.Value, nil
		},
		Parameters: []*registry.Parameter{},
	}

	// isEmpty() method
	isEmptyImpl := &registry.Function{
		Name:      "isEmpty",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 || !args[0].IsObject() {
				return values.NewBool(true), nil
			}
			obj := args[0].Data.(*values.Object)
			if obj.Properties != nil {
				if count, ok := obj.Properties["__count"]; ok && count != nil {
					return values.NewBool(count.ToInt() == 0), nil
				}
			}
			return values.NewBool(true), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// Create method descriptors
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
		"push": {
			Name:       "push",
			Visibility: "public",
			Parameters: []*registry.ParameterDescriptor{
				{Name: "value", Type: "mixed"},
			},
			Implementation: NewBuiltinMethodImpl(pushImpl),
		},
		"unshift": {
			Name:       "unshift",
			Visibility: "public",
			Parameters: []*registry.ParameterDescriptor{
				{Name: "value", Type: "mixed"},
			},
			Implementation: NewBuiltinMethodImpl(unshiftImpl),
		},
		"pop": {
			Name:           "pop",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(popImpl),
		},
		"shift": {
			Name:           "shift",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(shiftImpl),
		},
		"top": {
			Name:           "top",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(topImpl),
		},
		"bottom": {
			Name:           "bottom",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(bottomImpl),
		},
		"isEmpty": {
			Name:           "isEmpty",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(isEmptyImpl),
		},
	}

	return &registry.ClassDescriptor{
		Name:       "SplDoublyLinkedList",
		Parent:     "",
		Interfaces: []string{"Iterator", "Countable"},
		Properties: make(map[string]*registry.PropertyDescriptor),
		Methods:    methods,
		Constants:  make(map[string]*registry.ConstantDescriptor),
	}
}