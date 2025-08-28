package parser

import (
	"fmt"

	"github.com/wudi/php-parser/ast"
	"github.com/wudi/php-parser/lexer"
)

// ============= MODERN PHP FEATURES PARSERS =============

// Attribute parser initialization
func (p *PrattParser) initializeAttributeParsers() {
	p.attributeParsers = map[lexer.TokenType]AttributeParseFn{
		lexer.T_ATTRIBUTE: p.parseAttributeList,
	}
}

// Type parser initialization  
func (p *PrattParser) initializeTypeParsers() {
	p.typeParsers = map[lexer.TokenType]TypeParseFn{
		lexer.T_STRING:     p.parseNamedType,
		lexer.T_ARRAY:      p.parseArrayType,
		lexer.T_CALLABLE:   p.parseCallableType,
		lexer.T_STATIC:     p.parseStaticType,
		lexer.TOKEN_QUESTION: p.parseNullableType,
	}
}

// ============= ATTRIBUTE PARSING (PHP 8.0+) =============

func (p *PrattParser) parseAttributes() ast.AttributeList {
	var attributeGroups []*ast.AttributeGroup
	
	for p.currentTokenIs(lexer.T_ATTRIBUTE) {
		group := p.parseAttributeGroup()
		if group != nil {
			attributeGroups = append(attributeGroups, group)
		}
	}
	
	return &ast.AttributeListExpression{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTAttributeList,
			Position: attributeGroups[0].GetPosition(),
			LineNo:   attributeGroups[0].GetLineNo(),
		},
		Groups: attributeGroups,
	}
}

func (p *PrattParser) parseAttributeGroup() *ast.AttributeGroup {
	position := p.currentToken.Position
	
	// Parse #[
	if !p.currentTokenIs(lexer.T_ATTRIBUTE) {
		return nil
	}
	
	p.nextToken()
	
	var attributes []*ast.AttributeExpression
	
	// Parse first attribute
	attributes = append(attributes, p.parseAttributeExpression())
	
	// Parse additional attributes separated by commas
	for p.peekTokenIs(lexer.TOKEN_COMMA) {
		p.nextToken() // consume comma
		p.nextToken()
		
		// Allow trailing comma
		if p.currentTokenIs(lexer.TOKEN_RBRACKET) {
			break
		}
		
		attributes = append(attributes, p.parseAttributeExpression())
	}
	
	// Parse ]
	if !p.expectPeek(lexer.TOKEN_RBRACKET) {
		return nil
	}
	
	return &ast.AttributeGroup{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTAttributeGroup,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Attributes: attributes,
	}
}

func (p *PrattParser) parseAttributeExpression() *ast.AttributeExpression {
	position := p.currentToken.Position
	
	// Parse attribute name (can be namespaced)
	name := p.parseNamespaceName()
	
	// Parse optional arguments
	var arguments []ast.Expression
	if p.peekTokenIs(lexer.TOKEN_LPAREN) {
		p.nextToken()
		arguments = p.parseArgumentList()
	}
	
	return &ast.AttributeExpression{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTAttribute,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Name:      name,
		Arguments: arguments,
	}
}

func (p *PrattParser) parseAttributeList() ast.AttributeList {
	return p.parseAttributes()
}

// ============= TYPE PARSING (PHP 7.0+, Enhanced in 8.0+) =============

func (p *PrattParser) parseType() ast.Type {
	p.enterContext("union_type")
	defer p.exitContext("union_type")
	
	// Handle nullable types (?Type)
	var nullable bool
	if p.currentTokenIs(lexer.TOKEN_QUESTION) {
		nullable = true
		p.nextToken()
	}
	
	// Parse the base type
	baseType := p.parseBaseType()
	if baseType == nil {
		return nil
	}
	
	// Check for union types (Type1|Type2)
	if p.peekTokenIs(lexer.TOKEN_PIPE) {
		return p.parseUnionType(baseType, nullable)
	}
	
	// Check for intersection types (Type1&Type2)  
	if p.peekTokenIs(lexer.TOKEN_AMPERSAND) {
		return p.parseIntersectionType(baseType, nullable)
	}
	
	// Simple type
	if nullable {
		return &ast.NullableType{
			BaseNode: ast.BaseNode{
				Kind:     ast.ASTType,
				Position: baseType.GetPosition(),
				LineNo:   baseType.GetLineNo(),
			},
			Type: baseType,
		}
	}
	
	return baseType
}

func (p *PrattParser) parseBaseType() ast.Type {
	switch p.currentToken.Type {
	case lexer.T_ARRAY:
		return p.parseArrayType()
	case lexer.T_CALLABLE:
		return p.parseCallableType()
	case lexer.T_STRING:
		return p.parseScalarType()
	case lexer.T_STATIC:
		return p.parseStaticType()
	default:
		// Assume it's a class name
		return p.parseNamedType()
	}
}

func (p *PrattParser) parseUnionType(firstType ast.Type, nullable bool) ast.Type {
	position := firstType.GetPosition()
	p.enterContext("union_type")
	defer p.exitContext("union_type")
	
	var types []ast.Type
	types = append(types, firstType)
	
	for p.peekTokenIs(lexer.TOKEN_PIPE) {
		p.nextToken() // consume |
		p.nextToken()
		
		// Handle parenthesized intersection types in unions
		var nextType ast.Type
		if p.currentTokenIs(lexer.TOKEN_LPAREN) {
			p.nextToken()
			nextType = p.parseIntersectionType(p.parseBaseType(), false)
			if !p.expectPeek(lexer.TOKEN_RPAREN) {
				return nil
			}
		} else {
			nextType = p.parseBaseType()
		}
		
		if nextType != nil {
			types = append(types, nextType)
		}
	}
	
	unionType := &ast.UnionType{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTTypeUnion,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Types: types,
	}
	
	if nullable {
		return &ast.NullableType{
			BaseNode: ast.BaseNode{
				Kind:     ast.ASTType,
				Position: position,
				LineNo:   uint32(position.Line),
			},
			Type: unionType,
		}
	}
	
	return unionType
}

func (p *PrattParser) parseIntersectionType(firstType ast.Type, nullable bool) ast.Type {
	position := firstType.GetPosition()
	p.enterContext("intersection_type")
	defer p.exitContext("intersection_type")
	
	var types []ast.Type
	types = append(types, firstType)
	
	for p.peekTokenIs(lexer.TOKEN_AMPERSAND) {
		p.nextToken() // consume &
		p.nextToken()
		
		nextType := p.parseBaseType()
		if nextType != nil {
			types = append(types, nextType)
		}
	}
	
	intersectionType := &ast.IntersectionType{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTTypeIntersection,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Types: types,
	}
	
	if nullable {
		return &ast.NullableType{
			BaseNode: ast.BaseNode{
				Kind:     ast.ASTType,
				Position: position,
				LineNo:   uint32(position.Line),
			},
			Type: intersectionType,
		}
	}
	
	return intersectionType
}

func (p *PrattParser) parseNamedType() ast.Type {
	position := p.currentToken.Position
	name := p.parseNamespaceName()
	
	return &ast.NamedType{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTType,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Name: name,
	}
}

func (p *PrattParser) parseArrayType() ast.Type {
	position := p.currentToken.Position
	
	return &ast.ArrayType{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTType,
			Position: position,
			LineNo:   uint32(position.Line),
		},
	}
}

func (p *PrattParser) parseCallableType() ast.Type {
	position := p.currentToken.Position
	
	return &ast.CallableType{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTType,
			Position: position,
			LineNo:   uint32(position.Line),
		},
	}
}

func (p *PrattParser) parseScalarType() ast.Type {
	position := p.currentToken.Position
	typeName := p.currentToken.Value
	
	return &ast.ScalarType{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTType,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Type: typeName,
	}
}

func (p *PrattParser) parseStaticType() ast.Type {
	position := p.currentToken.Position
	
	return &ast.StaticType{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTType,
			Position: position,
			LineNo:   uint32(position.Line),
		},
	}
}

func (p *PrattParser) parseNullableType() ast.Type {
	position := p.currentToken.Position
	p.nextToken()
	
	baseType := p.parseBaseType()
	if baseType == nil {
		return nil
	}
	
	return &ast.NullableType{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTType,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Type: baseType,
	}
}

// ============= MATCH EXPRESSION PARSING (PHP 8.0+) =============

func (p *PrattParser) parseMatchExpression() ast.Expression {
	position := p.currentToken.Position
	p.enterContext("match")
	defer p.exitContext("match")
	
	// Parse match keyword
	if !p.currentTokenIs(lexer.T_MATCH) {
		return nil
	}
	
	// Parse (expression)
	if !p.expectPeek(lexer.TOKEN_LPAREN) {
		return nil
	}
	
	p.nextToken()
	condition := p.parseExpression(LOWEST)
	if condition == nil {
		return nil
	}
	
	if !p.expectPeek(lexer.TOKEN_RPAREN) {
		return nil
	}
	
	// Parse { match arms }
	if !p.expectPeek(lexer.TOKEN_LBRACE) {
		return nil
	}
	
	p.nextToken()
	arms := p.parseMatchArms()
	
	if !p.expectPeek(lexer.TOKEN_RBRACE) {
		return nil
	}
	
	return &ast.MatchExpression{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTMatch,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Condition: condition,
		Arms:      arms,
	}
}

func (p *PrattParser) parseMatchArms() []*ast.MatchArm {
	var arms []*ast.MatchArm
	
	for !p.currentTokenIs(lexer.TOKEN_RBRACE) && !p.currentTokenIs(lexer.T_EOF) {
		arm := p.parseMatchArm()
		if arm != nil {
			arms = append(arms, arm)
		}
		
		// Expect comma after each arm (except possibly the last)
		if p.peekTokenIs(lexer.TOKEN_COMMA) {
			p.nextToken()
		}
		
		p.nextToken()
		
		// Allow trailing comma
		if p.currentTokenIs(lexer.TOKEN_RBRACE) {
			break
		}
	}
	
	return arms
}

func (p *PrattParser) parseMatchArm() *ast.MatchArm {
	position := p.currentToken.Position
	
	var conditions []ast.Expression
	
	// Parse default arm
	if p.currentTokenIs(lexer.T_DEFAULT) {
		// Allow optional comma after default
		if p.peekTokenIs(lexer.TOKEN_COMMA) {
			p.nextToken()
		}
		
		if !p.expectPeek(lexer.T_DOUBLE_ARROW) {
			return nil
		}
		
		p.nextToken()
		expression := p.parseExpression(LOWEST)
		if expression == nil {
			return nil
		}
		
		return &ast.MatchArm{
			BaseNode: ast.BaseNode{
				Kind:     ast.ASTMatchArm,
				Position: position,
				LineNo:   uint32(position.Line),
			},
			Conditions: nil, // nil indicates default arm
			Expression: expression,
		}
	}
	
	// Parse condition list (can be multiple conditions separated by commas)
	conditions = append(conditions, p.parseExpression(LOWEST))
	
	for p.peekTokenIs(lexer.TOKEN_COMMA) {
		// Peek ahead to see if this comma is followed by => (end of conditions)
		// or by another expression (more conditions)
		if p.peekNextTokenIs(lexer.T_DOUBLE_ARROW) {
			break
		}
		
		p.nextToken() // consume comma
		p.nextToken()
		conditions = append(conditions, p.parseExpression(LOWEST))
	}
	
	// Allow optional comma before =>
	if p.peekTokenIs(lexer.TOKEN_COMMA) {
		p.nextToken()
	}
	
	if !p.expectPeek(lexer.T_DOUBLE_ARROW) {
		return nil
	}
	
	p.nextToken()
	expression := p.parseExpression(LOWEST)
	if expression == nil {
		return nil
	}
	
	return &ast.MatchArm{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTMatchArm,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Conditions: conditions,
		Expression: expression,
	}
}

// ============= ENUM PARSING (PHP 8.1+) =============

func (p *PrattParser) parseEnumMembers() []ast.ClassMember {
	var members []ast.ClassMember
	
	for !p.currentTokenIs(lexer.TOKEN_RBRACE) && !p.currentTokenIs(lexer.T_EOF) {
		// Parse attributes first
		var attributes ast.AttributeList
		if p.currentTokenIs(lexer.T_ATTRIBUTE) {
			attributes = p.parseAttributes()
		}
		
		var member ast.ClassMember
		switch p.currentToken.Type {
		case lexer.T_CASE:
			member = p.parseEnumCase()
		default:
			// Regular class member (method, constant, etc.)
			member = p.parseClassMember()
		}
		
		if member != nil {
			if attributes != nil {
				if attributable, ok := member.(ast.AttributableClassMember); ok {
					attributable.SetAttributes(attributes)
				}
			}
			members = append(members, member)
		}
		
		p.nextToken()
	}
	
	return members
}

func (p *PrattParser) parseEnumCase() ast.ClassMember {
	position := p.currentToken.Position
	
	// Parse case keyword
	if !p.currentTokenIs(lexer.T_CASE) {
		return nil
	}
	
	// Parse case name
	if !p.expectPeek(lexer.T_STRING) {
		return nil
	}
	
	caseName := p.currentToken.Value
	
	// Parse optional value for backed enums
	var value ast.Expression
	if p.peekTokenIs(lexer.TOKEN_EQUAL) {
		p.nextToken() // consume =
		p.nextToken()
		value = p.parseExpression(LOWEST)
	}
	
	if !p.expectPeek(lexer.TOKEN_SEMICOLON) {
		return nil
	}
	
	return &ast.EnumCase{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTEnumCase,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Name:  caseName,
		Value: value,
	}
}

// ============= PROPERTY HOOKS PARSING (PHP 8.4+) =============

func (p *PrattParser) parsePropertyDeclaration(modifiers []string) ast.ClassMember {
	position := p.currentToken.Position
	
	// Parse optional type
	var propertyType ast.Type
	if p.isTypeToken(p.currentToken.Type) {
		propertyType = p.parseType()
		p.nextToken()
	}
	
	// Parse property name
	if !p.currentTokenIs(lexer.T_VARIABLE) {
		p.errors = append(p.errors, "expected property name")
		return nil
	}
	
	propertyName := p.currentToken.Value
	
	// Parse optional default value
	var defaultValue ast.Expression
	if p.peekTokenIs(lexer.TOKEN_EQUAL) {
		p.nextToken() // consume =
		p.nextToken()
		defaultValue = p.parseExpression(LOWEST)
	}
	
	// Check for property hooks
	var hooks []*ast.PropertyHook
	if p.peekTokenIs(lexer.TOKEN_LBRACE) {
		p.nextToken()
		p.nextToken()
		hooks = p.parsePropertyHooks()
		
		if !p.expectPeek(lexer.TOKEN_RBRACE) {
			return nil
		}
	} else {
		// Regular property - expect semicolon
		if !p.expectPeek(lexer.TOKEN_SEMICOLON) {
			return nil
		}
	}
	
	return &ast.PropertyDeclaration{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTPropDecl,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Name:         propertyName,
		Type:         propertyType,
		DefaultValue: defaultValue,
		Modifiers:    modifiers,
		Hooks:        hooks,
	}
}

func (p *PrattParser) parsePropertyHooks() []*ast.PropertyHook {
	p.enterContext("property_hook")
	defer p.exitContext("property_hook")
	
	var hooks []*ast.PropertyHook
	
	for !p.currentTokenIs(lexer.TOKEN_RBRACE) && !p.currentTokenIs(lexer.T_EOF) {
		// Parse attributes
		var attributes ast.AttributeList
		if p.currentTokenIs(lexer.T_ATTRIBUTE) {
			attributes = p.parseAttributes()
		}
		
		hook := p.parsePropertyHook()
		if hook != nil {
			if attributes != nil {
				hook.SetAttributes(attributes)
			}
			hooks = append(hooks, hook)
		}
		
		p.nextToken()
	}
	
	return hooks
}

func (p *PrattParser) parsePropertyHook() *ast.PropertyHook {
	position := p.currentToken.Position
	
	// Parse hook modifiers
	var modifiers []string
	for p.isVisibilityOrModifier(p.currentToken.Type) {
		modifiers = append(modifiers, p.currentToken.Value)
		p.nextToken()
	}
	
	// Parse optional reference return
	var returnsReference bool
	if p.currentTokenIs(lexer.TOKEN_AMPERSAND) {
		returnsReference = true
		p.nextToken()
	}
	
	// Parse hook name (get, set)
	if !p.currentTokenIs(lexer.T_STRING) {
		p.errors = append(p.errors, "expected hook name (get or set)")
		return nil
	}
	
	hookName := p.currentToken.Value
	if hookName != "get" && hookName != "set" {
		p.errors = append(p.errors, fmt.Sprintf("invalid hook name: %s (expected 'get' or 'set')", hookName))
		return nil
	}
	
	// Parse optional parameters (only for set hook)
	var parameters []*ast.Parameter
	if p.peekTokenIs(lexer.TOKEN_LPAREN) {
		p.nextToken()
		parameters = p.parseParameterList()
	}
	
	// Parse hook body
	var body ast.Statement
	var expression ast.Expression
	
	if p.peekTokenIs(lexer.T_DOUBLE_ARROW) {
		// Short syntax: => expression;
		p.nextToken() // consume =>
		p.nextToken()
		expression = p.parseExpression(LOWEST)
		
		if !p.expectPeek(lexer.TOKEN_SEMICOLON) {
			return nil
		}
	} else if p.peekTokenIs(lexer.TOKEN_LBRACE) {
		// Block syntax: { statements }
		p.nextToken()
		p.nextToken()
		body = p.parseStatementList([]lexer.TokenType{lexer.TOKEN_RBRACE})
		
		if !p.expectPeek(lexer.TOKEN_RBRACE) {
			return nil
		}
	} else if p.peekTokenIs(lexer.TOKEN_SEMICOLON) {
		// Abstract hook
		p.nextToken()
	} else {
		p.errors = append(p.errors, "expected hook body, => expression, or ;")
		return nil
	}
	
	return &ast.PropertyHook{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTPropertyHook,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Name:             hookName,
		Parameters:       parameters,
		Body:             body,
		Expression:       expression,
		Modifiers:        modifiers,
		ReturnsReference: returnsReference,
	}
}

// ============= PIPE OPERATOR (PHP 8.4+) =============

func (p *PrattParser) parsePipeExpression(left ast.Expression) ast.Expression {
	position := p.currentToken.Position
	precedence := p.getPrecedence(p.currentToken.Type)
	
	p.nextToken()
	right := p.parseExpression(precedence)
	if right == nil {
		return nil
	}
	
	return &ast.PipeExpression{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTPipe,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Left:  left,
		Right: right,
	}
}

// ============= UTILITY FUNCTIONS =============

func (p *PrattParser) parseArgumentList() []ast.Expression {
	var arguments []ast.Expression
	
	p.nextToken()
	
	if p.currentTokenIs(lexer.TOKEN_RPAREN) {
		return arguments
	}
	
	arguments = append(arguments, p.parseArgument())
	
	for p.peekTokenIs(lexer.TOKEN_COMMA) {
		p.nextToken() // consume comma
		p.nextToken()
		arguments = append(arguments, p.parseArgument())
	}
	
	if !p.expectPeek(lexer.TOKEN_RPAREN) {
		return nil
	}
	
	return arguments
}

func (p *PrattParser) parseArgument() ast.Expression {
	// Handle named arguments (PHP 8.0+)
	if p.currentTokenIs(lexer.T_STRING) && p.peekTokenIs(lexer.TOKEN_COLON) {
		return p.parseNamedArgument()
	}
	
	// Handle unpacking (...$args)
	if p.currentTokenIs(lexer.T_ELLIPSIS) {
		return p.parseUnpackArgument()
	}
	
	// Regular argument
	return p.parseExpression(LOWEST)
}

func (p *PrattParser) parseNamedArgument() ast.Expression {
	position := p.currentToken.Position
	name := p.currentToken.Value
	
	p.nextToken() // consume name
	p.nextToken() // consume :
	
	value := p.parseExpression(LOWEST)
	if value == nil {
		return nil
	}
	
	return &ast.NamedArgument{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTNamedArg,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Name:  name,
		Value: value,
	}
}

func (p *PrattParser) parseUnpackArgument() ast.Expression {
	position := p.currentToken.Position
	
	p.nextToken() // consume ...
	expression := p.parseExpression(LOWEST)
	if expression == nil {
		return nil
	}
	
	return &ast.UnpackExpression{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTUnpack,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Expression: expression,
	}
}

// peekNextTokenIs looks two tokens ahead
func (p *PrattParser) peekNextTokenIs(t lexer.TokenType) bool {
	// TODO: Implement proper 2-token lookahead
	// For now, just return false as a fallback
	return false
}