package main

import (
	"fmt"
	"github.com/yourname/php-parser/lexer"
)

func main() {
	input := `<?php echo "Hello, World!"; ?>`
	lex := lexer.New(input)
	
	fmt.Printf("Input: %q\n", input)
	fmt.Printf("Initial state: %s\n", lex.State().String())
	
	for i := 0; i < 10; i++ { // 限制循环次数以防无限循环
		tok := lex.NextToken()
		fmt.Printf("Token[%d]: %s\n", i, tok.String())
		
		if tok.Type == lexer.T_EOF {
			break
		}
	}
}