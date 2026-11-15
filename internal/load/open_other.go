//go:build !unix

package load

import (
	"fmt"
	"os"
)

// openNoFollow refuses symlinks with Lstat, then opens the path. The caller still checks the
// opened descriptor with Stat. Without O_NOFOLLOW a swap between the two calls is possible on
// this platform.
func openNoFollow(path string) (*os.File, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("%s is a symlink", path)
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("%s is not a regular file", path)
	}
	return os.Open(path)
}
