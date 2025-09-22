package runtime

import (
	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)


// GetInterfaces returns all iterator-related interface definitions
func GetInterfaces() []*registry.Interface {
	return []*registry.Interface{
		getTraversableInterface(),
		getIteratorInterface(),
	}
}

// GetIteratorClasses returns all iterator-related class definitions
func GetIteratorClasses() []*registry.ClassDescriptor {
	return []*registry.ClassDescriptor{
		getGeneratorClass(),
	}
}

func getTraversableInterface() *registry.Interface {
	return &registry.Interface{
		Name:    "Traversable",
		Methods: make(map[string]*registry.InterfaceMethod),
		Extends: []string{},
	}
}

func getIteratorInterface() *registry.Interface {
	// Create method definitions for Iterator interface
	methods := map[string]*registry.InterfaceMethod{
		"current": {
			Name:       "current",
			Visibility: "public",
			Parameters: []*registry.Parameter{},
			ReturnType: "mixed",
		},
		"key": {
			Name:       "key",
			Visibility: "public",
			Parameters: []*registry.Parameter{},
			ReturnType: "mixed",
		},
		"next": {
			Name:       "next",
			Visibility: "public",
			Parameters: []*registry.Parameter{},
			ReturnType: "void",
		},
		"rewind": {
			Name:       "rewind",
			Visibility: "public",
			Parameters: []*registry.Parameter{},
			ReturnType: "void",
		},
		"valid": {
			Name:       "valid",
			Visibility: "public",
			Parameters: []*registry.Parameter{},
			ReturnType: "bool",
		},
	}

	return &registry.Interface{
		Name:    "Iterator",
		Methods: methods,
		Extends: []string{"Traversable"},
	}
}

func getGeneratorClass() *registry.ClassDescriptor {
	// Create method implementations for Generator class
	currentImpl := &registry.Function{
		Name:      "current",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 || !args[0].IsObject() {
				return values.NewNull(), nil
			}
			obj := args[0].Data.(*values.Object)

			// Get generator
			if genVal, ok := obj.Properties["__channel_generator"]; ok && genVal != nil {
				if gen, ok := genVal.Data.(*Generator); ok {
					return gen.Current(), nil
				}
			}

			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{},
	}

	keyImpl := &registry.Function{
		Name:      "key",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 || !args[0].IsObject() {
				return values.NewNull(), nil
			}
			obj := args[0].Data.(*values.Object)

			// Get generator
			if genVal, ok := obj.Properties["__channel_generator"]; ok && genVal != nil {
				if gen, ok := genVal.Data.(*Generator); ok {
					return gen.Key(), nil
				}
			}

			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{},
	}

	nextImpl := &registry.Function{
		Name:      "next",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 || !args[0].IsObject() {
				return values.NewNull(), nil
			}
			obj := args[0].Data.(*values.Object)

			// Get generator
			if genVal, ok := obj.Properties["__channel_generator"]; ok && genVal != nil {
				if gen, ok := genVal.Data.(*Generator); ok {
					gen.Next() // Advance to next value
					return values.NewNull(), nil
				}
			}

			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{},
	}

	rewindImpl := &registry.Function{
		Name:      "rewind",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 || !args[0].IsObject() {
				return values.NewNull(), nil
			}
			obj := args[0].Data.(*values.Object)

			// Get generator
			if genVal, ok := obj.Properties["__channel_generator"]; ok && genVal != nil {
				if gen, ok := genVal.Data.(*Generator); ok {
					if err := gen.Rewind(); err != nil {
						return values.NewNull(), err
					}
					gen.Next() // Start and get first value
					return values.NewNull(), nil
				}
			}

			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{},
	}

	validImpl := &registry.Function{
		Name:      "valid",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 || !args[0].IsObject() {
				return values.NewBool(false), nil
			}
			obj := args[0].Data.(*values.Object)

			// Get generator
			if genVal, ok := obj.Properties["__channel_generator"]; ok && genVal != nil {
				if gen, ok := genVal.Data.(*Generator); ok {
					return values.NewBool(gen.Valid()), nil
				}
			}

			return values.NewBool(false), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// Create method descriptors
	methods := map[string]*registry.MethodDescriptor{
		"current": {
			Name:           "current",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: &BuiltinMethodImpl{function: currentImpl},
		},
		"key": {
			Name:           "key",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: &BuiltinMethodImpl{function: keyImpl},
		},
		"next": {
			Name:           "next",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: &BuiltinMethodImpl{function: nextImpl},
		},
		"rewind": {
			Name:           "rewind",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: &BuiltinMethodImpl{function: rewindImpl},
		},
		"valid": {
			Name:           "valid",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: &BuiltinMethodImpl{function: validImpl},
		},
	}

	return &registry.ClassDescriptor{
		Name:       "Generator",
		IsFinal:    true,
		Interfaces: []string{"Iterator"},
		Methods:    methods,
		Properties: make(map[string]*registry.PropertyDescriptor),
		Constants:  make(map[string]*registry.ConstantDescriptor),
	}
}