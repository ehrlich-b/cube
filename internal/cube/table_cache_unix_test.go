//go:build darwin || linux

package cube

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestTableCacheRejectsFIFOReplacement(t *testing.T) {
	for _, kind := range []string{"fifo", "symlink-to-fifo"} {
		for _, limited := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/deadline=%t", kind, limited), func(t *testing.T) {
				dir := t.TempDir()
				path := filepath.Join(dir, "cache.bin")
				const size = 1 << 30
				if err := os.WriteFile(path, nil, 0600); err != nil {
					t.Fatal(err)
				}
				if err := os.Truncate(path, size); err != nil {
					t.Fatal(err)
				}
				fifo := filepath.Join(dir, "fifo")
				if err := syscall.Mkfifo(fifo, 0600); err != nil {
					t.Fatal(err)
				}
				replacement := fifo
				if kind == "symlink-to-fifo" {
					replacement = filepath.Join(dir, "link")
					if err := os.Symlink(fifo, replacement); err != nil {
						t.Fatal(err)
					}
				}
				done := make(chan error, 1)
				go func() {
					deadline := time.Time{}
					if limited {
						deadline = time.Now().Add(10 * time.Millisecond)
					}
					f, gotSize, err := openTableCacheWithOpener(path, size, size, deadline, func(path string) (*os.File, error) {
						// The helper has already statted the valid, large regular
						// file. Swap in a FIFO at precisely the vulnerable point.
						if err := os.Rename(replacement, path); err != nil {
							return nil, err
						}
						return openTableCacheFile(path)
					})
					if f != nil {
						f.Close()
						done <- fmt.Errorf("accepted replacement: size %d, error %v", gotSize, err)
					} else if gotSize != 0 || err == nil || !strings.Contains(err.Error(), "invalid table cache: "+path) {
						done <- fmt.Errorf("unclear rejection: size %d, error %v", gotSize, err)
					} else {
						done <- nil
					}
				}()
				select {
				case err := <-done:
					if err != nil {
						t.Fatal(err)
					}
				case <-time.After(time.Second):
					// Release a regressed blocking open so the test fails
					// promptly without leaving a stuck goroutine in the suite.
					fifoPath := fifo
					if kind == "fifo" {
						fifoPath = path
					}
					writer, err := os.OpenFile(fifoPath, os.O_RDWR|syscall.O_NONBLOCK, 0)
					if err == nil {
						defer writer.Close()
						select {
						case <-done:
						case <-time.After(time.Second):
						}
					}
					t.Fatalf("cache open blocked after a FIFO replaced the statted regular file (release: %v)", err)
				}
			})
		}
	}
}

func TestTableCacheRestoresBlockingReads(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cache.bin")
	const data = "regular cache"
	if err := os.WriteFile(path, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	f, size, err := openTableCache(path, int64(len(data)), int64(len(data)), time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	raw, err := f.SyscallConn()
	if err != nil {
		t.Fatal(err)
	}
	var flags uintptr
	var errno syscall.Errno
	if err := raw.Control(func(fd uintptr) {
		flags, _, errno = syscall.Syscall(syscall.SYS_FCNTL, fd, syscall.F_GETFL, 0)
	}); err != nil {
		t.Fatal(err)
	}
	if errno != 0 || flags&syscall.O_NONBLOCK != 0 {
		t.Fatal("regular cache descriptor is not blocking", flags, errno)
	}
	got, err := io.ReadAll(f)
	if err != nil || string(got) != data || size != int64(len(data)) {
		t.Fatal("regular cache read failed", string(got), size, err)
	}
}
