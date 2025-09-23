package vm

import (
	"testing"

	"github.com/wudi/hey/opcodes"
	"github.com/wudi/hey/values"
)

// MockExecutor implements InstructionExecutor for testing
type MockExecutor struct {
	*BaseExecutor
	executeFunc func() (*ExecutionResult, error)
}

func (m *MockExecutor) Execute() (*ExecutionResult, error) {
	if m.executeFunc != nil {
		return m.executeFunc()
	}
	return m.CreateAdvanceResult(values.NewString("mock_result"))
}

func TestBaseExecutor_ReadOperands(t *testing.T) {
	ctx := NewExecutionContext()
	frame := newCallFrame("test", nil, nil, []*values.Value{
		values.NewString("const1"),
		values.NewInt(42),
	})

	// Set up instruction with constant operands
	inst := &opcodes.Instruction{
		OpType1: (byte(opcodes.IS_CONST) << 4) | byte(opcodes.IS_CONST),
		Op1:     0,
		Op2:     1,
	}

	executor := NewBaseExecutor(ctx, frame, inst)

	// Test ReadOperand1
	op1, err := executor.ReadOperand1()
	if err != nil {
		t.Fatalf("ReadOperand1() error = %v", err)
	}
	if !op1.Equal(values.NewString("const1")) {
		t.Errorf("ReadOperand1() = %v, want %v", op1, values.NewString("const1"))
	}

	// Test ReadOperand2
	op2, err := executor.ReadOperand2()
	if err != nil {
		t.Fatalf("ReadOperand2() error = %v", err)
	}
	if !op2.Equal(values.NewInt(42)) {
		t.Errorf("ReadOperand2() = %v, want %v", op2, values.NewInt(42))
	}

	// Test ReadBothOperands
	both1, both2, err := executor.ReadBothOperands()
	if err != nil {
		t.Fatalf("ReadBothOperands() error = %v", err)
	}
	if !both1.Equal(values.NewString("const1")) {
		t.Errorf("ReadBothOperands() op1 = %v, want %v", both1, values.NewString("const1"))
	}
	if !both2.Equal(values.NewInt(42)) {
		t.Errorf("ReadBothOperands() op2 = %v, want %v", both2, values.NewInt(42))
	}
}

func TestBaseExecutor_WriteResult(t *testing.T) {
	ctx := NewExecutionContext()
	frame := newCallFrame("test", nil, nil, nil)

	// Set up instruction to write to temp var
	inst := &opcodes.Instruction{
		OpType2: (byte(opcodes.IS_TMP_VAR) << 4),
		Result:  5,
	}

	executor := NewBaseExecutor(ctx, frame, inst)

	result := values.NewString("test_result")
	err := executor.WriteResult(result)
	if err != nil {
		t.Fatalf("WriteResult() error = %v", err)
	}

	// Verify result was written
	if !frame.getTemp(5).Equal(result) {
		t.Errorf("WriteResult() did not write correctly")
	}
}

func TestBaseExecutor_CreateResults(t *testing.T) {
	ctx := NewExecutionContext()
	frame := newCallFrame("test", nil, nil, nil)
	inst := &opcodes.Instruction{
		OpType2: (byte(opcodes.IS_TMP_VAR) << 4),
		Result:  0,
	}

	executor := NewBaseExecutor(ctx, frame, inst)

	// Test CreateAdvanceResult
	result := values.NewString("advance_result")
	execResult, err := executor.CreateAdvanceResult(result)
	if err != nil {
		t.Fatalf("CreateAdvanceResult() error = %v", err)
	}
	if !execResult.ShouldAdvanceIP {
		t.Errorf("CreateAdvanceResult() ShouldAdvanceIP = false, want true")
	}
	if execResult.JumpTo != -1 {
		t.Errorf("CreateAdvanceResult() JumpTo = %d, want -1", execResult.JumpTo)
	}

	// Test CreateJumpResult
	execResult, err = executor.CreateJumpResult(100)
	if err != nil {
		t.Fatalf("CreateJumpResult() error = %v", err)
	}
	if execResult.ShouldAdvanceIP {
		t.Errorf("CreateJumpResult() ShouldAdvanceIP = true, want false")
	}
	if execResult.JumpTo != 100 {
		t.Errorf("CreateJumpResult() JumpTo = %d, want 100", execResult.JumpTo)
	}

	// Test CreateNoAdvanceResult
	execResult, err = executor.CreateNoAdvanceResult()
	if err != nil {
		t.Fatalf("CreateNoAdvanceResult() error = %v", err)
	}
	if execResult.ShouldAdvanceIP {
		t.Errorf("CreateNoAdvanceResult() ShouldAdvanceIP = true, want false")
	}
	if execResult.JumpTo != -1 {
		t.Errorf("CreateNoAdvanceResult() JumpTo = %d, want -1", execResult.JumpTo)
	}
}

func TestMockExecutor_Execute(t *testing.T) {
	ctx := NewExecutionContext()
	frame := newCallFrame("test", nil, nil, nil)
	inst := &opcodes.Instruction{}

	baseExecutor := NewBaseExecutor(ctx, frame, inst)
	mockExecutor := &MockExecutor{BaseExecutor: baseExecutor}

	// Test default behavior
	result, err := mockExecutor.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !result.ShouldAdvanceIP {
		t.Errorf("Execute() ShouldAdvanceIP = false, want true")
	}

	// Test custom execute function
	called := false
	mockExecutor.executeFunc = func() (*ExecutionResult, error) {
		called = true
		return mockExecutor.CreateJumpResult(42)
	}

	result, err = mockExecutor.Execute()
	if err != nil {
		t.Fatalf("Execute() with custom func error = %v", err)
	}
	if !called {
		t.Errorf("Execute() custom function was not called")
	}
	if result.JumpTo != 42 {
		t.Errorf("Execute() JumpTo = %d, want 42", result.JumpTo)
	}
}

func TestBaseExecutor_Getters(t *testing.T) {
	ctx := NewExecutionContext()
	frame := newCallFrame("test", nil, nil, nil)
	inst := &opcodes.Instruction{Opcode: opcodes.OP_ADD}

	executor := NewBaseExecutor(ctx, frame, inst)

	if executor.GetContext() != ctx {
		t.Errorf("GetContext() did not return the correct context")
	}
	if executor.GetFrame() != frame {
		t.Errorf("GetFrame() did not return the correct frame")
	}
	if executor.GetInstruction() != inst {
		t.Errorf("GetInstruction() did not return the correct instruction")
	}
}