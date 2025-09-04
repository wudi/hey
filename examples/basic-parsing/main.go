package main

import (
	"fmt"

	"github.com/wudi/php-parser/lexer"
	"github.com/wudi/php-parser/parser"
)

func main() {
	// Sample PHP code to parse
	phpCode := `<?php
$name = "World";
echo "Hello, " . $name . "!";

function greet($person) {
    return "Hello, " . $person;
}

class User {
    private $name;
    
    public function __construct($name) {
        $this->name = $name;
    }
    
    public function getName() {
        return $this->name;
    }
}
?>`

	// Create lexer
	l := lexer.New(phpCode)

	// Create parser
	p := parser.New(l)

	// Parse the code
	program := p.ParseProgram()

	// Check for parsing errors
	if len(p.Errors()) > 0 {
		fmt.Println("Parsing errors:")
		for _, err := range p.Errors() {
			fmt.Printf("  %s\n", err)
		}
		return
	}

	// Print basic information about the parsed program
	fmt.Printf("Successfully parsed PHP code!\n")
	fmt.Printf("Number of statements: %d\n", len(program.Body))
	fmt.Println("\nParsed statements:")

	for i, stmt := range program.Body {
		fmt.Printf("  %d. %s\n", i+1, stmt.String())
	}

	// Print the full AST structure
	fmt.Println("\nFull AST:")
	fmt.Println(program.String())
}
