package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ehrlich-b/cube/internal/cube"
	"github.com/ehrlich-b/cube/internal/cfen"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go \"R U R' U'\"")
		os.Exit(1)
	}

	moves := os.Args[1]
	
	// Generate pattern for the given moves
	pattern, err := generatePattern(moves)
	if err != nil {
		log.Fatalf("Error generating pattern: %v", err)
	}

	fmt.Printf("Moves: %s\n", moves)
	fmt.Printf("Pattern: %s\n", pattern)
	
	// Also show move count
	parsedMoves, err := cube.ParseScramble(moves)
	if err != nil {
		log.Fatalf("Error parsing moves: %v", err)
	}
	
	fmt.Printf("Move Count: %d\n", len(parsedMoves))
}

// generatePattern creates a CFEN pattern by applying algorithm to solved cube
func generatePattern(moves string) (string, error) {
	// Create solved cube in canonical YB orientation
	c := cube.NewCube(3)
	
	// Parse and apply the algorithm moves
	parsedMoves, err := cube.ParseScramble(moves)
	if err != nil {
		return "", err
	}
	
	// Apply moves
	for _, move := range parsedMoves {
		c.ApplyMove(move)
	}
	
	// Get after state as CFEN
	afterCFEN, err := cfen.GenerateCFEN(c)
	if err != nil {
		return "", err
	}
	
	return afterCFEN, nil
}