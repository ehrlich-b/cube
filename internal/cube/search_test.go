package cube

import (
	"testing"
)

// isSolvedPredicate wraps Cube.IsSolved to fit SearchToTarget's
// func(*Cube) bool signature.
func isSolvedPredicate(c *Cube) bool { return c.IsSolved() }

// applyOnClone returns a fresh copy of the scrambled cube with seq applied, so
// tests can verify a search solution without disturbing the original cube.
func applyOnClone(t *testing.T, scrambled *Cube, seq []Move) *Cube {
	t.Helper()
	work := scrambled.clone()
	work.ApplyMoves(seq)
	return work
}

// TestSearchToTargetSingleMove verifies that a one-move scramble is undone by
// exactly its inverse.
func TestSearchToTargetSingleMove(t *testing.T) {
	cube := NewCube(3)
	moves, err := ParseScramble("F")
	if err != nil {
		t.Fatalf("failed to parse F: %v", err)
	}
	cube.ApplyMoves(moves)

	sol, ok := SearchToTarget(cube, isSolvedPredicate, 1)
	if !ok {
		t.Fatal("SearchToTarget should find a solution for a single F move")
	}
	if len(sol) != 1 {
		t.Fatalf("expected exactly one move, got %v", sol)
	}

	want, err := ParseMove("F'")
	if err != nil {
		t.Fatalf("failed to parse F': %v", err)
	}
	if sol[0] != want {
		t.Errorf("expected move %v, got %v", want, sol[0])
	}
	if !applyOnClone(t, cube, sol).IsSolved() {
		t.Error("returned move does not actually solve the cube")
	}
}

// TestSearchToTargetTwoMoves verifies that a two-move scramble needs exactly
// two moves: a depth-1 search finds nothing, then the depth-2 search returns a
// solution that provably solves the cube.
func TestSearchToTargetTwoMoves(t *testing.T) {
	cube := NewCube(3)
	moves, err := ParseScramble("F R")
	if err != nil {
		t.Fatalf("failed to parse F R: %v", err)
	}
	cube.ApplyMoves(moves)

	if _, ok := SearchToTarget(cube, isSolvedPredicate, 1); ok {
		t.Fatal("depth-1 search should not solve a two-move scramble")
	}

	sol, ok := SearchToTarget(cube, isSolvedPredicate, 2)
	if !ok {
		t.Fatal("SearchToTarget should find a solution at depth 2")
	}
	if len(sol) != 2 {
		t.Fatalf("expected exactly two moves, got %d: %v", len(sol), sol)
	}
	if !applyOnClone(t, cube, sol).IsSolved() {
		t.Error("returned sequence does not actually solve the cube")
	}
}

// TestSearchToTargetThreeMovesOptimal verifies iterative deepening returns the
// shortest solution: neither depth 1 nor depth 2 finds anything for F R U, and
// the depth-3 solution has length 3.
func TestSearchToTargetThreeMovesOptimal(t *testing.T) {
	cube := NewCube(3)
	moves, err := ParseScramble("F R U")
	if err != nil {
		t.Fatalf("failed to parse F R U: %v", err)
	}
	cube.ApplyMoves(moves)

	if _, ok := SearchToTarget(cube, isSolvedPredicate, 2); ok {
		t.Fatal("search must not solve F R U in two moves")
	}

	sol, ok := SearchToTarget(cube, isSolvedPredicate, 3)
	if !ok {
		t.Fatal("SearchToTarget should find a solution at depth 3")
	}
	if len(sol) != 3 {
		t.Fatalf("expected exactly three moves, got %d: %v", len(sol), sol)
	}
	if !applyOnClone(t, cube, sol).IsSolved() {
		t.Error("returned sequence does not actually solve the cube")
	}
}

// TestSearchToTargetNeverTrue verifies that an unsatisfiable predicate reports
// not-found and leaves the input cube completely untouched.
func TestSearchToTargetNeverTrue(t *testing.T) {
	cube := NewCube(3)
	moves, err := ParseScramble("F R U")
	if err != nil {
		t.Fatalf("failed to parse F R U: %v", err)
	}
	cube.ApplyMoves(moves)
	before := cube.String()

	sol, ok := SearchToTarget(cube, func(*Cube) bool { return false }, 3)
	if ok {
		t.Fatal("a never-true predicate should never be satisfied")
	}
	if len(sol) != 0 {
		t.Errorf("expected no solution moves, got %v", sol)
	}
	if after := cube.String(); after != before {
		t.Error("the input cube was mutated by a failed search")
	}
}

// TestSearchToTargetAlreadySatisfied verifies that an already-satisfied
// predicate yields an empty (but non-nil) solution.
func TestSearchToTargetAlreadySatisfied(t *testing.T) {
	cube := NewCube(3)

	sol, ok := SearchToTarget(cube, isSolvedPredicate, 3)
	if !ok {
		t.Fatal("search on an already-solved cube should succeed")
	}
	if sol == nil {
		t.Error("expected a non-nil empty solution, got nil")
	}
	if len(sol) != 0 {
		t.Errorf("expected zero moves, got %v", sol)
	}
	if !cube.IsSolved() {
		t.Error("solved cube should remain solved")
	}
}

// TestSearchToTargetRestoresWhiteCross verifies the search composes with the
// stage-1 pieces predicates: a single F move breaks the white cross, and a
// depth-1 search for WhiteCrossSolved restores it.
func TestSearchToTargetRestoresWhiteCross(t *testing.T) {
	cube := NewCube(3)
	moves, err := ParseScramble("F")
	if err != nil {
		t.Fatalf("failed to parse F: %v", err)
	}
	cube.ApplyMoves(moves)

	if WhiteCrossSolved(cube) {
		t.Fatal("white cross should be broken after a single F move")
	}

	sol, ok := SearchToTarget(cube, WhiteCrossSolved, 1)
	if !ok {
		t.Fatal("SearchToTarget should find a move that restores the white cross")
	}
	if len(sol) != 1 {
		t.Fatalf("expected exactly one move, got %d: %v", len(sol), sol)
	}
	if !WhiteCrossSolved(applyOnClone(t, cube, sol)) {
		t.Errorf("sequence %v does not restore the white cross", sol)
	}
}
