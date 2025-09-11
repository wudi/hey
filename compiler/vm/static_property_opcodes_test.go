package vm

import (
	"testing"

	"github.com/wudi/hey/compiler/opcodes"
	"github.com/wudi/hey/compiler/registry"
	"github.com/wudi/hey/compiler/values"
)

func TestFetchStaticPropertyIssetOpcode(t *testing.T) {
	tests := []struct {
		name        string
		className   string
		propName    string
		setupClass  bool
		setupProp   bool
		propValue   *values.Value
		expectedSet bool
		expectError bool
	}{
		{
			name:        "property exists and is set",
			className:   "TestClass",
			propName:    "staticProp",
			setupClass:  true,
			setupProp:   true,
			propValue:   values.NewString("test_value"),
			expectedSet: true,
			expectError: false,
		},
		{
			name:        "property exists but is null",
			className:   "TestClass",
			propName:    "nullProp",
			setupClass:  true,
			setupProp:   true,
			propValue:   values.NewNull(),
			expectedSet: false,
			expectError: false,
		},
		{
			name:        "property doesn't exist",
			className:   "TestClass",
			propName:    "nonExistentProp",
			setupClass:  true,
			setupProp:   false,
			propValue:   nil,
			expectedSet: false,
			expectError: false,
		},
		{
			name:        "class doesn't exist",
			className:   "NonExistentClass",
			propName:    "anyProp",
			setupClass:  false,
			setupProp:   false,
			propValue:   nil,
			expectedSet: false,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize registry for each test
			registry.Initialize()

			vm := NewVirtualMachine()
			ctx := NewExecutionContext()

			// Setup class if needed using registry
			if tt.setupClass {
				vm.setStaticPropertyInRegistry(tt.className, tt.propName+"_dummy", values.NewNull())

				// Setup property if needed
				if tt.setupProp {
					vm.setStaticPropertyInRegistry(tt.className, tt.propName, tt.propValue)
				}
			}

			// Setup instruction operands
			ctx.Temporaries[1] = values.NewString(tt.className)
			ctx.Temporaries[2] = values.NewString(tt.propName)

			// Create FETCH_STATIC_PROP_IS instruction
			inst := &opcodes.Instruction{
				Opcode: opcodes.OP_FETCH_STATIC_PROP_IS,
				Op1:    1, // Class name
				Op2:    2, // Property name
				Result: 3, // Result location
			}
			inst.OpType1, inst.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR)

			err := vm.executeFetchStaticPropertyIsset(ctx, inst)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("executeFetchStaticPropertyIsset failed: %v", err)
			}

			// Check result
			result := ctx.Temporaries[3]
			if result == nil {
				t.Fatal("Result should not be nil")
			}

			if !result.IsBool() {
				t.Errorf("Expected boolean result, got %T", result.Data)
			}

			if result.ToBool() != tt.expectedSet {
				t.Errorf("Expected isset=%t, got isset=%t", tt.expectedSet, result.ToBool())
			}
		})
	}
}

func TestFetchStaticPropertyReadWriteOpcode(t *testing.T) {
	tests := []struct {
		name         string
		className    string
		propName     string
		setupClass   bool
		setupProp    bool
		initialValue *values.Value
		expectError  bool
	}{
		{
			name:         "existing property read-write",
			className:    "TestClass",
			propName:     "existingProp",
			setupClass:   true,
			setupProp:    true,
			initialValue: values.NewString("initial_value"),
			expectError:  false,
		},
		{
			name:         "non-existing property - should create",
			className:    "TestClass",
			propName:     "newProp",
			setupClass:   true,
			setupProp:    false,
			initialValue: nil,
			expectError:  false,
		},
		{
			name:         "non-existing class - should create",
			className:    "NewClass",
			propName:     "newProp",
			setupClass:   false,
			setupProp:    false,
			initialValue: nil,
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize registry for each test
			registry.Initialize()

			vm := NewVirtualMachine()
			ctx := NewExecutionContext()

			// Setup class if needed
			if tt.setupClass {
				vm.setStaticPropertyInRegistry(tt.className, tt.propName+"_dummy", values.NewNull())

				// Setup property if needed
				if tt.setupProp {
					vm.setStaticPropertyInRegistry(tt.className, tt.propName, tt.initialValue)
				}
			}

			// Setup instruction operands
			ctx.Temporaries[1] = values.NewString(tt.className)
			ctx.Temporaries[2] = values.NewString(tt.propName)

			// Create FETCH_STATIC_PROP_RW instruction
			inst := &opcodes.Instruction{
				Opcode: opcodes.OP_FETCH_STATIC_PROP_RW,
				Op1:    1, // Class name
				Op2:    2, // Property name
				Result: 3, // Result location
			}
			inst.OpType1, inst.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR)

			err := vm.executeFetchStaticPropertyReadWrite(ctx, inst)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("executeFetchStaticPropertyReadWrite failed: %v", err)
			}

			// Check result
			result := ctx.Temporaries[3]
			if result == nil {
				t.Fatal("Result should not be nil")
			}

			// Check that class and property were created if needed
			// Verify class exists in registry
			if _, err := registry.GlobalRegistry.GetClass(tt.className); err != nil {
				t.Errorf("Class %s should exist", tt.className)
			}

			if _, exists := vm.getStaticPropertyFromRegistry(tt.className, tt.propName); !exists {
				t.Errorf("Property %s should exist", tt.propName)
			}

			// Check that the property is marked as static
			if classDesc, err := registry.GlobalRegistry.GetClass(tt.className); err == nil {
				if prop, exists := classDesc.Properties[tt.propName]; exists {
					if !prop.IsStatic {
						t.Error("Property should be marked as static")
					}
				}
			}

			// Check that result matches expected value
			if tt.initialValue != nil {
				if !valuesEqualForTest(result, tt.initialValue) {
					t.Errorf("Expected result %v, got %v", tt.initialValue, result)
				}
			} else {
				// Should be null for newly created properties
				if !result.IsNull() {
					t.Errorf("Expected null for new property, got %v", result)
				}
			}
		})
	}
}

func TestFetchStaticPropertyUnsetOpcode(t *testing.T) {
	// Initialize registry for test
	registry.Initialize()

	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	// Setup class with property
	className := "TestClass"
	propName := "testProp"
	// Setup class with property using compatibility layer
	vm.setStaticPropertyInRegistry(className, propName+"_dummy", values.NewNull())
	vm.setStaticPropertyInRegistry(className, propName, values.NewString("test_value"))

	// Setup instruction operands
	ctx.Temporaries[1] = values.NewString(className)
	ctx.Temporaries[2] = values.NewString(propName)

	// Create FETCH_STATIC_PROP_UNSET instruction
	inst := &opcodes.Instruction{
		Opcode: opcodes.OP_FETCH_STATIC_PROP_UNSET,
		Op1:    1, // Class name
		Op2:    2, // Property name
		Result: 0, // Unused
	}
	inst.OpType1, inst.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR, opcodes.IS_UNUSED)

	err := vm.executeFetchStaticPropertyUnset(ctx, inst)
	if err != nil {
		t.Fatalf("executeFetchStaticPropertyUnset failed: %v", err)
	}

	// Check that property was removed
	if _, exists := vm.getStaticPropertyFromRegistry(className, propName); exists {
		t.Error("Property should have been removed")
	}
}

func TestStaticPropertyOpcodeErrors(t *testing.T) {
	// Initialize registry for test
	registry.Initialize()

	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	// Test FETCH_STATIC_PROP_IS with non-string class name
	ctx.Temporaries[1] = values.NewInt(123) // Invalid class name
	ctx.Temporaries[2] = values.NewString("prop")

	inst1 := &opcodes.Instruction{
		Opcode: opcodes.OP_FETCH_STATIC_PROP_IS,
		Op1:    1,
		Op2:    2,
		Result: 3,
	}
	inst1.OpType1, inst1.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR)

	err := vm.executeFetchStaticPropertyIsset(ctx, inst1)
	if err == nil {
		t.Error("Expected error for non-string class name in FETCH_STATIC_PROP_IS")
	}

	// Test FETCH_STATIC_PROP_RW with non-string property name
	ctx.Temporaries[1] = values.NewString("TestClass")
	ctx.Temporaries[2] = values.NewInt(456) // Invalid property name

	inst2 := &opcodes.Instruction{
		Opcode: opcodes.OP_FETCH_STATIC_PROP_RW,
		Op1:    1,
		Op2:    2,
		Result: 3,
	}
	inst2.OpType1, inst2.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR)

	err = vm.executeFetchStaticPropertyReadWrite(ctx, inst2)
	if err == nil {
		t.Error("Expected error for non-string property name in FETCH_STATIC_PROP_RW")
	}
}

// Helper function to compare values for testing
func valuesEqualForTest(a, b *values.Value) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	if a.Type != b.Type {
		return false
	}

	switch a.Type {
	case values.TypeNull:
		return true
	case values.TypeBool:
		return a.ToBool() == b.ToBool()
	case values.TypeInt:
		return a.ToInt() == b.ToInt()
	case values.TypeFloat:
		return a.ToFloat() == b.ToFloat()
	case values.TypeString:
		return a.ToString() == b.ToString()
	default:
		return a == b
	}
}
