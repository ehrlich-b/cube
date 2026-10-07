package cube

import (
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func init() {
	// Keep generated caches in this standalone clone, never in /tmp or another
	// checkout. The normal executable uses the user's cache directory.
	if os.Getenv("CUBE_CACHE_DIR") == "" {
		root, _ := filepath.Abs("../../.scratch/cube-cache")
		_ = os.Setenv("CUBE_CACHE_DIR", root)
	}
}

func TestCoordinateDifferential(t *testing.T) {
	r := rand.New(rand.NewSource(20261007))
	for n := 0; n < 200; n++ {
		c, s := NewCube(3), identityCubie()
		for step := 0; step < 30; step++ {
			m := r.Intn(18)
			c.ApplyMove(coordinateMoves[m])
			s = s.mul(cubieMoves[m])
			if got := readCubie(c); got != s {
				t.Fatalf("case %d step %d: coordinate/sticker disagreement", n, step)
			}
			if s.mul(s.inverse()) != identityCubie() {
				t.Fatal("cubie inverse failed")
			}
		}
	}
	for x := 0; x < 40320; x++ {
		var p [8]uint8
		setPermutation(p[:], x)
		if permutationRank(p[:]) != x {
			t.Fatal("permutation rank round trip", x)
		}
	}
	for x := 0; x < 2187; x++ {
		if twistCubie(x).twist() != x {
			t.Fatal("twist round trip", x)
		}
	}
	for x := 0; x < 2048; x++ {
		if flipCubie(x).flip() != x {
			t.Fatal("flip round trip", x)
		}
	}
	for x := 0; x < 495; x++ {
		if sliceCubie(x).slice() != x {
			t.Fatal("slice round trip", x)
		}
	}
}

func TestKociembaOracle200(t *testing.T) {
	load := time.Now()
	solverTables()
	t.Logf("table initialization: %v", time.Since(load))
	r := rand.New(rand.NewSource(2026100701))
	total, longest := 0, 0
	var sumTime, maxTime time.Duration
	for n := 0; n < 200; n++ {
		c := NewCube(3)
		var scramble []Move
		prev := -1
		for i := 0; i < 30; i++ {
			m := r.Intn(18)
			for skipCoordinateFace(m, prev) {
				m = r.Intn(18)
			}
			prev = m
			scramble = append(scramble, coordinateMoves[m])
		}
		c.ApplyMoves(scramble)
		before := readCubie(c)
		started := time.Now()
		result, err := (&KociembaSolver{}).Solve(c)
		elapsed := time.Since(started)
		if err != nil {
			t.Fatalf("case %d (%s): %v", n, FormatMoves(scramble), err)
		}
		if readCubie(c) != before {
			t.Fatal("solver mutated input")
		}
		if len(result.Solution) == 0 || len(result.Solution) > 22 {
			t.Fatalf("case %d: invalid length %d", n, len(result.Solution))
		}
		c.ApplyMoves(result.Solution)
		if !c.IsSolved() {
			t.Fatalf("case %d: returned moves do not solve", n)
		}
		total += len(result.Solution)
		longest = max(longest, len(result.Solution))
		sumTime += elapsed
		maxTime = max(maxTime, elapsed)
	}
	t.Logf("200 scrambles: mean %.3f max %d face turns; mean %v max %v", float64(total)/200, longest, sumTime/200, maxTime)
}

func cubeFromCoordinates(s cubie) *Cube {
	c, home := NewCube(3), NewCube(3)
	for slot, coords := range cornerFacelets {
		for i, p := range coords {
			c.Faces[p.Face][p.Row][p.Col] = sticker(home, cornerFacelets[s.cp[slot]][(i-int(s.co[slot])+3)%3])
		}
	}
	for slot, coords := range edgeFacelets {
		for i, p := range coords {
			c.Faces[p.Face][p.Row][p.Col] = sticker(home, edgeFacelets[s.ep[slot]][i^int(s.eo[slot])])
		}
	}
	return c
}

func TestKociembaUniformStates200(t *testing.T) {
	r := rand.New(rand.NewSource(2026100705))
	total, longest := 0, 0
	var elapsed, maxTime time.Duration
	for n := 0; n < 200; n++ {
		s := identityCubie()
		r.Shuffle(8, func(i, j int) { s.cp[i], s.cp[j] = s.cp[j], s.cp[i] })
		r.Shuffle(12, func(i, j int) { s.ep[i], s.ep[j] = s.ep[j], s.ep[i] })
		cp, ep := make([]int, 8), make([]int, 12)
		for i, p := range s.cp {
			cp[i] = int(p)
		}
		for i, p := range s.ep {
			ep[i] = int(p)
		}
		if permutationParity(cp) != permutationParity(ep) {
			s.ep[0], s.ep[1] = s.ep[1], s.ep[0]
		}
		sum := 0
		for i := 0; i < 7; i++ {
			s.co[i] = uint8(r.Intn(3))
			sum += int(s.co[i])
		}
		s.co[7] = uint8((3 - sum%3) % 3)
		flip := uint8(0)
		for i := 0; i < 11; i++ {
			s.eo[i] = uint8(r.Intn(2))
			flip ^= s.eo[i]
		}
		s.eo[11] = flip
		c := cubeFromCoordinates(s)
		if err := Validate3x3(c); err != nil || readCubie(c) != s {
			t.Fatal("uniform-state fixture is invalid", err)
		}
		started := time.Now()
		result, err := (&KociembaSolver{}).Solve(c)
		duration := time.Since(started)
		if err != nil {
			t.Fatalf("uniform state %d: %v", n, err)
		}
		if len(result.Solution) > 22 {
			t.Fatal("uniform state exceeds 22 turns")
		}
		c.ApplyMoves(result.Solution)
		if !c.IsSolved() {
			t.Fatalf("uniform state %d did not solve", n)
		}
		total += len(result.Solution)
		longest = max(longest, len(result.Solution))
		elapsed += duration
		maxTime = max(maxTime, duration)
	}
	t.Logf("200 uniform states: mean %.3f max %d face turns; mean %v max %v", float64(total)/200, longest, elapsed/200, maxTime)
}

func TestKociembaFramesAndValidation(t *testing.T) {
	for _, text := range []string{"x y R U F2", "M E S R U", "Rw Fw Uw2 L'"} {
		c := NewCube(3)
		moves, _ := ParseMoves(text)
		c.ApplyMoves(moves)
		result, err := (&KociembaSolver{}).Solve(c)
		if err != nil {
			t.Fatal(text, err)
		}
		c.ApplyMoves(result.Solution)
		if !c.IsSolved() {
			t.Fatal(text, "solution failed")
		}
	}
	for _, c := range []*Cube{nil, NewCube(2), NewCube(4)} {
		if _, err := (&KociembaSolver{}).Solve(c); err == nil {
			t.Fatal("unsupported cube accepted")
		}
	}
	c := NewCube(3)
	a, b := edgeFacelets[0][0], edgeFacelets[0][1]
	c.Faces[a.Face][a.Row][a.Col], c.Faces[b.Face][b.Row][b.Col] = sticker(c, b), sticker(c, a)
	if _, err := (&KociembaSolver{}).Solve(c); err == nil {
		t.Fatal("unreachable flipped edge accepted")
	}
}

func TestKociembaBeatsBeginnerLength(t *testing.T) {
	r := rand.New(rand.NewSource(2026100707))
	fastTotal, beginnerTotal := 0, 0
	for n := 0; n < 30; n++ {
		c := NewCube(3)
		for i := 0; i < 30; i++ {
			c.ApplyMove(coordinateMoves[r.Intn(18)])
		}
		fast, err := (&KociembaSolver{}).Solve(c)
		if err != nil {
			t.Fatal(err)
		}
		beginner, err := (&BeginnerSolver{}).Solve(c)
		if err != nil {
			t.Fatal(err)
		}
		fastTotal += len(fast.Solution)
		beginnerTotal += len(beginner.Solution)
	}
	if fastTotal >= beginnerTotal {
		t.Fatal("Kociemba does not beat beginner on paired length sample")
	}
	t.Logf("30 paired scrambles: Kociemba mean %.2f vs beginner %.2f", float64(fastTotal)/30, float64(beginnerTotal)/30)
}

func TestKociembaSpecialStates(t *testing.T) {
	for _, kind := range []string{"superflip", "corner twists", "checkerboard"} {
		s := identityCubie()
		if kind == "superflip" {
			for i := range s.eo {
				s.eo[i] = 1
			}
		}
		if kind == "corner twists" {
			for i := range s.co {
				s.co[i] = 1
			}
			s.co[7] = 2
		}
		c := cubeFromCoordinates(s)
		if kind == "checkerboard" {
			moves, _ := ParseMoves("R2 L2 U2 D2 F2 B2")
			c.ApplyMoves(moves)
		}
		started := time.Now()
		result, err := (&KociembaSolver{}).Solve(c)
		if err != nil {
			t.Fatal(kind, err)
		}
		c.ApplyMoves(result.Solution)
		if !c.IsSolved() {
			t.Fatal(kind, "did not solve")
		}
		t.Logf("%s: %d turns, %v", kind, len(result.Solution), time.Since(started))
	}
}

func TestCoordinateCacheIntegrity(t *testing.T) {
	tables := solverTables()
	base := os.Getenv("CUBE_CACHE_DIR")
	if err := os.MkdirAll(base, 0700); err != nil {
		t.Fatal(err)
	}
	dir, err := os.MkdirTemp(base, "cache-test-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	t.Setenv("CUBE_CACHE_DIR", dir)
	saveCoordinateTables(tables)
	if cached := loadCoordinateTables(); !reflect.DeepEqual(cached, tables) {
		t.Fatal("cached tables differ from generated tables")
	}
	data, err := os.ReadFile(coordinateCachePath())
	if err != nil {
		t.Fatal(err)
	}
	data[len(data)/2] ^= 1
	if err := os.WriteFile(coordinateCachePath(), data, 0600); err != nil {
		t.Fatal(err)
	}
	if loadCoordinateTables() != nil {
		t.Fatal("corrupted cache accepted")
	}
	if err := os.WriteFile(coordinateCachePath(), data[:10], 0600); err != nil {
		t.Fatal(err)
	}
	if loadCoordinateTables() != nil {
		t.Fatal("truncated cache accepted")
	}
	t.Setenv("CUBE_CACHE_DIR", filepath.Join(coordinateCachePath(), "unwritable"))
	saveCoordinateTables(tables) // A non-directory cache path must not panic.
	if loadCoordinateTables() != nil {
		t.Fatal("invalid cache path accepted")
	}
}
