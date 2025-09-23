package vm

import (
	"testing"

	"github.com/wudi/hey/opcodes"
	"github.com/wudi/hey/values"
)

func TestArithmeticExecutor_Addition(t *testing.T) {
	ctx := NewExecutionContext()
	frame := newCallFrame("test", nil, nil, []*values.Value{
		values.NewInt(10),
		values.NewInt(20),
	})

	inst := &opcodes.Instruction{
		Opcode:  opcodes.OP_ADD,
		OpType1: (byte(opcodes.IS_CONST) << 4) | byte(opcodes.IS_CONST),
		Op1:     0,
		Op2:     1,
		OpType2: (byte(opcodes.IS_TMP_VAR) << 4),
		Result:  0,
	}

	executor := NewArithmeticExecutor(ctx, frame, inst)
	execResult, err := executor.Execute()

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !execResult.ShouldAdvanceIP {
		t.Errorf("Execute() ShouldAdvanceIP = false, want true")
	}

	// Check that result was written to temp var
	tempResult := frame.getTemp(0)
	if !tempResult.Equal(values.NewInt(30)) {
		t.Errorf("Addition result = %v, want %v", tempResult, values.NewInt(30))
	}
}

func TestArithmeticExecutor_FloatAddition(t *testing.T) {
	ctx := NewExecutionContext()
	frame := newCallFrame("test", nil, nil, []*values.Value{
		values.NewFloat(1.5),
		values.NewFloat(2.5),
	})

	inst := &opcodes.Instruction{
		Opcode:  opcodes.OP_ADD,
		OpType1: (byte(opcodes.IS_CONST) << 4) | byte(opcodes.IS_CONST),
		Op1:     0,
		Op2:     1,
		OpType2: (byte(opcodes.IS_TMP_VAR) << 4),
		Result:  1,
	}

	executor := NewArithmeticExecutor(ctx, frame, inst)
	_, err := executor.Execute()

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	tempResult := frame.getTemp(1)
	if !tempResult.Equal(values.NewFloat(4.0)) {
		t.Errorf("Float addition result = %v, want %v", tempResult, values.NewFloat(4.0))
	}
}

func TestArithmeticExecutor_Subtraction(t *testing.T) {
	ctx := NewExecutionContext()
	frame := newCallFrame("test", nil, nil, []*values.Value{
		values.NewInt(30),
		values.NewInt(12),
	})

	inst := &opcodes.Instruction{
		Opcode:  opcodes.OP_SUB,
		OpType1: (byte(opcodes.IS_CONST) << 4) | byte(opcodes.IS_CONST),
		Op1:     0,
		Op2:     1,
		OpType2: (byte(opcodes.IS_TMP_VAR) << 4),
		Result:  2,
	}

	executor := NewArithmeticExecutor(ctx, frame, inst)
	_, err := executor.Execute()

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	tempResult := frame.getTemp(2)
	if !tempResult.Equal(values.NewInt(18)) {
		t.Errorf("Subtraction result = %v, want %v", tempResult, values.NewInt(18))
	}
}

func TestArithmeticExecutor_Multiplication(t *testing.T) {
	ctx := NewExecutionContext()
	frame := newCallFrame("test", nil, nil, []*values.Value{
		values.NewInt(6),
		values.NewInt(7),
	})

	inst := &opcodes.Instruction{
		Opcode:  opcodes.OP_MUL,
		OpType1: (byte(opcodes.IS_CONST) << 4) | byte(opcodes.IS_CONST),
		Op1:     0,
		Op2:     1,
		OpType2: (byte(opcodes.IS_TMP_VAR) << 4),
		Result:  3,
	}

	executor := NewArithmeticExecutor(ctx, frame, inst)
	_, err := executor.Execute()

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	tempResult := frame.getTemp(3)
	if !tempResult.Equal(values.NewInt(42)) {
		t.Errorf("Multiplication result = %v, want %v", tempResult, values.NewInt(42))
	}
}

func TestArithmeticExecutor_Division(t *testing.T) {
	ctx := NewExecutionContext()
	frame := newCallFrame("test", nil, nil, []*values.Value{
		values.NewFloat(15.0),
		values.NewFloat(3.0),
	})

	inst := &opcodes.Instruction{
		Opcode:  opcodes.OP_DIV,
		OpType1: (byte(opcodes.IS_CONST) << 4) | byte(opcodes.IS_CONST),
		Op1:     0,
		Op2:     1,
		OpType2: (byte(opcodes.IS_TMP_VAR) << 4),
		Result:  4,
	}

	executor := NewArithmeticExecutor(ctx, frame, inst)
	_, err := executor.Execute()

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	tempResult := frame.getTemp(4)
	if !tempResult.Equal(values.NewFloat(5.0)) {
		t.Errorf("Division result = %v, want %v", tempResult, values.NewFloat(5.0))
	}
}

func TestArithmeticExecutor_DivisionByZero(t *testing.T) {
	ctx := NewExecutionContext()
	frame := newCallFrame("test", nil, nil, []*values.Value{
		values.NewInt(10),
		values.NewInt(0),
	})

	inst := &opcodes.Instruction{
		Opcode:  opcodes.OP_DIV,
		OpType1: (byte(opcodes.IS_CONST) << 4) | byte(opcodes.IS_CONST),
		Op1:     0,
		Op2:     1,
		OpType2: (byte(opcodes.IS_TMP_VAR) << 4),
		Result:  5,
	}

	executor := NewArithmeticExecutor(ctx, frame, inst)
	_, err := executor.Execute()

	if err == nil {
		t.Errorf("Execute() expected division by zero error, got nil")
	}
}

func TestArithmeticExecutor_Modulo(t *testing.T) {
	ctx := NewExecutionContext()
	frame := newCallFrame("test", nil, nil, []*values.Value{
		values.NewInt(17),
		values.NewInt(5),
	})

	inst := &opcodes.Instruction{
		Opcode:  opcodes.OP_MOD,
		OpType1: (byte(opcodes.IS_CONST) << 4) | byte(opcodes.IS_CONST),
		Op1:     0,
		Op2:     1,
		OpType2: (byte(opcodes.IS_TMP_VAR) << 4),
		Result:  6,
	}

	executor := NewArithmeticExecutor(ctx, frame, inst)
	_, err := executor.Execute()

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	tempResult := frame.getTemp(6)
	if !tempResult.Equal(values.NewInt(2)) {
		t.Errorf("Modulo result = %v, want %v", tempResult, values.NewInt(2))
	}
}

func TestArithmeticExecutor_Power(t *testing.T) {
	ctx := NewExecutionContext()
	frame := newCallFrame("test", nil, nil, []*values.Value{
		values.NewInt(2),
		values.NewInt(3),
	})

	inst := &opcodes.Instruction{
		Opcode:  opcodes.OP_POW,
		OpType1: (byte(opcodes.IS_CONST) << 4) | byte(opcodes.IS_CONST),
		Op1:     0,
		Op2:     1,
		OpType2: (byte(opcodes.IS_TMP_VAR) << 4),
		Result:  7,
	}

	executor := NewArithmeticExecutor(ctx, frame, inst)
	_, err := executor.Execute()

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	tempResult := frame.getTemp(7)
	if !tempResult.Equal(values.NewInt(8)) {
		t.Errorf("Power result = %v, want %v", tempResult, values.NewInt(8))
	}
}

func TestArithmeticExecutor_UnsupportedOperation(t *testing.T) {
	ctx := NewExecutionContext()
	frame := newCallFrame("test", nil, nil, []*values.Value{
		values.NewInt(1),
		values.NewInt(2),
	})

	inst := &opcodes.Instruction{
		Opcode:  opcodes.OP_CONCAT, // Not an arithmetic operation
		OpType1: (byte(opcodes.IS_CONST) << 4) | byte(opcodes.IS_CONST),
		Op1:     0,
		Op2:     1,
		OpType2: (byte(opcodes.IS_TMP_VAR) << 4),
		Result:  8,
	}

	executor := NewArithmeticExecutor(ctx, frame, inst)
	_, err := executor.Execute()

	if err == nil {
		t.Errorf("Execute() expected unsupported operation error, got nil")
	}
}