package cube

import (
	"math/rand"
	"os"
	"os/exec"
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
	start := time.Now()
	solverTables()
	phase1PatternTables(solverTables())
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
	durations := make([]time.Duration, 1000)
	hist := make(map[int]int)
	var total time.Duration
	for i := range durations {
		state := uniformCubie(r)
		c := cubeFromCoordinates(state)
		start = time.Now()
		result, err := SolveKociemba(c, KociembaOptions{TargetLength: 20, TimeLimit: time.Second})
		durations[i] = time.Since(start)
		total += durations[i]
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
	sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })
	t.Logf("1000 uniform states; warm mean %v p99 %v max %v; lengths %v", total/1000, durations[989], durations[999], hist)
	for length, count := range hist {
		if length > 20 {
			t.Errorf("length target missed: %d states at %d turns", count, length)
		}
	}
	if total/1000 >= 50*time.Millisecond || durations[989] >= 250*time.Millisecond {
		t.Error("warm latency target missed: mean <50ms and p99 <250ms required")
	}
}

func TestTwoPhaseColdBenchmark(t *testing.T) {
	if os.Getenv("CUBE_COLD_BENCH") != "1" {
		t.Skip("benchmark subprocess")
	}
	c := cubeFromCoordinates(uniformCubie(rand.New(rand.NewSource(2026100709))))
	result, err := SolveKociemba(c, KociembaOptions{TargetLength: 20, TimeLimit: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	c.ApplyMoves(result.Solution)
	if !c.IsSolved() {
		t.Fatal("cold cached process solver contract")
	}
}
