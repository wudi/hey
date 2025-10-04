package main

import (
	"errors"
	"fmt"
	"testing"
)

func TestFormatErrorMessageIncludeStack(t *testing.T) {
	root := errors.New("require: failed to read /tmp/missing.php: open /tmp/missing.php: No such file or directory")
	vmErr := fmt.Errorf("vm error at ip=361 opcode=REQUIRE_ONCE in /tmp/test_missing.php on line 2: %w", root)
	execErr := fmt.Errorf("execution error in /tmp/bootstrap.php: %w", vmErr)

	got := formatErrorMessage(execErr)
	want := "Error: require: failed to read /tmp/missing.php: open /tmp/missing.php: No such file or directory\nInclude stack:\n  - /tmp/test_missing.php:2 (opcode REQUIRE_ONCE)"

	if got != want {
		t.Fatalf("unexpected format output\nwant:\n%s\n\ngot:\n%s", want, got)
	}
}

func TestFormatErrorMessageDeduplicatesFrames(t *testing.T) {
	root := errors.New("require: failed to read /tmp/missing.php: open /tmp/missing.php: No such file or directory")
	vmInner := fmt.Errorf("vm error at ip=361 opcode=REQUIRE_ONCE in /tmp/test_missing.php on line 2: %w", root)
	vmOuter := fmt.Errorf("vm error at ip=361 opcode=REQUIRE_ONCE in /tmp/test_missing.php on line 2: %w", vmInner)

	got := formatErrorMessage(vmOuter)
	want := "Error: require: failed to read /tmp/missing.php: open /tmp/missing.php: No such file or directory\nInclude stack:\n  - /tmp/test_missing.php:2 (opcode REQUIRE_ONCE)"

	if got != want {
		t.Fatalf("unexpected format output with duplicates\nwant:\n%s\n\ngot:\n%s", want, got)
	}
}

func TestFormatErrorMessageExecutionFallback(t *testing.T) {
	root := errors.New("syntax error, unexpected T_STRING in /tmp/file.php on line 1")
	execErr := fmt.Errorf("execution error in /tmp/file.php: %w", root)

	got := formatErrorMessage(execErr)
	want := "Error: syntax error, unexpected T_STRING in /tmp/file.php on line 1\nInclude stack:\n  - /tmp/file.php"

	if got != want {
		t.Fatalf("unexpected execution fallback output\nwant:\n%s\n\ngot:\n%s", want, got)
	}
}
