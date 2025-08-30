package cli

import (
	"fmt"

	"github.com/ehrlich-b/cube/internal/cube"
	"github.com/spf13/cobra"
)

var showAlgCmd = &cobra.Command{
	Use:   "show-alg [algorithm-name]",
	Short: "Display algorithm pattern with start state, moves, and end state",
	Long:  `Display an algorithm from the database showing the start state, algorithm moves, and target state for learning purposes.`,
	Example: `  cube show-alg "Sune"
  cube show-alg "T-Perm" --animate
  cube show-alg "Cross OLL" --color`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		algorithmName := args[0]
		color, _ := cmd.Flags().GetBool("color")
		animate, _ := cmd.Flags().GetBool("animate")
		viewMode, _ := cmd.Flags().GetString("view")

		// Find algorithm in database
		results := cube.LookupAlgorithm(algorithmName)
		if len(results) == 0 {
			return fmt.Errorf("algorithm '%s' not found in database", algorithmName)
		}

		// Use first match
		alg := results[0]

		// Check if algorithm has pattern
		if alg.Pattern == "" {
			return fmt.Errorf("algorithm '%s' does not have pattern for visualization", alg.Name)
		}

		// Display algorithm header
		fmt.Printf("=== %s (%s) ===\n", alg.Name, alg.CaseID)
		if alg.Description != "" {
			fmt.Printf("Description: %s\n", alg.Description)
		}
		fmt.Printf("Category: %s\n", alg.Category)
		fmt.Printf("Moves: %s (%d moves)\n\n", alg.Moves, alg.MoveCount)

		// TODO: Implement pattern visualization with new Pattern field
		fmt.Printf("Pattern: %s\n\n", alg.Pattern)

		// Show algorithm execution
		if animate {
			return showAlgorithmAnimated(alg, color)
		}

		// Parse and apply algorithm moves
		moves, err := cube.ParseScramble(alg.Moves)
		if err != nil {
			return fmt.Errorf("failed to parse algorithm moves: %w", err)
		}

		// Create cubes for before/after comparison
		beforeCube := cube.NewCube(3)
		afterCube := cube.NewCube(3)

		// Apply all moves to after cube
		for _, move := range moves {
			afterCube.ApplyMove(move)
		}

		// Determine view mode
		useLastLayer := false
		if viewMode == "last" {
			useLastLayer = true
		} else if viewMode == "auto" {
			// Auto-detect based on algorithm category and changes
			useLastLayer = (alg.Category == "OLL" || alg.Category == "PLL" || alg.Category == "CFOP-OLL" || alg.Category == "CFOP-PLL") &&
				cube.AffectsOnlyLastLayers(beforeCube, afterCube)
		}

		// Display final state
		fmt.Println("🎯 FINAL STATE:")
		if viewMode == "both" {
			fmt.Println("Last layer view:")
			output := afterCube.LastLayerString(color, color)
			fmt.Println(output)
			fmt.Println("\nFull cube view:")
			output = afterCube.StringWithColor(color)
			fmt.Println(output)
		} else if useLastLayer {
			fmt.Println("(Last layer view)")
			output := afterCube.LastLayerString(color, color)
			fmt.Println(output)
		} else {
			output := afterCube.StringWithColor(color)
			fmt.Println(output)
		}

		// TODO: Implement pattern verification with new Pattern field
		fmt.Println("✅ Algorithm executed successfully")

		return nil
	},
}

func showAlgorithmAnimated(alg cube.Algorithm, color bool) error {
	fmt.Println("🎬 ANIMATED ALGORITHM EXECUTION:")
	fmt.Println("Press Enter between each step...")

	// Parse moves
	moves, err := cube.ParseScramble(alg.Moves)
	if err != nil {
		return fmt.Errorf("failed to parse moves: %w", err)
	}

	// Create working cube (start from solved state)
	// TODO: Implement pattern-based starting state
	workingCube := cube.NewCube(3)

	// Show each step
	for i, move := range moves {
		fmt.Printf("\nStep %d/%d: %s\n", i+1, len(moves), move.String())

		workingCube.ApplyMove(move)

		output := workingCube.StringWithColor(color)
		fmt.Println(output)

		if i < len(moves)-1 {
			fmt.Print("Press Enter to continue...")
			fmt.Scanln()
		}
	}

	fmt.Println("\n✅ Algorithm execution complete!")
	return nil
}

func init() {
	showAlgCmd.Flags().BoolP("color", "c", false, "Use colored output")
	showAlgCmd.Flags().Bool("animate", false, "Show step-by-step animation")
	showAlgCmd.Flags().String("view", "auto", "View mode: auto, last, full, both")
	rootCmd.AddCommand(showAlgCmd)
}
