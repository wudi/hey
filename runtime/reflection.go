package runtime

import (
	"strings"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// GetReflectionFunctions returns reflection-related PHP functions
func GetReflectionFunctions() []*registry.Function {
	return []*registry.Function{
		{
			Name: "function_exists",
			Parameters: []*registry.Parameter{
				{Name: "function_name", Type: "string"},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    2,
			IsBuiltin: true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewBool(false), nil
				}
				name := args[0].ToString()
				if name == "" {
					return values.NewBool(false), nil
				}
				if fn, ok := ctx.LookupUserFunction(name); ok && fn != nil {
					return values.NewBool(true), nil
				}
				if reg := ctx.SymbolRegistry(); reg != nil {
					if _, ok := reg.GetFunction(name); ok {
						return values.NewBool(true), nil
					}
				}
				return values.NewBool(false), nil
			},
		},
		{
			Name: "class_exists",
			Parameters: []*registry.Parameter{
				{Name: "class_name", Type: "string"},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    2,
			IsBuiltin: true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewBool(false), nil
				}
				name := args[0].ToString()
				if name == "" {
					return values.NewBool(false), nil
				}
				if _, ok := ctx.LookupUserClass(name); ok {
					return values.NewBool(true), nil
				}
				if reg := ctx.SymbolRegistry(); reg != nil {
					if _, err := reg.GetClass(name); err == nil {
						return values.NewBool(true), nil
					}
				}
				// Note: builtinClassStubs check removed - should be handled by builtins.go
				return values.NewBool(false), nil
			},
		},
		{
			Name: "get_class",
			Parameters: []*registry.Parameter{
				{Name: "object", Type: "object"},
			},
			ReturnType: "string",
			MinArgs:    0,
			MaxArgs:    1,
			IsBuiltin: true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 {
					// Called without arguments - return current class name if in class context
					// For now, return false (which would be an error in real PHP)
					return values.NewBool(false), nil
				}

				if args[0] == nil || !args[0].IsObject() {
					return values.NewBool(false), nil
				}

				obj := args[0].Data.(*values.Object)
				return values.NewString(obj.ClassName), nil
			},
		},
		{
			Name: "method_exists",
			Parameters: []*registry.Parameter{
				{Name: "object_or_class", Type: "mixed"},
				{Name: "method_name", Type: "string"},
			},
			ReturnType: "bool",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin: true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 || args[1] == nil {
					return values.NewBool(false), nil
				}
				methodName := strings.ToLower(args[1].ToString())
				if methodName == "" {
					return values.NewBool(false), nil
				}
				var className string
				if args[0] != nil && args[0].IsObject() {
					className = args[0].Data.(*values.Object).ClassName
				} else if args[0] != nil {
					className = args[0].ToString()
				}
				if className == "" {
					return values.NewBool(false), nil
				}
				if classInfo, ok := ctx.LookupUserClass(className); ok && classInfo != nil {
					for name := range classInfo.Methods {
						if strings.ToLower(name) == methodName {
							return values.NewBool(true), nil
						}
					}
				}
				if reg := ctx.SymbolRegistry(); reg != nil {
					if desc, err := reg.GetClass(className); err == nil && desc != nil {
						for name := range desc.Methods {
							if strings.ToLower(name) == methodName {
								return values.NewBool(true), nil
							}
						}
					}
				}
				// Note: builtinClassStubs check removed - should be handled by builtins.go
				return values.NewBool(false), nil
			},
		},
		{
			Name: "func_num_args",
			Parameters: []*registry.Parameter{},
			ReturnType: "int",
			MinArgs:    0,
			MaxArgs:    0,
			IsBuiltin: true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// func_num_args() returns the number of arguments passed to the calling function
				// We need to access the calling function's arguments
				// For now, return -1 to indicate it's not implemented correctly
				return values.NewInt(-1), nil
			},
		},
		{
			Name: "func_get_args",
			Parameters: []*registry.Parameter{},
			ReturnType: "array",
			MinArgs:    0,
			MaxArgs:    0,
			IsBuiltin: true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// func_get_args() returns an array of all arguments passed to the calling function
				// For now, return empty array
				return values.NewArray(), nil
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
			IsBuiltin: true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// func_get_arg(n) returns the nth argument passed to the calling function
				// For now, return null
				return values.NewNull(), nil
			},
		},

		// property_exists - Check if object has property
		{
			Name: "property_exists",
			Parameters: []*registry.Parameter{
				{Name: "class", Type: "mixed"},
				{Name: "property", Type: "string"},
			},
			ReturnType: "bool",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin: true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 {
					return values.NewBool(false), nil
				}

				propertyName := args[1].ToString()

				// Handle object input
				if args[0].IsObject() {
					obj := args[0]

					// Check if object has the property
					if obj.Data.(*values.Object) != nil {
						objData := obj.Data.(*values.Object)
						_, exists := objData.Properties[propertyName]
						return values.NewBool(exists), nil
					}
				}

				// Handle class string input (simplified - just return false for now)
				// In full implementation, would check class definition
				return values.NewBool(false), nil
			},
		},

		// is_a - Check object type
		{
			Name: "is_a",
			Parameters: []*registry.Parameter{
				{Name: "object", Type: "mixed"},
				{Name: "class_name", Type: "string"},
				{Name: "allow_string", Type: "bool", HasDefault: true, DefaultValue: values.NewBool(false)},
			},
			ReturnType: "bool",
			MinArgs:    2,
			MaxArgs:    3,
			IsBuiltin: true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 {
					return values.NewBool(false), nil
				}

				targetClassName := args[1].ToString()
				allowString := false
				if len(args) >= 3 {
					allowString = args[2].ToBool()
				}

				// Handle string class name input
				if args[0].IsString() {
					if allowString {
						className := args[0].ToString()
						return values.NewBool(className == targetClassName), nil
					}
					return values.NewBool(false), nil
				}

				// Handle object input
				if args[0].IsObject() {
					obj := args[0]
					if objData := obj.Data.(*values.Object); objData != nil {
						return values.NewBool(objData.ClassName == targetClassName), nil
					}
				}

				return values.NewBool(false), nil
			},
		},
		{
			Name: "get_class_methods",
			Parameters: []*registry.Parameter{
				{Name: "class_name_or_object", Type: "mixed"},
			},
			ReturnType: "array|false",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin: true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewBool(false), nil
				}

				var className string

				// If argument is an object, get its class name
				if args[0].IsObject() {
					obj := args[0].Data.(*values.Object)
					className = obj.ClassName
				} else {
					// Otherwise treat as class name string
					className = args[0].ToString()
				}

				if className == "" {
					return values.NewBool(false), nil
				}

				// Look up the class in registry
				reg := ctx.SymbolRegistry()
				if reg == nil {
					return values.NewBool(false), nil
				}

				classDesc, err := reg.GetClass(className)
				if err != nil || classDesc == nil {
					return values.NewBool(false), nil
				}

				// Collect all method names
				methods := values.NewArray()
				index := int64(0)

				for methodName := range classDesc.Methods {
					methods.ArraySet(values.NewInt(index), values.NewString(methodName))
					index++
				}

				return methods, nil
			},
		},
	}
}

