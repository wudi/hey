//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"strings"

	"github.com/wudi/hey/compiler/ast"
	"github.com/wudi/hey/compiler/lexer"
	"github.com/wudi/hey/compiler/parser"
)

// ErrorReporter 自定义错误报告器
type ErrorReporter struct {
	SyntaxErrors   []string
	LexicalErrors  []string
	SemanticErrors []string
	TotalErrors    int
}

func NewErrorReporter() *ErrorReporter {
	return &ErrorReporter{
		SyntaxErrors:   make([]string, 0),
		LexicalErrors:  make([]string, 0),
		SemanticErrors: make([]string, 0),
	}
}

func (er *ErrorReporter) ReportError(errorType, message string) {
	er.TotalErrors++
	switch errorType {
	case "syntax":
		er.SyntaxErrors = append(er.SyntaxErrors, message)
	case "lexical":
		er.LexicalErrors = append(er.LexicalErrors, message)
	case "semantic":
		er.SemanticErrors = append(er.SemanticErrors, message)
	}
}

func (er *ErrorReporter) PrintReport() {
	fmt.Printf("=== Error Analysis Report ===\n")
	fmt.Printf("Total errors found: %d\n\n", er.TotalErrors)

	if len(er.SyntaxErrors) > 0 {
		fmt.Printf("Syntax Errors (%d):\n", len(er.SyntaxErrors))
		for i, err := range er.SyntaxErrors {
			fmt.Printf("  %d. %s\n", i+1, err)
		}
		fmt.Println()
	}

	if len(er.LexicalErrors) > 0 {
		fmt.Printf("Lexical Errors (%d):\n", len(er.LexicalErrors))
		for i, err := range er.LexicalErrors {
			fmt.Printf("  %d. %s\n", i+1, err)
		}
		fmt.Println()
	}

	if len(er.SemanticErrors) > 0 {
		fmt.Printf("Semantic Errors (%d):\n", len(er.SemanticErrors))
		for i, err := range er.SemanticErrors {
			fmt.Printf("  %d. %s\n", i+1, err)
		}
		fmt.Println()
	}
}

// parseWithErrorHandling 解析代码并处理各种错误
func parseWithErrorHandling(code, description string) {
	fmt.Printf("=== %s ===\n", description)
	fmt.Println("PHP Code:")
	fmt.Printf("```php\n%s\n```\n", code)

	reporter := NewErrorReporter()

	// 创建lexer和parser
	l := lexer.New(code)
	p := parser.New(l)

	// 解析代码
	program := p.ParseProgram()

	// 收集parser错误
	parserErrors := p.Errors()
	for _, err := range parserErrors {
		if strings.Contains(err, "unexpected token") ||
			strings.Contains(err, "expected") ||
			strings.Contains(err, "no prefix parse function") {
			reporter.ReportError("syntax", err)
		} else {
			reporter.ReportError("lexical", err)
		}
	}

	// 基本语义检查（示例）
	if program != nil {
		checkBasicSemantics(program, reporter)
	}

	// 打印错误报告
	reporter.PrintReport()

	// 显示是否成功解析
	if reporter.TotalErrors == 0 {
		fmt.Printf("✅ Parsing completed successfully!\n")
		fmt.Printf("Generated %d statements in AST\n", len(program.Body))
	} else {
		fmt.Printf("❌ Parsing completed with %d errors\n", reporter.TotalErrors)
		if program != nil {
			fmt.Printf("Partial AST generated with %d statements\n", len(program.Body))
		}
	}

	fmt.Println(strings.Repeat("-", 60))
	fmt.Println()
}

// checkBasicSemantics 基本语义检查示例
func checkBasicSemantics(program *ast.Program, reporter *ErrorReporter) {
	// 这里可以添加各种语义检查
	// 例如：检查变量使用、函数调用等

	// 示例：检查是否有空的语句块（这只是演示，实际语义分析会更复杂）
	if len(program.Body) == 0 {
		reporter.ReportError("semantic", "Empty program - no statements found")
	}
}

func main() {
	fmt.Println("=== PHP Parser Error Handling Examples ===\n")

	// 示例1：正确的PHP代码
	validCode := `<?php
$message = "Hello, World!";
echo $message;

function greet($name) {
    return "Hello, " . $name;
}

$greeting = greet("PHP");
echo $greeting;
?>`

	parseWithErrorHandling(validCode, "Example 1: Valid PHP Code")

	// 示例2：语法错误 - 缺少分号
	missingSemicolon := `<?php
$name = "John"
echo $name;
?>`

	parseWithErrorHandling(missingSemicolon, "Example 2: Missing Semicolon")

	// 示例3：语法错误 - 不匹配的括号
	unmatchedParentheses := `<?php
function test() {
    echo "Hello World";
// 缺少右括号
?>`

	parseWithErrorHandling(unmatchedParentheses, "Example 3: Unmatched Parentheses")

	// 示例4：语法错误 - 无效的变量名
	invalidVariableName := `<?php
$1invalid_name = "test";
echo $1invalid_name;
?>`

	parseWithErrorHandling(invalidVariableName, "Example 4: Invalid Variable Name")

	// 示例5：语法错误 - 错误的函数语法
	invalidFunctionSyntax := `<?php
function {
    echo "Missing function name";
}
?>`

	parseWithErrorHandling(invalidFunctionSyntax, "Example 5: Invalid Function Syntax")

	// 示例6：复杂的错误情况
	complexErrors := `<?php
class Test {
    public function method1() {
        $var = "test"  // 缺少分号
        return $var
    }  // 缺少分号
    
    public function method2() {
        // 不匹配的引号
        echo "Hello World';
    }
    
    // 语法错误的属性声明
    public $property = function() { return "test"; };
}
?>`

	parseWithErrorHandling(complexErrors, "Example 6: Multiple Complex Errors")

	// 示例7：字符串相关错误
	stringErrors := `<?php
$str1 = "Unclosed string
$str2 = 'Another unclosed string;
echo $str1 . $str2;
?>`

	parseWithErrorHandling(stringErrors, "Example 7: String-related Errors")

	// 示例8：部分恢复解析示例
	partialRecovery := `<?php
$good_var = "This is fine";
echo $good_var;

// 下面有错误
$bad_var = "Missing quote;
echo $bad_var;

// 但解析器应该能继续处理后面的代码
$another_good_var = "This should still work";
echo $another_good_var;
?>`

	parseWithErrorHandling(partialRecovery, "Example 8: Error Recovery")

	// 总结
	fmt.Println("=== Error Handling Summary ===")
	fmt.Println("This example demonstrates:")
	fmt.Println("  • How to collect and categorize parsing errors")
	fmt.Println("  • Different types of syntax errors PHP parsers encounter")
	fmt.Println("  • Error reporting with detailed information")
	fmt.Println("  • Parser error recovery capabilities")
	fmt.Println("  • Integration of lexical and syntax error handling")
	fmt.Println()
	fmt.Println("In production use, you would:")
	fmt.Println("  • Add line/column information to error reports")
	fmt.Println("  • Implement more sophisticated semantic analysis")
	fmt.Println("  • Provide suggestions for fixing common errors")
	fmt.Println("  • Support IDE integration for real-time error highlighting")
}
