package runtime

import (
	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// GetWordPressFunctions returns stub implementations of core WordPress functions
// These are minimal stubs to allow WordPress core files to load
func GetWordPressFunctions() []*registry.Function {
	return []*registry.Function{
		// wp_doing_ajax() - Check if this is an AJAX request
		{
			Name:       "wp_doing_ajax",
			Parameters: []*registry.Parameter{},
			ReturnType: "bool",
			MinArgs:    0,
			MaxArgs:    0,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// Stub: Return false (not an AJAX request)
				return values.NewBool(false), nil
			},
		},
		// wp_is_json_request() - Check if this is a JSON request
		{
			Name:       "wp_is_json_request",
			Parameters: []*registry.Parameter{},
			ReturnType: "bool",
			MinArgs:    0,
			MaxArgs:    0,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				return values.NewBool(false), nil
			},
		},
		// wp_is_serving_rest_request() - Check if this is a REST API request
		{
			Name:       "wp_is_serving_rest_request",
			Parameters: []*registry.Parameter{},
			ReturnType: "bool",
			MinArgs:    0,
			MaxArgs:    0,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				return values.NewBool(false), nil
			},
		},
		// wp_is_jsonp_request() - Check if this is a JSONP request
		{
			Name:       "wp_is_jsonp_request",
			Parameters: []*registry.Parameter{},
			ReturnType: "bool",
			MinArgs:    0,
			MaxArgs:    0,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				return values.NewBool(false), nil
			},
		},
		// wp_is_xml_request() - Check if this is an XML request
		{
			Name:       "wp_is_xml_request",
			Parameters: []*registry.Parameter{},
			ReturnType: "bool",
			MinArgs:    0,
			MaxArgs:    0,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				return values.NewBool(false), nil
			},
		},
		// apply_filters() - WordPress hook system - call filters on a value
		{
			Name: "apply_filters",
			Parameters: []*registry.Parameter{
				{Name: "hook_name", Type: "string"},
				{Name: "value", Type: "mixed"},
			},
			ReturnType: "mixed",
			MinArgs:    2,
			MaxArgs:    -1, // Variable arguments
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 {
					return values.NewNull(), nil
				}
				// Stub: Just return the value without any filtering
				return args[1], nil
			},
		},
		// wp_get_server_protocol() - Get HTTP protocol version
		{
			Name:       "wp_get_server_protocol",
			Parameters: []*registry.Parameter{},
			ReturnType: "string",
			MinArgs:    0,
			MaxArgs:    0,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// Stub: Return HTTP/1.1
				return values.NewString("HTTP/1.1"), nil
			},
		},
		// implode() is a PHP built-in, but might be needed
		// call_user_func() - Call a user function with arguments
		{
			Name: "call_user_func",
			Parameters: []*registry.Parameter{
				{Name: "callback", Type: "callable"},
			},
			ReturnType: "mixed",
			MinArgs:    1,
			MaxArgs:    -1,
			IsBuiltin:  true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 {
					return values.NewNull(), nil
				}

				callback := args[0]
				callArgs := args[1:]

				// If it's a string, try to call it as a function name
				if callback.Type == values.TypeString {
					funcName := callback.ToString()
					if userFunc, ok := ctx.LookupUserFunction(funcName); ok {
						return ctx.CallUserFunction(userFunc, callArgs)
					}
				}

				// Stub: Return null if we can't handle the callback
				return values.NewNull(), nil
			},
		},
	}
}
