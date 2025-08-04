package cli

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ehrlich-b/cube/internal/cfen"
	"github.com/ehrlich-b/cube/internal/cube"
	"github.com/spf13/cobra"
)

type AlgorithmMatch struct {
	Algorithm cube.Algorithm
	MatchType string // "exact", "start_pattern", "target_pattern"
	Confidence float64
}

var identifyCmd = &cobra.Command{
	Use:     "identify [cfen-pattern]",
	Short:   "Identify cube patterns and suggest matching algorithms",
	Long:    `Analyze a cube state (in CFEN format) and identify matching OLL/PLL cases or suggest applicable algorithms.`,
	Example: `  cube identify "YB|Y9/R9/B9/W9/O9/G9"  # Solved state
  cube identify "YB|BY5RYG/YO2R6/YBOB6/W9/YG2O6/BR2G6"  # Sune pattern
  cube identify --suggest --category OLL  # Show OLL algorithms for current pattern`,
	Args: cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		suggest, _ := cmd.Flags().GetBool("suggest")
		category, _ := cmd.Flags().GetString("category")
		color, _ := cmd.Flags().GetBool("color")
		
		var pattern string
		if len(args) > 0 {
			pattern = args[0]
		} else if suggest {
			// Use solved state as default for suggestions
			pattern = "YB|Y9/R9/B9/W9/O9/G9"
		} else {
			return fmt.Errorf("CFEN pattern required (or use --suggest)")
		}
		
		// Parse the pattern
		inputState, err := cfen.ParseCFEN(pattern)
		if err != nil {
			return fmt.Errorf("failed to parse CFEN pattern: %w", err)
		}
		
		inputCube, err := inputState.ToCube()
		if err != nil {
			return fmt.Errorf("failed to convert state to cube: %w", err)
		}
		
		// Display the input pattern
		fmt.Println("🔍 ANALYZING PATTERN:")
		output := inputCube.StringWithColor(color)
		fmt.Println(output)
		fmt.Printf("CFEN: %s\n\n", pattern)
		
		// Find matching algorithms
		matches := findMatchingAlgorithms(inputCube, pattern, category)
		
		if len(matches) == 0 {
			fmt.Println("❌ No matching algorithms found in database")
			if suggest {
				fmt.Println("\n💡 SUGGESTIONS:")
				fmt.Println("• Try expanding the algorithm database with more patterns")
				fmt.Println("• Check if your pattern uses standard orientation (Yellow top, Blue front)")
				fmt.Println("• Consider that this might be an intermediate state requiring multiple algorithms")
			}
			return nil
		}
		
		// Sort matches by confidence
		sort.Slice(matches, func(i, j int) bool {
			return matches[i].Confidence > matches[j].Confidence
		})
		
		// Display results
		fmt.Printf("✅ FOUND %d MATCHING ALGORITHMS:\n\n", len(matches))
		
		for i, match := range matches {
			fmt.Printf("[%d] %s (%s)\n", i+1, match.Algorithm.Name, match.Algorithm.CaseID)
			fmt.Printf("    Category: %s\n", match.Algorithm.Category)
			fmt.Printf("    Match Type: %s (%.1f%% confidence)\n", match.MatchType, match.Confidence*100)
			fmt.Printf("    Moves: %s (%d moves)\n", match.Algorithm.Moves, match.Algorithm.MoveCount)
			if match.Algorithm.Description != "" {
				fmt.Printf("    Description: %s\n", match.Algorithm.Description)
			}
			fmt.Println()
		}
		
		// Show suggestions if requested
		if suggest && len(matches) > 0 {
			fmt.Println("💡 RECOMMENDED ACTIONS:")
			topMatch := matches[0]
			
			switch topMatch.MatchType {
			case "exact_start":
				fmt.Printf("• Execute '%s' algorithm: %s\n", topMatch.Algorithm.Name, topMatch.Algorithm.Moves)
				fmt.Printf("• This should solve the %s case you have\n", topMatch.Algorithm.Category)
			case "exact_target":
				fmt.Printf("• This pattern is the target state for '%s'\n", topMatch.Algorithm.Name)
				fmt.Printf("• You may have already solved this case, or need to work backwards\n")
			case "partial_match":
				fmt.Printf("• This partially matches '%s' - check orientation\n", topMatch.Algorithm.Name)
				fmt.Printf("• Try cube rotations: x, y, z before applying algorithm\n")
			}
			
			if len(matches) > 1 {
				fmt.Printf("• %d other algorithms also match - consider efficiency and preference\n", len(matches)-1)
			}
		}
		
		return nil
	},
}

func findMatchingAlgorithms(inputCube *cube.Cube, pattern, categoryFilter string) []AlgorithmMatch {
	var matches []AlgorithmMatch
	
	// Get all algorithms with CFEN patterns
	allAlgorithms := cube.AlgorithmDatabase
	
	for _, alg := range allAlgorithms {
		// Filter by category if specified
		if categoryFilter != "" && !strings.EqualFold(alg.Category, categoryFilter) {
			continue
		}
		
		// TODO: Update verification system for new Pattern field
		// Skip algorithms without patterns
		if alg.Pattern == "" {
			continue
		}
		
		// TODO: Implement pattern matching with new Pattern field
		// For now, add algorithm as potential match if it has a pattern
		if alg.Pattern != "" {
			match := AlgorithmMatch{
				Algorithm:  alg,
				MatchType:  "pattern_available",
				Confidence: 0.5,
			}
			matches = append(matches, match)
		}
	}
	
	return matches
}

func calculatePartialMatch(inputPattern string, alg cube.Algorithm) float64 {
	// TODO: Implement pattern matching with new Pattern field
	// For now, return basic confidence for algorithms with patterns
	
	if alg.Pattern == "" {
		return 0.0
	}
	
	// Basic confidence for having a pattern
	return 0.3
}

func init() {
	identifyCmd.Flags().BoolP("suggest", "s", false, "Show algorithm suggestions and recommendations")
	identifyCmd.Flags().StringP("category", "C", "", "Filter by algorithm category (OLL, PLL, F2L, etc.)")
	identifyCmd.Flags().BoolP("color", "c", false, "Use colored output")
	rootCmd.AddCommand(identifyCmd)
}