package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wudi/php-parser/ast"
	"github.com/wudi/php-parser/lexer"
)

func TestParsing_AnonymousClass(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected func(t *testing.T, result ast.Node)
	}{
		{
			name:  "basic anonymous class",
			input: `<?php $obj = new class {};`,
			expected: func(t *testing.T, result ast.Node) {
				program := result.(*ast.Program)
				require.Len(t, program.Body, 1)

				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				assign, ok := exprStmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)

				anonClass, ok := assign.Right.(*ast.AnonymousClass)
				require.True(t, ok)

				assert.Empty(t, anonClass.Arguments)
				assert.Nil(t, anonClass.Extends)
				assert.Empty(t, anonClass.Implements)
				assert.Empty(t, anonClass.Body)
			},
		},
		{
			name:  "anonymous class with constructor arguments",
			input: `<?php $obj = new class($arg1, $arg2) {};`,
			expected: func(t *testing.T, result ast.Node) {
				program := result.(*ast.Program)
				require.Len(t, program.Body, 1)

				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				assign, ok := exprStmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)

				anonClass, ok := assign.Right.(*ast.AnonymousClass)
				require.True(t, ok)

				require.Len(t, anonClass.Arguments, 2)

				arg1, ok := anonClass.Arguments[0].(*ast.Variable)
				require.True(t, ok)
				assert.Equal(t, "$arg1", arg1.Name)

				arg2, ok := anonClass.Arguments[1].(*ast.Variable)
				require.True(t, ok)
				assert.Equal(t, "$arg2", arg2.Name)
			},
		},
		{
			name:  "anonymous class with extends",
			input: `<?php $obj = new class extends BaseClass {};`,
			expected: func(t *testing.T, result ast.Node) {
				program := result.(*ast.Program)
				require.Len(t, program.Body, 1)

				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				assign, ok := exprStmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)

				anonClass, ok := assign.Right.(*ast.AnonymousClass)
				require.True(t, ok)

				require.NotNil(t, anonClass.Extends)
				extends, ok := anonClass.Extends.(*ast.IdentifierNode)
				require.True(t, ok)
				assert.Equal(t, "BaseClass", extends.Name)
			},
		},
		{
			name:  "anonymous class with implements",
			input: `<?php $obj = new class implements Interface1, Interface2 {};`,
			expected: func(t *testing.T, result ast.Node) {
				program := result.(*ast.Program)
				require.Len(t, program.Body, 1)

				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				assign, ok := exprStmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)

				anonClass, ok := assign.Right.(*ast.AnonymousClass)
				require.True(t, ok)

				require.Len(t, anonClass.Implements, 2)

				iface1, ok := anonClass.Implements[0].(*ast.IdentifierNode)
				require.True(t, ok)
				assert.Equal(t, "Interface1", iface1.Name)

				iface2, ok := anonClass.Implements[1].(*ast.IdentifierNode)
				require.True(t, ok)
				assert.Equal(t, "Interface2", iface2.Name)
			},
		},
		{
			name:  "anonymous class with class body",
			input: `<?php $obj = new class { private $prop; public function method() {} };`,
			expected: func(t *testing.T, result ast.Node) {
				program := result.(*ast.Program)
				require.Len(t, program.Body, 1)

				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				assign, ok := exprStmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)

				anonClass, ok := assign.Right.(*ast.AnonymousClass)
				require.True(t, ok)

				// Should have property and method in body
				require.Len(t, anonClass.Body, 2)

				// First should be property declaration
				_, ok = anonClass.Body[0].(*ast.PropertyDeclaration)
				require.True(t, ok, "Expected PropertyDeclaration, got %T", anonClass.Body[0])

				// Second should be method declaration
				_, ok = anonClass.Body[1].(*ast.FunctionDeclaration)
				require.True(t, ok, "Expected FunctionDeclaration, got %T", anonClass.Body[1])
			},
		},
		{
			name:  "anonymous class with final modifier",
			input: `<?php $obj = new final class {};`,
			expected: func(t *testing.T, result ast.Node) {
				program := result.(*ast.Program)
				require.Len(t, program.Body, 1)

				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				assign, ok := exprStmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)

				anonClass, ok := assign.Right.(*ast.AnonymousClass)
				require.True(t, ok)

				assert.Equal(t, []string{"final"}, anonClass.Modifiers)
				assert.Empty(t, anonClass.Arguments)
				assert.Nil(t, anonClass.Extends)
				assert.Empty(t, anonClass.Implements)
				assert.Empty(t, anonClass.Body)
			},
		},
		{
			name:  "anonymous class with readonly modifier",
			input: `<?php $obj = new readonly class {};`,
			expected: func(t *testing.T, result ast.Node) {
				program := result.(*ast.Program)
				require.Len(t, program.Body, 1)

				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				assign, ok := exprStmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)

				anonClass, ok := assign.Right.(*ast.AnonymousClass)
				require.True(t, ok)

				assert.Equal(t, []string{"readonly"}, anonClass.Modifiers)
				assert.Empty(t, anonClass.Arguments)
				assert.Nil(t, anonClass.Extends)
				assert.Empty(t, anonClass.Implements)
				assert.Empty(t, anonClass.Body)
			},
		},
		{
			name:  "anonymous class with abstract modifier",
			input: `<?php $obj = new abstract class {};`,
			expected: func(t *testing.T, result ast.Node) {
				program := result.(*ast.Program)
				require.Len(t, program.Body, 1)

				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				assign, ok := exprStmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)

				anonClass, ok := assign.Right.(*ast.AnonymousClass)
				require.True(t, ok)

				assert.Equal(t, []string{"abstract"}, anonClass.Modifiers)
				assert.Empty(t, anonClass.Arguments)
				assert.Nil(t, anonClass.Extends)
				assert.Empty(t, anonClass.Implements)
				assert.Empty(t, anonClass.Body)
			},
		},
		{
			name:  "anonymous class with multiple modifiers and complex structure",
			input: `<?php $obj = new final readonly class($param) extends Parent implements Interface { private $prop; };`,
			expected: func(t *testing.T, result ast.Node) {
				program := result.(*ast.Program)
				require.Len(t, program.Body, 1)

				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				assign, ok := exprStmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)

				anonClass, ok := assign.Right.(*ast.AnonymousClass)
				require.True(t, ok)

				assert.Equal(t, []string{"final", "readonly"}, anonClass.Modifiers)
				
				// Check constructor argument
				require.Len(t, anonClass.Arguments, 1)
				arg, ok := anonClass.Arguments[0].(*ast.Variable)
				require.True(t, ok)
				assert.Equal(t, "$param", arg.Name)

				// Check extends
				require.NotNil(t, anonClass.Extends)
				extends, ok := anonClass.Extends.(*ast.IdentifierNode)
				require.True(t, ok)
				assert.Equal(t, "Parent", extends.Name)

				// Check implements
				require.Len(t, anonClass.Implements, 1)
				iface, ok := anonClass.Implements[0].(*ast.IdentifierNode)
				require.True(t, ok)
				assert.Equal(t, "Interface", iface.Name)

				// Check class body
				require.Len(t, anonClass.Body, 1)
			},
		},
		{
			name:  "anonymous class with attributes",
			input: `<?php $obj = new #[Attribute] class {};`,
			expected: func(t *testing.T, result ast.Node) {
				program := result.(*ast.Program)
				require.Len(t, program.Body, 1)

				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				assign, ok := exprStmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)

				anonClass, ok := assign.Right.(*ast.AnonymousClass)
				require.True(t, ok)

				// Check attributes
				require.Len(t, anonClass.Attributes, 1)
				attrGroup := anonClass.Attributes[0]
				require.Len(t, attrGroup.Attributes, 1)
				attr := attrGroup.Attributes[0]
				assert.Equal(t, "Attribute", attr.Name.Name)
				assert.Empty(t, attr.Arguments)

				assert.Empty(t, anonClass.Modifiers)
				assert.Empty(t, anonClass.Arguments)
				assert.Nil(t, anonClass.Extends)
				assert.Empty(t, anonClass.Implements)
				assert.Empty(t, anonClass.Body)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			result := p.ParseProgram()

			require.NotNil(t, result)
			assert.Empty(t, p.Errors(), "Parser errors: %v", p.Errors())

			tt.expected(t, result)
		})
	}
}