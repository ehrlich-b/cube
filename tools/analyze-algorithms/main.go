package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/ehrlich-b/cube/internal/cube"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	command := os.Args[1]
	switch command {
	case "relationships":
		analyzeRelationships()
	case "duplicates":
		findDuplicates()
	case "statistics":
		showStatistics()
	case "validate":
		validateDatabase()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
	}
}

func printUsage() {
	fmt.Println(`Algorithm Database Analyzer

Usage: analyze-algorithms <command>

Commands:
  relationships  Find inverse and mirror relationships
  duplicates     Find potential duplicate algorithms
  statistics     Show database statistics
  validate       Validate database consistency
`)
}

func analyzeRelationships() {
	fmt.Println("Analyzing algorithm relationships...\n")

	algorithms := cube.GetAllAlgorithms()

	// Find potential inverse relationships
	fmt.Println("=== POTENTIAL INVERSE RELATIONSHIPS ===")
	inverseCount := 0

	for i, alg1 := range algorithms {
		for j, alg2 := range algorithms {
			if i >= j {
				continue // Avoid duplicates and self-comparison
			}

			if areInverse(alg1.Moves, alg2.Moves) {
				fmt.Printf("INVERSE PAIR:\n")
				fmt.Printf("  %s (%s): %s\n", alg1.Name, alg1.CaseID, alg1.Moves)
				fmt.Printf("  %s (%s): %s\n", alg2.Name, alg2.CaseID, alg2.Moves)
				fmt.Println()
				inverseCount++
			}
		}
	}

	// Find potential mirror relationships
	fmt.Println("=== POTENTIAL MIRROR RELATIONSHIPS ===")
	mirrorCount := 0

	for i, alg1 := range algorithms {
		for j, alg2 := range algorithms {
			if i >= j {
				continue
			}

			if areMirror(alg1, alg2) {
				fmt.Printf("MIRROR PAIR:\n")
				fmt.Printf("  %s (%s): %s\n", alg1.Name, alg1.CaseID, alg1.Moves)
				fmt.Printf("  %s (%s): %s\n", alg2.Name, alg2.CaseID, alg2.Moves)
				if alg1.Description != "" && alg2.Description != "" {
					fmt.Printf("  Descriptions: '%s' vs '%s'\n", alg1.Description, alg2.Description)
				}
				fmt.Println()
				mirrorCount++
			}
		}
	}

	fmt.Printf("Summary: Found %d inverse pairs and %d mirror pairs\n", inverseCount, mirrorCount)
}

func areInverse(moves1, moves2 string) bool {
	// Parse moves and check if one is the inverse of the other
	parsed1, err1 := cube.ParseScramble(moves1)
	parsed2, err2 := cube.ParseScramble(moves2)

	if err1 != nil || err2 != nil {
		return false
	}

	// Generate inverse of moves1
	inverse1 := generateInverse(parsed1)

	// Check if inverse1 matches parsed2
	if len(inverse1) != len(parsed2) {
		return false
	}

	for i, move := range inverse1 {
		if move != parsed2[i] {
			return false
		}
	}

	return true
}

func areMirror(alg1, alg2 cube.Algorithm) bool {
	// Check if algorithms are mirrors based on:
	// 1. Same category
	// 2. Similar move count
	// 3. Names suggesting mirroring (e.g., "A" vs "B", "Left" vs "Right")
	// 4. Similar descriptions

	if alg1.Category != alg2.Category {
		return false
	}

	// Move count should be similar (within 2 moves)
	if abs(alg1.MoveCount-alg2.MoveCount) > 2 {
		return false
	}

	// Check for name patterns suggesting mirrors
	name1 := strings.ToLower(alg1.Name)
	name2 := strings.ToLower(alg2.Name)

	// Common mirror patterns
	mirrorPatterns := [][]string{
		{"sune", "anti-sune"},
		{"antisune", "sune"},
		{" a", " b"},
		{"-a", "-b"},
		{"left", "right"},
		{"clockwise", "counterclockwise"},
		{"cw", "ccw"},
	}

	for _, pattern := range mirrorPatterns {
		if (strings.Contains(name1, pattern[0]) && strings.Contains(name2, pattern[1])) ||
			(strings.Contains(name1, pattern[1]) && strings.Contains(name2, pattern[0])) {
			return true
		}
	}

	// Check case ID patterns
	case1 := strings.ToLower(alg1.CaseID)
	case2 := strings.ToLower(alg2.CaseID)

	// Pattern like "OLL-1" vs "OLL-2" or "PLL-Aa" vs "PLL-Ab"
	if strings.Contains(case1, "a") && strings.Contains(case2, "b") {
		base1 := strings.ReplaceAll(case1, "a", "")
		base2 := strings.ReplaceAll(case2, "b", "")
		if base1 == base2 {
			return true
		}
	}

	return false
}

func findDuplicates() {
	fmt.Println("Finding potential duplicate algorithms...\n")

	algorithms := cube.GetAllAlgorithms()

	// Group by normalized moves
	moveGroups := make(map[string][]cube.Algorithm)

	for _, alg := range algorithms {
		normalizedMoves := normalizeMoves(alg.Moves)
		moveGroups[normalizedMoves] = append(moveGroups[normalizedMoves], alg)
	}

	duplicateCount := 0
	for moves, group := range moveGroups {
		if len(group) > 1 {
			fmt.Printf("DUPLICATE MOVES: %s\n", moves)
			for _, alg := range group {
				fmt.Printf("  %s (%s) - %s: %s\n", alg.Name, alg.CaseID, alg.Category, alg.Description)
			}
			fmt.Println()
			duplicateCount++
		}
	}

	fmt.Printf("Found %d sets of algorithms with identical moves\n", duplicateCount)
}

func normalizeMoves(moves string) string {
	// Remove all whitespace and convert to lowercase for comparison
	return strings.ReplaceAll(strings.ToLower(moves), " ", "")
}

func showStatistics() {
	fmt.Println("Database Statistics\n")

	algorithms := cube.GetAllAlgorithms()

	// Count by category
	categoryCount := make(map[string]int)
	totalMoves := 0

	for _, alg := range algorithms {
		categoryCount[alg.Category]++
		totalMoves += alg.MoveCount
	}

	fmt.Printf("Total algorithms: %d\n", len(algorithms))
	fmt.Printf("Average moves per algorithm: %.1f\n", float64(totalMoves)/float64(len(algorithms)))
	fmt.Println()

	fmt.Println("Algorithms by category:")

	// Sort categories by count
	type categoryInfo struct {
		name  string
		count int
	}

	var categories []categoryInfo
	for cat, count := range categoryCount {
		categories = append(categories, categoryInfo{cat, count})
	}

	sort.Slice(categories, func(i, j int) bool {
		return categories[i].count > categories[j].count
	})

	for _, cat := range categories {
		fmt.Printf("  %-15s: %d\n", cat.name, cat.count)
	}

	// Move count distribution
	fmt.Println("\nMove count distribution:")
	moveCountDist := make(map[int]int)
	for _, alg := range algorithms {
		moveCountDist[alg.MoveCount]++
	}

	var moveCounts []int
	for count := range moveCountDist {
		moveCounts = append(moveCounts, count)
	}
	sort.Ints(moveCounts)

	for _, count := range moveCounts {
		fmt.Printf("  %2d moves: %d algorithms\n", count, moveCountDist[count])
	}
}

func validateDatabase() {
	fmt.Println("Validating database consistency...\n")

	algorithms := cube.GetAllAlgorithms()
	issues := 0

	fmt.Println("=== VALIDATION ISSUES ===")

	for i, alg := range algorithms {
		// Check for empty required fields
		if alg.Name == "" {
			fmt.Printf("Algorithm %d: Missing name\n", i+1)
			issues++
		}

		if alg.CaseID == "" {
			fmt.Printf("Algorithm '%s': Missing case ID\n", alg.Name)
			issues++
		}

		if alg.Category == "" {
			fmt.Printf("Algorithm '%s': Missing category\n", alg.Name)
			issues++
		}

		if alg.Moves == "" {
			fmt.Printf("Algorithm '%s': Missing moves\n", alg.Name)
			issues++
		}

		// Check if moves parse correctly
		if alg.Moves != "" {
			_, err := cube.ParseScramble(alg.Moves)
			if err != nil {
				fmt.Printf("Algorithm '%s': Invalid moves '%s': %v\n", alg.Name, alg.Moves, err)
				issues++
			}
		}

		// Check move count consistency
		if alg.Moves != "" {
			moves, err := cube.ParseScramble(alg.Moves)
			if err == nil && len(moves) != alg.MoveCount {
				fmt.Printf("Algorithm '%s': Move count mismatch (stated: %d, actual: %d)\n",
					alg.Name, alg.MoveCount, len(moves))
				issues++
			}
		}
	}

	if issues == 0 {
		fmt.Println("✅ Database validation passed - no issues found!")
	} else {
		fmt.Printf("❌ Found %d validation issues\n", issues)
	}

	fmt.Printf("\nDatabase summary: %d algorithms validated\n", len(algorithms))
}

func generateInverse(moves []cube.Move) []cube.Move {
	// Reverse the order and invert each move
	var inverse []cube.Move
	for i := len(moves) - 1; i >= 0; i-- {
		move := moves[i]
		// Invert the move
		invertedMove := cube.Move{
			Face:      move.Face,
			Clockwise: !move.Clockwise,
			Double:    move.Double,
			Wide:      move.Wide,
			WideDepth: move.WideDepth,
			Layer:     move.Layer,
			Slice:     move.Slice,
			Rotation:  move.Rotation,
		}
		// Handle double moves (they are their own inverse)
		if move.Double {
			invertedMove.Clockwise = move.Clockwise
		}
		inverse = append(inverse, invertedMove)
	}
	return inverse
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
