package vm

import (
	"testing"

	"github.com/wudi/hey/compiler/opcodes"
	"github.com/wudi/hey/compiler/values"
)

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
