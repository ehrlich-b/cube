package cli

import (
	"fmt"
	"os"

	"github.com/ehrlich-b/cube/internal/cfen"
	"github.com/ehrlich-b/cube/internal/cube"
	"github.com/spf13/cobra"
)

var verifyCmd = &cobra.Command{
	Use:   "verify <algorithm>",
	Short: "Verify an algorithm transforms start state to target state",
	Long: `Verify that an algorithm correctly transforms a cube from a start state to a target state.
Both states are specified using CFEN notation with wildcard support.

Examples:
  # Verify Sune algorithm (OLL case)
  cube verify "R U R' U R U2 R'" \
    --start "YB|Y9/R3G3R3/B3W3B3/W9/O3Y3O3/G3R3G3" \
    --target "YB|Y9/?9/?9/?9/?9/?9"

  # Verify T-Perm (PLL case)
  cube verify "R U R' U' R' F R2 U' R' U' R U R' F'" \
    --start "YB|Y9/?9/?9/W9/?9/?9" \
    --target "YB|Y9/R9/B9/W9/O9/G9"

  # Verify simple inverse (defaults to solved start/target)
  cube verify "R U R' U' U R U' R'"  # Should solve to default state

  # Verify F2L pair insertion
  cube verify "U R U' R'" \
    --start "YB|?Y?YYY?Y?/?9/?9/W9/?9/?9" \
    --target "YB|?Y?YYY?Y?/??R??R??R/??B??B??B/W9/??O??O??O/??G??G??G"`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		algorithm := args[0]

		// Get flags
		startCFEN, _ := cmd.Flags().GetString("start")
		targetCFEN, _ := cmd.Flags().GetString("target")
		verbose, _ := cmd.Flags().GetBool("verbose")
		headless, _ := cmd.Flags().GetBool("headless")
		useColor, _ := cmd.Flags().GetBool("color")
		useLetters, _ := cmd.Flags().GetBool("letters")
		useUnicode := useColor && !useLetters

		// Default to solved cube if not specified
		if startCFEN == "" {
			startCFEN = "YB|Y9/R9/B9/W9/O9/G9" // 3x3 solved
		}
		if targetCFEN == "" {
			targetCFEN = "YB|Y9/R9/B9/W9/O9/G9" // 3x3 solved
		}

		// Parse start CFEN
		startState, err := cfen.ParseCFEN(startCFEN)
		if err != nil {
			if !headless {
				fmt.Printf("Error parsing start CFEN: %v\n", err)
			}
			os.Exit(1)
		}

		// Parse target CFEN
		targetState, err := cfen.ParseCFEN(targetCFEN)
		if err != nil {
			if !headless {
				fmt.Printf("Error parsing target CFEN: %v\n", err)
			}
			os.Exit(1)
		}

		// Validate dimensions match
		if startState.Dimension != targetState.Dimension {
			if !headless {
				fmt.Printf("Error: Start and target dimensions must match (%d vs %d)\n",
					startState.Dimension, targetState.Dimension)
			}
			os.Exit(1)
		}

		// Convert start CFEN to cube
		c, err := startState.ToCube()
		if err != nil {
			if !headless {
				fmt.Printf("Error converting start CFEN to cube: %v\n", err)
			}
			os.Exit(1)
		}

		// Show start state if verbose
		if verbose && !headless {
			fmt.Printf("Start state (from CFEN):\n")
			fmt.Println(c.UnfoldedString(useColor, useUnicode))
		}

		// Parse and apply algorithm
		moves, err := cube.ParseScramble(algorithm)
		if err != nil {
			if !headless {
				fmt.Printf("Error parsing algorithm: %v\n", err)
			}
			os.Exit(1)
		}

		c.ApplyMoves(moves)

		// Show result state if verbose
		if verbose && !headless {
			fmt.Printf("\nAfter algorithm (%s):\n", algorithm)
			fmt.Println(c.UnfoldedString(useColor, useUnicode))
		}

		// Check if result matches target
		matches, err := targetState.MatchesCube(c)
		if err != nil {
			if !headless {
				fmt.Printf("Error matching result to target: %v\n", err)
			}
			os.Exit(1)
		}

		if matches {
			if !headless {
				fmt.Printf("✅ PASS: Algorithm correctly transforms start to target state\n")
				fmt.Printf("Algorithm: %s\n", algorithm)
				fmt.Printf("Move count: %d\n", len(moves))
				if verbose {
					fmt.Printf("Start:  %s\n", startCFEN)
					fmt.Printf("Target: %s\n", targetCFEN)
					actualCFEN, _ := cfen.GenerateCFEN(c)
					fmt.Printf("Actual: %s\n", actualCFEN)
				}
			}
			os.Exit(0)
		} else {
			if !headless {
				fmt.Printf("❌ FAIL: Algorithm does not achieve target state\n")
				fmt.Printf("Algorithm: %s\n", algorithm)
				if !verbose {
					fmt.Printf("\nTip: Use --verbose to see the cube states\n")
				} else {
					fmt.Printf("Start:  %s\n", startCFEN)
					fmt.Printf("Target: %s\n", targetCFEN)
					actualCFEN, _ := cfen.GenerateCFEN(c)
					fmt.Printf("Actual: %s\n", actualCFEN)
				}
			}
			os.Exit(1)
		}
	},
}

func init() {
	verifyCmd.Flags().String("start", "", "Starting CFEN state (defaults to solved)")
	verifyCmd.Flags().String("target", "", "Target CFEN state (defaults to solved)")
	verifyCmd.Flags().BoolP("verbose", "v", false, "Show cube states and transformations")
	verifyCmd.Flags().Bool("headless", false, "Exit with code 0 for pass, 1 for fail (no output)")
	verifyCmd.Flags().BoolP("color", "c", false, "Use colored output")
	verifyCmd.Flags().Bool("letters", false, "Use colored letters instead of blocks")
}
