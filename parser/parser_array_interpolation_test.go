package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wudi/php-parser/ast"
	"github.com/wudi/php-parser/lexer"
)

// TestParser_InterpolatedStringArrayAccess tests that the parser correctly
// recognizes array access within interpolated strings
func TestParser_InterpolatedStringArrayAccess(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		validate func(t *testing.T, prog *ast.Program)
	}{
		{
			name:  "simple array access in interpolated string",
			input: `<?php echo "$arr[0]";`,
			validate: func(t *testing.T, prog *ast.Program) {
				assert.Len(t, prog.Body, 1)
				
				echoStmt, ok := prog.Body[0].(*ast.EchoStatement)
				assert.True(t, ok, "Expected echo statement")
				assert.Len(t, echoStmt.Arguments.Arguments, 1)
				
				interpolatedStr, ok := echoStmt.Arguments.Arguments[0].(*ast.InterpolatedStringExpression)
				assert.True(t, ok, "Expected interpolated string")
				assert.Len(t, interpolatedStr.Parts, 1)
				
				arrayAccess, ok := interpolatedStr.Parts[0].(*ast.ArrayAccessExpression)
				assert.True(t, ok, "Expected array access expression")
				
				variable, ok := arrayAccess.Array.(*ast.Variable)
				assert.True(t, ok, "Expected variable")
				assert.Equal(t, "$arr", variable.Name)
				
				index, ok := (*arrayAccess.Index).(*ast.NumberLiteral)
				assert.True(t, ok, "Expected number literal")
				assert.Equal(t, "0", index.Value)
			},
		},
		{
			name:  "array access with variable index",
			input: `<?php echo "$arr[$i]";`,
			validate: func(t *testing.T, prog *ast.Program) {
				assert.Len(t, prog.Body, 1)
				
				echoStmt, ok := prog.Body[0].(*ast.EchoStatement)
				assert.True(t, ok, "Expected echo statement")
				assert.Len(t, echoStmt.Arguments.Arguments, 1)
				
				interpolatedStr, ok := echoStmt.Arguments.Arguments[0].(*ast.InterpolatedStringExpression)
				assert.True(t, ok, "Expected interpolated string")
				assert.Len(t, interpolatedStr.Parts, 1)
				
				arrayAccess, ok := interpolatedStr.Parts[0].(*ast.ArrayAccessExpression)
				assert.True(t, ok, "Expected array access expression")
				
				variable, ok := arrayAccess.Array.(*ast.Variable)
				assert.True(t, ok, "Expected variable")
				assert.Equal(t, "$arr", variable.Name)
				
				indexVar, ok := (*arrayAccess.Index).(*ast.Variable)
				assert.True(t, ok, "Expected variable")
				assert.Equal(t, "$i", indexVar.Name)
			},
		},
		{
			name:  "array access with text prefix and suffix",
			input: `<?php echo "Value: $arr[0] found";`,
			validate: func(t *testing.T, prog *ast.Program) {
				assert.Len(t, prog.Body, 1)
				
				echoStmt, ok := prog.Body[0].(*ast.EchoStatement)
				assert.True(t, ok, "Expected echo statement")
				assert.Len(t, echoStmt.Arguments.Arguments, 1)
				
				interpolatedStr, ok := echoStmt.Arguments.Arguments[0].(*ast.InterpolatedStringExpression)
				assert.True(t, ok, "Expected interpolated string")
				assert.Len(t, interpolatedStr.Parts, 3)
				
				// Check first part (prefix)
				prefix, ok := interpolatedStr.Parts[0].(*ast.StringLiteral)
				assert.True(t, ok, "Expected string literal")
				assert.Equal(t, "Value: ", prefix.Value)
				
				// Check second part (array access)
				_, ok = interpolatedStr.Parts[1].(*ast.ArrayAccessExpression)
				assert.True(t, ok, "Expected array access expression")
				
				// Check third part (suffix)  
				suffix, ok := interpolatedStr.Parts[2].(*ast.StringLiteral)
				assert.True(t, ok, "Expected string literal")
				assert.Equal(t, " found", suffix.Value)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New(lexer.New(tt.input))
			prog := p.ParseProgram()
			
			assert.Empty(t, p.Errors(), "Parser should not have errors")
			assert.NotNil(t, prog, "Program should not be nil")
			
			tt.validate(t, prog)
		})
	}
}