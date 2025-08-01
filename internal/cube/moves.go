package cube

// ApplyMove applies a single move to the cube
func (c *Cube) ApplyMove(move Move) {
	moveType, quarterTurns := moveToMoveType(move)
	layers := getAffectedLayers(move, c.Size)

	for _, layer := range layers {
		perm := getPermutation(c.Size, moveType, layer, quarterTurns)
		applyPermutation(c, perm)
	}
}

// ApplyMoves applies a sequence of moves to the cube
func (c *Cube) ApplyMoves(moves []Move) {
	for _, move := range moves {
		c.ApplyMove(move)
	}
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

	// Handle cube rotations (affect all layers)
	if move.Rotation != NoRotation {
		layers := make([]int, N)
		for i := 0; i < N; i++ {
			layers[i] = i
		}
		return layers
	}

	// Handle face moves
	if move.Wide {
		// Wide moves affect outer two layers by default
		depth := move.WideDepth
		if depth <= 0 {
			depth = 2
		}
		layers := make([]int, depth)
		for i := 0; i < depth; i++ {
			layers[i] = i
		}
		return layers
	} else {
		// Regular moves affect only outer layer
		return []int{0}
	}
}
