package main

import (
	"fmt"
	"github.com/yourname/php-parser/lexer"
)

func main() {
	input := `<?php
$name = "John";
$age = 25;`
	
	lex := lexer.New(input)
	
	fmt.Printf("Input: %q\n", input)
	fmt.Printf("Lines in input:\n")
	
	lines := 1
	for i, ch := range input {
		fmt.Printf("  input[%d] = '%c'", i, ch)
		if ch == '\n' {
			lines++
			fmt.Printf("  (LINE %d STARTS HERE)", lines)
		}
		fmt.Println()
		if i >= 20 { // 只显示前面一些字符
			break
		}
	}
	
	fmt.Println("\nTokens:")
	for i := 0; i < 10; i++ {
		tok := lex.NextToken()
		fmt.Printf("Token[%d]: %s = %q at Line:%d Column:%d\n", 
			i, lexer.TokenNames[tok.Type], tok.Value, tok.Position.Line, tok.Position.Column)
		
		if tok.Type == lexer.T_EOF {
			break
		}
	}
}