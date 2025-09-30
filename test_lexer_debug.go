package main

import (
	"fmt"
	"github.com/wudi/hey/compiler/lexer"
)

func main() {
	input := `<?php
if (!did_action('admin_head')) {
	?>
<!DOCTYPE html>
<html>
<head>
	<title>Test</title>
</head>
</html>
<?php
}
?>`

	l := lexer.New(input)

	for {
		tok := l.NextToken()
		if tok.Type == lexer.T_EOF {
			break
		}

		if tok.Type == lexer.T_INLINE_HTML {
			fmt.Printf("T_INLINE_HTML: %q (len=%d)\n", tok.Value, len(tok.Value))
			// Print character by character
			for i, ch := range []byte(tok.Value) {
				if ch == '\n' {
					fmt.Printf("  [%d] = '\\n'\n", i)
				} else if ch == '\t' {
					fmt.Printf("  [%d] = '\\t'\n", i)
				} else if ch == ' ' {
					fmt.Printf("  [%d] = ' '\n", i)
				} else {
					fmt.Printf("  [%d] = %q\n", i, ch)
				}
			}
		} else if tok.Type == lexer.T_CLOSE_TAG {
			fmt.Printf("T_CLOSE_TAG\n")
		} else {
			fmt.Printf("%s: %s\n", lexer.TokenNames[tok.Type], tok.Value)
		}
	}
}