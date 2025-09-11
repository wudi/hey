package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/urfave/cli/v3"
	"github.com/wudi/hey/compiler"
	"github.com/wudi/hey/compiler/ast"
	"github.com/wudi/hey/compiler/lexer"
	"github.com/wudi/hey/compiler/parser"
	"github.com/wudi/hey/compiler/runtime"
	"github.com/wudi/hey/compiler/values"
	"github.com/wudi/hey/compiler/vm"
	"github.com/wudi/hey/version"
)

func main() {
	app := &cli.Command{
		Name:  "hey",
		Usage: "A PHP interpreter written in Go",
		Commands: []*cli.Command{
			initCommand,     // hey init
			requireCommand,  // hey require
			installCommand,  // hey install
			updateCommand,   // hey update
			validateCommand, // hey validate
			fpmCommand,      // hey fpm
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "code",
				Local:   true,
				Aliases: []string{"r"},
				Usage:   "Run PHP <code> without using script tags <?..?>",
				Action: func(ctx context.Context, cmd *cli.Command, s string) error {
					return parseAndExecuteCode(s, true)
				},
			},
			&cli.StringFlag{
				Name:        "version",
				Local:       true,
				Aliases:     []string{"v"},
				Usage:       "Show version",
				Destination: nil,
				Action: func(ctx context.Context, cmd *cli.Command, s string) error {
					fmt.Println(version.Version())
					return nil
				},
			},
			&cli.StringFlag{
				Name:    "file",
				Local:   true,
				Aliases: []string{"f"},
				Usage:   "Parse and execute <file>.",
				Action: func(ctx context.Context, cmd *cli.Command, s string) error {
					return parseAndExecuteFile(s)
				},
			},
			&cli.StringFlag{
				Name:  "S",
				Local: true,
				Usage: "<addr>:<port> Run with built-in web server.",
				Action: func(ctx context.Context, cmd *cli.Command, s string) error {
					return runWebServer(s)
				},
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			// if flowed file input is provided, use it
			if len(os.Args) > 1 {
				return parseAndExecuteFile(os.Args[1])
			}

			// read from stdin if no file or code is provided
			code, err := io.ReadAll(os.Stdin)
			if err != nil {
				return err
			}

			fmt.Println(string(code))

			return parseAndExecuteCode(string(code), false)
		},
	}

	err := app.Run(context.Background(), os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func parseAndExecuteFile(filename string) error {
	code, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	return parseAndExecuteCode(string(code), false)
}

func parseAndExecuteCode(code string, inScript bool) error {
	var l *lexer.Lexer
	if inScript {
		l = lexer.NewInScripting(code)
	} else {
		l = lexer.New(code)
	}

	p := parser.New(l)
	prog := p.ParseProgram()
	if len(p.Errors()) != 0 {
		for _, msg := range p.Errors() {
			fmt.Println(msg)
		}
		os.Exit(1)
	}

	comp := compiler.NewCompiler()
	if err := comp.Compile(prog); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Initialize runtime if not already done
	if err := runtime.Bootstrap(); err != nil {
		fmt.Println("Failed to bootstrap runtime:", err)
		os.Exit(1)
	}

	// Initialize VM integration
	if err := runtime.InitializeVMIntegration(); err != nil {
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
		err := vmachine.Execute(includeCtx, comp.GetBytecode(), comp.GetConstants(), comp.GetVMFunctions(), comp.GetVMClasses())
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
	return vmachine.Execute(vmCtx, comp.GetBytecode(), comp.GetConstants(), comp.GetVMFunctions(), comp.GetVMClasses())
}

func runWebServer(addr string) error {
	// Placeholder for built-in web server functionality
	fmt.Printf("Starting built-in web server at %s (not yet implemented)\n", addr)
	return nil
}
