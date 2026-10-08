package cube

import (
	"math/rand"
	"os"
	"testing"
	"time"
)

func TestPhaseOneSymmetryCoordinates(t *testing.T) {
	db := phase1SymmetryCoordinates(solverTables())
	r := rand.New(rand.NewSource(2026100715))
	sym := phase1Symmetries()
	for i, rotation := range sym {
		for m, move := range cubieMoves {
			conjugated := rotation.inverse().mul(move).mul(rotation)
			found := false
			for _, candidate := range cubieMoves {
				found = found || conjugated == candidate
			}
			if !found {
				t.Fatalf("symmetry %d does not preserve move %d", i, m)
			}
		}
	}
	for trial := 0; trial < 1000; trial++ {
		state := uniformCubie(r)
		code := int(db.code[state.twist()])
		rotation := sym[code&7]
		next := rotation.inverse().mul(state).mul(rotation)
		if next.twist() != int(db.reps[code/8]) || db.conjugate(state.flip()*495+state.slice(), code&7) != next.flip()*495+next.slice() {
			t.Fatal("symmetry coordinates disagree with full cubie conjugation")
		}
		for i, rotation := range sym {
			next := rotation.inverse().mul(state).mul(rotation)
			if db.conjugate(state.flip()*495+state.slice(), i) != next.flip()*495+next.slice() {
				t.Fatal("flip/slice conjugation depends on omitted permutations")
			}
		}
	}
	t.Logf("%d twist classes, %d pruning states, %d packed bytes", len(db.reps), len(db.reps)*flipSliceStates, len(db.reps)*flipSliceStates/2)
}

func TestPhaseOneLargeTableOracle(t *testing.T) {
	if os.Getenv("CUBE_PHASE1_TABLES") != "1" {
		t.Skip("make test-phase1-tables")
	}
	tables := solverTables()
	started := time.Now()
	db := phase1PatternTables(tables)
	if db == nil {
		t.Fatal("symmetry-reduced phase-one table failed")
	}
	t.Logf("phase-one generation/load: %v; %d classes; %d packed bytes", time.Since(started), len(db.reps), len(db.distance))
	r := rand.New(rand.NewSource(2026100716))
	for trial := 0; trial < 1000; trial++ {
		state := identityCubie()
		for depth := 0; depth < 14; depth++ {
			h := db.bound(state.twist(), state.flip(), state.slice())
			if h > depth {
				t.Fatal("inadmissible symmetry-reduced bound", depth, h)
			}
			m := r.Intn(18)
			next := state.mul(cubieMoves[m])
			nh := db.bound(next.twist(), next.flip(), next.slice())
			if h > nh+1 || nh > h+1 {
				t.Fatal("inconsistent symmetry-reduced bound", h, nh)
			}
			state = next
		}
	}
	// A descending witness for uniformly sampled coordinates checks the table
	// all the way to the subgroup, including representatives with stabilizers.
	for trial := 0; trial < 1000; trial++ {
		state := uniformCubie(r)
		co, eo, sl := state.twist(), state.flip(), state.slice()
		for h := db.bound(co, eo, sl); h > 0; h-- {
			found := false
			for m := 0; m < 18; m++ {
				a, b, c := int(tables.Twist[co*18+m]), int(tables.Flip[eo*18+m]), int(tables.Slice[sl*18+m])
				if db.bound(a, b, c) == h-1 {
					co, eo, sl = a, b, c
					found = true
					break
				}
			}
			if !found {
				t.Fatal("pruning distance has no descending witness")
			}
		}
		if co != 0 || eo != 0 || sl != 0 {
			t.Fatal("zero distance is not subgroup membership")
		}
	}
	var hist [16]int
	for _, d := range db.distance {
		hist[d&15]++
		hist[d>>4]++
	}
	if hist[0] != 1 || hist[15] != 0 {
		t.Fatal("incomplete phase-one table", hist)
	}
	t.Logf("phase-one distance histogram: %v", hist)
}
