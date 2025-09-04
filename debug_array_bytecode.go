package main

import (
	"fmt"

	"github.com/wudi/php-parser/compiler"
	"github.com/wudi/php-parser/lexer"
	"github.com/wudi/php-parser/parser"
)

func main() {
	// Test simple array creation
	code := `<?php $arr = [1,2,3]; echo "Done"; ?>`

	// Parse
	p := parser.New(lexer.New(code))
	prog := p.ParseProgram()
	
	if len(p.Errors()) > 0 {
		fmt.Printf("Parse errors: %v\n", p.Errors())
		return
	}

	// Compile
	comp := compiler.NewCompiler()
	err := comp.Compile(prog)
	if err != nil {
		fmt.Printf("Compilation error: %v\n", err)
		return
	}

	// Show new bytecode
	bytecode := comp.GetBytecode()
	constants := comp.GetConstants()
	
	fmt.Printf("Generated %d bytecode instructions:\n", len(bytecode))
	for i, inst := range bytecode {
		fmt.Printf("[%04d] %s\n", i, inst.String())
	}
	
	fmt.Printf("\nConstants (%d):\n", len(constants))
	for i, c := range constants {
		fmt.Printf("[%04d] %s\n", i, c.String())
	}
}