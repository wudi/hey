package compiler

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wudi/php-parser/lexer"
	"github.com/wudi/php-parser/parser"
)

// TestArrowFunctions tests arrow function compilation and execution
func TestArrowFunctions(t *testing.T) {
	tests := []struct {
		name    string
		phpCode string
	}{
		{
			name: "Simple arrow function",
			phpCode: `<?php
$add = fn($a, $b) => $a + $b;
echo $add(1, 2);
?>`,
		},
		{
			name: "Arrow function with single parameter",
			phpCode: `<?php
$double = fn($x) => $x * 2;
echo $double(5);
?>`,
		},
		{
			name: "Arrow function with no parameters",
			phpCode: `<?php
$getValue = fn() => 42;
echo $getValue();
?>`,
		},
		{
			name: "Arrow function with type hints",
			phpCode: `<?php
$calculate = fn(int $a, int $b): int => $a * $b + 1;
echo $calculate(3, 4);
?>`,
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

			// For now, just test that compilation succeeds
			// TODO: Add execution testing when runtime supports closures better
		})
	}
}

// TestSpreadExpressions tests spread expression compilation
func TestSpreadExpressions(t *testing.T) {
	tests := []struct {
		name    string
		phpCode string
	}{
		{
			name: "Spread in array literal",
			phpCode: `<?php
$arr1 = [1, 2];
$arr2 = [...$arr1, 3, 4];
print_r($arr2);
?>`,
		},
		{
			name: "Multiple spreads in array",
			phpCode: `<?php
$first = [1, 2];
$second = [3, 4];
$combined = [...$first, 5, ...$second, 6];
print_r($combined);
?>`,
		},
		{
			name: "Spread with empty array",
			phpCode: `<?php
$empty = [];
$result = [...$empty, 1, 2, 3];
print_r($result);
?>`,
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

			// For now, just test that compilation succeeds
			// TODO: Add execution testing when VM supports spreads better
		})
	}
}

// TestModernPHPFeaturesCombined tests combining modern features
func TestModernPHPFeaturesCombined(t *testing.T) {
	tests := []struct {
		name    string
		phpCode string
	}{
		{
			name: "Arrow function with spread",
			phpCode: `<?php
$numbers = [1, 2, 3];
$process = fn($arr) => [...$arr, 4, 5];
$result = $process($numbers);
print_r($result);
?>`,
		},
		{
			name: "Complex arrow function",
			phpCode: `<?php
$data = [10, 20, 30];
$transform = fn($values) => array_sum([...$values, 40]);
echo $transform($data);
?>`,
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

			// Test that compilation succeeds for combined features
		})
	}
}
