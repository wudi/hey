package parser

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wudi/php-parser/ast"
	"github.com/wudi/php-parser/lexer"
)

func TestParsing_FirstClassCallable_Complete(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected func(t *testing.T, program *ast.Program)
	}{
		{
			name: "Simple function first-class callable",
			input: `<?php
$func = strlen(...);`,
			expected: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)
				
				stmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)
				
				assignment, ok := stmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)
				
				fcc, ok := assignment.Right.(*ast.FirstClassCallable)
				require.True(t, ok)
				
				identNode, ok := fcc.Callable.(*ast.IdentifierNode)
				require.True(t, ok)
				assert.Equal(t, "strlen", identNode.Name)
			},
		},
		{
			name: "Object method first-class callable",
			input: `<?php
$obj = new stdClass();
$method = $obj->method(...);`,
			expected: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 2)
				
				stmt2, ok := program.Body[1].(*ast.ExpressionStatement)
				require.True(t, ok)
				
				assignment, ok := stmt2.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)
				
				fcc, ok := assignment.Right.(*ast.FirstClassCallable)
				require.True(t, ok)
				
				propAccess, ok := fcc.Callable.(*ast.PropertyAccessExpression)
				require.True(t, ok)
				propIdent, ok := propAccess.Property.(*ast.IdentifierNode)
				require.True(t, ok, "Property should be IdentifierNode")
				assert.Equal(t, "method", propIdent.Name)
				
				objVar, ok := propAccess.Object.(*ast.Variable)
				require.True(t, ok)
				assert.Equal(t, "$obj", objVar.Name)
			},
		},
		{
			name: "Static method first-class callable",
			input: `<?php
$static = MyClass::staticMethod(...);`,
			expected: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)
				
				stmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)
				
				assignment, ok := stmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)
				
				fcc, ok := assignment.Right.(*ast.FirstClassCallable)
				require.True(t, ok)
				
				staticAccess, ok := fcc.Callable.(*ast.StaticAccessExpression)
				require.True(t, ok)
				
				propertyNode, ok := staticAccess.Property.(*ast.IdentifierNode)
				require.True(t, ok)
				assert.Equal(t, "staticMethod", propertyNode.Name)
				
				className, ok := staticAccess.Class.(*ast.IdentifierNode)
				require.True(t, ok)
				assert.Equal(t, "MyClass", className.Name)
			},
		},
		{
			name: "Variable function first-class callable",
			input: `<?php
$funcName = 'strlen';
$variableFunc = $funcName(...);`,
			expected: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 2)
				
				stmt2, ok := program.Body[1].(*ast.ExpressionStatement)
				require.True(t, ok)
				
				assignment, ok := stmt2.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)
				
				fcc, ok := assignment.Right.(*ast.FirstClassCallable)
				require.True(t, ok)
				
				varNode, ok := fcc.Callable.(*ast.Variable)
				require.True(t, ok)
				assert.Equal(t, "$funcName", varNode.Name)
			},
		},
		{
			name: "Self static method first-class callable",
			input: `<?php
class TestClass {
    public function test() {
        $self = self::method(...);
        $parent = parent::method(...);
        $static = static::method(...);
    }
}`,
			expected: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)
				
				stmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)
				
				classExpr, ok := stmt.Expression.(*ast.ClassExpression)
				require.True(t, ok)
				
				require.Len(t, classExpr.Body, 1)
				method, ok := classExpr.Body[0].(*ast.FunctionDeclaration)
				require.True(t, ok)
				
				require.Len(t, method.Body, 3)
				
				// Test self::method(...)
				selfStmt, ok := method.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)
				
				selfAssign, ok := selfStmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)
				
				selfFCC, ok := selfAssign.Right.(*ast.FirstClassCallable)
				require.True(t, ok)
				
				selfStatic, ok := selfFCC.Callable.(*ast.StaticAccessExpression)
				require.True(t, ok)
				
				selfClass, ok := selfStatic.Class.(*ast.IdentifierNode)
				require.True(t, ok)
				assert.Equal(t, "self", selfClass.Name)
				
				selfPropertyNode, ok := selfStatic.Property.(*ast.IdentifierNode)
				require.True(t, ok)
				assert.Equal(t, "method", selfPropertyNode.Name)
				
				// Test parent::method(...)
				parentStmt, ok := method.Body[1].(*ast.ExpressionStatement)
				require.True(t, ok)
				
				parentAssign, ok := parentStmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)
				
				parentFCC, ok := parentAssign.Right.(*ast.FirstClassCallable)
				require.True(t, ok)
				
				parentStatic, ok := parentFCC.Callable.(*ast.StaticAccessExpression)
				require.True(t, ok)
				
				parentClass, ok := parentStatic.Class.(*ast.IdentifierNode)
				require.True(t, ok)
				assert.Equal(t, "parent", parentClass.Name)
				
				parentPropertyNode, ok := parentStatic.Property.(*ast.IdentifierNode)
				require.True(t, ok)
				assert.Equal(t, "method", parentPropertyNode.Name)
				
				// Test static::method(...)
				staticStmt, ok := method.Body[2].(*ast.ExpressionStatement)
				require.True(t, ok)
				
				staticAssign, ok := staticStmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)
				
				staticFCC, ok := staticAssign.Right.(*ast.FirstClassCallable)
				require.True(t, ok)
				
				staticStatic, ok := staticFCC.Callable.(*ast.StaticAccessExpression)
				require.True(t, ok)
				
				staticClass, ok := staticStatic.Class.(*ast.IdentifierNode)
				require.True(t, ok)
				assert.Equal(t, "static", staticClass.Name)
				
				staticPropertyNode, ok := staticStatic.Property.(*ast.IdentifierNode)
				require.True(t, ok)
				assert.Equal(t, "method", staticPropertyNode.Name)
			},
		},
		{
			name: "Closure first-class callable",
			input: `<?php
$closure = function() { return 'test'; };
$closureFCC = $closure(...);`,
			expected: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 2)
				
				stmt2, ok := program.Body[1].(*ast.ExpressionStatement)
				require.True(t, ok)
				
				assignment, ok := stmt2.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)
				
				fcc, ok := assignment.Right.(*ast.FirstClassCallable)
				require.True(t, ok)
				
				varNode, ok := fcc.Callable.(*ast.Variable)
				require.True(t, ok)
				assert.Equal(t, "$closure", varNode.Name)
			},
		},
		{
			name: "Complex array method first-class callable",
			input: `<?php
$arrayMethod = [new stdClass(), 'toString'](...);`,
			expected: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)
				
				stmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)
				
				assignment, ok := stmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)
				
				fcc, ok := assignment.Right.(*ast.FirstClassCallable)
				require.True(t, ok)
				
				arrayLit, ok := fcc.Callable.(*ast.ArrayExpression)
				require.True(t, ok)
				
				require.Len(t, arrayLit.Elements, 2)
				
				// First element should be new stdClass()
				newExpr, ok := arrayLit.Elements[0].(*ast.NewExpression)
				require.True(t, ok)
				
				className, ok := newExpr.Class.(*ast.CallExpression)
				require.True(t, ok)
				
				classIdent, ok := className.Callee.(*ast.IdentifierNode)
				require.True(t, ok)
				assert.Equal(t, "stdClass", classIdent.Name)
				
				// Second element should be 'toString' string
				stringLit, ok := arrayLit.Elements[1].(*ast.StringLiteral)
				require.True(t, ok)
				assert.Equal(t, "toString", stringLit.Value)
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

// Test edge cases and error handling
func TestParsing_FirstClassCallable_EdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		shouldHaveError bool
	}{
		{
			name: "Valid ellipsis syntax",
			input: `<?php
$func = strlen(...);`,
			shouldHaveError: false,
		},
		{
			name: "Invalid ellipsis usage (not in function call)",
			input: `<?php  
$invalid = ...;`,
			shouldHaveError: true,
		},
		{
			name: "Nested function calls with first-class callable",
			input: `<?php
$result = call_user_func(strlen(...), 'test');`,
			shouldHaveError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			l := lexer.New(test.input)
			parser := New(l)
			program := parser.ParseProgram()
			
			require.NotNil(t, program)
			errors := parser.Errors()
			
			if test.shouldHaveError {
				assert.NotEmpty(t, errors, "Expected parser errors but got none")
			} else {
				if len(errors) > 0 {
					t.Errorf("Unexpected parser errors: %v", errors)
				}
			}
		})
	}
}