package spl

import (
	"fmt"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// GetLimitIteratorClass returns the LimitIterator class descriptor
func GetLimitIteratorClass() *registry.ClassDescriptor {
	// Constructor
	constructorImpl := &registry.Function{
		Name:      "__construct",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 {
				return values.NewNull(), fmt.Errorf("LimitIterator::__construct() expects at least 1 argument")
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("__construct called on non-object")
			}

			// Get parameters: iterator, offset=0, limit=-1
			var iterator *values.Value
			var offset *values.Value = values.NewInt(0)
			var limit *values.Value = values.NewInt(-1)

			if len(args) > 1 && !args[1].IsNull() {
				iterator = args[1]
			}
			if len(args) > 2 {
				offset = args[2]
			}
			if len(args) > 3 {
				limit = args[3]
			}

			// Note: Temporarily disabled validation due to VM parameter passing issue
			// TODO: Re-enable when VM constructor parameter passing is fixed

			obj := thisObj.Data.(*values.Object)
			if obj.Properties == nil {
				obj.Properties = make(map[string]*values.Value)
			}

			// Store the iterator, offset, and limit
			obj.Properties["__iterator"] = iterator
			obj.Properties["__offset"] = offset
			obj.Properties["__limit"] = limit
			obj.Properties["__position"] = values.NewInt(0)

			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{
			{Name: "iterator", Type: "Iterator"},
			{Name: "offset", Type: "int", HasDefault: true, DefaultValue: values.NewInt(0)},
			{Name: "count", Type: "int", HasDefault: true, DefaultValue: values.NewInt(-1)},
		},
	}

	// getInnerIterator() method
	getInnerIteratorImpl := &registry.Function{
		Name:      "getInnerIterator",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("LimitIterator::getInnerIterator() expects exactly 1 argument (this), %d given", len(args))
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("getInnerIterator called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			if iterator, ok := obj.Properties["__iterator"]; ok && iterator != nil {
				return iterator, nil
			}

			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// getPosition() method
	getPositionImpl := &registry.Function{
		Name:      "getPosition",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("LimitIterator::getPosition() expects exactly 1 argument (this), %d given", len(args))
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("getPosition called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			if position, ok := obj.Properties["__position"]; ok && position.IsInt() {
				return position, nil
			}

			return values.NewInt(0), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// current() method - delegates to inner iterator if within bounds
	currentImpl := &registry.Function{
		Name:      "current",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("LimitIterator::current() expects exactly 1 argument (this), %d given", len(args))
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("current called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			if iterator, ok := obj.Properties["__iterator"]; ok && iterator != nil && iterator.IsObject() {
				// Check if we're within bounds
				if valid, _ := isWithinBounds(obj); valid {
					// Call current() on the inner iterator
					return callIteratorMethod(ctx, iterator, "current", []*values.Value{iterator})
				}
			}

			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// key() method - delegates to inner iterator if within bounds
	keyImpl := &registry.Function{
		Name:      "key",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("LimitIterator::key() expects exactly 1 argument (this), %d given", len(args))
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("key called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			if iterator, ok := obj.Properties["__iterator"]; ok && iterator != nil && iterator.IsObject() {
				// Check if we're within bounds
				if valid, _ := isWithinBounds(obj); valid {
					// Call key() on the inner iterator
					return callIteratorMethod(ctx, iterator, "key", []*values.Value{iterator})
				}
			}

			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// valid() method - checks if within bounds and inner iterator is valid
	validImpl := &registry.Function{
		Name:      "valid",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("LimitIterator::valid() expects exactly 1 argument (this), %d given", len(args))
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("valid called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			if iterator, ok := obj.Properties["__iterator"]; ok && iterator != nil && iterator.IsObject() {
				// Check if we're within bounds
				valid, _ := isWithinBounds(obj)
				if !valid {
					return values.NewBool(false), nil
				}

				// Call valid() on the inner iterator
				result, err := callIteratorMethod(ctx, iterator, "valid", []*values.Value{iterator})
				if err != nil {
					return values.NewBool(false), nil
				}
				return result, nil
			}

			return values.NewBool(false), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// next() method - moves to next position and updates position counter
	nextImpl := &registry.Function{
		Name:      "next",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("LimitIterator::next() expects exactly 1 argument (this), %d given", len(args))
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("next called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			if iterator, ok := obj.Properties["__iterator"]; ok && iterator != nil && iterator.IsObject() {
				// Call next() on the inner iterator
				callIteratorMethod(ctx, iterator, "next", []*values.Value{iterator})

				// Increment position counter
				if position, ok := obj.Properties["__position"]; ok && position.IsInt() {
					obj.Properties["__position"] = values.NewInt(position.ToInt() + 1)
				}
			}

			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// rewind() method - rewinds inner iterator and seeks to offset
	rewindImpl := &registry.Function{
		Name:      "rewind",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("LimitIterator::rewind() expects exactly 1 argument (this), %d given", len(args))
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("rewind called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			if iterator, ok := obj.Properties["__iterator"]; ok && iterator != nil && iterator.IsObject() {
				// Rewind the inner iterator
				callIteratorMethod(ctx, iterator, "rewind", []*values.Value{iterator})

				// Skip to offset position
				if offset, ok := obj.Properties["__offset"]; ok && offset.IsInt() {
					offsetVal := offset.ToInt()
					for i := int64(0); i < offsetVal; i++ {
						// Check if iterator is still valid before calling next
						if validResult, err := callIteratorMethod(ctx, iterator, "valid", []*values.Value{iterator}); err == nil && validResult.ToBool() {
							callIteratorMethod(ctx, iterator, "next", []*values.Value{iterator})
						} else {
							break
						}
					}
				}

				// Reset position counter
				obj.Properties["__position"] = values.NewInt(0)
			}

			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// seek() method - seeks to a specific position within the limited range
	seekImpl := &registry.Function{
		Name:      "seek",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 2 {
				return values.NewNull(), fmt.Errorf("LimitIterator::seek() expects exactly 2 arguments, %d given", len(args))
			}

			thisObj := args[0]
			position := args[1]

			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("seek called on non-object")
			}

			if !position.IsInt() {
				return values.NewNull(), fmt.Errorf("seek position must be integer")
			}

			obj := thisObj.Data.(*values.Object)

			// Rewind first
			rewindResult, err := rewindImpl.Builtin(ctx, []*values.Value{thisObj})
			if err != nil {
				return rewindResult, err
			}

			// Move to the requested position
			targetPos := position.ToInt()
			if iterator, ok := obj.Properties["__iterator"]; ok && iterator != nil && iterator.IsObject() {
				for i := int64(0); i < targetPos; i++ {
					// Check if we're still within bounds and valid
					if valid, _ := isWithinBounds(obj); !valid {
						return values.NewNull(), fmt.Errorf("seek position %d is out of bounds", targetPos)
					}

					if validResult, err := callIteratorMethod(ctx, iterator, "valid", []*values.Value{iterator}); err == nil && validResult.ToBool() {
						callIteratorMethod(ctx, iterator, "next", []*values.Value{iterator})
						obj.Properties["__position"] = values.NewInt(i + 1)
					} else {
						return values.NewNull(), fmt.Errorf("seek position %d is out of bounds", targetPos)
					}
				}
			}

			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{
			{Name: "position", Type: "int"},
		},
	}

	// Create method descriptors
	methods := map[string]*registry.MethodDescriptor{
		"__construct": {
			Name:       "__construct",
			Visibility: "public",
			Parameters: []*registry.ParameterDescriptor{
				{Name: "iterator", Type: "Iterator", HasDefault: true, DefaultValue: values.NewNull()},
				{Name: "offset", Type: "int", HasDefault: true, DefaultValue: values.NewInt(0)},
				{Name: "count", Type: "int", HasDefault: true, DefaultValue: values.NewInt(-1)},
			},
			Implementation: NewBuiltinMethodImpl(constructorImpl),
		},
		"getInnerIterator": {
			Name:           "getInnerIterator",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(getInnerIteratorImpl),
		},
		"getPosition": {
			Name:           "getPosition",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(getPositionImpl),
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
		"seek": {
			Name:       "seek",
			Visibility: "public",
			Parameters: []*registry.ParameterDescriptor{
				{Name: "position", Type: "int"},
			},
			Implementation: NewBuiltinMethodImpl(seekImpl),
		},
	}

	return &registry.ClassDescriptor{
		Name:        "LimitIterator",
		Methods:     methods,
		Constants:   make(map[string]*registry.ConstantDescriptor),
		Interfaces:  []string{"Iterator", "OuterIterator", "SeekableIterator"},
	}
}

// Helper function to check if current position is within bounds
func isWithinBounds(obj *values.Object) (bool, error) {
	position, hasPos := obj.Properties["__position"]
	limit, hasLimit := obj.Properties["__limit"]

	if !hasPos || !position.IsInt() {
		return false, fmt.Errorf("position not found or invalid")
	}

	if !hasLimit || !limit.IsInt() {
		return true, nil // No limit
	}

	posVal := position.ToInt()
	limitVal := limit.ToInt()

	// If limit is -1, no limit
	if limitVal == -1 {
		return true, nil
	}

	// Check if position is within limit
	return posVal < limitVal, nil
}

