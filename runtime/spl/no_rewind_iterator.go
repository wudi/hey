package spl

import (
	"fmt"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// GetNoRewindIteratorClass returns the NoRewindIterator class descriptor
func GetNoRewindIteratorClass() *registry.ClassDescriptor {
	// Constructor
	constructorImpl := &registry.Function{
		Name:      "__construct",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 {
				return values.NewNull(), fmt.Errorf("NoRewindIterator::__construct() expects at least 1 argument")
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("__construct called on non-object")
			}

			// Handle VM parameter passing issue - make iterator parameter optional
			var iterator *values.Value = values.NewNull()

			if len(args) > 1 && !args[1].IsNull() {
				iterator = args[1]
			}

			obj := thisObj.Data.(*values.Object)
			if obj.Properties == nil {
				obj.Properties = make(map[string]*values.Value)
			}

			// Store the iterator and rewind state
			obj.Properties["__iterator"] = iterator
			obj.Properties["__rewind_called"] = values.NewBool(false) // Track if rewind has been called

			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{
			{Name: "iterator", Type: "Iterator", HasDefault: true, DefaultValue: values.NewNull()},
		},
	}

	// getInnerIterator() method - implements OuterIterator
	getInnerIteratorImpl := &registry.Function{
		Name:      "getInnerIterator",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("NoRewindIterator::getInnerIterator() expects exactly 1 argument (this), %d given", len(args))
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("getInnerIterator called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			if iterator, ok := obj.Properties["__iterator"]; ok && iterator != nil {
				return iterator, nil
			}

			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// rewind() method - only works once, subsequent calls are ignored
	rewindImpl := &registry.Function{
		Name:      "rewind",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("NoRewindIterator::rewind() expects exactly 1 argument (this), %d given", len(args))
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("rewind called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			iterator, hasIterator := obj.Properties["__iterator"]
			rewindCalled, hasRewindState := obj.Properties["__rewind_called"]

			if !hasIterator || !hasRewindState {
				return values.NewNull(), nil
			}

			// Only allow rewind if it hasn't been called before
			if !rewindCalled.ToBool() && iterator != nil && iterator.IsObject() {
				// Mark that rewind has been called
				obj.Properties["__rewind_called"] = values.NewBool(true)

				// Call rewind on the inner iterator
				callIteratorMethod(ctx, iterator, "rewind", []*values.Value{iterator})
			}
			// If rewind has already been called, ignore subsequent rewind calls

			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// current() method - delegates to inner iterator
	currentImpl := &registry.Function{
		Name:      "current",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("NoRewindIterator::current() expects exactly 1 argument (this), %d given", len(args))
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("current called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			if iterator, ok := obj.Properties["__iterator"]; ok && iterator != nil && iterator.IsObject() {
				return callIteratorMethod(ctx, iterator, "current", []*values.Value{iterator})
			}

			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// key() method - delegates to inner iterator
	keyImpl := &registry.Function{
		Name:      "key",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("NoRewindIterator::key() expects exactly 1 argument (this), %d given", len(args))
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("key called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			if iterator, ok := obj.Properties["__iterator"]; ok && iterator != nil && iterator.IsObject() {
				return callIteratorMethod(ctx, iterator, "key", []*values.Value{iterator})
			}

			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// valid() method - delegates to inner iterator
	validImpl := &registry.Function{
		Name:      "valid",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("NoRewindIterator::valid() expects exactly 1 argument (this), %d given", len(args))
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("valid called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			if iterator, ok := obj.Properties["__iterator"]; ok && iterator != nil && iterator.IsObject() {
				result, err := callIteratorMethod(ctx, iterator, "valid", []*values.Value{iterator})
				if err != nil {
					return values.NewBool(false), nil
				}
				return result, nil
			}

			return values.NewBool(false), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// next() method - delegates to inner iterator
	nextImpl := &registry.Function{
		Name:      "next",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("NoRewindIterator::next() expects exactly 1 argument (this), %d given", len(args))
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("next called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			if iterator, ok := obj.Properties["__iterator"]; ok && iterator != nil && iterator.IsObject() {
				callIteratorMethod(ctx, iterator, "next", []*values.Value{iterator})
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
			Parameters: []*registry.ParameterDescriptor{
				{Name: "iterator", Type: "Iterator", HasDefault: true, DefaultValue: values.NewNull()},
			},
			Implementation: NewBuiltinMethodImpl(constructorImpl),
		},
		"getInnerIterator": {
			Name:           "getInnerIterator",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(getInnerIteratorImpl),
		},
		"rewind": {
			Name:           "rewind",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(rewindImpl),
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
	}

	return &registry.ClassDescriptor{
		Name:       "NoRewindIterator",
		Methods:    methods,
		Constants:  make(map[string]*registry.ConstantDescriptor),
		Interfaces: []string{"Iterator", "OuterIterator"},
	}
}