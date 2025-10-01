package runtime

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

var supportedExtensions = map[string]bool{
	"standard": true,
	"core":     true,
	"date":     true,
	"pcre":     true,
	"json":     true,
	"mbstring": true,
	"ctype":    true,
	"hash":     true, // We support hash functions (md5, sha1, sha256, etc.)
}

// GetVariableFunctions returns variable-related PHP functions
func GetVariableFunctions() []*registry.Function {
	return []*registry.Function{
		{
			Name: "isset",
			Parameters: []*registry.Parameter{
				{Name: "value", Type: "mixed"},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    999, // PHP's isset() can take multiple arguments
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// In PHP, isset() returns true only if all arguments are set and not null
				for _, arg := range args {
					if arg == nil || arg.IsNull() {
						return values.NewBool(false), nil
					}
				}
				return values.NewBool(true), nil
			},
		},
		{
			Name: "define",
			Parameters: []*registry.Parameter{
				{Name: "name", Type: "string"},
				{Name: "value", Type: "mixed"},
				{Name: "case_insensitive", Type: "bool", HasDefault: true, DefaultValue: values.NewBool(false)},
			},
			ReturnType: "bool",
			MinArgs:    2,
			MaxArgs:    3,
			IsBuiltin: true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 {
					return values.NewBool(false), nil
				}

				name := args[0].ToString()
				value := args[1]
				// Third parameter (case_insensitive) is deprecated in modern PHP but we accept it

				if name == "" {
					return values.NewBool(false), nil
				}

				// Check if constant already exists
				reg := ctx.SymbolRegistry()
				if _, exists := reg.GetConstant(name); exists {
					return values.NewBool(false), nil
				}

				// Create and register the constant
				constantDesc := &registry.ConstantDescriptor{
					Name:       name,
					Visibility: "public",
					Value:      value,
					IsFinal:    true,
				}

				err := reg.RegisterConstant(constantDesc)
				if err != nil {
					return values.NewBool(false), nil
				}

				return values.NewBool(true), nil
			},
		},
		{
			Name: "defined",
			Parameters: []*registry.Parameter{
				{Name: "name", Type: "string"},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin: true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 {
					return values.NewBool(false), nil
				}

				name := args[0].ToString()
				reg := ctx.SymbolRegistry()
				_, exists := reg.GetConstant(name)

				return values.NewBool(exists), nil
			},
		},

		// constant - Get constant value
		{
			Name: "constant",
			Parameters: []*registry.Parameter{
				{Name: "name", Type: "string"},
			},
			ReturnType: "mixed",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 {
					return values.NewNull(), nil
				}

				name := args[0].ToString()
				if reg := ctx.SymbolRegistry(); reg != nil {
					if constant, exists := reg.GetConstant(name); exists {
						return constant.Value, nil
					}
				}

				// Return null for undefined constant (PHP behavior)
				return values.NewNull(), nil
			},
		},

		// get_defined_vars - Get all defined variables
		{
			Name: "get_defined_vars",
			Parameters: []*registry.Parameter{},
			ReturnType: "array",
			MinArgs:    0,
			MaxArgs:    0,
			IsBuiltin:  true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// For now, return empty array - getting current scope variables
				// requires deeper integration with VM execution context
				// In real PHP, this would return local variables from current scope
				return values.NewArray(), nil
			},
		},

		// get_defined_constants - Get all defined constants
		{
			Name: "get_defined_constants",
			Parameters: []*registry.Parameter{
				{Name: "categorize", Type: "bool", HasDefault: true, DefaultValue: values.NewBool(false)},
			},
			ReturnType: "array",
			MinArgs:    0,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				categorize := false
				if len(args) > 0 {
					categorize = args[0].ToBool()
				}

				if reg := ctx.SymbolRegistry(); reg != nil {
					constants := reg.GetAllConstants()

					if categorize {
						// Group constants by category
						result := values.NewArray()
						coreGroup := values.NewArray()

						for _, constant := range constants {
							coreGroup.ArraySet(values.NewString(constant.Name), constant.Value)
						}

						result.ArraySet(values.NewString("Core"), coreGroup)
						return result, nil
					} else {
						// Return flat array of constants
						result := values.NewArray()
						for _, constant := range constants {
							result.ArraySet(values.NewString(constant.Name), constant.Value)
						}
						return result, nil
					}
				}

				return values.NewArray(), nil
			},
		},
		{
			Name: "empty",
			Parameters: []*registry.Parameter{
				{Name: "value", Type: "mixed"},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 {
					return values.NewBool(true), nil
				}

				value := args[0]
				if value == nil || value.IsNull() {
					return values.NewBool(true), nil
				}

				// PHP empty() rules:
				// - false, 0, 0.0, "", "0", array(), null return true
				// - everything else returns false
				switch value.Type {
				case values.TypeBool:
					return values.NewBool(!value.ToBool()), nil
				case values.TypeInt:
					return values.NewBool(value.ToInt() == 0), nil
				case values.TypeFloat:
					return values.NewBool(value.ToFloat() == 0.0), nil
				case values.TypeString:
					str := value.ToString()
					return values.NewBool(str == "" || str == "0"), nil
				case values.TypeArray:
					arr := value.Data.(*values.Array)
					return values.NewBool(len(arr.Elements) == 0), nil
				case values.TypeNull:
					return values.NewBool(true), nil
				default:
					return values.NewBool(false), nil
				}
			},
		},
		{
			Name: "var_export",
			Parameters: []*registry.Parameter{
				{Name: "expression", Type: "mixed"},
				{Name: "return", Type: "bool", HasDefault: true, DefaultValue: values.NewBool(false)},
			},
			ReturnType: "mixed",
			MinArgs:    1,
			MaxArgs:    2,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 {
					return values.NewNull(), nil
				}

				value := args[0]
				returnResult := false
				if len(args) > 1 && args[1] != nil {
					returnResult = args[1].ToBool()
				}

				// Generate PHP-like representation of the value
				export := varExportValue(value, 0)

				if returnResult {
					return values.NewString(export), nil
				} else {
					// Print to output
					fmt.Print(export)
					return values.NewNull(), nil
				}
			},
		},
		{
			Name: "unset",
			Parameters: []*registry.Parameter{
				{Name: "var", Type: "mixed"},
			},
			ReturnType: "void",
			MinArgs:    1,
			MaxArgs:    999, // PHP's unset() can take multiple arguments
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// In a full implementation, unset() would need VM support to actually unset variables
				// For now, we'll return null as unset() returns void/null
				// TODO: Implement proper variable unsetting with VM support
				return values.NewNull(), nil
			},
		},
		{
			Name: "serialize",
			Parameters: []*registry.Parameter{
				{Name: "value", Type: "mixed"},
			},
			ReturnType: "string",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 {
					return values.NewString(""), nil
				}

				value := args[0]
				serialized := serializeValue(value)
				return values.NewString(serialized), nil
			},
		},
		{
			Name: "unserialize",
			Parameters: []*registry.Parameter{
				{Name: "data", Type: "string"},
			},
			ReturnType: "mixed",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 {
					return values.NewBool(false), nil
				}

				data := args[0].ToString()
				value, err := unserializeValue(data)
				if err != nil {
					// Return false on error, as per PHP behavior
					return values.NewBool(false), nil
				}
				return value, nil
			},
		},
		{
			Name: "is_callable",
			Parameters: []*registry.Parameter{
				{Name: "value", Type: "mixed"},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 {
					return values.NewBool(false), nil
				}

				value := args[0]
				if value == nil || value.IsNull() {
					return values.NewBool(false), nil
				}

				switch value.Type {
				case values.TypeString:
					funcName := value.ToString()
					if funcName == "" {
						return values.NewBool(false), nil
					}

					if reg := ctx.SymbolRegistry(); reg != nil {
						if _, exists := reg.GetFunction(funcName); exists {
							return values.NewBool(true), nil
						}
					}

					if _, exists := ctx.LookupUserFunction(funcName); exists {
						return values.NewBool(true), nil
					}

					return values.NewBool(false), nil

				case values.TypeArray:
					arr := value.Data.(*values.Array)
					if len(arr.Elements) != 2 {
						return values.NewBool(false), nil
					}

					classNameVal := value.ArrayGet(values.NewInt(0))
					methodNameVal := value.ArrayGet(values.NewInt(1))

					if classNameVal.IsNull() || methodNameVal.IsNull() {
						return values.NewBool(false), nil
					}

					if classNameVal.Type != values.TypeString || methodNameVal.Type != values.TypeString {
						return values.NewBool(false), nil
					}

					className := classNameVal.ToString()
					methodName := methodNameVal.ToString()

					if classMap, exists := ctx.LookupUserClass(className); exists {
						if _, methodExists := classMap.Methods[methodName]; methodExists {
							return values.NewBool(true), nil
						}
					}

					return values.NewBool(false), nil

				default:
					return values.NewBool(false), nil
				}
			},
		},
		{
			Name: "extension_loaded",
			Parameters: []*registry.Parameter{
				{Name: "extension", Type: "string"},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 {
					return values.NewBool(false), nil
				}

				extName := args[0].ToString()
				if extName == "" {
					return values.NewBool(false), nil
				}

				extNameLower := strings.ToLower(extName)
				return values.NewBool(supportedExtensions[extNameLower]), nil
			},
		},
	}
}

// varExportValue generates a PHP-like string representation of a value
func varExportValue(value *values.Value, indent int) string {
	if value == nil {
		return "NULL"
	}

	switch value.Type {
	case values.TypeNull:
		return "NULL"
	case values.TypeBool:
		if value.ToBool() {
			return "true"
		}
		return "false"
	case values.TypeInt:
		return fmt.Sprintf("%d", value.ToInt())
	case values.TypeFloat:
		return fmt.Sprintf("%g", value.ToFloat())
	case values.TypeString:
		// Escape string and wrap in single quotes
		str := value.ToString()
		escaped := strings.ReplaceAll(str, "'", "\\'")
		escaped = strings.ReplaceAll(escaped, "\\", "\\\\")
		return fmt.Sprintf("'%s'", escaped)
	case values.TypeArray:
		arr := value.Data.(*values.Array)
		if len(arr.Elements) == 0 {
			return "array ()"
		}

		result := "array (\n"
		for key, element := range arr.Elements {
			indentStr := strings.Repeat("  ", indent+1)
			result += fmt.Sprintf("%s%v => %s,\n", indentStr, key, varExportValue(element, indent+1))
		}
		indentStr := strings.Repeat("  ", indent)
		result += indentStr + ")"
		return result
	default:
		return "NULL"
	}
}

// serializeValue generates a PHP serialize format string for a value
func serializeValue(value *values.Value) string {
	if value == nil {
		return "N;"
	}

	switch value.Type {
	case values.TypeNull:
		return "N;"
	case values.TypeBool:
		if value.ToBool() {
			return "b:1;"
		}
		return "b:0;"
	case values.TypeInt:
		return fmt.Sprintf("i:%d;", value.ToInt())
	case values.TypeFloat:
		return fmt.Sprintf("d:%g;", value.ToFloat())
	case values.TypeString:
		str := value.ToString()
		return fmt.Sprintf("s:%d:\"%s\";", len(str), str)
	case values.TypeArray:
		arr := value.Data.(*values.Array)
		result := fmt.Sprintf("a:%d:{", len(arr.Elements))
		for key, element := range arr.Elements {
			// Serialize key
			if keyInt, ok := key.(int64); ok {
				result += fmt.Sprintf("i:%d;", keyInt)
			} else {
				keyStr := fmt.Sprintf("%v", key)
				result += fmt.Sprintf("s:%d:\"%s\";", len(keyStr), keyStr)
			}
			// Serialize value
			result += serializeValue(element)
		}
		result += "}"
		return result
	default:
		return "N;"
	}
}

// unserializeValue parses a PHP serialize format string back to a value
// This is a simplified implementation that handles basic types
func unserializeValue(data string) (*values.Value, error) {
	if len(data) < 2 {
		return nil, fmt.Errorf("invalid serialized data")
	}

	switch data[0] {
	case 'N':
		if data == "N;" {
			return values.NewNull(), nil
		}
		return nil, fmt.Errorf("invalid null format")
	case 'b':
		if len(data) >= 4 && data[1] == ':' && data[3] == ';' {
			if data[2] == '1' {
				return values.NewBool(true), nil
			} else if data[2] == '0' {
				return values.NewBool(false), nil
			}
		}
		return nil, fmt.Errorf("invalid bool format")
	case 'i':
		// i:123;
		if len(data) >= 4 && data[1] == ':' {
			semicolonPos := strings.Index(data[2:], ";")
			if semicolonPos != -1 {
				numStr := data[2 : 2+semicolonPos]
				if num, err := strconv.ParseInt(numStr, 10, 64); err == nil {
					return values.NewInt(num), nil
				}
			}
		}
		return nil, fmt.Errorf("invalid int format")
	case 'd':
		// d:1.5;
		if len(data) >= 4 && data[1] == ':' {
			semicolonPos := strings.Index(data[2:], ";")
			if semicolonPos != -1 {
				numStr := data[2 : 2+semicolonPos]
				if num, err := strconv.ParseFloat(numStr, 64); err == nil {
					return values.NewFloat(num), nil
				}
			}
		}
		return nil, fmt.Errorf("invalid float format")
	case 's':
		// s:5:"hello";
		if len(data) >= 6 && data[1] == ':' {
			colonPos := strings.Index(data[2:], ":")
			if colonPos != -1 {
				lengthStr := data[2 : 2+colonPos]
				if length, err := strconv.Atoi(lengthStr); err == nil {
					startPos := 2 + colonPos + 2 // s:length:"
					if len(data) >= startPos+length+2 { // +2 for closing ";
						str := data[startPos : startPos+length]
						return values.NewString(str), nil
					}
				}
			}
		}
		return nil, fmt.Errorf("invalid string format")
	case 'a':
		// Simple array unserialization - a:2:{i:0;s:5:"hello";i:1;s:5:"world";}
		// For now, return false for complex array deserialization
		return values.NewBool(false), nil
	default:
		return nil, fmt.Errorf("unsupported serialized type: %c", data[0])
	}
}