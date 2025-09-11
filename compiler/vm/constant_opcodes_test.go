package vm

import (
	"testing"

	"github.com/wudi/hey/compiler/opcodes"
	"github.com/wudi/hey/compiler/values"
)

func TestFetchConstantOpcode(t *testing.T) {
	tests := []struct {
		name           string
		constantName   string
		constantValue  *values.Value
		expectedResult *values.Value
	}{
		{
			name:           "fetch existing string constant",
			constantName:   "TEST_CONST",
			constantValue:  values.NewString("hello world"),
			expectedResult: values.NewString("hello world"),
		},
		{
			name:           "fetch existing int constant",
			constantName:   "MAX_SIZE",
			constantValue:  values.NewInt(1000),
			expectedResult: values.NewInt(1000),
		},
		{
			name:           "fetch existing bool constant",
			constantName:   "DEBUG_MODE",
			constantValue:  values.NewBool(true),
			expectedResult: values.NewBool(true),
		},
		{
			name:           "fetch non-existent constant",
			constantName:   "UNDEFINED_CONST",
			constantValue:  nil, // Don't set in constants map
			expectedResult: values.NewNull(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := NewVirtualMachine()
			ctx := NewExecutionContext()

			// Setup constant if it should exist
			if tt.constantValue != nil {
				ctx.GlobalConstants[tt.constantName] = tt.constantValue
			}

			// Store the constant name as a string value
			ctx.Temporaries[1] = values.NewString(tt.constantName)

			// Create FETCH_CONSTANT instruction
			inst := &opcodes.Instruction{
				Opcode: opcodes.OP_FETCH_CONSTANT,
				Op1:    1, // Temporary containing constant name
				Op2:    0, // Unused
				Result: 2, // Result location
			}
			inst.OpType1, inst.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)

			err := vm.executeFetchConstant(ctx, inst)
			if err != nil {
				t.Fatalf("executeFetchConstant failed: %v", err)
			}

			// Check result
			result := ctx.Temporaries[2]
			if result == nil {
				t.Fatal("Expected result value, got nil")
			}

			// Compare result values based on type
			if tt.expectedResult.IsNull() && !result.IsNull() {
				t.Errorf("Expected null result, got %v", result)
			} else if !tt.expectedResult.IsNull() {
				if tt.expectedResult.IsString() {
					if !result.IsString() || result.ToString() != tt.expectedResult.ToString() {
						t.Errorf("Expected string %q, got %v", tt.expectedResult.ToString(), result)
					}
				} else if tt.expectedResult.IsInt() {
					if !result.IsInt() || result.ToInt() != tt.expectedResult.ToInt() {
						t.Errorf("Expected int %d, got %v", tt.expectedResult.ToInt(), result)
					}
				} else if tt.expectedResult.IsBool() {
					if !result.IsBool() || result.ToBool() != tt.expectedResult.ToBool() {
						t.Errorf("Expected bool %t, got %v", tt.expectedResult.ToBool(), result)
					}
				}
			}
		})
	}
}

func TestCoalesceOpcode(t *testing.T) {
	tests := []struct {
		name        string
		leftValue   *values.Value
		rightValue  *values.Value
		expectedVal string
	}{
		{
			name:        "left value non-null string",
			leftValue:   values.NewString("left"),
			rightValue:  values.NewString("right"),
			expectedVal: "left",
		},
		{
			name:        "left value null, right value string",
			leftValue:   values.NewNull(),
			rightValue:  values.NewString("right"),
			expectedVal: "right",
		},
		{
			name:        "left value non-null int",
			leftValue:   values.NewInt(42),
			rightValue:  values.NewInt(99),
			expectedVal: "42",
		},
		{
			name:        "left value null, right value int",
			leftValue:   values.NewNull(),
			rightValue:  values.NewInt(99),
			expectedVal: "99",
		},
		{
			name:        "both values null",
			leftValue:   values.NewNull(),
			rightValue:  values.NewNull(),
			expectedVal: "null",
		},
		{
			name:        "left empty string (non-null)",
			leftValue:   values.NewString(""),
			rightValue:  values.NewString("fallback"),
			expectedVal: "",
		},
		{
			name:        "left zero (non-null)",
			leftValue:   values.NewInt(0),
			rightValue:  values.NewInt(100),
			expectedVal: "0",
		},
		{
			name:        "left false (non-null)",
			leftValue:   values.NewBool(false),
			rightValue:  values.NewBool(true),
			expectedVal: "false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := NewVirtualMachine()
			ctx := NewExecutionContext()

			// Setup operands
			ctx.Temporaries[1] = tt.leftValue
			ctx.Temporaries[2] = tt.rightValue

			// Create COALESCE instruction
			inst := &opcodes.Instruction{
				Opcode: opcodes.OP_COALESCE,
				Op1:    1, // Left operand
				Op2:    2, // Right operand
				Result: 3, // Result location
			}
			inst.OpType1, inst.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR)

			err := vm.executeCoalesce(ctx, inst)
			if err != nil {
				t.Fatalf("executeCoalesce failed: %v", err)
			}

			// Check result
			result := ctx.Temporaries[3]
			if result == nil {
				t.Fatal("Expected result value, got nil")
			}

			var actualVal string
			if result.IsNull() {
				actualVal = "null"
			} else if result.IsString() {
				actualVal = result.ToString()
			} else if result.IsInt() {
				actualVal = result.ToString()
			} else if result.IsBool() {
				if result.ToBool() {
					actualVal = "true"
				} else {
					actualVal = "false"
				}
			} else {
				actualVal = "unknown"
			}

			if actualVal != tt.expectedVal {
				t.Errorf("Expected %q, got %q", tt.expectedVal, actualVal)
			}
		})
	}
}

func TestCoalesceWithUndefinedOperands(t *testing.T) {
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	// Test with undefined left operand (should be treated as null)
	ctx.Temporaries[2] = values.NewString("fallback")

	// Create COALESCE instruction with undefined left operand
	inst := &opcodes.Instruction{
		Opcode: opcodes.OP_COALESCE,
		Op1:    1, // Undefined operand
		Op2:    2, // Right operand
		Result: 3, // Result location
	}
	inst.OpType1, inst.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR)

	err := vm.executeCoalesce(ctx, inst)
	if err != nil {
		t.Fatalf("executeCoalesce failed: %v", err)
	}

	// Should use right operand when left is undefined
	result := ctx.Temporaries[3]
	if result == nil || !result.IsString() || result.ToString() != "fallback" {
		t.Errorf("Expected 'fallback', got %v", result)
	}
}
