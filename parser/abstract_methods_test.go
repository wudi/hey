package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wudi/php-parser/ast"
	"github.com/wudi/php-parser/lexer"
)

// Helper function to get class name from ClassExpression
func getClassName(t *testing.T, classExpr *ast.ClassExpression) string {
	className, ok := classExpr.Name.(*ast.IdentifierNode)
	assert.True(t, ok, "Expected class name to be IdentifierNode")
	return className.Name
}

// Helper function to get method name from FunctionDeclaration
func getMethodName(t *testing.T, funcDecl *ast.FunctionDeclaration) string {
	methodName, ok := funcDecl.Name.(*ast.IdentifierNode)
	assert.True(t, ok, "Expected method name to be IdentifierNode")
	return methodName.Name
}

// Helper function to get interface name from Expression
func getInterfaceName(t *testing.T, expr ast.Expression) string {
	interfaceName, ok := expr.(*ast.IdentifierNode)
	assert.True(t, ok, "Expected interface name to be IdentifierNode")
	return interfaceName.Name
}

func TestParsing_AbstractMethods(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedErrors int
		validate       func(t *testing.T, classExpr *ast.ClassExpression)
	}{
		{
			name: "Basic abstract class with abstract methods",
			input: `<?php
abstract class SeekableFileContent implements FileContent {
    protected abstract function doRead($offset, $count);
    protected abstract function getDefaultPermissions();
    protected abstract function doWrite($data, $offset, $length);
}
?>`,
			expectedErrors: 0,
			validate: func(t *testing.T, classExpr *ast.ClassExpression) {
				assert.Equal(t, "SeekableFileContent", getClassName(t, classExpr))
				assert.True(t, classExpr.Abstract, "Expected class to be abstract")
				assert.Equal(t, 1, len(classExpr.Implements), "Expected one interface")
				assert.Equal(t, "FileContent", getInterfaceName(t, classExpr.Implements[0]))
				
				assert.Equal(t, 3, len(classExpr.Body), "Expected three abstract methods")
				
				for i, method := range classExpr.Body {
					funcDecl, ok := method.(*ast.FunctionDeclaration)
					assert.True(t, ok, "Expected FunctionDeclaration at index %d", i)
					assert.Equal(t, "protected", funcDecl.Visibility)
					assert.True(t, funcDecl.IsAbstract, "Expected method to be abstract")
					assert.Nil(t, funcDecl.Body, "Abstract methods should have no body")
				}
				
				// Check specific method names
				method1 := classExpr.Body[0].(*ast.FunctionDeclaration)
				assert.Equal(t, "doRead", getMethodName(t, method1))
				assert.Equal(t, 2, len(method1.Parameters.Parameters))
				
				method2 := classExpr.Body[1].(*ast.FunctionDeclaration)
				assert.Equal(t, "getDefaultPermissions", getMethodName(t, method2))
				if method2.Parameters != nil {
					assert.Equal(t, 0, len(method2.Parameters.Parameters))
				} else {
					assert.Equal(t, 0, 0) // No parameters
				}
				
				method3 := classExpr.Body[2].(*ast.FunctionDeclaration)
				assert.Equal(t, "doWrite", getMethodName(t, method3))
				assert.Equal(t, 3, len(method3.Parameters.Parameters))
			},
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

			// Extract the class from ExpressionStatement
			stmt := program.Body[0]
			exprStmt, ok := stmt.(*ast.ExpressionStatement)
			assert.True(t, ok, "Expected ExpressionStatement, got %T", stmt)

			classExpr, ok := exprStmt.Expression.(*ast.ClassExpression)
			assert.True(t, ok, "Expected ClassExpression, got %T", exprStmt.Expression)

			// Run specific validation
			if tt.validate != nil {
				tt.validate(t, classExpr)
			}
		})
	}
}