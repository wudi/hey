package testutils

import (
	"fmt"
	"strings"
	"testing"

	"github.com/wudi/hey/compiler/lexer"
)

// PHPSourceBuilder PHP源码构建器
type PHPSourceBuilder struct {
	parts []string
}

// NewPHPSource 创建PHP源码构建器
func NewPHPSource() *PHPSourceBuilder {
	return &PHPSourceBuilder{
		parts: []string{"<?php"},
	}
}

// Add 添加PHP代码
func (b *PHPSourceBuilder) Add(code string) *PHPSourceBuilder {
	b.parts = append(b.parts, code)
	return b
}

// AddLine 添加一行PHP代码
func (b *PHPSourceBuilder) AddLine(code string) *PHPSourceBuilder {
	return b.Add(code + ";")
}

// AddVariable 添加变量赋值
func (b *PHPSourceBuilder) AddVariable(name, value string) *PHPSourceBuilder {
	return b.AddLine(fmt.Sprintf("%s = %s", name, value))
}

// AddEcho 添加echo语句
func (b *PHPSourceBuilder) AddEcho(args ...string) *PHPSourceBuilder {
	return b.AddLine("echo " + strings.Join(args, ", "))
}

// AddFunction 添加函数定义
func (b *PHPSourceBuilder) AddFunction(name string, params []string, body string) *PHPSourceBuilder {
	paramsStr := strings.Join(params, ", ")
	return b.Add(fmt.Sprintf("function %s(%s) { %s }", name, paramsStr, body))
}

// AddClass 添加类定义
func (b *PHPSourceBuilder) AddClass(name, body string) *PHPSourceBuilder {
	return b.Add(fmt.Sprintf("class %s { %s }", name, body))
}

// AddIf 添加if语句
func (b *PHPSourceBuilder) AddIf(condition, body string) *PHPSourceBuilder {
	return b.Add(fmt.Sprintf("if (%s) { %s }", condition, body))
}

// AddWhile 添加while语句
func (b *PHPSourceBuilder) AddWhile(condition, body string) *PHPSourceBuilder {
	return b.Add(fmt.Sprintf("while (%s) { %s }", condition, body))
}

// AddFor 添加for语句
func (b *PHPSourceBuilder) AddFor(init, condition, increment, body string) *PHPSourceBuilder {
	return b.Add(fmt.Sprintf("for (%s; %s; %s) { %s }", init, condition, increment, body))
}

// AddTryCatch 添加try-catch语句
func (b *PHPSourceBuilder) AddTryCatch(tryBody, catchVar, catchBody string) *PHPSourceBuilder {
	return b.Add(fmt.Sprintf("try { %s } catch (Exception %s) { %s }", tryBody, catchVar, catchBody))
}

// Build 构建最终的PHP源码
func (b *PHPSourceBuilder) Build() string {
	return strings.Join(b.parts, " ") + " ?>"
}

// String 实现Stringer接口
func (b *PHPSourceBuilder) String() string {
	return b.Build()
}

// QuickPHP 快速创建简单PHP源码的辅助函数
func QuickPHP(code string) string {
	return fmt.Sprintf("<?php %s ?>", code)
}

// QuickClass 快速创建类的PHP源码
func QuickClass(name, body string) string {
	return QuickPHP(fmt.Sprintf("class %s { %s }", name, body))
}

// QuickFunction 快速创建函数的PHP源码
func QuickFunction(name, params, body string) string {
	return QuickPHP(fmt.Sprintf("function %s(%s) { %s }", name, params, body))
}

// ParserTester 解析器测试器，用于快速测试
type ParserTester struct {
	parserFactory ParserFactory
	t             *testing.T
}

// NewParserTester 创建解析器测试器
func NewParserTester(t *testing.T, parserFactory ParserFactory) *ParserTester {
	return &ParserTester{
		parserFactory: parserFactory,
		t:             t,
	}
}

// Quick 快速测试
func (pt *ParserTester) Quick(source string, validator ValidationFunc) {
	builder := NewParserTestBuilder(pt.parserFactory)
	builder.Test(pt.t, source, validator)
}

// QuickVariable 快速测试变量赋值
func (pt *ParserTester) QuickVariable(varName, value string) {
	pt.Quick(QuickPHP(varName+" = "+value+";"), ValidateVariable(varName))
}

// QuickString 快速测试字符串赋值
func (pt *ParserTester) QuickString(varName, value, raw string) {
	pt.Quick(QuickPHP(varName+" = "+raw+";"), ValidateStringAssignment(varName, value, raw))
}

// QuickError 快速测试错误
func (pt *ParserTester) QuickError(source, errorKeyword string) {
	errorBuilder := NewErrorTestBuilder(pt.parserFactory)
	errorBuilder.ExpectError(pt.t, source, errorKeyword)
}

// QuickSuccess 快速测试成功解析
func (pt *ParserTester) QuickSuccess(source string) {
	errorBuilder := NewErrorTestBuilder(pt.parserFactory)
	errorBuilder.ExpectNoError(pt.t, source)
}

// TestDataset 测试数据集
type TestDataset struct {
	Name  string
	Items []TestDataItem
}

// TestDataItem 测试数据项
type TestDataItem struct {
	Name        string
	Source      string
	ExpectedAST interface{} // 可以是任何期望的AST结构表示
	Tags        []string
	Skip        bool
}

// CreateDefaultParserFactory 创建默认解析器工厂的辅助函数
// 这个函数需要在parser包中实现
var CreateDefaultParserFactory func() ParserFactory

// MustCreateParser 必须创建解析器，失败则panic
func MustCreateParser(source string) (ParserInterface, *TestContext) {
	if CreateDefaultParserFactory == nil {
		panic("CreateDefaultParserFactory not set")
	}

	factory := CreateDefaultParserFactory()
	l := lexer.New(source)
	p := factory(l)

	ctx := &TestContext{
		Parser: p,
		Lexer:  l,
		Config: &TestConfig{StrictMode: false, ValidateAST: false},
	}

	return p, ctx
}

// DebugParser 调试解析器，输出详细信息
func DebugParser(t *testing.T, source string, parserFactory ParserFactory) {
	t.Helper()

	t.Logf("Parsing source: %s", source)

	l := lexer.New(source)
	p := parserFactory(l)
	program := p.ParseProgram()

	errors := p.Errors()
	if len(errors) > 0 {
		t.Logf("Parser errors: %v", errors)
	}

	if program != nil {
		t.Logf("Program body length: %d", len(program.Body))
		t.Logf("Program AST: %s", program.String())
	} else {
		t.Log("Program is nil")
	}
}
