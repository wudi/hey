package main

import (
	"fmt"
)

// 我们需要创建一个包装器来访问 lexer 的内部方法
func main() {
	input := `<?php echo "Hello"; ?>`
	
	fmt.Printf("Input: %q\n", input)
	fmt.Printf("Length: %d\n", len(input))
	
	// 手动检查 peekCharN 逻辑
	for i := 0; i < len(input); i++ {
		ch := input[i]
		fmt.Printf("input[%d] = '%c'\n", i, ch)
		
		if ch == '<' && i+1 < len(input) && input[i+1] == '?' {
			fmt.Printf("Found '<?' at position %d\n", i)
			
			if i+2 < len(input) {
				fmt.Printf("  Next char (pos %d): '%c'\n", i+2, input[i+2])
			}
			if i+3 < len(input) {
				fmt.Printf("  Next char (pos %d): '%c'\n", i+3, input[i+3])
			}
			if i+4 < len(input) {
				fmt.Printf("  Next char (pos %d): '%c'\n", i+4, input[i+4])
			}
			
			// 检查是否匹配 "php"
			if i+4 < len(input) &&
				input[i+2] == 'p' &&
				input[i+3] == 'h' &&
				input[i+4] == 'p' {
				fmt.Printf("  Found complete '<?php' tag!\n")
			} else {
				fmt.Printf("  Not a complete '<?php' tag\n")
			}
			break
		}
	}
}