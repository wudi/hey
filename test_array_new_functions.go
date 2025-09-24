package main

import (
	"fmt"
	"os"

	"github.com/wudi/hey/runtime"
	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// Simple test context for builtin functions
type testContext struct{}

func (t *testContext) WriteOutput(val *values.Value) error { return nil }
func (t *testContext) GetGlobal(name string) (*values.Value, bool) { return nil, false }
func (t *testContext) SetGlobal(name string, val *values.Value) {}
func (t *testContext) SymbolRegistry() *registry.Registry {
	reg := registry.NewRegistry()
	// Register functions
	functions := runtime.GetArrayFunctions()
	for _, fn := range functions {
		reg.RegisterFunction(fn)
	}
	return reg
}
func (t *testContext) LookupUserFunction(name string) (*registry.Function, bool) { return nil, false }
func (t *testContext) CallUserFunction(function *registry.Function, args []*values.Value) (*values.Value, error) { return nil, fmt.Errorf("user function calls not supported in test") }
func (t *testContext) SimpleCallUserFunction(function *registry.Function, args []*values.Value) (*values.Value, error) { return nil, fmt.Errorf("user function calls not supported in test") }

func main() {
	fmt.Println("Testing new array functions...")
	ctx := &testContext{}

	// Test array_key_exists
	fmt.Println("\n=== Testing array_key_exists ===")
	arr := values.NewArray()
	arrData := arr.Data.(*values.Array)
	arrData.Elements["a"] = values.NewInt(1)
	arrData.Elements[int64(2)] = values.NewString("two")

	functions := runtime.GetArrayFunctions()
	var keyExistsFn *registry.Function
	for _, fn := range functions {
		if fn.Name == "array_key_exists" {
			keyExistsFn = fn
			break
		}
	}

	if keyExistsFn == nil {
		fmt.Println("array_key_exists function not found!")
		os.Exit(1)
	}

	// Test with string key
	result1, err := keyExistsFn.Builtin(*ctx, []*values.Value{values.NewString("a"), arr})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("array_key_exists('a', arr): %v (should be true)\n", result1.ToBool())
	}

	// Test with int key
	result2, err := keyExistsFn.Builtin(*ctx, []*values.Value{values.NewInt(2), arr})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("array_key_exists(2, arr): %v (should be true)\n", result2.ToBool())
	}

	// Test with missing key
	result3, err := keyExistsFn.Builtin(*ctx, []*values.Value{values.NewString("missing"), arr})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("array_key_exists('missing', arr): %v (should be false)\n", result3.ToBool())
	}

	// Test sorting functions
	fmt.Println("\n=== Testing sorting functions ===")
	testArr := values.NewArray()
	testArrData := testArr.Data.(*values.Array)
	testArrData.Elements[int64(0)] = values.NewInt(3)
	testArrData.Elements[int64(1)] = values.NewInt(1)
	testArrData.Elements[int64(2)] = values.NewInt(4)
	testArrData.NextIndex = 3

	var sortFn *registry.Function
	for _, fn := range functions {
		if fn.Name == "sort" {
			sortFn = fn
			break
		}
	}

	if sortFn != nil {
		result, err := sortFn.Builtin(*ctx, []*values.Value{testArr})
		if err != nil {
			fmt.Printf("Sort error: %v\n", err)
		} else {
			fmt.Printf("sort() returned: %v\n", result.ToBool())
			fmt.Printf("Array after sort: ")
			for i := int64(0); i < testArrData.NextIndex; i++ {
				if val, exists := testArrData.Elements[i]; exists {
					fmt.Printf("%d ", val.ToInt())
				}
			}
			fmt.Println()
		}
	}

	fmt.Println("\n=== Testing array_product ===")
	var productFn *registry.Function
	for _, fn := range functions {
		if fn.Name == "array_product" {
			productFn = fn
			break
		}
	}

	if productFn != nil {
		// Test with numbers
		prodArr := values.NewArray()
		prodArrData := prodArr.Data.(*values.Array)
		prodArrData.Elements[int64(0)] = values.NewInt(2)
		prodArrData.Elements[int64(1)] = values.NewInt(3)
		prodArrData.Elements[int64(2)] = values.NewInt(4)

		result, err := productFn.Builtin(*ctx, []*values.Value{prodArr})
		if err != nil {
			fmt.Printf("Product error: %v\n", err)
		} else {
			fmt.Printf("array_product([2,3,4]): %v (should be 24)\n", result.ToInt())
		}

		// Test with empty array
		emptyArr := values.NewArray()
		result2, err := productFn.Builtin(*ctx, []*values.Value{emptyArr})
		if err != nil {
			fmt.Printf("Product error: %v\n", err)
		} else {
			fmt.Printf("array_product([]): %v (should be 0 in our implementation, 1 in PHP)\n", result2.ToInt())
		}
	}

	fmt.Println("\nAll basic tests completed!")
}