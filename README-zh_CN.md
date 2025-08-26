# PHP 解析器

一个用 Go 语言编写的全面的 PHP 解析器，能够将 PHP 源代码转换为抽象语法树（AST）。该解析器支持 PHP 7+ 语法，并与 PHP 官方实现保持高度兼容。

## 功能特性

- **完整的 PHP 词法分析器**: 150+ 种符合 PHP 8.4 规范的 Token 类型
- **递归下降解析器**: 采用 Pratt 解析算法，正确处理操作符优先级
- **完整的 AST 支持**: 基于接口的 AST 节点，支持访问者模式
- **PHP 兼容性**: Token ID 和 AST 结构与 PHP 官方实现保持一致
- **全面的语法支持**:
  - 变量、函数、类和控制流
  - 带可见性修饰符的类常量、属性和方法
  - Try-catch-finally 语句
  - 支持插值的 Heredoc/Nowdoc 字符串
  - 类型化参数和引用参数
  - 复杂表达式和方法链调用

## 安装

```bash
git clone https://github.com/wudi/php-parser.git
cd php-parser
go build -o php-parser ./cmd/php-parser
```

## 使用方法

### 命令行界面

```bash
# 从标准输入解析 PHP 代码
echo '<?php echo "Hello, World!"; ?>' | ./php-parser

# 解析 PHP 文件
./php-parser -i example.php

# 显示 Token 和 AST 结构
./php-parser -tokens -ast example.php

# 输出格式：json（默认）、ast、tokens
./php-parser -format json example.php

# 仅显示解析错误
./php-parser -errors example.php
```

### 编程接口使用

```go
package main

import (
    "fmt"
    "github.com/wudi/php-parser/lexer"
    "github.com/wudi/php-parser/parser"
)

func main() {
    input := `<?php
    $name = "World";
    echo "Hello, " . $name;
    ?>`
    
    l := lexer.New(input)
    p := parser.New(l)
    program := p.ParseProgram()
    
    // 检查解析错误
    if errors := p.Errors(); len(errors) > 0 {
        for _, err := range errors {
            fmt.Printf("解析错误: %s\n", err)
        }
        return
    }
    
    // 使用 AST
    fmt.Printf("解析了 %d 条语句\n", len(program.Body))
}
```

## 架构设计

解析器遵循经典的编译器前端设计：

```
PHP 源代码 → 词法分析器 → Token 流 → 语法分析器 → 抽象语法树
```

### 核心模块

- **`lexer/`**: 带状态机的词法分析器（11 种状态）
- **`parser/`**: 采用 Pratt 解析的递归下降解析器
- **`ast/`**: AST 节点定义和工具
- **`cmd/php-parser/`**: 命令行界面
- **`errors/`**: 带位置跟踪的错误处理

## 测试

```bash
# 运行所有测试
go test ./...

# 运行解析器测试，输出详细信息
go test ./parser -v

# 运行特定测试套件
go test ./parser -run=TestParsing_TryCatchWithStatements
go test ./parser -run=TestParsing_ClassMethodsWithVisibility

# 运行基准测试
go test ./parser -bench=.
```

## PHP 兼容性

解析器与 PHP 官方实现保持高度兼容：

- Token ID 符合 PHP 8.4 规范
- AST 节点类型与 PHP 的 `zend_ast.h` 保持一致
- 词法分析器状态与 PHP 的词法分析器相同
- 全面的测试套件通过 PHP 的 `token_get_all()` 进行验证

## 支持的 PHP 语法

### 基础构造
- 变量和常量
- 所有数据类型（整数、浮点数、字符串、数组、对象）
- 运算符（算术、比较、逻辑、位运算）

### 控制流
- If/else 语句
- Switch 语句
- 循环（for、foreach、while、do-while）
- Try-catch-finally 块

### 面向对象特性
- 带继承的类声明
- 带可见性修饰符和类型提示的属性
- 带可见性修饰符的方法
- 带可见性修饰符的类常量
- 静态访问和方法调用

### 函数
- 函数声明和调用
- 匿名函数（闭包）
- 类型化参数和返回类型
- 引用参数
- 可变参数

### 高级特性
- Heredoc 和 Nowdoc 字符串
- 字符串插值
- 数组语法（`array()` 和 `[]` 两种形式）
- 复杂表达式和方法链调用

## 示例

### 基础解析

```php
<?php
$users = [
    ['name' => 'John', 'age' => 30],
    ['name' => 'Jane', 'age' => 25]
];

foreach ($users as $user) {
    echo $user['name'] . " is " . $user['age'] . " years old\n";
}
```

### 带方法的类

```php
<?php
class UserManager {
    private array $users = [];
    
    public function addUser(string $name, int $age): void {
        $this->users[] = ['name' => $name, 'age' => $age];
    }
    
    protected function getUserCount(): int {
        return count($this->users);
    }
}
```

### 带复杂表达式的 Try-Catch

```php
<?php
try {
    $result = $service->processData($input);
    $processed = $result->transform()->validate();
} catch (ValidationException $e) {
    $logger->error($e->getMessage());
    throw new ProcessingException('Validation failed');
} finally {
    $cleanup->execute();
}

$finalResult = $processor->complete($processed);
```

## 开发

### 环境要求
- Go 1.21+
- PHP 8.4（用于兼容性测试）

### 运行测试
```bash
# 所有测试
go test ./...

# 特定模块
go test ./lexer -v
go test ./parser -v
go test ./ast -v

# 带基准测试
go test ./parser -bench=. -benchmem
```

### 贡献指南
1. 遵循 Go 编码规范
2. 保持 PHP 兼容性
3. 为新功能添加全面的测试
4. 参考 PHP 官方语法（`/home/ubuntu/php-src/Zend/zend_language_parser.y`）

## 许可证

本项目为开源项目。请查看仓库以了解许可证详情。