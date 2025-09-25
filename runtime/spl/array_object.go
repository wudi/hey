package spl

import (
	"fmt"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// GetArrayObjectClass returns the ArrayObject class descriptor
func GetArrayObjectClass() *registry.ClassDescriptor {
	// Constructor
	constructorImpl := &registry.Function{
		Name:      "__construct",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 {
				return values.NewNull(), fmt.Errorf("ArrayObject::__construct() expects at least 1 argument, %d given", len(args))
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
					// Clone the array
					array = values.NewArray()
					arr := args[1].Data.(*values.Array)
					for k, v := range arr.Elements {
						keyVal := values.NewNull()
						switch key := k.(type) {
						case int64:
							keyVal = values.NewInt(key)
						case string:
							keyVal = values.NewString(key)
						}
						array.ArraySet(keyVal, v)
					}
				} else if args[1].IsObject() {
					// Convert object properties to array
					objData := args[1].Data.(*values.Object)
					array = values.NewArray()
					if objData.Properties != nil {
						for k, v := range objData.Properties {
							array.ArraySet(values.NewString(k), v)
						}
					}
				} else {
					return values.NewNull(), fmt.Errorf("ArrayObject::__construct() expects parameter 1 to be array or object")
				}
			} else {
				array = values.NewArray()
			}

			// Store internal data
			obj.Properties["__array"] = array

			// Store flags (second argument)
			flags := int64(0)
			if len(args) > 2 && args[2] != nil {
				flags = args[2].ToInt()
			}
			obj.Properties["__flags"] = values.NewInt(flags)

			return thisObj, nil
		},
		Parameters: []*registry.Parameter{
			{Name: "array", Type: "array|object", HasDefault: true, DefaultValue: values.NewArray()},
			{Name: "flags", Type: "int", HasDefault: true, DefaultValue: values.NewInt(0)},
			{Name: "iterator_class", Type: "string", HasDefault: true, DefaultValue: values.NewString("ArrayIterator")},
		},
	}

	// count() method (Countable interface)
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

	// getIterator() method (IteratorAggregate interface)
	getIteratorImpl := &registry.Function{
		Name:      "getIterator",
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

			// Create an ArrayIterator for this array
			iteratorObj := &values.Object{
				ClassName:  "ArrayIterator",
				Properties: make(map[string]*values.Value),
			}
			iterator := &values.Value{
				Type: values.TypeObject,
				Data: iteratorObj,
			}

			// Initialize the iterator with our array
			arrayIteratorClass := GetArrayIteratorClass()
			constructor := arrayIteratorClass.Methods["__construct"]
			constructorImpl := constructor.Implementation.(*BuiltinMethodImpl)
			constructorArgs := []*values.Value{iterator, arr}
			_, err := constructorImpl.GetFunction().Builtin(ctx, constructorArgs)
			if err != nil {
				return values.NewNull(), err
			}

			return iterator, nil
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

			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{
			{Name: "index", Type: "mixed"},
		},
	}

	// getArrayCopy() method
	getArrayCopyImpl := &registry.Function{
		Name:      "getArrayCopy",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 || !args[0].IsObject() {
				return values.NewArray(), nil
			}
			obj := args[0].Data.(*values.Object)
			if obj.Properties == nil {
				return values.NewArray(), nil
			}

			arr, ok := obj.Properties["__array"]
			if !ok || arr == nil {
				return values.NewArray(), nil
			}

			// Create a copy of the array
			copy := values.NewArray()
			if arr.IsArray() {
				arrData := arr.Data.(*values.Array)
				for k, v := range arrData.Elements {
					keyVal := values.NewNull()
					switch key := k.(type) {
					case int64:
						keyVal = values.NewInt(key)
					case string:
						keyVal = values.NewString(key)
					}
					copy.ArraySet(keyVal, v)
				}
			}

			return copy, nil
		},
		Parameters: []*registry.Parameter{},
	}

	// exchangeArray() method
	exchangeArrayImpl := &registry.Function{
		Name:      "exchangeArray",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 2 {
				return values.NewNull(), fmt.Errorf("exchangeArray expects 2 arguments")
			}
			if !args[0].IsObject() {
				return values.NewNull(), fmt.Errorf("exchangeArray called on non-object")
			}
			obj := args[0].Data.(*values.Object)
			if obj.Properties == nil {
				obj.Properties = make(map[string]*values.Value)
			}

			// Get current array
			currentArr, ok := obj.Properties["__array"]
			if !ok || currentArr == nil {
				currentArr = values.NewArray()
			}

			// Set new array
			var newArray *values.Value
			if args[1].IsArray() {
				// Clone the array
				newArray = values.NewArray()
				arr := args[1].Data.(*values.Array)
				for k, v := range arr.Elements {
					keyVal := values.NewNull()
					switch key := k.(type) {
					case int64:
						keyVal = values.NewInt(key)
					case string:
						keyVal = values.NewString(key)
					}
					newArray.ArraySet(keyVal, v)
				}
			} else if args[1].IsObject() {
				// Convert object properties to array
				objData := args[1].Data.(*values.Object)
				newArray = values.NewArray()
				if objData.Properties != nil {
					for k, v := range objData.Properties {
						newArray.ArraySet(values.NewString(k), v)
					}
				}
			} else {
				return values.NewNull(), fmt.Errorf("exchangeArray expects parameter 1 to be array or object")
			}

			obj.Properties["__array"] = newArray

			return currentArr, nil
		},
		Parameters: []*registry.Parameter{
			{Name: "array", Type: "array|object"},
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
				{Name: "iterator_class", Type: "string", HasDefault: true, DefaultValue: values.NewString("ArrayIterator")},
			},
			Implementation: NewBuiltinMethodImpl(constructorImpl),
		},
		"count": {
			Name:           "count",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(countImpl),
		},
		"getIterator": {
			Name:           "getIterator",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(getIteratorImpl),
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
		"getArrayCopy": {
			Name:           "getArrayCopy",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(getArrayCopyImpl),
		},
		"exchangeArray": {
			Name:       "exchangeArray",
			Visibility: "public",
			Parameters: []*registry.ParameterDescriptor{
				{Name: "array", Type: "array|object"},
			},
			Implementation: NewBuiltinMethodImpl(exchangeArrayImpl),
		},
	}

	return &registry.ClassDescriptor{
		Name:       "ArrayObject",
		Parent:     "",
		Interfaces: []string{"IteratorAggregate", "ArrayAccess", "Countable"},
		Properties: make(map[string]*registry.PropertyDescriptor),
		Methods:    methods,
		Constants:  make(map[string]*registry.ConstantDescriptor),
	}
}