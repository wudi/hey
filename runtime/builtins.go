package runtime

import (
	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

var builtinClassStubs = map[string]map[string]struct{}{
	"stdclass": {},
	"exception": {
		"getmessage":       {},
		"getcode":          {},
		"getfile":          {},
		"getline":          {},
		"gettrace":         {},
		"gettraceasstring": {},
	},
}

// GetAllBuiltinFunctions returns all builtin functions from all modules
func GetAllBuiltinFunctions() []*registry.Function {
	var functions []*registry.Function

	// Add functions from each module
	functions = append(functions, GetArrayFunctions()...)
	functions = append(functions, GetStringFunctions()...)
	functions = append(functions, GetRegexFunctions()...)
	functions = append(functions, GetTypeFunctions()...)
	functions = append(functions, GetEncodingFunctions()...)
	functions = append(functions, GetFilesystemFunctions()...)
	functions = append(functions, GetSystemFunctions()...)
	functions = append(functions, GetTimeFunctions()...)
	functions = append(functions, GetMathFunctions()...)
	functions = append(functions, GetOutputFunctions()...)
	functions = append(functions, GetReflectionFunctions()...)
	functions = append(functions, GetVariableFunctions()...)
	functions = append(functions, GetConcurrencyFunctions()...)

	return functions
}

// GetAllBuiltinClasses returns all builtin classes from all modules
func GetAllBuiltinClasses() []*registry.ClassDescriptor {
	var classes []*registry.ClassDescriptor

	// Add stdClass - PHP's generic object class
	stdClass := &registry.ClassDescriptor{
		Name:       "stdClass",
		Parent:     "",
		Interfaces: []string{},
		Traits:     []string{},
		Methods:    make(map[string]*registry.MethodDescriptor),
		Properties: make(map[string]*registry.PropertyDescriptor),
		Constants:  make(map[string]*registry.ConstantDescriptor),
		IsAbstract: false,
		IsFinal:    false,
	}
	classes = append(classes, stdClass)

	// Add classes from exception module
	classes = append(classes, GetClasses()...)

	// Add classes from iterator module
	classes = append(classes, GetIteratorClasses()...)

	// Add classes from concurrency module
	classes = append(classes, GetConcurrencyClasses()...)

	return classes
}

// GetAllBuiltinInterfaces returns all builtin interfaces from all modules
func GetAllBuiltinInterfaces() []*registry.Interface {
	var interfaces []*registry.Interface

	// Add interfaces from iterator module
	interfaces = append(interfaces, GetInterfaces()...)

	return interfaces
}

// GetAllBuiltinConstants returns all builtin constants
func GetAllBuiltinConstants() []*registry.ConstantDescriptor {
	return []*registry.ConstantDescriptor{
		{
			Name:  "CASE_LOWER",
			Value: values.NewInt(0),
		},
		{
			Name:  "CASE_UPPER",
			Value: values.NewInt(1),
		},
		{
			Name:  "SORT_REGULAR",
			Value: values.NewInt(0),
		},
		{
			Name:  "SORT_NUMERIC",
			Value: values.NewInt(1),
		},
		{
			Name:  "SORT_STRING",
			Value: values.NewInt(2),
		},
		{
			Name:  "SORT_DESC",
			Value: values.NewInt(3),
		},
		{
			Name:  "SORT_ASC",
			Value: values.NewInt(4),
		},
		{
			Name:  "SORT_LOCALE_STRING",
			Value: values.NewInt(5),
		},
		{
			Name:  "SORT_NATURAL",
			Value: values.NewInt(6),
		},
		{
			Name:  "SORT_FLAG_CASE",
			Value: values.NewInt(8),
		},
	}
}

// Legacy variable for backwards compatibility
var builtinFunctionSpecs = GetAllBuiltinFunctions()

// BuiltinMethodImpl represents a builtin method implementation
type BuiltinMethodImpl struct {
	function *registry.Function
}

func NewBuiltinMethodImpl(function *registry.Function) *BuiltinMethodImpl {
	return &BuiltinMethodImpl{function: function}
}

func (b *BuiltinMethodImpl) ImplementationKind() string { return "builtin" }

func (b *BuiltinMethodImpl) GetFunction() *registry.Function {
	return b.function
}