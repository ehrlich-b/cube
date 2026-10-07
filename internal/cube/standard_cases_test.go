package cube

import (
	"fmt"
	"strings"
	"testing"
)

func TestStandardFixturesArePhysical(t *testing.T) {
	var ids []string
	for i := 1; i <= 57; i++ {
		ids = append(ids, fmt.Sprintf("OLL-%d", i))
	}
	for i := 1; i <= 41; i++ {
		ids = append(ids, fmt.Sprintf("F2L-%d", i))
	}
	for id := range standardPLL {
		ids = append(ids, "PLL-"+id)
	}
	for _, id := range ids {
		s, _, err := standardCaseState(id)
		if err != nil {
			t.Fatal(err)
		}
		if err := Validate3x3(cfopCube(s)); err != nil {
			t.Fatal(id, err)
		}
	}
}

func TestDatabaseStandardCases(t *testing.T) {
	counts := map[string]int{}
	for _, a := range AlgorithmDatabase {
		if err := VerifyAlgorithmCases(a); err != nil {
			t.Error(a.CaseID, err)
		}
		if a.Dimension != 3 {
			continue
		}
		for _, category := range []string{"OLL", "PLL", "F2L"} {
			if a.HasCategory(category) {
				counts[category]++
			}
		}
	}
	t.Logf("independent physical database checks: %v", counts)
}

func TestNamedCasesRejectWrongAlgorithms(t *testing.T) {
	for _, test := range []struct{ id, wrong, right string }{
		{"PLL-T", "R U R' F' R U R' U' R' F R2 U' R'", "R U R' U' R' F R2 U' R' U' R U R' F'"},
		{"OLL-12", "R U R' U' R' F R F'", "M' R' U' R U' R' U2 R U' M"},
		{"F2L-1", "Rw U R' U' Rw' F R F'", "U R U' R'"},
	} {
		if err := VerifyStandardCase(Algorithm{Moves: test.wrong}, test.id); err == nil {
			t.Fatal("wrong case accepted", test.id)
		}
		if err := VerifyStandardCase(Algorithm{Moves: test.right}, test.id); err != nil {
			t.Fatal(test.id, err)
		}
	}
}

func TestCFOPPLLLabels(t *testing.T) {
	for _, name := range []string{"T", "Jb"} {
		s, _, _ := standardCaseState("PLL-" + name)
		result, err := (&CFOPSolver{}).Solve(cfopCube(s))
		if err != nil {
			t.Fatal(err)
		}
		labels := strings.Join(result.Stages[6].Cases, "; ")
		if !strings.Contains(labels, "PLL-"+name+" ·") {
			t.Fatalf("%s fixture mislabeled: %s", name, labels)
		}
	}
	c := NewCube(3)
	moves, _ := ParseMoves("R U R2 F' R U R U' R' F R U' R'")
	c.ApplyMoves(moves)
	result, err := (&CFOPSolver{}).Solve(c)
	if err != nil {
		t.Fatal(err)
	}
	if label := strings.Join(result.Stages[6].Cases, "; "); !strings.Contains(label, "PLL-Jb ·") || strings.Contains(label, "PLL-T ·") {
		t.Fatal("reported Jb scramble mislabeled", label)
	}
}

func TestFourByFourOLLParityFixture(t *testing.T) {
	// A flipped UF dedge, assembled by exchanging the two wing sticker
	// colors directly; no database pattern or inverse algorithm is involved.
	c := NewCube(4)
	for col := 1; col <= 2; col++ {
		c.Faces[Up][3][col], c.Faces[Front][0][col] = Blue, Yellow
	}
	moves, _ := ParseMoves(LookupAlgorithm("4x4-OLL-PARITY")[0].Moves)
	c.ApplyMoves(moves)
	if !c.IsSolved() {
		t.Fatal("4x4 OLL parity does not solve an independently flipped UF dedge", c.String())
	}
}

func TestFourByFourPLLParityFixture(t *testing.T) {
	// Two opposite dedges exchanged independently of the stored pattern.
	// The standard six-turn algorithm leaves a final U2 alignment.
	c := NewCube(4)
	for col := 1; col <= 2; col++ {
		c.Faces[Front][0][col], c.Faces[Back][0][col] = Green, Blue
	}
	moves, _ := ParseMoves(LookupAlgorithm("4x4-PLL-PARITY")[0].Moves)
	c.ApplyMoves(moves)
	auf, _ := ParseMove("U2")
	c.ApplyMove(auf)
	if !c.IsSolved() {
		t.Fatal("4x4 PLL parity does not solve independently exchanged UF/UB dedges", c.String())
	}
}

func TestFourByFourParityCenters(t *testing.T) {
	for _, id := range []string{"4x4-OLL-PARITY", "4x4-PLL-PARITY"} {
		a := LookupAlgorithm(id)[0]
		c := NewCube(4)
		moves, _ := ParseMoves(a.Moves)
		c.ApplyMoves(moves)
		for f := range c.Faces {
			for r := 1; r <= 2; r++ {
				for col := 1; col <= 2; col++ {
					if c.Faces[f][r][col] != NewCube(4).Faces[f][r][col] {
						t.Errorf("%s scrambles %s center (%d,%d)", id, Face(f), r, col)
					}
				}
			}
		}
	}
}
