// Package parser
// with object pooling to reduce allocations.
package parser

import (
	"sync"

	"github.com/wudi/php-parser/lexer"
)

var parserPool = sync.Pool{New: func() any { return &Parser{} }}

// NewFromPool 从对象池中获取一个 Parser 实例
func NewFromPool(l *lexer.Lexer) *Parser {
	p := parserPool.Get().(*Parser)
	p.lexer = l
	p.errors = p.errors[:0] // 重置但保留容量

	p.nextToken()
	p.nextToken()
	return p
}

// Release 将 Parser 归还给对象池
func (p *Parser) Release() {
	p.lexer = nil
	p.peekToken = lexer.Token{}
	p.currentToken = p.peekToken

	parserPool.Put(p)
}
