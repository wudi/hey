package parser

import (
	"testing"

	"github.com/wudi/php-parser/lexer"
)

func BenchmarkParserCreation(b *testing.B) {
	source := `<?php $x = 1 + 2;`
	b.ResetTimer()
	b.Run("Original", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			l := lexer.New(source)
			p := New(l) // 原版本
			_ = p
		}
	})

	b.Run("WithPool", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			l := lexer.New(source)
			p := NewFromPool(l)
			p.Release()
		}
	})
}
