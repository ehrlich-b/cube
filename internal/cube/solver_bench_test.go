package cube

import (
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"testing"
	"time"
)

// Shuffle independent physical coordinates, then enforce the three conservation
// laws. Unlike a random walk this samples the legal cube group uniformly.
func uniformCubie(r *rand.Rand) cubie {
	s := identityCubie()
	r.Shuffle(8, func(i, j int) { s.cp[i], s.cp[j] = s.cp[j], s.cp[i] })
	r.Shuffle(12, func(i, j int) { s.ep[i], s.ep[j] = s.ep[j], s.ep[i] })
	parity := 0
	for i, p := range s.cp {
		for _, q := range s.cp[i+1:] {
			if p > q {
				parity ^= 1
			}
		}
	}
	for i, p := range s.ep {
		for _, q := range s.ep[i+1:] {
			if p > q {
				parity ^= 1
			}
		}
	}
	if parity != 0 {
		s.ep[0], s.ep[1] = s.ep[1], s.ep[0]
	}
	sum := 0
	for i := 0; i < 7; i++ {
		s.co[i] = uint8(r.Intn(3))
		sum += int(s.co[i])
	}
	s.co[7] = uint8((3 - sum%3) % 3)
	for i := 0; i < 11; i++ {
		s.eo[i] = uint8(r.Intn(2))
		s.eo[11] ^= s.eo[i]
	}
	return s
}

func TestTwoPhaseBenchmark1000(t *testing.T) {
	if os.Getenv("CUBE_BENCH") != "1" {
		t.Skip("make bench-kociemba")
	}
	runTwoPhaseBenchmark(t, 1000)
}

func TestTwoPhaseDeterminism10000(t *testing.T) {
	if os.Getenv("CUBE_DETERMINISM") != "1" {
		t.Skip("CUBE_DETERMINISM=1 go test ./internal/cube -run '^TestTwoPhaseDeterminism10000$'")
	}
	runTwoPhaseBenchmark(t, 10000)
}

func runTwoPhaseBenchmark(t *testing.T, count int) {
	t.Helper()
	start := time.Now()
	solverTables()

	t.Logf("cold table setup: %v", time.Since(start))
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(executable, "-test.run=^TestTwoPhaseColdBenchmark$", "-test.v")
	cmd.Env = append(os.Environ(), "CUBE_COLD_BENCH=1")
	start = time.Now()
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("cold cached process: %v %s", err, output)
	}
	t.Logf("cold cached process including solve: %v", time.Since(start))
	r := rand.New(rand.NewSource(2026100709))
	durations := make([]time.Duration, count)
	hist := make(map[int]int)
	var total time.Duration
	var maxNodes uint64
	worstWork, worstTime := 0, 0
	for i := range durations {
		state := uniformCubie(r)
		c := cubeFromCoordinates(state)
		start = time.Now()
		search := &kociembaSearch{softLimit: true}
		result, err := solveKociemba(c, KociembaOptions{TargetLength: 20, TimeLimit: time.Second}, search)
		durations[i] = time.Since(start)
		total += durations[i]
		if durations[i] > durations[worstTime] {
			worstTime = i
		}
		if search.nodes > maxNodes {
			maxNodes, worstWork = search.nodes, i
		}

		if err != nil {
			t.Fatalf("state %d: %v", i, err)
		}
		if readCubie(c) != state {
			t.Fatal("input mutated")
		}
		c.ApplyMoves(result.Solution)
		if !c.IsSolved() {
			t.Fatalf("state %d violates solver contract", i)
		}
		hist[TurnCount(result.Solution)]++
	}
	maxTime := durations[worstTime]
	sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })
	p99 := durations[(count*99+99)/100-1]
	t.Logf("%d uniform states; warm mean %v p99 %v max %v (state %d); max DFS operations %d (state %d); lengths %v", count, total/time.Duration(count), p99, maxTime, worstTime, maxNodes, worstWork, hist)
	for length, count := range hist {
		if length > 20 {
			t.Errorf("length target missed: %d states at %d turns", count, length)
		}
	}
	if total/time.Duration(count) >= 50*time.Millisecond || p99 >= 250*time.Millisecond {
		t.Error("warm latency target missed: mean <50ms and p99 <250ms required")
	}
}

func TestTwoPhaseColdBenchmark(t *testing.T) {
	if os.Getenv("CUBE_COLD_BENCH") != "1" {
		t.Skip("benchmark subprocess")
	}
	c := cubeFromCoordinates(uniformCubie(rand.New(rand.NewSource(2026100709))))
	result, err := (&KociembaSolver{}).Solve(c)
	if err != nil {
		t.Fatal(err)
	}
	c.ApplyMoves(result.Solution)
	if !c.IsSolved() {
		t.Fatal("cold cached process solver contract")
	}
}

// Export the exact seeded benchmark populations in the standard reference
// solver facelet order. This is a gated data-generation target, not a solver.
func TestReferenceFixtureExport(t *testing.T) {
	if os.Getenv("CUBE_REFERENCE_FIXTURES") != "1" {
		t.Skip("make export-reference-fixtures")
	}
	home := NewCube(3)
	order := []Face{Up, Right, Front, Down, Left, Back}
	var letters [6]byte
	for i, f := range order {
		letters[home.Faces[f][1][1]] = "URFDLB"[i]
	}
	for _, spec := range []struct {
		seed  int64
		count int
		name  string
	}{
		{2026100709, 1000, "reference-uniform-1000.txt"},
		{2026100714, 10, "reference-optimal-10.txt"},
	} {
		r := rand.New(rand.NewSource(spec.seed))
		data := make([]byte, 0, spec.count*55)
		for i := 0; i < spec.count; i++ {
			c := cubeFromCoordinates(uniformCubie(r))
			if err := Validate3x3(c); err != nil {
				t.Fatal(err)
			}
			for _, f := range order {
				for _, row := range c.Faces[f] {
					for _, color := range row {
						data = append(data, letters[color])
					}
				}
			}
			data = append(data, '\n')
		}
		path := filepath.Join("..", "..", ".scratch", spec.name)
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, data, 0600); err != nil {
			t.Fatal(err)
		}
		t.Logf("exported %d uniform URFDLB fixtures, seed %d, to %s", spec.count, spec.seed, path)
	}
}
