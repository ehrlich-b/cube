package cube

import (
	"os"
	"runtime"
	"testing"
	"time"
)

func TestCacheLoadersRejectOversizedFilesBeforeAllocation(t *testing.T) {
	for _, test := range []struct {
		name string
		path func() string
		load func() bool
	}{
		{"coordinates", coordinateCachePath, func() bool { return loadCoordinateTables() != nil }},
		{"coordinates with deadline", coordinateCachePath, func() bool {
			return loadCoordinateTablesLimit(time.Now().Add(time.Second)) != nil
		}},
		{"edges", edgeCachePath, func() bool { return loadEdgePatterns(time.Time{}) != nil }},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("CUBE_CACHE_DIR", t.TempDir())
			f, err := os.Create(test.path())
			if err != nil {
				t.Fatal(err)
			}
			if err := f.Truncate(64 << 20); err != nil {
				f.Close()
				t.Fatal(err)
			}
			if err := f.Close(); err != nil {
				t.Fatal(err)
			}
			var before, after runtime.MemStats
			runtime.GC()
			runtime.ReadMemStats(&before)
			accepted := test.load()
			runtime.ReadMemStats(&after)
			if accepted {
				t.Fatal("accepted oversized cache")
			}
			// Allow bookkeeping, but no allocation proportional to file size.
			if allocated := after.TotalAlloc - before.TotalAlloc; allocated > 1<<20 {
				t.Fatalf("oversized cache rejection allocated %d bytes", allocated)
			}
		})
	}
}
