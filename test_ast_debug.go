package main

import (
	"encoding/json"
	"fmt"
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

	// Print AST as JSON
	data, err := json.MarshalIndent(program, "", "  ")
	if err != nil {
		fmt.Println("JSON error:", err)
		return
	}
	fmt.Println(string(data))
}