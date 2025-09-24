package main

import (
	"fmt"
	"regexp"
	"strconv"
)

// naturalCompare implements PHP's natural string comparison
func naturalCompare(s1, s2 string) int {
	// Regular expression to split strings into chunks of digits and non-digits
	re := regexp.MustCompile(`(\d+|\D+)`)

	parts1 := re.FindAllString(s1, -1)
	parts2 := re.FindAllString(s2, -1)

	minLen := len(parts1)
	if len(parts2) < minLen {
		minLen = len(parts2)
	}

	for i := 0; i < minLen; i++ {
		p1, p2 := parts1[i], parts2[i]

		// Check if both parts are numeric
		if isNumeric(p1) && isNumeric(p2) {
			n1, _ := strconv.ParseInt(p1, 10, 64)
			n2, _ := strconv.ParseInt(p2, 10, 64)
			if n1 < n2 {
				return -1
			} else if n1 > n2 {
				return 1
			}
		} else {
			// String comparison
			if p1 < p2 {
				return -1
			} else if p1 > p2 {
				return 1
			}
		}
	}

	// If all compared parts are equal, compare lengths
	if len(parts1) < len(parts2) {
		return -1
	} else if len(parts1) > len(parts2) {
		return 1
	}

	return 0
}

func isNumeric(s string) bool {
	_, err := strconv.ParseInt(s, 10, 64)
	return err == nil
}

func main() {
	fmt.Println("Testing naturalCompare function:")

	tests := [][]string{
		{"item1", "item2"},
		{"item2", "item10"},
		{"item10", "item20"},
		{"item1", "item10"},
	}

	for _, test := range tests {
		result := naturalCompare(test[0], test[1])
		fmt.Printf("naturalCompare('%s', '%s') = %d\n", test[0], test[1], result)
	}
}