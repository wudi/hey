package compiler

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wudi/php-parser/lexer"
	"github.com/wudi/php-parser/parser"
	"github.com/wudi/php-parser/compiler/vm"
)

func TestEcho(t *testing.T) {
	p := parser.New(lexer.New(`<?php echo "Hello, World!";`))
	prog := p.ParseProgram()

	comp := NewSimpleCompiler()
	err := comp.CompileNode(prog)
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

			comp := NewSimpleCompiler()
			err := comp.CompileNode(prog)
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

			comp := NewSimpleCompiler()
			err := comp.CompileNode(prog)
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

			comp := NewSimpleCompiler()
			err := comp.CompileNode(prog)
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

			comp := NewSimpleCompiler()
			err := comp.CompileNode(prog)
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

			comp := NewSimpleCompiler()
			err := comp.CompileNode(prog)
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

			comp := NewSimpleCompiler()
			err := comp.CompileNode(prog)
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

			comp := NewSimpleCompiler()
			err := comp.CompileNode(prog)
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

			comp := NewSimpleCompiler()
			err := comp.CompileNode(prog)
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

			comp := NewSimpleCompiler()
			err := comp.CompileNode(prog)
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

			comp := NewSimpleCompiler()
			err := comp.CompileNode(prog)
			require.NoError(t, err, "Failed to compile %s", tc.name)

			vmCtx := vm.NewExecutionContext()
			err = vm.NewVirtualMachine().Execute(vmCtx, comp.GetBytecode(), comp.GetConstants())
			require.NoError(t, err, "Failed to execute %s", tc.name)
		})
	}
}