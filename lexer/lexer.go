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
	
	// 跳过 shebang 行（如 #!/usr/bin/php）
	l.skipShebang()
	
	l.readChar() // 读取第一个字符
	return l
}

// skipShebang 跳过文件开头的 shebang 行
func (l *Lexer) skipShebang() {
	// 检查是否以 #! 开头
	if len(l.input) >= 2 && l.input[0] == '#' && l.input[1] == '!' {
		// 找到第一个换行符，跳过整行
		i := 0
		for i < len(l.input) && l.input[i] != '\n' && l.input[i] != '\r' {
			i++
		}
		
		// 处理不同的行尾格式
		if i < len(l.input) {
			if l.input[i] == '\r' {
				i++ // 跳过 \r
				// 检查是否有 \n 跟在 \r 后面 (CRLF)
				if i < len(l.input) && l.input[i] == '\n' {
					i++ // 跳过 \n
				}
			} else if l.input[i] == '\n' {
				i++ // 跳过 \n (LF)
			}
		}
		
		// 更新输入，从 shebang 行之后开始
		if i > 0 && i < len(l.input) {
			l.input = l.input[i:]
		} else if i >= len(l.input) {
			// 整个文件都是 shebang 行
			l.input = ""
		}
	}
}

// readChar 读取下一个字符并前进指针
func (l *Lexer) readChar() {
	// 先更新位置指针
	l.position = l.readPosition
	l.readPosition++
	
	// 根据当前读取的字符更新行列信息
	if l.position >= len(l.input) {
		l.ch = 0 // EOF
		return
	}
	
	l.ch = l.input[l.position]
	
	// 更新行列信息：基于当前字符位置
	if l.position == 0 {
		// 第一个字符
		l.line = 1
		l.column = 0
	} else {
		// 检查前一个字符是否是换行符
		prevChar := l.input[l.position-1]
		if prevChar == '\n' {
			l.line++
			l.column = 0
		} else {
			l.column++
		}
	}
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
	// 直接使用已维护的行列信息，避免重复遍历
	return Position{
		Line:   l.line,
		Column: l.column,
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
		// 检查是否需要返回到之前的状态（如从 {$var} 插值返回到 Heredoc）
		if !l.stateStack.IsEmpty() {
			l.state = l.stateStack.Pop()
		}
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
			isDocComment := l.peekChar() == '*' && l.peekCharN(1) == '*' // 检查是否为 /**
			l.readChar() // 跳过 /
			comment := l.readBlockComment()
			fullComment := "/" + comment
			
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
			if l.peekCharN(1) == '<' {
				// Heredoc/Nowdoc 检测 <<<
				return l.handleHeredocStart(pos)
			}
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
		// 双引号字符串 - 检查是否包含变量插值
		if l.containsInterpolation('"') {
			// 包含变量插值，切换到 ST_DOUBLE_QUOTES 状态
			l.readChar() // 跳过开头的引号
			l.state = ST_DOUBLE_QUOTES
			return Token{Type: '"', Value: "\"", Position: pos}
		} else {
			// 简单字符串，无插值
			str, err := l.readString('"')
			if err != nil {
				l.addError(err.Error())
				return Token{Type: T_UNKNOWN, Value: "", Position: pos}
			}
			return Token{Type: T_CONSTANT_ENCAPSED_STRING, Value: `"` + str + `"`, Position: pos}
		}
		
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
	pos := l.getCurrentPosition()
	
	// 检查是否到达字符串结尾
	if l.ch == '"' {
		l.readChar() // 跳过结束引号
		l.state = ST_IN_SCRIPTING
		return Token{Type: '"', Value: "\"", Position: pos}
	}
	
	if l.ch == 0 {
		l.addError("unterminated string")
		return Token{Type: T_EOF, Value: "", Position: pos}
	}
	
	var content strings.Builder
	
	for l.ch != '"' && l.ch != 0 {
		// 检查 {$variable} 语法
		if l.ch == '{' && l.peekChar() == '$' {
			// 如果已经有内容，先返回内容
			if content.Len() > 0 {
				return Token{Type: T_ENCAPSED_AND_WHITESPACE, Value: content.String(), Position: pos}
			}
			// 推入当前状态到栈
			l.stateStack.Push(l.state)
			l.state = ST_IN_SCRIPTING
			l.readChar() // 跳过 {
			return Token{Type: T_CURLY_OPEN, Value: "{", Position: pos}
		} else if l.ch == '$' && (isLetter(l.peekChar()) || l.peekChar() == '_') {
			// 直接变量插值 $variable
			if content.Len() > 0 {
				return Token{Type: T_ENCAPSED_AND_WHITESPACE, Value: content.String(), Position: pos}
			}
			// 读取变量
			l.readChar() // 跳过 $
			identifier := l.readIdentifier()
			return Token{Type: T_VARIABLE, Value: "$" + identifier, Position: pos}
		}
		
		// 处理转义字符
		if l.ch == '\\' {
			l.readChar() // 跳过反斜杠
			if l.ch != 0 {
				switch l.ch {
				case 'n':
					content.WriteByte('\n')
				case 'r':
					content.WriteByte('\r')
				case 't':
					content.WriteByte('\t')
				case '\\':
					content.WriteByte('\\')
				case '"':
					content.WriteByte('"')
				case '$':
					content.WriteByte('$')
				default:
					content.WriteByte(l.ch)
				}
				l.readChar()
			}
		} else {
			content.WriteByte(l.ch)
			l.readChar()
		}
	}
		
	if content.Len() > 0 {
		return Token{Type: T_ENCAPSED_AND_WHITESPACE, Value: content.String(), Position: pos}
	}
	
	return Token{Type: T_EOF, Value: "", Position: pos}
}

// handleHeredocStart 处理 Heredoc/Nowdoc 开始标记
func (l *Lexer) handleHeredocStart(pos Position) Token {
	l.readChar() // 跳过第一个 <
	l.readChar() // 跳过第二个 <
	l.readChar() // 跳过第三个 <
	
	// 跳过空白
	for l.ch == ' ' || l.ch == '\t' {
		l.readChar()
	}
	
	isNowdoc := false
	var label string
	
	// 检查是否为 Nowdoc (<<<'LABEL')
	if l.ch == '\'' {
		isNowdoc = true
		l.readChar() // 跳过 '
		label = l.readHeredocLabel()
		if l.ch == '\'' {
			l.readChar() // 跳过结束的 '
		}
	} else if l.ch == '"' {
		// 支持 <<<"LABEL" 格式 (等同于 <<<LABEL)
		l.readChar() // 跳过 "
		label = l.readHeredocLabel()
		if l.ch == '"' {
			l.readChar() // 跳过结束的 "
		}
	} else {
		// 普通 Heredoc <<<LABEL
		label = l.readHeredocLabel()
	}
	
	if label == "" {
		l.addError("invalid heredoc/nowdoc label")
		return Token{Type: T_UNKNOWN, Value: "<<<", Position: pos}
	}
	
	// 跳过到行尾，并记录换行符用于token值
	var lineEnding string
	for l.ch != '\n' && l.ch != '\r' && l.ch != 0 {
		l.readChar()
	}
	if l.ch == '\r' {
		lineEnding += string(l.ch)
		l.readChar()
	}
	if l.ch == '\n' {
		lineEnding += string(l.ch)
		l.readChar()
	}
	
	// 保存标签并切换状态
	l.heredocLabel = label
	if isNowdoc {
		l.state = ST_NOWDOC
		return Token{Type: T_NOWDOC, Value: "<<<'" + label + "'" + lineEnding, Position: pos}
	} else {
		l.state = ST_HEREDOC
		return Token{Type: T_START_HEREDOC, Value: "<<<" + label + lineEnding, Position: pos}
	}
}

// readHeredocLabel 读取 Heredoc/Nowdoc 标签
func (l *Lexer) readHeredocLabel() string {
	var label strings.Builder
	
	// 第一个字符必须是字母或下划线
	if !isLetter(l.ch) {
		return ""
	}
	
	for isLetter(l.ch) || isDigit(l.ch) {
		label.WriteByte(l.ch)
		l.readChar()
	}
	
	return label.String()
}

// nextTokenInHeredoc 在Heredoc中获取token
func (l *Lexer) nextTokenInHeredoc() Token {
	pos := l.getCurrentPosition()
	
	// 检查是否到达结束标签
	if l.isAtHeredocEnd() {
		// 计算缩进长度
		indentStart := l.position
		for indentStart > 0 && l.input[indentStart-1] != '\n' && l.input[indentStart-1] != '\r' {
			indentStart--
		}
		
		// 读取结束标签（包含缩进）
		endTokenValue := l.input[indentStart:l.position+len(l.heredocLabel)]
		
		// 移动到标签结束位置
		for i := 0; i < len(l.heredocLabel); i++ {
			l.readChar()
		}
		
		l.heredocLabel = ""
		l.state = ST_IN_SCRIPTING
		return Token{Type: T_END_HEREDOC, Value: endTokenValue, Position: pos}
	}
	
	// 读取 Heredoc 内容
	var content strings.Builder
	for !l.isAtHeredocEnd() && l.ch != 0 {
		if l.ch == '{' && l.peekChar() == '$' {
			// {$variable} 模式 - 返回 T_CURLY_OPEN，切换到脚本状态处理变量
			if content.Len() > 0 {
				return Token{Type: T_ENCAPSED_AND_WHITESPACE, Value: content.String(), Position: pos}
			}
			l.stateStack.Push(l.state) // 保存当前 Heredoc 状态
			l.state = ST_IN_SCRIPTING  // 切换到脚本状态
			l.readChar()               // 跳过 {
			return Token{Type: T_CURLY_OPEN, Value: "{", Position: pos}
		} else if l.ch == '$' && (isLetter(l.peekChar()) || l.peekChar() == '_') {
			// 直接的变量插值 $variable
			if content.Len() > 0 {
				return Token{Type: T_ENCAPSED_AND_WHITESPACE, Value: content.String(), Position: pos}
			}
			// 读取变量
			l.readChar() // 跳过 $
			identifier := l.readIdentifier()
			return Token{Type: T_VARIABLE, Value: "$" + identifier, Position: pos}
		}
		content.WriteByte(l.ch)
		l.readChar()
	}
	
	if content.Len() > 0 {
		return Token{Type: T_ENCAPSED_AND_WHITESPACE, Value: content.String(), Position: pos}
	}
	
	return Token{Type: T_EOF, Value: "", Position: pos}
}

// nextTokenInNowdoc 在Nowdoc中获取token
func (l *Lexer) nextTokenInNowdoc() Token {
	pos := l.getCurrentPosition()
	
	// 检查是否到达结束标签
	if l.isAtHeredocEnd() {
		// 计算缩进长度
		indentStart := l.position
		for indentStart > 0 && l.input[indentStart-1] != '\n' && l.input[indentStart-1] != '\r' {
			indentStart--
		}
		
		// 读取结束标签（包含缩进）
		endTokenValue := l.input[indentStart:l.position+len(l.heredocLabel)]
		
		// 移动到标签结束位置
		for i := 0; i < len(l.heredocLabel); i++ {
			l.readChar()
		}
		
		l.heredocLabel = ""
		l.state = ST_IN_SCRIPTING
		return Token{Type: T_END_HEREDOC, Value: endTokenValue, Position: pos}
	}
	
	// 读取 Nowdoc 内容（无变量插值）
	var content strings.Builder
	for !l.isAtHeredocEnd() && l.ch != 0 {
		content.WriteByte(l.ch)
		l.readChar()
	}
	
	if content.Len() > 0 {
		return Token{Type: T_ENCAPSED_AND_WHITESPACE, Value: content.String(), Position: pos}
	}
	
	return Token{Type: T_EOF, Value: "", Position: pos}
}

// isAtHeredocEnd 检查当前位置是否到达 Heredoc/Nowdoc 结束标签
func (l *Lexer) isAtHeredocEnd() bool {
	if l.heredocLabel == "" {
		return false
	}
	
	// 检查当前位置是否在行首（允许缩进）
	if l.column != 0 {
		// 如果不在第0列，检查是否在行首的缩进位置
		// 向前查找直到行首，确保只有空格或制表符
		pos := l.position - 1
		for pos >= 0 && l.input[pos] != '\n' && l.input[pos] != '\r' {
			if l.input[pos] != ' ' && l.input[pos] != '\t' {
				return false // 不是纯缩进
			}
			pos--
		}
		// 如果到达这里，说明从行首到当前位置都是缩进字符
	}
	
	labelLen := len(l.heredocLabel)
	if l.position+labelLen > len(l.input) {
		return false
	}
	
	// 检查是否匹配标签
	candidateLabel := l.input[l.position:l.position+labelLen]
	if candidateLabel != l.heredocLabel {
		return false
	}
	
	// 检查标签后面的字符是否不是标签的延续（参考PHP的IS_LABEL_SUCCESSOR逻辑）
	nextPos := l.position + labelLen
	if nextPos >= len(l.input) {
		return true // 文件结尾
	}
	
	nextChar := l.input[nextPos]
	// 如果下一个字符不是字母、数字、下划线，则是有效的结束标记
	// 这与 PHP 的 !IS_LABEL_SUCCESSOR() 检查一致
	isLabelSuccessor := (nextChar >= 'a' && nextChar <= 'z') || 
	                   (nextChar >= 'A' && nextChar <= 'Z') || 
	                   (nextChar >= '0' && nextChar <= '9') || 
	                   nextChar == '_'
	return !isLabelSuccessor
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

// containsInterpolation 检查字符串是否包含变量插值
func (l *Lexer) containsInterpolation(delimiter byte) bool {
	pos := l.position + 1 // 跳过开头的引号
	
	for pos < len(l.input) && l.input[pos] != delimiter {
		if l.input[pos] == '\\' {
			// 跳过转义字符
			pos += 2
			continue
		}
		
		// 检查变量插值
		if l.input[pos] == '$' && pos+1 < len(l.input) {
			nextChar := l.input[pos+1]
			if isLetter(nextChar) || nextChar == '{' {
				return true
			}
		}
		
		// 检查花括号插值 {$var}
		if l.input[pos] == '{' && pos+1 < len(l.input) && l.input[pos+1] == '$' {
			return true
		}
		
		pos++
	}
	
	return false
}
