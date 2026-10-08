package cube

import (
	"math/bits"
	"sync/atomic"
	"time"
)

type axisCoordinate struct{ co, eo, sl int }

type largeOptimalSearch struct {
	t         *coordinateTables
	db        *optimalPatterns
	phase1    *phase1Patterns
	huge      *optimalPhase1
	sorted    [21][3]int
	hugeBound [21][3]int
	axes      [21][3]axisCoordinate
	moveMap   [3][18]int
	path      [21]int
	deadline  time.Time
	nodes     uint64
	timedOut  bool
	rootMask  int
	stop      *atomic.Bool
	cancelled bool
}

func (s *largeOptimalSearch) initializeAxes(state cubie) {
	for axis, rotation := range []RotationType{NoRotation, X_Rotation, Z_Rotation} {
		frame := NewCube(3)
		if rotation != NoRotation {
			frame.ApplyMove(Move{Rotation: rotation, Clockwise: true})
		}
		sym := readCubie(frame)
		inv := sym.inverse()
		view := inv.mul(state).mul(sym)
		s.axes[0][axis] = axisCoordinate{view.twist(), view.flip(), view.slice()}
		if s.huge != nil {
			s.sorted[0][axis] = sliceSorted(view)
			s.hugeBound[0][axis] = s.huge.rootBound(view.twist(), view.flip(), s.sorted[0][axis])
			if s.hugeBound[0][axis] < 0 {
				s.timedOut = true
			}
		}
		for m, move := range cubieMoves {
			transformed := inv.mul(move).mul(sym)
			for n, candidate := range cubieMoves {
				if transformed == candidate {
					s.moveMap[axis][m] = n
					break
				}
			}
		}
	}
}

func (s *largeOptimalSearch) bound(cp, e0, e1, level int) int {
	root := s.axes[level][0]
	h := max(nibbleDistance(s.db.corners, cp*2187+root.co), nibbleDistance(s.db.edges[0], e0), nibbleDistance(s.db.edges[1], e1))
	if s.huge != nil {
		return max(h, threeAxisBound(s.hugeBound[level]))
	}
	if s.phase1 == nil {
		return max(h, s.t.phase1Bound(root.co, root.eo, root.sl))
	}
	var bounds [3]int
	for i, a := range s.axes[level] {
		bounds[i] = s.phase1.bound(a.co, a.eo, a.sl)
	}
	return max(h, threeAxisBound(bounds))
}

// Every face turn preserves one of the three phase-one goal subgroups.
// If all three admissible bounds reach n>0, a length-n solution is impossible:
// its first turn from the goal would leave one axis at distance at most n-1.
func threeAxisBound(h [3]int) int {
	bound := max(h[0], h[1], h[2])
	if h[0] > 0 && h[0] == h[1] && h[1] == h[2] {
		bound++
	}
	return bound
}

func (s *largeOptimalSearch) dfs(cp, e0, e1, left, level, prev int) bool {
	s.nodes++
	if s.nodes&1023 == 0 {
		if s.stop != nil && s.stop.Load() {
			s.cancelled = true
			return false
		}
		if tableDeadlineExceeded(s.deadline) {
			s.timedOut = true
		}
	}
	co := s.axes[level][0].co
	if s.timedOut || s.cancelled || nibbleDistance(s.db.corners, cp*2187+co) > left || nibbleDistance(s.db.edges[0], e0) > left || nibbleDistance(s.db.edges[1], e1) > left {
		return false
	}
	// With the large phase-one table every child was checked in all three
	// frames before recursion. Avoid repeating those cache reads here.
	if s.phase1 == nil && s.huge == nil {
		a := s.axes[level][0]
		if s.t.phase1Bound(a.co, a.eo, a.sl) > left {
			return false
		}
	}
	if left == 0 {
		return true
	}
	count := 1
	if s.phase1 != nil || s.huge != nil {
		count = 3
	}
	candidates := phase1Candidates[prev]
	if level == 0 && s.rootMask != 0 {
		candidates &= s.rootMask
	}
	for ; candidates != 0; candidates &= candidates - 1 {
		m := bits.TrailingZeros(uint(candidates))
		pruned := false
		var bounds [3]int
		for axis := 0; axis < count; axis++ {
			a := s.axes[level][axis]
			mapped := s.moveMap[axis][m]
			next := axisCoordinate{int(s.t.Twist[a.co*18+mapped]), int(s.t.Flip[a.eo*18+mapped]), int(s.t.Slice[a.sl*18+mapped])}
			s.axes[level+1][axis] = next
			if s.huge != nil {
				sorted := int(s.huge.sortedMove[s.sorted[level][axis]*18+mapped])
				s.sorted[level+1][axis] = sorted
				bounds[axis] = s.huge.childBound(next.co, next.eo, sorted, s.hugeBound[level][axis])
				s.hugeBound[level+1][axis] = bounds[axis]
			} else if s.phase1 != nil {
				bounds[axis] = s.phase1.bound(next.co, next.eo, next.sl)
			}
			if bounds[axis] > left-1 {
				pruned = true
				break
			}
		}
		if pruned || threeAxisBound(bounds) > left-1 {
			continue
		}
		s.path[level] = m
		a := int(s.db.moves[e0/64*18+m]) ^ (e0 & 63)
		b := int(s.db.moves[e1/64*18+m]) ^ (e1 & 63)
		if s.dfs(int(s.t.Corner[cp*18+m]), a, b, left-1, level+1, m) {
			return true
		}
		if s.timedOut || s.cancelled {
			return false
		}
	}
	return false
}

// At most two workers share immutable tables and partition the eighteen
// root moves. A depth is disproved only when both partitions finish. All
// smaller depths are still exhausted before a solution can be returned.
func (s *largeOptimalSearch) searchDepth(cp, e0, e1, depth int, parallel bool) bool {
	if !parallel || depth < 13 {
		return s.dfs(cp, e0, e1, depth, 0, 18)
	}
	type answer struct {
		worker largeOptimalSearch
		found  bool
	}
	answers := make(chan answer, 2)
	stop := &atomic.Bool{}
	for parity := 0; parity < 2; parity++ {
		worker := *s
		worker.nodes, worker.rootMask, worker.stop = 0, 0, stop
		for m := parity; m < 18; m += 2 {
			worker.rootMask |= 1 << m
		}
		go func(worker largeOptimalSearch) {
			found := worker.dfs(cp, e0, e1, depth, 0, 18)
			if found {
				stop.Store(true)
			}
			answers <- answer{worker, found}
		}(worker)
	}
	found := false
	for i := 0; i < 2; i++ {
		result := <-answers
		s.nodes += result.worker.nodes
		s.timedOut = s.timedOut || result.worker.timedOut
		if result.found {
			s.path = result.worker.path
			found = true
		}
	}
	return found
}

func largeOptimalSearchLimit(state cubie, t *coordinateTables, db *optimalPatterns, deadline time.Time) ([]Move, bool, bool) {
	s := largeOptimalSearch{t: t, db: db, phase1: phase1LargeDB.Load(), huge: optimalPhase1DB.Load(), deadline: deadline}
	s.initializeAxes(state)
	cp := permutationRank(state.cp[:])
	e0, e1 := sixEdgeCoordinate(state, 0), sixEdgeCoordinate(state, 1)
	forwardBound := s.bound(cp, e0, e1, 0)
	inverse := state.inverse()
	s.initializeAxes(inverse)
	icp := permutationRank(inverse.cp[:])
	i0, i1 := sixEdgeCoordinate(inverse, 0), sixEdgeCoordinate(inverse, 1)
	inverted := s.bound(icp, i0, i1, 0) > forwardBound
	if inverted {
		state, cp, e0, e1 = inverse, icp, i0, i1
	}
	s.initializeAxes(state)
	if s.timedOut {
		return nil, false, true
	}
	for depth := s.bound(cp, e0, e1, 0); depth <= 20; depth++ {
		if tableDeadlineExceeded(deadline) {
			return nil, false, true
		}
		if s.searchDepth(cp, e0, e1, depth, true) {
			if tableDeadlineExceeded(deadline) {
				return nil, false, true
			}
			moves := make([]Move, depth)
			for i := range moves {
				moves[i] = coordinateMoves[s.path[i]]
			}
			if inverted {
				moves = inverseSequence(moves)
			}
			return moves, true, false
		}
		if s.timedOut {
			return nil, false, true
		}
	}
	return nil, false, false
}

// This exact state has a published 20-FTM lower bound (Michael Reid, 1995),
// rather than a new exhaustive search at runtime. The 20-turn upper-bound
// witness is replayed by SolveOptimal and by the independent physical oracle.
// https://www.math.rwth-aachen.de/~Martin.Schoenert/Cube-Lovers/michael_reid__superflip_requires_20_face_turns.html
func certifiedSuperflip(state cubie) ([]Move, bool) {
	goal := identityCubie()
	for i := range goal.eo {
		goal.eo[i] = 1
	}
	if state != goal {
		return nil, false
	}
	moves, err := ParseMoves("U R2 F B R B2 R U2 L B2 R U' D' R2 F R' L B2 U2 F2")
	if err != nil {
		return nil, false
	}
	return inverseSequence(moves), true
}
