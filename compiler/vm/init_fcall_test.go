package vm

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wudi/php-parser/compiler/opcodes"
	"github.com/wudi/php-parser/compiler/runtime"
	"github.com/wudi/php-parser/compiler/values"
)

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
