package vm

import (
	"testing"

	"github.com/wudi/hey/compiler/opcodes"
	"github.com/wudi/hey/compiler/values"
)

func TestCloneOpcode(t *testing.T) {
	tests := []struct {
		name    string
		object  *values.Value
		wantErr bool
		errMsg  string
	}{
		{
			name: "clone simple object",
			object: &values.Value{
				Type: values.TypeObject,
				Data: values.Object{
					ClassName: "TestClass",
					Properties: map[string]*values.Value{
						"name": values.NewString("original"),
						"age":  values.NewInt(25),
					},
				},
			},
			wantErr: false,
		},
		{
			name:    "clone non-object (string)",
			object:  values.NewString("not an object"),
			wantErr: true,
			errMsg:  "__clone method called on non-object",
		},
		{
			name:    "clone non-object (int)",
			object:  values.NewInt(42),
			wantErr: true,
			errMsg:  "__clone method called on non-object",
		},
		{
			name:    "clone non-object (array)",
			object:  values.NewArray(),
			wantErr: true,
			errMsg:  "__clone method called on non-object",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vm := NewVirtualMachine()
			ctx := NewExecutionContext()

			ctx.Temporaries = make(map[uint32]*values.Value)
			ctx.Temporaries[0] = test.object

			op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)
			inst := opcodes.Instruction{
				Opcode:  opcodes.OP_CLONE,
				OpType1: op1Type,
				OpType2: op2Type,
				Op1:     0,
				Result:  1,
			}

			err := vm.executeClone(ctx, &inst)

			if test.wantErr {
				if err == nil {
					t.Fatal("Expected error, but got none")
				}
				if err.Error() != test.errMsg {
					t.Errorf("Expected error message '%s', got '%s'", test.errMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Check result
			result := ctx.Temporaries[1]
			if result == nil {
				t.Fatal("Clone result is nil")
			}

			if !result.IsObject() {
				t.Fatalf("Expected cloned result to be object, got %v", result.Type)
			}

			// Verify it's a different object instance
			if result == test.object {
				t.Error("Cloned object is the same instance as original (should be different)")
			}

			// Verify deep copy semantics
			originalObj := test.object.Data.(values.Object)
			clonedObj := result.Data.(values.Object)

			if clonedObj.ClassName != originalObj.ClassName {
				t.Errorf("Expected cloned class name '%s', got '%s'", originalObj.ClassName, clonedObj.ClassName)
			}

			// Check properties are copied
			if len(clonedObj.Properties) != len(originalObj.Properties) {
				t.Errorf("Expected %d properties in clone, got %d", len(originalObj.Properties), len(clonedObj.Properties))
			}

			for key, originalProp := range originalObj.Properties {
				clonedProp, exists := clonedObj.Properties[key]
				if !exists {
					t.Errorf("Property '%s' missing in cloned object", key)
					continue
				}

				// Properties should have same value but be different instances
				if clonedProp == originalProp {
					t.Errorf("Property '%s' is same instance in clone (should be copied)", key)
				}

				if clonedProp.Type != originalProp.Type {
					t.Errorf("Property '%s' type mismatch: expected %v, got %v", key, originalProp.Type, clonedProp.Type)
				}

				if clonedProp.Data != originalProp.Data {
					t.Errorf("Property '%s' value mismatch: expected %v, got %v", key, originalProp.Data, clonedProp.Data)
				}
			}
		})
	}
}

func TestCloneObjectWithNestedObjects(t *testing.T) {
	// Create nested object structure
	innerObject := &values.Value{
		Type: values.TypeObject,
		Data: values.Object{
			ClassName: "InnerClass",
			Properties: map[string]*values.Value{
				"value": values.NewString("inner value"),
			},
		},
	}

	outerObject := &values.Value{
		Type: values.TypeObject,
		Data: values.Object{
			ClassName: "OuterClass",
			Properties: map[string]*values.Value{
				"name":  values.NewString("outer"),
				"inner": innerObject,
			},
		},
	}

	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	ctx.Temporaries = make(map[uint32]*values.Value)
	ctx.Temporaries[0] = outerObject

	op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)
	inst := opcodes.Instruction{
		Opcode:  opcodes.OP_CLONE,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     0,
		Result:  1,
	}

	err := vm.executeClone(ctx, &inst)
	if err != nil {
		t.Fatalf("Clone failed: %v", err)
	}

	result := ctx.Temporaries[1]
	if result == nil || !result.IsObject() {
		t.Fatal("Clone result should be an object")
	}

	clonedObj := result.Data.(values.Object)
	originalObj := outerObject.Data.(values.Object)

	// Check outer object properties
	clonedInner := clonedObj.Properties["inner"]
	originalInner := originalObj.Properties["inner"]

	// Verify deep cloning - inner objects should be different instances
	if clonedInner == originalInner {
		t.Error("Inner object should be deep cloned (different instance)")
	}

	// But should have same values
	clonedInnerObj := clonedInner.Data.(values.Object)
	originalInnerObj := originalInner.Data.(values.Object)

	if clonedInnerObj.ClassName != originalInnerObj.ClassName {
		t.Error("Inner object class name should be preserved")
	}

	clonedInnerValue := clonedInnerObj.Properties["value"]
	originalInnerValue := originalInnerObj.Properties["value"]

	if clonedInnerValue.Data.(string) != originalInnerValue.Data.(string) {
		t.Error("Inner object property value should be preserved")
	}
}

func TestCloneObjectWithArrays(t *testing.T) {
	// Create object with array property
	array := values.NewArray()
	array.ArraySet(values.NewString("0"), values.NewString("item1"))
	array.ArraySet(values.NewString("1"), values.NewString("item2"))

	objectWithArray := &values.Value{
		Type: values.TypeObject,
		Data: values.Object{
			ClassName: "ArrayClass",
			Properties: map[string]*values.Value{
				"items": array,
				"count": values.NewInt(2),
			},
		},
	}

	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	ctx.Temporaries = make(map[uint32]*values.Value)
	ctx.Temporaries[0] = objectWithArray

	op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)
	inst := opcodes.Instruction{
		Opcode:  opcodes.OP_CLONE,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     0,
		Result:  1,
	}

	err := vm.executeClone(ctx, &inst)
	if err != nil {
		t.Fatalf("Clone failed: %v", err)
	}

	result := ctx.Temporaries[1]
	if result == nil || !result.IsObject() {
		t.Fatal("Clone result should be an object")
	}

	clonedObj := result.Data.(values.Object)
	originalObj := objectWithArray.Data.(values.Object)

	// Check array property is deep cloned
	clonedItems := clonedObj.Properties["items"]
	originalItems := originalObj.Properties["items"]

	if clonedItems == originalItems {
		t.Error("Array property should be deep cloned (different instance)")
	}

	// Verify array contents are preserved
	clonedArray := clonedItems.Data.(*values.Array)
	originalArray := originalItems.Data.(*values.Array)

	if len(clonedArray.Elements) != len(originalArray.Elements) {
		t.Error("Array elements count should be preserved")
	}

	for key, originalElement := range originalArray.Elements {
		clonedElement, exists := clonedArray.Elements[key]
		if !exists {
			t.Errorf("Array element '%s' missing in clone", key)
			continue
		}

		// Elements should be different instances but same values
		if clonedElement == originalElement {
			t.Errorf("Array element '%s' should be deep cloned", key)
		}

		if clonedElement.Data.(string) != originalElement.Data.(string) {
			t.Errorf("Array element '%s' value mismatch", key)
		}
	}
}

func TestCloneEmptyObject(t *testing.T) {
	emptyObject := &values.Value{
		Type: values.TypeObject,
		Data: values.Object{
			ClassName:  "EmptyClass",
			Properties: make(map[string]*values.Value),
		},
	}

	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	ctx.Temporaries = make(map[uint32]*values.Value)
	ctx.Temporaries[0] = emptyObject

	op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)
	inst := opcodes.Instruction{
		Opcode:  opcodes.OP_CLONE,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     0,
		Result:  1,
	}

	err := vm.executeClone(ctx, &inst)
	if err != nil {
		t.Fatalf("Clone of empty object failed: %v", err)
	}

	result := ctx.Temporaries[1]
	if result == nil || !result.IsObject() {
		t.Fatal("Clone result should be an object")
	}

	clonedObj := result.Data.(values.Object)
	if clonedObj.ClassName != "EmptyClass" {
		t.Error("Cloned empty object should preserve class name")
	}

	if len(clonedObj.Properties) != 0 {
		t.Error("Cloned empty object should have no properties")
	}
}
