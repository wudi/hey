package vm

import (
	"testing"

	"github.com/wudi/php-parser/compiler/opcodes"
	"github.com/wudi/php-parser/compiler/values"
)

func TestSilenceOpcodes(t *testing.T) {
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	// Initially, no error suppression should be active
	if ctx.IsSilenced() {
		t.Error("Expected no error suppression initially")
	}

	// Create BEGIN_SILENCE instruction
	op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_UNUSED, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)
	beginInst := opcodes.Instruction{
		Opcode:  opcodes.OP_BEGIN_SILENCE,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     0,
		Op2:     0,
		Result:  0, // Store result in temporary variable 0
	}

	// Set up temporaries
	ctx.Temporaries = make(map[uint32]*values.Value)

	// Execute BEGIN_SILENCE
	err := vm.executeBeginSilence(ctx, &beginInst)
	if err != nil {
		t.Fatalf("BEGIN_SILENCE execution failed: %v", err)
	}

	// After BEGIN_SILENCE, error suppression should be active
	if !ctx.IsSilenced() {
		t.Error("Expected error suppression to be active after BEGIN_SILENCE")
	}

	// Check that the result was stored
	result := ctx.Temporaries[0]
	if result == nil {
		t.Fatal("BEGIN_SILENCE result is nil")
	}

	if result.Type != values.TypeBool || !result.Data.(bool) {
		t.Error("Expected BEGIN_SILENCE to return true")
	}

	// Create END_SILENCE instruction
	endInst := opcodes.Instruction{
		Opcode:  opcodes.OP_END_SILENCE,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     0, // Use the result from BEGIN_SILENCE
		Op2:     0,
		Result:  1, // Store result in temporary variable 1
	}

	// Execute END_SILENCE
	err = vm.executeEndSilence(ctx, &endInst)
	if err != nil {
		t.Fatalf("END_SILENCE execution failed: %v", err)
	}

	// After END_SILENCE, error suppression should be inactive
	if ctx.IsSilenced() {
		t.Error("Expected error suppression to be inactive after END_SILENCE")
	}

	// Check that the result was stored
	result = ctx.Temporaries[1]
	if result == nil {
		t.Fatal("END_SILENCE result is nil")
	}

	if result.Type != values.TypeBool || result.Data.(bool) {
		t.Error("Expected END_SILENCE to return false")
	}
}

func TestNestedSilence(t *testing.T) {
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	ctx.Temporaries = make(map[uint32]*values.Value)

	// Create instructions
	op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_UNUSED, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)

	beginInst1 := opcodes.Instruction{
		Opcode:  opcodes.OP_BEGIN_SILENCE,
		OpType1: op1Type,
		OpType2: op2Type,
		Result:  0,
	}

	beginInst2 := opcodes.Instruction{
		Opcode:  opcodes.OP_BEGIN_SILENCE,
		OpType1: op1Type,
		OpType2: op2Type,
		Result:  1,
	}

	endInst1 := opcodes.Instruction{
		Opcode:  opcodes.OP_END_SILENCE,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     1, // Use result from second BEGIN_SILENCE
		Result:  2,
	}

	endInst2 := opcodes.Instruction{
		Opcode:  opcodes.OP_END_SILENCE,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     0, // Use result from first BEGIN_SILENCE
		Result:  3,
	}

	// Execute nested BEGIN_SILENCE operations
	err := vm.executeBeginSilence(ctx, &beginInst1)
	if err != nil {
		t.Fatalf("First BEGIN_SILENCE failed: %v", err)
	}

	if !ctx.IsSilenced() {
		t.Error("Expected silenced after first BEGIN_SILENCE")
	}

	err = vm.executeBeginSilence(ctx, &beginInst2)
	if err != nil {
		t.Fatalf("Second BEGIN_SILENCE failed: %v", err)
	}

	if !ctx.IsSilenced() {
		t.Error("Expected silenced after second BEGIN_SILENCE")
	}

	if len(ctx.SilenceStack) != 2 {
		t.Errorf("Expected silence stack length 2, got %d", len(ctx.SilenceStack))
	}

	// Execute first END_SILENCE (should still be silenced)
	err = vm.executeEndSilence(ctx, &endInst1)
	if err != nil {
		t.Fatalf("First END_SILENCE failed: %v", err)
	}

	if !ctx.IsSilenced() {
		t.Error("Expected to still be silenced after first END_SILENCE")
	}

	if len(ctx.SilenceStack) != 1 {
		t.Errorf("Expected silence stack length 1, got %d", len(ctx.SilenceStack))
	}

	// Execute second END_SILENCE (should no longer be silenced)
	err = vm.executeEndSilence(ctx, &endInst2)
	if err != nil {
		t.Fatalf("Second END_SILENCE failed: %v", err)
	}

	if ctx.IsSilenced() {
		t.Error("Expected not to be silenced after second END_SILENCE")
	}

	if len(ctx.SilenceStack) != 0 {
		t.Errorf("Expected silence stack length 0, got %d", len(ctx.SilenceStack))
	}
}

// Test the @ operator simulation
func TestErrorSuppressionSimulation(t *testing.T) {
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	// Simulate: @some_operation()
	// This would involve: BEGIN_SILENCE, some_operation, END_SILENCE

	ctx.Temporaries = make(map[uint32]*values.Value)

	// 1. BEGIN_SILENCE
	op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_UNUSED, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)
	beginInst := opcodes.Instruction{
		Opcode:  opcodes.OP_BEGIN_SILENCE,
		OpType1: op1Type,
		OpType2: op2Type,
		Result:  0,
	}

	err := vm.executeBeginSilence(ctx, &beginInst)
	if err != nil {
		t.Fatalf("BEGIN_SILENCE failed: %v", err)
	}

	// At this point, errors should be suppressed
	if !ctx.IsSilenced() {
		t.Error("Expected errors to be suppressed during @ operation")
	}

	// 2. Simulate some operation that might generate errors
	// (In real implementation, any errors during this phase would be suppressed)

	// 3. END_SILENCE
	endInst := opcodes.Instruction{
		Opcode:  opcodes.OP_END_SILENCE,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     0, // Use result from BEGIN_SILENCE
		Result:  1,
	}

	err = vm.executeEndSilence(ctx, &endInst)
	if err != nil {
		t.Fatalf("END_SILENCE failed: %v", err)
	}

	// After the @ operation, errors should no longer be suppressed
	if ctx.IsSilenced() {
		t.Error("Expected errors to no longer be suppressed after @ operation")
	}
}
