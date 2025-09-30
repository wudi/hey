package main

import (
	"fmt"
	"github.com/wudi/hey/compiler/ast"
	"github.com/wudi/hey/compiler/lexer"
	"github.com/wudi/hey/compiler/parser"
)

func main() {
	input := `<?php
if (true) {
	?>HTML inside if
<?php
}
?>`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		fmt.Println("Parser errors:")
		for _, err := range p.Errors() {
			fmt.Println("  ", err)
		}
		return
	}

	// Check first statement
	if len(program.Body) > 0 {
		stmt := program.Body[0]
		fmt.Printf("Statement type: %T\n", stmt)
		if ifStmt, ok := stmt.(*ast.IfStatement); ok {
			fmt.Printf("If statement consequent has %d statements\n", len(ifStmt.Consequent))
			if len(ifStmt.Consequent) > 0 {
				conseq := ifStmt.Consequent[0]
				fmt.Printf("Consequent[0] type: %T\n", conseq)
				if exprStmt, ok := conseq.(*ast.ExpressionStatement); ok {
					fmt.Printf("Expression type: %T\n", exprStmt.Expression)
				}
			}
		}
	}
}