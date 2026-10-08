package cube

import (
	"math/rand"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestOptimalSolverMissingSortedCacheDeadline(t *testing.T) {
	if os.Getenv("CUBE_OPTIMAL_TABLES") != "1" {
		t.Skip("make test-optimal")
	}
	if os.Getenv("CUBE_TEST_SORTED_DEADLINE") != "1" {
		executable, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(executable, "-test.run=^TestOptimalSolverMissingSortedCacheDeadline$", "-test.v")
		cmd.Env = append(os.Environ(), "CUBE_TEST_SORTED_DEADLINE=1")
		output, err := cmd.CombinedOutput()
		t.Logf("%s", output)
		if err != nil {
			t.Fatalf("missing sorted cache subprocess: %v", err)
		}
		return
	}
	// Keep genuine smaller databases ready while withholding only the sorted
	// cache. A fresh process avoids a sorted table published by another test.
	tables := solverTables()
	searchEdgePatterns()
	if optimalPatternTables(tables, time.Time{}) == nil {
		t.Fatal("smaller optimal tables unavailable")
	}
	t.Setenv("CUBE_CACHE_DIR", t.TempDir())
	t.Setenv("CUBE_SMALL_TABLES", "0")
	state := uniformCubie(rand.New(rand.NewSource(2026100714)))
	if _, found, timedOut := exactCoordinateSearchLimit(state, nil, 10, time.Now().Add(time.Second)); found || timedOut {
		t.Fatal("fixture must reach sorted table initialization")
	}
	for attempt := 0; attempt < 3; attempt++ {
		// Retrying may reuse heap pages from the previous interrupted allocation.
		runtime.GC()
		const limit = 200 * time.Millisecond
		started := time.Now()
		result, err := SolveOptimal(cubeFromCoordinates(state), limit)
		elapsed := time.Since(started)
		t.Logf("200ms limit with missing sorted cache, attempt %d: %v", attempt+1, elapsed)
		if result != nil || err == nil || !strings.Contains(err.Error(), "time limit exceeded") {
			t.Fatalf("missing sorted cache: result %v, error %v", result, err)
		}
		if elapsed > limit+100*time.Millisecond {
			t.Fatalf("200ms limit took %v during sorted table setup", elapsed)
		}
		if optimalPhase1DB.Load() != nil {
			t.Fatal("interrupted setup published an incomplete sorted table")
		}
		files, err := os.ReadDir(os.Getenv("CUBE_CACHE_DIR"))
		if err != nil || len(files) != 0 {
			t.Fatalf("interrupted setup saved a cache: %v, %v", files, err)
		}
	}
}

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
	// Prerequisite setup is outside the budget for waiting on each held lock.
	solverTables()
	for _, tc := range []struct {
		name string
		lock chan struct{}
		load func(time.Time) bool
	}{
		{"coordinates", tablesLock, func(d time.Time) bool { return solverTablesLimit(d) != nil }},
		{"edges", edgeLock, func(d time.Time) bool { return searchEdgePatternsLimit(d) != nil }},
		{"sorted phase one", optimalPhase1Lock, func(d time.Time) bool { return optimalPhase1Tables(solverTables(), d) != nil }},
		{"large optimal", optimalLock, func(d time.Time) bool { return optimalPatternTables(solverTables(), d) != nil }},
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
