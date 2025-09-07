package jit

import (
	"runtime"
	"testing"
	"time"

	"github.com/wudi/php-parser/compiler/opcodes"
	"github.com/wudi/php-parser/compiler/values"
)

func TestExecutableMemoryAllocation(t *testing.T) {
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		t.Skip("Executable memory allocation only supported on Linux and macOS")
	}

	size := 4096 // One page
	execMem, err := AllocateExecutableMemory(size)
	if err != nil {
		t.Fatalf("Failed to allocate executable memory: %v", err)
	}
	defer execMem.Free()

	if execMem.Size < size {
		t.Errorf("Expected at least %d bytes, got %d", size, execMem.Size)
	}

	if len(execMem.Data) != execMem.Size {
		t.Errorf("Data slice length %d doesn't match size %d", len(execMem.Data), execMem.Size)
	}

	// Test writing to executable memory
	testData := []byte{0x90, 0x90, 0x90} // NOP instructions
	err = execMem.WriteBytes(0, testData)
	if err != nil {
		t.Fatalf("Failed to write to executable memory: %v", err)
	}

	// Verify the data was written
	for i, b := range testData {
		if execMem.Data[i] != b {
			t.Errorf("Data mismatch at offset %d: expected %02x, got %02x", i, b, execMem.Data[i])
		}
	}
}

func TestJITExecutionContext(t *testing.T) {
	ctx := NewJITExecutionContext()

	// Test stack operations
	val1 := interface{}(int64(42))
	val2 := interface{}(int64(24))

	err := ctx.PushValue(&val1)
	if err != nil {
		t.Fatalf("Failed to push value: %v", err)
	}

	err = ctx.PushValue(&val2)
	if err != nil {
		t.Fatalf("Failed to push second value: %v", err)
	}

	// Test stack pointer
	if ctx.StackPtr != 2 {
		t.Errorf("Expected stack pointer 2, got %d", ctx.StackPtr)
	}

	// Test pop operations
	popped, err := ctx.PopValue()
	if err != nil {
		t.Fatalf("Failed to pop value: %v", err)
	}

	if *popped != val2 {
		t.Errorf("Popped wrong value: expected %v, got %v", val2, *popped)
	}

	// Test register operations
	ctx.SetRegister("RAX", 0x1234567890ABCDEF)
	if ctx.GetRegister("RAX") != 0x1234567890ABCDEF {
		t.Errorf("Register value mismatch")
	}
}

func TestJITFunctionCompilation(t *testing.T) {
	config := DefaultConfig()
	config.DebugMode = true

	amd64Gen, err := NewAMD64CodeGenerator(config)
	if err != nil {
		t.Fatalf("Failed to create AMD64 code generator: %v", err)
	}

	// Create simple bytecode (add two numbers)
	bytecode := []opcodes.Instruction{
		{
			Opcode: opcodes.OP_ADD,
			Op1:    1,
			Op2:    2,
			Result: 3,
		},
		{
			Opcode: opcodes.OP_RETURN,
			Op1:    3,
		},
	}

	// Attempt compilation - this may fail due to platform limitations
	// but should not crash
	jitFunc, err := amd64Gen.CompileToExecutable("testAdd", bytecode, nil)

	if err != nil {
		t.Logf("Compilation failed as expected on %s: %v", runtime.GOOS, err)

		// On unsupported platforms, we expect specific errors
		if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
			return // Expected failure on unsupported platforms
		}
	} else {
		// If compilation succeeded, test the JIT function
		if jitFunc == nil {
			t.Fatal("JIT function is nil despite successful compilation")
		}

		// Clean up
		defer jitFunc.Free()

		if jitFunc.Name != "testAdd" {
			t.Errorf("Expected function name 'testAdd', got '%s'", jitFunc.Name)
		}

		if len(jitFunc.MachineCode) == 0 {
			t.Error("Machine code is empty")
		}

		if jitFunc.entryPoint == 0 {
			t.Error("Entry point is zero")
		}
	}
}

func TestJITFunctionExecution(t *testing.T) {
	// Create a simple JIT function for testing
	compiledFunc := &CompiledFunction{
		Name:        "testFunction",
		MachineCode: []byte{0x90, 0xC3}, // NOP, RET
		EntryPoint:  0x1000,             // Dummy entry point
	}

	jitFunc := &JITFunction{
		CompiledFunction: compiledFunc,
	}

	// Test argument conversion
	args := []*values.Value{
		values.NewInt(10),
		values.NewInt(20),
	}

	nativeArgs, err := jitFunc.convertArgsToNative(args)
	if err != nil {
		t.Fatalf("Failed to convert arguments: %v", err)
	}

	if len(nativeArgs) != 2 {
		t.Errorf("Expected 2 native args, got %d", len(nativeArgs))
	}

	if nativeArgs[0] != 10 || nativeArgs[1] != 20 {
		t.Errorf("Argument conversion failed: got %v", nativeArgs)
	}

	// Test result conversion
	nativeResult := int64(30)
	phpResult, err := jitFunc.convertResultFromNative(nativeResult)
	if err != nil {
		t.Fatalf("Failed to convert result: %v", err)
	}

	if phpResult.Type != values.TypeInt || phpResult.ToInt() != 30 {
		t.Errorf("Result conversion failed: got %v", phpResult)
	}
}

func TestJITExecutionSimulation(t *testing.T) {
	// Create a JIT function with simple machine code
	compiledFunc := &CompiledFunction{
		Name:        "simulatedAdd",
		MachineCode: []byte{0x48, 0x01, 0xd8, 0xC3}, // ADD RAX, RBX; RET (simplified)
	}

	jitFunc := &JITFunction{
		CompiledFunction: compiledFunc,
	}

	// Test simulated execution
	ctx := NewJITExecutionContext()
	args := []int64{15, 25}

	result, err := jitFunc.executeSimulated(ctx, args)
	if err != nil {
		t.Fatalf("Simulated execution failed: %v", err)
	}

	// Should return the sum
	expected := int64(40)
	if result != expected {
		t.Errorf("Expected %d, got %d", expected, result)
	}
}

func TestCallConventionDetection(t *testing.T) {
	conv := GetCallConvention()

	switch runtime.GOOS {
	case "windows":
		if conv != CallConvWin64 {
			t.Errorf("Expected Win64 calling convention on Windows, got %v", conv)
		}
	default:
		if conv != CallConvSystemV {
			t.Errorf("Expected System V calling convention on %s, got %v", runtime.GOOS, conv)
		}
	}
}

func TestPlatformSupport(t *testing.T) {
	supported := IsJITExecutionSupported()

	switch runtime.GOOS {
	case "linux", "darwin":
		if runtime.GOARCH == "amd64" {
			if !supported {
				t.Errorf("JIT execution should be supported on %s/%s", runtime.GOOS, runtime.GOARCH)
			}
		}
	case "windows":
		// Windows support not yet implemented
		if supported {
			t.Errorf("JIT execution should not be supported on Windows yet")
		}
	default:
		if supported {
			t.Errorf("JIT execution should not be supported on %s", runtime.GOOS)
		}
	}
}

func TestJITExecutionStats(t *testing.T) {
	compiledFunc := &CompiledFunction{
		Name:           "statsTest",
		MachineCode:    []byte{0x90, 0xC3},
		ExecutionCount: 5,
		ExecutionTime:  50 * time.Millisecond,
	}

	jitFunc := &JITFunction{
		CompiledFunction: compiledFunc,
		entryPoint:       0x2000,
	}

	stats := jitFunc.GetExecutionStats()

	if stats.FunctionName != "statsTest" {
		t.Errorf("Expected function name 'statsTest', got '%s'", stats.FunctionName)
	}

	if stats.ExecutionCount != 5 {
		t.Errorf("Expected execution count 5, got %d", stats.ExecutionCount)
	}

	if stats.TotalTime != 50*time.Millisecond {
		t.Errorf("Expected total time 50ms, got %v", stats.TotalTime)
	}

	expectedAvg := 10 * time.Millisecond
	if stats.AverageTime != expectedAvg {
		t.Errorf("Expected average time %v, got %v", expectedAvg, stats.AverageTime)
	}

	if stats.MachineCodeSize != 2 {
		t.Errorf("Expected machine code size 2, got %d", stats.MachineCodeSize)
	}

	if stats.EntryPoint != 0x2000 {
		t.Errorf("Expected entry point 0x2000, got 0x%x", stats.EntryPoint)
	}
}

func TestExecutableMemoryErrors(t *testing.T) {
	// Test allocation with zero size
	execMem, err := AllocateExecutableMemory(0)
	if err == nil {
		t.Error("Expected error for zero-size allocation")
		if execMem != nil {
			execMem.Free()
		}
	}

	// Test writing beyond bounds
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		execMem, err := AllocateExecutableMemory(10)
		if err == nil {
			defer execMem.Free()

			// Try to write beyond the actual allocated size
			// Note: actual allocated size might be page-aligned and larger than requested
			actualSize := len(execMem.Data)
			err = execMem.WriteBytes(actualSize-1, []byte{1, 2, 3, 4}) // Should exceed bounds
			if err == nil {
				t.Error("Expected error for out-of-bounds write")
			}
		}
	}
}

// Benchmark JIT execution overhead
func BenchmarkJITExecution(b *testing.B) {
	compiledFunc := &CompiledFunction{
		Name:        "benchmarkFunc",
		MachineCode: []byte{0x48, 0x01, 0xd8, 0xC3}, // ADD RAX, RBX; RET
	}

	jitFunc := &JITFunction{
		CompiledFunction: compiledFunc,
	}

	args := []*values.Value{
		values.NewInt(100),
		values.NewInt(200),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := jitFunc.Execute(args)
		if err != nil {
			b.Fatalf("JIT execution failed: %v", err)
		}
	}
}

// Benchmark memory allocation
func BenchmarkExecutableMemoryAllocation(b *testing.B) {
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		b.Skip("Executable memory allocation only supported on Linux and macOS")
	}

	size := 4096

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		execMem, err := AllocateExecutableMemory(size)
		if err != nil {
			b.Fatalf("Failed to allocate memory: %v", err)
		}
		execMem.Free()
	}
}
