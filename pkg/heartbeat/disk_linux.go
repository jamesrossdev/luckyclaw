//go:build linux
// +build linux

package heartbeat

import "syscall"

// getDiskUsage returns the percentage of the root filesystem used.
func getDiskUsage() float64 {
	var stat syscall.Statfs_t
	if err := syscall.Statfs("/", &stat); err != nil {
		return 0
	}
	total := stat.Blocks * uint64(stat.Bsize)
	free := stat.Bavail * uint64(stat.Bsize)
	used := total - free

	if total == 0 {
		return 0
	}
	return float64(used) / float64(total) * 100.0
}
