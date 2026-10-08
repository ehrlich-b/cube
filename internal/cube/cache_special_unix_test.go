//go:build darwin || linux

package cube

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func TestCacheLoadersRejectSpecialFiles(t *testing.T) {
	loaders := []struct {
		name string
		path func() string
		load func() bool
	}{
		{"coordinates", coordinateCachePath, func() bool { return loadCoordinateTables() != nil }},
		{"coordinates-2x2", twoByTwoCachePath, func() bool { return loadTwoByTwoTables() != nil }},
		{"coordinates-deadline", coordinateCachePath, func() bool {
			return loadCoordinateTablesLimit(time.Now().Add(100*time.Millisecond)) != nil
		}},
		{"edges", edgeCachePath, func() bool { return loadEdgePatterns(time.Time{}) != nil }},
		{"edges-deadline", edgeCachePath, func() bool {
			return loadEdgePatterns(time.Now().Add(100*time.Millisecond)) != nil
		}},
		{"phase-one", func() string { return filepath.Join(filepath.Dir(coordinateCachePath()), "phase1-sym8-v1.bin") }, func() bool {
			return loadPackedPatternLimit("phase1-sym8-v1.bin", 16, time.Now().Add(100*time.Millisecond)) != nil
		}},
		{"optimal-sorted-phase-one", func() string { return tableCachePath("optimal-phase1-sorted-sym16-cap11-v1.bin") }, func() bool {
			return loadPatternBytes("optimal-phase1-sorted-sym16-cap11-v1.bin", 16, time.Now().Add(100*time.Millisecond)) != nil
		}},
		{"optimal", optimalCachePath, func() bool {
			return loadOptimalPatterns(time.Now().Add(100*time.Millisecond)) != nil
		}},
	}
	// Isolate each loader so a blocking open fails without hanging the suite.
	if name := os.Getenv("CUBE_TEST_SPECIAL_CACHE"); name != "" {
		for _, loader := range loaders {
			if loader.name == name {
				if loader.load() {
					t.Fatal("accepted special cache file")
				}
				return
			}
		}
		t.Fatal("unknown loader", name)
	}
	for _, loader := range loaders {
		for _, kind := range []string{"fifo", "symlink-to-fifo", "directory"} {
			t.Run(loader.name+"/"+kind, func(t *testing.T) {
				t.Setenv("CUBE_CACHE_DIR", t.TempDir())
				path := loader.path()
				switch kind {
				case "fifo":
					if err := syscall.Mkfifo(path, 0600); err != nil {
						t.Fatal(err)
					}
				case "symlink-to-fifo":
					fifo := filepath.Join(filepath.Dir(path), "fifo-target")
					if err := syscall.Mkfifo(fifo, 0600); err != nil {
						t.Fatal(err)
					}
					if err := os.Symlink(fifo, path); err != nil {
						t.Fatal(err)
					}
				case "directory":
					if err := os.Mkdir(path, 0700); err != nil {
						t.Fatal(err)
					}
				}
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestCacheLoadersRejectSpecialFiles$")
				cmd.Env = append(os.Environ(), "CUBE_TEST_SPECIAL_CACHE="+loader.name)
				if out, err := cmd.CombinedOutput(); err != nil {
					t.Fatalf("cache rejection failed: %v (%v)\n%s", err, ctx.Err(), out)
				}
			})
		}
	}
}
