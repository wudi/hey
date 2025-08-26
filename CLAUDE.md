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
go test ./...                    # Run all tests
go test ./lexer -v               # Run lexer tests with verbose output
go test ./parser -v              # Run parser tests with verbose output
go test ./ast -v                 # Run AST tests with verbose output
go test ./parser -bench=.        # Run performance benchmarks
go test ./parser -bench=. -run=^$  # Run only benchmarks (no unit tests)
go test ./parser -run=TestParsing_NowdocStrings  # Run specific test
go test ./parser -run=TestParsing_ClassMethodsWithVisibility  # Run class methods test
go test ./parser -run=TestParsing_TryCatchWithStatements  # Run try-catch parsing tests
go test ./parser -run=TestParsing_TypedReferenceParameters  # Run typed reference parameter tests
```

**Build CLI Tool:**
```bash
go build -o php-parser ./cmd/php-parser  # Build command-line tool
./php-parser -i example.php              # Parse a PHP file
./php-parser -tokens -ast                # Show tokens and AST structure
./php-parser -format json example.php    # Output AST as JSON
./php-parser -errors example.php         # Show only parsing errors
echo '<?php echo "Hello"; ?>' | ./php-parser  # Parse from stdin
```

**PHP Compatibility Testing:**
```bash
php -r "var_export(token_get_all('<?php \$x = 1; ?>'));"  # Compare with PHP tokens
/bin/php test_shebang.php                                # Test shebang handling
go run test_shebang_demo.go                             # Test lexer with shebang files
```

**Cleanup:**
```bash
go clean                         # Clean build artifacts
rm -f php-parser main              # Remove built binaries
```

## Architecture Overview

This is a PHP parser implementation in Go with the following structure:

### Core Modules
- **`lexer/`**: PHP lexical analyzer with state machine
  - `token.go`: PHP token definitions (150+ tokens matching PHP 8.4)
  - `states.go`: Lexer state management (11 states including ST_IN_SCRIPTING, ST_DOUBLE_QUOTES)
  - `lexer.go`: Main lexer implementation with shebang support and PHP tag recognition

- **`parser/`**: Recursive descent parser with Pratt parsing
  - `parser.go`: 2370+ lines implementing 40+ parse expression functions
  - Comprehensive PHP syntax support (variables, functions, classes, control flow)
  - Class method visibility parsing with public/private/protected modifiers
  - Operator precedence handling (LOWEST to HIGHEST including TERNARY)
  - Expression parsing: binary ops, unary ops, method calls, array access, etc.

- **`ast/`**: Abstract Syntax Tree nodes
  - `node.go`: Interface-based AST node system with visitor pattern
  - `kind.go`: AST node type constants (150+ kinds matching PHP's zend_ast.h)
  - `builder.go`: AST construction utilities
  - Full JSON serialization and string representation support

- **`errors/`**: Error handling with precise position tracking

### Command Line Interface  
- **`cmd/php-parser/`**: Feature-rich CLI tool (244 lines)
  - Multiple output formats: JSON, AST structure, raw tokens
  - File and stdin input with comprehensive error handling
  - Debugging flags: -tokens, -ast, -errors, -v (verbose)
  - Position-aware error reporting with line:column information

### Key Design Features
- **PHP Compatibility**: Token IDs match PHP 8.4 official implementation
- **Pratt Parser**: Elegant operator precedence handling with prefix/infix functions
- **State Machine**: Lexer supports 11 states including shebang recognition
- **Interface-Based AST**: Visitor pattern support with 150+ node types
- **Position Tracking**: Precise error location with line/column information
- **Performance**: Benchmarking support for parser optimization

### Critical Implementation Details

**Parser Architecture:**
- Prefix parse functions: handle tokens that start expressions (variables, literals, unary ops)
- Infix parse functions: handle binary operators, method calls, array access
- Precedence levels: LOWEST, EQUALS, LESSGREATER, SUM, PRODUCT, EXPONENT, PREFIX, CALL, INDEX, TERNARY, HIGHEST
- Error recovery: continues parsing after errors to find multiple issues

**Lexer States:**
- ST_IN_SCRIPTING: Main PHP code parsing
- ST_DOUBLE_QUOTES: String interpolation with variable recognition  
- ST_HEREDOC/ST_NOWDOC: Multi-line string handling
- ST_LOOKING_FOR_PROPERTY: Object member access

**AST Node System:**
- All nodes implement Node interface with GetChildren(), Accept(), String() methods
- AST kinds match PHP's official zend_ast.h constants (ASTVar = 256, ASTCall = 516, etc.)
- Class constants use ASTClassConstGroup (776) for declarations and ASTConstElem (775) for individual constants
- FunctionDeclaration includes Visibility field for class method access modifiers
- PropertyDeclaration supports typed properties with visibility modifiers
- JSON serialization preserves full AST structure for external tools

### Testing Strategy
- **Unit Tests**: Comprehensive table-driven tests using testify framework (180+ tests)
  - 40+ parser tests covering variables, expressions, arrays, functions, classes, class constants
  - Class method visibility parsing tests (TestParsing_ClassMethodsWithVisibility)
  - Property declaration tests with type hints and visibility modifiers
  - Try-catch statement parsing with subsequent statements (TestParsing_TryCatchWithStatements)
  - Typed reference parameter parsing tests (TestParsing_TypedReferenceParameters)
  - Heredoc/Nowdoc string parsing tests
  - String interpolation and complex expression tests
  - Class constant parsing with all visibility modifiers
  - Error handling and syntax error recovery tests
- **Benchmark Tests**: Performance testing for different complexity levels
  - Simple assignments (~1.6μs), complex expressions (~4μs)  
  - String parsing benchmarks (simple strings, interpolation, heredoc/nowdoc)
  - Memory allocation tracking with `-benchmem`
- **PHP Compatibility**: Validation using `/bin/php` and `token_get_all()`
- **Shebang Support**: Tests for executable PHP files with shebang headers
- **Edge Cases**: Malformed syntax, error recovery, and boundary conditions

## Important Reminders

**When Adding New PHP Syntax Support:**
1. Add new token types to `lexer/token.go` if needed (maintain PHP compatibility)
2. Implement prefix or infix parse functions in `parser/parser.go`
3. Create corresponding AST node types in `ast/node.go` with full interface implementation
4. Add AST kind constants to `ast/kind.go` (follow PHP's zend_ast.h numbering)
5. Update the String() method in `ast/kind.go` for new node types
6. Add constructor functions (NewXXXExpression) following existing patterns

**When Adding New Class Member Types:**
1. Analyze PHP grammar rules in `/home/ubuntu/php-src/Zend/zend_language_parser.y`
2. Check if visibility modifiers are supported for the new member type
3. Update `parseClassStatement` logic at `parser.go:2117` to handle the new case
4. Create dedicated parsing function (e.g., `parseClassConstantDeclaration`)
5. Add comprehensive test cases covering all visibility modifiers and edge cases
6. Ensure AST nodes implement full Node interface (GetChildren, Accept, String)

**Parser Error Debugging:**
- "no prefix parse function found" → add prefix parse function to parser initialization in `parser.go:80-100`
- "no infix parse function found" → add infix parse function with correct precedence in `parser.go:100-120`
- "expected next token to be T_VARIABLE, got T_STRING instead" for class constants → check `parseClassStatement` logic at `parser.go:2117`
- Class method visibility parsing issues → verify `parseFunctionDeclaration` handles visibility at `parser.go:608-614`
- Property parsing with visibility modifiers → check `parsePropertyDeclaration` function
- Try-catch parsing with statements after → verify token advancement in `parseTryStatement` at `parser.go:1492-1503`
- Missing AST constructors → add NewXXXExpression functions in `ast/node.go`
- Nowdoc/Heredoc parsing issues → check `parseNowdocExpression` and `parseHeredoc` functions
- String interpolation problems → verify `InterpolatedStringExpression` handling
- Class constant parsing errors → verify `parseClassConstantDeclaration` function at `parser.go:2139`

**PHP Compatibility Requirements:**
- Token IDs must match PHP 8.4 official implementation exactly
- AST node kinds should align with zend_ast.h when possible
- Test against `/bin/php` using `token_get_all()` for validation
- Reference `/home/ubuntu/php-src` for implementation details
- Reference `/home/ubuntu/php-src/Zend/zend_language_parser.y` for grammar rules
- Reference `/home/ubuntu/php-src/Zend/zend_ast.h` for AST node kinds
- Reference `/home/ubuntu/php-src/Zend/zend_language_scanner.l` for lexer and lexer states and tokenization
- Before performing any fixes or refactoring, analyze the original PHP code's lexical and syntactic structure first.

## Recent Improvements

**Class Method Visibility Parsing (Latest):**
- Enhanced `FunctionDeclaration` AST node with `Visibility` field at `ast/node.go:890`
- Updated `parseFunctionDeclaration` to handle visibility modifiers at `parser.go:608-614`
- Fixed `parseClassStatement` to properly route visibility + function combinations at `parser.go:2123-2124`
- Support for all visibility modifiers: `private function`, `protected function`, `public function`
- Methods without explicit visibility leave Visibility field empty (following PHP defaults)
- Comprehensive test suite with 3 test scenarios (`TestParsing_ClassMethodsWithVisibility`)
  - Public constructor with complex parameter type hints
  - All three visibility modifiers with different parameter types
  - Default methods without visibility modifier

**Class Property Declaration Support:**
- Added `PropertyDeclaration` AST node with visibility, type hints, and default values
- Full class declaration parsing with extends/implements support 
- Enhanced static access operator (::) as infix operator for complex expressions
- Support for typed properties: `private bool $prop`, `private array $data = []`

**Class Constants Parsing:**
- Added `ClassConstantDeclaration` and `ConstantDeclarator` AST nodes at `ast/node.go:3047-3145`
- Implemented `parseClassConstantDeclaration` function at `parser.go:2139-2186`
- Fixed parsing logic in `parseClassStatement` to handle visibility modifiers followed by `const`
- Support for all visibility modifiers: `private const`, `protected const`, `public const`
- Support for multiple constants per declaration: `const A = 1, B = 2, C = "hello"`
- Comprehensive test suite with 5 test cases covering all scenarios (`TestParsing_ClassConstants`)

**Nowdoc/Heredoc Parsing:**
- Fixed multi-token nowdoc parsing in `parseNowdocExpression()` at `parser.go:1747`
- Added support for variables in heredoc content in `parseHeredoc()` at `parser.go:930`  
- Comprehensive test coverage for both nowdoc and heredoc scenarios

**Function Declaration Refactoring:**
- Major refactoring based on PHP's official zend_language_parser.y grammar
- Enhanced AST structure with structured TypeHint objects instead of simple strings
- Support for nullable types (?Type), union types (Type1|Type2), intersection types (Type1&Type2)
- Function by-reference support (function &foo()) and variadic parameters (...$params)
- Parameter visibility modifiers (public, private, protected, readonly)
- Fixed T_ELLIPSIS (...) token generation for variadic parameters

**Try-Catch Statement Parsing Fix (Latest):**
- Fixed critical parsing issue where statements following try-catch blocks failed to parse
- Root cause: Improper token advancement in catch clause loop causing "no prefix parse function for `=`" errors
- Fixed token positioning logic in `parseTryStatement` at `parser.go:1492-1503`
- Added comprehensive test suite `TestParsing_TryCatchWithStatements` with 5 scenarios:
  - Basic try-catch with assignment after
  - Try-catch with statements in blocks and after
  - Multiple catch clauses with statements after
  - Nested try-catch structures with complex expressions
  - Empty try-catch followed by multiple statements
- Enhanced finally block detection and parsing
- Full support for complex expressions after try-catch: `$obj->method1()->method2()->method3()`

**Typed Reference Parameters Support:**
- Enhanced `FunctionDeclaration` to support typed reference parameters like `function foo(array &$data)`
- Added comprehensive parsing for nullable, union, and intersection types with reference
- Support for mixed reference and non-reference parameters in same function
- Extensive test coverage with `TestParsing_TypedReferenceParameters`

**Enhanced Test Suite:**
- Added table-driven tests for string interpolation scenarios
- Class method visibility test scenarios with comprehensive parameter validation
- Property declaration tests with all visibility combinations
- Try-catch parsing tests with edge cases and complex scenarios
- Benchmark tests for different parsing complexity levels
- Error case testing with proper validation
- All tests follow Go testing best practices with testify framework

## Statement Parsing Token Management

When fixing statement parsing issues, particularly for control structures (try-catch, if-else, loops), pay careful attention to token advancement:

1. **Token Position Consistency**: After parsing a statement, ensure `currentToken` is positioned correctly for the main parsing loop
2. **Loop Token Advancement**: In loops that parse multiple similar constructs (like catch clauses), only advance tokens when certain there are more constructs to parse
3. **Finally Block Detection**: Use `peekToken` to check for optional constructs like finally blocks without advancing prematurely
4. **Break vs Continue**: Use proper loop control to avoid unnecessary token advancement that can skip subsequent statements

**Common Pattern for Multi-Construct Parsing:**
```go
for p.currentToken.Type == EXPECTED_TOKEN {
    // Parse construct
    // ...
    
    // Only advance if there's another construct
    if p.peekToken.Type == EXPECTED_TOKEN {
        p.nextToken()
    } else {
        break
    }
}
```

# important-instruction-reminders
Do what has been asked; nothing more, nothing less.
NEVER create files unless they're absolutely necessary for achieving your goal.
ALWAYS prefer editing an existing file to creating a new one.
NEVER proactively create documentation files (*.md) or README files. Only create documentation files if explicitly requested by the User.
- to memorize