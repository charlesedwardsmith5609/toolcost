//go:build unix

package load

import (
	"os"
	"syscall"
)

// openNoFollow opens path read-only without following a final symlink and without blocking
// on a FIFO. The caller checks the opened descriptor with Stat.
func openNoFollow(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0)
}
