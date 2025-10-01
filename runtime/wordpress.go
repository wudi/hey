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
		// did_action() - Check if an action has been executed
		{
			Name: "did_action",
			Parameters: []*registry.Parameter{
				{Name: "hook_name", Type: "string"},
			},
			ReturnType: "int",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// Stub: Return 0 (action not executed yet)
				return values.NewInt(0), nil
			},
		},
		// status_header() - Set HTTP status header
		{
			Name: "status_header",
			Parameters: []*registry.Parameter{
				{Name: "code", Type: "int"},
				{Name: "description", Type: "string", DefaultValue: values.NewString("")},
			},
			ReturnType: "void",
			MinArgs:    1,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 {
					return values.NewNull(), nil
				}
				// Stub: Use header() function if available
				code := args[0].ToInt()
				httpCtx := ctx.GetHTTPContext()
				if httpCtx != nil {
					httpCtx.SetResponseCode(int(code))
				}
				return values.NewNull(), nil
			},
		},
		// nocache_headers() - Set headers to prevent caching
		{
			Name:       "nocache_headers",
			Parameters: []*registry.Parameter{},
			ReturnType: "void",
			MinArgs:    0,
			MaxArgs:    0,
			IsBuiltin:  true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// Stub: Set cache-control headers
				httpCtx := ctx.GetHTTPContext()
				if httpCtx != nil {
					httpCtx.AddHeader("Cache-Control", "no-cache, must-revalidate, max-age=0", true)
					httpCtx.AddHeader("Pragma", "no-cache", true)
					httpCtx.AddHeader("Expires", "Wed, 11 Jan 1984 05:00:00 GMT", true)
				}
				return values.NewNull(), nil
			},
		},
		// wp_list_pluck() - Pluck a certain field out of each object in a list
		{
			Name: "wp_list_pluck",
			Parameters: []*registry.Parameter{
				{Name: "list", Type: "array"},
				{Name: "field", Type: "string"},
			},
			ReturnType: "array",
			MinArgs:    2,
			MaxArgs:    3,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 || args[0].Type != values.TypeArray {
					return values.NewArray(), nil
				}

				list := args[0].Data.(map[string]*values.Value)
				field := args[1].ToString()
				result := make(map[string]*values.Value)

				for key, item := range list {
					if item.Type == values.TypeArray {
						if itemMap, ok := item.Data.(map[string]*values.Value); ok {
							if val, exists := itemMap[field]; exists {
								result[key] = val
							}
						}
					}
				}

				return &values.Value{Type: values.TypeArray, Data: result}, nil
			},
		},
		// get_language_attributes() - Get language attributes for HTML tag
		{
			Name:       "get_language_attributes",
			Parameters: []*registry.Parameter{},
			ReturnType: "string",
			MinArgs:    0,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// Stub: Return default language attributes
				return values.NewString("lang=\"en-US\" dir=\"ltr\""), nil
			},
		},
		// language_attributes() - Display language attributes for HTML tag
		{
			Name:       "language_attributes",
			Parameters: []*registry.Parameter{},
			ReturnType: "void",
			MinArgs:    0,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// Stub: Echo language attributes
				ctx.WriteOutput(values.NewString("lang=\"en-US\" dir=\"ltr\""))
				return values.NewNull(), nil
			},
		},
		// is_rtl() - Check if current locale is RTL (right-to-left)
		{
			Name:       "is_rtl",
			Parameters: []*registry.Parameter{},
			ReturnType: "bool",
			MinArgs:    0,
			MaxArgs:    0,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// Stub: Return false (LTR)
				return values.NewBool(false), nil
			},
		},
		// wp_parse_str() - WordPress wrapper for parse_str with filters
		// Note: This is a stub that just calls parse_str without actual filtering
		{
			Name: "wp_parse_str",
			Parameters: []*registry.Parameter{
				{Name: "input_string", Type: "string"},
				{Name: "result", Type: "array", IsReference: true},
			},
			ReturnType: "void",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// Stub: For now, just accept the call and do nothing
				// The actual parse_str would have already been called by PHP code
				// We're just here to satisfy function_exists() check
				return values.NewNull(), nil
			},
		},
	}
}
