package spl

import (
	"fmt"
	"sort"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// GetRecursiveArrayIteratorClass returns the RecursiveArrayIterator class descriptor
func GetRecursiveArrayIteratorClass() *registry.ClassDescriptor {
	// Constructor - similar to ArrayIterator
	constructorImpl := &registry.Function{
		Name:      "__construct",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 {
				return values.NewNull(), fmt.Errorf("RecursiveArrayIterator::__construct() expects at least 1 argument")
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("__construct called on non-object")
			}

			// Handle VM parameter passing issue - make array parameter optional
			var array *values.Value = values.NewArray() // Default empty array

			if len(args) > 1 && !args[1].IsNull() {
				array = args[1]
			}

			obj := thisObj.Data.(*values.Object)
			if obj.Properties == nil {
				obj.Properties = make(map[string]*values.Value)
			}

			// Store the array and initialize position
			obj.Properties["__array"] = array
			obj.Properties["__position"] = values.NewInt(0)

			// Build keys array for iteration in sorted order (same as ArrayIterator)
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

			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{
			{Name: "array", Type: "array", HasDefault: true, DefaultValue: values.NewArray()},
		},
	}

	// hasChildren() method - implements RecursiveIterator
	hasChildrenImpl := &registry.Function{
		Name:      "hasChildren",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("RecursiveArrayIterator::hasChildren() expects exactly 1 argument (this), %d given", len(args))
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("hasChildren called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			array, hasArray := obj.Properties["__array"]
			position, hasPosition := obj.Properties["__position"]

			if !hasArray || !hasPosition || !array.IsArray() || !position.IsInt() {
				return values.NewBool(false), nil
			}

			// Get current element using keys array (same approach as other methods)
			keys, hasKeys := obj.Properties["__keys"]
			if !hasKeys || keys == nil {
				return values.NewBool(false), nil
			}

			pos := int(position.ToInt())
			if pos < 0 || pos >= keys.ArrayCount() {
				return values.NewBool(false), nil
			}

			currentKey := keys.ArrayGet(values.NewInt(int64(pos)))
			if currentKey == nil {
				return values.NewBool(false), nil
			}

			currentValue := array.ArrayGet(currentKey)
			if currentValue == nil {
				return values.NewBool(false), nil
			}

			// Has children if current value is an array
			return values.NewBool(currentValue.IsArray()), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// getChildren() method - implements RecursiveIterator
	getChildrenImpl := &registry.Function{
		Name:      "getChildren",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("RecursiveArrayIterator::getChildren() expects exactly 1 argument (this), %d given", len(args))
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("getChildren called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			array, hasArray := obj.Properties["__array"]
			position, hasPosition := obj.Properties["__position"]

			if !hasArray || !hasPosition || !array.IsArray() || !position.IsInt() {
				return values.NewNull(), nil
			}

			// Get current element using keys array (same approach as other methods)
			keys, hasKeys := obj.Properties["__keys"]
			if !hasKeys || keys == nil {
				return values.NewNull(), nil
			}

			pos := int(position.ToInt())
			if pos < 0 || pos >= keys.ArrayCount() {
				return values.NewNull(), nil
			}

			currentKey := keys.ArrayGet(values.NewInt(int64(pos)))
			if currentKey == nil {
				return values.NewNull(), nil
			}

			currentValue := array.ArrayGet(currentKey)
			if currentValue == nil || !currentValue.IsArray() {
				return values.NewNull(), nil
			}

			// Create new RecursiveArrayIterator for the child array
			childObj := &values.Object{
				ClassName:  "RecursiveArrayIterator",
				Properties: make(map[string]*values.Value),
			}

			// Initialize child iterator with the nested array
			childObj.Properties["__array"] = currentValue
			childObj.Properties["__position"] = values.NewInt(0)

			// Build keys array for child iterator (same logic as constructor)
			childKeys := values.NewArray()
			if currentValue.IsArray() {
				arr := currentValue.Data.(*values.Array)

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
					childKeys.ArraySet(values.NewInt(int64(index)), values.NewInt(key))
					index++
				}
				for _, key := range stringKeys {
					childKeys.ArraySet(values.NewInt(int64(index)), values.NewString(key))
					index++
				}
			}
			childObj.Properties["__keys"] = childKeys

			return &values.Value{
				Type: values.TypeObject,
				Data: childObj,
			}, nil
		},
		Parameters: []*registry.Parameter{},
	}

	// current() method - inherited from ArrayIterator behavior
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

	// key() method - inherited from ArrayIterator behavior
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

	// valid() method - inherited from ArrayIterator behavior
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

	// next() method - inherited from ArrayIterator behavior
	nextImpl := &registry.Function{
		Name:      "next",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("RecursiveArrayIterator::next() expects exactly 1 argument (this), %d given", len(args))
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("next called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			if position, hasPosition := obj.Properties["__position"]; hasPosition && position.IsInt() {
				obj.Properties["__position"] = values.NewInt(position.ToInt() + 1)
			}

			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// rewind() method - inherited from ArrayIterator behavior
	rewindImpl := &registry.Function{
		Name:      "rewind",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("RecursiveArrayIterator::rewind() expects exactly 1 argument (this), %d given", len(args))
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("rewind called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			obj.Properties["__position"] = values.NewInt(0)

			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// Create method descriptors
	methods := map[string]*registry.MethodDescriptor{
		"__construct": {
			Name:       "__construct",
			Visibility: "public",
			Parameters: []*registry.ParameterDescriptor{
				{Name: "array", Type: "array", HasDefault: true, DefaultValue: values.NewArray()},
			},
			Implementation: NewBuiltinMethodImpl(constructorImpl),
		},
		"hasChildren": {
			Name:           "hasChildren",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(hasChildrenImpl),
		},
		"getChildren": {
			Name:           "getChildren",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(getChildrenImpl),
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
		"valid": {
			Name:           "valid",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(validImpl),
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
	}

	return &registry.ClassDescriptor{
		Name:       "RecursiveArrayIterator",
		Parent:     "ArrayIterator", // Extends ArrayIterator
		Methods:    methods,
		Constants:  make(map[string]*registry.ConstantDescriptor),
		Interfaces: []string{"Iterator", "RecursiveIterator"},
	}
}

// Helper function removed - all methods now use sorted __keys array for consistent iteration order