package stdlib

import (
	"github.com/wudi/hey/compiler/values"
	"github.com/wudi/hey/compiler/vm"
)

// StandardLibrary represents the PHP standard library
type StandardLibrary struct {
	Constants map[string]*values.Value
	Variables map[string]*values.Value
	Functions map[string]BuiltinFunction
	Classes   map[string]*Class
}

// BuiltinFunction represents a built-in PHP function
type BuiltinFunction struct {
	Name       string
	Handler    FunctionHandler
	Parameters []Parameter
	IsVariadic bool
	MinArgs    int
	MaxArgs    int
}

// FunctionHandler is the type for built-in function handlers
type FunctionHandler func(ctx *vm.ExecutionContext, args []*values.Value) (*values.Value, error)

// Parameter represents a function parameter
type Parameter struct {
	Name         string
	Type         string
	IsReference  bool
	HasDefault   bool
	DefaultValue *values.Value
}

// Class represents a built-in PHP class
type Class struct {
	Name       string
	Parent     string
	Properties map[string]*Property
	Methods    map[string]*Method
	Constants  map[string]*values.Value
	IsAbstract bool
	IsFinal    bool
}

// Property represents a class property
type Property struct {
	Name         string
	Visibility   string // public, private, protected
	IsStatic     bool
	Type         string
	DefaultValue *values.Value
}

// Method represents a class method
type Method struct {
	Name       string
	Visibility string
	IsStatic   bool
	IsAbstract bool
	IsFinal    bool
	Parameters []Parameter
	Handler    FunctionHandler
	IsVariadic bool
}

// NewStandardLibrary creates a new standard library instance
func NewStandardLibrary() *StandardLibrary {
	stdlib := &StandardLibrary{
		Constants: make(map[string]*values.Value),
		Variables: make(map[string]*values.Value),
		Functions: make(map[string]BuiltinFunction),
		Classes:   make(map[string]*Class),
	}

	// Initialize built-in constants
	stdlib.initConstants()

	// Initialize built-in variables
	stdlib.initVariables()

	// Initialize built-in functions
	stdlib.initFunctions()

	// Initialize built-in classes
	stdlib.initClasses()

	return stdlib
}

// RegisterExtension allows external extensions to register functions and classes
func (stdlib *StandardLibrary) RegisterExtension(ext Extension) error {
	// Register extension constants
	for name, value := range ext.GetConstants() {
		stdlib.Constants[name] = value
	}

	// Register extension variables
	for name, value := range ext.GetVariables() {
		stdlib.Variables[name] = value
	}

	// Register extension functions
	for name, fn := range ext.GetFunctions() {
		stdlib.Functions[name] = fn
	}

	// Register extension classes
	for name, class := range ext.GetClasses() {
		stdlib.Classes[name] = class
	}

	return nil
}

// Extension interface for external extensions
type Extension interface {
	GetName() string
	GetVersion() string
	GetConstants() map[string]*values.Value
	GetVariables() map[string]*values.Value
	GetFunctions() map[string]BuiltinFunction
	GetClasses() map[string]*Class
}
