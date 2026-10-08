package cube

import (
	"fmt"
	"math/bits"
	"os"
	"runtime"
	"sync/atomic"
	"time"
)

const sortedSliceStates = 495 * 24
const optimalFlipSliceStates = 2048 * sortedSliceStates
const optimalPhase1Cap = 11

// The huge phase-one coordinate fixes the four slice edges in place, in
// addition to orienting all cubies. Corner twists are quotiented by the 16
// U/D-axis symmetries. Two bits store exact distances modulo three through
// depth ten; code three means an admissible lower bound of eleven. Saturation
// preserves the one-move Lipschitz bound, so child distances follow from the
// parent's bound without a second search. The three axis goals are subgroups.
type optimalPhase1 struct {
	t           *coordinateTables
	sortedMove  []uint16
	sortedConj  []uint16
	flipDelta   []uint16
	twistMove   []uint16
	stabilizers []uint16
	distance    []uint8
}

var optimalPhase1Lock = make(chan struct{}, 1)
var optimalPhase1DB atomic.Pointer[optimalPhase1]

func sliceSorted(s cubie) int {
	var permutation [4]uint8
	n := 0
	for _, id := range s.ep {
		if id >= 8 {
			permutation[n] = id - 8
			n++
		}
	}
	return s.slice()*24 + permutationRank(permutation[:])
}

func sortedSliceCubie(x int) cubie {
	s := sliceCubie(x / 24)
	var permutation [4]uint8
	setPermutation(permutation[:], x%24)
	for p, id := range s.ep {
		if id >= 8 {
			s.ep[p] = permutation[id-8] + 8
		}
	}
	return s
}

func optimalPhase1Coordinates(t *coordinateTables, deadline time.Time) *optimalPhase1 {
	if tableDeadlineExceeded(deadline) {
		return nil
	}
	db := &optimalPhase1{t: t, sortedMove: make([]uint16, sortedSliceStates*18), sortedConj: make([]uint16, sortedSliceStates*16), flipDelta: make([]uint16, sortedSliceStates*16)}
	for x := 0; x < sortedSliceStates; x++ {
		if x&255 == 0 && tableDeadlineExceeded(deadline) {
			return nil
		}
		s := sortedSliceCubie(x)
		for m, move := range cubieMoves {
			db.sortedMove[x*18+m] = uint16(sliceSorted(s.mul(move)))
		}
		for sym := 0; sym < 16; sym++ {
			view := compactConjugate(s, sym)
			db.sortedConj[x*16+sym], db.flipDelta[x*16+sym] = uint16(sliceSorted(view)), uint16(view.flip())
		}
	}
	db.twistMove = make([]uint16, len(t.TwistRepresentatives)*18)
	db.stabilizers = make([]uint16, len(t.TwistRepresentatives))
	for c, rep := range t.TwistRepresentatives {
		for m := 0; m < 18; m++ {
			db.twistMove[c*18+m] = t.TwistSym[t.Twist[int(rep)*18+m]]
		}
		for sym := 0; sym < 16; sym++ {
			if t.SymTwist[int(rep)*16+sym] == rep {
				db.stabilizers[c] |= 1 << sym
			}
		}
	}
	return db
}

func (db *optimalPhase1) conjugate(flip, sorted, sym int) (int, int) {
	return int(db.t.SymFlip[flip*16+sym] ^ db.flipDelta[sorted*16+sym]), int(db.sortedConj[sorted*16+sym])
}

func (db *optimalPhase1) index(co, eo, sorted int) uint32 {
	code := int(db.t.TwistSym[co])
	flip, slice := db.conjugate(eo, sorted, code&15)
	return uint32(code/16*optimalFlipSliceStates + flip*sortedSliceStates + slice)
}

func (db *optimalPhase1) residue(co, eo, sorted int) int {
	x := db.index(co, eo, sorted)
	return int(db.distance[x>>2] >> ((x & 3) * 2) & 3)
}

func (db *optimalPhase1) rootBound(co, eo, sorted int) int {
	if db.residue(co, eo, sorted) == 3 {
		return optimalPhase1Cap
	}
	h := 0
	for co != 0 || eo != 0 || sorted != 0 {
		want := (db.residue(co, eo, sorted) + 2) % 3
		found := false
		for m := 0; m < 18; m++ {
			a, b, c := int(db.t.Twist[co*18+m]), int(db.t.Flip[eo*18+m]), int(db.sortedMove[sorted*18+m])
			if db.residue(a, b, c) == want {
				co, eo, sorted = a, b, c
				h++
				found = true
				break
			}
		}
		// A complete generated table always has a descending witness. Treat a
		// malformed cached table as unavailable rather than looping indefinitely.
		if !found || h > 10 {
			return -1
		}
	}
	return h
}

var optimalPhase1Children = func() [optimalPhase1Cap + 1][4]int8 {
	var children [optimalPhase1Cap + 1][4]int8
	for parent := range children {
		for residue := 0; residue < 3; residue++ {
			delta := (residue - parent%3 + 3) % 3
			if delta == 2 {
				delta = -1
			}
			children[parent][residue] = int8(parent + delta)
		}
		children[parent][3] = optimalPhase1Cap
	}
	return children
}()

func (db *optimalPhase1) childBound(co, eo, sorted, parent int) int {
	return int(optimalPhase1Children[parent][db.residue(co, eo, sorted)])
}

func (db *optimalPhase1) pruning(deadline time.Time) []uint8 {
	if tableDeadlineExceeded(deadline) {
		return nil
	}
	size := len(db.stabilizers) * optimalFlipSliceStates
	d := make([]uint8, size/4)
	for i := range d {
		d[i] = 255
	}
	d[0] &= 252
	frontier := []uint32{0}
	for depth := 1; depth < optimalPhase1Cap; depth++ {
		var next []uint32
		if depth < optimalPhase1Cap-1 {
			next = make([]uint32, 0, min(len(frontier)*4, 1<<26))
		}
		value := uint8(depth % 3)
		for head, x := range frontier {
			if head&4095 == 0 && tableDeadlineExceeded(deadline) {
				return nil
			}
			c, fs := int(x)/optimalFlipSliceStates, int(x)%optimalFlipSliceStates
			flip, sorted := fs/sortedSliceStates, fs%sortedSliceStates
			for m := 0; m < 18; m++ {
				code := int(db.twistMove[c*18+m])
				nf, ns := db.conjugate(int(db.t.Flip[flip*18+m]), int(db.sortedMove[sorted*18+m]), code&15)
				nc := code / 16
				y := uint32(nc*optimalFlipSliceStates + nf*sortedSliceStates + ns)
				shift := (y & 3) * 2
				if d[y>>2]>>shift&3 != 3 {
					continue
				}
				d[y>>2] = (d[y>>2] &^ (3 << shift)) | value<<shift
				if depth < optimalPhase1Cap-1 {
					next = append(next, y)
				}
				for stabilizers := db.stabilizers[nc] &^ 1; stabilizers != 0; stabilizers &= stabilizers - 1 {
					sym := bits.TrailingZeros16(stabilizers)
					f, s := db.conjugate(nf, ns, sym)
					z := uint32(nc*optimalFlipSliceStates + f*sortedSliceStates + s)
					shift := (z & 3) * 2
					if d[z>>2]>>shift&3 == 3 {
						d[z>>2] = (d[z>>2] &^ (3 << shift)) | value<<shift
						if depth < optimalPhase1Cap-1 {
							next = append(next, z)
						}
					}
				}
			}
		}
		frontier = next
	}
	if tableDeadlineExceeded(deadline) {
		return nil
	}
	return d
}

func optimalPhase1Tables(t *coordinateTables, deadline time.Time) *optimalPhase1 {
	if runtime.GOARCH == "wasm" || os.Getenv("CUBE_SMALL_TABLES") == "1" {
		return nil
	}
	if !lockSearchTables(optimalPhase1Lock, deadline) {
		return nil
	}
	defer func() { <-optimalPhase1Lock }()
	if db := optimalPhase1DB.Load(); db != nil {
		return db
	}
	db := optimalPhase1Coordinates(t, deadline)
	if db == nil {
		return nil
	}
	filename := "optimal-phase1-sorted-sym16-cap11-v1.bin"
	size := len(db.stabilizers) * optimalFlipSliceStates / 4
	db.distance = loadPatternBytes(filename, size, deadline)
	if db.distance != nil && db.distance[0]&3 != 0 {
		db.distance = nil
	}
	if db.distance == nil {
		if tableDeadlineExceeded(deadline) {
			return nil
		}
		fmt.Fprintf(os.Stderr, "Building the optimal sorted phase-one database once: %.2f MiB of cache; allow tens of minutes; initialization counts toward --time-limit.\n", float64(size+64)/(1<<20))
		db.distance = db.pruning(deadline)
		if db.distance == nil {
			return nil
		}
		if deadline.IsZero() || time.Until(deadline) > time.Second {
			savePackedPattern(filename, db.distance)
		}
	}
	if tableDeadlineExceeded(deadline) {
		return nil
	}
	optimalPhase1DB.Store(db)
	return db
}
