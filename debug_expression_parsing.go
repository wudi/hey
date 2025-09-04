package main

import (
	"encoding/json"
	"fmt"

	"github.com/wudi/php-parser/lexer"
	"github.com/wudi/php-parser/parser"
)

func main() {
	// Test what parseExpression returns for key => value
	code := `<?php $key => $value ?>`

	// Parse as a general expression
	p := parser.New(lexer.New(code))
	prog := p.ParseProgram()
	
	if len(p.Errors()) > 0 {
		fmt.Printf("Parse errors: %v\n", p.Errors())
		return
	}

	// Print the AST
	jsonBytes, err := json.MarshalIndent(prog, "", "  ")
	if err != nil {
		fmt.Printf("JSON marshal error: %v\n", err)
		return
	}
	
	fmt.Println("AST for '$key => $value':")
	fmt.Println(string(jsonBytes))
}