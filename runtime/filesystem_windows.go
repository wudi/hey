//go:build windows

package runtime

import (
	"syscall"
	"time"
)

// getStatTimes extracts access and change times from syscall.Stat_t
// Windows-specific implementation
// Note: Windows syscall.Stat_t doesn't have the same time fields as Unix systems
func getStatTimes(stat *syscall.Stat_t) (atime int64, ctime int64) {
	// Windows syscall.Stat_t structure is different and doesn't expose
	// access/change times in the same way as Unix systems
	// We return current time as fallback, which is consistent with some PHP Windows behavior
	now := time.Now().Unix()
	return now, now
}