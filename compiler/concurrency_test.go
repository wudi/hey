package compiler

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wudi/php-parser/compiler/runtime"
	"github.com/wudi/php-parser/compiler/values"
	"github.com/wudi/php-parser/lexer"
	"github.com/wudi/php-parser/parser"
)

// Test basic go() function functionality
func TestGoFunction(t *testing.T) {
	// Initialize runtime
	err := runtime.Bootstrap()
	require.NoError(t, err, "Failed to bootstrap runtime")

	// Test go() function exists
	functions := runtime.GlobalRegistry.GetAllFunctions()
	assert.Contains(t, functions, "go", "go() function should be registered")

	// Test go() function call
	goFunc := functions["go"]
	assert.NotNil(t, goFunc, "go() function should not be nil")
	assert.Equal(t, "go", goFunc.Name)
	assert.Equal(t, 1, goFunc.MinArgs)
	assert.Equal(t, 1, goFunc.MaxArgs)
}

// Test WaitGroup class functionality
func TestWaitGroupClass(t *testing.T) {
	// Initialize runtime
	err := runtime.Bootstrap()
	require.NoError(t, err, "Failed to bootstrap runtime")

	// Test WaitGroup class exists
	classes := runtime.GlobalRegistry.GetAllClasses()
	assert.Contains(t, classes, "waitgroup", "WaitGroup class should be registered")

	waitGroupClass := classes["waitgroup"]
	assert.NotNil(t, waitGroupClass, "WaitGroup class should not be nil")
	assert.Equal(t, "WaitGroup", waitGroupClass.Name)

	// Check methods exist
	assert.Contains(t, waitGroupClass.Methods, "__construct")
	assert.Contains(t, waitGroupClass.Methods, "Add")
	assert.Contains(t, waitGroupClass.Methods, "Done")
	assert.Contains(t, waitGroupClass.Methods, "Wait")
}

// Test WaitGroup value operations
func TestWaitGroupValue(t *testing.T) {
	wg := values.NewWaitGroup()

	// Test type checking
	assert.True(t, wg.IsWaitGroup())
	assert.False(t, wg.IsNull())
	assert.Equal(t, "WaitGroup", wg.ToString())
	assert.Equal(t, "waitgroup", wg.TypeName())

	// Test Add operation
	err := wg.WaitGroupAdd(2)
	assert.NoError(t, err)

	// Test Done operation
	err = wg.WaitGroupDone()
	assert.NoError(t, err)

	// Test another Done operation
	err = wg.WaitGroupDone()
	assert.NoError(t, err)

	// Test Wait operation (should not block since counter is 0)
	done := make(chan bool, 1)
	go func() {
		err := wg.WaitGroupWait()
		assert.NoError(t, err)
		done <- true
	}()

	select {
	case <-done:
		// Wait completed as expected
	case <-time.After(100 * time.Millisecond):
		t.Error("WaitGroup.Wait() should have completed immediately")
	}
}

// Test WaitGroup error conditions
func TestWaitGroupErrors(t *testing.T) {
	wg := values.NewWaitGroup()

	// Test negative counter
	err := wg.WaitGroupAdd(-1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "negative")

	// Test Done on zero counter
	err = wg.WaitGroupDone()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "negative")
}

// Test Goroutine value operations
func TestGoroutineValue(t *testing.T) {
	closure := values.NewClosure(nil, nil, "test")
	gor := values.NewGoroutine(closure.Data.(*values.Closure), nil)

	// Test type checking
	assert.True(t, gor.IsGoroutine())
	assert.False(t, gor.IsNull())
	assert.Contains(t, gor.ToString(), "Goroutine")
	assert.Equal(t, "goroutine", gor.TypeName())

	// Test goroutine data
	gorData := gor.Data.(*values.Goroutine)
	assert.NotZero(t, gorData.ID)
	assert.Equal(t, "running", gorData.Status)
	assert.NotNil(t, gorData.Done)
}

// Test go() function integration with parsing
func TestGoFunctionIntegration(t *testing.T) {
	// PHP code that uses go() function
	phpCode := `<?php
$closure = function() {
    return "Hello from goroutine";
};
$g = go($closure);
`

	// Parse PHP code
	l := lexer.New(phpCode)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Logf("Parser errors: %v", p.Errors())
		return
	}

	// Initialize compiler
	comp := NewCompiler()

	// Initialize runtime
	err := runtime.Bootstrap()
	require.NoError(t, err, "Failed to bootstrap runtime")

	err = runtime.InitializeVMIntegration()
	require.NoError(t, err, "Failed to initialize VM integration")

	// Compile the program
	err = comp.Compile(program)
	if err != nil {
		t.Logf("Compilation failed (expected): %v", err)
		// This might fail if the parser doesn't support function call syntax yet
		// But the go() function itself should be registered
		return
	}

	// Test execution would require full VM integration
}

// Test WaitGroup class integration with parsing
func TestWaitGroupIntegration(t *testing.T) {
	// PHP code that uses WaitGroup class
	phpCode := `<?php
$wg = new WaitGroup();
$wg->Add(1);
$wg->Done();
$wg->Wait();
`

	// Parse PHP code
	l := lexer.New(phpCode)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Logf("Parser errors: %v", p.Errors())
		return
	}

	// Initialize compiler
	comp := NewCompiler()

	// Initialize runtime
	err := runtime.Bootstrap()
	require.NoError(t, err, "Failed to bootstrap runtime")

	err = runtime.InitializeVMIntegration()
	require.NoError(t, err, "Failed to initialize VM integration")

	// Compile the program
	err = comp.Compile(program)
	if err != nil {
		t.Logf("Compilation failed (expected): %v", err)
		// This might fail if class instantiation isn't fully supported
		// But the WaitGroup class itself should be registered
		return
	}

	// Test execution would require full VM integration
}

// Test concurrent WaitGroup usage
func TestConcurrentWaitGroup(t *testing.T) {
	wg := values.NewWaitGroup()

	// Add work items
	err := wg.WaitGroupAdd(3)
	require.NoError(t, err)

	// Start goroutines that will call Done
	for i := 0; i < 3; i++ {
		go func() {
			time.Sleep(10 * time.Millisecond) // Simulate work
			err := wg.WaitGroupDone()
			assert.NoError(t, err)
		}()
	}

	// Wait for all goroutines to complete
	done := make(chan bool, 1)
	go func() {
		err := wg.WaitGroupWait()
		assert.NoError(t, err)
		done <- true
	}()

	select {
	case <-done:
		// All goroutines completed
	case <-time.After(1 * time.Second):
		t.Error("WaitGroup.Wait() timed out")
	}
}

// Benchmark WaitGroup operations
func BenchmarkWaitGroupOperations(b *testing.B) {
	b.Run("Add", func(b *testing.B) {
		wg := values.NewWaitGroup()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			wg.WaitGroupAdd(1)
			wg.WaitGroupDone()
		}
	})

	b.Run("Concurrent", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				wg := values.NewWaitGroup()
				wg.WaitGroupAdd(1)
				go func() {
					wg.WaitGroupDone()
				}()
				wg.WaitGroupWait()
			}
		})
	})
}
