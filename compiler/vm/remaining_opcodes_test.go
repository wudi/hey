package vm

import (
	"testing"

	"github.com/wudi/php-parser/compiler/opcodes"
	"github.com/wudi/php-parser/compiler/values"
)

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
			vm := NewVirtualMachine()
			ctx := NewExecutionContext()

			// Setup class if needed
			if tt.setupClass {
				class := &Class{
					Name:       tt.className,
					Properties: make(map[string]*Property),
					Methods:    make(map[string]*Function),
					Constants:  make(map[string]*values.Value),
				}

				if tt.isAbstract {
					// Add abstract method to make class abstract
					if tt.className == "MyClass" {
						// Add an abstract method (no implementation)
						class.Methods["abstractMethod"] = &Function{
							Name:         "abstractMethod",
							Instructions: []opcodes.Instruction{}, // Empty = abstract
							Parameters:   []Parameter{},
							Constants:    []*values.Value{},
							IsVariadic:   false,
							IsGenerator:  false,
						}
					}
				}

				ctx.Classes[tt.className] = class
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
