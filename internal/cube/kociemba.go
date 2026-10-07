package cube

import (
	"fmt"
	"time"
)

// KociembaOptions controls an anytime two-phase search. TargetLength is a
// stopping goal in face turns, not a guarantee or a proof of optimality.
// TimeLimit covers search; one-time table initialization is separate so cold
// Web Workers can finish building tables and still receive a search budget.
type KociembaOptions struct {
	TargetLength int
	TimeLimit    time.Duration
}

type phaseNode struct {
	a, b, c int
	next    int
}

// Explicit DFS stacks let all six views share the budget without restarting
// searches. A difficult phase-two endpoint cannot monopolize the entire solve.
type twoPhase struct {
	root           cubie
	faceMap        [6]Face
	inverse        bool
	path           [30]int
	one            [13]phaseNode
	two            [19]phaseNode
	depth, level   int
	depth2, level2 int
	active2        bool
}

func (s *twoPhase) startDepth(depth int) {
	s.depth, s.level = depth, 0
	s.one[0] = phaseNode{a: s.root.twist(), b: s.root.flip(), c: s.root.slice()}
}

// advance visits a bounded number of nodes, retaining both DFS cursors. The
// total bound tightens immediately when any view finds a shorter solution.
func (s *twoPhase) advance(t *coordinateTables, bound, quantum int) ([]int, bool) {
	for visited := 0; visited < quantum; visited++ {
		if s.active2 {
			if s.depth2 > min(18, bound-s.depth) {
				s.active2 = false
				continue
			}
			if s.level2 < 0 {
				s.depth2++
				s.level2 = 0
				s.two[0].next = 0
				continue
			}
			if s.level2 == s.depth2 {
				s.active2 = false
				return append([]int(nil), s.path[:s.depth+s.depth2]...), false
			}
			n := &s.two[s.level2]
			if n.next == len(phase2Moves) {
				s.level2--
				continue
			}
			m := phase2Moves[n.next]
			n.next++
			level := s.depth + s.level2
			if level > 0 && skipCoordinateFace(m, s.path[level-1]) {
				continue
			}
			a, b, c := int(t.Corner[n.a*18+m]), int(t.Edge2[n.b*18+m]), int(t.Slice2[n.c*18+m])
			if t.phase2Bound(a, b, c) > s.depth2-s.level2-1 {
				continue
			}
			s.path[level] = m
			s.level2++
			s.two[s.level2] = phaseNode{a: a, b: b, c: c}
			continue
		}
		if s.level < 0 || s.depth > bound {
			return nil, true
		}
		if s.level == s.depth {
			// A final subgroup move just duplicates an endpoint at a smaller
			// phase-one depth. Move it into phase two instead.
			s.level--
			if s.depth > 0 && isPhase2Move(s.path[s.depth-1]) {
				continue
			}
			state := s.root
			for _, m := range s.path[:s.depth] {
				state = state.mul(cubieMoves[m])
			}
			s.two[0] = phaseNode{a: permutationRank(state.cp[:]), b: permutationRank(state.ep[:8]), c: permutationRank(state.ep[8:])}
			s.depth2 = t.phase2Bound(s.two[0].a, s.two[0].b, s.two[0].c)
			s.level2, s.active2 = 0, true
			continue
		}
		n := &s.one[s.level]
		if n.next == 18 {
			s.level--
			continue
		}
		m := n.next
		n.next++
		if s.level > 0 && skipCoordinateFace(m, s.path[s.level-1]) {
			continue
		}
		a, b, c := int(t.Twist[n.a*18+m]), int(t.Flip[n.b*18+m]), int(t.Slice[n.c*18+m])
		if t.phase1Bound(a, b, c) > s.depth-s.level-1 {
			continue
		}
		s.path[s.level] = m
		s.level++
		s.one[s.level] = phaseNode{a: a, b: b, c: c}
	}
	return nil, false
}

func isPhase2Move(m int) bool { return m >= 12 || m%3 == 1 }

// Search the U/D, F/B and L/R reductions, each in both directions. Rotating
// positions AND relabeling colors conjugates the state; faceMap converts the
// answer back without adding search-orientation rotations to the solution.
func twoPhaseViews(c *Cube, t *coordinateTables) [6]twoPhase {
	var views [6]twoPhase
	home := NewCube(3)
	for axis, rotation := range []RotationType{NoRotation, X_Rotation, Z_Rotation} {
		work, frame := c.clone(), home.clone()
		if rotation != NoRotation {
			m := Move{Rotation: rotation, Clockwise: true}
			work.ApplyMove(m)
			frame.ApplyMove(m)
		}
		var recolor [6]Color
		var faces [6]Face
		for f := Front; f <= Down; f++ {
			color := frame.Faces[f][1][1]
			recolor[color] = home.Faces[f][1][1]
			for original := Front; original <= Down; original++ {
				if home.Faces[original][1][1] == color {
					faces[f] = original
				}
			}
		}
		for f := range work.Faces {
			for row := range work.Faces[f] {
				for col, color := range work.Faces[f][row] {
					work.Faces[f][row][col] = recolor[color]
				}
			}
		}
		state := readCubie(work)
		for direction := 0; direction < 2; direction++ {
			view := &views[axis*2+direction]
			view.root, view.faceMap, view.inverse = state, faces, direction == 1
			if view.inverse {
				view.root = state.inverse()
			}
			view.startDepth(t.phase1Bound(view.root.twist(), view.root.flip(), view.root.slice()))
		}
	}
	return views
}

func (s *twoPhase) moves(path []int) []Move {
	moves := make([]Move, len(path))
	for i, m := range path {
		move := coordinateMoves[m]
		move.Face = s.faceMap[move.Face]
		moves[i] = move
	}
	if s.inverse {
		return inverseSequence(moves)
	}
	return moves
}

func solveTwoPhase(c *Cube) (*SolverResult, error) {
	return SolveKociemba(c, KociembaOptions{TargetLength: 20, TimeLimit: time.Second})
}

// SolveKociemba returns the best verified solution found before the search
// budget expires, or stops early on reaching TargetLength. If the budget is
// too short to find even one solution, it returns an explicit timeout error.
func SolveKociemba(c *Cube, options KociembaOptions) (*SolverResult, error) {
	start := time.Now()
	if options.TargetLength < 1 || options.TargetLength > 30 {
		return nil, fmt.Errorf("Kociemba target length must be between 1 and 30")
	}
	if options.TimeLimit <= 0 {
		return nil, fmt.Errorf("Kociemba time limit must be positive")
	}
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
	t := solverTables()
	views := twoPhaseViews(work, t)
	deadline := time.Now().Add(options.TimeLimit)
	bound := 30
	var best []Move
search:
	for {
		active := false
		for i := range views {
			if !time.Now().Before(deadline) {
				break search
			}
			s := &views[i]
			if s.depth > min(12, bound) {
				continue
			}
			active = true
			path, done := s.advance(t, bound, 1024)
			if path != nil {
				best = s.moves(path)
				bound = len(path) - 1
				if len(path) <= options.TargetLength {
					break search
				}
			}
			if done {
				s.startDepth(s.depth + 1)
			}
		}
		if !active {
			break
		}
	}
	if best == nil {
		return nil, fmt.Errorf("Kociemba time limit exceeded (%s) before finding a solution", options.TimeLimit)
	}
	moves := append(compactGrip(rotations), best...)
	check := c.clone()
	check.ApplyMoves(moves)
	if !check.IsSolved() {
		return nil, fmt.Errorf("Kociemba solution failed full-cube verification")
	}
	return &SolverResult{Solution: moves, Steps: len(moves), Duration: time.Since(start)}, nil
}
