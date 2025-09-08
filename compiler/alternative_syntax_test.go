package compiler

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wudi/php-parser/compiler/vm"
	"github.com/wudi/php-parser/lexer"
	"github.com/wudi/php-parser/parser"
)

// TestAlternativeIfStatement tests alternative if syntax (if/endif)
func TestAlternativeIfStatement(t *testing.T) {
	tests := []struct {
		name    string
		phpCode string
	}{
		{
			name: "Simple alternative if",
			phpCode: `<?php
$x = 5;
if ($x > 0):
    echo "positive";
endif;
?>`,
		},
		{
			name: "Alternative if with else",
			phpCode: `<?php
$x = -1;
if ($x > 0):
    echo "positive";
else:
    echo "not positive";
endif;
?>`,
		},
		{
			name: "Alternative if with elseif",
			phpCode: `<?php
$x = 0;
if ($x > 0):
    echo "positive";
elseif ($x == 0):
    echo "zero";
else:
    echo "negative";
endif;
?>`,
		},
		{
			name: "Nested alternative if",
			phpCode: `<?php
$x = 5;
$y = 10;
if ($x > 0):
    if ($y > 5):
        echo "both positive and y > 5";
    endif;
endif;
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

			// Test execution with runtime
			err = executeWithRuntime(t, comp)
			require.NoError(t, err, "Execution failed for test: %s", tt.name)
		})
	}
}

// TestAlternativeWhileStatement tests alternative while syntax (while/endwhile)
func TestAlternativeWhileStatement(t *testing.T) {
	tests := []struct {
		name    string
		phpCode string
	}{
		{
			name: "Simple alternative while",
			phpCode: `<?php
$i = 0;
while ($i < 3):
    echo $i;
    $i++;
endwhile;
?>`,
		},
		{
			name: "Alternative while with break",
			phpCode: `<?php
$i = 0;
while (true):
    if ($i >= 2) break;
    echo $i;
    $i++;
endwhile;
?>`,
		},
		{
			name: "Nested alternative while",
			phpCode: `<?php
$i = 0;
while ($i < 2):
    $j = 0;
    while ($j < 2):
        echo "$i,$j ";
        $j++;
    endwhile;
    $i++;
endwhile;
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

			// Test execution with runtime
			err = executeWithRuntime(t, comp)
			require.NoError(t, err, "Execution failed for test: %s", tt.name)
		})
	}
}

// TestAlternativeForStatement tests alternative for syntax (for/endfor)
func TestAlternativeForStatement(t *testing.T) {
	tests := []struct {
		name    string
		phpCode string
	}{
		{
			name: "Simple alternative for",
			phpCode: `<?php
for ($i = 0; $i < 3; $i++):
    echo $i;
endfor;
?>`,
		},
		{
			name: "Alternative for with multiple init",
			phpCode: `<?php
for ($i = 0, $j = 0; $i < 2; $i++, $j += 2):
    echo "$i,$j ";
endfor;
?>`,
		},
		{
			name: "Alternative for with break",
			phpCode: `<?php
for ($i = 0; $i < 10; $i++):
    if ($i >= 2) break;
    echo $i;
endfor;
?>`,
		},
		{
			name: "Nested alternative for",
			phpCode: `<?php
for ($i = 0; $i < 2; $i++):
    for ($j = 0; $j < 2; $j++):
        echo "$i,$j ";
    endfor;
endfor;
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

			// Test execution with runtime
			err = executeWithRuntime(t, comp)
			require.NoError(t, err, "Execution failed for test: %s", tt.name)
		})
	}
}

// TestAlternativeForeachStatement tests alternative foreach syntax (foreach/endforeach)
func TestAlternativeForeachStatement(t *testing.T) {
	tests := []struct {
		name    string
		phpCode string
	}{
		{
			name: "Simple alternative foreach",
			phpCode: `<?php
$arr = array(1, 2, 3);
foreach ($arr as $value):
    echo $value;
endforeach;
?>`,
		},
		{
			name: "Alternative foreach with key",
			phpCode: `<?php
$arr = array("a" => 1, "b" => 2);
foreach ($arr as $key => $value):
    echo "$key:$value ";
endforeach;
?>`,
		},
		{
			name: "Nested alternative foreach",
			phpCode: `<?php
$matrix = array(array(1, 2), array(3, 4));
foreach ($matrix as $row):
    foreach ($row as $value):
        echo "$value ";
    endforeach;
endforeach;
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

			// Now test execution with the full implementation
			vmCtx := vm.NewExecutionContext()
			err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants(), comp.GetFunctions(), comp.GetClasses())
			require.NoError(t, err, "Execution failed for test: %s", tt.name)
		})
	}
}

// TestMixedAlternativeSyntax tests mixing alternative and regular syntax
func TestMixedAlternativeSyntax(t *testing.T) {
	tests := []struct {
		name    string
		phpCode string
	}{
		{
			name: "Alternative if with regular for",
			phpCode: `<?php
$x = 5;
if ($x > 0):
    for ($i = 0; $i < 2; $i++) {
        echo $i;
    }
endif;
?>`,
		},
		{
			name: "Regular if with alternative while",
			phpCode: `<?php
$x = 5;
if ($x > 0) {
    $i = 0;
    while ($i < 2):
        echo $i;
        $i++;
    endwhile;
}
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

			// Test execution with runtime
			err = executeWithRuntime(t, comp)
			require.NoError(t, err, "Execution failed for test: %s", tt.name)
		})
	}
}
