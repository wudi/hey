package runtime

import (
	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// GetOutputFunctions returns all output buffering related functions
func GetOutputFunctions() []*registry.Function {
	return []*registry.Function{
		// print - Output a string
		{
			Name:       "print",
			Parameters: []*registry.Parameter{{Name: "value", Type: "mixed"}},
			ReturnType: "int",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) > 0 {
					if err := ctx.WriteOutput(args[0]); err != nil {
						return nil, err
					}
				}
				return values.NewInt(1), nil
			},
		},

		// var_dump - Dumps information about a variable
		{
			Name:       "var_dump",
			Parameters: []*registry.Parameter{{Name: "value", Type: "mixed"}},
			ReturnType: "bool",
			IsVariadic: true,
			MinArgs:    1,
			MaxArgs:    -1,
			IsBuiltin:  true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if ctx != nil {
					for _, arg := range args {
						_ = ctx.WriteOutput(values.NewString(arg.VarDump()))
					}
				}
				return values.NewNull(), nil
			},
		},

		// ob_start - Start output buffering
		{
			Name: "ob_start",
			Parameters: []*registry.Parameter{
				{Name: "callback", Type: "callable", HasDefault: true, DefaultValue: values.NewNull()},
				{Name: "chunk_size", Type: "int", HasDefault: true, DefaultValue: values.NewInt(0)},
				{Name: "flags", Type: "int", HasDefault: true, DefaultValue: values.NewInt(0)},
			},
			IsBuiltin: true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				obs := ctx.GetOutputBufferStack()
				if obs == nil {
					return values.NewBool(false), nil
				}

				// Get optional parameters
				var handler string
				chunkSize := int64(0)
				flags := int64(0)

				if len(args) > 0 && !args[0].IsNull() {
					// TODO: Handle callback functions properly
					handler = args[0].String()
				}

				if len(args) > 1 {
					chunkSize = args[1].ToInt()
				}

				if len(args) > 2 {
					flags = args[2].ToInt()
				}

				// Start output buffering
				success := obs.Start(handler, int(chunkSize), int(flags))
				return values.NewBool(success), nil
			},
		},

		// ob_get_contents - Return the contents of the output buffer
		{
			Name:      "ob_get_contents",
			Parameters: []*registry.Parameter{},
			IsBuiltin: true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				obs := ctx.GetOutputBufferStack()
				if obs == nil || obs.GetLevel() == 0 {
					return values.NewBool(false), nil
				}
				contents := obs.GetContents()
				return values.NewString(contents), nil
			},
		},

		// ob_get_length - Return the length of the output buffer
		{
			Name:      "ob_get_length",
			Parameters: []*registry.Parameter{},
			IsBuiltin: true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				obs := ctx.GetOutputBufferStack()
				if obs == nil || obs.GetLevel() == 0 {
					return values.NewBool(false), nil
				}
				length := obs.GetLength()
				return values.NewInt(int64(length)), nil
			},
		},

		// ob_get_level - Return the nesting level of the output buffering mechanism
		{
			Name:      "ob_get_level",
			Parameters: []*registry.Parameter{},
			IsBuiltin: true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				obs := ctx.GetOutputBufferStack()
				if obs == nil {
					return values.NewInt(0), nil
				}
				level := obs.GetLevel()
				return values.NewInt(int64(level)), nil
			},
		},

		// ob_clean - Clean (erase) the contents of the active output buffer
		{
			Name:      "ob_clean",
			Parameters: []*registry.Parameter{},
			IsBuiltin: true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				obs := ctx.GetOutputBufferStack()
				if obs == nil {
					return values.NewBool(false), nil
				}
				success := obs.Clean()
				return values.NewBool(success), nil
			},
		},

		// ob_end_clean - Clean (erase) the contents of the active output buffer and turn it off
		{
			Name:      "ob_end_clean",
			Parameters: []*registry.Parameter{},
			IsBuiltin: true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				obs := ctx.GetOutputBufferStack()
				if obs == nil {
					return values.NewBool(false), nil
				}
				success := obs.EndClean()
				return values.NewBool(success), nil
			},
		},

		// ob_flush - Flush (send) the return value of the active output handler
		{
			Name:      "ob_flush",
			Parameters: []*registry.Parameter{},
			IsBuiltin: true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				obs := ctx.GetOutputBufferStack()
				if obs == nil {
					return values.NewBool(false), nil
				}
				success := obs.Flush()
				return values.NewBool(success), nil
			},
		},

		// ob_end_flush - Flush (send) the return value of the active output handler and turn the active output buffer off
		{
			Name:      "ob_end_flush",
			Parameters: []*registry.Parameter{},
			IsBuiltin: true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				obs := ctx.GetOutputBufferStack()
				if obs == nil {
					return values.NewBool(false), nil
				}
				success := obs.EndFlush()
				return values.NewBool(success), nil
			},
		},

		// ob_get_clean - Get the contents of the active output buffer and turn it off
		{
			Name:      "ob_get_clean",
			Parameters: []*registry.Parameter{},
			IsBuiltin: true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				obs := ctx.GetOutputBufferStack()
				if obs == nil {
					return values.NewBool(false), nil
				}
				contents, success := obs.GetClean()
				if !success {
					return values.NewBool(false), nil
				}
				return values.NewString(contents), nil
			},
		},

		// ob_get_flush - Flush (send) the return value of the active output handler, return the contents and turn it off
		{
			Name:      "ob_get_flush",
			Parameters: []*registry.Parameter{},
			IsBuiltin: true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				obs := ctx.GetOutputBufferStack()
				if obs == nil {
					return values.NewBool(false), nil
				}
				contents, success := obs.GetFlush()
				if !success {
					return values.NewBool(false), nil
				}
				return values.NewString(contents), nil
			},
		},

		// ob_get_status - Get status of output buffers
		{
			Name: "ob_get_status",
			Parameters: []*registry.Parameter{
				{Name: "full_status", Type: "bool", HasDefault: true, DefaultValue: values.NewBool(false)},
			},
			IsBuiltin: true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				obs := ctx.GetOutputBufferStack()
				if obs == nil {
					return values.NewArray(), nil
				}

				fullStatus := false
				if len(args) > 0 {
					fullStatus = args[0].ToBool()
				}

				if fullStatus {
					return obs.GetStatusFull(), nil
				}
				return obs.GetStatus(), nil
			},
		},

		// ob_implicit_flush - Turn implicit flush on/off
		{
			Name: "ob_implicit_flush",
			Parameters: []*registry.Parameter{
				{Name: "enable", Type: "bool", HasDefault: true, DefaultValue: values.NewBool(true)},
			},
			IsBuiltin: true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				obs := ctx.GetOutputBufferStack()
				if obs == nil {
					return values.NewNull(), nil
				}

				enable := true
				if len(args) > 0 {
					enable = args[0].ToBool()
				}

				obs.SetImplicitFlush(enable)
				return values.NewNull(), nil
			},
		},

		// ob_list_handlers - List all output handlers in use
		{
			Name:      "ob_list_handlers",
			Parameters: []*registry.Parameter{},
			IsBuiltin: true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				obs := ctx.GetOutputBufferStack()
				if obs == nil {
					return values.NewArray(), nil
				}

				handlers := obs.ListHandlers()
				result := values.NewArray()
				for _, handler := range handlers {
					result.ArraySet(nil, values.NewString(handler))
				}
				return result, nil
			},
		},

		// flush - Flush system output buffer
		{
			Name:      "flush",
			Parameters: []*registry.Parameter{},
			IsBuiltin: true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				obs := ctx.GetOutputBufferStack()
				if obs != nil {
					obs.FlushSystem()
				}
				return values.NewNull(), nil
			},
		},

		// print_r - Prints human-readable information about a variable
		{
			Name: "print_r",
			Parameters: []*registry.Parameter{
				{Name: "value", Type: "mixed"},
				{Name: "return", Type: "bool", HasDefault: true, DefaultValue: values.NewBool(false)},
			},
			ReturnType: "mixed",
			MinArgs:    1,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 1 {
					return values.NewNull(), nil
				}

				returnOutput := false
				if len(args) > 1 {
					returnOutput = args[1].ToBool()
				}

				output := args[0].PrintR()

				if returnOutput {
					return values.NewString(output), nil
				}

				// Output directly
				if ctx != nil {
					_ = ctx.WriteOutput(values.NewString(output))
				}
				return values.NewBool(true), nil
			},
		},

		// output_add_rewrite_var - Add URL rewriter values (stub implementation)
		{
			Name: "output_add_rewrite_var",
			Parameters: []*registry.Parameter{
				{Name: "name", Type: "string"},
				{Name: "value", Type: "string"},
			},
			IsBuiltin: true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// Stub implementation - in a real web environment this would modify URLs
				return values.NewBool(true), nil
			},
		},

		// output_reset_rewrite_vars - Reset URL rewriter values (stub implementation)
		{
			Name:      "output_reset_rewrite_vars",
			Parameters: []*registry.Parameter{},
			IsBuiltin: true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// Stub implementation - in a real web environment this would reset URL rewriting
				return values.NewBool(true), nil
			},
		},
	}
}