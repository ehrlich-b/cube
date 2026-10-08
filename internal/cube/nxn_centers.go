package cube

import (
	"fmt"
	"math/bits"
	"sync"
)

// Four indistinguishable centers have just C(24,4)=10626 coordinates.
// These small, in-memory pattern databases guide searches for opposite center
// blocks on 4x4/5x5. Once U/D are complete, only outer turns and horizontal
// slices are allowed. Center cycles finish any remaining bars before pairing.
const nxnCenterCoordinates = 10626

type centerPatterns struct {
	masks     []uint32
	next      [][]uint16
	goal      [6]uint32
	distMu    sync.Mutex
	distCache map[string][]uint8
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
	db := centerPatternShape(n, t, o)
	local := make(map[int]int, 24)
	for i, pos := range o.positions {
		local[pos] = i
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

func centerPatternShape(n int, t *reductionTables, o *reductionOrbit) *centerPatterns {
	db := &centerPatterns{masks: make([]uint32, nxnCenterCoordinates), next: make([][]uint16, len(t.moves))}
	for i, pos := range o.positions {
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
	return db
}

func centerDistanceKey(face Face, allowed []int) string {
	key := make([]byte, len(allowed)+1)
	key[0] = byte(face)
	for i, g := range allowed {
		key[i+1] = byte(g)
	}
	return string(key)
}

func (db *centerPatterns) distances(face Face, allowed []int) []uint8 {
	key := centerDistanceKey(face, allowed)
	db.distMu.Lock()
	defer db.distMu.Unlock()
	if dist := db.distCache[key]; dist != nil {
		return dist
	}
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
	if db.distCache == nil {
		db.distCache = make(map[string][]uint8)
	}
	db.distCache[key] = dist
	return dist
}

type centerSearchNode struct {
	key    [8]uint16
	parent int16 // at most 1 + 36*512 nodes
	move   int16
	score  uint16 // at most 8*(255*16 + 4)
}

type centerSearchBuffers struct {
	nodes, candidates, selected []centerSearchNode
	beam                        []int
	seen, level                 centerKeySet
	offsets                     []int
}

func (b *centerSearchBuffers) init(allowed int) {
	const width = 512
	if b.nodes == nil {
		b.nodes = make([]centerSearchNode, 1, 1+width*36)
		b.selected = make([]centerSearchNode, width)
		b.beam = make([]int, 1, width)
		b.seen = newCenterKeySet(1 + width*36)
		b.offsets = make([]int, 8*(255*16+4)+1)
	}
	if cap(b.candidates) < width*allowed {
		b.candidates = make([]centerSearchNode, 0, width*allowed)
		b.level = newCenterKeySet(width * allowed)
	}
}

// Exact open addressing for beam coordinates. Small fingerprints keep most
// probes in a compact byte array; collisions still compare every coordinate.
type centerKeySet struct {
	keys [][2]uint64
	tags []uint8
}

func newCenterKeySet(max int) centerKeySet {
	size := 1
	for size < max*2 {
		size *= 2
	}
	return centerKeySet{make([][2]uint64, size), make([]uint8, size)}
}

func centerPackedKey(key [8]uint16) ([2]uint64, uint64) {
	a := uint64(key[0]) | uint64(key[1])<<14 | uint64(key[2])<<28 | uint64(key[3])<<42
	b := uint64(key[4]) | uint64(key[5])<<14 | uint64(key[6])<<28 | uint64(key[7])<<42
	h := (a ^ b*0x9e3779b97f4a7c15) * 0xbf58476d1ce4e5b9
	h = (h ^ h>>32) * 0x94d049bb133111eb
	return [2]uint64{a, b}, h ^ h>>32
}

func (s *centerKeySet) slot(key [2]uint64, hash uint64) int {
	tag := uint8(hash>>56) | 1
	at := int(hash) & (len(s.keys) - 1)
	for s.tags[at] != 0 {
		if s.tags[at] == tag && s.keys[at] == key {
			return at
		}
		at = (at + 1) & (len(s.keys) - 1)
	}
	return at
}

func (s *centerKeySet) contains(key [2]uint64, hash uint64) bool {
	return s.tags[s.slot(key, hash)] != 0
}

func (s *centerKeySet) add(key [2]uint64, hash uint64) bool {
	at := s.slot(key, hash)
	if s.tags[at] != 0 {
		return false
	}
	s.keys[at], s.tags[at] = key, uint8(hash>>56)|1
	return true
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
	var buffers centerSearchBuffers
	for _, faces := range [][]Face{{Up}, {Up, Down}, {Left}, {Left, Right}, {Left, Right, Front, Back}} {
		allowed := nxnAllowedCenterMoves(t, locked[Down])
		part := nxnSearchCentersUsing(c, t, patterns, faces, allowed, &buffers)
		if err := nxnApplyMoves(c, part); err != nil {
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

func nxnAllowedCenterMoves(t *reductionTables, lockedUD bool) []int {
	var allowed []int
	for g, m := range t.moves {
		if t.size%2 == 1 && m.Layer == t.size/2 {
			continue // Keep the fixed-center frame canonical.
		}
		if !lockedUD || !m.Wide && (m.Layer == 0 || m.Layer == t.size-1) || nxnAxis(m.Face) == nxnAxis(Up) {
			allowed = append(allowed, g)
		}
	}
	return allowed
}

// A bounded beam supplies useful partial blocks on the one or two movable
// center orbits of a 4x4/5x5, even when it cannot finish the joint coordinate. The constructive
// color-cycle finish guarantees termination without any search timeout risk.
func nxnSearchCenters(c *Cube, t *reductionTables, dbs []*centerPatterns, faces []Face, allowed []int) []Move {
	var buffers centerSearchBuffers
	return nxnSearchCentersUsing(c, t, dbs, faces, allowed, &buffers)
}

func nxnSearchCentersUsing(c *Cube, t *reductionTables, dbs []*centerPatterns, faces []Face, allowed []int, buffers *centerSearchBuffers) []Move {
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
	var scores [8][]uint16
	for i, dist := range dists {
		db := dbs[i/len(faces)]
		scores[i] = make([]uint16, nxnCenterCoordinates)
		for key, d := range dist {
			wrong := 4 - bits.OnesCount32(db.masks[key]&db.goal[faces[i%len(faces)]])
			scores[i][key] = uint16(d)*16 + uint16(wrong)
		}
	}
	score := func(key [8]uint16) uint16 {
		h := uint16(0)
		for i := range dists {
			h += scores[i][key[i]]
		}
		return h
	}
	const width = 512
	startScore := score(start)
	if startScore == 0 {
		return nil
	}
	buffers.init(len(allowed))
	transitions := make([][8][]uint16, len(allowed))
	var allowedTurns [128]bool
	for _, move := range allowed {
		allowedTurns[move] = true
	}
	canBound := true
	for g, move := range allowed {
		if move%3 != 1 && !allowedTurns[move/3*3+2-move%3] {
			canBound = false
		}
		for i := range dists {
			transitions[g][i] = dbs[i/len(faces)].next[move]
		}
	}
	nodes := buffers.nodes[:1]
	nodes[0] = centerSearchNode{key: start, parent: -1, move: -1, score: startScore}
	beam, best := buffers.beam[:1], 0
	beam[0] = 0
	seen, levelSeen := &buffers.seen, &buffers.level
	clear(seen.tags)
	startKey, startHash := centerPackedKey(start)
	seen.add(startKey, startHash)
	candidateBuffer, selected, offsets := buffers.candidates, buffers.selected, buffers.offsets
	clear(offsets)
	minWritten, maxWritten := uint16(len(offsets)-1), uint16(0)
	for depth := 0; depth < 36 && nodes[best].score != 0; depth++ {
		candidates := candidateBuffer[:0]
		clear(levelSeen.tags)
		if minWritten <= maxWritten {
			clear(offsets[minWritten : maxWritten+1])
		}
		minWritten, maxWritten = uint16(len(offsets)-1), 0
		threshold, retained := uint16(len(offsets)-1), 0
		for _, parent := range beam {
			prev := nodes[parent]
			// With inverse-closed generators, a coordinate's distance can
			// fall by at most one turn. Bound the unread coordinates once
			// per parent, then reject partial scores before more lookups.
			var lower [4]uint16
			if canBound {
				sum := uint16(0)
				for i := 0; i < len(dists)/2; i++ {
					h := scores[i][prev.key[i]] >> 4
					if h == 255 {
						sum += 255 * 16 // An unreachable component stays unreachable.
					} else if h > 0 {
						sum += (h - 1) * 16
					}
					lower[i] = sum
				}
			}
			for moveIndex, g := range allowed {
				if prev.move >= 0 && g/3 == int(prev.move)/3 {
					continue
				}
				next := centerSearchNode{parent: int16(parent), move: int16(g)}
				step := &transitions[moveIndex]
				// The normal stages have 1, 2, 4 or 8 coordinates. Fixed
				// indices avoid repeated slice/array bounds in this hot loop.
				switch len(dists) {
				case 8:
					k4, k5 := step[4][prev.key[4]], step[5][prev.key[5]]
					k6, k7 := step[6][prev.key[6]], step[7][prev.key[7]]
					next.key[4], next.key[5], next.key[6], next.key[7] = k4, k5, k6, k7
					next.score += scores[4][k4] + scores[5][k5] + scores[6][k6] + scores[7][k7]
					if retained >= width && next.score+lower[3] >= threshold {
						continue
					}
					fallthrough
				case 4:
					k2, k3 := step[2][prev.key[2]], step[3][prev.key[3]]
					next.key[2], next.key[3] = k2, k3
					next.score += scores[2][k2] + scores[3][k3]
					if retained >= width && next.score+lower[1] >= threshold {
						continue
					}
					fallthrough
				case 2:
					k1 := step[1][prev.key[1]]
					next.key[1] = k1
					next.score += scores[1][k1]
					if retained >= width && next.score+lower[0] >= threshold {
						continue
					}
					fallthrough
				case 1:
					k0 := step[0][prev.key[0]]
					next.key[0] = k0
					next.score += scores[0][k0]
				default:
					for i := range dists {
						next.key[i] = step[i][prev.key[i]]
						next.score += scores[i][next.key[i]]
					}
				}
				// A zero score has one coordinate key. Its first generated
				// path is also the first goal selected by stable beam ordering.
				if next.score == 0 {
					best = len(nodes)
					nodes = append(nodes, next)
					goto centerResult
				}
				// A coordinate's score is independent of its path. Once a
				// stable top-width prefix exists, later equal/worse scores
				// cannot enter it, so they need neither hashing nor storage.
				if retained >= width && next.score >= threshold {
					continue
				}
				key, hash := centerPackedKey(next.key)
				if seen.contains(key, hash) || !levelSeen.add(key, hash) {
					continue
				}
				candidates = append(candidates, next)
				if next.score < minWritten {
					minWritten = next.score
				}
				if next.score > maxWritten {
					maxWritten = next.score
				}
				offsets[next.score]++
				retained++
				if len(candidates) == width {
					threshold = maxWritten
				}
				for retained-offsets[threshold] >= width {
					retained -= offsets[threshold]
					threshold--
				}
			}
		}
		if len(candidates) == 0 {
			break
		}
		candidates = nxnCenterBeamInto(candidates, selected, offsets)
		beam = beam[:0]
		for _, next := range candidates {
			key, hash := centerPackedKey(next.key)
			seen.add(key, hash)
			id := len(nodes)
			nodes = append(nodes, next)
			beam = append(beam, id)
			if next.score < nodes[best].score {
				best = id
			}
		}
	}
centerResult:
	var reversed []Move
	for best != 0 {
		reversed = append(reversed, t.moves[nodes[best].move])
		best = int(nodes[best].parent)
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
func nxnCenterBeamInto(candidates, selected []centerSearchNode, offsets []int) []centerSearchNode {
	lo, hi := candidates[0].score, candidates[0].score
	for _, node := range candidates {
		if node.score < lo {
			lo = node.score
		}
		if node.score > hi {
			hi = node.score
		}
	}
	counts := offsets[lo : hi+1]
	clear(counts)
	for _, node := range candidates {
		counts[node.score-lo]++
	}
	total := 0
	for score, count := range counts {
		counts[score] = total
		total += count
	}
	if len(candidates) < len(selected) {
		selected = selected[:len(candidates)]
	}
	for _, node := range candidates {
		at := counts[node.score-lo]
		counts[node.score-lo]++
		if at < len(selected) {
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
		bestCost := 0
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
					cost := o.cycleCost(t, a, b, d)
					if bestCost == 0 || gain*bestCost > bestGain*cost {
						bestCost, bestGain, chosen = cost, gain, [3]int{a, b, d}
					}
				}
			}
		}
		if bestCost == 0 {
			return nil, fmt.Errorf("center color cycles stalled")
		}
		a, b, d := chosen[0], chosen[1], chosen[2]
		best := o.cycleMoves(t, a, b, d)
		colors[b], colors[d], colors[a] = colors[a], colors[b], colors[d]
		if err := nxnApplyMoves(c, best); err != nil {
			return nil, err
		}
		moves = append(moves, best...)
	}
}
