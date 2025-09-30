package runtime

import (
	"testing"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

func TestPregMatch(t *testing.T) {
	// Initialize runtime to ensure functions are available
	err := Bootstrap()
	if err != nil {
		t.Fatalf("Failed to bootstrap runtime: %v", err)
	}

	// Get the preg_match function
	pregMatchFunc, found := registry.GlobalRegistry.GetFunction("preg_match")
	if !found || pregMatchFunc == nil {
		t.Fatal("preg_match function not found in registry")
	}

	tests := []struct {
		name            string
		pattern         string
		subject         string
		expectedResult  int64
		expectedMatches []string
	}{
		{
			name:            "Basic match",
			pattern:         "/hello/",
			subject:         "hello world",
			expectedResult:  1,
			expectedMatches: []string{"hello"},
		},
		{
			name:            "No match",
			pattern:         "/xyz/",
			subject:         "hello world",
			expectedResult:  0,
			expectedMatches: []string{},
		},
		{
			name:            "Capture groups",
			pattern:         "/(\\w+)\\s+(\\w+)/",
			subject:         "hello world",
			expectedResult:  1,
			expectedMatches: []string{"hello world", "hello", "world"},
		},
		{
			name:            "Case insensitive",
			pattern:         "/HELLO/i",
			subject:         "hello world",
			expectedResult:  1,
			expectedMatches: []string{"hello"},
		},
		{
			name:            "Empty subject",
			pattern:         "/test/",
			subject:         "",
			expectedResult:  0,
			expectedMatches: []string{},
		},
		{
			name:            "Complex pattern",
			pattern:         "/\\d{3}-\\d{4}/",
			subject:         "Call me at 123-4567",
			expectedResult:  1,
			expectedMatches: []string{"123-4567"},
		},
		{
			name:            "Optional group not present",
			pattern:         "/(foo)(bar)?/i",
			subject:         "FooXYZ",
			expectedResult:  1,
			expectedMatches: []string{"Foo", "Foo"}, // Note: only 2 elements, not 3
		},
		{
			name:            "Alternation - second branch",
			pattern:         "/(foo)|(bar)/i",
			subject:         "BAR",
			expectedResult:  1,
			expectedMatches: []string{"BAR", "", "BAR"}, // Empty string for first group
		},
		{
			name:            "Multiple optional groups",
			pattern:         "/(first)(second)?(third)?/",
			subject:         "firstthird",
			expectedResult:  1,
			expectedMatches: []string{"firstthird", "first", "", "third"}, // Middle empty preserved
		},
		{
			name:            "Empty capture group",
			pattern:         "/()/",
			subject:         "test",
			expectedResult:  1,
			expectedMatches: []string{""}, // Only one element for empty match
		},
		{
			name:            "Nested capture groups",
			pattern:         "/((inner)outer)/",
			subject:         "innerouter",
			expectedResult:  1,
			expectedMatches: []string{"innerouter", "innerouter", "inner"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear any previous regex errors
			clearRegexError()

			// Prepare arguments
			pattern := values.NewString(tt.pattern)
			subject := values.NewString(tt.subject)
			matches := values.NewArray()

			args := []*values.Value{pattern, subject, matches}

			// Call preg_match
			result, err := pregMatchFunc.Builtin(nil, args)
			if err != nil {
				t.Fatalf("preg_match failed with error: %v", err)
			}

			// Check result value
			if result.Type != values.TypeInt && result.Type != values.TypeBool {
				t.Fatalf("Expected int or bool result, got %v", result.Type)
			}

			var actualResult int64
			if result.Type == values.TypeInt {
				actualResult = result.ToInt()
			} else if result.Type == values.TypeBool {
				if result.ToBool() {
					actualResult = 1
				} else {
					actualResult = 0
				}
			}

			if actualResult != tt.expectedResult {
				t.Errorf("Expected result %d, got %d", tt.expectedResult, actualResult)
			}

			// Check matches array if there should be matches
			if tt.expectedResult > 0 {
				if matches.Type != values.TypeArray {
					t.Fatal("Expected matches to be array type")
				}

				arr := matches.Data.(*values.Array)
				if len(arr.Elements) != len(tt.expectedMatches) {
					t.Errorf("Expected %d matches, got %d", len(tt.expectedMatches), len(arr.Elements))
				}

				for i, expectedMatch := range tt.expectedMatches {
					if actualMatch, exists := arr.Elements[int64(i)]; exists {
						if actualMatch.ToString() != expectedMatch {
							t.Errorf("Match %d: expected %q, got %q", i, expectedMatch, actualMatch.ToString())
						}
					} else {
						t.Errorf("Match %d not found", i)
					}
				}
			} else {
				// No matches expected, array should be empty
				if matches.Type == values.TypeArray {
					arr := matches.Data.(*values.Array)
					if len(arr.Elements) != 0 {
						t.Errorf("Expected empty matches array, got %d elements", len(arr.Elements))
					}
				}
			}
		})
	}
}

func TestPregMatchInvalidPattern(t *testing.T) {
	// Initialize runtime
	err := Bootstrap()
	if err != nil {
		t.Fatalf("Failed to bootstrap runtime: %v", err)
	}

	pregMatchFunc, found := registry.GlobalRegistry.GetFunction("preg_match")
	if !found || pregMatchFunc == nil {
		t.Fatal("preg_match function not found in registry")
	}

	tests := []struct {
		name           string
		pattern        string
		subject        string
		expectingFalse bool
		expectedError  int
	}{
		{
			name:           "Unterminated character class",
			pattern:        "/[/",
			subject:        "hello world",
			expectingFalse: true,
			expectedError:  PREG_INTERNAL_ERROR,
		},
		{
			name:           "Invalid delimiter",
			pattern:        "hello",
			subject:        "hello world",
			expectingFalse: true,
			expectedError:  PREG_INTERNAL_ERROR,
		},
		{
			name:           "Missing closing delimiter",
			pattern:        "/hello",
			subject:        "hello world",
			expectingFalse: true,
			expectedError:  PREG_INTERNAL_ERROR,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear previous errors
			clearRegexError()

			pattern := values.NewString(tt.pattern)
			subject := values.NewString(tt.subject)
			matches := values.NewArray()

			args := []*values.Value{pattern, subject, matches}

			result, err := pregMatchFunc.Builtin(nil, args)
			if err != nil {
				t.Fatalf("Unexpected Go error: %v", err)
			}

			if tt.expectingFalse {
				if result.Type != values.TypeBool || result.ToBool() != false {
					t.Errorf("Expected false result for invalid pattern, got %v", result)
				}

				// Check error was set
				errorCode, _ := getRegexError()
				if errorCode != tt.expectedError {
					t.Errorf("Expected error code %d, got %d", tt.expectedError, errorCode)
				}
			}
		})
	}
}

func TestPregLastError(t *testing.T) {
	// Initialize runtime
	err := Bootstrap()
	if err != nil {
		t.Fatalf("Failed to bootstrap runtime: %v", err)
	}

	lastErrorFunc, found := registry.GlobalRegistry.GetFunction("preg_last_error")
	if !found || lastErrorFunc == nil {
		t.Fatal("preg_last_error function not found in registry")
	}

	// Clear any previous errors
	clearRegexError()

	// Check initial state
	result, err := lastErrorFunc.Builtin(nil, []*values.Value{})
	if err != nil {
		t.Fatalf("preg_last_error failed: %v", err)
	}

	if result.ToInt() != PREG_NO_ERROR {
		t.Errorf("Expected PREG_NO_ERROR (%d), got %d", PREG_NO_ERROR, result.ToInt())
	}

	// Set an error manually
	setRegexError(PREG_INTERNAL_ERROR, "test error")

	// Check error was set
	result, err = lastErrorFunc.Builtin(nil, []*values.Value{})
	if err != nil {
		t.Fatalf("preg_last_error failed: %v", err)
	}

	if result.ToInt() != PREG_INTERNAL_ERROR {
		t.Errorf("Expected PREG_INTERNAL_ERROR (%d), got %d", PREG_INTERNAL_ERROR, result.ToInt())
	}
}

func TestPregLastErrorMsg(t *testing.T) {
	// Initialize runtime
	err := Bootstrap()
	if err != nil {
		t.Fatalf("Failed to bootstrap runtime: %v", err)
	}

	lastErrorMsgFunc, found := registry.GlobalRegistry.GetFunction("preg_last_error_msg")
	if !found || lastErrorMsgFunc == nil {
		t.Fatal("preg_last_error_msg function not found in registry")
	}

	// Clear any previous errors
	clearRegexError()

	// Check initial state
	result, err := lastErrorMsgFunc.Builtin(nil, []*values.Value{})
	if err != nil {
		t.Fatalf("preg_last_error_msg failed: %v", err)
	}

	if result.ToString() != "" {
		t.Errorf("Expected empty error message, got %q", result.ToString())
	}

	// Set an error manually
	testMessage := "test error message"
	setRegexError(PREG_INTERNAL_ERROR, testMessage)

	// Check error message was set
	result, err = lastErrorMsgFunc.Builtin(nil, []*values.Value{})
	if err != nil {
		t.Fatalf("preg_last_error_msg failed: %v", err)
	}

	if result.ToString() != testMessage {
		t.Errorf("Expected error message %q, got %q", testMessage, result.ToString())
	}
}

func TestPregMatchAll(t *testing.T) {
	// Initialize runtime
	err := Bootstrap()
	if err != nil {
		t.Fatalf("Failed to bootstrap runtime: %v", err)
	}

	pregMatchAllFunc, found := registry.GlobalRegistry.GetFunction("preg_match_all")
	if !found || pregMatchAllFunc == nil {
		t.Fatal("preg_match_all function not found in registry")
	}

	tests := []struct {
		name            string
		pattern         string
		subject         string
		expectedResult  int64
		expectedMatches [][]string // [group][match_index]
	}{
		{
			name:           "Basic multiple matches",
			pattern:        "/\\d+/",
			subject:        "I have 123 apples and 456 oranges",
			expectedResult: 2,
			expectedMatches: [][]string{
				{"123", "456"}, // Group 0: full matches
			},
		},
		{
			name:            "No matches",
			pattern:         "/xyz/",
			subject:         "hello world",
			expectedResult:  0,
			expectedMatches: [][]string{{}}, // Empty array at index 0
		},
		{
			name:           "Capture groups",
			pattern:        "/(\\w+)\\s+(\\d+)/",
			subject:        "apple 123 banana 456",
			expectedResult: 2,
			expectedMatches: [][]string{
				{"apple 123", "banana 456"}, // Group 0: full matches
				{"apple", "banana"},          // Group 1: first capture
				{"123", "456"},               // Group 2: second capture
			},
		},
		{
			name:           "Single match",
			pattern:        "/hello/",
			subject:        "hello world",
			expectedResult: 1,
			expectedMatches: [][]string{
				{"hello"}, // Group 0: full match
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear any previous regex errors
			clearRegexError()

			// Prepare arguments
			pattern := values.NewString(tt.pattern)
			subject := values.NewString(tt.subject)
			matches := values.NewArray()

			args := []*values.Value{pattern, subject, matches}

			// Call preg_match_all
			result, err := pregMatchAllFunc.Builtin(nil, args)
			if err != nil {
				t.Fatalf("preg_match_all failed with error: %v", err)
			}

			// Check result value
			actualResult := result.ToInt()
			if actualResult != tt.expectedResult {
				t.Errorf("Expected result %d, got %d", tt.expectedResult, actualResult)
			}

			// Check matches array structure
			if matches.Type != values.TypeArray {
				t.Fatal("Expected matches to be array type")
			}

			arr := matches.Data.(*values.Array)
			if len(arr.Elements) != len(tt.expectedMatches) {
				t.Errorf("Expected %d groups, got %d", len(tt.expectedMatches), len(arr.Elements))
			}

			// Check each group
			for groupIndex, expectedGroup := range tt.expectedMatches {
				if groupArray, exists := arr.Elements[int64(groupIndex)]; exists {
					if groupArray.Type != values.TypeArray {
						t.Errorf("Group %d should be array type", groupIndex)
						continue
					}

					groupArr := groupArray.Data.(*values.Array)
					if len(groupArr.Elements) != len(expectedGroup) {
						t.Errorf("Group %d: expected %d matches, got %d", groupIndex, len(expectedGroup), len(groupArr.Elements))
					}

					// Check each match in this group
					for matchIndex, expectedMatch := range expectedGroup {
						if actualMatch, exists := groupArr.Elements[int64(matchIndex)]; exists {
							if actualMatch.ToString() != expectedMatch {
								t.Errorf("Group %d, Match %d: expected %q, got %q", groupIndex, matchIndex, expectedMatch, actualMatch.ToString())
							}
						} else {
							t.Errorf("Group %d, Match %d not found", groupIndex, matchIndex)
						}
					}
				} else {
					t.Errorf("Group %d not found", groupIndex)
				}
			}
		})
	}
}

func TestPregQuote(t *testing.T) {
	// Initialize runtime
	err := Bootstrap()
	if err != nil {
		t.Fatalf("Failed to bootstrap runtime: %v", err)
	}

	pregQuoteFunc, found := registry.GlobalRegistry.GetFunction("preg_quote")
	if !found || pregQuoteFunc == nil {
		t.Fatal("preg_quote function not found in registry")
	}

	tests := []struct {
		name           string
		input          string
		delimiter      string
		hasDelimiter   bool
		expectedOutput string
	}{
		{
			name:           "Basic meta characters",
			input:          "Hello. World? (test) [array] {object} *star* +plus+ ^start$ \\backslash|pipe",
			hasDelimiter:   false,
			expectedOutput: "Hello\\. World\\? \\(test\\) \\[array\\] \\{object\\} \\*star\\* \\+plus\\+ \\^start\\$ \\\\backslash\\|pipe",
		},
		{
			name:           "With delimiter",
			input:          "Hello/world/test",
			delimiter:      "/",
			hasDelimiter:   true,
			expectedOutput: "Hello\\/world\\/test",
		},
		{
			name:           "Empty string",
			input:          "",
			hasDelimiter:   false,
			expectedOutput: "",
		},
		{
			name:           "Only delimiters",
			input:          "////",
			delimiter:      "/",
			hasDelimiter:   true,
			expectedOutput: "\\/\\/\\/\\/",
		},
		{
			name:           "Special delimiter hash",
			input:          "test#hash@at~tilde",
			delimiter:      "#",
			hasDelimiter:   true,
			expectedOutput: "test\\#hash@at~tilde",
		},
		{
			name:           "Normal text no meta chars",
			input:          "hello world 123",
			hasDelimiter:   false,
			expectedOutput: "hello world 123",
		},
		{
			name:           "Mixed meta and normal",
			input:          "test.exe",
			hasDelimiter:   false,
			expectedOutput: "test\\.exe",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var args []*values.Value

			if tt.hasDelimiter {
				args = []*values.Value{
					values.NewString(tt.input),
					values.NewString(tt.delimiter),
				}
			} else {
				args = []*values.Value{
					values.NewString(tt.input),
				}
			}

			result, err := pregQuoteFunc.Builtin(nil, args)
			if err != nil {
				t.Fatalf("preg_quote failed with error: %v", err)
			}

			if result.Type != values.TypeString {
				t.Fatalf("Expected string result, got %v", result.Type)
			}

			actualOutput := result.ToString()
			if actualOutput != tt.expectedOutput {
				t.Errorf("Expected %q, got %q", tt.expectedOutput, actualOutput)
			}
		})
	}
}

func TestPregReplace(t *testing.T) {
	functions := GetRegexFunctions()

	var pregReplaceFunc *registry.Function
	for _, f := range functions {
		if f.Name == "preg_replace" {
			pregReplaceFunc = f
			break
		}
	}

	if pregReplaceFunc == nil {
		t.Fatal("preg_replace function not found")
	}

	tests := []struct {
		name        string
		pattern     string
		replacement string
		subject     string
		expected    string
	}{
		{
			name:        "Simple string replacement",
			pattern:     "/hello/",
			replacement: "hi",
			subject:     "hello world",
			expected:    "hi world",
		},
		{
			name:        "Number replacement",
			pattern:     "/[0-9]+/",
			replacement: "X",
			subject:     "abc123def456ghi",
			expected:    "abcXdefXghi",
		},
		{
			name:        "Capture group replacement",
			pattern:     "/a(.)c/",
			replacement: "X$1Y",
			subject:     "abc def",
			expected:    "XbY def",
		},
		{
			name:        "Case insensitive replacement",
			pattern:     "/hello/i",
			replacement: "hi",
			subject:     "Hello World",
			expected:    "hi World",
		},
		{
			name:        "Multiple replacements",
			pattern:     "/a/",
			replacement: "e",
			subject:     "banana",
			expected:    "benene",
		},
		{
			name:        "No match",
			pattern:     "/xyz/",
			replacement: "replaced",
			subject:     "hello world",
			expected:    "hello world",
		},
		{
			name:        "Empty replacement",
			pattern:     "/test/",
			replacement: "",
			subject:     "test string test",
			expected:    " string ",
		},
		{
			name:        "Word boundary",
			pattern:     "/\\bcat\\b/",
			replacement: "dog",
			subject:     "cat catch cat",
			expected:    "dog catch dog",
		},
		{
			name:        "Dot matches any character",
			pattern:     "/c.t/",
			replacement: "dog",
			subject:     "cat cut cot",
			expected:    "dog dog dog",
		},
		{
			name:        "Start and end anchors",
			pattern:     "/^hello/",
			replacement: "hi",
			subject:     "hello world hello",
			expected:    "hi world hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := []*values.Value{
				values.NewString(tt.pattern),
				values.NewString(tt.replacement),
				values.NewString(tt.subject),
			}

			result, err := pregReplaceFunc.Builtin(nil, args)
			if err != nil {
				t.Fatalf("preg_replace failed: %v", err)
			}

			if result == nil {
				t.Fatal("preg_replace returned nil")
			}

			if result.Type != values.TypeString {
				t.Fatalf("Expected string result, got %v", result.Type)
			}

			actualOutput := result.ToString()
			if actualOutput != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, actualOutput)
			}
		})
	}
}