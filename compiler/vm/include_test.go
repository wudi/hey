package vm

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wudi/hey/compiler/opcodes"
	"github.com/wudi/hey/compiler/values"
)

func TestIncludeOpcode(t *testing.T) {
	// Create test files
	testDir := t.TempDir()
	testFile := filepath.Join(testDir, "test.php")
	testContent := []byte("<?php echo 'Hello from included file'; ?>")

	err := os.WriteFile(testFile, testContent, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name          string
		filepath      string
		expectSuccess bool
		expectResult  bool
	}{
		{"include existing file", testFile, true, true},
		{"include non-existent file", "/non/existent/file.php", false, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vm := NewVirtualMachine()
			ctx := NewExecutionContext()

			ctx.Temporaries = make(map[uint32]*values.Value)
			ctx.Temporaries[0] = values.NewString(test.filepath)

			op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)
			inst := opcodes.Instruction{
				Opcode:  opcodes.OP_INCLUDE,
				OpType1: op1Type,
				OpType2: op2Type,
				Op1:     0,
				Result:  1,
			}

			err := vm.executeInclude(ctx, &inst)
			if err != nil {
				t.Fatalf("INCLUDE execution failed: %v", err)
			}

			// Check result
			result := ctx.Temporaries[1]
			if result == nil {
				t.Fatal("INCLUDE result is nil")
			}

			if test.expectSuccess {
				// Should return file size (int)
				if !result.IsInt() {
					t.Errorf("Expected int result for successful include, got %v", result)
				}
				expectedSize := int64(len(testContent))
				if result.Data.(int64) != expectedSize {
					t.Errorf("Expected result %d, got %d", expectedSize, result.Data.(int64))
				}
			} else {
				// Should return false for failed include
				if !result.IsBool() || result.Data.(bool) != false {
					t.Errorf("Expected false for failed include, got %v", result)
				}
			}
		})
	}
}

func TestIncludeOnceOpcode(t *testing.T) {
	// Create test files
	testDir := t.TempDir()
	testFile := filepath.Join(testDir, "test_once.php")
	testContent := []byte("<?php echo 'Hello from included once file'; ?>")

	err := os.WriteFile(testFile, testContent, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	ctx.Temporaries = make(map[uint32]*values.Value)
	ctx.Temporaries[0] = values.NewString(testFile)

	op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)

	// First include_once - should succeed
	inst1 := opcodes.Instruction{
		Opcode:  opcodes.OP_INCLUDE_ONCE,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     0,
		Result:  1,
	}

	err = vm.executeIncludeOnce(ctx, &inst1)
	if err != nil {
		t.Fatalf("First INCLUDE_ONCE execution failed: %v", err)
	}

	result1 := ctx.Temporaries[1]
	if result1 == nil || !result1.IsInt() {
		t.Fatalf("First INCLUDE_ONCE should return int, got %v", result1)
	}

	// Second include_once - should return true (already included)
	inst2 := opcodes.Instruction{
		Opcode:  opcodes.OP_INCLUDE_ONCE,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     0,
		Result:  2,
	}

	err = vm.executeIncludeOnce(ctx, &inst2)
	if err != nil {
		t.Fatalf("Second INCLUDE_ONCE execution failed: %v", err)
	}

	result2 := ctx.Temporaries[2]
	if result2 == nil || !result2.IsBool() || result2.Data.(bool) != true {
		t.Errorf("Second INCLUDE_ONCE should return true, got %v", result2)
	}
}

func TestRequireOpcode(t *testing.T) {
	// Create test files
	testDir := t.TempDir()
	testFile := filepath.Join(testDir, "test_require.php")
	testContent := []byte("<?php echo 'Hello from required file'; ?>")

	err := os.WriteFile(testFile, testContent, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name          string
		filepath      string
		expectSuccess bool
		expectError   bool
	}{
		{"require existing file", testFile, true, false},
		{"require non-existent file", "/non/existent/require.php", false, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vm := NewVirtualMachine()
			ctx := NewExecutionContext()

			ctx.Temporaries = make(map[uint32]*values.Value)
			ctx.Temporaries[0] = values.NewString(test.filepath)

			op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)
			inst := opcodes.Instruction{
				Opcode:  opcodes.OP_REQUIRE,
				OpType1: op1Type,
				OpType2: op2Type,
				Op1:     0,
				Result:  1,
			}

			err := vm.executeRequire(ctx, &inst)

			if test.expectError {
				if err == nil {
					t.Fatal("Expected REQUIRE to fail, but it succeeded")
				}
				// Error expected, test passed
				return
			}

			if err != nil {
				t.Fatalf("REQUIRE execution failed: %v", err)
			}

			// Check result
			result := ctx.Temporaries[1]
			if result == nil {
				t.Fatal("REQUIRE result is nil")
			}

			if test.expectSuccess {
				// Should return file size (int)
				if !result.IsInt() {
					t.Errorf("Expected int result for successful require, got %v", result)
				}
				expectedSize := int64(len(testContent))
				if result.Data.(int64) != expectedSize {
					t.Errorf("Expected result %d, got %d", expectedSize, result.Data.(int64))
				}
			}
		})
	}
}

func TestRequireOnceOpcode(t *testing.T) {
	// Create test files
	testDir := t.TempDir()
	testFile := filepath.Join(testDir, "test_require_once.php")
	testContent := []byte("<?php echo 'Hello from required once file'; ?>")

	err := os.WriteFile(testFile, testContent, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	ctx.Temporaries = make(map[uint32]*values.Value)
	ctx.Temporaries[0] = values.NewString(testFile)

	op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)

	// First require_once - should succeed
	inst1 := opcodes.Instruction{
		Opcode:  opcodes.OP_REQUIRE_ONCE,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     0,
		Result:  1,
	}

	err = vm.executeRequireOnce(ctx, &inst1)
	if err != nil {
		t.Fatalf("First REQUIRE_ONCE execution failed: %v", err)
	}

	result1 := ctx.Temporaries[1]
	if result1 == nil || !result1.IsInt() {
		t.Fatalf("First REQUIRE_ONCE should return int, got %v", result1)
	}

	// Second require_once - should return true (already included)
	inst2 := opcodes.Instruction{
		Opcode:  opcodes.OP_REQUIRE_ONCE,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     0,
		Result:  2,
	}

	err = vm.executeRequireOnce(ctx, &inst2)
	if err != nil {
		t.Fatalf("Second REQUIRE_ONCE execution failed: %v", err)
	}

	result2 := ctx.Temporaries[2]
	if result2 == nil || !result2.IsBool() || result2.Data.(bool) != true {
		t.Errorf("Second REQUIRE_ONCE should return true, got %v", result2)
	}
}

func TestRequireOnceFailure(t *testing.T) {
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	ctx.Temporaries = make(map[uint32]*values.Value)
	ctx.Temporaries[0] = values.NewString("/non/existent/require_once.php")

	op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)
	inst := opcodes.Instruction{
		Opcode:  opcodes.OP_REQUIRE_ONCE,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     0,
		Result:  1,
	}

	err := vm.executeRequireOnce(ctx, &inst)
	if err == nil {
		t.Fatal("Expected REQUIRE_ONCE to fail for non-existent file")
	}

	// Should get proper error message
	if !strings.Contains(err.Error(), "require(") {
		t.Errorf("Expected require error message, got: %s", err.Error())
	}
}

func TestIncludeFilePath(t *testing.T) {
	// Test absolute path handling for once variants
	testDir := t.TempDir()
	testFile := filepath.Join(testDir, "path_test.php")
	testContent := []byte("<?php // test ?>")

	err := os.WriteFile(testFile, testContent, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	// Test with relative path (should be converted to absolute)
	relPath := filepath.Base(testFile)

	// Change to test directory so relative path works
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(testDir)

	ctx.Temporaries = make(map[uint32]*values.Value)
	ctx.Temporaries[0] = values.NewString(relPath)

	op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)

	// First include_once with relative path
	inst1 := opcodes.Instruction{
		Opcode:  opcodes.OP_INCLUDE_ONCE,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     0,
		Result:  1,
	}

	err = vm.executeIncludeOnce(ctx, &inst1)
	if err != nil {
		t.Fatalf("First INCLUDE_ONCE with relative path failed: %v", err)
	}

	// Now try with absolute path - should still recognize as already included
	ctx.Temporaries[0] = values.NewString(testFile)
	inst2 := opcodes.Instruction{
		Opcode:  opcodes.OP_INCLUDE_ONCE,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     0,
		Result:  2,
	}

	err = vm.executeIncludeOnce(ctx, &inst2)
	if err != nil {
		t.Fatalf("Second INCLUDE_ONCE with absolute path failed: %v", err)
	}

	result2 := ctx.Temporaries[2]
	if result2 == nil || !result2.IsBool() || result2.Data.(bool) != true {
		t.Errorf("Second INCLUDE_ONCE with absolute path should return true (already included), got %v", result2)
	}
}
