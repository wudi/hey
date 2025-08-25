package main

import (
	"fmt"
	"github.com/yourname/php-parser/lexer"
	"github.com/yourname/php-parser/parser"
	"github.com/yourname/php-parser/ast"
)

func main() {
	input := `<?php $x = 1 + 2 * 3; ?>`
	
	fmt.Printf("Input: %q\n", input)
	
	lex := lexer.New(input)
	p := parser.New(lex)
	program := p.ParseProgram()
	
	fmt.Printf("Errors: %v\n", p.Errors())
	
	if len(program.Body) > 0 {
		stmt := program.Body[0]
		exprStmt := stmt.(*ast.ExpressionStatement)
		assignment := exprStmt.Expression.(*ast.AssignmentExpression)
		
		fmt.Printf("Assignment right side: %T\n", assignment.Right)
		fmt.Printf("Assignment right side string: %s\n", assignment.Right.String())
		
		// 期望的是: "((1 + (2 * 3)))"
		// 实际得到: "(1 + (2 * 3))"
		
		if binaryExpr, ok := assignment.Right.(*ast.BinaryExpression); ok {
			fmt.Printf("Binary operator: %s\n", binaryExpr.Operator)
			fmt.Printf("Left: %s\n", binaryExpr.Left.String())
			fmt.Printf("Right: %s\n", binaryExpr.Right.String())
		}
	}
}