package cube

// patterns.go - Semantic pattern recognition for cube solving
//
// This file implements pattern recognition for common cube solving stages
// like white cross, first two layers, OLL, PLL, etc.

// Pattern represents a recognizable cube state or partial state
type Pattern interface {
	Name() string
	Matches(cube *Cube) bool
	Description() string
	CompletionPercent(cube *Cube) float64
}

// WhiteCrossPattern checks if the white cross is solved
type WhiteCrossPattern struct{}

func (p WhiteCrossPattern) Name() string {
	return "White Cross"
}

func (p WhiteCrossPattern) Description() string {
	return "Four white edge pieces correctly positioned on the bottom face"
}

func (p WhiteCrossPattern) Matches(cube *Cube) bool {
	if cube.Size != 3 {
		return false
	}
	
	// Check that all four white edges are in correct positions
	whiteEdges := [][]Color{
		{White, Blue},   // White-Blue edge (front)
		{White, Red},    // White-Red edge (right) 
		{White, Green},  // White-Green edge (back)
		{White, Orange}, // White-Orange edge (left)
	}
	
	for _, edgeColors := range whiteEdges {
		edge := cube.GetPieceByColors(edgeColors)
		if edge == nil || !cube.IsPieceInCorrectPosition(edgeColors) {
			return false
		}
	}
	
	return true
}

func (p WhiteCrossPattern) CompletionPercent(cube *Cube) float64 {
	if cube.Size != 3 {
		return 0.0
	}
	
	whiteEdges := [][]Color{
		{White, Blue}, {White, Red}, {White, Green}, {White, Orange},
	}
	
	correct := 0
	for _, edgeColors := range whiteEdges {
		if cube.IsPieceInCorrectPosition(edgeColors) {
			correct++
		}
	}
	
	return float64(correct) / 4.0 * 100.0
}

// WhiteLayerPattern checks if the entire white (bottom) layer is solved
type WhiteLayerPattern struct{}

func (p WhiteLayerPattern) Name() string {
	return "White Layer"
}

func (p WhiteLayerPattern) Description() string {
	return "Complete white layer with four corners correctly positioned"
}

func (p WhiteLayerPattern) Matches(cube *Cube) bool {
	// First check if white cross is solved
	crossPattern := WhiteCrossPattern{}
	if !crossPattern.Matches(cube) {
		return false
	}
	
	// Check that all four white corners are in correct positions
	whiteCorners := [][]Color{
		{White, Blue, Red},    // White-Blue-Red corner
		{White, Red, Green},   // White-Red-Green corner  
		{White, Green, Orange}, // White-Green-Orange corner
		{White, Orange, Blue}, // White-Orange-Blue corner
	}
	
	for _, cornerColors := range whiteCorners {
		if !cube.IsPieceInCorrectPosition(cornerColors) {
			return false
		}
	}
	
	return true
}

func (p WhiteLayerPattern) CompletionPercent(cube *Cube) float64 {
	if cube.Size != 3 {
		return 0.0
	}
	
	whiteCorners := [][]Color{
		{White, Blue, Red}, {White, Red, Green}, 
		{White, Green, Orange}, {White, Orange, Blue},
	}
	
	correct := 0
	for _, cornerColors := range whiteCorners {
		if cube.IsPieceInCorrectPosition(cornerColors) {
			correct++
		}
	}
	
	// Factor in cross completion too
	crossPattern := WhiteCrossPattern{}
	crossPercent := crossPattern.CompletionPercent(cube)
	
	cornerPercent := float64(correct) / 4.0 * 100.0
	
	// Weighted average: cross is 50%, corners are 50%
	return (crossPercent + cornerPercent) / 2.0
}

// F2LSlotPattern checks if a specific F2L slot is solved
type F2LSlotPattern struct {
	Slot int // 0-3 for the four F2L slots
}

func (p F2LSlotPattern) Name() string {
	slotNames := []string{"Front-Right F2L", "Back-Right F2L", "Back-Left F2L", "Front-Left F2L"}
	if p.Slot >= 0 && p.Slot < 4 {
		return slotNames[p.Slot]
	}
	return "F2L Slot"
}

func (p F2LSlotPattern) Description() string {
	return "Corner-edge pair correctly positioned in F2L slot"
}

func (p F2LSlotPattern) Matches(cube *Cube) bool {
	if cube.Size != 3 || p.Slot < 0 || p.Slot > 3 {
		return false
	}
	
	// Define F2L slot pieces
	slotCorners := [][]Color{
		{White, Blue, Red},    // Slot 0: Front-Right corner
		{White, Red, Green},   // Slot 1: Back-Right corner
		{White, Green, Orange}, // Slot 2: Back-Left corner
		{White, Orange, Blue}, // Slot 3: Front-Left corner
	}
	
	slotEdges := [][]Color{
		{Blue, Red},    // Slot 0: Front-Right edge
		{Red, Green},   // Slot 1: Back-Right edge  
		{Green, Orange}, // Slot 2: Back-Left edge
		{Orange, Blue}, // Slot 3: Front-Left edge
	}
	
	if p.Slot >= len(slotCorners) || p.Slot >= len(slotEdges) {
		return false
	}
	
	cornerColors := slotCorners[p.Slot]
	edgeColors := slotEdges[p.Slot]
	
	return cube.IsPieceInCorrectPosition(cornerColors) &&
		   cube.IsPieceInCorrectPosition(edgeColors)
}

func (p F2LSlotPattern) CompletionPercent(cube *Cube) float64 {
	if cube.Size != 3 {
		return 0.0
	}
	
	if p.Matches(cube) {
		return 100.0
	}
	
	// Could add partial completion logic here
	return 0.0
}

// OLLSolvedPattern checks if the last layer orientation is solved (all yellow on top)
type OLLSolvedPattern struct{}

func (p OLLSolvedPattern) Name() string {
	return "OLL (Last Layer Oriented)"
}

func (p OLLSolvedPattern) Description() string {
	return "All yellow stickers on the top face"
}

func (p OLLSolvedPattern) Matches(cube *Cube) bool {
	if cube.Size != 3 {
		return false
	}
	
	// Check that all stickers on the Up face are yellow
	for row := 0; row < cube.Size; row++ {
		for col := 0; col < cube.Size; col++ {
			if cube.Faces[Up][row][col] != Yellow {
				return false
			}
		}
	}
	
	return true
}

func (p OLLSolvedPattern) CompletionPercent(cube *Cube) float64 {
	if cube.Size != 3 {
		return 0.0
	}
	
	yellowCount := 0
	totalStickers := cube.Size * cube.Size
	
	for row := 0; row < cube.Size; row++ {
		for col := 0; col < cube.Size; col++ {
			if cube.Faces[Up][row][col] == Yellow {
				yellowCount++
			}
		}
	}
	
	return float64(yellowCount) / float64(totalStickers) * 100.0
}

// PLLSolvedPattern checks if the last layer is completely solved
type PLLSolvedPattern struct{}

func (p PLLSolvedPattern) Name() string {
	return "PLL (Cube Solved)"
}

func (p PLLSolvedPattern) Description() string {
	return "Last layer correctly permuted - cube completely solved"
}

func (p PLLSolvedPattern) Matches(cube *Cube) bool {
	return cube.IsSolved()
}

func (p PLLSolvedPattern) CompletionPercent(cube *Cube) float64 {
	if cube.IsSolved() {
		return 100.0
	}
	
	// Check OLL first
	ollPattern := OLLSolvedPattern{}
	if !ollPattern.Matches(cube) {
		return 0.0 // Can't have PLL without OLL
	}
	
	// Count correctly positioned last layer pieces
	// This is a simplified check - could be more sophisticated
	lastLayerPieces := [][]Color{
		{Yellow, Blue}, {Yellow, Red}, {Yellow, Green}, {Yellow, Orange}, // Edges
		{Yellow, Blue, Red}, {Yellow, Red, Green}, {Yellow, Green, Orange}, {Yellow, Orange, Blue}, // Corners
	}
	
	correct := 0
	for _, pieceColors := range lastLayerPieces {
		if cube.IsPieceInCorrectPosition(pieceColors) {
			correct++
		}
	}
	
	return float64(correct) / float64(len(lastLayerPieces)) * 100.0
}

// GetAllPatterns returns all available patterns for recognition
func GetAllPatterns() []Pattern {
	return []Pattern{
		WhiteCrossPattern{},
		WhiteLayerPattern{},
		F2LSlotPattern{Slot: 0}, F2LSlotPattern{Slot: 1},
		F2LSlotPattern{Slot: 2}, F2LSlotPattern{Slot: 3},
		OLLSolvedPattern{},
		PLLSolvedPattern{},
	}
}

// AnalyzeCubeState returns which patterns match the current cube state
func AnalyzeCubeState(cube *Cube) map[string]float64 {
	patterns := GetAllPatterns()
	results := make(map[string]float64)
	
	for _, pattern := range patterns {
		completion := pattern.CompletionPercent(cube)
		if completion > 0 {
			results[pattern.Name()] = completion
		}
	}
	
	return results
}

// GetNextStep suggests the next logical solving step based on current state
func GetNextStep(cube *Cube) string {
	if cube.IsSolved() {
		return "Cube is already solved! 🎉"
	}
	
	// Check patterns in solving order
	crossPattern := WhiteCrossPattern{}
	if !crossPattern.Matches(cube) {
		return "Solve the white cross on the bottom"
	}
	
	layerPattern := WhiteLayerPattern{}
	if !layerPattern.Matches(cube) {
		return "Complete the white layer (insert white corners)"
	}
	
	// Check F2L slots
	allF2LSolved := true
	for slot := 0; slot < 4; slot++ {
		f2lPattern := F2LSlotPattern{Slot: slot}
		if !f2lPattern.Matches(cube) {
			allF2LSolved = false
			break
		}
	}
	
	if !allF2LSolved {
		return "Complete F2L (First Two Layers)"
	}
	
	ollPattern := OLLSolvedPattern{}
	if !ollPattern.Matches(cube) {
		return "Orient last layer (OLL - make top face all yellow)"
	}
	
	return "Permute last layer (PLL - solve remaining pieces)"
}