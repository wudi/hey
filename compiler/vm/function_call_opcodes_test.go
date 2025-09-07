package vm

import (
	"testing"

	"github.com/wudi/php-parser/compiler/opcodes"
	"github.com/wudi/php-parser/compiler/values"
)

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
