package cube

import (
	"testing"
	"time"
)

// Run alone with -count=1 to measure fresh-process setup independently of search.
func TestCFOPSetupTiming(t *testing.T) {
	started := time.Now()
	searchEdgePatterns()
	edges := time.Since(started)
	started = time.Now()
	cfopOnce.Do(loadCFOPDatabase)
	if cfopDB.err != nil {
		t.Fatal(cfopDB.err)
	}
	t.Logf("CFOP setup: edges %v, database %v", edges, time.Since(started))
}

func TestFixedFrameAllocationBudget(t *testing.T) {
	moves, _ := ParseMoves("y R U R' U' R' F R2 U' R' U' R U R' F' y'")
	allocs := testing.AllocsPerRun(5, func() { fixedFrameMoves(moves) })
	if allocs > 300 {
		t.Fatalf("fixed-frame rewrite allocates %.0f objects; budget 300", allocs)
	}
}

// Original implementation retained as an exact output oracle for the faster rewrite.
func referenceFixedFrameMoves(moves []Move) []Move {
	frame := NewCube(3)
	var rotations, result []Move
	for _, m := range moves {
		if m.Rotation != NoRotation {
			frame.ApplyMove(m)
			rotations = append(rotations, m)
			continue
		}
		probe := frame.clone()
		probe.ApplyMove(m)
		probe.ApplyMoves(inverseSequence(rotations))
		found := false
		for _, candidate := range referenceCFOPMoveAlphabet() {
			test := NewCube(3)
			test.ApplyMove(candidate)
			if test.String() == probe.String() {
				result = append(result, candidate)
				found = true
				break
			}
		}
		if !found {
			return append([]Move(nil), moves...)
		}
	}
	_, restore, _ := canonical3x3(frame)
	result = append(result, inverseSequence(restore)...)
	return OptimizeMoves(result)
}

func referenceCFOPMoveAlphabet() []Move {
	result := append([]Move(nil), coordinateMoves[:]...)
	extra, _ := ParseMoves("M M2 M' E E2 E' S S2 S' Rw Rw2 Rw' Lw Lw2 Lw' Uw Uw2 Uw' Dw Dw2 Dw' Fw Fw2 Fw' Bw Bw2 Bw'")
	return append(result, extra...)
}

func TestFixedFrameRewriteMatchesReference(t *testing.T) {
	for _, a := range AlgorithmDatabase {
		if a.Dimension != 3 {
			continue
		}
		moves, _ := ParseMoves(a.Moves)
		for yaw := 0; yaw < 4; yaw++ {
			seq := append(append(yawMoves(yaw), moves...), inverseSequence(yawMoves(yaw))...)
			got, want := fixedFrameMoves(seq), referenceFixedFrameMoves(seq)
			if FormatMoves(got) != FormatMoves(want) {
				t.Fatalf("%s yaw %d: %s; want %s", a.CaseID, yaw, FormatMoves(got), FormatMoves(want))
			}
			original, fixed := geometryLabels(3), geometryLabels(3)
			original.ApplyMoves(seq)
			fixed.ApplyMoves(got)
			assertGeometryEqual(t, fixed, original)
		}
	}
}
