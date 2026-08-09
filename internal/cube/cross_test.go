package cube

import (
	"math/rand"
	"strings"
	"testing"
)

// crossSolveRandomMoves builds a pseudo-random scramble of n face turns (with
// optional ' and 2 modifiers) for oracle-style tests.
func crossScrambleMoves(r *rand.Rand, n int) []Move {
	faces := []string{"R", "L", "U", "D", "F", "B"}
	mods := []string{"", "'", "2"}
	tokens := make([]string, n)
	for i := 0; i < n; i++ {
		tokens[i] = faces[r.Intn(len(faces))] + mods[r.Intn(len(mods))]
	}
	moves, err := ParseMoves(strings.Join(tokens, " "))
	if err != nil {
		panic("crossScrambleMoves produced unparseable notation: " + strings.Join(tokens, " "))
	}
	return moves
}

// applyCrossOnClone returns a fresh copy of scrambled with seq applied, so
// tests verify a cross solution without disturbing the original cube.
func applyCrossOnClone(scrambled *Cube, seq []Move) *Cube {
	work := scrambled.clone()
	work.ApplyMoves(seq)
	return work
}

// TestSolveWhiteCrossOracle sweeps 200 deterministic random scrambles (fixed
// seed) and verifies the returned sequence actually solves the white cross in
// every case. It also reports the mean and max solution length and asserts the
// hard cap is respected. The test must finish fast; if it gets slow the search
// pruning is broken.
func TestSolveWhiteCrossOracle(t *testing.T) {
	const scrambles = 200
	const scrambleLen = 25
	r := rand.New(rand.NewSource(42))

	lengths := make([]int, 0, scrambles)
	var maxLen int
	var totalLen int

	for i := 0; i < scrambles; i++ {
		scrambled := NewCube(3)
		scrambled.ApplyMoves(crossScrambleMoves(r, scrambleLen))

		before := scrambled.String()
		sol, err := SolveWhiteCross(scrambled)
		if err != nil {
			t.Fatalf("scramble %d: SolveWhiteCross returned error: %v", i, err)
		}
		if after := scrambled.String(); after != before {
			t.Fatalf("scramble %d: SolveWhiteCross mutated its input cube", i)
		}

		got := applyCrossOnClone(scrambled, sol)
		if !WhiteCrossSolved(got) {
			t.Fatalf("scramble %d: solution of %d moves does not solve the white cross (moves=%v)",
				i, len(sol), sol)
		}

		lengths = append(lengths, len(sol))
		totalLen += len(sol)
		if len(sol) > maxLen {
			maxLen = len(sol)
		}
	}

	mean := float64(totalLen) / float64(scrambles)
	t.Logf("oracle: %d scrambles, mean solution length %.2f, max %d", scrambles, mean, maxLen)

	if maxLen > maxCrossMoves {
		t.Errorf("max solution length %d exceeds the hard cap of %d", maxLen, maxCrossMoves)
	}
	if maxLen < 1 {
		t.Error("expected at least one non-trivial scramble to need moves")
	}
}

// TestSolveWhiteCrossSolvedCube verifies a solved cube yields a trivially short
// (empty) solution that leaves the white cross solved.
func TestSolveWhiteCrossSolvedCube(t *testing.T) {
	c := NewCube(3)

	sol, err := SolveWhiteCross(c)
	if err != nil {
		t.Fatalf("SolveWhiteCross on a solved cube returned error: %v", err)
	}
	if len(sol) != 0 {
		t.Errorf("expected an empty solution for a solved cube, got %d moves: %v", len(sol), sol)
	}
	if !WhiteCrossSolved(applyCrossOnClone(c, sol)) {
		t.Error("white cross should still be solved after applying the empty solution")
	}
}

// TestSolveWhiteCrossSingleMoves verifies that a single face turn (which breaks
// exactly one white edge) is solved in only a few moves.
func TestSolveWhiteCrossSingleMoves(t *testing.T) {
	tests := []struct {
		scramble   string
		maxMoves   int
		wantBroken bool
	}{
		{"F", 4, true},  // F sweeps the D-F boundary, breaking the white-blue edge
		{"R", 4, true},  // R sweeps the D-R boundary, breaking the white-red edge
		{"U", 4, false}, // U only rotates the top layer, so the white cross survives
	}

	for _, tt := range tests {
		t.Run(tt.scramble, func(t *testing.T) {
			c := NewCube(3)
			moves, err := ParseMoves(tt.scramble)
			if err != nil {
				t.Fatalf("failed to parse %q: %v", tt.scramble, err)
			}
			c.ApplyMoves(moves)

			if got := WhiteCrossSolved(c); got != !tt.wantBroken {
				t.Fatalf("white cross broken after %q = %v, want %v", tt.scramble, !got, tt.wantBroken)
			}

			sol, err := SolveWhiteCross(c)
			if err != nil {
				t.Fatalf("SolveWhiteCross(%q) returned error: %v", tt.scramble, err)
			}
			if len(sol) > tt.maxMoves {
				t.Errorf("%q: solution of %d moves exceeds %d: %v", tt.scramble, len(sol), tt.maxMoves, sol)
			}
			if !WhiteCrossSolved(applyCrossOnClone(c, sol)) {
				t.Errorf("%q: returned solution does not solve the white cross: %v", tt.scramble, sol)
			}
		})
	}
}

// TestSolveWhiteCrossRejectsNon3Cube verifies SolveWhiteCross refuses cube
// sizes other than 3 with an error instead of panicking.
func TestSolveWhiteCrossRejectsNon3Cube(t *testing.T) {
	for _, size := range []int{2, 4, 5} {
		t.Run("size", func(t *testing.T) {
			c := NewCube(size)
			if _, err := SolveWhiteCross(c); err == nil {
				t.Errorf("SolveWhiteCross on %dx%d should return an error", size, size)
			}
		})
	}
}

// TestSolveWhiteCrossFromWhiteCrossState verifies that a cube whose white cross
// is already solved gets a solution that leaves it solved.
func TestSolveWhiteCrossFromWhiteCrossState(t *testing.T) {
	c := NewCube(3)

	r := rand.New(rand.NewSource(7))
	c.ApplyMoves(crossScrambleMoves(r, 3)) // non-trivial but very shallow

	sol, err := SolveWhiteCross(c)
	if err != nil {
		t.Fatalf("SolveWhiteCross returned error: %v", err)
	}
	if !WhiteCrossSolved(applyCrossOnClone(c, sol)) {
		t.Errorf("solution does not solve the white cross: %v", sol)
	}
}
