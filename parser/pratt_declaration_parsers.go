package parser

import (
	"fmt"

	"github.com/wudi/php-parser/ast"
	"github.com/wudi/php-parser/lexer"
)

// ============= DECLARATION PARSER INITIALIZATION =============

func (p *PrattParser) initializeDeclarationParsers() {
	p.declarationParsers = map[lexer.TokenType]DeclarationParseFn{
		// Function and method declarations
		lexer.T_FUNCTION: p.parseFunctionDeclaration,
		lexer.T_FN:       p.parseArrowFunctionDeclaration,
		
		// Class-related declarations
		lexer.T_CLASS:     p.parseClassDeclaration,
		lexer.T_INTERFACE: p.parseInterfaceDeclaration,
		lexer.T_TRAIT:     p.parseTraitDeclaration,
		lexer.T_ENUM:      p.parseEnumDeclaration,
		
		// Namespace and use declarations
		lexer.T_NAMESPACE: p.parseNamespaceDeclaration,
		lexer.T_USE:       p.parseUseDeclaration,
		
		// Constant declarations
		lexer.T_CONST: p.parseConstantDeclaration,
		
		// Visibility modifiers (might start various declarations)
		lexer.T_PUBLIC:    p.parseModifiedDeclaration,
		lexer.T_PROTECTED: p.parseModifiedDeclaration,
		lexer.T_PRIVATE:   p.parseModifiedDeclaration,
		lexer.T_STATIC:    p.parseStaticDeclaration,
		lexer.T_ABSTRACT:  p.parseAbstractDeclaration,
		lexer.T_FINAL:     p.parseFinalDeclaration,
		lexer.T_READONLY:  p.parseReadonlyDeclaration,
	}
}

// ============= FUNCTION DECLARATIONS =============

func (p *PrattParser) parseFunctionDeclaration() ast.Declaration {
	position := p.currentToken.Position
	p.enterContext("function")
	defer p.exitContext("function")
	
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
	
	// Parse function name
	if !p.expectPeek(lexer.T_STRING) {
		return nil
	}
	
	functionName := p.currentToken.Value
	p.parsingContext.FunctionName = functionName
	
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
	
	// Parse function body
	if !p.expectPeek(lexer.TOKEN_LBRACE) {
		return nil
	}
	
	p.nextToken()
	body := p.parseStatementList([]lexer.TokenType{lexer.TOKEN_RBRACE})
	
	if !p.expectPeek(lexer.TOKEN_RBRACE) {
		return nil
	}
	
	return &ast.FunctionDeclaration{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTFuncDecl,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Name:             functionName,
		Parameters:       parameters,
		ReturnType:       returnType,
		Body:             body,
		ReturnsReference: returnsReference,
		Visibility:       "", // Top-level functions have no visibility
	}
}

func (p *PrattParser) parseArrowFunctionDeclaration() ast.Declaration {
	position := p.currentToken.Position
	p.enterContext("arrow_function")
	defer p.exitContext("arrow_function")
	
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
	
	// Parse => and expression body
	if !p.expectPeek(lexer.T_DOUBLE_ARROW) {
		return nil
	}
	
	p.nextToken()
	expression := p.parseExpression(LOWEST)
	if expression == nil {
		return nil
	}
	
	return &ast.ArrowFunctionDeclaration{
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

// ============= CLASS DECLARATIONS =============

func (p *PrattParser) parseClassDeclaration() ast.Declaration {
	position := p.currentToken.Position
	p.enterContext("class")
	defer p.exitContext("class")
	
	// Parse class modifiers (abstract, final, readonly)
	var modifiers []string
	if p.currentTokenIs(lexer.T_ABSTRACT) || p.currentTokenIs(lexer.T_FINAL) || p.currentTokenIs(lexer.T_READONLY) {
		modifiers = append(modifiers, p.currentToken.Value)
		p.nextToken()
	}
	
	// Parse class keyword
	if !p.currentTokenIs(lexer.T_CLASS) {
		return nil
	}
	
	// Parse class name
	if !p.expectPeek(lexer.T_STRING) {
		return nil
	}
	
	className := p.currentToken.Value
	
	// Parse optional extends clause
	var extends ast.Expression
	if p.peekTokenIs(lexer.T_EXTENDS) {
		p.nextToken()
		p.nextToken()
		extends = p.parseExpression(LOWEST)
	}
	
	// Parse optional implements clause
	var implements []ast.Expression
	if p.peekTokenIs(lexer.T_IMPLEMENTS) {
		p.nextToken()
		p.nextToken()
		
		implements = append(implements, p.parseExpression(LOWEST))
		
		for p.peekTokenIs(lexer.TOKEN_COMMA) {
			p.nextToken() // consume comma
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
	
	return &ast.ClassDeclaration{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTClass,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Name:       className,
		Modifiers:  modifiers,
		Extends:    extends,
		Implements: implements,
		Members:    members,
	}
}

func (p *PrattParser) parseInterfaceDeclaration() ast.Declaration {
	position := p.currentToken.Position
	p.enterContext("interface")
	defer p.exitContext("interface")
	
	// Parse interface keyword
	if !p.currentTokenIs(lexer.T_INTERFACE) {
		return nil
	}
	
	// Parse interface name
	if !p.expectPeek(lexer.T_STRING) {
		return nil
	}
	
	interfaceName := p.currentToken.Value
	
	// Parse optional extends clause (interfaces can extend multiple interfaces)
	var extends []ast.Expression
	if p.peekTokenIs(lexer.T_EXTENDS) {
		p.nextToken()
		p.nextToken()
		
		extends = append(extends, p.parseExpression(LOWEST))
		
		for p.peekTokenIs(lexer.TOKEN_COMMA) {
			p.nextToken() // consume comma
			p.nextToken()
			extends = append(extends, p.parseExpression(LOWEST))
		}
	}
	
	// Parse interface body
	if !p.expectPeek(lexer.TOKEN_LBRACE) {
		return nil
	}
	
	p.nextToken()
	members := p.parseInterfaceMembers()
	
	if !p.expectPeek(lexer.TOKEN_RBRACE) {
		return nil
	}
	
	return &ast.InterfaceDeclaration{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTClass, // PHP uses same AST kind for interface
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Name:    interfaceName,
		Extends: extends,
		Members: members,
	}
}

func (p *PrattParser) parseTraitDeclaration() ast.Declaration {
	position := p.currentToken.Position
	p.enterContext("trait")
	defer p.exitContext("trait")
	
	// Parse trait keyword
	if !p.currentTokenIs(lexer.T_TRAIT) {
		return nil
	}
	
	// Parse trait name
	if !p.expectPeek(lexer.T_STRING) {
		return nil
	}
	
	traitName := p.currentToken.Value
	
	// Parse trait body
	if !p.expectPeek(lexer.TOKEN_LBRACE) {
		return nil
	}
	
	p.nextToken()
	members := p.parseTraitMembers()
	
	if !p.expectPeek(lexer.TOKEN_RBRACE) {
		return nil
	}
	
	return &ast.TraitDeclaration{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTClass, // PHP uses same AST kind for trait
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Name:    traitName,
		Members: members,
	}
}

func (p *PrattParser) parseEnumDeclaration() ast.Declaration {
	position := p.currentToken.Position
	p.enterContext("enum")
	defer p.exitContext("enum")
	
	// Parse enum keyword
	if !p.currentTokenIs(lexer.T_ENUM) {
		return nil
	}
	
	// Parse enum name
	if !p.expectPeek(lexer.T_STRING) {
		return nil
	}
	
	enumName := p.currentToken.Value
	
	// Parse optional backing type (: string, : int)
	var backingType ast.Type
	if p.peekTokenIs(lexer.TOKEN_COLON) {
		p.nextToken()
		p.nextToken()
		backingType = p.parseType()
	}
	
	// Parse optional implements clause
	var implements []ast.Expression
	if p.peekTokenIs(lexer.T_IMPLEMENTS) {
		p.nextToken()
		p.nextToken()
		
		implements = append(implements, p.parseExpression(LOWEST))
		
		for p.peekTokenIs(lexer.TOKEN_COMMA) {
			p.nextToken() // consume comma
			p.nextToken()
			implements = append(implements, p.parseExpression(LOWEST))
		}
	}
	
	// Parse enum body
	if !p.expectPeek(lexer.TOKEN_LBRACE) {
		return nil
	}
	
	p.nextToken()
	members := p.parseEnumMembers()
	
	if !p.expectPeek(lexer.TOKEN_RBRACE) {
		return nil
	}
	
	return &ast.EnumDeclaration{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTClass, // PHP uses same AST kind for enum
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Name:        enumName,
		BackingType: backingType,
		Implements:  implements,
		Members:     members,
	}
}

// ============= NAMESPACE AND USE DECLARATIONS =============

func (p *PrattParser) parseNamespaceDeclaration() ast.Declaration {
	position := p.currentToken.Position
	p.enterContext("namespace")
	defer p.exitContext("namespace")
	
	// Parse namespace keyword
	if !p.currentTokenIs(lexer.T_NAMESPACE) {
		return nil
	}
	
	// Parse optional namespace name
	var namespaceName ast.Expression
	if p.peekTokenIs(lexer.T_STRING) {
		p.nextToken()
		namespaceName = p.parseNamespaceName()
	}
	
	// Check for block or statement namespace
	if p.peekTokenIs(lexer.TOKEN_LBRACE) {
		// Block namespace: namespace Name { ... }
		p.nextToken()
		p.nextToken()
		
		var statements []ast.Statement
		for !p.currentTokenIs(lexer.TOKEN_RBRACE) && !p.currentTokenIs(lexer.T_EOF) {
			stmt := p.parseStatement()
			if stmt != nil {
				statements = append(statements, stmt)
			}
			p.nextToken()
		}
		
		if !p.expectPeek(lexer.TOKEN_RBRACE) {
			return nil
		}
		
		return &ast.NamespaceDeclaration{
			BaseNode: ast.BaseNode{
				Kind:     ast.ASTNamespace,
				Position: position,
				LineNo:   uint32(position.Line),
			},
			Name:       namespaceName,
			Statements: statements,
			IsBlock:    true,
		}
	} else {
		// Statement namespace: namespace Name;
		if !p.expectPeek(lexer.TOKEN_SEMICOLON) {
			return nil
		}
		
		return &ast.NamespaceDeclaration{
			BaseNode: ast.BaseNode{
				Kind:     ast.ASTNamespace,
				Position: position,
				LineNo:   uint32(position.Line),
			},
			Name:       namespaceName,
			Statements: nil,
			IsBlock:    false,
		}
	}
}

func (p *PrattParser) parseUseDeclaration() ast.Declaration {
	position := p.currentToken.Position
	
	// Parse use keyword
	if !p.currentTokenIs(lexer.T_USE) {
		return nil
	}
	
	// Parse optional use type (function, const)
	var useType string
	if p.peekTokenIs(lexer.T_FUNCTION) || p.peekTokenIs(lexer.T_CONST) {
		p.nextToken()
		useType = p.currentToken.Value
	}
	
	p.nextToken()
	
	// Check for group use syntax
	if p.peekTokenIs(lexer.T_NS_SEPARATOR) {
		return p.parseGroupUseDeclaration(useType, position)
	}
	
	// Parse regular use declaration
	var uses []*ast.UseClause
	
	uses = append(uses, p.parseUseClause())
	
	for p.peekTokenIs(lexer.TOKEN_COMMA) {
		p.nextToken() // consume comma
		p.nextToken()
		uses = append(uses, p.parseUseClause())
	}
	
	if !p.expectPeek(lexer.TOKEN_SEMICOLON) {
		return nil
	}
	
	return &ast.UseDeclaration{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTUse,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Type:  useType,
		Uses:  uses,
		IsGroupUse: false,
	}
}

func (p *PrattParser) parseGroupUseDeclaration(useType string, position lexer.Position) ast.Declaration {
	// Parse the common prefix
	prefix := p.parseNamespaceName()
	
	if !p.expectPeek(lexer.T_NS_SEPARATOR) {
		return nil
	}
	
	if !p.expectPeek(lexer.TOKEN_LBRACE) {
		return nil
	}
	
	p.nextToken()
	
	var uses []*ast.UseClause
	
	// Parse first use clause
	if p.currentTokenIs(lexer.T_FUNCTION) || p.currentTokenIs(lexer.T_CONST) {
		// Individual type specification
		uses = append(uses, p.parseTypedUseClause())
	} else {
		uses = append(uses, p.parseUseClause())
	}
	
	for p.peekTokenIs(lexer.TOKEN_COMMA) {
		p.nextToken() // consume comma
		p.nextToken()
		
		if p.currentTokenIs(lexer.T_FUNCTION) || p.currentTokenIs(lexer.T_CONST) {
			uses = append(uses, p.parseTypedUseClause())
		} else {
			uses = append(uses, p.parseUseClause())
		}
	}
	
	// Optional comma
	if p.peekTokenIs(lexer.TOKEN_COMMA) {
		p.nextToken()
	}
	
	if !p.expectPeek(lexer.TOKEN_RBRACE) {
		return nil
	}
	
	if !p.expectPeek(lexer.TOKEN_SEMICOLON) {
		return nil
	}
	
	return &ast.UseDeclaration{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTGroupUse,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Type:       useType,
		Uses:       uses,
		Prefix:     prefix,
		IsGroupUse: true,
	}
}

// ============= CONSTANT DECLARATIONS =============

func (p *PrattParser) parseConstantDeclaration() ast.Declaration {
	position := p.currentToken.Position
	
	// Parse const keyword
	if !p.currentTokenIs(lexer.T_CONST) {
		return nil
	}
	
	p.nextToken()
	
	var constants []*ast.ConstantClause
	
	// Parse first constant
	constants = append(constants, p.parseConstantClause())
	
	// Parse additional constants
	for p.peekTokenIs(lexer.TOKEN_COMMA) {
		p.nextToken() // consume comma
		p.nextToken()
		constants = append(constants, p.parseConstantClause())
	}
	
	if !p.expectPeek(lexer.TOKEN_SEMICOLON) {
		return nil
	}
	
	return &ast.ConstantDeclaration{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTConstDecl,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Constants: constants,
	}
}

// ============= MEMBER DECLARATIONS =============

func (p *PrattParser) parseClassMembers() []ast.ClassMember {
	var members []ast.ClassMember
	
	for !p.currentTokenIs(lexer.TOKEN_RBRACE) && !p.currentTokenIs(lexer.T_EOF) {
		// Parse attributes first
		var attributes ast.AttributeList
		if p.currentTokenIs(lexer.T_ATTRIBUTE) {
			attributes = p.parseAttributes()
		}
		
		// Parse member with potential visibility modifiers
		member := p.parseClassMember()
		if member != nil {
			if attributes != nil && member != nil {
				// Attach attributes to member
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

func (p *PrattParser) parseClassMember() ast.ClassMember {
	// Parse visibility and other modifiers
	var modifiers []string
	for p.isVisibilityOrModifier(p.currentToken.Type) {
		modifiers = append(modifiers, p.currentToken.Value)
		p.nextToken()
	}
	
	switch p.currentToken.Type {
	case lexer.T_FUNCTION:
		return p.parseMethodDeclaration(modifiers)
	case lexer.T_CONST:
		return p.parseClassConstantDeclaration(modifiers)
	case lexer.T_USE:
		return p.parseTraitUseClause()
	case lexer.T_CASE:
		if p.parsingContext.InEnum {
			return p.parseEnumCase()
		}
		fallthrough
	default:
		// Property declaration
		return p.parsePropertyDeclaration(modifiers)
	}
}

func (p *PrattParser) parseMethodDeclaration(modifiers []string) ast.ClassMember {
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
	
	// Parse method name
	if !p.expectPeek(lexer.T_STRING) {
		return nil
	}
	
	methodName := p.currentToken.Value
	
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
	
	// Parse method body (or semicolon for abstract methods)
	var body ast.Statement
	if p.peekTokenIs(lexer.TOKEN_SEMICOLON) {
		p.nextToken()
		// Abstract method - no body
	} else if p.expectPeek(lexer.TOKEN_LBRACE) {
		p.nextToken()
		body = p.parseStatementList([]lexer.TokenType{lexer.TOKEN_RBRACE})
		
		if !p.expectPeek(lexer.TOKEN_RBRACE) {
			return nil
		}
	} else {
		return nil
	}
	
	return &ast.MethodDeclaration{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTMethod,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Name:             methodName,
		Modifiers:        modifiers,
		Parameters:       parameters,
		ReturnType:       returnType,
		Body:             body,
		ReturnsReference: returnsReference,
	}
}

// ============= UTILITY FUNCTIONS =============

func (p *PrattParser) parseParameterList() []*ast.Parameter {
	var parameters []*ast.Parameter
	
	p.nextToken()
	
	if p.currentTokenIs(lexer.TOKEN_RPAREN) {
		return parameters
	}
	
	parameters = append(parameters, p.parseParameter())
	
	for p.peekTokenIs(lexer.TOKEN_COMMA) {
		p.nextToken() // consume comma
		p.nextToken()
		parameters = append(parameters, p.parseParameter())
	}
	
	if !p.expectPeek(lexer.TOKEN_RPAREN) {
		return nil
	}
	
	return parameters
}

func (p *PrattParser) parseParameter() *ast.Parameter {
	position := p.currentToken.Position
	
	// Parse parameter modifiers (PHP 8.0+ constructor promotion)
	var modifiers []string
	for p.isVisibilityOrModifier(p.currentToken.Type) {
		modifiers = append(modifiers, p.currentToken.Value)
		p.nextToken()
	}
	
	// Parse type hint
	var typeHint ast.Type
	if p.isTypeToken(p.currentToken.Type) {
		typeHint = p.parseType()
		p.nextToken()
	}
	
	// Parse reference
	var isReference bool
	if p.currentTokenIs(lexer.TOKEN_AMPERSAND) {
		isReference = true
		p.nextToken()
	}
	
	// Parse variadic
	var isVariadic bool
	if p.currentTokenIs(lexer.T_ELLIPSIS) {
		isVariadic = true
		p.nextToken()
	}
	
	// Parse parameter name
	if !p.currentTokenIs(lexer.T_VARIABLE) {
		p.errors = append(p.errors, "expected parameter name")
		return nil
	}
	
	parameterName := p.currentToken.Value
	
	// Parse default value
	var defaultValue ast.Expression
	if p.peekTokenIs(lexer.TOKEN_EQUAL) {
		p.nextToken() // consume =
		p.nextToken()
		defaultValue = p.parseExpression(LOWEST)
	}
	
	return &ast.Parameter{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTParam,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Name:         parameterName,
		Type:         typeHint,
		DefaultValue: defaultValue,
		IsReference:  isReference,
		IsVariadic:   isVariadic,
		Modifiers:    modifiers,
	}
}

func (p *PrattParser) parseNamespaceName() ast.Expression {
	parts := []string{p.currentToken.Value}
	position := p.currentToken.Position
	
	for p.peekTokenIs(lexer.T_NS_SEPARATOR) {
		p.nextToken() // consume \
		if !p.expectPeek(lexer.T_STRING) {
			break
		}
		parts = append(parts, p.currentToken.Value)
	}
	
	return &ast.NamespaceNameExpression{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTNamespaceName,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Parts: parts,
	}
}

func (p *PrattParser) parseUseClause() *ast.UseClause {
	position := p.currentToken.Position
	
	name := p.parseNamespaceName()
	
	var alias ast.Expression
	if p.peekTokenIs(lexer.T_AS) {
		p.nextToken() // consume 'as'
		if !p.expectPeek(lexer.T_STRING) {
			return nil
		}
		alias = p.parseIdentifier()
	}
	
	return &ast.UseClause{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTUseElem,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Name:  name,
		Alias: alias,
	}
}

func (p *PrattParser) parseTypedUseClause() *ast.UseClause {
	position := p.currentToken.Position
	
	// Parse type (function or const)
	useType := p.currentToken.Value
	p.nextToken()
	
	name := p.parseNamespaceName()
	
	var alias ast.Expression
	if p.peekTokenIs(lexer.T_AS) {
		p.nextToken() // consume 'as'
		if !p.expectPeek(lexer.T_STRING) {
			return nil
		}
		alias = p.parseIdentifier()
	}
	
	return &ast.UseClause{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTUseElem,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Name:  name,
		Alias: alias,
		Type:  useType,
	}
}

func (p *PrattParser) parseConstantClause() *ast.ConstantClause {
	position := p.currentToken.Position
	
	if !p.currentTokenIs(lexer.T_STRING) {
		p.errors = append(p.errors, "expected constant name")
		return nil
	}
	
	name := p.currentToken.Value
	
	if !p.expectPeek(lexer.TOKEN_EQUAL) {
		return nil
	}
	
	p.nextToken()
	value := p.parseExpression(LOWEST)
	if value == nil {
		return nil
	}
	
	return &ast.ConstantClause{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTConstElem,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Name:  name,
		Value: value,
	}
}

func (p *PrattParser) isVisibilityOrModifier(tokenType lexer.TokenType) bool {
	switch tokenType {
	case lexer.T_PUBLIC, lexer.T_PROTECTED, lexer.T_PRIVATE,
		 lexer.T_STATIC, lexer.T_ABSTRACT, lexer.T_FINAL, lexer.T_READONLY:
		return true
	default:
		return false
	}
}

func (p *PrattParser) isTypeToken(tokenType lexer.TokenType) bool {
	switch tokenType {
	case lexer.T_ARRAY, lexer.T_CALLABLE, lexer.T_STRING, lexer.T_STATIC:
		return true
	default:
		return false
	}
}

// ============= MODIFIER-BASED DECLARATION ROUTING =============

func (p *PrattParser) parseModifiedDeclaration() ast.Declaration {
	// This handles cases where declarations start with modifiers
	// We need to collect all modifiers first, then route to appropriate parser
	var modifiers []string
	
	for p.isVisibilityOrModifier(p.currentToken.Type) {
		modifiers = append(modifiers, p.currentToken.Value)
		p.nextToken()
	}
	
	// Now check what kind of declaration follows
	switch p.currentToken.Type {
	case lexer.T_FUNCTION:
		// This is a method declaration within a class context
		if p.parsingContext.InClass || p.parsingContext.InTrait || p.parsingContext.InInterface {
			member := p.parseMethodDeclaration(modifiers)
			return &ast.DeclarationStatement{Declaration: member}
		}
		// Regular function with modifiers (shouldn't happen, but handle gracefully)
		return p.parseFunctionDeclaration()
		
	case lexer.T_CONST:
		// Class constant with visibility
		if p.parsingContext.InClass || p.parsingContext.InEnum {
			member := p.parseClassConstantDeclaration(modifiers)
			return &ast.DeclarationStatement{Declaration: member}
		}
		// Regular constant
		return p.parseConstantDeclaration()
		
	case lexer.T_CLASS:
		// Class with modifiers (abstract/final)
		return p.parseClassDeclaration()
		
	default:
		// Property declaration
		if p.parsingContext.InClass || p.parsingContext.InTrait {
			member := p.parsePropertyDeclaration(modifiers)
			return &ast.DeclarationStatement{Declaration: member}
		}
		
		p.errors = append(p.errors, fmt.Sprintf("unexpected token after modifiers: %s", p.currentToken.Type))
		return nil
	}
}

func (p *PrattParser) parseStaticDeclaration() ast.Declaration {
	// Handle static keyword - could be static function, static property, or static variable
	if p.peekTokenIs(lexer.T_FUNCTION) {
		return p.parseModifiedDeclaration()
	}
	
	// In class context, this might be a static property
	if p.parsingContext.InClass || p.parsingContext.InTrait {
		return p.parseModifiedDeclaration()
	}
	
	// Otherwise, treat as static variable statement (not a declaration)
	// This should be handled by statement parser
	p.errors = append(p.errors, "static keyword in declaration context")
	return nil
}

func (p *PrattParser) parseAbstractDeclaration() ast.Declaration {
	return p.parseModifiedDeclaration()
}

func (p *PrattParser) parseFinalDeclaration() ast.Declaration {
	return p.parseModifiedDeclaration()
}

func (p *PrattParser) parseReadonlyDeclaration() ast.Declaration {
	return p.parseModifiedDeclaration()
}

// parseInterfaceMembers parses the members of an interface
func (p *PrattParser) parseInterfaceMembers() []ast.ClassMember {
	var members []ast.ClassMember
	
	for !p.currentTokenIs(lexer.TOKEN_RBRACE) && !p.currentTokenIs(lexer.T_EOF) {
		// Interface members can have attributes
		var attributes ast.AttributeList
		if p.currentTokenIs(lexer.T_ATTRIBUTE) {
			attributes = p.parseAttributeList()
		}
		
		// Parse visibility modifiers and other modifiers
		var modifiers []string
		for p.isVisibilityOrModifier(p.currentToken.Type) {
			modifiers = append(modifiers, p.currentToken.Value)
			p.nextToken()
		}
		
		switch p.currentToken.Type {
		case lexer.T_FUNCTION:
			// Interface method declaration (no body)
			method := p.parseInterfaceMethodDeclaration(modifiers)
			if method != nil {
				if attributes != nil {
					method.SetAttributes(attributes)
				}
				members = append(members, method)
			}
			
		case lexer.T_CONST:
			// Interface constant declaration
			constant := p.parseClassConstantDeclaration(modifiers)
			if constant != nil {
				if attributes != nil {
					constant.SetAttributes(attributes)
				}
				members = append(members, constant)
			}
			
		default:
			p.errors = append(p.errors, fmt.Sprintf("unexpected token in interface: %s", p.currentToken.Type))
			p.nextToken() // Skip the problematic token
		}
		
		// Skip semicolons if present
		if p.currentTokenIs(lexer.TOKEN_SEMICOLON) {
			p.nextToken()
		}
	}
	
	return members
}

// parseInterfaceMethodDeclaration parses a method declaration within an interface (no body)
func (p *PrattParser) parseInterfaceMethodDeclaration(modifiers []string) *ast.MethodDeclaration {
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
	
	// Parse method name
	if !p.expectPeek(lexer.T_STRING) {
		return nil
	}
	
	methodName := p.currentToken.Value
	
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
	
	return &ast.MethodDeclaration{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTMethod,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Name:             methodName,
		Modifiers:        modifiers,
		Parameters:       parameters,
		ReturnType:       returnType,
		Body:             nil, // Interface methods have no body
		ReturnsReference: returnsReference,
	}
}

// ============= MISSING METHOD PLACEHOLDERS =============

func (p *PrattParser) parseTraitMembers() []ast.ClassMember {
	// TODO: Implement full trait member parsing
	var members []ast.ClassMember
	
	for !p.currentTokenIs(lexer.TOKEN_RBRACE) && !p.currentTokenIs(lexer.T_EOF) {
		// Skip for now - just advance to avoid infinite loop
		p.nextToken()
	}
	
	return members
}

func (p *PrattParser) parseClassConstantDeclaration(modifiers []string) *ast.ClassConstantDeclaration {
	// TODO: Implement full class constant parsing
	position := p.currentToken.Position
	
	// Skip parsing for now
	for !p.currentTokenIs(lexer.TOKEN_SEMICOLON) && !p.currentTokenIs(lexer.T_EOF) {
		p.nextToken()
	}
	
	return &ast.ClassConstantDeclaration{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTConstDecl,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Constants: []*ast.ConstantClause{},
		Modifiers: modifiers,
	}
}

func (p *PrattParser) parseTraitUseClause() *ast.TraitUseClause {
	// TODO: Implement trait use clause parsing
	position := p.currentToken.Position
	
	// Skip parsing for now
	for !p.currentTokenIs(lexer.TOKEN_SEMICOLON) && !p.currentTokenIs(lexer.T_EOF) {
		p.nextToken()
	}
	
	return &ast.TraitUseClause{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTUse,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Traits: []ast.Expression{},
	}
}