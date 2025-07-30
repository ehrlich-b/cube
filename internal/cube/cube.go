package cube

import (
	"fmt"
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
)

func (c Color) String() string {
	return []string{"W", "Y", "R", "O", "B", "G"}[c]
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
	}
	return colors[c]
}

// UnicodeString returns a colored Unicode square representation
func (c Color) UnicodeString() string {
	squares := []string{"⬜", "🟨", "🟥", "🟧", "🟦", "🟩"}
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
	faceColors := []Color{White, Yellow, Red, Orange, Blue, Green}

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
	var sb strings.Builder

	faceNames := []string{"Front", "Back", "Left", "Right", "Up", "Down"}

	for face := 0; face < 6; face++ {
		sb.WriteString(fmt.Sprintf("%s face:\n", faceNames[face]))
		for row := 0; row < c.Size; row++ {
			for col := 0; col < c.Size; col++ {
				if useColor {
					sb.WriteString(c.Faces[face][row][col].ColoredString())
				} else {
					sb.WriteString(c.Faces[face][row][col].String())
				}
				sb.WriteString(" ")
			}
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
