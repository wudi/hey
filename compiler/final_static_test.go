package compiler

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wudi/php-parser/lexer"
	"github.com/wudi/php-parser/parser"
)

func TestFinalStaticImplementation(t *testing.T) {
	code := `<?php
// Test 1: Basic static counter
function counter() {
    static $count = 0;
    $count++;
    return $count;
}

echo "Counter test: ";
echo counter() . " ";
echo counter() . " ";
echo counter() . "\n";

// Test 2: Multiple static variables
function multi_static() {
    static $name = "PHP", $version = 8.4;
    echo "Language: $name $version\n";
}

multi_static();
multi_static();

// Test 3: Static without default
function uninitialized_static() {
    static $unset;
    if (is_null($unset)) {
        $unset = "now_set";
        echo "Was null, now set\n";
    } else {
        echo "Still set: $unset\n";
    }
}

uninitialized_static();
uninitialized_static();

echo "All tests completed!\n";
?>`

	expected := `Counter test: 1 2 3
Language: PHP 8.4
Language: PHP 8.4
Was null, now set
Still set: now_set
All tests completed!
`

	// Compile the AST
	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()
	require.NotNil(t, program)

	comp := NewCompiler()
	err := comp.Compile(program)
	require.NoError(t, err, "Compilation failed")

	// Capture output
	buf := &bytes.Buffer{}
	r, w, _ := os.Pipe()
	oldStdout := os.Stdout
	os.Stdout = w

	// Execute with runtime
	err = executeWithRuntime(t, comp)
	require.NoError(t, err, "Execution failed")

	// Restore stdout and get output
	w.Close()
	os.Stdout = oldStdout
	buf.ReadFrom(r)

	output := buf.String()
	require.Equal(t, expected, output, "Output mismatch - static statements not working correctly")
}
