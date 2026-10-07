package cube

import "fmt"

// ValidateMoves checks a sequence before applying it, preventing partial changes
// when a later move is invalid for the cube's dimension.
func ValidateMoves(moves []Move, size int) error {
	if size < 2 {
		return fmt.Errorf("cube dimension must be at least 2 (got %d)", size)
	}
	for _, move := range moves {
		if err := move.Validate(size); err != nil {
			return err
		}
	}
	return nil
}

// Validate rejects moves that cannot be represented on a cube of this size.
func (move Move) Validate(size int) error {
	if size < 2 {
		return fmt.Errorf("cube dimension must be at least 2 (got %d)", size)
	}
	if move.Slice != NoSlice {
		if move.Slice < M_Slice || move.Slice > S_Slice || move.Rotation != NoRotation || move.Wide || move.Layer != 0 || move.WideDepth != 0 {
			return fmt.Errorf("invalid slice move")
		}
		if size%2 == 0 {
			return fmt.Errorf("slice move %s is unsupported on an even %dx%dx%d cube; use numbered layer turns", move, size, size, size)
		}
		return nil
	}
	if move.Rotation != NoRotation {
		if move.Rotation < X_Rotation || move.Rotation > Z_Rotation || move.Wide || move.Layer != 0 || move.WideDepth != 0 {
			return fmt.Errorf("invalid cube rotation")
		}
		return nil
	}
	if move.Face < Front || move.Face > Down {
		return fmt.Errorf("invalid move face %d", move.Face)
	}
	if move.Layer < 0 || move.Layer >= size {
		return fmt.Errorf("move %s selects layer %d outside a %dx%dx%d cube (layers 1-%d)", move, move.Layer+1, size, size, size, size)
	}
	if move.Wide {
		depth := move.WideDepth
		if depth == 0 {
			depth = 2
		}
		if depth < 1 || depth > size || move.Layer != 0 {
			return fmt.Errorf("move %s selects invalid wide layer depth %d for a %dx%dx%d cube", move, depth, size, size, size)
		}
	} else if move.WideDepth != 0 {
		return fmt.Errorf("wide layer depth requires a wide turn")
	}
	return nil
}

// ApplyMove applies a single valid move, returning an error without mutation
// when a layer is out of range or a slice is unsupported.
func (c *Cube) ApplyMove(move Move) error {
	if err := move.Validate(c.Size); err != nil {
		return err
	}
	c.applyMove(move)
	return nil
}

func (c *Cube) applyMove(move Move) {
	moveType, quarterTurns := moveToMoveType(move)
	layers := getAffectedLayers(move, c.Size)

	for _, layer := range layers {
		perm := getPermutation(c.Size, moveType, layer, quarterTurns)
		applyPermutation(c, perm)
	}
}

// ApplyMoves applies a sequence of moves to the cube
func (c *Cube) ApplyMoves(moves []Move) error {
	if err := ValidateMoves(moves, c.Size); err != nil {
		return err
	}
	for _, move := range moves {
		c.applyMove(move)
	}
	return nil
}

// moveToMoveType converts a Move struct to MoveType and determines quarter turns
func moveToMoveType(move Move) (MoveType, int) {
	var moveType MoveType
	var quarterTurns int

	// Handle slice moves
	if move.Slice != NoSlice {
		switch move.Slice {
		case M_Slice:
			moveType = MoveM
		case E_Slice:
			moveType = MoveE
		case S_Slice:
			moveType = MoveS
		default:
			return MoveR, 0 // Default fallback
		}
	} else if move.Rotation != NoRotation {
		// Handle cube rotations
		switch move.Rotation {
		case X_Rotation:
			moveType = MoveX
		case Y_Rotation:
			moveType = MoveY
		case Z_Rotation:
			moveType = MoveZ
		default:
			return MoveR, 0 // Default fallback
		}
	} else {
		// Handle face moves
		switch move.Face {
		case Right:
			moveType = MoveR
		case Left:
			moveType = MoveL
		case Up:
			moveType = MoveU
		case Down:
			moveType = MoveD
		case Front:
			moveType = MoveF
		case Back:
			moveType = MoveB
		default:
			return MoveR, 0 // Default fallback
		}
	}

	// Determine quarter turns
	if move.Double {
		quarterTurns = 2
	} else if move.Clockwise {
		quarterTurns = 1 // Clockwise = 1 quarter turn
	} else {
		quarterTurns = 3 // Counter-clockwise = 3 quarter turns clockwise
	}

	return moveType, quarterTurns
}

// getAffectedLayers determines which layers are affected by a move
func getAffectedLayers(move Move, N int) []int {
	// Handle slice moves
	if move.Slice != NoSlice {
		if N%2 == 0 {
			return []int{} // Slice moves undefined for even cubes
		}
		return []int{N / 2} // Middle layer
	}

	// A cube-rotation permutation already includes every layer. Apply it once,
	// rather than once per layer (which changes the angle with cube size).
	if move.Rotation != NoRotation {
		return []int{0}
	}

	// Handle face moves
	if move.Wide {
		// Wide moves affect outer N layers (default 2)
		depth := move.WideDepth
		if depth <= 0 {
			depth = 2
		}
		layers := make([]int, depth)
		for i := 0; i < depth; i++ {
			layers[i] = i
		}
		return layers
	} else if move.Layer > 0 {
		// Layer moves (2R, 3L, etc.) affect only the specified layer
		return []int{move.Layer}
	} else {
		// Regular moves affect only outer layer (standard cubing convention)
		return []int{0}
	}
}
