package runtime

import (
	"fmt"

	"github.com/wudi/hey/values"
)

// VMIntegration provides integration between the runtime registry and VM
type VMIntegration struct {
	registry *RuntimeRegistry
}

// NewVMIntegration creates a new VM integration
func NewVMIntegration(registry *RuntimeRegistry) *VMIntegration {
	return &VMIntegration{
		registry: registry,
	}
}

// GetSuperGlobal retrieves a superglobal variable
func (vi *VMIntegration) GetSuperGlobal(name string) (*values.Value, bool) {
	if vi.registry == nil {
		return nil, false
	}
	return vi.registry.GetVariable(name)
}

// SetSuperGlobal sets a superglobal variable
func (vi *VMIntegration) SetSuperGlobal(name string, value *values.Value) error {
	if vi.registry == nil {
		return fmt.Errorf("runtime registry not initialized")
	}

	// Check if it's a built-in variable
	if vi.registry.builtinVariables[name] {
		return vi.registry.RegisterVariable(name, value, true, "")
	}

	return vi.registry.RegisterVariable(name, value, false, "")
}

// GetConstant retrieves a constant
func (vi *VMIntegration) GetConstant(name string) (*values.Value, bool) {
	if vi.registry == nil {
		return nil, false
	}
	return vi.registry.GetConstant(name)
}

// GetAllVariables returns all registered variables for VM initialization
func (vi *VMIntegration) GetAllVariables() map[string]*values.Value {
	if vi.registry == nil {
		return make(map[string]*values.Value)
	}
	return vi.registry.GetAllVariables()
}

// GetAllConstants returns all registered constants for VM initialization
func (vi *VMIntegration) GetAllConstants() map[string]*values.Value {
	if vi.registry == nil {
		return make(map[string]*values.Value)
	}
	return vi.registry.GetAllConstants()
}

// CallFunction calls a registered function
func (vi *VMIntegration) CallFunction(ctx ExecutionContext, name string, args []*values.Value) (*values.Value, error) {
	if vi.registry == nil {
		return nil, fmt.Errorf("runtime registry not initialized")
	}
	return vi.registry.CallFunction(ctx, name, args)
}

// HasFunction checks if a function is registered
func (vi *VMIntegration) HasFunction(name string) bool {
	if vi.registry == nil {
		return false
	}
	return vi.registry.HasFunction(name)
}

// IsBuiltinVariable checks if a variable is built-in
func (vi *VMIntegration) IsBuiltinVariable(name string) bool {
	if vi.registry == nil {
		return false
	}
	return vi.registry.builtinVariables[name]
}

// Global integration instance
var GlobalVMIntegration *VMIntegration

// InitializeVMIntegration initializes the global VM integration
func InitializeVMIntegration() error {
	if GlobalRegistry == nil {
		return fmt.Errorf("runtime registry not initialized - call Bootstrap() first")
	}

	GlobalVMIntegration = NewVMIntegration(GlobalRegistry)
	return nil
}

// Helper functions for easy VM integration

// GetBuiltinConstantForVM retrieves a built-in constant for VM use
func GetBuiltinConstantForVM(name string) (*values.Value, bool) {
	if GlobalVMIntegration == nil {
		return nil, false
	}
	return GlobalVMIntegration.GetConstant(name)
}

// GetSuperGlobalForVM retrieves a superglobal for VM use
func GetSuperGlobalForVM(name string) (*values.Value, bool) {
	if GlobalVMIntegration == nil {
		return nil, false
	}
	return GlobalVMIntegration.GetSuperGlobal(name)
}

// SetSuperGlobalForVM sets a superglobal from VM
func SetSuperGlobalForVM(name string, value *values.Value) error {
	if GlobalVMIntegration == nil {
		return fmt.Errorf("VM integration not initialized")
	}
	return GlobalVMIntegration.SetSuperGlobal(name, value)
}
