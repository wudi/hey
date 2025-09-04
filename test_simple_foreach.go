package main

import (
	"fmt"

	"github.com/wudi/php-parser/compiler"
	"github.com/wudi/php-parser/compiler/vm"
	"github.com/wudi/php-parser/lexer"
	"github.com/wudi/php-parser/parser"
)

func main() {
	// Test value-only foreach (no key)
	code := `<?php $arr = [1,2,3]; foreach($arr as $value) { echo "Value: ", $value, "\n"; } ?>`

	fmt.Println("Testing value-only foreach:")

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

	// Execute
	vmCtx := vm.NewExecutionContext()
	virtualMachine := vm.NewVirtualMachine()
	err = virtualMachine.Execute(vmCtx, comp.GetBytecode(), comp.GetConstants())
	if err != nil {
		fmt.Printf("Execution error: %v\n", err)
		return
	}
	
	fmt.Println("Test completed")
}