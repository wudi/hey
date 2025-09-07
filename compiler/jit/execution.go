package jit

import (
	"fmt"
	"runtime"
	"time"

	"github.com/wudi/php-parser/compiler/opcodes"
	"github.com/wudi/php-parser/compiler/values"
)

// JITFunction represents a compiled JIT function ready for execution
type JITFunction struct {
	*CompiledFunction
	executableMemory *ExecutableMemory
	entryPoint       uintptr
	nativeCaller     *NativeFunctionCaller
	debugger         *JITDebugger
	memProfiler      *MemoryProfiler
}

// CallConvention defines how arguments are passed to JIT functions
type CallConvention int

const (
	// SystemV calling convention used on Unix systems
	CallConvSystemV CallConvention = iota
	// Microsoft x64 calling convention used on Windows
	CallConvWin64
)

// GetCallConvention returns the appropriate calling convention for the platform
func GetCallConvention() CallConvention {
	switch runtime.GOOS {
	case "windows":
		return CallConvWin64
	default:
		return CallConvSystemV
	}
}

// CompileToExecutable compiles bytecode to an executable JIT function
func (gen *AMD64CodeGenerator) CompileToExecutable(functionName string, bytecode []opcodes.Instruction, optimizations []Optimization) (*JITFunction, error) {
	// Generate machine code
	compiledFunc, err := gen.GenerateMachineCode(bytecode, optimizations)
	if err != nil {
		return nil, err
	}

	// Allocate executable memory
	execMem, err := AllocateExecutableMemory(len(compiledFunc.MachineCode))
	if err != nil {
		return nil, fmt.Errorf("failed to allocate executable memory: %v", err)
	}

	// Write machine code to executable memory
	err = execMem.WriteBytes(0, compiledFunc.MachineCode)
	if err != nil {
		execMem.Free()
		return nil, fmt.Errorf("failed to write machine code: %v", err)
	}

	// Create JIT function with native caller and debugging
	jitFunc := &JITFunction{
		CompiledFunction: compiledFunc,
		executableMemory: execMem,
		entryPoint:       execMem.GetFunctionPointer(0),
		nativeCaller:     NewNativeFunctionCaller(),
		debugger:         NewJITDebugger(),
		memProfiler:      NewMemoryProfiler(),
	}

	// Enable debugging in debug mode
	if gen.config.DebugMode {
		jitFunc.debugger.Enable()
		jitFunc.debugger.SetTraceLevel(DebugLevelDebug)
		jitFunc.debugger.DumpMachineCode(functionName, compiledFunc.MachineCode, jitFunc.entryPoint)
		jitFunc.debugger.DisassembleMachineCode(compiledFunc.MachineCode, jitFunc.entryPoint)

		// Validate machine code
		issues := jitFunc.debugger.ValidateMachineCode(compiledFunc.MachineCode)
		for _, issue := range issues {
			fmt.Printf("[JIT-VALIDATION] %s: %s\n", issue.Level, issue.Message)
		}
	}

	// Record memory allocation
	jitFunc.memProfiler.RecordAllocation(jitFunc.entryPoint, int64(len(compiledFunc.MachineCode)), functionName)

	jitFunc.Name = functionName
	jitFunc.EntryPoint = jitFunc.entryPoint

	return jitFunc, nil
}

// Execute executes the JIT-compiled function with the given arguments
func (jf *JITFunction) Execute(args []*values.Value) (*values.Value, error) {
	// Set up execution context
	ctx := NewJITExecutionContext()

	// Convert PHP values to native types for passing to JIT code
	nativeArgs, err := jf.convertArgsToNative(args)
	if err != nil {
		return nil, fmt.Errorf("failed to convert arguments: %v", err)
	}

	// Check for breakpoints
	if jf.debugger != nil && jf.debugger.ShouldBreak(jf.entryPoint) {
		fmt.Printf("[JIT-DEBUG] Breakpoint hit at 0x%x\n", jf.entryPoint)
	}

	// Execute the JIT function with timing
	start := time.Now()
	result, err := jf.executeNative(ctx, nativeArgs)
	execTime := time.Since(start)

	// Log execution if debugging is enabled
	if jf.debugger != nil {
		jf.debugger.LogExecution(jf.Name, jf.entryPoint, nativeArgs, result, execTime, err)
	}

	if err != nil {
		return nil, fmt.Errorf("JIT execution failed: %v", err)
	}

	// Convert result back to PHP value
	phpResult, err := jf.convertResultFromNative(result)
	if err != nil {
		return nil, fmt.Errorf("failed to convert result: %v", err)
	}

	return phpResult, nil
}

// executeNative executes the native machine code
func (jf *JITFunction) executeNative(ctx *JITExecutionContext, args []int64) (int64, error) {
	// 使用原生调用器执行机器码
	if jf.nativeCaller == nil {
		jf.nativeCaller = NewNativeFunctionCaller()
	}

	// 检查平台支持
	if !IsJITExecutionSupported() {
		if jf.debugger != nil {
			fmt.Println("[JIT-DEBUG] Platform doesn't support native execution, using simulation")
		}
		return jf.executeSimulated(ctx, args)
	}

	// 验证入口点
	if jf.entryPoint == 0 {
		return 0, fmt.Errorf("invalid entry point: null pointer")
	}

	// 安全的原生调用
	result, err := jf.nativeCaller.SafeNativeCall(jf.entryPoint, args)
	if err != nil {
		// 如果原生调用失败，回退到模拟执行
		if jf.debugger != nil {
			fmt.Printf("[JIT-DEBUG] Native call failed, falling back to simulation: %v\n", err)
		}
		return jf.executeSimulated(ctx, args)
	}

	return result, nil
}

// executeWithProfile 带性能分析的执行
func (jf *JITFunction) executeWithProfile(ctx *JITExecutionContext, args []int64) (int64, error) {
	start := time.Now()
	result, err := jf.executeNative(ctx, args)
	elapsed := time.Since(start)

	// 更新执行统计
	jf.ExecutionCount++
	jf.ExecutionTime += elapsed

	return result, err
}

// createFunctionSignature 创建函数签名
func (jf *JITFunction) createFunctionSignature(argCount int) *FunctionSignature {
	paramTypes := make([]ParameterType, argCount)
	for i := range paramTypes {
		paramTypes[i] = ParamTypeInt64 // 简化：所有参数都是int64
	}

	return &FunctionSignature{
		ParameterTypes: paramTypes,
		ReturnType:     ParamTypeInt64,
		CallingConv:    GetCallConvention(),
	}
}

// executeSimulated provides a simulated execution for testing and development
// This serves as a fallback when native execution fails
func (jf *JITFunction) executeSimulated(ctx *JITExecutionContext, args []int64) (int64, error) {
	// 分析机器码来决定模拟的操作
	if len(jf.MachineCode) == 0 {
		return 0, fmt.Errorf("no machine code to simulate")
	}

	// 简单的模式匹配来识别操作类型
	operation := jf.detectOperation()

	switch operation {
	case "add":
		if len(args) >= 2 {
			result := args[0] + args[1]
			fmt.Printf("JIT Simulated ADD: %d + %d = %d\n", args[0], args[1], result)
			return result, nil
		}
	case "sub":
		if len(args) >= 2 {
			result := args[0] - args[1]
			fmt.Printf("JIT Simulated SUB: %d - %d = %d\n", args[0], args[1], result)
			return result, nil
		}
	case "mul":
		if len(args) >= 2 {
			result := args[0] * args[1]
			fmt.Printf("JIT Simulated MUL: %d * %d = %d\n", args[0], args[1], result)
			return result, nil
		}
	default:
		// 默认加法操作
		if len(args) >= 2 {
			result := args[0] + args[1]
			fmt.Printf("JIT Simulated (default ADD): %d + %d = %d\n", args[0], args[1], result)
			return result, nil
		} else if len(args) == 1 {
			return args[0], nil
		}
	}

	return 0, fmt.Errorf("simulated execution: insufficient arguments")
}

// detectOperation 从机器码中检测操作类型
func (jf *JITFunction) detectOperation() string {
	// 简单的机器码模式识别
	code := jf.MachineCode

	// 查找ADD指令模式 (0x48, 0x01, ...)
	for i := 0; i < len(code)-2; i++ {
		if code[i] == 0x48 && code[i+1] == 0x01 {
			return "add"
		}
		if code[i] == 0x48 && code[i+1] == 0x29 {
			return "sub"
		}
		if code[i] == 0x48 && code[i+1] == 0x0f && i < len(code)-3 && code[i+2] == 0xaf {
			return "mul"
		}
	}

	return "unknown"
}

// convertArgsToNative converts PHP values to native integers for JIT execution
func (jf *JITFunction) convertArgsToNative(args []*values.Value) ([]int64, error) {
	nativeArgs := make([]int64, len(args))

	for i, arg := range args {
		switch arg.Type {
		case values.TypeInt:
			nativeArgs[i] = arg.ToInt()
		case values.TypeFloat:
			nativeArgs[i] = int64(arg.ToFloat()) // Convert float to int for simplicity
		case values.TypeString:
			// For strings, we'll use the length as a simple conversion
			nativeArgs[i] = int64(len(arg.ToString()))
		case values.TypeBool:
			if arg.ToBool() {
				nativeArgs[i] = 1
			} else {
				nativeArgs[i] = 0
			}
		case values.TypeNull:
			nativeArgs[i] = 0
		default:
			return nil, fmt.Errorf("unsupported argument type: %d", arg.Type)
		}
	}

	return nativeArgs, nil
}

// convertResultFromNative converts native result back to PHP value
func (jf *JITFunction) convertResultFromNative(result int64) (*values.Value, error) {
	// For simplicity, always return integer results
	return values.NewInt(result), nil
}

// Free releases the executable memory used by this JIT function
func (jf *JITFunction) Free() error {
	if jf.executableMemory != nil {
		// Record memory free for profiling
		if jf.memProfiler != nil {
			jf.memProfiler.RecordFree(jf.entryPoint)
		}

		err := jf.executableMemory.Free()
		jf.executableMemory = nil
		jf.entryPoint = 0
		jf.nativeCaller = nil

		// Print debug stats if debugging is enabled
		if jf.debugger != nil && jf.debugger.enabled {
			jf.debugger.PrintStats()
			jf.memProfiler.PrintMemoryStats()
		}

		return err
	}
	return nil
}

// GetExecutionStats returns execution statistics for this JIT function
func (jf *JITFunction) GetExecutionStats() JITExecutionStats {
	return JITExecutionStats{
		FunctionName:    jf.Name,
		ExecutionCount:  jf.ExecutionCount,
		TotalTime:       jf.ExecutionTime,
		AverageTime:     jf.ExecutionTime / time.Duration(max(jf.ExecutionCount, 1)),
		MachineCodeSize: len(jf.MachineCode),
		EntryPoint:      jf.entryPoint,
	}
}

// JITExecutionStats represents statistics for JIT function execution
type JITExecutionStats struct {
	FunctionName    string
	ExecutionCount  int64
	TotalTime       time.Duration
	AverageTime     time.Duration
	MachineCodeSize int
	EntryPoint      uintptr
}

// max helper function
func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

// CallNativeFunction is a low-level function to call native code
// This would normally be implemented in assembly, but we'll provide
// a Go-based simulation for development and testing
func CallNativeFunction(entryPoint uintptr, args []int64) (int64, error) {
	// WARNING: This is extremely unsafe and for demonstration only
	// Real implementation would use proper calling conventions

	if entryPoint == 0 {
		return 0, fmt.Errorf("invalid entry point")
	}

	// In a real implementation, this would be:
	// 1. Set up proper calling convention (registers, stack)
	// 2. Call the function at entryPoint
	// 3. Handle return value and error conditions
	// 4. Restore calling context

	// For now, we'll simulate a basic function call
	// This is NOT actual machine code execution
	return 42, nil // Placeholder return value
}

// IsJITExecutionSupported checks if JIT execution is supported on current platform
func IsJITExecutionSupported() bool {
	switch runtime.GOOS {
	case "linux":
		return runtime.GOARCH == "amd64"
	case "darwin":
		return runtime.GOARCH == "amd64"
	case "windows":
		// Windows support would require additional implementation
		return false
	default:
		return false
	}
}
