package cube

import (
	"fmt"
	"sort"
	"time"
)

// The eight rotations preserving U/D, together with their left/right mirrors,
// form the 16-element subgroup used to quotient phase-one pair coordinates.
// Reflection reverses corner winding; edge flip is unchanged in our facelet
// convention. Conjugating a flip also needs the slice membership, so retain
// its affine (permutation XOR delta) form instead of a million-entry map.
func mirrorCubie(s cubie) cubie {
	corners := [8]uint8{1, 0, 3, 2, 5, 4, 7, 6}
	edges := [12]uint8{2, 1, 0, 3, 6, 5, 4, 7, 9, 8, 11, 10}
	var n cubie
	for i, p := range corners {
		n.cp[i], n.co[i] = corners[s.cp[p]], (3-s.co[p])%3
	}
	for i, p := range edges {
		n.ep[i], n.eo[i] = edges[s.ep[p]], s.eo[p]
	}
	return n
}

var compactRotations = phase1Symmetries()

func compactConjugate(s cubie, sym int) cubie {
	if sym >= 8 {
		s = mirrorCubie(s)
	}
	r := compactRotations[sym&7]
	return r.inverse().mul(s).mul(r)
}

func symmetryClasses(conj []uint16, size int) ([]uint16, []int) {
	return symmetryClassesStep(conj, size, 1)
}

func symmetryClassesStep(conj []uint16, size, step int) ([]uint16, []int) {
	codes, classes := make([]uint16, size), make([]uint16, size)
	var reps []int
	for x := 0; x < size; x++ {
		rep, selected := x, 0
		for sym := 0; sym < 16; sym += step {
			if y := int(conj[x*16+sym]); y < rep {
				rep, selected = y, sym
			}
		}
		if rep == x {
			classes[x] = uint16(len(reps))
			reps = append(reps, x)
		}
		codes[x] = classes[rep]*16 + uint16(selected)
	}
	return codes, reps
}

func (t *coordinateTables) compactPhase1() {
	t.SymTwist = make([]uint16, 2187*16)
	t.SymFlip = make([]uint16, 2048*16)
	t.SymFlipDelta, t.SymSlice = make([]uint16, 495*16), make([]uint16, 495*16)
	rotations := phase1Symmetries()
	for sym := 0; sym < 16; sym++ {
		r, inv := rotations[sym&7], rotations[sym&7].inverse()
		conjugate := func(s cubie) cubie {
			if sym >= 8 {
				s = mirrorCubie(s)
			}
			return inv.mul(s).mul(r)
		}
		for x := 0; x < 2187; x++ {
			t.SymTwist[x*16+sym] = uint16(conjugate(twistCubie(x)).twist())
		}
		for x := 0; x < 2048; x++ {
			// Subtract the conjugated identity to isolate the linear map.
			t.SymFlip[x*16+sym] = uint16(conjugate(flipCubie(x)).flip())
		}
		for x := 0; x < 495; x++ {
			n := conjugate(sliceCubie(x))
			t.SymFlipDelta[x*16+sym], t.SymSlice[x*16+sym] = uint16(n.flip()), uint16(n.slice())
		}
	}
	var twists, slices []int
	t.TwistSym, twists = symmetryClasses(t.SymTwist, 2187)
	t.TwistRepresentatives = make([]uint16, len(twists))
	for i, rep := range twists {
		t.TwistRepresentatives[i] = uint16(rep)
	}
	t.SliceSym, slices = symmetryClasses(t.SymSlice, 495)
	ts, fs := make([]uint8, len(twists)*495), make([]uint8, len(slices)*2048)
	for c, rep := range twists {
		copy(ts[c*495:], t.TwistSlice[rep*495:(rep+1)*495])
	}
	for c, rep := range slices {
		for eo := 0; eo < 2048; eo++ {
			fs[c*2048+eo] = t.FlipSlice[eo*495+rep]
		}
	}
	t.TwistSlice, t.FlipSlice = packNibbles(ts), packNibbles(fs)
	// The even subgroup preserves edge-flip convention without knowing the
	// omitted slice membership. Its eight symmetries yield 324 twist classes.
	var tfReps []int
	t.TwistFlipSym, tfReps = symmetryClassesStep(t.SymTwist, 2187, 2)
	tf := make([]uint8, len(tfReps)*2048)
	for c, rep := range tfReps {
		copy(tf[c*2048:], t.TwistFlip[rep*2048:(rep+1)*2048])
	}
	t.TwistFlip = packNibbles(tf)
}

func packNibbles(d []uint8) []uint8 {
	p := make([]uint8, (len(d)+1)/2)
	for i, v := range d {
		if v > 15 {
			panic(fmt.Sprintf("compact distance %d at %d", v, i))
		}
		p[i/2] |= v << (4 * (i & 1))
	}
	return p
}

func (t *coordinateTables) twistSliceBound(co, sl int) int {
	code := int(t.TwistSym[co])
	return nibbleDistance(t.TwistSlice, code/16*495+int(t.SymSlice[sl*16+(code&15)]))
}

func (t *coordinateTables) flipSliceBound(eo, sl int) int {
	code := int(t.SliceSym[sl])
	sym := code & 15
	flip := t.SymFlip[eo*16+sym] ^ t.SymFlipDelta[sl*16+sym]
	return nibbleDistance(t.FlipSlice, code/16*2048+int(flip))
}

func maxSliceClass(codes []uint16) uint16 {
	var n uint16
	for _, code := range codes {
		n = max(n, code/16)
	}
	return n
}

// Phase two quotients each eight-piece permutation by the same 16 symmetries.
// Pairing representatives with slice permutation and the other piece group's
// four-piece combination keeps both correlation bounds in a few hundred KB.
func (t *coordinateTables) compactPhase2() {
	cpConj, epConj := make([]uint16, 40320*16), make([]uint16, 40320*16)
	t.SymSlice2 = make([]uint16, 24*16)
	t.SymCornerComb, t.SymEdgeComb = make([]uint16, 140*16), make([]uint16, 140*16)
	for x := 0; x < 40320; x++ {
		s := identityCubie()
		setPermutation(s.cp[:], x)
		e := identityCubie()
		setPermutation(e.ep[:8], x)
		for sym := 0; sym < 16; sym++ {
			nc, ne := compactConjugate(s, sym), compactConjugate(e, sym)
			cpConj[x*16+sym] = uint16(permutationRank(nc.cp[:]))
			epConj[x*16+sym] = uint16(permutationRank(ne.ep[:8]))
			t.SymCornerComb[cornerCombination(s)*16+sym] = uint16(cornerCombination(nc))
			t.SymEdgeComb[cornerCombination(s)*16+sym] = uint16(edgeCombination(ne))
		}
	}
	for x := 0; x < 24; x++ {
		s := identityCubie()
		setPermutation(s.ep[8:], x)
		for i := 8; i < 12; i++ {
			s.ep[i] += 8
		}
		for sym := 0; sym < 16; sym++ {
			n := compactConjugate(s, sym)
			t.SymSlice2[x*16+sym] = uint16(permutationRank(n.ep[8:]))
		}
	}
	var cr, er []int
	t.PermSymCorner, cr = symmetryClasses(cpConj, 40320)
	t.PermSymEdge, er = symmetryClasses(epConj, 40320)
	reduce := func(raw []uint8, reps []int, size int) []uint8 {
		d := make([]uint8, len(reps)*size)
		for c, rep := range reps {
			copy(d[c*size:], raw[rep*size:(rep+1)*size])
		}
		return packNibbles(d)
	}
	t.CornerSlice, t.EdgeSlice = reduce(t.CornerSlice, cr, 24), reduce(t.EdgeSlice, er, 24)
	t.EdgeComb, t.CornerEdgeComb = reduce(t.EdgeComb, er, 140), reduce(t.CornerEdgeComb, cr, 140)
}

func (t *coordinateTables) cornerSliceBound(cp, sp int) int {
	code := int(t.PermSymCorner[cp])
	return nibbleDistance(t.CornerSlice, code/16*24+int(t.SymSlice2[sp*16+(code&15)]))
}
func (t *coordinateTables) edgeSliceBound(ep, sp int) int {
	code := int(t.PermSymEdge[ep])
	return nibbleDistance(t.EdgeSlice, code/16*24+int(t.SymSlice2[sp*16+(code&15)]))
}
func (t *coordinateTables) edgeCornerBound(ep, comb int) int {
	code := int(t.PermSymEdge[ep])
	return nibbleDistance(t.EdgeComb, code/16*140+int(t.SymCornerComb[comb*16+(code&15)]))
}
func (t *coordinateTables) cornerEdgeBound(cp, comb int) int {
	code := int(t.PermSymCorner[cp])
	return nibbleDistance(t.CornerEdgeComb, code/16*140+int(t.SymEdgeComb[comb*16+(code&15)]))
}
func edgeCombination(s cubie) int { copy(s.cp[:], s.ep[:8]); return cornerCombination(s) }
func edgeCombinationMoves() []uint16 {
	return makeMoveTable(140, func(x int) cubie {
		// Find one permutation for this combination and parity; this small setup
		// scan is independent of the transition action being generated.
		for p := 0; p < 40320; p++ {
			s := identityCubie()
			setPermutation(s.ep[:8], p)
			if edgeCombination(s) == x {
				return s
			}
		}
		panic("missing edge combination")
	}, edgeCombination, true, time.Time{})
}

func (t *coordinateTables) twistFlipBound(co, eo int) int {
	code := int(t.TwistFlipSym[co])
	return nibbleDistance(t.TwistFlip, code/16*2048+int(t.SymFlip[eo*16+(code&15)]))
}

const nearPhase1Depth = 5

func (t *coordinateTables) generateNearPhase1() {
	t.nearPhase1 = map[uint32]uint8{0: 0}
	queue := []uint32{0}
	for head := 0; head < len(queue); head++ {
		key := queue[head]
		d := t.nearPhase1[key]
		if d == nearPhase1Depth {
			continue
		}
		co, fs := int(key)/flipSliceStates, int(key)%flipSliceStates
		eo, sl := fs/495, fs%495
		for m := 0; m < 18; m++ {
			next := uint32(int(t.Twist[co*18+m])*flipSliceStates + int(t.Flip[eo*18+m])*495 + int(t.Slice[sl*18+m]))
			if _, found := t.nearPhase1[next]; !found {
				t.nearPhase1[next] = d + 1
				queue = append(queue, next)
			}
		}
	}
	compact := make(map[uint32]uint8)
	for key, d := range t.nearPhase1 {
		co, fs := int(key)/flipSliceStates, int(key)%flipSliceStates
		canonical := t.nearPhase1Key(co, fs/495, fs%495)
		compact[canonical] = d
	}
	t.nearPhase1 = compact
	keys := make([]uint32, 0, len(compact))
	for key := range compact {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	t.NearPhase1Keys = keys
	t.NearPhase1Distances = make([]uint8, len(keys))
	for i, key := range keys {
		t.NearPhase1Distances[i] = compact[key]
	}

}

func (t *coordinateTables) nearPhase1Bound(co, eo, sl int) int {
	if d, ok := t.nearPhase1[t.nearPhase1Key(co, eo, sl)]; ok {
		return int(d)
	}
	return nearPhase1Depth + 1
}

func (t *coordinateTables) nearPhase1Key(co, eo, sl int) uint32 {
	code := int(t.TwistSym[co])
	sym := code & 15
	flip := int(t.SymFlip[eo*16+sym] ^ t.SymFlipDelta[sl*16+sym])
	slice := int(t.SymSlice[sl*16+sym])
	return uint32(code/16*flipSliceStates + flip*495 + slice)
}
