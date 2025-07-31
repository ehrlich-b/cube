package cfen

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/ehrlich-b/cube/internal/cube"
)

// CFENOrientation represents the Up and Front face colors
type CFENOrientation struct {
	Up    cube.Color
	Front cube.Color
}

// CFENFace represents a face with stickers and wildcard support
type CFENFace struct {
	Stickers []cube.Color // Flattened array of stickers (row-major order)
	Size     int          // Dimension (N for NxN face)
}

// CFENState represents a complete cube state in CFEN format
type CFENState struct {
	Orientation CFENOrientation
	Faces       [6]CFENFace // U, R, F, D, L, B order
	Dimension   int         // Cube dimension (N for NxN cube)
}

// String returns the CFEN string representation
func (state *CFENState) String() string {
	var sb strings.Builder

	// Orientation field
	sb.WriteString(state.Orientation.Up.String())
	sb.WriteString(state.Orientation.Front.String())
	sb.WriteString("|")

	// Faces field (U/R/F/D/L/B order)
	for i, face := range state.Faces {
		if i > 0 {
			sb.WriteString("/")
		}
		sb.WriteString(face.compactString())
	}

	return sb.String()
}

// compactString returns run-length encoded representation of face stickers
func (face *CFENFace) compactString() string {
	if len(face.Stickers) == 0 {
		return ""
	}

	var sb strings.Builder
	currentColor := face.Stickers[0]
	count := 1

	for i := 1; i < len(face.Stickers); i++ {
		if face.Stickers[i] == currentColor {
			count++
		} else {
			// Write current run
			sb.WriteString(currentColor.String())
			if count > 1 {
				sb.WriteString(strconv.Itoa(count))
			}

			// Start new run
			currentColor = face.Stickers[i]
			count = 1
		}
	}

	// Write final run
	sb.WriteString(currentColor.String())
	if count > 1 {
		sb.WriteString(strconv.Itoa(count))
	}

	return sb.String()
}

// ParseCFEN parses a CFEN string into a CFENState
func ParseCFEN(cfenStr string) (*CFENState, error) {
	// Split on | to separate orientation and faces
	parts := strings.Split(cfenStr, "|")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid CFEN format: expected 'orientation|faces', got '%s'", cfenStr)
	}

	// Parse orientation
	orientation, err := parseOrientation(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid orientation '%s': %v", parts[0], err)
	}

	// Parse faces
	faces, dimension, err := parseFaces(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid faces '%s': %v", parts[1], err)
	}

	return &CFENState{
		Orientation: *orientation,
		Faces:       faces,
		Dimension:   dimension,
	}, nil
}

// parseOrientation parses the orientation field (e.g., "WG")
func parseOrientation(orientStr string) (*CFENOrientation, error) {
	if len(orientStr) != 2 {
		return nil, fmt.Errorf("orientation must be exactly 2 characters, got %d", len(orientStr))
	}

	upColor, err := parseColor(rune(orientStr[0]))
	if err != nil {
		return nil, fmt.Errorf("invalid up color '%c': %v", orientStr[0], err)
	}

	frontColor, err := parseColor(rune(orientStr[1]))
	if err != nil {
		return nil, fmt.Errorf("invalid front color '%c': %v", orientStr[1], err)
	}

	return &CFENOrientation{
		Up:    upColor,
		Front: frontColor,
	}, nil
}

// parseFaces parses the faces field (e.g., "W9/R9/G9/Y9/O9/B9")
func parseFaces(facesStr string) ([6]CFENFace, int, error) {
	faceStrs := strings.Split(facesStr, "/")
	if len(faceStrs) != 6 {
		return [6]CFENFace{}, 0, fmt.Errorf("expected 6 faces separated by '/', got %d", len(faceStrs))
	}

	var faces [6]CFENFace
	var dimension int

	for i, faceStr := range faceStrs {
		face, err := parseFace(faceStr)
		if err != nil {
			return [6]CFENFace{}, 0, fmt.Errorf("face %d: %v", i, err)
		}

		// Validate face size consistency
		if i == 0 {
			// Determine dimension from first face
			stickers := len(face.Stickers)
			dim := int(sqrt(float64(stickers)))
			if dim*dim != stickers {
				return [6]CFENFace{}, 0, fmt.Errorf("face %d has %d stickers, not a perfect square", i, stickers)
			}
			dimension = dim
		} else {
			// Verify all faces have same size
			if len(face.Stickers) != dimension*dimension {
				return [6]CFENFace{}, 0, fmt.Errorf("face %d has %d stickers, expected %d", i, len(face.Stickers), dimension*dimension)
			}
		}

		face.Size = dimension
		faces[i] = *face
	}

	return faces, dimension, nil
}

// parseFace parses a single face string with run-length encoding
func parseFace(faceStr string) (*CFENFace, error) {
	var stickers []cube.Color

	// Regular expression to match color+optional_count patterns
	re := regexp.MustCompile(`([WYROGB?])(\d*)`)
	matches := re.FindAllStringSubmatch(faceStr, -1)

	if len(matches) == 0 {
		return nil, fmt.Errorf("no valid color tokens found in '%s'", faceStr)
	}

	for _, match := range matches {
		colorChar := match[1]
		countStr := match[2]

		// Parse color
		color, err := parseColor(rune(colorChar[0]))
		if err != nil {
			return nil, fmt.Errorf("invalid color '%s': %v", colorChar, err)
		}

		// Parse count (default 1)
		count := 1
		if countStr != "" {
			var err error
			count, err = strconv.Atoi(countStr)
			if err != nil || count < 1 {
				return nil, fmt.Errorf("invalid count '%s': must be positive integer", countStr)
			}
		}

		// Add stickers
		for i := 0; i < count; i++ {
			stickers = append(stickers, color)
		}
	}

	// Verify we consumed the entire string
	reconstructed := ""
	for _, match := range matches {
		reconstructed += match[0]
	}
	if reconstructed != faceStr {
		return nil, fmt.Errorf("failed to parse entire face string '%s', parsed '%s'", faceStr, reconstructed)
	}

	return &CFENFace{
		Stickers: stickers,
	}, nil
}

// parseColor converts a character to a Color
func parseColor(ch rune) (cube.Color, error) {
	switch ch {
	case 'W':
		return cube.White, nil
	case 'Y':
		return cube.Yellow, nil
	case 'R':
		return cube.Red, nil
	case 'O':
		return cube.Orange, nil
	case 'G':
		return cube.Green, nil
	case 'B':
		return cube.Blue, nil
	case '?':
		return cube.Grey, nil // Wildcard
	default:
		return cube.White, fmt.Errorf("unknown color character '%c'", ch)
	}
}

// sqrt returns the integer square root (simple implementation)
func sqrt(x float64) float64 {
	if x == 0 {
		return 0
	}

	// Newton's method for integer square root
	result := x
	for {
		next := 0.5 * (result + x/result)
		if abs(next-result) < 0.000001 {
			break
		}
		result = next
	}
	return result
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
