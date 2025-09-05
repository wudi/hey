package main

import (
	"fmt"
	"time"

	"github.com/wudi/php-parser/compiler/runtime"
	"github.com/wudi/php-parser/compiler/values"
)

func main() {
	// Initialize runtime system
	err := runtime.Bootstrap()
	if err != nil {
		panic(fmt.Sprintf("Failed to bootstrap runtime: %v", err))
	}

	// Initialize VM integration
	err = runtime.InitializeVMIntegration()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize VM integration: %v", err))
	}

	// Create and register the Go extension
	goExt := runtime.NewGoExtension()
	
	// Create extension manager and register the extension
	extManager := runtime.NewExtensionManager(runtime.GlobalRegistry)
	err = extManager.RegisterExtension(goExt)
	if err != nil {
		panic(fmt.Sprintf("Failed to register Go extension: %v", err))
	}

	// Load the extension
	err = extManager.LoadExtension("go")
	if err != nil {
		panic(fmt.Sprintf("Failed to load Go extension: %v", err))
	}

	fmt.Println("Go extension registered successfully!")

	// Example 1: Create a simple closure and execute it with go()
	fmt.Println("\n=== Example 1: Simple closure ===")
	
	// Create a simple Go function as a closure
	simpleHandler := func(ctx runtime.ExecutionContext, args []*values.Value) (*values.Value, error) {
		fmt.Println("Hello from goroutine!")
		time.Sleep(100 * time.Millisecond) // Simulate work
		return values.NewString("completed"), nil
	}
	
	simpleClosure := values.NewClosure(simpleHandler, nil, "simple_task")
	
	// Call go() function with the closure
	goFunc, exists := runtime.GlobalRegistry.GetFunction("go")
	if !exists {
		panic("go() function not found")
	}
	
	result, err := goFunc.Handler(nil, []*values.Value{simpleClosure})
	if err != nil {
		panic(fmt.Sprintf("Failed to call go(): %v", err))
	}
	
	fmt.Printf("go() returned: %s\n", result.String())

	// Example 2: Closure with bound variables
	fmt.Println("\n=== Example 2: Closure with bound variables ===")
	
	boundVars := map[string]*values.Value{
		"message": values.NewString("Hello from bound variable!"),
		"counter": values.NewInt(42),
	}
	
	boundHandler := func(ctx runtime.ExecutionContext, args []*values.Value) (*values.Value, error) {
		fmt.Println("Executing closure with bound variables...")
		// In a real implementation, you'd access bound variables here
		time.Sleep(200 * time.Millisecond)
		return values.NewString("bound_task_completed"), nil
	}
	
	boundClosure := values.NewClosure(boundHandler, boundVars, "bound_task")
	
	result, err = goFunc.Handler(nil, []*values.Value{boundClosure})
	if err != nil {
		panic(fmt.Sprintf("Failed to call go() with bound closure: %v", err))
	}
	
	fmt.Printf("go() with bound variables returned: %s\n", result.String())

	// Example 3: Multiple concurrent executions
	fmt.Println("\n=== Example 3: Multiple concurrent executions ===")
	
	for i := 0; i < 3; i++ {
		taskID := i
		taskHandler := func(ctx runtime.ExecutionContext, args []*values.Value) (*values.Value, error) {
			fmt.Printf("Task %d started in goroutine\n", taskID)
			time.Sleep(time.Duration(100*(taskID+1)) * time.Millisecond)
			fmt.Printf("Task %d completed\n", taskID)
			return values.NewInt(int64(taskID)), nil
		}
		
		taskClosure := values.NewClosure(taskHandler, nil, fmt.Sprintf("task_%d", taskID))
		
		result, err := goFunc.Handler(nil, []*values.Value{taskClosure})
		if err != nil {
			fmt.Printf("Failed to start task %d: %v\n", taskID, err)
		} else {
			fmt.Printf("Task %d started, go() returned: %s\n", taskID, result.String())
		}
	}

	// Wait for all goroutines to complete
	fmt.Println("\nWaiting for all goroutines to complete...")
	goExt.WaitForAll()
	fmt.Println("All goroutines completed!")

	// Example 4: Error handling
	fmt.Println("\n=== Example 4: Error handling ===")
	
	// Try to call go() with non-callable value
	nonCallable := values.NewString("not a function")
	result, err = goFunc.Handler(nil, []*values.Value{nonCallable})
	if err != nil {
		fmt.Printf("Expected error for non-callable: %v\n", err)
	}

	// Try to call go() with wrong number of arguments
	result, err = goFunc.Handler(nil, []*values.Value{})
	if err != nil {
		fmt.Printf("Expected error for no arguments: %v\n", err)
	}

	result, err = goFunc.Handler(nil, []*values.Value{simpleClosure, simpleClosure})
	if err != nil {
		fmt.Printf("Expected error for too many arguments: %v\n", err)
	}

	fmt.Println("\nExample completed successfully!")
}