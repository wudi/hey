package main

import (
	"fmt"
	"strings"

	"github.com/wudi/php-parser/lexer"
	"github.com/wudi/php-parser/parser"
)

func main() {
	// 示例 PHP 代码
	phpCode := `<?php
function greet($name) {
    echo "Hello, ", $name, "!";
    return true;
}

$user = "World";
$age = 25;

if ($age > 18) {
    greet($user);
} else {
    echo "Too young!";
}

for ($i = 0; $i < 5; $i++) {
    echo $i, " ";
}

$numbers = array(1, 2, 3, 4, 5);
$sum = 0;

while ($sum < 10) {
    $sum++;
}

echo "Final sum: ", $sum;
?>`

	fmt.Println("=== PHP Parser 示例 ===")
	fmt.Printf("输入的 PHP 代码:\n%s\n", phpCode)
	fmt.Println(strings.Repeat("=", 60))

	// 词法分析
	fmt.Println("步骤 1: 词法分析 (Lexical Analysis)")
	fmt.Println("正在将 PHP 代码转换为 Token 流...")

	lex := lexer.New(phpCode)
	tokens := []lexer.Token{}

	for {
		token := lex.NextToken()
		tokens = append(tokens, token)
		if token.Type == lexer.T_EOF {
			break
		}
	}

	fmt.Printf("生成了 %d 个 tokens\n", len(tokens)-1) // 减去 EOF

	// 显示前几个 tokens
	fmt.Println("前 10 个 tokens:")
	for i := 0; i < 10 && i < len(tokens)-1; i++ {
		token := tokens[i]
		fmt.Printf("  %2d: %-25s %q\n", i+1, lexer.TokenNames[token.Type], token.Value)
	}

	fmt.Println(strings.Repeat("-", 60))

	// 语法分析
	fmt.Println("步骤 2: 语法分析 (Syntax Analysis)")
	fmt.Println("正在构建抽象语法树 (AST)...")

	// 重新创建 lexer 用于 parser
	lex2 := lexer.New(phpCode)
	p := parser.New(lex2)
	program := p.ParseProgram()

	// 检查解析错误
	parserErrors := p.Errors()
	if len(parserErrors) > 0 {
		fmt.Printf("发现 %d 个解析错误:\n", len(parserErrors))
		for i, err := range parserErrors {
			fmt.Printf("  错误 %d: %s\n", i+1, err)
		}
		fmt.Println(strings.Repeat("-", 60))
	}

	// 检查词法错误
	lexerErrors := lex2.GetErrors()
	if len(lexerErrors) > 0 {
		fmt.Printf("发现 %d 个词法错误:\n", len(lexerErrors))
		for i, err := range lexerErrors {
			fmt.Printf("  错误 %d: %s\n", i+1, err)
		}
		fmt.Println(strings.Repeat("-", 60))
	}

	fmt.Printf("解析完成! 生成了包含 %d 个语句的程序节点\n", len(program.Body))

	// 显示 AST 结构
	fmt.Println("AST 结构概览:")
	for i, stmt := range program.Body {
		fmt.Printf("  语句 %d: %T\n", i+1, stmt)
	}

	fmt.Println(strings.Repeat("-", 60))

	// 步骤 3: AST 重建源代码
	fmt.Println("步骤 3: 从 AST 重建源代码")
	fmt.Println("正在从 AST 生成 PHP 代码...")

	reconstructed := program.String()
	fmt.Printf("重建的代码:\n%s\n", reconstructed)

	fmt.Println(strings.Repeat("-", 60))

	// 步骤 4: 生成 JSON 表示
	fmt.Println("步骤 4: 生成 JSON 表示")
	fmt.Println("正在将 AST 转换为 JSON...")

	jsonData, err := program.ToJSON()
	if err != nil {
		fmt.Printf("JSON 转换失败: %v\n", err)
	} else {
		fmt.Printf("AST JSON 表示 (前 500 字符):\n%s...\n",
			truncateString(string(jsonData), 500))
	}

	fmt.Println(strings.Repeat("=", 60))

	// 步骤 5: 错误处理示例
	fmt.Println("步骤 5: 错误处理示例")
	testErrorHandling()

	fmt.Println("=== 示例完成 ===")
}

// testErrorHandling 测试错误处理
func testErrorHandling() {
	fmt.Println("正在测试错误处理...")

	// 包含语法错误的 PHP 代码
	badPHPCode := `<?php
echo "Missing semicolon"
$invalid = ;
if ($x > 5 {
    echo "Missing parenthesis";
}
?>`

	fmt.Printf("包含错误的 PHP 代码:\n%s\n", badPHPCode)

	lex := lexer.New(badPHPCode)
	p := parser.New(lex)
	program := p.ParseProgram()

	parserErrors := p.Errors()
	if len(parserErrors) > 0 {
		fmt.Printf("发现 %d 个解析错误:\n", len(parserErrors))
		for i, err := range parserErrors {
			fmt.Printf("  错误 %d: %s\n", i+1, err)
		}
	} else {
		fmt.Println("未发现解析错误")
	}

	fmt.Printf("尽管有错误，仍然解析了 %d 个语句\n", len(program.Body))
}

// truncateString 截断字符串
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}
