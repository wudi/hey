package spl

import (
	"fmt"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// GetParentIteratorClass returns the ParentIterator class descriptor
func GetParentIteratorClass() *registry.ClassDescriptor {
	// Constructor - inherits from RecursiveFilterIterator
	constructorImpl := &registry.Function{
		Name:      "__construct",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 2 {
				return nil, fmt.Errorf("ParentIterator::__construct() expects exactly 1 parameter, %d given", len(args)-1)
			}

			thisObj := args[0]
			iteratorArg := args[1]

			// ParentIterator requires a RecursiveIterator
			if !iteratorArg.IsObject() {
				return nil, fmt.Errorf("ParentIterator::__construct(): Argument #1 ($iterator) must be of type RecursiveIterator, %s given", iteratorArg.Type)
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

	// Get parent methods from RecursiveFilterIterator
	parentClass := GetRecursiveFilterIteratorClass()
	methods := make(map[string]*registry.MethodDescriptor)

	// Copy all parent methods except __construct and accept
	for name, method := range parentClass.Methods {
		if name != "__construct" && name != "accept" {
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

	// Override accept method - the core logic of ParentIterator
	acceptImpl := &registry.Function{
		Name:      "accept",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			objData := thisObj.Data.(*values.Object)

			innerIteratorValue := objData.Properties["__iterator"]
			if innerIteratorValue == nil {
				return values.NewBool(false), nil
			}

			innerIterator := innerIteratorValue

			// ParentIterator accepts elements that have children
			// This means calling hasChildren() on the inner iterator
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

				// ParentIterator accepts items that have children
				return result, nil
			}

			return values.NewBool(false), nil
		},
	}

	// Override getChildren to return ParentIterator wrapping the children
	getChildrenImpl := &registry.Function{
		Name:      "getChildren",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			objData := thisObj.Data.(*values.Object)

			innerIteratorValue := objData.Properties["__iterator"]
			if innerIteratorValue == nil {
				return nil, fmt.Errorf("ParentIterator::getChildren(): No inner iterator")
			}

			innerIterator := innerIteratorValue

			// Get children from the inner iterator
			if innerIterator.IsObject() {
				innerObj := innerIterator.Data.(*values.Object)

				// Get the getChildren method from the inner iterator's class
				className := innerObj.ClassName
				class, err := ctx.SymbolRegistry().GetClass(className)
				if err != nil {
					return nil, fmt.Errorf("ParentIterator::getChildren(): Inner iterator class not found: %v", err)
				}

				getChildrenMethod, exists := class.Methods["getChildren"]
				if !exists {
					return nil, fmt.Errorf("ParentIterator::getChildren(): Inner iterator does not implement getChildren")
				}

				// Call getChildren on the inner iterator
				getChildrenImpl := getChildrenMethod.Implementation.(*BuiltinMethodImpl)
				childrenResult, err := getChildrenImpl.GetFunction().Builtin(ctx, []*values.Value{innerIterator})
				if err != nil {
					return nil, err
				}

				// Create a new ParentIterator wrapping the children
				childParentObj := &values.Object{
					ClassName:  "ParentIterator",
					Properties: make(map[string]*values.Value),
				}
				childParentThis := &values.Value{
					Type: values.TypeObject,
					Data: childParentObj,
				}

				// Initialize the child ParentIterator with the children iterator
				childParentObj.Properties["__iterator"] = childrenResult

				return childParentThis, nil
			}

			return nil, fmt.Errorf("ParentIterator::getChildren(): Inner iterator is not an object")
		},
	}

	// Override methods
	methods["accept"] = &registry.MethodDescriptor{
		Name:           "accept",
		Visibility:     "public",
		Parameters:     []*registry.ParameterDescriptor{},
		Implementation: NewBuiltinMethodImpl(acceptImpl),
	}

	methods["getChildren"] = &registry.MethodDescriptor{
		Name:           "getChildren",
		Visibility:     "public",
		Parameters:     []*registry.ParameterDescriptor{},
		Implementation: NewBuiltinMethodImpl(getChildrenImpl),
	}

	// Override rewind to position on first accepted element
	rewindImpl := &registry.Function{
		Name:      "rewind",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			objData := thisObj.Data.(*values.Object)

			innerIteratorValue := objData.Properties["__iterator"]
			if innerIteratorValue == nil {
				return values.NewNull(), nil
			}

			innerIterator := innerIteratorValue

			// Rewind inner iterator
			if innerIterator.IsObject() {
				innerObj := innerIterator.Data.(*values.Object)
				className := innerObj.ClassName
				class, err := ctx.SymbolRegistry().GetClass(className)
				if err != nil {
					return values.NewNull(), nil
				}

				if rewindMethod, exists := class.Methods["rewind"]; exists {
					rewindImpl := rewindMethod.Implementation.(*BuiltinMethodImpl)
					_, err := rewindImpl.GetFunction().Builtin(ctx, []*values.Value{innerIterator})
					if err != nil {
						return values.NewNull(), nil
					}
				}

				// Find first accepted element
				return findNextAccepted(ctx, thisObj)
			}

			return values.NewNull(), nil
		},
	}

	// Override next to advance to next accepted element
	nextImpl := &registry.Function{
		Name:      "next",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			objData := thisObj.Data.(*values.Object)

			innerIteratorValue := objData.Properties["__iterator"]
			if innerIteratorValue == nil {
				return values.NewNull(), nil
			}

			innerIterator := innerIteratorValue

			// Move inner iterator to next
			if innerIterator.IsObject() {
				innerObj := innerIterator.Data.(*values.Object)
				className := innerObj.ClassName
				class, err := ctx.SymbolRegistry().GetClass(className)
				if err != nil {
					return values.NewNull(), nil
				}

				if nextMethod, exists := class.Methods["next"]; exists {
					nextImpl := nextMethod.Implementation.(*BuiltinMethodImpl)
					_, err := nextImpl.GetFunction().Builtin(ctx, []*values.Value{innerIterator})
					if err != nil {
						return values.NewNull(), nil
					}
				}

				// Find next accepted element
				return findNextAccepted(ctx, thisObj)
			}

			return values.NewNull(), nil
		},
	}

	// Override valid to check if current position has accepted element
	validImpl := &registry.Function{
		Name:      "valid",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			thisObj := args[0]
			objData := thisObj.Data.(*values.Object)

			innerIteratorValue := objData.Properties["__iterator"]
			if innerIteratorValue == nil {
				return values.NewBool(false), nil
			}

			innerIterator := innerIteratorValue

			// Check if inner iterator is valid
			if innerIterator.IsObject() {
				innerObj := innerIterator.Data.(*values.Object)
				className := innerObj.ClassName
				class, err := ctx.SymbolRegistry().GetClass(className)
				if err != nil {
					return values.NewBool(false), nil
				}

				if validMethod, exists := class.Methods["valid"]; exists {
					validImpl := validMethod.Implementation.(*BuiltinMethodImpl)
					validResult, err := validImpl.GetFunction().Builtin(ctx, []*values.Value{innerIterator})
					if err != nil || !validResult.ToBool() {
						return values.NewBool(false), nil
					}

					// If valid, check if accepted (use our own accept method)
					result, err := acceptImpl.Builtin(ctx, []*values.Value{thisObj})
					if err != nil {
						return values.NewBool(false), nil
					}
					return result, nil
				}
			}

			return values.NewBool(false), nil
		},
	}

	methods["rewind"] = &registry.MethodDescriptor{
		Name:           "rewind",
		Visibility:     "public",
		Parameters:     []*registry.ParameterDescriptor{},
		Implementation: NewBuiltinMethodImpl(rewindImpl),
	}

	methods["next"] = &registry.MethodDescriptor{
		Name:           "next",
		Visibility:     "public",
		Parameters:     []*registry.ParameterDescriptor{},
		Implementation: NewBuiltinMethodImpl(nextImpl),
	}

	methods["valid"] = &registry.MethodDescriptor{
		Name:           "valid",
		Visibility:     "public",
		Parameters:     []*registry.ParameterDescriptor{},
		Implementation: NewBuiltinMethodImpl(validImpl),
	}

	// Copy constants from parent
	constants := make(map[string]*registry.ConstantDescriptor)
	for name, constant := range parentClass.Constants {
		constants[name] = constant
	}

	return &registry.ClassDescriptor{
		Name:       "ParentIterator",
		Parent:     "RecursiveFilterIterator",
		Interfaces: []string{"Iterator", "OuterIterator", "RecursiveIterator"},
		Traits:     []string{},
		IsAbstract: false, // This class is concrete
		IsFinal:    false,
		Methods:    methods,
		Properties: map[string]*registry.PropertyDescriptor{},
		Constants:  constants,
	}
}

// findNextAccepted advances the inner iterator until an accepted element is found
func findNextAccepted(ctx registry.BuiltinCallContext, thisObj *values.Value) (*values.Value, error) {
	objData := thisObj.Data.(*values.Object)
	innerIteratorValue := objData.Properties["__iterator"]
	if innerIteratorValue == nil {
		return values.NewNull(), nil
	}

	innerIterator := innerIteratorValue

	if innerIterator.IsObject() {
		innerObj := innerIterator.Data.(*values.Object)
		className := innerObj.ClassName
		class, err := ctx.SymbolRegistry().GetClass(className)
		if err != nil {
			return values.NewNull(), nil
		}

		// Get methods from inner iterator
		validMethod, hasValid := class.Methods["valid"]
		nextMethod, hasNext := class.Methods["next"]

		if !hasValid || !hasNext {
			return values.NewNull(), nil
		}

		validImpl := validMethod.Implementation.(*BuiltinMethodImpl)
		nextImpl := nextMethod.Implementation.(*BuiltinMethodImpl)

		// Get accept method from ParentIterator class
		parentClass, err := ctx.SymbolRegistry().GetClass("ParentIterator")
		if err != nil {
			return values.NewNull(), nil
		}

		acceptMethod, hasAccept := parentClass.Methods["accept"]
		if !hasAccept {
			return values.NewNull(), nil
		}
		acceptImpl := acceptMethod.Implementation.(*BuiltinMethodImpl)

		// Loop until we find an accepted element or run out of elements
		for {
			// Check if inner iterator is still valid
			validResult, err := validImpl.GetFunction().Builtin(ctx, []*values.Value{innerIterator})
			if err != nil || !validResult.ToBool() {
				break
			}

			// Check if current element is accepted
			acceptResult, err := acceptImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
			if err != nil {
				// If accept fails, move to next
			} else if acceptResult.ToBool() {
				// Found an accepted element
				return values.NewNull(), nil
			}

			// Move to next element
			_, err = nextImpl.GetFunction().Builtin(ctx, []*values.Value{innerIterator})
			if err != nil {
				break
			}
		}
	}

	return values.NewNull(), nil
}