# PHP Parser - 待实现功能清单

基于对PHP官方语法文件 `zend_language_parser.y` 的分析，以下是当前parser缺失或不完整的功能。

## 高优先级 (High Priority)

### 1. 内部函数 (Internal Functions) - ✅ 已完成
- [x] `T_ISSET` - isset函数完整语法：`T_ISSET '(' isset_variables possible_comma ')'`
  - 支持单个和多个变量
  - 多个变量正确用AND连接（符合PHP语法规范）
  - 支持可选的尾随逗号
- [x] `T_EMPTY` - empty函数：`T_EMPTY '(' expr ')'`
- [x] `T_INCLUDE` - include语句：`T_INCLUDE expr`
- [x] `T_INCLUDE_ONCE` - include_once语句：`T_INCLUDE_ONCE expr`  
- [x] `T_REQUIRE` - require语句：`T_REQUIRE expr`
- [x] `T_REQUIRE_ONCE` - require_once语句：`T_REQUIRE_ONCE expr`
- [x] `T_EVAL` - eval语句：`T_EVAL '(' expr ')'`

### 2. Spaceship运算符 (<=>) - ✅ 已完成
- [x] 实现 `expr T_SPACESHIP expr` 语法
- [x] 正确的运算符优先级（LESSGREATER级别）
- [x] 支持复杂表达式作为操作数

### 3. T_PIPE表达式 (管道运算符) - ✅ 已完成
- [x] 实现 `expr T_PIPE expr` 语法 (`|>`)
- [x] Lexer支持 `|>` token识别
- [x] 正确的运算符优先级（SUM级别）
- [x] 左结合性（支持链式操作）
- [x] 支持简单函数名和函数调用作为右操作数

### 4. 数组展开语法 (...) - ✅ 已完成
- [x] 数组中的展开语法：`T_ELLIPSIS expr` 在 `array_pair` 中
- [x] 参数中的展开语法：`T_ELLIPSIS expr` 在 `argument` 中
- [x] SpreadExpression AST节点实现
- [x] 支持数组字面量和array()语法中的展开
- [x] 支持函数调用参数的展开语法
- [x] 正确区分第一类可调用语法和展开语法

### 5. 命名参数 (Named Arguments) - ✅ 已完成
- [x] 实现 `identifier ':' expr` 语法
- [x] 在函数调用中支持命名参数
- [x] NamedArgument AST节点实现
- [x] 支持单个和多个命名参数
- [x] 支持位置参数与命名参数混合使用
- [x] 支持复杂表达式作为命名参数值

## 中优先级 (Medium Priority)

### 6. 属性完整支持 (Attributes) - ✅ 已完成
- [x] 完整的属性语法：`attribute_group`, `attribute_list`
- [x] 属性参数列表：`attribute_decl`
- [x] 多个属性组合
- [x] AttributeGroup AST节点实现
- [x] AttributeList AST节点实现
- [x] 支持单个属性组中的多个属性：`#[Attr1, Attr2, ...]`
- [x] 支持带参数的属性：`#[Route("/api", method: "GET")]`
- [x] 支持命名参数在属性中的使用
- [x] 更新现有测试以适应新的AST结构
- [x] 完整的错误处理和边界条件测试

### 7. 匿名类 (Anonymous Classes) - ✅ 已完成
- [x] 基本匿名类语法：`new class {}`
- [x] 带构造参数的匿名类：`new class($arg1, $arg2) {}`
- [x] 匿名类修饰符：`new final class {}`, `new readonly class {}`, `new abstract class {}`
- [x] 继承和接口实现：`new class extends Parent implements Interface1, Interface2 {}`
- [x] 属性支持：`new #[Attribute] class {}`
- [x] 完整类体：属性、方法、构造函数等
- [x] AnonymousClass AST节点实现
- [x] 支持多修饰符组合：`new final readonly class {}`
- [x] 完整的AST结构包含Attributes和Modifiers字段
- [x] 智能匿名类模式检测（isAnonymousClassPattern）
- [x] 11个测试场景覆盖所有功能和边界条件

### 8. Trait适配 (Trait Adaptations) - ✅ 已完成
- [x] `trait_adaptations` - trait使用适配
- [x] `trait_precedence` - insteadof语法：`TraitA::method insteadof TraitB, TraitC`
- [x] `trait_alias` - as语法重命名：`method as newName`, `method as private`, `TraitA::method as public newMethod`
- [x] 完整的trait使用语句解析：`use TraitA, TraitB { ... }`
- [x] TraitMethodReference AST节点支持完全限定和简单方法引用
- [x] TraitPrecedenceStatement和TraitAliasStatement AST节点实现
- [x] 支持多个insteadof traits和复杂适配规则组合
- [x] 8个综合测试用例覆盖所有功能和边界条件

### 9. 保留关键字处理 - ✅ 已完成  
- [x] `reserved_non_modifiers` - 保留非修饰符关键字：支持所有PHP官方保留关键字
- [x] `semi_reserved` - 半保留关键字：包含保留关键字 + 可见性修饰符
- [x] 标识符中的保留字使用：在类常量名、方法名、属性访问中允许保留关键字
- [x] 实现isReservedNonModifier()和isSemiReserved()辅助函数
- [x] 更新类常量声明解析以支持保留关键字作为常量名
- [x] 更新函数/方法声明解析以支持保留关键字作为方法名
- [x] 更新属性访问解析以支持保留关键字作为属性名（$obj->class, $obj?->function）
- [x] 更新trait适配解析以支持保留关键字作为方法引用和别名
- [x] 16个综合测试用例覆盖所有保留关键字使用场景

### 10. First-class Callable - ⚠️ 部分实现
- [ ] 完整的callable语法：`'(' T_ELLIPSIS ')'`
- [ ] 对象方法的first-class callable
- [ ] 静态方法的first-class callable

## 低优先级 (Low Priority)

### 11. 属性钩子 (Property Hooks) - ❌ 未实现
- [ ] `hooked_property` - 带钩子的属性
- [ ] `property_hook_list` - 属性钩子列表
- [ ] `property_hook` - 单个属性钩子
- [ ] `property_hook_body` - 钩子体

### 12. 高级表达式功能 - ❌ 未实现
- [ ] `T_YIELD_FROM` - yield from语法
- [ ] Clone参数列表：`clone_argument_list`
- [ ] Shell执行：backticks语法
- [ ] `T_VOID_CAST` - void转换

### 13. 类型系统增强 - ⚠️ 部分实现
- [ ] 静态类型：`T_STATIC` 在类型表达式中
- [ ] 组合类型括号：`'(' intersection_type ')'`
- [ ] 完整的 `type_expr_without_static`

### 14. 顶层语句完善 - ❌ 未实现
- [ ] `T_HALT_COMPILER` - halt compiler语句
- [ ] 分组Use语句：`group_use_declaration`
- [ ] 混合分组Use：`mixed_group_use_declaration`
- [ ] 内联Use声明：`inline_use_declarations`

### 15. 类成员增强 - ❌ 未实现
- [ ] 构造器参数属性：`optional_cpp_modifiers`
- [ ] 抽象方法声明：method_body为`;`
- [ ] 枚举Case：`enum_case`
- [ ] 枚举后备类型：`enum_backing_type`

### 16. 复杂变量和访问 - ❌ 未实现
- [ ] `variable_class_name` - 变量类名
- [ ] `fully_dereferenceable` - 完全可解引用
- [ ] 动态类常量访问：`class_name T_PAAMAYIM_NEKUDOTAYIM '{' expr '}'`
- [ ] 复杂的字符串插值：完整的 `encaps_var`

### 17. 数组和列表增强 - ❌ 未实现
- [ ] 表达式中的数组解构：`T_LIST '(' array_pair_list ')'`
- [ ] 可变参数完整支持：`is_variadic`
- [ ] 数组引用元素：`ampersand variable` 在数组中

### 18. Match表达式完善 - ⚠️ 部分实现
- [ ] 完整的match语法验证
- [ ] `match_arm_list` - match分支列表
- [ ] `match_arm_cond_list` - match条件列表

## 实现状态说明

- ❌ 未实现：完全缺失该功能
- ⚠️ 部分实现：有基本实现但可能不完整或有问题
- ✅ 已实现：功能完整实现并通过测试

## 实现顺序

按优先级顺序实现，每完成一个功能后：
1. 添加相应的测试用例
2. 更新此TODO.md标记完成状态
3. 提交并推送代码
4. 继续下一个功能

## 测试策略

每个功能实现后都需要：
1. 单元测试覆盖基本用法
2. 边界条件测试
3. 与PHP官方行为对比测试
4. 性能基准测试（如适用）