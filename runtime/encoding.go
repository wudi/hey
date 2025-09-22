package runtime

import (
	"encoding/json"
	"fmt"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

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
		result := make(map[string]interface{})
		for key, value := range arr.Elements {
			keyStr := fmt.Sprintf("%v", key)
			result[keyStr] = phpValueToGoValue(value)
		}
		return result
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
				{Name: "associative", Type: "bool"},
			},
			ReturnType: "mixed",
			MinArgs:    1,
			MaxArgs:    4,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil {
					return values.NewNull(), nil
				}

				jsonStr := args[0].ToString()

				// Parse JSON into generic interface
				var result interface{}
				err := json.Unmarshal([]byte(jsonStr), &result)
				if err != nil {
					return values.NewNull(), nil // PHP returns null on error
				}

				// Convert Go value back to PHP value
				return goValueToPhpValue(result), nil
			},
		},
	}
}