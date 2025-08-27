# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

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

# Code style
- Follow Go conventions (gofmt, effective Go)
- Use interfaces and structs for abstraction
- Lexer test case write into `lexer/lexer_test.go` 
- Parser test case write into `parser/parser_test.go`
- AST test case write into `ast/ast_test.go`

**Workflow**
- Plan todo list before coding in `TODO.md` and update it frequently
- Before fix syntax error, ultrathink top-down analysis of the original PHP syntax tree
- Be sure to unit test when you’re done making a series of code changes
- After fixing a bug, add a test case to cover the bug
- After fixing a bug, run all tests to ensure no regressions
- After code changes, consider globally whether there are better implementations, or even refactorings
- Prefer running single tests, and not the whole test suite, for performance
- Create a descriptive commit message

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