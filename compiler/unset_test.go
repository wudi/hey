package compiler

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wudi/php-parser/compiler/runtime"
	"github.com/wudi/php-parser/compiler/vm"
	"github.com/wudi/php-parser/lexer"
	"github.com/wudi/php-parser/parser"
)

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
