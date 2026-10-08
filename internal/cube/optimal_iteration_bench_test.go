package cube

import (
	"math/rand"
	"os"
	"strconv"
	"testing"
	"time"
)

// Completed iterations make search changes comparable without conditioning on
// which uniform states happen to finish a complete optimality proof.
func TestOptimalIterationBenchmark(t *testing.T) {
	if os.Getenv("CUBE_OPTIMAL_ITERATION_BENCH") != "1" {
		t.Skip("CUBE_OPTIMAL_ITERATION_BENCH=1")
	}
	depth, cases := 16, 4
	for name, value := range map[string]*int{"CUBE_OPTIMAL_DEPTH": &depth, "CUBE_OPTIMAL_CASES": &cases} {
		if text := os.Getenv(name); text != "" {
			n, err := strconv.Atoi(text)
			if err != nil || n < 1 {
				t.Fatal("invalid benchmark parameter", name)
			}
			*value = n
		}
	}
	if depth > 20 {
		t.Fatal("depth exceeds God's number")
	}
	tables := solverTables()
	db := optimalPatternTables(tables, time.Time{})
	huge := optimalPhase1Tables(tables, time.Time{})
	if db == nil || huge == nil {
		t.Fatal("optimal tables unavailable")
	}
	r := rand.New(rand.NewSource(2026100714))
	for trial := 0; trial < cases; trial++ {
		state := uniformCubie(r)
		start := time.Now()
		s := largeOptimalSearch{t: tables, db: db, huge: huge, deadline: start.Add(time.Minute)}
		s.initializeAxes(state)
		found := s.searchDepth(permutationRank(state.cp[:]), sixEdgeCoordinate(state, 0), sixEdgeCoordinate(state, 1), depth, true)
		if s.timedOut {
			t.Fatal("iteration did not finish", trial)
		}
		if found {
			c := cubeFromCoordinates(state)
			for _, m := range s.path[:depth] {
				c.ApplyMove(coordinateMoves[m])
			}
			if !c.IsSolved() {
				t.Fatal("iteration solver contract", trial)
			}
		}
		t.Logf("state %d depth %d: %d nodes, %v, found %v", trial, depth, s.nodes, time.Since(start), found)
	}
}
