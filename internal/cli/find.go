package cli

import (
	"fmt"
	"strings"

	"github.com/ehrlich-b/cube/internal/cfen"
	"github.com/ehrlich-b/cube/internal/cube"
	"github.com/spf13/cobra"
)

var findCmd = &cobra.Command{
	Use:   "find",
	Short: "Find sequences that create specific patterns or states",
	Long: `Find move sequences that achieve specific cube states or patterns.

Examples:
  cube find pattern solved --max-moves 4 --from "R U"
  cube find pattern cross --max-moves 8 --from "F R U"
  cube find sequence "R U" --max-moves 5
  cube find --target "YB|Y9/?9/?9/?9/?9/?9" --max-moves 4`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		target, _ := cmd.Flags().GetString("target")
		if target == "" {
			return cmd.Help()
		}
		maxMoves, _ := cmd.Flags().GetInt("max-moves")
		start, _ := cmd.Flags().GetString("start")
		options, err := findOptions(cmd)
		if err != nil {
			return err
		}
		return runPatternSearchWithOptions(target, maxMoves, start, false, options)
	},
}

var findPatternCmd = &cobra.Command{
	Use:   "pattern [pattern-name]",
	Short: "Find sequences that create a specific pattern",
	Long: `Find move sequences that create a specific named pattern.

Available patterns:
  - solved: Return cube to solved state
  - checkerboard: Create checkerboard pattern
  - cross: Create yellow cross on top
  - CFEN: Match a concrete state or wildcard sticker pattern

Examples:
  cube find pattern solved --max-moves 6
  cube find pattern cross --max-moves 8`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pattern := args[0]
		maxMoves, _ := cmd.Flags().GetInt("max-moves")
		fromState, _ := cmd.Flags().GetString("from")
		showSteps, _ := cmd.Flags().GetBool("steps")

		options, err := findOptions(cmd)
		if err != nil {
			return err
		}
		start, _ := cmd.Flags().GetString("start")
		if start != "" {
			if fromState != "" {
				return fmt.Errorf("use only one of --start and --from")
			}
			fromState = start
		}
		return runPatternSearchWithOptions(pattern, maxMoves, fromState, showSteps, options)
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

		options, err := findOptions(cmd)
		if err != nil {
			return err
		}
		start, _ := cmd.Flags().GetString("start")
		return runSequenceSearchWithOptions(scramble, maxMoves, showSteps, start, options)
	},
}

func runPatternSearch(pattern string, maxMoves int, fromState string, showSteps bool) error {
	return runPatternSearchWithOptions(pattern, maxMoves, fromState, showSteps, findSearchOptions{dimension: 3})
}

type findSearchOptions struct {
	dimension int
	moves     []cube.Move
}

func findOptions(cmd *cobra.Command) (findSearchOptions, error) {
	dimension, _ := cmd.Flags().GetInt("dimension")
	if dimension < 2 || dimension > 20 {
		return findSearchOptions{}, fmt.Errorf("dimension must be between 2 and 20")
	}
	text, _ := cmd.Flags().GetString("moves")
	var moves []cube.Move
	if text != "" {
		var err error
		moves, err = cube.ParseMoves(strings.ReplaceAll(text, ",", " "))
		if err != nil {
			return findSearchOptions{}, err
		}
		if len(moves) == 0 {
			return findSearchOptions{}, fmt.Errorf("empty move set")
		}
	}
	return findSearchOptions{dimension: dimension, moves: moves}, nil
}

func findStart(text string, dimension int) (*cube.Cube, error) {
	if strings.Contains(text, "|") {
		state, err := cfen.ParseCFEN(text)
		if err != nil {
			return nil, err
		}
		if state.Dimension != dimension {
			return nil, fmt.Errorf("start CFEN dimension differs from --dimension")
		}
		c, err := state.ToCube()
		if err != nil {
			return nil, err
		}
		for _, face := range c.Faces {
			for _, row := range face {
				for _, color := range row {
					if color == cube.Grey {
						return nil, fmt.Errorf("start state must be concrete")
					}
				}
			}
		}
		if dimension == 3 {
			if err := cube.Validate3x3(c); err != nil {
				return nil, err
			}
		}
		return c, nil
	}
	c := cube.NewCube(dimension)
	moves, err := cube.ParseScramble(text)
	if err != nil {
		return nil, err
	}
	c.ApplyMoves(moves)
	return c, nil
}

func runPatternSearchWithOptions(pattern string, maxMoves int, fromState string, showSteps bool, options findSearchOptions) error {
	fmt.Printf("Searching for sequences to create '%s' pattern (max %d moves)...\n", pattern, maxMoves)

	// Create starting cube
	startCube, err := findStart(fromState, options.dimension)
	if err != nil {
		return fmt.Errorf("error parsing from-state '%s': %v", fromState, err)
	}
	if fromState != "" {
		fmt.Printf("Starting from state: %s\n", fromState)
	}

	// Define target checker based on pattern
	target := cube.NewCube(options.dimension)
	if strings.Contains(pattern, "|") {
		state, err := cfen.ParseCFEN(pattern)
		if err != nil {
			return err
		}
		if state.Dimension != options.dimension {
			return fmt.Errorf("target CFEN dimension differs from --dimension")
		}
		target, err = state.ToCube()
		if err != nil {
			return err
		}
	} else {
		switch strings.ToLower(pattern) {
		case "solved":
			for f := range target.Faces {
				for r := range target.Faces[f] {
					for col := range target.Faces[f][r] {
						target.Faces[f][r][col] = startCube.Faces[f][options.dimension/2][options.dimension/2]
					}
				}
			}
		case "cross":
			if options.dimension != 3 {
				return fmt.Errorf("cross pattern requires dimension 3")
			}
			for f := range target.Faces {
				for r := range target.Faces[f] {
					for col := range target.Faces[f][r] {
						target.Faces[f][r][col] = cube.Grey
					}
				}
			}
			for _, pos := range [][2]int{{1, 1}, {0, 1}, {1, 0}, {1, 2}, {2, 1}} {
				target.Faces[cube.Up][pos[0]][pos[1]] = cube.Yellow
			}
		case "checkerboard":
			moves, _ := cube.ParseMoves("R2 L2 U2 D2 F2 B2")
			target.ApplyMoves(moves)
		default:
			return fmt.Errorf("unknown pattern '%s'. Available: solved, cross, checkerboard", pattern)
		}
	}

	// Simple brute force search
	var results []searchResult
	if options.dimension != 3 && strings.EqualFold(pattern, "solved") {
		results = breadthFirstSearchLimit(startCube, func(c *cube.Cube) bool { return c.IsSolved() }, maxMoves, options.moves, 1)
	} else {
		results = coordinatePatternResults(startCube, target, options.moves, maxMoves)
	}

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
			testCube := copyCube(startCube)

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
	return runSequenceSearchWithOptions(scramble, maxMoves, showSteps, "", findSearchOptions{dimension: 3})
}

func runSequenceSearchWithOptions(scramble string, maxMoves int, showSteps bool, start string, options findSearchOptions) error {
	fmt.Printf("Searching for solutions to '%s' (max %d moves)...\n", scramble, maxMoves)

	// Parse and apply scramble
	scrambleMoves, err := cube.ParseScramble(scramble)
	if err != nil {
		return fmt.Errorf("error parsing scramble: %v", err)
	}

	startCube, err := findStart(start, options.dimension)
	if err != nil {
		return err
	}
	startCube.ApplyMoves(scrambleMoves)

	// Search for solutions
	target := copyCube(startCube)
	for f := range target.Faces {
		for r := range target.Faces[f] {
			for col := range target.Faces[f][r] {
				target.Faces[f][r][col] = startCube.Faces[f][options.dimension/2][options.dimension/2]
			}
		}
	}
	var results []searchResult
	if options.dimension != 3 {
		results = breadthFirstSearchLimit(startCube, func(c *cube.Cube) bool { return c.IsSolved() }, maxMoves, options.moves, 1)
	} else {
		results = coordinatePatternResults(startCube, target, options.moves, maxMoves)
	}

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

func coordinatePatternResults(start, target *cube.Cube, moves []cube.Move, maxDepth int) []searchResult {
	path, ok := cube.FindPattern(start, target, moves, maxDepth)
	if !ok {
		return nil
	}
	notation := cube.FormatMoves(path)
	if len(path) == 0 {
		notation = "(already at target)"
	}
	return []searchResult{{moves: path, notation: notation}}
}

// breadthFirstSearch performs BFS to find sequences that satisfy the target condition
func breadthFirstSearch(startCube *cube.Cube, isTarget func(*cube.Cube) bool, maxDepth int) []searchResult {
	return breadthFirstSearchMoves(startCube, isTarget, maxDepth, nil)
}

func breadthFirstSearchMoves(startCube *cube.Cube, isTarget func(*cube.Cube) bool, maxDepth int, moves []cube.Move) []searchResult {
	return breadthFirstSearchLimit(startCube, isTarget, maxDepth, moves, 10)
}

func breadthFirstSearchLimit(startCube *cube.Cube, isTarget func(*cube.Cube) bool, maxDepth int, moves []cube.Move, resultLimit int) []searchResult {
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
	if moves == nil {
		moves, _ = cube.ParseMoves("R R' R2 L L' L2 U U' U2 D D' D2 F F' F2 B B' B2")
	}

	for len(queue) > 0 && len(results) < resultLimit {
		current := queue[0]
		queue = queue[1:]

		if current.depth >= maxDepth {
			continue
		}

		for _, move := range moves {

			// Apply move to a copy
			newCube := copyCube(current.cube)
			newCube.ApplyMove(move)

			cubeStr := newCube.String()
			if visited[cubeStr] {
				continue
			}
			visited[cubeStr] = true

			// Copy before append: current.moves may have spare capacity from an
			// earlier append, so a bare append would write into a backing array
			// shared by sibling states and corrupt their recorded solutions.
			newMoves := make([]cube.Move, len(current.moves), len(current.moves)+1)
			copy(newMoves, current.moves)
			newMoves = append(newMoves, move)

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
				if len(results) == resultLimit {
					return results
				}
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
	findCmd.Flags().IntP("max-moves", "m", 8, "Maximum number of moves to search")
	findCmd.Flags().String("target", "", "Target CFEN, with optional wildcards")
	for _, cmd := range []*cobra.Command{findCmd, findPatternCmd, findSequenceCmd} {
		cmd.Flags().IntP("dimension", "d", 3, "Cube dimension (3x3 uses coordinate IDA*)")
		cmd.Flags().String("start", "", "Concrete starting CFEN state")
		cmd.Flags().String("moves", "", "Move set, separated by spaces or commas (default: 18 face turns)")
	}
}
