package vm

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/wudi/php-parser/compiler/opcodes"
	"github.com/wudi/php-parser/compiler/values"
)

// TestPerformanceMetrics tests the performance metrics tracking
func TestPerformanceMetrics(t *testing.T) {
	metrics := NewPerformanceMetrics()

	// Record some instructions
	metrics.RecordInstruction("OP_ADD")
	metrics.RecordInstruction("OP_ADD")
	metrics.RecordInstruction("OP_ECHO")
	metrics.RecordInstruction("OP_ADD")

	// Record function calls
	metrics.RecordFunctionCall("test_function")
	metrics.RecordFunctionCall("another_function")
	metrics.RecordFunctionCall("test_function")

	// Record memory allocations
	metrics.RecordMemoryAllocation(1024)
	metrics.RecordMemoryAllocation(2048)
	metrics.RecordMemoryDeallocation(512)

	// Verify metrics
	require.Equal(t, uint64(4), metrics.TotalInstructions)
	require.Equal(t, uint64(3), metrics.InstructionCounts["OP_ADD"])
	require.Equal(t, uint64(1), metrics.InstructionCounts["OP_ECHO"])
	require.Equal(t, uint64(2), metrics.FunctionCallCounts["test_function"])
	require.Equal(t, uint64(1), metrics.FunctionCallCounts["another_function"])
	require.Equal(t, uint64(2), metrics.MemoryAllocations)
	require.Equal(t, uint64(1), metrics.MemoryDeallocations)
	require.Equal(t, uint64(2560), metrics.CurrentMemoryUsage) // 1024 + 2048 - 512
	require.Equal(t, uint64(3072), metrics.PeakMemoryUsage)    // 1024 + 2048

	// Test report generation
	report := metrics.GetReport()
	require.Contains(t, report, "Total Instructions: 4")
	require.Contains(t, report, "OP_ADD: 3")
	require.Contains(t, report, "test_function: 2 calls")
}

// TestDebugger tests the debugging functionality
func TestDebugger(t *testing.T) {
	debugger := NewDebugger(DebugLevelDetailed, nil)

	// Test breakpoint management
	debugger.SetBreakpoint(100)
	debugger.SetBreakpoint(200)

	require.True(t, debugger.ShouldBreak(100))
	require.True(t, debugger.ShouldBreak(200))
	require.False(t, debugger.ShouldBreak(300))

	debugger.RemoveBreakpoint(100)
	require.False(t, debugger.ShouldBreak(100))
	require.True(t, debugger.ShouldBreak(200))

	// Test variable watching
	debugger.WatchVariable("$testVar")
	require.True(t, debugger.WatchVariables["$testVar"])

	// Test instruction tracing
	ctx := &ExecutionContext{
		Variables: make(map[uint32]*values.Value),
		SP:        5,
	}
	ctx.Variables[0] = values.NewInt(42)

	inst := &opcodes.Instruction{
		Opcode: opcodes.OP_ADD,
		Op1:    1,
		Op2:    2,
		Result: 3,
	}

	debugger.TraceInstruction(10, inst, ctx, time.Microsecond*100)

	require.Len(t, debugger.InstructionLog, 1)
	trace := debugger.InstructionLog[0]
	require.Equal(t, 10, trace.IP)
	require.Equal(t, opcodes.OP_ADD, trace.Instruction.Opcode)
	require.Equal(t, 5, trace.StackSize)
	require.Equal(t, time.Microsecond*100, trace.Duration)

	// Test function call tracing
	args := []*values.Value{values.NewString("arg1"), values.NewInt(42)}
	debugger.TraceFunctionCall("testFunction", args, 0)

	require.Len(t, debugger.CallStack, 1)
	callTrace := debugger.CallStack[0]
	require.Equal(t, "testFunction", callTrace.FunctionName)
	require.Len(t, callTrace.Arguments, 2)
	require.Equal(t, 0, callTrace.CallDepth)

	// Test function return tracing
	returnValue := values.NewString("result")
	debugger.TraceFunctionReturn("testFunction", returnValue, 0)

	// Get the updated call trace
	updatedCallTrace := debugger.CallStack[0]
	require.Equal(t, returnValue, updatedCallTrace.ReturnValue)
	require.True(t, updatedCallTrace.Duration > 0)

	// Test report generation
	report := debugger.GenerateReport()
	require.Contains(t, report, "Instructions traced: 1")
	require.Contains(t, report, "Function calls: 1")
}

// TestVMOptimizer tests the VM optimization features
func TestVMOptimizer(t *testing.T) {
	optimizer := NewVMOptimizer()

	// Record some hot spots
	optimizer.RecordHotSpot(100)
	optimizer.RecordHotSpot(200)
	optimizer.RecordHotSpot(100) // Duplicate
	optimizer.RecordHotSpot(100) // Another duplicate
	optimizer.RecordHotSpot(300)
	optimizer.RecordHotSpot(200) // Duplicate

	// Test hot spot detection
	require.True(t, optimizer.IsHotSpot(100, 2))  // 3 executions >= 2
	require.True(t, optimizer.IsHotSpot(200, 2))  // 2 executions >= 2
	require.False(t, optimizer.IsHotSpot(300, 2)) // 1 execution < 2

	// Test getting hot spots
	hotSpots := optimizer.GetHotSpots(2)
	require.Len(t, hotSpots, 2)

	// Should be sorted by count (descending)
	require.Equal(t, 100, hotSpots[0].IP)
	require.Equal(t, uint64(3), hotSpots[0].Count)
	require.Equal(t, 200, hotSpots[1].IP)
	require.Equal(t, uint64(2), hotSpots[1].Count)
}

// TestMemoryPool tests the memory pool functionality
func TestMemoryPool(t *testing.T) {
	pool := NewMemoryPool()

	// Test value pooling
	val1 := pool.GetValue()
	val2 := pool.GetValue()
	require.NotNil(t, val1)
	require.NotNil(t, val2)

	// Set some values
	val1.Type = values.TypeInt
	val1.Data = int64(42)
	val2.Type = values.TypeString
	val2.Data = "test"

	// Return to pool
	pool.PutValue(val1)
	pool.PutValue(val2)

	// Get again (should be recycled)
	val3 := pool.GetValue()
	require.NotNil(t, val3)
	// Value should be reset
	require.Equal(t, values.ValueType(0), val3.Type)
	require.Nil(t, val3.Data)

	// Test statistics
	allocs, deallocs := pool.GetStats()
	require.Equal(t, uint64(3), allocs)   // val1, val2, val3
	require.Equal(t, uint64(2), deallocs) // val1, val2 returned

	// Test execution context pooling
	ctx1 := pool.GetExecutionContext()
	ctx2 := pool.GetExecutionContext()
	require.NotNil(t, ctx1)
	require.NotNil(t, ctx2)

	// Add some data
	ctx1.Variables[100] = values.NewInt(123)
	ctx1.Temporaries[200] = values.NewString("temp")

	// Return to pool
	pool.PutExecutionContext(ctx1)

	// Get again
	ctx3 := pool.GetExecutionContext()
	require.NotNil(t, ctx3)
	// Maps should be cleared
	require.Len(t, ctx3.Variables, 0)
	require.Len(t, ctx3.Temporaries, 0)
}

// TestEnhancedVMIntegration tests the integration of enhanced features with the VM
func TestEnhancedVMIntegration(t *testing.T) {
	// Create VM with profiling enabled
	vm := NewVirtualMachineWithProfiling(DebugLevelBasic)

	require.True(t, vm.EnableProfiling)
	require.True(t, vm.Debugger.ProfilerEnabled)
	require.Equal(t, DebugLevelBasic, vm.Debugger.Level)
	require.NotNil(t, vm.Metrics)
	require.NotNil(t, vm.Optimizer)
	require.NotNil(t, vm.MemoryPool)

	// Test utility methods
	vm.SetDebugLevel(DebugLevelVerbose)
	require.Equal(t, DebugLevelVerbose, vm.Debugger.Level)

	vm.SetBreakpoint(42)
	require.True(t, vm.Debugger.BreakPoints[42])

	vm.WatchVariable("$testVar")
	require.True(t, vm.Debugger.WatchVariables["$testVar"])

	// Test advanced profiling enablement
	vm.EnableAdvancedProfiling()
	require.True(t, vm.EnableProfiling)
	require.True(t, vm.DebugMode)
	require.Equal(t, DebugLevelDetailed, vm.Debugger.Level)
	require.True(t, vm.Debugger.ProfilerEnabled)

	// Test reports (should not crash with empty data)
	perfReport := vm.GetPerformanceReport()
	require.Contains(t, perfReport, "VM Performance Report")

	debugReport := vm.GetDebugReport()
	require.Contains(t, debugReport, "VM Debugger Report")

	// Test memory stats
	allocs, deallocs := vm.GetMemoryStats()
	require.Equal(t, uint64(0), allocs)   // No allocations yet
	require.Equal(t, uint64(0), deallocs) // No deallocations yet

	// Test hot spots (should be empty initially)
	hotSpots := vm.GetHotSpots(10)
	require.Len(t, hotSpots, 0)
}

// TestProfileDataAnalysis tests analysis of profiling data
func TestProfileDataAnalysis(t *testing.T) {
	profileData := &ProfileData{
		FunctionProfiles: make(map[string]*FunctionProfile),
		InstructionTimes: make(map[string]time.Duration),
		HotPaths:         make([]HotPath, 0),
		MemoryProfile: &MemoryProfile{
			AllocationsPerType:   make(map[string]uint64),
			DeallocationsPerType: make(map[string]uint64),
			PeakUsagePerType:     make(map[string]uint64),
			LeakDetection:        make(map[string]uint64),
		},
	}

	// Add some function profiles
	profileData.FunctionProfiles["function1"] = &FunctionProfile{
		Name:        "function1",
		CallCount:   100,
		TotalTime:   time.Millisecond * 500,
		AverageTime: time.Microsecond * 5000, // 5ms
		MinTime:     time.Microsecond * 1000, // 1ms
		MaxTime:     time.Millisecond * 50,   // 50ms
	}

	profileData.FunctionProfiles["function2"] = &FunctionProfile{
		Name:        "function2",
		CallCount:   50,
		TotalTime:   time.Millisecond * 1000, // 1s
		AverageTime: time.Millisecond * 20,   // 20ms
		MinTime:     time.Millisecond * 5,    // 5ms
		MaxTime:     time.Millisecond * 100,  // 100ms
	}

	// Add instruction timings
	profileData.InstructionTimes["OP_ADD"] = time.Millisecond * 100
	profileData.InstructionTimes["OP_ECHO"] = time.Millisecond * 50
	profileData.InstructionTimes["OP_CALL"] = time.Millisecond * 300

	// Add memory profile data
	profileData.MemoryProfile.AllocationsPerType["Value"] = 1000
	profileData.MemoryProfile.AllocationsPerType["Array"] = 200
	profileData.MemoryProfile.DeallocationsPerType["Value"] = 950
	profileData.MemoryProfile.DeallocationsPerType["Array"] = 180
	profileData.MemoryProfile.PeakUsagePerType["Value"] = 50 * 1024  // 50KB
	profileData.MemoryProfile.PeakUsagePerType["Array"] = 100 * 1024 // 100KB
	profileData.MemoryProfile.LeakDetection["Value"] = 50            // 50 leaked values
	profileData.MemoryProfile.LeakDetection["Array"] = 20            // 20 leaked arrays

	// Verify the data
	require.Len(t, profileData.FunctionProfiles, 2)
	require.Len(t, profileData.InstructionTimes, 3)

	function1 := profileData.FunctionProfiles["function1"]
	require.Equal(t, uint64(100), function1.CallCount)
	require.Equal(t, time.Millisecond*500, function1.TotalTime)

	function2 := profileData.FunctionProfiles["function2"]
	require.Equal(t, uint64(50), function2.CallCount)
	require.Equal(t, time.Millisecond*1000, function2.TotalTime)

	require.Equal(t, time.Millisecond*100, profileData.InstructionTimes["OP_ADD"])
	require.Equal(t, time.Millisecond*300, profileData.InstructionTimes["OP_CALL"])

	// Test memory analysis
	require.Equal(t, uint64(1000), profileData.MemoryProfile.AllocationsPerType["Value"])
	require.Equal(t, uint64(950), profileData.MemoryProfile.DeallocationsPerType["Value"])
	require.Equal(t, uint64(50), profileData.MemoryProfile.LeakDetection["Value"])
}

// BenchmarkVMPerformance benchmarks VM execution with and without profiling
func BenchmarkVMPerformance(b *testing.B) {
	// Create a simple instruction sequence
	instructions := []opcodes.Instruction{
		{Opcode: opcodes.OP_QM_ASSIGN, Op1: 0, Op2: 0, Result: 100},
		{Opcode: opcodes.OP_QM_ASSIGN, Op1: 1, Op2: 0, Result: 101},
		{Opcode: opcodes.OP_ADD, Op1: 100, Op2: 101, Result: 102},
		{Opcode: opcodes.OP_ECHO, Op1: 102, Op2: 0, Result: 0},
	}

	constants := []*values.Value{
		values.NewInt(10),
		values.NewInt(20),
	}

	b.Run("WithoutProfiling", func(b *testing.B) {
		vm := NewVirtualMachine()
		for i := 0; i < b.N; i++ {
			ctx := NewExecutionContext()
			ctx.SetOutputWriter(&strings.Builder{}) // Discard output
			vm.Execute(ctx, instructions, constants, nil, nil)
		}
	})

	b.Run("WithProfiling", func(b *testing.B) {
		vm := NewVirtualMachineWithProfiling(DebugLevelNone)
		for i := 0; i < b.N; i++ {
			ctx := NewExecutionContext()
			ctx.SetOutputWriter(&strings.Builder{}) // Discard output
			vm.Execute(ctx, instructions, constants, nil, nil)
		}
	})

	b.Run("WithDetailedProfiling", func(b *testing.B) {
		vm := NewVirtualMachineWithProfiling(DebugLevelDetailed)
		for i := 0; i < b.N; i++ {
			ctx := NewExecutionContext()
			ctx.SetOutputWriter(&strings.Builder{}) // Discard output
			vm.Execute(ctx, instructions, constants, nil, nil)
		}
	})
}
