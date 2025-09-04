package main

import (
	"fmt"

	"github.com/wudi/php-parser/compiler"
	"github.com/wudi/php-parser/compiler/vm"
	"github.com/wudi/php-parser/lexer"
	"github.com/wudi/php-parser/parser"
)

func main() {
	// Test simple echo first to verify basics work
	code1 := `<?php echo "Test works\n"; ?>`
	fmt.Println("=== Test 1: Simple Echo ===")
	testCode(code1)
	
	// Test simple foreach with simple loop body
	code2 := `<?php $arr = [1,2,3]; foreach($arr as $value) { echo $value, "\n"; } ?>`
	fmt.Println("\n=== Test 2: Simple Foreach ===")
	testCode(code2)
}

func testCode(code string) {
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

	// Execute without debug mode first
	fmt.Println("Output:")
	vmCtx := vm.NewExecutionContext()
	virtualMachine := vm.NewVirtualMachine()
	err = virtualMachine.Execute(vmCtx, comp.GetBytecode(), comp.GetConstants())
	if err != nil {
		fmt.Printf("Execution error: %v\n", err)
		return
	}
	fmt.Println("--- End Output ---")
}