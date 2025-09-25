package spl

import (
	"fmt"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// GetRecursiveRegexIteratorClass returns the RecursiveRegexIterator class descriptor
func GetRecursiveRegexIteratorClass() *registry.ClassDescriptor {
	// Constructor - inherits from RegexIterator
	constructorImpl := &registry.Function{
		Name:      "__construct",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 {
				return values.NewNull(), fmt.Errorf("RecursiveRegexIterator::__construct() expects at least 1 argument")
			}

			thisObj := args[0]

			// Handle VM parameter passing issue - make parameters optional
			var iterator *values.Value = values.NewNull()
			if len(args) > 1 && !args[1].IsNull() {
				iterator = args[1]
			}

			var regex *values.Value = values.NewNull()
			if len(args) > 2 && !args[2].IsNull() {
				regex = args[2]
			}

			var mode *values.Value = values.NewInt(0) // Default MATCH mode
			if len(args) > 3 && !args[3].IsNull() {
				mode = args[3]
			}

			var flags *values.Value = values.NewInt(0) // Default no flags
			if len(args) > 4 && !args[4].IsNull() {
				flags = args[4]
			}

			var pregFlags *values.Value = values.NewInt(0) // Default no preg flags
			if len(args) > 5 && !args[5].IsNull() {
				pregFlags = args[5]
			}

			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("__construct called on non-object")
			}

			// RecursiveRegexIterator requires a RecursiveIterator
			if !iterator.IsNull() && !iterator.IsObject() {
				return values.NewNull(), fmt.Errorf("RecursiveRegexIterator::__construct(): Argument #1 ($iterator) must be of type RecursiveIterator, %s given", iterator.Type)
			}

			// Store all properties like RegexIterator
			objData := thisObj.Data.(*values.Object)
			objData.Properties["__iterator"] = iterator
			objData.Properties["__regex"] = regex
			objData.Properties["__mode"] = mode
			objData.Properties["__flags"] = flags
			objData.Properties["__preg_flags"] = pregFlags

			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{
			{Name: "iterator", Type: "RecursiveIterator"},
			{Name: "regex", Type: "string"},
			{Name: "mode", Type: "int", DefaultValue: values.NewInt(0)},
			{Name: "flags", Type: "int", DefaultValue: values.NewInt(0)},
			{Name: "preg_flags", Type: "int", DefaultValue: values.NewInt(0)},
		},
	}

	// Get parent methods from RegexIterator
	parentClass := GetRegexIteratorClass()
	methods := make(map[string]*registry.MethodDescriptor)

	// Copy all parent methods except __construct, hasChildren, getChildren
	for name, method := range parentClass.Methods {
		if name != "__construct" && name != "hasChildren" && name != "getChildren" {
			methods[name] = method
		}
	}

	// Override constructor
	methods["__construct"] = &registry.MethodDescriptor{
		Name:       "__construct",
		Visibility: "public",
		Parameters: []*registry.ParameterDescriptor{
			{Name: "iterator", Type: "RecursiveIterator"},
			{Name: "regex", Type: "string"},
			{Name: "mode", Type: "int"},
			{Name: "flags", Type: "int"},
			{Name: "preg_flags", Type: "int"},
		},
		Implementation: NewBuiltinMethodImpl(constructorImpl),
	}

	// Add hasChildren implementation - delegate to inner iterator
	hasChildrenImpl := &registry.Function{
		Name:      "hasChildren",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			objData := thisObj.Data.(*values.Object)

			innerIteratorValue := objData.Properties["__iterator"]
			if innerIteratorValue == nil || innerIteratorValue.IsNull() {
				return values.NewBool(false), nil
			}

			innerIterator := innerIteratorValue

			// Call hasChildren on the inner iterator
			if innerIterator.IsObject() {
				innerObj := innerIterator.Data.(*values.Object)
				className := innerObj.ClassName
				class, err := ctx.SymbolRegistry().GetClass(className)
				if err != nil {
					return values.NewBool(false), nil
				}

				hasChildrenMethod, exists := class.Methods["hasChildren"]
				if !exists {
					return values.NewBool(false), nil
				}

				// Call hasChildren on the inner iterator
				hasChildrenImpl := hasChildrenMethod.Implementation.(*BuiltinMethodImpl)
				result, err := hasChildrenImpl.GetFunction().Builtin(ctx, []*values.Value{innerIterator})
				if err != nil {
					return values.NewBool(false), nil
				}

				return result, nil
			}

			return values.NewBool(false), nil
		},
	}

	// Add getChildren implementation - return RecursiveRegexIterator wrapping the children
	getChildrenImpl := &registry.Function{
		Name:      "getChildren",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			objData := thisObj.Data.(*values.Object)

			innerIteratorValue := objData.Properties["__iterator"]
			if innerIteratorValue == nil || innerIteratorValue.IsNull() {
				return nil, fmt.Errorf("RecursiveRegexIterator::getChildren(): No inner iterator")
			}

			innerIterator := innerIteratorValue

			// Get children from the inner iterator
			if innerIterator.IsObject() {
				innerObj := innerIterator.Data.(*values.Object)
				className := innerObj.ClassName
				class, err := ctx.SymbolRegistry().GetClass(className)
				if err != nil {
					return nil, fmt.Errorf("RecursiveRegexIterator::getChildren(): Inner iterator class not found: %v", err)
				}

				getChildrenMethod, exists := class.Methods["getChildren"]
				if !exists {
					return nil, fmt.Errorf("RecursiveRegexIterator::getChildren(): Inner iterator does not implement getChildren")
				}

				// Call getChildren on the inner iterator
				getChildrenImpl := getChildrenMethod.Implementation.(*BuiltinMethodImpl)
				childrenResult, err := getChildrenImpl.GetFunction().Builtin(ctx, []*values.Value{innerIterator})
				if err != nil {
					return nil, err
				}

				// Get current regex settings from this iterator
				regexValue := objData.Properties["__regex"]
				modeValue := objData.Properties["__mode"]
				flagsValue := objData.Properties["__flags"]
				pregFlagsValue := objData.Properties["__preg_flags"]

				// Create a new RecursiveRegexIterator wrapping the children
				childRecursiveRegexObj := &values.Object{
					ClassName:  "RecursiveRegexIterator",
					Properties: make(map[string]*values.Value),
				}
				childRecursiveRegexThis := &values.Value{
					Type: values.TypeObject,
					Data: childRecursiveRegexObj,
				}

				// Initialize the child RecursiveRegexIterator with same settings
				childRecursiveRegexObj.Properties["__iterator"] = childrenResult
				childRecursiveRegexObj.Properties["__regex"] = regexValue
				childRecursiveRegexObj.Properties["__mode"] = modeValue
				childRecursiveRegexObj.Properties["__flags"] = flagsValue
				childRecursiveRegexObj.Properties["__preg_flags"] = pregFlagsValue

				return childRecursiveRegexThis, nil
			}

			return nil, fmt.Errorf("RecursiveRegexIterator::getChildren(): Inner iterator is not an object")
		},
	}

	// Add the RecursiveIterator methods
	methods["hasChildren"] = &registry.MethodDescriptor{
		Name:           "hasChildren",
		Visibility:     "public",
		Parameters:     []*registry.ParameterDescriptor{},
		Implementation: NewBuiltinMethodImpl(hasChildrenImpl),
	}

	methods["getChildren"] = &registry.MethodDescriptor{
		Name:           "getChildren",
		Visibility:     "public",
		Parameters:     []*registry.ParameterDescriptor{},
		Implementation: NewBuiltinMethodImpl(getChildrenImpl),
	}

	// Copy constants from parent
	constants := make(map[string]*registry.ConstantDescriptor)
	for name, constant := range parentClass.Constants {
		constants[name] = constant
	}

	return &registry.ClassDescriptor{
		Name:       "RecursiveRegexIterator",
		Parent:     "RegexIterator",
		Interfaces: []string{"Iterator", "OuterIterator", "RecursiveIterator"},
		Traits:     []string{},
		IsAbstract: false,
		IsFinal:    false,
		Methods:    methods,
		Properties: map[string]*registry.PropertyDescriptor{},
		Constants:  constants,
	}
}