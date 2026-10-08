package cube

import (
	"math/rand"
	"os"
	"testing"
	"time"
)

func TestOptimalSortedCoordinates(t *testing.T) {
	tables := solverTables()
	db := optimalPhase1Coordinates(tables, time.Time{})
	r := rand.New(rand.NewSource(2026100722))
	for x := 0; x < sortedSliceStates; x++ {
		if sliceSorted(sortedSliceCubie(x)) != x {
			t.Fatal("sorted slice round trip", x)
		}
	}
	for trial := 0; trial < 1000; trial++ {
		state := uniformCubie(r)
		co, eo, sorted := state.twist(), state.flip(), sliceSorted(state)
		for m, move := range cubieMoves {
			if sliceSorted(state.mul(move)) != int(db.sortedMove[sorted*18+m]) {
				t.Fatal("sorted slice transition")
			}
		}
		for sym := 0; sym < 16; sym++ {
			next := compactConjugate(state, sym)
			f, s := db.conjugate(eo, sorted, sym)
			if next.flip() != f || sliceSorted(next) != s {
				t.Fatal("sorted flip conjugation")
			}
			code := int(tables.TwistSym[co])
			if sym == code&15 && next.twist() != int(tables.TwistRepresentatives[code/16]) {
				t.Fatal("twist quotient")
			}
		}
	}
	if optimalPhase1Coordinates(tables, time.Now().Add(-time.Second)) != nil {
		t.Fatal("expired coordinate generation")
	}
	t.Logf("%d classes, %d states, %d packed bytes", len(db.stabilizers), len(db.stabilizers)*optimalFlipSliceStates, len(db.stabilizers)*optimalFlipSliceStates/4)
}

func TestOptimalSortedTableOracle(t *testing.T) {
	if os.Getenv("CUBE_OPTIMAL_TABLES") != "1" {
		t.Skip("make test-optimal")
	}
	start := time.Now()
	tables := solverTables()
	db := optimalPhase1Tables(tables, time.Time{})
	if db == nil {
		t.Fatal("sorted phase-one database failed")
	}
	t.Logf("sorted phase-one generation/load %v; %d pruning bytes", time.Since(start), len(db.distance))
	r := rand.New(rand.NewSource(2026100723))
	for trial := 0; trial < 1000; trial++ {
		state := identityCubie()
		for depth := 0; depth <= 14; depth++ {
			co, eo, sorted := state.twist(), state.flip(), sliceSorted(state)
			h := db.rootBound(co, eo, sorted)
			if trial < 100 {
				s := largeOptimalSearch{t: tables, huge: db}
				s.initializeAxes(state)
				if s.timedOut || threeAxisBound(s.hugeBound[0]) > depth {
					t.Fatal("inadmissible three-axis bound", depth, s.hugeBound[0])
				}
			}
			if h < 0 || h > depth {
				t.Fatal("inadmissible sorted bound", depth, h)
			}
			for m, move := range cubieMoves {
				next := state.mul(move)
				a, b, c := next.twist(), next.flip(), sliceSorted(next)
				nh := db.childBound(a, b, c, h)
				if nh != db.rootBound(a, b, c) || h > nh+1 || nh > h+1 {
					t.Fatal("inconsistent modulo-three bound", depth, h, nh)
				}
				if a != int(tables.Twist[co*18+m]) || b != int(tables.Flip[eo*18+m]) || c != int(db.sortedMove[sorted*18+m]) {
					t.Fatal("sorted coordinate move")
				}
			}
			state = state.mul(cubieMoves[r.Intn(18)])
		}
	}
	var hist [4]uint64
	for _, v := range db.distance {
		for shift := 0; shift < 8; shift += 2 {
			hist[(v>>shift)&3]++
		}
	}
	t.Logf("modulo-three histogram, unknown means >=11: %v", hist)
}

func TestOptimalSortedGenerationDeadline(t *testing.T) {
	db := &optimalPhase1{}
	if db.pruning(time.Now().Add(-time.Second)) != nil {
		t.Fatal("expired sorted table generation")
	}
}
