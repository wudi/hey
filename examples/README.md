# PHP Parser Examples

This directory contains practical examples demonstrating various use cases and capabilities of the PHP parser.

## Available Examples

### 1. [Basic Parsing](./basic-parsing/)
**Difficulty**: Beginner  
**Focus**: Core parsing functionality

Learn the fundamentals of parsing PHP code into an Abstract Syntax Tree (AST). This example shows how to:
- Create a lexer and parser instance
- Parse PHP code and handle basic errors
- Examine the resulting AST structure
- Display parsed statements and their types

**Key concepts**: Lexer, Parser, AST, Program structure

---

### 2. [AST Visitor](./ast-visitor/)  
**Difficulty**: Intermediate  
**Focus**: Tree traversal and analysis

Demonstrates the visitor pattern for traversing and analyzing AST nodes. Includes:
- Custom visitor implementations for specific analysis tasks
- Built-in visitor functions for common operations
- Variable collection and function analysis
- Node type statistics and pattern detection

**Key concepts**: Visitor pattern, AST traversal, Node filtering

---

### 3. [Token Extraction](./token-extraction/)
**Difficulty**: Beginner  
**Focus**: Lexical analysis

Shows how to extract and analyze individual tokens from PHP source code:
- Token type identification and categorization
- Statistical analysis of token usage
- Content extraction (keywords, identifiers, literals)
- String interpolation tokenization examples

**Key concepts**: Lexer, Tokens, Token types, Lexical analysis

---

### 4. [Error Handling](./error-handling/)
**Difficulty**: Intermediate  
**Focus**: Error detection and reporting

Comprehensive error handling demonstration covering:
- Syntax error detection and categorization
- Error recovery and partial parsing
- Custom error reporting systems
- Common PHP syntax error scenarios

**Key concepts**: Error recovery, Error categorization, Partial parsing

---

### 5. [Code Analysis](./code-analysis/)
**Difficulty**: Advanced  
**Focus**: Static analysis and metrics

Advanced static code analysis example featuring:
- Code complexity metrics and quality assessment
- Pattern detection and issue identification  
- Comprehensive reporting with quality scores
- Practical suggestions for code improvement

**Key concepts**: Static analysis, Code metrics, Quality assessment

## Getting Started

Each example is self-contained and can be run independently:

```bash
cd examples/[example-name]
go run main.go
```

## Example Dependencies

All examples depend on the core PHP parser modules:
- `github.com/wudi/hey/compiler/lexer` - Lexical analysis
- `github.com/wudi/hey/compiler/parser` - Syntax analysis  
- `github.com/wudi/hey/compiler/ast` - Abstract Syntax Tree

## Learning Path

**Recommended order for beginners:**
1. **Basic Parsing** - Understand core concepts
2. **Token Extraction** - Learn about lexical analysis
3. **AST Visitor** - Master tree traversal
4. **Error Handling** - Handle real-world scenarios
5. **Code Analysis** - Build advanced analysis tools

## Use Case Scenarios

### Development Tools
- **Code Editors/IDEs**: Syntax highlighting, auto-completion
- **Linters**: Code quality and style checking
- **Formatters**: Automated code formatting
- **Refactoring Tools**: Safe code transformation

### Analysis Tools  
- **Static Analyzers**: Bug detection, security scanning
- **Metrics Tools**: Complexity analysis, technical debt
- **Documentation Generators**: API docs from code
- **Dependency Analyzers**: Code relationship mapping

### Language Tools
- **Transpilers**: PHP version compatibility
- **Code Generators**: Template-based code generation  
- **Migration Tools**: Framework/library migrations
- **Testing Tools**: Test generation and analysis

## Contributing Examples

When adding new examples:
1. Create a dedicated directory: `examples/example-name/`
2. Include `main.go` with comprehensive demonstrations
3. Add a detailed `README.md` explaining the concepts
4. Use realistic PHP code samples
5. Include error cases and edge scenarios
6. Follow the established documentation structure

## Common Patterns

### Error-Safe Parsing
```go
l := lexer.New(phpCode)
p := parser.New(l)
program := p.ParseProgram()

if len(p.Errors()) > 0 {
    // Handle parsing errors
    for _, err := range p.Errors() {
        fmt.Printf("Error: %s\n", err)
    }
}
```

### AST Traversal
```go
ast.Walk(visitor, program)
// or
results := ast.FindAllFunc(program, condition)
```

### Token Analysis
```go
l := lexer.New(code)
for {
    tok := l.NextToken()
    if tok.Type == token.T_EOF {
        break
    }
    // Process token
}
```

## Further Reading

- [Parser Implementation Details](../compiler/parser/)
- [AST Node Reference](../compiler/ast/)
- [Lexer Token Types](../compiler/lexer/)
- [PHP Language Grammar](https://github.com/php/php-src)