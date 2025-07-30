package cli

import (
	"fmt"
	"os"

	"github.com/ehrlich-b/cube/internal/cube"
	"github.com/spf13/cobra"
)

var verifyCmd = &cobra.Command{
	Use:   "verify [scramble] [solution]",
	Short: "Verify if a solution solves a scrambled cube",
	Long: `Verify takes a scramble and a solution, applies both in sequence,
and checks if the cube ends up in a solved state.

Examples:
  cube verify "R U R' U'" "U R U' R'"
  cube verify "F R U' R' F'" "F R U R' U' F'" --dimension 3

Use --headless for programmatic use (exit code 0 for solved, 1 for not solved).`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		scramble := args[0]
		solution := args[1]
		dimension, _ := cmd.Flags().GetInt("dimension")
		useColor, _ := cmd.Flags().GetBool("color")
		useLetters, _ := cmd.Flags().GetBool("letters")
		useUnicode := useColor && !useLetters
		verbose, _ := cmd.Flags().GetBool("verbose")
		headless, _ := cmd.Flags().GetBool("headless")

		// Create a new cube
		c := cube.NewCube(dimension)

		// Parse and apply scramble
		scrambleMoves, err := cube.ParseScramble(scramble)
		if err != nil {
			if !headless {
				fmt.Printf("Error parsing scramble: %v\n", err)
			}
			os.Exit(1)
		}

		c.ApplyMoves(scrambleMoves)

		if verbose && !headless {
			fmt.Printf("Cube after scramble (%s):\n", scramble)
			fmt.Println(c.UnfoldedString(useColor, useUnicode))
		}

		// Parse and apply solution
		solutionMoves, err := cube.ParseScramble(solution)
		if err != nil {
			if !headless {
				fmt.Printf("Error parsing solution: %v\n", err)
			}
			os.Exit(1)
		}

		c.ApplyMoves(solutionMoves)

		if verbose && !headless {
			fmt.Printf("Cube after solution (%s):\n", solution)
			fmt.Println(c.UnfoldedString(useColor, useUnicode))
		}

		// Check if solved
		if c.IsSolved() {
			if !headless {
				fmt.Printf("✅ SOLVED! The solution correctly solves the scramble.\n")
				fmt.Printf("Scramble: %s\n", scramble)
				fmt.Printf("Solution: %s\n", solution)
				fmt.Printf("Total moves: %d\n", len(scrambleMoves)+len(solutionMoves))
			}
			// Exit with code 0 for success (solved)
			os.Exit(0)
		} else {
			if !headless {
				fmt.Printf("❌ NOT SOLVED! The solution does not solve the scramble.\n")
				fmt.Printf("Scramble: %s\n", scramble)
				fmt.Printf("Solution: %s\n", solution)
				if !verbose {
					fmt.Printf("\nTip: Use --verbose to see the cube state after each step.\n")
				}
			}
			// Exit with code 1 for failure (not solved)
			os.Exit(1)
		}
	},
}

func init() {
	verifyCmd.Flags().IntP("dimension", "d", 3, "Cube dimension (2, 3, 4, etc.)")
	verifyCmd.Flags().BoolP("color", "c", false, "Use colored output (Unicode blocks by default)")
	verifyCmd.Flags().Bool("letters", false, "Use letters instead of Unicode blocks when using --color")
	verifyCmd.Flags().BoolP("verbose", "v", false, "Show cube state after each step")
	verifyCmd.Flags().Bool("headless", false, "Exit with code 0 for solved, 1 for not solved (no output)")
}
