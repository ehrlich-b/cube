package cube

import (
	"bytes"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestCompactSymmetryConjugation(t *testing.T) {
	tables := solverTables()
	r := rand.New(rand.NewSource(2026100720))
	for trial := 0; trial < 100; trial++ {
		s := uniformCubie(r)
		for sym := 0; sym < 16; sym++ {
			n := compactConjugate(s, sym)
			co, eo, sl := s.twist(), s.flip(), s.slice()
			if int(tables.SymTwist[co*16+sym]) != n.twist() || int(tables.SymSlice[sl*16+sym]) != n.slice() || int(tables.SymFlip[eo*16+sym]^tables.SymFlipDelta[sl*16+sym]) != n.flip() {
				t.Fatal("compact coordinates disagree with full conjugation", sym)
			}
			for m, move := range cubieMoves {
				mapped := compactConjugate(move, sym)
				found := false
				for _, candidate := range cubieMoves {
					found = found || mapped == candidate
				}
				if !found || compactConjugate(s.mul(move), sym) != n.mul(mapped) {
					t.Fatal("symmetry does not preserve move action", sym, m)
				}
			}
		}
	}
}

// Rebuild from this engine's move action, and check every quotient entry against
// independent unquotiented BFS distances. The embedded asset is reproducible;
// generation is explicit and never part of a normal first solve.
func TestCompactTableGeneration(t *testing.T) {
	if os.Getenv("CUBE_GENERATE_TABLES") != "1" {
		t.Skip("make generate-tables")
	}
	tables := generateCoordinateTables(time.Time{})
	t.Logf("twist classes %d; slice classes %d", (len(tables.TwistSlice)*2)/495, len(tables.FlipSlice)*2/2048)
	ts := pairPruning(tables.Twist, tables.Slice, 495, false, time.Time{})
	fs := pairPruning(tables.Flip, tables.Slice, 495, false, time.Time{})
	for co := 0; co < 2187; co++ {
		for sl := 0; sl < 495; sl++ {
			if tables.twistSliceBound(co, sl) != int(ts[co*495+sl]) {
				t.Fatal("twist quotient changes distance", co, sl)
			}
		}
	}
	for eo := 0; eo < 2048; eo++ {
		for sl := 0; sl < 495; sl++ {
			if tables.flipSliceBound(eo, sl) != int(fs[eo*495+sl]) {
				t.Fatal("flip quotient changes distance", eo, sl)
			}
		}
	}
	tf := pairPruning(tables.Twist, tables.Flip, 2048, false, time.Time{})
	for co := 0; co < 2187; co++ {
		for eo := 0; eo < 2048; eo++ {
			if tables.twistFlipBound(co, eo) != int(tf[co*2048+eo]) {
				t.Fatal("twist/flip quotient changes distance", co, eo)
			}
		}
	}
	for sl := 0; sl < 495; sl++ {
		for sym := 0; sym < 16; sym += 2 {
			if tables.SymFlipDelta[sl*16+sym] != 0 {
				t.Fatal("twist/flip subgroup requires omitted slice", sl, sym)
			}
		}
	}
	cs := pairPruning(tables.Corner, tables.Slice2, 24, true, time.Time{})
	es := pairPruning(tables.Edge2, tables.Slice2, 24, true, time.Time{})
	ec := pairPruning(tables.Edge2, cornerCombinationMoves(), 140, true, time.Time{})
	ce := pairPruning(tables.Corner, edgeCombinationMoves(), 140, true, time.Time{})
	for p := 0; p < 40320; p++ {
		for sp := 0; sp < 24; sp++ {
			if tables.cornerSliceBound(p, sp) != int(cs[p*24+sp]) || tables.edgeSliceBound(p, sp) != int(es[p*24+sp]) {
				t.Fatal("permutation/slice quotient changes distance", p, sp)
			}
		}
		for comb := 0; comb < 140; comb++ {
			if tables.edgeCornerBound(p, comb) != int(ec[p*140+comb]) || tables.cornerEdgeBound(p, comb) != int(ce[p*140+comb]) {
				t.Fatal("correlation quotient changes distance", p, comb)
			}
		}
	}
	t.Logf("exact phase-one radius-five frontier: %d entries", len(tables.NearPhase1Keys))
	t.Logf("phase-two corner/edge classes: %d/%d", maxSliceClass(tables.PermSymCorner)+1, maxSliceClass(tables.PermSymEdge)+1)
	data := encodeCoordinateTables(tables)
	if len(data) > 5_000_000 {
		t.Fatal("generated table exceeds repository size limit", len(data))
	}
	if got := decodeCoordinateTables(data, time.Time{}); !reflect.DeepEqual(got, tables) {
		t.Fatal("compact table roundtrip")
	}
	path := filepath.Join("tables", "coordinates-v5.bin.gz")
	if !bytes.Equal(data, embeddedCoordinates) {
		if err := os.WriteFile(path, data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("verified and generated %s: %d bytes", path, len(data))
}

func TestEmbeddedCoordinateTables(t *testing.T) {
	if decodeCoordinateTables(embeddedCoordinates, time.Time{}) == nil {
		t.Fatal("invalid embedded compact tables; run make generate-tables")
	}
	if len(embeddedCoordinates) > 5_000_000 {
		t.Fatal("embedded tables exceed size limit")
	}
}

// Descending witnesses prove each stored distance is attainable. Closure of
// every layer below the frontier proves no omitted state is within five turns.
func TestNearPhaseOneFrontier(t *testing.T) {
	tables := solverTables()
	for i, key := range tables.NearPhase1Keys {
		if i > 0 && key <= tables.NearPhase1Keys[i-1] {
			t.Fatal("unsorted or duplicate frontier key")
		}
		co, fs := int(tables.TwistRepresentatives[int(key)/flipSliceStates]), int(key)%flipSliceStates
		eo, sl := fs/495, fs%495
		d := tables.nearPhase1Bound(co, eo, sl)
		if d < 0 || d > nearPhase1Depth || (d == 0) != (key == 0) {
			t.Fatal("invalid frontier distance", key, d)
		}
		witness := d == 0
		for m := 0; m < 18; m++ {
			a, b, c := int(tables.Twist[co*18+m]), int(tables.Flip[eo*18+m]), int(tables.Slice[sl*18+m])
			n := tables.nearPhase1Bound(a, b, c)
			if n > d+1 || d > n+1 {
				t.Fatal("frontier layer is not closed or consistent", key, d, n)
			}
			witness = witness || n == d-1
		}
		if !witness {
			t.Fatal("frontier distance has no descending witness", key, d)
		}
	}
}
