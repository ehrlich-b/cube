package cube

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCacheWritersReportPersistenceFailures(t *testing.T) {
	for _, writer := range []struct {
		name string
		path func() string
		save func() error
	}{
		{"coordinates", coordinateCachePath, func() error { return saveCoordinateBytes(embeddedCoordinates) }},
		{"edges", edgeCachePath, func() error { return saveEdgePatterns(&edgePatterns{}) }},
		{"phase-one", func() string { return tableCachePath("phase1-sym8-v1.bin") }, func() error {
			return savePackedPattern("phase1-sym8-v1.bin", []byte{0x10})
		}},
		{"optimal", optimalCachePath, func() error { return saveOptimalPatterns(&optimalPatterns{}, time.Time{}) }},
	} {
		for _, failure := range []string{"parent-is-file", "rename-to-directory", "read-only-directory"} {
			t.Run(writer.name+"/"+failure, func(t *testing.T) {
				cache := filepath.Join(t.TempDir(), "cache")
				t.Setenv("CUBE_CACHE_DIR", cache)
				if failure == "parent-is-file" {
					if err := os.WriteFile(cache, []byte("keep"), 0600); err != nil {
						t.Fatal(err)
					}
				} else {
					if err := os.Mkdir(cache, 0700); err != nil {
						t.Fatal(err)
					}
					if failure == "rename-to-directory" {
						if err := os.Mkdir(writer.path(), 0700); err != nil {
							t.Fatal(err)
						}
					} else {
						if err := os.Chmod(cache, 0500); err != nil {
							t.Fatal(err)
						}
						t.Cleanup(func() { os.Chmod(cache, 0700) })
						// Root and some filesystems do not enforce Unix write bits.
						if f, err := os.CreateTemp(cache, "probe-*"); err == nil {
							f.Close()
							os.Remove(f.Name())
							t.Skip("filesystem permits writes to a read-only directory")
						}
					}
				}
				if err := writer.save(); err == nil {
					t.Fatal("persistence failure was discarded")
				}
				if failure == "parent-is-file" {
					if data, err := os.ReadFile(cache); err != nil || string(data) != "keep" {
						t.Fatal("cache failure changed existing file", err)
					}
				} else {
					files, err := os.ReadDir(cache)
					want := 0
					if failure == "rename-to-directory" {
						want = 1
					}
					if err != nil || len(files) != want {
						t.Fatal("cache failure left temporary files", files, err)
					}
				}
			})
		}
	}
}

func TestLargeTableBuildPersistsExistingMemoryTable(t *testing.T) {
	t.Setenv("CUBE_SMALL_TABLES", "0")
	old := phase1LargeDB.Swap(&phase1Patterns{distance: []byte{0x10}})
	t.Cleanup(func() { phase1LargeDB.Store(old) })
	t.Setenv("CUBE_CACHE_DIR", t.TempDir())
	if size, err := BuildLargePhase1Tables(); err != nil || size != 65 {
		t.Fatal("in-memory table was not persisted", size, err)
	}
	if got := loadPackedPattern("phase1-sym8-v1.bin", 1); !bytes.Equal(got, []byte{0x10}) {
		t.Fatal("builder returned success without a readable cache", got)
	}
	// A different destination must not be satisfied by the shared memory table.
	t.Setenv("CUBE_CACHE_DIR", t.TempDir())
	if err := os.Mkdir(tableCachePath("phase1-sym8-v1.bin"), 0700); err != nil {
		t.Fatal(err)
	}
	if size, err := BuildLargePhase1Tables(); err == nil || size != 0 {
		t.Fatal("builder reported ready after a failed rename", size, err)
	}
}

func TestTableCacheFailuresPreserveExistingCache(t *testing.T) {
	for _, failure := range []string{"write", "close", "deadline"} {
		t.Run(failure, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "cache.bin")
			if err := os.WriteFile(path, []byte("keep"), 0600); err != nil {
				t.Fatal(err)
			}
			deadline := time.Time{}
			if failure == "deadline" {
				deadline = time.Now().Add(10 * time.Millisecond)
			}
			err := writeTableCache(path, "cache-*.bin", deadline, func(f *os.File) error {
				if _, err := f.Write([]byte("partial")); err != nil {
					return err
				}
				switch failure {
				case "write":
					return io.ErrShortWrite
				case "close":
					return f.Close()
				case "deadline":
					time.Sleep(max(0, time.Until(deadline)) + time.Millisecond)
				}
				return nil
			})
			if err == nil {
				t.Fatal("cache publication discarded an error")
			}
			if failure == "deadline" && !errors.Is(err, context.DeadlineExceeded) {
				t.Fatal("wrong deadline error", err)
			}
			if got, err := os.ReadFile(path); err != nil || string(got) != "keep" {
				t.Fatal("failure replaced the existing cache", got, err)
			}
			if files, err := os.ReadDir(filepath.Dir(path)); err != nil || len(files) != 1 {
				t.Fatal("failure left temporary files", files, err)
			}
		})
	}
}

type expiringCacheIO struct {
	deadline time.Time
	calls    int
	size     int
}

func (s *expiringCacheIO) Read(p []byte) (int, error) {
	s.calls++
	s.size = len(p)
	time.Sleep(max(0, time.Until(s.deadline)) + time.Millisecond)
	return len(p), nil
}

func (s *expiringCacheIO) Write(p []byte) (int, error) { return s.Read(p) }

func TestCacheIOChecksDeadlineDuringTransfer(t *testing.T) {
	for _, operation := range []string{"read", "write"} {
		t.Run(operation, func(t *testing.T) {
			deadline := time.Now().Add(10 * time.Millisecond)
			underlying := &expiringCacheIO{deadline: deadline}
			data := make([]byte, 128<<10)
			var n int
			var err error
			if operation == "read" {
				n, err = io.ReadFull(tableDeadlineReader{underlying, deadline}, data)
			} else {
				n, err = (tableDeadlineWriter{underlying, deadline}).Write(data)
			}
			if !errors.Is(err, context.DeadlineExceeded) || underlying.calls != 1 || underlying.size > 64<<10 || n > 64<<10 {
				t.Fatal("cache I/O continued after its deadline", n, err, underlying)
			}
		})
	}
	cache := filepath.Join(t.TempDir(), "not-created")
	t.Setenv("CUBE_CACHE_DIR", cache)
	if err := saveOptimalPatterns(&optimalPatterns{}, time.Now().Add(-time.Second)); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal("expired cache write", err)
	}
	if _, err := os.Stat(cache); !os.IsNotExist(err) {
		t.Fatal("expired cache write touched the filesystem", err)
	}
}
