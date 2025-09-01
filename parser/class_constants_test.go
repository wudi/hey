package parser

import (
	"testing"
	
	"github.com/stretchr/testify/assert"
	"github.com/wudi/php-parser/ast"
	"github.com/wudi/php-parser/lexer"
)

func TestParsing_ClassConstantsWithModifiers(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		validate func(t *testing.T, program *ast.Program)
	}{
		{
			name: "basic const without visibility",
			input: `<?php
class Test {
    const BASIC = 1;
}`,
			validate: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 1)
				
				stmt, ok := program.Body[0].(*ast.ExpressionStatement)
				assert.True(t, ok)
				
				classExpr, ok := stmt.Expression.(*ast.ClassExpression)
				assert.True(t, ok)
				
				assert.Len(t, classExpr.Body, 1)
				constDecl, ok := classExpr.Body[0].(*ast.ClassConstantDeclaration)
				assert.True(t, ok)
				
				assert.Equal(t, "", constDecl.Visibility) // defaults to public
				assert.False(t, constDecl.IsFinal)
				assert.False(t, constDecl.IsAbstract)
				assert.Len(t, constDecl.Constants, 1)
				assert.Equal(t, "BASIC", constDecl.Constants[0].Name.(*ast.IdentifierNode).Name)
			},
		},
		{
			name: "visibility modifiers",
			input: `<?php
class Test {
    public const PUBLIC_CONST = 1;
    private const PRIVATE_CONST = 2;
    protected const PROTECTED_CONST = 3;
}`,
			validate: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 1)
				
				stmt, ok := program.Body[0].(*ast.ExpressionStatement)
				assert.True(t, ok)
				
				classExpr, ok := stmt.Expression.(*ast.ClassExpression)
				assert.True(t, ok)
				
				assert.Len(t, classExpr.Body, 3)
				
				// Public const
				publicConst, ok := classExpr.Body[0].(*ast.ClassConstantDeclaration)
				assert.True(t, ok)
				assert.Equal(t, "public", publicConst.Visibility)
				assert.Equal(t, "PUBLIC_CONST", publicConst.Constants[0].Name.(*ast.IdentifierNode).Name)
				
				// Private const
				privateConst, ok := classExpr.Body[1].(*ast.ClassConstantDeclaration)
				assert.True(t, ok)
				assert.Equal(t, "private", privateConst.Visibility)
				assert.Equal(t, "PRIVATE_CONST", privateConst.Constants[0].Name.(*ast.IdentifierNode).Name)
				
				// Protected const
				protectedConst, ok := classExpr.Body[2].(*ast.ClassConstantDeclaration)
				assert.True(t, ok)
				assert.Equal(t, "protected", protectedConst.Visibility)
				assert.Equal(t, "PROTECTED_CONST", protectedConst.Constants[0].Name.(*ast.IdentifierNode).Name)
			},
		},
		{
			name: "final const combinations",
			input: `<?php
class Test {
    final const FINAL_BASIC = 1;
    final public const FINAL_PUBLIC = 2;
    final protected const FINAL_PROTECTED = 3;
}`,
			validate: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 1)
				
				stmt, ok := program.Body[0].(*ast.ExpressionStatement)
				assert.True(t, ok)
				
				classExpr, ok := stmt.Expression.(*ast.ClassExpression)
				assert.True(t, ok)
				
				assert.Len(t, classExpr.Body, 3)
				
				// final const
				finalBasic, ok := classExpr.Body[0].(*ast.ClassConstantDeclaration)
				assert.True(t, ok)
				assert.Equal(t, "", finalBasic.Visibility) // defaults to public
				assert.True(t, finalBasic.IsFinal)
				assert.False(t, finalBasic.IsAbstract)
				assert.Equal(t, "FINAL_BASIC", finalBasic.Constants[0].Name.(*ast.IdentifierNode).Name)
				
				// final public const
				finalPublic, ok := classExpr.Body[1].(*ast.ClassConstantDeclaration)
				assert.True(t, ok)
				assert.Equal(t, "public", finalPublic.Visibility)
				assert.True(t, finalPublic.IsFinal)
				assert.False(t, finalPublic.IsAbstract)
				assert.Equal(t, "FINAL_PUBLIC", finalPublic.Constants[0].Name.(*ast.IdentifierNode).Name)
				
				// final protected const
				finalProtected, ok := classExpr.Body[2].(*ast.ClassConstantDeclaration)
				assert.True(t, ok)
				assert.Equal(t, "protected", finalProtected.Visibility)
				assert.True(t, finalProtected.IsFinal)
				assert.False(t, finalProtected.IsAbstract)
				assert.Equal(t, "FINAL_PROTECTED", finalProtected.Constants[0].Name.(*ast.IdentifierNode).Name)
			},
		},
		{
			name: "multiple constants in one declaration",
			input: `<?php
class Test {
    const A = 1, B = 2, C = 3;
    public const X = 'x', Y = 'y';
    final protected const P = true, Q = false;
}`,
			validate: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 1)
				
				stmt, ok := program.Body[0].(*ast.ExpressionStatement)
				assert.True(t, ok)
				
				classExpr, ok := stmt.Expression.(*ast.ClassExpression)
				assert.True(t, ok)
				
				assert.Len(t, classExpr.Body, 3)
				
				// Basic multiple constants
				basicMultiple, ok := classExpr.Body[0].(*ast.ClassConstantDeclaration)
				assert.True(t, ok)
				assert.Len(t, basicMultiple.Constants, 3)
				assert.Equal(t, "A", basicMultiple.Constants[0].Name.(*ast.IdentifierNode).Name)
				assert.Equal(t, "B", basicMultiple.Constants[1].Name.(*ast.IdentifierNode).Name)
				assert.Equal(t, "C", basicMultiple.Constants[2].Name.(*ast.IdentifierNode).Name)
				
				// Public multiple constants
				publicMultiple, ok := classExpr.Body[1].(*ast.ClassConstantDeclaration)
				assert.True(t, ok)
				assert.Equal(t, "public", publicMultiple.Visibility)
				assert.Len(t, publicMultiple.Constants, 2)
				assert.Equal(t, "X", publicMultiple.Constants[0].Name.(*ast.IdentifierNode).Name)
				assert.Equal(t, "Y", publicMultiple.Constants[1].Name.(*ast.IdentifierNode).Name)
				
				// Final protected multiple constants
				finalProtectedMultiple, ok := classExpr.Body[2].(*ast.ClassConstantDeclaration)
				assert.True(t, ok)
				assert.Equal(t, "protected", finalProtectedMultiple.Visibility)
				assert.True(t, finalProtectedMultiple.IsFinal)
				assert.Len(t, finalProtectedMultiple.Constants, 2)
				assert.Equal(t, "P", finalProtectedMultiple.Constants[0].Name.(*ast.IdentifierNode).Name)
				assert.Equal(t, "Q", finalProtectedMultiple.Constants[1].Name.(*ast.IdentifierNode).Name)
			},
		},
		{
			name: "original failing case - final protected const",
			input: `<?php
class BaseUri {
    final protected const WHATWG_SPECIAL_SCHEMES = ['ftp' => 1, 'http' => 1];
}`,
			validate: func(t *testing.T, program *ast.Program) {
				assert.Len(t, program.Body, 1)
				
				stmt, ok := program.Body[0].(*ast.ExpressionStatement)
				assert.True(t, ok)
				
				classExpr, ok := stmt.Expression.(*ast.ClassExpression)
				assert.True(t, ok)
				
				nameIdent, ok := classExpr.Name.(*ast.IdentifierNode)
				assert.True(t, ok)
				assert.Equal(t, "BaseUri", nameIdent.Name)
				
				assert.Len(t, classExpr.Body, 1)
				constDecl, ok := classExpr.Body[0].(*ast.ClassConstantDeclaration)
				assert.True(t, ok)
				
				assert.Equal(t, "protected", constDecl.Visibility)
				assert.True(t, constDecl.IsFinal)
				assert.False(t, constDecl.IsAbstract)
				assert.Len(t, constDecl.Constants, 1)
				assert.Equal(t, "WHATWG_SPECIAL_SCHEMES", constDecl.Constants[0].Name.(*ast.IdentifierNode).Name)
				
				// Check that the array value is parsed correctly
				arrayExpr, ok := constDecl.Constants[0].Value.(*ast.ArrayExpression)
				assert.True(t, ok)
				assert.Len(t, arrayExpr.Elements, 2) // 'ftp' => 1, 'http' => 1
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