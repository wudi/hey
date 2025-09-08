package vm

import (
	"bytes"
	"testing"

	"github.com/wudi/php-parser/compiler/opcodes"
	"github.com/wudi/php-parser/compiler/values"
)

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
