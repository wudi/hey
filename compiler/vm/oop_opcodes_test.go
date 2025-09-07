package vm

import (
	"testing"

	"github.com/wudi/php-parser/compiler/opcodes"
	"github.com/wudi/php-parser/compiler/values"
)

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
