package passes

import (
	"github.com/wudi/php-parser/compiler/opcodes"
	"github.com/wudi/php-parser/compiler/values"
)

// OptimizationPass represents a single optimization pass
type OptimizationPass interface {
	Name() string
	Optimize(instructions []opcodes.Instruction, constants []*values.Value) ([]opcodes.Instruction, []*values.Value, bool)
}

// Optimizer manages multiple optimization passes
type Optimizer struct {
	passes []OptimizationPass
}

// NewOptimizer creates a new optimizer with default passes
func NewOptimizer() *Optimizer {
	optimizer := &Optimizer{
		passes: []OptimizationPass{
			&ConstantFoldingPass{},
			&DeadCodeEliminationPass{},
			&PeepholeOptimizationPass{},
			&JumpOptimizationPass{},
			&TemporaryEliminationPass{},
		},
	}
	return optimizer
}

// AddPass adds an optimization pass
func (o *Optimizer) AddPass(pass OptimizationPass) {
	o.passes = append(o.passes, pass)
}

// Optimize applies all optimization passes until no more changes occur
func (o *Optimizer) Optimize(instructions []opcodes.Instruction, constants []*values.Value) ([]opcodes.Instruction, []*values.Value) {
	currentInstructions := make([]opcodes.Instruction, len(instructions))
	copy(currentInstructions, instructions)

	currentConstants := make([]*values.Value, len(constants))
	copy(currentConstants, constants)

	totalChanges := 0
	iteration := 0
	maxIterations := 10 // Prevent infinite loops

	for iteration < maxIterations {
		iterationChanges := 0

		for _, pass := range o.passes {
			optimized, newConstants, changed := pass.Optimize(currentInstructions, currentConstants)
			if changed {
				currentInstructions = optimized
				currentConstants = newConstants
				iterationChanges++
			}
		}

		if iterationChanges == 0 {
			break // No more optimizations possible
		}

		totalChanges += iterationChanges
		iteration++
	}

	return currentInstructions, currentConstants
}

// ConstantFoldingPass performs compile-time constant evaluation
type ConstantFoldingPass struct{}

func (p *ConstantFoldingPass) Name() string {
	return "ConstantFolding"
}

func (p *ConstantFoldingPass) Optimize(instructions []opcodes.Instruction, constants []*values.Value) ([]opcodes.Instruction, []*values.Value, bool) {
	optimized := make([]opcodes.Instruction, 0, len(instructions))
	newConstants := make([]*values.Value, len(constants))
	copy(newConstants, constants)

	changed := false

	for i := 0; i < len(instructions); i++ {
		inst := instructions[i]

		// Look for arithmetic operations on constants
		if p.isArithmeticOp(inst.Opcode) && p.areBothConstants(inst) {
			// Perform constant folding
			if folded, newConst := p.foldConstantOperation(inst, constants); folded {
				// Replace with a simple constant load
				constIndex := uint32(len(newConstants))
				newConstants = append(newConstants, newConst)

				opType1, opType2 := opcodes.EncodeOpTypes(opcodes.IS_CONST, opcodes.IS_UNUSED, opcodes.DecodeResultType(inst.OpType2))
				optimized = append(optimized, opcodes.Instruction{
					Opcode:  opcodes.OP_QM_ASSIGN,
					OpType1: opType1,
					OpType2: opType2,
					Op1:     constIndex,
					Op2:     0,
					Result:  inst.Result,
				})
				changed = true
				continue
			}
		}

		optimized = append(optimized, inst)
	}

	return optimized, newConstants, changed
}

func (p *ConstantFoldingPass) isArithmeticOp(opcode opcodes.Opcode) bool {
	switch opcode {
	case opcodes.OP_ADD, opcodes.OP_SUB, opcodes.OP_MUL, opcodes.OP_DIV, opcodes.OP_MOD, opcodes.OP_POW:
		return true
	case opcodes.OP_CONCAT:
		return true
	case opcodes.OP_IS_EQUAL, opcodes.OP_IS_NOT_EQUAL, opcodes.OP_IS_IDENTICAL, opcodes.OP_IS_NOT_IDENTICAL:
		return true
	case opcodes.OP_IS_SMALLER, opcodes.OP_IS_SMALLER_OR_EQUAL, opcodes.OP_IS_GREATER, opcodes.OP_IS_GREATER_OR_EQUAL:
		return true
	case opcodes.OP_BOOLEAN_AND, opcodes.OP_BOOLEAN_OR:
		return true
	case opcodes.OP_BW_AND, opcodes.OP_BW_OR, opcodes.OP_BW_XOR, opcodes.OP_SL, opcodes.OP_SR:
		return true
	default:
		return false
	}
}

func (p *ConstantFoldingPass) areBothConstants(inst opcodes.Instruction) bool {
	op1Type := opcodes.DecodeOpType1(inst.OpType1)
	op2Type := opcodes.DecodeOpType2(inst.OpType1)
	return op1Type == opcodes.IS_CONST && op2Type == opcodes.IS_CONST
}

func (p *ConstantFoldingPass) foldConstantOperation(inst opcodes.Instruction, constants []*values.Value) (bool, *values.Value) {
	if int(inst.Op1) >= len(constants) || int(inst.Op2) >= len(constants) {
		return false, nil
	}

	op1 := constants[inst.Op1]
	op2 := constants[inst.Op2]

	switch inst.Opcode {
	case opcodes.OP_ADD:
		return true, op1.Add(op2)
	case opcodes.OP_SUB:
		return true, op1.Subtract(op2)
	case opcodes.OP_MUL:
		return true, op1.Multiply(op2)
	case opcodes.OP_DIV:
		if op2.ToFloat() == 0.0 {
			return false, nil // Don't fold division by zero
		}
		return true, op1.Divide(op2)
	case opcodes.OP_MOD:
		if op2.ToInt() == 0 {
			return false, nil // Don't fold modulo by zero
		}
		return true, op1.Modulo(op2)
	case opcodes.OP_POW:
		return true, op1.Power(op2)
	case opcodes.OP_CONCAT:
		return true, op1.Concat(op2)
	case opcodes.OP_IS_EQUAL:
		return true, values.NewBool(op1.Equal(op2))
	case opcodes.OP_IS_NOT_EQUAL:
		return true, values.NewBool(!op1.Equal(op2))
	case opcodes.OP_IS_IDENTICAL:
		return true, values.NewBool(op1.Identical(op2))
	case opcodes.OP_IS_NOT_IDENTICAL:
		return true, values.NewBool(!op1.Identical(op2))
	case opcodes.OP_IS_SMALLER:
		return true, values.NewBool(op1.Compare(op2) < 0)
	case opcodes.OP_IS_SMALLER_OR_EQUAL:
		return true, values.NewBool(op1.Compare(op2) <= 0)
	case opcodes.OP_IS_GREATER:
		return true, values.NewBool(op1.Compare(op2) > 0)
	case opcodes.OP_IS_GREATER_OR_EQUAL:
		return true, values.NewBool(op1.Compare(op2) >= 0)
	case opcodes.OP_BOOLEAN_AND:
		return true, values.NewBool(op1.ToBool() && op2.ToBool())
	case opcodes.OP_BOOLEAN_OR:
		return true, values.NewBool(op1.ToBool() || op2.ToBool())
	case opcodes.OP_BW_AND:
		return true, values.NewInt(op1.ToInt() & op2.ToInt())
	case opcodes.OP_BW_OR:
		return true, values.NewInt(op1.ToInt() | op2.ToInt())
	case opcodes.OP_BW_XOR:
		return true, values.NewInt(op1.ToInt() ^ op2.ToInt())
	case opcodes.OP_SL:
		return true, values.NewInt(op1.ToInt() << uint(op2.ToInt()))
	case opcodes.OP_SR:
		return true, values.NewInt(op1.ToInt() >> uint(op2.ToInt()))
	default:
		return false, nil
	}
}

// DeadCodeEliminationPass removes unreachable code
type DeadCodeEliminationPass struct{}

func (p *DeadCodeEliminationPass) Name() string {
	return "DeadCodeElimination"
}

func (p *DeadCodeEliminationPass) Optimize(instructions []opcodes.Instruction, constants []*values.Value) ([]opcodes.Instruction, []*values.Value, bool) {
	// Build reachability graph
	reachable := make([]bool, len(instructions))
	p.markReachable(instructions, reachable, 0)

	// Remove unreachable instructions
	optimized := make([]opcodes.Instruction, 0, len(instructions))
	changed := false

	for i, inst := range instructions {
		if reachable[i] {
			optimized = append(optimized, inst)
		} else {
			changed = true
		}
	}

	return optimized, constants, changed
}

func (p *DeadCodeEliminationPass) markReachable(instructions []opcodes.Instruction, reachable []bool, start int) {
	if start >= len(instructions) || reachable[start] {
		return
	}

	reachable[start] = true
	inst := instructions[start]

	switch inst.Opcode {
	case opcodes.OP_JMP:
		// Unconditional jump
		if opcodes.DecodeOpType1(inst.OpType1) == opcodes.IS_CONST {
			target := int(inst.Op1)
			p.markReachable(instructions, reachable, target)
		}
		// Don't mark next instruction as reachable for unconditional jump

	case opcodes.OP_JMPZ, opcodes.OP_JMPNZ:
		// Conditional jump - both paths are reachable
		if opcodes.DecodeOpType2(inst.OpType1) == opcodes.IS_CONST {
			target := int(inst.Op2)
			p.markReachable(instructions, reachable, target)
		}
		p.markReachable(instructions, reachable, start+1)

	case opcodes.OP_RETURN, opcodes.OP_EXIT, opcodes.OP_THROW:
		// These instructions terminate execution
		// Don't mark next instruction as reachable

	default:
		// Regular instruction - mark next as reachable
		p.markReachable(instructions, reachable, start+1)
	}
}

// PeepholeOptimizationPass performs local optimizations
type PeepholeOptimizationPass struct{}

func (p *PeepholeOptimizationPass) Name() string {
	return "PeepholeOptimization"
}

func (p *PeepholeOptimizationPass) Optimize(instructions []opcodes.Instruction, constants []*values.Value) ([]opcodes.Instruction, []*values.Value, bool) {
	optimized := make([]opcodes.Instruction, 0, len(instructions))
	changed := false

	for i := 0; i < len(instructions); {
		// Look for patterns to optimize
		if i < len(instructions)-1 {
			// Pattern: QM_ASSIGN followed by FETCH_R of the same temp
			if p.isRedundantAssignFetch(instructions[i], instructions[i+1]) {
				// Skip the FETCH_R, just keep the QM_ASSIGN with the fetch's result
				inst := instructions[i]
				inst.Result = instructions[i+1].Result
				optimized = append(optimized, inst)
				i += 2
				changed = true
				continue
			}

			// Pattern: Assignment to temp that's never used
			if p.isUnusedTempAssignment(instructions, i) {
				// Skip this instruction
				i++
				changed = true
				continue
			}
		}

		optimized = append(optimized, instructions[i])
		i++
	}

	return optimized, constants, changed
}

func (p *PeepholeOptimizationPass) isRedundantAssignFetch(inst1, inst2 opcodes.Instruction) bool {
	return inst1.Opcode == opcodes.OP_QM_ASSIGN &&
		inst2.Opcode == opcodes.OP_FETCH_R &&
		opcodes.DecodeResultType(inst1.OpType2) == opcodes.IS_TMP_VAR &&
		opcodes.DecodeOpType1(inst2.OpType1) == opcodes.IS_TMP_VAR &&
		inst1.Result == inst2.Op1
}

func (p *PeepholeOptimizationPass) isUnusedTempAssignment(instructions []opcodes.Instruction, index int) bool {
	inst := instructions[index]

	// Only optimize temporary variable assignments
	if opcodes.DecodeResultType(inst.OpType2) != opcodes.IS_TMP_VAR {
		return false
	}

	tempVar := inst.Result

	// Check if this temp is used later
	for i := index + 1; i < len(instructions); i++ {
		later := instructions[i]

		// Check all operands
		if (opcodes.DecodeOpType1(later.OpType1) == opcodes.IS_TMP_VAR && later.Op1 == tempVar) ||
			(opcodes.DecodeOpType2(later.OpType1) == opcodes.IS_TMP_VAR && later.Op2 == tempVar) {
			return false // Temp is used
		}
	}

	return true // Temp is never used
}

// JumpOptimizationPass optimizes jump instructions
type JumpOptimizationPass struct{}

func (p *JumpOptimizationPass) Name() string {
	return "JumpOptimization"
}

func (p *JumpOptimizationPass) Optimize(instructions []opcodes.Instruction, constants []*values.Value) ([]opcodes.Instruction, []*values.Value, bool) {
	optimized := make([]opcodes.Instruction, len(instructions))
	copy(optimized, instructions)
	changed := false

	for i := 0; i < len(optimized); i++ {
		inst := optimized[i]

		// Optimize jumps to next instruction
		if p.isJumpToNext(inst, i) {
			optimized[i] = opcodes.Instruction{Opcode: opcodes.OP_NOP}
			changed = true
			continue
		}

		// Optimize jump chains (jump to jump)
		if p.isJump(inst.Opcode) {
			if newTarget := p.resolveJumpChain(optimized, inst, constants); newTarget != -1 {
				// Update jump target
				if opcodes.DecodeOpType1(inst.OpType1) == opcodes.IS_CONST {
					optimized[i].Op1 = uint32(newTarget)
				} else if opcodes.DecodeOpType2(inst.OpType1) == opcodes.IS_CONST {
					optimized[i].Op2 = uint32(newTarget)
				}
				changed = true
			}
		}
	}

	return optimized, constants, changed
}

func (p *JumpOptimizationPass) isJumpToNext(inst opcodes.Instruction, currentIndex int) bool {
	if !p.isJump(inst.Opcode) {
		return false
	}

	var target int
	if opcodes.DecodeOpType1(inst.OpType1) == opcodes.IS_CONST {
		target = int(inst.Op1)
	} else if opcodes.DecodeOpType2(inst.OpType1) == opcodes.IS_CONST {
		target = int(inst.Op2)
	} else {
		return false
	}

	return target == currentIndex+1
}

func (p *JumpOptimizationPass) isJump(opcode opcodes.Opcode) bool {
	switch opcode {
	case opcodes.OP_JMP, opcodes.OP_JMPZ, opcodes.OP_JMPNZ:
		return true
	default:
		return false
	}
}

func (p *JumpOptimizationPass) resolveJumpChain(instructions []opcodes.Instruction, inst opcodes.Instruction, constants []*values.Value) int {
	visited := make(map[int]bool)

	var target int
	if opcodes.DecodeOpType1(inst.OpType1) == opcodes.IS_CONST {
		target = int(inst.Op1)
	} else if opcodes.DecodeOpType2(inst.OpType1) == opcodes.IS_CONST {
		target = int(inst.Op2)
	} else {
		return -1
	}

	// Follow jump chain
	for target >= 0 && target < len(instructions) && !visited[target] {
		visited[target] = true
		targetInst := instructions[target]

		// Only follow unconditional jumps in chains
		if targetInst.Opcode == opcodes.OP_JMP && opcodes.DecodeOpType1(targetInst.OpType1) == opcodes.IS_CONST {
			newTarget := int(targetInst.Op1)
			if newTarget != target {
				target = newTarget
				continue
			}
		}

		break
	}

	// Return new target if it's different from original
	originalTarget := int(inst.Op1)
	if inst.Opcode != opcodes.OP_JMP {
		originalTarget = int(inst.Op2)
	}

	if target != originalTarget {
		return target
	}

	return -1
}

// TemporaryEliminationPass reduces temporary variable usage
type TemporaryEliminationPass struct{}

func (p *TemporaryEliminationPass) Name() string {
	return "TemporaryElimination"
}

func (p *TemporaryEliminationPass) Optimize(instructions []opcodes.Instruction, constants []*values.Value) ([]opcodes.Instruction, []*values.Value, bool) {
	// This is a simplified implementation
	// A full implementation would perform register allocation

	// For now, just identify and merge temporary variables that have the same lifetime
	optimized := make([]opcodes.Instruction, len(instructions))
	copy(optimized, instructions)

	// This pass would be quite complex to implement fully
	// It requires liveness analysis and register allocation

	return optimized, constants, false
}

// OptimizationStats tracks optimization statistics
type OptimizationStats struct {
	PassStats     map[string]int
	OriginalSize  int
	OptimizedSize int
	Iterations    int
}

func (o *Optimizer) OptimizeWithStats(instructions []opcodes.Instruction, constants []*values.Value) ([]opcodes.Instruction, []*values.Value, *OptimizationStats) {
	stats := &OptimizationStats{
		PassStats:    make(map[string]int),
		OriginalSize: len(instructions),
		Iterations:   0,
	}

	currentInstructions := make([]opcodes.Instruction, len(instructions))
	copy(currentInstructions, instructions)

	currentConstants := make([]*values.Value, len(constants))
	copy(currentConstants, constants)

	maxIterations := 10

	for stats.Iterations < maxIterations {
		iterationChanges := false

		for _, pass := range o.passes {
			optimized, newConstants, changed := pass.Optimize(currentInstructions, currentConstants)

			if changed {
				currentInstructions = optimized
				currentConstants = newConstants
				iterationChanges = true

				// Update stats
				passName := pass.Name()
				stats.PassStats[passName]++
			}
		}

		if !iterationChanges {
			break
		}

		stats.Iterations++
	}

	stats.OptimizedSize = len(currentInstructions)

	return currentInstructions, currentConstants, stats
}
