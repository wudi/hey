package spl

import (
	"fmt"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// GetRecursiveCallbackFilterIteratorClass returns the RecursiveCallbackFilterIterator class descriptor
func GetRecursiveCallbackFilterIteratorClass() *registry.ClassDescriptor {
	// Constructor - inherits from CallbackFilterIterator
	constructorImpl := &registry.Function{
		Name:      "__construct",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 {
				return values.NewNull(), fmt.Errorf("RecursiveCallbackFilterIterator::__construct() expects at least 1 argument")
			}

			thisObj := args[0]

			// Handle VM parameter passing issue - make parameters optional
			var iterator *values.Value = values.NewNull()
			if len(args) > 1 && !args[1].IsNull() {
				iterator = args[1]
			}

			var callback *values.Value = values.NewNull()
			if len(args) > 2 && !args[2].IsNull() {
				callback = args[2]
			}

			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("__construct called on non-object")
			}

			// RecursiveCallbackFilterIterator requires a RecursiveIterator
			if !iterator.IsNull() && !iterator.IsObject() {
				return values.NewNull(), fmt.Errorf("RecursiveCallbackFilterIterator::__construct(): Argument #1 ($iterator) must be of type RecursiveIterator, %s given", iterator.Type)
			}

			// For now, we'll store the callback but won't validate it as callable since
			// the VM doesn't have robust callable validation yet
			objData := thisObj.Data.(*values.Object)
			objData.Properties["__iterator"] = iterator
			objData.Properties["__callback"] = callback

			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{
			{Name: "iterator", Type: "RecursiveIterator"},
			{Name: "callback", Type: "callable"},
		},
	}

	// Get parent methods from CallbackFilterIterator
	parentClass := GetCallbackFilterIteratorClass()
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
			{Name: "callback", Type: "callable"},
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

	// Add getChildren implementation - return RecursiveCallbackFilterIterator wrapping the children
	getChildrenImpl := &registry.Function{
		Name:      "getChildren",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			objData := thisObj.Data.(*values.Object)

			innerIteratorValue := objData.Properties["__iterator"]
			if innerIteratorValue == nil || innerIteratorValue.IsNull() {
				return nil, fmt.Errorf("RecursiveCallbackFilterIterator::getChildren(): No inner iterator")
			}

			innerIterator := innerIteratorValue

			// Get children from the inner iterator
			if innerIterator.IsObject() {
				innerObj := innerIterator.Data.(*values.Object)
				className := innerObj.ClassName
				class, err := ctx.SymbolRegistry().GetClass(className)
				if err != nil {
					return nil, fmt.Errorf("RecursiveCallbackFilterIterator::getChildren(): Inner iterator class not found: %v", err)
				}

				getChildrenMethod, exists := class.Methods["getChildren"]
				if !exists {
					return nil, fmt.Errorf("RecursiveCallbackFilterIterator::getChildren(): Inner iterator does not implement getChildren")
				}

				// Call getChildren on the inner iterator
				getChildrenImpl := getChildrenMethod.Implementation.(*BuiltinMethodImpl)
				childrenResult, err := getChildrenImpl.GetFunction().Builtin(ctx, []*values.Value{innerIterator})
				if err != nil {
					return nil, err
				}

				// Get current callback from this iterator
				callbackValue := objData.Properties["__callback"]

				// Create a new RecursiveCallbackFilterIterator wrapping the children
				childRecursiveCallbackObj := &values.Object{
					ClassName:  "RecursiveCallbackFilterIterator",
					Properties: make(map[string]*values.Value),
				}
				childRecursiveCallbackThis := &values.Value{
					Type: values.TypeObject,
					Data: childRecursiveCallbackObj,
				}

				// Initialize the child RecursiveCallbackFilterIterator
				childRecursiveCallbackObj.Properties["__iterator"] = childrenResult
				childRecursiveCallbackObj.Properties["__callback"] = callbackValue

				return childRecursiveCallbackThis, nil
			}

			return nil, fmt.Errorf("RecursiveCallbackFilterIterator::getChildren(): Inner iterator is not an object")
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
		Name:       "RecursiveCallbackFilterIterator",
		Parent:     "CallbackFilterIterator",
		Interfaces: []string{"Iterator", "OuterIterator", "RecursiveIterator"},
		Traits:     []string{},
		IsAbstract: false,
		IsFinal:    false,
		Methods:    methods,
		Properties: map[string]*registry.PropertyDescriptor{},
		Constants:  constants,
	}
}