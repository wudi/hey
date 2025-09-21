package vm

import (
	"fmt"
	"strings"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// builtinContext adapts ExecutionContext operations to the registry's builtin
// call interface without creating package cycles.
type builtinContext struct {
	vm  *VirtualMachine
	ctx *ExecutionContext
}

func (b *builtinContext) WriteOutput(val *values.Value) error {
	if b.ctx == nil || b.ctx.OutputWriter == nil {
		return fmt.Errorf("no output writer configured")
	}
	_, err := fmt.Fprint(b.ctx.OutputWriter, val.ToString())
	return err
}

func (b *builtinContext) GetGlobal(name string) (*values.Value, bool) {
	if b.ctx == nil {
		return nil, false
	}
	v, ok := b.ctx.GlobalVars[name]
	return v, ok
}

func (b *builtinContext) SetGlobal(name string, val *values.Value) {
	if b.ctx == nil {
		return
	}
	b.ctx.GlobalVars[name] = val
}

func (b *builtinContext) SymbolRegistry() *registry.Registry {
	return registry.GlobalRegistry
}

func (b *builtinContext) LookupUserFunction(name string) (*registry.Function, bool) {
	if b.ctx == nil || b.ctx.UserFunctions == nil {
		return nil, false
	}
	fn, ok := b.ctx.UserFunctions[strings.ToLower(name)]
	return fn, ok
}

func (b *builtinContext) LookupUserClass(name string) (*registry.Class, bool) {
	if b.ctx == nil || b.ctx.UserClasses == nil {
		return nil, false
	}
	if class, ok := b.ctx.UserClasses[strings.ToLower(name)]; ok {
		return class, true
	}
	if b.ctx.ClassTable != nil {
		if runtimeCls, ok := b.ctx.ClassTable[strings.ToLower(name)]; ok && runtimeCls != nil {
			return runtimeCls.Descriptor, runtimeCls.Descriptor != nil
		}
	}
	return nil, false
}

func (b *builtinContext) Halt(exitCode int, message string) error {
	if b.ctx == nil {
		return fmt.Errorf("no execution context available")
	}

	// Print message if provided
	if message != "" {
		if b.ctx.OutputWriter != nil {
			fmt.Fprint(b.ctx.OutputWriter, message)
		}
	}

	// Set exit code and halt execution
	b.ctx.ExitCode = exitCode
	b.ctx.Halted = true

	return nil
}
