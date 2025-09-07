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
	// 支持基本的算术和控制流操作
	switch opcode {
	case opcodes.OP_ADD, opcodes.OP_SUB, opcodes.OP_MUL, opcodes.OP_DIV:
		return true
	case opcodes.OP_JMP, opcodes.OP_JMPZ, opcodes.OP_JMPNZ:
		return true
	case opcodes.OP_ASSIGN, opcodes.OP_FETCH_R, opcodes.OP_FETCH_W:
		return true
	case opcodes.OP_RETURN, opcodes.OP_NOP:
		return true
	// 增加更多支持的指令
	case opcodes.OP_BOOL, opcodes.OP_CAST:
		return true
	case opcodes.OP_IS_EQUAL, opcodes.OP_IS_NOT_EQUAL:
		return true
	case opcodes.OP_IS_SMALLER, opcodes.OP_IS_SMALLER_OR_EQUAL:
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
	case opcodes.OP_FETCH_R:
		return gen.compileFetchR(inst)
	case opcodes.OP_FETCH_W:
		return gen.compileFetchW(inst)
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

// 基本指令编译方法（增强实现）
func (gen *AMD64CodeGenerator) compileAdd(inst *opcodes.Instruction) error {
	// 获取操作数寄存器
	srcReg1, err := gen.getOperandRegister(inst.Op1, inst.OpType1)
	if err != nil {
		return fmt.Errorf("failed to get src1 register: %v", err)
	}

	srcReg2, err := gen.getOperandRegister(inst.Op2, inst.OpType2)
	if err != nil {
		return fmt.Errorf("failed to get src2 register: %v", err)
	}

	dstReg, err := gen.allocateResultRegister(inst.Result)
	if err != nil {
		return fmt.Errorf("failed to allocate result register: %v", err)
	}

	// 生成实际的ADD指令
	// mov dst, src1
	if err := gen.emitMov(dstReg, srcReg1); err != nil {
		return err
	}

	// add dst, src2
	if err := gen.emitAdd(dstReg, srcReg2); err != nil {
		return err
	}

	return nil
}

func (gen *AMD64CodeGenerator) compileSub(inst *opcodes.Instruction) error {
	// 获取操作数寄存器
	srcReg1, err := gen.getOperandRegister(inst.Op1, inst.OpType1)
	if err != nil {
		return fmt.Errorf("failed to get src1 register: %v", err)
	}

	srcReg2, err := gen.getOperandRegister(inst.Op2, inst.OpType2)
	if err != nil {
		return fmt.Errorf("failed to get src2 register: %v", err)
	}

	dstReg, err := gen.allocateResultRegister(inst.Result)
	if err != nil {
		return fmt.Errorf("failed to allocate result register: %v", err)
	}

	// 生成SUB指令: dst = src1 - src2
	// mov dst, src1
	if err := gen.emitMov(dstReg, srcReg1); err != nil {
		return err
	}

	// sub dst, src2
	if err := gen.emitSub(dstReg, srcReg2); err != nil {
		return err
	}

	return nil
}

func (gen *AMD64CodeGenerator) compileMul(inst *opcodes.Instruction) error {
	// 获取操作数寄存器
	srcReg1, err := gen.getOperandRegister(inst.Op1, inst.OpType1)
	if err != nil {
		return fmt.Errorf("failed to get src1 register: %v", err)
	}

	srcReg2, err := gen.getOperandRegister(inst.Op2, inst.OpType2)
	if err != nil {
		return fmt.Errorf("failed to get src2 register: %v", err)
	}

	dstReg, err := gen.allocateResultRegister(inst.Result)
	if err != nil {
		return fmt.Errorf("failed to allocate result register: %v", err)
	}

	// 生成MUL指令: dst = src1 * src2
	// mov dst, src1
	if err := gen.emitMov(dstReg, srcReg1); err != nil {
		return err
	}

	// imul dst, src2
	if err := gen.emitIMul(dstReg, srcReg2); err != nil {
		return err
	}

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
	// 使用真正的可执行内存分配
	execMem, err := AllocateExecutableMemory(len(code))
	if err != nil {
		return nil, fmt.Errorf("failed to allocate executable memory: %v", err)
	}

	// 将机器码写入可执行内存
	err = execMem.WriteBytes(0, code)
	if err != nil {
		execMem.Free()
		return nil, fmt.Errorf("failed to write machine code: %v", err)
	}

	// 返回可执行内存的数据切片
	return execMem.Data, nil
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

// 增强的指令编码方法

// getOperandRegister 获取操作数对应的寄存器
func (gen *AMD64CodeGenerator) getOperandRegister(operand uint32, opType byte) (string, error) {
	switch opcodes.OpType(opType) {
	case opcodes.IS_VAR:
		// 变量 - 分配寄存器
		return gen.regAllocator.allocateRegister(operand)
	case opcodes.IS_TMP_VAR:
		// 临时变量 - 分配寄存器
		return gen.regAllocator.allocateRegister(operand + 1000) // 偏移避免冲突
	case opcodes.IS_CONST:
		// 常量 - 加载到寄存器
		reg, err := gen.regAllocator.allocateRegister(operand + 2000) // 偏移避免冲突
		if err != nil {
			return "", err
		}
		// 加载常量到寄存器
		err = gen.emitLoadConstant(reg, int64(operand))
		return reg, err
	default:
		return "", fmt.Errorf("unsupported operand type: %d", opType)
	}
}

// allocateResultRegister 为结果分配寄存器
func (gen *AMD64CodeGenerator) allocateResultRegister(result uint32) (string, error) {
	return gen.regAllocator.allocateRegister(result + 3000) // 偏移避免冲突
}

// 具体指令编码方法

// emitMov 生成MOV指令 (mov dst, src)
func (gen *AMD64CodeGenerator) emitMov(dst, src string) error {
	dstCode := gen.getRegisterCode(dst)
	srcCode := gen.getRegisterCode(src)

	if dstCode == -1 || srcCode == -1 {
		return fmt.Errorf("invalid register: dst=%s, src=%s", dst, src)
	}

	// REX.W + ModR/M
	gen.emitByte(0x48)                                        // REX.W
	gen.emitByte(0x89)                                        // MOV r/m64, r64
	gen.emitByte(0xc0 | (byte(srcCode) << 3) | byte(dstCode)) // ModR/M

	return nil
}

// emitAdd 生成ADD指令 (add dst, src)
func (gen *AMD64CodeGenerator) emitAdd(dst, src string) error {
	dstCode := gen.getRegisterCode(dst)
	srcCode := gen.getRegisterCode(src)

	if dstCode == -1 || srcCode == -1 {
		return fmt.Errorf("invalid register: dst=%s, src=%s", dst, src)
	}

	// REX.W + ADD
	gen.emitByte(0x48)                                        // REX.W
	gen.emitByte(0x01)                                        // ADD r/m64, r64
	gen.emitByte(0xc0 | (byte(srcCode) << 3) | byte(dstCode)) // ModR/M

	return nil
}

// emitSub 生成SUB指令 (sub dst, src)
func (gen *AMD64CodeGenerator) emitSub(dst, src string) error {
	dstCode := gen.getRegisterCode(dst)
	srcCode := gen.getRegisterCode(src)

	if dstCode == -1 || srcCode == -1 {
		return fmt.Errorf("invalid register: dst=%s, src=%s", dst, src)
	}

	// REX.W + SUB
	gen.emitByte(0x48)                                        // REX.W
	gen.emitByte(0x29)                                        // SUB r/m64, r64
	gen.emitByte(0xc0 | (byte(srcCode) << 3) | byte(dstCode)) // ModR/M

	return nil
}

// emitIMul 生成IMUL指令 (imul dst, src)
func (gen *AMD64CodeGenerator) emitIMul(dst, src string) error {
	dstCode := gen.getRegisterCode(dst)
	srcCode := gen.getRegisterCode(src)

	if dstCode == -1 || srcCode == -1 {
		return fmt.Errorf("invalid register: dst=%s, src=%s", dst, src)
	}

	// REX.W + IMUL
	gen.emitByte(0x48)                                        // REX.W
	gen.emitByte(0x0f)                                        // Two-byte opcode prefix
	gen.emitByte(0xaf)                                        // IMUL r64, r/m64
	gen.emitByte(0xc0 | (byte(dstCode) << 3) | byte(srcCode)) // ModR/M

	return nil
}

// emitLoadConstant 加载常量到寄存器
func (gen *AMD64CodeGenerator) emitLoadConstant(reg string, value int64) error {
	regCode := gen.getRegisterCode(reg)
	if regCode == -1 {
		return fmt.Errorf("invalid register: %s", reg)
	}

	// MOV r64, imm64
	gen.emitByte(0x48)                 // REX.W
	gen.emitByte(0xb8 + byte(regCode)) // MOV r64, imm64

	// 小端序64位立即数
	gen.emitByte(byte(value))
	gen.emitByte(byte(value >> 8))
	gen.emitByte(byte(value >> 16))
	gen.emitByte(byte(value >> 24))
	gen.emitByte(byte(value >> 32))
	gen.emitByte(byte(value >> 40))
	gen.emitByte(byte(value >> 48))
	gen.emitByte(byte(value >> 56))

	return nil
}

// getRegisterCode 获取寄存器的编码
func (gen *AMD64CodeGenerator) getRegisterCode(reg string) int {
	switch reg {
	case RAX:
		return 0
	case RCX:
		return 1
	case RDX:
		return 2
	case RBX:
		return 3
	case RSP:
		return 4
	case RBP:
		return 5
	case RSI:
		return 6
	case RDI:
		return 7
	case R8:
		return 0 // R8-R15 need REX.B
	case R9:
		return 1
	case R10:
		return 2
	case R11:
		return 3
	case R12:
		return 4
	case R13:
		return 5
	case R14:
		return 6
	case R15:
		return 7
	default:
		return -1
	}
}

// compileFetchR 编译FETCH_R指令
func (gen *AMD64CodeGenerator) compileFetchR(inst *opcodes.Instruction) error {
	// FETCH_R 是读取变量值
	srcReg, err := gen.getOperandRegister(inst.Op1, inst.OpType1)
	if err != nil {
		return fmt.Errorf("failed to get source register: %v", err)
	}

	dstReg, err := gen.allocateResultRegister(inst.Result)
	if err != nil {
		return fmt.Errorf("failed to allocate result register: %v", err)
	}

	// 简单的移动操作
	return gen.emitMov(dstReg, srcReg)
}

// compileFetchW 编译FETCH_W指令
func (gen *AMD64CodeGenerator) compileFetchW(inst *opcodes.Instruction) error {
	// FETCH_W 是获取变量引用用于写入
	// 这里简化处理，与 FETCH_R 相同
	return gen.compileFetchR(inst)
}
