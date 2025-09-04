package main

import (
	"fmt"

	"github.com/wudi/php-parser/compiler"
	"github.com/wudi/php-parser/compiler/vm"
	"github.com/wudi/php-parser/lexer"
	"github.com/wudi/php-parser/parser"
)

func main() {
	// Minimal test case to debug FE_RESET and FE_FETCH
	code := `<?php $arr = [42]; foreach($arr as $value) { echo $value; } ?>`

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
	
	fmt.Printf("=== MINIMAL FOREACH DEBUG ===\n")
	fmt.Printf("Array should contain: [42]\n")
	fmt.Printf("Expected output: 42\n\n")
	
	// Show key instructions
	fmt.Printf("Key foreach instructions:\n")
	for i, inst := range bytecode {
		opStr := inst.Opcode.String()
		if opStr == "FE_RESET" || opStr == "FE_FETCH" || opStr == "ASSIGN" || (opStr == "FETCH_R" && i > 10) || opStr == "ECHO" {
			fmt.Printf("[%04d] %s\n", i, inst.String())
		}
	}
	
	fmt.Printf("\nConstants:\n")
	for i, c := range constants {
		fmt.Printf("[%04d] %s\n", i, c.String())
	}

	fmt.Printf("\nActual execution output: ")
	
	// Execute
	vmCtx := vm.NewExecutionContext()
	virtualMachine := vm.NewVirtualMachine()
	err = virtualMachine.Execute(vmCtx, comp.GetBytecode(), comp.GetConstants())
	if err != nil {
		fmt.Printf("Execution error: %v\n", err)
		return
	}
	
	fmt.Printf("\n--- End Debug ---\n")
}