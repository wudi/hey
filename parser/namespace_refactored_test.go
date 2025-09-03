package parser

import (
	"testing"
	
	"github.com/wudi/php-parser/parser/testutils"
)

// TestParsing_NamespaceStatements Namespace语句测试
func TestParsing_NamespaceStatements(t *testing.T) {
	suite := testutils.NewTestSuiteBuilder("NamespaceStatements", createParserFactory())
	
	suite.
		AddSimple("simple_namespace_declaration",
			`<?php
namespace App;
?>`,
			func(ctx *testutils.TestContext) {
				assertions := testutils.NewASTAssertions(ctx.T)
				body := assertions.AssertProgramBody(ctx.Program, 1)
				assertions.AssertNamespaceStatement(body[0], "App")
			}).
		AddSimple("multi_level_namespace",
			`<?php
namespace App\Http\Controllers;
?>`,
			func(ctx *testutils.TestContext) {
				assertions := testutils.NewASTAssertions(ctx.T)
				body := assertions.AssertProgramBody(ctx.Program, 1)
				assertions.AssertNamespaceStatement(body[0], "App\\Http\\Controllers")
			}).
		AddSimple("global_namespace",
			`<?php
namespace;
?>`,
			func(ctx *testutils.TestContext) {
				assertions := testutils.NewASTAssertions(ctx.T)
				body := assertions.AssertProgramBody(ctx.Program, 1)
				assertions.AssertGlobalNamespaceStatement(body[0])
			}).
		Run(t)
}

// TestParsing_NamespaceSeparator Namespace分隔符测试
func TestParsing_NamespaceSeparator(t *testing.T) {
	suite := testutils.NewTestSuiteBuilder("NamespaceSeparator", createParserFactory())
	
	suite.
		AddSimple("fully_qualified_namespace_call",
			`<?php \DateTime\createFromFormat();`,
			func(ctx *testutils.TestContext) {
				assertions := testutils.NewASTAssertions(ctx.T)
				body := assertions.AssertProgramBody(ctx.Program, 1)
				
				exprStmt := assertions.AssertExpressionStatement(body[0])
				callExpr := assertions.AssertCallExpression(exprStmt.Expression)
				assertions.AssertFullyQualifiedCall(callExpr, "DateTime\\createFromFormat")
			}).
		Run(t)
}