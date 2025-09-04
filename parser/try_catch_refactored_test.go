package parser

import (
	"testing"

	"github.com/wudi/php-parser/parser/testutils"
)

// TestParsing_TryCatchWithStatements TryCatch语句测试
func TestParsing_TryCatchWithStatements(t *testing.T) {
	suite := testutils.NewTestSuiteBuilder("TryCatchWithStatements", createParserFactory())

	suite.
		AddSimple("try_catch_with_assignment_after",
			`<?php
try {
} catch (Exception $ex) {
}
$tested = $test->getName();`,
			func(ctx *testutils.TestContext) {
				assertions := testutils.NewASTAssertions(ctx.T)
				body := assertions.AssertProgramBody(ctx.Program, 2)

				// Check try-catch statement
				tryStmt := assertions.AssertTryStatement(body[0])
				assertions.AssertTryBlockEmpty(tryStmt)
				assertions.AssertCatchClausesCount(tryStmt, 1)

				// Check assignment statement after try-catch
				exprStmt := assertions.AssertExpressionStatement(body[1])
				assignment := assertions.AssertAssignment(exprStmt.Expression, "=")
				assertions.AssertVariable(assignment.Left, "$tested")
				assertions.AssertMethodCall(assignment.Right, "$test", "getName")
			}).
		AddSimple("try_catch_with_statements_in_blocks",
			`<?php
try {
    $x = 1;
} catch (Exception $ex) {
    $y = 2;
}
$z = 3;`,
			func(ctx *testutils.TestContext) {
				assertions := testutils.NewASTAssertions(ctx.T)
				body := assertions.AssertProgramBody(ctx.Program, 2)

				// Check try-catch statement
				tryStmt := assertions.AssertTryStatement(body[0])
				assertions.AssertTryBlockStatements(tryStmt, 1)
				assertions.AssertCatchClausesCount(tryStmt, 1)

				// Check statement after try-catch
				exprStmt := assertions.AssertExpressionStatement(body[1])
				assignment := assertions.AssertAssignment(exprStmt.Expression, "=")
				assertions.AssertVariable(assignment.Left, "$z")
			}).
		AddSimple("multiple_catch_clauses",
			`<?php
try {
    throw new Exception();
} catch (InvalidArgumentException $e) {
    echo "invalid";
} catch (Exception $ex) {
    echo "general";
}
$done = true;`,
			func(ctx *testutils.TestContext) {
				assertions := testutils.NewASTAssertions(ctx.T)
				body := assertions.AssertProgramBody(ctx.Program, 2)

				// Check try-catch statement
				tryStmt := assertions.AssertTryStatement(body[0])
				assertions.AssertCatchClausesCount(tryStmt, 2)

				// Check statement after try-catch
				exprStmt := assertions.AssertExpressionStatement(body[1])
				assignment := assertions.AssertAssignment(exprStmt.Expression, "=")
				assertions.AssertVariable(assignment.Left, "$done")
			}).
		Run(t)
}
