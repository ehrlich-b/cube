package main

import (
	"fmt"
	"os"

	"github.com/ehrlich-b/cube/internal/cfen"
	"github.com/ehrlich-b/cube/internal/cube"
)

func main() {
	verbose := false
	category := ""

	// Simple argument parsing
	for i, arg := range os.Args[1:] {
		switch arg {
		case "--verbose", "-v":
			verbose = true
		case "--category":
			if i+1 < len(os.Args)-1 {
				category = os.Args[i+2]
			}
		}
	}

	algorithms := cube.AlgorithmDatabase
	var toVerify []cube.Algorithm

	// Filter by category if specified
	if category != "" {
		algorithms = cube.GetByCategory(category)
	}

	// Filter to only algorithms with patterns
	for _, alg := range algorithms {
		if alg.Pattern != "" {
			toVerify = append(toVerify, alg)
		}
	}

	if len(toVerify) == 0 {
		if category != "" {
			fmt.Printf("No algorithms with CFEN patterns found in category '%s'\n", category)
		} else {
			fmt.Println("No algorithms with CFEN patterns found in database")
		}
		return
	}

	// Counters for summary
	totalCount := len(toVerify)
	passedCount := 0
	failedCount := 0
	errorCount := 0

	fmt.Printf("Verifying %d algorithms", totalCount)
	if category != "" {
		fmt.Printf(" in category '%s'", category)
	}
	fmt.Printf("...\n\n")

	// Verify each algorithm
	for i, alg := range toVerify {
		fmt.Printf("[%d/%d] Testing %s (%s)...", i+1, totalCount, alg.Name, alg.CaseID)

		// Set up start and target CFENs
		startCFEN := "YB|Y9/R9/B9/W9/O9/G9" // Always start from solved cube
		targetCFEN := alg.Pattern // Expected pattern after applying algorithm

		// Perform verification
		err := verifyAlgorithm(alg, startCFEN, targetCFEN, false)

		if err != nil {
			if err.Error() == "verification failed" {
				failedCount++
				fmt.Printf(" ❌ FAIL\n")
				if verbose {
					fmt.Printf("    Reason: Algorithm does not achieve target state\n")
				}
			} else {
				errorCount++
				fmt.Printf(" ⚠️  ERROR\n")
				if verbose {
					fmt.Printf("    Reason: %v\n", err)
				}
			}
		} else {
			passedCount++
			fmt.Printf(" ✅ PASS\n")
		}

		// Show detailed output for verbose mode
		if verbose {
			fmt.Printf("    Algorithm: %s\n", alg.Moves)
			fmt.Printf("    Start:     %s\n", startCFEN)
			fmt.Printf("    Target:    %s\n", targetCFEN)
			fmt.Println()
		}
	}

	// Print summary
	fmt.Printf("\n=== Verification Summary ===\n")
	fmt.Printf("Total algorithms tested: %d\n", totalCount)
	fmt.Printf("✅ Passed: %d (%.1f%%)\n", passedCount, float64(passedCount)/float64(totalCount)*100)
	fmt.Printf("❌ Failed: %d (%.1f%%)\n", failedCount, float64(failedCount)/float64(totalCount)*100)
	if errorCount > 0 {
		fmt.Printf("⚠️  Errors: %d (%.1f%%)\n", errorCount, float64(errorCount)/float64(totalCount)*100)
	}

	if passedCount == totalCount {
		fmt.Printf("\n🎉 All algorithms verified successfully!\n")
	} else {
		fmt.Printf("\n⚠️  Some algorithms failed verification. Use --verbose for details.\n")
		os.Exit(1)
	}
}

func verifyAlgorithm(algorithm cube.Algorithm, startCFEN, targetCFEN string, verbose bool) error {
	// Parse start CFEN
	startState, err := cfen.ParseCFEN(startCFEN)
	if err != nil {
		return fmt.Errorf("parsing start CFEN: %v", err)
	}

	// Parse target CFEN
	targetState, err := cfen.ParseCFEN(targetCFEN)
	if err != nil {
		return fmt.Errorf("parsing target CFEN: %v", err)
	}

	// Convert start CFEN to cube
	c, err := startState.ToCube()
	if err != nil {
		return fmt.Errorf("converting start CFEN to cube: %v", err)
	}

	// Parse and apply algorithm
	moves, err := cube.ParseScramble(algorithm.Moves)
	if err != nil {
		return fmt.Errorf("parsing algorithm moves: %v", err)
	}

	c.ApplyMoves(moves)

	// Check if result matches target
	matches, err := targetState.MatchesCube(c)
	if err != nil {
		return fmt.Errorf("matching result to target: %v", err)
	}

	if !matches {
		return fmt.Errorf("verification failed")
	}

	return nil
}
