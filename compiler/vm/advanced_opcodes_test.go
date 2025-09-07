package vm

import (
	"testing"

	"github.com/wudi/php-parser/compiler/opcodes"
	"github.com/wudi/php-parser/compiler/values"
)

func TestYieldOpcode(t *testing.T) {
	tests := []struct {
		name         string
		yieldValue   *values.Value
		yieldKey     *values.Value
		hasGenerator bool
		expectedHalt bool
	}{
		{
			name:         "yield in generator context",
			yieldValue:   values.NewString("test_value"),
			yieldKey:     values.NewInt(1),
			hasGenerator: true,
			expectedHalt: true,
		},
		{
			name:         "yield outside generator",
			yieldValue:   values.NewInt(42),
			yieldKey:     nil,
			hasGenerator: false,
			expectedHalt: false,
		},
		{
			name:         "yield null value",
			yieldValue:   values.NewNull(),
			yieldKey:     values.NewString("key"),
			hasGenerator: true,
			expectedHalt: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := NewVirtualMachine()
			ctx := NewExecutionContext()

			// Setup yield value
			ctx.Temporaries[1] = tt.yieldValue

			// Setup yield key if provided
			op2 := uint32(0)
			if tt.yieldKey != nil {
				ctx.Temporaries[2] = tt.yieldKey
				op2 = 2
			}

			// Setup generator context if needed
			if tt.hasGenerator {
				ctx.CurrentGenerator = &Generator{
					Function:     nil,
					Context:      ctx,
					Variables:    make(map[uint32]*values.Value),
					IP:           0,
					YieldedKey:   nil,
					YieldedValue: nil,
					IsFinished:   false,
					IsSuspended:  false,
				}
			}

			// Create YIELD instruction
			inst := &opcodes.Instruction{
				Opcode: opcodes.OP_YIELD,
				Op1:    1,   // Yield value
				Op2:    op2, // Yield key (optional)
				Result: 3,   // Result location
			}
			inst.OpType1, inst.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR)

			err := vm.executeYield(ctx, inst)
			if err != nil {
				t.Fatalf("executeYield failed: %v", err)
			}

			// Check execution state
			if ctx.Halted != tt.expectedHalt {
				t.Errorf("Expected halted=%t, got halted=%t", tt.expectedHalt, ctx.Halted)
			}

			// Check generator state if applicable
			if tt.hasGenerator {
				if ctx.CurrentGenerator.YieldedValue == nil {
					t.Error("Generator should have yielded value")
				}
				if !ctx.CurrentGenerator.IsSuspended {
					t.Error("Generator should be suspended")
				}

				// Check yielded value matches
				if !valuesEqual(ctx.CurrentGenerator.YieldedValue, tt.yieldValue) {
					t.Errorf("Expected yielded value %v, got %v", tt.yieldValue, ctx.CurrentGenerator.YieldedValue)
				}
			}

			// Check result
			result := ctx.Temporaries[3]
			if result == nil {
				t.Error("Result should not be nil")
			}
		})
	}
}

func TestAddArrayUnpackOpcode(t *testing.T) {
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	// Create target array [1, 2]
	targetArray := values.NewArray()
	targetData := targetArray.Data.(*values.Array)
	targetData.Elements[0] = values.NewInt(1)
	targetData.Elements[1] = values.NewInt(2)
	targetData.NextIndex = 2

	// Create source array to unpack [3, 4, 5]
	sourceArray := values.NewArray()
	sourceData := sourceArray.Data.(*values.Array)
	sourceData.Elements[0] = values.NewInt(3)
	sourceData.Elements[1] = values.NewInt(4)
	sourceData.Elements[2] = values.NewInt(5)
	sourceData.NextIndex = 3

	// Setup operands
	ctx.Temporaries[1] = sourceArray
	ctx.Temporaries[2] = targetArray

	// Create ADD_ARRAY_UNPACK instruction
	inst := &opcodes.Instruction{
		Opcode: opcodes.OP_ADD_ARRAY_UNPACK,
		Op1:    1, // Source array
		Op2:    0, // Unused
		Result: 2, // Target array
	}
	inst.OpType1, inst.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)

	err := vm.executeAddArrayUnpack(ctx, inst)
	if err != nil {
		t.Fatalf("executeAddArrayUnpack failed: %v", err)
	}

	// Check that target array now contains [1, 2, 3, 4, 5]
	finalData := targetArray.Data.(*values.Array)
	expectedSize := 5
	if len(finalData.Elements) != expectedSize {
		t.Errorf("Expected array size %d, got %d", expectedSize, len(finalData.Elements))
	}

	// Check specific values
	expectedValues := []int64{1, 2, 3, 4, 5}
	for i, expectedVal := range expectedValues {
		if val, exists := finalData.Elements[i]; exists {
			if !val.IsInt() || val.ToInt() != expectedVal {
				t.Errorf("Expected element[%d] = %d, got %v", i, expectedVal, val)
			}
		} else {
			t.Errorf("Missing element at index %d", i)
		}
	}
}

func TestBindGlobalOpcode(t *testing.T) {
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	// Setup global variable
	ctx.GlobalVars["test_var"] = values.NewString("global_value")

	// Setup variable name
	ctx.Temporaries[1] = values.NewString("test_var")

	// Create BIND_GLOBAL instruction
	localSlot := uint32(5)
	inst := &opcodes.Instruction{
		Opcode: opcodes.OP_BIND_GLOBAL,
		Op1:    1,         // Variable name
		Op2:    localSlot, // Local variable slot
		Result: 0,         // Unused
	}
	inst.OpType1, inst.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_UNUSED)

	err := vm.executeBindGlobal(ctx, inst)
	if err != nil {
		t.Fatalf("executeBindGlobal failed: %v", err)
	}

	// Check that local variable is bound to global
	localVar := ctx.Variables[localSlot]
	if localVar == nil {
		t.Fatal("Local variable should be bound")
	}

	if localVar != ctx.GlobalVars["test_var"] {
		t.Error("Local variable should reference the same object as global variable")
	}

	if !localVar.IsString() || localVar.ToString() != "global_value" {
		t.Errorf("Expected local variable value 'global_value', got %v", localVar)
	}

	// Check variable name mapping
	if ctx.VarSlotNames[localSlot] != "test_var" {
		t.Errorf("Expected variable name mapping 'test_var', got %q", ctx.VarSlotNames[localSlot])
	}
}

func TestBindGlobalWithNewVariable(t *testing.T) {
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	// Don't setup global variable - should be created
	ctx.Temporaries[1] = values.NewString("new_var")

	localSlot := uint32(3)
	inst := &opcodes.Instruction{
		Opcode: opcodes.OP_BIND_GLOBAL,
		Op1:    1,
		Op2:    localSlot,
		Result: 0,
	}
	inst.OpType1, inst.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_UNUSED)

	err := vm.executeBindGlobal(ctx, inst)
	if err != nil {
		t.Fatalf("executeBindGlobal failed: %v", err)
	}

	// Check that global variable was created
	globalVar := ctx.GlobalVars["new_var"]
	if globalVar == nil {
		t.Fatal("Global variable should be created")
	}

	if !globalVar.IsNull() {
		t.Error("New global variable should be null")
	}

	// Check that local variable is bound
	localVar := ctx.Variables[localSlot]
	if localVar != globalVar {
		t.Error("Local variable should reference global variable")
	}
}

func TestMatchOpcode(t *testing.T) {
	tests := []struct {
		name           string
		matchValue     *values.Value
		cases          map[interface{}]*values.Value
		expectedResult *values.Value
	}{
		{
			name:       "match int value",
			matchValue: values.NewInt(2),
			cases: map[interface{}]*values.Value{
				1: values.NewString("one"),
				2: values.NewString("two"),
				3: values.NewString("three"),
			},
			expectedResult: values.NewString("two"),
		},
		{
			name:       "match string value",
			matchValue: values.NewString("hello"),
			cases: map[interface{}]*values.Value{
				"hello": values.NewString("greeting"),
				"world": values.NewString("planet"),
			},
			expectedResult: values.NewString("greeting"),
		},
		{
			name:       "no match found",
			matchValue: values.NewInt(99),
			cases: map[interface{}]*values.Value{
				1: values.NewString("one"),
				2: values.NewString("two"),
			},
			expectedResult: values.NewNull(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := NewVirtualMachine()
			ctx := NewExecutionContext()

			// Setup match value
			ctx.Temporaries[1] = tt.matchValue

			// Setup cases array
			casesArray := values.NewArray()
			casesData := casesArray.Data.(*values.Array)
			for key, value := range tt.cases {
				casesData.Elements[key] = value
			}
			ctx.Temporaries[2] = casesArray

			// Create MATCH instruction
			inst := &opcodes.Instruction{
				Opcode: opcodes.OP_MATCH,
				Op1:    1, // Match value
				Op2:    2, // Cases array
				Result: 3, // Result location
			}
			inst.OpType1, inst.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR)

			err := vm.executeMatch(ctx, inst)
			if err != nil {
				t.Fatalf("executeMatch failed: %v", err)
			}

			// Check result
			result := ctx.Temporaries[3]
			if result == nil {
				t.Fatal("Result should not be nil")
			}

			if !valuesEqual(result, tt.expectedResult) {
				t.Errorf("Expected result %v, got %v", tt.expectedResult, result)
			}
		})
	}
}

func TestYieldFromOpcode(t *testing.T) {
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	// Create an array to yield from
	sourceArray := values.NewArray()
	sourceData := sourceArray.Data.(*values.Array)
	sourceData.Elements[0] = values.NewString("a")
	sourceData.Elements[1] = values.NewString("b")
	sourceData.Elements[2] = values.NewString("c")

	ctx.Temporaries[1] = sourceArray

	// Setup generator context
	ctx.CurrentGenerator = &Generator{
		Function:     nil,
		Context:      ctx,
		Variables:    make(map[uint32]*values.Value),
		IP:           0,
		YieldedKey:   nil,
		YieldedValue: nil,
		IsFinished:   false,
		IsSuspended:  false,
	}

	// Create YIELD_FROM instruction
	inst := &opcodes.Instruction{
		Opcode: opcodes.OP_YIELD_FROM,
		Op1:    1, // Source array/iterator
		Op2:    0, // Unused
		Result: 2, // Result location
	}
	inst.OpType1, inst.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)

	err := vm.executeYieldFrom(ctx, inst)
	if err != nil {
		t.Fatalf("executeYieldFrom failed: %v", err)
	}

	// Check that result is set to null (default return value)
	result := ctx.Temporaries[2]
	if result == nil || !result.IsNull() {
		t.Errorf("Expected null result, got %v", result)
	}

	// Check that generator has yielded something (simplified implementation)
	if ctx.CurrentGenerator != nil {
		// In the current implementation, it yields the last value from the array
		if ctx.CurrentGenerator.YieldedValue == nil {
			t.Error("Generator should have yielded a value")
		}
	}
}

// Helper function to compare values
func valuesEqual(a, b *values.Value) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	if a.Type != b.Type {
		return false
	}

	switch a.Type {
	case values.TypeNull:
		return true
	case values.TypeBool:
		return a.ToBool() == b.ToBool()
	case values.TypeInt:
		return a.ToInt() == b.ToInt()
	case values.TypeFloat:
		return a.ToFloat() == b.ToFloat()
	case values.TypeString:
		return a.ToString() == b.ToString()
	default:
		return a == b
	}
}

func TestAdvancedOpcodeErrors(t *testing.T) {
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	// Test BIND_GLOBAL with non-string name
	ctx.Temporaries[1] = values.NewInt(123) // Invalid name

	inst := &opcodes.Instruction{
		Opcode: opcodes.OP_BIND_GLOBAL,
		Op1:    1,
		Op2:    2,
		Result: 0,
	}
	inst.OpType1, inst.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_UNUSED)

	err := vm.executeBindGlobal(ctx, inst)
	if err == nil {
		t.Error("Expected error for non-string variable name in BIND_GLOBAL")
	}

	// Test ADD_ARRAY_UNPACK with non-array target
	ctx.Temporaries[1] = values.NewArray()
	ctx.Temporaries[2] = values.NewString("not_array") // Invalid target

	inst2 := &opcodes.Instruction{
		Opcode: opcodes.OP_ADD_ARRAY_UNPACK,
		Op1:    1,
		Op2:    0,
		Result: 2,
	}
	inst2.OpType1, inst2.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)

	err = vm.executeAddArrayUnpack(ctx, inst2)
	if err == nil {
		t.Error("Expected error for non-array target in ADD_ARRAY_UNPACK")
	}
}
