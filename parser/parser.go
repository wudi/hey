package parser

import (
	"fmt"
	"strings"

	"github.com/yourname/php-parser/ast"
	"github.com/yourname/php-parser/lexer"
)

// Precedence 操作符优先级
type Precedence int

const (
	_ Precedence = iota
	LOWEST
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
	lexer.TOKEN_EQUAL:             ASSIGN,
	lexer.T_PLUS_EQUAL:            ASSIGN,
	lexer.T_MINUS_EQUAL:           ASSIGN,
	lexer.T_MUL_EQUAL:             ASSIGN,
	lexer.T_DIV_EQUAL:             ASSIGN,
	lexer.T_CONCAT_EQUAL:          ASSIGN,
	lexer.T_IS_EQUAL:              EQUALS,
	lexer.T_IS_NOT_EQUAL:          EQUALS,
	lexer.T_IS_IDENTICAL:          EQUALS,
	lexer.T_IS_NOT_IDENTICAL:      EQUALS,
	lexer.TOKEN_LT:                LESSGREATER,
	lexer.TOKEN_GT:                LESSGREATER,
	lexer.T_IS_SMALLER_OR_EQUAL:   LESSGREATER,
	lexer.T_IS_GREATER_OR_EQUAL:   LESSGREATER,
	lexer.TOKEN_PLUS:              SUM,
	lexer.TOKEN_MINUS:             SUM,
	lexer.TOKEN_DIVIDE:            PRODUCT,
	lexer.TOKEN_MULTIPLY:          PRODUCT,
	lexer.TOKEN_MODULO:            PRODUCT,
	lexer.T_INC:                   POSTFIX,
	lexer.T_DEC:                   POSTFIX,
	lexer.TOKEN_LPAREN:            CALL,
	lexer.TOKEN_LBRACKET:          INDEX,
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
	p.registerInfix(lexer.TOKEN_EQUAL, p.parseAssignmentExpression)
	p.registerInfix(lexer.T_PLUS_EQUAL, p.parseAssignmentExpression)
	p.registerInfix(lexer.T_MINUS_EQUAL, p.parseAssignmentExpression)
	p.registerInfix(lexer.T_MUL_EQUAL, p.parseAssignmentExpression)
	p.registerInfix(lexer.T_DIV_EQUAL, p.parseAssignmentExpression)
	p.registerInfix(lexer.T_CONCAT_EQUAL, p.parseAssignmentExpression)
	p.registerInfix(lexer.T_INC, p.parsePostfixExpression)
	p.registerInfix(lexer.T_DEC, p.parsePostfixExpression)

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
	case lexer.T_BREAK:
		return p.parseBreakStatement()
	case lexer.T_CONTINUE:
		return p.parseContinueStatement()
	case lexer.TOKEN_LBRACE:
		return p.parseBlockStatement()
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
	msg := fmt.Sprintf("expected next token to be %s, got %s instead",
		lexer.TokenNames[t], lexer.TokenNames[p.peekToken.Type])
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