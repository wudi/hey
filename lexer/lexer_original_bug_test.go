package lexer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestLexer_OriginalBugCase tests the specific bug case that was reported:
// Array access in string interpolation should work the same as PHP
func TestLexer_OriginalBugCase(t *testing.T) {
	input := `<?php
$arr = [1,2,3];

for($i=0; $i<3; $i++) {
    echo "$arr[$i]\n";
}`

	expected := []struct {
		expectedType  TokenType
		expectedValue string
	}{
		// <?php
		{T_OPEN_TAG, "<?php\n"},
		
		// $arr = [1,2,3];
		{T_VARIABLE, "$arr"},
		{TOKEN_EQUAL, "="},
		{TOKEN_LBRACKET, "["},
		{T_LNUMBER, "1"},
		{TOKEN_COMMA, ","},
		{T_LNUMBER, "2"},
		{TOKEN_COMMA, ","},
		{T_LNUMBER, "3"},
		{TOKEN_RBRACKET, "]"},
		{TOKEN_SEMICOLON, ";"},
		
		// for($i=0; $i<3; $i++) {
		{T_FOR, "for"},
		{TOKEN_LPAREN, "("},
		{T_VARIABLE, "$i"},
		{TOKEN_EQUAL, "="},
		{T_LNUMBER, "0"},
		{TOKEN_SEMICOLON, ";"},
		{T_VARIABLE, "$i"},
		{TOKEN_LT, "<"},
		{T_LNUMBER, "3"},
		{TOKEN_SEMICOLON, ";"},
		{T_VARIABLE, "$i"},
		{T_INC, "++"},
		{TOKEN_RPAREN, ")"},
		{TOKEN_LBRACE, "{"},
		
		// echo "$arr[$i]\n";
		{T_ECHO, "echo"},
		{TOKEN_QUOTE, "\""},
		{T_VARIABLE, "$arr"},
		{TOKEN_LBRACKET, "["},
		{T_VARIABLE, "$i"},
		{TOKEN_RBRACKET, "]"},
		{T_ENCAPSED_AND_WHITESPACE, "\n"},
		{TOKEN_QUOTE, "\""},
		{TOKEN_SEMICOLON, ";"},
		
		// }
		{TOKEN_RBRACE, "}"},
		{T_EOF, ""},
	}

	lexer := New(input)

	for i, expected := range expected {
		tok := lexer.NextToken()
		assert.Equal(t, expected.expectedType, tok.Type,
			"test[%d] - token type wrong. expected=%s, got=%s",
			i, TokenNames[expected.expectedType], TokenNames[tok.Type])
		assert.Equal(t, expected.expectedValue, tok.Value,
			"test[%d] - value wrong. expected=%q, got=%q",
			i, expected.expectedValue, tok.Value)
	}
}