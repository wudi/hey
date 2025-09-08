package compiler

import (
	"bytes"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wudi/php-parser/lexer"
	"github.com/wudi/php-parser/parser"
)

func TestStaticStatements(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{
			name: "basic static variable",
			code: `<?php
function test() {
    static $count = 0;
    $count++;
    echo $count;
}
test();
test();
test();
?>`,
			expected: "123",
		},
		{
			name: "static variable with string default",
			code: `<?php
function test() {
    static $name = "hello";
    echo $name;
}
test();
?>`,
			expected: "hello",
		},
		{
			name: "multiple static variables",
			code: `<?php
function test() {
    static $count = 0, $name = "test";
    $count++;
    echo $count . ":" . $name . "\n";
}
test();
test();
?>`,
			expected: "1:test\n2:test\n",
		},
		{
			name: "static variable without default value",
			code: `<?php
function test() {
    static $uninit;
    if (is_null($uninit)) {
        echo "null\n";
        $uninit = "initialized";
    } else {
        echo $uninit . "\n";
    }
}
test();
test();
?>`,
			expected: "null\ninitialized\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// First test with native PHP to verify expected behavior
			phpOutput := runPHPHelper(t, tt.code)
			require.Equal(t, tt.expected, phpOutput, "Native PHP output doesn't match expected for test: %s", tt.name)

			// Test with our compiler
			l := lexer.New(tt.code)
			p := parser.New(l)
			program := p.ParseProgram()
			require.NotNil(t, program)

			// Compile the AST
			comp := NewCompiler()
			err := comp.Compile(program)
			require.NoError(t, err, "Compilation failed for test: %s", tt.name)

			// Capture output
			buf := &bytes.Buffer{}
			r, w, _ := os.Pipe()
			oldStdout := os.Stdout
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

// runPHPHelper executes PHP code using the native PHP interpreter for comparison
func runPHPHelper(t *testing.T, code string) string {
	tmpfile, err := os.CreateTemp("", "test*.php")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.WriteString(code)
	require.NoError(t, err)
	tmpfile.Close()

	cmd := exec.Command("/usr/bin/php", tmpfile.Name())
	output, err := cmd.Output()
	require.NoError(t, err, "Failed to run native PHP")

	return string(output)
}
