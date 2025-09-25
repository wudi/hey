package spl

import (
	"fmt"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// RecursiveTreeIterator flag constants
const (
	RECURSIVE_TREE_ITERATOR_BYPASS_CURRENT = 4
	RECURSIVE_TREE_ITERATOR_BYPASS_KEY     = 8
	RECURSIVE_TREE_ITERATOR_SHOW_TREE      = 16
	RECURSIVE_TREE_ITERATOR_LEAVES_ONLY    = 32
	RECURSIVE_TREE_ITERATOR_CHILD_FIRST    = 2
)

// RecursiveTreeIterator prefix constants
const (
	RECURSIVE_TREE_PREFIX_LEFT        = 0
	RECURSIVE_TREE_PREFIX_MID_HAS_NEXT = 1
	RECURSIVE_TREE_PREFIX_END_HAS_NEXT = 2
	RECURSIVE_TREE_PREFIX_END_LAST     = 3
)

// GetRecursiveTreeIteratorClass returns the RecursiveTreeIterator class descriptor
func GetRecursiveTreeIteratorClass() *registry.ClassDescriptor {
	// Constructor - inherits from RecursiveIteratorIterator
	constructorImpl := &registry.Function{
		Name:      "__construct",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 {
				return values.NewNull(), fmt.Errorf("RecursiveTreeIterator::__construct() expects at least 1 argument")
			}

			thisObj := args[0]

			// Handle VM parameter passing issue - make parameters optional
			var iterator *values.Value = values.NewNull()
			if len(args) > 1 && !args[1].IsNull() {
				iterator = args[1]
			}

			var flags *values.Value = values.NewInt(1) // Default SELF_FIRST
			if len(args) > 2 && !args[2].IsNull() {
				flags = args[2]
			}

			var caching_it_flags *values.Value = values.NewInt(0)
			if len(args) > 3 && !args[3].IsNull() {
				caching_it_flags = args[3]
			}

			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("__construct called on non-object")
			}

			// RecursiveTreeIterator requires a RecursiveIterator
			if !iterator.IsNull() && !iterator.IsObject() {
				return values.NewNull(), fmt.Errorf("RecursiveTreeIterator::__construct(): Argument #1 ($iterator) must be of type RecursiveIterator, %s given", iterator.Type)
			}

			// Initialize properties
			objData := thisObj.Data.(*values.Object)
			objData.Properties["__iterator"] = iterator
			objData.Properties["__flags"] = flags
			objData.Properties["__caching_it_flags"] = caching_it_flags

			// Initialize tree rendering properties
			objData.Properties["__prefixes"] = createDefaultPrefixes()
			objData.Properties["__postfix"] = values.NewString("")

			// Initialize RecursiveIteratorIterator state
			objData.Properties["__depth"] = values.NewInt(0)
			objData.Properties["__max_depth"] = values.NewInt(-1) // No limit

			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{
			{Name: "iterator", Type: "RecursiveIterator"},
			{Name: "flags", Type: "int", DefaultValue: values.NewInt(1)},
			{Name: "caching_it_flags", Type: "int", DefaultValue: values.NewInt(0)},
		},
	}

	// Get parent methods from RecursiveIteratorIterator
	parentClass := GetRecursiveIteratorIteratorClass()
	methods := make(map[string]*registry.MethodDescriptor)

	// Copy all parent methods except __construct and current
	for name, method := range parentClass.Methods {
		if name != "__construct" && name != "current" {
			methods[name] = method
		}
	}

	// Override constructor
	methods["__construct"] = &registry.MethodDescriptor{
		Name:       "__construct",
		Visibility: "public",
		Parameters: []*registry.ParameterDescriptor{
			{Name: "iterator", Type: "RecursiveIterator"},
			{Name: "flags", Type: "int"},
			{Name: "caching_it_flags", Type: "int"},
		},
		Implementation: NewBuiltinMethodImpl(constructorImpl),
	}

	// Override current - returns formatted tree string
	currentImpl := &registry.Function{
		Name:      "current",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			objData := thisObj.Data.(*values.Object)

			innerIteratorValue := objData.Properties["__iterator"]
			if innerIteratorValue == nil || innerIteratorValue.IsNull() {
				return values.NewNull(), nil
			}

			// Get the prefix for current depth
			prefix := getTreePrefix(thisObj)

			// Get the entry (current value)
			entry := getTreeEntry(thisObj)

			// Combine prefix + entry to create tree representation
			treeString := prefix + entry
			return values.NewString(treeString), nil
		},
	}

	// Add getEntry method
	getEntryImpl := &registry.Function{
		Name:      "getEntry",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			objData := thisObj.Data.(*values.Object)

			innerIteratorValue := objData.Properties["__iterator"]
			if innerIteratorValue == nil || innerIteratorValue.IsNull() {
				return values.NewString(""), nil
			}

			// Get current value from inner iterator
			if innerIteratorValue.IsObject() {
				innerObj := innerIteratorValue.Data.(*values.Object)
				className := innerObj.ClassName
				class, err := ctx.SymbolRegistry().GetClass(className)
				if err != nil {
					return values.NewString(""), nil
				}

				currentMethod, exists := class.Methods["current"]
				if !exists {
					return values.NewString(""), nil
				}

				currentImpl := currentMethod.Implementation.(*BuiltinMethodImpl)
				result, err := currentImpl.GetFunction().Builtin(ctx, []*values.Value{innerIteratorValue})
				if err != nil {
					return values.NewString(""), nil
				}

				// Convert to string representation
				if result.IsArray() {
					return values.NewString("Array"), nil
				}
				return values.NewString(result.ToString()), nil
			}

			return values.NewString(""), nil
		},
	}

	// Add getPrefix method
	getPrefixImpl := &registry.Function{
		Name:      "getPrefix",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			prefix := getTreePrefix(thisObj)
			return values.NewString(prefix), nil
		},
	}

	// Add getPostfix method
	getPostfixImpl := &registry.Function{
		Name:      "getPostfix",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			objData := thisObj.Data.(*values.Object)

			postfixValue := objData.Properties["__postfix"]
			if postfixValue == nil {
				return values.NewString(""), nil
			}

			return postfixValue, nil
		},
	}

	// Add setPostfix method
	setPostfixImpl := &registry.Function{
		Name:      "setPostfix",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 2 {
				return values.NewNull(), fmt.Errorf("RecursiveTreeIterator::setPostfix() expects exactly 1 parameter")
			}

			thisObj := args[0]
			postfix := args[1]

			objData := thisObj.Data.(*values.Object)
			objData.Properties["__postfix"] = values.NewString(postfix.ToString())

			return values.NewNull(), nil
		},
	}

	// Add setPrefixPart method
	setPrefixPartImpl := &registry.Function{
		Name:      "setPrefixPart",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 3 {
				return values.NewNull(), fmt.Errorf("RecursiveTreeIterator::setPrefixPart() expects exactly 2 parameters")
			}

			thisObj := args[0]
			part := args[1]
			value := args[2]

			objData := thisObj.Data.(*values.Object)
			prefixes := objData.Properties["__prefixes"]

			if prefixes != nil && prefixes.IsArray() {
				prefixes.ArraySet(part, values.NewString(value.ToString()))
			}

			return values.NewNull(), nil
		},
	}

	// Add new methods
	methods["current"] = &registry.MethodDescriptor{
		Name:           "current",
		Visibility:     "public",
		Parameters:     []*registry.ParameterDescriptor{},
		Implementation: NewBuiltinMethodImpl(currentImpl),
	}

	methods["getEntry"] = &registry.MethodDescriptor{
		Name:           "getEntry",
		Visibility:     "public",
		Parameters:     []*registry.ParameterDescriptor{},
		Implementation: NewBuiltinMethodImpl(getEntryImpl),
	}

	methods["getPrefix"] = &registry.MethodDescriptor{
		Name:           "getPrefix",
		Visibility:     "public",
		Parameters:     []*registry.ParameterDescriptor{},
		Implementation: NewBuiltinMethodImpl(getPrefixImpl),
	}

	methods["getPostfix"] = &registry.MethodDescriptor{
		Name:           "getPostfix",
		Visibility:     "public",
		Parameters:     []*registry.ParameterDescriptor{},
		Implementation: NewBuiltinMethodImpl(getPostfixImpl),
	}

	methods["setPostfix"] = &registry.MethodDescriptor{
		Name:       "setPostfix",
		Visibility: "public",
		Parameters: []*registry.ParameterDescriptor{
			{Name: "postfix", Type: "string"},
		},
		Implementation: NewBuiltinMethodImpl(setPostfixImpl),
	}

	methods["setPrefixPart"] = &registry.MethodDescriptor{
		Name:       "setPrefixPart",
		Visibility: "public",
		Parameters: []*registry.ParameterDescriptor{
			{Name: "part", Type: "int"},
			{Name: "value", Type: "string"},
		},
		Implementation: NewBuiltinMethodImpl(setPrefixPartImpl),
	}

	// Copy constants from parent and add tree-specific ones
	constants := make(map[string]*registry.ConstantDescriptor)
	for name, constant := range parentClass.Constants {
		constants[name] = constant
	}

	// Add RecursiveTreeIterator constants
	constants["BYPASS_CURRENT"] = &registry.ConstantDescriptor{
		Name:  "BYPASS_CURRENT",
		Value: values.NewInt(RECURSIVE_TREE_ITERATOR_BYPASS_CURRENT),
	}
	constants["BYPASS_KEY"] = &registry.ConstantDescriptor{
		Name:  "BYPASS_KEY",
		Value: values.NewInt(RECURSIVE_TREE_ITERATOR_BYPASS_KEY),
	}
	constants["SHOW_TREE"] = &registry.ConstantDescriptor{
		Name:  "SHOW_TREE",
		Value: values.NewInt(RECURSIVE_TREE_ITERATOR_SHOW_TREE),
	}

	return &registry.ClassDescriptor{
		Name:       "RecursiveTreeIterator",
		Parent:     "RecursiveIteratorIterator",
		Interfaces: []string{"Iterator", "OuterIterator"},
		Traits:     []string{},
		IsAbstract: false,
		IsFinal:    false,
		Methods:    methods,
		Properties: map[string]*registry.PropertyDescriptor{},
		Constants:  constants,
	}
}

// Helper functions for tree rendering

func createDefaultPrefixes() *values.Value {
	prefixes := values.NewArray()
	prefixes.ArraySet(values.NewInt(RECURSIVE_TREE_PREFIX_LEFT), values.NewString("| "))
	prefixes.ArraySet(values.NewInt(RECURSIVE_TREE_PREFIX_MID_HAS_NEXT), values.NewString("|-"))
	prefixes.ArraySet(values.NewInt(RECURSIVE_TREE_PREFIX_END_HAS_NEXT), values.NewString("|-"))
	prefixes.ArraySet(values.NewInt(RECURSIVE_TREE_PREFIX_END_LAST), values.NewString("\\-"))
	return prefixes
}

func getTreePrefix(thisObj *values.Value) string {
	objData := thisObj.Data.(*values.Object)
	prefixes := objData.Properties["__prefixes"]

	if prefixes == nil || !prefixes.IsArray() {
		return "|-"
	}

	// For simplicity, return mid-has-next prefix
	// In full implementation, this would calculate based on position in tree
	midPrefix := prefixes.ArrayGet(values.NewInt(RECURSIVE_TREE_PREFIX_MID_HAS_NEXT))
	if midPrefix != nil {
		return midPrefix.ToString()
	}

	return "|-"
}

func getTreeEntry(thisObj *values.Value) string {
	objData := thisObj.Data.(*values.Object)
	innerIteratorValue := objData.Properties["__iterator"]

	if innerIteratorValue == nil || innerIteratorValue.IsNull() {
		return ""
	}

	// This is a simplified implementation
	// In full implementation, this would get the current value from the recursive iterator
	return "Array" // Placeholder
}