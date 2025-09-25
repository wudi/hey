package spl

import (
	"fmt"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// CachingIterator flag constants (matching PHP constants)
const (
	CACHING_ITERATOR_CALL_TOSTRING = 1
	CACHING_ITERATOR_FULL_CACHE    = 2
)

// GetCachingIteratorClass returns the CachingIterator class descriptor
func GetCachingIteratorClass() *registry.ClassDescriptor {
	// Constructor
	constructorImpl := &registry.Function{
		Name:      "__construct",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 {
				return values.NewNull(), fmt.Errorf("CachingIterator::__construct() expects at least 1 argument")
			}

			thisObj := args[0]

			// Handle VM parameter passing issue - make iterator parameter optional
			var iterator *values.Value = values.NewNull()
			if len(args) > 1 && !args[1].IsNull() {
				iterator = args[1]
			}
			flags := values.NewInt(CACHING_ITERATOR_CALL_TOSTRING) // Default flag

			if len(args) > 2 && !args[2].IsNull() {
				flags = args[2]
			}

			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("__construct called on non-object")
			}

			if !iterator.IsNull() && !iterator.IsObject() {
				return values.NewNull(), fmt.Errorf("Iterator must be an object")
			}

			obj := thisObj.Data.(*values.Object)
			if obj.Properties == nil {
				obj.Properties = make(map[string]*values.Value)
			}

			// Store the iterator, flags, and initialize cache
			obj.Properties["__iterator"] = iterator
			obj.Properties["__flags"] = flags
			obj.Properties["__cache"] = values.NewArray()
			obj.Properties["__cached_string"] = values.NewString("")

			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{
			{Name: "iterator", Type: "Iterator", HasDefault: true, DefaultValue: values.NewNull()},
			{Name: "flags", Type: "int", HasDefault: true, DefaultValue: values.NewInt(CACHING_ITERATOR_CALL_TOSTRING)},
		},
	}

	// current() method
	currentImpl := &registry.Function{
		Name:      "current",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("current() expects exactly 1 argument")
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("current() called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			iterator := obj.Properties["__iterator"]

			if iterator == nil || iterator.IsNull() {
				return values.NewNull(), nil
			}

			// Delegate to inner iterator's current method
			return callIteratorMethod(ctx, iterator, "current", []*values.Value{iterator})
		},
	}

	// key() method
	keyImpl := &registry.Function{
		Name:      "key",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("key() expects exactly 1 argument")
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("key() called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			iterator := obj.Properties["__iterator"]

			if iterator == nil || iterator.IsNull() {
				return values.NewNull(), nil
			}

			// Delegate to inner iterator's key method
			return callIteratorMethod(ctx, iterator, "key", []*values.Value{iterator})
		},
	}

	// next() method
	nextImpl := &registry.Function{
		Name:      "next",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("next() expects exactly 1 argument")
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("next() called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			iterator := obj.Properties["__iterator"]
			flags := obj.Properties["__flags"]

			if iterator == nil || iterator.IsNull() {
				return values.NewNull(), nil
			}

			// Cache current value before moving to next
			if flags != nil && flags.IsInt() {
				flagValue := flags.ToInt()

				// Get current value for caching
				currentValue, err := callIteratorMethod(ctx, iterator, "current", []*values.Value{iterator})
				if err == nil {
					// Handle CALL_TOSTRING flag
					if flagValue&CACHING_ITERATOR_CALL_TOSTRING != 0 {
						if currentValue != nil {
							obj.Properties["__cached_string"] = values.NewString(currentValue.ToString())
						}
					}

					// Handle FULL_CACHE flag
					if flagValue&CACHING_ITERATOR_FULL_CACHE != 0 {
						// Get current key for full cache
						keyValue, err := callIteratorMethod(ctx, iterator, "key", []*values.Value{iterator})
						if err == nil && keyValue != nil {
							cache := obj.Properties["__cache"]
							if cache != nil && cache.IsArray() {
								cacheArray := cache.Data.(*values.Array)
								// Store in cache using key as index
								if keyValue.IsInt() {
									cacheArray.Elements[keyValue.ToInt()] = currentValue
								} else if keyValue.IsString() {
									cacheArray.Elements[keyValue.ToString()] = currentValue
								}
							}
						}
					}
				}
			}

			// Call next() on inner iterator
			return callIteratorMethod(ctx, iterator, "next", []*values.Value{iterator})
		},
	}

	// rewind() method
	rewindImpl := &registry.Function{
		Name:      "rewind",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("rewind() expects exactly 1 argument")
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("rewind() called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			iterator := obj.Properties["__iterator"]

			if iterator == nil || iterator.IsNull() {
				return values.NewNull(), nil
			}

			// Clear cache when rewinding
			obj.Properties["__cache"] = values.NewArray()
			obj.Properties["__cached_string"] = values.NewString("")

			// Delegate to inner iterator's rewind method
			return callIteratorMethod(ctx, iterator, "rewind", []*values.Value{iterator})
		},
	}

	// valid() method
	validImpl := &registry.Function{
		Name:      "valid",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("valid() expects exactly 1 argument")
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewBool(false), fmt.Errorf("valid() called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			iterator := obj.Properties["__iterator"]

			if iterator == nil || iterator.IsNull() {
				return values.NewBool(false), nil
			}

			// Delegate to inner iterator's valid method
			return callIteratorMethod(ctx, iterator, "valid", []*values.Value{iterator})
		},
	}

	// __toString() method
	toStringImpl := &registry.Function{
		Name:      "__toString",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("__toString() expects exactly 1 argument")
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewString(""), fmt.Errorf("__toString() called on non-object")
			}

			obj := thisObj.Data.(*values.Object)

			// For CachingIterator with CALL_TOSTRING flag, always return current value as string
			iterator := obj.Properties["__iterator"]
			if iterator != nil && !iterator.IsNull() {
				currentValue, err := callIteratorMethod(ctx, iterator, "current", []*values.Value{iterator})
				if err == nil && currentValue != nil {
					return values.NewString(currentValue.ToString()), nil
				}
			}

			return values.NewString(""), nil
		},
	}

	// hasNext() method
	hasNextImpl := &registry.Function{
		Name:      "hasNext",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("hasNext() expects exactly 1 argument")
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewBool(false), fmt.Errorf("hasNext() called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			iterator := obj.Properties["__iterator"]

			if iterator == nil || iterator.IsNull() {
				return values.NewBool(false), nil
			}

			// Clone the iterator state to check if next position is valid
			// This is a simplified approach - we check if calling next() and then valid() returns true

			// First check if current position is valid
			validResult, err := callIteratorMethod(ctx, iterator, "valid", []*values.Value{iterator})
			if err != nil || !validResult.ToBool() {
				return values.NewBool(false), nil
			}

			// Try to get next element by temporarily advancing
			// This is complex to do without side effects, so we use a simpler heuristic
			// In a full implementation, we'd need to clone iterator state

			// For now, we assume if we're at a valid position and can get key/current,
			// we check if the key suggests there might be more elements
			keyResult, err := callIteratorMethod(ctx, iterator, "key", []*values.Value{iterator})
			if err != nil {
				return values.NewBool(false), nil
			}

			// For ArrayIterator, we can check if we're at the last element
			// by comparing current key with array bounds
			iteratorObj := iterator.Data.(*values.Object)
			if iteratorObj.ClassName == "ArrayIterator" {
				if arr, ok := iteratorObj.Properties["__array"]; ok && arr != nil && arr.IsArray() {
					arrayData := arr.Data.(*values.Array)

					if keyResult.IsInt() {
						currentKey := keyResult.ToInt()
						maxKey := int64(len(arrayData.Elements)) - 1

						// Has next if current key is less than max key
						return values.NewBool(currentKey < maxKey), nil
					}
				}
			}

			// For other iterator types, assume no next (simplified)
			return values.NewBool(false), nil
		},
	}

	// getCache() method (for FULL_CACHE mode)
	getCacheImpl := &registry.Function{
		Name:      "getCache",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("getCache() expects exactly 1 argument")
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewArray(), fmt.Errorf("getCache() called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			cache := obj.Properties["__cache"]

			if cache != nil && cache.IsArray() {
				return cache, nil
			}

			return values.NewArray(), nil
		},
	}

	// getInnerIterator() method - implements OuterIterator
	getInnerIteratorImpl := &registry.Function{
		Name:      "getInnerIterator",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("getInnerIterator() expects exactly 1 argument")
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("getInnerIterator() called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			iterator := obj.Properties["__iterator"]

			if iterator == nil {
				return values.NewNull(), nil
			}

			return iterator, nil
		},
	}

	return &registry.ClassDescriptor{
		Name: "CachingIterator",
		Methods: map[string]*registry.MethodDescriptor{
			"__construct": {
				Name:       "__construct",
				Visibility: "public",
				Parameters: []*registry.ParameterDescriptor{
					{Name: "iterator", Type: "Iterator", HasDefault: true, DefaultValue: values.NewNull()},
					{Name: "flags", Type: "int", HasDefault: true, DefaultValue: values.NewInt(CACHING_ITERATOR_CALL_TOSTRING)},
				},
				Implementation: NewBuiltinMethodImpl(constructorImpl),
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
			"__toString": {
				Name:           "__toString",
				Visibility:     "public",
				Parameters:     []*registry.ParameterDescriptor{},
				Implementation: NewBuiltinMethodImpl(toStringImpl),
			},
			"hasNext": {
				Name:           "hasNext",
				Visibility:     "public",
				Parameters:     []*registry.ParameterDescriptor{},
				Implementation: NewBuiltinMethodImpl(hasNextImpl),
			},
			"getCache": {
				Name:           "getCache",
				Visibility:     "public",
				Parameters:     []*registry.ParameterDescriptor{},
				Implementation: NewBuiltinMethodImpl(getCacheImpl),
			},
			"getInnerIterator": {
				Name:           "getInnerIterator",
				Visibility:     "public",
				Parameters:     []*registry.ParameterDescriptor{},
				Implementation: NewBuiltinMethodImpl(getInnerIteratorImpl),
			},
		},
		Interfaces: []string{"OuterIterator", "Iterator"},
	}
}