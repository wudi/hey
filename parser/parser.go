package parser

import (
	"fmt"
	"strings"

	"github.com/wudi/php-parser/ast"
	"github.com/wudi/php-parser/lexer"
)

// Precedence 操作符优先级 - 根据 PHP 官方优先级顺序
type Precedence int

const (
	_ Precedence = iota
	LOWEST
	ASSIGN        // = += -= *= /= .= %= &= |= ^= <<= >>= **= ??=
	TERNARY       // ? :
	COALESCE      // ??
	LOGICAL_OR    // || or
	LOGICAL_AND   // && and
	BITWISE_OR    // |
	BITWISE_XOR   // ^
	BITWISE_AND   // &
	EQUALS        // == != === !==
	LESSGREATER   // < <= > >= <=> instanceof
	BITWISE_SHIFT // << >>
	SUM           // + - .
	PRODUCT       // * / %
	EXPONENT      // **
	PREFIX        // ! ~ -X +X ++X --X (int) (float) (string) (array) (object) (bool) @
	POSTFIX       // X++ X--
	CALL          // myFunction(X) $obj->prop Class::$prop
	INDEX         // array[index]
)

// 操作符优先级映射
var precedences = map[lexer.TokenType]Precedence{
	// 赋值运算符
	lexer.TOKEN_EQUAL:      ASSIGN,
	lexer.T_PLUS_EQUAL:     ASSIGN,
	lexer.T_MINUS_EQUAL:    ASSIGN,
	lexer.T_MUL_EQUAL:      ASSIGN,
	lexer.T_DIV_EQUAL:      ASSIGN,
	lexer.T_CONCAT_EQUAL:   ASSIGN,
	lexer.T_MOD_EQUAL:      ASSIGN,
	lexer.T_AND_EQUAL:      ASSIGN,
	lexer.T_OR_EQUAL:       ASSIGN,
	lexer.T_XOR_EQUAL:      ASSIGN,
	lexer.T_SL_EQUAL:       ASSIGN,
	lexer.T_SR_EQUAL:       ASSIGN,
	lexer.T_POW_EQUAL:      ASSIGN,
	lexer.T_COALESCE_EQUAL: ASSIGN,

	// 三元运算符
	lexer.TOKEN_QUESTION: TERNARY,
	lexer.T_DOUBLE_ARROW: TERNARY,

	// 空合并运算符
	lexer.T_COALESCE: COALESCE,

	// 逻辑运算符
	lexer.T_BOOLEAN_OR:  LOGICAL_OR,
	lexer.T_LOGICAL_OR:  LOGICAL_OR,
	lexer.T_BOOLEAN_AND: LOGICAL_AND,
	lexer.T_LOGICAL_AND: LOGICAL_AND,
	lexer.T_LOGICAL_XOR: LOGICAL_AND, // xor has same precedence as and

	// 位运算符
	lexer.TOKEN_PIPE:      BITWISE_OR,
	lexer.TOKEN_CARET:     BITWISE_XOR,
	lexer.TOKEN_AMPERSAND: BITWISE_AND,

	// 比较运算符
	lexer.T_IS_EQUAL:            EQUALS,
	lexer.T_IS_NOT_EQUAL:        EQUALS,
	lexer.T_IS_IDENTICAL:        EQUALS,
	lexer.T_IS_NOT_IDENTICAL:    EQUALS,
	lexer.TOKEN_LT:              LESSGREATER,
	lexer.TOKEN_GT:              LESSGREATER,
	lexer.T_IS_SMALLER_OR_EQUAL: LESSGREATER,
	lexer.T_IS_GREATER_OR_EQUAL: LESSGREATER,
	lexer.T_SPACESHIP:           LESSGREATER,
	lexer.T_INSTANCEOF:          LESSGREATER,

	// 位移运算符
	lexer.T_SR: BITWISE_SHIFT,
	lexer.T_SL: BITWISE_SHIFT,

	// 加减运算符
	lexer.TOKEN_PLUS:  SUM,
	lexer.TOKEN_MINUS: SUM,
	lexer.TOKEN_DOT:   SUM, // 字符串连接

	// 乘除运算符
	lexer.TOKEN_DIVIDE:   PRODUCT,
	lexer.TOKEN_MULTIPLY: PRODUCT,
	lexer.TOKEN_MODULO:   PRODUCT,

	// 幂运算符
	lexer.T_POW: EXPONENT,

	// 后缀运算符
	lexer.T_INC: POSTFIX,
	lexer.T_DEC: POSTFIX,

	// 函数调用和属性访问
	lexer.T_OBJECT_OPERATOR:          CALL,
	lexer.T_NULLSAFE_OBJECT_OPERATOR: CALL,
	lexer.T_PAAMAYIM_NEKUDOTAYIM:     CALL,
	lexer.TOKEN_LPAREN:               CALL,

	// 数组访问
	lexer.TOKEN_LBRACKET: INDEX,
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
		lexer.T_NAME_FULLY_QUALIFIED:     parseIdentifier,
		lexer.T_NAME_QUALIFIED:           parseIdentifier,
		lexer.T_NAME_RELATIVE:            parseIdentifier,
		lexer.TOKEN_EXCLAMATION:          parsePrefixExpression,
		lexer.TOKEN_PLUS:                 parsePrefixExpression, // 一元正号操作符 +
		lexer.TOKEN_MINUS:                parsePrefixExpression,
		lexer.TOKEN_DOLLAR:               parseDollarBraceExpression, // ${...} syntax
		lexer.TOKEN_TILDE:                parsePrefixExpression,      // 位运算NOT操作符 ~
		lexer.T_INC:                      parsePrefixExpression,
		lexer.T_DEC:                      parsePrefixExpression,
		lexer.TOKEN_LPAREN:               parseGroupedExpression,
		lexer.TOKEN_LBRACE:               parseCurlyBraceExpression,
		lexer.T_ARRAY:                    parseArrayExpression,
		lexer.T_INLINE_HTML:              parseInlineHTML,
		lexer.T_OPEN_TAG:                 parseOpenTag,
		lexer.T_COMMENT:                  parseComment,
		lexer.T_DOC_COMMENT:              parseDocBlockComment,
		lexer.T_START_HEREDOC:            parseHeredoc,
		lexer.T_ENCAPSED_AND_WHITESPACE:  parseStringLiteral,
		lexer.T_END_HEREDOC:              parseEndHeredoc,
		lexer.TOKEN_COMMA:                parseComma,
		lexer.TOKEN_BACKTICK:             parseShellExecExpression,
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
		lexer.T_CURLY_OPEN:               parseCurlyOpenExpression,
		lexer.TOKEN_QUOTE:                parseInterpolatedString,
		lexer.T_INCLUDE:                  parseIncludeOrEvalExpression,
		lexer.T_INCLUDE_ONCE:             parseIncludeOrEvalExpression,
		lexer.T_REQUIRE:                  parseIncludeOrEvalExpression,
		lexer.T_REQUIRE_ONCE:             parseIncludeOrEvalExpression,
		lexer.T_STATIC:                   parseStaticOrArrowFunctionExpression,
		lexer.T_ABSTRACT:                 parseAbstractExpression,
		lexer.T_FINAL:                    parseFinalExpression,
		lexer.T_CLOSE_TAG:                parseCloseTagExpression,
		lexer.T_NS_SEPARATOR:             parseNamespaceExpression,
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
		lexer.T_MATCH:                    parseMatchToken,
		lexer.T_ENUM:                     parseIdentifier,
		lexer.T_PAAMAYIM_NEKUDOTAYIM:     parseStaticAccess,
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
		lexer.T_ATTRIBUTE:                parseAttributeExpression,
		lexer.T_FN:                       parseArrowFunctionExpression,
		lexer.T_YIELD:                    parseYieldExpression,
		lexer.T_YIELD_FROM:               parseYieldFromExpression,
		lexer.T_THROW:                    parseThrowExpression,
		lexer.T_POW:                      parseFallback, // ** 作为前缀时是无效的，但需要处理
	}
	globalInfixParseFns = map[lexer.TokenType]InfixParseFn{
		// 二元运算符
		lexer.TOKEN_PLUS:     parseInfixExpression,
		lexer.TOKEN_MINUS:    parseInfixExpression,
		lexer.TOKEN_DIVIDE:   parseInfixExpression,
		lexer.TOKEN_MULTIPLY: parseInfixExpression,
		lexer.TOKEN_MODULO:   parseInfixExpression,
		lexer.T_POW:          parseInfixExpression,
		lexer.TOKEN_DOT:      parseInfixExpression,

		// 比较运算符
		lexer.T_IS_EQUAL:            parseInfixExpression,
		lexer.T_IS_NOT_EQUAL:        parseInfixExpression,
		lexer.T_IS_IDENTICAL:        parseInfixExpression,
		lexer.T_IS_NOT_IDENTICAL:    parseInfixExpression,
		lexer.TOKEN_LT:              parseInfixExpression,
		lexer.TOKEN_GT:              parseInfixExpression,
		lexer.T_IS_SMALLER_OR_EQUAL: parseInfixExpression,
		lexer.T_IS_GREATER_OR_EQUAL: parseInfixExpression,
		lexer.T_SPACESHIP:           parseInfixExpression,
		lexer.T_INSTANCEOF:          parseInstanceofExpression,

		// 位运算符
		lexer.T_SR:            parseInfixExpression, // >> (right shift)
		lexer.T_SL:            parseInfixExpression, // << (left shift)
		lexer.TOKEN_AMPERSAND: parseInfixExpression, // & (bitwise AND)
		lexer.TOKEN_PIPE:      parseInfixExpression, // | (bitwise OR)
		lexer.TOKEN_CARET:     parseInfixExpression, // ^ (bitwise XOR)

		// 逻辑运算符
		lexer.T_BOOLEAN_AND: parseBooleanExpression,
		lexer.T_BOOLEAN_OR:  parseBooleanExpression,
		lexer.T_LOGICAL_AND: parseBooleanExpression,
		lexer.T_LOGICAL_OR:  parseBooleanExpression,
		lexer.T_LOGICAL_XOR: parseBooleanExpression,

		// 赋值运算符
		lexer.TOKEN_EQUAL:      parseAssignmentExpression,
		lexer.T_PLUS_EQUAL:     parseAssignmentExpression,
		lexer.T_MINUS_EQUAL:    parseAssignmentExpression,
		lexer.T_MUL_EQUAL:      parseAssignmentExpression,
		lexer.T_DIV_EQUAL:      parseAssignmentExpression,
		lexer.T_CONCAT_EQUAL:   parseAssignmentExpression,
		lexer.T_MOD_EQUAL:      parseAssignmentExpression,
		lexer.T_AND_EQUAL:      parseAssignmentExpression,
		lexer.T_OR_EQUAL:       parseAssignmentExpression,
		lexer.T_XOR_EQUAL:      parseAssignmentExpression,
		lexer.T_SL_EQUAL:       parseAssignmentExpression,
		lexer.T_SR_EQUAL:       parseAssignmentExpression,
		lexer.T_POW_EQUAL:      parseAssignmentExpression,
		lexer.T_COALESCE_EQUAL: parseAssignmentExpression,

		// 其他运算符
		lexer.T_COALESCE:     parseCoalesceExpression,
		lexer.TOKEN_QUESTION: parseTernaryExpression,
		lexer.T_DOUBLE_ARROW: parseDoubleArrowExpression,

		// 后缀运算符
		lexer.T_INC: parsePostfixExpression,
		lexer.T_DEC: parsePostfixExpression,

		// 访问运算符
		lexer.T_OBJECT_OPERATOR:          parsePropertyAccess,
		lexer.T_NULLSAFE_OBJECT_OPERATOR: parseNullsafePropertyAccess,
		lexer.T_PAAMAYIM_NEKUDOTAYIM:     parseStaticAccessExpression,
		lexer.TOKEN_LPAREN:               parseCallExpression,
		lexer.TOKEN_LBRACKET:             parseArrayAccess,
	}
}

type PrefixParseFn func(*Parser) ast.Expression
type InfixParseFn func(*Parser, ast.Expression) ast.Expression

// isNonSyntacticToken 判断token是否在语法解析中无意义，应该被跳过
func isNonSyntacticToken(tokenType lexer.TokenType) bool {
	switch tokenType {
	case lexer.T_COMMENT,
		lexer.T_DOC_COMMENT:
		// 注释token在语法解析中无意义，但保留在token流中供工具使用
		return true
	default:
		return false
	}
}

// isValidIdentifier 判断token是否可以作为标识符（如命名参数）
// 基于PHP语法中的identifier规则，支持T_STRING和semi_reserved tokens
func isValidIdentifier(tokenType lexer.TokenType) bool {
	switch tokenType {
	// T_STRING - regular identifiers
	case lexer.T_STRING:
		return true
	// reserved_non_modifiers (from PHP grammar)
	case lexer.T_INCLUDE, lexer.T_INCLUDE_ONCE, lexer.T_EVAL, lexer.T_REQUIRE, lexer.T_REQUIRE_ONCE,
		lexer.T_LOGICAL_OR, lexer.T_LOGICAL_XOR, lexer.T_LOGICAL_AND,
		lexer.T_INSTANCEOF, lexer.T_NEW, lexer.T_CLONE, lexer.T_EXIT, lexer.T_IF, lexer.T_ELSEIF,
		lexer.T_ELSE, lexer.T_ENDIF, lexer.T_ECHO, lexer.T_DO, lexer.T_WHILE, lexer.T_ENDWHILE,
		lexer.T_FOR, lexer.T_ENDFOR, lexer.T_FOREACH, lexer.T_ENDFOREACH, lexer.T_DECLARE,
		lexer.T_ENDDECLARE, lexer.T_AS, lexer.T_TRY, lexer.T_CATCH, lexer.T_FINALLY,
		lexer.T_THROW, lexer.T_USE, lexer.T_INSTEADOF, lexer.T_GLOBAL, lexer.T_VAR, lexer.T_UNSET,
		lexer.T_ISSET, lexer.T_EMPTY, lexer.T_CONTINUE, lexer.T_GOTO,
		lexer.T_FUNCTION, lexer.T_CONST, lexer.T_RETURN, lexer.T_PRINT, lexer.T_YIELD, lexer.T_LIST,
		lexer.T_SWITCH, lexer.T_ENDSWITCH, lexer.T_CASE, lexer.T_DEFAULT, lexer.T_BREAK,
		lexer.T_ARRAY, lexer.T_CALLABLE, lexer.T_EXTENDS, lexer.T_IMPLEMENTS, lexer.T_NAMESPACE,
		lexer.T_TRAIT, lexer.T_INTERFACE, lexer.T_CLASS,
		lexer.T_CLASS_C, lexer.T_TRAIT_C, lexer.T_FUNC_C, lexer.T_METHOD_C, lexer.T_LINE,
		lexer.T_FILE, lexer.T_DIR, lexer.T_NS_C, lexer.T_FN, lexer.T_MATCH, lexer.T_ENUM:
		return true
	// Modifiers (also part of semi_reserved)
	case lexer.T_STATIC, lexer.T_ABSTRACT, lexer.T_FINAL, lexer.T_PRIVATE, lexer.T_PROTECTED, lexer.T_PUBLIC, lexer.T_READONLY:
		return true
	default:
		return false
	}
}

// isNamedArgumentStart 检查当前是否为命名参数的开始 (identifier: value)
func (p *Parser) isNamedArgumentStart() bool {
	return isValidIdentifier(p.currentToken.Type) && p.peekToken.Type == lexer.TOKEN_COLON
}

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

	// 初始化时也需要跳过非语法token
	p.currentToken = l.NextToken()
	for isNonSyntacticToken(p.currentToken.Type) {
		p.currentToken = l.NextToken()
	}

	p.peekToken = l.NextToken()
	for isNonSyntacticToken(p.peekToken.Type) {
		p.peekToken = l.NextToken()
	}

	return p
}

// nextToken 前进到下一个语法有意义的token，自动跳过注释等无意义token
func (p *Parser) nextToken() {
	p.currentToken = p.peekToken
	p.peekToken = p.lexer.NextToken()

	// 自动跳过语法解析中无意义的token
	for isNonSyntacticToken(p.peekToken.Type) {
		p.peekToken = p.lexer.NextToken()
	}
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

// expectPeekAny expects any one of the provided token types
func (p *Parser) expectPeekAny(tokens ...lexer.TokenType) bool {
	for _, t := range tokens {
		if p.peekTokenIs(t) {
			p.nextToken()
			return true
		}
	}
	// Create error message with all expected tokens
	var expected []string
	for _, t := range tokens {
		expected = append(expected, lexer.TokenNames[t])
	}
	p.errors = append(p.errors, fmt.Sprintf("expected one of %v, got `%s` at %s", expected, lexer.TokenNames[p.peekToken.Type], p.peekToken.Position))
	return false
}

// expectSemicolon 期望分号，但在关闭标签之前分号是可选的
func (p *Parser) expectSemicolon() bool {
	// PHP规范：语句末尾的分号在?>之前是可选的
	if p.peekToken.Type == lexer.T_CLOSE_TAG {
		// 不前进到关闭标签，让主解析循环处理它
		return true
	}
	return p.expectPeek(lexer.TOKEN_SEMICOLON)
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
	msg := fmt.Sprintf("expected next token to be `%s`, got `%s` instead at position %s",
		lexer.TokenNames[t], lexer.TokenNames[p.peekToken.Type], p.peekToken.Position)
	p.errors = append(p.errors, msg)
}

// Errors 获取解析错误
func (p *Parser) Errors() []string {
	return p.errors
}

// isTypeToken 检查token是否为有效的类型token
func isTypeToken(tokenType lexer.TokenType) bool {
	switch tokenType {
	case lexer.T_STRING, lexer.T_ARRAY, lexer.T_CALLABLE, lexer.T_STATIC,
		lexer.T_NAME_FULLY_QUALIFIED, lexer.T_NAME_RELATIVE, lexer.T_NAME_QUALIFIED,
		lexer.T_NS_SEPARATOR:
		return true
	default:
		return false
	}
}

// isClassNameToken 检查token是否为有效的类名token（根据 class_name 语法规则）
func isClassNameToken(tokenType lexer.TokenType) bool {
	switch tokenType {
	case lexer.T_STATIC, lexer.T_STRING,
		lexer.T_NAME_FULLY_QUALIFIED, lexer.T_NAME_RELATIVE, lexer.T_NAME_QUALIFIED,
		lexer.T_NS_SEPARATOR:
		return true
	default:
		return false
	}
}

// parseTypeHint 解析类型提示，支持nullable, union和intersection类型
func parseTypeHint(p *Parser) *ast.TypeHint {
	// 检查nullable类型 ?Type 或 ?(Type1|Type2)
	nullable := false
	if p.currentToken.Type == lexer.TOKEN_QUESTION {
		nullable = true
		p.nextToken() // 移动到类型token
	}

	// 检查是否为带括号的复合类型 (Type1&Type2)|Type3
	if p.currentToken.Type == lexer.TOKEN_LPAREN {
		return parseParenthesizedTypeHint(p, nullable)
	}

	// 解析基本类型
	if !isTypeToken(p.currentToken.Type) {
		p.errors = append(p.errors, fmt.Sprintf("expected type name, got `%s` instead at position %s", p.currentToken.Value, p.currentToken.Position))
		return nil
	}

	// 创建基本类型提示
	baseType := ast.NewSimpleTypeHint(p.currentToken.Position, p.currentToken.Value, nullable)

	// 检查是否为union类型 Type1|Type2
	if p.peekToken.Type == lexer.TOKEN_PIPE {
		return parseUnionType(p, baseType)
	}

	// 检查是否为intersection类型 Type1&Type2
	if p.peekToken.Type == lexer.TOKEN_AMPERSAND {
		return parseIntersectionType(p, baseType)
	}

	return baseType
}

// parseParameterTypeHint 解析参数类型提示，支持nullable、union和intersection类型
// 在 PHP 8.1+ 中，intersection类型在参数中是支持的
func parseParameterTypeHint(p *Parser) *ast.TypeHint {
	// 检查nullable类型 ?Type 或 ?(Type1|Type2)
	nullable := false
	if p.currentToken.Type == lexer.TOKEN_QUESTION {
		nullable = true
		p.nextToken() // 移动到类型token
	}

	// 检查是否为带括号的复合类型 (Type1&Type2)|Type3
	if p.currentToken.Type == lexer.TOKEN_LPAREN {
		return parseParenthesizedTypeHint(p, nullable)
	}

	// 解析基本类型
	if !isTypeToken(p.currentToken.Type) {
		p.errors = append(p.errors, fmt.Sprintf("expected type name, got `%s` instead at position %s", p.currentToken.Value, p.currentToken.Position))
		return nil
	}

	// 解析完整的类型名（可能包含命名空间）
	var typeName string
	if p.currentToken.Type == lexer.T_NS_SEPARATOR || p.currentToken.Type == lexer.T_STRING {
		// 使用 parseClassName 来处理完整的命名空间名称
		className, err := parseClassName(p)
		if err != nil {
			p.errors = append(p.errors, err.Error())
			return nil
		}
		typeName = className
	} else {
		// 简单类型名
		typeName = p.currentToken.Value
	}

	// 创建基本类型提示
	baseType := ast.NewSimpleTypeHint(p.currentToken.Position, typeName, nullable)

	// 检查是否为union类型 Type1|Type2
	if p.peekToken.Type == lexer.TOKEN_PIPE {
		return parseUnionType(p, baseType)
	}

	// 暂时禁用intersection类型支持以修复引用参数解析
	// intersection类型会在后续版本中重新实现
	// if p.peekToken.Type == lexer.TOKEN_AMPERSAND {
	//     return parseIntersectionType(p, baseType)
	// }

	return baseType
}

// parseUnionType 解析联合类型 Type1|Type2|Type3
func parseUnionType(p *Parser, firstType *ast.TypeHint) *ast.TypeHint {
	pos := firstType.Position
	types := []*ast.TypeHint{firstType}

	for p.peekToken.Type == lexer.TOKEN_PIPE {
		p.nextToken() // 移动到 |
		p.nextToken() // 移动到类型token

		var typeHint *ast.TypeHint

		// 检查nullable类型
		nullable := false
		if p.currentToken.Type == lexer.TOKEN_QUESTION {
			nullable = true
			p.nextToken() // 移动到实际类型token
		}

		// 检查是否为括号表达式
		if p.currentToken.Type == lexer.TOKEN_LPAREN {
			typeHint = parseParenthesizedTypeHint(p, nullable)
		} else if isTypeToken(p.currentToken.Type) {
			typeHint = ast.NewSimpleTypeHint(p.currentToken.Position, p.currentToken.Value, nullable)
		} else {
			p.errors = append(p.errors, fmt.Sprintf("expected type name in union type, got `%s` instead", p.currentToken.Value))
			return nil
		}

		if typeHint != nil {
			types = append(types, typeHint)
		}
	}

	return ast.NewUnionTypeHint(pos, types)
}

// parseIntersectionType 解析交集类型 Type1&Type2&Type3
// 如果遇到非类型token，则返回nil表示这不是intersection类型
func parseIntersectionType(p *Parser, firstType *ast.TypeHint) *ast.TypeHint {
	pos := firstType.Position
	types := []*ast.TypeHint{firstType}

	for p.peekToken.Type == lexer.TOKEN_AMPERSAND {
		// 检查 & 后面是否跟着类型token
		savedCurrentToken := p.currentToken
		savedPeekToken := p.peekToken
		
		p.nextToken() // 移动到 &
		
		// 检查下一个token是否是类型
		if !isTypeToken(p.peekToken.Type) {
			// 不是intersection类型，恢复位置并返回nil
			p.currentToken = savedCurrentToken
			p.peekToken = savedPeekToken
			return nil
		}
		
		p.nextToken() // 移动到类型token
		
		typeHint := ast.NewSimpleTypeHint(p.currentToken.Position, p.currentToken.Value, false) // intersection types can't be nullable
		types = append(types, typeHint)
	}

	return ast.NewIntersectionTypeHint(pos, types)
}

// parseParenthesizedTypeHint 解析带括号的复杂类型 (Type1&Type2)|Type3
func parseParenthesizedTypeHint(p *Parser, nullable bool) *ast.TypeHint {
	// 跳过开括号 (
	p.nextToken() // 移动到括号内的第一个类型

	// 检查是否为空括号
	if p.currentToken.Type == lexer.TOKEN_RPAREN {
		p.errors = append(p.errors, "empty parentheses in type expression")
		return nil
	}

	// 递归解析括号内的类型
	innerType := parseComplexTypeHint(p)
	if innerType == nil {
		return nil
	}

	// 确保有结束括号
	if !p.expectPeek(lexer.TOKEN_RPAREN) {
		p.errors = append(p.errors, "expected closing parenthesis in type expression")
		return nil
	}

	// 应用nullable到括号内的类型
	if nullable {
		innerType.Nullable = true
	}

	// 检查括号后是否有union/intersection操作符
	if p.peekToken.Type == lexer.TOKEN_PIPE {
		return parseUnionType(p, innerType)
	} else if p.peekToken.Type == lexer.TOKEN_AMPERSAND {
		return parseIntersectionType(p, innerType)
	}

	return innerType
}

// parseComplexTypeHint 解析复杂类型（不处理最外层的nullable）
func parseComplexTypeHint(p *Parser) *ast.TypeHint {
	// 解析基本类型
	if !isTypeToken(p.currentToken.Type) {
		p.errors = append(p.errors, fmt.Sprintf("expected type name in complex type, got `%s` instead", p.currentToken.Value))
		return nil
	}

	// 创建基本类型提示
	baseType := ast.NewSimpleTypeHint(p.currentToken.Position, p.currentToken.Value, false)

	// 检查是否为union类型 Type1|Type2
	if p.peekToken.Type == lexer.TOKEN_PIPE {
		return parseUnionType(p, baseType)
	}

	// 检查是否为intersection类型 Type1&Type2
	if p.peekToken.Type == lexer.TOKEN_AMPERSAND {
		return parseIntersectionType(p, baseType)
	}

	return baseType
}

// ParseProgram 解析整个程序
func (p *Parser) ParseProgram() *ast.Program {
	program := ast.NewProgram(p.currentToken.Position)

	// 跳过 PHP 开始标签
	if p.currentToken.Type == lexer.T_OPEN_TAG {
		p.nextToken()
	}

	for !p.isAtEnd() {
		// Handle T_CLOSE_TAG: just skip it and continue parsing
		if p.currentToken.Type == lexer.T_CLOSE_TAG {
			p.nextToken()
			continue
		}

		// Handle T_INLINE_HTML as a statement (like echo)
		if p.currentToken.Type == lexer.T_INLINE_HTML {
			// Create a statement from the inline HTML content
			htmlExpr := parseInlineHTML(p)
			if htmlExpr != nil {
				// Wrap inline HTML as an expression statement
				program.Body = append(program.Body, ast.NewExpressionStatement(htmlExpr.(*ast.StringLiteral).Position, htmlExpr))
			}
			p.nextToken()
			continue
		}

		// Skip T_OPEN_TAG tokens in the middle of the program
		if p.currentToken.Type == lexer.T_OPEN_TAG {
			p.nextToken()
			continue
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
	p.errors = append(p.errors, fmt.Sprintf("expected `%s`, got `%s` at %s", tokenType, p.peekToken.Type, p.peekToken.Position))
	return false
}

// noPrefixParseFnError 添加前缀解析函数缺失错误
func (p *Parser) noPrefixParseFnError(t lexer.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for `%s` %s found as at position %s", lexer.TokenNames[t], t, p.currentToken.Position)
	p.errors = append(p.errors, msg)
}

// parseStatement 解析语句
func parseStatement(p *Parser) ast.Statement {
	switch p.currentToken.Type {
	case lexer.T_ECHO:
		return parseEchoStatement(p)
	case lexer.T_OPEN_TAG_WITH_ECHO:
		return parseShortEchoStatement(p)
	case lexer.T_PRINT:
		return parsePrintStatement(p)
	case lexer.T_IF:
		return parseIfStatement(p)
	case lexer.T_WHILE:
		return parseWhileStatement(p)
	case lexer.T_FOR:
		return parseForStatement(p)
	case lexer.T_FUNCTION:
		return parseFunctionDeclaration(p)
	case lexer.T_CLASS:
		// Check if this is class::method (static call) vs class declaration
		if p.peekToken.Type == lexer.T_PAAMAYIM_NEKUDOTAYIM {
			// This is a static method call like "Class::method()", treat as expression statement
			return parseExpressionStatement(p)
		}
		return parseClassDeclaration(p)
	case lexer.T_NAMESPACE:
		return parseNamespaceStatement(p)
	case lexer.T_USE:
		return parseUseStatement(p)
	case lexer.T_INTERFACE:
		return parseInterfaceDeclaration(p)
	case lexer.T_TRAIT:
		return parseTraitDeclaration(p)
	case lexer.T_ENUM:
		return parseEnumDeclaration(p)
	case lexer.T_READONLY:
		// Check if this is "readonly class"
		if p.peekToken.Type == lexer.T_CLASS {
			return parseReadonlyClassDeclaration(p)
		}
		// Otherwise fall through to expression statement
		return parseExpressionStatement(p)
	case lexer.T_RETURN:
		return parseReturnStatement(p)
	case lexer.T_GLOBAL:
		return parseGlobalStatement(p)
	case lexer.T_STATIC:
		// Check if this is late static binding (static::method or static::$property)
		if p.peekToken.Type == lexer.T_PAAMAYIM_NEKUDOTAYIM {
			// This is static::something - treat as expression statement
			return parseExpressionStatement(p)
		}
		// Otherwise it's a static variable declaration
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
	case lexer.T_DECLARE:
		return parseDeclareStatement(p)
	case lexer.T_THROW:
		return parseThrowStatement(p)
	case lexer.T_GOTO:
		return parseGotoStatement(p)
	case lexer.T_BREAK:
		return parseBreakStatement(p)
	case lexer.T_CONTINUE:
		return parseContinueStatement(p)
	case lexer.T_ENDIF:
		return parseEndifStatement(p)
	case lexer.T_HALT_COMPILER:
		return parseHaltCompilerStatement(p)
	case lexer.TOKEN_LBRACE:
		return parseBlockStatement(p)
	case lexer.T_FINAL:
		// Check if this is "final class"
		if p.peekToken.Type == lexer.T_CLASS {
			return parseFinalClassDeclaration(p)
		}
		// Otherwise fall through to expression statement
		return parseExpressionStatement(p)
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

	if !p.expectSemicolon() {
		return nil
	}

	return stmt
}

// parseShortEchoStatement 解析短回声语句 (<?= ... ?>)
func parseShortEchoStatement(p *Parser) *ast.EchoStatement {
	stmt := ast.NewEchoStatement(p.currentToken.Position)

	// 解析表达式 (<?= 之后的内容)
	p.nextToken()
	stmt.Arguments = append(stmt.Arguments, parseExpression(p, LOWEST))

	// 处理可选的分号（在短回声标签中可以有分号）
	if p.peekToken.Type == lexer.TOKEN_SEMICOLON {
		p.nextToken() // 移动到分号位置
	}

	return stmt
}

// parsePrintStatement 解析 print 语句
func parsePrintStatement(p *Parser) *ast.PrintStatement {
	stmt := ast.NewPrintStatement(p.currentToken.Position)

	// 解析 print 的参数
	p.nextToken()
	stmt.Arguments = append(stmt.Arguments, parseExpression(p, LOWEST))

	// 处理多个参数（用逗号分隔）
	for p.peekToken.Type == lexer.TOKEN_COMMA {
		p.nextToken() // 移动到逗号
		p.nextToken() // 移动到下一个表达式
		stmt.Arguments = append(stmt.Arguments, parseExpression(p, LOWEST))
	}

	if !p.expectSemicolon() {
		return nil
	}

	return stmt
}

// parseIfStatement 解析 if 语句，支持普通语法和Alternative语法
func parseIfStatement(p *Parser) ast.Statement {
	pos := p.currentToken.Position

	if !p.expectPeek(lexer.TOKEN_LPAREN) {
		return nil
	}

	p.nextToken()
	condition := parseExpression(p, LOWEST)

	if !p.expectPeek(lexer.TOKEN_RPAREN) {
		return nil
	}

	// 检查是否为Alternative语法 (if (...):)
	if p.peekToken.Type == lexer.TOKEN_COLON {
		p.nextToken() // 移动到 :
		return parseAlternativeIfStatement(p, pos, condition)
	}

	// 普通语法 (if (...) { ... } 或 if (...) statement;)
	p.nextToken() // 移动到下一个token

	ifStmt := ast.NewIfStatement(pos, condition)

	// 检查是否是块语句 ({})
	if p.currentToken.Type == lexer.TOKEN_LBRACE {
		ifStmt.Consequent = parseBlockStatements(p)
	} else {
		// 单行语句
		singleStmt := parseStatement(p)
		if singleStmt != nil {
			ifStmt.Consequent = []ast.Statement{singleStmt}
		}
	}

	// 检查是否有 elseif 或 else 子句
	for p.peekToken.Type == lexer.T_ELSEIF {
		p.nextToken() // 移动到 elseif

		if !p.expectPeek(lexer.TOKEN_LPAREN) {
			return ifStmt
		}

		p.nextToken()
		elseifCondition := parseExpression(p, LOWEST)

		if !p.expectPeek(lexer.TOKEN_RPAREN) {
			return ifStmt
		}

		// elseif 语句体
		p.nextToken()
		elseifStmt := ast.NewIfStatement(p.currentToken.Position, elseifCondition)

		if p.currentToken.Type == lexer.TOKEN_LBRACE {
			elseifStmt.Consequent = parseBlockStatements(p)
		} else {
			// 单行语句
			singleStmt := parseStatement(p)
			if singleStmt != nil {
				elseifStmt.Consequent = []ast.Statement{singleStmt}
			}
		}

		ifStmt.Alternate = append(ifStmt.Alternate, elseifStmt)
	}

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
		} else {
			// else block 或单行语句
			p.nextToken()
			if p.currentToken.Type == lexer.TOKEN_LBRACE {
				// 块语句
				ifStmt.Alternate = parseBlockStatements(p)
			} else {
				// 单行语句
				elseStmt := parseStatement(p)
				if elseStmt != nil {
					ifStmt.Alternate = []ast.Statement{elseStmt}
				}
			}
		}
	}

	return ifStmt
}

// parseAlternativeIfStatement 解析Alternative语法的if语句 (if: ... elseif: ... else: ... endif;)
func parseAlternativeIfStatement(p *Parser, pos lexer.Position, condition ast.Expression) *ast.AlternativeIfStatement {
	altIfStmt := ast.NewAlternativeIfStatement(pos, condition)

	// 解析第一个if块的语句列表
	for p.peekToken.Type != lexer.T_ELSEIF && p.peekToken.Type != lexer.T_ELSE && p.peekToken.Type != lexer.T_ENDIF {
		if p.peekToken.Type == lexer.T_EOF {
			p.errors = append(p.errors, "Expected endif, but reached end of file")
			return nil
		}
		p.nextToken()
		if stmt := parseStatement(p); stmt != nil {
			altIfStmt.Then = append(altIfStmt.Then, stmt)
		}
	}

	// 处理elseif子句
	for p.peekToken.Type == lexer.T_ELSEIF {
		p.nextToken() // 移动到 elseif

		if !p.expectPeek(lexer.TOKEN_LPAREN) {
			return nil
		}

		p.nextToken()
		elseifCondition := parseExpression(p, LOWEST)

		if !p.expectPeek(lexer.TOKEN_RPAREN) {
			return nil
		}

		if !p.expectPeek(lexer.TOKEN_COLON) {
			return nil
		}

		elseifClause := ast.NewElseIfClause(p.currentToken.Position, elseifCondition)

		// 解析elseif块的语句列表
		for p.peekToken.Type != lexer.T_ELSEIF && p.peekToken.Type != lexer.T_ELSE && p.peekToken.Type != lexer.T_ENDIF {
			if p.peekToken.Type == lexer.T_EOF {
				p.errors = append(p.errors, "Expected endif, but reached end of file")
				return nil
			}
			p.nextToken()
			if stmt := parseStatement(p); stmt != nil {
				elseifClause.Body = append(elseifClause.Body, stmt)
			}
		}

		altIfStmt.ElseIfs = append(altIfStmt.ElseIfs, elseifClause)
	}

	// 处理else子句
	if p.peekToken.Type == lexer.T_ELSE {
		p.nextToken() // 移动到 else

		if !p.expectPeek(lexer.TOKEN_COLON) {
			return nil
		}

		// 解析else块的语句列表
		for p.peekToken.Type != lexer.T_ENDIF {
			if p.peekToken.Type == lexer.T_EOF {
				p.errors = append(p.errors, "Expected endif, but reached end of file")
				return nil
			}
			p.nextToken()
			if stmt := parseStatement(p); stmt != nil {
				altIfStmt.Else = append(altIfStmt.Else, stmt)
			}
		}
	}

	// 期望 endif
	if !p.expectPeek(lexer.T_ENDIF) {
		return nil
	}

	// 期望分号
	if !p.expectSemicolon() {
		return nil
	}

	return altIfStmt
}

// parseAlternativeSwitchStatement 解析Alternative语法的switch语句 (switch: ... endswitch;)
func parseAlternativeSwitchStatement(p *Parser, pos lexer.Position, discriminant ast.Expression) ast.Statement {
	switchStmt := ast.NewSwitchStatement(pos, discriminant)

	p.nextToken() // 跳过冒号，移到第一个语句/case

	// 解析switch体中的语句，直到遇到endswitch
	for p.currentToken.Type != lexer.T_ENDSWITCH && p.currentToken.Type != lexer.T_EOF {
		if p.currentToken.Type == lexer.T_CASE {
			casePos := p.currentToken.Position
			p.nextToken()
			test := parseExpression(p, LOWEST)

			if !p.expectPeek(lexer.TOKEN_COLON) {
				continue
			}

			switchCase := ast.NewSwitchCase(casePos, test)
			p.nextToken()

			// 解析case体中的语句
			for p.currentToken.Type != lexer.T_CASE &&
				p.currentToken.Type != lexer.T_DEFAULT &&
				p.currentToken.Type != lexer.T_ENDSWITCH &&
				p.currentToken.Type != lexer.T_EOF {
				if stmt := parseStatement(p); stmt != nil {
					switchCase.Body = append(switchCase.Body, stmt)
				}
				p.nextToken()
			}
			switchStmt.Cases = append(switchStmt.Cases, switchCase)
		} else if p.currentToken.Type == lexer.T_DEFAULT {
			defaultPos := p.currentToken.Position

			if !p.expectPeek(lexer.TOKEN_COLON) {
				continue
			}

			// default case的test为nil
			defaultCase := ast.NewSwitchCase(defaultPos, nil)
			p.nextToken()

			// 解析default体中的语句
			for p.currentToken.Type != lexer.T_CASE &&
				p.currentToken.Type != lexer.T_DEFAULT &&
				p.currentToken.Type != lexer.T_ENDSWITCH &&
				p.currentToken.Type != lexer.T_EOF {
				if stmt := parseStatement(p); stmt != nil {
					defaultCase.Body = append(defaultCase.Body, stmt)
				}
				p.nextToken()
			}
			switchStmt.Cases = append(switchStmt.Cases, defaultCase)
		} else {
			// 如果遇到其他语句，跳过
			p.nextToken()
		}
	}

	// 期望 endswitch
	if p.currentToken.Type != lexer.T_ENDSWITCH {
		p.errors = append(p.errors, fmt.Sprintf("expected 'endswitch', got '%s'", p.currentToken.Type))
		return nil
	}

	// 期望分号
	if !p.expectSemicolon() {
		return nil
	}

	return switchStmt
}

// parseWhileStatement 解析 while 语句，支持普通语法和Alternative语法
func parseWhileStatement(p *Parser) ast.Statement {
	pos := p.currentToken.Position

	if !p.expectPeek(lexer.TOKEN_LPAREN) {
		return nil
	}

	p.nextToken()
	condition := parseExpression(p, LOWEST)

	if !p.expectPeek(lexer.TOKEN_RPAREN) {
		return nil
	}

	// 检查是否为Alternative语法 (while (...):)
	if p.peekToken.Type == lexer.TOKEN_COLON {
		p.nextToken() // 移动到 :
		return parseAlternativeWhileStatement(p, pos, condition)
	}

	// 普通语法 (while (...) { ... } 或 while (...) statement;)
	p.nextToken() // 移动到下一个token

	whileStmt := ast.NewWhileStatement(pos, condition)

	// 检查是否是块语句 ({})
	if p.currentToken.Type == lexer.TOKEN_LBRACE {
		whileStmt.Body = parseBlockStatements(p)
	} else {
		// 单行语句
		singleStmt := parseStatement(p)
		if singleStmt != nil {
			whileStmt.Body = []ast.Statement{singleStmt}
		}
	}

	return whileStmt
}

// parseAlternativeWhileStatement 解析Alternative语法的while语句 (while: ... endwhile;)
func parseAlternativeWhileStatement(p *Parser, pos lexer.Position, condition ast.Expression) *ast.AlternativeWhileStatement {
	altWhileStmt := ast.NewAlternativeWhileStatement(pos, condition)

	// 解析while块的语句列表
	for p.peekToken.Type != lexer.T_ENDWHILE {
		if p.peekToken.Type == lexer.T_EOF {
			p.errors = append(p.errors, "Expected endwhile, but reached end of file")
			return nil
		}
		p.nextToken()
		if stmt := parseStatement(p); stmt != nil {
			altWhileStmt.Body = append(altWhileStmt.Body, stmt)
		}
	}

	// 期望 endwhile
	if !p.expectPeek(lexer.T_ENDWHILE) {
		return nil
	}

	// 期望分号
	if !p.expectSemicolon() {
		return nil
	}

	return altWhileStmt
}

// parseCommaSeparatedExpressions parses comma-separated expression list for for loops
func parseCommaSeparatedExpressions(p *Parser, endTokenTypes ...lexer.TokenType) []ast.Expression {
	var expressions []ast.Expression

	// Check if we should stop (empty expression list)
	for _, tokenType := range endTokenTypes {
		if p.currentToken.Type == tokenType {
			return expressions
		}
	}

	// Parse first expression
	expressions = append(expressions, parseExpression(p, LOWEST))

	// Parse remaining comma-separated expressions
	for p.peekToken.Type == lexer.TOKEN_COMMA {
		p.nextToken() // move to comma
		p.nextToken() // move to next expression
		expressions = append(expressions, parseExpression(p, LOWEST))
	}

	return expressions
}

// createSingleExpression converts a slice of expressions to a single expression
// If there's only one expression, returns it directly
// If there are multiple expressions, returns a CommaExpression
func createSingleExpression(expressions []ast.Expression) ast.Expression {
	if len(expressions) == 0 {
		return nil
	}
	if len(expressions) == 1 {
		return expressions[0]
	}
	// Multiple expressions - create a comma expression
	return ast.NewCommaExpression(expressions[0].GetPosition(), expressions)
}

// parseExpressionList 解析表达式列表（用于函数调用参数等），支持命名参数
func parseExpressionList(p *Parser, end lexer.TokenType) []ast.Expression {
	var args []ast.Expression

	// 检查当前是否已经在end token
	if p.currentToken.Type == end {
		return args
	}

	// 如果没有在end token，但peek是end，则前进
	if p.peekToken.Type == end {
		p.nextToken()
		return args
	}

	// 如果当前不是T_ELLIPSIS，则需要前进到第一个参数
	if p.currentToken.Type != lexer.T_ELLIPSIS {
		p.nextToken()
	}

	// Check for different argument types
	if p.isNamedArgumentStart() {
		// Named argument (identifier: value)
		arg := parseNamedArgument(p)
		if arg != nil {
			args = append(args, arg)
		}
	} else if p.currentToken.Type == lexer.T_ELLIPSIS {
		// Spread argument (...expr)
		pos := p.currentToken.Position
		p.nextToken() // 跳过 ...
		expr := parseExpression(p, LOWEST)
		if expr != nil {
			spreadExpr := ast.NewSpreadExpression(pos, expr)
			args = append(args, spreadExpr)
		}
	} else {
		args = append(args, parseExpression(p, LOWEST))
	}

	for p.peekToken.Type == lexer.TOKEN_COMMA {
		p.nextToken() // 移动到逗号
		
		// 检查逗号后面是否直接是结束符（支持尾随逗号）
		if p.peekToken.Type == end {
			break
		}
		
		p.nextToken() // 移动到逗号后的token

		// Check for different argument types
		if p.isNamedArgumentStart() {
			// Named argument (identifier: value)
			arg := parseNamedArgument(p)
			if arg != nil {
				args = append(args, arg)
			}
		} else if p.currentToken.Type == lexer.T_ELLIPSIS {
			// Spread argument (...expr)
			pos := p.currentToken.Position
			p.nextToken() // 跳过 ...
			expr := parseExpression(p, LOWEST)
			if expr != nil {
				spreadExpr := ast.NewSpreadExpression(pos, expr)
				args = append(args, spreadExpr)
			}
		} else {
			args = append(args, parseExpression(p, LOWEST))
		}
	}

	if !p.expectPeek(end) {
		return nil
	}

	return args
}

// parseForStatement 解析 for 语句，支持普通语法和Alternative语法
func parseForStatement(p *Parser) ast.Statement {
	pos := p.currentToken.Position

	if !p.expectPeek(lexer.TOKEN_LPAREN) {
		return nil
	}

	var initExprs, conditionExprs, updateExprs []ast.Expression

	// 解析初始化表达式 (可以有多个，用逗号分隔)
	if p.peekToken.Type != lexer.TOKEN_SEMICOLON {
		p.nextToken()
		initExprs = parseCommaSeparatedExpressions(p, lexer.TOKEN_SEMICOLON)
	}

	if !p.expectSemicolon() {
		return nil
	}

	// 解析条件表达式 (可以有多个，用逗号分隔)
	if p.peekToken.Type != lexer.TOKEN_SEMICOLON {
		p.nextToken()
		conditionExprs = parseCommaSeparatedExpressions(p, lexer.TOKEN_SEMICOLON)
	}

	if !p.expectSemicolon() {
		return nil
	}

	// 解析更新表达式 (可以有多个，用逗号分隔)
	if p.peekToken.Type != lexer.TOKEN_RPAREN {
		p.nextToken()
		updateExprs = parseCommaSeparatedExpressions(p, lexer.TOKEN_RPAREN)
	}

	if !p.expectPeek(lexer.TOKEN_RPAREN) {
		return nil
	}

	// 检查是否为Alternative语法 (for (...):)
	if p.peekToken.Type == lexer.TOKEN_COLON {
		p.nextToken() // 移动到 :
		return parseAlternativeForStatement(p, pos, initExprs, conditionExprs, updateExprs)
	}

	// 普通语法 (for (...) { ... } 或 for (...) statement;)
	p.nextToken() // 移动到下一个token

	forStmt := ast.NewForStatement(pos)
	forStmt.Init = createSingleExpression(initExprs)
	forStmt.Test = createSingleExpression(conditionExprs)
	forStmt.Update = createSingleExpression(updateExprs)

	// 检查是否是块语句 ({})
	if p.currentToken.Type == lexer.TOKEN_LBRACE {
		forStmt.Body = parseBlockStatements(p)
	} else {
		// 单行语句
		singleStmt := parseStatement(p)
		if singleStmt != nil {
			forStmt.Body = []ast.Statement{singleStmt}
		}
	}

	return forStmt
}

// parseAlternativeForStatement 解析Alternative语法的for语句 (for: ... endfor;)
func parseAlternativeForStatement(p *Parser, pos lexer.Position, initExprs, conditionExprs, updateExprs []ast.Expression) *ast.AlternativeForStatement {
	altForStmt := ast.NewAlternativeForStatement(pos)
	altForStmt.Init = initExprs
	altForStmt.Condition = conditionExprs
	altForStmt.Update = updateExprs

	// 解析for块的语句列表
	for p.peekToken.Type != lexer.T_ENDFOR {
		if p.peekToken.Type == lexer.T_EOF {
			p.errors = append(p.errors, "Expected endfor, but reached end of file")
			return nil
		}
		p.nextToken()
		if stmt := parseStatement(p); stmt != nil {
			altForStmt.Body = append(altForStmt.Body, stmt)
		}
	}

	// 期望 endfor
	if !p.expectPeek(lexer.T_ENDFOR) {
		return nil
	}

	// 期望分号
	if !p.expectSemicolon() {
		return nil
	}

	return altForStmt
}

// parseFunctionDeclaration 解析函数声明
func parseFunctionDeclaration(p *Parser) *ast.FunctionDeclaration {
	pos := p.currentToken.Position

	// 检查是否有 abstract 修饰符
	var isAbstract bool
	if p.currentToken.Type == lexer.T_ABSTRACT {
		isAbstract = true
		p.nextToken() // 移动到下一个token (可能是visibility或function)
	}

	// 检查是否有可见性修饰符 (public, private, protected)
	var visibility string
	var isStatic bool
	if p.currentToken.Type == lexer.T_PUBLIC || p.currentToken.Type == lexer.T_PRIVATE || p.currentToken.Type == lexer.T_PROTECTED {
		visibility = p.currentToken.Value
		// 检查是否有 static 修饰符
		if p.peekToken.Type == lexer.T_STATIC {
			isStatic = true
			p.nextToken() // 移动到 static
			if !p.expectPeek(lexer.T_FUNCTION) {
				return nil
			}
		} else if !p.expectPeek(lexer.T_FUNCTION) {
			return nil
		}
	} else if p.currentToken.Type == lexer.T_STATIC {
		// Handle static function without explicit visibility
		isStatic = true
		if !p.expectPeek(lexer.T_FUNCTION) {
			return nil
		}
	}

	// 检查是否为引用返回函数 function &foo()
	byReference := false
	if p.peekToken.Type == lexer.TOKEN_AMPERSAND {
		byReference = true
		p.nextToken() // 移动到 &
	}

	// Expect method/function name (allow reserved keywords)
	p.nextToken()
	if p.currentToken.Type != lexer.T_STRING && !isSemiReserved(p.currentToken.Type) {
		p.errors = append(p.errors, fmt.Sprintf("expected function name, got %s at position %s",
			p.currentToken.Value, p.currentToken.Position))
		return nil
	}

	name := ast.NewIdentifierNode(p.currentToken.Position, p.currentToken.Value)
	funcDecl := ast.NewFunctionDeclaration(pos, name)
	funcDecl.ByReference = byReference
	funcDecl.Visibility = visibility
	funcDecl.IsStatic = isStatic

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
			
			// 检查逗号后面是否直接是结束符（支持尾随逗号）
			if p.peekToken.Type == lexer.TOKEN_RPAREN {
				break
			}
			
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

	// 检查是否有返回类型声明 ": type"
	if p.peekToken.Type == lexer.TOKEN_COLON {
		p.nextToken() // 移动到 ':'
		p.nextToken() // 移动到类型开始位置

		// 解析返回类型（支持复杂类型）
		returnType := parseTypeHint(p)
		if returnType == nil {
			return nil
		}
		funcDecl.ReturnType = returnType
	}

	// For abstract methods, expect semicolon instead of function body
	if isAbstract {
		if !p.expectPeek(lexer.TOKEN_SEMICOLON) {
			return nil
		}
		// Abstract methods have no body
		funcDecl.Body = nil
	} else {
		if !p.expectPeek(lexer.TOKEN_LBRACE) {
			return nil
		}
		funcDecl.Body = parseBlockStatements(p)
	}

	return funcDecl
}

// parseParameter 解析函数参数（支持类型提示、引用、可变参数、可见性修饰符等）
func parseParameter(p *Parser) *ast.Parameter {
	var param ast.Parameter

	// 首先解析属性 (attributes) #[...]
	// 根据PHP语法：attributed_parameter: attributes parameter | parameter
	var attributes []*ast.AttributeGroup
	for p.currentToken.Type == lexer.T_ATTRIBUTE {
		attrGroup := parseAttributeGroup(p)
		if attrGroup != nil {
			attributes = append(attributes, attrGroup)
		}
		p.nextToken() // 移动到下一个token
	}
	param.Attributes = attributes

	// 检查可见性修饰符 (public, private, protected, readonly)
	// 这些通常只在构造函数参数中使用
	if p.currentToken.Type == lexer.T_PUBLIC || p.currentToken.Type == lexer.T_PRIVATE || p.currentToken.Type == lexer.T_PROTECTED {
		param.Visibility = p.currentToken.Value
		p.nextToken()
	}

	if p.currentToken.Type == lexer.T_READONLY {
		param.ReadOnly = true
		p.nextToken()
	}

	// 检查是否有类型提示
	if isTypeToken(p.currentToken.Type) || p.currentToken.Type == lexer.TOKEN_QUESTION {
		typeHint := parseParameterTypeHint(p)
		if typeHint == nil {
			return nil
		}
		param.Type = typeHint
		p.nextToken() // 移动到下一个token
	}

	// 检查引用参数 &$param
	if p.currentToken.Type == lexer.TOKEN_AMPERSAND {
		param.ByReference = true
		p.nextToken()
	}

	// 检查可变参数 ...$params
	if p.currentToken.Type == lexer.T_ELLIPSIS {
		param.Variadic = true
		p.nextToken()
	}

	// 解析参数名
	if p.currentToken.Type == lexer.T_VARIABLE {
		param.Name = p.currentToken.Value
	} else {
		p.errors = append(p.errors, fmt.Sprintf("expected variable name, got `%s` instead at position %s", p.currentToken.Value, p.currentToken.Position))
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

// parseNamespaceStatement 解析命名空间声明语句
func parseNamespaceStatement(p *Parser) *ast.NamespaceStatement {
	pos := p.currentToken.Position

	// namespace 关键字后面可能是：
	// 1. namespace; (全局命名空间)
	// 2. namespace Foo\Bar; (简单命名空间声明)
	// 3. namespace Foo\Bar { ... } (带块的命名空间声明)

	if p.peekToken.Type == lexer.TOKEN_SEMICOLON {
		// 全局命名空间: namespace;
		p.nextToken() // 移动到 ;
		return ast.NewNamespaceStatement(pos, nil)
	}

	if p.peekToken.Type == lexer.TOKEN_LBRACE {
		// 匿名命名空间: namespace { ... }
		p.nextToken() // 移动到 {
		stmt := ast.NewNamespaceStatement(pos, nil)
		stmt.Body = parseBlockStatements(p)
		return stmt
	}

	// 解析命名空间名称 - 支持简单标识符和限定名
	var nameParts []string
	if p.peekToken.Type == lexer.T_STRING {
		p.nextToken() // 移动到标识符
		nameParts = []string{p.currentToken.Value}
		
		// 处理多级命名空间 (Foo\Bar\Baz) - 旧格式兼容
		for p.peekToken.Type == lexer.T_NS_SEPARATOR {
			p.nextToken() // 移动到 \
			if !p.expectPeek(lexer.T_STRING) {
				return nil
			}
			nameParts = append(nameParts, p.currentToken.Value)
		}
	} else if p.peekToken.Type == lexer.T_NAME_QUALIFIED || p.peekToken.Type == lexer.T_NAME_RELATIVE {
		p.nextToken() // 移动到限定名
		// 分割限定名为各个部分
		qualifiedName := p.currentToken.Value
		nameParts = strings.Split(qualifiedName, "\\")
		// 过滤空字符串（可能来自开头的\）
		var filteredParts []string
		for _, part := range nameParts {
			if part != "" {
				filteredParts = append(filteredParts, part)
			}
		}
		nameParts = filteredParts
	} else {
		p.errors = append(p.errors, fmt.Sprintf("expected namespace name, got %s at position %s", p.peekToken.Value, p.peekToken.Position))
		return nil
	}

	namespaceName := ast.NewNamespaceNameExpression(pos, nameParts)
	stmt := ast.NewNamespaceStatement(pos, namespaceName)

	// 检查是否有块语法
	if p.peekToken.Type == lexer.TOKEN_LBRACE {
		p.nextToken() // 移动到 {
		stmt.Body = parseBlockStatements(p)
	} else if p.peekToken.Type == lexer.TOKEN_SEMICOLON {
		p.nextToken() // 移动到 ;
	}

	return stmt
}

// parseUseNamespaceName parses a namespace name that can be a single qualified token or multiple string tokens
func parseUseNamespaceName(p *Parser) (*ast.NamespaceNameExpression, error) {
	pos := p.currentToken.Position
	
	// Handle single qualified tokens (T_NAME_QUALIFIED, T_NAME_FULLY_QUALIFIED, T_NAME_RELATIVE)
	if p.currentToken.Type == lexer.T_NAME_QUALIFIED || 
		p.currentToken.Type == lexer.T_NAME_FULLY_QUALIFIED || 
		p.currentToken.Type == lexer.T_NAME_RELATIVE {
		// Split the qualified name into parts
		qualifiedName := p.currentToken.Value
		var nameParts []string
		if strings.HasPrefix(qualifiedName, "\\") {
			// Remove leading backslash and split
			qualifiedName = strings.TrimPrefix(qualifiedName, "\\")
		}
		nameParts = strings.Split(qualifiedName, "\\")
		return ast.NewNamespaceNameExpression(pos, nameParts), nil
	}

	// Handle traditional T_STRING tokens separated by T_NS_SEPARATOR
	if p.currentToken.Type == lexer.T_STRING {
		nameParts := []string{p.currentToken.Value}

		// 处理多级命名空间 (Foo\Bar\Baz)
		for p.peekToken.Type == lexer.T_NS_SEPARATOR {
			p.nextToken() // 移动到 \
			if !p.expectPeek(lexer.T_STRING) {
				return nil, fmt.Errorf("expected string after namespace separator")
			}
			nameParts = append(nameParts, p.currentToken.Value)
		}

		return ast.NewNamespaceNameExpression(pos, nameParts), nil
	}

	return nil, fmt.Errorf("expected namespace name, got %s", p.currentToken.Type)
}

// parseUseStatement 解析 use 语句
func parseUseStatement(p *Parser) *ast.UseStatement {
	pos := p.currentToken.Position
	stmt := ast.NewUseStatement(pos)

	// 检查是否有类型修饰符 (function, const)
	var useType string
	if p.peekToken.Type == lexer.T_FUNCTION || p.peekToken.Type == lexer.T_CONST {
		p.nextToken()
		useType = p.currentToken.Value
	}

	// 解析第一个 use 子句 - 接受多种名称类型
	if !p.expectPeekAny(lexer.T_STRING, lexer.T_NAME_QUALIFIED, lexer.T_NAME_FULLY_QUALIFIED, lexer.T_NAME_RELATIVE) {
		return nil
	}

	// 解析命名空间名称
	namespaceName, err := parseUseNamespaceName(p)
	if err != nil {
		p.errors = append(p.errors, err.Error())
		return nil
	}

	// 检查是否有别名 (as Alias)
	var alias string
	if p.peekToken.Type == lexer.T_AS {
		p.nextToken() // 移动到 as
		if !p.expectPeek(lexer.T_STRING) {
			return nil
		}
		alias = p.currentToken.Value
	}

	// 添加第一个 use 子句
	stmt.Uses = append(stmt.Uses, ast.UseClause{
		Name:  namespaceName,
		Alias: alias,
		Type:  useType,
	})

	// 处理更多 use 子句 (use A, B, C;)
	for p.peekToken.Type == lexer.TOKEN_COMMA {
		p.nextToken() // 移动到逗号
		if !p.expectPeekAny(lexer.T_STRING, lexer.T_NAME_QUALIFIED, lexer.T_NAME_FULLY_QUALIFIED, lexer.T_NAME_RELATIVE) {
			return nil
		}

		// 解析命名空间名称
		namespaceName, err := parseUseNamespaceName(p)
		if err != nil {
			p.errors = append(p.errors, err.Error())
			return nil
		}

		// 检查是否有别名
		alias := ""
		if p.peekToken.Type == lexer.T_AS {
			p.nextToken() // 移动到 as
			if !p.expectPeek(lexer.T_STRING) {
				return nil
			}
			alias = p.currentToken.Value
		}

		// 添加 use 子句
		stmt.Uses = append(stmt.Uses, ast.UseClause{
			Name:  namespaceName,
			Alias: alias,
			Type:  useType, // 所有子句共享相同的类型
		})
	}

	// 期望分号结尾
	if p.peekToken.Type == lexer.TOKEN_SEMICOLON {
		p.nextToken()
	}

	return stmt
}

// parseInterfaceDeclaration 解析接口声明
func parseInterfaceDeclaration(p *Parser) *ast.InterfaceDeclaration {
	pos := p.currentToken.Position

	// 期望接口名称
	if !p.expectPeek(lexer.T_STRING) {
		return nil
	}

	name := ast.NewIdentifierNode(p.currentToken.Position, p.currentToken.Value)
	interfaceDecl := ast.NewInterfaceDeclaration(pos, name)

	// 检查是否有 extends 子句
	if p.peekToken.Type == lexer.T_EXTENDS {
		p.nextToken() // 移动到 extends

		// 解析第一个父接口
		p.nextToken() // 移动到类名
		className, err := parseClassName(p)
		if err != nil {
			p.errors = append(p.errors, err.Error())
			return nil
		}
		parentInterface := ast.NewIdentifierNode(p.currentToken.Position, className)
		interfaceDecl.Extends = append(interfaceDecl.Extends, parentInterface)

		// 处理多个父接口 (interface Child extends Parent1, Parent2)
		for p.peekToken.Type == lexer.TOKEN_COMMA {
			p.nextToken() // 移动到逗号
			p.nextToken() // 移动到类名
			className, err := parseClassName(p)
			if err != nil {
				p.errors = append(p.errors, err.Error())
				return nil
			}
			parentInterface := ast.NewIdentifierNode(p.currentToken.Position, className)
			interfaceDecl.Extends = append(interfaceDecl.Extends, parentInterface)
		}
	}

	// 期望开始大括号
	if !p.expectPeek(lexer.TOKEN_LBRACE) {
		return nil
	}

	// 解析接口体（方法声明）
	for p.peekToken.Type != lexer.TOKEN_RBRACE && p.peekToken.Type != lexer.T_EOF {
		p.nextToken()

		// 跳过空行和注释
		if p.currentToken.Type == lexer.T_WHITESPACE {
			continue
		}

		// 解析方法声明
		method := parseInterfaceMethod(p)
		if method != nil {
			interfaceDecl.Methods = append(interfaceDecl.Methods, method)
		}
	}

	if !p.expectPeek(lexer.TOKEN_RBRACE) {
		return nil
	}

	return interfaceDecl
}

// parseInterfaceMethod 解析接口方法声明
func parseInterfaceMethod(p *Parser) *ast.InterfaceMethod {
	// 默认可见性为 public（接口方法都是 public）
	visibility := "public"

	// 检查可见性修饰符（虽然接口方法通常是 public）
	if p.currentToken.Type == lexer.T_PUBLIC || p.currentToken.Type == lexer.T_PRIVATE || p.currentToken.Type == lexer.T_PROTECTED {
		visibility = p.currentToken.Value
		p.nextToken()
	}

	// 期望 function 关键字
	if p.currentToken.Type != lexer.T_FUNCTION {
		return nil
	}

	// 期望方法名 (allow reserved keywords)
	p.nextToken()
	if p.currentToken.Type != lexer.T_STRING && !isSemiReserved(p.currentToken.Type) {
		p.errors = append(p.errors, fmt.Sprintf("expected method name, got %s at position %s",
			p.currentToken.Value, p.currentToken.Position))
		return nil
	}

	methodName := ast.NewIdentifierNode(p.currentToken.Position, p.currentToken.Value)

	// 期望左括号
	if !p.expectPeek(lexer.TOKEN_LPAREN) {
		return nil
	}

	method := &ast.InterfaceMethod{
		Name:       methodName,
		Parameters: make([]ast.Parameter, 0),
		Visibility: visibility,
	}

	// 解析参数列表
	if p.peekToken.Type != lexer.TOKEN_RPAREN {
		p.nextToken()

		// 解析第一个参数
		param := parseParameter(p)
		if param != nil {
			method.Parameters = append(method.Parameters, *param)
		}

		// 处理更多参数
		for p.peekToken.Type == lexer.TOKEN_COMMA {
			p.nextToken() // 移动到逗号
			
			// 检查逗号后面是否直接是结束符（支持尾随逗号）
			if p.peekToken.Type == lexer.TOKEN_RPAREN {
				break
			}
			
			p.nextToken() // 移动到下一个参数
			param := parseParameter(p)
			if param != nil {
				method.Parameters = append(method.Parameters, *param)
			}
		}
	}

	if !p.expectPeek(lexer.TOKEN_RPAREN) {
		return nil
	}

	// 检查返回类型声明
	if p.peekToken.Type == lexer.TOKEN_COLON {
		p.nextToken() // 移动到 ':'
		p.nextToken() // 移动到类型开始位置

		// 解析返回类型
		returnType := parseTypeHint(p)
		if returnType != nil {
			method.ReturnType = returnType
		}
	}

	// 期望分号（接口方法没有方法体）
	if p.peekToken.Type == lexer.TOKEN_SEMICOLON {
		p.nextToken()
	}

	return method
}

// parseTraitDeclaration 解析 trait 声明
func parseTraitDeclaration(p *Parser) *ast.TraitDeclaration {
	pos := p.currentToken.Position

	// 期望 trait 名称
	if !p.expectPeek(lexer.T_STRING) {
		return nil
	}

	name := ast.NewIdentifierNode(p.currentToken.Position, p.currentToken.Value)
	traitDecl := ast.NewTraitDeclaration(pos, name)

	// 期望开始大括号
	if !p.expectPeek(lexer.TOKEN_LBRACE) {
		return nil
	}

	// 解析 trait 体 - 使用与类相同的解析逻辑
	p.nextToken() // 进入类体
	for !p.currentTokenIs(lexer.TOKEN_RBRACE) && !p.isAtEnd() {
		stmt := parseClassStatement(p)
		if stmt != nil {
			// 根据语句类型添加到相应的列表
			switch s := stmt.(type) {
			case *ast.FunctionDeclaration:
				traitDecl.Methods = append(traitDecl.Methods, s)
			case *ast.PropertyDeclaration:
				traitDecl.Properties = append(traitDecl.Properties, s)
			case *ast.UseTraitStatement:
				// trait 使用其他 trait 的语句，添加到 Body
				traitDecl.Body = append(traitDecl.Body, s)
			case *ast.ClassConstantDeclaration:
				// trait 常量声明，添加到 Body
				traitDecl.Body = append(traitDecl.Body, s)
			default:
				// 其他语句类型
				traitDecl.Body = append(traitDecl.Body, s)
			}
		}
		p.nextToken()
	}

	return traitDecl
}

// parseTraitProperty 解析 trait 属性
func parseTraitProperty(p *Parser, visibility string) *ast.PropertyDeclaration {
	// 检查类型提示
	var typeHint *ast.TypeHint
	if isTypeToken(p.currentToken.Type) {
		typeHint = parseTypeHint(p)
		if !p.expectPeek(lexer.T_VARIABLE) {
			return nil
		}
	} else if p.currentToken.Type != lexer.T_VARIABLE {
		return nil
	}

	// 解析变量名（去掉 $ 符号）
	varName := p.currentToken.Value
	if len(varName) > 1 && varName[0] == '$' {
		varName = varName[1:]
	}

	pos := p.currentToken.Position
	property := ast.NewPropertyDeclaration(pos, visibility, varName, false, false, typeHint, nil)

	// 检查是否有默认值
	if p.peekToken.Type == lexer.TOKEN_EQUAL {
		p.nextToken() // 移动到 =
		p.nextToken() // 移动到默认值
		property.DefaultValue = parseExpression(p, LOWEST)
	}

	// 期望分号
	if p.peekToken.Type == lexer.TOKEN_SEMICOLON {
		p.nextToken()
	}

	return property
}

// parseMatchToken handles T_MATCH tokens - can be either match expressions or identifiers
// Based on PHP's grammar: expr includes `match` rule, and `identifier` includes semi_reserved tokens
func parseMatchToken(p *Parser) ast.Expression {
	// Look ahead: if T_MATCH is followed by '(', it's a match expression
	// Otherwise, it's used as an identifier (semi-reserved keyword)
	if p.peekToken.Type == lexer.TOKEN_LPAREN {
		return parseMatchExpression(p)
	}
	
	// Treat as identifier (semi-reserved keyword usage)
	return parseIdentifier(p)
}


// isConstantDeclaration 检查当前位置是否为常量声明
// 例如: public const, private const, protected const
func isConstantDeclaration(p *Parser) bool {
	// 如果当前 token 是 const，则肯定是常量声明
	if p.currentToken.Type == lexer.T_CONST {
		return true
	}
	
	// 如果当前 token 是可见性修饰符，检查下一个 token 是否为 const
	if p.currentToken.Type == lexer.T_PUBLIC ||
		p.currentToken.Type == lexer.T_PRIVATE ||
		p.currentToken.Type == lexer.T_PROTECTED {
		return p.peekToken.Type == lexer.T_CONST
	}
	
	return false
}

// parseEnumDeclaration 解析 enum 声明 (PHP 8.1+)
func parseEnumDeclaration(p *Parser) *ast.EnumDeclaration {
	pos := p.currentToken.Position

	// 期望 enum 名称
	if !p.expectPeek(lexer.T_STRING) {
		return nil
	}

	name := ast.NewIdentifierNode(p.currentToken.Position, p.currentToken.Value)
	enumDecl := ast.NewEnumDeclaration(pos, name)

	// 检查是否有支撑类型 (: string 或 : int)
	if p.peekToken.Type == lexer.TOKEN_COLON {
		p.nextToken() // 移动到 ':'
		p.nextToken() // 移动到类型

		backingType := parseTypeHint(p)
		if backingType != nil {
			enumDecl.BackingType = backingType
		}
	}

	// 检查是否有 implements 子句
	if p.peekToken.Type == lexer.T_IMPLEMENTS {
		p.nextToken() // 移动到 implements

		// 解析第一个接口
		if !p.expectPeek(lexer.T_STRING) {
			return nil
		}
		interface1 := ast.NewIdentifierNode(p.currentToken.Position, p.currentToken.Value)
		enumDecl.Implements = append(enumDecl.Implements, interface1)

		// 处理多个接口 (enum Status implements Interface1, Interface2)
		for p.peekToken.Type == lexer.TOKEN_COMMA {
			p.nextToken() // 移动到逗号
			if !p.expectPeek(lexer.T_STRING) {
				return nil
			}
			interfaceN := ast.NewIdentifierNode(p.currentToken.Position, p.currentToken.Value)
			enumDecl.Implements = append(enumDecl.Implements, interfaceN)
		}
	}

	// 期望开始大括号
	if !p.expectPeek(lexer.TOKEN_LBRACE) {
		return nil
	}

	// 解析 enum 体（案例和方法）
	for p.peekToken.Type != lexer.TOKEN_RBRACE && p.peekToken.Type != lexer.T_EOF {
		p.nextToken()

		// 跳过空行和注释
		if p.currentToken.Type == lexer.T_WHITESPACE {
			continue
		}

		if p.currentToken.Type == lexer.T_CASE {
			// 解析 enum 案例
			enumCase := parseEnumCase(p)
			if enumCase != nil {
				enumDecl.Cases = append(enumDecl.Cases, enumCase)
			}
		} else if p.currentToken.Type == lexer.T_FUNCTION ||
			p.currentToken.Type == lexer.T_PUBLIC ||
			p.currentToken.Type == lexer.T_PRIVATE ||
			p.currentToken.Type == lexer.T_PROTECTED {
			// 检查这是方法还是常量声明
			if isConstantDeclaration(p) {
				// 解析 enum 常量声明
				constant := parseClassConstantDeclaration(p)
				if constant != nil {
					if constDecl, ok := constant.(*ast.ClassConstantDeclaration); ok {
						enumDecl.Constants = append(enumDecl.Constants, constDecl)
					}
				}
			} else {
				// 解析 enum 方法
				method := parseFunctionDeclaration(p)
				if method != nil {
					enumDecl.Methods = append(enumDecl.Methods, method)
				}
			}
		} else if p.currentToken.Type == lexer.T_CONST {
			// 解析没有可见性修饰符的常量声明
			constant := parseClassConstantDeclaration(p)
			if constant != nil {
				if constDecl, ok := constant.(*ast.ClassConstantDeclaration); ok {
					enumDecl.Constants = append(enumDecl.Constants, constDecl)
				}
			}
		}
	}

	if !p.expectPeek(lexer.TOKEN_RBRACE) {
		return nil
	}

	return enumDecl
}

// parseEnumCase 解析 enum 案例
func parseEnumCase(p *Parser) *ast.EnumCase {
	// 期望案例名称（可以是T_STRING或半保留关键字）
	p.nextToken()
	if p.currentToken.Type != lexer.T_STRING && !isSemiReserved(p.currentToken.Type) {
		p.peekError(lexer.T_STRING)
		return nil
	}

	caseName := ast.NewIdentifierNode(p.currentToken.Position, p.currentToken.Value)
	var caseValue ast.Expression

	// 检查是否有值 (case SUCCESS = 'success')
	if p.peekToken.Type == lexer.TOKEN_EQUAL {
		p.nextToken() // 移动到 =
		p.nextToken() // 移动到值
		caseValue = parseExpression(p, LOWEST)
	}

	// 期望分号
	if p.peekToken.Type == lexer.TOKEN_SEMICOLON {
		p.nextToken()
	}

	return ast.NewEnumCase(caseName, caseValue)
}

// parsePropertyAccess 解析属性访问表达式
// Supports: $obj->property, $obj->$variable, $obj->{expression}
func parsePropertyAccess(p *Parser, left ast.Expression) ast.Expression {
	pos := p.currentToken.Position
	p.nextToken()

	var property ast.Expression

	switch p.currentToken.Type {
	case lexer.T_STRING:
		// Simple property name: $obj->property
		property = ast.NewIdentifierNode(p.currentToken.Position, p.currentToken.Value)

	case lexer.T_VARIABLE:
		// Variable property name: $obj->$property_name
		// Parse as a variable expression
		property = parseVariable(p)

	case lexer.TOKEN_LBRACE:
		// Brace-enclosed expression: $obj->{$expr} or $obj->{"property"}
		p.nextToken()
		property = parseExpression(p, LOWEST)
		if !p.expectPeek(lexer.TOKEN_RBRACE) {
			return nil
		}

	default:
		// Check if it's a semi-reserved keyword (can be used as property name)
		if isSemiReserved(p.currentToken.Type) {
			property = ast.NewIdentifierNode(p.currentToken.Position, p.currentToken.Value)
		} else {
			p.errors = append(p.errors, fmt.Sprintf("expected property name, variable, or brace-enclosed expression, got %s at position %s",
				p.currentToken.Value, p.currentToken.Position))
			return nil
		}
	}

	return ast.NewPropertyAccessExpression(pos, left, property)
}

// parseNullsafePropertyAccess 解析空安全属性访问表达式
// Supports: $obj?->property, $obj?->$variable, $obj?->{expression}
func parseNullsafePropertyAccess(p *Parser, left ast.Expression) ast.Expression {
	pos := p.currentToken.Position
	p.nextToken()

	var property ast.Expression

	switch p.currentToken.Type {
	case lexer.T_STRING:
		// Simple property name: $obj?->property
		property = ast.NewIdentifierNode(p.currentToken.Position, p.currentToken.Value)

	case lexer.T_VARIABLE:
		// Variable property name: $obj?->$property_name
		// Parse as a variable expression
		property = parseVariable(p)

	case lexer.TOKEN_LBRACE:
		// Brace-enclosed expression: $obj?->{$expr} or $obj?->{"property"}
		p.nextToken()
		property = parseExpression(p, LOWEST)
		if !p.expectPeek(lexer.TOKEN_RBRACE) {
			return nil
		}

	default:
		// Check if it's a semi-reserved keyword (can be used as property name)
		if isSemiReserved(p.currentToken.Type) {
			property = ast.NewIdentifierNode(p.currentToken.Position, p.currentToken.Value)
		} else {
			p.errors = append(p.errors, fmt.Sprintf("expected property name, variable, or brace-enclosed expression, got %s at position %s",
				p.currentToken.Value, p.currentToken.Position))
			return nil
		}
	}

	return ast.NewNullsafePropertyAccessExpression(pos, left, property)
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

// parseEndifStatement 解析 endif 语句
func parseEndifStatement(p *Parser) ast.Statement {
	pos := p.currentToken.Position

	if p.peekToken.Type == lexer.TOKEN_SEMICOLON {
		p.nextToken()
	}

	// Create a simple statement to represent the endif
	return ast.NewExpressionStatement(pos, ast.NewIdentifierNode(pos, "endif"))
}

// parseHaltCompilerStatement 解析 __halt_compiler() 语句
func parseHaltCompilerStatement(p *Parser) ast.Statement {
	pos := p.currentToken.Position

	// 期望 (
	if !p.expectPeek(lexer.TOKEN_LPAREN) {
		return nil
	}

	// 期望 )
	if !p.expectPeek(lexer.TOKEN_RPAREN) {
		return nil
	}

	// 期望 ;
	if !p.expectSemicolon() {
		return nil
	}

	// 创建 halt compiler 语句
	haltStmt := ast.NewHaltCompilerStatement(pos)

	// 停止解析后续代码 - 设置当前token为EOF来终止解析
	// 这样模拟PHP中__halt_compiler()的行为
	p.currentToken = lexer.Token{Type: lexer.T_EOF, Position: pos}
	p.peekToken = lexer.Token{Type: lexer.T_EOF, Position: pos}

	return haltStmt
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

	// 分号是可选的，特别是在关闭标签之前
	if p.peekToken.Type == lexer.TOKEN_SEMICOLON {
		p.nextToken()
	} else if p.peekToken.Type == lexer.T_CLOSE_TAG {
		// 在关闭标签之前分号是可选的，按PHP规范
		// 不消耗关闭标签，让主循环处理
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

	for p.peekToken.Type != lexer.TOKEN_SEMICOLON && p.peekToken.Type != lexer.T_OPEN_TAG && precedence < p.peekPrecedence() {
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
	pos := p.currentToken.Position

	// Check if this is part of a qualified namespace (Parse\Date)
	if p.peekToken.Type == lexer.T_NS_SEPARATOR {
		// Parse as qualified name using parseClassName
		className, err := parseClassName(p)
		if err != nil {
			// Fallback to simple identifier if parsing fails
			return ast.NewIdentifierNode(pos, p.currentToken.Value)
		}
		return ast.NewIdentifierNode(pos, className)
	}

	// Simple identifier
	return ast.NewIdentifierNode(pos, p.currentToken.Value)
}

// isReservedNonModifier 检查是否为保留非修饰符关键字
func isReservedNonModifier(tokenType lexer.TokenType) bool {
	switch tokenType {
	case lexer.T_INCLUDE, lexer.T_INCLUDE_ONCE, lexer.T_EVAL, lexer.T_REQUIRE, lexer.T_REQUIRE_ONCE,
		lexer.T_LOGICAL_OR, lexer.T_LOGICAL_XOR, lexer.T_LOGICAL_AND,
		lexer.T_INSTANCEOF, lexer.T_NEW, lexer.T_CLONE, lexer.T_EXIT, lexer.T_IF, lexer.T_ELSEIF, lexer.T_ELSE, lexer.T_ENDIF,
		lexer.T_ECHO, lexer.T_DO, lexer.T_WHILE, lexer.T_ENDWHILE,
		lexer.T_FOR, lexer.T_ENDFOR, lexer.T_FOREACH, lexer.T_ENDFOREACH, lexer.T_DECLARE, lexer.T_ENDDECLARE,
		lexer.T_AS, lexer.T_TRY, lexer.T_CATCH, lexer.T_FINALLY,
		lexer.T_THROW, lexer.T_USE, lexer.T_INSTEADOF, lexer.T_GLOBAL, lexer.T_VAR, lexer.T_UNSET, lexer.T_ISSET, lexer.T_EMPTY,
		lexer.T_CONTINUE, lexer.T_GOTO,
		lexer.T_FUNCTION, lexer.T_CONST, lexer.T_RETURN, lexer.T_PRINT, lexer.T_YIELD, lexer.T_LIST,
		lexer.T_SWITCH, lexer.T_ENDSWITCH, lexer.T_CASE, lexer.T_DEFAULT, lexer.T_BREAK,
		lexer.T_ARRAY, lexer.T_CALLABLE, lexer.T_EXTENDS, lexer.T_IMPLEMENTS, lexer.T_NAMESPACE, lexer.T_TRAIT, lexer.T_INTERFACE, lexer.T_CLASS,
		lexer.T_CLASS_C, lexer.T_TRAIT_C, lexer.T_FUNC_C, lexer.T_METHOD_C, lexer.T_LINE, lexer.T_FILE, lexer.T_DIR, lexer.T_NS_C, lexer.T_FN, lexer.T_MATCH, lexer.T_ENUM,
		lexer.T_PROPERTY_C:
		return true
	default:
		return false
	}
}

// isSemiReserved 检查是否为半保留关键字
func isSemiReserved(tokenType lexer.TokenType) bool {
	return isReservedNonModifier(tokenType) ||
		tokenType == lexer.T_STATIC || tokenType == lexer.T_ABSTRACT || tokenType == lexer.T_FINAL ||
		tokenType == lexer.T_PRIVATE || tokenType == lexer.T_PROTECTED || tokenType == lexer.T_PUBLIC ||
		tokenType == lexer.T_READONLY
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

	// 解析第一个元素，可能是展开语法
	if p.currentToken.Type == lexer.T_ELLIPSIS {
		pos := p.currentToken.Position
		p.nextToken() // 跳过 ...
		expr := parseExpression(p, LOWEST)
		if expr != nil {
			spreadExpr := ast.NewSpreadExpression(pos, expr)
			array.Elements = append(array.Elements, spreadExpr)
		}
	} else {
		array.Elements = append(array.Elements, parseExpression(p, LOWEST))
	}

	for p.peekToken.Type == lexer.TOKEN_COMMA {
		p.nextToken() // 移动到逗号

		// 检查是否为尾随逗号：如果逗号后面是右括号，则不解析更多元素
		if p.peekToken.Type == lexer.TOKEN_RPAREN {
			break
		}

		p.nextToken() // 移动到下一个元素

		// 检查展开语法
		if p.currentToken.Type == lexer.T_ELLIPSIS {
			pos := p.currentToken.Position
			p.nextToken() // 跳过 ...
			expr := parseExpression(p, LOWEST)
			if expr != nil {
				spreadExpr := ast.NewSpreadExpression(pos, expr)
				array.Elements = append(array.Elements, spreadExpr)
			}
		} else {
			array.Elements = append(array.Elements, parseExpression(p, LOWEST))
		}
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

	// 检查第一类可调用语法 function(...)
	if p.peekToken.Type == lexer.T_ELLIPSIS {
		p.nextToken() // 移动到 ...
		if p.peekToken.Type == lexer.TOKEN_RPAREN {
			// 确实是 function(...) 语法
			p.nextToken() // 跳过 )
			return ast.NewFirstClassCallable(pos, fn)
		}
		// 如果不是 function(...) 语法，当前token现在是T_ELLIPSIS
		// 我们需要让parseExpressionList知道已经前进了一步
	} else {
		// 正常情况，当前token是(，需要前进到第一个参数或)
	}

	call := ast.NewCallExpression(pos, fn)

	// 解析参数列表
	call.Arguments = parseExpressionList(p, lexer.TOKEN_RPAREN)

	return call
}

// parseNamedArgument 解析命名参数 (PHP 8.0+) - name: value
func parseNamedArgument(p *Parser) ast.Expression {
	pos := p.currentToken.Position

	// 解析参数名
	name := ast.NewIdentifierNode(p.currentToken.Position, p.currentToken.Value)

	// 期望冒号
	if !p.expectPeek(lexer.TOKEN_COLON) {
		return nil
	}

	// 解析参数值
	p.nextToken()
	value := parseExpression(p, LOWEST)
	if value == nil {
		return nil
	}

	return ast.NewNamedArgument(pos, name, value)
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

// parseHeredoc 解析Heredoc（支持变量插值）
func parseHeredoc(p *Parser) ast.Expression {
	startPos := p.currentToken.Position
	var parts []ast.Expression

	// 跳过 T_START_HEREDOC
	p.nextToken()

	// 解析heredoc内容，直到遇到T_END_HEREDOC
	for p.currentToken.Type != lexer.T_END_HEREDOC && !p.isAtEnd() {
		switch p.currentToken.Type {
		case lexer.T_ENCAPSED_AND_WHITESPACE:
			// 字符串片段
			if p.currentToken.Value != "" {
				stringPart := ast.NewStringLiteral(p.currentToken.Position, p.currentToken.Value, p.currentToken.Value)
				parts = append(parts, stringPart)
			}
		case lexer.T_VARIABLE:
			// 直接变量插值 $var
			variable := ast.NewVariable(p.currentToken.Position, p.currentToken.Value)
			parts = append(parts, variable)
		case lexer.T_CURLY_OPEN:
			// 复杂变量插值 {$var} 或 {$var['key']} 或 {$var->prop}
			p.nextToken() // 跳过 T_CURLY_OPEN

			// 解析大括号内的表达式
			expr := parseExpression(p, LOWEST)
			if expr != nil {
				parts = append(parts, expr)
			}

			// 期望结束的 }
			if p.peekToken.Type == lexer.TOKEN_RBRACE {
				p.nextToken() // 移动到 }
			} else {
				p.errors = append(p.errors, "expected '}' after expression in heredoc interpolation")
				return nil
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

	// 检查是否正确结束
	if p.currentToken.Type != lexer.T_END_HEREDOC {
		p.errors = append(p.errors, "expected T_END_HEREDOC in heredoc")
		return nil
	}

	// 如果只有一个部分且是简单字符串，返回字符串字面量
	if len(parts) == 1 {
		if stringLit, ok := parts[0].(*ast.StringLiteral); ok {
			return stringLit
		}
	}

	// 返回字符串插值表达式
	return ast.NewInterpolatedStringExpression(startPos, parts)
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
			p.errors = append(p.errors, fmt.Sprintf("expected variable in static statement, got %s at position %s", p.currentToken.Value, p.currentToken.Position))
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

	if !p.expectSemicolon() {
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
	if !p.expectSemicolon() {
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

	// 检查是否为Alternative语法 (foreach (...):)
	if p.peekToken.Type == lexer.TOKEN_COLON {
		p.nextToken() // 移动到 :
		return parseAlternativeForeachStatement(p, pos, iterable, key, value)
	}

	// 普通语法 (foreach (...) { ... })
	p.nextToken()
	body := parseStatement(p)

	return ast.NewForeachStatement(pos, iterable, key, value, body)
}

// parseAlternativeForeachStatement 解析Alternative语法的foreach语句 (foreach: ... endforeach;)
func parseAlternativeForeachStatement(p *Parser, pos lexer.Position, iterable, key, value ast.Expression) *ast.AlternativeForeachStatement {
	altForeachStmt := ast.NewAlternativeForeachStatement(pos, iterable, value)
	altForeachStmt.Key = key

	// 解析foreach块的语句列表
	for p.peekToken.Type != lexer.T_ENDFOREACH {
		if p.peekToken.Type == lexer.T_EOF {
			p.errors = append(p.errors, "Expected endforeach, but reached end of file")
			return nil
		}
		p.nextToken()
		if stmt := parseStatement(p); stmt != nil {
			altForeachStmt.Body = append(altForeachStmt.Body, stmt)
		}
	}

	// 期望 endforeach
	if !p.expectPeek(lexer.T_ENDFOREACH) {
		return nil
	}

	// 期望分号
	if !p.expectSemicolon() {
		return nil
	}

	return altForeachStmt
}

// parseDeclareStatement 解析declare语句，支持普通语法和Alternative语法
func parseDeclareStatement(p *Parser) ast.Statement {
	pos := p.currentToken.Position

	if !p.expectPeek(lexer.TOKEN_LPAREN) {
		return nil
	}

	// 解析声明列表 (例如: ticks=1, strict_types=1)
	var declarations []ast.Expression

	p.nextToken()
	for {
		// 解析声明 (identifier = expression)
		if p.currentToken.Type != lexer.T_STRING {
			p.errors = append(p.errors, "Expected identifier in declare statement")
			return nil
		}

		name := ast.NewIdentifierNode(p.currentToken.Position, p.currentToken.Value)

		if !p.expectPeek(lexer.TOKEN_EQUAL) {
			return nil
		}

		p.nextToken()
		value := parseExpression(p, LOWEST)

		// 创建赋值表达式作为声明
		declaration := ast.NewAssignmentExpression(name.GetPosition(), name, "=", value)
		declarations = append(declarations, declaration)

		p.nextToken()
		if p.currentToken.Type == lexer.TOKEN_RPAREN {
			break
		} else if p.currentToken.Type == lexer.TOKEN_COMMA {
			p.nextToken()
			continue
		} else {
			p.errors = append(p.errors, "Expected ',' or ')' in declare statement")
			return nil
		}
	}

	// 检查是否为Alternative语法 (declare (...):)
	if p.peekToken.Type == lexer.TOKEN_COLON {
		p.nextToken() // 移动到 :
		return parseAlternativeDeclareStatement(p, pos, declarations)
	}

	// 普通语法可以是: declare(...); 或 declare(...) { ... }
	if p.peekToken.Type == lexer.TOKEN_SEMICOLON {
		p.nextToken() // 移动到 ;
		return ast.NewDeclareStatement(pos, declarations, false)
	} else if p.peekToken.Type == lexer.TOKEN_LBRACE {
		p.nextToken() // 移动到 {
		declareStmt := ast.NewDeclareStatement(pos, declarations, false)
		declareStmt.Body = parseBlockStatements(p)
		return declareStmt
	} else {
		p.errors = append(p.errors, "Expected ';' or '{' after declare")
		return nil
	}
}

// parseAlternativeDeclareStatement 解析Alternative语法的declare语句 (declare(): ... enddeclare;)
func parseAlternativeDeclareStatement(p *Parser, pos lexer.Position, declarations []ast.Expression) *ast.DeclareStatement {
	declareStmt := ast.NewDeclareStatement(pos, declarations, true)

	// 解析declare块的语句列表
	for p.peekToken.Type != lexer.T_ENDDECLARE {
		if p.peekToken.Type == lexer.T_EOF {
			p.errors = append(p.errors, "Expected enddeclare, but reached end of file")
			return nil
		}
		p.nextToken()
		if stmt := parseStatement(p); stmt != nil {
			declareStmt.Body = append(declareStmt.Body, stmt)
		}
	}

	// 期望 enddeclare
	if !p.expectPeek(lexer.T_ENDDECLARE) {
		return nil
	}

	// 期望分号
	if !p.expectSemicolon() {
		return nil
	}

	return declareStmt
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

	// 检查是否为 Alternative 语法 (switch (...):)
	if p.peekToken.Type == lexer.TOKEN_COLON {
		p.nextToken() // 移动到 :
		return parseAlternativeSwitchStatement(p, pos, discriminant)
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
			exceptionType := parseQualifiedName(p)
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

		// Check what's next after the catch clause
		// We're currently on the closing brace of the catch block
		if p.peekToken.Type == lexer.T_CATCH {
			p.nextToken() // Move past the closing brace
			// Now currentToken should be the next T_CATCH
		} else {
			// No more catch clauses, we'll exit the loop
			// currentToken is on the closing brace of the last catch
			break
		}
	}

	// Check if there's a finally block (advance to check)
	if p.currentToken.Type != lexer.T_FINALLY && p.peekToken.Type == lexer.T_FINALLY {
		p.nextToken() // Advance to the finally token
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
		// After finally block, we're on the closing brace
	}

	return tryStmt
}

// parseThrowStatement 解析throw语句
func parseThrowStatement(p *Parser) ast.Statement {
	pos := p.currentToken.Position

	p.nextToken()
	argument := parseExpression(p, LOWEST)

	if !p.expectSemicolon() {
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

	if !p.expectSemicolon() {
		return nil
	}

	return ast.NewGotoStatement(pos, label)
}

// parseNewExpression 解析new表达式
// isAnonymousClassPattern 检查是否是匿名类模式
func isAnonymousClassPattern(p *Parser) bool {
	// 检查修饰符或属性，这些肯定是匿名类
	if p.peekToken.Type == lexer.T_FINAL ||
		p.peekToken.Type == lexer.T_ABSTRACT ||
		p.peekToken.Type == lexer.T_READONLY ||
		p.peekToken.Type == lexer.T_ATTRIBUTE {
		return true
	}
	
	// 如果下一个token是 T_CLASS，需要向前看来确定
	// 这已经在parseNewExpression中处理，所以这里直接检查T_CLASS
	return p.peekToken.Type == lexer.T_CLASS
}

func parseNewExpression(p *Parser) ast.Expression {
	pos := p.currentToken.Position

	// 检查是否是匿名类 "new [modifiers] class"
	if isAnonymousClassPattern(p) {
		// 对于T_CLASS的情况，我们需要更仔细地检查
		// 但由于无法安全地进行lookahead而不破坏解析状态，
		// 我们会在parseAnonymousClass中处理这种边缘情况
		return parseAnonymousClass(p, pos)
	}

	p.nextToken()
	class := parseExpression(p, PREFIX)

	newExpr := ast.NewNewExpression(pos, class)

	// 检查是否有构造函数参数（支持trailing comma）
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
				// 检查trailing comma情况
				if p.currentToken.Type == lexer.TOKEN_RPAREN {
					break // 允许trailing comma
				}
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
	right := parseQualifiedName(p)

	return ast.NewInstanceofExpression(pos, left, right)
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

// parseAttributeExpression 解析属性表达式 #[AttributeName(args)] 或 #[Attr1, Attr2, ...]
func parseAttributeExpression(p *Parser) ast.Expression {
	// 当前 token 是 T_ATTRIBUTE (#[)
	// 解析属性组：可能包含多个属性，用逗号分隔
	attributeGroup := parseAttributeGroup(p)
	if attributeGroup == nil {
		return nil
	}

	return attributeGroup
}

// parseAttributeGroup 解析单个属性组 #[Attr1, Attr2, ...]
func parseAttributeGroup(p *Parser) *ast.AttributeGroup {
	pos := p.currentToken.Position
	var attributes []*ast.Attribute

	// 当前token应该是 T_ATTRIBUTE (#[)，需要移动到第一个属性名
	// 属性名遵循 class_name 语法规则：T_STATIC | T_STRING | T_NAME_* | T_NS_SEPARATOR
	p.nextToken() // 移动到属性名的第一个token
	if !isClassNameToken(p.currentToken.Type) {
		p.errors = append(p.errors, fmt.Sprintf("expected attribute name, got `%s` instead at position %s", p.currentToken.Value, p.currentToken.Position))
		return nil
	}

	firstAttr := parseAttributeDecl(p)
	if firstAttr == nil {
		return nil
	}
	attributes = append(attributes, firstAttr)

	// 解析可能的后续属性（用逗号分隔）
	for p.peekToken.Type == lexer.TOKEN_COMMA {
		p.nextToken() // 移动到逗号
		p.nextToken() // 移动到下一个属性名
		if !isClassNameToken(p.currentToken.Type) {
			p.errors = append(p.errors, fmt.Sprintf("expected attribute name after comma, got `%s` instead at position %s", p.currentToken.Value, p.currentToken.Position))
			break
		}

		attr := parseAttributeDecl(p)
		if attr != nil {
			attributes = append(attributes, attr)
		}
	}

	// 期望结束符 ]
	if !p.expectPeek(lexer.TOKEN_RBRACKET) {
		return nil
	}

	return ast.NewAttributeGroup(pos, attributes)
}

// parseClassName 解析类名，支持所有类型的名称（根据 PHP 语法中的 class_name 规则）
// class_name: T_STATIC | name
// name: T_STRING | T_NAME_QUALIFIED | T_NAME_FULLY_QUALIFIED | T_NAME_RELATIVE
func parseClassName(p *Parser) (string, error) {
	var className string

	switch p.currentToken.Type {
	case lexer.T_STATIC:
		return "static", nil
	case lexer.T_STRING:
		className = p.currentToken.Value

		// 检查是否是相对限定名 (WpOrg\Requests\Requests)
		for p.peekToken.Type == lexer.T_NS_SEPARATOR {
			p.nextToken()                     // 移动到 \
			className += p.currentToken.Value // 添加 "\"
			p.nextToken()                     // 移动到下一个名称部分
			if p.currentToken.Type != lexer.T_STRING {
				return "", fmt.Errorf("expected class name after namespace separator, got `%s` instead at position %s", p.currentToken.Value, p.currentToken.Position)
			}
			className += p.currentToken.Value
		}
		return className, nil
	case lexer.T_NAME_FULLY_QUALIFIED, lexer.T_NAME_QUALIFIED, lexer.T_NAME_RELATIVE:
		return p.currentToken.Value, nil
	case lexer.T_NS_SEPARATOR:
		// 处理由多个token组成的完全限定名称: \ + 名称部分
		className = p.currentToken.Value // 开始于 "\"
		p.nextToken()

		// 解析完整的命名空间路径: Namespace\SubNamespace\ClassName
		for {
			if p.currentToken.Type != lexer.T_STRING {
				return "", fmt.Errorf("expected class name after namespace separator, got `%s` instead at position %s", p.currentToken.Value, p.currentToken.Position)
			}
			className += p.currentToken.Value

			// 检查是否还有更多的命名空间分隔符
			if p.peekToken.Type == lexer.T_NS_SEPARATOR {
				p.nextToken()                     // 移动到 \
				className += p.currentToken.Value // 添加 "\"
				p.nextToken()                     // 移动到下一个名称部分
			} else {
				break
			}
		}
		return className, nil
	default:
		return "", fmt.Errorf("expected class name, got `%s` instead at position %s", p.currentToken.Value, p.currentToken.Position)
	}
}

// parseAttributeDecl 解析单个属性声明 AttributeName(args)
func parseAttributeDecl(p *Parser) *ast.Attribute {
	pos := p.currentToken.Position

	// 使用 class_name 语法规则解析属性名
	attributeName, err := parseClassName(p)
	if err != nil {
		p.errors = append(p.errors, err.Error())
		return nil
	}

	name := ast.NewIdentifierNode(pos, attributeName)

	var arguments []ast.Expression

	// 检查是否有参数列表
	if p.peekToken.Type == lexer.TOKEN_LPAREN {
		p.nextToken() // 移动到 (

		// 解析参数列表
		if p.peekToken.Type != lexer.TOKEN_RPAREN {
			p.nextToken()

			// Check for named argument (identifier: value)
			if p.isNamedArgumentStart() {
				arg := parseNamedArgument(p)
				if arg != nil {
					arguments = append(arguments, arg)
				}
			} else {
				arguments = append(arguments, parseExpression(p, LOWEST))
			}

			for p.peekToken.Type == lexer.TOKEN_COMMA {
				p.nextToken() // 移动到逗号
				
				// 检查是否为尾随逗号 (trailing comma)
				if p.peekToken.Type == lexer.TOKEN_RPAREN {
					break // 如果逗号后面是右括号，说明是尾随逗号，跳出循环
				}
				
				p.nextToken() // 移动到下一个参数

				// Check for named argument
				if p.isNamedArgumentStart() {
					arg := parseNamedArgument(p)
					if arg != nil {
						arguments = append(arguments, arg)
					}
				} else {
					arguments = append(arguments, parseExpression(p, LOWEST))
				}
			}
		}

		if !p.expectPeek(lexer.TOKEN_RPAREN) {
			return nil
		}
	}

	return ast.NewAttribute(pos, name, arguments)
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

		// Check if we've reached the end after skipping comments
		if p.currentToken.Type == lexer.TOKEN_RBRACKET {
			break
		}

		// 解析数组元素表达式
		// 检查是否是展开语法：T_ELLIPSIS expr
		if p.currentToken.Type == lexer.T_ELLIPSIS {
			pos := p.currentToken.Position
			p.nextToken() // 跳过 ...

			// 解析展开的表达式
			expr := parseExpression(p, LOWEST)
			if expr != nil {
				// 创建展开表达式节点
				spreadExpr := ast.NewSpreadExpression(pos, expr)
				array.Elements = append(array.Elements, spreadExpr)
			}
		} else {
			element := parseExpression(p, LOWEST)
			if element != nil {
				array.Elements = append(array.Elements, element)
			}
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
// 根据PHP语法: T_ISSET '(' isset_variables possible_comma ')'
// 多个变量用 AND 连接：isset_variables ',' isset_variable -> zend_ast_create(ZEND_AST_AND, $1, $3)
func parseIssetExpression(p *Parser) ast.Expression {
	pos := p.currentToken.Position

	if !p.expectToken(lexer.TOKEN_LPAREN) {
		return nil
	}

	// 至少需要一个参数
	if p.peekToken.Type == lexer.TOKEN_RPAREN {
		p.errors = append(p.errors, "isset() expects at least 1 parameter, 0 given")
		return nil
	}

	p.nextToken()

	// 创建第一个 isset 检查
	firstExpr := parseExpression(p, LOWEST)
	var result ast.Expression = ast.NewIssetExpression(pos, []ast.Expression{firstExpr})

	// 处理多个参数 - 用 AND 连接
	for p.peekToken.Type == lexer.TOKEN_COMMA {
		p.nextToken() // 跳过逗号

		// 检查是否有可选的尾随逗号 (possible_comma)
		if p.peekToken.Type == lexer.TOKEN_RPAREN {
			break
		}

		p.nextToken() // 移动到下一个参数
		nextExpr := parseExpression(p, LOWEST)
		nextIsset := ast.NewIssetExpression(pos, []ast.Expression{nextExpr})

		// 用 AND 连接多个 isset 检查
		result = ast.NewBinaryExpression(pos, result, "&&", nextIsset)
	}

	if !p.expectToken(lexer.TOKEN_RPAREN) {
		return nil
	}

	return result
}

// parseListExpression 解析 list() 表达式
func parseListExpression(p *Parser) ast.Expression {
	pos := p.currentToken.Position

	if !p.expectPeek(lexer.TOKEN_LPAREN) {
		return nil
	}

	elements := make([]ast.Expression, 0)

	// Handle empty list: list()
	if p.peekToken.Type == lexer.TOKEN_RPAREN {
		p.nextToken()
		listExpr := ast.NewArrayExpression(pos)
		listExpr.Elements = elements
		return listExpr
	}

	// Parse list elements with support for empty elements (consecutive commas)
	// The key insight: we need to handle the token stream correctly

	for p.peekToken.Type != lexer.TOKEN_RPAREN {
		// Check if the current position represents an empty element
		if p.peekToken.Type == lexer.TOKEN_COMMA {
			// Empty element (we're right before a comma)
			elements = append(elements, nil)
			p.nextToken() // move to the comma

			// After consuming the comma representing empty element,
			// check what's next
			if p.peekToken.Type == lexer.TOKEN_RPAREN {
				// This was a trailing comma after empty element
				break
			}
			// Continue to parse the next element
		} else {
			// Non-empty element: parse it
			p.nextToken() // move to the element
			elements = append(elements, parseExpression(p, LOWEST))

			// After parsing the element, check if there's more
			if p.peekToken.Type == lexer.TOKEN_COMMA {
				p.nextToken() // consume the comma

				// Check for trailing comma
				if p.peekToken.Type == lexer.TOKEN_RPAREN {
					break
				}
				// Otherwise continue to next element
			} else if p.peekToken.Type != lexer.TOKEN_RPAREN {
				// Unexpected token after element
				break
			}
		}
	}

	if !p.expectPeek(lexer.TOKEN_RPAREN) {
		return nil
	}

	// Create ArrayExpression for list() - PHP uses same AST structure
	listExpr := ast.NewArrayExpression(pos)
	listExpr.Elements = elements
	return listExpr
}

// parseAnonymousFunctionExpression 解析匿名函数表达式
func parseAnonymousFunctionExpression(p *Parser) ast.Expression {
	pos := p.currentToken.Position

	// 检查是否为引用返回匿名函数 function &()
	byReference := false
	if p.peekToken.Type == lexer.TOKEN_AMPERSAND {
		byReference = true
		p.nextToken() // 移动到 &
	}

	// 解析参数列表
	if !p.expectToken(lexer.TOKEN_LPAREN) {
		return nil
	}

	var parameters []ast.Parameter

	// 解析参数（支持类型提示）
	if p.peekToken.Type != lexer.TOKEN_RPAREN {
		p.nextToken()

		// 解析第一个参数
		param := parseParameter(p)
		if param != nil {
			parameters = append(parameters, *param)
		}

		// 处理更多参数
		for p.peekToken.Type == lexer.TOKEN_COMMA {
			p.nextToken() // 移动到逗号
			
			// 检查逗号后面是否直接是结束符（支持尾随逗号）
			if p.peekToken.Type == lexer.TOKEN_RPAREN {
				break
			}
			
			p.nextToken() // 移动到下一个参数
			param := parseParameter(p)
			if param != nil {
				parameters = append(parameters, *param)
			}
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
				} else if p.currentToken.Type == lexer.TOKEN_AMPERSAND {
					// 处理引用变量 &$var
					refExpr := parseReferenceExpression(p)
					useClause = append(useClause, refExpr)
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

	// 检查返回类型声明 - 支持多种类型格式
	if p.peekToken.Type == lexer.TOKEN_COLON {
		p.nextToken() // 跳过 ':'
		p.nextToken() // 移动到类型

		// 处理可空类型 ?Type
		if p.currentToken.Type == lexer.TOKEN_QUESTION {
			p.nextToken() // 移动到实际类型
		}

		// 解析类型（可以是 T_STRING, T_ARRAY 等）
		if p.currentToken.Type != lexer.T_STRING &&
			p.currentToken.Type != lexer.T_ARRAY &&
			p.currentToken.Type != lexer.T_CALLABLE {
			// 对于其他类型，尝试继续解析
		}

		// 处理联合类型和交叉类型
		for p.peekToken.Type == lexer.TOKEN_PIPE || p.peekToken.Type == lexer.TOKEN_AMPERSAND {
			p.nextToken() // 跳过 | 或 &
			p.nextToken() // 移动到下一个类型
		}
	}

	// 解析函数体
	if !p.expectToken(lexer.TOKEN_LBRACE) {
		return nil
	}

	body := parseBlockStatements(p)

	return ast.NewAnonymousFunctionExpression(pos, parameters, body, useClause, byReference)
}

// parseUseExpression 解析 use 表达式（在表达式上下文中）
func parseUseExpression(p *Parser) ast.Expression {
	pos := p.currentToken.Position

	// 这里简单处理为标识符，实际的 use 语句应该在语句级别处理
	return ast.NewIdentifierNode(pos, "use")
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

// parseClassDeclaration 解析类声明语句
func parseClassDeclaration(p *Parser) ast.Statement {
	pos := p.currentToken.Position

	// 跳过 'class' 关键字
	if !p.expectPeek(lexer.T_STRING) {
		return nil
	}

	// 解析类名
	name := ast.NewIdentifierNode(p.currentToken.Position, p.currentToken.Value)
	class := ast.NewClassExpression(pos, name, nil, nil, false, false) // final = false, readOnly = false

	// 检查是否有 extends 子句
	if p.peekToken.Type == lexer.T_EXTENDS {
		p.nextToken() // 跳过当前token移动到extends
		p.nextToken() // 移动到类名
		className, err := parseClassName(p)
		if err != nil {
			p.errors = append(p.errors, err.Error())
			return nil
		}
		extends := ast.NewIdentifierNode(p.currentToken.Position, className)
		class.Extends = extends
	}

	// 检查是否有 implements 子句
	if p.peekToken.Type == lexer.T_IMPLEMENTS {
		p.nextToken() // 移动到 implements
		p.nextToken() // 移动到第一个接口名

		interfaceName, err := parseClassName(p)
		if err != nil {
			p.errors = append(p.errors, err.Error())
			return nil
		}
		interfaceNode := ast.NewIdentifierNode(p.currentToken.Position, interfaceName)
		class.Implements = append(class.Implements, interfaceNode)

		// 处理多个接口，用逗号分隔
		for p.peekToken.Type == lexer.TOKEN_COMMA {
			p.nextToken() // 移动到逗号
			p.nextToken() // 移动到下一个接口名
			interfaceName, err := parseClassName(p)
			if err != nil {
				p.errors = append(p.errors, err.Error())
				break
			}
			interfaceNode := ast.NewIdentifierNode(p.currentToken.Position, interfaceName)
			class.Implements = append(class.Implements, interfaceNode)
		}
	}

	// 期望类体开始
	if !p.expectPeek(lexer.TOKEN_LBRACE) {
		return nil
	}

	// 解析类体
	p.nextToken() // 进入类体
	for !p.currentTokenIs(lexer.TOKEN_RBRACE) && !p.isAtEnd() {
		stmt := parseClassStatement(p)
		if stmt != nil {
			class.Body = append(class.Body, stmt)
		}
		p.nextToken()
	}

	return ast.NewExpressionStatement(pos, class)
}

// parseReadonlyClassDeclaration 解析 readonly class 声明语句
func parseReadonlyClassDeclaration(p *Parser) ast.Statement {
	pos := p.currentToken.Position

	// 跳过 'readonly' 关键字，移动到 'class'
	if !p.expectPeek(lexer.T_CLASS) {
		return nil
	}

	// 跳过 'class' 关键字
	if !p.expectPeek(lexer.T_STRING) {
		return nil
	}

	// 解析类名
	name := ast.NewIdentifierNode(p.currentToken.Position, p.currentToken.Value)
	class := ast.NewClassExpression(pos, name, nil, nil, false, true) // final = false, readOnly = true

	// 检查是否有 extends 子句
	if p.peekToken.Type == lexer.T_EXTENDS {
		p.nextToken() // 跳过当前token移动到extends
		p.nextToken() // 移动到类名
		className, err := parseClassName(p)
		if err != nil {
			p.errors = append(p.errors, err.Error())
			return nil
		}
		extends := ast.NewIdentifierNode(p.currentToken.Position, className)
		class.Extends = extends
	}

	// 检查是否有 implements 子句
	if p.peekToken.Type == lexer.T_IMPLEMENTS {
		p.nextToken() // 移动到 implements
		p.nextToken() // 移动到第一个接口名

		interfaceName, err := parseClassName(p)
		if err != nil {
			p.errors = append(p.errors, err.Error())
			return nil
		}
		interfaceNode := ast.NewIdentifierNode(p.currentToken.Position, interfaceName)
		class.Implements = append(class.Implements, interfaceNode)

		// 处理多个接口，用逗号分隔
		for p.peekToken.Type == lexer.TOKEN_COMMA {
			p.nextToken() // 移动到逗号
			p.nextToken() // 移动到下一个接口名
			interfaceName, err := parseClassName(p)
			if err != nil {
				p.errors = append(p.errors, err.Error())
				break
			}
			interfaceNode := ast.NewIdentifierNode(p.currentToken.Position, interfaceName)
			class.Implements = append(class.Implements, interfaceNode)
		}
	}

	// 期望左大括号开始类体
	if !p.expectPeek(lexer.TOKEN_LBRACE) {
		return nil
	}

	// 解析类体
	p.nextToken() // 进入类体
	for !p.currentTokenIs(lexer.TOKEN_RBRACE) && !p.isAtEnd() {
		stmt := parseClassStatement(p)
		if stmt != nil {
			class.Body = append(class.Body, stmt)
		}
		p.nextToken()
	}

	return ast.NewExpressionStatement(pos, class)
}

// parseFinalClassDeclaration 解析 final class 声明语句
func parseFinalClassDeclaration(p *Parser) ast.Statement {
	pos := p.currentToken.Position

	// 跳过 'final' 关键字，移动到 'class'
	if !p.expectPeek(lexer.T_CLASS) {
		return nil
	}

	// 跳过 'class' 关键字
	if !p.expectPeek(lexer.T_STRING) {
		return nil
	}

	// 解析类名
	name := ast.NewIdentifierNode(p.currentToken.Position, p.currentToken.Value)
	class := ast.NewClassExpression(pos, name, nil, nil, true, false) // final = true, readOnly = false

	// 检查是否有 extends 子句
	if p.peekToken.Type == lexer.T_EXTENDS {
		p.nextToken() // 跳过当前token移动到extends
		p.nextToken() // 移动到类名
		className, err := parseClassName(p)
		if err != nil {
			p.errors = append(p.errors, err.Error())
			return nil
		}
		extends := ast.NewIdentifierNode(p.currentToken.Position, className)
		class.Extends = extends
	}

	// 检查是否有 implements 子句
	if p.peekToken.Type == lexer.T_IMPLEMENTS {
		p.nextToken() // 移动到 implements
		p.nextToken() // 移动到第一个接口名

		interfaceName, err := parseClassName(p)
		if err != nil {
			p.errors = append(p.errors, err.Error())
			return nil
		}
		class.Implements = append(class.Implements, ast.NewIdentifierNode(p.currentToken.Position, interfaceName))

		// 处理多个接口
		for p.peekToken.Type == lexer.TOKEN_COMMA {
			p.nextToken() // 移动到逗号
			p.nextToken() // 移动到下一个接口名
			interfaceName, err := parseClassName(p)
			if err != nil {
				p.errors = append(p.errors, err.Error())
				return nil
			}
			class.Implements = append(class.Implements, ast.NewIdentifierNode(p.currentToken.Position, interfaceName))
		}
	}

	// 解析类主体
	if !p.expectPeek(lexer.TOKEN_LBRACE) {
		return nil
	}

	p.nextToken() // 移动到类主体第一个token
	for p.currentToken.Type != lexer.TOKEN_RBRACE && p.currentToken.Type != lexer.T_EOF {
		stmt := parseClassStatement(p)
		if stmt != nil {
			class.Body = append(class.Body, stmt)
		}
		p.nextToken()
	}

	return ast.NewExpressionStatement(pos, class)
}

// parseMatchExpression 解析 match 表达式 (PHP 8.0+)
func parseMatchExpression(p *Parser) ast.Expression {
	pos := p.currentToken.Position

	// 跳过 'match'，期望左括号
	if !p.expectPeek(lexer.TOKEN_LPAREN) {
		return nil
	}

	// 解析匹配主体表达式
	p.nextToken()
	subject := parseExpression(p, LOWEST)
	if subject == nil {
		return nil
	}

	// 期望右括号
	if !p.expectPeek(lexer.TOKEN_RPAREN) {
		return nil
	}

	// 期望左大括号开始 match arms
	if !p.expectPeek(lexer.TOKEN_LBRACE) {
		return nil
	}

	matchExpr := ast.NewMatchExpression(pos, subject)

	// 解析 match arms
	p.nextToken() // 进入大括号内
	for !p.currentTokenIs(lexer.TOKEN_RBRACE) && !p.isAtEnd() {
		arm := parseMatchArm(p)
		if arm != nil {
			matchExpr.Arms = append(matchExpr.Arms, arm)
		}

		// 期望逗号或右大括号
		if p.peekToken.Type == lexer.TOKEN_COMMA {
			p.nextToken() // 跳过逗号
			// 检查是否为trailing comma（逗号后直接是右大括号）
			if p.peekToken.Type == lexer.TOKEN_RBRACE {
				break
			}
			p.nextToken() // 移动到下一个 arm
		} else if p.peekToken.Type == lexer.TOKEN_RBRACE {
			break
		} else {
			break
		}
	}

	// 期望右大括号结束
	if !p.expectPeek(lexer.TOKEN_RBRACE) {
		return nil
	}

	return matchExpr
}

// parseMatchArm 解析 match 表达式的一个分支
func parseMatchArm(p *Parser) *ast.MatchArm {
	pos := p.currentToken.Position

	var conditions []ast.Expression
	isDefault := false

	// 检查是否为 default 分支
	if p.currentToken.Type == lexer.T_DEFAULT {
		isDefault = true
	} else {
		// 解析条件列表
		for {
			// 解析条件表达式，允许逻辑运算符（||、&&）作为条件的一部分
			// 使用 COALESCE 优先级，这样会包含 || 和 && 但停在 => 之前
			condition := parseExpression(p, COALESCE)
			if condition == nil {
				return nil
			}
			conditions = append(conditions, condition)

			// 检查下一个 token 是什么
			if p.peekToken.Type == lexer.TOKEN_COMMA {
				// 跳过逗号
				p.nextToken()
				// 检查是否为 trailing comma（逗号后跟 '=>'）
				if p.peekToken.Type == lexer.T_DOUBLE_ARROW {
					// 这是一个 trailing comma，停止解析条件
					break
				}
				// 有更多条件，移动到下一个条件
				p.nextToken()
			} else if p.peekToken.Type == lexer.T_DOUBLE_ARROW {
				// 找到 =>，停止解析条件
				break
			} else {
				// 意外的 token，停止解析
				break
			}
		}
	}

	// 期望 '=>'
	if !p.expectPeek(lexer.T_DOUBLE_ARROW) {
		return nil
	}

	// 解析分支体
	p.nextToken()
	body := parseExpression(p, LOWEST)
	if body == nil {
		return nil
	}

	return ast.NewMatchArm(pos, conditions, body, isDefault)
}

// parseClassExpression 解析类表达式（保持向后兼容）
func parseClassExpression(p *Parser) ast.Expression {
	pos := p.currentToken.Position

	// 检查是否是静态访问：class::method 的情况
	// 在这种情况下，"class" 应该被视为类名标识符，而不是类声明
	if p.peekToken.Type == lexer.T_PAAMAYIM_NEKUDOTAYIM {
		// 这是 class::something 的情况，将 "class" 作为标识符返回
		return ast.NewIdentifierNode(pos, "class")
	}

	// 类声明通常需要名称和可能的扩展
	p.nextToken()

	var name ast.Expression
	if p.currentToken.Type == lexer.T_STRING {
		name = ast.NewIdentifierNode(p.currentToken.Position, p.currentToken.Value)
	}

	return ast.NewClassExpression(pos, name, nil, nil, false, false) // final = false, readOnly = false
}

// parseVisibilityStaticFunction 解析 visibility static function 声明
// 在调用时，currentToken 是 static，peekToken 是 function
func parseVisibilityStaticFunction(p *Parser, visibility string, pos lexer.Position) ast.Statement {
	// 当前：static token，下一个：function token
	if !p.expectPeek(lexer.T_FUNCTION) {
		return nil
	}

	// 现在开始解析函数，类似 parseFunctionDeclaration 但已知是 visibility + static

	// 检查是否为引用返回函数 function &foo()
	byReference := false
	if p.peekToken.Type == lexer.TOKEN_AMPERSAND {
		byReference = true
		p.nextToken() // 移动到 &
	}

	// Expect method/function name (allow reserved keywords)
	p.nextToken()
	if p.currentToken.Type != lexer.T_STRING && !isSemiReserved(p.currentToken.Type) {
		p.errors = append(p.errors, fmt.Sprintf("expected function name, got %s at position %s",
			p.currentToken.Value, p.currentToken.Position))
		return nil
	}

	funcName := ast.NewIdentifierNode(p.currentToken.Position, p.currentToken.Value)

	// 创建函数声明，设置可见性和静态标志
	funcDecl := ast.NewFunctionDeclaration(pos, funcName)
	funcDecl.Visibility = visibility
	funcDecl.IsStatic = true
	funcDecl.ByReference = byReference

	// Expect '('
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
			
			// 检查逗号后面是否直接是结束符（支持尾随逗号）
			if p.peekToken.Type == lexer.TOKEN_RPAREN {
				break
			}
			
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

	// 检查是否有返回类型声明 ": type"
	if p.peekToken.Type == lexer.TOKEN_COLON {
		p.nextToken() // 移动到 ':'
		p.nextToken() // 移动到类型开始位置

		// 解析返回类型（支持复杂类型）
		returnType := parseTypeHint(p)
		if returnType != nil {
			funcDecl.ReturnType = returnType
		}
	}

	// 解析函数体
	if p.peekToken.Type == lexer.TOKEN_LBRACE {
		p.nextToken()
		funcDecl.Body = parseBlockStatements(p)
	} else if p.peekToken.Type == lexer.TOKEN_SEMICOLON {
		// 抽象方法
		p.nextToken()
	}

	return funcDecl
}

// parseClassStatement 解析类体内的语句（属性和方法）
// 基于 PHP 官方语法实现：先识别修饰符序列，再根据后续 token 确定声明类型
func parseClassStatement(p *Parser) ast.Statement {
	switch p.currentToken.Type {
	case lexer.T_PRIVATE, lexer.T_PROTECTED, lexer.T_PUBLIC:
		// 可见性修饰符后可能跟着：const, function, static, readonly, 或直接是属性
		if p.peekToken.Type == lexer.T_CONST {
			return parseClassConstantDeclaration(p)
		} else if p.peekToken.Type == lexer.T_FUNCTION {
			return parseFunctionDeclaration(p)
		} else if p.peekToken.Type == lexer.T_STATIC {
			// visibility static 组合：统一让 parsePropertyDeclaration 处理
			// parsePropertyDeclaration 内部会检查是否为函数并委派
			return parsePropertyDeclaration(p)
		} else if p.peekToken.Type == lexer.T_READONLY {
			// visibility readonly property
			return parsePropertyDeclaration(p)
		} else {
			// 默认为属性声明（可能有类型提示）
			return parsePropertyDeclaration(p)
		}
	case lexer.T_FUNCTION:
		return parseFunctionDeclaration(p)
	case lexer.T_CONST:
		// const without visibility modifier (defaults to public)
		return parseClassConstantDeclaration(p)
	case lexer.T_READONLY:
		// readonly without visibility modifier (defaults to public)
		return parsePropertyDeclaration(p)
	case lexer.T_USE:
		return parseUseTraitStatement(p)
	case lexer.T_STATIC:
		// Handle static function or static property
		if p.peekToken.Type == lexer.T_FUNCTION {
			return parseFunctionDeclaration(p)
		} else {
			// static property (e.g., static $var)
			return parsePropertyDeclaration(p)
		}
	case lexer.T_ABSTRACT:
		// Handle abstract methods: abstract [visibility] function name();
		return parseFunctionDeclaration(p)
	default:
		// 跳过未识别的token
		return nil
	}
}

// parseClassConstantDeclaration 解析类常量声明
func parseClassConstantDeclaration(p *Parser) ast.Statement {
	pos := p.currentToken.Position

	// Parse visibility modifier (if present)
	visibility := "" // Empty by default (implicitly public when not specified)
	if p.currentToken.Type == lexer.T_PRIVATE || p.currentToken.Type == lexer.T_PROTECTED || p.currentToken.Type == lexer.T_PUBLIC {
		visibility = p.currentToken.Value
		p.nextToken() // Move to 'const'
	}

	// Expect 'const' keyword
	if p.currentToken.Type != lexer.T_CONST {
		p.errors = append(p.errors, fmt.Sprintf("expected 'const', got %s at position %s", p.currentToken.Value, p.currentToken.Position))
		return nil
	}

	p.nextToken() // Move past 'const'

	// Check for optional type annotation (PHP 8.3+)
	var constType *ast.TypeHint
	if isTypeToken(p.currentToken.Type) && p.peekToken.Type != lexer.TOKEN_EQUAL {
		// Only parse as type if the next token is not '=', meaning current token is type, not constant name
		constType = parseTypeHint(p)
		if constType == nil {
			return nil
		}
		p.nextToken() // Move past the type to the constant name
	}

	// Parse constant list (can have multiple constants in one declaration)
	var constants []ast.ConstantDeclarator

	for {
		// Parse constant name (allow reserved keywords)
		if p.currentToken.Type != lexer.T_STRING && !isSemiReserved(p.currentToken.Type) {
			p.errors = append(p.errors, fmt.Sprintf("expected constant name, got %s at position %s", p.currentToken.Value, p.currentToken.Position))
			return nil
		}

		constPos := p.currentToken.Position
		name := ast.NewIdentifierNode(p.currentToken.Position, p.currentToken.Value)

		// Expect '='
		if !p.expectPeek(lexer.TOKEN_EQUAL) {
			return nil
		}

		p.nextToken() // Move to value
		value := parseExpression(p, LOWEST)

		// Create constant declarator
		constants = append(constants, *ast.NewConstantDeclarator(constPos, name, value))

		// Check for comma (multiple constants) using peekToken
		if p.peekToken.Type == lexer.TOKEN_COMMA {
			p.nextToken() // Move to comma
			p.nextToken() // Skip comma and continue
			continue
		}

		// Should end with semicolon
		if p.peekToken.Type == lexer.TOKEN_SEMICOLON {
			break
		}

		// If not comma or semicolon, we're done
		break
	}

	return ast.NewClassConstantDeclaration(pos, visibility, constType, constants)
}

// parsePropertyDeclaration 解析属性声明
func parsePropertyDeclaration(p *Parser) ast.Statement {
	pos := p.currentToken.Position

	var visibility string
	var static bool
	var readOnly bool

	// Parse modifiers in order: visibility, static, readonly
	if p.currentToken.Type == lexer.T_PRIVATE || p.currentToken.Type == lexer.T_PROTECTED || p.currentToken.Type == lexer.T_PUBLIC {
		visibility = p.currentToken.Value

		// Check for static modifier
		if p.peekToken.Type == lexer.T_STATIC {
			static = true
			p.nextToken() // Move to static

			// 检查 static 后面是否为 function
			if p.peekToken.Type == lexer.T_FUNCTION {
				// 这是静态方法：visibility static function
				// 从当前位置（static token）开始解析，直接调用专门的函数
				return parseVisibilityStaticFunction(p, visibility, pos)
			}
		}

		// Check for readonly modifier
		if p.peekToken.Type == lexer.T_READONLY {
			readOnly = true
			p.nextToken() // Move to readonly
		}
	} else if p.currentToken.Type == lexer.T_STATIC {
		// Static without visibility modifier (defaults to public)
		visibility = "public"
		static = true
	} else if p.currentToken.Type == lexer.T_READONLY {
		// Readonly without visibility modifier (defaults to public)
		visibility = "public"
		readOnly = true
	} else {
		// Default visibility
		visibility = "public"
	}

	// 检查下一个token是否为类型提示
	var typeHint *ast.TypeHint
	if p.peekToken.Type != lexer.T_VARIABLE {
		p.nextToken()
		// 这是一个类型提示 (包括可空类型 ?Type)
		if isTypeToken(p.currentToken.Type) || p.currentToken.Type == lexer.TOKEN_QUESTION {
			typeHint = parseTypeHint(p)
		}
	}

	// 移动到变量名
	if !p.expectPeek(lexer.T_VARIABLE) {
		return nil
	}

	// 解析属性名（去掉$）
	propertyName := p.currentToken.Value
	if strings.HasPrefix(propertyName, "$") {
		propertyName = propertyName[1:]
	}

	// 检查是否为property hooks (用 { 而不是 ; 或 =)
	if p.peekToken.Type == lexer.TOKEN_LBRACE {
		// 解析property hooks
		p.nextToken() // 移动到 {
		hooks := parsePropertyHookList(p)
		return ast.NewHookedPropertyDeclaration(pos, visibility, propertyName, static, readOnly, typeHint, hooks)
	}

	var defaultValue ast.Expression

	// 检查是否有默认值
	if p.peekToken.Type == lexer.TOKEN_EQUAL {
		p.nextToken() // 跳到 =
		p.nextToken() // 移动到值
		defaultValue = parseExpression(p, LOWEST)
	}

	// 期望分号
	if p.peekToken.Type == lexer.TOKEN_SEMICOLON {
		p.nextToken()
	}

	return ast.NewPropertyDeclaration(pos, visibility, propertyName, static, readOnly, typeHint, defaultValue)
}

// parsePropertyHookList 解析属性钩子列表
func parsePropertyHookList(p *Parser) []*ast.PropertyHook {
	var hooks []*ast.PropertyHook

	p.nextToken() // 移动到第一个hook或结束符

	// 解析hook列表，直到遇到 }
	for p.currentToken.Type != lexer.TOKEN_RBRACE && !p.isAtEnd() {
		if (p.currentToken.Type == lexer.T_STRING && (p.currentToken.Value == "get" || p.currentToken.Value == "set")) || p.currentToken.Type == lexer.TOKEN_AMPERSAND {
			hook := parsePropertyHook(p)
			if hook != nil {
				hooks = append(hooks, hook)
			}
		}

		// 检查是否有分号分隔符
		if p.currentToken.Type == lexer.TOKEN_SEMICOLON {
			p.nextToken()
		} else if p.currentToken.Type != lexer.TOKEN_RBRACE {
			// 期望分号或结束符
			p.nextToken()
		}
	}

	return hooks
}

// parsePropertyHook 解析单个属性钩子 (get 或 set)
func parsePropertyHook(p *Parser) *ast.PropertyHook {
	pos := p.currentToken.Position

	// 检查是否为引用钩子 (&get)
	byRef := false
	if p.currentToken.Type == lexer.TOKEN_AMPERSAND {
		byRef = true
		p.nextToken()
	}

	// 必须是 get 或 set
	if p.currentToken.Type != lexer.T_STRING || (p.currentToken.Value != "get" && p.currentToken.Value != "set") {
		p.errors = append(p.errors, fmt.Sprintf("expected 'get' or 'set', got %s at position %s",
			p.currentToken.Value, p.currentToken.Position))
		return nil
	}

	hookType := p.currentToken.Value

	// 对于set hook，检查是否有参数
	var parameter *ast.Parameter
	if hookType == "set" && p.peekToken.Type == lexer.TOKEN_LPAREN {
		p.nextToken() // 移动到 (
		parameter = parsePropertyHookParameter(p)
		if !p.expectPeek(lexer.TOKEN_RPAREN) {
			return nil
		}
	}

	p.nextToken() // 移动到下一个token

	// 检查是arrow syntax (=>) 还是 block syntax ({})
	if p.currentToken.Type == lexer.T_DOUBLE_ARROW {
		// Arrow syntax: get => expression;
		p.nextToken() // 移动到表达式
		body := parseExpression(p, LOWEST)
		return ast.NewPropertyHook(pos, hookType, parameter, body, nil, byRef)
	} else if p.currentToken.Type == lexer.TOKEN_LBRACE {
		// Block syntax: get { statements }
		statements := parseBlockStatement(p).Body
		return ast.NewPropertyHook(pos, hookType, parameter, nil, statements, byRef)
	} else {
		p.errors = append(p.errors, fmt.Sprintf("expected '=>' or '{' after hook, got %s at position %s",
			p.currentToken.Value, p.currentToken.Position))
		return nil
	}
}

// parsePropertyHookParameter 解析set hook的参数
func parsePropertyHookParameter(p *Parser) *ast.Parameter {
	p.nextToken() // 移动到参数

	// 可能有类型提示
	var typeHint *ast.TypeHint
	if isTypeToken(p.currentToken.Type) {
		typeHint = parseTypeHint(p)
		p.nextToken() // 移动到变量名
	}

	// 必须是变量
	if p.currentToken.Type != lexer.T_VARIABLE {
		p.errors = append(p.errors, fmt.Sprintf("expected variable in set hook parameter, got %s at position %s",
			p.currentToken.Value, p.currentToken.Position))
		return nil
	}

	paramName := p.currentToken.Value

	return &ast.Parameter{
		Name: paramName,
		Type: typeHint,
	}
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

// parseStaticAccess 解析静态访问表达式 Class::method 或 Class::$property （前缀版本）
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

// parseDollarBraceExpression 解析 ${...} 或 $$var 语法的变量表达式
func parseDollarBraceExpression(p *Parser) ast.Expression {
	pos := p.currentToken.Position

	// 当前是 $, 检查下一个 token
	if p.peekToken.Type == lexer.TOKEN_LBRACE {
		// ${...} 语法
		// 跳过 $ 和 {
		p.nextToken() // 移动到 {
		p.nextToken() // 移动到表达式

		// 解析内部表达式
		expr := parseExpression(p, LOWEST)

		// 期望 }
		if !p.expectPeek(lexer.TOKEN_RBRACE) {
			return nil
		}

		// 创建一个表示动态变量名的表达式
		// 使用 Variable 节点，但名称为特殊格式
		return &ast.Variable{
			BaseNode: ast.BaseNode{
				Kind:     ast.ASTVar,
				Position: pos,
				LineNo:   uint32(pos.Line),
			},
			Name: fmt.Sprintf("${%s}", expr.String()),
		}
	} else if p.peekToken.Type == lexer.T_VARIABLE {
		// $$var 语法 - variable variable
		p.nextToken() // 移动到 $var
		varName := p.currentToken.Value

		// 创建一个表示 variable variable 的表达式
		return &ast.Variable{
			BaseNode: ast.BaseNode{
				Kind:     ast.ASTVar,
				Position: pos,
				LineNo:   uint32(pos.Line),
			},
			Name: fmt.Sprintf("$%s", varName),
		}
	} else {
		// 不支持的语法
		p.errors = append(p.errors, fmt.Sprintf("unexpected token after $ at position %s", pos))
		return nil
	}
}

// parseStaticAccessExpression 解析静态访问表达式 Class::method 或 Class::$property （中缀版本）
func parseStaticAccessExpression(p *Parser, left ast.Expression) ast.Expression {
	pos := p.currentToken.Position

	// 跳过 ::
	p.nextToken()

	var property ast.Expression
	if p.currentToken.Type == lexer.T_CLASS {
		// Special handling for ::class magic constant
		property = ast.NewIdentifierNode(p.currentToken.Position, "class")
	} else if p.currentToken.Type == lexer.T_NEW {
		// Special handling for ::new as a method name (not the new keyword)
		property = ast.NewIdentifierNode(p.currentToken.Position, "new")
	} else if isSemiReserved(p.currentToken.Type) {
		// Handle reserved keywords as method/property names (e.g., ::for, ::if, ::return)
		// This matches PHP's grammar where reserved_non_modifiers can be used as member_name
		property = ast.NewIdentifierNode(p.currentToken.Position, p.currentToken.Value)
	} else if p.currentToken.Type == lexer.TOKEN_DOLLAR && p.peekToken.Type == lexer.TOKEN_LBRACE {
		// Handle complex variable property syntax: static::${$var}
		// Use the parseDollarBraceExpression function we just created
		property = parseDollarBraceExpression(p)
	} else if p.currentToken.Type == lexer.TOKEN_DOLLAR && p.peekToken.Type == lexer.T_VARIABLE {
		// Handle variable variable property: static::$$var
		property = parseDollarBraceExpression(p)
	} else if p.currentToken.Type != lexer.T_EOF {
		property = parseExpression(p, CALL)
	}

	// 创建一个静态访问表达式
	return ast.NewStaticAccessExpression(pos, left, property)
}

// parseQualifiedNameExpression 解析限定名表达式 (例如 Parse\Date)
func parseQualifiedNameExpression(p *Parser, left ast.Expression) ast.Expression {
	pos := p.currentToken.Position

	// 确保左边是一个标识符
	leftIdent, ok := left.(*ast.IdentifierNode)
	if !ok {
		p.errors = append(p.errors, fmt.Sprintf("expected identifier before namespace separator at position %s", pos))
		return left
	}

	// 跳过 \ (namespace separator)
	p.nextToken()

	// 继续构建完整的限定名
	nameStr := leftIdent.Name

	if p.currentToken.Type == lexer.T_STRING {
		nameStr += "\\" + p.currentToken.Value

		// 继续解析后续的命名空间部分
		for p.peekToken.Type == lexer.T_NS_SEPARATOR {
			p.nextToken() // 跳到 \
			p.nextToken() // 跳过 \
			if p.currentToken.Type == lexer.T_STRING {
				nameStr += "\\" + p.currentToken.Value
			}
		}
	}

	// 创建新的标识符表达式包含完整的限定名
	return ast.NewIdentifierNode(pos, nameStr)
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

// parseShellExecExpression 解析命令执行表达式 (反引号)
func parseShellExecExpression(p *Parser) ast.Expression {
	pos := p.currentToken.Position
	var parts []ast.Expression

	// 跳过开始的反引号
	p.nextToken()

	// 解析命令内容，直到遇到结束的反引号
	for p.currentToken.Type != lexer.TOKEN_BACKTICK && !p.isAtEnd() {
		switch p.currentToken.Type {
		case lexer.T_VARIABLE:
			// 变量插值
			variable := ast.NewVariable(p.currentToken.Position, p.currentToken.Value)
			parts = append(parts, variable)
		case lexer.T_ENCAPSED_AND_WHITESPACE:
			// 命令片段
			if p.currentToken.Value != "" {
				stringPart := ast.NewStringLiteral(p.currentToken.Position, p.currentToken.Value, p.currentToken.Value)
				parts = append(parts, stringPart)
			}
		case lexer.T_CURLY_OPEN:
			// {$expression} 形式的复杂表达式
			p.nextToken() // 跳过 {
			expr := parseExpression(p, LOWEST)
			if expr != nil {
				parts = append(parts, expr)
			}
			// 期待右花括号
			if p.currentToken.Type == lexer.TOKEN_RBRACE {
				// 已经在正确位置，不需要额外移动
			} else {
				p.errors = append(p.errors, "expected '}' after expression in shell execution")
			}
		default:
			// 其他内容，当作命令片段处理
			if p.currentToken.Value != "" {
				stringPart := ast.NewStringLiteral(p.currentToken.Position, p.currentToken.Value, p.currentToken.Value)
				parts = append(parts, stringPart)
			}
		}
		p.nextToken()
	}

	// 如果没有找到结束的反引号，报错
	if p.currentToken.Type != lexer.TOKEN_BACKTICK {
		p.errors = append(p.errors, "unterminated shell execution expression")
	}

	// 如果只有一个部分且是简单字符串，仍返回 ShellExecExpression
	return ast.NewShellExecExpression(pos, parts)
}

// parseVisibilityModifier 解析可见性修饰符 public/private/protected
func parseVisibilityModifier(p *Parser) ast.Expression {
	pos := p.currentToken.Position
	modifier := p.currentToken.Value

	// 可见性修饰符后面应该跟着属性或方法声明
	return ast.NewVisibilityModifierExpression(pos, modifier)
}

// parseIncludeOrEvalExpression 解析 include/require/eval 表达式
func parseIncludeOrEvalExpression(p *Parser) ast.Expression {
	pos := p.currentToken.Position
	tokenType := p.currentToken.Type

	p.nextToken() // 跳过 include/require/eval 关键字

	var expr ast.Node
	if tokenType == lexer.T_EVAL {
		// eval 需要括号
		if !p.expectPeek(lexer.TOKEN_LPAREN) {
			return nil
		}
		p.nextToken() // 跳过 '('
		expr = parseExpression(p, LOWEST)
		if !p.expectPeek(lexer.TOKEN_RPAREN) {
			return nil
		}
	} else {
		// include/require 后面跟表达式
		expr = parseExpression(p, LOWEST)
	}

	return ast.NewIncludeOrEvalExpression(pos, tokenType, expr)
}

// parseStaticExpression 解析 static 关键字表达式
func parseStaticExpression(p *Parser) ast.Expression {
	pos := p.currentToken.Position

	// 检查下一个token来决定如何解析
	switch p.peekToken.Type {
	case lexer.T_VARIABLE:
		// static $var = value; 这实际上是语句级别的，但在表达式上下文中先返回标识符
		// 实际的静态声明会在语句解析中处理
		name := ast.NewIdentifierNode(pos, "static")
		return name
	case lexer.T_FUNCTION:
		// static function() { ... } 静态匿名函数
		p.nextToken() // 跳过 static
		return parseAnonymousFunctionExpression(p)
	case lexer.T_PAAMAYIM_NEKUDOTAYIM:
		// static::method() 静态类引用
		name := ast.NewIdentifierNode(pos, "static")
		return name
	default:
		// 单独的 static 关键字，可能用于类型声明或其他用途
		name := ast.NewIdentifierNode(pos, "static")
		return name
	}
}

// parseAbstractExpression 解析 abstract 关键字
func parseAbstractExpression(p *Parser) ast.Expression {
	pos := p.currentToken.Position

	// abstract 通常用作类或方法修饰符
	// 在这里作为标识符返回，实际的类声明解析会在其他地方处理
	return ast.NewIdentifierNode(pos, "abstract")
}

// parseFinalExpression 解析 final 关键字
func parseFinalExpression(p *Parser) ast.Expression {
	pos := p.currentToken.Position

	// final 通常用作类修饰符，防止类被继承
	// 在这里作为标识符返回，实际的类声明解析会在其他地方处理
	return ast.NewIdentifierNode(pos, "final")
}

// parseYieldExpression 解析 yield 表达式
func parseYieldExpression(p *Parser) ast.Expression {
	pos := p.currentToken.Position

	p.nextToken() // 跳过 yield 关键字

	// yield 可能有三种形式：
	// 1. yield;               (无值)
	// 2. yield $value;        (只有值)
	// 3. yield $key => $value; (键值对)

	// 检查是否是空的 yield
	if p.currentToken.Type == lexer.TOKEN_SEMICOLON || p.currentToken.Type == lexer.T_EOF {
		return ast.NewYieldExpression(pos, nil, nil)
	}

	// 解析第一个表达式，使用较高的优先级以防止 => 被作为中缀操作符解析
	firstExpr := parseExpression(p, LOGICAL_OR)
	if firstExpr == nil {
		return ast.NewYieldExpression(pos, nil, nil)
	}

	// 检查是否有 =>，表示这是键值对形式
	if p.peekToken.Type == lexer.T_DOUBLE_ARROW {
		p.nextToken() // 跳过 =>
		p.nextToken() // 移动到值表达式
		valueExpr := parseExpression(p, LOWEST)
		return ast.NewYieldExpression(pos, firstExpr, valueExpr)
	}

	// 只有值的形式
	return ast.NewYieldExpression(pos, nil, firstExpr)
}

// parseYieldFromExpression 解析 yield from 表达式
func parseYieldFromExpression(p *Parser) ast.Expression {
	pos := p.currentToken.Position

	p.nextToken() // 跳过 "yield from" token

	// yield from 后面必须跟一个表达式
	if p.currentToken.Type == lexer.TOKEN_SEMICOLON || p.currentToken.Type == lexer.T_EOF {
		p.errors = append(p.errors, "yield from requires an expression")
		return nil
	}

	expr := parseExpression(p, LOWEST)
	if expr == nil {
		p.errors = append(p.errors, "invalid expression after yield from")
		return nil
	}

	return ast.NewYieldFromExpression(pos, expr)
}

// parseThrowExpression 解析 throw 表达式 (PHP 8.0+)
func parseThrowExpression(p *Parser) ast.Expression {
	pos := p.currentToken.Position

	p.nextToken() // 跳过 throw 关键字

	// throw 后面必须跟一个表达式
	if p.currentToken.Type == lexer.TOKEN_SEMICOLON || p.currentToken.Type == lexer.T_EOF {
		p.errors = append(p.errors, "throw requires an expression")
		return nil
	}

	// 解析异常表达式，使用较高的优先级以防止 ?: 等操作符干扰
	expr := parseExpression(p, PREFIX)
	if expr == nil {
		p.errors = append(p.errors, "invalid expression after throw")
		return nil
	}

	return ast.NewThrowExpression(pos, expr)
}

// parseCloseTagExpression 解析 PHP 结束标签
func parseCloseTagExpression(p *Parser) ast.Expression {
	pos := p.currentToken.Position
	content := p.currentToken.Value

	// 结束标签后可能有HTML内容
	return ast.NewCloseTagExpression(pos, content)
}

// parseNamespaceExpression 解析命名空间表达式（以 \ 开始）
func parseNamespaceExpression(p *Parser) ast.Expression {
	pos := p.currentToken.Position

	p.nextToken() // 跳过 \

	if p.currentToken.Type == lexer.T_STRING {
		// \Namespace\Class 形式
		nameStr := p.currentToken.Value

		// 继续解析后续的命名空间部分，构建完整的命名空间字符串
		for p.peekToken.Type == lexer.T_NS_SEPARATOR {
			p.nextToken() // 跳到 \
			p.nextToken() // 跳过 \
			if p.currentToken.Type == lexer.T_STRING {
				nameStr += "\\" + p.currentToken.Value
			}
		}

		// 创建完整的命名空间标识符
		name := ast.NewIdentifierNode(pos, nameStr)
		return ast.NewNamespaceExpression(pos, name)
	}

	// 单独的 \
	return ast.NewNamespaceExpression(pos, nil)
}

// parseQualifiedName 解析限定名称，支持命名空间（如 Foo\Bar\Baz）
func parseQualifiedName(p *Parser) ast.Expression {
	pos := p.currentToken.Position

	// 处理以 \ 开头的绝对命名空间
	if p.currentToken.Type == lexer.T_NS_SEPARATOR {
		return parseNamespaceExpression(p)
	}

	// 处理以标识符开头的相对命名空间
	if p.currentToken.Type == lexer.T_STRING {
		nameStr := p.currentToken.Value

		// 继续解析后续的命名空间部分
		for p.peekToken.Type == lexer.T_NS_SEPARATOR {
			p.nextToken() // 跳到 \
			p.nextToken() // 跳过 \
			if p.currentToken.Type == lexer.T_STRING {
				nameStr += "\\" + p.currentToken.Value
			} else {
				break
			}
		}

		// 创建标识符节点
		return ast.NewIdentifierNode(pos, nameStr)
	}

	// 回退到普通表达式解析
	return parseExpression(p, LESSGREATER)
}

// parseArrowFunctionExpression 解析箭头函数表达式 (PHP 7.4+)
func parseArrowFunctionExpression(p *Parser) ast.Expression {
	pos := p.currentToken.Position
	var static bool

	// 检查是否是静态箭头函数
	if p.currentToken.Type == lexer.T_STATIC {
		static = true
		if !p.expectPeek(lexer.T_FN) {
			return nil
		}
	}

	// 跳过 'fn'，期望左括号
	if !p.expectPeek(lexer.TOKEN_LPAREN) {
		return nil
	}

	// 解析参数列表
	var parameters []ast.Parameter

	// 解析参数列表
	if p.peekToken.Type != lexer.TOKEN_RPAREN {
		p.nextToken()

		// 解析第一个参数
		param := parseParameter(p)
		if param != nil {
			parameters = append(parameters, *param)
		}

		// 处理更多参数
		for p.peekToken.Type == lexer.TOKEN_COMMA {
			p.nextToken() // 移动到逗号
			
			// 检查逗号后面是否直接是结束符（支持尾随逗号）
			if p.peekToken.Type == lexer.TOKEN_RPAREN {
				break
			}
			
			p.nextToken() // 移动到下一个参数
			param := parseParameter(p)
			if param != nil {
				parameters = append(parameters, *param)
			}
		}
	}

	// 期望右括号
	if !p.expectPeek(lexer.TOKEN_RPAREN) {
		return nil
	}

	// 可选的返回类型
	var returnType *ast.TypeHint
	if p.peekToken.Type == lexer.TOKEN_COLON {
		p.nextToken() // 跳过冒号
		p.nextToken() // 进入返回类型
		returnType = parseTypeHint(p)
	}

	// 期望双箭头 =>
	if !p.expectPeek(lexer.T_DOUBLE_ARROW) {
		return nil
	}

	// 解析函数体表达式
	p.nextToken()
	body := parseExpression(p, LOWEST)
	if body == nil {
		return nil
	}

	return ast.NewArrowFunctionExpression(pos, parameters, returnType, body, static)
}

// parseStaticOrArrowFunctionExpression 处理 static 关键字（可能是静态箭头函数或其他静态表达式）
func parseStaticOrArrowFunctionExpression(p *Parser) ast.Expression {
	// 检查是否是静态箭头函数 static fn
	if p.peekToken.Type == lexer.T_FN {
		// 这是静态箭头函数，委托给箭头函数解析器
		return parseArrowFunctionExpression(p)
	}

	// 否则使用原来的静态表达式解析
	return parseStaticExpression(p)
}

// parseAnonymousClass 解析匿名类表达式 new class(args) extends Parent implements Interface { ... }
func parseAnonymousClass(p *Parser, pos lexer.Position) ast.Expression {
	// 跳过 'new' 关键字
	p.nextToken()

	// 解析属性 (如果存在)
	var attributes []*ast.AttributeGroup
	for p.currentToken.Type == lexer.T_ATTRIBUTE {
		attrGroup := parseAttributeGroup(p)
		if attrGroup != nil {
			attributes = append(attributes, attrGroup)
		}
		p.nextToken() // 移动到下一个token
	}

	// 解析类修饰符
	var modifiers []string
	for p.currentToken.Type == lexer.T_FINAL || p.currentToken.Type == lexer.T_ABSTRACT || p.currentToken.Type == lexer.T_READONLY {
		modifiers = append(modifiers, p.currentToken.Value)
		p.nextToken()
	}

	// 期望 'class' 关键字
	if p.currentToken.Type != lexer.T_CLASS {
		return nil
	}

	var arguments []ast.Expression

	// 检查这是否真的是匿名类还是尝试实例化"class"类
	// 匿名类的模式：
	// - new class { } (直接有类体)
	// - new class extends Base { }  
	// - new class implements Interface { }
	// - new class() { } (有构造函数参数)
	// 非匿名类的模式：
	// - new class() (没有类体，只有构造函数调用)
	if p.peekToken.Type == lexer.TOKEN_LPAREN {
		// 需要检查括号后面是否有类体
		// 保存当前位置
		savedPos := pos
		
		// 跳过构造函数参数来看后面是否有类体
		p.nextToken() // 移动到 (
		arguments = parseExpressionList(p, lexer.TOKEN_RPAREN)
		
		// 现在检查后面是否有 {, extends, implements
		hasClassBody := p.peekToken.Type == lexer.TOKEN_LBRACE ||
			p.peekToken.Type == lexer.T_EXTENDS ||
			p.peekToken.Type == lexer.T_IMPLEMENTS
		
		if !hasClassBody {
			// 这不是匿名类，而是尝试实例化"class"类
			// 创建一个普通的new表达式
			class := ast.NewIdentifierNode(p.currentToken.Position, "class")
			newExpr := ast.NewNewExpression(savedPos, class)
			newExpr.Arguments = arguments
			return newExpr
		}
		
		// 这是真正的匿名类，继续正常处理
		// arguments已经被解析了，不需要重新解析
	} else {
		// 没有构造函数参数，检查是否有类体
		if p.peekToken.Type != lexer.TOKEN_LBRACE &&
			p.peekToken.Type != lexer.T_EXTENDS &&
			p.peekToken.Type != lexer.T_IMPLEMENTS {
			// 这看起来不像匿名类，但也不是正确的语法
			// 返回nil让外层处理
			return nil
		}
		
		// 没有构造函数参数的匿名类
		arguments = []ast.Expression{}
	}

	// 检查 extends 子句
	var extends ast.Expression
	if p.peekToken.Type == lexer.T_EXTENDS {
		p.nextToken() // 移动到 extends
		p.nextToken() // 移动到类名
		extends = ast.NewIdentifierNode(p.currentToken.Position, p.currentToken.Value)
	}

	// 检查 implements 子句
	var implements []ast.Expression
	if p.peekToken.Type == lexer.T_IMPLEMENTS {
		p.nextToken() // 移动到 implements
		p.nextToken() // 移动到第一个接口名
		implements = append(implements, ast.NewIdentifierNode(p.currentToken.Position, p.currentToken.Value))

		// 处理多个接口
		for p.peekToken.Type == lexer.TOKEN_COMMA {
			p.nextToken() // 移动到逗号
			p.nextToken() // 移动到下一个接口名
			implements = append(implements, ast.NewIdentifierNode(p.currentToken.Position, p.currentToken.Value))
		}
	}

	// 期望类体开始
	if !p.expectPeek(lexer.TOKEN_LBRACE) {
		return nil
	}

	// 解析类体
	body := parseClassBody(p)

	return ast.NewAnonymousClass(pos, attributes, modifiers, arguments, extends, implements, body)
}

// parseClassBody 解析类体
func parseClassBody(p *Parser) []ast.Statement {
	var statements []ast.Statement

	p.nextToken() // 进入类体

	for !p.currentTokenIs(lexer.TOKEN_RBRACE) && !p.isAtEnd() {
		stmt := parseClassStatement(p)
		if stmt != nil {
			statements = append(statements, stmt)
		}
		p.nextToken()
	}

	return statements
}

// isTraitName 检查当前 token 是否可以作为 trait 名称
// trait 名称可以是简单名称或限定名称
func isTraitName(tokenType lexer.TokenType) bool {
	return tokenType == lexer.T_STRING ||
		tokenType == lexer.T_NAME_QUALIFIED ||
		tokenType == lexer.T_NAME_FULLY_QUALIFIED ||
		tokenType == lexer.T_NAME_RELATIVE
}

// parseUseTraitStatement 解析 trait 使用语句 (use TraitA, TraitB { ... })
func parseUseTraitStatement(p *Parser) ast.Statement {
	pos := p.currentToken.Position

	// 期望 trait 名称列表
	p.nextToken() // 跳过 'use'

	var traits []*ast.IdentifierNode

	// 解析第一个 trait 名称
	if !isTraitName(p.currentToken.Type) {
		p.errors = append(p.errors, fmt.Sprintf("expected trait name, got %s at position %s", p.currentToken.Value, p.currentToken.Position))
		return nil
	}

	traits = append(traits, ast.NewIdentifierNode(p.currentToken.Position, p.currentToken.Value))

	// 解析多个 trait（用逗号分隔）
	for p.peekToken.Type == lexer.TOKEN_COMMA {
		p.nextToken() // 跳过逗号
		p.nextToken() // 移动到下一个 trait 名称

		if !isTraitName(p.currentToken.Type) {
			p.errors = append(p.errors, fmt.Sprintf("expected trait name, got %s at position %s", p.currentToken.Value, p.currentToken.Position))
			return nil
		}

		traits = append(traits, ast.NewIdentifierNode(p.currentToken.Position, p.currentToken.Value))
	}

	// 检查是否有适配规则
	var adaptations []ast.TraitAdaptation

	if p.peekToken.Type == lexer.TOKEN_LBRACE {
		p.nextToken() // 跳过 '{'

		// 解析适配规则
		for p.peekToken.Type != lexer.TOKEN_RBRACE && p.peekToken.Type != lexer.T_EOF {
			p.nextToken()
			adaptation := parseTraitAdaptation(p)
			if adaptation != nil {
				adaptations = append(adaptations, adaptation)
			}

			// 期望分号
			if p.peekToken.Type == lexer.TOKEN_SEMICOLON {
				p.nextToken() // 跳过分号
			}
		}

		// 期望右大括号
		if !p.expectPeek(lexer.TOKEN_RBRACE) {
			return nil
		}
	} else if p.peekToken.Type == lexer.TOKEN_SEMICOLON {
		p.nextToken() // 跳过分号
	}

	return ast.NewUseTraitStatement(pos, traits, adaptations)
}

// parseTraitAdaptation 解析 trait 适配规则
func parseTraitAdaptation(p *Parser) ast.TraitAdaptation {
	// 解析 trait 方法引用
	methodRef := parseTraitMethodReference(p)
	if methodRef == nil {
		return nil
	}

	// 检查是 precedence 还是 alias
	if p.peekToken.Type == lexer.T_INSTEADOF {
		return parseTraitPrecedence(p, methodRef)
	} else if p.peekToken.Type == lexer.T_AS {
		return parseTraitAlias(p, methodRef)
	}

	p.errors = append(p.errors, fmt.Sprintf("expected 'insteadof' or 'as', got %s at position %s", p.peekToken.Value, p.peekToken.Position))
	return nil
}

// parseTraitMethodReference 解析 trait 方法引用 (TraitName::methodName or methodName)
func parseTraitMethodReference(p *Parser) *ast.TraitMethodReference {
	pos := p.currentToken.Position

	if p.currentToken.Type != lexer.T_STRING && !isSemiReserved(p.currentToken.Type) {
		p.errors = append(p.errors, fmt.Sprintf("expected method or trait name, got %s at position %s", p.currentToken.Value, p.currentToken.Position))
		return nil
	}

	firstPart := ast.NewIdentifierNode(p.currentToken.Position, p.currentToken.Value)

	// 检查是否是完全限定的方法引用 (TraitName::methodName)
	if p.peekToken.Type == lexer.T_PAAMAYIM_NEKUDOTAYIM {
		p.nextToken() // 跳过 '::'

		// 在trait适配上下文中，'as' 和 'insteadof' 应该被视为关键字，而不是方法名
		if p.peekToken.Type == lexer.T_AS || p.peekToken.Type == lexer.T_INSTEADOF {
			p.errors = append(p.errors, fmt.Sprintf("expected method name, got %s at position %s", p.peekToken.Value, p.peekToken.Position))
			return nil
		}

		// 检查下一个token是否是有效的方法名
		if p.peekToken.Type != lexer.T_STRING && !isSemiReserved(p.peekToken.Type) {
			p.errors = append(p.errors, fmt.Sprintf("expected method name, got %s at position %s", p.peekToken.Value, p.peekToken.Position))
			return nil
		}

		p.nextToken() // 移动到方法名
		methodName := ast.NewIdentifierNode(p.currentToken.Position, p.currentToken.Value)
		return ast.NewTraitMethodReference(pos, firstPart, methodName)
	}

	// 简单的方法名引用
	return ast.NewTraitMethodReference(pos, nil, firstPart)
}

// parseTraitPrecedence 解析 trait 优先级声明 (A::method insteadof B, C)
func parseTraitPrecedence(p *Parser, methodRef *ast.TraitMethodReference) ast.TraitAdaptation {
	pos := methodRef.Position

	// 期望 'insteadof'
	if !p.expectPeek(lexer.T_INSTEADOF) {
		return nil
	}

	// 解析替代的 trait 列表
	p.nextToken() // 移动到第一个 trait 名称

	var insteadOfTraits []*ast.IdentifierNode

	if !isTraitName(p.currentToken.Type) {
		p.errors = append(p.errors, fmt.Sprintf("expected trait name, got %s at position %s", p.currentToken.Value, p.currentToken.Position))
		return nil
	}

	insteadOfTraits = append(insteadOfTraits, ast.NewIdentifierNode(p.currentToken.Position, p.currentToken.Value))

	// 解析多个 trait（用逗号分隔）
	for p.peekToken.Type == lexer.TOKEN_COMMA {
		p.nextToken() // 跳过逗号
		p.nextToken() // 移动到下一个 trait 名称

		if !isTraitName(p.currentToken.Type) {
			p.errors = append(p.errors, fmt.Sprintf("expected trait name, got %s at position %s", p.currentToken.Value, p.currentToken.Position))
			return nil
		}

		insteadOfTraits = append(insteadOfTraits, ast.NewIdentifierNode(p.currentToken.Position, p.currentToken.Value))
	}

	return ast.NewTraitPrecedenceStatement(pos, methodRef, insteadOfTraits)
}

// parseTraitAlias 解析 trait 别名声明 (A::method as newName; A::method as public newName)
func parseTraitAlias(p *Parser, methodRef *ast.TraitMethodReference) ast.TraitAdaptation {
	pos := methodRef.Position

	// 期望 'as'
	if !p.expectPeek(lexer.T_AS) {
		return nil
	}

	p.nextToken() // 移动到别名或可见性修饰符

	var visibility string
	var alias *ast.IdentifierNode

	// 检查是否是可见性修饰符
	if p.currentToken.Type == lexer.T_PRIVATE || p.currentToken.Type == lexer.T_PROTECTED || p.currentToken.Type == lexer.T_PUBLIC {
		visibility = p.currentToken.Value

		// 可选的新方法名 (allow reserved keywords)
		if p.peekToken.Type == lexer.T_STRING || isReservedNonModifier(p.peekToken.Type) {
			p.nextToken() // 移动到方法名
			alias = ast.NewIdentifierNode(p.currentToken.Position, p.currentToken.Value)
		}
	} else if p.currentToken.Type == lexer.T_STRING || isReservedNonModifier(p.currentToken.Type) {
		// 只是别名，没有可见性修饰符 (allow reserved keywords)
		alias = ast.NewIdentifierNode(p.currentToken.Position, p.currentToken.Value)
	} else {
		p.errors = append(p.errors, fmt.Sprintf("expected visibility modifier or method name, got %s at position %s", p.currentToken.Value, p.currentToken.Position))
		return nil
	}

	return ast.NewTraitAliasStatement(pos, methodRef, alias, visibility)
}

// parseCurlyBraceExpression 解析花括号表达式 {expression}
// 根据 PHP 语法，member_name 可以是 '{' expr '}'，直接返回内部表达式
func parseCurlyBraceExpression(p *Parser) ast.Expression {
	// 当前 token 是 '{'
	// 跳过 '{'
	p.nextToken()
	
	// 解析内部表达式
	expr := parseExpression(p, LOWEST)
	
	// 期望 '}'
	if !p.expectPeek(lexer.TOKEN_RBRACE) {
		return nil
	}
	
	// 根据 PHP 语法，直接返回内部表达式，不需要包装节点
	return expr
}
