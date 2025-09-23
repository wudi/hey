package vm

import (
	"github.com/wudi/hey/opcodes"
	"github.com/wudi/hey/values"
)

// InstructionExecutor defines the interface for executing VM instructions
type InstructionExecutor interface {
	Execute() (*ExecutionResult, error)
}

// ExecutionResult represents the result of instruction execution
type ExecutionResult struct {
	ShouldAdvanceIP bool           // Whether to advance the instruction pointer
	Result          *values.Value  // Optional result value
	JumpTo          int           // For jump instructions (-1 means no jump)
}

// BaseExecutor provides common functionality for instruction executors
type BaseExecutor struct {
	ctx    *ExecutionContext
	frame  *CallFrame
	inst   *opcodes.Instruction
	reader *OperandReader
}

// NewBaseExecutor creates a new base executor
func NewBaseExecutor(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) *BaseExecutor {
	return &BaseExecutor{
		ctx:    ctx,
		frame:  frame,
		inst:   inst,
		reader: NewOperandReader(ctx, frame, inst),
	}
}

// ReadOperand1 reads the first operand
func (b *BaseExecutor) ReadOperand1() (*values.Value, error) {
	return b.reader.ReadOperand1()
}

// ReadOperand2 reads the second operand
func (b *BaseExecutor) ReadOperand2() (*values.Value, error) {
	return b.reader.ReadOperand2()
}

// ReadBothOperands reads both operands
func (b *BaseExecutor) ReadBothOperands() (*values.Value, *values.Value, error) {
	return b.reader.ReadBothOperands()
}

// WriteResult writes the result
func (b *BaseExecutor) WriteResult(result *values.Value) error {
	return b.reader.WriteResult(result)
}

// CreateAdvanceResult creates a result that advances the IP
func (b *BaseExecutor) CreateAdvanceResult(result *values.Value) (*ExecutionResult, error) {
	var err error
	if result != nil {
		err = b.WriteResult(result)
	}
	return &ExecutionResult{
		ShouldAdvanceIP: true,
		Result:          result,
		JumpTo:          -1,
	}, err
}

// CreateJumpResult creates a result that jumps to a specific IP
func (b *BaseExecutor) CreateJumpResult(jumpTo int) (*ExecutionResult, error) {
	return &ExecutionResult{
		ShouldAdvanceIP: false,
		Result:          nil,
		JumpTo:          jumpTo,
	}, nil
}

// CreateNoAdvanceResult creates a result that doesn't advance the IP
func (b *BaseExecutor) CreateNoAdvanceResult() (*ExecutionResult, error) {
	return &ExecutionResult{
		ShouldAdvanceIP: false,
		Result:          nil,
		JumpTo:          -1,
	}, nil
}

// GetContext returns the execution context
func (b *BaseExecutor) GetContext() *ExecutionContext {
	return b.ctx
}

// GetFrame returns the call frame
func (b *BaseExecutor) GetFrame() *CallFrame {
	return b.frame
}

// GetInstruction returns the instruction
func (b *BaseExecutor) GetInstruction() *opcodes.Instruction {
	return b.inst
}