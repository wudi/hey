package compiler

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wudi/php-parser/lexer"
	"github.com/wudi/php-parser/parser"
)

// TestGotoStatement tests goto statement compilation
func TestGotoStatement(t *testing.T) {
	tests := []struct {
		name    string
		phpCode string
	}{
		{
			name: "Simple goto forward",
			phpCode: `<?php
echo "Before goto\n";
goto target;
echo "This will be skipped\n";
target:
echo "After goto\n";
?>`,
		},
		{
			name: "Goto backward",
			phpCode: `<?php
$i = 0;
loop:
echo $i . "\n";
$i++;
if ($i < 3) goto loop;
echo "Done\n";
?>`,
		},
		{
			name: "Goto with conditional",
			phpCode: `<?php
$x = 5;
if ($x > 0) goto positive;
echo "Not positive\n";
goto end;
positive:
echo "Is positive\n";
end:
echo "Finished\n";
?>`,
		},
		{
			name: "Multiple labels",
			phpCode: `<?php
$choice = 2;
if ($choice == 1) goto first;
if ($choice == 2) goto second;
goto third;

first:
echo "First\n";
goto end;

second:
echo "Second\n";
goto end;

third:
echo "Third\n";

end:
echo "End\n";
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

// TestLabelStatement tests label statement compilation
func TestLabelStatement(t *testing.T) {
	tests := []struct {
		name    string
		phpCode string
	}{
		{
			name: "Label without goto (no effect)",
			phpCode: `<?php
echo "Before label\n";
my_label:
echo "After label\n";
?>`,
		},
		{
			name: "Multiple labels in sequence",
			phpCode: `<?php
echo "Start\n";
label1:
label2:
label3:
echo "All labels defined\n";
?>`,
		},
		{
			name: "Label in conditional block",
			phpCode: `<?php
$x = true;
if ($x) {
    inner_label:
    echo "Inside conditional\n";
}
echo "Outside\n";
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

// TestDeclareStatement tests declare statement compilation
func TestDeclareStatement(t *testing.T) {
	tests := []struct {
		name    string
		phpCode string
	}{
		{
			name: "Simple declare strict_types",
			phpCode: `<?php
declare(strict_types=1);
echo "Strict types enabled\n";
?>`,
		},
		{
			name: "Declare with ticks",
			phpCode: `<?php
declare(ticks=1);
echo "Ticks enabled\n";
?>`,
		},
		{
			name: "Declare with encoding",
			phpCode: `<?php
declare(encoding='UTF-8');
echo "Encoding set\n";
?>`,
		},
		{
			name: "Multiple declare directives",
			phpCode: `<?php
declare(strict_types=1, ticks=1);
echo "Multiple directives\n";
?>`,
		},
		{
			name: "Declare block syntax",
			phpCode: `<?php
declare(ticks=1) {
    echo "Inside declare block\n";
    echo "Still inside\n";
}
echo "Outside block\n";
?>`,
		},
		{
			name: "Declare alternative syntax",
			phpCode: `<?php
declare(ticks=1):
    echo "Alternative declare syntax\n";
enddeclare;
echo "After declare\n";
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

			// Declare statements mostly affect compile-time behavior
			// For now, just test that they compile successfully
		})
	}
}

// TestAdvancedControlFlow tests complex combinations of advanced features
func TestAdvancedControlFlow(t *testing.T) {
	tests := []struct {
		name    string
		phpCode string
	}{
		{
			name: "Goto with loops",
			phpCode: `<?php
$i = 0;
start_loop:
if ($i >= 3) goto end_loop;
echo "Iteration: $i\n";
$i++;
goto start_loop;
end_loop:
echo "Loop finished\n";
?>`,
		},
		{
			name: "Nested goto and labels",
			phpCode: `<?php
$outer = 2;
$inner = 1;

outer_loop:
if ($outer <= 0) goto done;
echo "Outer: $outer\n";

inner_check:
if ($inner > 3) {
    $inner = 1;
    $outer--;
    goto outer_loop;
}
echo "  Inner: $inner\n";
$inner++;
goto inner_check;

done:
echo "All done\n";
?>`,
		},
		{
			name: "Declare with goto",
			phpCode: `<?php
declare(strict_types=1);
$x = 10;
if ($x > 5) goto skip_message;
echo "x is small\n";
skip_message:
echo "x is: $x\n";
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
