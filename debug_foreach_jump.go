package main

import (
	"fmt"

	"github.com/wudi/php-parser/compiler"
	"github.com/wudi/php-parser/lexer"
	"github.com/wudi/php-parser/parser"
)

func main() {
	code := `<?php $arr = [1,2,3]; foreach($arr as $value) { echo $value, "\n"; } ?>`

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
	
	fmt.Printf("\nJump instructions:\n")
	for i, inst := range bytecode {
		instStr := inst.String()
		if inst.Opcode.String() == "JMP" || inst.Opcode.String() == "JMPNZ" {
			fmt.Printf("[%04d] %s\n", i, instStr)
		}
	}
}