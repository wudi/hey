package parser

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wudi/php-parser/ast"
	"github.com/wudi/php-parser/lexer"
)

func TestParsing_ReservedKeywords(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected func(t *testing.T, program *ast.Program)
	}{
		{
			name: "Class constants with reserved keywords",
			input: `<?php
class TestClass {
    const class = 'class_value';
    const function = 'function_value';
    const if = 'if_value';
    public const new = 'new_value';
    private const while = 'while_value';
}`,
			expected: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)
				
				classStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)
				
				classExpr, ok := classStmt.Expression.(*ast.ClassExpression)
				require.True(t, ok)
				
				require.Len(t, classExpr.Body, 5)
				
				// Test first constant: const class = 'class_value';
				constDecl1, ok := classExpr.Body[0].(*ast.ClassConstantDeclaration)
				require.True(t, ok)
				assert.Equal(t, "public", constDecl1.Visibility)
				assert.Len(t, constDecl1.Constants, 1)
				nameNode1, ok := constDecl1.Constants[0].Name.(*ast.IdentifierNode)
				require.True(t, ok)
				assert.Equal(t, "class", nameNode1.Name)
				
				// Test public const new = 'new_value';
				constDecl4, ok := classExpr.Body[3].(*ast.ClassConstantDeclaration)
				require.True(t, ok)
				assert.Equal(t, "public", constDecl4.Visibility)
				nameNode4, ok := constDecl4.Constants[0].Name.(*ast.IdentifierNode)
				require.True(t, ok)
				assert.Equal(t, "new", nameNode4.Name)
				
				// Test private const while = 'while_value';
				constDecl5, ok := classExpr.Body[4].(*ast.ClassConstantDeclaration)
				require.True(t, ok)
				assert.Equal(t, "private", constDecl5.Visibility)
				nameNode5, ok := constDecl5.Constants[0].Name.(*ast.IdentifierNode)
				require.True(t, ok)
				assert.Equal(t, "while", nameNode5.Name)
			},
		},
		{
			name: "Method names with reserved keywords",
			input: `<?php
class TestClass {
    public function class() {
        return 'class';
    }
    
    private function if() {
        return 'if';
    }
    
    protected function while() {
        return 'while';
    }
    
    public function function() {
        return 'function';
    }
}`,
			expected: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)
				
				classStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)
				
				classExpr, ok := classStmt.Expression.(*ast.ClassExpression)
				require.True(t, ok)
				
				require.Len(t, classExpr.Body, 4)
				
				// Test public function class()
				method1, ok := classExpr.Body[0].(*ast.FunctionDeclaration)
				require.True(t, ok)
				nameNode1, ok := method1.Name.(*ast.IdentifierNode)
				require.True(t, ok)
				assert.Equal(t, "class", nameNode1.Name)
				assert.Equal(t, "public", method1.Visibility)
				
				// Test private function if()
				method2, ok := classExpr.Body[1].(*ast.FunctionDeclaration)
				require.True(t, ok)
				nameNode2, ok := method2.Name.(*ast.IdentifierNode)
				require.True(t, ok)
				assert.Equal(t, "if", nameNode2.Name)
				assert.Equal(t, "private", method2.Visibility)
				
				// Test protected function while()
				method3, ok := classExpr.Body[2].(*ast.FunctionDeclaration)
				require.True(t, ok)
				nameNode3, ok := method3.Name.(*ast.IdentifierNode)
				require.True(t, ok)
				assert.Equal(t, "while", nameNode3.Name)
				assert.Equal(t, "protected", method3.Visibility)
				
				// Test public function function()
				method4, ok := classExpr.Body[3].(*ast.FunctionDeclaration)
				require.True(t, ok)
				nameNode4, ok := method4.Name.(*ast.IdentifierNode)
				require.True(t, ok)
				assert.Equal(t, "function", nameNode4.Name)
				assert.Equal(t, "public", method4.Visibility)
			},
		},
		{
			name: "Property access with reserved keywords",
			input: `<?php
$obj->class;
$obj->function;
$obj->if;
$obj->while;
$obj->new;`,
			expected: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 5)
				
				// Test $obj->class
				stmt1, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)
				propAccess1, ok := stmt1.Expression.(*ast.PropertyAccessExpression)
				require.True(t, ok)
				propIdent1, ok := propAccess1.Property.(*ast.IdentifierNode)
				require.True(t, ok, "Expected IdentifierNode for property")
				assert.Equal(t, "class", propIdent1.Name)
				
				// Test $obj->function
				stmt2, ok := program.Body[1].(*ast.ExpressionStatement)
				require.True(t, ok)
				propAccess2, ok := stmt2.Expression.(*ast.PropertyAccessExpression)
				require.True(t, ok)
				propIdent2, ok := propAccess2.Property.(*ast.IdentifierNode)
				require.True(t, ok, "Expected IdentifierNode for property")
				assert.Equal(t, "function", propIdent2.Name)
				
				// Test $obj->if
				stmt3, ok := program.Body[2].(*ast.ExpressionStatement)
				require.True(t, ok)
				propAccess3, ok := stmt3.Expression.(*ast.PropertyAccessExpression)
				require.True(t, ok)
				propIdent3, ok := propAccess3.Property.(*ast.IdentifierNode)
				require.True(t, ok, "Expected IdentifierNode for property")
				assert.Equal(t, "if", propIdent3.Name)
				
				// Test $obj->while
				stmt4, ok := program.Body[3].(*ast.ExpressionStatement)
				require.True(t, ok)
				propAccess4, ok := stmt4.Expression.(*ast.PropertyAccessExpression)
				require.True(t, ok)
				propIdent4, ok := propAccess4.Property.(*ast.IdentifierNode)
				require.True(t, ok, "Expected IdentifierNode for property")
				assert.Equal(t, "while", propIdent4.Name)
				
				// Test $obj->new
				stmt5, ok := program.Body[4].(*ast.ExpressionStatement)
				require.True(t, ok)
				propAccess5, ok := stmt5.Expression.(*ast.PropertyAccessExpression)
				require.True(t, ok)
				propIdent5, ok := propAccess5.Property.(*ast.IdentifierNode)
				require.True(t, ok, "Expected IdentifierNode for property")
				assert.Equal(t, "new", propIdent5.Name)
			},
		},
		{
			name: "Nullsafe property access with reserved keywords",
			input: `<?php
$obj?->class;
$obj?->function;
$obj?->if;`,
			expected: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 3)
				
				// Test $obj?->class
				stmt1, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)
				propAccess1, ok := stmt1.Expression.(*ast.NullsafePropertyAccessExpression)
				require.True(t, ok)
				propIdent1, ok := propAccess1.Property.(*ast.IdentifierNode)
				require.True(t, ok, "Expected IdentifierNode for property")
				assert.Equal(t, "class", propIdent1.Name)
				
				// Test $obj?->function
				stmt2, ok := program.Body[1].(*ast.ExpressionStatement)
				require.True(t, ok)
				propAccess2, ok := stmt2.Expression.(*ast.NullsafePropertyAccessExpression)
				require.True(t, ok)
				propIdent2, ok := propAccess2.Property.(*ast.IdentifierNode)
				require.True(t, ok, "Expected IdentifierNode for property")
				assert.Equal(t, "function", propIdent2.Name)
				
				// Test $obj?->if
				stmt3, ok := program.Body[2].(*ast.ExpressionStatement)
				require.True(t, ok)
				propAccess3, ok := stmt3.Expression.(*ast.NullsafePropertyAccessExpression)
				require.True(t, ok)
				propIdent3, ok := propAccess3.Property.(*ast.IdentifierNode)
				require.True(t, ok, "Expected IdentifierNode for property")
				assert.Equal(t, "if", propIdent3.Name)
			},
		},
		{
			name: "Trait adaptations with reserved keywords",
			input: `<?php
class TestClass {
    use TestTrait {
        class as function;
        if as while;
        TestTrait::class as public echo;
        function as private new;
    }
}`,
			expected: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)
				
				classStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)
				
				classExpr, ok := classStmt.Expression.(*ast.ClassExpression)
				require.True(t, ok)
				
				require.Len(t, classExpr.Body, 1)
				useTraitStmt, ok := classExpr.Body[0].(*ast.UseTraitStatement)
				require.True(t, ok)
				
				require.Len(t, useTraitStmt.Adaptations, 4)
				
				// Test class as function
				alias1, ok := useTraitStmt.Adaptations[0].(*ast.TraitAliasStatement)
				require.True(t, ok)
				assert.Equal(t, "class", alias1.Method.Method.Name)
				assert.Equal(t, "function", alias1.Alias.Name)
				
				// Test if as while
				alias2, ok := useTraitStmt.Adaptations[1].(*ast.TraitAliasStatement)
				require.True(t, ok)
				assert.Equal(t, "if", alias2.Method.Method.Name)
				assert.Equal(t, "while", alias2.Alias.Name)
				
				// Test TestTrait::class as public echo
				alias3, ok := useTraitStmt.Adaptations[2].(*ast.TraitAliasStatement)
				require.True(t, ok)
				assert.Equal(t, "TestTrait", alias3.Method.Trait.Name)
				assert.Equal(t, "class", alias3.Method.Method.Name)
				assert.Equal(t, "public", alias3.Visibility)
				assert.Equal(t, "echo", alias3.Alias.Name)
				
				// Test function as private new
				alias4, ok := useTraitStmt.Adaptations[3].(*ast.TraitAliasStatement)
				require.True(t, ok)
				assert.Equal(t, "function", alias4.Method.Method.Name)
				assert.Equal(t, "private", alias4.Visibility)
				assert.Equal(t, "new", alias4.Alias.Name)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			l := lexer.New(test.input)
			parser := New(l)
			program := parser.ParseProgram()
			
			require.NotNil(t, program)
			if len(parser.Errors()) > 0 {
				t.Errorf("Parser errors: %v", parser.Errors())
			}
			
			test.expected(t, program)
		})
	}
}

func TestParsing_ReservedKeywords_IsHelperFunctions(t *testing.T) {
	tests := []struct {
		name       string
		tokenType  lexer.TokenType
		isReserved bool
		isSemi     bool
	}{
		{"T_CLASS is reserved non-modifier", lexer.T_CLASS, true, true},
		{"T_FUNCTION is reserved non-modifier", lexer.T_FUNCTION, true, true},
		{"T_IF is reserved non-modifier", lexer.T_IF, true, true},
		{"T_WHILE is reserved non-modifier", lexer.T_WHILE, true, true},
		{"T_NEW is reserved non-modifier", lexer.T_NEW, true, true},
		{"T_PRIVATE is not reserved non-modifier but is semi-reserved", lexer.T_PRIVATE, false, true},
		{"T_PUBLIC is not reserved non-modifier but is semi-reserved", lexer.T_PUBLIC, false, true},
		{"T_PROTECTED is not reserved non-modifier but is semi-reserved", lexer.T_PROTECTED, false, true},
		{"T_STATIC is not reserved non-modifier but is semi-reserved", lexer.T_STATIC, false, true},
		{"T_STRING is not reserved", lexer.T_STRING, false, false},
		{"T_VARIABLE is not reserved", lexer.T_VARIABLE, false, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.isReserved, isReservedNonModifier(test.tokenType))
			assert.Equal(t, test.isSemi, isSemiReserved(test.tokenType))
		})
	}
}