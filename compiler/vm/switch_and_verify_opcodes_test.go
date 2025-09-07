package vm

import (
	"testing"

	"github.com/wudi/php-parser/compiler/opcodes"
	"github.com/wudi/php-parser/compiler/values"
)

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
