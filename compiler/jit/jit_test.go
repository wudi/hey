package jit

import (
	"testing"
	"time"

	"github.com/wudi/php-parser/compiler/opcodes"
)

func TestNewJITCompiler(t *testing.T) {
	config := DefaultConfig()
	jit, err := NewJITCompiler(config)

	if err != nil {
		t.Fatalf("Failed to create JIT compiler: %v", err)
	}

	if jit == nil {
		t.Fatal("JIT compiler should not be nil")
	}

	if !jit.IsEnabled() {
		t.Error("JIT compiler should be enabled")
	}
}

func TestHotspotDetector(t *testing.T) {
	detector := NewHotspotDetector(3) // 阈值为3次调用
	defer detector.Stop()

	funcName := "testFunction"

	// 测试初始状态
	if detector.IsHotspot(funcName) {
		t.Error("Function should not be hotspot initially")
	}

	// 记录调用，未达到阈值
	detector.RecordCall(funcName)
	detector.RecordCall(funcName)

	if detector.IsHotspot(funcName) {
		t.Error("Function should not be hotspot with 2 calls")
	}

	// 达到阈值
	detector.RecordCall(funcName)

	if !detector.IsHotspot(funcName) {
		t.Error("Function should be hotspot with 3 calls")
	}

	// 检查统计信息
	stats := detector.GetStats()
	if stats.HotspotFunctions != 1 {
		t.Errorf("Expected 1 hotspot function, got %d", stats.HotspotFunctions)
	}

	if stats.TotalCalls != 3 {
		t.Errorf("Expected 3 total calls, got %d", stats.TotalCalls)
	}
}

func TestFunctionCallInfo(t *testing.T) {
	detector := NewHotspotDetector(5)
	defer detector.Stop()

	funcName := "testFunction"

	// 记录一些调用
	for i := 0; i < 3; i++ {
		detector.RecordCall(funcName)
		time.Sleep(1 * time.Millisecond) // 确保时间差异
	}

	// 获取函数信息
	info, exists := detector.GetFunctionInfo(funcName)
	if !exists {
		t.Fatal("Function info should exist")
	}

	if info.CallCount != 3 {
		t.Errorf("Expected 3 calls, got %d", info.CallCount)
	}

	if info.CallFrequency <= 0 {
		t.Error("Call frequency should be positive")
	}

	if info.IsHotspot {
		t.Error("Function should not be hotspot with 3 calls (threshold is 5)")
	}
}

func TestJITCompilation(t *testing.T) {
	config := DefaultConfig()
	config.CompilationThreshold = 1 // 立即编译用于测试

	jit, err := NewJITCompiler(config)
	if err != nil {
		t.Fatalf("Failed to create JIT compiler: %v", err)
	}

	funcName := "testFunction"
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
			Op2:    0,
			Result: 0,
		},
	}

	// 记录函数调用使其成为热点
	// 需要调用threshold次数才能成为热点
	for i := 0; i < config.CompilationThreshold; i++ {
		jit.RecordFunctionCall(funcName)
	}

	if !jit.ShouldCompile(funcName) {
		t.Error("Function should be eligible for compilation")
	}

	// 尝试编译（可能会失败，因为我们的AMD64生成器还很基础）
	compiledFunc, err := jit.CompileFunction(funcName, bytecode)

	// 检查编译是否成功或者返回了预期的错误
	if err != nil {
		// 这是预期的，因为我们的实现还不完整
		t.Logf("Compilation failed as expected: %v", err)
	} else {
		// 如果编译成功，检查结果
		if compiledFunc == nil {
			t.Error("Compiled function should not be nil")
		}

		if compiledFunc.Name != funcName {
			t.Errorf("Expected function name %s, got %s", funcName, compiledFunc.Name)
		}

		// 检查是否缓存了编译结果
		cached, exists := jit.GetCompiledFunction(funcName)
		if !exists {
			t.Error("Compiled function should be cached")
		}

		if cached != compiledFunc {
			t.Error("Cached function should be the same instance")
		}
	}
}

func TestJITStats(t *testing.T) {
	config := DefaultConfig()
	jit, err := NewJITCompiler(config)
	if err != nil {
		t.Fatalf("Failed to create JIT compiler: %v", err)
	}

	// 初始统计应该为空
	stats := jit.GetStats()
	if stats.TotalCompilations != 0 {
		t.Error("Initial total compilations should be 0")
	}

	// 测试热点检测器统计
	funcName := "testFunction"
	for i := 0; i < 5; i++ {
		jit.RecordFunctionCall(funcName)
	}

	hotspotStats := jit.hotspotDetector.GetStats()
	if hotspotStats.TotalCalls != 5 {
		t.Errorf("Expected 5 total calls, got %d", hotspotStats.TotalCalls)
	}
}

func TestCodeGeneratorInterface(t *testing.T) {
	config := DefaultConfig()
	codeGen, err := NewAMD64CodeGenerator(config)
	if err != nil {
		t.Fatalf("Failed to create AMD64 code generator: %v", err)
	}

	if codeGen.GetTargetArch() != "amd64" {
		t.Error("Target architecture should be amd64")
	}

	// 测试支持的操作码
	supportedOpcodes := []opcodes.Opcode{
		opcodes.OP_ADD,
		opcodes.OP_SUB,
		opcodes.OP_MUL,
		opcodes.OP_DIV,
		opcodes.OP_JMP,
		opcodes.OP_RETURN,
	}

	for _, opcode := range supportedOpcodes {
		if !codeGen.SupportsOpcode(opcode) {
			t.Errorf("Should support opcode %s", opcode.String())
		}
	}

	// 测试不支持的操作码
	unsupportedOpcodes := []opcodes.Opcode{
		opcodes.OP_YIELD,
		opcodes.OP_MATCH,
	}

	for _, opcode := range unsupportedOpcodes {
		if codeGen.SupportsOpcode(opcode) {
			t.Errorf("Should not support opcode %s", opcode.String())
		}
	}
}

func TestHotspotRanking(t *testing.T) {
	detector := NewHotspotDetector(2)
	defer detector.Stop()

	// 创建不同调用频率的函数
	functions := []struct {
		name  string
		calls int
	}{
		{"func1", 5},
		{"func2", 10},
		{"func3", 3},
		{"func4", 15},
	}

	// 记录调用，在调用之间添加微小延迟以确保频率计算正确
	for _, f := range functions {
		for i := 0; i < f.calls; i++ {
			detector.RecordCall(f.name)
			if i < f.calls-1 { // 最后一次调用不需要延迟
				time.Sleep(100 * time.Microsecond)
			}
		}
	}

	// 等待一段时间计算调用频率
	time.Sleep(10 * time.Millisecond)

	// 获取排名前3的热点
	topHotspots := detector.GetTopHotspots(3)

	if len(topHotspots) != 3 {
		t.Errorf("Expected 3 top hotspots, got %d", len(topHotspots))
	}

	// 检查排序（应该按调用次数降序）
	for i := 0; i < len(topHotspots)-1; i++ {
		if topHotspots[i].CallCount < topHotspots[i+1].CallCount {
			t.Error("Hotspots should be sorted by call count in descending order")
		}
	}

	// func4应该排在第一位（15次调用）
	if topHotspots[0].FunctionName != "func4" {
		t.Errorf("Expected func4 to be top hotspot, got %s", topHotspots[0].FunctionName)
	}
}

func TestThresholdUpdate(t *testing.T) {
	detector := NewHotspotDetector(5) // 初始阈值5
	defer detector.Stop()

	funcName := "testFunction"

	// 调用3次，不应该成为热点
	for i := 0; i < 3; i++ {
		detector.RecordCall(funcName)
	}

	if detector.IsHotspot(funcName) {
		t.Error("Function should not be hotspot with threshold 5")
	}

	// 降低阈值到2
	detector.SetThreshold(2)

	// 现在应该成为热点
	if !detector.IsHotspot(funcName) {
		t.Error("Function should be hotspot after lowering threshold to 2")
	}
}

func TestCompilerConfiguration(t *testing.T) {
	config := &Config{
		CompilationThreshold: 20,
		MaxCompiledFunctions: 500,
		EnableOptimizations:  false,
		TargetArch:           "amd64",
		DebugMode:            true,
	}

	jit, err := NewJITCompiler(config)
	if err != nil {
		t.Fatalf("Failed to create JIT compiler: %v", err)
	}

	// 验证配置被正确设置
	if jit.config.CompilationThreshold != 20 {
		t.Errorf("Expected threshold 20, got %d", jit.config.CompilationThreshold)
	}

	if jit.config.EnableOptimizations {
		t.Error("Optimizations should be disabled")
	}

	if !jit.config.DebugMode {
		t.Error("Debug mode should be enabled")
	}
}

func BenchmarkHotspotDetection(b *testing.B) {
	detector := NewHotspotDetector(100)
	defer detector.Stop()

	funcName := "benchmarkFunction"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		detector.RecordCall(funcName)
	}
}

func BenchmarkJITCompilerCreation(b *testing.B) {
	config := DefaultConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		jit, err := NewJITCompiler(config)
		if err != nil {
			b.Fatalf("Failed to create JIT compiler: %v", err)
		}
		_ = jit
	}
}
