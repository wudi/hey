package compiler

import (
	"github.com/wudi/hey/compiler/opcodes"
	"github.com/wudi/hey/compiler/values"
)

// CompilerFunction represents a function during compilation (replaces vm.Function)
type CompilerFunction struct {
	Name         string
	Instructions []opcodes.Instruction
	Constants    []*values.Value
	Parameters   []CompilerParameter
	IsVariadic   bool
	IsGenerator  bool
	IsAnonymous  bool
}

// CompilerParameter represents a function parameter (replaces vm.Parameter)
type CompilerParameter struct {
	Name         string
	Type         string
	IsReference  bool
	HasDefault   bool
	DefaultValue *values.Value
}

// CompilerClass represents a class during compilation (replaces vm.Class)
type CompilerClass struct {
	Name       string
	Parent     string
	Properties map[string]*CompilerProperty
	Methods    map[string]*CompilerFunction
	Constants  map[string]*CompilerClassConstant
	IsAbstract bool
	IsFinal    bool
}

// CompilerProperty represents a class property (replaces vm.Property)
type CompilerProperty struct {
	Name         string
	Type         string
	Visibility   string // public, private, protected
	IsStatic     bool
	DefaultValue *values.Value
}

// CompilerClassConstant represents a class constant (replaces vm.ClassConstant)
type CompilerClassConstant struct {
	Name       string
	Value      *values.Value
	Visibility string // public, private, protected
	Type       string
	IsFinal    bool
	IsAbstract bool
}

// CompilerInterface represents an interface during compilation (replaces vm.Interface)
type CompilerInterface struct {
	Name    string
	Methods map[string]*CompilerInterfaceMethod
	Extends []string
}

// CompilerInterfaceMethod represents an interface method
type CompilerInterfaceMethod struct {
	Name       string
	Visibility string
	Parameters []*CompilerParameter
}

// CompilerTrait represents a trait during compilation (replaces vm.Trait)
type CompilerTrait struct {
	Name       string
	Properties map[string]*CompilerProperty
	Methods    map[string]*CompilerFunction
}

// CompilerEnum represents an enum during compilation
type CompilerEnum struct {
	Name        string
	BackingType string
	Cases       map[string]*values.Value
	Methods     map[string]*CompilerFunction
}