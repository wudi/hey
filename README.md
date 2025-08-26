# PHP Parser

A comprehensive PHP parser written in Go that converts PHP source code into Abstract Syntax Trees (AST). The parser supports PHP 7+ syntax and maintains high compatibility with PHP's official implementation.

## Features

- **Complete PHP Lexer**: 150+ token types matching PHP 8.4 specification
- **Recursive Descent Parser**: Pratt parsing with proper operator precedence
- **Full AST Support**: Interface-based AST nodes with visitor pattern
- **PHP Compatibility**: Token IDs and AST structure align with PHP's official implementation
- **Comprehensive Syntax Support**:
  - Variables, functions, classes, and control flow
  - Class constants, properties, and methods with visibility modifiers
  - Try-catch-finally statements
  - Heredoc/Nowdoc strings with interpolation
  - Typed parameters and reference parameters
  - Complex expressions and method chaining

## Installation

```bash
git clone https://github.com/wudi/php-parser.git
cd php-parser
go build -o php-parser ./cmd/php-parser
```

## Usage

### Command Line Interface

```bash
# Parse PHP code from stdin
echo '<?php echo "Hello, World!"; ?>' | ./php-parser

# Parse a PHP file
./php-parser -i example.php

# Show tokens and AST structure
./php-parser -tokens -ast example.php

# Output formats: json (default), ast, tokens
./php-parser -format json example.php

# Show only parsing errors
./php-parser -errors example.php
```

### Programmatic Usage

```go
package main

import (
    "fmt"
    "github.com/wudi/php-parser/lexer"
    "github.com/wudi/php-parser/parser"
)

func main() {
    input := `<?php
    $name = "World";
    echo "Hello, " . $name;
    ?>`
    
    l := lexer.New(input)
    p := parser.New(l)
    program := p.ParseProgram()
    
    // Check for parsing errors
    if errors := p.Errors(); len(errors) > 0 {
        for _, err := range errors {
            fmt.Printf("Parser error: %s\n", err)
        }
        return
    }
    
    // Use the AST
    fmt.Printf("Parsed %d statements\n", len(program.Body))
}
```

## Architecture

The parser follows a classic compiler frontend design:

```
PHP Source Code → Lexer → Token Stream → Parser → Abstract Syntax Tree
```

### Core Modules

- **`lexer/`**: Lexical analyzer with state machine (11 states)
- **`parser/`**: Recursive descent parser with Pratt parsing
- **`ast/`**: AST node definitions and utilities
- **`cmd/php-parser/`**: Command-line interface
- **`errors/`**: Error handling with position tracking

## Testing

```bash
# Run all tests
go test ./...

# Run parser tests with verbose output
go test ./parser -v

# Run specific test suites
go test ./parser -run=TestParsing_TryCatchWithStatements
go test ./parser -run=TestParsing_ClassMethodsWithVisibility

# Run benchmarks
go test ./parser -bench=.
```

## PHP Compatibility

The parser maintains high compatibility with PHP's official implementation:

- Token IDs match PHP 8.4 specification
- AST node kinds align with PHP's `zend_ast.h`
- Lexer states mirror PHP's lexical analyzer
- Comprehensive test suite validates against PHP's `token_get_all()`

## Supported PHP Syntax

### Basic Constructs
- Variables and constants
- All data types (integers, floats, strings, arrays, objects)
- Operators (arithmetic, comparison, logical, bitwise)

### Control Flow
- If/else statements
- Switch statements
- Loops (for, foreach, while, do-while)
- Try-catch-finally blocks

### Object-Oriented Features
- Class declarations with inheritance
- Properties with visibility modifiers and type hints
- Methods with visibility modifiers
- Class constants with visibility modifiers
- Static access and method calls

### Functions
- Function declarations and calls
- Anonymous functions (closures)
- Typed parameters and return types
- Reference parameters
- Variadic parameters

### Advanced Features
- Heredoc and Nowdoc strings
- String interpolation
- Array syntax (both `array()` and `[]`)
- Complex expressions and method chaining

## Examples

### Basic Parsing

```php
<?php
$users = [
    ['name' => 'John', 'age' => 30],
    ['name' => 'Jane', 'age' => 25]
];

foreach ($users as $user) {
    echo $user['name'] . " is " . $user['age'] . " years old\n";
}
```

### Class with Methods

```php
<?php
class UserManager {
    private array $users = [];
    
    public function addUser(string $name, int $age): void {
        $this->users[] = ['name' => $name, 'age' => $age];
    }
    
    protected function getUserCount(): int {
        return count($this->users);
    }
}
```

### Try-Catch with Complex Expressions

```php
<?php
try {
    $result = $service->processData($input);
    $processed = $result->transform()->validate();
} catch (ValidationException $e) {
    $logger->error($e->getMessage());
    throw new ProcessingException('Validation failed');
} finally {
    $cleanup->execute();
}

$finalResult = $processor->complete($processed);
```

## Development

### Requirements
- Go 1.21+
- PHP 8.4 (for compatibility testing)

### Running Tests
```bash
# All tests
go test ./...

# Specific modules
go test ./lexer -v
go test ./parser -v
go test ./ast -v

# With benchmarks
go test ./parser -bench=. -benchmem
```

### Contributing
1. Follow Go coding standards
2. Maintain PHP compatibility
3. Add comprehensive tests for new features
4. Reference PHP's official grammar (`/home/ubuntu/php-src/Zend/zend_language_parser.y`)

## License

This project is open source. Please check the repository for license details.