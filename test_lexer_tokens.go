package main

import (
	"fmt"
	"github.com/wudi/hey/compiler/lexer"
)

func main() {
	input := `<?php
$d = "dir='ltr'";
echo "$d";
?>`

	l := lexer.New(input)

	fmt.Println("=== Tokens ===")
	for {
		tok := l.NextToken()
		if tok.Type == lexer.T_EOF {
			break
		}

		fmt.Printf("%s: %q\n", lexer.TokenNames[tok.Type], tok.Value)
	}
}