package cube

import (
	"math/bits"
	"time"
)

type axisCoordinate struct{ co, eo, sl int }

type largeOptimalSearch struct {
	t        *coordinateTables
	db       *optimalPatterns
	phase1   *phase1Patterns
	axes     [21][3]axisCoordinate
	moveMap  [3][18]int
	path     [21]int
	deadline time.Time
	nodes    uint64
	timedOut bool
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
	if s.phase1 == nil {
		return max(h, s.t.phase1Bound(root.co, root.eo, root.sl))
	}
	for _, a := range s.axes[level] {
		h = max(h, s.phase1.bound(a.co, a.eo, a.sl))
	}
	return h
}

func (s *largeOptimalSearch) dfs(cp, e0, e1, left, level, prev int) bool {
	s.nodes++
	if s.nodes&1023 == 0 && tableDeadlineExceeded(s.deadline) {
		s.timedOut = true
	}
	co := s.axes[level][0].co
	if s.timedOut || nibbleDistance(s.db.corners, cp*2187+co) > left || nibbleDistance(s.db.edges[0], e0) > left || nibbleDistance(s.db.edges[1], e1) > left {
		return false
	}
	// With the large phase-one table every child was checked in all three
	// frames before recursion. Avoid repeating those cache reads here.
	if s.phase1 == nil {
		a := s.axes[level][0]
		if s.t.phase1Bound(a.co, a.eo, a.sl) > left {
			return false
		}
	}
	if left == 0 {
		return true
	}
	count := 1
	if s.phase1 != nil {
		count = 3
	}
	for candidates := phase1Candidates[prev]; candidates != 0; candidates &= candidates - 1 {
		m := bits.TrailingZeros(uint(candidates))
		pruned := false
		for axis := 0; axis < count; axis++ {
			a := s.axes[level][axis]
			mapped := s.moveMap[axis][m]
			next := axisCoordinate{int(s.t.Twist[a.co*18+mapped]), int(s.t.Flip[a.eo*18+mapped]), int(s.t.Slice[a.sl*18+mapped])}
			s.axes[level+1][axis] = next
			if s.phase1 != nil && s.phase1.bound(next.co, next.eo, next.sl) > left-1 {
				pruned = true
				break
			}
		}
		if pruned {
			continue
		}
		s.path[level] = m
		a := int(s.db.moves[e0/64*18+m]) ^ (e0 & 63)
		b := int(s.db.moves[e1/64*18+m]) ^ (e1 & 63)
		if s.dfs(int(s.t.Corner[cp*18+m]), a, b, left-1, level+1, m) {
			return true
		}
		if s.timedOut {
			return false
		}
	}
	return false
}

func largeOptimalSearchLimit(state cubie, t *coordinateTables, db *optimalPatterns, deadline time.Time) ([]Move, bool, bool) {
	s := largeOptimalSearch{t: t, db: db, phase1: phase1LargeDB.Load(), deadline: deadline}
	s.initializeAxes(state)
	cp := permutationRank(state.cp[:])
	e0, e1 := sixEdgeCoordinate(state, 0), sixEdgeCoordinate(state, 1)
	for depth := s.bound(cp, e0, e1, 0); depth <= 20; depth++ {
		if tableDeadlineExceeded(deadline) {
			return nil, false, true
		}
		if s.dfs(cp, e0, e1, depth, 0, 18) {
			if tableDeadlineExceeded(deadline) {
				return nil, false, true
			}
			moves := make([]Move, depth)
			for i := range moves {
				moves[i] = coordinateMoves[s.path[i]]
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
