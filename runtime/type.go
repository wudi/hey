package runtime

import (
	"strconv"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// GetTypeFunctions returns all type-checking functions
func GetTypeFunctions() []*registry.Function {
	return []*registry.Function{
		{
			Name:       "is_string",
			Parameters: []*registry.Parameter{{Name: "value", Type: "mixed"}},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewBool(false), nil
				}
				return values.NewBool(args[0].IsString()), nil
			},
		},
		{
			Name:       "is_int",
			Parameters: []*registry.Parameter{{Name: "value", Type: "mixed"}},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewBool(false), nil
				}
				return values.NewBool(args[0].IsInt()), nil
			},
		},
		{
			Name:       "is_array",
			Parameters: []*registry.Parameter{{Name: "value", Type: "mixed"}},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewBool(false), nil
				}
				return values.NewBool(args[0].IsArray()), nil
			},
		},
		{
			Name:       "is_object",
			Parameters: []*registry.Parameter{{Name: "value", Type: "mixed"}},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewBool(false), nil
				}
				return values.NewBool(args[0].IsObject()), nil
			},
		},
		{
			Name:       "is_bool",
			Parameters: []*registry.Parameter{{Name: "value", Type: "mixed"}},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewBool(false), nil
				}
				return values.NewBool(args[0].IsBool()), nil
			},
		},
		{
			Name:       "is_float",
			Parameters: []*registry.Parameter{{Name: "value", Type: "mixed"}},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewBool(false), nil
				}
				return values.NewBool(args[0].IsFloat()), nil
			},
		},
		{
			Name:       "is_null",
			Parameters: []*registry.Parameter{{Name: "value", Type: "mixed"}},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewBool(true), nil
				}
				return values.NewBool(args[0].IsNull()), nil
			},
		},
		{
			Name:       "is_numeric",
			Parameters: []*registry.Parameter{{Name: "value", Type: "mixed"}},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewBool(false), nil
				}
				// is_numeric returns true for numbers and numeric strings
				if args[0].IsInt() || args[0].IsFloat() {
					return values.NewBool(true), nil
				}
				if args[0].IsString() {
					str := args[0].ToString()
					_, err := strconv.ParseFloat(str, 64)
					return values.NewBool(err == nil), nil
				}
				return values.NewBool(false), nil
			},
		},
		{
			Name: "gettype",
			Parameters: []*registry.Parameter{
				{Name: "value", Type: "mixed"},
			},
			ReturnType: "string",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 {
					return values.NewString("NULL"), nil
				}

				val := args[0]
				if val == nil {
					return values.NewString("NULL"), nil
				}

				// Return PHP type names
				switch val.Type {
				case values.TypeNull:
					return values.NewString("NULL"), nil
				case values.TypeBool:
					return values.NewString("boolean"), nil
				case values.TypeInt:
					return values.NewString("integer"), nil
				case values.TypeFloat:
					return values.NewString("double"), nil
				case values.TypeString:
					return values.NewString("string"), nil
				case values.TypeArray:
					return values.NewString("array"), nil
				case values.TypeObject:
					return values.NewString("object"), nil
				case values.TypeResource:
					return values.NewString("resource"), nil
				default:
					return values.NewString("unknown type"), nil
				}
			},
		},
		{
			Name:       "is_resource",
			Parameters: []*registry.Parameter{{Name: "value", Type: "mixed"}},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewBool(false), nil
				}
				// In Hey-Codex, resources are represented as integers
				return values.NewBool(args[0].Type == values.TypeResource), nil
			},
		},
		{
			Name:       "is_scalar",
			Parameters: []*registry.Parameter{{Name: "value", Type: "mixed"}},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewBool(false), nil
				}
				// Scalars are: string, int, float, bool (not null, array, object, resource)
				return values.NewBool(
					args[0].IsString() ||
					args[0].IsInt() ||
					args[0].IsFloat() ||
					args[0].IsBool(),
				), nil
			},
		},
		{
			Name: "settype",
			Parameters: []*registry.Parameter{
				{Name: "var", Type: "mixed"},
				{Name: "type", Type: "string"},
			},
			ReturnType: "bool",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 {
					return values.NewBool(false), nil
				}

				variable := args[0]
				typeName := args[1].ToString()

				// In a real implementation, settype() would modify the variable in place
				// This simplified version creates a new value with the converted type
				switch typeName {
				case "boolean", "bool":
					boolVal := variable.ToBool()
					variable.Type = values.TypeBool
					variable.Data = boolVal
				case "integer", "int":
					intVal := variable.ToInt()
					variable.Type = values.TypeInt
					variable.Data = intVal
				case "float", "double":
					floatVal := variable.ToFloat()
					variable.Type = values.TypeFloat
					variable.Data = floatVal
				case "string":
					strVal := variable.ToString()
					variable.Type = values.TypeString
					variable.Data = strVal
				case "array":
					if variable.Type != values.TypeArray {
						// Convert non-array to array with single element
						newArray := values.NewArray()
						newArray.ArraySet(values.NewInt(0), variable)
						variable.Type = values.TypeArray
						variable.Data = newArray.Data
					}
				case "null":
					variable.Type = values.TypeNull
					variable.Data = nil
				default:
					return values.NewBool(false), nil
				}

				return values.NewBool(true), nil
			},
		},
	}
}