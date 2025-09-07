package vm

import (
	"testing"

	"github.com/wudi/php-parser/compiler/opcodes"
	"github.com/wudi/php-parser/compiler/values"
)

func TestSendReferenceOpcode(t *testing.T) {
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	// Setup a variable to send as reference
	testValue := values.NewString("test value")
	ctx.Variables[1] = testValue

	// Create SEND_REF instruction
	inst := &opcodes.Instruction{
		Opcode: opcodes.OP_SEND_REF,
		Op1:    1, // Variable to send as reference
		Op2:    0, // Unused
		Result: 0, // Unused
	}
	inst.OpType1, inst.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_VAR, opcodes.IS_UNUSED, opcodes.IS_UNUSED)

	// Clear call arguments before test
	ctx.CallArguments = nil

	err := vm.executeSendReference(ctx, inst)
	if err != nil {
		t.Fatalf("executeSendReference failed: %v", err)
	}

	// Check that argument was added to call arguments
	if len(ctx.CallArguments) != 1 {
		t.Fatalf("Expected 1 call argument, got %d", len(ctx.CallArguments))
	}

	if ctx.CallArguments[0] != testValue {
		t.Errorf("Expected reference to original value, got different value")
	}
}

func TestSendVariableOpcode(t *testing.T) {
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	// Setup a variable to send
	testValue := values.NewInt(42)
	ctx.Variables[1] = testValue

	// Create SEND_VAR instruction
	inst := &opcodes.Instruction{
		Opcode: opcodes.OP_SEND_VAR,
		Op1:    1, // Variable to send
		Op2:    0, // Unused
		Result: 0, // Unused
	}
	inst.OpType1, inst.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_VAR, opcodes.IS_UNUSED, opcodes.IS_UNUSED)

	// Clear call arguments before test
	ctx.CallArguments = nil

	err := vm.executeSendVariable(ctx, inst)
	if err != nil {
		t.Fatalf("executeSendVariable failed: %v", err)
	}

	// Check that argument was added to call arguments
	if len(ctx.CallArguments) != 1 {
		t.Fatalf("Expected 1 call argument, got %d", len(ctx.CallArguments))
	}

	if ctx.CallArguments[0].ToInt() != 42 {
		t.Errorf("Expected argument value 42, got %d", ctx.CallArguments[0].ToInt())
	}
}

func TestUnsetVarOpcode(t *testing.T) {
	tests := []struct {
		name     string
		varType  opcodes.OpType
		setupVar func(ctx *ExecutionContext, slot uint32)
		checkVar func(ctx *ExecutionContext, slot uint32) bool
	}{
		{
			name:    "unset regular variable",
			varType: opcodes.IS_VAR,
			setupVar: func(ctx *ExecutionContext, slot uint32) {
				ctx.Variables[slot] = values.NewString("test")
			},
			checkVar: func(ctx *ExecutionContext, slot uint32) bool {
				return ctx.Variables[slot] == nil
			},
		},
		{
			name:    "unset temporary variable",
			varType: opcodes.IS_TMP_VAR,
			setupVar: func(ctx *ExecutionContext, slot uint32) {
				ctx.Temporaries[slot] = values.NewInt(123)
			},
			checkVar: func(ctx *ExecutionContext, slot uint32) bool {
				return ctx.Temporaries[slot] == nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := NewVirtualMachine()
			ctx := NewExecutionContext()
			slot := uint32(5)

			// Setup variable
			tt.setupVar(ctx, slot)

			// Create UNSET_VAR instruction
			inst := &opcodes.Instruction{
				Opcode: opcodes.OP_UNSET_VAR,
				Op1:    slot,
				Op2:    0, // Unused
				Result: 0, // Unused
			}
			inst.OpType1, inst.OpType2 = opcodes.EncodeOpTypes(tt.varType, opcodes.IS_UNUSED, opcodes.IS_UNUSED)

			err := vm.executeUnsetVar(ctx, inst)
			if err != nil {
				t.Fatalf("executeUnsetVar failed: %v", err)
			}

			// Check that variable was unset
			if !tt.checkVar(ctx, slot) {
				t.Error("Variable should be unset after UNSET_VAR")
			}
		})
	}
}

func TestIssetIsEmptyVarOpcode(t *testing.T) {
	tests := []struct {
		name          string
		setupVariable func(ctx *ExecutionContext)
		expectedIsset bool
	}{
		{
			name: "isset on existing variable",
			setupVariable: func(ctx *ExecutionContext) {
				ctx.Variables[1] = values.NewString("hello")
			},
			expectedIsset: true,
		},
		{
			name: "isset on null variable",
			setupVariable: func(ctx *ExecutionContext) {
				ctx.Variables[1] = values.NewNull()
			},
			expectedIsset: false,
		},
		{
			name: "isset on non-existent variable",
			setupVariable: func(ctx *ExecutionContext) {
				// Don't set variable 1
			},
			expectedIsset: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := NewVirtualMachine()
			ctx := NewExecutionContext()

			// Setup variable state
			tt.setupVariable(ctx)

			// Create ISSET_ISEMPTY_VAR instruction
			inst := &opcodes.Instruction{
				Opcode: opcodes.OP_ISSET_ISEMPTY_VAR,
				Op1:    1, // Variable to check
				Op2:    0, // Unused
				Result: 2, // Result location
			}
			inst.OpType1, inst.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_VAR, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)

			err := vm.executeIssetIsEmptyVar(ctx, inst)
			if err != nil {
				t.Fatalf("executeIssetIsEmptyVar failed: %v", err)
			}

			// Check result
			result := ctx.Temporaries[2]
			if result == nil || !result.IsBool() {
				t.Fatalf("Expected boolean result, got %v", result)
			}

			if result.ToBool() != tt.expectedIsset {
				t.Errorf("Expected isset=%v, got isset=%v", tt.expectedIsset, result.ToBool())
			}
		})
	}
}

func TestDoInternalCallOpcode(t *testing.T) {
	// This is a basic test to ensure DO_ICALL doesn't crash
	// The actual functionality is delegated to DO_FCALL
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	// Setup a simple function call context
	ctx.CallContext = &CallContext{
		FunctionName: "strlen",
	}
	ctx.CallArguments = []*values.Value{values.NewString("test")}

	inst := &opcodes.Instruction{
		Opcode: opcodes.OP_DO_ICALL,
		Op1:    0,
		Op2:    0,
		Result: 1,
	}
	inst.OpType1, inst.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_UNUSED, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)

	// This should delegate to DO_FCALL and not crash
	err := vm.executeDoInternalCall(ctx, inst)
	// We expect this might fail due to function not being set up properly,
	// but it shouldn't crash with a panic
	if err != nil {
		t.Logf("DO_ICALL failed as expected (no function setup): %v", err)
	}
}

func TestDoUserCallOpcode(t *testing.T) {
	// Similar basic test for DO_UCALL
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	// Setup a simple function call context
	ctx.CallContext = &CallContext{
		FunctionName: "user_function",
	}

	inst := &opcodes.Instruction{
		Opcode: opcodes.OP_DO_UCALL,
		Op1:    0,
		Op2:    0,
		Result: 1,
	}
	inst.OpType1, inst.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_UNUSED, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)

	// This should delegate to DO_FCALL and not crash
	err := vm.executeDoUserCall(ctx, inst)
	// We expect this might fail due to function not being set up properly,
	// but it shouldn't crash with a panic
	if err != nil {
		t.Logf("DO_UCALL failed as expected (no function setup): %v", err)
	}
}
