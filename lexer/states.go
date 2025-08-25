package lexer

// LexerState 表示词法分析器的状态
type LexerState int

// PHP Lexer 状态枚举，基于 PHP 官方实现
const (
	// 初始状态 - 解析 HTML 内容
	ST_INITIAL LexerState = iota

	// 脚本状态 - 解析 PHP 代码
	ST_IN_SCRIPTING

	// 双引号字符串状态
	ST_DOUBLE_QUOTES

	// Heredoc 状态
	ST_HEREDOC

	// Nowdoc 状态  
	ST_NOWDOC

	// 字符串中的变量偏移状态 (如 "$arr[index]" 中的 index)
	ST_VAR_OFFSET

	// 查找对象属性状态 (如 "->property")
	ST_LOOKING_FOR_PROPERTY

	// 查找变量名状态 (如 "${var_name}")
	ST_LOOKING_FOR_VARNAME

	// 反引号命令执行状态
	ST_BACKQUOTE

	// 注释状态
	ST_COMMENT

	// 文档注释状态
	ST_DOC_COMMENT
)

// StateNames 提供状态到名称的映射，便于调试
var StateNames = map[LexerState]string{
	ST_INITIAL:                "ST_INITIAL",
	ST_IN_SCRIPTING:           "ST_IN_SCRIPTING", 
	ST_DOUBLE_QUOTES:          "ST_DOUBLE_QUOTES",
	ST_HEREDOC:                "ST_HEREDOC",
	ST_NOWDOC:                 "ST_NOWDOC",
	ST_VAR_OFFSET:             "ST_VAR_OFFSET",
	ST_LOOKING_FOR_PROPERTY:   "ST_LOOKING_FOR_PROPERTY",
	ST_LOOKING_FOR_VARNAME:    "ST_LOOKING_FOR_VARNAME",
	ST_BACKQUOTE:              "ST_BACKQUOTE",
	ST_COMMENT:                "ST_COMMENT",
	ST_DOC_COMMENT:            "ST_DOC_COMMENT",
}

// String 返回状态的字符串表示
func (s LexerState) String() string {
	if name, exists := StateNames[s]; exists {
		return name
	}
	return "UNKNOWN_STATE"
}

// StateStack 状态栈，用于嵌套状态管理
type StateStack struct {
	states []LexerState
}

// NewStateStack 创建新的状态栈
func NewStateStack() *StateStack {
	return &StateStack{
		states: make([]LexerState, 0, 8), // 预分配容量
	}
}

// Push 压入新状态
func (s *StateStack) Push(state LexerState) {
	s.states = append(s.states, state)
}

// Pop 弹出栈顶状态
func (s *StateStack) Pop() LexerState {
	if len(s.states) == 0 {
		return ST_INITIAL // 默认状态
	}
	
	last := len(s.states) - 1
	state := s.states[last]
	s.states = s.states[:last]
	return state
}

// Peek 查看栈顶状态而不弹出
func (s *StateStack) Peek() LexerState {
	if len(s.states) == 0 {
		return ST_INITIAL
	}
	return s.states[len(s.states)-1]
}

// IsEmpty 检查栈是否为空
func (s *StateStack) IsEmpty() bool {
	return len(s.states) == 0
}

// Size 返回栈大小
func (s *StateStack) Size() int {
	return len(s.states)
}

// Clear 清空栈
func (s *StateStack) Clear() {
	s.states = s.states[:0]
}