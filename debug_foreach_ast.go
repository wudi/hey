package main

import (
	"encoding/json"
	"fmt"

	"github.com/wudi/php-parser/lexer"
	"github.com/wudi/php-parser/parser"
)

func main() {
	code := `<?php $arr = [1,2,3]; foreach($arr as $key => $value) { echo $key, ":", $value, "\n"; } ?>`

	// Parse
	p := parser.New(lexer.New(code))
	prog := p.ParseProgram()
	
	if len(p.Errors()) > 0 {
		fmt.Printf("Parse errors: %v\n", p.Errors())
		return
	}

	// Convert to JSON to see the structure
	jsonBytes, err := json.MarshalIndent(prog, "", "  ")
	if err != nil {
		fmt.Printf("JSON marshal error: %v\n", err)
		return
	}
	
	fmt.Println("AST Structure:")
	fmt.Println(string(jsonBytes))
}