package spl

import (
	"fmt"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// GetAppendIteratorClass returns the AppendIterator class descriptor
func GetAppendIteratorClass() *registry.ClassDescriptor {
	// Constructor
	constructorImpl := &registry.Function{
		Name:      "__construct",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 {
				return values.NewNull(), fmt.Errorf("AppendIterator::__construct() expects at least 1 argument")
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("__construct called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			if obj.Properties == nil {
				obj.Properties = make(map[string]*values.Value)
			}

			// Initialize empty array of iterators and current index
			obj.Properties["__iterators"] = values.NewArray()
			obj.Properties["__current_index"] = values.NewInt(0)

			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// append() method
	appendImpl := &registry.Function{
		Name:      "append",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 2 {
				return values.NewNull(), fmt.Errorf("AppendIterator::append() expects exactly 2 arguments, %d given", len(args))
			}

			thisObj := args[0]
			iterator := args[1]

			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("append called on non-object")
			}

			if !iterator.IsObject() {
				return values.NewNull(), fmt.Errorf("append() expects parameter 1 to be an iterator object")
			}

			obj := thisObj.Data.(*values.Object)
			iterators := obj.Properties["__iterators"]

			// Add iterator to the array
			nextIndex := int64(iterators.ArrayCount())
			iterators.ArraySet(values.NewInt(nextIndex), iterator)

			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{
			{Name: "iterator", Type: "Iterator"},
		},
	}

	// getIteratorIndex() method
	getIteratorIndexImpl := &registry.Function{
		Name:      "getIteratorIndex",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("AppendIterator::getIteratorIndex() expects exactly 1 argument (this), %d given", len(args))
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("getIteratorIndex called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			if currentIndex, ok := obj.Properties["__current_index"]; ok && currentIndex.IsInt() {
				return currentIndex, nil
			}

			return values.NewInt(0), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// getInnerIterator() method
	getInnerIteratorImpl := &registry.Function{
		Name:      "getInnerIterator",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("AppendIterator::getInnerIterator() expects exactly 1 argument (this), %d given", len(args))
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("getInnerIterator called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			iterators := obj.Properties["__iterators"]
			currentIndex := obj.Properties["__current_index"]

			if !iterators.IsArray() || !currentIndex.IsInt() {
				return values.NewNull(), nil
			}

			// Get current iterator
			currentIter := iterators.ArrayGet(currentIndex)
			if currentIter != nil && !currentIter.IsNull() && currentIter.IsObject() {
				return currentIter, nil
			}

			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// current() method - delegates to current active iterator
	currentImpl := &registry.Function{
		Name:      "current",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("AppendIterator::current() expects exactly 1 argument (this), %d given", len(args))
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("current called on non-object")
			}

			// Get current iterator
			currentIterator, err := getCurrentIterator(thisObj)
			if err != nil {
				return values.NewNull(), nil
			}

			// Call current() on the active iterator
			return callIteratorMethod(ctx, currentIterator, "current", []*values.Value{currentIterator})
		},
		Parameters: []*registry.Parameter{},
	}

	// key() method - delegates to current active iterator
	keyImpl := &registry.Function{
		Name:      "key",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("AppendIterator::key() expects exactly 1 argument (this), %d given", len(args))
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("key called on non-object")
			}

			// Get current iterator
			currentIterator, err := getCurrentIterator(thisObj)
			if err != nil {
				return values.NewNull(), nil
			}

			// Call key() on the active iterator
			return callIteratorMethod(ctx, currentIterator, "key", []*values.Value{currentIterator})
		},
		Parameters: []*registry.Parameter{},
	}

	// valid() method - checks if current iterator is valid
	validImpl := &registry.Function{
		Name:      "valid",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("AppendIterator::valid() expects exactly 1 argument (this), %d given", len(args))
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("valid called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			iterators := obj.Properties["__iterators"]
			currentIndex := obj.Properties["__current_index"]

			if !iterators.IsArray() || !currentIndex.IsInt() {
				return values.NewBool(false), nil
			}

			currentIdx := currentIndex.ToInt()
			iteratorCount := int64(iterators.ArrayCount())

			// If we've gone past all iterators, not valid
			if currentIdx >= iteratorCount {
				return values.NewBool(false), nil
			}

			// Get current iterator
			currentIterator, err := getCurrentIterator(thisObj)
			if err != nil {
				return values.NewBool(false), nil
			}

			// Check if current iterator is valid
			result, err := callIteratorMethod(ctx, currentIterator, "valid", []*values.Value{currentIterator})
			if err != nil {
				return values.NewBool(false), nil
			}

			return result, nil
		},
		Parameters: []*registry.Parameter{},
	}

	// next() method - advances current iterator or moves to next iterator
	nextImpl := &registry.Function{
		Name:      "next",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("AppendIterator::next() expects exactly 1 argument (this), %d given", len(args))
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("next called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			iterators := obj.Properties["__iterators"]
			currentIndex := obj.Properties["__current_index"]

			if !iterators.IsArray() || !currentIndex.IsInt() {
				return values.NewNull(), nil
			}

			// Get current iterator
			currentIterator, err := getCurrentIterator(thisObj)
			if err != nil {
				return values.NewNull(), nil
			}

			// Call next() on current iterator
			callIteratorMethod(ctx, currentIterator, "next", []*values.Value{currentIterator})

			// Check if current iterator is still valid
			validResult, err := callIteratorMethod(ctx, currentIterator, "valid", []*values.Value{currentIterator})
			if err != nil || !validResult.ToBool() {
				// Move to next iterator
				currentIdx := currentIndex.ToInt()
				iteratorCount := int64(iterators.ArrayCount())

				// Keep advancing to next non-empty iterator
				for currentIdx+1 < iteratorCount {
					currentIdx++
					obj.Properties["__current_index"] = values.NewInt(currentIdx)

					// Get next iterator
					nextIterator, err := getCurrentIterator(thisObj)
					if err != nil {
						continue
					}

					// Rewind the next iterator
					callIteratorMethod(ctx, nextIterator, "rewind", []*values.Value{nextIterator})

					// Check if it's valid
					validResult, err := callIteratorMethod(ctx, nextIterator, "valid", []*values.Value{nextIterator})
					if err == nil && validResult.ToBool() {
						break // Found a valid iterator
					}
				}
			}

			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// rewind() method - rewinds to first iterator
	rewindImpl := &registry.Function{
		Name:      "rewind",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("AppendIterator::rewind() expects exactly 1 argument (this), %d given", len(args))
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("rewind called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			iterators := obj.Properties["__iterators"]

			if !iterators.IsArray() {
				return values.NewNull(), nil
			}

			// Reset to first iterator
			obj.Properties["__current_index"] = values.NewInt(0)

			// Find first valid iterator
			iteratorCount := int64(iterators.ArrayCount())
			for i := int64(0); i < iteratorCount; i++ {
				obj.Properties["__current_index"] = values.NewInt(i)

				currentIterator, err := getCurrentIterator(thisObj)
				if err != nil {
					continue
				}

				// Rewind this iterator
				callIteratorMethod(ctx, currentIterator, "rewind", []*values.Value{currentIterator})

				// Check if it's valid
				validResult, err := callIteratorMethod(ctx, currentIterator, "valid", []*values.Value{currentIterator})
				if err == nil && validResult.ToBool() {
					break // Found first valid iterator
				}
			}

			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// Create method descriptors
	methods := map[string]*registry.MethodDescriptor{
		"__construct": {
			Name:       "__construct",
			Visibility: "public",
			Parameters: []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(constructorImpl),
		},
		"append": {
			Name:       "append",
			Visibility: "public",
			Parameters: []*registry.ParameterDescriptor{
				{Name: "iterator", Type: "Iterator"},
			},
			Implementation: NewBuiltinMethodImpl(appendImpl),
		},
		"getIteratorIndex": {
			Name:           "getIteratorIndex",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(getIteratorIndexImpl),
		},
		"getInnerIterator": {
			Name:           "getInnerIterator",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(getInnerIteratorImpl),
		},
		"current": {
			Name:           "current",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(currentImpl),
		},
		"key": {
			Name:           "key",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(keyImpl),
		},
		"valid": {
			Name:           "valid",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(validImpl),
		},
		"next": {
			Name:           "next",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(nextImpl),
		},
		"rewind": {
			Name:           "rewind",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(rewindImpl),
		},
	}

	return &registry.ClassDescriptor{
		Name:        "AppendIterator",
		Methods:     methods,
		Constants:   make(map[string]*registry.ConstantDescriptor),
		Interfaces:  []string{"Iterator", "OuterIterator"},
	}
}

// Helper function to get current active iterator
func getCurrentIterator(thisObj *values.Value) (*values.Value, error) {
	obj := thisObj.Data.(*values.Object)
	iterators := obj.Properties["__iterators"]
	currentIndex := obj.Properties["__current_index"]

	if !iterators.IsArray() || !currentIndex.IsInt() {
		return nil, fmt.Errorf("invalid AppendIterator state")
	}

	// Get current iterator
	currentIter := iterators.ArrayGet(currentIndex)
	if currentIter == nil || currentIter.IsNull() || !currentIter.IsObject() {
		return nil, fmt.Errorf("current iterator not found")
	}

	return currentIter, nil
}