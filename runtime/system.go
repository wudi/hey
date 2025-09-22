package runtime

import (
	"os"
	"strings"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// GetSystemFunctions returns system-related PHP functions
func GetSystemFunctions() []*registry.Function {
	return []*registry.Function{
		{
			Name: "getenv",
			Parameters: []*registry.Parameter{
				{Name: "varname", Type: "string", HasDefault: true, DefaultValue: values.NewNull()},
			},
			ReturnType: "string|array|false",
			MinArgs:    0,
			MaxArgs:    1,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || (len(args) > 0 && args[0].IsNull()) {
					// Return all environment variables as an associative array
					envVars := os.Environ()
					result := values.NewArray()
					arr := result.Data.(*values.Array)

					for _, env := range envVars {
						parts := strings.SplitN(env, "=", 2)
						if len(parts) == 2 {
							arr.Elements[parts[0]] = values.NewString(parts[1])
						}
					}
					return result, nil
				}

				// Get specific environment variable
				varname := args[0].ToString()
				value, exists := os.LookupEnv(varname)

				if !exists {
					return values.NewBool(false), nil
				}

				return values.NewString(value), nil
			},
		},
		{
			Name: "putenv",
			Parameters: []*registry.Parameter{
				{Name: "setting", Type: "string"},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 {
					return values.NewBool(false), nil
				}

				setting := args[0].ToString()

				// Parse the setting string (format: "NAME=value")
				parts := strings.SplitN(setting, "=", 2)
				if len(parts) != 2 {
					return values.NewBool(false), nil
				}

				name := parts[0]
				value := parts[1]

				// Set the environment variable
				err := os.Setenv(name, value)
				if err != nil {
					return values.NewBool(false), nil
				}

				return values.NewBool(true), nil
			},
		},
		{
			Name:       "getcwd",
			Parameters: []*registry.Parameter{},
			ReturnType: "string|false",
			MinArgs:    0,
			MaxArgs:    0,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// Get the current working directory
				cwd, err := os.Getwd()
				if err != nil {
					// Return false on error (PHP behavior)
					return values.NewBool(false), nil
				}
				return values.NewString(cwd), nil
			},
		},
		{
			Name: "exit",
			Parameters: []*registry.Parameter{
				{Name: "status", Type: "mixed", HasDefault: true, DefaultValue: values.NewInt(0)},
			},
			ReturnType: "void",
			MinArgs:    0,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				var exitCode int
				var message string

				if len(args) > 0 && args[0] != nil {
					if args[0].IsString() {
						// String argument: print message and exit with code 0
						message = args[0].ToString()
						exitCode = 0
					} else {
						// Numeric argument: exit with this code
						exitCode = int(args[0].ToInt())
					}
				} else {
					// No argument: exit with code 0
					exitCode = 0
				}

				// Halt execution
				ctx.Halt(exitCode, message)
				return values.NewNull(), nil
			},
		},
		{
			Name: "die",
			Parameters: []*registry.Parameter{
				{Name: "status", Type: "mixed", HasDefault: true, DefaultValue: values.NewInt(0)},
			},
			ReturnType: "void",
			MinArgs:    0,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				var exitCode int
				var message string

				if len(args) > 0 && args[0] != nil {
					if args[0].IsString() {
						// String argument: print message and exit with code 0
						message = args[0].ToString()
						exitCode = 0
					} else {
						// Numeric argument: exit with this code
						exitCode = int(args[0].ToInt())
					}
				} else {
					// No argument: exit with code 0
					exitCode = 0
				}

				// Halt execution (die is an alias of exit)
				ctx.Halt(exitCode, message)
				return values.NewNull(), nil
			},
		},
	}
}