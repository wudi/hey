package runtime

import (
	"fmt"
	"sync"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

type AssertState struct {
	mu       sync.RWMutex
	active   int64
	warning  int64
	bail     int64
}

var globalAssertState = &AssertState{
	active:  1,
	warning: 1,
	bail:    0,
}

func GetAssertFunctions() []*registry.Function {
	return []*registry.Function{
		{
			Name: "assert_options",
			Parameters: []*registry.Parameter{
				{Name: "option", Type: "int"},
				{Name: "value", Type: "mixed", HasDefault: true, DefaultValue: values.NewNull()},
			},
			ReturnType: "mixed",
			MinArgs:    1,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 {
					return values.NewNull(), nil
				}

				option := args[0].ToInt()

				globalAssertState.mu.Lock()
				defer globalAssertState.mu.Unlock()

				var currentValue int64
				switch option {
				case 1: // ASSERT_ACTIVE
					currentValue = globalAssertState.active
					if len(args) > 1 && !args[1].IsNull() {
						globalAssertState.active = args[1].ToInt()
					}
				case 4: // ASSERT_WARNING
					currentValue = globalAssertState.warning
					if len(args) > 1 && !args[1].IsNull() {
						globalAssertState.warning = args[1].ToInt()
					}
				case 3: // ASSERT_BAIL
					currentValue = globalAssertState.bail
					if len(args) > 1 && !args[1].IsNull() {
						globalAssertState.bail = args[1].ToInt()
					}
				default:
					return values.NewBool(false), nil
				}

				return values.NewInt(currentValue), nil
			},
		},
		{
			Name: "assert",
			Parameters: []*registry.Parameter{
				{Name: "assertion", Type: "mixed"},
				{Name: "description", Type: "string|Throwable", HasDefault: true, DefaultValue: values.NewNull()},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 {
					return values.NewBool(false), nil
				}

				globalAssertState.mu.RLock()
				active := globalAssertState.active
				globalAssertState.mu.RUnlock()

				if active == 0 {
					return values.NewBool(true), nil
				}

				assertion := args[0]

				assertionResult := assertion.ToBool()
				if assertionResult {
					return values.NewBool(true), nil
				}

				description := "assert(false)"
				if len(args) > 1 && !args[1].IsNull() {
					description = args[1].ToString()
				}

				return values.NewBool(false), fmt.Errorf("AssertionError: %s", description)
			},
		},
	}
}