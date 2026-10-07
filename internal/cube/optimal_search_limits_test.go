package cube

import (
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestOptimalSolverColdCacheDeadline(t *testing.T) {
	if os.Getenv("CUBE_TEST_COLD_DEADLINE") == "1" {
		c := NewCube(3)
		moves, _ := ParseMoves("R")
		c.ApplyMoves(moves)
		started := time.Now()
		result, err := SolveOptimal(c, time.Millisecond)
		elapsed := time.Since(started)
		if result != nil || err == nil || !strings.Contains(err.Error(), "time limit") {
			t.Fatalf("cold short limit: result %v, error %v", result, err)
		}
		if elapsed > 100*time.Millisecond {
			t.Fatalf("1ms limit took %v during cold table setup", elapsed)
		}
		if tables != nil || edgeDB != nil {
			t.Fatal("interrupted setup published incomplete tables")
		}
		// Check the second initialization stage with coordinate tables ready.
		solverTables()
		started = time.Now()
		result, err = SolveOptimal(c, time.Millisecond)
		elapsed = time.Since(started)
		if result != nil || err == nil || !strings.Contains(err.Error(), "time limit") || elapsed > 100*time.Millisecond {
			t.Fatalf("cold edge tables: result %v, error %v, elapsed %v", result, err, elapsed)
		}
		if edgeDB != nil {
			t.Fatal("interrupted edge setup published incomplete tables")
		}
		// An interrupted initialization must be safe to retry.
		result, err = SolveOptimal(c, 5*time.Second)
		if err != nil {
			t.Fatal("retry after interrupted setup:", err)
		}
		c.ApplyMoves(result.Solution)
		if !c.IsSolved() {
			t.Fatal("retry answer failed replay")
		}
		return
	}
	cacheRoot := os.Getenv("CUBE_CACHE_DIR")
	if err := os.MkdirAll(cacheRoot, 0700); err != nil {
		t.Fatal(err)
	}
	cache, err := os.MkdirTemp(cacheRoot, "cold-deadline-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(cache)
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(executable, "-test.run=^TestOptimalSolverColdCacheDeadline$", "-test.v")
	cmd.Env = append(os.Environ(), "CUBE_TEST_COLD_DEADLINE=1", "CUBE_CACHE_DIR="+cache)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("cold-cache subprocess: %v\n%s", err, output)
	}
}

func TestOptimalTableInitializationWaitDeadline(t *testing.T) {
	for _, tc := range []struct {
		name string
		lock chan struct{}
		load func(time.Time) bool
	}{
		{"coordinates", tablesLock, func(d time.Time) bool { return solverTablesLimit(d) != nil }},
		{"edges", edgeLock, func(d time.Time) bool { return searchEdgePatternsLimit(d) != nil }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tc.lock <- struct{}{}
			defer func() { <-tc.lock }()
			started := time.Now()
			if tc.load(started.Add(time.Millisecond)) || time.Since(started) > 100*time.Millisecond {
				t.Fatal("waiting for table initialization ignored the deadline")
			}
		})
	}
}
