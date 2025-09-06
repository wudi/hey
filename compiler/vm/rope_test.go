package vm

import (
	"testing"

	"github.com/wudi/php-parser/compiler/opcodes"
	"github.com/wudi/php-parser/compiler/values"
)

func TestRopeBasicConcatenation(t *testing.T) {
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	// Test ROPE operations simulating: "Hello" . " " . "World" . "!"
	ctx.Temporaries = make(map[uint32]*values.Value)
	ctx.Temporaries[0] = values.NewString("Hello")
	ctx.Temporaries[1] = values.NewString(" ")
	ctx.Temporaries[2] = values.NewString("World")
	ctx.Temporaries[3] = values.NewString("!")

	// ROPE_INIT: Start with "Hello"
	op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_UNUSED)
	initInst := opcodes.Instruction{
		Opcode:  opcodes.OP_ROPE_INIT,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     0,  // "Hello" from temp 0
		Result:  10, // Buffer ID 10
	}

	err := vm.executeRopeInit(ctx, &initInst)
	if err != nil {
		t.Fatalf("ROPE_INIT failed: %v", err)
	}

	// Check buffer was created
	if buffer, exists := ctx.RopeBuffers[10]; !exists || len(buffer) != 1 || buffer[0] != "Hello" {
		t.Errorf("ROPE_INIT buffer incorrect: %v", buffer)
	}

	// ROPE_ADD: Add " "
	op1TypeAdd, op2TypeAdd := opcodes.EncodeOpTypes(opcodes.IS_UNUSED, opcodes.IS_TMP_VAR, opcodes.IS_UNUSED)
	addInst1 := opcodes.Instruction{
		Opcode:  opcodes.OP_ROPE_ADD,
		OpType1: op1TypeAdd,
		OpType2: op2TypeAdd,
		Op1:     10, // Buffer ID
		Op2:     1,  // " " from temp 1
	}

	err = vm.executeRopeAdd(ctx, &addInst1)
	if err != nil {
		t.Fatalf("ROPE_ADD failed: %v", err)
	}

	// ROPE_ADD: Add "World"
	addInst2 := opcodes.Instruction{
		Opcode:  opcodes.OP_ROPE_ADD,
		OpType1: op1TypeAdd,
		OpType2: op2TypeAdd,
		Op1:     10, // Buffer ID
		Op2:     2,  // "World" from temp 2
	}

	err = vm.executeRopeAdd(ctx, &addInst2)
	if err != nil {
		t.Fatalf("ROPE_ADD failed: %v", err)
	}

	// ROPE_ADD: Add "!"
	addInst3 := opcodes.Instruction{
		Opcode:  opcodes.OP_ROPE_ADD,
		OpType1: op1TypeAdd,
		OpType2: op2TypeAdd,
		Op1:     10, // Buffer ID
		Op2:     3,  // "!" from temp 3
	}

	err = vm.executeRopeAdd(ctx, &addInst3)
	if err != nil {
		t.Fatalf("ROPE_ADD failed: %v", err)
	}

	// Check buffer has all strings
	if buffer, exists := ctx.RopeBuffers[10]; !exists || len(buffer) != 4 {
		t.Errorf("ROPE buffer should have 4 strings, got: %v", buffer)
	}

	// ROPE_END: Finalize concatenation
	op1TypeEnd, op2TypeEnd := opcodes.EncodeOpTypes(opcodes.IS_UNUSED, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)
	endInst := opcodes.Instruction{
		Opcode:  opcodes.OP_ROPE_END,
		OpType1: op1TypeEnd,
		OpType2: op2TypeEnd,
		Op1:     10, // Buffer ID
		Result:  5,  // Store result in temp 5
	}

	err = vm.executeRopeEnd(ctx, &endInst)
	if err != nil {
		t.Fatalf("ROPE_END failed: %v", err)
	}

	// Check result
	result := ctx.Temporaries[5]
	if result == nil {
		t.Fatal("ROPE_END result is nil")
	}

	expected := "Hello World!"
	if result.ToString() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result.ToString())
	}

	// Check buffer was cleaned up
	if _, exists := ctx.RopeBuffers[10]; exists {
		t.Error("ROPE buffer should be cleaned up after ROPE_END")
	}
}

func TestRopeEmptyStrings(t *testing.T) {
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	ctx.Temporaries = make(map[uint32]*values.Value)
	ctx.Temporaries[0] = values.NewString("")
	ctx.Temporaries[1] = values.NewString("test")
	ctx.Temporaries[2] = values.NewString("")

	// ROPE_INIT with empty string
	op1TypeInit, op2TypeInit := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_UNUSED)
	initInst := opcodes.Instruction{
		Opcode:  opcodes.OP_ROPE_INIT,
		OpType1: op1TypeInit,
		OpType2: op2TypeInit,
		Op1:     0,
		Result:  1,
	}

	err := vm.executeRopeInit(ctx, &initInst)
	if err != nil {
		t.Fatalf("ROPE_INIT failed: %v", err)
	}

	// Add non-empty string
	op1TypeAdd2, op2TypeAdd2 := opcodes.EncodeOpTypes(opcodes.IS_UNUSED, opcodes.IS_TMP_VAR, opcodes.IS_UNUSED)
	addInst1 := opcodes.Instruction{
		Opcode:  opcodes.OP_ROPE_ADD,
		OpType1: op1TypeAdd2,
		OpType2: op2TypeAdd2,
		Op1:     1,
		Op2:     1,
	}

	err = vm.executeRopeAdd(ctx, &addInst1)
	if err != nil {
		t.Fatalf("ROPE_ADD failed: %v", err)
	}

	// Add empty string
	addInst2 := opcodes.Instruction{
		Opcode:  opcodes.OP_ROPE_ADD,
		OpType1: op1TypeAdd2,
		OpType2: op2TypeAdd2,
		Op1:     1,
		Op2:     2,
	}

	err = vm.executeRopeAdd(ctx, &addInst2)
	if err != nil {
		t.Fatalf("ROPE_ADD failed: %v", err)
	}

	// Finalize
	op1TypeEnd2, op2TypeEnd2 := opcodes.EncodeOpTypes(opcodes.IS_UNUSED, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)
	endInst := opcodes.Instruction{
		Opcode:  opcodes.OP_ROPE_END,
		OpType1: op1TypeEnd2,
		OpType2: op2TypeEnd2,
		Op1:     1,
		Result:  3,
	}

	err = vm.executeRopeEnd(ctx, &endInst)
	if err != nil {
		t.Fatalf("ROPE_END failed: %v", err)
	}

	result := ctx.Temporaries[3]
	if result.ToString() != "test" {
		t.Errorf("Expected 'test', got '%s'", result.ToString())
	}
}

func TestRopeEndWithoutInit(t *testing.T) {
	// Test ROPE_END with non-existent buffer (should handle gracefully)
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	ctx.Temporaries = make(map[uint32]*values.Value)

	op1TypeEnd3, op2TypeEnd3 := opcodes.EncodeOpTypes(opcodes.IS_UNUSED, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)
	endInst := opcodes.Instruction{
		Opcode:  opcodes.OP_ROPE_END,
		OpType1: op1TypeEnd3,
		OpType2: op2TypeEnd3,
		Op1:     99, // Non-existent buffer
		Result:  1,
	}

	err := vm.executeRopeEnd(ctx, &endInst)
	if err != nil {
		t.Fatalf("ROPE_END failed: %v", err)
	}

	result := ctx.Temporaries[1]
	if result.ToString() != "" {
		t.Errorf("Expected empty string, got '%s'", result.ToString())
	}
}

func TestFastConcatBasic(t *testing.T) {
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	ctx.Temporaries = make(map[uint32]*values.Value)
	ctx.Temporaries[0] = values.NewString("Hello")
	ctx.Temporaries[1] = values.NewString(" World")

	op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR)
	inst := opcodes.Instruction{
		Opcode:  opcodes.OP_FAST_CONCAT,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     0,
		Op2:     1,
		Result:  2,
	}

	err := vm.executeFastConcat(ctx, &inst)
	if err != nil {
		t.Fatalf("FAST_CONCAT failed: %v", err)
	}

	result := ctx.Temporaries[2]
	if result == nil {
		t.Fatal("FAST_CONCAT result is nil")
	}

	expected := "Hello World"
	if result.ToString() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result.ToString())
	}
}

func TestFastConcatWithNumbers(t *testing.T) {
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	ctx.Temporaries = make(map[uint32]*values.Value)
	ctx.Temporaries[0] = values.NewInt(42)
	ctx.Temporaries[1] = values.NewFloat(3.14)

	op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR)
	inst := opcodes.Instruction{
		Opcode:  opcodes.OP_FAST_CONCAT,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     0,
		Op2:     1,
		Result:  2,
	}

	err := vm.executeFastConcat(ctx, &inst)
	if err != nil {
		t.Fatalf("FAST_CONCAT failed: %v", err)
	}

	result := ctx.Temporaries[2]
	expected := "423.14"
	if result.ToString() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result.ToString())
	}
}
