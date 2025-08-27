package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wudi/php-parser/ast"
	"github.com/wudi/php-parser/lexer"
)

func TestParsing_SpreadSyntax(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected func(t *testing.T, result ast.Node)
	}{
		{
			name:  "array spread with single array",
			input: `<?php $arr2 = [...$arr1];`,
			expected: func(t *testing.T, result ast.Node) {
				program, ok := result.(*ast.Program)
				require.True(t, ok)
				require.Len(t, program.Body, 1)

				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				assign, ok := exprStmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)

				arrayExpr, ok := assign.Right.(*ast.ArrayExpression)
				require.True(t, ok)
				require.Len(t, arrayExpr.Elements, 1)

				spread, ok := arrayExpr.Elements[0].(*ast.SpreadExpression)
				require.True(t, ok)

				variable, ok := spread.Argument.(*ast.Variable)
				require.True(t, ok)
				assert.Equal(t, "$arr1", variable.Name)
			},
		},
		{
			name:  "array spread with mixed elements",
			input: `<?php $arr2 = [...$arr1, 4, 5];`,
			expected: func(t *testing.T, result ast.Node) {
				program, ok := result.(*ast.Program)
				require.True(t, ok)
				require.Len(t, program.Body, 1)

				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				assign, ok := exprStmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)

				arrayExpr, ok := assign.Right.(*ast.ArrayExpression)
				require.True(t, ok)
				require.Len(t, arrayExpr.Elements, 3)

				// First element: spread
				spread, ok := arrayExpr.Elements[0].(*ast.SpreadExpression)
				require.True(t, ok)
				variable, ok := spread.Argument.(*ast.Variable)
				require.True(t, ok)
				assert.Equal(t, "$arr1", variable.Name)

				// Second element: number 4
				num1, ok := arrayExpr.Elements[1].(*ast.NumberLiteral)
				require.True(t, ok)
				assert.Equal(t, "4", num1.Value)

				// Third element: number 5
				num2, ok := arrayExpr.Elements[2].(*ast.NumberLiteral)
				require.True(t, ok)
				assert.Equal(t, "5", num2.Value)
			},
		},
		{
			name:  "array() spread with mixed elements",
			input: `<?php $arr3 = array(0, ...$arr1, 6);`,
			expected: func(t *testing.T, result ast.Node) {
				program, ok := result.(*ast.Program)
				require.True(t, ok)
				require.Len(t, program.Body, 1)

				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				assign, ok := exprStmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)

				arrayExpr, ok := assign.Right.(*ast.ArrayExpression)
				require.True(t, ok)
				require.Len(t, arrayExpr.Elements, 3)

				// First element: number 0
				num0, ok := arrayExpr.Elements[0].(*ast.NumberLiteral)
				require.True(t, ok)
				assert.Equal(t, "0", num0.Value)

				// Second element: spread
				spread, ok := arrayExpr.Elements[1].(*ast.SpreadExpression)
				require.True(t, ok)
				variable, ok := spread.Argument.(*ast.Variable)
				require.True(t, ok)
				assert.Equal(t, "$arr1", variable.Name)

				// Third element: number 6
				num6, ok := arrayExpr.Elements[2].(*ast.NumberLiteral)
				require.True(t, ok)
				assert.Equal(t, "6", num6.Value)
			},
		},
		{
			name:  "function call with spread arguments",
			input: `<?php $result = test(...$arr1);`,
			expected: func(t *testing.T, result ast.Node) {
				program, ok := result.(*ast.Program)
				require.True(t, ok)
				require.Len(t, program.Body, 1)

				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				assign, ok := exprStmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)

				call, ok := assign.Right.(*ast.CallExpression)
				require.True(t, ok)
				require.Len(t, call.Arguments, 1)

				spread, ok := call.Arguments[0].(*ast.SpreadExpression)
				require.True(t, ok)

				variable, ok := spread.Argument.(*ast.Variable)
				require.True(t, ok)
				assert.Equal(t, "$arr1", variable.Name)
			},
		},
		{
			name:  "function call with mixed arguments",
			input: `<?php $mixed = test(1, ...[2, 3]);`,
			expected: func(t *testing.T, result ast.Node) {
				program, ok := result.(*ast.Program)
				require.True(t, ok)
				require.Len(t, program.Body, 1)

				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				assign, ok := exprStmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)

				call, ok := assign.Right.(*ast.CallExpression)
				require.True(t, ok)
				require.Len(t, call.Arguments, 2)

				// First argument: number 1
				num1, ok := call.Arguments[0].(*ast.NumberLiteral)
				require.True(t, ok)
				assert.Equal(t, "1", num1.Value)

				// Second argument: spread of array literal
				spread, ok := call.Arguments[1].(*ast.SpreadExpression)
				require.True(t, ok)

				arrayExpr, ok := spread.Argument.(*ast.ArrayExpression)
				require.True(t, ok)
				require.Len(t, arrayExpr.Elements, 2)

				num2, ok := arrayExpr.Elements[0].(*ast.NumberLiteral)
				require.True(t, ok)
				assert.Equal(t, "2", num2.Value)

				num3, ok := arrayExpr.Elements[1].(*ast.NumberLiteral)
				require.True(t, ok)
				assert.Equal(t, "3", num3.Value)
			},
		},
		{
			name:  "multiple spread in array",
			input: `<?php $arr = [...$a, ...$b, ...$c];`,
			expected: func(t *testing.T, result ast.Node) {
				program, ok := result.(*ast.Program)
				require.True(t, ok)
				require.Len(t, program.Body, 1)

				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				assign, ok := exprStmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)

				arrayExpr, ok := assign.Right.(*ast.ArrayExpression)
				require.True(t, ok)
				require.Len(t, arrayExpr.Elements, 3)

				// All elements should be spread expressions
				for i, element := range arrayExpr.Elements {
					spread, ok := element.(*ast.SpreadExpression)
					require.True(t, ok, "Element %d should be spread expression", i)

					variable, ok := spread.Argument.(*ast.Variable)
					require.True(t, ok)

					expectedNames := []string{"$a", "$b", "$c"}
					assert.Equal(t, expectedNames[i], variable.Name)
				}
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

func TestParsing_SpreadSyntaxErrors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name:          "spread without expression",
			input:         `<?php $arr = [...];`,
			expectedError: "expected ',' or ']' in array",
		},
		{
			name:          "spread in wrong context",
			input:         `<?php $x = ...5;`,
			expectedError: "no prefix parse function for `T_ELLIPSIS`",
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

// Helper function to check if error message contains expected text
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr ||
		   len(s) > len(substr) && 
		   (s[len(s)-len(substr):] == substr ||
		    indexOf(s, substr) != -1)
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}