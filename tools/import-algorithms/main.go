package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/ehrlich-b/cube/internal/cube"
)

// CSVRecord represents the format of algorithm CSV files
type CSVRecord struct {
	CaseID      string
	Name        string
	Category    string
	Moves       string
	Description string
	Recognition string
	Reference   string
}

// ImportConfig controls the import process
type ImportConfig struct {
	InputDir         string
	OutputFile       string
	Verbose          bool
	DryRun           bool
	SkipDuplicates   bool
	GeneratePatterns bool
}

func main() {
	config := ImportConfig{
		InputDir:         "../../alg_dumps",
		OutputFile:       "../../internal/cube/algorithms_imported.go",
		Verbose:          true,
		DryRun:           false,
		SkipDuplicates:   true,
		GeneratePatterns: true,
	}

	// Parse command line arguments
	for i := 1; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "--input":
			if i+1 < len(os.Args) {
				config.InputDir = os.Args[i+1]
				i++
			}
		case "--output":
			if i+1 < len(os.Args) {
				config.OutputFile = os.Args[i+1]
				i++
			}
		case "--dry-run":
			config.DryRun = true
		case "--no-patterns":
			config.GeneratePatterns = false
		case "--allow-duplicates":
			config.SkipDuplicates = false
		case "--quiet":
			config.Verbose = false
		case "--help":
			printUsage()
			return
		}
	}

	if err := runImport(config); err != nil {
		log.Fatalf("Import failed: %v", err)
	}
}

func printUsage() {
	fmt.Println(`Algorithm CSV Import Tool

Usage: import-algorithms [options]

Options:
  --input DIR        Directory containing CSV files (default: ../../alg_dumps)
  --output FILE      Output Go file (default: ../../internal/cube/algorithms_imported.go)
  --dry-run          Show what would be imported without writing files
  --no-patterns      Skip pattern generation
  --allow-duplicates Allow duplicate algorithms
  --quiet            Reduce output verbosity
  --help             Show this help message`)
}

func runImport(config ImportConfig) error {
	// Find all CSV files in input directory
	csvFiles, err := findCSVFiles(config.InputDir)
	if err != nil {
		return fmt.Errorf("finding CSV files: %w", err)
	}

	if config.Verbose {
		fmt.Printf("Found %d CSV files to import\n", len(csvFiles))
	}

	// Import all algorithms
	var allAlgorithms []cube.Algorithm
	duplicateCount := 0

	for _, csvFile := range csvFiles {
		algorithms, dupes, err := importCSVFile(csvFile, config)
		if err != nil {
			if config.Verbose {
				fmt.Printf("Warning: failed to import %s: %v\n", csvFile, err)
			}
			continue
		}

		allAlgorithms = append(allAlgorithms, algorithms...)
		duplicateCount += dupes

		if config.Verbose {
			fmt.Printf("Imported %d algorithms from %s\n", len(algorithms), filepath.Base(csvFile))
		}
	}

	// Remove duplicates if requested
	if config.SkipDuplicates {
		allAlgorithms = removeDuplicates(allAlgorithms)
	}

	// Generate patterns if requested
	if config.GeneratePatterns {
		if config.Verbose {
			fmt.Printf("Generating patterns for %d algorithms...\n", len(allAlgorithms))
		}

		for i := range allAlgorithms {
			pattern, err := generateAlgorithmPattern(&allAlgorithms[i])
			if err != nil {
				if config.Verbose {
					fmt.Printf("Warning: failed to generate pattern for %s: %v\n", allAlgorithms[i].Name, err)
				}
				continue
			}
			allAlgorithms[i].Pattern = pattern
		}
	}

	// Sort algorithms by category and case ID
	sort.Slice(allAlgorithms, func(i, j int) bool {
		if allAlgorithms[i].Category != allAlgorithms[j].Category {
			return allAlgorithms[i].Category < allAlgorithms[j].Category
		}
		return allAlgorithms[i].CaseID < allAlgorithms[j].CaseID
	})

	if config.Verbose {
		fmt.Printf("\nImport Summary:\n")
		fmt.Printf("  Total algorithms: %d\n", len(allAlgorithms))
		fmt.Printf("  Duplicates skipped: %d\n", duplicateCount)
		categoryCounts := make(map[string]int)
		for _, alg := range allAlgorithms {
			categoryCounts[alg.Category]++
		}
		fmt.Printf("  By category:\n")
		for category, count := range categoryCounts {
			fmt.Printf("    %s: %d\n", category, count)
		}
	}

	if config.DryRun {
		fmt.Printf("\nDry run complete - would write %d algorithms to %s\n", len(allAlgorithms), config.OutputFile)
		return nil
	}

	// Write the algorithms to Go file
	if err := writeAlgorithmsFile(allAlgorithms, config.OutputFile); err != nil {
		return fmt.Errorf("writing algorithms file: %w", err)
	}

	if config.Verbose {
		fmt.Printf("Successfully wrote %d algorithms to %s\n", len(allAlgorithms), config.OutputFile)
	}

	return nil
}

func findCSVFiles(inputDir string) ([]string, error) {
	var csvFiles []string

	err := filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(strings.ToLower(path), ".csv") {
			csvFiles = append(csvFiles, path)
		}

		return nil
	})

	return csvFiles, err
}

func importCSVFile(filename string, config ImportConfig) ([]cube.Algorithm, int, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, 0, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true

	var algorithms []cube.Algorithm
	duplicateCount := 0
	lineNum := 0

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, 0, fmt.Errorf("reading CSV line %d: %w", lineNum, err)
		}

		lineNum++

		// Skip empty records or records with insufficient columns
		if len(record) < 6 {
			continue
		}

		csvRecord := CSVRecord{
			CaseID:      strings.TrimSpace(record[0]),
			Name:        strings.TrimSpace(record[1]),
			Category:    strings.TrimSpace(record[2]),
			Moves:       strings.TrimSpace(record[3]),
			Description: strings.TrimSpace(record[4]),
			Recognition: strings.TrimSpace(record[5]),
		}

		if len(record) > 6 {
			csvRecord.Reference = strings.TrimSpace(record[6])
		}

		// Skip empty or invalid records
		if csvRecord.CaseID == "" || csvRecord.Moves == "" {
			continue
		}

		algorithm, err := convertCSVToAlgorithm(csvRecord)
		if err != nil {
			if config.Verbose {
				fmt.Printf("Warning: skipping invalid algorithm at line %d: %v\n", lineNum, err)
			}
			continue
		}

		algorithms = append(algorithms, *algorithm)
	}

	return algorithms, duplicateCount, nil
}

func convertCSVToAlgorithm(record CSVRecord) (*cube.Algorithm, error) {
	// Skip algorithms with obviously invalid notation
	if shouldSkipAlgorithm(record.Moves) {
		return nil, fmt.Errorf("algorithm contains unsupported notation: %s", record.Moves)
	}

	// Clean up the algorithm moves - standardize notation
	normalizedMoves := normalizeAlgorithmMoves(record.Moves)

	// Parse and validate moves
	moves, err := cube.ParseScramble(normalizedMoves)
	if err != nil {
		return nil, fmt.Errorf("invalid moves '%s' (normalized from '%s'): %w", normalizedMoves, record.Moves, err)
	}

	algorithm := &cube.Algorithm{
		Name:        record.Name,
		CaseID:      record.CaseID,
		Category:    record.Category,
		Moves:       normalizedMoves,
		MoveCount:   len(moves),
		Description: record.Description,
		Recognition: record.Recognition,
		Pattern:     "", // Will be generated if requested
	}

	return algorithm, nil
}

func shouldSkipAlgorithm(moves string) bool {
	// Skip algorithms with obviously invalid or unsupported notation
	invalidPatterns := []string{
		"^$",           // Empty moves
		"no algorithm", // Placeholder text
		"similar to",   // Reference text
		"apply.*alg",   // Placeholder text
		"turn cube",    // Non-standard instruction
		"U2\\)",        // Malformed parentheses
		"\\^",          // Exponent notation not supported
	}

	for _, pattern := range invalidPatterns {
		matched, _ := regexp.MatchString("(?i)"+pattern, moves)
		if matched {
			return true
		}
	}

	return false
}

func normalizeAlgorithmMoves(moves string) string {
	// Remove extra whitespace and normalize move notation
	moves = regexp.MustCompile(`\s+`).ReplaceAllString(strings.TrimSpace(moves), " ")

	// Remove parentheses that are just for grouping (but preserve cube rotations)
	moves = strings.ReplaceAll(moves, "(", "")
	moves = strings.ReplaceAll(moves, ")", "")

	// Handle cube rotations: (x), (y), (z) -> x, y, z
	moves = regexp.MustCompile(`\b([xyz])\b`).ReplaceAllString(moves, "$1")
	moves = regexp.MustCompile(`\b([xyz]')\b`).ReplaceAllString(moves, "$1")
	moves = regexp.MustCompile(`\b([xyz]2)\b`).ReplaceAllString(moves, "$1")

	// Convert lowercase wide moves to uppercase with 'w' suffix
	// r -> Rw, f -> Fw, etc.
	lowerCasePattern := regexp.MustCompile(`\b([rflbud])([2']?)\b`)
	moves = lowerCasePattern.ReplaceAllStringFunc(moves, func(match string) string {
		parts := lowerCasePattern.FindStringSubmatch(match)
		face := strings.ToUpper(parts[1])
		modifier := parts[2]
		if modifier == "" {
			return face + "w"
		}
		return face + "w" + modifier
	})

	// Convert M, E, S slice moves (they're already uppercase but need to be handled properly)
	// No changes needed for these as they should already be supported

	// Normalize some common variations
	moves = strings.ReplaceAll(moves, "2'", "'2") // R2' -> R'2 (not standard but sometimes seen)

	// Handle specific notations that might be problematic
	moves = strings.ReplaceAll(moves, " '", "'") // Fix spacing issues like "R '"

	return moves
}

func removeDuplicates(algorithms []cube.Algorithm) []cube.Algorithm {
	seen := make(map[string]bool)
	var unique []cube.Algorithm

	for _, alg := range algorithms {
		// Create a key based on normalized moves
		key := strings.ToLower(strings.ReplaceAll(alg.Moves, " ", ""))

		if !seen[key] {
			seen[key] = true
			unique = append(unique, alg)
		}
	}

	return unique
}

func generateAlgorithmPattern(algorithm *cube.Algorithm) (string, error) {
	// Create a solved YB cube
	c := cube.NewCube(3)

	// Apply the algorithm
	moves, err := cube.ParseScramble(algorithm.Moves)
	if err != nil {
		return "", err
	}

	for _, move := range moves {
		c.ApplyMove(move)
	}

	// Generate CFEN pattern (this would need to be implemented)
	// For now, return empty string - patterns will be generated separately
	return "", nil
}

func writeAlgorithmsFile(algorithms []cube.Algorithm, filename string) error {
	// Create the directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		return err
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write the Go file header
	fmt.Fprintf(file, `package cube

// This file is auto-generated by tools/import-algorithms
// DO NOT EDIT MANUALLY

// ImportedAlgorithms contains algorithms imported from CSV dumps
var ImportedAlgorithms = []Algorithm{
`)

	// Write each algorithm
	for _, alg := range algorithms {
		fmt.Fprintf(file, "\t{\n")
		fmt.Fprintf(file, "\t\tName:        %s,\n", strconv.Quote(alg.Name))
		fmt.Fprintf(file, "\t\tCaseID:      %s,\n", strconv.Quote(alg.CaseID))
		fmt.Fprintf(file, "\t\tCategory:    %s,\n", strconv.Quote(alg.Category))
		fmt.Fprintf(file, "\t\tMoves:       %s,\n", strconv.Quote(alg.Moves))
		fmt.Fprintf(file, "\t\tMoveCount:   %d,\n", alg.MoveCount)
		fmt.Fprintf(file, "\t\tDescription: %s,\n", strconv.Quote(alg.Description))
		fmt.Fprintf(file, "\t\tRecognition: %s,\n", strconv.Quote(alg.Recognition))
		if alg.Pattern != "" {
			fmt.Fprintf(file, "\t\tPattern:     %s,\n", strconv.Quote(alg.Pattern))
		}
		fmt.Fprintf(file, "\t},\n")
	}

	fmt.Fprintf(file, "}\n")

	return nil
}
