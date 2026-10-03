package cube

import (
	"fmt"
	"testing"
)

// This oracle tracks a sticker's position AND outward normal in 3D. It does not
// call the engine's rings, permutations, cubie addresses, or coordinate helpers.
// Coordinates are scaled by N-1 so every grid cell has an integer position.
type geometrySticker struct {
	position [3]int
	normal   [3]int
}

func geometryAddress(face Face, row, col, size int) geometrySticker {
	m := size - 1
	a, b := 2*col-m, m-2*row
	switch face {
	case Front:
		return geometrySticker{[3]int{a, b, m}, [3]int{0, 0, 1}}
	case Back:
		return geometrySticker{[3]int{-a, b, -m}, [3]int{0, 0, -1}}
	case Left:
		return geometrySticker{[3]int{-m, b, a}, [3]int{-1, 0, 0}}
	case Right:
		return geometrySticker{[3]int{m, b, -a}, [3]int{1, 0, 0}}
	case Up:
		return geometrySticker{[3]int{a, m, -b}, [3]int{0, 1, 0}}
	case Down:
		return geometrySticker{[3]int{a, -m, b}, [3]int{0, -1, 0}}
	}
	panic("invalid face")
}

// Negative right-handed quarter turns about the positive x/y/z axes are
// clockwise when looking at the R/U/F face from outside the cube.
func geometryTurn(v [3]int, axis int) [3]int {
	switch axis {
	case 0:
		return [3]int{v[0], v[2], -v[1]}
	case 1:
		return [3]int{-v[2], v[1], v[0]}
	case 2:
		return [3]int{v[1], -v[0], v[2]}
	}
	panic("invalid axis")
}

func geometryLabels(size int) *Cube {
	c := NewCube(size)
	for f := range c.Faces {
		for r := 0; r < size; r++ {
			for col := 0; col < size; col++ {
				// Labels are deliberately unique: uniform solved faces would hide
				// a reversed grid or a separated corner's stickers.
				c.Faces[f][r][col] = Color(f*size*size + r*size + col)
			}
		}
	}
	return c
}

func geometryOracle(before *Cube, axis, turns int, selected func([3]int) bool) *Cube {
	size := before.Size
	addresses := make(map[geometrySticker]Coord, 6*size*size)
	for f := range before.Faces {
		for r := 0; r < size; r++ {
			for col := 0; col < size; col++ {
				addresses[geometryAddress(Face(f), r, col, size)] = Coord{Face(f), r, col}
			}
		}
	}
	after := NewCube(size)
	for from, to := range addresses {
		destination := from
		if selected(from.position) {
			for i := 0; i < (turns+4)%4; i++ {
				destination.position = geometryTurn(destination.position, axis)
				destination.normal = geometryTurn(destination.normal, axis)
			}
		}
		p, ok := addresses[destination]
		if !ok {
			panic("rigid transform left the cube surface")
		}
		after.Faces[p.Face][p.Row][p.Col] = before.Faces[to.Face][to.Row][to.Col]
	}
	return after
}

func assertGeometryEqual(t *testing.T, got, want *Cube) {
	t.Helper()
	for f := range got.Faces {
		for r := 0; r < got.Size; r++ {
			for col := 0; col < got.Size; col++ {
				if got.Faces[f][r][col] != want.Faces[f][r][col] {
					t.Fatalf("destination %s[%d][%d]: source label %d, want %d",
						Face(f), r, col, int(got.Faces[f][r][col]), int(want.Faces[f][r][col]))
				}
			}
		}
	}
}

func TestCubeRotationsRigidGeometry(t *testing.T) {
	for size := 2; size <= 6; size++ {
		for axis, token := range []string{"x", "y", "z"} {
			for _, modifier := range []struct {
				suffix string
				turns  int
			}{{"", 1}, {"2", 2}, {"'", 3}} {
				t.Run(fmt.Sprintf("%dx%d/%s%s", size, size, token, modifier.suffix), func(t *testing.T) {
					got := geometryLabels(size)
					want := geometryOracle(got, axis, modifier.turns, func([3]int) bool { return true })
					move, err := ParseMove(token + modifier.suffix)
					if err != nil {
						t.Fatal(err)
					}
					got.ApplyMove(move)
					assertGeometryEqual(t, got, want)
				})
			}
		}
	}
}

func TestFaceAndWideTurnsRigidGeometry(t *testing.T) {
	for size := 2; size <= 6; size++ {
		for _, face := range []struct {
			token string
			axis  int
			sign  int
		}{{"R", 0, 1}, {"L", 0, -1}, {"U", 1, 1}, {"D", 1, -1}, {"F", 2, 1}, {"B", 2, -1}} {
			for width := 1; width <= size; width++ {
				for _, modifier := range []struct {
					suffix string
					turns  int
				}{{"", 1}, {"2", 2}, {"'", 3}} {
					token := face.token
					if width == 2 {
						token += "w"
					} else if width > 2 {
						token = fmt.Sprintf("%d%sw", width, token)
					}
					token += modifier.suffix
					t.Run(fmt.Sprintf("%dx%d/%s", size, size, token), func(t *testing.T) {
						got := geometryLabels(size)
						want := geometryOracle(got, face.axis, face.sign*modifier.turns, func(p [3]int) bool {
							return face.sign*p[face.axis] >= size-1-2*(width-1)
						})
						move, err := ParseMove(token)
						if err != nil {
							t.Fatal(err)
						}
						got.ApplyMove(move)
						assertGeometryEqual(t, got, want)
					})
				}
			}
		}
	}
}

func TestLayerTurnsRigidGeometry(t *testing.T) {
	for size := 2; size <= 6; size++ {
		for _, face := range []struct {
			token string
			axis  int
			sign  int
		}{{"R", 0, 1}, {"L", 0, -1}, {"U", 1, 1}, {"D", 1, -1}, {"F", 2, 1}, {"B", 2, -1}} {
			for layer := 1; layer < size; layer++ {
				for _, modifier := range []struct {
					suffix string
					turns  int
				}{{"", 1}, {"2", 2}, {"'", 3}} {
					token := fmt.Sprintf("%d%s%s", layer+1, face.token, modifier.suffix)
					t.Run(fmt.Sprintf("%dx%d/%s", size, size, token), func(t *testing.T) {
						got := geometryLabels(size)
						want := geometryOracle(got, face.axis, face.sign*modifier.turns, func(p [3]int) bool {
							return face.sign*p[face.axis] == size-1-2*layer
						})
						move, err := ParseMove(token)
						if err != nil {
							t.Fatal(err)
						}
						got.ApplyMove(move)
						assertGeometryEqual(t, got, want)
					})
				}
			}
		}
	}
}

func TestRotationSliceIdentities(t *testing.T) {
	for _, pair := range [][2]string{{"x", "R M' L'"}, {"y", "U E' D'"}, {"z", "F S B'"}} {
		t.Run(pair[0], func(t *testing.T) {
			got, want := geometryLabels(3), geometryLabels(3)
			for i, c := range []*Cube{got, want} {
				moves, err := ParseMoves(pair[i])
				if err != nil {
					t.Fatal(err)
				}
				c.ApplyMoves(moves)
			}
			assertGeometryEqual(t, got, want)
		})
	}
}

func TestMiddleSliceTurnsRigidGeometry(t *testing.T) {
	for _, size := range []int{3, 5} {
		// Singmaster M turns as L, E as D, and S as F.
		for _, slice := range []struct {
			token string
			axis  int
			sign  int
		}{{"M", 0, -1}, {"E", 1, -1}, {"S", 2, 1}} {
			for _, modifier := range []struct {
				suffix string
				turns  int
			}{{"", 1}, {"2", 2}, {"'", 3}} {
				token := slice.token + modifier.suffix
				t.Run(fmt.Sprintf("%dx%d/%s", size, size, token), func(t *testing.T) {
					got := geometryLabels(size)
					want := geometryOracle(got, slice.axis, slice.sign*modifier.turns, func(p [3]int) bool {
						return p[slice.axis] == 0
					})
					move, err := ParseMove(token)
					if err != nil {
						t.Fatal(err)
					}
					got.ApplyMove(move)
					assertGeometryEqual(t, got, want)
				})
			}
		}
	}
}
