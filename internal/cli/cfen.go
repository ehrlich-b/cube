package cli

import (
	"fmt"

	"github.com/ehrlich-b/cube/internal/cfen"
	"github.com/ehrlich-b/cube/internal/cube"
	"github.com/spf13/cobra"
)

var parseCfenCmd = &cobra.Command{
	Use:   "parse-cfen <cfen-string>",
	Short: "Parse and display a CFEN string as a cube state",
	Long: `Parse a CFEN (Cube Forsyth-Edwards Notation) string and display the resulting cube state.

Examples:
  cube parse-cfen "WG|W9/R9/G9/Y9/O9/B9"                    # Solved 3x3
  cube parse-cfen "WG|?W?WWW?W?/?9/?9/?9/?9/?9"              # White cross only
  cube parse-cfen "WG|W16/R16/G16/Y16/O16/B16"               # Solved 4x4
  cube parse-cfen "WG|Y25/?25/?25/?25/?25/?25"               # 5x5 OLL drill`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfenStr := args[0]

		// Parse CFEN string
		cfenState, err := cfen.ParseCFEN(cfenStr)
		if err != nil {
			return fmt.Errorf("failed to parse CFEN: %v", err)
		}

		// Convert to cube for display
		cube, err := cfenState.ToCube()
		if err != nil {
			return fmt.Errorf("failed to convert CFEN to cube: %v", err)
		}

		// Get display flags
		useColor, _ := cmd.Flags().GetBool("color")
		useUnicode := useColor // Use Unicode blocks when color is enabled
		useLetters, _ := cmd.Flags().GetBool("letters")
		if useLetters {
			useUnicode = false // Use colored letters instead of blocks
		}

		// Display cube information
		fmt.Printf("CFEN: %s\n", cfenStr)
		fmt.Printf("Dimension: %dx%dx%d\n", cfenState.Dimension, cfenState.Dimension, cfenState.Dimension)
		fmt.Printf("Orientation: %s up, %s front\n",
			cfenState.Orientation.Up.String(),
			cfenState.Orientation.Front.String())
		fmt.Printf("Solved: %t\n\n", cube.IsSolved())

		// Display cube state
		fmt.Print(cube.UnfoldedString(useColor && !useUnicode, useUnicode))

		return nil
	},
}

var generateCfenCmd = &cobra.Command{
	Use:   "generate-cfen <scramble>",
	Short: "Apply scramble moves and output the resulting CFEN string",
	Long: `Apply a scramble sequence to a solved cube and output the resulting state as a CFEN string.

Examples:
  cube generate-cfen "R U R' U'"                    # Simple scramble
  cube generate-cfen "R U R' U'" --dimension 4      # 4x4 cube
  cube generate-cfen "R U R' U'" --start "WG|..."   # Custom starting state`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		scramble := args[0]

		// Get dimension
		dimension, _ := cmd.Flags().GetInt("dimension")
		if dimension < 2 {
			dimension = 3 // Default to 3x3
		}

		// Get starting state
		startCfen, _ := cmd.Flags().GetString("start")
		var resultCube *cube.Cube
		var err error

		if startCfen != "" {
			// Parse starting CFEN
			cfenState, err := cfen.ParseCFEN(startCfen)
			if err != nil {
				return fmt.Errorf("invalid starting CFEN: %v", err)
			}

			// Validate dimension if specified
			if dimension != 3 && cfenState.Dimension != dimension {
				return fmt.Errorf("CFEN dimension %d doesn't match specified dimension %d",
					cfenState.Dimension, dimension)
			}

			resultCube, err = cfenState.ToCube()
			if err != nil {
				return fmt.Errorf("failed to convert starting CFEN to cube: %v", err)
			}
		} else {
			// Start with solved cube
			resultCube = cube.NewCube(dimension)
		}

		// Parse and apply scramble
		if scramble != "" {
			moves, err := cube.ParseScramble(scramble)
			if err != nil {
				return fmt.Errorf("invalid scramble: %v", err)
			}
			resultCube.ApplyMoves(moves)
		}

		// Generate CFEN
		cfenStr, err := cfen.GenerateCFEN(resultCube)
		if err != nil {
			return fmt.Errorf("failed to generate CFEN: %v", err)
		}

		fmt.Println(cfenStr)
		return nil
	},
}

var verifyCfenCmd = &cobra.Command{
	Use:   "verify-cfen <scramble> <solution> --target <cfen>",
	Short: "Verify that a solution reaches the target CFEN state",
	Long: `Apply a scramble and solution, then verify the result matches the target CFEN pattern.
Supports wildcard matching where '?' positions are ignored.

Examples:
  # Verify white cross solution
  cube verify-cfen "R U R' U'" "U R U' R'" --target "WG|?W?WWW?W?/?9/?9/?9/?9/?9"

  # Verify full solve
  cube verify-cfen "R U R' U'" "solution" --target "WG|W9/R9/G9/Y9/O9/B9"

  # Verify OLL completion
  cube verify-cfen "scramble" "solution" --target "WG|Y9/?9/?9/?9/?9/?9"`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		scramble := args[0]
		solution := args[1]

		// Get target CFEN
		targetCfen, _ := cmd.Flags().GetString("target")
		if targetCfen == "" {
			return fmt.Errorf("--target flag is required")
		}

		// Parse target CFEN
		targetState, err := cfen.ParseCFEN(targetCfen)
		if err != nil {
			return fmt.Errorf("invalid target CFEN: %v", err)
		}

		// Get dimension
		dimension, _ := cmd.Flags().GetInt("dimension")
		if dimension < 2 {
			dimension = targetState.Dimension // Use CFEN dimension as default
		} else if dimension != targetState.Dimension {
			return fmt.Errorf("specified dimension %d doesn't match target CFEN dimension %d",
				dimension, targetState.Dimension)
		}

		// Start with solved cube
		testCube := cube.NewCube(dimension)

		// Apply scramble
		if scramble != "" {
			scrambleMoves, err := cube.ParseScramble(scramble)
			if err != nil {
				return fmt.Errorf("invalid scramble: %v", err)
			}
			testCube.ApplyMoves(scrambleMoves)
		}

		// Apply solution
		if solution != "" {
			solutionMoves, err := cube.ParseScramble(solution)
			if err != nil {
				return fmt.Errorf("invalid solution: %v", err)
			}
			testCube.ApplyMoves(solutionMoves)
		}

		// Check if result matches target pattern
		matches, err := targetState.MatchesCube(testCube)
		if err != nil {
			return fmt.Errorf("failed to match against target: %v", err)
		}

		// Get verbose flag
		verbose, _ := cmd.Flags().GetBool("verbose")

		if matches {
			fmt.Println("✅ PASS: Solution matches target CFEN pattern")
			if verbose {
				actualCfen, _ := cfen.GenerateCFEN(testCube)
				fmt.Printf("Target:  %s\n", targetCfen)
				fmt.Printf("Actual:  %s\n", actualCfen)
			}
		} else {
			fmt.Println("❌ FAIL: Solution does not match target CFEN pattern")
			if verbose {
				actualCfen, _ := cfen.GenerateCFEN(testCube)
				fmt.Printf("Target:  %s\n", targetCfen)
				fmt.Printf("Actual:  %s\n", actualCfen)
			}
			return fmt.Errorf("verification failed")
		}

		return nil
	},
}

var matchCfenCmd = &cobra.Command{
	Use:   "match-cfen <current-cfen> <target-cfen>",
	Short: "Compare two CFEN strings and show differences",
	Long: `Compare two CFEN strings and show which positions differ.
Supports wildcard matching where '?' positions are ignored.

Examples:
  cube match-cfen "WG|W9/R9/G9/Y9/O9/B9" "WG|W9/R9/G9/Y9/O9/B9"     # Perfect match
  cube match-cfen "WG|YWY..." "WG|?W?..."                             # Partial match`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		currentCfen := args[0]
		targetCfen := args[1]

		// Parse both CFEN strings
		currentState, err := cfen.ParseCFEN(currentCfen)
		if err != nil {
			return fmt.Errorf("invalid current CFEN: %v", err)
		}

		targetState, err := cfen.ParseCFEN(targetCfen)
		if err != nil {
			return fmt.Errorf("invalid target CFEN: %v", err)
		}

		// Validate dimensions match
		if currentState.Dimension != targetState.Dimension {
			return fmt.Errorf("dimension mismatch: current %d vs target %d",
				currentState.Dimension, targetState.Dimension)
		}

		// Convert current to cube for matching
		currentCube, err := currentState.ToCube()
		if err != nil {
			return fmt.Errorf("failed to convert current CFEN to cube: %v", err)
		}

		// Check match
		matches, err := targetState.MatchesCube(currentCube)
		if err != nil {
			return fmt.Errorf("failed to match: %v", err)
		}

		if matches {
			fmt.Println("✅ MATCH: Current state matches target pattern")
		} else {
			fmt.Println("❌ NO MATCH: Current state does not match target pattern")
		}

		fmt.Printf("Current: %s\n", currentCfen)
		fmt.Printf("Target:  %s\n", targetCfen)

		return nil
	},
}

func init() {
	// Add flags to parse-cfen
	parseCfenCmd.Flags().Bool("color", false, "Use colored output")
	parseCfenCmd.Flags().Bool("letters", false, "Use colored letters instead of blocks")

	// Add flags to generate-cfen
	generateCfenCmd.Flags().Int("dimension", 3, "Cube dimension (2-20)")
	generateCfenCmd.Flags().String("start", "", "Starting CFEN state (default: solved)")

	// Add flags to verify-cfen
	verifyCfenCmd.Flags().String("target", "", "Target CFEN pattern (required)")
	verifyCfenCmd.Flags().Int("dimension", 0, "Cube dimension (auto-detect from target if not specified)")
	verifyCfenCmd.Flags().Bool("verbose", false, "Show detailed comparison")
	verifyCfenCmd.MarkFlagRequired("target")

	// Register commands
	rootCmd.AddCommand(parseCfenCmd)
	rootCmd.AddCommand(generateCfenCmd)
	rootCmd.AddCommand(verifyCfenCmd)
	rootCmd.AddCommand(matchCfenCmd)
}
