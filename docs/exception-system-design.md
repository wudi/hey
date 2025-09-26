# 异常系统重构设计方案

## 问题定义

当前 hey-codex VM 架构中，内置函数返回的 Go `error` 会直接中断执行，不会转换为可被 PHP `try-catch` 捕获的异常对象。

### 当前流程
```go
// vm/instructions.go:3231-3234
ret, err := fn.Builtin(ctxBuiltin, args)
if err != nil {
    return false, err  // ❌ 直接返回，无法被 try-catch 捕获
}
```

### 期望流程
```php
try {
    assert(false);  // 应该抛出 AssertionError
} catch (AssertionError $e) {
    echo "Caught!";  // ✅ 应该能捕获
}
```

---

## PHP Zend VM 异常机制分析

### 核心数据结构
```c
// zend_globals.h
struct _zend_executor_globals {
    zend_object *exception;        // 当前待处理异常
    zend_object *prev_exception;   // 异常链
    // ...
};

// 访问宏
#define EG(v) (executor_globals.v)
```

### 异常抛出流程
```c
// zend_exceptions.c
ZEND_API void zend_throw_exception_internal(zend_object *exception) {
    zend_exception_set_previous(exception, EG(exception));
    EG(exception) = exception;  // 设置全局异常指针

    // VM 在每条指令后检查 EG(exception)
    // 如果非空，触发 HANDLE_EXCEPTION 流程
}
```

### 内置函数抛出异常
```c
// 示例：array_merge 参数错误
ZEND_API zend_object *zend_throw_exception(
    zend_class_entry *exception_ce,  // 异常类
    const char *message,              // 错误消息
    zend_long code                    // 错误码
);

// 使用
if (invalid_argument) {
    zend_throw_exception(zend_ce_type_error, "Invalid type", 0);
    RETURN_THROWS();  // 宏：return
}
```

---

## Hey-Codex 当前架构

### 异常处理流程

**1. throw 语句编译**
```go
// compiler/compiler.go
case *ast.ThrowNode:
    // 编译异常对象创建
    c.compileExpression(node.Exception)
    c.emit(opcodes.OP_THROW, ...)
```

**2. VM 执行 THROW**
```go
// vm/instructions.go:2542
func (vm *VirtualMachine) execThrow(...) (bool, error) {
    val, _ := vm.readOperand(...)
    return vm.raiseException(ctx, frame, val)
}
```

**3. 异常传播**
```go
// vm/vm.go:660
func (vm *VirtualMachine) raiseException(ctx, frame, value) {
    for {
        if handler := frame.popExceptionHandler(); handler != nil {
            frame.pendingException = value  // 设置到 CallFrame
            frame.IP = handler.catchIP      // 跳转到 catch 块
            return false, nil
        }
        ctx.popFrame()  // 向上展开调用栈
        frame = ctx.currentFrame()
    }
}
```

### 问题所在

**内置函数调用点**
```go
// vm/instructions.go:3231
ret, err := fn.Builtin(ctxBuiltin, args)
if err != nil {
    return false, err  // ❌ 问题：直接返回 Go error
}
```

**内置函数无法访问 CallFrame**
```go
// registry/types.go
type BuiltinImplementation func(
    ctx BuiltinCallContext,    // 只有有限接口
    args []*values.Value
) (*values.Value, error)       // 返回的 error 不是 PHP 异常
```

---

## 设计方案

### **方案 A：扩展 BuiltinCallContext（推荐）**

#### 架构改动

**1. 扩展 BuiltinCallContext 接口**
```go
// registry/types.go
type BuiltinCallContext interface {
    // 现有方法
    WriteOutput(val *values.Value) error
    GetGlobal(name string) (*values.Value, bool)
    // ...

    // 新增：异常抛出方法
    ThrowException(exception *values.Value) error
}
```

**2. 实现 ThrowException**
```go
// vm/vm.go
type builtinContext struct {
    vm  *VirtualMachine
    ctx *ExecutionContext
    frame *CallFrame  // 新增：保存当前帧
}

func (bc *builtinContext) ThrowException(exception *values.Value) error {
    _, err := bc.vm.raiseException(bc.ctx, bc.frame, exception)
    if err != nil {
        return err
    }
    // 返回特殊标记 error，告诉 execDoFCall 已处理
    return ErrExceptionThrown
}

var ErrExceptionThrown = errors.New("__EXCEPTION_THROWN__")
```

**3. 修改 execDoFCall**
```go
// vm/instructions.go:3218
ctxBuiltin := &builtinContext{
    vm:    vm,
    ctx:   ctx,
    frame: frame,  // 传入当前帧
}

ret, err := fn.Builtin(ctxBuiltin, args)
if err != nil {
    if errors.Is(err, ErrExceptionThrown) {
        // 异常已设置到 frame.pendingException，继续执行
        return true, nil
    }
    // 其他 error 仍然直接返回
    return false, err
}
```

**4. 修改 assert 函数**
```go
// runtime/assert.go
Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
    // ... 检查断言失败

    // 创建 AssertionError 对象
    exceptionClass := ctx.SymbolRegistry().GetClass("AssertionError")
    exceptionObj := values.NewObject(exceptionClass.Name)
    exceptionObj.SetProperty("message", values.NewString(description))

    // 抛出异常
    return nil, ctx.ThrowException(exceptionObj)
}
```

#### 优点
- ✅ **最小侵入性**：只扩展接口，不改变现有代码
- ✅ **向后兼容**：旧代码继续返回 `error`，新代码选择性使用 `ThrowException`
- ✅ **类型安全**：编译期检查
- ✅ **符合 PHP 语义**：内置函数可以像 `throw` 一样抛出异常

#### 缺点
- ⚠️ 需要内置函数主动调用 `ThrowException`
- ⚠️ 两种错误处理路径（error 和 exception）可能混淆

---

### **方案 B：特殊 Error 类型**

#### 架构改动

**1. 定义 PHPException 类型**
```go
// values/exception.go
type PHPException struct {
    Object *values.Value  // 异常对象
}

func (e *PHPException) Error() string {
    return fmt.Sprintf("PHP Exception: %s", e.Object.ToString())
}
```

**2. 修改 execDoFCall**
```go
ret, err := fn.Builtin(ctxBuiltin, args)
if err != nil {
    var phpEx *PHPException
    if errors.As(err, &phpEx) {
        // 转换为 VM 异常
        return vm.raiseException(ctx, frame, phpEx.Object)
    }
    return false, err  // 普通 error
}
```

**3. 修改 assert 函数**
```go
Builtin: func(ctx, args) (*values.Value, error) {
    // 创建异常对象
    exceptionObj := values.NewObject("AssertionError")
    // ...

    // 返回特殊 error
    return nil, &PHPException{Object: exceptionObj}
}
```

#### 优点
- ✅ 不需要扩展接口
- ✅ 自动转换机制

#### 缺点
- ⚠️ 需要类型断言，不够直观
- ⚠️ 错误处理逻辑隐藏在 error 类型中

---

### **方案 C：修改返回签名（不推荐）**

```go
type BuiltinImplementation func(
    ctx BuiltinCallContext,
    args []*values.Value
) (result *values.Value, exception *values.Value, err error)
```

#### 缺点
- ❌ **破坏性改动**：需要重写所有 65+ 内置函数
- ❌ 三个返回值过于复杂
- ❌ 不符合 Go 惯用法

---

## 推荐实施方案

### **Phase 1: 基础架构（1-2天）**

**Step 1: 定义异常标记**
```go
// errors/exceptions.go
package errors

var ErrExceptionThrown = errors.New("__EXCEPTION_THROWN__")
```

**Step 2: 扩展 BuiltinCallContext**
```go
// registry/types.go
type BuiltinCallContext interface {
    // ... 现有方法

    // ThrowException 抛出 PHP 异常对象
    // 返回 ErrExceptionThrown 标记
    ThrowException(exception *values.Value) error
}
```

**Step 3: 实现 ThrowException**
```go
// vm/vm.go
type builtinContext struct {
    vm    *VirtualMachine
    ctx   *ExecutionContext
    frame *CallFrame  // 新增
}

func (bc *builtinContext) ThrowException(exception *values.Value) error {
    _, err := bc.vm.raiseException(bc.ctx, bc.frame, exception)
    if err != nil {
        return err
    }
    return ErrExceptionThrown
}
```

**Step 4: 修改 execDoFCall**
```go
// vm/instructions.go:3218
ctxBuiltin := &builtinContext{
    vm:    vm,
    ctx:   ctx,
    frame: frame,  // 新增
}

ret, err := fn.Builtin(ctxBuiltin, args)
if err != nil {
    if errors.Is(err, ErrExceptionThrown) {
        return true, nil  // 异常已处理，继续执行
    }
    return false, err
}
```

### **Phase 2: assert 函数重构（1天）**

**Step 1: 创建异常对象辅助函数**
```go
// runtime/exception_helpers.go
func CreateException(ctx registry.BuiltinCallContext, className, message string) *values.Value {
    class := ctx.SymbolRegistry().GetClass(className)
    if class == nil {
        return nil
    }

    obj := values.NewObject(className)
    obj.SetProperty("message", values.NewString(message))
    obj.SetProperty("code", values.NewInt(0))
    obj.SetProperty("file", values.NewString(""))
    obj.SetProperty("line", values.NewInt(0))

    return obj
}
```

**Step 2: 重构 assert**
```go
// runtime/assert.go
Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
    // ... 检查逻辑

    if !assertionResult {
        description := "assert(false)"
        if len(args) > 1 && !args[1].IsNull() {
            description = args[1].ToString()
        }

        exception := CreateException(ctx, "AssertionError", description)
        if exception == nil {
            return nil, fmt.Errorf("AssertionError class not found")
        }

        return nil, ctx.ThrowException(exception)
    }

    return values.NewBool(true), nil
}
```

### **Phase 3: 测试验证（1天）**

**Step 1: 单元测试**
```go
// vm/exception_test.go
func TestBuiltinThrowException(t *testing.T) {
    code := `<?php
    try {
        assert(false);
        echo "NOT REACHED\n";
    } catch (AssertionError $e) {
        echo "Caught\n";
    }`

    output, err := compileAndExecute(t, code)
    require.NoError(t, err)
    assert.Equal(t, "Caught\n", output)
}
```

**Step 2: 集成测试**
```bash
./build/hey -r 'try { assert(false); } catch (AssertionError $e) { echo "OK"; }'
# 期望输出: OK
```

---

## 影响范围评估

### 需要修改的文件
1. `registry/types.go` - 扩展 BuiltinCallContext 接口
2. `vm/vm.go` - 实现 ThrowException，修改 builtinContext
3. `vm/instructions.go` - 修改 execDoFCall 错误处理
4. `errors/exceptions.go` - 新增异常标记常量
5. `runtime/assert.go` - 重构 assert 函数
6. `runtime/exception_helpers.go` - 新增异常创建辅助函数（可选）

### 向后兼容性
- ✅ 所有现有内置函数继续工作
- ✅ 旧代码返回 `error` 仍然有效
- ✅ 无需修改编译器和字节码

### 未来扩展
该架构可支持更多内置函数抛出异常：
- `json_decode()` → `JsonException`
- `file_get_contents()` → `RuntimeException`
- `array_key_exists()` → `TypeError`

---

## 里程碑

| 阶段 | 任务 | 工作量 | 状态 |
|------|------|--------|------|
| Phase 1 | 基础架构实现 | 1-2天 | ⏸ 待开始 |
| Phase 2 | assert 函数重构 | 1天 | ⏸ 待开始 |
| Phase 3 | 测试验证 | 1天 | ⏸ 待开始 |
| **总计** | | **3-4天** | |

---

## 风险与挑战

### 技术风险
1. **并发安全**：`frame.pendingException` 在多线程环境下的访问
   - 缓解：当前 VM 是单线程执行

2. **内存泄漏**：异常对象的生命周期管理
   - 缓解：Go GC 自动管理

3. **性能影响**：每次内置函数调用增加 error 检查
   - 影响：可忽略（仅一次 `errors.Is` 判断）

### 实施风险
1. **测试覆盖不足**：边界情况未考虑
   - 缓解：先实现 assert，再推广到其他函数

2. **文档缺失**：开发者不知如何使用新 API
   - 缓解：在 `registry/types.go` 添加详细注释

---

## 总结

**推荐采用方案 A**，理由：
1. ✅ 最小侵入性，易于实施
2. ✅ 向后兼容，渐进式迁移
3. ✅ 清晰的 API 语义
4. ✅ 3-4天工作量可接受

**Next Steps:**
1. Review 本设计文档
2. 实施 Phase 1 基础架构
3. 验证 assert 函数测试通过
4. 考虑推广到其他内置函数