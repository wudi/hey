package spl

import (
	"fmt"
	"sort"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// GetArrayIteratorClass returns the ArrayIterator class descriptor
func GetArrayIteratorClass() *registry.ClassDescriptor {
	// Constructor
	constructorImpl := &registry.Function{
		Name:      "__construct",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 {
				return values.NewNull(), fmt.Errorf("ArrayIterator::__construct() expects at least 1 argument, %d given", len(args))
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("__construct called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			if obj.Properties == nil {
				obj.Properties = make(map[string]*values.Value)
			}

			// Handle array or object argument
			var array *values.Value
			if len(args) > 1 && args[1] != nil {
				if args[1].IsArray() {
					array = args[1]
				} else if args[1].IsObject() {
					// Convert object properties to array
					objData := args[1].Data.(*values.Object)
					newArray := values.NewArray()
					if objData.Properties != nil {
						for k, v := range objData.Properties {
							newArray.ArraySet(values.NewString(k), v)
						}
					}
					array = newArray
				} else {
					return values.NewNull(), fmt.Errorf("ArrayIterator::__construct() expects parameter 1 to be array or object")
				}
			} else {
				array = values.NewArray()
			}

			// Store internal data
			obj.Properties["__array"] = array
			obj.Properties["__position"] = values.NewInt(0)

			// Build keys array for iteration in sorted order
			keys := values.NewArray()
			if array.IsArray() {
				arr := array.Data.(*values.Array)

				// Collect all keys first
				var intKeys []int64
				var stringKeys []string

				for k := range arr.Elements {
					switch v := k.(type) {
					case int64:
						intKeys = append(intKeys, v)
					case string:
						stringKeys = append(stringKeys, v)
					}
				}

				// Sort keys to ensure consistent iteration order
				sort.Slice(intKeys, func(i, j int) bool { return intKeys[i] < intKeys[j] })
				sort.Strings(stringKeys)

				// Add sorted keys to keys array
				index := 0
				for _, key := range intKeys {
					keys.ArraySet(values.NewInt(int64(index)), values.NewInt(key))
					index++
				}
				for _, key := range stringKeys {
					keys.ArraySet(values.NewInt(int64(index)), values.NewString(key))
					index++
				}
			}
			obj.Properties["__keys"] = keys

			return thisObj, nil
		},
		Parameters: []*registry.Parameter{
			{Name: "array", Type: "array|object", HasDefault: true, DefaultValue: values.NewArray()},
			{Name: "flags", Type: "int", HasDefault: true, DefaultValue: values.NewInt(0)},
		},
	}

	// count() method
	countImpl := &registry.Function{
		Name:      "count",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 || !args[0].IsObject() {
				return values.NewInt(0), nil
			}
			obj := args[0].Data.(*values.Object)
			if obj.Properties != nil {
				if arr, ok := obj.Properties["__array"]; ok && arr != nil {
					return values.NewInt(int64(arr.ArrayCount())), nil
				}
			}
			return values.NewInt(0), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// current() method
	currentImpl := &registry.Function{
		Name:      "current",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 || !args[0].IsObject() {
				return values.NewNull(), nil
			}
			obj := args[0].Data.(*values.Object)
			if obj.Properties == nil {
				return values.NewNull(), nil
			}

			arr, ok := obj.Properties["__array"]
			if !ok || arr == nil {
				return values.NewNull(), nil
			}

			keys, ok := obj.Properties["__keys"]
			if !ok || keys == nil {
				return values.NewNull(), nil
			}

			pos, ok := obj.Properties["__position"]
			if !ok || pos == nil {
				return values.NewNull(), nil
			}

			position := int(pos.ToInt())
			if position < 0 || position >= keys.ArrayCount() {
				return values.NewNull(), nil
			}

			key := keys.ArrayGet(values.NewInt(int64(position)))
			if key == nil {
				return values.NewNull(), nil
			}

			// Get value from array using the key
			val := arr.ArrayGet(key)
			return val, nil
		},
		Parameters: []*registry.Parameter{},
	}

	// key() method
	keyImpl := &registry.Function{
		Name:      "key",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 || !args[0].IsObject() {
				return values.NewNull(), nil
			}
			obj := args[0].Data.(*values.Object)
			if obj.Properties == nil {
				return values.NewNull(), nil
			}

			keys, ok := obj.Properties["__keys"]
			if !ok || keys == nil {
				return values.NewNull(), nil
			}

			pos, ok := obj.Properties["__position"]
			if !ok || pos == nil {
				return values.NewNull(), nil
			}

			position := int(pos.ToInt())
			if position < 0 || position >= keys.ArrayCount() {
				return values.NewNull(), nil
			}

			key := keys.ArrayGet(values.NewInt(int64(position)))
			if key == nil {
				return values.NewNull(), nil
			}

			return key, nil
		},
		Parameters: []*registry.Parameter{},
	}

	// next() method
	nextImpl := &registry.Function{
		Name:      "next",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 || !args[0].IsObject() {
				return values.NewNull(), nil
			}
			obj := args[0].Data.(*values.Object)
			if obj.Properties == nil {
				return values.NewNull(), nil
			}

			pos, ok := obj.Properties["__position"]
			if !ok || pos == nil {
				obj.Properties["__position"] = values.NewInt(1)
			} else {
				obj.Properties["__position"] = values.NewInt(pos.ToInt() + 1)
			}

			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// rewind() method
	rewindImpl := &registry.Function{
		Name:      "rewind",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 || !args[0].IsObject() {
				return values.NewNull(), nil
			}
			obj := args[0].Data.(*values.Object)
			if obj.Properties == nil {
				obj.Properties = make(map[string]*values.Value)
			}
			obj.Properties["__position"] = values.NewInt(0)
			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// valid() method
	validImpl := &registry.Function{
		Name:      "valid",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 || !args[0].IsObject() {
				return values.NewBool(false), nil
			}
			obj := args[0].Data.(*values.Object)
			if obj.Properties == nil {
				return values.NewBool(false), nil
			}

			keys, ok := obj.Properties["__keys"]
			if !ok || keys == nil {
				return values.NewBool(false), nil
			}

			pos, ok := obj.Properties["__position"]
			if !ok || pos == nil {
				return values.NewBool(false), nil
			}

			position := int(pos.ToInt())
			return values.NewBool(position >= 0 && position < keys.ArrayCount()), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// offsetExists() method (ArrayAccess interface)
	offsetExistsImpl := &registry.Function{
		Name:      "offsetExists",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 2 {
				return values.NewBool(false), nil
			}
			if !args[0].IsObject() {
				return values.NewBool(false), nil
			}
			obj := args[0].Data.(*values.Object)
			if obj.Properties == nil {
				return values.NewBool(false), nil
			}

			arr, ok := obj.Properties["__array"]
			if !ok || arr == nil {
				return values.NewBool(false), nil
			}

			val := arr.ArrayGet(args[1])
			return values.NewBool(!val.IsNull()), nil
		},
		Parameters: []*registry.Parameter{
			{Name: "index", Type: "mixed"},
		},
	}

	// offsetGet() method
	offsetGetImpl := &registry.Function{
		Name:      "offsetGet",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 2 {
				return values.NewNull(), nil
			}
			if !args[0].IsObject() {
				return values.NewNull(), nil
			}
			obj := args[0].Data.(*values.Object)
			if obj.Properties == nil {
				return values.NewNull(), nil
			}

			arr, ok := obj.Properties["__array"]
			if !ok || arr == nil {
				return values.NewNull(), nil
			}

			val := arr.ArrayGet(args[1])
			return val, nil
		},
		Parameters: []*registry.Parameter{
			{Name: "index", Type: "mixed"},
		},
	}

	// offsetSet() method
	offsetSetImpl := &registry.Function{
		Name:      "offsetSet",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 3 {
				return values.NewNull(), fmt.Errorf("offsetSet expects 3 arguments")
			}
			if !args[0].IsObject() {
				return values.NewNull(), fmt.Errorf("offsetSet called on non-object")
			}
			obj := args[0].Data.(*values.Object)
			if obj.Properties == nil {
				obj.Properties = make(map[string]*values.Value)
			}

			arr, ok := obj.Properties["__array"]
			if !ok || arr == nil {
				return values.NewNull(), fmt.Errorf("internal array not initialized")
			}

			// Set the value in array
			arr.ArraySet(args[1], args[2])

			// Rebuild keys
			keys := values.NewArray()
			if arr.IsArray() {
				arrData := arr.Data.(*values.Array)
				index := 0
				for k := range arrData.Elements {
					keyVal := values.NewNull()
					switch v := k.(type) {
					case int64:
						keyVal = values.NewInt(v)
					case string:
						keyVal = values.NewString(v)
					}
					keys.ArraySet(values.NewInt(int64(index)), keyVal)
					index++
				}
			}
			obj.Properties["__keys"] = keys

			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{
			{Name: "index", Type: "mixed"},
			{Name: "value", Type: "mixed"},
		},
	}

	// offsetUnset() method
	offsetUnsetImpl := &registry.Function{
		Name:      "offsetUnset",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 2 {
				return values.NewNull(), nil
			}
			if !args[0].IsObject() {
				return values.NewNull(), nil
			}
			obj := args[0].Data.(*values.Object)
			if obj.Properties == nil {
				return values.NewNull(), nil
			}

			arr, ok := obj.Properties["__array"]
			if !ok || arr == nil {
				return values.NewNull(), nil
			}

			arr.ArrayUnset(args[1])

			// Rebuild keys
			keys := values.NewArray()
			if arr.IsArray() {
				arrData := arr.Data.(*values.Array)
				index := 0
				for k := range arrData.Elements {
					keyVal := values.NewNull()
					switch v := k.(type) {
					case int64:
						keyVal = values.NewInt(v)
					case string:
						keyVal = values.NewString(v)
					}
					keys.ArraySet(values.NewInt(int64(index)), keyVal)
					index++
				}
			}
			obj.Properties["__keys"] = keys

			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{
			{Name: "index", Type: "mixed"},
		},
	}

	// Create method descriptors
	methods := map[string]*registry.MethodDescriptor{
		"__construct": {
			Name:       "__construct",
			Visibility: "public",
			Parameters: []*registry.ParameterDescriptor{
				{Name: "array", Type: "array|object", HasDefault: true, DefaultValue: values.NewArray()},
				{Name: "flags", Type: "int", HasDefault: true, DefaultValue: values.NewInt(0)},
			},
			Implementation: NewBuiltinMethodImpl(constructorImpl),
		},
		"count": {
			Name:           "count",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(countImpl),
		},
		"current": {
			Name:           "current",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(currentImpl),
		},
		"key": {
			Name:           "key",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(keyImpl),
		},
		"next": {
			Name:           "next",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(nextImpl),
		},
		"rewind": {
			Name:           "rewind",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(rewindImpl),
		},
		"valid": {
			Name:           "valid",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(validImpl),
		},
		"offsetExists": {
			Name:       "offsetExists",
			Visibility: "public",
			Parameters: []*registry.ParameterDescriptor{
				{Name: "index", Type: "mixed"},
			},
			Implementation: NewBuiltinMethodImpl(offsetExistsImpl),
		},
		"offsetGet": {
			Name:       "offsetGet",
			Visibility: "public",
			Parameters: []*registry.ParameterDescriptor{
				{Name: "index", Type: "mixed"},
			},
			Implementation: NewBuiltinMethodImpl(offsetGetImpl),
		},
		"offsetSet": {
			Name:       "offsetSet",
			Visibility: "public",
			Parameters: []*registry.ParameterDescriptor{
				{Name: "index", Type: "mixed"},
				{Name: "value", Type: "mixed"},
			},
			Implementation: NewBuiltinMethodImpl(offsetSetImpl),
		},
		"offsetUnset": {
			Name:       "offsetUnset",
			Visibility: "public",
			Parameters: []*registry.ParameterDescriptor{
				{Name: "index", Type: "mixed"},
			},
			Implementation: NewBuiltinMethodImpl(offsetUnsetImpl),
		},
	}

	return &registry.ClassDescriptor{
		Name:       "ArrayIterator",
		Parent:     "",
		Interfaces: []string{"Iterator", "ArrayAccess", "Countable"},
		Properties: make(map[string]*registry.PropertyDescriptor),
		Methods:    methods,
		Constants:  make(map[string]*registry.ConstantDescriptor),
	}
}