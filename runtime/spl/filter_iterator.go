package spl

import (
	"fmt"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// GetFilterIteratorClass returns the FilterIterator class descriptor
func GetFilterIteratorClass() *registry.ClassDescriptor {
	// Constructor
	constructorImpl := &registry.Function{
		Name:      "__construct",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 2 {
				return values.NewNull(), fmt.Errorf("FilterIterator::__construct() expects at least 2 arguments")
			}

			thisObj := args[0]
			iterator := args[1]

			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("__construct called on non-object")
			}

			if !iterator.IsObject() {
				return values.NewNull(), fmt.Errorf("FilterIterator::__construct() expects parameter 1 to be an iterator")
			}

			obj := thisObj.Data.(*values.Object)
			if obj.Properties == nil {
				obj.Properties = make(map[string]*values.Value)
			}

			// Store the iterator (similar to IteratorIterator)
			obj.Properties["__iterator"] = iterator

			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{
			{Name: "iterator", Type: "Iterator"},
		},
	}

	// getInnerIterator() method
	getInnerIteratorImpl := &registry.Function{
		Name:      "getInnerIterator",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("FilterIterator::getInnerIterator() expects exactly 1 argument (this), %d given", len(args))
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

	// accept() method - abstract in FilterIterator
	acceptImpl := &registry.Function{
		Name:      "accept",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			// This is an abstract method that should be overridden by subclasses
			return values.NewNull(), fmt.Errorf("FilterIterator::accept() is abstract and must be implemented by subclasses")
		},
		Parameters: []*registry.Parameter{},
	}

	// current() method - delegates to inner iterator after finding valid item
	currentImpl := &registry.Function{
		Name:      "current",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("FilterIterator::current() expects exactly 1 argument (this), %d given", len(args))
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("current called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			if iterator, ok := obj.Properties["__iterator"]; ok && iterator != nil && iterator.IsObject() {
				// Call current() on the inner iterator
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
				return values.NewNull(), fmt.Errorf("FilterIterator::key() expects exactly 1 argument (this), %d given", len(args))
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("key called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			if iterator, ok := obj.Properties["__iterator"]; ok && iterator != nil && iterator.IsObject() {
				// Call key() on the inner iterator
				return callIteratorMethod(ctx, iterator, "key", []*values.Value{iterator})
			}

			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// valid() method - checks if current position is valid and accepted
	validImpl := &registry.Function{
		Name:      "valid",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("FilterIterator::valid() expects exactly 1 argument (this), %d given", len(args))
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("valid called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			if iterator, ok := obj.Properties["__iterator"]; ok && iterator != nil && iterator.IsObject() {
				// Check if inner iterator is valid
				validResult, err := callIteratorMethod(ctx, iterator, "valid", []*values.Value{iterator})
				if err != nil || !validResult.ToBool() {
					return values.NewBool(false), nil
				}

				// For the base FilterIterator class, we can't call accept() since it's abstract
				// In a real implementation, subclasses would override accept()
				// For now, we'll return the inner iterator's validity
				return values.NewBool(true), nil
			}

			return values.NewBool(false), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// next() method - advances to next valid item
	nextImpl := &registry.Function{
		Name:      "next",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("FilterIterator::next() expects exactly 1 argument (this), %d given", len(args))
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("next called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			if iterator, ok := obj.Properties["__iterator"]; ok && iterator != nil && iterator.IsObject() {
				// Move inner iterator to next position
				callIteratorMethod(ctx, iterator, "next", []*values.Value{iterator})

				// For abstract FilterIterator, we can't call accept()
				// In concrete subclasses, they would implement logic to keep advancing
				// until accept() returns true or iterator becomes invalid
			}

			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// rewind() method - rewinds inner iterator and positions to first valid item
	rewindImpl := &registry.Function{
		Name:      "rewind",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("FilterIterator::rewind() expects exactly 1 argument (this), %d given", len(args))
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("rewind called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			if iterator, ok := obj.Properties["__iterator"]; ok && iterator != nil && iterator.IsObject() {
				// Rewind the inner iterator
				callIteratorMethod(ctx, iterator, "rewind", []*values.Value{iterator})

				// For abstract FilterIterator, we can't call accept() to find first valid item
				// In concrete subclasses, they would implement logic to advance to first valid item
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
				{Name: "iterator", Type: "Iterator"},
			},
			Implementation: NewBuiltinMethodImpl(constructorImpl),
		},
		"getInnerIterator": {
			Name:           "getInnerIterator",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(getInnerIteratorImpl),
		},
		"accept": {
			Name:           "accept",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(acceptImpl),
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
		Name:        "FilterIterator",
		Methods:     methods,
		Constants:   make(map[string]*registry.ConstantDescriptor),
		Interfaces:  []string{"Iterator", "OuterIterator"},
		IsAbstract:  true, // FilterIterator is an abstract class
	}
}