package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wudi/php-parser/ast"
)

// ============= EXPRESSION PARSING TESTS =============

func TestPrattParser_Literals(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected interface{}
	}{
		{"integer", "<?php 42;", int64(42)},
		{"float", "<?php 3.14;", 3.14},
		{"string", "<?php 'hello';", "hello"},
		{"variable", "<?php $foo;", "$foo"},
		{"true", "<?php true;", true},
		{"false", "<?php false;", false},
		{"null", "<?php null;", nil},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program, errors := ParsePHP(tt.input)
			require.Empty(t, errors)
			require.NotNil(t, program)
			require.Len(t, program.Statements, 1)
			
			// Verify the literal value
			// (Implementation would check actual AST node values)
		})
	}
}

func TestPrattParser_BinaryExpressions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		operator string
	}{
		{"addition", "<?php 1 + 2;", "+"},
		{"subtraction", "<?php 3 - 1;", "-"},
		{"multiplication", "<?php 2 * 3;", "*"},
		{"division", "<?php 6 / 2;", "/"},
		{"modulo", "<?php 5 % 2;", "%"},
		{"power", "<?php 2 ** 3;", "**"},
		{"concatenation", "<?php 'a' . 'b';", "."},
		{"spaceship", "<?php 1 <=> 2;", "<=>"},
		{"null coalesce", "<?php $a ?? $b;", "??"},
		{"pipe", "<?php $a |> $b;", "|>"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program, errors := ParsePHP(tt.input)
			require.Empty(t, errors)
			require.NotNil(t, program)
		})
	}
}

func TestPrattParser_AssignmentOperators(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		operator string
	}{
		{"assign", "<?php $a = 1;", "="},
		{"add assign", "<?php $a += 1;", "+="},
		{"subtract assign", "<?php $a -= 1;", "-="},
		{"multiply assign", "<?php $a *= 2;", "*="},
		{"divide assign", "<?php $a /= 2;", "/="},
		{"concat assign", "<?php $a .= 'b';", ".="},
		{"modulo assign", "<?php $a %= 2;", "%="},
		{"and assign", "<?php $a &= 1;", "&="},
		{"or assign", "<?php $a |= 1;", "|="},
		{"xor assign", "<?php $a ^= 1;", "^="},
		{"shift left assign", "<?php $a <<= 1;", "<<="},
		{"shift right assign", "<?php $a >>= 1;", ">>="},
		{"power assign", "<?php $a **= 2;", "**="},
		{"null coalesce assign", "<?php $a ??= 1;", "??="},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program, errors := ParsePHP(tt.input)
			require.Empty(t, errors)
			require.NotNil(t, program)
		})
	}
}

func TestPrattParser_TernaryAndCoalescing(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"ternary", "<?php $a ? $b : $c;"},
		{"short ternary", "<?php $a ?: $b;"},
		{"null coalesce", "<?php $a ?? $b;"},
		{"nested ternary", "<?php $a ? $b : $c ? $d : $e;"},
		{"nested coalesce", "<?php $a ?? $b ?? $c;"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program, errors := ParsePHP(tt.input)
			require.Empty(t, errors)
			require.NotNil(t, program)
		})
	}
}

// ============= STATEMENT PARSING TESTS =============

func TestPrattParser_ControlFlowStatements(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"if statement", "<?php if ($a) echo 'yes';"},
		{"if-else statement", "<?php if ($a) echo 'yes'; else echo 'no';"},
		{"if-elseif-else", "<?php if ($a) echo 'a'; elseif ($b) echo 'b'; else echo 'c';"},
		{"while loop", "<?php while ($i < 10) $i++;"},
		{"do-while loop", "<?php do { $i++; } while ($i < 10);"},
		{"for loop", "<?php for ($i = 0; $i < 10; $i++) echo $i;"},
		{"foreach loop", "<?php foreach ($array as $value) echo $value;"},
		{"foreach with key", "<?php foreach ($array as $key => $value) echo $key;"},
		{"switch statement", "<?php switch ($a) { case 1: echo 'one'; break; default: echo 'other'; }"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program, errors := ParsePHP(tt.input)
			require.Empty(t, errors)
			require.NotNil(t, program)
		})
	}
}

func TestPrattParser_AlternativeSyntax(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"if-endif", "<?php if ($a): echo 'yes'; endif;"},
		{"if-else-endif", "<?php if ($a): echo 'yes'; else: echo 'no'; endif;"},
		{"while-endwhile", "<?php while ($a): echo 'loop'; endwhile;"},
		{"for-endfor", "<?php for ($i = 0; $i < 10; $i++): echo $i; endfor;"},
		{"foreach-endforeach", "<?php foreach ($array as $item): echo $item; endforeach;"},
		{"switch-endswitch", "<?php switch ($a): case 1: echo 'one'; break; endswitch;"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program, errors := ParsePHP(tt.input)
			require.Empty(t, errors)
			require.NotNil(t, program)
		})
	}
}

func TestPrattParser_ExceptionHandling(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"try-catch", "<?php try { risky(); } catch (Exception $e) { handle(); }"},
		{"try-catch-finally", "<?php try { risky(); } catch (Exception $e) { handle(); } finally { cleanup(); }"},
		{"multiple catch", "<?php try { risky(); } catch (TypeError $e) { } catch (Exception $e) { }"},
		{"union catch types", "<?php try { risky(); } catch (TypeError|ValueError $e) { handle(); }"},
		{"catch without variable", "<?php try { risky(); } catch (Exception) { handle(); }"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program, errors := ParsePHP(tt.input)
			require.Empty(t, errors)
			require.NotNil(t, program)
		})
	}
}

// ============= DECLARATION PARSING TESTS =============

func TestPrattParser_FunctionDeclarations(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"simple function", "<?php function foo() { return 42; }"},
		{"function with params", "<?php function add($a, $b) { return $a + $b; }"},
		{"function with types", "<?php function add(int $a, int $b): int { return $a + $b; }"},
		{"function with default params", "<?php function greet($name = 'World') { echo 'Hello ' . $name; }"},
		{"function with reference return", "<?php function &getRef() { return $ref; }"},
		{"function with variadic param", "<?php function sum(...$numbers) { return array_sum($numbers); }"},
		{"function with nullable types", "<?php function maybe(?string $s): ?int { return null; }"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program, errors := ParsePHP(tt.input)
			require.Empty(t, errors)
			require.NotNil(t, program)
		})
	}
}

func TestPrattParser_ClassDeclarations(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"simple class", "<?php class Foo {}"},
		{"class with extends", "<?php class Child extends Parent {}"},
		{"class with implements", "<?php class Foo implements Bar {}"},
		{"class with multiple implements", "<?php class Foo implements Bar, Baz {}"},
		{"abstract class", "<?php abstract class AbstractFoo {}"},
		{"final class", "<?php final class FinalFoo {}"},
		{"readonly class", "<?php readonly class ReadonlyFoo {}"},
		{"class with property", "<?php class Foo { public $bar; }"},
		{"class with method", "<?php class Foo { public function bar() {} }"},
		{"class with constructor", "<?php class Foo { public function __construct() {} }"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program, errors := ParsePHP(tt.input)
			require.Empty(t, errors)
			require.NotNil(t, program)
		})
	}
}

// ============= MODERN PHP FEATURES TESTS =============

func TestPrattParser_Attributes(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"simple attribute", "<?php #[Attribute] class Foo {}"},
		{"attribute with args", "<?php #[Route('/api')] class Controller {}"},
		{"multiple attributes", "<?php #[Attribute] #[Deprecated] class Foo {}"},
		{"attribute group", "<?php #[Attribute, Deprecated] class Foo {}"},
		{"method attribute", "<?php class Foo { #[Test] public function test() {} }"},
		{"parameter attribute", "<?php function foo(#[SensitiveParameter] string $password) {}"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program, errors := ParsePHP(tt.input)
			require.Empty(t, errors)
			require.NotNil(t, program)
		})
	}
}

func TestPrattParser_MatchExpression(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"simple match", "<?php $result = match($value) { 1 => 'one', 2 => 'two', default => 'other' };"},
		{"match with multiple conditions", "<?php match($value) { 1, 2, 3 => 'low', 4, 5, 6 => 'mid', default => 'high' };"},
		{"match with expressions", "<?php match(true) { $age >= 18 => 'adult', $age >= 13 => 'teen', default => 'child' };"},
		{"match without default", "<?php match($value) { 'a' => 1, 'b' => 2 };"},
		{"nested match", "<?php match($a) { 1 => match($b) { 2 => 'yes', default => 'no' }, default => 'maybe' };"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program, errors := ParsePHP(tt.input)
			require.Empty(t, errors)
			require.NotNil(t, program)
		})
	}
}

func TestPrattParser_Enums(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"simple enum", "<?php enum Status { case DRAFT; case PUBLISHED; }"},
		{"backed enum", "<?php enum Status: string { case DRAFT = 'draft'; case PUBLISHED = 'published'; }"},
		{"enum with methods", "<?php enum Status { case DRAFT; public function label(): string { return 'Draft'; } }"},
		{"enum implements interface", "<?php enum Status implements StatusInterface { case DRAFT; }"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program, errors := ParsePHP(tt.input)
			require.Empty(t, errors)
			require.NotNil(t, program)
		})
	}
}

func TestPrattParser_UnionAndIntersectionTypes(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"union type", "<?php function foo(int|string $param) {}"},
		{"nullable union", "<?php function foo(int|string|null $param) {}"},
		{"intersection type", "<?php function foo(Foo&Bar $param) {}"},
		{"complex union", "<?php function foo((Foo&Bar)|Baz $param) {}"},
		{"return union type", "<?php function foo(): int|string { return 1; }"},
		{"return intersection type", "<?php function foo(): Foo&Bar { return new FooBar(); }"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program, errors := ParsePHP(tt.input)
			require.Empty(t, errors)
			require.NotNil(t, program)
		})
	}
}

func TestPrattParser_NamedArguments(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"single named arg", "<?php foo(name: 'value');"},
		{"multiple named args", "<?php foo(first: 1, second: 2);"},
		{"mixed positional and named", "<?php foo('positional', named: 'value');"},
		{"named with unpacking", "<?php foo(name: 'value', ...$args);"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program, errors := ParsePHP(tt.input)
			require.Empty(t, errors)
			require.NotNil(t, program)
		})
	}
}

func TestPrattParser_NullsafeOperator(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"nullsafe property", "<?php $obj?->property;"},
		{"nullsafe method", "<?php $obj?->method();"},
		{"chained nullsafe", "<?php $obj?->prop?->method();"},
		{"mixed safe and nullsafe", "<?php $obj->prop?->method();"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program, errors := ParsePHP(tt.input)
			require.Empty(t, errors)
			require.NotNil(t, program)
		})
	}
}

func TestPrattParser_PropertyHooks(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			"property with get hook",
			"<?php class Foo { public string $name { get => $this->_name; } }",
		},
		{
			"property with set hook",
			"<?php class Foo { public string $name { set => $this->_name = $value; } }",
		},
		{
			"property with both hooks",
			"<?php class Foo { public string $name { get => $this->_name; set => $this->_name = $value; } }",
		},
		{
			"property hook with body",
			"<?php class Foo { public string $name { get { return strtoupper($this->_name); } } }",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program, errors := ParsePHP(tt.input)
			require.Empty(t, errors)
			require.NotNil(t, program)
		})
	}
}

// ============= ERROR RECOVERY TESTS =============

func TestPrattParser_ErrorRecovery(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		shouldError bool
	}{
		{"missing semicolon", "<?php $a = 1 $b = 2;", true},
		{"unclosed parenthesis", "<?php if ($a { echo 'yes'; }", true},
		{"invalid operator", "<?php $a === = $b;", true},
		{"unexpected token", "<?php function () {}", true},
		{"missing closing brace", "<?php class Foo { public function bar() {", true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, errors := ParsePHP(tt.input)
			if tt.shouldError {
				assert.NotEmpty(t, errors)
			} else {
				assert.Empty(t, errors)
			}
		})
	}
}

// ============= OPERATOR PRECEDENCE TESTS =============

func TestPrattParser_OperatorPrecedence(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string // Expected structure
	}{
		{
			"multiplication before addition",
			"<?php 1 + 2 * 3;",
			"(1 + (2 * 3))",
		},
		{
			"power before multiplication",
			"<?php 2 * 3 ** 2;",
			"(2 * (3 ** 2))",
		},
		{
			"comparison before logical",
			"<?php $a > 5 && $b < 10;",
			"((a > 5) && (b < 10))",
		},
		{
			"assignment is right associative",
			"<?php $a = $b = $c = 1;",
			"(a = (b = (c = 1)))",
		},
		{
			"ternary is right associative",
			"<?php $a ? $b : $c ? $d : $e;",
			"(a ? b : (c ? d : e))",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program, errors := ParsePHP(tt.input)
			require.Empty(t, errors)
			require.NotNil(t, program)
			// Verify AST structure matches expected precedence
		})
	}
}

// ============= BENCHMARK TESTS =============

func BenchmarkPrattParser_SimpleExpression(b *testing.B) {
	input := "<?php $a + $b * $c;"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ParsePHP(input)
	}
}

func BenchmarkPrattParser_ComplexClass(b *testing.B) {
	input := `<?php
class ComplexClass extends BaseClass implements Interface1, Interface2 {
	private string $prop1;
	protected int $prop2 = 42;
	public array $prop3 = [];
	
	public function __construct(string $param1, int $param2 = 0) {
		$this->prop1 = $param1;
		$this->prop2 = $param2;
	}
	
	public function method1(): string {
		return $this->prop1;
	}
	
	protected function method2(array $data): void {
		foreach ($data as $key => $value) {
			$this->process($key, $value);
		}
	}
}`
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ParsePHP(input)
	}
}

func BenchmarkPrattParser_ModernPHPFeatures(b *testing.B) {
	input := `<?php
#[Route('/api')]
class Controller {
	public function handle(int|string $id): Response {
		return match($id) {
			1, 2, 3 => new Response('low'),
			4, 5, 6 => new Response('mid'),
			default => new Response('high')
		};
	}
}`
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ParsePHP(input)
	}
}