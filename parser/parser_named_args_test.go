package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wudi/php-parser/ast"
	"github.com/wudi/php-parser/lexer"
)

func TestParsing_NamedArguments(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected func(t *testing.T, result ast.Node)
	}{
		{
			name:  "single named argument",
			input: `<?php test(name: "John");`,
			expected: func(t *testing.T, result ast.Node) {
				program := result.(*ast.Program)
				require.Len(t, program.Body, 1)

				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				call, ok := exprStmt.Expression.(*ast.CallExpression)
				require.True(t, ok)
				require.Len(t, call.Arguments, 1)

				namedArg, ok := call.Arguments[0].(*ast.NamedArgument)
				require.True(t, ok)

				assert.Equal(t, "name", namedArg.Name.Name)
				
				stringLiteral, ok := namedArg.Value.(*ast.StringLiteral)
				require.True(t, ok)
				assert.Equal(t, "John", stringLiteral.Value)
			},
		},
		{
			name:  "multiple named arguments",
			input: `<?php calculate(x: 10, y: 20, operation: "add");`,
			expected: func(t *testing.T, result ast.Node) {
				program := result.(*ast.Program)
				require.Len(t, program.Body, 1)

				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				call, ok := exprStmt.Expression.(*ast.CallExpression)
				require.True(t, ok)
				require.Len(t, call.Arguments, 3)

				// First argument: x: 10
				namedArg1, ok := call.Arguments[0].(*ast.NamedArgument)
				require.True(t, ok)
				assert.Equal(t, "x", namedArg1.Name.Name)

				num1, ok := namedArg1.Value.(*ast.NumberLiteral)
				require.True(t, ok)
				assert.Equal(t, "10", num1.Value)

				// Second argument: y: 20
				namedArg2, ok := call.Arguments[1].(*ast.NamedArgument)
				require.True(t, ok)
				assert.Equal(t, "y", namedArg2.Name.Name)

				num2, ok := namedArg2.Value.(*ast.NumberLiteral)
				require.True(t, ok)
				assert.Equal(t, "20", num2.Value)

				// Third argument: operation: "add"
				namedArg3, ok := call.Arguments[2].(*ast.NamedArgument)
				require.True(t, ok)
				assert.Equal(t, "operation", namedArg3.Name.Name)

				str, ok := namedArg3.Value.(*ast.StringLiteral)
				require.True(t, ok)
				assert.Equal(t, "add", str.Value)
			},
		},
		{
			name:  "mixed positional and named arguments",
			input: `<?php mixed_args(1, 2, name: "Alice", value: 42);`,
			expected: func(t *testing.T, result ast.Node) {
				program := result.(*ast.Program)
				require.Len(t, program.Body, 1)

				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				call, ok := exprStmt.Expression.(*ast.CallExpression)
				require.True(t, ok)
				require.Len(t, call.Arguments, 4)

				// First argument: 1 (positional)
				num1, ok := call.Arguments[0].(*ast.NumberLiteral)
				require.True(t, ok)
				assert.Equal(t, "1", num1.Value)

				// Second argument: 2 (positional)
				num2, ok := call.Arguments[1].(*ast.NumberLiteral)
				require.True(t, ok)
				assert.Equal(t, "2", num2.Value)

				// Third argument: name: "Alice" (named)
				namedArg1, ok := call.Arguments[2].(*ast.NamedArgument)
				require.True(t, ok)
				assert.Equal(t, "name", namedArg1.Name.Name)

				str, ok := namedArg1.Value.(*ast.StringLiteral)
				require.True(t, ok)
				assert.Equal(t, "Alice", str.Value)

				// Fourth argument: value: 42 (named)
				namedArg2, ok := call.Arguments[3].(*ast.NamedArgument)
				require.True(t, ok)
				assert.Equal(t, "value", namedArg2.Name.Name)

				num4, ok := namedArg2.Value.(*ast.NumberLiteral)
				require.True(t, ok)
				assert.Equal(t, "42", num4.Value)
			},
		},
		{
			name:  "named argument with variable value",
			input: `<?php test(name: $userName);`,
			expected: func(t *testing.T, result ast.Node) {
				program := result.(*ast.Program)
				require.Len(t, program.Body, 1)

				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				call, ok := exprStmt.Expression.(*ast.CallExpression)
				require.True(t, ok)
				require.Len(t, call.Arguments, 1)

				namedArg, ok := call.Arguments[0].(*ast.NamedArgument)
				require.True(t, ok)
				assert.Equal(t, "name", namedArg.Name.Name)

				variable, ok := namedArg.Value.(*ast.Variable)
				require.True(t, ok)
				assert.Equal(t, "$userName", variable.Name)
			},
		},
		{
			name:  "named argument with complex expression",
			input: `<?php test(result: $a + $b * 2);`,
			expected: func(t *testing.T, result ast.Node) {
				program := result.(*ast.Program)
				require.Len(t, program.Body, 1)

				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				call, ok := exprStmt.Expression.(*ast.CallExpression)
				require.True(t, ok)
				require.Len(t, call.Arguments, 1)

				namedArg, ok := call.Arguments[0].(*ast.NamedArgument)
				require.True(t, ok)
				assert.Equal(t, "result", namedArg.Name.Name)

				// Should be a binary expression: $a + ($b * 2)
				binaryExpr, ok := namedArg.Value.(*ast.BinaryExpression)
				require.True(t, ok)
				assert.Equal(t, "+", binaryExpr.Operator)
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

func TestParsing_NamedArgumentsErrors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name:          "named argument without value",
			input:         `<?php test(name:);`,
			expectedError: "expected next token to be `)`, got `;`",
		},
		{
			name:          "named argument without colon",
			input:         `<?php test(name "value");`,
			expectedError: "expected next token to be `)`, got `T_CONSTANT_ENCAPSED_STRING`",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			result := p.ParseProgram()

			// Should have parsing errors
			errors := p.Errors()
			require.NotEmpty(t, errors, "Expected parsing errors but got none")

			// Check if expected error message is present
			found := false
			for _, err := range errors {
				if contains(err, tt.expectedError) {
					found = true
					break
				}
			}
			assert.True(t, found, "Expected error containing '%s', got: %v", tt.expectedError, errors)

			// Result should still be parseable (error recovery)
			require.NotNil(t, result)
		})
	}
}