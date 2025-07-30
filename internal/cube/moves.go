package cube

import (
	"fmt"
	"strings"
)

// Move represents a cube move
type Move struct {
	Face      Face
	Clockwise bool
	Double    bool
}

// String returns the standard notation for the move
func (m Move) String() string {
	notation := m.Face.String()
	if m.Double {
		notation += "2"
	} else if !m.Clockwise {
		notation += "'"
	}
	return notation
}

// ParseMove parses a move from standard notation (e.g., "R", "U'", "F2")
func ParseMove(notation string) (Move, error) {
	notation = strings.TrimSpace(strings.ToUpper(notation))
	if len(notation) == 0 {
		return Move{}, fmt.Errorf("empty move notation")
	}

	var face Face
	switch notation[0] {
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
		return Move{}, fmt.Errorf("invalid face: %c", notation[0])
	}

	move := Move{Face: face, Clockwise: true}

	// Check for modifiers
	for i := 1; i < len(notation); i++ {
		switch notation[i] {
		case '\'':
			move.Clockwise = false
		case '2':
			move.Double = true
		default:
			return Move{}, fmt.Errorf("invalid modifier: %c", notation[i])
		}
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

// ApplyMove applies a single move to the cube
func (c *Cube) ApplyMove(move Move) {
	if move.Double {
		c.rotateFace(move.Face, true)
		c.rotateFace(move.Face, true)
	} else {
		c.rotateFace(move.Face, move.Clockwise)
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
func (c *Cube) rotateAdjacentEdges(face Face, clockwise bool) {
	size := c.Size

	switch face {
	case Front:
		// Front face: rotate edges between Up, Right, Down, Left
		temp := make([]Color, size)

		if clockwise {
			// Save Up bottom row
			for i := 0; i < size; i++ {
				temp[i] = c.Faces[Up][size-1][i]
			}
			// Up ← Left right column (reversed)
			for i := 0; i < size; i++ {
				c.Faces[Up][size-1][i] = c.Faces[Left][size-1-i][size-1]
			}
			// Left right column ← Down top row (reversed)
			for i := 0; i < size; i++ {
				c.Faces[Left][size-1-i][size-1] = c.Faces[Down][0][size-1-i]
			}
			// Down top row ← Right left column
			for i := 0; i < size; i++ {
				c.Faces[Down][0][i] = c.Faces[Right][i][0]
			}
			// Right left column ← Up bottom row (saved)
			for i := 0; i < size; i++ {
				c.Faces[Right][i][0] = temp[i]
			}
		} else {
			// Save Up bottom row
			for i := 0; i < size; i++ {
				temp[i] = c.Faces[Up][size-1][i]
			}
			// Up ← Right left column
			for i := 0; i < size; i++ {
				c.Faces[Up][size-1][i] = c.Faces[Right][i][0]
			}
			// Right left column ← Down top row (reversed)
			for i := 0; i < size; i++ {
				c.Faces[Right][i][0] = c.Faces[Down][0][size-1-i]
			}
			// Down top row ← Left right column (reversed)
			for i := 0; i < size; i++ {
				c.Faces[Down][0][size-1-i] = c.Faces[Left][size-1-i][size-1]
			}
			// Left right column ← Up bottom row (saved)
			for i := 0; i < size; i++ {
				c.Faces[Left][i][size-1] = temp[size-1-i]
			}
		}

	case Back:
		// Back face: rotate edges between Up, Left, Down, Right
		temp := make([]Color, size)

		if clockwise {
			// Save Up top row
			for i := 0; i < size; i++ {
				temp[i] = c.Faces[Up][0][i]
			}
			// Up ← Right right column (reversed)
			for i := 0; i < size; i++ {
				c.Faces[Up][0][i] = c.Faces[Right][size-1-i][size-1]
			}
			// Right right column ← Down bottom row (reversed)
			for i := 0; i < size; i++ {
				c.Faces[Right][size-1-i][size-1] = c.Faces[Down][size-1][size-1-i]
			}
			// Down bottom row ← Left left column
			for i := 0; i < size; i++ {
				c.Faces[Down][size-1][i] = c.Faces[Left][i][0]
			}
			// Left left column ← Up top row (saved)
			for i := 0; i < size; i++ {
				c.Faces[Left][i][0] = temp[size-1-i]
			}
		} else {
			// Save Up top row
			for i := 0; i < size; i++ {
				temp[i] = c.Faces[Up][0][i]
			}
			// Up ← Left left column
			for i := 0; i < size; i++ {
				c.Faces[Up][0][i] = c.Faces[Left][size-1-i][0]
			}
			// Left left column ← Down bottom row (reversed)
			for i := 0; i < size; i++ {
				c.Faces[Left][i][0] = c.Faces[Down][size-1][size-1-i]
			}
			// Down bottom row ← Right right column (reversed)
			for i := 0; i < size; i++ {
				c.Faces[Down][size-1][size-1-i] = c.Faces[Right][size-1-i][size-1]
			}
			// Right right column ← Up top row (saved)
			for i := 0; i < size; i++ {
				c.Faces[Right][i][size-1] = temp[i]
			}
		}

	case Left:
		// Left face: rotate edges between Up, Front, Down, Back
		temp := make([]Color, size)

		if clockwise {
			// Save Up left column
			for i := 0; i < size; i++ {
				temp[i] = c.Faces[Up][i][0]
			}
			// Up ← Back right column (reversed)
			for i := 0; i < size; i++ {
				c.Faces[Up][i][0] = c.Faces[Back][size-1-i][size-1]
			}
			// Back right column ← Down left column (reversed)
			for i := 0; i < size; i++ {
				c.Faces[Back][size-1-i][size-1] = c.Faces[Down][size-1-i][0]
			}
			// Down left column ← Front left column
			for i := 0; i < size; i++ {
				c.Faces[Down][i][0] = c.Faces[Front][i][0]
			}
			// Front left column ← Up left column (saved)
			for i := 0; i < size; i++ {
				c.Faces[Front][i][0] = temp[i]
			}
		} else {
			// Save Up left column
			for i := 0; i < size; i++ {
				temp[i] = c.Faces[Up][i][0]
			}
			// Up ← Front left column
			for i := 0; i < size; i++ {
				c.Faces[Up][i][0] = c.Faces[Front][i][0]
			}
			// Front left column ← Down left column
			for i := 0; i < size; i++ {
				c.Faces[Front][i][0] = c.Faces[Down][i][0]
			}
			// Down left column ← Back right column (reversed)
			for i := 0; i < size; i++ {
				c.Faces[Down][i][0] = c.Faces[Back][size-1-i][size-1]
			}
			// Back right column ← Up left column (saved, reversed)
			for i := 0; i < size; i++ {
				c.Faces[Back][size-1-i][size-1] = temp[i]
			}
		}

	case Right:
		// Right face: rotate edges between Up, Back, Down, Front
		temp := make([]Color, size)

		if clockwise {
			// Save Up right column
			for i := 0; i < size; i++ {
				temp[i] = c.Faces[Up][i][size-1]
			}
			// Up ← Front right column
			for i := 0; i < size; i++ {
				c.Faces[Up][i][size-1] = c.Faces[Front][i][size-1]
			}
			// Front right column ← Down right column
			for i := 0; i < size; i++ {
				c.Faces[Front][i][size-1] = c.Faces[Down][i][size-1]
			}
			// Down right column ← Back left column (reversed)
			for i := 0; i < size; i++ {
				c.Faces[Down][i][size-1] = c.Faces[Back][size-1-i][0]
			}
			// Back left column ← Up right column (saved, reversed)
			for i := 0; i < size; i++ {
				c.Faces[Back][size-1-i][0] = temp[i]
			}
		} else {
			// Save Up right column
			for i := 0; i < size; i++ {
				temp[i] = c.Faces[Up][i][size-1]
			}
			// Up ← Back left column (reversed)
			for i := 0; i < size; i++ {
				c.Faces[Up][i][size-1] = c.Faces[Back][size-1-i][0]
			}
			// Back left column ← Down right column (reversed)
			for i := 0; i < size; i++ {
				c.Faces[Back][size-1-i][0] = c.Faces[Down][size-1-i][size-1]
			}
			// Down right column ← Front right column
			for i := 0; i < size; i++ {
				c.Faces[Down][i][size-1] = c.Faces[Front][i][size-1]
			}
			// Front right column ← Up right column (saved)
			for i := 0; i < size; i++ {
				c.Faces[Front][i][size-1] = temp[i]
			}
		}

	case Up:
		// Up face: rotate edges between Back, Right, Front, Left
		temp := make([]Color, size)

		if clockwise {
			// Save Back top row
			for i := 0; i < size; i++ {
				temp[i] = c.Faces[Back][0][i]
			}
			// Back ← Left top row
			for i := 0; i < size; i++ {
				c.Faces[Back][0][i] = c.Faces[Left][0][i]
			}
			// Left ← Front top row
			for i := 0; i < size; i++ {
				c.Faces[Left][0][i] = c.Faces[Front][0][i]
			}
			// Front ← Right top row
			for i := 0; i < size; i++ {
				c.Faces[Front][0][i] = c.Faces[Right][0][i]
			}
			// Right ← Back top row (saved)
			for i := 0; i < size; i++ {
				c.Faces[Right][0][i] = temp[i]
			}
		} else {
			// Save Back top row
			for i := 0; i < size; i++ {
				temp[i] = c.Faces[Back][0][i]
			}
			// Back ← Right top row
			for i := 0; i < size; i++ {
				c.Faces[Back][0][i] = c.Faces[Right][0][i]
			}
			// Right ← Front top row
			for i := 0; i < size; i++ {
				c.Faces[Right][0][i] = c.Faces[Front][0][i]
			}
			// Front ← Left top row
			for i := 0; i < size; i++ {
				c.Faces[Front][0][i] = c.Faces[Left][0][i]
			}
			// Left ← Back top row (saved)
			for i := 0; i < size; i++ {
				c.Faces[Left][0][i] = temp[i]
			}
		}

	case Down:
		// Down face: rotate edges between Front, Right, Back, Left
		temp := make([]Color, size)

		if clockwise {
			// Save Front bottom row
			for i := 0; i < size; i++ {
				temp[i] = c.Faces[Front][size-1][i]
			}
			// Front ← Left bottom row
			for i := 0; i < size; i++ {
				c.Faces[Front][size-1][i] = c.Faces[Left][size-1][i]
			}
			// Left ← Back bottom row
			for i := 0; i < size; i++ {
				c.Faces[Left][size-1][i] = c.Faces[Back][size-1][i]
			}
			// Back ← Right bottom row
			for i := 0; i < size; i++ {
				c.Faces[Back][size-1][i] = c.Faces[Right][size-1][i]
			}
			// Right ← Front bottom row (saved)
			for i := 0; i < size; i++ {
				c.Faces[Right][size-1][i] = temp[i]
			}
		} else {
			// Save Front bottom row
			for i := 0; i < size; i++ {
				temp[i] = c.Faces[Front][size-1][i]
			}
			// Front ← Right bottom row
			for i := 0; i < size; i++ {
				c.Faces[Front][size-1][i] = c.Faces[Right][size-1][i]
			}
			// Right ← Back bottom row
			for i := 0; i < size; i++ {
				c.Faces[Right][size-1][i] = c.Faces[Back][size-1][i]
			}
			// Back ← Left bottom row
			for i := 0; i < size; i++ {
				c.Faces[Back][size-1][i] = c.Faces[Left][size-1][i]
			}
			// Left ← Front bottom row (saved)
			for i := 0; i < size; i++ {
				c.Faces[Left][size-1][i] = temp[i]
			}
		}
	}
}
