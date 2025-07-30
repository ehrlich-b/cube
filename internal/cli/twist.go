package cli

import (
	"fmt"
	"os"

	"github.com/ehrlich-b/cube/internal/cube"
	"github.com/spf13/cobra"
)

var twistCmd = &cobra.Command{
	Use:   "twist [moves]",
	Short: "Apply moves to a cube and display the result",
	Long: `Apply a sequence of moves to a cube and display the resulting state.
This command does not solve the cube - it just applies the moves and shows
the result. Perfect for learning algorithms, exploring patterns, and visualization.

Examples:
  cube twist "R U R' U'"
  cube twist "F R U' R' F'" --color
  cube twist "Rw Uw Fw" --dimension 4`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		moves := args[0]
		dimension, _ := cmd.Flags().GetInt("dimension")

		fmt.Printf("Applying moves to %dx%dx%d cube: %s\n", dimension, dimension, dimension, moves)

		// Create cube
		c := cube.NewCube(dimension)

		// Parse and apply moves
		parsedMoves, err := cube.ParseScramble(moves)
		if err != nil {
			fmt.Printf("Error parsing moves: %v\n", err)
			os.Exit(1)
		}

		c.ApplyMoves(parsedMoves)

		// Get display options
		useColor, _ := cmd.Flags().GetBool("color")
		useLetters, _ := cmd.Flags().GetBool("letters")
		useUnicode := useColor && !useLetters

		// Display result
		fmt.Printf("\nCube state after applying moves:\n%s\n", c.UnfoldedString(useColor, useUnicode))

		// Show move count
		fmt.Printf("Moves applied: %d\n", len(parsedMoves))

		// Check if solved
		if c.IsSolved() {
			fmt.Printf("Status: ✅ SOLVED!\n")
		} else {
			fmt.Printf("Status: 🔄 Scrambled\n")
		}
	},
}

func init() {
	twistCmd.Flags().IntP("dimension", "d", 3, "Cube dimension (2, 3, 4, etc.)")
	twistCmd.Flags().BoolP("color", "c", false, "Use colored output (Unicode blocks by default)")
	twistCmd.Flags().Bool("letters", false, "Use letters instead of Unicode blocks when using --color")
}
