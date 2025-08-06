package cli

import (
	"fmt"
	"strings"

	"github.com/ehrlich-b/cube/internal/cube"
	"github.com/spf13/cobra"
)

var lookupCmd = &cobra.Command{
	Use:   "lookup [query]",
	Short: "Look up cube algorithms by name, pattern, or category",
	Long: `Look up algorithms in the database by searching names, move sequences,
descriptions, or case numbers. You can also filter by category.

Examples:
  cube lookup sune
  cube lookup "R U R' U'"
  cube lookup --category OLL
  cube lookup "T-Perm"
  cube lookup --pattern "R U R' U'"
  cube lookup --fuzzy "sun"  # fuzzy matches "Sune", "Anti-Sune"`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := ""
		if len(args) > 0 {
			query = args[0]
		}

		pattern, _ := cmd.Flags().GetString("pattern")
		category, _ := cmd.Flags().GetString("category")
		listAll, _ := cmd.Flags().GetBool("all")
		fuzzy, _ := cmd.Flags().GetBool("fuzzy")

		var results []cube.Algorithm

		// Determine lookup method
		if pattern != "" {
			results = cube.LookupByMoves(pattern)
			fmt.Printf("Algorithms matching pattern '%s':\n\n", pattern)
		} else if category != "" {
			results = cube.GetByCategory(category)
			fmt.Printf("Algorithms in category '%s':\n\n", strings.ToUpper(category))
		} else if listAll {
			results = cube.GetAllAlgorithms()
			fmt.Println("All algorithms in database:")
		} else if query != "" {
			if fuzzy {
				results = cube.FuzzyLookupAlgorithm(query)
				fmt.Printf("Fuzzy search results for '%s':\n\n", query)
			} else {
				results = cube.LookupAlgorithm(query)
				fmt.Printf("Algorithms matching '%s':\n\n", query)
			}
		} else {
			fmt.Println("Please provide a query, use --pattern, --category, or --all")
			fmt.Println("\nExample: cube lookup sune")
			fmt.Println("         cube lookup --category OLL")
			fmt.Println("         cube lookup --all")
			return
		}

		// Display results
		if len(results) == 0 {
			fmt.Println("No algorithms found.")
			return
		}

		for i, alg := range results {
			if i > 0 {
				fmt.Println(strings.Repeat("-", 50))
			}

			if alg.CaseID != "" {
				fmt.Printf("%s - %s\n", alg.CaseID, alg.Name)
			} else {
				fmt.Printf("%s (%s)\n", alg.Name, alg.Category)
			}

			fmt.Printf("Moves: %s\n", alg.Moves)
			fmt.Printf("Description: %s\n", alg.Description)

			// Show a preview if color is enabled
			useColor, _ := cmd.Flags().GetBool("color")
			preview, _ := cmd.Flags().GetBool("preview")
			if preview {
				fmt.Println("\nPreview (applied to solved cube):")
				previewAlgorithm(alg.Moves, useColor)
			}
		}

		if len(results) > 1 {
			fmt.Printf("\nFound %d algorithms.\n", len(results))
		}
	},
}

func previewAlgorithm(moves string, useColor bool) {
	c := cube.NewCube(3)
	parsedMoves, err := cube.ParseScramble(moves)
	if err != nil {
		fmt.Printf("Error parsing moves: %v\n", err)
		return
	}

	c.ApplyMoves(parsedMoves)

	// Show only the top face for OLL/PLL preview
	fmt.Println("Top face after algorithm:")
	for row := 0; row < 3; row++ {
		for col := 0; col < 3; col++ {
			color := c.Faces[4][row][col] // Up face
			if useColor {
				fmt.Print(color.ColoredString())
			} else {
				fmt.Print(color.String())
			}
			fmt.Print(" ")
		}
		fmt.Println()
	}
}

func init() {
	lookupCmd.Flags().StringP("pattern", "p", "", "Look up by exact move sequence")
	lookupCmd.Flags().StringP("category", "c", "", "Filter by category (OLL, PLL, F2L)")
	lookupCmd.Flags().BoolP("all", "a", false, "List all algorithms")
	lookupCmd.Flags().Bool("color", false, "Use colored output")
	lookupCmd.Flags().Bool("preview", false, "Show preview of algorithm effect")
	lookupCmd.Flags().BoolP("fuzzy", "f", false, "Use fuzzy string matching for better search")
}
