package vm

import (
	"bytes"
	"io"
	"sync"
	"github.com/wudi/hey/values"
)

// OutputBuffer represents a single output buffer in the stack
type OutputBuffer struct {
	buffer      *bytes.Buffer
	name        string
	flags       int
	chunkSize   int
	handler     string // Handler function name (empty for default)
	level       int
}

// OutputBufferStack manages nested output buffers
type OutputBufferStack struct {
	mu          sync.Mutex
	buffers     []*OutputBuffer
	baseWriter  io.Writer      // Original output writer (stdout)
	implicitFlush bool
	execCtx     *ExecutionContext // Reference to track headers_sent
}

// NewOutputBufferStack creates a new output buffer stack
func NewOutputBufferStack(baseWriter io.Writer) *OutputBufferStack {
	return &OutputBufferStack{
		buffers:     make([]*OutputBuffer, 0),
		baseWriter:  baseWriter,
		implicitFlush: false,
	}
}

// Start creates and pushes a new buffer onto the stack
func (obs *OutputBufferStack) Start(handler string, chunkSize int, flags int) bool {
	obs.mu.Lock()
	defer obs.mu.Unlock()

	level := len(obs.buffers)
	buffer := &OutputBuffer{
		buffer:    &bytes.Buffer{},
		name:      "default output handler",
		flags:     flags,
		chunkSize: chunkSize,
		handler:   handler,
		level:     level,
	}

	obs.buffers = append(obs.buffers, buffer)
	return true
}

// GetContents returns the contents of the active buffer without removing it
func (obs *OutputBufferStack) GetContents() string {
	obs.mu.Lock()
	defer obs.mu.Unlock()

	if len(obs.buffers) == 0 {
		return ""
	}

	return obs.buffers[len(obs.buffers)-1].buffer.String()
}

// GetLength returns the length of the active buffer
func (obs *OutputBufferStack) GetLength() int {
	obs.mu.Lock()
	defer obs.mu.Unlock()

	if len(obs.buffers) == 0 {
		return 0
	}

	return obs.buffers[len(obs.buffers)-1].buffer.Len()
}

// GetLevel returns the nesting level of output buffering
func (obs *OutputBufferStack) GetLevel() int {
	obs.mu.Lock()
	defer obs.mu.Unlock()

	return len(obs.buffers)
}

// Clean erases the contents of the active buffer
func (obs *OutputBufferStack) Clean() bool {
	obs.mu.Lock()
	defer obs.mu.Unlock()

	if len(obs.buffers) == 0 {
		return false
	}

	obs.buffers[len(obs.buffers)-1].buffer.Reset()
	return true
}

// EndClean erases and removes the active buffer
func (obs *OutputBufferStack) EndClean() bool {
	obs.mu.Lock()
	defer obs.mu.Unlock()

	if len(obs.buffers) == 0 {
		return false
	}

	// Remove the last buffer without outputting
	obs.buffers = obs.buffers[:len(obs.buffers)-1]
	return true
}

// Flush sends the active buffer contents to the output
func (obs *OutputBufferStack) Flush() bool {
	obs.mu.Lock()
	defer obs.mu.Unlock()

	if len(obs.buffers) == 0 {
		return false
	}

	activeBuffer := obs.buffers[len(obs.buffers)-1]
	content := activeBuffer.buffer.Bytes()

	// Write to the next level down or base writer
	if len(obs.buffers) > 1 {
		// Write to parent buffer
		obs.buffers[len(obs.buffers)-2].buffer.Write(content)
	} else {
		// Write to base writer
		obs.baseWriter.Write(content)
	}

	// Clear the current buffer after flushing
	activeBuffer.buffer.Reset()
	return true
}

// EndFlush flushes and removes the active buffer
func (obs *OutputBufferStack) EndFlush() bool {
	obs.mu.Lock()
	defer obs.mu.Unlock()

	if len(obs.buffers) == 0 {
		return false
	}

	activeBuffer := obs.buffers[len(obs.buffers)-1]
	content := activeBuffer.buffer.Bytes()

	// Remove the buffer first
	obs.buffers = obs.buffers[:len(obs.buffers)-1]

	// Write to the next level down or base writer
	if len(obs.buffers) > 0 {
		// Write to parent buffer
		obs.buffers[len(obs.buffers)-1].buffer.Write(content)
	} else {
		// Write to base writer
		obs.baseWriter.Write(content)
	}

	return true
}

// GetClean returns contents and removes the active buffer
func (obs *OutputBufferStack) GetClean() (string, bool) {
	obs.mu.Lock()
	defer obs.mu.Unlock()

	if len(obs.buffers) == 0 {
		return "", false
	}

	content := obs.buffers[len(obs.buffers)-1].buffer.String()
	obs.buffers = obs.buffers[:len(obs.buffers)-1]
	return content, true
}

// GetFlush returns contents, flushes, and removes the active buffer
func (obs *OutputBufferStack) GetFlush() (string, bool) {
	obs.mu.Lock()
	defer obs.mu.Unlock()

	if len(obs.buffers) == 0 {
		return "", false
	}

	activeBuffer := obs.buffers[len(obs.buffers)-1]
	content := activeBuffer.buffer.String()

	// Remove the buffer first
	obs.buffers = obs.buffers[:len(obs.buffers)-1]

	// Write to the next level down or base writer
	if len(obs.buffers) > 0 {
		// Write to parent buffer
		obs.buffers[len(obs.buffers)-1].buffer.Write([]byte(content))
	} else {
		// Write to base writer
		obs.baseWriter.Write([]byte(content))
	}

	return content, true
}

// GetStatus returns status information for the active buffer
func (obs *OutputBufferStack) GetStatus() *values.Value {
	obs.mu.Lock()
	defer obs.mu.Unlock()

	if len(obs.buffers) == 0 {
		return values.NewArray()
	}

	activeBuffer := obs.buffers[len(obs.buffers)-1]
	status := values.NewArray()
	status.ArraySet(values.NewString("name"), values.NewString(activeBuffer.name))
	status.ArraySet(values.NewString("type"), values.NewInt(0)) // 0 for internal handler
	status.ArraySet(values.NewString("flags"), values.NewInt(int64(activeBuffer.flags)))
	status.ArraySet(values.NewString("level"), values.NewInt(int64(activeBuffer.level)))
	status.ArraySet(values.NewString("chunk_size"), values.NewInt(int64(activeBuffer.chunkSize)))
	status.ArraySet(values.NewString("buffer_size"), values.NewInt(int64(activeBuffer.buffer.Len())))
	status.ArraySet(values.NewString("buffer_used"), values.NewInt(int64(activeBuffer.buffer.Len())))

	return status
}

// GetStatusFull returns status information for all buffers
func (obs *OutputBufferStack) GetStatusFull() *values.Value {
	obs.mu.Lock()
	defer obs.mu.Unlock()

	result := values.NewArray()

	for _, buffer := range obs.buffers {
		status := values.NewArray()
		status.ArraySet(values.NewString("name"), values.NewString(buffer.name))
		status.ArraySet(values.NewString("type"), values.NewInt(0))
		status.ArraySet(values.NewString("flags"), values.NewInt(int64(buffer.flags)))
		status.ArraySet(values.NewString("level"), values.NewInt(int64(buffer.level)))
		status.ArraySet(values.NewString("chunk_size"), values.NewInt(int64(buffer.chunkSize)))
		status.ArraySet(values.NewString("buffer_size"), values.NewInt(int64(buffer.buffer.Len())))
		status.ArraySet(values.NewString("buffer_used"), values.NewInt(int64(buffer.buffer.Len())))
		result.ArraySet(nil, status)
	}

	return result
}

// ListHandlers returns a list of active output handler names
func (obs *OutputBufferStack) ListHandlers() []string {
	obs.mu.Lock()
	defer obs.mu.Unlock()

	handlers := make([]string, 0, len(obs.buffers))
	for _, buffer := range obs.buffers {
		if buffer.handler != "" {
			handlers = append(handlers, buffer.handler)
		} else {
			handlers = append(handlers, "default output handler")
		}
	}
	return handlers
}

// SetImplicitFlush sets the implicit flush mode
func (obs *OutputBufferStack) SetImplicitFlush(on bool) {
	obs.mu.Lock()
	defer obs.mu.Unlock()
	obs.implicitFlush = on
}

// Write implements io.Writer interface for intercepting output
func (obs *OutputBufferStack) Write(p []byte) (n int, err error) {
	obs.mu.Lock()
	defer obs.mu.Unlock()

	// Mark headers as sent if this is the first non-buffered output
	if obs.execCtx != nil && obs.execCtx.HTTPContext != nil && len(obs.buffers) == 0 && len(p) > 0 {
		obs.execCtx.HTTPContext.MarkHeadersSent("output:1") // Track location later if needed
	}

	// If we have active buffers, write to the topmost buffer
	if len(obs.buffers) > 0 {
		return obs.buffers[len(obs.buffers)-1].buffer.Write(p)
	}

	// Otherwise, write to the base writer
	if obs.implicitFlush {
		// In implicit flush mode, immediately flush to base writer
		return obs.baseWriter.Write(p)
	}
	return obs.baseWriter.Write(p)
}

// FlushSystem flushes the system output buffer (for flush() function)
func (obs *OutputBufferStack) FlushSystem() {
	// In a real web server environment, this would flush headers and content
	// For our implementation, we ensure base writer is flushed if it's a buffered writer
	if flusher, ok := obs.baseWriter.(interface{ Flush() }); ok {
		flusher.Flush()
	}
}