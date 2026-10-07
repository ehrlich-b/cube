package cube

import (
	"math/rand"
	"reflect"
	"testing"
	"time"
)

func TestCFOPUniformStates200(t *testing.T) {
	r := rand.New(rand.NewSource(2026100712))
	var sums, maxima [8]int
	var elapsed, longest time.Duration
	stageNames := []string{"Cross", "F2L 1", "F2L 2", "F2L 3", "F2L 4", "OLL", "PLL"}
	for n := 0; n < 200; n++ {
		s := identityCubie()
		r.Shuffle(8, func(i, j int) { s.cp[i], s.cp[j] = s.cp[j], s.cp[i] })
		r.Shuffle(12, func(i, j int) { s.ep[i], s.ep[j] = s.ep[j], s.ep[i] })
		cp, ep := make([]int, 8), make([]int, 12)
		for i, p := range s.cp {
			cp[i] = int(p)
		}
		for i, p := range s.ep {
			ep[i] = int(p)
		}
		if permutationParity(cp) != permutationParity(ep) {
			s.ep[0], s.ep[1] = s.ep[1], s.ep[0]
		}
		sum := 0
		for i := 0; i < 7; i++ {
			s.co[i] = uint8(r.Intn(3))
			sum += int(s.co[i])
		}
		s.co[7] = uint8((3 - sum%3) % 3)
		flip := uint8(0)
		for i := 0; i < 11; i++ {
			s.eo[i] = uint8(r.Intn(2))
			flip ^= s.eo[i]
		}
		s.eo[11] = flip
		c := cubeFromCoordinates(s)
		before := c.String()
		if err := Validate3x3(c); err != nil {
			t.Fatal(err)
		}
		result, err := (&CFOPSolver{}).Solve(c)
		if err != nil {
			t.Fatalf("uniform state %d: %v", n, err)
		}
		if c.String() != before {
			t.Fatal("solver mutated input")
		}
		if len(result.Stages) != 7 {
			t.Fatal("missing stages")
		}
		var moves []Move
		protected := uint8(0)
		for i, stage := range result.Stages {
			if stage.Name != stageNames[i] || len(stage.Cases) == 0 {
				t.Fatal("invalid stage metadata", stage)
			}
			moves = append(moves, stage.Moves...)
			c.ApplyMoves(stage.Moves)
			if c.String() != stage.After.String() {
				t.Fatalf("state %d %s snapshot differs from replay", n, stage.Name)
			}
			if !WhiteCrossSolved(c) {
				t.Fatalf("state %d %s broke the cross", n, stage.Name)
			}
			if i >= 1 && i <= 4 {
				slots := cfopSlots(c)
				if slots&protected != protected {
					t.Fatal("F2L broke a solved pair")
				}
				protected = slots
			}
			if i >= 4 && cfopSlots(c) != 15 {
				t.Fatal("F2L incomplete")
			}
			if i >= 5 && !OLLSolved(c) {
				t.Fatal("OLL incomplete")
			}
			turns := TurnCount(stage.Moves)
			sums[i] += turns
			maxima[i] = max(maxima[i], turns)
		}
		if !c.IsSolved() || !reflect.DeepEqual(moves, result.Solution) || result.Steps != len(moves) {
			t.Fatalf("state %d full solver contract failed", n)
		}
		turns := TurnCount(moves)
		sums[7] += turns
		maxima[7] = max(maxima[7], turns)
		elapsed += result.Duration
		longest = max(longest, result.Duration)
	}
	for i, name := range append(stageNames, "Total") {
		t.Logf("%s: mean %.3f max %d turns", name, float64(sums[i])/200, maxima[i])
	}
	t.Logf("200 uniform random states: total %v, mean %v, max %v (includes first-use setup)", elapsed, elapsed/200, longest)
}

func TestCFOPFramesValidationAndSolved(t *testing.T) {
	for _, text := range []string{"", "x y R U F2 L' B", "M E S R U", "Rw Fw Uw2 L'"} {
		c := NewCube(3)
		moves, _ := ParseMoves(text)
		c.ApplyMoves(moves)
		result, err := (&CFOPSolver{}).Solve(c)
		if err != nil {
			t.Fatal(text, err)
		}
		c.ApplyMoves(result.Solution)
		if !c.IsSolved() {
			t.Fatal(text, "did not solve")
		}
		if text == "" && len(result.Solution) != 0 {
			t.Fatal("solved cube should skip every stage")
		}
	}
	for _, c := range []*Cube{nil, NewCube(2), NewCube(4)} {
		if _, err := (&CFOPSolver{}).Solve(c); err == nil {
			t.Fatal("unsupported cube accepted")
		}
	}
	c := NewCube(3)
	a, b := edgeFacelets[0][0], edgeFacelets[0][1]
	c.Faces[a.Face][a.Row][a.Col], c.Faces[b.Face][b.Row][b.Col] = sticker(c, b), sticker(c, a)
	if _, err := (&CFOPSolver{}).Solve(c); err == nil {
		t.Fatal("single flipped edge accepted")
	}
}

func TestCFOPLastLayerCoverageAndReplay(t *testing.T) {
	cfopOnce.Do(loadCFOPDatabase)
	if cfopDB.err != nil {
		t.Fatal(cfopDB.err)
	}
	if len(cfopDB.oll) != 216 || len(cfopDB.pll) != 288 {
		t.Fatal("incomplete LL coverage")
	}
	for _, cases := range [][]cfopCase{cfopDB.oll, cfopDB.pll} {
		for _, candidate := range cases {
			c := cubeFromCoordinates(candidate.state)
			if !matchesStickerPattern(c, candidate.pattern) {
				t.Fatal("generated recognition pattern mismatch")
			}
			c.ApplyMoves(candidate.moves)
			if cfopSlots(c) != 15 || !OLLSolved(c) {
				t.Fatal("LL algorithm failed checkpoint")
			}
			if len(cases) == 288 && !c.IsSolved() {
				t.Fatal("PLL failed full solve")
			}
		}
	}
}

func TestFixedFrameMovesPreserveEffects(t *testing.T) {
	for _, a := range AlgorithmDatabase {
		if a.Dimension != 3 {
			continue
		}
		moves, _ := ParseMoves(a.Moves)
		for yaw := 0; yaw < 4; yaw++ {
			seq := append(append(yawMoves(yaw), moves...), inverseSequence(yawMoves(yaw))...)
			original, fixed := NewCube(3), NewCube(3)
			original.ApplyMoves(seq)
			fixed.ApplyMoves(fixedFrameMoves(seq))
			if original.String() != fixed.String() {
				t.Fatal(a.CaseID, "frame rewrite changed effect")
			}
		}
	}
}

func TestImportedRecognitionPatterns(t *testing.T) {
	for _, a := range AlgorithmDatabase {
		if a.Dimension != 3 {
			continue
		}
		c, err := readAlgorithmPattern(a.Pattern)
		if err != nil {
			t.Fatal(a.CaseID, err)
		}
		if err := Validate3x3(c); err != nil {
			t.Fatal(a.CaseID, err)
		}
		moves, _ := ParseMoves(a.Moves)
		c.ApplyMoves(moves)
		if c.String() != NewCube(3).String() {
			t.Fatal(a.CaseID, "imported pattern does not verify")
		}
	}
	for _, text := range []string{"", "WB|W9/R9/B9/Y9/O9/G9", "YB|Y4/R4/B4/W4/O4/G4", "YB|Y10/R9/B9/W9/O9/G9", "YB|?9/R9/B9/W9/O9/G9", "YB|Y9/R9/B9/W9/O9"} {
		if _, err := readAlgorithmPattern(text); err == nil {
			t.Fatal("invalid generated pattern accepted", text)
		}
	}
}

func TestCFOPResultDoesNotExposeCache(t *testing.T) {
	c := NewCube(3)
	scramble, _ := ParseMoves("R2 U F' D B2 L' U2 F R' D2 L B' U R2 F2 D' L2 U' B R")
	c.ApplyMoves(scramble)
	first, err := (&CFOPSolver{}).Solve(c)
	if err != nil {
		t.Fatal(err)
	}
	want := append([]Move(nil), first.Solution...)
	for _, stage := range first.Stages {
		for i := range stage.Cases {
			stage.Cases[i] = "modified"
		}
		for i := range stage.Moves {
			stage.Moves[i] = Move{Face: Down, Clockwise: true}
		}
	}
	second, err := (&CFOPSolver{}).Solve(c)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first.Solution, want) || !reflect.DeepEqual(second.Solution, want) {
		t.Fatal("mutating returned stages changed the solution cache")
	}
	for _, stage := range second.Stages {
		for _, name := range stage.Cases {
			if name == "modified" {
				t.Fatal("returned names exposed the cache")
			}
		}
	}
}
