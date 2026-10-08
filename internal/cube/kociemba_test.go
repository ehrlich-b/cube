package cube

import (
	"math/rand"
	"strings"
	"testing"
	"time"
)

func TestTwoPhaseViewsAndResumption(t *testing.T) {
	tables := solverTables()
	for _, text := range []string{"R U F2 L' B", "F U' R D B2 L", "R2 U D' F2"} {
		c := NewCube(3)
		scramble, _ := ParseMoves(text)
		c.ApplyMoves(scramble)
		views := twoPhaseViews(c, tables)
		for i := range views {
			s := &views[i]
			var path []int
			for calls := 0; calls < 1000000 && path == nil; calls++ {
				var done bool
				// Pause after every node, including within phase two.
				path, done = s.advance(tables, 30, 1)
				if done {
					s.startDepth(s.depth + 1)
				}
			}
			if path == nil {
				t.Fatalf("%s view %d did not finish", text, i)
			}
			check := c.clone()
			check.ApplyMoves(s.moves(path))
			if !check.IsSolved() {
				t.Fatalf("%s view %d failed inverse/axis replay", text, i)
			}
		}
	}
}

func TestTwoPhasePreMoveViews(t *testing.T) {
	tables := solverTables()
	c := NewCube(3)
	scramble, _ := ParseMoves("R U F2 L' B")
	c.ApplyMoves(scramble)
	views := preMoveViews(twoPhaseViews(c, tables), tables)
	if len(views) != 54 {
		t.Fatal("missing pre-move/inverse/axis views", len(views))
	}
	for i := range views {
		s := &views[i]
		var path []int
		for calls := 0; calls < 1000000 && path == nil; calls++ {
			var done bool
			path, done = s.advance(tables, 29, 1)
			if done {
				s.startDepth(s.depth + 1)
			}
		}
		if path == nil {
			t.Fatal("pre-move view did not finish", i)
		}
		check := c.clone()
		check.ApplyMoves(s.moves(path))
		if !check.IsSolved() {
			t.Fatal("pre-move/inverse/axis view failed replay", i)
		}
	}
}

func TestTwoPhaseDoublePreMoveReplay(t *testing.T) {
	tables := solverTables()
	c := NewCube(3)
	scramble, _ := ParseMoves("R U F2 L' B")
	c.ApplyMoves(scramble)
	base := twoPhaseViews(c, tables)
	for _, original := range base {
		s := original
		var path []int
		for path == nil {
			var done bool
			path, done = s.advance(tables, 30, 256)
			if done {
				s.startDepth(s.depth + 1)
			}
		}
		var only [6]twoPhase
		for i := range only {
			only[i] = original
		}
		views := extraPreMoveViews(only[:], tables)
		if len(views) != 288 {
			t.Fatal("missing double pre-moves", len(views))
		}
		for _, v := range views {
			for _, alternate := range []bool{false, true} {
				v.alternateActive = alternate
				suffix := v.preSuffix()
				witness := append([]int(nil), path...)
				for i := len(suffix) - 1; i >= 0; i-- {
					m := suffix[i]
					witness = append(witness, m/3*3+2-m%3)
				}
				state := v.root
				if alternate {
					state = cubieMoves[v.preMoves[0]/3*3+1].mul(state)
					if state.twist() != v.root.twist() || state.flip() != v.root.flip() || state.slice() != v.root.slice() {
						t.Fatal("alternate changes phase-one coordinates")
					}
				}
				for _, m := range witness {
					state = state.mul(cubieMoves[m])
				}
				if state != identityCubie() {
					t.Fatal("left multiplication suffix order")
				}
				check := c.clone()
				check.ApplyMoves(v.moves(witness))
				if !check.IsSolved() {
					t.Fatal("double pre-move/inverse/axis/alternate replay")
				}
			}
		}
	}
}

func TestPhaseTwoInverseBackjumpWitnesses(t *testing.T) {
	tables := solverTables()
	r := rand.New(rand.NewSource(2026100721))
	for trial := 0; trial < 200; trial++ {
		depth := 1 + trial%10
		state := identityCubie()
		for i := 0; i < depth; i++ {
			state = state.mul(cubieMoves[phase2Moves[r.Intn(len(phase2Moves))]])
		}
		s := twoPhase{root: state}
		s.startDepth(0)
		var path []int
		for path == nil {
			var done bool
			path, done = s.advance(tables, depth, 1)
			if done {
				t.Fatal("pruning discarded a known phase-two witness", trial, depth)
			}
		}
		if len(path) > depth {
			t.Fatal("phase-two budget exceeded")
		}
		for _, m := range path {
			state = state.mul(cubieMoves[m])
		}
		if state != identityCubie() {
			t.Fatal("phase-two backjump replay")
		}
	}
}

func TestAlternatePreMoveCursorResumption(t *testing.T) {
	tables := solverTables()
	c := NewCube(3)
	scramble, _ := ParseMoves("R U F2 L' B")
	c.ApplyMoves(scramble)
	base := twoPhaseViews(c, tables)
	views := additionalPreMoveViews(base[:], tables, 1, 6)
	if len(views) != 24 {
		t.Fatal("missing paired terminal pre-moves", len(views))
	}
	for i := 0; i < len(views); i += 4 {
		v := views[i]
		v.startDepth(tables.twoPhaseBound(v.root.twist(), v.root.flip(), v.root.slice()))
		var seen [2]bool
		for calls := 0; calls < 1_000_000 && !(seen[0] && seen[1]); calls++ {
			path, done := v.advance(tables, 29, 1)
			if done {
				v.startDepth(v.depth + 1)
			}
			if path == nil {
				continue
			}
			variant := 0
			if v.alternateActive {
				variant = 1
			}
			seen[variant] = true
			check := c.clone()
			check.ApplyMoves(v.moves(path))
			if !check.IsSolved() {
				t.Fatal("paused alternate pre-move failed replay", i, variant)
			}
		}
		if !seen[0] || !seen[1] {
			t.Fatal("alternate cursor lost a view", i, seen)
		}
	}
}

func TestPhaseOneCombinedPruning(t *testing.T) {
	tables := solverTables()
	rng := rand.New(rand.NewSource(2026100711))
	stronger := false
	for trial := 0; trial < 1000; trial++ {
		state := identityCubie()
		for depth := 0; depth < 12; depth++ {
			co, eo, sl := state.twist(), state.flip(), state.slice()
			bound := tables.phase1Bound(co, eo, sl)
			if bound > depth {
				t.Fatalf("inadmissible phase-one bound %d at depth %d", bound, depth)
			}
			old := max(tables.twistSliceBound(co, sl), tables.flipSliceBound(eo, sl))
			stronger = stronger || bound > old
			state = state.mul(cubieMoves[rng.Intn(18)])
		}
	}
	if !stronger {
		t.Fatal("twist/flip pruning never strengthens the existing bound")
	}
}

func TestKociembaBudgetAndTarget(t *testing.T) {
	solverTables()

	c := NewCube(3)
	moves, _ := ParseMoves("R U F2 L' B D2 R F U2 B' L2 U R2 D F' L B2 D' R U2")
	c.ApplyMoves(moves)
	before := readCubie(c)
	// Measure the work to the first <=20 incumbent, then expire the injected
	// clock immediately after that work. CPU scheduling cannot change the test.
	first := &kociembaSearch{softLimit: true}
	if _, err := solveKociemba(c, KociembaOptions{20, time.Second}, first); err != nil {
		t.Fatal(err)
	}
	epoch := time.Unix(0, 0)
	search := &kociembaSearch{}
	search.now = func() time.Time {
		if search.nodes >= first.nodes {
			return epoch.Add(100 * time.Millisecond)
		}
		return epoch
	}
	result, err := solveKociemba(c, KociembaOptions{1, 100 * time.Millisecond}, search)
	if err != nil {
		t.Fatal(err)
	}
	if search.nodes != first.nodes {
		t.Fatalf("expired search continued: %d operations, want %d", search.nodes, first.nodes)
	}
	if len(result.Solution) == 0 || TurnCount(result.Solution) > 20 || readCubie(c) != before {
		t.Fatal("timeout lost incumbent or mutated input")
	}
	check := c.clone()
	check.ApplyMoves(result.Solution)
	if !check.IsSolved() {
		t.Fatal("budget-expired incumbent violates solver contract")
	}
	if result, err := SolveKociemba(c, KociembaOptions{TargetLength: 20, TimeLimit: time.Nanosecond}); result != nil || err == nil || !strings.Contains(err.Error(), "before finding") {
		t.Fatal("timeout without an incumbent must be an explicit error", result, err)
	}
	for _, options := range []KociembaOptions{{0, time.Second}, {31, time.Second}, {20, 0}, {20, -time.Second}} {
		if result, err := SolveKociemba(c, options); result != nil || err == nil {
			t.Fatal("invalid options accepted", options)
		}
	}
	near := NewCube(3)
	near.ApplyMove(coordinateMoves[0])
	result, err = SolveKociemba(near, KociembaOptions{TargetLength: 1, TimeLimit: time.Second})
	if err != nil || TurnCount(result.Solution) != 1 {
		t.Fatal("one-turn target did not stop on its solution", result, err)
	}
}

func TestKociembaDeadlineCase198(t *testing.T) {
	c := NewCube(3)
	moves, err := ParseMoves("D' F2 D' R' U' F' U' B2 D F' R B2 U' R' U2 D2 B' D2 B D' R2 F' B' D' L U' B U' F U")
	if err != nil {
		t.Fatal(err)
	}
	c.ApplyMoves(moves)
	before := readCubie(c)
	epoch := time.Unix(0, 0)
	calls := 0
	now := func() time.Time {
		calls++
		if calls <= 2 {
			return epoch
		}
		if calls < 250 {
			return epoch.Add(850 * time.Millisecond)
		}
		return epoch.Add(time.Second)
	}
	// Before the fix this exact clock triggered the 80% restart at bound 30
	// and returned a verified 21-turn solution at the deadline.
	result, err := solveKociemba(c, KociembaOptions{20, time.Second}, &kociembaSearch{now: now})
	if result != nil || err == nil || !strings.Contains(err.Error(), "time limit exceeded") {
		t.Fatal("hard deadline silently missed the <=20 requirement", result, err)
	}
	for _, soft := range []bool{false, true} {
		calls = 0
		tiny := &kociembaSearch{softLimit: soft, now: func() time.Time {
			calls++
			return epoch.Add(time.Duration(calls) * time.Second)
		}}
		result, err = solveKociemba(c, KociembaOptions{20, time.Nanosecond}, tiny)
		if !soft {
			if result != nil || err == nil || !strings.Contains(err.Error(), "time limit exceeded") || tiny.nodes != 0 {
				t.Fatal("tiny hard deadline must error before search", result, err, tiny.nodes)
			}
			continue
		}
		if err != nil || TurnCount(result.Solution) > 20 {
			t.Fatal("default quality must survive a tiny soft deadline", result, err)
		}
		check := c.clone()
		check.ApplyMoves(result.Solution)
		if !check.IsSolved() || readCubie(c) != before {
			t.Fatal("default quality violated replay/input contract")
		}
		t.Logf("case 198: %d turns, %d deterministic DFS operations", TurnCount(result.Solution), tiny.nodes)
	}
}

func TestKociembaCompleteFallback(t *testing.T) {
	tables := solverTables()
	for _, text := range []string{"U R2 D' F2", "R U F'", "R U R' U' F2"} {
		c := NewCube(3)
		scramble, err := ParseMoves(text)
		if err != nil {
			t.Fatal(err)
		}
		c.ApplyMoves(scramble)
		base := twoPhaseViews(c, tables)
		var solution []Move
		for depth := 0; depth <= len(scramble) && solution == nil; depth++ {
			views := phaseOneStageViews(base, tables, 13+depth)
			if len(views) == 0 {
				if tables.twoPhaseBound(base[0].root.twist(), base[0].root.flip(), base[0].root.slice()) <= depth {
					t.Fatal("fallback omitted a feasible reduction depth")
				}
				continue
			}
			if len(views) != 1 || !views[0].complete || views[0].phaseTwoDepth(20) != 20-depth {
				t.Fatal("fallback retained a fast-search depth restriction")
			}
			s := &views[0]
			for {
				path, done := s.advance(tables, len(scramble), 1)
				if path != nil {
					solution = s.moves(path)
					break
				}
				if done {
					break
				}
			}
		}
		if solution == nil || TurnCount(solution) > len(scramble) {
			t.Fatal("complete fallback lost an existing bounded solution", text)
		}
		c.ApplyMoves(solution)
		if !c.IsSolved() {
			t.Fatal("complete fallback failed replay", text)
		}
	}
}
