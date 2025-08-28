package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wudi/php-parser/lexer"
)

func TestPrattParser_Creation(t *testing.T) {
	input := "<?php 42;"
	l := lexer.New(input)
	p := NewPrattParser(l)
	
	assert.NotNil(t, p)
	assert.NotNil(t, p.lexer)
	assert.NotNil(t, p.parsingContext)
}

func TestPrattParser_SimpleIntegerLiteral(t *testing.T) {
	input := "<?php 42;"
	program, errors := ParsePHP(input)
	
	assert.Empty(t, errors)
	assert.NotNil(t, program)
	assert.Len(t, program.Statements, 1)
}

func TestPrattParser_SimpleBinaryExpression(t *testing.T) {
	input := "<?php 1 + 2;"
	program, errors := ParsePHP(input)
	
	assert.Empty(t, errors)
	assert.NotNil(t, program)
	assert.Len(t, program.Statements, 1)
}