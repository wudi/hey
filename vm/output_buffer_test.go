package vm

import (
	"bytes"
	"testing"

	"github.com/wudi/hey/values"
)

func TestOutputBufferStack(t *testing.T) {
	var output bytes.Buffer
	obs := NewOutputBufferStack(&output)

	// Test initial state
	if obs.GetLevel() != 0 {
		t.Errorf("Expected initial level 0, got %d", obs.GetLevel())
	}

	// Test start buffer
	success := obs.Start("", 0, 0)
	if !success {
		t.Error("Start should return true")
	}

	if obs.GetLevel() != 1 {
		t.Errorf("Expected level 1 after start, got %d", obs.GetLevel())
	}

	// Test write to buffer
	n, err := obs.Write([]byte("Hello World"))
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if n != 11 {
		t.Errorf("Expected to write 11 bytes, wrote %d", n)
	}

	// Test get contents
	contents := obs.GetContents()
	if contents != "Hello World" {
		t.Errorf("Expected 'Hello World', got '%s'", contents)
	}

	// Test get length
	length := obs.GetLength()
	if length != 11 {
		t.Errorf("Expected length 11, got %d", length)
	}

	// Test clean
	success = obs.Clean()
	if !success {
		t.Error("Clean should return true")
	}

	if obs.GetContents() != "" {
		t.Error("Buffer should be empty after clean")
	}

	// Write new content
	obs.Write([]byte("New Content"))

	// Test end flush
	success = obs.EndFlush()
	if !success {
		t.Error("EndFlush should return true")
	}

	if obs.GetLevel() != 0 {
		t.Errorf("Expected level 0 after end flush, got %d", obs.GetLevel())
	}

	if output.String() != "New Content" {
		t.Errorf("Expected 'New Content' in output, got '%s'", output.String())
	}
}

func TestNestedOutputBuffers(t *testing.T) {
	var output bytes.Buffer
	obs := NewOutputBufferStack(&output)

	// Start first buffer
	obs.Start("", 0, 0)
	obs.Write([]byte("Level1"))

	// Start second buffer
	obs.Start("", 0, 0)
	obs.Write([]byte("Level2"))

	if obs.GetLevel() != 2 {
		t.Errorf("Expected level 2, got %d", obs.GetLevel())
	}

	// Get contents of second buffer
	contents := obs.GetContents()
	if contents != "Level2" {
		t.Errorf("Expected 'Level2', got '%s'", contents)
	}

	// Flush second buffer to first
	obs.Flush()
	obs.EndClean()

	// Should be back to level 1
	if obs.GetLevel() != 1 {
		t.Errorf("Expected level 1, got %d", obs.GetLevel())
	}

	// First buffer should contain both
	contents = obs.GetContents()
	if contents != "Level1Level2" {
		t.Errorf("Expected 'Level1Level2', got '%s'", contents)
	}

	// End first buffer
	obs.EndFlush()

	if obs.GetLevel() != 0 {
		t.Errorf("Expected level 0, got %d", obs.GetLevel())
	}

	if output.String() != "Level1Level2" {
		t.Errorf("Expected 'Level1Level2' in output, got '%s'", output.String())
	}
}

func TestOutputBufferGetClean(t *testing.T) {
	var output bytes.Buffer
	obs := NewOutputBufferStack(&output)

	obs.Start("", 0, 0)
	obs.Write([]byte("Test Content"))

	contents, success := obs.GetClean()
	if !success {
		t.Error("GetClean should return true")
	}

	if contents != "Test Content" {
		t.Errorf("Expected 'Test Content', got '%s'", contents)
	}

	if obs.GetLevel() != 0 {
		t.Errorf("Expected level 0 after GetClean, got %d", obs.GetLevel())
	}
}

func TestOutputBufferGetFlush(t *testing.T) {
	var output bytes.Buffer
	obs := NewOutputBufferStack(&output)

	obs.Start("", 0, 0)
	obs.Write([]byte("Flush Content"))

	contents, success := obs.GetFlush()
	if !success {
		t.Error("GetFlush should return true")
	}

	if contents != "Flush Content" {
		t.Errorf("Expected 'Flush Content', got '%s'", contents)
	}

	if obs.GetLevel() != 0 {
		t.Errorf("Expected level 0 after GetFlush, got %d", obs.GetLevel())
	}

	if output.String() != "Flush Content" {
		t.Errorf("Expected 'Flush Content' in output, got '%s'", output.String())
	}
}

func TestOutputBufferStatus(t *testing.T) {
	var output bytes.Buffer
	obs := NewOutputBufferStack(&output)

	// Test empty status
	status := obs.GetStatus()
	if !status.IsArray() {
		t.Error("Status should be an array")
	}

	// Start buffer and test status
	obs.Start("test_handler", 1024, 0)
	obs.Write([]byte("Test"))

	status = obs.GetStatus()
	if !status.IsArray() {
		t.Error("Status should be an array")
	}

	// Check status fields
	name := status.ArrayGet(values.NewString("name"))
	if name == nil || name.ToString() != "default output handler" {
		t.Errorf("Expected 'default output handler', got '%s'", name.ToString())
	}

	level := status.ArrayGet(values.NewString("level"))
	if level == nil || level.ToInt() != 0 {
		t.Errorf("Expected level 0, got %d", level.ToInt())
	}

	bufferSize := status.ArrayGet(values.NewString("buffer_size"))
	if bufferSize == nil || bufferSize.ToInt() != 4 {
		t.Errorf("Expected buffer size 4, got %d", bufferSize.ToInt())
	}
}

func TestOutputBufferStatusFull(t *testing.T) {
	var output bytes.Buffer
	obs := NewOutputBufferStack(&output)

	// Start two buffers
	obs.Start("handler1", 1024, 0)
	obs.Start("handler2", 2048, 0)

	statusArray := obs.GetStatusFull()
	if !statusArray.IsArray() {
		t.Error("StatusFull should return an array")
	}

	// Should have 2 buffer statuses
	if statusArray.ArrayCount() != 2 {
		t.Errorf("Expected 2 buffer statuses, got %d", statusArray.ArrayCount())
	}
}

func TestOutputBufferListHandlers(t *testing.T) {
	var output bytes.Buffer
	obs := NewOutputBufferStack(&output)

	// Test empty handlers
	handlers := obs.ListHandlers()
	if len(handlers) != 0 {
		t.Errorf("Expected 0 handlers, got %d", len(handlers))
	}

	// Start buffers
	obs.Start("custom_handler", 0, 0)
	obs.Start("", 0, 0)

	handlers = obs.ListHandlers()
	if len(handlers) != 2 {
		t.Errorf("Expected 2 handlers, got %d", len(handlers))
	}

	if handlers[0] != "custom_handler" {
		t.Errorf("Expected 'custom_handler', got '%s'", handlers[0])
	}

	if handlers[1] != "default output handler" {
		t.Errorf("Expected 'default output handler', got '%s'", handlers[1])
	}
}

func TestImplicitFlush(t *testing.T) {
	var output bytes.Buffer
	obs := NewOutputBufferStack(&output)

	// Test implicit flush off (default)
	obs.SetImplicitFlush(false)
	obs.Write([]byte("Test"))

	if output.String() != "Test" {
		t.Errorf("Expected 'Test' with implicit flush off, got '%s'", output.String())
	}

	// Reset output
	output.Reset()

	// Test implicit flush on
	obs.SetImplicitFlush(true)
	obs.Write([]byte("Test2"))

	if output.String() != "Test2" {
		t.Errorf("Expected 'Test2' with implicit flush on, got '%s'", output.String())
	}
}