package cube

import (
	"fmt"
	"runtime"
	"time"
)

// ReductionSolver solves 2x2 corners and 4x4 through 7x7 reduction. It never
// mutates its input and verifies the complete returned sequence independently.
type ReductionSolver struct{}

func (s *ReductionSolver) Name() string { return "Reduction" }

func (s *ReductionSolver) Solve(c *Cube) (*SolverResult, error) {
	return SolveNxN(c, KociembaOptions{TargetLength: 20, TimeLimit: time.Second})
}

// SolveNxN places centers, pairs each wing orbit to a legal reduced edge
// state, and finishes through the public 3x3 solver. Options apply to that
// final search, not the constructive reduction. Each printed turn or rotation
// counts once, including wide turns.
func SolveNxN(c *Cube, options KociembaOptions) (*SolverResult, error) {
	started := time.Now()
	if err := validateNxNShape(c); err != nil {
		return nil, err
	}
	if options.TargetLength < 1 || options.TargetLength > 30 || options.TimeLimit <= 0 {
		return nil, fmt.Errorf("reduced Kociemba needs target length 1-30 and a positive time limit")
	}
	work := c.clone()
	var moves []Move
	var preloadDone <-chan struct{}
	if c.Size%2 == 1 {
		// Odd fixed centers define the frame even after slice/rotation moves.
		frame := NewCube(3)
		for f := range frame.Faces {
			frame.Faces[f][1][1] = work.Faces[f][c.Size/2][c.Size/2]
		}
		_, rotations, err := canonical3x3(frame)
		if err != nil {
			return nil, fmt.Errorf("fixed centers: %w", err)
		}
		moves = append(moves, rotations...)
		if err := nxnApplyMoves(work, rotations); err != nil {
			return nil, err
		}
	}
	reduced := nxnReducedSeed(work)
	// Validate even-cube corners with hypothetical legal edges of matching
	// parity. The actual reduced edge state is chosen later by pairing.
	if c.Size%2 == 0 {
		parity, err := nxnCornerParity(reduced)
		if err != nil {
			return nil, err
		}
		if parity != 0 {
			a, b := edgeFacelets[0], edgeFacelets[1]
			for i := range a {
				reduced.Faces[a[i].Face][a[i].Row][a[i].Col], reduced.Faces[b[i].Face][b[i].Row][b[i].Col] =
					sticker(reduced, b[i]), sticker(reduced, a[i])
			}
		}
	}
	if err := Validate3x3(reduced); err != nil {
		return nil, fmt.Errorf("reduced corners/central edges: %w", err)
	}
	if c.Size > 3 {
		// Native cores can load the final 3x3 tables while reduction runs.
		// Keep browser allocation peaks separate on its single execution thread.
		if runtime.GOARCH != "wasm" && !work.IsSolved() {
			done := make(chan struct{})
			go func() {
				solverTables()
				close(done)
			}()
			defer func() { <-done }()
			preloadDone = done
		}
		t, err := nxnTables(c.Size)
		if err != nil {
			return nil, err
		}
		part, err := nxnCenterBlocks(work, t)
		if err != nil {
			return nil, err
		}
		moves = append(moves, part...)
		part, paired, err := nxnPairWings(work, t)
		if err != nil {
			return nil, err
		}
		moves = append(moves, part...)
		reduced = paired
	}
	if preloadDone != nil {
		<-preloadDone
	}
	if c.Size == 2 && !reduced.IsSolved() {
		nxnTwoByTwoTables()
	}
	finish, err := SolveKociemba(reduced, options)
	if err != nil {
		return nil, fmt.Errorf("reduced 3x3 solve: %w", err)
	}
	moves = nxnOptimizeMoves(append(moves, finish.Solution...), c.Size)
	if !nxnVerifySolution(c, moves) {
		return nil, fmt.Errorf("reduction solution failed uniform, center-matched verification")
	}
	return &SolverResult{Solution: moves, Steps: len(moves), Duration: time.Since(started)}, nil
}

// Verify every sticker against its canonical destination through the complete
// solution permutation. This is equivalent to a uniform, center-matched full
// replay and needs one permutation instead of allocating a cube per turn.
func nxnVerifySolution(c *Cube, moves []Move) bool {
	if err := ValidateMoves(moves, c.Size); err != nil {
		return false
	}
	home := [6]Color{Blue, Green, Orange, Red, Yellow, White}
	for src, dst := range nxnPermutation(c.Size, moves) {
		if nxnColor(c, src) != home[dst/(c.Size*c.Size)] {
			return false
		}
	}
	return true
}

func nxnCornerParity(c *Cube) (int, error) {
	home := NewCube(3)
	p, seen := make([]int, 8), [8]bool{}
	for slot, coords := range cornerFacelets {
		found := false
		for id, target := range cornerFacelets {
			if cornerColorsMatch(sticker(c, coords[0]), sticker(c, coords[1]), sticker(c, coords[2]),
				sticker(home, target[0]), sticker(home, target[1]), sticker(home, target[2])) {
				if seen[id] {
					return 0, fmt.Errorf("duplicate corner in reduced cube")
				}
				p[slot], seen[id], found = id, true, true
				break
			}
		}
		if !found {
			return 0, fmt.Errorf("invalid corner colors in reduced cube")
		}
	}
	return permutationParity(p), nil
}

func validateNxNShape(c *Cube) error {
	if c == nil || c.Size < 2 || c.Size > 7 || c.Size == 3 {
		return fmt.Errorf("reduction supports dimensions 2 and 4-7; use the 3x3 solver for dimension 3")
	}
	var counts [6]int
	for f, face := range c.Faces {
		if len(face) != c.Size {
			return fmt.Errorf("face %d needs %d rows", f, c.Size)
		}
		for _, row := range face {
			if len(row) != c.Size {
				return fmt.Errorf("face %d needs %d stickers per row", f, c.Size)
			}
			for _, color := range row {
				if color < White || color > Green {
					return fmt.Errorf("a concrete cube needs real colors, without wildcards")
				}
				counts[color]++
			}
		}
	}
	for color, count := range counts {
		if count != c.Size*c.Size {
			return fmt.Errorf("color %s has %d stickers; expected %d", Color(color), count, c.Size*c.Size)
		}
	}
	return nil
}

// Corners scale exactly. Odd central edges are sampled; even/2x2 edges
// initially have canonical colors and are adjusted for corner parity above.
func nxnReducedSeed(c *Cube) *Cube {
	r := NewCube(3)
	for _, corner := range cornerFacelets {
		for _, p := range corner {
			r.Faces[p.Face][p.Row][p.Col] = c.Faces[p.Face][p.Row*(c.Size-1)/2][p.Col*(c.Size-1)/2]
		}
	}
	if c.Size%2 == 1 {
		for _, edge := range edgeFacelets {
			for _, p := range edge {
				r.Faces[p.Face][p.Row][p.Col] = c.Faces[p.Face][p.Row*(c.Size-1)/2][p.Col*(c.Size-1)/2]
			}
		}
	}
	return r
}

func nxnColor(c *Cube, index int) Color {
	f, r, col := indexToCoord(index, c.Size)
	return c.Faces[f][r][col]
}

// Reduction applies many short algorithms to an already validated cube. Keep
// its stickers in two flat stack buffers for a whole algorithm rather than
// allocating six faces and their rows on every turn. Final verification uses
// the independently composed full solution permutation above.
func nxnApplyMoves(c *Cube, moves []Move) error {
	if err := ValidateMoves(moves, c.Size); err != nil {
		return err
	}
	if len(moves) == 0 {
		return nil
	}
	var first, second [6 * 7 * 7]Color
	state, after := &first, &second
	n := c.Size
	for face, rows := range c.Faces {
		for row, colors := range rows {
			copy(state[face*n*n+row*n:], colors)
		}
	}
	for _, m := range moves {
		kind, turns := moveToMoveType(m)
		lo, hi := m.Layer, m.Layer+1
		if m.Slice != NoSlice {
			lo, hi = n/2, n/2+1
		} else if m.Rotation != NoRotation {
			lo, hi = 0, 1
		} else if m.Wide {
			lo, hi = 0, m.WideDepth
			if hi == 0 {
				hi = 2
			}
		}
		for layer := lo; layer < hi; layer++ {
			for src, dst := range getPermutation(n, kind, layer, turns) {
				after[dst] = state[src]
			}
			state, after = after, state
		}
	}
	for face, rows := range c.Faces {
		for row, colors := range rows {
			copy(colors, state[face*n*n+row*n:face*n*n+(row+1)*n])
		}
	}
	return nil
}

func nxnReducedColor(c *Cube, index, n int) Color {
	f, r, col := indexToCoord(index, n)
	compress := func(x int) int {
		if x == 0 {
			return 0
		}
		if x == n-1 {
			return 2
		}
		return 1
	}
	return c.Faces[f][compress(r)][compress(col)]
}

func nxnWingPermutation(c, reduced *Cube, o *reductionOrbit) ([]int, error) {
	var targets [36]int
	for i := range targets {
		targets[i] = -1
	}
	for i, pos := range o.positions {
		a, b := nxnReducedColor(reduced, pos, c.Size), nxnReducedColor(reduced, o.partners[i], c.Size)
		key := int(a)*6 + int(b)
		if targets[key] >= 0 {
			return nil, fmt.Errorf("duplicate reduced wing target %s-%s", a, b)
		}
		targets[key] = i
	}
	p, seen := make([]int, 24), [24]bool{}
	for i, pos := range o.positions {
		a, b := nxnColor(c, pos), nxnColor(c, o.partners[i])
		id := targets[int(a)*6+int(b)]
		if id < 0 || seen[id] {
			return nil, fmt.Errorf("invalid or duplicate wing %s-%s in layer %d", a, b, o.layer+1)
		}
		p[i], seen[id] = id, true
	}
	return p, nil
}

func nxnValidateCenterOrbit(c *Cube, o *reductionOrbit) error {
	home := NewCube(c.Size)
	var actual, expected [6]int
	for _, pos := range o.positions {
		actual[nxnColor(c, pos)]++
		expected[nxnColor(home, pos)]++
	}
	if actual != expected {
		return fmt.Errorf("incorrect color counts in a center orbit")
	}
	return nil
}

func nxnCenterMatched(c *Cube) bool {
	home := NewCube(c.Size)
	for f := range c.Faces {
		if c.Faces[f][0][0] != home.Faces[f][0][0] {
			return false
		}
	}
	return true
}
