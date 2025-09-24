package registry

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/wudi/hey/opcodes"
	"github.com/wudi/hey/values"
)

// BuiltinImplementation defines the function signature for builtin functions
// implemented in Go and callable from the VM.
type BuiltinImplementation func(ctx BuiltinCallContext, args []*values.Value) (*values.Value, error)

// BuiltinCallContext exposes the minimal VM services that builtin implementations
// need without creating a dependency cycle back to the vm package.
type BuiltinCallContext interface {
	// WriteOutput should render the provided value to the active output stream.
	WriteOutput(val *values.Value) error
	// GetGlobal fetches a global variable by name if it exists.
	GetGlobal(name string) (*values.Value, bool)
	// SetGlobal updates or creates a global variable by name.
	SetGlobal(name string, val *values.Value)
	// SymbolRegistry returns the unified registry to allow builtins to inspect
	// other symbols (functions, classes, etc.).
	SymbolRegistry() *Registry
	// LookupUserFunction returns a user-defined function registered inside the
	// active execution context, if available.
	LookupUserFunction(name string) (*Function, bool)
	// LookupUserClass returns a user-defined class registered inside the active
	// execution context, if available.
	LookupUserClass(name string) (*Class, bool)
	// Halt stops execution with the given exit code and optional message.
	Halt(exitCode int, message string) error
	// GetExecutionContext returns the execution context for timeout management
	GetExecutionContext() ExecutionContextInterface
	// GetOutputBufferStack returns the output buffer stack for output control
	GetOutputBufferStack() OutputBufferStackInterface
}

// ExecutionContextInterface provides minimal interface for timeout management
type ExecutionContextInterface interface {
	SetTimeLimit(seconds int) bool
}

// OutputBufferStackInterface provides minimal interface for output buffer control
type OutputBufferStackInterface interface {
	Start(handler string, chunkSize int, flags int) bool
	GetContents() string
	GetLength() int
	GetLevel() int
	Clean() bool
	EndClean() bool
	Flush() bool
	EndFlush() bool
	GetClean() (string, bool)
	GetFlush() (string, bool)
	GetStatus() *values.Value
	GetStatusFull() *values.Value
	ListHandlers() []string
	SetImplicitFlush(on bool)
	FlushSystem()
}

// Function describes a PHP function that can either be user-defined (bytecode)
// or builtin (Go implementation).
type Function struct {
	Name         string
	Parameters   []*Parameter
	ReturnType   string
	Instructions []*opcodes.Instruction
	Constants    []*values.Value
	IsVariadic        bool
	IsGenerator       bool
	IsAnonymous       bool
	IsBuiltin         bool
	IsAbstract        bool
	ReturnsByReference bool
	Builtin           BuiltinImplementation
	Handler      func(interface{}, []*values.Value) (*values.Value, error)
	MinArgs      int
	MaxArgs      int
	Attributes   []*Attribute
	// Variable slot mapping for proper local variable allocation in goroutines
	VariableSlots map[string]uint32 // variable name -> slot number
	MaxLocalSlot  uint32            // highest slot number used + 1
}

// Clone creates a shallow copy of the function metadata. Instructions and
// constants are re-used, mirroring PHP's copy-on-write semantics for op arrays.
func (f *Function) Clone() *Function {
	if f == nil {
		return nil
	}
	clone := *f
	return &clone
}

// Parameter captures metadata about a compiled parameter.
type Parameter struct {
	Name         string
	Type         string
	IsReference  bool
	HasDefault   bool
	DefaultValue *values.Value
	Attributes   []*Attribute
}

// Attribute represents a compiled PHP attribute.
type Attribute struct {
	Name      string
	Arguments []*values.Value
}

// Class models a compiled PHP class definition used by the compiler and VM.
type Class struct {
	Name       string
	Parent     string
	Interfaces []string
	Traits     []string
	Properties map[string]*Property
	Methods    map[string]*Function
	Constants  map[string]*ClassConstant
	IsAbstract bool
	IsFinal    bool
	Attributes []*Attribute
}

// Property represents a class property.
type Property struct {
	Name         string
	Visibility   string
	IsStatic     bool
	IsReadonly   bool
	Type         string
	DefaultValue *values.Value
	DocComment   string
	Attributes   []*Attribute
}

// ClassConstant represents a class constant.
type ClassConstant struct {
	Name       string
	Value      *values.Value
	Visibility string
	IsFinal    bool
	Type       string
	IsAbstract bool
}

// Interface models an interface declaration.
type Interface struct {
	Name    string
	Methods map[string]*InterfaceMethod
	Extends []string
}

// InterfaceMethod represents a method requirement within an interface.
type InterfaceMethod struct {
	Name       string
	Visibility string
	Parameters []*Parameter
	ReturnType string
}

// Trait models a PHP trait definition.
type Trait struct {
	Name       string
	Properties map[string]*Property
	Methods    map[string]*Function
}

// Constant represents a global constant.
type Constant struct {
	Name  string
	Value *values.Value
}

// The descriptor layer is used by the unified symbol system to expose
// lightweight metadata to the runtime and tools without binding to the VM
// execution data structures.

type ParameterDescriptor struct {
	Name         string
	Type         string
	IsReference  bool
	HasDefault   bool
	DefaultValue *values.Value
}

type PropertyDescriptor struct {
	Name         string
	Visibility   string
	IsStatic     bool
	Type         string
	DefaultValue *values.Value
	IsReadonly   bool
}

type ConstantDescriptor struct {
	Name       string
	Visibility string
	Value      *values.Value
	IsFinal    bool
}

type MethodImplementation interface {
	ImplementationKind() string
}

type ParameterInfo struct {
	Name         string
	HasDefault   bool
	DefaultValue *values.Value
	IsVariadic   bool
}

type BytecodeMethodImpl struct {
	Instructions []*opcodes.Instruction
	Constants    []*values.Value
	LocalVars    int
	Parameters   []*ParameterInfo
}

func (b *BytecodeMethodImpl) ImplementationKind() string { return "bytecode" }

type MethodDescriptor struct {
	Name           string
	Visibility     string
	IsStatic       bool
	IsAbstract     bool
	IsFinal        bool
	IsVariadic     bool
	Parameters     []*ParameterDescriptor
	Implementation MethodImplementation
}

type ClassDescriptor struct {
	Name       string
	Parent     string
	Interfaces []string
	Traits     []string
	IsAbstract bool
	IsFinal    bool
	Properties map[string]*PropertyDescriptor
	Methods    map[string]*MethodDescriptor
	Constants  map[string]*ConstantDescriptor
}

// Registry is a threadsafe container for all globally registered symbols.
type Registry struct {
	mu         sync.RWMutex
	functions  map[string]*Function
	classes    map[string]*ClassDescriptor
	constants  map[string]*ConstantDescriptor
	interfaces map[string]*Interface
	traits     map[string]*Trait
}

var (
	initOnce       sync.Once
	GlobalRegistry *Registry
)

// Initialize ensures the global registry instance is created exactly once.
func Initialize() {
	initOnce.Do(func() {
		GlobalRegistry = &Registry{
			functions:  make(map[string]*Function),
			classes:    make(map[string]*ClassDescriptor),
			constants:  make(map[string]*ConstantDescriptor),
			interfaces: make(map[string]*Interface),
			traits:     make(map[string]*Trait),
		}
	})
}

func keyFor(name string) string {
	return strings.ToLower(name)
}

// RegisterFunction registers a function (builtin or user) with the registry.
// The last registration wins which mirrors PHP's function redeclaration rules
// for user code executed at runtime.
func (r *Registry) RegisterFunction(fn *Function) error {
	if fn == nil {
		return errors.New("cannot register nil function")
	}
	if fn.Name == "" {
		return errors.New("cannot register function with empty name")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.functions[keyFor(fn.Name)] = fn
	return nil
}

// GetAllFunctions returns a shallow copy of all registered functions.
func (r *Registry) GetAllFunctions() map[string]*Function {
	if r == nil {
		return nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make(map[string]*Function, len(r.functions))
	for name, fn := range r.functions {
		out[name] = fn
	}
	return out
}

// GetFunction fetches a function by case-insensitive name.
func (r *Registry) GetFunction(name string) (*Function, bool) {
	if r == nil {
		return nil, false
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	fn, ok := r.functions[keyFor(name)]
	if !ok {
		return nil, false
	}
	return fn, true
}

// RegisterClass stores or replaces a class descriptor.
func (r *Registry) RegisterClass(class *ClassDescriptor) error {
	if class == nil {
		return errors.New("cannot register nil class descriptor")
	}
	if class.Name == "" {
		return errors.New("cannot register class with empty name")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.classes[keyFor(class.Name)] = class
	return nil
}

// GetAllClasses returns a shallow copy of all registered class descriptors.
func (r *Registry) GetAllClasses() map[string]*ClassDescriptor {
	if r == nil {
		return nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make(map[string]*ClassDescriptor, len(r.classes))
	for name, class := range r.classes {
		out[name] = class
	}
	return out
}

// GetClass retrieves a class descriptor when present.
func (r *Registry) GetClass(name string) (*ClassDescriptor, error) {
	if r == nil {
		return nil, errors.New("registry not initialized")
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	class, ok := r.classes[keyFor(name)]
	if !ok {
		return nil, fmt.Errorf("class %s not registered", name)
	}
	return class, nil
}

// RegisterConstant stores a global constant descriptor.
func (r *Registry) RegisterConstant(constant *ConstantDescriptor) error {
	if constant == nil {
		return errors.New("cannot register nil constant descriptor")
	}
	if constant.Name == "" {
		return errors.New("cannot register constant with empty name")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.constants[keyFor(constant.Name)] = constant
	return nil
}

// GetConstant retrieves a constant descriptor if available.
func (r *Registry) GetConstant(name string) (*ConstantDescriptor, bool) {
	if r == nil {
		return nil, false
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.constants[keyFor(name)]
	return c, ok
}

// RegisterInterface stores an interface definition.
func (r *Registry) RegisterInterface(iface *Interface) error {
	if iface == nil {
		return errors.New("cannot register nil interface")
	}
	if iface.Name == "" {
		return errors.New("cannot register interface with empty name")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.interfaces[keyFor(iface.Name)] = iface
	return nil
}

// GetInterface fetches an interface definition.
func (r *Registry) GetInterface(name string) (*Interface, bool) {
	if r == nil {
		return nil, false
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	iface, ok := r.interfaces[keyFor(name)]
	return iface, ok
}

// RegisterTrait stores a trait definition.
func (r *Registry) RegisterTrait(trait *Trait) error {
	if trait == nil {
		return errors.New("cannot register nil trait")
	}
	if trait.Name == "" {
		return errors.New("cannot register trait with empty name")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.traits[keyFor(trait.Name)] = trait
	return nil
}

// GetTrait fetches a trait definition.
func (r *Registry) GetTrait(name string) (*Trait, bool) {
	if r == nil {
		return nil, false
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	trait, ok := r.traits[keyFor(name)]
	return trait, ok
}

// IsInstanceOf checks if className is an instance of typeName
// This includes checking for exact match, parent classes, and interfaces
func (r *Registry) IsInstanceOf(className, typeName string) bool {
	if r == nil {
		return false
	}

	// Normalize names to lowercase for case-insensitive comparison
	className = keyFor(className)
	typeName = keyFor(typeName)

	// Special case for built-in types
	if typeName == "throwable" {
		// All exceptions implement Throwable
		return r.isExceptionType(className)
	}

	// Check for exact match
	if className == typeName {
		return true
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	// Check class hierarchy
	class, ok := r.classes[className]
	if !ok {
		return false
	}

	// Check parent classes recursively
	if class.Parent != "" {
		if r.IsInstanceOf(class.Parent, typeName) {
			return true
		}
	}

	// Check implemented interfaces
	for _, iface := range class.Interfaces {
		if keyFor(iface) == typeName {
			return true
		}
		// Check interface inheritance
		if r.isInterfaceExtends(iface, typeName) {
			return true
		}
	}
	return false
}

// isExceptionType checks if a class is an exception type
func (r *Registry) isExceptionType(className string) bool {
	// Built-in exception types
	exceptionTypes := []string{"exception", "error", "errorexception", "typeerror", "parseerror", "arithmeticerror"}
	for _, exType := range exceptionTypes {
		if className == exType {
			return true
		}
	}

	// Check if it extends Exception or Error
	r.mu.RLock()
	defer r.mu.RUnlock()

	class, ok := r.classes[className]
	if ok {
		parent := keyFor(class.Parent)
		if parent == "exception" || parent == "error" {
			return true
		}
		// Recursively check parent
		if parent != "" {
			return r.isExceptionType(parent)
		}
	}

	return false
}

// isInterfaceExtends checks if an interface extends another interface
func (r *Registry) isInterfaceExtends(ifaceName, parentName string) bool {
	iface, ok := r.interfaces[keyFor(ifaceName)]
	if !ok {
		return false
	}

	for _, extends := range iface.Extends {
		if keyFor(extends) == parentName {
			return true
		}
		// Recursive check
		if r.isInterfaceExtends(extends, parentName) {
			return true
		}
	}

	return false
}
