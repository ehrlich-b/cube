package cube

import "time"

// Four labeled edges occupy 12P4 positions with 2^4 orientations. The three
// disjoint groups share the same transitions and have separate solved goals.
type edgePatterns struct {
	moves    []uint32
	distance [3][]uint8
}

var edgeLock = make(chan struct{}, 1)
var edgeDB *edgePatterns

func edgeRank(pos [4]uint8, flip int) int {
	x, used := 0, uint16(0)
	for i, p := range pos {
		d := int(p)
		for j := uint8(0); j < p; j++ {
			if used&(1<<j) != 0 {
				d--
			}
		}
		x = x*(12-i) + d
		used |= 1 << p
	}
	return x*16 + flip
}

func edgeUnrank(x int) ([4]uint8, int) {
	flip := x % 16
	x /= 16
	var digits [4]int
	for i := 3; i >= 0; i-- {
		digits[i] = x % (12 - i)
		x /= 12 - i
	}
	var pos [4]uint8
	used := uint16(0)
	for i, d := range digits {
		for p := uint8(0); p < 12; p++ {
			if used&(1<<p) != 0 {
				continue
			}
			if d == 0 {
				pos[i] = p
				used |= 1 << p
				break
			}
			d--
		}
	}
	return pos, flip
}

func edgeCoordinate(s cubie, group int) int {
	var pos [4]uint8
	flip := 0
	for p, id := range s.ep {
		if int(id)/4 == group {
			i := int(id) % 4
			pos[i] = uint8(p)
			flip |= int(s.eo[p]) << i
		}
	}
	return edgeRank(pos, flip)
}

func searchEdgePatterns() *edgePatterns {
	return searchEdgePatternsLimit(time.Time{})
}

func searchEdgePatternsLimit(deadline time.Time) *edgePatterns {
	if !lockSearchTables(edgeLock, deadline) {
		return nil
	}
	defer func() { <-edgeLock }()
	if edgeDB == nil {
		if cached := loadEdgePatterns(deadline); cached != nil {
			edgeDB = cached
			return edgeDB
		}
		const size = edgePatternSize
		db := &edgePatterns{moves: make([]uint32, size*18)}
		var dest [18][12]uint8
		for m, state := range cubieMoves {
			for p, source := range state.ep {
				dest[m][source] = uint8(p)
			}
		}
		for x := 0; x < size; x++ {
			if x&255 == 0 && tableDeadlineExceeded(deadline) {
				return nil
			}
			pos, flip := edgeUnrank(x)
			for m, state := range cubieMoves {
				var next [4]uint8
				f := flip
				for i, p := range pos {
					next[i] = dest[m][p]
					f ^= int(state.eo[next[i]]) << i
				}
				db.moves[x*18+m] = uint32(edgeRank(next, f))
			}
		}
		for g := range db.distance {
			d := make([]uint8, size)
			for i := range d {
				d[i] = 255
			}
			goal := edgeCoordinate(identityCubie(), g)
			d[goal] = 0
			queue := make([]uint32, 1, size)
			queue[0] = uint32(goal)
			for head := 0; head < len(queue); head++ {
				if head&1023 == 0 && tableDeadlineExceeded(deadline) {
					return nil
				}
				x := queue[head]
				for _, y := range db.moves[int(x)*18 : int(x)*18+18] {
					if d[y] == 255 {
						d[y] = d[x] + 1
						queue = append(queue, y)
					}
				}
			}
			db.distance[g] = d
		}
		if tableDeadlineExceeded(deadline) {
			return nil
		}
		edgeDB = db
		if deadline.IsZero() {
			saveEdgePatterns(db)
		}
	}
	return edgeDB
}

type exactSearch struct {
	t        *coordinateTables
	edges    *edgePatterns
	allowed  []int
	present  [18]bool
	path     []int
	deadline time.Time
	nodes    uint64
	timedOut bool
}

// With a restricted alphabet, adjacent turns are redundant only when their
// composition is identity or is itself an allowed single move.
func (s *exactSearch) skip(m, prev int) bool {
	if prev < 0 {
		return false
	}
	f, p := Face(m/3), Face(prev/3)
	if oppositeFaces(f, p) && f < p {
		return true
	}
	if f != p {
		return false
	}
	turns := (quarterTurnIndex(m) + quarterTurnIndex(prev)) % 4
	if turns == 0 {
		return true
	}
	idx := int(f)*3 + turns - 1
	return s.present[idx]
}

// Our move order is clockwise, half turn, inverse.
func quarterTurnIndex(m int) int { return m%3 + 1 }

func (s *exactSearch) dfs(co, eo, sl, cp int, e [3]int, left, level, prev int) bool {
	s.nodes++
	if s.timedOut {
		return false
	}
	if s.nodes&1023 == 0 && !s.deadline.IsZero() && time.Now().After(s.deadline) {
		s.timedOut = true
		return false
	}
	h := max(s.t.phase1Bound(co, eo, sl), int(s.t.CornerDistance[cp]))
	for g, x := range e {
		h = max(h, int(s.edges.distance[g][x]))
	}
	if h > left {
		return false
	}
	if left == 0 {
		return h == 0
	}
	for _, m := range s.allowed {
		if s.skip(m, prev) {
			continue
		}
		var next [3]int
		for g, x := range e {
			next[g] = int(s.edges.moves[x*18+m])
		}
		s.path[level] = m
		if s.dfs(int(s.t.Twist[co*18+m]), int(s.t.Flip[eo*18+m]), int(s.t.Slice[sl*18+m]), int(s.t.Corner[cp*18+m]), next, left-1, level+1, m) {
			return true
		}
	}
	return false
}

func coordinateMoveIndices(moves []Move) ([]int, bool) {
	if moves == nil {
		moves = coordinateMoves[:]
	}
	indices := make([]int, 0, len(moves))
	seen := [18]bool{}
	for _, m := range moves {
		if m.Face < Front || m.Face > Down || m.Wide || m.Layer != 0 || m.Slice != NoSlice || m.Rotation != NoRotation {
			return nil, false
		}
		index := int(m.Face) * 3
		if m.Double {
			index++
		} else if !m.Clockwise {
			index += 2
		}
		if !seen[index] {
			seen[index] = true
			indices = append(indices, index)
		}
	}
	return indices, true
}

func exactCoordinateSearch(state cubie, moves []Move, maxDepth int) ([]Move, bool) {
	result, ok, _ := exactCoordinateSearchLimit(state, moves, maxDepth, time.Time{})
	return result, ok
}

func exactCoordinateSearchLimit(state cubie, moves []Move, maxDepth int, deadline time.Time) ([]Move, bool, bool) {
	if tableDeadlineExceeded(deadline) {
		return nil, false, true
	}
	if state == identityCubie() {
		return []Move{}, true, false
	}
	indices, _ := coordinateMoveIndices(moves)
	t := solverTablesLimit(deadline)
	if t == nil {
		return nil, false, true
	}
	edges := searchEdgePatternsLimit(deadline)
	if edges == nil {
		return nil, false, true
	}
	s := exactSearch{t: t, edges: edges, allowed: indices, deadline: deadline}
	for _, m := range indices {
		s.present[m] = true
	}
	e := [3]int{edgeCoordinate(state, 0), edgeCoordinate(state, 1), edgeCoordinate(state, 2)}
	co, eo, sl, cp := state.twist(), state.flip(), state.slice(), permutationRank(state.cp[:])
	lower := max(s.t.phase1Bound(co, eo, sl), int(s.t.CornerDistance[cp]))
	for g, x := range e {
		lower = max(lower, int(s.edges.distance[g][x]))
	}
	for depth := lower; depth <= maxDepth; depth++ {
		if !deadline.IsZero() && time.Now().After(deadline) {
			return nil, false, true
		}
		s.path = make([]int, depth)
		if s.dfs(co, eo, sl, cp, e, depth, 0, -1) {
			if tableDeadlineExceeded(deadline) {
				return nil, false, true
			}
			result := make([]Move, depth)
			for i, m := range s.path {
				result[i] = coordinateMoves[m]
			}
			return result, true, false
		}
		if s.timedOut {
			return nil, false, true
		}
	}
	return nil, false, false
}
