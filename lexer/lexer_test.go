package lexer

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLexer_BasicTokens(t *testing.T) {
	input := `<?php echo "Hello, World!"; ?>`
	
	tests := []struct {
		expectedType  TokenType
		expectedValue string
	}{
		{T_OPEN_TAG, "<?php "},
		{T_ECHO, "echo"},
		{T_CONSTANT_ENCAPSED_STRING, `"Hello, World!"`},
		{TOKEN_SEMICOLON, ";"},
		{T_CLOSE_TAG, "?>"},
		{T_EOF, ""},
	}
	
	lexer := New(input)
	
	for i, tt := range tests {
		tok := lexer.NextToken()
		assert.Equal(t, tt.expectedType, tok.Type, "test[%d] - tokentype wrong. expected=%q, got=%q", i, TokenNames[tt.expectedType], TokenNames[tok.Type])
		assert.Equal(t, tt.expectedValue, tok.Value, "test[%d] - value wrong. expected=%q, got=%q", i, tt.expectedValue, tok.Value)
	}
}

func TestLexer_Variables(t *testing.T) {
	input := `<?php $name = "John"; $age = 25; ?>`
	
	tests := []struct {
		expectedType  TokenType
		expectedValue string
	}{
		{T_OPEN_TAG, "<?php "},
		{T_VARIABLE, "$name"},
		{TOKEN_EQUAL, "="},
		{T_CONSTANT_ENCAPSED_STRING, `"John"`},
		{TOKEN_SEMICOLON, ";"},
		{T_VARIABLE, "$age"},
		{TOKEN_EQUAL, "="},
		{T_LNUMBER, "25"},
		{TOKEN_SEMICOLON, ";"},
		{T_CLOSE_TAG, "?>"},
		{T_EOF, ""},
	}
	
	lexer := New(input)
	
	for i, tt := range tests {
		tok := lexer.NextToken()
		assert.Equal(t, tt.expectedType, tok.Type, "test[%d] - tokentype wrong. expected=%q, got=%q", i, TokenNames[tt.expectedType], TokenNames[tok.Type])
		assert.Equal(t, tt.expectedValue, tok.Value, "test[%d] - value wrong. expected=%q, got=%q", i, tt.expectedValue, tok.Value)
	}
}

func TestLexer_Operators(t *testing.T) {
	input := `<?php $a + $b - $c * $d / $e % $f; ?>`
	
	tests := []struct {
		expectedType  TokenType
		expectedValue string
	}{
		{T_OPEN_TAG, "<?php "},
		{T_VARIABLE, "$a"},
		{TOKEN_PLUS, "+"},
		{T_VARIABLE, "$b"},
		{TOKEN_MINUS, "-"},
		{T_VARIABLE, "$c"},
		{TOKEN_MULTIPLY, "*"},
		{T_VARIABLE, "$d"},
		{TOKEN_DIVIDE, "/"},
		{T_VARIABLE, "$e"},
		{TOKEN_MODULO, "%"},
		{T_VARIABLE, "$f"},
		{TOKEN_SEMICOLON, ";"},
		{T_CLOSE_TAG, "?>"},
		{T_EOF, ""},
	}
	
	lexer := New(input)
	
	for i, tt := range tests {
		tok := lexer.NextToken()
		assert.Equal(t, tt.expectedType, tok.Type, "test[%d] - tokentype wrong. expected=%q, got=%q", i, TokenNames[tt.expectedType], TokenNames[tok.Type])
		assert.Equal(t, tt.expectedValue, tok.Value, "test[%d] - value wrong. expected=%q, got=%q", i, tt.expectedValue, tok.Value)
	}
}

func TestLexer_ComparisonOperators(t *testing.T) {
	input := `<?php $a == $b != $c === $d !== $e <= $f >= $g <=> $h; ?>`
	
	tests := []struct {
		expectedType  TokenType
		expectedValue string
	}{
		{T_OPEN_TAG, "<?php "},
		{T_VARIABLE, "$a"},
		{T_IS_EQUAL, "=="},
		{T_VARIABLE, "$b"},
		{T_IS_NOT_EQUAL, "!="},
		{T_VARIABLE, "$c"},
		{T_IS_IDENTICAL, "==="},
		{T_VARIABLE, "$d"},
		{T_IS_NOT_IDENTICAL, "!=="},
		{T_VARIABLE, "$e"},
		{T_IS_SMALLER_OR_EQUAL, "<="},
		{T_VARIABLE, "$f"},
		{T_IS_GREATER_OR_EQUAL, ">="},
		{T_VARIABLE, "$g"},
		{T_SPACESHIP, "<=>"},
		{T_VARIABLE, "$h"},
		{TOKEN_SEMICOLON, ";"},
		{T_CLOSE_TAG, "?>"},
		{T_EOF, ""},
	}
	
	lexer := New(input)
	
	for i, tt := range tests {
		tok := lexer.NextToken()
		assert.Equal(t, tt.expectedType, tok.Type, "test[%d] - tokentype wrong. expected=%q, got=%q", i, TokenNames[tt.expectedType], TokenNames[tok.Type])
		assert.Equal(t, tt.expectedValue, tok.Value, "test[%d] - value wrong. expected=%q, got=%q", i, tt.expectedValue, tok.Value)
	}
}

func TestLexer_AssignmentOperators(t *testing.T) {
	input := `<?php $a += $b -= $c *= $d /= $e .= $f; ?>`
	
	tests := []struct {
		expectedType  TokenType
		expectedValue string
	}{
		{T_OPEN_TAG, "<?php "},
		{T_VARIABLE, "$a"},
		{T_PLUS_EQUAL, "+="},
		{T_VARIABLE, "$b"},
		{T_MINUS_EQUAL, "-="},
		{T_VARIABLE, "$c"},
		{T_MUL_EQUAL, "*="},
		{T_VARIABLE, "$d"},
		{T_DIV_EQUAL, "/="},
		{T_VARIABLE, "$e"},
		{T_CONCAT_EQUAL, ".="},
		{T_VARIABLE, "$f"},
		{TOKEN_SEMICOLON, ";"},
		{T_CLOSE_TAG, "?>"},
		{T_EOF, ""},
	}
	
	lexer := New(input)
	
	for i, tt := range tests {
		tok := lexer.NextToken()
		assert.Equal(t, tt.expectedType, tok.Type, "test[%d] - tokentype wrong. expected=%q, got=%q", i, TokenNames[tt.expectedType], TokenNames[tok.Type])
		assert.Equal(t, tt.expectedValue, tok.Value, "test[%d] - value wrong. expected=%q, got=%q", i, tt.expectedValue, tok.Value)
	}
}

func TestLexer_Keywords(t *testing.T) {
	input := `<?php if ($condition) { echo "true"; } else { echo "false"; } ?>`
	
	tests := []struct {
		expectedType  TokenType
		expectedValue string
	}{
		{T_OPEN_TAG, "<?php "},
		{T_IF, "if"},
		{TOKEN_LPAREN, "("},
		{T_VARIABLE, "$condition"},
		{TOKEN_RPAREN, ")"},
		{TOKEN_LBRACE, "{"},
		{T_ECHO, "echo"},
		{T_CONSTANT_ENCAPSED_STRING, `"true"`},
		{TOKEN_SEMICOLON, ";"},
		{TOKEN_RBRACE, "}"},
		{T_ELSE, "else"},
		{TOKEN_LBRACE, "{"},
		{T_ECHO, "echo"},
		{T_CONSTANT_ENCAPSED_STRING, `"false"`},
		{TOKEN_SEMICOLON, ";"},
		{TOKEN_RBRACE, "}"},
		{T_CLOSE_TAG, "?>"},
		{T_EOF, ""},
	}
	
	lexer := New(input)
	
	for i, tt := range tests {
		tok := lexer.NextToken()
		assert.Equal(t, tt.expectedType, tok.Type, "test[%d] - tokentype wrong. expected=%q, got=%q", i, TokenNames[tt.expectedType], TokenNames[tok.Type])
		assert.Equal(t, tt.expectedValue, tok.Value, "test[%d] - value wrong. expected=%q, got=%q", i, tt.expectedValue, tok.Value)
	}
}

func TestLexer_Numbers(t *testing.T) {
	tests := []struct {
		input         string
		expectedType  TokenType
		expectedValue string
	}{
		{"123", T_LNUMBER, "123"},
		{"0", T_LNUMBER, "0"},
		{"0x1F", T_LNUMBER, "0x1F"},
		{"0X1f", T_LNUMBER, "0X1f"},
		{"0123", T_LNUMBER, "0123"},
		{"0b1010", T_LNUMBER, "0b1010"},
		{"0B1010", T_LNUMBER, "0B1010"},
		{"3.14", T_DNUMBER, "3.14"},
		{"2.5e2", T_DNUMBER, "2.5e2"},
		{"1E-3", T_DNUMBER, "1E-3"},
		{".5", T_DNUMBER, ".5"},
	}
	
	for _, tt := range tests {
		lexer := New("<?php " + tt.input + " ?>")
		lexer.NextToken() // Skip T_OPEN_TAG
		
		tok := lexer.NextToken()
		assert.Equal(t, tt.expectedType, tok.Type, "input=%q - tokentype wrong. expected=%q, got=%q", tt.input, TokenNames[tt.expectedType], TokenNames[tok.Type])
		assert.Equal(t, tt.expectedValue, tok.Value, "input=%q - value wrong. expected=%q, got=%q", tt.input, tt.expectedValue, tok.Value)
	}
}

func TestLexer_Comments(t *testing.T) {
	input := `<?php 
// This is a single line comment
/* This is a 
   block comment */
/** This is a doc comment */
# Hash comment
echo "Hello";
?>`
	
	lexer := New(input)
	
	// 跳过开始标签
	tok := lexer.NextToken()
	assert.Equal(t, T_OPEN_TAG, tok.Type)
	
	// 第一个注释 //
	tok = lexer.NextToken()
	assert.Equal(t, T_COMMENT, tok.Type)
	assert.True(t, strings.HasPrefix(tok.Value, "// This is a single line comment"))
	
	// 块注释 /* */
	tok = lexer.NextToken()
	assert.Equal(t, T_COMMENT, tok.Type)
	assert.Contains(t, tok.Value, "This is a")
	assert.Contains(t, tok.Value, "block comment")
	
	// 文档注释 /** */
	tok = lexer.NextToken()
	assert.Equal(t, T_DOC_COMMENT, tok.Type)
	assert.Contains(t, tok.Value, "This is a doc comment")
	
	// Hash 注释 #
	tok = lexer.NextToken()
	assert.Equal(t, T_COMMENT, tok.Type)
	assert.True(t, strings.HasPrefix(tok.Value, "# Hash comment"))
}

func TestLexer_Position(t *testing.T) {
	input := `<?php
$name = "John";
$age = 25;`
	
	lexer := New(input)
	
	// 检查位置信息是否正确
	tok := lexer.NextToken() // <?php
	assert.Equal(t, 1, tok.Position.Line)
	assert.Equal(t, 0, tok.Position.Column)
	
	tok = lexer.NextToken() // $name
	assert.Equal(t, 2, tok.Position.Line)
	assert.Equal(t, 0, tok.Position.Column)
	
	tok = lexer.NextToken() // =
	assert.Equal(t, 2, tok.Position.Line)
	
	tok = lexer.NextToken() // "John"
	assert.Equal(t, 2, tok.Position.Line)
	
	tok = lexer.NextToken() // ;
	assert.Equal(t, 2, tok.Position.Line)
	
	tok = lexer.NextToken() // $age
	assert.Equal(t, 3, tok.Position.Line)
	assert.Equal(t, 0, tok.Position.Column)
}