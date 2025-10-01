package vmfactory

import (
	"fmt"

	"github.com/wudi/hey/compiler/ast"
	"github.com/wudi/hey/opcodes"
	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
	"github.com/wudi/hey/vm"
)

// Compiler interface to avoid circular dependency with concrete compiler package
type Compiler interface {
	SetCurrentFile(path string)
	Compile(node ast.Node) error
	GetBytecode() []*opcodes.Instruction
	GetConstants() []*values.Value
	Functions() map[string]*registry.Function
	Classes() map[string]*registry.Class
	Interfaces() map[string]*registry.Interface
	Traits() map[string]*registry.Trait
}

// CompilerFactory creates compiler instances to avoid direct import
type CompilerFactory func() Compiler

// VMFactory creates VirtualMachine instances with pre-configured compiler callbacks.
// This eliminates the need for manual CompilerCallback setup in every usage.
type VMFactory struct {
	compilerFactory CompilerFactory
}

// NewVMFactory creates a new VM factory with the provided compiler factory.
// The compiler factory is injected to avoid circular dependencies.
func NewVMFactory(compilerFactory CompilerFactory) *VMFactory {
	return &VMFactory{
		compilerFactory: compilerFactory,
	}
}

// CreateVM creates a new VirtualMachine with properly configured CompilerCallback.
// This replaces the manual setupCompilerCallback pattern throughout the codebase.
func (f *VMFactory) CreateVM() *vm.VirtualMachine {
	vmachine := vm.NewVirtualMachine()
	vmachine.CompilerCallback = f.createCompilerCallback(vmachine)
	return vmachine
}

// createCompilerCallback returns the standard compiler callback implementation.
// This consolidates the duplicated callback logic from cmd/hey/main.go.
func (f *VMFactory) createCompilerCallback(vmachine *vm.VirtualMachine) vm.CompilerCallback {
	return func(ctx *vm.ExecutionContext, program *ast.Program, filePath string, isRequired bool) (*values.Value, error) {
		comp := f.compilerFactory()
		if filePath != "" {
			comp.SetCurrentFile(filePath)
		}
		if err := comp.Compile(program); err != nil {
			return nil, fmt.Errorf("compilation error in %s: %v", filePath, err)
		}

		// Execute the included file directly in the same context
		// This ensures variables defined in the include are accessible to the caller
		err := vmachine.Execute(ctx, comp.GetBytecode(), comp.GetConstants(),
			comp.Functions(), comp.Classes(), comp.Interfaces(), comp.Traits())
		if err != nil {
			return nil, fmt.Errorf("execution error in %s: %v", filePath, err)
		}

		// After include execution, check the stack for return value
		// The handleReturn function pushes return values onto the stack when returning from include files
		if ctx.Halted && len(ctx.Stack) > 0 {
			returnValue := ctx.Stack[len(ctx.Stack)-1]
			ctx.Stack = ctx.Stack[:len(ctx.Stack)-1]
			ctx.Halted = false // Reset Halted state so main script can continue
			if returnValue.IsNull() {
				return values.NewInt(1), nil
			}
			return returnValue, nil
		}

		// No return statement in included file, return 1 by default
		ctx.Halted = false // Reset Halted state
		return values.NewInt(1), nil
	}
}