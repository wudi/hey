package vm

import (
	"github.com/wudi/hey/opcodes"
	"github.com/wudi/hey/values"
)

// ComparisonExecutor handles comparison operations
type ComparisonExecutor struct {
	*BaseExecutor
}

// NewComparisonExecutor creates a new comparison executor
func NewComparisonExecutor(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) *ComparisonExecutor {
	return &ComparisonExecutor{
		BaseExecutor: NewBaseExecutor(ctx, frame, inst),
	}
}

// Execute performs comparison operations
func (c *ComparisonExecutor) Execute() (*ExecutionResult, error) {
	op1, op2, err := c.ReadBothOperands()
	if err != nil {
		return nil, err
	}

	var result *values.Value
	inst := c.GetInstruction()

	switch inst.Opcode {
	case opcodes.OP_IS_EQUAL:
		result = c.isEqual(op1, op2)
	case opcodes.OP_IS_NOT_EQUAL:
		result = c.isNotEqual(op1, op2)
	case opcodes.OP_IS_IDENTICAL:
		result = c.isIdentical(op1, op2)
	case opcodes.OP_IS_NOT_IDENTICAL:
		result = c.isNotIdentical(op1, op2)
	case opcodes.OP_IS_SMALLER:
		result = c.isSmaller(op1, op2)
	case opcodes.OP_IS_SMALLER_OR_EQUAL:
		result = c.isSmallerOrEqual(op1, op2)
	default:
		return nil, NewOpcodeError(inst.Opcode)
	}

	return c.CreateAdvanceResult(result)
}

func (c *ComparisonExecutor) isEqual(op1, op2 *values.Value) *values.Value {
	val1 := op1.Deref()
	val2 := op2.Deref()

	// PHP's == comparison with type juggling
	if val1.Type == val2.Type {
		return values.NewBool(val1.Equal(val2))
	}

	// Handle type juggling for different types
	if val1.IsNumeric() && val2.IsNumeric() {
		return values.NewBool(val1.ToFloat() == val2.ToFloat())
	}

	if val1.IsString() || val2.IsString() {
		return values.NewBool(val1.ToString() == val2.ToString())
	}

	return values.NewBool(false)
}

func (c *ComparisonExecutor) isNotEqual(op1, op2 *values.Value) *values.Value {
	result := c.isEqual(op1, op2)
	return values.NewBool(!result.ToBool())
}

func (c *ComparisonExecutor) isIdentical(op1, op2 *values.Value) *values.Value {
	val1 := op1.Deref()
	val2 := op2.Deref()

	// PHP's === comparison - strict type and value check
	if val1.Type != val2.Type {
		return values.NewBool(false)
	}

	return values.NewBool(val1.Equal(val2))
}

func (c *ComparisonExecutor) isNotIdentical(op1, op2 *values.Value) *values.Value {
	result := c.isIdentical(op1, op2)
	return values.NewBool(!result.ToBool())
}

func (c *ComparisonExecutor) isSmaller(op1, op2 *values.Value) *values.Value {
	val1 := op1.Deref()
	val2 := op2.Deref()

	if val1.IsNumeric() && val2.IsNumeric() {
		return values.NewBool(val1.ToFloat() < val2.ToFloat())
	}

	// String comparison
	return values.NewBool(val1.ToString() < val2.ToString())
}

func (c *ComparisonExecutor) isSmallerOrEqual(op1, op2 *values.Value) *values.Value {
	val1 := op1.Deref()
	val2 := op2.Deref()

	if val1.IsNumeric() && val2.IsNumeric() {
		return values.NewBool(val1.ToFloat() <= val2.ToFloat())
	}

	// String comparison
	return values.NewBool(val1.ToString() <= val2.ToString())
}