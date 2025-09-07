package vm

import (
	"testing"

	"github.com/wudi/php-parser/compiler/opcodes"
	"github.com/wudi/php-parser/compiler/values"
)

func TestAssignStaticPropertyOpcode(t *testing.T) {
	tests := []struct {
		name           string
		className      string
		propertyName   string
		assignValue    *values.Value
		expectedResult *values.Value
	}{
		{
			name:           "assign string to static property",
			className:      "TestClass",
			propertyName:   "testProp",
			assignValue:    values.NewString("hello world"),
			expectedResult: values.NewString("hello world"),
		},
		{
			name:           "assign int to static property",
			className:      "MathClass",
			propertyName:   "counter",
			assignValue:    values.NewInt(42),
			expectedResult: values.NewInt(42),
		},
		{
			name:           "assign bool to static property",
			className:      "ConfigClass",
			propertyName:   "enabled",
			assignValue:    values.NewBool(true),
			expectedResult: values.NewBool(true),
		},
		{
			name:           "assign null to static property",
			className:      "TestClass",
			propertyName:   "nullable",
			assignValue:    values.NewNull(),
			expectedResult: values.NewNull(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := NewVirtualMachine()
			ctx := NewExecutionContext()

			// Setup operands: class name, property name, and value
			ctx.Temporaries[1] = values.NewString(tt.className)
			ctx.Temporaries[2] = values.NewString(tt.propertyName)
			ctx.Temporaries[3] = tt.assignValue

			// Create ASSIGN_STATIC_PROP instruction
			inst := &opcodes.Instruction{
				Opcode: opcodes.OP_ASSIGN_STATIC_PROP,
				Op1:    1, // Class name
				Op2:    2, // Property name
				Result: 3, // Value to assign
			}
			inst.OpType1, inst.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR)

			err := vm.executeAssignStaticProperty(ctx, inst)
			if err != nil {
				t.Fatalf("executeAssignStaticProperty failed: %v", err)
			}

			// Check that the class was created
			if ctx.Classes[tt.className] == nil {
				t.Fatal("Class should have been created")
			}

			// Check that the property was created
			prop := ctx.Classes[tt.className].Properties[tt.propertyName]
			if prop == nil {
				t.Fatal("Property should have been created")
			}

			// Check that the property is marked as static
			if !prop.IsStatic {
				t.Error("Property should be marked as static")
			}

			// Check that the property has correct value
			if prop.DefaultValue == nil {
				t.Fatal("Property value should not be nil")
			}

			// Compare values based on type
			if tt.expectedResult.IsNull() {
				if !prop.DefaultValue.IsNull() {
					t.Errorf("Expected null value, got %v", prop.DefaultValue)
				}
			} else if tt.expectedResult.IsString() {
				if !prop.DefaultValue.IsString() || prop.DefaultValue.ToString() != tt.expectedResult.ToString() {
					t.Errorf("Expected string %q, got %v", tt.expectedResult.ToString(), prop.DefaultValue)
				}
			} else if tt.expectedResult.IsInt() {
				if !prop.DefaultValue.IsInt() || prop.DefaultValue.ToInt() != tt.expectedResult.ToInt() {
					t.Errorf("Expected int %d, got %v", tt.expectedResult.ToInt(), prop.DefaultValue)
				}
			} else if tt.expectedResult.IsBool() {
				if !prop.DefaultValue.IsBool() || prop.DefaultValue.ToBool() != tt.expectedResult.ToBool() {
					t.Errorf("Expected bool %t, got %v", tt.expectedResult.ToBool(), prop.DefaultValue)
				}
			}
		})
	}
}

func TestAssignStaticPropertyUpdate(t *testing.T) {
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	className := "TestClass"
	propertyName := "updateProp"

	// First assignment
	ctx.Temporaries[1] = values.NewString(className)
	ctx.Temporaries[2] = values.NewString(propertyName)
	ctx.Temporaries[3] = values.NewString("initial")

	inst := &opcodes.Instruction{
		Opcode: opcodes.OP_ASSIGN_STATIC_PROP,
		Op1:    1,
		Op2:    2,
		Result: 3,
	}
	inst.OpType1, inst.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR)

	err := vm.executeAssignStaticProperty(ctx, inst)
	if err != nil {
		t.Fatalf("First assignment failed: %v", err)
	}

	// Check first assignment
	if ctx.Classes[className].Properties[propertyName].DefaultValue.ToString() != "initial" {
		t.Errorf("Expected 'initial', got %v", ctx.Classes[className].Properties[propertyName].DefaultValue)
	}

	// Second assignment to same property
	ctx.Temporaries[3] = values.NewString("updated")

	err = vm.executeAssignStaticProperty(ctx, inst)
	if err != nil {
		t.Fatalf("Second assignment failed: %v", err)
	}

	// Check that property was updated
	if ctx.Classes[className].Properties[propertyName].DefaultValue.ToString() != "updated" {
		t.Errorf("Expected 'updated', got %v", ctx.Classes[className].Properties[propertyName].DefaultValue)
	}
}

func TestAssignStaticPropertyOpOpcode(t *testing.T) {
	tests := []struct {
		name           string
		initialValue   *values.Value
		operandValue   *values.Value
		expectedResult interface{}
	}{
		{
			name:           "int addition",
			initialValue:   values.NewInt(10),
			operandValue:   values.NewInt(5),
			expectedResult: int64(15),
		},
		{
			name:           "string concatenation",
			initialValue:   values.NewString("hello"),
			operandValue:   values.NewString(" world"),
			expectedResult: "hello world",
		},
		{
			name:           "float addition",
			initialValue:   values.NewFloat(3.14),
			operandValue:   values.NewFloat(2.86),
			expectedResult: 6.0,
		},
		{
			name:           "mixed int and float addition",
			initialValue:   values.NewInt(10),
			operandValue:   values.NewFloat(5.5),
			expectedResult: 15.5,
		},
		{
			name:           "null property initialization",
			initialValue:   nil, // Property doesn't exist
			operandValue:   values.NewInt(42),
			expectedResult: int64(42), // null + 42 = 42
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := NewVirtualMachine()
			ctx := NewExecutionContext()

			className := "TestClass"
			propertyName := "compoundProp"

			// Setup class and property if initial value is provided
			if tt.initialValue != nil {
				ctx.Classes[className] = &Class{
					Name:       className,
					Properties: make(map[string]*Property),
					Methods:    make(map[string]*Function),
					Constants:  make(map[string]*values.Value),
				}
				ctx.Classes[className].Properties[propertyName] = &Property{
					Name:         propertyName,
					DefaultValue: tt.initialValue,
					Visibility:   "public",
					IsStatic:     true,
				}
			}

			// Setup operands
			ctx.Temporaries[1] = values.NewString(className)
			ctx.Temporaries[2] = values.NewString(propertyName)
			ctx.Temporaries[3] = tt.operandValue

			// Create ASSIGN_STATIC_PROP_OP instruction
			inst := &opcodes.Instruction{
				Opcode: opcodes.OP_ASSIGN_STATIC_PROP_OP,
				Op1:    1, // Class name
				Op2:    2, // Property name
				Result: 3, // Operand value
			}
			inst.OpType1, inst.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR)

			err := vm.executeAssignStaticPropertyOp(ctx, inst)
			if err != nil {
				t.Fatalf("executeAssignStaticPropertyOp failed: %v", err)
			}

			// Check result
			prop := ctx.Classes[className].Properties[propertyName]
			if prop == nil {
				t.Fatal("Property should exist after operation")
			}

			result := prop.DefaultValue
			if result == nil {
				t.Fatal("Property value should not be nil")
			}

			// Check result value based on expected type
			switch expected := tt.expectedResult.(type) {
			case int64:
				if !result.IsInt() || result.ToInt() != expected {
					t.Errorf("Expected int %d, got %v", expected, result)
				}
			case string:
				if !result.IsString() || result.ToString() != expected {
					t.Errorf("Expected string %q, got %v", expected, result)
				}
			case float64:
				if !result.IsFloat() || result.ToFloat() != expected {
					t.Errorf("Expected float %f, got %v", expected, result)
				}
			default:
				t.Fatalf("Unknown expected result type: %T", expected)
			}
		})
	}
}

func TestAssignStaticPropertyErrors(t *testing.T) {
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	// Test with non-string class name
	ctx.Temporaries[1] = values.NewInt(123) // Invalid class name
	ctx.Temporaries[2] = values.NewString("prop")
	ctx.Temporaries[3] = values.NewString("value")

	inst := &opcodes.Instruction{
		Opcode: opcodes.OP_ASSIGN_STATIC_PROP,
		Op1:    1,
		Op2:    2,
		Result: 3,
	}
	inst.OpType1, inst.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR)

	err := vm.executeAssignStaticProperty(ctx, inst)
	if err == nil {
		t.Error("Expected error for non-string class name")
	}

	// Test with non-string property name
	ctx.Temporaries[1] = values.NewString("TestClass")
	ctx.Temporaries[2] = values.NewInt(456) // Invalid property name

	err = vm.executeAssignStaticProperty(ctx, inst)
	if err == nil {
		t.Error("Expected error for non-string property name")
	}
}
