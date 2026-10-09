package cube

import (
	"fmt"
	"math/bits"
	"time"
)

// KociembaOptions controls a time-limited two-phase search. TargetLength is a
// stopping goal in face turns, not a proof of optimality. Successful results
// have at most max(20, TargetLength) face turns, including on timeout.
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
	root                                       cubie
	faceMap                                    [6]Face
	inverse                                    bool
	preMoves                                   []int
	swapPre, pendingAlternate, alternateActive bool
	endpoint                                   cubie
	endpointPrev                               int
	path                                       [30]int
	one                                        [31]phaseNode
	two                                        [31]phaseNode
	depth, level                               int
	depth2, level2                             int
	active2                                    bool
	nodes                                      uint64
	complete                                   bool
}

func (s *twoPhase) startDepth(depth int) {
	s.depth, s.level = depth, 0
	s.active2, s.pendingAlternate, s.alternateActive = false, false, false
	s.one[0] = phaseNode{a: s.root.twist(), b: s.root.flip(), c: s.root.slice(), next: phase1SearchCandidates[18]}
}

// advance visits a bounded number of nodes, retaining both DFS cursors. The
// total bound tightens immediately when any view finds a shorter solution.
func (s *twoPhase) advance(t *coordinateTables, bound, quantum int) ([]int, bool) {
	for visited := 0; visited < quantum; visited++ {
		s.nodes++
		if !s.active2 && s.pendingAlternate {
			s.pendingAlternate, s.alternateActive = false, true
			state := cubieMoves[s.preMoves[0]/3*3+1].mul(s.endpoint)
			s.startPhaseTwo(t, state, s.endpointPrev, bound)
			continue
		}
		if s.active2 {
			if s.depth2 > s.phaseTwoDepth(bound) {
				s.active2 = false
				continue
			}
			if s.level2 < 0 {
				s.active2 = false
				continue
			}
			if s.level2 == s.depth2 || (s.two[s.level2].a == 0 && s.two[s.level2].b == 0 && s.two[s.level2].c == 0) {
				path := s.path[:s.depth+s.level2]
				// A half turn ending phase two can merge with the first
				// pre-move. Charge the simplified answer to the total budget.
				if s.solutionLength(path) > bound+len(s.preMoves) {
					s.level2--
					continue
				}
				s.active2 = false
				return append([]int{}, path...), false
			}
			n := &s.two[s.level2]
			if n.next == 0 {
				s.level2--
				continue
			}
			rank := bits.TrailingZeros(uint(n.next))
			m := searchMoveOrder[rank]
			n.next &= n.next - 1
			level := s.depth + s.level2
			a, b, c := int(t.Corner[n.a*18+m]), int(t.Edge2[n.b*18+m]), int(t.Slice2[n.c*18+m])
			left := s.depth2 - s.level2
			inverse := t.phase2InverseBound(a, b, c)
			if inverse > left+1 {
				// The child is one turn from its parent. An admissible bound
				// above left+1 proves the parent's entire remaining subtree
				// impossible, even when the inverse projection is inconsistent.
				// Each further ancestor adds one available turn and removes
				// one turn from this lower bound, hence two per skipped level.
				s.level2 -= min(s.level2+1, (inverse-left)/2)
				continue
			}
			if inverse >= left {
				if inverse > left {
					n.next &^= 7 << (rank / 3 * 3)
				}
				continue
			}
			if t.phase2Pruned(a, b, c, left-1) {
				continue
			}
			s.path[level] = m
			s.level2++
			s.two[s.level2] = phaseNode{a: a, b: b, c: c, next: phase2SearchCandidates[m]}
			continue
		}
		if s.level < 0 || s.depth > bound {
			return nil, true
		}
		if s.level == s.depth {
			// A final subgroup move just duplicates an endpoint at a smaller
			// phase-one depth. Move it into phase two instead.
			s.level--
			if !s.complete && s.depth > 0 && isPhase2Move(s.path[s.depth-1]) {
				continue
			}
			state := s.root
			for _, m := range s.path[:s.depth] {
				state = state.mul(cubieMoves[m])
			}
			prev := 18
			if s.depth > 0 {
				prev = s.path[s.depth-1]
			}
			s.endpoint, s.endpointPrev = state, prev
			s.pendingAlternate, s.alternateActive = s.swapPre, false
			s.startPhaseTwo(t, state, prev, bound)
			continue
		}
		n := &s.one[s.level]
		// A subgroup excursion of fewer than five turns cannot produce a
		// new phase-one endpoint with our canonical consecutive-face rule.
		// Its subgroup turns belong at the start of phase two instead.
		if !s.complete && n.a == 0 && n.b == 0 && n.c == 0 && s.depth-s.level < 5 {
			s.level--
			continue
		}
		if n.next == 0 {
			s.level--
			continue
		}
		rank := bits.TrailingZeros(uint(n.next))
		m := searchMoveOrder[rank]
		n.next &= n.next - 1
		a, b, c := int(t.Twist[n.a*18+m]), int(t.Flip[n.b*18+m]), int(t.Slice[n.c*18+m])
		h := t.phase1Prune(a, b, c, s.depth-s.level-1)
		if useLargePhase1 {
			h = t.twoPhaseBound(a, b, c)
		}
		if h > s.depth-s.level-1 {
			// Every other power on this face is one turn from this child.
			if h > s.depth-s.level {
				n.next &^= 7 << (rank / 3 * 3)
			}
			continue
		}
		s.path[s.level] = m
		s.level++
		s.one[s.level] = phaseNode{a: a, b: b, c: c, next: phase1SearchCandidates[m]}
	}
	return nil, false
}

func isPhase2Move(m int) bool { return m >= 12 || m%3 == 1 }

var phase1Candidates, phase2Candidates = func() ([19]int, [19]int) {
	var one, two [19]int
	for prev := range one {
		for m := range coordinateMoves {
			if prev != 18 && skipCoordinateFace(m, prev) {
				continue
			}
			one[prev] |= 1 << m
			if isPhase2Move(m) {
				two[prev] |= 1 << m
			}
		}
	}
	return one, two
}()

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
			view.startDepth(t.twoPhaseBound(view.root.twist(), view.root.flip(), view.root.slice()))
		}
	}
	return views
}

func (s *twoPhase) moves(path []int) []Move {
	path = append(path, s.preSuffix()...)
	moves := make([]Move, len(path))
	for i, m := range path {
		move := coordinateMoves[m]
		move.Face = s.faceMap[move.Face]
		moves[i] = move
	}
	if s.inverse {
		return OptimizeMoves(inverseSequence(moves))
	}
	return OptimizeMoves(moves)
}

func preMoveViews(base [6]twoPhase, t *coordinateTables) []twoPhase {
	views := make([]twoPhase, 0, 54)
	views = append(views, base[:]...)
	for _, s := range base {
		for m := 0; m < 12; m++ {
			if isPhase2Move(m) {
				continue
			}
			v := s
			// Left multiplication produces a solution ending in this move:
			// m * root * path = I implies root * path * m = I.
			v.root, v.preMoves = cubieMoves[m].mul(s.root), []int{m}
			v.startDepth(t.twoPhaseBound(v.root.twist(), v.root.flip(), v.root.slice()))
			views = append(views, v)
		}
	}
	return views
}

func solveTwoPhase(c *Cube) (*SolverResult, error) {
	return solveKociemba(c, KociembaOptions{TargetLength: 20, TimeLimit: time.Second}, &kociembaSearch{softLimit: true})
}

// SolveKociemba stops on reaching TargetLength or the explicit search deadline.
// At expiry it may return an incumbent only within max(20, TargetLength) face
// turns; otherwise it returns a time-limit error. Use KociembaSolver.Solve for
// the default quality search, which continues until <=20 without a deadline.
func SolveKociemba(c *Cube, options KociembaOptions) (*SolverResult, error) {
	return solveKociemba(c, options, &kociembaSearch{})
}

// Keep clock injection local to each solve so regression tests do not race with
// other callers. Nodes count DFS cursor operations, including pruning/backtrack
// work, rather than depending on elapsed time or processor speed.
type kociembaSearch struct {
	now             func() time.Time
	softLimit       bool
	requireSolution bool
	nodes           uint64
}

func solveKociemba(c *Cube, options KociembaOptions, search *kociembaSearch) (*SolverResult, error) {
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
	if certified, ok := certifiedSuperflip(readCubie(work)); ok && options.TargetLength >= 20 {
		moves := append(compactGrip(rotations), certified...)
		check := c.clone()
		check.ApplyMoves(moves)
		if !check.IsSolved() {
			return nil, fmt.Errorf("Kociemba solution failed full-cube verification")
		}
		return &SolverResult{Solution: moves, Steps: len(moves), Duration: time.Since(start)}, nil
	}
	t := solverTables()
	if useLargePhase1 {
		phase1PatternTables(t)
	}
	base := twoPhaseViews(work, t)
	now := search.now
	if now == nil {
		now = time.Now
	}
	deadline := now().Add(options.TimeLimit)
	var best []Move
	// Reduction must finish its last stage even if the optimization budget
	// expires before the first solution, as can happen on slower WASM hosts.
	expired := func() bool {
		return !search.softLimit && (!search.requireSolution || best != nil) && !now().Before(deadline)
	}
	bound := max(20, options.TargetLength)
	stage := 0
searchStages:
	for stage <= 13+max(20, options.TargetLength) {
		views := phaseOneStageViews(base, t, stage)

		pending := make([]bool, len(views))
		active := len(views)
		for i := range pending {
			pending[i] = true
		}
		for active > 0 {
			if expired() {
				break searchStages
			}
			for i := range views {
				if !pending[i] {
					continue
				}
				if expired() {
					break searchStages
				}
				s := &views[i]
				before := s.nodes
				path, done := s.advance(t, bound-len(s.preMoves), 1024)
				search.nodes += s.nodes - before
				if path != nil {
					candidate := s.moves(path)
					if best == nil || len(candidate) < len(best) {
						best = candidate
						bound = len(best) - 1
						if len(best) <= options.TargetLength {
							break searchStages
						}
					}
				}
				if done {
					pending[i] = false
					active--
				}
			}
		}
		stage++
	}
	if best == nil {
		if expired() {
			return nil, fmt.Errorf("Kociemba time limit exceeded (%s) before finding a solution within %d face turns", options.TimeLimit, max(20, options.TargetLength))
		}
		return nil, fmt.Errorf("Kociemba search exhausted without a solution within %d face turns", max(20, options.TargetLength))
	}
	moves := append(compactGrip(rotations), best...)
	check := c.clone()
	check.ApplyMoves(moves)
	if !check.IsSolved() {
		return nil, fmt.Errorf("Kociemba solution failed full-cube verification")
	}
	return &SolverResult{Solution: moves, Steps: len(moves), Duration: time.Since(start)}, nil
}

// A terminal side quarter turn keeps the reduction useful. Earlier pre-moves
// include all faces; subgroup turns before the terminal turn expose different
// piece correlations. Longer suffixes permit seven-turn phase-one reductions.
func extraPreMoveViews(base []twoPhase, t *coordinateTables) []twoPhase {
	return additionalPreMoveViews(base, t, 2, 7)
}
func additionalPreMoveViews(base []twoPhase, t *coordinateTables, count, depth int) []twoPhase {
	var views []twoPhase
	var prefix [3]int
	for _, s := range base {
		if len(s.preMoves) != 0 {
			continue
		}
		var visit func(cubie, int, int)
		visit = func(root cubie, level, prev int) {
			if level == count {
				if t.twoPhaseBound(root.twist(), root.flip(), root.slice()) > depth {
					return
				}
				moves := make([]int, count)
				for i := range moves {
					moves[i] = prefix[count-1-i]
				}
				v := twoPhase{root: root, faceMap: s.faceMap, inverse: s.inverse, preMoves: moves, swapPre: true}
				v.startDepth(depth)
				views = append(views, v)
				return
			}
			for _, m := range searchMoveOrder {
				if prev >= 0 && (Face(m/3) == Face(prev/3) || oppositeFaces(Face(m/3), Face(prev/3))) {
					continue
				}
				if level == count-1 && (isPhase2Move(m) || m%3 != 2) {
					continue
				}
				prefix[level] = m
				visit(cubieMoves[m].mul(root), level+1, m)
			}
		}
		visit(s.root, 0, -1)
	}
	return views
}

func (s *twoPhase) solutionLength(path []int) int {
	var stack [40]int
	n := 0
	appendMove := func(m int) {
		if n > 0 && stack[n-1]/3 == m/3 {
			turns := (stack[n-1]%3 + m%3 + 2) % 4
			if turns == 0 {
				n--
			} else {
				stack[n-1] = m/3*3 + turns - 1
			}
		} else {
			stack[n] = m
			n++
		}
	}
	for _, m := range path {
		appendMove(m)
	}
	for _, m := range s.preSuffix() {
		appendMove(m)
	}
	return n
}

// Use U, R, F, D, L, B turn order for the two-phase probes. Keep the public
// coordinate indices and optimal-search tables unchanged.
var searchMoveOrder = func() [18]int {
	var order [18]int
	for rank, face := range []Face{Up, Right, Front, Down, Left, Back} {
		for power := 0; power < 3; power++ {
			order[rank*3+power] = int(face)*3 + power
		}
	}
	return order
}()

var phase1SearchCandidates, phase2SearchCandidates = func() ([19]int, [19]int) {
	var one, two [19]int
	for prev := range one {
		for rank, m := range searchMoveOrder {
			if prev != 18 && skipCoordinateFace(m, prev) {
				continue
			}
			one[prev] |= 1 << rank
			if isPhase2Move(m) {
				two[prev] |= 1 << rank
			}
		}
	}
	return one, two
}()

// Synchronize phase-one work by total length, including the pre-move suffix.
// A view cannot advance to an expensive deeper reduction while another axis
// still has cheap endpoints at this length. Pre-moves keep depth at least seven.
func phaseOneStageViews(base [6]twoPhase, t *coordinateTables, stage int) []twoPhase {
	if stage > 12 {
		// The fast reductions cap each phase. If they exhaust, enumerate every
		// phase-one depth from zero through the total bound in one original
		// view, with unrestricted phase-two suffixes and no endpoint shortcuts.
		// Any solution splits at its last non-subgroup turn (or at depth zero
		// for an all-subgroup solution), so this finite fallback is complete.
		s := base[0]
		s.complete = true
		depth := stage - 13
		if t.twoPhaseBound(s.root.twist(), s.root.flip(), s.root.slice()) > depth {
			return nil
		}
		s.startDepth(depth)
		return []twoPhase{s}
	}
	var views []twoPhase
	for _, s := range base {
		if t.twoPhaseBound(s.root.twist(), s.root.flip(), s.root.slice()) <= stage {
			s.startDepth(stage)
			views = append(views, s)
		}
	}
	for count := 1; count <= min(3, stage-7); count++ {
		views = append(views, additionalPreMoveViews(base[:], t, count, stage-count)...)
	}
	return views
}

// A terminal side quarter turn and its inverse differ by a left subgroup
// half turn. That changes the phase-two permutations while leaving every
// phase-one coordinate unchanged, so share the complete phase-one DFS.
func (s *twoPhase) preSuffix() []int {
	if !s.alternateActive {
		return s.preMoves
	}
	suffix := append([]int(nil), s.preMoves...)
	m := suffix[0]
	suffix[0] = m/3*3 + 2 - m%3
	return suffix
}

func (s *twoPhase) startPhaseTwo(t *coordinateTables, state cubie, prev, bound int) {
	s.two[0] = phaseNode{a: permutationRank(state.cp[:]), b: permutationRank(state.ep[:8]), c: permutationRank(state.ep[8:]), next: phase2SearchCandidates[prev]}
	s.depth2 = s.phaseTwoDepth(bound)
	if max(t.phase2Bound(s.two[0].a, s.two[0].b, s.two[0].c), t.phase2InverseBound(s.two[0].a, s.two[0].b, s.two[0].c)) > s.depth2 {
		s.active2 = false
		return
	}
	s.level2, s.active2 = 0, true
}

func (s *twoPhase) phaseTwoDepth(bound int) int {
	if s.complete {
		return bound - s.depth
	}
	return min(12, bound-s.depth)
}
