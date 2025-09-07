package jit

import (
	"fmt"
	"unsafe"

	"github.com/wudi/php-parser/compiler/opcodes"
)

// AMD64CodeGenerator AMD64架构的机器码生成器
type AMD64CodeGenerator struct {
	config *Config

	// 寄存器分配器
	regAllocator *RegisterAllocator

	// 指令缓冲区
	codeBuffer []byte

	// 标签和跳转表
	labels map[string]int
	fixups []JumpFixup
}

// JumpFixup 跳转修复信息
type JumpFixup struct {
	Position int    // 需要修复的位置
	Label    string // 目标标签
	Type     int    // 跳转类型 (相对/绝对)
}

// RegisterAllocator 寄存器分配器
type RegisterAllocator struct {
	// AMD64 通用寄存器
	registers map[string]bool // 寄存器使用状态

	// 寄存器到变量的映射
	regToVar map[string]uint32
	varToReg map[uint32]string
}

// AMD64寄存器定义
const (
	RAX = "rax"
	RBX = "rbx"
	RCX = "rcx"
	RDX = "rdx"
	RSI = "rsi"
	RDI = "rdi"
	RSP = "rsp"
	RBP = "rbp"
	R8  = "r8"
	R9  = "r9"
	R10 = "r10"
	R11 = "r11"
	R12 = "r12"
	R13 = "r13"
	R14 = "r14"
	R15 = "r15"
)

// NewAMD64CodeGenerator 创建新的AMD64代码生成器
func NewAMD64CodeGenerator(config *Config) (*AMD64CodeGenerator, error) {
	regAllocator := &RegisterAllocator{
		registers: make(map[string]bool),
		regToVar:  make(map[string]uint32),
		varToReg:  make(map[uint32]string),
	}

	// 初始化可用寄存器（保留一些系统寄存器）
	availableRegs := []string{RAX, RBX, RCX, RDX, RSI, RDI, R8, R9, R10, R11}
	for _, reg := range availableRegs {
		regAllocator.registers[reg] = false // false = 空闲
	}

	return &AMD64CodeGenerator{
		config:       config,
		regAllocator: regAllocator,
		labels:       make(map[string]int),
	}, nil
}

// GenerateMachineCode 生成AMD64机器码
func (gen *AMD64CodeGenerator) GenerateMachineCode(bytecode []opcodes.Instruction, optimizations []Optimization) (*CompiledFunction, error) {
	// 重置状态
	gen.codeBuffer = nil
	gen.labels = make(map[string]int)
	gen.fixups = nil
	gen.regAllocator.reset()

	// 函数序言
	if err := gen.emitProlog(); err != nil {
		return nil, fmt.Errorf("failed to emit prolog: %v", err)
	}

	// 编译字节码指令
	for i, inst := range bytecode {
		if err := gen.compileInstruction(&inst, i); err != nil {
			return nil, fmt.Errorf("failed to compile instruction %d (%s): %v", i, inst.Opcode.String(), err)
		}
	}

	// 函数尾声
	if err := gen.emitEpilog(); err != nil {
		return nil, fmt.Errorf("failed to emit epilog: %v", err)
	}

	// 修复跳转地址
	if err := gen.fixupJumps(); err != nil {
		return nil, fmt.Errorf("failed to fixup jumps: %v", err)
	}

	// 分配可执行内存
	executableCode, err := gen.allocateExecutableMemory(gen.codeBuffer)
	if err != nil {
		return nil, fmt.Errorf("failed to allocate executable memory: %v", err)
	}

	compiledFunc := &CompiledFunction{
		MachineCode:       executableCode,
		EntryPoint:        uintptr(unsafe.Pointer(&executableCode[0])),
		OptimizationLevel: len(optimizations),
	}

	// 记录优化信息
	for _, opt := range optimizations {
		compiledFunc.OptimizationFlags = append(compiledFunc.OptimizationFlags, opt.Name())
	}

	return compiledFunc, nil
}

// GetTargetArch 返回目标架构
func (gen *AMD64CodeGenerator) GetTargetArch() string {
	return "amd64"
}

// SupportsOpcode 检查是否支持指定的opcode
func (gen *AMD64CodeGenerator) SupportsOpcode(opcode opcodes.Opcode) bool {
	// 目前支持基本的算术和控制流操作
	switch opcode {
	case opcodes.OP_ADD, opcodes.OP_SUB, opcodes.OP_MUL, opcodes.OP_DIV:
		return true
	case opcodes.OP_JMP, opcodes.OP_JMPZ, opcodes.OP_JMPNZ:
		return true
	case opcodes.OP_ASSIGN, opcodes.OP_FETCH_R, opcodes.OP_FETCH_W:
		return true
	case opcodes.OP_RETURN, opcodes.OP_NOP:
		return true
	default:
		return false
	}
}

// compileInstruction 编译单条字节码指令
func (gen *AMD64CodeGenerator) compileInstruction(inst *opcodes.Instruction, index int) error {
	switch inst.Opcode {
	case opcodes.OP_ADD:
		return gen.compileAdd(inst)
	case opcodes.OP_SUB:
		return gen.compileSub(inst)
	case opcodes.OP_MUL:
		return gen.compileMul(inst)
	case opcodes.OP_DIV:
		return gen.compileDiv(inst)
	case opcodes.OP_JMP:
		return gen.compileJmp(inst)
	case opcodes.OP_JMPZ:
		return gen.compileJmpZ(inst)
	case opcodes.OP_JMPNZ:
		return gen.compileJmpNZ(inst)
	case opcodes.OP_ASSIGN:
		return gen.compileAssign(inst)
	case opcodes.OP_RETURN:
		return gen.compileReturn(inst)
	case opcodes.OP_NOP:
		// 空操作，不生成任何代码
		return nil
	default:
		// 不支持的操作，生成函数调用到VM
		return gen.compileVMCall(inst)
	}
}

// 函数序言 - 设置栈帧
func (gen *AMD64CodeGenerator) emitProlog() error {
	// push rbp
	gen.emitByte(0x55)

	// mov rbp, rsp
	gen.emitBytes([]byte{0x48, 0x89, 0xe5})

	// sub rsp, 128  (为局部变量预留空间)
	gen.emitBytes([]byte{0x48, 0x83, 0xec, 0x80})

	return nil
}

// 函数尾声 - 恢复栈帧并返回
func (gen *AMD64CodeGenerator) emitEpilog() error {
	// mov rsp, rbp
	gen.emitBytes([]byte{0x48, 0x89, 0xec})

	// pop rbp
	gen.emitByte(0x5d)

	// ret
	gen.emitByte(0xc3)

	return nil
}

// 基本指令编译方法（简化实现）
func (gen *AMD64CodeGenerator) compileAdd(inst *opcodes.Instruction) error {
	// 这里实现ADD指令的机器码生成
	// 简化版本：生成 add rax, rbx
	gen.emitBytes([]byte{0x48, 0x01, 0xd8})
	return nil
}

func (gen *AMD64CodeGenerator) compileSub(inst *opcodes.Instruction) error {
	// 生成 sub rax, rbx
	gen.emitBytes([]byte{0x48, 0x29, 0xd8})
	return nil
}

func (gen *AMD64CodeGenerator) compileMul(inst *opcodes.Instruction) error {
	// 生成 imul rax, rbx
	gen.emitBytes([]byte{0x48, 0x0f, 0xaf, 0xc3})
	return nil
}

func (gen *AMD64CodeGenerator) compileDiv(inst *opcodes.Instruction) error {
	// 除法需要更复杂的处理，这里简化
	// idiv rbx
	gen.emitBytes([]byte{0x48, 0xf7, 0xfb})
	return nil
}

func (gen *AMD64CodeGenerator) compileJmp(inst *opcodes.Instruction) error {
	// 无条件跳转
	label := fmt.Sprintf("L%d", inst.Op1)
	return gen.emitJump(0xe9, label) // jmp rel32
}

func (gen *AMD64CodeGenerator) compileJmpZ(inst *opcodes.Instruction) error {
	// 零跳转 - test rax, rax; jz label
	gen.emitBytes([]byte{0x48, 0x85, 0xc0}) // test rax, rax

	label := fmt.Sprintf("L%d", inst.Op1)
	return gen.emitJump(0x74, label) // jz rel8 (简化版)
}

func (gen *AMD64CodeGenerator) compileJmpNZ(inst *opcodes.Instruction) error {
	// 非零跳转 - test rax, rax; jnz label
	gen.emitBytes([]byte{0x48, 0x85, 0xc0}) // test rax, rax

	label := fmt.Sprintf("L%d", inst.Op1)
	return gen.emitJump(0x75, label) // jnz rel8 (简化版)
}

func (gen *AMD64CodeGenerator) compileAssign(inst *opcodes.Instruction) error {
	// 简单赋值 - mov 操作
	// mov rax, rbx
	gen.emitBytes([]byte{0x48, 0x89, 0xd8})
	return nil
}

func (gen *AMD64CodeGenerator) compileReturn(inst *opcodes.Instruction) error {
	// 返回指令已经在epilog中处理
	// 这里可以设置返回值
	return nil
}

func (gen *AMD64CodeGenerator) compileVMCall(inst *opcodes.Instruction) error {
	// 对于不支持的指令，生成调用VM解释器的代码
	// 这是一个回退机制

	// 这里需要设置调用约定，传递指令给VM
	// 简化版本：生成一个调用指令
	// call vm_interpreter

	// 目前返回错误表示不支持
	return fmt.Errorf("opcode %s requires VM fallback (not yet implemented)", inst.Opcode.String())
}

// 辅助方法
func (gen *AMD64CodeGenerator) emitByte(b byte) {
	gen.codeBuffer = append(gen.codeBuffer, b)
}

func (gen *AMD64CodeGenerator) emitBytes(bytes []byte) {
	gen.codeBuffer = append(gen.codeBuffer, bytes...)
}

func (gen *AMD64CodeGenerator) emitJump(opcode byte, label string) error {
	// 记录需要修复的跳转
	gen.fixups = append(gen.fixups, JumpFixup{
		Position: len(gen.codeBuffer) + 1,
		Label:    label,
		Type:     1, // 相对跳转
	})

	// 发出跳转指令（地址将稍后修复）
	gen.emitByte(opcode)
	gen.emitBytes([]byte{0x00, 0x00, 0x00, 0x00}) // 占位符

	return nil
}

func (gen *AMD64CodeGenerator) fixupJumps() error {
	// 修复所有跳转地址
	for _, fixup := range gen.fixups {
		targetPos, exists := gen.labels[fixup.Label]
		if !exists {
			return fmt.Errorf("undefined label: %s", fixup.Label)
		}

		// 计算相对偏移
		offset := int32(targetPos - fixup.Position - 4)

		// 写入偏移量（小端序）
		gen.codeBuffer[fixup.Position] = byte(offset)
		gen.codeBuffer[fixup.Position+1] = byte(offset >> 8)
		gen.codeBuffer[fixup.Position+2] = byte(offset >> 16)
		gen.codeBuffer[fixup.Position+3] = byte(offset >> 24)
	}

	return nil
}

// 分配可执行内存（平台相关）
func (gen *AMD64CodeGenerator) allocateExecutableMemory(code []byte) ([]byte, error) {
	// 这里需要使用系统调用分配可执行内存
	// 在实际实现中需要使用mmap或VirtualAlloc

	// 简化版本：直接返回代码（仅用于演示）
	// 注意：这在实际运行中是不安全的
	executableCode := make([]byte, len(code))
	copy(executableCode, code)

	return executableCode, nil
}

// 寄存器分配器方法
func (ra *RegisterAllocator) reset() {
	for reg := range ra.registers {
		ra.registers[reg] = false
	}
	ra.regToVar = make(map[string]uint32)
	ra.varToReg = make(map[uint32]string)
}

func (ra *RegisterAllocator) allocateRegister(varID uint32) (string, error) {
	// 如果变量已经分配了寄存器，返回它
	if reg, exists := ra.varToReg[varID]; exists {
		return reg, nil
	}

	// 寻找空闲寄存器
	for reg, inUse := range ra.registers {
		if !inUse {
			ra.registers[reg] = true
			ra.regToVar[reg] = varID
			ra.varToReg[varID] = reg
			return reg, nil
		}
	}

	// 没有空闲寄存器，需要溢出到内存
	return "", fmt.Errorf("no free registers available, spilling not implemented")
}

func (ra *RegisterAllocator) freeRegister(reg string) {
	if varID, exists := ra.regToVar[reg]; exists {
		ra.registers[reg] = false
		delete(ra.regToVar, reg)
		delete(ra.varToReg, varID)
	}
}
