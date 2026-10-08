package cube

// Adjacent turns on one axis commute, including numbered far layers and wide
// blocks. Normalize them to physical layers, cancel, then choose the shortest
// of individual slices or differences of prefix/suffix blocks. Rotations are
// boundaries because they change the frame of the following axes.
func nxnOptimizeMoves(moves []Move, n int) []Move {
	var result []Move
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
			layers := getAffectedLayers(m, n)
			for _, layer := range layers {
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
		var best []Move
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
			var blocks []Move
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
				best = blocks
			}
		}
		if len(best) > end-start {
			best = moves[start:end]
		}
		result = append(result, best...)
		start = end
	}
	return OptimizeMoves(result)
}
