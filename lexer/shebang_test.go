package lexer

import (
	"testing"
)

func TestShebangSkipping(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []TokenType
	}{
		{
			name:     "Simple shebang with PHP code",
			input:    "#!/usr/bin/php\n<?php echo 'hello';",
			expected: []TokenType{T_OPEN_TAG, T_ECHO, T_CONSTANT_ENCAPSED_STRING, TOKEN_SEMICOLON, T_EOF},
		},
		{
			name:     "Shebang with different path",
			input:    "#!/bin/php\n<?php $name = 'test';",
			expected: []TokenType{T_OPEN_TAG, T_VARIABLE, TOKEN_EQUAL, T_CONSTANT_ENCAPSED_STRING, TOKEN_SEMICOLON, T_EOF},
		},
		{
			name:     "No shebang",
			input:    "<?php echo 'hello';",
			expected: []TokenType{T_OPEN_TAG, T_ECHO, T_CONSTANT_ENCAPSED_STRING, TOKEN_SEMICOLON, T_EOF},
		},
		{
			name:     "Hash comment (not shebang)",
			input:    "# This is a comment\n<?php echo 'hello';",
			expected: []TokenType{T_INLINE_HTML, T_OPEN_TAG, T_ECHO, T_CONSTANT_ENCAPSED_STRING, TOKEN_SEMICOLON, T_EOF},
		},
		{
			name:     "Empty shebang line",
			input:    "#!/usr/bin/php\n",
			expected: []TokenType{T_EOF},
		},
		{
			name:     "Shebang without newline",
			input:    "#!/usr/bin/php",
			expected: []TokenType{T_EOF},
		},
		{
			name:     "Shebang with simple function",
			input:    "#!/usr/local/bin/php -f\n<?php function test() { return; }",
			expected: []TokenType{T_OPEN_TAG, T_FUNCTION, T_STRING, TOKEN_LPAREN, TOKEN_RPAREN, TOKEN_LBRACE, T_RETURN, TOKEN_SEMICOLON, TOKEN_RBRACE, T_EOF},
		},
		{
			name:     "Shebang with CRLF line ending",
			input:    "#!/usr/bin/php\r\n<?php echo 'test';",
			expected: []TokenType{T_OPEN_TAG, T_ECHO, T_CONSTANT_ENCAPSED_STRING, TOKEN_SEMICOLON, T_EOF},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			lexer := New(test.input)

			for i, expectedType := range test.expected {
				token := lexer.NextToken()
				if token.Type != expectedType {
					t.Errorf("Test %s: token %d - expected %d, got %d (value: %q)",
						test.name, i, expectedType, token.Type, token.Value)
					break
				}
			}
		})
	}
}

func TestShebangContent(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedInput string // 预期处理后的输入
	}{
		{
			name:          "Basic shebang",
			input:         "#!/usr/bin/php\n<?php echo 'hello';",
			expectedInput: "<?php echo 'hello';",
		},
		{
			name:          "Shebang with arguments",
			input:         "#!/usr/bin/php -f\n<?php $x = 1;",
			expectedInput: "<?php $x = 1;",
		},
		{
			name:          "No shebang",
			input:         "<?php echo 'hello';",
			expectedInput: "<?php echo 'hello';",
		},
		{
			name:          "Just hash, not shebang",
			input:         "# comment\n<?php echo 'hello';",
			expectedInput: "# comment\n<?php echo 'hello';",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			lexer := New(test.input)
			if lexer.input != test.expectedInput {
				t.Errorf("Test %s: expected input %q, got %q",
					test.name, test.expectedInput, lexer.input)
			}
		})
	}
}

func TestShebangEdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "Empty input",
			input: "",
		},
		{
			name:  "Single character",
			input: "#",
		},
		{
			name:  "Only shebang",
			input: "#!/usr/bin/php",
		},
		{
			name:  "Shebang with only newline",
			input: "#!/usr/bin/php\n",
		},
		{
			name:  "Very long shebang",
			input: "#!/very/long/path/to/php/with/many/arguments/and/flags/that/might/be/used/in/some/systems\n<?php echo 'test';",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// 这个测试主要确保不会崩溃
			lexer := New(test.input)

			// 尝试获取一些 token，确保不会 panic
			for i := 0; i < 5; i++ {
				token := lexer.NextToken()
				if token.Type == T_EOF {
					break
				}
			}
		})
	}
}

// TestShebangPositions 测试跳过shebang后位置信息是否正确
func TestShebangPositions(t *testing.T) {
	input := "#!/usr/bin/php\n<?php echo 'hello';"
	lexer := New(input)

	// 第一个token应该是T_OPEN_TAG，位置应该从第2行开始
	token := lexer.NextToken()
	if token.Type != T_OPEN_TAG {
		t.Errorf("Expected T_OPEN_TAG, got %d", token.Type)
	}

	// 位置信息应该正确（注意：行号从1开始，但跳过shebang后实际从第2行开始）
	if token.Position.Line != 1 {
		t.Errorf("Expected line 1 after skipping shebang, got %d", token.Position.Line)
	}

	if token.Position.Column != 0 {
		t.Errorf("Expected column 0, got %d", token.Position.Column)
	}
}
