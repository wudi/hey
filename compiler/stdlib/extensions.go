package stdlib

import (
	"fmt"

	"github.com/wudi/php-parser/compiler/values"
	"github.com/wudi/php-parser/compiler/vm"
)

// ExtensionManager manages external PHP extensions
type ExtensionManager struct {
	extensions map[string]Extension
	stdlib     *StandardLibrary
}

// NewExtensionManager creates a new extension manager
func NewExtensionManager(stdlib *StandardLibrary) *ExtensionManager {
	return &ExtensionManager{
		extensions: make(map[string]Extension),
		stdlib:     stdlib,
	}
}

// RegisterExtension registers an extension with the standard library
func (em *ExtensionManager) RegisterExtension(ext Extension) error {
	name := ext.GetName()

	// Check if extension is already registered
	if _, exists := em.extensions[name]; exists {
		return fmt.Errorf("extension %s is already registered", name)
	}

	// Validate extension
	if err := em.validateExtension(ext); err != nil {
		return fmt.Errorf("extension validation failed: %v", err)
	}

	// Register with stdlib
	if err := em.stdlib.RegisterExtension(ext); err != nil {
		return fmt.Errorf("failed to register extension: %v", err)
	}

	// Store extension reference
	em.extensions[name] = ext

	return nil
}

// UnregisterExtension removes an extension from the standard library
func (em *ExtensionManager) UnregisterExtension(name string) error {
	ext, exists := em.extensions[name]
	if !exists {
		return fmt.Errorf("extension %s is not registered", name)
	}

	// Remove extension constants
	for constName := range ext.GetConstants() {
		delete(em.stdlib.Constants, constName)
	}

	// Remove extension variables
	for varName := range ext.GetVariables() {
		delete(em.stdlib.Variables, varName)
	}

	// Remove extension functions
	for funcName := range ext.GetFunctions() {
		delete(em.stdlib.Functions, funcName)
	}

	// Remove extension classes
	for className := range ext.GetClasses() {
		delete(em.stdlib.Classes, className)
	}

	// Remove from manager
	delete(em.extensions, name)

	return nil
}

// GetExtension returns a registered extension by name
func (em *ExtensionManager) GetExtension(name string) (Extension, bool) {
	ext, exists := em.extensions[name]
	return ext, exists
}

// GetRegisteredExtensions returns a list of all registered extension names
func (em *ExtensionManager) GetRegisteredExtensions() []string {
	var names []string
	for name := range em.extensions {
		names = append(names, name)
	}
	return names
}

// IsExtensionLoaded checks if an extension is loaded
func (em *ExtensionManager) IsExtensionLoaded(name string) bool {
	_, exists := em.extensions[name]
	return exists
}

// validateExtension validates an extension before registration
func (em *ExtensionManager) validateExtension(ext Extension) error {
	// Validate extension name
	name := ext.GetName()
	if name == "" {
		return fmt.Errorf("extension name cannot be empty")
	}

	// Validate extension version
	version := ext.GetVersion()
	if version == "" {
		return fmt.Errorf("extension version cannot be empty")
	}

	// Check for naming conflicts
	constants := ext.GetConstants()
	for constName := range constants {
		if _, exists := em.stdlib.Constants[constName]; exists {
			return fmt.Errorf("constant %s already exists", constName)
		}
	}

	functions := ext.GetFunctions()
	for funcName := range functions {
		if _, exists := em.stdlib.Functions[funcName]; exists {
			return fmt.Errorf("function %s already exists", funcName)
		}
	}

	classes := ext.GetClasses()
	for className := range classes {
		if _, exists := em.stdlib.Classes[className]; exists {
			return fmt.Errorf("class %s already exists", className)
		}
	}

	return nil
}

// BaseExtension provides a base implementation for extensions
type BaseExtension struct {
	name      string
	version   string
	constants map[string]*values.Value
	variables map[string]*values.Value
	functions map[string]BuiltinFunction
	classes   map[string]*Class
}

// NewBaseExtension creates a new base extension
func NewBaseExtension(name, version string) *BaseExtension {
	return &BaseExtension{
		name:      name,
		version:   version,
		constants: make(map[string]*values.Value),
		variables: make(map[string]*values.Value),
		functions: make(map[string]BuiltinFunction),
		classes:   make(map[string]*Class),
	}
}

// GetName returns the extension name
func (be *BaseExtension) GetName() string {
	return be.name
}

// GetVersion returns the extension version
func (be *BaseExtension) GetVersion() string {
	return be.version
}

// GetConstants returns extension constants
func (be *BaseExtension) GetConstants() map[string]*values.Value {
	return be.constants
}

// GetVariables returns extension variables
func (be *BaseExtension) GetVariables() map[string]*values.Value {
	return be.variables
}

// GetFunctions returns extension functions
func (be *BaseExtension) GetFunctions() map[string]BuiltinFunction {
	return be.functions
}

// GetClasses returns extension classes
func (be *BaseExtension) GetClasses() map[string]*Class {
	return be.classes
}

// AddConstant adds a constant to the extension
func (be *BaseExtension) AddConstant(name string, value *values.Value) {
	be.constants[name] = value
}

// AddVariable adds a variable to the extension
func (be *BaseExtension) AddVariable(name string, value *values.Value) {
	be.variables[name] = value
}

// AddFunction adds a function to the extension
func (be *BaseExtension) AddFunction(name string, handler FunctionHandler, params []Parameter, isVariadic bool, minArgs, maxArgs int) {
	be.functions[name] = BuiltinFunction{
		Name:       name,
		Handler:    handler,
		Parameters: params,
		IsVariadic: isVariadic,
		MinArgs:    minArgs,
		MaxArgs:    maxArgs,
	}
}

// AddClass adds a class to the extension
func (be *BaseExtension) AddClass(name string, class *Class) {
	be.classes[name] = class
}

// Example extension implementations

// MathExtension provides additional math functions
type MathExtension struct {
	*BaseExtension
}

// NewMathExtension creates a new math extension
func NewMathExtension() *MathExtension {
	ext := &MathExtension{
		BaseExtension: NewBaseExtension("math", "1.0.0"),
	}

	ext.initMathExtension()
	return ext
}

func (me *MathExtension) initMathExtension() {
	// Add math constants
	me.AddConstant("MATH_PI", values.NewFloat(3.14159265358979323846))
	me.AddConstant("MATH_E", values.NewFloat(2.71828182845904523536))

	// Add math functions
	me.AddFunction("deg2rad", deg2radHandler, []Parameter{
		{Name: "number", Type: "float", IsReference: false, HasDefault: false},
	}, false, 1, 1)

	me.AddFunction("rad2deg", rad2degHandler, []Parameter{
		{Name: "number", Type: "float", IsReference: false, HasDefault: false},
	}, false, 1, 1)

	me.AddFunction("hypot", hypotHandler, []Parameter{
		{Name: "x", Type: "float", IsReference: false, HasDefault: false},
		{Name: "y", Type: "float", IsReference: false, HasDefault: false},
	}, false, 2, 2)
}

// JSON extension provides JSON functions
type JsonExtension struct {
	*BaseExtension
}

// NewJsonExtension creates a new JSON extension
func NewJsonExtension() *JsonExtension {
	ext := &JsonExtension{
		BaseExtension: NewBaseExtension("json", "1.0.0"),
	}

	ext.initJsonExtension()
	return ext
}

func (je *JsonExtension) initJsonExtension() {
	// Add JSON constants
	je.AddConstant("JSON_HEX_TAG", values.NewInt(1))
	je.AddConstant("JSON_HEX_AMP", values.NewInt(2))
	je.AddConstant("JSON_HEX_APOS", values.NewInt(4))
	je.AddConstant("JSON_HEX_QUOT", values.NewInt(8))
	je.AddConstant("JSON_FORCE_OBJECT", values.NewInt(16))
	je.AddConstant("JSON_NUMERIC_CHECK", values.NewInt(32))
	je.AddConstant("JSON_UNESCAPED_SLASHES", values.NewInt(64))
	je.AddConstant("JSON_PRETTY_PRINT", values.NewInt(128))
	je.AddConstant("JSON_UNESCAPED_UNICODE", values.NewInt(256))

	// Add JSON functions
	je.AddFunction("json_encode", jsonEncodeHandler, []Parameter{
		{Name: "value", Type: "mixed", IsReference: false, HasDefault: false},
		{Name: "flags", Type: "int", IsReference: false, HasDefault: true, DefaultValue: values.NewInt(0)},
		{Name: "depth", Type: "int", IsReference: false, HasDefault: true, DefaultValue: values.NewInt(512)},
	}, false, 1, 3)

	je.AddFunction("json_decode", jsonDecodeHandler, []Parameter{
		{Name: "json", Type: "string", IsReference: false, HasDefault: false},
		{Name: "associative", Type: "bool", IsReference: false, HasDefault: true, DefaultValue: values.NewBool(false)},
		{Name: "depth", Type: "int", IsReference: false, HasDefault: true, DefaultValue: values.NewInt(512)},
		{Name: "flags", Type: "int", IsReference: false, HasDefault: true, DefaultValue: values.NewInt(0)},
	}, false, 1, 4)
}

// Extension function handlers

func deg2radHandler(ctx *vm.ExecutionContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("deg2rad() expects exactly 1 parameter, %d given", len(args))
	}

	degrees := args[0].ToFloat()
	radians := degrees * (3.14159265358979323846 / 180.0)
	return values.NewFloat(radians), nil
}

func rad2degHandler(ctx *vm.ExecutionContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("rad2deg() expects exactly 1 parameter, %d given", len(args))
	}

	radians := args[0].ToFloat()
	degrees := radians * (180.0 / 3.14159265358979323846)
	return values.NewFloat(degrees), nil
}

func hypotHandler(ctx *vm.ExecutionContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("hypot() expects exactly 2 parameters, %d given", len(args))
	}

	x := args[0].ToFloat()
	y := args[1].ToFloat()
	result := (x*x + y*y) // Simplified sqrt calculation
	return values.NewFloat(result), nil
}

func jsonEncodeHandler(ctx *vm.ExecutionContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("json_encode() expects at least 1 parameter, %d given", len(args))
	}

	// Simplified JSON encoding
	value := args[0]
	switch value.Type {
	case values.TypeString:
		return values.NewString(fmt.Sprintf(`"%s"`, value.ToString())), nil
	case values.TypeInt:
		return values.NewString(fmt.Sprintf("%d", value.ToInt())), nil
	case values.TypeFloat:
		return values.NewString(fmt.Sprintf("%g", value.ToFloat())), nil
	case values.TypeBool:
		if value.ToBool() {
			return values.NewString("true"), nil
		}
		return values.NewString("false"), nil
	case values.TypeNull:
		return values.NewString("null"), nil
	case values.TypeArray:
		return values.NewString("[]"), nil // Simplified
	default:
		return values.NewString("null"), nil
	}
}

func jsonDecodeHandler(ctx *vm.ExecutionContext, args []*values.Value) (*values.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("json_decode() expects at least 1 parameter, %d given", len(args))
	}

	// Simplified JSON decoding
	json := args[0].ToString()

	switch json {
	case "null":
		return values.NewNull(), nil
	case "true":
		return values.NewBool(true), nil
	case "false":
		return values.NewBool(false), nil
	case "[]":
		return values.NewArray(), nil
	default:
		// Try to parse as number or string
		if json[0] == '"' && json[len(json)-1] == '"' {
			return values.NewString(json[1 : len(json)-1]), nil
		}
		return values.NewNull(), nil
	}
}

// Built-in extension functions for stdlib

func (stdlib *StandardLibrary) GetBuiltinExtensions() []Extension {
	return []Extension{
		NewMathExtension(),
		NewJsonExtension(),
	}
}

// LoadBuiltinExtensions loads all built-in extensions
func (stdlib *StandardLibrary) LoadBuiltinExtensions() error {
	extensions := stdlib.GetBuiltinExtensions()

	for _, ext := range extensions {
		if err := stdlib.RegisterExtension(ext); err != nil {
			return fmt.Errorf("failed to load built-in extension %s: %v", ext.GetName(), err)
		}
	}

	return nil
}
