package main

import (
	"fmt"
	"github.com/yourname/php-parser/lexer"
)

func main() {
	input := `<?php 
// This is a single line comment
/* This is a 
   block comment */
/** This is a doc comment */
# Hash comment
echo "Hello";
?>`
	
	lex := lexer.New(input)
	
	fmt.Printf("Input: %q\n", input)
	
	for i := 0; i < 20; i++ { // 限制循环次数
		tok := lex.NextToken()
		fmt.Printf("Token[%d]: %s = %q\n", i, lexer.TokenNames[tok.Type], tok.Value)
		
		if tok.Type == lexer.T_EOF {
			break
		}
	}
}