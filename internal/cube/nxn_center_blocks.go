package cube

import "math/bits"

type centerTransfer struct{ src, dst uint8 }
type centerBlockAction struct {
	moves []Move
	trans []centerTransfer
	masks []centerBlockMask
}

type centerBlockMask struct {
	color       Color
	added, lost [3]uint64
}

func (a *centerBlockAction) buildMasks(targets []Color) {
	var masks [6]centerBlockMask
	for color := range masks {
		masks[color].color = Color(color)
	}
	for _, step := range a.trans {
		m := &masks[targets[step.dst]]
		m.added[step.src/64] |= 1 << (step.src % 64)
		m.lost[step.dst/64] |= 1 << (step.dst % 64)
	}
	var compact [6]centerBlockMask
	count := 0
	for _, m := range masks {
		var nonzero uint64
		for word := range m.added {
			common := m.added[word] & m.lost[word]
			m.added[word] &^= common
			m.lost[word] &^= common
			nonzero |= m.added[word] | m.lost[word]
		}
		if nonzero != 0 {
			compact[count] = m
			count++
		}
	}
	a.masks = append([]centerBlockMask(nil), compact[:count]...)
}

func (a *centerBlockAction) gain(colors *[6][3]uint64) int {
	gain := 0
	for i := range a.masks {
		m := &a.masks[i]
		for word, present := range colors[m.color] {
			gain += bits.OnesCount64(present&m.added[word]) - bits.OnesCount64(present&m.lost[word])
		}
	}
	return gain
}

// Each gain is a linear sum of color-equality terms. Record the affected
// actions by position/color so a selected bar updates only changed terms.
func nxnCenterInfluences(actions []centerBlockAction) (out [144][6][]uint32) {
	for id := range actions {
		for _, m := range actions[id].masks {
			for word := range m.added {
				for sign, mask := range [2]uint64{m.added[word], m.lost[word]} {
					for mask != 0 {
						pos := word*64 + bits.TrailingZeros64(mask)
						out[pos][m.color] = append(out[pos][m.color], uint32(id*2+sign))
						mask &= mask - 1
					}
				}
			}
		}
	}
	return out
}

// Generalize a center commutator from one intersecting row/column to entire
// bars. Wings may move during this stage; the fixed-center frame is preserved.
// Full permutations include oblique orbits and central rows of odd cubes.
type centerRange struct {
	moves         []Move
	perm, inverse [6 * 7 * 7]uint16
	conjugates    [6]centerConjugate
}

type centerConjugate struct {
	perm, inverse [6 * 7 * 7]uint16
}

func nxnCenterRange(n int, face Face, lo, hi int, clockwise bool) *centerRange {
	r := &centerRange{}
	for layer := lo; layer <= hi; layer++ {
		r.moves = append(r.moves, Move{Face: face, Layer: layer, Clockwise: clockwise})
	}
	r.moves = nxnOptimizeMoves(r.moves, n)
	for src, dst := range nxnPermutation(n, r.moves) {
		r.perm[src], r.inverse[dst] = uint16(dst), uint16(src)
	}
	return r
}

func nxnCenterBlockActions(n int, t *reductionTables) []centerBlockAction {
	var positions []int
	var local [6 * 7 * 7]uint8
	var targets []Color
	home := NewCube(n)
	for _, o := range t.centers {
		for _, pos := range o.positions {
			local[pos] = uint8(len(positions))
			positions = append(positions, pos)
			targets = append(targets, nxnColor(home, pos))
		}
	}
	seen := map[string]bool{}
	var actions []centerBlockAction
	for axis := 0; axis < 3; axis++ {
		rFace := []Face{Right, Up, Front}[axis]
		uFace := []Face{Up, Front, Right}[axis]
		fFace := []Face{Front, Right, Up}[axis]
		var outer [6]Move
		copy(outer[:3], faceMoves(uFace))
		copy(outer[3:], faceMoves([]Face{Down, Back, Left}[axis]))
		var uPerms [6]Permutation
		for i, u := range outer {
			uPerms[i] = nxnPermutation(n, []Move{u})
		}
		var rs, fs [7][7][2]*centerRange
		for lo := 1; lo < n-1; lo++ {
			for hi := lo; hi < n-1; hi++ {
				for q := 0; q < 2; q++ {
					r := nxnCenterRange(n, rFace, lo, hi, q == 0)
					rs[lo][hi][q] = r
					fs[lo][hi][q] = nxnCenterRange(n, fFace, lo, hi, q == 0)
					for i, u := range uPerms {
						g := &r.conjugates[i]
						for src := range u {
							dst := r.inverse[u[r.perm[src]]]
							g.perm[src], g.inverse[dst] = dst, uint16(src)
						}
					}
				}
			}
		}
		// Preserve the original generator order, including all cost ties.
		for ra := 1; ra < n-1; ra++ {
			for rb := ra; rb < n-1; rb++ {
				for fa := 1; fa < n-1; fa++ {
					for fb := fa; fb < n-1; fb++ {
						for rq := 0; rq < 2; rq++ {
							r := rs[ra][rb][rq]
							for fq := 0; fq < 2; fq++ {
								f := fs[fa][fb][fq]
								for ui, u := range outer {
									g := &r.conjugates[ui]
									after := func(pos int) int {
										return int(f.inverse[g.inverse[f.perm[g.perm[pos]]]])
									}
									// Only fixed centers can leave the movable-center
									// union. Wings and corners are free at this stage.
									valid := true
									if n%2 == 1 {
										for face := 0; face < 6; face++ {
											pos := face*n*n + (n/2)*n + n/2
											if after(pos) != pos {
												valid = false
												break
											}
										}
									}
									if !valid {
										continue
									}
									var keyBuffer [144]byte
									key := keyBuffer[:len(positions)]
									moved := false
									for i, pos := range positions {
										dst := after(pos)
										key[i] = local[dst]
										moved = moved || dst != pos
									}
									if !moved || seen[string(key)] {
										continue
									}
									seen[string(key)] = true
									var conjugateBuffer [16]Move
									conjugate := append(conjugateBuffer[:0], r.moves...)
									conjugate = nxnAppendInverse(append(conjugate, u), r.moves)
									var rawBuffer, optimizedBuffer [64]Move
									raw := nxnCommutatorInto(rawBuffer[:0], conjugate, f.moves)
									optimized := nxnOptimizeInto(optimizedBuffer[:0], raw, n)
									a := centerBlockAction{moves: append([]Move(nil), optimized...)}
									movedCount := 0
									for i, dst := range key {
										if int(dst) != i {
											movedCount++
										}
									}
									a.trans = make([]centerTransfer, 0, movedCount)
									for i, dst := range key {
										if int(dst) != i {
											a.trans = append(a.trans, centerTransfer{uint8(i), dst})
										}
									}
									a.buildMasks(targets)
									actions = append(actions, a)
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

func nxnBulkCenters(c *Cube, t *reductionTables) ([]Move, error) {
	var positions []int
	for _, o := range t.centers {
		if err := nxnValidateCenterOrbit(c, o); err != nil {
			return nil, err
		}
		positions = append(positions, o.positions...)
	}
	home := NewCube(c.Size)
	colors, targets := make([]Color, len(positions)), make([]Color, len(positions))
	local := make(map[int]int)
	for i, pos := range positions {
		colors[i], targets[i] = nxnColor(c, pos), nxnColor(home, pos)
		local[pos] = i
	}
	t.blocksOnce.Do(func() {
		t.blocks = nxnCenterBlockActions(c.Size, t)
		t.blockInfluences = nxnCenterInfluences(t.blocks)
	})
	actions := t.blocks
	var setups [][144]uint8
	setupMoves := []Move{{}}
	var identity [144]uint8
	for i := range identity {
		identity[i] = uint8(i)
	}
	setups = append(setups, identity)
	for _, m := range coordinateMoves {
		p := nxnPermutation(c.Size, []Move{m})
		var trans [144]uint8
		for i, pos := range positions {
			trans[local[p[pos]]] = uint8(i)
		}
		setups = append(setups, trans)
		setupMoves = append(setupMoves, m)
	}
	var inverses [19][144]uint8
	gains := make([]int16, len(setups)*len(actions))
	for setup, trans := range setups {
		var present [6][3]uint64
		for i := range positions {
			present[colors[trans[i]]][i/64] |= 1 << (i % 64)
			inverses[setup][trans[i]] = uint8(i)
		}
		for id := range actions {
			gains[setup*len(actions)+id] = int16(actions[id].gain(&present))
		}
	}
	var moves []Move
	for {
		bestGain, bestCost, bestAction, bestSetup := 0, 1, -1, 0
		for setup := range setups {
			for id := range actions {
				action := &actions[id]
				gain := int(gains[setup*len(actions)+id])
				cost := len(action.moves)
				if setup != 0 {
					cost++
				}
				if gain > 0 && gain*bestCost > bestGain*cost {
					bestGain, bestCost, bestAction, bestSetup = gain, cost, id, setup
				}
			}
		}
		// Individual color cycles are more economical than a block algorithm
		// gaining fewer than one correct center per four turns.
		if bestAction < 0 || bestGain*4 < bestCost {
			break
		}
		part := actions[bestAction].moves
		if bestSetup != 0 {
			part = append([]Move{setupMoves[bestSetup]}, part...)
		}
		if err := c.ApplyMoves(part); err != nil {
			return nil, err
		}
		moves = append(moves, part...)
		for i, pos := range positions {
			after := nxnColor(c, pos)
			if after == colors[i] {
				continue
			}
			for setup := range setups {
				terms := &t.blockInfluences[inverses[setup][i]]
				base := setup * len(actions)
				for _, term := range terms[colors[i]] {
					gains[base+int(term>>1)] -= 1 - 2*int16(term&1)
				}
				for _, term := range terms[after] {
					gains[base+int(term>>1)] += 1 - 2*int16(term&1)
				}
			}
			colors[i] = after
		}
	}
	for _, o := range t.centers {
		part, err := nxnColorCycles(c, t, o, []Face{Front, Back, Left, Right, Up, Down}, nil)
		if err != nil {
			return nil, err
		}
		moves = append(moves, part...)
	}
	return moves, nil
}
