package version

import "fmt"

const (
	VERSION = "0.1.0"
	COMMIT  = "dev"
	BUILT   = ""
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
