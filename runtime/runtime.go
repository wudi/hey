package runtime

import (
	"fmt"
	"sync"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

var (
	bootstrapOnce sync.Once
	bootstrapErr  error

	vmInitOnce sync.Once
	vmInitErr  error

	// GlobalRegistry mirrors registry.GlobalRegistry for backward compatibility.
	GlobalRegistry *registry.Registry
)

// Bootstrap sets up the shared runtime components such as the unified registry
// and builtin function catalogue. It can be called repeatedly; initialization
// happens only once.
func Bootstrap() error {
	bootstrapOnce.Do(func() {
		registry.Initialize()
		GlobalRegistry = registry.GlobalRegistry
		bootstrapErr = registerBuiltinSymbols()
	})
	return bootstrapErr
}

// UnifiedBootstrap is a compatibility helper used by code paths that expect a
// more feature-rich initialization routine. Currently it reuses Bootstrap.
func UnifiedBootstrap() error {
	return Bootstrap()
}

// VMIntegration exposes runtime level globals that need to be injected into
// execution contexts before running user bytecode.
type VMIntegration struct {
	mu      sync.RWMutex
	globals map[string]*values.Value
}

// GlobalVMIntegration is populated by InitializeVMIntegration and provides
// shared state (e.g. superglobals) to new execution contexts.
var GlobalVMIntegration *VMIntegration

// InitializeVMIntegration prepares the VM integration layer. It ensures the
// runtime has been bootstrapped and then allocates the integration structure.
func InitializeVMIntegration() error {
	if err := Bootstrap(); err != nil {
		return err
	}

	vmInitOnce.Do(func() {
		GlobalVMIntegration = &VMIntegration{
			globals: make(map[string]*values.Value),
		}
		// Populate canonical PHP superglobals as empty structures so that new
		// execution contexts can copy them lazily.
		GlobalVMIntegration.globals["$GLOBALS"] = values.NewArray()
		GlobalVMIntegration.globals["$_SERVER"] = values.NewArray()
		GlobalVMIntegration.globals["$_GET"] = values.NewArray()
		GlobalVMIntegration.globals["$_POST"] = values.NewArray()
		GlobalVMIntegration.globals["$_FILES"] = values.NewArray()
		GlobalVMIntegration.globals["$_COOKIE"] = values.NewArray()
		GlobalVMIntegration.globals["$_SESSION"] = values.NewArray()
		GlobalVMIntegration.globals["$_REQUEST"] = values.NewArray()
		GlobalVMIntegration.globals["$_ENV"] = values.NewArray()
	})

	return vmInitErr
}

// GetAllVariables returns a shallow copy of the currently registered globals.
func (v *VMIntegration) GetAllVariables() map[string]*values.Value {
	v.mu.RLock()
	defer v.mu.RUnlock()
	out := make(map[string]*values.Value, len(v.globals))
	for k, val := range v.globals {
		out[k] = val
	}
	return out
}

// SetGlobal sets a global variable value.
func (v *VMIntegration) SetGlobal(name string, val *values.Value) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.globals[name] = val
}

// GetGlobal fetches a global variable by name.
func (v *VMIntegration) GetGlobal(name string) (*values.Value, bool) {
	v.mu.RLock()
	defer v.mu.RUnlock()
	val, ok := v.globals[name]
	return val, ok
}

// registerBuiltinSymbols wires up the builtin function catalogue. Additional
// runtime entities (constants, classes) can be registered here as needed.
func registerBuiltinSymbols() error {
	// Ensure the registry is ready before we attempt to register builtins.
	if registry.GlobalRegistry == nil {
		registry.Initialize()
	}

	for _, spec := range builtinFunctionSpecs {
		fn := &registry.Function{
			Name:       spec.Name,
			Parameters: spec.Parameters,
			ReturnType: spec.ReturnType,
			IsVariadic: spec.IsVariadic,
			IsBuiltin:  true,
			Builtin:    spec.Impl,
		}

		if spec.MinArgs == 0 && len(spec.Parameters) > 0 {
			spec.MinArgs = len(spec.Parameters)
		}
		if spec.MaxArgs == 0 {
			if spec.IsVariadic {
				spec.MaxArgs = -1
			} else {
				spec.MaxArgs = len(spec.Parameters)
			}
		}

		fn.MinArgs = spec.MinArgs
		fn.MaxArgs = spec.MaxArgs
		if spec.Name == "go" {
			fn.Handler = func(_ interface{}, args []*values.Value) (*values.Value, error) {
				return spec.Impl(nil, args)
			}
		}

		if err := registry.GlobalRegistry.RegisterFunction(fn); err != nil {
			return fmt.Errorf("register builtin %s: %w", spec.Name, err)
		}
	}

	if err := registerWaitGroupClass(); err != nil {
		return err
	}

	// Register builtin constants.
	for _, v := range builtinConstants {
		if err := registry.GlobalRegistry.RegisterConstant(v); err != nil {
			return fmt.Errorf("register constant %s: %w", v.Name, err)
		}
	}

	return nil
}
