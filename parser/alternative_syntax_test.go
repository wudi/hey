package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wudi/php-parser/ast"
	"github.com/wudi/php-parser/lexer"
)

func TestParsing_AlternativeIfStatement(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedErrors int
	}{
		{
			name: "Alternative If with elseif and else",
			input: `<?php
if ($condition):
    echo "true";
elseif ($other):
    echo "other";
else:
    echo "false";
endif;
?>`,
			expectedErrors: 0,
		},
		{
			name: "Alternative If simple",
			input: `<?php
if ($x > 0):
    echo "positive";
endif;
?>`,
			expectedErrors: 0,
		},
		{
			name: "Alternative If with multiple elseifs",
			input: `<?php
if ($x == 1):
    echo "one";
elseif ($x == 2):
    echo "two";
elseif ($x == 3):
    echo "three";
else:
    echo "other";
endif;
?>`,
			expectedErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()

			assert.Equal(t, tt.expectedErrors, len(p.Errors()), "Parser errors: %v", p.Errors())
			assert.NotNil(t, program)
			assert.Greater(t, len(program.Body), 0, "Expected at least one statement")

			// Check that we have an AlternativeIfStatement
			stmt := program.Body[0]
			altIfStmt, ok := stmt.(*ast.AlternativeIfStatement)
			assert.True(t, ok, "Expected AlternativeIfStatement, got %T", stmt)
			assert.NotNil(t, altIfStmt.Condition, "Expected condition to be set")
		})
	}
}

func TestParsing_AlternativeWhileStatement(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedErrors int
	}{
		{
			name: "Alternative While loop",
			input: `<?php
while ($counter < 10):
    echo $counter;
    $counter++;
endwhile;
?>`,
			expectedErrors: 0,
		},
		{
			name: "Alternative While with complex condition",
			input: `<?php
while ($i < count($array) && $flag):
    process($array[$i]);
    $i++;
endwhile;
?>`,
			expectedErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()

			assert.Equal(t, tt.expectedErrors, len(p.Errors()), "Parser errors: %v", p.Errors())
			assert.NotNil(t, program)
			assert.Greater(t, len(program.Body), 0, "Expected at least one statement")

			// Check that we have an AlternativeWhileStatement
			stmt := program.Body[0]
			altWhileStmt, ok := stmt.(*ast.AlternativeWhileStatement)
			assert.True(t, ok, "Expected AlternativeWhileStatement, got %T", stmt)
			assert.NotNil(t, altWhileStmt.Condition, "Expected condition to be set")
			assert.Greater(t, len(altWhileStmt.Body), 0, "Expected body statements")
		})
	}
}

func TestParsing_AlternativeForStatement(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedErrors int
	}{
		{
			name: "Alternative For loop",
			input: `<?php
for ($i = 0; $i < 5; $i++):
    echo $i;
endfor;
?>`,
			expectedErrors: 0,
		},
		{
			name: "Alternative For with empty init",
			input: `<?php
for (; $i < 10; $i++):
    echo $i;
endfor;
?>`,
			expectedErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()

			assert.Equal(t, tt.expectedErrors, len(p.Errors()), "Parser errors: %v", p.Errors())
			assert.NotNil(t, program)
			assert.Greater(t, len(program.Body), 0, "Expected at least one statement")

			// Check that we have an AlternativeForStatement
			stmt := program.Body[0]
			altForStmt, ok := stmt.(*ast.AlternativeForStatement)
			assert.True(t, ok, "Expected AlternativeForStatement, got %T", stmt)
			assert.Greater(t, len(altForStmt.Body), 0, "Expected body statements")
		})
	}
}

func TestParsing_AlternativeForeachStatement(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedErrors int
	}{
		{
			name: "Alternative Foreach with value only",
			input: `<?php
foreach ($array as $value):
    echo $value;
endforeach;
?>`,
			expectedErrors: 0,
		},
		{
			name: "Alternative Foreach with key and value",
			input: `<?php
foreach ($array as $key => $value):
    echo $key . ": " . $value;
endforeach;
?>`,
			expectedErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()

			assert.Equal(t, tt.expectedErrors, len(p.Errors()), "Parser errors: %v", p.Errors())
			assert.NotNil(t, program)
			assert.Greater(t, len(program.Body), 0, "Expected at least one statement")

			// Check that we have an AlternativeForeachStatement
			stmt := program.Body[0]
			altForeachStmt, ok := stmt.(*ast.AlternativeForeachStatement)
			assert.True(t, ok, "Expected AlternativeForeachStatement, got %T", stmt)
			assert.NotNil(t, altForeachStmt.Iterable, "Expected iterable to be set")
			assert.NotNil(t, altForeachStmt.Value, "Expected value to be set")
			assert.Greater(t, len(altForeachStmt.Body), 0, "Expected body statements")
		})
	}
}

func TestParsing_DeclareStatement(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedErrors int
		expectAlt      bool
	}{
		{
			name: "Alternative Declare statement",
			input: `<?php
declare(strict_types=1):
    function test(): int {
        return 42;
    }
enddeclare;
?>`,
			expectedErrors: 0,
			expectAlt:      true,
		},
		{
			name: "Alternative Declare with multiple declarations",
			input: `<?php
declare(ticks=1, encoding='UTF-8'):
    echo "Hello World";
enddeclare;
?>`,
			expectedErrors: 0,
			expectAlt:      true,
		},
		{
			name: "Regular Declare statement",
			input: `<?php
declare(strict_types=1);
?>`,
			expectedErrors: 0,
			expectAlt:      false,
		},
		{
			name: "Regular Declare with block",
			input: `<?php
declare(strict_types=1) {
    function test() {}
}
?>`,
			expectedErrors: 0,
			expectAlt:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()

			assert.Equal(t, tt.expectedErrors, len(p.Errors()), "Parser errors: %v", p.Errors())
			assert.NotNil(t, program)
			assert.Greater(t, len(program.Body), 0, "Expected at least one statement")

			// Check that we have a DeclareStatement
			stmt := program.Body[0]
			declareStmt, ok := stmt.(*ast.DeclareStatement)
			assert.True(t, ok, "Expected DeclareStatement, got %T", stmt)
			assert.Greater(t, len(declareStmt.Declarations), 0, "Expected at least one declaration")
			assert.Equal(t, tt.expectAlt, declareStmt.Alternative, "Expected alternative flag to match")
		})
	}
}