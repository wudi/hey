package runtime

import (
	"fmt"
	"sort"
	"strings"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

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
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
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
			Name:       "array_merge",
			Parameters: []*registry.Parameter{{Name: "array", Type: "array"}},
			ReturnType: "array",
			IsVariadic: true,
			MinArgs:    1,
			MaxArgs:    -1,
			IsBuiltin: true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				result := values.NewArray()
				targetArr := result.Data.(*values.Array)
				for _, arg := range args {
					if arg == nil || !arg.IsArray() {
						continue
					}
					src := arg.Data.(*values.Array)
					for key, val := range src.Elements {
						targetArr.Elements[key] = val
					}
					if src.NextIndex > targetArr.NextIndex {
						targetArr.NextIndex = src.NextIndex
					}
				}
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

				// Search for the needle in the array values
				for _, value := range arr.Elements {
					if value != nil && value.ToString() == needle.ToString() {
						return values.NewBool(true), nil
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
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				if len(args) < 2 || args[0] == nil || !args[0].IsArray() {
					return values.NewArray(), nil
				}
				arr := args[0].Data.(*values.Array)
				chunkSize := int(args[1].ToInt())
				if chunkSize <= 0 {
					return values.NewArray(), nil
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
					return values.NewBool(false), nil
				}
				keysArr := args[0].Data.(*values.Array)
				valsArr := args[1].Data.(*values.Array)
				if args[0].ArrayCount() != args[1].ArrayCount() {
					return values.NewBool(false), nil
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
				for i, j := 0, len(elements)-1; i < j; i, j = i+1, j-1 {
					elements[i], elements[j] = elements[j], elements[i]
				}
				result := values.NewArray()
				resultArr := result.Data.(*values.Array)
				if preserveKeys {
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
			Name:       "array_unique",
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
				result := values.NewArray()
				resultArr := result.Data.(*values.Array)
				seen := make(map[string]bool)
				for key, val := range arr.Elements {
					if val == nil {
						continue
					}
					valStr := val.ToString()
					if !seen[valStr] {
						seen[valStr] = true
						resultArr.Elements[key] = val
					}
				}
				return result, nil
			},
		},
		{
			Name: "array_filter",
			Parameters: []*registry.Parameter{
				{Name: "array", Type: "array"},
			},
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

				// Filter truthy values (simplified version without callback support)
				for key, val := range arr.Elements {
					if val != nil && val.ToBool() {
						resultArr.Elements[key] = val
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

				// Extract values and sort them
				var sortedValues []*values.Value
				for _, value := range arr.Elements {
					sortedValues = append(sortedValues, value)
				}

				// Simple bubble sort for string comparison
				for i := 0; i < len(sortedValues); i++ {
					for j := i + 1; j < len(sortedValues); j++ {
						if sortedValues[i].ToString() > sortedValues[j].ToString() {
							sortedValues[i], sortedValues[j] = sortedValues[j], sortedValues[i]
						}
					}
				}

				// Rebuild array with new sorted order and numeric indices
				newElements := make(map[interface{}]*values.Value)
				for i, value := range sortedValues {
					newElements[int64(i)] = value
				}
				arr.Elements = newElements

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

				// Handle string callback (builtin function names)
				if callback.IsString() {
					funcName := callback.ToString()
					reg := ctx.SymbolRegistry()
					if reg == nil {
						return values.NewArray(), fmt.Errorf("array_map(): Argument #1 ($callback) must be a valid callback or null, function \"%s\" not found or invalid function name", funcName)
					}

					func_, ok := reg.GetFunction(funcName)
					if !ok || func_ == nil || !func_.IsBuiltin {
						return values.NewArray(), fmt.Errorf("array_map(): Argument #1 ($callback) must be a valid callback or null, function \"%s\" not found or invalid function name", funcName)
					}

					// Use first array's keys for result (preserves associative keys)
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

						// Call the builtin function
						if len(callArgs) > 0 {
							if result_val, err := func_.Builtin(ctx, callArgs); err == nil && result_val != nil {
								resultArr.Elements[key] = result_val
							}
						}
					}
					return result, nil
				}

				// For other callback types (closures, user functions), return empty for now
				// Full callback support would require VM integration
				return values.NewArray(), fmt.Errorf("array_map(): User-defined callbacks not yet supported")
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
					if *length <= 0 {
						return values.NewArray(), nil
					}
					end = offset + *length
					if end > arrLen {
						end = arrLen
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
			Name: "array_reverse",
			Parameters: []*registry.Parameter{
				{Name: "array", Type: "array"},
				{Name: "preserve_keys", Type: "bool", DefaultValue: values.NewBool(false)},
			},
			ReturnType: "array",
			MinArgs:    1,
			MaxArgs:    2,
			IsBuiltin:  true,
			Builtin: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
				array := args[0]
				preserveKeys := false
				if len(args) >= 2 && args[1].Type == values.TypeBool {
					preserveKeys = args[1].Data.(bool)
				}

				if array.Type != values.TypeArray {
					return values.NewArray(), nil
				}

				result := values.NewArray()
				resultArr := result.Data.(*values.Array)
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

				// Check if this is a purely numeric sequential array
				allNumeric := true
				for _, k := range keys {
					if _, isInt := k.(int64); !isInt {
						allNumeric = false
						break
					}
				}

				// Process keys in reverse order
				newIndex := int64(0)
				for i := len(keys) - 1; i >= 0; i-- {
					key := keys[i]
					value := arrayData.Elements[key]

					if preserveKeys {
						// Keep original keys with reversed values
						resultArr.Elements[key] = value
						if numKey, ok := key.(int64); ok && numKey >= resultArr.NextIndex {
							resultArr.NextIndex = numKey + 1
						}
					} else {
						if allNumeric {
							// Reindex numeric arrays with new indices
							resultArr.Elements[newIndex] = value
							if newIndex >= resultArr.NextIndex {
								resultArr.NextIndex = newIndex + 1
							}
							newIndex++
						} else {
							// String keys are always preserved, but values are reversed
							resultArr.Elements[key] = value
						}
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
	}
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