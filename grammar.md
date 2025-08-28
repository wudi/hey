# PHP 语法分析完整报告

基于对 PHP 官方实现 (`/home/ubuntu/php-src/Zend/`) 的完整分析，本文档记录了 PHP 语言的完整语法结构、AST 节点类型和词法规则。

## 1. 语法规则层次结构

### 1.1 顶层语法规则

```yacc
start → top_statement_list

top_statement_list → top_statement_list top_statement 
                  | ε

top_statement → statement
              | attributed_top_statement  
              | attributes attributed_top_statement
              | T_HALT_COMPILER '(' ')' ';'
              | namespace_statement
              | use_statement
```

### 1.2 语句分类

**控制结构语句:**
```yacc
statement → '{' inner_statement_list '}'
          | if_stmt | alt_if_stmt
          | T_WHILE '(' expr ')' while_statement
          | T_DO statement T_WHILE '(' expr ')' ';'
          | T_FOR '(' for_exprs ';' for_cond_exprs ';' for_exprs ')' for_statement
          | T_SWITCH '(' expr ')' switch_case_list
          | T_FOREACH '(' expr T_AS foreach_variable ')' foreach_statement
          | T_FOREACH '(' expr T_AS foreach_variable T_DOUBLE_ARROW foreach_variable ')' foreach_statement
          | T_TRY '{' inner_statement_list '}' catch_list finally_statement
```

**声明语句:**
```yacc
attributed_statement → function_declaration_statement
                     | class_declaration_statement
                     | trait_declaration_statement  
                     | interface_declaration_statement
                     | enum_declaration_statement
```

**简单语句:**
```yacc
statement → T_BREAK optional_expr ';'
          | T_CONTINUE optional_expr ';'
          | T_RETURN optional_expr ';'
          | T_GLOBAL global_var_list ';'
          | T_STATIC static_var_list ';'
          | T_ECHO echo_expr_list ';'
          | T_UNSET '(' unset_variables possible_comma ')' ';'
          | expr ';'
          | ';'  /* empty statement */
          | T_GOTO T_STRING ';'
          | T_STRING ':'  /* label */
```

### 1.3 表达式层次结构

**表达式优先级 (从低到高):**

| 级别 | 运算符 | 结合性 | AST节点类型 |
|------|--------|--------|-------------|
| 1 | `throw` | precedence | `ZEND_AST_THROW` |
| 2 | Arrow functions | precedence | `ZEND_AST_ARROW_FUNC` |
| 3 | `include`, `include_once`, `require`, `require_once` | precedence | `ZEND_AST_INCLUDE_OR_EVAL` |
| 4 | `or` | left | `ZEND_AST_OR` |
| 5 | `xor` | left | `ZEND_AST_XOR` |
| 6 | `and` | left | `ZEND_AST_AND` |
| 7 | `print` | precedence | `ZEND_AST_PRINT` |
| 8 | `yield`, `yield from` | precedence | `ZEND_AST_YIELD` |
| 9 | 赋值运算符 (`=`, `+=`, `-=`, etc.) | precedence | `ZEND_AST_ASSIGN*` |
| 10 | `? :` | left | `ZEND_AST_CONDITIONAL` |
| 11 | `??` | right | `ZEND_AST_COALESCE` |
| 12 | `\|\|` | left | `ZEND_AST_OR` |
| 13 | `&&` | left | `ZEND_AST_AND` |
| 14 | `\|` | left | `ZEND_AST_BINARY_OP` |
| 15 | `^` | left | `ZEND_AST_BINARY_OP` |
| 16 | `&` | left | `ZEND_AST_BINARY_OP` |
| 17 | `==`, `!=`, `===`, `!==`, `<=>` | nonassoc | `ZEND_AST_BINARY_OP` |
| 18 | `<`, `<=`, `>`, `>=` | nonassoc | `ZEND_AST_BINARY_OP` |
| 19 | `\|>` | left | `ZEND_AST_PIPE` |
| 20 | `.` | left | `ZEND_AST_CONCAT` |
| 21 | `<<`, `>>` | left | `ZEND_AST_BINARY_OP` |
| 22 | `+`, `-` | left | `ZEND_AST_BINARY_OP` |
| 23 | `*`, `/`, `%` | left | `ZEND_AST_BINARY_OP` |
| 24 | `!` | precedence | `ZEND_AST_UNARY_OP` |
| 25 | `instanceof` | precedence | `ZEND_AST_INSTANCEOF` |
| 26 | `~`, casts, `@` | precedence | `ZEND_AST_UNARY_OP` |
| 27 | `**` | right | `ZEND_AST_BINARY_OP` |
| 28 | `clone` | precedence | `ZEND_AST_CLONE` |

## 2. AST 节点类型完整列表

### 2.1 节点分类系统

**编码规则:**
- 位0-5: 节点ID
- 位6: ZEND_AST_SPECIAL 标志 
- 位7: ZEND_AST_IS_LIST 标志
- 位8+: 子节点数量

**节点类别:**

### 2.2 特殊节点 (0-3)
```c
ZEND_AST_ZVAL = 0           // 字面量值
ZEND_AST_CONSTANT = 1       // 命名常量  
ZEND_AST_ZNODE = 2          // 编译时节点
ZEND_AST_FUNC_DECL = 3      // 函数声明
```

### 2.3 声明节点 (64-69)
```c
ZEND_AST_CLOSURE = 64       // 匿名函数
ZEND_AST_METHOD = 65        // 类方法
ZEND_AST_CLASS = 66         // 类声明
ZEND_AST_ARROW_FUNC = 67    // 箭头函数
ZEND_AST_ENUM = 68          // 枚举声明
```

### 2.4 列表节点 (128-149) - 可变长度
```c
ZEND_AST_ARG_LIST = 128           // 参数列表
ZEND_AST_ARRAY = 129              // 数组字面量
ZEND_AST_ENCAPS_LIST = 130        // 字符串插值列表
ZEND_AST_EXPR_LIST = 131          // 表达式列表
ZEND_AST_STMT_LIST = 132          // 语句列表
ZEND_AST_IF = 133                 // if语句链
ZEND_AST_SWITCH_LIST = 134        // switch案例列表
ZEND_AST_CATCH_LIST = 135         // catch子句列表
ZEND_AST_PARAM_LIST = 136         // 形参列表
ZEND_AST_CLOSURE_USES = 137       // use变量列表
ZEND_AST_PROP_GROUP = 138         // 属性组
ZEND_AST_CONST_DECL = 139         // 常量声明列表
ZEND_AST_CLASS_CONST_GROUP = 140  // 类常量组
ZEND_AST_NAME_LIST = 141          // 名称列表
ZEND_AST_TRAIT_ADAPTATIONS = 142  // trait适配列表
ZEND_AST_USE = 143                // use声明列表
ZEND_AST_ATTRIBUTE_GROUP = 144    // 属性组
ZEND_AST_MATCH_ARM_LIST = 145     // match分支列表
ZEND_AST_ENUM_CASE_LIST = 146     // 枚举案例列表
ZEND_AST_PROPERTY_HOOK_LIST = 147 // 属性钩子列表
```

### 2.5 表达式节点 (256-287) - 固定子节点数
```c
// 0个子节点
ZEND_AST_MAGIC_CONST = 256        // 魔术常量
ZEND_AST_TYPE = 257               // 类型声明

// 1个子节点  
ZEND_AST_VAR = 1*64 + 256         // 变量
ZEND_AST_CONST = 1*64 + 257       // 常量引用
ZEND_AST_UNPACK = 1*64 + 258      // 解包操作
ZEND_AST_UNARY_PLUS = 1*64 + 259  // 一元加
ZEND_AST_UNARY_MINUS = 1*64 + 260 // 一元减
ZEND_AST_CAST = 1*64 + 261        // 类型转换
ZEND_AST_EMPTY = 1*64 + 262       // empty()
ZEND_AST_ISSET = 1*64 + 263       // isset()
ZEND_AST_SILENCE = 1*64 + 264     // @ 操作符
ZEND_AST_SHELL_EXEC = 1*64 + 265  // 反引号执行
ZEND_AST_CLONE = 1*64 + 266       // clone
ZEND_AST_EXIT = 1*64 + 267        // exit/die
ZEND_AST_PRINT = 1*64 + 268       // print
ZEND_AST_INCLUDE_OR_EVAL = 1*64 + 269  // include/eval
ZEND_AST_UNARY_OP = 1*64 + 270    // 通用一元运算
ZEND_AST_PRE_INC = 1*64 + 271     // 前缀++
ZEND_AST_PRE_DEC = 1*64 + 272     // 前缀--
ZEND_AST_POST_INC = 1*64 + 273    // 后缀++
ZEND_AST_POST_DEC = 1*64 + 274    // 后缀--
ZEND_AST_YIELD_FROM = 1*64 + 275  // yield from
ZEND_AST_GLOBAL = 1*64 + 276      // global声明
ZEND_AST_UNSET = 1*64 + 277       // unset()
ZEND_AST_RETURN = 1*64 + 278      // return
ZEND_AST_LABEL = 1*64 + 279       // 标签
ZEND_AST_REF = 1*64 + 280         // 引用
ZEND_AST_HALT_COMPILER = 1*64 + 281  // __halt_compiler
ZEND_AST_ECHO = 1*64 + 282        // echo
ZEND_AST_THROW = 1*64 + 283       // throw
ZEND_AST_GOTO = 1*64 + 284        // goto

// 2个子节点
ZEND_AST_DIM = 2*64 + 256         // 数组访问
ZEND_AST_PROP = 2*64 + 257        // 属性访问
ZEND_AST_NULLSAFE_PROP = 2*64 + 258  // 空安全属性
ZEND_AST_STATIC_PROP = 2*64 + 259 // 静态属性
ZEND_AST_CALL = 2*64 + 260        // 函数调用
ZEND_AST_CLASS_CONST = 2*64 + 261 // 类常量
ZEND_AST_ASSIGN = 2*64 + 262      // 赋值
ZEND_AST_ASSIGN_REF = 2*64 + 263  // 引用赋值
ZEND_AST_ASSIGN_OP = 2*64 + 264   // 复合赋值
ZEND_AST_BINARY_OP = 2*64 + 265   // 二元运算
ZEND_AST_ARRAY_ELEM = 2*64 + 266  // 数组元素
ZEND_AST_NEW = 2*64 + 267         // new表达式
ZEND_AST_INSTANCEOF = 2*64 + 268  // instanceof
ZEND_AST_YIELD = 2*64 + 269       // yield
ZEND_AST_COALESCE = 2*64 + 270    // ??操作符
ZEND_AST_ASSIGN_COALESCE = 2*64 + 271  // ??=操作符
ZEND_AST_STATIC = 2*64 + 272      // static声明
ZEND_AST_WHILE = 2*64 + 273       // while循环
ZEND_AST_DO_WHILE = 2*64 + 274    // do-while循环
ZEND_AST_IF_ELEM = 2*64 + 275     // if条件元素
ZEND_AST_SWITCH_CASE = 2*64 + 276 // switch案例
ZEND_AST_CATCH = 2*64 + 277       // catch子句
ZEND_AST_PARAM = 2*64 + 278       // 参数
ZEND_AST_TYPE_UNION = 2*64 + 279  // 联合类型
ZEND_AST_TYPE_INTERSECTION = 2*64 + 280  // 交集类型
ZEND_AST_ATTRIBUTE = 2*64 + 281   // 属性
ZEND_AST_MATCH_ARM = 2*64 + 282   // match分支
ZEND_AST_ENUM_CASE = 2*64 + 283   // 枚举案例
ZEND_AST_PROPERTY_HOOK = 2*64 + 284  // 属性钩子

// 3个子节点
ZEND_AST_METHOD_CALL = 3*64 + 256      // 方法调用
ZEND_AST_NULLSAFE_METHOD_CALL = 3*64 + 257  // 空安全方法调用
ZEND_AST_STATIC_CALL = 3*64 + 258      // 静态方法调用
ZEND_AST_CONDITIONAL = 3*64 + 259      // 三元运算符
ZEND_AST_TRY = 3*64 + 260              // try语句
ZEND_AST_FOREACH = 3*64 + 261          // foreach循环
ZEND_AST_DECLARE = 3*64 + 262          // declare语句

// 4个子节点
ZEND_AST_FOR = 4*64 + 256              // for循环
ZEND_AST_SWITCH = 4*64 + 257           // switch语句
```

### 2.6 声明元素节点 (768-777)
```c
ZEND_AST_PROP_ELEM = 768          // 属性元素
ZEND_AST_CONST_ELEM = 769         // 常量元素  
ZEND_AST_USE_TRAIT = 770          // trait使用
ZEND_AST_TRAIT_PRECEDENCE = 771   // trait优先级
ZEND_AST_METHOD_REFERENCE = 772   // 方法引用
ZEND_AST_NAMESPACE = 773          // 命名空间
ZEND_AST_USE_ELEM = 774           // use元素
ZEND_AST_TRAIT_ALIAS = 775        // trait别名
ZEND_AST_GROUP_USE = 776          // 分组use
ZEND_AST_CLASS_NAME = 777         // 类名
```

## 3. 词法分析器状态和Token

### 3.1 词法状态机

**完整状态列表:**
```c
INITIAL                    // 初始状态，寻找PHP开始标签
SHEBANG                   // 处理Unix shebang行 (#!)
ST_IN_SCRIPTING           // PHP代码主要解析状态
ST_DOUBLE_QUOTES          // 双引号字符串内部状态  
ST_SINGLE_QUOTES          // 单引号字符串状态
ST_BACKQUOTE              // 反引号字符串(shell命令)状态
ST_HEREDOC                // Heredoc字符串状态
ST_NOWDOC                 // Nowdoc字符串状态
ST_END_HEREDOC            // Heredoc/Nowdoc结束标签状态
ST_LOOKING_FOR_PROPERTY   // 寻找对象属性状态
ST_LOOKING_FOR_VARNAME    // 寻找变量名状态
ST_VAR_OFFSET             // 数组索引/偏移量状态
```

### 3.2 Token类型分类

**基本数据类型Token (12个):**
```c
T_LNUMBER                 // 整数字面量
T_DNUMBER                 // 浮点数字面量  
T_STRING                  // 标识符
T_VARIABLE                // 变量 ($var)
T_CONSTANT_ENCAPSED_STRING // 字符串常量
T_ENCAPSED_AND_WHITESPACE // 字符串插值片段
T_NUM_STRING              // 数字字符串
T_INLINE_HTML             // HTML内联代码
T_NAME_FULLY_QUALIFIED    // 完全限定名 (\Foo\Bar)
T_NAME_QUALIFIED          // 限定名 (Foo\Bar)  
T_NAME_RELATIVE           // 相对名 (namespace\Foo)
T_STRING_VARNAME          // 字符串中的变量名
```

**关键字Token (84个):**
```c
// 控制结构
T_IF, T_ELSE, T_ELSEIF, T_ENDIF
T_WHILE, T_ENDWHILE, T_DO
T_FOR, T_ENDFOR, T_FOREACH, T_ENDFOREACH  
T_SWITCH, T_ENDSWITCH, T_CASE, T_DEFAULT, T_MATCH
T_TRY, T_CATCH, T_FINALLY, T_THROW
T_BREAK, T_CONTINUE, T_GOTO, T_RETURN

// 函数和类
T_FUNCTION, T_FN, T_YIELD, T_YIELD_FROM
T_CLASS, T_INTERFACE, T_TRAIT, T_ENUM
T_EXTENDS, T_IMPLEMENTS, T_ABSTRACT, T_FINAL  
T_PUBLIC, T_PRIVATE, T_PROTECTED, T_READONLY
T_STATIC, T_CONST, T_VAR, T_NEW, T_CLONE

// 其他关键字
T_INCLUDE, T_INCLUDE_ONCE, T_REQUIRE, T_REQUIRE_ONCE
T_ECHO, T_PRINT, T_EVAL, T_EXIT
T_GLOBAL, T_UNSET, T_ISSET, T_EMPTY, T_LIST, T_ARRAY
T_NAMESPACE, T_USE, T_INSTEADOF, T_AS
T_DECLARE, T_ENDDECLARE
T_HALT_COMPILER, T_CALLABLE, T_INSTANCEOF
```

**运算符Token (51个):**
```c
// 算术运算符
T_PLUS_EQUAL (+=), T_MINUS_EQUAL (-=), T_MUL_EQUAL (*=)
T_DIV_EQUAL (/=), T_MOD_EQUAL (%=), T_POW_EQUAL (**=)
T_POW (**), T_INC (++), T_DEC (--)

// 比较运算符  
T_IS_EQUAL (==), T_IS_NOT_EQUAL (!=)
T_IS_IDENTICAL (===), T_IS_NOT_IDENTICAL (!==)
T_IS_SMALLER_OR_EQUAL (<=), T_IS_GREATER_OR_EQUAL (>=)
T_SPACESHIP (<=>), T_COALESCE (??)

// 逻辑运算符
T_BOOLEAN_AND (&&), T_BOOLEAN_OR (||)
T_LOGICAL_AND (and), T_LOGICAL_OR (or), T_LOGICAL_XOR (xor)

// 位运算符
T_AND_EQUAL (&=), T_OR_EQUAL (|=), T_XOR_EQUAL (^=)
T_SL (<<), T_SR (>>), T_SL_EQUAL (<<=), T_SR_EQUAL (>>=)

// 字符串运算符
T_CONCAT_EQUAL (.=), T_COALESCE_EQUAL (??=)

// 对象运算符
T_OBJECT_OPERATOR (->), T_NULLSAFE_OBJECT_OPERATOR (?->)
T_PAAMAYIM_NEKUDOTAYIM (::), T_DOUBLE_ARROW (=>)

// 类型转换
T_INT_CAST, T_DOUBLE_CAST, T_STRING_CAST
T_ARRAY_CAST, T_OBJECT_CAST, T_BOOL_CAST, T_UNSET_CAST
```

**特殊Token (24个):**
```c
// PHP标签
T_OPEN_TAG, T_OPEN_TAG_WITH_ECHO, T_CLOSE_TAG

// 字符串相关
T_START_HEREDOC, T_END_HEREDOC
T_CURLY_OPEN, T_DOLLAR_OPEN_CURLY_BRACES
T_ATTRIBUTE (#[)

// 魔术常量
T_LINE, T_FILE, T_DIR, T_CLASS_C, T_TRAIT_C
T_METHOD_C, T_FUNC_C, T_PROPERTY_C, T_NS_C

// 其他特殊
T_ELLIPSIS (...), T_NS_SEPARATOR (\)
T_PIPE (|>), T_AMPERSAND_FOLLOWED_BY_VAR_OR_VARARG
T_AMPERSAND_NOT_FOLLOWED_BY_VAR_OR_VARARG
T_COMMENT, T_DOC_COMMENT, T_WHITESPACE
T_BAD_CHARACTER, T_ERROR
```

### 3.3 字符串处理规则

**单引号字符串:**
- 不支持变量插值
- 仅处理 `\'` 和 `\\` 转义
- Token: `T_CONSTANT_ENCAPSED_STRING`

**双引号字符串:**
- 支持变量插值: `$var`, `$obj->prop`, `$arr[key]`
- 支持复杂插值: `{$expr}`
- 转义: Unicode `\u{xxxx}`, 十六进制 `\xFF`, 八进制 `\377`
- Token: `T_ENCAPSED_AND_WHITESPACE` (插值), `T_CONSTANT_ENCAPSED_STRING` (简单)

**Heredoc/Nowdoc:**
- Heredoc: `<<<LABEL`, 支持变量插值
- Nowdoc: `<<<'LABEL'`, 不支持插值
- 支持灵活缩进语法 (PHP 7.3+)
- Token: `T_START_HEREDOC`, `T_END_HEREDOC`, 内容为插值token

## 4. 语法构造的解析方式

### 4.1 类声明解析
```yacc
class_declaration_statement:
    optional_attributes class_modifiers T_CLASS identifier extends_from implements_list '{' class_statement_list '}'
    
class_statement_list:
    class_statement_list attributed_class_statement
    | %empty

attributed_class_statement:  
    class_statement
    | attributes class_statement

class_statement:
    property_modifiers optional_type property_list ';'
    | class_const_modifiers T_CONST class_const_list ';'  
    | method_modifiers function identifier '(' parameter_list ')' return_type method_body
    | T_USE name_list trait_adaptations
```

### 4.2 函数声明解析
```yacc
function_declaration_statement:
    optional_attributes T_FUNCTION optional_ref identifier '(' parameter_list ')' return_type '{' inner_statement_list '}'

parameter_list:
    non_empty_parameter_list
    | %empty
    
parameter:
    optional_attributes optional_visibility optional_type optional_ref optional_ellipsis T_VARIABLE optional_default
```

### 4.3 表达式解析模式

**前缀表达式:**
- 变量: `$var` → `ZEND_AST_VAR`
- 字面量: `123`, `"string"` → `ZEND_AST_ZVAL` 
- 函数调用: `func()` → `ZEND_AST_CALL`
- 一元运算: `-expr`, `!expr` → `ZEND_AST_UNARY_OP`

**中缀表达式:**  
- 二元运算: `a + b` → `ZEND_AST_BINARY_OP`
- 方法调用: `$obj->method()` → `ZEND_AST_METHOD_CALL`
- 数组访问: `$arr[key]` → `ZEND_AST_DIM`
- 赋值: `$var = value` → `ZEND_AST_ASSIGN`

**后缀表达式:**
- 自增/自减: `$var++` → `ZEND_AST_POST_INC`
- 数组访问: `expr[key]` → `ZEND_AST_DIM`

### 4.4 控制结构解析

**条件语句:**
```yacc
if_stmt:
    T_IF '(' expr ')' statement elseif_list else_single
    
elseif_list:
    elseif_list T_ELSEIF '(' expr ')' statement
    | %empty
```

**循环语句:**
```yacc  
for_statement:
    T_FOR '(' for_exprs ';' for_cond_exprs ';' for_exprs ')' statement
    
foreach_statement:
    T_FOREACH '(' expr T_AS foreach_variable ')' statement
    | T_FOREACH '(' expr T_AS foreach_variable T_DOUBLE_ARROW foreach_variable ')' statement
```

**异常处理:**
```yacc
try_statement:
    T_TRY '{' inner_statement_list '}' catch_list finally_statement
    
catch_list:
    catch_list catch
    | %empty
    
catch:
    T_CATCH '(' name_list optional_variable ')' '{' inner_statement_list '}'
```

## 5. 现代PHP特性支持

### 5.1 Match表达式 (PHP 8.0)
```yacc
match:
    T_MATCH '(' expr ')' '{' match_arm_list '}'
    
match_arm:
    match_arm_cond_list T_DOUBLE_ARROW expr
    | T_DEFAULT T_DOUBLE_ARROW expr
```

### 5.2 属性 (PHP 8.0)  
```yacc
attributes:
    attribute
    | attributes attribute
    
attribute:
    T_ATTRIBUTE attribute_group possible_comma ']'
    
attribute_decl:
    class_name
    | class_name argument_list
```

### 5.3 枚举 (PHP 8.1)
```yacc
enum_declaration_statement:
    optional_attributes enum_modifiers T_ENUM identifier enum_backing_type implements_list '{' enum_case_list '}'
    
enum_case:
    optional_attributes T_CASE identifier enum_case_expr ';'
```

### 5.4 属性钩子 (PHP 8.4)
```yacc
property_hook:
    optional_attributes property_hook_modifiers identifier optional_parameter_list property_hook_body
    
hooked_property:
    property_modifiers optional_type T_VARIABLE '{' optional_property_hook_list '}'
```

### 5.5 联合和交集类型
```yacc
union_type:
    type '|' type
    | union_type '|' type
    
intersection_type:
    type '&' type  
    | intersection_type '&' type
```

## 6. AST构建规则

### 6.1 节点创建模式
```c
// 创建固定子节点数的节点
$$ = zend_ast_create(ZEND_AST_BINARY_OP, $1, $3);

// 创建列表节点
$$ = zend_ast_create_list(1, ZEND_AST_STMT_LIST, $1);

// 添加到列表 
$$ = zend_ast_list_add($1, $3);

// 带属性的节点
$$ = zend_ast_with_attributes($1, $2);
```

### 6.2 类型和属性标记
```c
// 设置节点属性
$$->attr = ZEND_NAME_FQ;                    // 完全限定名
$$->attr = ZEND_SYMBOL_CLASS;               // 符号类型  
$$->attr = ZEND_ACC_PUBLIC;                 // 访问修饰符
$$->attr = ZEND_TYPE_NULLABLE;              // 可空类型
$$->attr = ZEND_ASSIGN_ADD;                 // 复合赋值类型
```

### 6.3 位置信息跟踪
所有AST节点都包含位置信息:
- 行号 (lineno)
- 列号 (col_offset) 
- 文件名引用
- 用于错误报告和调试

## 7. 重构指导原则

基于以上分析，重构PHP parser时应遵循以下原则:

### 7.1 AST节点设计
1. **兼容性**: 节点ID和结构必须与Zend引擎兼容
2. **扩展性**: 支持新的PHP特性和语法
3. **内存效率**: 使用合适的数据结构减少内存占用
4. **类型安全**: 强类型的节点接口设计

### 7.2 解析器架构
1. **递归下降**: 主要使用递归下降解析
2. **优先级解析**: 表达式使用Pratt解析器处理优先级
3. **错误恢复**: 提供完善的语法错误恢复机制
4. **性能优化**: 针对常见语法模式进行优化

### 7.3 词法分析器要求
1. **状态机**: 完整实现所有11个词法状态
2. **字符串处理**: 正确处理各种字符串语法和插值
3. **Unicode支持**: 完整的多字节字符支持
4. **兼容模式**: 支持不同PHP版本的兼容选项

这份完整的语法分析为重构提供了精确的技术规范和实现指导。