package runtime

import (
	"testing"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// TestSystemFunctions tests system-related functions
func TestSystemFunctions(t *testing.T) {
	functions := GetSystemFunctions()
	functionMap := make(map[string]*registry.Function)
	for _, fn := range functions {
		functionMap[fn.Name] = fn
	}

	t.Run("getenv", func(t *testing.T) {
		fn := functionMap["getenv"]
		if fn == nil {
			t.Fatal("getenv function not found")
		}

		// Test getting a specific environment variable
		result, err := fn.Builtin(nil, []*values.Value{values.NewString("PATH")})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}

		if result == nil {
			t.Error("result is nil")
			return
		}

		// PATH should exist and be a string
		if result.IsBool() && !result.ToBool() {
			t.Error("PATH environment variable should exist")
		}
	})

	t.Run("getcwd", func(t *testing.T) {
		fn := functionMap["getcwd"]
		if fn == nil {
			t.Fatal("getcwd function not found")
		}

		result, err := fn.Builtin(nil, []*values.Value{})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}

		if result == nil {
			t.Error("result is nil")
			return
		}

		// Should return a string path
		if !result.IsString() {
			t.Errorf("expected string result, got %T", result.Data)
		}

		cwd := result.ToString()
		if len(cwd) == 0 {
			t.Error("current working directory should not be empty")
		}
	})

	t.Run("escapeshellarg", func(t *testing.T) {
		fn := functionMap["escapeshellarg"]
		if fn == nil {
			t.Fatal("escapeshellarg function not found")
		}

		testCases := []struct {
			input    string
			expected string
			desc     string
		}{
			{"hello world", "'hello world'", "spaces should be quoted"},
			{"hello'world", "'hello'\\''world'", "single quotes should be escaped"},
			{"hello\"world", "'hello\"world'", "double quotes should be wrapped"},
			{"", "''", "empty string should be quoted"},
		}

		for _, tc := range testCases {
			result, err := fn.Builtin(nil, []*values.Value{values.NewString(tc.input)})
			if err != nil {
				t.Errorf("escapeshellarg(%q): unexpected error: %v", tc.input, err)
				continue
			}

			if result.ToString() != tc.expected {
				t.Errorf("escapeshellarg(%q) = %q, want %q", tc.input, result.ToString(), tc.expected)
			}
		}
	})

	t.Run("escapeshellcmd", func(t *testing.T) {
		fn := functionMap["escapeshellcmd"]
		if fn == nil {
			t.Fatal("escapeshellcmd function not found")
		}

		testCases := []struct {
			input    string
			expected string
			desc     string
		}{
			{"echo hello", "echo hello", "spaces should not be escaped"},
			{"echo hello; rm file", "echo hello\\; rm file", "semicolons should be escaped"},
			{"echo hello|grep world", "echo hello\\|grep world", "pipes should be escaped"},
			{"", "", "empty string should remain empty"},
		}

		for _, tc := range testCases {
			result, err := fn.Builtin(nil, []*values.Value{values.NewString(tc.input)})
			if err != nil {
				t.Errorf("escapeshellcmd(%q): unexpected error: %v", tc.input, err)
				continue
			}

			if result.ToString() != tc.expected {
				t.Errorf("escapeshellcmd(%q) = %q, want %q", tc.input, result.ToString(), tc.expected)
			}
		}
	})

	t.Run("shell_exec", func(t *testing.T) {
		fn := functionMap["shell_exec"]
		if fn == nil {
			t.Fatal("shell_exec function not found")
		}

		// Test basic command
		result, err := fn.Builtin(nil, []*values.Value{values.NewString("echo 'Hello World'")})
		if err != nil {
			t.Errorf("shell_exec: unexpected error: %v", err)
			return
		}

		expected := "Hello World\n"
		if result.ToString() != expected {
			t.Errorf("shell_exec('echo Hello World') = %q, want %q", result.ToString(), expected)
		}

		// Test empty command
		result, err = fn.Builtin(nil, []*values.Value{values.NewString("")})
		if err != nil {
			t.Errorf("shell_exec(''): unexpected error: %v", err)
			return
		}

		if !result.IsNull() {
			t.Error("shell_exec('') should return null")
		}
	})

	t.Run("exec", func(t *testing.T) {
		fn := functionMap["exec"]
		if fn == nil {
			t.Fatal("exec function not found")
		}

		// Test basic command
		output := values.NewArray()
		returnCode := values.NewInt(0)
		result, err := fn.Builtin(nil, []*values.Value{
			values.NewString("echo 'Hello World'"),
			output,
			returnCode,
		})
		if err != nil {
			t.Errorf("exec: unexpected error: %v", err)
			return
		}

		// Check last line
		expected := "Hello World"
		if result.ToString() != expected {
			t.Errorf("exec('echo Hello World') last line = %q, want %q", result.ToString(), expected)
		}

		// Check return code
		if returnCode.ToInt() != 0 {
			t.Errorf("exec('echo Hello World') return code = %d, want 0", returnCode.ToInt())
		}

		// Check output array
		arr := output.Data.(*values.Array)
		if len(arr.Elements) != 1 {
			t.Errorf("exec output array should have 1 element, got %d", len(arr.Elements))
		} else if arr.Elements[int64(0)].ToString() != "Hello World" {
			t.Errorf("exec output array[0] = %q, want %q", arr.Elements[int64(0)].ToString(), "Hello World")
		}
	})

	t.Run("system", func(t *testing.T) {
		fn := functionMap["system"]
		if fn == nil {
			t.Fatal("system function not found")
		}

		// Test basic command
		returnCode := values.NewInt(-1)
		result, err := fn.Builtin(nil, []*values.Value{
			values.NewString("echo 'Hello World' > /dev/null"),
			returnCode,
		})
		if err != nil {
			t.Errorf("system: unexpected error: %v", err)
			return
		}

		// system() should return empty string in our implementation
		if !result.IsString() {
			t.Errorf("system should return string, got %T", result.Data)
		}

		// Check return code
		if returnCode.ToInt() != 0 {
			t.Errorf("system('echo Hello World > /dev/null') return code = %d, want 0", returnCode.ToInt())
		}
	})

	t.Run("passthru", func(t *testing.T) {
		fn := functionMap["passthru"]
		if fn == nil {
			t.Fatal("passthru function not found")
		}

		// Test basic command
		returnCode := values.NewInt(-1)
		result, err := fn.Builtin(nil, []*values.Value{
			values.NewString("echo 'Hello World' > /dev/null"),
			returnCode,
		})
		if err != nil {
			t.Errorf("passthru: unexpected error: %v", err)
			return
		}

		// passthru() should return null
		if !result.IsNull() {
			t.Errorf("passthru should return null, got %T", result.Data)
		}

		// Check return code
		if returnCode.ToInt() != 0 {
			t.Errorf("passthru('echo Hello World > /dev/null') return code = %d, want 0", returnCode.ToInt())
		}
	})

	t.Run("proc_open", func(t *testing.T) {
		fn := functionMap["proc_open"]
		if fn == nil {
			t.Fatal("proc_open function not found")
		}

		// Create descriptor spec for pipes
		descriptorSpec := values.NewArray()
		descArr := descriptorSpec.Data.(*values.Array)

		// stdin pipe
		stdin := values.NewArray()
		stdinArr := stdin.Data.(*values.Array)
		stdinArr.Elements[int64(0)] = values.NewString("pipe")
		stdinArr.Elements[int64(1)] = values.NewString("r")
		descArr.Elements[int64(0)] = stdin

		// stdout pipe
		stdout := values.NewArray()
		stdoutArr := stdout.Data.(*values.Array)
		stdoutArr.Elements[int64(0)] = values.NewString("pipe")
		stdoutArr.Elements[int64(1)] = values.NewString("w")
		descArr.Elements[int64(1)] = stdout

		// stderr pipe
		stderr := values.NewArray()
		stderrArr := stderr.Data.(*values.Array)
		stderrArr.Elements[int64(0)] = values.NewString("pipe")
		stderrArr.Elements[int64(1)] = values.NewString("w")
		descArr.Elements[int64(2)] = stderr

		// Pipes array to be filled
		pipes := values.NewArray()

		// Open process
		process, err := fn.Builtin(nil, []*values.Value{
			values.NewString("cat"),
			descriptorSpec,
			pipes,
		})
		if err != nil {
			t.Errorf("proc_open: unexpected error: %v", err)
			return
		}

		// Should return a resource
		if !process.IsResource() {
			t.Error("proc_open should return a resource")
			return
		}

		// Close the process
		if closeFn := functionMap["proc_close"]; closeFn != nil {
			exitCode, err := closeFn.Builtin(nil, []*values.Value{process})
			if err != nil {
				t.Errorf("proc_close: unexpected error: %v", err)
			}
			if exitCode.ToInt() != 0 {
				t.Errorf("proc_close exit code = %d, want 0", exitCode.ToInt())
			}
		}
	})

	t.Run("proc_get_status", func(t *testing.T) {
		openFn := functionMap["proc_open"]
		statusFn := functionMap["proc_get_status"]
		closeFn := functionMap["proc_close"]

		if openFn == nil || statusFn == nil || closeFn == nil {
			t.Fatal("proc functions not found")
		}

		// Create descriptor spec
		descriptorSpec := values.NewArray()
		descArr := descriptorSpec.Data.(*values.Array)

		stdin := values.NewArray()
		stdinArr := stdin.Data.(*values.Array)
		stdinArr.Elements[int64(0)] = values.NewString("pipe")
		stdinArr.Elements[int64(1)] = values.NewString("r")
		descArr.Elements[int64(0)] = stdin

		stdout := values.NewArray()
		stdoutArr := stdout.Data.(*values.Array)
		stdoutArr.Elements[int64(0)] = values.NewString("pipe")
		stdoutArr.Elements[int64(1)] = values.NewString("w")
		descArr.Elements[int64(1)] = stdout

		pipes := values.NewArray()

		// Open a sleep process
		process, err := openFn.Builtin(nil, []*values.Value{
			values.NewString("sleep 0.1"),
			descriptorSpec,
			pipes,
		})
		if err != nil {
			t.Errorf("proc_open: unexpected error: %v", err)
			return
		}

		// Get status
		status, err := statusFn.Builtin(nil, []*values.Value{process})
		if err != nil {
			t.Errorf("proc_get_status: unexpected error: %v", err)
		}

		if !status.IsArray() {
			t.Error("proc_get_status should return an array")
		} else {
			statusArr := status.Data.(*values.Array)

			// Check if running field exists
			if running, ok := statusArr.Elements["running"]; ok {
				if !running.IsBool() {
					t.Error("running field should be a boolean")
				}
			} else {
				t.Error("status array should have 'running' field")
			}

			// Check if pid field exists
			if pid, ok := statusArr.Elements["pid"]; ok {
				if !pid.IsInt() {
					t.Error("pid field should be an integer")
				}
			} else {
				t.Error("status array should have 'pid' field")
			}
		}

		// Close process
		closeFn.Builtin(nil, []*values.Value{process})
	})

	t.Run("proc_terminate", func(t *testing.T) {
		openFn := functionMap["proc_open"]
		terminateFn := functionMap["proc_terminate"]
		closeFn := functionMap["proc_close"]

		if openFn == nil || terminateFn == nil || closeFn == nil {
			t.Fatal("proc functions not found")
		}

		// Create descriptor spec
		descriptorSpec := values.NewArray()
		descArr := descriptorSpec.Data.(*values.Array)

		stdin := values.NewArray()
		stdinArr := stdin.Data.(*values.Array)
		stdinArr.Elements[int64(0)] = values.NewString("pipe")
		stdinArr.Elements[int64(1)] = values.NewString("r")
		descArr.Elements[int64(0)] = stdin

		pipes := values.NewArray()

		// Open a long-running process
		process, err := openFn.Builtin(nil, []*values.Value{
			values.NewString("sleep 10"),
			descriptorSpec,
			pipes,
		})
		if err != nil {
			t.Errorf("proc_open: unexpected error: %v", err)
			return
		}

		// Terminate the process
		result, err := terminateFn.Builtin(nil, []*values.Value{process})
		if err != nil {
			t.Errorf("proc_terminate: unexpected error: %v", err)
		}

		if !result.ToBool() {
			t.Error("proc_terminate should return true on success")
		}

		// Close process
		closeFn.Builtin(nil, []*values.Value{process})
	})

	t.Run("proc_nice", func(t *testing.T) {
		fn := functionMap["proc_nice"]
		if fn == nil {
			t.Fatal("proc_nice function not found")
		}

		// Get current priority
		result, err := fn.Builtin(nil, []*values.Value{values.NewInt(0)})
		if err != nil {
			t.Errorf("proc_nice(0): unexpected error: %v", err)
			return
		}

		// Should return an integer (current priority)
		if !result.IsInt() {
			t.Error("proc_nice(0) should return current priority as int")
		}

		// Try to increase niceness (may fail without privileges)
		result, err = fn.Builtin(nil, []*values.Value{values.NewInt(1)})
		if err != nil {
			t.Errorf("proc_nice(1): unexpected error: %v", err)
			return
		}

		// Should return bool for non-zero increment
		if !result.IsBool() {
			t.Error("proc_nice(1) should return bool")
		}
	})
}

func TestPhpVersion(t *testing.T) {
	functions := GetSystemFunctions()
	functionMap := make(map[string]*registry.Function)
	for _, fn := range functions {
		functionMap[fn.Name] = fn
	}

	phpversionFunc := functionMap["phpversion"]
	if phpversionFunc == nil {
		t.Fatal("phpversion function not found")
	}

	ctx := &mockBuiltinContext{}

	tests := []struct {
		name     string
		args     []*values.Value
		expected interface{}
	}{
		{
			name:     "no arguments",
			args:     []*values.Value{},
			expected: "8.0.30",
		},
		{
			name:     "core extension",
			args:     []*values.Value{values.NewString("core")},
			expected: "8.0.30",
		},
		{
			name:     "standard extension",
			args:     []*values.Value{values.NewString("standard")},
			expected: "8.0.30",
		},
		{
			name:     "nonexistent extension",
			args:     []*values.Value{values.NewString("nonexistent")},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := phpversionFunc.Builtin(ctx, tt.args)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			switch expected := tt.expected.(type) {
			case string:
				if result.Type != values.TypeString || result.ToString() != expected {
					t.Errorf("Expected string %q, got %v", expected, result)
				}
			case bool:
				if result.Type != values.TypeBool || result.ToBool() != expected {
					t.Errorf("Expected bool %v, got %v", expected, result)
				}
			}
		})
	}
}

func TestGetLoadedExtensions(t *testing.T) {
	functions := GetSystemFunctions()
	functionMap := make(map[string]*registry.Function)
	for _, fn := range functions {
		functionMap[fn.Name] = fn
	}

	extFunc := functionMap["get_loaded_extensions"]
	if extFunc == nil {
		t.Fatal("get_loaded_extensions function not found")
	}

	ctx := &mockBuiltinContext{}

	result, err := extFunc.Builtin(ctx, []*values.Value{})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Type != values.TypeArray {
		t.Fatalf("Expected array, got %v", result.Type)
	}

	arr := result.Data.(*values.Array)
	if len(arr.Elements) == 0 {
		t.Fatal("Expected non-empty array of extensions")
	}

	// Check that "Core" extension is included
	found := false
	for _, val := range arr.Elements {
		if val != nil && val.Type == values.TypeString && val.ToString() == "Core" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected 'Core' extension to be in the list")
	}
}