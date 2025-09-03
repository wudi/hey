package parser

import (
	"testing"
	
	"github.com/wudi/php-parser/parser/testutils"
)

// TestRefactored_StaticMethods 重构后的静态方法测试
func TestRefactored_StaticMethods(t *testing.T) {
	suite := testutils.NewTestSuiteBuilder("StaticMethods", createParserFactory())
	
	// 基础静态方法
	suite.AddSimple("static_public_function",
		`<?php
class MyClass {
    static public function fromArray($array) {
        return 1;
    }
}`,
		testutils.ValidateClass("MyClass",
			testutils.ValidateClassMethod("fromArray", "public")))
	
	// 不同顺序的修饰符
	suite.AddSimple("public_static_function",
		`<?php
class MyClass {
    public static function create() {
        return new self();
    }
}`,
		testutils.ValidateClass("MyClass",
			testutils.ValidateClassMethod("create", "public")))
	
	// 私有静态方法
	suite.AddSimple("private_static_function",
		`<?php
class MyClass {
    private static function validateInput($data) {
        return true;
    }
}`,
		testutils.ValidateClass("MyClass",
			testutils.ValidateClassMethod("validateInput", "private")))
	
	// 受保护的静态方法
	suite.AddSimple("protected_static_function",
		`<?php
class MyClass {
    protected static function processData($data) {
        return $data;
    }
}`,
		testutils.ValidateClass("MyClass",
			testutils.ValidateClassMethod("processData", "protected")))
	
	// 带参数的静态方法
	suite.AddSimple("static_method_with_parameters",
		`<?php
class Calculator {
    public static function add($a, $b, $c = 0) {
        return $a + $b + $c;
    }
}`,
		testutils.ValidateClass("Calculator",
			testutils.ValidateClassMethod("add", "public")))
	
	// 带返回类型的静态方法
	suite.AddSimple("static_method_with_return_type",
		`<?php
class Factory {
    public static function createInstance(): self {
        return new self();
    }
}`,
		testutils.ValidateClass("Factory",
			testutils.ValidateClassMethod("createInstance", "public")))
			
	// final静态方法
	suite.AddSimple("final_static_method",
		`<?php
class BaseClass {
    final public static function getInstance() {
        return new static();
    }
}`,
		testutils.ValidateClass("BaseClass",
			testutils.ValidateClassMethod("getInstance", "public")))
	
	suite.Run(t)
}