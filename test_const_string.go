package main

import (
	"fmt"
	"github.com/wudi/hey/compiler/lexer"
)

func main() {
	input := `<?php $a = "x'"; ?>`

	l := lexer.New(input)

	fmt.Println("=== Tokens ===")
	for {
		tok := l.NextToken()
		if tok.Type == lexer.T_EOF {
			break
		}

		fmt.Printf("%s: %q (len=%d)\n", lexer.TokenNames[tok.Type], tok.Value, len(tok.Value))

		if tok.Type == lexer.T_CONSTANT_ENCAPSED_STRING {
			fmt.Println("  Bytes:")
			for i, ch := range []byte(tok.Value) {
				fmt.Printf("    [%d] = %q (%d)\n", i, ch, ch)
			}
		}
	}
}