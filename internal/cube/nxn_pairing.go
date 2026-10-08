package cube

import (
	"encoding/binary"
	"fmt"
	"math/bits"
)

type wingAction struct {
	moves []Move
	full  [24]uint8
	outer [24]uint8
	want  [3]uint64
	pairs [12][2]uint8
}

func nxnOuterMoves(moves []Move) []Move {
	var outer []Move
	for _, m := range moves {
		if m.Layer == 0 {
			outer = append(outer, m)
		}
	}
	return outer
}

func nxnWingAction(n int, o *reductionOrbit, moves []Move) wingAction {
	p, q := nxnPermutation(n, moves), nxnPermutation(n, nxnOuterMoves(moves))
	return nxnWingActionFromPerm(o, moves, p, q)
}

func nxnWingActionFromPerm(o *reductionOrbit, moves []Move, p, q Permutation) wingAction {
	a := wingAction{moves: moves}
	var local [6 * 7 * 7]uint8
	for i, pos := range o.positions {
		local[pos] = uint8(i)
	}
	for i, pos := range o.positions {
		a.full[i], a.outer[i] = local[p[pos]], local[q[pos]]
	}
	a.prepare(o.mate)
	return a
}

func (a *wingAction) prepare(mate [24]uint8) {
	var inverseOuter, inverseFull, want [24]uint8
	for i := range a.full {
		inverseOuter[a.outer[i]], inverseFull[a.full[i]] = uint8(i), uint8(i)
	}
	pair := 0
	for i, dst := range a.full {
		want[i] = inverseOuter[dst]
		other := inverseFull[mate[dst]]
		if i < int(other) {
			a.pairs[pair] = [2]uint8{uint8(i), other}
			pair++
		}
	}
	a.want = nxnPackWings(want)
}

func nxnPackWings(p [24]uint8) [3]uint64 {
	return [3]uint64{binary.LittleEndian.Uint64(p[:8]), binary.LittleEndian.Uint64(p[8:16]), binary.LittleEndian.Uint64(p[16:])}
}

func (a *wingAction) fixedScore(p [3]uint64) int {
	matched := 0
	for i, target := range a.want {
		x := p[i] ^ target
		// Exact zero-byte count, without the cross-byte borrow of x-0x0101...
		zeros := ^(((x & 0x7f7f7f7f7f7f7f7f) + 0x7f7f7f7f7f7f7f7f) | x | 0x7f7f7f7f7f7f7f7f)
		matched += bits.OnesCount64(zeros)
	}
	return matched
}

func (a *wingAction) pairedScore(p, mates [24]uint8) int {
	matched := 0
	for _, pair := range a.pairs {
		if p[pair[0]] == mates[pair[1]] {
			matched += 2
		}
	}
	return matched
}

func nxnPreservesCenterColors(n int, p Permutation) bool {
	for src, dst := range p {
		face, row, col := indexToCoord(src, n)
		if !isBoundaryCell(face, row, col, n) && src/(n*n) != dst/(n*n) {
			return false
		}
	}
	return true
}

func nxnPairingActions(n int, o *reductionOrbit) []wingAction {
	var actions []wingAction
	seen := map[[48]uint8]bool{}
	add := func(moves []Move) {
		p := nxnPermutation(n, moves)
		if !nxnPreservesCenterColors(n, p) {
			return
		}
		a := nxnWingActionFromPerm(o, moves, p, nxnPermutation(n, nxnOuterMoves(moves)))
		var key [48]uint8
		copy(key[:24], a.full[:])
		copy(key[24:], a.outer[:])
		if !seen[key] {
			seen[key] = true
			actions = append(actions, a)
		}
	}
	// Slice, extract a completed pair to an outer layer, restore the slice.
	// The full sticker permutation verifies center preservation for every
	// orientation and layer; no hand-entered wing coordinates are assumed.
	for _, axis := range []Face{Right, Up, Front} {
		for _, layer := range []int{o.layer, n - 1 - o.layer} {
			for _, slice := range faceMoves(axis) {
				slice.Layer = layer
				for face := Front; face <= Down; face++ {
					if nxnAxis(face) == nxnAxis(axis) {
						continue
					}
					for _, a := range []Move{{Face: face, Clockwise: true}, {Face: face}} {
						for b := Front; b <= Down; b++ {
							if nxnAxis(b) != nxnAxis(axis) {
								continue
							}
							for _, turn := range faceMoves(b) {
								moves := append([]Move{slice, a, turn}, nxnInverse([]Move{a, slice})...)
								add(moves)
							}
							// Slice-flip-slice handles the last two edges.
							for f := Front; f <= Down; f++ {
								if nxnAxis(f) == nxnAxis(axis) || nxnAxis(f) == nxnAxis(face) {
									continue
								}
								for _, flip := range []Move{{Face: f, Clockwise: true}, {Face: f}} {
									moves := []Move{slice, a, {Face: b, Clockwise: true}}
									moves = append(moves, nxnInverse([]Move{a})...)
									moves = append(moves, flip)
									moves = append(moves, nxnInverse([]Move{a, flip})...)
									moves = append(moves, a)
									moves = append(moves, nxnInverse([]Move{slice})...)
									add(moves)
								}
							}
						}
					}
				}
			}
		}
	}
	return actions
}

func nxnAxis(face Face) int {
	switch face {
	case Right, Left:
		return 0
	case Up, Down:
		return 1
	default:
		return 2
	}
}

func nxnWingAfter(p [24]uint8, a wingAction) (next [24]uint8) {
	for i, id := range p {
		next[a.full[i]] = a.outer[id]
	}
	return next
}

func nxnWingMates(n int, o *reductionOrbit) (mate [24]uint8) {
	home := NewCube(n)
	for i, pos := range o.positions {
		for j, other := range o.positions {
			if nxnColor(home, pos) == nxnColor(home, o.partners[j]) && nxnColor(home, o.partners[i]) == nxnColor(home, other) {
				mate[i] = uint8(j)
			}
		}
	}
	return mate
}

func nxnSampleReduced(c *Cube, o *reductionOrbit) *Cube {
	r := nxnReducedSeed(c)
	if c.Size%2 == 0 {
		for i, pos := range o.positions {
			for _, index := range []int{pos, o.partners[i]} {
				f, row, col := indexToCoord(index, c.Size)
				compress := func(v int) int {
					if v == 0 {
						return 0
					}
					if v == c.Size-1 {
						return 2
					}
					return 1
				}
				r.Faces[f][compress(row)][compress(col)] = nxnColor(c, index)
			}
		}
	}
	return r
}

func nxnOLLParity(layer int) []Move {
	// Fifteen turns exchange the two wings of one edge and restore centers.
	// Opposite inner slices select the same wing orbit on every dimension.
	moves, _ := ParseMoves("2R2 B2 U2 2L U2 2R' U2 2R U2 F2 2R F2 2L' B2 2R2")
	for i := range moves {
		if moves[i].Layer != 0 {
			moves[i].Layer = layer
		}
	}
	return moves
}

func nxnPLLParity(n int) []Move {
	// Use the entire half-width block on larger even cubes. A single inner
	// slice with a two-layer Uw turn would disturb oblique center orbits.
	if n == 4 {
		moves, _ := ParseMoves("2R2 U2 2R2 Uw2 2R2 Uw2")
		return moves
	}
	r := []Move{{Face: Right, Wide: true, WideDepth: n / 2, Double: true}, {Face: Right, Double: true}}
	u := Move{Face: Up, Wide: true, WideDepth: n / 2, Double: true}
	moves := append(append([]Move{}, r...), Move{Face: Up, Double: true})
	moves = append(moves, r...)
	moves = append(moves, u)
	moves = append(moves, r...)
	return append(moves, u)
}

func nxnPairWings(c *Cube, t *reductionTables) ([]Move, *Cube, error) {
	var moves []Move
	var reduced *Cube
	for orbit, o := range t.wings {
		natural := c.Size%2 == 0 && orbit == 0
		if reduced == nil {
			reduced = nxnReducedSeed(c)
		}
		p, err := nxnWingPermutation(c, reduced, o)
		if err != nil {
			return nil, nil, err
		}
		if !natural && permutationParity(p) != 0 {
			part := nxnOLLParity(o.layer)
			if !nxnPreservesCenterColors(c.Size, nxnPermutation(c.Size, part)) {
				return nil, nil, fmt.Errorf("OLL parity disturbed centers")
			}
			if err := nxnApplyMoves(c, part); err != nil {
				return nil, nil, err
			}
			moves = append(moves, part...)
			reduced = nxnSampleReduced(c, t.wings[0])
			p, err = nxnWingPermutation(c, reduced, o)
			if err != nil {
				return nil, nil, err
			}
		}
		var state [24]uint8
		for i, id := range p {
			state[i] = uint8(id)
		}
		mate := nxnWingMates(c.Size, o)
		score := func(p [24]uint8) int {
			matched := 0
			for i, id := range p {
				if natural {
					if p[mate[i]] == mate[id] {
						matched++
					}
				} else if int(id) == i {
					matched++
				}
			}
			return matched
		}
		o.actionsOnce.Do(func() { o.actions = nxnPairingActions(c.Size, o) })
		actions := o.actions
		var outer []wingAction
		for _, m := range coordinateMoves {
			outer = append(outer, nxnWingAction(c.Size, o, []Move{m}))
		}
		for score(state) < 24 {
			before, bestGain, bestCost := score(state), 0, 1
			var best []Move
			var bestState [24]uint8
			consider := func(pre [24]uint8, setup []Move) {
				packed := nxnPackWings(pre)
				var mates [24]uint8
				if natural {
					for i, id := range pre {
						mates[i] = mate[id]
					}
				}
				for i := range actions {
					a := &actions[i]
					matched := 0
					if natural {
						matched = a.pairedScore(pre, mates)
					} else {
						matched = a.fixedScore(packed)
					}
					gain, cost := matched-before, len(a.moves)+len(setup)
					if gain > 0 && (best == nil || gain*bestCost > bestGain*cost) {
						bestGain, bestCost, bestState = gain, cost, nxnWingAfter(pre, *a)
						best = append(append([]Move{}, setup...), a.moves...)
					}
				}
			}
			consider(state, nil)
			for i, a := range outer {
				pre := nxnWingAfter(state, a)
				consider(pre, a.moves)
				for j, b := range outer {
					if i/3 == j/3 {
						continue
					}
					consider(nxnWingAfter(pre, b), []Move{a.moves[0], b.moves[0]})
				}
			}
			// A pure cycle is a cheap exact finish for unusual last-edge cases.
			for a := 0; a < 24; a++ {
				for b := a + 1; b < 24; b++ {
					for d := a + 1; d < 24; d++ {
						if d == b {
							continue
						}
						next := state
						next[b], next[d], next[a] = state[a], state[b], state[d]
						gain := score(next) - before
						if gain <= 0 {
							continue
						}
						cost := o.cycleCost(t, a, b, d)
						if best == nil || gain*bestCost > bestGain*cost {
							best, bestGain, bestCost, bestState = o.cycleMoves(t, a, b, d), gain, cost, next
						}
					}
				}
			}
			if best == nil {
				return nil, nil, fmt.Errorf("wing pairing stalled in layer %d at %d/24", o.layer+1, before)
			}
			if err := nxnApplyMoves(c, best); err != nil {
				return nil, nil, err
			}
			moves = append(moves, best...)
			state = bestState
		}
		reduced = nxnSampleReduced(c, t.wings[0])
		if natural {
			cubies := readCubie(reduced)
			flips := uint8(0)
			for _, flip := range cubies.eo {
				flips ^= flip
			}
			if flips != 0 {
				part := nxnOLLParity(o.layer)
				if err := nxnApplyMoves(c, part); err != nil {
					return nil, nil, err
				}
				moves = append(moves, part...)
				reduced = nxnSampleReduced(c, o)
			}
			cubies = readCubie(reduced)
			cp, ep := make([]int, 8), make([]int, 12)
			for i, id := range cubies.cp {
				cp[i] = int(id)
			}
			for i, id := range cubies.ep {
				ep[i] = int(id)
			}
			if permutationParity(cp) != permutationParity(ep) {
				part := nxnPLLParity(c.Size)
				if !nxnPreservesCenterColors(c.Size, nxnPermutation(c.Size, part)) {
					return nil, nil, fmt.Errorf("PLL parity disturbed centers")
				}
				if err := nxnApplyMoves(c, part); err != nil {
					return nil, nil, err
				}
				moves = append(moves, part...)
				reduced = nxnSampleReduced(c, o)
			}
		}
		if err := Validate3x3(reduced); err != nil {
			return nil, nil, fmt.Errorf("paired 3x3: %w", err)
		}
	}
	return moves, reduced, nil
}
