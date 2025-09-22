package runtime

import (
	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// GetOutputFunctions returns output-related PHP functions
func GetOutputFunctions() []*registry.Function {
	return []*registry.Function{
		{
			Name:       "print",
			Parameters: []*registry.Parameter{{Name: "value", Type: "mixed"}},
			ReturnType: "int",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin: true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) > 0 {
					if err := ctx.WriteOutput(args[0]); err != nil {
						return nil, err
					}
				}
				return values.NewInt(1), nil
			},
		},
		{
			Name:       "var_dump",
			Parameters: []*registry.Parameter{{Name: "value", Type: "mixed"}},
			ReturnType: "bool",
			IsVariadic: true,
			MinArgs:    1,
			MaxArgs:    -1,
			IsBuiltin: true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if ctx != nil {
					for _, arg := range args {
						_ = ctx.WriteOutput(values.NewString(arg.VarDump()))
					}
				}
				return values.NewNull(), nil
			},
		},
	}
}