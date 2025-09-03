package parser

import (
	"testing"
	
	"github.com/wudi/php-parser/ast"
	"github.com/wudi/php-parser/parser/testutils"
)

// TestFunctions_BasicDeclaration 基础函数声明测试
func TestFunctions_BasicDeclaration(t *testing.T) {
	suite := testutils.NewTestSuiteBuilder("BasicFunctionDeclaration", createParserFactory())
	
	suite.
		AddFunction("no_params", "test", []string{},
			func(funcDecl *ast.FunctionDeclaration, t *testing.T) {
				// 可以添加更多验证
			}).
		AddFunction("single_param", "greet", []string{"$name"}).
		AddFunction("multiple_params", "calculate", []string{"$a", "$b", "$c"}).
		AddFunction("typed_param", "process", []string{"string $data"}).
		AddFunction("default_param", "init", []string{"$value = 42"}).
		AddFunction("mixed_params", "complex", []string{"string $name", "$age = 18", "bool $active = true"}).
		Run(t)
}

// TestFunctions_ReturnTypes 函数返回类型测试
func TestFunctions_ReturnTypes(t *testing.T) {
	suite := testutils.NewTestSuiteBuilder("FunctionReturnTypes", createParserFactory())
	
	suite.
		AddSimple("string_return", "<?php function getName(): string { return 'test'; } ?>",
			testutils.ValidateFunction("getName", 0)).
		AddSimple("int_return", "<?php function getAge(): int { return 25; } ?>",
			testutils.ValidateFunction("getAge", 0)).
		AddSimple("bool_return", "<?php function isActive(): bool { return true; } ?>",
			testutils.ValidateFunction("isActive", 0)).
		AddSimple("void_return", "<?php function process(): void { echo 'done'; } ?>",
			testutils.ValidateFunction("process", 0)).
		AddSimple("nullable_return", "<?php function getData(): ?array { return null; } ?>",
			testutils.ValidateFunction("getData", 0)).
		AddSimple("union_return", "<?php function getValue(): int|string { return 42; } ?>",
			testutils.ValidateFunction("getValue", 0)).
		Run(t)
}

// TestFunctions_ByReference 引用函数测试
func TestFunctions_ByReference(t *testing.T) {
	suite := testutils.NewTestSuiteBuilder("ByReferenceFunctions", createParserFactory())
	
	suite.
		AddSimple("ref_function", "<?php function &getReference() { global $data; return $data; } ?>",
			testutils.ValidateFunction("getReference", 0)).
		AddSimple("ref_param", "<?php function modify(&$data) { $data = 'modified'; } ?>",
			testutils.ValidateFunction("modify", 1)).
		AddSimple("mixed_ref_params", "<?php function update($id, &$data, $options = []) { } ?>",
			testutils.ValidateFunction("update", 3)).
		Run(t)
}

// TestFunctions_Variadic 变参函数测试  
func TestFunctions_Variadic(t *testing.T) {
	suite := testutils.NewTestSuiteBuilder("VariadicFunctions", createParserFactory())
	
	suite.
		AddSimple("simple_variadic", "<?php function sum(...$numbers) { } ?>",
			testutils.ValidateFunction("sum", 1)).
		AddSimple("typed_variadic", "<?php function process(string ...$items) { } ?>",
			testutils.ValidateFunction("process", 1)).
		AddSimple("mixed_variadic", "<?php function handle($required, ...$optional) { } ?>",
			testutils.ValidateFunction("handle", 2)).
		Run(t)
}

// TestFunctions_Anonymous 匿名函数测试
func TestFunctions_Anonymous(t *testing.T) {
	suite := testutils.NewTestSuiteBuilder("AnonymousFunctions", createParserFactory())
	
	suite.
		AddSimple("simple_closure", "<?php $fn = function() { return 42; }; ?>",
			testutils.ValidateVariable("$fn")).
		AddSimple("closure_with_params", "<?php $fn = function($a, $b) { return $a + $b; }; ?>",
			testutils.ValidateVariable("$fn")).
		AddSimple("closure_with_use", "<?php $fn = function($x) use ($y) { return $x + $y; }; ?>",
			testutils.ValidateVariable("$fn")).
		AddSimple("closure_with_ref_use", "<?php $fn = function($x) use (&$y) { $y = $x; }; ?>",
			testutils.ValidateVariable("$fn")).
		AddSimple("arrow_function", "<?php $fn = fn($x) => $x * 2; ?>",
			testutils.ValidateVariable("$fn")).
		AddSimple("arrow_function_complex", "<?php $fn = fn($a, $b) => $a + $b + $external; ?>",
			testutils.ValidateVariable("$fn")).
		Run(t)
}