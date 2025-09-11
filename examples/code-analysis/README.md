# Code Analysis Example

This example demonstrates comprehensive static code analysis using the PHP parser to extract metrics, detect patterns, and assess code quality.

## What it does

- Performs comprehensive static analysis of PHP code
- Calculates various code metrics and complexity scores
- Identifies potential code quality issues
- Provides detailed reports with actionable insights
- Generates overall code quality scores

## Key Features

### Code Metrics
- **Lines of Code**: Total line count
- **Function/Class Counts**: Structure analysis  
- **Variable Usage**: Frequency analysis
- **Nesting Depth**: Complexity measurement
- **Cyclomatic Complexity**: Control flow complexity

### Analysis Components
- **FunctionInfo**: Parameter count, visibility, complexity
- **ClassInfo**: Methods, properties, visibility distribution
- **Variable Tracking**: Usage patterns and frequency
- **Issue Detection**: Quality problems with suggestions

### Quality Checks
- **Parameter Count**: Warns about functions with too many parameters
- **Complexity Analysis**: Identifies overly complex code paths  
- **Nesting Depth**: Detects deeply nested structures
- **Code Patterns**: Analyzes expressions and control structures

## Metrics Analyzed

### Structural Metrics
- Function and class declarations
- Method and property counts
- Variable usage frequency
- Expression complexity

### Complexity Metrics
- Cyclomatic complexity (control flow branches)
- Maximum nesting depth
- Binary expression density
- Assignment pattern analysis

### Quality Indicators
- Parameter count warnings
- Complexity thresholds
- Pattern detection
- Best practice suggestions

## Running the example

```bash
cd examples/code-analysis
go run main.go
```

## Expected output

The program provides:
- **Code Metrics**: Quantitative analysis of code structure
- **Function Analysis**: Details about each function's complexity
- **Class Analysis**: Object-oriented structure examination
- **Variable Usage**: Frequency and pattern analysis
- **Issues Found**: Quality problems with improvement suggestions
- **Quality Score**: Overall assessment (0-100 scale)
- **Pattern Analysis**: Additional structural insights
- **Improvement Suggestions**: Actionable recommendations

## Score Card System

The quality score is calculated based on:
- **Base Score**: 100 points
- **Complexity Penalty**: -10 for high cyclomatic complexity
- **Nesting Penalty**: -15 for deep nesting (>5 levels)
- **Issue Penalty**: -5 per detected issue

Score ranges:
- 90-100: 🌟 Excellent
- 80-89: ✅ Good  
- 70-79: 👍 Fair
- 60-69: ⚠️ Needs Improvement
- 0-59: ❌ Poor

## Production Applications

This analysis framework is ideal for:
- **Code Review Tools**: Automated quality assessment
- **CI/CD Integration**: Quality gates and metrics tracking
- **Technical Debt Analysis**: Identifying refactoring priorities
- **Coding Standards**: Enforcement of best practices
- **Documentation Generation**: Automated code insights
- **Refactoring Planning**: Data-driven improvement strategies

## Extending the Analyzer

To add more analysis features:
- Implement semantic analysis rules
- Add design pattern detection
- Include security vulnerability scanning
- Support custom quality rules
- Add historical trend analysis
- Integrate with external quality tools