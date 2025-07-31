package cli

import (
	"fmt"
	"os"

	"github.com/ehrlich-b/cube/internal/cfen"
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
		useCfenOutput, _ := cmd.Flags().GetBool("cfen")
		startCfen, _ := cmd.Flags().GetString("start")

		// Create cube from starting position
		var c *cube.Cube
		if startCfen != "" {
			// Parse starting CFEN
			cfenState, err := cfen.ParseCFEN(startCfen)
			if err != nil {
				fmt.Printf("Error parsing starting CFEN: %v\n", err)
				os.Exit(1)
			}

			// Validate dimension if specified
			if dimension != 3 && cfenState.Dimension != dimension {
				fmt.Printf("CFEN dimension %d doesn't match specified dimension %d\n",
					cfenState.Dimension, dimension)
				os.Exit(1)
			}
			dimension = cfenState.Dimension // Use CFEN dimension

			c, err = cfenState.ToCube()
			if err != nil {
				fmt.Printf("Error converting CFEN to cube: %v\n", err)
				os.Exit(1)
			}
		} else {
			// Start with solved cube
			c = cube.NewCube(dimension)
		}

		if !useCfenOutput {
			fmt.Printf("Applying moves to %dx%dx%d cube: %s\n", dimension, dimension, dimension, moves)
			if startCfen != "" {
				fmt.Printf("Starting from CFEN: %s\n", startCfen)
			}
		}

		// Parse and apply moves
		parsedMoves, err := cube.ParseScramble(moves)
		if err != nil {
			if !useCfenOutput {
				fmt.Printf("Error parsing moves: %v\n", err)
			}
			os.Exit(1)
		}

		c.ApplyMoves(parsedMoves)

		if useCfenOutput {
			// CFEN output mode
			cfenStr, err := cfen.GenerateCFEN(c)
			if err != nil {
				fmt.Printf("Error generating CFEN: %v\n", err)
				os.Exit(1)
			}
			fmt.Print(cfenStr)
		} else {
			// Normal display mode
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
		}
	},
}

func init() {
	twistCmd.Flags().IntP("dimension", "d", 3, "Cube dimension (2, 3, 4, etc.)")
	twistCmd.Flags().BoolP("color", "c", false, "Use colored output (Unicode blocks by default)")
	twistCmd.Flags().Bool("letters", false, "Use letters instead of Unicode blocks when using --color")
	twistCmd.Flags().Bool("cfen", false, "Output final cube state as CFEN string")
	twistCmd.Flags().String("start", "", "Starting cube state as CFEN string (default: solved)")
}
