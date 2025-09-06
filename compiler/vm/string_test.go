package vm

import (
	"testing"

	"github.com/wudi/php-parser/compiler/opcodes"
	"github.com/wudi/php-parser/compiler/values"
)

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
