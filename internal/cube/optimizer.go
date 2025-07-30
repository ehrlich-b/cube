package cube

import (
	"strings"
)

// OptimizeMoves takes a sequence of moves and optimizes it by:
// - Combining consecutive moves on same face: R R -> R2, R R R -> R'
// - Removing canceling moves: R R' -> (nothing), R2 R2 -> (nothing)
// - Simplifying double moves: R2 R2 -> (nothing), R2 R -> R', R2 R' -> R
func OptimizeMoves(moves []Move) []Move {
	if len(moves) == 0 {
		return moves
	}

	optimized := make([]Move, 0, len(moves))

	for i := 0; i < len(moves); i++ {
		currentMove := moves[i]

		// Skip moves that don't affect the cube (only cube rotations for now)
		if currentMove.Rotation != NoRotation {
			optimized = append(optimized, currentMove)
			continue
		}

		// Try to combine with previous move if it's the same face
		if len(optimized) > 0 {
			lastMove := &optimized[len(optimized)-1]

			// Same face moves can be combined
			if lastMove.Face == currentMove.Face &&
				lastMove.Wide == currentMove.Wide &&
				lastMove.Layer == currentMove.Layer &&
				lastMove.Slice == NoSlice && currentMove.Slice == NoSlice {

				combined := combineSameFaceMoves(*lastMove, currentMove)
				if combined == nil {
					// Moves cancel out - remove the last move
					optimized = optimized[:len(optimized)-1]
				} else {
					// Update the last move with combined result
					optimized[len(optimized)-1] = *combined
				}
				continue
			}
		}

		// No combination possible, add the move
		optimized = append(optimized, currentMove)
	}

	return optimized
}

// combineSameFaceMoves combines two moves on the same face
// Returns nil if the moves cancel out completely
func combineSameFaceMoves(first, second Move) *Move {
	// Convert moves to "quarter turn count" for easier math
	firstCount := moveToQuarterTurns(first)
	secondCount := moveToQuarterTurns(second)

	totalCount := (firstCount + secondCount) % 4

	// If total is 0, moves cancel out
	if totalCount == 0 {
		return nil
	}

	// Create optimized move from total quarter turns
	return quarterTurnsToMove(first.Face, first.Wide, first.Layer, totalCount)
}

// moveToQuarterTurns converts a move to number of quarter turns (1-3)
func moveToQuarterTurns(move Move) int {
	if move.Double {
		return 2
	} else if move.Clockwise {
		return 1
	} else {
		return 3 // Counter-clockwise = 3 quarter turns clockwise
	}
}

// quarterTurnsToMove converts quarter turn count back to a Move
func quarterTurnsToMove(face Face, wide bool, layer int, quarterTurns int) *Move {
	switch quarterTurns {
	case 1:
		return &Move{Face: face, Wide: wide, Layer: layer, Clockwise: true, Double: false}
	case 2:
		return &Move{Face: face, Wide: wide, Layer: layer, Clockwise: true, Double: true}
	case 3:
		return &Move{Face: face, Wide: wide, Layer: layer, Clockwise: false, Double: false}
	default:
		return nil // Should never happen
	}
}

// OptimizeScramble takes a scramble string and returns an optimized version
func OptimizeScramble(scramble string) (string, error) {
	moves, err := ParseScramble(scramble)
	if err != nil {
		return "", err
	}

	optimized := OptimizeMoves(moves)

	// Convert back to string
	var result []string
	for _, move := range optimized {
		result = append(result, move.String())
	}

	return strings.Join(result, " "), nil
}

// GetMoveCount returns the total number of moves in a sequence after optimization
func GetMoveCount(moves []Move) int {
	return len(OptimizeMoves(moves))
}

// IsCancellingSequence checks if a sequence of moves results in no net change
func IsCancellingSequence(moves []Move) bool {
	return len(OptimizeMoves(moves)) == 0
}
