package runtime

import (
	"encoding/json"
	"fmt"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// isSequentialArray checks if a PHP array is sequential (list-like)
// A sequential array has integer keys starting from 0 with no gaps
func isSequentialArray(arr *values.Array) bool {
	if len(arr.Elements) == 0 {
		return true // Empty arrays are considered sequential
	}

	// Check if all keys are integers and sequential from 0
	for i := 0; i < len(arr.Elements); i++ {
		if _, exists := arr.Elements[int64(i)]; !exists {
			return false
		}
	}

	// Check there are no additional non-integer keys
	for key := range arr.Elements {
		if _, ok := key.(int64); !ok {
			return false
		}
		intKey := key.(int64)
		if intKey < 0 || intKey >= int64(len(arr.Elements)) {
			return false
		}
	}

	return true
}

// phpValueToGoValue converts PHP value to Go value for JSON encoding
func phpValueToGoValue(val *values.Value) interface{} {
	if val == nil {
		return nil
	}

	switch val.Type {
	case values.TypeNull:
		return nil
	case values.TypeBool:
		return val.Data.(bool)
	case values.TypeInt:
		return val.Data.(int64)
	case values.TypeFloat:
		return val.Data.(float64)
	case values.TypeString:
		return val.Data.(string)
	case values.TypeArray:
		arr := val.Data.(*values.Array)

		// Check if this is a sequential array (should be JSON array)
		if isSequentialArray(arr) {
			// Convert to slice for JSON array encoding
			result := make([]interface{}, len(arr.Elements))
			for i := 0; i < len(arr.Elements); i++ {
				result[i] = phpValueToGoValue(arr.Elements[int64(i)])
			}
			return result
		} else {
			// Convert to map for JSON object encoding
			result := make(map[string]interface{})
			for key, value := range arr.Elements {
				keyStr := fmt.Sprintf("%v", key)
				result[keyStr] = phpValueToGoValue(value)
			}
			return result
		}
	case values.TypeObject:
		obj := val.Data.(*values.Object)
		result := make(map[string]interface{})
		for key, value := range obj.Properties {
			result[key] = phpValueToGoValue(value)
		}
		return result
	default:
		return val.ToString()
	}
}

// goValueToPhpValue converts Go value to PHP value after JSON decoding
func goValueToPhpValue(val interface{}) *values.Value {
	if val == nil {
		return values.NewNull()
	}

	switch v := val.(type) {
	case bool:
		return values.NewBool(v)
	case float64:
		// JSON numbers are always float64, check if it's actually an int
		if v == float64(int64(v)) {
			return values.NewInt(int64(v))
		}
		return values.NewFloat(v)
	case string:
		return values.NewString(v)
	case map[string]interface{}:
		arr := values.NewArray()
		arrData := arr.Data.(*values.Array)
		for key, value := range v {
			arrData.Elements[key] = goValueToPhpValue(value)
		}
		return arr
	case []interface{}:
		arr := values.NewArray()
		arrData := arr.Data.(*values.Array)
		for i, value := range v {
			arrData.Elements[int64(i)] = goValueToPhpValue(value)
		}
		return arr
	default:
		return values.NewString(fmt.Sprintf("%v", v))
	}
}

// GetEncodingFunctions returns all encoding-related functions
func GetEncodingFunctions() []*registry.Function {
	return []*registry.Function{
		{
			Name: "json_encode",
			Parameters: []*registry.Parameter{
				{Name: "value", Type: "mixed"},
			},
			ReturnType: "string",
			MinArgs:    1,
			MaxArgs:    4,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewString("null"), nil
				}

				// Convert PHP value to Go value for JSON encoding
				goValue := phpValueToGoValue(args[0])

				// Encode to JSON
				jsonBytes, err := json.Marshal(goValue)
				if err != nil {
					return values.NewBool(false), nil // PHP returns false on error
				}

				return values.NewString(string(jsonBytes)), nil
			},
		},
		{
			Name: "json_decode",
			Parameters: []*registry.Parameter{
				{Name: "json", Type: "string"},
				{Name: "associative", Type: "bool", HasDefault: true, DefaultValue: values.NewBool(false)},
				{Name: "depth", Type: "int", HasDefault: true, DefaultValue: values.NewInt(512)},
				{Name: "flags", Type: "int", HasDefault: true, DefaultValue: values.NewInt(0)},
			},
			ReturnType: "mixed",
			MinArgs:    1,
			MaxArgs:    4,
			IsBuiltin:  true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewNull(), nil
				}

				jsonStr := args[0].ToString()

				flags := int64(0)
				if len(args) >= 4 {
					flags = args[3].ToInt()
				}

				var result interface{}
				err := json.Unmarshal([]byte(jsonStr), &result)
				if err != nil {
					const JSON_THROW_ON_ERROR = 4194304
					if (flags & JSON_THROW_ON_ERROR) != 0 {
						exception := CreateException(ctx, "JsonException", err.Error())
						if exception == nil {
							return nil, fmt.Errorf("JsonException class not found")
						}
						return nil, ctx.ThrowException(exception)
					}
					return values.NewNull(), nil
				}

				return goValueToPhpValue(result), nil
			},
		},
	}
}