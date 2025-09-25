package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/chzyer/readline"
	"github.com/urfave/cli/v3"
	"github.com/wudi/hey/compiler"
	"github.com/wudi/hey/compiler/lexer"
	"github.com/wudi/hey/compiler/parser"
	"github.com/wudi/hey/runtime"
	"github.com/wudi/hey/values"
	"github.com/wudi/hey/version"
	"github.com/wudi/hey/vm"
	"github.com/wudi/hey/vmfactory"
)

func main() {
	// Check if the first argument is a PHP file
	// If it is, bypass CLI framework and directly execute the file with all arguments
	if len(os.Args) > 1 {
		filename := os.Args[1]
		// Check if it's not a flag and the file exists
		if !strings.HasPrefix(filename, "-") {
			if _, err := os.Stat(filename); err == nil {
				// File exists, execute it directly with all arguments
				scriptArgs := append([]string{filename}, os.Args[2:]...)
				if err := parseAndExecuteFileWithArgs(filename, scriptArgs); err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(1)
				}
				return
			}
		}
	}

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
				Name:  "a",
				Local: true,
				Usage: "Run as interactive shell",
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
			&cli.BoolFlag{
				Name:    "version",
				Local:   true,
				Aliases: []string{"v"},
				Usage:   "Show version information",
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
			// Check if version is requested
			if cmd.Bool("version") {
				fmt.Printf("Hey %s\n", version.FullVersion())
				fmt.Printf("Build: %s\n", version.Build())
				fmt.Printf("Commit: %s\n", version.Commit())
				return nil
			}

			// Check if interactive mode is requested
			if cmd.Bool("a") {
				return runInteractiveShell()
			}

			// This path is no longer reached for PHP files as they're handled before CLI parsing
			// This remains for backward compatibility with non-file arguments
			if len(os.Args) > 1 && !strings.HasPrefix(os.Args[1], "-") {
				// If we get here, the file doesn't exist
				return fmt.Errorf("file not found: %s", os.Args[1])
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

func parseAndExecuteFileWithArgs(filename string, args []string) error {
	code, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	return parseAndExecuteCodeWithFileAndArgs(string(code), false, filename, args)
}

func parseAndExecuteCodeWithFile(code string, inScript bool, filename string) error {
	return parseAndExecuteCodeWithFileAndArgs(code, inScript, filename, []string{filename})
}

func parseAndExecuteCodeWithFileAndArgs(code string, inScript bool, filename string, args []string) error {
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
	if err := runtime.Bootstrap(); err != nil {
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
	if err := runtime.InitializeVMIntegration(); err != nil {
		fmt.Println("Failed to initialize VM integration:", err)
		os.Exit(1)
	}

	vmCtx := vm.NewExecutionContext()

	// GlobalVars is already initialized as sync.Map in NewExecutionContext()

	variables := runtime.GlobalVMIntegration.GetAllVariables()
	for name, value := range variables {
		vmCtx.GlobalVars.Store(name, value)
	}

	// Populate $argc and $argv
	if len(args) > 0 {
		argc := values.NewInt(int64(len(args)))
		vmCtx.GlobalVars.Store("$argc", argc)

		argv := values.NewArray()
		for i, arg := range args {
			argv.ArraySet(values.NewInt(int64(i)), values.NewString(arg))
		}
		vmCtx.GlobalVars.Store("$argv", argv)
	}

	// Create VM with pre-configured compiler callback
	factory := vmfactory.NewVMFactory(func() vmfactory.Compiler {
		return compiler.NewCompiler()
	})
	vmachine := factory.CreateVM()

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
	if err := runtime.Bootstrap(); err != nil {
		fmt.Println("Failed to bootstrap runtime:", err)
		os.Exit(1)
	}

	comp := compiler.NewCompiler()
	if err := comp.Compile(prog); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Initialize VM integration
	if err := runtime.InitializeVMIntegration(); err != nil {
		fmt.Println("Failed to initialize VM integration:", err)
		os.Exit(1)
	}

	vmCtx := vm.NewExecutionContext()

	// GlobalVars is already initialized as sync.Map in NewExecutionContext()

	variables := runtime.GlobalVMIntegration.GetAllVariables()
	for name, value := range variables {
		vmCtx.GlobalVars.Store(name, value)
	}

	// Create VM with pre-configured compiler callback
	factory := vmfactory.NewVMFactory(func() vmfactory.Compiler {
		return compiler.NewCompiler()
	})
	vmachine := factory.CreateVM()

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

// trackingWriter wraps io.Writer and tracks if output was written
type trackingWriter struct {
	w              io.Writer
	hasOutput      bool
	lastWasNewline bool
}

func (tw *trackingWriter) Write(p []byte) (n int, err error) {
	tw.hasOutput = true
	if len(p) > 0 {
		tw.lastWasNewline = p[len(p)-1] == '\n'
	}
	return tw.w.Write(p)
}

func (tw *trackingWriter) Reset() {
	tw.hasOutput = false
	tw.lastWasNewline = false
}

func runInteractiveShell() error {
	fmt.Printf("Welcome to Hey %s. Build: %s\n", version.FullVersion(), version.Build())

	// Initialize runtime and VM integration once
	if err := runtime.Bootstrap(); err != nil {
		return fmt.Errorf("Failed to bootstrap runtime: %v", err)
	}

	if err := runtime.InitializeVMIntegration(); err != nil {
		return fmt.Errorf("Failed to initialize VM integration: %v", err)
	}

	// Create persistent VM context and machine
	vmCtx := vm.NewExecutionContext()
	// GlobalVars is already initialized as sync.Map in NewExecutionContext()

	// Create tracking writer for output
	outputTracker := &trackingWriter{w: os.Stdout}
	vmCtx.OutputWriter = outputTracker

	variables := runtime.GlobalVMIntegration.GetAllVariables()
	for name, value := range variables {
		vmCtx.GlobalVars.Store(name, value)
	}

	factory := vmfactory.NewVMFactory(func() vmfactory.Compiler {
		return compiler.NewCompiler()
	})
	vmachine := factory.CreateVM()

	// Get history file path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "/tmp"
	}
	historyFile := homeDir + "/.hey_history"

	// Create readline instance with configuration
	config := &readline.Config{
		Prompt:            ">",
		HistoryFile:       historyFile,
		HistoryLimit:      1000,
		InterruptPrompt:   "^C",
		EOFPrompt:         "exit",
		HistorySearchFold: true, // Case-insensitive history search

		// Enable vi mode keys (can be toggled by user preference)
		VimMode: false,

		// Custom key bindings (these are defaults but we specify them explicitly)
		// Ctrl+A: move to beginning of line
		// Ctrl+E: move to end of line
		// Ctrl+W: delete word before cursor
		// Ctrl+K: delete from cursor to end of line
		// Ctrl+U: delete from cursor to beginning of line
		// Arrow keys: move cursor left/right
		// Ctrl+D: exit if line is empty
		// Ctrl+R: reverse history search
		// Tab: auto-completion (could be enhanced with PHP function names)
	}

	rl, err := readline.NewEx(config)
	if err != nil {
		return err
	}
	defer rl.Close()

	multilineBuffer := ""
	inMultiline := false

	for {
		// Set the appropriate prompt
		if inMultiline {
			rl.SetPrompt("... ")
		} else {
			rl.SetPrompt("> ")
		}

		line, err := rl.Readline()
		if err != nil {
			if err == readline.ErrInterrupt {
				if multilineBuffer != "" {
					// Clear multiline buffer on interrupt
					multilineBuffer = ""
					inMultiline = false
					fmt.Println("^C")
					continue
				}
				// Exit on second interrupt
				break
			} else if err == io.EOF {
				break
			}
			return err
		}

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
		outputTracker.Reset()
		executeREPLCode(code, vmCtx, vmachine)

		// Add newline after output if needed
		if outputTracker.hasOutput && !outputTracker.lastWasNewline {
			fmt.Println()
		}
	}

	return nil
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

	// Reset Halted state from any previous execution
	vmCtx.Halted = false

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
