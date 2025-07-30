package cli

import (
	"fmt"
	"strings"

	"github.com/ehrlich-b/cube/internal/cube"
	"github.com/spf13/cobra"
)

var findCmd = &cobra.Command{
	Use:   "find",
	Short: "Find sequences that create specific patterns or states",
	Long: `Find move sequences that achieve specific cube states or patterns.

Examples:
  cube find --pattern solved --max-moves 4     # Find ways to solve in 4 moves
  cube find --pattern cross --max-moves 8      # Find cross-solving sequences
  cube find --scramble "R U" --max-moves 5     # Find ways to solve R U scramble
  cube find --state "solved" --from "R U R'"   # Find moves from R U R' to solved`,
}

var findPatternCmd = &cobra.Command{
	Use:   "pattern [pattern-name]",
	Short: "Find sequences that create a specific pattern",
	Long: `Find move sequences that create a specific named pattern.

Available patterns:
  - solved: Return cube to solved state
  - checkerboard: Create checkerboard pattern
  - cross: Create yellow cross on top
  - more patterns coming soon...

Examples:
  cube find pattern solved --max-moves 6
  cube find pattern cross --max-moves 8`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pattern := args[0]
		maxMoves, _ := cmd.Flags().GetInt("max-moves")
		fromState, _ := cmd.Flags().GetString("from")
		showSteps, _ := cmd.Flags().GetBool("steps")

		return runPatternSearch(pattern, maxMoves, fromState, showSteps)
	},
}

var findSequenceCmd = &cobra.Command{
	Use:   "sequence [scramble]",
	Short: "Find sequences that solve a specific scramble",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		scramble := args[0]
		maxMoves, _ := cmd.Flags().GetInt("max-moves")
		showSteps, _ := cmd.Flags().GetBool("steps")

		return runSequenceSearch(scramble, maxMoves, showSteps)
	},
}

func runPatternSearch(pattern string, maxMoves int, fromState string, showSteps bool) error {
	fmt.Printf("Searching for sequences to create '%s' pattern (max %d moves)...\n", pattern, maxMoves)

	// Create starting cube
	startCube := cube.NewCube(3)
	if fromState != "" {
		moves, err := cube.ParseScramble(fromState)
		if err != nil {
			return fmt.Errorf("error parsing from-state '%s': %v", fromState, err)
		}
		startCube.ApplyMoves(moves)
		fmt.Printf("Starting from state: %s\n", fromState)
	}

	// Define target checker based on pattern
	var isTarget func(*cube.Cube) bool
	switch strings.ToLower(pattern) {
	case "solved":
		isTarget = func(c *cube.Cube) bool { return c.IsSolved() }
	case "cross":
		isTarget = func(c *cube.Cube) bool {
			// Simple cross check - yellow center and edges on top
			return c.Faces[cube.Up][1][1] == cube.Yellow &&
				c.Faces[cube.Up][0][1] == cube.Yellow &&
				c.Faces[cube.Up][1][0] == cube.Yellow &&
				c.Faces[cube.Up][1][2] == cube.Yellow &&
				c.Faces[cube.Up][2][1] == cube.Yellow
		}
	case "checkerboard":
		isTarget = func(c *cube.Cube) bool {
			// Simplified checkerboard detection (would need more complex logic)
			return false // TODO: implement checkerboard pattern detection
		}
	default:
		return fmt.Errorf("unknown pattern '%s'. Available: solved, cross, checkerboard", pattern)
	}

	// Simple brute force search
	results := breadthFirstSearch(startCube, isTarget, maxMoves)

	if len(results) == 0 {
		fmt.Printf("No sequences found within %d moves.\n", maxMoves)
		return nil
	}

	fmt.Printf("\nFound %d sequence(s):\n", len(results))
	for i, result := range results {
		optimized := cube.OptimizeMoves(result.moves)
		fmt.Printf("%d. %s (%d moves", i+1, result.notation, len(result.moves))
		if len(optimized) != len(result.moves) {
			fmt.Printf(", %d optimized", len(optimized))
		}
		fmt.Printf(")\n")

		if showSteps {
			// Show intermediate states
			testCube := cube.NewCube(3)
			if fromState != "" {
				fromMoves, _ := cube.ParseScramble(fromState)
				testCube.ApplyMoves(fromMoves)
			}

			fmt.Printf("   Steps:\n")
			for j, move := range result.moves {
				testCube.ApplyMove(move)
				fmt.Printf("   %d. %s\n", j+1, move.String())
			}
		}
	}

	return nil
}

func runSequenceSearch(scramble string, maxMoves int, showSteps bool) error {
	fmt.Printf("Searching for solutions to '%s' (max %d moves)...\n", scramble, maxMoves)

	// Parse and apply scramble
	scrambleMoves, err := cube.ParseScramble(scramble)
	if err != nil {
		return fmt.Errorf("error parsing scramble: %v", err)
	}

	startCube := cube.NewCube(3)
	startCube.ApplyMoves(scrambleMoves)

	// Search for solutions
	isTarget := func(c *cube.Cube) bool { return c.IsSolved() }
	results := breadthFirstSearch(startCube, isTarget, maxMoves)

	if len(results) == 0 {
		fmt.Printf("No solutions found within %d moves.\n", maxMoves)
		return nil
	}

	fmt.Printf("\nFound %d solution(s):\n", len(results))
	for i, result := range results {
		optimized := cube.OptimizeMoves(result.moves)
		fmt.Printf("%d. %s (%d moves", i+1, result.notation, len(result.moves))
		if len(optimized) != len(result.moves) {
			fmt.Printf(", %d optimized", len(optimized))
		}
		fmt.Printf(")\n")
	}

	return nil
}

type searchResult struct {
	moves    []cube.Move
	notation string
}

// breadthFirstSearch performs BFS to find sequences that satisfy the target condition
func breadthFirstSearch(startCube *cube.Cube, isTarget func(*cube.Cube) bool, maxDepth int) []searchResult {
	if isTarget(startCube) {
		return []searchResult{{moves: []cube.Move{}, notation: "(already at target)"}}
	}

	type state struct {
		cube  *cube.Cube
		moves []cube.Move
		depth int
	}

	queue := []state{{cube: copyCube(startCube), moves: []cube.Move{}, depth: 0}}
	visited := make(map[string]bool)
	visited[startCube.String()] = true

	var results []searchResult
	basicMoves := []string{"R", "R'", "R2", "L", "L'", "L2", "U", "U'", "U2", "D", "D'", "D2", "F", "F'", "F2", "B", "B'", "B2"}

	for len(queue) > 0 && len(results) < 10 { // Limit results to avoid too much output
		current := queue[0]
		queue = queue[1:]

		if current.depth >= maxDepth {
			continue
		}

		for _, moveStr := range basicMoves {
			move, err := cube.ParseMove(moveStr)
			if err != nil {
				continue
			}

			// Apply move to a copy
			newCube := copyCube(current.cube)
			newCube.ApplyMove(move)

			cubeStr := newCube.String()
			if visited[cubeStr] {
				continue
			}
			visited[cubeStr] = true

			newMoves := append(current.moves, move)

			if isTarget(newCube) {
				// Found a solution
				var notation []string
				for _, m := range newMoves {
					notation = append(notation, m.String())
				}
				results = append(results, searchResult{
					moves:    newMoves,
					notation: strings.Join(notation, " "),
				})
			} else if current.depth+1 < maxDepth {
				// Continue searching
				queue = append(queue, state{
					cube:  newCube,
					moves: newMoves,
					depth: current.depth + 1,
				})
			}
		}
	}

	return results
}

func copyCube(original *cube.Cube) *cube.Cube {
	copy := cube.NewCube(original.Size)
	for face := 0; face < 6; face++ {
		for row := 0; row < original.Size; row++ {
			for col := 0; col < original.Size; col++ {
				copy.Faces[face][row][col] = original.Faces[face][row][col]
			}
		}
	}
	return copy
}

func init() {
	rootCmd.AddCommand(findCmd)
	findCmd.AddCommand(findPatternCmd)
	findCmd.AddCommand(findSequenceCmd)

	// Flags for both subcommands
	findPatternCmd.Flags().IntP("max-moves", "m", 6, "Maximum number of moves to search")
	findPatternCmd.Flags().StringP("from", "f", "", "Starting cube state (default: solved)")
	findPatternCmd.Flags().BoolP("steps", "s", false, "Show intermediate steps")

	findSequenceCmd.Flags().IntP("max-moves", "m", 8, "Maximum number of moves to search")
	findSequenceCmd.Flags().BoolP("steps", "s", false, "Show intermediate steps")
}
