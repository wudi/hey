package vm

import (
	"fmt"
	"sort"
	"sync"

	"github.com/wudi/hey/opcodes"
)

type profileState struct {
	mu sync.Mutex

	instructionCounts map[int]int
	opcodeCounts      map[opcodes.Opcode]int

	allocs int
	frees  int

	debug []string
}

func newProfileState() *profileState {
	return &profileState{
		instructionCounts: make(map[int]int),
		opcodeCounts:      make(map[opcodes.Opcode]int),
		debug:             make([]string, 0, 64),
	}
}

func (ps *profileState) observe(ip int, opcode opcodes.Opcode) {
	ps.mu.Lock()
	ps.instructionCounts[ip]++
	ps.opcodeCounts[opcode]++
	ps.mu.Unlock()
}

func (ps *profileState) addDebug(message string) {
	ps.mu.Lock()
	ps.debug = append(ps.debug, message)
	ps.mu.Unlock()
}

func (ps *profileState) debugRecords() []string {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	out := make([]string, len(ps.debug))
	copy(out, ps.debug)
	return out
}

func (ps *profileState) hotSpots(n int) []HotSpot {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	spots := make([]HotSpot, 0, len(ps.instructionCounts))
	for ip, count := range ps.instructionCounts {
		spots = append(spots, HotSpot{IP: ip, Count: count})
	}
	sort.Slice(spots, func(i, j int) bool {
		if spots[i].Count == spots[j].Count {
			return spots[i].IP < spots[j].IP
		}
		return spots[i].Count > spots[j].Count
	})
	if n <= 0 || n >= len(spots) {
		return spots
	}
	return spots[:n]
}

func (ps *profileState) render() string {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	if len(ps.instructionCounts) == 0 {
		return "(no profiling data)"
	}
	total := 0
	for _, count := range ps.instructionCounts {
		total += count
	}
	return fmt.Sprintf("instructions executed: %d, unique ips: %d", total, len(ps.instructionCounts))
}

func (ps *profileState) recordAlloc(delta int) {
	ps.mu.Lock()
	if delta > 0 {
		ps.allocs += delta
	} else {
		ps.frees += -delta
	}
	ps.mu.Unlock()
}
