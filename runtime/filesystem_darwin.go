//go:build darwin

package runtime

import (
	"syscall"
)

// getStatTimes extracts access and change times from syscall.Stat_t
// macOS-specific implementation using Atimespec/Ctimespec fields
func getStatTimes(stat *syscall.Stat_t) (atime int64, ctime int64) {
	return int64(stat.Atimespec.Sec), int64(stat.Ctimespec.Sec)
}