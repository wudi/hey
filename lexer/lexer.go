package lexer

import (
	"fmt"
	"strings"
)

// Lexer 词法分析器结构体
type Lexer struct {
	input         string        // 输入源代码
	position      int           // 当前位置（指向当前字符）
	readPosition  int           // 当前读取位置（指向当前字符之后的字符）
	ch            byte          // 当前字符
	line          int           // 当前行号
	column        int           // 当前列号
	
	// 状态管理
	state         LexerState    // 当前状态
	stateStack    *StateStack   // 状态栈
	
	// Heredoc/Nowdoc 支持
	heredocLabel  string        // 当前 Heredoc 标签
	heredocLabels []string      // Heredoc 标签栈
	
	// 错误处理
	errors        []string      // 错误列表
}

// New 创建新的词法分析器
func New(input string) *Lexer {
	l := &Lexer{
		input:         input,
		line:          1,
		column:        0, // 从 0 开始计数
		state:         ST_INITIAL,
		stateStack:    NewStateStack(),
		heredocLabels: make([]string, 0),
		errors:        make([]string, 0),
	}
	
	l.readChar() // 读取第一个字符
	return l
}

// readChar 读取下一个字符并前进指针
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0 // EOF
	} else {
		l.ch = l.input[l.readPosition]
	}
	
	// 更新位置信息
	if l.position < len(l.input) && l.input[l.position] == '\n' {
		l.line++
		l.column = 0
	} else {
		l.column++
	}
	
	l.position = l.readPosition
	l.readPosition++
}

// peekChar 查看下一个字符但不移动指针
func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

// peekCharN 查看第 n 个字符后的字符（0-based）
func (l *Lexer) peekCharN(n int) byte {
	pos := l.readPosition + n
	if pos >= len(l.input) {
		return 0
	}
	return l.input[pos]
}

// getCurrentPosition 获取当前位置（token开始位置）
func (l *Lexer) getCurrentPosition() Position {
	// 需要计算当前字符的位置
	line, column := 1, 0
	for i := 0; i < l.position && i < len(l.input); i++ {
		if l.input[i] == '\n' {
			line++
			column = 0
		} else {
			column++
		}
	}
	
	return Position{
		Line:   line,
		Column: column,
		Offset: l.position,
	}
}

// skipWhitespace 跳过空白字符
func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

// readIdentifier 读取标识符
func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

// readNumber 读取数字
func (l *Lexer) readNumber() (string, TokenType) {
	position := l.position
	tokenType := T_LNUMBER // 默认为整数
	
	// 处理十六进制
	if l.ch == '0' && (l.peekChar() == 'x' || l.peekChar() == 'X') {
		l.readChar() // 跳过 '0'
		l.readChar() // 跳过 'x'
		for isHexDigit(l.ch) {
			l.readChar()
		}
		return l.input[position:l.position], T_LNUMBER
	}
	
	// 处理八进制
	if l.ch == '0' && isDigit(l.peekChar()) {
		for isOctalDigit(l.ch) {
			l.readChar()
		}
		return l.input[position:l.position], T_LNUMBER
	}
	
	// 处理二进制
	if l.ch == '0' && (l.peekChar() == 'b' || l.peekChar() == 'B') {
		l.readChar() // 跳过 '0'
		l.readChar() // 跳过 'b'
		for isBinaryDigit(l.ch) {
			l.readChar()
		}
		return l.input[position:l.position], T_LNUMBER
	}
	
	// 处理十进制
	for isDigit(l.ch) {
		l.readChar()
	}
	
	// 检查是否为浮点数
	if l.ch == '.' && isDigit(l.peekChar()) {
		tokenType = T_DNUMBER
		l.readChar() // 跳过 '.'
		for isDigit(l.ch) {
			l.readChar()
		}
	}
	
	// 检查科学计数法
	if l.ch == 'e' || l.ch == 'E' {
		tokenType = T_DNUMBER
		l.readChar()
		if l.ch == '+' || l.ch == '-' {
			l.readChar()
		}
		for isDigit(l.ch) {
			l.readChar()
		}
	}
	
	return l.input[position:l.position], tokenType
}

// readString 读取字符串
func (l *Lexer) readString(delimiter byte) (string, error) {
	l.readChar() // 移动到引号后
	
	var result strings.Builder
	
	for l.ch != delimiter && l.ch != 0 {
		if l.ch == '\\' {
			// 处理转义字符
			l.readChar()
			switch l.ch {
			case 'n':
				result.WriteByte('\n')
			case 'r':
				result.WriteByte('\r')
			case 't':
				result.WriteByte('\t')
			case '\\':
				result.WriteByte('\\')
			case '\'':
				result.WriteByte('\'')
			case '"':
				result.WriteByte('"')
			case '$':
				result.WriteByte('$')
			default:
				result.WriteByte(l.ch)
			}
		} else {
			result.WriteByte(l.ch)
		}
		l.readChar()
	}
	
	if l.ch != delimiter {
		return "", fmt.Errorf("unterminated string at line %d, column %d", l.line, l.column)
	}
	
	l.readChar() // 跳过结束的引号
	return result.String(), nil
}

// readLineComment 读取单行注释
func (l *Lexer) readLineComment() string {
	position := l.position
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
	return l.input[position:l.position]
}

// readBlockComment 读取块注释
func (l *Lexer) readBlockComment() string {
	position := l.position
	
	for {
		if l.ch == 0 {
			break
		}
		if l.ch == '*' && l.peekChar() == '/' {
			l.readChar() // 跳过 *
			l.readChar() // 跳过 /
			break
		}
		l.readChar()
	}
	
	return l.input[position:l.position]
}

// NextToken 获取下一个 token
func (l *Lexer) NextToken() Token {
	// 根据当前状态处理
	switch l.state {
	case ST_INITIAL:
		return l.nextTokenInitial()
	case ST_IN_SCRIPTING:
		return l.nextTokenInScripting()
	case ST_DOUBLE_QUOTES:
		return l.nextTokenInDoubleQuotes()
	case ST_HEREDOC:
		return l.nextTokenInHeredoc()
	case ST_NOWDOC:
		return l.nextTokenInNowdoc()
	default:
		return l.nextTokenInScripting()
	}
}

// nextTokenInitial 在初始状态（HTML模式）获取下一个token
func (l *Lexer) nextTokenInitial() Token {
	var content strings.Builder
	pos := l.getCurrentPosition()
	
	// 查找 PHP 开始标签
	for l.ch != 0 {
		if l.ch == '<' {
			if l.peekChar() == '?' {
				// 检查是否是 <?php
				if l.peekCharN(1) == 'p' && l.peekCharN(2) == 'h' && l.peekCharN(3) == 'p' {
					// 如果有内容，先返回HTML内容
					if content.Len() > 0 {
						return Token{Type: T_INLINE_HTML, Value: content.String(), Position: pos}
					}
					
					// 读取 <?php
					result := ""
					for i := 0; i < 5; i++ {
						result += string(l.ch)
						l.readChar()
					}
					
					// 跳过可能的空白
					if l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
						result += string(l.ch)
						l.readChar()
					}
					
					l.state = ST_IN_SCRIPTING
					return Token{Type: T_OPEN_TAG, Value: result, Position: pos}
				} else if l.peekCharN(1) == '=' {
					// <?= 标签
					if content.Len() > 0 {
						return Token{Type: T_INLINE_HTML, Value: content.String(), Position: pos}
					}
					
					result := string(l.ch) + string(l.peekChar()) + string(l.peekCharN(1))
					l.readChar() // <
					l.readChar() // ?
					l.readChar() // =
					
					l.state = ST_IN_SCRIPTING
					return Token{Type: T_OPEN_TAG_WITH_ECHO, Value: result, Position: pos}
				}
			}
		}
		
		content.WriteByte(l.ch)
		l.readChar()
	}
	
	// 文件结束
	if content.Len() > 0 {
		return Token{Type: T_INLINE_HTML, Value: content.String(), Position: pos}
	}
	
	return Token{Type: T_EOF, Value: "", Position: l.getCurrentPosition()}
}

// nextTokenInScripting 在 PHP 脚本模式获取下一个token
func (l *Lexer) nextTokenInScripting() Token {
	l.skipWhitespace()
	
	pos := l.getCurrentPosition()
	
	switch l.ch {
	case 0:
		return Token{Type: T_EOF, Value: "", Position: pos}
		
	// 单字符 tokens
	case ';':
		l.readChar()
		return Token{Type: TOKEN_SEMICOLON, Value: ";", Position: pos}
	case ',':
		l.readChar()
		return Token{Type: TOKEN_COMMA, Value: ",", Position: pos}
	case '{':
		l.readChar()
		return Token{Type: TOKEN_LBRACE, Value: "{", Position: pos}
	case '}':
		l.readChar()
		return Token{Type: TOKEN_RBRACE, Value: "}", Position: pos}
	case '(':
		l.readChar()
		return Token{Type: TOKEN_LPAREN, Value: "(", Position: pos}
	case ')':
		l.readChar()
		return Token{Type: TOKEN_RPAREN, Value: ")", Position: pos}
	case '[':
		l.readChar()
		return Token{Type: TOKEN_LBRACKET, Value: "[", Position: pos}
	case ']':
		l.readChar()
		return Token{Type: TOKEN_RBRACKET, Value: "]", Position: pos}
	case '~':
		l.readChar()
		return Token{Type: TOKEN_TILDE, Value: "~", Position: pos}
	case '@':
		l.readChar()
		return Token{Type: TOKEN_AT, Value: "@", Position: pos}
		
	// 可能是多字符的操作符
	case '+':
		if l.peekChar() == '+' {
			l.readChar()
			l.readChar()
			return Token{Type: T_INC, Value: "++", Position: pos}
		} else if l.peekChar() == '=' {
			l.readChar()
			l.readChar()
			return Token{Type: T_PLUS_EQUAL, Value: "+=", Position: pos}
		}
		l.readChar()
		return Token{Type: TOKEN_PLUS, Value: "+", Position: pos}
		
	case '-':
		if l.peekChar() == '-' {
			l.readChar()
			l.readChar()
			return Token{Type: T_DEC, Value: "--", Position: pos}
		} else if l.peekChar() == '=' {
			l.readChar()
			l.readChar()
			return Token{Type: T_MINUS_EQUAL, Value: "-=", Position: pos}
		} else if l.peekChar() == '>' {
			l.readChar()
			l.readChar()
			return Token{Type: T_OBJECT_OPERATOR, Value: "->", Position: pos}
		}
		l.readChar()
		return Token{Type: TOKEN_MINUS, Value: "-", Position: pos}
		
	case '*':
		if l.peekChar() == '*' {
			l.readChar()
			l.readChar()
			if l.ch == '=' {
				l.readChar()
				return Token{Type: T_POW_EQUAL, Value: "**=", Position: pos}
			}
			return Token{Type: T_POW, Value: "**", Position: pos}
		} else if l.peekChar() == '=' {
			l.readChar()
			l.readChar()
			return Token{Type: T_MUL_EQUAL, Value: "*=", Position: pos}
		}
		l.readChar()
		return Token{Type: TOKEN_MULTIPLY, Value: "*", Position: pos}
		
	case '/':
		if l.peekChar() == '/' {
			// 单行注释
			comment := l.readLineComment()
			return Token{Type: T_COMMENT, Value: comment, Position: pos}
		} else if l.peekChar() == '*' {
			// 块注释 - 先检查是否为文档注释
			isDocComment := l.peekCharN(1) == '*' // 检查是否为 /**
			l.readChar() // 跳过 /
			comment := l.readBlockComment()
			fullComment := "/*" + comment
			
			if isDocComment {
				return Token{Type: T_DOC_COMMENT, Value: fullComment, Position: pos}
			}
			return Token{Type: T_COMMENT, Value: fullComment, Position: pos}
		} else if l.peekChar() == '=' {
			l.readChar()
			l.readChar()
			return Token{Type: T_DIV_EQUAL, Value: "/=", Position: pos}
		}
		l.readChar()
		return Token{Type: TOKEN_DIVIDE, Value: "/", Position: pos}
		
	case '%':
		if l.peekChar() == '=' {
			l.readChar()
			l.readChar()
			return Token{Type: T_MOD_EQUAL, Value: "%=", Position: pos}
		}
		l.readChar()
		return Token{Type: TOKEN_MODULO, Value: "%", Position: pos}
		
	case '=':
		if l.peekChar() == '=' {
			l.readChar()
			l.readChar()
			if l.ch == '=' {
				l.readChar()
				return Token{Type: T_IS_IDENTICAL, Value: "===", Position: pos}
			}
			return Token{Type: T_IS_EQUAL, Value: "==", Position: pos}
		} else if l.peekChar() == '>' {
			l.readChar()
			l.readChar()
			return Token{Type: T_DOUBLE_ARROW, Value: "=>", Position: pos}
		}
		l.readChar()
		return Token{Type: TOKEN_EQUAL, Value: "=", Position: pos}
		
	case '!':
		if l.peekChar() == '=' {
			l.readChar()
			l.readChar()
			if l.ch == '=' {
				l.readChar()
				return Token{Type: T_IS_NOT_IDENTICAL, Value: "!==", Position: pos}
			}
			return Token{Type: T_IS_NOT_EQUAL, Value: "!=", Position: pos}
		}
		l.readChar()
		return Token{Type: TOKEN_EXCLAMATION, Value: "!", Position: pos}
		
	case '<':
		if l.peekChar() == '=' {
			l.readChar()
			l.readChar()
			if l.ch == '>' {
				l.readChar()
				return Token{Type: T_SPACESHIP, Value: "<=>", Position: pos}
			}
			return Token{Type: T_IS_SMALLER_OR_EQUAL, Value: "<=", Position: pos}
		} else if l.peekChar() == '<' {
			l.readChar()
			l.readChar()
			if l.ch == '=' {
				l.readChar()
				return Token{Type: T_SL_EQUAL, Value: "<<=", Position: pos}
			}
			return Token{Type: T_SL, Value: "<<", Position: pos}
		}
		l.readChar()
		return Token{Type: TOKEN_LT, Value: "<", Position: pos}
		
	case '>':
		if l.peekChar() == '=' {
			l.readChar()
			l.readChar()
			return Token{Type: T_IS_GREATER_OR_EQUAL, Value: ">=", Position: pos}
		} else if l.peekChar() == '>' {
			l.readChar()
			l.readChar()
			if l.ch == '=' {
				l.readChar()
				return Token{Type: T_SR_EQUAL, Value: ">>=", Position: pos}
			}
			return Token{Type: T_SR, Value: ">>", Position: pos}
		}
		l.readChar()
		return Token{Type: TOKEN_GT, Value: ">", Position: pos}
		
	case '&':
		if l.peekChar() == '&' {
			l.readChar()
			l.readChar()
			return Token{Type: T_BOOLEAN_AND, Value: "&&", Position: pos}
		} else if l.peekChar() == '=' {
			l.readChar()
			l.readChar()
			return Token{Type: T_AND_EQUAL, Value: "&=", Position: pos}
		}
		l.readChar()
		return Token{Type: TOKEN_AMPERSAND, Value: "&", Position: pos}
		
	case '|':
		if l.peekChar() == '|' {
			l.readChar()
			l.readChar()
			return Token{Type: T_BOOLEAN_OR, Value: "||", Position: pos}
		} else if l.peekChar() == '=' {
			l.readChar()
			l.readChar()
			return Token{Type: T_OR_EQUAL, Value: "|=", Position: pos}
		}
		l.readChar()
		return Token{Type: TOKEN_PIPE, Value: "|", Position: pos}
		
	case '^':
		if l.peekChar() == '=' {
			l.readChar()
			l.readChar()
			return Token{Type: T_XOR_EQUAL, Value: "^=", Position: pos}
		}
		l.readChar()
		return Token{Type: TOKEN_CARET, Value: "^", Position: pos}
		
	case '.':
		if l.peekChar() == '=' {
			l.readChar()
			l.readChar()
			return Token{Type: T_CONCAT_EQUAL, Value: ".=", Position: pos}
		} else if l.peekChar() == '.' && l.peekCharN(3) == '.' {
			l.readChar()
			l.readChar()
			l.readChar()
			return Token{Type: T_ELLIPSIS, Value: "...", Position: pos}
		} else if isDigit(l.peekChar()) {
			// 浮点数
			number, tokenType := l.readNumber()
			return Token{Type: tokenType, Value: number, Position: pos}
		}
		l.readChar()
		return Token{Type: TOKEN_DOT, Value: ".", Position: pos}
		
	case '?':
		if l.peekChar() == '?' {
			l.readChar()
			l.readChar()
			if l.ch == '=' {
				l.readChar()
				return Token{Type: T_COALESCE_EQUAL, Value: "??=", Position: pos}
			}
			return Token{Type: T_COALESCE, Value: "??", Position: pos}
		} else if l.peekChar() == '-' && l.peekCharN(3) == '>' {
			l.readChar()
			l.readChar()
			l.readChar()
			return Token{Type: T_NULLSAFE_OBJECT_OPERATOR, Value: "?->", Position: pos}
		} else if l.peekChar() == '>' {
			// PHP 结束标签
			l.readChar()
			l.readChar()
			l.state = ST_INITIAL
			return Token{Type: T_CLOSE_TAG, Value: "?>", Position: pos}
		}
		l.readChar()
		return Token{Type: TOKEN_QUESTION, Value: "?", Position: pos}
		
	case ':':
		if l.peekChar() == ':' {
			l.readChar()
			l.readChar()
			return Token{Type: T_PAAMAYIM_NEKUDOTAYIM, Value: "::", Position: pos}
		}
		l.readChar()
		return Token{Type: TOKEN_COLON, Value: ":", Position: pos}
		
	case '$':
		if isLetter(l.peekChar()) || l.peekChar() == '_' {
			// 变量
			l.readChar() // 跳过 $
			identifier := l.readIdentifier()
			return Token{Type: T_VARIABLE, Value: "$" + identifier, Position: pos}
		}
		l.readChar()
		return Token{Type: TOKEN_DOLLAR, Value: "$", Position: pos}
		
	case '\\':
		l.readChar()
		return Token{Type: T_NS_SEPARATOR, Value: "\\", Position: pos}
		
	case '"':
		// 双引号字符串
		str, err := l.readString('"')
		if err != nil {
			l.addError(err.Error())
			return Token{Type: T_UNKNOWN, Value: "", Position: pos}
		}
		return Token{Type: T_CONSTANT_ENCAPSED_STRING, Value: `"` + str + `"`, Position: pos}
		
	case '\'':
		// 单引号字符串
		str, err := l.readString('\'')
		if err != nil {
			l.addError(err.Error())
			return Token{Type: T_UNKNOWN, Value: "", Position: pos}
		}
		return Token{Type: T_CONSTANT_ENCAPSED_STRING, Value: "'" + str + "'", Position: pos}
		
	case '#':
		// 单行注释 (# 形式)
		comment := l.readLineComment()
		return Token{Type: T_COMMENT, Value: comment, Position: pos}
		
	default:
		if isLetter(l.ch) || l.ch == '_' {
			// 标识符或关键字
			identifier := l.readIdentifier()
			
			// 检查是否为关键字
			if tokenType, isKeyword := IsKeyword(identifier); isKeyword {
				return Token{Type: tokenType, Value: identifier, Position: pos}
			}
			
			return Token{Type: T_STRING, Value: identifier, Position: pos}
		} else if isDigit(l.ch) {
			// 数字
			number, tokenType := l.readNumber()
			return Token{Type: tokenType, Value: number, Position: pos}
		} else {
			// 未知字符
			ch := l.ch
			l.readChar()
			l.addError(fmt.Sprintf("unexpected character '%c' at line %d, column %d", ch, pos.Line, pos.Column))
			return Token{Type: T_BAD_CHARACTER, Value: string(ch), Position: pos}
		}
	}
}

// nextTokenInDoubleQuotes 在双引号字符串中获取token
func (l *Lexer) nextTokenInDoubleQuotes() Token {
	// 这里简化实现，实际需要处理变量插值等
	return l.nextTokenInScripting()
}

// nextTokenInHeredoc 在Heredoc中获取token
func (l *Lexer) nextTokenInHeredoc() Token {
	// 这里简化实现，实际需要处理Heredoc语法
	return l.nextTokenInScripting()
}

// nextTokenInNowdoc 在Nowdoc中获取token
func (l *Lexer) nextTokenInNowdoc() Token {
	// 这里简化实现，实际需要处理Nowdoc语法
	return l.nextTokenInScripting()
}

// addError 添加错误信息
func (l *Lexer) addError(msg string) {
	l.errors = append(l.errors, msg)
}

// GetErrors 获取所有错误
func (l *Lexer) GetErrors() []string {
	return l.errors
}

// State 获取当前状态
func (l *Lexer) State() LexerState {
	return l.state
}

// 辅助函数

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func isHexDigit(ch byte) bool {
	return isDigit(ch) || ('a' <= ch && ch <= 'f') || ('A' <= ch && ch <= 'F')
}

func isOctalDigit(ch byte) bool {
	return '0' <= ch && ch <= '7'
}

func isBinaryDigit(ch byte) bool {
	return ch == '0' || ch == '1'
}