package optimizer

import (
	"github.com/wudi/hey/opcodes"
	"github.com/wudi/hey/values"
)

// OptimizationStats captures bookkeeping for optimizer runs.
type OptimizationStats struct {
	OriginalSize  int
	OptimizedSize int
	Iterations    int
	PassStats     map[string]int
}

// Optimizer is a placeholder optimization pipeline.
type Optimizer struct{}

// NewOptimizer constructs a new optimizer instance.
func NewOptimizer() *Optimizer {
	return &Optimizer{}
}

// OptimizeWithStats currently performs no transformations and returns the
// original instruction stream along with trivial statistics. The API mirrors
// what a future optimizer would provide so callers remain stable.
func (o *Optimizer) OptimizeWithStats(instr []opcodes.Instruction, consts []*values.Value) ([]opcodes.Instruction, []*values.Value, OptimizationStats) {
	stats := OptimizationStats{
		OriginalSize:  len(instr),
		OptimizedSize: len(instr),
		Iterations:    0,
		PassStats:     make(map[string]int),
	}
	return instr, consts, stats
}
