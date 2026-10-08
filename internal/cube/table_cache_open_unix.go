//go:build unix

package cube

import (
	"os"
	"syscall"
)

func openTableCacheFile(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_RDONLY|syscall.O_NONBLOCK, 0)
}

// Only called after the opened descriptor has been verified as a regular file.
func setTableCacheBlocking(f *os.File) error {
	if err := syscall.SetNonblock(int(f.Fd()), false); err != nil {
		return &os.PathError{Op: "set blocking", Path: f.Name(), Err: err}
	}
	return nil
}
