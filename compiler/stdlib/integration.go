package stdlib

import (
	"fmt"

	"github.com/wudi/php-parser/compiler/values"
	"github.com/wudi/php-parser/compiler/vm"
)

// StdlibIntegration provides integration between the standard library and VM
type StdlibIntegration struct {
	stdlib           *StandardLibrary
	extensionManager *ExtensionManager
}

// NewStdlibIntegration creates a new stdlib integration
func NewStdlibIntegration() *StdlibIntegration {
	stdlib := NewStandardLibrary()
	extensionManager := NewExtensionManager(stdlib)
	
	// Load built-in extensions
	stdlib.LoadBuiltinExtensions()
	
	return &StdlibIntegration{
		stdlib:           stdlib,
		extensionManager: extensionManager,
	}
}

// InitializeExecutionContext initializes an execution context with stdlib support
func (si *StdlibIntegration) InitializeExecutionContext(ctx *vm.ExecutionContext) error {
	// Initialize constants
	for name, value := range si.stdlib.Constants {
		// Add constants to the context (would need constant pool integration)
		_ = name
		_ = value
	}
	
	// Initialize global variables
	if ctx.GlobalVars == nil {
		ctx.GlobalVars = make(map[string]*values.Value)
	}
	
	for name, value := range si.stdlib.Variables {
		ctx.GlobalVars[name] = value
	}
	
	// Initialize built-in functions
	if ctx.Functions == nil {
		ctx.Functions = make(map[string]*vm.Function)
	}
	
	for name, builtinFunc := range si.stdlib.Functions {
		// Convert builtin function to VM function
		vmFunc := si.createVMFunctionWrapper(name, builtinFunc)
		ctx.Functions[name] = vmFunc
	}
	
	// Initialize built-in classes
	if ctx.Classes == nil {
		ctx.Classes = make(map[string]*vm.Class)
	}
	
	for name, stdlibClass := range si.stdlib.Classes {
		// Convert stdlib class to VM class
		vmClass := si.convertClassToVM(stdlibClass)
		ctx.Classes[name] = vmClass
	}
	
	return nil
}

// createVMFunctionWrapper creates a VM function wrapper for builtin functions
func (si *StdlibIntegration) createVMFunctionWrapper(name string, builtinFunc BuiltinFunction) *vm.Function {
	return &vm.Function{
		Name:         name,
		Instructions: nil, // Built-in functions don't have bytecode
		Constants:    nil,
		Parameters:   si.convertParametersToVM(builtinFunc.Parameters),
		IsVariadic:   builtinFunc.IsVariadic,
		IsGenerator:  false,
	}
}

// convertParametersToVM converts stdlib parameters to VM parameters
func (si *StdlibIntegration) convertParametersToVM(params []Parameter) []vm.Parameter {
	vmParams := make([]vm.Parameter, len(params))
	for i, param := range params {
		vmParams[i] = vm.Parameter{
			Name:         param.Name,
			Type:         param.Type,
			IsReference:  param.IsReference,
			HasDefault:   param.HasDefault,
			DefaultValue: param.DefaultValue,
		}
	}
	return vmParams
}

// convertClassToVM converts stdlib class to VM class
func (si *StdlibIntegration) convertClassToVM(stdlibClass *Class) *vm.Class {
	vmClass := &vm.Class{
		Name:        stdlibClass.Name,
		ParentClass: stdlibClass.Parent,
		Properties:  make(map[string]*vm.Property),
		Methods:     make(map[string]*vm.Function),
		Constants:   stdlibClass.Constants,
		IsAbstract:  stdlibClass.IsAbstract,
		IsFinal:     stdlibClass.IsFinal,
	}
	
	// Convert properties
	for name, prop := range stdlibClass.Properties {
		vmClass.Properties[name] = &vm.Property{
			Name:         prop.Name,
			Type:         prop.Type,
			Visibility:   prop.Visibility,
			IsStatic:     prop.IsStatic,
			DefaultValue: prop.DefaultValue,
		}
	}
	
	// Convert methods
	for name, method := range stdlibClass.Methods {
		vmClass.Methods[name] = &vm.Function{
			Name:         method.Name,
			Instructions: nil, // Built-in methods don't have bytecode
			Constants:    nil,
			Parameters:   si.convertParametersToVM(method.Parameters),
			IsVariadic:   method.IsVariadic,
			IsGenerator:  false,
		}
	}
	
	return vmClass
}

// HandleBuiltinFunctionCall handles calls to built-in functions
func (si *StdlibIntegration) HandleBuiltinFunctionCall(ctx *vm.ExecutionContext, functionName string, args []*values.Value) (*values.Value, error) {
	builtinFunc, exists := si.stdlib.Functions[functionName]
	if !exists {
		return nil, fmt.Errorf("undefined function: %s", functionName)
	}
	
	// Validate argument count
	if len(args) < builtinFunc.MinArgs {
		return nil, fmt.Errorf("%s() expects at least %d parameters, %d given", functionName, builtinFunc.MinArgs, len(args))
	}
	
	if builtinFunc.MaxArgs != -1 && len(args) > builtinFunc.MaxArgs {
		return nil, fmt.Errorf("%s() expects at most %d parameters, %d given", functionName, builtinFunc.MaxArgs, len(args))
	}
	
	// Call the handler
	return builtinFunc.Handler(ctx, args)
}

// HandleBuiltinMethodCall handles calls to built-in class methods
func (si *StdlibIntegration) HandleBuiltinMethodCall(ctx *vm.ExecutionContext, className, methodName string, thisObject *values.Value, args []*values.Value) (*values.Value, error) {
	stdlibClass, exists := si.stdlib.Classes[className]
	if !exists {
		return nil, fmt.Errorf("undefined class: %s", className)
	}
	
	method, exists := stdlibClass.Methods[methodName]
	if !exists {
		return nil, fmt.Errorf("undefined method: %s::%s", className, methodName)
	}
	
	// Check visibility (simplified - would need proper visibility checking)
	if method.Visibility == "private" || method.Visibility == "protected" {
		// Would need proper access control checking here
	}
	
	// Call the method handler
	if method.Handler != nil {
		return method.Handler(ctx, args)
	}
	
	return values.NewNull(), nil
}

// IsBuiltinFunction checks if a function is a built-in function
func (si *StdlibIntegration) IsBuiltinFunction(functionName string) bool {
	_, exists := si.stdlib.Functions[functionName]
	return exists
}

// IsBuiltinClass checks if a class is a built-in class
func (si *StdlibIntegration) IsBuiltinClass(className string) bool {
	_, exists := si.stdlib.Classes[className]
	return exists
}

// GetBuiltinConstant retrieves a built-in constant value
func (si *StdlibIntegration) GetBuiltinConstant(constantName string) (*values.Value, bool) {
	value, exists := si.stdlib.Constants[constantName]
	return value, exists
}

// GetSuperGlobal retrieves a superglobal variable
func (si *StdlibIntegration) GetSuperGlobal(name string) *values.Value {
	return si.stdlib.GetSuperGlobal(name)
}

// SetSuperGlobal sets a superglobal variable
func (si *StdlibIntegration) SetSuperGlobal(name string, value *values.Value) {
	si.stdlib.SetSuperGlobal(name, value)
}

// RegisterExtension registers an external extension
func (si *StdlibIntegration) RegisterExtension(ext Extension) error {
	return si.extensionManager.RegisterExtension(ext)
}

// UnregisterExtension unregisters an external extension
func (si *StdlibIntegration) UnregisterExtension(name string) error {
	return si.extensionManager.UnregisterExtension(name)
}

// GetExtension retrieves a registered extension
func (si *StdlibIntegration) GetExtension(name string) (Extension, bool) {
	return si.extensionManager.GetExtension(name)
}

// IsExtensionLoaded checks if an extension is loaded
func (si *StdlibIntegration) IsExtensionLoaded(name string) bool {
	return si.extensionManager.IsExtensionLoaded(name)
}

// GetLoadedExtensions returns list of loaded extensions
func (si *StdlibIntegration) GetLoadedExtensions() []string {
	return si.extensionManager.GetRegisteredExtensions()
}

// GetFunctionInfo returns information about a built-in function
func (si *StdlibIntegration) GetFunctionInfo(functionName string) (*BuiltinFunction, bool) {
	fn, exists := si.stdlib.Functions[functionName]
	if !exists {
		return nil, false
	}
	return &fn, true
}

// GetClassInfo returns information about a built-in class
func (si *StdlibIntegration) GetClassInfo(className string) (*Class, bool) {
	class, exists := si.stdlib.Classes[className]
	if !exists {
		return nil, false
	}
	return class, true
}

// ValidateBuiltinFunctionCall validates a built-in function call
func (si *StdlibIntegration) ValidateBuiltinFunctionCall(functionName string, argCount int) error {
	builtinFunc, exists := si.stdlib.Functions[functionName]
	if !exists {
		return fmt.Errorf("undefined function: %s", functionName)
	}
	
	if argCount < builtinFunc.MinArgs {
		return fmt.Errorf("%s() expects at least %d parameters, %d given", functionName, builtinFunc.MinArgs, argCount)
	}
	
	if builtinFunc.MaxArgs != -1 && argCount > builtinFunc.MaxArgs {
		return fmt.Errorf("%s() expects at most %d parameters, %d given", functionName, builtinFunc.MaxArgs, argCount)
	}
	
	return nil
}

// GetAllBuiltinFunctions returns a map of all built-in functions
func (si *StdlibIntegration) GetAllBuiltinFunctions() map[string]BuiltinFunction {
	return si.stdlib.Functions
}

// GetAllBuiltinClasses returns a map of all built-in classes
func (si *StdlibIntegration) GetAllBuiltinClasses() map[string]*Class {
	return si.stdlib.Classes
}

// GetAllBuiltinConstants returns a map of all built-in constants
func (si *StdlibIntegration) GetAllBuiltinConstants() map[string]*values.Value {
	return si.stdlib.Constants
}

// Global stdlib integration instance
var GlobalStdlibIntegration *StdlibIntegration

// InitializeGlobalStdlib initializes the global stdlib integration
func InitializeGlobalStdlib() {
	GlobalStdlibIntegration = NewStdlibIntegration()
}

// Helper functions for easy access

// CallBuiltinFunction calls a built-in function
func CallBuiltinFunction(ctx *vm.ExecutionContext, functionName string, args []*values.Value) (*values.Value, error) {
	if GlobalStdlibIntegration == nil {
		return nil, fmt.Errorf("standard library not initialized")
	}
	return GlobalStdlibIntegration.HandleBuiltinFunctionCall(ctx, functionName, args)
}

// CallBuiltinMethod calls a built-in class method
func CallBuiltinMethod(ctx *vm.ExecutionContext, className, methodName string, thisObject *values.Value, args []*values.Value) (*values.Value, error) {
	if GlobalStdlibIntegration == nil {
		return nil, fmt.Errorf("standard library not initialized")
	}
	return GlobalStdlibIntegration.HandleBuiltinMethodCall(ctx, className, methodName, thisObject, args)
}

// GetBuiltinConstantValue retrieves a built-in constant
func GetBuiltinConstantValue(constantName string) (*values.Value, bool) {
	if GlobalStdlibIntegration == nil {
		return nil, false
	}
	return GlobalStdlibIntegration.GetBuiltinConstant(constantName)
}

// IsBuiltinFunctionGlobal checks if a function is built-in (global helper)
func IsBuiltinFunctionGlobal(functionName string) bool {
	if GlobalStdlibIntegration == nil {
		return false
	}
	return GlobalStdlibIntegration.IsBuiltinFunction(functionName)
}