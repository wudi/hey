package runtime

import (
	"fmt"
	"sync"

	"github.com/wudi/php-parser/compiler/values"
)

// RuntimeRegistry is the unified registration system for all runtime entities
type RuntimeRegistry struct {
	mu sync.RWMutex
	
	// Core registry maps
	constants map[string]*values.Value
	variables map[string]*values.Value
	functions map[string]*FunctionDescriptor
	classes   map[string]*ClassDescriptor
	
	// Extension management
	extensions map[string]*ExtensionDescriptor
	
	// Built-in flags to prevent overrides
	builtinConstants map[string]bool
	builtinVariables map[string]bool
	builtinFunctions map[string]bool
	builtinClasses   map[string]bool
}

// FunctionDescriptor describes a runtime function
type FunctionDescriptor struct {
	Name         string
	Handler      FunctionHandler
	Parameters   []ParameterDescriptor
	IsVariadic   bool
	MinArgs      int
	MaxArgs      int
	IsBuiltin    bool
	ExtensionName string
}

// ParameterDescriptor describes a function parameter
type ParameterDescriptor struct {
	Name         string
	Type         string
	IsReference  bool
	HasDefault   bool
	DefaultValue *values.Value
}

// ClassDescriptor describes a runtime class
type ClassDescriptor struct {
	Name         string
	Parent       string
	Properties   map[string]*PropertyDescriptor
	Methods      map[string]*MethodDescriptor
	Constants    map[string]*values.Value
	IsAbstract   bool
	IsFinal      bool
	IsBuiltin    bool
	ExtensionName string
}

// PropertyDescriptor describes a class property
type PropertyDescriptor struct {
	Name         string
	Type         string
	Visibility   string // public, private, protected
	IsStatic     bool
	DefaultValue *values.Value
}

// MethodDescriptor describes a class method
type MethodDescriptor struct {
	Name         string
	Visibility   string
	IsStatic     bool
	IsAbstract   bool
	IsFinal      bool
	Parameters   []ParameterDescriptor
	Handler      FunctionHandler
	IsVariadic   bool
}

// ExtensionDescriptor describes a registered extension
type ExtensionDescriptor struct {
	Name         string
	Version      string
	Description  string
	LoadOrder    int
	Dependencies []string
}

// ExecutionContext interface for runtime function handlers 
type ExecutionContext interface {
	// Add methods as needed for function handlers
}

// FunctionHandler is the type for runtime function handlers
type FunctionHandler func(ctx ExecutionContext, args []*values.Value) (*values.Value, error)

// NewRuntimeRegistry creates a new unified runtime registry
func NewRuntimeRegistry() *RuntimeRegistry {
	return &RuntimeRegistry{
		constants: make(map[string]*values.Value),
		variables: make(map[string]*values.Value),
		functions: make(map[string]*FunctionDescriptor),
		classes:   make(map[string]*ClassDescriptor),
		extensions: make(map[string]*ExtensionDescriptor),
		builtinConstants: make(map[string]bool),
		builtinVariables: make(map[string]bool),
		builtinFunctions: make(map[string]bool),
		builtinClasses:   make(map[string]bool),
	}
}

// RegisterConstant registers a runtime constant
func (r *RuntimeRegistry) RegisterConstant(name string, value *values.Value, isBuiltin bool, extensionName string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	// Check for conflicts with built-ins
	if r.builtinConstants[name] && !isBuiltin {
		return fmt.Errorf("cannot override built-in constant: %s", name)
	}
	
	// Check for existing registration
	if _, exists := r.constants[name]; exists {
		return fmt.Errorf("constant already registered: %s", name)
	}
	
	r.constants[name] = value
	if isBuiltin {
		r.builtinConstants[name] = true
	}
	
	return nil
}

// RegisterVariable registers a runtime variable
func (r *RuntimeRegistry) RegisterVariable(name string, value *values.Value, isBuiltin bool, extensionName string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	// Check for conflicts with built-ins
	if r.builtinVariables[name] && !isBuiltin {
		return fmt.Errorf("cannot override built-in variable: %s", name)
	}
	
	// Check for existing registration
	if _, exists := r.variables[name]; exists && r.builtinVariables[name] {
		return fmt.Errorf("variable already registered: %s", name)
	}
	
	r.variables[name] = value
	if isBuiltin {
		r.builtinVariables[name] = true
	}
	
	return nil
}

// RegisterFunction registers a runtime function
func (r *RuntimeRegistry) RegisterFunction(descriptor *FunctionDescriptor) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	name := descriptor.Name
	
	// Check for conflicts with built-ins
	if r.builtinFunctions[name] && !descriptor.IsBuiltin {
		return fmt.Errorf("cannot override built-in function: %s", name)
	}
	
	// Check for existing registration
	if _, exists := r.functions[name]; exists {
		return fmt.Errorf("function already registered: %s", name)
	}
	
	r.functions[name] = descriptor
	if descriptor.IsBuiltin {
		r.builtinFunctions[name] = true
	}
	
	return nil
}

// RegisterClass registers a runtime class
func (r *RuntimeRegistry) RegisterClass(descriptor *ClassDescriptor) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	name := descriptor.Name
	
	// Check for conflicts with built-ins
	if r.builtinClasses[name] && !descriptor.IsBuiltin {
		return fmt.Errorf("cannot override built-in class: %s", name)
	}
	
	// Check for existing registration
	if _, exists := r.classes[name]; exists {
		return fmt.Errorf("class already registered: %s", name)
	}
	
	r.classes[name] = descriptor
	if descriptor.IsBuiltin {
		r.builtinClasses[name] = true
	}
	
	return nil
}

// RegisterExtension registers an extension
func (r *RuntimeRegistry) RegisterExtension(descriptor *ExtensionDescriptor) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	name := descriptor.Name
	
	// Check for existing registration
	if _, exists := r.extensions[name]; exists {
		return fmt.Errorf("extension already registered: %s", name)
	}
	
	r.extensions[name] = descriptor
	return nil
}

// Lookup methods
func (r *RuntimeRegistry) GetConstant(name string) (*values.Value, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	value, exists := r.constants[name]
	return value, exists
}

func (r *RuntimeRegistry) GetVariable(name string) (*values.Value, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	value, exists := r.variables[name]
	return value, exists
}

func (r *RuntimeRegistry) GetFunction(name string) (*FunctionDescriptor, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	descriptor, exists := r.functions[name]
	return descriptor, exists
}

func (r *RuntimeRegistry) GetClass(name string) (*ClassDescriptor, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	descriptor, exists := r.classes[name]
	return descriptor, exists
}

func (r *RuntimeRegistry) GetExtension(name string) (*ExtensionDescriptor, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	descriptor, exists := r.extensions[name]
	return descriptor, exists
}

// Query methods
func (r *RuntimeRegistry) IsBuiltinConstant(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	return r.builtinConstants[name]
}

func (r *RuntimeRegistry) IsBuiltinFunction(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	return r.builtinFunctions[name]
}

func (r *RuntimeRegistry) IsBuiltinClass(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	return r.builtinClasses[name]
}

func (r *RuntimeRegistry) HasFunction(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	_, exists := r.functions[name]
	return exists
}

func (r *RuntimeRegistry) HasClass(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	_, exists := r.classes[name]
	return exists
}

// Bulk operations
func (r *RuntimeRegistry) GetAllConstants() map[string]*values.Value {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	result := make(map[string]*values.Value, len(r.constants))
	for k, v := range r.constants {
		result[k] = v
	}
	return result
}

func (r *RuntimeRegistry) GetAllVariables() map[string]*values.Value {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	result := make(map[string]*values.Value, len(r.variables))
	for k, v := range r.variables {
		result[k] = v
	}
	return result
}

func (r *RuntimeRegistry) GetAllFunctions() map[string]*FunctionDescriptor {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	result := make(map[string]*FunctionDescriptor, len(r.functions))
	for k, v := range r.functions {
		// Create a copy to avoid external mutation
		desc := *v
		result[k] = &desc
	}
	return result
}

func (r *RuntimeRegistry) GetAllClasses() map[string]*ClassDescriptor {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	result := make(map[string]*ClassDescriptor, len(r.classes))
	for k, v := range r.classes {
		// Create a copy to avoid external mutation
		desc := *v
		result[k] = &desc
	}
	return result
}

// Extension listing
func (r *RuntimeRegistry) GetRegisteredExtensions() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var names []string
	for name := range r.extensions {
		names = append(names, name)
	}
	return names
}

// Validation
func (r *RuntimeRegistry) ValidateFunctionCall(name string, argCount int) error {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	descriptor, exists := r.functions[name]
	if !exists {
		return fmt.Errorf("undefined function: %s", name)
	}
	
	if argCount < descriptor.MinArgs {
		return fmt.Errorf("%s() expects at least %d parameters, %d given", name, descriptor.MinArgs, argCount)
	}
	
	if descriptor.MaxArgs != -1 && argCount > descriptor.MaxArgs {
		return fmt.Errorf("%s() expects at most %d parameters, %d given", name, descriptor.MaxArgs, argCount)
	}
	
	return nil
}

// Function execution
func (r *RuntimeRegistry) CallFunction(ctx ExecutionContext, name string, args []*values.Value) (*values.Value, error) {
	r.mu.RLock()
	descriptor, exists := r.functions[name]
	r.mu.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("undefined function: %s", name)
	}
	
	// Validate arguments
	if err := r.ValidateFunctionCall(name, len(args)); err != nil {
		return nil, err
	}
	
	// Call the handler
	return descriptor.Handler(ctx, args)
}

// Global registry instance
var GlobalRegistry *RuntimeRegistry

// Initialize initializes the global registry
func Initialize() {
	GlobalRegistry = NewRuntimeRegistry()
}

// Convenience functions for global registry
func RegisterBuiltinConstant(name string, value *values.Value) error {
	if GlobalRegistry == nil {
		return fmt.Errorf("runtime registry not initialized")
	}
	return GlobalRegistry.RegisterConstant(name, value, true, "")
}

func RegisterBuiltinFunction(descriptor *FunctionDescriptor) error {
	if GlobalRegistry == nil {
		return fmt.Errorf("runtime registry not initialized")
	}
	descriptor.IsBuiltin = true
	return GlobalRegistry.RegisterFunction(descriptor)
}

func CallBuiltinFunction(ctx ExecutionContext, name string, args []*values.Value) (*values.Value, error) {
	if GlobalRegistry == nil {
		return nil, fmt.Errorf("runtime registry not initialized")
	}
	return GlobalRegistry.CallFunction(ctx, name, args)
}

func HasBuiltinFunction(name string) bool {
	if GlobalRegistry == nil {
		return false
	}
	return GlobalRegistry.HasFunction(name)
}