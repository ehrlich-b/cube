//go:build !wasm

package cube

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestTwoByTwoCachePreservesEveryCoordinate(t *testing.T) {
	t.Setenv("CUBE_CACHE_DIR", t.TempDir())
	want := solverTables()
	if err := saveTwoByTwoTables(want); err != nil {
		t.Fatal(err)
	}
	if got := loadTwoByTwoTables(); !reflect.DeepEqual(got, want) {
		t.Fatal("2x2 cache changed coordinate transitions or pruning distances")
	}
	data, err := os.ReadFile(twoByTwoCachePath())
	if err != nil {
		t.Fatal(err)
	}
	for name, broken := range map[string][]byte{
		"truncated":   data[:len(data)-1],
		"appended":    append(append([]byte{}, data...), 0),
		"stale asset": append([]byte{}, data...),
		"corrupt":     append([]byte{}, data...),
	} {
		t.Run(name, func(t *testing.T) {
			if name == "stale asset" {
				broken[0] ^= 1
			}
			if name == "corrupt" {
				broken[len(broken)-1] ^= 1
			}
			if err := os.WriteFile(twoByTwoCachePath(), broken, 0600); err != nil {
				t.Fatal(err)
			}
			if loadTwoByTwoTables() != nil {
				t.Fatal("accepted invalid 2x2 cache")
			}
		})
	}
}

func TestTwoByTwoCacheRejectsNonRegularAndOversizedFiles(t *testing.T) {
	t.Setenv("CUBE_CACHE_DIR", t.TempDir())
	if err := os.Mkdir(twoByTwoCachePath(), 0700); err != nil {
		t.Fatal(err)
	}
	if loadTwoByTwoTables() != nil {
		t.Fatal("accepted a directory as a cache")
	}
	if err := os.Remove(twoByTwoCachePath()); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(twoByTwoCachePath())
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
	if loadTwoByTwoTables() != nil {
		t.Fatal("accepted oversized 2x2 cache")
	}
}

func TestTwoByTwoSolveWithUnavailableCache(t *testing.T) {
	// A regular file in place of the cache directory prevents every write,
	// including for privileged test runners where chmod is insufficient.
	path := filepath.Join(t.TempDir(), "unavailable")
	if err := os.WriteFile(path, []byte("blocked"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CUBE_CACHE_DIR", path)
	previous := tables
	tables = nil
	t.Cleanup(func() { tables = previous })
	c := NewCube(2)
	moves, _ := ParseMoves("R U F2 L' B")
	if err := c.ApplyMoves(moves); err != nil {
		t.Fatal(err)
	}
	result, err := SolveNxN(c, KociembaOptions{TargetLength: 20, TimeLimit: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	if err := c.ApplyMoves(result.Solution); err != nil || !c.IsSolved() || !nxnCenterMatched(c) {
		t.Fatal("2x2 fallback failed center-matched replay", err)
	}
}
