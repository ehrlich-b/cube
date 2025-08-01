package cube

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseMove parses a move from advanced notation
// Supports: R, U', F2, 2R, Rw, 2Fw, M, E', S2, x, y', z2
func ParseMove(notation string) (Move, error) {
	notation = strings.TrimSpace(notation)
	if len(notation) == 0 {
		return Move{}, fmt.Errorf("empty move notation")
	}

	move := Move{Clockwise: true} // Default to clockwise

	// Parse modifiers at the end
	for len(notation) > 0 {
		lastChar := notation[len(notation)-1]
		if lastChar == '\'' {
			move.Clockwise = false
			notation = notation[:len(notation)-1]
		} else if lastChar == '2' {
			move.Double = true
			notation = notation[:len(notation)-1]
		} else {
			break
		}
	}

	if len(notation) == 0 {
		return Move{}, fmt.Errorf("invalid move notation")
	}

	// Check for wide moves (w suffix)
	if strings.HasSuffix(notation, "w") {
		move.Wide = true
		notation = notation[:len(notation)-1]
	}

	// Check for numbered moves (starts with digit)
	if len(notation) > 0 && notation[0] >= '0' && notation[0] <= '9' {
		// Extract number
		numStr := ""
		i := 0
		for i < len(notation) && notation[i] >= '0' && notation[i] <= '9' {
			numStr += string(notation[i])
			i++
		}
		if len(numStr) > 0 {
			num, err := strconv.Atoi(numStr)
			if err != nil {
				return Move{}, fmt.Errorf("invalid number in move: %s", numStr)
			}
			if move.Wide {
				move.WideDepth = num
			} else {
				move.Layer = num - 1 // Convert to 0-indexed
			}
			notation = notation[i:]
		}
	}

	// Parse the face/slice/rotation
	switch notation {
	case "R":
		move.Face = Right
	case "L":
		move.Face = Left
	case "U":
		move.Face = Up
	case "D":
		move.Face = Down
	case "F":
		move.Face = Front
	case "B":
		move.Face = Back
	case "M":
		move.Slice = M_Slice
	case "E":
		move.Slice = E_Slice
	case "S":
		move.Slice = S_Slice
	case "x":
		move.Rotation = X_Rotation
	case "y":
		move.Rotation = Y_Rotation
	case "z":
		move.Rotation = Z_Rotation
	default:
		return Move{}, fmt.Errorf("unknown move notation: %s", notation)
	}

	return move, nil
}

// ParseMoves parses a sequence of moves from a string
func ParseMoves(sequence string) ([]Move, error) {
	sequence = strings.TrimSpace(sequence)
	if len(sequence) == 0 {
		return []Move{}, nil
	}

	// Split by spaces
	parts := strings.Fields(sequence)
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

// ParseScramble is an alias for ParseMoves for backward compatibility
func ParseScramble(sequence string) ([]Move, error) {
	return ParseMoves(sequence)
}

// String returns a string representation of the move
func (m Move) String() string {
	var result string

	// Handle slice moves
	if m.Slice != NoSlice {
		switch m.Slice {
		case M_Slice:
			result = "M"
		case E_Slice:
			result = "E"
		case S_Slice:
			result = "S"
		}
	} else if m.Rotation != NoRotation {
		// Handle cube rotations
		switch m.Rotation {
		case X_Rotation:
			result = "x"
		case Y_Rotation:
			result = "y"
		case Z_Rotation:
			result = "z"
		}
	} else {
		// Handle face moves with layer/wide notation

		// Add layer number prefix for numbered layer moves (2R, 3L, etc.)
		if m.Layer > 0 {
			result += fmt.Sprintf("%d", m.Layer+1) // Convert back to 1-indexed
		}

		// Add face letter
		switch m.Face {
		case Right:
			result += "R"
		case Left:
			result += "L"
		case Up:
			result += "U"
		case Down:
			result += "D"
		case Front:
			result += "F"
		case Back:
			result += "B"
		}

		// Add wide suffix
		if m.Wide {
			result += "w"
		}
	}

	// Add modifiers
	if m.Double {
		result += "2"
	} else if !m.Clockwise {
		result += "'"
	}

	return result
}
