package vm

import (
	"testing"

	"github.com/wudi/hey/compiler/opcodes"
	"github.com/wudi/hey/compiler/values"
)

func TestFetchIssetOpcode(t *testing.T) {
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

			// Create FETCH_IS instruction: FETCH_IS TMP:0, VAR:1 -> TMP:2
			inst := &opcodes.Instruction{
				Opcode: opcodes.OP_FETCH_IS,
				Op1:    1, // Variable to check
				Op2:    0, // Unused for FETCH_IS
				Result: 2, // Result location
			}
			inst.OpType1, inst.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_VAR, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)

			err := vm.executeFetchIsset(ctx, inst)
			if err != nil {
				t.Fatalf("executeFetchIsset failed: %v", err)
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

func TestFetchDimIssetOpcode(t *testing.T) {
	tests := []struct {
		name          string
		setupArray    func() *values.Value
		key           *values.Value
		expectedIsset bool
	}{
		{
			name: "isset on existing array element",
			setupArray: func() *values.Value {
				arr := values.NewArray()
				arr.ArraySet(values.NewString("key"), values.NewString("value"))
				return arr
			},
			key:           values.NewString("key"),
			expectedIsset: true,
		},
		{
			name: "isset on null array element",
			setupArray: func() *values.Value {
				arr := values.NewArray()
				arr.ArraySet(values.NewString("key"), values.NewNull())
				return arr
			},
			key:           values.NewString("key"),
			expectedIsset: false,
		},
		{
			name: "isset on non-existent array element",
			setupArray: func() *values.Value {
				arr := values.NewArray()
				return arr
			},
			key:           values.NewString("missing"),
			expectedIsset: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := NewVirtualMachine()
			ctx := NewExecutionContext()

			// Setup array
			array := tt.setupArray()
			ctx.Variables[1] = array
			ctx.Constants = append(ctx.Constants, tt.key)

			// Create FETCH_DIM_IS instruction
			inst := &opcodes.Instruction{
				Opcode: opcodes.OP_FETCH_DIM_IS,
				Op1:    1, // Array variable
				Op2:    0, // Key constant
				Result: 2, // Result location
			}
			inst.OpType1, inst.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_VAR, opcodes.IS_CONST, opcodes.IS_TMP_VAR)

			err := vm.executeFetchDimIsset(ctx, inst)
			if err != nil {
				t.Fatalf("executeFetchDimIsset failed: %v", err)
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

func TestFetchUnsetOpcode(t *testing.T) {
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	// Setup variable
	ctx.Variables[1] = values.NewString("test")

	// Verify variable exists
	if ctx.Variables[1] == nil {
		t.Fatal("Variable should exist before unset")
	}

	// Create FETCH_UNSET instruction
	inst := &opcodes.Instruction{
		Opcode: opcodes.OP_FETCH_UNSET,
		Op1:    1, // Variable to unset
		Op2:    0, // Unused
		Result: 0, // Unused
	}
	inst.OpType1, inst.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_VAR, opcodes.IS_UNUSED, opcodes.IS_UNUSED)

	err := vm.executeFetchUnset(ctx, inst)
	if err != nil {
		t.Fatalf("executeFetchUnset failed: %v", err)
	}

	// Verify variable is unset
	if ctx.Variables[1] != nil {
		t.Error("Variable should be unset after FETCH_UNSET")
	}
}

func TestFetchDimUnsetOpcode(t *testing.T) {
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	// Setup array with element
	array := values.NewArray()
	key := values.NewString("key")
	array.ArraySet(key, values.NewString("value"))
	ctx.Variables[1] = array
	ctx.Constants = append(ctx.Constants, key)

	// First verify element exists using isset
	issetInst := &opcodes.Instruction{
		Opcode: opcodes.OP_FETCH_DIM_IS,
		Op1:    1, // Array variable
		Op2:    0, // Key constant
		Result: 3, // Result location
	}
	issetInst.OpType1, issetInst.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_VAR, opcodes.IS_CONST, opcodes.IS_TMP_VAR)

	err := vm.executeFetchDimIsset(ctx, issetInst)
	if err != nil {
		t.Fatalf("executeFetchDimIsset failed: %v", err)
	}

	if !ctx.Temporaries[3].ToBool() {
		t.Fatal("Array element should be set before unset")
	}

	// Create FETCH_DIM_UNSET instruction
	unsetInst := &opcodes.Instruction{
		Opcode: opcodes.OP_FETCH_DIM_UNSET,
		Op1:    1, // Array variable
		Op2:    0, // Key constant
		Result: 0, // Unused
	}
	unsetInst.OpType1, unsetInst.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_VAR, opcodes.IS_CONST, opcodes.IS_UNUSED)

	err = vm.executeFetchDimUnset(ctx, unsetInst)
	if err != nil {
		t.Fatalf("executeFetchDimUnset failed: %v", err)
	}

	// Verify element is unset by testing isset again
	issetInst2 := &opcodes.Instruction{
		Opcode: opcodes.OP_FETCH_DIM_IS,
		Op1:    1, // Array variable
		Op2:    0, // Key constant
		Result: 4, // Result location (different from before)
	}
	issetInst2.OpType1, issetInst2.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_VAR, opcodes.IS_CONST, opcodes.IS_TMP_VAR)

	err = vm.executeFetchDimIsset(ctx, issetInst2)
	if err != nil {
		t.Fatalf("executeFetchDimIsset failed after unset: %v", err)
	}

	if ctx.Temporaries[4].ToBool() {
		t.Error("Array element should not be set after FETCH_DIM_UNSET")
	}
}
