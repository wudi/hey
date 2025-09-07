package vm

import (
	"testing"

	"github.com/wudi/php-parser/compiler/opcodes"
	"github.com/wudi/php-parser/compiler/values"
)

func TestForeachFreeOpcode(t *testing.T) {
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	// Setup a foreach iterator in slot 5
	iteratorSlot := uint32(5)

	// Create a mock iterator
	ctx.ForeachIterators[iteratorSlot] = &ForeachIterator{
		Array:   values.NewArray(),
		Index:   0,
		Keys:    []*values.Value{values.NewString("key1"), values.NewString("key2")},
		Values:  []*values.Value{values.NewString("value1"), values.NewString("value2")},
		HasMore: true,
	}

	// Also setup some temporary variables that might be associated
	ctx.Temporaries[iteratorSlot] = values.NewString("iterator_value")
	ctx.Temporaries[iteratorSlot+1] = values.NewString("iterator_key")

	// Verify setup
	if ctx.ForeachIterators[iteratorSlot] == nil {
		t.Fatal("Iterator should be set up")
	}
	if ctx.Temporaries[iteratorSlot] == nil {
		t.Fatal("Temporary value should be set up")
	}

	// Create FE_FREE instruction
	inst := &opcodes.Instruction{
		Opcode: opcodes.OP_FE_FREE,
		Op1:    iteratorSlot, // Iterator slot to free
		Op2:    0,            // Unused
		Result: 0,            // Unused
	}
	inst.OpType1, inst.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_UNUSED, opcodes.IS_UNUSED, opcodes.IS_UNUSED)

	err := vm.executeForeachFree(ctx, inst)
	if err != nil {
		t.Fatalf("executeForeachFree failed: %v", err)
	}

	// Check that iterator was removed
	if ctx.ForeachIterators[iteratorSlot] != nil {
		t.Error("Iterator should have been freed")
	}

	// Check that associated temporaries were cleaned up
	if ctx.Temporaries[iteratorSlot] != nil {
		t.Error("Iterator temporary value should have been cleaned up")
	}
	if ctx.Temporaries[iteratorSlot+1] != nil {
		t.Error("Iterator temporary key should have been cleaned up")
	}
}

func TestForeachFreeWithEmptyContext(t *testing.T) {
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	// Don't set up any iterators or temporaries
	iteratorSlot := uint32(10)

	// Create FE_FREE instruction
	inst := &opcodes.Instruction{
		Opcode: opcodes.OP_FE_FREE,
		Op1:    iteratorSlot,
		Op2:    0,
		Result: 0,
	}
	inst.OpType1, inst.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_UNUSED, opcodes.IS_UNUSED, opcodes.IS_UNUSED)

	// This should not crash even if there's nothing to free
	err := vm.executeForeachFree(ctx, inst)
	if err != nil {
		t.Fatalf("executeForeachFree should not fail with empty context: %v", err)
	}
}

func TestEvalOpcode(t *testing.T) {
	tests := []struct {
		name         string
		evalCode     string
		expectedNull bool
		shouldError  bool
	}{
		{
			name:         "eval empty string",
			evalCode:     "",
			expectedNull: true,
			shouldError:  false,
		},
		{
			name:         "eval simple PHP code",
			evalCode:     "echo 'Hello World';",
			expectedNull: true, // Our stub implementation returns NULL
			shouldError:  false,
		},
		{
			name:         "eval return statement",
			evalCode:     "return 42;",
			expectedNull: true, // Our stub implementation returns NULL
			shouldError:  false,
		},
		{
			name:         "eval variable assignment",
			evalCode:     "$x = 10; return $x;",
			expectedNull: true, // Our stub implementation returns NULL
			shouldError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := NewVirtualMachine()
			ctx := NewExecutionContext()

			// Setup the code to eval
			ctx.Temporaries[1] = values.NewString(tt.evalCode)

			// Create EVAL instruction
			inst := &opcodes.Instruction{
				Opcode: opcodes.OP_EVAL,
				Op1:    1, // Code to evaluate
				Op2:    0, // Unused
				Result: 2, // Result location
			}
			inst.OpType1, inst.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)

			err := vm.executeEval(ctx, inst)

			if tt.shouldError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("executeEval failed: %v", err)
			}

			// Check result
			result := ctx.Temporaries[2]
			if result == nil {
				t.Fatal("Result should not be nil")
			}

			if tt.expectedNull {
				if !result.IsNull() {
					t.Errorf("Expected null result, got %v", result)
				}
			}
		})
	}
}

func TestEvalOpcodeWithNonStringCode(t *testing.T) {
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	// Setup non-string value to eval
	ctx.Temporaries[1] = values.NewInt(123)

	// Create EVAL instruction
	inst := &opcodes.Instruction{
		Opcode: opcodes.OP_EVAL,
		Op1:    1,
		Op2:    0,
		Result: 2,
	}
	inst.OpType1, inst.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)

	err := vm.executeEval(ctx, inst)
	if err == nil {
		t.Error("Expected error for non-string eval code")
	}

	expectedError := "EVAL requires string code to evaluate"
	if err.Error() != expectedError {
		t.Errorf("Expected error %q, got %q", expectedError, err.Error())
	}
}

func TestEvalOpcodeWithNullCode(t *testing.T) {
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	// Setup null value to eval
	ctx.Temporaries[1] = values.NewNull()

	// Create EVAL instruction
	inst := &opcodes.Instruction{
		Opcode: opcodes.OP_EVAL,
		Op1:    1,
		Op2:    0,
		Result: 2,
	}
	inst.OpType1, inst.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)

	err := vm.executeEval(ctx, inst)
	if err == nil {
		t.Error("Expected error for null eval code")
	}
}

func TestForeachCleanupFlow(t *testing.T) {
	// This test simulates a typical foreach loop cleanup scenario
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	// Setup multiple iterators as would happen in nested foreach loops
	for i := uint32(0); i < 3; i++ {
		ctx.ForeachIterators[i] = &ForeachIterator{
			Array:   values.NewArray(),
			Index:   int(i),
			Keys:    []*values.Value{},
			Values:  []*values.Value{values.NewInt(int64(i))},
			HasMore: false,
		}
		ctx.Temporaries[i] = values.NewInt(int64(i * 10))
		ctx.Temporaries[i+10] = values.NewString("key_" + string(rune('0'+i)))
	}

	// Free iterators in reverse order (as would happen when exiting nested loops)
	for i := uint32(2); i >= 0 && i <= 2; i-- { // Handle underflow
		inst := &opcodes.Instruction{
			Opcode: opcodes.OP_FE_FREE,
			Op1:    i,
			Op2:    0,
			Result: 0,
		}
		inst.OpType1, inst.OpType2 = opcodes.EncodeOpTypes(opcodes.IS_UNUSED, opcodes.IS_UNUSED, opcodes.IS_UNUSED)

		err := vm.executeForeachFree(ctx, inst)
		if err != nil {
			t.Fatalf("Failed to free iterator %d: %v", i, err)
		}

		// Verify this iterator was freed
		if ctx.ForeachIterators[i] != nil {
			t.Errorf("Iterator %d should have been freed", i)
		}
	}

	// Verify all iterators were freed
	if len(ctx.ForeachIterators) > 0 {
		t.Errorf("Expected all iterators to be freed, but %d remain", len(ctx.ForeachIterators))
	}
}
