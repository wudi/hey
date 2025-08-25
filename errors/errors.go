package errors

import (
	"fmt"
	"strings"

	"github.com/wudi/php-parser/lexer"
)

// ErrorType 错误类型
type ErrorType int

const (
	SyntaxError ErrorType = iota
	LexicalError
	SemanticError
)

// Error 表示解析错误
type Error struct {
	Type     ErrorType      `json:"type"`
	Message  string         `json:"message"`
	Position lexer.Position `json:"position"`
	Source   string         `json:"source,omitempty"`
}

// NewSyntaxError 创建语法错误
func NewSyntaxError(message string, pos lexer.Position) *Error {
	return &Error{
		Type:     SyntaxError,
		Message:  message,
		Position: pos,
	}
}

// NewLexicalError 创建词法错误
func NewLexicalError(message string, pos lexer.Position) *Error {
	return &Error{
		Type:     LexicalError,
		Message:  message,
		Position: pos,
	}
}

// NewSemanticError 创建语义错误
func NewSemanticError(message string, pos lexer.Position) *Error {
	return &Error{
		Type:     SemanticError,
		Message:  message,
		Position: pos,
	}
}

// String 返回错误的字符串表示
func (e *Error) String() string {
	var typeStr string
	switch e.Type {
	case SyntaxError:
		typeStr = "Syntax Error"
	case LexicalError:
		typeStr = "Lexical Error"
	case SemanticError:
		typeStr = "Semantic Error"
	}

	return fmt.Sprintf("%s at line %d, column %d: %s",
		typeStr, e.Position.Line, e.Position.Column, e.Message)
}

// Error 实现 error 接口
func (e *Error) Error() string {
	return e.String()
}

// WithSource 添加源代码上下文
func (e *Error) WithSource(source string) *Error {
	e.Source = source
	return e
}

// ErrorList 错误列表
type ErrorList []*Error

// Add 添加错误
func (el *ErrorList) Add(err *Error) {
	*el = append(*el, err)
}

// AddSyntaxError 添加语法错误
func (el *ErrorList) AddSyntaxError(message string, pos lexer.Position) {
	el.Add(NewSyntaxError(message, pos))
}

// AddLexicalError 添加词法错误
func (el *ErrorList) AddLexicalError(message string, pos lexer.Position) {
	el.Add(NewLexicalError(message, pos))
}

// AddSemanticError 添加语义错误
func (el *ErrorList) AddSemanticError(message string, pos lexer.Position) {
	el.Add(NewSemanticError(message, pos))
}

// HasErrors 检查是否有错误
func (el ErrorList) HasErrors() bool {
	return len(el) > 0
}

// Count 返回错误数量
func (el ErrorList) Count() int {
	return len(el)
}

// String 返回所有错误的字符串表示
func (el ErrorList) String() string {
	var builder strings.Builder
	for i, err := range el {
		if i > 0 {
			builder.WriteString("\n")
		}
		builder.WriteString(err.String())
	}
	return builder.String()
}

// Error 实现 error 接口
func (el ErrorList) Error() string {
	return el.String()
}

// FilterByType 按类型过滤错误
func (el ErrorList) FilterByType(errorType ErrorType) ErrorList {
	var filtered ErrorList
	for _, err := range el {
		if err.Type == errorType {
			filtered = append(filtered, err)
		}
	}
	return filtered
}

// GetSyntaxErrors 获取语法错误
func (el ErrorList) GetSyntaxErrors() ErrorList {
	return el.FilterByType(SyntaxError)
}

// GetLexicalErrors 获取词法错误
func (el ErrorList) GetLexicalErrors() ErrorList {
	return el.FilterByType(LexicalError)
}

// GetSemanticErrors 获取语义错误
func (el ErrorList) GetSemanticErrors() ErrorList {
	return el.FilterByType(SemanticError)
}

// ErrorReporter 错误报告器
type ErrorReporter struct {
	errors ErrorList
	source string
}

// NewErrorReporter 创建新的错误报告器
func NewErrorReporter(source string) *ErrorReporter {
	return &ErrorReporter{
		errors: make(ErrorList, 0),
		source: source,
	}
}

// Report 报告错误
func (er *ErrorReporter) Report(err *Error) {
	if er.source != "" {
		err.WithSource(er.source)
	}
	er.errors.Add(err)
}

// ReportSyntaxError 报告语法错误
func (er *ErrorReporter) ReportSyntaxError(message string, pos lexer.Position) {
	er.Report(NewSyntaxError(message, pos))
}

// ReportLexicalError 报告词法错误
func (er *ErrorReporter) ReportLexicalError(message string, pos lexer.Position) {
	er.Report(NewLexicalError(message, pos))
}

// ReportSemanticError 报告语义错误
func (er *ErrorReporter) ReportSemanticError(message string, pos lexer.Position) {
	er.Report(NewSemanticError(message, pos))
}

// GetErrors 获取所有错误
func (er *ErrorReporter) GetErrors() ErrorList {
	return er.errors
}

// HasErrors 检查是否有错误
func (er *ErrorReporter) HasErrors() bool {
	return er.errors.HasErrors()
}

// Clear 清除所有错误
func (er *ErrorReporter) Clear() {
	er.errors = make(ErrorList, 0)
}

// GetErrorCount 获取错误数量
func (er *ErrorReporter) GetErrorCount() int {
	return er.errors.Count()
}

// PrintFormatted 格式化打印错误（带源代码上下文）
func (e *Error) PrintFormatted() string {
	if e.Source == "" {
		return e.String()
	}

	lines := strings.Split(e.Source, "\n")
	if e.Position.Line <= 0 || e.Position.Line > len(lines) {
		return e.String()
	}

	var builder strings.Builder
	builder.WriteString(e.String())
	builder.WriteString("\n")

	// 显示错误行
	errorLine := lines[e.Position.Line-1]
	builder.WriteString(fmt.Sprintf("  %d | %s\n", e.Position.Line, errorLine))

	// 显示错误位置指示器
	builder.WriteString("      | ")
	for i := 0; i < e.Position.Column; i++ {
		builder.WriteString(" ")
	}
	builder.WriteString("^\n")

	return builder.String()
}
