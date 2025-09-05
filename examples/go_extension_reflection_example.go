package main

import (
	"fmt"
	"strings"

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

	fmt.Println("Go extension with reflection capabilities registered successfully!")

	// Get the go function
	goFunc, exists := runtime.GlobalRegistry.GetFunction("go")
	if !exists {
		panic("go() function not found")
	}

	// Example 1: Simple function with string parameter and return
	fmt.Println("\n=== Example 1: Simple String Function ===")
	
	stringFunc := func(name string) string {
		return "Hello, " + name + "!"
	}
	
	stringClosure := values.NewClosure(stringFunc, nil, "string_function")
	
	result, err := goFunc.Handler(nil, []*values.Value{stringClosure})
	if err != nil {
		panic(fmt.Sprintf("Failed to call go(): %v", err))
	}
	fmt.Printf("go() returned: %s\n", result.String())

	// Example 2: Function with multiple parameters and return values
	fmt.Println("\n=== Example 2: Multiple Parameters and Return Values ===")
	
	mathFunc := func(a, b int) (int, int, string) {
		sum := a + b
		product := a * b
		operation := fmt.Sprintf("%d + %d = %d, %d * %d = %d", a, b, sum, a, b, product)
		return sum, product, operation
	}
	
	mathClosure := values.NewClosure(mathFunc, nil, "math_function")
	
	result, err = goFunc.Handler(nil, []*values.Value{mathClosure})
	if err != nil {
		panic(fmt.Sprintf("Failed to call go(): %v", err))
	}
	fmt.Printf("go() returned: %s\n", result.String())

	// Example 3: Function returning error
	fmt.Println("\n=== Example 3: Function with Error Return ===")
	
	errorFunc := func(shouldFail bool) (string, error) {
		if shouldFail {
			return "", fmt.Errorf("intentional failure")
		}
		return "success", nil
	}
	
	// Test success case
	successClosure := values.NewClosure(errorFunc, nil, "error_function_success")
	result, err = goFunc.Handler(nil, []*values.Value{successClosure})
	if err != nil {
		panic(fmt.Sprintf("Failed to call go(): %v", err))
	}
	fmt.Printf("Success case - go() returned: %s\n", result.String())

	// Example 4: Function with various types
	fmt.Println("\n=== Example 4: Various Parameter Types ===")
	
	typedFunc := func(str string, num int64, flag bool, flt float64) map[string]interface{} {
		return map[string]interface{}{
			"string":    str,
			"number":    num,
			"boolean":   flag,
			"float":     flt,
			"processed": strings.ToUpper(str),
		}
	}
	
	typedClosure := values.NewClosure(typedFunc, nil, "typed_function")
	
	result, err = goFunc.Handler(nil, []*values.Value{typedClosure})
	if err != nil {
		panic(fmt.Sprintf("Failed to call go(): %v", err))
	}
	fmt.Printf("go() returned: %s\n", result.String())

	// Example 5: Function returning slice
	fmt.Println("\n=== Example 5: Slice Return Value ===")
	
	sliceFunc := func(count int) []string {
		result := make([]string, count)
		for i := 0; i < count; i++ {
			result[i] = fmt.Sprintf("item_%d", i+1)
		}
		return result
	}
	
	sliceClosure := values.NewClosure(sliceFunc, nil, "slice_function")
	
	result, err = goFunc.Handler(nil, []*values.Value{sliceClosure})
	if err != nil {
		panic(fmt.Sprintf("Failed to call go(): %v", err))
	}
	fmt.Printf("go() returned: %s\n", result.String())

	// Example 6: Void function
	fmt.Println("\n=== Example 6: Void Function ===")
	
	voidFunc := func(message string) {
		fmt.Printf("  Void function executed with message: '%s'\n", message)
	}
	
	voidClosure := values.NewClosure(voidFunc, nil, "void_function")
	
	result, err = goFunc.Handler(nil, []*values.Value{voidClosure})
	if err != nil {
		panic(fmt.Sprintf("Failed to call go(): %v", err))
	}
	fmt.Printf("go() returned: %s\n", result.String())

	// Example 7: String-based callable
	fmt.Println("\n=== Example 7: String-based Callable ===")
	
	stringCallable := values.NewClosure("my_function_name", nil, "string_callable")
	
	result, err = goFunc.Handler(nil, []*values.Value{stringCallable})
	if err != nil {
		panic(fmt.Sprintf("Failed to call go(): %v", err))
	}
	fmt.Printf("go() returned: %s\n", result.String())

	// Example 8: Map-based callable (PHP-style callable array)
	fmt.Println("\n=== Example 8: Map-based Callable ===")
	
	mapCallable := values.NewClosure(
		map[string]string{
			"class":  "MyClass", 
			"method": "myMethod",
		}, 
		nil, 
		"map_callable",
	)
	
	result, err = goFunc.Handler(nil, []*values.Value{mapCallable})
	if err != nil {
		panic(fmt.Sprintf("Failed to call go(): %v", err))
	}
	fmt.Printf("go() returned: %s\n", result.String())

	// Example 9: Function with complex return type (struct-like)
	fmt.Println("\n=== Example 9: Complex Return Type ===")
	
	type Person struct {
		Name string
		Age  int
		Tags []string
	}
	
	structFunc := func(name string, age int) Person {
		return Person{
			Name: name,
			Age:  age,
			Tags: []string{"golang", "reflection", "test"},
		}
	}
	
	structClosure := values.NewClosure(structFunc, nil, "struct_function")
	
	result, err = goFunc.Handler(nil, []*values.Value{structClosure})
	if err != nil {
		panic(fmt.Sprintf("Failed to call go(): %v", err))
	}
	fmt.Printf("go() returned: %s\n", result.String())

	// Wait for all goroutines to complete
	fmt.Println("\nWaiting for all goroutines to complete...")
	goExt.WaitForAll()
	fmt.Println("All goroutines completed!")

	fmt.Println("\nReflection-based callable invocation examples completed successfully!")
}