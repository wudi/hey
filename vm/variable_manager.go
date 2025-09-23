package vm

import (
	"sync"

	"github.com/wudi/hey/values"
)

// VariableManager manages variable storage and access in the VM
type VariableManager struct {
	// Global variables shared across execution
	GlobalVars *sync.Map // map[string]*values.Value

	// Local variables in current scope
	Variables *sync.Map // map[string]*values.Value

	// Temporary variables for expression evaluation
	Temporaries *sync.Map // map[uint32]*values.Value
}

// NewVariableManager creates a new variable manager
func NewVariableManager() *VariableManager {
	return &VariableManager{
		GlobalVars:  &sync.Map{},
		Variables:   &sync.Map{},
		Temporaries: &sync.Map{},
	}
}

// GetGlobalVar safely retrieves a global variable
func (vm *VariableManager) GetGlobalVar(name string) (*values.Value, bool) {
	if val, ok := vm.GlobalVars.Load(name); ok {
		return val.(*values.Value), true
	}
	return nil, false
}

// SetGlobalVar safely sets a global variable
func (vm *VariableManager) SetGlobalVar(name string, val *values.Value) {
	vm.GlobalVars.Store(name, val)
}

// GetVariable safely retrieves a local variable
func (vm *VariableManager) GetVariable(name string) (*values.Value, bool) {
	if val, ok := vm.Variables.Load(name); ok {
		return val.(*values.Value), true
	}
	return nil, false
}

// SetVariable safely sets a local variable
func (vm *VariableManager) SetVariable(name string, value *values.Value) {
	vm.Variables.Store(name, value)
	if sanitized := sanitizeVariableName(name); sanitized != "" && sanitized != name {
		vm.Variables.Store(sanitized, value)
	}
}

// UnsetVariable safely removes a local variable
func (vm *VariableManager) UnsetVariable(name string) {
	vm.Variables.Delete(name)
	if sanitized := sanitizeVariableName(name); sanitized != "" && sanitized != name {
		vm.Variables.Delete(sanitized)
	}
}

// GetTemporary safely retrieves a temporary variable
func (vm *VariableManager) GetTemporary(slot uint32) (*values.Value, bool) {
	if val, ok := vm.Temporaries.Load(slot); ok {
		return val.(*values.Value), true
	}
	return nil, false
}

// SetTemporary safely sets a temporary variable
func (vm *VariableManager) SetTemporary(slot uint32, value *values.Value) {
	vm.Temporaries.Store(slot, value)
}

// EnsureGlobal ensures a global variable exists, creating it as null if not
func (vm *VariableManager) EnsureGlobal(name string) *values.Value {
	// First check GlobalVars
	for _, variant := range globalNameVariants(name) {
		if val, ok := vm.GlobalVars.Load(variant); ok {
			vm.BindGlobalValue(name, val.(*values.Value))
			return val.(*values.Value)
		}
	}

	// Then check Variables
	for _, variant := range globalNameVariants(name) {
		if val, ok := vm.Variables.Load(variant); ok {
			vm.BindGlobalValue(name, val.(*values.Value))
			return val.(*values.Value)
		}
	}

	// Create new null value
	null := values.NewNull()
	vm.BindGlobalValue(name, null)
	return null
}

// BindGlobalValue binds a value to all variants of a global variable name
func (vm *VariableManager) BindGlobalValue(name string, value *values.Value) {
	variants := globalNameVariants(name)
	for _, variant := range variants {
		vm.GlobalVars.Store(variant, value)
	}
}

// UnsetGlobal removes a global variable
func (vm *VariableManager) UnsetGlobal(name string) {
	for _, variant := range globalNameVariants(name) {
		vm.GlobalVars.Delete(variant)
	}
}

// Clear resets all variable storage (useful for testing)
func (vm *VariableManager) Clear() {
	vm.GlobalVars = &sync.Map{}
	vm.Variables = &sync.Map{}
	vm.Temporaries = &sync.Map{}
}

// Copy creates a deep copy of the variable manager state
func (vm *VariableManager) Copy() *VariableManager {
	copy := NewVariableManager()

	// Copy global variables
	vm.GlobalVars.Range(func(key, value interface{}) bool {
		copy.GlobalVars.Store(key, copyValue(value.(*values.Value)))
		return true
	})

	// Copy local variables
	vm.Variables.Range(func(key, value interface{}) bool {
		copy.Variables.Store(key, copyValue(value.(*values.Value)))
		return true
	})

	// Copy temporary variables
	vm.Temporaries.Range(func(key, value interface{}) bool {
		copy.Temporaries.Store(key, copyValue(value.(*values.Value)))
		return true
	})

	return copy
}