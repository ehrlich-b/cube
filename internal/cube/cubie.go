package cube

// CUBIE ADDRESSING SYSTEM - CURRENTLY UNUSED
//
// This file implements a linear addressing system for cube positions that could be useful
// for advanced solving algorithms that need to track individual pieces. The system maps
// 3D cube positions to linear indices for efficient computation.
//
// Status: Implemented but not integrated into current solving system
// Future use: Piece tracking for advanced solvers (Phase 3+ in TODO.md)
//
// CubiePosition represents a 3D position on the cube using linear addressing
// For an NxN cube, positions are numbered 1 to (6 * N * N):
//
// Face layout (reading positions like a book, left-to-right, top-to-bottom):
// For 3x3:     For 4x4:
//   1 2 3        1  2  3  4
//   4 5 6        5  6  7  8
//   7 8 9        9 10 11 12
//               13 14 15 16
//
// 3D cube numbering (example for 3x3):
// Up face (U):    1-9     (Yellow in canonical orientation)
// Left face (L):  10-18   (Orange in canonical orientation)
// Front face (F): 19-27   (Blue in canonical orientation)
// Right face (R): 28-36   (Red in canonical orientation)
// Back face (B):  37-45   (Green in canonical orientation)
// Down face (D):  46-54   (White in canonical orientation)

type CubieAddress int

// Dynamic cubie addressing for any NxN cube
// GetFacePositions returns all cubie addresses for a given face
func GetFacePositions(face Face, size int) []CubieAddress {
	faceSize := size * size
	startPos := int(face)*faceSize + 1

	positions := make([]CubieAddress, faceSize)
	for i := 0; i < faceSize; i++ {
		positions[i] = CubieAddress(startPos + i)
	}
	return positions
}

// Get3x3SpecificPositions returns commonly used position sets for 3x3 cubes only
func Get3x3SpecificPositions() map[string][]CubieAddress {
	if size := 3; size == 3 {
		return map[string][]CubieAddress{
			// Layer aliases for 3x3
			"TL": GetFacePositions(Up, 3),          // Top Layer
			"BL": GetFacePositions(Down, 3),        // Bottom Layer
			"ML": {12, 16, 21, 25, 30, 34, 39, 43}, // Middle layer edges for 3x3

			// Piece type aliases for 3x3
			"TC": {1, 3, 7, 9},                     // Top corners
			"TE": {2, 4, 6, 8},                     // Top edges
			"ME": {12, 16, 21, 25, 30, 34, 39, 43}, // Middle edges
			"BC": {46, 48, 52, 54},                 // Bottom corners
			"BE": {47, 49, 51, 53},                 // Bottom edges

			// Face aliases for 3x3
			"UF": GetFacePositions(Up, 3),
			"LF": GetFacePositions(Left, 3),
			"FF": GetFacePositions(Front, 3),
			"RF": GetFacePositions(Right, 3),
			"BF": GetFacePositions(Back, 3),
			"DF": GetFacePositions(Down, 3),

			// Special combinations for 3x3
			"WC": {2, 4, 6, 8, 50},          // White cross (edges + center)
			"WF": GetFacePositions(Up, 3),   // White face
			"YC": {47, 49, 51, 53, 5},       // Yellow cross (edges + center)
			"YF": GetFacePositions(Down, 3), // Yellow face
		}
	}
	return make(map[string][]CubieAddress)
}

// CubieToFacePos converts a cubie address to face and position within that face
func CubieToFacePos(address CubieAddress, size int) (Face, int, int) {
	pos := int(address) - 1 // Convert to 0-based indexing
	faceSize := size * size

	faceIndex := pos / faceSize
	posInFace := pos % faceSize

	face := Face(faceIndex)
	row := posInFace / size
	col := posInFace % size

	return face, row, col
}

// FacePosToCubie converts face and position to cubie address
func FacePosToCubie(face Face, row, col, size int) CubieAddress {
	faceSize := size * size
	posInFace := row*size + col
	absolutePos := int(face)*faceSize + posInFace + 1

	return CubieAddress(absolutePos)
}

// GetCubieColor returns the color at a specific cubie address
func (c *Cube) GetCubieColor(address CubieAddress) Color {
	face, row, col := CubieToFacePos(address, c.Size)
	return c.Faces[face][row][col]
}

// SetCubieColor sets the color at a specific cubie address
func (c *Cube) SetCubieColor(address CubieAddress, color Color) {
	face, row, col := CubieToFacePos(address, c.Size)
	c.Faces[face][row][col] = color
}

// ParseCubieSpec parses cubie specification strings like "1,2,3" or "1-9" or "TL,ML,BC"
// For 3x3 cubes only initially - can be extended later
func ParseCubieSpec(spec string, size int) ([]CubieAddress, error) {
	// TODO: Implement parsing of cubie specifications
	// Should handle:
	// - Individual addresses: "1,2,3"
	// - Ranges: "1-9", "46-54"
	// - Aliases: "TL" (top layer), "WC" (white cross), etc.

	var result []CubieAddress

	// For now, return empty slice - will implement parsing logic
	return result, nil
}
