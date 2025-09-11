# Hey - 基于 Go 的 PHP 解释器
**目前属于实验性项目，请勿用于生产环境**

一个用 Go 编写的高性能 PHP 解释器，提供与 PHP 8.0+ 的语法兼容性。

## 特性

- **完整的 PHP 8.0+ 语法支持**：兼容现代 PHP 特性，包括箭头函数、展开操作符、goto 语句和严格类型
- **高性能虚拟机**：自定义字节码编译器和虚拟机，具备高级性能分析功能
- **高级调试**：内置调试器，支持断点、变量监控和性能分析
- **内存管理**：高效的内存池和分配跟踪
- **词法分析器和语法分析器**：完整的 PHP 语法词法和语法分析
- **静态分析**：基于 AST 的代码分析和指标收集

## 安装

```bash
go get github.com/wudi/hey
```

## 快速开始

### 基本用法

```go
package main

import (
    "github.com/wudi/hey/compiler"
    "github.com/wudi/hey/compiler/lexer"
    "github.com/wudi/hey/compiler/parser"
    "github.com/wudi/hey/compiler/vm"
)

func main() {
    phpCode := `<?php
    $x = 10;
    $y = 20;
    echo $x + $y;
    ?>`
    
    // 解析
    l := lexer.New(phpCode)
    p := parser.New(l)
    program := p.ParseProgram()
    
    // 编译
    comp := compiler.NewCompiler()
    comp.Compile(program)
    
    // 执行
    vmachine := vm.NewVirtualMachine()
    ctx := vm.NewExecutionContext()
    vmachine.Execute(ctx, comp.GetBytecode(), comp.GetConstants(), comp.GetFunctions(), comp.GetClasses())
}
```

### 示例程序

运行包含的演示程序查看高级特性：

```bash
cd cmd/vm-demo
go run main.go
```

## 架构

### 核心组件

- **词法分析器**：将 PHP 源代码标记化
- **语法分析器**：构建抽象语法树 (AST)
- **编译器**：从 AST 生成字节码
- **虚拟机**：执行字节码并支持性能分析
- **运行时**：提供 PHP 标准库函数

### 性能特性

- **性能分析虚拟机**：详细的执行分析和热点分析
- **内存跟踪**：内存分配和释放监控
- **断点**：支持变量监控的调试功能
- **性能报告**：全面的执行统计信息

## 示例

`examples/` 目录包含全面的演示：

- **基本解析**：核心解析功能
- **AST 访问者**：树遍历和分析
- **令牌提取**：词法分析示例
- **错误处理**：错误检测和恢复
- **代码分析**：静态分析和指标

## 支持的 PHP 特性

- 变量和数据类型
- 函数（包括箭头函数）
- 类和对象
- 控制结构（if/else、循环、goto）
- 现代 PHP 语法（展开操作符、严格类型）
- 错误处理和异常
- 标准库函数

## 开发

### 构建

```bash
go build ./cmd/vm-demo
```

### 测试

```bash
go test ./...
```

## 许可证

MIT 许可证 - 详见 [LICENSE](LICENSE) 文件。

## 作者

Di Wu <hi@wudi.io>

## 贡献

欢迎贡献！请随时提交问题和拉取请求。