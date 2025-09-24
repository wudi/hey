package vm

import (
	"fmt"
	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// SimpleCallUserFunction provides a lightweight user function calling mechanism
// specifically designed for callback scenarios to avoid VM execution context interference
func (b *builtinContext) SimpleCallUserFunction(function *registry.Function, args []*values.Value) (*values.Value, error) {
	if b.ctx == nil || b.vm == nil {
		return nil, fmt.Errorf("no execution context or VM available")
	}

	if function == nil {
		return nil, fmt.Errorf("function is nil")
	}

	if function.IsBuiltin {
		return nil, fmt.Errorf("function %s is builtin, not user-defined", function.Name)
	}

	// For simple callback functions, we'll create a minimal execution environment
	// that doesn't interfere with the host builtin function's execution context

	// Create a new VM instance for isolated execution
	callbackVM := &VirtualMachine{}

	// Create a minimal execution context that inherits necessary state
	callbackCtx := &ExecutionContext{
		GlobalVars:    b.ctx.GlobalVars,
		IncludedFiles: b.ctx.IncludedFiles,
		Variables:     b.ctx.Variables,
		Temporaries:   b.ctx.Temporaries,
		ClassTable:    b.ctx.ClassTable,

		UserFunctions:  b.ctx.UserFunctions,
		UserClasses:    b.ctx.UserClasses,
		UserInterfaces: b.ctx.UserInterfaces,
		UserTraits:     b.ctx.UserTraits,

		Stack:      make([]*values.Value, 0),
		CallStack:  make([]*CallFrame, 0),
		Constants:  make([]*values.Value, 0),

		OutputWriter: b.ctx.OutputWriter,
		Halted:       false,
		ExitCode:     0,
	}

	// Create call frame for the user function
	frame := newCallFrame(function.Name, function, function.Instructions, function.Constants)

	// Bind parameters
	for i, param := range function.Parameters {
		var arg *values.Value
		if i < len(args) {
			arg = args[i]
		} else if param.HasDefault && param.DefaultValue != nil {
			arg = param.DefaultValue
		} else {
			arg = values.NewNull()
		}
		frame.setLocal(uint32(i), arg)
		frame.bindSlotName(uint32(i), param.Name)
	}

	// Push frame and execute in isolated context
	callbackCtx.pushFrame(frame)

	// Execute the user function in complete isolation
	for !callbackCtx.Halted && callbackCtx.currentFrame() == frame {
		currentFrame := callbackCtx.currentFrame()
		if currentFrame == nil {
			break
		}

		if currentFrame.IP >= len(currentFrame.Instructions) {
			// Reached end without explicit return
			callbackCtx.popFrame()
			return values.NewNull(), nil
		}

		inst := currentFrame.Instructions[currentFrame.IP]
		currentFrame.IP++

		continued, err := callbackVM.executeInstruction(callbackCtx, currentFrame, inst)
		if err != nil {
			callbackCtx.popFrame()
			return nil, err
		}

		if !continued {
			// Function returned
			break
		}
	}

	// Get return value from isolated stack
	result := values.NewNull()
	if len(callbackCtx.Stack) > 0 {
		result = callbackCtx.Stack[len(callbackCtx.Stack)-1]
	}

	return result, nil
}