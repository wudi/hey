package parser

import (
	"fmt"
	"strings"

	"github.com/wudi/php-parser/ast"
	"github.com/wudi/php-parser/lexer"
)

// Precedence 操作符优先级
type Precedence int

const (
	_ Precedence = iota
	LOWEST
	TERNARY     // ? :
	LOGICAL_OR  // ||
	LOGICAL_AND // &&
	COALESCE    // ??
	ASSIGN      // =
	EQUALS      // ==
	LESSGREATER // > or <
	SUM         // +
	PRODUCT     // *
	PREFIX      // -X or !X
	POSTFIX     // X++ or X--
	CALL        // myFunction(X)
	INDEX       // array[index]
)

// 操作符优先级映射
var precedences = map[lexer.TokenType]Precedence{
	lexer.TOKEN_QUESTION:        TERNARY,
	lexer.T_DOUBLE_ARROW:        TERNARY,
	lexer.T_BOOLEAN_OR:          LOGICAL_OR,
	lexer.T_BOOLEAN_AND:         LOGICAL_AND,
	lexer.T_COALESCE:            COALESCE,
	lexer.TOKEN_EQUAL:           ASSIGN,
	lexer.T_PLUS_EQUAL:          ASSIGN,
	lexer.T_MINUS_EQUAL:         ASSIGN,
	lexer.T_MUL_EQUAL:           ASSIGN,
	lexer.T_DIV_EQUAL:           ASSIGN,
	lexer.T_CONCAT_EQUAL:        ASSIGN,
	lexer.T_IS_EQUAL:            EQUALS,
	lexer.T_IS_NOT_EQUAL:        EQUALS,
	lexer.T_IS_IDENTICAL:        EQUALS,
	lexer.T_IS_NOT_IDENTICAL:    EQUALS,
	lexer.TOKEN_LT:              LESSGREATER,
	lexer.TOKEN_GT:              LESSGREATER,
	lexer.T_IS_SMALLER_OR_EQUAL: LESSGREATER,
	lexer.T_IS_GREATER_OR_EQUAL: LESSGREATER,
	lexer.T_INSTANCEOF:          LESSGREATER,
	lexer.TOKEN_PLUS:            SUM,
	lexer.TOKEN_MINUS:           SUM,
	lexer.TOKEN_DOT:             SUM,
	lexer.T_OBJECT_OPERATOR:     CALL,
	lexer.TOKEN_LBRACKET:        INDEX,
	lexer.TOKEN_DIVIDE:          PRODUCT,
	lexer.TOKEN_MULTIPLY:        PRODUCT,
	lexer.TOKEN_MODULO:          PRODUCT,
	lexer.T_INC:                 POSTFIX,
	lexer.T_DEC:                 POSTFIX,
	lexer.TOKEN_LPAREN:          CALL,
}

// 前缀解析函数类型
type prefixParseFn func() ast.Expression

// 中缀解析函数类型
type infixParseFn func(ast.Expression) ast.Expression

// Parser 解析器结构体
type Parser struct {
	lexer *lexer.Lexer

	currentToken lexer.Token
	peekToken    lexer.Token

	// 前缀解析函数表
	prefixParseFns map[lexer.TokenType]prefixParseFn
	// 中缀解析函数表
	infixParseFns map[lexer.TokenType]infixParseFn

	errors []string
}

// New 创建新的解析器
func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		lexer:  l,
		errors: []string{},
	}

	// 注册前缀解析函数
	p.prefixParseFns = make(map[lexer.TokenType]prefixParseFn)
	p.registerPrefix(lexer.T_VARIABLE, p.parseVariable)
	p.registerPrefix(lexer.T_LNUMBER, p.parseIntegerLiteral)
	p.registerPrefix(lexer.T_DNUMBER, p.parseFloatLiteral)
	p.registerPrefix(lexer.T_CONSTANT_ENCAPSED_STRING, p.parseStringLiteral)
	p.registerPrefix(lexer.T_STRING, p.parseIdentifier)
	p.registerPrefix(lexer.TOKEN_EXCLAMATION, p.parsePrefixExpression)
	p.registerPrefix(lexer.TOKEN_MINUS, p.parsePrefixExpression)
	p.registerPrefix(lexer.T_INC, p.parsePrefixExpression)
	p.registerPrefix(lexer.T_DEC, p.parsePrefixExpression)
	p.registerPrefix(lexer.TOKEN_LPAREN, p.parseGroupedExpression)
	p.registerPrefix(lexer.T_ARRAY, p.parseArrayExpression)

	// PHP 特定的前缀解析函数
	p.registerPrefix(lexer.T_INLINE_HTML, p.parseInlineHTML)
	p.registerPrefix(lexer.T_OPEN_TAG, p.parseOpenTag)
	p.registerPrefix(lexer.T_COMMENT, p.parseComment)
	p.registerPrefix(lexer.T_DOC_COMMENT, p.parseDocBlockComment)
	p.registerPrefix(lexer.T_START_HEREDOC, p.parseHeredoc)
	p.registerPrefix(lexer.T_ENCAPSED_AND_WHITESPACE, p.parseStringLiteral)
	p.registerPrefix(lexer.T_END_HEREDOC, p.parseEndHeredoc)
	p.registerPrefix(lexer.TOKEN_COMMA, p.parseComma)
	p.registerPrefix(lexer.T_NEW, p.parseNewExpression)
	p.registerPrefix(lexer.T_CLONE, p.parseCloneExpression)

	// Error suppression and special operators
	p.registerPrefix(lexer.TOKEN_AT, p.parseErrorSuppression)
	p.registerPrefix(lexer.T_EMPTY, p.parseEmptyExpression)
	p.registerPrefix(lexer.TOKEN_LBRACKET, p.parseArrayLiteral)

	// Missing tokens - PHP built-in functions and constructs
	p.registerPrefix(lexer.T_EXIT, p.parseExitExpression)
	p.registerPrefix(lexer.T_ISSET, p.parseIssetExpression)
	p.registerPrefix(lexer.T_LIST, p.parseListExpression)
	p.registerPrefix(lexer.T_FUNCTION, p.parseAnonymousFunctionExpression)
	p.registerPrefix(lexer.T_USE, p.parseUseExpression)

	// String interpolation and nowdoc
	p.registerPrefix(lexer.T_NOWDOC, p.parseNowdocExpression)
	p.registerPrefix(lexer.T_CURLY_OPEN, p.parseCurlyOpenExpression)

	// Control flow tokens that might appear as expressions in some contexts
	p.registerPrefix(lexer.T_ELSEIF, p.parseFallback)
	p.registerPrefix(lexer.T_ELSE, p.parseFallback)

	// Ternary operator components
	p.registerPrefix(lexer.TOKEN_QUESTION, p.parseFallback) // ? should be infix
	p.registerPrefix(lexer.TOKEN_COLON, p.parseFallback)    // : should be infix

	// Double arrow (for arrays) - should be infix but register to avoid errors
	p.registerPrefix(lexer.T_DOUBLE_ARROW, p.parseFallback)

	// Fallback handlers for punctuation
	p.registerPrefix(lexer.TOKEN_SEMICOLON, p.parseFallback)
	p.registerPrefix(lexer.TOKEN_RPAREN, p.parseFallback)
	p.registerPrefix(lexer.TOKEN_RBRACKET, p.parseFallback)

	// Additional missing tokens
	p.registerPrefix(lexer.T_EOF, p.parseFallback)                      // End of file
	p.registerPrefix(lexer.TOKEN_DOT, p.parseStringConcatenation)       // String concatenation as prefix
	p.registerPrefix(lexer.TOKEN_RBRACE, p.parseFallback)               // Closing brace
	p.registerPrefix(lexer.TOKEN_AMPERSAND, p.parseReferenceExpression) // Reference operator &$var

	// Keywords that can start expressions or statements
	p.registerPrefix(lexer.T_AS, p.parseFallback)               // 'as' keyword - usually infix
	p.registerPrefix(lexer.T_CASE, p.parseCaseExpression)       // case in switch
	p.registerPrefix(lexer.T_CLASS, p.parseClassExpression)     // class declaration or class expression
	p.registerPrefix(lexer.T_CONST, p.parseConstExpression)     // const declaration
	p.registerPrefix(lexer.T_DEFAULT, p.parseDefaultExpression) // default in switch
	p.registerPrefix(lexer.T_EVAL, p.parseEvalExpression)       // eval() construct
	p.registerPrefix(lexer.T_EXTENDS, p.parseFallback)          // extends keyword

	// Operators that could be prefix in some contexts
	p.registerPrefix(lexer.T_LOGICAL_OR, p.parseFallback)               // 'or' logical operator
	p.registerPrefix(lexer.T_PAAMAYIM_NEKUDOTAYIM, p.parseStaticAccess) // :: static access
	p.registerPrefix(lexer.T_SR, p.parseFallback)                       // >> right shift - usually infix

	// Visibility modifiers
	p.registerPrefix(lexer.T_PRIVATE, p.parseVisibilityModifier)   // private visibility
	p.registerPrefix(lexer.T_PROTECTED, p.parseVisibilityModifier) // protected visibility
	p.registerPrefix(lexer.T_PUBLIC, p.parseVisibilityModifier)    // public visibility

	// Type cast operators
	p.registerPrefix(lexer.T_INT_CAST, p.parseTypeCast)    // (int) cast
	p.registerPrefix(lexer.T_BOOL_CAST, p.parseTypeCast)   // (bool) cast
	p.registerPrefix(lexer.T_DOUBLE_CAST, p.parseTypeCast) // (float) cast
	p.registerPrefix(lexer.T_STRING_CAST, p.parseTypeCast) // (string) cast
	p.registerPrefix(lexer.T_ARRAY_CAST, p.parseTypeCast)  // (array) cast
	p.registerPrefix(lexer.T_OBJECT_CAST, p.parseTypeCast) // (object) cast
	p.registerPrefix(lexer.T_UNSET_CAST, p.parseTypeCast)  // (unset) cast

	// 注册中缀解析函数
	p.infixParseFns = make(map[lexer.TokenType]infixParseFn)
	p.registerInfix(lexer.TOKEN_PLUS, p.parseInfixExpression)
	p.registerInfix(lexer.TOKEN_MINUS, p.parseInfixExpression)
	p.registerInfix(lexer.TOKEN_DIVIDE, p.parseInfixExpression)
	p.registerInfix(lexer.TOKEN_MULTIPLY, p.parseInfixExpression)
	p.registerInfix(lexer.TOKEN_MODULO, p.parseInfixExpression)
	p.registerInfix(lexer.T_IS_EQUAL, p.parseInfixExpression)
	p.registerInfix(lexer.T_IS_NOT_EQUAL, p.parseInfixExpression)
	p.registerInfix(lexer.T_IS_IDENTICAL, p.parseInfixExpression)
	p.registerInfix(lexer.T_IS_NOT_IDENTICAL, p.parseInfixExpression)
	p.registerInfix(lexer.TOKEN_LT, p.parseInfixExpression)
	p.registerInfix(lexer.TOKEN_GT, p.parseInfixExpression)
	p.registerInfix(lexer.T_IS_SMALLER_OR_EQUAL, p.parseInfixExpression)
	p.registerInfix(lexer.T_IS_GREATER_OR_EQUAL, p.parseInfixExpression)
	p.registerInfix(lexer.T_INSTANCEOF, p.parseInstanceofExpression)
	p.registerInfix(lexer.TOKEN_DOT, p.parseInfixExpression)
	p.registerInfix(lexer.T_OBJECT_OPERATOR, p.parsePropertyAccess)
	p.registerInfix(lexer.TOKEN_EQUAL, p.parseAssignmentExpression)
	p.registerInfix(lexer.T_PLUS_EQUAL, p.parseAssignmentExpression)
	p.registerInfix(lexer.T_MINUS_EQUAL, p.parseAssignmentExpression)
	p.registerInfix(lexer.T_MUL_EQUAL, p.parseAssignmentExpression)
	p.registerInfix(lexer.T_DIV_EQUAL, p.parseAssignmentExpression)
	p.registerInfix(lexer.T_CONCAT_EQUAL, p.parseAssignmentExpression)
	p.registerInfix(lexer.T_INC, p.parsePostfixExpression)
	p.registerInfix(lexer.T_DEC, p.parsePostfixExpression)
	p.registerInfix(lexer.TOKEN_LPAREN, p.parseCallExpression)

	// Coalesce and boolean operators
	p.registerInfix(lexer.T_COALESCE, p.parseCoalesceExpression)
	p.registerInfix(lexer.T_BOOLEAN_AND, p.parseBooleanExpression)
	p.registerInfix(lexer.T_BOOLEAN_OR, p.parseBooleanExpression)

	// Array access
	p.registerInfix(lexer.TOKEN_LBRACKET, p.parseArrayAccess)

	// Ternary operator and double arrow
	p.registerInfix(lexer.TOKEN_QUESTION, p.parseTernaryExpression)
	p.registerInfix(lexer.T_DOUBLE_ARROW, p.parseDoubleArrowExpression)

	// 读取两个 token，初始化 currentToken 和 peekToken
	p.nextToken()
	p.nextToken()

	return p
}

// registerPrefix 注册前缀解析函数
func (p *Parser) registerPrefix(tokenType lexer.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

// registerInfix 注册中缀解析函数
func (p *Parser) registerInfix(tokenType lexer.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

// nextToken 前进到下一个 token
func (p *Parser) nextToken() {
	p.currentToken = p.peekToken
	p.peekToken = p.lexer.NextToken()
}

// ParseProgram 解析整个程序
func (p *Parser) ParseProgram() *ast.Program {
	program := ast.NewProgram(p.currentToken.Position)

	// 跳过 PHP 开始标签
	if p.currentToken.Type == lexer.T_OPEN_TAG {
		p.nextToken()
	}

	for !p.isAtEnd() {
		if p.currentToken.Type == lexer.T_CLOSE_TAG {
			break
		}

		stmt := p.parseStatement()
		if stmt != nil {
			program.Body = append(program.Body, stmt)
		}
		p.nextToken()
	}

	return program
}

// isAtEnd 检查是否到达文件结尾
func (p *Parser) isAtEnd() bool {
	return p.currentToken.Type == lexer.T_EOF
}

// getCurrentPrecedence 获取当前token的优先级
func (p *Parser) getCurrentPrecedence() Precedence {
	if precedence, ok := precedences[p.currentToken.Type]; ok {
		return precedence
	}
	return LOWEST
}

// expectToken 检查下一个token是否为期望类型，如果是则前进
func (p *Parser) expectToken(tokenType lexer.TokenType) bool {
	if p.peekToken.Type == tokenType {
		p.nextToken()
		return true
	}
	p.errors = append(p.errors, fmt.Sprintf("expected `%s`, got `%s` at line: %d col: %d", tokenType, p.peekToken.Type, p.peekToken.Position.Line, p.peekToken.Position.Column))
	return false
}

// parseStatement 解析语句
func (p *Parser) parseStatement() ast.Statement {
	switch p.currentToken.Type {
	case lexer.T_ECHO:
		return p.parseEchoStatement()
	case lexer.T_IF:
		return p.parseIfStatement()
	case lexer.T_WHILE:
		return p.parseWhileStatement()
	case lexer.T_FOR:
		return p.parseForStatement()
	case lexer.T_FUNCTION:
		return p.parseFunctionDeclaration()
	case lexer.T_RETURN:
		return p.parseReturnStatement()
	case lexer.T_GLOBAL:
		return p.parseGlobalStatement()
	case lexer.T_STATIC:
		return p.parseStaticStatement()
	case lexer.T_UNSET:
		return p.parseUnsetStatement()
	case lexer.T_DO:
		return p.parseDoWhileStatement()
	case lexer.T_FOREACH:
		return p.parseForeachStatement()
	case lexer.T_SWITCH:
		return p.parseSwitchStatement()
	case lexer.T_TRY:
		return p.parseTryStatement()
	case lexer.T_THROW:
		return p.parseThrowStatement()
	case lexer.T_GOTO:
		return p.parseGotoStatement()
	case lexer.T_BREAK:
		return p.parseBreakStatement()
	case lexer.T_CONTINUE:
		return p.parseContinueStatement()
	case lexer.TOKEN_LBRACE:
		return p.parseBlockStatement()
	case lexer.T_STRING:
		// 检查是否是标签（T_STRING followed by :)
		if p.peekToken.Type == lexer.TOKEN_COLON {
			pos := p.currentToken.Position
			name := p.parseIdentifier()
			p.nextToken() // 跳过 :
			return ast.NewLabelStatement(pos, name)
		}
		return p.parseExpressionStatement()
	default:
		return p.parseExpressionStatement()
	}
}

// parseEchoStatement 解析 echo 语句
func (p *Parser) parseEchoStatement() *ast.EchoStatement {
	stmt := ast.NewEchoStatement(p.currentToken.Position)

	// 解析 echo 的参数
	p.nextToken()
	stmt.Arguments = append(stmt.Arguments, p.parseExpression(LOWEST))

	// 处理多个参数（用逗号分隔）
	for p.peekToken.Type == lexer.TOKEN_COMMA {
		p.nextToken() // 移动到逗号
		p.nextToken() // 移动到下一个表达式
		stmt.Arguments = append(stmt.Arguments, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(lexer.TOKEN_SEMICOLON) {
		return nil
	}

	return stmt
}

// parseIfStatement 解析 if 语句
func (p *Parser) parseIfStatement() *ast.IfStatement {
	pos := p.currentToken.Position

	if !p.expectPeek(lexer.TOKEN_LPAREN) {
		return nil
	}

	p.nextToken()
	condition := p.parseExpression(LOWEST)

	if !p.expectPeek(lexer.TOKEN_RPAREN) {
		return nil
	}

	if !p.expectPeek(lexer.TOKEN_LBRACE) {
		return nil
	}

	ifStmt := ast.NewIfStatement(pos, condition)
	ifStmt.Consequent = p.parseBlockStatements()

	// 检查是否有 else 子句
	if p.peekToken.Type == lexer.T_ELSE {
		p.nextToken() // 移动到 else

		if p.peekToken.Type == lexer.T_IF {
			// else if - 递归解析
			p.nextToken()
			elseIfStmt := p.parseIfStatement()
			if elseIfStmt != nil {
				ifStmt.Alternate = append(ifStmt.Alternate, elseIfStmt)
			}
		} else if p.peekToken.Type == lexer.TOKEN_LBRACE {
			// else block
			p.nextToken()
			ifStmt.Alternate = p.parseBlockStatements()
		}
	}

	return ifStmt
}

// parseWhileStatement 解析 while 语句
func (p *Parser) parseWhileStatement() *ast.WhileStatement {
	pos := p.currentToken.Position

	if !p.expectPeek(lexer.TOKEN_LPAREN) {
		return nil
	}

	p.nextToken()
	condition := p.parseExpression(LOWEST)

	if !p.expectPeek(lexer.TOKEN_RPAREN) {
		return nil
	}

	if !p.expectPeek(lexer.TOKEN_LBRACE) {
		return nil
	}

	whileStmt := ast.NewWhileStatement(pos, condition)
	whileStmt.Body = p.parseBlockStatements()

	return whileStmt
}

// parseForStatement 解析 for 语句
func (p *Parser) parseForStatement() *ast.ForStatement {
	pos := p.currentToken.Position
	forStmt := ast.NewForStatement(pos)

	if !p.expectPeek(lexer.TOKEN_LPAREN) {
		return nil
	}

	// 解析初始化表达式
	if p.peekToken.Type != lexer.TOKEN_SEMICOLON {
		p.nextToken()
		forStmt.Init = p.parseExpression(LOWEST)
	}

	if !p.expectPeek(lexer.TOKEN_SEMICOLON) {
		return nil
	}

	// 解析条件表达式
	if p.peekToken.Type != lexer.TOKEN_SEMICOLON {
		p.nextToken()
		forStmt.Test = p.parseExpression(LOWEST)
	}

	if !p.expectPeek(lexer.TOKEN_SEMICOLON) {
		return nil
	}

	// 解析更新表达式
	if p.peekToken.Type != lexer.TOKEN_RPAREN {
		p.nextToken()
		forStmt.Update = p.parseExpression(LOWEST)
	}

	if !p.expectPeek(lexer.TOKEN_RPAREN) {
		return nil
	}

	if !p.expectPeek(lexer.TOKEN_LBRACE) {
		return nil
	}

	forStmt.Body = p.parseBlockStatements()

	return forStmt
}

// parseFunctionDeclaration 解析函数声明
func (p *Parser) parseFunctionDeclaration() *ast.FunctionDeclaration {
	pos := p.currentToken.Position

	if !p.expectPeek(lexer.T_STRING) {
		return nil
	}

	name := ast.NewIdentifierNode(p.currentToken.Position, p.currentToken.Value)
	funcDecl := ast.NewFunctionDeclaration(pos, name)

	if !p.expectPeek(lexer.TOKEN_LPAREN) {
		return nil
	}

	// 解析参数列表
	if p.peekToken.Type != lexer.TOKEN_RPAREN {
		p.nextToken()

		// 解析第一个参数
		if p.currentToken.Type == lexer.T_VARIABLE {
			param := ast.Parameter{Name: p.currentToken.Value}
			funcDecl.Parameters = append(funcDecl.Parameters, param)

			// 处理更多参数
			for p.peekToken.Type == lexer.TOKEN_COMMA {
				p.nextToken() // 移动到逗号
				if !p.expectPeek(lexer.T_VARIABLE) {
					return nil
				}
				param := ast.Parameter{Name: p.currentToken.Value}
				funcDecl.Parameters = append(funcDecl.Parameters, param)
			}
		}
	}

	if !p.expectPeek(lexer.TOKEN_RPAREN) {
		return nil
	}

	// 检查是否有返回类型声明 ": type"
	if p.peekToken.Type == lexer.TOKEN_COLON {
		p.nextToken() // 移动到 ':'

		if !p.expectPeek(lexer.T_STRING) { // 期望类型名
			return nil
		}

		// 解析返回类型 (这里简单处理为字符串，可以扩展为更复杂的类型系统)
		funcDecl.ReturnType = p.currentToken.Value
	}

	if !p.expectPeek(lexer.TOKEN_LBRACE) {
		return nil
	}

	funcDecl.Body = p.parseBlockStatements()

	return funcDecl
}

// parseReturnStatement 解析 return 语句
func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	pos := p.currentToken.Position

	var returnValue ast.Expression

	if p.peekToken.Type != lexer.TOKEN_SEMICOLON {
		p.nextToken()
		returnValue = p.parseExpression(LOWEST)
	}

	if p.peekToken.Type == lexer.TOKEN_SEMICOLON {
		p.nextToken()
	}

	return ast.NewReturnStatement(pos, returnValue)
}

// parseBreakStatement 解析 break 语句
func (p *Parser) parseBreakStatement() *ast.BreakStatement {
	pos := p.currentToken.Position

	if p.peekToken.Type == lexer.TOKEN_SEMICOLON {
		p.nextToken()
	}

	return ast.NewBreakStatement(pos)
}

// parseContinueStatement 解析 continue 语句
func (p *Parser) parseContinueStatement() *ast.ContinueStatement {
	pos := p.currentToken.Position

	if p.peekToken.Type == lexer.TOKEN_SEMICOLON {
		p.nextToken()
	}

	return ast.NewContinueStatement(pos)
}

// parseBlockStatement 解析块语句
func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	pos := p.currentToken.Position
	block := ast.NewBlockStatement(pos)

	p.nextToken() // 跳过 {

	for !p.currentTokenIs(lexer.TOKEN_RBRACE) && !p.isAtEnd() {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Body = append(block.Body, stmt)
		}
		p.nextToken()
	}

	return block
}

// parseBlockStatements 解析块语句中的语句列表
func (p *Parser) parseBlockStatements() []ast.Statement {
	statements := []ast.Statement{}

	p.nextToken() // 跳过 {

	for !p.currentTokenIs(lexer.TOKEN_RBRACE) && !p.isAtEnd() {
		stmt := p.parseStatement()
		if stmt != nil {
			statements = append(statements, stmt)
		}
		p.nextToken()
	}

	return statements
}

// parseExpressionStatement 解析表达式语句
func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	pos := p.currentToken.Position
	expr := p.parseExpression(LOWEST)

	if expr == nil {
		return nil
	}

	if p.peekToken.Type == lexer.TOKEN_SEMICOLON {
		p.nextToken()
	}

	return ast.NewExpressionStatement(pos, expr)
}

// parseExpression 解析表达式
func (p *Parser) parseExpression(precedence Precedence) ast.Expression {
	prefix := p.prefixParseFns[p.currentToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.currentToken.Type)
		return nil
	}

	leftExp := prefix()

	for p.peekToken.Type != lexer.TOKEN_SEMICOLON && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()
		leftExp = infix(leftExp)
	}

	return leftExp
}

// 前缀解析函数

// parseVariable 解析变量
func (p *Parser) parseVariable() ast.Expression {
	return ast.NewVariable(p.currentToken.Position, p.currentToken.Value)
}

// parseIdentifier 解析标识符
func (p *Parser) parseIdentifier() ast.Expression {
	return ast.NewIdentifierNode(p.currentToken.Position, p.currentToken.Value)
}

// parseIntegerLiteral 解析整数字面量
func (p *Parser) parseIntegerLiteral() ast.Expression {
	return ast.NewNumberLiteral(p.currentToken.Position, p.currentToken.Value, "integer")
}

// parseFloatLiteral 解析浮点数字面量
func (p *Parser) parseFloatLiteral() ast.Expression {
	return ast.NewNumberLiteral(p.currentToken.Position, p.currentToken.Value, "float")
}

// parseStringLiteral 解析字符串字面量
func (p *Parser) parseStringLiteral() ast.Expression {
	// 移除引号获取实际值
	value := strings.Trim(p.currentToken.Value, `"'`)
	return ast.NewStringLiteral(p.currentToken.Position, value, p.currentToken.Value)
}

// parsePrefixExpression 解析前缀表达式
func (p *Parser) parsePrefixExpression() ast.Expression {
	pos := p.currentToken.Position
	operator := p.currentToken.Value

	p.nextToken()
	operand := p.parseExpression(PREFIX)

	return ast.NewUnaryExpression(pos, operator, operand, true)
}

// parseTypeCast 解析类型转换表达式
func (p *Parser) parseTypeCast() ast.Expression {
	pos := p.currentToken.Position
	castType := p.currentToken.Value

	p.nextToken()
	operand := p.parseExpression(PREFIX)

	return ast.NewCastExpression(pos, castType, operand)
}

// parseGroupedExpression 解析括号表达式
func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()

	exp := p.parseExpression(LOWEST)

	if !p.expectPeek(lexer.TOKEN_RPAREN) {
		return nil
	}

	return exp
}

// parseArrayExpression 解析数组表达式
func (p *Parser) parseArrayExpression() ast.Expression {
	pos := p.currentToken.Position
	array := ast.NewArrayExpression(pos)

	if !p.expectPeek(lexer.TOKEN_LPAREN) {
		return nil
	}

	if p.peekToken.Type == lexer.TOKEN_RPAREN {
		p.nextToken()
		return array
	}

	p.nextToken()
	array.Elements = append(array.Elements, p.parseExpression(LOWEST))

	for p.peekToken.Type == lexer.TOKEN_COMMA {
		p.nextToken() // 移动到逗号
		p.nextToken() // 移动到下一个元素
		array.Elements = append(array.Elements, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(lexer.TOKEN_RPAREN) {
		return nil
	}

	return array
}

// 中缀解析函数

// parseInfixExpression 解析中缀表达式
func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	pos := p.currentToken.Position
	operator := p.currentToken.Value
	precedence := p.curPrecedence()

	p.nextToken()
	right := p.parseExpression(precedence)

	return ast.NewBinaryExpression(pos, left, operator, right)
}

// parseAssignmentExpression 解析赋值表达式
func (p *Parser) parseAssignmentExpression(left ast.Expression) ast.Expression {
	pos := p.currentToken.Position
	operator := p.currentToken.Value

	p.nextToken()
	right := p.parseExpression(LOWEST)

	return ast.NewAssignmentExpression(pos, left, operator, right)
}

// parsePostfixExpression 解析后缀表达式
func (p *Parser) parsePostfixExpression(left ast.Expression) ast.Expression {
	pos := p.currentToken.Position
	operator := p.currentToken.Value

	return ast.NewUnaryExpression(pos, operator, left, false)
}

// parseCallExpression 解析函数调用表达式
func (p *Parser) parseCallExpression(fn ast.Expression) ast.Expression {
	pos := p.currentToken.Position
	call := ast.NewCallExpression(pos, fn)

	// 解析参数列表
	call.Arguments = p.parseExpressionList(lexer.TOKEN_RPAREN)

	return call
}

// parseExpressionList 解析表达式列表（用于函数调用参数等）
func (p *Parser) parseExpressionList(end lexer.TokenType) []ast.Expression {
	args := []ast.Expression{}

	if p.peekToken.Type == end {
		p.nextToken()
		return args
	}

	p.nextToken()
	args = append(args, p.parseExpression(LOWEST))

	for p.peekToken.Type == lexer.TOKEN_COMMA {
		p.nextToken()
		p.nextToken()
		args = append(args, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(end) {
		return nil
	}

	return args
}

// 辅助方法

// currentTokenIs 检查当前 token 是否为指定类型
func (p *Parser) currentTokenIs(t lexer.TokenType) bool {
	return p.currentToken.Type == t
}

// peekTokenIs 检查下一个 token 是否为指定类型
func (p *Parser) peekTokenIs(t lexer.TokenType) bool {
	return p.peekToken.Type == t
}

// expectPeek 检查下一个 token 并前进
func (p *Parser) expectPeek(t lexer.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

// peekPrecedence 获取下一个 token 的优先级
func (p *Parser) peekPrecedence() Precedence {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

// curPrecedence 获取当前 token 的优先级
func (p *Parser) curPrecedence() Precedence {
	if p, ok := precedences[p.currentToken.Type]; ok {
		return p
	}
	return LOWEST
}

// 错误处理

// peekError 添加类型检查错误
func (p *Parser) peekError(t lexer.TokenType) {
	msg := fmt.Sprintf("expected next token to be `%s`, got `%s` instead at line: %d col: %d",
		lexer.TokenNames[t], lexer.TokenNames[p.peekToken.Type], p.peekToken.Position.Line, p.peekToken.Position.Column)
	p.errors = append(p.errors, msg)
}

// noPrefixParseFnError 添加前缀解析函数缺失错误
func (p *Parser) noPrefixParseFnError(t lexer.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", lexer.TokenNames[t])
	p.errors = append(p.errors, msg)
}

// Errors 获取解析错误
func (p *Parser) Errors() []string {
	return p.errors
}

// Current 获取当前 token (调试用)
func (p *Parser) Current() lexer.Token {
	return p.currentToken
}

// Peek 获取下一个 token (调试用)
func (p *Parser) Peek() lexer.Token {
	return p.peekToken
}

// PHP 特定的解析函数

// parseInlineHTML 解析内联HTML
func (p *Parser) parseInlineHTML() ast.Expression {
	// 对于内联HTML，创建一个特殊的字面量表达式
	return ast.NewStringLiteral(p.currentToken.Position, p.currentToken.Value, p.currentToken.Value)
}

// parseOpenTag 解析PHP开放标签
func (p *Parser) parseOpenTag() ast.Expression {
	// PHP开放标签通常不作为表达式使用，但为了完整性，创建一个特殊节点
	return ast.NewStringLiteral(p.currentToken.Position, p.currentToken.Value, p.currentToken.Value)
}

// parseComment 解析注释
func (p *Parser) parseComment() ast.Expression {
	// 注释也创建为特殊的字面量
	return ast.NewStringLiteral(p.currentToken.Position, p.currentToken.Value, p.currentToken.Value)
}

// parseHeredoc 解析Heredoc
func (p *Parser) parseHeredoc() ast.Expression {
	startPos := p.currentToken.Position
	startValue := p.currentToken.Value

	// 构建完整的heredoc内容
	var content strings.Builder
	content.WriteString(startValue)

	// 期望下一个token是T_ENCAPSED_AND_WHITESPACE或T_END_HEREDOC
	p.nextToken()

	// 收集heredoc内容
	for p.currentToken.Type == lexer.T_ENCAPSED_AND_WHITESPACE {
		content.WriteString(p.currentToken.Value)
		p.nextToken()
	}

	// 期望T_END_HEREDOC
	if p.currentToken.Type == lexer.T_END_HEREDOC {
		content.WriteString(p.currentToken.Value)
	} else {
		p.errors = append(p.errors, "expected T_END_HEREDOC in heredoc")
	}

	// 返回完整的heredoc作为字符串字面量
	return ast.NewStringLiteral(startPos, content.String(), content.String())
}

// parseEndHeredoc 解析Heredoc结束标记
func (p *Parser) parseEndHeredoc() ast.Expression {
	// 这通常不应该被单独调用，因为T_END_HEREDOC应该在parseHeredoc中处理
	// 但为了避免"no prefix parse function"错误，提供一个基本实现
	return ast.NewStringLiteral(p.currentToken.Position, p.currentToken.Value, p.currentToken.Value)
}

// parseDocBlockComment 解析文档块注释
func (p *Parser) parseDocBlockComment() ast.Expression {
	// 创建文档块注释节点
	return ast.NewDocBlockComment(p.currentToken.Position, p.currentToken.Value, p.currentToken.Value)
}

// parseGlobalStatement 解析全局变量声明
func (p *Parser) parseGlobalStatement() ast.Statement {
	pos := p.currentToken.Position
	globalStmt := ast.NewGlobalStatement(pos)

	// 跳过 'global' 关键字
	p.nextToken()

	// 解析变量列表
	for {
		if p.currentToken.Type == lexer.T_VARIABLE {
			variable := p.parseVariable()
			globalStmt.Variables = append(globalStmt.Variables, variable)
			p.nextToken()

			// 检查是否有更多变量（以逗号分隔）
			if p.currentToken.Type == lexer.TOKEN_COMMA {
				p.nextToken() // 跳过逗号
				continue
			}
			break
		} else {
			p.errors = append(p.errors, "expected variable in global statement")
			break
		}
	}

	return globalStmt
}

// parseComma 解析逗号（通常不应该单独调用，但为了避免错误提供基本实现）
func (p *Parser) parseComma() ast.Expression {
	// 逗号通常不应该作为表达式单独解析，但为了避免解析错误，返回一个简单的字符串字面量
	return ast.NewStringLiteral(p.currentToken.Position, ",", ",")
}

// parseStaticStatement 解析静态变量声明
func (p *Parser) parseStaticStatement() ast.Statement {
	pos := p.currentToken.Position
	staticStmt := ast.NewStaticStatement(pos)

	// 跳过 'static' 关键字
	p.nextToken()

	// 解析变量列表
	for {
		if p.currentToken.Type == lexer.T_VARIABLE {
			variable := p.parseVariable()

			var defaultValue ast.Expression = nil
			p.nextToken()

			// 检查是否有默认值
			if p.currentToken.Type == lexer.TOKEN_EQUAL {
				p.nextToken() // 跳过 =
				defaultValue = p.parseExpression(LOWEST)
				p.nextToken()
			}

			staticVar := ast.NewStaticVariable(pos, variable, defaultValue)
			staticStmt.Variables = append(staticStmt.Variables, staticVar)

			// 检查是否有更多变量（以逗号分隔）
			if p.currentToken.Type == lexer.TOKEN_COMMA {
				p.nextToken() // 跳过逗号
				continue
			}
			break
		} else {
			p.errors = append(p.errors, "expected variable in static statement")
			break
		}
	}

	return staticStmt
}

// parseUnsetStatement 解析unset语句
func (p *Parser) parseUnsetStatement() ast.Statement {
	pos := p.currentToken.Position
	unsetStmt := ast.NewUnsetStatement(pos)

	// 跳过 'unset' 关键字
	if !p.expectPeek(lexer.TOKEN_LPAREN) {
		return nil
	}

	p.nextToken() // 进入括号内部

	// 解析变量列表
	for p.currentToken.Type != lexer.TOKEN_RPAREN && p.currentToken.Type != lexer.T_EOF {
		variable := p.parseExpression(LOWEST)
		unsetStmt.Variables = append(unsetStmt.Variables, variable)

		p.nextToken()
		if p.currentToken.Type == lexer.TOKEN_COMMA {
			p.nextToken() // 跳过逗号
		}
	}

	if !p.expectPeek(lexer.TOKEN_SEMICOLON) {
		return nil
	}

	return unsetStmt
}

// parseDoWhileStatement 解析do-while语句
func (p *Parser) parseDoWhileStatement() ast.Statement {
	pos := p.currentToken.Position

	// 跳过 'do'
	p.nextToken()

	// 解析循环体
	body := p.parseStatement()

	// 期望 'while'
	if !p.expectPeek(lexer.T_WHILE) {
		return nil
	}

	// 期望 '('
	if !p.expectPeek(lexer.TOKEN_LPAREN) {
		return nil
	}

	p.nextToken()
	condition := p.parseExpression(LOWEST)

	// 期望 ')'
	if !p.expectPeek(lexer.TOKEN_RPAREN) {
		return nil
	}

	// 期望 ';'
	if !p.expectPeek(lexer.TOKEN_SEMICOLON) {
		return nil
	}

	return ast.NewDoWhileStatement(pos, body, condition)
}

// parseForeachStatement 解析foreach语句
func (p *Parser) parseForeachStatement() ast.Statement {
	pos := p.currentToken.Position

	// 跳过 'foreach'
	if !p.expectPeek(lexer.TOKEN_LPAREN) {
		return nil
	}

	p.nextToken()
	iterable := p.parseExpression(LOWEST)

	// 期望 'as'
	if !p.expectPeek(lexer.T_AS) {
		return nil
	}

	p.nextToken()
	var key, value ast.Expression

	// 解析第一个变量
	firstVar := p.parseExpression(LOWEST)

	p.nextToken()
	// 检查是否有 '=>'（表示有key）
	if p.currentToken.Type == lexer.T_DOUBLE_ARROW {
		key = firstVar
		p.nextToken()
		value = p.parseExpression(LOWEST)
		p.nextToken()
	} else {
		value = firstVar
	}

	// 期望 ')'
	if p.currentToken.Type != lexer.TOKEN_RPAREN {
		p.errors = append(p.errors, "expected ')' in foreach statement")
		return nil
	}

	p.nextToken()
	body := p.parseStatement()

	return ast.NewForeachStatement(pos, iterable, key, value, body)
}

// parseSwitchStatement 解析switch语句
func (p *Parser) parseSwitchStatement() ast.Statement {
	pos := p.currentToken.Position

	// 跳过 'switch'
	if !p.expectPeek(lexer.TOKEN_LPAREN) {
		return nil
	}

	p.nextToken()
	discriminant := p.parseExpression(LOWEST)

	if !p.expectPeek(lexer.TOKEN_RPAREN) {
		return nil
	}

	if !p.expectPeek(lexer.TOKEN_LBRACE) {
		return nil
	}

	switchStmt := ast.NewSwitchStatement(pos, discriminant)

	p.nextToken()
	// 解析case语句
	for p.currentToken.Type != lexer.TOKEN_RBRACE && p.currentToken.Type != lexer.T_EOF {
		if p.currentToken.Type == lexer.T_CASE {
			casePos := p.currentToken.Position
			p.nextToken()
			test := p.parseExpression(LOWEST)

			if !p.expectPeek(lexer.TOKEN_COLON) {
				continue
			}

			switchCase := ast.NewSwitchCase(casePos, test)

			p.nextToken()
			// 解析case体内的语句
			for p.currentToken.Type != lexer.T_CASE && p.currentToken.Type != lexer.T_DEFAULT &&
				p.currentToken.Type != lexer.TOKEN_RBRACE && p.currentToken.Type != lexer.T_EOF {
				stmt := p.parseStatement()
				if stmt != nil {
					switchCase.Body = append(switchCase.Body, stmt)
				}
				p.nextToken()
			}

			switchStmt.Cases = append(switchStmt.Cases, switchCase)
		} else if p.currentToken.Type == lexer.T_DEFAULT {
			casePos := p.currentToken.Position

			if !p.expectPeek(lexer.TOKEN_COLON) {
				continue
			}

			defaultCase := ast.NewSwitchCase(casePos, nil) // nil表示default

			p.nextToken()
			// 解析default体内的语句
			for p.currentToken.Type != lexer.T_CASE && p.currentToken.Type != lexer.T_DEFAULT &&
				p.currentToken.Type != lexer.TOKEN_RBRACE && p.currentToken.Type != lexer.T_EOF {
				stmt := p.parseStatement()
				if stmt != nil {
					defaultCase.Body = append(defaultCase.Body, stmt)
				}
				p.nextToken()
			}

			switchStmt.Cases = append(switchStmt.Cases, defaultCase)
		} else {
			p.nextToken()
		}
	}

	return switchStmt
}

// parseTryStatement 解析try-catch语句
func (p *Parser) parseTryStatement() ast.Statement {
	pos := p.currentToken.Position
	tryStmt := ast.NewTryStatement(pos)

	// 跳过 'try'
	if !p.expectPeek(lexer.TOKEN_LBRACE) {
		return nil
	}

	p.nextToken()
	// 解析try块
	for p.currentToken.Type != lexer.TOKEN_RBRACE && p.currentToken.Type != lexer.T_EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			tryStmt.Body = append(tryStmt.Body, stmt)
		}
		p.nextToken()
	}

	p.nextToken()

	// 解析catch子句
	for p.currentToken.Type == lexer.T_CATCH {
		catchPos := p.currentToken.Position

		if !p.expectPeek(lexer.TOKEN_LPAREN) {
			continue
		}

		p.nextToken()
		// 解析异常类型
		var parameter ast.Expression
		types := make([]ast.Expression, 0)

		// 解析类型（可能有多个，用|分隔）
		for {
			exceptionType := p.parseExpression(LOWEST)
			types = append(types, exceptionType)

			p.nextToken()
			if p.currentToken.Type == lexer.TOKEN_PIPE { // |
				p.nextToken() // 跳过 |
				continue
			}
			break
		}

		// 解析参数变量
		if p.currentToken.Type == lexer.T_VARIABLE {
			parameter = p.parseVariable()
			p.nextToken()
		}

		if p.currentToken.Type != lexer.TOKEN_RPAREN {
			p.errors = append(p.errors, "expected ')' in catch clause")
			continue
		}

		if !p.expectPeek(lexer.TOKEN_LBRACE) {
			continue
		}

		catchClause := ast.NewCatchClause(catchPos, parameter)
		catchClause.Types = types

		p.nextToken()
		// 解析catch块
		for p.currentToken.Type != lexer.TOKEN_RBRACE && p.currentToken.Type != lexer.T_EOF {
			stmt := p.parseStatement()
			if stmt != nil {
				catchClause.Body = append(catchClause.Body, stmt)
			}
			p.nextToken()
		}

		tryStmt.CatchClauses = append(tryStmt.CatchClauses, catchClause)
		p.nextToken()
	}

	// 解析finally块
	if p.currentToken.Type == lexer.T_FINALLY {
		if !p.expectPeek(lexer.TOKEN_LBRACE) {
			return tryStmt
		}

		p.nextToken()
		// 解析finally块
		for p.currentToken.Type != lexer.TOKEN_RBRACE && p.currentToken.Type != lexer.T_EOF {
			stmt := p.parseStatement()
			if stmt != nil {
				tryStmt.FinallyBlock = append(tryStmt.FinallyBlock, stmt)
			}
			p.nextToken()
		}
	}

	return tryStmt
}

// parseThrowStatement 解析throw语句
func (p *Parser) parseThrowStatement() ast.Statement {
	pos := p.currentToken.Position

	p.nextToken()
	argument := p.parseExpression(LOWEST)

	if !p.expectPeek(lexer.TOKEN_SEMICOLON) {
		return nil
	}

	return ast.NewThrowStatement(pos, argument)
}

// parseGotoStatement 解析goto语句
func (p *Parser) parseGotoStatement() ast.Statement {
	pos := p.currentToken.Position

	p.nextToken()
	if p.currentToken.Type != lexer.T_STRING {
		p.errors = append(p.errors, "expected label name in goto statement")
		return nil
	}

	label := p.parseIdentifier()

	if !p.expectPeek(lexer.TOKEN_SEMICOLON) {
		return nil
	}

	return ast.NewGotoStatement(pos, label)
}

// parseNewExpression 解析new表达式
func (p *Parser) parseNewExpression() ast.Expression {
	pos := p.currentToken.Position

	p.nextToken()
	class := p.parseExpression(PREFIX)

	newExpr := ast.NewNewExpression(pos, class)

	// 检查是否有构造函数参数
	if p.peekToken.Type == lexer.TOKEN_LPAREN {
		p.nextToken() // 移动到 (
		p.nextToken() // 进入括号内部

		// 解析参数列表
		for p.currentToken.Type != lexer.TOKEN_RPAREN && p.currentToken.Type != lexer.T_EOF {
			arg := p.parseExpression(LOWEST)
			newExpr.Arguments = append(newExpr.Arguments, arg)

			p.nextToken()
			if p.currentToken.Type == lexer.TOKEN_COMMA {
				p.nextToken() // 跳过逗号
			}
		}
	}

	return newExpr
}

// parseCloneExpression 解析clone表达式
func (p *Parser) parseCloneExpression() ast.Expression {
	pos := p.currentToken.Position

	p.nextToken()
	object := p.parseExpression(PREFIX)

	return ast.NewCloneExpression(pos, object)
}

// parseInstanceofExpression 解析instanceof表达式
func (p *Parser) parseInstanceofExpression(left ast.Expression) ast.Expression {
	pos := p.currentToken.Position

	p.nextToken()
	right := p.parseExpression(LESSGREATER)

	return ast.NewInstanceofExpression(pos, left, right)
}

// parsePropertyAccess 解析属性访问表达式
func (p *Parser) parsePropertyAccess(left ast.Expression) ast.Expression {
	pos := p.currentToken.Position

	p.nextToken()
	property := p.parseExpression(CALL)

	return ast.NewPropertyAccessExpression(pos, left, property)
}

// parseErrorSuppression 解析错误抑制操作符 @
func (p *Parser) parseErrorSuppression() ast.Expression {
	pos := p.currentToken.Position

	p.nextToken()
	expr := p.parseExpression(PREFIX)

	return ast.NewErrorSuppressionExpression(pos, expr)
}

// parseCoalesceExpression 解析 null 合并操作符 ??
func (p *Parser) parseCoalesceExpression(left ast.Expression) ast.Expression {
	pos := p.currentToken.Position

	precedence := p.getCurrentPrecedence()
	p.nextToken()
	right := p.parseExpression(precedence)

	return ast.NewCoalesceExpression(pos, left, right)
}

// parseBooleanExpression 解析布尔逻辑操作符 && ||
func (p *Parser) parseBooleanExpression(left ast.Expression) ast.Expression {
	pos := p.currentToken.Position
	operator := p.currentToken.Value

	precedence := p.getCurrentPrecedence()
	p.nextToken()
	right := p.parseExpression(precedence)

	return ast.NewBinaryExpression(pos, left, operator, right)
}

// parseArrayAccess 解析数组访问表达式 []
func (p *Parser) parseArrayAccess(left ast.Expression) ast.Expression {
	pos := p.currentToken.Position

	p.nextToken()

	var index *ast.Expression
	if p.currentToken.Type != lexer.TOKEN_RBRACKET {
		expr := p.parseExpression(LOWEST)
		index = &expr
	}

	if !p.expectToken(lexer.TOKEN_RBRACKET) {
		return nil
	}

	return ast.NewArrayAccessExpression(pos, left, index)
}

// parseEmptyExpression 解析 empty() 函数
func (p *Parser) parseEmptyExpression() ast.Expression {
	pos := p.currentToken.Position

	if !p.expectToken(lexer.TOKEN_LPAREN) {
		return nil
	}

	p.nextToken()
	expr := p.parseExpression(LOWEST)

	if !p.expectToken(lexer.TOKEN_RPAREN) {
		return nil
	}

	return ast.NewEmptyExpression(pos, expr)
}

// parseArrayLiteral 解析数组字面量 [] - 根据 PHP 官方语法
func (p *Parser) parseArrayLiteral() ast.Expression {
	pos := p.currentToken.Position

	array := ast.NewArrayExpression(pos)

	// 空数组 []
	if p.peekToken.Type == lexer.TOKEN_RBRACKET {
		p.nextToken()
		return array
	}

	p.nextToken()

	// 解析数组元素列表 (non_empty_array_pair_list)
	for p.currentToken.Type != lexer.TOKEN_RBRACKET && p.currentToken.Type != lexer.T_EOF {
		// possible_array_pair: 可以为空（trailing comma情况）
		if p.currentToken.Type == lexer.TOKEN_COMMA {
			// 跳过空元素（连续的逗号或开头的逗号）
			p.nextToken()
			continue
		}

		// 解析数组元素表达式
		element := p.parseExpression(LOWEST)
		if element != nil {
			array.Elements = append(array.Elements, element)
		}

		p.nextToken()

		// 检查是否到达数组结尾
		if p.currentToken.Type == lexer.TOKEN_RBRACKET {
			break
		}

		// 期望逗号分隔符，但允许省略（在结尾）
		if p.currentToken.Type == lexer.TOKEN_COMMA {
			p.nextToken()
			// 允许 trailing comma: [1, 2, 3,]
			if p.currentToken.Type == lexer.TOKEN_RBRACKET {
				break
			}
		} else if p.currentToken.Type != lexer.TOKEN_RBRACKET {
			// 如果不是逗号也不是结尾括号，则报错
			p.errors = append(p.errors, "expected ',' or ']' in array")
			return nil
		}
	}

	// 确保我们在正确的位置（应该是 ']'）
	if p.currentToken.Type != lexer.TOKEN_RBRACKET {
		p.errors = append(p.errors, "expected ']' to close array")
		return nil
	}

	return array
}

// parseFallback 解析回退处理（用于标点符号）
func (p *Parser) parseFallback() ast.Expression {
	// 对于标点符号，通常意味着表达式结束
	// 返回 nil 让上级处理
	pos := p.currentToken.Position
	value := p.currentToken.Value

	// 创建一个简单的标识符作为占位符
	return ast.NewIdentifierNode(pos, value)
}

// ============== 新增的解析函数实现 ==============

// parseExitExpression 解析 exit/die 表达式
func (p *Parser) parseExitExpression() ast.Expression {
	pos := p.currentToken.Position

	var argument ast.Expression

	// exit 可能带有括号和参数，也可能不带
	if p.peekToken.Type == lexer.TOKEN_LPAREN {
		p.nextToken() // 移动到 (
		if p.peekToken.Type != lexer.TOKEN_RPAREN {
			p.nextToken() // 进入括号
			argument = p.parseExpression(LOWEST)
		}
		if p.peekToken.Type == lexer.TOKEN_RPAREN {
			p.nextToken() // 跳过 )
		}
	} else if p.peekToken.Type != lexer.TOKEN_SEMICOLON &&
		p.peekToken.Type != lexer.T_EOF {
		// exit 后直接跟表达式，不带括号
		p.nextToken()
		argument = p.parseExpression(LOWEST)
	}

	return ast.NewExitExpression(pos, argument)
}

// parseIssetExpression 解析 isset() 表达式
func (p *Parser) parseIssetExpression() ast.Expression {
	pos := p.currentToken.Position

	if !p.expectToken(lexer.TOKEN_LPAREN) {
		return nil
	}

	var arguments []ast.Expression

	// 解析参数列表
	if p.peekToken.Type != lexer.TOKEN_RPAREN {
		p.nextToken()
		arguments = append(arguments, p.parseExpression(LOWEST))

		// 处理多个参数
		for p.peekToken.Type == lexer.TOKEN_COMMA {
			p.nextToken() // 跳过逗号
			p.nextToken() // 移动到下一个参数
			arguments = append(arguments, p.parseExpression(LOWEST))
		}
	}

	if !p.expectToken(lexer.TOKEN_RPAREN) {
		return nil
	}

	return ast.NewIssetExpression(pos, arguments)
}

// parseListExpression 解析 list() 表达式
func (p *Parser) parseListExpression() ast.Expression {
	pos := p.currentToken.Position

	if !p.expectToken(lexer.TOKEN_LPAREN) {
		return nil
	}

	var elements []ast.Expression

	// 解析 list 元素
	if p.peekToken.Type != lexer.TOKEN_RPAREN {
		p.nextToken()

		for {
			if p.currentToken.Type == lexer.TOKEN_COMMA {
				// 空元素（如 list(, $b) ）
				elements = append(elements, nil)
			} else {
				elements = append(elements, p.parseExpression(LOWEST))
			}

			if p.peekToken.Type == lexer.TOKEN_RPAREN {
				break
			}

			if p.peekToken.Type != lexer.TOKEN_COMMA {
				break
			}

			p.nextToken() // 跳过逗号
			p.nextToken() // 移动到下一个元素
		}
	}

	if !p.expectToken(lexer.TOKEN_RPAREN) {
		return nil
	}

	return ast.NewListExpression(pos, elements)
}

// parseAnonymousFunctionExpression 解析匿名函数表达式
func (p *Parser) parseAnonymousFunctionExpression() ast.Expression {
	pos := p.currentToken.Position

	// 解析参数列表
	if !p.expectToken(lexer.TOKEN_LPAREN) {
		return nil
	}

	var parameters []ast.Parameter

	// 解析参数
	if p.peekToken.Type != lexer.TOKEN_RPAREN {
		p.nextToken()

		for {
			if p.currentToken.Type == lexer.T_VARIABLE {
				param := ast.Parameter{Name: p.currentToken.Value}
				parameters = append(parameters, param)
			}

			if p.peekToken.Type == lexer.TOKEN_RPAREN {
				break
			}

			if p.peekToken.Type != lexer.TOKEN_COMMA {
				break
			}

			p.nextToken() // 跳过逗号
			p.nextToken() // 移动到下一个参数
		}
	}

	if !p.expectToken(lexer.TOKEN_RPAREN) {
		return nil
	}

	// 检查 use 子句
	var useClause []ast.Expression
	if p.peekToken.Type == lexer.T_USE {
		p.nextToken() // 跳过 use
		if !p.expectToken(lexer.TOKEN_LPAREN) {
			return nil
		}

		if p.peekToken.Type != lexer.TOKEN_RPAREN {
			p.nextToken()

			for {
				if p.currentToken.Type == lexer.T_VARIABLE {
					useClause = append(useClause, p.parseVariable())
				}

				if p.peekToken.Type == lexer.TOKEN_RPAREN {
					break
				}

				if p.peekToken.Type != lexer.TOKEN_COMMA {
					break
				}

				p.nextToken() // 跳过逗号
				p.nextToken() // 移动到下一个变量
			}
		}

		if !p.expectToken(lexer.TOKEN_RPAREN) {
			return nil
		}
	}

	// 解析函数体
	if !p.expectToken(lexer.TOKEN_LBRACE) {
		return nil
	}

	body := p.parseBlockStatements()

	return ast.NewAnonymousFunctionExpression(pos, parameters, body, useClause)
}

// parseUseExpression 解析 use 表达式（在表达式上下文中）
func (p *Parser) parseUseExpression() ast.Expression {
	pos := p.currentToken.Position

	// 这里简单处理为标识符，实际的 use 语句应该在语句级别处理
	return ast.NewIdentifierNode(pos, "use")
}

// parseNowdocExpression 解析 nowdoc 表达式
func (p *Parser) parseNowdocExpression() ast.Expression {
	pos := p.currentToken.Position
	content := p.currentToken.Value

	// Nowdoc 是单一token，直接返回字符串字面量
	return ast.NewStringLiteral(pos, content, content)
}

// parseCurlyOpenExpression 解析字符串插值开始标记
func (p *Parser) parseCurlyOpenExpression() ast.Expression {
	pos := p.currentToken.Position

	// 这通常在字符串插值中出现，这里简单处理
	return ast.NewStringLiteral(pos, p.currentToken.Value, p.currentToken.Value)
}

// parseTernaryExpression 解析三元运算符表达式
func (p *Parser) parseTernaryExpression(condition ast.Expression) ast.Expression {
	pos := p.currentToken.Position

	p.nextToken()

	var consequent ast.Expression

	// 检查是否是短三元运算符 ?:
	if p.currentToken.Type == lexer.TOKEN_COLON {
		// 短三元运算符，consequent 为 nil
		consequent = nil
	} else {
		consequent = p.parseExpression(LOWEST)

		// 期望冒号
		if !p.expectToken(lexer.TOKEN_COLON) {
			return nil
		}
	}

	p.nextToken()
	alternate := p.parseExpression(TERNARY)

	return ast.NewTernaryExpression(pos, condition, consequent, alternate)
}

// parseDoubleArrowExpression 解析 => 表达式（数组元素）
func (p *Parser) parseDoubleArrowExpression(key ast.Expression) ast.Expression {
	pos := p.currentToken.Position

	p.nextToken()
	value := p.parseExpression(TERNARY)

	return ast.NewArrayElementExpression(pos, key, value)
}

// ============== 新增的缺失函数实现 ==============

// parseStringConcatenation 解析字符串连接（当 . 作为前缀使用时）
func (p *Parser) parseStringConcatenation() ast.Expression {
	pos := p.currentToken.Position

	// . 作为前缀通常意味着错误的语法，返回占位符
	return ast.NewIdentifierNode(pos, ".")
}

// parseReferenceExpression 解析引用表达式 &$var
func (p *Parser) parseReferenceExpression() ast.Expression {
	pos := p.currentToken.Position

	p.nextToken()
	operand := p.parseExpression(PREFIX)

	return ast.NewUnaryExpression(pos, "&", operand, true)
}

// parseCaseExpression 解析 case 表达式（switch 语句中的 case）
func (p *Parser) parseCaseExpression() ast.Expression {
	pos := p.currentToken.Position

	p.nextToken()
	test := p.parseExpression(LOWEST)

	// 返回一个特殊的case表达式节点
	return ast.NewCaseExpression(pos, test)
}

// parseClassExpression 解析类表达式或类声明
func (p *Parser) parseClassExpression() ast.Expression {
	pos := p.currentToken.Position

	// 类声明通常需要名称和可能的扩展
	p.nextToken()

	var name ast.Expression
	if p.currentToken.Type == lexer.T_STRING {
		name = ast.NewIdentifierNode(p.currentToken.Position, p.currentToken.Value)
	}

	return ast.NewClassExpression(pos, name, nil, nil) // name, extends, implements
}

// parseConstExpression 解析常量声明表达式
func (p *Parser) parseConstExpression() ast.Expression {
	pos := p.currentToken.Position

	p.nextToken()

	// 解析常量名
	var name ast.Expression
	if p.currentToken.Type == lexer.T_STRING {
		name = ast.NewIdentifierNode(p.currentToken.Position, p.currentToken.Value)
	}

	return ast.NewConstExpression(pos, name, nil) // name, value
}

// parseDefaultExpression 解析 default 表达式（switch 语句中的 default）
func (p *Parser) parseDefaultExpression() ast.Expression {
	pos := p.currentToken.Position

	// default case 没有测试表达式
	return ast.NewCaseExpression(pos, nil)
}

// parseEvalExpression 解析 eval() 表达式
func (p *Parser) parseEvalExpression() ast.Expression {
	pos := p.currentToken.Position

	var argument ast.Expression

	// eval 可能带有括号和参数
	if p.peekToken.Type == lexer.TOKEN_LPAREN {
		p.nextToken() // 移动到 (
		if p.peekToken.Type != lexer.TOKEN_RPAREN {
			p.nextToken() // 进入括号
			argument = p.parseExpression(LOWEST)
		}

		if p.peekToken.Type == lexer.TOKEN_RPAREN {
			p.nextToken() // 跳过右括号
		}
	}

	return ast.NewEvalExpression(pos, argument)
}

// parseStaticAccess 解析静态访问表达式 Class::method 或 Class::$property
func (p *Parser) parseStaticAccess() ast.Expression {
	pos := p.currentToken.Position

	// :: 作为前缀通常意味着省略了左边的类名，如 ::method()
	p.nextToken()

	var property ast.Expression
	if p.currentToken.Type != lexer.T_EOF {
		property = p.parseExpression(CALL)
	}

	// 创建一个静态访问表达式，左边为 nil 表示省略了类名
	return ast.NewStaticAccessExpression(pos, nil, property)
}

// parseVisibilityModifier 解析可见性修饰符 public/private/protected
func (p *Parser) parseVisibilityModifier() ast.Expression {
	pos := p.currentToken.Position
	modifier := p.currentToken.Value

	// 可见性修饰符后面应该跟着属性或方法声明
	return ast.NewVisibilityModifierExpression(pos, modifier)
}
