package parser

import (
	"testing"

	"github.com/wudi/php-parser/parser/testutils"
)

// TestParsing_TraitDeclarations Trait声明测试
func TestParsing_TraitDeclarations(t *testing.T) {
	suite := testutils.NewTestSuiteBuilder("TraitDeclarations", createParserFactory())

	suite.
		AddSimple("simple_trait_declaration",
			`<?php
trait LoggerTrait {
    public function log(string $message): void {
        echo $message;
    }
}
?>`,
			func(ctx *testutils.TestContext) {
				assertions := testutils.NewASTAssertions(ctx.T)
				body := assertions.AssertProgramBody(ctx.Program, 1)
				assertions.AssertTraitDeclaration(body[0], "LoggerTrait")
			}).
		AddSimple("trait_with_properties",
			`<?php
trait DatabaseTrait {
    protected $connection;
    public function connect() {}
}
?>`,
			func(ctx *testutils.TestContext) {
				assertions := testutils.NewASTAssertions(ctx.T)
				body := assertions.AssertProgramBody(ctx.Program, 1)
				assertions.AssertTraitDeclaration(body[0], "DatabaseTrait")
			}).
		Run(t)
}

// TestParsing_MatchExpressions Match表达式测试
func TestParsing_MatchExpressions(t *testing.T) {
	suite := testutils.NewTestSuiteBuilder("MatchExpressions", createParserFactory())

	suite.
		AddSimple("simple_match_expression",
			`<?php
$result = match ($status) {
    'pending' => 'waiting',
    'approved' => 'done',
    default => 'unknown'
};
?>`,
			func(ctx *testutils.TestContext) {
				assertions := testutils.NewASTAssertions(ctx.T)
				body := assertions.AssertProgramBody(ctx.Program, 1)

				exprStmt := assertions.AssertExpressionStatement(body[0])
				assignment := assertions.AssertAssignment(exprStmt.Expression, "=")
				assertions.AssertVariable(assignment.Left, "$result")
				assertions.AssertMatchExpression(assignment.Right)
			}).
		Run(t)
}

// TestParsing_YieldExpressions Yield表达式测试
func TestParsing_YieldExpressions(t *testing.T) {
	suite := testutils.NewTestSuiteBuilder("YieldExpressions", createParserFactory())

	suite.
		AddSimple("simple_yield",
			`<?php
function generator() {
    yield $value;
}
?>`,
			func(ctx *testutils.TestContext) {
				assertions := testutils.NewASTAssertions(ctx.T)
				body := assertions.AssertProgramBody(ctx.Program, 1)

				funcDecl := assertions.AssertFunctionDeclaration(body[0], "generator")
				assertions.AssertFunctionBody(funcDecl, 1)
			}).
		AddSimple("yield_from",
			`<?php
function gen() {
    yield from $other;
}
?>`,
			func(ctx *testutils.TestContext) {
				assertions := testutils.NewASTAssertions(ctx.T)
				body := assertions.AssertProgramBody(ctx.Program, 1)

				funcDecl := assertions.AssertFunctionDeclaration(body[0], "gen")
				assertions.AssertFunctionBody(funcDecl, 1)
			}).
		Run(t)
}

// TestParsing_FirstClassCallable FirstClassCallable测试
func TestParsing_FirstClassCallable(t *testing.T) {
	suite := testutils.NewTestSuiteBuilder("FirstClassCallable", createParserFactory())

	suite.
		AddSimple("function_reference",
			`<?php
$fn = strlen(...);
?>`,
			func(ctx *testutils.TestContext) {
				assertions := testutils.NewASTAssertions(ctx.T)
				body := assertions.AssertProgramBody(ctx.Program, 1)

				exprStmt := assertions.AssertExpressionStatement(body[0])
				assignment := assertions.AssertAssignment(exprStmt.Expression, "=")
				assertions.AssertVariable(assignment.Left, "$fn")
				// FirstClassCallable might be represented as a special expression type
			}).
		Run(t)
}

// TestParsing_ReturnStatement Return语句测试
func TestParsing_ReturnStatement(t *testing.T) {
	suite := testutils.NewTestSuiteBuilder("ReturnStatement", createParserFactory())

	suite.
		AddSimple("simple_return",
			`<?php
function test() {
    return $value;
}
?>`,
			func(ctx *testutils.TestContext) {
				assertions := testutils.NewASTAssertions(ctx.T)
				body := assertions.AssertProgramBody(ctx.Program, 1)

				funcDecl := assertions.AssertFunctionDeclaration(body[0], "test")
				assertions.AssertFunctionBody(funcDecl, 1)

				returnStmt := assertions.AssertReturnStatement(funcDecl.Body[0])
				assertions.AssertVariable(returnStmt.Argument, "$value")
			}).
		AddSimple("return_expression",
			`<?php
function add($a, $b) {
    return $a + $b;
}
?>`,
			func(ctx *testutils.TestContext) {
				assertions := testutils.NewASTAssertions(ctx.T)
				body := assertions.AssertProgramBody(ctx.Program, 1)

				funcDecl := assertions.AssertFunctionDeclaration(body[0], "add")
				returnStmt := assertions.AssertReturnStatement(funcDecl.Body[0])
				assertions.AssertBinaryExpression(returnStmt.Argument, "+")
			}).
		Run(t)
}
