package spl

import (
	"fmt"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// GetIteratorIteratorClass returns the IteratorIterator class descriptor
func GetIteratorIteratorClass() *registry.ClassDescriptor {
	// Constructor
	constructorImpl := &registry.Function{
		Name:      "__construct",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 {
				return values.NewNull(), fmt.Errorf("IteratorIterator::__construct() expects at least 1 argument")
			}

			thisObj := args[0]
			var iterator *values.Value
			if len(args) > 1 {
				iterator = args[1]
			} else {
				iterator = values.NewNull()
			}

			// Note: Temporarily disabled validation due to VM parameter passing issue
			// TODO: Re-enable when VM constructor parameter passing is fixed
			// if iterator.IsNull() {
			//     return values.NewNull(), fmt.Errorf("IteratorIterator::__construct() expects iterator argument")
			// }

			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("__construct called on non-object")
			}

			// Note: Temporarily disabled Traversable validation due to VM parameter passing issue
			// TODO: Re-enable when VM constructor parameter passing is fixed

			// Basic validation - ensure we have an object
			if !iterator.IsNull() && iterator.IsObject() {
				iteratorObj := iterator.Data.(*values.Object)

				// For IteratorAggregate objects, get their iterator
				if iteratorObj.ClassName == "ArrayObject" {
					registry := ctx.SymbolRegistry()
					if registry != nil {
						if class, err := registry.GetClass("ArrayObject"); err == nil && class != nil {
							if getIterMethod, ok := class.Methods["getIterator"]; ok {
								impl := getIterMethod.Implementation.(*BuiltinMethodImpl)
								actualIterator, err := impl.GetFunction().Builtin(ctx, []*values.Value{iterator})
								if err == nil && actualIterator.IsObject() {
									iterator = actualIterator
								}
							}
						}
					}
				}
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
			{Name: "iterator", Type: "Traversable"},
			{Name: "class", Type: "string", HasDefault: true, DefaultValue: values.NewString("")},
		},
	}

	// getInnerIterator() method
	getInnerIteratorImpl := &registry.Function{
		Name:      "getInnerIterator",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("IteratorIterator::getInnerIterator() expects exactly 1 argument (this), %d given", len(args))
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("getInnerIterator called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			if iterator, ok := obj.Properties["__iterator"]; ok && iterator.IsObject() {
				return iterator, nil
			}

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
				return values.NewNull(), fmt.Errorf("IteratorIterator::current() expects exactly 1 argument (this), %d given", len(args))
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("current called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			if iterator, ok := obj.Properties["__iterator"]; ok && iterator.IsObject() {
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
				return values.NewNull(), fmt.Errorf("IteratorIterator::key() expects exactly 1 argument (this), %d given", len(args))
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("key called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			if iterator, ok := obj.Properties["__iterator"]; ok && iterator.IsObject() {
				// Call key() on the inner iterator
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
				return values.NewNull(), fmt.Errorf("IteratorIterator::valid() expects exactly 1 argument (this), %d given", len(args))
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("valid called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			if iterator, ok := obj.Properties["__iterator"]; ok && iterator.IsObject() {
				// Call valid() on the inner iterator
				return callIteratorMethod(ctx, iterator, "valid", []*values.Value{iterator})
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
				return values.NewNull(), fmt.Errorf("IteratorIterator::next() expects exactly 1 argument (this), %d given", len(args))
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("next called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			if iterator, ok := obj.Properties["__iterator"]; ok && iterator.IsObject() {
				// Call next() on the inner iterator
				return callIteratorMethod(ctx, iterator, "next", []*values.Value{iterator})
			}

			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// rewind() method - delegates to inner iterator
	rewindImpl := &registry.Function{
		Name:      "rewind",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("IteratorIterator::rewind() expects exactly 1 argument (this), %d given", len(args))
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("rewind called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			if iterator, ok := obj.Properties["__iterator"]; ok && iterator.IsObject() {
				// Call rewind() on the inner iterator
				return callIteratorMethod(ctx, iterator, "rewind", []*values.Value{iterator})
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
				{Name: "iterator", Type: "Traversable", HasDefault: true, DefaultValue: values.NewNull()},
				{Name: "class", Type: "string", HasDefault: true, DefaultValue: values.NewString("")},
			},
			Implementation: NewBuiltinMethodImpl(constructorImpl),
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
		Name:        "IteratorIterator",
		Methods:     methods,
		Constants:   make(map[string]*registry.ConstantDescriptor),
		Interfaces:  []string{"Iterator", "OuterIterator"},
	}
}

// Helper function to call methods on iterator objects
func callIteratorMethod(ctx registry.BuiltinCallContext, iterator *values.Value, methodName string, args []*values.Value) (*values.Value, error) {
	if !iterator.IsObject() {
		return values.NewNull(), fmt.Errorf("cannot call method on non-object")
	}

	iteratorObj := iterator.Data.(*values.Object)
	registry := ctx.SymbolRegistry()
	if registry == nil {
		return values.NewNull(), fmt.Errorf("registry not available")
	}

	class, err := registry.GetClass(iteratorObj.ClassName)
	if err != nil || class == nil {
		return values.NewNull(), fmt.Errorf("class %s not found", iteratorObj.ClassName)
	}

	method, ok := class.Methods[methodName]
	if !ok {
		return values.NewNull(), fmt.Errorf("method %s not found in class %s", methodName, iteratorObj.ClassName)
	}

	impl := method.Implementation.(*BuiltinMethodImpl)
	return impl.GetFunction().Builtin(ctx, args)
}