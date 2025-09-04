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
		{
			name:  "invalid expression in array index - graceful fallback",
			input: `<?php echo "$arr[$i+1]";`,
			validate: func(t *testing.T, prog *ast.Program) {
				assert.Len(t, prog.Body, 1)
				
				echoStmt, ok := prog.Body[0].(*ast.EchoStatement)
				assert.True(t, ok, "Expected echo statement")
				assert.Len(t, echoStmt.Arguments.Arguments, 1)
				
				interpolatedStr, ok := echoStmt.Arguments.Arguments[0].(*ast.InterpolatedStringExpression)
				assert.True(t, ok, "Expected interpolated string")
				
				// This is a edge case: invalid syntax gets parsed with graceful fallback
				// The exact parsing behavior may vary, but should not crash
				assert.GreaterOrEqual(t, len(interpolatedStr.Parts), 1, "Should have at least one part")
				
				// First part should be just the variable (not array access)
				variable, ok := interpolatedStr.Parts[0].(*ast.Variable)
				assert.True(t, ok, "Expected variable (not array access)")
				assert.Equal(t, "$arr", variable.Name)
				
				// Additional parts may contain the invalid syntax as literals
			},
		},
		{
			name:  "multiple array access in same string",
			input: `<?php echo "$a[0] and $b[1]";`,
			validate: func(t *testing.T, prog *ast.Program) {
				assert.Len(t, prog.Body, 1)
				
				echoStmt, ok := prog.Body[0].(*ast.EchoStatement)
				assert.True(t, ok, "Expected echo statement")
				
				interpolatedStr, ok := echoStmt.Arguments.Arguments[0].(*ast.InterpolatedStringExpression)
				assert.True(t, ok, "Expected interpolated string")
				
				// Should have parts: $a[0], " and ", $b[1]
				assert.Len(t, interpolatedStr.Parts, 3)
				
				// First array access
				arrayAccess1, ok := interpolatedStr.Parts[0].(*ast.ArrayAccessExpression)
				assert.True(t, ok, "Expected first array access")
				variable1, _ := arrayAccess1.Array.(*ast.Variable)
				assert.Equal(t, "$a", variable1.Name)
				
				// Middle text
				text, ok := interpolatedStr.Parts[1].(*ast.StringLiteral)
				assert.True(t, ok, "Expected string literal")
				assert.Equal(t, " and ", text.Value)
				
				// Second array access
				arrayAccess2, ok := interpolatedStr.Parts[2].(*ast.ArrayAccessExpression)
				assert.True(t, ok, "Expected second array access")
				variable2, _ := arrayAccess2.Array.(*ast.Variable)
				assert.Equal(t, "$b", variable2.Name)
			},
		},
		{
			name:  "string key array access",
			input: `<?php echo "$arr[key]";`,
			validate: func(t *testing.T, prog *ast.Program) {
				assert.Len(t, prog.Body, 1)
				
				echoStmt, ok := prog.Body[0].(*ast.EchoStatement)
				assert.True(t, ok, "Expected echo statement")
				
				interpolatedStr, ok := echoStmt.Arguments.Arguments[0].(*ast.InterpolatedStringExpression)
				assert.True(t, ok, "Expected interpolated string")
				assert.Len(t, interpolatedStr.Parts, 1)
				
				arrayAccess, ok := interpolatedStr.Parts[0].(*ast.ArrayAccessExpression)
				assert.True(t, ok, "Expected array access expression")
				
				variable, ok := arrayAccess.Array.(*ast.Variable)
				assert.True(t, ok, "Expected variable")
				assert.Equal(t, "$arr", variable.Name)
				
				// String key should be parsed as string literal (T_STRING becomes StringLiteral)
				stringLit, ok := (*arrayAccess.Index).(*ast.StringLiteral)
				assert.True(t, ok, "Expected string literal for string key")
				assert.Equal(t, "key", stringLit.Value)
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

// TestLexer_VarOffsetStateHandling tests that the lexer correctly handles
// the VAR_OFFSET state for array access in interpolated strings
func TestLexer_VarOffsetStateHandling(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []lexer.TokenType
		values   []string
	}{
		{
			name:  "valid array access with numeric index",
			input: `"$arr[0]"`,
			expected: []lexer.TokenType{
				lexer.TOKEN_QUOTE,
				lexer.T_VARIABLE,
				lexer.TOKEN_LBRACKET,
				lexer.T_LNUMBER,
				lexer.TOKEN_RBRACKET,
				lexer.TOKEN_QUOTE,
			},
			values: []string{`"`, "$arr", "[", "0", "]", `"`},
		},
		{
			name:  "valid array access with variable index",
			input: `"$arr[$i]"`,
			expected: []lexer.TokenType{
				lexer.TOKEN_QUOTE,
				lexer.T_VARIABLE,
				lexer.TOKEN_LBRACKET,
				lexer.T_VARIABLE,
				lexer.TOKEN_RBRACKET,
				lexer.TOKEN_QUOTE,
			},
			values: []string{`"`, "$arr", "[", "$i", "]", `"`},
		},
		{
			name:  "invalid expression in array index",
			input: `"$arr[$i+1]"`,
			expected: []lexer.TokenType{
				lexer.TOKEN_QUOTE,
				lexer.T_VARIABLE,
				lexer.TOKEN_LBRACKET,
				lexer.T_VARIABLE,
				lexer.T_ENCAPSED_AND_WHITESPACE, // "+" exits VAR_OFFSET state
				lexer.T_ENCAPSED_AND_WHITESPACE, // "1]" as literal text
				lexer.TOKEN_QUOTE,
			},
			values: []string{`"`, "$arr", "[", "$i", "+", "1]", `"`},
		},
		{
			name:  "array access with string key",
			input: `"$arr[key]"`,
			expected: []lexer.TokenType{
				lexer.TOKEN_QUOTE,
				lexer.T_VARIABLE,
				lexer.TOKEN_LBRACKET,
				lexer.T_STRING,
				lexer.TOKEN_RBRACKET,
				lexer.TOKEN_QUOTE,
			},
			values: []string{`"`, "$arr", "[", "key", "]", `"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New("<?php echo " + tt.input + ";")
			
			// Skip opening tokens (T_OPEN_TAG, T_ECHO)
			l.NextToken() // T_OPEN_TAG
			l.NextToken() // T_ECHO
			
			// Test the interpolated string tokens
			for i, expectedType := range tt.expected {
				token := l.NextToken()
				assert.Equal(t, expectedType, token.Type, 
					"Token %d: expected %s, got %s", i, expectedType.String(), token.Type.String())
				assert.Equal(t, tt.values[i], token.Value,
					"Token %d value: expected %q, got %q", i, tt.values[i], token.Value)
			}
		})
	}
}