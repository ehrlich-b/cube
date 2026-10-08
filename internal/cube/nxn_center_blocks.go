package cube

type centerTransfer struct{ src, dst uint8 }
type centerBlockAction struct {
	moves []Move
	trans []centerTransfer
}

// Generalize a center commutator from one intersecting row/column to entire
// bars. Wings may move during this stage; the fixed-center frame is preserved.
// Full permutations include oblique orbits and central rows of odd cubes.
func nxnCenterBlockActions(n int, t *reductionTables) []centerBlockAction {
	var positions []int
	local := make(map[int]uint8)
	for _, o := range t.centers {
		for _, pos := range o.positions {
			local[pos] = uint8(len(positions))
			positions = append(positions, pos)
		}
	}
	seen := map[string]bool{}
	var actions []centerBlockAction
	for axis := 0; axis < 3; axis++ {
		rFace := []Face{Right, Up, Front}[axis]
		uFace := []Face{Up, Front, Right}[axis]
		fFace := []Face{Front, Right, Up}[axis]
		for ra := 1; ra < n-1; ra++ {
			for rb := ra; rb < n-1; rb++ {
				for fa := 1; fa < n-1; fa++ {
					for fb := fa; fb < n-1; fb++ {
						for _, rq := range []int{1, 3} {
							var r []Move
							for layer := ra; layer <= rb; layer++ {
								r = append(r, Move{Face: rFace, Layer: layer, Clockwise: rq == 1})
							}
							r = nxnOptimizeMoves(r, n)
							for _, fq := range []int{1, 3} {
								var f []Move
								for layer := fa; layer <= fb; layer++ {
									f = append(f, Move{Face: fFace, Layer: layer, Clockwise: fq == 1})
								}
								f = nxnOptimizeMoves(f, n)
								for _, uf := range []Face{uFace, []Face{Down, Back, Left}[axis]} {
									for _, u := range faceMoves(uf) {
										conjugate := append(append(append([]Move{}, r...), u), nxnInverse(r)...)
										moves := nxnCommutator(conjugate, f)
										p := nxnPermutation(n, moves)
										valid := true
										for src, dst := range p {
											if src != dst {
												face, row, col := indexToCoord(src, n)
												if _, ok := local[src]; !ok && !isBoundaryCell(face, row, col, n) {
													valid = false
													break
												}
											}
										}
										if !valid {
											continue
										}
										key := make([]byte, len(positions))
										a := centerBlockAction{moves: nxnOptimizeMoves(moves, n)}
										for i, pos := range positions {
											key[i] = local[p[pos]]
											if p[pos] != pos {
												a.trans = append(a.trans, centerTransfer{uint8(i), key[i]})
											}
										}
										if len(a.trans) != 0 && !seen[string(key)] {
											seen[string(key)] = true
											actions = append(actions, a)
										}
									}
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
	t.blocksOnce.Do(func() { t.blocks = nxnCenterBlockActions(c.Size, t) })
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
	var moves []Move
	for {
		bestGain, bestCost, bestAction, bestSetup := 0, 1, -1, 0
		for setup, trans := range setups {
			for id, action := range actions {
				gain := 0
				for _, step := range action.trans {
					if colors[trans[step.src]] == targets[step.dst] {
						gain++
					}
					if colors[trans[step.dst]] == targets[step.dst] {
						gain--
					}
				}
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
			colors[i] = nxnColor(c, pos)
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
