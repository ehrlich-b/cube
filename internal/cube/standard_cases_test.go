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

func TestDatabaseF2LDescriptionsMatchPhysicalCases(t *testing.T) {
	for _, a := range AlgorithmDatabase {
		if a.Dimension != 3 || !a.HasCategory("F2L") {
			continue
		}
		id := a.CaseID
		if !strings.HasPrefix(id, "F2L-") {
			for _, alias := range a.Aliases {
				if strings.HasPrefix(alias, "F2L-") {
					id = alias
					break
				}
			}
		}
		s, _, err := standardCaseState(id)
		if err != nil {
			t.Fatal(err)
		}
		c := cfopCube(s)
		// Read the target pair's actual stickers, independently of the prose
		// generator and candidate algorithm, in the standard fixture frame.
		for _, piece := range []struct {
			name, colors, marker string
			positions            [][]Coord
		}{
			{"corner", "BWR", "W", cornerPositions()},
			{"edge", "BR", "B", edgePositions()},
		} {
			found := false
			for _, coords := range piece.positions {
				var position, colors, markerFace string
				for _, p := range coords {
					position += p.Face.String()
					color := sticker(c, p).String()
					colors += color
					if color == piece.marker {
						markerFace = p.Face.String()
					}
				}
				if len(colors) != len(piece.colors) || strings.Trim(colors, piece.colors) != "" {
					continue
				}
				found = true
				layer := "FR slot"
				if coords[0].Face == Up {
					layer = "top layer"
				}
				colorName := "white"
				if piece.marker == "B" {
					colorName = "blue"
				}
				fact := fmt.Sprintf("%s in %s at %s (%s on %s)", piece.name, layer, position, colorName, markerFace)
				if !strings.Contains(a.Description, fact) || !strings.Contains(a.Recognition, fact) {
					t.Errorf("%s missing physical fact %q: description=%q, recognition=%q", id, fact, a.Description, a.Recognition)
				}
			}
			if !found {
				t.Fatal(id, "target", piece.name, "missing from physical fixture")
			}
		}
	}
}

func cornerPositions() [][]Coord {
	var positions [][]Coord
	for _, coords := range cornerFacelets {
		positions = append(positions, coords[:])
	}
	return positions
}

func edgePositions() [][]Coord {
	var positions [][]Coord
	for _, coords := range edgeFacelets {
		positions = append(positions, coords[:])
	}
	return positions
}

func TestNamedF2LCasesRejectWrongDescriptions(t *testing.T) {
	for _, id := range []string{"F2L-3", "F2L-4", "F2L-21", "F2L-22"} {
		t.Run(id, func(t *testing.T) {
			a := LookupAlgorithm(id)[0]
			for _, mutate := range []func(*Algorithm){
				func(a *Algorithm) { a.Description = strings.ReplaceAll(a.Description, "top layer", "FR slot") },
				func(a *Algorithm) { a.Description = strings.ReplaceAll(a.Description, "at URF", "at UFL") },
				func(a *Algorithm) { a.Description = strings.ReplaceAll(a.Description, "white on", "blue on") },
				func(a *Algorithm) { a.Recognition = "Edge in slot; corner in slot" },
			} {
				wrong := a
				mutate(&wrong)
				if err := VerifyAlgorithmCases(wrong); err == nil {
					t.Fatal("wrong physical description accepted", wrong.Description, wrong.Recognition)
				}
			}
		})
	}
}

func TestStandardF2LTextPositionsAndOrientations(t *testing.T) {
	// Literal expectations cover the reported defects and both pieces in-slot,
	// with every corner twist and edge flip. They are not generator output.
	for _, test := range []struct{ id, corner, edge string }{
		{"F2L-3", "corner in top layer at URF (white on F)", "edge in top layer at UL (blue on L)"},
		{"F2L-4", "corner in top layer at URF (white on R)", "edge in top layer at UB (blue on U)"},
		{"F2L-21", "corner in top layer at URF (white on U)", "edge in top layer at UL (blue on U)"},
		{"F2L-22", "corner in top layer at URF (white on U)", "edge in top layer at UB (blue on B)"},
		{"F2L-25", "corner in FR slot at DFR (white on D)", "edge in top layer at UR (blue on U)"},
		{"F2L-31", "corner in top layer at URF (white on U)", "edge in FR slot at FR (blue on R)"},
		{"F2L-37", "corner in FR slot at DFR (white on D)", "edge in FR slot at FR (blue on R)"},
		{"F2L-38", "corner in FR slot at DFR (white on F)", "edge in FR slot at FR (blue on F)"},
		{"F2L-39", "corner in FR slot at DFR (white on R)", "edge in FR slot at FR (blue on F)"},
	} {
		a := Algorithm{CaseID: test.id, Category: "F2L", Dimension: 3}
		description, recognition, err := StandardF2LText(a)
		if err != nil {
			t.Fatal(err)
		}
		for _, fact := range []string{test.corner, test.edge} {
			if !strings.Contains(description, fact) || !strings.Contains(recognition, fact) {
				t.Errorf("%s: missing %q from %q / %q", test.id, fact, description, recognition)
			}
		}
	}
	merged := Algorithm{CaseID: "TRIG-PAIR", Category: "Trigger", Categories: []string{"F2L"},
		Dimension: 3, Aliases: []string{"F2L-1", "F2L-1-CSV", "F2L-2"}}
	_, recognition, err := StandardF2LText(merged)
	if err != nil || strings.Count(recognition, "F2L-1 (") != 1 || !strings.Contains(recognition, "F2L-2 (") {
		t.Fatal("merged named cases missing or duplicated", recognition, err)
	}
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
