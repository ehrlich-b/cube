package cli

import (
	"fmt"
	"sort"

	"github.com/ehrlich-b/cube/internal/cube"
	"github.com/spf13/cobra"
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze [scramble]",
	Short: "Analyze cube state and identify patterns",
	Long: `Analyze displays detailed information about the current cube state including:
- Piece positions and orientations
- Pattern recognition (white cross, F2L, OLL, PLL)  
- Solving progress and next steps
- Piece tracking information

Examples:
  cube analyze ""                    # Analyze solved cube
  cube analyze "R U R' U'"          # Analyze after scramble
  cube analyze "F R U R' U' F'"     # Analyze OLL algorithm result`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		scramble := ""
		if len(args) > 0 {
			scramble = args[0]
		}

		dimension, _ := cmd.Flags().GetInt("dimension")
		verbose, _ := cmd.Flags().GetBool("verbose")
		pieces, _ := cmd.Flags().GetBool("pieces")

		// Create cube
		c := cube.NewCube(dimension)

		// Apply scramble if provided
		if scramble != "" {
			moves, err := cube.ParseScramble(scramble)
			if err != nil {
				return fmt.Errorf("failed to parse scramble: %w", err)
			}
			c.ApplyMoves(moves)
			fmt.Printf("🔍 Analyzing cube after scramble: %s\n\n", scramble)
		} else {
			fmt.Println("🔍 Analyzing solved cube state:")
		}

		// Show basic cube state
		if verbose {
			fmt.Println("Cube state:")
			fmt.Println(c.UnfoldedString(true, true))
			fmt.Println()
		}

		// Pattern analysis (only for 3x3)
		if dimension == 3 {
			fmt.Println("📊 PATTERN ANALYSIS:")
			patterns := cube.AnalyzeCubeState(c)
			
			if len(patterns) == 0 {
				fmt.Println("No recognized patterns found")
			} else {
				// Sort patterns by completion percentage
				type patternResult struct {
					name       string
					completion float64
				}
				
				var sortedPatterns []patternResult
				for name, completion := range patterns {
					sortedPatterns = append(sortedPatterns, patternResult{name, completion})
				}
				
				sort.Slice(sortedPatterns, func(i, j int) bool {
					return sortedPatterns[i].completion > sortedPatterns[j].completion
				})
				
				for _, p := range sortedPatterns {
					status := "🔄"
					if p.completion == 100.0 {
						status = "✅"
					} else if p.completion >= 50.0 {
						status = "🟡"
					}
					
					fmt.Printf("%s %s: %.1f%% complete\n", status, p.name, p.completion)
				}
			}
			
			// Next step suggestion
			fmt.Println("\n🎯 NEXT STEP:")
			nextStep := cube.GetNextStep(c)
			fmt.Printf("→ %s\n", nextStep)
		}

		// Piece analysis (only for 3x3)
		if pieces && dimension == 3 {
			fmt.Println("\n🧩 PIECE ANALYSIS:")
			
			// Analyze edges
			fmt.Println("\nEdge pieces:")
			edges := c.GetAllEdges()
			for i, edge := range edges {
				if edge != nil && len(edge.Colors) >= 2 {
					correctPos := c.IsPieceInCorrectPosition(edge.Colors)
					correctOri := c.IsPieceCorrectlyOriented(edge.Colors)
					
					status := "❌"
					if correctPos && correctOri {
						status = "✅"
					} else if correctPos {
						status = "🔄" // Right position, wrong orientation
					}
					
					fmt.Printf("  %s Edge %d: %s-%s at %s,%d,%d\n", 
						status, i+1, 
						edge.Colors[0], edge.Colors[1],
						edge.Position.Face, edge.Position.Row, edge.Position.Col)
				}
			}
			
			// Analyze corners
			fmt.Println("\nCorner pieces:")
			corners := c.GetAllCorners()
			for i, corner := range corners {
				if corner != nil && len(corner.Colors) >= 3 {
					correctPos := c.IsPieceInCorrectPosition(corner.Colors)
					correctOri := c.IsPieceCorrectlyOriented(corner.Colors)
					
					status := "❌"
					if correctPos && correctOri {
						status = "✅"
					} else if correctPos {
						status = "🔄" // Right position, wrong orientation
					}
					
					fmt.Printf("  %s Corner %d: %s-%s-%s at %s,%d,%d\n", 
						status, i+1,
						corner.Colors[0], corner.Colors[1], corner.Colors[2],
						corner.Position.Face, corner.Position.Row, corner.Position.Col)
				}
			}
			
			// Analyze centers
			fmt.Println("\nCenter pieces:")
			centers := c.GetAllCenters()
			for i, center := range centers {
				if center != nil {
					fmt.Printf("  ✅ Center %d: %s at %s,%d,%d\n", 
						i+1, center.Colors[0],
						center.Position.Face, center.Position.Row, center.Position.Col)
				}
			}
		}

		return nil
	},
}

func init() {
	analyzeCmd.Flags().IntP("dimension", "d", 3, "Cube dimension (NxNxN)")
	analyzeCmd.Flags().BoolP("verbose", "v", false, "Show detailed cube state")
	analyzeCmd.Flags().BoolP("pieces", "p", false, "Show detailed piece analysis")
	rootCmd.AddCommand(analyzeCmd)
}