package cube

import (
	"fmt"
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
// final search, not the constructive reduction. Move count is slice turns.
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
		if err := work.ApplyMoves(rotations); err != nil {
			return nil, err
		}
	}
	reduced := nxnReducedSeed(work)
	// On even cubes there is no midge. Choose unflipped edges, exchanging
	// two if required to match corner permutation parity (PLL parity).
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
		t, err := nxnTables(c.Size)
		if err != nil {
			return nil, err
		}
		// An odd wing permutation cannot be paired by 3-cycles. One inner
		// quarter turn makes it even (OLL parity); centers are restored next.
		// This also fixes individual inner-orbit parity on odd/larger cubes.
		for _, o := range t.wings {
			p, err := nxnWingPermutation(work, reduced, o)
			if err != nil {
				return nil, err
			}
			if permutationParity(p) != 0 {
				m := Move{Face: Right, Layer: o.layer, Clockwise: true}
				if err := work.ApplyMove(m); err != nil {
					return nil, err
				}
				moves = append(moves, m)
			}
		}
		for _, o := range t.centers {
			p, err := nxnCenterPermutation(work, o)
			if err != nil {
				return nil, err
			}
			part, err := nxnPlaceOrbit(work, t, o, p)
			if err != nil {
				return nil, err
			}
			moves = append(moves, part...)
		}
		for _, o := range t.wings {
			p, err := nxnWingPermutation(work, reduced, o)
			if err != nil {
				return nil, err
			}
			part, err := nxnPlaceOrbit(work, t, o, p)
			if err != nil {
				return nil, err
			}
			moves = append(moves, part...)
		}
	}
	finish, err := SolveKociemba(reduced, options)
	if err != nil {
		return nil, fmt.Errorf("reduced 3x3 solve: %w", err)
	}
	moves = OptimizeMoves(append(moves, finish.Solution...))
	check := c.clone()
	if err := check.ApplyMoves(moves); err != nil {
		return nil, err
	}
	if !check.IsSolved() || !nxnCenterMatched(check) {
		return nil, fmt.Errorf("reduction solution failed uniform, center-matched verification")
	}
	return &SolverResult{Solution: moves, Steps: len(moves), Duration: time.Since(started)}, nil
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

func nxnCenterPermutation(c *Cube, o *reductionOrbit) ([]int, error) {
	home := NewCube(c.Size)
	p, used := make([]int, 24), [24]bool{}
	for i := range p {
		p[i] = -1
		if nxnColor(c, o.positions[i]) == nxnColor(home, o.positions[i]) {
			p[i], used[i] = i, true
		}
	}
	for i, pos := range o.positions {
		if p[i] >= 0 {
			continue
		}
		for j, target := range o.positions {
			if !used[j] && nxnColor(c, pos) == nxnColor(home, target) {
				p[i], used[j] = j, true
				break
			}
		}
		if p[i] < 0 {
			return nil, fmt.Errorf("incorrect color counts in a center orbit")
		}
	}
	if permutationParity(p) != 0 {
		// Same-colored center labels are interchangeable, so choose an even
		// labeling without changing the requested physical colors.
		for i, pos := range o.positions {
			for j := i + 1; j < len(o.positions); j++ {
				if nxnColor(c, pos) == nxnColor(c, o.positions[j]) {
					p[i], p[j] = p[j], p[i]
					return p, nil
				}
			}
		}
	}
	return p, nil
}

func nxnPlaceOrbit(c *Cube, t *reductionTables, o *reductionOrbit, p []int) ([]Move, error) {
	if permutationParity(p) != 0 {
		return nil, fmt.Errorf("odd reduction permutation after parity correction")
	}
	var moves []Move
	for i := 0; i < len(p); i++ {
		for p[i] != i {
			j := i + 1
			for j < len(p) && p[j] != i {
				j++
			}
			k := p[i]
			if k == j {
				k = i + 1
				for k < len(p) && (k == j || p[k] == k) {
					k++
				}
			}
			if j >= len(p) || k >= len(p) || k <= i {
				return nil, fmt.Errorf("reduction permutation cannot be placed by 3-cycles")
			}
			part := o.cycleMoves(t, j, i, k)
			if err := c.ApplyMoves(part); err != nil {
				return nil, err
			}
			moves = append(moves, part...)
			p[i], p[k], p[j] = p[j], p[i], p[k]
		}
	}
	return moves, nil
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
