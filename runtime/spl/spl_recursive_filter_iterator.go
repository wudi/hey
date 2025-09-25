package spl

import (
	"fmt"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// GetRecursiveFilterIteratorClass returns the RecursiveFilterIterator class descriptor
func GetRecursiveFilterIteratorClass() *registry.ClassDescriptor {
	// Constructor - inherits from FilterIterator
	constructorImpl := &registry.Function{
		Name:      "__construct",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 2 {
				return nil, fmt.Errorf("RecursiveFilterIterator::__construct() expects exactly 1 parameter, %d given", len(args)-1)
			}

			thisObj := args[0]
			iteratorArg := args[1]

			// RecursiveFilterIterator requires a RecursiveIterator
			if !iteratorArg.IsObject() {
				return nil, fmt.Errorf("RecursiveFilterIterator::__construct(): Argument #1 must be RecursiveIterator, %s given", iteratorArg.Type)
			}

			// Store the inner iterator (using same property name as FilterIterator)
			objData := thisObj.Data.(*values.Object)
			objData.Properties["__iterator"] = iteratorArg

			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{
			{Name: "iterator", Type: "RecursiveIterator"},
		},
	}

	// Get parent methods from FilterIterator
	parentClass := GetFilterIteratorClass()
	methods := make(map[string]*registry.MethodDescriptor)

	// Copy all parent methods except __construct and accept
	for name, method := range parentClass.Methods {
		if name != "__construct" {
			methods[name] = method
		}
	}

	// Override constructor
	methods["__construct"] = &registry.MethodDescriptor{
		Name:       "__construct",
		Visibility: "public",
		Parameters: []*registry.ParameterDescriptor{
			{Name: "iterator", Type: "RecursiveIterator"},
		},
		Implementation: NewBuiltinMethodImpl(constructorImpl),
	}

	// hasChildren method - implements RecursiveIterator
	hasChildrenImpl := &registry.Function{
		Name:      "hasChildren",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			objData := thisObj.Data.(*values.Object)

			innerIteratorValue := objData.Properties["__iterator"]
			if innerIteratorValue == nil {
				return values.NewBool(false), nil
			}

			innerIterator := innerIteratorValue

			// Check if the inner iterator is a RecursiveIterator and has hasChildren method
			if innerIterator.IsObject() {
				innerObj := innerIterator.Data.(*values.Object)

				// Get the hasChildren method from the inner iterator's class
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

	// getChildren method - implements RecursiveIterator
	getChildrenImpl := &registry.Function{
		Name:      "getChildren",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			objData := thisObj.Data.(*values.Object)

			innerIteratorValue := objData.Properties["__iterator"]
			if innerIteratorValue == nil {
				return nil, fmt.Errorf("RecursiveFilterIterator::getChildren(): No inner iterator")
			}

			innerIterator := innerIteratorValue

			// Check if the inner iterator has getChildren method
			if innerIterator.IsObject() {
				innerObj := innerIterator.Data.(*values.Object)

				// Get the getChildren method from the inner iterator's class
				className := innerObj.ClassName
				class, err := ctx.SymbolRegistry().GetClass(className)
				if err != nil {
					return nil, fmt.Errorf("RecursiveFilterIterator::getChildren(): Inner iterator class not found: %v", err)
				}

				getChildrenMethod, exists := class.Methods["getChildren"]
				if !exists {
					return nil, fmt.Errorf("RecursiveFilterIterator::getChildren(): Inner iterator does not implement getChildren")
				}

				// Call getChildren on the inner iterator
				getChildrenImpl := getChildrenMethod.Implementation.(*BuiltinMethodImpl)
				childrenResult, err := getChildrenImpl.GetFunction().Builtin(ctx, []*values.Value{innerIterator})
				if err != nil {
					return nil, err
				}

				// Create a new instance of the same filter class wrapping the children
				thisClass := objData.ClassName
				childFilterObj := &values.Object{
					ClassName:  thisClass,
					Properties: make(map[string]*values.Value),
				}
				childFilterThis := &values.Value{
					Type: values.TypeObject,
					Data: childFilterObj,
				}

				// Initialize the child filter with the children iterator
				childFilterObj.Properties["__iterator"] = childrenResult

				return childFilterThis, nil
			}

			return nil, fmt.Errorf("RecursiveFilterIterator::getChildren(): Inner iterator is not an object")
		},
	}

	// Override accept method to make it abstract (will be overridden by concrete classes)
	acceptImpl := &registry.Function{
		Name:      "accept",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			// This is abstract - concrete classes must override this method
			return nil, fmt.Errorf("RecursiveFilterIterator::accept() is abstract and must be implemented by subclasses")
		},
	}

	// Add new methods
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

	methods["accept"] = &registry.MethodDescriptor{
		Name:           "accept",
		Visibility:     "public",
		Parameters:     []*registry.ParameterDescriptor{},
		Implementation: NewBuiltinMethodImpl(acceptImpl),
	}

	// Copy constants from parent
	constants := make(map[string]*registry.ConstantDescriptor)
	for name, constant := range parentClass.Constants {
		constants[name] = constant
	}

	return &registry.ClassDescriptor{
		Name:       "RecursiveFilterIterator",
		Parent:     "FilterIterator",
		Interfaces: []string{"Iterator", "OuterIterator", "RecursiveIterator"},
		Traits:     []string{},
		IsAbstract: true, // This class is abstract
		IsFinal:    false,
		Methods:    methods,
		Properties: map[string]*registry.PropertyDescriptor{},
		Constants:  constants,
	}
}