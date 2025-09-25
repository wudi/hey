package spl

import (
	"fmt"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// GetEmptyIteratorClass returns the EmptyIterator class descriptor
func GetEmptyIteratorClass() *registry.ClassDescriptor {
	// Constructor
	constructorImpl := &registry.Function{
		Name:      "__construct",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("EmptyIterator constructor expects exactly 1 argument (this), %d given", len(args))
			}
			// No initialization needed for empty iterator
			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// current() method - throws BadMethodCallException
	currentImpl := &registry.Function{
		Name:      "current",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("EmptyIterator::current() expects exactly 1 argument (this), %d given", len(args))
			}
			// Throw BadMethodCallException
			return values.NewNull(), fmt.Errorf("BadMethodCallException: Accessing the value of an EmptyIterator")
		},
		Parameters: []*registry.Parameter{},
	}

	// key() method - throws BadMethodCallException
	keyImpl := &registry.Function{
		Name:      "key",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("EmptyIterator::key() expects exactly 1 argument (this), %d given", len(args))
			}
			// Throw BadMethodCallException
			return values.NewNull(), fmt.Errorf("BadMethodCallException: Accessing the key of an EmptyIterator")
		},
		Parameters: []*registry.Parameter{},
	}

	// valid() method - always returns false
	validImpl := &registry.Function{
		Name:      "valid",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("EmptyIterator::valid() expects exactly 1 argument (this), %d given", len(args))
			}
			// Always return false for empty iterator
			return values.NewBool(false), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// next() method - does nothing
	nextImpl := &registry.Function{
		Name:      "next",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("EmptyIterator::next() expects exactly 1 argument (this), %d given", len(args))
			}
			// No-op for empty iterator
			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// rewind() method - does nothing
	rewindImpl := &registry.Function{
		Name:      "rewind",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("EmptyIterator::rewind() expects exactly 1 argument (this), %d given", len(args))
			}
			// No-op for empty iterator
			return values.NewNull(), nil
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
		Name:        "EmptyIterator",
		Methods:     methods,
		Constants:   make(map[string]*registry.ConstantDescriptor),
		Interfaces:  []string{"Iterator"},
	}
}

