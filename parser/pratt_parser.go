package parser

import (
	"fmt"

	"github.com/wudi/php-parser/ast"
	"github.com/wudi/php-parser/lexer"
)

// PrattParser implements a Pratt parser for PHP expressions with enhanced architecture
type PrattParser struct {
	lexer        *lexer.Lexer
	currentToken lexer.Token
	peekToken    lexer.Token
	errors       []string
	
	// Parser function registries
	prefixParseFns map[lexer.TokenType]PrefixParseFn
	infixParseFns  map[lexer.TokenType]InfixParseFn
	
	// Statement parsers organized by grammar categories
	statementParsers    map[lexer.TokenType]StatementParseFn
	declarationParsers  map[lexer.TokenType]DeclarationParseFn
	
	// Modern PHP feature parsers
	attributeParsers    map[lexer.TokenType]AttributeParseFn
	typeParsers        map[lexer.TokenType]TypeParseFn
	
	// Context tracking for complex parsing
	parsingContext *ParsingContext
}

// ParsingContext tracks the current parsing state and context
type ParsingContext struct {
	InClass          bool
	InFunction       bool
	InInterface      bool
	InTrait          bool
	InEnum           bool
	InNamespace      bool
	InAttribute      bool
	ClassVisibility  string
	FunctionName     string
	NamespaceName    string
	PHPVersion       PHPVersion
	
	// Modern PHP 8+ contexts
	InMatch          bool
	InPropertyHook   bool
	InArrowFunction  bool
	InUnionType      bool
	InIntersectionType bool
}

// PHPVersion represents supported PHP versions for feature compatibility
type PHPVersion int

const (
	PHP74 PHPVersion = iota
	PHP80
	PHP81
	PHP82
	PHP83
	PHP84
)

// Parser function types for different grammar categories
type (
	PrefixParseFn      func() ast.Expression
	InfixParseFn       func(ast.Expression) ast.Expression
	StatementParseFn   func() ast.Statement
	DeclarationParseFn func() ast.Declaration
	AttributeParseFn   func() ast.AttributeList
	TypeParseFn        func() ast.Type
)

// Declaration interface for all declaration nodes
type Declaration interface {
	ast.Node
	declarationNode()
}

// NewPrattParser creates a new enhanced Pratt parser instance
func NewPrattParser(l *lexer.Lexer) *PrattParser {
	p := &PrattParser{
		lexer:  l,
		errors: []string{},
		parsingContext: &ParsingContext{
			PHPVersion: PHP84, // Default to latest supported version
		},
	}
	
	// Initialize parser function registries
	p.initializePrefixParsers()
	p.initializeInfixParsers()
	p.initializeStatementParsers()
	p.initializeDeclarationParsers()
	p.initializeAttributeParsers()
	p.initializeTypeParsers()
	
	// Initialize token position
	p.nextToken()
	p.nextToken()
	
	return p
}

// ============= EXPRESSION PARSING ARCHITECTURE =============

// Precedence levels following PHP 8.4 official specification
type Precedence int

const (
	_ Precedence = iota
	LOWEST
	PIPE          // |> (PHP 8.4+)
	ASSIGNMENT    // = += -= *= /= .= %= &= |= ^= <<= >>= **= ??=
	TERNARY       // ? :
	COALESCE      // ??
	LOGICAL_OR    // || or
	LOGICAL_AND   // && and
	LOGICAL_XOR   // xor
	BITWISE_OR    // |
	BITWISE_XOR   // ^
	BITWISE_AND   // &
	EQUALITY      // == != === !==
	RELATIONAL    // < <= > >= <=> instanceof
	SHIFT         // << >>
	CONCATENATION // .
	ADDITIVE      // + -
	MULTIPLICATIVE // * / %
	EXPONENTIAL   // **
	UNARY         // ! ~ -X +X ++X --X cast @
	POSTFIX       // X++ X--
	MEMBER_ACCESS // -> ?-> :: []
	PRIMARY       // () literals variables
)

// Precedence mapping for all PHP operators
var precedenceMap = map[lexer.TokenType]Precedence{
	// PHP 8.4+ Pipe operator
	lexer.T_PIPE:                   PIPE,
	
	// Assignment operators
	lexer.TOKEN_EQUAL:              ASSIGNMENT,
	lexer.T_PLUS_EQUAL:             ASSIGNMENT,
	lexer.T_MINUS_EQUAL:            ASSIGNMENT,
	lexer.T_MUL_EQUAL:              ASSIGNMENT,
	lexer.T_DIV_EQUAL:              ASSIGNMENT,
	lexer.T_CONCAT_EQUAL:           ASSIGNMENT,
	lexer.T_MOD_EQUAL:              ASSIGNMENT,
	lexer.T_AND_EQUAL:              ASSIGNMENT,
	lexer.T_OR_EQUAL:               ASSIGNMENT,
	lexer.T_XOR_EQUAL:              ASSIGNMENT,
	lexer.T_SL_EQUAL:               ASSIGNMENT,
	lexer.T_SR_EQUAL:               ASSIGNMENT,
	lexer.T_POW_EQUAL:              ASSIGNMENT,
	lexer.T_COALESCE_EQUAL:         ASSIGNMENT,
	
	// Ternary and coalesce
	lexer.TOKEN_QUESTION:           TERNARY,
	lexer.T_COALESCE:               COALESCE,
	
	// Logical operators
	lexer.T_BOOLEAN_OR:             LOGICAL_OR,
	lexer.T_LOGICAL_OR:             LOGICAL_OR,
	lexer.T_BOOLEAN_AND:            LOGICAL_AND,
	lexer.T_LOGICAL_AND:            LOGICAL_AND,
	lexer.T_LOGICAL_XOR:            LOGICAL_XOR,
	
	// Bitwise operators
	lexer.TOKEN_PIPE:               BITWISE_OR,
	lexer.TOKEN_CARET:              BITWISE_XOR,
	lexer.TOKEN_AMPERSAND:          BITWISE_AND,
	
	// Equality and relational
	lexer.T_IS_EQUAL:               EQUALITY,
	lexer.T_IS_NOT_EQUAL:           EQUALITY,
	lexer.T_IS_IDENTICAL:           EQUALITY,
	lexer.T_IS_NOT_IDENTICAL:       EQUALITY,
	lexer.TOKEN_LT:                 RELATIONAL,
	lexer.TOKEN_GT:                 RELATIONAL,
	lexer.T_IS_SMALLER_OR_EQUAL:    RELATIONAL,
	lexer.T_IS_GREATER_OR_EQUAL:    RELATIONAL,
	lexer.T_SPACESHIP:              RELATIONAL,
	lexer.T_INSTANCEOF:             RELATIONAL,
	
	// Shift operators
	lexer.T_SL:                     SHIFT,
	lexer.T_SR:                     SHIFT,
	
	// String concatenation
	lexer.TOKEN_DOT:                CONCATENATION,
	
	// Arithmetic operators
	lexer.TOKEN_PLUS:               ADDITIVE,
	lexer.TOKEN_MINUS:              ADDITIVE,
	lexer.TOKEN_MULTIPLY:           MULTIPLICATIVE,
	lexer.TOKEN_DIVIDE:             MULTIPLICATIVE,
	lexer.TOKEN_MODULO:             MULTIPLICATIVE,
	lexer.T_POW:                    EXPONENTIAL,
	
	// Postfix operators
	lexer.T_INC:                    POSTFIX,
	lexer.T_DEC:                    POSTFIX,
	
	// Member access operators
	lexer.T_OBJECT_OPERATOR:        MEMBER_ACCESS,
	lexer.T_NULLSAFE_OBJECT_OPERATOR: MEMBER_ACCESS,
	lexer.T_PAAMAYIM_NEKUDOTAYIM:   MEMBER_ACCESS,
	lexer.TOKEN_LBRACKET:           MEMBER_ACCESS,
	lexer.TOKEN_LPAREN:             MEMBER_ACCESS, // Function calls
}

// getPrecedence returns the precedence of the given token type
func (p *PrattParser) getPrecedence(tokenType lexer.TokenType) Precedence {
	if prec, exists := precedenceMap[tokenType]; exists {
		return prec
	}
	return LOWEST
}

// parseExpression is the core Pratt parser expression parsing method
func (p *PrattParser) parseExpression(precedence Precedence) ast.Expression {
	// Get prefix parser for current token
	prefix := p.prefixParseFns[p.currentToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.currentToken.Type)
		return nil
	}
	
	leftExp := prefix()
	
	// Continue parsing while we have infix operators with higher precedence
	for !p.peekTokenIs(lexer.TOKEN_SEMICOLON) && precedence < p.getPrecedence(p.peekToken.Type) {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}
		
		p.nextToken()
		leftExp = infix(leftExp)
	}
	
	return leftExp
}

// ============= STATEMENT PARSING ARCHITECTURE =============

// parseStatement routes to appropriate statement parser based on token type
func (p *PrattParser) parseStatement() ast.Statement {
	// Handle attributes for statements
	var attributes ast.AttributeList
	if p.currentTokenIs(lexer.T_ATTRIBUTE) {
		attributes = p.parseAttributes()
	}
	
	// Route to specific statement parser
	if parser, exists := p.statementParsers[p.currentToken.Type]; exists {
		stmt := parser()
		if attributes != nil {
			// Attach attributes to statement
			if attributable, ok := stmt.(ast.AttributableStatement); ok {
				attributable.SetAttributes(attributes)
			}
		}
		return stmt
	}
	
	// Default to expression statement
	return p.parseExpressionStatement()
}

// ============= DECLARATION PARSING ARCHITECTURE =============

// parseDeclaration routes to appropriate declaration parser
func (p *PrattParser) parseDeclaration() ast.Declaration {
	// Handle attributes for declarations
	var attributes ast.AttributeList
	if p.currentTokenIs(lexer.T_ATTRIBUTE) {
		attributes = p.parseAttributes()
	}
	
	// Route to specific declaration parser
	if parser, exists := p.declarationParsers[p.currentToken.Type]; exists {
		decl := parser()
		if attributes != nil {
			// Attach attributes to declaration
			if attributable, ok := decl.(ast.AttributableDeclaration); ok {
				attributable.SetAttributes(attributes)
			}
		}
		return decl
	}
	
	p.errors = append(p.errors, fmt.Sprintf("no declaration parser for token %s", p.currentToken.Type))
	return nil
}

// ============= UTILITY METHODS =============

// nextToken advances to the next meaningful token, skipping whitespace and comments
func (p *PrattParser) nextToken() {
	p.currentToken = p.peekToken
	p.peekToken = p.lexer.NextToken()
	
	// Skip non-syntactic tokens
	for p.isNonSyntacticToken(p.peekToken.Type) {
		p.peekToken = p.lexer.NextToken()
	}
}

// Token checking utilities
func (p *PrattParser) currentTokenIs(t lexer.TokenType) bool {
	return p.currentToken.Type == t
}

func (p *PrattParser) peekTokenIs(t lexer.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *PrattParser) expectPeek(t lexer.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

func (p *PrattParser) isNonSyntacticToken(tokenType lexer.TokenType) bool {
	switch tokenType {
	case lexer.T_WHITESPACE, lexer.T_COMMENT, lexer.T_DOC_COMMENT:
		return true
	default:
		return false
	}
}

// Error handling
func (p *PrattParser) noPrefixParseFnError(t lexer.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
}

func (p *PrattParser) peekError(t lexer.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead",
		t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *PrattParser) Errors() []string {
	return p.errors
}

// Context management
func (p *PrattParser) enterContext(contextType string) {
	switch contextType {
	case "class":
		p.parsingContext.InClass = true
	case "function":
		p.parsingContext.InFunction = true
	case "interface":
		p.parsingContext.InInterface = true
	case "trait":
		p.parsingContext.InTrait = true
	case "enum":
		p.parsingContext.InEnum = true
	case "namespace":
		p.parsingContext.InNamespace = true
	case "match":
		p.parsingContext.InMatch = true
	case "property_hook":
		p.parsingContext.InPropertyHook = true
	case "arrow_function":
		p.parsingContext.InArrowFunction = true
	}
}

func (p *PrattParser) exitContext(contextType string) {
	switch contextType {
	case "class":
		p.parsingContext.InClass = false
	case "function":
		p.parsingContext.InFunction = false
	case "interface":
		p.parsingContext.InInterface = false
	case "trait":
		p.parsingContext.InTrait = false
	case "enum":
		p.parsingContext.InEnum = false
	case "namespace":
		p.parsingContext.InNamespace = false
	case "match":
		p.parsingContext.InMatch = false
	case "property_hook":
		p.parsingContext.InPropertyHook = false
	case "arrow_function":
		p.parsingContext.InArrowFunction = false
	}
}

// Main parsing entry point
func (p *PrattParser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}
	
	for !p.currentTokenIs(lexer.T_EOF) {
		// Skip PHP opening tags
		if p.currentTokenIs(lexer.T_OPEN_TAG) {
			p.nextToken()
			continue
		}
		
		// Parse top-level statement or declaration
		var stmt ast.Statement
		if p.isDeclarationToken(p.currentToken.Type) {
			decl := p.parseDeclaration()
			if decl != nil {
				stmt = &ast.DeclarationStatement{Declaration: decl}
			}
		} else {
			stmt = p.parseStatement()
		}
		
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		
		p.nextToken()
	}
	
	return program
}

// isDeclarationToken checks if the token starts a declaration
func (p *PrattParser) isDeclarationToken(tokenType lexer.TokenType) bool {
	switch tokenType {
	case lexer.T_FUNCTION, lexer.T_CLASS, lexer.T_INTERFACE, 
		 lexer.T_TRAIT, lexer.T_ENUM, lexer.T_NAMESPACE, 
		 lexer.T_USE, lexer.T_CONST:
		return true
	case lexer.T_ABSTRACT, lexer.T_FINAL, lexer.T_READONLY,
		 lexer.T_PUBLIC, lexer.T_PROTECTED, lexer.T_PRIVATE, lexer.T_STATIC:
		// These might start declarations depending on context
		return true
	default:
		return false
	}
}