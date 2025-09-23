package vm

import (
	"sync"

	"github.com/wudi/hey/values"
)

// CallStackManager manages the call stack for function execution
type CallStackManager struct {
	frames []*CallFrame
	mu     sync.Mutex
}

// NewCallStackManager creates a new call stack manager
func NewCallStackManager() *CallStackManager {
	return &CallStackManager{
		frames: make([]*CallFrame, 0, 8),
	}
}

// PushFrame adds a new call frame to the call stack
func (cs *CallStackManager) PushFrame(frame *CallFrame) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.frames = append(cs.frames, frame)
}

// PopFrame removes and returns the current call frame. Returns nil when the stack is empty.
func (cs *CallStackManager) PopFrame() *CallFrame {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	if len(cs.frames) == 0 {
		return nil
	}
	idx := len(cs.frames) - 1
	frame := cs.frames[idx]
	cs.frames = cs.frames[:idx]
	return frame
}

// CurrentFrame returns the actively executing call frame
func (cs *CallStackManager) CurrentFrame() *CallFrame {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	if len(cs.frames) == 0 {
		return nil
	}
	return cs.frames[len(cs.frames)-1]
}

// Depth returns the current call stack depth
func (cs *CallStackManager) Depth() int {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	return len(cs.frames)
}

// IsEmpty returns true if the call stack is empty
func (cs *CallStackManager) IsEmpty() bool {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	return len(cs.frames) == 0
}

// GetFrames returns a copy of all frames (for debugging)
func (cs *CallStackManager) GetFrames() []*CallFrame {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	frames := make([]*CallFrame, len(cs.frames))
	copy(frames, cs.frames)
	return frames
}

// UpdateGlobalBindings updates global variable bindings across all frames
func (cs *CallStackManager) UpdateGlobalBindings(names []string, value *values.Value) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	for _, frame := range cs.frames {
		if frame == nil || len(frame.GlobalSlots) == 0 {
			continue
		}
		for slot, bound := range frame.GlobalSlots {
			for _, candidate := range names {
				if bound == candidate {
					frame.Locals[slot] = value
					break
				}
			}
		}
	}
}

// Clear resets the call stack (useful for testing)
func (cs *CallStackManager) Clear() {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.frames = cs.frames[:0]
}

// Copy creates a shallow copy of the call stack (frames are shared)
func (cs *CallStackManager) Copy() *CallStackManager {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	copy := NewCallStackManager()
	copy.frames = make([]*CallFrame, len(cs.frames))
	for i, frame := range cs.frames {
		copy.frames[i] = frame // Shallow copy - frames are shared
	}
	return copy
}