package version

import "fmt"

const (
	VERSION = "0.1.0"
	COMMIT  = "dev"
	BUILT   = ""
)

func Version() string {
	return fmt.Sprintf("%s (%s)", VERSION, BUILT)
}
