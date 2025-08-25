package main

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/yourname/php-parser/lexer"
	"github.com/yourname/php-parser/parser"
)

func main() {
	fmt.Println("=== PHP Parser 与官方 PHP 实现对比测试 ===")
	fmt.Println()

	testCases := []string{
		`<?php echo "Hello, World!"; ?>`,
		`<?php $name = "John"; echo $name; ?>`,
		`<?php $a = 1 + 2 * 3; ?>`,
		`<?php if ($x > 5) { echo "big"; } ?>`,
		`<?php for ($i = 0; $i < 5; $i++) { echo $i; } ?>`,
		`<?php function test($param) { return $param * 2; } ?>`,
	}

	for i, testCase := range testCases {
		fmt.Printf("测试案例 %d: %s\n", i+1, testCase)
		fmt.Println(strings.Repeat("-", 80))

		// 使用我们的 parser
		fmt.Println("我们的解析结果:")
		testWithOurParser(testCase)
		
		fmt.Println()
		
		// 使用 PHP 官方实现
		fmt.Println("PHP 官方 token_get_all() 结果:")
		testWithPHPOfficial(testCase)
		
		fmt.Println()
		fmt.Println(strings.Repeat("=", 80))
		fmt.Println()
	}
}

func testWithOurParser(code string) {
	lex := lexer.New(code)
	p := parser.New(lex)
	
	// 收集 tokens
	lex2 := lexer.New(code)
	fmt.Println("Tokens:")
	tokenCount := 0
	for {
		token := lex2.NextToken()
		if token.Type == lexer.T_EOF {
			break
		}
		tokenCount++
		fmt.Printf("  %d: %-25s %q\n", tokenCount, lexer.TokenNames[token.Type], token.Value)
	}
	
	// 解析 AST
	program := p.ParseProgram()
	errors := p.Errors()
	
	fmt.Printf("语句数量: %d\n", len(program.Body))
	
	if len(errors) > 0 {
		fmt.Printf("解析错误: %d\n", len(errors))
		for i, err := range errors {
			fmt.Printf("  错误 %d: %s\n", i+1, err)
		}
	} else {
		fmt.Println("解析成功，无错误")
	}
}

func testWithPHPOfficial(code string) {
	// 构建 PHP 代码来调用 token_get_all
	phpCode := fmt.Sprintf(`<?php
$tokens = token_get_all(%s);
$count = 0;
foreach ($tokens as $token) {
    $count++;
    if (is_array($token)) {
        printf("  %%d: %%s \"%%s\"\n", $count, token_name($token[0]), addcslashes($token[1], "\r\n\t\"\\"));
    } else {
        printf("  %%d: CHAR \"%%s\"\n", $count, $token);
    }
}
printf("Token 总数: %%d\n", $count);
?>`, phpStringLiteral(code))

	// 执行 PHP 代码
	cmd := exec.Command("/bin/php", "-r", phpCode)
	output, err := cmd.Output()
	
	if err != nil {
		fmt.Printf("执行 PHP 失败: %v\n", err)
		return
	}
	
	fmt.Print(string(output))
}

// phpStringLiteral 将字符串转换为 PHP 字符串字面量
func phpStringLiteral(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\t", "\\t")
	return "\"" + s + "\""
}