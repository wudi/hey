package main

import (
	"fmt"

	"github.com/wudi/php-parser/compiler"
	"github.com/wudi/php-parser/compiler/vm"
	"github.com/wudi/php-parser/lexer"
	"github.com/wudi/php-parser/parser"
)

func testCode(name, code string) {
	fmt.Printf("\n=== Testing: %s ===\n", name)
	fmt.Printf("Code: %s\n", code)

	// Parse the PHP code
	p := parser.New(lexer.New(code))
	prog := p.ParseProgram()
	
	if len(p.Errors()) > 0 {
		fmt.Printf("Parse errors:\n")
		for _, err := range p.Errors() {
			fmt.Printf("  %s\n", err)
		}
		return
	}

	// Compile to bytecode
	comp := compiler.NewCompiler()
	err := comp.Compile(prog)
	if err != nil {
		fmt.Printf("Compilation error: %v\n", err)
		return
	}

	fmt.Printf("Bytecode instructions: %d\n", len(comp.GetBytecode()))
	
	// Execute the bytecode
	fmt.Println("Output:")
	vmCtx := vm.NewExecutionContext()
	virtualMachine := vm.NewVirtualMachine()
	err = virtualMachine.Execute(vmCtx, comp.GetBytecode(), comp.GetConstants())
	if err != nil {
		fmt.Printf("Execution error: %v\n", err)
		return
	}
	fmt.Println("\n--- End Output ---")
}

func main() {
	// Test simple echo first
	testCode("Simple echo", `<?php echo "Hello"; ?>`)
	
	// Test array creation and echo
	testCode("Array creation", `<?php $arr = [1,2,3]; echo "Created array"; ?>`)
	
	// Test simple foreach
	testCode("Simple foreach", `<?php $arr = [1,2,3]; foreach($arr as $value) { echo $value; } ?>`)
	
	// Test foreach with key-value
	testCode("Foreach with key-value", `<?php $arr = [1,2,3]; foreach($arr as $key => $value) { echo $key, ":", $value, "\n"; } ?>`)
}