package cube

// Adjacent turns on one axis commute, including numbered far layers and wide
// blocks. Normalize them to physical layers, cancel, then choose the shortest
// of individual slices or differences of prefix/suffix blocks. Rotations are
// boundaries because they change the frame of the following axes.
func nxnOptimizeMoves(moves []Move, n int) []Move {
	return nxnOptimizeInto(make([]Move, 0, len(moves)), moves, n)
}

// Callers comparing cycle costs can supply a stack buffer instead of allocating
// a move list for every candidate. Input and output must not overlap.
func nxnOptimizeInto(result []Move, moves []Move, n int) []Move {
	for start := 0; start < len(moves); {
		first := moves[start]
		if first.Rotation != NoRotation || first.Slice != NoSlice {
			result = append(result, first)
			start++
			continue
		}
		axis := nxnAxis(first.Face)
		var turns [7]int
		end := start
		for end < len(moves) {
			m := moves[end]
			if m.Rotation != NoRotation || m.Slice != NoSlice || nxnAxis(m.Face) != axis {
				break
			}
			q := moveToQuarterTurns(m)
			positive := m.Face == Right || m.Face == Up || m.Face == Front
			if !positive {
				q = (4 - q) % 4
			}
			lo, hi := m.Layer, m.Layer+1
			if m.Wide {
				lo, hi = 0, m.WideDepth
				if hi <= 0 {
					hi = 2
				}
			}
			for at := lo; at < hi; at++ {
				layer := at
				if !positive {
					layer = n - 1 - layer
				}
				turns[layer] = (turns[layer] + q) % 4
			}
			end++
		}
		face := []Face{Right, Up, Front}[axis]
		opposite := []Face{Left, Down, Back}[axis]
		makeTurn := func(layer, depth, q int, reverse bool) Move {
			f := face
			if reverse {
				f, q = opposite, (4-q)%4
			}
			m := Move{Face: f, Layer: layer, Clockwise: q == 1, Double: q == 2}
			if depth > 1 {
				m.Layer, m.Wide, m.WideDepth = 0, true, depth
			}
			if depth == n {
				m = Move{Rotation: []RotationType{X_Rotation, Y_Rotation, Z_Rotation}[axis], Clockwise: q == 1, Double: q == 2}
				if reverse {
					m.Clockwise = q == 3
				}
			}
			return m
		}
		var bestBuffer, blockBuffer [7]Move
		best := bestBuffer[:0]
		for layer, q := range turns[:n] {
			if q != 0 {
				if layer >= n/2 {
					best = append(best, makeTurn(n-1-layer, 1, q, true))
				} else {
					best = append(best, makeTurn(layer, 1, q, false))
				}
			}
		}
		for _, reverse := range []bool{false, true} {
			blocks := blockBuffer[:0]
			for depth := n; depth > 0; depth-- {
				at, beyond := depth-1, depth
				if reverse {
					at, beyond = n-depth, n-depth-1
				}
				q := turns[at]
				if beyond >= 0 && beyond < n {
					q = (q - turns[beyond] + 4) % 4
				}
				if q != 0 {
					blocks = append(blocks, makeTurn(0, depth, q, reverse))
				}
			}
			if len(blocks) < len(best) {
				best = append(best[:0], blocks...)
			}
		}
		if len(best) > end-start {
			best = moves[start:end]
		}
		result = append(result, best...)
		start = end
	}
	// Fold adjacent turns in place, including axes brought together by a
	// cancellation. This is the same final pass as OptimizeMoves.
	write := 0
	for _, m := range result {
		if write > 0 {
			last := result[write-1]
			sameRotation := m.Rotation != NoRotation && last.Rotation == m.Rotation
			sameFace := last.Face == m.Face && last.Rotation == NoRotation && m.Rotation == NoRotation &&
				last.Wide == m.Wide && last.WideDepth == m.WideDepth && last.Layer == m.Layer &&
				last.Slice == NoSlice && m.Slice == NoSlice
			if sameRotation || sameFace {
				q := (moveToQuarterTurns(last) + moveToQuarterTurns(m)) % 4
				if q == 0 {
					write--
				} else {
					last.Clockwise, last.Double = q != 3, q == 2
					result[write-1] = last
				}
				continue
			}
		}
		result[write] = m
		write++
	}
	return result[:write]
}
