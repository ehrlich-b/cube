package cube

// solving_db_test.go — first-ever tests for internal/cube/solving_db.go.
//
// The file under test is unwired (NewSolvingDB is called from nowhere) and, until
// this file, completely unexercised. These tests document what it ACTUALLY does
// and pin several suspected defects (deliberately NOT fixed — see the PINNED
// comments below). The wildcard semantics rely on Color::Grey being a "don't
// care" as implemented by SolvingPattern.Matches.

import (
	"testing"
)

// --- helpers ---------------------------------------------------------------

// mustParse parses a move sequence or fails the test. Standard-library-only.
func mustParse(t *testing.T, seq string) []Move {
	t.Helper()
	moves, err := ParseMoves(seq)
	if err != nil {
		t.Fatalf("ParseMoves(%q) error: %v", seq, err)
	}
	return moves
}

// apply builds a solved 3x3 cube and applies seq to it.
func apply(t *testing.T, seq string) *Cube {
	t.Helper()
	c := NewCube(3)
	c.ApplyMoves(mustParse(t, seq))
	return c
}

// greyPattern builds an all-Grey SolvingPattern of the given size, the wildcard
// base case. Built programmatically rather than as a literal.
func greyPattern(size int) *SolvingPattern {
	p := &SolvingPattern{}
	for f := 0; f < 6; f++ {
		p.Faces[f] = make([][]Color, size)
		for r := 0; r < size; r++ {
			p.Faces[f][r] = make([]Color, size)
			for c := 0; c < size; c++ {
				p.Faces[f][r][c] = Grey
			}
		}
	}
	return p
}

// copyFaces deep-copies a cube's faces so a pattern is immune to later mutation
// of the cube it was derived from.
func copyFaces(c *Cube) [6][][]Color {
	var out [6][][]Color
	for f := 0; f < 6; f++ {
		out[f] = make([][]Color, c.Size)
		for r := 0; r < c.Size; r++ {
			out[f][r] = append([]Color(nil), c.Faces[f][r]...)
		}
	}
	return out
}

// facesEqual2D reports whether two faces are identical sticker-for-sticker.
// (Named 2D to avoid colliding with facesEqual(a, b *Cube) in invariants_test.go.)
func facesEqual2D(a, b [][]Color) bool {
	if len(a) != len(b) {
		return false
	}
	for r := range a {
		for c := range a[r] {
			if a[r][c] != b[r][c] {
				return false
			}
		}
	}
	return true
}

// --- 1. all-Grey wildcard base case -----------------------------------------

func TestSolvingPatternAllGreyWildcard(t *testing.T) {
	pat := greyPattern(3)
	solved := NewCube(3)
	scrambled := apply(t, "R U R' U' F U R U' R' F' R U2")

	if !pat.Matches(solved) {
		t.Errorf("all-Grey pattern must match a solved 3x3 cube (every cell is a wildcard)")
	}
	if !pat.Matches(scrambled) {
		t.Errorf("all-Grey pattern must match a scrambled 3x3 cube (every cell is a wildcard)")
	}
}

// --- 2. exact copy of a solved cube's faces ---------------------------------

func TestSolvingPatternExactCopy(t *testing.T) {
	solved := NewCube(3)
	pat := &SolvingPattern{Faces: copyFaces(solved), Name: "exact-solved-copy"}

	if !pat.Matches(solved) {
		t.Fatalf("exact-copy pattern must match the solved cube it was copied from")
	}

	turned := NewCube(3)
	turned.ApplyMoves(mustParse(t, "R"))
	if pat.Matches(turned) {
		t.Errorf("exact-copy pattern must NOT match a cube after a single R turn")
	}

	// The pattern is a defensively-copied snapshot: it stays a solved pattern.
	if !pat.Matches(NewCube(3)) {
		t.Errorf("exact-copy pattern should still match a fresh solved cube (copy, not alias)")
	}
}

// --- 3. size guard ----------------------------------------------------------

func TestSolvingPatternSizeGuard(t *testing.T) {
	// A 3x3 pattern against a 2x2 cube. Matches checks cube.Size != len(p.Faces[0])
	// first and must return false rather than indexing a 2x2 cube with 3x3 ranges.
	pat3 := greyPattern(3)
	tiny := NewCube(2)

	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Matches panicked on size mismatch (expected clean false): %v", r)
			}
		}()
		if pat3.Matches(tiny) {
			t.Errorf("3x3 pattern must not match a 2x2 cube")
		}
	}()

	// And the mirror direction: a 2x2 pattern vs a 3x3 cube.
	pat2 := greyPattern(2)
	big := NewCube(3)
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Matches panicked on size mismatch (expected clean false): %v", r)
			}
		}()
		if pat2.Matches(big) {
			t.Errorf("2x2 pattern must not match a 3x3 cube")
		}
	}()
}

// --- 4. partial wildcard: exactly one face fixed ----------------------------
//
// The fixed face is DOWN (all White on a solved cube).
//   - "U" does NOT disturb the Down face: a U turn permutes only the Up face and
//     the top ring of the four side faces; the Down face is not part of that ring,
//     so its stickers are provably untouched. This is why U is the "unchanged"
//     move.
//   - "R" DOES disturb the Down face: an R turn rotates the entire right layer,
//     which includes the Down face's right column; that column is fed from the
//     Back face's right column (Green on a solved cube), so after R the Down
//     face right column is no longer White. This is why R is the "changed" move.

func TestSolvingPatternPartialWildcardDownFace(t *testing.T) {
	pat := greyPattern(3)
	for r := 0; r < 3; r++ {
		for c := 0; c < 3; c++ {
			pat.Faces[Down][r][c] = White // fix exactly one face; the other five stay Grey
		}
	}

	solved := NewCube(3)
	if !pat.Matches(solved) {
		t.Fatalf("Down-fixed pattern must match a solved cube (Down is all White)")
	}

	uTurned := apply(t, "U")
	if !facesEqual2D(uTurned.Faces[Down], solved.Faces[Down]) {
		t.Fatalf("test precondition broken: this engine's U turn touched the Down face")
	}
	if uTurned.IsSolved() {
		t.Fatalf("test precondition broken: U turn from solved should not be a solved cube")
	}
	if !pat.Matches(uTurned) {
		t.Errorf("pattern must match the U-turned cube: the fixed (Down) face is provably unchanged; all other changes are wildcarded away")
	}

	rTurned := apply(t, "R")
	if facesEqual2D(rTurned.Faces[Down], solved.Faces[Down]) {
		t.Fatalf("test precondition broken: this engine's R turn left the Down face White-only")
	}
	if pat.Matches(rTurned) {
		t.Errorf("pattern must NOT match the R-turned cube: the Down face right column changed (no longer all White)")
	}
}

// --- 5. NewSolvingDB / GetPhases ---------------------------------------------
//
// Read from load4LookLL (not guessed): it appends to exactly two phase keys and
// never creates any other:
//   db.algorithms["oll"] -> {4-Look-OLL-Cross, 4-Look-OLL-Sune}
//   db.algorithms["pll"] -> {4-Look-PLL-T-Perm, 4-Look-PLL-U-Perm}
// GetPhases therefore returns exactly these two names; map iteration order is
// unordered, so we assert membership and length, not order.

func TestNewSolvingDBAndPhases(t *testing.T) {
	db := NewSolvingDB()
	if db == nil {
		t.Fatalf("NewSolvingDB() returned nil")
	}

	phases := db.GetPhases()
	have := map[string]bool{}
	for _, p := range phases {
		have[p] = true
	}
	if len(phases) != 2 {
		t.Fatalf("GetPhases() = %v (len %d); want exactly the 2 phases load4LookLL populates: oll, pll", phases, len(phases))
	}
	if !have["oll"] || !have["pll"] {
		t.Errorf("GetPhases() = %v; missing 'oll' and/or 'pll'", phases)
	}

	// PINNED (suspected defect / dead load): every algorithm is stored with
	// Pattern == nil and neither findOLLMove nor findPLLMove ever reads
	// algo.Pattern, so SolvingPattern/Matches is unreachable from the database.
	// The "database" is really two hand-written if/else chains.
	for _, phase := range phases {
		for _, algo := range db.algorithms[phase] {
			if algo.Pattern != nil {
				t.Errorf("%q: expected nil Pattern (none are ever populated); got a pattern", algo.Name)
			}
		}
	}
}

// --- 6. FindNextMove routing -------------------------------------------------
//
// The function has three exits, all of which must return nil-without-panic:
//   (a) phase key absent from the map  -> the explicit `exists` check fails.
//   (b) phase == "oll"                 -> findOLLMove (covered elsewhere).
//   (c) phase == "pll"                 -> findPLLMove (covered elsewhere).
//   (d) phase key PRESENT but not oll/pll -> the if/else falls through to nil.
//
// Real DB keys are only ever "oll"/"pll" (load4LookLL), so (a) covers anything
// else on a real DB. To reach (d) we build a synthetic DB containing a phase key
// that is neither oll nor pll.

func TestFindNextMoveRouting(t *testing.T) {
	db := NewSolvingDB()
	solved := NewCube(3)

	// (a) keys load4LookLL never creates: f2l, cross are documented phases that
	// are never populated; "" and "OLL" are just non-keys ("OLL" != "oll").
	for _, phase := range []string{"f2l", "cross", "", "OLL", "PLL", "oll!", "xyz"} {
		if got := db.FindNextMove(solved, phase); got != nil {
			t.Errorf("FindNextMove(%q) on real DB = %+v; want nil (key absent)", phase, got)
		}
	}

	// (d) a key that exists but is neither "oll" nor "pll" must fall through to nil.
	synthetic := &SolvingDB{
		algorithms: map[string][]SolvingAlgorithm{
			"bogus": {{Name: "Bogus-Algo", Priority: 1, Phase: "bogus"}},
		},
	}
	if got := synthetic.FindNextMove(solved, "bogus"); got != nil {
		t.Errorf("FindNextMove(\"bogus\") on synthetic DB = %+v; want nil (exists but not oll/pll)", got)
	}
}

// --- 7. FindNextMove on a solved cube ----------------------------------------
//
// Reasoning from the source, before asserting:
//   - "oll": findOLLMove forms two branches.
//       * if !hasCross -> cross algo. On a solved cube the Up center and all four
//         Up edges are Yellow, so hasCross is TRUE and this branch is skipped.
//       * if hasCross && !isOLLComplete -> sune algo. On a solved cube the whole
//         Up face is Yellow, so isOLLComplete is TRUE and this branch is skipped.
//       The function falls through to `return nil`.
//   - "pll": findPLLMove guards on !isOLLComplete (false on solved) and then
//       cube.IsSolved() (true on solved), returning nil directly.
//
// Returning nil for both phases on a solved cube is sensible: the last layer has
// nothing left to do. (Note this is the one case the DB handles correctly.)

func TestFindNextMoveOnSolvedCube(t *testing.T) {
	db := NewSolvingDB()
	solved := NewCube(3)

	if got := db.FindNextMove(solved, "oll"); got != nil {
		t.Errorf(`FindNextMove(solved, "oll") = %q; want nil (OLL already complete)`, got.Name)
	}
	if got := db.FindNextMove(solved, "pll"); got != nil {
		t.Errorf(`FindNextMove(solved, "pll") = %q; want nil (cube already solved)`, got.Name)
	}
}

// --- 8. OLL branch routing + PINNED no-case-detection --------------------------

func TestFindNextMoveOLLBranchesAndPinNoCaseDetection(t *testing.T) {
	db := NewSolvingDB()

	// "F" and "R U R' U'" are structurally different scrambles, but hasCross is
	// false for both (a yellow Up-edge is displaced in each), so the code routes
	// BOTH to the single priority-1 Cross algorithm with a fixed move set.
	// PIN (suspected defect): findOLLMove has no dot/line/L discrimination at
	// all — !hasCross is its only cross branch (the "4-Look-OLL-Cross" description
	// claims it "handles dot, line, L-shape" but one fixed sequence is returned
	// for every non-cross state, with no state routing, and hasCross inspects only
	// the Up face, never whether the cross edges sit in their correct side slots).
	for _, seq := range []string{"F", "R U R' U'"} {
		st := apply(t, seq)
		got := db.FindNextMove(st, "oll")
		if got == nil || got.Name != "4-Look-OLL-Cross" {
			t.Errorf("non-cross state %q routed to %+v; want 4-Look-OLL-Cross", seq, got)
		}
	}

	// Sune branch: "R U R' U R U2 R'" (the Sune sequence) from solved leaves the
	// yellow cross intact but un-orients a corner, so hasCross==true && !isOLLComplete
	// -> priority-2 Sune algo. (Verified below via the same predicates the DB uses.)
	suneState := apply(t, "R U R' U R U2 R'")
	if !db.hasCross(suneState) {
		t.Fatalf("test precondition broken: expected the Sune sequence to preserve the cross")
	}
	if db.isOLLComplete(suneState) {
		t.Fatalf("test precondition broken: expected the Sune sequence to leave OLL incomplete")
	}
	s := db.FindNextMove(suneState, "oll")
	if s == nil || s.Name != "4-Look-OLL-Sune" {
		t.Errorf("cross-intact but OLL-incomplete state routed to %+v; want 4-Look-OLL-Sune", s)
	}
}

// --- Pin: findPLLMove is case-blind and U-Perm is unreachable -------------------

// PINNED SUSPECTED DEFECT: findPLLMove performs no PLL case detection. For ANY
// cube passing its two guards (OLL complete, not solved) it executes
//
//	for priority := 1; priority <= 2; priority++ {
//	    for _, algo := range algorithms { if algo.Priority == priority { return &algo } }
//	}
//
// which returns on the FIRST priority-1 hit — the T-Perm — every time. The
// priority-2 U-Perm is structurally unreachable, and the same T-Perm is offered
// regardless of the actual case. The in-code comment ("try T-perm first, then
// U-perm ... cycle through them") describes an intent the loop cannot fulfill.
//
// Demonstration: a solved cube rotated once by U. This is a legitimate PLL state
// whose correct fix is a single U' turn (or a U-perm); findPLLMove can only offer
// the T-Perm, which does not solve it, and a second call offers the T-Perm again
// — the DB can never make progress on this case.

func TestPinFindNextMovePLLAlwaysTPer(t *testing.T) {
	db := NewSolvingDB()

	state := apply(t, "U") // top still all Yellow (OLL complete), cube not solved
	if !db.isOLLComplete(state) {
		t.Fatalf("test precondition broken: U from solved should leave OLL complete")
	}
	if state.IsSolved() {
		t.Fatalf("test precondition broken: U from solved must not be solved")
	}

	first := db.FindNextMove(state, "pll")
	if first == nil {
		t.Fatalf("FindNextMove(pll) unexpectedly nil for an unsolved OLL-complete state")
	}
	if first.Name != "4-Look-PLL-T-Perm" {
		t.Errorf("expected case-blind T-Perm selection; got %q", first.Name)
	}

	// Applying the returned T-Perm must NOT solve a U case; the DB must also keep
	// offering the T-Perm afterwards (no U-Perm is ever reachable).
	progress := apply(t, "U")
	progress.ApplyMoves(first.Moves)
	if progress.IsSolved() {
		t.Errorf("PIN: T-Perm apparently solved a U-only case, contradicting cube theory — investigate")
	}
	again := db.FindNextMove(progress, "pll")
	if again == nil {
		t.Fatalf("expected the DB to keep returning a PLL move after T-Perm; got nil")
	}
	if again.Name != "4-Look-PLL-T-Perm" {
		t.Errorf("expected the DB to return T-Perm again; got %q", again.Name)
	}
}

// --- Pin: documented-but-unpopulated phases --------------------------------------

// PINNED: the SolvingDB struct comment documents phases "oll", "pll", "f2l",
// "cross", but load4LookLL only ever populates "oll" and "pll". FindNextMove for
// "f2l" and "cross" is therefore guaranteed to return nil — documented yet
// unreachable phases. (Also exercised by TestFindNextMoveRouting; kept here as an
// explicit pin.)

func TestPinDocumentedPhasesNeverPopulated(t *testing.T) {
	db := NewSolvingDB()
	for _, phase := range []string{"f2l", "cross"} {
		if got := db.FindNextMove(NewCube(3), phase); got != nil {
			t.Errorf("documented phase %q returned %+v; expected nil (never populated)", phase, got)
		}
	}
	for _, p := range db.GetPhases() {
		if p == "f2l" || p == "cross" {
			t.Errorf("phase %q should never appear in GetPhases(); got %v", p, db.GetPhases())
		}
	}
}

// --- Pin: Matches' size guard trusts Faces[0] alone --------------------------------

// PINNED SUSPECTED DEFECT: Matches validates size only against len(p.Faces[0]).
// If another face of the pattern has a different dimension, the guard is bypassed
// and the nested loops index out of range — a panic instead of a clean false. A
// predicate meant to *safely* test a state should never panic on malformed input.

func TestPinMalformedPatternPanics(t *testing.T) {
	pat := greyPattern(3)
	pat.Faces[1] = make([][]Color, 2) // shrink one non-Faces[0] face below the cube size
	pat.Faces[1][0] = make([]Color, 2)
	pat.Faces[1][1] = make([]Color, 2)
	for r := range pat.Faces[1] {
		for c := range pat.Faces[1][r] {
			pat.Faces[1][r][c] = Grey // keep it a wildcard so Matches doesn't early-return on color
		}
	}

	var recovered any
	func() {
		defer func() { recovered = recover() }()
		pat.Matches(NewCube(3))
	}()
	if recovered == nil {
		t.Logf("PIN: malformed pattern did NOT panic — Matches silently tolerated an inconsistent face")
	} else {
		t.Logf("PIN CONFIRMED: malformed pattern panicked (%v); a robust size guard would return false", recovered)
	}
}
