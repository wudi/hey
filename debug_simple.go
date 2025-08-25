package main

import (
	"fmt"
	"github.com/yourname/php-parser/lexer"
	"github.com/yourname/php-parser/parser"
	"github.com/yourname/php-parser/ast"
)

func main() {
	input := `<?php echo "Hello"; ?>`
	
	fmt.Printf("Input: %q\n", input)
	
	lex := lexer.New(input)
	p := parser.New(lex)
	program := p.ParseProgram()
	
	fmt.Printf("Errors: %v\n", p.Errors())
	fmt.Printf("Program statements: %d\n", len(program.Body))
	
	if len(program.Body) > 0 {
		stmt := program.Body[0]
		fmt.Printf("First statement type: %T\n", stmt)
		fmt.Printf("First statement string: %s\n", stmt.String())
		
		if echoStmt, ok := stmt.(*ast.EchoStatement); ok {
			fmt.Printf("Echo arguments: %d\n", len(echoStmt.Arguments))
			for i, arg := range echoStmt.Arguments {
				fmt.Printf("  Arg[%d]: %T = %s\n", i, arg, arg.String())
			}
		}
	}
}