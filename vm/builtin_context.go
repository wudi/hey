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
	return b.ctx.GetGlobalVar(name)
}

func (b *builtinContext) SetGlobal(name string, val *values.Value) {
	if b.ctx == nil {
		return
	}
	b.ctx.SetGlobalVar(name, val)
}

func (b *builtinContext) SymbolRegistry() *registry.Registry {
	return registry.GlobalRegistry
}

func (b *builtinContext) LookupUserFunction(name string) (*registry.Function, bool) {
	if b.ctx == nil {
		return nil, false
	}
	return b.ctx.GetUserFunction(strings.ToLower(name))
}

func (b *builtinContext) CallUserFunction(function *registry.Function, args []*values.Value) (*values.Value, error) {
	if b.ctx == nil || b.vm == nil {
		return nil, fmt.Errorf("no execution context or VM available")
	}

	if function == nil {
		return nil, fmt.Errorf("function is nil")
	}

	if function.IsBuiltin {
		return nil, fmt.Errorf("function %s is builtin, not user-defined", function.Name)
	}

	// Create completely isolated execution environment
	// Save ALL current VM state that might be affected
	savedStack := make([]*values.Value, len(b.ctx.Stack))
	copy(savedStack, b.ctx.Stack)
	savedCallStack := make([]*CallFrame, len(b.ctx.CallStack))
	copy(savedCallStack, b.ctx.CallStack)
	savedHalted := b.ctx.Halted
	savedExitCode := b.ctx.ExitCode

	// Reset VM state for isolated execution
	b.ctx.Stack = nil
	b.ctx.Halted = false
	b.ctx.ExitCode = 0

	// Create a new call frame for the user function
	child := newCallFrame(function.Name, function, function.Instructions, function.Constants)

	// Bind parameters using the same logic as the VM (by slot index)
	for i, param := range function.Parameters {
		var arg *values.Value

		if i < len(args) {
			// Normal parameter - copy the value
			arg = args[i]
		} else if param.HasDefault && param.DefaultValue != nil {
			// Use default parameter value
			arg = param.DefaultValue
		} else {
			// No argument provided and no default - use null
			arg = values.NewNull()
		}

		// Set the parameter in the local slot
		child.setLocal(uint32(i), arg)
		// Bind the slot name for debugging/reflection
		child.bindSlotName(uint32(i), param.Name)
	}

	// Push the frame onto the execution stack
	b.ctx.pushFrame(child)

	// Execute until the frame returns
	var userResult *values.Value = values.NewNull()
	for !b.ctx.Halted && b.ctx.currentFrame() == child {
		frame := b.ctx.currentFrame()
		if frame == nil {
			break
		}

		if frame.IP >= len(frame.Instructions) {
			// Reached end without explicit return - return null
			b.ctx.popFrame()
			userResult = values.NewNull()
			break
		}

		inst := frame.Instructions[frame.IP]
		frame.IP++

		continued, err := b.vm.executeInstruction(b.ctx, frame, inst)
		if err != nil {
			b.ctx.popFrame() // Clean up the frame on error
			// Restore original VM state completely
			b.ctx.Stack = savedStack
			b.ctx.CallStack = savedCallStack
			b.ctx.Halted = savedHalted
			b.ctx.ExitCode = savedExitCode
			return nil, err
		}

		if !continued {
			// Execution should stop (e.g., return was called)
			if len(b.ctx.Stack) > 0 {
				userResult = b.ctx.Stack[len(b.ctx.Stack)-1]
			}
			break
		}
	}

	// If we still have the result on stack, get it
	if len(b.ctx.Stack) > 0 && userResult.IsNull() {
		userResult = b.ctx.Stack[len(b.ctx.Stack)-1]
	}

	// Completely restore original VM state
	b.ctx.Stack = savedStack
	b.ctx.CallStack = savedCallStack
	b.ctx.Halted = savedHalted
	b.ctx.ExitCode = savedExitCode

	return userResult, nil
}

func (b *builtinContext) LookupUserClass(name string) (*registry.Class, bool) {
	if b.ctx == nil {
		return nil, false
	}
	if class, ok := b.ctx.GetUserClass(strings.ToLower(name)); ok {
		return class, true
	}
	if runtimeCls, ok := b.ctx.getClass(strings.ToLower(name)); ok && runtimeCls != nil {
		return runtimeCls.Descriptor, runtimeCls.Descriptor != nil
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

func (b *builtinContext) GetExecutionContext() registry.ExecutionContextInterface {
	return b.ctx
}

func (b *builtinContext) GetOutputBufferStack() registry.OutputBufferStackInterface {
	if b.ctx == nil || b.ctx.OutputBufferStack == nil {
		return nil
	}
	return b.ctx.OutputBufferStack
}
