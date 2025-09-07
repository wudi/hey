//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"time"

	"github.com/wudi/php-parser/compiler/jit"
	"github.com/wudi/php-parser/compiler/opcodes"
	"github.com/wudi/php-parser/compiler/values"
	"github.com/wudi/php-parser/compiler/vm"
)

func main() {
	fmt.Println("=== PHP JIT 编译器演示 ===\n")

	// 1. 创建带JIT的虚拟机
	fmt.Println("1. 初始化虚拟机和JIT编译器...")

	jitConfig := &jit.Config{
		CompilationThreshold: 3, // 调用3次后开始编译
		MaxCompiledFunctions: 100,
		EnableOptimizations:  true,
		TargetArch:           "amd64",
		DebugMode:            true,
	}

	vmWithJIT, err := vm.NewVirtualMachineWithJIT(jitConfig)
	if err != nil {
		fmt.Printf("创建JIT虚拟机失败: %v\n", err)
		return
	}

	vmWithoutJIT := vm.NewVirtualMachine()
	vmWithoutJIT.JITEnabled = false

	fmt.Printf("✅ JIT虚拟机创建成功 (编译阈值: %d)\n", jitConfig.CompilationThreshold)
	fmt.Printf("✅ 标准虚拟机创建成功\n\n")

	// 2. 准备测试函数
	fmt.Println("2. 准备测试函数...")

	// 创建一个简单的数学计算函数
	testFunction := &vm.Function{
		Name: "calculateSum",
		Parameters: []vm.Parameter{
			{Name: "a", Type: "int"},
			{Name: "b", Type: "int"},
		},
		Instructions: []opcodes.Instruction{
			{
				Opcode:  opcodes.OP_FETCH_R,
				Op1:     0, // 参数a (变量槽0)
				Result:  1, // 临时变量1
				OpType1: byte(opcodes.IS_VAR),
				OpType2: byte(opcodes.IS_TMP_VAR),
			},
			{
				Opcode:  opcodes.OP_FETCH_R,
				Op1:     1, // 参数b (变量槽1)
				Result:  2, // 临时变量2
				OpType1: byte(opcodes.IS_VAR),
				OpType2: byte(opcodes.IS_TMP_VAR),
			},
			{
				Opcode:  opcodes.OP_ADD,
				Op1:     1, // 临时变量1
				Op2:     2, // 临时变量2
				Result:  3, // 结果临时变量3
				OpType1: byte(opcodes.IS_TMP_VAR),
				OpType2: byte(opcodes.IS_TMP_VAR),
			},
			{
				Opcode:  opcodes.OP_RETURN,
				Op1:     3, // 返回临时变量3
				OpType1: byte(opcodes.IS_TMP_VAR),
			},
		},
		Constants: []*values.Value{},
	}

	// 准备执行上下文
	setupContext := func() *vm.ExecutionContext {
		ctx := vm.NewExecutionContext()
		ctx.Functions = make(map[string]*vm.Function)
		ctx.Functions["calculateSum"] = testFunction
		return ctx
	}

	fmt.Println("✅ 测试函数准备完成: calculateSum(a, b) = a + b\n")

	// 3. 热点检测演示
	fmt.Println("3. 热点检测演示...")

	fmt.Printf("编译阈值: %d 次调用\n", jitConfig.CompilationThreshold)

	for i := 1; i <= 5; i++ {
		vmWithJIT.JITCompiler.RecordFunctionCall("calculateSum")
		isHotspot := vmWithJIT.JITCompiler.ShouldCompile("calculateSum")

		hotspotStats := vmWithJIT.JITCompiler.GetStats()
		fmt.Printf("第 %d 次调用 - 热点状态: %v, 总编译次数: %d\n",
			i, isHotspot, hotspotStats.TotalCompilations)

		if isHotspot && i == jitConfig.CompilationThreshold {
			fmt.Printf("🔥 函数 'calculateSum' 已被识别为热点！\n")
		}

		time.Sleep(10 * time.Millisecond) // 模拟执行间隔
	}
	fmt.Println()

	// 4. JIT编译演示
	fmt.Println("4. JIT编译演示...")

	if vmWithJIT.JITCompiler.ShouldCompile("calculateSum") {
		fmt.Println("正在尝试JIT编译 'calculateSum'...")

		compiledFunc, err := vmWithJIT.JITCompiler.CompileFunction("calculateSum", testFunction.Instructions)
		if err != nil {
			fmt.Printf("❌ JIT编译失败: %v\n", err)
			fmt.Println("   这是预期的，因为完整的机器码执行尚未实现")
		} else {
			fmt.Printf("✅ JIT编译成功！\n")
			fmt.Printf("   函数名: %s\n", compiledFunc.Name)
			fmt.Printf("   机器码大小: %d 字节\n", len(compiledFunc.MachineCode))
			fmt.Printf("   优化级别: %d\n", compiledFunc.OptimizationLevel)
		}
	}
	fmt.Println()

	// 5. 性能对比演示
	fmt.Println("5. 性能对比演示...")

	runTest := func(vm *vm.VirtualMachine, name string) time.Duration {
		start := time.Now()
		iterations := 100

		for i := 0; i < iterations; i++ {
			ctx := setupContext()

			// 模拟函数调用指令序列
			instructions := []opcodes.Instruction{
				{
					Opcode:  opcodes.OP_INIT_FCALL,
					Op1:     0, // 函数名索引
					OpType1: byte(opcodes.IS_CONST),
				},
				{
					Opcode:  opcodes.OP_SEND_VAL,
					Op1:     1, // 参数值10
					OpType1: byte(opcodes.IS_CONST),
				},
				{
					Opcode:  opcodes.OP_SEND_VAL,
					Op1:     2, // 参数值20
					OpType1: byte(opcodes.IS_CONST),
				},
			}

			// 准备常量池
			constants := []*values.Value{
				values.NewString("calculateSum"), // 函数名
				values.NewInt(10),                // 参数a
				values.NewInt(20),                // 参数b
			}

			// 执行（会自动处理JIT优化）
			vm.Execute(ctx, instructions, constants, ctx.Functions, nil)
		}

		elapsed := time.Since(start)
		fmt.Printf("%s: %d 次迭代耗时 %v (平均: %v/次)\n",
			name, iterations, elapsed, elapsed/time.Duration(iterations))

		return elapsed
	}

	// 运行标准解释器
	interpretTime := runTest(vmWithoutJIT, "标准解释器")

	// 运行JIT版本（实际会回退到解释器，但会进行热点检测和编译尝试）
	jitTime := runTest(vmWithJIT, "JIT编译器")

	// 计算性能差异
	if jitTime < interpretTime {
		speedup := float64(interpretTime) / float64(jitTime)
		fmt.Printf("🚀 JIT版本快 %.2fx\n", speedup)
	} else {
		overhead := float64(jitTime) / float64(interpretTime)
		fmt.Printf("⚠️ JIT版本有 %.2fx 开销（编译开销）\n", overhead)
	}
	fmt.Println()

	// 6. 统计信息展示
	fmt.Println("6. JIT编译器统计信息...")

	stats := vmWithJIT.JITCompiler.GetStats()
	fmt.Printf("总编译次数: %d\n", stats.TotalCompilations)
	fmt.Printf("成功编译: %d\n", stats.SuccessfulCompilations)
	fmt.Printf("失败编译: %d\n", stats.FailedCompilations)
	fmt.Printf("平均编译时间: %v\n", stats.AverageCompileTime)
	fmt.Printf("编译代码总大小: %d 字节\n", stats.CompiledCodeSize)
	fmt.Println()

	// 7. 热点函数排名
	fmt.Println("7. 热点函数排名...")

	// 添加更多函数调用以生成排名
	testFunctions := []string{"calculateSum", "processData", "validateInput", "formatOutput"}
	callCounts := []int{10, 15, 5, 8}

	for i, funcName := range testFunctions {
		for j := 0; j < callCounts[i]; j++ {
			vmWithJIT.JITCompiler.RecordFunctionCall(funcName)
			time.Sleep(time.Microsecond) // 确保时间差异
		}
	}

	// 等待频率计算
	time.Sleep(10 * time.Millisecond)

	topHotspots := vmWithJIT.JITCompiler.GetTopHotspots(5)
	for i, hotspot := range topHotspots {
		status := "❄️"
		if hotspot.IsHotspot {
			status = "🔥"
		}
		fmt.Printf("%d. %s %s - 调用次数: %d, 频率: %.2f 次/秒\n",
			i+1, status, hotspot.FunctionName, hotspot.CallCount, hotspot.CallFrequency)
	}
	fmt.Println()

	// 8. 代码生成器测试
	fmt.Println("8. 代码生成器能力测试...")

	codeGen, err := jit.NewAMD64CodeGenerator(jitConfig)
	if err != nil {
		fmt.Printf("❌ 创建代码生成器失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 目标架构: %s\n", codeGen.GetTargetArch())

	supportedOpcodes := []opcodes.Opcode{
		opcodes.OP_ADD, opcodes.OP_SUB, opcodes.OP_MUL, opcodes.OP_DIV,
		opcodes.OP_JMP, opcodes.OP_RETURN, opcodes.OP_ASSIGN,
	}

	fmt.Println("支持的指令:")
	for _, opcode := range supportedOpcodes {
		supported := codeGen.SupportsOpcode(opcode)
		status := "❌"
		if supported {
			status = "✅"
		}
		fmt.Printf("  %s %s\n", status, opcode.String())
	}
	fmt.Println()

	// 9. 总结
	fmt.Println("=== 演示总结 ===")
	fmt.Println("✅ JIT编译器核心功能已实现:")
	fmt.Println("   • 热点检测和自适应编译")
	fmt.Println("   • 多架构代码生成框架")
	fmt.Println("   • 与虚拟机无缝集成")
	fmt.Println("   • 完整的错误回退机制")
	fmt.Println("   • 详细的性能统计")
	fmt.Println("   • 扩展性架构设计")
	fmt.Println()
	fmt.Println("⚠️  当前限制:")
	fmt.Println("   • 机器码执行需要系统级内存管理")
	fmt.Println("   • 需要实现完整的调用约定")
	fmt.Println("   • 需要添加更多指令支持")
	fmt.Println()
	fmt.Println("🚀 未来扩展:")
	fmt.Println("   • 完整的x86-64机器码执行")
	fmt.Println("   • ARM64架构支持")
	fmt.Println("   • 高级优化算法")
	fmt.Println("   • 调试和性能分析工具")
}
