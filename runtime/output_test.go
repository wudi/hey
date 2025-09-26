package runtime

import (
	"fmt"
	"strings"
	"testing"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// mockOutputContext extends mockBuiltinContext to capture output
type mockOutputContext struct {
	output   strings.Builder
	registry *registry.Registry
}

func (m *mockOutputContext) WriteOutput(val *values.Value) error {
	m.output.WriteString(val.ToString())
	return nil
}

func (m *mockOutputContext) GetGlobal(name string) (*values.Value, bool) {
	return nil, false
}

func (m *mockOutputContext) SetGlobal(name string, val *values.Value) {}

func (m *mockOutputContext) SymbolRegistry() *registry.Registry {
	if m.registry == nil {
		registry.Initialize()
		m.registry = registry.GlobalRegistry
	}
	return m.registry
}

func (m *mockOutputContext) LookupUserFunction(name string) (*registry.Function, bool) {
	return nil, false
}

func (m *mockOutputContext) CallUserFunction(function *registry.Function, args []*values.Value) (*values.Value, error) {
	return nil, fmt.Errorf("user function calls not supported in test mock")
}

func (m *mockOutputContext) SimpleCallUserFunction(function *registry.Function, args []*values.Value) (*values.Value, error) {
	return nil, fmt.Errorf("user function calls not supported in test mock")
}

func (m *mockOutputContext) LookupUserClass(name string) (*registry.Class, bool) {
	return nil, false
}

func (m *mockOutputContext) Halt(exitCode int, message string) error {
	return nil
}

func (m *mockOutputContext) GetExecutionContext() registry.ExecutionContextInterface {
	return nil
}

func (m *mockOutputContext) GetOutputBufferStack() registry.OutputBufferStackInterface {
	return nil
}

func (m *mockOutputContext) GetCurrentFunctionArgCount() (int, error) {
	return 0, fmt.Errorf("cannot be called from the global scope")
}

func (m *mockOutputContext) GetCurrentFunctionArg(index int) (*values.Value, error) {
	return nil, fmt.Errorf("cannot be called from the global scope")
}

func (m *mockOutputContext) GetCurrentFunctionArgs() ([]*values.Value, error) {
	return nil, fmt.Errorf("cannot be called from the global scope")
}

func (m *mockOutputContext) ThrowException(exception *values.Value) error {
	return fmt.Errorf("exception thrown in test mock: %v", exception)
}

func TestPrintR(t *testing.T) {
	printRFunc := findFunction("print_r", GetOutputFunctions())
	if printRFunc == nil {
		t.Fatal("print_r function not found")
	}

	tests := []struct {
		name     string
		args     []*values.Value
		expected string
		returns  *values.Value
	}{
		{
			name:     "null value",
			args:     []*values.Value{values.NewNull()},
			expected: "",
			returns:  values.NewBool(true),
		},
		{
			name:     "true boolean",
			args:     []*values.Value{values.NewBool(true)},
			expected: "1",
			returns:  values.NewBool(true),
		},
		{
			name:     "false boolean",
			args:     []*values.Value{values.NewBool(false)},
			expected: "",
			returns:  values.NewBool(true),
		},
		{
			name:     "integer",
			args:     []*values.Value{values.NewInt(42)},
			expected: "42",
			returns:  values.NewBool(true),
		},
		{
			name:     "float",
			args:     []*values.Value{values.NewFloat(3.14)},
			expected: "3.14",
			returns:  values.NewBool(true),
		},
		{
			name:     "string",
			args:     []*values.Value{values.NewString("Hello World")},
			expected: "Hello World",
			returns:  values.NewBool(true),
		},
		{
			name: "empty array",
			args: []*values.Value{values.NewArray()},
			expected: `Array
(
)
`,
			returns: values.NewBool(true),
		},
		{
			name: "indexed array",
			args: []*values.Value{func() *values.Value {
				arr := values.NewArray()
				arr.ArraySet(values.NewInt(0), values.NewInt(1))
				arr.ArraySet(values.NewInt(1), values.NewInt(2))
				arr.ArraySet(values.NewInt(2), values.NewInt(3))
				return arr
			}()},
			expected: `Array
(
    [0] => 1
    [1] => 2
    [2] => 3
)
`,
			returns: values.NewBool(true),
		},
		{
			name: "associative array",
			args: []*values.Value{func() *values.Value {
				arr := values.NewArray()
				arr.ArraySet(values.NewString("a"), values.NewInt(1))
				arr.ArraySet(values.NewString("b"), values.NewInt(2))
				arr.ArraySet(values.NewString("c"), values.NewInt(3))
				return arr
			}()},
			expected: `Array
(
    [a] => 1
    [b] => 2
    [c] => 3
)
`,
			returns: values.NewBool(true),
		},
		{
			name: "mixed array",
			args: []*values.Value{func() *values.Value {
				arr := values.NewArray()
				arr.ArraySet(values.NewInt(0), values.NewInt(1))
				arr.ArraySet(values.NewString("a"), values.NewInt(2))
				arr.ArraySet(values.NewInt(1), values.NewInt(3))
				arr.ArraySet(values.NewString("b"), values.NewInt(4))
				return arr
			}()},
			expected: `Array
(
    [0] => 1
    [1] => 3
    [a] => 2
    [b] => 4
)
`,
			returns: values.NewBool(true),
		},
		{
			name: "nested array",
			args: []*values.Value{func() *values.Value {
				arr := values.NewArray()
				arr.ArraySet(values.NewString("name"), values.NewString("John"))
				arr.ArraySet(values.NewString("age"), values.NewInt(30))

				hobbies := values.NewArray()
				hobbies.ArraySet(values.NewInt(0), values.NewString("reading"))
				hobbies.ArraySet(values.NewInt(1), values.NewString("coding"))
				hobbies.ArraySet(values.NewInt(2), values.NewString("gaming"))
				arr.ArraySet(values.NewString("hobbies"), hobbies)

				address := values.NewArray()
				address.ArraySet(values.NewString("city"), values.NewString("New York"))
				address.ArraySet(values.NewString("country"), values.NewString("USA"))
				arr.ArraySet(values.NewString("address"), address)

				return arr
			}()},
			expected: `Array
(
    [address] => Array
        (
            [city] => New York
            [country] => USA
        )

    [age] => 30
    [hobbies] => Array
        (
            [0] => reading
            [1] => coding
            [2] => gaming
        )

    [name] => John
)
`,
			returns: values.NewBool(true),
		},
		{
			name: "return true - simple array",
			args: []*values.Value{
				func() *values.Value {
					arr := values.NewArray()
					arr.ArraySet(values.NewInt(0), values.NewInt(1))
					arr.ArraySet(values.NewInt(1), values.NewInt(2))
					arr.ArraySet(values.NewInt(2), values.NewInt(3))
					return arr
				}(),
				values.NewBool(true),
			},
			expected: "",
			returns: values.NewString(`Array
(
    [0] => 1
    [1] => 2
    [2] => 3
)
`),
		},
		{
			name: "simple object",
			args: []*values.Value{func() *values.Value {
				obj := values.NewObject("TestClass")
				objData := obj.Data.(*values.Object)
				objData.Properties["pub"] = values.NewString("public")
				objData.Properties["priv"] = values.NewString("private")
				objData.Properties["prot"] = values.NewString("protected")
				return obj
			}()},
			expected: `TestClass Object
(
    [priv] => private
    [prot] => protected
    [pub] => public
)
`,
			returns: values.NewBool(true),
		},
		{
			name: "recursive array detection",
			args: []*values.Value{func() *values.Value {
				arr := values.NewArray()
				arr.ArraySet(values.NewInt(0), values.NewInt(1))
				arr.ArraySet(values.NewInt(1), values.NewInt(2))
				// Create a reference to the same array
				arr.ArraySet(values.NewInt(2), arr)
				return arr
			}()},
			expected: `Array
(
    [0] => 1
    [1] => 2
    [2] => Array
 *RECURSION*
)
`,
			returns: values.NewBool(true),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &mockOutputContext{}
			result, err := printRFunc.Builtin(ctx, tt.args)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Check output if not using return mode
			if len(tt.args) == 1 || !tt.args[1].ToBool() {
				if ctx.output.String() != tt.expected {
					t.Errorf("Output mismatch\nExpected:\n%q\nGot:\n%q", tt.expected, ctx.output.String())
				}
			}

			// Check return value
			if !valuesEqual(result, tt.returns) {
				t.Errorf("Return value mismatch\nExpected: %v\nGot: %v", tt.returns, result)
			}
		})
	}
}

func findFunction(name string, functions []*registry.Function) *registry.Function {
	for _, fn := range functions {
		if fn.Name == name {
			return fn
		}
	}
	return nil
}

func valuesEqual(a, b *values.Value) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	if a.Type != b.Type {
		return false
	}

	switch a.Type {
	case values.TypeNull:
		return true
	case values.TypeBool:
		return a.ToBool() == b.ToBool()
	case values.TypeInt:
		return a.ToInt() == b.ToInt()
	case values.TypeFloat:
		return a.ToFloat() == b.ToFloat()
	case values.TypeString:
		return a.ToString() == b.ToString()
	default:
		// For complex types, compare string representations
		return a.String() == b.String()
	}
}