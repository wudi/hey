//go:build linux

package runtime

import (
	"syscall"
)

// getStatTimes extracts access and change times from syscall.Stat_t
// Linux-specific implementation using Atim/Ctim fields
func getStatTimes(stat *syscall.Stat_t) (atime int64, ctime int64) {
	return int64(stat.Atim.Sec), int64(stat.Ctim.Sec)
}