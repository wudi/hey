package spl

import (
	"fmt"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// GetSplFixedArrayClass returns the SplFixedArray class descriptor
func GetSplFixedArrayClass() *registry.ClassDescriptor {
	// Constructor
	constructorImpl := &registry.Function{
		Name:      "__construct",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 {
				return values.NewNull(), fmt.Errorf("SplFixedArray::__construct() expects at least 1 argument, %d given", len(args))
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("__construct called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			if obj.Properties == nil {
				obj.Properties = make(map[string]*values.Value)
			}

			// Get size argument
			size := int64(0)
			if len(args) > 1 && args[1] != nil {
				size = args[1].ToInt()
				if size < 0 {
					return values.NewNull(), fmt.Errorf("SplFixedArray::__construct(): size must be >= 0")
				}
			}

			// Initialize fixed array data
			obj.Properties["__size"] = values.NewInt(size)

			// Create internal storage as a Go slice of Value pointers
			data := make([]*values.Value, size)
			for i := range data {
				data[i] = values.NewNull()
			}

			// Store the slice in a special wrapper value
			dataVal := &values.Value{
				Type: values.TypeArray, // Use array type but with special behavior
				Data: data,             // Store Go slice directly
			}
			obj.Properties["__data"] = dataVal

			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{
			{Name: "size", Type: "int", HasDefault: false},
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
				if size, ok := obj.Properties["__size"]; ok && size != nil {
					return size, nil
				}
			}
			return values.NewInt(0), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// getSize() method
	getSizeImpl := &registry.Function{
		Name:      "getSize",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 || !args[0].IsObject() {
				return values.NewInt(0), nil
			}
			obj := args[0].Data.(*values.Object)
			if obj.Properties != nil {
				if size, ok := obj.Properties["__size"]; ok && size != nil {
					return size, nil
				}
			}
			return values.NewInt(0), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// setSize() method
	setSizeImpl := &registry.Function{
		Name:      "setSize",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 2 {
				return values.NewNull(), fmt.Errorf("setSize expects 2 arguments")
			}
			if !args[0].IsObject() {
				return values.NewNull(), fmt.Errorf("setSize called on non-object")
			}

			obj := args[0].Data.(*values.Object)
			newSize := args[1].ToInt()
			if newSize < 0 {
				return values.NewNull(), fmt.Errorf("SplFixedArray::setSize(): size must be >= 0")
			}

			oldSize := int64(0)
			if size, ok := obj.Properties["__size"]; ok && size != nil {
				oldSize = size.ToInt()
			}

			// Get current data
			var oldData []*values.Value
			if dataVal, ok := obj.Properties["__data"]; ok && dataVal != nil {
				if data, ok := dataVal.Data.([]*values.Value); ok {
					oldData = data
				}
			}

			// Create new data array
			newData := make([]*values.Value, newSize)
			for i := range newData {
				newData[i] = values.NewNull()
			}

			// Copy old data that fits
			copyLen := int(newSize)
			if int(oldSize) < copyLen {
				copyLen = int(oldSize)
			}
			for i := 0; i < copyLen; i++ {
				if oldData != nil && i < len(oldData) {
					newData[i] = oldData[i]
				}
			}

			// Update size and data
			obj.Properties["__size"] = values.NewInt(newSize)
			obj.Properties["__data"] = &values.Value{
				Type: values.TypeArray,
				Data: newData,
			}

			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{
			{Name: "size", Type: "int"},
		},
	}

	// offsetExists() method
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
			index := args[1].ToInt()

			// Check bounds
			if size, ok := obj.Properties["__size"]; ok && size != nil {
				arraySize := size.ToInt()
				return values.NewBool(index >= 0 && index < arraySize), nil
			}

			return values.NewBool(false), nil
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
			index := args[1].ToInt()

			// Check bounds
			size := int64(0)
			if sizeVal, ok := obj.Properties["__size"]; ok && sizeVal != nil {
				size = sizeVal.ToInt()
			}

			if index < 0 || index >= size {
				return values.NewNull(), fmt.Errorf("SplFixedArray: index %d out of bounds", index)
			}

			// Get data
			if dataVal, ok := obj.Properties["__data"]; ok && dataVal != nil {
				if data, ok := dataVal.Data.([]*values.Value); ok {
					if int(index) < len(data) {
						return data[index], nil
					}
				}
			}

			return values.NewNull(), nil
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
			index := args[1].ToInt()
			value := args[2]

			// Check bounds
			size := int64(0)
			if sizeVal, ok := obj.Properties["__size"]; ok && sizeVal != nil {
				size = sizeVal.ToInt()
			}

			if index < 0 || index >= size {
				return values.NewNull(), fmt.Errorf("SplFixedArray: index %d out of bounds", index)
			}

			// Set data
			if dataVal, ok := obj.Properties["__data"]; ok && dataVal != nil {
				if data, ok := dataVal.Data.([]*values.Value); ok {
					if int(index) < len(data) {
						data[index] = value
					}
				}
			}

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
			index := args[1].ToInt()

			// Check bounds
			size := int64(0)
			if sizeVal, ok := obj.Properties["__size"]; ok && sizeVal != nil {
				size = sizeVal.ToInt()
			}

			if index < 0 || index >= size {
				return values.NewNull(), fmt.Errorf("SplFixedArray: index %d out of bounds", index)
			}

			// Unset data (set to null)
			if dataVal, ok := obj.Properties["__data"]; ok && dataVal != nil {
				if data, ok := dataVal.Data.([]*values.Value); ok {
					if int(index) < len(data) {
						data[index] = values.NewNull()
					}
				}
			}

			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{
			{Name: "index", Type: "mixed"},
		},
	}

	// toArray() method
	toArrayImpl := &registry.Function{
		Name:      "toArray",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 || !args[0].IsObject() {
				return values.NewArray(), nil
			}

			obj := args[0].Data.(*values.Object)

			// Create regular array
			array := values.NewArray()

			// Copy data from fixed array
			if dataVal, ok := obj.Properties["__data"]; ok && dataVal != nil {
				if data, ok := dataVal.Data.([]*values.Value); ok {
					for i, val := range data {
						array.ArraySet(values.NewInt(int64(i)), val)
					}
				}
			}

			return array, nil
		},
		Parameters: []*registry.Parameter{},
	}

	// Create method descriptors
	methods := map[string]*registry.MethodDescriptor{
		"__construct": {
			Name:       "__construct",
			Visibility: "public",
			Parameters: []*registry.ParameterDescriptor{
				{Name: "size", Type: "int"},
			},
			Implementation: NewBuiltinMethodImpl(constructorImpl),
		},
		"count": {
			Name:           "count",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(countImpl),
		},
		"getSize": {
			Name:           "getSize",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(getSizeImpl),
		},
		"setSize": {
			Name:       "setSize",
			Visibility: "public",
			Parameters: []*registry.ParameterDescriptor{
				{Name: "size", Type: "int"},
			},
			Implementation: NewBuiltinMethodImpl(setSizeImpl),
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
		"toArray": {
			Name:           "toArray",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(toArrayImpl),
		},
	}

	return &registry.ClassDescriptor{
		Name:       "SplFixedArray",
		Parent:     "",
		Interfaces: []string{"ArrayAccess", "Countable"},
		Properties: make(map[string]*registry.PropertyDescriptor),
		Methods:    methods,
		Constants:  make(map[string]*registry.ConstantDescriptor),
	}
}