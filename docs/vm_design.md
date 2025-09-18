# VM 执行系统总体设计

本设计在参考 PHP 官方实现（`php-src/Zend/zend_vm_execute.c`、`Zend/zend_execute.h` 等文件）与主流脚本语言虚拟机结构（Python CPython Eval Loop、Lua VM、Ruby YARV）的基础上，结合项目既有 `compiler`、`ast`、`parser`、`lexer` 模块，对执行期进行完整建模。目标是在 Go 语言中实现一个具备可扩展、可调试、可与统一符号系统协作的 PHP 字节码虚拟机。

## 1. 架构概览

```
┌────────────────────────────────────────────────────────────┐
│                        Application CLI                      │
│        (cmd/hey, cmd/vm-demo, 外部工具)                     │
└──────────────┬──────────────────────────────────────────────┘
               │
               ▼
┌────────────────────────────────────────────────────────────┐
│                        runtime 包                           │
│   - Bootstrap / UnifiedBootstrap                            │
│   - VMIntegration（全局变量、超全局）                        │
│   - registerBuiltinSymbols（内建函数注册）                  │
└──────────────┬──────────────────────────────────────────────┘
               │
               ▼
┌────────────────────────────────────────────────────────────┐
│                       registry 包                            │
│   - 全局统一注册表(GlobalRegistry)                          │
│   - Function / Class / Trait / Interface 描述               │
│   - BuiltinCallContext / BuiltinImplementation              │
└──────────────┬──────────────────────────────────────────────┘
               │
               ▼
┌────────────────────────────────────────────────────────────┐
│                          vm 包                              │
│   - VirtualMachine：执行主循环、调试、性能统计              │
│   - ExecutionContext：执行上下文、变量/常量快照            │
│   - CallFrame：函数调用帧、局部变量、临时变量               │
│   - 指令调度与处理：instructions.go                         │
│   - Profiling & Debug：profiling.go                         │
└──────────────┬──────────────────────────────────────────────┘
               │
               ▼
┌────────────────────────────────────────────────────────────┐
│                      compiler 输出物                        │
│   - opcodes.Instruction 列表                                │
│   - 常量池 []*values.Value                                  │
│   - registry.Function / Class                               │
└────────────────────────────────────────────────────────────┘
```

核心思路：编译器持续输出 PHP 字节码与函数/类描述；运行时通过 registry 统一管理符号；虚拟机使用 ExecutionContext 运行字节码并提供调试、性能、内建函数支持。

## 2. 模块职责

### 2.1 registry
- 线程安全的 `Registry` 全局注册表。
- `Function` 结构兼容用户函数与内建函数(BuiltinImplementation)。
- `Class`、`ClassDescriptor`、`BytecodeMethodImpl` 等描述结构，满足编译器的统一系统需求。
- `BuiltinCallContext` 解耦 vm 与 runtime，实现内建函数在纯 Go 环境下执行。

### 2.2 runtime
- `Bootstrap()` 初始化 registry 并注册内建符号。
- `UnifiedBootstrap()` 兼容旧流程。
- `VMIntegration` 持有需要注入 ExecutionContext 的超全局变量。
- `builtinFunctionSpecs` 列出 Go 实现的内建函数：`count`、`array_*`、`strlen`、`substr`、`strpos`、`print` 等。

### 2.3 vm
- `VirtualMachine`：执行主循环、调试断点、变量监视、性能统计。
- `ExecutionContext`：
  - `CallStack` 执行帧栈；
  - `Variables` / `Temporaries` / `Constants` 快照；
  - `debugLog` 支持调试报告；
  - `GlobalVars` / `IncludedFiles` 与 CLI 协同。
- `CallFrame`：
  - `Locals` / `TempVars` 存储运行期变量；
  - `SlotNames` / `NameSlots` 解决 variable-variable 和 watch 需求；
  - `PendingCall` 协调 `INIT_FCALL`/`SEND_*`/`DO_FCALL` 序列；
  - `ReturnTarget` 指示返回值存放位置。
- 指令处理：
  - 算术、比较、布尔、赋值、跳转、数组、字符串、函数调用、闭包创建、include/require 等核心指令。
  - 未实现指令统一返回 `opcode ... not implemented`，便于增量扩展。
- 调试 & Profiling：`profileState` 统计 IP/Opcode 热点，收集 debug 日志，输出性能报告。

### 2.4 optimizer（占位）
- 提供兼容接口 `OptimizeWithStats`，未来可替换为真正优化器。

## 3. 核心数据结构

```go
// VirtualMachine 关键字段
struct VirtualMachine {
    debugLevel        DebugLevel
    DebugMode         bool
    breakpoints       map[int]struct{}
    watchVars         map[string]struct{}
    profile           *profileState
    advancedProfiling bool
    CompilerCallback  CompilerCallback
    lastContext       *ExecutionContext
}

// ExecutionContext
struct ExecutionContext {
    Stack         []*values.Value
    OutputWriter  io.Writer
    GlobalVars    map[string]*values.Value
    IncludedFiles map[string]bool
    Variables     map[string]*values.Value // 按名字快速查看/共享变量
    Temporaries   map[uint32]*values.Value
    Constants     []*values.Value
    CallStack     []*CallFrame
}

// CallFrame
struct CallFrame {
    Function     *registry.Function
    Instructions []*opcodes.Instruction
    Constants    []*values.Value
    Locals       map[uint32]*values.Value
    TempVars     map[uint32]*values.Value
    SlotNames    map[uint32]string
    NameSlots    map[string]uint32
    PendingCall  *PendingCall
    ReturnTarget operandTarget
}
```

## 4. 指令执行主循环（伪代码）

```text
while true:
    frame = ctx.currentFrame()
    if frame == nil:
        ctx.Halted = true
        break

    if frame.IP out of range:
        handleReturn(ctx, NULL)
        continue

    inst = frame.Instructions[frame.IP]
    profile.observe(frame.IP, inst.Opcode)
    if isBreakpoint(frame.IP):
        recordDebug()

    advance, err = executeInstruction(ctx, frame, inst)
    if err: return decorateError(err)
    if ctx.Halted: break
    if advance: frame.IP++
```

指令执行通过 `switch inst.Opcode` 分发至具体处理函数（`execAssign`、`execArithmetic` 等），通过 `decodeOperand`/`readOperand`/`writeOperand` 完成操作数读取与写入。

## 5. 函数调用流程

1. `OP_INIT_FCALL` 解析被调用实体（字符串函数名、闭包等），构建 `PendingCall`。
2. `OP_SEND_*` 按顺序压入参数。
3. `OP_DO_FCALL`：
   - 若目标为 Go 实现的内建函数，通过 `registry.BuiltinImplementation` 立即执行，写回返回值。
   - 若为用户函数：创建 `CallFrame`，绑定参数，压入 `CallStack`，暂停当前帧。
4. `OP_RETURN`/`OP_RETURN_BY_REF`：
   - 调用 `handleReturn`，出栈当前帧，将返回值写入调用者指定的 `ReturnTarget`。

该流程与 Zend VM 的 `INIT_FCALL`/`SEND_*`/`DO_FCALL` 机制保持一致，方便未来扩展到 `DO_UCALL`、`DO_ICALL`、方法调用等。

## 6. include/require 处理

- `OP_INCLUDE*` 读取文件、调用 `lexer`+`parser` 构造 AST，然后交给 `CompilerCallback`。
- `CompilerCallback` 由 CLI 注入，可复用 `compiler` 生成的字节码并在同一 VM 实例中执行。
- `ExecutionContext.IncludedFiles` 用于实现 `_once` 语义。

## 7. 内建符号系统

- runtime.bootstrap → registry.Module
- 统一在 registry 中维护：
  - `RegisterFunction/Class/Interface/Trait/Constant` 等接口；
  - `ClassDescriptor`/`MethodDescriptor`/`BytecodeMethodImpl` 支持统一类加载。
- 内建函数通过 `BuiltinCallContext` 访问输出流、全局变量、registry。

## 8. 调试 & 性能分析

- `DebugMode`（布尔）与 `DebugLevel`（枚举）双重控制：保持向后兼容。
- `SetBreakpoint` / `WatchVariable` 记录断点、变量观测目标。
- `profileState` 收集指令执行次数、热点 IP/Opcode、调试信息，提供 `GetPerformanceReport`、`GetHotSpots`、`GetDebugReport`、`GetMemoryStats` 接口。

## 9. 扩展点

- 指令覆盖度：当前实现了主要运算、数组、函数、闭包、include 等指令。其余指令保留 `TODO` 错误提示，便于按需实现。
- 内建函数：已实现常用字符串/数组函数，可逐步扩充。
- 类/接口/trait：registry 结构已经与编译器对接，后续可在 VM 中实现方法调用、属性访问、trait 合并等高级特性。
- 并发/插件：ExecutionContext 与 registry 均为线程安全设计，可在未来添加协程支持或插件系统。

## 10. 参考资料

- PHP 官方实现：`php-src/Zend/zend_vm_execute.c`、`zend_execute.h`
- CPython eval loop、Lua VM、Ruby YARV 设计思路
- 项目内 `compiler/compiler.go`、`opcodes/opcodes.go` 对应的指令生成逻辑

上述设计与实现，为已完成的 `compiler/ast/parser/lexer` 模块提供了完整的执行后端，可在保持 API 兼容的前提下逐步完善 PHP 运行时的更多细节。

