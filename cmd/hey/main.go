package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/urfave/cli/v3"
	"github.com/wudi/hey/compiler"
	"github.com/wudi/hey/compiler/ast"
	"github.com/wudi/hey/compiler/lexer"
	"github.com/wudi/hey/compiler/parser"
	runtime2 "github.com/wudi/hey/runtime"
	"github.com/wudi/hey/values"
	"github.com/wudi/hey/version"
	"github.com/wudi/hey/vm"
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
			&cli.BoolFlag{
				Name:    "a",
				Local:   true,
				Usage:   "Run as interactive shell",
			},
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
			// Check if interactive mode is requested
			if cmd.Bool("a") {
				return runInteractiveShell()
			}

			// if flowed file input is provided, use it
			if len(os.Args) > 1 {
				return parseAndExecuteFile(os.Args[1])
			}

			// read from stdin if no file or code is provided
			code, err := io.ReadAll(os.Stdin)
			if err != nil {
				return err
			}

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
	return parseAndExecuteCodeWithFile(string(code), false, filename)
}

func parseAndExecuteCodeWithFile(code string, inScript bool, filename string) error {
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

	// Initialize runtime first so constants are registered
	if err := runtime2.Bootstrap(); err != nil {
		fmt.Println("Failed to bootstrap runtime:", err)
		os.Exit(1)
	}

	comp := compiler.NewCompiler()
	// Set the current file for magic constants
	if filename != "" {
		comp.SetCurrentFile(filename)
	}
	if err := comp.Compile(prog); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Initialize VM integration
	if err := runtime2.InitializeVMIntegration(); err != nil {
		fmt.Println("Failed to initialize VM integration:", err)
		os.Exit(1)
	}

	vmCtx := vm.NewExecutionContext()

	if vmCtx.GlobalVars == nil {
		vmCtx.GlobalVars = make(map[string]*values.Value)
	}

	variables := runtime2.GlobalVMIntegration.GetAllVariables()
	for name, value := range variables {
		vmCtx.GlobalVars[name] = value
	}

	// Create VM and set up compiler callback
	vmachine := vm.NewVirtualMachine()

	// Set up the compiler callback for include functionality
	vmachine.CompilerCallback = func(ctx *vm.ExecutionContext, program *ast.Program, filePath string, isRequired bool) (*values.Value, error) {
		// Create a new compiler for the included file
		comp := compiler.NewCompiler()
		// Set the file path for the included file
		if filePath != "" {
			comp.SetCurrentFile(filePath)
		}
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
		err := vmachine.Execute(includeCtx, comp.GetBytecode(), comp.GetConstants(), comp.Functions(), comp.Classes(), comp.Interfaces(), comp.Traits())
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
	// Execute the script
	err := vmachine.Execute(vmCtx, comp.GetBytecode(), comp.GetConstants(), comp.Functions(), comp.Classes(), comp.Interfaces(), comp.Traits())

	// Call destructors on all remaining objects at script end
	vmachine.CallAllDestructors(vmCtx)

	// Check if exit() or die() was called
	if vmCtx.Halted {
		os.Exit(vmCtx.ExitCode)
	}

	return err
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

	// Initialize runtime first so constants are registered
	if err := runtime2.Bootstrap(); err != nil {
		fmt.Println("Failed to bootstrap runtime:", err)
		os.Exit(1)
	}

	comp := compiler.NewCompiler()
	if err := comp.Compile(prog); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Initialize VM integration
	if err := runtime2.InitializeVMIntegration(); err != nil {
		fmt.Println("Failed to initialize VM integration:", err)
		os.Exit(1)
	}

	vmCtx := vm.NewExecutionContext()

	if vmCtx.GlobalVars == nil {
		vmCtx.GlobalVars = make(map[string]*values.Value)
	}

	variables := runtime2.GlobalVMIntegration.GetAllVariables()
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
		err := vmachine.Execute(includeCtx, comp.GetBytecode(), comp.GetConstants(), comp.Functions(), comp.Classes(), comp.Interfaces(), comp.Traits())
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
	// Execute the script
	err := vmachine.Execute(vmCtx, comp.GetBytecode(), comp.GetConstants(), comp.Functions(), comp.Classes(), comp.Interfaces(), comp.Traits())

	// Call destructors on all remaining objects at script end
	vmachine.CallAllDestructors(vmCtx)

	// Check if exit() or die() was called
	if vmCtx.Halted {
		os.Exit(vmCtx.ExitCode)
	}

	return err
}

func runWebServer(addr string) error {
	// Placeholder for built-in web server functionality
	fmt.Printf("Starting built-in web server at %s (not yet implemented)\n", addr)
	return nil
}

func runInteractiveShell() error {
	fmt.Println("Interactive mode enabled.")

	// Initialize runtime and VM integration once
	if err := runtime2.Bootstrap(); err != nil {
		return fmt.Errorf("Failed to bootstrap runtime: %v", err)
	}

	if err := runtime2.InitializeVMIntegration(); err != nil {
		return fmt.Errorf("Failed to initialize VM integration: %v", err)
	}

	// Create persistent VM context and machine
	vmCtx := vm.NewExecutionContext()
	if vmCtx.GlobalVars == nil {
		vmCtx.GlobalVars = make(map[string]*values.Value)
	}

	variables := runtime2.GlobalVMIntegration.GetAllVariables()
	for name, value := range variables {
		vmCtx.GlobalVars[name] = value
	}

	vmachine := vm.NewVirtualMachine()
	setupCompilerCallback(vmachine)

	scanner := bufio.NewScanner(os.Stdin)
	multilineBuffer := ""
	inMultiline := false

	for {
		if inMultiline {
			fmt.Print("... ")
		} else {
			fmt.Print("hey > ")
		}

		if !scanner.Scan() {
			break
		}

		line := scanner.Text()

		// Handle exit commands
		if !inMultiline && (line == "exit" || line == "quit" || line == "exit()" || line == "quit()") {
			fmt.Println("Bye!")
			break
		}

		// Accumulate multiline input
		multilineBuffer += line + "\n"

		// Check if we need to continue multiline input
		if needsMoreInput(multilineBuffer) {
			inMultiline = true
			continue
		}

		// Process the complete input
		inMultiline = false
		code := strings.TrimSpace(multilineBuffer)
		multilineBuffer = ""

		if code == "" {
			continue
		}

		// Wrap code in PHP tags if not present
		if !strings.HasPrefix(code, "<?") {
			code = "<?php " + code + " ?>"
		}

		// Execute the code
		executeREPLCode(code, vmCtx, vmachine)
	}

	return scanner.Err()
}

func needsMoreInput(code string) bool {
	// Simple heuristic: check for unclosed braces, quotes, etc.
	openBraces := 0
	openParens := 0
	openBrackets := 0
	inSingleQuote := false
	inDoubleQuote := false
	escaped := false

	for _, ch := range code {
		if escaped {
			escaped = false
			continue
		}

		if ch == '\\' {
			escaped = true
			continue
		}

		if !inSingleQuote && !inDoubleQuote {
			switch ch {
			case '\'':
				inSingleQuote = true
			case '"':
				inDoubleQuote = true
			case '{':
				openBraces++
			case '}':
				openBraces--
			case '(':
				openParens++
			case ')':
				openParens--
			case '[':
				openBrackets++
			case ']':
				openBrackets--
			}
		} else if inSingleQuote && ch == '\'' {
			inSingleQuote = false
		} else if inDoubleQuote && ch == '"' {
			inDoubleQuote = false
		}
	}

	// Need more input if we have unclosed constructs
	return openBraces > 0 || openParens > 0 || openBrackets > 0 ||
	       inSingleQuote || inDoubleQuote
}

func executeREPLCode(code string, vmCtx *vm.ExecutionContext, vmachine *vm.VirtualMachine) {
	// Store initial stack size to detect if expression left a value
	initialStackSize := len(vmCtx.Stack)

	// Parse the code
	l := lexer.New(code)
	p := parser.New(l)
	prog := p.ParseProgram()

	if len(p.Errors()) != 0 {
		for _, msg := range p.Errors() {
			fmt.Printf("Parse error: %s\n", msg)
		}
		return
	}

	// Compile the code
	comp := compiler.NewCompiler()
	if err := comp.Compile(prog); err != nil {
		fmt.Printf("Compile error: %v\n", err)
		return
	}

	// Execute the code in the persistent context
	err := vmachine.Execute(vmCtx, comp.GetBytecode(), comp.GetConstants(),
	                        comp.Functions(), comp.Classes(), comp.Interfaces(), comp.Traits())

	if err != nil {
		fmt.Printf("Runtime error: %v\n", err)
		return
	}

	// If execution left a value on the stack (expression evaluation), print it
	if len(vmCtx.Stack) > initialStackSize && !vmCtx.Halted {
		// Get the top value(s) added during this execution
		for i := initialStackSize; i < len(vmCtx.Stack); i++ {
			topValue := vmCtx.Stack[i]
			// Only print if it's not null/void
			if !topValue.IsNull() {
				fmt.Println(topValue.String())
			}
		}
		// Clear any expression results from the stack
		vmCtx.Stack = vmCtx.Stack[:initialStackSize]
	}
}

func setupCompilerCallback(vmachine *vm.VirtualMachine) {
	vmachine.CompilerCallback = func(ctx *vm.ExecutionContext, program *ast.Program, filePath string, isRequired bool) (*values.Value, error) {
		comp := compiler.NewCompiler()
		if filePath != "" {
			comp.SetCurrentFile(filePath)
		}
		if err := comp.Compile(program); err != nil {
			return nil, fmt.Errorf("compilation error in %s: %v", filePath, err)
		}

		includeCtx := vm.NewExecutionContext()
		includeCtx.Variables = ctx.Variables
		includeCtx.Stack = ctx.Stack
		includeCtx.IncludedFiles = ctx.IncludedFiles
		includeCtx.OutputWriter = ctx.OutputWriter

		err := vmachine.Execute(includeCtx, comp.GetBytecode(), comp.GetConstants(),
		                        comp.Functions(), comp.Classes(), comp.Interfaces(), comp.Traits())
		if err != nil {
			return nil, fmt.Errorf("execution error in %s: %v", filePath, err)
		}

		ctx.Variables = includeCtx.Variables
		ctx.Stack = includeCtx.Stack
		ctx.IncludedFiles = includeCtx.IncludedFiles

		if includeCtx.Halted && len(includeCtx.Stack) > 0 {
			returnValue := includeCtx.Stack[len(includeCtx.Stack)-1]
			if returnValue.IsNull() {
				return values.NewInt(1), nil
			}
			includeCtx.Stack = includeCtx.Stack[:len(includeCtx.Stack)-1]
			ctx.Stack = includeCtx.Stack
			return returnValue, nil
		}

		return values.NewInt(1), nil
	}
}
