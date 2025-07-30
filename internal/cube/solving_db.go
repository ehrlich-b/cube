package cube

import (
	"fmt"
)

// SolvingPattern represents a cube state pattern with Grey wildcards for solving
type SolvingPattern struct {
	// Same structure as Cube.Faces, but Grey means "don't care"
	Faces [6][][]Color
	Focus string // Which face matters most ("up", "down", etc.)
	Name  string // Human readable name for this pattern
}

// Matches checks if a cube matches this pattern
func (p *SolvingPattern) Matches(cube *Cube) bool {
	if cube.Size != len(p.Faces[0]) {
		return false // Pattern size must match cube size
	}

	for face := 0; face < 6; face++ {
		for row := 0; row < cube.Size; row++ {
			for col := 0; col < cube.Size; col++ {
				patternColor := p.Faces[face][row][col]
				cubeColor := cube.Faces[face][row][col]

				// Grey is wildcard - matches anything
				if patternColor == Grey {
					continue
				}

				// Must match exactly
				if patternColor != cubeColor {
					return false
				}
			}
		}
	}

	return true
}

// SolvingAlgorithm represents an algorithm that can be applied when pattern matches
type SolvingAlgorithm struct {
	Name        string          // "4-Look-OLL-Dot", "Sune", etc.
	Description string          // Human-readable description
	Moves       []Move          // Parsed moves ready to apply
	Pattern     *SolvingPattern // When to apply this algorithm
	Priority    int             // Lower = try first (for multiple solutions)
	Phase       string          // "cross", "f2l", "oll", "pll"
}

// SolvingDB represents a database of solving algorithms with pattern matching
type SolvingDB struct {
	algorithms map[string][]SolvingAlgorithm // "oll", "pll", "f2l", "cross"
}

// NewSolvingDB creates a new solving database with your 4-look LL method
func NewSolvingDB() *SolvingDB {
	db := &SolvingDB{
		algorithms: make(map[string][]SolvingAlgorithm),
	}

	// Load your 4-look LL method
	db.load4LookLL()

	return db
}

// load4LookLL adds your 4-look last layer method with proper state detection
func (db *SolvingDB) load4LookLL() {
	// Instead of pattern matching, use custom logic for 4-look LL states
	// This is more reliable than trying to encode every case as a pattern

	// 4-Look OLL Step 1: Make yellow cross if we don't have it
	crossMoves, _ := ParseScramble("F R U R' U' F'")
	db.algorithms["oll"] = append(db.algorithms["oll"], SolvingAlgorithm{
		Name:        "4-Look-OLL-Cross",
		Description: "Make yellow cross on top (handles dot, line, L-shape)",
		Moves:       crossMoves,
		Pattern:     nil, // Custom logic in FindNextMove
		Priority:    1,
		Phase:       "oll",
	})

	// 4-Look OLL Step 2: Orient corners after cross is complete
	suneMoves, _ := ParseScramble("R U R' U R U2 R'")
	db.algorithms["oll"] = append(db.algorithms["oll"], SolvingAlgorithm{
		Name:        "4-Look-OLL-Sune",
		Description: "Orient corners to complete OLL",
		Moves:       suneMoves,
		Pattern:     nil, // Custom logic in FindNextMove
		Priority:    2,
		Phase:       "oll",
	})

	// 4-Look PLL Step 1: Position corners
	tpermMoves, _ := ParseScramble("R U R' F' R U R' U' R' F R2 U' R'")
	db.algorithms["pll"] = append(db.algorithms["pll"], SolvingAlgorithm{
		Name:        "4-Look-PLL-T-Perm",
		Description: "Position corners correctly",
		Moves:       tpermMoves,
		Pattern:     nil, // Custom logic in FindNextMove
		Priority:    1,
		Phase:       "pll",
	})

	// 4-Look PLL Step 2: Position edges
	upermMoves, _ := ParseScramble("R U' R U R U R U' R' U' R2")
	db.algorithms["pll"] = append(db.algorithms["pll"], SolvingAlgorithm{
		Name:        "4-Look-PLL-U-Perm",
		Description: "Cycle edges to solve cube",
		Moves:       upermMoves,
		Pattern:     nil, // Custom logic in FindNextMove
		Priority:    2,
		Phase:       "pll",
	})
}

// FindNextMove finds the best algorithm for current cube state in given phase
func (db *SolvingDB) FindNextMove(cube *Cube, phase string) *SolvingAlgorithm {
	algorithms, exists := db.algorithms[phase]
	if !exists {
		return nil
	}

	// Use custom 4-look LL logic instead of generic pattern matching
	if phase == "oll" {
		return db.findOLLMove(cube, algorithms)
	} else if phase == "pll" {
		return db.findPLLMove(cube, algorithms)
	}

	return nil
}

// findOLLMove determines which OLL algorithm to apply based on cube state
func (db *SolvingDB) findOLLMove(cube *Cube, algorithms []SolvingAlgorithm) *SolvingAlgorithm {
	// Check if we need to make cross first
	if !db.hasCross(cube) {
		// Return cross algorithm (priority 1)
		for _, algo := range algorithms {
			if algo.Priority == 1 { // Cross algorithm
				return &algo
			}
		}
	}

	// Cross exists, check if we need to orient corners
	if db.hasCross(cube) && !db.isOLLComplete(cube) {
		// Return Sune algorithm (priority 2)
		for _, algo := range algorithms {
			if algo.Priority == 2 { // Sune algorithm
				return &algo
			}
		}
	}

	// OLL is complete, no more OLL moves needed
	return nil
}

// findPLLMove determines which PLL algorithm to apply based on cube state
func (db *SolvingDB) findPLLMove(cube *Cube, algorithms []SolvingAlgorithm) *SolvingAlgorithm {
	// Only apply PLL if OLL is complete
	if !db.isOLLComplete(cube) {
		return nil
	}

	// If cube is already solved, no PLL needed
	if cube.IsSolved() {
		return nil
	}

	// For 4-look LL, we try both T-perm and U-perm
	// Since proper case detection is complex, we'll rely on the solver
	// to cycle through them with AUF moves

	// Try T-perm first (priority 1), then U-perm (priority 2)
	for priority := 1; priority <= 2; priority++ {
		for _, algo := range algorithms {
			if algo.Priority == priority {
				return &algo
			}
		}
	}

	return nil
}

// hasCross checks if yellow cross exists on top face
func (db *SolvingDB) hasCross(cube *Cube) bool {
	center := cube.Size / 2

	// Center must be Yellow
	if cube.Faces[Up][center][center] != Yellow {
		return false
	}

	// All 4 edges must be Yellow
	if cube.Faces[Up][0][center] != Yellow { // Top edge
		return false
	}
	if cube.Faces[Up][center][0] != Yellow { // Left edge
		return false
	}
	if cube.Faces[Up][center][cube.Size-1] != Yellow { // Right edge
		return false
	}
	if cube.Faces[Up][cube.Size-1][center] != Yellow { // Bottom edge
		return false
	}

	return true
}

// isOLLComplete checks if entire top face is Yellow
func (db *SolvingDB) isOLLComplete(cube *Cube) bool {
	for row := 0; row < cube.Size; row++ {
		for col := 0; col < cube.Size; col++ {
			if cube.Faces[Up][row][col] != Yellow {
				return false
			}
		}
	}
	return true
}

// GetPhases returns all available solving phases
func (db *SolvingDB) GetPhases() []string {
	phases := make([]string, 0, len(db.algorithms))
	for phase := range db.algorithms {
		phases = append(phases, phase)
	}
	return phases
}

// LoadFromYAML loads solving algorithms from YAML (future expansion hook)
func (db *SolvingDB) LoadFromYAML(yamlContent string) error {
	// TODO: Design YAML format for solving algorithms with patterns
	// Could look like:
	// ---
	// phase: "oll"
	// algorithms:
	//   - name: "Sune"
	//     moves: ["R", "U", "R'", "U", "R", "U2", "R'"]
	//     pattern:
	//       focus: "up"
	//       faces:
	//         up:
	//           - ["Y", ".", "Y"]  # . = Grey wildcard
	//           - [".", "Y", "."]
	//           - [".", "Y", "."]
	return fmt.Errorf("YAML loading not implemented yet - but hook is ready")
}
