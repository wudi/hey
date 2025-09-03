# AST Visitor Example

This example demonstrates how to use the visitor pattern to traverse and analyze the Abstract Syntax Tree (AST).

## What it does

- Shows how to implement custom visitors to collect specific information from the AST
- Demonstrates the use of built-in visitor functions for common tasks
- Provides examples of different analysis patterns:
  - Variable collection with duplicate removal
  - Function declaration analysis
  - AST depth calculation
  - Node type statistics

## Key concepts

### Custom Visitors
- `VariableCollector`: Finds all variable usage in the code
- `FunctionCollector`: Collects function declarations with parameter counts  
- `DepthCounter`: Calculates the maximum depth of the AST

### Built-in Visitor Functions
- `ast.FindAllFunc()`: Find all nodes matching a condition
- `ast.CountFunc()`: Count nodes matching a condition
- `ast.WalkFunc()`: Simple traversal with a function

## Running the example

```bash
cd examples/ast-visitor
go run main.go
```

## Expected output

The program will display:
- List of all unique variables found in the code
- Information about each function declaration
- Maximum depth of the AST
- Statistics about different node types
- Count of various expression types

This pattern is useful for building code analysis tools, linters, and documentation generators.