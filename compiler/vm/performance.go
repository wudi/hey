package vm

import (
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/wudi/hey/compiler/values"
)

// PerformanceMetrics tracks VM execution statistics
type PerformanceMetrics struct {
	mutex                sync.RWMutex
	StartTime            time.Time
	TotalInstructions    uint64
	InstructionCounts    map[string]uint64
	FunctionCallCounts   map[string]uint64
	MemoryAllocations    uint64
	MemoryDeallocations  uint64
	GCCollections        uint32
	TotalExecutionTime   time.Duration
	AverageInstPerSecond float64
	PeakMemoryUsage      uint64
	CurrentMemoryUsage   uint64
}

// NewPerformanceMetrics creates a new performance metrics tracker
func NewPerformanceMetrics() *PerformanceMetrics {
	return &PerformanceMetrics{
		StartTime:          time.Now(),
		InstructionCounts:  make(map[string]uint64),
		FunctionCallCounts: make(map[string]uint64),
	}
}

// RecordInstruction records the execution of an instruction
func (pm *PerformanceMetrics) RecordInstruction(opcodeString string) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	pm.TotalInstructions++
	pm.InstructionCounts[opcodeString]++
}

// RecordFunctionCall records a function call
func (pm *PerformanceMetrics) RecordFunctionCall(functionName string) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	pm.FunctionCallCounts[functionName]++
}

// RecordMemoryAllocation records memory allocation
func (pm *PerformanceMetrics) RecordMemoryAllocation(size uint64) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	pm.MemoryAllocations++
	pm.CurrentMemoryUsage += size
	if pm.CurrentMemoryUsage > pm.PeakMemoryUsage {
		pm.PeakMemoryUsage = pm.CurrentMemoryUsage
	}
}

// RecordMemoryDeallocation records memory deallocation
func (pm *PerformanceMetrics) RecordMemoryDeallocation(size uint64) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	pm.MemoryDeallocations++
	if pm.CurrentMemoryUsage >= size {
		pm.CurrentMemoryUsage -= size
	}
}

// UpdateExecutionTime updates the total execution time and calculates averages
func (pm *PerformanceMetrics) UpdateExecutionTime() {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	pm.TotalExecutionTime = time.Since(pm.StartTime)
	if pm.TotalExecutionTime > 0 {
		pm.AverageInstPerSecond = float64(pm.TotalInstructions) / pm.TotalExecutionTime.Seconds()
	}
}

// GetReport generates a performance report
func (pm *PerformanceMetrics) GetReport() string {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	// Update timing (avoid calling UpdateExecutionTime to prevent deadlock)
	totalTime := time.Since(pm.StartTime)
	avgPerSecond := float64(0)
	if totalTime > 0 {
		avgPerSecond = float64(pm.TotalInstructions) / totalTime.Seconds()
	}

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	report := fmt.Sprintf(`
=== VM Performance Report ===
Execution Time: %v
Total Instructions: %d
Instructions/Second: %.2f
Peak Memory Usage: %d bytes (%.2f MB)
Current Memory Usage: %d bytes (%.2f MB)
Go Memory Stats:
  - Allocated: %d bytes
  - Total Allocations: %d
  - GC Cycles: %d
  - Heap Objects: %d

Top Instructions:`,
		totalTime,
		pm.TotalInstructions,
		avgPerSecond,
		pm.PeakMemoryUsage,
		float64(pm.PeakMemoryUsage)/1024/1024,
		pm.CurrentMemoryUsage,
		float64(pm.CurrentMemoryUsage)/1024/1024,
		memStats.Alloc,
		memStats.TotalAlloc,
		memStats.NumGC,
		memStats.HeapObjects,
	)

	// Top 10 most executed instructions
	for opcode, count := range pm.InstructionCounts {
		if count > 0 {
			percentage := float64(count) / float64(pm.TotalInstructions) * 100
			report += fmt.Sprintf("\n  - %s: %d (%.2f%%)", opcode, count, percentage)
		}
	}

	report += "\n\nFunction Calls:"
	for funcName, count := range pm.FunctionCallCounts {
		if count > 0 {
			report += fmt.Sprintf("\n  - %s: %d calls", funcName, count)
		}
	}

	return report
}

// VMOptimizer provides optimization strategies for the VM
type VMOptimizer struct {
	InstructionCache map[string]func(*VirtualMachine, *ExecutionContext) error
	HotSpots         map[int]uint64 // IP -> execution count
	OptimizedPaths   map[string]bool
}

// NewVMOptimizer creates a new VM optimizer
func NewVMOptimizer() *VMOptimizer {
	return &VMOptimizer{
		InstructionCache: make(map[string]func(*VirtualMachine, *ExecutionContext) error),
		HotSpots:         make(map[int]uint64),
		OptimizedPaths:   make(map[string]bool),
	}
}

// RecordHotSpot records frequently executed instruction positions
func (opt *VMOptimizer) RecordHotSpot(ip int) {
	opt.HotSpots[ip]++
}

// IsHotSpot determines if an instruction position is a hot spot
func (opt *VMOptimizer) IsHotSpot(ip int, threshold uint64) bool {
	return opt.HotSpots[ip] >= threshold
}

// GetHotSpots returns the most executed instruction positions
func (opt *VMOptimizer) GetHotSpots(limit int) []HotSpot {
	hotSpots := make([]HotSpot, 0, len(opt.HotSpots))
	for ip, count := range opt.HotSpots {
		hotSpots = append(hotSpots, HotSpot{IP: ip, Count: count})
	}

	// Simple sorting by count (descending)
	for i := 0; i < len(hotSpots)-1; i++ {
		for j := i + 1; j < len(hotSpots); j++ {
			if hotSpots[j].Count > hotSpots[i].Count {
				hotSpots[i], hotSpots[j] = hotSpots[j], hotSpots[i]
			}
		}
	}

	if limit < len(hotSpots) {
		result := make([]HotSpot, limit)
		copy(result, hotSpots[:limit])
		return result
	}
	result := make([]HotSpot, len(hotSpots))
	copy(result, hotSpots)
	return result
}

// HotSpot represents a frequently executed instruction position
type HotSpot struct {
	IP    int
	Count uint64
}

// MemoryPool provides efficient memory allocation for VM objects
type MemoryPool struct {
	valuePool     sync.Pool
	contextPool   sync.Pool
	framePool     sync.Pool
	allocations   uint64
	deallocations uint64
}

// NewMemoryPool creates a new memory pool
func NewMemoryPool() *MemoryPool {
	return &MemoryPool{
		valuePool: sync.Pool{
			New: func() interface{} {
				return &values.Value{}
			},
		},
		contextPool: sync.Pool{
			New: func() interface{} {
				return &ExecutionContext{
					Variables:   make(map[uint32]*values.Value),
					Temporaries: make(map[uint32]*values.Value),
					GlobalVars:  make(map[string]*values.Value),
					// Classes now handled by unified registry
				}
			},
		},
		framePool: sync.Pool{
			New: func() interface{} {
				return &CallFrame{
					Variables: make(map[uint32]*values.Value),
				}
			},
		},
	}
}

// GetValue gets a reusable Value from the pool
func (mp *MemoryPool) GetValue() *values.Value {
	mp.allocations++
	return mp.valuePool.Get().(*values.Value)
}

// PutValue returns a Value to the pool
func (mp *MemoryPool) PutValue(v *values.Value) {
	// Reset the value
	v.Type = 0
	v.Data = nil
	mp.deallocations++
	mp.valuePool.Put(v)
}

// GetExecutionContext gets a reusable ExecutionContext from the pool
func (mp *MemoryPool) GetExecutionContext() *ExecutionContext {
	ctx := mp.contextPool.Get().(*ExecutionContext)
	// Clear the context for reuse
	for k := range ctx.Variables {
		delete(ctx.Variables, k)
	}
	for k := range ctx.Temporaries {
		delete(ctx.Temporaries, k)
	}
	for k := range ctx.GlobalVars {
		delete(ctx.GlobalVars, k)
	}
	return ctx
}

// PutExecutionContext returns an ExecutionContext to the pool
func (mp *MemoryPool) PutExecutionContext(ctx *ExecutionContext) {
	mp.contextPool.Put(ctx)
}

// GetStats returns memory pool statistics
func (mp *MemoryPool) GetStats() (uint64, uint64) {
	return mp.allocations, mp.deallocations
}
