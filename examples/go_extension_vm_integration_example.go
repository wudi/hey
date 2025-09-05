package main

import (
	"fmt"

	"github.com/wudi/php-parser/compiler/opcodes"
	"github.com/wudi/php-parser/compiler/runtime"
	"github.com/wudi/php-parser/compiler/values"
	"github.com/wudi/php-parser/compiler/vm"
)

// VMClosureExecutor implements the ClosureExecutor interface for VM integration
type VMClosureExecutor struct {
	vm *vm.VirtualMachine
}

func (vce *VMClosureExecutor) ExecuteClosure(ctx interface{}, closure *values.Closure, args []*values.Value) (*values.Value, error) {
	// Convert the runtime context to VM execution context if needed
	var vmCtx *vm.ExecutionContext
	
	if vmExecutionCtx, ok := ctx.(*vm.ExecutionContext); ok {
		vmCtx = vmExecutionCtx
	} else {
		// Create a new VM context for this execution
		vmCtx = vm.NewExecutionContext()
	}
	
	return vce.vm.ExecuteClosure(vmCtx, closure, args)
}

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

	// Create VM and Go extension with VM integration
	vmInstance := vm.NewVirtualMachine()
	goExt := runtime.NewGoExtension()
	
	// Set up VM closure executor to enable proper VM-based closure execution
	vmExecutor := &VMClosureExecutor{vm: vmInstance}
	goExt.SetClosureExecutor(vmExecutor)
	
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

	fmt.Println("Go extension with VM closure execution registered successfully!")

	// Get the go function
	goFunc, exists := runtime.GlobalRegistry.GetFunction("go")
	if !exists {
		panic("go() function not found")
	}

	// Example 1: Runtime function handler with VM integration
	fmt.Println("\n=== Example 1: Runtime Function Handler ===")
	
	runtimeHandler := func(ctx runtime.ExecutionContext, args []*values.Value) (*values.Value, error) {
		fmt.Println("Executing runtime function handler in goroutine")
		if len(args) > 0 {
			fmt.Printf("  First argument: %s\n", args[0].ToString())
		}
		return values.NewString("runtime_handler_result"), nil
	}
	
	runtimeClosure := values.NewClosure(runtimeHandler, nil, "runtime_function")
	
	result, err := goFunc.Handler(nil, []*values.Value{runtimeClosure})
	if err != nil {
		panic(fmt.Sprintf("Failed to call go() with runtime handler: %v", err))
	}
	fmt.Printf("go() returned: %s\n", result.String())

	// Example 2: VM function simulation (would be a real VM function in practice)
	fmt.Println("\n=== Example 2: VM Function Simulation ===")
	
	// Create a VM Function struct (normally compiled from PHP code)
	vmFunction := &vm.Function{
		Name:         "test_vm_function",
		Instructions: []opcodes.Instruction{}, // Would contain actual bytecode
		Constants:    []*values.Value{values.NewString("VM function result")},
		Parameters: []vm.Parameter{
			{Name: "param1", Type: "string", HasDefault: false},
		},
		IsVariadic:  false,
		IsGenerator: false,
	}
	
	vmClosure := values.NewClosure(vmFunction, nil, "vm_function")
	
	result, err = goFunc.Handler(nil, []*values.Value{vmClosure})
	if err != nil {
		fmt.Printf("Expected VM execution to fail gracefully: %v\n", err)
	} else {
		fmt.Printf("go() returned: %s\n", result.String())
	}

	// Example 3: String-based function name
	fmt.Println("\n=== Example 3: String-based Function Name ===")
	
	stringClosure := values.NewClosure("strlen", nil, "string_function")
	
	result, err = goFunc.Handler(nil, []*values.Value{stringClosure})
	if err != nil {
		panic(fmt.Sprintf("Failed to call go() with string function: %v", err))
	}
	fmt.Printf("go() returned: %s\n", result.String())

	// Example 4: Closure with bound variables
	fmt.Println("\n=== Example 4: Closure with Bound Variables ===")
	
	boundVars := map[string]*values.Value{
		"message": values.NewString("Hello from bound variable"),
		"count":   values.NewInt(42),
	}
	
	boundHandler := func(ctx runtime.ExecutionContext, args []*values.Value) (*values.Value, error) {
		fmt.Println("Executing bound variable handler")
		// The VM would provide access to bound variables
		return values.NewString("bound_variable_result"), nil
	}
	
	boundClosure := values.NewClosure(boundHandler, boundVars, "bound_function")
	
	result, err = goFunc.Handler(nil, []*values.Value{boundClosure})
	if err != nil {
		panic(fmt.Sprintf("Failed to call go() with bound closure: %v", err))
	}
	fmt.Printf("go() returned: %s\n", result.String())

	// Example 5: Legacy function without context
	fmt.Println("\n=== Example 5: Legacy Function Handler ===")
	
	legacyHandler := func(args []*values.Value) (*values.Value, error) {
		fmt.Println("Executing legacy function handler")
		return values.NewString("legacy_result"), nil
	}
	
	legacyClosure := values.NewClosure(legacyHandler, nil, "legacy_function")
	
	result, err = goFunc.Handler(nil, []*values.Value{legacyClosure})
	if err != nil {
		panic(fmt.Sprintf("Failed to call go() with legacy handler: %v", err))
	}
	fmt.Printf("go() returned: %s\n", result.String())

	// Example 6: Reflection fallback for unknown types
	fmt.Println("\n=== Example 6: Reflection Fallback ===")
	
	reflectionFunc := func(message string) string {
		return "Reflected: " + message
	}
	
	reflectionClosure := values.NewClosure(reflectionFunc, nil, "reflection_function")
	
	result, err = goFunc.Handler(nil, []*values.Value{reflectionClosure})
	if err != nil {
		panic(fmt.Sprintf("Failed to call go() with reflection function: %v", err))
	}
	fmt.Printf("go() returned: %s\n", result.String())

	// Wait for all goroutines to complete
	fmt.Println("\nWaiting for all goroutines to complete...")
	goExt.WaitForAll()
	fmt.Println("All goroutines completed!")

	// Test direct VM closure execution
	fmt.Println("\n=== Example 7: Direct VM Closure Execution ===")
	
	directHandler := func(ctx runtime.ExecutionContext, args []*values.Value) (*values.Value, error) {
		return values.NewString("direct_vm_result"), nil
	}
	
	closureValue := values.NewClosure(directHandler, nil, "direct_test")
	
	// Test direct VM call
	vmResult, err := vmInstance.CallClosure(closureValue, []*values.Value{
		values.NewString("test argument"),
	})
	if err != nil {
		fmt.Printf("Direct VM call failed: %v\n", err)
	} else {
		fmt.Printf("Direct VM call result: %s\n", vmResult.String())
	}

	fmt.Println("\nVM-integrated Go extension example completed successfully!")
}