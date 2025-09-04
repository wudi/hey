package parser

import (
	"testing"

	"github.com/wudi/php-parser/parser/testutils"
)

// TestRefactored_ControlFlowStatements 重构后的控制流语句测试
func TestRefactored_ControlFlowStatements(t *testing.T) {
	suite := testutils.NewTestSuiteBuilder("ControlFlowStatements", createParserFactory())

	// If语句测试
	suite.AddSimple("if_statement",
		`<?php if ($x > 5) { echo "big"; } ?>`,
		testutils.ValidateIfStatement(
			testutils.ValidateBinaryExpression("$x", ">", "5"),
			testutils.ValidateEchoArgs([]string{`"big"`})))

	// If-else语句测试
	suite.AddSimple("if_else_statement",
		`<?php if ($x > 5) { echo "big"; } else { echo "small"; } ?>`,
		testutils.ValidateIfElseStatement(
			testutils.ValidateBinaryExpression("$x", ">", "5"),
			testutils.ValidateEchoArgs([]string{`"big"`}),
			testutils.ValidateEchoArgs([]string{`"small"`})))

	// While语句测试
	suite.AddSimple("while_statement",
		`<?php while ($i < 10) { $i++; } ?>`,
		testutils.ValidateWhileStatement(
			testutils.ValidateBinaryExpression("$i", "<", "10"),
			testutils.ValidatePostfixExpression("$i", "++")))

	// For语句测试
	suite.AddSimple("for_statement",
		`<?php for ($i = 0; $i < 10; $i++) { echo $i; } ?>`,
		testutils.ValidateForStatement(
			testutils.ValidateAssignmentExpression("$i", "0"),
			testutils.ValidateBinaryExpression("$i", "<", "10"),
			testutils.ValidatePostfixExpression("$i", "++"),
			testutils.ValidateEchoVariable("$i")))

	suite.Run(t)
}

// TestRefactored_AlternativeSyntax 重构后的替代语法测试
func TestRefactored_AlternativeSyntax(t *testing.T) {
	suite := testutils.NewTestSuiteBuilder("AlternativeSyntax", createParserFactory())

	// 替代if语法
	suite.AddSimple("alternative_if_statement",
		`<?php if ($x > 0): echo "positive"; endif; ?>`,
		testutils.ValidateIfStatement(
			testutils.ValidateBinaryExpression("$x", ">", "0"),
			testutils.ValidateEchoArgs([]string{`"positive"`})))

	// 替代while语法
	suite.AddSimple("alternative_while_statement",
		`<?php while ($i < 5): $i++; endwhile; ?>`,
		testutils.ValidateWhileStatement(
			testutils.ValidateBinaryExpression("$i", "<", "5"),
			testutils.ValidatePostfixExpression("$i", "++")))

	// 替代for语法
	suite.AddSimple("alternative_for_statement",
		`<?php for ($i = 0; $i < 3; $i++): echo $i; endfor; ?>`,
		testutils.ValidateForStatement(
			testutils.ValidateAssignmentExpression("$i", "0"),
			testutils.ValidateBinaryExpression("$i", "<", "3"),
			testutils.ValidatePostfixExpression("$i", "++"),
			testutils.ValidateEchoVariable("$i")))

	// 替代foreach语法
	suite.AddSimple("alternative_foreach_statement",
		`<?php foreach ($items as $item): echo $item; endforeach; ?>`,
		testutils.ValidateForeachStatement("$items", "", "$item",
			testutils.ValidateEchoVariable("$item")))

	suite.Run(t)
}

// TestRefactored_SimpleControlFlow 简化的控制流测试
func TestRefactored_SimpleControlFlow(t *testing.T) {
	suite := testutils.NewTestSuiteBuilder("SimpleControlFlow", createParserFactory())

	// 基础if测试
	suite.AddSimple("simple_if",
		`<?php if ($x) echo "true"; ?>`,
		testutils.ValidateIfStatement(
			testutils.ValidateVariableExpression("$x"),
			testutils.ValidateEchoArgs([]string{`"true"`})))

	// 基础while测试
	suite.AddSimple("simple_while",
		`<?php while ($i--) doSomething(); ?>`,
		testutils.ValidateWhileStatement(
			testutils.ValidatePostfixExpression("$i", "--"),
			testutils.ValidateFunctionCall("doSomething")))

	suite.Run(t)
}
