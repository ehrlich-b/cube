package cfen

import (
	"fmt"

	"github.com/ehrlich-b/cube/internal/cube"
)

// ToCube converts a CFENState to an internal Cube representation
func (state *CFENState) ToCube() (*cube.Cube, error) {
	rotations, err := orientationMoves(state.Orientation)
	if err != nil {
		return nil, err
	}
	// Create new cube with correct dimension
	newCube := cube.NewCube(state.Dimension)

	// CFEN stores the faces of its oriented frame in U/R/F/D/L/B order.
	faceMapping := [...]cube.Face{cube.Up, cube.Right, cube.Front, cube.Down, cube.Left, cube.Back}

	// Copy the grids before rotating them back into canonical coordinates.
	for cfenFaceIdx, cfenFace := range state.Faces {
		internalFace := faceMapping[cfenFaceIdx]

		// Convert flattened sticker array to 2D array
		for stickerIdx, color := range cfenFace.Stickers {
			row := stickerIdx / state.Dimension
			col := stickerIdx % state.Dimension
			newCube.Faces[internalFace][row][col] = color
		}
	}

	// Restore the canonical coordinate frame, including each face's grid angle.
	for i := len(rotations) - 1; i >= 0; i-- {
		move := rotations[i]
		move.Clockwise = !move.Clockwise
		newCube.ApplyMove(move)
	}
	return newCube, nil
}

// FromCube converts an internal Cube to CFENState
func FromCube(c *cube.Cube, orientation CFENOrientation) (*CFENState, error) {
	if c == nil {
		return nil, fmt.Errorf("cube cannot be nil")
	}

	rotations, err := orientationMoves(orientation)
	if err != nil {
		return nil, err
	}
	oriented := c
	if len(rotations) > 0 {
		// Work on a copy; exporting an orientation must not mutate the source.
		oriented = cube.NewCube(c.Size)
		for face := range c.Faces {
			for row := range c.Faces[face] {
				copy(oriented.Faces[face][row], c.Faces[face][row])
			}
		}
		if err := oriented.ApplyMoves(rotations); err != nil {
			return nil, err
		}
	}
	reverseFaceMapping := [...]cube.Face{cube.Up, cube.Right, cube.Front, cube.Down, cube.Left, cube.Back}

	var faces [6]CFENFace

	for cfenFaceIdx := 0; cfenFaceIdx < 6; cfenFaceIdx++ {
		internalFace := reverseFaceMapping[cfenFaceIdx]

		// Convert 2D array to flattened sticker array
		stickers := make([]cube.Color, c.Size*c.Size)
		for row := 0; row < c.Size; row++ {
			for col := 0; col < c.Size; col++ {
				stickerIdx := row*c.Size + col
				stickers[stickerIdx] = oriented.Faces[internalFace][row][col]
			}
		}

		faces[cfenFaceIdx] = CFENFace{
			Stickers: stickers,
			Size:     c.Size,
		}
	}

	return &CFENState{
		Orientation: orientation,
		Faces:       faces,
		Dimension:   c.Size,
	}, nil
}

// GenerateCFEN creates a CFEN string from a cube with default orientation
func GenerateCFEN(c *cube.Cube) (string, error) {
	// Use default orientation matching cube's canonical orientation (Yellow up, Blue front)
	orientation := CFENOrientation{
		Up:    cube.Yellow,
		Front: cube.Blue,
	}

	cfenState, err := FromCube(c, orientation)
	if err != nil {
		return "", err
	}

	return cfenState.String(), nil
}

// MatchesPattern checks if the cube state matches a CFEN pattern with wildcards
func (state *CFENState) MatchesCube(c *cube.Cube) (bool, error) {
	if c.Size != state.Dimension {
		return false, fmt.Errorf("cube dimension %d doesn't match CFEN dimension %d", c.Size, state.Dimension)
	}

	// Convert cube to CFEN for comparison
	cubeState, err := FromCube(c, state.Orientation)
	if err != nil {
		return false, err
	}

	// Compare each face, ignoring wildcards (Grey color)
	for faceIdx := 0; faceIdx < 6; faceIdx++ {
		patternFace := state.Faces[faceIdx]
		cubeFace := cubeState.Faces[faceIdx]

		if len(patternFace.Stickers) != len(cubeFace.Stickers) {
			return false, fmt.Errorf("face %d sticker count mismatch", faceIdx)
		}

		for stickerIdx := 0; stickerIdx < len(patternFace.Stickers); stickerIdx++ {
			patternColor := patternFace.Stickers[stickerIdx]
			cubeColor := cubeFace.Stickers[stickerIdx]

			// Skip wildcard positions (Grey color)
			if patternColor == cube.Grey {
				continue
			}

			// Exact match required for non-wildcard positions
			if patternColor != cubeColor {
				return false, nil
			}
		}
	}

	return true, nil
}

// ValidateCFEN validates a CFEN string format and returns any errors
func ValidateCFEN(cfenStr string) error {
	_, err := ParseCFEN(cfenStr)
	return err
}

// Each valid Up/Front color pair identifies one of the 24 rigid orientations.
// Derive their transforms from physical rotations rather than face-only maps,
// which can mirror centers or lose the rotation of the sticker grids.
var orientationRotations = func() map[CFENOrientation][]cube.Move {
	result := make(map[CFENOrientation][]cube.Move)
	type frame struct {
		c     *cube.Cube
		moves []cube.Move
	}
	queue := []frame{{cube.NewCube(3), nil}}
	axes, _ := cube.ParseMoves("x y z")
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		key := CFENOrientation{current.c.Faces[cube.Up][1][1], current.c.Faces[cube.Front][1][1]}
		if _, ok := result[key]; ok {
			continue
		}
		result[key] = current.moves
		for _, move := range axes {
			next := cube.NewCube(3)
			for f := range next.Faces {
				for r := range next.Faces[f] {
					copy(next.Faces[f][r], current.c.Faces[f][r])
				}
			}
			next.ApplyMove(move)
			moves := append(append([]cube.Move(nil), current.moves...), move)
			queue = append(queue, frame{next, moves})
		}
	}
	return result
}()

func orientationMoves(orientation CFENOrientation) ([]cube.Move, error) {
	moves, ok := orientationRotations[orientation]
	if !ok {
		return nil, fmt.Errorf("orientation requires adjacent real Up and Front colors")
	}
	return moves, nil
}
