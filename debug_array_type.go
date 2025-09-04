package main

import (
	"fmt"

	"github.com/wudi/php-parser/compiler"
	"github.com/wudi/php-parser/compiler/vm"
	"github.com/wudi/php-parser/compiler/values"
	"github.com/wudi/php-parser/lexer"
	"github.com/wudi/php-parser/parser"
)

func main() {
	// Test simple array creation
	code := `<?php $arr = [1,2,3]; echo "Done"; ?>`

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

	// Execute and check the array variable
	vmCtx := vm.NewExecutionContext()
	virtualMachine := vm.NewVirtualMachine()
	
	// Execute bytecode
	err = virtualMachine.Execute(vmCtx, comp.GetBytecode(), comp.GetConstants())
	if err != nil {
		fmt.Printf("Execution error: %v\n", err)
		return
	}
	
	// Check the $arr variable (VAR:41 from our previous debug)
	if arrVal, exists := vmCtx.Variables[41]; exists {
		fmt.Printf("Array variable found:\n")
		fmt.Printf("  Type: %d\n", arrVal.Type)
		fmt.Printf("  TypeArray constant: %d\n", values.TypeArray)
		fmt.Printf("  String representation: %s\n", arrVal.String())
		fmt.Printf("  Is TypeArray: %v\n", arrVal.Type == values.TypeArray)
		
		if arrVal.Type == values.TypeArray {
			arrayData := arrVal.Data.(*values.Array)
			fmt.Printf("  Array elements count: %d\n", len(arrayData.Elements))
			for k, v := range arrayData.Elements {
				fmt.Printf("    [%v] = %s\n", k, v.String())
			}
		}
	} else {
		fmt.Printf("Array variable not found in Variables map\n")
		fmt.Printf("Variables in context: %d\n", len(vmCtx.Variables))
		for k, v := range vmCtx.Variables {
			fmt.Printf("  VAR:%d = %s\n", k, v.String())
		}
	}
}