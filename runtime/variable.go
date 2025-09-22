package runtime

import (
	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// GetVariableFunctions returns variable-related PHP functions
func GetVariableFunctions() []*registry.Function {
	return []*registry.Function{
		{
			Name: "isset",
			Parameters: []*registry.Parameter{
				{Name: "value", Type: "mixed"},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    999, // PHP's isset() can take multiple arguments
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// In PHP, isset() returns true only if all arguments are set and not null
				for _, arg := range args {
					if arg == nil || arg.IsNull() {
						return values.NewBool(false), nil
					}
				}
				return values.NewBool(true), nil
			},
		},
		{
			Name: "define",
			Parameters: []*registry.Parameter{
				{Name: "name", Type: "string"},
				{Name: "value", Type: "mixed"},
				{Name: "case_insensitive", Type: "bool", HasDefault: true, DefaultValue: values.NewBool(false)},
			},
			ReturnType: "bool",
			MinArgs:    2,
			MaxArgs:    3,
			IsBuiltin: true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 {
					return values.NewBool(false), nil
				}

				name := args[0].ToString()
				value := args[1]
				// Third parameter (case_insensitive) is deprecated in modern PHP but we accept it

				if name == "" {
					return values.NewBool(false), nil
				}

				// Check if constant already exists
				reg := ctx.SymbolRegistry()
				if _, exists := reg.GetConstant(name); exists {
					return values.NewBool(false), nil
				}

				// Create and register the constant
				constantDesc := &registry.ConstantDescriptor{
					Name:       name,
					Visibility: "public",
					Value:      value,
					IsFinal:    true,
				}

				err := reg.RegisterConstant(constantDesc)
				if err != nil {
					return values.NewBool(false), nil
				}

				return values.NewBool(true), nil
			},
		},
		{
			Name: "defined",
			Parameters: []*registry.Parameter{
				{Name: "name", Type: "string"},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin: true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 {
					return values.NewBool(false), nil
				}

				name := args[0].ToString()
				reg := ctx.SymbolRegistry()
				_, exists := reg.GetConstant(name)

				return values.NewBool(exists), nil
			},
		},
	}
}