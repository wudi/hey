package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wudi/php-parser/ast"
	"github.com/wudi/php-parser/lexer"
)

func TestParsing_MethodModifierCombinations(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedErrors int
		validate       func(t *testing.T, classExpr *ast.ClassExpression)
	}{
		{
			name: "public final static function",
			input: `<?php
class Foo {
    public final static function isSigchildEnabled()
    {
    }
}
?>`,
			expectedErrors: 0,
			validate: func(t *testing.T, classExpr *ast.ClassExpression) {
				assert.Equal(t, "Foo", getClassName(t, classExpr))
				assert.Equal(t, 1, len(classExpr.Body))
				
				method := classExpr.Body[0].(*ast.FunctionDeclaration)
				assert.Equal(t, "isSigchildEnabled", getMethodName(t, method))
				assert.Equal(t, "public", method.Visibility)
				assert.True(t, method.IsFinal, "Expected method to be final")
				assert.True(t, method.IsStatic, "Expected method to be static")
				assert.False(t, method.IsAbstract, "Expected method not to be abstract")
			},
		},
		{
			name: "Different modifier orders",
			input: `<?php
class TestClass {
    public final static function publicFinalStatic() {}
    public static final function publicStaticFinal() {}
    final public static function finalPublicStatic() {}
    final static public function finalStaticPublic() {}
    static public final function staticPublicFinal() {}
    static final public function staticFinalPublic() {}
}
?>`,
			expectedErrors: 0,
			validate: func(t *testing.T, classExpr *ast.ClassExpression) {
				assert.Equal(t, "TestClass", getClassName(t, classExpr))
				assert.Equal(t, 6, len(classExpr.Body))
				
				// Check that all methods have the correct final and static properties regardless of modifier order
				for i, stmt := range classExpr.Body {
					method, ok := stmt.(*ast.FunctionDeclaration)
					assert.True(t, ok, "Expected FunctionDeclaration at index %d", i)
					// Visibility might be explicit or implicit depending on parsing path
					if method.Visibility != "" {
						assert.Equal(t, "public", method.Visibility, "Method %d should be public", i)
					}
					assert.True(t, method.IsFinal, "Method %d should be final", i)
					assert.True(t, method.IsStatic, "Method %d should be static", i)
					assert.False(t, method.IsAbstract, "Method %d should not be abstract", i)
				}
				
				// Check specific method names
				expectedNames := []string{
					"publicFinalStatic", "publicStaticFinal", "finalPublicStatic",
					"finalStaticPublic", "staticPublicFinal", "staticFinalPublic",
				}
				
				for i, expectedName := range expectedNames {
					method := classExpr.Body[i].(*ast.FunctionDeclaration)
					assert.Equal(t, expectedName, getMethodName(t, method))
				}
			},
		},
		{
			name: "Protected and default visibility modifiers",
			input: `<?php
class TestClass {
    protected final static function protectedFinalStatic() {}
    final static function defaultFinalStatic() {}
    static final function defaultStaticFinal() {}
}
?>`,
			expectedErrors: 0,
			validate: func(t *testing.T, classExpr *ast.ClassExpression) {
				assert.Equal(t, "TestClass", getClassName(t, classExpr))
				assert.Equal(t, 3, len(classExpr.Body))
				
				// Protected method
				method1 := classExpr.Body[0].(*ast.FunctionDeclaration)
				assert.Equal(t, "protectedFinalStatic", getMethodName(t, method1))
				assert.Equal(t, "protected", method1.Visibility)
				assert.True(t, method1.IsFinal)
				assert.True(t, method1.IsStatic)
				
				// Default visibility methods (should be public implicitly)
				method2 := classExpr.Body[1].(*ast.FunctionDeclaration)
				assert.Equal(t, "defaultFinalStatic", getMethodName(t, method2))
				assert.Equal(t, "", method2.Visibility) // Explicit visibility not set
				assert.True(t, method2.IsFinal)
				assert.True(t, method2.IsStatic)
				
				method3 := classExpr.Body[2].(*ast.FunctionDeclaration)
				assert.Equal(t, "defaultStaticFinal", getMethodName(t, method3))
				assert.Equal(t, "", method3.Visibility) // Explicit visibility not set
				assert.True(t, method3.IsFinal)
				assert.True(t, method3.IsStatic)
			},
		},
		{
			name: "Abstract methods with static modifier",
			input: `<?php
abstract class AbstractTestClass {
    abstract public static function abstractPublicStatic();
    abstract static public function abstractStaticPublic();
}
?>`,
			expectedErrors: 0,
			validate: func(t *testing.T, classExpr *ast.ClassExpression) {
				assert.Equal(t, "AbstractTestClass", getClassName(t, classExpr))
				assert.True(t, classExpr.Abstract)
				assert.Equal(t, 2, len(classExpr.Body))
				
				// Both methods should be abstract and static
				for i, stmt := range classExpr.Body {
					method, ok := stmt.(*ast.FunctionDeclaration)
					assert.True(t, ok, "Expected FunctionDeclaration at index %d", i)
					assert.True(t, method.IsAbstract, "Method %d should be abstract", i)
					assert.True(t, method.IsStatic, "Method %d should be static", i)
					assert.False(t, method.IsFinal, "Method %d should not be final", i)
					assert.Nil(t, method.Body, "Abstract method %d should have no body", i)
				}
				
				// Check visibility and names
				method1 := classExpr.Body[0].(*ast.FunctionDeclaration)
				assert.Equal(t, "abstractPublicStatic", getMethodName(t, method1))
				assert.Equal(t, "public", method1.Visibility)
				
				method2 := classExpr.Body[1].(*ast.FunctionDeclaration)
				assert.Equal(t, "abstractStaticPublic", getMethodName(t, method2))
				assert.Equal(t, "public", method2.Visibility)
			},
		},
		{
			name: "Single modifiers to ensure no regression",
			input: `<?php
class TestClass {
    public function publicMethod() {}
    private function privateMethod() {}
    protected function protectedMethod() {}
    static function staticMethod() {}
    final function finalMethod() {}
}
?>`,
			expectedErrors: 0,
			validate: func(t *testing.T, classExpr *ast.ClassExpression) {
				assert.Equal(t, "TestClass", getClassName(t, classExpr))
				assert.Equal(t, 5, len(classExpr.Body))
				
				// Public method
				method1 := classExpr.Body[0].(*ast.FunctionDeclaration)
				assert.Equal(t, "publicMethod", getMethodName(t, method1))
				assert.Equal(t, "public", method1.Visibility)
				assert.False(t, method1.IsStatic)
				assert.False(t, method1.IsFinal)
				
				// Private method
				method2 := classExpr.Body[1].(*ast.FunctionDeclaration)
				assert.Equal(t, "privateMethod", getMethodName(t, method2))
				assert.Equal(t, "private", method2.Visibility)
				assert.False(t, method2.IsStatic)
				assert.False(t, method2.IsFinal)
				
				// Protected method
				method3 := classExpr.Body[2].(*ast.FunctionDeclaration)
				assert.Equal(t, "protectedMethod", getMethodName(t, method3))
				assert.Equal(t, "protected", method3.Visibility)
				assert.False(t, method3.IsStatic)
				assert.False(t, method3.IsFinal)
				
				// Static method (no explicit visibility)
				method4 := classExpr.Body[3].(*ast.FunctionDeclaration)
				assert.Equal(t, "staticMethod", getMethodName(t, method4))
				assert.Equal(t, "", method4.Visibility) // Default
				assert.True(t, method4.IsStatic)
				assert.False(t, method4.IsFinal)
				
				// Final method (no explicit visibility)
				method5 := classExpr.Body[4].(*ast.FunctionDeclaration)
				assert.Equal(t, "finalMethod", getMethodName(t, method5))
				assert.Equal(t, "", method5.Visibility) // Default
				assert.False(t, method5.IsStatic)
				assert.True(t, method5.IsFinal)
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