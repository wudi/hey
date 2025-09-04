package compiler

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wudi/php-parser/compiler/vm"
	"github.com/wudi/php-parser/lexer"
	"github.com/wudi/php-parser/parser"
)

func TestEcho(t *testing.T) {
	p := parser.New(lexer.New(`<?php echo "Hello, World!";`))
	prog := p.ParseProgram()

	comp := NewCompiler()
	err := comp.Compile(prog)
	require.NoError(t, err)

	vmCtx := vm.NewExecutionContext()
	err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants())
	require.NoError(t, err)
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
			err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants())
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
			err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants())
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
			err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants())
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
			err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants())
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
			err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants())
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
			err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants())
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
			err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants())
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
			err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants())
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
			err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants())
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
			err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants())
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
			err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants())
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
			err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants())
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
			err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants())
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
	err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants())
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
	err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants())
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
	err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants())
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
	err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants())
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
	err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants())
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
			err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants())
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
			err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants())
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
	err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants())
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
			err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants())
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
			err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants())
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
			err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants())
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
			
			err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants())
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
			err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants())
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
			err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants())
			require.NoError(t, err, "Failed to execute array access outside interpolation: %s", tt.name)
		})
	}
}
