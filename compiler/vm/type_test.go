package vm

import (
	"testing"

	"github.com/wudi/php-parser/compiler/opcodes"
	"github.com/wudi/php-parser/compiler/values"
)

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
