package compiler

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/wudi/hey/compiler/lexer"
	"github.com/wudi/hey/compiler/parser"
	"github.com/wudi/hey/registry"
	"github.com/stretchr/testify/require"
)

func TestTraitAdaptations(t *testing.T) {
	tests := []struct {
		name     string
		phpCode  string
		expected string
	}{
		{
			name: "Trait precedence (insteadof)",
			phpCode: `<?php
trait A {
	public function test() {
		return "A::test";
	}
}

trait B {
	public function test() {
		return "B::test";
	}
}

class TestClass {
	use A, B {
		A::test insteadof B;
	}
}

$obj = new TestClass();
echo $obj->test();
?>`,
			expected: "A::test",
		},
		{
			name: "Trait aliasing",
			phpCode: `<?php
trait A {
	public function test() {
		return "A::test";
	}
}

trait B {
	public function test() {
		return "B::test";
	}
}

class TestClass {
	use A, B {
		A::test as testA;
		B::test as testB;
		A::test insteadof B;
	}
}

$obj = new TestClass();
echo $obj->test() . "\n";
echo $obj->testA() . "\n";
echo $obj->testB() . "\n";
?>`,
			expected: "A::test\nA::test\nB::test\n",
		},
		{
			name: "Multiple trait precedence",
			phpCode: `<?php
trait A {
	public function test() {
		return "A::test";
	}
}

trait B {
	public function test() {
		return "B::test";
	}
}

trait C {
	public function test() {
		return "C::test";
	}
}

class TestClass {
	use A, B, C {
		A::test insteadof B, C;
		B::test as testFromB;
		C::test as testFromC;
	}
}

$obj = new TestClass();
echo $obj->test() . "\n";
echo $obj->testFromB() . "\n";
echo $obj->testFromC() . "\n";
?>`,
			expected: "A::test\nB::test\nC::test\n",
		},
		{
			name: "Trait visibility change",
			phpCode: `<?php
trait A {
	public function hello() {
		return "hello";
	}
}

class TestClass {
	use A {
		A::hello as private;
	}

	public function publicAccess() {
		return $this->hello();
	}
}

$obj = new TestClass();
echo $obj->publicAccess();
?>`,
			expected: "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize registry
			registry.Initialize()

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

func TestTraitAdaptationErrors(t *testing.T) {
	tests := []struct {
		name        string
		phpCode     string
		expectedErr string
	}{
		{
			name: "Trait not found in precedence",
			phpCode: `<?php
trait A {
	public function test() {
		return "A::test";
	}
}

class TestClass {
	use A {
		NonExistent::test insteadof A;
	}
}
?>`,
			expectedErr: "trait NonExistent not found in precedence rule",
		},
		{
			name: "Method not found in alias",
			phpCode: `<?php
trait A {
	public function test() {
		return "A::test";
	}
}

class TestClass {
	use A {
		A::nonExistentMethod as aliasMethod;
	}
}
?>`,
			expectedErr: "method nonExistentMethod not found in trait A",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize registry
			registry.Initialize()

			// Parse the PHP code
			p := parser.New(lexer.New(tt.phpCode))
			prog := p.ParseProgram()
			require.NotNil(t, prog, "Failed to parse program for test: %s", tt.name)

			// Compile the program - should fail
			comp := NewCompiler()
			err := comp.Compile(prog)
			if err == nil {
				t.Fatalf("Expected compilation to fail for %s", tt.name)
			}

			if !strings.Contains(err.Error(), tt.expectedErr) {
				t.Errorf("Expected error containing %q, got: %v", tt.expectedErr, err)
			}
		})
	}
}