package cli

// find_harness_test.go is the CORRECTNESS HARNESS for the `find` pattern search
// (breadthFirstSearch in find.go). It exists to pin down what "correct" means so
// a future optimized search (stage 2) can be swapped in and validated against it.
// Every test here asserts on the search OUTPUT — that the returned move sequences
// genuinely produce a state matching the target pattern — never on internal
// implementation details of the search itself.

import (
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/ehrlich-b/cube/internal/cfen"
	"github.com/ehrlich-b/cube/internal/cube"
)

// scrambleAndTarget starts from a solved 3x3 cube, applies `depth` random moves
// (seeded math/rand for reproducibility, never repeating the same face twice in
// a row so adjacent moves cannot trivially cancel), and returns the scrambled
// cube plus the solved cube's exact CFEN (no wildcards) as the target pattern.
func scrambleAndTarget(t *testing.T, seed int64, depth int) (*cube.Cube, string) {
	t.Helper()

	rng := rand.New(rand.NewSource(seed))
	faces := []cube.Face{cube.Right, cube.Left, cube.Up, cube.Down, cube.Front, cube.Back}

	scrambled := cube.NewCube(3)
	prevFace := cube.Face(-1)
	for applied := 0; applied < depth; {
		f := faces[rng.Intn(len(faces))]
		if f == prevFace {
			continue
		}
		prevFace = f

		var move cube.Move
		switch rng.Intn(3) {
		case 0:
			move = cube.Move{Face: f, Clockwise: false}
		case 1:
			move = cube.Move{Face: f, Clockwise: true}
		default:
			move = cube.Move{Face: f, Double: true}
		}
		scrambled.ApplyMove(move)
		applied++
	}

	target, err := cfen.GenerateCFEN(cube.NewCube(3))
	if err != nil {
		t.Fatalf("GenerateCFEN(solved): %v", err)
	}
	return scrambled, target
}

// TestFindSolutionsSatisfyPattern is the core regression net. For depths 1..4
// (25 seeded cases each) it verifies the search (a) finds a solution within
// maxDepth, (b) actually reaches a state matching the solved-CFEN target, and
// (c) never exceeds maxDepth (BFS optimality).
func TestFindSolutionsSatisfyPattern(t *testing.T) {
	const casesPerDepth = 25

	checked := 0
	for depth := 1; depth <= 4; depth++ {
		for seed := int64(1); seed <= casesPerDepth; seed++ {
			checked++
			scrambled, targetCFEN := scrambleAndTarget(t, seed, depth)

			target, err := cfen.ParseCFEN(targetCFEN)
			if err != nil {
				t.Fatalf("depth %d seed %d: ParseCFEN(%q): %v", depth, seed, targetCFEN, err)
			}
			isTarget := func(c *cube.Cube) bool {
				matches, err := target.MatchesCube(c)
				return err == nil && matches
			}

			results := breadthFirstSearch(scrambled, isTarget, depth)
			if len(results) == 0 {
				t.Errorf("depth %d seed %d: no solution found within maxDepth=%d", depth, seed, depth)
				continue
			}

			for _, result := range results {
				got := copyCube(scrambled)
				got.ApplyMoves(result.moves)

				matches, err := target.MatchesCube(got)
				if err != nil || !matches {
					t.Errorf("depth %d seed %d: solution %q does not reach target (err=%v)",
						depth, seed, result.notation, err)
				}
				if len(result.moves) > depth {
					t.Errorf("depth %d seed %d: solution length %d exceeds maxDepth=%d",
						depth, seed, len(result.moves), depth)
				}
			}
		}
	}
	t.Logf("checked %d scramble/target cases (depths 1-4)", checked)
}

// maskCross returns a CFEN fix predicate that pins only the cross-shaped stickers
// (4 edges + center) of the given CFEN face index, leaving everything else grey.
// CFEN face order for YB orientation is U(0), R(1), F(2), D(3), L(4), B(5).
func maskCross(face int) func(face, row, col int) bool {
	cross := [][2]int{{0, 1}, {1, 0}, {1, 1}, {1, 2}, {2, 1}}
	return func(f, row, col int) bool {
		if f != face {
			return false
		}
		for _, p := range cross {
			if p[0] == row && p[1] == col {
				return true
			}
		}
		return false
	}
}

// maskFullFace pins every sticker of the given CFEN face index.
func maskFullFace(face int) func(face, row, col int) bool {
	return func(f, _, _ int) bool { return f == face }
}

// wildcardState builds a wildcard CFEN pattern from the state reached by applying
// `scramble` to a solved cube, keeping only the stickers selected by fix() and
// wildcarding (Grey) everything else.
func wildcardState(t *testing.T, scramble string, fix func(face, row, col int) bool) *cfen.CFENState {
	t.Helper()

	c := cube.NewCube(3)
	moves, err := cube.ParseScramble(scramble)
	if err != nil {
		t.Fatalf("ParseScramble(%q): %v", scramble, err)
	}
	c.ApplyMoves(moves)

	state, err := cfen.FromCube(c, cfen.CFENOrientation{Up: cube.Yellow, Front: cube.Blue})
	if err != nil {
		t.Fatalf("FromCube for %q: %v", scramble, err)
	}
	for fi := 0; fi < 6; fi++ {
		for i := 0; i < state.Dimension*state.Dimension; i++ {
			if !fix(fi, i/state.Dimension, i%state.Dimension) {
				state.Faces[fi].Stickers[i] = cube.Grey
			}
		}
	}
	return state
}

// TestFindWildcardPattern builds hand-made wildcard patterns (white cross only,
// yellow cross only, full faces, two-faced cross) from states 1-3 moves off
// solved and asserts that every returned solution satisfies MatchesCube under
// the wildcard semantics.
func TestFindWildcardPattern(t *testing.T) {
	cases := []struct {
		name     string
		scramble string
		fix      func(face, row, col int) bool
		maxDepth int
	}{
		{"white cross only (D face)", "R", maskCross(3), 1},
		{"yellow cross only (U face)", "F", maskCross(0), 1},
		{"full top face", "F D", maskFullFace(0), 2},
		{"full front face", "R U", maskFullFace(2), 2},
		{"U and D cross", "F R U", func(f, row, col int) bool {
			return maskCross(0)(f, row, col) || maskCross(3)(f, row, col)
		}, 3},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			target := wildcardState(t, tc.scramble, tc.fix)
			solved := cube.NewCube(3)
			isTarget := func(c *cube.Cube) bool {
				matches, err := target.MatchesCube(c)
				return err == nil && matches
			}

			results := breadthFirstSearch(solved, isTarget, tc.maxDepth)
			if len(results) == 0 {
				t.Fatalf("no solution found within maxDepth=%d for pattern %q", tc.maxDepth, target.String())
			}

			for _, result := range results {
				got := cube.NewCube(3)
				got.ApplyMoves(result.moves)

				matches, err := target.MatchesCube(got)
				if err != nil || !matches {
					t.Errorf("solution %q does not satisfy wildcard pattern %q (err=%v)",
						result.notation, target.String(), err)
				}
			}
			t.Logf("pattern %s: %d solution(s), first=%q", target.String(), len(results), results[0].notation)
		})
	}
}

// TestFindNoSolutionWithinDepth verifies that a depth-4 scramble searched with
// maxDepth=2 reports no solution — and, crucially, terminates promptly instead
// of panicking or hanging.
func TestFindNoSolutionWithinDepth(t *testing.T) {
	const scramble = "R U F B" // verified: no solution within 2 moves

	start := cube.NewCube(3)
	moves, err := cube.ParseScramble(scramble)
	if err != nil {
		t.Fatalf("ParseScramble(%q): %v", scramble, err)
	}
	start.ApplyMoves(moves)

	isTarget := func(c *cube.Cube) bool { return c.IsSolved() }
	results := breadthFirstSearch(start, isTarget, 2)
	if len(results) != 0 {
		t.Fatalf("expected no solution within 2 moves, got %d (e.g. %q)", len(results), results[0].notation)
	}
}

// TestFindDepthTiming is a skipped-by-default benchmark scaffold for stage 2.
// It times the current search across depths 1..5 (stage 2 will extend this to
// 8). Enable with: RUN_FIND_TIMING=1 go test -run TestFindDepthTiming ./internal/cli/ -v
// (it also refuses to run under -short).
func TestFindDepthTiming(t *testing.T) {
	if testing.Short() {
		t.Skip("depth timing disabled under -short")
	}
	if os.Getenv("RUN_FIND_TIMING") == "" {
		t.Skip("set RUN_FIND_TIMING=1 to enable depth timing")
	}

	const casesPerDepth = 3
	maxDepth := 5 // stage 2: extend to 8

	for depth := 1; depth <= maxDepth; depth++ {
		start := time.Now()
		for seed := int64(1); seed <= casesPerDepth; seed++ {
			scrambled, targetCFEN := scrambleAndTarget(t, seed*1000, depth)
			target, err := cfen.ParseCFEN(targetCFEN)
			if err != nil {
				t.Fatalf("depth %d: ParseCFEN(%q): %v", depth, targetCFEN, err)
			}
			isTarget := func(c *cube.Cube) bool {
				matches, err := target.MatchesCube(c)
				return err == nil && matches
			}
			results := breadthFirstSearch(scrambled, isTarget, depth)
			if len(results) == 0 {
				t.Errorf("depth %d seed %d: expected a solution", depth, seed)
			}
		}
		t.Logf("depth %d: %d cases in %v", depth, casesPerDepth, time.Since(start))
	}
}
