package version

import "fmt"

const (
	VERSION = "0.1.0"
)

var (
	COMMIT = "dev"   // Set via ldflags during build
	BUILT  = ""      // Set via ldflags during build
)

func Version() string {
	return fmt.Sprintf("v%s", VERSION)
}

func Build() string {
	if BUILT != "" {
		return BUILT
	}
	return "unknown"
}

func Commit() string {
	if COMMIT != "" {
		return COMMIT
	}
	return "unknown"
}

func FullVersion() string {
	if COMMIT != "dev" && COMMIT != "" {
		return fmt.Sprintf("v%s (%s)", VERSION, COMMIT)
	}
	return fmt.Sprintf("v%s", VERSION)
}
