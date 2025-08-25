# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

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