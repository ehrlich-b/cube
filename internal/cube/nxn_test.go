package cube

import (
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestNxNReductionContract(t *testing.T) {
	rng := rand.New(rand.NewSource(20261007))
	for _, size := range []int{2, 4, 5, 6, 7} {
		t.Run(fmt.Sprintf("%dx%d", size, size), func(t *testing.T) {
			for trial := 0; trial < 4; trial++ {
				c := NewCube(size)
				var scramble []Move
				for i := 0; i < 80; i++ {
					turns := rng.Intn(3) + 1
					scramble = append(scramble, Move{Face: Face(rng.Intn(6)), Layer: rng.Intn(size), Clockwise: turns == 1, Double: turns == 2})
				}
				if err := c.ApplyMoves(scramble); err != nil {
					t.Fatal(err)
				}
				before := c.clone()
				result, err := SolveNxN(c, KociembaOptions{TargetLength: 21, TimeLimit: time.Second})
				if err != nil {
					t.Fatalf("trial %d: %v; scramble %s", trial, err, FormatMoves(scramble))
				}
				if !facesEqual(c, before) {
					t.Fatal("solver mutated its input")
				}
				if result == nil || result.Steps != len(result.Solution) || len(result.Solution) == 0 {
					t.Fatal("invalid solver result")
				}
				if err := before.ApplyMoves(result.Solution); err != nil {
					t.Fatal(err)
				}
				if !before.IsSolved() || !nxnCenterMatched(before) {
					t.Fatal("solution did not produce a uniform center-matched cube")
				}
				t.Logf("trial %d: %d moves in %s", trial, result.Steps, result.Duration)
			}
		})
	}
}

func TestNxNParityFixtures(t *testing.T) {
	fixtures := []struct {
		name, scramble, invalid string
		oddWing                 bool
	}{
		{"OLL", "2R2 B2 U2 2L U2 2R' U2 2R U2 F2 2R F2 2L' B2 2R2", "flipped edges", true},
		{"PLL", "2R2 U2 2R2 Uw2 2R2 Uw2", "permutation parity", false},
	}
	for _, fixture := range fixtures {
		t.Run(fixture.name, func(t *testing.T) {
			c := NewCube(4)
			moves, err := ParseMoves(fixture.scramble)
			if err != nil {
				t.Fatal(err)
			}
			if err := c.ApplyMoves(moves); err != nil {
				t.Fatal(err)
			}
			// These are reduced parity states, rather than general scrambles:
			// their centers remain solved, but their sampled 3x3 is illegal.
			sample := NewCube(3)
			coordinates := [3]int{0, 1, 3}
			for f := range sample.Faces {
				for r := 0; r < 3; r++ {
					for col := 0; col < 3; col++ {
						sample.Faces[f][r][col] = c.Faces[f][coordinates[r]][coordinates[col]]
					}
				}
				for r := 1; r < 3; r++ {
					for col := 1; col < 3; col++ {
						if c.Faces[f][r][col] != NewCube(4).Faces[f][r][col] {
							t.Fatal("parity fixture disturbed centers")
						}
					}
				}
			}
			if err := Validate3x3(sample); err == nil || !strings.Contains(err.Error(), fixture.invalid) {
				t.Fatalf("fixture did not exhibit %s: %v", fixture.name, err)
			}
			tables, err := nxnTables(4)
			if err != nil {
				t.Fatal(err)
			}
			p, err := nxnWingPermutation(c, NewCube(3), tables.wings[0])
			if err != nil || (permutationParity(p) != 0) != fixture.oddWing {
				t.Fatalf("wrong wing parity: %v (%v)", p, err)
			}
			result, err := (&ReductionSolver{}).Solve(c)
			if err != nil {
				t.Fatal(err)
			}
			if err := c.ApplyMoves(result.Solution); err != nil || !c.IsSolved() || !nxnCenterMatched(c) {
				t.Fatalf("parity solution failed: %v", err)
			}
		})
	}
}

func TestNxNOddFixedCenterOrientation(t *testing.T) {
	for _, size := range []int{5, 7} {
		for _, scramble := range []string{"M E S x y z", "3R 3U 3F", "Rw Fw' Uw2 x'", "y2 2R 2U' 2F2"} {
			c := NewCube(size)
			moves, _ := ParseMoves(scramble)
			if err := c.ApplyMoves(moves); err != nil {
				t.Fatal(err)
			}
			result, err := (&ReductionSolver{}).Solve(c)
			if err != nil {
				t.Fatalf("%dx%d %s: %v", size, size, scramble, err)
			}
			if err := c.ApplyMoves(result.Solution); err != nil || !c.IsSolved() || !nxnCenterMatched(c) {
				t.Fatalf("fixed-center solve failed: %v", err)
			}
		}
	}
}

func TestNxNRejectsInvalidStatesWithoutMutation(t *testing.T) {
	fixtures := map[string]*Cube{"nil": nil, "unsupported": NewCube(8), "3x3": NewCube(3)}
	malformed := NewCube(4)
	malformed.Faces[Front][0] = malformed.Faces[Front][0][:3]
	fixtures["malformed"] = malformed
	wildcard := NewCube(4)
	wildcard.Faces[Front][0][0] = Grey
	fixtures["wildcard"] = wildcard
	mirrored := NewCube(2)
	mirrored.Faces[Front][0][1], mirrored.Faces[Right][0][0] = mirrored.Faces[Right][0][0], mirrored.Faces[Front][0][1]
	fixtures["mirrored corner"] = mirrored
	fixed := NewCube(5)
	fixed.Faces[Up][2][2], fixed.Faces[Front][2][2] = fixed.Faces[Front][2][2], fixed.Faces[Up][2][2]
	fixtures["invalid fixed centers"] = fixed
	centers := NewCube(5)
	centers.Faces[Up][1][1], centers.Faces[Front][1][2] = centers.Faces[Front][1][2], centers.Faces[Up][1][1]
	fixtures["wrong center orbit"] = centers
	wings := NewCube(4)
	wings.Faces[Up][3][1], wings.Faces[Front][0][1] = wings.Faces[Front][0][1], wings.Faces[Up][3][1]
	fixtures["impossible single wing flip"] = wings
	for name, c := range fixtures {
		t.Run(name, func(t *testing.T) {
			var before *Cube
			if c != nil {
				before = c.clone()
			}
			if result, err := (&ReductionSolver{}).Solve(c); err == nil || result != nil {
				t.Fatalf("invalid state returned success: %v", result)
			}
			if !reflect.DeepEqual(c, before) {
				t.Fatal("invalid input was mutated")
			}
		})
	}
	for _, size := range []int{2, 4, 5, 6, 7} {
		c := NewCube(size)
		result, err := (&ReductionSolver{}).Solve(c)
		if err != nil || result == nil || len(result.Solution) != 0 || !facesEqual(c, NewCube(size)) {
			t.Fatalf("solved %dx%d contract: %v (%v)", size, size, result, err)
		}
	}
}
