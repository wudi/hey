# PHP 解析器

[![Go Report Card](https://goreportcard.com/badge/github.com/wudi/php-parser)](https://goreportcard.com/report/github.com/wudi/php-parser)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org/doc/install)

一个用 Go 语言实现的高性能、全面的 PHP 解析器，完全支持 PHP 8.4 语法。该解析器提供了完整的词法分析、语法解析和 AST 生成功能。

[English](README.md) | [文档](docs/) | [示例](examples/)

## 特性

### 🚀 核心功能
- **完全兼容 PHP 8.4**: 包括现代 PHP 特性的完整语法支持
- **高性能**: 优化的词法分析器和解析器，支持性能基准测试
- **完整的 AST**: 丰富的抽象语法树，包含 150+ 种节点类型
- **错误恢复**: 强大的错误处理和部分解析功能
- **位置跟踪**: 所有标记和节点的精确行/列信息
- **访问者模式**: 全面的 AST 遍历和转换工具

### 📦 组件
- **词法分析器**: 基于状态机的分词器，包含 11 种解析状态
- **语法解析器**: 递归下降解析器，表达式采用 Pratt 解析
- **AST**: 基于接口的节点系统，支持访问者模式
- **命令行工具**: 功能丰富的命令行界面，用于解析和分析
- **示例**: 5 个全面的示例，演示不同的使用场景

### 🎯 应用场景
- **静态分析**: 代码质量工具、代码检查器、安全扫描器
- **开发工具**: IDE、语言服务器、代码格式化工具
- **文档生成**: API 文档生成器、代码可视化
- **重构工具**: 自动化代码转换和迁移工具
- **测试工具**: 代码覆盖率分析、变异测试
- **代码转译**: PHP 到 PHP 的转换、版本兼容性

## 快速开始

### 安装

```bash
go get github.com/wudi/php-parser
```

### 基本用法

```go
package main

import (
    "fmt"
    "github.com/wudi/php-parser/lexer"
    "github.com/wudi/php-parser/parser"
)

func main() {
    code := `<?php
    function hello($name) {
        return "Hello, " . $name . "!";
    }
    echo hello("World");
    ?>`

    l := lexer.New(code)
    p := parser.New(l)
    program := p.ParseProgram()

    if len(p.Errors()) > 0 {
        for _, err := range p.Errors() {
            fmt.Printf("错误: %s\n", err)
        }
        return
    }

    fmt.Printf("解析了 %d 条语句\n", len(program.Body))
    fmt.Printf("AST: %s\n", program.String())
}
```

### 命令行工具

构建和使用 CLI 工具：

```bash
# 构建解析器
go build -o php-parser ./cmd/php-parser

# 解析 PHP 文件
./php-parser example.php

# 显示标记和 AST
./php-parser -tokens -ast example.php

# 输出为 JSON 格式
./php-parser -format json example.php

# 从标准输入解析
echo '<?php echo "Hello"; ?>' | ./php-parser
```

## 架构

### 项目结构

```
php-parser/
├── ast/            # 抽象语法树实现
│   ├── node.go     # 150+ AST 节点类型（6K+ 行）
│   ├── kind.go     # AST 节点类型常量
│   ├── visitor.go  # 访问者模式工具
│   └── builder.go  # AST 构建帮助器
├── lexer/          # 词法分析器
│   ├── lexer.go    # 主词法分析器，状态机（1.5K+ 行）
│   ├── token.go    # PHP 标记定义（150+ 标记）
│   └── states.go   # 词法分析器状态管理
├── parser/         # 语法解析器
│   ├── parser.go   # 递归下降解析器（7K+ 行）
│   ├── pool.go     # 解析器池，支持并发
│   └── testdata/   # 测试用例和固件
├── cmd/            # 命令行界面
│   └── php-parser/ # CLI 实现（244 行）
├── examples/       # 使用示例和教程
│   ├── basic-parsing/      # 基本解析概念
│   ├── ast-visitor/        # 访问者模式示例
│   ├── token-extraction/   # 词法分析
│   ├── error-handling/     # 错误恢复示例
│   └── code-analysis/      # 静态分析工具
├── errors/         # 错误处理工具
└── scripts/        # 开发和测试脚本
```

**代码统计**: 16 个源文件中包含 18,500+ 行 Go 代码，29 个测试文件

### 核心组件

#### 词法分析器（分词器）
- **11 种解析状态**: 包括 `ST_IN_SCRIPTING`、`ST_DOUBLE_QUOTES`、`ST_HEREDOC`
- **150+ 标记类型**: 完全兼容 PHP 8.4 标记
- **状态机**: 处理复杂的 PHP 语法，如字符串插值
- **Shebang 支持**: 识别可执行的 PHP 文件
- **位置跟踪**: 行号、列号和偏移信息

#### 语法解析器（语法分析器）
- **递归下降**: 清晰、可维护的解析架构
- **Pratt 解析**: 优雅的操作符优先级处理（14 个级别）
- **错误恢复**: 遇到错误后继续解析以发现多个问题
- **50+ 解析函数**: 全面的 PHP 语法覆盖
- **替代语法**: 完全支持 `if:...endif;` 风格的构造

#### AST（抽象语法树）
- **基于接口**: 节点类型之间的清晰分离
- **150+ 节点类型**: 匹配 PHP 官方的 `zend_ast.h` 常量
- **访问者模式**: 轻松遍历和转换
- **JSON 序列化**: 导出 AST 供外部工具使用
- **位置保留**: 所有节点都保留源位置

## PHP 8.4 语言支持

### 操作符
- **算术运算**: `+`、`-`、`*`、`/`、`%`、`**`（幂运算）
- **赋值运算**: `=`、`+=`、`-=`、`*=`、`/=`、`%=`、`**=`、`.=`、`??=` 等
- **比较运算**: `==`、`===`、`!=`、`!==`、`<`、`<=`、`>`、`>=`、`<=>` （太空船操作符）
- **逻辑运算**: `&&`、`||`、`!`、`and`、`or`、`xor`
- **位运算**: `&`、`|`、`^`、`~`、`<<`、`>>`
- **空合并**: `??`、`??=`

### 语言构造
- **控制结构**: `if`、`else`、`elseif`、`while`、`for`、`foreach`、`switch`
- **替代语法**: `if:...endif;`、`while:...endwhile;`、`switch:...endswitch;`
- **函数**: 参数、返回类型、引用、可变参数
- **类**: 属性、方法、常量、继承、接口
- **命名空间**: 完整的命名空间支持与 `use` 语句
- **特殊**: `__halt_compiler()`、匹配表达式、属性

### 现代 PHP 特性
- **类型化属性**: `private int $id`、`public ?string $name`
- **联合类型**: `int|string`、`Foo|Bar|null`
- **交集类型**: `Foo&Bar`
- **匹配表达式**: 使用 `match()` 进行模式匹配
- **属性**: `#[Route('/api')]` 语法
- **空安全操作符**: `$user?->getProfile()?->getName()`
- **可见性类常量**: `private const SECRET = 'value'`

## 示例

`examples/` 目录包含全面的演示：

### 1. [基本解析](examples/basic-parsing/)
学习基本的解析概念和 AST 检查。

```bash
cd examples/basic-parsing && go run main.go
```

### 2. [AST 访问者模式](examples/ast-visitor/)
实现自定义访问者进行代码分析和遍历。

```bash
cd examples/ast-visitor && go run main.go
```

### 3. [标记提取](examples/token-extraction/)
探索词法分析和标记统计。

```bash
cd examples/token-extraction && go run main.go
```

### 4. [错误处理](examples/error-handling/)
了解解析器错误恢复和报告。

```bash
cd examples/error-handling && go run main.go
```

### 5. [代码分析](examples/code-analysis/)
构建具有度量和质量评估的静态分析工具。

```bash
cd examples/code-analysis && go run main.go
```

每个示例都包含：
- **可运行代码**: 完整的工作示例
- **文档**: 详细的 README 和说明
- **真实 PHP 样本**: 用于测试的真实代码
- **渐进复杂性**: 从初学者到高级

## 测试

### 运行测试

```bash
# 运行所有测试
go test ./...

# 详细输出运行
go test ./... -v

# 运行特定组件测试
go test ./lexer -v
go test ./parser -v
go test ./ast -v

# 运行基准测试
go test ./parser -bench=.
go test ./parser -bench=. -benchmem

# 运行特定测试用例
go test ./parser -run=TestParsing_TryCatchWithStatements
go test ./parser -run=TestParsing_ClassMethodsWithVisibility
```

### 测试覆盖率

项目维护全面的测试覆盖率：
- **29 个测试文件**: 覆盖所有主要组件
- **200+ 测试用例**: 包括边缘情况和错误条件
- **真实世界测试**: 针对主要 PHP 框架的兼容性测试
- **基准测试**: 性能验证和优化

### 兼容性测试

针对流行的 PHP 代码库进行测试：

```bash
# 使用 WordPress 测试
go run scripts/test_folder.go /path/to/wordpress-develop

# 使用 Laravel 测试
go run scripts/test_folder.go /path/to/framework

# 使用 Symfony 测试
go run scripts/test_folder.go /path/to/symfony
```

## 性能

### 基准测试
- **简单语句**: 每次解析约 1.6μs
- **复杂表达式**: 每次解析约 4μs
- **大文件**: 流式处理，高效的内存使用
- **并发解析**: 解析器池支持并行处理

### 内存使用
- **低内存占用**: 每次解析的内存分配最少
- **流式处理**: 大文件支持，无需加载整个内容
- **池模式**: 服务器应用程序的可重用解析器实例

## 开发

### 要求
- Go 1.21+
- 可选：PHP 8.4 用于兼容性测试

### 开发命令

```bash
# 构建 CLI 工具
go build -o php-parser ./cmd/php-parser

# 运行所有测试
go test ./...

# 运行性能基准测试
go test ./parser -bench=. -run=^$

# 内存分析测试
go test ./parser -bench=. -benchmem

# 兼容性测试
go run scripts/test_folder.go /path/to/php/codebase

# 清理构建产物
go clean
rm -f php-parser main
```

## 贡献

我们欢迎贡献！请查看我们的贡献指南：

1. **Fork** 仓库
2. **创建** 功能分支：`git checkout -b feature-name`
3. **编写** 变更测试
4. **确保** 所有测试通过：`go test ./...`
5. **运行** 代码检查：`go fmt ./...`
6. **提交** Pull Request

### 开发指南
- 遵循 Go 编码标准（`gofmt`、有效的 Go）
- 与官方实现保持 PHP 兼容性
- 为新功能添加全面的测试
- 参考 PHP 官方语法文件 `/php-src/Zend/zend_language_parser.y`
- 为新功能更新文档

## 生产应用场景

### 静态分析工具
- **代码质量**: 检测反模式、复杂性度量
- **安全扫描**: 查找漏洞、注入风险
- **样式检查**: 执行编码标准和约定

### 开发工具
- **IDE**: 语法高亮、自动完成、错误检测
- **语言服务器**: 编辑器支持的 LSP 实现
- **代码格式化工具**: 一致的代码样式和格式化

### 文档工具
- **API 生成器**: 从代码和注释中提取文档
- **代码可视化**: 生成类图、调用图
- **依赖分析**: 跟踪代码关系和耦合

### 迁移和重构
- **版本升级**: PHP 版本兼容性转换
- **框架迁移**: 自动化代码模式更新
- **代码现代化**: 应用现代 PHP 实践和语法

## 许可证

该项目在 MIT 许可证下授权 - 请查看 [LICENSE](LICENSE) 文件了解详情。

## 致谢

- **PHP 团队**: 提供官方 PHP 语言规范
- **Go 社区**: 提供优秀的工具和生态系统
- **贡献者**: 所有为该项目做出贡献的人

## 链接

- **仓库**: [github.com/wudi/php-parser](https://github.com/wudi/php-parser)
- **问题**: [报告错误和功能请求](https://github.com/wudi/php-parser/issues)
- **文档**: [详细的 API 文档](docs/)
- **示例**: [代码示例和教程](examples/)

---

**用 ❤️ 构建于 Go | PHP 8.4 兼容 | 生产就绪**