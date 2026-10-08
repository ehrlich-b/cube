package cube

import (
	"math/rand"
	"strings"
	"testing"
	"time"
)

func TestTwoPhaseViewsAndResumption(t *testing.T) {
	tables := solverTables()
	for _, text := range []string{"R U F2 L' B", "F U' R D B2 L", "R2 U D' F2"} {
		c := NewCube(3)
		scramble, _ := ParseMoves(text)
		c.ApplyMoves(scramble)
		views := twoPhaseViews(c, tables)
		for i := range views {
			s := &views[i]
			var path []int
			for calls := 0; calls < 1000000 && path == nil; calls++ {
				var done bool
				// Pause after every node, including within phase two.
				path, done = s.advance(tables, 30, 1)
				if done {
					s.startDepth(s.depth + 1)
				}
			}
			if path == nil {
				t.Fatalf("%s view %d did not finish", text, i)
			}
			check := c.clone()
			check.ApplyMoves(s.moves(path))
			if !check.IsSolved() {
				t.Fatalf("%s view %d failed inverse/axis replay", text, i)
			}
		}
	}
}

func TestTwoPhasePreMoveViews(t *testing.T) {
	tables := solverTables()
	c := NewCube(3)
	scramble, _ := ParseMoves("R U F2 L' B")
	c.ApplyMoves(scramble)
	views := preMoveViews(twoPhaseViews(c, tables), tables)
	if len(views) != 54 {
		t.Fatal("missing pre-move/inverse/axis views", len(views))
	}
	for i := range views {
		s := &views[i]
		var path []int
		for calls := 0; calls < 1000000 && path == nil; calls++ {
			var done bool
			path, done = s.advance(tables, 29, 1)
			if done {
				s.startDepth(s.depth + 1)
			}
		}
		if path == nil {
			t.Fatal("pre-move view did not finish", i)
		}
		check := c.clone()
		check.ApplyMoves(s.moves(path))
		if !check.IsSolved() {
			t.Fatal("pre-move/inverse/axis view failed replay", i)
		}
	}
}

func TestPhaseOneCombinedPruning(t *testing.T) {
	tables := solverTables()
	rng := rand.New(rand.NewSource(2026100711))
	stronger := false
	for trial := 0; trial < 1000; trial++ {
		state := identityCubie()
		for depth := 0; depth < 12; depth++ {
			co, eo, sl := state.twist(), state.flip(), state.slice()
			bound := tables.phase1Bound(co, eo, sl)
			if bound > depth {
				t.Fatalf("inadmissible phase-one bound %d at depth %d", bound, depth)
			}
			old := max(int(tables.TwistSlice[co*495+sl]), int(tables.FlipSlice[eo*495+sl]))
			stronger = stronger || bound > old
			state = state.mul(cubieMoves[rng.Intn(18)])
		}
	}
	if !stronger {
		t.Fatal("twist/flip pruning never strengthens the existing bound")
	}
}

func TestKociembaBudgetAndTarget(t *testing.T) {
	solverTables()
	phase1PatternTables(solverTables()) // Search budget excludes all one-time tables.
	c := NewCube(3)
	moves, _ := ParseMoves("R U F2 L' B D2 R F U2 B' L2 U R2 D F' L B2 D' R U2")
	c.ApplyMoves(moves)
	before := readCubie(c)
	// An impossible stopping goal forces budget expiry AFTER an incumbent.
	started := time.Now()
	result, err := SolveKociemba(c, KociembaOptions{TargetLength: 1, TimeLimit: 100 * time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	if elapsed := time.Since(started); elapsed < 100*time.Millisecond || elapsed > time.Second {
		t.Fatalf("budget not respected: %v", elapsed)
	}
	if len(result.Solution) == 0 || len(result.Solution) > 30 || readCubie(c) != before {
		t.Fatal("timeout lost incumbent or mutated input")
	}
	check := c.clone()
	check.ApplyMoves(result.Solution)
	if !check.IsSolved() {
		t.Fatal("budget-expired incumbent violates solver contract")
	}
	if result, err := SolveKociemba(c, KociembaOptions{TargetLength: 20, TimeLimit: time.Nanosecond}); result != nil || err == nil || !strings.Contains(err.Error(), "before finding") {
		t.Fatal("timeout without an incumbent must be an explicit error", result, err)
	}
	for _, options := range []KociembaOptions{{0, time.Second}, {31, time.Second}, {20, 0}, {20, -time.Second}} {
		if result, err := SolveKociemba(c, options); result != nil || err == nil {
			t.Fatal("invalid options accepted", options)
		}
	}
	near := NewCube(3)
	near.ApplyMove(coordinateMoves[0])
	result, err = SolveKociemba(near, KociembaOptions{TargetLength: 1, TimeLimit: time.Second})
	if err != nil || TurnCount(result.Solution) != 1 {
		t.Fatal("one-turn target did not stop on its solution", result, err)
	}
}
