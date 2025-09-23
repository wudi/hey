package vm

import (
	"github.com/wudi/hey/opcodes"
)

// InstructionFactory creates appropriate executor for each instruction type
type InstructionFactory struct {
	executors map[opcodes.Opcode]func(*ExecutionContext, *CallFrame, *opcodes.Instruction) InstructionExecutor
}

// NewInstructionFactory creates a new instruction factory with all executors registered
func NewInstructionFactory() *InstructionFactory {
	factory := &InstructionFactory{
		executors: make(map[opcodes.Opcode]func(*ExecutionContext, *CallFrame, *opcodes.Instruction) InstructionExecutor),
	}

	factory.registerExecutors()
	return factory
}

// CreateExecutor creates the appropriate executor for the given instruction
func (f *InstructionFactory) CreateExecutor(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (InstructionExecutor, error) {
	if creator, exists := f.executors[inst.Opcode]; exists {
		return creator(ctx, frame, inst), nil
	}

	return nil, NewOpcodeError(inst.Opcode)
}

// registerExecutors registers all instruction executors
func (f *InstructionFactory) registerExecutors() {
	// Arithmetic operations
	f.executors[opcodes.OP_ADD] = func(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) InstructionExecutor {
		return NewArithmeticExecutor(ctx, frame, inst)
	}
	f.executors[opcodes.OP_SUB] = func(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) InstructionExecutor {
		return NewArithmeticExecutor(ctx, frame, inst)
	}
	f.executors[opcodes.OP_MUL] = func(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) InstructionExecutor {
		return NewArithmeticExecutor(ctx, frame, inst)
	}
	f.executors[opcodes.OP_DIV] = func(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) InstructionExecutor {
		return NewArithmeticExecutor(ctx, frame, inst)
	}
	f.executors[opcodes.OP_MOD] = func(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) InstructionExecutor {
		return NewArithmeticExecutor(ctx, frame, inst)
	}
	f.executors[opcodes.OP_POW] = func(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) InstructionExecutor {
		return NewArithmeticExecutor(ctx, frame, inst)
	}

	// Variable operations
	f.executors[opcodes.OP_ASSIGN] = func(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) InstructionExecutor {
		return NewVariableExecutor(ctx, frame, inst)
	}
	f.executors[opcodes.OP_ASSIGN_REF] = func(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) InstructionExecutor {
		return NewVariableExecutor(ctx, frame, inst)
	}

	// Comparison operations
	f.executors[opcodes.OP_IS_EQUAL] = func(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) InstructionExecutor {
		return NewComparisonExecutor(ctx, frame, inst)
	}
	f.executors[opcodes.OP_IS_NOT_EQUAL] = func(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) InstructionExecutor {
		return NewComparisonExecutor(ctx, frame, inst)
	}
	f.executors[opcodes.OP_IS_IDENTICAL] = func(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) InstructionExecutor {
		return NewComparisonExecutor(ctx, frame, inst)
	}
	f.executors[opcodes.OP_IS_NOT_IDENTICAL] = func(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) InstructionExecutor {
		return NewComparisonExecutor(ctx, frame, inst)
	}
	f.executors[opcodes.OP_IS_SMALLER] = func(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) InstructionExecutor {
		return NewComparisonExecutor(ctx, frame, inst)
	}
	f.executors[opcodes.OP_IS_SMALLER_OR_EQUAL] = func(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) InstructionExecutor {
		return NewComparisonExecutor(ctx, frame, inst)
	}

	// Control flow operations are still handled by legacy code

	// Additional operations will be added to the factory incrementally
	// For now, only core operations (arithmetic, comparison, assignment, etc.) are implemented

	// String operations
	f.executors[opcodes.OP_CONCAT] = func(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) InstructionExecutor {
		return NewStringExecutor(ctx, frame, inst)
	}

	// Special operations
	f.executors[opcodes.OP_NOP] = func(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) InstructionExecutor {
		return NewNopExecutor(ctx, frame, inst)
	}
	f.executors[opcodes.OP_EXIT] = func(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) InstructionExecutor {
		return NewExitExecutor(ctx, frame, inst)
	}
}

// IsSupported checks if an opcode is supported by the factory
func (f *InstructionFactory) IsSupported(opcode opcodes.Opcode) bool {
	_, exists := f.executors[opcode]
	return exists
}

// GetSupportedOpcodes returns all supported opcodes
func (f *InstructionFactory) GetSupportedOpcodes() []opcodes.Opcode {
	opcodes := make([]opcodes.Opcode, 0, len(f.executors))
	for opcode := range f.executors {
		opcodes = append(opcodes, opcode)
	}
	return opcodes
}