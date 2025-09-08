package compiler

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wudi/php-parser/lexer"
	"github.com/wudi/php-parser/parser"
)

// TestNumberLiterals tests compilation and execution of various number literal formats
func TestNumberLiterals(t *testing.T) {
	tests := []struct {
		name     string
		phpCode  string
		expected string
	}{
		{
			name:     "Binary literals",
			phpCode:  `<?php echo 0b1111 . "\n"; echo 0B1010 . "\n"; echo 0b0 . "\n"; ?>`,
			expected: "15\n10\n0\n",
		},
		{
			name:     "Hexadecimal literals",
			phpCode:  `<?php echo 0x1F . "\n"; echo 0X1f . "\n"; echo 0xff . "\n"; ?>`,
			expected: "31\n31\n255\n",
		},
		{
			name:     "Octal literals",
			phpCode:  `<?php echo 0123 . "\n"; echo 0o123 . "\n"; echo 0O777 . "\n"; ?>`,
			expected: "83\n83\n511\n",
		},
		{
			name:     "Decimal integers",
			phpCode:  `<?php echo 123 . "\n"; echo 0 . "\n"; echo 999 . "\n"; ?>`,
			expected: "123\n0\n999\n",
		},
		{
			name:     "Numbers with underscores",
			phpCode:  `<?php echo 1_000_000 . "\n"; echo 0xFF_AA . "\n"; echo 0b1010_1010 . "\n"; ?>`,
			expected: "1000000\n65450\n170\n",
		},
		{
			name:     "Float literals",
			phpCode:  `<?php echo 1.23 . "\n"; echo 3.14159 . "\n"; echo 1.0 . "\n"; ?>`,
			expected: "1.23\n3.14159\n1\n",
		},
		{
			name:     "Scientific notation",
			phpCode:  `<?php echo 1.23e4 . "\n"; echo 1.23E-4 . "\n"; echo 5e2 . "\n"; ?>`,
			expected: "12300\n0.000123\n500\n",
		},
		{
			name:     "Mixed number types in expressions",
			phpCode:  `<?php echo (0b1111 + 0x10) . "\n"; echo (123 + 1.5) . "\n"; ?>`,
			expected: "31\n124.5\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the PHP code
			p := parser.New(lexer.New(tt.phpCode))
			prog := p.ParseProgram()
			require.NotNil(t, prog, "Failed to parse program for test: %s", tt.name)

			// Compile the program
			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Compilation failed for test: %s", tt.name)

			// Capture output for verification
			var buf bytes.Buffer
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Execute with runtime
			err = executeWithRuntime(t, comp)
			require.NoError(t, err, "Execution failed for test: %s", tt.name)

			// Restore stdout and get output
			w.Close()
			os.Stdout = oldStdout
			buf.ReadFrom(r)

			output := buf.String()
			require.Equal(t, tt.expected, output, "Output mismatch for test: %s", tt.name)
		})
	}
}
