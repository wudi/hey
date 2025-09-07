package jit

import (
	"fmt"
	"time"

	"github.com/wudi/php-parser/compiler/values"
)

// SetDebugMode 设置调试模式
func (jf *JITFunction) SetDebugMode(enabled bool) {
	if jf.debugger == nil {
		jf.debugger = NewJITDebugger()
	}

	if enabled {
		jf.debugger.Enable()
		jf.debugger.SetTraceLevel(DebugLevelDebug)
	} else {
		jf.debugger.Disable()
	}
}

// AddBreakpoint 添加断点
func (jf *JITFunction) AddBreakpoint() {
	if jf.debugger != nil {
		jf.debugger.AddBreakpoint(jf.entryPoint)
	}
}

// RemoveBreakpoint 移除断点
func (jf *JITFunction) RemoveBreakpoint() {
	if jf.debugger != nil {
		jf.debugger.RemoveBreakpoint(jf.entryPoint)
	}
}

// GetDebugStats 获取调试统计
func (jf *JITFunction) GetDebugStats() DebugStats {
	if jf.debugger != nil {
		return jf.debugger.GetStats()
	}
	return DebugStats{}
}

// GetMemoryStats 获取内存统计
func (jf *JITFunction) GetMemoryStats() MemoryStats {
	if jf.memProfiler != nil {
		return jf.memProfiler.GetMemoryStats()
	}
	return MemoryStats{}
}

// GetPerformanceMetrics 获取性能指标
func (jf *JITFunction) GetPerformanceMetrics() JITPerformanceMetrics {
	debugStats := jf.GetDebugStats()
	memStats := jf.GetMemoryStats()
	execStats := jf.GetExecutionStats()

	totalExecutions := debugStats.TotalExecutions
	if totalExecutions == 0 {
		totalExecutions = 1 // 避免除零
	}

	return JITPerformanceMetrics{
		FunctionName:         jf.Name,
		ExecutionCount:       debugStats.TotalExecutions,
		SuccessRate:          float64(debugStats.SuccessfulExecutions) / float64(totalExecutions),
		AverageExecutionTime: debugStats.AverageExecutionTime,
		TotalExecutionTime:   debugStats.TotalExecutionTime,
		MachineCodeSize:      int64(execStats.MachineCodeSize),
		MemoryUsage:          memStats.CurrentUsage,
		OptimizationLevel:    int64(jf.OptimizationLevel),
	}
}

// JITPerformanceMetrics JIT性能指标
type JITPerformanceMetrics struct {
	FunctionName         string
	ExecutionCount       int64
	SuccessRate          float64
	AverageExecutionTime time.Duration
	TotalExecutionTime   time.Duration
	MachineCodeSize      int64
	MemoryUsage          int64
	OptimizationLevel    int64
}

// Validate 验证JIT函数的有效性
func (jf *JITFunction) Validate() error {
	if jf.entryPoint == 0 {
		return fmt.Errorf("invalid entry point")
	}

	if jf.executableMemory == nil {
		return fmt.Errorf("no executable memory allocated")
	}

	if len(jf.MachineCode) == 0 {
		return fmt.Errorf("no machine code generated")
	}

	// 验证机器码
	if jf.debugger != nil {
		issues := jf.debugger.ValidateMachineCode(jf.MachineCode)
		for _, issue := range issues {
			if issue.Level == "ERROR" {
				return fmt.Errorf("machine code validation error: %s", issue.Message)
			}
		}
	}

	return nil
}

// Clone 克隆JIT函数
func (jf *JITFunction) Clone(newName string) (*JITFunction, error) {
	// 分配新的可执行内存
	newExecMem, err := AllocateExecutableMemory(len(jf.MachineCode))
	if err != nil {
		return nil, fmt.Errorf("failed to allocate memory for clone: %v", err)
	}

	// 复制机器码
	err = newExecMem.WriteBytes(0, jf.MachineCode)
	if err != nil {
		newExecMem.Free()
		return nil, fmt.Errorf("failed to write machine code to clone: %v", err)
	}

	// 创建新的JIT函数
	clone := &JITFunction{
		CompiledFunction: &CompiledFunction{
			Name:              newName,
			MachineCode:       make([]byte, len(jf.MachineCode)),
			EntryPoint:        newExecMem.GetFunctionPointer(0),
			OptimizationLevel: jf.OptimizationLevel,
			OptimizationFlags: make([]string, len(jf.OptimizationFlags)),
		},
		executableMemory: newExecMem,
		entryPoint:       newExecMem.GetFunctionPointer(0),
		nativeCaller:     NewNativeFunctionCaller(),
		debugger:         NewJITDebugger(),
		memProfiler:      NewMemoryProfiler(),
	}

	// 复制数据
	copy(clone.MachineCode, jf.MachineCode)
	copy(clone.OptimizationFlags, jf.OptimizationFlags)

	// 记录内存分配
	clone.memProfiler.RecordAllocation(clone.entryPoint, int64(len(clone.MachineCode)), newName)

	return clone, nil
}

// ExecuteWithContext 使用上下文执行JIT函数
func (jf *JITFunction) ExecuteWithContext(ctx *JITExecutionContext, args []*values.Value) (*values.Value, error) {
	// Convert PHP values to native types
	nativeArgs, err := jf.convertArgsToNative(args)
	if err != nil {
		return nil, fmt.Errorf("failed to convert arguments: %v", err)
	}

	// Execute with the provided context
	result, err := jf.executeNative(ctx, nativeArgs)
	if err != nil {
		return nil, err
	}

	// Convert result back to PHP value
	return jf.convertResultFromNative(result)
}

// CreateTrampoline 为JIT函数创建跳转代码
func (jf *JITFunction) CreateTrampoline() (*ExecutableMemory, error) {
	if jf.nativeCaller == nil {
		jf.nativeCaller = NewNativeFunctionCaller()
	}

	return jf.nativeCaller.CreateFunctionTrampoline(jf.entryPoint)
}

// WarmUp 预热函数（执行几次来确保系统准备就绪）
func (jf *JITFunction) WarmUp() error {
	// 创建测试参数
	warmupArgs := []*values.Value{
		values.NewInt(1),
		values.NewInt(2),
	}

	// 执行几次预热
	for i := 0; i < 3; i++ {
		_, err := jf.Execute(warmupArgs)
		if err != nil {
			return fmt.Errorf("warmup execution %d failed: %v", i+1, err)
		}
	}

	return nil
}

// ExecuteWithTimeout 带超时的执行
func (jf *JITFunction) ExecuteWithTimeout(args []*values.Value, timeout time.Duration) (*values.Value, error) {
	resultChan := make(chan *values.Value, 1)
	errorChan := make(chan error, 1)

	// 在goroutine中执行
	go func() {
		result, err := jf.Execute(args)
		if err != nil {
			errorChan <- err
		} else {
			resultChan <- result
		}
	}()

	// 等待结果或超时
	select {
	case result := <-resultChan:
		return result, nil
	case err := <-errorChan:
		return nil, err
	case <-time.After(timeout):
		return nil, fmt.Errorf("JIT execution timed out after %v", timeout)
	}
}

// GetInstructionCount 获取指令数量估计
func (jf *JITFunction) GetInstructionCount() int {
	// 简单估计：平均每条x86指令3-4字节
	return len(jf.MachineCode) / 3
}

// PrintDebugInfo 打印调试信息
func (jf *JITFunction) PrintDebugInfo() {
	fmt.Printf("=== JIT Function Debug Info: %s ===\n", jf.Name)
	fmt.Printf("Entry Point: 0x%x\n", jf.entryPoint)
	fmt.Printf("Machine Code Size: %d bytes\n", len(jf.MachineCode))
	fmt.Printf("Estimated Instructions: %d\n", jf.GetInstructionCount())
	fmt.Printf("Optimization Level: %d\n", jf.OptimizationLevel)

	if jf.debugger != nil {
		fmt.Println("\nExecution Stats:")
		jf.debugger.PrintStats()

		fmt.Println("\nMemory Stats:")
		jf.memProfiler.PrintMemoryStats()

		fmt.Println("\nMachine Code:")
		jf.debugger.DumpMachineCode(jf.Name, jf.MachineCode, jf.entryPoint)
	}

	fmt.Println("=====================================")
}

// IsHealthy 检查函数健康状态
func (jf *JITFunction) IsHealthy() bool {
	err := jf.Validate()
	if err != nil {
		return false
	}

	// 检查执行成功率
	metrics := jf.GetPerformanceMetrics()
	return metrics.SuccessRate > 0.8 // 至少80%成功率
}
