# PHP Parser

**EXPERIMENTAL**

[![Go Report Card](https://goreportcard.com/badge/github.com/wudi/php-parser)](https://goreportcard.com/report/github.com/wudi/php-parser)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org/doc/install)

A high-performance, comprehensive PHP parser implementation in Go with full PHP 8.4 syntax support. This parser provides complete lexical analysis, syntax parsing, and AST generation capabilities for PHP code.

[中文文档](README-zh_CN.md) | [Documentation](docs/) | [Examples](examples/)

## Features

### 🚀 Core Capabilities
- **Full PHP 8.4 Compatibility**: Complete syntax support including modern PHP features
- **High Performance**: Optimized lexer and parser with benchmark support
- **Complete AST**: Rich Abstract Syntax Tree with 150+ node types
- **Error Recovery**: Robust error handling with partial parsing capabilities
- **Position Tracking**: Precise line/column information for all tokens and nodes
- **Visitor Pattern**: Comprehensive AST traversal and transformation utilities

### 📦 Components
- **Lexer**: State-machine based tokenizer with 11 parsing states
- **Parser**: Recursive descent parser with Pratt parsing for expressions  
- **AST**: Interface-based node system with visitor pattern support
- **CLI Tool**: Feature-rich command-line interface for parsing and analysis
- **Examples**: 5 comprehensive examples demonstrating different use cases

### 🎯 Use Cases
- **Static Analysis**: Code quality tools, linters, security scanners
- **Development Tools**: IDEs, language servers, code formatters
- **Documentation**: API documentation generators, code visualization
- **Refactoring**: Automated code transformation and migration tools
- **Testing**: Code coverage analysis, mutation testing
- **Transpilation**: PHP-to-PHP transformations, version compatibility

## Quick Start

### Installation

```bash
go get github.com/wudi/php-parser
```

### Basic Usage

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
            fmt.Printf("Error: %s\n", err)
        }
        return
    }

    fmt.Printf("Parsed %d statements\n", len(program.Body))
    fmt.Printf("AST: %s\n", program.String())
}
```

### Command Line Tool

Build and use the CLI tool:

```bash
# Build the parser
go build -o php-parser ./cmd/php-parser

# Parse a PHP file
./php-parser example.php

# Show tokens and AST
./php-parser -tokens -ast example.php

# Output as JSON
./php-parser -format json example.php

# Parse from stdin
echo '<?php echo "Hello"; ?>' | ./php-parser
```

## Architecture

### Project Structure

```
php-parser/
├── ast/            # Abstract Syntax Tree implementation
│   ├── node.go     # 150+ AST node types (6K+ lines)
│   ├── kind.go     # AST node type constants  
│   ├── visitor.go  # Visitor pattern utilities
│   └── builder.go  # AST construction helpers
├── lexer/          # Lexical analyzer  
│   ├── lexer.go    # Main lexer with state machine (1.5K+ lines)
│   ├── token.go    # PHP token definitions (150+ tokens)
│   └── states.go   # Lexer state management
├── parser/         # Syntax parser
│   ├── parser.go   # Recursive descent parser (7K+ lines)
│   ├── pool.go     # Parser pooling for concurrency
│   └── testdata/   # Test cases and fixtures
├── cmd/            # Command-line interface
│   └── php-parser/ # CLI implementation (244 lines)
├── examples/       # Usage examples and tutorials
│   ├── basic-parsing/      # Fundamental parsing concepts
│   ├── ast-visitor/        # Visitor pattern examples
│   ├── token-extraction/   # Lexical analysis
│   ├── error-handling/     # Error recovery examples
│   └── code-analysis/      # Static analysis tools
├── errors/         # Error handling utilities  
└── scripts/        # Development and testing scripts
```

**Code Statistics**: 18,500+ lines of Go code across 16 source files with 29 test files

### Core Components

#### Lexer (Tokenizer)
- **11 Parsing States**: Including `ST_IN_SCRIPTING`, `ST_DOUBLE_QUOTES`, `ST_HEREDOC`
- **150+ Token Types**: Complete PHP 8.4 token compatibility
- **State Machine**: Handles complex PHP syntax like string interpolation
- **Shebang Support**: Recognizes executable PHP files
- **Position Tracking**: Line, column, and offset information

#### Parser (Syntax Analyzer)  
- **Recursive Descent**: Clean, maintainable parsing architecture
- **Pratt Parsing**: Elegant operator precedence handling (14 levels)
- **Error Recovery**: Continues parsing after errors to find multiple issues
- **50+ Parse Functions**: Comprehensive PHP syntax coverage
- **Alternative Syntax**: Full support for `if:...endif;` style constructs

#### AST (Abstract Syntax Tree)
- **Interface-Based**: Clean separation between node types
- **150+ Node Types**: Matching PHP's official `zend_ast.h` constants
- **Visitor Pattern**: Easy traversal and transformation
- **JSON Serialization**: Export AST for external tools
- **Position Preservation**: All nodes retain source location

## PHP 8.4 Language Support

### Operators
- **Arithmetic**: `+`, `-`, `*`, `/`, `%`, `**` (power)
- **Assignment**: `=`, `+=`, `-=`, `*=`, `/=`, `%=`, `**=`, `.=`, `??=`, etc.
- **Comparison**: `==`, `===`, `!=`, `!==`, `<`, `<=`, `>`, `>=`, `<=>` (spaceship)
- **Logical**: `&&`, `||`, `!`, `and`, `or`, `xor`
- **Bitwise**: `&`, `|`, `^`, `~`, `<<`, `>>`
- **Null Coalescing**: `??`, `??=`

### Language Constructs  
- **Control Structures**: `if`, `else`, `elseif`, `while`, `for`, `foreach`, `switch`
- **Alternative Syntax**: `if:...endif;`, `while:...endwhile;`, `switch:...endswitch;`
- **Functions**: Parameters, return types, references, variadic
- **Classes**: Properties, methods, constants, inheritance, interfaces
- **Namespaces**: Full namespace support with `use` statements
- **Special**: `__halt_compiler()`, match expressions, attributes

### Modern PHP Features
- **Typed Properties**: `private int $id`, `public ?string $name`
- **Union Types**: `int|string`, `Foo|Bar|null`
- **Intersection Types**: `Foo&Bar`  
- **Match Expressions**: Pattern matching with `match()`
- **Attributes**: `#[Route('/api')]` syntax
- **Nullsafe Operator**: `$user?->getProfile()?->getName()`
- **Class Constants with Visibility**: `private const SECRET = 'value'`

## Examples

The `examples/` directory contains comprehensive demonstrations:

### 1. [Basic Parsing](examples/basic-parsing/)
Learn fundamental parsing concepts and AST examination.

```bash
cd examples/basic-parsing && go run main.go
```

### 2. [AST Visitor Pattern](examples/ast-visitor/)  
Implement custom visitors for code analysis and traversal.

```bash
cd examples/ast-visitor && go run main.go
```

### 3. [Token Extraction](examples/token-extraction/)
Explore lexical analysis and token statistics.

```bash
cd examples/token-extraction && go run main.go
```

### 4. [Error Handling](examples/error-handling/)
Understand parser error recovery and reporting.

```bash
cd examples/error-handling && go run main.go  
```

### 5. [Code Analysis](examples/code-analysis/)
Build static analysis tools with metrics and quality assessment.

```bash
cd examples/code-analysis && go run main.go
```

Each example includes:
- **Runnable Code**: Complete working examples
- **Documentation**: Detailed README with explanations  
- **Real PHP Samples**: Realistic code for testing
- **Progressive Complexity**: From beginner to advanced

## Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run with verbose output
go test ./... -v

# Run specific component tests
go test ./lexer -v
go test ./parser -v  
go test ./ast -v

# Run benchmarks
go test ./parser -bench=.
go test ./parser -bench=. -benchmem

# Run specific test cases
go test ./parser -run=TestParsing_TryCatchWithStatements
go test ./parser -run=TestParsing_ClassMethodsWithVisibility
```

### Test Coverage

The project maintains comprehensive test coverage with:
- **29 Test Files**: Covering all major components
- **200+ Test Cases**: Including edge cases and error conditions  
- **Real-world Testing**: Compatibility tests against major PHP frameworks
- **Benchmark Tests**: Performance validation and optimization

### Compatibility Testing

Test against popular PHP codebases:

```bash
# Test with WordPress
go run scripts/test_folder.go /path/to/wordpress-develop

# Test with Laravel  
go run scripts/test_folder.go /path/to/framework

# Test with Symfony
go run scripts/test_folder.go /path/to/symfony
```

## Performance

### Benchmarks
- **Simple Statements**: ~1.6μs per parse
- **Complex Expressions**: ~4μs per parse
- **Large Files**: Efficient memory usage with streaming
- **Concurrent Parsing**: Parser pool support for parallel processing

### Memory Usage
- **Low Footprint**: Minimal memory allocation per parse
- **Streaming**: Large file support without loading entire content
- **Pool Pattern**: Reusable parser instances for server applications

## Development

### Requirements
- Go 1.21+
- Optional: PHP 8.4 for compatibility testing

### Development Commands

```bash
# Build CLI tool
go build -o php-parser ./cmd/php-parser

# Run all tests
go test ./...

# Run performance benchmarks
go test ./parser -bench=. -run=^$

# Test with memory profiling
go test ./parser -bench=. -benchmem

# Compatibility testing
go run scripts/test_folder.go /path/to/php/codebase

# Clean build artifacts
go clean
rm -f php-parser main
```

## Contributing

We welcome contributions! Please see our contributing guidelines:

1. **Fork** the repository
2. **Create** a feature branch: `git checkout -b feature-name`
3. **Write** tests for your changes
4. **Ensure** all tests pass: `go test ./...`
5. **Run** linting: `go fmt ./...`
6. **Submit** a pull request

### Development Guidelines
- Follow Go coding standards (`gofmt`, effective Go)
- Maintain PHP compatibility with official implementation
- Add comprehensive tests for new features
- Reference PHP's official grammar at `/php-src/Zend/zend_language_parser.y`
- Update documentation for new features

## Use Cases in Production

### Static Analysis Tools
- **Code Quality**: Detect anti-patterns, complexity metrics
- **Security Scanning**: Find vulnerabilities, injection risks
- **Style Checking**: Enforce coding standards and conventions

### Development Tools
- **IDEs**: Syntax highlighting, auto-completion, error detection
- **Language Servers**: LSP implementation for editor support  
- **Code Formatters**: Consistent code styling and formatting

### Documentation Tools
- **API Generators**: Extract documentation from code and comments
- **Code Visualization**: Generate class diagrams, call graphs
- **Dependency Analysis**: Track code relationships and coupling

### Migration and Refactoring
- **Version Upgrades**: PHP version compatibility transformations
- **Framework Migrations**: Automated code pattern updates
- **Code Modernization**: Apply modern PHP practices and syntax

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- **PHP Team**: For the official PHP language specification  
- **Go Community**: For excellent tooling and ecosystem
- **Contributors**: Everyone who has contributed to this project

## Links

- **Repository**: [github.com/wudi/php-parser](https://github.com/wudi/php-parser)
- **Issues**: [Report bugs and feature requests](https://github.com/wudi/php-parser/issues)
- **Documentation**: [Detailed API documentation](docs/)
- **Examples**: [Code examples and tutorials](examples/)

---

**Built with ❤️ in Go | PHP 8.4 Compatible | Production Ready**
