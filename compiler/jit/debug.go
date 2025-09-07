package jit

import (
	"fmt"
	"strings"
	"time"
)

// JITDebugger JIT调试器
type JITDebugger struct {
	enabled       bool
	traceLevel    DebugLevel
	executionLogs []ExecutionLog
	breakpoints   map[uintptr]bool
}

// DebugLevel 调试级别
type DebugLevel int

const (
	DebugLevelNone DebugLevel = iota
	DebugLevelError
	DebugLevelWarn
	DebugLevelInfo
	DebugLevelDebug
	DebugLevelTrace
)

// ExecutionLog 执行日志记录
type ExecutionLog struct {
	Timestamp     time.Time
	FunctionName  string
	EntryPoint    uintptr
	Arguments     []int64
	Result        int64
	ExecutionTime time.Duration
	Success       bool
	Error         error
}

// NewJITDebugger 创建JIT调试器
func NewJITDebugger() *JITDebugger {
	return &JITDebugger{
		enabled:     false,
		traceLevel:  DebugLevelInfo,
		breakpoints: make(map[uintptr]bool),
	}
}

// Enable 启用调试
func (d *JITDebugger) Enable() {
	d.enabled = true
}

// Disable 禁用调试
func (d *JITDebugger) Disable() {
	d.enabled = false
}

// SetTraceLevel 设置跟踪级别
func (d *JITDebugger) SetTraceLevel(level DebugLevel) {
	d.traceLevel = level
}

// LogExecution 记录执行信息
func (d *JITDebugger) LogExecution(functionName string, entryPoint uintptr, args []int64, result int64, execTime time.Duration, err error) {
	if !d.enabled {
		return
	}

	log := ExecutionLog{
		Timestamp:     time.Now(),
		FunctionName:  functionName,
		EntryPoint:    entryPoint,
		Arguments:     make([]int64, len(args)),
		Result:        result,
		ExecutionTime: execTime,
		Success:       err == nil,
		Error:         err,
	}
	copy(log.Arguments, args)

	d.executionLogs = append(d.executionLogs, log)

	if d.traceLevel >= DebugLevelInfo {
		d.printExecutionLog(log)
	}
}

// printExecutionLog 打印执行日志
func (d *JITDebugger) printExecutionLog(log ExecutionLog) {
	status := "✅"
	if !log.Success {
		status = "❌"
	}

	argsStr := make([]string, len(log.Arguments))
	for i, arg := range log.Arguments {
		argsStr[i] = fmt.Sprintf("%d", arg)
	}

	fmt.Printf("[JIT-DEBUG] %s %s(%s) -> %d [%v]\n",
		status,
		log.FunctionName,
		strings.Join(argsStr, ", "),
		log.Result,
		log.ExecutionTime)

	if !log.Success && log.Error != nil {
		fmt.Printf("[JIT-DEBUG] Error: %v\n", log.Error)
	}
}

// AddBreakpoint 添加断点
func (d *JITDebugger) AddBreakpoint(entryPoint uintptr) {
	d.breakpoints[entryPoint] = true
}

// RemoveBreakpoint 移除断点
func (d *JITDebugger) RemoveBreakpoint(entryPoint uintptr) {
	delete(d.breakpoints, entryPoint)
}

// ShouldBreak 检查是否应该中断
func (d *JITDebugger) ShouldBreak(entryPoint uintptr) bool {
	if !d.enabled {
		return false
	}
	return d.breakpoints[entryPoint]
}

// DumpMachineCode 转储机器码
func (d *JITDebugger) DumpMachineCode(functionName string, code []byte, entryPoint uintptr) {
	if !d.enabled || d.traceLevel < DebugLevelDebug {
		return
	}

	fmt.Printf("[JIT-DEBUG] Machine code for %s (entry: 0x%x, size: %d bytes):\n",
		functionName, entryPoint, len(code))

	for i := 0; i < len(code); i += 16 {
		end := i + 16
		if end > len(code) {
			end = len(code)
		}

		// 地址
		fmt.Printf("0x%08x: ", entryPoint+uintptr(i))

		// 十六进制字节
		for j := i; j < end; j++ {
			fmt.Printf("%02x ", code[j])
		}

		// 填充空格
		for j := end; j < i+16; j++ {
			fmt.Print("   ")
		}

		// ASCII表示
		fmt.Print(" |")
		for j := i; j < end; j++ {
			if code[j] >= 32 && code[j] <= 126 {
				fmt.Printf("%c", code[j])
			} else {
				fmt.Print(".")
			}
		}
		fmt.Println("|")
	}
	fmt.Println()
}

// GetExecutionLogs 获取执行日志
func (d *JITDebugger) GetExecutionLogs() []ExecutionLog {
	return d.executionLogs
}

// ClearLogs 清空日志
func (d *JITDebugger) ClearLogs() {
	d.executionLogs = nil
}

// GetStats 获取调试统计信息
func (d *JITDebugger) GetStats() DebugStats {
	stats := DebugStats{
		TotalExecutions:      int64(len(d.executionLogs)),
		SuccessfulExecutions: 0,
		FailedExecutions:     0,
		TotalExecutionTime:   0,
		AverageExecutionTime: 0,
	}

	for _, log := range d.executionLogs {
		stats.TotalExecutionTime += log.ExecutionTime
		if log.Success {
			stats.SuccessfulExecutions++
		} else {
			stats.FailedExecutions++
		}
	}

	if stats.TotalExecutions > 0 {
		stats.AverageExecutionTime = stats.TotalExecutionTime / time.Duration(stats.TotalExecutions)
	}

	return stats
}

// DebugStats 调试统计信息
type DebugStats struct {
	TotalExecutions      int64
	SuccessfulExecutions int64
	FailedExecutions     int64
	TotalExecutionTime   time.Duration
	AverageExecutionTime time.Duration
}

// PrintStats 打印统计信息
func (d *JITDebugger) PrintStats() {
	stats := d.GetStats()

	fmt.Println("[JIT-DEBUG] Execution Statistics:")
	fmt.Printf("  Total executions: %d\n", stats.TotalExecutions)
	fmt.Printf("  Successful: %d (%.1f%%)\n", stats.SuccessfulExecutions,
		float64(stats.SuccessfulExecutions)/float64(stats.TotalExecutions)*100)
	fmt.Printf("  Failed: %d (%.1f%%)\n", stats.FailedExecutions,
		float64(stats.FailedExecutions)/float64(stats.TotalExecutions)*100)
	fmt.Printf("  Total time: %v\n", stats.TotalExecutionTime)
	fmt.Printf("  Average time: %v\n", stats.AverageExecutionTime)
}

// ValidateMachineCode 验证机器码
func (d *JITDebugger) ValidateMachineCode(code []byte) []ValidationIssue {
	var issues []ValidationIssue

	if len(code) == 0 {
		issues = append(issues, ValidationIssue{
			Level:   "ERROR",
			Message: "Machine code is empty",
		})
		return issues
	}

	// 检查基本的x86-64指令格式
	if len(code) < 3 {
		issues = append(issues, ValidationIssue{
			Level:   "WARN",
			Message: "Machine code is very short, might be incomplete",
		})
	}

	// 检查是否有适当的函数结尾（RET指令）
	hasRet := false
	for i := 0; i < len(code); i++ {
		if code[i] == 0xc3 { // RET
			hasRet = true
			break
		}
	}

	if !hasRet {
		issues = append(issues, ValidationIssue{
			Level:   "WARN",
			Message: "No RET instruction found, function might not return properly",
		})
	}

	// 检查是否有函数序言（push rbp; mov rbp, rsp）
	if len(code) >= 4 {
		if code[0] == 0x55 && code[1] == 0x48 && code[2] == 0x89 && code[3] == 0xe5 {
			// 找到标准函数序言
		} else {
			issues = append(issues, ValidationIssue{
				Level:   "INFO",
				Message: "Non-standard function prolog detected",
			})
		}
	}

	return issues
}

// ValidationIssue 验证问题
type ValidationIssue struct {
	Level   string
	Message string
}

// DisassembleMachineCode 反汇编机器码（简单版本）
func (d *JITDebugger) DisassembleMachineCode(code []byte, entryPoint uintptr) {
	if !d.enabled || d.traceLevel < DebugLevelTrace {
		return
	}

	fmt.Printf("[JIT-DEBUG] Disassembly for entry point 0x%x:\n", entryPoint)

	i := 0
	for i < len(code) {
		addr := entryPoint + uintptr(i)
		fmt.Printf("0x%08x: ", addr)

		// 简单的指令识别
		if i < len(code) {
			switch code[i] {
			case 0x55:
				fmt.Println("55              push rbp")
				i++
			case 0x48:
				if i+3 < len(code) && code[i+1] == 0x89 && code[i+2] == 0xe5 {
					fmt.Printf("48 89 e5        mov rbp, rsp\n")
					i += 3
				} else if i+3 < len(code) && code[i+1] == 0x83 && code[i+2] == 0xec {
					fmt.Printf("48 83 ec %02x     sub rsp, 0x%x\n", code[i+3], code[i+3])
					i += 4
				} else if i+2 < len(code) && code[i+1] == 0x01 {
					fmt.Printf("48 01 %02x        add %%r%s, %%r%s\n", code[i+2],
						getRegisterName(int((code[i+2]>>3)&7)), getRegisterName(int(code[i+2]&7)))
					i += 3
				} else if i+2 < len(code) && code[i+1] == 0x29 {
					fmt.Printf("48 29 %02x        sub %%r%s, %%r%s\n", code[i+2],
						getRegisterName(int((code[i+2]>>3)&7)), getRegisterName(int(code[i+2]&7)))
					i += 3
				} else if i+2 < len(code) && code[i+1] == 0x89 {
					fmt.Printf("48 89 %02x        mov %%r%s, %%r%s\n", code[i+2],
						getRegisterName(int((code[i+2]>>3)&7)), getRegisterName(int(code[i+2]&7)))
					i += 3
				} else {
					fmt.Printf("48              (REX.W prefix)\n")
					i++
				}
			case 0x5d:
				fmt.Println("5d              pop rbp")
				i++
			case 0xc3:
				fmt.Println("c3              ret")
				i++
			case 0x90:
				fmt.Println("90              nop")
				i++
			default:
				fmt.Printf("%02x              (unknown)\n", code[i])
				i++
			}
		} else {
			break
		}
	}
	fmt.Println()
}

// getRegisterName 获取寄存器名称
func getRegisterName(regCode int) string {
	switch regCode {
	case 0:
		return "ax"
	case 1:
		return "cx"
	case 2:
		return "dx"
	case 3:
		return "bx"
	case 4:
		return "sp"
	case 5:
		return "bp"
	case 6:
		return "si"
	case 7:
		return "di"
	default:
		return fmt.Sprintf("r%d", regCode)
	}
}

// MemoryProfiler 内存分析器
type MemoryProfiler struct {
	allocations    []MemoryAllocation
	totalAllocated int64
	totalFreed     int64
	currentUsage   int64
}

// MemoryAllocation 内存分配记录
type MemoryAllocation struct {
	Address      uintptr
	Size         int64
	Timestamp    time.Time
	FunctionName string
	Freed        bool
	FreeTime     time.Time
}

// NewMemoryProfiler 创建内存分析器
func NewMemoryProfiler() *MemoryProfiler {
	return &MemoryProfiler{}
}

// RecordAllocation 记录内存分配
func (mp *MemoryProfiler) RecordAllocation(addr uintptr, size int64, funcName string) {
	mp.allocations = append(mp.allocations, MemoryAllocation{
		Address:      addr,
		Size:         size,
		Timestamp:    time.Now(),
		FunctionName: funcName,
		Freed:        false,
	})
	mp.totalAllocated += size
	mp.currentUsage += size
}

// RecordFree 记录内存释放
func (mp *MemoryProfiler) RecordFree(addr uintptr) {
	for i := range mp.allocations {
		if mp.allocations[i].Address == addr && !mp.allocations[i].Freed {
			mp.allocations[i].Freed = true
			mp.allocations[i].FreeTime = time.Now()
			mp.totalFreed += mp.allocations[i].Size
			mp.currentUsage -= mp.allocations[i].Size
			break
		}
	}
}

// GetMemoryStats 获取内存统计
func (mp *MemoryProfiler) GetMemoryStats() MemoryStats {
	return MemoryStats{
		TotalAllocated:  mp.totalAllocated,
		TotalFreed:      mp.totalFreed,
		CurrentUsage:    mp.currentUsage,
		AllocationCount: int64(len(mp.allocations)),
	}
}

// MemoryStats 内存统计信息
type MemoryStats struct {
	TotalAllocated  int64
	TotalFreed      int64
	CurrentUsage    int64
	AllocationCount int64
}

// PrintMemoryStats 打印内存统计
func (mp *MemoryProfiler) PrintMemoryStats() {
	stats := mp.GetMemoryStats()
	fmt.Println("[JIT-DEBUG] Memory Statistics:")
	fmt.Printf("  Total allocated: %d bytes\n", stats.TotalAllocated)
	fmt.Printf("  Total freed: %d bytes\n", stats.TotalFreed)
	fmt.Printf("  Current usage: %d bytes\n", stats.CurrentUsage)
	fmt.Printf("  Allocation count: %d\n", stats.AllocationCount)
}
