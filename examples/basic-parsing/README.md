# Basic Parsing Example

This example demonstrates how to use the PHP parser to parse PHP code and examine the resulting AST.

## What it does

- Takes a sample PHP code string containing variables, functions, and classes
- Creates a lexer and parser instance
- Parses the code into an Abstract Syntax Tree (AST)
- Displays basic information about the parsed statements
- Shows the full AST structure

## Running the example

```bash
cd examples/basic-parsing
go run main.go
```

## Expected output

The program will display:
- Number of parsed statements
- A list of each statement type
- The complete AST structure in string format

This is the foundation for more complex PHP code analysis tasks.