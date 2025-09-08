package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/wudi/php-parser/ast"
	"github.com/wudi/php-parser/compiler"
	"github.com/wudi/php-parser/compiler/runtime"
	"github.com/wudi/php-parser/compiler/values"
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

	// Initialize runtime if not already done
	if err = runtime.Bootstrap(); err != nil {
		fmt.Println("Failed to bootstrap runtime:", err)
		os.Exit(1)
	}

	// Initialize VM integration
	if err = runtime.InitializeVMIntegration(); err != nil {
		fmt.Println("Failed to initialize VM integration:", err)
		os.Exit(1)
	}

	vmCtx := vm.NewExecutionContext()

	if vmCtx.GlobalVars == nil {
		vmCtx.GlobalVars = make(map[string]*values.Value)
	}

	variables := runtime.GlobalVMIntegration.GetAllVariables()
	for name, value := range variables {
		vmCtx.GlobalVars[name] = value
	}

	// Create VM and set up compiler callback
	vmachine := vm.NewVirtualMachine()

	// Set up the compiler callback for include functionality
	vmachine.CompilerCallback = func(ctx *vm.ExecutionContext, program *ast.Program, filePath string, isRequired bool) (*values.Value, error) {
		// Create a new compiler for the included file
		comp := compiler.NewCompiler()
		if err := comp.Compile(program); err != nil {
			return nil, fmt.Errorf("compilation error in %s: %v", filePath, err)
		}

		// Create a new execution context for the included file but copy the variables
		// This allows variable sharing while preserving the main script's instruction state
		includeCtx := vm.NewExecutionContext()
		includeCtx.Variables = ctx.Variables         // Share variables
		includeCtx.Stack = ctx.Stack                 // Share stack
		includeCtx.IncludedFiles = ctx.IncludedFiles // Share included files tracking
		includeCtx.OutputWriter = ctx.OutputWriter   // Share output writer

		// Execute the compiled bytecode in the separate context
		err := vmachine.Execute(includeCtx, comp.GetBytecode(), comp.GetConstants(), comp.GetFunctions(), comp.GetClasses())
		if err != nil {
			return nil, fmt.Errorf("execution error in %s: %v", filePath, err)
		}

		// Copy back any changes to the shared state
		ctx.Variables = includeCtx.Variables
		ctx.Stack = includeCtx.Stack
		ctx.IncludedFiles = includeCtx.IncludedFiles
		// Output merging is now handled automatically by shared OutputWriter

		// Check if the included file executed an explicit return statement
		if includeCtx.Halted && len(includeCtx.Stack) > 0 {
			// Get the return value from the stack
			returnValue := includeCtx.Stack[len(includeCtx.Stack)-1]

			// Check if this is an explicit return (not just end of file)
			// In PHP, if a file ends without explicit return, it should return 1, not null
			if returnValue.IsNull() {
				// This is likely end-of-file, not an explicit return null
				return values.NewInt(1), nil
			}

			// Remove the return value from the stack and update both contexts
			includeCtx.Stack = includeCtx.Stack[:len(includeCtx.Stack)-1]
			ctx.Stack = includeCtx.Stack
			return returnValue, nil
		}

		// Return 1 on successful inclusion (PHP convention when no return statement)
		return values.NewInt(1), nil
	}

	// Execute the program
	err = vmachine.Execute(vmCtx, comp.GetBytecode(), comp.GetConstants(), comp.GetFunctions(), comp.GetClasses())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
