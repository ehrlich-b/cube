package cli

import (
	"fmt"

	"github.com/ehrlich-b/cube/internal/cube"
	"github.com/spf13/cobra"
)

var optimizeCmd = &cobra.Command{
	Use:   "optimize [moves]",
	Short: "Optimize a sequence of moves",
	Long: `Optimize a sequence of moves by combining consecutive moves and removing cancellations.

Examples:
  cube optimize "R R"           # Outputs: R2
  cube optimize "R R'"          # Outputs: (empty - moves cancel)
  cube optimize "R U R' U'"     # Outputs: R U R' U' (no optimization possible)
  cube optimize "R R R"         # Outputs: R'
  cube optimize "F2 F2"         # Outputs: (empty - moves cancel)`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		moves := args[0]

		// Parse the moves
		parsedMoves, err := cube.ParseScramble(moves)
		if err != nil {
			return fmt.Errorf("error parsing moves: %v", err)
		}

		// Get original count
		originalCount := len(parsedMoves)

		// Optimize the moves
		optimized, err := cube.OptimizeScramble(moves)
		if err != nil {
			return fmt.Errorf("error optimizing moves: %v", err)
		}

		// Get optimized count
		optimizedMoves, _ := cube.ParseScramble(optimized)
		optimizedCount := len(optimizedMoves)

		// Display results
		fmt.Printf("Original:  %s (%d moves)\n", moves, originalCount)
		if optimized == "" {
			fmt.Printf("Optimized: (empty - all moves cancel out)\n")
		} else {
			fmt.Printf("Optimized: %s (%d moves)\n", optimized, optimizedCount)
		}

		if originalCount != optimizedCount {
			saved := originalCount - optimizedCount
			fmt.Printf("Saved %d move(s)\n", saved)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(optimizeCmd)
}
