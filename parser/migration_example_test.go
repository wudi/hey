package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wudi/php-parser/ast"
	"github.com/wudi/php-parser/lexer"
	"github.com/wudi/php-parser/parser/testutils"
)

// TestMigrationExample_Before 原有测试风格
func TestMigrationExample_Before(t *testing.T) {
	input := `<?php $name = "John"; ?>`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	checkParserErrors(t, p)
	assert.NotNil(t, program)
	assert.Len(t, program.Body, 1)

	stmt := program.Body[0]
	exprStmt, ok := stmt.(*ast.ExpressionStatement)
	assert.True(t, ok, "Statement should be ExpressionStatement")

	assignment, ok := exprStmt.Expression.(*ast.AssignmentExpression)
	assert.True(t, ok, "Expression should be AssignmentExpression")

	// 检查左侧变量
	variable, ok := assignment.Left.(*ast.Variable)
	assert.True(t, ok, "Left side should be Variable")
	assert.Equal(t, "$name", variable.Name)

	// 检查操作符
	assert.Equal(t, "=", assignment.Operator)

	// 检查右侧字符串字面量
	stringLit, ok := assignment.Right.(*ast.StringLiteral)
	assert.True(t, ok, "Right side should be StringLiteral")
	assert.Equal(t, "John", stringLit.Value)
	assert.Equal(t, `"John"`, stringLit.Raw)
}

// TestMigrationExample_After 新架构测试风格
func TestMigrationExample_After(t *testing.T) {
	parserFactory := func(l *lexer.Lexer) testutils.ParserInterface {
		return New(l)
	}
	builder := testutils.NewParserTestBuilder(parserFactory)

	builder.Test(t,
		`<?php $name = "John"; ?>`,
		testutils.ValidateStringAssignment("$name", "John", `"John"`),
	)
}

// TestMigrationExample_TableDriven 表驱动测试迁移示例
func TestMigrationExample_TableDriven(t *testing.T) {
	parserFactory := func(l *lexer.Lexer) testutils.ParserInterface {
		return New(l)
	}
	builder := testutils.NewParserTestBuilder(parserFactory)

	tests := []struct {
		Name      string
		Source    string
		Validator func(*testutils.TestContext)
	}{
		{
			Name:      "string_assignment",
			Source:    `<?php $name = "John"; ?>`,
			Validator: testutils.ValidateStringAssignment("$name", "John", `"John"`),
		},
		{
			Name:      "integer_assignment",
			Source:    `<?php $age = 25; ?>`,
			Validator: testutils.ValidateVariable("$age"),
		},
		{
			Name:   "complex_assignment",
			Source: `<?php $greeting = "Hello " . $name; ?>`,
			Validator: func(ctx *testutils.TestContext) {
				assertions := testutils.NewASTAssertions(t)
				body := assertions.AssertProgramBody(ctx.Program, 1)

				exprStmt := assertions.AssertExpressionStatement(body[0])
				assignment := assertions.AssertAssignment(exprStmt.Expression, "=")
				assertions.AssertVariable(assignment.Left, "$greeting")

				// 验证右侧是二元表达式（字符串连接）
				assertions.AssertBinaryExpression(assignment.Right, ".")
			},
		},
	}

	builder.TestTableDriven(t, tests)
}
