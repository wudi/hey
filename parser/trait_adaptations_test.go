package parser

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wudi/php-parser/ast"
	"github.com/wudi/php-parser/lexer"
)

func TestParsing_TraitAdaptations(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected func(t *testing.T, program *ast.Program)
	}{
		{
			name: "Simple trait usage without adaptations",
			input: `<?php
class TestClass {
    use TraitA;
}`,
			expected: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)
				
				classStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)
				
				classExpr, ok := classStmt.Expression.(*ast.ClassExpression)
				require.True(t, ok)
				
				require.Len(t, classExpr.Body, 1)
				useTraitStmt, ok := classExpr.Body[0].(*ast.UseTraitStatement)
				require.True(t, ok)
				
				assert.Len(t, useTraitStmt.Traits, 1)
				assert.Equal(t, "TraitA", useTraitStmt.Traits[0].Name)
				assert.Nil(t, useTraitStmt.Adaptations)
			},
		},
		{
			name: "Multiple traits usage without adaptations",
			input: `<?php
class TestClass {
    use TraitA, TraitB, TraitC;
}`,
			expected: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)
				
				classStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)
				
				classExpr, ok := classStmt.Expression.(*ast.ClassExpression)
				require.True(t, ok)
				
				require.Len(t, classExpr.Body, 1)
				useTraitStmt, ok := classExpr.Body[0].(*ast.UseTraitStatement)
				require.True(t, ok)
				
				assert.Len(t, useTraitStmt.Traits, 3)
				assert.Equal(t, "TraitA", useTraitStmt.Traits[0].Name)
				assert.Equal(t, "TraitB", useTraitStmt.Traits[1].Name)
				assert.Equal(t, "TraitC", useTraitStmt.Traits[2].Name)
				assert.Nil(t, useTraitStmt.Adaptations)
			},
		},
		{
			name: "Trait precedence (insteadof)",
			input: `<?php
class TestClass {
    use TraitA, TraitB {
        TraitA::foo insteadof TraitB;
    }
}`,
			expected: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)
				
				classStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)
				
				classExpr, ok := classStmt.Expression.(*ast.ClassExpression)
				require.True(t, ok)
				
				require.Len(t, classExpr.Body, 1)
				useTraitStmt, ok := classExpr.Body[0].(*ast.UseTraitStatement)
				require.True(t, ok)
				
				assert.Len(t, useTraitStmt.Traits, 2)
				require.Len(t, useTraitStmt.Adaptations, 1)
				
				adaptation := useTraitStmt.Adaptations[0]
				precedence, ok := adaptation.(*ast.TraitPrecedenceStatement)
				require.True(t, ok)
				
				assert.Equal(t, "TraitA", precedence.Method.Trait.Name)
				assert.Equal(t, "foo", precedence.Method.Method.Name)
				assert.Len(t, precedence.InsteadOf, 1)
				assert.Equal(t, "TraitB", precedence.InsteadOf[0].Name)
			},
		},
		{
			name: "Trait alias with new name",
			input: `<?php
class TestClass {
    use TraitA {
        foo as newFoo;
    }
}`,
			expected: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)
				
				classStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)
				
				classExpr, ok := classStmt.Expression.(*ast.ClassExpression)
				require.True(t, ok)
				
				require.Len(t, classExpr.Body, 1)
				useTraitStmt, ok := classExpr.Body[0].(*ast.UseTraitStatement)
				require.True(t, ok)
				
				require.Len(t, useTraitStmt.Adaptations, 1)
				
				adaptation := useTraitStmt.Adaptations[0]
				alias, ok := adaptation.(*ast.TraitAliasStatement)
				require.True(t, ok)
				
				assert.Nil(t, alias.Method.Trait) // Simple method reference
				assert.Equal(t, "foo", alias.Method.Method.Name)
				assert.Equal(t, "newFoo", alias.Alias.Name)
				assert.Empty(t, alias.Visibility)
			},
		},
		{
			name: "Trait alias with visibility change",
			input: `<?php
class TestClass {
    use TraitA {
        foo as private;
    }
}`,
			expected: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)
				
				classStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)
				
				classExpr, ok := classStmt.Expression.(*ast.ClassExpression)
				require.True(t, ok)
				
				require.Len(t, classExpr.Body, 1)
				useTraitStmt, ok := classExpr.Body[0].(*ast.UseTraitStatement)
				require.True(t, ok)
				
				require.Len(t, useTraitStmt.Adaptations, 1)
				
				adaptation := useTraitStmt.Adaptations[0]
				alias, ok := adaptation.(*ast.TraitAliasStatement)
				require.True(t, ok)
				
				assert.Equal(t, "foo", alias.Method.Method.Name)
				assert.Nil(t, alias.Alias) // No new name, only visibility change
				assert.Equal(t, "private", alias.Visibility)
			},
		},
		{
			name: "Trait alias with visibility and new name",
			input: `<?php
class TestClass {
    use TraitA {
        TraitA::bar as protected newBar;
    }
}`,
			expected: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)
				
				classStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)
				
				classExpr, ok := classStmt.Expression.(*ast.ClassExpression)
				require.True(t, ok)
				
				require.Len(t, classExpr.Body, 1)
				useTraitStmt, ok := classExpr.Body[0].(*ast.UseTraitStatement)
				require.True(t, ok)
				
				require.Len(t, useTraitStmt.Adaptations, 1)
				
				adaptation := useTraitStmt.Adaptations[0]
				alias, ok := adaptation.(*ast.TraitAliasStatement)
				require.True(t, ok)
				
				assert.Equal(t, "TraitA", alias.Method.Trait.Name)
				assert.Equal(t, "bar", alias.Method.Method.Name)
				assert.Equal(t, "newBar", alias.Alias.Name)
				assert.Equal(t, "protected", alias.Visibility)
			},
		},
		{
			name: "Multiple adaptations",
			input: `<?php
class TestClass {
    use TraitA, TraitB {
        TraitA::foo insteadof TraitB;
        TraitB::foo as fooFromB;
        bar as private privateBar;
        TraitA::baz as public publicBaz;
    }
}`,
			expected: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)
				
				classStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)
				
				classExpr, ok := classStmt.Expression.(*ast.ClassExpression)
				require.True(t, ok)
				
				require.Len(t, classExpr.Body, 1)
				useTraitStmt, ok := classExpr.Body[0].(*ast.UseTraitStatement)
				require.True(t, ok)
				
				require.Len(t, useTraitStmt.Adaptations, 4)
				
				// First adaptation - precedence
				precedence, ok := useTraitStmt.Adaptations[0].(*ast.TraitPrecedenceStatement)
				require.True(t, ok)
				assert.Equal(t, "TraitA", precedence.Method.Trait.Name)
				assert.Equal(t, "foo", precedence.Method.Method.Name)
				assert.Equal(t, "TraitB", precedence.InsteadOf[0].Name)
				
				// Second adaptation - alias
				alias1, ok := useTraitStmt.Adaptations[1].(*ast.TraitAliasStatement)
				require.True(t, ok)
				assert.Equal(t, "TraitB", alias1.Method.Trait.Name)
				assert.Equal(t, "foo", alias1.Method.Method.Name)
				assert.Equal(t, "fooFromB", alias1.Alias.Name)
				
				// Third adaptation - visibility change
				alias2, ok := useTraitStmt.Adaptations[2].(*ast.TraitAliasStatement)
				require.True(t, ok)
				assert.Nil(t, alias2.Method.Trait)
				assert.Equal(t, "bar", alias2.Method.Method.Name)
				assert.Equal(t, "privateBar", alias2.Alias.Name)
				assert.Equal(t, "private", alias2.Visibility)
				
				// Fourth adaptation - visibility with new name
				alias3, ok := useTraitStmt.Adaptations[3].(*ast.TraitAliasStatement)
				require.True(t, ok)
				assert.Equal(t, "TraitA", alias3.Method.Trait.Name)
				assert.Equal(t, "baz", alias3.Method.Method.Name)
				assert.Equal(t, "publicBaz", alias3.Alias.Name)
				assert.Equal(t, "public", alias3.Visibility)
			},
		},
		{
			name: "Multiple insteadof traits",
			input: `<?php
class TestClass {
    use TraitA, TraitB, TraitC {
        TraitA::foo insteadof TraitB, TraitC;
    }
}`,
			expected: func(t *testing.T, program *ast.Program) {
				require.Len(t, program.Body, 1)
				
				classStmt, ok := program.Body[0].(*ast.ExpressionStatement)
				require.True(t, ok)
				
				classExpr, ok := classStmt.Expression.(*ast.ClassExpression)
				require.True(t, ok)
				
				require.Len(t, classExpr.Body, 1)
				useTraitStmt, ok := classExpr.Body[0].(*ast.UseTraitStatement)
				require.True(t, ok)
				
				require.Len(t, useTraitStmt.Adaptations, 1)
				
				adaptation := useTraitStmt.Adaptations[0]
				precedence, ok := adaptation.(*ast.TraitPrecedenceStatement)
				require.True(t, ok)
				
				assert.Equal(t, "TraitA", precedence.Method.Trait.Name)
				assert.Equal(t, "foo", precedence.Method.Method.Name)
				assert.Len(t, precedence.InsteadOf, 2)
				assert.Equal(t, "TraitB", precedence.InsteadOf[0].Name)
				assert.Equal(t, "TraitC", precedence.InsteadOf[1].Name)
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

func TestParsing_TraitAdaptations_ErrorCases(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name: "Missing trait name after use",
			input: `<?php
class TestClass {
    use ;
}`,
			expectedError: "expected trait name",
		},
		{
			name: "Missing method name after ::",
			input: `<?php
class TestClass {
    use TraitA {
        TraitA:: as foo;
    }
}`,
			expectedError: "expected method name",
		},
		{
			name: "Missing insteadof trait name",
			input: `<?php
class TestClass {
    use TraitA {
        foo insteadof ;
    }
}`,
			expectedError: "expected trait name",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			l := lexer.New(test.input)
			parser := New(l)
			program := parser.ParseProgram()
			
			require.NotNil(t, program)
			errors := parser.Errors()
			require.NotEmpty(t, errors, "Expected parser errors but got none")
			
			found := false
			for _, err := range errors {
				if containsSubstring(err, test.expectedError) {
					found = true
					break
				}
			}
			assert.True(t, found, "Expected error containing '%s', but got: %v", test.expectedError, errors)
		})
	}
}

func containsSubstring(str, substr string) bool {
	return len(str) >= len(substr) && str[:len(substr)] == substr
}