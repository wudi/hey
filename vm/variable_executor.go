package vm

import (
	"github.com/wudi/hey/opcodes"
	"github.com/wudi/hey/values"
)

// VariableExecutor handles variable assignment operations
type VariableExecutor struct {
	*BaseExecutor
}

// NewVariableExecutor creates a new variable executor
func NewVariableExecutor(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) *VariableExecutor {
	return &VariableExecutor{
		BaseExecutor: NewBaseExecutor(ctx, frame, inst),
	}
}

// Execute performs variable operations
func (v *VariableExecutor) Execute() (*ExecutionResult, error) {
	inst := v.GetInstruction()

	switch inst.Opcode {
	case opcodes.OP_ASSIGN:
		return v.executeAssign()
	case opcodes.OP_ASSIGN_REF:
		return v.executeAssignRef()
	default:
		return nil, NewOpcodeError(inst.Opcode)
	}
}

func (v *VariableExecutor) executeAssign() (*ExecutionResult, error) {
	op1, err := v.ReadOperand1()
	if err != nil {
		return nil, err
	}

	// For assignment, we write the value to the result location
	if err := v.WriteResult(op1); err != nil {
		return nil, err
	}

	return v.CreateAdvanceResult(op1)
}

func (v *VariableExecutor) executeAssignRef() (*ExecutionResult, error) {
	op1, err := v.ReadOperand1()
	if err != nil {
		return nil, err
	}

	// Create a reference to the value
	ref := values.NewReference(op1)

	if err := v.WriteResult(ref); err != nil {
		return nil, err
	}

	return v.CreateAdvanceResult(ref)
}