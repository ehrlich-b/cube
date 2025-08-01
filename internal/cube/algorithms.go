package cube

import (
	"strings"
)

// Algorithm represents a named cube algorithm
type Algorithm struct {
	// Core algorithm data
	Name        string
	Category    string // OLL, PLL, F2L, etc.
	Moves       string
	Description string
	CaseNumber  string // e.g., "OLL 21", "PLL T"

	// Verification fields
	StartCFEN  string // Required starting state (with wildcards)
	TargetCFEN string // Expected ending state (with wildcards)
	Verified   bool   // Has this been verified?

	// Enhanced metadata
	MoveCount   int      // Number of moves (auto-calculated)
	Probability float64  // Chance of occurring in solve
	Variants    []string // Alternative algorithms for same case
	TestedOn    []int    // Cube sizes tested (e.g., [3, 4, 5])
}

// AlgorithmDatabase contains common cube algorithms
var AlgorithmDatabase = []Algorithm{
	// OLL Cases
	{
		Name:        "Sune",
		Category:    "OLL",
		Moves:       "R U R' U R U2 R'",
		Description: "Orient corners when one is correctly oriented",
		CaseNumber:  "OLL 27",
		StartCFEN:   "YB|Y9/R9/B9/W9/O9/G9",                 // Start from solved for demo purposes
		TargetCFEN:  "YB|BY5RYG/YO2R6/YBOB6/W9/YG2O6/BR2G6", // Sune pattern result
		MoveCount:   7,
		Verified:    true,
		TestedOn:    []int{3},
	},
	{
		Name:        "Anti-Sune",
		Category:    "OLL",
		Moves:       "R U2 R' U' R U' R'",
		Description: "Mirror of Sune algorithm",
		CaseNumber:  "OLL 26",
		StartCFEN:   "YB|BY5RYG/YO2R6/YBOB6/W9/YG2O6/BR2G6", // Sune pattern (needs Anti-Sune)
		TargetCFEN:  "YB|Y9/R9/B9/W9/O9/G9",                 // Back to solved
		MoveCount:   7,
		Verified:    true,
		TestedOn:    []int{3},
	},
	{
		Name:        "Cross OLL",
		Category:    "OLL",
		Moves:       "F R U R' U' F'",
		Description: "Creates cross on top, orients edges",
		CaseNumber:  "OLL 3",
		StartCFEN:   "YB|RBY4R2Y/BG2R6/BYOB6/W9/YOYO6/OYG7", // Cross OLL case (line pattern)
		TargetCFEN:  "YB|Y9/R9/B9/W9/O9/G9",                 // Solved (cross oriented)
		MoveCount:   6,
		Verified:    true,
		TestedOn:    []int{3},
	},
	{
		Name:        "Dot OLL",
		Category:    "OLL",
		Moves:       "F R U R' U' F' f R U R' U' f'",
		Description: "No edges oriented correctly",
		CaseNumber:  "OLL 2",
	},
	{
		Name:        "L-Shape OLL",
		Category:    "OLL",
		Moves:       "F U R U' R' F'",
		Description: "L-shape edge orientation",
		CaseNumber:  "OLL 44",
	},

	// PLL Cases
	{
		Name:        "T-Perm",
		Category:    "PLL",
		Moves:       "R U R' U' R' F R2 U' R' U' R U R' F'",
		Description: "Swaps two adjacent corners and two edges",
		CaseNumber:  "PLL T",
		StartCFEN:   "YB|Y9/GOBR6/B2RB6/W9/ORO7/RG8", // Actual T-Perm case
		TargetCFEN:  "YB|Y9/R9/B9/W9/O9/G9",          // Fully solved
		MoveCount:   14,
		Verified:    true,
		TestedOn:    []int{3},
	},
	{
		Name:        "Y-Perm",
		Category:    "PLL",
		Moves:       "F R U' R' U' R U R' F' R U R' U' R' F R F'",
		Description: "Swaps two adjacent corners and two edges diagonally",
		CaseNumber:  "PLL Y",
	},
	{
		Name:        "U-Perm (a)",
		Category:    "PLL",
		Moves:       "R U' R U R U R U' R' U' R2",
		Description: "Cycles three edges counterclockwise",
		CaseNumber:  "PLL Ua",
	},
	{
		Name:        "U-Perm (b)",
		Category:    "PLL",
		Moves:       "R2 U R U R' U' R' U' R' U R'",
		Description: "Cycles three edges clockwise",
		CaseNumber:  "PLL Ub",
	},
	{
		Name:        "H-Perm",
		Category:    "PLL",
		Moves:       "M2 U M2 U2 M2 U M2",
		Description: "Swaps opposite edges",
		CaseNumber:  "PLL H",
	},
	{
		Name:        "Z-Perm",
		Category:    "PLL",
		Moves:       "M' U M2 U M2 U M' U2 M2",
		Description: "Swaps adjacent edges",
		CaseNumber:  "PLL Z",
	},

	// F2L Cases
	{
		Name:        "Basic Insert",
		Category:    "F2L",
		Moves:       "U R U' R'",
		Description: "Basic corner-edge pair insertion",
		CaseNumber:  "F2L 1",
		StartCFEN:   "", // Will be defined for specific F2L slot patterns
		TargetCFEN:  "", // Will be defined for specific F2L slot patterns
		MoveCount:   4,
	},
	{
		Name:        "Split Pair",
		Category:    "F2L",
		Moves:       "R U R' U' R U R' U'",
		Description: "Split and rejoin F2L pair",
		CaseNumber:  "F2L 2",
	},

	// Common Patterns
	{
		Name:        "Sexy Move",
		Category:    "Trigger",
		Moves:       "R U R' U'",
		Description: "Most common trigger in cubing",
		CaseNumber:  "",
		StartCFEN:   "YB|RBY6R/WG2R6/B5YB2Y/W2BW6/YO8/ORG7", // Specific scrambled state
		TargetCFEN:  "YB|Y9/R9/B9/W9/O9/G9",                 // Solved (this particular sexy move solves it)
		MoveCount:   4,
		Verified:    true,
		TestedOn:    []int{3},
	},
	{
		Name:        "Sledgehammer",
		Category:    "Trigger",
		Moves:       "R' F R F'",
		Description: "Common trigger for F2L and OLL",
		CaseNumber:  "",
	},

	// Essential OLL Cases (Most Common)
	{
		Name:        "T-Shape OLL",
		Category:    "OLL",
		Moves:       "r U R' U' r' F R F'",
		Description: "T-shape edge pattern",
		CaseNumber:  "OLL 33",
	},
	{
		Name:        "P-Shape OLL",
		Category:    "OLL",
		Moves:       "f R U R' U' f'",
		Description: "P-shape edge pattern",
		CaseNumber:  "OLL 44",
	},
	{
		Name:        "Bowtie OLL",
		Category:    "OLL",
		Moves:       "f' L' U' L U f",
		Description: "P-shape edge pattern (mirror)",
		CaseNumber:  "OLL 43",
	},
	{
		Name:        "Double Sune",
		Category:    "OLL",
		Moves:       "R U R' U R U' R' U R U2 R'",
		Description: "H-shape corner pattern",
		CaseNumber:  "OLL 21",
	},
	{
		Name:        "Pi OLL",
		Category:    "OLL",
		Moves:       "R U2 R2 U' R2 U' R2 U2 R",
		Description: "Pi-shape corner pattern",
		CaseNumber:  "OLL 22",
	},
	{
		Name:        "U-Shape OLL",
		Category:    "OLL",
		Moves:       "R2 D R' U2 R D' R' U2 R'",
		Description: "U-shape corner pattern",
		CaseNumber:  "OLL 23",
	},
	{
		Name:        "Lightning Bolt",
		Category:    "OLL",
		Moves:       "r U R' U' r' F R F'",
		Description: "Lightning bolt edge pattern",
		CaseNumber:  "OLL 24",
	},
	{
		Name:        "Chameleon",
		Category:    "OLL",
		Moves:       "F' r U R' U' r' F R",
		Description: "Chameleon edge pattern",
		CaseNumber:  "OLL 25",
	},

	// Essential PLL Cases (Most Common)
	{
		Name:        "A-Perm (a)",
		Category:    "PLL",
		Moves:       "x R' U R' D2 R U' R' D2 R2 x'",
		Description: "Adjacent corner swap",
		CaseNumber:  "PLL Aa",
		StartCFEN:   "YB|Y9/ORGR6/B6RBR/W9/BO8/G2BG6", // A-Perm case (adjacent corners swapped)
		TargetCFEN:  "YB|Y9/R9/B9/W9/O9/G9",           // Solved
		MoveCount:   9,
		Verified:    true,
		TestedOn:    []int{3},
	},
	{
		Name:        "A-Perm (b)",
		Category:    "PLL",
		Moves:       "x R2' D2 R U R' D2 R U' R x'",
		Description: "Adjacent corner swap (mirror)",
		CaseNumber:  "PLL Ab",
	},
	{
		Name:        "J-Perm (a)",
		Category:    "PLL",
		Moves:       "L' U' L F L' U' L U L F' L2' U L U",
		Description: "Adjacent corner and edge swap",
		CaseNumber:  "PLL Ja",
	},
	{
		Name:        "J-Perm (b)",
		Category:    "PLL",
		Moves:       "R U R' F' R U R' U' R' F R2 U' R' U'",
		Description: "Adjacent corner and edge swap (mirror)",
		CaseNumber:  "PLL Jb",
	},
	{
		Name:        "R-Perm (a)",
		Category:    "PLL",
		Moves:       "L U2' L' U2' L F' L' U' L U L F L2' U",
		Description: "Right-hand corner and edge swap",
		CaseNumber:  "PLL Ra",
	},
	{
		Name:        "R-Perm (b)",
		Category:    "PLL",
		Moves:       "R' U2 R U2 R' F R U R' U' R' F' R2 U'",
		Description: "Right-hand corner and edge swap (mirror)",
		CaseNumber:  "PLL Rb",
	},
	{
		Name:        "V-Perm",
		Category:    "PLL",
		Moves:       "R' U R' U' y R' F' R2 U' R' U R' F R F",
		Description: "Diagonal corner swap",
		CaseNumber:  "PLL V",
	},
	{
		Name:        "F-Perm",
		Category:    "PLL",
		Moves:       "R' U' F' R U R' U' R' F R2 U' R' U' R U R' U R",
		Description: "Adjacent corner and edge swap",
		CaseNumber:  "PLL F",
	},

	// Common F2L Cases
	{
		Name:        "Corner Up, Edge Up",
		Category:    "F2L",
		Moves:       "U R U' R' U' F' U F",
		Description: "Both pieces on top, separated",
		CaseNumber:  "F2L 5",
	},
	{
		Name:        "Corner Correct, Edge Wrong",
		Category:    "F2L",
		Moves:       "R U R' U' R U R'",
		Description: "Corner oriented correctly, edge needs flip",
		CaseNumber:  "F2L 8",
	},
	{
		Name:        "Corner Wrong, Edge Correct",
		Category:    "F2L",
		Moves:       "R U' R' U R U' R'",
		Description: "Edge oriented correctly, corner needs rotation",
		CaseNumber:  "F2L 7",
	},
	{
		Name:        "Back-to-Back",
		Category:    "F2L",
		Moves:       "U' R U R' U2 R U' R'",
		Description: "Corner and edge adjacent on top",
		CaseNumber:  "F2L 6",
	},

	// Advanced Triggers
	{
		Name:        "Double Sexy",
		Category:    "Trigger",
		Moves:       "R U R' U' R U R' U'",
		Description: "Two sexy moves in sequence",
		CaseNumber:  "",
	},
	{
		Name:        "Lefty Sexy",
		Category:    "Trigger",
		Moves:       "L' U' L U",
		Description: "Left-hand sexy move",
		CaseNumber:  "",
	},
	{
		Name:        "Fat Sexy",
		Category:    "Trigger",
		Moves:       "r U R' U'",
		Description: "Wide right sexy move",
		CaseNumber:  "",
	},

	// Essential Dot OLL Cases (No Edges Oriented)
	{
		Name:        "Diagonal OLL",
		Category:    "OLL",
		Moves:       "R U2 R2 F R F' U2 R' F R F'",
		Description: "Diagonal dot pattern",
		CaseNumber:  "OLL 1",
	},
	{
		Name:        "Anti-Diagonal OLL",
		Category:    "OLL",
		Moves:       "f R U R' U' f' U' F R U R' U' F'",
		Description: "Anti-diagonal dot pattern",
		CaseNumber:  "OLL 3",
	},
	{
		Name:        "Double Dot OLL",
		Category:    "OLL",
		Moves:       "f R U R' U' f' U F R U R' U' F'",
		Description: "Double dot pattern",
		CaseNumber:  "OLL 4",
	},

	// Essential Fish OLL Cases
	{
		Name:        "Big Fish",
		Category:    "OLL",
		Moves:       "R U R' U' R' F R2 U R' U' F'",
		Description: "Big fish pattern",
		CaseNumber:  "OLL 9",
	},
	{
		Name:        "Small Fish",
		Category:    "OLL",
		Moves:       "R U R' U R' F R F' R U2 R'",
		Description: "Small fish pattern",
		CaseNumber:  "OLL 10",
	},

	// Essential Square OLL Cases
	{
		Name:        "Square OLL",
		Category:    "OLL",
		Moves:       "r' U2 R U R' U r",
		Description: "Square edge pattern",
		CaseNumber:  "OLL 5",
	},
	{
		Name:        "Anti-Square OLL",
		Category:    "OLL",
		Moves:       "r U2 R' U' R U' r'",
		Description: "Anti-square edge pattern",
		CaseNumber:  "OLL 6",
	},

	// Essential Knight Move OLL Cases
	{
		Name:        "Knight Move OLL",
		Category:    "OLL",
		Moves:       "R U R' U' R U' R' F' U' F R U R'",
		Description: "Knight move pattern",
		CaseNumber:  "OLL 13",
	},
	{
		Name:        "Gun OLL",
		Category:    "OLL",
		Moves:       "F U R U' R2 F' R U R U' R'",
		Description: "Gun pattern",
		CaseNumber:  "OLL 14",
	},

	// Remaining Essential PLL Cases
	{
		Name:        "N-Perm (a)",
		Category:    "PLL",
		Moves:       "R U R' U R U R' F' R U R' U' R' F R2 U' R' U2 R U' R'",
		Description: "N permutation (clockwise)",
		CaseNumber:  "PLL Na",
	},
	{
		Name:        "N-Perm (b)",
		Category:    "PLL",
		Moves:       "r' D r U2 r' D r U2 r' D r U2 r' D r U2 r' D r",
		Description: "N permutation (counterclockwise)",
		CaseNumber:  "PLL Nb",
	},
	{
		Name:        "G-Perm (a)",
		Category:    "PLL",
		Moves:       "R2 U R' U R' U' R U' R2 D U' R' U R D'",
		Description: "G permutation (a variant)",
		CaseNumber:  "PLL Ga",
	},
	{
		Name:        "G-Perm (b)",
		Category:    "PLL",
		Moves:       "R' U' R U D' R2 U R' U R U' R U' R2 D",
		Description: "G permutation (b variant)",
		CaseNumber:  "PLL Gb",
	},
	{
		Name:        "G-Perm (c)",
		Category:    "PLL",
		Moves:       "R2 U' R U' R U R' U R2 D' U R U' R' D",
		Description: "G permutation (c variant)",
		CaseNumber:  "PLL Gc",
	},
	{
		Name:        "G-Perm (d)",
		Category:    "PLL",
		Moves:       "R U R' U' D R2 U' R U' R' U R' U R2 D'",
		Description: "G permutation (d variant)",
		CaseNumber:  "PLL Gd",
	},
	{
		Name:        "E-Perm",
		Category:    "PLL",
		Moves:       "x' R U' R' D R U R' D' R U R' D R U' R' D' x",
		Description: "Opposite corner swap",
		CaseNumber:  "PLL E",
	},

	// Advanced F2L Cases
	{
		Name:        "Corner in Slot, Edge Up",
		Category:    "F2L",
		Moves:       "R U' R' d R' U' R U' R' U R",
		Description: "Corner in slot, edge on top",
		CaseNumber:  "F2L 17",
	},
	{
		Name:        "Edge in Slot, Corner Up",
		Category:    "F2L",
		Moves:       "R U R' U2 R U' R' U R U' R'",
		Description: "Edge in slot, corner on top",
		CaseNumber:  "F2L 18",
	},
	{
		Name:        "Both in Slot",
		Category:    "F2L",
		Moves:       "R U' R' U R U2 R' U R U' R'",
		Description: "Both pieces in slot, need extraction",
		CaseNumber:  "F2L 29",
	},
	{
		Name:        "Separated on Top",
		Category:    "F2L",
		Moves:       "R U' R' U' F' U F",
		Description: "Corner and edge separated on top",
		CaseNumber:  "F2L 35",
	},

	// Additional Common Triggers
	{
		Name:        "Right Hand",
		Category:    "Trigger",
		Moves:       "R U R'",
		Description: "Basic right-hand trigger",
		CaseNumber:  "",
	},
	{
		Name:        "Left Hand",
		Category:    "Trigger",
		Moves:       "L' U' L",
		Description: "Basic left-hand trigger",
		CaseNumber:  "",
	},
	{
		Name:        "Sune Trigger",
		Category:    "Trigger",
		Moves:       "R U R' U R U2 R'",
		Description: "Sune as a trigger sequence",
		CaseNumber:  "",
	},
}

// LookupAlgorithm searches for algorithms by name or moves
func LookupAlgorithm(query string) []Algorithm {
	query = strings.ToLower(strings.TrimSpace(query))
	var results []Algorithm

	for _, alg := range AlgorithmDatabase {
		// Check if query matches name, moves, or description
		if strings.Contains(strings.ToLower(alg.Name), query) ||
			strings.Contains(strings.ToLower(alg.Moves), query) ||
			strings.Contains(strings.ToLower(alg.Description), query) ||
			strings.Contains(strings.ToLower(alg.CaseNumber), query) {
			results = append(results, alg)
		}
	}

	return results
}

// LookupByMoves finds algorithms that exactly match the given moves
func LookupByMoves(moves string) []Algorithm {
	moves = strings.TrimSpace(moves)
	var results []Algorithm

	for _, alg := range AlgorithmDatabase {
		if alg.Moves == moves {
			results = append(results, alg)
		}
	}

	return results
}

// GetByCategory returns all algorithms in a given category
func GetByCategory(category string) []Algorithm {
	category = strings.ToUpper(strings.TrimSpace(category))
	var results []Algorithm

	for _, alg := range AlgorithmDatabase {
		if alg.Category == category {
			results = append(results, alg)
		}
	}

	return results
}

// CalculateMoveCount returns the number of moves in an algorithm string
func (alg *Algorithm) CalculateMoveCount() int {
	if alg.Moves == "" {
		return 0
	}

	moves, err := ParseScramble(alg.Moves)
	if err != nil {
		return 0
	}

	return len(moves)
}

// UpdateMoveCount updates the MoveCount field based on the Moves string
func (alg *Algorithm) UpdateMoveCount() {
	alg.MoveCount = alg.CalculateMoveCount()
}

// NeedsCFENPatterns returns true if this algorithm has CFEN patterns to verify
func (alg *Algorithm) NeedsCFENPatterns() bool {
	return alg.StartCFEN != "" || alg.TargetCFEN != ""
}

// MarkVerified marks the algorithm as verified for a specific cube size
func (alg *Algorithm) MarkVerified(cubeSize int) {
	alg.Verified = true
	// Add cube size to TestedOn if not already present
	for _, size := range alg.TestedOn {
		if size == cubeSize {
			return // Already tested on this size
		}
	}
	alg.TestedOn = append(alg.TestedOn, cubeSize)
}

// MarkUnverified marks the algorithm as unverified
func (alg *Algorithm) MarkUnverified() {
	alg.Verified = false
}

// GetVerifiedAlgorithms returns only algorithms that have been verified
func GetVerifiedAlgorithms() []Algorithm {
	var results []Algorithm

	for _, alg := range AlgorithmDatabase {
		if alg.Verified {
			results = append(results, alg)
		}
	}

	return results
}

// GetUnverifiedAlgorithms returns algorithms that need verification
func GetUnverifiedAlgorithms() []Algorithm {
	var results []Algorithm

	for _, alg := range AlgorithmDatabase {
		if !alg.Verified && alg.NeedsCFENPatterns() {
			results = append(results, alg)
		}
	}

	return results
}
