package cube

import (
	"fmt"
	"time"
)

// twoPhase searches multiple phase-one endpoints, rather than accepting the
// first reduction. Phase two uses only U/D turns and side-face half turns.
type twoPhase struct {
	t        *coordinateTables
	root     cubie
	path     [30]int
	length   int
	limit    int
	deadline time.Time
	nodes    uint64
	timedOut bool
}

func (s *twoPhase) phase1(co, eo, sl, left, level, prev int) bool {
	if s.expired() {
		return false
	}
	if s.t.phase1Bound(co, eo, sl) > left {
		return false
	}
	if left == 0 {
		if co != 0 || eo != 0 || sl != 0 {
			return false
		}
		state := s.root
		for _, m := range s.path[:level] {
			state = state.mul(cubieMoves[m])
		}
		cp, ep, sp := permutationRank(state.cp[:]), permutationRank(state.ep[:8]), permutationRank(state.ep[8:])
		limit := min(18, s.limit-level)
		for depth := s.t.phase2Bound(cp, ep, sp); depth <= limit; depth++ {
			if s.phase2(cp, ep, sp, depth, level, prev) {
				s.length = level + depth
				return true
			}
		}
		return false
	}
	for m := range cubieMoves {
		if skipCoordinateFace(m, prev) {
			continue
		}
		s.path[level] = m
		if s.phase1(int(s.t.Twist[co*18+m]), int(s.t.Flip[eo*18+m]), int(s.t.Slice[sl*18+m]), left-1, level+1, m) {
			return true
		}
	}
	return false
}

func (s *twoPhase) phase2(cp, ep, sp, left, level, prev int) bool {
	if s.expired() {
		return false
	}
	if s.t.phase2Bound(cp, ep, sp) > left {
		return false
	}
	if left == 0 {
		return cp == 0 && ep == 0 && sp == 0
	}
	for _, m := range phase2Moves {
		if skipCoordinateFace(m, prev) {
			continue
		}
		s.path[level] = m
		if s.phase2(int(s.t.Corner[cp*18+m]), int(s.t.Edge2[ep*18+m]), int(s.t.Slice2[sp*18+m]), left-1, level+1, m) {
			return true
		}
	}
	return false
}

func (s *twoPhase) expired() bool {
	s.nodes++
	if s.timedOut {
		return true
	}
	if s.nodes&4095 == 0 && !s.deadline.IsZero() && time.Now().After(s.deadline) {
		s.timedOut = true
	}
	return s.timedOut
}

func solveTwoPhase(c *Cube) (*SolverResult, error) {
	start := time.Now()
	if err := Validate3x3(c); err != nil {
		return nil, err
	}
	if c.IsSolved() {
		return &SolverResult{Solution: []Move{}, Duration: time.Since(start)}, nil
	}
	work, rotations, err := canonical3x3(c)
	if err != nil {
		return nil, err
	}
	state := readCubie(work)
	s := twoPhase{t: solverTables(), root: state, limit: 22}
	s.deadline = time.Now().Add(time.Second)
	co, eo, sl := state.twist(), state.flip(), state.slice()
	// Prefer <=22 turns for a bounded search, then retain completeness through
	// the full phase-one (12) + phase-two (18) diameter bounds. Hard cases never
	// become empty successes or fail just because the preferred bound expired.
	for pass := 0; pass < 2; pass++ {
		for depth := s.t.phase1Bound(co, eo, sl); depth <= 12; depth++ {
			if s.phase1(co, eo, sl, depth, 0, -1) {
				moves := compactGrip(rotations)
				for _, m := range s.path[:s.length] {
					moves = append(moves, coordinateMoves[m])
				}
				check := c.clone()
				check.ApplyMoves(moves)
				if !check.IsSolved() {
					return nil, fmt.Errorf("Kociemba solution failed full-cube verification")
				}
				return &SolverResult{Solution: moves, Steps: len(moves), Duration: time.Since(start)}, nil
			}
			if s.timedOut {
				break
			}
		}
		if pass == 0 {
			s = twoPhase{t: s.t, root: state, limit: 30}
		}
	}
	return nil, fmt.Errorf("no two-phase solution within 30 face turns")
}
