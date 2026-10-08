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
	if err := savePackedPattern(filename, want); err != nil {
		t.Fatal(err)
	}
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

func TestSaturatedPatternCacheIntegrity(t *testing.T) {
	t.Setenv("CUBE_CACHE_DIR", t.TempDir())
	const filename = "optimal-phase1-sorted-sym16-cap11-v1.bin"
	// Code three is a valid saturated two-bit bound, including packed 0xff;
	// the four-bit phase-one reader must still reject unfinished nibbles.
	want := []byte{0xfc, 0xff, 0x39, 0xe4}
	if err := savePackedPatternLimit(filename, want, time.Now().Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	if got := loadPatternBytes(filename, len(want), time.Now().Add(time.Second)); !bytes.Equal(got, want) {
		t.Fatal("saturated pattern cache round trip", got)
	}
	if loadPackedPattern(filename, len(want)) != nil {
		t.Fatal("four-bit reader accepted saturated two-bit data")
	}
	if loadPatternBytes(filename, len(want), time.Now().Add(-time.Second)) != nil {
		t.Fatal("expired saturated pattern cache load")
	}
	path := tableCachePath(filename)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, mutate := range []func([]byte) []byte{
		func(d []byte) []byte { d[len(d)-1] ^= 1; return d },
		func(d []byte) []byte { return d[:len(d)-1] },
		func(d []byte) []byte { return append(d, 0) },
		func(d []byte) []byte { d[32] ^= 1; sum := sha256.Sum256(d[32:]); copy(d, sum[:]); return d },
	} {
		if err := os.WriteFile(path, mutate(append([]byte(nil), data...)), 0600); err != nil {
			t.Fatal(err)
		}
		if loadPatternBytes(filename, len(want), time.Now().Add(time.Second)) != nil {
			t.Fatal("invalid saturated pattern cache accepted")
		}
	}
}

func TestChunkedPatternCacheCompatibility(t *testing.T) {
	t.Setenv("CUBE_CACHE_DIR", t.TempDir())
	const filename = "pattern-chunks.bin"
	// Cross a backing-array boundary, including a partial final block, while
	// retaining the byte-for-byte format of the existing contiguous cache.
	want := bytes.Repeat([]byte{0xff}, patternByteChunkSize+7)
	want[0], want[patternByteChunkSize-1], want[patternByteChunkSize] = 0xfc, 0x39, 0xe4
	if err := savePackedPattern(filename, want); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(tableCachePath(filename))
	if err != nil {
		t.Fatal(err)
	}
	d := patternByteChunks(loadPatternData(filename, len(want), patternByteChunkSize, time.Time{}))
	if d.size() != len(want) {
		t.Fatal("chunked cache size", d.size())
	}
	for i, value := range want {
		if d.at(uint32(i)) != value {
			t.Fatal("chunked cache differs from contiguous bytes", i)
		}
	}
	if err := savePatternPartsLimit(filename, d, time.Time{}); err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(tableCachePath(filename))
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("chunked persistence changed the cache format", err)
	}
	*d.byte(patternByteChunkSize) ^= 1
	if d.at(patternByteChunkSize) != want[patternByteChunkSize]^1 || d.at(patternByteChunkSize-1) != want[patternByteChunkSize-1] {
		t.Fatal("write across chunk boundary changed the wrong byte")
	}
	filled := filledPatternByteChunks(len(want), 255, time.Time{})
	if filled.size() != len(want) {
		t.Fatal("fresh pruning storage size", filled.size())
	}
	for _, chunk := range filled {
		if !bytes.Equal(chunk, bytes.Repeat([]byte{255}, len(chunk))) {
			t.Fatal("fresh pruning storage contains a visited entry")
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

func TestLargePatternCacheLoadDeadline(t *testing.T) {
	t.Setenv("CUBE_CACHE_DIR", t.TempDir())
	const filename = "pattern-deadline.bin"
	const size = 168 * optimalFlipSliceStates / 4
	f, err := os.Create(tableCachePath(filename))
	if err != nil {
		t.Fatal(err)
	}
	// A sparse cache with a valid move fingerprint reaches payload allocation,
	// reads and hashing without generating another gigabyte-scale database.
	var header [64]byte
	fingerprint := edgeMoveFingerprint()
	copy(header[32:], fingerprint[:])
	_, writeErr := f.Write(header[:])
	truncateErr := f.Truncate(size + 64)
	closeErr := f.Close()
	if writeErr != nil || truncateErr != nil || closeErr != nil {
		t.Fatalf("prepare sparse cache: %v, %v, %v", writeErr, truncateErr, closeErr)
	}
	const limit = 200 * time.Millisecond
	started := time.Now()
	if loadPatternData(filename, size, patternByteChunkSize, started.Add(limit)) != nil {
		t.Fatal("interrupted payload load returned a table")
	}
	elapsed := time.Since(started)
	t.Logf("200ms limit while loading a large pattern cache: %v", elapsed)
	if elapsed < limit || elapsed > limit+100*time.Millisecond {
		t.Fatalf("large cache load did not stop promptly at its deadline: %v", elapsed)
	}
	// Cancellation of a large load must leave a subsequent valid load usable.
	want := []byte{0xfc, 0xff, 0x39, 0xe4}
	if err := savePackedPattern(filename, want); err != nil {
		t.Fatal(err)
	}
	if got := loadPatternBytes(filename, len(want), time.Now().Add(time.Second)); !bytes.Equal(got, want) {
		t.Fatal("cache load retry after cancellation", got)
	}
}
