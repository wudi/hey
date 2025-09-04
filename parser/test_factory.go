package parser

import (
	"github.com/wudi/php-parser/lexer"
	"github.com/wudi/php-parser/parser/testutils"
)

// createParserFactory 创建解析器工厂函数 - 共享的测试工厂函数
func createParserFactory() testutils.ParserFactory {
	return func(l *lexer.Lexer) testutils.ParserInterface {
		return New(l)
	}
}
