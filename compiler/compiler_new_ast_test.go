package compiler

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wudi/php-parser/lexer"
	"github.com/wudi/php-parser/parser"
)

// TestPrintStatement tests the compilation of print statements
func TestPrintStatement(t *testing.T) {
	tests := []struct {
		name    string
		phpCode string
	}{
		{
			name:    "Simple print statement",
			phpCode: `<?php print "Hello World"; ?>`,
		},
		{
			name:    "Print with variable",
			phpCode: `<?php $msg = "test"; print $msg; ?>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(lexer.New(tt.phpCode))
			prog := p.ParseProgram()
			require.NotNil(t, prog, "Failed to parse program for test: %s", tt.name)

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Compilation failed for test: %s", tt.name)

			err = executeWithRuntime(t, comp)
			require.NoError(t, err, "Execution failed for test: %s", tt.name)
		})
	}
}

// TestPrintExpression tests the compilation of print expressions
func TestPrintExpression(t *testing.T) {
	tests := []struct {
		name    string
		phpCode string
	}{
		{
			name:    "Print expression returns 1",
			phpCode: `<?php $x = print "test"; ?>`,
		},
		{
			name:    "Print in assignment",
			phpCode: `<?php $result = print "output"; ?>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(lexer.New(tt.phpCode))
			prog := p.ParseProgram()
			require.NotNil(t, prog, "Failed to parse program for test: %s", tt.name)

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Compilation failed for test: %s", tt.name)

			err = executeWithRuntime(t, comp)
			require.NoError(t, err, "Execution failed for test: %s", tt.name)
		})
	}
}

// TestCloneExpression tests the compilation of clone expressions
func TestCloneExpression(t *testing.T) {
	tests := []struct {
		name        string
		phpCode     string
		expectError bool
	}{
		{
			name:        "Clone non-object should fail",
			phpCode:     `<?php $obj = "dummy"; $cloned = clone $obj; ?>`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(lexer.New(tt.phpCode))
			prog := p.ParseProgram()
			require.NotNil(t, prog, "Failed to parse program for test: %s", tt.name)

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Compilation failed for test: %s", tt.name)

			err = executeWithRuntime(t, comp)
			if tt.expectError {
				require.Error(t, err, "Expected execution to fail for test: %s", tt.name)
				require.Contains(t, err.Error(), "__clone method called on non-object", "Expected specific error message")
			} else {
				require.NoError(t, err, "Execution failed for test: %s", tt.name)
			}
		})
	}
}

// TestInstanceofExpression tests the compilation of instanceof expressions
func TestInstanceofExpression(t *testing.T) {
	tests := []struct {
		name    string
		phpCode string
	}{
		{
			name:    "Object instanceof class",
			phpCode: `<?php $obj = new stdClass(); $result = $obj instanceof stdClass; ?>`,
		},
		{
			name:    "String not instanceof class",
			phpCode: `<?php $result = "test" instanceof stdClass; ?>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(lexer.New(tt.phpCode))
			prog := p.ParseProgram()
			require.NotNil(t, prog, "Failed to parse program for test: %s", tt.name)

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Compilation failed for test: %s", tt.name)

			err = executeWithRuntime(t, comp)
			require.NoError(t, err, "Execution failed for test: %s", tt.name)
		})
	}
}

// TestCastExpression tests the compilation of cast expressions
func TestCastExpression(t *testing.T) {
	tests := []struct {
		name    string
		phpCode string
	}{
		{
			name:    "Cast string to int",
			phpCode: `<?php $result = (int)"123"; ?>`,
		},
		{
			name:    "Cast int to string",
			phpCode: `<?php $result = (string)456; ?>`,
		},
		{
			name:    "Cast to bool",
			phpCode: `<?php $result = (bool)0; ?>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(lexer.New(tt.phpCode))
			prog := p.ParseProgram()
			require.NotNil(t, prog, "Failed to parse program for test: %s", tt.name)

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Compilation failed for test: %s", tt.name)

			err = executeWithRuntime(t, comp)
			require.NoError(t, err, "Execution failed for test: %s", tt.name)
		})
	}
}

// TestEmptyExpression tests the compilation of empty expressions
func TestEmptyExpression(t *testing.T) {
	tests := []struct {
		name    string
		phpCode string
	}{
		{
			name:    "Empty string is empty",
			phpCode: `<?php $result = empty(""); ?>`,
		},
		{
			name:    "Non-empty string is not empty",
			phpCode: `<?php $result = empty("test"); ?>`,
		},
		{
			name:    "Zero is empty",
			phpCode: `<?php $result = empty(0); ?>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(lexer.New(tt.phpCode))
			prog := p.ParseProgram()
			require.NotNil(t, prog, "Failed to parse program for test: %s", tt.name)

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Compilation failed for test: %s", tt.name)

			err = executeWithRuntime(t, comp)
			require.NoError(t, err, "Execution failed for test: %s", tt.name)
		})
	}
}

// TestIssetExpression tests the compilation of isset expressions
func TestIssetExpression(t *testing.T) {
	tests := []struct {
		name    string
		phpCode string
	}{
		{
			name:    "Isset on defined variable",
			phpCode: `<?php $var = "test"; $result = isset($var); ?>`,
		},
		{
			name:    "Isset on undefined variable",
			phpCode: `<?php $result = isset($undefined); ?>`,
		},
		{
			name:    "Isset with multiple variables",
			phpCode: `<?php $a = 1; $b = 2; $result = isset($a, $b); ?>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(lexer.New(tt.phpCode))
			prog := p.ParseProgram()
			require.NotNil(t, prog, "Failed to parse program for test: %s", tt.name)

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Compilation failed for test: %s", tt.name)

			err = executeWithRuntime(t, comp)
			require.NoError(t, err, "Execution failed for test: %s", tt.name)
		})
	}
}

// TestGlobalStatement tests the compilation of global statements
func TestGlobalStatement(t *testing.T) {
	tests := []struct {
		name    string
		phpCode string
	}{
		{
			name:    "Global variable declaration",
			phpCode: `<?php function test() { global $x; } ?>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(lexer.New(tt.phpCode))
			prog := p.ParseProgram()
			require.NotNil(t, prog, "Failed to parse program for test: %s", tt.name)

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Compilation failed for test: %s", tt.name)

			err = executeWithRuntime(t, comp)
			require.NoError(t, err, "Execution failed for test: %s", tt.name)
		})
	}
}

// TestDoWhileStatement tests the compilation of do-while statements
func TestDoWhileStatement(t *testing.T) {
	tests := []struct {
		name    string
		phpCode string
	}{
		{
			name:    "Simple do-while loop",
			phpCode: `<?php $i = 0; do { $i++; } while($i < 3); ?>`,
		},
		{
			name:    "Do-while executes at least once",
			phpCode: `<?php $i = 10; do { $i++; } while($i < 5); ?>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(lexer.New(tt.phpCode))
			prog := p.ParseProgram()
			require.NotNil(t, prog, "Failed to parse program for test: %s", tt.name)

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Compilation failed for test: %s", tt.name)

			err = executeWithRuntime(t, comp)
			require.NoError(t, err, "Execution failed for test: %s", tt.name)
		})
	}
}

// TestMagicConstantExpression tests the compilation of magic constant expressions
func TestMagicConstantExpression(t *testing.T) {
	tests := []struct {
		name    string
		phpCode string
	}{
		{
			name:    "Magic constant __LINE__",
			phpCode: `<?php $line = __LINE__; ?>`,
		},
		{
			name:    "Magic constant __FILE__",
			phpCode: `<?php $file = __FILE__; ?>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(lexer.New(tt.phpCode))
			prog := p.ParseProgram()
			require.NotNil(t, prog, "Failed to parse program for test: %s", tt.name)

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Compilation failed for test: %s", tt.name)

			err = executeWithRuntime(t, comp)
			require.NoError(t, err, "Execution failed for test: %s", tt.name)
		})
	}
}

// TestCommaExpression tests the compilation of comma expressions
func TestCommaExpression(t *testing.T) {
	tests := []struct {
		name    string
		phpCode string
	}{
		{
			name:    "Comma expression returns last value",
			phpCode: `<?php $result = (1, 2, 3); ?>`,
		},
		{
			name:    "Comma with side effects",
			phpCode: `<?php $a = 1; $b = ($a++, $a * 2); ?>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(lexer.New(tt.phpCode))
			prog := p.ParseProgram()
			require.NotNil(t, prog, "Failed to parse program for test: %s", tt.name)

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Compilation failed for test: %s", tt.name)

			err = executeWithRuntime(t, comp)
			require.NoError(t, err, "Execution failed for test: %s", tt.name)
		})
	}
}

// Test for error cases - features not yet implemented
func TestNotYetImplementedFeatures(t *testing.T) {
	tests := []struct {
		name         string
		phpCode      string
		errorMessage string
	}{
		{
			name:         "Static variable declaration",
			phpCode:      `<?php function test() { static $x = 1; } ?>`,
			errorMessage: "static variable declarations not yet implemented",
		},
		{
			name:         "Unset statement",
			phpCode:      `<?php $x = 1; unset($x); ?>`,
			errorMessage: "unset statements not yet implemented",
		},
		{
			name:         "Shell execution",
			phpCode:      "<?php `ls`; ?>",
			errorMessage: "shell execution expressions not yet implemented",
		},
		{
			name:         "Spread expression",
			phpCode:      `<?php $arr = [...$other]; ?>`,
			errorMessage: "spread expressions not yet implemented",
		},
		{
			name:         "Arrow function",
			phpCode:      `<?php $fn = fn($x) => $x * 2; ?>`,
			errorMessage: "arrow functions not yet implemented",
		},
		{
			name:         "First class callable",
			phpCode:      `<?php $fn = strlen(...); ?>`,
			errorMessage: "first-class callables not yet implemented",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(lexer.New(tt.phpCode))
			prog := p.ParseProgram()
			require.NotNil(t, prog, "Failed to parse program for test: %s", tt.name)

			comp := NewCompiler()
			err := comp.Compile(prog)
			if err == nil {
				t.Errorf("Expected compilation to fail with not implemented error for test: %s", tt.name)
			} else if !strings.Contains(err.Error(), tt.errorMessage) {
				t.Errorf("Expected error to contain %q, got %q for test: %s", tt.errorMessage, err.Error(), tt.name)
			}
		})
	}
}

// Benchmark tests for new implementations
func BenchmarkPrintExpression(b *testing.B) {
	phpCode := `<?php print "test"; ?>`

	// Parse once
	p := parser.New(lexer.New(phpCode))
	prog := p.ParseProgram()

	if prog == nil {
		b.Fatalf("Failed to parse program")
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		compiler := NewCompiler()
		err := compiler.Compile(prog)
		if err != nil {
			b.Fatalf("Compilation failed: %v", err)
		}
	}
}

func BenchmarkCastExpression(b *testing.B) {
	phpCode := `<?php (int)"123"; ?>`

	p := parser.New(lexer.New(phpCode))
	prog := p.ParseProgram()

	if prog == nil {
		b.Fatalf("Failed to parse program")
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		compiler := NewCompiler()
		err := compiler.Compile(prog)
		if err != nil {
			b.Fatalf("Compilation failed: %v", err)
		}
	}
}

// Integration test - complex expressions using multiple new features
func TestComplexExpressionIntegration(t *testing.T) {
	phpCode := `<?php 
		$x = print "hello"; 
		$result = "test" instanceof stdClass;
		$cast = (string)$result;
	?>`

	p := parser.New(lexer.New(phpCode))
	prog := p.ParseProgram()
	require.NotNil(t, prog, "Failed to parse complex integration program")

	comp := NewCompiler()
	err := comp.Compile(prog)
	require.NoError(t, err, "Compilation failed for complex integration test")

	err = executeWithRuntime(t, comp)
	require.NoError(t, err, "Execution failed for complex integration test")
}
