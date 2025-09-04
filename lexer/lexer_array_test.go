package lexer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestLexer_StringInterpolationArrayAccess tests that array access in double-quoted strings
// is tokenized correctly according to PHP's behavior
func TestLexer_StringInterpolationArrayAccess(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []struct {
			expectedType  TokenType
			expectedValue string
		}
	}{
		{
			name:  "simple array access in double quotes",
			input: `<?php echo "$arr[0]";`,
			expected: []struct {
				expectedType  TokenType
				expectedValue string
			}{
				{T_OPEN_TAG, "<?php "},
				{T_ECHO, "echo"},
				{TOKEN_QUOTE, "\""},
				{T_VARIABLE, "$arr"},
				{TOKEN_LBRACKET, "["},
				{T_LNUMBER, "0"},
				{TOKEN_RBRACKET, "]"},
				{TOKEN_QUOTE, "\""},
				{TOKEN_SEMICOLON, ";"},
				{T_EOF, ""},
			},
		},
		{
			name:  "variable index array access in double quotes",
			input: `<?php echo "$arr[$i]";`,
			expected: []struct {
				expectedType  TokenType
				expectedValue string
			}{
				{T_OPEN_TAG, "<?php "},
				{T_ECHO, "echo"},
				{TOKEN_QUOTE, "\""},
				{T_VARIABLE, "$arr"},
				{TOKEN_LBRACKET, "["},
				{T_VARIABLE, "$i"},
				{TOKEN_RBRACKET, "]"},
				{TOKEN_QUOTE, "\""},
				{TOKEN_SEMICOLON, ";"},
				{T_EOF, ""},
			},
		},
		{
			name:  "array access with text prefix and suffix",
			input: `<?php echo "Value: $arr[$key] found";`,
			expected: []struct {
				expectedType  TokenType
				expectedValue string
			}{
				{T_OPEN_TAG, "<?php "},
				{T_ECHO, "echo"},
				{TOKEN_QUOTE, "\""},
				{T_ENCAPSED_AND_WHITESPACE, "Value: "},
				{T_VARIABLE, "$arr"},
				{TOKEN_LBRACKET, "["},
				{T_VARIABLE, "$key"},
				{TOKEN_RBRACKET, "]"},
				{T_ENCAPSED_AND_WHITESPACE, " found"},
				{TOKEN_QUOTE, "\""},
				{TOKEN_SEMICOLON, ";"},
				{T_EOF, ""},
			},
		},
		{
			name:  "multiple array accesses in one string",
			input: `<?php echo "$arr[0] and $brr[1]";`,
			expected: []struct {
				expectedType  TokenType
				expectedValue string
			}{
				{T_OPEN_TAG, "<?php "},
				{T_ECHO, "echo"},
				{TOKEN_QUOTE, "\""},
				{T_VARIABLE, "$arr"},
				{TOKEN_LBRACKET, "["},
				{T_LNUMBER, "0"},
				{TOKEN_RBRACKET, "]"},
				{T_ENCAPSED_AND_WHITESPACE, " and "},
				{T_VARIABLE, "$brr"},
				{TOKEN_LBRACKET, "["},
				{T_LNUMBER, "1"},
				{TOKEN_RBRACKET, "]"},
				{TOKEN_QUOTE, "\""},
				{TOKEN_SEMICOLON, ";"},
				{T_EOF, ""},
			},
		},
		{
			name:  "array access with string index",
			input: `<?php echo "$arr[key]";`,
			expected: []struct {
				expectedType  TokenType
				expectedValue string
			}{
				{T_OPEN_TAG, "<?php "},
				{T_ECHO, "echo"},
				{TOKEN_QUOTE, "\""},
				{T_VARIABLE, "$arr"},
				{TOKEN_LBRACKET, "["},
				{T_STRING, "key"},
				{TOKEN_RBRACKET, "]"},
				{TOKEN_QUOTE, "\""},
				{TOKEN_SEMICOLON, ";"},
				{T_EOF, ""},
			},
		},
		{
			name:  "array access with newline",
			input: `<?php echo "$arr[$i]\n";`,
			expected: []struct {
				expectedType  TokenType
				expectedValue string
			}{
				{T_OPEN_TAG, "<?php "},
				{T_ECHO, "echo"},
				{TOKEN_QUOTE, "\""},
				{T_VARIABLE, "$arr"},
				{TOKEN_LBRACKET, "["},
				{T_VARIABLE, "$i"},
				{TOKEN_RBRACKET, "]"},
				{T_ENCAPSED_AND_WHITESPACE, "\n"},
				{TOKEN_QUOTE, "\""},
				{TOKEN_SEMICOLON, ";"},
				{T_EOF, ""},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := New(tt.input)

			for i, expected := range tt.expected {
				tok := lexer.NextToken()
				assert.Equal(t, expected.expectedType, tok.Type,
					"test[%d] - token type wrong. expected=%s, got=%s",
					i, TokenNames[expected.expectedType], TokenNames[tok.Type])
				assert.Equal(t, expected.expectedValue, tok.Value,
					"test[%d] - value wrong. expected=%q, got=%q",
					i, expected.expectedValue, tok.Value)
			}
		})
	}
}

// TestLexer_StringInterpolationEdgeCases tests edge cases for string interpolation
func TestLexer_StringInterpolationEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []struct {
			expectedType  TokenType
			expectedValue string
		}
	}{
		{
			name:  "variable without array access should not enter ST_VAR_OFFSET",
			input: `<?php echo "$arr and more text";`,
			expected: []struct {
				expectedType  TokenType
				expectedValue string
			}{
				{T_OPEN_TAG, "<?php "},
				{T_ECHO, "echo"},
				{TOKEN_QUOTE, "\""},
				{T_VARIABLE, "$arr"},
				{T_ENCAPSED_AND_WHITESPACE, " and more text"},
				{TOKEN_QUOTE, "\""},
				{TOKEN_SEMICOLON, ";"},
				{T_EOF, ""},
			},
		},
		{
			name:  "escaped bracket should not trigger array access",
			input: `<?php echo "$arr\\[0\\]";`,
			expected: []struct {
				expectedType  TokenType
				expectedValue string
			}{
				{T_OPEN_TAG, "<?php "},
				{T_ECHO, "echo"},
				{TOKEN_QUOTE, "\""},
				{T_VARIABLE, "$arr"},
				{T_ENCAPSED_AND_WHITESPACE, `\[0\]`},
				{TOKEN_QUOTE, "\""},
				{TOKEN_SEMICOLON, ";"},
				{T_EOF, ""},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := New(tt.input)

			for i, expected := range tt.expected {
				tok := lexer.NextToken()
				assert.Equal(t, expected.expectedType, tok.Type,
					"test[%d] - token type wrong. expected=%s, got=%s",
					i, TokenNames[expected.expectedType], TokenNames[tok.Type])
				assert.Equal(t, expected.expectedValue, tok.Value,
					"test[%d] - value wrong. expected=%q, got=%q",
					i, expected.expectedValue, tok.Value)
			}
		})
	}
}