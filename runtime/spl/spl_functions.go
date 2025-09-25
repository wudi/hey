package spl

import (
	"fmt"
	"unsafe"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// GetSplFunctions returns all SPL utility functions
func GetSplFunctions() []*registry.Function {
	functions := []*registry.Function{
		getSplObjectIdFunction(),
		getSplObjectHashFunction(),
		getSplClassesFunction(),
		getIteratorToArrayFunction(),
		getIteratorCountFunction(),
		getIteratorApplyFunction(),
	}

	// Add class reflection functions
	functions = append(functions, GetClassReflectionFunctions()...)

	return functions
}

// spl_object_id() function
func getSplObjectIdFunction() *registry.Function {
	return &registry.Function{
		Name:      "spl_object_id",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("spl_object_id() expects exactly 1 argument, %d given", len(args))
			}

			if !args[0].IsObject() {
				return values.NewNull(), fmt.Errorf("spl_object_id() expects parameter 1 to be object")
			}

			// Generate unique ID based on object's memory address
			obj := args[0].Data.(*values.Object)
			objectId := int64(uintptr(unsafe.Pointer(obj)))
			if objectId < 0 {
				objectId = -objectId
			}

			return values.NewInt(objectId), nil
		},
		Parameters: []*registry.Parameter{
			{Name: "object", Type: "object"},
		},
		MinArgs: 1,
		MaxArgs: 1,
	}
}

// spl_object_hash() function
func getSplObjectHashFunction() *registry.Function {
	return &registry.Function{
		Name:      "spl_object_hash",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("spl_object_hash() expects exactly 1 argument, %d given", len(args))
			}

			if !args[0].IsObject() {
				return values.NewNull(), fmt.Errorf("spl_object_hash() expects parameter 1 to be object")
			}

			// Generate hash based on object's memory address
			obj := args[0].Data.(*values.Object)
			objectId := uintptr(unsafe.Pointer(obj))

			// Create hash string similar to PHP format
			hash := fmt.Sprintf("%032x", objectId)

			return values.NewString(hash), nil
		},
		Parameters: []*registry.Parameter{
			{Name: "object", Type: "object"},
		},
		MinArgs: 1,
		MaxArgs: 1,
	}
}

// spl_classes() function
func getSplClassesFunction() *registry.Function {
	return &registry.Function{
		Name:      "spl_classes",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 0 {
				return values.NewNull(), fmt.Errorf("spl_classes() expects no arguments, %d given", len(args))
			}

			// Get all SPL classes
			splClasses := []string{
				"ArrayIterator",
				"ArrayObject",
				"SplDoublyLinkedList",
				"SplStack",
				"SplQueue",
				"SplFixedArray",
				"SplObjectStorage",
				"EmptyIterator",
				"SplFileInfo",
				"IteratorIterator",
				"LimitIterator",
				"AppendIterator",
				"FilterIterator",
				"CallbackFilterIterator",
				"RecursiveArrayIterator",
				"RecursiveIteratorIterator",
				"NoRewindIterator",
				"InfiniteIterator",
				"MultipleIterator",
				"CachingIterator",
				"RegexIterator",
				// Add more as they're implemented
			}

			result := values.NewArray()
			for i, className := range splClasses {
				result.ArraySet(values.NewInt(int64(i)), values.NewString(className))
			}

			return result, nil
		},
		Parameters: []*registry.Parameter{},
		MinArgs:    0,
		MaxArgs:    0,
	}
}

// iterator_to_array() function
func getIteratorToArrayFunction() *registry.Function {
	return &registry.Function{
		Name:      "iterator_to_array",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 || len(args) > 2 {
				return values.NewNull(), fmt.Errorf("iterator_to_array() expects 1 or 2 arguments, %d given", len(args))
			}

			iterator := args[0]
			if !iterator.IsObject() {
				return values.NewNull(), fmt.Errorf("iterator_to_array() expects parameter 1 to be an iterator")
			}

			// Check for use_keys parameter
			useKeys := true
			if len(args) > 1 {
				useKeys = args[1].ToBool()
			}

			result := values.NewArray()
			obj := iterator.Data.(*values.Object)

			// Check if object has iterator methods
			if obj.ClassName == "ArrayIterator" || obj.ClassName == "SplObjectStorage" {
				// Try to iterate through the object
				// This is a simplified version - in a full implementation,
				// we'd need to call the actual iterator methods

				// For ArrayIterator, get the internal array
				if obj.ClassName == "ArrayIterator" {
					if arr, ok := obj.Properties["__array"]; ok && arr != nil && arr.IsArray() {
						// Copy array data
						arrData := arr.Data.(*values.Array)
						index := int64(0)
						for k, v := range arrData.Elements {
							if useKeys {
								// Use original keys
								switch key := k.(type) {
								case int64:
									result.ArraySet(values.NewInt(key), v)
								case string:
									result.ArraySet(values.NewString(key), v)
								}
							} else {
								// Use numeric indices
								result.ArraySet(values.NewInt(index), v)
								index++
							}
						}
					}
				}
			}

			return result, nil
		},
		Parameters: []*registry.Parameter{
			{Name: "iterator", Type: "mixed"},
			{Name: "use_keys", Type: "bool", HasDefault: true, DefaultValue: values.NewBool(true)},
		},
		MinArgs: 1,
		MaxArgs: 2,
	}
}

// iterator_count() function
func getIteratorCountFunction() *registry.Function {
	return &registry.Function{
		Name:      "iterator_count",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("iterator_count() expects exactly 1 argument, %d given", len(args))
			}

			iterator := args[0]
			if !iterator.IsObject() {
				return values.NewNull(), fmt.Errorf("iterator_count() expects parameter 1 to be an iterator")
			}

			obj := iterator.Data.(*values.Object)

			// Check if object implements Countable
			if obj.ClassName == "ArrayIterator" || obj.ClassName == "ArrayObject" ||
				obj.ClassName == "SplDoublyLinkedList" || obj.ClassName == "SplStack" ||
				obj.ClassName == "SplQueue" || obj.ClassName == "SplFixedArray" ||
				obj.ClassName == "SplObjectStorage" {

				// For countable objects, get count directly
				if obj.ClassName == "ArrayIterator" {
					if arr, ok := obj.Properties["__array"]; ok && arr != nil && arr.IsArray() {
						return values.NewInt(int64(arr.ArrayCount())), nil
					}
				}
				// Add other specific cases as needed
			}

			// Fallback: manually count by iterating (simplified)
			return values.NewInt(0), nil
		},
		Parameters: []*registry.Parameter{
			{Name: "iterator", Type: "mixed"},
		},
		MinArgs: 1,
		MaxArgs: 1,
	}
}

// iterator_apply() function
func getIteratorApplyFunction() *registry.Function {
	return &registry.Function{
		Name:      "iterator_apply",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 2 || len(args) > 3 {
				return values.NewNull(), fmt.Errorf("iterator_apply() expects 2 or 3 arguments, %d given", len(args))
			}

			iterator := args[0]
			// callback := args[1]  // Will be used in full implementation
			// var funcArgs []*values.Value // Will be used in full implementation

			if !iterator.IsObject() {
				return values.NewNull(), fmt.Errorf("iterator_apply() expects parameter 1 to be an iterator")
			}

			// For now, return a basic count
			// Full implementation would iterate and call the callback function
			count := int64(0)

			obj := iterator.Data.(*values.Object)
			if obj.ClassName == "ArrayIterator" {
				if arr, ok := obj.Properties["__array"]; ok && arr != nil && arr.IsArray() {
					count = int64(arr.ArrayCount())
				}
			}

			return values.NewInt(count), nil
		},
		Parameters: []*registry.Parameter{
			{Name: "iterator", Type: "mixed"},
			{Name: "function", Type: "callable"},
			{Name: "args", Type: "array", HasDefault: true, DefaultValue: values.NewArray()},
		},
		MinArgs: 2,
		MaxArgs: 3,
	}
}