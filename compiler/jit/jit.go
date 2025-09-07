package jit

import (
	"fmt"
	"sync"
	"time"

	"github.com/wudi/php-parser/compiler/opcodes"
)

// JITCompiler 代表即时编译器
type JITCompiler struct {
	// 配置
	config *Config

	// 热点检测
	hotspotDetector *HotspotDetector

	// 代码生成器
	codeGenerator CodeGenerator

	// 编译后的代码缓存
	compiledCode sync.Map // map[string]*CompiledFunction

	// 统计信息
	stats *CompilerStats

	// 互斥锁保护编译过程
	mu sync.RWMutex
}

// Config JIT编译器配置
type Config struct {
	// 编译阈值：函数被调用多少次后开始JIT编译
	CompilationThreshold int

	// 最大编译函数数量
	MaxCompiledFunctions int

	// 是否启用优化
	EnableOptimizations bool

	// 目标架构
	TargetArch string // "amd64", "arm64"

	// 调试模式
	DebugMode bool
}

// CompiledFunction 代表一个JIT编译后的函数
type CompiledFunction struct {
	// 函数名称
	Name string

	// 原始字节码
	ByteCode []opcodes.Instruction

	// 编译后的机器码
	MachineCode []byte

	// 入口点地址
	EntryPoint uintptr

	// 编译时间
	CompileTime time.Time

	// 执行统计
	ExecutionCount int64
	ExecutionTime  time.Duration

	// 优化信息
	OptimizationLevel int
	OptimizationFlags []string
}

// CodeGenerator 代码生成器接口
type CodeGenerator interface {
	// 生成机器码
	GenerateMachineCode(bytecode []opcodes.Instruction, optimizations []Optimization) (*CompiledFunction, error)

	// 获取目标架构
	GetTargetArch() string

	// 是否支持指定的opcode
	SupportsOpcode(opcode opcodes.Opcode) bool
}

// Optimization 优化接口
type Optimization interface {
	// 优化名称
	Name() string

	// 应用优化
	Apply(bytecode []opcodes.Instruction) ([]opcodes.Instruction, error)

	// 是否适用于指定的字节码
	IsApplicable(bytecode []opcodes.Instruction) bool
}

// CompilerStats JIT编译器统计信息
type CompilerStats struct {
	// 编译统计
	TotalCompilations      int64
	SuccessfulCompilations int64
	FailedCompilations     int64

	// 性能统计
	TotalCompileTime   time.Duration
	AverageCompileTime time.Duration

	// 执行统计
	TotalJITExecutions    int64
	TotalJITExecutionTime time.Duration

	// 内存使用
	CompiledCodeSize int64
	MaxCodeCacheSize int64

	// 互斥锁保护统计数据
	mu sync.RWMutex
}

// NewJITCompiler 创建新的JIT编译器
func NewJITCompiler(config *Config) (*JITCompiler, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// 创建热点检测器
	hotspotDetector := NewHotspotDetector(config.CompilationThreshold)

	// 根据目标架构创建代码生成器
	var codeGenerator CodeGenerator
	var err error

	switch config.TargetArch {
	case "amd64":
		codeGenerator, err = NewAMD64CodeGenerator(config)
	case "arm64":
		return nil, fmt.Errorf("ARM64 code generation not yet implemented")
	default:
		return nil, fmt.Errorf("unsupported target architecture: %s", config.TargetArch)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create code generator: %v", err)
	}

	return &JITCompiler{
		config:          config,
		hotspotDetector: hotspotDetector,
		codeGenerator:   codeGenerator,
		stats:           &CompilerStats{},
	}, nil
}

// DefaultConfig 返回默认的JIT编译器配置
func DefaultConfig() *Config {
	return &Config{
		CompilationThreshold: 10,   // 调用10次后开始JIT编译
		MaxCompiledFunctions: 1000, // 最多编译1000个函数
		EnableOptimizations:  true,
		TargetArch:           "amd64",
		DebugMode:            false,
	}
}

// ShouldCompile 检查函数是否应该被JIT编译
func (jit *JITCompiler) ShouldCompile(functionName string) bool {
	// 检查是否已经编译过
	if _, exists := jit.compiledCode.Load(functionName); exists {
		return false
	}

	// 检查热点检测器
	return jit.hotspotDetector.IsHotspot(functionName)
}

// CompileFunction 编译指定的函数
func (jit *JITCompiler) CompileFunction(functionName string, bytecode []opcodes.Instruction) (*CompiledFunction, error) {
	jit.mu.Lock()
	defer jit.mu.Unlock()

	// 再次检查是否已经编译（双重检查锁定）
	if compiled, exists := jit.compiledCode.Load(functionName); exists {
		return compiled.(*CompiledFunction), nil
	}

	// 更新统计
	jit.stats.mu.Lock()
	jit.stats.TotalCompilations++
	jit.stats.mu.Unlock()

	startTime := time.Now()

	// 应用优化（如果启用）
	optimizedBytecode := bytecode
	var appliedOptimizations []Optimization

	if jit.config.EnableOptimizations {
		optimizedBytecode, appliedOptimizations = jit.applyOptimizations(bytecode)
	}

	// 生成机器码
	compiledFunc, err := jit.codeGenerator.GenerateMachineCode(optimizedBytecode, appliedOptimizations)
	if err != nil {
		jit.stats.mu.Lock()
		jit.stats.FailedCompilations++
		jit.stats.mu.Unlock()
		return nil, fmt.Errorf("failed to generate machine code for %s: %v", functionName, err)
	}

	compiledFunc.Name = functionName
	compiledFunc.ByteCode = bytecode
	compiledFunc.CompileTime = time.Now()

	// 缓存编译结果
	jit.compiledCode.Store(functionName, compiledFunc)

	// 更新统计
	compileTime := time.Since(startTime)
	jit.stats.mu.Lock()
	jit.stats.SuccessfulCompilations++
	jit.stats.TotalCompileTime += compileTime
	jit.stats.AverageCompileTime = jit.stats.TotalCompileTime / time.Duration(jit.stats.SuccessfulCompilations)
	jit.stats.CompiledCodeSize += int64(len(compiledFunc.MachineCode))
	jit.stats.mu.Unlock()

	if jit.config.DebugMode {
		fmt.Printf("JIT: Compiled function %s in %v, machine code size: %d bytes\n",
			functionName, compileTime, len(compiledFunc.MachineCode))
	}

	return compiledFunc, nil
}

// GetCompiledFunction 获取已编译的函数
func (jit *JITCompiler) GetCompiledFunction(functionName string) (*CompiledFunction, bool) {
	if compiled, exists := jit.compiledCode.Load(functionName); exists {
		return compiled.(*CompiledFunction), true
	}
	return nil, false
}

// RecordFunctionCall 记录函数调用，用于热点检测
func (jit *JITCompiler) RecordFunctionCall(functionName string) {
	jit.hotspotDetector.RecordCall(functionName)
}

// GetStats 获取编译器统计信息
func (jit *JITCompiler) GetStats() CompilerStats {
	jit.stats.mu.RLock()
	defer jit.stats.mu.RUnlock()
	return *jit.stats
}

// applyOptimizations 应用优化
func (jit *JITCompiler) applyOptimizations(bytecode []opcodes.Instruction) ([]opcodes.Instruction, []Optimization) {
	// 这里将实现各种JIT特定的优化
	// 目前返回原始字节码
	return bytecode, nil
}

// ClearCompiledCode 清除所有已编译的代码（用于调试和测试）
func (jit *JITCompiler) ClearCompiledCode() {
	jit.mu.Lock()
	defer jit.mu.Unlock()

	jit.compiledCode = sync.Map{}

	jit.stats.mu.Lock()
	jit.stats.CompiledCodeSize = 0
	jit.stats.mu.Unlock()
}

// IsEnabled 检查JIT编译器是否启用
func (jit *JITCompiler) IsEnabled() bool {
	return jit != nil && jit.config != nil
}

// GetTopHotspots 获取调用频率最高的N个函数
func (jit *JITCompiler) GetTopHotspots(n int) []HotspotRank {
	return jit.hotspotDetector.GetTopHotspots(n)
}
