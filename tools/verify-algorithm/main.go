package main

import (
	"fmt"
	"os"

	"github.com/ehrlich-b/cube/internal/cfen"
	"github.com/ehrlich-b/cube/internal/cube"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: verify-algorithm <algorithm-name> [--verbose]")
		fmt.Println("       verify-algorithm --list")
		os.Exit(1)
	}

	if os.Args[1] == "--list" {
		listAlgorithms()
		return
	}

	algorithmName := os.Args[1]
	verbose := len(os.Args) > 2 && os.Args[2] == "--verbose"

	// Look up algorithm by name
	algorithms := cube.LookupAlgorithm(algorithmName)
	if len(algorithms) == 0 {
		fmt.Printf("Error: algorithm '%s' not found\n", algorithmName)
		os.Exit(1)
	}

	algorithm := algorithms[0]

	// Check if algorithm has pattern
	if algorithm.Pattern == "" {
		fmt.Printf("Error: algorithm '%s' has no pattern defined\n", algorithm.Name)
		os.Exit(1)
	}

	// Update move count
	algorithm.UpdateMoveCount()

	// Use solved cube as start state for pattern verification
	startCFEN := "YB|Y9/R9/B9/W9/O9/G9" // Solved cube
	targetCFEN := algorithm.Pattern // Expected pattern after applying algorithm

	// Perform verification
	err := verifyAlgorithm(algorithm, startCFEN, targetCFEN, verbose)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func listAlgorithms() {
	algorithms := cube.AlgorithmDatabase
	verifiableCount := 0

	fmt.Println("Algorithms with CFEN patterns for verification:")
	fmt.Println()

	for _, alg := range algorithms {
		if alg.Pattern != "" {
			verifiableCount++
			status := "✅ HAS PATTERN"

			fmt.Printf("%s (%s) - %s\n", alg.Name, alg.CaseID, status)
			fmt.Printf("  Moves: %s\n", alg.Moves)
			fmt.Printf("  Pattern: %s\n", alg.Pattern)
			if alg.Recognition != "" {
				fmt.Printf("  Recognition: %s\n", alg.Recognition)
			}
			fmt.Println()
		}
	}

	fmt.Printf("Total algorithms with CFEN patterns: %d\n", verifiableCount)
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

	if verbose {
		fmt.Printf("Algorithm: %s (%s)\n", algorithm.Name, algorithm.CaseID)
		fmt.Printf("Moves: %s\n", algorithm.Moves)
		fmt.Printf("Move count: %d\n", algorithm.MoveCount)
		fmt.Printf("\nStart state:\n")
		fmt.Println(c.UnfoldedString(false, false))
	}

	// Parse and apply algorithm
	moves, err := cube.ParseScramble(algorithm.Moves)
	if err != nil {
		return fmt.Errorf("parsing algorithm moves: %v", err)
	}

	c.ApplyMoves(moves)

	if verbose {
		fmt.Printf("\nAfter algorithm:\n")
		fmt.Println(c.UnfoldedString(false, false))
	}

	// Check if result matches target
	matches, err := targetState.MatchesCube(c)
	if err != nil {
		return fmt.Errorf("matching result to target: %v", err)
	}

	if matches {
		fmt.Printf("✅ PASS: %s correctly transforms start to target state\n", algorithm.Name)
		if verbose {
			fmt.Printf("Start:  %s\n", startCFEN)
			fmt.Printf("Target: %s\n", targetCFEN)
			actualCFEN, _ := cfen.GenerateCFEN(c)
			fmt.Printf("Actual: %s\n", actualCFEN)
		}
		return nil
	} else {
		fmt.Printf("❌ FAIL: %s does not achieve target state\n", algorithm.Name)
		if verbose {
			fmt.Printf("Start:  %s\n", startCFEN)
			fmt.Printf("Target: %s\n", targetCFEN)
			actualCFEN, _ := cfen.GenerateCFEN(c)
			fmt.Printf("Actual: %s\n", actualCFEN)
		}
		return fmt.Errorf("verification failed")
	}
}
