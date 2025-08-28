package parser

import (
	"fmt"

	"github.com/wudi/php-parser/ast"
	"github.com/wudi/php-parser/lexer"
)

// ============= REMAINING EXPRESSION PARSERS =============

func (p *PrattParser) parseAnonymousFunction() ast.Expression {
	position := p.currentToken.Position
	
	// Parse function keyword  
	if !p.currentTokenIs(lexer.T_FUNCTION) {
		return nil
	}
	
	// Parse optional reference return
	var returnsReference bool
	if p.peekTokenIs(lexer.TOKEN_AMPERSAND) {
		returnsReference = true
		p.nextToken()
	}
	
	// Parse parameter list
	if !p.expectPeek(lexer.TOKEN_LPAREN) {
		return nil
	}
	
	parameters := p.parseParameterList()
	
	// Parse optional use clause
	var useVars []*ast.UseVariable
	if p.peekTokenIs(lexer.T_USE) {
		p.nextToken()
		if !p.expectPeek(lexer.TOKEN_LPAREN) {
			return nil
		}
		
		useVars = p.parseUseVariableList()
	}
	
	// Parse return type
	var returnType ast.Type
	if p.peekTokenIs(lexer.TOKEN_COLON) {
		p.nextToken()
		p.nextToken()
		returnType = p.parseType()
	}
	
	// Parse function body
	if !p.expectPeek(lexer.TOKEN_LBRACE) {
		return nil
	}
	
	p.nextToken()
	body := p.parseStatementList([]lexer.TokenType{lexer.TOKEN_RBRACE})
	
	if !p.expectPeek(lexer.TOKEN_RBRACE) {
		return nil
	}
	
	return &ast.AnonymousFunctionExpression{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTClosure,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Parameters:       parameters,
		UseVariables:     useVars,
		ReturnType:       returnType,
		Body:             body,
		ReturnsReference: returnsReference,
	}
}

func (p *PrattParser) parseArrowFunction() ast.Expression {
	position := p.currentToken.Position
	
	// Parse fn keyword
	if !p.currentTokenIs(lexer.T_FN) {
		return nil
	}
	
	// Parse optional reference return
	var returnsReference bool
	if p.peekTokenIs(lexer.TOKEN_AMPERSAND) {
		returnsReference = true
		p.nextToken()
	}
	
	// Parse parameter list
	if !p.expectPeek(lexer.TOKEN_LPAREN) {
		return nil
	}
	
	parameters := p.parseParameterList()
	
	// Parse return type
	var returnType ast.Type
	if p.peekTokenIs(lexer.TOKEN_COLON) {
		p.nextToken()
		p.nextToken()
		returnType = p.parseType()
	}
	
	// Parse => and expression
	if !p.expectPeek(lexer.T_DOUBLE_ARROW) {
		return nil
	}
	
	p.nextToken()
	expression := p.parseExpression(LOWEST)
	if expression == nil {
		return nil
	}
	
	return &ast.ArrowFunctionExpression{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTArrowFunc,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Parameters:       parameters,
		ReturnType:       returnType,
		Expression:       expression,
		ReturnsReference: returnsReference,
	}
}

func (p *PrattParser) parseStaticExpression() ast.Expression {
	// This handles 'static' in expression context
	// Could be static::, static function, etc.
	if p.peekTokenIs(lexer.T_PAAMAYIM_NEKUDOTAYIM) {
		// static::something - treat as identifier
		return p.parseIdentifier()
	}
	
	if p.peekTokenIs(lexer.T_FUNCTION) {
		// static function - parse as anonymous function with static
		p.nextToken()
		closure := p.parseAnonymousFunction()
		if closure != nil {
			if anon, ok := closure.(*ast.AnonymousFunctionExpression); ok {
				anon.IsStatic = true
			}
		}
		return closure
	}
	
	if p.peekTokenIs(lexer.T_FN) {
		// static fn - parse as arrow function with static
		p.nextToken()
		arrow := p.parseArrowFunction()
		if arrow != nil {
			if arrowFn, ok := arrow.(*ast.ArrowFunctionExpression); ok {
				arrowFn.IsStatic = true
			}
		}
		return arrow
	}
	
	// Just static keyword - treat as identifier
	return p.parseIdentifier()
}

// ============= TYPE CASTING =============

func (p *PrattParser) parseTypeCast() ast.Expression {
	position := p.currentToken.Position
	castType := p.currentToken.Type
	
	p.nextToken()
	expression := p.parseExpression(UNARY)
	if expression == nil {
		return nil
	}
	
	var castToType string
	switch castType {
	case lexer.T_INT_CAST:
		castToType = "int"
	case lexer.T_BOOL_CAST:
		castToType = "bool"
	case lexer.T_DOUBLE_CAST:
		castToType = "float"
	case lexer.T_STRING_CAST:
		castToType = "string"
	case lexer.T_ARRAY_CAST:
		castToType = "array"
	case lexer.T_OBJECT_CAST:
		castToType = "object"
	case lexer.T_UNSET_CAST:
		castToType = "unset"
	default:
		castToType = "unknown"
	}
	
	return &ast.CastExpression{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTCast,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Type:       castToType,
		Expression: expression,
	}
}

// ============= OBJECT CREATION =============

func (p *PrattParser) parseNewExpression() ast.Expression {
	position := p.currentToken.Position
	
	if !p.currentTokenIs(lexer.T_NEW) {
		return nil
	}
	
	p.nextToken()
	
	// Parse class name or expression
	var className ast.Expression
	switch p.currentToken.Type {
	case lexer.T_STRING:
		className = p.parseNamespaceName()
	case lexer.T_VARIABLE:
		className = p.parseVariable()
	case lexer.TOKEN_LPAREN:
		p.nextToken()
		className = p.parseExpression(LOWEST)
		if !p.expectPeek(lexer.TOKEN_RPAREN) {
			return nil
		}
	case lexer.T_CLASS:
		// Anonymous class
		return p.parseAnonymousClassExpression(position)
	default:
		p.errors = append(p.errors, fmt.Sprintf("unexpected token in new expression: %s", p.currentToken.Type))
		return nil
	}
	
	// Parse optional arguments
	var arguments []ast.Expression
	if p.peekTokenIs(lexer.TOKEN_LPAREN) {
		p.nextToken()
		arguments = p.parseArgumentList()
	}
	
	return &ast.NewExpression{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTNew,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Class: className,
		Arguments: arguments,
	}
}

func (p *PrattParser) parseAnonymousClassExpression(position lexer.Position) ast.Expression {
	// Parse class modifiers
	var modifiers []string
	if p.currentTokenIs(lexer.T_ABSTRACT) || p.currentTokenIs(lexer.T_FINAL) {
		modifiers = append(modifiers, p.currentToken.Value)
		p.nextToken()
	}
	
	// Parse class keyword
	if !p.currentTokenIs(lexer.T_CLASS) {
		return nil
	}
	
	// Parse optional constructor arguments
	var constructorArgs []ast.Expression
	if p.peekTokenIs(lexer.TOKEN_LPAREN) {
		p.nextToken()
		constructorArgs = p.parseArgumentList()
	}
	
	// Parse optional extends
	var extends ast.Expression
	if p.peekTokenIs(lexer.T_EXTENDS) {
		p.nextToken()
		p.nextToken()
		extends = p.parseExpression(LOWEST)
	}
	
	// Parse optional implements
	var implements []ast.Expression
	if p.peekTokenIs(lexer.T_IMPLEMENTS) {
		p.nextToken()
		p.nextToken()
		
		implements = append(implements, p.parseExpression(LOWEST))
		
		for p.peekTokenIs(lexer.TOKEN_COMMA) {
			p.nextToken()
			p.nextToken()
			implements = append(implements, p.parseExpression(LOWEST))
		}
	}
	
	// Parse class body
	if !p.expectPeek(lexer.TOKEN_LBRACE) {
		return nil
	}
	
	p.nextToken()
	members := p.parseClassMembers()
	
	if !p.expectPeek(lexer.TOKEN_RBRACE) {
		return nil
	}
	
	return &ast.AnonymousClassExpression{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTNew,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Modifiers:       modifiers,
		ConstructorArgs: constructorArgs,
		Extends:         extends,
		Implements:      implements,
		Members:         members,
	}
}

func (p *PrattParser) parseCloneExpression() ast.Expression {
	position := p.currentToken.Position
	
	p.nextToken()
	expression := p.parseExpression(UNARY)
	if expression == nil {
		return nil
	}
	
	return &ast.CloneExpression{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTClone,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Expression: expression,
	}
}

// ============= LANGUAGE CONSTRUCTS =============

func (p *PrattParser) parseIssetExpression() ast.Expression {
	position := p.currentToken.Position
	
	if !p.expectPeek(lexer.TOKEN_LPAREN) {
		return nil
	}
	
	p.nextToken()
	
	var variables []ast.Expression
	variables = append(variables, p.parseExpression(LOWEST))
	
	for p.peekTokenIs(lexer.TOKEN_COMMA) {
		p.nextToken()
		p.nextToken()
		variables = append(variables, p.parseExpression(LOWEST))
	}
	
	if !p.expectPeek(lexer.TOKEN_RPAREN) {
		return nil
	}
	
	return &ast.IssetExpression{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTIsset,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Variables: variables,
	}
}

func (p *PrattParser) parseEmptyExpression() ast.Expression {
	position := p.currentToken.Position
	
	if !p.expectPeek(lexer.TOKEN_LPAREN) {
		return nil
	}
	
	p.nextToken()
	expression := p.parseExpression(LOWEST)
	if expression == nil {
		return nil
	}
	
	if !p.expectPeek(lexer.TOKEN_RPAREN) {
		return nil
	}
	
	return &ast.EmptyExpression{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTEmpty,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Variable: expression,
	}
}

func (p *PrattParser) parseListExpression() ast.Expression {
	position := p.currentToken.Position
	
	if !p.expectPeek(lexer.TOKEN_LPAREN) {
		return nil
	}
	
	p.nextToken()
	elements := p.parseListElements()
	
	if !p.expectPeek(lexer.TOKEN_RPAREN) {
		return nil
	}
	
	return &ast.ListExpression{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTList,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Elements: elements,
	}
}

func (p *PrattParser) parseExitExpression() ast.Expression {
	position := p.currentToken.Position
	
	var expression ast.Expression
	if p.peekTokenIs(lexer.TOKEN_LPAREN) {
		p.nextToken()
		p.nextToken()
		
		if !p.currentTokenIs(lexer.TOKEN_RPAREN) {
			expression = p.parseExpression(LOWEST)
		}
		
		if !p.expectPeek(lexer.TOKEN_RPAREN) {
			return nil
		}
	}
	
	return &ast.ExitExpression{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTExit,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Expression: expression,
	}
}

func (p *PrattParser) parseEvalExpression() ast.Expression {
	position := p.currentToken.Position
	
	if !p.expectPeek(lexer.TOKEN_LPAREN) {
		return nil
	}
	
	p.nextToken()
	expression := p.parseExpression(LOWEST)
	if expression == nil {
		return nil
	}
	
	if !p.expectPeek(lexer.TOKEN_RPAREN) {
		return nil
	}
	
	return &ast.EvalExpression{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTIncludeOrEval,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Code: expression,
	}
}

func (p *PrattParser) parsePrintExpression() ast.Expression {
	position := p.currentToken.Position
	
	p.nextToken()
	expression := p.parseExpression(UNARY)
	if expression == nil {
		return nil
	}
	
	return &ast.PrintExpression{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTPrint,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Expression: expression,
	}
}

func (p *PrattParser) parseIncludeExpression() ast.Expression {
	position := p.currentToken.Position
	includeType := p.currentToken.Type
	
	p.nextToken()
	expression := p.parseExpression(LOWEST)
	if expression == nil {
		return nil
	}
	
	var includeTypeStr string
	switch includeType {
	case lexer.T_INCLUDE:
		includeTypeStr = "include"
	case lexer.T_INCLUDE_ONCE:
		includeTypeStr = "include_once"
	case lexer.T_REQUIRE:
		includeTypeStr = "require"
	case lexer.T_REQUIRE_ONCE:
		includeTypeStr = "require_once"
	default:
		includeTypeStr = "include"
	}
	
	return &ast.IncludeExpression{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTIncludeOrEval,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Type:       includeTypeStr,
		Expression: expression,
	}
}

// ============= GENERATOR EXPRESSIONS =============

func (p *PrattParser) parseYieldExpression() ast.Expression {
	position := p.currentToken.Position
	
	p.nextToken()
	
	var key, value ast.Expression
	
	if !p.isStatementEnd() {
		firstExpr := p.parseExpression(LOWEST)
		
		// Check if this is key => value syntax
		if p.peekTokenIs(lexer.T_DOUBLE_ARROW) {
			key = firstExpr
			p.nextToken() // consume =>
			p.nextToken()
			value = p.parseExpression(LOWEST)
		} else {
			value = firstExpr
		}
	}
	
	return &ast.YieldExpression{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTYield,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Key:   key,
		Value: value,
	}
}

func (p *PrattParser) parseYieldFromExpression() ast.Expression {
	position := p.currentToken.Position
	
	p.nextToken()
	expression := p.parseExpression(LOWEST)
	if expression == nil {
		return nil
	}
	
	return &ast.YieldFromExpression{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTYieldFrom,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Expression: expression,
	}
}

func (p *PrattParser) parseThrowExpression() ast.Expression {
	position := p.currentToken.Position
	
	p.nextToken()
	expression := p.parseExpression(LOWEST)
	if expression == nil {
		return nil
	}
	
	return &ast.ThrowExpression{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTThrowExpr,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Expression: expression,
	}
}

// ============= ARRAY EXPRESSIONS =============

func (p *PrattParser) parseArrayExpression() ast.Expression {
	position := p.currentToken.Position
	
	if !p.expectPeek(lexer.TOKEN_LPAREN) {
		return nil
	}
	
	p.nextToken()
	elements := p.parseArrayElements()
	
	if !p.expectPeek(lexer.TOKEN_RPAREN) {
		return nil
	}
	
	return &ast.ArrayExpression{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTArray,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Elements: elements,
		IsShort:   false,
	}
}

func (p *PrattParser) parseArrayLiteral() ast.Expression {
	position := p.currentToken.Position
	
	p.nextToken()
	elements := p.parseArrayElements()
	
	if !p.expectPeek(lexer.TOKEN_RBRACKET) {
		return nil
	}
	
	return &ast.ArrayExpression{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTArray,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Elements: elements,
		IsShort:   true,
	}
}

func (p *PrattParser) parseArrayElements() []*ast.ArrayElement {
	var elements []*ast.ArrayElement
	
	if p.currentTokenIs(lexer.TOKEN_RBRACKET) || p.currentTokenIs(lexer.TOKEN_RPAREN) {
		return elements
	}
	
	elements = append(elements, p.parseArrayElement())
	
	for p.peekTokenIs(lexer.TOKEN_COMMA) {
		p.nextToken() // consume comma
		p.nextToken()
		
		// Allow trailing comma
		if p.currentTokenIs(lexer.TOKEN_RBRACKET) || p.currentTokenIs(lexer.TOKEN_RPAREN) {
			break
		}
		
		elements = append(elements, p.parseArrayElement())
	}
	
	return elements
}

func (p *PrattParser) parseArrayElement() *ast.ArrayElement {
	position := p.currentToken.Position
	
	// Handle unpacking (...$array)
	if p.currentTokenIs(lexer.T_ELLIPSIS) {
		p.nextToken()
		expression := p.parseExpression(LOWEST)
		
		return &ast.ArrayElement{
			BaseNode: ast.BaseNode{
				Kind:     ast.ASTUnpack,
				Position: position,
				LineNo:   uint32(position.Line),
			},
			Value:    expression,
			IsUnpack: true,
		}
	}
	
	// Handle reference (&$var)
	var isReference bool
	if p.currentTokenIs(lexer.TOKEN_AMPERSAND) {
		isReference = true
		p.nextToken()
	}
	
	// Parse first expression
	firstExpr := p.parseExpression(LOWEST)
	
	// Check for key => value syntax
	var key, value ast.Expression
	if p.peekTokenIs(lexer.T_DOUBLE_ARROW) {
		key = firstExpr
		p.nextToken() // consume =>
		p.nextToken()
		
		// Check for reference on value
		if p.currentTokenIs(lexer.TOKEN_AMPERSAND) {
			isReference = true
			p.nextToken()
		}
		
		value = p.parseExpression(LOWEST)
	} else {
		value = firstExpr
	}
	
	return &ast.ArrayElement{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTArrayElem,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Key:         key,
		Value:       value,
		IsReference: isReference,
	}
}

// ============= STRING EXPRESSIONS =============

func (p *PrattParser) parseInterpolatedString() ast.Expression {
	position := p.currentToken.Position
	var parts []ast.Expression
	
	p.nextToken()
	
	for !p.currentTokenIs(lexer.TOKEN_QUOTE) && !p.currentTokenIs(lexer.T_EOF) {
		if p.currentTokenIs(lexer.T_ENCAPSED_AND_WHITESPACE) {
			parts = append(parts, &ast.StringLiteral{
				BaseNode: ast.BaseNode{
					Kind:     ast.ASTZval,
					Position: p.currentToken.Position,
					LineNo:   uint32(p.currentToken.Position.Line),
				},
				Value: p.currentToken.Value,
			})
		} else if p.currentTokenIs(lexer.T_VARIABLE) {
			parts = append(parts, p.parseVariable())
		} else {
			// Parse complex expression
			parts = append(parts, p.parseExpression(LOWEST))
		}
		p.nextToken()
	}
	
	return &ast.InterpolatedStringExpression{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTEncapsList,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Parts: parts,
	}
}

func (p *PrattParser) parseHeredocExpression() ast.Expression {
	position := p.currentToken.Position
	label := p.currentToken.Value
	
	var parts []ast.Expression
	
	p.nextToken()
	
	for !p.currentTokenIs(lexer.T_END_HEREDOC) && !p.currentTokenIs(lexer.T_EOF) {
		if p.currentTokenIs(lexer.T_ENCAPSED_AND_WHITESPACE) {
			parts = append(parts, &ast.StringLiteral{
				BaseNode: ast.BaseNode{
					Kind:     ast.ASTZval,
					Position: p.currentToken.Position,
					LineNo:   uint32(p.currentToken.Position.Line),
				},
				Value: p.currentToken.Value,
			})
		} else if p.currentTokenIs(lexer.T_VARIABLE) {
			parts = append(parts, p.parseVariable())
		} else {
			parts = append(parts, p.parseExpression(LOWEST))
		}
		p.nextToken()
	}
	
	return &ast.HeredocExpression{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTEncapsList,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Label: label,
		Parts: parts,
	}
}

func (p *PrattParser) parseNowdocExpression() ast.Expression {
	position := p.currentToken.Position
	content := p.currentToken.Value
	
	return &ast.NowdocExpression{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTZval,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Content: content,
	}
}

func (p *PrattParser) parseShellExecExpression() ast.Expression {
	position := p.currentToken.Position
	
	p.nextToken()
	
	var parts []ast.Expression
	
	for !p.currentTokenIs(lexer.T_EOF) { // TODO: Add TOKEN_BACKTICK when available
		if p.currentTokenIs(lexer.T_ENCAPSED_AND_WHITESPACE) {
			parts = append(parts, &ast.StringLiteral{
				BaseNode: ast.BaseNode{
					Kind:     ast.ASTZval,
					Position: p.currentToken.Position,
					LineNo:   uint32(p.currentToken.Position.Line),
				},
				Value: p.currentToken.Value,
			})
		} else {
			parts = append(parts, p.parseExpression(LOWEST))
		}
		p.nextToken()
	}
	
	return &ast.ShellExecExpression{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTShellExec,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Command: nil, // TODO: Fix command parsing from parts
	}
}

// ============= MAGIC CONSTANTS =============

func (p *PrattParser) parseMagicConstant() ast.Expression {
	position := p.currentToken.Position
	constantType := p.currentToken.Type
	
	return &ast.MagicConstantExpression{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTMagicConst,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Name: constantType,
	}
}

// ============= FUNCTION CALLS =============

func (p *PrattParser) parseFunctionCall(left ast.Expression) ast.Expression {
	position := p.currentToken.Position
	
	arguments := p.parseArgumentList()
	
	return &ast.FunctionCallExpression{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTCall,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Function:  left,
		Arguments: arguments,
	}
}

func (p *PrattParser) parseInstanceofExpression(left ast.Expression) ast.Expression {
	position := p.currentToken.Position
	precedence := p.getPrecedence(p.currentToken.Type)
	
	p.nextToken()
	right := p.parseExpression(precedence)
	if right == nil {
		return nil
	}
	
	return &ast.InstanceofExpression{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTInstanceof,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Expression: left,
		Class:      right,
	}
}

// ============= UTILITY FUNCTIONS =============

func (p *PrattParser) parseUseVariableList() []*ast.UseVariable {
	var useVars []*ast.UseVariable
	
	p.nextToken()
	
	if p.currentTokenIs(lexer.TOKEN_RPAREN) {
		return useVars
	}
	
	useVars = append(useVars, p.parseUseVariable())
	
	for p.peekTokenIs(lexer.TOKEN_COMMA) {
		p.nextToken() // consume comma
		p.nextToken()
		useVars = append(useVars, p.parseUseVariable())
	}
	
	if !p.expectPeek(lexer.TOKEN_RPAREN) {
		return nil
	}
	
	return useVars
}

func (p *PrattParser) parseUseVariable() *ast.UseVariable {
	position := p.currentToken.Position
	
	var isReference bool
	if p.currentTokenIs(lexer.TOKEN_AMPERSAND) {
		isReference = true
		p.nextToken()
	}
	
	if !p.currentTokenIs(lexer.T_VARIABLE) {
		p.errors = append(p.errors, "expected variable in use clause")
		return nil
	}
	
	variable := p.currentToken.Value
	
	return &ast.UseVariable{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTClosureUses,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Name:        variable,
		IsReference: isReference,
	}
}

func (p *PrattParser) parseListElements() []*ast.ListElement {
	var elements []*ast.ListElement
	
	for !p.currentTokenIs(lexer.TOKEN_RPAREN) && !p.currentTokenIs(lexer.T_EOF) {
		var element *ast.ListElement
		
		if p.currentTokenIs(lexer.TOKEN_COMMA) {
			// Empty element
			element = nil
		} else {
			position := p.currentToken.Position
			expr := p.parseExpression(LOWEST)
			
			element = &ast.ListElement{
				BaseNode: ast.BaseNode{
					Kind:     ast.ASTArrayElem,
					Position: position,
					LineNo:   uint32(position.Line),
				},
				Variable: expr,
			}
		}
		
		elements = append(elements, element)
		
		if p.peekTokenIs(lexer.TOKEN_COMMA) {
			p.nextToken()
			p.nextToken()
		} else {
			break
		}
	}
	
	return elements
}

func (p *PrattParser) isStatementEnd() bool {
	switch p.currentToken.Type {
	case lexer.TOKEN_SEMICOLON, lexer.TOKEN_RBRACE, lexer.T_EOF,
		 lexer.TOKEN_COMMA, lexer.TOKEN_RPAREN, lexer.TOKEN_RBRACKET:
		return true
	default:
		return false
	}
}