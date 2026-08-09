package cube

import (
	"fmt"
)

// EdgePiece identifies a physical edge cubie by the two sticker addresses of
// its two colored stickers.
type EdgePiece struct {
	A, B CubieAddress
}

// CornerPiece identifies a physical corner cubie by the three sticker
// addresses of its three colored stickers.
type CornerPiece struct {
	A, B, C CubieAddress
}

// isBoundaryCell reports whether the cell lies on any boundary row or column
// of its face (i.e. it is an edge or corner sticker, not a center sticker).
func isBoundaryCell(face Face, row, col, size int) bool {
	return row == 0 || row == size-1 || col == 0 || col == size-1
}

// isCornerCell reports whether the cell lies on both a boundary row and a
// boundary column of its face.
func isCornerCell(face Face, row, col, size int) bool {
	return (row == 0 || row == size-1) && (col == 0 || col == size-1)
}

// isEdgeCell reports whether the cell lies on exactly one boundary (a boundary
// row or a boundary column, but not both) of its face.
func isEdgeCell(face Face, row, col, size int) bool {
	return isBoundaryCell(face, row, col, size) && !isCornerCell(face, row, col, size)
}

// boundaryPartner returns the sticker address on the adjacent face that shares
// an edge with the given cell. It is only well-defined for edge cells (cells on
// exactly one boundary of their face); corner and center cells return an error.
//
// The mappings mirror the ring definitions in ring_generators.go: on a face of
// size N, row 0 faces the "Up"-side neighbor, row N-1 the "Down"-side neighbor,
// col 0 the "Left"-side neighbor and col N-1 the "Right"-side neighbor.
func boundaryPartner(face Face, row, col, size int) (CubieAddress, error) {
	last := size - 1
	var nf Face
	var nr, nc int

	switch face {
	case Up:
		switch {
		case row == 0:
			nf, nr, nc = Back, 0, last-col
		case row == last:
			nf, nr, nc = Front, 0, col
		case col == 0:
			nf, nr, nc = Left, 0, row
		case col == last:
			nf, nr, nc = Right, 0, last-row
		default:
			return 0, fmt.Errorf("cell (%d,%d) on %s is not an edge cell", row, col, face)
		}
	case Down:
		switch {
		case row == 0:
			nf, nr, nc = Front, last, col
		case row == last:
			nf, nr, nc = Back, last, last-col
		case col == 0:
			nf, nr, nc = Left, last, last-row
		case col == last:
			nf, nr, nc = Right, last, row
		default:
			return 0, fmt.Errorf("cell (%d,%d) on %s is not an edge cell", row, col, face)
		}
	case Front:
		switch {
		case row == 0:
			nf, nr, nc = Up, last, col
		case row == last:
			nf, nr, nc = Down, 0, col
		case col == 0:
			nf, nr, nc = Left, row, last
		case col == last:
			nf, nr, nc = Right, row, 0
		default:
			return 0, fmt.Errorf("cell (%d,%d) on %s is not an edge cell", row, col, face)
		}
	case Back:
		switch {
		case row == 0:
			nf, nr, nc = Up, 0, last-col
		case row == last:
			nf, nr, nc = Down, last, last-col
		case col == 0:
			nf, nr, nc = Right, row, last
		case col == last:
			nf, nr, nc = Left, row, 0
		default:
			return 0, fmt.Errorf("cell (%d,%d) on %s is not an edge cell", row, col, face)
		}
	case Left:
		switch {
		case row == 0:
			nf, nr, nc = Up, col, 0
		case row == last:
			nf, nr, nc = Down, last-col, 0
		case col == 0:
			nf, nr, nc = Back, row, last
		case col == last:
			nf, nr, nc = Front, row, 0
		default:
			return 0, fmt.Errorf("cell (%d,%d) on %s is not an edge cell", row, col, face)
		}
	case Right:
		switch {
		case row == 0:
			nf, nr, nc = Up, last-col, last
		case row == last:
			nf, nr, nc = Down, col, last
		case col == 0:
			nf, nr, nc = Front, row, last
		case col == last:
			nf, nr, nc = Back, row, 0
		default:
			return 0, fmt.Errorf("cell (%d,%d) on %s is not an edge cell", row, col, face)
		}
	default:
		return 0, fmt.Errorf("invalid face %v", face)
	}

	return FacePosToCubie(nf, nr, nc, size), nil
}

// FindEdge locates the edge cubie whose two stickers are exactly colorA and
// colorB, in either order. It only supports 3x3 cubes.
func FindEdge(c *Cube, colorA, colorB Color) (EdgePiece, error) {
	if c.Size != 3 {
		return EdgePiece{}, fmt.Errorf("FindEdge requires a 3x3 cube, got size %d", c.Size)
	}
	if colorA == colorB {
		return EdgePiece{}, fmt.Errorf("FindEdge requires two distinct colors, got %s twice", colorA)
	}

	for addr := 1; addr <= 6*c.Size*c.Size; addr++ {
		a := CubieAddress(addr)
		face, row, col := CubieToFacePos(a, c.Size)
		if !isEdgeCell(face, row, col, c.Size) {
			continue
		}
		b, err := boundaryPartner(face, row, col, c.Size)
		if err != nil {
			continue
		}
		ca := c.GetCubieColor(a)
		cb := c.GetCubieColor(b)
		if (ca == colorA && cb == colorB) || (ca == colorB && cb == colorA) {
			return EdgePiece{A: a, B: b}, nil
		}
	}

	return EdgePiece{}, fmt.Errorf("no edge found with colors %s and %s", colorA, colorB)
}

// stickerCoord returns the cell coordinate on the -1..+1 cube centered at the
// origin, following the same face orientation used by the move engine. The
// values are integers: -1, 0 or +1.
func stickerCoord(face Face, row, col, size int) (x, y, z int) {
	m := size - 1
	switch face {
	case Up:
		return 2*col/m - 1, 1, 2*row/m - 1
	case Down:
		return 2*col/m - 1, -1, 1 - 2*row/m
	case Front:
		return 2*col/m - 1, 1 - 2*row/m, 1
	case Back:
		return 1 - 2*col/m, 1 - 2*row/m, -1
	case Left:
		return -1, 1 - 2*row/m, 2*col/m - 1
	case Right:
		return 1, 1 - 2*row/m, 1 - 2*col/m
	}
	return 0, 0, 0
}

// cornerKey returns a stable identifier for the cube corner the given cell
// occupies. It is meaningful only for corner cells; center and edge cells
// return ok == false. Two cells sharing a corner key belong to the same
// physical cube corner.
func cornerKey(face Face, row, col, size int) (key int, ok bool) {
	if !isCornerCell(face, row, col, size) {
		return 0, false
	}
	x, y, z := stickerCoord(face, row, col, size)
	if x < 0 {
		key |= 1
	}
	if y < 0 {
		key |= 2
	}
	if z < 0 {
		key |= 4
	}
	return key, true
}

// colorTripleSorted returns its three arguments sorted in ascending order.
func colorTripleSorted(a, b, c Color) (Color, Color, Color) {
	if a > b {
		a, b = b, a
	}
	if b > c {
		b, c = c, b
	}
	if a > b {
		a, b = b, a
	}
	return a, b, c
}

// cornerColorsMatch reports whether the three colors match the target triple in
// any order.
func cornerColorsMatch(a1, a2, a3, b1, b2, b3 Color) bool {
	sa1, sa2, sa3 := colorTripleSorted(a1, a2, a3)
	sb1, sb2, sb3 := colorTripleSorted(b1, b2, b3)
	return sa1 == sb1 && sa2 == sb2 && sa3 == sb3
}

// FindCorner locates the corner cubie whose three stickers are exactly c1, c2
// and c3, in any order. It only supports 3x3 cubes.
func FindCorner(c *Cube, c1, c2, c3 Color) (CornerPiece, error) {
	if c.Size != 3 {
		return CornerPiece{}, fmt.Errorf("FindCorner requires a 3x3 cube, got size %d", c.Size)
	}
	if c1 == c2 || c1 == c3 || c2 == c3 {
		return CornerPiece{}, fmt.Errorf("FindCorner requires three distinct colors, got %s, %s, %s", c1, c2, c3)
	}

	for key := 0; key < 8; key++ {
		var addrs [3]CubieAddress
		n := 0
		for addr := 1; addr <= 6*c.Size*c.Size; addr++ {
			a := CubieAddress(addr)
			face, row, col := CubieToFacePos(a, c.Size)
			k, ok := cornerKey(face, row, col, c.Size)
			if !ok || k != key {
				continue
			}
			addrs[n] = a
			n++
		}
		if n != 3 {
			continue
		}
		got1 := c.GetCubieColor(addrs[0])
		got2 := c.GetCubieColor(addrs[1])
		got3 := c.GetCubieColor(addrs[2])
		if cornerColorsMatch(got1, got2, got3, c1, c2, c3) {
			return CornerPiece{A: addrs[0], B: addrs[1], C: addrs[2]}, nil
		}
	}

	return CornerPiece{}, fmt.Errorf("no corner found with colors %s, %s, %s", c1, c2, c3)
}

// faceColor returns the color of a face's center sticker, i.e. the fixed color
// of that face.
func (c *Cube) faceColor(face Face) Color {
	return c.Faces[face][c.Size/2][c.Size/2]
}

// pieceInHomeSlot reports whether each sticker of the piece occupies a cell on
// the face whose center color matches the sticker's own color. For a 3x3 cube
// this is exactly the "home slot with correct orientation" condition.
func pieceInHomeSlot(c *Cube, addrs []CubieAddress) bool {
	for _, addr := range addrs {
		face, _, _ := CubieToFacePos(addr, c.Size)
		if c.GetCubieColor(addr) != c.faceColor(face) {
			return false
		}
	}
	return true
}

// EdgeSolved reports whether the edge cubie with stickers colorA and colorB sits
// in its home slot with the correct orientation, i.e. each sticker's color
// equals the center color of the face it is currently on.
func EdgeSolved(c *Cube, colorA, colorB Color) bool {
	edge, err := FindEdge(c, colorA, colorB)
	if err != nil {
		return false
	}
	return pieceInHomeSlot(c, []CubieAddress{edge.A, edge.B})
}

// CornerSolved reports whether the corner cubie with stickers c1, c2 and c3
// sits in its home slot with the correct orientation, i.e. each sticker's color
// equals the center color of the face it is currently on.
func CornerSolved(c *Cube, c1, c2, c3 Color) bool {
	corner, err := FindCorner(c, c1, c2, c3)
	if err != nil {
		return false
	}
	return pieceInHomeSlot(c, []CubieAddress{corner.A, corner.B, corner.C})
}

// WhiteCrossSolved reports whether all four white edges are solved. Given the
// canonical orientation (white on the Down face), these are the white-blue,
// white-orange, white-red and white-green edges.
func WhiteCrossSolved(c *Cube) bool {
	return EdgeSolved(c, White, Blue) &&
		EdgeSolved(c, White, Orange) &&
		EdgeSolved(c, White, Red) &&
		EdgeSolved(c, White, Green)
}

// FirstLayerSolved reports whether the white cross and all four white corners
// are solved (i.e. the entire first layer is solved).
func FirstLayerSolved(c *Cube) bool {
	if !WhiteCrossSolved(c) {
		return false
	}
	return CornerSolved(c, White, Blue, Red) &&
		CornerSolved(c, White, Blue, Orange) &&
		CornerSolved(c, White, Green, Red) &&
		CornerSolved(c, White, Green, Orange)
}
