package cli

import (
	"fmt"
	"strings"

	"github.com/ehrlich-b/cube/internal/cube"
	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:   "show [scramble]",
	Short: "Show cube state with optional pattern highlighting",
	Long: `Show displays the cube state after applying a scramble.
It can highlight specific patterns to help with learning algorithms.

Examples:
  cube show "R U R' U'"
  cube show "R U R' U'" --color
  cube show "R U R' U'" --highlight-cross
  cube show "" --highlight-oll`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		scramble := ""
		if len(args) > 0 {
			scramble = args[0]
		}

		dimension, _ := cmd.Flags().GetInt("dimension")
		useColor, _ := cmd.Flags().GetBool("color")
		useLetters, _ := cmd.Flags().GetBool("letters")
		useUnicode := useColor && !useLetters
		highlightCross, _ := cmd.Flags().GetBool("highlight-cross")
		highlightOLL, _ := cmd.Flags().GetBool("highlight-oll")
		highlightPLL, _ := cmd.Flags().GetBool("highlight-pll")
		highlightF2L, _ := cmd.Flags().GetBool("highlight-f2l")

		// Create cube
		c := cube.NewCube(dimension)

		// Apply scramble if provided
		if scramble != "" {
			moves, err := cube.ParseScramble(scramble)
			if err != nil {
				fmt.Printf("Error parsing scramble: %v\n", err)
				return
			}
			c.ApplyMoves(moves)
			fmt.Printf("Cube state after scramble: %s\n\n", scramble)
		} else {
			fmt.Println("Solved cube state:")
		}

		// Determine highlight mode
		highlightMode := ""
		if highlightCross {
			highlightMode = "cross"
		} else if highlightOLL {
			highlightMode = "oll"
		} else if highlightPLL {
			highlightMode = "pll"
		} else if highlightF2L {
			highlightMode = "f2l"
		}

		// Display cube with highlighting
		if highlightMode != "" {
			fmt.Printf("Highlighting: %s pattern\n\n", strings.ToUpper(highlightMode))
			displayWithHighlight(c, highlightMode, useColor, useUnicode)
		} else {
			fmt.Println(c.UnfoldedString(useColor, useUnicode))
		}
	},
}

func displayWithHighlight(c *cube.Cube, mode string, useColor bool, useUnicode bool) {
	// Display cube in unfolded cross format with highlighting
	var sb strings.Builder

	dimColor := "\033[90m" // Dark gray for dimmed pieces
	resetColor := "\033[0m"
	dimUnicode := "⬛" // Black square for dimmed pieces

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
			color := c.Faces[4][row][col] // Up face
			highlight := shouldHighlight(4, row, col, c.Size, mode)

			if highlight {
				sb.WriteString(c.FormatSticker(color, useColor, useUnicode))
			} else {
				if useUnicode {
					sb.WriteString(dimUnicode)
				} else if useColor {
					sb.WriteString(fmt.Sprintf("%s%s%s", dimColor, color.String(), resetColor))
				} else {
					sb.WriteString(".")
				}
			}
		}
		sb.WriteString("\n")
	}
	sb.WriteString("\n")

	// Middle row: Left, Front, Right, Back
	faces := [4]int{2, 0, 3, 1} // Left, Front, Right, Back
	for row := 0; row < c.Size; row++ {
		for i, face := range faces {
			for col := 0; col < c.Size; col++ {
				color := c.Faces[face][row][col]
				highlight := shouldHighlight(face, row, col, c.Size, mode)

				if highlight {
					sb.WriteString(c.FormatSticker(color, useColor, useUnicode))
				} else {
					if useUnicode {
						sb.WriteString(dimUnicode)
					} else if useColor {
						sb.WriteString(fmt.Sprintf("%s%s%s", dimColor, color.String(), resetColor))
					} else {
						sb.WriteString(".")
					}
				}
			}
			if i < 3 { // Add space between faces (but not after the last one)
				sb.WriteString(" ")
			}
		}
		sb.WriteString("\n")
	}
	sb.WriteString("\n")

	// Bottom face (Down) - aligned with Front face
	for row := 0; row < c.Size; row++ {
		sb.WriteString(leftPadding)
		for col := 0; col < c.Size; col++ {
			color := c.Faces[5][row][col] // Down face
			highlight := shouldHighlight(5, row, col, c.Size, mode)

			if highlight {
				sb.WriteString(c.FormatSticker(color, useColor, useUnicode))
			} else {
				if useUnicode {
					sb.WriteString(dimUnicode)
				} else if useColor {
					sb.WriteString(fmt.Sprintf("%s%s%s", dimColor, color.String(), resetColor))
				} else {
					sb.WriteString(".")
				}
			}
		}
		sb.WriteString("\n")
	}

	fmt.Print(sb.String())
}

func shouldHighlight(face, row, col, size int, mode string) bool {
	// Simple highlighting logic - can be made much more sophisticated
	switch mode {
	case "cross":
		// Highlight white cross on bottom (Down face)
		if face == 5 { // Down face
			// Center and edge pieces
			center := size / 2
			if (row == center && col == center) ||
				(row == 0 && col == center) ||
				(row == size-1 && col == center) ||
				(row == center && col == 0) ||
				(row == center && col == size-1) {
				return true
			}
		}
		// Also highlight corresponding edge pieces on adjacent faces
		if face >= 0 && face <= 3 { // Front, Back, Left, Right
			if row == size-1 && col == size/2 {
				return true
			}
		}

	case "oll":
		// Highlight top layer (Up face) for OLL
		if face == 4 { // Up face
			return true
		}
		// Also highlight top row of side faces
		if face >= 0 && face <= 3 { // Front, Back, Left, Right
			if row == 0 {
				return true
			}
		}

	case "pll":
		// Highlight top layer for PLL (same as OLL for now)
		if face == 4 { // Up face
			return true
		}
		if face >= 0 && face <= 3 { // Front, Back, Left, Right
			if row == 0 {
				return true
			}
		}

	case "f2l":
		// Highlight first two layers (bottom 2/3 of side faces)
		if face == 5 { // Down face
			return true
		}
		if face >= 0 && face <= 3 { // Front, Back, Left, Right
			if row >= size/3 {
				return true
			}
		}
	}

	return false
}

func init() {
	showCmd.Flags().IntP("dimension", "d", 3, "Cube dimension (2, 3, 4, etc.)")
	showCmd.Flags().BoolP("color", "c", false, "Use colored output (Unicode blocks by default)")
	showCmd.Flags().Bool("letters", false, "Use letters instead of Unicode blocks when using --color")
	showCmd.Flags().Bool("highlight-cross", false, "Highlight cross pattern")
	showCmd.Flags().Bool("highlight-oll", false, "Highlight OLL (Orientation of Last Layer)")
	showCmd.Flags().Bool("highlight-pll", false, "Highlight PLL (Permutation of Last Layer)")
	showCmd.Flags().Bool("highlight-f2l", false, "Highlight F2L (First Two Layers)")
}
