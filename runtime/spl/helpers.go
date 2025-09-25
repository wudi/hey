package spl

import (
	"github.com/wudi/hey/registry"
)

// BuiltinMethodImpl represents a builtin method implementation
type BuiltinMethodImpl struct {
	function *registry.Function
}

// NewBuiltinMethodImpl creates a new builtin method implementation
func NewBuiltinMethodImpl(function *registry.Function) *BuiltinMethodImpl {
	return &BuiltinMethodImpl{function: function}
}

// ImplementationKind returns "builtin"
func (b *BuiltinMethodImpl) ImplementationKind() string { return "builtin" }

// GetFunction returns the underlying function
func (b *BuiltinMethodImpl) GetFunction() *registry.Function {
	return b.function
}