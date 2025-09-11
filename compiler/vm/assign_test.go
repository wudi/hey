package vm

import (
	"testing"

	"github.com/wudi/hey/compiler/opcodes"
	"github.com/wudi/hey/compiler/values"
)

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
