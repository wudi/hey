package parser

import (
	"fmt"
	"strconv"

	"github.com/wudi/php-parser/ast"
	"github.com/wudi/php-parser/lexer"
)

// ============= PREFIX PARSER INITIALIZATION =============

func (p *PrattParser) initializePrefixParsers() {
	p.prefixParseFns = map[lexer.TokenType]PrefixParseFn{
		// Literals and identifiers
		lexer.T_LNUMBER:                  p.parseIntegerLiteral,
		lexer.T_DNUMBER:                  p.parseFloatLiteral,
		lexer.T_CONSTANT_ENCAPSED_STRING: p.parseStringLiteral,
		lexer.T_STRING:                   p.parseIdentifier,
		lexer.T_VARIABLE:                 p.parseVariable,
		
		// Unary operators
		lexer.TOKEN_EXCLAMATION:          p.parsePrefixExpression,
		lexer.TOKEN_MINUS:                p.parsePrefixExpression,
		lexer.TOKEN_PLUS:                 p.parsePrefixExpression,
		lexer.TOKEN_TILDE:                p.parsePrefixExpression,
		lexer.T_INC:                      p.parsePrefixExpression,
		lexer.T_DEC:                      p.parsePrefixExpression,
		lexer.TOKEN_AT:                   p.parseErrorSuppression,
		lexer.TOKEN_AMPERSAND:            p.parseReference,
		
		// Grouping and arrays
		lexer.TOKEN_LPAREN:               p.parseGroupedExpression,
		lexer.TOKEN_LBRACKET:             p.parseArrayLiteral,
		lexer.T_ARRAY:                    p.parseArrayExpression,
		
		// Type casting
		lexer.T_INT_CAST:                 p.parseTypeCast,
		lexer.T_BOOL_CAST:                p.parseTypeCast,
		lexer.T_DOUBLE_CAST:              p.parseTypeCast,
		lexer.T_STRING_CAST:              p.parseTypeCast,
		lexer.T_ARRAY_CAST:               p.parseTypeCast,
		lexer.T_OBJECT_CAST:              p.parseTypeCast,
		lexer.T_UNSET_CAST:               p.parseTypeCast,
		
		// Object creation and cloning
		lexer.T_NEW:                      p.parseNewExpression,
		lexer.T_CLONE:                    p.parseCloneExpression,
		
		// Functions and closures
		lexer.T_FUNCTION:                 p.parseAnonymousFunction,
		lexer.T_FN:                       p.parseArrowFunction,
		lexer.T_STATIC:                   p.parseStaticExpression,
		
		// Language constructs
		lexer.T_ISSET:                    p.parseIssetExpression,
		lexer.T_EMPTY:                    p.parseEmptyExpression,
		lexer.T_LIST:                     p.parseListExpression,
		lexer.T_EXIT:                     p.parseExitExpression,
		lexer.T_EVAL:                     p.parseEvalExpression,
		lexer.T_PRINT:                    p.parsePrintExpression,
		
		// Include/require
		lexer.T_INCLUDE:                  p.parseIncludeExpression,
		lexer.T_INCLUDE_ONCE:             p.parseIncludeExpression,
		lexer.T_REQUIRE:                  p.parseIncludeExpression,
		lexer.T_REQUIRE_ONCE:             p.parseIncludeExpression,
		
		// Generators and async
		lexer.T_YIELD:                    p.parseYieldExpression,
		lexer.T_YIELD_FROM:               p.parseYieldFromExpression,
		
		// PHP 8.0+ features
		lexer.T_MATCH:                    p.parseMatchExpression,
		lexer.T_THROW:                    p.parseThrowExpression,
		lexer.T_ATTRIBUTE:                func() ast.Expression { return p.parseAttributeExpression() },
		
		// String interpolation
		lexer.TOKEN_QUOTE:                p.parseInterpolatedString,
		lexer.T_START_HEREDOC:            p.parseHeredocExpression,
		lexer.T_NOWDOC:                   p.parseNowdocExpression,
		
		// Constants and magic constants
		lexer.T_LINE:                     p.parseMagicConstant,
		lexer.T_FILE:                     p.parseMagicConstant,
		lexer.T_DIR:                      p.parseMagicConstant,
		lexer.T_CLASS_C:                  p.parseMagicConstant,
		lexer.T_METHOD_C:                 p.parseMagicConstant,
		lexer.T_TRAIT_C:                  p.parseMagicConstant,
	}
}

// ============= INFIX PARSER INITIALIZATION =============

func (p *PrattParser) initializeInfixParsers() {
	p.infixParseFns = map[lexer.TokenType]InfixParseFn{
		// Arithmetic operators
		lexer.TOKEN_PLUS:                 p.parseInfixExpression,
		lexer.TOKEN_MINUS:                p.parseInfixExpression,
		lexer.TOKEN_MULTIPLY:             p.parseInfixExpression,
		lexer.TOKEN_DIVIDE:               p.parseInfixExpression,
		lexer.TOKEN_MODULO:               p.parseInfixExpression,
		lexer.T_POW:                      p.parseInfixExpression,
		
		// String concatenation
		lexer.TOKEN_DOT:                  p.parseInfixExpression,
		
		// Comparison operators
		lexer.T_IS_EQUAL:                 p.parseInfixExpression,
		lexer.T_IS_NOT_EQUAL:             p.parseInfixExpression,
		lexer.T_IS_IDENTICAL:             p.parseInfixExpression,
		lexer.T_IS_NOT_IDENTICAL:         p.parseInfixExpression,
		lexer.TOKEN_LT:                   p.parseInfixExpression,
		lexer.TOKEN_GT:                   p.parseInfixExpression,
		lexer.T_IS_SMALLER_OR_EQUAL:      p.parseInfixExpression,
		lexer.T_IS_GREATER_OR_EQUAL:      p.parseInfixExpression,
		lexer.T_SPACESHIP:                p.parseInfixExpression,
		
		// Logical operators
		lexer.T_BOOLEAN_AND:              p.parseInfixExpression,
		lexer.T_BOOLEAN_OR:               p.parseInfixExpression,
		lexer.T_LOGICAL_AND:              p.parseInfixExpression,
		lexer.T_LOGICAL_OR:               p.parseInfixExpression,
		lexer.T_LOGICAL_XOR:              p.parseInfixExpression,
		
		// Bitwise operators
		lexer.TOKEN_AMPERSAND:            p.parseInfixExpression,
		lexer.TOKEN_PIPE:                 p.parseInfixExpression,
		lexer.TOKEN_CARET:                p.parseInfixExpression,
		lexer.T_SL:                       p.parseInfixExpression,
		lexer.T_SR:                       p.parseInfixExpression,
		
		// Assignment operators
		lexer.TOKEN_EQUAL:                p.parseAssignmentExpression,
		lexer.T_PLUS_EQUAL:               p.parseAssignmentExpression,
		lexer.T_MINUS_EQUAL:              p.parseAssignmentExpression,
		lexer.T_MUL_EQUAL:                p.parseAssignmentExpression,
		lexer.T_DIV_EQUAL:                p.parseAssignmentExpression,
		lexer.T_CONCAT_EQUAL:             p.parseAssignmentExpression,
		lexer.T_MOD_EQUAL:                p.parseAssignmentExpression,
		lexer.T_AND_EQUAL:                p.parseAssignmentExpression,
		lexer.T_OR_EQUAL:                 p.parseAssignmentExpression,
		lexer.T_XOR_EQUAL:                p.parseAssignmentExpression,
		lexer.T_SL_EQUAL:                 p.parseAssignmentExpression,
		lexer.T_SR_EQUAL:                 p.parseAssignmentExpression,
		lexer.T_POW_EQUAL:                p.parseAssignmentExpression,
		lexer.T_COALESCE_EQUAL:           p.parseAssignmentExpression,
		
		// Ternary and coalescing
		lexer.TOKEN_QUESTION:             p.parseTernaryExpression,
		lexer.T_COALESCE:                 p.parseCoalescingExpression,
		
		// Postfix operators
		lexer.T_INC:                      p.parsePostfixExpression,
		lexer.T_DEC:                      p.parsePostfixExpression,
		
		// Member access
		lexer.T_OBJECT_OPERATOR:          p.parseMemberAccess,
		lexer.T_NULLSAFE_OBJECT_OPERATOR: p.parseNullsafeMemberAccess,
		lexer.T_PAAMAYIM_NEKUDOTAYIM:     p.parseStaticMemberAccess,
		lexer.TOKEN_LBRACKET:             p.parseArrayAccess,
		lexer.TOKEN_LPAREN:               p.parseFunctionCall,
		
		// Special operators
		lexer.T_INSTANCEOF:               p.parseInstanceofExpression,
		
		// PHP 8.4+ Pipe operator
		lexer.T_PIPE:                     p.parsePipeExpression,
	}
}

// ============= LITERAL PARSERS =============

func (p *PrattParser) parseIntegerLiteral() ast.Expression {
	value, err := strconv.ParseInt(p.currentToken.Value, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.currentToken.Value)
		p.errors = append(p.errors, msg)
		return nil
	}
	
	return &ast.IntegerLiteral{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTZval,
			Position: p.currentToken.Position,
			LineNo:   uint32(p.currentToken.Position.Line),
		},
		Value: value,
	}
}

func (p *PrattParser) parseFloatLiteral() ast.Expression {
	value, err := strconv.ParseFloat(p.currentToken.Value, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as float", p.currentToken.Value)
		p.errors = append(p.errors, msg)
		return nil
	}
	
	return &ast.FloatLiteral{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTZval,
			Position: p.currentToken.Position,
			LineNo:   uint32(p.currentToken.Position.Line),
		},
		Value: value,
	}
}

func (p *PrattParser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTZval,
			Position: p.currentToken.Position,
			LineNo:   uint32(p.currentToken.Position.Line),
		},
		Value: p.currentToken.Value,
	}
}

func (p *PrattParser) parseIdentifier() ast.Expression {
	return &ast.IdentifierNode{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTVar,
			Position: p.currentToken.Position,
			LineNo:   uint32(p.currentToken.Position.Line),
		},
		Value: p.currentToken.Value,
	}
}

func (p *PrattParser) parseVariable() ast.Expression {
	return &ast.Variable{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTVar,
			Position: p.currentToken.Position,
			LineNo:   uint32(p.currentToken.Position.Line),
		},
		Name: p.currentToken.Value,
	}
}

// ============= UNARY EXPRESSION PARSERS =============

func (p *PrattParser) parsePrefixExpression() ast.Expression {
	operator := p.currentToken.Value
	operatorToken := p.currentToken.Type
	
	p.nextToken()
	
	right := p.parseExpression(UNARY)
	if right == nil {
		return nil
	}
	
	// Map token type to AST kind
	var kind ast.ASTKind
	switch operatorToken {
	case lexer.TOKEN_EXCLAMATION:
		kind = ast.ASTUnaryOp
	case lexer.TOKEN_MINUS:
		kind = ast.ASTUnaryMinus
	case lexer.TOKEN_PLUS:
		kind = ast.ASTUnaryPlus
	case lexer.TOKEN_TILDE:
		kind = ast.ASTUnaryOp
	case lexer.T_INC:
		kind = ast.ASTPreInc
	case lexer.T_DEC:
		kind = ast.ASTPreDec
	default:
		kind = ast.ASTUnaryOp
	}
	
	return &ast.UnaryExpression{
		BaseNode: ast.BaseNode{
			Kind:     kind,
			Position: p.currentToken.Position,
			LineNo:   uint32(p.currentToken.Position.Line),
		},
		Operator: operator,
		Right:    right,
	}
}

func (p *PrattParser) parsePostfixExpression(left ast.Expression) ast.Expression {
	operator := p.currentToken.Value
	operatorToken := p.currentToken.Type
	
	var kind ast.ASTKind
	switch operatorToken {
	case lexer.T_INC:
		kind = ast.ASTPostInc
	case lexer.T_DEC:
		kind = ast.ASTPostDec
	default:
		kind = ast.ASTPostInc // fallback
	}
	
	return &ast.PostfixExpression{
		BaseNode: ast.BaseNode{
			Kind:     kind,
			Position: p.currentToken.Position,
			LineNo:   uint32(p.currentToken.Position.Line),
		},
		Left:     left,
		Operator: operator,
	}
}

func (p *PrattParser) parseErrorSuppression() ast.Expression {
	position := p.currentToken.Position
	
	p.nextToken()
	expression := p.parseExpression(UNARY)
	if expression == nil {
		return nil
	}
	
	return &ast.ErrorSuppressionExpression{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTSilence,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Expression: expression,
	}
}

func (p *PrattParser) parseReference() ast.Expression {
	position := p.currentToken.Position
	
	p.nextToken()
	expression := p.parseExpression(UNARY)
	if expression == nil {
		return nil
	}
	
	return &ast.ReferenceExpression{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTRef,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Expression: expression,
	}
}

// ============= BINARY EXPRESSION PARSERS =============

func (p *PrattParser) parseInfixExpression(left ast.Expression) ast.Expression {
	operator := p.currentToken.Value
	operatorPosition := p.currentToken.Position
	precedence := p.getPrecedence(p.currentToken.Type)
	
	p.nextToken()
	
	right := p.parseExpression(precedence)
	if right == nil {
		return nil
	}
	
	return &ast.BinaryExpression{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTBinaryOp,
			Position: operatorPosition,
			LineNo:   uint32(operatorPosition.Line),
		},
		Left:     left,
		Operator: operator,
		Right:    right,
	}
}

func (p *PrattParser) parseAssignmentExpression(left ast.Expression) ast.Expression {
	operator := p.currentToken.Value
	operatorPosition := p.currentToken.Position
	operatorToken := p.currentToken.Type
	
	p.nextToken()
	
	// Handle reference assignment
	var isRef bool
	if p.currentTokenIs(lexer.TOKEN_AMPERSAND) {
		isRef = true
		p.nextToken()
	}
	
	right := p.parseExpression(ASSIGNMENT)
	if right == nil {
		return nil
	}
	
	// Determine AST kind based on operator
	var kind ast.ASTKind
	if operatorToken == lexer.TOKEN_EQUAL {
		if isRef {
			kind = ast.ASTAssignRef
		} else {
			kind = ast.ASTAssign
		}
	} else {
		kind = ast.ASTAssignOp
	}
	
	return &ast.AssignmentExpression{
		BaseNode: ast.BaseNode{
			Kind:     kind,
			Position: operatorPosition,
			LineNo:   uint32(operatorPosition.Line),
		},
		Left:        left,
		Operator:    operator,
		Right:       right,
		IsReference: isRef,
	}
}

// ============= TERNARY AND COALESCING OPERATORS =============

func (p *PrattParser) parseTernaryExpression(left ast.Expression) ast.Expression {
	position := p.currentToken.Position
	
	p.nextToken()
	
	// Handle short ternary (?:) 
	var trueExpression ast.Expression
	if p.currentTokenIs(lexer.TOKEN_COLON) {
		trueExpression = nil
	} else {
		trueExpression = p.parseExpression(LOWEST)
	}
	
	if !p.expectPeek(lexer.TOKEN_COLON) {
		return nil
	}
	
	p.nextToken()
	falseExpression := p.parseExpression(TERNARY)
	if falseExpression == nil {
		return nil
	}
	
	return &ast.TernaryExpression{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTConditional,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Condition: left,
		TrueExp:   trueExpression,
		FalseExp:  falseExpression,
	}
}

func (p *PrattParser) parseCoalescingExpression(left ast.Expression) ast.Expression {
	position := p.currentToken.Position
	precedence := p.getPrecedence(p.currentToken.Type)
	
	p.nextToken()
	right := p.parseExpression(precedence)
	if right == nil {
		return nil
	}
	
	return &ast.CoalescingExpression{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTCoalesce,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Left:  left,
		Right: right,
	}
}

// ============= MEMBER ACCESS PARSERS =============

func (p *PrattParser) parseMemberAccess(left ast.Expression) ast.Expression {
	position := p.currentToken.Position
	
	p.nextToken()
	
	// Parse property name (can be identifier, variable, or expression)
	var property ast.Expression
	switch p.currentToken.Type {
	case lexer.T_STRING:
		property = p.parseIdentifier()
	case lexer.T_VARIABLE:
		property = p.parseVariable()
	case lexer.TOKEN_LBRACE:
		p.nextToken()
		property = p.parseExpression(LOWEST)
		if !p.expectPeek(lexer.TOKEN_RBRACE) {
			return nil
		}
	default:
		p.errors = append(p.errors, fmt.Sprintf("unexpected token in member access: %s", p.currentToken.Type))
		return nil
	}
	
	return &ast.MemberAccessExpression{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTProp,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Object:   left,
		Property: property,
	}
}

func (p *PrattParser) parseNullsafeMemberAccess(left ast.Expression) ast.Expression {
	position := p.currentToken.Position
	
	p.nextToken()
	
	var property ast.Expression
	switch p.currentToken.Type {
	case lexer.T_STRING:
		property = p.parseIdentifier()
	case lexer.T_VARIABLE:
		property = p.parseVariable()
	case lexer.TOKEN_LBRACE:
		p.nextToken()
		property = p.parseExpression(LOWEST)
		if !p.expectPeek(lexer.TOKEN_RBRACE) {
			return nil
		}
	default:
		p.errors = append(p.errors, fmt.Sprintf("unexpected token in nullsafe member access: %s", p.currentToken.Type))
		return nil
	}
	
	return &ast.NullsafeMemberAccessExpression{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTNullsafeProp,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Object:   left,
		Property: property,
	}
}

func (p *PrattParser) parseStaticMemberAccess(left ast.Expression) ast.Expression {
	position := p.currentToken.Position
	
	p.nextToken()
	
	var member ast.Expression
	switch p.currentToken.Type {
	case lexer.T_STRING:
		member = p.parseIdentifier()
	case lexer.T_VARIABLE:
		member = p.parseVariable()
	case lexer.TOKEN_LBRACE:
		p.nextToken()
		member = p.parseExpression(LOWEST)
		if !p.expectPeek(lexer.TOKEN_RBRACE) {
			return nil
		}
	default:
		p.errors = append(p.errors, fmt.Sprintf("unexpected token in static member access: %s", p.currentToken.Type))
		return nil
	}
	
	return &ast.StaticMemberAccessExpression{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTStaticProp,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Class:  left,
		Member: member,
	}
}

// ============= ARRAY ACCESS PARSER =============

func (p *PrattParser) parseArrayAccess(left ast.Expression) ast.Expression {
	position := p.currentToken.Position
	
	p.nextToken()
	
	var index ast.Expression
	if !p.currentTokenIs(lexer.TOKEN_RBRACKET) {
		index = p.parseExpression(LOWEST)
	}
	
	if !p.expectPeek(lexer.TOKEN_RBRACKET) {
		return nil
	}
	
	return &ast.ArrayAccessExpression{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTDim,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Array: left,
		Index: index,
	}
}

// ============= GROUPING PARSER =============

func (p *PrattParser) parseGroupedExpression() ast.Expression {
	p.nextToken()
	
	exp := p.parseExpression(LOWEST)
	
	if !p.expectPeek(lexer.TOKEN_RPAREN) {
		return nil
	}
	
	return exp
}