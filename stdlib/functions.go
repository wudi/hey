package stdlib

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/wudi/hey/values"
	"github.com/wudi/hey/vm"
)

// initFunctions initializes built-in PHP functions
func (stdlib *StandardLibrary) initFunctions() {
	// String functions
	stdlib.Functions["strlen"] = BuiltinFunction{
		Name:    "strlen",
		Handler: strlenHandler,
		Parameters: []*Parameter{
			{Name: "string", Type: "string", IsReference: false, HasDefault: false},
		},
		IsVariadic: false,
		MinArgs:    1,
		MaxArgs:    1,
	}

	stdlib.Functions["substr"] = BuiltinFunction{
		Name:    "substr",
		Handler: substrHandler,
		Parameters: []*Parameter{
			{Name: "string", Type: "string", IsReference: false, HasDefault: false},
			{Name: "start", Type: "int", IsReference: false, HasDefault: false},
			{Name: "length", Type: "int", IsReference: false, HasDefault: true, DefaultValue: values.NewNull()},
		},
		IsVariadic: false,
		MinArgs:    2,
		MaxArgs:    3,
	}

	stdlib.Functions["strpos"] = BuiltinFunction{
		Name:    "strpos",
		Handler: strposHandler,
		Parameters: []*Parameter{
			{Name: "haystack", Type: "string", IsReference: false, HasDefault: false},
			{Name: "needle", Type: "string", IsReference: false, HasDefault: false},
			{Name: "offset", Type: "int", IsReference: false, HasDefault: true, DefaultValue: values.NewInt(0)},
		},
		IsVariadic: false,
		MinArgs:    2,
		MaxArgs:    3,
	}

	stdlib.Functions["str_replace"] = BuiltinFunction{
		Name:    "str_replace",
		Handler: strReplaceHandler,
		Parameters: []*Parameter{
			{Name: "search", Type: "mixed", IsReference: false, HasDefault: false},
			{Name: "replace", Type: "mixed", IsReference: false, HasDefault: false},
			{Name: "subject", Type: "mixed", IsReference: false, HasDefault: false},
		},
		IsVariadic: false,
		MinArgs:    3,
		MaxArgs:    3,
	}

	stdlib.Functions["strtolower"] = BuiltinFunction{
		Name:    "strtolower",
		Handler: strtolowerHandler,
		Parameters: []*Parameter{
			{Name: "string", Type: "string", IsReference: false, HasDefault: false},
		},
		IsVariadic: false,
		MinArgs:    1,
		MaxArgs:    1,
	}

	stdlib.Functions["strtoupper"] = BuiltinFunction{
		Name:    "strtoupper",
		Handler: strtoupperHandler,
		Parameters: []*Parameter{
			{Name: "string", Type: "string", IsReference: false, HasDefault: false},
		},
		IsVariadic: false,
		MinArgs:    1,
		MaxArgs:    1,
	}

	stdlib.Functions["trim"] = BuiltinFunction{
		Name:    "trim",
		Handler: trimHandler,
		Parameters: []*Parameter{
			{Name: "string", Type: "string", IsReference: false, HasDefault: false},
			{Name: "characters", Type: "string", IsReference: false, HasDefault: true, DefaultValue: values.NewString(" \t\n\r\x00\x0B")},
		},
		IsVariadic: false,
		MinArgs:    1,
		MaxArgs:    2,
	}

	stdlib.Functions["explode"] = BuiltinFunction{
		Name:    "explode",
		Handler: explodeHandler,
		Parameters: []*Parameter{
			{Name: "delimiter", Type: "string", IsReference: false, HasDefault: false},
			{Name: "string", Type: "string", IsReference: false, HasDefault: false},
			{Name: "limit", Type: "int", IsReference: false, HasDefault: true, DefaultValue: values.NewInt(math.MaxInt32)},
		},
		IsVariadic: false,
		MinArgs:    2,
		MaxArgs:    3,
	}

	stdlib.Functions["implode"] = BuiltinFunction{
		Name:    "implode",
		Handler: implodeHandler,
		Parameters: []*Parameter{
			{Name: "separator", Type: "string", IsReference: false, HasDefault: false},
			{Name: "array", Type: "array", IsReference: false, HasDefault: false},
		},
		IsVariadic: false,
		MinArgs:    2,
		MaxArgs:    2,
	}

	// Array functions
	stdlib.Functions["count"] = BuiltinFunction{
		Name:    "count",
		Handler: countHandler,
		Parameters: []*Parameter{
			{Name: "value", Type: "mixed", IsReference: false, HasDefault: false},
			{Name: "mode", Type: "int", IsReference: false, HasDefault: true, DefaultValue: values.NewInt(0)},
		},
		IsVariadic: false,
		MinArgs:    1,
		MaxArgs:    2,
	}

	stdlib.Functions["array_push"] = BuiltinFunction{
		Name:    "array_push",
		Handler: arrayPushHandler,
		Parameters: []*Parameter{
			{Name: "array", Type: "array", IsReference: true, HasDefault: false},
		},
		IsVariadic: true,
		MinArgs:    2,
		MaxArgs:    -1,
	}

	stdlib.Functions["array_pop"] = BuiltinFunction{
		Name:    "array_pop",
		Handler: arrayPopHandler,
		Parameters: []*Parameter{
			{Name: "array", Type: "array", IsReference: true, HasDefault: false},
		},
		IsVariadic: false,
		MinArgs:    1,
		MaxArgs:    1,
	}

	stdlib.Functions["array_keys"] = BuiltinFunction{
		Name:    "array_keys",
		Handler: arrayKeysHandler,
		Parameters: []*Parameter{
			{Name: "array", Type: "array", IsReference: false, HasDefault: false},
		},
		IsVariadic: false,
		MinArgs:    1,
		MaxArgs:    1,
	}

	stdlib.Functions["array_values"] = BuiltinFunction{
		Name:    "array_values",
		Handler: arrayValuesHandler,
		Parameters: []*Parameter{
			{Name: "array", Type: "array", IsReference: false, HasDefault: false},
		},
		IsVariadic: false,
		MinArgs:    1,
		MaxArgs:    1,
	}

	stdlib.Functions["in_array"] = BuiltinFunction{
		Name:    "in_array",
		Handler: inArrayHandler,
		Parameters: []*Parameter{
			{Name: "needle", Type: "mixed", IsReference: false, HasDefault: false},
			{Name: "haystack", Type: "array", IsReference: false, HasDefault: false},
			{Name: "strict", Type: "bool", IsReference: false, HasDefault: true, DefaultValue: values.NewBool(false)},
		},
		IsVariadic: false,
		MinArgs:    2,
		MaxArgs:    3,
	}

	// Math functions
	stdlib.Functions["abs"] = BuiltinFunction{
		Name:    "abs",
		Handler: absHandler,
		Parameters: []*Parameter{
			{Name: "number", Type: "mixed", IsReference: false, HasDefault: false},
		},
		IsVariadic: false,
		MinArgs:    1,
		MaxArgs:    1,
	}

	stdlib.Functions["max"] = BuiltinFunction{
		Name:    "max",
		Handler: maxHandler,
		Parameters: []*Parameter{
			{Name: "value", Type: "mixed", IsReference: false, HasDefault: false},
		},
		IsVariadic: true,
		MinArgs:    1,
		MaxArgs:    -1,
	}

	stdlib.Functions["min"] = BuiltinFunction{
		Name:    "min",
		Handler: minHandler,
		Parameters: []*Parameter{
			{Name: "value", Type: "mixed", IsReference: false, HasDefault: false},
		},
		IsVariadic: true,
		MinArgs:    1,
		MaxArgs:    -1,
	}

	stdlib.Functions["round"] = BuiltinFunction{
		Name:    "round",
		Handler: roundHandler,
		Parameters: []*Parameter{
			{Name: "val", Type: "float", IsReference: false, HasDefault: false},
			{Name: "precision", Type: "int", IsReference: false, HasDefault: true, DefaultValue: values.NewInt(0)},
			{Name: "mode", Type: "int", IsReference: false, HasDefault: true, DefaultValue: values.NewInt(1)},
		},
		IsVariadic: false,
		MinArgs:    1,
		MaxArgs:    3,
	}

	stdlib.Functions["floor"] = BuiltinFunction{
		Name:    "floor",
		Handler: floorHandler,
		Parameters: []*Parameter{
			{Name: "value", Type: "float", IsReference: false, HasDefault: false},
		},
		IsVariadic: false,
		MinArgs:    1,
		MaxArgs:    1,
	}

	stdlib.Functions["ceil"] = BuiltinFunction{
		Name:    "ceil",
		Handler: ceilHandler,
		Parameters: []*Parameter{
			{Name: "value", Type: "float", IsReference: false, HasDefault: false},
		},
		IsVariadic: false,
		MinArgs:    1,
		MaxArgs:    1,
	}

	stdlib.Functions["pow"] = BuiltinFunction{
		Name:    "pow",
		Handler: powHandler,
		Parameters: []*Parameter{
			{Name: "base", Type: "mixed", IsReference: false, HasDefault: false},
			{Name: "exp", Type: "mixed", IsReference: false, HasDefault: false},
		},
		IsVariadic: false,
		MinArgs:    2,
		MaxArgs:    2,
	}

	stdlib.Functions["sqrt"] = BuiltinFunction{
		Name:    "sqrt",
		Handler: sqrtHandler,
		Parameters: []*Parameter{
			{Name: "arg", Type: "float", IsReference: false, HasDefault: false},
		},
		IsVariadic: false,
		MinArgs:    1,
		MaxArgs:    1,
	}

	// Type functions
	stdlib.Functions["is_string"] = BuiltinFunction{
		Name:    "is_string",
		Handler: isStringHandler,
		Parameters: []*Parameter{
			{Name: "value", Type: "mixed", IsReference: false, HasDefault: false},
		},
		IsVariadic: false,
		MinArgs:    1,
		MaxArgs:    1,
	}

	stdlib.Functions["is_int"] = BuiltinFunction{
		Name:    "is_int",
		Handler: isIntHandler,
		Parameters: []*Parameter{
			{Name: "value", Type: "mixed", IsReference: false, HasDefault: false},
		},
		IsVariadic: false,
		MinArgs:    1,
		MaxArgs:    1,
	}

	stdlib.Functions["is_float"] = BuiltinFunction{
		Name:    "is_float",
		Handler: isFloatHandler,
		Parameters: []*Parameter{
			{Name: "value", Type: "mixed", IsReference: false, HasDefault: false},
		},
		IsVariadic: false,
		MinArgs:    1,
		MaxArgs:    1,
	}

	stdlib.Functions["is_bool"] = BuiltinFunction{
		Name:    "is_bool",
		Handler: isBoolHandler,
		Parameters: []*Parameter{
			{Name: "value", Type: "mixed", IsReference: false, HasDefault: false},
		},
		IsVariadic: false,
		MinArgs:    1,
		MaxArgs:    1,
	}

	stdlib.Functions["is_array"] = BuiltinFunction{
		Name:    "is_array",
		Handler: isArrayHandler,
		Parameters: []*Parameter{
			{Name: "value", Type: "mixed", IsReference: false, HasDefault: false},
		},
		IsVariadic: false,
		MinArgs:    1,
		MaxArgs:    1,
	}

	stdlib.Functions["is_null"] = BuiltinFunction{
		Name:    "is_null",
		Handler: isNullHandler,
		Parameters: []*Parameter{
			{Name: "value", Type: "mixed", IsReference: false, HasDefault: false},
		},
		IsVariadic: false,
		MinArgs:    1,
		MaxArgs:    1,
	}

	stdlib.Functions["is_numeric"] = BuiltinFunction{
		Name:    "is_numeric",
		Handler: isNumericHandler,
		Parameters: []*Parameter{
			{Name: "value", Type: "mixed", IsReference: false, HasDefault: false},
		},
		IsVariadic: false,
		MinArgs:    1,
		MaxArgs:    1,
	}

	stdlib.Functions["isset"] = BuiltinFunction{
		Name:    "isset",
		Handler: issetHandler,
		Parameters: []*Parameter{
			{Name: "var", Type: "mixed", IsReference: false, HasDefault: false},
		},
		IsVariadic: true,
		MinArgs:    1,
		MaxArgs:    -1,
	}

	stdlib.Functions["empty"] = BuiltinFunction{
		Name:    "empty",
		Handler: emptyHandler,
		Parameters: []*Parameter{
			{Name: "var", Type: "mixed", IsReference: false, HasDefault: false},
		},
		IsVariadic: false,
		MinArgs:    1,
		MaxArgs:    1,
	}

	// Conversion functions
	stdlib.Functions["intval"] = BuiltinFunction{
		Name:    "intval",
		Handler: intvalHandler,
		Parameters: []*Parameter{
			{Name: "value", Type: "mixed", IsReference: false, HasDefault: false},
			{Name: "base", Type: "int", IsReference: false, HasDefault: true, DefaultValue: values.NewInt(10)},
		},
		IsVariadic: false,
		MinArgs:    1,
		MaxArgs:    2,
	}

	stdlib.Functions["floatval"] = BuiltinFunction{
		Name:    "floatval",
		Handler: floatvalHandler,
		Parameters: []*Parameter{
			{Name: "value", Type: "mixed", IsReference: false, HasDefault: false},
		},
		IsVariadic: false,
		MinArgs:    1,
		MaxArgs:    1,
	}

	stdlib.Functions["strval"] = BuiltinFunction{
		Name:    "strval",
		Handler: strvalHandler,
		Parameters: []*Parameter{
			{Name: "value", Type: "mixed", IsReference: false, HasDefault: false},
		},
		IsVariadic: false,
		MinArgs:    1,
		MaxArgs:    1,
	}

	stdlib.Functions["boolval"] = BuiltinFunction{
		Name:    "boolval",
		Handler: boolvalHandler,
		Parameters: []*Parameter{
			{Name: "value", Type: "mixed", IsReference: false, HasDefault: false},
		},
		IsVariadic: false,
		MinArgs:    1,
		MaxArgs:    1,
	}

	// Output functions
	stdlib.Functions["var_dump"] = BuiltinFunction{
		Name:    "var_dump",
		Handler: varDumpHandler,
		Parameters: []*Parameter{
			{Name: "value", Type: "mixed", IsReference: false, HasDefault: false},
		},
		IsVariadic: true,
		MinArgs:    1,
		MaxArgs:    -1,
	}

	stdlib.Functions["print_r"] = BuiltinFunction{
		Name:    "print_r",
		Handler: printRHandler,
		Parameters: []*Parameter{
			{Name: "value", Type: "mixed", IsReference: false, HasDefault: false},
			{Name: "return", Type: "bool", IsReference: false, HasDefault: true, DefaultValue: values.NewBool(false)},
		},
		IsVariadic: false,
		MinArgs:    1,
		MaxArgs:    2,
	}

	// Time functions
	stdlib.Functions["time"] = BuiltinFunction{
		Name:       "time",
		Handler:    timeHandler,
		Parameters: []*Parameter{},
		IsVariadic: false,
		MinArgs:    0,
		MaxArgs:    0,
	}

	stdlib.Functions["microtime"] = BuiltinFunction{
		Name:    "microtime",
		Handler: microtimeHandler,
		Parameters: []*Parameter{
			{Name: "as_float", Type: "bool", IsReference: false, HasDefault: true, DefaultValue: values.NewBool(false)},
		},
		IsVariadic: false,
		MinArgs:    0,
		MaxArgs:    1,
	}

	stdlib.Functions["date"] = BuiltinFunction{
		Name:    "date",
		Handler: dateHandler,
		Parameters: []*Parameter{
			{Name: "format", Type: "string", IsReference: false, HasDefault: false},
			{Name: "timestamp", Type: "int", IsReference: false, HasDefault: true, DefaultValue: values.NewNull()},
		},
		IsVariadic: false,
		MinArgs:    1,
		MaxArgs:    2,
	}
}

// String function handlers

func strlenHandler(ctx *vm.ExecutionContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("strlen() expects exactly 1 parameter, %d given", len(args))
	}
	str := args[0].ToString()
	return values.NewInt(int64(len(str))), nil
}

func substrHandler(ctx *vm.ExecutionContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("substr() expects at least 2 parameters, %d given", len(args))
	}

	str := args[0].ToString()
	start := int(args[1].ToInt())

	if start < 0 {
		start = len(str) + start
		if start < 0 {
			start = 0
		}
	}

	if start >= len(str) {
		return values.NewString(""), nil
	}

	if len(args) >= 3 && !args[2].IsNull() {
		length := int(args[2].ToInt())
		if length < 0 {
			end := len(str) + length
			if end <= start {
				return values.NewString(""), nil
			}
			return values.NewString(str[start:end]), nil
		}
		end := start + length
		if end > len(str) {
			end = len(str)
		}
		return values.NewString(str[start:end]), nil
	}

	return values.NewString(str[start:]), nil
}

func strposHandler(ctx *vm.ExecutionContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("strpos() expects at least 2 parameters, %d given", len(args))
	}

	haystack := args[0].ToString()
	needle := args[1].ToString()
	offset := 0

	if len(args) >= 3 {
		offset = int(args[2].ToInt())
	}

	if offset >= len(haystack) {
		return values.NewBool(false), nil
	}

	pos := strings.Index(haystack[offset:], needle)
	if pos == -1 {
		return values.NewBool(false), nil
	}

	return values.NewInt(int64(offset + pos)), nil
}

func strReplaceHandler(ctx *vm.ExecutionContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 3 {
		return nil, fmt.Errorf("str_replace() expects at least 3 parameters, %d given", len(args))
	}

	search := args[0].ToString()
	replace := args[1].ToString()
	subject := args[2].ToString()

	result := strings.ReplaceAll(subject, search, replace)
	return values.NewString(result), nil
}

func strtolowerHandler(ctx *vm.ExecutionContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("strtolower() expects exactly 1 parameter, %d given", len(args))
	}
	str := args[0].ToString()
	return values.NewString(strings.ToLower(str)), nil
}

func strtoupperHandler(ctx *vm.ExecutionContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("strtoupper() expects exactly 1 parameter, %d given", len(args))
	}
	str := args[0].ToString()
	return values.NewString(strings.ToUpper(str)), nil
}

func trimHandler(ctx *vm.ExecutionContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("trim() expects at least 1 parameter, %d given", len(args))
	}

	str := args[0].ToString()
	chars := " \t\n\r\x00\x0B"

	if len(args) >= 2 {
		chars = args[1].ToString()
	}

	return values.NewString(strings.Trim(str, chars)), nil
}

func explodeHandler(ctx *vm.ExecutionContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("explode() expects at least 2 parameters, %d given", len(args))
	}

	delimiter := args[0].ToString()
	str := args[1].ToString()
	limit := math.MaxInt32

	if len(args) >= 3 {
		limit = int(args[2].ToInt())
	}

	var parts []string
	if limit == 1 {
		parts = []string{str}
	} else if limit > 1 {
		parts = strings.SplitN(str, delimiter, limit)
	} else {
		parts = strings.Split(str, delimiter)
	}

	result := values.NewArray()
	for i, part := range parts {
		result.ArraySet(values.NewInt(int64(i)), values.NewString(part))
	}

	return result, nil
}

func implodeHandler(ctx *vm.ExecutionContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("implode() expects exactly 2 parameters, %d given", len(args))
	}

	separator := args[0].ToString()
	arr := args[1]

	if !arr.IsArray() {
		return nil, fmt.Errorf("implode(): Argument must be an array")
	}

	var parts []string
	arrayData := arr.Data.(*values.Array)

	// Get array values in order
	for i := int64(0); i < int64(len(arrayData.Elements)); i++ {
		if val, exists := arrayData.Elements[i]; exists {
			parts = append(parts, val.ToString())
		}
	}

	result := strings.Join(parts, separator)
	return values.NewString(result), nil
}

// Array function handlers

func countHandler(ctx *vm.ExecutionContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("count() expects at least 1 parameter, %d given", len(args))
	}

	value := args[0]
	if value.IsArray() {
		return values.NewInt(int64(value.ArrayCount())), nil
	}

	if value.IsNull() {
		return values.NewInt(0), nil
	}

	return values.NewInt(1), nil
}

func arrayPushHandler(ctx *vm.ExecutionContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("array_push() expects at least 2 parameters, %d given", len(args))
	}

	arr := args[0]
	if !arr.IsArray() {
		return nil, fmt.Errorf("array_push(): Argument must be an array")
	}

	for i := 1; i < len(args); i++ {
		arr.ArraySet(nil, args[i])
	}

	return values.NewInt(int64(arr.ArrayCount())), nil
}

func arrayPopHandler(ctx *vm.ExecutionContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("array_pop() expects exactly 1 parameter, %d given", len(args))
	}

	arr := args[0]
	if !arr.IsArray() {
		return nil, fmt.Errorf("array_pop(): Argument must be an array")
	}

	arrayData := arr.Data.(*values.Array)
	if len(arrayData.Elements) == 0 {
		return values.NewNull(), nil
	}

	// Find the highest numeric key
	var maxKey interface{}
	var maxValue *values.Value

	for key, value := range arrayData.Elements {
		if intKey, ok := key.(int64); ok {
			if maxKey == nil || intKey > maxKey.(int64) {
				maxKey = key
				maxValue = value
			}
		}
	}

	if maxKey != nil {
		delete(arrayData.Elements, maxKey)
		return maxValue, nil
	}

	return values.NewNull(), nil
}

func arrayKeysHandler(ctx *vm.ExecutionContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("array_keys() expects at least 1 parameter, %d given", len(args))
	}

	arr := args[0]
	if !arr.IsArray() {
		return nil, fmt.Errorf("array_keys(): Argument must be an array")
	}

	result := values.NewArray()
	arrayData := arr.Data.(*values.Array)
	index := int64(0)

	for key := range arrayData.Elements {
		var keyValue *values.Value
		switch k := key.(type) {
		case int64:
			keyValue = values.NewInt(k)
		case string:
			keyValue = values.NewString(k)
		default:
			keyValue = values.NewString(fmt.Sprintf("%v", k))
		}
		result.ArraySet(values.NewInt(index), keyValue)
		index++
	}

	return result, nil
}

func arrayValuesHandler(ctx *vm.ExecutionContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("array_values() expects exactly 1 parameter, %d given", len(args))
	}

	arr := args[0]
	if !arr.IsArray() {
		return nil, fmt.Errorf("array_values(): Argument must be an array")
	}

	result := values.NewArray()
	arrayData := arr.Data.(*values.Array)
	index := int64(0)

	for _, value := range arrayData.Elements {
		result.ArraySet(values.NewInt(index), value)
		index++
	}

	return result, nil
}

func inArrayHandler(ctx *vm.ExecutionContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("in_array() expects at least 2 parameters, %d given", len(args))
	}

	needle := args[0]
	haystack := args[1]
	strict := false

	if len(args) >= 3 {
		strict = args[2].ToBool()
	}

	if !haystack.IsArray() {
		return nil, fmt.Errorf("in_array(): Argument must be an array")
	}

	arrayData := haystack.Data.(*values.Array)
	for _, value := range arrayData.Elements {
		var matches bool
		if strict {
			matches = needle.Identical(value)
		} else {
			matches = needle.Equal(value)
		}

		if matches {
			return values.NewBool(true), nil
		}
	}

	return values.NewBool(false), nil
}

// Math function handlers

func absHandler(ctx *vm.ExecutionContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("abs() expects exactly 1 parameter, %d given", len(args))
	}

	value := args[0]
	if value.IsFloat() {
		f := value.ToFloat()
		if f < 0 {
			return values.NewFloat(-f), nil
		}
		return values.NewFloat(f), nil
	}

	i := value.ToInt()
	if i < 0 {
		return values.NewInt(-i), nil
	}
	return values.NewInt(i), nil
}

func maxHandler(ctx *vm.ExecutionContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("max() expects at least 1 parameter, %d given", len(args))
	}

	max := args[0]
	for i := 1; i < len(args); i++ {
		if args[i].Compare(max) > 0 {
			max = args[i]
		}
	}

	return max, nil
}

func minHandler(ctx *vm.ExecutionContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("min() expects at least 1 parameter, %d given", len(args))
	}

	min := args[0]
	for i := 1; i < len(args); i++ {
		if args[i].Compare(min) < 0 {
			min = args[i]
		}
	}

	return min, nil
}

func roundHandler(ctx *vm.ExecutionContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("round() expects at least 1 parameter, %d given", len(args))
	}

	val := args[0].ToFloat()
	precision := int64(0)

	if len(args) >= 2 {
		precision = args[1].ToInt()
	}

	multiplier := math.Pow(10, float64(precision))
	result := math.Round(val*multiplier) / multiplier

	if precision == 0 && result == float64(int64(result)) {
		return values.NewInt(int64(result)), nil
	}

	return values.NewFloat(result), nil
}

func floorHandler(ctx *vm.ExecutionContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("floor() expects exactly 1 parameter, %d given", len(args))
	}

	val := args[0].ToFloat()
	result := math.Floor(val)

	return values.NewFloat(result), nil
}

func ceilHandler(ctx *vm.ExecutionContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("ceil() expects exactly 1 parameter, %d given", len(args))
	}

	val := args[0].ToFloat()
	result := math.Ceil(val)

	return values.NewFloat(result), nil
}

func powHandler(ctx *vm.ExecutionContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("pow() expects exactly 2 parameters, %d given", len(args))
	}

	base := args[0].ToFloat()
	exp := args[1].ToFloat()
	result := math.Pow(base, exp)

	if result == float64(int64(result)) && result >= -9223372036854775808 && result <= 9223372036854775807 {
		return values.NewInt(int64(result)), nil
	}

	return values.NewFloat(result), nil
}

func sqrtHandler(ctx *vm.ExecutionContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("sqrt() expects exactly 1 parameter, %d given", len(args))
	}

	val := args[0].ToFloat()
	result := math.Sqrt(val)

	return values.NewFloat(result), nil
}

// Type checking function handlers

func isStringHandler(ctx *vm.ExecutionContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("is_string() expects exactly 1 parameter, %d given", len(args))
	}
	return values.NewBool(args[0].IsString()), nil
}

func isIntHandler(ctx *vm.ExecutionContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("is_int() expects exactly 1 parameter, %d given", len(args))
	}
	return values.NewBool(args[0].IsInt()), nil
}

func isFloatHandler(ctx *vm.ExecutionContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("is_float() expects exactly 1 parameter, %d given", len(args))
	}
	return values.NewBool(args[0].IsFloat()), nil
}

func isBoolHandler(ctx *vm.ExecutionContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("is_bool() expects exactly 1 parameter, %d given", len(args))
	}
	return values.NewBool(args[0].IsBool()), nil
}

func isArrayHandler(ctx *vm.ExecutionContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("is_array() expects exactly 1 parameter, %d given", len(args))
	}
	return values.NewBool(args[0].IsArray()), nil
}

func isNullHandler(ctx *vm.ExecutionContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("is_null() expects exactly 1 parameter, %d given", len(args))
	}
	return values.NewBool(args[0].IsNull()), nil
}

func isNumericHandler(ctx *vm.ExecutionContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("is_numeric() expects exactly 1 parameter, %d given", len(args))
	}

	value := args[0]
	if value.IsNumeric() {
		return values.NewBool(true), nil
	}

	if value.IsString() {
		str := strings.TrimSpace(value.ToString())
		if _, err := strconv.ParseFloat(str, 64); err == nil {
			return values.NewBool(true), nil
		}
	}

	return values.NewBool(false), nil
}

func issetHandler(ctx *vm.ExecutionContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("isset() expects at least 1 parameter, %d given", len(args))
	}

	for _, arg := range args {
		if arg.IsNull() {
			return values.NewBool(false), nil
		}
	}

	return values.NewBool(true), nil
}

func emptyHandler(ctx *vm.ExecutionContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("empty() expects exactly 1 parameter, %d given", len(args))
	}

	value := args[0]
	isEmpty := value.IsNull() || !value.ToBool()

	return values.NewBool(isEmpty), nil
}

// Conversion function handlers

func intvalHandler(ctx *vm.ExecutionContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("intval() expects at least 1 parameter, %d given", len(args))
	}

	return values.NewInt(args[0].ToInt()), nil
}

func floatvalHandler(ctx *vm.ExecutionContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("floatval() expects exactly 1 parameter, %d given", len(args))
	}

	return values.NewFloat(args[0].ToFloat()), nil
}

func strvalHandler(ctx *vm.ExecutionContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("strval() expects exactly 1 parameter, %d given", len(args))
	}

	return values.NewString(args[0].ToString()), nil
}

func boolvalHandler(ctx *vm.ExecutionContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("boolval() expects exactly 1 parameter, %d given", len(args))
	}

	return values.NewBool(args[0].ToBool()), nil
}

// Output function handlers

func varDumpHandler(ctx *vm.ExecutionContext, args []*values.Value) (*values.Value, error) {
	for _, arg := range args {
		fmt.Println(arg.String())
	}
	return values.NewNull(), nil
}

func printRHandler(ctx *vm.ExecutionContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("print_r() expects at least 1 parameter, %d given", len(args))
	}

	returnOutput := false
	if len(args) >= 2 {
		returnOutput = args[1].ToBool()
	}

	output := args[0].String()
	if returnOutput {
		return values.NewString(output), nil
	}

	fmt.Print(output)
	return values.NewBool(true), nil
}

// Time function handlers

func timeHandler(ctx *vm.ExecutionContext, args []*values.Value) (*values.Value, error) {
	return values.NewInt(time.Now().Unix()), nil
}

func microtimeHandler(ctx *vm.ExecutionContext, args []*values.Value) (*values.Value, error) {
	asFloat := false
	if len(args) >= 1 {
		asFloat = args[0].ToBool()
	}

	now := time.Now()
	if asFloat {
		timestamp := float64(now.Unix()) + float64(now.Nanosecond())/1e9
		return values.NewFloat(timestamp), nil
	}

	micro := now.Nanosecond() / 1000
	sec := now.Unix()
	result := fmt.Sprintf("0.%06d %d", micro, sec)
	return values.NewString(result), nil
}

func dateHandler(ctx *vm.ExecutionContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("date() expects at least 1 parameter, %d given", len(args))
	}

	format := args[0].ToString()
	var timestamp time.Time

	if len(args) >= 2 && !args[1].IsNull() {
		timestamp = time.Unix(args[1].ToInt(), 0)
	} else {
		timestamp = time.Now()
	}

	// Simplified date formatting - in a full implementation, this would support all PHP date format characters
	goFormat := strings.ReplaceAll(format, "Y", "2006")
	goFormat = strings.ReplaceAll(goFormat, "m", "01")
	goFormat = strings.ReplaceAll(goFormat, "d", "02")
	goFormat = strings.ReplaceAll(goFormat, "H", "15")
	goFormat = strings.ReplaceAll(goFormat, "i", "04")
	goFormat = strings.ReplaceAll(goFormat, "s", "05")

	result := timestamp.Format(goFormat)
	return values.NewString(result), nil
}
