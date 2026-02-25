//go:build !linux
// +build !linux

package heartbeat

// getDiskUsage returns 0 for non-Linux platforms, bypassing the check.
func getDiskUsage() float64 {
	return 0
}
