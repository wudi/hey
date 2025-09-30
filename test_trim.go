package main

import (
	"fmt"
	"strings"
)

func main() {
	tests := []string{
		`"x'y"`,
		`"x'"`,
		`"'y"`,
		`"dir='ltr'"`,
	}

	for _, test := range tests {
		result := strings.Trim(test, `"'`)
		fmt.Printf("%q -> %q\n", test, result)
	}
}