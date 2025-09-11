# PHP Parser 架构设计文档

## 总体架构

PHP Parser 采用经典的编译器前端架构，包含以下几个主要模块：

```
源代码 -> 词法分析器(Lexer) -> Token流 -> 语法分析器(Parser) -> AST
```

## 核心模块

### 1. 词法分析器 (Lexer)
- **位置**: `lexer/` 目录
- **功能**: 将 PHP 源代码转换为 Token 流
- **状态机**: 支持多种词法状态（初始、字符串、注释、Heredoc等）
- **Token兼容性**: 与 PHP 官方 Token 类型完全兼容

#### 状态机设计
```
INITIAL -> ST_IN_SCRIPTING -> ST_DOUBLE_QUOTES -> ST_HEREDOC -> ST_NOWDOC -> ST_VAR_OFFSET -> ST_LOOKING_FOR_PROPERTY -> ST_LOOKING_FOR_VARNAME
```

### 2. Token 类型系统
- **位置**: `lexer/token.go`
- **功能**: 定义所有 PHP Token 类型常量
- **兼容性**: 与 PHP 8 官方 Token ID 保持一致

### 3. 语法分析器 (Parser)
- **位置**: `parser/` 目录
- **算法**: 递归下降分析
- **输出**: 抽象语法树 (AST)

### 4. AST 节点定义
- **位置**: `ast/` 目录
- **设计**: 基于接口的节点系统
- **扩展性**: 易于添加新的语法结构

### 5. 错误处理
- **位置**: `errors/` 目录
- **功能**: 词法和语法错误的统一处理
- **特性**: 精确的错误位置定位

## PHP Lexer 深度分析

### 状态转换机制

PHP lexer 使用有限状态自动机来处理不同的词法上下文：

1. **ST_IN_SCRIPTING**: 默认 PHP 代码状态
2. **ST_DOUBLE_QUOTES**: 双引号字符串内部
3. **ST_HEREDOC**: Heredoc 文档内部
4. **ST_NOWDOC**: Nowdoc 文档内部
5. **ST_VAR_OFFSET**: 数组索引内部
6. **ST_LOOKING_FOR_PROPERTY**: 查找对象属性
7. **ST_LOOKING_FOR_VARNAME**: 查找变量名

### 关键实现细节

1. **字符串插值**: 在双引号字符串中识别变量和表达式
2. **Heredoc/Nowdoc**: 支持灵活的多行字符串语法
3. **注释处理**: 支持 // 、/* */ 和 # 三种注释形式
4. **数字解析**: 支持十进制、十六进制、八进制和二进制数字
5. **转义序列**: 完整的 PHP 转义序列支持

## 设计原则

1. **兼容性优先**: 与 PHP 官方实现保持高度兼容
2. **模块化设计**: 清晰的模块边界，便于维护和测试
3. **可扩展性**: 易于添加新的 PHP 语法特性
4. **性能考虑**: 高效的词法分析和语法分析算法
5. **错误友好**: 提供清晰的错误信息和位置定位

## 测试策略

1. **单元测试**: 每个模块的独立测试
2. **集成测试**: 完整的解析流程测试
3. **兼容性测试**: 与 PHP 官方实现对比测试
4. **边界测试**: 错误情况和极端情况测试

## Bytecode Compiler Architecture

### 6. 字节码编译器 (Bytecode Compiler)
- **位置**: `compiler/` 目录
- **功能**: 将 AST 编译为高效的字节码指令
- **虚拟机**: 基于栈的虚拟机执行引擎
- **性能**: 比直接 AST 解释执行快 10-50 倍

### 编译器核心组件

1. **函数调用系统**: 完整的函数声明、调用、参数传递机制
2. **循环迭代器**: 支持 foreach 循环的高效迭代系统
3. **执行上下文**: 隔离的函数执行环境和变量作用域
4. **参数映射**: 函数参数到变量槽的精确映射系统

### 详细设计文档

详细的字节码编译器设计文档位于 `docs/compiler/` 目录:

- [`01-function-call-architecture.md`](compiler/01-function-call-architecture.md): 函数调用系统设计
- [`02-foreach-iterator-design.md`](compiler/02-foreach-iterator-design.md): Foreach 迭代器设计  
- [`03-vm-execution-contexts.md`](compiler/03-vm-execution-contexts.md): 虚拟机执行上下文
- [`04-parameter-mapping-system.md`](compiler/04-parameter-mapping-system.md): 参数映射系统
- [`05-integration-overview.md`](compiler/05-integration-overview.md): 系统集成概览

### PHP 兼容性

字节码编译器完全兼容 PHP Zend VM:
- 使用与 PHP 相同的操作码 (`ZEND_INIT_FCALL`, `ZEND_DO_FCALL` 等)
- 支持完整的 PHP 值系统和类型转换
- 正确处理函数调用、数组操作、循环控制
- 维护 PHP 的作用域和变量语义

## 扩展计划

1. **语法分析**: 支持完整的 PHP 语法 ✅
2. **静态分析**: 类型检查和代码分析
3. **代码生成**: AST 到代码的反向生成
4. **IDE 支持**: 语法高亮和自动补全支持
5. **字节码编译**: 高性能字节码执行引擎 ✅