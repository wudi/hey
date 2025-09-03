package parser

import (
	"testing"
	
	"github.com/wudi/php-parser/parser/testutils"
)

// TestRefactored_UnaryExpressions 重构后的一元表达式测试
func TestRefactored_UnaryExpressions(t *testing.T) {
	suite := testutils.NewTestSuiteBuilder("UnaryExpressions", createParserFactory())
	
	// 前缀递增
	suite.AddSimple("prefix_increment",
		`<?php $result = ++$i; ?>`,
		testutils.ValidatePrefixExpression("$result", "$i", "++"))
	
	// 后缀递增
	suite.AddSimple("postfix_increment",
		`<?php $result = $i++; ?>`,
		testutils.ValidatePostfixAssignment("$result", "$i", "++"))
	
	// 前缀递减
	suite.AddSimple("prefix_decrement",
		`<?php $result = --$i; ?>`,
		testutils.ValidatePrefixExpression("$result", "$i", "--"))
	
	// 后缀递减
	suite.AddSimple("postfix_decrement",
		`<?php $result = $i--; ?>`,
		testutils.ValidatePostfixAssignment("$result", "$i", "--"))
	
	// 一元正号
	suite.AddSimple("unary_plus",
		`<?php $result = +$value; ?>`,
		testutils.ValidatePrefixExpression("$result", "$value", "+"))
	
	// 一元负号
	suite.AddSimple("unary_minus",
		`<?php $result = -$value; ?>`,
		testutils.ValidatePrefixExpression("$result", "$value", "-"))
	
	// 逻辑非
	suite.AddSimple("logical_not",
		`<?php $result = !$flag; ?>`,
		testutils.ValidatePrefixExpression("$result", "$flag", "!"))
	
	// 位非
	suite.AddSimple("bitwise_not",
		`<?php $result = ~$value; ?>`,
		testutils.ValidatePrefixExpression("$result", "$value", "~"))
	
	suite.Run(t)
}

// TestRefactored_BinaryExpressions 重构后的二元表达式测试
func TestRefactored_BinaryExpressions(t *testing.T) {
	suite := testutils.NewTestSuiteBuilder("BinaryExpressions", createParserFactory())
	
	// 算术运算
	suite.AddSimple("addition",
		`<?php $result = $a + $b; ?>`,
		testutils.ValidateBinaryAssignment("$result", "$a", "+", "$b"))
	
	suite.AddSimple("subtraction", 
		`<?php $result = $a - $b; ?>`,
		testutils.ValidateBinaryAssignment("$result", "$a", "-", "$b"))
	
	suite.AddSimple("multiplication",
		`<?php $result = $a * $b; ?>`,
		testutils.ValidateBinaryAssignment("$result", "$a", "*", "$b"))
	
	suite.AddSimple("division",
		`<?php $result = $a / $b; ?>`,
		testutils.ValidateBinaryAssignment("$result", "$a", "/", "$b"))
	
	suite.AddSimple("modulus",
		`<?php $result = $a % $b; ?>`,
		testutils.ValidateBinaryAssignment("$result", "$a", "%", "$b"))
	
	suite.AddSimple("power",
		`<?php $result = $a ** $b; ?>`,
		testutils.ValidateBinaryAssignment("$result", "$a", "**", "$b"))
	
	// 比较运算
	suite.AddSimple("equal",
		`<?php $result = $a == $b; ?>`,
		testutils.ValidateBinaryAssignment("$result", "$a", "==", "$b"))
	
	suite.AddSimple("not_equal",
		`<?php $result = $a != $b; ?>`,
		testutils.ValidateBinaryAssignment("$result", "$a", "!=", "$b"))
	
	suite.AddSimple("identical",
		`<?php $result = $a === $b; ?>`,
		testutils.ValidateBinaryAssignment("$result", "$a", "===", "$b"))
	
	suite.AddSimple("not_identical",
		`<?php $result = $a !== $b; ?>`,
		testutils.ValidateBinaryAssignment("$result", "$a", "!==", "$b"))
	
	suite.AddSimple("less_than",
		`<?php $result = $a < $b; ?>`,
		testutils.ValidateBinaryAssignment("$result", "$a", "<", "$b"))
	
	suite.AddSimple("greater_than",
		`<?php $result = $a > $b; ?>`,
		testutils.ValidateBinaryAssignment("$result", "$a", ">", "$b"))
	
	suite.AddSimple("less_equal",
		`<?php $result = $a <= $b; ?>`,
		testutils.ValidateBinaryAssignment("$result", "$a", "<=", "$b"))
	
	suite.AddSimple("greater_equal",
		`<?php $result = $a >= $b; ?>`,
		testutils.ValidateBinaryAssignment("$result", "$a", ">=", "$b"))
	
	suite.AddSimple("spaceship",
		`<?php $result = $a <=> $b; ?>`,
		testutils.ValidateBinaryAssignment("$result", "$a", "<=>", "$b"))
	
	// 逻辑运算
	suite.AddSimple("logical_and",
		`<?php $result = $a && $b; ?>`,
		testutils.ValidateBinaryAssignment("$result", "$a", "&&", "$b"))
	
	suite.AddSimple("logical_or",
		`<?php $result = $a || $b; ?>`,
		testutils.ValidateBinaryAssignment("$result", "$a", "||", "$b"))
	
	// 位运算
	suite.AddSimple("bitwise_and",
		`<?php $result = $a & $b; ?>`,
		testutils.ValidateBinaryAssignment("$result", "$a", "&", "$b"))
	
	suite.AddSimple("bitwise_or",
		`<?php $result = $a | $b; ?>`,
		testutils.ValidateBinaryAssignment("$result", "$a", "|", "$b"))
	
	suite.AddSimple("bitwise_xor",
		`<?php $result = $a ^ $b; ?>`,
		testutils.ValidateBinaryAssignment("$result", "$a", "^", "$b"))
	
	suite.AddSimple("left_shift",
		`<?php $result = $a << $b; ?>`,
		testutils.ValidateBinaryAssignment("$result", "$a", "<<", "$b"))
	
	suite.AddSimple("right_shift",
		`<?php $result = $a >> $b; ?>`,
		testutils.ValidateBinaryAssignment("$result", "$a", ">>", "$b"))
	
	// 字符串连接
	suite.AddSimple("string_concatenation",
		`<?php $result = $a . $b; ?>`,
		testutils.ValidateBinaryAssignment("$result", "$a", ".", "$b"))
	
	// instanceof
	suite.AddSimple("instanceof",
		`<?php $result = $obj instanceof MyClass; ?>`,
		testutils.ValidateInstanceofExpression("$result", "$obj", "MyClass"))
	
	suite.Run(t)
}

// TestRefactored_TernaryExpressions 重构后的三元表达式测试
func TestRefactored_TernaryExpressions(t *testing.T) {
	suite := testutils.NewTestSuiteBuilder("TernaryExpressions", createParserFactory())
	
	// 基础三元表达式
	suite.AddSimple("basic_ternary",
		`<?php $result = $condition ? $true_val : $false_val; ?>`,
		testutils.ValidateTernaryExpression("$result", "$condition", "$true_val", "$false_val"))
	
	// 空合并运算符
	suite.AddSimple("null_coalescing",
		`<?php $result = $value ?? $default; ?>`,
		testutils.ValidateCoalesceExpression("$result", "$value", "$default"))
	
	// 空合并赋值运算符
	suite.AddSimple("null_coalescing_assignment",
		`<?php $value ??= $default; ?>`,
		testutils.ValidateAssignmentOperation("$value", "??=", "$default"))
	
	suite.Run(t)
}

// TestRefactored_AssignmentExpressions 重构后的赋值表达式测试
func TestRefactored_AssignmentExpressions(t *testing.T) {
	suite := testutils.NewTestSuiteBuilder("AssignmentExpressions", createParserFactory())
	
	// 复合赋值运算符
	suite.AddSimple("addition_assignment",
		`<?php $a += $b; ?>`,
		testutils.ValidateAssignmentOperation("$a", "+=", "$b"))
	
	suite.AddSimple("subtraction_assignment",
		`<?php $a -= $b; ?>`,
		testutils.ValidateAssignmentOperation("$a", "-=", "$b"))
	
	suite.AddSimple("multiplication_assignment",
		`<?php $a *= $b; ?>`,
		testutils.ValidateAssignmentOperation("$a", "*=", "$b"))
	
	suite.AddSimple("division_assignment",
		`<?php $a /= $b; ?>`,
		testutils.ValidateAssignmentOperation("$a", "/=", "$b"))
	
	suite.AddSimple("modulus_assignment",
		`<?php $a %= $b; ?>`,
		testutils.ValidateAssignmentOperation("$a", "%=", "$b"))
	
	suite.AddSimple("power_assignment",
		`<?php $a **= $b; ?>`,
		testutils.ValidateAssignmentOperation("$a", "**=", "$b"))
	
	suite.AddSimple("concatenation_assignment",
		`<?php $a .= $b; ?>`,
		testutils.ValidateAssignmentOperation("$a", ".=", "$b"))
	
	suite.AddSimple("bitwise_and_assignment",
		`<?php $a &= $b; ?>`,
		testutils.ValidateAssignmentOperation("$a", "&=", "$b"))
	
	suite.AddSimple("bitwise_or_assignment",
		`<?php $a |= $b; ?>`,
		testutils.ValidateAssignmentOperation("$a", "|=", "$b"))
	
	suite.AddSimple("bitwise_xor_assignment",
		`<?php $a ^= $b; ?>`,
		testutils.ValidateAssignmentOperation("$a", "^=", "$b"))
	
	suite.AddSimple("left_shift_assignment",
		`<?php $a <<= $b; ?>`,
		testutils.ValidateAssignmentOperation("$a", "<<=", "$b"))
	
	suite.AddSimple("right_shift_assignment",
		`<?php $a >>= $b; ?>`,
		testutils.ValidateAssignmentOperation("$a", ">>=", "$b"))
	
	suite.Run(t)
}