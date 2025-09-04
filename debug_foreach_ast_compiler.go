package main

import (
	"encoding/json"
	"fmt"

	"github.com/wudi/php-parser/ast"
	"github.com/wudi/php-parser/lexer"
	"github.com/wudi/php-parser/parser"
)

func main() {
	code := `<?php $arr = [1,2,3]; foreach($arr as $key => $value) { echo $key; } ?>`

	// Parse
	p := parser.New(lexer.New(code))
	prog := p.ParseProgram()
	
	if len(p.Errors()) > 0 {
		fmt.Printf("Parse errors: %v\n", p.Errors())
		return
	}

	// Find the foreach statement in the AST
	for _, stmt := range prog.Body {
		if _, ok := stmt.(*ast.ExpressionStatement); ok {
			// Skip assignments
		} else if blockStmt, ok := stmt.(*ast.BlockStatement); ok {
			for _, innerStmt := range blockStmt.Body {
				if foreachStmt, ok := innerStmt.(*ast.ForeachStatement); ok {
					fmt.Println("Found ForeachStatement:")
					fmt.Printf("  Key: %v\n", foreachStmt.Key)
					fmt.Printf("  Value: %v\n", foreachStmt.Value)
					
					if foreachStmt.Key != nil {
						fmt.Printf("  Key type: %T\n", foreachStmt.Key)
						if keyVar, ok := foreachStmt.Key.(*ast.Variable); ok {
							fmt.Printf("  Key variable name: %s\n", keyVar.Name)
						}
					} else {
						fmt.Printf("  Key is nil!\n")
					}
					
					if foreachStmt.Value != nil {
						fmt.Printf("  Value type: %T\n", foreachStmt.Value)
						if valueVar, ok := foreachStmt.Value.(*ast.Variable); ok {
							fmt.Printf("  Value variable name: %s\n", valueVar.Name)
						} else {
							fmt.Printf("  Value is not a simple Variable!\n")
							// Try to serialize it to see the structure
							valueBytes, _ := json.MarshalIndent(foreachStmt.Value, "    ", "  ")
							fmt.Printf("  Value structure: %s\n", string(valueBytes))
						}
					}
				}
			}
		}
	}
}