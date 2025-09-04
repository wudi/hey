package main

import (
	"fmt"

	"github.com/wudi/php-parser/compiler"
	"github.com/wudi/php-parser/compiler/vm"
	"github.com/wudi/php-parser/lexer"
	"github.com/wudi/php-parser/parser"
)

func main() {
	// Test both simple assignment and foreach to see what changed
	fmt.Println("=== Testing simple variable assignment ===")
	testCode("Simple assign", `<?php $x = 123; echo $x; ?>`)
	
	fmt.Println("\n=== Testing foreach (minimal) ===")
	testCode("Foreach", `<?php $arr=[42]; foreach($arr as $v) { echo "[",$v,"]"; } ?>`)
}

func testCode(name, code string) {
	fmt.Printf("Test: %s\n", name)
	
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

	fmt.Print("Output: ")
	vmCtx := vm.NewExecutionContext()
	virtualMachine := vm.NewVirtualMachine()
	err = virtualMachine.Execute(vmCtx, comp.GetBytecode(), comp.GetConstants())
	if err != nil {
		fmt.Printf("Execution error: %v\n", err)
		return
	}
	fmt.Println()
}