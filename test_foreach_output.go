package main

import (
	"fmt"
	"os"

	"github.com/wudi/php-parser/compiler"
	"github.com/wudi/php-parser/compiler/vm"
	"github.com/wudi/php-parser/lexer"
	"github.com/wudi/php-parser/parser"
)

func main() {
	code := `<?php
$arr = [1,2,3];

foreach($arr as $key => $value) {
    echo $key, ":", $value, "\n";
}`

	// Parse the PHP code
	p := parser.New(lexer.New(code))
	prog := p.ParseProgram()
	
	if len(p.Errors()) > 0 {
		fmt.Printf("Parse errors:\n")
		for _, err := range p.Errors() {
			fmt.Printf("  %s\n", err)
		}
		os.Exit(1)
	}

	// Compile to bytecode
	comp := compiler.NewCompiler()
	err := comp.Compile(prog)
	if err != nil {
		fmt.Printf("Compilation error: %v\n", err)
		os.Exit(1)
	}

	// Execute the bytecode
	fmt.Println("Executing foreach loop:")
	vmCtx := vm.NewExecutionContext()
	virtualMachine := vm.NewVirtualMachine()
	err = virtualMachine.Execute(vmCtx, comp.GetBytecode(), comp.GetConstants())
	if err != nil {
		fmt.Printf("Execution error: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Println("Execution completed successfully")
}