package vm

import (
	"testing"

	"github.com/wudi/hey/opcodes"
	"github.com/wudi/hey/values"
)

func TestInstructionFactory_CreateExecutor(t *testing.T) {
	factory := NewInstructionFactory()
	ctx := NewExecutionContext()
	frame := &CallFrame{
		Constants: []*values.Value{
			values.NewInt(10),
			values.NewInt(5),
		},
		Locals:   make(map[uint32]*values.Value),
		TempVars: make(map[uint32]*values.Value),
	}

	tests := []struct {
		name    string
		opcode  opcodes.Opcode
		wantErr bool
	}{
		{
			name:    "arithmetic - ADD",
			opcode:  opcodes.OP_ADD,
			wantErr: false,
		},
		{
			name:    "arithmetic - SUB",
			opcode:  opcodes.OP_SUB,
			wantErr: false,
		},
		{
			name:    "arithmetic - MUL",
			opcode:  opcodes.OP_MUL,
			wantErr: false,
		},
		{
			name:    "arithmetic - DIV",
			opcode:  opcodes.OP_DIV,
			wantErr: false,
		},
		{
			name:    "comparison - IS_EQUAL",
			opcode:  opcodes.OP_IS_EQUAL,
			wantErr: false,
		},
		{
			name:    "comparison - IS_IDENTICAL",
			opcode:  opcodes.OP_IS_IDENTICAL,
			wantErr: false,
		},
		{
			name:    "variable - ASSIGN",
			opcode:  opcodes.OP_ASSIGN,
			wantErr: false,
		},
		{
			name:    "variable - ASSIGN_REF",
			opcode:  opcodes.OP_ASSIGN_REF,
			wantErr: false,
		},
		{
			name:    "string - CONCAT",
			opcode:  opcodes.OP_CONCAT,
			wantErr: false,
		},
		{
			name:    "special - NOP",
			opcode:  opcodes.OP_NOP,
			wantErr: false,
		},
		{
			name:    "special - EXIT",
			opcode:  opcodes.OP_EXIT,
			wantErr: false,
		},
		{
			name:    "unsupported opcode",
			opcode:  opcodes.OP_PLUS, // Not in factory yet
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inst := &opcodes.Instruction{
				Opcode:  tt.opcode,
				OpType1: byte(opcodes.IS_CONST)<<4 | byte(opcodes.IS_CONST), // Op1=CONST, Op2=CONST
				OpType2: byte(opcodes.IS_TMP_VAR) << 4,                      // Result=TMP_VAR
				Op1:     0,
				Op2:     1,
				Result:  2,
			}

			executor, err := factory.CreateExecutor(ctx, frame, inst)

			if tt.wantErr {
				if err == nil {
					t.Errorf("CreateExecutor() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("CreateExecutor() unexpected error: %v", err)
				return
			}

			if executor == nil {
				t.Errorf("CreateExecutor() returned nil executor")
			}
		})
	}
}

func TestInstructionFactory_IsSupported(t *testing.T) {
	factory := NewInstructionFactory()

	supportedOpcodes := []opcodes.Opcode{
		opcodes.OP_ADD,
		opcodes.OP_SUB,
		opcodes.OP_IS_EQUAL,
		opcodes.OP_ASSIGN,
		opcodes.OP_CONCAT,
		opcodes.OP_NOP,
	}

	unsupportedOpcodes := []opcodes.Opcode{
		opcodes.OP_PLUS,  // Unary operations not yet in factory
		opcodes.OP_MINUS, // Unary operations not yet in factory
	}

	for _, opcode := range supportedOpcodes {
		if !factory.IsSupported(opcode) {
			t.Errorf("IsSupported(%v) should return true", opcode)
		}
	}

	for _, opcode := range unsupportedOpcodes {
		if factory.IsSupported(opcode) {
			t.Errorf("IsSupported(%v) should return false", opcode)
		}
	}
}

func TestInstructionFactory_GetSupportedOpcodes(t *testing.T) {
	factory := NewInstructionFactory()
	supportedOpcodes := factory.GetSupportedOpcodes()

	if len(supportedOpcodes) == 0 {
		t.Error("GetSupportedOpcodes() should return non-empty slice")
	}

	// Verify some expected opcodes are present
	expectedOpcodes := map[opcodes.Opcode]bool{
		opcodes.OP_ADD:      false,
		opcodes.OP_SUB:      false,
		opcodes.OP_IS_EQUAL: false,
		opcodes.OP_ASSIGN:   false,
	}

	for _, opcode := range supportedOpcodes {
		if _, exists := expectedOpcodes[opcode]; exists {
			expectedOpcodes[opcode] = true
		}
	}

	for opcode, found := range expectedOpcodes {
		if !found {
			t.Errorf("Expected opcode %v not found in supported opcodes", opcode)
		}
	}
}

func TestInstructionFactory_Integration(t *testing.T) {
	// Test that the instruction factory integrates properly with the VM
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()
	frame := &CallFrame{
		Constants: []*values.Value{
			values.NewInt(10),
			values.NewInt(5),
		},
		Locals:   make(map[uint32]*values.Value),
		TempVars: make(map[uint32]*values.Value),
	}

	// Test arithmetic operation through factory
	inst := &opcodes.Instruction{
		Opcode:  opcodes.OP_ADD,
		OpType1: byte(opcodes.IS_CONST)<<4 | byte(opcodes.IS_CONST), // Op1=CONST, Op2=CONST
		OpType2: byte(opcodes.IS_TMP_VAR) << 4,                      // Result=TMP_VAR
		Op1:     0,                                                  // First constant (10)
		Op2:     1,                                                  // Second constant (5)
		Result:  2,                                                  // Store in temp 2
	}

	// The VM should use the factory to execute the instruction
	advance, err := vm.executeInstruction(ctx, frame, inst)
	if err != nil {
		t.Fatalf("executeInstruction() failed: %v", err)
	}

	if !advance {
		t.Error("executeInstruction() should return true for advancing IP")
	}

	// Verify the result was computed correctly (10 + 5 = 15)
	result, exists := ctx.Temporaries.Load(uint32(2))
	if !exists {
		t.Fatal("Result not stored in temporaries")
	}

	resultValue := result.(*values.Value)
	if resultValue.ToInt() != 15 {
		t.Errorf("Expected result 15, got %v", resultValue.ToInt())
	}
}