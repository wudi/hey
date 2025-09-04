package parser

import (
	"testing"

	"github.com/wudi/php-parser/parser/testutils"
)

// TestBasic_VariableDeclaration 基础变量声明测试
func TestBasic_VariableDeclaration(t *testing.T) {
	suite := testutils.NewTestSuiteBuilder("VariableDeclaration", createParserFactory())

	suite.
		AddStringAssignment("simple_string", "$name", "John", `"John"`).
		AddStringAssignment("empty_string", "$empty", "", `""`).
		AddStringAssignment("single_quotes", "$msg", "Hello World", "'Hello World'").
		AddVariableAssignment("integer", "$age", "25").
		AddVariableAssignment("float", "$price", "19.99").
		AddVariableAssignment("boolean_true", "$flag", "true").
		AddVariableAssignment("boolean_false", "$active", "false").
		AddVariableAssignment("null_value", "$data", "null").
		Run(t)
}

// TestBasic_EchoStatement 基础echo语句测试
func TestBasic_EchoStatement(t *testing.T) {
	suite := testutils.NewTestSuiteBuilder("EchoStatement", createParserFactory())

	suite.
		AddEcho("single_string", []string{`"Hello World"`},
			testutils.ValidateStringArg("Hello World", `"Hello World"`)).
		AddEcho("multiple_strings", []string{`"Hello"`, `" "`, `"World"`},
			testutils.ValidateStringArg("Hello", `"Hello"`),
			testutils.ValidateStringArg(" ", `" "`),
			testutils.ValidateStringArg("World", `"World"`)).
		AddEcho("mixed_args", []string{`"Count:"`, "$count", "42"},
			testutils.ValidateStringArg("Count:", `"Count:"`)).
		Run(t)
}

// TestBasic_EchoWithoutSemicolon echo语句无分号测试
func TestBasic_EchoWithoutSemicolon(t *testing.T) {
	suite := testutils.NewTestSuiteBuilder("EchoWithoutSemicolon", createParserFactory())

	suite.
		AddSimple("simple_string_echo", `<?php echo 'hello' ?>`,
			func(ctx *testutils.TestContext) {
				assertions := testutils.NewASTAssertions(ctx.T)
				body := assertions.AssertProgramBody(ctx.Program, 1)

				echoStmt := assertions.AssertEchoStatement(body[0], 1)
				assertions.AssertStringLiteral(echoStmt.Arguments.Arguments[0], "hello", "'hello'")
			}).
		AddSimple("variable_echo", `<?php echo $var ?>`,
			func(ctx *testutils.TestContext) {
				assertions := testutils.NewASTAssertions(ctx.T)
				body := assertions.AssertProgramBody(ctx.Program, 1)

				echoStmt := assertions.AssertEchoStatement(body[0], 1)
				assertions.AssertVariable(echoStmt.Arguments.Arguments[0], "$var")
			}).
		AddSimple("ternary_expression_echo", `<?php echo $active_frames_tab == 'application' ? 'frames-container-application' : '' ?>`,
			func(ctx *testutils.TestContext) {
				assertions := testutils.NewASTAssertions(ctx.T)
				body := assertions.AssertProgramBody(ctx.Program, 1)

				echoStmt := assertions.AssertEchoStatement(body[0], 1)
				assertions.AssertTernaryExpression(echoStmt.Arguments.Arguments[0])
			}).
		AddSimple("multiple_arguments_echo", `<?php echo 'Hello', ' ', 'World' ?>`,
			func(ctx *testutils.TestContext) {
				assertions := testutils.NewASTAssertions(ctx.T)
				body := assertions.AssertProgramBody(ctx.Program, 1)

				echoStmt := assertions.AssertEchoStatement(body[0], 3)
				assertions.AssertStringLiteral(echoStmt.Arguments.Arguments[0], "Hello", "'Hello'")
				assertions.AssertStringLiteral(echoStmt.Arguments.Arguments[1], " ", "' '")
				assertions.AssertStringLiteral(echoStmt.Arguments.Arguments[2], "World", "'World'")
			}).
		AddSimple("complex_expression_echo", `<?php echo $a + $b * 2 ?>`,
			func(ctx *testutils.TestContext) {
				assertions := testutils.NewASTAssertions(ctx.T)
				body := assertions.AssertProgramBody(ctx.Program, 1)

				echoStmt := assertions.AssertEchoStatement(body[0], 1)
				assertions.AssertBinaryExpression(echoStmt.Arguments.Arguments[0], "+")
			}).
		Run(t)
}

// TestBasic_FloatLiteralEdgeCases 浮点数字面量边界情况测试
func TestBasic_FloatLiteralEdgeCases(t *testing.T) {
	suite := testutils.NewTestSuiteBuilder("FloatLiteralEdgeCases", createParserFactory())

	suite.
		AddSimple("float_ending_with_decimal", `<?php $x = 1.; ?>`,
			func(ctx *testutils.TestContext) {
				assertions := testutils.NewASTAssertions(ctx.T)
				body := assertions.AssertProgramBody(ctx.Program, 1)

				exprStmt := assertions.AssertExpressionStatement(body[0])
				assignment := assertions.AssertAssignment(exprStmt.Expression, "=")
				assertions.AssertVariable(assignment.Left, "$x")
				assertions.AssertNumberLiteral(assignment.Right, "1.")
			}).
		AddSimple("float_with_zero_decimal", `<?php $x = 1.0; ?>`,
			func(ctx *testutils.TestContext) {
				assertions := testutils.NewASTAssertions(ctx.T)
				body := assertions.AssertProgramBody(ctx.Program, 1)

				exprStmt := assertions.AssertExpressionStatement(body[0])
				assignment := assertions.AssertAssignment(exprStmt.Expression, "=")
				assertions.AssertVariable(assignment.Left, "$x")
				assertions.AssertNumberLiteral(assignment.Right, "1.0")
			}).
		AddSimple("float_in_array_context", `<?php $arr = [1., 1.0, 1.23]; ?>`,
			func(ctx *testutils.TestContext) {
				assertions := testutils.NewASTAssertions(ctx.T)
				body := assertions.AssertProgramBody(ctx.Program, 1)

				exprStmt := assertions.AssertExpressionStatement(body[0])
				assignment := assertions.AssertAssignment(exprStmt.Expression, "=")
				assertions.AssertVariable(assignment.Left, "$arr")

				arrayExpr := assertions.AssertArray(assignment.Right, 3)
				assertions.AssertNumberLiteral(arrayExpr.Elements[0], "1.")
				assertions.AssertNumberLiteral(arrayExpr.Elements[1], "1.0")
				assertions.AssertNumberLiteral(arrayExpr.Elements[2], "1.23")
			}).
		Run(t)
}

// TestBasic_BinaryExpressions 基础二元表达式测试
func TestBasic_BinaryExpressions(t *testing.T) {
	suite := testutils.NewTestSuiteBuilder("BinaryExpressions", createParserFactory())

	// 算术运算符
	suite.
		AddSimple("addition", "<?php $result = 1 + 2; ?>",
			testutils.ValidateBinaryOperation("+",
				testutils.ValidateNumberArg("1"),
				testutils.ValidateNumberArg("2"))).
		AddSimple("subtraction", "<?php $result = 10 - 3; ?>",
			testutils.ValidateBinaryOperation("-",
				testutils.ValidateNumberArg("10"),
				testutils.ValidateNumberArg("3"))).
		AddSimple("multiplication", "<?php $result = 4 * 5; ?>",
			testutils.ValidateBinaryOperation("*",
				testutils.ValidateNumberArg("4"),
				testutils.ValidateNumberArg("5"))).
		AddSimple("division", "<?php $result = 20 / 4; ?>",
			testutils.ValidateBinaryOperation("/",
				testutils.ValidateNumberArg("20"),
				testutils.ValidateNumberArg("4"))).
		AddSimple("modulo", "<?php $result = 10 % 3; ?>",
			testutils.ValidateBinaryOperation("%",
				testutils.ValidateNumberArg("10"),
				testutils.ValidateNumberArg("3"))).
		AddSimple("power", "<?php $result = 2 ** 3; ?>",
			testutils.ValidateBinaryOperation("**",
				testutils.ValidateNumberArg("2"),
				testutils.ValidateNumberArg("3"))).
		Run(t)
}

// TestBasic_ComparisonExpressions 基础比较表达式测试
func TestBasic_ComparisonExpressions(t *testing.T) {
	suite := testutils.NewTestSuiteBuilder("ComparisonExpressions", createParserFactory())

	suite.
		AddSimple("equals", "<?php $result = $a == $b; ?>",
			testutils.ValidateBinaryOperation("==", nil, nil)).
		AddSimple("not_equals", "<?php $result = $a != $b; ?>",
			testutils.ValidateBinaryOperation("!=", nil, nil)).
		AddSimple("not_equals_alt", "<?php $result = $a <> $b; ?>",
			testutils.ValidateBinaryOperation("<>", nil, nil)).
		AddSimple("identical", "<?php $result = $a === $b; ?>",
			testutils.ValidateBinaryOperation("===", nil, nil)).
		AddSimple("not_identical", "<?php $result = $a !== $b; ?>",
			testutils.ValidateBinaryOperation("!==", nil, nil)).
		AddSimple("less_than", "<?php $result = $a < $b; ?>",
			testutils.ValidateBinaryOperation("<", nil, nil)).
		AddSimple("less_equal", "<?php $result = $a <= $b; ?>",
			testutils.ValidateBinaryOperation("<=", nil, nil)).
		AddSimple("greater_than", "<?php $result = $a > $b; ?>",
			testutils.ValidateBinaryOperation(">", nil, nil)).
		AddSimple("greater_equal", "<?php $result = $a >= $b; ?>",
			testutils.ValidateBinaryOperation(">=", nil, nil)).
		AddSimple("spaceship", "<?php $result = $a <=> $b; ?>",
			testutils.ValidateBinaryOperation("<=>", nil, nil)).
		Run(t)
}

// TestBasic_LogicalExpressions 基础逻辑表达式测试
func TestBasic_LogicalExpressions(t *testing.T) {
	suite := testutils.NewTestSuiteBuilder("LogicalExpressions", createParserFactory())

	suite.
		AddSimple("logical_and", "<?php $result = $a && $b; ?>",
			testutils.ValidateBinaryOperation("&&", nil, nil)).
		AddSimple("logical_or", "<?php $result = $a || $b; ?>",
			testutils.ValidateBinaryOperation("||", nil, nil)).
		AddSimple("logical_and_word", "<?php $result = $a and $b; ?>",
			testutils.ValidateBinaryOperation("and", nil, nil)).
		AddSimple("logical_or_word", "<?php $result = $a or $b; ?>",
			testutils.ValidateBinaryOperation("or", nil, nil)).
		AddSimple("logical_xor", "<?php $result = $a xor $b; ?>",
			testutils.ValidateBinaryOperation("xor", nil, nil)).
		Run(t)
}

// TestBasic_StringConcatenation 基础字符串连接测试
func TestBasic_StringConcatenation(t *testing.T) {
	suite := testutils.NewTestSuiteBuilder("StringConcatenation", createParserFactory())

	suite.
		AddSimple("simple_concat", "<?php $result = \"Hello\" . \" World\"; ?>",
			testutils.ValidateBinaryOperation(".",
				testutils.ValidateStringArg("Hello", `"Hello"`),
				testutils.ValidateStringArg(" World", `" World"`))).
		AddSimple("variable_concat", "<?php $result = $greeting . $name; ?>",
			testutils.ValidateBinaryOperation(".", nil, nil)).
		AddSimple("multiple_concat", "<?php $result = $a . $b . $c; ?>",
			testutils.ValidateBinaryOperation(".", nil, nil)). // 左结合
		Run(t)
}

// TestBasic_AssignmentOperators 基础赋值操作符测试
func TestBasic_AssignmentOperators(t *testing.T) {
	suite := testutils.NewTestSuiteBuilder("AssignmentOperators", createParserFactory())

	suite.
		AddSimple("basic_assignment", "<?php $a = 5; ?>",
			testutils.ValidateBasicAssignment("$a")).
		AddSimple("addition_assignment", "<?php $a += 5; ?>",
			testutils.ValidateAdditionAssignment("$a")).
		AddSimple("subtraction_assignment", "<?php $a -= 3; ?>",
			testutils.ValidateSubtractionAssignment("$a")).
		AddSimple("multiplication_assignment", "<?php $a *= 2; ?>",
			testutils.ValidateMultiplicationAssignment("$a")).
		AddSimple("division_assignment", "<?php $a /= 2; ?>",
			testutils.ValidateDivisionAssignment("$a")).
		AddSimple("modulo_assignment", "<?php $a %= 3; ?>",
			testutils.ValidateModuloAssignment("$a")).
		AddSimple("power_assignment", "<?php $a **= 2; ?>",
			testutils.ValidatePowerAssignment("$a")).
		AddSimple("concat_assignment", "<?php $a .= \"text\"; ?>",
			testutils.ValidateConcatenationAssignment("$a")).
		AddSimple("coalesce_assignment", "<?php $a ??= \"default\"; ?>",
			testutils.ValidateCoalesceAssignment("$a")).
		Run(t)
}

// TestBasic_BitwiseOperators 基础位运算符测试
func TestBasic_BitwiseOperators(t *testing.T) {
	suite := testutils.NewTestSuiteBuilder("BitwiseOperators", createParserFactory())

	suite.
		AddSimple("bitwise_and", "<?php $result = $a & $b; ?>",
			testutils.ValidateBinaryOperation("&", nil, nil)).
		AddSimple("bitwise_or", "<?php $result = $a | $b; ?>",
			testutils.ValidateBinaryOperation("|", nil, nil)).
		AddSimple("bitwise_xor", "<?php $result = $a ^ $b; ?>",
			testutils.ValidateBinaryOperation("^", nil, nil)).
		AddSimple("bitwise_and_assignment", "<?php $a &= $b; ?>",
			testutils.ValidateBitwiseAndAssignment("$a")).
		AddSimple("bitwise_or_assignment", "<?php $a |= $b; ?>",
			testutils.ValidateBitwiseOrAssignment("$a")).
		AddSimple("bitwise_xor_assignment", "<?php $a ^= $b; ?>",
			testutils.ValidateBitwiseXorAssignment("$a")).
		AddSimple("left_shift", "<?php $result = $a << 2; ?>",
			testutils.ValidateBinaryOperation("<<", nil, nil)).
		AddSimple("right_shift", "<?php $result = $a >> 2; ?>",
			testutils.ValidateBinaryOperation(">>", nil, nil)).
		AddSimple("left_shift_assignment", "<?php $a <<= 2; ?>",
			testutils.ValidateLeftShiftAssignment("$a")).
		AddSimple("right_shift_assignment", "<?php $a >>= 2; ?>",
			testutils.ValidateRightShiftAssignment("$a")).
		Run(t)
}
