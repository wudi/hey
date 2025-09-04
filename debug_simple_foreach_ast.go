package main

import (
	"encoding/json"
	"fmt"

	"github.com/wudi/php-parser/lexer"
	"github.com/wudi/php-parser/parser"
)

func main() {
	code := `<?php foreach($arr as $key => $value) { echo $key; } ?>`

	// Parse
	p := parser.New(lexer.New(code))
	prog := p.ParseProgram()
	
	if len(p.Errors()) > 0 {
		fmt.Printf("Parse errors: %v\n", p.Errors())
		return
	}

	// Print the full AST to find the foreach statement
	jsonBytes, err := json.MarshalIndent(prog, "", "  ")
	if err != nil {
		fmt.Printf("JSON marshal error: %v\n", err)
		return
	}
	
	fmt.Println("Full AST:")
	fmt.Println(string(jsonBytes))
}