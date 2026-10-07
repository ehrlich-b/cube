package cube

import (
	"crypto/sha256"
	"encoding/binary"
	"os"
	"reflect"
	"testing"
	"time"
)

func TestEdgeCacheReplayAndCorruption(t *testing.T) {
	want := searchEdgePatterns()
	t.Setenv("CUBE_CACHE_DIR", t.TempDir())
	saveEdgePatterns(want)
	if got := loadEdgePatterns(time.Time{}); !reflect.DeepEqual(got, want) {
		t.Fatal("cached edge tables differ")
	}
	data, err := os.ReadFile(edgeCachePath())
	if err != nil {
		t.Fatal(err)
	}
	for _, mutate := range []func([]byte) []byte{
		func(data []byte) []byte { data[len(data)-1] ^= 1; return data },
		func(data []byte) []byte { return data[:64] },
		func(data []byte) []byte {
			binary.LittleEndian.PutUint32(data[64:], edgePatternSize)
			sum := sha256.Sum256(data[32:])
			copy(data[:32], sum[:])
			return data
		},
		func(data []byte) []byte {
			data[32] ^= 1
			sum := sha256.Sum256(data[32:])
			copy(data[:32], sum[:])
			return data
		},
	} {
		if err := os.WriteFile(edgeCachePath(), mutate(append([]byte(nil), data...)), 0600); err != nil {
			t.Fatal(err)
		}
		if loadEdgePatterns(time.Time{}) != nil {
			t.Fatal("invalid edge cache accepted")
		}
	}
	saveEdgePatterns(want)
	if loadEdgePatterns(time.Now().Add(-time.Second)) != nil {
		t.Fatal("expired setup deadline ignored")
	}
	t.Setenv("CUBE_CACHE_DIR", edgeCachePath()) // A file cannot hold a cache directory.
	saveEdgePatterns(want)
	if loadEdgePatterns(time.Time{}) != nil {
		t.Fatal("unavailable cache accepted")
	}
}
