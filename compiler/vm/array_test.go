package vm

import (
	"testing"

	"github.com/wudi/hey/compiler/opcodes"
	"github.com/wudi/hey/compiler/values"
)

func TestCountOpcode(t *testing.T) {
	tests := []struct {
		name     string
		input    *values.Value
		expected int64
	}{
		{"empty array count", createArray(), 0},
		{"single element array count", createArrayWithElements([]interface{}{int64(0)}, []*values.Value{values.NewString("hello")}), 1},
		{"multi element array count", createArrayWithElements([]interface{}{int64(0), int64(1), "key"}, []*values.Value{values.NewInt(1), values.NewInt(2), values.NewString("value")}), 3},
		{"string count (length)", values.NewString("hello"), 5},
		{"empty string count", values.NewString(""), 0},
		{"int count (not array/string)", values.NewInt(42), 0},
		{"null count", values.NewNull(), 0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vm := NewVirtualMachine()
			ctx := NewExecutionContext()

			ctx.Temporaries = make(map[uint32]*values.Value)
			ctx.Temporaries[0] = test.input

			// Create COUNT instruction
			op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)
			inst := opcodes.Instruction{
				Opcode:  opcodes.OP_COUNT,
				OpType1: op1Type,
				OpType2: op2Type,
				Op1:     0, // Input from temporary variable 0
				Result:  1, // Store in temporary variable 1
			}

			// Execute COUNT
			err := vm.executeCount(ctx, &inst)
			if err != nil {
				t.Fatalf("COUNT execution failed: %v", err)
			}

			// Check result
			result := ctx.Temporaries[1]
			if result == nil {
				t.Fatal("COUNT result is nil")
			}

			if result.Type != values.TypeInt {
				t.Errorf("Expected int type, got %v", result.Type)
			}

			if result.Data.(int64) != test.expected {
				t.Errorf("Expected count %v, got %v", test.expected, result.Data.(int64))
			}
		})
	}
}

func TestInArrayOpcode(t *testing.T) {
	// Create test array: [1, "hello", 3.14, true]
	testArray := createArrayWithElements(
		[]interface{}{int64(0), int64(1), int64(2), int64(3)},
		[]*values.Value{
			values.NewInt(1),
			values.NewString("hello"),
			values.NewFloat(3.14),
			values.NewBool(true),
		},
	)

	tests := []struct {
		name     string
		needle   *values.Value
		haystack *values.Value
		expected bool
	}{
		{"find int in array", values.NewInt(1), testArray, true},
		{"find string in array", values.NewString("hello"), testArray, true},
		{"find float in array", values.NewFloat(3.14), testArray, true},
		{"find bool in array", values.NewBool(true), testArray, true},
		{"find via loose comparison int", values.NewInt(99), testArray, true},            // 99 == true (loose)
		{"find via loose comparison string", values.NewString("world"), testArray, true}, // "world" == true (loose)
		{"not find int zero in array", values.NewInt(0), testArray, false},               // 0 doesn't match any element
		{"not find false in array", values.NewBool(false), testArray, false},             // false doesn't match any element
		{"not find empty string in array", values.NewString(""), testArray, false},       // "" doesn't match any element
		{"not find null in array", values.NewNull(), testArray, false},                   // null doesn't match any element
		{"search in non-array", values.NewInt(1), values.NewString("hello"), false},
		{"search in empty array", values.NewInt(1), createArray(), false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vm := NewVirtualMachine()
			ctx := NewExecutionContext()

			ctx.Temporaries = make(map[uint32]*values.Value)
			ctx.Temporaries[0] = test.needle
			ctx.Temporaries[1] = test.haystack

			// Create IN_ARRAY instruction
			op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR)
			inst := opcodes.Instruction{
				Opcode:  opcodes.OP_IN_ARRAY,
				OpType1: op1Type,
				OpType2: op2Type,
				Op1:     0, // Needle from temporary variable 0
				Op2:     1, // Haystack from temporary variable 1
				Result:  2, // Store in temporary variable 2
			}

			// Execute IN_ARRAY
			err := vm.executeInArray(ctx, &inst)
			if err != nil {
				t.Fatalf("IN_ARRAY execution failed: %v", err)
			}

			// Check result
			result := ctx.Temporaries[2]
			if result == nil {
				t.Fatal("IN_ARRAY result is nil")
			}

			if result.Type != values.TypeBool {
				t.Errorf("Expected bool type, got %v", result.Type)
			}

			if result.Data.(bool) != test.expected {
				t.Errorf("Expected %v, got %v", test.expected, result.Data.(bool))
			}
		})
	}
}

func TestArrayKeyExistsOpcode(t *testing.T) {
	// Create test array: [0 => "zero", "key1" => "value1", 5 => "five"]
	testArray := createArrayWithElements(
		[]interface{}{int64(0), "key1", int64(5)},
		[]*values.Value{
			values.NewString("zero"),
			values.NewString("value1"),
			values.NewString("five"),
		},
	)

	tests := []struct {
		name     string
		key      *values.Value
		array    *values.Value
		expected bool
	}{
		{"key exists - int key", values.NewInt(0), testArray, true},
		{"key exists - string key", values.NewString("key1"), testArray, true},
		{"key exists - large int key", values.NewInt(5), testArray, true},
		{"key doesn't exist - int", values.NewInt(1), testArray, false},
		{"key doesn't exist - string", values.NewString("key2"), testArray, false},
		{"search in non-array", values.NewInt(0), values.NewString("hello"), false},
		{"search in empty array", values.NewInt(0), createArray(), false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vm := NewVirtualMachine()
			ctx := NewExecutionContext()

			ctx.Temporaries = make(map[uint32]*values.Value)
			ctx.Temporaries[0] = test.key
			ctx.Temporaries[1] = test.array

			// Create ARRAY_KEY_EXISTS instruction
			op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR)
			inst := opcodes.Instruction{
				Opcode:  opcodes.OP_ARRAY_KEY_EXISTS,
				OpType1: op1Type,
				OpType2: op2Type,
				Op1:     0, // Key from temporary variable 0
				Op2:     1, // Array from temporary variable 1
				Result:  2, // Store in temporary variable 2
			}

			// Execute ARRAY_KEY_EXISTS
			err := vm.executeArrayKeyExists(ctx, &inst)
			if err != nil {
				t.Fatalf("ARRAY_KEY_EXISTS execution failed: %v", err)
			}

			// Check result
			result := ctx.Temporaries[2]
			if result == nil {
				t.Fatal("ARRAY_KEY_EXISTS result is nil")
			}

			if result.Type != values.TypeBool {
				t.Errorf("Expected bool type, got %v", result.Type)
			}

			if result.Data.(bool) != test.expected {
				t.Errorf("Expected %v, got %v", test.expected, result.Data.(bool))
			}
		})
	}
}

func TestArrayValuesOpcode(t *testing.T) {
	// Create test array: [0 => "zero", "key1" => "value1", 5 => "five"]
	testArray := createArrayWithElements(
		[]interface{}{int64(0), "key1", int64(5)},
		[]*values.Value{
			values.NewString("zero"),
			values.NewString("value1"),
			values.NewString("five"),
		},
	)

	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	ctx.Temporaries = make(map[uint32]*values.Value)
	ctx.Temporaries[0] = testArray

	// Create ARRAY_VALUES instruction
	op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)
	inst := opcodes.Instruction{
		Opcode:  opcodes.OP_ARRAY_VALUES,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     0, // Array from temporary variable 0
		Result:  1, // Store in temporary variable 1
	}

	// Execute ARRAY_VALUES
	err := vm.executeArrayValues(ctx, &inst)
	if err != nil {
		t.Fatalf("ARRAY_VALUES execution failed: %v", err)
	}

	// Check result
	result := ctx.Temporaries[1]
	if result == nil {
		t.Fatal("ARRAY_VALUES result is nil")
	}

	if result.Type != values.TypeArray {
		t.Errorf("Expected array type, got %v", result.Type)
	}

	// Should have 3 elements with sequential numeric keys
	if result.ArrayCount() != 3 {
		t.Errorf("Expected array with 3 elements, got %d", result.ArrayCount())
	}

	// Check that values are present (order may vary due to map iteration)
	elem0 := result.ArrayGet(values.NewInt(0))
	elem1 := result.ArrayGet(values.NewInt(1))
	elem2 := result.ArrayGet(values.NewInt(2))

	if elem0 == nil || elem1 == nil || elem2 == nil {
		t.Error("Expected all elements to be present")
	}
}

func TestArrayKeysOpcode(t *testing.T) {
	// Create test array: [0 => "zero", "key1" => "value1", 5 => "five"]
	testArray := createArrayWithElements(
		[]interface{}{int64(0), "key1", int64(5)},
		[]*values.Value{
			values.NewString("zero"),
			values.NewString("value1"),
			values.NewString("five"),
		},
	)

	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	ctx.Temporaries = make(map[uint32]*values.Value)
	ctx.Temporaries[0] = testArray

	// Create ARRAY_KEYS instruction
	op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)
	inst := opcodes.Instruction{
		Opcode:  opcodes.OP_ARRAY_KEYS,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     0, // Array from temporary variable 0
		Result:  1, // Store in temporary variable 1
	}

	// Execute ARRAY_KEYS
	err := vm.executeArrayKeys(ctx, &inst)
	if err != nil {
		t.Fatalf("ARRAY_KEYS execution failed: %v", err)
	}

	// Check result
	result := ctx.Temporaries[1]
	if result == nil {
		t.Fatal("ARRAY_KEYS result is nil")
	}

	if result.Type != values.TypeArray {
		t.Errorf("Expected array type, got %v", result.Type)
	}

	// Should have 3 elements with sequential numeric keys
	if result.ArrayCount() != 3 {
		t.Errorf("Expected array with 3 elements, got %d", result.ArrayCount())
	}

	// Check that keys are present (order may vary due to map iteration)
	elem0 := result.ArrayGet(values.NewInt(0))
	elem1 := result.ArrayGet(values.NewInt(1))
	elem2 := result.ArrayGet(values.NewInt(2))

	if elem0 == nil || elem1 == nil || elem2 == nil {
		t.Error("Expected all key elements to be present")
	}
}

func TestArrayMergeOpcode(t *testing.T) {
	// Create first array: [0 => "a", 1 => "b"]
	array1 := createArrayWithElements(
		[]interface{}{int64(0), int64(1)},
		[]*values.Value{values.NewString("a"), values.NewString("b")},
	)

	// Create second array: [2 => "c", "key" => "d"]
	array2 := createArrayWithElements(
		[]interface{}{int64(2), "key"},
		[]*values.Value{values.NewString("c"), values.NewString("d")},
	)

	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	ctx.Temporaries = make(map[uint32]*values.Value)
	ctx.Temporaries[0] = array1
	ctx.Temporaries[1] = array2

	// Create ARRAY_MERGE instruction
	op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR)
	inst := opcodes.Instruction{
		Opcode:  opcodes.OP_ARRAY_MERGE,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     0, // First array from temporary variable 0
		Op2:     1, // Second array from temporary variable 1
		Result:  2, // Store in temporary variable 2
	}

	// Execute ARRAY_MERGE
	err := vm.executeArrayMerge(ctx, &inst)
	if err != nil {
		t.Fatalf("ARRAY_MERGE execution failed: %v", err)
	}

	// Check result
	result := ctx.Temporaries[2]
	if result == nil {
		t.Fatal("ARRAY_MERGE result is nil")
	}

	if result.Type != values.TypeArray {
		t.Errorf("Expected array type, got %v", result.Type)
	}

	// Should have 4 elements total
	if result.ArrayCount() != 4 {
		t.Errorf("Expected merged array with 4 elements, got %d", result.ArrayCount())
	}

	// Check that merged elements are accessible
	valA := result.ArrayGet(values.NewInt(0))
	valB := result.ArrayGet(values.NewInt(1))
	valC := result.ArrayGet(values.NewInt(2))
	valD := result.ArrayGet(values.NewString("key"))

	if valA == nil || valA.Data.(string) != "a" {
		t.Error("Expected merged array to contain 'a' at index 0")
	}
	if valB == nil || valB.Data.(string) != "b" {
		t.Error("Expected merged array to contain 'b' at index 1")
	}
	if valC == nil || valC.Data.(string) != "c" {
		t.Error("Expected merged array to contain 'c' at index 2")
	}
	if valD == nil || valD.Data.(string) != "d" {
		t.Error("Expected merged array to contain 'd' at key 'key'")
	}
}

// Test comprehensive array operations simulation
func TestArrayOperationsSimulation(t *testing.T) {
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	// Simulate: count(array_merge($arr1, $arr2)) where $arr1 has 2 elements, $arr2 has 1 element
	// Expected result: 3

	// Create arrays
	arr1 := createArrayWithElements(
		[]interface{}{int64(0), int64(1)},
		[]*values.Value{values.NewString("first"), values.NewString("second")},
	)
	arr2 := createArrayWithElements(
		[]interface{}{int64(0)},
		[]*values.Value{values.NewString("third")},
	)

	ctx.Temporaries = make(map[uint32]*values.Value)
	ctx.Temporaries[0] = arr1
	ctx.Temporaries[1] = arr2

	op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR)
	op1TypeSingle, op2TypeSingle := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)

	// 1. Merge arrays: array_merge($arr1, $arr2)
	mergeInst := opcodes.Instruction{
		Opcode:  opcodes.OP_ARRAY_MERGE,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     0, // arr1
		Op2:     1, // arr2
		Result:  2, // Store merged result in temp var 2
	}

	err := vm.executeArrayMerge(ctx, &mergeInst)
	if err != nil {
		t.Fatalf("ARRAY_MERGE execution failed: %v", err)
	}

	mergedArray := ctx.Temporaries[2]
	if mergedArray.ArrayCount() != 3 {
		t.Errorf("Expected merged array with 3 elements, got %d", mergedArray.ArrayCount())
	}

	// 2. Count merged array: count(merged_array)
	countInst := opcodes.Instruction{
		Opcode:  opcodes.OP_COUNT,
		OpType1: op1TypeSingle,
		OpType2: op2TypeSingle,
		Op1:     2, // Merged array from step 1
		Result:  3, // Store count in temp var 3
	}

	err = vm.executeCount(ctx, &countInst)
	if err != nil {
		t.Fatalf("COUNT execution failed: %v", err)
	}

	count := ctx.Temporaries[3]
	if !count.IsInt() || count.Data.(int64) != 3 {
		t.Errorf("Expected count 3, got %v", count.Data)
	}

	// 3. Test in_array: in_array("second", merged_array)
	needle := values.NewString("second")
	ctx.Temporaries[4] = needle

	inArrayInst := opcodes.Instruction{
		Opcode:  opcodes.OP_IN_ARRAY,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     4, // "second" from temp var 4
		Op2:     2, // merged array from step 1
		Result:  5, // Store result in temp var 5
	}

	err = vm.executeInArray(ctx, &inArrayInst)
	if err != nil {
		t.Fatalf("IN_ARRAY execution failed: %v", err)
	}

	inArrayResult := ctx.Temporaries[5]
	if !inArrayResult.IsBool() || !inArrayResult.Data.(bool) {
		t.Errorf("Expected in_array to return true, got %v", inArrayResult.Data)
	}

	// 4. Test array_key_exists: array_key_exists(1, merged_array)
	key := values.NewInt(1)
	ctx.Temporaries[6] = key

	keyExistsInst := opcodes.Instruction{
		Opcode:  opcodes.OP_ARRAY_KEY_EXISTS,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     6, // key 1 from temp var 6
		Op2:     2, // merged array from step 1
		Result:  7, // Store result in temp var 7
	}

	err = vm.executeArrayKeyExists(ctx, &keyExistsInst)
	if err != nil {
		t.Fatalf("ARRAY_KEY_EXISTS execution failed: %v", err)
	}

	keyExistsResult := ctx.Temporaries[7]
	if !keyExistsResult.IsBool() || !keyExistsResult.Data.(bool) {
		t.Errorf("Expected array_key_exists to return true, got %v", keyExistsResult.Data)
	}
}

// Helper function to create an empty array
func createArray() *values.Value {
	return values.NewArray()
}

// Helper function to create an array with elements
func createArrayWithElements(keys []interface{}, vals []*values.Value) *values.Value {
	array := values.NewArray()
	for i, key := range keys {
		if i < len(vals) {
			var keyValue *values.Value
			switch k := key.(type) {
			case int64:
				keyValue = values.NewInt(k)
			case string:
				keyValue = values.NewString(k)
			default:
				keyValue = values.NewInt(0)
			}
			array.ArraySet(keyValue, vals[i])
		}
	}
	return array
}
