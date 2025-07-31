package cube

import (
	"fmt"
	"strconv"
	"strings"
)

// SliceType represents middle slice moves
type SliceType int

const (
	NoSlice SliceType = iota
	M_Slice           // Middle slice between L/R
	E_Slice           // Equatorial slice between U/D
	S_Slice           // Standing slice between F/B
)

// String returns the notation for slice moves
func (s SliceType) String() string {
	switch s {
	case M_Slice:
		return "M"
	case E_Slice:
		return "E"
	case S_Slice:
		return "S"
	default:
		return ""
	}
}

// RotationType represents whole cube rotations
type RotationType int

const (
	NoRotation RotationType = iota
	X_Rotation              // Around R axis
	Y_Rotation              // Around U axis
	Z_Rotation              // Around F axis
)

// String returns the notation for cube rotations
func (r RotationType) String() string {
	switch r {
	case X_Rotation:
		return "x"
	case Y_Rotation:
		return "y"
	case Z_Rotation:
		return "z"
	default:
		return ""
	}
}

// Move represents a cube move with advanced notation support
type Move struct {
	Face      Face // F, B, L, R, U, D (or invalid for rotations/slices)
	Clockwise bool
	Double    bool
	Wide      bool         // For wide moves (Rw, Fw, etc.)
	WideDepth int          // How many layers for wide moves (default 2 for "Rw")
	Layer     int          // For layer-specific moves (2R, 3L, etc.) - 0 means outermost
	Slice     SliceType    // For middle slice moves (M, E, S)
	Rotation  RotationType // For cube rotations (x, y, z)
}

// String returns the standard notation for the move
func (m Move) String() string {
	var notation string

	// Handle slice moves
	if m.Slice != NoSlice {
		notation = m.Slice.String()
	} else if m.Rotation != NoRotation {
		// Handle cube rotations
		notation = m.Rotation.String()
	} else {
		// Handle face moves with layer and wide notation
		if m.Layer > 0 {
			notation = strconv.Itoa(m.Layer + 1) // Convert 0-based to 1-based
		}
		notation += m.Face.String()
		if m.Wide {
			notation += "w"
		}
	}

	// Add modifiers
	if m.Double {
		notation += "2"
	} else if !m.Clockwise {
		notation += "'"
	}

	return notation
}

// ParseMove parses a move from advanced notation
// Supports: R, U', F2, 2R, Rw, 2Fw, M, E', S2, x, y', z2
func ParseMove(notation string) (Move, error) {
	notation = strings.TrimSpace(notation)
	if len(notation) == 0 {
		return Move{}, fmt.Errorf("empty move notation")
	}

	move := Move{Clockwise: true, WideDepth: 2} // Default wide depth is 2 layers
	i := 0

	// Check for slice moves first (M, E, S)
	if len(notation) >= 1 {
		switch strings.ToUpper(notation)[0] {
		case 'M':
			move.Slice = M_Slice
			i = 1
		case 'E':
			move.Slice = E_Slice
			i = 1
		case 'S':
			move.Slice = S_Slice
			i = 1
		}
	}

	// Check for cube rotations (x, y, z) - case sensitive
	if move.Slice == NoSlice && len(notation) >= 1 {
		switch notation[0] {
		case 'x':
			move.Rotation = X_Rotation
			i = 1
		case 'y':
			move.Rotation = Y_Rotation
			i = 1
		case 'z':
			move.Rotation = Z_Rotation
			i = 1
		}
	}

	// If not slice or rotation, parse face moves
	if move.Slice == NoSlice && move.Rotation == NoRotation {
		// Check for layer number at start (2R, 3L, etc.)
		if i < len(notation) && notation[i] >= '2' && notation[i] <= '9' {
			layer, err := strconv.Atoi(string(notation[i]))
			if err != nil {
				return Move{}, fmt.Errorf("invalid layer number: %c", notation[i])
			}
			move.Layer = layer - 1 // Convert to 0-based
			i++
		}

		// Parse face
		if i >= len(notation) {
			return Move{}, fmt.Errorf("missing face in notation")
		}

		var face Face
		switch strings.ToUpper(notation)[i] {
		case 'F':
			face = Front
		case 'B':
			face = Back
		case 'L':
			face = Left
		case 'R':
			face = Right
		case 'U':
			face = Up
		case 'D':
			face = Down
		default:
			return Move{}, fmt.Errorf("invalid face: %c", notation[i])
		}
		move.Face = face
		i++

		// Check for wide move marker (w)
		if i < len(notation) && strings.ToLower(notation)[i] == 'w' {
			move.Wide = true
			i++
		}
	}

	// Parse modifiers (', 2)
	for i < len(notation) {
		switch notation[i] {
		case '\'':
			move.Clockwise = false
		case '2':
			move.Double = true
		default:
			return Move{}, fmt.Errorf("invalid modifier: %c", notation[i])
		}
		i++
	}

	return move, nil
}

// ParseScramble parses a scramble string into a slice of moves
func ParseScramble(scramble string) ([]Move, error) {
	if scramble == "" {
		return []Move{}, nil
	}

	parts := strings.Fields(scramble)
	moves := make([]Move, 0, len(parts))

	for _, part := range parts {
		move, err := ParseMove(part)
		if err != nil {
			return nil, fmt.Errorf("error parsing move '%s': %v", part, err)
		}
		moves = append(moves, move)
	}

	return moves, nil
}

// ApplyMove applies a single move to the cube with advanced notation support
func (c *Cube) ApplyMove(move Move) {
	// Handle slice moves
	if move.Slice != NoSlice {
		c.applySliceMove(move)
		return
	}

	// Handle cube rotations
	if move.Rotation != NoRotation {
		c.applyCubeRotation(move)
		return
	}

	// Handle face moves (including wide and layer moves)
	if move.Wide {
		c.applyWideMove(move)
	} else if move.Layer > 0 {
		c.applyLayerMove(move)
	} else {
		// Standard face move
		if move.Double {
			c.rotateFace(move.Face, true)
			c.rotateFace(move.Face, true)
		} else {
			c.rotateFace(move.Face, move.Clockwise)
		}
	}
}

// ApplyMoves applies a sequence of moves to the cube
func (c *Cube) ApplyMoves(moves []Move) {
	for _, move := range moves {
		c.ApplyMove(move)
	}
}

// rotateFace rotates a face and its adjacent edges
func (c *Cube) rotateFace(face Face, clockwise bool) {
	// Rotate the face itself
	c.rotateFaceMatrix(int(face), clockwise)

	// Rotate adjacent edges
	c.rotateAdjacentEdges(face, clockwise)
}

// rotateFaceMatrix rotates a face matrix 90 degrees
func (c *Cube) rotateFaceMatrix(face int, clockwise bool) {
	size := c.Size
	temp := make([][]Color, size)
	for i := range temp {
		temp[i] = make([]Color, size)
	}

	// Copy current face
	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			temp[i][j] = c.Faces[face][i][j]
		}
	}

	// Rotate
	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			if clockwise {
				c.Faces[face][i][j] = temp[size-1-j][i]
			} else {
				c.Faces[face][i][j] = temp[j][size-1-i]
			}
		}
	}
}

// rotateAdjacentEdges rotates the edges adjacent to a face
// For standard face moves, only rotate the outermost layer (layer 0)
func (c *Cube) rotateAdjacentEdges(face Face, clockwise bool) {
	// Standard face moves (R, U, F, etc.) should only affect the outermost layer
	// Wide moves and layer moves are handled separately
	c.rotateAdjacentEdgesLayer(face, clockwise, 0)
}

// rotateAdjacentEdgesLayer rotates the edges for a specific layer
func (c *Cube) rotateAdjacentEdgesLayer(face Face, clockwise bool, layer int) {
	size := c.Size
	temp := make([]Color, size)

	switch face {
	case Front:
		// Front face: rotate edges between Up, Right, Down, Left
		// Layer 0 = outermost, layer 1 = second from outside, etc.
		upRow := size - 1 - layer
		downRow := layer
		leftCol := size - 1 - layer
		rightCol := layer

		if clockwise {
			// Save Up row (upRow)
			for i := 0; i < size; i++ {
				temp[i] = c.Faces[Up][upRow][i]
			}
			// Up ← Left column (leftCol, reversed)
			for i := 0; i < size; i++ {
				c.Faces[Up][upRow][i] = c.Faces[Left][size-1-i][leftCol]
			}
			// Left column ← Down row (downRow, reversed)
			for i := 0; i < size; i++ {
				c.Faces[Left][size-1-i][leftCol] = c.Faces[Down][downRow][i]
			}
			// Down row ← Right column (rightCol, reversed)
			for i := 0; i < size; i++ {
				c.Faces[Down][downRow][i] = c.Faces[Right][size-1-i][rightCol]
			}
			// Right column ← Up row (saved, reversed for correct visual orientation)
			for i := 0; i < size; i++ {
				c.Faces[Right][i][rightCol] = temp[size-1-i]
			}
		} else {
			// Save Up row
			for i := 0; i < size; i++ {
				temp[i] = c.Faces[Up][upRow][i]
			}
			// Up ← Right column
			for i := 0; i < size; i++ {
				c.Faces[Up][upRow][i] = c.Faces[Right][i][rightCol]
			}
			// Right column ← Down row (reversed for correct visual orientation)
			for i := 0; i < size; i++ {
				c.Faces[Right][i][rightCol] = c.Faces[Down][downRow][size-1-i]
			}
			// Down row ← Left column (reversed)
			for i := 0; i < size; i++ {
				c.Faces[Down][downRow][i] = c.Faces[Left][size-1-i][leftCol]
			}
			// Left column ← Up row (saved, reversed for correct visual orientation)
			for i := 0; i < size; i++ {
				c.Faces[Left][i][leftCol] = temp[size-1-i]
			}
		}

	case Back:
		// Back face: rotate edges between Up, Right, Down, Left (clockwise when viewed from back)
		upRow := layer
		downRow := size - 1 - layer
		leftCol := layer
		rightCol := size - 1 - layer

		if clockwise {
			// Save Up row
			for i := 0; i < size; i++ {
				temp[i] = c.Faces[Up][upRow][i]
			}
			// Up ← Right column (reversed)
			for i := 0; i < size; i++ {
				c.Faces[Up][upRow][i] = c.Faces[Right][size-1-i][rightCol]
			}
			// Right column ← Down row (reversed)
			for i := 0; i < size; i++ {
				c.Faces[Right][size-1-i][rightCol] = c.Faces[Down][downRow][size-1-i]
			}
			// Down row ← Left column (reversed)
			for i := 0; i < size; i++ {
				c.Faces[Down][downRow][i] = c.Faces[Left][size-1-i][leftCol]
			}
			// Left column ← Up row (saved)
			for i := 0; i < size; i++ {
				c.Faces[Left][i][leftCol] = temp[i]
			}
		} else {
			// Save Up row
			for i := 0; i < size; i++ {
				temp[i] = c.Faces[Up][upRow][i]
			}
			// Up ← Left column (reversed)
			for i := 0; i < size; i++ {
				c.Faces[Up][upRow][i] = c.Faces[Left][size-1-i][leftCol]
			}
			// Left column ← Down row (reversed)
			for i := 0; i < size; i++ {
				c.Faces[Left][i][leftCol] = c.Faces[Down][downRow][size-1-i]
			}
			// Down row ← Right column (reversed)
			for i := 0; i < size; i++ {
				c.Faces[Down][downRow][size-1-i] = c.Faces[Right][size-1-i][rightCol]
			}
			// Right column ← Up row (saved, NOT reversed)
			for i := 0; i < size; i++ {
				c.Faces[Right][i][rightCol] = temp[i]
			}
		}

	case Left:
		// Left face: rotate edges between Up, Front, Down, Back
		upCol := layer
		downCol := layer
		frontCol := layer
		backCol := size - 1 - layer

		if clockwise {
			// Save Up column
			for i := 0; i < size; i++ {
				temp[i] = c.Faces[Up][i][upCol]
			}
			// Up ← Back column (reversed)
			for i := 0; i < size; i++ {
				c.Faces[Up][i][upCol] = c.Faces[Back][size-1-i][backCol]
			}
			// Back column ← Down column (reversed)
			for i := 0; i < size; i++ {
				c.Faces[Back][size-1-i][backCol] = c.Faces[Down][i][downCol]
			}
			// Down column ← Front column
			for i := 0; i < size; i++ {
				c.Faces[Down][i][downCol] = c.Faces[Front][i][frontCol]
			}
			// Front column ← Up column (saved, reversed for correct visual orientation)
			for i := 0; i < size; i++ {
				c.Faces[Front][i][frontCol] = temp[size-1-i]
			}
		} else {
			// Save Up column
			for i := 0; i < size; i++ {
				temp[i] = c.Faces[Up][i][upCol]
			}
			// Up ← Front column
			for i := 0; i < size; i++ {
				c.Faces[Up][i][upCol] = c.Faces[Front][i][frontCol]
			}
			// Front column ← Down column
			for i := 0; i < size; i++ {
				c.Faces[Front][i][frontCol] = c.Faces[Down][i][downCol]
			}
			// Down column ← Back column (reversed)
			for i := 0; i < size; i++ {
				c.Faces[Down][i][downCol] = c.Faces[Back][size-1-i][backCol]
			}
			// Back column ← Up column (saved, reversed)
			for i := 0; i < size; i++ {
				c.Faces[Back][size-1-i][backCol] = temp[i]
			}
		}

	case Right:
		// Right face: rotate edges between Up, Back, Down, Front (clockwise from right side view)
		upCol := size - 1 - layer
		downCol := size - 1 - layer
		frontCol := size - 1 - layer
		backCol := layer

		if clockwise {
			// Save Up column
			for i := 0; i < size; i++ {
				temp[i] = c.Faces[Up][i][upCol]
			}
			// Up ← Front column
			for i := 0; i < size; i++ {
				c.Faces[Up][i][upCol] = c.Faces[Front][i][frontCol]
			}
			// Front column ← Down column
			for i := 0; i < size; i++ {
				c.Faces[Front][i][frontCol] = c.Faces[Down][i][downCol]
			}
			// Down column ← Back column (reversed)
			for i := 0; i < size; i++ {
				c.Faces[Down][i][downCol] = c.Faces[Back][size-1-i][backCol]
			}
			// Back column ← Up column (saved, reversed)
			for i := 0; i < size; i++ {
				c.Faces[Back][size-1-i][backCol] = temp[i]
			}
		} else {
			// Save Up column
			for i := 0; i < size; i++ {
				temp[i] = c.Faces[Up][i][upCol]
			}
			// Up ← Back column (reversed)
			for i := 0; i < size; i++ {
				c.Faces[Up][i][upCol] = c.Faces[Back][size-1-i][backCol]
			}
			// Back column ← Down column (reversed)
			for i := 0; i < size; i++ {
				c.Faces[Back][size-1-i][backCol] = c.Faces[Down][i][downCol]
			}
			// Down column ← Front column
			for i := 0; i < size; i++ {
				c.Faces[Down][i][downCol] = c.Faces[Front][i][frontCol]
			}
			// Front column ← Up column (saved)
			for i := 0; i < size; i++ {
				c.Faces[Front][i][frontCol] = temp[i]
			}
		}

	case Up:
		// Up face: rotate edges between Back, Right, Front, Left
		backRow := layer
		rightRow := layer
		frontRow := layer
		leftRow := layer

		if clockwise {
			// Save Back row
			for i := 0; i < size; i++ {
				temp[i] = c.Faces[Back][backRow][i]
			}
			// Back ← Left row
			for i := 0; i < size; i++ {
				c.Faces[Back][backRow][i] = c.Faces[Left][leftRow][i]
			}
			// Left ← Front row
			for i := 0; i < size; i++ {
				c.Faces[Left][leftRow][i] = c.Faces[Front][frontRow][i]
			}
			// Front ← Right row
			for i := 0; i < size; i++ {
				c.Faces[Front][frontRow][i] = c.Faces[Right][rightRow][i]
			}
			// Right ← Back row (saved)
			for i := 0; i < size; i++ {
				c.Faces[Right][rightRow][i] = temp[i]
			}
		} else {
			// Save Back row
			for i := 0; i < size; i++ {
				temp[i] = c.Faces[Back][backRow][i]
			}
			// Back ← Right row
			for i := 0; i < size; i++ {
				c.Faces[Back][backRow][i] = c.Faces[Right][rightRow][i]
			}
			// Right ← Front row
			for i := 0; i < size; i++ {
				c.Faces[Right][rightRow][i] = c.Faces[Front][frontRow][i]
			}
			// Front ← Left row
			for i := 0; i < size; i++ {
				c.Faces[Front][frontRow][i] = c.Faces[Left][leftRow][i]
			}
			// Left ← Back row (saved)
			for i := 0; i < size; i++ {
				c.Faces[Left][leftRow][i] = temp[i]
			}
		}

	case Down:
		// Down face: rotate edges between Front, Right, Back, Left
		frontRow := size - 1 - layer
		rightRow := size - 1 - layer
		backRow := size - 1 - layer
		leftRow := size - 1 - layer

		if clockwise {
			// Save Front row
			for i := 0; i < size; i++ {
				temp[i] = c.Faces[Front][frontRow][i]
			}
			// Front ← Left row
			for i := 0; i < size; i++ {
				c.Faces[Front][frontRow][i] = c.Faces[Left][leftRow][i]
			}
			// Left ← Back row
			for i := 0; i < size; i++ {
				c.Faces[Left][leftRow][i] = c.Faces[Back][backRow][i]
			}
			// Back ← Right row
			for i := 0; i < size; i++ {
				c.Faces[Back][backRow][i] = c.Faces[Right][rightRow][i]
			}
			// Right ← Front row (saved)
			for i := 0; i < size; i++ {
				c.Faces[Right][rightRow][i] = temp[i]
			}
		} else {
			// Save Front row
			for i := 0; i < size; i++ {
				temp[i] = c.Faces[Front][frontRow][i]
			}
			// Front ← Right row
			for i := 0; i < size; i++ {
				c.Faces[Front][frontRow][i] = c.Faces[Right][rightRow][i]
			}
			// Right ← Back row
			for i := 0; i < size; i++ {
				c.Faces[Right][rightRow][i] = c.Faces[Back][backRow][i]
			}
			// Back ← Left row
			for i := 0; i < size; i++ {
				c.Faces[Back][backRow][i] = c.Faces[Left][leftRow][i]
			}
			// Left ← Front row (saved)
			for i := 0; i < size; i++ {
				c.Faces[Left][leftRow][i] = temp[i]
			}
		}
	}
}

// applySliceMove applies middle slice moves (M, E, S)
func (c *Cube) applySliceMove(move Move) {
	if move.Double {
		c.executeSliceMove(move.Slice, move.Clockwise)
		c.executeSliceMove(move.Slice, move.Clockwise)
	} else {
		c.executeSliceMove(move.Slice, move.Clockwise)
	}
}

// executeSliceMove performs the actual slice rotation
func (c *Cube) executeSliceMove(slice SliceType, clockwise bool) {
	size := c.Size
	temp := make([]Color, size)

	switch slice {
	case M_Slice:
		// M slice: between L and R faces, moves like L but affects middle
		// Only works on odd-sized cubes (3x3, 5x5), affects center column
		if size%2 == 0 {
			return // M slice undefined for even cubes
		}
		centerCol := size / 2

		if clockwise {
			// M moves like L: Up ← Back ← Down ← Front ← Up
			for i := 0; i < size; i++ {
				temp[i] = c.Faces[Up][i][centerCol]
			}
			for i := 0; i < size; i++ {
				c.Faces[Up][i][centerCol] = c.Faces[Back][size-1-i][centerCol]
			}
			for i := 0; i < size; i++ {
				c.Faces[Back][size-1-i][centerCol] = c.Faces[Down][size-1-i][centerCol]
			}
			for i := 0; i < size; i++ {
				c.Faces[Down][i][centerCol] = c.Faces[Front][i][centerCol]
			}
			for i := 0; i < size; i++ {
				c.Faces[Front][i][centerCol] = temp[i]
			}
		} else {
			// M' moves opposite to M
			for i := 0; i < size; i++ {
				temp[i] = c.Faces[Up][i][centerCol]
			}
			for i := 0; i < size; i++ {
				c.Faces[Up][i][centerCol] = c.Faces[Front][i][centerCol]
			}
			for i := 0; i < size; i++ {
				c.Faces[Front][i][centerCol] = c.Faces[Down][i][centerCol]
			}
			for i := 0; i < size; i++ {
				c.Faces[Down][size-1-i][centerCol] = c.Faces[Back][size-1-i][centerCol]
			}
			for i := 0; i < size; i++ {
				c.Faces[Back][size-1-i][centerCol] = temp[i]
			}
		}

	case E_Slice:
		// E slice: between U and D faces, moves like D but affects middle
		if size%2 == 0 {
			return // E slice undefined for even cubes
		}
		centerRow := size / 2

		if clockwise {
			// E moves like D: Front ← Left ← Back ← Right ← Front
			for i := 0; i < size; i++ {
				temp[i] = c.Faces[Front][centerRow][i]
			}
			for i := 0; i < size; i++ {
				c.Faces[Front][centerRow][i] = c.Faces[Left][centerRow][i]
			}
			for i := 0; i < size; i++ {
				c.Faces[Left][centerRow][i] = c.Faces[Back][centerRow][i]
			}
			for i := 0; i < size; i++ {
				c.Faces[Back][centerRow][i] = c.Faces[Right][centerRow][i]
			}
			for i := 0; i < size; i++ {
				c.Faces[Right][centerRow][i] = temp[i]
			}
		} else {
			// E' moves opposite
			for i := 0; i < size; i++ {
				temp[i] = c.Faces[Front][centerRow][i]
			}
			for i := 0; i < size; i++ {
				c.Faces[Front][centerRow][i] = c.Faces[Right][centerRow][i]
			}
			for i := 0; i < size; i++ {
				c.Faces[Right][centerRow][i] = c.Faces[Back][centerRow][i]
			}
			for i := 0; i < size; i++ {
				c.Faces[Back][centerRow][i] = c.Faces[Left][centerRow][i]
			}
			for i := 0; i < size; i++ {
				c.Faces[Left][centerRow][i] = temp[i]
			}
		}

	case S_Slice:
		// S slice: between F and B faces, moves like F but affects middle
		if size%2 == 0 {
			return // S slice undefined for even cubes
		}
		centerLayer := size / 2

		if clockwise {
			// S moves like F
			for i := 0; i < size; i++ {
				temp[i] = c.Faces[Up][size-1-centerLayer][i]
			}
			for i := 0; i < size; i++ {
				c.Faces[Up][size-1-centerLayer][i] = c.Faces[Left][size-1-i][centerLayer]
			}
			for i := 0; i < size; i++ {
				c.Faces[Left][size-1-i][centerLayer] = c.Faces[Down][centerLayer][size-1-i]
			}
			for i := 0; i < size; i++ {
				c.Faces[Down][centerLayer][i] = c.Faces[Right][i][size-1-centerLayer]
			}
			for i := 0; i < size; i++ {
				c.Faces[Right][i][size-1-centerLayer] = temp[i]
			}
		} else {
			// S' moves opposite
			for i := 0; i < size; i++ {
				temp[i] = c.Faces[Up][size-1-centerLayer][i]
			}
			for i := 0; i < size; i++ {
				c.Faces[Up][size-1-centerLayer][i] = c.Faces[Right][i][size-1-centerLayer]
			}
			for i := 0; i < size; i++ {
				c.Faces[Right][i][size-1-centerLayer] = c.Faces[Down][centerLayer][size-1-i]
			}
			for i := 0; i < size; i++ {
				c.Faces[Down][centerLayer][size-1-i] = c.Faces[Left][size-1-i][centerLayer]
			}
			for i := 0; i < size; i++ {
				c.Faces[Left][i][centerLayer] = temp[size-1-i]
			}
		}
	}
}

// applyWideMove applies wide moves (Rw, Fw, etc.)
func (c *Cube) applyWideMove(move Move) {
	// Wide move = face move + adjacent inner layers
	// For Rw: apply R + inner layer moves
	layers := move.WideDepth
	if layers <= 0 {
		layers = 2 // Default to 2 layers
	}

	for layer := 0; layer < layers; layer++ {
		if layer == 0 {
			// Outermost layer includes face rotation
			if move.Double {
				c.rotateFace(move.Face, move.Clockwise)
				c.rotateFace(move.Face, move.Clockwise)
			} else {
				c.rotateFace(move.Face, move.Clockwise)
			}
		} else {
			// Inner layers - just edge rotation
			if move.Double {
				c.rotateAdjacentEdgesLayer(move.Face, move.Clockwise, layer)
				c.rotateAdjacentEdgesLayer(move.Face, move.Clockwise, layer)
			} else {
				c.rotateAdjacentEdgesLayer(move.Face, move.Clockwise, layer)
			}
		}
	}
}

// applyLayerMove applies layer-specific moves (2R, 3L, etc.)
func (c *Cube) applyLayerMove(move Move) {
	// Layer move affects only the specified inner layer (no face rotation)
	if move.Double {
		c.rotateAdjacentEdgesLayer(move.Face, move.Clockwise, move.Layer)
		c.rotateAdjacentEdgesLayer(move.Face, move.Clockwise, move.Layer)
	} else {
		c.rotateAdjacentEdgesLayer(move.Face, move.Clockwise, move.Layer)
	}
}

// applyCubeRotation applies whole cube rotations (x, y, z)
func (c *Cube) applyCubeRotation(move Move) {
	if move.Double {
		c.executeCubeRotation(move.Rotation, move.Clockwise)
		c.executeCubeRotation(move.Rotation, move.Clockwise)
	} else {
		c.executeCubeRotation(move.Rotation, move.Clockwise)
	}
}

// executeCubeRotation performs the actual cube rotation
func (c *Cube) executeCubeRotation(rotation RotationType, clockwise bool) {
	switch rotation {
	case X_Rotation:
		// x rotation: rotate around R face axis
		// This reorients the entire cube - faces change positions
		if clockwise {
			// F→D, D→B, B→U, U→F, L rotates CCW, R rotates CW
			c.swapFaces(Front, Down, Back, Up)
			c.rotateFaceMatrix(int(Left), false) // Left face rotates CCW
			c.rotateFaceMatrix(int(Right), true) // Right face rotates CW
		} else {
			// F→U, U→B, B→D, D→F, L rotates CW, R rotates CCW
			c.swapFaces(Front, Up, Back, Down)
			c.rotateFaceMatrix(int(Left), true)   // Left face rotates CW
			c.rotateFaceMatrix(int(Right), false) // Right face rotates CCW
		}

	case Y_Rotation:
		// y rotation: rotate around U face axis
		if clockwise {
			// F→L, L→B, B→R, R→F, U rotates CW, D rotates CCW
			c.swapFaces(Front, Left, Back, Right)
			c.rotateFaceMatrix(int(Up), true)    // Up face rotates CW
			c.rotateFaceMatrix(int(Down), false) // Down face rotates CCW
		} else {
			// F→R, R→B, B→L, L→F, U rotates CCW, D rotates CW
			c.swapFaces(Front, Right, Back, Left)
			c.rotateFaceMatrix(int(Up), false)  // Up face rotates CCW
			c.rotateFaceMatrix(int(Down), true) // Down face rotates CW
		}

	case Z_Rotation:
		// z rotation: rotate around F face axis
		if clockwise {
			// U→L, L→D, D→R, R→U, F rotates CW, B rotates CCW
			c.swapFaces(Up, Left, Down, Right)
			c.rotateFaceMatrix(int(Front), true) // Front face rotates CW
			c.rotateFaceMatrix(int(Back), false) // Back face rotates CCW
		} else {
			// U→R, R→D, D→L, L→U, F rotates CCW, B rotates CW
			c.swapFaces(Up, Right, Down, Left)
			c.rotateFaceMatrix(int(Front), false) // Front face rotates CCW
			c.rotateFaceMatrix(int(Back), true)   // Back face rotates CW
		}
	}
}

// swapFaces swaps the positions of four faces in a cycle
func (c *Cube) swapFaces(f1, f2, f3, f4 Face) {
	size := c.Size
	temp := make([][]Color, size)
	for i := range temp {
		temp[i] = make([]Color, size)
		for j := range temp[i] {
			temp[i][j] = c.Faces[f1][i][j]
		}
	}

	// f1 ← f4
	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			c.Faces[f1][i][j] = c.Faces[f4][i][j]
		}
	}
	// f4 ← f3
	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			c.Faces[f4][i][j] = c.Faces[f3][i][j]
		}
	}
	// f3 ← f2
	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			c.Faces[f3][i][j] = c.Faces[f2][i][j]
		}
	}
	// f2 ← f1 (saved)
	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			c.Faces[f2][i][j] = temp[i][j]
		}
	}
}
