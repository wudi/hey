//go:build !linux && !darwin && !windows

package runtime

import (
	"syscall"
	"time"
)

// getStatTimes extracts access and change times from syscall.Stat_t
// Fallback implementation for platforms where we can't access raw syscall fields
// Returns current time as fallback - this is a limitation of cross-platform support
func getStatTimes(stat *syscall.Stat_t) (atime int64, ctime int64) {
	// On unsupported platforms, we can't access the raw syscall.Stat_t fields
	// Return current time as a fallback (this matches some PHP behavior on Windows)
	now := time.Now().Unix()
	return now, now
}