package cfen

// cfen_test.go pins down the CFEN verification backbone the solver will rely on:
// the canonical solved string, lossless round-tripping between a cube and its
// CFEN form, and wildcard pattern matching. The default YB orientation is the
// one every command uses today and is exercised most heavily; the other three
// supported orientations are checked for round-trip consistency.

import (
	"testing"

	"github.com/ehrlich-b/cube/internal/cube"
)

// cubesEqual reports whether two cubes have identical sticker layouts.
func cubesEqual(a, b *cube.Cube) bool {
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

func scrambledCube(t *testing.T, size int, scramble string) *cube.Cube {
	t.Helper()
	c := cube.NewCube(size)
	if scramble != "" {
		moves, err := cube.ParseMoves(scramble)
		if err != nil {
			t.Fatalf("parse scramble %q: %v", scramble, err)
		}
		c.ApplyMoves(moves)
	}
	return c
}

// The solved cube has exactly one canonical CFEN string in the default YB
// orientation. If this changes, either the orientation convention or the face
// ordering has drifted, and every stored pattern is now suspect.
func TestInvariant_SolvedCubeCanonicalCFEN(t *testing.T) {
	cases := map[int]string{
		2: "YB|Y4/R4/B4/W4/O4/G4",
		3: "YB|Y9/R9/B9/W9/O9/G9",
	}
	for size, want := range cases {
		got, err := GenerateCFEN(cube.NewCube(size))
		if err != nil {
			t.Fatalf("size %d: GenerateCFEN: %v", size, err)
		}
		if got != want {
			t.Errorf("size %d: solved CFEN = %q, want %q", size, got, want)
		}
	}
}

// cube -> CFEN -> cube must be lossless in the default YB orientation, for the
// solved state and arbitrary scrambles. This is the contract that lets verify
// trust a parsed start state.
func TestInvariant_CFENRoundTripYB(t *testing.T) {
	scrambles := []string{
		"",
		"R",
		"R U R' U'",
		"R U F2 L' B",
		"F R U' R' U' R U R' F'",
	}
	for _, scramble := range scrambles {
		original := scrambledCube(t, 3, scramble)

		cfenStr, err := GenerateCFEN(original)
		if err != nil {
			t.Fatalf("scramble %q: GenerateCFEN: %v", scramble, err)
		}
		state, err := ParseCFEN(cfenStr)
		if err != nil {
			t.Fatalf("scramble %q: ParseCFEN(%q): %v", scramble, cfenStr, err)
		}
		restored, err := state.ToCube()
		if err != nil {
			t.Fatalf("scramble %q: ToCube: %v", scramble, err)
		}
		if !cubesEqual(original, restored) {
			t.Errorf("scramble %q: cube -> CFEN -> cube changed the cube (CFEN was %q)", scramble, cfenStr)
		}
	}
}

// FromCube -> ToCube must be the identity for every supported orientation, not
// just YB. (The forward and reverse face mappings must agree for this to hold.)
func TestInvariant_CFENRoundTripAllOrientations(t *testing.T) {
	orientations := []struct {
		name string
		o    CFENOrientation
	}{
		{"YB", CFENOrientation{Up: cube.Yellow, Front: cube.Blue}},
		{"WB", CFENOrientation{Up: cube.White, Front: cube.Blue}},
		{"YG", CFENOrientation{Up: cube.Yellow, Front: cube.Green}},
		{"WG", CFENOrientation{Up: cube.White, Front: cube.Green}},
	}
	scrambles := []string{"", "R U R' U'", "R U F2 L' B"}

	for _, oc := range orientations {
		for _, scramble := range scrambles {
			original := scrambledCube(t, 3, scramble)

			state, err := FromCube(original, oc.o)
			if err != nil {
				t.Fatalf("%s / %q: FromCube: %v", oc.name, scramble, err)
			}
			restored, err := state.ToCube()
			if err != nil {
				t.Fatalf("%s / %q: ToCube: %v", oc.name, scramble, err)
			}
			if !cubesEqual(original, restored) {
				t.Errorf("%s / %q: FromCube -> ToCube was not the identity", oc.name, scramble)
			}
		}
	}
}

// Wildcard ('?', stored as Grey) positions must match any color, while concrete
// colors must match exactly. This is what makes patterns like "yellow on top,
// rest anything" usable for last-layer recognition.
func TestInvariant_MatchesCubeWildcards(t *testing.T) {
	mustMatch := func(t *testing.T, pattern string, c *cube.Cube, want bool) {
		t.Helper()
		state, err := ParseCFEN(pattern)
		if err != nil {
			t.Fatalf("ParseCFEN(%q): %v", pattern, err)
		}
		got, err := state.MatchesCube(c)
		if err != nil {
			t.Fatalf("MatchesCube against %q: %v", pattern, err)
		}
		if got != want {
			t.Errorf("pattern %q matched=%v, want %v", pattern, got, want)
		}
	}

	solved := cube.NewCube(3)
	afterR := scrambledCube(t, 3, "R")   // disturbs the U face
	afterY := scrambledCube(t, 3, "y")   // whole-cube spin: yellow stays on top
	afterX2 := scrambledCube(t, 3, "x2") // flips top/bottom: white now on top

	const yellowTop = "YB|Y9/?9/?9/?9/?9/?9"
	const solvedPat = "YB|Y9/R9/B9/W9/O9/G9"

	mustMatch(t, solvedPat, solved, true)   // exact solved pattern matches solved cube
	mustMatch(t, solvedPat, afterR, false)  // a real turn no longer matches solved
	mustMatch(t, yellowTop, solved, true)   // yellow-top wildcard matches solved
	mustMatch(t, yellowTop, afterY, true)   // ...and still matches after a y spin
	mustMatch(t, yellowTop, afterX2, false) // ...but not when white is on top
}

// Mirrors what `cube verify` does: parse a target pattern, apply an algorithm to
// the start state, and check the result against the pattern. A round-trip
// algorithm (sexy move x6 = identity) must land on solved; a single sexy move
// must not.
func TestInvariant_VerifyAlgorithmSemantics(t *testing.T) {
	target, err := ParseCFEN("YB|Y9/R9/B9/W9/O9/G9")
	if err != nil {
		t.Fatalf("ParseCFEN: %v", err)
	}

	identity := scrambledCube(t, 3, "R U R' U' R U R' U' R U R' U' R U R' U' R U R' U' R U R' U'")
	if ok, _ := target.MatchesCube(identity); !ok {
		t.Error("sexy move applied 6 times should return to solved")
	}

	single := scrambledCube(t, 3, "R U R' U'")
	if ok, _ := target.MatchesCube(single); ok {
		t.Error("a single sexy move should not match the solved state")
	}
}
