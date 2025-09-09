package registry

import (
	"strings"

	"github.com/wudi/php-parser/compiler/opcodes"
	"github.com/wudi/php-parser/compiler/values"
)

// ClassBuilder provides a fluent API for building class descriptors
type ClassBuilder struct {
	descriptor *ClassDescriptor
}

// NewClass creates a new class builder
func NewClass(name string) *ClassBuilder {
	return &ClassBuilder{
		descriptor: &ClassDescriptor{
			Name:       name,
			Properties: make(map[string]*PropertyDescriptor),
			Methods:    make(map[string]*MethodDescriptor),
			Constants:  make(map[string]*ConstantDescriptor),
		},
	}
}

// Extends sets the parent class
func (b *ClassBuilder) Extends(parent string) *ClassBuilder {
	b.descriptor.Parent = parent
	return b
}

// Abstract marks the class as abstract
func (b *ClassBuilder) Abstract() *ClassBuilder {
	b.descriptor.IsAbstract = true
	return b
}

// Final marks the class as final
func (b *ClassBuilder) Final() *ClassBuilder {
	b.descriptor.IsFinal = true
	return b
}

// AddProperty adds a property to the class
func (b *ClassBuilder) AddProperty(name, visibility, phpType string, defaultValue *values.Value) *ClassBuilder {
	b.descriptor.Properties[name] = &PropertyDescriptor{
		Name:         name,
		Type:         phpType,
		Visibility:   visibility,
		DefaultValue: defaultValue,
	}
	return b
}

// AddStaticProperty adds a static property to the class
func (b *ClassBuilder) AddStaticProperty(name, visibility, phpType string, defaultValue *values.Value) *ClassBuilder {
	b.descriptor.Properties[name] = &PropertyDescriptor{
		Name:         name,
		Type:         phpType,
		Visibility:   visibility,
		IsStatic:     true,
		DefaultValue: defaultValue,
	}
	return b
}

// AddConstant adds a constant to the class
func (b *ClassBuilder) AddConstant(name string, value *values.Value) *ClassBuilder {
	b.descriptor.Constants[name] = &ConstantDescriptor{
		Name:       name,
		Value:      value,
		Visibility: "public",
	}
	return b
}

// AddMethod starts building a method
func (b *ClassBuilder) AddMethod(name, visibility string) *MethodBuilder {
	return &MethodBuilder{
		classBuilder: b,
		method: &MethodDescriptor{
			Name:       name,
			Visibility: visibility,
			Parameters: []ParameterDescriptor{},
		},
	}
}

// Build finalizes and returns the class descriptor
func (b *ClassBuilder) Build() *ClassDescriptor {
	return b.descriptor
}

// MethodBuilder provides a fluent API for building methods
type MethodBuilder struct {
	classBuilder *ClassBuilder
	method       *MethodDescriptor
}

// Static marks the method as static
func (mb *MethodBuilder) Static() *MethodBuilder {
	mb.method.IsStatic = true
	return mb
}

// Abstract marks the method as abstract
func (mb *MethodBuilder) Abstract() *MethodBuilder {
	mb.method.IsAbstract = true
	return mb
}

// Final marks the method as final
func (mb *MethodBuilder) Final() *MethodBuilder {
	mb.method.IsFinal = true
	return mb
}

// Variadic marks the method as variadic
func (mb *MethodBuilder) Variadic() *MethodBuilder {
	mb.method.IsVariadic = true
	return mb
}

// WithParameter adds a parameter to the method
func (mb *MethodBuilder) WithParameter(name, phpType string) *MethodBuilder {
	mb.method.Parameters = append(mb.method.Parameters, ParameterDescriptor{
		Name: name,
		Type: phpType,
	})
	return mb
}

// WithOptionalParameter adds an optional parameter with default value
func (mb *MethodBuilder) WithOptionalParameter(name, phpType string, defaultValue *values.Value) *MethodBuilder {
	mb.method.Parameters = append(mb.method.Parameters, ParameterDescriptor{
		Name:         name,
		Type:         phpType,
		HasDefault:   true,
		DefaultValue: defaultValue,
	})
	return mb
}

// WithReferenceParameter adds a reference parameter
func (mb *MethodBuilder) WithReferenceParameter(name, phpType string) *MethodBuilder {
	mb.method.Parameters = append(mb.method.Parameters, ParameterDescriptor{
		Name:        name,
		Type:        phpType,
		IsReference: true,
	})
	return mb
}

// WithNativeHandler sets a Go native function as the implementation
func (mb *MethodBuilder) WithNativeHandler(handler func(ExecutionContext, []*values.Value) (*values.Value, error)) *MethodBuilder {
	mb.method.Implementation = &NativeMethodImpl{Handler: handler}
	return mb
}

// WithRuntimeHandler sets a runtime handler as the implementation
func (mb *MethodBuilder) WithRuntimeHandler(handler func(ExecutionContext, []*values.Value) (*values.Value, error)) *MethodBuilder {
	mb.method.Implementation = &RuntimeHandlerImpl{Handler: handler}
	return mb
}

// WithBytecode sets bytecode instructions as the implementation
func (mb *MethodBuilder) WithBytecode(instructions []opcodes.Instruction, constants []*values.Value, localVars int) *MethodBuilder {
	mb.method.Implementation = &BytecodeMethodImpl{
		Instructions: instructions,
		Constants:    constants,
		LocalVars:    localVars,
	}
	return mb
}

// Done completes the method and returns to class builder
func (mb *MethodBuilder) Done() *ClassBuilder {
	mb.classBuilder.descriptor.Methods[mb.method.Name] = mb.method
	return mb.classBuilder
}

// Helper functions for common method patterns

// Constructor creates a constructor method builder
func (b *ClassBuilder) Constructor() *MethodBuilder {
	return b.AddMethod("__construct", "public")
}

// Destructor creates a destructor method builder
func (b *ClassBuilder) Destructor() *MethodBuilder {
	return b.AddMethod("__destruct", "public")
}

// ToString creates a __toString method builder
func (b *ClassBuilder) ToString() *MethodBuilder {
	return b.AddMethod("__toString", "public")
}

// Getter creates a getter method
func (b *ClassBuilder) Getter(propertyName string, returnType string) *MethodBuilder {
	methodName := "get" + strings.ToUpper(propertyName[:1]) + propertyName[1:]
	return b.AddMethod(methodName, "public")
}

// Setter creates a setter method
func (b *ClassBuilder) Setter(propertyName string, paramType string) *MethodBuilder {
	methodName := "set" + strings.ToUpper(propertyName[:1]) + propertyName[1:]
	return b.AddMethod(methodName, "public").
		WithParameter("value", paramType)
}

// strings package already imported at top

// BuiltinClass marks the class as built-in and registers it
func (b *ClassBuilder) BuiltinClass(extensionName string) *ClassBuilder {
	if b.descriptor.Metadata == nil {
		b.descriptor.Metadata = &ClassMetadata{}
	}
	b.descriptor.Metadata.IsBuiltin = true
	b.descriptor.Metadata.ExtensionName = extensionName
	return b
}

// Register registers the class in the global registry
func (b *ClassBuilder) Register() error {
	if b.descriptor.Metadata != nil && b.descriptor.Metadata.IsBuiltin {
		return RegisterBuiltinClass(b.descriptor)
	}
	return GlobalRegistry.RegisterClass(b.descriptor)
}

// BuildAndRegister builds and registers the class in one step
func (b *ClassBuilder) BuildAndRegister() (*ClassDescriptor, error) {
	class := b.Build()
	var err error

	if class.Metadata != nil && class.Metadata.IsBuiltin {
		err = RegisterBuiltinClass(class)
	} else {
		err = GlobalRegistry.RegisterClass(class)
	}

	if err != nil {
		return nil, err
	}

	return class, nil
}
