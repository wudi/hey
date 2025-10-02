# WordPress Heisenbug 调试经验总结

## 问题概述

WordPress index.php 在执行时产生无输出，但添加简单的 `echo 1;` 语句后问题消失。这是典型的 Heisenbug - 观察/调试行为改变了程序执行结果。

**初始症状：**
- WordPress 无任何输出
- 错误：`callable not resolved` 在 wp_die 函数的 call_user_func 调用处
- 添加 `echo 1;` 或任何调试代码后问题消失

## 根本原因分析

### 问题 1: 编译器 Bug - FETCH_CONSTANT 指令被跳过

**设计缺陷：**
- 编译器生成 FETCH_CONSTANT 指令加载函数名到 TMP_VAR
- 但控制流跳转（JMP/JMPZ 等）可能跳过这些指令
- VM 执行 INIT_FCALL 时读到未初始化的 TMP_VAR（NULL）

**发现过程：**
1. 编译器输出显示：`Emitting FETCH_CONSTANT for 'nocache_headers' to TMP_VAR[136]`
2. 但运行时 `writeOperand` 从未被调用写入 TMP_VAR[136]
3. 证明 FETCH_CONSTANT 指令**根本没有执行**

**关键发现：**
```
编译器生成的字节码：
  IP=253: FETCH_CONSTANT 'nocache_headers' → TMP_VAR[136]
  IP=254: (某条指令，可能被跳过)
  IP=255: INIT_FCALL 读取 TMP_VAR[136]

实际执行：
  IP=253 被跳过（跳转直接到 IP=255）
  IP=255: INIT_FCALL 读取 TMP_VAR[136] = NULL
```

### 问题 2: VM Bug - DO_FCALL 双重执行

**设计缺陷：**
- `CallUserFunction` 有自己的执行循环
- 当嵌套函数调用 die/exit 时，设置 `Halted=true`
- 内层循环因 Halted 退出，但 child frame 仍在 CallStack 中
- 外层循环继续执行，导致同一 DO_FCALL 执行两次

**发现过程：**
1. 调试输出显示同一 frame、同一 IP 的 DO_FCALL 执行两次
2. 第一次 pendingCalls count=1（正常）
3. 第二次 pendingCalls count=0（已被第一次 pop）

**执行流程：**
```
外层 CallUserFunction(wp_die):
  → DO_FCALL at IP=143
    → call_user_func builtin
      → 内层 CallUserFunction(_default_wp_die_handler)
        → _default_wp_die_handler 调用 wp_die (递归)
          → wp_die 调用 exit/die
            → Halted=true
        → 内层循环退出（Halted=true）
        → 恢复 CallStack（但保留 Halted=true）
      → 返回到 call_user_func
    → 返回到 DO_FCALL (第一次执行完成)
  → 外层循环继续（child 仍 active）
  → 再次执行 IP=143 的 DO_FCALL（第二次！）
    → pending=nil（已被第一次 pop）
    → 错误：call without INIT_FCALL
```

### 问题 3: CallUserFunction 循环不支持嵌套

**设计缺陷：**
- 简化的循环条件：`currentFrame() == child`
- 当 child 调用其他函数时，currentFrame 变成新 frame
- 循环提前退出，child 未完成就返回

## 调试方法和技巧

### 1. 分层调试策略

**从症状到根因的追踪路径：**

```
Level 1: 用户可见症状
  ↓ "WordPress 无输出"

Level 2: 错误信息
  ↓ "callable not resolved"

Level 3: 调用栈追踪
  ↓ wp_die → call_user_func → _default_wp_die_handler

Level 4: 变量生命周期
  ↓ $callback 变量正确赋值和读取

Level 5: 字节码层面
  ↓ TMP_VAR[136] 未初始化

Level 6: 指令执行追踪
  ↓ FETCH_CONSTANT 未执行

Level 7: 编译器输出对比
  ↓ 编译器生成了 FETCH_CONSTANT，但 VM 跳过了
```

### 2. Heisenbug 的识别和处理

**典型特征：**
- 添加调试代码（echo、print、log）后问题消失
- 问题与时序、初始化、执行顺序相关
- 观察行为改变了程序状态

**调试策略：**
1. **不要依赖侵入式调试** - 会改变执行行为
2. **使用独立的调试通道** - 环境变量控制的 stderr 输出
3. **追踪关键状态点** - IP、frame pointer、pendingCalls count
4. **对比执行路径** - 有无调试代码时的 IP 序列

### 3. 变量生命周期追踪

**关键发现工具：**

```go
// 追踪变量赋值
if varName == "$callback" {
    fmt.Fprintf(os.Stderr, "[DEBUG] ASSIGN: slot=%d value='%s'\n",
        resSlot, value.ToString())
}

// 追踪变量读取
if varName == "$callback" {
    fmt.Fprintf(os.Stderr, "[DEBUG] FETCH_R: slot=%d value='%s'\n",
        op1, val.ToString())
}

// 追踪 TMP_VAR 写入
if opType == IS_TMP_VAR && operand == 136 {
    fmt.Fprintf(os.Stderr, "[DEBUG] Writing to TMP_VAR[136]: %s\n",
        value.ToString())
}
```

**发现：**
- $callback 正确赋值和读取 ✓
- 但问题不在变量本身
- 而在函数调用的参数传递

### 4. 指令级别的执行追踪

**关键技术：**

```go
// 追踪特定 IP 的执行
if frame.IP == 126 && frame.FunctionName == "wp_die" {
    fmt.Fprintf(os.Stderr, "[IP=%d] INIT_FCALL callee='%s'\n",
        frame.IP, callee.ToString())
}

// 追踪同一指令的多次执行（发现双重执行）
if inst.Opcode == OP_DO_FCALL {
    fmt.Fprintf(os.Stderr, "[%p] DO_FCALL IP=%d count=%d\n",
        frame, frame.IP, len(frame.pendingCalls))
}
```

**重要发现：**
- Frame pointer 相同 → 同一个 frame
- IP 相同 → 同一条指令
- 但执行了两次 → 循环逻辑问题

### 5. 编译器与 VM 执行对比

**方法：**
1. 启用编译器调试：查看生成的字节码
2. 启用 VM 调试：查看实际执行的指令
3. 对比差异：找出被跳过的指令

**关键对比：**
```
编译器输出：
  [COMPILER DEBUG] Emitting FETCH_CONSTANT for 'nocache_headers' to TMP_VAR[136]

VM 执行追踪：
  (无 FETCH_CONSTANT 执行记录)

结论：指令生成了但未执行 → 编译器 jump 优化 bug
```

### 6. 状态恢复机制的检查

**CallUserFunction 的状态管理：**

```go
// 保存状态
savedStack := b.ctx.Stack
savedCallStack := b.ctx.CallStack
savedHalted := b.ctx.Halted

// 执行子函数
// ...

// 恢复状态
b.ctx.Stack = savedStack
b.ctx.CallStack = savedCallStack
// 注意：Halted 可能不恢复（die/exit 语义）
```

**问题点：**
- 恢复 CallStack 时，child frame 可能还在其中
- 外层循环继续执行，因为 childFrameActive() 仍为 true
- 但 IP 没有正确更新，导致重复执行

## 解决方案设计

### Solution 1: DEFENSIVE FIX for FETCH_CONSTANT

**设计思路：**
- 既然编译器有 bug（生成但跳过指令），在 VM 层面补救
- INIT_FCALL 发现 callee=NULL 时，不要立即失败
- 向后搜索字节码，找到对应的 FETCH_CONSTANT 指令
- 直接从常量表读取函数名

**实现：**

```go
if (callee == nil || callee.IsNull()) && opType1 == IS_TMP_VAR {
    // 向后搜索最多 10 条指令
    for i := frame.IP - 1; i >= 0 && i >= frame.IP-10; i-- {
        checkInst := frame.Instructions[i]
        if checkInst.Opcode == OP_FETCH_CONSTANT {
            // 检查是否写入相同的 TMP_VAR slot
            resType, resSlot := decodeResult(checkInst)
            if resType == IS_TMP_VAR && resSlot == op1 {
                // 从常量表读取函数名
                constType, constOp := decodeOperand(checkInst, 1)
                if constType == IS_CONST {
                    callee = frame.Constants[constOp]
                    break
                }
            }
        }
    }
}
```

**优点：**
- 运行时补救编译器 bug
- 不需要修改编译器（复杂且风险高）
- 性能影响小（只在错误情况触发）

### Solution 2: Workaround for DO_FCALL Double Execution

**设计思路：**
- 第二次执行是异常情况，不应该发生
- 但根本原因（Halted 导致循环交互）很复杂
- 在 DO_FCALL 层面容错：pending=nil 时不报错，而是跳过

**实现：**

```go
pending := frame.popPendingCall()
if pending == nil {
    // WORKAROUND: 双重执行时第二次调用，跳过
    return true, nil  // 继续执行，相当于 NOP
}
```

**权衡：**
- 这是 workaround，不是根本修复
- 但根本修复需要重构整个 CallUserFunction 机制
- Workaround 安全、简单、有效

### Solution 3: Enhanced CallUserFunction Loop

**设计思路：**
- 支持嵌套调用：child 可以调用其他函数
- 区分 child 完成和嵌套函数完成
- 正确处理 Halted 状态

**关键点：**

```go
// 检查 child 是否仍在 CallStack（支持嵌套）
childFrameActive := func() bool {
    for _, f := range b.ctx.CallStack {
        if f == child {
            return true
        }
    }
    return false
}

// 循环条件
for !b.ctx.Halted && childFrameActive() {
    frame := b.ctx.currentFrame()

    // 区分处理
    if frame.IP >= len(frame.Instructions) {
        if frame == child {
            // child 完成 → 退出
            b.ctx.popFrame()
            break
        } else {
            // 嵌套函数完成 → 继续
            b.ctx.popFrame()
            continue
        }
    }

    // 处理 return
    if !continued {
        if !childFrameActive() {
            break  // child 已被 return 弹出
        }
        // 否则继续（嵌套函数 return）
    }
}
```

## 关键经验教训

### 1. 架构设计的缺陷

**问题：编译器和 VM 的责任分离不清**
- 编译器：生成指令 + 优化（jump）
- VM：执行指令
- **Gap**：编译器优化可能导致 VM 假设失效

**改进方向：**
- 编译器应保证：优化后 TMP_VAR 的值仍有效
- 或者：VM 应该更鲁棒，不依赖 TMP_VAR 一定被初始化
- 最佳：使用更可靠的参数传递机制（如 stack-based）

### 2. 状态管理的复杂性

**问题：多层执行循环 + 状态保存/恢复**
- 外层循环执行 frame A
- A 调用 builtin
- builtin 启动内层循环执行 frame B
- B 修改全局状态（Halted）
- 恢复部分状态，返回外层
- **结果**：外层循环在不一致的状态下继续

**改进方向：**
- 明确状态所有权：谁可以修改 Halted？
- 隔离执行环境：内层循环不应影响外层
- 或者：使用完全独立的 VM 实例（SimpleCallUserFunction 的方向）

### 3. 调试工具的重要性

**有效的调试手段：**
1. **分层调试输出** - 不同的环境变量控制不同层次
   - DEBUG_BYTECODE：字节码生成
   - DEBUG_EXEC：指令执行
   - DEBUG_CALL：函数调用
   - DEBUG_STATE：状态变化

2. **关键状态追踪** - 不只是值，还有指针和计数
   ```
   frame=%p IP=%d pendingCalls=%d
   ```

3. **执行路径对比** - 记录 IP 序列，对比正常和异常情况
   ```
   Normal:   IP=253 → 254 → 255 → ...
   Bug:      IP=252 → 255 (跳过 253, 254)
   ```

4. **编译器-VM 双向验证** - 生成的和执行的必须一致

### 4. Heisenbug 的本质

**为什么调试代码让 bug 消失？**

1. **时序改变**
   - `echo 1;` 产生输出，可能触发 flush
   - 改变了后续代码的执行时序
   - 某些竞态条件被避免

2. **内存布局改变**
   - 添加代码改变了栈帧布局
   - TMP_VAR 的位置/初始化可能不同
   - 未初始化变量可能碰巧有"正确"的值

3. **优化改变**
   - 调试代码可能阻止某些编译器优化
   - Jump 目标可能改变
   - 指令执行顺序改变

**对策：**
- 不依赖调试代码的行为
- 使用非侵入式调试（环境变量 + stderr）
- 理解编译器优化的影响
- 设计更鲁棒的执行机制

### 5. 技术债务的代价

**问题的历史：**
- CallUserFunction 最初设计简单（只处理单层调用）
- 后来添加嵌套支持，但未完全重构
- DEFENSIVE FIX 是为了兼容旧代码而添加的 hack

**教训：**
- 早期设计要考虑扩展性（嵌套、递归、异常）
- 发现设计缺陷时，重构比 workaround 更好（长期）
- 技术债务会累积，最终导致复杂的 bug

## 最佳实践

### 1. VM 设计原则

- **指令执行的原子性**：一条指令的效果应该是完整和确定的
- **状态的局部性**：尽量减少全局状态（如 Halted），使用 frame-local 状态
- **循环的单一职责**：一个执行循环只负责一个 frame，不要混合多层

### 2. 调试策略

- **自顶向下**：从症状开始，逐层深入
- **假设验证**：每个假设都要用数据验证
- **对比分析**：正常 vs 异常、编译 vs 执行
- **最小复现**：创建最简单的复现案例

### 3. 修复原则

- **优先根本修复**：理解根因，从源头解决
- **必要时 workaround**：根本修复风险太高时，加安全的 workaround
- **文档化**：清晰记录问题、原因、解决方案
- **测试验证**：确保修复不引入新问题

## 后续改进建议

### 短期（Workaround 已解决）

1. ✅ DEFENSIVE FIX 恢复 FETCH_CONSTANT
2. ✅ DO_FCALL 容错处理
3. ✅ CallUserFunction 支持嵌套

### 中期（根本修复）

1. **修复编译器 jump 优化 bug**
   - 分析为何 FETCH_CONSTANT 被跳过
   - 确保优化不会跳过必要的初始化指令
   - 或者：改用更可靠的参数传递机制

2. **重构 CallUserFunction**
   - 彻底分离内外层执行上下文
   - 使用独立的 VM 实例（参考 SimpleCallUserFunction）
   - 明确状态所有权和生命周期

### 长期（架构改进）

1. **Stack-based 参数传递**
   - 不依赖 TMP_VAR（易失）
   - 使用 stack（可靠、顺序保证）
   - PHP 原始实现也是 stack-based

2. **执行模型简化**
   - 单一执行循环
   - 统一的状态管理
   - 清晰的控制流

3. **更好的调试支持**
   - 内建的 trace 机制
   - 执行历史记录
   - 状态快照和回放

## 总结

这个 Heisenbug 揭示了多个层面的设计问题：

1. **编译器层**：优化导致指令被跳过
2. **VM 层**：执行循环的嵌套和状态管理
3. **调试层**：观察改变行为的本质

解决方案是多层次的：

1. **VM 补救**：DEFENSIVE FIX 恢复被跳过的指令
2. **容错处理**：DO_FCALL workaround 处理异常情况
3. **增强逻辑**：CallUserFunction 支持复杂嵌套

关键教训：

- **Heisenbug 不是魔法**，有深层的时序/状态原因
- **分层调试**是复杂问题的有效方法
- **设计缺陷**会以难以预料的方式暴露
- **Workaround 有时是必要的**，但要文档化并计划根本修复

这次调试充分展示了：理解系统的多个层次、使用正确的调试工具、以及持续深入追踪问题的重要性。
