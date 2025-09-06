package vm

import (
	"testing"

	"github.com/wudi/php-parser/compiler/opcodes"
	"github.com/wudi/php-parser/compiler/values"
)

func TestCaseOpcode(t *testing.T) {
	tests := []struct {
		name     string
		op1      *values.Value
		op2      *values.Value
		expected bool
		opcode   opcodes.Opcode
	}{
		{
			name:     "CASE: string '2' == int 2 (loose comparison)",
			op1:      values.NewString("2"),
			op2:      values.NewInt(2),
			expected: true,
			opcode:   opcodes.OP_CASE,
		},
		{
			name:     "CASE: int 2 == string '2' (loose comparison)",
			op1:      values.NewInt(2),
			op2:      values.NewString("2"),
			expected: true,
			opcode:   opcodes.OP_CASE,
		},
		{
			name:     "CASE: int 1 != int 2",
			op1:      values.NewInt(1),
			op2:      values.NewInt(2),
			expected: false,
			opcode:   opcodes.OP_CASE,
		},
		{
			name:     "CASE_STRICT: string '2' !== int 2 (strict comparison)",
			op1:      values.NewString("2"),
			op2:      values.NewInt(2),
			expected: false,
			opcode:   opcodes.OP_CASE_STRICT,
		},
		{
			name:     "CASE_STRICT: int 2 === int 2 (strict comparison)",
			op1:      values.NewInt(2),
			op2:      values.NewInt(2),
			expected: true,
			opcode:   opcodes.OP_CASE_STRICT,
		},
		{
			name:     "CASE_STRICT: string 'test' === string 'test'",
			op1:      values.NewString("test"),
			op2:      values.NewString("test"),
			expected: true,
			opcode:   opcodes.OP_CASE_STRICT,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vm := NewVirtualMachine()
			ctx := NewExecutionContext()

			// Create instruction
			op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR)
			inst := opcodes.Instruction{
				Opcode:  test.opcode,
				OpType1: op1Type,
				OpType2: op2Type,
				Op1:     0, // temporary variable slot 0
				Op2:     1, // temporary variable slot 1
				Result:  2, // temporary variable slot 2
			}

			// Set up operands in temporary variables
			ctx.Temporaries = make(map[uint32]*values.Value)
			ctx.Temporaries[0] = test.op1
			ctx.Temporaries[1] = test.op2

			// Execute the instruction
			var err error
			switch test.opcode {
			case opcodes.OP_CASE:
				err = vm.executeCase(ctx, &inst)
			case opcodes.OP_CASE_STRICT:
				err = vm.executeCaseStrict(ctx, &inst)
			}

			if err != nil {
				t.Fatalf("Execution failed: %v", err)
			}

			// Check result
			result := ctx.Temporaries[2]
			if result == nil {
				t.Fatal("Result is nil")
			}

			if result.Type != values.TypeBool {
				t.Errorf("Expected bool result, got %v", result.Type)
			}

			actualResult := result.Data.(bool)
			if actualResult != test.expected {
				t.Errorf("Expected %v, got %v", test.expected, actualResult)
			}
		})
	}
}

// Test switch statement simulation using CASE opcodes
func TestSwitchStatementSimulation(t *testing.T) {
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	// Simulate: switch("2") { case 2: ... }
	switchValue := values.NewString("2")
	caseValue := values.NewInt(2)

	ctx.Temporaries = make(map[uint32]*values.Value)
	ctx.Temporaries[0] = switchValue
	ctx.Temporaries[1] = caseValue

	// Create CASE instruction (loose comparison)
	op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR)
	inst := opcodes.Instruction{
		Opcode:  opcodes.OP_CASE,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     0,
		Op2:     1,
		Result:  2,
	}

	// Execute case comparison
	err := vm.executeCase(ctx, &inst)
	if err != nil {
		t.Fatalf("CASE execution failed: %v", err)
	}

	// Check result - should be true due to loose comparison
	result := ctx.Temporaries[2]
	if result == nil {
		t.Fatal("CASE result is nil")
	}

	if result.Type != values.TypeBool || !result.Data.(bool) {
		t.Error("Expected CASE to return true for loose comparison of '2' and 2")
	}

	// Test with strict comparison - should be false
	inst.Opcode = opcodes.OP_CASE_STRICT
	err = vm.executeCaseStrict(ctx, &inst)
	if err != nil {
		t.Fatalf("CASE_STRICT execution failed: %v", err)
	}

	result = ctx.Temporaries[2]
	if result == nil {
		t.Fatal("CASE_STRICT result is nil")
	}

	if result.Type != values.TypeBool || result.Data.(bool) {
		t.Error("Expected CASE_STRICT to return false for strict comparison of '2' and 2")
	}
}
