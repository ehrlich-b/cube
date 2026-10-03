package cube

import "fmt"

// These ordered facelets define cubie orientation. Corner triples wind in the
// same direction, starting with U/D; edges start with U/D or, in the belt, F/B.
// Keeping the order is essential: an unordered color set misses mirrored corners.
var edgeFacelets = [12][2]Coord{
	{{Up, 1, 2}, {Right, 0, 1}}, {{Up, 2, 1}, {Front, 0, 1}},
	{{Up, 1, 0}, {Left, 0, 1}}, {{Up, 0, 1}, {Back, 0, 1}},
	{{Down, 1, 2}, {Right, 2, 1}}, {{Down, 0, 1}, {Front, 2, 1}},
	{{Down, 1, 0}, {Left, 2, 1}}, {{Down, 2, 1}, {Back, 2, 1}},
	{{Front, 1, 2}, {Right, 1, 0}}, {{Front, 1, 0}, {Left, 1, 2}},
	{{Back, 1, 2}, {Left, 1, 0}}, {{Back, 1, 0}, {Right, 1, 2}},
}

var cornerFacelets = [8][3]Coord{
	{{Up, 2, 2}, {Right, 0, 0}, {Front, 0, 2}},
	{{Up, 2, 0}, {Front, 0, 0}, {Left, 0, 2}},
	{{Up, 0, 0}, {Left, 0, 0}, {Back, 0, 2}},
	{{Up, 0, 2}, {Back, 0, 0}, {Right, 0, 2}},
	{{Down, 0, 2}, {Front, 2, 2}, {Right, 2, 0}},
	{{Down, 0, 0}, {Left, 2, 2}, {Front, 2, 0}},
	{{Down, 2, 0}, {Back, 2, 2}, {Left, 2, 0}},
	{{Down, 2, 2}, {Right, 2, 2}, {Back, 2, 0}},
}

func sticker(c *Cube, p Coord) Color { return c.Faces[p.Face][p.Row][p.Col] }

// Validate3x3 rejects incomplete or physically unreachable states in the
// project's color scheme, including single flips/twists and permutation parity.
// Whole-cube rotations, slice turns and wide turns are accepted. Input is read only.
func Validate3x3(c *Cube) error {
	if c == nil || c.Size != 3 {
		return fmt.Errorf("a complete 3x3 cube is required")
	}
	var counts [6]int
	for f := range c.Faces {
		if len(c.Faces[f]) != 3 {
			return fmt.Errorf("face %d must have three rows", f)
		}
		for _, row := range c.Faces[f] {
			if len(row) != 3 {
				return fmt.Errorf("face %d must have three stickers per row", f)
			}
			for _, color := range row {
				if color < White || color > Green {
					return fmt.Errorf("a concrete state needs real colors on all 54 stickers (no wildcards)")
				}
				counts[color]++
			}
		}
	}
	for color, count := range counts {
		if count != 9 {
			return fmt.Errorf("color %s has %d stickers; expected 9", Color(color), count)
		}
	}
	work, _, err := canonical3x3(c)
	if err != nil {
		return err
	}
	home := NewCube(3)
	ep := make([]int, 12)
	seenEdges := [12]bool{}
	flips := 0
	for slot, coords := range edgeFacelets {
		a, b := sticker(work, coords[0]), sticker(work, coords[1])
		found := false
		for id, target := range edgeFacelets {
			x, y := sticker(home, target[0]), sticker(home, target[1])
			if (a == x && b == y) || (a == y && b == x) {
				if seenEdges[id] {
					return fmt.Errorf("duplicate edge %s-%s", x, y)
				}
				seenEdges[id], ep[slot], found = true, id, true
				if a != x {
					flips++
				}
				break
			}
		}
		if !found {
			return fmt.Errorf("invalid edge colors %s-%s", a, b)
		}
	}
	cp := make([]int, 8)
	seenCorners := [8]bool{}
	twists := 0
	for slot, coords := range cornerFacelets {
		got := [3]Color{sticker(work, coords[0]), sticker(work, coords[1]), sticker(work, coords[2])}
		found := false
		for id, target := range cornerFacelets {
			want := [3]Color{sticker(home, target[0]), sticker(home, target[1]), sticker(home, target[2])}
			for twist := 0; twist < 3; twist++ {
				if got[twist] == want[0] && got[(twist+1)%3] == want[1] && got[(twist+2)%3] == want[2] {
					if seenCorners[id] {
						return fmt.Errorf("duplicate corner %s-%s-%s", want[0], want[1], want[2])
					}
					seenCorners[id], cp[slot], found = true, id, true
					twists += twist
				}
			}
		}
		if !found {
			return fmt.Errorf("invalid or mirrored corner %s-%s-%s", got[0], got[1], got[2])
		}
	}
	if flips%2 != 0 {
		return fmt.Errorf("unreachable state: an odd number of flipped edges; check sticker entry")
	}
	if twists%3 != 0 {
		return fmt.Errorf("unreachable state: corner twists do not balance; check sticker entry")
	}
	if permutationParity(ep) != permutationParity(cp) {
		return fmt.Errorf("unreachable state: edge and corner permutation parity differ; check sticker entry")
	}
	return nil
}

func permutationParity(p []int) int {
	parity := 0
	for i := range p {
		for j := i + 1; j < len(p); j++ {
			if p[i] > p[j] {
				parity ^= 1
			}
		}
	}
	return parity
}

func centerKey(c *Cube) [6]Color {
	var key [6]Color
	for f := range key {
		key[f] = c.Faces[f][1][1]
	}
	return key
}

// canonical3x3 explores just the 24 rigid orientations, never face turns.
// The returned rotations are part of the lesson, so the physical cube and model
// remain in the same frame even after a rotated scramble or recovery input.
func canonical3x3(c *Cube) (*Cube, []Move, error) {
	type orientation struct {
		cube *Cube
		path []Move
	}
	want := centerKey(NewCube(3))
	queue := []orientation{{c.clone(), nil}}
	seen := map[[6]Color]bool{centerKey(c): true}
	for head := 0; head < len(queue); head++ {
		current := queue[head]
		if centerKey(current.cube) == want {
			return current.cube, current.path, nil
		}
		for _, axis := range []RotationType{X_Rotation, Y_Rotation, Z_Rotation} {
			move := Move{Rotation: axis, Clockwise: true}
			next := current.cube.clone()
			next.ApplyMove(move)
			key := centerKey(next)
			if !seen[key] {
				seen[key] = true
				path := append(append([]Move(nil), current.path...), move)
				queue = append(queue, orientation{next, path})
			}
		}
	}
	return nil, nil, fmt.Errorf("centers do not match the Cube color scheme under any rotation")
}
