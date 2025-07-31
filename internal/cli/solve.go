package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/ehrlich-b/cube/internal/cfen"
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
		useCfenOutput, _ := cmd.Flags().GetBool("cfen")
		startCfen, _ := cmd.Flags().GetString("start")

		// Create cube from starting position
		var c *cube.Cube
		if startCfen != "" {
			// Parse starting CFEN
			cfenState, err := cfen.ParseCFEN(startCfen)
			if err != nil {
				if !headless {
					fmt.Printf("Error parsing starting CFEN: %v\n", err)
				}
				os.Exit(1)
			}

			// Validate dimension if specified
			if dimension != 3 && cfenState.Dimension != dimension {
				if !headless {
					fmt.Printf("CFEN dimension %d doesn't match specified dimension %d\n",
						cfenState.Dimension, dimension)
				}
				os.Exit(1)
			}
			dimension = cfenState.Dimension // Use CFEN dimension

			c, err = cfenState.ToCube()
			if err != nil {
				if !headless {
					fmt.Printf("Error converting CFEN to cube: %v\n", err)
				}
				os.Exit(1)
			}
		} else {
			// Start with solved cube
			c = cube.NewCube(dimension)
		}

		if !headless {
			fmt.Printf("Solving %dx%dx%d cube with scramble: %s\n", dimension, dimension, dimension, scramble)
			fmt.Printf("Using algorithm: %s\n", algorithm)
			if startCfen != "" {
				fmt.Printf("Starting from CFEN: %s\n", startCfen)
			}
		}

		// Apply scramble to cube
		if scramble != "" {
			moves, err := cube.ParseScramble(scramble)
			if err != nil {
				if !headless {
					fmt.Printf("Error parsing scramble: %v\n", err)
				}
				os.Exit(1)
			}
			c.ApplyMoves(moves)
		}

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

		// Apply solution to get final state
		c.ApplyMoves(result.Solution)

		// Format solution
		var solutionStr strings.Builder
		for i, move := range result.Solution {
			if i > 0 {
				solutionStr.WriteString(" ")
			}
			solutionStr.WriteString(move.String())
		}

		if useCfenOutput {
			// CFEN output mode
			cfenStr, err := cfen.GenerateCFEN(c)
			if err != nil {
				if !headless {
					fmt.Printf("Error generating CFEN: %v\n", err)
				}
				os.Exit(1)
			}
			fmt.Print(cfenStr)
		} else if headless {
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
	solveCmd.Flags().Bool("cfen", false, "Output final cube state as CFEN string instead of moves")
	solveCmd.Flags().String("start", "", "Starting cube state as CFEN string (default: solved)")
}
