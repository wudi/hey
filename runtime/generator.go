package runtime

import (
	"fmt"
	"strconv"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// GeneratorExecutor is used to execute generator functions with yield support
type GeneratorExecutor struct {
	generator *Generator
}

// Generator implements PHP generators with proper execution suspension/resumption
type Generator struct {
	function     *registry.Function
	args         []*values.Value
	vm           interface{} // VM interface to avoid import cycles

	// Generator state
	started      bool
	finished     bool
	suspended    bool
	currentKey   *values.Value
	currentValue *values.Value

	// Yield from delegation state
	delegating        bool
	delegateIterable  *values.Value
	delegateKeys      []string // For array iteration
	delegateIndex     int      // Current index in array
	delegateGenerator *Generator // For generator delegation

	// Suspended execution state
	suspendedContext *GeneratorExecutionState
}

// GeneratorExecutionState preserves VM execution state at yield points
type GeneratorExecutionState struct {
	frame interface{} // Actual CallFrame object
	ctx   interface{} // Actual ExecutionContext object
}

// NewGenerator creates a new generator
func NewGenerator(function *registry.Function, args []*values.Value, vm interface{}) *Generator {
	return &Generator{
		function:         function,
		args:             args,
		vm:               vm,
		started:          false,
		finished:         false,
		suspended:        false,
		currentKey:       values.NewNull(),
		currentValue:     values.NewNull(),
		delegating:       false,
		delegateIterable: nil,
		delegateKeys:     nil,
		delegateIndex:    0,
		delegateGenerator: nil,
		suspendedContext: nil,
	}
}

// NewChannelGenerator creates a new channel-based generator (DEPRECATED - for compatibility)
func NewChannelGenerator(function interface{}, args []*values.Value, vm interface{}) *Generator {
	fn, ok := function.(*registry.Function)
	if !ok {
		return nil
	}
	return NewGenerator(fn, args, vm)
}

// Next advances the generator to the next value
func (g *Generator) Next() bool {
	if g.finished {
		return false
	}

	// Check if we're delegating to another iterable
	if g.delegating {
		return g.handleDelegateNext()
	}

	if !g.started {
		// First call - start execution from beginning
		g.started = true
		return g.executeUntilYield()
	} else if g.suspended {
		// Resume from suspended state
		return g.resumeFromYield()
	}

	return false
}

// executeUntilYield starts generator execution from the beginning
func (g *Generator) executeUntilYield() bool {
	vm, ok := g.vm.(interface {
		CreateExecutionContext() interface{}
		CreateCallFrame(*registry.Function, []*values.Value) interface{}
		ExecuteUntilYield(interface{}, interface{}) (bool, error)
	})
	if !ok {
		return false
	}

	// Create fresh execution context and call frame
	ctx := vm.CreateExecutionContext()
	frame := vm.CreateCallFrame(g.function, g.args)

	// Set generator reference in call frame
	if frameTyped, ok := frame.(interface{ SetGenerator(interface{}) }); ok {
		frameTyped.SetGenerator(g)
	}

	// Execute until first yield
	yielded, err := vm.ExecuteUntilYield(ctx, frame)
	if err != nil {
		g.finished = true
		return false
	}

	if !yielded {
		// Function completed without yield
		g.finished = true
		return false
	}

	// Save execution state for resumption
	g.saveExecutionState(ctx, frame)
	g.suspended = true
	return true
}

// resumeFromYield resumes generator execution from saved state
func (g *Generator) resumeFromYield() bool {
	if g.suspendedContext == nil {
		return false
	}

	vm, ok := g.vm.(interface {
		ResumeFromYield(interface{}, interface{}) (bool, error)
	})
	if !ok {
		return false
	}

	// Restore execution state
	ctx, frame := g.restoreExecutionState()

	// Resume execution until next yield
	yielded, err := vm.ResumeFromYield(ctx, frame)
	if err != nil {
		g.finished = true
		return false
	}

	if !yielded {
		// Function completed
		g.finished = true
		g.suspended = false
		g.suspendedContext = nil
		return false
	}

	// Update saved state for next resumption
	g.saveExecutionState(ctx, frame)
	return true
}

// saveExecutionState preserves VM state for resumption
func (g *Generator) saveExecutionState(ctx, frame interface{}) {
	// Store the actual execution context and call frame objects
	// These will be reused for resumption
	g.suspendedContext = &GeneratorExecutionState{
		frame: frame,
		ctx:   ctx,
	}
}

// restoreExecutionState recreates VM state from saved state
func (g *Generator) restoreExecutionState() (interface{}, interface{}) {
	if g.suspendedContext == nil {
		return nil, nil
	}
	return g.suspendedContext.ctx, g.suspendedContext.frame
}

// Current returns the current value
func (g *Generator) Current() *values.Value {
	return g.currentValue
}

// Key returns the current key
func (g *Generator) Key() *values.Value {
	return g.currentKey
}

// Valid returns whether the generator has more values
func (g *Generator) Valid() bool {
	return !g.finished && g.started
}

// Rewind resets the generator (not supported for generators)
func (g *Generator) Rewind() error {
	if g.started {
		return fmt.Errorf("Cannot rewind a generator that was already run")
	}
	return nil
}

// Yield is called from within the generator function to yield a value
func (g *Generator) Yield(key, value *values.Value) {
	// Store the yielded values for retrieval
	g.currentKey = key
	g.currentValue = value
	// Actual suspension logic will be handled by VM.execYield
}

// StartDelegation begins delegating to another iterable
func (g *Generator) StartDelegation(iterable *values.Value) error {
	g.delegating = true
	g.delegateIterable = iterable
	g.delegateIndex = 0
	g.suspended = true // Mark as suspended so Next() will handle delegation

	if iterable.IsArray() {
		// Prepare array keys for iteration
		if arr, ok := iterable.Data.(*values.Array); ok {
			g.delegateKeys = make([]string, 0, len(arr.Elements))
			for key := range arr.Elements {
				keyStr := fmt.Sprintf("%v", key)
				g.delegateKeys = append(g.delegateKeys, keyStr)
			}
		}

		// Set the first value immediately if array is not empty
		if len(g.delegateKeys) > 0 {
			keyStr := g.delegateKeys[0]
			arr := g.delegateIterable.Data.(*values.Array)

			// Convert string back to interface{} key for lookup
			var key interface{}
			if intKey, err := strconv.Atoi(keyStr); err == nil {
				key = int64(intKey)
			} else {
				key = keyStr
			}

			value := arr.Elements[key]

			// Convert to appropriate Value type
			var keyValue *values.Value
			if intKey, err := strconv.Atoi(keyStr); err == nil {
				keyValue = values.NewInt(int64(intKey))
			} else {
				keyValue = values.NewString(keyStr)
			}

			// Set current values and advance index
			g.currentKey = keyValue
			g.currentValue = value
			g.delegateIndex++
		}
	} else if iterable.IsObject() && iterable.Data.(*values.Object).ClassName == "Generator" {
		// Get the delegate generator
		obj := iterable.Data.(*values.Object)
		if genVal, ok := obj.Properties["__channel_generator"]; ok {
			if delegateGen, ok := genVal.Data.(*Generator); ok {
				g.delegateGenerator = delegateGen
				// Get the first value from the delegate generator
				if g.delegateGenerator.Next() {
					g.currentKey = g.delegateGenerator.Key()
					g.currentValue = g.delegateGenerator.Current()
				} else {
				}
			} else {
				return fmt.Errorf("invalid generator for delegation")
			}
		} else {
			return fmt.Errorf("generator object missing __channel_generator property")
		}
	} else {
		return fmt.Errorf("yield from requires an iterable (array or Generator)")
	}

	return nil
}

// handleDelegateNext handles the next iteration when delegating
func (g *Generator) handleDelegateNext() bool {
	if g.delegateIterable.IsArray() {
		// Array delegation
		if g.delegateIndex >= len(g.delegateKeys) {
			// Array exhausted, stop delegating and resume normal execution
			g.delegating = false
			return g.resumeFromYield()
		}

		// Get current array item
		keyStr := g.delegateKeys[g.delegateIndex]
		arr := g.delegateIterable.Data.(*values.Array)

		// Convert string back to interface{} key for lookup
		var key interface{}
		if intKey, err := strconv.Atoi(keyStr); err == nil {
			key = int64(intKey)
		} else {
			key = keyStr
		}

		value := arr.Elements[key]

		// Convert to appropriate Value type
		var keyValue *values.Value
		if intKey, err := strconv.Atoi(keyStr); err == nil {
			keyValue = values.NewInt(int64(intKey))
		} else {
			keyValue = values.NewString(keyStr)
		}

		// Set current values and advance index
		g.currentKey = keyValue
		g.currentValue = value
		g.delegateIndex++

		return true
	} else if g.delegateGenerator != nil {
		// Generator delegation
		if g.delegateGenerator.Next() {
			// Delegate generator has a value
			g.currentKey = g.delegateGenerator.Key()
			g.currentValue = g.delegateGenerator.Current()
			return true
		} else {
			// Delegate generator is exhausted, stop delegating and resume normal execution
			g.delegating = false
			return g.resumeFromYield()
		}
	}

	return false
}