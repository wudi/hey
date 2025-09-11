package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/wudi/hey/compiler"
	"github.com/wudi/hey/compiler/lexer"
	"github.com/wudi/hey/compiler/parser"
	"github.com/wudi/hey/compiler/runtime"
	"github.com/wudi/hey/compiler/vm"
)

func main() {
	fmt.Println("=== Enhanced PHP VM Demonstration ===")

	// Initialize runtime
	err := runtime.Bootstrap()
	if err != nil {
		fmt.Printf("Runtime bootstrap failed: %v\n", err)
		os.Exit(1)
	}

	err = runtime.InitializeVMIntegration()
	if err != nil {
		fmt.Printf("VM integration failed: %v\n", err)
		os.Exit(1)
	}

	// Demonstrate different VM configurations
	demonstrateBasicVM()
	demonstrateProfilingVM()
	demonstrateAdvancedFeatures()
}

func demonstrateBasicVM() {
	fmt.Println("\n--- Basic VM Performance ---")

	phpCode := `<?php
// Basic arithmetic and output
$x = 10;
$y = 20;
$result = $x + $y;
echo "Basic calculation: $result\n";

// Loop performance
for ($i = 0; $i < 5; $i++) {
    echo "Iteration $i\n";
}
?>`

	// Basic VM without profiling
	basicVM := vm.NewVirtualMachine()
	executeCode(basicVM, phpCode, "Basic VM")
}

func demonstrateProfilingVM() {
	fmt.Println("\n--- VM with Performance Profiling ---")

	phpCode := `<?php
// Arrow functions (modern PHP feature)
$multiply = fn($a, $b) => $a * $b;
echo "Arrow function result: " . $multiply(6, 7) . "\n";

// Spread expressions
$arr1 = [1, 2, 3];
$arr2 = [4, 5];
$combined = [...$arr1, ...$arr2, 6];
echo "Spread array count: " . count($combined) . "\n";

// Advanced control flow
goto skip_this;
echo "This should be skipped\n";
skip_this:
echo "Jumped successfully\n";

// Declare statement
declare(strict_types=1);
echo "Strict types enabled\n";
?>`

	// VM with detailed profiling
	profilingVM := vm.NewVirtualMachineWithProfiling(vm.DebugLevelBasic)
	profilingVM.SetBreakpoint(50) // Set a breakpoint
	profilingVM.WatchVariable("$multiply")

	executeCode(profilingVM, phpCode, "Profiling VM")

	// Show performance report
	fmt.Println("\n--- Performance Report ---")
	fmt.Println(profilingVM.GetPerformanceReport())

	// Show debug report
	fmt.Println("\n--- Debug Report ---")
	fmt.Println(profilingVM.GetDebugReport())

	// Show hot spots
	fmt.Println("\n--- Hot Spots Analysis ---")
	hotSpots := profilingVM.GetHotSpots(5)
	for i, spot := range hotSpots {
		fmt.Printf("#%d: IP %d executed %d times\n", i+1, spot.IP, spot.Count)
	}

	// Show memory statistics
	fmt.Println("\n--- Memory Pool Statistics ---")
	allocs, deallocs := profilingVM.GetMemoryStats()
	fmt.Printf("Allocations: %d, Deallocations: %d\n", allocs, deallocs)
}

func demonstrateAdvancedFeatures() {
	fmt.Println("\n--- Advanced VM Features ---")

	phpCode := `<?php
// Complex modern PHP with all features
class Calculator {
    private $operations = [];
    
    public function add($a, $b) {
        $this->operations[] = "add";
        return $a + $b;
    }
}

$calc = new Calculator();
$numbers = [10, 20, 30];

// Arrow function with spread and object method call
$processNumbers = fn($nums) => $calc->add(...$nums[0:2]);
$result = $processNumbers($numbers);

echo "Advanced calculation result: $result\n";

// Alternative syntax with goto
if ($result > 25):
    goto success;
else:
    echo "Result too small\n";
endif;

success:
echo "Calculation successful!\n";
?>`

	// Advanced VM with all features enabled
	advancedVM := vm.NewVirtualMachineWithProfiling(vm.DebugLevelDetailed)
	advancedVM.EnableAdvancedProfiling()

	executeCode(advancedVM, phpCode, "Advanced VM")

	fmt.Println("\n--- Advanced Analysis ---")
	fmt.Println(advancedVM.GetPerformanceReport())
}

func executeCode(vmachine *vm.VirtualMachine, phpCode, description string) {
	fmt.Printf("\nExecuting %s:\n", description)
	fmt.Println(strings.Repeat("-", 40))

	// Parse
	l := lexer.New(phpCode)
	p := parser.New(l)
	program := p.ParseProgram()

	if program == nil {
		fmt.Printf("❌ Parse failed for %s\n", description)
		return
	}

	// Compile
	comp := compiler.NewCompiler()
	err := comp.Compile(program)
	if err != nil {
		fmt.Printf("❌ Compilation failed for %s: %v\n", description, err)
		return
	}

	// Execute
	ctx := vm.NewExecutionContext()
	var output strings.Builder
	ctx.SetOutputWriter(&output)

	err = vmachine.Execute(ctx, comp.GetBytecode(), comp.GetConstants(), comp.GetFunctions(), comp.GetClasses())
	if err != nil {
		fmt.Printf("❌ Execution failed for %s: %v\n", description, err)
		return
	}

	fmt.Print(output.String())
	fmt.Printf("✅ %s completed successfully\n", description)
}
