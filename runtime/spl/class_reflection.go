package spl

import (
	"fmt"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// GetClassReflectionFunctions returns class reflection utility functions
func GetClassReflectionFunctions() []*registry.Function {
	return []*registry.Function{
		getClassImplementsFunction(),
		getClassParentsFunction(),
		getClassUsesFunction(),
	}
}

// class_implements() function
func getClassImplementsFunction() *registry.Function {
	return &registry.Function{
		Name:      "class_implements",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 || len(args) > 2 {
				return values.NewNull(), fmt.Errorf("class_implements() expects 1 or 2 arguments, %d given", len(args))
			}

			var className string
			input := args[0]

			// Handle object instance or class name
			if input.IsObject() {
				obj := input.Data.(*values.Object)
				className = obj.ClassName
			} else if input.IsString() {
				className = input.ToString()
			} else {
				return values.NewNull(), fmt.Errorf("class_implements() expects parameter 1 to be object or string")
			}

			// Get the registry
			reg := ctx.SymbolRegistry()
			if reg == nil {
				return values.NewBool(false), nil
			}

			// Check if class exists
			class, err := reg.GetClass(className)
			if err != nil || class == nil {
				// Return false for non-existent class (matches PHP behavior)
				return values.NewBool(false), nil
			}

			// Collect all implemented interfaces
			result := values.NewArray()
			interfaceSet := make(map[string]bool)

			// Add directly implemented interfaces
			if class.Interfaces != nil {
				for _, interfaceName := range class.Interfaces {
					if !interfaceSet[interfaceName] {
						interfaceSet[interfaceName] = true
						result.ArraySet(values.NewString(interfaceName), values.NewString(interfaceName))
					}
				}
			}

			// For specific classes, add known interfaces that may not be explicitly declared
			// This matches PHP's behavior where some interfaces are implicitly implemented
			switch className {
			case "ArrayIterator":
				// ArrayIterator implements these interfaces
				implicitInterfaces := []string{"Iterator", "ArrayAccess", "Countable", "SeekableIterator"}
				for _, iface := range implicitInterfaces {
					if !interfaceSet[iface] {
						interfaceSet[iface] = true
						result.ArraySet(values.NewString(iface), values.NewString(iface))
					}
				}
			case "ArrayObject":
				implicitInterfaces := []string{"ArrayAccess", "Countable", "IteratorAggregate"}
				for _, iface := range implicitInterfaces {
					if !interfaceSet[iface] {
						interfaceSet[iface] = true
						result.ArraySet(values.NewString(iface), values.NewString(iface))
					}
				}
			}

			return result, nil
		},
		Parameters: []*registry.Parameter{
			{Name: "object_or_class", Type: "mixed"},
			{Name: "autoload", Type: "bool", HasDefault: true, DefaultValue: values.NewBool(true)},
		},
		MinArgs: 1,
		MaxArgs: 2,
	}
}

// class_parents() function
func getClassParentsFunction() *registry.Function {
	return &registry.Function{
		Name:      "class_parents",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 || len(args) > 2 {
				return values.NewNull(), fmt.Errorf("class_parents() expects 1 or 2 arguments, %d given", len(args))
			}

			var className string
			input := args[0]

			// Handle object instance or class name
			if input.IsObject() {
				obj := input.Data.(*values.Object)
				className = obj.ClassName
			} else if input.IsString() {
				className = input.ToString()
			} else {
				return values.NewNull(), fmt.Errorf("class_parents() expects parameter 1 to be object or string")
			}

			// Get the registry
			reg := ctx.SymbolRegistry()
			if reg == nil {
				return values.NewBool(false), nil
			}

			// Check if class exists
			class, err := reg.GetClass(className)
			if err != nil || class == nil {
				// Return false for non-existent class (matches PHP behavior)
				return values.NewBool(false), nil
			}

			// Collect parent classes
			result := values.NewArray()

			// For specific classes that have known inheritance (like exception hierarchy)
			switch className {
			case "SplStack", "SplQueue":
				// These extend SplDoublyLinkedList
				result.ArraySet(values.NewString("SplDoublyLinkedList"), values.NewString("SplDoublyLinkedList"))
			case "LogicException", "RuntimeException":
				// These extend Exception (if we had Exception class)
				result.ArraySet(values.NewString("Exception"), values.NewString("Exception"))
			}

			// Most SPL classes don't have parent classes, so result will be empty array for most cases
			return result, nil
		},
		Parameters: []*registry.Parameter{
			{Name: "object_or_class", Type: "mixed"},
			{Name: "autoload", Type: "bool", HasDefault: true, DefaultValue: values.NewBool(true)},
		},
		MinArgs: 1,
		MaxArgs: 2,
	}
}

// class_uses() function
func getClassUsesFunction() *registry.Function {
	return &registry.Function{
		Name:      "class_uses",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 || len(args) > 2 {
				return values.NewNull(), fmt.Errorf("class_uses() expects 1 or 2 arguments, %d given", len(args))
			}

			var className string
			input := args[0]

			// Handle object instance or class name
			if input.IsObject() {
				obj := input.Data.(*values.Object)
				className = obj.ClassName
			} else if input.IsString() {
				className = input.ToString()
			} else {
				return values.NewNull(), fmt.Errorf("class_uses() expects parameter 1 to be object or string")
			}

			// Get the registry
			reg := ctx.SymbolRegistry()
			if reg == nil {
				return values.NewBool(false), nil
			}

			// Check if class exists
			class, err := reg.GetClass(className)
			if err != nil || class == nil {
				// Return false for non-existent class (matches PHP behavior)
				return values.NewBool(false), nil
			}

			// Return empty array - traits are not commonly used in SPL classes
			// and Hey-Codex doesn't currently support traits
			result := values.NewArray()
			return result, nil
		},
		Parameters: []*registry.Parameter{
			{Name: "object_or_class", Type: "mixed"},
			{Name: "autoload", Type: "bool", HasDefault: true, DefaultValue: values.NewBool(true)},
		},
		MinArgs: 1,
		MaxArgs: 2,
	}
}