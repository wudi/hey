package runtime

import (
	"fmt"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// GetConcurrencyFunctions returns concurrency-related PHP functions
func GetConcurrencyFunctions() []*registry.Function {
	return []*registry.Function{
		{
			Name:       "go",
			Parameters: []*registry.Parameter{{Name: "closure", Type: "callable"}},
			ReturnType: "Goroutine",
			IsVariadic: true,
			MinArgs:    1,
			MaxArgs:    -1,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 {
					return nil, fmt.Errorf("go() expects at least one argument")
				}
				closureVal := args[0]
				if closureVal == nil || !closureVal.IsCallable() {
					return nil, fmt.Errorf("go() expects a callable as first argument")
				}
				closure := closureVal.ClosureGet()
				if closure == nil {
					return nil, fmt.Errorf("go() invalid closure")
				}
				useVars := make(map[string]*values.Value)
				for i, arg := range args[1:] {
					useVars[fmt.Sprintf("var_%d", i)] = arg
				}
				return values.NewGoroutine(closure, useVars), nil
			},
		},
	}
}