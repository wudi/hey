package testutils

import (
	"testing"

	"github.com/wudi/php-parser/ast"
)

// ParserInterface 定义解析器接口以避免循环依赖
type ParserInterface interface {
	ParseProgram() *ast.Program
	Errors() []string
}

// CheckParserErrors 检查解析器错误 - 从原有代码迁移
func CheckParserErrors(t *testing.T, p ParserInterface) {
	t.Helper()
	errors := p.Errors()
	if len(errors) == 0 {
		return
	}

	t.Errorf("parser has %d errors", len(errors))
	for _, msg := range errors {
		t.Errorf("parser error: %q", msg)
	}
	t.FailNow()
}
