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
	
	// Rotate adjacent edges - this is a simplified implementation
	// For a complete implementation, we would need to handle all edge rotations
	// TODO: Implement complete edge rotation logic for all cube sizes
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