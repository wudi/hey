package parser

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wudi/php-parser/parser/testutils"
)

// 批量重构剩余的重要测试案例

// TestParsing_TypedParameters 类型化参数测试
func TestParsing_TypedParameters(t *testing.T) {
	suite := testutils.NewTestSuiteBuilder("TypedParameters", createParserFactory())

	suite.
		AddSimple("typed_parameter",
			`<?php function test(string $param) {} ?>`,
			func(ctx *testutils.TestContext) {
				assertions := testutils.NewASTAssertions(ctx.T)
				body := assertions.AssertProgramBody(ctx.Program, 1)
				assertions.AssertFunctionDeclaration(body[0], "test")
			}).
		Run(t)
}

// TestParsing_FunctionReturnTypes 函数返回类型测试
func TestParsing_FunctionReturnTypes(t *testing.T) {
	suite := testutils.NewTestSuiteBuilder("FunctionReturnTypes", createParserFactory())

	suite.
		AddSimple("string_return_type",
			`<?php function getName(): string { return "test"; } ?>`,
			func(ctx *testutils.TestContext) {
				assertions := testutils.NewASTAssertions(ctx.T)
				body := assertions.AssertProgramBody(ctx.Program, 1)
				assertions.AssertFunctionDeclaration(body[0], "getName")
			}).
		Run(t)
}

// TestParsing_BitwiseOperations 位运算测试
func TestParsing_BitwiseOperations(t *testing.T) {
	suite := testutils.NewTestSuiteBuilder("BitwiseOperations", createParserFactory())

	suite.
		AddSimple("bitwise_and",
			`<?php $result = $a & $b; ?>`,
			func(ctx *testutils.TestContext) {
				assertions := testutils.NewASTAssertions(ctx.T)
				body := assertions.AssertProgramBody(ctx.Program, 1)

				exprStmt := assertions.AssertExpressionStatement(body[0])
				assignment := assertions.AssertAssignment(exprStmt.Expression, "=")
				assertions.AssertBinaryExpression(assignment.Right, "&")
			}).
		Run(t)
}

// TestParsing_ArrayExpression 数组表达式测试
func TestParsing_ArrayExpression(t *testing.T) {
	suite := testutils.NewTestSuiteBuilder("ArrayExpression", createParserFactory())

	suite.
		AddSimple("basic_array",
			`<?php $arr = [1, 2, 3]; ?>`,
			func(ctx *testutils.TestContext) {
				assertions := testutils.NewASTAssertions(ctx.T)
				body := assertions.AssertProgramBody(ctx.Program, 1)

				exprStmt := assertions.AssertExpressionStatement(body[0])
				assignment := assertions.AssertAssignment(exprStmt.Expression, "=")
				assertions.AssertArray(assignment.Right, 3)
			}).
		Run(t)
}

// TestParsing_ArrayTrailingCommas 数组尾随逗号测试
func TestParsing_ArrayTrailingCommas(t *testing.T) {
	suite := testutils.NewTestSuiteBuilder("ArrayTrailingCommas", createParserFactory())

	suite.
		AddSimple("array_with_trailing_comma",
			`<?php $arr = [1, 2, 3,]; ?>`,
			func(ctx *testutils.TestContext) {
				assertions := testutils.NewASTAssertions(ctx.T)
				body := assertions.AssertProgramBody(ctx.Program, 1)

				exprStmt := assertions.AssertExpressionStatement(body[0])
				assignment := assertions.AssertAssignment(exprStmt.Expression, "=")
				assertions.AssertArray(assignment.Right, 3)
			}).
		Run(t)
}

// TestParsing_GroupedExpression 分组表达式测试
func TestParsing_GroupedExpression(t *testing.T) {
	suite := testutils.NewTestSuiteBuilder("GroupedExpression", createParserFactory())

	suite.
		AddSimple("parenthesized_expression",
			`<?php $result = ($a + $b) * $c; ?>`,
			func(ctx *testutils.TestContext) {
				assertions := testutils.NewASTAssertions(ctx.T)
				body := assertions.AssertProgramBody(ctx.Program, 1)

				exprStmt := assertions.AssertExpressionStatement(body[0])
				assignment := assertions.AssertAssignment(exprStmt.Expression, "=")
				assertions.AssertBinaryExpression(assignment.Right, "*")
			}).
		Run(t)
}

// TestParsing_OperatorPrecedence 操作符优先级测试
func TestParsing_OperatorPrecedence(t *testing.T) {
	suite := testutils.NewTestSuiteBuilder("OperatorPrecedence", createParserFactory())

	suite.
		AddSimple("arithmetic_precedence",
			`<?php $result = $a + $b * $c; ?>`,
			func(ctx *testutils.TestContext) {
				assertions := testutils.NewASTAssertions(ctx.T)
				body := assertions.AssertProgramBody(ctx.Program, 1)

				exprStmt := assertions.AssertExpressionStatement(body[0])
				assignment := assertions.AssertAssignment(exprStmt.Expression, "=")
				assertions.AssertBinaryExpression(assignment.Right, "+")
			}).
		Run(t)
}

// TestParsing_HeredocStrings Heredoc字符串测试
func TestParsing_HeredocStrings(t *testing.T) {
	suite := testutils.NewTestSuiteBuilder("HeredocStrings", createParserFactory())

	suite.
		AddSimple("simple_heredoc",
			`<?php $str = <<<EOT
Hello World
EOT; ?>`,
			func(ctx *testutils.TestContext) {
				assertions := testutils.NewASTAssertions(ctx.T)
				body := assertions.AssertProgramBody(ctx.Program, 1)

				exprStmt := assertions.AssertExpressionStatement(body[0])
				assignment := assertions.AssertAssignment(exprStmt.Expression, "=")
				// Heredoc might be represented as HeredocExpression or StringLiteral
				require.NotNil(ctx.T, assignment.Right)
			}).
		Run(t)
}

// TestParsing_NowdocStrings Nowdoc字符串测试
func TestParsing_NowdocStrings(t *testing.T) {
	suite := testutils.NewTestSuiteBuilder("NowdocStrings", createParserFactory())

	suite.
		AddSimple("simple_nowdoc",
			`<?php $str = <<<'EOT'
Hello World
EOT; ?>`,
			func(ctx *testutils.TestContext) {
				assertions := testutils.NewASTAssertions(ctx.T)
				body := assertions.AssertProgramBody(ctx.Program, 1)

				exprStmt := assertions.AssertExpressionStatement(body[0])
				assignment := assertions.AssertAssignment(exprStmt.Expression, "=")
				require.NotNil(ctx.T, assignment.Right)
			}).
		Run(t)
}

// TestParsing_StringInterpolation 字符串插值测试
func TestParsing_StringInterpolation(t *testing.T) {
	suite := testutils.NewTestSuiteBuilder("StringInterpolation", createParserFactory())

	suite.
		AddSimple("variable_interpolation",
			`<?php $str = "Hello $name"; ?>`,
			func(ctx *testutils.TestContext) {
				assertions := testutils.NewASTAssertions(ctx.T)
				body := assertions.AssertProgramBody(ctx.Program, 1)

				exprStmt := assertions.AssertExpressionStatement(body[0])
				assignment := assertions.AssertAssignment(exprStmt.Expression, "=")
				// String interpolation might be InterpolatedStringExpression
				require.NotNil(ctx.T, assignment.Right)
			}).
		Run(t)
}
