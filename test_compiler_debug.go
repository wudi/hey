package main

import (
	"fmt"
	"github.com/wudi/hey/compiler"
	"github.com/wudi/hey/compiler/ast"
	"github.com/wudi/hey/compiler/lexer"
	"github.com/wudi/hey/compiler/parser"
)

func main() {
	input := `<?php
if (true) {
	?>HTML inside if
<?php
}
?>`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		fmt.Println("Parser errors:")
		for _, err := range p.Errors() {
			fmt.Println("  ", err)
		}
		return
	}

	// Print AST
	fmt.Println("=== AST ===")
	printAST(program, 0)

	// Compile
	fmt.Println("\n=== Bytecode ===")
	c := compiler.New()
	err := c.Compile(program)
	if err != nil {
		fmt.Println("Compile error:", err)
		return
	}

	bytecode := c.Bytecode()
	fmt.Printf("Instructions: %d bytes\n", len(bytecode.Instructions))
	fmt.Printf("Constants: %d\n", len(bytecode.Constants))

	// Disassemble
	fmt.Println("\nInstructions:")
	fmt.Println(bytecode.Instructions.String())
}

func printAST(node ast.Node, indent int) {
	prefix := ""
	for i := 0; i < indent; i++ {
		prefix += "  "
	}

	switch n := node.(type) {
	case *ast.Program:
		fmt.Printf("%sProgram (%d statements)\n", prefix, len(n.Body))
		for _, stmt := range n.Body {
			printAST(stmt, indent+1)
		}
	case *ast.IfStatement:
		fmt.Printf("%sIfStatement\n", prefix)
		fmt.Printf("%s  Condition:\n", prefix)
		printAST(n.Condition, indent+2)
		fmt.Printf("%s  Then:\n", prefix)
		printAST(n.ThenBranch, indent+2)
		if n.ElseBranch != nil {
			fmt.Printf("%s  Else:\n", prefix)
			printAST(n.ElseBranch, indent+2)
		}
	case *ast.BlockStatement:
		fmt.Printf("%sBlockStatement (%d statements)\n", prefix, len(n.Body))
		for _, stmt := range n.Body {
			printAST(stmt, indent+1)
		}
	case *ast.EchoStatement:
		fmt.Printf("%sEchoStatement (%d args)\n", prefix, len(n.Arguments.Arguments))
		for _, arg := range n.Arguments.Arguments {
			printAST(arg, indent+1)
		}
	case *ast.BooleanLiteral:
		fmt.Printf("%sBooleanLiteral: %v\n", prefix, n.Value)
	case *ast.StringLiteral:
		fmt.Printf("%sStringLiteral: %q\n", prefix, n.Value)
	default:
		fmt.Printf("%s%T\n", prefix, n)
	}
}