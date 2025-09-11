# Hey - PHP Interpreter in Go

**Currently experimental, do not use in production.**

A high-performance PHP interpreter written in Go, providing syntax compatibility with PHP 8.0+.

## Features

- **Full PHP 8.0+ Syntax Support**: Compatible with modern PHP features including arrow functions, spread operators, goto statements, and strict types
- **High-Performance Virtual Machine**: Custom bytecode compiler and VM with advanced profiling capabilities
- **Advanced Debugging**: Built-in debugger with breakpoints, variable watching, and performance analysis
- **Memory Management**: Efficient memory pool with allocation tracking
- **Lexer & Parser**: Complete lexical analysis and parsing for PHP syntax
- **Static Analysis**: AST-based code analysis and metrics collection

## Installation Binary

```bash
go install github.com/wudi/hey/cmd/hey@master
```

```
hey -h

NAME:
   hey - A PHP interpreter written in Go

USAGE:
   hey [global options] [command [command options]]

COMMANDS:
   init        Creates a composer.json file in the current directory
   require     Adds required packages to your composer.json and installs them
   install, i  Installs your composer.json and installs them
   update, u   Updates your dependencies to the latest version according to composer.json, and updates the composer.lock file
   validate    Validates a composer.json file
   fpm         FastCGI process manager
   help, h     Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --code string, -r string     Run PHP <code> without using script tags <?..?>
   --version string, -v string  Show version
   --file string, -f string     Parse and execute <file>.
   -S string                    <addr>:<port> Run with built-in web server.
   --help, -h                   show help

```


## Using modules 

### Basic Usage 

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
    
    // Parse
    l := lexer.New(phpCode)
    p := parser.New(l)
    program := p.ParseProgram()
    
    // Compile
    comp := compiler.NewCompiler()
    comp.Compile(program)
    
    // Execute
    vmachine := vm.NewVirtualMachine()
    ctx := vm.NewExecutionContext()
    vmachine.Execute(ctx, comp.GetBytecode(), comp.GetConstants(), comp.GetFunctions(), comp.GetClasses())
}
```

### Demo Application

Run the included demo to see advanced features:

```bash
cd cmd/vm-demo
go run main.go
```

## Architecture

### Core Components

- **Lexer**: Tokenizes PHP source code
- **Parser**: Builds Abstract Syntax Tree (AST)
- **Compiler**: Generates bytecode from AST
- **Virtual Machine**: Executes bytecode with profiling support
- **Runtime**: Provides PHP standard library functions

### Performance Features

- **Profiling VM**: Detailed execution profiling and hot spot analysis
- **Memory Tracking**: Allocation and deallocation monitoring
- **Breakpoints**: Debug support with variable watching
- **Performance Reports**: Comprehensive execution statistics

## Examples

The `examples/` directory contains comprehensive demonstrations:

- **Basic Parsing**: Core parsing functionality
- **AST Visitor**: Tree traversal and analysis
- **Token Extraction**: Lexical analysis examples
- **Error Handling**: Error detection and recovery
- **Code Analysis**: Static analysis and metrics

## Supported PHP Features

- Variables and data types
- Functions (including arrow functions)
- Classes and objects
- Control structures (if/else, loops, goto)
- Modern PHP syntax (spread operators, strict types)
- Error handling and exceptions
- Standard library functions

## Development

### Building

```bash
go build ./cmd/vm-demo
```

### Testing

```bash
go test ./...
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

### Bytecode Compiler

Experience next-generation PHP execution with our complete bytecode compiler:

```bash
# Build the bytecode demo
go build -o bytecode-demo ./cmd/bytecode-demo

# Run bytecode compilation examples
./bytecode-demo
```

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Author

Di Wu <hi@wudi.io>

## Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.
