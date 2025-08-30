package cube

import (
	"strings"
)

// Face represents a face of the cube
type Face int

const (
	Front Face = iota
	Back
	Left
	Right
	Up
	Down
)

func (f Face) String() string {
	return []string{"F", "B", "L", "R", "U", "D"}[f]
}

// Color represents a sticker color
type Color int

const (
	White Color = iota
	Yellow
	Red
	Orange
	Blue
	Green
	Grey // Wildcard for pattern matching
)

func (c Color) String() string {
	return []string{"W", "Y", "R", "O", "B", "G", "."}[c]
}

// ColoredString returns a muted colored string representation
func (c Color) ColoredString() string {
	// Much more muted colors that won't burn eyes
	colors := []string{
		"\033[37mW\033[0m", // Light gray for white
		"\033[33mY\033[0m", // Muted yellow
		"\033[31mR\033[0m", // Muted red
		"\033[35mO\033[0m", // Muted magenta for orange
		"\033[34mB\033[0m", // Muted blue
		"\033[32mG\033[0m", // Muted green
		"\033[90m.\033[0m", // Dark gray for wildcard
	}
	return colors[c]
}

// UnicodeString returns a colored Unicode square representation
func (c Color) UnicodeString() string {
	squares := []string{"⬜", "🟨", "🟥", "🟧", "🟦", "🟩", "⬛"}
	return squares[c]
}

// Cube represents an NxNxN cube
type Cube struct {
	Size  int          // Dimension of the cube (3 for 3x3x3)
	Faces [6][][]Color // Six faces, each Size x Size
}

// NewCube creates a new solved cube of the given size
func NewCube(size int) *Cube {
	if size < 2 {
		size = 2 // Minimum 2x2x2
	}

	cube := &Cube{Size: size}

	// Initialize faces with solved colors
	// Canonical orientation: Yellow on top, White on bottom, Blue facing front
	// Face order: Front, Back, Left, Right, Up, Down
	faceColors := []Color{Blue, Green, Orange, Red, Yellow, White}

	for face := 0; face < 6; face++ {
		cube.Faces[face] = make([][]Color, size)
		for row := 0; row < size; row++ {
			cube.Faces[face][row] = make([]Color, size)
			for col := 0; col < size; col++ {
				cube.Faces[face][row][col] = faceColors[face]
			}
		}
	}

	return cube
}

// IsSolved checks if the cube is in a solved state
func (c *Cube) IsSolved() bool {
	for face := 0; face < 6; face++ {
		firstColor := c.Faces[face][0][0]
		for row := 0; row < c.Size; row++ {
			for col := 0; col < c.Size; col++ {
				if c.Faces[face][row][col] != firstColor {
					return false
				}
			}
		}
	}
	return true
}

// String returns a string representation of the cube
func (c *Cube) String() string {
	return c.StringWithColor(false)
}

// StringWithColor returns a string representation with optional colors
func (c *Cube) StringWithColor(useColor bool) string {
	return c.UnfoldedString(useColor, false)
}

// UnfoldedString returns the cube in an unfolded cross layout
func (c *Cube) UnfoldedString(useColor bool, useUnicode bool) string {
	var sb strings.Builder

	// Create padding to align top/bottom with front face
	var leftPadding string
	if useUnicode {
		// Unicode blocks are double-width: (c.Size * 2) + 1 space
		leftPadding = strings.Repeat(" ", (c.Size*2)+1)
	} else {
		// Single-width characters: c.Size + 1 space
		leftPadding = strings.Repeat(" ", c.Size) + " "
	}

	// Top face (Up) - aligned with Front face
	for row := 0; row < c.Size; row++ {
		sb.WriteString(leftPadding)
		for col := 0; col < c.Size; col++ {
			sb.WriteString(c.FormatSticker(c.Faces[Up][row][col], useColor, useUnicode))
		}
		sb.WriteString("\n")
	}
	sb.WriteString("\n")

	// Middle row: Left, Front, Right, Back
	for row := 0; row < c.Size; row++ {
		// Left face
		for col := 0; col < c.Size; col++ {
			sb.WriteString(c.FormatSticker(c.Faces[Left][row][col], useColor, useUnicode))
		}
		sb.WriteString(" ") // Space between faces

		// Front face
		for col := 0; col < c.Size; col++ {
			sb.WriteString(c.FormatSticker(c.Faces[Front][row][col], useColor, useUnicode))
		}
		sb.WriteString(" ") // Space between faces

		// Right face
		for col := 0; col < c.Size; col++ {
			sb.WriteString(c.FormatSticker(c.Faces[Right][row][col], useColor, useUnicode))
		}
		sb.WriteString(" ") // Space between faces

		// Back face
		for col := 0; col < c.Size; col++ {
			sb.WriteString(c.FormatSticker(c.Faces[Back][row][col], useColor, useUnicode))
		}
		sb.WriteString("\n")
	}
	sb.WriteString("\n")

	// Bottom face (Down) - aligned with Front face
	for row := 0; row < c.Size; row++ {
		sb.WriteString(leftPadding)
		for col := 0; col < c.Size; col++ {
			sb.WriteString(c.FormatSticker(c.Faces[Down][row][col], useColor, useUnicode))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// FormatSticker returns the appropriate representation for a sticker
func (c *Cube) FormatSticker(color Color, useColor bool, useUnicode bool) string {
	if useUnicode {
		return color.UnicodeString()
	} else if useColor {
		return color.ColoredString()
	} else {
		return color.String()
	}
}

// MonoString returns a single-width colored block for perfect monospace alignment
func (c Color) MonoString() string {
	// Use colored █ (full block) which is guaranteed single-width
	colors := []string{
		"\033[47m█\033[0m", // White background
		"\033[43m█\033[0m", // Yellow background
		"\033[41m█\033[0m", // Red background
		"\033[45m█\033[0m", // Magenta background (orange)
		"\033[44m█\033[0m", // Blue background
		"\033[42m█\033[0m", // Green background
	}
	return colors[c]
}

// AffectsOnlyLastLayers detects if changes are limited to the top two layers
func AffectsOnlyLastLayers(beforeCube, afterCube *Cube) bool {
	if beforeCube.Size != afterCube.Size {
		return false
	}
	
	size := beforeCube.Size
	
	// For each face, check if only the top two layers are affected
	for face := 0; face < 6; face++ {
		switch Face(face) {
		case Up:
			// Up face: all changes are considered last-layer
			continue
		case Down:
			// Down face: no changes allowed for last-layer algorithms
			for row := 0; row < size; row++ {
				for col := 0; col < size; col++ {
					if beforeCube.Faces[face][row][col] != afterCube.Faces[face][row][col] {
						return false
					}
				}
			}
		case Front, Back, Left, Right:
			// Side faces: only top two layers (rows 0 and 1) can change
			for row := 2; row < size; row++ {
				for col := 0; col < size; col++ {
					if beforeCube.Faces[face][row][col] != afterCube.Faces[face][row][col] {
						return false
					}
				}
			}
		}
	}
	
	return true
}

// LastLayerString returns a 5x5 grid view focused on the last layer
func (c *Cube) LastLayerString(useColor bool, useUnicode bool) string {
	var sb strings.Builder
	
	if c.Size < 3 {
		// For 2x2, just show the unfolded view
		return c.UnfoldedString(useColor, useUnicode)
	}
	
	// Layout for 3x3:
	//     [back edge row]
	// [left] [top face] [right] 
	//     [front edge row]
	
	// Top row: back edge pieces (Back face, row 0)
	sb.WriteString("  ")
	for col := 0; col < c.Size; col++ {
		sb.WriteString(c.FormatSticker(c.Faces[Back][0][col], useColor, useUnicode))
	}
	sb.WriteString("\n")
	
	// Middle rows: left edges, top face, right edges
	for row := 0; row < c.Size; row++ {
		// Left edge (Left face, column size-1)
		sb.WriteString(c.FormatSticker(c.Faces[Left][row][c.Size-1], useColor, useUnicode))
		sb.WriteString(" ")
		
		// Top face row
		for col := 0; col < c.Size; col++ {
			sb.WriteString(c.FormatSticker(c.Faces[Up][row][col], useColor, useUnicode))
		}
		sb.WriteString(" ")
		
		// Right edge (Right face, column 0)
		sb.WriteString(c.FormatSticker(c.Faces[Right][row][0], useColor, useUnicode))
		sb.WriteString("\n")
	}
	
	// Bottom row: front edge pieces (Front face, row 0)
	sb.WriteString("  ")
	for col := 0; col < c.Size; col++ {
		sb.WriteString(c.FormatSticker(c.Faces[Front][0][col], useColor, useUnicode))
	}
	sb.WriteString("\n")
	
	return sb.String()
}
