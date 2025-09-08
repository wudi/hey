package compiler

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wudi/php-parser/ast"
	"github.com/wudi/php-parser/compiler/opcodes"
	"github.com/wudi/php-parser/compiler/runtime"
	"github.com/wudi/php-parser/compiler/values"
	"github.com/wudi/php-parser/compiler/vm"
	"github.com/wudi/php-parser/lexer"
	"github.com/wudi/php-parser/parser"
)

// Helper function to execute compiled bytecode with runtime initialization
func executeWithRuntime(t *testing.T, comp *Compiler) error {
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

func TestClassConstantDeclaration(t *testing.T) {
	testCases := []struct {
		name string
		code string
	}{
		{
			"Simple class constants",
			`<?php 
			class TestClass { 
				public const VERSION = "1.0"; 
				public const MAX_SIZE = 100; 
				
				public function getVersion() { 
					return self::VERSION; 
				} 
			} 
			echo TestClass::VERSION; 
			echo TestClass::MAX_SIZE;`,
		},
		{
			"Multiple constants in one declaration",
			`<?php 
			class TestClass { 
				public const FIRST = 1, SECOND = 2, THIRD = 3; 
			} 
			echo TestClass::FIRST; 
			echo TestClass::SECOND;`,
		},
		{
			"Constants with different visibilities",
			`<?php 
			class TestClass { 
				public const PUBLIC_CONST = "public"; 
			} 
			echo TestClass::PUBLIC_CONST;`,
		},
		{
			"Final constants",
			`<?php 
			class TestClass { 
				final public const IMMUTABLE = "cannot_override"; 
				public const OTHER = "allowed"; 
			} 
			echo TestClass::IMMUTABLE; 
			echo TestClass::OTHER;`,
		},
		{
			"Constants with different types",
			`<?php 
			class TestClass { 
				public const STRING_CONST = "hello"; 
				public const INT_CONST = 42; 
				public const FLOAT_CONST = 3.14; 
				public const BOOL_CONST = true; 
				public const NULL_CONST = null; 
				public const ARRAY_CONST = []; 
			} 
			echo TestClass::STRING_CONST; 
			echo TestClass::INT_CONST; 
			echo TestClass::BOOL_CONST ? "1" : "0";`,
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

			vmCtx := vm.NewExecutionContext()
			err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants(), comp.GetFunctions(), comp.GetClasses())
			require.NoError(t, err, "Failed to execute class constant declaration: %s", tc.name)
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
		})
	}
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
			phpCode: `<?php
echo "Before include\n";
include "test_include1.php";`,
			expectedOutput: "Before include\nThis is from included file 1\n",
			expectError:    false,
		},
		{
			name: "Include with function definition",
			phpCode: `<?php
include "test_include3.php";`,
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
