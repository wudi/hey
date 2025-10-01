package runtime

import (
	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// GetMySQLiFunctions returns stub implementations of MySQLi functions
// These are minimal stubs to allow WordPress to detect mysqli support
func GetMySQLiFunctions() []*registry.Function {
	return []*registry.Function{
		// mysqli_connect() - Open a new connection to the MySQL server
		{
			Name: "mysqli_connect",
			Parameters: []*registry.Parameter{
				{Name: "hostname", Type: "string", DefaultValue: values.NewString("localhost")},
				{Name: "username", Type: "string", DefaultValue: values.NewString("")},
				{Name: "password", Type: "string", DefaultValue: values.NewString("")},
				{Name: "database", Type: "string", DefaultValue: values.NewString("")},
				{Name: "port", Type: "int", DefaultValue: values.NewInt(3306)},
				{Name: "socket", Type: "string", DefaultValue: values.NewString("")},
			},
			ReturnType: "resource",
			MinArgs:    0,
			MaxArgs:    6,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// Stub: Return a fake resource to indicate mysqli is available
				// Real implementation would connect to MySQL
				return values.NewResource("mysqli_link"), nil
			},
		},
		// mysqli_close() - Close a previously opened database connection
		{
			Name: "mysqli_close",
			Parameters: []*registry.Parameter{
				{Name: "mysql", Type: "resource"},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewBool(false), nil
				}
				return values.NewBool(true), nil
			},
		},
		// mysqli_query() - Perform a query on the database
		{
			Name: "mysqli_query",
			Parameters: []*registry.Parameter{
				{Name: "mysql", Type: "resource"},
				{Name: "query", Type: "string"},
			},
			ReturnType: "mixed",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 {
					return values.NewBool(false), nil
				}
				// Stub: Return empty result set
				return values.NewResource("mysqli_result"), nil
			},
		},
		// mysqli_error() - Return a description of the last error
		{
			Name: "mysqli_error",
			Parameters: []*registry.Parameter{
				{Name: "mysql", Type: "resource"},
			},
			ReturnType: "string",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				return values.NewString(""), nil
			},
		},
		// mysqli_errno() - Return the error code for the most recent function call
		{
			Name: "mysqli_errno",
			Parameters: []*registry.Parameter{
				{Name: "mysql", Type: "resource"},
			},
			ReturnType: "int",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				return values.NewInt(0), nil
			},
		},
		// mysqli_fetch_assoc() - Fetch the next row of a result set as an associative array
		{
			Name: "mysqli_fetch_assoc",
			Parameters: []*registry.Parameter{
				{Name: "result", Type: "resource"},
			},
			ReturnType: "array",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// Stub: Return null (no more rows)
				return values.NewNull(), nil
			},
		},
		// mysqli_free_result() - Free the memory associated with a result
		{
			Name: "mysqli_free_result",
			Parameters: []*registry.Parameter{
				{Name: "result", Type: "resource"},
			},
			ReturnType: "void",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				return values.NewNull(), nil
			},
		},
	}
}
