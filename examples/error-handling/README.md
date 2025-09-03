# Error Handling Example

This example demonstrates comprehensive error handling capabilities of the PHP parser, including error detection, categorization, and reporting.

## What it does

- Shows how to handle different types of parsing errors
- Demonstrates error categorization (syntax, lexical, semantic)
- Provides examples of common PHP syntax errors
- Shows parser error recovery capabilities
- Implements custom error reporting

## Key Features

### Error Types
- **Syntax Errors**: Malformed PHP syntax (missing semicolons, unmatched brackets)
- **Lexical Errors**: Invalid tokens (malformed strings, invalid characters)
- **Semantic Errors**: Valid syntax but semantic issues (basic checks only)

### Error Examples Covered
1. **Valid Code**: Demonstrates successful parsing
2. **Missing Semicolon**: Common syntax error
3. **Unmatched Parentheses**: Bracket/brace mismatches
4. **Invalid Variable Names**: Lexical token errors
5. **Invalid Function Syntax**: Structural syntax errors
6. **Complex Multiple Errors**: Real-world scenarios with multiple issues
7. **String-related Errors**: Unclosed strings and quote mismatches
8. **Error Recovery**: Parser's ability to continue after errors

### Custom Error Reporter
- Categorizes errors by type
- Counts total errors
- Provides detailed error reports
- Shows partial parsing results

## Running the example

```bash
cd examples/error-handling
go run main.go
```

## Expected output

For each test case, the program displays:
- The PHP code being tested
- Categorized error reports
- Success/failure status
- AST generation results (when possible)

## Error Recovery

The parser demonstrates error recovery by:
- Continuing to parse after encountering errors
- Generating partial AST when possible
- Collecting multiple errors in a single pass
- Providing meaningful error messages

## Production Use Cases

This error handling pattern is essential for:
- **IDE Integration**: Real-time syntax error highlighting
- **Code Linters**: Comprehensive code quality checking
- **CI/CD Pipelines**: Automated code validation
- **Development Tools**: Error reporting and suggestions
- **Language Servers**: Editor support with error diagnostics

## Extending Error Handling

To enhance error handling:
- Add line/column position information
- Implement error recovery strategies
- Provide fix suggestions
- Add more semantic analysis rules
- Support error severity levels