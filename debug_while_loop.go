package main

import (
	"fmt"

	"github.com/wudi/php-parser/compiler"
	"github.com/wudi/php-parser/lexer"
	"github.com/wudi/php-parser/parser"
)

func main() {
	// Test simple while loop to see if jump/label system works
	code := `<?php $i = 0; while($i < 2) { echo $i, "\n"; $i++; } ?>`

	// Parse and compile
	p := parser.New(lexer.New(code))
	prog := p.ParseProgram()
	
	if len(p.Errors()) > 0 {
		fmt.Printf("Parse errors: %v\n", p.Errors())
		return
	}

	comp := compiler.NewCompiler()
	err := comp.Compile(prog)
	if err != nil {
		fmt.Printf("Compilation error: %v\n", err)
		return
	}

	bytecode := comp.GetBytecode()
	constants := comp.GetConstants()
	
	fmt.Printf("Constants:\n")
	for i, c := range constants {
		fmt.Printf("[%04d] %s\n", i, c.String())
	}
	
	fmt.Printf("\nBytecode:\n")
	for i, inst := range bytecode {
		fmt.Printf("[%04d] %s\n", i, inst.String())
	}
	
	fmt.Printf("\nJump instructions:\n")
	for i, inst := range bytecode {
		if inst.Opcode.String() == "JMP" || inst.Opcode.String() == "JMPZ" || inst.Opcode.String() == "JMPNZ" {
			fmt.Printf("[%04d] %s\n", i, inst.String())
		}
	}
}