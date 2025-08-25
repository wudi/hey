package main

import (
	"fmt"
)

// 模拟 lexer 的 peekCharN 逻辑
func main() {
	input := `<?php echo "Hello"; ?>`
	position := 0      // 当前位置（指向当前字符）
	readPosition := 1  // 读取位置（指向下一个字符）
	
	fmt.Printf("Input: %q\n", input)
	fmt.Printf("Current position: %d ('%c')\n", position, input[position])
	fmt.Printf("Read position: %d ('%c')\n", readPosition, input[readPosition])
	
	// 模拟 peekChar() - 查看下一个字符
	peekChar := func() byte {
		if readPosition >= len(input) {
			return 0
		}
		return input[readPosition]
	}
	
	// 模拟 peekCharN(n) - 查看第 n 个字符后的字符
	peekCharN := func(n int) byte {
		pos := readPosition + n
		if pos >= len(input) {
			return 0
		}
		return input[pos]
	}
	
	ch := input[position] // 当前字符 '<'
	
	fmt.Printf("\nChecking logic:\n")
	fmt.Printf("ch = '%c'\n", ch)
	fmt.Printf("peekChar() = '%c'\n", peekChar())
	
	if ch == '<' {
		fmt.Println("Found '<'")
		
		if peekChar() == '?' {
			fmt.Println("Found '<?'")
			
			fmt.Printf("peekCharN(0) = '%c' (should be '?')\n", peekCharN(0))
			fmt.Printf("peekCharN(1) = '%c' (should be 'p')\n", peekCharN(1))
			fmt.Printf("peekCharN(2) = '%c' (should be 'h')\n", peekCharN(2))
			fmt.Printf("peekCharN(3) = '%c' (should be 'p')\n", peekCharN(3))
			
			// 原始代码中的检查
			if peekCharN(1) == 'p' && peekCharN(2) == 'h' && peekCharN(3) == 'p' {
				fmt.Println("SUCCESS: Found complete '<?php' tag!")
			} else {
				fmt.Println("FAILED: Not a complete '<?php' tag")
			}
		}
	}
}