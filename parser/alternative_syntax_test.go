package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wudi/php-parser/ast"
	"github.com/wudi/php-parser/lexer"
)

func TestParsing_AlternativeIfStatement(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedErrors int
	}{
		{
			name: "Alternative If with elseif and else",
			input: `<?php
if ($condition):
    echo "true";
elseif ($other):
    echo "other";
else:
    echo "false";
endif;
?>`,
			expectedErrors: 0,
		},
		{
			name: "Alternative If simple",
			input: `<?php
if ($x > 0):
    echo "positive";
endif;
?>`,
			expectedErrors: 0,
		},
		{
			name: "Alternative If with multiple elseifs",
			input: `<?php
if ($x == 1):
    echo "one";
elseif ($x == 2):
    echo "two";
elseif ($x == 3):
    echo "three";
else:
    echo "other";
endif;
?>`,
			expectedErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()

			assert.Equal(t, tt.expectedErrors, len(p.Errors()), "Parser errors: %v", p.Errors())
			assert.NotNil(t, program)
			assert.Greater(t, len(program.Body), 0, "Expected at least one statement")

			// Check that we have an AlternativeIfStatement
			stmt := program.Body[0]
			altIfStmt, ok := stmt.(*ast.AlternativeIfStatement)
			assert.True(t, ok, "Expected AlternativeIfStatement, got %T", stmt)
			assert.NotNil(t, altIfStmt.Condition, "Expected condition to be set")
		})
	}
}

func TestParsing_AlternativeWhileStatement(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedErrors int
	}{
		{
			name: "Alternative While loop",
			input: `<?php
while ($counter < 10):
    echo $counter;
    $counter++;
endwhile;
?>`,
			expectedErrors: 0,
		},
		{
			name: "Alternative While with complex condition",
			input: `<?php
while ($i < count($array) && $flag):
    process($array[$i]);
    $i++;
endwhile;
?>`,
			expectedErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()

			assert.Equal(t, tt.expectedErrors, len(p.Errors()), "Parser errors: %v", p.Errors())
			assert.NotNil(t, program)
			assert.Greater(t, len(program.Body), 0, "Expected at least one statement")

			// Check that we have an AlternativeWhileStatement
			stmt := program.Body[0]
			altWhileStmt, ok := stmt.(*ast.AlternativeWhileStatement)
			assert.True(t, ok, "Expected AlternativeWhileStatement, got %T", stmt)
			assert.NotNil(t, altWhileStmt.Condition, "Expected condition to be set")
			assert.Greater(t, len(altWhileStmt.Body), 0, "Expected body statements")
		})
	}
}

func TestParsing_AlternativeForStatement(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedErrors int
	}{
		{
			name: "Alternative For loop",
			input: `<?php
for ($i = 0; $i < 5; $i++):
    echo $i;
endfor;
?>`,
			expectedErrors: 0,
		},
		{
			name: "Alternative For with empty init",
			input: `<?php
for (; $i < 10; $i++):
    echo $i;
endfor;
?>`,
			expectedErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()

			assert.Equal(t, tt.expectedErrors, len(p.Errors()), "Parser errors: %v", p.Errors())
			assert.NotNil(t, program)
			assert.Greater(t, len(program.Body), 0, "Expected at least one statement")

			// Check that we have an AlternativeForStatement
			stmt := program.Body[0]
			altForStmt, ok := stmt.(*ast.AlternativeForStatement)
			assert.True(t, ok, "Expected AlternativeForStatement, got %T", stmt)
			assert.Greater(t, len(altForStmt.Body), 0, "Expected body statements")
		})
	}
}

func TestParsing_AlternativeForeachStatement(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedErrors int
	}{
		{
			name: "Alternative Foreach with value only",
			input: `<?php
foreach ($array as $value):
    echo $value;
endforeach;
?>`,
			expectedErrors: 0,
		},
		{
			name: "Alternative Foreach with key and value",
			input: `<?php
foreach ($array as $key => $value):
    echo $key . ": " . $value;
endforeach;
?>`,
			expectedErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()

			assert.Equal(t, tt.expectedErrors, len(p.Errors()), "Parser errors: %v", p.Errors())
			assert.NotNil(t, program)
			assert.Greater(t, len(program.Body), 0, "Expected at least one statement")

			// Check that we have an AlternativeForeachStatement
			stmt := program.Body[0]
			altForeachStmt, ok := stmt.(*ast.AlternativeForeachStatement)
			assert.True(t, ok, "Expected AlternativeForeachStatement, got %T", stmt)
			assert.NotNil(t, altForeachStmt.Iterable, "Expected iterable to be set")
			assert.NotNil(t, altForeachStmt.Value, "Expected value to be set")
			assert.Greater(t, len(altForeachStmt.Body), 0, "Expected body statements")
		})
	}
}

func TestParsing_DeclareStatement(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedErrors int
		expectAlt      bool
	}{
		{
			name: "Alternative Declare statement",
			input: `<?php
declare(strict_types=1):
    function test(): int {
        return 42;
    }
enddeclare;
?>`,
			expectedErrors: 0,
			expectAlt:      true,
		},
		{
			name: "Alternative Declare with multiple declarations",
			input: `<?php
declare(ticks=1, encoding='UTF-8'):
    echo "Hello World";
enddeclare;
?>`,
			expectedErrors: 0,
			expectAlt:      true,
		},
		{
			name: "Regular Declare statement",
			input: `<?php
declare(strict_types=1);
?>`,
			expectedErrors: 0,
			expectAlt:      false,
		},
		{
			name: "Regular Declare with block",
			input: `<?php
declare(strict_types=1) {
    function test() {}
}
?>`,
			expectedErrors: 0,
			expectAlt:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()

			assert.Equal(t, tt.expectedErrors, len(p.Errors()), "Parser errors: %v", p.Errors())
			assert.NotNil(t, program)
			assert.Greater(t, len(program.Body), 0, "Expected at least one statement")

			// Check that we have a DeclareStatement
			stmt := program.Body[0]
			declareStmt, ok := stmt.(*ast.DeclareStatement)
			assert.True(t, ok, "Expected DeclareStatement, got %T", stmt)
			assert.Greater(t, len(declareStmt.Declarations), 0, "Expected at least one declaration")
			assert.Equal(t, tt.expectAlt, declareStmt.Alternative, "Expected alternative flag to match")
		})
	}
}

func TestParsing_NamespaceStatements(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedErrors int
	}{
		{
			name: "Simple namespace declaration",
			input: `<?php
namespace App;
?>`,
			expectedErrors: 0,
		},
		{
			name: "Multi-level namespace declaration",
			input: `<?php
namespace App\Http\Controllers;
?>`,
			expectedErrors: 0,
		},
		{
			name: "Global namespace declaration",
			input: `<?php
namespace;
?>`,
			expectedErrors: 0,
		},
		{
			name: "Namespace with block syntax",
			input: `<?php
namespace App {
    function test() {}
}
?>`,
			expectedErrors: 0,
		},
		{
			name: "Anonymous namespace with block syntax",
			input: `<?php
namespace {
    function test() {}
}
?>`,
			expectedErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()

			assert.Equal(t, tt.expectedErrors, len(p.Errors()), "Parser errors: %v", p.Errors())
			assert.NotNil(t, program)
			assert.Greater(t, len(program.Body), 0, "Expected at least one statement")

			// Check that we have a NamespaceStatement
			stmt := program.Body[0]
			namespaceStmt, ok := stmt.(*ast.NamespaceStatement)
			assert.True(t, ok, "Expected NamespaceStatement, got %T", stmt)
			
			// Test specific behavior based on test case
			switch tt.name {
			case "Simple namespace declaration":
				assert.NotNil(t, namespaceStmt.Name, "Expected namespace name to be set")
				assert.Equal(t, []string{"App"}, namespaceStmt.Name.Parts, "Expected namespace parts to match")
				assert.Equal(t, 0, len(namespaceStmt.Body), "Expected empty body for simple declaration")
			case "Multi-level namespace declaration":
				assert.NotNil(t, namespaceStmt.Name, "Expected namespace name to be set")
				assert.Equal(t, []string{"App", "Http", "Controllers"}, namespaceStmt.Name.Parts, "Expected namespace parts to match")
				assert.Equal(t, 0, len(namespaceStmt.Body), "Expected empty body for simple declaration")
			case "Global namespace declaration":
				assert.Nil(t, namespaceStmt.Name, "Expected namespace name to be nil for global namespace")
				assert.Equal(t, 0, len(namespaceStmt.Body), "Expected empty body for global namespace")
			case "Namespace with block syntax":
				assert.NotNil(t, namespaceStmt.Name, "Expected namespace name to be set")
				assert.Equal(t, []string{"App"}, namespaceStmt.Name.Parts, "Expected namespace parts to match")
				assert.Greater(t, len(namespaceStmt.Body), 0, "Expected body statements for block syntax")
			case "Anonymous namespace with block syntax":
				assert.Nil(t, namespaceStmt.Name, "Expected namespace name to be nil for anonymous namespace")
				assert.Greater(t, len(namespaceStmt.Body), 0, "Expected body statements for block syntax")
			}
		})
	}
}

func TestParsing_UseStatements(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedErrors int
	}{
		{
			name: "Simple use statement",
			input: `<?php
use App\Http\Controller;
?>`,
			expectedErrors: 0,
		},
		{
			name: "Use statement with alias",
			input: `<?php
use App\Http\Controller as BaseController;
?>`,
			expectedErrors: 0,
		},
		{
			name: "Multiple use statements",
			input: `<?php
use App\Http\Request, App\Http\Response;
?>`,
			expectedErrors: 0,
		},
		{
			name: "Mixed use statements with aliases",
			input: `<?php
use App\Http\Request as Req, App\Http\Response, App\Http\Controller as BaseController;
?>`,
			expectedErrors: 0,
		},
		{
			name: "Use function statement",
			input: `<?php
use function App\Http\helper_function;
?>`,
			expectedErrors: 0,
		},
		{
			name: "Use const statement",
			input: `<?php
use const App\Http\SOME_CONSTANT;
?>`,
			expectedErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()

			assert.Equal(t, tt.expectedErrors, len(p.Errors()), "Parser errors: %v", p.Errors())
			assert.NotNil(t, program)
			assert.Greater(t, len(program.Body), 0, "Expected at least one statement")

			// Check that we have a UseStatement
			stmt := program.Body[0]
			useStmt, ok := stmt.(*ast.UseStatement)
			assert.True(t, ok, "Expected UseStatement, got %T", stmt)
			assert.Greater(t, len(useStmt.Uses), 0, "Expected at least one use clause")

			// Test specific behavior based on test case
			switch tt.name {
			case "Simple use statement":
				assert.Equal(t, 1, len(useStmt.Uses), "Expected one use clause")
				assert.Equal(t, []string{"App", "Http", "Controller"}, useStmt.Uses[0].Name.Parts, "Expected namespace parts to match")
				assert.Equal(t, "", useStmt.Uses[0].Alias, "Expected no alias")
				assert.Equal(t, "", useStmt.Uses[0].Type, "Expected no type")
			case "Use statement with alias":
				assert.Equal(t, 1, len(useStmt.Uses), "Expected one use clause")
				assert.Equal(t, []string{"App", "Http", "Controller"}, useStmt.Uses[0].Name.Parts, "Expected namespace parts to match")
				assert.Equal(t, "BaseController", useStmt.Uses[0].Alias, "Expected alias to match")
				assert.Equal(t, "", useStmt.Uses[0].Type, "Expected no type")
			case "Multiple use statements":
				assert.Equal(t, 2, len(useStmt.Uses), "Expected two use clauses")
				assert.Equal(t, []string{"App", "Http", "Request"}, useStmt.Uses[0].Name.Parts, "Expected first namespace to match")
				assert.Equal(t, []string{"App", "Http", "Response"}, useStmt.Uses[1].Name.Parts, "Expected second namespace to match")
				assert.Equal(t, "", useStmt.Uses[0].Alias, "Expected no alias on first")
				assert.Equal(t, "", useStmt.Uses[1].Alias, "Expected no alias on second")
			case "Mixed use statements with aliases":
				assert.Equal(t, 3, len(useStmt.Uses), "Expected three use clauses")
				assert.Equal(t, "Req", useStmt.Uses[0].Alias, "Expected alias on first")
				assert.Equal(t, "", useStmt.Uses[1].Alias, "Expected no alias on second")
				assert.Equal(t, "BaseController", useStmt.Uses[2].Alias, "Expected alias on third")
			case "Use function statement":
				assert.Equal(t, 1, len(useStmt.Uses), "Expected one use clause")
				assert.Equal(t, "function", useStmt.Uses[0].Type, "Expected function type")
				assert.Equal(t, []string{"App", "Http", "helper_function"}, useStmt.Uses[0].Name.Parts, "Expected namespace parts to match")
			case "Use const statement":
				assert.Equal(t, 1, len(useStmt.Uses), "Expected one use clause")
				assert.Equal(t, "const", useStmt.Uses[0].Type, "Expected const type")
				assert.Equal(t, []string{"App", "Http", "SOME_CONSTANT"}, useStmt.Uses[0].Name.Parts, "Expected namespace parts to match")
			}
		})
	}
}

func TestParsing_InterfaceDeclarations(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedErrors int
	}{
		{
			name: "Simple interface declaration",
			input: `<?php
interface UserInterface {
    public function getName(): string;
}
?>`,
			expectedErrors: 0,
		},
		{
			name: "Interface with multiple methods",
			input: `<?php
interface DatabaseInterface {
    public function connect(): bool;
    public function query(string $sql, array $params): array;
    public function disconnect(): void;
}
?>`,
			expectedErrors: 0,
		},
		{
			name: "Interface extending another interface",
			input: `<?php
interface AdminInterface extends UserInterface {
    public function getPermissions(): array;
}
?>`,
			expectedErrors: 0,
		},
		{
			name: "Interface extending multiple interfaces",
			input: `<?php
interface SuperAdminInterface extends UserInterface, AdminInterface {
    public function deleteUser(int $id): bool;
}
?>`,
			expectedErrors: 0,
		},
		{
			name: "Interface with complex method signatures",
			input: `<?php
interface ServiceInterface {
    public function process(string $data, ?array $options = null): ?object;
    public function validate(array &$data): bool;
}
?>`,
			expectedErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()

			assert.Equal(t, tt.expectedErrors, len(p.Errors()), "Parser errors: %v", p.Errors())
			assert.NotNil(t, program)
			assert.Greater(t, len(program.Body), 0, "Expected at least one statement")

			// Check that we have an InterfaceDeclaration
			stmt := program.Body[0]
			interfaceDecl, ok := stmt.(*ast.InterfaceDeclaration)
			assert.True(t, ok, "Expected InterfaceDeclaration, got %T", stmt)
			assert.NotNil(t, interfaceDecl.Name, "Expected interface name to be set")

			// Test specific behavior based on test case
			switch tt.name {
			case "Simple interface declaration":
				assert.Equal(t, "UserInterface", interfaceDecl.Name.Name, "Expected interface name to match")
				assert.Equal(t, 0, len(interfaceDecl.Extends), "Expected no parent interfaces")
				assert.Equal(t, 1, len(interfaceDecl.Methods), "Expected one method")
				assert.Equal(t, "getName", interfaceDecl.Methods[0].Name.Name, "Expected method name to match")
				assert.Equal(t, "public", interfaceDecl.Methods[0].Visibility, "Expected public visibility")
				assert.NotNil(t, interfaceDecl.Methods[0].ReturnType, "Expected return type")
			case "Interface with multiple methods":
				assert.Equal(t, "DatabaseInterface", interfaceDecl.Name.Name, "Expected interface name to match")
				assert.Equal(t, 3, len(interfaceDecl.Methods), "Expected three methods")
				assert.Equal(t, "connect", interfaceDecl.Methods[0].Name.Name, "Expected first method name")
				assert.Equal(t, "query", interfaceDecl.Methods[1].Name.Name, "Expected second method name")
				assert.Equal(t, "disconnect", interfaceDecl.Methods[2].Name.Name, "Expected third method name")
			case "Interface extending another interface":
				assert.Equal(t, "AdminInterface", interfaceDecl.Name.Name, "Expected interface name to match")
				assert.Equal(t, 1, len(interfaceDecl.Extends), "Expected one parent interface")
				assert.Equal(t, "UserInterface", interfaceDecl.Extends[0].Name, "Expected parent interface name")
				assert.Equal(t, 1, len(interfaceDecl.Methods), "Expected one method")
			case "Interface extending multiple interfaces":
				assert.Equal(t, "SuperAdminInterface", interfaceDecl.Name.Name, "Expected interface name to match")
				assert.Equal(t, 2, len(interfaceDecl.Extends), "Expected two parent interfaces")
				assert.Equal(t, "UserInterface", interfaceDecl.Extends[0].Name, "Expected first parent interface")
				assert.Equal(t, "AdminInterface", interfaceDecl.Extends[1].Name, "Expected second parent interface")
			case "Interface with complex method signatures":
				assert.Equal(t, "ServiceInterface", interfaceDecl.Name.Name, "Expected interface name to match")
				assert.Equal(t, 2, len(interfaceDecl.Methods), "Expected two methods")
				assert.Equal(t, "process", interfaceDecl.Methods[0].Name.Name, "Expected first method name")
				assert.Equal(t, "validate", interfaceDecl.Methods[1].Name.Name, "Expected second method name")
				// Test that the second method has reference parameter
				assert.True(t, interfaceDecl.Methods[1].Parameters[0].ByReference, "Expected reference parameter")
			}
		})
	}
}

func TestParsing_TraitDeclarations(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedErrors int
	}{
		{
			name: "Simple trait declaration",
			input: `<?php
trait LoggerTrait {
    public function log(string $message): void {
        echo $message;
    }
}
?>`,
			expectedErrors: 0,
		},
		{
			name: "Trait with properties and methods",
			input: `<?php
trait DatabaseTrait {
    private $connection;
    protected $config;
    
    public function connect(): bool {
        return true;
    }
    
    private function getConfig(): array {
        return $this->config;
    }
}
?>`,
			expectedErrors: 0,
		},
		{
			name: "Trait with mixed visibility modifiers",
			input: `<?php
trait UtilityTrait {
    public $publicProperty;
    private $privateProperty;
    protected $protectedProperty;
    
    public function publicMethod(): void {}
    private function privateMethod(): string { return "private"; }
    protected function protectedMethod(array $data): array { return $data; }
}
?>`,
			expectedErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()

			assert.Equal(t, tt.expectedErrors, len(p.Errors()), "Parser errors: %v", p.Errors())
			assert.NotNil(t, program)
			assert.Greater(t, len(program.Body), 0, "Expected at least one statement")

			// Check that we have a TraitDeclaration
			stmt := program.Body[0]
			traitDecl, ok := stmt.(*ast.TraitDeclaration)
			assert.True(t, ok, "Expected TraitDeclaration, got %T", stmt)
			assert.NotNil(t, traitDecl.Name, "Expected trait name to be set")

			// Test specific behavior based on test case
			switch tt.name {
			case "Simple trait declaration":
				assert.Equal(t, "LoggerTrait", traitDecl.Name.Name, "Expected trait name to match")
				assert.Equal(t, 0, len(traitDecl.Properties), "Expected no properties")
				assert.Equal(t, 1, len(traitDecl.Methods), "Expected one method")
				assert.Equal(t, "log", traitDecl.Methods[0].Name.String(), "Expected method name to match")
				assert.Equal(t, "public", traitDecl.Methods[0].Visibility, "Expected public visibility")
			case "Trait with properties and methods":
				assert.Equal(t, "DatabaseTrait", traitDecl.Name.Name, "Expected trait name to match")
				assert.Equal(t, 2, len(traitDecl.Properties), "Expected two properties")
				assert.Equal(t, 2, len(traitDecl.Methods), "Expected two methods")
				assert.Equal(t, "connection", traitDecl.Properties[0].Name, "Expected first property name")
				assert.Equal(t, "config", traitDecl.Properties[1].Name, "Expected second property name")
				assert.Equal(t, "private", traitDecl.Properties[0].Visibility, "Expected private visibility")
				assert.Equal(t, "protected", traitDecl.Properties[1].Visibility, "Expected protected visibility")
			case "Trait with mixed visibility modifiers":
				assert.Equal(t, "UtilityTrait", traitDecl.Name.Name, "Expected trait name to match")
				assert.Equal(t, 3, len(traitDecl.Properties), "Expected three properties")
				assert.Equal(t, 3, len(traitDecl.Methods), "Expected three methods")
				// Test property visibilities
				assert.Equal(t, "public", traitDecl.Properties[0].Visibility, "Expected public visibility")
				assert.Equal(t, "private", traitDecl.Properties[1].Visibility, "Expected private visibility")
				assert.Equal(t, "protected", traitDecl.Properties[2].Visibility, "Expected protected visibility")
				// Test method visibilities
				assert.Equal(t, "public", traitDecl.Methods[0].Visibility, "Expected public method visibility")
				assert.Equal(t, "private", traitDecl.Methods[1].Visibility, "Expected private method visibility")
				assert.Equal(t, "protected", traitDecl.Methods[2].Visibility, "Expected protected method visibility")
			}
		})
	}
}

func TestParsing_ComprehensiveModernPHP(t *testing.T) {
	// This test demonstrates all the newly implemented features working together
	input := `<?php
namespace App\Http\Controllers;

use App\Models\User;
use App\Services\AuthService as Auth;
use function App\Helpers\sanitize_input;
use const App\Config\DEFAULT_TIMEOUT;

interface UserRepositoryInterface {
    public function findById(int $id): ?User;
    public function create(array $data): User;
    public function update(int $id, array $data): bool;
}

trait LoggerTrait {
    protected string $logFile = 'app.log';
    
    public function log(string $message): void {
        // Implementation here
    }
}
?>`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	assert.Equal(t, 0, len(p.Errors()), "Parser errors: %v", p.Errors())
	assert.NotNil(t, program)
	assert.Greater(t, len(program.Body), 5, "Expected multiple top-level statements")

	// Test namespace declaration
	namespaceStmt, ok := program.Body[0].(*ast.NamespaceStatement)
	assert.True(t, ok, "Expected first statement to be namespace")
	assert.Equal(t, []string{"App", "Http", "Controllers"}, namespaceStmt.Name.Parts)

	// Find and test use statements (they might be separate statements)
	var useStatements []*ast.UseStatement
	var interfaceDecl *ast.InterfaceDeclaration
	var traitDecl *ast.TraitDeclaration
	
	for _, stmt := range program.Body {
		switch s := stmt.(type) {
		case *ast.UseStatement:
			useStatements = append(useStatements, s)
		case *ast.InterfaceDeclaration:
			interfaceDecl = s
		case *ast.TraitDeclaration:
			traitDecl = s
		}
	}

	// Test use statements
	assert.Greater(t, len(useStatements), 0, "Expected at least one use statement")
	if len(useStatements) > 0 {
		// Test first use statement
		assert.Equal(t, []string{"App", "Models", "User"}, useStatements[0].Uses[0].Name.Parts)
	}
	if len(useStatements) > 1 {
		// Test aliased use statement
		assert.Equal(t, "Auth", useStatements[1].Uses[0].Alias, "Expected alias")
	}

	// Test interface declaration
	assert.NotNil(t, interfaceDecl, "Expected interface declaration")
	if interfaceDecl != nil {
		assert.Equal(t, "UserRepositoryInterface", interfaceDecl.Name.Name)
		assert.Equal(t, 3, len(interfaceDecl.Methods), "Expected 3 interface methods")
	}

	// Test trait declaration
	assert.NotNil(t, traitDecl, "Expected trait declaration")
	if traitDecl != nil {
		assert.Equal(t, "LoggerTrait", traitDecl.Name.Name)
		assert.Equal(t, 1, len(traitDecl.Properties), "Expected 1 trait property")
		assert.Equal(t, 1, len(traitDecl.Methods), "Expected 1 trait method")
	}
}

func TestParsing_AlternativeSyntaxWithModernFeatures(t *testing.T) {
	// This test combines alternative syntax with modern PHP features
	input := `<?php
namespace App\Utils;

use App\Services\Logger;

trait CacheTrait {
    private array $cache = [];
}

interface ProcessorInterface {
    public function process(array $data): bool;
}

if ($condition):
    echo "Processing...";
endif;

foreach ($items as $item):
    echo $item;
endforeach;

while ($running):
    echo "Running...";
endwhile;
?>`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	assert.Equal(t, 0, len(p.Errors()), "Parser errors: %v", p.Errors())
	assert.NotNil(t, program)
	assert.Equal(t, 7, len(program.Body), "Expected 7 statements combining modern features with alternative syntax")

	// Verify we have all the different types of statements
	statementTypes := make(map[string]bool)
	for _, stmt := range program.Body {
		switch stmt.(type) {
		case *ast.NamespaceStatement:
			statementTypes["namespace"] = true
		case *ast.UseStatement:
			statementTypes["use"] = true
		case *ast.TraitDeclaration:
			statementTypes["trait"] = true
		case *ast.InterfaceDeclaration:
			statementTypes["interface"] = true
		case *ast.AlternativeIfStatement:
			statementTypes["alt_if"] = true
		case *ast.AlternativeForeachStatement:
			statementTypes["alt_foreach"] = true
		case *ast.AlternativeWhileStatement:
			statementTypes["alt_while"] = true
		}
	}

	// Verify we have all expected statement types
	expectedTypes := []string{"namespace", "use", "trait", "interface", "alt_if", "alt_foreach", "alt_while"}
	for _, expectedType := range expectedTypes {
		assert.True(t, statementTypes[expectedType], "Expected %s statement type", expectedType)
	}
}

func TestParsing_EnumDeclarations(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedErrors int
	}{
		{
			name: "Simple enum declaration",
			input: `<?php
enum Status {
    case PENDING;
    case APPROVED;
    case REJECTED;
}
?>`,
			expectedErrors: 0,
		},
		{
			name: "Backed enum with string values",
			input: `<?php
enum Status: string {
    case PENDING = 'pending';
    case APPROVED = 'approved';
    case REJECTED = 'rejected';
}
?>`,
			expectedErrors: 0,
		},
		{
			name: "Backed enum with integer values",
			input: `<?php
enum Priority: int {
    case LOW = 1;
    case MEDIUM = 2;
    case HIGH = 3;
}
?>`,
			expectedErrors: 0,
		},
		{
			name: "Enum implementing interfaces",
			input: `<?php
enum Color implements ColorInterface {
    case RED;
    case GREEN;
    case BLUE;
    
    public function getHex(): string {
        return '#FF0000';
    }
}
?>`,
			expectedErrors: 0,
		},
		{
			name: "Enum with multiple interfaces and methods",
			input: `<?php
enum HttpStatus: int implements StatusInterface, JsonSerializable {
    case OK = 200;
    case NOT_FOUND = 404;
    case SERVER_ERROR = 500;
    
    public function isError(): bool {
        return $this->value >= 400;
    }
    
    public function jsonSerialize(): mixed {
        return $this->value;
    }
}
?>`,
			expectedErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()

			assert.Equal(t, tt.expectedErrors, len(p.Errors()), "Parser errors: %v", p.Errors())
			assert.NotNil(t, program)
			assert.Greater(t, len(program.Body), 0, "Expected at least one statement")

			// Check that we have an EnumDeclaration
			stmt := program.Body[0]
			enumDecl, ok := stmt.(*ast.EnumDeclaration)
			assert.True(t, ok, "Expected EnumDeclaration, got %T", stmt)
			assert.NotNil(t, enumDecl.Name, "Expected enum name to be set")

			// Test specific behavior based on test case
			switch tt.name {
			case "Simple enum declaration":
				assert.Equal(t, "Status", enumDecl.Name.Name, "Expected enum name to match")
				assert.Nil(t, enumDecl.BackingType, "Expected no backing type")
				assert.Equal(t, 0, len(enumDecl.Implements), "Expected no interfaces")
				assert.Equal(t, 3, len(enumDecl.Cases), "Expected 3 enum cases")
				assert.Equal(t, 0, len(enumDecl.Methods), "Expected no methods")
				assert.Equal(t, "PENDING", enumDecl.Cases[0].Name.Name, "Expected first case name")
				assert.Nil(t, enumDecl.Cases[0].Value, "Expected no value for pure enum case")
			case "Backed enum with string values":
				assert.Equal(t, "Status", enumDecl.Name.Name, "Expected enum name to match")
				assert.NotNil(t, enumDecl.BackingType, "Expected backing type")
				assert.Equal(t, 3, len(enumDecl.Cases), "Expected 3 enum cases")
				assert.NotNil(t, enumDecl.Cases[0].Value, "Expected value for backed enum case")
			case "Backed enum with integer values":
				assert.Equal(t, "Priority", enumDecl.Name.Name, "Expected enum name to match")
				assert.NotNil(t, enumDecl.BackingType, "Expected backing type")
				assert.Equal(t, 3, len(enumDecl.Cases), "Expected 3 enum cases")
			case "Enum implementing interfaces":
				assert.Equal(t, "Color", enumDecl.Name.Name, "Expected enum name to match")
				assert.Equal(t, 1, len(enumDecl.Implements), "Expected one interface")
				assert.Equal(t, "ColorInterface", enumDecl.Implements[0].Name, "Expected interface name")
				assert.Equal(t, 3, len(enumDecl.Cases), "Expected 3 enum cases")
				assert.Equal(t, 1, len(enumDecl.Methods), "Expected 1 method")
			case "Enum with multiple interfaces and methods":
				assert.Equal(t, "HttpStatus", enumDecl.Name.Name, "Expected enum name to match")
				assert.NotNil(t, enumDecl.BackingType, "Expected backing type")
				assert.Equal(t, 2, len(enumDecl.Implements), "Expected two interfaces")
				assert.Equal(t, "StatusInterface", enumDecl.Implements[0].Name, "Expected first interface")
				assert.Equal(t, "JsonSerializable", enumDecl.Implements[1].Name, "Expected second interface")
				assert.Equal(t, 3, len(enumDecl.Cases), "Expected 3 enum cases")
				assert.Equal(t, 2, len(enumDecl.Methods), "Expected 2 methods")
			}
		})
	}
}

