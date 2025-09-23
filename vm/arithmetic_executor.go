package vm

import (
	"github.com/wudi/hey/opcodes"
	"github.com/wudi/hey/values"
)

// ArithmeticExecutor handles arithmetic operations
type ArithmeticExecutor struct {
	*BaseExecutor
}

// NewArithmeticExecutor creates a new arithmetic executor
func NewArithmeticExecutor(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) *ArithmeticExecutor {
	return &ArithmeticExecutor{
		BaseExecutor: NewBaseExecutor(ctx, frame, inst),
	}
}

// Execute performs the arithmetic operation
func (a *ArithmeticExecutor) Execute() (*ExecutionResult, error) {
	op1, op2, err := a.ReadBothOperands()
	if err != nil {
		return nil, err
	}

	var result *values.Value
	inst := a.GetInstruction()

	switch inst.Opcode {
	case opcodes.OP_ADD:
		result = a.add(op1, op2)
	case opcodes.OP_SUB:
		result = a.subtract(op1, op2)
	case opcodes.OP_MUL:
		result = a.multiply(op1, op2)
	case opcodes.OP_DIV:
		result, err = a.divide(op1, op2)
		if err != nil {
			return nil, err
		}
	case opcodes.OP_MOD:
		result, err = a.modulo(op1, op2)
		if err != nil {
			return nil, err
		}
	case opcodes.OP_POW:
		result = a.power(op1, op2)
	default:
		return nil, NewOpcodeError(inst.Opcode)
	}

	return a.CreateAdvanceResult(result)
}

func (a *ArithmeticExecutor) add(op1, op2 *values.Value) *values.Value {
	// Dereference if needed
	val1 := op1.Deref()
	val2 := op2.Deref()

	// Handle numeric addition
	if val1.IsNumeric() && val2.IsNumeric() {
		if val1.IsFloat() || val2.IsFloat() {
			return values.NewFloat(val1.ToFloat() + val2.ToFloat())
		}
		return values.NewInt(val1.ToInt() + val2.ToInt())
	}

	// Handle array addition (PHP array + array)
	if val1.IsArray() && val2.IsArray() {
		result := values.NewArray()
		arr1 := val1.Data.(*values.Array)
		arr2 := val2.Data.(*values.Array)

		// Copy all elements from first array
		for k, v := range arr1.Elements {
			result.Data.(*values.Array).Elements[k] = v
		}

		// Add elements from second array that don't exist in first
		for k, v := range arr2.Elements {
			if _, exists := result.Data.(*values.Array).Elements[k]; !exists {
				result.Data.(*values.Array).Elements[k] = v
			}
		}

		// Update NextIndex
		if arr1.NextIndex > result.Data.(*values.Array).NextIndex {
			result.Data.(*values.Array).NextIndex = arr1.NextIndex
		}
		if arr2.NextIndex > result.Data.(*values.Array).NextIndex {
			result.Data.(*values.Array).NextIndex = arr2.NextIndex
		}

		return result
	}

	// Default: convert to numbers and add
	return values.NewFloat(val1.ToFloat() + val2.ToFloat())
}

func (a *ArithmeticExecutor) subtract(op1, op2 *values.Value) *values.Value {
	val1 := op1.Deref()
	val2 := op2.Deref()

	if val1.IsFloat() || val2.IsFloat() {
		return values.NewFloat(val1.ToFloat() - val2.ToFloat())
	}
	return values.NewInt(val1.ToInt() - val2.ToInt())
}

func (a *ArithmeticExecutor) multiply(op1, op2 *values.Value) *values.Value {
	val1 := op1.Deref()
	val2 := op2.Deref()

	if val1.IsFloat() || val2.IsFloat() {
		return values.NewFloat(val1.ToFloat() * val2.ToFloat())
	}
	return values.NewInt(val1.ToInt() * val2.ToInt())
}

func (a *ArithmeticExecutor) divide(op1, op2 *values.Value) (*values.Value, error) {
	val1 := op1.Deref()
	val2 := op2.Deref()

	divisor := val2.ToFloat()
	if divisor == 0 {
		return nil, NewDivisionByZeroError()
	}

	return values.NewFloat(val1.ToFloat() / divisor), nil
}

func (a *ArithmeticExecutor) modulo(op1, op2 *values.Value) (*values.Value, error) {
	val1 := op1.Deref()
	val2 := op2.Deref()

	divisor := val2.ToInt()
	if divisor == 0 {
		return nil, NewModuloByZeroError()
	}

	return values.NewInt(val1.ToInt() % divisor), nil
}

func (a *ArithmeticExecutor) power(op1, op2 *values.Value) *values.Value {
	val1 := op1.Deref()
	val2 := op2.Deref()

	base := val1.ToFloat()
	exponent := val2.ToFloat()

	// Simple power implementation (for now)
	result := 1.0
	if exponent > 0 {
		for i := 0; i < int(exponent); i++ {
			result *= base
		}
	} else if exponent < 0 {
		for i := 0; i < int(-exponent); i++ {
			result /= base
		}
	}

	if val1.IsInt() && val2.IsInt() && exponent >= 0 {
		return values.NewInt(int64(result))
	}
	return values.NewFloat(result)
}