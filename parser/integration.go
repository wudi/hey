package parser

import (
	"io/ioutil"

	"github.com/wudi/php-parser/ast"
	"github.com/wudi/php-parser/lexer"
)

// ParsePHP is the main entry point for parsing PHP code using the new Pratt parser
func ParsePHP(input string) (*ast.Program, []string) {
	l := lexer.New(input)
	p := NewPrattParser(l)
	program := p.ParseProgram()
	return program, p.Errors()
}

// ParsePHPWithVersion parses PHP code with a specific PHP version compatibility
func ParsePHPWithVersion(input string, version PHPVersion) (*ast.Program, []string) {
	l := lexer.New(input)
	p := NewPrattParser(l)
	p.parsingContext.PHPVersion = version
	program := p.ParseProgram()
	return program, p.Errors()
}

// ParseFile parses a PHP file and returns the AST
func ParseFile(filename string) (*ast.Program, []string, error) {
	// Read file content
	content, err := readFile(filename)
	if err != nil {
		return nil, nil, err
	}
	
	// Parse the content
	program, errors := ParsePHP(content)
	return program, errors, nil
}

// ParseExpression parses a single PHP expression
func ParseExpression(input string) (ast.Expression, []string) {
	l := lexer.New(input)
	p := NewPrattParser(l)
	
	// Skip opening PHP tag if present
	if p.currentTokenIs(lexer.T_OPEN_TAG) {
		p.nextToken()
	}
	
	expr := p.parseExpression(LOWEST)
	return expr, p.Errors()
}

// ParseStatement parses a single PHP statement
func ParseStatement(input string) (ast.Statement, []string) {
	l := lexer.New(input)
	p := NewPrattParser(l)
	
	// Skip opening PHP tag if present
	if p.currentTokenIs(lexer.T_OPEN_TAG) {
		p.nextToken()
	}
	
	stmt := p.parseStatement()
	return stmt, p.Errors()
}

// ParseDeclaration parses a single PHP declaration
func ParseDeclaration(input string) (ast.Declaration, []string) {
	l := lexer.New(input)
	p := NewPrattParser(l)
	
	// Skip opening PHP tag if present
	if p.currentTokenIs(lexer.T_OPEN_TAG) {
		p.nextToken()
	}
	
	decl := p.parseDeclaration()
	return decl, p.Errors()
}


// ============= UTILITY FUNCTIONS =============

func readFile(filename string) (string, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(content), nil
}


// ============= FEATURE FLAGS =============

// ParserConfig allows configuring parser features
type ParserConfig struct {
	UseNewParser      bool
	PHPVersion        PHPVersion
	EnableAttributes  bool
	EnableEnums       bool
	EnableMatchExpr   bool
	EnableNullsafe    bool
	EnableNamedArgs   bool
	EnablePropertyHooks bool
}

// DefaultConfig returns default parser configuration
func DefaultConfig() *ParserConfig {
	return &ParserConfig{
		UseNewParser:      true,
		PHPVersion:        PHP84,
		EnableAttributes:  true,
		EnableEnums:       true,
		EnableMatchExpr:   true,
		EnableNullsafe:    true,
		EnableNamedArgs:   true,
		EnablePropertyHooks: true,
	}
}

// ParseWithConfig parses PHP code with specific configuration
func ParseWithConfig(input string, config *ParserConfig) (*ast.Program, []string) {
	// Use new Pratt parser
	l := lexer.New(input)
	p := NewPrattParser(l)
	p.parsingContext.PHPVersion = config.PHPVersion
	
	// Configure features based on config
	// (Features are already implemented in the parser, this is for future fine-tuning)
	
	program := p.ParseProgram()
	return program, p.Errors()
}