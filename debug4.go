package main

import (
	"fmt"
	"github.com/yourname/php-parser/lexer"
)

// 我们通过直接检查来调试
func main() {
	input := `<?php echo "Hello"; ?>`
	lex := lexer.New(input)
	
	fmt.Printf("Input: %q\n", input)
	
	// 创建一个 lexer 并手动调用 nextTokenInitial
	// 但我们无法直接访问私有方法，所以让我们检查第一个 token
	fmt.Println("Getting first token...")
	tok1 := lex.NextToken()
	fmt.Printf("Token 1: Type=%s, Value=%q\n", lexer.TokenNames[tok1.Type], tok1.Value)
	
	if tok1.Type == lexer.T_INLINE_HTML {
		fmt.Println("ERROR: Got INLINE_HTML instead of OPEN_TAG!")
		fmt.Println("This means the <?php detection is not working")
		
		// 让我们看看这个值包含什么
		fmt.Printf("HTML Content: %q\n", tok1.Value)
	}
	
	tok2 := lex.NextToken()
	fmt.Printf("Token 2: Type=%s, Value=%q\n", lexer.TokenNames[tok2.Type], tok2.Value)
}