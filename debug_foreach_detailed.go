package main

import (
	"fmt"

	"github.com/wudi/php-parser/compiler"
	"github.com/wudi/php-parser/compiler/vm"
	"github.com/wudi/php-parser/lexer"
	"github.com/wudi/php-parser/parser"
)

func main() {
	code := `<?php $arr = [1,2,3]; foreach($arr as $key => $value) { echo $key, ":", $value, "\n"; } ?>`

	// Parse
	p := parser.New(lexer.New(code))
	prog := p.ParseProgram()
	
	if len(p.Errors()) > 0 {
		fmt.Printf("Parse errors:\n")
		for _, err := range p.Errors() {
			fmt.Printf("  %s\n", err)
		}
		return
	}

	// Compile
	comp := compiler.NewCompiler()
	err := comp.Compile(prog)
	if err != nil {
		fmt.Printf("Compilation error: %v\n", err)
		return
	}

	// Show bytecode
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

	// Execute with debug mode
	fmt.Println("\n=== Executing with VM Debug Mode ===")
	vmCtx := vm.NewExecutionContext()
	virtualMachine := vm.NewVirtualMachine()
	virtualMachine.DebugMode = true  // Enable debug mode
	err = virtualMachine.Execute(vmCtx, bytecode, constants)
	if err != nil {
		fmt.Printf("Execution error: %v\n", err)
		return
	}
	
	fmt.Println("Execution completed successfully")
}