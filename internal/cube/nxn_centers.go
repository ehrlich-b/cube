package cube

import (
	"fmt"
	"math/bits"
)

// Four indistinguishable centers have just C(24,4)=10626 coordinates.
// These small, in-memory pattern databases guide searches for opposite center
// blocks on 4x4/5x5. Once U/D are complete, only outer turns and horizontal
// slices are allowed. Center cycles finish any remaining bars before pairing.
const nxnCenterCoordinates = 10626

type centerPatterns struct {
	masks []uint32
	next  [][]uint16
	goal  [6]uint32
}

var centerChoose = func() [25][5]int {
	var c [25][5]int
	for n := range c {
		c[n][0] = 1
		for k := 1; k <= 4 && k <= n; k++ {
			c[n][k] = c[n-1][k-1] + c[n-1][k]
		}
	}
	return c
}()

func centerRank(mask uint32) uint16 {
	rank, k := 0, 1
	for mask != 0 {
		pos := bits.TrailingZeros32(mask)
		rank += centerChoose[pos][k]
		k++
		mask &= mask - 1
	}
	return uint16(rank)
}

func centerPatternTable(n int, t *reductionTables, o *reductionOrbit) *centerPatterns {
	db := &centerPatterns{masks: make([]uint32, nxnCenterCoordinates), next: make([][]uint16, len(t.moves))}
	local := make(map[int]int, 24)
	for i, pos := range o.positions {
		local[pos] = i
		db.goal[pos/(n*n)] |= 1 << i
	}
	for a := 0; a < 24; a++ {
		for b := a + 1; b < 24; b++ {
			for c := b + 1; c < 24; c++ {
				for d := c + 1; d < 24; d++ {
					mask := uint32(1<<a | 1<<b | 1<<c | 1<<d)
					db.masks[centerRank(mask)] = mask
				}
			}
		}
	}
	for g, p := range t.perms {
		var trans [24]uint32
		for i, pos := range o.positions {
			trans[i] = 1 << local[p[pos]]
		}
		db.next[g] = make([]uint16, nxnCenterCoordinates)
		for i, mask := range db.masks {
			var after uint32
			for mask != 0 {
				after |= trans[bits.TrailingZeros32(mask)]
				mask &= mask - 1
			}
			db.next[g][i] = centerRank(after)
		}
	}
	return db
}

func (db *centerPatterns) distances(face Face, allowed []int) []uint8 {
	dist := make([]uint8, nxnCenterCoordinates)
	for i := range dist {
		dist[i] = 255
	}
	root := centerRank(db.goal[face])
	dist[root] = 0
	queue := make([]uint16, 1, nxnCenterCoordinates)
	queue[0] = root
	for head := 0; head < len(queue); head++ {
		key := queue[head]
		for _, g := range allowed {
			next := db.next[g][key]
			if dist[next] == 255 {
				dist[next] = dist[key] + 1
				queue = append(queue, next)
			}
		}
	}
	return dist
}

type centerSearchNode struct {
	key    [8]uint16
	parent int
	move   int
	score  int
}

func nxnCenterBlocks(c *Cube, t *reductionTables) ([]Move, error) {
	if c.Size >= 6 {
		return nxnBulkCenters(c, t)
	}
	var moves []Move

	for _, o := range t.centers {
		if err := nxnValidateCenterOrbit(c, o); err != nil {
			return nil, err
		}
	}
	t.patternsOnce.Do(func() {
		t.patterns = make([]*centerPatterns, len(t.centers))
		for i, o := range t.centers {
			t.patterns[i] = centerPatternTable(c.Size, t, o)
		}
	})
	patterns := t.patterns

	locked := map[Face]bool{}
	for _, faces := range [][]Face{{Up}, {Up, Down}, {Left}, {Left, Right}, {Left, Right, Front, Back}} {
		var allowed []int
		for g, m := range t.moves {
			if c.Size%2 == 1 && m.Layer == c.Size/2 {
				continue // Keep the fixed-center frame canonical.
			}
			if !locked[Down] || !m.Wide && (m.Layer == 0 || m.Layer == c.Size-1) || nxnAxis(m.Face) == nxnAxis(Up) {
				allowed = append(allowed, g)
			}
		}
		part := nxnSearchCenters(c, t, patterns, faces, allowed)
		if err := c.ApplyMoves(part); err != nil {
			return nil, err
		}
		moves = append(moves, part...)
		for _, o := range t.centers {
			part, err := nxnColorCycles(c, t, o, faces, locked)
			if err != nil {
				return nil, err
			}
			moves = append(moves, part...)
		}
		if len(faces) == 2 && faces[0] == Up {
			locked[Up], locked[Down] = true, true
		}
	}
	return moves, nil
}

// A bounded beam supplies useful partial blocks on the one or two movable
// center orbits of a 4x4/5x5, even when it cannot finish the joint coordinate. The constructive
// color-cycle finish guarantees termination without any search timeout risk.
func nxnSearchCenters(c *Cube, t *reductionTables, dbs []*centerPatterns, faces []Face, allowed []int) []Move {
	var start [8]uint16
	dists := make([][]uint8, len(faces)*len(dbs))
	for orbit, o := range t.centers {
		for k, face := range faces {
			var mask uint32
			color := NewCube(c.Size).Faces[face][0][0]
			for i, pos := range o.positions {
				if nxnColor(c, pos) == color {
					mask |= 1 << i
				}
			}
			start[len(faces)*orbit+k] = centerRank(mask)
			dists[len(faces)*orbit+k] = dbs[orbit].distances(face, allowed)
		}
	}
	score := func(key [8]uint16) int {
		h, wrong := 0, 0
		for i, dist := range dists {
			h += int(dist[key[i]])
			db := dbs[i/len(faces)]
			wrong += 4 - bits.OnesCount32(db.masks[key[i]]&db.goal[faces[i%len(faces)]])
		}
		return h*16 + wrong
	}
	nodes := []centerSearchNode{{key: start, parent: -1, move: -1, score: score(start)}}
	if nodes[0].score == 0 {
		return nil
	}
	beam, best := []int{0}, 0
	seen := map[[8]uint16]bool{start: true}
	width := 512
	for depth := 0; depth < 36 && nodes[best].score != 0; depth++ {
		candidates := make([]centerSearchNode, 0, width*len(allowed))
		levelSeen := make(map[[8]uint16]bool, width*len(allowed))
		for _, parent := range beam {
			prev := nodes[parent]
			for _, g := range allowed {
				if prev.move >= 0 && g/3 == prev.move/3 {
					continue
				}
				next := centerSearchNode{parent: parent, move: g}
				for i := range dists {
					next.key[i] = dbs[i/len(faces)].next[g][prev.key[i]]
				}
				if seen[next.key] || levelSeen[next.key] {
					continue
				}
				levelSeen[next.key] = true
				next.score = score(next.key)
				candidates = append(candidates, next)
			}
		}
		if len(candidates) == 0 {
			break
		}
		candidates = nxnCenterBeam(candidates, width)
		beam = beam[:0]
		for _, next := range candidates {
			seen[next.key] = true
			id := len(nodes)
			nodes = append(nodes, next)
			beam = append(beam, id)
			if next.score < nodes[best].score {
				best = id
			}
		}
	}
	var reversed []Move
	for best != 0 {
		reversed = append(reversed, t.moves[nodes[best].move])
		best = nodes[best].parent
	}
	part := make([]Move, len(reversed))
	for i, m := range reversed {
		part[len(part)-1-i] = m
	}
	return part
}

// Integer scores allow a stable counting selection instead of sorting tens
// of thousands of nodes at every depth. Ties retain generator order, so this
// chooses exactly the same beam as a stable comparison sort.
func nxnCenterBeam(candidates []centerSearchNode, width int) []centerSearchNode {
	maxScore := 0
	for _, node := range candidates {
		if node.score > maxScore {
			maxScore = node.score
		}
	}
	offsets := make([]int, maxScore+1)
	for _, node := range candidates {
		offsets[node.score]++
	}
	total := 0
	for score, count := range offsets {
		offsets[score] = total
		total += count
	}
	if len(candidates) < width {
		width = len(candidates)
	}
	selected := make([]centerSearchNode, width)
	for _, node := range candidates {
		at := offsets[node.score]
		offsets[node.score]++
		if at < width {
			selected[at] = node
		}
	}
	return selected
}

// Centers are colors, not labeled pieces. Choose cycles by the number of
// useful placements per turn, preserving every previously completed color.
func nxnColorCycles(c *Cube, t *reductionTables, o *reductionOrbit, faces []Face, locked map[Face]bool) ([]Move, error) {
	home := NewCube(c.Size)
	var colors, target [24]Color
	var active, keep [6]bool
	for _, f := range faces {
		active[home.Faces[f][0][0]] = true
	}
	for f := range locked {
		keep[home.Faces[f][0][0]] = true
	}
	for i, pos := range o.positions {
		colors[i], target[i] = nxnColor(c, pos), nxnColor(home, pos)
	}
	var moves []Move
	for {
		wrong := 0
		for i, color := range colors {
			if active[target[i]] && color != target[i] {
				wrong++
			}
		}
		if wrong == 0 {
			return moves, nil
		}
		bestGain := 0
		var best []Move
		var chosen [3]int
		for a := 0; a < 24; a++ {
			for b := a + 1; b < 24; b++ {
				for d := a + 1; d < 24; d++ {
					if b == d {
						continue
					}
					gain, valid := 0, true
					for _, pair := range [][2]int{{a, b}, {b, d}, {d, a}} {
						src, dst := pair[0], pair[1]
						if keep[target[dst]] && colors[src] != target[dst] || keep[colors[src]] && colors[src] != target[dst] {
							valid = false
							break
						}
						if active[target[dst]] {
							if colors[src] == target[dst] {
								gain++
							}
							if colors[dst] == target[dst] {
								gain--
							}
						}
					}
					if !valid || gain <= 0 {
						continue
					}
					part := o.cycleMoves(t, a, b, d)
					if best == nil || gain*len(best) > bestGain*len(part) {
						best, bestGain, chosen = part, gain, [3]int{a, b, d}
					}
				}
			}
		}
		if best == nil {
			return nil, fmt.Errorf("center color cycles stalled")
		}
		a, b, d := chosen[0], chosen[1], chosen[2]
		colors[b], colors[d], colors[a] = colors[a], colors[b], colors[d]
		if err := c.ApplyMoves(best); err != nil {
			return nil, err
		}
		moves = append(moves, best...)
	}
}
