package cube

import (
	"math/rand"
	"testing"
	"time"
)

func TestOptimalSearchDepths(t *testing.T) {
	load := time.Now()
	solverTables()
	searchEdgePatterns()
	t.Logf("search table initialization: %v", time.Since(load))
	r := rand.New(rand.NewSource(2026100702))
	for depth := 1; depth <= 10; depth++ {
		var elapsed time.Duration
		longest := 0
		for n := 0; n < 10; n++ {
			start := NewCube(3)
			var scramble []Move
			prev := -1
			for i := 0; i < depth; i++ {
				m := r.Intn(18)
				for skipCoordinateFace(m, prev) {
					m = r.Intn(18)
				}
				prev = m
				scramble = append(scramble, coordinateMoves[m])
			}
			start.ApplyMoves(scramble)
			before := readCubie(start)
			started := time.Now()
			result, ok := FindPattern(start, NewCube(3), nil, depth)
			elapsed += time.Since(started)
			if !ok {
				t.Fatalf("depth %d scramble %s: no result", depth, FormatMoves(scramble))
			}
			if readCubie(start) != before {
				t.Fatal("search mutated input")
			}
			check := start.clone()
			check.ApplyMoves(result)
			if !check.IsSolved() {
				t.Fatalf("depth %d: invalid sequence %s", depth, FormatMoves(result))
			}
			if len(result) > depth {
				t.Fatal("search exceeded max depth")
			}
			if _, ok := FindPattern(start, NewCube(3), nil, len(result)-1); ok && len(result) > 0 {
				t.Fatal("answer is not shortest")
			}
			longest = max(longest, len(result))
		}
		t.Logf("depth %d: 10 targets, max solution %d, total %v, mean %v", depth, longest, elapsed, elapsed/10)
	}
}

func TestPatternSearchRestrictedAndFallback(t *testing.T) {
	for _, dimension := range []int{2, 3, 4} {
		t.Run(string(rune('0'+dimension)), func(t *testing.T) {
			start, target := NewCube(dimension), NewCube(dimension)
			move, _ := ParseMove("R")
			target.ApplyMove(move)
			target.ApplyMove(move)
			result, ok := FindPattern(start, target, []Move{move}, 2)
			if !ok || len(result) != 2 {
				t.Fatal("restricted alphabet lost shortest answer", result, ok)
			}
			start.ApplyMoves(result)
			if start.String() != target.String() {
				t.Fatal("fallback/restricted answer failed replay")
			}
		})
	}
	start := NewCube(3)
	rotated, _ := ParseMoves("x R U")
	start.ApplyMoves(rotated)
	target := start.clone()
	for f := range target.Faces {
		for r := range target.Faces[f] {
			for col := range target.Faces[f][r] {
				target.Faces[f][r][col] = target.Faces[f][1][1]
			}
		}
	}
	result, ok := FindPattern(start, target, nil, 2)
	if !ok || len(result) != 2 {
		t.Fatal("rotated frame search failed", result)
	}
	start.ApplyMoves(result)
	if !start.IsSolved() {
		t.Fatal("rotated result does not solve")
	}
	start, target = NewCube(3), NewCube(3)
	rotation, _ := ParseMove("x")
	target.ApplyMove(rotation)
	result, ok = FindPattern(start, target, []Move{rotation}, 1)
	if !ok || len(result) != 1 {
		t.Fatal("rotation move alphabet fallback failed")
	}
}

func TestWildcardSearchMatchesStickerBFS(t *testing.T) {
	r := rand.New(rand.NewSource(2026100703))
	for n := 0; n < 100; n++ {
		start, target := NewCube(3), NewCube(3)
		for i := 0; i < 3; i++ {
			start.ApplyMove(coordinateMoves[r.Intn(18)])
			target.ApplyMove(coordinateMoves[r.Intn(18)])
		}
		for f := range target.Faces {
			for row := range target.Faces[f] {
				for col := range target.Faces[f][row] {
					if r.Intn(10) < 9 {
						target.Faces[f][row][col] = Grey
					}
				}
			}
		}
		want, found := stickerPatternBFS(start, target, nil, 3)
		if matchesStickerPattern(start, target) {
			want, found = []Move{}, true
		}
		got, ok := FindPattern(start, target, nil, 3)
		if ok != found || (ok && len(got) != len(want)) {
			t.Fatalf("case %d: IDA* %v/%v BFS %v/%v", n, got, ok, want, found)
		}
		if ok {
			check := start.clone()
			check.ApplyMoves(got)
			if !matchesStickerPattern(check, target) {
				t.Fatal("wildcard answer fails replay")
			}
		}
	}
}

func TestOptimalSolverLimitAndContract(t *testing.T) {
	solverTables()
	searchEdgePatterns()
	for _, text := range []string{"R", "R U F2 L' B", "x R U F2 L' B"} {
		c := NewCube(3)
		scramble, _ := ParseMoves(text)
		c.ApplyMoves(scramble)
		result, err := SolveOptimal(c, time.Second)
		if err != nil {
			t.Fatal(text, err)
		}
		check := c.clone()
		check.ApplyMoves(result.Solution)
		if !check.IsSolved() {
			t.Fatal("optimal output violates solver contract")
		}
		if len(result.Solution) > len(scramble) {
			t.Fatal("optimal result is longer than known inverse")
		}
	}
	c := NewCube(3)
	moves, _ := ParseMoves("R U F2 L' B D2 R F U2 B'")
	c.ApplyMoves(moves)
	if result, err := SolveOptimal(c, time.Nanosecond); err == nil || result != nil {
		t.Fatal("expired limit returned an unproved answer")
	}
	if _, err := SolveOptimal(c, 0); err == nil {
		t.Fatal("zero limit accepted")
	}
}
