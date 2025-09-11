package registry

import (
	"fmt"
	"strings"
	"sync"

	"github.com/wudi/hey/compiler/opcodes"
	"github.com/wudi/hey/compiler/values"
)

// MethodType defines the type of method implementation
type MethodType int

const (
	MethodTypeBytecode MethodType = iota // Compiled bytecode
	MethodTypeNative                     // Go native function
	MethodTypeHandler                    // Runtime handler
)

// ExecutionContext interface to avoid circular imports
type ExecutionContext interface {
	WriteOutput(output string)
	HasFunction(name string) bool
	ExecuteBytecodeMethod(instructions []opcodes.Instruction, constants []*values.Value, args []*values.Value) (*values.Value, error)
	ExecuteBytecodeMethodWithParams(instructions []opcodes.Instruction, constants []*values.Value, parameters []ParameterInfo, args []*values.Value) (*values.Value, error)
}

// MethodImplementation is the unified interface for method implementations
type MethodImplementation interface {
	Execute(ctx ExecutionContext, args []*values.Value) (*values.Value, error)
	GetType() MethodType
}

// ParameterInfo contains information about a method parameter
type ParameterInfo struct {
	Name         string
	HasDefault   bool
	DefaultValue *values.Value
	IsVariadic   bool
}

// BytecodeMethodImpl implements method via bytecode execution
type BytecodeMethodImpl struct {
	Instructions []opcodes.Instruction
	Constants    []*values.Value
	LocalVars    int
	Parameters   []ParameterInfo
}

func (b *BytecodeMethodImpl) Execute(ctx ExecutionContext, args []*values.Value) (*values.Value, error) {
	// Execute bytecode in VM - this will be implemented in VM layer
	return ctx.ExecuteBytecodeMethodWithParams(b.Instructions, b.Constants, b.Parameters, args)
}

func (b *BytecodeMethodImpl) GetType() MethodType { return MethodTypeBytecode }

// NativeMethodImpl implements method via Go native function
type NativeMethodImpl struct {
	Handler func(ctx ExecutionContext, args []*values.Value) (*values.Value, error)
}

func (n *NativeMethodImpl) Execute(ctx ExecutionContext, args []*values.Value) (*values.Value, error) {
	return n.Handler(ctx, args)
}

func (n *NativeMethodImpl) GetType() MethodType { return MethodTypeNative }

// RuntimeHandlerImpl implements method via runtime handler
type RuntimeHandlerImpl struct {
	Handler func(ctx ExecutionContext, args []*values.Value) (*values.Value, error)
}

func (r *RuntimeHandlerImpl) Execute(ctx ExecutionContext, args []*values.Value) (*values.Value, error) {
	// Bridge to runtime handler interface
	return r.Handler(ctx, args)
}

func (r *RuntimeHandlerImpl) GetType() MethodType { return MethodTypeHandler }

// Remove duplicate ExecutionContext interface

// Unified class descriptor
type ClassDescriptor struct {
	// Basic information
	Name       string
	Parent     string
	IsAbstract bool
	IsFinal    bool

	// Properties
	Properties map[string]*PropertyDescriptor

	// Methods - supports multiple implementation types
	Methods map[string]*MethodDescriptor

	// Constants
	Constants map[string]*ConstantDescriptor

	// Metadata
	Metadata *ClassMetadata
}

// PropertyDescriptor describes a class property
type PropertyDescriptor struct {
	Name         string
	Type         string
	Visibility   string // public, private, protected
	IsStatic     bool
	DefaultValue *values.Value
}

// MethodDescriptor describes a class method with unified implementation
type MethodDescriptor struct {
	Name       string
	Visibility string
	IsStatic   bool
	IsAbstract bool
	IsFinal    bool
	Parameters []ParameterDescriptor

	// Unified method implementation
	Implementation MethodImplementation
	IsVariadic     bool
}

// ParameterDescriptor describes method parameters
type ParameterDescriptor struct {
	Name         string
	Type         string
	IsReference  bool
	HasDefault   bool
	DefaultValue *values.Value
}

// ConstantDescriptor describes class constants
type ConstantDescriptor struct {
	Name       string
	Value      *values.Value
	Visibility string // public, private, protected
	Type       string // Type hint for PHP 8.3+
	IsFinal    bool   // final const
	IsAbstract bool   // abstract const (interfaces/abstract classes)
}

// ClassMetadata contains additional class information
type ClassMetadata struct {
	IsBuiltin     bool
	ExtensionName string
	LoadOrder     int
}

// ClassRegistry is the unified registration system for all classes
type ClassRegistry struct {
	mu sync.RWMutex

	// Core registry
	classes map[string]*ClassDescriptor

	// Class loaders for lazy loading
	loaders map[string]ClassLoader

	// Built-in flags to prevent overrides
	builtinClasses map[string]bool
}

// ClassLoader interface for lazy class loading
type ClassLoader interface {
	LoadClass(name string) (*ClassDescriptor, error)
	CanLoad(name string) bool
}

// NewClassRegistry creates a new unified class registry
func NewClassRegistry() *ClassRegistry {
	return &ClassRegistry{
		classes:        make(map[string]*ClassDescriptor),
		loaders:        make(map[string]ClassLoader),
		builtinClasses: make(map[string]bool),
	}
}

// RegisterClass registers a class in the registry
func (r *ClassRegistry) RegisterClass(class *ClassDescriptor) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := strings.ToLower(class.Name)

	// Check for conflicts with built-ins
	if r.builtinClasses[name] && (class.Metadata == nil || !class.Metadata.IsBuiltin) {
		return fmt.Errorf("cannot override built-in class: %s", name)
	}

	// Check for existing registration
	if _, exists := r.classes[name]; exists {
		return fmt.Errorf("class already registered: %s", name)
	}

	r.classes[name] = class

	if class.Metadata != nil && class.Metadata.IsBuiltin {
		r.builtinClasses[name] = true
	}

	return nil
}

// RegisterClassLoader registers a class loader for lazy loading
func (r *ClassRegistry) RegisterClassLoader(pattern string, loader ClassLoader) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.loaders[pattern] = loader
}

// GetClass retrieves a class, loading it if necessary
func (r *ClassRegistry) GetClass(name string) (*ClassDescriptor, error) {
	r.mu.RLock()
	if class, exists := r.classes[strings.ToLower(name)]; exists {
		r.mu.RUnlock()
		return class, nil
	}
	r.mu.RUnlock()

	// Try lazy loading
	return r.loadClass(name)
}

// loadClass attempts to load a class using registered loaders
func (r *ClassRegistry) loadClass(name string) (*ClassDescriptor, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if already loaded (double-check pattern)
	if class, exists := r.classes[strings.ToLower(name)]; exists {
		return class, nil
	}

	// Try each loader
	for pattern, loader := range r.loaders {
		if loader.CanLoad(name) {
			class, err := loader.LoadClass(name)
			if err != nil {
				continue // Try next loader
			}

			// Register the loaded class
			r.classes[strings.ToLower(name)] = class
			return class, nil
		}
		_ = pattern // Use pattern for matching in future
	}

	return nil, fmt.Errorf("class not found: %s", name)
}

// HasClass checks if a class exists
func (r *ClassRegistry) HasClass(name string) bool {
	_, err := r.GetClass(name)
	return err == nil
}

// IsBuiltinClass checks if a class is built-in
func (r *ClassRegistry) IsBuiltinClass(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.builtinClasses[strings.ToLower(name)]
}

// GetAllClasses returns all registered classes
func (r *ClassRegistry) GetAllClasses() map[string]*ClassDescriptor {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]*ClassDescriptor, len(r.classes))
	for k, v := range r.classes {
		// Deep copy to prevent external mutation
		result[k] = r.copyClassDescriptor(v)
	}
	return result
}

// copyClassDescriptor creates a deep copy of a class descriptor
func (r *ClassRegistry) copyClassDescriptor(original *ClassDescriptor) *ClassDescriptor {
	classCopy := &ClassDescriptor{
		Name:       original.Name,
		Parent:     original.Parent,
		IsAbstract: original.IsAbstract,
		IsFinal:    original.IsFinal,
		Properties: make(map[string]*PropertyDescriptor),
		Methods:    make(map[string]*MethodDescriptor),
		Constants:  make(map[string]*ConstantDescriptor),
	}

	// Copy properties
	for k, v := range original.Properties {
		classCopy.Properties[k] = &PropertyDescriptor{
			Name:         v.Name,
			Type:         v.Type,
			Visibility:   v.Visibility,
			IsStatic:     v.IsStatic,
			DefaultValue: v.DefaultValue, // Value is immutable
		}
	}

	// Copy methods
	for k, v := range original.Methods {
		classCopy.Methods[k] = &MethodDescriptor{
			Name:           v.Name,
			Visibility:     v.Visibility,
			IsStatic:       v.IsStatic,
			IsAbstract:     v.IsAbstract,
			IsFinal:        v.IsFinal,
			Parameters:     make([]ParameterDescriptor, len(v.Parameters)),
			Implementation: v.Implementation, // Interface is immutable
			IsVariadic:     v.IsVariadic,
		}
		copy(classCopy.Methods[k].Parameters, v.Parameters)
	}

	// Copy constants
	for k, v := range original.Constants {
		classCopy.Constants[k] = &ConstantDescriptor{
			Name:       v.Name,
			Value:      v.Value,
			Visibility: v.Visibility,
			Type:       v.Type,
			IsFinal:    v.IsFinal,
			IsAbstract: v.IsAbstract,
		}
	}

	// Copy metadata
	if original.Metadata != nil {
		classCopy.Metadata = &ClassMetadata{
			IsBuiltin:     original.Metadata.IsBuiltin,
			ExtensionName: original.Metadata.ExtensionName,
			LoadOrder:     original.Metadata.LoadOrder,
		}
	}

	return classCopy
}

// resolveMethod resolves a method by traversing the inheritance hierarchy
func (r *ClassRegistry) resolveMethod(className, methodName string) (*MethodDescriptor, string, error) {
	currentClassName := className
	visited := make(map[string]bool) // Prevent circular inheritance

	for currentClassName != "" {
		if visited[currentClassName] {
			return nil, "", fmt.Errorf("circular inheritance detected for class %s", className)
		}
		visited[currentClassName] = true

		class, err := r.GetClass(currentClassName)
		if err != nil {
			return nil, "", err
		}

		if method, exists := class.Methods[methodName]; exists {
			return method, currentClassName, nil
		}

		// Move up the inheritance chain
		currentClassName = class.Parent
	}

	return nil, "", fmt.Errorf("method %s not found in class %s or its parent classes", methodName, className)
}

// ExecuteMethodCall executes a method on a class
func (r *ClassRegistry) ExecuteMethodCall(ctx ExecutionContext, className, methodName string, args []*values.Value) (*values.Value, error) {
	method, actualClassName, err := r.resolveMethod(className, methodName)
	if err != nil {
		return nil, err
	}

	if method.Implementation == nil {
		return nil, fmt.Errorf("method %s has no implementation in class %s", methodName, actualClassName)
	}

	// Direct method execution without type conversion
	return method.Implementation.Execute(ctx, args)
}

// Global registry instance
var GlobalRegistry *ClassRegistry

// Initialize initializes the global registry
func Initialize() {
	GlobalRegistry = NewClassRegistry()
}

// Convenience functions for global registry
func RegisterBuiltinClass(class *ClassDescriptor) error {
	if GlobalRegistry == nil {
		return fmt.Errorf("class registry not initialized")
	}

	// Mark as built-in
	if class.Metadata == nil {
		class.Metadata = &ClassMetadata{}
	}
	class.Metadata.IsBuiltin = true

	return GlobalRegistry.RegisterClass(class)
}

func GetClass(name string) (*ClassDescriptor, error) {
	if GlobalRegistry == nil {
		return nil, fmt.Errorf("class registry not initialized")
	}
	return GlobalRegistry.GetClass(name)
}

func HasClass(name string) bool {
	if GlobalRegistry == nil {
		return false
	}
	return GlobalRegistry.HasClass(name)
}

// CallBuiltinFunction is a placeholder - this should be implemented by runtime package
func CallBuiltinFunction(ctx ExecutionContext, name string, args []*values.Value) (*values.Value, error) {
	// This is a placeholder - actual implementation should be in runtime package
	return values.NewNull(), fmt.Errorf("function %s not implemented in unified system", name)
}
