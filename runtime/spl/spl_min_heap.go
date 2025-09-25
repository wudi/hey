package spl

import (
	"fmt"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// GetSplMinHeapClass returns the SplMinHeap class descriptor
func GetSplMinHeapClass() *registry.ClassDescriptor {
	// Constructor
	constructorImpl := &registry.Function{
		Name:      "__construct",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// compare() method
	compareImpl := &registry.Function{
		Name:      "compare",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 3 {
				return nil, fmt.Errorf("SplMinHeap::compare() expects 2 parameters, %d given", len(args)-1)
			}

			value1 := args[1]
			value2 := args[2]

			// Min heap: return negative of comparison (reverse max heap logic)
			result := -compareValues(value1, value2)
			return values.NewInt(int64(result)), nil
		},
		Parameters: []*registry.Parameter{
			{Name: "value1", Type: "mixed"},
			{Name: "value2", Type: "mixed"},
		},
	}

	// Get parent methods from SplHeap
	parentClass := GetSplHeapClass()
	methods := make(map[string]*registry.MethodDescriptor)

	// Copy all parent methods
	for name, method := range parentClass.Methods {
		methods[name] = method
	}

	// Override specific methods
	methods["__construct"] = &registry.MethodDescriptor{
		Name:           "__construct",
		Visibility:     "public",
		Parameters:     []*registry.ParameterDescriptor{},
		Implementation: NewBuiltinMethodImpl(constructorImpl),
	}
	methods["compare"] = &registry.MethodDescriptor{
		Name:       "compare",
		Visibility: "protected",
		Parameters: []*registry.ParameterDescriptor{
			{Name: "value1", Type: "mixed"},
			{Name: "value2", Type: "mixed"},
		},
		Implementation: NewBuiltinMethodImpl(compareImpl),
	}

	return &registry.ClassDescriptor{
		Name:       "SplMinHeap",
		Parent:     "SplHeap",
		Interfaces: []string{"Iterator", "Countable"},
		Traits:     []string{},
		IsAbstract: false,
		IsFinal:    false,
		Methods:    methods,
		Properties: map[string]*registry.PropertyDescriptor{},
		Constants:  map[string]*registry.ConstantDescriptor{},
	}
}