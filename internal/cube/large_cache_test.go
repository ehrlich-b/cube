package cube

import (
	"bytes"
	"crypto/sha256"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestPackedPatternCacheIntegrity(t *testing.T) {
	t.Setenv("CUBE_CACHE_DIR", t.TempDir())
	const filename = "pattern-test.bin"
	want := bytes.Repeat([]byte{0x11}, 16)
	want[0] = 0x10
	savePackedPattern(filename, want)
	if got := loadPackedPattern(filename, len(want)); !bytes.Equal(got, want) {
		t.Fatal("pattern cache round trip", got)
	}
	path := filepath.Join(os.Getenv("CUBE_CACHE_DIR"), filename)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, mutate := range []func([]byte) []byte{
		func(d []byte) []byte { d[len(d)-1] ^= 1; return d },
		func(d []byte) []byte { return d[:len(d)-1] },
		func(d []byte) []byte { return append(d, 0) },
		func(d []byte) []byte { d[32] ^= 1; sum := sha256.Sum256(d[32:]); copy(d, sum[:]); return d },
		func(d []byte) []byte { d[len(d)-1] = 0xff; sum := sha256.Sum256(d[32:]); copy(d, sum[:]); return d },
	} {
		if err := os.WriteFile(path, mutate(append([]byte(nil), data...)), 0600); err != nil {
			t.Fatal(err)
		}
		if loadPackedPattern(filename, len(want)) != nil {
			t.Fatal("invalid packed pattern cache accepted")
		}
	}
}

func TestLargeCacheSizeAndDeadlineGuards(t *testing.T) {
	t.Setenv("CUBE_CACHE_DIR", t.TempDir())
	f, err := os.Create(optimalCachePath())
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Truncate(optimalCacheBytes + 1); err != nil {
		f.Close()
		t.Fatal(err)
	}
	f.Close()
	var before, after runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&before)
	if loadOptimalPatterns(time.Time{}) != nil {
		t.Fatal("oversized optimal cache accepted")
	}
	runtime.ReadMemStats(&after)
	if after.TotalAlloc-before.TotalAlloc > 1<<20 {
		t.Fatal("oversized optimal cache allocated its payload")
	}
	if loadOptimalPatterns(time.Now().Add(-time.Second)) != nil {
		t.Fatal("expired cache load")
	}
	if packDistances([]uint8{0, 255}, time.Time{}) != nil {
		t.Fatal("incomplete distances packed")
	}
	if packDistances([]uint8{0, 1}, time.Now().Add(-time.Second)) != nil {
		t.Fatal("expired packing")
	}
}
