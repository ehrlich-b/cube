package cube

import (
	"math/rand"
	"os"
	"strconv"
	"testing"
	"time"
)

func TestSixEdgeCoordinate(t *testing.T) {
	for x := 0; x < sixEdgePermutations; x++ {
		if sixEdgeRank(sixEdgeUnrank(x)) != x {
			t.Fatal("six-edge rank round trip", x)
		}
	}
	if sixEdgeMoves(time.Now().Add(-time.Second)) != nil {
		t.Fatal("expired transition generation")
	}
}

func TestCertifiedSuperflip(t *testing.T) {
	state := identityCubie()
	for i := range state.eo {
		state.eo[i] = 1
	}
	for _, grip := range []string{"", "x", "y z'", "x2 y'"} {
		c := cubeFromCoordinates(state)
		rotations, _ := ParseMoves(grip)
		c.ApplyMoves(rotations)
		result, err := SolveOptimal(c, time.Second)
		if err != nil {
			t.Fatal(err)
		}
		if TurnCount(result.Solution) != 20 {
			t.Fatal("superflip must have distance 20", result.Solution)
		}
		c.ApplyMoves(result.Solution)
		if !c.IsSolved() {
			t.Fatal("superflip certificate fails sticker replay", grip)
		}
	}
	for _, different := range []cubie{identityCubie(), state.mul(cubieMoves[0])} {
		if _, ok := certifiedSuperflip(different); ok {
			t.Fatal("certificate accepted a different state")
		}
	}
}

func TestOptimalAxisCoordinateTransitions(t *testing.T) {
	tables := solverTables()
	r := rand.New(rand.NewSource(2026100718))
	for trial := 0; trial < 200; trial++ {
		state := uniformCubie(r)
		s := largeOptimalSearch{t: tables}
		s.initializeAxes(state)
		views := twoPhaseViews(cubeFromCoordinates(state), tables)
		for axis := 0; axis < 3; axis++ {
			root := views[axis*2].root
			inverse := root.inverse()
			co, eo, sorted := s.inverseAxis(state.inverse(), axis)
			if co != inverse.twist() || eo != inverse.flip() || sorted != sliceSorted(inverse) {
				t.Fatal("inverse coordinates disagree with sticker rotation/recoloring")
			}
			if s.axes[0][axis] != (axisCoordinate{root.twist(), root.flip(), root.slice()}) {
				t.Fatal("optimal axis disagrees with sticker rotation/recoloring")
			}
			for m := 0; m < 18; m++ {
				mapped := s.moveMap[axis][m]
				if views[axis*2].faceMap[Face(mapped/3)] != Face(m/3) || mapped%3 != m%3 {
					t.Fatal("optimal move map disagrees with physical face mapping")
				}
				a := s.axes[0][axis]
				got := axisCoordinate{int(tables.Twist[a.co*18+mapped]), int(tables.Flip[a.eo*18+mapped]), int(tables.Slice[a.sl*18+mapped])}
				next := root.mul(cubieMoves[mapped])
				if got != (axisCoordinate{next.twist(), next.flip(), next.slice()}) {
					t.Fatal("optimal axis transition disagrees with cubie move")
				}
				if s.axisMoves[axis]&(1<<m) != 0 {
					a, b, c := s.inverseAxis(state.mul(cubieMoves[m]).inverse(), axis)
					if a != co || b != eo || c != sorted {
						t.Fatal("subgroup move changed the inverse coordinate")
					}
				}
			}
		}
	}
}

// Explicit target: ordinary tests and short --optimal solves stay lightweight.
// This exercises the complete generated/cache-loaded databases and independent
// exact IDA* distance oracle, rather than reducing the random-state benchmark.
func TestOptimalLargeTableOracle(t *testing.T) {
	if os.Getenv("CUBE_OPTIMAL_TABLES") != "1" {
		t.Skip("make test-optimal")
	}
	tables := solverTables()
	started := time.Now()
	db := optimalPatternTables(tables, time.Time{})
	if db == nil {
		t.Fatal("large pruning table generation failed")
	}
	t.Logf("large table generation/load: %v; corner %d bytes; edges %d + %d bytes; transitions %d bytes; cache %d bytes", time.Since(started), len(db.corners), len(db.edges[0]), len(db.edges[1]), len(db.moves)*4, optimalCacheBytes)
	if optimalPhase1Tables(tables, time.Time{}) == nil {
		t.Fatal("sorted phase-one tables failed")
	}
	r := rand.New(rand.NewSource(2026100713))
	for trial := 0; trial < 1000; trial++ {
		state := identityCubie()
		for depth := 0; depth <= 12; depth++ {
			cp, co := permutationRank(state.cp[:]), state.twist()
			if nibbleDistance(db.corners, cp*2187+co) > depth {
				t.Fatal("inadmissible full-corner bound", depth)
			}
			m := r.Intn(18)
			next := state.mul(cubieMoves[m])
			for g := range db.edges {
				x, y := sixEdgeCoordinate(state, g), sixEdgeCoordinate(next, g)
				if int(db.moves[x/64*18+m])^(x&63) != y {
					t.Fatal("six-edge transition disagrees with cubie move")
				}
				a, b := nibbleDistance(db.edges[g], x), nibbleDistance(db.edges[g], y)
				if a > depth || a > b+1 || b > a+1 {
					t.Fatal("six-edge distance is inadmissible/inconsistent", depth, a, b)
				}
			}
			state = next
		}
	}
	for depth := 1; depth <= 10; depth++ {
		for trial := 0; trial < 5; trial++ {
			state := identityCubie()
			for i := 0; i < depth; i++ {
				state = state.mul(cubieMoves[r.Intn(18)])
			}
			want, found := exactCoordinateSearch(state, nil, depth)
			got, ok, timedOut := largeOptimalSearchLimit(state, tables, db, time.Now().Add(5*time.Second))
			if !found || !ok || timedOut || len(got) != len(want) {
				t.Fatalf("depth %d: large %d/%v/%v, finder %d/%v", depth, len(got), ok, timedOut, len(want), found)
			}
			c := cubeFromCoordinates(state)
			c.ApplyMoves(got)
			if !c.IsSolved() {
				t.Fatal("large-table optimal answer violates solver contract")
			}
		}
	}
	for _, text := range []string{"R U F2 L' B D2 R F U2 B' L2 U", "F2 R U' L2 D B R2 U F' D2 L B2", "R U F2 L' B D2 R F U2 B' L2 U R2", "F2 R U' L2 D B R2 U F' D2 L B2 U R"} {
		c := NewCube(3)
		scramble, _ := ParseMoves(text)
		c.ApplyMoves(scramble)
		want, found, timedOut := exactCoordinateSearchLimit(readCubie(c), nil, len(scramble), time.Now().Add(2*time.Minute))
		if !found || timedOut {
			t.Fatal("deep finder oracle did not complete", text)
		}
		started := time.Now()
		result, err := SolveOptimal(c, 20*time.Second)
		if err != nil || TurnCount(result.Solution) != len(want) {
			t.Fatal("deep optimal distance differs from finder", text, result, err)
		}
		c.ApplyMoves(result.Solution)
		if !c.IsSolved() {
			t.Fatal("deep optimal answer violates solver contract")
		}
		t.Logf("deep finder cross-check: proved %d turns in %v", len(want), time.Since(started))
	}
	for _, data := range [][]uint8{db.corners, db.edges[0], db.edges[1]} {
		var hist [16]int
		for _, d := range data {
			hist[d&15]++
			hist[d>>4]++
		}
		if hist[0] != 1 || hist[15] != 0 {
			t.Fatal("incomplete pattern database", hist)
		}
		t.Logf("distance histogram: %v", hist)
	}
}

func TestOptimalInversePruningOracle(t *testing.T) {
	if os.Getenv("CUBE_OPTIMAL_TABLES") != "1" {
		t.Skip("make test-optimal")
	}
	tables := solverTables()
	huge := optimalPhase1Tables(tables, time.Time{})
	if huge == nil {
		t.Fatal("sorted phase-one tables failed")
	}
	r := rand.New(rand.NewSource(2026100801))
	for trial := 0; trial < 1000; trial++ {
		state := identityCubie()
		var scramble [14]int
		for i := range scramble {
			scramble[i] = r.Intn(18)
			state = state.mul(cubieMoves[scramble[i]])
		}
		last := (1 << 18) - 1
		for left := len(scramble); left > 0; left-- {
			s := largeOptimalSearch{t: tables, huge: huge}
			s.initializeAxes(state)
			m := scramble[left-1]/3*3 + 2 - scramble[left-1]%3
			candidates := s.inverseCandidates(left, 0, (1<<18)-1)
			if candidates&(1<<m) == 0 {
				t.Fatal("inverse pruning removed a known solving suffix", trial, left)
			}
			for axis, h := range s.hugeBound[0] {
				if h == left {
					last &^= s.axisMoves[axis]
				}
			}
			final := scramble[0]/3*3 + 2 - scramble[0]%3
			if last&(1<<final) == 0 {
				t.Fatal("forward pruning removed a known last move", trial, left)
			}
			state = state.mul(cubieMoves[m])
		}
		if state != identityCubie() {
			t.Fatal("known suffix did not solve")
		}
	}
}

func TestOptimalUniformBenchmark(t *testing.T) {
	if os.Getenv("CUBE_OPTIMAL_BENCH") != "1" {
		t.Skip("make bench-optimal")
	}
	limit := 3 * time.Minute
	if text := os.Getenv("CUBE_OPTIMAL_LIMIT"); text != "" {
		var err error
		limit, err = time.ParseDuration(text)
		if err != nil {
			t.Fatal(err)
		}
	}
	cases := 20
	if text := os.Getenv("CUBE_OPTIMAL_CASES"); text != "" {
		var err error
		cases, err = strconv.Atoi(text)
		if err != nil || cases < 1 {
			t.Fatal("invalid case count")
		}
	}
	tables := solverTables()
	started := time.Now()
	if optimalPatternTables(tables, time.Time{}) == nil || optimalPhase1Tables(tables, time.Time{}) == nil {
		t.Fatal("large tables failed")
	}
	t.Logf("large table generation/load: %v; %d uniform states; per-state limit %v", time.Since(started), cases, limit)
	r := rand.New(rand.NewSource(2026100714))
	solved, timedOut := 0, 0
	for i := 0; i < cases; i++ {
		state := uniformCubie(r)
		c := cubeFromCoordinates(state)
		started = time.Now()
		stats := &optimalSearchStats{}
		result, err := solveOptimal(c, limit, stats)
		elapsed := time.Since(started)
		if readCubie(c) != state {
			t.Fatal("optimal solver mutated input")
		}
		if err != nil {
			if elapsed < limit {
				t.Fatal("unexpected early failure", err)
			}
			timedOut++
			t.Logf("state %d: censored >%v (%v), %d nodes; iterations %v", i, limit, elapsed, stats.nodes, stats.iterations)
			continue
		}
		c.ApplyMoves(result.Solution)
		if !c.IsSolved() {
			t.Fatal("optimal solver contract")
		}
		solved++
		t.Logf("state %d: proved %d turns in %v, %d nodes; iterations %v", i, TurnCount(result.Solution), elapsed, stats.nodes, stats.iterations)
	}
	t.Logf("uniform distribution: %d/%d proved, %d/%d timed out at %v", solved, cases, timedOut, cases, limit)
}

// Compare completed IDA* iterations on identical states. Node counts isolate
// heuristic strength from elapsed-time changes caused by background scheduling.
func TestOptimalHeuristicBenchmark(t *testing.T) {
	if os.Getenv("CUBE_OPTIMAL_COMPARE") != "1" {
		t.Skip("CUBE_OPTIMAL_COMPARE=1")
	}
	tables := solverTables()
	db := optimalPatternTables(tables, time.Time{})
	phase1 := phase1PatternTables(tables)
	huge := optimalPhase1Tables(tables, time.Time{})
	if db == nil || phase1 == nil || huge == nil {
		t.Fatal("optimal tables unavailable")
	}
	r := rand.New(rand.NewSource(2026100714))
	for trial := 0; trial < 2; trial++ {
		state := uniformCubie(r)
		for _, depth := range []int{14, 15} {
			for _, variant := range []int{0, 1, 2} {
				var pattern *optimalPhase1
				if variant > 0 {
					pattern = huge
				}
				start := time.Now()
				s := largeOptimalSearch{t: tables, db: db, phase1: phase1, huge: pattern, deadline: start.Add(time.Minute)}
				s.initializeAxes(state)
				found := s.searchDepth(permutationRank(state.cp[:]), sixEdgeCoordinate(state, 0), sixEdgeCoordinate(state, 1), depth, variant == 2)
				if found {
					c := cubeFromCoordinates(state)
					for _, m := range s.path[:depth] {
						c.ApplyMove(coordinateMoves[m])
					}
					if !c.IsSolved() {
						t.Fatal("comparison solver contract")
					}
				}
				t.Logf("state %d, depth %d, sorted phase one %v, two workers %v: %d nodes, %v, found %v, censored %v", trial, depth, pattern != nil, variant == 2, s.nodes, time.Since(start), found, s.timedOut)
			}
		}
	}
}
