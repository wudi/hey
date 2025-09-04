package parser

import (
	"testing"

	"github.com/wudi/php-parser/parser/testutils"
)

// TestRefactored_MethodModifierCombinations 重构后的方法修饰符组合测试
func TestRefactored_MethodModifierCombinations(t *testing.T) {
	suite := testutils.NewTestSuiteBuilder("MethodModifierCombinations", createParserFactory())

	// public final static function
	suite.AddSimple("public_final_static_function",
		`<?php
class Foo {
    public final static function isSigchildEnabled()
    {
    }
}
?>`,
		testutils.ValidateClass("Foo",
			testutils.ValidateClassMethod("isSigchildEnabled", "public")))

	// final public static function (不同顺序)
	suite.AddSimple("final_public_static_function",
		`<?php
class Bar {
    final public static function create()
    {
        return new self();
    }
}`,
		testutils.ValidateClass("Bar",
			testutils.ValidateClassMethod("create", "public")))

	// static final public function
	suite.AddSimple("static_final_public_function",
		`<?php
class Baz {
    static final public function getInstance()
    {
        return null;
    }
}`,
		testutils.ValidateClass("Baz",
			testutils.ValidateClassMethod("getInstance", "public")))

	// private final function
	suite.AddSimple("private_final_function",
		`<?php
class Test {
    private final function process()
    {
        return true;
    }
}`,
		testutils.ValidateClass("Test",
			testutils.ValidateClassMethod("process", "private")))

	// protected final static function
	suite.AddSimple("protected_final_static_function",
		`<?php
class Service {
    protected final static function validate($data)
    {
        return !empty($data);
    }
}`,
		testutils.ValidateClass("Service",
			testutils.ValidateClassMethod("validate", "protected")))

	// abstract public function
	suite.AddSimple("abstract_public_function",
		`<?php
abstract class AbstractClass {
    abstract public function execute();
}`,
		testutils.ValidateClass("AbstractClass",
			testutils.ValidateClassMethod("execute", "public")))

	// abstract protected function
	suite.AddSimple("abstract_protected_function",
		`<?php
abstract class AbstractService {
    abstract protected function process($data);
}`,
		testutils.ValidateClass("AbstractService",
			testutils.ValidateClassMethod("process", "protected")))

	// public abstract function (顺序不同)
	suite.AddSimple("public_abstract_function",
		`<?php
abstract class Handler {
    public abstract function handle($request);
}`,
		testutils.ValidateClass("Handler",
			testutils.ValidateClassMethod("handle", "public")))

	// abstract static function (如果支持)
	suite.AddSimple("abstract_static_function",
		`<?php
abstract class Factory {
    abstract static function create();
}`,
		testutils.ValidateClass("Factory",
			testutils.ValidateClassMethod("create", "")))

	// 复杂的修饰符组合
	suite.AddSimple("complex_modifiers",
		`<?php
class Complex {
    final public static function method1() {}
    abstract public function method2();
    private final function method3() {}
    protected static function method4() {}
}`,
		testutils.ValidateClass("Complex",
			testutils.ValidateClassMethod("method1", "public"),
			testutils.ValidateClassMethod("method2", "public"),
			testutils.ValidateClassMethod("method3", "private"),
			testutils.ValidateClassMethod("method4", "protected")))

	suite.Run(t)
}
