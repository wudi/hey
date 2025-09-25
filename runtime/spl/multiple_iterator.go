package spl

import (
	"fmt"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// GetMultipleIteratorClass returns the MultipleIterator class descriptor
func GetMultipleIteratorClass() *registry.ClassDescriptor {
	// Constructor
	constructorImpl := &registry.Function{
		Name:      "__construct",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 {
				return values.NewNull(), fmt.Errorf("MultipleIterator::__construct() expects at least 1 argument")
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("__construct called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			if obj.Properties == nil {
				obj.Properties = make(map[string]*values.Value)
			}

			// Initialize storage for iterators and their info/keys
			obj.Properties["__iterators"] = values.NewArray()
			obj.Properties["__iterator_info"] = values.NewArray()

			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// attachIterator() method
	attachIteratorImpl := &registry.Function{
		Name:      "attachIterator",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 2 || len(args) > 3 {
				return values.NewNull(), fmt.Errorf("attachIterator() expects 2 or 3 arguments, %d given", len(args))
			}

			thisObj := args[0]
			iterator := args[1]
			var info *values.Value = values.NewNull()

			if len(args) > 2 {
				info = args[2]
			}

			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("attachIterator() called on non-object")
			}

			if !iterator.IsObject() {
				return values.NewNull(), fmt.Errorf("Iterator must be an object")
			}

			obj := thisObj.Data.(*values.Object)
			iterators := obj.Properties["__iterators"]
			iteratorInfo := obj.Properties["__iterator_info"]

			if iterators == nil || !iterators.IsArray() {
				return values.NewNull(), fmt.Errorf("Iterators array not initialized")
			}

			iteratorsArray := iterators.Data.(*values.Array)
			infoArray := iteratorInfo.Data.(*values.Array)

			// Add iterator to the end of the array
			nextIndex := int64(len(iteratorsArray.Elements))
			iteratorsArray.Elements[nextIndex] = iterator
			infoArray.Elements[nextIndex] = info

			return values.NewNull(), nil
		},
	}

	// detachIterator() method
	detachIteratorImpl := &registry.Function{
		Name:      "detachIterator",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 2 {
				return values.NewNull(), fmt.Errorf("detachIterator() expects exactly 2 arguments")
			}

			thisObj := args[0]
			iterator := args[1]

			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("detachIterator() called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			iterators := obj.Properties["__iterators"]
			iteratorInfo := obj.Properties["__iterator_info"]

			if iterators == nil || !iterators.IsArray() {
				return values.NewNull(), fmt.Errorf("Iterators array not initialized")
			}

			iteratorsArray := iterators.Data.(*values.Array)
			infoArray := iteratorInfo.Data.(*values.Array)

			// Find and remove the iterator
			newIterators := make(map[interface{}]*values.Value)
			newInfo := make(map[interface{}]*values.Value)
			newIndex := int64(0)

			for i := int64(0); i < int64(len(iteratorsArray.Elements)); i++ {
				if iter, exists := iteratorsArray.Elements[i]; exists {
					// Compare iterator objects by reference
					if iter != iterator {
						newIterators[newIndex] = iter
						if info, infoExists := infoArray.Elements[i]; infoExists {
							newInfo[newIndex] = info
						} else {
							newInfo[newIndex] = values.NewNull()
						}
						newIndex++
					}
				}
			}

			// Replace arrays with new ones
			iteratorsArray.Elements = newIterators
			infoArray.Elements = newInfo

			return values.NewNull(), nil
		},
	}

	// containsIterator() method
	containsIteratorImpl := &registry.Function{
		Name:      "containsIterator",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 2 {
				return values.NewNull(), fmt.Errorf("containsIterator() expects exactly 2 arguments")
			}

			thisObj := args[0]
			iterator := args[1]

			if !thisObj.IsObject() {
				return values.NewBool(false), fmt.Errorf("containsIterator() called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			iterators := obj.Properties["__iterators"]

			if iterators == nil || !iterators.IsArray() {
				return values.NewBool(false), nil
			}

			iteratorsArray := iterators.Data.(*values.Array)

			// Search for the iterator
			for _, iter := range iteratorsArray.Elements {
				if iter == iterator {
					return values.NewBool(true), nil
				}
			}

			return values.NewBool(false), nil
		},
	}

	// countIterators() method
	countIteratorsImpl := &registry.Function{
		Name:      "countIterators",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("countIterators() expects exactly 1 argument")
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewInt(0), fmt.Errorf("countIterators() called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			iterators := obj.Properties["__iterators"]

			if iterators == nil || !iterators.IsArray() {
				return values.NewInt(0), nil
			}

			iteratorsArray := iterators.Data.(*values.Array)
			return values.NewInt(int64(len(iteratorsArray.Elements))), nil
		},
	}

	// current() method
	currentImpl := &registry.Function{
		Name:      "current",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("current() expects exactly 1 argument")
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("current() called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			iterators := obj.Properties["__iterators"]

			if iterators == nil || !iterators.IsArray() {
				return values.NewNull(), nil
			}

			iteratorsArray := iterators.Data.(*values.Array)
			result := values.NewArray()

			// Get current value from each iterator
			for i := int64(0); i < int64(len(iteratorsArray.Elements)); i++ {
				if iter, exists := iteratorsArray.Elements[i]; exists {
					current, err := callIteratorMethod(ctx, iter, "current", []*values.Value{iter})
					if err != nil {
						return values.NewNull(), err
					}
					result.ArraySet(values.NewInt(i), current)
				}
			}

			return result, nil
		},
	}

	// key() method
	keyImpl := &registry.Function{
		Name:      "key",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("key() expects exactly 1 argument")
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("key() called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			iterators := obj.Properties["__iterators"]

			if iterators == nil || !iterators.IsArray() {
				return values.NewNull(), nil
			}

			iteratorsArray := iterators.Data.(*values.Array)
			result := values.NewArray()

			// Get key from each iterator
			for i := int64(0); i < int64(len(iteratorsArray.Elements)); i++ {
				if iter, exists := iteratorsArray.Elements[i]; exists {
					key, err := callIteratorMethod(ctx, iter, "key", []*values.Value{iter})
					if err != nil {
						return values.NewNull(), err
					}
					result.ArraySet(values.NewInt(i), key)
				}
			}

			return result, nil
		},
	}

	// next() method
	nextImpl := &registry.Function{
		Name:      "next",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("next() expects exactly 1 argument")
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("next() called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			iterators := obj.Properties["__iterators"]

			if iterators == nil || !iterators.IsArray() {
				return values.NewNull(), nil
			}

			iteratorsArray := iterators.Data.(*values.Array)

			// Call next() on all iterators
			for i := int64(0); i < int64(len(iteratorsArray.Elements)); i++ {
				if iter, exists := iteratorsArray.Elements[i]; exists {
					_, err := callIteratorMethod(ctx, iter, "next", []*values.Value{iter})
					if err != nil {
						return values.NewNull(), err
					}
				}
			}

			return values.NewNull(), nil
		},
	}

	// rewind() method
	rewindImpl := &registry.Function{
		Name:      "rewind",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("rewind() expects exactly 1 argument")
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("rewind() called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			iterators := obj.Properties["__iterators"]

			if iterators == nil || !iterators.IsArray() {
				return values.NewNull(), nil
			}

			iteratorsArray := iterators.Data.(*values.Array)

			// Call rewind() on all iterators
			for i := int64(0); i < int64(len(iteratorsArray.Elements)); i++ {
				if iter, exists := iteratorsArray.Elements[i]; exists {
					_, err := callIteratorMethod(ctx, iter, "rewind", []*values.Value{iter})
					if err != nil {
						return values.NewNull(), err
					}
				}
			}

			return values.NewNull(), nil
		},
	}

	// valid() method
	validImpl := &registry.Function{
		Name:      "valid",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("valid() expects exactly 1 argument")
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewBool(false), fmt.Errorf("valid() called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			iterators := obj.Properties["__iterators"]

			if iterators == nil || !iterators.IsArray() {
				return values.NewBool(false), nil
			}

			iteratorsArray := iterators.Data.(*values.Array)

			if len(iteratorsArray.Elements) == 0 {
				return values.NewBool(false), nil
			}

			// MultipleIterator is valid only if ALL iterators are valid
			// (stops at shortest iterator)
			for i := int64(0); i < int64(len(iteratorsArray.Elements)); i++ {
				if iter, exists := iteratorsArray.Elements[i]; exists {
					validResult, err := callIteratorMethod(ctx, iter, "valid", []*values.Value{iter})
					if err != nil {
						return values.NewBool(false), err
					}
					if !validResult.ToBool() {
						return values.NewBool(false), nil
					}
				}
			}

			return values.NewBool(true), nil
		},
	}

	return &registry.ClassDescriptor{
		Name: "MultipleIterator",
		Methods: map[string]*registry.MethodDescriptor{
			"__construct": {
				Name:           "__construct",
				Visibility:     "public",
				Parameters:     []*registry.ParameterDescriptor{},
				Implementation: NewBuiltinMethodImpl(constructorImpl),
			},
			"attachIterator": {
				Name:       "attachIterator",
				Visibility: "public",
				Parameters: []*registry.ParameterDescriptor{
					{Name: "iterator", Type: "Iterator"},
					{Name: "info", Type: "string", HasDefault: true, DefaultValue: values.NewNull()},
				},
				Implementation: NewBuiltinMethodImpl(attachIteratorImpl),
			},
			"detachIterator": {
				Name:       "detachIterator",
				Visibility: "public",
				Parameters: []*registry.ParameterDescriptor{
					{Name: "iterator", Type: "Iterator"},
				},
				Implementation: NewBuiltinMethodImpl(detachIteratorImpl),
			},
			"containsIterator": {
				Name:       "containsIterator",
				Visibility: "public",
				Parameters: []*registry.ParameterDescriptor{
					{Name: "iterator", Type: "Iterator"},
				},
				Implementation: NewBuiltinMethodImpl(containsIteratorImpl),
			},
			"countIterators": {
				Name:           "countIterators",
				Visibility:     "public",
				Parameters:     []*registry.ParameterDescriptor{},
				Implementation: NewBuiltinMethodImpl(countIteratorsImpl),
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
			"valid": {
				Name:           "valid",
				Visibility:     "public",
				Parameters:     []*registry.ParameterDescriptor{},
				Implementation: NewBuiltinMethodImpl(validImpl),
			},
		},
		Interfaces: []string{"Iterator"},
	}
}