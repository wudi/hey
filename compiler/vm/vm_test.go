package vm

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/wudi/hey/compiler/opcodes"
	"github.com/wudi/hey/compiler/registry"
	"github.com/wudi/hey/compiler/runtime"
	"github.com/wudi/hey/compiler/values"
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

func TestCountOpcode(t *testing.T) {
	tests := []struct {
		name     string
		input    *values.Value
		expected int64
	}{
		{"empty array count", createArray(), 0},
		{"single element array count", createArrayWithElements([]interface{}{int64(0)}, []*values.Value{values.NewString("hello")}), 1},
		{"multi element array count", createArrayWithElements([]interface{}{int64(0), int64(1), "key"}, []*values.Value{values.NewInt(1), values.NewInt(2), values.NewString("value")}), 3},
		{"string count (length)", values.NewString("hello"), 5},
		{"empty string count", values.NewString(""), 0},
		{"int count (not array/string)", values.NewInt(42), 0},
		{"null count", values.NewNull(), 0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vm := NewVirtualMachine()
			ctx := NewExecutionContext()

			ctx.Temporaries = make(map[uint32]*values.Value)
			ctx.Temporaries[0] = test.input

			// Create COUNT instruction
			op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)
			inst := opcodes.Instruction{
				Opcode:  opcodes.OP_COUNT,
				OpType1: op1Type,
				OpType2: op2Type,
				Op1:     0, // Input from temporary variable 0
				Result:  1, // Store in temporary variable 1
			}

			// Execute COUNT
			err := vm.executeCount(ctx, &inst)
			if err != nil {
				t.Fatalf("COUNT execution failed: %v", err)
			}

			// Check result
			result := ctx.Temporaries[1]
			if result == nil {
				t.Fatal("COUNT result is nil")
			}

			if result.Type != values.TypeInt {
				t.Errorf("Expected int type, got %v", result.Type)
			}

			if result.Data.(int64) != test.expected {
				t.Errorf("Expected count %v, got %v", test.expected, result.Data.(int64))
			}
		})
	}
}

func TestInArrayOpcode(t *testing.T) {
	// Create test array: [1, "hello", 3.14, true]
	testArray := createArrayWithElements(
		[]interface{}{int64(0), int64(1), int64(2), int64(3)},
		[]*values.Value{
			values.NewInt(1),
			values.NewString("hello"),
			values.NewFloat(3.14),
			values.NewBool(true),
		},
	)

	tests := []struct {
		name     string
		needle   *values.Value
		haystack *values.Value
		expected bool
	}{
		{"find int in array", values.NewInt(1), testArray, true},
		{"find string in array", values.NewString("hello"), testArray, true},
		{"find float in array", values.NewFloat(3.14), testArray, true},
		{"find bool in array", values.NewBool(true), testArray, true},
		{"find via loose comparison int", values.NewInt(99), testArray, true},            // 99 == true (loose)
		{"find via loose comparison string", values.NewString("world"), testArray, true}, // "world" == true (loose)
		{"not find int zero in array", values.NewInt(0), testArray, false},               // 0 doesn't match any element
		{"not find false in array", values.NewBool(false), testArray, false},             // false doesn't match any element
		{"not find empty string in array", values.NewString(""), testArray, false},       // "" doesn't match any element
		{"not find null in array", values.NewNull(), testArray, false},                   // null doesn't match any element
		{"search in non-array", values.NewInt(1), values.NewString("hello"), false},
		{"search in empty array", values.NewInt(1), createArray(), false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vm := NewVirtualMachine()
			ctx := NewExecutionContext()

			ctx.Temporaries = make(map[uint32]*values.Value)
			ctx.Temporaries[0] = test.needle
			ctx.Temporaries[1] = test.haystack

			// Create IN_ARRAY instruction
			op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR)
			inst := opcodes.Instruction{
				Opcode:  opcodes.OP_IN_ARRAY,
				OpType1: op1Type,
				OpType2: op2Type,
				Op1:     0, // Needle from temporary variable 0
				Op2:     1, // Haystack from temporary variable 1
				Result:  2, // Store in temporary variable 2
			}

			// Execute IN_ARRAY
			err := vm.executeInArray(ctx, &inst)
			if err != nil {
				t.Fatalf("IN_ARRAY execution failed: %v", err)
			}

			// Check result
			result := ctx.Temporaries[2]
			if result == nil {
				t.Fatal("IN_ARRAY result is nil")
			}

			if result.Type != values.TypeBool {
				t.Errorf("Expected bool type, got %v", result.Type)
			}

			if result.Data.(bool) != test.expected {
				t.Errorf("Expected %v, got %v", test.expected, result.Data.(bool))
			}
		})
	}
}

func TestArrayKeyExistsOpcode(t *testing.T) {
	// Create test array: [0 => "zero", "key1" => "value1", 5 => "five"]
	testArray := createArrayWithElements(
		[]interface{}{int64(0), "key1", int64(5)},
		[]*values.Value{
			values.NewString("zero"),
			values.NewString("value1"),
			values.NewString("five"),
		},
	)

	tests := []struct {
		name     string
		key      *values.Value
		array    *values.Value
		expected bool
	}{
		{"key exists - int key", values.NewInt(0), testArray, true},
		{"key exists - string key", values.NewString("key1"), testArray, true},
		{"key exists - large int key", values.NewInt(5), testArray, true},
		{"key doesn't exist - int", values.NewInt(1), testArray, false},
		{"key doesn't exist - string", values.NewString("key2"), testArray, false},
		{"search in non-array", values.NewInt(0), values.NewString("hello"), false},
		{"search in empty array", values.NewInt(0), createArray(), false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vm := NewVirtualMachine()
			ctx := NewExecutionContext()

			ctx.Temporaries = make(map[uint32]*values.Value)
			ctx.Temporaries[0] = test.key
			ctx.Temporaries[1] = test.array

			// Create ARRAY_KEY_EXISTS instruction
			op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR)
			inst := opcodes.Instruction{
				Opcode:  opcodes.OP_ARRAY_KEY_EXISTS,
				OpType1: op1Type,
				OpType2: op2Type,
				Op1:     0, // Key from temporary variable 0
				Op2:     1, // Array from temporary variable 1
				Result:  2, // Store in temporary variable 2
			}

			// Execute ARRAY_KEY_EXISTS
			err := vm.executeArrayKeyExists(ctx, &inst)
			if err != nil {
				t.Fatalf("ARRAY_KEY_EXISTS execution failed: %v", err)
			}

			// Check result
			result := ctx.Temporaries[2]
			if result == nil {
				t.Fatal("ARRAY_KEY_EXISTS result is nil")
			}

			if result.Type != values.TypeBool {
				t.Errorf("Expected bool type, got %v", result.Type)
			}

			if result.Data.(bool) != test.expected {
				t.Errorf("Expected %v, got %v", test.expected, result.Data.(bool))
			}
		})
	}
}

func TestArrayValuesOpcode(t *testing.T) {
	// Create test array: [0 => "zero", "key1" => "value1", 5 => "five"]
	testArray := createArrayWithElements(
		[]interface{}{int64(0), "key1", int64(5)},
		[]*values.Value{
			values.NewString("zero"),
			values.NewString("value1"),
			values.NewString("five"),
		},
	)

	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	ctx.Temporaries = make(map[uint32]*values.Value)
	ctx.Temporaries[0] = testArray

	// Create ARRAY_VALUES instruction
	op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)
	inst := opcodes.Instruction{
		Opcode:  opcodes.OP_ARRAY_VALUES,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     0, // Array from temporary variable 0
		Result:  1, // Store in temporary variable 1
	}

	// Execute ARRAY_VALUES
	err := vm.executeArrayValues(ctx, &inst)
	if err != nil {
		t.Fatalf("ARRAY_VALUES execution failed: %v", err)
	}

	// Check result
	result := ctx.Temporaries[1]
	if result == nil {
		t.Fatal("ARRAY_VALUES result is nil")
	}

	if result.Type != values.TypeArray {
		t.Errorf("Expected array type, got %v", result.Type)
	}

	// Should have 3 elements with sequential numeric keys
	if result.ArrayCount() != 3 {
		t.Errorf("Expected array with 3 elements, got %d", result.ArrayCount())
	}

	// Check that values are present (order may vary due to map iteration)
	elem0 := result.ArrayGet(values.NewInt(0))
	elem1 := result.ArrayGet(values.NewInt(1))
	elem2 := result.ArrayGet(values.NewInt(2))

	if elem0 == nil || elem1 == nil || elem2 == nil {
		t.Error("Expected all elements to be present")
	}
}

func TestArrayKeysOpcode(t *testing.T) {
	// Create test array: [0 => "zero", "key1" => "value1", 5 => "five"]
	testArray := createArrayWithElements(
		[]interface{}{int64(0), "key1", int64(5)},
		[]*values.Value{
			values.NewString("zero"),
			values.NewString("value1"),
			values.NewString("five"),
		},
	)

	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	ctx.Temporaries = make(map[uint32]*values.Value)
	ctx.Temporaries[0] = testArray

	// Create ARRAY_KEYS instruction
	op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)
	inst := opcodes.Instruction{
		Opcode:  opcodes.OP_ARRAY_KEYS,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     0, // Array from temporary variable 0
		Result:  1, // Store in temporary variable 1
	}

	// Execute ARRAY_KEYS
	err := vm.executeArrayKeys(ctx, &inst)
	if err != nil {
		t.Fatalf("ARRAY_KEYS execution failed: %v", err)
	}

	// Check result
	result := ctx.Temporaries[1]
	if result == nil {
		t.Fatal("ARRAY_KEYS result is nil")
	}

	if result.Type != values.TypeArray {
		t.Errorf("Expected array type, got %v", result.Type)
	}

	// Should have 3 elements with sequential numeric keys
	if result.ArrayCount() != 3 {
		t.Errorf("Expected array with 3 elements, got %d", result.ArrayCount())
	}

	// Check that keys are present (order may vary due to map iteration)
	elem0 := result.ArrayGet(values.NewInt(0))
	elem1 := result.ArrayGet(values.NewInt(1))
	elem2 := result.ArrayGet(values.NewInt(2))

	if elem0 == nil || elem1 == nil || elem2 == nil {
		t.Error("Expected all key elements to be present")
	}
}

func TestArrayMergeOpcode(t *testing.T) {
	// Create first array: [0 => "a", 1 => "b"]
	array1 := createArrayWithElements(
		[]interface{}{int64(0), int64(1)},
		[]*values.Value{values.NewString("a"), values.NewString("b")},
	)

	// Create second array: [2 => "c", "key" => "d"]
	array2 := createArrayWithElements(
		[]interface{}{int64(2), "key"},
		[]*values.Value{values.NewString("c"), values.NewString("d")},
	)

	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	ctx.Temporaries = make(map[uint32]*values.Value)
	ctx.Temporaries[0] = array1
	ctx.Temporaries[1] = array2

	// Create ARRAY_MERGE instruction
	op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR)
	inst := opcodes.Instruction{
		Opcode:  opcodes.OP_ARRAY_MERGE,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     0, // First array from temporary variable 0
		Op2:     1, // Second array from temporary variable 1
		Result:  2, // Store in temporary variable 2
	}

	// Execute ARRAY_MERGE
	err := vm.executeArrayMerge(ctx, &inst)
	if err != nil {
		t.Fatalf("ARRAY_MERGE execution failed: %v", err)
	}

	// Check result
	result := ctx.Temporaries[2]
	if result == nil {
		t.Fatal("ARRAY_MERGE result is nil")
	}

	if result.Type != values.TypeArray {
		t.Errorf("Expected array type, got %v", result.Type)
	}

	// Should have 4 elements total
	if result.ArrayCount() != 4 {
		t.Errorf("Expected merged array with 4 elements, got %d", result.ArrayCount())
	}

	// Check that merged elements are accessible
	valA := result.ArrayGet(values.NewInt(0))
	valB := result.ArrayGet(values.NewInt(1))
	valC := result.ArrayGet(values.NewInt(2))
	valD := result.ArrayGet(values.NewString("key"))

	if valA == nil || valA.Data.(string) != "a" {
		t.Error("Expected merged array to contain 'a' at index 0")
	}
	if valB == nil || valB.Data.(string) != "b" {
		t.Error("Expected merged array to contain 'b' at index 1")
	}
	if valC == nil || valC.Data.(string) != "c" {
		t.Error("Expected merged array to contain 'c' at index 2")
	}
	if valD == nil || valD.Data.(string) != "d" {
		t.Error("Expected merged array to contain 'd' at key 'key'")
	}
}

// Test comprehensive array operations simulation
func TestArrayOperationsSimulation(t *testing.T) {
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	// Simulate: count(array_merge($arr1, $arr2)) where $arr1 has 2 elements, $arr2 has 1 element
	// Expected result: 3

	// Create arrays
	arr1 := createArrayWithElements(
		[]interface{}{int64(0), int64(1)},
		[]*values.Value{values.NewString("first"), values.NewString("second")},
	)
	arr2 := createArrayWithElements(
		[]interface{}{int64(0)},
		[]*values.Value{values.NewString("third")},
	)

	ctx.Temporaries = make(map[uint32]*values.Value)
	ctx.Temporaries[0] = arr1
	ctx.Temporaries[1] = arr2

	op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR)
	op1TypeSingle, op2TypeSingle := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)

	// 1. Merge arrays: array_merge($arr1, $arr2)
	mergeInst := opcodes.Instruction{
		Opcode:  opcodes.OP_ARRAY_MERGE,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     0, // arr1
		Op2:     1, // arr2
		Result:  2, // Store merged result in temp var 2
	}

	err := vm.executeArrayMerge(ctx, &mergeInst)
	if err != nil {
		t.Fatalf("ARRAY_MERGE execution failed: %v", err)
	}

	mergedArray := ctx.Temporaries[2]
	if mergedArray.ArrayCount() != 3 {
		t.Errorf("Expected merged array with 3 elements, got %d", mergedArray.ArrayCount())
	}

	// 2. Count merged array: count(merged_array)
	countInst := opcodes.Instruction{
		Opcode:  opcodes.OP_COUNT,
		OpType1: op1TypeSingle,
		OpType2: op2TypeSingle,
		Op1:     2, // Merged array from step 1
		Result:  3, // Store count in temp var 3
	}

	err = vm.executeCount(ctx, &countInst)
	if err != nil {
		t.Fatalf("COUNT execution failed: %v", err)
	}

	count := ctx.Temporaries[3]
	if !count.IsInt() || count.Data.(int64) != 3 {
		t.Errorf("Expected count 3, got %v", count.Data)
	}

	// 3. Test in_array: in_array("second", merged_array)
	needle := values.NewString("second")
	ctx.Temporaries[4] = needle

	inArrayInst := opcodes.Instruction{
		Opcode:  opcodes.OP_IN_ARRAY,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     4, // "second" from temp var 4
		Op2:     2, // merged array from step 1
		Result:  5, // Store result in temp var 5
	}

	err = vm.executeInArray(ctx, &inArrayInst)
	if err != nil {
		t.Fatalf("IN_ARRAY execution failed: %v", err)
	}

	inArrayResult := ctx.Temporaries[5]
	if !inArrayResult.IsBool() || !inArrayResult.Data.(bool) {
		t.Errorf("Expected in_array to return true, got %v", inArrayResult.Data)
	}

	// 4. Test array_key_exists: array_key_exists(1, merged_array)
	key := values.NewInt(1)
	ctx.Temporaries[6] = key

	keyExistsInst := opcodes.Instruction{
		Opcode:  opcodes.OP_ARRAY_KEY_EXISTS,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     6, // key 1 from temp var 6
		Op2:     2, // merged array from step 1
		Result:  7, // Store result in temp var 7
	}

	err = vm.executeArrayKeyExists(ctx, &keyExistsInst)
	if err != nil {
		t.Fatalf("ARRAY_KEY_EXISTS execution failed: %v", err)
	}

	keyExistsResult := ctx.Temporaries[7]
	if !keyExistsResult.IsBool() || !keyExistsResult.Data.(bool) {
		t.Errorf("Expected array_key_exists to return true, got %v", keyExistsResult.Data)
	}
}

// Helper function to create an empty array
func createArray() *values.Value {
	return values.NewArray()
}

// Helper function to create an array with elements
func createArrayWithElements(keys []interface{}, vals []*values.Value) *values.Value {
	array := values.NewArray()
	for i, key := range keys {
		if i < len(vals) {
			var keyValue *values.Value
			switch k := key.(type) {
			case int64:
				keyValue = values.NewInt(k)
			case string:
				keyValue = values.NewString(k)
			default:
				keyValue = values.NewInt(0)
			}
			array.ArraySet(keyValue, vals[i])
		}
	}
	return array
}

func TestAssignOpOpcode(t *testing.T) {
	tests := []struct {
		name     string
		initial  *values.Value
		value    *values.Value
		opType   byte
		expected interface{}
		expType  values.ValueType
	}{
		{"assign add (+=)", values.NewInt(10), values.NewInt(5), ZEND_ADD, int64(15), values.TypeInt},
		{"assign sub (-=)", values.NewInt(20), values.NewInt(3), ZEND_SUB, int64(17), values.TypeInt},
		{"assign mul (*=)", values.NewInt(4), values.NewInt(3), ZEND_MUL, int64(12), values.TypeInt},
		{"assign div (/=)", values.NewInt(15), values.NewInt(3), ZEND_DIV, int64(5), values.TypeInt}, // PHP returns int when division result is whole number
		{"assign mod (%=)", values.NewInt(17), values.NewInt(5), ZEND_MOD, int64(2), values.TypeInt},
		{"assign pow (**=)", values.NewInt(2), values.NewInt(3), ZEND_POW, int64(8), values.TypeInt},
		{"assign concat (.=)", values.NewString("Hello"), values.NewString(" World"), ZEND_CONCAT, "Hello World", values.TypeString},
		{"assign bw_or (|=)", values.NewInt(12), values.NewInt(3), ZEND_BW_OR, int64(15), values.TypeInt},    // 1100 | 0011 = 1111
		{"assign bw_and (&=)", values.NewInt(15), values.NewInt(7), ZEND_BW_AND, int64(7), values.TypeInt},   // 1111 & 0111 = 0111
		{"assign bw_xor (^=)", values.NewInt(12), values.NewInt(10), ZEND_BW_XOR, int64(6), values.TypeInt},  // 1100 ^ 1010 = 0110
		{"assign shift left (<<=)", values.NewInt(5), values.NewInt(2), ZEND_SL, int64(20), values.TypeInt},  // 5 << 2 = 20
		{"assign shift right (>>=)", values.NewInt(20), values.NewInt(2), ZEND_SR, int64(5), values.TypeInt}, // 20 >> 2 = 5
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vm := NewVirtualMachine()
			ctx := NewExecutionContext()

			ctx.Temporaries = make(map[uint32]*values.Value)
			ctx.Temporaries[0] = test.initial
			ctx.Temporaries[1] = test.value

			// Create ASSIGN_OP instruction
			op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR)
			inst := opcodes.Instruction{
				Opcode:   opcodes.OP_ASSIGN_OP,
				OpType1:  op1Type,
				OpType2:  op2Type,
				Reserved: test.opType, // Store operation type
				Op1:      0,           // Initial value from temporary variable 0
				Op2:      1,           // Value to operate with from temporary variable 1
				Result:   2,           // Store in temporary variable 2
			}

			// Execute ASSIGN_OP
			err := vm.executeAssignOp(ctx, &inst)
			if err != nil {
				t.Fatalf("ASSIGN_OP execution failed: %v", err)
			}

			// Check result
			result := ctx.Temporaries[2]
			if result == nil {
				t.Fatal("ASSIGN_OP result is nil")
			}

			if result.Type != test.expType {
				t.Errorf("Expected type %v, got %v", test.expType, result.Type)
			}

			switch test.expType {
			case values.TypeInt:
				if result.Data.(int64) != test.expected.(int64) {
					t.Errorf("Expected %v, got %v", test.expected, result.Data.(int64))
				}
			case values.TypeFloat:
				if result.Data.(float64) != test.expected.(float64) {
					t.Errorf("Expected %v, got %v", test.expected, result.Data.(float64))
				}
			case values.TypeString:
				if result.Data.(string) != test.expected.(string) {
					t.Errorf("Expected '%v', got '%v'", test.expected, result.Data.(string))
				}
			}
		})
	}
}

func TestAssignOpWithPhpSemantics(t *testing.T) {
	// Test PHP-like behavior with different types
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	// Test string + number (PHP concatenates strings, adds numbers)
	ctx.Temporaries = make(map[uint32]*values.Value)
	ctx.Temporaries[0] = values.NewString("10")
	ctx.Temporaries[1] = values.NewString("5")

	op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR)
	inst := opcodes.Instruction{
		Opcode:   opcodes.OP_ASSIGN_OP,
		OpType1:  op1Type,
		OpType2:  op2Type,
		Reserved: ZEND_ADD, // += operation
		Op1:      0,
		Op2:      1,
		Result:   2,
	}

	err := vm.executeAssignOp(ctx, &inst)
	if err != nil {
		t.Fatalf("ASSIGN_OP execution failed: %v", err)
	}

	result := ctx.Temporaries[2]
	// PHP would convert string numbers to actual numbers for arithmetic
	if result.Type != values.TypeInt && result.Type != values.TypeFloat {
		t.Errorf("Expected numeric result for string number addition, got %v", result.Type)
	}
}

func TestAssignOpUnknownOperation(t *testing.T) {
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	ctx.Temporaries = make(map[uint32]*values.Value)
	ctx.Temporaries[0] = values.NewInt(10)
	ctx.Temporaries[1] = values.NewInt(5)

	op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR)
	inst := opcodes.Instruction{
		Opcode:   opcodes.OP_ASSIGN_OP,
		OpType1:  op1Type,
		OpType2:  op2Type,
		Reserved: 99, // Invalid operation type
		Op1:      0,
		Op2:      1,
		Result:   2,
	}

	// Should return error for unknown operation
	err := vm.executeAssignOp(ctx, &inst)
	if err == nil {
		t.Fatal("Expected error for unknown operation type, but got none")
	}
}

func TestAssignDimOpcode(t *testing.T) {
	tests := []struct {
		name     string
		array    *values.Value
		key      *values.Value
		value    *values.Value
		expected string
	}{
		{
			"assign to string key",
			values.NewArray(),
			values.NewString("key1"),
			values.NewString("value1"),
			"value1",
		},
		{
			"assign to int key",
			values.NewArray(),
			values.NewInt(5),
			values.NewString("value2"),
			"value2",
		},
		{
			"assign to new array (auto-creation)",
			values.NewNull(), // Non-array, should auto-convert
			values.NewInt(0),
			values.NewString("auto"),
			"auto",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vm := NewVirtualMachine()
			ctx := NewExecutionContext()

			ctx.Temporaries = make(map[uint32]*values.Value)
			ctx.Temporaries[0] = test.array // Array variable
			ctx.Temporaries[1] = test.key   // Key
			ctx.Temporaries[2] = test.value // Value to assign

			// Create ASSIGN_DIM instruction
			op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR)
			inst := opcodes.Instruction{
				Opcode:   opcodes.OP_ASSIGN_DIM,
				OpType1:  op1Type,
				OpType2:  op2Type,
				Reserved: 2, // Value is in temporary variable 2
				Op1:      0, // Array from temporary variable 0
				Op2:      1, // Key from temporary variable 1
				Result:   3, // Store result in temporary variable 3
			}

			// Execute ASSIGN_DIM
			err := vm.executeAssignDim(ctx, &inst)
			if err != nil {
				t.Fatalf("ASSIGN_DIM execution failed: %v", err)
			}

			// Check that array was updated (or created)
			resultArray := ctx.Temporaries[0]
			if !resultArray.IsArray() {
				t.Fatal("Expected array after ASSIGN_DIM")
			}

			// Check that the value was assigned correctly
			retrievedValue := resultArray.ArrayGet(test.key)
			if retrievedValue == nil {
				t.Fatal("Could not retrieve assigned value from array")
			}

			if retrievedValue.ToString() != test.expected {
				t.Errorf("Expected '%v', got '%v'", test.expected, retrievedValue.ToString())
			}

			// Check that result value is returned
			result := ctx.Temporaries[3]
			if result == nil {
				t.Fatal("ASSIGN_DIM result is nil")
			}

			if result.ToString() != test.expected {
				t.Errorf("Expected result '%v', got '%v'", test.expected, result.ToString())
			}
		})
	}
}

func TestAssignObjOpcode(t *testing.T) {
	tests := []struct {
		name     string
		object   *values.Value
		property string
		value    *values.Value
		expected string
	}{
		{
			"assign to object property",
			values.NewObject("TestClass"),
			"prop1",
			values.NewString("value1"),
			"value1",
		},
		{
			"assign to different property",
			values.NewObject("TestClass"),
			"prop2",
			values.NewInt(42),
			"42",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vm := NewVirtualMachine()
			ctx := NewExecutionContext()

			ctx.Temporaries = make(map[uint32]*values.Value)
			ctx.Temporaries[0] = test.object                     // Object variable
			ctx.Temporaries[1] = values.NewString(test.property) // Property name
			ctx.Temporaries[2] = test.value                      // Value to assign

			// Create ASSIGN_OBJ instruction
			op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR)
			inst := opcodes.Instruction{
				Opcode:   opcodes.OP_ASSIGN_OBJ,
				OpType1:  op1Type,
				OpType2:  op2Type,
				Reserved: 2, // Value is in temporary variable 2
				Op1:      0, // Object from temporary variable 0
				Op2:      1, // Property name from temporary variable 1
				Result:   3, // Store result in temporary variable 3
			}

			// Execute ASSIGN_OBJ
			err := vm.executeAssignObj(ctx, &inst)
			if err != nil {
				t.Fatalf("ASSIGN_OBJ execution failed: %v", err)
			}

			// Check that object property was set
			resultObject := ctx.Temporaries[0]
			if !resultObject.IsObject() {
				t.Fatal("Expected object after ASSIGN_OBJ")
			}

			// Check that the property was assigned correctly
			retrievedValue := resultObject.ObjectGet(test.property)
			if retrievedValue == nil {
				t.Fatal("Could not retrieve assigned property from object")
			}

			if retrievedValue.ToString() != test.expected {
				t.Errorf("Expected '%v', got '%v'", test.expected, retrievedValue.ToString())
			}

			// Check that result value is returned
			result := ctx.Temporaries[3]
			if result == nil {
				t.Fatal("ASSIGN_OBJ result is nil")
			}

			if result.ToString() != test.expected {
				t.Errorf("Expected result '%v', got '%v'", test.expected, result.ToString())
			}
		})
	}
}

func TestAssignObjNonObject(t *testing.T) {
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	ctx.Temporaries = make(map[uint32]*values.Value)
	ctx.Temporaries[0] = values.NewString("not an object") // Not an object
	ctx.Temporaries[1] = values.NewString("prop")          // Property name
	ctx.Temporaries[2] = values.NewString("value")         // Value to assign

	op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR)
	inst := opcodes.Instruction{
		Opcode:   opcodes.OP_ASSIGN_OBJ,
		OpType1:  op1Type,
		OpType2:  op2Type,
		Reserved: 2,
		Op1:      0,
		Op2:      1,
		Result:   3,
	}

	// Should return error for non-object
	err := vm.executeAssignObj(ctx, &inst)
	if err == nil {
		t.Fatal("Expected error when assigning to non-object, but got none")
	}
}

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

func TestCastToInt(t *testing.T) {
	tests := []struct {
		name     string
		input    *values.Value
		expected int64
	}{
		{"string number", values.NewString("42"), 42},
		{"string float", values.NewString("3.14"), 3},
		{"string invalid", values.NewString("hello"), 0},
		{"float", values.NewFloat(3.14), 3},
		{"bool true", values.NewBool(true), 1},
		{"bool false", values.NewBool(false), 0},
		{"null", values.NewNull(), 0},
		{"array empty", values.NewArray(), 0},
		{"array with items", func() *values.Value {
			arr := values.NewArray()
			arr.ArraySet(values.NewInt(0), values.NewString("test"))
			return arr
		}(), 1},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vm := NewVirtualMachine()
			ctx := NewExecutionContext()

			ctx.Temporaries = make(map[uint32]*values.Value)
			ctx.Temporaries[0] = test.input

			op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)
			inst := opcodes.Instruction{
				Opcode:   opcodes.OP_CAST,
				OpType1:  op1Type,
				OpType2:  op2Type,
				Reserved: opcodes.CAST_IS_LONG,
				Op1:      0,
				Result:   1,
			}

			err := vm.executeCast(ctx, &inst)
			if err != nil {
				t.Fatalf("CAST to int failed: %v", err)
			}

			result := ctx.Temporaries[1]
			if result == nil {
				t.Fatal("CAST result is nil")
			}

			if !result.IsInt() {
				t.Errorf("Expected int result, got %v", result.Type)
			}

			if result.Data.(int64) != test.expected {
				t.Errorf("Expected %d, got %d", test.expected, result.Data.(int64))
			}
		})
	}
}

func TestCastToFloat(t *testing.T) {
	tests := []struct {
		name     string
		input    *values.Value
		expected float64
	}{
		{"string number", values.NewString("3.14"), 3.14},
		{"string int", values.NewString("42"), 42.0},
		{"int", values.NewInt(42), 42.0},
		{"bool true", values.NewBool(true), 1.0},
		{"bool false", values.NewBool(false), 0.0},
		{"null", values.NewNull(), 0.0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vm := NewVirtualMachine()
			ctx := NewExecutionContext()

			ctx.Temporaries = make(map[uint32]*values.Value)
			ctx.Temporaries[0] = test.input

			op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)
			inst := opcodes.Instruction{
				Opcode:   opcodes.OP_CAST,
				OpType1:  op1Type,
				OpType2:  op2Type,
				Reserved: opcodes.CAST_IS_DOUBLE,
				Op1:      0,
				Result:   1,
			}

			err := vm.executeCast(ctx, &inst)
			if err != nil {
				t.Fatalf("CAST to float failed: %v", err)
			}

			result := ctx.Temporaries[1]
			if result == nil {
				t.Fatal("CAST result is nil")
			}

			if !result.IsFloat() {
				t.Errorf("Expected float result, got %v", result.Type)
			}

			if result.Data.(float64) != test.expected {
				t.Errorf("Expected %f, got %f", test.expected, result.Data.(float64))
			}
		})
	}
}

func TestCastToString(t *testing.T) {
	tests := []struct {
		name     string
		input    *values.Value
		expected string
	}{
		{"int", values.NewInt(42), "42"},
		{"float", values.NewFloat(3.14), "3.14"},
		{"bool true", values.NewBool(true), "1"},
		{"bool false", values.NewBool(false), ""},
		{"null", values.NewNull(), ""},
		{"string", values.NewString("hello"), "hello"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vm := NewVirtualMachine()
			ctx := NewExecutionContext()

			ctx.Temporaries = make(map[uint32]*values.Value)
			ctx.Temporaries[0] = test.input

			op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)
			inst := opcodes.Instruction{
				Opcode:   opcodes.OP_CAST,
				OpType1:  op1Type,
				OpType2:  op2Type,
				Reserved: opcodes.CAST_IS_STRING,
				Op1:      0,
				Result:   1,
			}

			err := vm.executeCast(ctx, &inst)
			if err != nil {
				t.Fatalf("CAST to string failed: %v", err)
			}

			result := ctx.Temporaries[1]
			if result == nil {
				t.Fatal("CAST result is nil")
			}

			if !result.IsString() {
				t.Errorf("Expected string result, got %v", result.Type)
			}

			if result.Data.(string) != test.expected {
				t.Errorf("Expected '%s', got '%s'", test.expected, result.Data.(string))
			}
		})
	}
}

func TestCastToArray(t *testing.T) {
	tests := []struct {
		name     string
		input    *values.Value
		expected func(*values.Value) bool
	}{
		{
			"null to array",
			values.NewNull(),
			func(result *values.Value) bool {
				return result.IsArray() && result.ArrayCount() == 0
			},
		},
		{
			"int to array",
			values.NewInt(42),
			func(result *values.Value) bool {
				if !result.IsArray() || result.ArrayCount() != 1 {
					return false
				}
				elem := result.ArrayGet(values.NewInt(0))
				return elem != nil && elem.IsInt() && elem.Data.(int64) == 42
			},
		},
		{
			"string to array",
			values.NewString("hello"),
			func(result *values.Value) bool {
				if !result.IsArray() || result.ArrayCount() != 1 {
					return false
				}
				elem := result.ArrayGet(values.NewInt(0))
				return elem != nil && elem.IsString() && elem.Data.(string) == "hello"
			},
		},
		{
			"array to array",
			func() *values.Value {
				arr := values.NewArray()
				arr.ArraySet(values.NewString("key"), values.NewString("value"))
				return arr
			}(),
			func(result *values.Value) bool {
				return result.IsArray() && result.ArrayCount() == 1
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vm := NewVirtualMachine()
			ctx := NewExecutionContext()

			ctx.Temporaries = make(map[uint32]*values.Value)
			ctx.Temporaries[0] = test.input

			op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)
			inst := opcodes.Instruction{
				Opcode:   opcodes.OP_CAST,
				OpType1:  op1Type,
				OpType2:  op2Type,
				Reserved: opcodes.CAST_IS_ARRAY,
				Op1:      0,
				Result:   1,
			}

			err := vm.executeCast(ctx, &inst)
			if err != nil {
				t.Fatalf("CAST to array failed: %v", err)
			}

			result := ctx.Temporaries[1]
			if result == nil {
				t.Fatal("CAST result is nil")
			}

			if !test.expected(result) {
				t.Errorf("Array cast result validation failed for %s", test.name)
			}
		})
	}
}

func TestCastToObject(t *testing.T) {
	tests := []struct {
		name  string
		input *values.Value
	}{
		{"null to object", values.NewNull()},
		{"int to object", values.NewInt(42)},
		{"string to object", values.NewString("hello")},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vm := NewVirtualMachine()
			ctx := NewExecutionContext()

			ctx.Temporaries = make(map[uint32]*values.Value)
			ctx.Temporaries[0] = test.input

			op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)
			inst := opcodes.Instruction{
				Opcode:   opcodes.OP_CAST,
				OpType1:  op1Type,
				OpType2:  op2Type,
				Reserved: opcodes.CAST_IS_OBJECT,
				Op1:      0,
				Result:   1,
			}

			err := vm.executeCast(ctx, &inst)
			if err != nil {
				t.Fatalf("CAST to object failed: %v", err)
			}

			result := ctx.Temporaries[1]
			if result == nil {
				t.Fatal("CAST result is nil")
			}

			if !result.IsObject() {
				t.Errorf("Expected object result, got %v", result.Type)
			}
		})
	}
}

func TestCastToNull(t *testing.T) {
	tests := []struct {
		name  string
		input *values.Value
	}{
		{"int to null", values.NewInt(42)},
		{"string to null", values.NewString("hello")},
		{"bool to null", values.NewBool(true)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vm := NewVirtualMachine()
			ctx := NewExecutionContext()

			ctx.Temporaries = make(map[uint32]*values.Value)
			ctx.Temporaries[0] = test.input

			op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)
			inst := opcodes.Instruction{
				Opcode:   opcodes.OP_CAST,
				OpType1:  op1Type,
				OpType2:  op2Type,
				Reserved: opcodes.CAST_IS_NULL,
				Op1:      0,
				Result:   1,
			}

			err := vm.executeCast(ctx, &inst)
			if err != nil {
				t.Fatalf("CAST to null failed: %v", err)
			}

			result := ctx.Temporaries[1]
			if result == nil {
				t.Fatal("CAST result is nil")
			}

			if !result.IsNull() {
				t.Errorf("Expected null result, got %v", result.Type)
			}
		})
	}
}

func TestCastUnknownType(t *testing.T) {
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	ctx.Temporaries = make(map[uint32]*values.Value)
	ctx.Temporaries[0] = values.NewInt(42)

	op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)
	inst := opcodes.Instruction{
		Opcode:   opcodes.OP_CAST,
		OpType1:  op1Type,
		OpType2:  op2Type,
		Reserved: 99, // Invalid cast type
		Op1:      0,
		Result:   1,
	}

	err := vm.executeCast(ctx, &inst)
	if err == nil {
		t.Fatal("Expected error for unknown cast type, but got none")
	}

	expectedError := "unknown cast type: 99"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestBoolConversion(t *testing.T) {
	tests := []struct {
		name     string
		input    *values.Value
		expected bool
	}{
		{"int zero", values.NewInt(0), false},
		{"int positive", values.NewInt(42), true},
		{"int negative", values.NewInt(-5), true},
		{"float zero", values.NewFloat(0.0), false},
		{"float positive", values.NewFloat(3.14), true},
		{"float negative", values.NewFloat(-2.5), true},
		{"string empty", values.NewString(""), false},
		{"string zero", values.NewString("0"), false},
		{"string non-empty", values.NewString("hello"), true},
		{"bool true", values.NewBool(true), true},
		{"bool false", values.NewBool(false), false},
		{"null", values.NewNull(), false},
		{"empty array", values.NewArray(), false},
		{"non-empty array", func() *values.Value {
			arr := values.NewArray()
			arr.ArraySet(values.NewInt(0), values.NewString("test"))
			return arr
		}(), true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vm := NewVirtualMachine()
			ctx := NewExecutionContext()

			ctx.Temporaries = make(map[uint32]*values.Value)
			ctx.Temporaries[0] = test.input

			op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)
			inst := opcodes.Instruction{
				Opcode:  opcodes.OP_BOOL,
				OpType1: op1Type,
				OpType2: op2Type,
				Op1:     0,
				Result:  1,
			}

			err := vm.executeBool(ctx, &inst)
			if err != nil {
				t.Fatalf("BOOL conversion failed: %v", err)
			}

			result := ctx.Temporaries[1]
			if result == nil {
				t.Fatal("BOOL result is nil")
			}

			if !result.IsBool() {
				t.Errorf("Expected bool result, got %v", result.Type)
			}

			if result.Data.(bool) != test.expected {
				t.Errorf("Expected %v, got %v", test.expected, result.Data.(bool))
			}
		})
	}
}

func TestCastOperationsIntegration(t *testing.T) {
	// Test multiple cast operations in sequence
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	ctx.Temporaries = make(map[uint32]*values.Value)
	ctx.Temporaries[0] = values.NewString("42")

	// First cast string to int
	op1Type1, op2Type1 := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)
	inst1 := opcodes.Instruction{
		Opcode:   opcodes.OP_CAST,
		OpType1:  op1Type1,
		OpType2:  op2Type1,
		Reserved: opcodes.CAST_IS_LONG,
		Op1:      0,
		Result:   1,
	}

	err := vm.executeCast(ctx, &inst1)
	if err != nil {
		t.Fatalf("First cast failed: %v", err)
	}

	// Then cast int to float
	op1Type2, op2Type2 := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)
	inst2 := opcodes.Instruction{
		Opcode:   opcodes.OP_CAST,
		OpType1:  op1Type2,
		OpType2:  op2Type2,
		Reserved: opcodes.CAST_IS_DOUBLE,
		Op1:      1,
		Result:   2,
	}

	err = vm.executeCast(ctx, &inst2)
	if err != nil {
		t.Fatalf("Second cast failed: %v", err)
	}

	// Finally convert to bool
	op1Type3, op2Type3 := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)
	inst3 := opcodes.Instruction{
		Opcode:  opcodes.OP_BOOL,
		OpType1: op1Type3,
		OpType2: op2Type3,
		Op1:     2,
		Result:  3,
	}

	err = vm.executeBool(ctx, &inst3)
	if err != nil {
		t.Fatalf("Bool conversion failed: %v", err)
	}

	// Verify all results
	intResult := ctx.Temporaries[1]
	if !intResult.IsInt() || intResult.Data.(int64) != 42 {
		t.Errorf("Int cast failed: %v", intResult)
	}

	floatResult := ctx.Temporaries[2]
	if !floatResult.IsFloat() || floatResult.Data.(float64) != 42.0 {
		t.Errorf("Float cast failed: %v", floatResult)
	}

	boolResult := ctx.Temporaries[3]
	if !boolResult.IsBool() || boolResult.Data.(bool) != true {
		t.Errorf("Bool conversion failed: %v", boolResult)
	}
}

func TestCloneOpcode(t *testing.T) {
	tests := []struct {
		name    string
		object  *values.Value
		wantErr bool
		errMsg  string
	}{
		{
			name: "clone simple object",
			object: &values.Value{
				Type: values.TypeObject,
				Data: values.Object{
					ClassName: "TestClass",
					Properties: map[string]*values.Value{
						"name": values.NewString("original"),
						"age":  values.NewInt(25),
					},
				},
			},
			wantErr: false,
		},
		{
			name:    "clone non-object (string)",
			object:  values.NewString("not an object"),
			wantErr: true,
			errMsg:  "__clone method called on non-object",
		},
		{
			name:    "clone non-object (int)",
			object:  values.NewInt(42),
			wantErr: true,
			errMsg:  "__clone method called on non-object",
		},
		{
			name:    "clone non-object (array)",
			object:  values.NewArray(),
			wantErr: true,
			errMsg:  "__clone method called on non-object",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vm := NewVirtualMachine()
			ctx := NewExecutionContext()

			ctx.Temporaries = make(map[uint32]*values.Value)
			ctx.Temporaries[0] = test.object

			op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)
			inst := opcodes.Instruction{
				Opcode:  opcodes.OP_CLONE,
				OpType1: op1Type,
				OpType2: op2Type,
				Op1:     0,
				Result:  1,
			}

			err := vm.executeClone(ctx, &inst)

			if test.wantErr {
				if err == nil {
					t.Fatal("Expected error, but got none")
				}
				if err.Error() != test.errMsg {
					t.Errorf("Expected error message '%s', got '%s'", test.errMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Check result
			result := ctx.Temporaries[1]
			if result == nil {
				t.Fatal("Clone result is nil")
			}

			if !result.IsObject() {
				t.Fatalf("Expected cloned result to be object, got %v", result.Type)
			}

			// Verify it's a different object instance
			if result == test.object {
				t.Error("Cloned object is the same instance as original (should be different)")
			}

			// Verify deep copy semantics
			originalObj := test.object.Data.(values.Object)
			clonedObj := result.Data.(values.Object)

			if clonedObj.ClassName != originalObj.ClassName {
				t.Errorf("Expected cloned class name '%s', got '%s'", originalObj.ClassName, clonedObj.ClassName)
			}

			// Check properties are copied
			if len(clonedObj.Properties) != len(originalObj.Properties) {
				t.Errorf("Expected %d properties in clone, got %d", len(originalObj.Properties), len(clonedObj.Properties))
			}

			for key, originalProp := range originalObj.Properties {
				clonedProp, exists := clonedObj.Properties[key]
				if !exists {
					t.Errorf("Property '%s' missing in cloned object", key)
					continue
				}

				// Properties should have same value but be different instances
				if clonedProp == originalProp {
					t.Errorf("Property '%s' is same instance in clone (should be copied)", key)
				}

				if clonedProp.Type != originalProp.Type {
					t.Errorf("Property '%s' type mismatch: expected %v, got %v", key, originalProp.Type, clonedProp.Type)
				}

				if clonedProp.Data != originalProp.Data {
					t.Errorf("Property '%s' value mismatch: expected %v, got %v", key, originalProp.Data, clonedProp.Data)
				}
			}
		})
	}
}

func TestCloneObjectWithNestedObjects(t *testing.T) {
	// Create nested object structure
	innerObject := &values.Value{
		Type: values.TypeObject,
		Data: values.Object{
			ClassName: "InnerClass",
			Properties: map[string]*values.Value{
				"value": values.NewString("inner value"),
			},
		},
	}

	outerObject := &values.Value{
		Type: values.TypeObject,
		Data: values.Object{
			ClassName: "OuterClass",
			Properties: map[string]*values.Value{
				"name":  values.NewString("outer"),
				"inner": innerObject,
			},
		},
	}

	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	ctx.Temporaries = make(map[uint32]*values.Value)
	ctx.Temporaries[0] = outerObject

	op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)
	inst := opcodes.Instruction{
		Opcode:  opcodes.OP_CLONE,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     0,
		Result:  1,
	}

	err := vm.executeClone(ctx, &inst)
	if err != nil {
		t.Fatalf("Clone failed: %v", err)
	}

	result := ctx.Temporaries[1]
	if result == nil || !result.IsObject() {
		t.Fatal("Clone result should be an object")
	}

	clonedObj := result.Data.(values.Object)
	originalObj := outerObject.Data.(values.Object)

	// Check outer object properties
	clonedInner := clonedObj.Properties["inner"]
	originalInner := originalObj.Properties["inner"]

	// Verify deep cloning - inner objects should be different instances
	if clonedInner == originalInner {
		t.Error("Inner object should be deep cloned (different instance)")
	}

	// But should have same values
	clonedInnerObj := clonedInner.Data.(values.Object)
	originalInnerObj := originalInner.Data.(values.Object)

	if clonedInnerObj.ClassName != originalInnerObj.ClassName {
		t.Error("Inner object class name should be preserved")
	}

	clonedInnerValue := clonedInnerObj.Properties["value"]
	originalInnerValue := originalInnerObj.Properties["value"]

	if clonedInnerValue.Data.(string) != originalInnerValue.Data.(string) {
		t.Error("Inner object property value should be preserved")
	}
}

func TestCloneObjectWithArrays(t *testing.T) {
	// Create object with array property
	array := values.NewArray()
	array.ArraySet(values.NewString("0"), values.NewString("item1"))
	array.ArraySet(values.NewString("1"), values.NewString("item2"))

	objectWithArray := &values.Value{
		Type: values.TypeObject,
		Data: values.Object{
			ClassName: "ArrayClass",
			Properties: map[string]*values.Value{
				"items": array,
				"count": values.NewInt(2),
			},
		},
	}

	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	ctx.Temporaries = make(map[uint32]*values.Value)
	ctx.Temporaries[0] = objectWithArray

	op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)
	inst := opcodes.Instruction{
		Opcode:  opcodes.OP_CLONE,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     0,
		Result:  1,
	}

	err := vm.executeClone(ctx, &inst)
	if err != nil {
		t.Fatalf("Clone failed: %v", err)
	}

	result := ctx.Temporaries[1]
	if result == nil || !result.IsObject() {
		t.Fatal("Clone result should be an object")
	}

	clonedObj := result.Data.(values.Object)
	originalObj := objectWithArray.Data.(values.Object)

	// Check array property is deep cloned
	clonedItems := clonedObj.Properties["items"]
	originalItems := originalObj.Properties["items"]

	if clonedItems == originalItems {
		t.Error("Array property should be deep cloned (different instance)")
	}

	// Verify array contents are preserved
	clonedArray := clonedItems.Data.(*values.Array)
	originalArray := originalItems.Data.(*values.Array)

	if len(clonedArray.Elements) != len(originalArray.Elements) {
		t.Error("Array elements count should be preserved")
	}

	for key, originalElement := range originalArray.Elements {
		clonedElement, exists := clonedArray.Elements[key]
		if !exists {
			t.Errorf("Array element '%s' missing in clone", key)
			continue
		}

		// Elements should be different instances but same values
		if clonedElement == originalElement {
			t.Errorf("Array element '%s' should be deep cloned", key)
		}

		if clonedElement.Data.(string) != originalElement.Data.(string) {
			t.Errorf("Array element '%s' value mismatch", key)
		}
	}
}

func TestCloneEmptyObject(t *testing.T) {
	emptyObject := &values.Value{
		Type: values.TypeObject,
		Data: values.Object{
			ClassName:  "EmptyClass",
			Properties: make(map[string]*values.Value),
		},
	}

	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	ctx.Temporaries = make(map[uint32]*values.Value)
	ctx.Temporaries[0] = emptyObject

	op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)
	inst := opcodes.Instruction{
		Opcode:  opcodes.OP_CLONE,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     0,
		Result:  1,
	}

	err := vm.executeClone(ctx, &inst)
	if err != nil {
		t.Fatalf("Clone of empty object failed: %v", err)
	}

	result := ctx.Temporaries[1]
	if result == nil || !result.IsObject() {
		t.Fatal("Clone result should be an object")
	}

	clonedObj := result.Data.(values.Object)
	if clonedObj.ClassName != "EmptyClass" {
		t.Error("Cloned empty object should preserve class name")
	}

	if len(clonedObj.Properties) != 0 {
		t.Error("Cloned empty object should have no properties")
	}
}

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

// TestPerformanceMetrics tests the performance metrics tracking
func TestPerformanceMetrics(t *testing.T) {
	metrics := NewPerformanceMetrics()

	// Record some instructions
	metrics.RecordInstruction("OP_ADD")
	metrics.RecordInstruction("OP_ADD")
	metrics.RecordInstruction("OP_ECHO")
	metrics.RecordInstruction("OP_ADD")

	// Record function calls
	metrics.RecordFunctionCall("test_function")
	metrics.RecordFunctionCall("another_function")
	metrics.RecordFunctionCall("test_function")

	// Record memory allocations
	metrics.RecordMemoryAllocation(1024)
	metrics.RecordMemoryAllocation(2048)
	metrics.RecordMemoryDeallocation(512)

	// Verify metrics
	require.Equal(t, uint64(4), metrics.TotalInstructions)
	require.Equal(t, uint64(3), metrics.InstructionCounts["OP_ADD"])
	require.Equal(t, uint64(1), metrics.InstructionCounts["OP_ECHO"])
	require.Equal(t, uint64(2), metrics.FunctionCallCounts["test_function"])
	require.Equal(t, uint64(1), metrics.FunctionCallCounts["another_function"])
	require.Equal(t, uint64(2), metrics.MemoryAllocations)
	require.Equal(t, uint64(1), metrics.MemoryDeallocations)
	require.Equal(t, uint64(2560), metrics.CurrentMemoryUsage) // 1024 + 2048 - 512
	require.Equal(t, uint64(3072), metrics.PeakMemoryUsage)    // 1024 + 2048

	// Test report generation
	report := metrics.GetReport()
	require.Contains(t, report, "Total Instructions: 4")
	require.Contains(t, report, "OP_ADD: 3")
	require.Contains(t, report, "test_function: 2 calls")
}

// TestDebugger tests the debugging functionality
func TestDebugger(t *testing.T) {
	debugger := NewDebugger(DebugLevelDetailed, nil)

	// Test breakpoint management
	debugger.SetBreakpoint(100)
	debugger.SetBreakpoint(200)

	require.True(t, debugger.ShouldBreak(100))
	require.True(t, debugger.ShouldBreak(200))
	require.False(t, debugger.ShouldBreak(300))

	debugger.RemoveBreakpoint(100)
	require.False(t, debugger.ShouldBreak(100))
	require.True(t, debugger.ShouldBreak(200))

	// Test variable watching
	debugger.WatchVariable("$testVar")
	require.True(t, debugger.WatchVariables["$testVar"])

	// Test instruction tracing
	ctx := &ExecutionContext{
		Variables: make(map[uint32]*values.Value),
		SP:        5,
	}
	ctx.Variables[0] = values.NewInt(42)

	inst := &opcodes.Instruction{
		Opcode: opcodes.OP_ADD,
		Op1:    1,
		Op2:    2,
		Result: 3,
	}

	debugger.TraceInstruction(10, inst, ctx, time.Microsecond*100)

	require.Len(t, debugger.InstructionLog, 1)
	trace := debugger.InstructionLog[0]
	require.Equal(t, 10, trace.IP)
	require.Equal(t, opcodes.OP_ADD, trace.Instruction.Opcode)
	require.Equal(t, 5, trace.StackSize)
	require.Equal(t, time.Microsecond*100, trace.Duration)

	// Test function call tracing
	args := []*values.Value{values.NewString("arg1"), values.NewInt(42)}
	debugger.TraceFunctionCall("testFunction", args, 0)

	require.Len(t, debugger.CallStack, 1)
	callTrace := debugger.CallStack[0]
	require.Equal(t, "testFunction", callTrace.FunctionName)
	require.Len(t, callTrace.Arguments, 2)
	require.Equal(t, 0, callTrace.CallDepth)

	// Test function return tracing
	returnValue := values.NewString("result")
	debugger.TraceFunctionReturn("testFunction", returnValue, 0)

	// Get the updated call trace
	updatedCallTrace := debugger.CallStack[0]
	require.Equal(t, returnValue, updatedCallTrace.ReturnValue)
	require.True(t, updatedCallTrace.Duration > 0)

	// Test report generation
	report := debugger.GenerateReport()
	require.Contains(t, report, "Instructions traced: 1")
	require.Contains(t, report, "Function calls: 1")
}

// TestVMOptimizer tests the VM optimization features
func TestVMOptimizer(t *testing.T) {
	optimizer := NewVMOptimizer()

	// Record some hot spots
	optimizer.RecordHotSpot(100)
	optimizer.RecordHotSpot(200)
	optimizer.RecordHotSpot(100) // Duplicate
	optimizer.RecordHotSpot(100) // Another duplicate
	optimizer.RecordHotSpot(300)
	optimizer.RecordHotSpot(200) // Duplicate

	// Test hot spot detection
	require.True(t, optimizer.IsHotSpot(100, 2))  // 3 executions >= 2
	require.True(t, optimizer.IsHotSpot(200, 2))  // 2 executions >= 2
	require.False(t, optimizer.IsHotSpot(300, 2)) // 1 execution < 2

	// Test getting hot spots
	hotSpots := optimizer.GetHotSpots(2)
	require.Len(t, hotSpots, 2)

	// Should be sorted by count (descending)
	require.Equal(t, 100, hotSpots[0].IP)
	require.Equal(t, uint64(3), hotSpots[0].Count)
	require.Equal(t, 200, hotSpots[1].IP)
	require.Equal(t, uint64(2), hotSpots[1].Count)
}

// TestMemoryPool tests the memory pool functionality
func TestMemoryPool(t *testing.T) {
	pool := NewMemoryPool()

	// Test value pooling
	val1 := pool.GetValue()
	val2 := pool.GetValue()
	require.NotNil(t, val1)
	require.NotNil(t, val2)

	// Set some values
	val1.Type = values.TypeInt
	val1.Data = int64(42)
	val2.Type = values.TypeString
	val2.Data = "test"

	// Return to pool
	pool.PutValue(val1)
	pool.PutValue(val2)

	// Get again (should be recycled)
	val3 := pool.GetValue()
	require.NotNil(t, val3)
	// Value should be reset
	require.Equal(t, values.ValueType(0), val3.Type)
	require.Nil(t, val3.Data)

	// Test statistics
	allocs, deallocs := pool.GetStats()
	require.Equal(t, uint64(3), allocs)   // val1, val2, val3
	require.Equal(t, uint64(2), deallocs) // val1, val2 returned

	// Test execution context pooling
	ctx1 := pool.GetExecutionContext()
	ctx2 := pool.GetExecutionContext()
	require.NotNil(t, ctx1)
	require.NotNil(t, ctx2)

	// Add some data
	ctx1.Variables[100] = values.NewInt(123)
	ctx1.Temporaries[200] = values.NewString("temp")

	// Return to pool
	pool.PutExecutionContext(ctx1)

	// Get again
	ctx3 := pool.GetExecutionContext()
	require.NotNil(t, ctx3)
	// Maps should be cleared
	require.Len(t, ctx3.Variables, 0)
	require.Len(t, ctx3.Temporaries, 0)
}

// TestEnhancedVMIntegration tests the integration of enhanced features with the VM
func TestEnhancedVMIntegration(t *testing.T) {
	// Create VM with profiling enabled
	vm := NewVirtualMachineWithProfiling(DebugLevelBasic)

	require.True(t, vm.EnableProfiling)
	require.True(t, vm.Debugger.ProfilerEnabled)
	require.Equal(t, DebugLevelBasic, vm.Debugger.Level)
	require.NotNil(t, vm.Metrics)
	require.NotNil(t, vm.Optimizer)
	require.NotNil(t, vm.MemoryPool)

	// Test utility methods
	vm.SetDebugLevel(DebugLevelVerbose)
	require.Equal(t, DebugLevelVerbose, vm.Debugger.Level)

	vm.SetBreakpoint(42)
	require.True(t, vm.Debugger.BreakPoints[42])

	vm.WatchVariable("$testVar")
	require.True(t, vm.Debugger.WatchVariables["$testVar"])

	// Test advanced profiling enablement
	vm.EnableAdvancedProfiling()
	require.True(t, vm.EnableProfiling)
	require.True(t, vm.DebugMode)
	require.Equal(t, DebugLevelDetailed, vm.Debugger.Level)
	require.True(t, vm.Debugger.ProfilerEnabled)

	// Test reports (should not crash with empty data)
	perfReport := vm.GetPerformanceReport()
	require.Contains(t, perfReport, "VM Performance Report")

	debugReport := vm.GetDebugReport()
	require.Contains(t, debugReport, "VM Debugger Report")

	// Test memory stats
	allocs, deallocs := vm.GetMemoryStats()
	require.Equal(t, uint64(0), allocs)   // No allocations yet
	require.Equal(t, uint64(0), deallocs) // No deallocations yet

	// Test hot spots (should be empty initially)
	hotSpots := vm.GetHotSpots(10)
	require.Len(t, hotSpots, 0)
}

// TestProfileDataAnalysis tests analysis of profiling data
func TestProfileDataAnalysis(t *testing.T) {
	profileData := &ProfileData{
		FunctionProfiles: make(map[string]*FunctionProfile),
		InstructionTimes: make(map[string]time.Duration),
		HotPaths:         make([]HotPath, 0),
		MemoryProfile: &MemoryProfile{
			AllocationsPerType:   make(map[string]uint64),
			DeallocationsPerType: make(map[string]uint64),
			PeakUsagePerType:     make(map[string]uint64),
			LeakDetection:        make(map[string]uint64),
		},
	}

	// Add some function profiles
	profileData.FunctionProfiles["function1"] = &FunctionProfile{
		Name:        "function1",
		CallCount:   100,
		TotalTime:   time.Millisecond * 500,
		AverageTime: time.Microsecond * 5000, // 5ms
		MinTime:     time.Microsecond * 1000, // 1ms
		MaxTime:     time.Millisecond * 50,   // 50ms
	}

	profileData.FunctionProfiles["function2"] = &FunctionProfile{
		Name:        "function2",
		CallCount:   50,
		TotalTime:   time.Millisecond * 1000, // 1s
		AverageTime: time.Millisecond * 20,   // 20ms
		MinTime:     time.Millisecond * 5,    // 5ms
		MaxTime:     time.Millisecond * 100,  // 100ms
	}

	// Add instruction timings
	profileData.InstructionTimes["OP_ADD"] = time.Millisecond * 100
	profileData.InstructionTimes["OP_ECHO"] = time.Millisecond * 50
	profileData.InstructionTimes["OP_CALL"] = time.Millisecond * 300

	// Add memory profile data
	profileData.MemoryProfile.AllocationsPerType["Value"] = 1000
	profileData.MemoryProfile.AllocationsPerType["Array"] = 200
	profileData.MemoryProfile.DeallocationsPerType["Value"] = 950
	profileData.MemoryProfile.DeallocationsPerType["Array"] = 180
	profileData.MemoryProfile.PeakUsagePerType["Value"] = 50 * 1024  // 50KB
	profileData.MemoryProfile.PeakUsagePerType["Array"] = 100 * 1024 // 100KB
	profileData.MemoryProfile.LeakDetection["Value"] = 50            // 50 leaked values
	profileData.MemoryProfile.LeakDetection["Array"] = 20            // 20 leaked arrays

	// Verify the data
	require.Len(t, profileData.FunctionProfiles, 2)
	require.Len(t, profileData.InstructionTimes, 3)

	function1 := profileData.FunctionProfiles["function1"]
	require.Equal(t, uint64(100), function1.CallCount)
	require.Equal(t, time.Millisecond*500, function1.TotalTime)

	function2 := profileData.FunctionProfiles["function2"]
	require.Equal(t, uint64(50), function2.CallCount)
	require.Equal(t, time.Millisecond*1000, function2.TotalTime)

	require.Equal(t, time.Millisecond*100, profileData.InstructionTimes["OP_ADD"])
	require.Equal(t, time.Millisecond*300, profileData.InstructionTimes["OP_CALL"])

	// Test memory analysis
	require.Equal(t, uint64(1000), profileData.MemoryProfile.AllocationsPerType["Value"])
	require.Equal(t, uint64(950), profileData.MemoryProfile.DeallocationsPerType["Value"])
	require.Equal(t, uint64(50), profileData.MemoryProfile.LeakDetection["Value"])
}

// BenchmarkVMPerformance benchmarks VM execution with and without profiling
func BenchmarkVMPerformance(b *testing.B) {
	// Create a simple instruction sequence
	instructions := []opcodes.Instruction{
		{Opcode: opcodes.OP_QM_ASSIGN, Op1: 0, Op2: 0, Result: 100},
		{Opcode: opcodes.OP_QM_ASSIGN, Op1: 1, Op2: 0, Result: 101},
		{Opcode: opcodes.OP_ADD, Op1: 100, Op2: 101, Result: 102},
		{Opcode: opcodes.OP_ECHO, Op1: 102, Op2: 0, Result: 0},
	}

	constants := []*values.Value{
		values.NewInt(10),
		values.NewInt(20),
	}

	b.Run("WithoutProfiling", func(b *testing.B) {
		vm := NewVirtualMachine()
		for i := 0; i < b.N; i++ {
			ctx := NewExecutionContext()
			ctx.SetOutputWriter(&strings.Builder{}) // Discard output
			vm.Execute(ctx, instructions, constants, nil, nil)
		}
	})

	b.Run("WithProfiling", func(b *testing.B) {
		vm := NewVirtualMachineWithProfiling(DebugLevelNone)
		for i := 0; i < b.N; i++ {
			ctx := NewExecutionContext()
			ctx.SetOutputWriter(&strings.Builder{}) // Discard output
			vm.Execute(ctx, instructions, constants, nil, nil)
		}
	})

	b.Run("WithDetailedProfiling", func(b *testing.B) {
		vm := NewVirtualMachineWithProfiling(DebugLevelDetailed)
		for i := 0; i < b.N; i++ {
			ctx := NewExecutionContext()
			ctx.SetOutputWriter(&strings.Builder{}) // Discard output
			vm.Execute(ctx, instructions, constants, nil, nil)
		}
	})
}

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

func TestForeachFreeOpcode(t *testing.T) {
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	// Setup a foreach iterator in slot 5
	iteratorSlot := uint32(5)

	// Create a mock iterator
	ctx.ForeachIterators[iteratorSlot] = &ForeachIterator{
		Array:   values.NewArray(),
		Index:   0,
		Keys:    []*values.Value{values.NewString("key1"), values.NewString("key2")},
		Values:  []*values.Value{values.NewString("value1"), values.NewString("value2")},
		HasMore: true,
	}

	// Also setup some temporary variables that might be associated
	ctx.Temporaries[iteratorSlot] = values.NewString("iterator_value")
	ctx.Temporaries[iteratorSlot+1] = values.NewString("iterator_key")

	// Verify setup
	if ctx.ForeachIterators[iteratorSlot] == nil {
		t.Fatal("Iterator should be set up")
	}
	if ctx.Temporaries[iteratorSlot] == nil {
		t.Fatal("Temporary value should be set up")
	}

	// Create FE_FREE instruction
	inst := &opcodes.Instruction{
		Opcode: opcodes.OP_FE_FREE,
		Op1:    iteratorSlot, // Iterator slot to free
		Op2:    0,            // Unused
		Result: 0,            // Unused
	}
	inst.OpType1, inst.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_UNUSED, opcodes.IS_UNUSED, opcodes.IS_UNUSED)

	err := vm.executeForeachFree(ctx, inst)
	if err != nil {
		t.Fatalf("executeForeachFree failed: %v", err)
	}

	// Check that iterator was removed
	if ctx.ForeachIterators[iteratorSlot] != nil {
		t.Error("Iterator should have been freed")
	}

	// Check that associated temporaries were cleaned up
	if ctx.Temporaries[iteratorSlot] != nil {
		t.Error("Iterator temporary value should have been cleaned up")
	}
	if ctx.Temporaries[iteratorSlot+1] != nil {
		t.Error("Iterator temporary key should have been cleaned up")
	}
}

func TestForeachFreeWithEmptyContext(t *testing.T) {
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	// Don't set up any iterators or temporaries
	iteratorSlot := uint32(10)

	// Create FE_FREE instruction
	inst := &opcodes.Instruction{
		Opcode: opcodes.OP_FE_FREE,
		Op1:    iteratorSlot,
		Op2:    0,
		Result: 0,
	}
	inst.OpType1, inst.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_UNUSED, opcodes.IS_UNUSED, opcodes.IS_UNUSED)

	// This should not crash even if there's nothing to free
	err := vm.executeForeachFree(ctx, inst)
	if err != nil {
		t.Fatalf("executeForeachFree should not fail with empty context: %v", err)
	}
}

func TestEvalOpcode(t *testing.T) {
	tests := []struct {
		name         string
		evalCode     string
		expectedNull bool
		shouldError  bool
	}{
		{
			name:         "eval empty string",
			evalCode:     "",
			expectedNull: true,
			shouldError:  false,
		},
		{
			name:         "eval simple PHP code",
			evalCode:     "echo 'Hello World';",
			expectedNull: true, // Our stub implementation returns NULL
			shouldError:  false,
		},
		{
			name:         "eval return statement",
			evalCode:     "return 42;",
			expectedNull: true, // Our stub implementation returns NULL
			shouldError:  false,
		},
		{
			name:         "eval variable assignment",
			evalCode:     "$x = 10; return $x;",
			expectedNull: true, // Our stub implementation returns NULL
			shouldError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := NewVirtualMachine()
			ctx := NewExecutionContext()

			// Setup the code to eval
			ctx.Temporaries[1] = values.NewString(tt.evalCode)

			// Create EVAL instruction
			inst := &opcodes.Instruction{
				Opcode: opcodes.OP_EVAL,
				Op1:    1, // Code to evaluate
				Op2:    0, // Unused
				Result: 2, // Result location
			}
			inst.OpType1, inst.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)

			err := vm.executeEval(ctx, inst)

			if tt.shouldError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("executeEval failed: %v", err)
			}

			// Check result
			result := ctx.Temporaries[2]
			if result == nil {
				t.Fatal("Result should not be nil")
			}

			if tt.expectedNull {
				if !result.IsNull() {
					t.Errorf("Expected null result, got %v", result)
				}
			}
		})
	}
}

func TestEvalOpcodeWithNonStringCode(t *testing.T) {
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	// Setup non-string value to eval
	ctx.Temporaries[1] = values.NewInt(123)

	// Create EVAL instruction
	inst := &opcodes.Instruction{
		Opcode: opcodes.OP_EVAL,
		Op1:    1,
		Op2:    0,
		Result: 2,
	}
	inst.OpType1, inst.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)

	err := vm.executeEval(ctx, inst)
	if err == nil {
		t.Error("Expected error for non-string eval code")
	}

	expectedError := "EVAL requires string code to evaluate"
	if err.Error() != expectedError {
		t.Errorf("Expected error %q, got %q", expectedError, err.Error())
	}
}

func TestEvalOpcodeWithNullCode(t *testing.T) {
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	// Setup null value to eval
	ctx.Temporaries[1] = values.NewNull()

	// Create EVAL instruction
	inst := &opcodes.Instruction{
		Opcode: opcodes.OP_EVAL,
		Op1:    1,
		Op2:    0,
		Result: 2,
	}
	inst.OpType1, inst.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)

	err := vm.executeEval(ctx, inst)
	if err == nil {
		t.Error("Expected error for null eval code")
	}
}

func TestForeachCleanupFlow(t *testing.T) {
	// This test simulates a typical foreach loop cleanup scenario
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	// Setup multiple iterators as would happen in nested foreach loops
	for i := uint32(0); i < 3; i++ {
		ctx.ForeachIterators[i] = &ForeachIterator{
			Array:   values.NewArray(),
			Index:   int(i),
			Keys:    []*values.Value{},
			Values:  []*values.Value{values.NewInt(int64(i))},
			HasMore: false,
		}
		ctx.Temporaries[i] = values.NewInt(int64(i * 10))
		ctx.Temporaries[i+10] = values.NewString("key_" + string(rune('0'+i)))
	}

	// Free iterators in reverse order (as would happen when exiting nested loops)
	for i := uint32(2); i >= 0 && i <= 2; i-- { // Handle underflow
		inst := &opcodes.Instruction{
			Opcode: opcodes.OP_FE_FREE,
			Op1:    i,
			Op2:    0,
			Result: 0,
		}
		inst.OpType1, inst.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_UNUSED, opcodes.IS_UNUSED, opcodes.IS_UNUSED)

		err := vm.executeForeachFree(ctx, inst)
		if err != nil {
			t.Fatalf("Failed to free iterator %d: %v", i, err)
		}

		// Verify this iterator was freed
		if ctx.ForeachIterators[i] != nil {
			t.Errorf("Iterator %d should have been freed", i)
		}
	}

	// Verify all iterators were freed
	if len(ctx.ForeachIterators) > 0 {
		t.Errorf("Expected all iterators to be freed, but %d remain", len(ctx.ForeachIterators))
	}
}

func TestInitFunctionCallByNameOpcode(t *testing.T) {
	tests := []struct {
		name         string
		functionName string
		argCount     *values.Value
		expectError  bool
	}{
		{
			name:         "init simple function call",
			functionName: "strlen",
			argCount:     nil,
			expectError:  false,
		},
		{
			name:         "init function call with arg count",
			functionName: "substr",
			argCount:     values.NewInt(3),
			expectError:  false,
		},
		{
			name:         "init function call with complex name",
			functionName: "MyClass::staticMethod",
			argCount:     values.NewInt(2),
			expectError:  false,
		},
		{
			name:         "init function call with namespace",
			functionName: "\\Namespace\\function_name",
			argCount:     values.NewInt(1),
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := NewVirtualMachine()
			ctx := NewExecutionContext()

			// Setup function name
			ctx.Temporaries[1] = values.NewString(tt.functionName)

			// Create instruction
			inst := &opcodes.Instruction{
				Opcode: opcodes.OP_INIT_FCALL_BY_NAME,
				Op1:    1, // Function name
				Op2:    0, // Arg count (optional)
				Result: 0, // Unused
			}

			if tt.argCount != nil {
				ctx.Temporaries[2] = tt.argCount
				inst.Op2 = 2
			}

			inst.OpType1, inst.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR, opcodes.IS_UNUSED)

			err := vm.executeInitFunctionCallByName(ctx, inst)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("executeInitFunctionCallByName failed: %v", err)
			}

			// Check that CallContext was set up
			if ctx.CallContext == nil {
				t.Fatal("CallContext should be initialized")
			}

			if ctx.CallContext.FunctionName != tt.functionName {
				t.Errorf("Expected function name %q, got %q", tt.functionName, ctx.CallContext.FunctionName)
			}

			// Check argument count if specified
			if tt.argCount != nil {
				expectedCount := int(tt.argCount.ToInt())
				if ctx.CallContext.NumArgs != expectedCount {
					t.Errorf("Expected arg count %d, got %d", expectedCount, ctx.CallContext.NumArgs)
				}
			}

			// Check that call arguments were cleared
			if ctx.CallArguments != nil && len(ctx.CallArguments) > 0 {
				t.Error("Call arguments should be cleared")
			}
		})
	}
}

func TestInitFunctionCallByNameWithInvalidName(t *testing.T) {
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	// Setup non-string function name
	ctx.Temporaries[1] = values.NewInt(123)

	inst := &opcodes.Instruction{
		Opcode: opcodes.OP_INIT_FCALL_BY_NAME,
		Op1:    1,
		Op2:    0,
		Result: 0,
	}
	inst.OpType1, inst.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_UNUSED)

	err := vm.executeInitFunctionCallByName(ctx, inst)
	if err == nil {
		t.Error("Expected error for non-string function name")
	}

	expectedError := "INIT_FCALL_BY_NAME requires string function name"
	if err.Error() != expectedError {
		t.Errorf("Expected error %q, got %q", expectedError, err.Error())
	}
}

func TestReturnByRefOpcode(t *testing.T) {
	tests := []struct {
		name             string
		returnValue      *values.Value
		hasCallStack     bool
		expectedHalted   bool
		expectedExitCode int
	}{
		{
			name:             "return by ref with null value",
			returnValue:      values.NewNull(),
			hasCallStack:     false,
			expectedHalted:   true,
			expectedExitCode: 0,
		},
		{
			name:             "return by ref with string value",
			returnValue:      values.NewString("test"),
			hasCallStack:     false,
			expectedHalted:   true,
			expectedExitCode: 0,
		},
		{
			name:             "return by ref with int value",
			returnValue:      values.NewInt(42),
			hasCallStack:     false,
			expectedHalted:   true,
			expectedExitCode: 0,
		},
		{
			name:             "return by ref in function",
			returnValue:      values.NewString("function_result"),
			hasCallStack:     true,
			expectedHalted:   false,
			expectedExitCode: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := NewVirtualMachine()
			ctx := NewExecutionContext()

			// Setup return value
			ctx.Temporaries[1] = tt.returnValue

			// Setup call stack if needed
			if tt.hasCallStack {
				ctx.CallStack = append(ctx.CallStack, CallFrame{
					Function:    nil, // Mock function
					ReturnIP:    100,
					Variables:   make(map[uint32]*values.Value),
					ThisObject:  nil,
					Arguments:   nil,
					ReturnValue: nil,
					ReturnByRef: false,
				})
			}

			originalStackSize := len(ctx.CallStack)

			// Create RETURN_BY_REF instruction
			inst := &opcodes.Instruction{
				Opcode: opcodes.OP_RETURN_BY_REF,
				Op1:    1, // Return value
				Op2:    0, // Unused
				Result: 0, // Unused
			}
			inst.OpType1, inst.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_UNUSED)

			err := vm.executeReturnByRef(ctx, inst)
			if err != nil {
				t.Fatalf("executeReturnByRef failed: %v", err)
			}

			// Check execution state
			if ctx.Halted != tt.expectedHalted {
				t.Errorf("Expected halted=%t, got halted=%t", tt.expectedHalted, ctx.Halted)
			}

			if ctx.ExitCode != tt.expectedExitCode {
				t.Errorf("Expected exit code %d, got %d", tt.expectedExitCode, ctx.ExitCode)
			}

			// Check call stack behavior
			if tt.hasCallStack {
				// Should have popped the call stack
				if len(ctx.CallStack) != originalStackSize-1 {
					t.Errorf("Expected call stack size %d, got %d", originalStackSize-1, len(ctx.CallStack))
				}

				// The return value and reference flag should be set on the popped frame
				// (In a real implementation, this would be available to the caller)
			}
		})
	}
}

func TestReturnByRefWithUndefinedValue(t *testing.T) {
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	// Don't set up any return value (should default to NULL)

	inst := &opcodes.Instruction{
		Opcode: opcodes.OP_RETURN_BY_REF,
		Op1:    1, // Undefined temporary
		Op2:    0,
		Result: 0,
	}
	inst.OpType1, inst.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_UNUSED)

	err := vm.executeReturnByRef(ctx, inst)
	if err != nil {
		t.Fatalf("executeReturnByRef should not fail with undefined value: %v", err)
	}

	// Should halt execution (global return)
	if !ctx.Halted {
		t.Error("Expected execution to be halted")
	}
}

func TestFunctionCallSequence(t *testing.T) {
	// This test simulates a typical function call sequence
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	// Step 1: Initialize function call
	ctx.Temporaries[1] = values.NewString("test_function")
	ctx.Temporaries[2] = values.NewInt(2) // 2 arguments

	initInst := &opcodes.Instruction{
		Opcode: opcodes.OP_INIT_FCALL_BY_NAME,
		Op1:    1,
		Op2:    2,
		Result: 0,
	}
	initInst.OpType1, initInst.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR, opcodes.IS_UNUSED)

	err := vm.executeInitFunctionCallByName(ctx, initInst)
	if err != nil {
		t.Fatalf("Failed to initialize function call: %v", err)
	}

	// Verify call context is set up
	if ctx.CallContext == nil {
		t.Fatal("Call context should be initialized")
	}
	if ctx.CallContext.FunctionName != "test_function" {
		t.Error("Function name not set correctly")
	}
	if ctx.CallContext.NumArgs != 2 {
		t.Error("Argument count not set correctly")
	}

	// Step 2: In a real scenario, SEND_VAL/SEND_VAR opcodes would add arguments
	// For this test, we'll simulate having arguments ready
	ctx.CallArguments = []*values.Value{
		values.NewString("arg1"),
		values.NewInt(42),
	}

	// Step 3: Execute the actual function call (this would normally be DO_FCALL)
	// We'll just simulate a successful call that pushes a call frame

	callFrame := CallFrame{
		Function:    nil, // Mock function
		ReturnIP:    ctx.IP + 1,
		Variables:   make(map[uint32]*values.Value),
		ThisObject:  nil,
		Arguments:   ctx.CallArguments,
		ReturnValue: nil,
		ReturnByRef: false,
	}
	ctx.CallStack = append(ctx.CallStack, callFrame)

	// Step 4: Return by reference from the function
	ctx.Temporaries[3] = values.NewString("result_value")

	returnInst := &opcodes.Instruction{
		Opcode: opcodes.OP_RETURN_BY_REF,
		Op1:    3,
		Op2:    0,
		Result: 0,
	}
	returnInst.OpType1, returnInst.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_UNUSED)

	err = vm.executeReturnByRef(ctx, returnInst)
	if err != nil {
		t.Fatalf("Failed to return by reference: %v", err)
	}

	// Verify that the call stack was popped
	if len(ctx.CallStack) != 0 {
		t.Error("Call stack should have been popped")
	}

	// Verify that execution is not halted (we returned from a function, not globally)
	if ctx.Halted {
		t.Error("Execution should not be halted when returning from function")
	}
}

func TestIncludeOpcode(t *testing.T) {
	// Create test files
	testDir := t.TempDir()
	testFile := filepath.Join(testDir, "test.php")
	testContent := []byte("<?php echo 'Hello from included file'; ?>")

	err := os.WriteFile(testFile, testContent, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name          string
		filepath      string
		expectSuccess bool
		expectResult  bool
	}{
		{"include existing file", testFile, true, true},
		{"include non-existent file", "/non/existent/file.php", false, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vm := NewVirtualMachine()
			ctx := NewExecutionContext()

			ctx.Temporaries = make(map[uint32]*values.Value)
			ctx.Temporaries[0] = values.NewString(test.filepath)

			op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)
			inst := opcodes.Instruction{
				Opcode:  opcodes.OP_INCLUDE,
				OpType1: op1Type,
				OpType2: op2Type,
				Op1:     0,
				Result:  1,
			}

			err := vm.executeInclude(ctx, &inst)
			if err != nil {
				t.Fatalf("INCLUDE execution failed: %v", err)
			}

			// Check result
			result := ctx.Temporaries[1]
			if result == nil {
				t.Fatal("INCLUDE result is nil")
			}

			if test.expectSuccess {
				// Should return file size (int)
				if !result.IsInt() {
					t.Errorf("Expected int result for successful include, got %v", result)
				}
				expectedSize := int64(len(testContent))
				if result.Data.(int64) != expectedSize {
					t.Errorf("Expected result %d, got %d", expectedSize, result.Data.(int64))
				}
			} else {
				// Should return false for failed include
				if !result.IsBool() || result.Data.(bool) != false {
					t.Errorf("Expected false for failed include, got %v", result)
				}
			}
		})
	}
}

func TestIncludeOnceOpcode(t *testing.T) {
	// Create test files
	testDir := t.TempDir()
	testFile := filepath.Join(testDir, "test_once.php")
	testContent := []byte("<?php echo 'Hello from included once file'; ?>")

	err := os.WriteFile(testFile, testContent, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	ctx.Temporaries = make(map[uint32]*values.Value)
	ctx.Temporaries[0] = values.NewString(testFile)

	op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)

	// First include_once - should succeed
	inst1 := opcodes.Instruction{
		Opcode:  opcodes.OP_INCLUDE_ONCE,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     0,
		Result:  1,
	}

	err = vm.executeIncludeOnce(ctx, &inst1)
	if err != nil {
		t.Fatalf("First INCLUDE_ONCE execution failed: %v", err)
	}

	result1 := ctx.Temporaries[1]
	if result1 == nil || !result1.IsInt() {
		t.Fatalf("First INCLUDE_ONCE should return int, got %v", result1)
	}

	// Second include_once - should return true (already included)
	inst2 := opcodes.Instruction{
		Opcode:  opcodes.OP_INCLUDE_ONCE,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     0,
		Result:  2,
	}

	err = vm.executeIncludeOnce(ctx, &inst2)
	if err != nil {
		t.Fatalf("Second INCLUDE_ONCE execution failed: %v", err)
	}

	result2 := ctx.Temporaries[2]
	if result2 == nil || !result2.IsBool() || result2.Data.(bool) != true {
		t.Errorf("Second INCLUDE_ONCE should return true, got %v", result2)
	}
}

func TestRequireOpcode(t *testing.T) {
	// Create test files
	testDir := t.TempDir()
	testFile := filepath.Join(testDir, "test_require.php")
	testContent := []byte("<?php echo 'Hello from required file'; ?>")

	err := os.WriteFile(testFile, testContent, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name          string
		filepath      string
		expectSuccess bool
		expectError   bool
	}{
		{"require existing file", testFile, true, false},
		{"require non-existent file", "/non/existent/require.php", false, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vm := NewVirtualMachine()
			ctx := NewExecutionContext()

			ctx.Temporaries = make(map[uint32]*values.Value)
			ctx.Temporaries[0] = values.NewString(test.filepath)

			op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)
			inst := opcodes.Instruction{
				Opcode:  opcodes.OP_REQUIRE,
				OpType1: op1Type,
				OpType2: op2Type,
				Op1:     0,
				Result:  1,
			}

			err := vm.executeRequire(ctx, &inst)

			if test.expectError {
				if err == nil {
					t.Fatal("Expected REQUIRE to fail, but it succeeded")
				}
				// Error expected, test passed
				return
			}

			if err != nil {
				t.Fatalf("REQUIRE execution failed: %v", err)
			}

			// Check result
			result := ctx.Temporaries[1]
			if result == nil {
				t.Fatal("REQUIRE result is nil")
			}

			if test.expectSuccess {
				// Should return file size (int)
				if !result.IsInt() {
					t.Errorf("Expected int result for successful require, got %v", result)
				}
				expectedSize := int64(len(testContent))
				if result.Data.(int64) != expectedSize {
					t.Errorf("Expected result %d, got %d", expectedSize, result.Data.(int64))
				}
			}
		})
	}
}

func TestRequireOnceOpcode(t *testing.T) {
	// Create test files
	testDir := t.TempDir()
	testFile := filepath.Join(testDir, "test_require_once.php")
	testContent := []byte("<?php echo 'Hello from required once file'; ?>")

	err := os.WriteFile(testFile, testContent, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	ctx.Temporaries = make(map[uint32]*values.Value)
	ctx.Temporaries[0] = values.NewString(testFile)

	op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)

	// First require_once - should succeed
	inst1 := opcodes.Instruction{
		Opcode:  opcodes.OP_REQUIRE_ONCE,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     0,
		Result:  1,
	}

	err = vm.executeRequireOnce(ctx, &inst1)
	if err != nil {
		t.Fatalf("First REQUIRE_ONCE execution failed: %v", err)
	}

	result1 := ctx.Temporaries[1]
	if result1 == nil || !result1.IsInt() {
		t.Fatalf("First REQUIRE_ONCE should return int, got %v", result1)
	}

	// Second require_once - should return true (already included)
	inst2 := opcodes.Instruction{
		Opcode:  opcodes.OP_REQUIRE_ONCE,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     0,
		Result:  2,
	}

	err = vm.executeRequireOnce(ctx, &inst2)
	if err != nil {
		t.Fatalf("Second REQUIRE_ONCE execution failed: %v", err)
	}

	result2 := ctx.Temporaries[2]
	if result2 == nil || !result2.IsBool() || result2.Data.(bool) != true {
		t.Errorf("Second REQUIRE_ONCE should return true, got %v", result2)
	}
}

func TestRequireOnceFailure(t *testing.T) {
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	ctx.Temporaries = make(map[uint32]*values.Value)
	ctx.Temporaries[0] = values.NewString("/non/existent/require_once.php")

	op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)
	inst := opcodes.Instruction{
		Opcode:  opcodes.OP_REQUIRE_ONCE,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     0,
		Result:  1,
	}

	err := vm.executeRequireOnce(ctx, &inst)
	if err == nil {
		t.Fatal("Expected REQUIRE_ONCE to fail for non-existent file")
	}

	// Should get proper error message
	if !strings.Contains(err.Error(), "require(") {
		t.Errorf("Expected require error message, got: %s", err.Error())
	}
}

func TestIncludeFilePath(t *testing.T) {
	// Test absolute path handling for once variants
	testDir := t.TempDir()
	testFile := filepath.Join(testDir, "path_test.php")
	testContent := []byte("<?php // test ?>")

	err := os.WriteFile(testFile, testContent, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	// Test with relative path (should be converted to absolute)
	relPath := filepath.Base(testFile)

	// Change to test directory so relative path works
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(testDir)

	ctx.Temporaries = make(map[uint32]*values.Value)
	ctx.Temporaries[0] = values.NewString(relPath)

	op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)

	// First include_once with relative path
	inst1 := opcodes.Instruction{
		Opcode:  opcodes.OP_INCLUDE_ONCE,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     0,
		Result:  1,
	}

	err = vm.executeIncludeOnce(ctx, &inst1)
	if err != nil {
		t.Fatalf("First INCLUDE_ONCE with relative path failed: %v", err)
	}

	// Now try with absolute path - should still recognize as already included
	ctx.Temporaries[0] = values.NewString(testFile)
	inst2 := opcodes.Instruction{
		Opcode:  opcodes.OP_INCLUDE_ONCE,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     0,
		Result:  2,
	}

	err = vm.executeIncludeOnce(ctx, &inst2)
	if err != nil {
		t.Fatalf("Second INCLUDE_ONCE with absolute path failed: %v", err)
	}

	result2 := ctx.Temporaries[2]
	if result2 == nil || !result2.IsBool() || result2.Data.(bool) != true {
		t.Errorf("Second INCLUDE_ONCE with absolute path should return true (already included), got %v", result2)
	}
}

func TestInitFunctionCallOpcode(t *testing.T) {
	// Initialize runtime for built-in function checks
	err := runtime.Bootstrap()
	require.NoError(t, err, "Failed to bootstrap runtime")

	tests := []struct {
		name         string
		numArgs      int
		functionName string
		expectError  bool
		errorMsg     string
	}{
		{
			name:         "init simple function call",
			numArgs:      1,
			functionName: "strlen",
			expectError:  false,
		},
		{
			name:         "init function call with multiple args",
			numArgs:      3,
			functionName: "substr",
			expectError:  false,
		},
		{
			name:         "init function call with no args",
			numArgs:      0,
			functionName: "phpversion",
			expectError:  false,
		},
		{
			name:         "init user function call",
			numArgs:      2,
			functionName: "custom_function",
			expectError:  false,
		},
		{
			name:         "init namespaced function call",
			numArgs:      1,
			functionName: "\\Namespace\\function_name",
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := NewVirtualMachine()
			ctx := NewExecutionContext()

			// Setup operands: functionName in Op1, numArgs in Op2 (consistent with INIT_FCALL_BY_NAME)
			ctx.Temporaries[1] = values.NewString(tt.functionName)
			ctx.Temporaries[2] = values.NewInt(int64(tt.numArgs))

			// Create INIT_FCALL instruction
			inst := &opcodes.Instruction{
				Opcode: opcodes.OP_INIT_FCALL,
				Op1:    1, // Function name
				Op2:    2, // Number of arguments
				Result: 0, // Unused
			}
			inst.OpType1, inst.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR, opcodes.IS_UNUSED)

			// Execute the instruction
			err := vm.executeInitFunctionCall(ctx, inst)

			if tt.expectError {
				require.Error(t, err, "Expected error but got none")
				if tt.errorMsg != "" {
					require.Contains(t, err.Error(), tt.errorMsg)
				}
				return
			}

			require.NoError(t, err, "executeInitFunctionCall failed")

			// Verify call context was set up correctly
			require.NotNil(t, ctx.CallContext, "CallContext should be initialized")
			require.Equal(t, tt.functionName, ctx.CallContext.FunctionName, "Function name mismatch")
			require.Equal(t, tt.numArgs, ctx.CallContext.NumArgs, "Argument count mismatch")
			require.Equal(t, 0, len(ctx.CallContext.Arguments), "Arguments should be empty initially")
			require.False(t, ctx.CallContext.IsMethod, "IsMethod should be false")
			require.Nil(t, ctx.CallContext.Object, "Object should be nil for regular functions")

			// Verify call arguments were cleared
			require.Nil(t, ctx.CallArguments, "Call arguments should be cleared")
		})
	}
}

func TestInitFunctionCallWithInvalidOperands(t *testing.T) {
	tests := []struct {
		name        string
		op1Value    *values.Value // functionName
		op2Value    *values.Value // numArgs
		expectError bool
		errorMsg    string
	}{
		{
			name:        "non-string function name",
			op1Value:    values.NewInt(123),
			op2Value:    values.NewInt(1),
			expectError: true,
			errorMsg:    "function name must be a string or callable object",
		},
		{
			name:        "null function name",
			op1Value:    values.NewNull(),
			op2Value:    values.NewInt(1),
			expectError: true,
			errorMsg:    "function name must be a string or callable object",
		},
		{
			name:        "non-integer numArgs",
			op1Value:    values.NewString("strlen"),
			op2Value:    values.NewString("not_a_number"),
			expectError: true,
			errorMsg:    "number of arguments must be an integer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := NewVirtualMachine()
			ctx := NewExecutionContext()

			// Setup operands: Op1=function name, Op2=numArgs
			ctx.Temporaries[1] = tt.op1Value // Should be function name
			ctx.Temporaries[2] = tt.op2Value // Should be numArgs

			inst := &opcodes.Instruction{
				Opcode: opcodes.OP_INIT_FCALL,
				Op1:    1,
				Op2:    2,
				Result: 0,
			}
			inst.OpType1, inst.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR, opcodes.IS_UNUSED)

			err := vm.executeInitFunctionCall(ctx, inst)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorMsg != "" {
					require.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestInitFunctionCallWithClosures(t *testing.T) {
	// Test closure calls like $closure()
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	// Create a mock callable object (closure)
	closureValue := values.NewCallable(&values.Closure{
		Function: &Function{
			Name:         "__closure__",
			Instructions: []opcodes.Instruction{},
			Constants:    []*values.Value{},
			Parameters:   []Parameter{},
		},
	})

	// Setup operands: callable object, numArgs=1
	ctx.Temporaries[1] = closureValue
	ctx.Temporaries[2] = values.NewInt(1)

	inst := &opcodes.Instruction{
		Opcode: opcodes.OP_INIT_FCALL,
		Op1:    1, // Callable object
		Op2:    2, // Number of arguments
		Result: 0,
	}
	inst.OpType1, inst.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR, opcodes.IS_UNUSED)

	err := vm.executeInitFunctionCall(ctx, inst)
	require.NoError(t, err, "executeInitFunctionCall with closure failed")

	// Verify call context setup for closures
	require.NotNil(t, ctx.CallContext, "CallContext should be initialized")
	require.Equal(t, "__closure__", ctx.CallContext.FunctionName, "Function name should be __closure__")
	require.Equal(t, 1, ctx.CallContext.NumArgs, "Argument count mismatch")
	require.Equal(t, closureValue, ctx.CallContext.Object, "Callable object should be stored in context")
}

func TestInitFunctionCallNested(t *testing.T) {
	// Test nested function calls
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	// First, set up an existing call context
	ctx.CallContext = &CallContext{
		FunctionName: "outer_function",
		Arguments:    []*values.Value{values.NewString("outer_arg")},
		NumArgs:      1,
	}

	// Setup operands for nested call
	ctx.Temporaries[1] = values.NewString("inner_function")
	ctx.Temporaries[2] = values.NewInt(2)

	inst := &opcodes.Instruction{
		Opcode: opcodes.OP_INIT_FCALL,
		Op1:    1,
		Op2:    2,
		Result: 0,
	}
	inst.OpType1, inst.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR, opcodes.IS_UNUSED)

	err := vm.executeInitFunctionCall(ctx, inst)
	require.NoError(t, err, "Nested function call initialization failed")

	// Verify the outer context was pushed to stack
	require.Len(t, ctx.CallContextStack, 1, "Call context stack should have one item")
	require.Equal(t, "outer_function", ctx.CallContextStack[0].FunctionName, "Outer context not preserved")

	// Verify new context is set up
	require.NotNil(t, ctx.CallContext, "New call context should be set")
	require.Equal(t, "inner_function", ctx.CallContext.FunctionName, "Inner function name mismatch")
	require.Equal(t, 2, ctx.CallContext.NumArgs, "Inner argument count mismatch")
}

func TestInitFunctionCallInstructionPointer(t *testing.T) {
	// Test that instruction pointer is correctly incremented
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()
	initialIP := ctx.IP

	// Setup valid operands
	ctx.Temporaries[1] = values.NewString("test_function")
	ctx.Temporaries[2] = values.NewInt(1)

	inst := &opcodes.Instruction{
		Opcode: opcodes.OP_INIT_FCALL,
		Op1:    1,
		Op2:    2,
		Result: 0,
	}
	inst.OpType1, inst.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR, opcodes.IS_UNUSED)

	err := vm.executeInitFunctionCall(ctx, inst)
	require.NoError(t, err, "executeInitFunctionCall failed")

	// Verify instruction pointer was incremented
	require.Equal(t, initialIP+1, ctx.IP, "Instruction pointer should be incremented")
}

func TestCallConstructorOpcode(t *testing.T) {
	tests := []struct {
		name               string
		object             *values.Value
		numArgs            int
		callArguments      []*values.Value
		expectError        bool
		expectedProperties []string
	}{
		{
			name:               "call constructor with no args",
			object:             createTestObject("TestClass"),
			numArgs:            0,
			callArguments:      nil,
			expectError:        false,
			expectedProperties: []string{"__constructed"},
		},
		{
			name:    "call constructor with args",
			object:  createTestObject("User"),
			numArgs: 2,
			callArguments: []*values.Value{
				values.NewString("john"),
				values.NewInt(25),
			},
			expectError:        false,
			expectedProperties: []string{"__constructed", "prop0", "prop1"},
		},
		{
			name:          "call constructor on non-object",
			object:        values.NewString("not_object"),
			numArgs:       0,
			callArguments: nil,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := NewVirtualMachine()
			ctx := NewExecutionContext()

			// Setup object
			ctx.Temporaries[1] = tt.object

			// Setup number of arguments
			if tt.numArgs > 0 {
				ctx.Temporaries[2] = values.NewInt(int64(tt.numArgs))
				ctx.CallArguments = tt.callArguments
			}

			// Create CALL_CTOR instruction
			inst := &opcodes.Instruction{
				Opcode: opcodes.OP_CALL_CTOR,
				Op1:    1, // Object
				Op2:    2, // Argument count
				Result: 0, // Unused
			}
			if tt.numArgs == 0 {
				inst.Op2 = 0
			}
			inst.OpType1, inst.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR, opcodes.IS_UNUSED)

			err := vm.executeCallConstructor(ctx, inst)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("executeCallConstructor failed: %v", err)
			}

			// Check that object properties were set
			if tt.object.IsObject() {
				obj := tt.object.Data.(*values.Object)
				for _, propName := range tt.expectedProperties {
					if _, exists := obj.Properties[propName]; !exists {
						t.Errorf("Expected property %s not found", propName)
					}
				}
			}
		})
	}
}

func TestInitConstructorCallOpcode(t *testing.T) {
	tests := []struct {
		name        string
		target      *values.Value
		numArgs     int
		expectError bool
	}{
		{
			name:        "init constructor call with class name",
			target:      values.NewString("TestClass"),
			numArgs:     2,
			expectError: false,
		},
		{
			name:        "init constructor call with object",
			target:      createTestObject("User"),
			numArgs:     1,
			expectError: false,
		},
		{
			name:        "init constructor call with invalid target",
			target:      values.NewInt(123),
			numArgs:     0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := NewVirtualMachine()
			ctx := NewExecutionContext()

			// Setup target
			ctx.Temporaries[1] = tt.target

			// Setup argument count
			if tt.numArgs > 0 {
				ctx.Temporaries[2] = values.NewInt(int64(tt.numArgs))
			}

			// Create INIT_CTOR_CALL instruction
			inst := &opcodes.Instruction{
				Opcode: opcodes.OP_INIT_CTOR_CALL,
				Op1:    1, // Target (class name or object)
				Op2:    2, // Argument count
				Result: 0, // Unused
			}
			if tt.numArgs == 0 {
				inst.Op2 = 0
			}
			inst.OpType1, inst.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR, opcodes.IS_UNUSED)

			err := vm.executeInitConstructorCall(ctx, inst)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("executeInitConstructorCall failed: %v", err)
			}

			// Check that call context was set up
			if ctx.CallContext == nil {
				t.Fatal("CallContext should be initialized")
			}

			if !ctx.CallContext.IsMethod {
				t.Error("CallContext should be marked as method call")
			}

			if ctx.CallContext.NumArgs != tt.numArgs {
				t.Errorf("Expected NumArgs %d, got %d", tt.numArgs, ctx.CallContext.NumArgs)
			}

			// Check that function name contains constructor
			expectedFuncName := ""
			if tt.target.IsString() {
				expectedFuncName = tt.target.ToString() + "::__construct"
			} else if tt.target.IsObject() {
				obj := tt.target.Data.(*values.Object)
				expectedFuncName = obj.ClassName + "::__construct"
			}

			if ctx.CallContext.FunctionName != expectedFuncName {
				t.Errorf("Expected FunctionName %s, got %s", expectedFuncName, ctx.CallContext.FunctionName)
			}

			// Check that call arguments were cleared
			if ctx.CallArguments != nil && len(ctx.CallArguments) > 0 {
				t.Error("Call arguments should be cleared")
			}
		})
	}
}

func TestOOPOpcodeErrors(t *testing.T) {
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	// Test CALL_CTOR with nil object
	inst1 := &opcodes.Instruction{
		Opcode: opcodes.OP_CALL_CTOR,
		Op1:    1, // Non-existent object
		Op2:    0,
		Result: 0,
	}
	inst1.OpType1, inst1.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_UNUSED)

	err := vm.executeCallConstructor(ctx, inst1)
	if err == nil {
		t.Error("Expected error for nil object in CALL_CTOR")
	}

	// Test INIT_CTOR_CALL with nil target
	inst2 := &opcodes.Instruction{
		Opcode: opcodes.OP_INIT_CTOR_CALL,
		Op1:    1, // Non-existent target
		Op2:    0,
		Result: 0,
	}
	inst2.OpType1, inst2.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_UNUSED)

	err = vm.executeInitConstructorCall(ctx, inst2)
	if err == nil {
		t.Error("Expected error for nil target in INIT_CTOR_CALL")
	}
}

// Helper function to create a test object
func createTestObject(className string) *values.Value {
	obj := &values.Object{
		ClassName:  className,
		Properties: make(map[string]*values.Value),
	}

	return &values.Value{
		Type: values.TypeObject,
		Data: obj,
	}
}

func TestEchoOpcode(t *testing.T) {
	tests := []struct {
		name     string
		input    *values.Value
		expected string
	}{
		{"echo string", values.NewString("Hello World"), "Hello World"},
		{"echo int", values.NewInt(42), "42"},
		{"echo float", values.NewFloat(3.14), "3.14"},
		{"echo bool true", values.NewBool(true), "1"},
		{"echo bool false", values.NewBool(false), ""},
		{"echo null", values.NewNull(), ""},
		{"echo empty string", values.NewString(""), ""},
		{"echo zero", values.NewInt(0), "0"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vm := NewVirtualMachine()
			ctx := NewExecutionContext()

			// Set up output capture
			var buf bytes.Buffer
			ctx.SetOutputWriter(&buf)

			ctx.Temporaries = make(map[uint32]*values.Value)
			ctx.Temporaries[0] = test.input

			op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_UNUSED)
			inst := opcodes.Instruction{
				Opcode:  opcodes.OP_ECHO,
				OpType1: op1Type,
				OpType2: op2Type,
				Op1:     0,
			}

			err := vm.executeEcho(ctx, &inst)
			if err != nil {
				t.Fatalf("ECHO execution failed: %v", err)
			}

			// Check output
			actualOutput := buf.String()
			if actualOutput != test.expected {
				t.Errorf("Expected output '%s', got '%s'", test.expected, actualOutput)
			}
		})
	}
}

func TestMultipleEchoOpcodes(t *testing.T) {
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	// Set up output capture
	var buf bytes.Buffer
	ctx.SetOutputWriter(&buf)

	ctx.Temporaries = make(map[uint32]*values.Value)
	ctx.Temporaries[0] = values.NewString("Hello")
	ctx.Temporaries[1] = values.NewString(" ")
	ctx.Temporaries[2] = values.NewString("World")
	ctx.Temporaries[3] = values.NewString("!")

	// Echo "Hello"
	op1Type1, op2Type1 := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_UNUSED)
	inst1 := opcodes.Instruction{
		Opcode:  opcodes.OP_ECHO,
		OpType1: op1Type1,
		OpType2: op2Type1,
		Op1:     0,
	}

	err := vm.executeEcho(ctx, &inst1)
	if err != nil {
		t.Fatalf("First ECHO failed: %v", err)
	}

	// Echo " "
	inst2 := opcodes.Instruction{
		Opcode:  opcodes.OP_ECHO,
		OpType1: op1Type1,
		OpType2: op2Type1,
		Op1:     1,
	}

	err = vm.executeEcho(ctx, &inst2)
	if err != nil {
		t.Fatalf("Second ECHO failed: %v", err)
	}

	// Echo "World"
	inst3 := opcodes.Instruction{
		Opcode:  opcodes.OP_ECHO,
		OpType1: op1Type1,
		OpType2: op2Type1,
		Op1:     2,
	}

	err = vm.executeEcho(ctx, &inst3)
	if err != nil {
		t.Fatalf("Third ECHO failed: %v", err)
	}

	// Echo "!"
	inst4 := opcodes.Instruction{
		Opcode:  opcodes.OP_ECHO,
		OpType1: op1Type1,
		OpType2: op2Type1,
		Op1:     3,
	}

	err = vm.executeEcho(ctx, &inst4)
	if err != nil {
		t.Fatalf("Fourth ECHO failed: %v", err)
	}

	// Check output
	actualOutput := buf.String()
	expectedOutput := "Hello World!"
	if actualOutput != expectedOutput {
		t.Errorf("Expected output '%s', got '%s'", expectedOutput, actualOutput)
	}
}

func TestPrintOpcode(t *testing.T) {
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	ctx.Temporaries = make(map[uint32]*values.Value)
	ctx.Temporaries[0] = values.NewString("Hello Print")

	op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)
	inst := opcodes.Instruction{
		Opcode:  opcodes.OP_PRINT,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     0,
		Result:  1,
	}

	err := vm.executePrint(ctx, &inst)
	if err != nil {
		t.Fatalf("PRINT execution failed: %v", err)
	}

	// Check that print always returns 1
	result := ctx.Temporaries[1]
	if result == nil {
		t.Fatal("PRINT result is nil")
	}

	if !result.IsInt() || result.Data.(int64) != 1 {
		t.Errorf("Expected print to return 1, got %v", result)
	}
}

func TestExitOpcode(t *testing.T) {
	tests := []struct {
		name         string
		input        *values.Value
		expectedCode int
		expectedHalt bool
	}{
		{"exit without code", nil, 0, true},
		{"exit with int code", values.NewInt(42), 42, true},
		{"exit with zero", values.NewInt(0), 0, true},
		{"exit with negative", values.NewInt(-1), -1, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vm := NewVirtualMachine()
			ctx := NewExecutionContext()

			ctx.Temporaries = make(map[uint32]*values.Value)
			if test.input != nil {
				ctx.Temporaries[0] = test.input
			}

			var op1Type, op2Type byte
			var op1 uint32

			if test.input != nil {
				op1Type, op2Type = opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_UNUSED)
				op1 = 0
			} else {
				op1Type, op2Type = opcodes.EncodeOpTypes(opcodes.IS_UNUSED, opcodes.IS_UNUSED, opcodes.IS_UNUSED)
				op1 = 0
			}

			inst := opcodes.Instruction{
				Opcode:  opcodes.OP_EXIT,
				OpType1: op1Type,
				OpType2: op2Type,
				Op1:     op1,
			}

			err := vm.executeExit(ctx, &inst)
			if err != nil {
				t.Fatalf("EXIT execution failed: %v", err)
			}

			if ctx.Halted != test.expectedHalt {
				t.Errorf("Expected halted=%v, got %v", test.expectedHalt, ctx.Halted)
			}

			if ctx.ExitCode != test.expectedCode {
				t.Errorf("Expected exit code %d, got %d", test.expectedCode, ctx.ExitCode)
			}
		})
	}
}

func TestExitWithStringMessage(t *testing.T) {
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	ctx.Temporaries = make(map[uint32]*values.Value)
	ctx.Temporaries[0] = values.NewString("Error: Something went wrong")

	op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_UNUSED)
	inst := opcodes.Instruction{
		Opcode:  opcodes.OP_EXIT,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     0,
	}

	err := vm.executeExit(ctx, &inst)
	if err != nil {
		t.Fatalf("EXIT with string failed: %v", err)
	}

	// Should be halted with exit code 0
	if !ctx.Halted {
		t.Error("Expected execution to be halted")
	}

	if ctx.ExitCode != 0 {
		t.Errorf("Expected exit code 0 for string message, got %d", ctx.ExitCode)
	}

	// The string should not be in output for exit (unlike die/exit with message)
	// This is a simplified implementation - no output should be generated for exit with string
}

func TestReturnOpcode(t *testing.T) {
	tests := []struct {
		name     string
		input    *values.Value
		hasInput bool
	}{
		{"return without value", nil, false},
		{"return with int", values.NewInt(42), true},
		{"return with string", values.NewString("hello"), true},
		{"return with null", values.NewNull(), true},
		{"return with bool", values.NewBool(true), true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vm := NewVirtualMachine()
			ctx := NewExecutionContext()

			ctx.Temporaries = make(map[uint32]*values.Value)
			if test.hasInput {
				ctx.Temporaries[0] = test.input
			}

			var op1Type, op2Type byte
			var op1 uint32

			if test.hasInput {
				op1Type, op2Type = opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)
				op1 = 0
			} else {
				op1Type, op2Type = opcodes.EncodeOpTypes(opcodes.IS_UNUSED, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)
				op1 = 0
			}

			inst := opcodes.Instruction{
				Opcode:  opcodes.OP_RETURN,
				OpType1: op1Type,
				OpType2: op2Type,
				Op1:     op1,
				Result:  1,
			}

			err := vm.executeReturn(ctx, &inst)
			if err != nil {
				t.Fatalf("RETURN execution failed: %v", err)
			}

			// Check return value
			result := ctx.Temporaries[1]
			if result == nil {
				t.Fatal("RETURN result is nil")
			}

			if test.hasInput {
				// Should return the input value
				expectedType := test.input.Type
				if result.Type != expectedType {
					t.Errorf("Expected return type %v, got %v", expectedType, result.Type)
				}

				switch expectedType {
				case values.TypeInt:
					if result.Data.(int64) != test.input.Data.(int64) {
						t.Errorf("Expected return value %v, got %v", test.input.Data, result.Data)
					}
				case values.TypeString:
					if result.Data.(string) != test.input.Data.(string) {
						t.Errorf("Expected return value '%s', got '%s'", test.input.Data.(string), result.Data.(string))
					}
				case values.TypeBool:
					if result.Data.(bool) != test.input.Data.(bool) {
						t.Errorf("Expected return value %v, got %v", test.input.Data, result.Data)
					}
				}
			} else {
				// Should return null when no return value specified
				if !result.IsNull() {
					t.Errorf("Expected null return value, got %v", result)
				}
			}
		})
	}
}

func TestIntegratedOutputFlow(t *testing.T) {
	// Test a sequence of output operations
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	// Set up output capture
	var buf bytes.Buffer
	ctx.SetOutputWriter(&buf)

	ctx.Temporaries = make(map[uint32]*values.Value)
	ctx.Temporaries[0] = values.NewString("Start: ")
	ctx.Temporaries[1] = values.NewInt(42)
	ctx.Temporaries[2] = values.NewString(" End")

	// Echo "Start: "
	op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_UNUSED)
	inst1 := opcodes.Instruction{
		Opcode:  opcodes.OP_ECHO,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     0,
	}

	err := vm.executeEcho(ctx, &inst1)
	if err != nil {
		t.Fatalf("First echo failed: %v", err)
	}

	// Print 42 (should also return 1)
	op1TypePrint, op2TypePrint := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)
	inst2 := opcodes.Instruction{
		Opcode:  opcodes.OP_PRINT,
		OpType1: op1TypePrint,
		OpType2: op2TypePrint,
		Op1:     1,
		Result:  3, // Store print result
	}

	err = vm.executePrint(ctx, &inst2)
	if err != nil {
		t.Fatalf("Print failed: %v", err)
	}

	// Echo " End"
	inst3 := opcodes.Instruction{
		Opcode:  opcodes.OP_ECHO,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     2,
	}

	err = vm.executeEcho(ctx, &inst3)
	if err != nil {
		t.Fatalf("Second echo failed: %v", err)
	}

	// Verify output
	actualOutput := buf.String()
	expectedOutput := "Start: 42 End"
	if actualOutput != expectedOutput {
		t.Errorf("Expected output '%s', got '%s'", expectedOutput, actualOutput)
	}

	// Verify print result
	printResult := ctx.Temporaries[3]
	if printResult == nil || !printResult.IsInt() || printResult.Data.(int64) != 1 {
		t.Errorf("Expected print result to be 1, got %v", printResult)
	}
}

func TestRecvOpcode(t *testing.T) {
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	// Set up parameters
	ctx.Parameters = []*values.Value{
		values.NewString("param1"),
		values.NewInt(42),
		values.NewBool(true),
	}
	ctx.Temporaries = make(map[uint32]*values.Value)

	// Create RECV instruction to receive parameter 1 (42)
	op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_UNUSED, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)
	inst := opcodes.Instruction{
		Opcode:  opcodes.OP_RECV,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     1, // Parameter index
		Op2:     0,
		Result:  0, // Store in temporary variable 0
	}

	// Execute RECV
	err := vm.executeRecv(ctx, &inst)
	if err != nil {
		t.Fatalf("RECV execution failed: %v", err)
	}

	// Check result
	result := ctx.Temporaries[0]
	if result == nil {
		t.Fatal("RECV result is nil")
	}

	if result.Type != values.TypeInt || result.Data.(int64) != 42 {
		t.Errorf("Expected parameter value 42, got %v", result.Data)
	}
}

func TestRecvNonExistentParameter(t *testing.T) {
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	// Set up only 2 parameters
	ctx.Parameters = []*values.Value{
		values.NewString("param1"),
		values.NewInt(42),
	}
	ctx.Temporaries = make(map[uint32]*values.Value)

	// Create RECV instruction to receive parameter 5 (doesn't exist)
	op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_UNUSED, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)
	inst := opcodes.Instruction{
		Opcode:  opcodes.OP_RECV,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     5, // Parameter index (out of bounds)
		Result:  0, // Store in temporary variable 0
	}

	// Execute RECV
	err := vm.executeRecv(ctx, &inst)
	if err != nil {
		t.Fatalf("RECV execution failed: %v", err)
	}

	// Check result - should be null
	result := ctx.Temporaries[0]
	if result == nil {
		t.Fatal("RECV result is nil")
	}

	if !result.IsNull() {
		t.Error("Expected null for non-existent parameter")
	}
}

func TestRecvInitOpcode(t *testing.T) {
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	// Set up parameters (missing parameter 2)
	ctx.Parameters = []*values.Value{
		values.NewString("param1"),
		values.NewInt(42),
	}
	ctx.Temporaries = make(map[uint32]*values.Value)
	// Set up default value in temporary variable 1
	ctx.Temporaries[1] = values.NewString("default_value")

	// Create RECV_INIT instruction to receive parameter 2 with default
	op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_UNUSED, opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR)
	inst := opcodes.Instruction{
		Opcode:  opcodes.OP_RECV_INIT,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     2, // Parameter index (doesn't exist)
		Op2:     1, // Default value from temporary variable 1
		Result:  0, // Store in temporary variable 0
	}

	// Execute RECV_INIT
	err := vm.executeRecvInit(ctx, &inst)
	if err != nil {
		t.Fatalf("RECV_INIT execution failed: %v", err)
	}

	// Check result - should be default value
	result := ctx.Temporaries[0]
	if result == nil {
		t.Fatal("RECV_INIT result is nil")
	}

	if result.Type != values.TypeString || result.Data.(string) != "default_value" {
		t.Errorf("Expected default value 'default_value', got %v", result.Data)
	}
}

func TestRecvInitWithProvidedParameter(t *testing.T) {
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	// Set up parameters (parameter 1 exists)
	ctx.Parameters = []*values.Value{
		values.NewString("param1"),
		values.NewInt(42),
		values.NewString("provided_param"),
	}
	ctx.Temporaries = make(map[uint32]*values.Value)
	// Set up default value in temporary variable 1
	ctx.Temporaries[1] = values.NewString("default_value")

	// Create RECV_INIT instruction to receive parameter 2 with default
	op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_UNUSED, opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR)
	inst := opcodes.Instruction{
		Opcode:  opcodes.OP_RECV_INIT,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     2, // Parameter index (exists)
		Op2:     1, // Default value from temporary variable 1
		Result:  0, // Store in temporary variable 0
	}

	// Execute RECV_INIT
	err := vm.executeRecvInit(ctx, &inst)
	if err != nil {
		t.Fatalf("RECV_INIT execution failed: %v", err)
	}

	// Check result - should be provided parameter, not default
	result := ctx.Temporaries[0]
	if result == nil {
		t.Fatal("RECV_INIT result is nil")
	}

	if result.Type != values.TypeString || result.Data.(string) != "provided_param" {
		t.Errorf("Expected provided parameter 'provided_param', got %v", result.Data)
	}
}

func TestRecvVariadicOpcode(t *testing.T) {
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	// Set up parameters (5 total, variadic starts from index 2)
	ctx.Parameters = []*values.Value{
		values.NewString("param1"),
		values.NewInt(42),
		values.NewString("variadic1"),
		values.NewInt(123),
		values.NewBool(true),
	}
	ctx.Temporaries = make(map[uint32]*values.Value)

	// Create RECV_VARIADIC instruction starting from parameter 2
	op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_UNUSED, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)
	inst := opcodes.Instruction{
		Opcode:  opcodes.OP_RECV_VARIADIC,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     2, // Start collecting from parameter 2
		Result:  0, // Store array in temporary variable 0
	}

	// Execute RECV_VARIADIC
	err := vm.executeRecvVariadic(ctx, &inst)
	if err != nil {
		t.Fatalf("RECV_VARIADIC execution failed: %v", err)
	}

	// Check result - should be array with 3 elements
	result := ctx.Temporaries[0]
	if result == nil {
		t.Fatal("RECV_VARIADIC result is nil")
	}

	if result.Type != values.TypeArray {
		t.Errorf("Expected array type, got %v", result.Type)
	}

	// Check array contents
	if result.ArrayCount() != 3 {
		t.Errorf("Expected 3 variadic parameters, got %d", result.ArrayCount())
	}

	// Check first variadic element (index 0 in array, parameter 2 in call)
	elem0 := result.ArrayGet(values.NewInt(0))
	if elem0.Type != values.TypeString || elem0.Data.(string) != "variadic1" {
		t.Errorf("Expected first variadic element 'variadic1', got %v", elem0.Data)
	}

	// Check second variadic element
	elem1 := result.ArrayGet(values.NewInt(1))
	if elem1.Type != values.TypeInt || elem1.Data.(int64) != 123 {
		t.Errorf("Expected second variadic element 123, got %v", elem1.Data)
	}
}

func TestSendVarExOpcode(t *testing.T) {
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	ctx.Temporaries = make(map[uint32]*values.Value)
	ctx.Temporaries[0] = values.NewString("send_value")

	// Create SEND_VAR_EX instruction
	op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)
	inst := opcodes.Instruction{
		Opcode:  opcodes.OP_SEND_VAR_EX,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     0, // Send from temporary variable 0
		Result:  1, // Store result in temporary variable 1
	}

	// Execute SEND_VAR_EX
	err := vm.executeSendVarEx(ctx, &inst)
	if err != nil {
		t.Fatalf("SEND_VAR_EX execution failed: %v", err)
	}

	// Check result - should be same as sent value
	result := ctx.Temporaries[1]
	if result == nil {
		t.Fatal("SEND_VAR_EX result is nil")
	}

	if result.Type != values.TypeString || result.Data.(string) != "send_value" {
		t.Errorf("Expected sent value 'send_value', got %v", result.Data)
	}

	// Check call arguments
	if len(ctx.CallArguments) != 1 {
		t.Errorf("Expected 1 call argument, got %d", len(ctx.CallArguments))
	}

	if ctx.CallArguments[0].Data.(string) != "send_value" {
		t.Errorf("Expected call argument 'send_value', got %v", ctx.CallArguments[0].Data)
	}
}

func TestSendVarNoRefOpcode(t *testing.T) {
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	ctx.Temporaries = make(map[uint32]*values.Value)
	ctx.Temporaries[0] = values.NewString("send_value")

	// Create SEND_VAR_NO_REF instruction
	op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)
	inst := opcodes.Instruction{
		Opcode:  opcodes.OP_SEND_VAR_NO_REF,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     0, // Send from temporary variable 0
		Result:  1, // Store result in temporary variable 1
	}

	// Execute SEND_VAR_NO_REF
	err := vm.executeSendVarNoRef(ctx, &inst)
	if err != nil {
		t.Fatalf("SEND_VAR_NO_REF execution failed: %v", err)
	}

	// Check result - should be copy of sent value
	result := ctx.Temporaries[1]
	if result == nil {
		t.Fatal("SEND_VAR_NO_REF result is nil")
	}

	if result.Type != values.TypeString || result.Data.(string) != "send_value" {
		t.Errorf("Expected copied value 'send_value', got %v", result.Data)
	}

	// Check call arguments
	if len(ctx.CallArguments) != 1 {
		t.Errorf("Expected 1 call argument, got %d", len(ctx.CallArguments))
	}

	// Verify the argument is a copy (different pointer but same data)
	if ctx.CallArguments[0] == ctx.Temporaries[0] {
		t.Error("Expected copied value, but got same pointer (reference)")
	}

	if ctx.CallArguments[0].Data.(string) != "send_value" {
		t.Errorf("Expected call argument 'send_value', got %v", ctx.CallArguments[0].Data)
	}
}

// Test a complete parameter passing simulation
func TestParameterPassingSimulation(t *testing.T) {
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	// Simulate function call: func(param1, param2="default", ...rest)
	// Called with: func("hello", 42, "extra1", "extra2")

	// Set up parameters for the function
	ctx.Parameters = []*values.Value{
		values.NewString("hello"),  // param1
		values.NewInt(42),          // param2 (overrides default)
		values.NewString("extra1"), // variadic[0]
		values.NewString("extra2"), // variadic[1]
	}
	ctx.Temporaries = make(map[uint32]*values.Value)
	ctx.Temporaries[10] = values.NewString("default") // Default value for param2

	// 1. Receive first parameter
	op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_UNUSED, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)
	recvInst1 := opcodes.Instruction{
		Opcode:  opcodes.OP_RECV,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     0, // Parameter 0
		Result:  0, // Store in temp var 0
	}

	err := vm.executeRecv(ctx, &recvInst1)
	if err != nil {
		t.Fatalf("First RECV failed: %v", err)
	}

	param1 := ctx.Temporaries[0]
	if param1.Data.(string) != "hello" {
		t.Errorf("Expected param1 'hello', got %v", param1.Data)
	}

	// 2. Receive second parameter with default
	op1Type2, op2Type2 := opcodes.EncodeOpTypes(opcodes.IS_UNUSED, opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR)
	recvInitInst := opcodes.Instruction{
		Opcode:  opcodes.OP_RECV_INIT,
		OpType1: op1Type2,
		OpType2: op2Type2,
		Op1:     1,  // Parameter 1
		Op2:     10, // Default from temp var 10
		Result:  1,  // Store in temp var 1
	}

	err = vm.executeRecvInit(ctx, &recvInitInst)
	if err != nil {
		t.Fatalf("RECV_INIT failed: %v", err)
	}

	param2 := ctx.Temporaries[1]
	if param2.Data.(int64) != 42 {
		t.Errorf("Expected param2 42, got %v", param2.Data)
	}

	// 3. Receive variadic parameters
	variadicInst := opcodes.Instruction{
		Opcode:  opcodes.OP_RECV_VARIADIC,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     2, // Start from parameter 2
		Result:  2, // Store in temp var 2
	}

	err = vm.executeRecvVariadic(ctx, &variadicInst)
	if err != nil {
		t.Fatalf("RECV_VARIADIC failed: %v", err)
	}

	variadicArray := ctx.Temporaries[2]
	if variadicArray.ArrayCount() != 2 {
		t.Errorf("Expected 2 variadic parameters, got %d", variadicArray.ArrayCount())
	}

	extra1 := variadicArray.ArrayGet(values.NewInt(0))
	if extra1.Data.(string) != "extra1" {
		t.Errorf("Expected variadic[0] 'extra1', got %v", extra1.Data)
	}

	extra2 := variadicArray.ArrayGet(values.NewInt(1))
	if extra2.Data.(string) != "extra2" {
		t.Errorf("Expected variadic[1] 'extra2', got %v", extra2.Data)
	}
}

func TestFetchGlobalsOpcode(t *testing.T) {
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	// Setup some global variables
	ctx.GlobalVars["testVar"] = values.NewString("test_value")
	ctx.GlobalVars["numVar"] = values.NewInt(42)
	ctx.GlobalVars["boolVar"] = values.NewBool(true)

	// Create FETCH_GLOBALS instruction
	inst := &opcodes.Instruction{
		Opcode: opcodes.OP_FETCH_GLOBALS,
		Op1:    0, // Unused
		Op2:    0, // Unused
		Result: 1, // Result location
	}
	inst.OpType1, inst.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_UNUSED, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)

	err := vm.executeFetchGlobals(ctx, inst)
	if err != nil {
		t.Fatalf("executeFetchGlobals failed: %v", err)
	}

	// Check result
	result := ctx.Temporaries[1]
	if result == nil {
		t.Fatal("Result should not be nil")
	}

	if !result.IsArray() {
		t.Errorf("Expected array result, got %T", result.Data)
	}

	globalsData := result.Data.(*values.Array)

	// Check that global variables are included
	if val, exists := globalsData.Elements["testVar"]; !exists {
		t.Error("testVar should be in $GLOBALS")
	} else if val.ToString() != "test_value" {
		t.Errorf("Expected testVar='test_value', got %v", val)
	}

	if val, exists := globalsData.Elements["numVar"]; !exists {
		t.Error("numVar should be in $GLOBALS")
	} else if val.ToInt() != 42 {
		t.Errorf("Expected numVar=42, got %v", val)
	}

	// Check that superglobals are initialized
	superglobals := []string{"_SERVER", "_GET", "_POST", "_SESSION", "_COOKIE", "_FILES", "_REQUEST", "_ENV"}
	for _, name := range superglobals {
		if val, exists := globalsData.Elements[name]; !exists {
			t.Errorf("Superglobal %s should be initialized", name)
		} else if !val.IsArray() {
			t.Errorf("Superglobal %s should be an array", name)
		}
	}

	// Check that $GLOBALS contains itself
	if _, exists := globalsData.Elements["GLOBALS"]; !exists {
		t.Error("$GLOBALS should contain itself")
	}
}

func TestGeneratorReturnOpcode(t *testing.T) {
	tests := []struct {
		name           string
		returnValue    *values.Value
		hasGenerator   bool
		expectedHalted bool
		expectedResult *values.Value
	}{
		{
			name:           "generator return with value",
			returnValue:    values.NewString("final_value"),
			hasGenerator:   true,
			expectedHalted: true,
			expectedResult: values.NewString("final_value"),
		},
		{
			name:           "generator return without value",
			returnValue:    nil,
			hasGenerator:   true,
			expectedHalted: true,
			expectedResult: values.NewNull(),
		},
		{
			name:           "regular return in function",
			returnValue:    values.NewInt(123),
			hasGenerator:   false,
			expectedHalted: false,
			expectedResult: values.NewInt(123),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize registry for each test
			registry.Initialize()

			vm := NewVirtualMachine()
			ctx := NewExecutionContext()

			// Setup return value
			if tt.returnValue != nil {
				ctx.Temporaries[1] = tt.returnValue
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
			} else {
				// Setup call stack for regular function return
				ctx.CallStack = append(ctx.CallStack, CallFrame{
					Function:    nil,
					ReturnIP:    100,
					Variables:   make(map[uint32]*values.Value),
					ThisObject:  nil,
					Arguments:   nil,
					ReturnValue: nil,
					ReturnByRef: false,
				})
			}

			// Create GENERATOR_RETURN instruction
			inst := &opcodes.Instruction{
				Opcode: opcodes.OP_GENERATOR_RETURN,
				Op1:    1, // Return value
				Op2:    0, // Unused
				Result: 0, // Unused
			}
			inst.OpType1, inst.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_UNUSED)

			err := vm.executeGeneratorReturn(ctx, inst)
			if err != nil {
				t.Fatalf("executeGeneratorReturn failed: %v", err)
			}

			// Check execution state
			if ctx.Halted != tt.expectedHalted {
				t.Errorf("Expected halted=%t, got halted=%t", tt.expectedHalted, ctx.Halted)
			}

			// Check generator state if applicable
			if tt.hasGenerator {
				if ctx.CurrentGenerator.YieldedValue == nil {
					t.Error("Generator should have return value")
				} else if !valuesEqualForTest(ctx.CurrentGenerator.YieldedValue, tt.expectedResult) {
					t.Errorf("Expected generator return %v, got %v", tt.expectedResult, ctx.CurrentGenerator.YieldedValue)
				}

				if !ctx.CurrentGenerator.IsFinished {
					t.Error("Generator should be marked as finished")
				}

				if ctx.CurrentGenerator.IsSuspended {
					t.Error("Generator should not be suspended after return")
				}
			}
		})
	}
}

func TestVerifyAbstractClassOpcode(t *testing.T) {
	tests := []struct {
		name         string
		className    string
		setupClass   bool
		isAbstract   bool
		expectError  bool
		errorMessage string
	}{
		{
			name:         "concrete class - should pass",
			className:    "ConcreteClass",
			setupClass:   true,
			isAbstract:   false,
			expectError:  false,
			errorMessage: "",
		},
		{
			name:         "abstract class by name - should fail",
			className:    "AbstractBaseClass",
			setupClass:   true,
			isAbstract:   true,
			expectError:  true,
			errorMessage: "Cannot instantiate abstract class AbstractBaseClass",
		},
		{
			name:         "class with abstract suffix - should fail",
			className:    "HandlerAbstract",
			setupClass:   true,
			isAbstract:   true,
			expectError:  true,
			errorMessage: "Cannot instantiate abstract class HandlerAbstract",
		},
		{
			name:         "class with abstract method - should fail",
			className:    "MyClass",
			setupClass:   true,
			isAbstract:   true,
			expectError:  true,
			errorMessage: "Cannot instantiate abstract class MyClass",
		},
		{
			name:         "non-existent class - should pass",
			className:    "NonExistentClass",
			setupClass:   false,
			isAbstract:   false,
			expectError:  false,
			errorMessage: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize registry for each test
			registry.Initialize()

			vm := NewVirtualMachine()
			ctx := NewExecutionContext()

			// Setup class if needed
			if tt.setupClass {
				classDesc := &registry.ClassDescriptor{
					Name:       tt.className,
					Properties: make(map[string]*registry.PropertyDescriptor),
					Methods:    make(map[string]*registry.MethodDescriptor),
					Constants:  make(map[string]*registry.ConstantDescriptor),
					IsAbstract: tt.isAbstract,
				}

				if tt.isAbstract {
					// Add abstract method to make class abstract
					if tt.className == "MyClass" {
						// Add an abstract method
						methodDesc := &registry.MethodDescriptor{
							Name:           "abstractMethod",
							Parameters:     []registry.ParameterDescriptor{},
							Visibility:     "public",
							IsStatic:       false,
							IsAbstract:     true,
							IsFinal:        false,
							Implementation: nil, // Abstract method has no implementation
						}
						classDesc.Methods["abstractMethod"] = methodDesc
					}
				}

				// Register class in unified registry
				registry.GlobalRegistry.RegisterClass(classDesc)
			}

			// Setup instruction operands
			ctx.Temporaries[1] = values.NewString(tt.className)

			// Create VERIFY_ABSTRACT_CLASS instruction
			inst := &opcodes.Instruction{
				Opcode: opcodes.OP_VERIFY_ABSTRACT_CLASS,
				Op1:    1, // Class name
				Op2:    0, // Unused
				Result: 0, // Unused
			}
			inst.OpType1, inst.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_UNUSED)

			err := vm.executeVerifyAbstractClass(ctx, inst)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if err.Error() != tt.errorMessage {
					t.Errorf("Expected error %q, got %q", tt.errorMessage, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("executeVerifyAbstractClass failed: %v", err)
			}
		})
	}
}

func TestRemainingOpcodeErrors(t *testing.T) {
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	// Test VERIFY_ABSTRACT_CLASS with non-string class name
	ctx.Temporaries[1] = values.NewInt(123) // Invalid class name

	inst := &opcodes.Instruction{
		Opcode: opcodes.OP_VERIFY_ABSTRACT_CLASS,
		Op1:    1,
		Op2:    0,
		Result: 0,
	}
	inst.OpType1, inst.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_UNUSED)

	err := vm.executeVerifyAbstractClass(ctx, inst)
	if err == nil {
		t.Error("Expected error for non-string class name in VERIFY_ABSTRACT_CLASS")
	}

	expectedError := "VERIFY_ABSTRACT_CLASS requires string class name"
	if err.Error() != expectedError {
		t.Errorf("Expected error %q, got %q", expectedError, err.Error())
	}
}

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

func TestFetchStaticPropertyIssetOpcode(t *testing.T) {
	tests := []struct {
		name        string
		className   string
		propName    string
		setupClass  bool
		setupProp   bool
		propValue   *values.Value
		expectedSet bool
		expectError bool
	}{
		{
			name:        "property exists and is set",
			className:   "TestClass",
			propName:    "staticProp",
			setupClass:  true,
			setupProp:   true,
			propValue:   values.NewString("test_value"),
			expectedSet: true,
			expectError: false,
		},
		{
			name:        "property exists but is null",
			className:   "TestClass",
			propName:    "nullProp",
			setupClass:  true,
			setupProp:   true,
			propValue:   values.NewNull(),
			expectedSet: false,
			expectError: false,
		},
		{
			name:        "property doesn't exist",
			className:   "TestClass",
			propName:    "nonExistentProp",
			setupClass:  true,
			setupProp:   false,
			propValue:   nil,
			expectedSet: false,
			expectError: false,
		},
		{
			name:        "class doesn't exist",
			className:   "NonExistentClass",
			propName:    "anyProp",
			setupClass:  false,
			setupProp:   false,
			propValue:   nil,
			expectedSet: false,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize registry for each test
			registry.Initialize()

			vm := NewVirtualMachine()
			ctx := NewExecutionContext()

			// Setup class if needed using registry
			if tt.setupClass {
				vm.setStaticPropertyInRegistry(tt.className, tt.propName+"_dummy", values.NewNull())

				// Setup property if needed
				if tt.setupProp {
					vm.setStaticPropertyInRegistry(tt.className, tt.propName, tt.propValue)
				}
			}

			// Setup instruction operands
			ctx.Temporaries[1] = values.NewString(tt.className)
			ctx.Temporaries[2] = values.NewString(tt.propName)

			// Create FETCH_STATIC_PROP_IS instruction
			inst := &opcodes.Instruction{
				Opcode: opcodes.OP_FETCH_STATIC_PROP_IS,
				Op1:    1, // Class name
				Op2:    2, // Property name
				Result: 3, // Result location
			}
			inst.OpType1, inst.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR)

			err := vm.executeFetchStaticPropertyIsset(ctx, inst)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("executeFetchStaticPropertyIsset failed: %v", err)
			}

			// Check result
			result := ctx.Temporaries[3]
			if result == nil {
				t.Fatal("Result should not be nil")
			}

			if !result.IsBool() {
				t.Errorf("Expected boolean result, got %T", result.Data)
			}

			if result.ToBool() != tt.expectedSet {
				t.Errorf("Expected isset=%t, got isset=%t", tt.expectedSet, result.ToBool())
			}
		})
	}
}

func TestFetchStaticPropertyReadWriteOpcode(t *testing.T) {
	tests := []struct {
		name         string
		className    string
		propName     string
		setupClass   bool
		setupProp    bool
		initialValue *values.Value
		expectError  bool
	}{
		{
			name:         "existing property read-write",
			className:    "TestClass",
			propName:     "existingProp",
			setupClass:   true,
			setupProp:    true,
			initialValue: values.NewString("initial_value"),
			expectError:  false,
		},
		{
			name:         "non-existing property - should create",
			className:    "TestClass",
			propName:     "newProp",
			setupClass:   true,
			setupProp:    false,
			initialValue: nil,
			expectError:  false,
		},
		{
			name:         "non-existing class - should create",
			className:    "NewClass",
			propName:     "newProp",
			setupClass:   false,
			setupProp:    false,
			initialValue: nil,
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize registry for each test
			registry.Initialize()

			vm := NewVirtualMachine()
			ctx := NewExecutionContext()

			// Setup class if needed
			if tt.setupClass {
				vm.setStaticPropertyInRegistry(tt.className, tt.propName+"_dummy", values.NewNull())

				// Setup property if needed
				if tt.setupProp {
					vm.setStaticPropertyInRegistry(tt.className, tt.propName, tt.initialValue)
				}
			}

			// Setup instruction operands
			ctx.Temporaries[1] = values.NewString(tt.className)
			ctx.Temporaries[2] = values.NewString(tt.propName)

			// Create FETCH_STATIC_PROP_RW instruction
			inst := &opcodes.Instruction{
				Opcode: opcodes.OP_FETCH_STATIC_PROP_RW,
				Op1:    1, // Class name
				Op2:    2, // Property name
				Result: 3, // Result location
			}
			inst.OpType1, inst.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR)

			err := vm.executeFetchStaticPropertyReadWrite(ctx, inst)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("executeFetchStaticPropertyReadWrite failed: %v", err)
			}

			// Check result
			result := ctx.Temporaries[3]
			if result == nil {
				t.Fatal("Result should not be nil")
			}

			// Check that class and property were created if needed
			// Verify class exists in registry
			if _, err := registry.GlobalRegistry.GetClass(tt.className); err != nil {
				t.Errorf("Class %s should exist", tt.className)
			}

			if _, exists := vm.getStaticPropertyFromRegistry(tt.className, tt.propName); !exists {
				t.Errorf("Property %s should exist", tt.propName)
			}

			// Check that the property is marked as static
			if classDesc, err := registry.GlobalRegistry.GetClass(tt.className); err == nil {
				if prop, exists := classDesc.Properties[tt.propName]; exists {
					if !prop.IsStatic {
						t.Error("Property should be marked as static")
					}
				}
			}

			// Check that result matches expected value
			if tt.initialValue != nil {
				if !valuesEqualForTest(result, tt.initialValue) {
					t.Errorf("Expected result %v, got %v", tt.initialValue, result)
				}
			} else {
				// Should be null for newly created properties
				if !result.IsNull() {
					t.Errorf("Expected null for new property, got %v", result)
				}
			}
		})
	}
}

func TestFetchStaticPropertyUnsetOpcode(t *testing.T) {
	// Initialize registry for test
	registry.Initialize()

	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	// Setup class with property
	className := "TestClass"
	propName := "testProp"
	// Setup class with property using compatibility layer
	vm.setStaticPropertyInRegistry(className, propName+"_dummy", values.NewNull())
	vm.setStaticPropertyInRegistry(className, propName, values.NewString("test_value"))

	// Setup instruction operands
	ctx.Temporaries[1] = values.NewString(className)
	ctx.Temporaries[2] = values.NewString(propName)

	// Create FETCH_STATIC_PROP_UNSET instruction
	inst := &opcodes.Instruction{
		Opcode: opcodes.OP_FETCH_STATIC_PROP_UNSET,
		Op1:    1, // Class name
		Op2:    2, // Property name
		Result: 0, // Unused
	}
	inst.OpType1, inst.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR, opcodes.IS_UNUSED)

	err := vm.executeFetchStaticPropertyUnset(ctx, inst)
	if err != nil {
		t.Fatalf("executeFetchStaticPropertyUnset failed: %v", err)
	}

	// Check that property was removed
	if _, exists := vm.getStaticPropertyFromRegistry(className, propName); exists {
		t.Error("Property should have been removed")
	}
}

func TestStaticPropertyOpcodeErrors(t *testing.T) {
	// Initialize registry for test
	registry.Initialize()

	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	// Test FETCH_STATIC_PROP_IS with non-string class name
	ctx.Temporaries[1] = values.NewInt(123) // Invalid class name
	ctx.Temporaries[2] = values.NewString("prop")

	inst1 := &opcodes.Instruction{
		Opcode: opcodes.OP_FETCH_STATIC_PROP_IS,
		Op1:    1,
		Op2:    2,
		Result: 3,
	}
	inst1.OpType1, inst1.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR)

	err := vm.executeFetchStaticPropertyIsset(ctx, inst1)
	if err == nil {
		t.Error("Expected error for non-string class name in FETCH_STATIC_PROP_IS")
	}

	// Test FETCH_STATIC_PROP_RW with non-string property name
	ctx.Temporaries[1] = values.NewString("TestClass")
	ctx.Temporaries[2] = values.NewInt(456) // Invalid property name

	inst2 := &opcodes.Instruction{
		Opcode: opcodes.OP_FETCH_STATIC_PROP_RW,
		Op1:    1,
		Op2:    2,
		Result: 3,
	}
	inst2.OpType1, inst2.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR)

	err = vm.executeFetchStaticPropertyReadWrite(ctx, inst2)
	if err == nil {
		t.Error("Expected error for non-string property name in FETCH_STATIC_PROP_RW")
	}
}

// Helper function to compare values for testing
func valuesEqualForTest(a, b *values.Value) bool {
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

func TestStrlenOpcode(t *testing.T) {
	tests := []struct {
		name     string
		input    *values.Value
		expected int64
	}{
		{"string 'hello' length", values.NewString("hello"), 5},
		{"string 'world!' length", values.NewString("world!"), 6},
		{"empty string length", values.NewString(""), 0},
		{"string with spaces length", values.NewString("hello world"), 11},
		{"unicode string length", values.NewString("café"), 5},         // Note: Go counts bytes, not runes
		{"int converted to string length", values.NewInt(123), 3},      // "123"
		{"float converted to string length", values.NewFloat(3.14), 4}, // "3.14"
		{"null converted to string length", values.NewNull(), 0},       // ""
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vm := NewVirtualMachine()
			ctx := NewExecutionContext()

			ctx.Temporaries = make(map[uint32]*values.Value)
			ctx.Temporaries[0] = test.input

			// Create STRLEN instruction
			op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)
			inst := opcodes.Instruction{
				Opcode:  opcodes.OP_STRLEN,
				OpType1: op1Type,
				OpType2: op2Type,
				Op1:     0, // Input from temporary variable 0
				Result:  1, // Store in temporary variable 1
			}

			// Execute STRLEN
			err := vm.executeStrlen(ctx, &inst)
			if err != nil {
				t.Fatalf("STRLEN execution failed: %v", err)
			}

			// Check result
			result := ctx.Temporaries[1]
			if result == nil {
				t.Fatal("STRLEN result is nil")
			}

			if result.Type != values.TypeInt {
				t.Errorf("Expected int type, got %v", result.Type)
			}

			if result.Data.(int64) != test.expected {
				t.Errorf("Expected length %v, got %v", test.expected, result.Data.(int64))
			}
		})
	}
}

func TestSubstrOpcode(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		start    int64
		expected string
	}{
		{"substr from beginning", "hello world", 0, "hello world"},
		{"substr from middle", "hello world", 6, "world"},
		{"substr from position 1", "hello", 1, "ello"},
		{"negative start position", "hello", -3, "llo"},
		{"start beyond string", "hello", 10, ""},
		{"negative start beyond string", "hello", -10, "hello"}, // In PHP, this would be from beginning
		{"empty string", "", 0, ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vm := NewVirtualMachine()
			ctx := NewExecutionContext()

			ctx.Temporaries = make(map[uint32]*values.Value)
			ctx.Temporaries[0] = values.NewString(test.input)
			ctx.Temporaries[1] = values.NewInt(test.start)

			// Create SUBSTR instruction
			op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR)
			inst := opcodes.Instruction{
				Opcode:  opcodes.OP_SUBSTR,
				OpType1: op1Type,
				OpType2: op2Type,
				Op1:     0, // String from temporary variable 0
				Op2:     1, // Start position from temporary variable 1
				Result:  2, // Store in temporary variable 2
			}

			// Execute SUBSTR
			err := vm.executeSubstr(ctx, &inst)
			if err != nil {
				t.Fatalf("SUBSTR execution failed: %v", err)
			}

			// Check result
			result := ctx.Temporaries[2]
			if result == nil {
				t.Fatal("SUBSTR result is nil")
			}

			if result.Type != values.TypeString {
				t.Errorf("Expected string type, got %v", result.Type)
			}

			actual := result.Data.(string)
			if actual != test.expected {
				t.Errorf("Expected substring '%v', got '%v'", test.expected, actual)
			}
		})
	}
}

func TestStrposOpcode(t *testing.T) {
	tests := []struct {
		name      string
		haystack  string
		needle    string
		expected  interface{} // int64 for found position, bool for false (not found)
		expectInt bool        // true if expecting int result, false if expecting bool
	}{
		{"find at beginning", "hello world", "hello", int64(0), true},
		{"find in middle", "hello world", "world", int64(6), true},
		{"find single character", "hello", "e", int64(1), true},
		{"not found", "hello world", "xyz", false, false},
		{"empty needle", "hello", "", int64(0), true}, // Empty needle found at position 0
		{"needle longer than haystack", "hi", "hello", false, false},
		{"case sensitive search", "Hello", "hello", false, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vm := NewVirtualMachine()
			ctx := NewExecutionContext()

			ctx.Temporaries = make(map[uint32]*values.Value)
			ctx.Temporaries[0] = values.NewString(test.haystack)
			ctx.Temporaries[1] = values.NewString(test.needle)

			// Create STRPOS instruction
			op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR)
			inst := opcodes.Instruction{
				Opcode:  opcodes.OP_STRPOS,
				OpType1: op1Type,
				OpType2: op2Type,
				Op1:     0, // Haystack from temporary variable 0
				Op2:     1, // Needle from temporary variable 1
				Result:  2, // Store in temporary variable 2
			}

			// Execute STRPOS
			err := vm.executeStrpos(ctx, &inst)
			if err != nil {
				t.Fatalf("STRPOS execution failed: %v", err)
			}

			// Check result
			result := ctx.Temporaries[2]
			if result == nil {
				t.Fatal("STRPOS result is nil")
			}

			if test.expectInt {
				if result.Type != values.TypeInt {
					t.Errorf("Expected int type, got %v", result.Type)
				}
				if result.Data.(int64) != test.expected.(int64) {
					t.Errorf("Expected position %v, got %v", test.expected, result.Data.(int64))
				}
			} else {
				if result.Type != values.TypeBool {
					t.Errorf("Expected bool type, got %v", result.Type)
				}
				if result.Data.(bool) != test.expected.(bool) {
					t.Errorf("Expected %v, got %v", test.expected, result.Data.(bool))
				}
			}
		})
	}
}

func TestStrtolowerOpcode(t *testing.T) {
	tests := []struct {
		name     string
		input    *values.Value
		expected string
	}{
		{"uppercase string", values.NewString("HELLO WORLD"), "hello world"},
		{"mixed case string", values.NewString("Hello World"), "hello world"},
		{"already lowercase", values.NewString("hello world"), "hello world"},
		{"string with numbers", values.NewString("Hello123"), "hello123"},
		{"empty string", values.NewString(""), ""},
		{"string with special chars", values.NewString("HELLO-WORLD!"), "hello-world!"},
		{"int converted to string", values.NewInt(123), "123"}, // Numbers remain unchanged
		{"null converted to string", values.NewNull(), ""},     // Empty string
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vm := NewVirtualMachine()
			ctx := NewExecutionContext()

			ctx.Temporaries = make(map[uint32]*values.Value)
			ctx.Temporaries[0] = test.input

			// Create STRTOLOWER instruction
			op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)
			inst := opcodes.Instruction{
				Opcode:  opcodes.OP_STRTOLOWER,
				OpType1: op1Type,
				OpType2: op2Type,
				Op1:     0, // Input from temporary variable 0
				Result:  1, // Store in temporary variable 1
			}

			// Execute STRTOLOWER
			err := vm.executeStrtolower(ctx, &inst)
			if err != nil {
				t.Fatalf("STRTOLOWER execution failed: %v", err)
			}

			// Check result
			result := ctx.Temporaries[1]
			if result == nil {
				t.Fatal("STRTOLOWER result is nil")
			}

			if result.Type != values.TypeString {
				t.Errorf("Expected string type, got %v", result.Type)
			}

			actual := result.Data.(string)
			if actual != test.expected {
				t.Errorf("Expected '%v', got '%v'", test.expected, actual)
			}
		})
	}
}

func TestStrtoupperOpcode(t *testing.T) {
	tests := []struct {
		name     string
		input    *values.Value
		expected string
	}{
		{"lowercase string", values.NewString("hello world"), "HELLO WORLD"},
		{"mixed case string", values.NewString("Hello World"), "HELLO WORLD"},
		{"already uppercase", values.NewString("HELLO WORLD"), "HELLO WORLD"},
		{"string with numbers", values.NewString("hello123"), "HELLO123"},
		{"empty string", values.NewString(""), ""},
		{"string with special chars", values.NewString("hello-world!"), "HELLO-WORLD!"},
		{"int converted to string", values.NewInt(123), "123"}, // Numbers remain unchanged
		{"null converted to string", values.NewNull(), ""},     // Empty string
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vm := NewVirtualMachine()
			ctx := NewExecutionContext()

			ctx.Temporaries = make(map[uint32]*values.Value)
			ctx.Temporaries[0] = test.input

			// Create STRTOUPPER instruction
			op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)
			inst := opcodes.Instruction{
				Opcode:  opcodes.OP_STRTOUPPER,
				OpType1: op1Type,
				OpType2: op2Type,
				Op1:     0, // Input from temporary variable 0
				Result:  1, // Store in temporary variable 1
			}

			// Execute STRTOUPPER
			err := vm.executeStrtoupper(ctx, &inst)
			if err != nil {
				t.Fatalf("STRTOUPPER execution failed: %v", err)
			}

			// Check result
			result := ctx.Temporaries[1]
			if result == nil {
				t.Fatal("STRTOUPPER result is nil")
			}

			if result.Type != values.TypeString {
				t.Errorf("Expected string type, got %v", result.Type)
			}

			actual := result.Data.(string)
			if actual != test.expected {
				t.Errorf("Expected '%v', got '%v'", test.expected, actual)
			}
		})
	}
}

// Test comprehensive string operations simulation
func TestStringOperationsSimulation(t *testing.T) {
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	// Simulate: strlen(strtoupper(substr("Hello World", 6)))
	// Expected: 5 (length of "WORLD")

	originalString := values.NewString("Hello World")
	startPos := values.NewInt(6)

	ctx.Temporaries = make(map[uint32]*values.Value)
	ctx.Temporaries[0] = originalString
	ctx.Temporaries[1] = startPos

	op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR)
	op1TypeSingle, op2TypeSingle := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)

	// 1. Extract substring: substr("Hello World", 6) -> "World"
	substrInst := opcodes.Instruction{
		Opcode:  opcodes.OP_SUBSTR,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     0, // Original string
		Op2:     1, // Start position
		Result:  2, // Store substring in temp var 2
	}

	err := vm.executeSubstr(ctx, &substrInst)
	if err != nil {
		t.Fatalf("SUBSTR execution failed: %v", err)
	}

	substring := ctx.Temporaries[2]
	if substring.Data.(string) != "World" {
		t.Errorf("Expected substring 'World', got '%v'", substring.Data.(string))
	}

	// 2. Convert to uppercase: strtoupper("World") -> "WORLD"
	upperInst := opcodes.Instruction{
		Opcode:  opcodes.OP_STRTOUPPER,
		OpType1: op1TypeSingle,
		OpType2: op2TypeSingle,
		Op1:     2, // Substring from step 1
		Result:  3, // Store uppercase result in temp var 3
	}

	err = vm.executeStrtoupper(ctx, &upperInst)
	if err != nil {
		t.Fatalf("STRTOUPPER execution failed: %v", err)
	}

	upperString := ctx.Temporaries[3]
	if upperString.Data.(string) != "WORLD" {
		t.Errorf("Expected uppercase 'WORLD', got '%v'", upperString.Data.(string))
	}

	// 3. Get length: strlen("WORLD") -> 5
	strlenInst := opcodes.Instruction{
		Opcode:  opcodes.OP_STRLEN,
		OpType1: op1TypeSingle,
		OpType2: op2TypeSingle,
		Op1:     3, // Uppercase string from step 2
		Result:  4, // Store length in temp var 4
	}

	err = vm.executeStrlen(ctx, &strlenInst)
	if err != nil {
		t.Fatalf("STRLEN execution failed: %v", err)
	}

	length := ctx.Temporaries[4]
	if !length.IsInt() || length.Data.(int64) != 5 {
		t.Errorf("Expected length 5, got %v", length.Data)
	}

	// 4. Test strpos: strpos("WORLD", "OR") -> 1
	needle := values.NewString("OR")
	ctx.Temporaries[5] = needle

	strposInst := opcodes.Instruction{
		Opcode:  opcodes.OP_STRPOS,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     3, // "WORLD" from step 2
		Op2:     5, // "OR"
		Result:  6, // Store position in temp var 6
	}

	err = vm.executeStrpos(ctx, &strposInst)
	if err != nil {
		t.Fatalf("STRPOS execution failed: %v", err)
	}

	position := ctx.Temporaries[6]
	if !position.IsInt() || position.Data.(int64) != 1 {
		t.Errorf("Expected position 1, got %v", position.Data)
	}
}

func TestSwitchLongOpcode(t *testing.T) {
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	// Create jump table: {1: 10, 2: 20, 3: 30, -1: 99} (default case)
	jumpTable := values.NewArray()
	jumpTableData := jumpTable.Data.(*values.Array)
	jumpTableData.Elements[1] = values.NewInt(10)  // case 1: jump to IP 10
	jumpTableData.Elements[2] = values.NewInt(20)  // case 2: jump to IP 20
	jumpTableData.Elements[3] = values.NewInt(30)  // case 3: jump to IP 30
	jumpTableData.Elements[-1] = values.NewInt(99) // default: jump to IP 99

	tests := []struct {
		name       string
		switchVal  *values.Value
		expectedIP int
	}{
		{"match case 1", values.NewInt(1), 10},
		{"match case 2", values.NewInt(2), 20},
		{"match case 3", values.NewInt(3), 30},
		{"no match - default", values.NewInt(999), 99},
		{"string converts to int", values.NewString("2"), 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx.IP = 0 // Reset IP
			ctx.Temporaries[1] = tt.switchVal
			ctx.Temporaries[2] = jumpTable

			inst := &opcodes.Instruction{
				Opcode: opcodes.OP_SWITCH_LONG,
				Op1:    1, // Switch value
				Op2:    2, // Jump table
				Result: 0, // Unused
			}
			inst.OpType1, inst.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR, opcodes.IS_UNUSED)

			err := vm.executeSwitchLong(ctx, inst)
			if err != nil {
				t.Fatalf("executeSwitchLong failed: %v", err)
			}

			if ctx.IP != tt.expectedIP {
				t.Errorf("Expected IP %d, got %d", tt.expectedIP, ctx.IP)
			}
		})
	}
}

func TestSwitchStringOpcode(t *testing.T) {
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	// Create jump table for string cases
	jumpTable := values.NewArray()
	jumpTableData := jumpTable.Data.(*values.Array)
	jumpTableData.Elements["hello"] = values.NewInt(10)
	jumpTableData.Elements["world"] = values.NewInt(20)
	jumpTableData.Elements["test"] = values.NewInt(30)
	jumpTableData.Elements["__default__"] = values.NewInt(99)

	tests := []struct {
		name       string
		switchVal  *values.Value
		expectedIP int
	}{
		{"match hello", values.NewString("hello"), 10},
		{"match world", values.NewString("world"), 20},
		{"match test", values.NewString("test"), 30},
		{"no match - default", values.NewString("unknown"), 99},
		{"int converts to string", values.NewInt(123), 99}, // No match, go to default
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx.IP = 0 // Reset IP
			ctx.Temporaries[1] = tt.switchVal
			ctx.Temporaries[2] = jumpTable

			inst := &opcodes.Instruction{
				Opcode: opcodes.OP_SWITCH_STRING,
				Op1:    1, // Switch value
				Op2:    2, // Jump table
				Result: 0, // Unused
			}
			inst.OpType1, inst.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR, opcodes.IS_UNUSED)

			err := vm.executeSwitchString(ctx, inst)
			if err != nil {
				t.Fatalf("executeSwitchString failed: %v", err)
			}

			if ctx.IP != tt.expectedIP {
				t.Errorf("Expected IP %d, got %d", tt.expectedIP, ctx.IP)
			}
		})
	}
}

func TestDeclareConstOpcode(t *testing.T) {
	tests := []struct {
		name         string
		constName    string
		constValue   *values.Value
		expectRedecl bool
	}{
		{"declare string constant", "TEST_STRING", values.NewString("hello"), false},
		{"declare int constant", "TEST_INT", values.NewInt(42), false},
		{"declare bool constant", "TEST_BOOL", values.NewBool(true), false},
		{"declare null constant", "TEST_NULL", values.NewNull(), false},
		{"redeclare constant", "TEST_STRING", values.NewString("world"), true},
	}

	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx.Temporaries[1] = values.NewString(tt.constName)
			ctx.Temporaries[2] = tt.constValue

			inst := &opcodes.Instruction{
				Opcode: opcodes.OP_DECLARE_CONST,
				Op1:    1, // Constant name
				Op2:    2, // Constant value
				Result: 0, // Unused
			}
			inst.OpType1, inst.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR, opcodes.IS_UNUSED)

			err := vm.executeDeclareConst(ctx, inst)
			if err != nil {
				t.Fatalf("executeDeclareConst failed: %v", err)
			}

			// Check that constant exists in global constants
			if !tt.expectRedecl {
				declaredValue := ctx.GlobalConstants[tt.constName]
				if declaredValue == nil {
					t.Fatal("Constant should be declared")
				}

				if !valuesEqual(declaredValue, tt.constValue) {
					t.Errorf("Expected constant value %v, got %v", tt.constValue, declaredValue)
				}
			}
		})
	}
}

func TestVerifyReturnTypeOpcode(t *testing.T) {
	tests := []struct {
		name         string
		returnValue  *values.Value
		expectedType string
		expectError  bool
	}{
		{"valid int", values.NewInt(42), "int", false},
		{"valid string", values.NewString("test"), "string", false},
		{"valid bool", values.NewBool(true), "bool", false},
		{"valid float", values.NewFloat(3.14), "float", false},
		{"valid array", values.NewArray(), "array", false},
		{"valid null", values.NewNull(), "null", false},
		{"mixed allows anything", values.NewString("anything"), "mixed", false},
		{"invalid type", values.NewString("not_int"), "int", true},
		{"no type constraint", values.NewInt(123), "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := NewVirtualMachine()
			ctx := NewExecutionContext()

			ctx.Temporaries[1] = tt.returnValue
			if tt.expectedType != "" {
				ctx.Temporaries[2] = values.NewString(tt.expectedType)
			}

			inst := &opcodes.Instruction{
				Opcode: opcodes.OP_VERIFY_RETURN_TYPE,
				Op1:    1, // Return value
				Op2:    2, // Expected type
				Result: 0, // Unused
			}
			inst.OpType1, inst.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR, opcodes.IS_UNUSED)

			err := vm.executeVerifyReturnType(ctx, inst)

			if tt.expectError {
				if err == nil {
					t.Error("Expected type verification error")
				}
			} else {
				if err != nil {
					t.Fatalf("executeVerifyReturnType failed: %v", err)
				}
			}
		})
	}
}

func TestSendUnpackOpcode(t *testing.T) {
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	// Create array of arguments to unpack
	argsArray := values.NewArray()
	argsData := argsArray.Data.(*values.Array)
	argsData.Elements[0] = values.NewString("arg1")
	argsData.Elements[1] = values.NewInt(42)
	argsData.Elements[2] = values.NewBool(true)
	argsData.NextIndex = 3

	// Setup call context
	ctx.CallContext = &CallContext{
		FunctionName: "test_function",
		NumArgs:      0,
	}
	ctx.CallArguments = []*values.Value{values.NewString("existing_arg")}

	ctx.Temporaries[1] = argsArray

	inst := &opcodes.Instruction{
		Opcode: opcodes.OP_SEND_UNPACK,
		Op1:    1, // Arguments to unpack
		Op2:    0, // Unused
		Result: 0, // Unused
	}
	inst.OpType1, inst.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_UNUSED)

	err := vm.executeSendUnpack(ctx, inst)
	if err != nil {
		t.Fatalf("executeSendUnpack failed: %v", err)
	}

	// Check that arguments were unpacked and added
	expectedArgCount := 4 // 1 existing + 3 unpacked
	if len(ctx.CallArguments) != expectedArgCount {
		t.Errorf("Expected %d arguments, got %d", expectedArgCount, len(ctx.CallArguments))
	}

	// Check specific arguments
	if ctx.CallArguments[0].ToString() != "existing_arg" {
		t.Error("First argument should be existing_arg")
	}
	if ctx.CallArguments[1].ToString() != "arg1" {
		t.Error("Second argument should be arg1")
	}
	if ctx.CallArguments[2].ToInt() != 42 {
		t.Error("Third argument should be 42")
	}
	if !ctx.CallArguments[3].IsBool() || !ctx.CallArguments[3].ToBool() {
		t.Error("Fourth argument should be true")
	}

	// Check that argument count was updated
	if ctx.CallContext.NumArgs != expectedArgCount {
		t.Errorf("Expected NumArgs %d, got %d", expectedArgCount, ctx.CallContext.NumArgs)
	}
}

func TestSendUnpackWithNonArray(t *testing.T) {
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	// Setup call context
	ctx.CallContext = &CallContext{
		FunctionName: "test_function",
		NumArgs:      0,
	}
	ctx.CallArguments = nil

	// Send non-array value (should be treated as single argument)
	ctx.Temporaries[1] = values.NewString("single_arg")

	inst := &opcodes.Instruction{
		Opcode: opcodes.OP_SEND_UNPACK,
		Op1:    1,
		Op2:    0,
		Result: 0,
	}
	inst.OpType1, inst.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_UNUSED)

	err := vm.executeSendUnpack(ctx, inst)
	if err != nil {
		t.Fatalf("executeSendUnpack failed: %v", err)
	}

	// Check that single argument was added
	if len(ctx.CallArguments) != 1 {
		t.Errorf("Expected 1 argument, got %d", len(ctx.CallArguments))
	}

	if ctx.CallArguments[0].ToString() != "single_arg" {
		t.Errorf("Expected 'single_arg', got %v", ctx.CallArguments[0])
	}
}

func TestSwitchOpcodeErrors(t *testing.T) {
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	// Test SWITCH_LONG with non-array jump table
	ctx.Temporaries[1] = values.NewInt(1)
	ctx.Temporaries[2] = values.NewString("not_array")

	inst := &opcodes.Instruction{
		Opcode: opcodes.OP_SWITCH_LONG,
		Op1:    1,
		Op2:    2,
		Result: 0,
	}
	inst.OpType1, inst.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR, opcodes.IS_UNUSED)

	err := vm.executeSwitchLong(ctx, inst)
	if err == nil {
		t.Error("Expected error for non-array jump table in SWITCH_LONG")
	}

	// Test DECLARE_CONST with non-string name
	ctx.Temporaries[1] = values.NewInt(123) // Invalid name
	ctx.Temporaries[2] = values.NewString("value")

	inst2 := &opcodes.Instruction{
		Opcode: opcodes.OP_DECLARE_CONST,
		Op1:    1,
		Op2:    2,
		Result: 0,
	}
	inst2.OpType1, inst2.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR, opcodes.IS_UNUSED)

	err = vm.executeDeclareConst(ctx, inst2)
	if err == nil {
		t.Error("Expected error for non-string constant name in DECLARE_CONST")
	}
}

func TestCastBoolOpcode(t *testing.T) {
	tests := []struct {
		name     string
		input    *values.Value
		expected bool
	}{
		{"cast string '1' to bool", values.NewString("1"), true},
		{"cast string '0' to bool", values.NewString("0"), false},
		{"cast string '' to bool", values.NewString(""), false},
		{"cast int 42 to bool", values.NewInt(42), true},
		{"cast int 0 to bool", values.NewInt(0), false},
		{"cast float 3.14 to bool", values.NewFloat(3.14), true},
		{"cast float 0.0 to bool", values.NewFloat(0.0), false},
		{"cast null to bool", values.NewNull(), false},
		{"cast true to bool", values.NewBool(true), true},
		{"cast false to bool", values.NewBool(false), false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vm := NewVirtualMachine()
			ctx := NewExecutionContext()

			ctx.Temporaries = make(map[uint32]*values.Value)
			ctx.Temporaries[0] = test.input

			// Create CAST_BOOL instruction
			op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)
			inst := opcodes.Instruction{
				Opcode:  opcodes.OP_CAST_BOOL,
				OpType1: op1Type,
				OpType2: op2Type,
				Op1:     0, // Input from temporary variable 0
				Result:  1, // Store in temporary variable 1
			}

			// Execute CAST_BOOL
			err := vm.executeCastBool(ctx, &inst)
			if err != nil {
				t.Fatalf("CAST_BOOL execution failed: %v", err)
			}

			// Check result
			result := ctx.Temporaries[1]
			if result == nil {
				t.Fatal("CAST_BOOL result is nil")
			}

			if result.Type != values.TypeBool {
				t.Errorf("Expected bool type, got %v", result.Type)
			}

			if result.Data.(bool) != test.expected {
				t.Errorf("Expected %v, got %v", test.expected, result.Data.(bool))
			}
		})
	}
}

func TestCastLongOpcode(t *testing.T) {
	tests := []struct {
		name     string
		input    *values.Value
		expected int64
	}{
		{"cast string '42' to int", values.NewString("42"), 42},
		{"cast string '3.14' to int", values.NewString("3.14"), 3},
		{"cast string 'abc' to int", values.NewString("abc"), 0},
		{"cast float 3.14 to int", values.NewFloat(3.14), 3},
		{"cast float 7.99 to int", values.NewFloat(7.99), 7},
		{"cast bool true to int", values.NewBool(true), 1},
		{"cast bool false to int", values.NewBool(false), 0},
		{"cast null to int", values.NewNull(), 0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vm := NewVirtualMachine()
			ctx := NewExecutionContext()

			ctx.Temporaries = make(map[uint32]*values.Value)
			ctx.Temporaries[0] = test.input

			// Create CAST_LONG instruction
			op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)
			inst := opcodes.Instruction{
				Opcode:  opcodes.OP_CAST_LONG,
				OpType1: op1Type,
				OpType2: op2Type,
				Op1:     0, // Input from temporary variable 0
				Result:  1, // Store in temporary variable 1
			}

			// Execute CAST_LONG
			err := vm.executeCastLong(ctx, &inst)
			if err != nil {
				t.Fatalf("CAST_LONG execution failed: %v", err)
			}

			// Check result
			result := ctx.Temporaries[1]
			if result == nil {
				t.Fatal("CAST_LONG result is nil")
			}

			if result.Type != values.TypeInt {
				t.Errorf("Expected int type, got %v", result.Type)
			}

			if result.Data.(int64) != test.expected {
				t.Errorf("Expected %v, got %v", test.expected, result.Data.(int64))
			}
		})
	}
}

func TestCastStringOpcode(t *testing.T) {
	tests := []struct {
		name     string
		input    *values.Value
		expected string
	}{
		{"cast int 42 to string", values.NewInt(42), "42"},
		{"cast float 3.14 to string", values.NewFloat(3.14), "3.14"},
		{"cast bool true to string", values.NewBool(true), "1"},
		{"cast bool false to string", values.NewBool(false), ""},
		{"cast null to string", values.NewNull(), ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vm := NewVirtualMachine()
			ctx := NewExecutionContext()

			ctx.Temporaries = make(map[uint32]*values.Value)
			ctx.Temporaries[0] = test.input

			// Create CAST_STRING instruction
			op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)
			inst := opcodes.Instruction{
				Opcode:  opcodes.OP_CAST_STRING,
				OpType1: op1Type,
				OpType2: op2Type,
				Op1:     0, // Input from temporary variable 0
				Result:  1, // Store in temporary variable 1
			}

			// Execute CAST_STRING
			err := vm.executeCastString(ctx, &inst)
			if err != nil {
				t.Fatalf("CAST_STRING execution failed: %v", err)
			}

			// Check result
			result := ctx.Temporaries[1]
			if result == nil {
				t.Fatal("CAST_STRING result is nil")
			}

			if result.Type != values.TypeString {
				t.Errorf("Expected string type, got %v", result.Type)
			}

			if result.Data.(string) != test.expected {
				t.Errorf("Expected '%v', got '%v'", test.expected, result.Data.(string))
			}
		})
	}
}

func TestCastArrayOpcode(t *testing.T) {
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	ctx.Temporaries = make(map[uint32]*values.Value)

	// Test casting string to array
	ctx.Temporaries[0] = values.NewString("hello")

	op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)
	inst := opcodes.Instruction{
		Opcode:  opcodes.OP_CAST_ARRAY,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     0, // Input from temporary variable 0
		Result:  1, // Store in temporary variable 1
	}

	// Execute CAST_ARRAY
	err := vm.executeCastArray(ctx, &inst)
	if err != nil {
		t.Fatalf("CAST_ARRAY execution failed: %v", err)
	}

	// Check result
	result := ctx.Temporaries[1]
	if result == nil {
		t.Fatal("CAST_ARRAY result is nil")
	}

	if result.Type != values.TypeArray {
		t.Errorf("Expected array type, got %v", result.Type)
	}

	if result.ArrayCount() != 1 {
		t.Errorf("Expected array with 1 element, got %d elements", result.ArrayCount())
	}

	// Check array element
	elem := result.ArrayGet(values.NewInt(0))
	if elem.Data.(string) != "hello" {
		t.Errorf("Expected array element 'hello', got %v", elem.Data)
	}
}

func TestIsTypeOpcode(t *testing.T) {
	tests := []struct {
		name     string
		value    *values.Value
		typeName string
		expected bool
	}{
		{"int is int", values.NewInt(42), "int", true},
		{"int is integer", values.NewInt(42), "integer", true},
		{"int is not string", values.NewInt(42), "string", false},
		{"string is string", values.NewString("hello"), "string", true},
		{"string is not int", values.NewString("hello"), "int", false},
		{"float is float", values.NewFloat(3.14), "float", true},
		{"float is double", values.NewFloat(3.14), "double", true},
		{"bool is bool", values.NewBool(true), "bool", true},
		{"bool is boolean", values.NewBool(true), "boolean", true},
		{"null is null", values.NewNull(), "null", true},
		{"null is not int", values.NewNull(), "int", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vm := NewVirtualMachine()
			ctx := NewExecutionContext()

			ctx.Temporaries = make(map[uint32]*values.Value)
			ctx.Temporaries[0] = test.value
			ctx.Temporaries[1] = values.NewString(test.typeName)

			// Create IS_TYPE instruction
			op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR)
			inst := opcodes.Instruction{
				Opcode:  opcodes.OP_IS_TYPE,
				OpType1: op1Type,
				OpType2: op2Type,
				Op1:     0, // Value from temporary variable 0
				Op2:     1, // Type name from temporary variable 1
				Result:  2, // Store in temporary variable 2
			}

			// Execute IS_TYPE
			err := vm.executeIsType(ctx, &inst)
			if err != nil {
				t.Fatalf("IS_TYPE execution failed: %v", err)
			}

			// Check result
			result := ctx.Temporaries[2]
			if result == nil {
				t.Fatal("IS_TYPE result is nil")
			}

			if result.Type != values.TypeBool {
				t.Errorf("Expected bool type, got %v", result.Type)
			}

			if result.Data.(bool) != test.expected {
				t.Errorf("Expected %v, got %v", test.expected, result.Data.(bool))
			}
		})
	}
}

func TestVerifyArgTypeOpcode(t *testing.T) {
	tests := []struct {
		name         string
		argument     *values.Value
		expectedType string
		shouldError  bool
	}{
		{"valid int argument", values.NewInt(42), "int", false},
		{"valid string argument", values.NewString("hello"), "string", false},
		{"invalid int argument", values.NewString("hello"), "int", true},
		{"invalid string argument", values.NewInt(42), "string", true},
		{"unknown type allows anything", values.NewInt(42), "unknown", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vm := NewVirtualMachine()
			ctx := NewExecutionContext()

			ctx.Temporaries = make(map[uint32]*values.Value)
			ctx.Temporaries[0] = test.argument
			ctx.Temporaries[1] = values.NewString(test.expectedType)

			// Create VERIFY_ARG_TYPE instruction
			op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR)
			inst := opcodes.Instruction{
				Opcode:  opcodes.OP_VERIFY_ARG_TYPE,
				OpType1: op1Type,
				OpType2: op2Type,
				Op1:     0, // Argument from temporary variable 0
				Op2:     1, // Expected type from temporary variable 1
				Result:  2, // Store in temporary variable 2
			}

			// Execute VERIFY_ARG_TYPE
			err := vm.executeVerifyArgType(ctx, &inst)

			if test.shouldError {
				if err == nil {
					t.Error("Expected error for type verification failure, but got none")
				}
			} else {
				if err != nil {
					t.Fatalf("VERIFY_ARG_TYPE execution failed: %v", err)
				}

				// Check result - should be the original argument if valid
				result := ctx.Temporaries[2]
				if result == nil {
					t.Fatal("VERIFY_ARG_TYPE result is nil")
				}

				if result.Type != test.argument.Type {
					t.Errorf("Expected result type %v, got %v", test.argument.Type, result.Type)
				}
			}
		})
	}
}

func TestInstanceofOpcode(t *testing.T) {
	tests := []struct {
		name      string
		object    *values.Value
		className string
		expected  bool
	}{
		{"object instanceof correct class", values.NewObject("MyClass"), "MyClass", true},
		{"object instanceof wrong class", values.NewObject("MyClass"), "OtherClass", false},
		{"non-object instanceof class", values.NewString("hello"), "MyClass", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vm := NewVirtualMachine()
			ctx := NewExecutionContext()

			ctx.Temporaries = make(map[uint32]*values.Value)
			ctx.Temporaries[0] = test.object
			ctx.Temporaries[1] = values.NewString(test.className)

			// Create INSTANCEOF instruction
			op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR)
			inst := opcodes.Instruction{
				Opcode:  opcodes.OP_INSTANCEOF,
				OpType1: op1Type,
				OpType2: op2Type,
				Op1:     0, // Object from temporary variable 0
				Op2:     1, // Class name from temporary variable 1
				Result:  2, // Store in temporary variable 2
			}

			// Execute INSTANCEOF
			err := vm.executeInstanceof(ctx, &inst)
			if err != nil {
				t.Fatalf("INSTANCEOF execution failed: %v", err)
			}

			// Check result
			result := ctx.Temporaries[2]
			if result == nil {
				t.Fatal("INSTANCEOF result is nil")
			}

			if result.Type != values.TypeBool {
				t.Errorf("Expected bool type, got %v", result.Type)
			}

			if result.Data.(bool) != test.expected {
				t.Errorf("Expected %v, got %v", test.expected, result.Data.(bool))
			}
		})
	}
}

// Test comprehensive type casting simulation
func TestTypeCastingSimulation(t *testing.T) {
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	// Simulate: (bool)$value, (int)$value, (string)$value
	inputValue := values.NewString("42")

	ctx.Temporaries = make(map[uint32]*values.Value)
	ctx.Temporaries[0] = inputValue

	op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)

	// 1. Cast to bool
	boolInst := opcodes.Instruction{
		Opcode:  opcodes.OP_CAST_BOOL,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     0, // Input
		Result:  1, // Store bool result
	}

	err := vm.executeCastBool(ctx, &boolInst)
	if err != nil {
		t.Fatalf("Bool cast failed: %v", err)
	}

	boolResult := ctx.Temporaries[1]
	if !boolResult.IsBool() || !boolResult.Data.(bool) {
		t.Errorf("Expected bool true, got %v", boolResult.Data)
	}

	// 2. Cast to int
	intInst := opcodes.Instruction{
		Opcode:  opcodes.OP_CAST_LONG,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     0, // Input
		Result:  2, // Store int result
	}

	err = vm.executeCastLong(ctx, &intInst)
	if err != nil {
		t.Fatalf("Int cast failed: %v", err)
	}

	intResult := ctx.Temporaries[2]
	if !intResult.IsInt() || intResult.Data.(int64) != 42 {
		t.Errorf("Expected int 42, got %v", intResult.Data)
	}

	// 3. Cast to string (should be unchanged)
	stringInst := opcodes.Instruction{
		Opcode:  opcodes.OP_CAST_STRING,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     0, // Input
		Result:  3, // Store string result
	}

	err = vm.executeCastString(ctx, &stringInst)
	if err != nil {
		t.Fatalf("String cast failed: %v", err)
	}

	stringResult := ctx.Temporaries[3]
	if !stringResult.IsString() || stringResult.Data.(string) != "42" {
		t.Errorf("Expected string '42', got %v", stringResult.Data)
	}
}
