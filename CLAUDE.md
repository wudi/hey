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
go test ./parser -bench=. -benchmem # Run benchmarks with memory allocation stats
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

**WordPress Compatibility Testing:**
```bash
go run test_wordpress.go                 # Run WordPress parsing test suite (1,648 files)
```

# Code style
- Follow Go conventions (gofmt, effective Go)
- Use interfaces and structs for abstraction
- Lexer test case write into `lexer/lexer_test.go` 
- Parser test case write into `parser/parser_test.go`
- AST test case write into `ast/ast_test.go`
- Use testify framework for assertions in tests

**Workflow**
- Plan todo list before coding using TodoWrite tool and update frequently
- Before implementing new syntax, analyze PHP grammar from `/home/ubuntu/php-src/Zend/zend_language_parser.y` 
- Always add test cases first, then implement functionality (TDD approach)
- After adding new syntax support, create test files in `/tmp/test_*.php` to verify functionality
- Use CLI tool to verify parsing: `./php-parser test_file.php`
- Run targeted tests: `go test ./parser -run=TestSpecificFeature -v`
- After fixing a bug, run all tests to ensure no regressions: `go test ./...`
- For major changes, run benchmarks: `go test ./parser -bench=. -run=^$`
- Test WordPress compatibility after major changes: `go run test_wordpress.go`
- Create descriptive commit messages with detailed changelog

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
  - `lexer_test.go`: Comprehensive token generation tests

- **`parser/`**: Recursive descent parser with Pratt parsing
  - `parser.go`: 4000+ lines implementing 50+ parse expression functions
  - `parser_test.go`: 180+ test cases covering all PHP syntax
  - `benchmark_test.go`: Performance benchmarks for parser optimization
  - `pool.go`: Parser pool for concurrent parsing scenarios
  - Comprehensive PHP 8.4 syntax support (variables, functions, classes, control flow)
  - Complete operator support including assignment operators (??=, **=, &=, |=, ^=, <<=, >>=), power (**), spaceship (<=>), unary plus (+)
  - Alternative syntax support for all control structures (if/endif, while/endwhile, for/endfor, foreach/endforeach, switch/endswitch)
  - Special statement handling (__halt_compiler()) with proper parsing termination
  - Class method visibility parsing with public/private/protected modifiers
  - Enhanced operator precedence with 14 levels matching PHP 8.4 specification
  - Expression parsing: binary ops, unary ops, method calls, array access, match expressions
  - Namespaced class support: `class A implements B\C\D`, `class A extends B\C\D`

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
- **PHP 8.4 Compatibility**: Token IDs and grammar rules match PHP 8.4 official implementation exactly
- **Complete Operator Support**: All PHP operators including modern additions (**, <=>, ??=, etc.)
- **Alternative Syntax**: Full support for alternative control structure syntax (switch:...endswitch;, etc.)
- **Special Constructs**: __halt_compiler() with proper parsing termination, match expressions
- **Pratt Parser**: Elegant operator precedence handling with 14 distinct precedence levels
- **State Machine**: Lexer supports 11 states including shebang recognition and string interpolation
- **Interface-Based AST**: Visitor pattern support with 150+ node types matching zend_ast.h
- **Position Tracking**: Precise error location with line/column information
- **Performance**: Benchmarking support for parser optimization

### Critical Implementation Details

**Parser Architecture:**
- Prefix parse functions: handle tokens that start expressions (variables, literals, unary ops)
- Infix parse functions: handle binary operators, method calls, array access
- Precedence levels (PHP 8.4 compliant): ASSIGN, TERNARY, COALESCE, LOGICAL_OR, LOGICAL_AND, BITWISE_OR, BITWISE_XOR, BITWISE_AND, EQUALS, LESSGREATER, BITWISE_SHIFT, SUM, PRODUCT, EXPONENT, PREFIX, POSTFIX, CALL, INDEX
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
1. Add new token types to `lexer/token.go` if needed (maintain PHP compatibility) - check Keywords map for reserved words
2. Add prefix/infix parse functions in `parser/parser.go` (globalPrefixParseFns/globalInfixParseFns maps at lines 87-272)
3. Set correct operator precedence in precedences map (follow PHP 8.4 precedence levels)
4. Create corresponding AST node types in `ast/node.go` with full interface implementation
5. Add AST kind constants to `ast/kind.go` (follow PHP's zend_ast.h numbering)
6. Update the String() method in `ast/kind.go` for new node types
7. Add constructor functions (NewXXXExpression) following existing patterns
8. Update `ast/visitor.go` Walk function to handle new node types (lines 28-95)
9. Add comprehensive test cases covering all syntax variations and edge cases
10. Test with WordPress codebase using `go run test_wordpress.go`

**When Adding New Class Member Types:**
1. Analyze PHP grammar rules in `/home/ubuntu/php-src/Zend/zend_language_parser.y`
2. Check if visibility modifiers are supported for the new member type
3. Update `parseClassStatement` logic at `parser.go:2117` to handle the new case
4. Create dedicated parsing function (e.g., `parseClassConstantDeclaration`)
5. Add comprehensive test cases covering all visibility modifiers and edge cases
6. Ensure AST nodes implement full Node interface (GetChildren, Accept, String)

**Parser Error Debugging:**
- "no prefix parse function found" → add prefix parse function to parser initialization in `parser.go:87-204`
- "no infix parse function found" → add infix parse function with correct precedence in `parser.go:205-272`
- "expected next token to be T_VARIABLE, got T_STRING instead" for class constants → check `parseClassStatement` logic at `parser.go:2117`
- Class method visibility parsing issues → verify `parseFunctionDeclaration` handles visibility at `parser.go:608-614`
- Property parsing with visibility modifiers → check `parsePropertyDeclaration` function
- Try-catch parsing with statements after → verify token advancement in `parseTryStatement` at `parser.go:1492-1503`
- Missing AST constructors → add NewXXXExpression functions in `ast/node.go`
- Nowdoc/Heredoc parsing issues → check `parseNowdocExpression` and `parseHeredoc` functions
- String interpolation problems → verify `InterpolatedStringExpression` handling
- Class constant parsing errors → verify `parseClassConstantDeclaration` function at `parser.go:2139`
- Namespace parsing in class declarations → check `parseClassName` function for qualified name support
- Duplicate class declaration functions → both `parseClassDeclaration` and `parseReadonlyClassDeclaration` need updates

**PHP Compatibility Requirements:**
- Token IDs must match PHP 8.4 official implementation exactly
- AST node kinds should align with zend_ast.h when possible
- Test against `/bin/php` using `token_get_all()` for validation
- Reference `/home/ubuntu/php-src` for implementation details
- Reference `/home/ubuntu/php-src/Zend/zend_language_parser.y` for grammar rules
- Reference `/home/ubuntu/php-src/Zend/zend_ast.h` for AST node kinds
- Reference `/home/ubuntu/php-src/Zend/zend_language_scanner.l` for lexer and lexer states and tokenization
- Before performing any fixes or refactoring, analyze the original PHP code's lexical and syntactic structure first.

## Critical Parser Functions

**Qualified Name Parsing:**
- `parseClassName()` at `parser.go:3914` - handles namespaced class names (A\B\C)
- Used by both `parseClassDeclaration` and `parseReadonlyClassDeclaration`
- Essential for extends and implements clauses with namespaces

**Class Declaration Functions:**
- `parseClassDeclaration()` at `parser.go:3858` - regular class declarations
- `parseReadonlyClassDeclaration()` at `parser.go:4004` - readonly class declarations
- Both must be updated in sync when fixing class-related parsing issues

**Operator Parsing:**
- Prefix operators registered in `globalPrefixParseFns` (lines 87-204)
- Infix operators registered in `globalInfixParseFns` (lines 205-272)
- Precedence levels defined in `precedences` map (lines 273-340)

## Recent Improvements

**100% WordPress Compatibility (Latest - 2024-12-28):**
- Fixed namespace parsing in implements clauses (`class A implements B\C\D`)
- Added unary plus operator support (`+(expression)`)
- Successfully parses all 1,648 PHP files in WordPress 6.7.1
- Both `parseClassDeclaration` and `parseReadonlyClassDeclaration` updated

**PHP 8.4 Grammar Enhancements (2024-12-19):**
- **Complete Operator Support**: Added all missing assignment operators (??=, **=, &=, |=, ^=, <<=, >>=), power operator (**), spaceship operator (<=>), logical XOR (xor)
- **Alternative Syntax**: Full support for alternative switch syntax (switch: ... endswitch;) alongside existing if/endif, while/endwhile, for/endfor, foreach/endforeach
- **Special Statements**: __halt_compiler() statement with proper parsing termination behavior
- **Enhanced Precedence**: Updated operator precedence to exactly match PHP 8.4 official specification with 14 distinct levels
- **AST Improvements**: Added HaltCompilerStatement AST node, improved error handling and position tracking
- **Comprehensive Testing**: Added test coverage for new operators, match expressions, alternative syntax, and halt compiler functionality

**Class Method Visibility Parsing:**
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