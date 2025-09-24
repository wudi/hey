package runtime

import (
	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// GetWindowsFunctions returns Windows-specific PHP functions
func GetWindowsFunctions() []*registry.Function {
	return []*registry.Function{
		{
			Name: "sapi_windows_vt100_support",
			Parameters: []*registry.Parameter{
				{Name: "stream", Type: "resource"},
				{Name: "enable", Type: "bool", HasDefault: true, DefaultValue: values.NewNull()},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// sapi_windows_vt100_support — Get or set VT100 support for the specified stream
				// associated to an output buffer of a Windows console.
				//
				// This function is Windows-specific and not applicable on Unix systems.
				// Always return false as VT100 support is not implemented in hey-codex.

				if len(args) == 0 || args[0] == nil {
					return values.NewBool(false), nil
				}

				// Validate that the first parameter is a resource type
				stream := args[0]
				if stream.Type != values.TypeResource {
					return values.NewBool(false), nil
				}

				// The second parameter (enable) is optional and defaults to null
				// If provided, it would set the VT100 support state, but we always return false
				// regardless of the enable parameter since VT100 support is not implemented

				return values.NewBool(false), nil
			},
		},
	}
}