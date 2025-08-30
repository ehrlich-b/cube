package cube

// piece_mapping.go - 3D cube piece position mappings
//
// This file contains the complex 3D geometry mappings needed to identify
// edges and corners that span multiple faces of the cube.

// EdgeMap defines which faces and positions make up each edge piece
type EdgeMap struct {
	Face1 Face
	Row1  int
	Col1  int
	Face2 Face  
	Row2  int
	Col2  int
}

// CornerMap defines which faces and positions make up each corner piece  
type CornerMap struct {
	Face1 Face
	Row1  int
	Col1  int
	Face2 Face
	Row2  int
	Col2  int
	Face3 Face
	Row3  int
	Col3  int
}

// Get3x3EdgeMappings returns all edge piece mappings for a 3x3 cube
func Get3x3EdgeMappings() []EdgeMap {
	return []EdgeMap{
		// Up face edges
		{Up, 0, 1, Back, 0, 1},   // UB - Up-Back edge
		{Up, 1, 0, Left, 0, 1},   // UL - Up-Left edge  
		{Up, 1, 2, Right, 0, 1},  // UR - Up-Right edge
		{Up, 2, 1, Front, 0, 1},  // UF - Up-Front edge
		
		// Middle layer edges (between side faces)
		{Front, 1, 0, Left, 1, 2},   // FL - Front-Left edge
		{Front, 1, 2, Right, 1, 0},  // FR - Front-Right edge
		{Back, 1, 0, Right, 1, 2},   // BR - Back-Right edge
		{Back, 1, 2, Left, 1, 0},    // BL - Back-Left edge
		
		// Down face edges  
		{Down, 0, 1, Front, 2, 1},  // DF - Down-Front edge
		{Down, 1, 0, Left, 2, 1},   // DL - Down-Left edge
		{Down, 1, 2, Right, 2, 1},  // DR - Down-Right edge
		{Down, 2, 1, Back, 2, 1},   // DB - Down-Back edge
	}
}

// Get3x3CornerMappings returns all corner piece mappings for a 3x3 cube
func Get3x3CornerMappings() []CornerMap {
	return []CornerMap{
		// Up face corners
		{Up, 0, 0, Back, 0, 2, Left, 0, 0},   // UBL - Up-Back-Left corner
		{Up, 0, 2, Back, 0, 0, Right, 0, 2},  // UBR - Up-Back-Right corner  
		{Up, 2, 0, Front, 0, 0, Left, 0, 2},  // UFL - Up-Front-Left corner
		{Up, 2, 2, Front, 0, 2, Right, 0, 0}, // UFR - Up-Front-Right corner
		
		// Down face corners
		{Down, 0, 0, Front, 2, 0, Left, 2, 2},  // DFL - Down-Front-Left corner
		{Down, 0, 2, Front, 2, 2, Right, 2, 0}, // DFR - Down-Front-Right corner
		{Down, 2, 0, Back, 2, 2, Left, 2, 0},   // DBL - Down-Back-Left corner  
		{Down, 2, 2, Back, 2, 0, Right, 2, 2},  // DBR - Down-Back-Right corner
	}
}

// getEdgeColors retrieves the actual colors of an edge piece using proper mappings
func (c *Cube) getEdgeColorsProper(edgeMap EdgeMap) []Color {
	if c.Size != 3 {
		return nil // Only supports 3x3 for now
	}
	
	color1 := c.Faces[edgeMap.Face1][edgeMap.Row1][edgeMap.Col1]
	color2 := c.Faces[edgeMap.Face2][edgeMap.Row2][edgeMap.Col2]
	
	return []Color{color1, color2}
}

// getCornerColors retrieves the actual colors of a corner piece using proper mappings
func (c *Cube) getCornerColorsProper(cornerMap CornerMap) []Color {
	if c.Size != 3 {
		return nil // Only supports 3x3 for now
	}
	
	color1 := c.Faces[cornerMap.Face1][cornerMap.Row1][cornerMap.Col1]
	color2 := c.Faces[cornerMap.Face2][cornerMap.Row2][cornerMap.Col2]
	color3 := c.Faces[cornerMap.Face3][cornerMap.Row3][cornerMap.Col3]
	
	return []Color{color1, color2, color3}
}

// GetAllEdges returns all edge pieces on the cube with their colors and positions
func (c *Cube) GetAllEdges() []*Piece {
	if c.Size != 3 {
		return nil // Only supports 3x3 for now
	}
	
	edgeMappings := Get3x3EdgeMappings()
	edges := make([]*Piece, len(edgeMappings))
	
	for i, mapping := range edgeMappings {
		colors := c.getEdgeColorsProper(mapping)
		piece := NewPiece(colors)
		piece.Position = Position{
			Face: mapping.Face1, // Primary face position
			Row:  mapping.Row1,
			Col:  mapping.Col1,
		}
		edges[i] = piece
	}
	
	return edges
}

// GetAllCorners returns all corner pieces on the cube with their colors and positions
func (c *Cube) GetAllCorners() []*Piece {
	if c.Size != 3 {
		return nil // Only supports 3x3 for now
	}
	
	cornerMappings := Get3x3CornerMappings()
	corners := make([]*Piece, len(cornerMappings))
	
	for i, mapping := range cornerMappings {
		colors := c.getCornerColorsProper(mapping)
		piece := NewPiece(colors)
		piece.Position = Position{
			Face: mapping.Face1, // Primary face position
			Row:  mapping.Row1,
			Col:  mapping.Col1,
		}
		corners[i] = piece
	}
	
	return corners
}

// GetAllCenters returns all center pieces on the cube
func (c *Cube) GetAllCenters() []*Piece {
	centers := make([]*Piece, 6) // 6 centers on any cube
	centerPos := c.Size / 2
	
	for face := 0; face < 6; face++ {
		color := c.Faces[face][centerPos][centerPos]
		piece := NewPiece([]Color{color})
		piece.Position = Position{
			Face: Face(face),
			Row:  centerPos,
			Col:  centerPos,
		}
		centers[face] = piece
	}
	
	return centers
}

// findEdgeMappingByPosition finds the edge mapping for a given primary position
func findEdgeMappingByPosition(face Face, row, col int) *EdgeMap {
	mappings := Get3x3EdgeMappings()
	for _, mapping := range mappings {
		if mapping.Face1 == face && mapping.Row1 == row && mapping.Col1 == col {
			return &mapping
		}
		// Also check secondary position
		if mapping.Face2 == face && mapping.Row2 == row && mapping.Col2 == col {
			// Return with swapped primary/secondary
			return &EdgeMap{
				Face1: mapping.Face2, Row1: mapping.Row2, Col1: mapping.Col2,
				Face2: mapping.Face1, Row2: mapping.Row1, Col2: mapping.Col1,
			}
		}
	}
	return nil
}

// findCornerMappingByPosition finds the corner mapping for a given primary position  
func findCornerMappingByPosition(face Face, row, col int) *CornerMap {
	mappings := Get3x3CornerMappings()
	for _, mapping := range mappings {
		if mapping.Face1 == face && mapping.Row1 == row && mapping.Col1 == col {
			return &mapping
		}
		// Also check secondary and tertiary positions
		if mapping.Face2 == face && mapping.Row2 == row && mapping.Col2 == col {
			// Return with rotated positions
			return &CornerMap{
				Face1: mapping.Face2, Row1: mapping.Row2, Col1: mapping.Col2,
				Face2: mapping.Face3, Row2: mapping.Row3, Col2: mapping.Col3,
				Face3: mapping.Face1, Row3: mapping.Row1, Col3: mapping.Col1,
			}
		}
		if mapping.Face3 == face && mapping.Row3 == row && mapping.Col3 == col {
			// Return with rotated positions
			return &CornerMap{
				Face1: mapping.Face3, Row1: mapping.Row3, Col1: mapping.Col3,
				Face2: mapping.Face1, Row2: mapping.Row1, Col2: mapping.Col1,
				Face3: mapping.Face2, Row3: mapping.Row2, Col3: mapping.Col2,
			}
		}
	}
	return nil
}