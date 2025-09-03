# Token Extraction Example

This example demonstrates how to use the lexer to extract and analyze tokens from PHP code.

## What it does

- Extracts all tokens from PHP code using the lexer
- Analyzes token statistics and distributions
- Demonstrates different token types and their properties
- Shows how to filter and categorize tokens
- Includes examples of string interpolation tokenization

## Key Features

### Token Analysis
- **TokenAnalyzer**: Collects statistics about token usage
- **Token Type Counting**: Tracks frequency of each token type
- **Content Extraction**: Separates keywords, identifiers, strings, and numbers
- **Position Tracking**: Shows line and column information for each token

### Token Categories
- **Keywords**: PHP language keywords (class, function, if, etc.)
- **Identifiers**: User-defined names (variables, function names, etc.)
- **Literals**: String literals, numbers, and constants
- **Operators**: Arithmetic, comparison, and logical operators
- **Punctuation**: Brackets, parentheses, semicolons, etc.

## Running the example

```bash
cd examples/token-extraction
go run main.go
```

## Expected output

The program will display:
- Complete token stream with positions
- Statistical analysis of token types
- Categorized content (keywords, identifiers, literals)
- Operator and punctuation counts
- String interpolation tokenization example

## Use Cases

This pattern is useful for:
- **Syntax Highlighting**: Identify different token types for coloring
- **Code Metrics**: Analyze code complexity and patterns  
- **Refactoring Tools**: Find and replace specific token patterns
- **Language Servers**: Provide editor features like auto-completion
- **Code Analysis**: Detect coding patterns and potential issues