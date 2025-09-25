package spl

import (
	"fmt"
	"os"
	"sort"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// RecursiveIteratorIterator mode constants
const (
	RECURSIVE_ITERATOR_ITERATOR_LEAVES_ONLY = 0
	RECURSIVE_ITERATOR_ITERATOR_SELF_FIRST  = 1
	RECURSIVE_ITERATOR_ITERATOR_CHILD_FIRST = 2
)

// GetRecursiveIteratorIteratorClass returns the RecursiveIteratorIterator class descriptor
func GetRecursiveIteratorIteratorClass() *registry.ClassDescriptor {
	// Constructor
	constructorImpl := &registry.Function{
		Name:      "__construct",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 {
				return values.NewNull(), fmt.Errorf("RecursiveIteratorIterator::__construct() expects at least 1 argument")
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("__construct called on non-object")
			}

			// Handle VM parameter passing issue - make parameters optional with defaults
			var iterator *values.Value = values.NewNull()
			if len(args) > 1 && !args[1].IsNull() {
				iterator = args[1]
			}

			// If this is the auto-constructor call (no parameters), just initialize the object
			// and return. The real constructor call with parameters will come later.
			if iterator.IsNull() && len(args) == 1 {
				// Just initialize basic properties for auto-constructor call
				obj := thisObj.Data.(*values.Object)
				if obj.Properties == nil {
					obj.Properties = make(map[string]*values.Value)
				}
				return values.NewNull(), nil
			}

			if iterator.IsNull() {
				return values.NewNull(), fmt.Errorf("RecursiveIteratorIterator::__construct() expects parameter 1 to be RecursiveIterator, null given")
			}

			if !iterator.IsObject() {
				return values.NewNull(), fmt.Errorf("RecursiveIteratorIterator::__construct() expects parameter 1 to be RecursiveIterator")
			}

			// Default mode is LEAVES_ONLY
			mode := int64(RECURSIVE_ITERATOR_ITERATOR_LEAVES_ONLY)
			if len(args) > 2 && args[2].IsInt() {
				mode = args[2].ToInt()
			}

			// Default flags is 0
			flags := int64(0)
			if len(args) > 3 && args[3].IsInt() {
				flags = args[3].ToInt()
			}

			obj := thisObj.Data.(*values.Object)
			if obj.Properties == nil {
				obj.Properties = make(map[string]*values.Value)
			}

			// Store parameters
			obj.Properties["__iterator"] = iterator
			obj.Properties["__mode"] = values.NewInt(mode)
			obj.Properties["__flags"] = values.NewInt(flags)
			obj.Properties["__maxDepth"] = values.NewInt(-1) // No limit by default

			// Initialize iterator stack and depth tracking
			iteratorStack := values.NewArray()
			iteratorStack.ArraySet(values.NewInt(0), iterator)
			obj.Properties["__iteratorStack"] = iteratorStack

			obj.Properties["__depth"] = values.NewInt(0)
			obj.Properties["__position"] = values.NewInt(0)
			obj.Properties["__valid"] = values.NewBool(false)

			// Initialize by rewinding
			rewindRecursiveIterator(obj)

			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{
			{Name: "iterator", Type: "RecursiveIterator", HasDefault: true, DefaultValue: values.NewNull()},
			{Name: "mode", Type: "int", HasDefault: true, DefaultValue: values.NewInt(RECURSIVE_ITERATOR_ITERATOR_LEAVES_ONLY)},
			{Name: "flags", Type: "int", HasDefault: true, DefaultValue: values.NewInt(0)},
		},
	}

	// rewind() method
	rewindImpl := &registry.Function{
		Name:      "rewind",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("RecursiveIteratorIterator::rewind() expects exactly 1 argument")
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("rewind called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			rewindRecursiveIterator(obj)

			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// valid() method
	validImpl := &registry.Function{
		Name:      "valid",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("RecursiveIteratorIterator::valid() expects exactly 1 argument")
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewBool(false), nil
			}

			obj := thisObj.Data.(*values.Object)
			validValue, exists := obj.Properties["__valid"]
			if !exists {
				return values.NewBool(false), nil
			}

			return validValue, nil
		},
		Parameters: []*registry.Parameter{},
	}

	// current() method
	currentImpl := &registry.Function{
		Name:      "current",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("RecursiveIteratorIterator::current() expects exactly 1 argument")
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), nil
			}

			obj := thisObj.Data.(*values.Object)

			// Get current iterator from stack
			currentIter := getCurrentIteratorFromStack(obj)
			if currentIter == nil {
				return values.NewNull(), nil
			}

			// For now, return a simple implementation for basic functionality
			// TODO: Implement full recursive traversal logic
			if currentIter.IsObject() {
				iterObj := currentIter.Data.(*values.Object)

				// Try to call current() method directly if it's an ArrayIterator
				if iterObj.ClassName == "RecursiveArrayIterator" || iterObj.ClassName == "ArrayIterator" {
					if arr, exists := iterObj.Properties["__array"]; exists && arr.IsArray() {
						if pos, exists := iterObj.Properties["__position"]; exists {
							position := int(pos.ToInt())
							if keys, exists := iterObj.Properties["__keys"]; exists {
								if position >= 0 && position < keys.ArrayCount() {
									key := keys.ArrayGet(values.NewInt(int64(position)))
									if key != nil {
										currentValue := arr.ArrayGet(key)
										if currentValue != nil {
											return currentValue, nil
										}
									}
								}
							}
						}
					}
				} else if iterObj.ClassName == "RecursiveDirectoryIterator" {
					if iteratorData, exists := iterObj.Properties["_iteratorData"]; exists && iteratorData.Type == values.TypeResource {
						data := iteratorData.Data.(*RecursiveDirectoryIteratorData)

						if data.currentIdx >= 0 && data.currentIdx < len(data.entries) {
							currentEntry := data.entries[data.currentIdx]
							currentPath := data.path + "/" + currentEntry.Name()

							// Check flag to determine return type
							flags := data.flags

							// CURRENT_AS_PATHNAME flag (from FilesystemIterator)
							if flags&32 != 0 { // CURRENT_AS_PATHNAME = 32
								return values.NewString(currentPath), nil
							}

							// Default: return SplFileInfo object
							fileInfoObj := &values.Object{
								ClassName:  "SplFileInfo",
								Properties: make(map[string]*values.Value),
							}
							fileInfoObj.Properties["__filepath"] = values.NewString(currentPath)

							return &values.Value{
								Type: values.TypeObject,
								Data: fileInfoObj,
							}, nil
						}
					}
				}
			}

			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// key() method
	keyImpl := &registry.Function{
		Name:      "key",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("RecursiveIteratorIterator::key() expects exactly 1 argument")
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), nil
			}

			obj := thisObj.Data.(*values.Object)

			// Get current iterator from stack
			currentIter := getCurrentIteratorFromStack(obj)
			if currentIter == nil {
				return values.NewNull(), nil
			}

			// For now, return a simple implementation for basic functionality
			if currentIter.IsObject() {
				iterObj := currentIter.Data.(*values.Object)

				// Try to get key directly if it's an ArrayIterator
				if iterObj.ClassName == "RecursiveArrayIterator" || iterObj.ClassName == "ArrayIterator" {
					if pos, exists := iterObj.Properties["__position"]; exists {
						position := int(pos.ToInt())
						if keys, exists := iterObj.Properties["__keys"]; exists {
							if position >= 0 && position < keys.ArrayCount() {
								key := keys.ArrayGet(values.NewInt(int64(position)))
								if key != nil {
									return key, nil
								}
							}
						}
					}
				} else if iterObj.ClassName == "RecursiveDirectoryIterator" {
					if iteratorData, exists := iterObj.Properties["_iteratorData"]; exists && iteratorData.Type == values.TypeResource {
						data := iteratorData.Data.(*RecursiveDirectoryIteratorData)

						if data.currentIdx >= 0 && data.currentIdx < len(data.entries) {
							currentEntry := data.entries[data.currentIdx]
							flags := data.flags

							// KEY_AS_FILENAME flag (from FilesystemIterator)
							if flags&16 != 0 { // KEY_AS_FILENAME = 16
								return values.NewString(currentEntry.Name()), nil
							}

							// Default: return full path
							currentPath := data.path + "/" + currentEntry.Name()
							return values.NewString(currentPath), nil
						}
					}
				}
			}

			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// next() method
	nextImpl := &registry.Function{
		Name:      "next",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("RecursiveIteratorIterator::next() expects exactly 1 argument")
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("next called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			advanceRecursiveIterator(ctx, obj)

			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// getDepth() method
	getDepthImpl := &registry.Function{
		Name:      "getDepth",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("RecursiveIteratorIterator::getDepth() expects exactly 1 argument")
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewInt(0), nil
			}

			obj := thisObj.Data.(*values.Object)
			depth, exists := obj.Properties["__depth"]
			if !exists {
				return values.NewInt(0), nil
			}

			return depth, nil
		},
		Parameters: []*registry.Parameter{},
	}

	// setMaxDepth() method
	setMaxDepthImpl := &registry.Function{
		Name:      "setMaxDepth",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 2 {
				return values.NewNull(), fmt.Errorf("RecursiveIteratorIterator::setMaxDepth() expects exactly 2 arguments")
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("setMaxDepth called on non-object")
			}

			maxDepth := args[1]
			if !maxDepth.IsInt() {
				maxDepth = values.NewInt(int64(maxDepth.ToInt()))
			}

			obj := thisObj.Data.(*values.Object)
			obj.Properties["__maxDepth"] = maxDepth

			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{
			{Name: "maxDepth", Type: "int", HasDefault: false},
		},
	}

	// getMaxDepth() method
	getMaxDepthImpl := &registry.Function{
		Name:      "getMaxDepth",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) != 1 {
				return values.NewNull(), fmt.Errorf("RecursiveIteratorIterator::getMaxDepth() expects exactly 1 argument")
			}

			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewInt(-1), nil
			}

			obj := thisObj.Data.(*values.Object)
			maxDepth, exists := obj.Properties["__maxDepth"]
			if !exists {
				return values.NewInt(-1), nil
			}

			return maxDepth, nil
		},
		Parameters: []*registry.Parameter{},
	}

	// Create method descriptors
	methods := map[string]*registry.MethodDescriptor{
		"__construct": {
			Name:       "__construct",
			Visibility: "public",
			Parameters: []*registry.ParameterDescriptor{
				{Name: "iterator", Type: "RecursiveIterator", HasDefault: true, DefaultValue: values.NewNull()},
				{Name: "mode", Type: "int", HasDefault: true, DefaultValue: values.NewInt(RECURSIVE_ITERATOR_ITERATOR_LEAVES_ONLY)},
				{Name: "flags", Type: "int", HasDefault: true, DefaultValue: values.NewInt(0)},
			},
			Implementation: NewBuiltinMethodImpl(constructorImpl),
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
		"getDepth": {
			Name:           "getDepth",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(getDepthImpl),
		},
		"setMaxDepth": {
			Name:       "setMaxDepth",
			Visibility: "public",
			Parameters: []*registry.ParameterDescriptor{
				{Name: "maxDepth", Type: "int", HasDefault: false},
			},
			Implementation: NewBuiltinMethodImpl(setMaxDepthImpl),
		},
		"getMaxDepth": {
			Name:           "getMaxDepth",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: NewBuiltinMethodImpl(getMaxDepthImpl),
		},
	}

	// Create constants
	constants := map[string]*registry.ConstantDescriptor{
		"LEAVES_ONLY": {
			Name:  "LEAVES_ONLY",
			Value: values.NewInt(RECURSIVE_ITERATOR_ITERATOR_LEAVES_ONLY),
		},
		"SELF_FIRST": {
			Name:  "SELF_FIRST",
			Value: values.NewInt(RECURSIVE_ITERATOR_ITERATOR_SELF_FIRST),
		},
		"CHILD_FIRST": {
			Name:  "CHILD_FIRST",
			Value: values.NewInt(RECURSIVE_ITERATOR_ITERATOR_CHILD_FIRST),
		},
	}

	return &registry.ClassDescriptor{
		Name:       "RecursiveIteratorIterator",
		Parent:     "",
		Methods:    methods,
		Constants:  constants,
		Interfaces: []string{"Iterator", "OuterIterator"},
	}
}

// Helper functions for RecursiveIteratorIterator implementation

// rewindRecursiveIterator initializes the iterator to the beginning
func rewindRecursiveIterator(obj *values.Object) {
	// Reset iterator stack to contain only the root iterator
	rootIterator, exists := obj.Properties["__iterator"]
	if !exists {
		return
	}

	iteratorStack := values.NewArray()
	iteratorStack.ArraySet(values.NewInt(0), rootIterator)
	obj.Properties["__iteratorStack"] = iteratorStack
	obj.Properties["__depth"] = values.NewInt(0)

	// Rewind the root iterator
	callDirectIteratorRewind(rootIterator)

	// Find the first valid position based on mode
	findNextValidPosition(obj)
}

// getCurrentIteratorFromStack gets the current iterator from the stack
func getCurrentIteratorFromStack(obj *values.Object) *values.Value {
	stack, exists := obj.Properties["__iteratorStack"]
	if !exists || !stack.IsArray() {
		return nil
	}

	depth, exists := obj.Properties["__depth"]
	if !exists || !depth.IsInt() {
		return nil
	}

	currentDepth := depth.ToInt()
	return stack.ArrayGet(values.NewInt(currentDepth))
}

// advanceRecursiveIterator moves to the next valid position
func advanceRecursiveIterator(ctx registry.BuiltinCallContext, obj *values.Object) {
	currentIter := getCurrentIteratorFromStack(obj)
	if currentIter == nil {
		obj.Properties["__valid"] = values.NewBool(false)
		return
	}

	mode, exists := obj.Properties["__mode"]
	if !exists {
		mode = values.NewInt(RECURSIVE_ITERATOR_ITERATOR_LEAVES_ONLY)
	}

	switch mode.ToInt() {
	case RECURSIVE_ITERATOR_ITERATOR_LEAVES_ONLY:
		advanceForLeavesOnly(ctx, obj)
	case RECURSIVE_ITERATOR_ITERATOR_SELF_FIRST:
		advanceForSelfFirst(ctx, obj)
	case RECURSIVE_ITERATOR_ITERATOR_CHILD_FIRST:
		advanceForChildFirst(ctx, obj)
	}
}

// Helper functions for direct iterator method calls without registry context

// callDirectIteratorNext calls next() on an iterator object directly
func callDirectIteratorNext(iter *values.Value) bool {
	if !iter.IsObject() {
		return false
	}

	iterObj := iter.Data.(*values.Object)
	if iterObj.ClassName == "RecursiveArrayIterator" || iterObj.ClassName == "ArrayIterator" {
		if pos, exists := iterObj.Properties["__position"]; exists {
			newPos := pos.ToInt() + 1
			iterObj.Properties["__position"] = values.NewInt(newPos)
			return true
		}
	} else if iterObj.ClassName == "RecursiveDirectoryIterator" {
		if iteratorData, exists := iterObj.Properties["_iteratorData"]; exists && iteratorData.Type == values.TypeResource {
			data := iteratorData.Data.(*RecursiveDirectoryIteratorData)
			data.currentIdx++

			// If SKIP_DOTS is set, skip dot entries automatically
			if data.flags&4096 != 0 { // SKIP_DOTS = 4096
				for data.currentIdx < len(data.entries) {
					currentEntry := data.entries[data.currentIdx]
					if currentEntry.Name() != "." && currentEntry.Name() != ".." {
						break
					}
					data.currentIdx++
				}
			}
			return true
		}
	}
	return false
}

// callDirectIteratorValid calls valid() on an iterator object directly
func callDirectIteratorValid(iter *values.Value) bool {
	if !iter.IsObject() {
		return false
	}

	iterObj := iter.Data.(*values.Object)
	if iterObj.ClassName == "RecursiveArrayIterator" || iterObj.ClassName == "ArrayIterator" {
		if pos, exists := iterObj.Properties["__position"]; exists {
			if keys, exists := iterObj.Properties["__keys"]; exists {
				position := int(pos.ToInt())
				return position >= 0 && position < keys.ArrayCount()
			}
		}
	} else if iterObj.ClassName == "RecursiveDirectoryIterator" {
		if iteratorData, exists := iterObj.Properties["_iteratorData"]; exists && iteratorData.Type == values.TypeResource {
			data := iteratorData.Data.(*RecursiveDirectoryIteratorData)

			// Skip invalid positions and handle SKIP_DOTS flag
			for data.currentIdx >= 0 && data.currentIdx < len(data.entries) {
				currentEntry := data.entries[data.currentIdx]

				// Check SKIP_DOTS flag
				if data.flags&4096 != 0 { // SKIP_DOTS = 4096
					if currentEntry.Name() == "." || currentEntry.Name() == ".." {
						data.currentIdx++
						continue
					}
				}

				return true
			}
		}
	}
	return false
}

// callDirectIteratorHasChildren calls hasChildren() on an iterator object directly
func callDirectIteratorHasChildren(iter *values.Value) bool {
	if !iter.IsObject() {
		return false
	}

	iterObj := iter.Data.(*values.Object)
	if iterObj.ClassName == "RecursiveArrayIterator" {
		if arr, exists := iterObj.Properties["__array"]; exists && arr.IsArray() {
			if pos, exists := iterObj.Properties["__position"]; exists {
				position := int(pos.ToInt())
				if keys, exists := iterObj.Properties["__keys"]; exists {
					if position >= 0 && position < keys.ArrayCount() {
						key := keys.ArrayGet(values.NewInt(int64(position)))
						if key != nil {
							currentValue := arr.ArrayGet(key)
							if currentValue != nil && currentValue.IsArray() {
								return currentValue.ArrayCount() > 0
							}
						}
					}
				}
			}
		}
	} else if iterObj.ClassName == "RecursiveDirectoryIterator" {
		if iteratorData, exists := iterObj.Properties["_iteratorData"]; exists && iteratorData.Type == values.TypeResource {
			data := iteratorData.Data.(*RecursiveDirectoryIteratorData)

			if data.currentIdx < 0 || data.currentIdx >= len(data.entries) {
				return false
			}

			currentEntry := data.entries[data.currentIdx]

			// Only directories can have children, and skip '.' and '..' unless specifically requested
			if !currentEntry.IsDir() {
				return false
			}

			// Special handling for '.' and '..' - they never have children
			if currentEntry.Name() == "." || currentEntry.Name() == ".." {
				return false
			}

			return true
		}
	}
	return false
}

// callDirectIteratorGetChildren calls getChildren() on an iterator object directly
func callDirectIteratorGetChildren(iter *values.Value) *values.Value {
	if !iter.IsObject() {
		return nil
	}

	iterObj := iter.Data.(*values.Object)
	if iterObj.ClassName == "RecursiveArrayIterator" {
		if arr, exists := iterObj.Properties["__array"]; exists && arr.IsArray() {
			if pos, exists := iterObj.Properties["__position"]; exists {
				position := int(pos.ToInt())
				if keys, exists := iterObj.Properties["__keys"]; exists {
					if position >= 0 && position < keys.ArrayCount() {
						key := keys.ArrayGet(values.NewInt(int64(position)))
						if key != nil {
							currentValue := arr.ArrayGet(key)
							if currentValue != nil && currentValue.IsArray() {
								// Create new RecursiveArrayIterator for child array
								childObj := &values.Object{
									ClassName:  "RecursiveArrayIterator",
									Properties: make(map[string]*values.Value),
								}
								childObj.Properties["__array"] = currentValue
								childObj.Properties["__position"] = values.NewInt(0)

								// Build sorted keys for child iterator (same logic as RecursiveArrayIterator constructor)
								keys := values.NewArray()
								childArr := currentValue.Data.(*values.Array)

								// Use same key sorting logic as ArrayIterator
								var intKeys []int64
								var stringKeys []string
								for k := range childArr.Elements {
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

								index := 0
								for _, key := range intKeys {
									keys.ArraySet(values.NewInt(int64(index)), values.NewInt(key))
									index++
								}
								for _, key := range stringKeys {
									keys.ArraySet(values.NewInt(int64(index)), values.NewString(key))
									index++
								}

								childObj.Properties["__keys"] = keys
								return &values.Value{Type: values.TypeObject, Data: childObj}
							}
						}
					}
				}
			}
		}
	} else if iterObj.ClassName == "RecursiveDirectoryIterator" {
		if iteratorData, exists := iterObj.Properties["_iteratorData"]; exists && iteratorData.Type == values.TypeResource {
			data := iteratorData.Data.(*RecursiveDirectoryIteratorData)

			if data.currentIdx < 0 || data.currentIdx >= len(data.entries) {
				return nil
			}

			currentEntry := data.entries[data.currentIdx]

			if !currentEntry.IsDir() {
				return nil
			}

			// Build path to child directory
			childPath := data.path + "/" + currentEntry.Name()

			// Create new RecursiveDirectoryIterator for the child directory
			childObj := &values.Object{
				ClassName:  "RecursiveDirectoryIterator",
				Properties: make(map[string]*values.Value),
			}
			childThis := &values.Value{
				Type: values.TypeObject,
				Data: childObj,
			}

			// Initialize child iterator using the same constructor logic
			if err := initRecursiveDirectoryIterator(childThis, childPath, data.flags); err != nil {
				return nil
			}

			return childThis
		}
	}
	return nil
}

// callDirectIteratorRewind calls rewind() on an iterator object directly
func callDirectIteratorRewind(iter *values.Value) bool {
	if !iter.IsObject() {
		return false
	}

	iterObj := iter.Data.(*values.Object)
	if iterObj.ClassName == "RecursiveArrayIterator" || iterObj.ClassName == "ArrayIterator" {
		iterObj.Properties["__position"] = values.NewInt(0)
		return true
	} else if iterObj.ClassName == "RecursiveDirectoryIterator" {
		if iteratorData, exists := iterObj.Properties["_iteratorData"]; exists && iteratorData.Type == values.TypeResource {
			data := iteratorData.Data.(*RecursiveDirectoryIteratorData)
			data.currentIdx = 0

			// If SKIP_DOTS is set, skip initial dot entries
			if data.flags&4096 != 0 { // SKIP_DOTS = 4096
				for data.currentIdx < len(data.entries) {
					currentEntry := data.entries[data.currentIdx]
					if currentEntry.Name() != "." && currentEntry.Name() != ".." {
						break
					}
					data.currentIdx++
				}
			}
			return true
		}
	}
	return false
}

// advanceForLeavesOnly advances for LEAVES_ONLY mode
func advanceForLeavesOnly(ctx registry.BuiltinCallContext, obj *values.Object) {
	for {
		currentIter := getCurrentIteratorFromStack(obj)
		if currentIter == nil {
			obj.Properties["__valid"] = values.NewBool(false)
			return
		}

		// Move current iterator to next position
		if !callDirectIteratorNext(currentIter) {
			obj.Properties["__valid"] = values.NewBool(false)
			return
		}

		// Check if current position is valid
		if !callDirectIteratorValid(currentIter) {
			// Current iterator exhausted, pop from stack
			if !popIteratorFromStack(obj) {
				obj.Properties["__valid"] = values.NewBool(false)
				return
			}
			continue
		}

		// Check if current element has children and we can descend
		if callDirectIteratorHasChildren(currentIter) {
			// Check max depth limit
			if !exceedsMaxDepth(obj) {
				// Descend into children
				childrenResult := callDirectIteratorGetChildren(currentIter)
				if childrenResult != nil && childrenResult.IsObject() {
					pushIteratorToStack(obj, childrenResult)
					// Rewind child iterator
					callDirectIteratorRewind(childrenResult)

					// After descending into children, check if the child is valid
					// Don't call next() on the child - it should already be at first position
					if callDirectIteratorValid(childrenResult) {
						// Check if this child also has children (recursive descent)
						if callDirectIteratorHasChildren(childrenResult) && !exceedsMaxDepth(obj) {
							continue // Will descend further
						}
						// Child is a leaf - return it
						obj.Properties["__valid"] = values.NewBool(true)
						return
					} else {
						// Child iterator is empty, pop and continue
						popIteratorFromStack(obj)
						continue
					}
				}
			}
		}

		// This is a leaf node - return it
		obj.Properties["__valid"] = values.NewBool(true)
		return
	}
}

// advanceForSelfFirst advances for SELF_FIRST mode
func advanceForSelfFirst(ctx registry.BuiltinCallContext, obj *values.Object) {
	currentIter := getCurrentIteratorFromStack(obj)
	if currentIter == nil {
		obj.Properties["__valid"] = values.NewBool(false)
		return
	}

	// Move to next position
	if !callDirectIteratorNext(currentIter) {
		obj.Properties["__valid"] = values.NewBool(false)
		return
	}

	findNextValidPosition(obj)
}

// advanceForChildFirst advances for CHILD_FIRST mode
func advanceForChildFirst(ctx registry.BuiltinCallContext, obj *values.Object) {
	currentIter := getCurrentIteratorFromStack(obj)
	if currentIter == nil {
		obj.Properties["__valid"] = values.NewBool(false)
		return
	}

	// Move to next position
	if !callDirectIteratorNext(currentIter) {
		obj.Properties["__valid"] = values.NewBool(false)
		return
	}

	findNextValidPosition(obj)
}

// findNextValidPosition finds the next valid position based on mode
func findNextValidPosition(obj *values.Object) {
	modeValue, exists := obj.Properties["__mode"]
	if !exists {
		obj.Properties["__mode"] = values.NewInt(RECURSIVE_ITERATOR_ITERATOR_LEAVES_ONLY)
		modeValue = obj.Properties["__mode"]
	}

	mode := int(modeValue.ToInt())

	// Get current iterator from stack
	currentIter := getCurrentIteratorFromStack(obj)
	if currentIter == nil {
		obj.Properties["__valid"] = values.NewBool(false)
		return
	}

	// Check if current iterator is valid
	if !callDirectIteratorValid(currentIter) {
		obj.Properties["__valid"] = values.NewBool(false)
		return
	}

	switch mode {
	case RECURSIVE_ITERATOR_ITERATOR_LEAVES_ONLY:
		// In LEAVES_ONLY mode, if current element has children, descend into them
		for callDirectIteratorHasChildren(currentIter) {
			// Check max depth limit
			if exceedsMaxDepth(obj) {
				break
			}

			// Descend into children
			childrenResult := callDirectIteratorGetChildren(currentIter)
			if childrenResult == nil || !childrenResult.IsObject() {
				break
			}

			pushIteratorToStack(obj, childrenResult)
			callDirectIteratorRewind(childrenResult)

			// Update current iterator
			currentIter = getCurrentIteratorFromStack(obj)
			if currentIter == nil || !callDirectIteratorValid(currentIter) {
				// Child iterator is empty, pop and try next
				popIteratorFromStack(obj)
				break
			}
		}
		obj.Properties["__valid"] = values.NewBool(true)

	case RECURSIVE_ITERATOR_ITERATOR_SELF_FIRST:
		// In SELF_FIRST mode, return current position (both containers and leaves)
		obj.Properties["__valid"] = values.NewBool(true)

	case RECURSIVE_ITERATOR_ITERATOR_CHILD_FIRST:
		// In CHILD_FIRST mode, visit children before parent
		// TODO: Implement full CHILD_FIRST logic
		obj.Properties["__valid"] = values.NewBool(true)

	default:
		obj.Properties["__valid"] = values.NewBool(true)
	}
}

// pushIteratorToStack adds an iterator to the stack and increases depth
func pushIteratorToStack(obj *values.Object, iterator *values.Value) {
	stack, _ := obj.Properties["__iteratorStack"]
	depth, _ := obj.Properties["__depth"]

	newDepth := depth.ToInt() + 1
	stack.ArraySet(values.NewInt(newDepth), iterator)
	obj.Properties["__depth"] = values.NewInt(newDepth)
}

// popIteratorFromStack removes the current iterator from stack and decreases depth
func popIteratorFromStack(obj *values.Object) bool {
	depth, exists := obj.Properties["__depth"]
	if !exists || depth.ToInt() <= 0 {
		return false
	}

	newDepth := depth.ToInt() - 1
	obj.Properties["__depth"] = values.NewInt(newDepth)
	return true
}

// exceedsMaxDepth checks if descending would exceed the maximum depth limit
func exceedsMaxDepth(obj *values.Object) bool {
	maxDepth, exists := obj.Properties["__maxDepth"]
	if !exists || maxDepth.ToInt() < 0 {
		return false // No limit
	}

	currentDepth, exists := obj.Properties["__depth"]
	if !exists {
		return false
	}

	return currentDepth.ToInt() >= maxDepth.ToInt()
}

// initRecursiveDirectoryIterator initializes a RecursiveDirectoryIterator with the given path and flags
// This is a helper function to avoid calling constructor with registry context
func initRecursiveDirectoryIterator(thisObj *values.Value, path string, flags int64) error {
	if path == "" {
		return fmt.Errorf("RecursiveDirectoryIterator::__construct(): Path cannot be empty")
	}

	// Check if path exists and is a directory
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("RecursiveDirectoryIterator::__construct(%s): Failed to open directory: No such file or directory", path)
		}
		return fmt.Errorf("RecursiveDirectoryIterator::__construct(%s): Failed to open directory: %v", path, err)
	}

	if !info.IsDir() {
		return fmt.Errorf("RecursiveDirectoryIterator::__construct(%s): Failed to open directory: Not a directory", path)
	}

	// Read directory entries
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("RecursiveDirectoryIterator::__construct(%s): Failed to open directory: %v", path, err)
	}
	defer file.Close()

	entries, err := file.Readdir(-1)
	if err != nil {
		return fmt.Errorf("RecursiveDirectoryIterator::__construct(%s): Failed to read directory: %v", path, err)
	}

	// Create iterator data compatible with FilesystemIterator
	var allEntries []os.FileInfo

	// Add . and .. entries if SKIP_DOTS is not set
	if (flags & 4096) == 0 { // SKIP_DOTS = 4096
		dotEntry := &fakeFileInfo{name: ".", isDir: true}
		dotDotEntry := &fakeFileInfo{name: "..", isDir: true}
		allEntries = make([]os.FileInfo, 0, len(entries)+2)
		allEntries = append(allEntries, dotEntry, dotDotEntry)
		allEntries = append(allEntries, entries...)
	} else {
		allEntries = entries
	}

	iteratorData := &RecursiveDirectoryIteratorData{
		path:       path,
		entries:    allEntries,
		currentIdx: 0,
		flags:      flags,
	}

	objData := thisObj.Data.(*values.Object)
	objData.Properties["_iteratorData"] = &values.Value{
		Type: values.TypeResource,
		Data: iteratorData,
	}

	return nil
}