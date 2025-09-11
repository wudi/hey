package registry

import (
	"github.com/wudi/hey/opcodes"
	"github.com/wudi/hey/values"
)

// Function represents a compiled PHP function
type Function struct {
	Name         string
	Instructions []*opcodes.Instruction
	Constants    []*values.Value
	Parameters   []*Parameter
	IsVariadic   bool
	IsGenerator  bool
	IsAnonymous  bool
}

// Parameter represents a function parameter
type Parameter struct {
	Name         string
	Type         string
	IsReference  bool
	HasDefault   bool
	DefaultValue *values.Value
}

// Class represents a compiled PHP class
type Class struct {
	Name       string
	Parent     string
	Properties map[string]*Property
	Methods    map[string]*Function
	Constants  map[string]*ClassConstant
	IsAbstract bool
	IsFinal    bool
}

// ClassConstant represents a class constant with metadata
type ClassConstant struct {
	Name       string
	Value      *values.Value
	Visibility string // public, private, protected
	Type       string // Type hint for PHP 8.3+
	IsFinal    bool   // final const
	IsAbstract bool   // abstract const (interfaces/abstract classes)
}

// Property represents a class property
type Property struct {
	Name         string
	Type         string
	Visibility   string // public, private, protected
	IsStatic     bool
	DefaultValue *values.Value
}

// Interface represents a PHP interface
type Interface struct {
	Name    string
	Methods map[string]*InterfaceMethod
	Extends []string // Parent interfaces
}

// InterfaceMethod represents a method in an interface
type InterfaceMethod struct {
	Name       string
	Visibility string
	Parameters []*Parameter
}

// Trait represents a PHP trait
type Trait struct {
	Name       string
	Properties map[string]*Property
	Methods    map[string]*Function
}

// Enum represents an enum during compilation
type Enum struct {
	Name        string
	BackingType string
	Cases       map[string]*values.Value
	Methods     map[string]*Function
}
