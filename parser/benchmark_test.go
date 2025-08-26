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

// BenchmarkParsingComplexity 测试不同复杂度的解析性能
func BenchmarkParsingComplexity(b *testing.B) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name:   "Simple Assignment",
			source: `<?php $x = 42; ?>`,
		},
		{
			name:   "Complex Expression",
			source: `<?php $result = ($a + $b * $c) / ($d - $e) ** 2; ?>`,
		},
		{
			name:   "Array Operations",
			source: `<?php $arr = [1, 2, 3, "key" => "value", $var => $other]; ?>`,
		},
		{
			name:   "Function Declaration",
			source: `<?php function test($a, $b = null) { return $a + $b; } ?>`,
		},
		{
			name:   "Control Structures",
			source: `<?php if ($x > 0) { for ($i = 0; $i < 10; $i++) { echo $i; } } ?>`,
		},
		{
			name:   "String Operations",
			source: `<?php $str = "Hello " . $name . " from " . $location; ?>`,
		},
		{
			name:   "Heredoc String",
			source: `<?php $html = <<<HTML
<div class="test">
    <p>Hello $name</p>
</div>
HTML; ?>`,
		},
		{
			name:   "Nowdoc String",
			source: `<?php $text = <<<'TEXT'
Raw text with $variables
TEXT; ?>`,
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				l := lexer.New(tt.source)
				p := New(l)
				program := p.ParseProgram()
				_ = program
			}
		})
	}
}

// BenchmarkStringParsing 专门测试字符串解析性能
func BenchmarkStringParsing(b *testing.B) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name:   "Simple String",
			source: `<?php $str = "Hello World"; ?>`,
		},
		{
			name:   "String with Variables",
			source: `<?php $str = "Hello $name, you are $age years old"; ?>`,
		},
		{
			name:   "Nowdoc",
			source: `<?php $text = <<<'TEXT'
This is raw text
With $variables that won't be interpolated
TEXT; ?>`,
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				l := lexer.New(tt.source)
				p := New(l)
				program := p.ParseProgram()
				_ = program
			}
		})
	}
}
