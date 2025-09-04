package parser

import (
	"testing"

	"github.com/wudi/php-parser/parser/testutils"
)

// TestRefactored_ClassConstantsWithModifiers 重构后的类常量修饰符测试
func TestRefactored_ClassConstantsWithModifiers(t *testing.T) {
	suite := testutils.NewTestSuiteBuilder("ClassConstantsWithModifiers", createParserFactory())

	// 基础常量（无可见性修饰符）
	suite.AddSimple("basic_const_without_visibility",
		`<?php
class Test {
    const BASIC = 1;
}`,
		testutils.ValidateClass("Test",
			testutils.ValidateClassConstant("BASIC", ""))) // 默认为public

	// 可见性修饰符测试
	suite.AddSimple("public_const",
		`<?php
class Test {
    public const PUBLIC_CONST = 1;
}`,
		testutils.ValidateClass("Test",
			testutils.ValidateClassConstant("PUBLIC_CONST", "public")))

	suite.AddSimple("private_const",
		`<?php
class Test {
    private const PRIVATE_CONST = 2;
}`,
		testutils.ValidateClass("Test",
			testutils.ValidateClassConstant("PRIVATE_CONST", "private")))

	suite.AddSimple("protected_const",
		`<?php
class Test {
    protected const PROTECTED_CONST = 3;
}`,
		testutils.ValidateClass("Test",
			testutils.ValidateClassConstant("PROTECTED_CONST", "protected")))

	// final修饰符测试
	suite.AddSimple("final_const",
		`<?php
class Test {
    final const FINAL_BASIC = 1;
}`,
		testutils.ValidateClass("Test",
			testutils.ValidateClassConstant("FINAL_BASIC", "")))

	suite.AddSimple("final_public_const",
		`<?php
class Test {
    final public const FINAL_PUBLIC = 2;
}`,
		testutils.ValidateClass("Test",
			testutils.ValidateClassConstant("FINAL_PUBLIC", "public")))

	suite.AddSimple("final_protected_const",
		`<?php
class Test {
    final protected const FINAL_PROTECTED = 3;
}`,
		testutils.ValidateClass("Test",
			testutils.ValidateClassConstant("FINAL_PROTECTED", "protected")))

	// 多个常量在一个声明中
	suite.AddSimple("multiple_constants_basic",
		`<?php
class Test {
    const A = 1, B = 2, C = 3;
}`,
		testutils.ValidateClass("Test",
			testutils.ValidateClassConstant("A", ""),
			testutils.ValidateClassConstant("B", ""),
			testutils.ValidateClassConstant("C", "")))

	suite.AddSimple("multiple_constants_public",
		`<?php
class Test {
    public const X = 'x', Y = 'y';
}`,
		testutils.ValidateClass("Test",
			testutils.ValidateClassConstant("X", "public"),
			testutils.ValidateClassConstant("Y", "public")))

	suite.AddSimple("multiple_constants_final_protected",
		`<?php
class Test {
    final protected const P = true, Q = false;
}`,
		testutils.ValidateClass("Test",
			testutils.ValidateClassConstant("P", "protected"),
			testutils.ValidateClassConstant("Q", "protected")))

	// 原始失败用例 - final protected const
	suite.AddSimple("original_failing_case",
		`<?php
class BaseUri {
    final protected const WHATWG_SPECIAL_SCHEMES = ['ftp' => 1, 'http' => 1];
}`,
		testutils.ValidateClass("BaseUri",
			testutils.ValidateClassConstant("WHATWG_SPECIAL_SCHEMES", "protected")))

	suite.Run(t)
}
