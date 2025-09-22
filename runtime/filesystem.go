package runtime

import (
	"os"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// GetFilesystemFunctions returns all file system related functions
func GetFilesystemFunctions() []*registry.Function {
	return []*registry.Function{
		{
			Name: "file_get_contents",
			Parameters: []*registry.Parameter{
				{Name: "filename", Type: "string"},
			},
			ReturnType: "string|false",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewBool(false), nil
				}
				filename := args[0].ToString()

				content, err := os.ReadFile(filename)
				if err != nil {
					return values.NewBool(false), nil
				}

				return values.NewString(string(content)), nil
			},
		},
		{
			Name: "file_put_contents",
			Parameters: []*registry.Parameter{
				{Name: "filename", Type: "string"},
				{Name: "data", Type: "mixed"},
			},
			ReturnType: "int|false",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 || args[0] == nil || args[1] == nil {
					return values.NewBool(false), nil
				}

				filename := args[0].ToString()
				data := args[1].ToString()

				err := os.WriteFile(filename, []byte(data), 0644)
				if err != nil {
					return values.NewBool(false), nil
				}

				return values.NewInt(int64(len(data))), nil
			},
		},
		{
			Name: "unlink",
			Parameters: []*registry.Parameter{
				{Name: "filename", Type: "string"},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewBool(false), nil
				}
				filename := args[0].ToString()

				err := os.Remove(filename)
				if err != nil {
					return values.NewBool(false), nil
				}

				return values.NewBool(true), nil
			},
		},
	}
}