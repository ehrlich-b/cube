package cfen

import (
	"fmt"

	"github.com/ehrlich-b/cube/internal/cube"
)

// ToCube converts a CFENState to an internal Cube representation
func (state *CFENState) ToCube() (*cube.Cube, error) {
	// Create new cube with correct dimension
	newCube := cube.NewCube(state.Dimension)

	// Get face mapping based on CFEN orientation
	faceMapping := getOrientationMapping(state.Orientation)

	// Copy stickers using orientation-aware mapping
	for cfenFaceIdx, cfenFace := range state.Faces {
		internalFace := faceMapping[cfenFaceIdx]

		// Convert flattened sticker array to 2D array
		for stickerIdx, color := range cfenFace.Stickers {
			row := stickerIdx / state.Dimension
			col := stickerIdx % state.Dimension
			newCube.Faces[internalFace][row][col] = color
		}
	}

	return newCube, nil
}

// FromCube converts an internal Cube to CFENState
func FromCube(c *cube.Cube, orientation CFENOrientation) (*CFENState, error) {
	if c == nil {
		return nil, fmt.Errorf("cube cannot be nil")
	}

	// Get reverse face mapping based on desired CFEN orientation
	reverseFaceMapping := getReverseOrientationMapping(orientation)

	var faces [6]CFENFace

	for cfenFaceIdx := 0; cfenFaceIdx < 6; cfenFaceIdx++ {
		internalFace := reverseFaceMapping[cfenFaceIdx]

		// Convert 2D array to flattened sticker array
		stickers := make([]cube.Color, c.Size*c.Size)
		for row := 0; row < c.Size; row++ {
			for col := 0; col < c.Size; col++ {
				stickerIdx := row*c.Size + col
				stickers[stickerIdx] = c.Faces[internalFace][row][col]
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

// getOrientationMapping returns face mapping from CFEN faces to internal cube faces
func getOrientationMapping(orientation CFENOrientation) [6]cube.Face {
	// Cube canonical: Yellow=Up, Blue=Front, Red=Right, White=Down, Orange=Left, Green=Back
	// CFEN order: U/R/F/D/L/B

	// Standard YB orientation (Yellow up, Blue front) - matches cube canonical
	if orientation.Up == cube.Yellow && orientation.Front == cube.Blue {
		return [6]cube.Face{cube.Up, cube.Right, cube.Front, cube.Down, cube.Left, cube.Back}
	}

	// WG orientation (White up, Green front) - cube rotated x' z
	if orientation.Up == cube.White && orientation.Front == cube.Green {
		return [6]cube.Face{cube.Down, cube.Left, cube.Back, cube.Up, cube.Right, cube.Front}
	}

	// WB orientation (White up, Blue front) - cube rotated x'
	if orientation.Up == cube.White && orientation.Front == cube.Blue {
		return [6]cube.Face{cube.Down, cube.Right, cube.Front, cube.Up, cube.Left, cube.Back}
	}

	// YG orientation (Yellow up, Green front) - cube rotated z
	if orientation.Up == cube.Yellow && orientation.Front == cube.Green {
		return [6]cube.Face{cube.Up, cube.Left, cube.Back, cube.Down, cube.Right, cube.Front}
	}

	// Default fallback to YB
	return [6]cube.Face{cube.Up, cube.Right, cube.Front, cube.Down, cube.Left, cube.Back}
}

// getReverseOrientationMapping returns face mapping from internal cube faces to CFEN faces
func getReverseOrientationMapping(orientation CFENOrientation) [6]cube.Face {
	// Standard YB orientation
	if orientation.Up == cube.Yellow && orientation.Front == cube.Blue {
		return [6]cube.Face{cube.Up, cube.Right, cube.Front, cube.Down, cube.Left, cube.Back}
	}

	// WG orientation - reverse of getOrientationMapping
	if orientation.Up == cube.White && orientation.Front == cube.Green {
		return [6]cube.Face{cube.Back, cube.Left, cube.Down, cube.Front, cube.Right, cube.Up}
	}

	// WB orientation
	if orientation.Up == cube.White && orientation.Front == cube.Blue {
		return [6]cube.Face{cube.Down, cube.Right, cube.Front, cube.Up, cube.Left, cube.Back}
	}

	// YG orientation
	if orientation.Up == cube.Yellow && orientation.Front == cube.Green {
		return [6]cube.Face{cube.Up, cube.Left, cube.Back, cube.Down, cube.Right, cube.Front}
	}

	// Default fallback
	return [6]cube.Face{cube.Up, cube.Right, cube.Front, cube.Down, cube.Left, cube.Back}
}
