package compiler

import (
	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// CompilerEnum represents an enum during compilation
type CompilerEnum struct {
	Name        string
	BackingType string
	Cases       map[string]*values.Value
	Methods     map[string]*registry.Function
}
