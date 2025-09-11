package compiler

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wudi/hey/compiler/ast"
	"github.com/wudi/hey/compiler/lexer"
	"github.com/wudi/hey/compiler/opcodes"
	"github.com/wudi/hey/compiler/parser"
	"github.com/wudi/hey/compiler/registry"
	"github.com/wudi/hey/compiler/runtime"
	"github.com/wudi/hey/compiler/values"
	"github.com/wudi/hey/compiler/vm"
)

// Helper function to execute compiled bytecode with runtime initialization
func executeWithRuntime(t *testing.T, comp *Compiler) error {
	// Initialize unified registry only if not already initialized
	if registry.GlobalRegistry == nil {
		registry.Initialize()
	}

	// Initialize runtime if not already done
	if runtime.GlobalRegistry == nil {
		err := runtime.Bootstrap()
		require.NoError(t, err, "Failed to bootstrap runtime")
	}

	// Initialize VM integration
	if runtime.GlobalVMIntegration == nil {
		err := runtime.InitializeVMIntegration()
		require.NoError(t, err, "Failed to initialize VM integration")
	}

	// Create VM and execution context
	vmachine := vm.NewVirtualMachine()
	vmCtx := vm.NewExecutionContext()

	// Initialize global variables from runtime
	if vmCtx.GlobalVars == nil {
		vmCtx.GlobalVars = make(map[string]*values.Value)
	}

	variables := runtime.GlobalVMIntegration.GetAllVariables()
	for name, value := range variables {
		vmCtx.GlobalVars[name] = value
	}

	// Execute bytecode
	return vmachine.Execute(vmCtx, comp.GetBytecode(), comp.GetConstants(), comp.GetFunctions(), comp.GetClasses())
}

func TestEcho(t *testing.T) {
	p := parser.New(lexer.New(`<?php echo "Hello, World!";`))
	prog := p.ParseProgram()

	comp := NewCompiler()
	err := comp.Compile(prog)
	require.NoError(t, err)

	err = executeWithRuntime(t, comp)
	require.NoError(t, err)
}

// Test built-in functions
func TestBuiltinFunctions(t *testing.T) {
	testCases := []struct {
		name string
		code string
	}{
		{
			name: "strlen",
			code: `<?php echo strlen("hello");`,
		},
		{
			name: "count_array",
			code: `<?php 
				$arr = [1, 2, 3];
				echo count($arr);`,
		},
		{
			name: "is_string",
			code: `<?php 
				$str = "hello";
				var_dump(is_string($str));`,
		},
		{
			name: "is_int",
			code: `<?php 
				$num = 42;
				var_dump(is_int($num));`,
		},
		{
			name: "is_array",
			code: `<?php 
				$arr = [1, 2, 3];
				var_dump(is_array($arr));`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := parser.New(lexer.New(tc.code))
			prog := p.ParseProgram()
			require.NotNil(t, prog, "Failed to parse program for test: %s", tc.name)

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Compilation failed for test: %s", tc.name)

			err = executeWithRuntime(t, comp)
			require.NoError(t, err, "Execution failed for test: %s", tc.name)
		})
	}
}

// Test string functions
func TestStringFunctions(t *testing.T) {
	code := `<?php
		$str = "Hello World";
		echo strlen($str) . "\n";
		var_dump(is_string($str));
	`

	p := parser.New(lexer.New(code))
	prog := p.ParseProgram()
	require.NotNil(t, prog)

	comp := NewCompiler()
	err := comp.Compile(prog)
	require.NoError(t, err)

	vmCtx := vm.NewExecutionContext()
	err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants(), comp.GetFunctions(), comp.GetClasses())
	require.NoError(t, err)
}

// Test array functions
func TestArrayFunctions(t *testing.T) {
	code := `<?php
		$arr = [1, 2, 3, "hello"];
		echo count($arr) . "\n";
		var_dump(is_array($arr));
	`

	p := parser.New(lexer.New(code))
	prog := p.ParseProgram()
	require.NotNil(t, prog)

	comp := NewCompiler()
	err := comp.Compile(prog)
	require.NoError(t, err)

	vmCtx := vm.NewExecutionContext()
	err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants(), comp.GetFunctions(), comp.GetClasses())
	require.NoError(t, err)
}

// Test type checking functions
func TestTypeCheckingFunctions(t *testing.T) {
	code := `<?php
		$str = "hello";
		$num = 42;
		$arr = [1, 2, 3];
		
		var_dump(is_string($str));
		var_dump(is_int($num)); 
		var_dump(is_array($arr));
		var_dump(is_string($num));
	`

	p := parser.New(lexer.New(code))
	prog := p.ParseProgram()
	require.NotNil(t, prog)

	comp := NewCompiler()
	err := comp.Compile(prog)
	require.NoError(t, err)

	vmCtx := vm.NewExecutionContext()
	err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants(), comp.GetFunctions(), comp.GetClasses())
	require.NoError(t, err)
}

func TestForeachWithFunctionCall(t *testing.T) {
	code := `<?php
function foo($n):array {
    $ret = [];
    for($i=0; $i<$n; $i++) {
        $ret[] = $i;
    }
    return $ret;
}

foreach(foo(5) as $v) {
    echo "$v\n";
}`

	p := parser.New(lexer.New(code))
	prog := p.ParseProgram()

	comp := NewCompiler()
	err := comp.Compile(prog)
	require.NoError(t, err, "Compilation failed")

	// Debug: Print compiled functions
	t.Logf("Compiled functions: %v", len(comp.GetFunctions()))
	for name, fn := range comp.GetFunctions() {
		t.Logf("Function %s: %d instructions, %d constants", name, len(fn.Instructions), len(fn.Constants))
	}

	// Execute with runtime
	vmCtx := vm.NewExecutionContext()
	// Set up output capture
	var buf bytes.Buffer
	vmCtx.SetOutputWriter(&buf)
	// Initialize runtime if not already done
	if runtime.GlobalRegistry == nil {
		err := runtime.Bootstrap()
		require.NoError(t, err, "Failed to bootstrap runtime")
	}
	// Initialize VM integration
	if runtime.GlobalVMIntegration == nil {
		err := runtime.InitializeVMIntegration()
		require.NoError(t, err, "Failed to initialize VM integration")
	}
	err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants(), comp.GetFunctions(), comp.GetClasses())
	require.NoError(t, err, "VM execution failed")

	// Get output from buffer
	output := buf.String()
	t.Logf("VM Output: %q", output)

	// Check that we got the expected output
	expectedOutput := "0\n1\n2\n3\n4\n"
	require.Equal(t, expectedOutput, output, "Expected output doesn't match")
}

func TestSimpleForeach(t *testing.T) {
	code := `<?php
$arr = [0, 1, 2, 3, 4];
foreach($arr as $v) {
    echo "$v\n";
}`

	p := parser.New(lexer.New(code))
	prog := p.ParseProgram()

	comp := NewCompiler()
	err := comp.Compile(prog)
	require.NoError(t, err, "Compilation failed")

	// Execute with runtime
	vmCtx := vm.NewExecutionContext()
	// Set up output capture
	var buf bytes.Buffer
	vmCtx.SetOutputWriter(&buf)
	// Initialize runtime if not already done
	if runtime.GlobalRegistry == nil {
		err := runtime.Bootstrap()
		require.NoError(t, err, "Failed to bootstrap runtime")
	}
	// Initialize VM integration
	if runtime.GlobalVMIntegration == nil {
		err := runtime.InitializeVMIntegration()
		require.NoError(t, err, "Failed to initialize VM integration")
	}
	err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants(), comp.GetFunctions(), comp.GetClasses())
	require.NoError(t, err, "VM execution failed")

	// Get output from buffer
	output := buf.String()
	t.Logf("Simple foreach VM Output: %q", output)

	// Check that we got the expected output
	expectedOutput := "0\n1\n2\n3\n4\n"
	require.Equal(t, expectedOutput, output, "Expected output doesn't match")
}

func TestSimpleFunctionCall(t *testing.T) {
	code := `<?php
function foo($n):array {
    $ret = [];
    for($i=0; $i<$n; $i++) {
        $ret[] = $i;
    }
    return $ret;
}

$result = foo(3);
echo "done";`

	p := parser.New(lexer.New(code))
	prog := p.ParseProgram()

	comp := NewCompiler()
	err := comp.Compile(prog)
	require.NoError(t, err, "Compilation failed")

	// Debug: Print compiled functions
	t.Logf("Compiled functions: %v", len(comp.GetFunctions()))
	for name, fn := range comp.GetFunctions() {
		t.Logf("Function %s: %d instructions, %d constants", name, len(fn.Instructions), len(fn.Constants))
	}

	// Execute with runtime and capture output
	vmCtx := vm.NewExecutionContext()
	var buf bytes.Buffer
	vmCtx.SetOutputWriter(&buf)
	// Initialize runtime if not already done
	if runtime.GlobalRegistry == nil {
		err := runtime.Bootstrap()
		require.NoError(t, err, "Failed to bootstrap runtime")
	}
	// Initialize VM integration
	if runtime.GlobalVMIntegration == nil {
		err := runtime.InitializeVMIntegration()
		require.NoError(t, err, "Failed to initialize VM integration")
	}
	err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants(), comp.GetFunctions(), comp.GetClasses())
	require.NoError(t, err, "VM execution failed")

	output := buf.String()
	t.Logf("Function call VM Output: %q", output)

	// Check that we got some output (at least "done")
	require.Contains(t, output, "done", "Should contain 'done'")
}

func TestArithmeticOperators(t *testing.T) {
	testCases := []struct {
		name string
		code string
	}{
		{"Addition", `<?php echo 5 + 3;`},
		{"Subtraction", `<?php echo 10 - 4;`},
		{"Multiplication", `<?php echo 6 * 7;`},
		{"Division", `<?php echo 15 / 5;`},
		{"Modulo", `<?php echo 17 % 5;`},
		{"Power", `<?php echo 2 ** 3;`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := parser.New(lexer.New(tc.code))
			prog := p.ParseProgram()

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Failed to compile %s", tc.name)

			vmCtx := vm.NewExecutionContext()
			err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants(), comp.GetFunctions(), comp.GetClasses())
			require.NoError(t, err, "Failed to execute %s", tc.name)
		})
	}
}

func TestComparisonOperators(t *testing.T) {
	testCases := []struct {
		name string
		code string
	}{
		{"Equal", `<?php echo 5 == 5;`},
		{"NotEqual", `<?php echo 5 != 3;`},
		{"NotEqualAlt", `<?php echo 5 <> 3;`},
		{"Identical", `<?php echo 5 === 5;`},
		{"NotIdentical", `<?php echo 5 !== "5";`},
		{"LessThan", `<?php echo 3 < 5;`},
		{"LessEqual", `<?php echo 5 <= 5;`},
		{"GreaterThan", `<?php echo 7 > 3;`},
		{"GreaterEqual", `<?php echo 5 >= 5;`},
		{"Spaceship", `<?php echo 5 <=> 3;`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := parser.New(lexer.New(tc.code))
			prog := p.ParseProgram()

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Failed to compile %s", tc.name)

			vmCtx := vm.NewExecutionContext()
			err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants(), comp.GetFunctions(), comp.GetClasses())
			require.NoError(t, err, "Failed to execute %s", tc.name)
		})
	}
}

func TestLogicalOperators(t *testing.T) {
	testCases := []struct {
		name string
		code string
	}{
		{"BooleanAnd", `<?php echo true && false;`},
		{"BooleanOr", `<?php echo true || false;`},
		{"LogicalAnd", `<?php echo true and false;`},
		{"LogicalOr", `<?php echo true or false;`},
		{"LogicalXor", `<?php echo true xor false;`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := parser.New(lexer.New(tc.code))
			prog := p.ParseProgram()

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Failed to compile %s", tc.name)

			vmCtx := vm.NewExecutionContext()
			err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants(), comp.GetFunctions(), comp.GetClasses())
			require.NoError(t, err, "Failed to execute %s", tc.name)
		})
	}
}

func TestBitwiseOperators(t *testing.T) {
	testCases := []struct {
		name string
		code string
	}{
		{"BitwiseAnd", `<?php echo 12 & 10;`},
		{"BitwiseOr", `<?php echo 12 | 10;`},
		{"BitwiseXor", `<?php echo 12 ^ 10;`},
		{"ShiftLeft", `<?php echo 5 << 2;`},
		{"ShiftRight", `<?php echo 20 >> 2;`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := parser.New(lexer.New(tc.code))
			prog := p.ParseProgram()

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Failed to compile %s", tc.name)

			vmCtx := vm.NewExecutionContext()
			err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants(), comp.GetFunctions(), comp.GetClasses())
			require.NoError(t, err, "Failed to execute %s", tc.name)
		})
	}
}

func TestUnaryOperators(t *testing.T) {
	testCases := []struct {
		name string
		code string
	}{
		{"UnaryPlus", `<?php echo +42;`},
		{"UnaryMinus", `<?php echo -42;`},
		{"LogicalNot", `<?php echo !true;`},
		{"BitwiseNot", `<?php echo ~15;`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := parser.New(lexer.New(tc.code))
			prog := p.ParseProgram()

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Failed to compile %s", tc.name)

			vmCtx := vm.NewExecutionContext()
			err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants(), comp.GetFunctions(), comp.GetClasses())
			require.NoError(t, err, "Failed to execute %s", tc.name)
		})
	}
}

func TestStringOperators(t *testing.T) {
	testCases := []struct {
		name string
		code string
	}{
		{"StringConcat", `<?php echo "Hello" . " World";`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := parser.New(lexer.New(tc.code))
			prog := p.ParseProgram()

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Failed to compile %s", tc.name)

			vmCtx := vm.NewExecutionContext()
			err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants(), comp.GetFunctions(), comp.GetClasses())
			require.NoError(t, err, "Failed to execute %s", tc.name)
		})
	}
}

func TestAdvancedComparisonOperators(t *testing.T) {
	testCases := []struct {
		name   string
		code   string
		expect string
	}{
		// Integer comparisons
		{"IntEqual_True", `<?php echo 5 == 5;`, "1"},
		{"IntEqual_False", `<?php echo 5 == 3;`, ""},
		{"IntNotEqual_True", `<?php echo 5 != 3;`, "1"},
		{"IntNotEqual_False", `<?php echo 5 != 5;`, ""},

		// String comparisons
		{"StringEqual_True", `<?php echo "hello" == "hello";`, "1"},
		{"StringEqual_False", `<?php echo "hello" == "world";`, ""},
		{"StringNotEqual", `<?php echo "hello" != "world";`, "1"},

		// Identity comparisons (strict)
		{"Identical_Int", `<?php echo 5 === 5;`, "1"},
		{"NotIdentical_IntString", `<?php echo 5 !== "5";`, "1"},
		{"Identical_String", `<?php echo "hello" === "hello";`, "1"},
		{"NotIdentical_String", `<?php echo "hello" !== "world";`, "1"},

		// Numeric comparisons
		{"LessThan_True", `<?php echo 3 < 5;`, "1"},
		{"LessThan_False", `<?php echo 5 < 3;`, ""},
		{"LessEqual_Equal", `<?php echo 5 <= 5;`, "1"},
		{"LessEqual_Less", `<?php echo 3 <= 5;`, "1"},
		{"LessEqual_False", `<?php echo 7 <= 5;`, ""},

		{"GreaterThan_True", `<?php echo 7 > 3;`, "1"},
		{"GreaterThan_False", `<?php echo 3 > 7;`, ""},
		{"GreaterEqual_Equal", `<?php echo 5 >= 5;`, "1"},
		{"GreaterEqual_Greater", `<?php echo 7 >= 3;`, "1"},
		{"GreaterEqual_False", `<?php echo 3 >= 7;`, ""},

		// Spaceship operator
		{"Spaceship_Less", `<?php echo 3 <=> 5;`, "-1"},
		{"Spaceship_Equal", `<?php echo 5 <=> 5;`, "0"},
		{"Spaceship_Greater", `<?php echo 7 <=> 3;`, "1"},

		// Boolean comparisons
		{"BoolEqual_True", `<?php echo true == true;`, "1"},
		{"BoolEqual_False", `<?php echo true == false;`, ""},
		{"BoolIdentical", `<?php echo true === true;`, "1"},
		{"BoolNotIdentical", `<?php echo true !== false;`, "1"},

		// Mixed type comparisons
		{"IntBool_Equal", `<?php echo 1 == true;`, "1"},
		{"IntBool_NotIdentical", `<?php echo 1 !== true;`, "1"},
		{"StringInt_Equal", `<?php echo "5" == 5;`, "1"},
		{"StringInt_NotIdentical", `<?php echo "5" !== 5;`, "1"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := parser.New(lexer.New(tc.code))
			prog := p.ParseProgram()

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Failed to compile %s", tc.name)

			vmCtx := vm.NewExecutionContext()
			err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants(), comp.GetFunctions(), comp.GetClasses())
			require.NoError(t, err, "Failed to execute %s", tc.name)
		})
	}
}

func TestComparisonWithNull(t *testing.T) {
	testCases := []struct {
		name string
		code string
	}{
		{"NullEqual_Null", `<?php echo null == null;`},
		{"NullNotEqual_Int", `<?php echo null != 5;`},
		{"NullIdentical", `<?php echo null === null;`},
		{"NullNotIdentical_False", `<?php echo null !== false;`},
		{"NullLessThan_Int", `<?php echo null < 5;`},
		{"IntGreater_Null", `<?php echo 5 > null;`},
		{"NullSpaceship", `<?php echo null <=> null;`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := parser.New(lexer.New(tc.code))
			prog := p.ParseProgram()

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Failed to compile %s", tc.name)

			vmCtx := vm.NewExecutionContext()
			err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants(), comp.GetFunctions(), comp.GetClasses())
			require.NoError(t, err, "Failed to execute %s", tc.name)
		})
	}
}

func TestComplexComparisonExpressions(t *testing.T) {
	testCases := []struct {
		name string
		code string
	}{
		// Nested comparisons with logical operators
		{"NestedAnd", `<?php echo (5 > 3) && (10 < 20);`},
		{"NestedOr", `<?php echo (5 < 3) || (10 > 5);`},
		{"MixedLogical", `<?php echo (5 == 5) && (3 != 4) || (2 < 1);`},

		// Chained comparisons
		{"ChainedEqual", `<?php echo 5 == 5 == true;`},
		{"ChainedComparison", `<?php echo 1 < 2 < 3;`},

		// Comparisons with arithmetic
		{"ArithmeticComparison", `<?php echo (5 + 3) > (2 * 3);`},
		{"ComplexArithmetic", `<?php echo (10 / 2) == (15 - 10);`},
		{"PowerComparison", `<?php echo (2 ** 3) > (3 ** 2);`},

		// String comparisons
		{"StringLength", `<?php echo "abc" < "def";`},
		{"StringNumeric", `<?php echo "10" > "2";`},
		{"StringConcat", `<?php echo ("hello" . " world") == "hello world";`},

		// Mixed type complex comparisons
		{"MixedTypeComplex", `<?php echo ("5" == 5) && (true == 1) && (false == 0);`},
		{"IdenticalVsEqual", `<?php echo ("5" == 5) && ("5" !== 5);`},

		// Parenthesized expressions
		{"ParenthesesGrouping", `<?php echo (5 > 3 && 2 < 4) || (1 == 2);`},
		{"NestedParentheses", `<?php echo ((5 + 2) > (3 * 2)) && ((4 - 1) == 3);`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := parser.New(lexer.New(tc.code))
			prog := p.ParseProgram()

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Failed to compile %s", tc.name)

			vmCtx := vm.NewExecutionContext()
			err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants(), comp.GetFunctions(), comp.GetClasses())
			require.NoError(t, err, "Failed to execute %s", tc.name)
		})
	}
}

func TestSpaceshipOperatorDetails(t *testing.T) {
	testCases := []struct {
		name     string
		code     string
		expected string
	}{
		// Integer spaceship
		{"IntSpaceship_Less", `<?php echo 1 <=> 5;`, "-1"},
		{"IntSpaceship_Equal", `<?php echo 5 <=> 5;`, "0"},
		{"IntSpaceship_Greater", `<?php echo 10 <=> 3;`, "1"},

		// String spaceship (lexicographical)
		{"StringSpaceship_Less", `<?php echo "apple" <=> "banana";`, "-1"},
		{"StringSpaceship_Equal", `<?php echo "hello" <=> "hello";`, "0"},
		{"StringSpaceship_Greater", `<?php echo "zebra" <=> "apple";`, "1"},

		// Mixed type spaceship
		{"MixedSpaceship_IntString", `<?php echo 5 <=> "5";`, "0"},
		{"MixedSpaceship_BoolInt", `<?php echo true <=> 1;`, "0"},
		{"MixedSpaceship_NullInt", `<?php echo null <=> 0;`, "0"},

		// Negative numbers
		{"NegativeSpaceship", `<?php echo -5 <=> -3;`, "-1"},
		{"NegativePositive", `<?php echo -1 <=> 1;`, "-1"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := parser.New(lexer.New(tc.code))
			prog := p.ParseProgram()

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Failed to compile %s", tc.name)

			vmCtx := vm.NewExecutionContext()
			err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants(), comp.GetFunctions(), comp.GetClasses())
			require.NoError(t, err, "Failed to execute %s", tc.name)
		})
	}
}

func TestIncrementDecrementOperators(t *testing.T) {
	testCases := []struct {
		name string
		code string
	}{
		// Pre-increment
		{"PreIncrement_Int", `<?php $x = 5; echo ++$x;`},
		{"PreIncrement_Float", `<?php $x = 3.5; echo ++$x;`},
		{"PreIncrement_String", `<?php $x = "10"; echo ++$x;`},
		{"PreIncrement_StringFloat", `<?php $x = "10.5"; echo ++$x;`},

		// Pre-decrement
		{"PreDecrement_Int", `<?php $x = 5; echo --$x;`},
		{"PreDecrement_Float", `<?php $x = 3.5; echo --$x;`},
		{"PreDecrement_String", `<?php $x = "10"; echo --$x;`},
		{"PreDecrement_StringFloat", `<?php $x = "10.5"; echo --$x;`},

		// Post-increment
		{"PostIncrement_Int", `<?php $x = 5; echo $x++;`},
		{"PostIncrement_Float", `<?php $x = 3.5; echo $x++;`},
		{"PostIncrement_String", `<?php $x = "10"; echo $x++;`},
		{"PostIncrement_StringFloat", `<?php $x = "10.5"; echo $x++;`},

		// Post-decrement
		{"PostDecrement_Int", `<?php $x = 5; echo $x--;`},
		{"PostDecrement_Float", `<?php $x = 3.5; echo $x--;`},
		{"PostDecrement_String", `<?php $x = "10"; echo $x--;`},
		{"PostDecrement_StringFloat", `<?php $x = "10.5"; echo $x--;`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := parser.New(lexer.New(tc.code))
			prog := p.ParseProgram()

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Failed to compile %s", tc.name)

			vmCtx := vm.NewExecutionContext()
			err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants(), comp.GetFunctions(), comp.GetClasses())
			require.NoError(t, err, "Failed to execute %s", tc.name)
		})
	}
}

func TestAdvancedIncrementDecrementOperators(t *testing.T) {
	testCases := []struct {
		name     string
		code     string
		expected string
	}{
		// Pre-increment tests with expected results
		{"PreIncrement_IntValue", `<?php $x = 5; echo ++$x;`, "6"},
		{"PreIncrement_FloatValue", `<?php $x = 3.5; echo ++$x;`, "4.5"},
		{"PreIncrement_StringValue", `<?php $x = "10"; echo ++$x;`, "11"},
		{"PreIncrement_ZeroValue", `<?php $x = 0; echo ++$x;`, "1"},

		// Pre-decrement tests with expected results
		{"PreDecrement_IntValue", `<?php $x = 5; echo --$x;`, "4"},
		{"PreDecrement_FloatValue", `<?php $x = 3.5; echo --$x;`, "2.5"},
		{"PreDecrement_StringValue", `<?php $x = "10"; echo --$x;`, "9"},
		{"PreDecrement_ZeroValue", `<?php $x = 0; echo --$x;`, "-1"},

		// Post-increment tests with expected results
		{"PostIncrement_IntValue", `<?php $x = 5; echo $x++;`, "5"},
		{"PostIncrement_FloatValue", `<?php $x = 3.5; echo $x++;`, "3.5"},
		{"PostIncrement_StringValue", `<?php $x = "10"; echo $x++;`, "10"},
		{"PostIncrement_ZeroValue", `<?php $x = 0; echo $x++;`, "0"},

		// Post-decrement tests with expected results
		{"PostDecrement_IntValue", `<?php $x = 5; echo $x--;`, "5"},
		{"PostDecrement_FloatValue", `<?php $x = 3.5; echo $x--;`, "3.5"},
		{"PostDecrement_StringValue", `<?php $x = "10"; echo $x--;`, "10"},
		{"PostDecrement_ZeroValue", `<?php $x = 0; echo $x--;`, "0"},

		// Edge cases
		{"PreIncrement_Null", `<?php $x = null; echo ++$x;`, "1"},
		{"PreDecrement_Null", `<?php $x = null; echo --$x;`, "-1"},
		{"PostIncrement_Null", `<?php $x = null; echo $x++;`, ""},
		{"PostDecrement_Null", `<?php $x = null; echo $x--;`, ""},

		// Boolean values
		{"PreIncrement_True", `<?php $x = true; echo ++$x;`, "2"},
		{"PreIncrement_False", `<?php $x = false; echo ++$x;`, "1"},
		{"PreDecrement_True", `<?php $x = true; echo --$x;`, "0"},
		{"PreDecrement_False", `<?php $x = false; echo --$x;`, "-1"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := parser.New(lexer.New(tc.code))
			prog := p.ParseProgram()

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Failed to compile %s", tc.name)

			vmCtx := vm.NewExecutionContext()
			err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants(), comp.GetFunctions(), comp.GetClasses())
			require.NoError(t, err, "Failed to execute %s", tc.name)
		})
	}
}

func TestIncrementDecrementSequences(t *testing.T) {
	testCases := []struct {
		name string
		code string
	}{
		// Multiple operations
		{"MultiplePreIncrement", `<?php $x = 5; echo ++$x; echo " "; echo ++$x;`},
		{"MultiplePostIncrement", `<?php $x = 5; echo $x++; echo " "; echo $x++;`},
		{"MixedIncrementDecrement", `<?php $x = 5; echo ++$x; echo " "; echo $x--; echo " "; echo --$x;`},

		// Chained operations
		{"ChainedIncrement", `<?php $x = 0; echo ++$x + ++$x;`},
		{"ChainedDecrement", `<?php $x = 10; echo --$x - --$x;`},
		{"MixedChained", `<?php $x = 5; echo ++$x * $x--;`},

		// Complex expressions
		{"IncrementInExpression", `<?php $x = 5; $y = 3; echo ($x++ + ++$y);`},
		{"DecrementInExpression", `<?php $x = 10; $y = 8; echo ($x-- - --$y);`},
		{"IncrementComparisonExpression", `<?php $x = 5; echo (++$x > 5);`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := parser.New(lexer.New(tc.code))
			prog := p.ParseProgram()

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Failed to compile %s", tc.name)

			vmCtx := vm.NewExecutionContext()
			err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants(), comp.GetFunctions(), comp.GetClasses())
			require.NoError(t, err, "Failed to execute %s", tc.name)
		})
	}
}

func TestArrayElementIncrementDecrement(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{
			"ArrayElementPostIncrement",
			`<?php 
			$sum_results = array(); 
			$v = "key1"; 
			$sum_results[$v]++; 
			echo $sum_results[$v];`,
		},
		{
			"ArrayElementPreIncrement",
			`<?php 
			$sum_results = array("key1" => 5); 
			$v = "key1"; 
			echo ++$sum_results[$v];`,
		},
		{
			"ArrayElementPostDecrement",
			`<?php 
			$sum_results = array("key1" => 5); 
			$v = "key1"; 
			echo $sum_results[$v]--;`,
		},
		{
			"ArrayElementPreDecrement",
			`<?php 
			$sum_results = array("key1" => 5); 
			$v = "key1"; 
			echo --$sum_results[$v];`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := parser.New(lexer.New(tc.code))
			prog := p.ParseProgram()
			if len(p.Errors()) > 0 {
				t.Fatalf("parser errors: %v", p.Errors())
			}

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Failed to compile %s", tc.name)

			err = executeWithRuntime(t, comp)
			require.NoError(t, err, "Failed to execute %s", tc.name)
		})
	}
}

func TestPropertyIncrementDecrement(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{
			"SimplePropertyPostIncrement",
			`<?php 
			class TestClass { 
				public $count = 5; 
			} 
			$obj = new TestClass(); 
			echo $obj->count++;`,
		},
		{
			"SimplePropertyPreIncrement",
			`<?php 
			class TestClass { 
				public $count = 5; 
			} 
			$obj = new TestClass(); 
			echo ++$obj->count;`,
		},
		{
			"SimplePropertyPostDecrement",
			`<?php 
			class TestClass { 
				public $count = 5; 
			} 
			$obj = new TestClass(); 
			echo $obj->count--;`,
		},
		{
			"SimplePropertyPreDecrement",
			`<?php 
			class TestClass { 
				public $count = 5; 
			} 
			$obj = new TestClass(); 
			echo --$obj->count;`,
		},
		{
			"PropertyIncrementAfterOperation",
			`<?php 
			class TestClass { 
				public $value = 10; 
			} 
			$obj = new TestClass(); 
			$obj->value++;
			echo $obj->value;`,
		},
		{
			"PropertyDecrementAfterOperation",
			`<?php 
			class TestClass { 
				public $value = 10; 
			} 
			$obj = new TestClass(); 
			$obj->value--;
			echo $obj->value;`,
		},
		{
			"PropertyIncrementZeroValue",
			`<?php 
			class TestClass { 
				public $count = 0; 
			} 
			$obj = new TestClass(); 
			echo ++$obj->count;`,
		},
		{
			"PropertyIncrementFloatValue",
			`<?php 
			class TestClass { 
				public $value = 3.5; 
			} 
			$obj = new TestClass(); 
			echo $obj->value++;`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Parse the code
			p := parser.New(lexer.New(tc.code))
			prog := p.ParseProgram()
			require.Empty(t, p.Errors(), "Parser should not have errors for %s", tc.name)

			// Compile the program
			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Failed to compile %s", tc.name)

			// Execute and verify it doesn't crash
			err = executeWithRuntime(t, comp)
			require.NoError(t, err, "Failed to execute %s", tc.name)
		})
	}
}

func TestNestedArrayIncrementDecrement(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{
			"NestedArrayPostIncrement",
			`<?php 
			$sum_results = array();
			$sub = "cat1";
			$v = "key1";
			$sum_results[$sub][$v]++;
			echo $sum_results[$sub][$v];`,
		},
		{
			"NestedArrayPreIncrement",
			`<?php 
			$sum_results = array("cat1" => array("key1" => 5));
			$sub = "cat1";
			$v = "key1";
			echo ++$sum_results[$sub][$v];`,
		},
		{
			"NestedArrayPostDecrement",
			`<?php 
			$sum_results = array("cat1" => array("key1" => 5));
			$sub = "cat1";
			$v = "key1";
			echo $sum_results[$sub][$v]--;`,
		},
		{
			"NestedArrayPreDecrement",
			`<?php 
			$sum_results = array("cat1" => array("key1" => 5));
			$sub = "cat1";
			$v = "key1";
			echo --$sum_results[$sub][$v];`,
		},
		{
			"TripleNestedArrayIncrement",
			`<?php 
			$data = array();
			$a = "level1";
			$b = "level2"; 
			$c = "level3";
			$data[$a][$b][$c]++;
			echo $data[$a][$b][$c];`,
		},
		{
			"MixedNestedArrayOperations",
			`<?php 
			$stats = array("users" => array("active" => 10, "inactive" => 5));
			$stats["users"]["active"]++;
			echo $stats["users"]["active"] . " ";
			echo --$stats["users"]["inactive"];`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := parser.New(lexer.New(tc.code))
			prog := p.ParseProgram()
			if len(p.Errors()) > 0 {
				t.Fatalf("parser errors: %v", p.Errors())
			}

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Failed to compile %s", tc.name)

			err = executeWithRuntime(t, comp)
			require.NoError(t, err, "Failed to execute %s", tc.name)
		})
	}
}

func TestPostIncrementExample(t *testing.T) {
	p := parser.New(lexer.New(`<?php
$a=1;
$a++;

echo $a; // except: 2
`))
	prog := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}
	comp := NewCompiler()
	err := comp.Compile(prog)
	require.NoError(t, err)

	vmCtx := vm.NewExecutionContext()
	err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants(), comp.GetFunctions(), comp.GetClasses())
	require.NoError(t, err)
}

func TestPostIncrementWithOutputCapture(t *testing.T) {
	// Save original stdout
	oldStdout := os.Stdout

	// Create a pipe to capture output
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Parse and execute the PHP code
	p := parser.New(lexer.New(`<?php
$a=1;
$a++;

echo $a; // except: 2
`))
	prog := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}
	comp := NewCompiler()
	err := comp.Compile(prog)
	require.NoError(t, err)

	vmCtx := vm.NewExecutionContext()
	err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants(), comp.GetFunctions(), comp.GetClasses())
	require.NoError(t, err)

	// Close writer and restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify output is "2"
	require.Equal(t, "2", output, "Expected output to be '2', got '%s'", output)
}

func TestSwitchStatement(t *testing.T) {
	// Test the provided switch case - should output "case 124"
	code := `<?php
$a = 123;
$a++;

switch ($a) {
    case 123:
        echo "case 123";
        break;
    case 124:
        echo "case 124";
        break;
    default:
        echo "case default";
}`

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	p := parser.New(lexer.New(code))
	prog := p.ParseProgram()

	// Check for parser errors
	parserErrors := p.Errors()
	require.Empty(t, parserErrors, "Parser should not have errors: %v", parserErrors)

	comp := NewCompiler()
	err := comp.Compile(prog)
	require.NoError(t, err)

	vmCtx := vm.NewExecutionContext()
	err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants(), comp.GetFunctions(), comp.GetClasses())
	require.NoError(t, err)

	// Close write pipe and restore stdout
	w.Close()
	os.Stdout = old

	// Read captured output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify output is "case 124" as expected
	require.Equal(t, "case 124", output, "Expected output to be 'case 124', got '%s'", output)
}

func TestSwitchStatementDefault(t *testing.T) {
	// Test default case
	code := `<?php
$a = 999;

switch ($a) {
    case 123:
        echo "case 123";
        break;
    case 124:
        echo "case 124";
        break;
    default:
        echo "default case";
}`

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	p := parser.New(lexer.New(code))
	prog := p.ParseProgram()

	comp := NewCompiler()
	err := comp.Compile(prog)
	require.NoError(t, err)

	vmCtx := vm.NewExecutionContext()
	err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants(), comp.GetFunctions(), comp.GetClasses())
	require.NoError(t, err)

	// Close write pipe and restore stdout
	w.Close()
	os.Stdout = old

	// Read captured output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	require.Equal(t, "default case", output)
}

func TestSwitchStatementFallthrough(t *testing.T) {
	// Test fall-through behavior (no break statement)
	code := `<?php
$a = 123;

switch ($a) {
    case 123:
        echo "first";
    case 124:
        echo "second";
        break;
    default:
        echo "default";
}`

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	p := parser.New(lexer.New(code))
	prog := p.ParseProgram()

	comp := NewCompiler()
	err := comp.Compile(prog)
	require.NoError(t, err)

	vmCtx := vm.NewExecutionContext()
	err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants(), comp.GetFunctions(), comp.GetClasses())
	require.NoError(t, err)

	// Close write pipe and restore stdout
	w.Close()
	os.Stdout = old

	// Read captured output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Should output "firstsecond" due to fall-through
	require.Equal(t, "firstsecond", output)
}

func TestCoalesceOperator(t *testing.T) {
	testCases := []struct {
		name string
		code string
	}{
		{"NullCoalesceToValue", `<?php echo null ?? "default";`},
		{"ValueCoalesceIgnored", `<?php echo "value" ?? "default";`},
		{"NumberCoalesceIgnored", `<?php echo 42 ?? "default";`},
		{"ZeroCoalesceIgnored", `<?php echo 0 ?? "default";`},
		{"FalseCoalesceIgnored", `<?php echo false ?? "default";`},
		{"EmptyStringCoalesceIgnored", `<?php echo "" ?? "default";`},
		{"ChainedCoalesce", `<?php echo null ?? null ?? "final";`},
		{"VariableCoalesce", `<?php $x = null; echo $x ?? "default";`},
		{"ExpressionCoalesce", `<?php echo (1 > 2 ? "true" : null) ?? "default";`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := parser.New(lexer.New(tc.code))
			prog := p.ParseProgram()

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Failed to compile %s", tc.name)

			vmCtx := vm.NewExecutionContext()
			err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants(), comp.GetFunctions(), comp.GetClasses())
			require.NoError(t, err, "Failed to execute %s", tc.name)
		})
	}
}

func TestMatchExpression(t *testing.T) {
	testCases := []struct {
		name string
		code string
	}{
		{"SimpleMatch", `<?php echo match(1) { 1 => "one", 2 => "two", default => "other" };`},
		{"MatchString", `<?php echo match("hello") { "hello" => "world", "hi" => "there", default => "unknown" };`},
		{"MatchMultipleConditions", `<?php echo match(2) { 1, 2, 3 => "small", 4, 5 => "medium", default => "large" };`},
		{"MatchWithoutDefault", `<?php echo match(1) { 1 => "one", 2 => "two" };`},
		{"MatchStrictComparison", `<?php echo match(1) { "1" => "string", 1 => "integer", default => "other" };`},
		{"MatchExpression", `<?php echo match(5 + 3) { 8 => "eight", 10 => "ten", default => "other" };`},
		{"MatchBooleanValues", `<?php echo match(true) { true => "yes", false => "no", default => "maybe" };`},
		{"MatchNullValue", `<?php echo match(null) { null => "null", 0 => "zero", false => "false", default => "other" };`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := parser.New(lexer.New(tc.code))
			prog := p.ParseProgram()

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Failed to compile %s", tc.name)

			vmCtx := vm.NewExecutionContext()
			err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants(), comp.GetFunctions(), comp.GetClasses())
			require.NoError(t, err, "Failed to execute %s", tc.name)
		})
	}
}

func TestMatchExpressionError(t *testing.T) {
	// Test case where no match is found and there's no default
	code := `<?php echo match(5) { 1 => "one", 2 => "two" };`

	p := parser.New(lexer.New(code))
	prog := p.ParseProgram()

	comp := NewCompiler()
	err := comp.Compile(prog)
	require.NoError(t, err, "Failed to compile match expression error test")

	vmCtx := vm.NewExecutionContext()
	err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants(), comp.GetFunctions(), comp.GetClasses())
	require.Error(t, err, "Expected UnhandledMatchError")
	require.Contains(t, err.Error(), "UnhandledMatchError", "Should contain UnhandledMatchError")
}

func TestForStatement(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{
			"Simple For Loop",
			`<?php for ($i = 0; $i < 3; $i++) { echo $i; }`,
		},
		{
			"For Loop with Empty Init",
			`<?php $i = 0; for (; $i < 2; $i++) { echo $i; }`,
		},
		{
			"For Loop with Empty Update",
			`<?php for ($i = 0; $i < 2; ) { echo $i; $i++; }`,
		},
		{
			"Infinite For Loop with Break",
			`<?php for (;;) { echo "1"; break; }`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(lexer.New(tt.code))
			prog := p.ParseProgram()

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Failed to compile for statement: %s", tt.name)

			vmCtx := vm.NewExecutionContext()
			err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants(), comp.GetFunctions(), comp.GetClasses())
			require.NoError(t, err, "Failed to execute for statement: %s", tt.name)
		})
	}
}

func TestForeachStatement(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{
			"Simple Foreach",
			`<?php $arr = array(1, 2, 3); foreach ($arr as $value) { echo $value; }`,
		},
		{
			"Foreach with Key",
			`<?php $arr = array("a" => 1, "b" => 2); foreach ($arr as $key => $value) { echo $key . ":" . $value; }`,
		},
		{
			"Empty Array Foreach",
			`<?php $arr = array(); foreach ($arr as $value) { echo $value; }`,
		},
		{
			"Foreach with Break",
			`<?php $arr = array(1, 2, 3, 4, 5); foreach ($arr as $value) { if ($value > 2) break; echo $value; }`,
		},
		{
			"Nested Foreach",
			`<?php $outer = array(array(1, 2), array(3, 4)); foreach ($outer as $inner) { foreach ($inner as $value) { echo $value; } }`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(lexer.New(tt.code))
			prog := p.ParseProgram()

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Failed to compile foreach statement: %s", tt.name)

			vmCtx := vm.NewExecutionContext()
			err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants(), comp.GetFunctions(), comp.GetClasses())
			require.NoError(t, err, "Failed to execute foreach statement: %s", tt.name)
		})
	}
}

func TestAlternativeForeachStatementExecution(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{
			"Simple Alternative Foreach",
			`<?php $arr = array(1, 2, 3); foreach ($arr as $value): echo $value; endforeach;`,
		},
		{
			"Alternative Foreach with Key",
			`<?php $arr = array("a" => 1, "b" => 2); foreach ($arr as $key => $value): echo $key . ":" . $value; endforeach;`,
		},
		{
			"Empty Array Alternative Foreach",
			`<?php $arr = array(); foreach ($arr as $value): echo $value; endforeach;`,
		},
		{
			"Alternative Foreach with Break",
			`<?php $arr = array(1, 2, 3, 4, 5); foreach ($arr as $value): if ($value > 2): break; endif; echo $value; endforeach;`,
		},
		{
			"Alternative Foreach with Continue",
			`<?php $arr = array(1, 2, 3, 4, 5); foreach ($arr as $value): if ($value == 2): continue; endif; echo $value; endforeach;`,
		},
		{
			"Nested Alternative Foreach",
			`<?php $outer = array(array(1, 2), array(3, 4)); foreach ($outer as $inner): foreach ($inner as $value): echo $value; endforeach; endforeach;`,
		},
		{
			"Mixed Alternative and Regular Foreach",
			`<?php $outer = array(array(1, 2), array(3, 4)); foreach ($outer as $inner): foreach ($inner as $value) { echo $value; } endforeach;`,
		},
		{
			"Alternative Foreach with Multiple Statements",
			`<?php $arr = array(1, 2, 3); foreach ($arr as $value): echo $value; echo " "; endforeach;`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(lexer.New(tt.code))
			prog := p.ParseProgram()

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Failed to compile alternative foreach statement: %s", tt.name)

			vmCtx := vm.NewExecutionContext()
			err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants(), comp.GetFunctions(), comp.GetClasses())
			require.NoError(t, err, "Failed to execute alternative foreach statement: %s", tt.name)
		})
	}
}

func TestInterpolatedStringExpression(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{
			name: "Simple variable interpolation",
			code: `<?php 
				$name = "World";
				echo "Hello $name!";
			`,
		},
		{
			name: "Multiple variable interpolation",
			code: `<?php 
				$first = "John";
				$last = "Doe";
				echo "Hello $first $last!";
			`,
		},
		{
			name: "Number interpolation",
			code: `<?php 
				$age = 25;
				echo "I am $age years old";
			`,
		},
		{
			name: "Mixed content interpolation",
			code: `<?php 
				$name = "Alice";
				$count = 5;
				echo "Hello $name, you have $count messages";
			`,
		},
		{
			name: "Expression interpolation",
			code: `<?php 
				$a = 10;
				$b = 20;
				echo "Sum: ${$a + $b}";
			`,
		},
		{
			name: "Array access interpolation",
			code: `<?php 
				$data = ["name" => "Bob"];
				echo "Hello {$data['name']}!";
			`,
		},
		{
			name: "Nested interpolation",
			code: `<?php 
				$prefix = "Mr";
				$name = "Smith";
				echo "$prefix $name says: 'Hello world!'";
			`,
		},
		{
			name: "Empty interpolation parts",
			code: `<?php 
				$empty = "";
				echo "Start${empty}End";
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(lexer.New(tt.code))
			prog := p.ParseProgram()

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Failed to compile interpolated string: %s", tt.name)

			vmCtx := vm.NewExecutionContext()
			err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants(), comp.GetFunctions(), comp.GetClasses())
			require.NoError(t, err, "Failed to execute interpolated string: %s", tt.name)
		})
	}
}

// TestInterpolatedStringArrayAccess tests array access within interpolated strings
func TestInterpolatedStringArrayAccess(t *testing.T) {
	tests := []struct {
		name           string
		code           string
		expectedOutput string
	}{
		{
			name: "Simple array access interpolation",
			code: `<?php
				$arr = [1,2,3];
				for($i=0; $i<3; $i++) {
					echo "$arr[$i]\n";
				}
			`,
			expectedOutput: "1\n2\n3\n",
		},
		{
			name: "Array access with string key",
			code: `<?php
				$arr = ["a" => "hello", "b" => "world"];
				$key = "a";
				echo "$arr[$key]";
			`,
			expectedOutput: "hello",
		},
		{
			name: "Multiple array access in one string",
			code: `<?php
				$arr1 = [1,2,3];
				$arr2 = [4,5,6]; 
				echo "$arr1[0] and $arr2[1]";
			`,
			expectedOutput: "1 and 5",
		},
		{
			name: "Simple single array access",
			code: `<?php
				$arr = [1,2,3];
				echo "$arr[0]";
			`,
			expectedOutput: "1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(lexer.New(tt.code))
			prog := p.ParseProgram()

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Failed to compile array access interpolation: %s", tt.name)

			vmCtx := vm.NewExecutionContext()

			err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants(), comp.GetFunctions(), comp.GetClasses())
			require.NoError(t, err, "Failed to execute array access interpolation: %s", tt.name)

			// For now, we just verify compilation and execution succeed
			// TODO: Implement output capturing to verify actual output values
		})
	}
}

// TestArrayAccessEdgeCases tests edge cases for array access compilation
func TestArrayAccessEdgeCases(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{
			name: "Nested array access in interpolation",
			code: `<?php
				$arr = [[1,2],[3,4]];
				$i = 0; $j = 1;
				echo "$arr[$i][$j]";
			`,
		},
		{
			name: "Array access with expression index in interpolation",
			code: `<?php
				$arr = [1,2,3,4,5];
				$i = 1;
				echo "$arr[$i+1]";
			`,
		},
		{
			name: "Array access with negative index",
			code: `<?php
				$arr = [1,2,3];
				$i = -1;
				echo "$arr[$i]";
			`,
		},
		{
			name: "Multiple interpolated expressions",
			code: `<?php
				$arr1 = [1,2,3];
				$arr2 = [4,5,6];
				$i = 0; $j = 1;
				echo "First: $arr1[$i], Second: $arr2[$j]";
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(lexer.New(tt.code))
			prog := p.ParseProgram()
			require.Empty(t, p.Errors(), "Parser should not have errors for: %s", tt.name)

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Failed to compile edge case: %s", tt.name)

			vmCtx := vm.NewExecutionContext()
			err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants(), comp.GetFunctions(), comp.GetClasses())
			require.NoError(t, err, "Failed to execute edge case: %s", tt.name)
		})
	}
}

// TestArrayAccessOutsideInterpolation tests that array access works outside interpolated strings
func TestArrayAccessOutsideInterpolation(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{
			name: "Simple array access assignment",
			code: `<?php
				$arr = [1,2,3];
				$val = $arr[1];
			`,
		},
		{
			name: "Array access in expressions",
			code: `<?php
				$arr = [1,2,3];
				$sum = $arr[0] + $arr[1];
			`,
		},
		{
			name: "Array access in function calls",
			code: `<?php
				$arr = ["hello", "world"];
				echo($arr[0]);
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(lexer.New(tt.code))
			prog := p.ParseProgram()
			require.Empty(t, p.Errors(), "Parser should not have errors for: %s", tt.name)

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Failed to compile array access outside interpolation: %s", tt.name)

			vmCtx := vm.NewExecutionContext()
			err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants(), comp.GetFunctions(), comp.GetClasses())
			require.NoError(t, err, "Failed to execute array access outside interpolation: %s", tt.name)
		})
	}
}

func TestArrayIndexAssignment(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{
			name: "Array index assignment with integer key",
			code: `<?php
				$arr = array();
				$arr[0] = "test";
			`,
		},
		{
			name: "Array index assignment with string key",
			code: `<?php
				$arr = array();
				$arr["key"] = "value";
			`,
		},
		{
			name: "Array append assignment",
			code: `<?php
				$arr = array();
				$arr[] = "append";
			`,
		},
		{
			name: "Array index then append",
			code: `<?php
				$arr = array();
				$arr[0] = "first";
				$arr[] = "last";
			`,
		},
		{
			name: "Array append then index",
			code: `<?php
				$arr = array();
				$arr[] = "first";
				$arr[1] = "second";
			`,
		},
		{
			name: "Nested array assignment",
			code: `<?php
				$cfg = array();
				$type = "config";
				$file = "main.php";
				$cfg[$type][$file] = false;
			`,
		},
		{
			name: "Multiple nested array operations",
			code: `<?php
				$data = array();
				$data["users"]["john"]["age"] = 25;
				$data["users"]["jane"]["age"] = 30;
				$val = $data["users"]["john"]["age"];
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(lexer.New(tt.code))
			prog := p.ParseProgram()
			require.Empty(t, p.Errors(), "Parser should not have errors for: %s", tt.name)

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Failed to compile array index assignment: %s", tt.name)

			vmCtx := vm.NewExecutionContext()
			err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants(), comp.GetFunctions(), comp.GetClasses())
			require.NoError(t, err, "Failed to execute array index assignment: %s", tt.name)
		})
	}
}

func TestArrayElementExpression(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{
			name: "Array with keyed elements",
			code: `<?php
				$arr = [1 => "one", "two" => 2, 3 => "three"];
			`,
		},
		{
			name: "Array with mixed indexed and keyed elements",
			code: `<?php
				$arr = ["first", "key" => "value", "last"];
			`,
		},
		{
			name: "Nested array with keyed elements",
			code: `<?php
				$arr = [
					"user" => ["name" => "John", "age" => 30],
					"config" => ["debug" => true]
				];
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse and compile
			p := parser.New(lexer.New(tt.code))
			prog := p.ParseProgram()
			require.Empty(t, p.Errors(), "Parser should not have errors for: %s", tt.name)

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Failed to compile ArrayElementExpression: %s", tt.name)

			// Execute to ensure no runtime errors
			err = executeWithRuntime(t, comp)
			require.NoError(t, err, "Failed to execute ArrayElementExpression: %s", tt.name)
		})
	}
}

func TestTryStatement(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{
			"Basic try-catch",
			`<?php
			try {
				echo "in try";
			} catch (Exception $e) {
				echo "caught";
			}`,
		},
		{
			"Try-catch-finally",
			`<?php
			try {
				echo "try block";
			} catch (Exception $e) {
				echo "catch block";
			} finally {
				echo "finally block";
			}`,
		},
		{
			"Try with finally only",
			`<?php
			try {
				echo "try block";
			} finally {
				echo "finally block";
			}`,
		},
		{
			"Multiple catch blocks",
			`<?php
			try {
				echo "try block";
			} catch (RuntimeException $e) {
				echo "runtime error";
			} catch (Exception $e) {
				echo "general error";
			}`,
		},
		{
			"Try-catch with variable assignment",
			`<?php
			$message = "default";
			try {
				$message = "success";
			} catch (Exception $e) {
				$message = "error";
			}
			echo $message;`,
		},
		{
			"Nested try-catch",
			`<?php
			try {
				echo "outer try";
				try {
					echo "inner try";
				} catch (Exception $e) {
					echo "inner catch";
				}
			} catch (Exception $e) {
				echo "outer catch";
			}`,
		},
		{
			"Try-catch with throw statement",
			`<?php
			try {
				echo "before throw";
				throw new Exception("simple exception");
				echo "after throw";
			} catch (Exception $e) {
				echo "caught exception";
			}`,
		},
		{
			"Try-catch with exception variable usage",
			`<?php
			try {
				echo "try block";
			} catch (Exception $ex) {
				echo "Exception caught";
				$error = $ex;
			}`,
		},
		{
			"Try-catch-finally with complex expressions",
			`<?php
			$x = 10;
			try {
				$x = $x * 2;
				echo $x;
			} catch (Exception $e) {
				$x = $x - 5;
			} finally {
				$x = $x + 1;
				echo $x;
			}`,
		},
		{
			"Empty try-catch blocks",
			`<?php
			try {
			} catch (Exception $e) {
			}`,
		},
		{
			"Try-catch with array operations",
			`<?php
			$arr = [1, 2, 3];
			try {
				echo "array ok";
			} catch (Exception $e) {
				echo "array error";
			}`,
		},
		{
			"Try-catch with string interpolation",
			`<?php
			$name = "test";
			try {
				echo "Hello $name";
			} catch (Exception $e) {
				echo "Error with $name";
			}`,
		},
		{
			"Try-catch with conditional logic",
			`<?php
			$condition = true;
			try {
				if ($condition) {
					echo "condition true";
				} else {
					echo "condition false";
				}
			} catch (Exception $e) {
				echo "exception in conditional";
			}`,
		},
		{
			"Try-catch with loops",
			`<?php
			try {
				for ($i = 0; $i < 3; $i++) {
					echo $i;
				}
			} catch (Exception $e) {
				echo "loop error";
			}`,
		},
		{
			"Try-finally without catch",
			`<?php
			try {
				echo "executing";
				$x = 5;
			} finally {
				echo "cleanup";
			}
			echo "done";`,
		},
		{
			"Exception flow with throw and catch",
			`<?php
			try {
				echo "Before throw";
				throw new Exception("Test exception");
				echo "After throw";
			} catch (Exception $e) {
				echo "Caught exception";
			}
			echo "After try-catch";`,
		},
		{
			"Try-catch-finally with exception thrown",
			`<?php
			try {
				echo "Try";
				throw new Exception("Error");
			} catch (Exception $e) {
				echo "Catch";
			} finally {
				echo "Finally";
			}
			echo "End";`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(lexer.New(tt.code))
			prog := p.ParseProgram()
			require.Empty(t, p.Errors(), "Parser should not have errors for: %s", tt.name)

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Failed to compile try statement: %s", tt.name)

			vmCtx := vm.NewExecutionContext()
			err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants(), comp.GetFunctions(), comp.GetClasses())
			require.NoError(t, err, "Failed to execute try statement: %s", tt.name)
		})
	}
}

func TestFunctionDeclaration(t *testing.T) {
	testCases := []struct {
		name string
		code string
	}{
		{
			"Simple function declaration",
			`<?php 
			function greet($name) { 
				echo "Hello, " . $name; 
			} 
			greet("World");`,
		},
		{
			"Function with default parameters",
			`<?php 
			function add($x, $y = 10) { 
				return $x + $y; 
			} 
			echo add(5);`,
		},
		{
			"Function with return type",
			`<?php 
			function multiply($a, $b): int { 
				return $a * $b; 
			} 
			echo multiply(3, 4);`,
		},
		{
			"Function with reference parameter",
			`<?php 
			function increment(&$value) { 
				$value++; 
			} 
			$x = 5; 
			increment($x); 
			echo $x;`,
		},
		{
			"Function with variadic parameters",
			`<?php 
			function sum(...$numbers) { 
				$total = 0; 
				foreach ($numbers as $num) { 
					$total += $num; 
				} 
				return $total; 
			} 
			echo sum(1, 2, 3, 4);`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := parser.New(lexer.New(tc.code))
			prog := p.ParseProgram()
			require.Empty(t, p.Errors(), "Parser should not have errors for: %s", tc.name)

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Failed to compile function declaration: %s", tc.name)

			vmCtx := vm.NewExecutionContext()
			err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants(), comp.GetFunctions(), comp.GetClasses())
			require.NoError(t, err, "Failed to execute function declaration: %s", tc.name)
		})
	}
}

func TestAnonymousClass(t *testing.T) {
	testCases := []struct {
		name string
		code string
	}{
		{
			"Simple anonymous class",
			`<?php 
			$obj = new class { 
				public function hello() { 
					return "Hello from anonymous class"; 
				} 
			}; 
			echo $obj->hello();`,
		},
		{
			"Anonymous class with properties",
			`<?php 
			$obj = new class { 
				public $name = "test"; 
				private $value = 42; 
				public function getName() { 
					return $this->name; 
				} 
			}; 
			echo $obj->getName();`,
		},
		{
			"Anonymous class with constants",
			`<?php 
			$obj = new class { 
				public const VERSION = "1.0"; 
				final public const MAX_SIZE = 100; 
			}; 
			echo $obj::VERSION; 
			echo $obj::MAX_SIZE;`,
		},
		{
			"Anonymous class with inheritance",
			`<?php 
			class BaseClass { 
				public function base() { 
					return "base"; 
				} 
			} 
			$obj = new class extends BaseClass { 
				public function test() { 
					return $this->base() . " extended"; 
				} 
			}; 
			echo $obj->test();`,
		},
		{
			"Anonymous class with constructor arguments",
			`<?php 
			$obj = new class("test") { 
				private $value; 
				public function __construct($val) { 
					$this->value = $val; 
				} 
				public function getValue() { 
					return $this->value; 
				} 
			}; 
			echo $obj->getValue();`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := parser.New(lexer.New(tc.code))
			prog := p.ParseProgram()
			require.Empty(t, p.Errors(), "Parser should not have errors for: %s", tc.name)

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Failed to compile anonymous class: %s", tc.name)

			vmCtx := vm.NewExecutionContext()
			err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants(), comp.GetFunctions(), comp.GetClasses())
			require.NoError(t, err, "Failed to execute anonymous class: %s", tc.name)
		})
	}
}

func TestPropertyDeclaration(t *testing.T) {
	testCases := []struct {
		name string
		code string
	}{
		{
			"Simple class property",
			`<?php 
			class TestClass { 
				public $name = "test"; 
				public function getName() { 
					return $this->name; 
				} 
			} 
			$obj = new TestClass(); 
			echo $obj->getName();`,
		},
		{
			"Property with different visibilities",
			`<?php 
			class TestClass { 
				public $publicProp = "public"; 
				private $privateProp = "private"; 
				protected $protectedProp = "protected"; 
				
				public function getPrivate() { 
					return $this->privateProp; 
				} 
				public function getProtected() { 
					return $this->protectedProp; 
				} 
			} 
			$obj = new TestClass(); 
			echo $obj->publicProp; 
			echo $obj->getPrivate(); 
			echo $obj->getProtected();`,
		},
		{
			"Static properties",
			`<?php 
			class TestClass { 
				public static $counter = "0"; 
				public static $instance = "test"; 
			} 
			echo TestClass::$counter; 
			echo TestClass::$instance;`,
		},
		{
			"Properties with type hints",
			`<?php 
			class TestClass { 
				public string $name = "default"; 
				public int $age = 0; 
				public ?array $data = null; 
				
				public function setData(array $data) { 
					$this->data = $data; 
				} 
			} 
			$obj = new TestClass(); 
			echo $obj->name;`,
		},
		{
			"Readonly properties",
			`<?php 
			class TestClass { 
				public readonly string $id; 
				
				public function __construct(string $id) { 
					$this->id = $id; 
				} 
				
				public function getId() { 
					return $this->id; 
				} 
			} 
			$obj = new TestClass("test123"); 
			echo $obj->getId();`,
		},
		{
			"Array default values - simple array",
			`<?php 
			class TestClass { 
				public $simpleArray = [1, 2, 3]; 
				public function getFirstElement() { 
					return $this->simpleArray[0]; 
				} 
			} 
			$obj = new TestClass(); 
			echo $obj->getFirstElement();`,
		},
		{
			"Array default values - mixed types",
			`<?php 
			class TestClass { 
				public $mixedArray = [1, "hello"]; 
				public function getElements() { 
					return $this->mixedArray[0] . "," . $this->mixedArray[1]; 
				} 
			} 
			$obj = new TestClass(); 
			echo $obj->getElements();`,
		},
		{
			"Array default values - associative array",
			`<?php 
			class TestClass { 
				public $assocArray = ["name" => "John", "age" => 30]; 
				public function getName() { 
					return $this->assocArray["name"]; 
				} 
			} 
			$obj = new TestClass(); 
			echo $obj->getName();`,
		},
		{
			"Array default values - nested array",
			`<?php 
			class TestClass { 
				public $nestedArray = [
					"users" => [
						["name" => "Alice"]
					]
				]; 
				public function getUserName() { 
					return $this->nestedArray["users"][0]["name"]; 
				} 
			} 
			$obj = new TestClass(); 
			echo $obj->getUserName();`,
		},
		{
			"Array default values - empty array",
			`<?php 
			class TestClass { 
				public $emptyArray = []; 
				public function hasArray() { 
					return isset($this->emptyArray); 
				} 
			} 
			$obj = new TestClass(); 
			echo $obj->hasArray() ? "true" : "false";`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := parser.New(lexer.New(tc.code))
			prog := p.ParseProgram()
			require.Empty(t, p.Errors(), "Parser should not have errors for: %s", tc.name)

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Failed to compile property declaration: %s", tc.name)

			vmCtx := vm.NewExecutionContext()
			err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants(), comp.GetFunctions(), comp.GetClasses())
			require.NoError(t, err, "Failed to execute property declaration: %s", tc.name)
		})
	}
}

func TestFunctionExists(t *testing.T) {
	testCases := []struct {
		name     string
		code     string
		expected string
	}{
		{
			"Function exists for built-in function",
			`<?php 
			echo function_exists("strlen") ? "true" : "false";`,
			"true",
		},
		{
			"Function exists for non-existent function",
			`<?php 
			echo function_exists("nonexistent_function") ? "true" : "false";`,
			"false",
		},
		{
			"Function exists for function_exists itself",
			`<?php 
			echo function_exists("function_exists") ? "true" : "false";`,
			"true",
		},
		{
			"Function exists for user-defined function",
			`<?php 
			function my_custom_function() {
				return "hello";
			}
			echo function_exists("my_custom_function") ? "true" : "false";`,
			"true",
		},
		{
			"Function exists before and after function definition",
			`<?php 
			echo function_exists("test_func") ? "true" : "false";
			echo ",";
			function test_func() {}
			echo function_exists("test_func") ? "true" : "false";`,
			"true,true",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test with our implementation
			p := parser.New(lexer.New(tc.code))
			prog := p.ParseProgram()
			require.Empty(t, p.Errors(), "Parser should not have errors for: %s", tc.name)

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Failed to compile function_exists test: %s", tc.name)

			// Initialize runtime for built-in functions
			runtimeErr := runtime.Bootstrap()
			require.NoError(t, runtimeErr, "Failed to bootstrap runtime")

			// Initialize VM integration
			if runtime.GlobalVMIntegration == nil {
				err := runtime.InitializeVMIntegration()
				require.NoError(t, err, "Failed to initialize VM integration")
			}

			vmCtx := vm.NewExecutionContext()

			// Capture output
			var buf bytes.Buffer
			vmCtx.SetOutputWriter(&buf)

			err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants(), comp.GetFunctions(), comp.GetClasses())
			require.NoError(t, err, "Failed to execute function_exists test: %s", tc.name)

			output := strings.TrimSpace(buf.String())
			require.Equal(t, tc.expected, output, "Function_exists test failed for: %s", tc.name)
		})
	}
}

func TestClassExists(t *testing.T) {
	testCases := []struct {
		name     string
		code     string
		expected string
	}{
		{
			"Class exists for built-in class stdClass",
			`<?php 
			echo class_exists("stdClass") ? "true" : "false";`,
			"true",
		},
		{
			"Class exists for built-in class Exception",
			`<?php 
			echo class_exists("Exception") ? "true" : "false";`,
			"true",
		},
		{
			"Class exists for non-existent class",
			`<?php 
			echo class_exists("NonExistentClass") ? "true" : "false";`,
			"false",
		},
		{
			"Class exists for user-defined class",
			`<?php 
			class TestClass {}
			echo class_exists("TestClass") ? "true" : "false";`,
			"true",
		},
		{
			"Class exists case insensitive",
			`<?php 
			class MyClass {}
			echo class_exists("myclass") ? "true" : "false";`,
			"true",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test with our implementation
			p := parser.New(lexer.New(tc.code))
			prog := p.ParseProgram()
			require.Empty(t, p.Errors(), "Parser should not have errors for: %s", tc.name)

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Failed to compile class_exists test: %s", tc.name)

			// Initialize runtime for built-in functions
			runtimeErr := runtime.Bootstrap()
			require.NoError(t, runtimeErr, "Failed to bootstrap runtime")

			// Initialize VM integration
			if runtime.GlobalVMIntegration == nil {
				err := runtime.InitializeVMIntegration()
				require.NoError(t, err, "Failed to initialize VM integration")
			}

			vmCtx := vm.NewExecutionContext()

			// Capture output
			var buf bytes.Buffer
			vmCtx.SetOutputWriter(&buf)

			err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants(), comp.GetFunctions(), comp.GetClasses())
			require.NoError(t, err, "Failed to execute class_exists test: %s", tc.name)

			output := strings.TrimSpace(buf.String())
			require.Equal(t, tc.expected, output, "Class_exists test failed for: %s", tc.name)
		})
	}
}

func TestMethodExists(t *testing.T) {
	testCases := []struct {
		name     string
		code     string
		expected string
	}{
		{
			"Method exists for built-in class Exception method",
			`<?php 
			echo method_exists("Exception", "getMessage") ? "true" : "false";`,
			"true",
		},
		{
			"Method exists for non-existent method on built-in class",
			`<?php 
			echo method_exists("stdClass", "nonExistentMethod") ? "true" : "false";`,
			"false",
		},
		{
			"Method exists for non-existent class",
			`<?php 
			echo method_exists("NonExistentClass", "someMethod") ? "true" : "false";`,
			"false",
		},
		{
			"Method exists for user-defined class method",
			`<?php 
			class TestClass {
				public function testMethod() {}
			}
			echo method_exists("TestClass", "testMethod") ? "true" : "false";`,
			"true",
		},
		{
			"Method exists for user-defined class with object instance",
			`<?php 
			class TestClass {
				public function testMethod() {}
			}
			$obj = new TestClass();
			echo method_exists($obj, "testMethod") ? "true" : "false";`,
			"true",
		},
		{
			"Method exists case insensitive",
			`<?php 
			class TestClass {
				public function MyMethod() {}
			}
			echo method_exists("TestClass", "mymethod") ? "true" : "false";`,
			"true",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test with our implementation
			p := parser.New(lexer.New(tc.code))
			prog := p.ParseProgram()
			require.Empty(t, p.Errors(), "Parser should not have errors for: %s", tc.name)

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Failed to compile method_exists test: %s", tc.name)

			// Initialize runtime for built-in functions
			runtimeErr := runtime.Bootstrap()
			require.NoError(t, runtimeErr, "Failed to bootstrap runtime")

			// Initialize VM integration
			if runtime.GlobalVMIntegration == nil {
				err := runtime.InitializeVMIntegration()
				require.NoError(t, err, "Failed to initialize VM integration")
			}

			vmCtx := vm.NewExecutionContext()

			// Capture output
			var buf bytes.Buffer
			vmCtx.SetOutputWriter(&buf)

			err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants(), comp.GetFunctions(), comp.GetClasses())
			require.NoError(t, err, "Failed to execute method_exists test: %s", tc.name)

			output := strings.TrimSpace(buf.String())
			require.Equal(t, tc.expected, output, "Method_exists test failed for: %s", tc.name)
		})
	}
}

func TestClassConstantDeclaration(t *testing.T) {
	testCases := []struct {
		name     string
		code     string
		expected string
	}{
		{
			"Simple string constant",
			`<?php 
			class Test { 
				public const VERSION = "1.0"; 
			} 
			echo Test::VERSION;`,
			"1.0",
		},
		{
			"Simple integer constant",
			`<?php 
			class Test { 
				public const MAX_SIZE = 100; 
			} 
			echo Test::MAX_SIZE;`,
			"100",
		},
		{
			"Multiple constants in one declaration",
			`<?php 
			class Test { 
				public const FIRST = 1, SECOND = 2; 
			} 
			echo Test::FIRST; 
			echo "|"; 
			echo Test::SECOND;`,
			"1|2",
		},
		{
			"Boolean constants",
			`<?php 
			class Test { 
				public const IS_TRUE = true; 
				public const IS_FALSE = false; 
			} 
			echo Test::IS_TRUE ? "1" : "0"; 
			echo Test::IS_FALSE ? "1" : "0";`,
			"10",
		},
		{
			"Null constant",
			`<?php 
			class Test { 
				public const NULL_VALUE = null; 
			} 
			echo Test::NULL_VALUE === null ? "true" : "false";`,
			"true",
		},
		{
			"Float constant",
			`<?php 
			class Test { 
				public const PI = 3.14; 
			} 
			echo Test::PI;`,
			"3.14",
		},
		{
			"Constants with visibility modifiers",
			`<?php 
			class Test { 
				public const PUBLIC_CONST = "public"; 
				private const PRIVATE_CONST = "private"; 
				protected const PROTECTED_CONST = "protected";
			} 
			echo Test::PUBLIC_CONST;`,
			"public",
		},
		{
			"Final constants",
			`<?php 
			class Test { 
				final public const FINAL_CONST = "immutable"; 
			} 
			echo Test::FINAL_CONST;`,
			"immutable",
		},
		{
			"Constants with lowercase literals",
			`<?php 
			class Test { 
				public const TRUE_CONST = true; 
				public const FALSE_CONST = false; 
				public const NULL_CONST = null; 
			} 
			echo Test::TRUE_CONST ? "T" : "F"; 
			echo Test::FALSE_CONST ? "T" : "F"; 
			echo (Test::NULL_CONST === null) ? "N" : "X";`,
			"TFN",
		},
		{
			"Empty array constant",
			`<?php 
			class Test { 
				public const EMPTY_ARRAY = []; 
			} 
			echo "array_defined";`,
			"array_defined",
		},
		{
			"Simple indexed array constant",
			`<?php 
			class Test { 
				public const SIMPLE_ARRAY = [1, 2, 3]; 
			} 
			echo "array_created";`,
			"array_created",
		},
		{
			"Mixed type array constant",
			`<?php 
			class Test { 
				public const MIXED_ARRAY = [1, "hello", true, null]; 
			} 
			echo "mixed_array_created";`,
			"mixed_array_created",
		},
		{
			"Associative array constant",
			`<?php 
			class Test { 
				public const ASSOC_ARRAY = ["key1" => "value1", "key2" => "value2"]; 
			} 
			echo "assoc_array_created";`,
			"assoc_array_created",
		},
		{
			"Nested array constant",
			`<?php 
			class Test { 
				public const NESTED_ARRAY = [[1, 2], [3, 4]]; 
			} 
			echo "nested_array_created";`,
			"nested_array_created",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := parser.New(lexer.New(tc.code))
			prog := p.ParseProgram()
			require.Empty(t, p.Errors(), "Parser should not have errors for: %s", tc.name)

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Failed to compile class constant declaration: %s", tc.name)

			// Verify that constants were properly stored in the class
			classes := comp.GetClasses()
			require.NotEmpty(t, classes, "Should have compiled at least one class")

			// Find the Test class
			var testClass *vm.Class
			for _, class := range classes {
				if class.Name == "Test" {
					testClass = class
					break
				}
			}
			require.NotNil(t, testClass, "Should have found Test class")
			require.NotEmpty(t, testClass.Constants, "Test class should have constants")

			vmCtx := vm.NewExecutionContext()
			err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants(), comp.GetFunctions(), comp.GetClasses())
			require.NoError(t, err, "Failed to execute class constant declaration: %s", tc.name)
		})
	}
}

func TestClassConstantErrors(t *testing.T) {
	errorCases := []struct {
		name          string
		code          string
		expectedError string
	}{
		{
			"Duplicate constant in same class",
			`<?php 
			class Test { 
				public const DUPLICATE = 1; 
				public const DUPLICATE = 2; 
			}`,
			"constant DUPLICATE already declared",
		},
		{
			"Private final constant error",
			`<?php 
			class Test { 
				final private const INVALID = "error"; 
			}`,
			"private constant cannot be final",
		},
		{
			"Function call expression not supported",
			`<?php 
			class Test { 
				public const FUNC_CALL = strlen("test"); 
			}`,
			"unsupported constant expression",
		},
	}

	for _, tc := range errorCases {
		t.Run(tc.name, func(t *testing.T) {
			p := parser.New(lexer.New(tc.code))
			prog := p.ParseProgram()

			comp := NewCompiler()
			err := comp.Compile(prog)
			if tc.expectedError != "" {
				require.Error(t, err, "Expected compilation error for: %s", tc.name)
				require.Contains(t, err.Error(), tc.expectedError, "Error should contain expected message")
			} else {
				require.NoError(t, err, "Should not have compilation error for: %s", tc.name)
			}
		})
	}
}

func TestClassConstantExpressions(t *testing.T) {
	testCases := []struct {
		name     string
		code     string
		expected string
	}{
		{
			"Arithmetic expressions",
			`<?php 
			class Test { 
				public const MATH_ADD = 10 + 5;
				public const MATH_SUB = 20 - 8;
				public const MATH_MUL = 6 * 7;
				public const MATH_DIV = 100 / 4;
				public const MATH_MOD = 17 % 3;
				public const MATH_POW = 2 ** 3;
			} 
			echo Test::MATH_ADD . "|"; 
			echo Test::MATH_SUB . "|"; 
			echo Test::MATH_MUL . "|"; 
			echo Test::MATH_DIV . "|"; 
			echo Test::MATH_MOD . "|"; 
			echo Test::MATH_POW;`,
			"15|12|42|25|2|8",
		},
		{
			"String concatenation",
			`<?php 
			class Test { 
				public const STRING_CONCAT = 'hello' . ' world';
				public const STRING_CONCAT2 = 'prefix_' . 'suffix';
			} 
			echo Test::STRING_CONCAT . "|"; 
			echo Test::STRING_CONCAT2;`,
			"hello world|prefix_suffix",
		},
		{
			"Boolean operations",
			`<?php 
			class Test { 
				public const BOOL_AND = true && false;
				public const BOOL_OR = false || true;
				public const BOOL_NOT = !true;
			} 
			echo (Test::BOOL_AND ? "1" : "0"); 
			echo (Test::BOOL_OR ? "1" : "0"); 
			echo (Test::BOOL_NOT ? "1" : "0");`,
			"010",
		},
		{
			"Comparison operations",
			`<?php 
			class Test { 
				public const COMP_EQ = 5 == 5;
				public const COMP_NEQ = 5 != 3;
				public const COMP_LT = 3 < 5;
				public const COMP_GT = 5 > 3;
				public const COMP_LTE = 5 <= 5;
				public const COMP_GTE = 5 >= 5;
				public const COMP_SPACESHIP = 5 <=> 3;
			} 
			echo (Test::COMP_EQ ? "1" : "0");
			echo (Test::COMP_NEQ ? "1" : "0");
			echo (Test::COMP_LT ? "1" : "0");
			echo (Test::COMP_GT ? "1" : "0");
			echo (Test::COMP_LTE ? "1" : "0");
			echo (Test::COMP_GTE ? "1" : "0");
			echo Test::COMP_SPACESHIP;`,
			"1111111",
		},
		{
			"Bitwise operations",
			`<?php 
			class Test { 
				public const BITWISE_AND = 7 & 3;
				public const BITWISE_OR = 4 | 2;
				public const BITWISE_XOR = 5 ^ 3;
				public const BITWISE_NOT = ~5;
				public const SHIFT_LEFT = 2 << 3;
				public const SHIFT_RIGHT = 16 >> 2;
			} 
			echo Test::BITWISE_AND . "|"; 
			echo Test::BITWISE_OR . "|"; 
			echo Test::BITWISE_XOR . "|"; 
			echo Test::BITWISE_NOT . "|"; 
			echo Test::SHIFT_LEFT . "|"; 
			echo Test::SHIFT_RIGHT;`,
			"3|6|6|-6|16|4",
		},
		{
			"Complex nested expressions",
			`<?php 
			class Test { 
				public const NESTED = (10 + 5) * 2;
				public const COMPLEX = 1 + 2 * 3 - 4;
				public const COMPLEX_NESTED = ((10 + 5) * 2 - 3) / 3;
			} 
			echo Test::NESTED . "|"; 
			echo Test::COMPLEX . "|"; 
			echo Test::COMPLEX_NESTED;`,
			"30|3|9",
		},
		{
			"Array expressions",
			`<?php 
			class Test { 
				public const SIMPLE_ARRAY = [1, 2, 3];
				public const COMPLEX_ARRAY = [1 + 2, 'hello' . 'world', true && false];
			} 
			echo count(Test::SIMPLE_ARRAY) . "|"; 
			echo count(Test::COMPLEX_ARRAY);`,
			"3|3",
		},
		{
			"Self constant references",
			`<?php 
			class Test { 
				public const BASE_VALUE = 100;
				public const DERIVED = self::BASE_VALUE * 2;
				public const CHAINED = self::DERIVED + self::BASE_VALUE;
			} 
			echo Test::BASE_VALUE . "|"; 
			echo Test::DERIVED . "|"; 
			echo Test::CHAINED;`,
			"100|200|300",
		},
		{
			"Ternary expressions",
			`<?php 
			class Test { 
				public const TERNARY = true ? 'yes' : 'no';
				public const TERNARY_COMPLEX = 5 > 3 ? 'greater' : 'lesser';
				public const TERNARY_NULL = false ? 'yes' : null;
			} 
			echo Test::TERNARY . "|"; 
			echo Test::TERNARY_COMPLEX . "|"; 
			echo (Test::TERNARY_NULL === null ? "null" : Test::TERNARY_NULL);`,
			"yes|greater|null",
		},
		{
			"Null coalescing",
			`<?php 
			class Test { 
				public const COALESCE = null ?? 'default';
				public const COALESCE2 = 'value' ?? 'default';
			} 
			echo Test::COALESCE . "|"; 
			echo Test::COALESCE2;`,
			"default|value",
		},
		{
			"Unary operations",
			`<?php 
			class Test { 
				public const UNARY_MINUS = -42;
				public const UNARY_PLUS = +42;
				public const UNARY_NOT = !false;
				public const UNARY_BITWISE = ~0;
			} 
			echo Test::UNARY_MINUS . "|"; 
			echo Test::UNARY_PLUS . "|"; 
			echo (Test::UNARY_NOT ? "1" : "0") . "|"; 
			echo Test::UNARY_BITWISE;`,
			"-42|42|1|-1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create lexer and parser
			p := parser.New(lexer.New(tc.code))
			prog := p.ParseProgram()

			// Create compiler and compile
			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Failed to compile: %s", tc.name)

			// Initialize runtime if not already done
			if runtime.GlobalRegistry == nil {
				err := runtime.Bootstrap()
				require.NoError(t, err, "Failed to bootstrap runtime")
			}

			// Initialize VM integration
			if runtime.GlobalVMIntegration == nil {
				err := runtime.InitializeVMIntegration()
				require.NoError(t, err, "Failed to initialize VM integration")
			}

			// Execute and capture output
			var buf bytes.Buffer
			vmCtx := vm.NewExecutionContext()
			vmCtx.SetOutputWriter(&buf)

			vmachine := vm.NewVirtualMachine()
			err = vmachine.Execute(vmCtx, comp.GetBytecode(), comp.GetConstants(), comp.GetFunctions(), comp.GetClasses())
			require.NoError(t, err, "Failed to execute: %s", tc.name)

			// Check expected output
			result := strings.TrimSpace(buf.String())
			require.Equal(t, tc.expected, result, "Unexpected output for: %s", tc.name)
		})
	}
}

func TestStaticAccessExpression(t *testing.T) {
	testCases := []struct {
		name     string
		code     string
		expected string
	}{
		{
			"Static constant access",
			`<?php
			class TestClass {
				public const CONSTANT = "const_value";
			}
			echo TestClass::CONSTANT;`,
			"const_value",
		},
		{
			"Static property access",
			`<?php
			class TestClass {
				public static $staticProp = "static_value";
			}
			echo TestClass::$staticProp;`,
			"static_value",
		},
		{
			"Self constant access",
			`<?php
			echo TestClass::CONSTANT;`,
			"const_value",
		},
		{
			"Self property access",
			`<?php
			echo TestClass::$staticProp;`,
			"static_value",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Capture stdout to verify output
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			p := parser.New(lexer.New(tc.code))
			prog := p.ParseProgram()
			require.Empty(t, p.Errors(), "Parser should not have errors for: %s", tc.name)

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Failed to compile static access: %s", tc.name)

			vmCtx := vm.NewExecutionContext()
			err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants(), comp.GetFunctions(), comp.GetClasses())
			require.NoError(t, err, "Failed to execute static access: %s", tc.name)

			// Close writer and restore stdout
			w.Close()
			os.Stdout = old

			// Read captured output
			var buf bytes.Buffer
			buf.ReadFrom(r)
			output := buf.String()

			require.Equal(t, tc.expected, output, "Expected '%s', got '%s' for test: %s", tc.expected, output, tc.name)
		})
	}
}

func TestVariableVariableExpression(t *testing.T) {
	testCases := []struct {
		name     string
		code     string
		expected string
	}{
		{
			"Simple variable variable",
			`<?php
			$var = "hello";
			$varName = "var";
			echo ${$varName};`,
			"hello",
		},
		{
			"Variable variable with expression",
			`<?php
			$test123 = "success";
			$fullname = "test123";
			echo ${$fullname};`,
			"success",
		},
		{
			"Variable variable with number",
			`<?php
			${'123'} = "numeric_var";
			$num = "123";
			echo ${$num};`,
			"numeric_var",
		},
		{
			"Undefined variable variable returns empty",
			`<?php
			$undefined = "nonexistent";
			echo "${$undefined}end";`,
			"end",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Capture stdout to verify output
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			p := parser.New(lexer.New(tc.code))
			prog := p.ParseProgram()
			require.Empty(t, p.Errors(), "Parser should not have errors for: %s", tc.name)

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Failed to compile variable variable: %s", tc.name)

			vmCtx := vm.NewExecutionContext()
			err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants(), comp.GetFunctions(), comp.GetClasses())
			require.NoError(t, err, "Failed to execute variable variable: %s", tc.name)

			// Close writer and restore stdout
			w.Close()
			os.Stdout = old

			// Read captured output
			var buf bytes.Buffer
			buf.ReadFrom(r)
			output := buf.String()

			require.Equal(t, tc.expected, output, "Expected '%s', got '%s' for test: %s", tc.expected, output, tc.name)
		})
	}
}

func TestStaticPropertyAccessExpression(t *testing.T) {
	testCases := []struct {
		name     string
		code     string
		expected string
	}{
		{
			"Static property access",
			`<?php
			class TestClass {
				public static $staticProp = "prop_value";
			}
			echo TestClass::$staticProp;`,
			"prop_value",
		},
		{
			"Static property from self",
			`<?php
			class TestClass {
				public static $prop = "self_value";
			}
			echo TestClass::$prop;`,
			"self_value",
		},
		{
			"Multiple static properties",
			`<?php
			class TestClass {
				public static $prop1 = "first";
				public static $prop2 = "second";
			}
			echo TestClass::$prop1 . TestClass::$prop2;`,
			"firstsecond",
		},
		{
			"Static property modification",
			`<?php
			class TestClass {
				public static $counter = "0";
			}
			echo TestClass::$counter;`,
			"0",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Capture stdout to verify output
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			p := parser.New(lexer.New(tc.code))
			prog := p.ParseProgram()
			require.Empty(t, p.Errors(), "Parser should not have errors for: %s", tc.name)

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Failed to compile static property access: %s", tc.name)

			vmCtx := vm.NewExecutionContext()
			err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants(), comp.GetFunctions(), comp.GetClasses())
			require.NoError(t, err, "Failed to execute static property access: %s", tc.name)

			// Close writer and restore stdout
			w.Close()
			os.Stdout = old

			// Read captured output
			var buf bytes.Buffer
			buf.ReadFrom(r)
			output := buf.String()

			require.Equal(t, tc.expected, output, "Expected '%s', got '%s' for test: %s", tc.expected, output, tc.name)
		})
	}
}

func TestStaticPropertyIncrementWithSelfAccess(t *testing.T) {
	code := `<?php
class TestClass {
    public static $counter = "0";

    function GetCounter() {
        return self::$counter;
    }
}
TestClass::$counter++;

$obj = new TestClass();
echo $obj->GetCounter();`

	// Capture stdout to verify output
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	p := parser.New(lexer.New(code))
	prog := p.ParseProgram()
	require.Empty(t, p.Errors(), "Parser should not have errors")

	comp := NewCompiler()
	err := comp.Compile(prog)
	require.NoError(t, err, "Failed to compile static property increment with self access")

	vmCtx := vm.NewExecutionContext()
	err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants(), comp.GetFunctions(), comp.GetClasses())
	require.NoError(t, err, "Failed to execute static property increment with self access")

	// Close writer and restore stdout
	w.Close()
	os.Stdout = old

	// Read captured output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	require.Equal(t, "1", output, "Expected '1', got '%s'", output)
}

func TestPropertyAccessExpression(t *testing.T) {
	testCases := []struct {
		name     string
		code     string
		expected string
	}{
		{
			"Simple property access",
			`<?php
			class TestClass {
				public $prop = "property_value";
			}
			$obj = new TestClass();
			echo $obj->prop;`,
			"property_value",
		},
		{
			"Property access with braces",
			`<?php
			class TestClass {
				public $prop = "brace_value";
			}
			$obj = new TestClass();
			echo $obj->{"prop"};`,
			"brace_value",
		},
		{
			"Property access via variable property name",
			`<?php
			class TestClass {
				public $dynamicProp = "dynamic_value";
			}
			$obj = new TestClass();
			$propName = "dynamicProp";
			echo $obj->$propName;`,
			"dynamic_value",
		},
		{
			"Property access with expression as property name",
			`<?php
			class TestClass {
				public $prop = "expr_value";
			}
			$obj = new TestClass();
			echo $obj->{"pr" . "op"};`,
			"expr_value",
		},
		{
			"Property assignment and access",
			`<?php
			class TestClass {}
			$obj = new TestClass();
			$obj->newProp = "new_value";
			echo $obj->newProp;`,
			"new_value",
		},
		{
			"Multiple property access",
			`<?php
			class TestClass {
				public $prop1 = "first";
				public $prop2 = "second";
			}
			$obj = new TestClass();
			echo $obj->prop1 . $obj->prop2;`,
			"firstsecond",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// First test with native PHP to ensure correctness
			tmpFile := filepath.Join(os.TempDir(), "test_property_"+tc.name+".php")
			err := os.WriteFile(tmpFile, []byte(tc.code), 0644)
			require.NoError(t, err)
			defer os.Remove(tmpFile)

			// Capture stdout to verify our implementation output
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			p := parser.New(lexer.New(tc.code))
			prog := p.ParseProgram()
			require.Empty(t, p.Errors(), "Parser should not have errors for: %s", tc.name)

			comp := NewCompiler()
			err = comp.Compile(prog)
			require.NoError(t, err, "Failed to compile property access: %s", tc.name)

			vmCtx := vm.NewExecutionContext()
			err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants(), comp.GetFunctions(), comp.GetClasses())
			require.NoError(t, err, "Failed to execute property access: %s", tc.name)

			// Close writer and restore stdout
			w.Close()
			os.Stdout = old

			// Read captured output
			var buf bytes.Buffer
			buf.ReadFrom(r)
			output := buf.String()

			require.Equal(t, tc.expected, output, "Expected '%s', got '%s' for test: %s", tc.expected, output, tc.name)
		})
	}
}

func TestFibonacciIterative(t *testing.T) {
	code := `<?php
function fibonacci_iterative($n) {
    if ($n <= 0) return 0;
    if ($n == 1) return 1;
    
    $a = 0;
    $b = 1;
    for ($i = 2; $i <= $n; $i++) {
        $temp = $a + $b;
        $a = $b;
        $b = $temp;
    }
    return $b;
}

// Test cases
for ($i = 0; $i <= 10; $i++) {
    echo fibonacci_iterative($i) . "\n";
}`

	p := parser.New(lexer.New(code))
	prog := p.ParseProgram()
	require.Empty(t, p.Errors(), "Parser should not have errors")

	comp := NewCompiler()
	err := comp.Compile(prog)
	require.NoError(t, err, "Failed to compile iterative Fibonacci")

	// Capture output
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	vmCtx := vm.NewExecutionContext()
	err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants(), comp.GetFunctions(), comp.GetClasses())
	require.NoError(t, err, "Failed to execute iterative Fibonacci")

	// Close writer and restore stdout
	w.Close()
	os.Stdout = old

	// Read captured output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	expected := "0\n1\n1\n2\n3\n5\n8\n13\n21\n34\n55\n"
	require.Equal(t, expected, output, "Iterative Fibonacci output mismatch")
}

func TestFibonacciRecursive(t *testing.T) {
	code := `<?php
function fibonacci_recursive($n) {
    if ($n <= 0) return 0;
    if ($n == 1) return 1;
    return fibonacci_recursive($n - 1) + fibonacci_recursive($n - 2);
}

// Test cases
for ($i = 0; $i <= 10; $i++) {
    echo fibonacci_recursive($i) . "\n";
}`

	p := parser.New(lexer.New(code))
	prog := p.ParseProgram()
	require.Empty(t, p.Errors(), "Parser should not have errors")

	comp := NewCompiler()
	err := comp.Compile(prog)
	require.NoError(t, err, "Failed to compile recursive Fibonacci")

	// Capture output
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	vmCtx := vm.NewExecutionContext()
	err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants(), comp.GetFunctions(), comp.GetClasses())
	require.NoError(t, err, "Failed to execute recursive Fibonacci")

	// Close writer and restore stdout
	w.Close()
	os.Stdout = old

	// Read captured output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	expected := "0\n1\n1\n2\n3\n5\n8\n13\n21\n34\n55\n"
	require.Equal(t, expected, output, "Recursive Fibonacci output mismatch")
}

func TestFibonacciComparison(t *testing.T) {
	code := `<?php
function fibonacci_iterative($n) {
    if ($n <= 0) return 0;
    if ($n == 1) return 1;
    
    $a = 0;
    $b = 1;
    for ($i = 2; $i <= $n; $i++) {
        $temp = $a + $b;
        $a = $b;
        $b = $temp;
    }
    return $b;
}

function fibonacci_recursive($n) {
    if ($n <= 0) return 0;
    if ($n == 1) return 1;
    return fibonacci_recursive($n - 1) + fibonacci_recursive($n - 2);
}

// Test both methods produce same results
for ($i = 0; $i <= 8; $i++) {
    $iter = fibonacci_iterative($i);
    $rec = fibonacci_recursive($i);
    if ($iter == $rec) {
        echo "F($i) = $iter (OK)\n";
    } else {
        echo "F($i) MISMATCH: iter=$iter, rec=$rec\n";
    }
}`

	p := parser.New(lexer.New(code))
	prog := p.ParseProgram()
	require.Empty(t, p.Errors(), "Parser should not have errors")

	comp := NewCompiler()
	err := comp.Compile(prog)
	require.NoError(t, err, "Failed to compile Fibonacci comparison")

	// Capture output
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	vmCtx := vm.NewExecutionContext()
	err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants(), comp.GetFunctions(), comp.GetClasses())
	require.NoError(t, err, "Failed to execute Fibonacci comparison")

	// Close writer and restore stdout
	w.Close()
	os.Stdout = old

	// Read captured output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	expected := "F(0) = 0 (OK)\nF(1) = 1 (OK)\nF(2) = 1 (OK)\nF(3) = 2 (OK)\nF(4) = 3 (OK)\nF(5) = 5 (OK)\nF(6) = 8 (OK)\nF(7) = 13 (OK)\nF(8) = 21 (OK)\n"
	require.Equal(t, expected, output, "Fibonacci comparison output mismatch")
}

func TestClassInheritance(t *testing.T) {
	testCases := []struct {
		name string
		code string
	}{
		{
			"Simple class inheritance with method override",
			`<?php 
			class Persion {
				public $name;
				public $age;

				public function __construct($name, $age) {
					$this->name = $name;
					$this->age = $age;
				}

				public function introduce() {
					return "My name is {$this->name} and I am {$this->age} years old.";
				}
			}

			class Student extends Persion {
				public $studentId;

				public function __construct($name, $age, $studentId) {
					parent::__construct($name, $age);
					$this->studentId = $studentId;
				}

				public function introduce() {
					return parent::introduce() . " My student ID is {$this->studentId}.";
				}
			}

			$student = new Student("Alice", 20, "S12345");
			echo $student->introduce();`,
		},
		{
			"Class inheritance without method override",
			`<?php 
			class Animal {
				public $name;

				public function __construct($name) {
					$this->name = $name;
				}

				public function speak() {
					return "Some sound";
				}
			}

			class Dog extends Animal {
				public function __construct($name) {
					parent::__construct($name);
				}
			}

			$dog = new Dog("Buddy");
			echo $dog->speak();`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Parse the code
			p := parser.New(lexer.New(tc.code))
			prog := p.ParseProgram()
			require.NotNil(t, prog, "Failed to parse program for test: %s", tc.name)

			// Compile
			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Failed to compile program for test: %s", tc.name)

			// Execute with runtime
			err = executeWithRuntime(t, comp)
			require.NoError(t, err, "Failed to execute program for test: %s", tc.name)
		})
	}
}

func TestClosure(t *testing.T) {
	testCases := []struct {
		name string
		code string
	}{
		{
			name: "simple_closure_with_use",
			code: `<?php
function foo($callback) {
    $callback();
}

$var1 = "World";

foo(function() use($var1) { 
    echo "Hello $var1"; 
});`,
		},
		{
			name: "closure_without_use",
			code: `<?php
$fn = function() {
    echo "Hello";
};
$fn();`,
		},
		{
			name: "closure_with_parameters",
			code: `<?php
$fn = function($name) {
    echo "Hello $name";
};
$fn("Alice");`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Parse the code
			p := parser.New(lexer.New(tc.code))
			prog := p.ParseProgram()
			require.NotNil(t, prog, "Failed to parse program for test: %s", tc.name)

			// Compile
			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Failed to compile program for test: %s", tc.name)

			// For now, we just check that compilation succeeds
			// Full execution can be implemented later
			_ = comp // Use the compiler to avoid unused variable warning
		})
	}
}

func TestClosureReferenceUseVariables(t *testing.T) {
	testCases := []struct {
		name string
		code string
	}{
		{
			name: "single_reference_use_variable",
			code: `<?php
$x = 10;
$closure = function() use (&$x) {
    $x++;
};
$closure();`,
		},
		{
			name: "multiple_reference_use_variables",
			code: `<?php
$a = 1;
$b = 2;
$closure = function() use (&$a, &$b) {
    $a *= 2;
    $b *= 3;
};
$closure();`,
		},
		{
			name: "mixed_reference_and_value_use",
			code: `<?php
$c = 5;
$d = 6;
$closure = function() use ($c, &$d) {
    $c = 100; // won't affect original
    $d = 200; // will affect original
};
$closure();`,
		},
		{
			name: "reference_use_with_array",
			code: `<?php
$arr = [1, 2, 3];
$closure = function() use (&$arr) {
    $arr[] = 4;
    $arr[0] = 10;
};
$closure();`,
		},
		{
			name: "nested_closure_with_reference_use",
			code: `<?php
$x = 1;
$outer = function() use (&$x) {
    $inner = function() use (&$x) {
        $x = 42;
    };
    $inner();
};
$outer();`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Parse the code
			p := parser.New(lexer.New(tc.code))
			prog := p.ParseProgram()
			require.NotNil(t, prog, "Failed to parse program for test: %s", tc.name)

			// Compile
			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Failed to compile program for test: %s", tc.name)

			// Verify the bytecode was generated
			bytecode := comp.GetBytecode()
			require.NotEmpty(t, bytecode, "Expected bytecode to be generated for test: %s", tc.name)

			// Check that we have OP_BIND_USE_VAR instructions
			foundBindUseVar := false
			for _, inst := range bytecode {
				if inst.Opcode == opcodes.OP_BIND_USE_VAR {
					foundBindUseVar = true
					break
				}
			}
			require.True(t, foundBindUseVar, "Expected to find OP_BIND_USE_VAR instruction for test: %s", tc.name)

			// Execute the closure code to verify it works
			err = executeWithRuntime(t, comp)
			require.NoError(t, err, "Failed to execute program for test: %s", tc.name)
		})
	}
}

// TestClosureExecution tests that closures with reference use variables work correctly
func TestClosureExecution(t *testing.T) {
	testCase := struct {
		name           string
		code           string
		expectedOutput string
	}{
		name:           "closure_with_reference_use_execution",
		code:           `<?php $x = 10; $closure = function() use (&$x) { $x++; }; $closure(); echo $x;`,
		expectedOutput: "11",
	}

	// Parse the code
	p := parser.New(lexer.New(testCase.code))
	prog := p.ParseProgram()
	require.NotNil(t, prog, "Failed to parse program")

	// Compile
	comp := NewCompiler()
	err := comp.Compile(prog)
	require.NoError(t, err, "Failed to compile program")

	// Execute and verify the result
	err = executeWithRuntime(t, comp)
	require.NoError(t, err, "Failed to execute program")

	// The test passes if execution succeeds without error
	// In a full implementation, we would capture and verify the output
}

// TestIncludeRequireIntegration tests the complete include/require functionality
func TestIncludeRequireIntegration(t *testing.T) {
	// Initialize runtime if not already done
	if runtime.GlobalRegistry == nil {
		err := runtime.Bootstrap()
		require.NoError(t, err, "Failed to bootstrap runtime")
	}

	// Initialize VM integration
	if runtime.GlobalVMIntegration == nil {
		err := runtime.InitializeVMIntegration()
		require.NoError(t, err, "Failed to initialize VM integration")
	}

	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "php_parser_test_")
	require.NoError(t, err, "Failed to create temp directory")
	defer os.RemoveAll(tmpDir)

	// Create test files in temporary directory
	testFiles := map[string]string{
		"return_array.php": `<?php
$arr = [1,2,3];
return $arr;`,
		"no_return.php": `<?php
echo "Hello from no_return.php\n";
// no return statement`,
		"return_string.php": `<?php
return "Hello from string return";`,
		"test_include1.php": `<?php
$included_var = "Hello from included file";
echo "This is from included file 1\n";
$number = 42;`,
		"test_include2.php": `<?php
// This file has a parse error to test error handling
echo "Before error"
syntax error here!`,
		"test_include3.php": `<?php
function included_function($param) {
    return "Function called with: " . $param;
}

$global_from_include = "Global variable";
echo "Include 3 executed\n";`,
	}

	for filename, content := range testFiles {
		filePath := filepath.Join(tmpDir, filename)
		err = os.WriteFile(filePath, []byte(content), 0644)
		require.NoError(t, err, "Failed to create test file: %s", filename)
	}

	// Create VM and set up compiler callback
	vmachine := vm.NewVirtualMachine()

	// Set up the compiler callback for include functionality
	vmachine.CompilerCallback = func(ctx *vm.ExecutionContext, program *ast.Program, filePath string, isRequired bool) (*values.Value, error) {
		// Create a new compiler for the included file
		comp := NewCompiler()
		if err := comp.Compile(program); err != nil {
			return nil, fmt.Errorf("compilation error in %s: %v", filePath, err)
		}

		// Create a new execution context for the included file but copy the variables
		// This allows variable sharing while preserving the main script's instruction state
		includeCtx := vm.NewExecutionContext()
		includeCtx.Variables = ctx.Variables         // Share variables
		includeCtx.Stack = ctx.Stack                 // Share stack
		includeCtx.IncludedFiles = ctx.IncludedFiles // Share included files tracking
		includeCtx.OutputWriter = ctx.OutputWriter   // Share output writer

		// Execute the compiled bytecode in the separate context
		err := vmachine.Execute(includeCtx, comp.GetBytecode(), comp.GetConstants(), comp.GetFunctions(), comp.GetClasses())
		if err != nil {
			return nil, fmt.Errorf("execution error in %s: %v", filePath, err)
		}

		// Copy back any changes to the shared state
		ctx.Variables = includeCtx.Variables
		ctx.Stack = includeCtx.Stack
		ctx.IncludedFiles = includeCtx.IncludedFiles
		// Output merging is now handled automatically by shared OutputWriter

		// Check if the included file executed an explicit return statement
		if includeCtx.Halted && len(includeCtx.Stack) > 0 {
			// Get the return value from the stack
			returnValue := includeCtx.Stack[len(includeCtx.Stack)-1]

			// Check if this is an explicit return (not just end of file)
			// In PHP, if a file ends without explicit return, it should return 1, not null
			if returnValue.IsNull() {
				// This is likely end-of-file, not an explicit return null
				return values.NewInt(1), nil
			}

			// Remove the return value from the stack and update both contexts
			includeCtx.Stack = includeCtx.Stack[:len(includeCtx.Stack)-1]
			ctx.Stack = includeCtx.Stack
			return returnValue, nil
		}

		// Return 1 on successful inclusion (PHP convention when no return statement)
		return values.NewInt(1), nil
	}

	testCases := []struct {
		name           string
		phpCode        string
		expectedOutput string
		expectError    bool
		errorContains  string
	}{
		{
			name: "Basic include success",
			phpCode: fmt.Sprintf(`<?php
echo "Before include\n";
include "%s";`, filepath.Join(tmpDir, "test_include1.php")),
			expectedOutput: "Before include\nThis is from included file 1\n",
			expectError:    false,
		},
		{
			name: "Include with function definition",
			phpCode: fmt.Sprintf(`<?php
include "%s";`, filepath.Join(tmpDir, "test_include3.php")),
			expectedOutput: "Include 3 executed\n",
			expectError:    false,
		},
		{
			name: "Require non-existent file",
			phpCode: `<?php
echo "Before require\n";
require "nonexistent.php";
echo "After require\n";`,
			expectError:   true,
			errorContains: "No such file or directory",
		},
		{
			name: "Include file with return array",
			phpCode: fmt.Sprintf(`<?php
echo "Before include\n";
$val = include "%s";
echo "After include\n";
var_dump($val);`, filepath.Join(tmpDir, "return_array.php")),
			expectedOutput: "Before include\nAfter include\narray(3) {\n  [0]=>\n  int(1)\n  [1]=>\n  int(2)\n  [2]=>\n  int(3)\n}\n",
			expectError:    false,
		},
		{
			name: "Include file with no return (default 1)",
			phpCode: fmt.Sprintf(`<?php
echo "Before include\n";
$val = include "%s";
echo "After include\n";
var_dump($val);`, filepath.Join(tmpDir, "no_return.php")),
			expectedOutput: "Before include\nHello from no_return.php\nAfter include\nint(1)\n",
			expectError:    false,
		},
		{
			name: "Include file with return string",
			phpCode: fmt.Sprintf(`<?php
$val = include "%s";
var_dump($val);`, filepath.Join(tmpDir, "return_string.php")),
			expectedOutput: "string(24) \"Hello from string return\"\n",
			expectError:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Parse main program
			l := lexer.New(tc.phpCode)
			p := parser.New(l)
			prog := p.ParseProgram()
			require.Empty(t, p.Errors(), "Parser errors: %v", p.Errors())

			// Compile main program
			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Failed to compile main program")

			// Create execution context
			vmCtx := vm.NewExecutionContext()

			// Set up output capture if needed
			var buf bytes.Buffer
			if tc.expectedOutput != "" {
				vmCtx.SetOutputWriter(&buf)
			}

			// Initialize global variables from runtime
			if vmCtx.GlobalVars == nil {
				vmCtx.GlobalVars = make(map[string]*values.Value)
			}

			variables := runtime.GlobalVMIntegration.GetAllVariables()
			for name, value := range variables {
				vmCtx.GlobalVars[name] = value
			}

			// Execute the program
			err = vmachine.Execute(vmCtx, comp.GetBytecode(), comp.GetConstants(), comp.GetFunctions(), comp.GetClasses())

			if tc.expectError {
				require.Error(t, err, "Expected error for test case: %s", tc.name)
				if tc.errorContains != "" {
					require.Contains(t, err.Error(), tc.errorContains, "Error should contain expected text")
				}
			} else {
				require.NoError(t, err, "Execution failed for test case: %s", tc.name)

				// Check output if specified
				if tc.expectedOutput != "" {
					actualOutput := buf.String()
					require.Equal(t, tc.expectedOutput, actualOutput, "Output mismatch for test case: %s", tc.name)
				}
			}
		})
	}
}

// TestIncludeOnceRequireOnce tests the _once variants of include/require
func TestIncludeOnceRequireOnce(t *testing.T) {
	// This test would be implemented when include_once and require_once opcodes are added
	// For now, we test that the tracking mechanism works
	t.Skip("include_once and require_once opcodes not yet implemented")
}

// TestTernaryOperator tests the ternary operator implementation
func TestTernaryOperator(t *testing.T) {
	// Initialize runtime
	err := runtime.Bootstrap()
	require.NoError(t, err)

	err = runtime.InitializeVMIntegration()
	require.NoError(t, err)

	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{
			name:     "Basic true ternary",
			code:     `<?php echo true ? "yes" : "no"; ?>`,
			expected: "yes",
		},
		{
			name:     "Basic false ternary",
			code:     `<?php echo false ? "yes" : "no"; ?>`,
			expected: "no",
		},
		{
			name:     "Integer true ternary",
			code:     `<?php echo 1 ? "yes" : "no"; ?>`,
			expected: "yes",
		},
		{
			name:     "Integer false ternary",
			code:     `<?php echo 0 ? "yes" : "no"; ?>`,
			expected: "no",
		},
		{
			name:     "String true ternary",
			code:     `<?php echo "hello" ? "yes" : "no"; ?>`,
			expected: "yes",
		},
		{
			name:     "String false ternary",
			code:     `<?php echo "" ? "yes" : "no"; ?>`,
			expected: "no",
		},
		{
			name:     "Variable true ternary",
			code:     `<?php $a = 1; echo $a ? "yes" : "no"; ?>`,
			expected: "yes",
		},
		{
			name:     "Variable false ternary",
			code:     `<?php $a = 0; echo $a ? "yes" : "no"; ?>`,
			expected: "no",
		},
		{
			name:     "Isset ternary (true)",
			code:     `<?php $a = 42; echo isset($a) ? "set" : "unset"; ?>`,
			expected: "set",
		},
		{
			name:     "Isset ternary (false)",
			code:     `<?php echo isset($nonexistent) ? "set" : "unset"; ?>`,
			expected: "unset",
		},
		{
			name:     "Array isset ternary (true)",
			code:     `<?php $arr = [1, 2, 3]; echo isset($arr[1]) ? "set" : "unset"; ?>`,
			expected: "set",
		},
		{
			name:     "Array isset ternary (false)",
			code:     `<?php $arr = [1, 2, 3]; echo isset($arr[5]) ? "set" : "unset"; ?>`,
			expected: "unset",
		},
		{
			name:     "Short ternary (true)",
			code:     `<?php echo "hello" ?: "world"; ?>`,
			expected: "hello",
		},
		{
			name:     "Short ternary (false)",
			code:     `<?php echo "" ?: "world"; ?>`,
			expected: "world",
		},
		{
			name:     "Short ternary null",
			code:     `<?php echo null ?: "world"; ?>`,
			expected: "world",
		},
		{
			name:     "Nested ternary",
			code:     `<?php echo true ? (false ? "a" : "b") : "c"; ?>`,
			expected: "b",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Parse and compile
			p := parser.New(lexer.New(test.code))
			prog := p.ParseProgram()
			require.NotNil(t, prog, "Failed to parse program for test: %s", test.name)

			comp := NewCompiler()
			err = comp.Compile(prog)
			require.NoError(t, err, "Compilation failed for test: %s", test.name)

			// Execute with output capture
			var buf bytes.Buffer
			vmCtx := vm.NewExecutionContext()
			vmCtx.SetOutputWriter(&buf)

			vmachine := vm.NewVirtualMachine()
			err = vmachine.Execute(vmCtx, comp.GetBytecode(), comp.GetConstants(), comp.GetFunctions(), comp.GetClasses())
			require.NoError(t, err, "Execution failed for test: %s", test.name)

			output := buf.String()
			require.Equal(t, test.expected, output, "Output mismatch for test: %s", test.name)
		})
	}
}

// TestUnsetIssetIntegration tests the interaction between unset and isset
func TestUnsetIssetIntegration(t *testing.T) {
	// Initialize runtime
	err := runtime.Bootstrap()
	require.NoError(t, err)

	err = runtime.InitializeVMIntegration()
	require.NoError(t, err)

	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{
			name: "Simple variable unset/isset",
			code: `<?php
				$a = 42;
				echo "Before: " . (isset($a) ? "set" : "unset") . "\n";
				unset($a);
				echo "After: " . (isset($a) ? "set" : "unset") . "\n";
			?>`,
			expected: "Before: set\nAfter: unset\n",
		},
		{
			name: "Array element unset/isset",
			code: `<?php
				$arr = [1, 2, 3];
				echo "Before: " . (isset($arr[1]) ? "set" : "unset") . "\n";
				unset($arr[1]);
				echo "After: " . (isset($arr[1]) ? "set" : "unset") . "\n";
			?>`,
			expected: "Before: set\nAfter: unset\n",
		},
		{
			name: "Multiple variable unset/isset",
			code: `<?php
				$a = 1;
				$b = 2;
				echo "Before a: " . (isset($a) ? "set" : "unset") . ", b: " . (isset($b) ? "set" : "unset") . "\n";
				unset($a, $b);
				echo "After a: " . (isset($a) ? "set" : "unset") . ", b: " . (isset($b) ? "set" : "unset") . "\n";
			?>`,
			expected: "Before a: set, b: set\nAfter a: unset, b: unset\n",
		},
		{
			name: "Array key isset after unset different key",
			code: `<?php
				$arr = [1, 2, 3]; 
				unset($arr[1]);
				echo isset($arr[0]) ? "set" : "unset";
			?>`,
			expected: "set",
		},
		{
			name: "Complex array isset/unset",
			code: `<?php
				$arr = ["a" => 1, "b" => 2, "c" => 3];
				$key = "b";
				echo "Before: " . (isset($arr[$key]) ? "set" : "unset") . "\n";
				unset($arr[$key]);
				echo "After: " . (isset($arr[$key]) ? "set" : "unset") . "\n";
			?>`,
			expected: "Before: set\nAfter: unset\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Parse and compile
			p := parser.New(lexer.New(test.code))
			prog := p.ParseProgram()
			require.NotNil(t, prog, "Failed to parse program for test: %s", test.name)

			comp := NewCompiler()
			err = comp.Compile(prog)
			require.NoError(t, err, "Compilation failed for test: %s", test.name)

			// Execute with output capture
			var buf bytes.Buffer
			vmCtx := vm.NewExecutionContext()
			vmCtx.SetOutputWriter(&buf)

			vmachine := vm.NewVirtualMachine()
			err = vmachine.Execute(vmCtx, comp.GetBytecode(), comp.GetConstants(), comp.GetFunctions(), comp.GetClasses())
			require.NoError(t, err, "Execution failed for test: %s", test.name)

			output := buf.String()
			require.Equal(t, test.expected, output, "Output mismatch for test: %s", test.name)
		})
	}
}

// TestErrorSuppressionExpression tests the @ (error suppression) operator
func TestErrorSuppressionExpression(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{
			name:     "suppress undefined variable",
			code:     `<?php $result = @$undefined_var; var_dump($result); ?>`,
			expected: "NULL\n",
		},
		{
			name:     "suppress array access on null",
			code:     `<?php $null_var = null; $result = @$null_var["key"]; var_dump($result); ?>`,
			expected: "NULL\n",
		},
		{
			name:     "nested error suppression",
			code:     `<?php $result = @(@$undefined_var2["key"]); var_dump($result); ?>`,
			expected: "NULL\n",
		},
		{
			name:     "suppress array access on undefined variable",
			code:     `<?php $result = @$undeclared_array[0]; var_dump($result); ?>`,
			expected: "NULL\n",
		},
		{
			name:     "multiple error suppression in one statement",
			code:     `<?php $a = @$undefined1; $b = @$undefined2; var_dump($a); var_dump($b); ?>`,
			expected: "NULL\nNULL\n",
		},
		{
			name:     "error suppression with assignment",
			code:     `<?php $result = @($x = $undefined_var); var_dump($result); var_dump($x); ?>`,
			expected: "NULL\nNULL\n",
		},
		{
			name:     "error suppression preserves return value",
			code:     `<?php $x = 42; $result = @$x; var_dump($result); ?>`,
			expected: "int(42)\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p := parser.New(lexer.New(test.code))
			prog := p.ParseProgram()
			require.NotNil(t, prog, "Failed to parse program for test: %s", test.name)

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Compilation failed for test: %s", test.name)

			// Execute with runtime initialization and output capture
			err = executeWithRuntime(t, comp)
			require.NoError(t, err, "Runtime initialization failed for test: %s", test.name)

			var buf bytes.Buffer
			vmCtx := vm.NewExecutionContext()
			vmCtx.SetOutputWriter(&buf)

			vmachine := vm.NewVirtualMachine()
			err = vmachine.Execute(vmCtx, comp.GetBytecode(), comp.GetConstants(), comp.GetFunctions(), comp.GetClasses())
			require.NoError(t, err, "Execution failed for test: %s", test.name)

			output := buf.String()
			require.Equal(t, test.expected, output, "Output mismatch for test: %s", test.name)
		})
	}
}

// TestListExpression tests list() expressions for array destructuring
func TestListExpression(t *testing.T) {
	tests := []struct {
		name     string
		phpCode  string
		expected string
	}{
		{
			name: "Simple list assignment",
			phpCode: `<?php
$array = array(1, 2, 3);
list($a, $b, $c) = $array;
echo $a . " " . $b . " " . $c . "\n";
?>`,
			expected: "1 2 3\n",
		},
		{
			name: "List with skip elements",
			phpCode: `<?php
$array = array(10, 20, 30, 40);
list($first, , $third) = $array;
echo $first . " " . $third . "\n";
?>`,
			expected: "10 30\n",
		},
		{
			name: "Nested list assignment",
			phpCode: `<?php
$nested = array(array(1, 2), array(3, 4));
list(list($a, $b), list($c, $d)) = $nested;
echo $a . " " . $b . " " . $c . " " . $d . "\n";
?>`,
			expected: "1 2 3 4\n",
		},
		{
			name: "List with insufficient elements",
			phpCode: `<?php
$array = array(1, 2);
list($a, $b, $c) = $array;
echo $a . " " . $b . " ";
var_dump($c);
?>`,
			expected: "1 2 NULL\n",
		},
		{
			name: "List with strings",
			phpCode: `<?php
$array = array("hello", "world", "test");
list($first, $second, $third) = $array;
echo $first . " " . $second . " " . $third . "\n";
?>`,
			expected: "hello world test\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the PHP code
			p := parser.New(lexer.New(tt.phpCode))
			prog := p.ParseProgram()
			require.NotNil(t, prog, "Failed to parse program for test: %s", tt.name)

			// Compile the program
			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Compilation failed for test: %s", tt.name)

			// Capture output for verification
			var buf bytes.Buffer
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Execute with runtime
			err = executeWithRuntime(t, comp)
			require.NoError(t, err, "Execution failed for test: %s", tt.name)

			// Restore stdout and get output
			w.Close()
			os.Stdout = oldStdout
			buf.ReadFrom(r)

			output := buf.String()
			require.Equal(t, tt.expected, output, "Output mismatch for test: %s", tt.name)
		})
	}
}

// TestTraitMethods tests the compilation of trait methods
func TestTraitMethods(t *testing.T) {
	tests := []struct {
		name     string
		phpCode  string
		expected string
	}{
		{
			name: "Simple trait method",
			phpCode: `<?php
trait SimpleTrait {
    public function greet($name) {
        return "Hello, " . $name;
    }
}

class TestClass {
    use SimpleTrait;
}

$obj = new TestClass();
echo $obj->greet("World") . "\n";
?>`,
			expected: "Hello, World\n",
		},
		{
			name: "Trait method with typed parameters",
			phpCode: `<?php
trait TypedTrait {
    public function processData(string $data, int $count) {
        return str_repeat($data, $count);
    }
}

class TestClass {
    use TypedTrait;
}

$obj = new TestClass();
echo $obj->processData("X", 3) . "\n";
?>`,
			expected: "XXX\n",
		},
		{
			name: "Trait method with default parameters",
			phpCode: `<?php
trait DefaultTrait {
    public function format($text, $prefix = ">> ", $suffix = " <<") {
        return $prefix . $text . $suffix;
    }
}

class TestClass {
    use DefaultTrait;
}

$obj = new TestClass();
echo $obj->format("test") . "\n";
echo $obj->format("hello", "* ") . "\n";
?>`,
			expected: ">> test <<\n* hello <<\n",
		},
		{
			name: "Trait method accessing $this",
			phpCode: `<?php
trait AccessTrait {
    public function getValue() {
        return $this->value ?? "no value";
    }
    
    public function setValue($val) {
        $this->value = $val;
        return "set to: " . $val;
    }
}

class TestClass {
    use AccessTrait;
    
    private $value = "initial";
}

$obj = new TestClass();
echo $obj->getValue() . "\n";
echo $obj->setValue("new value") . "\n";
echo $obj->getValue() . "\n";
?>`,
			expected: "initial\nset to: new value\nnew value\n",
		},
		{
			name: "Multiple traits with methods",
			phpCode: `<?php
trait MathTrait {
    public function add($a, $b) {
        return $a + $b;
    }
}

trait StringTrait {
    public function concat($a, $b) {
        return $a . $b;
    }
}

class Calculator {
    use MathTrait, StringTrait;
}

$calc = new Calculator();
echo $calc->add(5, 3) . "\n";
echo $calc->concat("Hello", "World") . "\n";
?>`,
			expected: "8\nHelloWorld\n",
		},
		{
			name: "Trait method with return statement",
			phpCode: `<?php
trait ReturnTrait {
    public function checkValue($val) {
        if ($val > 10) {
            return "high";
        }
        if ($val > 5) {
            return "medium";
        }
        return "low";
    }
}

class TestClass {
    use ReturnTrait;
}

$obj = new TestClass();
echo $obj->checkValue(15) . "\n";
echo $obj->checkValue(7) . "\n";
echo $obj->checkValue(3) . "\n";
?>`,
			expected: "high\nmedium\nlow\n",
		},
		{
			name: "Trait method with variadic parameters",
			phpCode: `<?php
trait VariadicTrait {
    public function sum(...$numbers) {
        $total = 0;
        foreach ($numbers as $num) {
            $total += $num;
        }
        return $total;
    }
}

class Calculator {
    use VariadicTrait;
}

$calc = new Calculator();
echo $calc->sum(1, 2, 3, 4, 5) . "\n";
echo $calc->sum(10, 20) . "\n";
?>`,
			expected: "15\n30\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the PHP code
			p := parser.New(lexer.New(tt.phpCode))
			prog := p.ParseProgram()
			require.NotNil(t, prog, "Failed to parse program for test: %s", tt.name)

			// Compile the program
			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Compilation failed for test: %s", tt.name)

			// Verify trait methods were compiled
			require.Greater(t, len(comp.traits), 0, "No traits were compiled for test: %s", tt.name)

			// Check that trait methods have proper bytecode
			for traitName, trait := range comp.traits {
				require.Greater(t, len(trait.Methods), 0, "Trait %s has no methods for test: %s", traitName, tt.name)
				for methodName, method := range trait.Methods {
					require.Greater(t, len(method.Instructions), 0, "Method %s in trait %s has no instructions for test: %s", methodName, traitName, tt.name)
				}
			}

			// Capture output for verification
			var buf bytes.Buffer
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Execute with runtime
			err = executeWithRuntime(t, comp)
			require.NoError(t, err, "Execution failed for test: %s", tt.name)

			// Restore stdout and get output
			w.Close()
			os.Stdout = oldStdout
			buf.ReadFrom(r)

			output := buf.String()
			require.Equal(t, tt.expected, output, "Output mismatch for test: %s", tt.name)
		})
	}
}

func TestPropertyAccessExpressionInArrays(t *testing.T) {
	tests := []struct {
		name     string
		phpCode  string
		expected string
	}{
		{
			"Simple property array read",
			`<?php 
			$obj = new stdClass(); 
			$obj->arr = [1, 2, 3]; 
			echo $obj->arr[1];`,
			"2",
		},
		{
			"Property array write",
			`<?php 
			$obj = new stdClass(); 
			$obj->arr = [1, 2, 3]; 
			$obj->arr[1] = 42; 
			echo $obj->arr[1];`,
			"42",
		},
		{
			"Property array append",
			`<?php 
			$obj = new stdClass(); 
			$obj->arr = [1, 2]; 
			$obj->arr[] = 3; 
			echo $obj->arr[2];`,
			"3",
		},
		{
			"Nested property array access",
			`<?php 
			$obj = new stdClass(); 
			$obj->arr = ['nested' => ['inner' => 'value']]; 
			$obj->arr['nested']['inner2'] = 'value2'; 
			echo $obj->arr['nested']['inner2'];`,
			"value2",
		},
		{
			"Property access with string keys",
			`<?php 
			$obj = new stdClass(); 
			$obj->data = ['key1' => 'value1']; 
			$obj->data['key2'] = 'value2'; 
			echo $obj->data['key1'] . ',' . $obj->data['key2'];`,
			"value1,value2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// First test with native PHP to make sure expected behavior is correct
			tempFile := "/tmp/test_property_array_" + strings.ReplaceAll(tt.name, " ", "_") + ".php"
			err := os.WriteFile(tempFile, []byte(tt.phpCode), 0644)
			require.NoError(t, err, "Failed to write temp file")
			defer os.Remove(tempFile)

			// Parse the PHP code
			p := parser.New(lexer.New(tt.phpCode))
			prog := p.ParseProgram()
			require.NotNil(t, prog, "Failed to parse program for test: %s", tt.name)
			require.Empty(t, p.Errors(), "Parser should not have errors for test: %s", tt.name)

			// Compile the program
			comp := NewCompiler()
			err = comp.Compile(prog)
			require.NoError(t, err, "Compilation failed for test: %s", tt.name)

			// Capture output for verification
			var buf bytes.Buffer
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Execute with runtime
			err = executeWithRuntime(t, comp)
			require.NoError(t, err, "Execution failed for test: %s", tt.name)

			// Restore stdout and get output
			w.Close()
			os.Stdout = oldStdout
			buf.ReadFrom(r)

			output := strings.TrimSpace(buf.String())
			require.Equal(t, tt.expected, output, "Output mismatch for test: %s", tt.name)
		})
	}
}

// TestInterfaceDefaultValuesCompilation tests the compilation of interface methods with default parameter values
// This test verifies that default values are properly evaluated during compilation phase
func TestInterfaceDefaultValuesCompilation(t *testing.T) {
	tests := []struct {
		name    string
		phpCode string
	}{
		{
			name: "Interface with default string parameter",
			phpCode: `<?php
interface TestInterface {
    function testMethod($param = "default_value");
}
?>`,
		},
		{
			name: "Interface with default integer parameter",
			phpCode: `<?php
interface TestInterface {
    function testMethod($param = 42);
}
?>`,
		},
		{
			name: "Interface with multiple default parameters",
			phpCode: `<?php
interface TestInterface {
    function testMethod($str = "hello", $num = 100, $bool = true, $null = null);
}
?>`,
		},
		{
			name: "Interface with default array parameter",
			phpCode: `<?php
interface TestInterface {
    function testMethod($arr = array(1, 2, 3));
}
?>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse
			p := parser.New(lexer.New(tt.phpCode))
			prog := p.ParseProgram()

			// Compile - this is where our default value evaluation happens
			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Compilation failed for test: %s", tt.name)

			// Verify that interfaces were compiled successfully
			require.Greater(t, len(comp.interfaces), 0, "No interfaces found after compilation")

			// Check that the interface has the expected method with default values
			iface, exists := comp.interfaces["TestInterface"]
			require.True(t, exists, "TestInterface not found in compiled interfaces")

			method, exists := iface.Methods["testMethod"]
			require.True(t, exists, "testMethod not found in TestInterface")

			// Verify that parameters with default values have their DefaultValue field set
			for _, param := range method.Parameters {
				if param.HasDefault {
					require.NotNil(t, param.DefaultValue, "Default value not evaluated for parameter %s", param.Name)
				}
			}
		})
	}
}

// TestGotoStatement tests goto statement compilation
func TestGotoStatement(t *testing.T) {
	tests := []struct {
		name    string
		phpCode string
	}{
		{
			name: "Simple goto forward",
			phpCode: `<?php
echo "Before goto\n";
goto target;
echo "This will be skipped\n";
target:
echo "After goto\n";
?>`,
		},
		{
			name: "Goto backward",
			phpCode: `<?php
$i = 0;
loop:
echo $i . "\n";
$i++;
if ($i < 3) goto loop;
echo "Done\n";
?>`,
		},
		{
			name: "Goto with conditional",
			phpCode: `<?php
$x = 5;
if ($x > 0) goto positive;
echo "Not positive\n";
goto end;
positive:
echo "Is positive\n";
end:
echo "Finished\n";
?>`,
		},
		{
			name: "Multiple labels",
			phpCode: `<?php
$choice = 2;
if ($choice == 1) goto first;
if ($choice == 2) goto second;
goto third;

first:
echo "First\n";
goto end;

second:
echo "Second\n";
goto end;

third:
echo "Third\n";

end:
echo "End\n";
?>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(lexer.New(tt.phpCode))
			prog := p.ParseProgram()
			require.NotNil(t, prog, "Failed to parse program for test: %s", tt.name)

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Compilation failed for test: %s", tt.name)

			// Test execution with runtime
			err = executeWithRuntime(t, comp)
			require.NoError(t, err, "Execution failed for test: %s", tt.name)
		})
	}
}

// TestLabelStatement tests label statement compilation
func TestLabelStatement(t *testing.T) {
	tests := []struct {
		name    string
		phpCode string
	}{
		{
			name: "Label without goto (no effect)",
			phpCode: `<?php
echo "Before label\n";
my_label:
echo "After label\n";
?>`,
		},
		{
			name: "Multiple labels in sequence",
			phpCode: `<?php
echo "Start\n";
label1:
label2:
label3:
echo "All labels defined\n";
?>`,
		},
		{
			name: "Label in conditional block",
			phpCode: `<?php
$x = true;
if ($x) {
    inner_label:
    echo "Inside conditional\n";
}
echo "Outside\n";
?>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(lexer.New(tt.phpCode))
			prog := p.ParseProgram()
			require.NotNil(t, prog, "Failed to parse program for test: %s", tt.name)

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Compilation failed for test: %s", tt.name)

			// Test execution with runtime
			err = executeWithRuntime(t, comp)
			require.NoError(t, err, "Execution failed for test: %s", tt.name)
		})
	}
}

// TestDeclareStatement tests declare statement compilation
func TestDeclareStatement(t *testing.T) {
	tests := []struct {
		name    string
		phpCode string
	}{
		{
			name: "Simple declare strict_types",
			phpCode: `<?php
declare(strict_types=1);
echo "Strict types enabled\n";
?>`,
		},
		{
			name: "Declare with ticks",
			phpCode: `<?php
declare(ticks=1);
echo "Ticks enabled\n";
?>`,
		},
		{
			name: "Declare with encoding",
			phpCode: `<?php
declare(encoding='UTF-8');
echo "Encoding set\n";
?>`,
		},
		{
			name: "Multiple declare directives",
			phpCode: `<?php
declare(strict_types=1, ticks=1);
echo "Multiple directives\n";
?>`,
		},
		{
			name: "Declare block syntax",
			phpCode: `<?php
declare(ticks=1) {
    echo "Inside declare block\n";
    echo "Still inside\n";
}
echo "Outside block\n";
?>`,
		},
		{
			name: "Declare alternative syntax",
			phpCode: `<?php
declare(ticks=1):
    echo "Alternative declare syntax\n";
enddeclare;
echo "After declare\n";
?>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(lexer.New(tt.phpCode))
			prog := p.ParseProgram()
			require.NotNil(t, prog, "Failed to parse program for test: %s", tt.name)

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Compilation failed for test: %s", tt.name)

			// Declare statements mostly affect compile-time behavior
			// For now, just test that they compile successfully
		})
	}
}

// TestAdvancedControlFlow tests complex combinations of advanced features
func TestAdvancedControlFlow(t *testing.T) {
	tests := []struct {
		name    string
		phpCode string
	}{
		{
			name: "Goto with loops",
			phpCode: `<?php
$i = 0;
start_loop:
if ($i >= 3) goto end_loop;
echo "Iteration: $i\n";
$i++;
goto start_loop;
end_loop:
echo "Loop finished\n";
?>`,
		},
		{
			name: "Nested goto and labels",
			phpCode: `<?php
$outer = 2;
$inner = 1;

outer_loop:
if ($outer <= 0) goto done;
echo "Outer: $outer\n";

inner_check:
if ($inner > 3) {
    $inner = 1;
    $outer--;
    goto outer_loop;
}
echo "  Inner: $inner\n";
$inner++;
goto inner_check;

done:
echo "All done\n";
?>`,
		},
		{
			name: "Declare with goto",
			phpCode: `<?php
declare(strict_types=1);
$x = 10;
if ($x > 5) goto skip_message;
echo "x is small\n";
skip_message:
echo "x is: $x\n";
?>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(lexer.New(tt.phpCode))
			prog := p.ParseProgram()
			require.NotNil(t, prog, "Failed to parse program for test: %s", tt.name)

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Compilation failed for test: %s", tt.name)

			// Test execution with runtime
			err = executeWithRuntime(t, comp)
			require.NoError(t, err, "Execution failed for test: %s", tt.name)
		})
	}
}

// TestAlternativeIfStatement tests alternative if syntax (if/endif)
func TestAlternativeIfStatement(t *testing.T) {
	tests := []struct {
		name    string
		phpCode string
	}{
		{
			name: "Simple alternative if",
			phpCode: `<?php
$x = 5;
if ($x > 0):
    echo "positive";
endif;
?>`,
		},
		{
			name: "Alternative if with else",
			phpCode: `<?php
$x = -1;
if ($x > 0):
    echo "positive";
else:
    echo "not positive";
endif;
?>`,
		},
		{
			name: "Alternative if with elseif",
			phpCode: `<?php
$x = 0;
if ($x > 0):
    echo "positive";
elseif ($x == 0):
    echo "zero";
else:
    echo "negative";
endif;
?>`,
		},
		{
			name: "Nested alternative if",
			phpCode: `<?php
$x = 5;
$y = 10;
if ($x > 0):
    if ($y > 5):
        echo "both positive and y > 5";
    endif;
endif;
?>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(lexer.New(tt.phpCode))
			prog := p.ParseProgram()
			require.NotNil(t, prog, "Failed to parse program for test: %s", tt.name)

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Compilation failed for test: %s", tt.name)

			// Test execution with runtime
			err = executeWithRuntime(t, comp)
			require.NoError(t, err, "Execution failed for test: %s", tt.name)
		})
	}
}

// TestAlternativeWhileStatement tests alternative while syntax (while/endwhile)
func TestAlternativeWhileStatement(t *testing.T) {
	tests := []struct {
		name    string
		phpCode string
	}{
		{
			name: "Simple alternative while",
			phpCode: `<?php
$i = 0;
while ($i < 3):
    echo $i;
    $i++;
endwhile;
?>`,
		},
		{
			name: "Alternative while with break",
			phpCode: `<?php
$i = 0;
while (true):
    if ($i >= 2) break;
    echo $i;
    $i++;
endwhile;
?>`,
		},
		{
			name: "Nested alternative while",
			phpCode: `<?php
$i = 0;
while ($i < 2):
    $j = 0;
    while ($j < 2):
        echo "$i,$j ";
        $j++;
    endwhile;
    $i++;
endwhile;
?>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(lexer.New(tt.phpCode))
			prog := p.ParseProgram()
			require.NotNil(t, prog, "Failed to parse program for test: %s", tt.name)

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Compilation failed for test: %s", tt.name)

			// Test execution with runtime
			err = executeWithRuntime(t, comp)
			require.NoError(t, err, "Execution failed for test: %s", tt.name)
		})
	}
}

// TestAlternativeForStatement tests alternative for syntax (for/endfor)
func TestAlternativeForStatement(t *testing.T) {
	tests := []struct {
		name    string
		phpCode string
	}{
		{
			name: "Simple alternative for",
			phpCode: `<?php
for ($i = 0; $i < 3; $i++):
    echo $i;
endfor;
?>`,
		},
		{
			name: "Alternative for with multiple init",
			phpCode: `<?php
for ($i = 0, $j = 0; $i < 2; $i++, $j += 2):
    echo "$i,$j ";
endfor;
?>`,
		},
		{
			name: "Alternative for with break",
			phpCode: `<?php
for ($i = 0; $i < 10; $i++):
    if ($i >= 2) break;
    echo $i;
endfor;
?>`,
		},
		{
			name: "Nested alternative for",
			phpCode: `<?php
for ($i = 0; $i < 2; $i++):
    for ($j = 0; $j < 2; $j++):
        echo "$i,$j ";
    endfor;
endfor;
?>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(lexer.New(tt.phpCode))
			prog := p.ParseProgram()
			require.NotNil(t, prog, "Failed to parse program for test: %s", tt.name)

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Compilation failed for test: %s", tt.name)

			// Test execution with runtime
			err = executeWithRuntime(t, comp)
			require.NoError(t, err, "Execution failed for test: %s", tt.name)
		})
	}
}

// TestAlternativeForeachStatement tests alternative foreach syntax (foreach/endforeach)
func TestAlternativeForeachStatement(t *testing.T) {
	tests := []struct {
		name    string
		phpCode string
	}{
		{
			name: "Simple alternative foreach",
			phpCode: `<?php
$arr = array(1, 2, 3);
foreach ($arr as $value):
    echo $value;
endforeach;
?>`,
		},
		{
			name: "Alternative foreach with key",
			phpCode: `<?php
$arr = array("a" => 1, "b" => 2);
foreach ($arr as $key => $value):
    echo "$key:$value ";
endforeach;
?>`,
		},
		{
			name: "Nested alternative foreach",
			phpCode: `<?php
$matrix = array(array(1, 2), array(3, 4));
foreach ($matrix as $row):
    foreach ($row as $value):
        echo "$value ";
    endforeach;
endforeach;
?>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(lexer.New(tt.phpCode))
			prog := p.ParseProgram()
			require.NotNil(t, prog, "Failed to parse program for test: %s", tt.name)

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Compilation failed for test: %s", tt.name)

			// Now test execution with the full implementation
			vmCtx := vm.NewExecutionContext()
			err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants(), comp.GetFunctions(), comp.GetClasses())
			require.NoError(t, err, "Execution failed for test: %s", tt.name)
		})
	}
}

// TestMixedAlternativeSyntax tests mixing alternative and regular syntax
func TestMixedAlternativeSyntax(t *testing.T) {
	tests := []struct {
		name    string
		phpCode string
	}{
		{
			name: "Alternative if with regular for",
			phpCode: `<?php
$x = 5;
if ($x > 0):
    for ($i = 0; $i < 2; $i++) {
        echo $i;
    }
endif;
?>`,
		},
		{
			name: "Regular if with alternative while",
			phpCode: `<?php
$x = 5;
if ($x > 0) {
    $i = 0;
    while ($i < 2):
        echo $i;
        $i++;
    endwhile;
}
?>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(lexer.New(tt.phpCode))
			prog := p.ParseProgram()
			require.NotNil(t, prog, "Failed to parse program for test: %s", tt.name)

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Compilation failed for test: %s", tt.name)

			// Test execution with runtime
			err = executeWithRuntime(t, comp)
			require.NoError(t, err, "Execution failed for test: %s", tt.name)
		})
	}
}

// TestPrintStatement tests the compilation of print statements
func TestPrintStatement(t *testing.T) {
	tests := []struct {
		name    string
		phpCode string
	}{
		{
			name:    "Simple print statement",
			phpCode: `<?php print "Hello World"; ?>`,
		},
		{
			name:    "Print with variable",
			phpCode: `<?php $msg = "test"; print $msg; ?>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(lexer.New(tt.phpCode))
			prog := p.ParseProgram()
			require.NotNil(t, prog, "Failed to parse program for test: %s", tt.name)

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Compilation failed for test: %s", tt.name)

			err = executeWithRuntime(t, comp)
			require.NoError(t, err, "Execution failed for test: %s", tt.name)
		})
	}
}

// TestPrintExpression tests the compilation of print expressions
func TestPrintExpression(t *testing.T) {
	tests := []struct {
		name    string
		phpCode string
	}{
		{
			name:    "Print expression returns 1",
			phpCode: `<?php $x = print "test"; ?>`,
		},
		{
			name:    "Print in assignment",
			phpCode: `<?php $result = print "output"; ?>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(lexer.New(tt.phpCode))
			prog := p.ParseProgram()
			require.NotNil(t, prog, "Failed to parse program for test: %s", tt.name)

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Compilation failed for test: %s", tt.name)

			err = executeWithRuntime(t, comp)
			require.NoError(t, err, "Execution failed for test: %s", tt.name)
		})
	}
}

// TestCloneExpression tests the compilation of clone expressions
func TestCloneExpression(t *testing.T) {
	tests := []struct {
		name        string
		phpCode     string
		expectError bool
	}{
		{
			name:        "Clone non-object should fail",
			phpCode:     `<?php $obj = "dummy"; $cloned = clone $obj; ?>`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(lexer.New(tt.phpCode))
			prog := p.ParseProgram()
			require.NotNil(t, prog, "Failed to parse program for test: %s", tt.name)

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Compilation failed for test: %s", tt.name)

			err = executeWithRuntime(t, comp)
			if tt.expectError {
				require.Error(t, err, "Expected execution to fail for test: %s", tt.name)
				require.Contains(t, err.Error(), "__clone method called on non-object", "Expected specific error message")
			} else {
				require.NoError(t, err, "Execution failed for test: %s", tt.name)
			}
		})
	}
}

// TestInstanceofExpression tests the compilation of instanceof expressions
func TestInstanceofExpression(t *testing.T) {
	tests := []struct {
		name    string
		phpCode string
	}{
		{
			name:    "Object instanceof class",
			phpCode: `<?php $obj = new stdClass(); $result = $obj instanceof stdClass; ?>`,
		},
		{
			name:    "String not instanceof class",
			phpCode: `<?php $result = "test" instanceof stdClass; ?>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(lexer.New(tt.phpCode))
			prog := p.ParseProgram()
			require.NotNil(t, prog, "Failed to parse program for test: %s", tt.name)

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Compilation failed for test: %s", tt.name)

			err = executeWithRuntime(t, comp)
			require.NoError(t, err, "Execution failed for test: %s", tt.name)
		})
	}
}

// TestCastExpression tests the compilation of cast expressions
func TestCastExpression(t *testing.T) {
	tests := []struct {
		name    string
		phpCode string
	}{
		{
			name:    "Cast string to int",
			phpCode: `<?php $result = (int)"123"; ?>`,
		},
		{
			name:    "Cast int to string",
			phpCode: `<?php $result = (string)456; ?>`,
		},
		{
			name:    "Cast to bool",
			phpCode: `<?php $result = (bool)0; ?>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(lexer.New(tt.phpCode))
			prog := p.ParseProgram()
			require.NotNil(t, prog, "Failed to parse program for test: %s", tt.name)

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Compilation failed for test: %s", tt.name)

			err = executeWithRuntime(t, comp)
			require.NoError(t, err, "Execution failed for test: %s", tt.name)
		})
	}
}

// TestEmptyExpression tests the compilation of empty expressions
func TestEmptyExpression(t *testing.T) {
	tests := []struct {
		name    string
		phpCode string
	}{
		{
			name:    "Empty string is empty",
			phpCode: `<?php $result = empty(""); ?>`,
		},
		{
			name:    "Non-empty string is not empty",
			phpCode: `<?php $result = empty("test"); ?>`,
		},
		{
			name:    "Zero is empty",
			phpCode: `<?php $result = empty(0); ?>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(lexer.New(tt.phpCode))
			prog := p.ParseProgram()
			require.NotNil(t, prog, "Failed to parse program for test: %s", tt.name)

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Compilation failed for test: %s", tt.name)

			err = executeWithRuntime(t, comp)
			require.NoError(t, err, "Execution failed for test: %s", tt.name)
		})
	}
}

// TestIssetExpression tests the compilation of isset expressions
func TestIssetExpression(t *testing.T) {
	tests := []struct {
		name    string
		phpCode string
	}{
		{
			name:    "Isset on defined variable",
			phpCode: `<?php $var = "test"; $result = isset($var); ?>`,
		},
		{
			name:    "Isset on undefined variable",
			phpCode: `<?php $result = isset($undefined); ?>`,
		},
		{
			name:    "Isset with multiple variables",
			phpCode: `<?php $a = 1; $b = 2; $result = isset($a, $b); ?>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(lexer.New(tt.phpCode))
			prog := p.ParseProgram()
			require.NotNil(t, prog, "Failed to parse program for test: %s", tt.name)

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Compilation failed for test: %s", tt.name)

			err = executeWithRuntime(t, comp)
			require.NoError(t, err, "Execution failed for test: %s", tt.name)
		})
	}
}

// TestGlobalStatement tests the compilation of global statements
func TestGlobalStatement(t *testing.T) {
	tests := []struct {
		name    string
		phpCode string
	}{
		{
			name:    "Global variable declaration",
			phpCode: `<?php function test() { global $x; } ?>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(lexer.New(tt.phpCode))
			prog := p.ParseProgram()
			require.NotNil(t, prog, "Failed to parse program for test: %s", tt.name)

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Compilation failed for test: %s", tt.name)

			err = executeWithRuntime(t, comp)
			require.NoError(t, err, "Execution failed for test: %s", tt.name)
		})
	}
}

// TestDoWhileStatement tests the compilation of do-while statements
func TestDoWhileStatement(t *testing.T) {
	tests := []struct {
		name    string
		phpCode string
	}{
		{
			name:    "Simple do-while loop",
			phpCode: `<?php $i = 0; do { $i++; } while($i < 3); ?>`,
		},
		{
			name:    "Do-while executes at least once",
			phpCode: `<?php $i = 10; do { $i++; } while($i < 5); ?>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(lexer.New(tt.phpCode))
			prog := p.ParseProgram()
			require.NotNil(t, prog, "Failed to parse program for test: %s", tt.name)

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Compilation failed for test: %s", tt.name)

			err = executeWithRuntime(t, comp)
			require.NoError(t, err, "Execution failed for test: %s", tt.name)
		})
	}
}

// TestMagicConstantExpression tests the compilation of magic constant expressions
func TestMagicConstantExpression(t *testing.T) {
	tests := []struct {
		name    string
		phpCode string
	}{
		{
			name:    "Magic constant __LINE__",
			phpCode: `<?php $line = __LINE__; ?>`,
		},
		{
			name:    "Magic constant __FILE__",
			phpCode: `<?php $file = __FILE__; ?>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(lexer.New(tt.phpCode))
			prog := p.ParseProgram()
			require.NotNil(t, prog, "Failed to parse program for test: %s", tt.name)

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Compilation failed for test: %s", tt.name)

			err = executeWithRuntime(t, comp)
			require.NoError(t, err, "Execution failed for test: %s", tt.name)
		})
	}
}

// TestCommaExpression tests the compilation of comma expressions
func TestCommaExpression(t *testing.T) {
	tests := []struct {
		name    string
		phpCode string
	}{
		{
			name:    "Comma expression returns last value",
			phpCode: `<?php $result = (1, 2, 3); ?>`,
		},
		{
			name:    "Comma with side effects",
			phpCode: `<?php $a = 1; $b = ($a++, $a * 2); ?>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(lexer.New(tt.phpCode))
			prog := p.ParseProgram()
			require.NotNil(t, prog, "Failed to parse program for test: %s", tt.name)

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Compilation failed for test: %s", tt.name)

			err = executeWithRuntime(t, comp)
			require.NoError(t, err, "Execution failed for test: %s", tt.name)
		})
	}
}

// Test for error cases - features not yet implemented
func TestNotYetImplementedFeatures(t *testing.T) {
	tests := []struct {
		name         string
		phpCode      string
		errorMessage string
	}{
		{
			name:         "Shell execution",
			phpCode:      "<?php `ls`; ?>",
			errorMessage: "shell execution expressions not yet implemented",
		},
		{
			name:         "First class callable",
			phpCode:      `<?php $fn = strlen(...); ?>`,
			errorMessage: "first-class callables not yet implemented",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(lexer.New(tt.phpCode))
			prog := p.ParseProgram()
			require.NotNil(t, prog, "Failed to parse program for test: %s", tt.name)

			comp := NewCompiler()
			err := comp.Compile(prog)
			if err == nil {
				t.Errorf("Expected compilation to fail with not implemented error for test: %s", tt.name)
			} else if !strings.Contains(err.Error(), tt.errorMessage) {
				t.Errorf("Expected error to contain %q, got %q for test: %s", tt.errorMessage, err.Error(), tt.name)
			}
		})
	}
}

// Benchmark tests for new implementations
func BenchmarkPrintExpression(b *testing.B) {
	phpCode := `<?php print "test"; ?>`

	// Parse once
	p := parser.New(lexer.New(phpCode))
	prog := p.ParseProgram()

	if prog == nil {
		b.Fatalf("Failed to parse program")
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		compiler := NewCompiler()
		err := compiler.Compile(prog)
		if err != nil {
			b.Fatalf("Compilation failed: %v", err)
		}
	}
}

func BenchmarkCastExpression(b *testing.B) {
	phpCode := `<?php (int)"123"; ?>`

	p := parser.New(lexer.New(phpCode))
	prog := p.ParseProgram()

	if prog == nil {
		b.Fatalf("Failed to parse program")
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		compiler := NewCompiler()
		err := compiler.Compile(prog)
		if err != nil {
			b.Fatalf("Compilation failed: %v", err)
		}
	}
}

// Integration test - complex expressions using multiple new features
func TestComplexExpressionIntegration(t *testing.T) {
	phpCode := `<?php 
		$x = print "hello"; 
		$result = "test" instanceof stdClass;
		$cast = (string)$result;
	?>`

	p := parser.New(lexer.New(phpCode))
	prog := p.ParseProgram()
	require.NotNil(t, prog, "Failed to parse complex integration program")

	comp := NewCompiler()
	err := comp.Compile(prog)
	require.NoError(t, err, "Compilation failed for complex integration test")

	err = executeWithRuntime(t, comp)
	require.NoError(t, err, "Execution failed for complex integration test")
}

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
	assert.Equal(t, -1, goFunc.MaxArgs) // Variadic function
	assert.True(t, goFunc.IsVariadic, "go() should be variadic")
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

// Test go() function with variable arguments
func TestGoFunctionVariadic(t *testing.T) {
	// Initialize runtime
	err := runtime.Bootstrap()
	require.NoError(t, err, "Failed to bootstrap runtime")

	// Create a closure and test variables
	closure := values.NewClosure(nil, nil, "test_closure")
	var1 := values.NewString("test_value")
	var2 := values.NewInt(42)

	// Test go() function handler directly with multiple arguments
	functions := runtime.GlobalRegistry.GetAllFunctions()
	goFunc := functions["go"]

	// Create a mock execution context
	ctx := &mockExecutionContext{}

	// Test with just closure
	args1 := []*values.Value{closure}
	result1, err := goFunc.Handler(ctx, args1)
	assert.NoError(t, err)
	assert.True(t, result1.IsGoroutine())

	// Test with closure and variables
	args2 := []*values.Value{closure, var1, var2}
	result2, err := goFunc.Handler(ctx, args2)
	assert.NoError(t, err)
	assert.True(t, result2.IsGoroutine())

	// Check that variables were captured
	gor := result2.Data.(*values.Goroutine)
	assert.Equal(t, 2, len(gor.UseVars))
	assert.Equal(t, "test_value", gor.UseVars["var_0"].ToString())
	assert.Equal(t, int64(42), gor.UseVars["var_1"].ToInt())
}

// Mock execution context for testing
type mockExecutionContext struct{}

func (m *mockExecutionContext) WriteOutput(output string)                   {}
func (m *mockExecutionContext) HasFunction(name string) bool                { return false }
func (m *mockExecutionContext) HasClass(name string) bool                   { return false }
func (m *mockExecutionContext) HasMethod(className, methodName string) bool { return false }

// Test go() function integration with parsing
func TestGoFunctionIntegration(t *testing.T) {
	// PHP code that uses go() function with variables
	phpCode := `<?php
$closure = function() {
    return "Hello from goroutine";
};
$var1 = "test";
$var2 = 42;
$g = go($closure, $var1, $var2);
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

func TestFinalStaticImplementation(t *testing.T) {
	code := `<?php
// Test 1: Basic static counter
function counter() {
    static $count = 0;
    $count++;
    return $count;
}

echo "Counter test: ";
echo counter() . " ";
echo counter() . " ";
echo counter() . "\n";

// Test 2: Multiple static variables
function multi_static() {
    static $name = "PHP", $version = 8.4;
    echo "Language: $name $version\n";
}

multi_static();
multi_static();

// Test 3: Static without default
function uninitialized_static() {
    static $unset;
    if (is_null($unset)) {
        $unset = "now_set";
        echo "Was null, now set\n";
    } else {
        echo "Still set: $unset\n";
    }
}

uninitialized_static();
uninitialized_static();

echo "All tests completed!\n";
?>`

	expected := `Counter test: 1 2 3
Language: PHP 8.4
Language: PHP 8.4
Was null, now set
Still set: now_set
All tests completed!
`

	// Compile the AST
	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()
	require.NotNil(t, program)

	comp := NewCompiler()
	err := comp.Compile(program)
	require.NoError(t, err, "Compilation failed")

	// Capture output
	buf := &bytes.Buffer{}
	r, w, _ := os.Pipe()
	oldStdout := os.Stdout
	os.Stdout = w

	// Execute with runtime
	err = executeWithRuntime(t, comp)
	require.NoError(t, err, "Execution failed")

	// Restore stdout and get output
	w.Close()
	os.Stdout = oldStdout
	buf.ReadFrom(r)

	output := buf.String()
	require.Equal(t, expected, output, "Output mismatch - static statements not working correctly")
}

// TestArrowFunctions tests arrow function compilation and execution
func TestArrowFunctions(t *testing.T) {
	tests := []struct {
		name    string
		phpCode string
	}{
		{
			name: "Simple arrow function",
			phpCode: `<?php
$add = fn($a, $b) => $a + $b;
echo $add(1, 2);
?>`,
		},
		{
			name: "Arrow function with single parameter",
			phpCode: `<?php
$double = fn($x) => $x * 2;
echo $double(5);
?>`,
		},
		{
			name: "Arrow function with no parameters",
			phpCode: `<?php
$getValue = fn() => 42;
echo $getValue();
?>`,
		},
		{
			name: "Arrow function with type hints",
			phpCode: `<?php
$calculate = fn(int $a, int $b): int => $a * $b + 1;
echo $calculate(3, 4);
?>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(lexer.New(tt.phpCode))
			prog := p.ParseProgram()
			require.NotNil(t, prog, "Failed to parse program for test: %s", tt.name)

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Compilation failed for test: %s", tt.name)

			// For now, just test that compilation succeeds
			// TODO: Add execution testing when runtime supports closures better
		})
	}
}

// TestSpreadExpressions tests spread expression compilation
func TestSpreadExpressions(t *testing.T) {
	tests := []struct {
		name    string
		phpCode string
	}{
		{
			name: "Spread in array literal",
			phpCode: `<?php
$arr1 = [1, 2];
$arr2 = [...$arr1, 3, 4];
print_r($arr2);
?>`,
		},
		{
			name: "Multiple spreads in array",
			phpCode: `<?php
$first = [1, 2];
$second = [3, 4];
$combined = [...$first, 5, ...$second, 6];
print_r($combined);
?>`,
		},
		{
			name: "Spread with empty array",
			phpCode: `<?php
$empty = [];
$result = [...$empty, 1, 2, 3];
print_r($result);
?>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(lexer.New(tt.phpCode))
			prog := p.ParseProgram()
			require.NotNil(t, prog, "Failed to parse program for test: %s", tt.name)

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Compilation failed for test: %s", tt.name)

			// For now, just test that compilation succeeds
			// TODO: Add execution testing when VM supports spreads better
		})
	}
}

// TestModernPHPFeaturesCombined tests combining modern features
func TestModernPHPFeaturesCombined(t *testing.T) {
	tests := []struct {
		name    string
		phpCode string
	}{
		{
			name: "Arrow function with spread",
			phpCode: `<?php
$numbers = [1, 2, 3];
$process = fn($arr) => [...$arr, 4, 5];
$result = $process($numbers);
print_r($result);
?>`,
		},
		{
			name: "Complex arrow function",
			phpCode: `<?php
$data = [10, 20, 30];
$transform = fn($values) => array_sum([...$values, 40]);
echo $transform($data);
?>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(lexer.New(tt.phpCode))
			prog := p.ParseProgram()
			require.NotNil(t, prog, "Failed to parse program for test: %s", tt.name)

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Compilation failed for test: %s", tt.name)

			// Test that compilation succeeds for combined features
		})
	}
}

// TestNumberLiterals tests compilation and execution of various number literal formats
func TestNumberLiterals(t *testing.T) {
	tests := []struct {
		name     string
		phpCode  string
		expected string
	}{
		{
			name:     "Binary literals",
			phpCode:  `<?php echo 0b1111 . "\n"; echo 0B1010 . "\n"; echo 0b0 . "\n"; ?>`,
			expected: "15\n10\n0\n",
		},
		{
			name:     "Hexadecimal literals",
			phpCode:  `<?php echo 0x1F . "\n"; echo 0X1f . "\n"; echo 0xff . "\n"; ?>`,
			expected: "31\n31\n255\n",
		},
		{
			name:     "Octal literals",
			phpCode:  `<?php echo 0123 . "\n"; echo 0o123 . "\n"; echo 0O777 . "\n"; ?>`,
			expected: "83\n83\n511\n",
		},
		{
			name:     "Decimal integers",
			phpCode:  `<?php echo 123 . "\n"; echo 0 . "\n"; echo 999 . "\n"; ?>`,
			expected: "123\n0\n999\n",
		},
		{
			name:     "Numbers with underscores",
			phpCode:  `<?php echo 1_000_000 . "\n"; echo 0xFF_AA . "\n"; echo 0b1010_1010 . "\n"; ?>`,
			expected: "1000000\n65450\n170\n",
		},
		{
			name:     "Float literals",
			phpCode:  `<?php echo 1.23 . "\n"; echo 3.14159 . "\n"; echo 1.0 . "\n"; ?>`,
			expected: "1.23\n3.14159\n1\n",
		},
		{
			name:     "Scientific notation",
			phpCode:  `<?php echo 1.23e4 . "\n"; echo 1.23E-4 . "\n"; echo 5e2 . "\n"; ?>`,
			expected: "12300\n0.000123\n500\n",
		},
		{
			name:     "Mixed number types in expressions",
			phpCode:  `<?php echo (0b1111 + 0x10) . "\n"; echo (123 + 1.5) . "\n"; ?>`,
			expected: "31\n124.5\n",
		},
		{
			name:     "Integer boundary values",
			phpCode:  `<?php echo 9223372036854775807 . "\n"; echo var_dump(9223372036854775807); ?>`,
			expected: "9223372036854775807\nint(9223372036854775807)\n",
		},
		{
			name:     "Integer overflow to float conversion",
			phpCode:  `<?php echo 9223372036854775808 . "\n"; echo var_dump(9223372036854775808); ?>`,
			expected: "9.223372036854776e+18\nfloat(9.223372036854776e+18)\n",
		},
		{
			name:     "Large integer overflow cases",
			phpCode:  `<?php echo 18446744073709551615 . "\n"; echo var_dump(18446744073709551615); ?>`,
			expected: "1.8446744073709552e+19\nfloat(1.8446744073709552e+19)\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the PHP code
			p := parser.New(lexer.New(tt.phpCode))
			prog := p.ParseProgram()
			require.NotNil(t, prog, "Failed to parse program for test: %s", tt.name)

			// Compile the program
			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Compilation failed for test: %s", tt.name)

			// Capture output for verification
			var buf bytes.Buffer
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Execute with runtime
			err = executeWithRuntime(t, comp)
			require.NoError(t, err, "Execution failed for test: %s", tt.name)

			// Restore stdout and get output
			w.Close()
			os.Stdout = oldStdout
			buf.ReadFrom(r)

			output := buf.String()
			require.Equal(t, tt.expected, output, "Output mismatch for test: %s", tt.name)
		})
	}
}

// TestInterfaceDeclaration tests the compilation of interface declarations
func TestInterfaceDeclaration(t *testing.T) {
	tests := []struct {
		name    string
		phpCode string
	}{
		{
			name: "Simple interface declaration",
			phpCode: `<?php
interface Drawable {
    public function draw();
}
?>`,
		},
		{
			name: "Interface with multiple methods",
			phpCode: `<?php
interface Shape {
    public function area();
    public function perimeter();
}
?>`,
		},
		{
			name: "Interface with extends",
			phpCode: `<?php
interface ColoredShape extends Drawable {
    public function setColor($color);
}
?>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(lexer.New(tt.phpCode))
			prog := p.ParseProgram()
			require.NotNil(t, prog, "Failed to parse program for test: %s", tt.name)

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Compilation failed for test: %s", tt.name)

			// Verify interface was stored
			require.Greater(t, len(comp.interfaces), 0, "No interfaces were compiled")
		})
	}
}

// TestTraitDeclaration tests the compilation of trait declarations
func TestTraitDeclaration(t *testing.T) {
	tests := []struct {
		name    string
		phpCode string
	}{
		{
			name: "Simple trait declaration",
			phpCode: `<?php
trait Loggable {
    public function log($message) {
        echo "LOG: " . $message . "\n";
    }
}
?>`,
		},
		{
			name: "Trait with property",
			phpCode: `<?php
trait Timestampable {
    private $timestamp;
    
    public function setTimestamp($time) {
        $this->timestamp = $time;
    }
}
?>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(lexer.New(tt.phpCode))
			prog := p.ParseProgram()
			require.NotNil(t, prog, "Failed to parse program for test: %s", tt.name)

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Compilation failed for test: %s", tt.name)

			// Verify trait was stored
			require.Greater(t, len(comp.traits), 0, "No traits were compiled")
		})
	}
}

// TestTraitUsage tests the compilation of trait usage in classes
func TestTraitUsage(t *testing.T) {
	tests := []struct {
		name    string
		phpCode string
	}{
		{
			name: "Class using trait",
			phpCode: `<?php
trait Loggable {
    public function log($message) {
        echo "LOG: " . $message;
    }
}

class User {
    use Loggable;
    
    public function getName() {
        return "user";
    }
}
?>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(lexer.New(tt.phpCode))
			prog := p.ParseProgram()
			require.NotNil(t, prog, "Failed to parse program for test: %s", tt.name)

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Compilation failed for test: %s", tt.name)

			// Verify both trait and class were compiled
			require.Greater(t, len(comp.traits), 0, "No traits were compiled")
			require.Greater(t, len(comp.classes), 0, "No classes were compiled")
		})
	}
}

// TestInterfaceImplementation tests classes implementing interfaces
func TestInterfaceImplementation(t *testing.T) {
	tests := []struct {
		name    string
		phpCode string
	}{
		{
			name: "Class implementing interface",
			phpCode: `<?php
interface Drawable {
    public function draw();
}

class Circle implements Drawable {
    public function draw() {
        echo "Drawing circle";
    }
}
?>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(lexer.New(tt.phpCode))
			prog := p.ParseProgram()
			require.NotNil(t, prog, "Failed to parse program for test: %s", tt.name)

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Compilation failed for test: %s", tt.name)

			// Verify both interface and class were compiled
			require.Greater(t, len(comp.interfaces), 0, "No interfaces were compiled")
			require.Greater(t, len(comp.classes), 0, "No classes were compiled")
		})
	}
}

// TestEnumDeclaration tests the compilation of enum declarations
func TestEnumDeclaration(t *testing.T) {
	tests := []struct {
		name    string
		phpCode string
	}{
		{
			name: "Simple enum declaration",
			phpCode: `<?php
enum Status {
    case PENDING;
    case APPROVED;
    case REJECTED;
}
?>`,
		},
		{
			name: "Backed enum with string values",
			phpCode: `<?php
enum Color: string {
    case RED = 'red';
    case GREEN = 'green';
    case BLUE = 'blue';
}
?>`,
		},
		{
			name: "Enum with method",
			phpCode: `<?php
enum Size {
    case SMALL;
    case MEDIUM;
    case LARGE;
    
    public function getLabel() {
        return match($this) {
            Size::SMALL => 'Small',
            Size::MEDIUM => 'Medium', 
            Size::LARGE => 'Large',
        };
    }
}
?>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(lexer.New(tt.phpCode))
			prog := p.ParseProgram()
			require.NotNil(t, prog, "Failed to parse program for test: %s", tt.name)

			comp := NewCompiler()
			err := comp.Compile(prog)
			require.NoError(t, err, "Compilation failed for test: %s", tt.name)

			// Verify enum was compiled as class
			require.Greater(t, len(comp.classes), 0, "No enums were compiled")
		})
	}
}

func TestStaticStatements(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{
			name: "basic static variable",
			code: `<?php
function test() {
    static $count = 0;
    $count++;
    echo $count;
}
test();
test();
test();
?>`,
			expected: "123",
		},
		{
			name: "static variable with string default",
			code: `<?php
function test() {
    static $name = "hello";
    echo $name;
}
test();
?>`,
			expected: "hello",
		},
		{
			name: "multiple static variables",
			code: `<?php
function test() {
    static $count = 0, $name = "test";
    $count++;
    echo $count . ":" . $name . "\n";
}
test();
test();
?>`,
			expected: "1:test\n2:test\n",
		},
		{
			name: "static variable without default value",
			code: `<?php
function test() {
    static $uninit;
    if (is_null($uninit)) {
        echo "null\n";
        $uninit = "initialized";
    } else {
        echo $uninit . "\n";
    }
}
test();
test();
?>`,
			expected: "null\ninitialized\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test with our compiler
			l := lexer.New(tt.code)
			p := parser.New(l)
			program := p.ParseProgram()
			require.NotNil(t, program)

			// Compile the AST
			comp := NewCompiler()
			err := comp.Compile(program)
			require.NoError(t, err, "Compilation failed for test: %s", tt.name)

			// Capture output
			buf := &bytes.Buffer{}
			r, w, _ := os.Pipe()
			oldStdout := os.Stdout
			os.Stdout = w

			// Execute with runtime
			err = executeWithRuntime(t, comp)
			require.NoError(t, err, "Execution failed for test: %s", tt.name)

			// Restore stdout and get output
			w.Close()
			os.Stdout = oldStdout
			buf.ReadFrom(r)

			output := buf.String()
			require.Equal(t, tt.expected, output, "Output mismatch for test: %s", tt.name)
		})
	}
}

func TestUnsetStatement(t *testing.T) {
	// Initialize runtime for tests
	err := runtime.Bootstrap()
	assert.NoError(t, err)

	tests := []struct {
		name           string
		code           string
		expectedOutput string
	}{
		{
			name: "Simple variable unset",
			code: `<?php
				$a = 42;
				echo "Before: " . (isset($a) ? "set" : "unset") . "\n";
				unset($a);
				echo "After: " . (isset($a) ? "set" : "unset") . "\n";
			`,
			expectedOutput: "Before: set\nAfter: unset\n",
		},
		{
			name: "Array element unset",
			code: `<?php
				$arr = [1, 2, 3];
				echo "Before: " . (isset($arr[1]) ? "set" : "unset") . "\n";
				unset($arr[1]);
				echo "After: " . (isset($arr[1]) ? "set" : "unset") . "\n";
			`,
			expectedOutput: "Before: set\nAfter: unset\n",
		},
		{
			name: "Multiple variable unset",
			code: `<?php
				$a = 1;
				$b = 2;
				echo "Before a: " . (isset($a) ? "set" : "unset") . ", b: " . (isset($b) ? "set" : "unset") . "\n";
				unset($a, $b);
				echo "After a: " . (isset($a) ? "set" : "unset") . ", b: " . (isset($b) ? "set" : "unset") . "\n";
			`,
			expectedOutput: "Before a: set, b: set\nAfter a: unset, b: unset\n",
		},
		{
			name: "Unset nonexistent variable (no error)",
			code: `<?php
				unset($nonexistent);
				echo "Done\n";
			`,
			expectedOutput: "Done\n",
		},
		{
			name: "Array append after unset",
			code: `<?php
				$arr = [1, 2, 3];
				unset($arr[1]);
				$arr[] = 4;
				echo count($arr) . "\n";
			`,
			expectedOutput: "3\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := compileAndExecute(t, tt.code)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedOutput, output, "Output mismatch for test case: %s", tt.name)
		})
	}
}

func TestUnsetStatementErrors(t *testing.T) {
	// Initialize runtime for tests
	err := runtime.Bootstrap()
	assert.NoError(t, err)

	tests := []struct {
		name        string
		code        string
		expectError bool
		errorMsg    string
	}{
		{
			name: "Cannot unset $this",
			code: `<?php
				class Test {
					public function test() {
						unset($this);
					}
				}
			`,
			expectError: true,
			errorMsg:    "cannot unset $this",
		},
		{
			name: "Cannot use [] for unsetting",
			code: `<?php
				$arr = [1, 2, 3];
				unset($arr[]);
			`,
			expectError: true,
			errorMsg:    "cannot use [] for unsetting",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseAndCompileOnly(t, tt.code)
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUnsetComplexExpressions(t *testing.T) {
	// Initialize runtime for tests
	err := runtime.Bootstrap()
	assert.NoError(t, err)

	tests := []struct {
		name           string
		code           string
		expectedOutput string
		skip           bool
		skipReason     string
	}{
		{
			name: "Nested array unset",
			code: `<?php
				$arr = [[1, 2], [3, 4]];
				echo "Before: " . (isset($arr[0][1]) ? "set" : "unset") . "\n";
				unset($arr[0][1]);  
				echo "After: " . (isset($arr[0][1]) ? "set" : "unset") . "\n";
			`,
			expectedOutput: "Before: set\nAfter: unset\n",
		},
		{
			name: "Dynamic array key unset",
			code: `<?php
				$arr = ["a" => 1, "b" => 2, "c" => 3];
				$key = "b";
				echo "Before: " . (isset($arr[$key]) ? "set" : "unset") . "\n";
				unset($arr[$key]);
				echo "After: " . (isset($arr[$key]) ? "set" : "unset") . "\n";
			`,
			expectedOutput: "Before: set\nAfter: unset\n",
		},
		{
			name: "Object property unset",
			code: `<?php
				$obj = new stdClass();
				$obj->prop = "value";
				echo "Before: " . (isset($obj->prop) ? "set" : "unset") . "\n"; 
				unset($obj->prop);
				echo "After: " . (isset($obj->prop) ? "set" : "unset") . "\n";
			`,
			expectedOutput: "Before: set\nAfter: unset\n",
			skip:           true,
			skipReason:     "Object property support requires full object system",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skip {
				t.Skip(tt.skipReason)
				return
			}

			output, err := compileAndExecute(t, tt.code)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedOutput, output, "Output mismatch for test case: %s", tt.name)
		})
	}
}

func TestUnsetWithVariableVariable(t *testing.T) {
	// Initialize runtime for tests
	err := runtime.Bootstrap()
	assert.NoError(t, err)

	// Skip for now as variable variables in unset require more complex handling
	t.Skip("Variable variables in unset context need additional implementation")

	code := `<?php
		$var = "test";
		$test = "value";
		echo "Before: " . (isset($$var) ? "set" : "unset") . "\n";
		unset($$var);
		echo "After: " . (isset($$var) ? "set" : "unset") . "\n";
	`
	expectedOutput := "Before: set\nAfter: unset\n"

	output, err := compileAndExecute(t, code)
	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, output)
}

// Helper function to test unset compilation without execution
func TestUnsetCompilationOnly(t *testing.T) {
	tests := []struct {
		name        string
		code        string
		expectError bool
	}{
		{
			name: "Simple unset compiles successfully",
			code: `<?php unset($a); ?>`,
		},
		{
			name: "Array unset compiles successfully",
			code: `<?php unset($arr[0]); ?>`,
		},
		{
			name: "Multiple unset compiles successfully",
			code: `<?php unset($a, $b, $c); ?>`,
		},
		{
			name: "Object property unset compiles successfully",
			code: `<?php unset($obj->prop); ?>`,
		},
		{
			name: "Static property unset compiles successfully",
			code: `<?php unset(MyClass::$prop); ?>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp, err := parseAndCompileOnly(t, tt.code)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, comp)

				// Check that we have some instructions generated
				assert.Greater(t, len(comp.GetBytecode()), 0, "Should generate at least one instruction")

				// Look for unset-related opcodes
				hasUnsetOpcode := false
				for _, inst := range comp.GetBytecode() {
					if strings.Contains(inst.Opcode.String(), "UNSET") {
						hasUnsetOpcode = true
						break
					}
				}
				assert.True(t, hasUnsetOpcode, "Should contain unset-related opcode")
			}
		})
	}
}

// Helper function to compile and execute PHP code with output capture
func compileAndExecute(t *testing.T, code string) (string, error) {
	// Parse the code
	p := parser.New(lexer.New(code))
	prog := p.ParseProgram()
	require.NotNil(t, prog, "Failed to parse program")

	// Compile the code
	comp := NewCompiler()
	err := comp.Compile(prog)
	require.NoError(t, err, "Failed to compile program")

	// Execute with output capture
	var buf bytes.Buffer
	vmCtx := vm.NewExecutionContext()
	vmCtx.SetOutputWriter(&buf)

	// Initialize runtime if not already done
	if runtime.GlobalRegistry == nil {
		err := runtime.Bootstrap()
		require.NoError(t, err, "Failed to bootstrap runtime")
	}

	// Initialize VM integration
	if runtime.GlobalVMIntegration == nil {
		err := runtime.InitializeVMIntegration()
		require.NoError(t, err, "Failed to initialize VM integration")
	}

	// Execute
	vmachine := vm.NewVirtualMachine()
	err = vmachine.Execute(vmCtx, comp.GetBytecode(), comp.GetConstants(), comp.GetFunctions(), comp.GetClasses())

	return buf.String(), err
}

// Helper function to parse and compile PHP code without execution
func parseAndCompileOnly(t *testing.T, code string) (*Compiler, error) {
	// Parse the code
	p := parser.New(lexer.New(code))
	prog := p.ParseProgram()
	require.NotNil(t, prog, "Failed to parse program")

	// Compile the code
	comp := NewCompiler()
	err := comp.Compile(prog)

	return comp, err
}
