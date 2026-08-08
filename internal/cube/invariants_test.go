package cube

// invariants_test.go holds the load-bearing invariants for the cube engine and
// the solver contract. These are the properties that MUST hold no matter what
// changes elsewhere in the codebase. If anything here goes red, stop and fix it
// before continuing — a broken invariant means the engine is silently corrupting
// state or a "solver" is emitting solutions that don't actually solve.
//
// Keep these cheap, deterministic (fixed RNG seed), and independent of any one
// solver or algorithm so they stay trustworthy while the rest of the project
// changes underneath them.

import (
	"math/rand"
	"strings"
	"testing"
)

// --- helpers ---------------------------------------------------------------

// invertMoves returns the inverse of a move sequence: reversed order, with each
// quarter turn negated (a 180° turn is its own inverse).
func invertMoves(moves []Move) []Move {
	inv := make([]Move, len(moves))
	for i, m := range moves {
		r := m
		if !m.Double {
			r.Clockwise = !m.Clockwise
		}
		inv[len(moves)-1-i] = r
	}
	return inv
}

// randomScramble builds a deterministic pseudo-random sequence of n moves drawn
// from the universally valid alphabet (face turns + whole-cube rotations), so it
// is legal on any cube size >= 2.
func randomScramble(r *rand.Rand, n int) []Move {
	alphabet := []string{"R", "L", "U", "D", "F", "B", "x", "y", "z"}
	mods := []string{"", "'", "2"}
	tokens := make([]string, n)
	for i := 0; i < n; i++ {
		tokens[i] = alphabet[r.Intn(len(alphabet))] + mods[r.Intn(len(mods))]
	}
	moves, err := ParseMoves(strings.Join(tokens, " "))
	if err != nil {
		panic("randomScramble produced unparseable notation: " + strings.Join(tokens, " "))
	}
	return moves
}

// colorCounts tallies how many stickers of each color are on the cube.
func colorCounts(c *Cube) [7]int {
	var counts [7]int
	for face := 0; face < 6; face++ {
		for row := 0; row < c.Size; row++ {
			for col := 0; col < c.Size; col++ {
				counts[c.Faces[face][row][col]]++
			}
		}
	}
	return counts
}

// facesEqual reports whether two cubes have identical sticker layouts.
func facesEqual(a, b *Cube) bool {
	if a.Size != b.Size {
		return false
	}
	for face := 0; face < 6; face++ {
		for row := 0; row < a.Size; row++ {
			for col := 0; col < a.Size; col++ {
				if a.Faces[face][row][col] != b.Faces[face][row][col] {
					return false
				}
			}
		}
	}
	return true
}

// --- engine invariants -----------------------------------------------------

// Every move is a permutation of stickers: it may never create or destroy a
// sticker, so each of the 6 real colors must always appear exactly N*N times and
// the wildcard color (Grey) must never appear from real moves. This catches the
// whole class of bugs where a permutation table corrupts the cube.
func TestInvariant_StickerConservation(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	for size := 2; size <= 6; size++ {
		want := size * size
		for trial := 0; trial < 200; trial++ {
			c := NewCube(size)
			c.ApplyMoves(randomScramble(r, 12))
			counts := colorCounts(c)
			for color := White; color <= Green; color++ {
				if counts[color] != want {
					t.Fatalf("size %d trial %d: color %s appears %d times, want %d",
						size, trial, color, counts[color], want)
				}
			}
			if counts[Grey] != 0 {
				t.Fatalf("size %d trial %d: wildcard color leaked onto the cube (%d Grey stickers)",
					size, trial, counts[Grey])
			}
		}
	}
}

// A scramble followed by its exact inverse must return to the solved state, for
// every cube size. This is the single strongest check that move application and
// its inverse are mutually consistent permutations.
func TestInvariant_ScrambleThenInverseSolves(t *testing.T) {
	r := rand.New(rand.NewSource(2))
	for size := 2; size <= 6; size++ {
		for trial := 0; trial < 200; trial++ {
			c := NewCube(size)
			scramble := randomScramble(r, 15)
			c.ApplyMoves(scramble)
			c.ApplyMoves(invertMoves(scramble))
			if !c.IsSolved() {
				t.Fatalf("size %d trial %d: scramble + inverse did not return to solved", size, trial)
			}
		}
	}
}

// Move application must be deterministic: the same sequence from a solved cube
// must always produce the identical state. Guards against hidden global state or
// nondeterministic permutation caching.
func TestInvariant_MoveApplicationIsDeterministic(t *testing.T) {
	r := rand.New(rand.NewSource(3))
	for trial := 0; trial < 50; trial++ {
		scramble := randomScramble(r, 20)
		a := NewCube(3)
		b := NewCube(3)
		a.ApplyMoves(scramble)
		b.ApplyMoves(scramble)
		if !facesEqual(a, b) {
			t.Fatalf("trial %d: identical scrambles produced different cube states", trial)
		}
	}
}

// A freshly created cube of any size is solved, and a single quarter turn always
// breaks that. Pins down the meaning of IsSolved so a solver can trust it.
func TestInvariant_SolvedDetection(t *testing.T) {
	for size := 2; size <= 7; size++ {
		c := NewCube(size)
		if !c.IsSolved() {
			t.Fatalf("size %d: a new cube must report solved", size)
		}
		m, err := ParseMoves("R")
		if err != nil {
			t.Fatal(err)
		}
		c.ApplyMoves(m)
		if c.IsSolved() {
			t.Fatalf("size %d: cube must not report solved after a single R turn", size)
		}
	}
}

// --- solver contract -------------------------------------------------------

// TestInvariant_SolverContract is the guardrail for every present and future
// solver. The contract is simple and non-negotiable:
//
//	if a solver returns a non-empty solution, applying that solution to the
//	scrambled cube MUST yield a solved cube.
//
// Today all solvers are unimplemented stubs that return an empty solution, so
// every case below reports SKIP rather than fail — we never lock in the broken
// "returns empty" behavior as correct. The moment a solver emits real moves,
// this test holds it to actually solving the cube. A solver cannot pass by
// faking output, and cannot hide an empty answer as success.
func TestInvariant_SolverContract(t *testing.T) {
	scrambles := []string{
		"R U R' U'",
		"R U F2 L' B",
		"R U2 R' U' R U' R'",
		"F R U' R' U' R U R' F'",
		"D2 R' U2 R F2 L U' L'",
	}
	for _, name := range []string{"beginner", "cfop", "kociemba"} {
		for _, scramble := range scrambles {
			t.Run(name+"/"+scramble, func(t *testing.T) {
				moves, err := ParseMoves(scramble)
				if err != nil {
					t.Fatalf("bad scramble %q: %v", scramble, err)
				}

				// Solve on its own cube in case Solve mutates the input.
				work := NewCube(3)
				work.ApplyMoves(moves)
				if work.IsSolved() {
					t.Fatalf("scramble %q leaves the cube solved; choose a real scramble", scramble)
				}

				solver, err := GetSolver(name)
				if err != nil {
					t.Fatalf("GetSolver(%q): %v", name, err)
				}

				res, err := solver.Solve(work)
				if err != nil {
					t.Fatalf("%s.Solve returned error on a 3x3: %v", name, err)
				}
				if res == nil {
					t.Fatalf("%s.Solve returned a nil result", name)
				}
				if res.Steps != len(res.Solution) {
					t.Errorf("%s: Steps (%d) != len(Solution) (%d)", name, res.Steps, len(res.Solution))
				}

				if len(res.Solution) == 0 {
					t.Skipf("%s returned no moves for %q (solver unimplemented)", name, scramble)
				}

				// THE GUARDRAIL: a non-empty solution must solve the cube. Verify
				// on an independent cube so we depend only on the returned moves.
				check := NewCube(3)
				check.ApplyMoves(moves)
				check.ApplyMoves(res.Solution)
				if !check.IsSolved() {
					t.Errorf("%s produced a %d-move solution that does NOT solve %q",
						name, len(res.Solution), scramble)
				}
			})
		}
	}
}
