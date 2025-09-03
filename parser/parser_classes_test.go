package parser

import (
	"testing"
	
	"github.com/wudi/php-parser/parser/testutils"
)

// TestClasses_BasicDeclaration 基础类声明测试
func TestClasses_BasicDeclaration(t *testing.T) {
	suite := testutils.NewTestSuiteBuilder("BasicClassDeclaration", createParserFactory())
	
	suite.
		AddClass("empty_class", "Test", "").
		AddClass("simple_class", "User", "$name = 'default';").
		AddClass("class_with_methods", "Calculator", "function add($a, $b) { return $a + $b; }").
		AddClass("class_with_constants", "Config", "const VERSION = '1.0';").
		Run(t)
}

// TestClasses_Inheritance 类继承测试
func TestClasses_Inheritance(t *testing.T) {
	suite := testutils.NewTestSuiteBuilder("ClassInheritance", createParserFactory())
	
	suite.
		AddSimple("extends_class", "<?php class Child extends Parent { } ?>",
			testutils.ValidateClass("Child")).
		AddSimple("implements_interface", "<?php class Service implements ServiceInterface { } ?>",
			testutils.ValidateClass("Service")).
		AddSimple("extends_and_implements", "<?php class Service extends BaseService implements ServiceInterface { } ?>",
			testutils.ValidateClass("Service")).
		AddSimple("multiple_interfaces", "<?php class Handler implements HandlerInterface, LoggerInterface { } ?>",
			testutils.ValidateClass("Handler")).
		Run(t)
}

// TestClasses_Properties 类属性测试
func TestClasses_Properties(t *testing.T) {
	suite := testutils.NewTestSuiteBuilder("ClassProperties", createParserFactory())
	
	suite.
		AddSimple("public_property", "<?php class Test { public $name; } ?>",
			testutils.ValidateClass("Test", 
				testutils.ValidateProperty("name", "public"))).
		AddSimple("private_property", "<?php class Test { private $data; } ?>",
			testutils.ValidateClass("Test",
				testutils.ValidateProperty("data", "private"))).
		AddSimple("protected_property", "<?php class Test { protected $value; } ?>",
			testutils.ValidateClass("Test",
				testutils.ValidateProperty("value", "protected"))).
		AddSimple("static_property", "<?php class Test { public static $count = 0; } ?>",
			testutils.ValidateClass("Test",
				testutils.ValidateProperty("count", "public"))).
		AddSimple("typed_property", "<?php class Test { private string $name; } ?>",
			testutils.ValidateClass("Test",
				testutils.ValidateProperty("name", "private"))).
		AddSimple("readonly_property", "<?php class Test { readonly string $id; } ?>",
			testutils.ValidateClass("Test",
				testutils.ValidateProperty("id", "public"))).
		Run(t)
}

// TestClasses_Methods 类方法测试
func TestClasses_Methods(t *testing.T) {
	suite := testutils.NewTestSuiteBuilder("ClassMethods", createParserFactory())
	
	suite.
		AddSimple("public_method", "<?php class Test { public function getName() { } } ?>",
			testutils.ValidateClass("Test",
				testutils.ValidateClassMethod("getName", "public"))).
		AddSimple("private_method", "<?php class Test { private function process() { } } ?>",
			testutils.ValidateClass("Test",
				testutils.ValidateClassMethod("process", "private"))).
		AddSimple("protected_method", "<?php class Test { protected function validate() { } } ?>",
			testutils.ValidateClass("Test",
				testutils.ValidateClassMethod("validate", "protected"))).
		AddSimple("static_method", "<?php class Test { public static function create() { } } ?>",
			testutils.ValidateClass("Test",
				testutils.ValidateClassMethod("create", "public"))).
		AddSimple("abstract_method", "<?php abstract class Test { abstract function process(); } ?>",
			testutils.ValidateClass("Test")).
		AddSimple("final_method", "<?php class Test { final public function process() { } } ?>",
			testutils.ValidateClass("Test",
				testutils.ValidateClassMethod("process", "public"))).
		Run(t)
}

// TestClasses_Constants 类常量测试
func TestClasses_Constants(t *testing.T) {
	suite := testutils.NewTestSuiteBuilder("ClassConstants", createParserFactory())
	
	suite.
		AddSimple("public_constant", "<?php class Test { public const VERSION = '1.0'; } ?>",
			testutils.ValidateClass("Test",
				testutils.ValidateClassConstant("VERSION", "public"))).
		AddSimple("private_constant", "<?php class Test { private const SECRET = 'hidden'; } ?>",
			testutils.ValidateClass("Test",
				testutils.ValidateClassConstant("SECRET", "private"))).
		AddSimple("protected_constant", "<?php class Test { protected const CONFIG = []; } ?>",
			testutils.ValidateClass("Test",
				testutils.ValidateClassConstant("CONFIG", "protected"))).
		AddSimple("multiple_constants", "<?php class Test { const A = 1, B = 2; } ?>",
			testutils.ValidateClass("Test",
				testutils.ValidateClassConstant("A", ""),
				testutils.ValidateClassConstant("B", ""))).
		AddSimple("final_constant", "<?php class Test { final const LIMIT = 100; } ?>",
			testutils.ValidateClass("Test",
				testutils.ValidateClassConstant("LIMIT", ""))).
		Run(t)
}

// TestClasses_Constructor 构造函数测试
func TestClasses_Constructor(t *testing.T) {
	suite := testutils.NewTestSuiteBuilder("ClassConstructor", createParserFactory())
	
	suite.
		AddSimple("basic_constructor", "<?php class Test { public function __construct() { } } ?>",
			testutils.ValidateClass("Test",
				testutils.ValidateClassMethod("__construct", "public"))).
		AddSimple("constructor_with_params", "<?php class Test { public function __construct($name, $age) { } } ?>",
			testutils.ValidateClass("Test",
				testutils.ValidateClassMethod("__construct", "public"))).
		AddSimple("constructor_promotion", "<?php class Test { public function __construct(public string $name) { } } ?>",
			testutils.ValidateClass("Test",
				testutils.ValidateClassMethod("__construct", "public"))).
		AddSimple("private_constructor", "<?php class Test { private function __construct() { } } ?>",
			testutils.ValidateClass("Test",
				testutils.ValidateClassMethod("__construct", "private"))).
		Run(t)
}

// TestClasses_ModifierCombinations 修饰符组合测试
func TestClasses_ModifierCombinations(t *testing.T) {
	suite := testutils.NewTestSuiteBuilder("ClassModifierCombinations", createParserFactory())
	
	suite.
		AddSimple("abstract_class", "<?php abstract class Test { } ?>",
			testutils.ValidateClass("Test")).
		AddSimple("final_class", "<?php final class Test { } ?>",
			testutils.ValidateClass("Test")).
		AddSimple("readonly_class", "<?php readonly class Test { } ?>",
			testutils.ValidateClass("Test")).
		AddSimple("final_abstract_method", "<?php abstract class Test { final abstract function process(); } ?>",
			testutils.ValidateClass("Test")).
		AddSimple("static_final_method", "<?php class Test { final public static function create() { } } ?>",
			testutils.ValidateClass("Test",
				testutils.ValidateClassMethod("create", "public"))).
		Run(t)
}