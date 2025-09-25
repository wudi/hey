package spl

import (
	"fmt"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// GetInfiniteIteratorClass returns the InfiniteIterator class descriptor
func GetInfiniteIteratorClass() *registry.ClassDescriptor {
	// Constructor
	constructorImpl := &registry.Function{
		Name:      "__construct",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 {
				return values.NewNull(), fmt.Errorf("InfiniteIterator::__construct() expects at least 1 argument")
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

			// Store the iterator
			obj.Properties["__iterator"] = iterator

			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{
			{Name: "iterator", Type: "Iterator", HasDefault: true, DefaultValue: values.NewNull()},
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
			iterator := obj.Properties["__iterator"]

			if iterator == nil || iterator.IsNull() {
				return values.NewNull(), nil
			}

			// Delegate to inner iterator's current method
			return callIteratorMethod(ctx, iterator, "current", []*values.Value{iterator})
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
			iterator := obj.Properties["__iterator"]

			if iterator == nil || iterator.IsNull() {
				return values.NewNull(), nil
			}

			// Delegate to inner iterator's key method
			return callIteratorMethod(ctx, iterator, "key", []*values.Value{iterator})
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
			iterator := obj.Properties["__iterator"]

			if iterator == nil || iterator.IsNull() {
				return values.NewNull(), nil
			}

			// Call next() on inner iterator
			_, err := callIteratorMethod(ctx, iterator, "next", []*values.Value{iterator})
			if err != nil {
				return values.NewNull(), err
			}

			// Check if inner iterator is still valid after next()
			validResult, err := callIteratorMethod(ctx, iterator, "valid", []*values.Value{iterator})
			if err != nil {
				return values.NewNull(), err
			}

			// If inner iterator becomes invalid, rewind it to create infinite loop
			if !validResult.ToBool() {
				_, err := callIteratorMethod(ctx, iterator, "rewind", []*values.Value{iterator})
				if err != nil {
					return values.NewNull(), err
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
			iterator := obj.Properties["__iterator"]

			if iterator == nil || iterator.IsNull() {
				return values.NewNull(), nil
			}

			// Delegate to inner iterator's rewind method
			return callIteratorMethod(ctx, iterator, "rewind", []*values.Value{iterator})
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
			iterator := obj.Properties["__iterator"]

			if iterator == nil || iterator.IsNull() {
				return values.NewBool(false), nil
			}

			// For InfiniteIterator, we're valid if the inner iterator has any elements
			// First check if inner iterator is currently valid
			validResult, err := callIteratorMethod(ctx, iterator, "valid", []*values.Value{iterator})
			if err != nil {
				return values.NewBool(false), nil
			}

			// If not valid, try rewinding and checking again to see if it has elements
			if !validResult.ToBool() {
				_, err := callIteratorMethod(ctx, iterator, "rewind", []*values.Value{iterator})
				if err != nil {
					return values.NewBool(false), nil
				}

				// Check validity after rewind
				validResult, err = callIteratorMethod(ctx, iterator, "valid", []*values.Value{iterator})
				if err != nil {
					return values.NewBool(false), nil
				}
			}

			// InfiniteIterator is valid if the inner iterator has at least one element
			return validResult, nil
		},
	}

	// getInnerIterator() method - implements OuterIterator
	getInnerIteratorImpl := &registry.Function{
		Name:      "getInnerIterator",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("getInnerIterator() expects exactly 1 argument")
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("getInnerIterator() called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			iterator := obj.Properties["__iterator"]

			if iterator == nil {
				return values.NewNull(), nil
			}

			return iterator, nil
		},
	}

	return &registry.ClassDescriptor{
		Name: "InfiniteIterator",
		Methods: map[string]*registry.MethodDescriptor{
			"__construct": {
				Name:           "__construct",
				Visibility:     "public",
				Parameters:     []*registry.ParameterDescriptor{},
				Implementation: NewBuiltinMethodImpl(constructorImpl),
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
			"getInnerIterator": {
				Name:           "getInnerIterator",
				Visibility:     "public",
				Parameters:     []*registry.ParameterDescriptor{},
				Implementation: NewBuiltinMethodImpl(getInnerIteratorImpl),
			},
		},
		Interfaces: []string{"OuterIterator", "Iterator"},
	}
}