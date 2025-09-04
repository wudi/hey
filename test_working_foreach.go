package main

import (
	"fmt"

	"github.com/wudi/php-parser/compiler"
	"github.com/wudi/php-parser/compiler/vm"
	"github.com/wudi/php-parser/lexer"
	"github.com/wudi/php-parser/parser"
)

func main() {
	// Test the exact same foreach case from the working tests
	testCode("Foreach with Key (from tests)", 
		`<?php $arr = array("a" => 1, "b" => 2); foreach ($arr as $key => $value) { echo $key . ":" . $value; }`)
	
	// Test simpler key-value case
	testCode("Simple key-value", 
		`<?php $arr = [1,2,3]; foreach ($arr as $key => $value) { echo $key; echo ":"; echo $value; echo "\n"; }`)
		
	// Test value only case
	testCode("Value only", 
		`<?php $arr = [1,2,3]; foreach ($arr as $value) { echo $value; echo "\n"; }`)
}

func testCode(name, code string) {
	fmt.Printf("\n=== %s ===\n", name)
	
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

	// Execute
	fmt.Printf("Output: ")
	vmCtx := vm.NewExecutionContext()
	virtualMachine := vm.NewVirtualMachine()
	err = virtualMachine.Execute(vmCtx, comp.GetBytecode(), comp.GetConstants())
	if err != nil {
		fmt.Printf("Execution error: %v\n", err)
		return
	}
	fmt.Println()
}