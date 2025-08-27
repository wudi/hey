package parser

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wudi/php-parser/ast"
	"github.com/wudi/php-parser/lexer"
)

func TestParsing_PropertyHooks_Complete(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected func(t *testing.T, program *ast.Program)
	}{
		{
			name: "Simple get hook with arrow syntax",
			input: `<?php
class Example {
    public string $name {
        get => 'test';
    }
}`,
			expected: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)
				
				classStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)
				
				classExpr, ok := classStmt.Expression.(*ast.ClassExpression)
				require.True(t, ok)
				
				require.Len(t, classExpr.Body, 1)
				hookedProp, ok := classExpr.Body[0].(*ast.HookedPropertyDeclaration)
				require.True(t, ok)
				
				assert.Equal(t, "public", hookedProp.Visibility)
				assert.Equal(t, "name", hookedProp.Name)
				assert.Equal(t, "string", hookedProp.Type.Name)
				
				require.Len(t, hookedProp.Hooks, 1)
				getHook := hookedProp.Hooks[0]
				assert.Equal(t, "get", getHook.Type)
				assert.False(t, getHook.ByRef)
				assert.NotNil(t, getHook.Body)
				assert.Nil(t, getHook.Parameter)
			},
		},
		{
			name: "Set hook with parameter and arrow syntax", 
			input: `<?php
class Example {
    public string $email {
        set(string $value) => strtolower($value);
    }
}`,
			expected: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)
				
				classStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)
				
				classExpr, ok := classStmt.Expression.(*ast.ClassExpression)
				require.True(t, ok)
				
				require.Len(t, classExpr.Body, 1)
				hookedProp, ok := classExpr.Body[0].(*ast.HookedPropertyDeclaration)
				require.True(t, ok)
				
				assert.Equal(t, "public", hookedProp.Visibility)
				assert.Equal(t, "email", hookedProp.Name)
				
				require.Len(t, hookedProp.Hooks, 1)
				setHook := hookedProp.Hooks[0]
				assert.Equal(t, "set", setHook.Type)
				assert.False(t, setHook.ByRef)
				assert.NotNil(t, setHook.Body)
				assert.NotNil(t, setHook.Parameter)
				assert.Equal(t, "$value", setHook.Parameter.Name)
			},
		},
		{
			name: "Both get and set hooks",
			input: `<?php
class Example {
    public string $fullName {
        get => $this->first . ' ' . $this->last;
        set(string $value) => $this->parseFullName($value);
    }
}`,
			expected: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)
				
				classStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)
				
				classExpr, ok := classStmt.Expression.(*ast.ClassExpression)
				require.True(t, ok)
				
				require.Len(t, classExpr.Body, 1)
				hookedProp, ok := classExpr.Body[0].(*ast.HookedPropertyDeclaration)
				require.True(t, ok)
				
				assert.Equal(t, "public", hookedProp.Visibility)
				assert.Equal(t, "fullName", hookedProp.Name)
				
				require.Len(t, hookedProp.Hooks, 2)
				
				getHook := hookedProp.Hooks[0]
				assert.Equal(t, "get", getHook.Type)
				assert.NotNil(t, getHook.Body)
				
				setHook := hookedProp.Hooks[1]
				assert.Equal(t, "set", setHook.Type)
				assert.NotNil(t, setHook.Body)
				assert.NotNil(t, setHook.Parameter)
			},
		},
		{
			name: "Reference get hook",
			input: `<?php
class Example {
    public array $data {
        &get => $this->internalData;
    }
}`,
			expected: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)
				
				classStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)
				
				classExpr, ok := classStmt.Expression.(*ast.ClassExpression)
				require.True(t, ok)
				
				require.Len(t, classExpr.Body, 1)
				hookedProp, ok := classExpr.Body[0].(*ast.HookedPropertyDeclaration)
				require.True(t, ok)
				
				require.Len(t, hookedProp.Hooks, 1)
				getHook := hookedProp.Hooks[0]
				assert.Equal(t, "get", getHook.Type)
				assert.True(t, getHook.ByRef)
				assert.NotNil(t, getHook.Body)
			},
		},
		{
			name: "Mixed hooked and regular properties",
			input: `<?php
class Example {
    public string $name {
        get => 'test';
    }
    private string $internal;
}`,
			expected: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)
				
				classStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)
				
				classExpr, ok := classStmt.Expression.(*ast.ClassExpression)
				require.True(t, ok)
				
				require.Len(t, classExpr.Body, 2)
				
				// First property should be hooked
				hookedProp, ok := classExpr.Body[0].(*ast.HookedPropertyDeclaration)
				require.True(t, ok)
				assert.Equal(t, "name", hookedProp.Name)
				require.Len(t, hookedProp.Hooks, 1)
				
				// Second property should be regular
				regularProp, ok := classExpr.Body[1].(*ast.PropertyDeclaration)
				require.True(t, ok)
				assert.Equal(t, "internal", regularProp.Name)
				assert.Equal(t, "private", regularProp.Visibility)
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