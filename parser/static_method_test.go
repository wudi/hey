package parser

import (
	"testing"
	
	"github.com/stretchr/testify/assert"
	"github.com/wudi/php-parser/ast"
	"github.com/wudi/php-parser/lexer"
)

func TestParsing_StaticMethods(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		validate func(t *testing.T, program *ast.Program)
	}{
		{
			name: "static public function",
			input: `<?php
class MyClass {
    static public function fromArray($array) {
        return 1;
    }
}`,
			validate: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 1)
				
				stmt, ok := program.Body[0].(*ast.ExpressionStatement)
				assert.True(t, ok)
				
				classExpr, ok := stmt.Expression.(*ast.ClassExpression)
				assert.True(t, ok)
				
				nameIdent, ok := classExpr.Name.(*ast.IdentifierNode)
				assert.True(t, ok)
				assert.Equal(t, "MyClass", nameIdent.Name)
				
				assert.Len(t, classExpr.Body, 1)
				method, ok := classExpr.Body[0].(*ast.FunctionDeclaration)
				assert.True(t, ok)
				
				methodName, ok := method.Name.(*ast.IdentifierNode)
				assert.True(t, ok)
				assert.Equal(t, "fromArray", methodName.Name)
				assert.Equal(t, "public", method.Visibility)
				assert.True(t, method.IsStatic)
				assert.False(t, method.IsAbstract)
				assert.False(t, method.IsFinal)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()

			checkParserErrors(t, p)

			assert.NotNil(t, program)
			tt.validate(t, program)
		})
	}
}