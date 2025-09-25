package spl

import (
	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// GetSplStackClass returns the SplStack class descriptor
// SplStack extends SplDoublyLinkedList with LIFO behavior
func GetSplStackClass() *registry.ClassDescriptor {
	// SplStack is just SplDoublyLinkedList with restricted iteration mode (LIFO)
	// It uses the same underlying implementation but provides different method names

	baseClass := GetSplDoublyLinkedListClass()

	// Override methods to provide stack-specific behavior
	methods := make(map[string]*registry.MethodDescriptor)

	// Copy all methods from base class first
	for name, method := range baseClass.Methods {
		methods[name] = method
	}

	// Add stack-specific methods

	// Constructor - sets LIFO mode
	constructorImpl := &registry.Function{
		Name:      "__construct",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			// Call parent constructor
			parentImpl := baseClass.Methods["__construct"].Implementation.(*BuiltinMethodImpl)
			result, err := parentImpl.GetFunction().Builtin(ctx, args)
			if err != nil {
				return result, err
			}

			// Set iterator mode to LIFO (2) - though this is just for completeness
			if len(args) > 0 && args[0].IsObject() {
				obj := args[0].Data.(*values.Object)
				if obj.Properties == nil {
					obj.Properties = make(map[string]*values.Value)
				}
				obj.Properties["__iterator_mode"] = values.NewInt(2) // LIFO
			}

			return result, nil
		},
		Parameters: []*registry.Parameter{},
	}

	methods["__construct"] = &registry.MethodDescriptor{
		Name:           "__construct",
		Visibility:     "public",
		Parameters:     []*registry.ParameterDescriptor{},
		Implementation: NewBuiltinMethodImpl(constructorImpl),
	}

	return &registry.ClassDescriptor{
		Name:       "SplStack",
		Parent:     "SplDoublyLinkedList",
		Interfaces: []string{"Iterator", "Countable"},
		Properties: make(map[string]*registry.PropertyDescriptor),
		Methods:    methods,
		Constants:  make(map[string]*registry.ConstantDescriptor),
	}
}

// GetSplQueueClass returns the SplQueue class descriptor
// SplQueue extends SplDoublyLinkedList with FIFO behavior
func GetSplQueueClass() *registry.ClassDescriptor {
	baseClass := GetSplDoublyLinkedListClass()

	// Override methods to provide queue-specific behavior
	methods := make(map[string]*registry.MethodDescriptor)

	// Copy all methods from base class first
	for name, method := range baseClass.Methods {
		methods[name] = method
	}

	// Add queue-specific methods

	// Constructor - sets FIFO mode
	constructorImpl := &registry.Function{
		Name:      "__construct",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			// Call parent constructor
			parentImpl := baseClass.Methods["__construct"].Implementation.(*BuiltinMethodImpl)
			result, err := parentImpl.GetFunction().Builtin(ctx, args)
			if err != nil {
				return result, err
			}

			// Set iterator mode to FIFO (0) - default but explicit
			if len(args) > 0 && args[0].IsObject() {
				obj := args[0].Data.(*values.Object)
				if obj.Properties == nil {
					obj.Properties = make(map[string]*values.Value)
				}
				obj.Properties["__iterator_mode"] = values.NewInt(0) // FIFO
			}

			return result, nil
		},
		Parameters: []*registry.Parameter{},
	}

	// enqueue() method - alias for push()
	enqueueImpl := &registry.Function{
		Name:      "enqueue",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			// Just call push()
			pushImpl := baseClass.Methods["push"].Implementation.(*BuiltinMethodImpl)
			return pushImpl.GetFunction().Builtin(ctx, args)
		},
		Parameters: []*registry.Parameter{
			{Name: "value", Type: "mixed"},
		},
	}

	// dequeue() method - alias for shift()
	dequeueImpl := &registry.Function{
		Name:      "dequeue",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			// Just call shift()
			shiftImpl := baseClass.Methods["shift"].Implementation.(*BuiltinMethodImpl)
			return shiftImpl.GetFunction().Builtin(ctx, args)
		},
		Parameters: []*registry.Parameter{},
	}

	methods["__construct"] = &registry.MethodDescriptor{
		Name:           "__construct",
		Visibility:     "public",
		Parameters:     []*registry.ParameterDescriptor{},
		Implementation: NewBuiltinMethodImpl(constructorImpl),
	}

	methods["enqueue"] = &registry.MethodDescriptor{
		Name:       "enqueue",
		Visibility: "public",
		Parameters: []*registry.ParameterDescriptor{
			{Name: "value", Type: "mixed"},
		},
		Implementation: NewBuiltinMethodImpl(enqueueImpl),
	}

	methods["dequeue"] = &registry.MethodDescriptor{
		Name:           "dequeue",
		Visibility:     "public",
		Parameters:     []*registry.ParameterDescriptor{},
		Implementation: NewBuiltinMethodImpl(dequeueImpl),
	}

	return &registry.ClassDescriptor{
		Name:       "SplQueue",
		Parent:     "SplDoublyLinkedList",
		Interfaces: []string{"Iterator", "Countable"},
		Properties: make(map[string]*registry.PropertyDescriptor),
		Methods:    methods,
		Constants:  make(map[string]*registry.ConstantDescriptor),
	}
}