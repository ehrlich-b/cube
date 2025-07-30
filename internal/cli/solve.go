package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/ehrlich-b/cube/internal/cube"
	"github.com/spf13/cobra"
)

var solveCmd = &cobra.Command{
	Use:   "solve [scramble]",
	Short: "Solve a scrambled cube",
	Long: `Solve a scrambled cube using the specified algorithm.
The scramble should be provided as a string of moves.

Use --headless for programmatic output (space-separated moves only).`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		scramble := args[0]
		algorithm, _ := cmd.Flags().GetString("algorithm")
		dimension, _ := cmd.Flags().GetInt("dimension")
		headless, _ := cmd.Flags().GetBool("headless")

		if !headless {
			fmt.Printf("Solving %dx%dx%d cube with scramble: %s\n", dimension, dimension, dimension, scramble)
			fmt.Printf("Using algorithm: %s\n", algorithm)
		}

		// Create cube and apply scramble
		c := cube.NewCube(dimension)
		moves, err := cube.ParseScramble(scramble)
		if err != nil {
			if !headless {
				fmt.Printf("Error parsing scramble: %v\n", err)
			}
			os.Exit(1)
		}

		c.ApplyMoves(moves)

		if !headless {
			useColor, _ := cmd.Flags().GetBool("color")
			useLetters, _ := cmd.Flags().GetBool("letters")
			useUnicode := useColor && !useLetters

			fmt.Printf("\nCube state after scramble:\n%s\n", c.UnfoldedString(useColor, useUnicode))
		}

		// Get solver and solve
		solver, err := cube.GetSolver(algorithm)
		if err != nil {
			if !headless {
				fmt.Printf("Error getting solver: %v\n", err)
			}
			os.Exit(1)
		}

		result, err := solver.Solve(c)
		if err != nil {
			if !headless {
				fmt.Printf("Error solving cube: %v\n", err)
			}
			os.Exit(1)
		}

		// Format solution
		var solutionStr strings.Builder
		for i, move := range result.Solution {
			if i > 0 {
				solutionStr.WriteString(" ")
			}
			solutionStr.WriteString(move.String())
		}

		if headless {
			// Headless mode: output only the space-separated move list
			fmt.Print(solutionStr.String())
		} else {
			// Normal mode: full output
			fmt.Printf("Solution: %s\n", solutionStr.String())
			fmt.Printf("Steps: %d\n", result.Steps)
			fmt.Printf("Time: %v\n", result.Duration)
		}
	},
}

func init() {
	solveCmd.Flags().StringP("algorithm", "a", "beginner", "Solving algorithm to use (beginner, cfop, kociemba)")
	solveCmd.Flags().IntP("dimension", "d", 3, "Cube dimension (2, 3, 4, etc.)")
	solveCmd.Flags().BoolP("color", "c", false, "Use colored output (Unicode blocks by default)")
	solveCmd.Flags().Bool("letters", false, "Use letters instead of Unicode blocks when using --color")
	solveCmd.Flags().Bool("headless", false, "Output only space-separated moves for programmatic use")
}
