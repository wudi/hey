package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wudi/php-parser/ast"
	"github.com/wudi/php-parser/lexer"
)

func TestParsing_AttributesEnhanced(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected func(t *testing.T, result ast.Node)
	}{
		{
			name:  "single attribute without parameters",
			input: `<?php $attr = #[Route];`,
			expected: func(t *testing.T, result ast.Node) {
				program := result.(*ast.Program)
				require.Len(t, program.Body, 1)

				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				assign, ok := exprStmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)

				attrGroup, ok := assign.Right.(*ast.AttributeGroup)
				require.True(t, ok)

				require.Len(t, attrGroup.Attributes, 1)
				assert.Equal(t, "Route", attrGroup.Attributes[0].Name.Name)
				assert.Empty(t, attrGroup.Attributes[0].Arguments)
			},
		},
		{
			name:  "single attribute with parameters",
			input: `<?php $attr = #[Route("/api/users")];`,
			expected: func(t *testing.T, result ast.Node) {
				program := result.(*ast.Program)
				require.Len(t, program.Body, 1)

				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				assign, ok := exprStmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)

				attrGroup, ok := assign.Right.(*ast.AttributeGroup)
				require.True(t, ok)

				require.Len(t, attrGroup.Attributes, 1)
				attr := attrGroup.Attributes[0]
				assert.Equal(t, "Route", attr.Name.Name)
				require.Len(t, attr.Arguments, 1)

				arg, ok := attr.Arguments[0].(*ast.StringLiteral)
				require.True(t, ok)
				assert.Equal(t, "/api/users", arg.Value)
			},
		},
		{
			name:  "attribute group with multiple attributes",
			input: `<?php $expr = #[Route("/api"), Method("GET")];`,
			expected: func(t *testing.T, result ast.Node) {
				program := result.(*ast.Program)
				require.Len(t, program.Body, 1)

				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				assign, ok := exprStmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)

				attrGroup, ok := assign.Right.(*ast.AttributeGroup)
				require.True(t, ok, "Expected AttributeGroup, got %T", assign.Right)

				require.Len(t, attrGroup.Attributes, 2)

				// First attribute: Route("/api")
				assert.Equal(t, "Route", attrGroup.Attributes[0].Name.Name)
				require.Len(t, attrGroup.Attributes[0].Arguments, 1)
				
				arg1, ok := attrGroup.Attributes[0].Arguments[0].(*ast.StringLiteral)
				require.True(t, ok)
				assert.Equal(t, "/api", arg1.Value)

				// Second attribute: Method("GET")
				assert.Equal(t, "Method", attrGroup.Attributes[1].Name.Name)
				require.Len(t, attrGroup.Attributes[1].Arguments, 1)
				
				arg2, ok := attrGroup.Attributes[1].Arguments[0].(*ast.StringLiteral)
				require.True(t, ok)
				assert.Equal(t, "GET", arg2.Value)
			},
		},
		{
			name:  "attribute with named parameters",
			input: `<?php $expr = #[Cache(ttl: 3600, tags: ["users"])];`,
			expected: func(t *testing.T, result ast.Node) {
				program := result.(*ast.Program)
				require.Len(t, program.Body, 1)

				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				assign, ok := exprStmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)

				attrGroup, ok := assign.Right.(*ast.AttributeGroup)
				require.True(t, ok)

				require.Len(t, attrGroup.Attributes, 1)
				attr := attrGroup.Attributes[0]

				assert.Equal(t, "Cache", attr.Name.Name)
				require.Len(t, attr.Arguments, 2)

				// First argument: ttl: 3600
				namedArg1, ok := attr.Arguments[0].(*ast.NamedArgument)
				require.True(t, ok)
				assert.Equal(t, "ttl", namedArg1.Name.Name)

				num, ok := namedArg1.Value.(*ast.NumberLiteral)
				require.True(t, ok)
				assert.Equal(t, "3600", num.Value)

				// Second argument: tags: ["users"]
				namedArg2, ok := attr.Arguments[1].(*ast.NamedArgument)
				require.True(t, ok)
				assert.Equal(t, "tags", namedArg2.Name.Name)

				arrayExpr, ok := namedArg2.Value.(*ast.ArrayExpression)
				require.True(t, ok)
				require.Len(t, arrayExpr.Elements, 1)
			},
		},
		{
			name:  "multiple attributes without parameters",
			input: `<?php $expr = #[Deprecated, Internal, Final];`,
			expected: func(t *testing.T, result ast.Node) {
				program := result.(*ast.Program)
				require.Len(t, program.Body, 1)

				exprStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)

				assign, ok := exprStmt.Expression.(*ast.AssignmentExpression)
				require.True(t, ok)

				attrGroup, ok := assign.Right.(*ast.AttributeGroup)
				require.True(t, ok)

				require.Len(t, attrGroup.Attributes, 3)

				expectedNames := []string{"Deprecated", "Internal", "Final"}
				for i, attr := range attrGroup.Attributes {
					assert.Equal(t, expectedNames[i], attr.Name.Name)
					assert.Empty(t, attr.Arguments)
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

func TestParsing_AttributeErrors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name:          "unclosed attribute group",
			input:         `<?php #[Route("/api");`,
			expectedError: "expected next token to be `]`",
		},
		{
			name:          "attribute without name",
			input:         `<?php #[];`,
			expectedError: "expected attribute name",
		},
		{
			name:          "malformed attribute parameters",
			input:         `<?php #[Route(;`,
			expectedError: "expected next token to be `)`",
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