package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/wudi/php-parser/compiler"
	"github.com/wudi/php-parser/compiler/vm"
	"github.com/wudi/php-parser/lexer"
	"github.com/wudi/php-parser/parser"
)

var input string

func main() {
	flag.StringVar(&input, "i", "", "Input file")
	flag.Parse()

	if input == "" {
		fmt.Println("Input file is required")
		os.Exit(1)
	}

	code, err := os.ReadFile(input)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	p := parser.New(lexer.New(string(code)))
	prog := p.ParseProgram()
	if len(p.Errors()) != 0 {
		for _, msg := range p.Errors() {
			fmt.Println(msg)
		}
		os.Exit(1)
	}

	comp := compiler.NewCompiler()
	err = comp.Compile(prog)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	vmCtx := vm.NewExecutionContext()
	err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants(), comp.GetFunctions(), comp.GetClasses())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
