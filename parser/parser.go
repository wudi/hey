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

var (
	// 前缀解析函数类型
	globalPrefixParseFns map[lexer.TokenType]PrefixParseFn
	// 中缀解析函数类型
	globalInfixParseFns map[lexer.TokenType]InfixParseFn
)

func init() {
	globalPrefixParseFns = map[lexer.TokenType]PrefixParseFn{
		lexer.T_VARIABLE:                 parseVariable,
		lexer.T_LNUMBER:                  parseIntegerLiteral,
		lexer.T_DNUMBER:                  parseFloatLiteral,
		lexer.T_CONSTANT_ENCAPSED_STRING: parseStringLiteral,
		lexer.T_STRING:                   parseIdentifier,
		lexer.TOKEN_EXCLAMATION:          parsePrefixExpression,
		lexer.TOKEN_MINUS:                parsePrefixExpression,
		lexer.T_INC:                      parsePrefixExpression,
		lexer.T_DEC:                      parsePrefixExpression,
		lexer.TOKEN_LPAREN:               parseGroupedExpression,
		lexer.T_ARRAY:                    parseArrayExpression,
		lexer.T_INLINE_HTML:              parseInlineHTML,
		lexer.T_OPEN_TAG:                 parseOpenTag,
		lexer.T_COMMENT:                  parseComment,
		lexer.T_DOC_COMMENT:              parseDocBlockComment,
		lexer.T_START_HEREDOC:            parseHeredoc,
		lexer.T_ENCAPSED_AND_WHITESPACE:  parseStringLiteral,
		lexer.T_END_HEREDOC:              parseEndHeredoc,
		lexer.TOKEN_COMMA:                parseComma,
		lexer.T_NEW:                      parseNewExpression,
		lexer.T_CLONE:                    parseCloneExpression,
		lexer.TOKEN_AT:                   parseErrorSuppression,
		lexer.T_EMPTY:                    parseEmptyExpression,
		lexer.TOKEN_LBRACKET:             parseArrayLiteral,
		lexer.T_EXIT:                     parseExitExpression,
		lexer.T_ISSET:                    parseIssetExpression,
		lexer.T_LIST:                     parseListExpression,
		lexer.T_FUNCTION:                 parseAnonymousFunctionExpression,
		lexer.T_USE:                      parseUseExpression,
		lexer.T_NOWDOC:                   parseNowdocExpression,
		lexer.T_CURLY_OPEN:               parseCurlyOpenExpression,
		lexer.TOKEN_QUOTE:                parseInterpolatedString,
		lexer.T_ELSEIF:                   parseFallback,
		lexer.T_ELSE:                     parseFallback,
		lexer.TOKEN_QUESTION:             parseFallback,
		lexer.TOKEN_COLON:                parseFallback,
		lexer.T_DOUBLE_ARROW:             parseFallback,
		lexer.TOKEN_SEMICOLON:            parseFallback,
		lexer.TOKEN_RPAREN:               parseFallback,
		lexer.TOKEN_RBRACKET:             parseFallback,
		lexer.T_EOF:                      parseFallback,
		lexer.TOKEN_DOT:                  parseStringConcatenation,
		lexer.TOKEN_RBRACE:               parseFallback,
		lexer.TOKEN_AMPERSAND:            parseReferenceExpression,
		lexer.T_AS:                       parseFallback,
		lexer.T_CASE:                     parseCaseExpression,
		lexer.T_CLASS:                    parseClassExpression,
		lexer.T_CONST:                    parseConstExpression,
		lexer.T_DEFAULT:                  parseDefaultExpression,
		lexer.T_EVAL:                     parseEvalExpression,
		lexer.T_EXTENDS:                  parseFallback,
		lexer.T_LOGICAL_OR:               parseFallback,
		lexer.T_PAAMAYIM_NEKUDOTAYIM:     parseStaticAccess,
		lexer.T_SR:                       parseFallback,
		lexer.T_PRIVATE:                  parseVisibilityModifier,
		lexer.T_PROTECTED:                parseVisibilityModifier,
		lexer.T_PUBLIC:                   parseVisibilityModifier,
		lexer.T_INT_CAST:                 parseTypeCast,
		lexer.T_BOOL_CAST:                parseTypeCast,
		lexer.T_DOUBLE_CAST:              parseTypeCast,
		lexer.T_STRING_CAST:              parseTypeCast,
		lexer.T_ARRAY_CAST:               parseTypeCast,
		lexer.T_OBJECT_CAST:              parseTypeCast,
		lexer.T_UNSET_CAST:               parseTypeCast,
	}
	globalInfixParseFns = map[lexer.TokenType]InfixParseFn{
		lexer.TOKEN_PLUS:            parseInfixExpression,
		lexer.TOKEN_MINUS:           parseInfixExpression,
		lexer.TOKEN_DIVIDE:          parseInfixExpression,
		lexer.TOKEN_MULTIPLY:        parseInfixExpression,
		lexer.TOKEN_MODULO:          parseInfixExpression,
		lexer.T_IS_EQUAL:            parseInfixExpression,
		lexer.T_IS_NOT_EQUAL:        parseInfixExpression,
		lexer.T_IS_IDENTICAL:        parseInfixExpression,
		lexer.T_IS_NOT_IDENTICAL:    parseInfixExpression,
		lexer.TOKEN_LT:              parseInfixExpression,
		lexer.TOKEN_GT:              parseInfixExpression,
		lexer.T_IS_SMALLER_OR_EQUAL: parseInfixExpression,
		lexer.T_IS_GREATER_OR_EQUAL: parseInfixExpression,
		lexer.T_INSTANCEOF:          parseInstanceofExpression,
		lexer.TOKEN_DOT:             parseInfixExpression,
		lexer.T_OBJECT_OPERATOR:     parsePropertyAccess,
		lexer.TOKEN_EQUAL:           parseAssignmentExpression,
		lexer.T_PLUS_EQUAL:          parseAssignmentExpression,
		lexer.T_MINUS_EQUAL:         parseAssignmentExpression,
		lexer.T_MUL_EQUAL:           parseAssignmentExpression,
		lexer.T_DIV_EQUAL:           parseAssignmentExpression,
		lexer.T_CONCAT_EQUAL:        parseAssignmentExpression,
		lexer.T_INC:                 parsePostfixExpression,
		lexer.T_DEC:                 parsePostfixExpression,
		lexer.TOKEN_LPAREN:          parseCallExpression,
		lexer.T_COALESCE:            parseCoalesceExpression,
		lexer.T_BOOLEAN_AND:         parseBooleanExpression,
		lexer.T_BOOLEAN_OR:          parseBooleanExpression,
		lexer.TOKEN_LBRACKET:        parseArrayAccess,
		lexer.TOKEN_QUESTION:        parseTernaryExpression,
		lexer.T_DOUBLE_ARROW:        parseDoubleArrowExpression,
	}
}

type PrefixParseFn func(*Parser) ast.Expression
type InfixParseFn func(*Parser, ast.Expression) ast.Expression

// Parser 解析器结构体
type Parser struct {
	lexer        *lexer.Lexer
	currentToken lexer.Token
	peekToken    lexer.Token
	errors       []string
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		lexer:  l,
		errors: []string{},
	}
	// 读取两个 token，初始化 currentToken 和 peekToken
	p.nextToken()
	p.nextToken()

	return p
}

// nextToken 前进到下一个 token
func (p *Parser) nextToken() {
	p.currentToken = p.peekToken
	p.peekToken = p.lexer.NextToken()
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

// peekError 添加类型检查错误
func (p *Parser) peekError(t lexer.TokenType) {
	msg := fmt.Sprintf("expected next token to be `%s`, got `%s` instead at line: %d col: %d",
		lexer.TokenNames[t], lexer.TokenNames[p.peekToken.Type], p.peekToken.Position.Line, p.peekToken.Position.Column)
	p.errors = append(p.errors, msg)
}

// Errors 获取解析错误
func (p *Parser) Errors() []string {
	return p.errors
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

		stmt := parseStatement(p)
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

// noPrefixParseFnError 添加前缀解析函数缺失错误
func (p *Parser) noPrefixParseFnError(t lexer.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for `%s` %s found", lexer.TokenNames[t], t)
	p.errors = append(p.errors, msg)
}

// parseStatement 解析语句
func parseStatement(p *Parser) ast.Statement {
	switch p.currentToken.Type {
	case lexer.T_ECHO:
		return parseEchoStatement(p)
	case lexer.T_IF:
		return parseIfStatement(p)
	case lexer.T_WHILE:
		return parseWhileStatement(p)
	case lexer.T_FOR:
		return parseForStatement(p)
	case lexer.T_FUNCTION:
		return parseFunctionDeclaration(p)
	case lexer.T_RETURN:
		return parseReturnStatement(p)
	case lexer.T_GLOBAL:
		return parseGlobalStatement(p)
	case lexer.T_STATIC:
		return parseStaticStatement(p)
	case lexer.T_UNSET:
		return parseUnsetStatement(p)
	case lexer.T_DO:
		return parseDoWhileStatement(p)
	case lexer.T_FOREACH:
		return parseForeachStatement(p)
	case lexer.T_SWITCH:
		return parseSwitchStatement(p)
	case lexer.T_TRY:
		return parseTryStatement(p)
	case lexer.T_THROW:
		return parseThrowStatement(p)
	case lexer.T_GOTO:
		return parseGotoStatement(p)
	case lexer.T_BREAK:
		return parseBreakStatement(p)
	case lexer.T_CONTINUE:
		return parseContinueStatement(p)
	case lexer.TOKEN_LBRACE:
		return parseBlockStatement(p)
	case lexer.T_STRING:
		// 检查是否是标签（T_STRING followed by :)
		if p.peekToken.Type == lexer.TOKEN_COLON {
			pos := p.currentToken.Position
			name := parseIdentifier(p)
			p.nextToken() // 跳过 :
			return ast.NewLabelStatement(pos, name)
		}
		return parseExpressionStatement(p)
	default:
		return parseExpressionStatement(p)
	}
}

// parseEchoStatement 解析 echo 语句
func parseEchoStatement(p *Parser) *ast.EchoStatement {
	stmt := ast.NewEchoStatement(p.currentToken.Position)

	// 解析 echo 的参数
	p.nextToken()
	stmt.Arguments = append(stmt.Arguments, parseExpression(p, LOWEST))

	// 处理多个参数（用逗号分隔）
	for p.peekToken.Type == lexer.TOKEN_COMMA {
		p.nextToken() // 移动到逗号
		p.nextToken() // 移动到下一个表达式
		stmt.Arguments = append(stmt.Arguments, parseExpression(p, LOWEST))
	}

	if !p.expectPeek(lexer.TOKEN_SEMICOLON) {
		return nil
	}

	return stmt
}

// parseIfStatement 解析 if 语句
func parseIfStatement(p *Parser) *ast.IfStatement {
	pos := p.currentToken.Position

	if !p.expectPeek(lexer.TOKEN_LPAREN) {
		return nil
	}

	p.nextToken()
	condition := parseExpression(p, LOWEST)

	if !p.expectPeek(lexer.TOKEN_RPAREN) {
		return nil
	}

	if !p.expectPeek(lexer.TOKEN_LBRACE) {
		return nil
	}

	ifStmt := ast.NewIfStatement(pos, condition)
	ifStmt.Consequent = parseBlockStatements(p)

	// 检查是否有 else 子句
	if p.peekToken.Type == lexer.T_ELSE {
		p.nextToken() // 移动到 else

		if p.peekToken.Type == lexer.T_IF {
			// else if - 递归解析
			p.nextToken()
			elseIfStmt := parseIfStatement(p)
			if elseIfStmt != nil {
				ifStmt.Alternate = append(ifStmt.Alternate, elseIfStmt)
			}
		} else if p.peekToken.Type == lexer.TOKEN_LBRACE {
			// else block
			p.nextToken()
			ifStmt.Alternate = parseBlockStatements(p)
		}
	}

	return ifStmt
}

// parseWhileStatement 解析 while 语句
func parseWhileStatement(p *Parser) *ast.WhileStatement {
	pos := p.currentToken.Position

	if !p.expectPeek(lexer.TOKEN_LPAREN) {
		return nil
	}

	p.nextToken()
	condition := parseExpression(p, LOWEST)

	if !p.expectPeek(lexer.TOKEN_RPAREN) {
		return nil
	}

	if !p.expectPeek(lexer.TOKEN_LBRACE) {
		return nil
	}

	whileStmt := ast.NewWhileStatement(pos, condition)
	whileStmt.Body = parseBlockStatements(p)

	return whileStmt
}

// parseForStatement 解析 for 语句
func parseForStatement(p *Parser) *ast.ForStatement {
	pos := p.currentToken.Position
	forStmt := ast.NewForStatement(pos)

	if !p.expectPeek(lexer.TOKEN_LPAREN) {
		return nil
	}

	// 解析初始化表达式
	if p.peekToken.Type != lexer.TOKEN_SEMICOLON {
		p.nextToken()
		forStmt.Init = parseExpression(p, LOWEST)
	}

	if !p.expectPeek(lexer.TOKEN_SEMICOLON) {
		return nil
	}

	// 解析条件表达式
	if p.peekToken.Type != lexer.TOKEN_SEMICOLON {
		p.nextToken()
		forStmt.Test = parseExpression(p, LOWEST)
	}

	if !p.expectPeek(lexer.TOKEN_SEMICOLON) {
		return nil
	}

	// 解析更新表达式
	if p.peekToken.Type != lexer.TOKEN_RPAREN {
		p.nextToken()
		forStmt.Update = parseExpression(p, LOWEST)
	}

	if !p.expectPeek(lexer.TOKEN_RPAREN) {
		return nil
	}

	if !p.expectPeek(lexer.TOKEN_LBRACE) {
		return nil
	}

	forStmt.Body = parseBlockStatements(p)

	return forStmt
}

// parseFunctionDeclaration 解析函数声明
func parseFunctionDeclaration(p *Parser) *ast.FunctionDeclaration {
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

		// 解析第一个参数（支持类型提示）
		param := parseParameter(p)
		if param != nil {
			funcDecl.Parameters = append(funcDecl.Parameters, *param)
		}

		// 处理更多参数
		for p.peekToken.Type == lexer.TOKEN_COMMA {
			p.nextToken() // 移动到逗号
			p.nextToken() // 移动到下一个参数
			param := parseParameter(p)
			if param != nil {
				funcDecl.Parameters = append(funcDecl.Parameters, *param)
			}
		}
	}

	if !p.expectPeek(lexer.TOKEN_RPAREN) {
		return nil
	}

	// 检查是否有返回类型声明 ": type" 或 ": ?type"
	if p.peekToken.Type == lexer.TOKEN_COLON {
		p.nextToken() // 移动到 ':'

		// 检查是否为可空类型
		nullable := false
		if p.peekToken.Type == lexer.TOKEN_QUESTION {
			nullable = true
			p.nextToken() // 移动到 '?'
		}

		if !p.expectPeek(lexer.T_STRING) { // 期望类型名
			return nil
		}

		// 解析返回类型
		returnType := p.currentToken.Value
		if nullable {
			returnType = "?" + returnType
		}
		funcDecl.ReturnType = returnType
	}

	if !p.expectPeek(lexer.TOKEN_LBRACE) {
		return nil
	}

	funcDecl.Body = parseBlockStatements(p)

	return funcDecl
}

// parseParameter 解析函数参数（支持类型提示、可空类型和默认值）
func parseParameter(p *Parser) *ast.Parameter {
	var param ast.Parameter
	
	// 检查是否有可空类型提示
	nullable := false
	if p.currentToken.Type == lexer.TOKEN_QUESTION {
		nullable = true
		p.nextToken()
	}
	
	// 检查是否有类型提示
	if p.currentToken.Type == lexer.T_STRING {
		typeName := p.currentToken.Value
		if nullable {
			typeName = "?" + typeName
		}
		param.Type = typeName
		if !p.expectPeek(lexer.T_VARIABLE) {
			return nil
		}
	}
	
	// 解析参数名
	if p.currentToken.Type == lexer.T_VARIABLE {
		param.Name = p.currentToken.Value
	} else {
		return nil
	}
	
	// 检查是否有默认值
	if p.peekToken.Type == lexer.TOKEN_EQUAL {
		p.nextToken() // 移动到 '='
		p.nextToken() // 移动到默认值
		param.DefaultValue = parseExpression(p, LOWEST)
	}
	
	return &param
}

// parseReturnStatement 解析 return 语句
func parseReturnStatement(p *Parser) *ast.ReturnStatement {
	pos := p.currentToken.Position

	var returnValue ast.Expression

	if p.peekToken.Type != lexer.TOKEN_SEMICOLON {
		p.nextToken()
		returnValue = parseExpression(p, LOWEST)
	}

	if p.peekToken.Type == lexer.TOKEN_SEMICOLON {
		p.nextToken()
	}

	return ast.NewReturnStatement(pos, returnValue)
}

// parseBreakStatement 解析 break 语句
func parseBreakStatement(p *Parser) *ast.BreakStatement {
	pos := p.currentToken.Position

	if p.peekToken.Type == lexer.TOKEN_SEMICOLON {
		p.nextToken()
	}

	return ast.NewBreakStatement(pos)
}

// parseContinueStatement 解析 continue 语句
func parseContinueStatement(p *Parser) *ast.ContinueStatement {
	pos := p.currentToken.Position

	if p.peekToken.Type == lexer.TOKEN_SEMICOLON {
		p.nextToken()
	}

	return ast.NewContinueStatement(pos)
}

// parseBlockStatement 解析块语句
func parseBlockStatement(p *Parser) *ast.BlockStatement {
	pos := p.currentToken.Position
	block := ast.NewBlockStatement(pos)

	p.nextToken() // 跳过 {

	for !p.currentTokenIs(lexer.TOKEN_RBRACE) && !p.isAtEnd() {
		stmt := parseStatement(p)
		if stmt != nil {
			block.Body = append(block.Body, stmt)
		}
		p.nextToken()
	}

	return block
}

// parseBlockStatements 解析块语句中的语句列表
func parseBlockStatements(p *Parser) []ast.Statement {
	var statements []ast.Statement

	p.nextToken() // 跳过 {

	for !p.currentTokenIs(lexer.TOKEN_RBRACE) && !p.isAtEnd() {
		stmt := parseStatement(p)
		if stmt != nil {
			statements = append(statements, stmt)
		}
		p.nextToken()
	}

	return statements
}

// parseExpressionStatement 解析表达式语句
func parseExpressionStatement(p *Parser) *ast.ExpressionStatement {
	pos := p.currentToken.Position
	expr := parseExpression(p, LOWEST)

	if expr == nil {
		return nil
	}

	if p.peekToken.Type == lexer.TOKEN_SEMICOLON {
		p.nextToken()
	}

	return ast.NewExpressionStatement(pos, expr)
}

// parseExpression 解析表达式
func parseExpression(p *Parser, precedence Precedence) ast.Expression {
	prefix := globalPrefixParseFns[p.currentToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.currentToken.Type)
		return nil
	}

	leftExp := prefix(p)

	for p.peekToken.Type != lexer.TOKEN_SEMICOLON && precedence < p.peekPrecedence() {
		infix := globalInfixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()
		leftExp = infix(p, leftExp)
	}

	return leftExp
}

// 前缀解析函数

// parseVariable 解析变量
func parseVariable(p *Parser) ast.Expression {
	return ast.NewVariable(p.currentToken.Position, p.currentToken.Value)
}

// parseIdentifier 解析标识符
func parseIdentifier(p *Parser) ast.Expression {
	return ast.NewIdentifierNode(p.currentToken.Position, p.currentToken.Value)
}

// parseIntegerLiteral 解析整数字面量
func parseIntegerLiteral(p *Parser) ast.Expression {
	return ast.NewNumberLiteral(p.currentToken.Position, p.currentToken.Value, "integer")
}

// parseFloatLiteral 解析浮点数字面量
func parseFloatLiteral(p *Parser) ast.Expression {
	return ast.NewNumberLiteral(p.currentToken.Position, p.currentToken.Value, "float")
}

// parseStringLiteral 解析字符串字面量
func parseStringLiteral(p *Parser) ast.Expression {
	// 移除引号获取实际值
	value := strings.Trim(p.currentToken.Value, `"'`)
	return ast.NewStringLiteral(p.currentToken.Position, value, p.currentToken.Value)
}

// parsePrefixExpression 解析前缀表达式
func parsePrefixExpression(p *Parser) ast.Expression {
	pos := p.currentToken.Position
	operator := p.currentToken.Value

	p.nextToken()
	operand := parseExpression(p, PREFIX)

	return ast.NewUnaryExpression(pos, operator, operand, true)
}

// parseTypeCast 解析类型转换表达式
func parseTypeCast(p *Parser) ast.Expression {
	pos := p.currentToken.Position
	castType := p.currentToken.Value

	p.nextToken()
	operand := parseExpression(p, PREFIX)

	return ast.NewCastExpression(pos, castType, operand)
}

// parseGroupedExpression 解析括号表达式
func parseGroupedExpression(p *Parser) ast.Expression {
	p.nextToken()

	exp := parseExpression(p, LOWEST)

	if !p.expectPeek(lexer.TOKEN_RPAREN) {
		return nil
	}

	return exp
}

// parseArrayExpression 解析数组表达式
func parseArrayExpression(p *Parser) ast.Expression {
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
	array.Elements = append(array.Elements, parseExpression(p, LOWEST))

	for p.peekToken.Type == lexer.TOKEN_COMMA {
		p.nextToken() // 移动到逗号
		p.nextToken() // 移动到下一个元素
		array.Elements = append(array.Elements, parseExpression(p, LOWEST))
	}

	if !p.expectPeek(lexer.TOKEN_RPAREN) {
		return nil
	}

	return array
}

// 中缀解析函数

// parseInfixExpression 解析中缀表达式
func parseInfixExpression(p *Parser, left ast.Expression) ast.Expression {
	pos := p.currentToken.Position
	operator := p.currentToken.Value
	precedence := p.curPrecedence()

	p.nextToken()
	right := parseExpression(p, precedence)

	return ast.NewBinaryExpression(pos, left, operator, right)
}

// parseAssignmentExpression 解析赋值表达式
func parseAssignmentExpression(p *Parser, left ast.Expression) ast.Expression {
	pos := p.currentToken.Position
	operator := p.currentToken.Value

	p.nextToken()
	right := parseExpression(p, LOWEST)

	return ast.NewAssignmentExpression(pos, left, operator, right)
}

// parsePostfixExpression 解析后缀表达式
func parsePostfixExpression(p *Parser, left ast.Expression) ast.Expression {
	pos := p.currentToken.Position
	operator := p.currentToken.Value

	return ast.NewUnaryExpression(pos, operator, left, false)
}

// parseCallExpression 解析函数调用表达式
func parseCallExpression(p *Parser, fn ast.Expression) ast.Expression {
	pos := p.currentToken.Position
	call := ast.NewCallExpression(pos, fn)

	// 解析参数列表
	call.Arguments = parseExpressionList(p, lexer.TOKEN_RPAREN)

	return call
}

// parseExpressionList 解析表达式列表（用于函数调用参数等）
func parseExpressionList(p *Parser, end lexer.TokenType) []ast.Expression {
	var args []ast.Expression

	if p.peekToken.Type == end {
		p.nextToken()
		return args
	}

	p.nextToken()
	// Skip comments before parsing expression
	for p.currentToken.Type == lexer.T_COMMENT {
		p.nextToken()
	}
	args = append(args, parseExpression(p, LOWEST))

	for p.peekToken.Type == lexer.TOKEN_COMMA {
		p.nextToken()
		p.nextToken()
		// Skip comments after comma
		for p.currentToken.Type == lexer.T_COMMENT {
			p.nextToken()
		}
		args = append(args, parseExpression(p, LOWEST))
	}

	if !p.expectPeek(end) {
		return nil
	}

	return args
}

// PHP 特定的解析函数

// parseInlineHTML 解析内联HTML
func parseInlineHTML(p *Parser) ast.Expression {
	// 对于内联HTML，创建一个特殊的字面量表达式
	return ast.NewStringLiteral(p.currentToken.Position, p.currentToken.Value, p.currentToken.Value)
}

// parseOpenTag 解析PHP开放标签
func parseOpenTag(p *Parser) ast.Expression {
	// PHP开放标签通常不作为表达式使用，但为了完整性，创建一个特殊节点
	return ast.NewStringLiteral(p.currentToken.Position, p.currentToken.Value, p.currentToken.Value)
}

// parseComment 解析注释
func parseComment(p *Parser) ast.Expression {
	// 注释也创建为特殊的字面量
	return ast.NewStringLiteral(p.currentToken.Position, p.currentToken.Value, p.currentToken.Value)
}

// parseHeredoc 解析Heredoc
func parseHeredoc(p *Parser) ast.Expression {
	startPos := p.currentToken.Position
	startValue := p.currentToken.Value

	// 构建完整的heredoc内容
	var content strings.Builder
	content.WriteString(startValue)

	// 期望下一个token是T_ENCAPSED_AND_WHITESPACE或T_END_HEREDOC
	p.nextToken()

	// 收集heredoc内容（包括变量）
	for p.currentToken.Type == lexer.T_ENCAPSED_AND_WHITESPACE || p.currentToken.Type == lexer.T_VARIABLE {
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
func parseEndHeredoc(p *Parser) ast.Expression {
	// 这通常不应该被单独调用，因为T_END_HEREDOC应该在parseHeredoc中处理
	// 但为了避免"no prefix parse function"错误，提供一个基本实现
	return ast.NewStringLiteral(p.currentToken.Position, p.currentToken.Value, p.currentToken.Value)
}

// parseDocBlockComment 解析文档块注释
func parseDocBlockComment(p *Parser) ast.Expression {
	// 创建文档块注释节点
	return ast.NewDocBlockComment(p.currentToken.Position, p.currentToken.Value, p.currentToken.Value)
}

// parseGlobalStatement 解析全局变量声明
func parseGlobalStatement(p *Parser) ast.Statement {
	pos := p.currentToken.Position
	globalStmt := ast.NewGlobalStatement(pos)

	// 跳过 'global' 关键字
	p.nextToken()

	// 解析变量列表
	for {
		if p.currentToken.Type == lexer.T_VARIABLE {
			variable := parseVariable(p)
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
func parseComma(p *Parser) ast.Expression {
	// 逗号通常不应该作为表达式单独解析，但为了避免解析错误，返回一个简单的字符串字面量
	return ast.NewStringLiteral(p.currentToken.Position, ",", ",")
}

// parseStaticStatement 解析静态变量声明
func parseStaticStatement(p *Parser) ast.Statement {
	pos := p.currentToken.Position
	staticStmt := ast.NewStaticStatement(pos)

	// 跳过 'static' 关键字
	p.nextToken()

	// 解析变量列表
	for {
		if p.currentToken.Type == lexer.T_VARIABLE {
			variable := parseVariable(p)

			var defaultValue ast.Expression = nil
			p.nextToken()

			// 检查是否有默认值
			if p.currentToken.Type == lexer.TOKEN_EQUAL {
				p.nextToken() // 跳过 =
				defaultValue = parseExpression(p, LOWEST)
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
func parseUnsetStatement(p *Parser) ast.Statement {
	pos := p.currentToken.Position
	unsetStmt := ast.NewUnsetStatement(pos)

	// 跳过 'unset' 关键字
	if !p.expectPeek(lexer.TOKEN_LPAREN) {
		return nil
	}

	p.nextToken() // 进入括号内部

	// 解析变量列表
	for p.currentToken.Type != lexer.TOKEN_RPAREN && p.currentToken.Type != lexer.T_EOF {
		variable := parseExpression(p, LOWEST)
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
func parseDoWhileStatement(p *Parser) ast.Statement {
	pos := p.currentToken.Position

	// 跳过 'do'
	p.nextToken()

	// 解析循环体
	body := parseStatement(p)

	// 期望 'while'
	if !p.expectPeek(lexer.T_WHILE) {
		return nil
	}

	// 期望 '('
	if !p.expectPeek(lexer.TOKEN_LPAREN) {
		return nil
	}

	p.nextToken()
	condition := parseExpression(p, LOWEST)

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
func parseForeachStatement(p *Parser) ast.Statement {
	pos := p.currentToken.Position

	// 跳过 'foreach'
	if !p.expectPeek(lexer.TOKEN_LPAREN) {
		return nil
	}

	p.nextToken()
	iterable := parseExpression(p, LOWEST)

	// 期望 'as'
	if !p.expectPeek(lexer.T_AS) {
		return nil
	}

	p.nextToken()
	var key, value ast.Expression

	// 解析第一个变量
	firstVar := parseExpression(p, LOWEST)

	p.nextToken()
	// 检查是否有 '=>'（表示有key）
	if p.currentToken.Type == lexer.T_DOUBLE_ARROW {
		key = firstVar
		p.nextToken()
		value = parseExpression(p, LOWEST)
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
	body := parseStatement(p)

	return ast.NewForeachStatement(pos, iterable, key, value, body)
}

// parseSwitchStatement 解析switch语句
func parseSwitchStatement(p *Parser) ast.Statement {
	pos := p.currentToken.Position

	// 跳过 'switch'
	if !p.expectPeek(lexer.TOKEN_LPAREN) {
		return nil
	}

	p.nextToken()
	discriminant := parseExpression(p, LOWEST)

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
			test := parseExpression(p, LOWEST)

			if !p.expectPeek(lexer.TOKEN_COLON) {
				continue
			}

			switchCase := ast.NewSwitchCase(casePos, test)

			p.nextToken()
			// 解析case体内的语句
			for p.currentToken.Type != lexer.T_CASE && p.currentToken.Type != lexer.T_DEFAULT &&
				p.currentToken.Type != lexer.TOKEN_RBRACE && p.currentToken.Type != lexer.T_EOF {
				stmt := parseStatement(p)
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
				stmt := parseStatement(p)
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
func parseTryStatement(p *Parser) ast.Statement {
	pos := p.currentToken.Position
	tryStmt := ast.NewTryStatement(pos)

	// 跳过 'try'
	if !p.expectPeek(lexer.TOKEN_LBRACE) {
		return nil
	}

	p.nextToken()
	// 解析try块
	for p.currentToken.Type != lexer.TOKEN_RBRACE && p.currentToken.Type != lexer.T_EOF {
		stmt := parseStatement(p)
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
			exceptionType := parseExpression(p, LOWEST)
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
			parameter = parseVariable(p)
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
			stmt := parseStatement(p)
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
			stmt := parseStatement(p)
			if stmt != nil {
				tryStmt.FinallyBlock = append(tryStmt.FinallyBlock, stmt)
			}
			p.nextToken()
		}
	}

	return tryStmt
}

// parseThrowStatement 解析throw语句
func parseThrowStatement(p *Parser) ast.Statement {
	pos := p.currentToken.Position

	p.nextToken()
	argument := parseExpression(p, LOWEST)

	if !p.expectPeek(lexer.TOKEN_SEMICOLON) {
		return nil
	}

	return ast.NewThrowStatement(pos, argument)
}

// parseGotoStatement 解析goto语句
func parseGotoStatement(p *Parser) ast.Statement {
	pos := p.currentToken.Position

	p.nextToken()
	if p.currentToken.Type != lexer.T_STRING {
		p.errors = append(p.errors, "expected label name in goto statement")
		return nil
	}

	label := parseIdentifier(p)

	if !p.expectPeek(lexer.TOKEN_SEMICOLON) {
		return nil
	}

	return ast.NewGotoStatement(pos, label)
}

// parseNewExpression 解析new表达式
func parseNewExpression(p *Parser) ast.Expression {
	pos := p.currentToken.Position

	p.nextToken()
	class := parseExpression(p, PREFIX)

	newExpr := ast.NewNewExpression(pos, class)

	// 检查是否有构造函数参数
	if p.peekToken.Type == lexer.TOKEN_LPAREN {
		p.nextToken() // 移动到 (
		p.nextToken() // 进入括号内部

		// 解析参数列表
		for p.currentToken.Type != lexer.TOKEN_RPAREN && p.currentToken.Type != lexer.T_EOF {
			arg := parseExpression(p, LOWEST)
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
func parseCloneExpression(p *Parser) ast.Expression {
	pos := p.currentToken.Position

	p.nextToken()
	object := parseExpression(p, PREFIX)

	return ast.NewCloneExpression(pos, object)
}

// parseInstanceofExpression 解析instanceof表达式
func parseInstanceofExpression(p *Parser, left ast.Expression) ast.Expression {
	pos := p.currentToken.Position

	p.nextToken()
	right := parseExpression(p, LESSGREATER)

	return ast.NewInstanceofExpression(pos, left, right)
}

// parsePropertyAccess 解析属性访问表达式
func parsePropertyAccess(p *Parser, left ast.Expression) ast.Expression {
	pos := p.currentToken.Position

	p.nextToken()
	property := parseExpression(p, CALL)

	return ast.NewPropertyAccessExpression(pos, left, property)
}

// parseErrorSuppression 解析错误抑制操作符 @
func parseErrorSuppression(p *Parser) ast.Expression {
	pos := p.currentToken.Position

	p.nextToken()
	expr := parseExpression(p, PREFIX)

	return ast.NewErrorSuppressionExpression(pos, expr)
}

// parseCoalesceExpression 解析 null 合并操作符 ??
func parseCoalesceExpression(p *Parser, left ast.Expression) ast.Expression {
	pos := p.currentToken.Position

	precedence := p.getCurrentPrecedence()
	p.nextToken()
	right := parseExpression(p, precedence)

	return ast.NewCoalesceExpression(pos, left, right)
}

// parseBooleanExpression 解析布尔逻辑操作符 && ||
func parseBooleanExpression(p *Parser, left ast.Expression) ast.Expression {
	pos := p.currentToken.Position
	operator := p.currentToken.Value

	precedence := p.getCurrentPrecedence()
	p.nextToken()
	right := parseExpression(p, precedence)

	return ast.NewBinaryExpression(pos, left, operator, right)
}

// parseArrayAccess 解析数组访问表达式 []
func parseArrayAccess(p *Parser, left ast.Expression) ast.Expression {
	pos := p.currentToken.Position

	p.nextToken()

	var index *ast.Expression
	if p.currentToken.Type != lexer.TOKEN_RBRACKET {
		expr := parseExpression(p, LOWEST)
		index = &expr

		if !p.expectPeek(lexer.TOKEN_RBRACKET) {
			return nil
		}
	}

	return ast.NewArrayAccessExpression(pos, left, index)
}

// parseEmptyExpression 解析 empty() 函数
func parseEmptyExpression(p *Parser) ast.Expression {
	pos := p.currentToken.Position

	if !p.expectToken(lexer.TOKEN_LPAREN) {
		return nil
	}

	p.nextToken()
	expr := parseExpression(p, LOWEST)

	if !p.expectToken(lexer.TOKEN_RPAREN) {
		return nil
	}

	return ast.NewEmptyExpression(pos, expr)
}

// parseArrayLiteral 解析数组字面量 [] - 根据 PHP 官方语法
func parseArrayLiteral(p *Parser) ast.Expression {
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
		element := parseExpression(p, LOWEST)
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
func parseFallback(p *Parser) ast.Expression {
	// 对于标点符号，通常意味着表达式结束
	// 返回 nil 让上级处理
	pos := p.currentToken.Position
	value := p.currentToken.Value

	// 创建一个简单的标识符作为占位符
	return ast.NewIdentifierNode(pos, value)
}

// ============== 新增的解析函数实现 ==============

// parseExitExpression 解析 exit/die 表达式
func parseExitExpression(p *Parser) ast.Expression {
	pos := p.currentToken.Position

	var argument ast.Expression

	// exit 可能带有括号和参数，也可能不带
	if p.peekToken.Type == lexer.TOKEN_LPAREN {
		p.nextToken() // 移动到 (
		if p.peekToken.Type != lexer.TOKEN_RPAREN {
			p.nextToken() // 进入括号
			argument = parseExpression(p, LOWEST)
		}
		if p.peekToken.Type == lexer.TOKEN_RPAREN {
			p.nextToken() // 跳过 )
		}
	} else if p.peekToken.Type != lexer.TOKEN_SEMICOLON &&
		p.peekToken.Type != lexer.T_EOF {
		// exit 后直接跟表达式，不带括号
		p.nextToken()
		argument = parseExpression(p, LOWEST)
	}

	return ast.NewExitExpression(pos, argument)
}

// parseIssetExpression 解析 isset() 表达式
func parseIssetExpression(p *Parser) ast.Expression {
	pos := p.currentToken.Position

	if !p.expectToken(lexer.TOKEN_LPAREN) {
		return nil
	}

	var arguments []ast.Expression

	// 解析参数列表
	if p.peekToken.Type != lexer.TOKEN_RPAREN {
		p.nextToken()
		arguments = append(arguments, parseExpression(p, LOWEST))

		// 处理多个参数
		for p.peekToken.Type == lexer.TOKEN_COMMA {
			p.nextToken() // 跳过逗号
			p.nextToken() // 移动到下一个参数
			arguments = append(arguments, parseExpression(p, LOWEST))
		}
	}

	if !p.expectToken(lexer.TOKEN_RPAREN) {
		return nil
	}

	return ast.NewIssetExpression(pos, arguments)
}

// parseListExpression 解析 list() 表达式
func parseListExpression(p *Parser) ast.Expression {
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
				elements = append(elements, parseExpression(p, LOWEST))
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
func parseAnonymousFunctionExpression(p *Parser) ast.Expression {
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
					useClause = append(useClause, parseVariable(p))
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

	body := parseBlockStatements(p)

	return ast.NewAnonymousFunctionExpression(pos, parameters, body, useClause)
}

// parseUseExpression 解析 use 表达式（在表达式上下文中）
func parseUseExpression(p *Parser) ast.Expression {
	pos := p.currentToken.Position

	// 这里简单处理为标识符，实际的 use 语句应该在语句级别处理
	return ast.NewIdentifierNode(pos, "use")
}

// parseNowdocExpression 解析 nowdoc 表达式
func parseNowdocExpression(p *Parser) ast.Expression {
	pos := p.currentToken.Position
	var contentBuilder strings.Builder
	
	// 跳过 T_NOWDOC 开始标记，收集内容
	p.nextToken()
	
	// 收集 T_ENCAPSED_AND_WHITESPACE 内容
	for p.currentToken.Type == lexer.T_ENCAPSED_AND_WHITESPACE {
		contentBuilder.WriteString(p.currentToken.Value)
		p.nextToken()
	}
	
	// 期望 T_END_HEREDOC 结束标记
	if p.currentToken.Type == lexer.T_END_HEREDOC {
		// 不需要添加结束标记到内容中，只是验证它存在
	} else {
		p.errors = append(p.errors, fmt.Sprintf("expected T_END_HEREDOC, got %s instead", p.currentToken.Type))
	}
	
	content := contentBuilder.String()
	return ast.NewStringLiteral(pos, content, content)
}

// parseCurlyOpenExpression 解析字符串插值开始标记
func parseCurlyOpenExpression(p *Parser) ast.Expression {
	pos := p.currentToken.Position

	// 这通常在字符串插值中出现，这里简单处理
	return ast.NewStringLiteral(pos, p.currentToken.Value, p.currentToken.Value)
}

// parseTernaryExpression 解析三元运算符表达式
func parseTernaryExpression(p *Parser, condition ast.Expression) ast.Expression {
	pos := p.currentToken.Position

	p.nextToken()

	var consequent ast.Expression

	// 检查是否是短三元运算符 ?:
	if p.currentToken.Type == lexer.TOKEN_COLON {
		// 短三元运算符，consequent 为 nil
		consequent = nil
	} else {
		consequent = parseExpression(p, LOWEST)

		// 期望冒号
		if !p.expectToken(lexer.TOKEN_COLON) {
			return nil
		}
	}

	p.nextToken()
	alternate := parseExpression(p, TERNARY)

	return ast.NewTernaryExpression(pos, condition, consequent, alternate)
}

// parseDoubleArrowExpression 解析 => 表达式（数组元素）
func parseDoubleArrowExpression(p *Parser, key ast.Expression) ast.Expression {
	pos := p.currentToken.Position

	p.nextToken()
	value := parseExpression(p, TERNARY)

	return ast.NewArrayElementExpression(pos, key, value)
}

// ============== 新增的缺失函数实现 ==============

// parseStringConcatenation 解析字符串连接（当 . 作为前缀使用时）
func parseStringConcatenation(p *Parser) ast.Expression {
	pos := p.currentToken.Position

	// . 作为前缀通常意味着错误的语法，返回占位符
	return ast.NewIdentifierNode(pos, ".")
}

// parseReferenceExpression 解析引用表达式 &$var
func parseReferenceExpression(p *Parser) ast.Expression {
	pos := p.currentToken.Position

	p.nextToken()
	operand := parseExpression(p, PREFIX)

	return ast.NewUnaryExpression(pos, "&", operand, true)
}

// parseCaseExpression 解析 case 表达式（switch 语句中的 case）
func parseCaseExpression(p *Parser) ast.Expression {
	pos := p.currentToken.Position

	p.nextToken()
	test := parseExpression(p, LOWEST)

	// 返回一个特殊的case表达式节点
	return ast.NewCaseExpression(pos, test)
}

// parseClassExpression 解析类表达式或类声明
func parseClassExpression(p *Parser) ast.Expression {
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
func parseConstExpression(p *Parser) ast.Expression {
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
func parseDefaultExpression(p *Parser) ast.Expression {
	pos := p.currentToken.Position

	// default case 没有测试表达式
	return ast.NewCaseExpression(pos, nil)
}

// parseEvalExpression 解析 eval() 表达式
func parseEvalExpression(p *Parser) ast.Expression {
	pos := p.currentToken.Position

	var argument ast.Expression

	// eval 可能带有括号和参数
	if p.peekToken.Type == lexer.TOKEN_LPAREN {
		p.nextToken() // 移动到 (
		if p.peekToken.Type != lexer.TOKEN_RPAREN {
			p.nextToken() // 进入括号
			argument = parseExpression(p, LOWEST)
		}

		if p.peekToken.Type == lexer.TOKEN_RPAREN {
			p.nextToken() // 跳过右括号
		}
	}

	return ast.NewEvalExpression(pos, argument)
}

// parseStaticAccess 解析静态访问表达式 Class::method 或 Class::$property
func parseStaticAccess(p *Parser) ast.Expression {
	pos := p.currentToken.Position

	// :: 作为前缀通常意味着省略了左边的类名，如 ::method()
	p.nextToken()

	var property ast.Expression
	if p.currentToken.Type != lexer.T_EOF {
		property = parseExpression(p, CALL)
	}

	// 创建一个静态访问表达式，左边为 nil 表示省略了类名
	return ast.NewStaticAccessExpression(pos, nil, property)
}

// parseInterpolatedString 解析包含变量插值的双引号字符串
func parseInterpolatedString(p *Parser) ast.Expression {
	pos := p.currentToken.Position
	var parts []ast.Expression

	// 跳过开始的引号
	p.nextToken()

	// 解析字符串内容，直到遇到结束的引号
	for p.currentToken.Type != lexer.TOKEN_QUOTE && !p.isAtEnd() {
		switch p.currentToken.Type {
		case lexer.T_VARIABLE:
			// 变量插值
			variable := ast.NewVariable(p.currentToken.Position, p.currentToken.Value)
			parts = append(parts, variable)
		case lexer.T_ENCAPSED_AND_WHITESPACE:
			// 字符串片段
			if p.currentToken.Value != "" {
				stringPart := ast.NewStringLiteral(p.currentToken.Position, p.currentToken.Value, p.currentToken.Value)
				parts = append(parts, stringPart)
			}
		default:
			// 其他内容，当作字符串处理
			if p.currentToken.Value != "" {
				stringPart := ast.NewStringLiteral(p.currentToken.Position, p.currentToken.Value, p.currentToken.Value)
				parts = append(parts, stringPart)
			}
		}
		p.nextToken()
	}

	// 如果只有一个部分且是简单字符串，返回字符串字面量
	if len(parts) == 1 {
		if stringLit, ok := parts[0].(*ast.StringLiteral); ok {
			return stringLit
		}
	}

	// 返回字符串插值表达式
	return ast.NewInterpolatedStringExpression(pos, parts)
}

// parseVisibilityModifier 解析可见性修饰符 public/private/protected
func parseVisibilityModifier(p *Parser) ast.Expression {
	pos := p.currentToken.Position
	modifier := p.currentToken.Value

	// 可见性修饰符后面应该跟着属性或方法声明
	return ast.NewVisibilityModifierExpression(pos, modifier)
}
