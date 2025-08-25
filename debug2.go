package main

import (
	"fmt"
)

func main() {
	input := `<?php echo "Hello, World!"; ?>`
	
	fmt.Printf("Input: %q\n", input)
	fmt.Printf("Length: %d\n", len(input))
	
	// 手动检查字符
	for i, ch := range []byte(input) {
		fmt.Printf("input[%d] = '%c' (0x%02x)\n", i, ch, ch)
		if i >= 8 { // 只检查前几个字符
			break
		}
	}
}