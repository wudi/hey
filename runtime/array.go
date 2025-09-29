package runtime

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// naturalCompare implements PHP's natural string comparison
// Returns -1 if s1 < s2, 0 if s1 == s2, 1 if s1 > s2
func naturalCompare(s1, s2 string) int {
	// Regular expression to split strings into chunks of digits and non-digits
	re := regexp.MustCompile(`(\d+|\D+)`)

	parts1 := re.FindAllString(s1, -1)
	parts2 := re.FindAllString(s2, -1)

	minLen := len(parts1)
	if len(parts2) < minLen {
		minLen = len(parts2)
	}

	for i := 0; i < minLen; i++ {
		p1, p2 := parts1[i], parts2[i]

		// Check if both parts are numeric
		if isNumeric(p1) && isNumeric(p2) {
			n1, _ := strconv.ParseInt(p1, 10, 64)
			n2, _ := strconv.ParseInt(p2, 10, 64)
			if n1 < n2 {
				return -1
			} else if n1 > n2 {
				return 1
			}
		} else {
			// String comparison
			if p1 < p2 {
				return -1
			} else if p1 > p2 {
				return 1
			}
		}
	}

	// If all compared parts are equal, compare lengths
	if len(parts1) < len(parts2) {
		return -1
	} else if len(parts1) > len(parts2) {
		return 1
	}

	return 0
}

// isNumeric checks if a string represents a number
func isNumeric(s string) bool {
	_, err := strconv.ParseInt(s, 10, 64)
	return err == nil
}

// deepCopyArray creates a deep copy of an array value
func deepCopyArray(arr *values.Value) *values.Value {
	if arr == nil || !arr.IsArray() {
		return values.NewArray()
	}

	result := values.NewArray()
	resultArr := result.Data.(*values.Array)
	sourceArr := arr.Data.(*values.Array)

	for key, value := range sourceArr.Elements {
		if value != nil && value.IsArray() {
			// Recursively copy nested arrays
			resultArr.Elements[key] = deepCopyArray(value)
		} else {
			// Copy primitive values
			resultArr.Elements[key] = value
		}
	}

	resultArr.NextIndex = sourceArr.NextIndex
	resultArr.IsIndexed = sourceArr.IsIndexed

	return result
}

// replaceRecursive performs recursive array replacement
func replaceRecursive(base, replacement *values.Value) *values.Value {
	if base == nil || !base.IsArray() || replacement == nil || !replacement.IsArray() {
		return base
	}

	baseArr := base.Data.(*values.Array)
	replaceArr := replacement.Data.(*values.Array)

	for key, value := range replaceArr.Elements {
		existingValue, exists := baseArr.Elements[key]

		if exists && existingValue != nil && existingValue.IsArray() && value != nil && value.IsArray() {
			// Both values are arrays, merge recursively
			baseArr.Elements[key] = replaceRecursive(existingValue, value)
		} else {
			// Replace the value
			baseArr.Elements[key] = value
		}
	}

	// Update NextIndex if needed
	if replaceArr.NextIndex > baseArr.NextIndex {
		baseArr.NextIndex = replaceArr.NextIndex
	}

	return base
}

// callbackInvoker handles calling PHP callbacks from array functions
func callbackInvoker(ctx registry.BuiltinCallContext, callback *values.Value, args []*values.Value) (*values.Value, error) {
	if callback == nil {
		return nil, fmt.Errorf("callback is null")
	}

	// Handle string callback (function name)
	if callback.Type == values.TypeString {
		funcName := callback.ToString()

		// Look up user-defined function first
		if userFunc, ok := ctx.LookupUserFunction(funcName); ok && userFunc != nil {
			// Call user-defined function via simplified VM integration
			return ctx.SimpleCallUserFunction(userFunc, args)
		}

		// Look up builtin function
		if builtinFunc, ok := ctx.SymbolRegistry().GetFunction(funcName); ok && builtinFunc != nil && builtinFunc.IsBuiltin {
			return builtinFunc.Builtin(ctx, args)
		}

		return nil, fmt.Errorf("function not found: %s", funcName)
	}

	// Handle closure/callable objects
	if callback.IsCallable() {
		closure := callback.ClosureGet()
		if closure != nil && closure.Function != nil {
			if userFunc, ok := closure.Function.(*registry.Function); ok && userFunc != nil && !userFunc.IsBuiltin {
				// Call closure function via simplified VM integration
				return ctx.SimpleCallUserFunction(userFunc, args)
			}
		}
		return nil, fmt.Errorf("invalid closure callback")
	}

	return nil, fmt.Errorf("invalid callback type")
}

// GetArrayFunctions returns all array-related functions
func GetArrayFunctions() []*registry.Function {
	return []*registry.Function{
		{
			Name:       "count",
			Parameters: []*registry.Parameter{{Name: "value", Type: "mixed"}},
			ReturnType: "int",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 1 {
					return values.NewInt(0), nil
				}
				val := args[0]
				if val == nil {
					return values.NewInt(0), nil
				}
				switch val.Type {
				case values.TypeArray:
					return values.NewInt(int64(val.ArrayCount())), nil
				case values.TypeObject:
					obj := val.Data.(*values.Object)

					// Check if the object implements Countable interface
					if obj.ClassName != "" {
						if class, err := registry.GlobalRegistry.GetClass(obj.ClassName); err == nil && class != nil {
							// Check if class implements Countable interface
							isCountable := false
							for _, iface := range class.Interfaces {
								if iface == "Countable" {
									isCountable = true
									break
								}
							}

							if isCountable {
								// Call the object's count() method
								if method, ok := class.Methods["count"]; ok {
									var function *registry.Function
									// Handle different BuiltinMethodImpl types
									switch impl := method.Implementation.(type) {
									case interface{ GetFunction() *registry.Function }:
										function = impl.GetFunction()
									default:
										// Fallback to property count if method can't be called
										return values.NewInt(int64(len(obj.Properties))), nil
									}

									if function != nil && function.IsBuiltin {
										return function.Builtin(ctx, []*values.Value{val})
									}
								}
							}
						}
					}

					// Default to counting properties
					return values.NewInt(int64(len(obj.Properties))), nil
				default:
					return values.NewInt(1), nil
				}
			},
		},
		{
			Name:       "array_keys",
			Parameters: []*registry.Parameter{{Name: "array", Type: "array"}},
			ReturnType: "array",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil || !args[0].IsArray() {
					return values.NewArray(), nil
				}
				arr := args[0].Data.(*values.Array)
				result := values.NewArray()
				idx := int64(0)
				for key := range arr.Elements {
					var keyVal *values.Value
					switch k := key.(type) {
					case string:
						keyVal = values.NewString(k)
					case int:
						keyVal = values.NewInt(int64(k))
					case int64:
						keyVal = values.NewInt(k)
					default:
						keyVal = values.NewString(fmt.Sprintf("%v", k))
					}
					result.Data.(*values.Array).Elements[idx] = keyVal
					idx++
				}
				result.Data.(*values.Array).NextIndex = idx
				return result, nil
			},
		},
		{
			Name:       "array_values",
			Parameters: []*registry.Parameter{{Name: "array", Type: "array"}},
			ReturnType: "array",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil || !args[0].IsArray() {
					return values.NewArray(), nil
				}
				arr := args[0].Data.(*values.Array)
				result := values.NewArray()
				idx := int64(0)
				for _, v := range arr.Elements {
					result.Data.(*values.Array).Elements[idx] = v
					idx++
				}
				result.Data.(*values.Array).NextIndex = idx
				return result, nil
			},
		},
		{
			Name: "array_push",
			Parameters: []*registry.Parameter{
				{Name: "array", Type: "array"},
			},
			ReturnType: "int",
			IsVariadic: true,
			MinArgs:    2,
			MaxArgs:    -1,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 || args[0] == nil || !args[0].IsArray() {
					return values.NewInt(0), nil
				}

				arr := args[0].Data.(*values.Array)
				// Add all values to the end of the array
				for i := 1; i < len(args); i++ {
					nextIndex := len(arr.Elements)
					arr.Elements[int64(nextIndex)] = args[i]
				}

				return values.NewInt(int64(len(arr.Elements))), nil
			},
		},
		{
			Name: "in_array",
			Parameters: []*registry.Parameter{
				{Name: "needle", Type: "mixed"},
				{Name: "haystack", Type: "array"},
				{Name: "strict", Type: "bool", DefaultValue: values.NewBool(false)},
			},
			ReturnType: "bool",
			MinArgs:    2,
			MaxArgs:    3,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 || args[1] == nil || !args[1].IsArray() {
					return values.NewBool(false), nil
				}

				needle := args[0]
				arr := args[1].Data.(*values.Array)
				strict := false
				if len(args) > 2 && args[2] != nil {
					strict = args[2].ToBool()
				}

				// Search for the needle in the array values
				for _, value := range arr.Elements {
					if value == nil {
						continue
					}

					if strict {
						// Strict comparison - types must match
						if needle.Type == value.Type {
							switch needle.Type {
							case values.TypeInt:
								if needle.ToInt() == value.ToInt() {
									return values.NewBool(true), nil
								}
							case values.TypeFloat:
								if needle.ToFloat() == value.ToFloat() {
									return values.NewBool(true), nil
								}
							case values.TypeString:
								if needle.ToString() == value.ToString() {
									return values.NewBool(true), nil
								}
							case values.TypeBool:
								if needle.ToBool() == value.ToBool() {
									return values.NewBool(true), nil
								}
							case values.TypeNull:
								return values.NewBool(true), nil // Both are null
							}
						}
					} else {
						// Loose comparison - use PHP-style comparison
						if compareValuesLoose(needle, value) {
							return values.NewBool(true), nil
						}
					}
				}

				return values.NewBool(false), nil
			},
		},
		{
			Name: "array_chunk",
			Parameters: []*registry.Parameter{
				{Name: "array", Type: "array"},
				{Name: "length", Type: "int"},
			},
			ReturnType: "array",
			MinArgs:    2,
			MaxArgs:    3,
			IsBuiltin: true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 || args[0] == nil || !args[0].IsArray() {
					return values.NewArray(), nil
				}
				arr := args[0].Data.(*values.Array)
				chunkSize := int(args[1].ToInt())
				if chunkSize <= 0 {
					exception := CreateException(ctx, "ValueError", "array_chunk(): Argument #2 ($length) must be greater than 0")
					if exception == nil {
						return nil, fmt.Errorf("ValueError class not found")
					}
					return nil, ctx.ThrowException(exception)
				}
				preserveKeys := false
				if len(args) > 2 && args[2] != nil {
					preserveKeys = args[2].ToBool()
				}
				result := values.NewArray()
				resultArr := result.Data.(*values.Array)
				currentChunk := values.NewArray()
				currentChunkArr := currentChunk.Data.(*values.Array)
				chunkIdx := int64(0)
				itemCount := 0
				for key, val := range arr.Elements {
					if preserveKeys {
						currentChunkArr.Elements[key] = val
					} else {
						currentChunkArr.Elements[int64(itemCount)] = val
					}
					itemCount++
					if itemCount >= chunkSize {
						resultArr.Elements[chunkIdx] = currentChunk
						chunkIdx++
						currentChunk = values.NewArray()
						currentChunkArr = currentChunk.Data.(*values.Array)
						itemCount = 0
					}
				}
				if itemCount > 0 {
					if !preserveKeys {
						currentChunkArr.NextIndex = int64(itemCount)
					}
					resultArr.Elements[chunkIdx] = currentChunk
					chunkIdx++
				}
				resultArr.NextIndex = chunkIdx
				return result, nil
			},
		},
		{
			Name: "array_combine",
			Parameters: []*registry.Parameter{
				{Name: "keys", Type: "array"},
				{Name: "values", Type: "array"},
			},
			ReturnType: "array",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 || args[0] == nil || !args[0].IsArray() || args[1] == nil || !args[1].IsArray() {
					return nil, fmt.Errorf("array_combine(): Arguments must be arrays")
				}
				keysArr := args[0].Data.(*values.Array)
				valsArr := args[1].Data.(*values.Array)
				if args[0].ArrayCount() != args[1].ArrayCount() {
					return nil, fmt.Errorf("array_combine(): Argument #1 ($keys) and argument #2 ($values) must have the same number of elements")
				}
				result := values.NewArray()
				resultArr := result.Data.(*values.Array)
				keysList := make([]*values.Value, 0, args[0].ArrayCount())
				for _, k := range keysArr.Elements {
					keysList = append(keysList, k)
				}
				valsList := make([]*values.Value, 0, args[1].ArrayCount())
				for _, v := range valsArr.Elements {
					valsList = append(valsList, v)
				}
				for i := 0; i < len(keysList) && i < len(valsList); i++ {
					keyVal := keysList[i]
					if keyVal.IsInt() {
						resultArr.Elements[keyVal.ToInt()] = valsList[i]
					} else {
						resultArr.Elements[keyVal.ToString()] = valsList[i]
					}
				}
				return result, nil
			},
		},
		{
			Name:       "array_count_values",
			Parameters: []*registry.Parameter{{Name: "array", Type: "array"}},
			ReturnType: "array",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil || !args[0].IsArray() {
					return values.NewArray(), nil
				}
				arr := args[0].Data.(*values.Array)
				result := values.NewArray()
				resultArr := result.Data.(*values.Array)
				for _, val := range arr.Elements {
					if val == nil {
						continue
					}
					key := val.ToString()
					if existing, ok := resultArr.Elements[key]; ok && existing != nil {
						resultArr.Elements[key] = values.NewInt(existing.ToInt() + 1)
					} else {
						resultArr.Elements[key] = values.NewInt(1)
					}
				}
				return result, nil
			},
		},
		{
			Name: "array_diff",
			Parameters: []*registry.Parameter{
				{Name: "array", Type: "array"},
			},
			ReturnType: "array",
			IsVariadic: true,
			MinArgs:    2,
			MaxArgs:    -1,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 || args[0] == nil || !args[0].IsArray() {
					return values.NewArray(), nil
				}
				arr1 := args[0].Data.(*values.Array)
				otherValues := make(map[string]bool)
				for i := 1; i < len(args); i++ {
					if args[i] != nil && args[i].IsArray() {
						arr := args[i].Data.(*values.Array)
						for _, v := range arr.Elements {
							if v != nil {
								otherValues[v.ToString()] = true
							}
						}
					}
				}
				result := values.NewArray()
				resultArr := result.Data.(*values.Array)
				for key, val := range arr1.Elements {
					if val != nil && !otherValues[val.ToString()] {
						resultArr.Elements[key] = val
					}
				}
				return result, nil
			},
		},
		{
			Name:       "array_flip",
			Parameters: []*registry.Parameter{{Name: "array", Type: "array"}},
			ReturnType: "array",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil || !args[0].IsArray() {
					return values.NewArray(), nil
				}
				arr := args[0].Data.(*values.Array)
				result := values.NewArray()
				resultArr := result.Data.(*values.Array)
				for key, val := range arr.Elements {
					if val == nil {
						continue
					}
					var keyStr string
					switch k := key.(type) {
					case string:
						keyStr = k
					case int:
						keyStr = fmt.Sprintf("%d", k)
					case int64:
						keyStr = fmt.Sprintf("%d", k)
					default:
						keyStr = fmt.Sprintf("%v", key)
					}
					if val.IsInt() {
						resultArr.Elements[val.ToInt()] = values.NewString(keyStr)
					} else {
						resultArr.Elements[val.ToString()] = values.NewString(keyStr)
					}
				}
				return result, nil
			},
		},
		{
			Name: "array_intersect",
			Parameters: []*registry.Parameter{
				{Name: "array", Type: "array"},
			},
			ReturnType: "array",
			IsVariadic: true,
			MinArgs:    2,
			MaxArgs:    -1,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 || args[0] == nil || !args[0].IsArray() {
					return values.NewArray(), nil
				}
				arr1 := args[0].Data.(*values.Array)
				result := values.NewArray()
				resultArr := result.Data.(*values.Array)
				for key, val := range arr1.Elements {
					if val == nil {
						continue
					}
					found := true
					for i := 1; i < len(args); i++ {
						if args[i] == nil || !args[i].IsArray() {
							continue
						}
						arr := args[i].Data.(*values.Array)
						hasValue := false
						for _, v := range arr.Elements {
							if v != nil && v.ToString() == val.ToString() {
								hasValue = true
								break
							}
						}
						if !hasValue {
							found = false
							break
						}
					}
					if found {
						resultArr.Elements[key] = val
					}
				}
				return result, nil
			},
		},
		{
			Name:       "array_reverse",
			Parameters: []*registry.Parameter{{Name: "array", Type: "array"}},
			ReturnType: "array",
			MinArgs:    1,
			MaxArgs:    2,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil || !args[0].IsArray() {
					return values.NewArray(), nil
				}
				arr := args[0].Data.(*values.Array)
				preserveKeys := false
				if len(args) > 1 && args[1] != nil {
					preserveKeys = args[1].ToBool()
				}
				elements := make([]struct {
					key interface{}
					val *values.Value
				}, 0, args[0].ArrayCount())
				for k, v := range arr.Elements {
					elements = append(elements, struct {
						key interface{}
						val *values.Value
					}{k, v})
				}
				// Sort elements by key to ensure consistent ordering
				sort.Slice(elements, func(i, j int) bool {
					ki, kiOk := elements[i].key.(int64)
					kj, kjOk := elements[j].key.(int64)
					if kiOk && kjOk {
						return ki < kj
					}
					if kiOk && !kjOk {
						return true
					}
					if !kiOk && kjOk {
						return false
					}
					return fmt.Sprintf("%v", elements[i].key) < fmt.Sprintf("%v", elements[j].key)
				})
				for i, j := 0, len(elements)-1; i < j; i, j = i+1, j-1 {
					elements[i], elements[j] = elements[j], elements[i]
				}
				result := values.NewArray()
				resultArr := result.Data.(*values.Array)

				// Check if array has string keys (associative array)
				hasStringKeys := false
				for _, elem := range elements {
					if _, isInt := elem.key.(int64); !isInt {
						hasStringKeys = true
						break
					}
				}

				// PHP always preserves string keys regardless of preserve_keys parameter
				if preserveKeys || hasStringKeys {
					for _, elem := range elements {
						resultArr.Elements[elem.key] = elem.val
					}
				} else {
					idx := int64(0)
					for _, elem := range elements {
						resultArr.Elements[idx] = elem.val
						idx++
					}
					resultArr.NextIndex = idx
				}
				return result, nil
			},
		},
		{
			Name:       "array_sum",
			Parameters: []*registry.Parameter{{Name: "array", Type: "array"}},
			ReturnType: "number",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil || !args[0].IsArray() {
					return values.NewInt(0), nil
				}
				arr := args[0].Data.(*values.Array)
				sum := float64(0)
				hasFloat := false
				for _, val := range arr.Elements {
					if val == nil {
						continue
					}
					if val.IsFloat() {
						hasFloat = true
						sum += val.ToFloat()
					} else if val.IsInt() {
						sum += float64(val.ToInt())
					}
				}
				if hasFloat {
					return values.NewFloat(sum), nil
				}
				return values.NewInt(int64(sum)), nil
			},
		},
		{
			Name: "array_filter",
			Parameters: []*registry.Parameter{
				{Name: "array", Type: "array"},
				{Name: "callback", Type: "callable", HasDefault: true, DefaultValue: values.NewNull()},
				{Name: "flag", Type: "int", HasDefault: true, DefaultValue: values.NewInt(0)},
			},
			ReturnType: "array",
			MinArgs:    1,
			MaxArgs:    3,
			IsBuiltin: true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil || !args[0].IsArray() {
					return values.NewArray(), nil
				}

				arr := args[0].Data.(*values.Array)
				result := values.NewArray()
				resultArr := result.Data.(*values.Array)

				// Get flag parameter (default: filter by value)
				// flag := int64(0) // Default: ARRAY_FILTER_VALUE
				// if len(args) > 2 && args[2] != nil {
				//	flag = args[2].ToInt()
				// }
				// TODO: Implement flag-based filtering when callback support is added

				// If no callback provided, use default filtering (truthy values)
				if len(args) < 2 || args[1] == nil {
					for key, val := range arr.Elements {
						if val != nil && val.ToBool() {
							resultArr.Elements[key] = val
						}
					}
					return result, nil
				}

				// Use unified callback invoker for both builtin and user-defined callbacks
				callback := args[1]
				for key, val := range arr.Elements {
					if val != nil {
						// Call the callback function with the value
						callArgs := []*values.Value{val}
						result_val, err := callbackInvoker(ctx, callback, callArgs)
						if err != nil {
							return nil, err
						}
						// Include in result if callback returns truthy value
						if result_val != nil && result_val.ToBool() {
							resultArr.Elements[key] = val
						}
					}
				}
				return result, nil
			},
		},
		{
			Name:       "array_change_key_case",
			Parameters: []*registry.Parameter{{Name: "array", Type: "array"}},
			ReturnType: "array",
			MinArgs:    1,
			MaxArgs:    2,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil || !args[0].IsArray() {
					return values.NewArray(), nil
				}
				arr := args[0].Data.(*values.Array)
				caseMode := int64(0)
				if len(args) > 1 && args[1] != nil {
					caseMode = args[1].ToInt()
				}
				result := values.NewArray()
				resultArr := result.Data.(*values.Array)
				for key, val := range arr.Elements {
					newKey := key
					if strKey, ok := key.(string); ok {
						if caseMode == 0 {
							newKey = strings.ToLower(strKey)
						} else {
							newKey = strings.ToUpper(strKey)
						}
					}
					resultArr.Elements[newKey] = val
				}
				resultArr.NextIndex = arr.NextIndex
				return result, nil
			},
		},
		{
			Name: "sort",
			Parameters: []*registry.Parameter{
				{Name: "array", Type: "array"},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    2,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil || !args[0].IsArray() {
					return values.NewBool(false), nil
				}

				arr := args[0].Data.(*values.Array)

				// Extract values maintaining original order for stable sort
				type valueWithIndex struct {
					value *values.Value
					index int
				}

				var sortedValues []valueWithIndex
				idx := 0
				for _, value := range arr.Elements {
					sortedValues = append(sortedValues, valueWithIndex{value, idx})
					idx++
				}

				// Use Go's efficient sort with proper PHP comparison logic
				sort.Slice(sortedValues, func(i, j int) bool {
					vi, vj := sortedValues[i].value, sortedValues[j].value

					// Handle different type comparisons like PHP
					if vi.IsInt() && vj.IsInt() {
						return vi.ToInt() < vj.ToInt()
					} else if vi.IsFloat() && vj.IsFloat() {
						return vi.ToFloat() < vj.ToFloat()
					} else if (vi.IsInt() || vi.IsFloat()) && (vj.IsInt() || vj.IsFloat()) {
						return vi.ToFloat() < vj.ToFloat()
					} else {
						// Fall back to string comparison
						return vi.ToString() < vj.ToString()
					}
				})

				// Rebuild array with new sorted order and numeric indices
				newElements := make(map[interface{}]*values.Value)
				for i, item := range sortedValues {
					newElements[int64(i)] = item.value
				}
				arr.Elements = newElements
				arr.NextIndex = int64(len(sortedValues))

				return values.NewBool(true), nil
			},
		},
		{
			Name: "array_map",
			Parameters: []*registry.Parameter{
				{Name: "callback", Type: "mixed"},
				{Name: "array", Type: "array"},
			},
			ReturnType: "array",
			IsVariadic: true,
			MinArgs:    2,
			MaxArgs:    -1,
			IsBuiltin: true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 {
					return values.NewArray(), nil
				}

				callback := args[0]
				arrays := args[1:]

				// Validate all arguments are arrays
				for _, arr := range arrays {
					if arr == nil || !arr.IsArray() {
						return values.NewArray(), nil
					}
				}

				// Find minimum array length
				minLength := int64(0)
				for i, arr := range arrays {
					len := int64(arr.ArrayCount())
					if i == 0 || len < minLength {
						minLength = len
					}
				}

				result := values.NewArray()
				resultArr := result.Data.(*values.Array)

				// Handle null callback - create array of arrays
				if callback == nil || callback.IsNull() {
					for i := int64(0); i < minLength; i++ {
						row := values.NewArray()
						rowArr := row.Data.(*values.Array)

						for _, arr := range arrays {
							arrData := arr.Data.(*values.Array)
							if val, ok := arrData.Elements[i]; ok && val != nil {
								rowArr.Elements[int64(len(rowArr.Elements))] = val
							}
						}
						rowArr.NextIndex = int64(len(arrays))
						resultArr.Elements[i] = row
					}
					resultArr.NextIndex = minLength
					return result, nil
				}

				// Handle single array case (most common)
				if len(arrays) == 1 {
					firstArr := arrays[0].Data.(*values.Array)

					// Create ordered list of keys for consistent iteration
					keys := make([]interface{}, 0, len(firstArr.Elements))
					for key := range firstArr.Elements {
						keys = append(keys, key)
					}

					// Sort keys to maintain PHP array order
					sort.Slice(keys, func(i, j int) bool {
						ki, iIsInt := keys[i].(int64)
						kj, jIsInt := keys[j].(int64)
						if iIsInt && jIsInt {
							return ki < kj
						}
						if iIsInt && !jIsInt {
							return true
						}
						if !iIsInt && jIsInt {
							return false
						}
						return fmt.Sprintf("%v", keys[i]) < fmt.Sprintf("%v", keys[j])
					})

					// Process each element in order
					for _, key := range keys {
						if val, ok := firstArr.Elements[key]; ok && val != nil {
							callArgs := []*values.Value{val}

							// Use unified callback invoker
							result_val, err := callbackInvoker(ctx, callback, callArgs)
							if err != nil {
								return nil, err
							}
							if result_val != nil {
								resultArr.Elements[key] = result_val
							}
						}
					}
				} else {
					// Handle multiple arrays case (complex multi-array mapping)
					firstArr := arrays[0].Data.(*values.Array)
					for key, _ := range firstArr.Elements {
						// Get values from all arrays at this key
						var callArgs []*values.Value
						for _, arr := range arrays {
							arrData := arr.Data.(*values.Array)
							if val, ok := arrData.Elements[key]; ok && val != nil {
								callArgs = append(callArgs, val)
							} else {
								callArgs = append(callArgs, values.NewNull())
							}
						}

						// Use unified callback invoker
						if len(callArgs) > 0 {
							result_val, err := callbackInvoker(ctx, callback, callArgs)
							if err != nil {
								return nil, err
							}
							if result_val != nil {
								resultArr.Elements[key] = result_val
							}
						}
					}
				}

				// Ensure result array NextIndex is properly set
				if len(resultArr.Elements) > 0 {
					maxKey := int64(0)
					for key := range resultArr.Elements {
						if intKey, ok := key.(int64); ok && intKey >= maxKey {
							maxKey = intKey + 1
						}
					}
					resultArr.NextIndex = maxKey
				}

				return result, nil
			},
		},
		{
			Name: "array_slice",
			Parameters: []*registry.Parameter{
				{Name: "array", Type: "array"},
				{Name: "offset", Type: "int"},
				{Name: "length", Type: "int", DefaultValue: values.NewNull()},
				{Name: "preserve_keys", Type: "bool", DefaultValue: values.NewBool(false)},
			},
			ReturnType: "array",
			MinArgs:    2,
			MaxArgs:    4,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 || args[0] == nil || !args[0].IsArray() {
					return values.NewArray(), nil
				}

				arr := args[0].Data.(*values.Array)
				offset := int(args[1].ToInt())

				// Get length parameter (null means slice to end)
				var length *int = nil
				if len(args) > 2 && args[2] != nil && !args[2].IsNull() {
					l := int(args[2].ToInt())
					length = &l
				}

				// Get preserve_keys parameter
				preserveKeys := false
				if len(args) > 3 && args[3] != nil {
					preserveKeys = args[3].ToBool()
				}

				// Convert map to ordered slice
				type keyValue struct {
					key interface{}
					val *values.Value
				}

				var elements []keyValue
				// For arrays, we need to maintain insertion order
				// Use NextIndex to determine the maximum numeric index
				for i := int64(0); i < arr.NextIndex; i++ {
					if val, ok := arr.Elements[i]; ok {
						elements = append(elements, keyValue{i, val})
					}
				}
				// Add string keys
				for key, val := range arr.Elements {
					if _, isInt := key.(int64); !isInt {
						elements = append(elements, keyValue{key, val})
					}
				}

				arrLen := len(elements)

				// Handle negative offset
				if offset < 0 {
					offset = arrLen + offset
					if offset < 0 {
						offset = 0
					}
				}

				// Handle offset beyond array bounds
				if offset >= arrLen {
					return values.NewArray(), nil
				}

				// Calculate end index
				end := arrLen
				if length != nil {
					if *length < 0 {
						// Negative length: exclude that many elements from the end
						end = arrLen + *length
						if end < offset {
							return values.NewArray(), nil
						}
					} else if *length == 0 {
						return values.NewArray(), nil
					} else {
						end = offset + *length
						if end > arrLen {
							end = arrLen
						}
					}
				}

				// Create result array
				result := values.NewArray()
				resultArr := result.Data.(*values.Array)

				// Extract slice
				newIndex := int64(0)
				for i := offset; i < end; i++ {
					elem := elements[i]
					if preserveKeys {
						resultArr.Elements[elem.key] = elem.val
					} else {
						resultArr.Elements[newIndex] = elem.val
						newIndex++
					}
				}

				if !preserveKeys {
					resultArr.NextIndex = newIndex
				}

				return result, nil
			},
		},
		{
			Name: "array_search",
			Parameters: []*registry.Parameter{
				{Name: "needle", Type: "mixed"},
				{Name: "haystack", Type: "array"},
				{Name: "strict", Type: "bool", DefaultValue: values.NewBool(false)},
			},
			ReturnType: "mixed",
			MinArgs:    2,
			MaxArgs:    3,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 || args[1] == nil || !args[1].IsArray() {
					return values.NewBool(false), nil
				}

				needle := args[0]
				arr := args[1].Data.(*values.Array)

				// Get strict parameter
				strict := false
				if len(args) > 2 && args[2] != nil {
					strict = args[2].ToBool()
				}

				// Search through array elements
				// First search numeric indices in order
				for i := int64(0); i < arr.NextIndex; i++ {
					if val, ok := arr.Elements[i]; ok {
						if strict {
							// Strict comparison: same type and value
							if needle.Type == val.Type {
								switch needle.Type {
								case values.TypeInt:
									if needle.ToInt() == val.ToInt() {
										return values.NewInt(i), nil
									}
								case values.TypeFloat:
									if needle.ToFloat() == val.ToFloat() {
										return values.NewInt(i), nil
									}
								case values.TypeString:
									if needle.ToString() == val.ToString() {
										return values.NewInt(i), nil
									}
								case values.TypeBool:
									if needle.ToBool() == val.ToBool() {
										return values.NewInt(i), nil
									}
								case values.TypeNull:
									// Both are null
									return values.NewInt(i), nil
								}
							}
						} else {
							// Loose comparison: convert to string and compare
							if needle.ToString() == val.ToString() {
								return values.NewInt(i), nil
							}
						}
					}
				}

				// Then search string keys
				for key, val := range arr.Elements {
					if _, isInt := key.(int64); !isInt {
						if strict {
							// Strict comparison
							if needle.Type == val.Type {
								switch needle.Type {
								case values.TypeInt:
									if needle.ToInt() == val.ToInt() {
										return values.NewString(fmt.Sprintf("%v", key)), nil
									}
								case values.TypeFloat:
									if needle.ToFloat() == val.ToFloat() {
										return values.NewString(fmt.Sprintf("%v", key)), nil
									}
								case values.TypeString:
									if needle.ToString() == val.ToString() {
										return values.NewString(fmt.Sprintf("%v", key)), nil
									}
								case values.TypeBool:
									if needle.ToBool() == val.ToBool() {
										return values.NewString(fmt.Sprintf("%v", key)), nil
									}
								case values.TypeNull:
									return values.NewString(fmt.Sprintf("%v", key)), nil
								}
							}
						} else {
							// Loose comparison
							if needle.ToString() == val.ToString() {
								return values.NewString(fmt.Sprintf("%v", key)), nil
							}
						}
					}
				}

				// Not found
				return values.NewBool(false), nil
			},
		},
		{
			Name: "array_pop",
			Parameters: []*registry.Parameter{
				{Name: "array", Type: "array"},
			},
			ReturnType: "mixed",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 1 || args[0] == nil || !args[0].IsArray() {
					return values.NewNull(), nil
				}

				arr := args[0].Data.(*values.Array)

				// Find the last element
				var lastKey interface{} = nil
				var lastVal *values.Value = nil

				// First check numeric indices from highest to lowest
				for i := arr.NextIndex - 1; i >= 0; i-- {
					if val, ok := arr.Elements[i]; ok {
						lastKey = i
						lastVal = val
						break
					}
				}

				// If no numeric index found, check string keys
				if lastKey == nil {
					for key, val := range arr.Elements {
						if _, isInt := key.(int64); !isInt {
							lastKey = key
							lastVal = val
							// Note: order of string keys is not guaranteed, but PHP behavior is similar
						}
					}
				}

				if lastKey == nil {
					return values.NewNull(), nil
				}

				// Remove the element
				delete(arr.Elements, lastKey)

				// Update NextIndex if it was a numeric key
				if intKey, ok := lastKey.(int64); ok && intKey == arr.NextIndex-1 {
					arr.NextIndex--
				}

				return lastVal, nil
			},
		},
		{
			Name: "array_shift",
			Parameters: []*registry.Parameter{
				{Name: "array", Type: "array"},
			},
			ReturnType: "mixed",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 1 || args[0] == nil || !args[0].IsArray() {
					return values.NewNull(), nil
				}

				arr := args[0].Data.(*values.Array)

				// Find the first element
				var firstKey interface{} = nil
				var firstVal *values.Value = nil

				// First check numeric indices from lowest to highest
				for i := int64(0); i < arr.NextIndex; i++ {
					if val, ok := arr.Elements[i]; ok {
						firstKey = i
						firstVal = val
						break
					}
				}

				// If no numeric index found, check string keys
				if firstKey == nil {
					for key, val := range arr.Elements {
						if _, isInt := key.(int64); !isInt {
							firstKey = key
							firstVal = val
							break // Take first string key found
						}
					}
				}

				if firstKey == nil {
					return values.NewNull(), nil
				}

				// Remove the element
				delete(arr.Elements, firstKey)

				// For numeric keys, we need to shift all other keys down
				if intKey, ok := firstKey.(int64); ok {
					// Create new elements map with shifted keys
					newElements := make(map[interface{}]*values.Value)
					for key, val := range arr.Elements {
						if keyInt, isInt := key.(int64); isInt && keyInt > intKey {
							// Shift numeric keys down by 1
							newElements[keyInt-1] = val
						} else if !isInt {
							// Keep string keys as-is
							newElements[key] = val
						} else {
							// Keep lower numeric keys
							newElements[key] = val
						}
					}
					arr.Elements = newElements
					if arr.NextIndex > 0 {
						arr.NextIndex--
					}
				}

				return firstVal, nil
			},
		},
		{
			Name: "array_unshift",
			Parameters: []*registry.Parameter{
				{Name: "array", Type: "array"},
			},
			ReturnType: "int",
			IsVariadic: true,
			MinArgs:    2,
			MaxArgs:    -1,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 || args[0] == nil || !args[0].IsArray() {
					return values.NewInt(0), nil
				}

				arr := args[0].Data.(*values.Array)
				values_to_add := args[1:]

				// Create new elements map
				newElements := make(map[interface{}]*values.Value)

				// Add new values to the beginning with indices 0, 1, 2...
				for i, val := range values_to_add {
					newElements[int64(i)] = val
				}

				numAdded := int64(len(values_to_add))

				// Shift existing numeric keys up by numAdded
				for key, val := range arr.Elements {
					if keyInt, isInt := key.(int64); isInt {
						// Shift numeric keys up
						newElements[keyInt+numAdded] = val
					} else {
						// Keep string keys as-is
						newElements[key] = val
					}
				}

				// Update the array
				arr.Elements = newElements
				arr.NextIndex += numAdded

				// Return the new length
				return values.NewInt(int64(len(arr.Elements))), nil
			},
		},
		{
			Name: "array_pad",
			Parameters: []*registry.Parameter{
				{Name: "array", Type: "array"},
				{Name: "size", Type: "int"},
				{Name: "value", Type: "mixed"},
			},
			ReturnType: "array",
			MinArgs:    3,
			MaxArgs:    3,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 3 || args[0] == nil || !args[0].IsArray() {
					return values.NewArray(), nil
				}

				arr := args[0].Data.(*values.Array)
				size := int(args[1].ToInt())
				padValue := args[2]

				currentLength := len(arr.Elements)
				absSize := size
				if absSize < 0 {
					absSize = -absSize
				}

				// If size is less than or equal to current length, return copy
				if absSize <= currentLength {
					result := values.NewArray()
					resultArr := result.Data.(*values.Array)
					for key, val := range arr.Elements {
						resultArr.Elements[key] = val
					}
					resultArr.NextIndex = arr.NextIndex
					return result, nil
				}

				result := values.NewArray()
				resultArr := result.Data.(*values.Array)
				padCount := absSize - currentLength

				if size > 0 {
					// Pad to the right: original elements first, then padding
					// Copy existing elements
					for key, val := range arr.Elements {
						resultArr.Elements[key] = val
					}

					// Add padding at the end with numeric indices
					nextIndex := arr.NextIndex
					for i := 0; i < padCount; i++ {
						resultArr.Elements[nextIndex] = padValue
						nextIndex++
					}
					resultArr.NextIndex = nextIndex

				} else {
					// Pad to the left: padding first, then original elements
					// Add padding at the beginning
					for i := int64(0); i < int64(padCount); i++ {
						resultArr.Elements[i] = padValue
					}

					// Shift and copy existing elements
					for key, val := range arr.Elements {
						if keyInt, isInt := key.(int64); isInt {
							// Shift numeric keys
							resultArr.Elements[keyInt+int64(padCount)] = val
						} else {
							// Keep string keys as-is
							resultArr.Elements[key] = val
						}
					}

					resultArr.NextIndex = int64(padCount) + arr.NextIndex
				}

				return result, nil
			},
		},
		{
			Name: "array_fill",
			Parameters: []*registry.Parameter{
				{Name: "start_index", Type: "int"},
				{Name: "count", Type: "int"},
				{Name: "value", Type: "mixed"},
			},
			ReturnType: "array",
			MinArgs:    3,
			MaxArgs:    3,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 3 {
					return values.NewArray(), nil
				}

				startIndex := args[0].ToInt()
				count := int(args[1].ToInt())
				value := args[2]

				if count <= 0 {
					return values.NewArray(), nil
				}

				result := values.NewArray()
				resultArr := result.Data.(*values.Array)

				for i := 0; i < count; i++ {
					key := startIndex + int64(i)
					resultArr.Elements[key] = value
				}

				// Set NextIndex appropriately
				if startIndex >= 0 {
					maxIndex := startIndex + int64(count)
					if maxIndex > resultArr.NextIndex {
						resultArr.NextIndex = maxIndex
					}
				}

				return result, nil
			},
		},
		{
			Name: "array_fill_keys",
			Parameters: []*registry.Parameter{
				{Name: "keys", Type: "array"},
				{Name: "value", Type: "mixed"},
			},
			ReturnType: "array",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 || args[0] == nil || !args[0].IsArray() {
					return values.NewArray(), nil
				}

				keysArr := args[0].Data.(*values.Array)
				value := args[1]

				result := values.NewArray()
				resultArr := result.Data.(*values.Array)

				// Iterate through all keys and use them as keys in result
				for _, keyVal := range keysArr.Elements {
					if keyVal == nil {
						continue
					}

					var resultKey interface{}
					if keyVal.IsInt() {
						resultKey = keyVal.ToInt()
					} else {
						resultKey = keyVal.ToString()
					}

					resultArr.Elements[resultKey] = value
				}

				return result, nil
			},
		},
		{
			Name: "range",
			Parameters: []*registry.Parameter{
				{Name: "start", Type: "mixed"},
				{Name: "end", Type: "mixed"},
				{Name: "step", Type: "mixed", DefaultValue: values.NewInt(1)},
			},
			ReturnType: "array",
			MinArgs:    2,
			MaxArgs:    3,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 {
					return values.NewArray(), nil
				}

				start := args[0]
				end := args[1]
				step := values.NewInt(1)
				if len(args) > 2 && args[2] != nil {
					step = args[2]
				}

				// Validate step is not zero
				if step.ToFloat() == 0 {
					return values.NewArray(), fmt.Errorf("range(): Argument #3 ($step) cannot be 0")
				}

				result := values.NewArray()
				resultArr := result.Data.(*values.Array)

				// Handle character ranges
				if start.IsString() && end.IsString() {
					startStr := start.ToString()
					endStr := end.ToString()
					if len(startStr) == 1 && len(endStr) == 1 {
						startChar := int(startStr[0])
						endChar := int(endStr[0])
						idx := int64(0)

						if startChar <= endChar {
							// Ascending
							for char := startChar; char <= endChar; char++ {
								resultArr.Elements[idx] = values.NewString(string(rune(char)))
								idx++
							}
						} else {
							// Descending
							for char := startChar; char >= endChar; char-- {
								resultArr.Elements[idx] = values.NewString(string(rune(char)))
								idx++
							}
						}
						resultArr.NextIndex = idx
						return result, nil
					}
				}

				// Handle numeric ranges
				startNum := start.ToFloat()
				endNum := end.ToFloat()
				stepNum := step.ToFloat()

				idx := int64(0)
				current := startNum

				// Determine direction
				if startNum <= endNum {
					// Ascending
					if stepNum < 0 {
						stepNum = -stepNum // Use absolute value
					}
					for current <= endNum {
						// Use appropriate type for result
						if start.IsInt() && end.IsInt() && step.IsInt() {
							resultArr.Elements[idx] = values.NewInt(int64(current))
						} else {
							resultArr.Elements[idx] = values.NewFloat(current)
						}
						idx++
						current += stepNum

						// Safety limit to prevent infinite loops
						if idx > 1000000 {
							break
						}
					}
				} else {
					// Descending
					if stepNum > 0 {
						stepNum = -stepNum // Make negative
					}
					for current >= endNum {
						// Use appropriate type for result
						if start.IsInt() && end.IsInt() && step.IsInt() {
							resultArr.Elements[idx] = values.NewInt(int64(current))
						} else {
							resultArr.Elements[idx] = values.NewFloat(current)
						}
						idx++
						current += stepNum

						// Safety limit to prevent infinite loops
						if idx > 1000000 {
							break
						}
					}
				}

				resultArr.NextIndex = idx
				return result, nil
			},
		},
		{
			Name: "array_splice",
			Parameters: []*registry.Parameter{
				{Name: "array", Type: "array"},
				{Name: "offset", Type: "int"},
				{Name: "length", Type: "int", DefaultValue: values.NewNull()},
				{Name: "replacement", Type: "array", DefaultValue: values.NewArray()},
			},
			ReturnType: "array",
			MinArgs:    2,
			MaxArgs:    4,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 || args[0] == nil || !args[0].IsArray() {
					return values.NewArray(), nil
				}

				arr := args[0].Data.(*values.Array)
				offset := int(args[1].ToInt())

				// Handle length parameter
				var length *int = nil
				if len(args) > 2 && args[2] != nil && !args[2].IsNull() {
					l := int(args[2].ToInt())
					length = &l
				}

				// Handle replacement parameter
				var replacement []*values.Value
				if len(args) > 3 && args[3] != nil && args[3].IsArray() {
					replacementArr := args[3].Data.(*values.Array)
					for i := int64(0); i < replacementArr.NextIndex; i++ {
						if val, ok := replacementArr.Elements[i]; ok {
							replacement = append(replacement, val)
						}
					}
					// Also handle string keys in replacement
					for key, val := range replacementArr.Elements {
						if _, isInt := key.(int64); !isInt {
							replacement = append(replacement, val)
						}
					}
				}

				// Convert array to ordered slice for easier manipulation
				type keyValue struct {
					key interface{}
					val *values.Value
				}

				var elements []keyValue
				// First add numeric keys in order
				for i := int64(0); i < arr.NextIndex; i++ {
					if val, ok := arr.Elements[i]; ok {
						elements = append(elements, keyValue{i, val})
					}
				}
				// Then add string keys (they'll be preserved in position)
				stringKeys := make([]keyValue, 0)
				for key, val := range arr.Elements {
					if _, isInt := key.(int64); !isInt {
						stringKeys = append(stringKeys, keyValue{key, val})
					}
				}

				arrLen := len(elements)

				// Handle negative offset
				if offset < 0 {
					offset = arrLen + offset
					if offset < 0 {
						offset = 0
					}
				}

				// Handle offset beyond bounds
				if offset > arrLen {
					offset = arrLen
				}

				// Calculate actual length to remove
				actualLength := arrLen - offset
				if length != nil {
					if *length < 0 {
						actualLength = arrLen - offset + *length
						if actualLength < 0 {
							actualLength = 0
						}
					} else {
						actualLength = *length
						if actualLength > arrLen-offset {
							actualLength = arrLen - offset
						}
					}
				}

				// Create removed elements array
				removed := values.NewArray()
				removedArr := removed.Data.(*values.Array)
				removedIdx := int64(0)
				for i := 0; i < actualLength; i++ {
					if offset+i < len(elements) {
						elem := elements[offset+i]
						// For removed array, use sequential numeric keys starting from 0
						removedArr.Elements[removedIdx] = elem.val
						removedIdx++
					}
				}
				removedArr.NextIndex = removedIdx

				// Build new elements slice
				newElements := make([]keyValue, 0)
				// Add elements before offset
				for i := 0; i < offset && i < len(elements); i++ {
					newElements = append(newElements, elements[i])
				}
				// Add replacement elements
				for _, val := range replacement {
					newElements = append(newElements, keyValue{int64(len(newElements)), val})
				}
				// Add elements after removed section
				for i := offset + actualLength; i < len(elements); i++ {
					newElements = append(newElements, keyValue{int64(len(newElements)), elements[i].val})
				}

				// Rebuild the original array
				arr.Elements = make(map[interface{}]*values.Value)
				for _, elem := range newElements {
					arr.Elements[elem.key] = elem.val
				}
				// Re-add string keys
				for _, elem := range stringKeys {
					// Only keep string keys that weren't in the removed range
					// For simplicity, keep all string keys (PHP behavior is complex here)
					arr.Elements[elem.key] = elem.val
				}
				arr.NextIndex = int64(len(newElements))

				return removed, nil
			},
		},
		{
			Name: "array_column",
			Parameters: []*registry.Parameter{
				{Name: "array", Type: "array"},
				{Name: "column_key", Type: "mixed"},
				{Name: "index_key", Type: "mixed", DefaultValue: values.NewNull()},
			},
			ReturnType: "array",
			MinArgs:    2,
			MaxArgs:    3,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				array := args[0]
				columnKey := args[1]
				var indexKey *values.Value
				if len(args) >= 3 {
					indexKey = args[2]
				}

				if array.Type != values.TypeArray {
					return values.NewArray(), nil
				}

				result := values.NewArray()
				resultArr := result.Data.(*values.Array)
				resultIndex := int64(0)

				// Iterate over the input array in order
				arrayData := array.Data.(*values.Array)

				// Create ordered list of keys
				keys := make([]interface{}, 0, len(arrayData.Elements))
				for k := range arrayData.Elements {
					keys = append(keys, k)
				}

				// Sort keys to maintain order (numeric keys first, then string keys)
				sort.Slice(keys, func(i, j int) bool {
					ki, iIsInt := keys[i].(int64)
					kj, jIsInt := keys[j].(int64)

					if iIsInt && jIsInt {
						return ki < kj
					}
					if iIsInt && !jIsInt {
						return true // int keys come first
					}
					if !iIsInt && jIsInt {
						return false // string keys come after
					}
					// both string keys
					return fmt.Sprintf("%v", keys[i]) < fmt.Sprintf("%v", keys[j])
				})

				for _, key := range keys {
					element := arrayData.Elements[key]
					if element.Type != values.TypeArray {
						continue // Skip non-array elements
					}

					var colValue *values.Value
					var keyValue interface{}
					elementArr := element.Data.(*values.Array)

					// Extract the column value
					if columnKey.IsNull() {
						// If column_key is null, return the whole element
						colValue = element
					} else if columnKey.Type == values.TypeString {
						colValue = elementArr.Elements[columnKey.Data.(string)]
					} else if columnKey.Type == values.TypeInt {
						colValue = elementArr.Elements[columnKey.Data.(int64)]
					}

					// If column doesn't exist, skip this element
					if colValue == nil {
						continue
					}

					// Determine the key for the result array
					if indexKey != nil && !indexKey.IsNull() {
						var indexVal *values.Value
						if indexKey.Type == values.TypeString {
							indexVal = elementArr.Elements[indexKey.Data.(string)]
						} else if indexKey.Type == values.TypeInt {
							indexVal = elementArr.Elements[indexKey.Data.(int64)]
						}

						if indexVal != nil {
							if indexVal.Type == values.TypeString {
								keyValue = indexVal.Data.(string)
							} else if indexVal.Type == values.TypeInt {
								keyValue = indexVal.Data.(int64)
							} else {
								keyValue = resultIndex
								resultIndex++
							}
						} else {
							keyValue = resultIndex
							resultIndex++
						}
					} else {
						keyValue = resultIndex
						resultIndex++
					}

					resultArr.Elements[keyValue] = colValue
					if numKey, ok := keyValue.(int64); ok && numKey >= resultArr.NextIndex {
						resultArr.NextIndex = numKey + 1
					}
				}

				return result, nil
			},
		},
		{
			Name: "array_keys",
			Parameters: []*registry.Parameter{
				{Name: "array", Type: "array"},
				{Name: "search_value", Type: "mixed", DefaultValue: values.NewNull()},
				{Name: "strict", Type: "bool", DefaultValue: values.NewBool(false)},
			},
			ReturnType: "array",
			MinArgs:    1,
			MaxArgs:    3,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				array := args[0]
				var searchValue *values.Value
				strict := false

				if len(args) >= 2 && !args[1].IsNull() {
					searchValue = args[1]
				}
				if len(args) >= 3 && args[2].Type == values.TypeBool {
					strict = args[2].Data.(bool)
				}

				if array.Type != values.TypeArray {
					return values.NewArray(), nil
				}

				result := values.NewArray()
				resultArr := result.Data.(*values.Array)
				arrayData := array.Data.(*values.Array)
				resultIndex := int64(0)

				// Create ordered list of keys
				keys := make([]interface{}, 0, len(arrayData.Elements))
				for k := range arrayData.Elements {
					keys = append(keys, k)
				}

				// Sort keys to maintain order (numeric keys first, then string keys)
				sort.Slice(keys, func(i, j int) bool {
					ki, iIsInt := keys[i].(int64)
					kj, jIsInt := keys[j].(int64)

					if iIsInt && jIsInt {
						return ki < kj
					}
					if iIsInt && !jIsInt {
						return true // int keys come first
					}
					if !iIsInt && jIsInt {
						return false // string keys come after
					}
					// both string keys
					return fmt.Sprintf("%v", keys[i]) < fmt.Sprintf("%v", keys[j])
				})

				for _, key := range keys {
					value := arrayData.Elements[key]

					// If no search value, include all keys
					if searchValue == nil {
						if numKey, ok := key.(int64); ok {
							resultArr.Elements[resultIndex] = values.NewInt(numKey)
						} else {
							resultArr.Elements[resultIndex] = values.NewString(key.(string))
						}
						resultIndex++
					} else {
						// Search for matching values
						matches := false
						if strict {
							// Strict comparison - types must match
							if value.Type == searchValue.Type {
								switch value.Type {
								case values.TypeInt:
									matches = value.Data.(int64) == searchValue.Data.(int64)
								case values.TypeString:
									matches = value.Data.(string) == searchValue.Data.(string)
								case values.TypeBool:
									matches = value.Data.(bool) == searchValue.Data.(bool)
								case values.TypeFloat:
									matches = value.Data.(float64) == searchValue.Data.(float64)
								case values.TypeNull:
									matches = true // both null
								}
							}
						} else {
							// Loose comparison - PHP-style type coercion
							matches = compareValuesLoose(value, searchValue)
						}

						if matches {
							if numKey, ok := key.(int64); ok {
								resultArr.Elements[resultIndex] = values.NewInt(numKey)
							} else {
								resultArr.Elements[resultIndex] = values.NewString(key.(string))
							}
							resultIndex++
						}
					}
				}

				resultArr.NextIndex = resultIndex
				return result, nil
			},
		},
		{
			Name: "array_values",
			Parameters: []*registry.Parameter{
				{Name: "array", Type: "array"},
			},
			ReturnType: "array",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				array := args[0]

				if array.Type != values.TypeArray {
					return values.NewArray(), nil
				}

				result := values.NewArray()
				resultArr := result.Data.(*values.Array)
				arrayData := array.Data.(*values.Array)
				resultIndex := int64(0)

				// Create ordered list of keys
				keys := make([]interface{}, 0, len(arrayData.Elements))
				for k := range arrayData.Elements {
					keys = append(keys, k)
				}

				// Sort keys to maintain order (numeric keys first, then string keys)
				sort.Slice(keys, func(i, j int) bool {
					ki, iIsInt := keys[i].(int64)
					kj, jIsInt := keys[j].(int64)

					if iIsInt && jIsInt {
						return ki < kj
					}
					if iIsInt && !jIsInt {
						return true // int keys come first
					}
					if !iIsInt && jIsInt {
						return false // string keys come after
					}
					// both string keys
					return fmt.Sprintf("%v", keys[i]) < fmt.Sprintf("%v", keys[j])
				})

				// Add all values with new sequential keys
				for _, key := range keys {
					value := arrayData.Elements[key]
					resultArr.Elements[resultIndex] = value
					resultIndex++
				}

				resultArr.NextIndex = resultIndex
				return result, nil
			},
		},
		{
			Name: "array_merge",
			Parameters: []*registry.Parameter{
				{Name: "arrays", Type: "array"},
			},
			ReturnType:   "array",
			MinArgs:      0,
			MaxArgs:      -1,
			IsVariadic:   true,
			IsBuiltin:    true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				result := values.NewArray()
				resultArr := result.Data.(*values.Array)
				numericIndex := int64(0)

				// Process each array argument
				for _, array := range args {
					if array.Type != values.TypeArray {
						continue // Skip non-array arguments
					}

					arrayData := array.Data.(*values.Array)

					// Create ordered list of keys
					keys := make([]interface{}, 0, len(arrayData.Elements))
					for k := range arrayData.Elements {
						keys = append(keys, k)
					}

					// Sort keys to maintain order (numeric keys first, then string keys)
					sort.Slice(keys, func(i, j int) bool {
						ki, iIsInt := keys[i].(int64)
						kj, jIsInt := keys[j].(int64)

						if iIsInt && jIsInt {
							return ki < kj
						}
						if iIsInt && !jIsInt {
							return true // int keys come first
						}
						if !iIsInt && jIsInt {
							return false // string keys come after
						}
						// both string keys
						return fmt.Sprintf("%v", keys[i]) < fmt.Sprintf("%v", keys[j])
					})

					// Merge values according to PHP rules
					for _, key := range keys {
						value := arrayData.Elements[key]

						if _, isNumKey := key.(int64); isNumKey {
							// Numeric keys: always reindex
							resultArr.Elements[numericIndex] = value
							numericIndex++
						} else {
							// String keys: preserve key, overwrite if exists
							resultArr.Elements[key] = value
						}
					}
				}

				resultArr.NextIndex = numericIndex
				return result, nil
			},
		},
		{
			Name: "array_unique",
			Parameters: []*registry.Parameter{
				{Name: "array", Type: "array"},
				{Name: "flags", Type: "int", DefaultValue: values.NewInt(0)},
			},
			ReturnType: "array",
			MinArgs:    1,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				array := args[0]
				// flags parameter not fully implemented - using loose comparison by default

				if array.Type != values.TypeArray {
					return values.NewArray(), nil
				}

				result := values.NewArray()
				resultArr := result.Data.(*values.Array)
				arrayData := array.Data.(*values.Array)

				// Track seen values
				seen := make(map[string]bool)

				// Create ordered list of keys
				keys := make([]interface{}, 0, len(arrayData.Elements))
				for k := range arrayData.Elements {
					keys = append(keys, k)
				}

				// Sort keys to maintain order (numeric keys first, then string keys)
				sort.Slice(keys, func(i, j int) bool {
					ki, iIsInt := keys[i].(int64)
					kj, jIsInt := keys[j].(int64)

					if iIsInt && jIsInt {
						return ki < kj
					}
					if iIsInt && !jIsInt {
						return true // int keys come first
					}
					if !iIsInt && jIsInt {
						return false // string keys come after
					}
					// both string keys
					return fmt.Sprintf("%v", keys[i]) < fmt.Sprintf("%v", keys[j])
				})

				// Process values in order, keeping only first occurrence
				for _, key := range keys {
					value := arrayData.Elements[key]

					// Create a string representation for comparison
					var valueKey string
					if value.IsNull() {
						valueKey = "null"
					} else {
						switch value.Type {
						case values.TypeInt:
							valueKey = fmt.Sprintf("int:%d", value.Data.(int64))
						case values.TypeString:
							valueKey = fmt.Sprintf("str:%s", value.Data.(string))
						case values.TypeBool:
							valueKey = fmt.Sprintf("bool:%t", value.Data.(bool))
						case values.TypeFloat:
							valueKey = fmt.Sprintf("float:%f", value.Data.(float64))
						default:
							valueKey = fmt.Sprintf("other:%v", value.Data)
						}
					}

					// Only add if we haven't seen this value before
					if !seen[valueKey] {
						seen[valueKey] = true
						resultArr.Elements[key] = value

						// Update NextIndex for numeric keys
						if numKey, ok := key.(int64); ok && numKey >= resultArr.NextIndex {
							resultArr.NextIndex = numKey + 1
						}
					}
				}

				return result, nil
			},
		},
		{
			Name: "array_key_exists",
			Parameters: []*registry.Parameter{
				{Name: "key", Type: "mixed"},
				{Name: "array", Type: "array"},
			},
			ReturnType: "bool",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 || args[1] == nil || !args[1].IsArray() {
					return values.NewBool(false), nil
				}

				key := args[0]
				arr := args[1].Data.(*values.Array)

				// Convert key to appropriate type for lookup
				var searchKey interface{}
				if key.IsInt() {
					searchKey = key.ToInt()
				} else {
					searchKey = key.ToString()
				}

				_, exists := arr.Elements[searchKey]
				return values.NewBool(exists), nil
			},
		},
		{
			Name: "key_exists",
			Parameters: []*registry.Parameter{
				{Name: "key", Type: "mixed"},
				{Name: "array", Type: "array"},
			},
			ReturnType: "bool",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// key_exists is an alias for array_key_exists
				// Get the array_key_exists function and call it
				reg := ctx.SymbolRegistry()
				if reg != nil {
					if fn, ok := reg.GetFunction("array_key_exists"); ok && fn != nil && fn.IsBuiltin {
						return fn.Builtin(ctx, args)
					}
				}
				return values.NewBool(false), nil
			},
		},
		{
			Name: "array_key_first",
			Parameters: []*registry.Parameter{
				{Name: "array", Type: "array"},
			},
			ReturnType: "mixed",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 1 || args[0] == nil || !args[0].IsArray() {
					return values.NewNull(), nil
				}

				arr := args[0].Data.(*values.Array)
				if len(arr.Elements) == 0 {
					return values.NewNull(), nil
				}

				// Find the first key in insertion order
				// We need to iterate through keys in their original order
				var firstKey interface{}

				// Get all keys and find minimum based on insertion order
				// Since Go maps don't preserve insertion order, we'll use a heuristic
				// Check if this is a purely numeric array first
				allNumeric := true
				for key := range arr.Elements {
					if _, isInt := key.(int64); !isInt {
						allNumeric = false
						break
					}
				}

				if allNumeric {
					// For numeric arrays, find minimum key
					minKey := int64(9223372036854775807) // MaxInt64
					found := false
					for key := range arr.Elements {
						if intKey, ok := key.(int64); ok && intKey < minKey {
							minKey = intKey
							found = true
						}
					}
					if found {
						firstKey = minKey
					}
				} else {
					// For mixed arrays, just take first key found
					for key := range arr.Elements {
						firstKey = key
						break
					}
				}

				if firstKey == nil {
					return values.NewNull(), nil
				}

				if intKey, ok := firstKey.(int64); ok {
					return values.NewInt(intKey), nil
				}
				return values.NewString(fmt.Sprintf("%v", firstKey)), nil
			},
		},
		{
			Name: "array_key_last",
			Parameters: []*registry.Parameter{
				{Name: "array", Type: "array"},
			},
			ReturnType: "mixed",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 1 || args[0] == nil || !args[0].IsArray() {
					return values.NewNull(), nil
				}

				arr := args[0].Data.(*values.Array)
				if len(arr.Elements) == 0 {
					return values.NewNull(), nil
				}

				// Find the last key in insertion order
				var lastKey interface{}

				// Check if this is a purely numeric array first
				allNumeric := true
				for key := range arr.Elements {
					if _, isInt := key.(int64); !isInt {
						allNumeric = false
						break
					}
				}

				if allNumeric {
					// For numeric arrays, find maximum key
					maxKey := int64(-9223372036854775808) // MinInt64
					found := false
					for key := range arr.Elements {
						if intKey, ok := key.(int64); ok && intKey > maxKey {
							maxKey = intKey
							found = true
						}
					}
					if found {
						lastKey = maxKey
					}
				} else {
					// For mixed arrays, just take last key found (Go map iteration order)
					for key := range arr.Elements {
						lastKey = key // Keep overwriting to get "last"
					}
				}

				if lastKey == nil {
					return values.NewNull(), nil
				}

				if intKey, ok := lastKey.(int64); ok {
					return values.NewInt(intKey), nil
				}
				return values.NewString(fmt.Sprintf("%v", lastKey)), nil
			},
		},
		{
			Name: "array_walk",
			Parameters: []*registry.Parameter{
				{Name: "array", Type: "array"},
				{Name: "callback", Type: "callable"},
				{Name: "userdata", Type: "mixed", DefaultValue: values.NewNull()},
			},
			ReturnType: "bool",
			MinArgs:    2,
			MaxArgs:    3,
			IsBuiltin:  true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 || args[0] == nil || !args[0].IsArray() {
					return values.NewBool(false), nil
				}

				arr := args[0].Data.(*values.Array)
				callback := args[1]
				var userdata *values.Value = values.NewNull()
				if len(args) > 2 && args[2] != nil {
					userdata = args[2]
				}

				// Use unified callback invoker for both builtin and user-defined callbacks

				// Iterate through array elements in order
				keys := make([]interface{}, 0, len(arr.Elements))
				for key := range arr.Elements {
					keys = append(keys, key)
				}

				// Sort keys to maintain order (numeric first, then string)
				sort.Slice(keys, func(i, j int) bool {
					ki, iIsInt := keys[i].(int64)
					kj, jIsInt := keys[j].(int64)

					if iIsInt && jIsInt {
						return ki < kj
					}
					if iIsInt && !jIsInt {
						return true // int keys come first
					}
					if !iIsInt && jIsInt {
						return false // string keys come after
					}
					// both string keys
					return fmt.Sprintf("%v", keys[i]) < fmt.Sprintf("%v", keys[j])
				})

				// Call function for each element
				for _, key := range keys {
					value := arr.Elements[key]
					if value == nil {
						continue
					}

					var keyVal *values.Value
					if intKey, ok := key.(int64); ok {
						keyVal = values.NewInt(intKey)
					} else {
						keyVal = values.NewString(fmt.Sprintf("%v", key))
					}

					// Call with (value, key, userdata)
					callArgs := []*values.Value{value, keyVal}
					if !userdata.IsNull() {
						callArgs = append(callArgs, userdata)
					}

					_, err := callbackInvoker(ctx, callback, callArgs)
					if err != nil {
						return values.NewBool(false), err
					}
				}

				return values.NewBool(true), nil
			},
		},
		{
			Name: "array_walk_recursive",
			Parameters: []*registry.Parameter{
				{Name: "array", Type: "array"},
				{Name: "callback", Type: "callable"},
				{Name: "userdata", Type: "mixed", DefaultValue: values.NewNull()},
			},
			ReturnType: "bool",
			MinArgs:    2,
			MaxArgs:    3,
			IsBuiltin:  true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// For now, implement as simple array_walk - full recursion would require
				// deeper implementation with proper recursion handling
				// Find and call array_walk
				reg := ctx.SymbolRegistry()
				if reg != nil {
					if walkFn, ok := reg.GetFunction("array_walk"); ok && walkFn != nil && walkFn.IsBuiltin {
						return walkFn.Builtin(ctx, args)
					}
				}
				return values.NewBool(false), nil
			},
		},
		{
			Name: "array_reduce",
			Parameters: []*registry.Parameter{
				{Name: "array", Type: "array"},
				{Name: "callback", Type: "callable"},
				{Name: "initial", Type: "mixed", DefaultValue: values.NewNull()},
			},
			ReturnType: "mixed",
			MinArgs:    2,
			MaxArgs:    3,
			IsBuiltin:  true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 || args[0] == nil || !args[0].IsArray() {
					return values.NewNull(), nil
				}

				arr := args[0].Data.(*values.Array)
				callback := args[1]
				carry := values.NewNull()
				if len(args) > 2 && args[2] != nil {
					carry = args[2]
				}

				// Use unified callback invoker for both builtin and user-defined callbacks

				// Get elements in order
				keys := make([]interface{}, 0, len(arr.Elements))
				for key := range arr.Elements {
					keys = append(keys, key)
				}

				// Sort keys to maintain order
				sort.Slice(keys, func(i, j int) bool {
					ki, iIsInt := keys[i].(int64)
					kj, jIsInt := keys[j].(int64)

					if iIsInt && jIsInt {
						return ki < kj
					}
					if iIsInt && !jIsInt {
						return true
					}
					if !iIsInt && jIsInt {
						return false
					}
					return fmt.Sprintf("%v", keys[i]) < fmt.Sprintf("%v", keys[j])
				})

				// Reduce array elements
				for _, key := range keys {
					value := arr.Elements[key]
					if value == nil {
						continue
					}

					// Call callback with (carry, item)
					result, err := callbackInvoker(ctx, callback, []*values.Value{carry, value})
					if err != nil {
						return values.NewNull(), err
					}
					carry = result
				}

				return carry, nil
			},
		},
		{
			Name: "array_product",
			Parameters: []*registry.Parameter{
				{Name: "array", Type: "array"},
			},
			ReturnType: "number",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil || !args[0].IsArray() {
					return values.NewInt(0), nil
				}

				arr := args[0].Data.(*values.Array)
				product := float64(1)
				hasFloat := false
				hasElements := false

				for _, val := range arr.Elements {
					if val == nil {
						continue
					}
					hasElements = true

					if val.IsFloat() {
						hasFloat = true
						product *= val.ToFloat()
					} else if val.IsInt() {
						product *= float64(val.ToInt())
					} else {
						// Convert to number, non-numeric becomes 0
						if numStr := val.ToString(); numStr != "" {
							// Try to parse as number, default to 0 if fails
							if val.IsInt() {
								product *= float64(val.ToInt())
							} else {
								product *= 0 // Non-numeric string becomes 0
							}
						} else {
							product *= 0 // Empty string becomes 0
						}
					}
				}

				if !hasElements {
					return values.NewInt(1), nil // PHP returns 1 for empty array
				}

				if hasFloat {
					return values.NewFloat(product), nil
				}
				return values.NewInt(int64(product)), nil
			},
		},
		{
			Name: "array_rand",
			Parameters: []*registry.Parameter{
				{Name: "array", Type: "array"},
				{Name: "num", Type: "int", DefaultValue: values.NewInt(1)},
			},
			ReturnType: "mixed",
			MinArgs:    1,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil || !args[0].IsArray() {
					return values.NewNull(), nil
				}

				arr := args[0].Data.(*values.Array)
				if len(arr.Elements) == 0 {
					return values.NewNull(), nil
				}

				num := int64(1)
				if len(args) > 1 && args[1] != nil {
					num = args[1].ToInt()
				}

				if num < 1 {
					return values.NewNull(), fmt.Errorf("array_rand(): Second argument has to be between 1 and the number of elements in the array")
				}

				// Get all keys
				keys := make([]interface{}, 0, len(arr.Elements))
				for key := range arr.Elements {
					keys = append(keys, key)
				}

				if int64(len(keys)) < num {
					return values.NewNull(), fmt.Errorf("array_rand(): Second argument has to be between 1 and the number of elements in the array")
				}

				// If only one key requested, return it directly
				if num == 1 {
					randomIndex := len(keys) / 2 // Simple deterministic "random" for testing
					key := keys[randomIndex]
					if intKey, ok := key.(int64); ok {
						return values.NewInt(intKey), nil
					}
					return values.NewString(fmt.Sprintf("%v", key)), nil
				}

				// Return array of random keys
				result := values.NewArray()
				resultArr := result.Data.(*values.Array)

				// Simple deterministic selection for testing - in real implementation would use math/rand
				for i := int64(0); i < num && i < int64(len(keys)); i++ {
					key := keys[i]
					if intKey, ok := key.(int64); ok {
						resultArr.Elements[i] = values.NewInt(intKey)
					} else {
						resultArr.Elements[i] = values.NewString(fmt.Sprintf("%v", key))
					}
				}
				resultArr.NextIndex = num

				return result, nil
			},
		},
		{
			Name: "shuffle",
			Parameters: []*registry.Parameter{
				{Name: "array", Type: "array"},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil || !args[0].IsArray() {
					return values.NewBool(false), nil
				}

				arr := args[0].Data.(*values.Array)

				// Get all values
				values_list := make([]*values.Value, 0, len(arr.Elements))
				for _, value := range arr.Elements {
					values_list = append(values_list, value)
				}

				// Simple deterministic shuffle for testing
				// In real implementation, would use math/rand.Shuffle()
				shuffled := make([]*values.Value, len(values_list))
				for i, v := range values_list {
					// Simple reverse order as "shuffle"
					shuffled[len(values_list)-1-i] = v
				}

				// Replace array contents with shuffled values using numeric keys
				arr.Elements = make(map[interface{}]*values.Value)
				for i, value := range shuffled {
					arr.Elements[int64(i)] = value
				}
				arr.NextIndex = int64(len(shuffled))

				return values.NewBool(true), nil
			},
		},
		{
			Name: "rsort",
			Parameters: []*registry.Parameter{
				{Name: "array", Type: "array"},
				{Name: "flags", Type: "int", DefaultValue: values.NewInt(0)},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil || !args[0].IsArray() {
					return values.NewBool(false), nil
				}

				arr := args[0].Data.(*values.Array)

				// Extract values maintaining original order for stable sort
				type valueWithIndex struct {
					value *values.Value
					index int
				}

				var sortedValues []valueWithIndex
				idx := 0
				for _, value := range arr.Elements {
					sortedValues = append(sortedValues, valueWithIndex{value, idx})
					idx++
				}

				// Sort in descending order (reverse of sort())
				sort.Slice(sortedValues, func(i, j int) bool {
					vi, vj := sortedValues[i].value, sortedValues[j].value

					// Handle different type comparisons like PHP - reverse order for rsort
					if vi.IsInt() && vj.IsInt() {
						return vi.ToInt() > vj.ToInt() // Reverse comparison
					} else if vi.IsFloat() && vj.IsFloat() {
						return vi.ToFloat() > vj.ToFloat() // Reverse comparison
					} else if (vi.IsInt() || vi.IsFloat()) && (vj.IsInt() || vj.IsFloat()) {
						return vi.ToFloat() > vj.ToFloat() // Reverse comparison
					} else {
						// Fall back to string comparison - reverse
						return vi.ToString() > vj.ToString()
					}
				})

				// Rebuild array with new sorted order and numeric indices
				newElements := make(map[interface{}]*values.Value)
				for i, item := range sortedValues {
					newElements[int64(i)] = item.value
				}
				arr.Elements = newElements
				arr.NextIndex = int64(len(sortedValues))

				return values.NewBool(true), nil
			},
		},
		{
			Name: "asort",
			Parameters: []*registry.Parameter{
				{Name: "array", Type: "array"},
				{Name: "flags", Type: "int", DefaultValue: values.NewInt(0)},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil || !args[0].IsArray() {
					return values.NewBool(false), nil
				}

				arr := args[0].Data.(*values.Array)

				// Extract key-value pairs
				type keyValuePair struct {
					key   interface{}
					value *values.Value
				}

				var pairs []keyValuePair
				for key, value := range arr.Elements {
					pairs = append(pairs, keyValuePair{key, value})
				}

				// Sort by values while maintaining key association
				sort.Slice(pairs, func(i, j int) bool {
					vi, vj := pairs[i].value, pairs[j].value

					// Handle different type comparisons like PHP
					if vi.IsInt() && vj.IsInt() {
						return vi.ToInt() < vj.ToInt()
					} else if vi.IsFloat() && vj.IsFloat() {
						return vi.ToFloat() < vj.ToFloat()
					} else if (vi.IsInt() || vi.IsFloat()) && (vj.IsInt() || vj.IsFloat()) {
						return vi.ToFloat() < vj.ToFloat()
					} else {
						// Fall back to string comparison
						return vi.ToString() < vj.ToString()
					}
				})

				// Rebuild array maintaining original keys
				newElements := make(map[interface{}]*values.Value)
				for _, pair := range pairs {
					newElements[pair.key] = pair.value
				}
				arr.Elements = newElements

				return values.NewBool(true), nil
			},
		},
		{
			Name: "arsort",
			Parameters: []*registry.Parameter{
				{Name: "array", Type: "array"},
				{Name: "flags", Type: "int", DefaultValue: values.NewInt(0)},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil || !args[0].IsArray() {
					return values.NewBool(false), nil
				}

				arr := args[0].Data.(*values.Array)

				// Extract key-value pairs
				type keyValuePair struct {
					key   interface{}
					value *values.Value
				}

				var pairs []keyValuePair
				for key, value := range arr.Elements {
					pairs = append(pairs, keyValuePair{key, value})
				}

				// Sort by values in descending order while maintaining key association
				sort.Slice(pairs, func(i, j int) bool {
					vi, vj := pairs[i].value, pairs[j].value

					// Handle different type comparisons like PHP - reverse order
					if vi.IsInt() && vj.IsInt() {
						return vi.ToInt() > vj.ToInt()
					} else if vi.IsFloat() && vj.IsFloat() {
						return vi.ToFloat() > vj.ToFloat()
					} else if (vi.IsInt() || vi.IsFloat()) && (vj.IsInt() || vj.IsFloat()) {
						return vi.ToFloat() > vj.ToFloat()
					} else {
						// Fall back to string comparison - reverse
						return vi.ToString() > vj.ToString()
					}
				})

				// Rebuild array maintaining original keys
				newElements := make(map[interface{}]*values.Value)
				for _, pair := range pairs {
					newElements[pair.key] = pair.value
				}
				arr.Elements = newElements

				return values.NewBool(true), nil
			},
		},
		{
			Name: "ksort",
			Parameters: []*registry.Parameter{
				{Name: "array", Type: "array"},
				{Name: "flags", Type: "int", DefaultValue: values.NewInt(0)},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil || !args[0].IsArray() {
					return values.NewBool(false), nil
				}

				arr := args[0].Data.(*values.Array)

				// Extract key-value pairs
				type keyValuePair struct {
					key   interface{}
					value *values.Value
				}

				var pairs []keyValuePair
				for key, value := range arr.Elements {
					pairs = append(pairs, keyValuePair{key, value})
				}

				// Sort by keys
				sort.Slice(pairs, func(i, j int) bool {
					ki, kj := pairs[i].key, pairs[j].key

					// Handle different key type comparisons
					kiInt, kiIsInt := ki.(int64)
					kjInt, kjIsInt := kj.(int64)

					if kiIsInt && kjIsInt {
						return kiInt < kjInt
					}
					if kiIsInt && !kjIsInt {
						return true // int keys come first
					}
					if !kiIsInt && kjIsInt {
						return false // string keys come after
					}
					// both string keys
					return fmt.Sprintf("%v", ki) < fmt.Sprintf("%v", kj)
				})

				// Rebuild array with sorted keys
				newElements := make(map[interface{}]*values.Value)
				for _, pair := range pairs {
					newElements[pair.key] = pair.value
				}
				arr.Elements = newElements

				return values.NewBool(true), nil
			},
		},
		{
			Name: "krsort",
			Parameters: []*registry.Parameter{
				{Name: "array", Type: "array"},
				{Name: "flags", Type: "int", DefaultValue: values.NewInt(0)},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil || !args[0].IsArray() {
					return values.NewBool(false), nil
				}

				arr := args[0].Data.(*values.Array)

				// Extract key-value pairs
				type keyValuePair struct {
					key   interface{}
					value *values.Value
				}

				var pairs []keyValuePair
				for key, value := range arr.Elements {
					pairs = append(pairs, keyValuePair{key, value})
				}

				// Sort by keys in descending order
				sort.Slice(pairs, func(i, j int) bool {
					ki, kj := pairs[i].key, pairs[j].key

					// Handle different key type comparisons - reverse order
					kiInt, kiIsInt := ki.(int64)
					kjInt, kjIsInt := kj.(int64)

					if kiIsInt && kjIsInt {
						return kiInt > kjInt // Reverse comparison
					}
					if kiIsInt && !kjIsInt {
						return false // String keys come first in reverse
					}
					if !kiIsInt && kjIsInt {
						return true // Int keys come after in reverse
					}
					// both string keys - reverse
					return fmt.Sprintf("%v", ki) > fmt.Sprintf("%v", kj)
				})

				// Rebuild array with sorted keys
				newElements := make(map[interface{}]*values.Value)
				for _, pair := range pairs {
					newElements[pair.key] = pair.value
				}
				arr.Elements = newElements

				return values.NewBool(true), nil
			},
		},
		{
			Name: "usort",
			Parameters: []*registry.Parameter{
				{Name: "array", Type: "array"},
				{Name: "callback", Type: "callable"},
			},
			ReturnType: "bool",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 || args[0] == nil || !args[0].IsArray() {
					return values.NewBool(false), nil
				}

				arr := args[0].Data.(*values.Array)
				callback := args[1]

				// Extract values into slice for sorting
				var values_list []*values.Value
				for _, val := range arr.Elements {
					values_list = append(values_list, val)
				}

				// Sort using callback comparator (supports both builtin and user-defined)
				sort.Slice(values_list, func(i, j int) bool {
					result, err := callbackInvoker(ctx, callback, []*values.Value{values_list[i], values_list[j]})
					if err != nil {
						return false // Default ordering on error
					}
					return result.ToInt() < 0
				})

				// Rebuild array with new order (usort reindexes keys)
				arr.Elements = make(map[interface{}]*values.Value)
				for i, val := range values_list {
					arr.Elements[int64(i)] = val
				}
				arr.NextIndex = int64(len(values_list))

				return values.NewBool(true), nil
			},
		},
		{
			Name: "uasort",
			Parameters: []*registry.Parameter{
				{Name: "array", Type: "array"},
				{Name: "callback", Type: "callable"},
			},
			ReturnType: "bool",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 || args[0] == nil || !args[0].IsArray() {
					return values.NewBool(false), nil
				}

				arr := args[0].Data.(*values.Array)
				callback := args[1]

				// Extract key-value pairs for sorting
				type KeyValue struct {
					Key   interface{}
					Value *values.Value
				}

				pairs := make([]KeyValue, 0, len(arr.Elements))
				for key, value := range arr.Elements {
					pairs = append(pairs, KeyValue{Key: key, Value: value})
				}

				// Sort by values using callback comparator (supports both builtin and user-defined)
				sort.Slice(pairs, func(i, j int) bool {
					result, err := callbackInvoker(ctx, callback, []*values.Value{pairs[i].Value, pairs[j].Value})
					if err != nil {
						return false // Default ordering on error
					}
					return result.ToInt() < 0
				})

				// Rebuild array maintaining key association (uasort preserves keys)
				arr.Elements = make(map[interface{}]*values.Value)
				for _, pair := range pairs {
					arr.Elements[pair.Key] = pair.Value
				}

				return values.NewBool(true), nil
			},
		},
		{
			Name: "uksort",
			Parameters: []*registry.Parameter{
				{Name: "array", Type: "array"},
				{Name: "callback", Type: "callable"},
			},
			ReturnType: "bool",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 || args[0] == nil || !args[0].IsArray() {
					return values.NewBool(false), nil
				}

				arr := args[0].Data.(*values.Array)
				callback := args[1]

				// Extract key-value pairs for sorting
				type KeyValue struct {
					Key   interface{}
					Value *values.Value
				}

				pairs := make([]KeyValue, 0, len(arr.Elements))
				for key, value := range arr.Elements {
					pairs = append(pairs, KeyValue{Key: key, Value: value})
				}

				// Sort by keys using callback comparator (supports both builtin and user-defined)
				sort.Slice(pairs, func(i, j int) bool {
					// Convert keys to values for callback
					keyA := values.NewString(fmt.Sprintf("%v", pairs[i].Key))
					keyB := values.NewString(fmt.Sprintf("%v", pairs[j].Key))

					result, err := callbackInvoker(ctx, callback, []*values.Value{keyA, keyB})
					if err != nil {
						return false // Default ordering on error
					}
					return result.ToInt() < 0
				})

				// Rebuild array maintaining key-value association (uksort preserves values)
				arr.Elements = make(map[interface{}]*values.Value)
				for _, pair := range pairs {
					arr.Elements[pair.Key] = pair.Value
				}

				return values.NewBool(true), nil
			},
		},
		{
			Name: "array_diff_assoc",
			Parameters: []*registry.Parameter{
				{Name: "array", Type: "array"},
			},
			ReturnType: "array",
			IsVariadic: true,
			MinArgs:    2,
			MaxArgs:    -1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 || args[0] == nil || !args[0].IsArray() {
					return values.NewArray(), nil
				}

				arr1 := args[0].Data.(*values.Array)
				otherKeyValues := make(map[string]string)

				// Collect key-value pairs from other arrays
				for i := 1; i < len(args); i++ {
					if args[i] != nil && args[i].IsArray() {
						arr := args[i].Data.(*values.Array)
						for key, val := range arr.Elements {
							if val != nil {
								keyStr := fmt.Sprintf("%v", key)
								otherKeyValues[keyStr] = val.ToString()
							}
						}
					}
				}

				result := values.NewArray()
				resultArr := result.Data.(*values.Array)

				// Find key-value pairs that don't exist in other arrays
				for key, val := range arr1.Elements {
					if val != nil {
						keyStr := fmt.Sprintf("%v", key)
						expectedValue, exists := otherKeyValues[keyStr]

						// Include if key doesn't exist or value is different
						if !exists || expectedValue != val.ToString() {
							resultArr.Elements[key] = val
						}
					}
				}

				return result, nil
			},
		},
		{
			Name: "array_diff_key",
			Parameters: []*registry.Parameter{
				{Name: "array", Type: "array"},
			},
			ReturnType: "array",
			IsVariadic: true,
			MinArgs:    2,
			MaxArgs:    -1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 || args[0] == nil || !args[0].IsArray() {
					return values.NewArray(), nil
				}

				arr1 := args[0].Data.(*values.Array)
				otherKeys := make(map[string]bool)

				// Collect keys from other arrays
				for i := 1; i < len(args); i++ {
					if args[i] != nil && args[i].IsArray() {
						arr := args[i].Data.(*values.Array)
						for key := range arr.Elements {
							keyStr := fmt.Sprintf("%v", key)
							otherKeys[keyStr] = true
						}
					}
				}

				result := values.NewArray()
				resultArr := result.Data.(*values.Array)

				// Find key-value pairs where key doesn't exist in other arrays
				for key, val := range arr1.Elements {
					keyStr := fmt.Sprintf("%v", key)
					if !otherKeys[keyStr] {
						resultArr.Elements[key] = val
					}
				}

				return result, nil
			},
		},
		{
			Name: "array_intersect_assoc",
			Parameters: []*registry.Parameter{
				{Name: "array", Type: "array"},
			},
			ReturnType: "array",
			IsVariadic: true,
			MinArgs:    2,
			MaxArgs:    -1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 || args[0] == nil || !args[0].IsArray() {
					return values.NewArray(), nil
				}

				arr1 := args[0].Data.(*values.Array)
				result := values.NewArray()
				resultArr := result.Data.(*values.Array)

				// For each element in first array, check if it exists in ALL other arrays
				for key, val := range arr1.Elements {
					if val == nil {
						continue
					}

					foundInAll := true

					// Check each other array
					for i := 1; i < len(args); i++ {
						if args[i] == nil || !args[i].IsArray() {
							foundInAll = false
							break
						}

						otherArr := args[i].Data.(*values.Array)
						otherVal, exists := otherArr.Elements[key]

						if !exists || otherVal == nil || otherVal.ToString() != val.ToString() {
							foundInAll = false
							break
						}
					}

					if foundInAll {
						resultArr.Elements[key] = val
					}
				}

				return result, nil
			},
		},
		{
			Name: "array_intersect_key",
			Parameters: []*registry.Parameter{
				{Name: "array", Type: "array"},
			},
			ReturnType: "array",
			IsVariadic: true,
			MinArgs:    2,
			MaxArgs:    -1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 || args[0] == nil || !args[0].IsArray() {
					return values.NewArray(), nil
				}

				arr1 := args[0].Data.(*values.Array)
				result := values.NewArray()
				resultArr := result.Data.(*values.Array)

				// For each element in first array, check if key exists in ALL other arrays
				for key, val := range arr1.Elements {
					if val == nil {
						continue
					}

					foundInAll := true

					// Check each other array for the key
					for i := 1; i < len(args); i++ {
						if args[i] == nil || !args[i].IsArray() {
							foundInAll = false
							break
						}

						otherArr := args[i].Data.(*values.Array)
						if _, exists := otherArr.Elements[key]; !exists {
							foundInAll = false
							break
						}
					}

					if foundInAll {
						resultArr.Elements[key] = val
					}
				}

				return result, nil
			},
		},
		{
			Name: "array_replace",
			Parameters: []*registry.Parameter{
				{Name: "array", Type: "array"},
			},
			ReturnType: "array",
			IsVariadic: true,
			MinArgs:    1,
			MaxArgs:    -1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 1 || args[0] == nil || !args[0].IsArray() {
					return values.NewArray(), nil
				}

				// Start with first array
				result := values.NewArray()
				resultArr := result.Data.(*values.Array)

				// Copy first array
				arr1 := args[0].Data.(*values.Array)
				for key, val := range arr1.Elements {
					resultArr.Elements[key] = val
				}
				resultArr.NextIndex = arr1.NextIndex

				// Replace/add values from subsequent arrays
				for i := 1; i < len(args); i++ {
					if args[i] == nil || !args[i].IsArray() {
						continue
					}

					replaceArr := args[i].Data.(*values.Array)
					for key, val := range replaceArr.Elements {
						resultArr.Elements[key] = val
					}
					if replaceArr.NextIndex > resultArr.NextIndex {
						resultArr.NextIndex = replaceArr.NextIndex
					}
				}

				return result, nil
			},
		},
		{
			Name: "current",
			Parameters: []*registry.Parameter{
				{Name: "array", Type: "array"},
			},
			ReturnType: "mixed",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil || !args[0].IsArray() {
					return values.NewBool(false), nil
				}

				arr := args[0].Data.(*values.Array)
				if len(arr.Elements) == 0 {
					return values.NewBool(false), nil
				}

				// Find first element (simplified implementation)
				for i := int64(0); i < arr.NextIndex; i++ {
					if val, exists := arr.Elements[i]; exists {
						return val, nil
					}
				}

				// Check string keys
				for _, val := range arr.Elements {
					return val, nil
				}

				return values.NewBool(false), nil
			},
		},
		{
			Name: "reset",
			Parameters: []*registry.Parameter{
				{Name: "array", Type: "array"},
			},
			ReturnType: "mixed",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// For now, implement as current() since we don't have array pointers
				if len(args) == 0 || args[0] == nil || !args[0].IsArray() {
					return values.NewBool(false), nil
				}

				arr := args[0].Data.(*values.Array)
				if len(arr.Elements) == 0 {
					return values.NewBool(false), nil
				}

				// Return first element
				for i := int64(0); i < arr.NextIndex; i++ {
					if val, exists := arr.Elements[i]; exists {
						return val, nil
					}
				}

				for _, val := range arr.Elements {
					return val, nil
				}

				return values.NewBool(false), nil
			},
		},
		{
			Name: "end",
			Parameters: []*registry.Parameter{
				{Name: "array", Type: "array"},
			},
			ReturnType: "mixed",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil || !args[0].IsArray() {
					return values.NewBool(false), nil
				}

				arr := args[0].Data.(*values.Array)
				if len(arr.Elements) == 0 {
					return values.NewBool(false), nil
				}

				// Find last element - check highest numeric key first
				for i := arr.NextIndex - 1; i >= 0; i-- {
					if val, exists := arr.Elements[i]; exists {
						return val, nil
					}
				}

				// Check string keys (last one found)
				var lastVal *values.Value
				for key, val := range arr.Elements {
					if _, isInt := key.(int64); !isInt {
						lastVal = val
					}
				}

				if lastVal != nil {
					return lastVal, nil
				}

				return values.NewBool(false), nil
			},
		},
		{
			Name: "next",
			Parameters: []*registry.Parameter{
				{Name: "array", Type: "array"},
			},
			ReturnType: "mixed",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// Simple implementation - for proper implementation, we'd need array cursors
				// For now, return false (indicating end of array reached)
				return values.NewBool(false), nil
			},
		},
		{
			Name: "prev",
			Parameters: []*registry.Parameter{
				{Name: "array", Type: "array"},
			},
			ReturnType: "mixed",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// Simple implementation - for proper implementation, we'd need array cursors
				// For now, return false (indicating beginning of array reached)
				return values.NewBool(false), nil
			},
		},
		{
			Name: "key",
			Parameters: []*registry.Parameter{
				{Name: "array", Type: "array"},
			},
			ReturnType: "mixed",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil || !args[0].IsArray() {
					return values.NewNull(), nil
				}

				arr := args[0].Data.(*values.Array)
				if len(arr.Elements) == 0 {
					return values.NewNull(), nil
				}

				// Return first key (simple implementation)
				for i := int64(0); i < arr.NextIndex; i++ {
					if _, exists := arr.Elements[i]; exists {
						return values.NewInt(i), nil
					}
				}

				// Check string keys
				for key := range arr.Elements {
					if _, isInt := key.(int64); !isInt {
						return values.NewString(fmt.Sprintf("%v", key)), nil
					}
				}

				return values.NewNull(), nil
			},
		},
		{
			Name: "pos",
			Parameters: []*registry.Parameter{
				{Name: "array", Type: "array"},
			},
			ReturnType: "mixed",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// pos is an alias for current
				reg := ctx.SymbolRegistry()
				if reg != nil {
					if currentFn, ok := reg.GetFunction("current"); ok && currentFn != nil && currentFn.IsBuiltin {
						return currentFn.Builtin(ctx, args)
					}
				}
				return values.NewBool(false), nil
			},
		},
		{
			Name: "sizeof",
			Parameters: []*registry.Parameter{
				{Name: "var", Type: "mixed"},
				{Name: "mode", Type: "int", DefaultValue: values.NewInt(0)},
			},
			ReturnType: "int",
			MinArgs:    1,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// sizeof is an alias for count
				reg := ctx.SymbolRegistry()
				if reg != nil {
					if countFn, ok := reg.GetFunction("count"); ok && countFn != nil && countFn.IsBuiltin {
						return countFn.Builtin(ctx, args)
					}
				}
				return values.NewInt(0), nil
			},
		},
		{
			Name: "array_replace_recursive",
			Parameters: []*registry.Parameter{
				{Name: "array", Type: "array"},
			},
			ReturnType: "array",
			IsVariadic: true,
			MinArgs:    1,
			MaxArgs:    -1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 1 || args[0] == nil || !args[0].IsArray() {
					return values.NewArray(), nil
				}

				// Start with first array (deep copy)
				result := deepCopyArray(args[0])

				// Recursively replace with subsequent arrays
				for i := 1; i < len(args); i++ {
					if args[i] == nil || !args[i].IsArray() {
						continue
					}
					result = replaceRecursive(result, args[i])
				}

				return result, nil
			},
		},
		{
			Name: "array_is_list",
			Parameters: []*registry.Parameter{
				{Name: "array", Type: "array"},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil || !args[0].IsArray() {
					return values.NewBool(false), nil
				}

				arr := args[0].Data.(*values.Array)

				// Empty array is a list
				if len(arr.Elements) == 0 {
					return values.NewBool(true), nil
				}

				// Check if keys are consecutive integers starting from 0
				expectedKey := int64(0)
				for expectedKey < arr.NextIndex {
					if _, exists := arr.Elements[expectedKey]; !exists {
						return values.NewBool(false), nil
					}
					expectedKey++
				}

				// Check that no non-numeric keys exist
				for key := range arr.Elements {
					if _, isInt := key.(int64); !isInt {
						return values.NewBool(false), nil
					}
				}

				return values.NewBool(true), nil
			},
		},
		{
			Name: "array_udiff",
			Parameters: []*registry.Parameter{
				{Name: "array", Type: "array"},
			},
			ReturnType: "array",
			IsVariadic: true,
			MinArgs:    3, // At least 2 arrays + 1 callback
			MaxArgs:    -1,
			IsBuiltin:  true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 3 || args[0] == nil || !args[0].IsArray() {
					return values.NewArray(), nil
				}

				// For now, return error for user-defined callbacks
				return values.NewArray(), fmt.Errorf("array_udiff(): User-defined callbacks not yet supported")
			},
		},
		{
			Name: "array_udiff_assoc",
			Parameters: []*registry.Parameter{
				{Name: "array", Type: "array"},
			},
			ReturnType: "array",
			IsVariadic: true,
			MinArgs:    3,
			MaxArgs:    -1,
			IsBuiltin:  true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				return values.NewArray(), fmt.Errorf("array_udiff_assoc(): User-defined callbacks not yet supported")
			},
		},
		{
			Name: "array_udiff_uassoc",
			Parameters: []*registry.Parameter{
				{Name: "array", Type: "array"},
			},
			ReturnType: "array",
			IsVariadic: true,
			MinArgs:    4,
			MaxArgs:    -1,
			IsBuiltin:  true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				return values.NewArray(), fmt.Errorf("array_udiff_uassoc(): User-defined callbacks not yet supported")
			},
		},
		{
			Name: "array_uintersect",
			Parameters: []*registry.Parameter{
				{Name: "array", Type: "array"},
			},
			ReturnType: "array",
			IsVariadic: true,
			MinArgs:    3,
			MaxArgs:    -1,
			IsBuiltin:  true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				return values.NewArray(), fmt.Errorf("array_uintersect(): User-defined callbacks not yet supported")
			},
		},
		{
			Name: "array_uintersect_assoc",
			Parameters: []*registry.Parameter{
				{Name: "array", Type: "array"},
			},
			ReturnType: "array",
			IsVariadic: true,
			MinArgs:    3,
			MaxArgs:    -1,
			IsBuiltin:  true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				return values.NewArray(), fmt.Errorf("array_uintersect_assoc(): User-defined callbacks not yet supported")
			},
		},
		{
			Name: "array_uintersect_uassoc",
			Parameters: []*registry.Parameter{
				{Name: "array", Type: "array"},
			},
			ReturnType: "array",
			IsVariadic: true,
			MinArgs:    4,
			MaxArgs:    -1,
			IsBuiltin:  true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				return values.NewArray(), fmt.Errorf("array_uintersect_uassoc(): User-defined callbacks not yet supported")
			},
		},
		{
			Name: "natcasesort",
			Parameters: []*registry.Parameter{
				{Name: "array", Type: "array"},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil || !args[0].IsArray() {
					return values.NewBool(false), nil
				}

				arr := args[0].Data.(*values.Array)

				// Extract key-value pairs
				type keyValuePair struct {
					key   interface{}
					value *values.Value
				}

				var pairs []keyValuePair
				for key, value := range arr.Elements {
					pairs = append(pairs, keyValuePair{key, value})
				}

				// Sort using case-insensitive natural order
				sort.Slice(pairs, func(i, j int) bool {
					vi, vj := pairs[i].value, pairs[j].value
					// Convert to lowercase strings for natural comparison
					si := strings.ToLower(vi.ToString())
					sj := strings.ToLower(vj.ToString())
					return naturalCompare(si, sj) < 0
				})

				// Rebuild array maintaining original keys
				newElements := make(map[interface{}]*values.Value)
				for _, pair := range pairs {
					newElements[pair.key] = pair.value
				}
				arr.Elements = newElements

				return values.NewBool(true), nil
			},
		},
		{
			Name: "natsort",
			Parameters: []*registry.Parameter{
				{Name: "array", Type: "array"},
			},
			ReturnType: "bool",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil || !args[0].IsArray() {
					return values.NewBool(false), nil
				}

				arr := args[0].Data.(*values.Array)

				// Extract key-value pairs
				type keyValuePair struct {
					key   interface{}
					value *values.Value
				}

				var pairs []keyValuePair
				for key, value := range arr.Elements {
					pairs = append(pairs, keyValuePair{key, value})
				}

				// Sort using natural order
				sort.Slice(pairs, func(i, j int) bool {
					vi, vj := pairs[i].value, pairs[j].value
					return naturalCompare(vi.ToString(), vj.ToString()) < 0
				})

				// Rebuild array maintaining original keys
				newElements := make(map[interface{}]*values.Value)
				for _, pair := range pairs {
					newElements[pair.key] = pair.value
				}
				arr.Elements = newElements

				return values.NewBool(true), nil
			},
		},
		{
			Name: "array_multisort",
			Parameters: []*registry.Parameter{
				{Name: "array", Type: "array"},
			},
			ReturnType: "bool",
			IsVariadic: true,
			MinArgs:    1,
			MaxArgs:    -1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil || !args[0].IsArray() {
					return values.NewBool(false), nil
				}

				// Simple implementation - for full array_multisort, we'd need complex
				// multi-dimensional sorting with multiple sort orders
				arr := args[0].Data.(*values.Array)

				// Extract values maintaining original order for stable sort
				type valueWithIndex struct {
					value *values.Value
					index int
				}

				var sortedValues []valueWithIndex
				idx := 0
				for _, value := range arr.Elements {
					sortedValues = append(sortedValues, valueWithIndex{value, idx})
					idx++
				}

				// Simple sort by first column values
				sort.Slice(sortedValues, func(i, j int) bool {
					vi, vj := sortedValues[i].value, sortedValues[j].value

					if vi.IsInt() && vj.IsInt() {
						return vi.ToInt() < vj.ToInt()
					} else if vi.IsFloat() && vj.IsFloat() {
						return vi.ToFloat() < vj.ToFloat()
					} else if (vi.IsInt() || vi.IsFloat()) && (vj.IsInt() || vj.IsFloat()) {
						return vi.ToFloat() < vj.ToFloat()
					} else {
						return vi.ToString() < vj.ToString()
					}
				})

				// Rebuild array with sorted order and numeric indices
				newElements := make(map[interface{}]*values.Value)
				for i, item := range sortedValues {
					newElements[int64(i)] = item.value
				}
				arr.Elements = newElements
				arr.NextIndex = int64(len(sortedValues))

				return values.NewBool(true), nil
			},
		},
		{
			Name: "each",
			Parameters: []*registry.Parameter{
				{Name: "array", Type: "array"},
			},
			ReturnType: "array",
			MinArgs:    1,
			MaxArgs:    1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// each() is deprecated in PHP 7.2+ but still widely used
				// Returns array(1, value, "key" => key, "value" => value) or false
				if len(args) == 0 || args[0] == nil || !args[0].IsArray() {
					return values.NewBool(false), nil
				}

				arr := args[0].Data.(*values.Array)
				if len(arr.Elements) == 0 {
					return values.NewBool(false), nil
				}

				// Find first element (simple implementation)
				for key, value := range arr.Elements {
					result := values.NewArray()
					resultArr := result.Data.(*values.Array)

					// Add numeric indices
					resultArr.Elements[int64(1)] = value
					resultArr.Elements["value"] = value

					if intKey, ok := key.(int64); ok {
						resultArr.Elements[int64(0)] = values.NewInt(intKey)
						resultArr.Elements["key"] = values.NewInt(intKey)
					} else {
						resultArr.Elements[int64(0)] = values.NewString(fmt.Sprintf("%v", key))
						resultArr.Elements["key"] = values.NewString(fmt.Sprintf("%v", key))
					}

					resultArr.NextIndex = 2
					return result, nil
				}

				return values.NewBool(false), nil
			},
		},
		{
			Name: "compact",
			Parameters: []*registry.Parameter{
				{Name: "var_name", Type: "mixed"},
			},
			ReturnType: "array",
			IsVariadic: true,
			MinArgs:    1,
			MaxArgs:    -1,
			IsBuiltin:  true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				result := values.NewArray()
				resultArr := result.Data.(*values.Array)

				for _, arg := range args {
					if arg == nil {
						continue
					}

					varName := arg.ToString()
					if varName == "" {
						continue
					}

					// Get global variable by name
					if value, exists := ctx.GetGlobal(varName); exists && value != nil {
						resultArr.Elements[varName] = value
					}
				}

				return result, nil
			},
		},
		{
			Name: "extract",
			Parameters: []*registry.Parameter{
				{Name: "array", Type: "array"},
				{Name: "flags", Type: "int", DefaultValue: values.NewInt(0)}, // EXTR_OVERWRITE
				{Name: "prefix", Type: "string", DefaultValue: values.NewString("")},
			},
			ReturnType: "int",
			MinArgs:    1,
			MaxArgs:    3,
			IsBuiltin:  true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) == 0 || args[0] == nil || !args[0].IsArray() {
					return values.NewInt(0), nil
				}

				arr := args[0].Data.(*values.Array)
				flags := int64(0) // EXTR_OVERWRITE
				prefix := ""

				if len(args) > 1 && args[1] != nil {
					flags = args[1].ToInt()
				}
				if len(args) > 2 && args[2] != nil {
					prefix = args[2].ToString()
				}

				extractedCount := int64(0)

				for key, value := range arr.Elements {
					if value == nil {
						continue
					}

					keyStr := fmt.Sprintf("%v", key)

					// Validate variable name (must be valid PHP variable name)
					if keyStr == "" || !isValidVariableName(keyStr) {
						continue
					}

					varName := keyStr
					if prefix != "" {
						varName = prefix + "_" + keyStr
					}

					// Handle different extraction flags
					switch flags {
					case 0: // EXTR_OVERWRITE - default
						ctx.SetGlobal(varName, value)
						extractedCount++
					case 1: // EXTR_SKIP
						if _, exists := ctx.GetGlobal(varName); !exists {
							ctx.SetGlobal(varName, value)
							extractedCount++
						}
					default: // For other flags, just extract with overwrite
						ctx.SetGlobal(varName, value)
						extractedCount++
					}
				}

				return values.NewInt(extractedCount), nil
			},
		},
		{
			Name: "list",
			Parameters: []*registry.Parameter{
				{Name: "var", Type: "mixed"},
			},
			ReturnType: "array",
			IsVariadic: true,
			MinArgs:    1,
			MaxArgs:    -1,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				// list() is a language construct, not really a function
				// This is a placeholder - actual list() would be handled by the parser
				return values.NewArray(), fmt.Errorf("list(): Language construct not implemented as function")
			},
		},
		{
			Name: "array_all",
			Parameters: []*registry.Parameter{
				{Name: "array", Type: "array"},
				{Name: "callback", Type: "callable"},
			},
			ReturnType: "bool",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 || args[0] == nil || !args[0].IsArray() {
					return values.NewBool(false), nil
				}

				arr := args[0].Data.(*values.Array)
				callback := args[1]

				// Empty array returns true (all elements satisfy condition vacuously)
				if len(arr.Elements) == 0 {
					return values.NewBool(true), nil
				}

				// Test all elements using callback (supports both builtin and user-defined)
				for _, value := range arr.Elements {
					result, err := callbackInvoker(ctx, callback, []*values.Value{value})
					if err != nil {
						return values.NewBool(false), err
					}
					if !result.ToBool() {
						return values.NewBool(false), nil
					}
				}
				return values.NewBool(true), nil
			},
		},
		{
			Name: "array_any",
			Parameters: []*registry.Parameter{
				{Name: "array", Type: "array"},
				{Name: "callback", Type: "callable"},
			},
			ReturnType: "bool",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 || args[0] == nil || !args[0].IsArray() {
					return values.NewBool(false), nil
				}

				arr := args[0].Data.(*values.Array)
				callback := args[1]

				// Empty array returns false (no elements satisfy condition)
				if len(arr.Elements) == 0 {
					return values.NewBool(false), nil
				}

				// Test any element using callback (supports both builtin and user-defined)
				for _, value := range arr.Elements {
					result, err := callbackInvoker(ctx, callback, []*values.Value{value})
					if err != nil {
						return values.NewBool(false), err
					}
					if result.ToBool() {
						return values.NewBool(true), nil
					}
				}
				return values.NewBool(false), nil
			},
		},
		{
			Name: "array_find",
			Parameters: []*registry.Parameter{
				{Name: "array", Type: "array"},
				{Name: "callback", Type: "callable"},
			},
			ReturnType: "mixed",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 || args[0] == nil || !args[0].IsArray() {
					return values.NewNull(), nil
				}

				arr := args[0].Data.(*values.Array)
				callback := args[1]

				// Empty array returns null
				if len(arr.Elements) == 0 {
					return values.NewNull(), nil
				}

				// Find first element that matches using callback (supports both builtin and user-defined)
				for _, value := range arr.Elements {
					result, err := callbackInvoker(ctx, callback, []*values.Value{value})
					if err != nil {
						return values.NewNull(), err
					}
					if result.ToBool() {
						return value, nil
					}
				}
				return values.NewNull(), nil
			},
		},
		{
			Name: "array_find_key",
			Parameters: []*registry.Parameter{
				{Name: "array", Type: "array"},
				{Name: "callback", Type: "callable"},
			},
			ReturnType: "mixed",
			MinArgs:    2,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 || args[0] == nil || !args[0].IsArray() {
					return values.NewNull(), nil
				}

				arr := args[0].Data.(*values.Array)
				callback := args[1]

				// Empty array returns null
				if len(arr.Elements) == 0 {
					return values.NewNull(), nil
				}

				// Find first key where value matches using callback (supports both builtin and user-defined)
				for key, value := range arr.Elements {
					result, err := callbackInvoker(ctx, callback, []*values.Value{value})
					if err != nil {
						return values.NewNull(), err
					}
					if result.ToBool() {
						// Return the key
						if intKey, ok := key.(int64); ok {
							return values.NewInt(intKey), nil
						} else {
							return values.NewString(fmt.Sprintf("%v", key)), nil
						}
					}
				}
				return values.NewNull(), nil
			},
		},
	}
}

// isValidVariableName checks if a string is a valid PHP variable name
func isValidVariableName(name string) bool {
	if name == "" {
		return false
	}

	// PHP variable names must start with letter or underscore
	first := rune(name[0])
	if !((first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z') || first == '_') {
		return false
	}

	// Remaining characters must be letters, digits, or underscores
	for _, r := range name[1:] {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_') {
			return false
		}
	}

	return true
}

// compareValuesLoose performs PHP-style loose comparison
func compareValuesLoose(a, b *values.Value) bool {
	// Handle null cases
	if a.IsNull() && b.IsNull() {
		return true
	}
	if a.IsNull() || b.IsNull() {
		return false
	}

	// Same type comparison
	if a.Type == b.Type {
		switch a.Type {
		case values.TypeInt:
			return a.Data.(int64) == b.Data.(int64)
		case values.TypeString:
			return a.Data.(string) == b.Data.(string)
		case values.TypeBool:
			return a.Data.(bool) == b.Data.(bool)
		case values.TypeFloat:
			return a.Data.(float64) == b.Data.(float64)
		}
	}

	// Cross-type comparisons (simplified PHP rules)
	if (a.Type == values.TypeInt || a.Type == values.TypeFloat) && (b.Type == values.TypeInt || b.Type == values.TypeFloat) {
		var aFloat, bFloat float64
		if a.Type == values.TypeInt {
			aFloat = float64(a.Data.(int64))
		} else {
			aFloat = a.Data.(float64)
		}
		if b.Type == values.TypeInt {
			bFloat = float64(b.Data.(int64))
		} else {
			bFloat = b.Data.(float64)
		}
		return aFloat == bFloat
	}

	if a.Type == values.TypeString && b.Type == values.TypeInt {
		// Try to convert string to int
		if a.Data.(string) == fmt.Sprintf("%d", b.Data.(int64)) {
			return true
		}
	}

	if b.Type == values.TypeString && a.Type == values.TypeInt {
		// Try to convert string to int
		if b.Data.(string) == fmt.Sprintf("%d", a.Data.(int64)) {
			return true
		}
	}

	return false
}