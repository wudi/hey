package main

import (
	"fmt"
	"github.com/yourname/php-parser/lexer"
	"github.com/yourname/php-parser/parser"
)

func main() {
	input := `<?php $name = "John"; ?>`
	
	fmt.Printf("Input: %q\n", input)
	
	// 首先看看 lexer 输出什么 tokens
	fmt.Println("\n=== Tokens ===")
	lex := lexer.New(input)
	for i := 0; i < 10; i++ {
		tok := lex.NextToken()
		fmt.Printf("Token[%d]: %s = %q\n", i, lexer.TokenNames[tok.Type], tok.Value)
		if tok.Type == lexer.T_EOF {
			break
		}
	}
	
	// 然后尝试解析
	fmt.Println("\n=== Parsing ===")
	lex2 := lexer.New(input)
	p := parser.New(lex2)
	
	fmt.Printf("Current token: %s = %q\n", lexer.TokenNames[p.Current().Type], p.Current().Value)
	fmt.Printf("Peek token: %s = %q\n", lexer.TokenNames[p.Peek().Type], p.Peek().Value)
	
	program := p.ParseProgram()
	
	fmt.Printf("Errors: %v\n", p.Errors())
	fmt.Printf("Program statements: %d\n", len(program.Body))
	
	if len(program.Body) > 0 {
		stmt := program.Body[0]
		fmt.Printf("Statement type: %T\n", stmt)
		fmt.Printf("Statement: %s\n", stmt.String())
	}
}