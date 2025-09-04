package main

import (
	"fmt"

	"github.com/wudi/php-parser/compiler"
	"github.com/wudi/php-parser/lexer"
	"github.com/wudi/php-parser/parser"
)

func main() {
	// Test simple foreach to debug label placement
	code := `<?php $arr = [1]; foreach($arr as $value) { echo $value; } ?>`

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
	
	fmt.Printf("Foreach bytecode analysis:\n")
	for i, inst := range bytecode {
		instStr := inst.String()
		fmt.Printf("[%04d] %s", i, instStr)
		
		// Highlight foreach-related instructions
		opStr := inst.Opcode.String()
		if opStr == "FE_RESET" || opStr == "FE_FETCH" {
			fmt.Printf("  <-- FOREACH")
		} else if opStr == "JMP" || opStr == "JMPNZ" {
			fmt.Printf("  <-- JUMP")
		}
		fmt.Println()
	}
	
	fmt.Printf("\nConstants:\n")
	for i, c := range constants {
		fmt.Printf("[%04d] %s\n", i, c.String())
	}
}