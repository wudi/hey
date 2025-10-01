package runtime

import (
	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// GetMySQLiAdvancedFunctions returns advanced/debug procedural functions
func GetMySQLiAdvancedFunctions() []*registry.Function {
	return []*registry.Function{
		// Debug functions
		{
			Name: "mysqli_dump_debug_info",
			Parameters: []*registry.Parameter{
				{Name: "mysql", Type: "object"},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				return values.NewBool(true), nil
			},
		},
		{
			Name: "mysqli_debug",
			Parameters: []*registry.Parameter{
				{Name: "options", Type: "string"},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				return values.NewBool(true), nil
			},
		},

		// Statistics functions
		{
			Name:       "mysqli_get_cache_stats",
			Parameters: []*registry.Parameter{},
			ReturnType: "array",
			MinArgs:    0,
			MaxArgs:    0,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				return values.NewArray(), nil
			},
		},
		{
			Name:       "mysqli_get_client_stats",
			Parameters: []*registry.Parameter{},
			ReturnType: "array",
			MinArgs:    0,
			MaxArgs:    0,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				return values.NewArray(), nil
			},
		},
		{
			Name: "mysqli_get_connection_stats",
			Parameters: []*registry.Parameter{
				{Name: "mysql", Type: "object"},
			},
			ReturnType: "array",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				return values.NewArray(), nil
			},
		},
		{
			Name:       "mysqli_get_links_stats",
			Parameters: []*registry.Parameter{},
			ReturnType: "array",
			MinArgs:    0,
			MaxArgs:    0,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				return values.NewArray(), nil
			},
		},

		// Connection management
		{
			Name: "mysqli_kill",
			Parameters: []*registry.Parameter{
				{Name: "mysql", Type: "object"},
				{Name: "process_id", Type: "int"},
			},
			ReturnType: "bool",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				return values.NewBool(true), nil
			},
		},
		{
			Name: "mysqli_refresh",
			Parameters: []*registry.Parameter{
				{Name: "mysql", Type: "object"},
				{Name: "flags", Type: "int"},
			},
			ReturnType: "bool",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				return values.NewBool(true), nil
			},
		},

		// Error reporting
		{
			Name: "mysqli_report",
			Parameters: []*registry.Parameter{
				{Name: "flags", Type: "int"},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				return values.NewBool(true), nil
			},
		},

		// Local infile handlers
		{
			Name: "mysqli_set_local_infile_default",
			Parameters: []*registry.Parameter{
				{Name: "mysql", Type: "object"},
			},
			ReturnType: "void",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				return values.NewNull(), nil
			},
		},
		{
			Name: "mysqli_set_local_infile_handler",
			Parameters: []*registry.Parameter{
				{Name: "mysql", Type: "object"},
				{Name: "callback", Type: "callable"},
			},
			ReturnType: "bool",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				return values.NewBool(true), nil
			},
		},

		// SSL configuration
		{
			Name: "mysqli_ssl_set",
			Parameters: []*registry.Parameter{
				{Name: "mysql", Type: "object"},
				{Name: "key", Type: "string"},
				{Name: "cert", Type: "string"},
				{Name: "ca", Type: "string"},
				{Name: "capath", Type: "string"},
				{Name: "cipher", Type: "string"},
			},
			ReturnType: "bool",
			MinArgs:    6,
			MaxArgs:    6,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				return values.NewBool(true), nil
			},
		},

		// Warnings
		{
			Name: "mysqli_stmt_get_warnings",
			Parameters: []*registry.Parameter{
				{Name: "statement", Type: "object"},
			},
			ReturnType: "object|false",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				return values.NewBool(false), nil
			},
		},
		{
			Name: "mysqli_get_warnings",
			Parameters: []*registry.Parameter{
				{Name: "mysql", Type: "object"},
			},
			ReturnType: "object|false",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				return values.NewBool(false), nil
			},
		},
	}
}
