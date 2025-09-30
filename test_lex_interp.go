package main

import (
	"fmt"
	"github.com/wudi/hey/compiler/lexer"
)

func main() {
	input := `<?php
$a = "x'";
echo "$a";
?>`

	l := lexer.New(input)

	fmt.Println("=== Tokens ===")
	for {
		tok := l.NextToken()
		if tok.Type == lexer.T_EOF {
			break
		}

		if tok.Type == lexer.T_ENCAPSED_AND_WHITESPACE {
			fmt.Printf("%s: %q (len=%d)\n", lexer.TokenNames[tok.Type], tok.Value, len(tok.Value))
			for i, ch := range []byte(tok.Value) {
				fmt.Printf("  [%d] = %q (%d)\n", i, ch, ch)
			}
		} else {
			fmt.Printf("%s: %q\n", lexer.TokenNames[tok.Type], tok.Value)
		}
	}
}