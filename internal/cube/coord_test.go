package cube

import (
	"math/rand"
	"strings"
	"testing"
)

// coord_test.go exercises the cubie-level representation of coord.go.
//
// The existing sticker engine is the oracle: almost every test is DIFFERENTIAL,
// comparing our cubie-level ApplyMoves against the engine (a *Cube advanced by
// ApplyMoves and then converted with ToCubieState). No test below hard-codes a
// move-table row and then "verifies" it against a copy of itself.

// randomScramble3 returns a deterministic pseudo-random sequence of n face turns
// (18-move alphabet only), using a caller-supplied generator. This is legal on
// a 3x3 cube and matches the move set the cubie representation understands.
func randomScramble3(r *rand.Rand, n int) []Move {
	faces := []string{"R", "L", "U", "D", "F", "B"}
	mods := []string{"", "'", "2"}
	tokens := make([]string, n)
	for i := 0; i < n; i++ {
		tokens[i] = faces[r.Intn(len(faces))] + mods[r.Intn(len(mods))]
	}
	moves, err := ParseMoves(strings.Join(tokens, " "))
	if err != nil {
		panic("randomScramble3 produced unparseable notation: " + strings.Join(tokens, " "))
	}
	return moves
}

// allSingleMoves returns the 18 face turns (6 faces x 3 variants) in a fixed
// order so an individual sub-test can name a single bad table entry.
func allSingleMoves() []struct {
	name string
	mv   Move
} {
	moves := []struct {
		name string
		mv   Move
	}{}
	names := []string{"R", "L", "U", "D", "F", "B"}
	mods := []string{"", "'", "2"}
	for _, n := range names {
		for _, m := range mods {
			parsed, err := ParseMoves(n + m)
			if err != nil {
				panic("allSingleMoves: " + err.Error())
			}
			moves = append(moves, struct {
				name string
				mv   Move
			}{n + m, parsed[0]})
		}
	}
	return moves
}

func cubieStateEqual(a, b CubieState) bool {
	return a == b
}

func TestToCubieStateRoundTrip(t *testing.T) {
	// 1. Round trip: a solved cube converts to the solved cubie state.
	c := NewCube(3)
	got, err := ToCubieState(c)
	if err != nil {
		t.Fatalf("ToCubieState(solved) returned error: %v", err)
	}
	want := SolvedCubieState()
	if !cubieStateEqual(got, want) {
		t.Fatalf("ToCubieState(solved) = %+v, want %+v", got, want)
	}
	if !got.IsSolved() {
		t.Fatal("ToCubieState(solved).IsSolved() = false, want true")
	}
}

// TestCubieDifferentialRandom is the central differential property: applying a
// scramble to a *Cube and converting with ToCubieState must equal applying the
// same scramble to SolvedCubieState() with the cubie-level ApplyMoves.
//
// The RNG seed is FIXED so any failure reproduces exactly. This test proves
// every move table row is right, because a single wrong row makes the two
// independent paths diverge.
func TestCubieDifferentialRandom(t *testing.T) {
	r := rand.New(rand.NewSource(20260825)) // fixed seed: failures reproduce
	for i := 0; i < 500; i++ {
		scramble := randomScramble3(r, 1+r.Intn(25)) // length 1..25
		c := NewCube(3)
		c.ApplyMoves(scramble)
		engine, err := ToCubieState(c)
		if err != nil {
			t.Fatalf("ToCubieState returned error on scramble #%d (%v): %v", i, scramble, err)
		}
		want := SolvedCubieState()
		for _, mv := range scramble {
			want = want.ApplyMove(mv)
		}
		if !cubieStateEqual(engine, want) {
			t.Fatalf("differential mismatch on scramble #%d: %v\nengine (via *Cube): %+v\ncubie  (ApplyMoves): %+v",
				i, scramble, engine, want)
		}
		// 6. The two classic invariants hold after every single scramble too:
		if cornerSumMod3(want) != 0 {
			t.Fatalf("corner orientation sum = %d mod 3 != 0 on scramble #%d (%v)", cornerSumMod3(want), i, scramble)
		}
		if edgeSumMod2(want) != 0 {
			t.Fatalf("edge orientation sum = %d mod 2 != 0 on scramble #%d (%v)", edgeSumMod2(want), i, scramble)
		}
	}
}

// TestCubieDifferentialSingleMoves checks each of the 18 face turns individually,
// one sub-test per move, so a single bad table entry names itself.
func TestCubieDifferentialSingleMoves(t *testing.T) {
	for _, sm := range allSingleMoves() {
		sm := sm
		t.Run(sm.name, func(t *testing.T) {
			c := NewCube(3)
			c.ApplyMove(sm.mv)
			engine, err := ToCubieState(c)
			if err != nil {
				t.Fatalf("ToCubieState returned error: %v", err)
			}
			want := SolvedCubieState().ApplyMove(sm.mv)
			if !cubieStateEqual(engine, want) {
				t.Fatalf("differential mismatch for %s (%+v)\nengine (via *Cube): %+v\ncubie  (ApplyMoves): %+v",
					sm.name, sm.mv, engine, want)
			}
		})
	}
}

// TestCubeFaceTurnOrder checks that any quarter turn applied four times returns
// to the start, and any 180-degree turn applied twice does too.
func TestCubeFaceTurnOrder(t *testing.T) {
	names := []string{"R", "L", "U", "D", "F", "B"}
	for _, n := range names {
		mv, err := ParseMoves(n)
		if err != nil {
			t.Fatalf("ParseMoves(%s): %v", n, err)
		}
		qt := mv[0]

		s := SolvedCubieState()
		for i := 0; i < 4; i++ {
			s = s.ApplyMove(qt)
		}
		if !s.IsSolved() {
			t.Errorf("%s applied four times is not the solved state: %+v", n, s)
		}

		d2, err := ParseMoves(n + "2")
		if err != nil {
			t.Fatalf("ParseMoves(%s2): %v", n, err)
		}
		twice := SolvedCubieState().ApplyMove(d2[0]).ApplyMove(d2[0])
		if !twice.IsSolved() {
			t.Errorf("%s2 applied twice is not the solved state: %+v", n, twice)
		}
	}
}

// TestCubieInverse checks that applying a random sequence and then its inverse
// returns to the solved state.
func TestCubieInverse(t *testing.T) {
	r := rand.New(rand.NewSource(77))
	for i := 0; i < 300; i++ {
		scramble := randomScramble3(r, 1+r.Intn(25))
		inv := invertMoves(scramble)
		s := SolvedCubieState()
		for _, mv := range append(append([]Move{}, scramble...), inv...) {
			s = s.ApplyMove(mv)
		}
		if !s.IsSolved() {
			t.Fatalf("scramble + inverse is not solved on iteration #%d (%v)", i, scramble)
		}
	}
}

func cornerSumMod3(s CubieState) int {
	sum := 0
	for i := 0; i < 8; i++ {
		sum += s.CO[i]
	}
	return sum % 3
}

func edgeSumMod2(s CubieState) int {
	sum := 0
	for i := 0; i < 12; i++ {
		sum += s.EO[i]
	}
	return sum % 2
}

// TestCubieOrientationInvariants asserts the two classic invariants — corner
// orientation sums to 0 mod 3 and edge orientation to 0 mod 2 — over many
// random scrambles. This class of bug (a corner-orientation update that drops
// the mod 3, for instance) is invisible to the differential test when the same
// mistake is shared by extraction and application, so it gets dedicated tests.
func TestCubieOrientationInvariants(t *testing.T) {
	r := rand.New(rand.NewSource(424242)) // fixed seed: failures reproduce
	for i := 0; i < 500; i++ {
		s := SolvedCubieState().ApplyMoves(randomScramble3(r, 1+r.Intn(25)))
		if got := cornerSumMod3(s); got != 0 {
			t.Fatalf("corner orientation sum = %d (mod 3) on scramble #%d", got, i)
		}
		if got := edgeSumMod2(s); got != 0 {
			t.Fatalf("edge orientation sum = %d (mod 2) on scramble #%d", got, i)
		}
	}
}

// permutationParity returns 0 for an even permutation and 1 for an odd one,
// computed directly from the cycle decomposition (parity = len - #cycles).
func permutationParity(p []int) int {
	visited := make([]bool, len(p))
	cycles := 0
	for i := 0; i < len(p); i++ {
		if visited[i] {
			continue
		}
		cycles++
		cur := i
		for !visited[cur] {
			visited[cur] = true
			cur = p[cur]
		}
	}
	return (len(p) - cycles) % 2
}

// TestCubieParityAgreement checks that the permutation parity of CP equals that
// of EP for any state reachable by face turns.
func TestCubieParityAgreement(t *testing.T) {
	r := rand.New(rand.NewSource(314159))
	for i := 0; i < 400; i++ {
		s := SolvedCubieState().ApplyMoves(randomScramble3(r, 1+r.Intn(25)))
		cp := permutationParity(s.CP[:])
		ep := permutationParity(s.EP[:])
		if cp != ep {
			t.Fatalf("parity mismatch on scramble #%d: CP parity %d, EP parity %d (%+v)", i, cp, ep, s)
		}
	}
}

// TestToCubieStateNonThree verifies that a non-3x3 cube yields an error and does
// not panic.
func TestToCubieStateNonThree(t *testing.T) {
	for _, size := range []int{2, 4, 5} {
		if _, err := ToCubieState(NewCube(size)); err == nil {
			t.Errorf("ToCubieState(NewCube(%d)) returned nil error, want an error", size)
		}
	}
}

// TestCubieIsSolvedDetectsEachComponent checks that IsSolved rejects a solved
// state once any one of the four arrays is perturbed.
func TestCubieIsSolvedDetectsEachComponent(t *testing.T) {
	cases := map[string]func(*CubieState){
		"corner permutation": func(c *CubieState) { c.CP[0], c.CP[1] = c.CP[1], c.CP[0] },
		"corner orientation": func(c *CubieState) { c.CO[3] = (c.CO[3] + 1) % 3 },
		"edge permutation":   func(c *CubieState) { c.EP[0], c.EP[1] = c.EP[1], c.EP[0] },
		"edge orientation":   func(c *CubieState) { c.EO[5] ^= 1 },
	}
	for name, mutate := range cases {
		c := SolvedCubieState()
		mutate(&c)
		if c.IsSolved() {
			t.Errorf("IsSolved() = true after perturbing the %s", name)
		}
	}
}
