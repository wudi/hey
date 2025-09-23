package vm

import (
	"testing"

	"github.com/wudi/hey/values"
)

func TestVariableManager_BasicOperations(t *testing.T) {
	vm := NewVariableManager()

	// Test setting and getting global variables
	vm.SetGlobalVar("test_global", values.NewString("global_value"))
	if val, ok := vm.GetGlobalVar("test_global"); !ok || !val.Equal(values.NewString("global_value")) {
		t.Errorf("GetGlobalVar failed")
	}

	// Test setting and getting local variables
	vm.SetVariable("test_local", values.NewInt(42))
	if val, ok := vm.GetVariable("test_local"); !ok || !val.Equal(values.NewInt(42)) {
		t.Errorf("GetVariable failed")
	}

	// Test temporary variables
	vm.SetTemporary(0, values.NewFloat(3.14))
	if val, ok := vm.GetTemporary(0); !ok || !val.Equal(values.NewFloat(3.14)) {
		t.Errorf("GetTemporary failed")
	}
}

func TestClassManager_BasicOperations(t *testing.T) {
	cm := NewClassManager()

	// Test creating and retrieving classes
	cls := &classRuntime{
		Name:        "TestClass",
		Properties:  make(map[string]*propertyRuntime),
		StaticProps: make(map[string]*values.Value),
		Constants:   make(map[string]*values.Value),
	}

	cm.StoreClass("TestClass", cls)
	if retrieved, ok := cm.GetClass("TestClass"); !ok || retrieved.Name != "TestClass" {
		t.Errorf("GetClass failed")
	}

	// Test current class operations
	cm.SetCurrentClass(cls)
	if current := cm.GetCurrentClass(); current == nil || current.Name != "TestClass" {
		t.Errorf("GetCurrentClass failed")
	}

	cm.ClearCurrentClass()
	if current := cm.GetCurrentClass(); current != nil {
		t.Errorf("ClearCurrentClass failed")
	}
}

func TestCallStackManager_BasicOperations(t *testing.T) {
	cs := NewCallStackManager()

	// Test empty stack
	if !cs.IsEmpty() {
		t.Errorf("IsEmpty should return true for new stack")
	}
	if cs.Depth() != 0 {
		t.Errorf("Depth should be 0 for new stack")
	}

	// Test pushing frames
	frame1 := newCallFrame("test1", nil, nil, nil)
	frame2 := newCallFrame("test2", nil, nil, nil)

	cs.PushFrame(frame1)
	if cs.IsEmpty() || cs.Depth() != 1 {
		t.Errorf("PushFrame failed")
	}

	cs.PushFrame(frame2)
	if cs.Depth() != 2 {
		t.Errorf("Second PushFrame failed")
	}

	// Test current frame
	if current := cs.CurrentFrame(); current == nil || current.FunctionName != "test2" {
		t.Errorf("CurrentFrame should return the top frame")
	}

	// Test popping frames
	popped := cs.PopFrame()
	if popped == nil || popped.FunctionName != "test2" {
		t.Errorf("PopFrame should return the top frame")
	}
	if cs.Depth() != 1 {
		t.Errorf("Depth should be 1 after popping")
	}

	// Test popping last frame
	popped = cs.PopFrame()
	if popped == nil || popped.FunctionName != "test1" {
		t.Errorf("PopFrame should return the remaining frame")
	}
	if !cs.IsEmpty() {
		t.Errorf("Stack should be empty after popping all frames")
	}

	// Test popping from empty stack
	if popped := cs.PopFrame(); popped != nil {
		t.Errorf("PopFrame should return nil for empty stack")
	}
}

func TestExecutionContextV2_Integration(t *testing.T) {
	ctx := NewExecutionContextV2()

	// Test variable operations
	ctx.setVariable("test_var", values.NewString("test_value"))
	if val, ok := ctx.GetVariable("test_var"); !ok || !val.Equal(values.NewString("test_value")) {
		t.Errorf("Variable operations failed")
	}

	// Test global variable operations
	ctx.SetGlobalVar("global_var", values.NewInt(123))
	if val, ok := ctx.GetGlobalVar("global_var"); !ok || !val.Equal(values.NewInt(123)) {
		t.Errorf("Global variable operations failed")
	}

	// Test call stack operations
	frame := newCallFrame("test_function", nil, nil, nil)
	ctx.pushFrame(frame)

	if current := ctx.currentFrame(); current == nil || current.FunctionName != "test_function" {
		t.Errorf("Call stack operations failed")
	}

	popped := ctx.popFrame()
	if popped == nil || popped.FunctionName != "test_function" {
		t.Errorf("Frame popping failed")
	}

	// Test class operations
	cls := &classRuntime{
		Name:        "TestClass",
		Properties:  make(map[string]*propertyRuntime),
		StaticProps: make(map[string]*values.Value),
		Constants:   make(map[string]*values.Value),
	}

	ctx.SetCurrentClass(cls)
	if current := ctx.GetCurrentClass(); current == nil || current.Name != "TestClass" {
		t.Errorf("Class operations failed")
	}

	// Test debug operations
	ctx.appendDebugRecord("test debug message")
	records := ctx.drainDebugRecords()
	if len(records) != 1 || records[0] != "test debug message" {
		t.Errorf("Debug operations failed")
	}
}

func TestExecutionContextV2_Copy(t *testing.T) {
	// Test variable manager copy
	vm := NewVariableManager()
	vm.SetGlobalVar("global", values.NewString("test"))
	vm.SetVariable("local", values.NewInt(42))
	vm.SetTemporary(0, values.NewFloat(3.14))

	vmCopy := vm.Copy()

	// Verify all values were copied
	if val, ok := vmCopy.GetGlobalVar("global"); !ok || !val.Equal(values.NewString("test")) {
		t.Errorf("Global variable not copied correctly")
	}
	if val, ok := vmCopy.GetVariable("local"); !ok || !val.Equal(values.NewInt(42)) {
		t.Errorf("Local variable not copied correctly")
	}
	if val, ok := vmCopy.GetTemporary(0); !ok || !val.Equal(values.NewFloat(3.14)) {
		t.Errorf("Temporary variable not copied correctly")
	}

	// Test class manager copy
	cm := NewClassManager()
	cls := &classRuntime{
		Name:        "TestClass",
		Properties:  make(map[string]*propertyRuntime),
		StaticProps: make(map[string]*values.Value),
		Constants:   make(map[string]*values.Value),
	}
	cm.StoreClass("TestClass", cls)
	cm.SetCurrentClass(cls)

	cmCopy := cm.Copy()

	if retrieved, ok := cmCopy.GetClass("TestClass"); !ok || retrieved.Name != "TestClass" {
		t.Errorf("Class not copied correctly")
	}
	if current := cmCopy.GetCurrentClass(); current == nil || current.Name != "TestClass" {
		t.Errorf("Current class not copied correctly")
	}
}

func TestExecutionContextV2_Clear(t *testing.T) {
	// Test variable manager clear
	vm := NewVariableManager()
	vm.SetGlobalVar("global", values.NewString("test"))
	vm.SetVariable("local", values.NewInt(42))
	vm.SetTemporary(0, values.NewFloat(3.14))

	vm.Clear()

	if _, ok := vm.GetGlobalVar("global"); ok {
		t.Errorf("Global variable not cleared")
	}
	if _, ok := vm.GetVariable("local"); ok {
		t.Errorf("Local variable not cleared")
	}
	if _, ok := vm.GetTemporary(0); ok {
		t.Errorf("Temporary variable not cleared")
	}

	// Test call stack clear
	cs := NewCallStackManager()
	cs.PushFrame(newCallFrame("test", nil, nil, nil))

	cs.Clear()

	if !cs.IsEmpty() {
		t.Errorf("Call stack not cleared")
	}
}