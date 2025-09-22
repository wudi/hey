package runtime

import (
	"fmt"
	"strings"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// GetStringFunctions returns all string-related PHP functions
func GetStringFunctions() []*registry.Function {
	return []*registry.Function{
		{
			Name:       "strlen",
			Parameters: []*registry.Parameter{{Name: "str", Type: "string"}},
			ReturnType: "int",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewInt(0), nil
				}
				return values.NewInt(int64(len(args[0].ToString()))), nil
			},
		},
		{
			Name: "strpos",
			Parameters: []*registry.Parameter{
				{Name: "haystack", Type: "string"},
				{Name: "needle", Type: "string"},
			},
			ReturnType: "int|false",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 {
					return values.NewBool(false), nil
				}
				haystack := args[0].ToString()
				needle := args[1].ToString()
				idx := strings.Index(haystack, needle)
				if idx == -1 {
					return values.NewBool(false), nil
				}
				return values.NewInt(int64(idx)), nil
			},
		},
		{
			Name: "substr",
			Parameters: []*registry.Parameter{
				{Name: "string", Type: "string"},
				{Name: "offset", Type: "int"},
				{Name: "length", Type: "int", HasDefault: true, DefaultValue: values.NewNull()},
			},
			ReturnType: "string|false",
			MinArgs:    2,
			MaxArgs:    3,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 {
					return values.NewBool(false), nil
				}

				str := args[0].ToString()
				offset := int(args[1].ToInt())
				strLen := len(str)

				// Handle negative offset
				if offset < 0 {
					offset = strLen + offset
					if offset < 0 {
						offset = 0
					}
				}

				// If offset is beyond string length, return false
				if offset >= strLen {
					return values.NewBool(false), nil
				}

				// Determine length
				var length int
				if len(args) >= 3 && !args[2].IsNull() {
					length = int(args[2].ToInt())
				} else {
					length = strLen - offset
				}

				// Handle negative length
				if length < 0 {
					endPos := strLen + length
					if endPos <= offset {
						return values.NewString(""), nil
					}
					length = endPos - offset
				}

				// Ensure we don't go past the end of the string
				if offset+length > strLen {
					length = strLen - offset
				}

				// Extract substring
				if offset < strLen && length > 0 {
					end := offset + length
					if end > strLen {
						end = strLen
					}
					return values.NewString(str[offset:end]), nil
				}

				return values.NewString(""), nil
			},
		},
		{
			Name: "str_repeat",
			Parameters: []*registry.Parameter{
				{Name: "input", Type: "string"},
				{Name: "multiplier", Type: "int"},
			},
			ReturnType: "string",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 {
					return values.NewString(""), nil
				}
				str := args[0].ToString()
				times := int(args[1].ToInt())
				if times <= 0 {
					return values.NewString(""), nil
				}
				return values.NewString(strings.Repeat(str, times)), nil
			},
		},
		{
			Name: "str_replace",
			Parameters: []*registry.Parameter{
				{Name: "search", Type: "mixed"},
				{Name: "replace", Type: "mixed"},
				{Name: "subject", Type: "mixed"},
			},
			ReturnType: "mixed",
			MinArgs:    3,
			MaxArgs:    4,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 3 {
					return values.NewString(""), nil
				}

				search := args[0].ToString()
				replace := args[1].ToString()
				subject := args[2].ToString()

				// PHP str_replace performs simple string replacement
				result := strings.ReplaceAll(subject, search, replace)
				return values.NewString(result), nil
			},
		},
		{
			Name:       "strtolower",
			Parameters: []*registry.Parameter{{Name: "string", Type: "string"}},
			ReturnType: "string",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewString(""), nil
				}
				return values.NewString(strings.ToLower(args[0].ToString())), nil
			},
		},
		{
			Name:       "strtoupper",
			Parameters: []*registry.Parameter{{Name: "string", Type: "string"}},
			ReturnType: "string",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewString(""), nil
				}
				return values.NewString(strings.ToUpper(args[0].ToString())), nil
			},
		},
		{
			Name:       "trim",
			Parameters: []*registry.Parameter{{Name: "string", Type: "string"}},
			ReturnType: "string",
			MinArgs:    1,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewString(""), nil
				}

				str := args[0].ToString()
				// PHP trim removes whitespace by default
				if len(args) == 1 {
					result := strings.TrimSpace(str)
					return values.NewString(result), nil
				}

				// Custom character mask (simplified version)
				mask := args[1].ToString()
				result := strings.Trim(str, mask)
				return values.NewString(result), nil
			},
		},
		{
			Name:       "ltrim",
			Parameters: []*registry.Parameter{{Name: "string", Type: "string"}},
			ReturnType: "string",
			MinArgs:    1,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewString(""), nil
				}

				str := args[0].ToString()
				// Left trim whitespace by default
				if len(args) == 1 {
					result := strings.TrimLeftFunc(str, func(r rune) bool {
						return r == ' ' || r == '\t' || r == '\n' || r == '\r'
					})
					return values.NewString(result), nil
				}

				// Custom character mask
				mask := args[1].ToString()
				result := strings.TrimLeft(str, mask)
				return values.NewString(result), nil
			},
		},
		{
			Name:       "rtrim",
			Parameters: []*registry.Parameter{{Name: "string", Type: "string"}},
			ReturnType: "string",
			MinArgs:    1,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewString(""), nil
				}

				str := args[0].ToString()
				// Right trim whitespace by default
				if len(args) == 1 {
					result := strings.TrimRightFunc(str, func(r rune) bool {
						return r == ' ' || r == '\t' || r == '\n' || r == '\r'
					})
					return values.NewString(result), nil
				}

				// Custom character mask
				mask := args[1].ToString()
				result := strings.TrimRight(str, mask)
				return values.NewString(result), nil
			},
		},
		{
			Name: "explode",
			Parameters: []*registry.Parameter{
				{Name: "delimiter", Type: "string"},
				{Name: "string", Type: "string"},
			},
			ReturnType: "array",
			MinArgs:    2,
			MaxArgs:    3,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 {
					return values.NewArray(), nil
				}

				delimiter := args[0].ToString()
				str := args[1].ToString()

				// Split the string
				parts := strings.Split(str, delimiter)

				// Convert to PHP array
				result := values.NewArray()
				resultArr := result.Data.(*values.Array)
				for i, part := range parts {
					resultArr.Elements[int64(i)] = values.NewString(part)
				}

				return result, nil
			},
		},
		{
			Name: "implode",
			Parameters: []*registry.Parameter{
				{Name: "separator", Type: "string"},
				{Name: "array", Type: "array"},
			},
			ReturnType: "string",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 || args[1] == nil || !args[1].IsArray() {
					return values.NewString(""), nil
				}

				separator := args[0].ToString()
				arr := args[1].Data.(*values.Array)

				// Convert array elements to strings and join
				var parts []string
				for _, value := range arr.Elements {
					if value != nil {
						parts = append(parts, value.ToString())
					}
				}

				result := strings.Join(parts, separator)
				return values.NewString(result), nil
			},
		},
		{
			Name: "sprintf",
			Parameters: []*registry.Parameter{
				{Name: "format", Type: "string"},
			},
			ReturnType: "string",
			IsVariadic: true,
			MinArgs:    1,
			MaxArgs:    -1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewString(""), nil
				}

				format := args[0].ToString()
				var goArgs []interface{}

				// Convert PHP values to Go interface{} for fmt.Sprintf
				for i := 1; i < len(args); i++ {
					if args[i] == nil {
						goArgs = append(goArgs, nil)
					} else if args[i].IsInt() {
						goArgs = append(goArgs, args[i].ToInt())
					} else if args[i].IsFloat() {
						goArgs = append(goArgs, args[i].ToFloat())
					} else if args[i].IsBool() {
						goArgs = append(goArgs, args[i].ToBool())
					} else {
						goArgs = append(goArgs, args[i].ToString())
					}
				}

				// Use Go's fmt.Sprintf which supports similar format specifiers
				result := fmt.Sprintf(format, goArgs...)
				return values.NewString(result), nil
			},
		},
	}
}