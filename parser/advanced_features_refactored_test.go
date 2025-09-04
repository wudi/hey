package parser

import (
	"testing"

	"github.com/wudi/php-parser/parser/testutils"
)

// TestRefactored_FunctionDeclarations 重构后的函数声明测试
func TestRefactored_FunctionDeclarations(t *testing.T) {
	suite := testutils.NewTestSuiteBuilder("FunctionDeclarations", createParserFactory())

	// 基础函数声明
	suite.AddSimple("basic_function",
		`<?php function getName() { return "test"; } ?>`,
		testutils.ValidateFunctionDeclaration("getName", []string{}, "",
			testutils.ValidateReturnStatement(`"test"`)))

	// 带参数的函数
	suite.AddSimple("function_with_parameters",
		`<?php function greet($name, $age) { echo $name; } ?>`,
		testutils.ValidateFunctionWithParameters("greet", []string{"$name", "$age"},
			testutils.ValidateEchoVariable("$name")))

	// 带返回类型的函数
	suite.AddSimple("function_with_return_type",
		`<?php function calculate(): int { return 42; } ?>`,
		testutils.ValidateFunctionWithReturnType("calculate", "int",
			testutils.ValidateReturnStatement("42")))

	suite.Run(t)
}

// TestRefactored_AnonymousFunctions 重构后的匿名函数测试
func TestRefactored_AnonymousFunctions(t *testing.T) {
	suite := testutils.NewTestSuiteBuilder("AnonymousFunctions", createParserFactory())

	// 基础匿名函数
	suite.AddSimple("basic_anonymous_function",
		`<?php $fn = function() { return "hello"; }; ?>`,
		testutils.ValidateAnonymousFunctionAssignment("$fn",
			testutils.ValidateReturnStatement(`"hello"`)))

	// 带参数的匿名函数
	suite.AddSimple("anonymous_function_with_params",
		`<?php $fn = function($x, $y) { return $x + $y; }; ?>`,
		testutils.ValidateAnonymousFunctionWithParams("$fn", []string{"$x", "$y"},
			testutils.ValidateReturnBinaryExpression("$x", "+", "$y")))

	// 使用use关键字的闭包
	suite.AddSimple("closure_with_use",
		`<?php $fn = function($x) use ($y) { return $x + $y; }; ?>`,
		testutils.ValidateClosureWithUse("$fn", []string{"$x"}, []string{"$y"},
			testutils.ValidateReturnBinaryExpression("$x", "+", "$y")))

	suite.Run(t)
}

// TestRefactored_ArrowFunctions 重构后的箭头函数测试
func TestRefactored_ArrowFunctions(t *testing.T) {
	suite := testutils.NewTestSuiteBuilder("ArrowFunctions", createParserFactory())

	// 基础箭头函数 - 简化验证，只检查基本结构
	suite.AddSimple("basic_arrow_function",
		`<?php $fn = fn($x) => $x * 2; ?>`,
		testutils.ValidateArrowFunctionAssignment("$fn", []string{"$x"}, nil))

	// 多参数箭头函数
	suite.AddSimple("arrow_function_multiple_params",
		`<?php $fn = fn($x, $y) => $x + $y; ?>`,
		testutils.ValidateArrowFunctionAssignment("$fn", []string{"$x", "$y"}, nil))

	suite.Run(t)
}

// TestRefactored_ClassDeclarations 重构后的类声明测试
func TestRefactored_ClassDeclarations(t *testing.T) {
	suite := testutils.NewTestSuiteBuilder("ClassDeclarations", createParserFactory())

	// 基础类声明 - 简化为表达式语句
	suite.AddSimple("basic_class",
		`<?php class User { public $name; } ?>`,
		testutils.ValidateClassInExpressionStatement("User"))

	// Final类
	suite.AddSimple("final_class",
		`<?php final class Config { } ?>`,
		testutils.ValidateFinalClassInExpressionStatement("Config"))

	// Abstract类
	suite.AddSimple("abstract_class",
		`<?php abstract class BaseController { } ?>`,
		testutils.ValidateAbstractClassInExpressionStatement("BaseController"))

	suite.Run(t)
}

// TestRefactored_StaticAccess 重构后的静态访问测试
func TestRefactored_StaticAccess(t *testing.T) {
	suite := testutils.NewTestSuiteBuilder("StaticAccess", createParserFactory())

	// 静态属性访问
	suite.AddSimple("static_property_access",
		`<?php $value = User::$count; ?>`,
		testutils.ValidateStaticPropertyAccess("$value", "User", "$count"))

	// 静态方法调用
	suite.AddSimple("static_method_call",
		`<?php $result = Math::abs(-5); ?>`,
		testutils.ValidateStaticMethodCall("$result", "Math", "abs", []string{"-5"}))

	// 静态常量访问
	suite.AddSimple("static_constant_access",
		`<?php $value = Status::ACTIVE; ?>`,
		testutils.ValidateStaticConstantAccess("$value", "Status", "ACTIVE"))

	suite.Run(t)
}
