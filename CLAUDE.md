# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Prompts
```
请使用 Go 语言帮助我设计和实现一个 PHP Parser，务必严格遵循以下技术及流程要求：

总体目标

编写一个能够将 PHP 源代码（如 <?php echo "Hello, world!"; ?>）解析为抽象语法树（AST）的 Go 程序。要求支持 PHP 7 及以上基础语法（如变量、函数、条件语句、循环语句、表达式等），并具备良好的可扩展性。

模块结构

词法分析器（Lexer）：将 PHP 源代码分割成 Token 流，识别关键字、标识符、符号、字符串、数字、注释等。
语法分析器（Parser）：根据 PHP 语法规则将 Token 流解析为 AST。
AST 节点定义：为不同的 PHP 语法结构定义相应的节点类型（如变量、函数、语句等）。
错误处理机制：优雅地报告语法和词法错误，指出出错位置和原因。
代码风格与设计原则

遵循 Go 的标准代码风格，模块化设计，结构清晰，易于维护。
使用接口和结构体抽象，便于扩展和单元测试。
注释详细，说明每个模块和关键函数的作用。
功能要求

能解析并输出 AST（建议使用 JSON 格式输出）。
能报告并定位语法错误。
提供至少一个完整的示例，用于演示如何解析一段 PHP 代码并输出结果。
项目结构与实现思路

请先输出项目整体目录结构及核心文件。
分模块详细说明实现思路和关键代码。
代码应便于扩展（如支持更多 PHP 语法或添加静态分析功能）。
扩展要求

词法分析器（Lexer）中，Token 的命名和取值需与原始 PHP Token 保持一致，参考官方实现，确保兼容性。
本地已安装 PHP 8.4，路径为 /bin/php，如有需要可通过命令行调用 PHP 进行 Token 流或 AST 的生成，对比测试解析结果。
PHP 的官方源代码路径为 /home/ubuntu/php-src，可以参考其实现细节，确保你的设计与 PHP 官方行为保持一致。
Lexer 深度要求

注意 Lexer 需要实现多个状态（state），如初始状态、字符串状态、注释状态等。请深入研究并阐述 PHP Lexer 的核心原理与状态转换机制，保证 Token 化过程的高还原性和准确性。
测试规范

单元测试需严格遵循 Go 语言规范，所有模块应逐步开发、逐步测试，确保每个小模块的正确性和稳定性。
设计文档要求

所有设计文档、技术说明、架构方案均需统一保存在项目的 docs 目录下，便于后续查阅和管理。
实现完整性要求

需要完整实现各个模块和功能，而不仅仅输出框架结构或伪代码。请给出可运行的、具备代表性的完整代码，并确保核心功能能够实际运行和测试。
其他要求

如有相关开源项目、文献或参考资料，请在最后简要列出供学习。
输出方式：

请分阶段、分模块详细输出设计思路和关键实现，最后给出完整示例代码，并说明如何运行和测试。
```

## Commands

**Build and Test:**
```bash
go test ./...              # Run all tests
go test ./lexer -v         # Run lexer tests with verbose output
go test ./parser -v        # Run parser tests with verbose output
go test ./ast -v           # Run AST tests with verbose output
```

**Build CLI Tool:**
```bash
go build -o phpparse ./cmd/phpparse  # Build command-line tool
./phpparse -i example.php            # Parse a PHP file
./phpparse -tokens -ast              # Show tokens and AST structure
echo '<?php echo "Hello"; ?>' | ./phpparse  # Parse from stdin
```

**Cleanup:**
```bash
go clean                         # Clean build artifacts
rm -f phpparse                   # Remove built binary
rm -f debug*.go                  # Remove debug files (if needed)
```

## Architecture Overview

This is a PHP parser implementation in Go with the following structure:

### Core Modules
- **`lexer/`**: PHP lexical analyzer with state machine
  - `token.go`: PHP token definitions (150+ tokens matching PHP 8.4)
  - `states.go`: Lexer state management (11 states)
  - `lexer.go`: Main lexer implementation with PHP tag recognition

- **`parser/`**: Recursive descent parser
  - `parser.go`: Main parser with operator precedence (Pratt parsing)
  - Handles expressions, statements, control structures

- **`ast/`**: Abstract Syntax Tree nodes
  - `node.go`: Interface-based AST node system
  - Supports JSON serialization and string representation

- **`errors/`**: Error handling with position tracking

### Command Line Interface
- **`cmd/phpparse/`**: CLI tool for parsing PHP code
  - Supports multiple output formats (JSON, AST, tokens)
  - File and stdin input support
  - Error reporting with position information

### Key Design Features
- **PHP Compatibility**: Token IDs match PHP 8.4 official implementation
- **State Machine**: Lexer handles multiple states (scripting, strings, heredoc, etc.)
- **Error Recovery**: Detailed error reporting with line/column positions
- **Modular Design**: Clean separation between lexer, parser, and AST

### Testing Strategy
- Unit tests for each module (`*_test.go` files)
- Integration tests for complete parsing workflow
- Compatibility tests comparing with PHP's `token_get_all()`
- Operator precedence and edge case testing

### Common Development Tasks
- Adding new PHP syntax support (extend parser and AST)
- Improving error messages and recovery
- Enhancing performance of lexer/parser
- Adding static analysis capabilities