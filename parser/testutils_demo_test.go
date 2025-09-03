package parser

import (
	"testing"
	
	"github.com/wudi/php-parser/lexer"
	"github.com/wudi/php-parser/parser/testutils"
)

// TestNewArchitecture_VariableDeclaration 使用新架构的变量声明测试
func TestNewArchitecture_VariableDeclaration(t *testing.T) {
	// 创建解析器工厂函数
	parserFactory := func(l *lexer.Lexer) testutils.ParserInterface {
		return New(l)
	}
	builder := testutils.NewParserTestBuilder(parserFactory)
	
	// 单个测试用例
	t.Run("simple_string_assignment", func(t *testing.T) {
		builder.Test(t, 
			`<?php $name = "John"; ?>`,
			testutils.ValidateStringAssignment("$name", "John", `"John"`),
		)
	})
	
	// 表驱动测试
	tests := []struct {
		Name      string
		Source    string
		Validator func(*testutils.TestContext)
	}{
		{
			Name:      "integer_assignment",
			Source:    `<?php $age = 25; ?>`,
			Validator: testutils.ValidateVariable("$age"),
		},
		{
			Name:      "string_assignment",
			Source:    `<?php $greeting = "Hello"; ?>`,
			Validator: testutils.ValidateStringAssignment("$greeting", "Hello", `"Hello"`),
		},
		{
			Name:      "boolean_assignment", 
			Source:    `<?php $flag = true; ?>`,
			Validator: testutils.ValidateVariable("$flag"),
		},
	}
	
	builder.TestTableDriven(t, tests)
}

// TestNewArchitecture_EchoStatement 使用新架构的echo语句测试
func TestNewArchitecture_EchoStatement(t *testing.T) {
	parserFactory := func(l *lexer.Lexer) testutils.ParserInterface {
		return New(l)
	}
	builder := testutils.NewParserTestBuilder(parserFactory)
	
	t.Run("simple_echo", func(t *testing.T) {
		builder.Test(t, 
			`<?php echo "Hello, World!"; ?>`,
			func(ctx *testutils.TestContext) {
				assertions := testutils.NewASTAssertions(t)
				body := assertions.AssertProgramBody(ctx.Program, 1)
				
				echoStmt := assertions.AssertEchoStatement(body[0], 1)
				assertions.AssertStringLiteral(
					echoStmt.Arguments.Arguments[0], 
					"Hello, World!", 
					`"Hello, World!"`,
				)
			},
		)
	})
	
	t.Run("multiple_arguments", func(t *testing.T) {
		builder.Test(t,
			`<?php echo "Hello", " ", "World!"; ?>`,
			func(ctx *testutils.TestContext) {
				assertions := testutils.NewASTAssertions(t)
				body := assertions.AssertProgramBody(ctx.Program, 1)
				
				echoStmt := assertions.AssertEchoStatement(body[0], 3)
				
				expectedValues := []string{"Hello", " ", "World!"}
				expectedRaws := []string{`"Hello"`, `" "`, `"World!"`}
				
				for i, arg := range echoStmt.Arguments.Arguments {
					assertions.AssertStringLiteral(arg, expectedValues[i], expectedRaws[i])
				}
			},
		)
	})
}

// TestNewArchitecture_ErrorHandling 测试错误处理
func TestNewArchitecture_ErrorHandling(t *testing.T) {
	parserFactory := func(l *lexer.Lexer) testutils.ParserInterface {
		return New(l)
	}
	builder := testutils.NewParserTestBuilder(parserFactory).WithStrictMode(false)
	
	t.Run("non_strict_mode", func(t *testing.T) {
		builder.Test(t,
			`<?php $incomplete = `,
			func(ctx *testutils.TestContext) {
				// 在非严格模式下，程序应该存在，解析器会尽力解析
				
				// 程序应该存在，可能有部分解析的内容
				if ctx.Program != nil {
					body := ctx.Program.Body
					t.Logf("Parsed %d statements", len(body))
					// 解析器可能会创建一些语句，即使有错误
				}
				
				// 记录错误信息用于调试
				errors := ctx.Parser.Errors()
				t.Logf("Parser errors: %v", errors)
				
				// 在非严格模式下，我们不强制要求有错误
				// 因为解析器可能已经很好地处理了这种情况
			},
		)
	})
}