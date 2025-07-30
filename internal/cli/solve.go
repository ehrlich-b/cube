package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ehrlich-b/cube/internal/cube"
)

var solveCmd = &cobra.Command{
	Use:   "solve [scramble]",
	Short: "Solve a scrambled cube",
	Long: `Solve a scrambled cube using the specified algorithm.
The scramble should be provided as a string of moves.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		scramble := args[0]
		algorithm, _ := cmd.Flags().GetString("algorithm")
		dimension, _ := cmd.Flags().GetInt("dimension")
		
		fmt.Printf("Solving %dx%dx%d cube with scramble: %s\n", dimension, dimension, dimension, scramble)
		fmt.Printf("Using algorithm: %s\n", algorithm)
		
		// Create cube and apply scramble
		c := cube.NewCube(dimension)
		moves, err := cube.ParseScramble(scramble)
		if err != nil {
			fmt.Printf("Error parsing scramble: %v\n", err)
			return
		}
		
		c.ApplyMoves(moves)
		
		// Get solver and solve
		solver, err := cube.GetSolver(algorithm)
		if err != nil {
			fmt.Printf("Error getting solver: %v\n", err)
			return
		}
		
		result, err := solver.Solve(c)
		if err != nil {
			fmt.Printf("Error solving cube: %v\n", err)
			return
		}
		
		// Format solution
		var solutionStr strings.Builder
		for i, move := range result.Solution {
			if i > 0 {
				solutionStr.WriteString(" ")
			}
			solutionStr.WriteString(move.String())
		}
		
		fmt.Printf("Solution: %s\n", solutionStr.String())
		fmt.Printf("Steps: %d\n", result.Steps)
		fmt.Printf("Time: %v\n", result.Duration)
	},
}

func init() {
	solveCmd.Flags().StringP("algorithm", "a", "beginner", "Solving algorithm to use (beginner, cfop, kociemba)")
	solveCmd.Flags().IntP("dimension", "d", 3, "Cube dimension (2, 3, 4, etc.)")
}