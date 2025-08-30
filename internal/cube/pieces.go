package cube

import (
	"sort"
)

// PieceType represents the type of a cube piece
type PieceType int

const (
	Corner PieceType = iota
	Edge
	Center
)

func (pt PieceType) String() string {
	return []string{"Corner", "Edge", "Center"}[pt]
}

// Position represents a piece's location on the cube
type Position struct {
	Face Face
	Row  int
	Col  int
}

// Piece represents a physical piece on the cube with its colors and type
type Piece struct {
	Type     PieceType
	Colors   []Color   // Colors in canonical order
	Position Position  // Current position on cube
	Orientation int     // Rotation state (0 = correct, 1/2 = rotated)
}

// NewPiece creates a piece from colors (automatically determines type)
func NewPiece(colors []Color) *Piece {
	piece := &Piece{
		Colors: make([]Color, len(colors)),
	}
	copy(piece.Colors, colors)
	
	// Sort colors to create canonical representation for identification
	sortedColors := make([]Color, len(colors))
	copy(sortedColors, colors)
	sort.Slice(sortedColors, func(i, j int) bool {
		return int(sortedColors[i]) < int(sortedColors[j])
	})
	piece.Colors = sortedColors
	
	// Determine piece type based on number of colors
	switch len(colors) {
	case 1:
		piece.Type = Center
	case 2:
		piece.Type = Edge
	case 3:
		piece.Type = Corner
	default:
		// Invalid piece
		piece.Type = Center
	}
	
	return piece
}

// GetPieceByColors finds a piece in the cube by its color combination
func (c *Cube) GetPieceByColors(colors []Color) *Piece {
	// Sort colors for canonical matching
	sortedColors := make([]Color, len(colors))
	copy(sortedColors, colors)
	sort.Slice(sortedColors, func(i, j int) bool {
		return int(sortedColors[i]) < int(sortedColors[j])
	})
	
	switch len(colors) {
	case 2: // Edge piece
		return c.findEdgeByColors(sortedColors)
	case 3: // Corner piece
		return c.findCornerByColors(sortedColors)
	case 1: // Center piece
		return c.findCenterByColor(colors[0])
	}
	
	return nil
}

// GetPieceLocation returns the position of a piece with given colors
func (c *Cube) GetPieceLocation(colors []Color) Position {
	piece := c.GetPieceByColors(colors)
	if piece != nil {
		return piece.Position
	}
	return Position{} // Invalid position
}

// IsPieceInCorrectPosition checks if a piece is in its solved position
func (c *Cube) IsPieceInCorrectPosition(colors []Color) bool {
	piece := c.GetPieceByColors(colors)
	if piece == nil {
		return false
	}
	
	// Get the solved position for this piece
	solvedPosition := c.getSolvedPosition(colors)
	
	return piece.Position.Face == solvedPosition.Face &&
		   piece.Position.Row == solvedPosition.Row &&
		   piece.Position.Col == solvedPosition.Col
}

// IsPieceCorrectlyOriented checks if a piece is oriented correctly
func (c *Cube) IsPieceCorrectlyOriented(colors []Color) bool {
	piece := c.GetPieceByColors(colors)
	if piece == nil {
		return false
	}
	
	// For now, consider orientation correct if the piece is in correct position
	// and the primary color is on the correct face
	if !c.IsPieceInCorrectPosition(colors) {
		return false
	}
	
	// Check orientation based on piece type
	switch piece.Type {
	case Corner:
		return c.isCornerCorrectlyOriented(piece)
	case Edge:
		return c.isEdgeCorrectlyOriented(piece)
	case Center:
		return true // Centers can't be misoriented on 3x3
	}
	
	return false
}

// findEdgeByColors locates an edge piece by its two colors
func (c *Cube) findEdgeByColors(sortedColors []Color) *Piece {
	if c.Size != 3 || len(sortedColors) != 2 {
		return nil
	}
	
	// Use proper edge mappings
	edges := c.GetAllEdges()
	for _, edge := range edges {
		// Sort edge colors for comparison
		edgeColors := make([]Color, len(edge.Colors))
		copy(edgeColors, edge.Colors)
		sort.Slice(edgeColors, func(i, j int) bool {
			return int(edgeColors[i]) < int(edgeColors[j])
		})
		
		// Check if colors match
		if len(edgeColors) == len(sortedColors) {
			match := true
			for i, color := range sortedColors {
				if edgeColors[i] != color {
					match = false
					break
				}
			}
			if match {
				return edge
			}
		}
	}
	
	return nil
}

// findCornerByColors locates a corner piece by its three colors  
func (c *Cube) findCornerByColors(sortedColors []Color) *Piece {
	if c.Size != 3 || len(sortedColors) != 3 {
		return nil
	}
	
	// Use proper corner mappings
	corners := c.GetAllCorners()
	for _, corner := range corners {
		// Sort corner colors for comparison
		cornerColors := make([]Color, len(corner.Colors))
		copy(cornerColors, corner.Colors)
		sort.Slice(cornerColors, func(i, j int) bool {
			return int(cornerColors[i]) < int(cornerColors[j])
		})
		
		// Check if colors match
		if len(cornerColors) == len(sortedColors) {
			match := true
			for i, color := range sortedColors {
				if cornerColors[i] != color {
					match = false
					break
				}
			}
			if match {
				return corner
			}
		}
	}
	
	return nil
}

// findCenterByColor locates a center piece by its color
func (c *Cube) findCenterByColor(color Color) *Piece {
	centerPos := c.Size / 2 // Center position for any face
	
	for face := 0; face < 6; face++ {
		if c.Faces[face][centerPos][centerPos] == color {
			piece := NewPiece([]Color{color})
			piece.Position = Position{
				Face: Face(face),
				Row:  centerPos,
				Col:  centerPos,
			}
			return piece
		}
	}
	
	return nil
}

// getEdgePositions returns all edge positions on the cube
func (c *Cube) getEdgePositions() []Position {
	positions := make([]Position, 0, 12) // 12 edges on a cube
	center := c.Size / 2
	
	// For 3x3, edges are at positions (0,1), (1,0), (1,2), (2,1) on each face
	// We'll focus on 3x3 for now
	if c.Size == 3 {
		for face := 0; face < 6; face++ {
			// Top edge
			positions = append(positions, Position{Face: Face(face), Row: 0, Col: center})
			// Left edge  
			positions = append(positions, Position{Face: Face(face), Row: center, Col: 0})
			// Right edge
			positions = append(positions, Position{Face: Face(face), Row: center, Col: c.Size-1})
			// Bottom edge
			positions = append(positions, Position{Face: Face(face), Row: c.Size-1, Col: center})
		}
	}
	
	return positions
}

// getCornerPositions returns all corner positions on the cube
func (c *Cube) getCornerPositions() []Position {
	positions := make([]Position, 0, 8) // 8 corners on a cube
	
	// For any cube size, corners are at the four corners of each face
	if c.Size >= 2 {
		for face := 0; face < 6; face++ {
			// Top-left corner
			positions = append(positions, Position{Face: Face(face), Row: 0, Col: 0})
			// Top-right corner
			positions = append(positions, Position{Face: Face(face), Row: 0, Col: c.Size-1})
			// Bottom-left corner
			positions = append(positions, Position{Face: Face(face), Row: c.Size-1, Col: 0})
			// Bottom-right corner
			positions = append(positions, Position{Face: Face(face), Row: c.Size-1, Col: c.Size-1})
		}
	}
	
	return positions
}

// getEdgeColors retrieves the colors of an edge piece at a given position
func (c *Cube) getEdgeColors(pos Position) []Color {
	if c.Size != 3 {
		return nil // Only supports 3x3 for now
	}
	
	// Find the edge mapping for this position
	edgeMapping := findEdgeMappingByPosition(pos.Face, pos.Row, pos.Col)
	if edgeMapping != nil {
		return c.getEdgeColorsProper(*edgeMapping)
	}
	
	return nil
}

// getCornerColors retrieves the colors of a corner piece at a given position
func (c *Cube) getCornerColors(pos Position) []Color {
	if c.Size != 3 {
		return nil // Only supports 3x3 for now
	}
	
	// Find the corner mapping for this position  
	cornerMapping := findCornerMappingByPosition(pos.Face, pos.Row, pos.Col)
	if cornerMapping != nil {
		return c.getCornerColorsProper(*cornerMapping)
	}
	
	return nil
}

// getSolvedPosition returns where a piece should be in the solved state
func (c *Cube) getSolvedPosition(colors []Color) Position {
	// Create a solved cube and find where this piece would be
	solvedCube := NewCube(c.Size)
	piece := solvedCube.GetPieceByColors(colors)
	if piece != nil {
		return piece.Position
	}
	return Position{}
}

// isCornerCorrectlyOriented checks corner orientation
func (c *Cube) isCornerCorrectlyOriented(piece *Piece) bool {
	// Simplified: check if the piece is in the right position
	// Full implementation would check 3D orientation of corner stickers
	return c.IsPieceInCorrectPosition(piece.Colors)
}

// isEdgeCorrectlyOriented checks edge orientation  
func (c *Cube) isEdgeCorrectlyOriented(piece *Piece) bool {
	// Simplified: check if the piece is in the right position
	// Full implementation would check 2D flip state of edge stickers
	return c.IsPieceInCorrectPosition(piece.Colors)
}