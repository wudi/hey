package runtime

import (
	"fmt"
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
	}
}