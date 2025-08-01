package cube

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
)

// Feature flag for legacy move system
var useLegacyMoves = os.Getenv("LEGACY_TWISTER") == "true"

// SliceType represents middle slice moves
type SliceType int

const (
	NoSlice SliceType = iota
	M_Slice           // Middle slice between L/R
	E_Slice           // Equatorial slice between U/D
	S_Slice           // Standing slice between F/B
)

// String returns the notation for slice moves
func (s SliceType) String() string {
	switch s {
	case M_Slice:
		return "M"
	case E_Slice:
		return "E"
	case S_Slice:
		return "S"
	default:
		return ""
	}
}

// RotationType represents whole cube rotations
type RotationType int

const (
	NoRotation RotationType = iota
	X_Rotation              // Around R axis
	Y_Rotation              // Around U axis
	Z_Rotation              // Around F axis
)

// String returns the notation for cube rotations
func (r RotationType) String() string {
	switch r {
	case X_Rotation:
		return "x"
	case Y_Rotation:
		return "y"
	case Z_Rotation:
		return "z"
	default:
		return ""
	}
}

// Move represents a cube move with advanced notation support
type Move struct {
	Face      Face // F, B, L, R, U, D (or invalid for rotations/slices)
	Clockwise bool
	Double    bool
	Wide      bool         // For wide moves (Rw, Fw, etc.)
	WideDepth int          // How many layers for wide moves (default 2 for "Rw")
	Layer     int          // For layer-specific moves (2R, 3L, etc.) - 0 means outermost
	Slice     SliceType    // For middle slice moves (M, E, S)
	Rotation  RotationType // For cube rotations (x, y, z)
}

// String returns the standard notation for the move
func (m Move) String() string {
	var notation string

	// Handle slice moves
	if m.Slice != NoSlice {
		notation = m.Slice.String()
	} else if m.Rotation != NoRotation {
		// Handle cube rotations
		notation = m.Rotation.String()
	} else {
		// Handle face moves with layer and wide notation
		if m.Layer > 0 {
			notation = strconv.Itoa(m.Layer + 1) // Convert 0-based to 1-based
		}
		notation += m.Face.String()
		if m.Wide {
			notation += "w"
		}
	}

	// Add modifiers
	if m.Double {
		notation += "2"
	} else if !m.Clockwise {
		notation += "'"
	}

	return notation
}

// Coord represents a sticker coordinate (face, row, col)
type Coord struct {
	Face Face
	Row  int
	Col  int
}

// stickerIndex converts (face, row, col) to flat index
func stickerIndex(face Face, row, col, N int) int {
	return int(face)*N*N + row*N + col
}

// indexToCoord converts flat index back to (face, row, col)
func indexToCoord(idx, N int) (Face, int, int) {
	face := Face(idx / (N * N))
	remainder := idx % (N * N)
	row := remainder / N
	col := remainder % N
	return face, row, col
}

// Permutation represents a mapping of sticker indices
type Permutation []int

// MoveType represents the base move type
type MoveType int

const (
	MoveR MoveType = iota
	MoveL
	MoveU
	MoveD
	MoveF
	MoveB
	MoveM
	MoveE
	MoveS
	MoveX
	MoveY
	MoveZ
)

// PermKey represents a cache key for permutations
type PermKey struct {
	N            int
	MoveType     MoveType
	Layer        int
	QuarterTurns int
}

// Permutation cache with thread-safe access
var permCache = make(map[PermKey]Permutation)
var permCacheMu sync.RWMutex

// getPermutation retrieves or generates a permutation from cache
func getPermutation(N int, moveType MoveType, layer int, quarterTurns int) Permutation {
	key := PermKey{N, moveType, layer, quarterTurns}

	permCacheMu.RLock()
	if perm, ok := permCache[key]; ok {
		permCacheMu.RUnlock()
		return perm
	}
	permCacheMu.RUnlock()

	// Generate and cache
	perm := generatePermutation(N, moveType, layer, quarterTurns)

	permCacheMu.Lock()
	permCache[key] = perm
	permCacheMu.Unlock()

	return perm
}

// Ring generators define which stickers move for each type of move

// ringR generates ring coordinates for R move at layer k
func ringR(N, k int) []Coord {
	var ring []Coord
	// Up face: column N-1-k, rows 0 to N-1
	for r := 0; r < N; r++ {
		ring = append(ring, Coord{Up, r, N - 1 - k})
	}
	// Back face: column k, rows N-1 to 0 (opposite direction due to 3D orientation)
	for r := N - 1; r >= 0; r-- {
		ring = append(ring, Coord{Back, r, k})
	}
	// Down face: column N-1-k, rows 0 to N-1 (same as Up, they're parallel)
	for r := 0; r < N; r++ {
		ring = append(ring, Coord{Down, r, N - 1 - k})
	}
	// Front face: column N-1-k, rows 0 to N-1
	for r := 0; r < N; r++ {
		ring = append(ring, Coord{Front, r, N - 1 - k})
	}
	return ring
}

// ringL generates ring coordinates for L move at layer k
func ringL(N, k int) []Coord {
	var ring []Coord
	// Up face: column k, rows 0 to N-1
	for r := 0; r < N; r++ {
		ring = append(ring, Coord{Up, r, k})
	}
	// Front face: column k, rows 0 to N-1
	for r := 0; r < N; r++ {
		ring = append(ring, Coord{Front, r, k})
	}
	// Down face: column k, rows N-1 to 0 (reversed)
	for r := N - 1; r >= 0; r-- {
		ring = append(ring, Coord{Down, r, k})
	}
	// Back face: column N-1-k, rows N-1 to 0 (reversed)
	for r := N - 1; r >= 0; r-- {
		ring = append(ring, Coord{Back, r, N - 1 - k})
	}
	return ring
}

// ringU generates ring coordinates for U move at layer k
func ringU(N, k int) []Coord {
	var ring []Coord
	// Back face: row k, columns 0 to N-1
	for c := 0; c < N; c++ {
		ring = append(ring, Coord{Back, k, c})
	}
	// Right face: row k, columns 0 to N-1
	for c := 0; c < N; c++ {
		ring = append(ring, Coord{Right, k, c})
	}
	// Front face: row k, columns 0 to N-1
	for c := 0; c < N; c++ {
		ring = append(ring, Coord{Front, k, c})
	}
	// Left face: row k, columns 0 to N-1
	for c := 0; c < N; c++ {
		ring = append(ring, Coord{Left, k, c})
	}
	return ring
}

// ringD generates ring coordinates for D move at layer k
func ringD(N, k int) []Coord {
	var ring []Coord
	// Front face: row N-1-k, columns 0 to N-1
	for c := 0; c < N; c++ {
		ring = append(ring, Coord{Front, N - 1 - k, c})
	}
	// Right face: row N-1-k, columns 0 to N-1
	for c := 0; c < N; c++ {
		ring = append(ring, Coord{Right, N - 1 - k, c})
	}
	// Back face: row N-1-k, columns 0 to N-1
	for c := 0; c < N; c++ {
		ring = append(ring, Coord{Back, N - 1 - k, c})
	}
	// Left face: row N-1-k, columns 0 to N-1
	for c := 0; c < N; c++ {
		ring = append(ring, Coord{Left, N - 1 - k, c})
	}
	return ring
}

// ringF generates ring coordinates for F move at layer k
func ringF(N, k int) []Coord {
	var ring []Coord
	// Up face: row N-1-k, columns 0 to N-1
	for c := 0; c < N; c++ {
		ring = append(ring, Coord{Up, N - 1 - k, c})
	}
	// Right face: column k, rows 0 to N-1
	for r := 0; r < N; r++ {
		ring = append(ring, Coord{Right, r, k})
	}
	// Down face: row k, columns N-1 to 0 (reversed)
	for c := N - 1; c >= 0; c-- {
		ring = append(ring, Coord{Down, k, c})
	}
	// Left face: column N-1-k, rows N-1 to 0 (reversed)
	for r := N - 1; r >= 0; r-- {
		ring = append(ring, Coord{Left, r, N - 1 - k})
	}
	return ring
}

// ringB generates ring coordinates for B move at layer k
func ringB(N, k int) []Coord {
	var ring []Coord
	// Up face: row k, columns N-1 to 0 (reversed)
	for c := N - 1; c >= 0; c-- {
		ring = append(ring, Coord{Up, k, c})
	}
	// Left face: column k, rows N-1 to 0 (reversed)
	for r := N - 1; r >= 0; r-- {
		ring = append(ring, Coord{Left, r, k})
	}
	// Down face: row N-1-k, columns 0 to N-1
	for c := 0; c < N; c++ {
		ring = append(ring, Coord{Down, N - 1 - k, c})
	}
	// Right face: column N-1-k, rows 0 to N-1
	for r := 0; r < N; r++ {
		ring = append(ring, Coord{Right, r, N - 1 - k})
	}
	return ring
}

// Slice move ring generators

// ringM generates ring coordinates for M slice move (between L and R)
func ringM(N, k int) []Coord {
	if N%2 == 0 {
		return nil // M slice undefined for even cubes for now
	}
	centerCol := N / 2
	var ring []Coord
	// Up face: column centerCol, rows 0 to N-1
	for r := 0; r < N; r++ {
		ring = append(ring, Coord{Up, r, centerCol})
	}
	// Front face: column centerCol, rows 0 to N-1
	for r := 0; r < N; r++ {
		ring = append(ring, Coord{Front, r, centerCol})
	}
	// Down face: column centerCol, rows 0 to N-1
	for r := 0; r < N; r++ {
		ring = append(ring, Coord{Down, r, centerCol})
	}
	// Back face: column centerCol, rows N-1 to 0 (reversed)
	for r := N - 1; r >= 0; r-- {
		ring = append(ring, Coord{Back, r, centerCol})
	}
	return ring
}

// ringE generates ring coordinates for E slice move (between U and D)
func ringE(N, k int) []Coord {
	if N%2 == 0 {
		return nil // E slice undefined for even cubes for now
	}
	centerRow := N / 2
	var ring []Coord
	// Front face: row centerRow, columns 0 to N-1
	for c := 0; c < N; c++ {
		ring = append(ring, Coord{Front, centerRow, c})
	}
	// Left face: row centerRow, columns 0 to N-1
	for c := 0; c < N; c++ {
		ring = append(ring, Coord{Left, centerRow, c})
	}
	// Back face: row centerRow, columns 0 to N-1
	for c := 0; c < N; c++ {
		ring = append(ring, Coord{Back, centerRow, c})
	}
	// Right face: row centerRow, columns 0 to N-1
	for c := 0; c < N; c++ {
		ring = append(ring, Coord{Right, centerRow, c})
	}
	return ring
}

// ringS generates ring coordinates for S slice move (between F and B)
func ringS(N, k int) []Coord {
	if N%2 == 0 {
		return nil // S slice undefined for even cubes for now
	}
	centerLayer := N / 2
	var ring []Coord
	// Up face: row N-1-centerLayer, columns 0 to N-1
	for c := 0; c < N; c++ {
		ring = append(ring, Coord{Up, N - 1 - centerLayer, c})
	}
	// Left face: column centerLayer, rows N-1 to 0 (reversed)
	for r := N - 1; r >= 0; r-- {
		ring = append(ring, Coord{Left, r, centerLayer})
	}
	// Down face: row centerLayer, columns N-1 to 0 (reversed)
	for c := N - 1; c >= 0; c-- {
		ring = append(ring, Coord{Down, centerLayer, c})
	}
	// Right face: column N-1-centerLayer, rows 0 to N-1
	for r := 0; r < N; r++ {
		ring = append(ring, Coord{Right, r, N - 1 - centerLayer})
	}
	return ring
}

// generateCubeRotationPermutation creates permutation for cube rotations (x, y, z)
func generateCubeRotationPermutation(N int, rotationType MoveType, quarterTurns int) Permutation {
	perm := make(Permutation, 6*N*N)
	// Initialize identity permutation
	for i := range perm {
		perm[i] = i
	}

	// Define face mappings for each rotation type
	var faceMappings [][]Face

	switch rotationType {
	case MoveX:
		// X rotation: around R face axis
		// Clockwise: F→D, D→B, B→U, U→F, L rotates CCW, R rotates CW
		if quarterTurns == 1 {
			faceMappings = [][]Face{
				{Front, Down},
				{Down, Back},
				{Back, Up},
				{Up, Front},
			}
		} else if quarterTurns == 2 {
			faceMappings = [][]Face{
				{Front, Back},
				{Back, Front},
				{Up, Down},
				{Down, Up},
			}
		} else { // quarterTurns == 3 (CCW)
			faceMappings = [][]Face{
				{Front, Up},
				{Up, Back},
				{Back, Down},
				{Down, Front},
			}
		}

	case MoveY:
		// Y rotation: around U face axis
		// Clockwise: F→L, L→B, B→R, R→F, U rotates CW, D rotates CCW
		if quarterTurns == 1 {
			faceMappings = [][]Face{
				{Front, Left},
				{Left, Back},
				{Back, Right},
				{Right, Front},
			}
		} else if quarterTurns == 2 {
			faceMappings = [][]Face{
				{Front, Back},
				{Back, Front},
				{Left, Right},
				{Right, Left},
			}
		} else { // quarterTurns == 3 (CCW)
			faceMappings = [][]Face{
				{Front, Right},
				{Right, Back},
				{Back, Left},
				{Left, Front},
			}
		}

	case MoveZ:
		// Z rotation: around F face axis
		// Clockwise: U→L, L→D, D→R, R→U, F rotates CW, B rotates CCW
		if quarterTurns == 1 {
			faceMappings = [][]Face{
				{Up, Left},
				{Left, Down},
				{Down, Right},
				{Right, Up},
			}
		} else if quarterTurns == 2 {
			faceMappings = [][]Face{
				{Up, Down},
				{Down, Up},
				{Left, Right},
				{Right, Left},
			}
		} else { // quarterTurns == 3 (CCW)
			faceMappings = [][]Face{
				{Up, Right},
				{Right, Down},
				{Down, Left},
				{Left, Up},
			}
		}

	default:
		return perm // Return identity for unknown rotations
	}

	// Apply face swaps
	for _, mapping := range faceMappings {
		srcFace := mapping[0]
		dstFace := mapping[1]

		// Copy entire face
		for row := 0; row < N; row++ {
			for col := 0; col < N; col++ {
				srcIdx := stickerIndex(srcFace, row, col, N)
				dstIdx := stickerIndex(dstFace, row, col, N)
				perm[srcIdx] = dstIdx
			}
		}
	}

	// Handle face rotations for the axis faces
	switch rotationType {
	case MoveX:
		// Left face rotates CCW, Right face rotates CW
		leftRotPerm := generateFaceRotationPermutation(N, MoveL, 4-quarterTurns) // CCW
		rightRotPerm := generateFaceRotationPermutation(N, MoveR, quarterTurns)  // CW

		// Compose permutations
		for i, dst := range leftRotPerm {
			if dst != i {
				perm[i] = dst
			}
		}
		for i, dst := range rightRotPerm {
			if dst != i {
				perm[i] = dst
			}
		}

	case MoveY:
		// Up face rotates CW, Down face rotates CCW
		upRotPerm := generateFaceRotationPermutation(N, MoveU, quarterTurns)     // CW
		downRotPerm := generateFaceRotationPermutation(N, MoveD, 4-quarterTurns) // CCW

		// Compose permutations
		for i, dst := range upRotPerm {
			if dst != i {
				perm[i] = dst
			}
		}
		for i, dst := range downRotPerm {
			if dst != i {
				perm[i] = dst
			}
		}

	case MoveZ:
		// Front face rotates CW, Back face rotates CCW
		frontRotPerm := generateFaceRotationPermutation(N, MoveF, quarterTurns)  // CW
		backRotPerm := generateFaceRotationPermutation(N, MoveB, 4-quarterTurns) // CCW

		// Compose permutations
		for i, dst := range frontRotPerm {
			if dst != i {
				perm[i] = dst
			}
		}
		for i, dst := range backRotPerm {
			if dst != i {
				perm[i] = dst
			}
		}
	}

	return perm
}

// rotateSlice rotates a slice of indices by quarterTurns
func rotateSlice(slice []int, quarterTurns int) []int {
	n := len(slice)
	if n == 0 {
		return slice
	}
	shift := (quarterTurns * n / 4) % n
	result := make([]int, n)
	for i := range slice {
		result[i] = slice[(i+shift)%n]
	}
	return result
}

// generatePermutation creates a permutation for a given move
func generatePermutation(N int, moveType MoveType, layer int, quarterTurns int) Permutation {
	perm := make(Permutation, 6*N*N)
	// Initialize identity permutation
	for i := range perm {
		perm[i] = i
	}

	// Get ring coordinates based on move type
	var ring []Coord
	switch moveType {
	case MoveR:
		ring = ringR(N, layer)
	case MoveL:
		ring = ringL(N, layer)
	case MoveU:
		ring = ringU(N, layer)
	case MoveD:
		ring = ringD(N, layer)
	case MoveF:
		ring = ringF(N, layer)
	case MoveB:
		ring = ringB(N, layer)
	case MoveM:
		ring = ringM(N, layer)
	case MoveE:
		ring = ringE(N, layer)
	case MoveS:
		ring = ringS(N, layer)
	case MoveX:
		// For cube rotations, invert the direction to match old system
		return generateCubeRotationPermutation(N, MoveX, (4-quarterTurns)%4)
	case MoveY:
		return generateCubeRotationPermutation(N, MoveY, (4-quarterTurns)%4)
	case MoveZ:
		return generateCubeRotationPermutation(N, MoveZ, (4-quarterTurns)%4)
	default:
		return perm // Return identity for unsupported moves for now
	}

	if ring == nil {
		return perm // Return identity if ring generation failed
	}

	// Convert to indices
	indices := make([]int, len(ring))
	for i, coord := range ring {
		indices[i] = stickerIndex(coord.Face, coord.Row, coord.Col, N)
	}

	// Apply rotation
	rotated := rotateSlice(indices, quarterTurns)
	for i, srcIdx := range indices {
		perm[srcIdx] = rotated[i]
	}

	// Handle face rotation if outer layer
	if layer == 0 {
		faceRotationPerm := generateFaceRotationPermutation(N, moveType, quarterTurns)
		// Compose with edge permutation
		for i, dst := range faceRotationPerm {
			if dst != i {
				perm[i] = dst
			}
		}
	}

	return perm
}

// generateFaceRotationPermutation creates permutation for rotating face stickers
func generateFaceRotationPermutation(N int, moveType MoveType, quarterTurns int) Permutation {
	perm := make(Permutation, 6*N*N)
	// Initialize identity permutation
	for i := range perm {
		perm[i] = i
	}

	var face Face
	switch moveType {
	case MoveR:
		face = Right
	case MoveL:
		face = Left
	case MoveU:
		face = Up
	case MoveD:
		face = Down
	case MoveF:
		face = Front
	case MoveB:
		face = Back
	default:
		return perm // No face rotation for slice moves
	}

	// Generate face rotation rings (concentric squares)
	for layer := 0; layer < N/2; layer++ {
		ring := generateFaceRing(face, N, layer)

		// Convert to indices
		indices := make([]int, len(ring))
		for i, coord := range ring {
			indices[i] = stickerIndex(coord.Face, coord.Row, coord.Col, N)
		}

		// Apply rotation
		rotated := rotateSlice(indices, quarterTurns)
		for i, srcIdx := range indices {
			perm[srcIdx] = rotated[i]
		}
	}

	return perm
}

// generateFaceRing generates coordinates for a ring on a face
func generateFaceRing(face Face, N, layer int) []Coord {
	var ring []Coord

	// Top edge (left to right)
	for c := layer; c < N-layer; c++ {
		ring = append(ring, Coord{face, layer, c})
	}

	// Right edge (top to bottom, excluding corner)
	for r := layer + 1; r < N-layer; r++ {
		ring = append(ring, Coord{face, r, N - 1 - layer})
	}

	// Bottom edge (right to left, excluding corner)
	if N-1-layer > layer {
		for c := N - 2 - layer; c >= layer; c-- {
			ring = append(ring, Coord{face, N - 1 - layer, c})
		}
	}

	// Left edge (bottom to top, excluding corners)
	if N-1-layer > layer {
		for r := N - 2 - layer; r > layer; r-- {
			ring = append(ring, Coord{face, r, layer})
		}
	}

	return ring
}

// applyPermutation applies a permutation to the cube
func applyPermutation(cube *Cube, perm Permutation) {
	N := cube.Size
	colors := make([]Color, 6*N*N)

	// Flatten cube to linear array
	idx := 0
	for face := 0; face < 6; face++ {
		for row := 0; row < N; row++ {
			for col := 0; col < N; col++ {
				colors[idx] = cube.Faces[face][row][col]
				idx++
			}
		}
	}

	// Apply permutation
	newColors := make([]Color, 6*N*N)
	for src, dst := range perm {
		newColors[dst] = colors[src]
	}

	// Unflatten back to cube
	idx = 0
	for face := 0; face < 6; face++ {
		for row := 0; row < N; row++ {
			for col := 0; col < N; col++ {
				cube.Faces[face][row][col] = newColors[idx]
				idx++
			}
		}
	}
}

// moveToMoveType converts a Move struct to MoveType and determines quarter turns
func moveToMoveType(move Move) (MoveType, int) {
	var moveType MoveType
	var quarterTurns int

	// Handle slice moves
	if move.Slice != NoSlice {
		switch move.Slice {
		case M_Slice:
			moveType = MoveM
		case E_Slice:
			moveType = MoveE
		case S_Slice:
			moveType = MoveS
		default:
			return MoveR, 0 // Default fallback
		}
	} else if move.Rotation != NoRotation {
		// Handle cube rotations
		switch move.Rotation {
		case X_Rotation:
			moveType = MoveX
		case Y_Rotation:
			moveType = MoveY
		case Z_Rotation:
			moveType = MoveZ
		default:
			return MoveR, 0 // Default fallback
		}
	} else {
		// Handle face moves
		switch move.Face {
		case Right:
			moveType = MoveR
		case Left:
			moveType = MoveL
		case Up:
			moveType = MoveU
		case Down:
			moveType = MoveD
		case Front:
			moveType = MoveF
		case Back:
			moveType = MoveB
		default:
			return MoveR, 0 // Default fallback
		}
	}

	// Determine quarter turns
	if move.Double {
		quarterTurns = 2
	} else if move.Clockwise {
		quarterTurns = 1 // Clockwise = 1 quarter turn
	} else {
		quarterTurns = 3 // Counter-clockwise = 3 quarter turns clockwise
	}

	return moveType, quarterTurns
}

// getAffectedLayers determines which layers are affected by a move
func getAffectedLayers(move Move, N int) []int {
	// Handle slice moves
	if move.Slice != NoSlice {
		if N%2 == 0 {
			return []int{} // Slice moves undefined for even cubes
		}
		return []int{N / 2} // Middle layer
	}

	// Handle cube rotations (affect all layers)
	if move.Rotation != NoRotation {
		layers := make([]int, N)
		for i := 0; i < N; i++ {
			layers[i] = i
		}
		return layers
	}

	// Handle face moves
	if move.Wide {
		// Wide moves affect outer two layers by default
		depth := move.WideDepth
		if depth <= 0 {
			depth = 2
		}
		layers := make([]int, depth)
		for i := 0; i < depth; i++ {
			layers[i] = i
		}
		return layers
	} else if move.Layer > 0 {
		// Layer-specific move
		return []int{move.Layer}
	} else {
		// Standard face move (outer layer only)
		return []int{0}
	}
}

// newApplyMove applies a move using the new permutation system
func (c *Cube) newApplyMove(move Move) {
	moveType, quarterTurns := moveToMoveType(move)
	layers := getAffectedLayers(move, c.Size)

	for _, layer := range layers {
		perm := getPermutation(c.Size, moveType, layer, quarterTurns)
		applyPermutation(c, perm)
	}
}

// ParseMove parses a move from advanced notation
// Supports: R, U', F2, 2R, Rw, 2Fw, M, E', S2, x, y', z2
func ParseMove(notation string) (Move, error) {
	notation = strings.TrimSpace(notation)
	if len(notation) == 0 {
		return Move{}, fmt.Errorf("empty move notation")
	}

	move := Move{Clockwise: true, WideDepth: 2} // Default wide depth is 2 layers
	i := 0

	// Check for slice moves first (M, E, S)
	if len(notation) >= 1 {
		switch strings.ToUpper(notation)[0] {
		case 'M':
			move.Slice = M_Slice
			i = 1
		case 'E':
			move.Slice = E_Slice
			i = 1
		case 'S':
			move.Slice = S_Slice
			i = 1
		}
	}

	// Check for cube rotations (x, y, z) - case sensitive
	if move.Slice == NoSlice && len(notation) >= 1 {
		switch notation[0] {
		case 'x':
			move.Rotation = X_Rotation
			i = 1
		case 'y':
			move.Rotation = Y_Rotation
			i = 1
		case 'z':
			move.Rotation = Z_Rotation
			i = 1
		}
	}

	// If not slice or rotation, parse face moves
	if move.Slice == NoSlice && move.Rotation == NoRotation {
		// Check for layer number at start (2R, 3L, etc.)
		if i < len(notation) && notation[i] >= '2' && notation[i] <= '9' {
			layer, err := strconv.Atoi(string(notation[i]))
			if err != nil {
				return Move{}, fmt.Errorf("invalid layer number: %c", notation[i])
			}
			move.Layer = layer - 1 // Convert to 0-based
			i++
		}

		// Parse face
		if i >= len(notation) {
			return Move{}, fmt.Errorf("missing face in notation")
		}

		var face Face
		switch strings.ToUpper(notation)[i] {
		case 'F':
			face = Front
		case 'B':
			face = Back
		case 'L':
			face = Left
		case 'R':
			face = Right
		case 'U':
			face = Up
		case 'D':
			face = Down
		default:
			return Move{}, fmt.Errorf("invalid face: %c", notation[i])
		}
		move.Face = face
		i++

		// Check for wide move marker (w)
		if i < len(notation) && strings.ToLower(notation)[i] == 'w' {
			move.Wide = true
			i++
		}
	}

	// Parse modifiers (', 2)
	for i < len(notation) {
		switch notation[i] {
		case '\'':
			move.Clockwise = false
		case '2':
			move.Double = true
		default:
			return Move{}, fmt.Errorf("invalid modifier: %c", notation[i])
		}
		i++
	}

	return move, nil
}

// ParseScramble parses a scramble string into a slice of moves
func ParseScramble(scramble string) ([]Move, error) {
	if scramble == "" {
		return []Move{}, nil
	}

	parts := strings.Fields(scramble)
	moves := make([]Move, 0, len(parts))

	for _, part := range parts {
		move, err := ParseMove(part)
		if err != nil {
			return nil, fmt.Errorf("error parsing move '%s': %v", part, err)
		}
		moves = append(moves, move)
	}

	return moves, nil
}

// ApplyMove applies a single move to the cube with advanced notation support
func (c *Cube) ApplyMove(move Move) {
	// Use legacy system only if explicitly enabled
	if useLegacyMoves {
		c.oldApplyMove(move)
		return
	}

	// Default to new permutation system
	c.newApplyMove(move)
}

// oldApplyMove is the original implementation
func (c *Cube) oldApplyMove(move Move) {
	// Handle slice moves
	if move.Slice != NoSlice {
		c.applySliceMove(move)
		return
	}

	// Handle cube rotations
	if move.Rotation != NoRotation {
		c.applyCubeRotation(move)
		return
	}

	// Handle face moves (including wide and layer moves)
	if move.Wide {
		c.applyWideMove(move)
	} else if move.Layer > 0 {
		c.applyLayerMove(move)
	} else {
		// Standard face move
		if move.Double {
			c.rotateFace(move.Face, true)
			c.rotateFace(move.Face, true)
		} else {
			c.rotateFace(move.Face, move.Clockwise)
		}
	}
}

// ApplyMoves applies a sequence of moves to the cube
func (c *Cube) ApplyMoves(moves []Move) {
	for _, move := range moves {
		c.ApplyMove(move)
	}
}

// rotateFace rotates a face and its adjacent edges
func (c *Cube) rotateFace(face Face, clockwise bool) {
	// Rotate the face itself
	c.rotateFaceMatrix(int(face), clockwise)

	// Rotate adjacent edges
	c.rotateAdjacentEdges(face, clockwise)
}

// rotateFaceMatrix rotates a face matrix 90 degrees
func (c *Cube) rotateFaceMatrix(face int, clockwise bool) {
	size := c.Size
	temp := make([][]Color, size)
	for i := range temp {
		temp[i] = make([]Color, size)
	}

	// Copy current face
	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			temp[i][j] = c.Faces[face][i][j]
		}
	}

	// Rotate
	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			if clockwise {
				c.Faces[face][i][j] = temp[size-1-j][i]
			} else {
				c.Faces[face][i][j] = temp[j][size-1-i]
			}
		}
	}
}

// rotateAdjacentEdges rotates the edges adjacent to a face
// For standard face moves, only rotate the outermost layer (layer 0)
func (c *Cube) rotateAdjacentEdges(face Face, clockwise bool) {
	// Standard face moves (R, U, F, etc.) should only affect the outermost layer
	// Wide moves and layer moves are handled separately
	c.rotateAdjacentEdgesLayer(face, clockwise, 0)
}

// rotateAdjacentEdgesLayer rotates the edges for a specific layer
func (c *Cube) rotateAdjacentEdgesLayer(face Face, clockwise bool, layer int) {
	size := c.Size
	temp := make([]Color, size)

	switch face {
	case Front:
		// Front face: rotate edges between Up, Right, Down, Left
		// Layer 0 = outermost, layer 1 = second from outside, etc.
		upRow := size - 1 - layer
		downRow := layer
		leftCol := size - 1 - layer
		rightCol := layer

		if clockwise {
			// Save Up row (upRow)
			for i := 0; i < size; i++ {
				temp[i] = c.Faces[Up][upRow][i]
			}
			// Up ← Left column (leftCol, reversed)
			for i := 0; i < size; i++ {
				c.Faces[Up][upRow][i] = c.Faces[Left][size-1-i][leftCol]
			}
			// Left column ← Down row (downRow, reversed)
			for i := 0; i < size; i++ {
				c.Faces[Left][size-1-i][leftCol] = c.Faces[Down][downRow][i]
			}
			// Down row ← Right column (rightCol, reversed)
			for i := 0; i < size; i++ {
				c.Faces[Down][downRow][i] = c.Faces[Right][size-1-i][rightCol]
			}
			// Right column ← Up row (saved, reversed for correct visual orientation)
			for i := 0; i < size; i++ {
				c.Faces[Right][i][rightCol] = temp[size-1-i]
			}
		} else {
			// Save Up row
			for i := 0; i < size; i++ {
				temp[i] = c.Faces[Up][upRow][i]
			}
			// Up ← Right column
			for i := 0; i < size; i++ {
				c.Faces[Up][upRow][i] = c.Faces[Right][i][rightCol]
			}
			// Right column ← Down row (reversed for correct visual orientation)
			for i := 0; i < size; i++ {
				c.Faces[Right][i][rightCol] = c.Faces[Down][downRow][size-1-i]
			}
			// Down row ← Left column (reversed)
			for i := 0; i < size; i++ {
				c.Faces[Down][downRow][i] = c.Faces[Left][size-1-i][leftCol]
			}
			// Left column ← Up row (saved, reversed for correct visual orientation)
			for i := 0; i < size; i++ {
				c.Faces[Left][i][leftCol] = temp[size-1-i]
			}
		}

	case Back:
		// Back face: rotate edges between Up, Right, Down, Left (clockwise when viewed from back)
		upRow := layer
		downRow := size - 1 - layer
		leftCol := layer
		rightCol := size - 1 - layer

		if clockwise {
			// Save Up row
			for i := 0; i < size; i++ {
				temp[i] = c.Faces[Up][upRow][i]
			}
			// Up ← Right column (reversed)
			for i := 0; i < size; i++ {
				c.Faces[Up][upRow][i] = c.Faces[Right][size-1-i][rightCol]
			}
			// Right column ← Down row (reversed)
			for i := 0; i < size; i++ {
				c.Faces[Right][size-1-i][rightCol] = c.Faces[Down][downRow][size-1-i]
			}
			// Down row ← Left column (reversed)
			for i := 0; i < size; i++ {
				c.Faces[Down][downRow][i] = c.Faces[Left][size-1-i][leftCol]
			}
			// Left column ← Up row (saved)
			for i := 0; i < size; i++ {
				c.Faces[Left][i][leftCol] = temp[i]
			}
		} else {
			// Save Up row
			for i := 0; i < size; i++ {
				temp[i] = c.Faces[Up][upRow][i]
			}
			// Up ← Left column (reversed)
			for i := 0; i < size; i++ {
				c.Faces[Up][upRow][i] = c.Faces[Left][size-1-i][leftCol]
			}
			// Left column ← Down row (reversed)
			for i := 0; i < size; i++ {
				c.Faces[Left][i][leftCol] = c.Faces[Down][downRow][size-1-i]
			}
			// Down row ← Right column (reversed)
			for i := 0; i < size; i++ {
				c.Faces[Down][downRow][size-1-i] = c.Faces[Right][size-1-i][rightCol]
			}
			// Right column ← Up row (saved, NOT reversed)
			for i := 0; i < size; i++ {
				c.Faces[Right][i][rightCol] = temp[i]
			}
		}

	case Left:
		// Left face: rotate edges between Up, Front, Down, Back
		upCol := layer
		downCol := layer
		frontCol := layer
		backCol := size - 1 - layer

		if clockwise {
			// Save Up column
			for i := 0; i < size; i++ {
				temp[i] = c.Faces[Up][i][upCol]
			}
			// Up ← Back column (reversed)
			for i := 0; i < size; i++ {
				c.Faces[Up][i][upCol] = c.Faces[Back][size-1-i][backCol]
			}
			// Back column ← Down column (reversed)
			for i := 0; i < size; i++ {
				c.Faces[Back][size-1-i][backCol] = c.Faces[Down][i][downCol]
			}
			// Down column ← Front column
			for i := 0; i < size; i++ {
				c.Faces[Down][i][downCol] = c.Faces[Front][i][frontCol]
			}
			// Front column ← Up column (saved, reversed for correct visual orientation)
			for i := 0; i < size; i++ {
				c.Faces[Front][i][frontCol] = temp[size-1-i]
			}
		} else {
			// Save Up column
			for i := 0; i < size; i++ {
				temp[i] = c.Faces[Up][i][upCol]
			}
			// Up ← Front column
			for i := 0; i < size; i++ {
				c.Faces[Up][i][upCol] = c.Faces[Front][i][frontCol]
			}
			// Front column ← Down column
			for i := 0; i < size; i++ {
				c.Faces[Front][i][frontCol] = c.Faces[Down][i][downCol]
			}
			// Down column ← Back column (reversed)
			for i := 0; i < size; i++ {
				c.Faces[Down][i][downCol] = c.Faces[Back][size-1-i][backCol]
			}
			// Back column ← Up column (saved, reversed)
			for i := 0; i < size; i++ {
				c.Faces[Back][size-1-i][backCol] = temp[i]
			}
		}

	case Right:
		// Right face: rotate edges between Up, Back, Down, Front (clockwise from right side view)
		upCol := size - 1 - layer
		downCol := size - 1 - layer
		frontCol := size - 1 - layer
		backCol := layer

		if clockwise {
			// Save Up column
			for i := 0; i < size; i++ {
				temp[i] = c.Faces[Up][i][upCol]
			}
			// Up ← Front column
			for i := 0; i < size; i++ {
				c.Faces[Up][i][upCol] = c.Faces[Front][i][frontCol]
			}
			// Front column ← Down column
			for i := 0; i < size; i++ {
				c.Faces[Front][i][frontCol] = c.Faces[Down][i][downCol]
			}
			// Down column ← Back column (reversed)
			for i := 0; i < size; i++ {
				c.Faces[Down][i][downCol] = c.Faces[Back][size-1-i][backCol]
			}
			// Back column ← Up column (saved, reversed)
			for i := 0; i < size; i++ {
				c.Faces[Back][size-1-i][backCol] = temp[i]
			}
		} else {
			// Save Up column
			for i := 0; i < size; i++ {
				temp[i] = c.Faces[Up][i][upCol]
			}
			// Up ← Back column (reversed)
			for i := 0; i < size; i++ {
				c.Faces[Up][i][upCol] = c.Faces[Back][size-1-i][backCol]
			}
			// Back column ← Down column (reversed)
			for i := 0; i < size; i++ {
				c.Faces[Back][size-1-i][backCol] = c.Faces[Down][i][downCol]
			}
			// Down column ← Front column
			for i := 0; i < size; i++ {
				c.Faces[Down][i][downCol] = c.Faces[Front][i][frontCol]
			}
			// Front column ← Up column (saved)
			for i := 0; i < size; i++ {
				c.Faces[Front][i][frontCol] = temp[i]
			}
		}

	case Up:
		// Up face: rotate edges between Back, Right, Front, Left
		backRow := layer
		rightRow := layer
		frontRow := layer
		leftRow := layer

		if clockwise {
			// Save Back row
			for i := 0; i < size; i++ {
				temp[i] = c.Faces[Back][backRow][i]
			}
			// Back ← Left row
			for i := 0; i < size; i++ {
				c.Faces[Back][backRow][i] = c.Faces[Left][leftRow][i]
			}
			// Left ← Front row
			for i := 0; i < size; i++ {
				c.Faces[Left][leftRow][i] = c.Faces[Front][frontRow][i]
			}
			// Front ← Right row
			for i := 0; i < size; i++ {
				c.Faces[Front][frontRow][i] = c.Faces[Right][rightRow][i]
			}
			// Right ← Back row (saved)
			for i := 0; i < size; i++ {
				c.Faces[Right][rightRow][i] = temp[i]
			}
		} else {
			// Save Back row
			for i := 0; i < size; i++ {
				temp[i] = c.Faces[Back][backRow][i]
			}
			// Back ← Right row
			for i := 0; i < size; i++ {
				c.Faces[Back][backRow][i] = c.Faces[Right][rightRow][i]
			}
			// Right ← Front row
			for i := 0; i < size; i++ {
				c.Faces[Right][rightRow][i] = c.Faces[Front][frontRow][i]
			}
			// Front ← Left row
			for i := 0; i < size; i++ {
				c.Faces[Front][frontRow][i] = c.Faces[Left][leftRow][i]
			}
			// Left ← Back row (saved)
			for i := 0; i < size; i++ {
				c.Faces[Left][leftRow][i] = temp[i]
			}
		}

	case Down:
		// Down face: rotate edges between Front, Right, Back, Left
		frontRow := size - 1 - layer
		rightRow := size - 1 - layer
		backRow := size - 1 - layer
		leftRow := size - 1 - layer

		if clockwise {
			// Save Front row
			for i := 0; i < size; i++ {
				temp[i] = c.Faces[Front][frontRow][i]
			}
			// Front ← Left row
			for i := 0; i < size; i++ {
				c.Faces[Front][frontRow][i] = c.Faces[Left][leftRow][i]
			}
			// Left ← Back row
			for i := 0; i < size; i++ {
				c.Faces[Left][leftRow][i] = c.Faces[Back][backRow][i]
			}
			// Back ← Right row
			for i := 0; i < size; i++ {
				c.Faces[Back][backRow][i] = c.Faces[Right][rightRow][i]
			}
			// Right ← Front row (saved)
			for i := 0; i < size; i++ {
				c.Faces[Right][rightRow][i] = temp[i]
			}
		} else {
			// Save Front row
			for i := 0; i < size; i++ {
				temp[i] = c.Faces[Front][frontRow][i]
			}
			// Front ← Right row
			for i := 0; i < size; i++ {
				c.Faces[Front][frontRow][i] = c.Faces[Right][rightRow][i]
			}
			// Right ← Back row
			for i := 0; i < size; i++ {
				c.Faces[Right][rightRow][i] = c.Faces[Back][backRow][i]
			}
			// Back ← Left row
			for i := 0; i < size; i++ {
				c.Faces[Back][backRow][i] = c.Faces[Left][leftRow][i]
			}
			// Left ← Front row (saved)
			for i := 0; i < size; i++ {
				c.Faces[Left][leftRow][i] = temp[i]
			}
		}
	}
}

// applySliceMove applies middle slice moves (M, E, S)
func (c *Cube) applySliceMove(move Move) {
	if move.Double {
		c.executeSliceMove(move.Slice, move.Clockwise)
		c.executeSliceMove(move.Slice, move.Clockwise)
	} else {
		c.executeSliceMove(move.Slice, move.Clockwise)
	}
}

// executeSliceMove performs the actual slice rotation
func (c *Cube) executeSliceMove(slice SliceType, clockwise bool) {
	size := c.Size
	temp := make([]Color, size)

	switch slice {
	case M_Slice:
		// M slice: between L and R faces, moves like L but affects middle
		// Only works on odd-sized cubes (3x3, 5x5), affects center column
		if size%2 == 0 {
			return // M slice undefined for even cubes
		}
		centerCol := size / 2

		if clockwise {
			// M moves like L: Up ← Back ← Down ← Front ← Up
			for i := 0; i < size; i++ {
				temp[i] = c.Faces[Up][i][centerCol]
			}
			for i := 0; i < size; i++ {
				c.Faces[Up][i][centerCol] = c.Faces[Back][size-1-i][centerCol]
			}
			for i := 0; i < size; i++ {
				c.Faces[Back][size-1-i][centerCol] = c.Faces[Down][size-1-i][centerCol]
			}
			for i := 0; i < size; i++ {
				c.Faces[Down][i][centerCol] = c.Faces[Front][i][centerCol]
			}
			for i := 0; i < size; i++ {
				c.Faces[Front][i][centerCol] = temp[i]
			}
		} else {
			// M' moves opposite to M
			for i := 0; i < size; i++ {
				temp[i] = c.Faces[Up][i][centerCol]
			}
			for i := 0; i < size; i++ {
				c.Faces[Up][i][centerCol] = c.Faces[Front][i][centerCol]
			}
			for i := 0; i < size; i++ {
				c.Faces[Front][i][centerCol] = c.Faces[Down][i][centerCol]
			}
			for i := 0; i < size; i++ {
				c.Faces[Down][size-1-i][centerCol] = c.Faces[Back][size-1-i][centerCol]
			}
			for i := 0; i < size; i++ {
				c.Faces[Back][size-1-i][centerCol] = temp[i]
			}
		}

	case E_Slice:
		// E slice: between U and D faces, moves like D but affects middle
		if size%2 == 0 {
			return // E slice undefined for even cubes
		}
		centerRow := size / 2

		if clockwise {
			// E moves like D: Front ← Left ← Back ← Right ← Front
			for i := 0; i < size; i++ {
				temp[i] = c.Faces[Front][centerRow][i]
			}
			for i := 0; i < size; i++ {
				c.Faces[Front][centerRow][i] = c.Faces[Left][centerRow][i]
			}
			for i := 0; i < size; i++ {
				c.Faces[Left][centerRow][i] = c.Faces[Back][centerRow][i]
			}
			for i := 0; i < size; i++ {
				c.Faces[Back][centerRow][i] = c.Faces[Right][centerRow][i]
			}
			for i := 0; i < size; i++ {
				c.Faces[Right][centerRow][i] = temp[i]
			}
		} else {
			// E' moves opposite
			for i := 0; i < size; i++ {
				temp[i] = c.Faces[Front][centerRow][i]
			}
			for i := 0; i < size; i++ {
				c.Faces[Front][centerRow][i] = c.Faces[Right][centerRow][i]
			}
			for i := 0; i < size; i++ {
				c.Faces[Right][centerRow][i] = c.Faces[Back][centerRow][i]
			}
			for i := 0; i < size; i++ {
				c.Faces[Back][centerRow][i] = c.Faces[Left][centerRow][i]
			}
			for i := 0; i < size; i++ {
				c.Faces[Left][centerRow][i] = temp[i]
			}
		}

	case S_Slice:
		// S slice: between F and B faces, moves like F but affects middle
		if size%2 == 0 {
			return // S slice undefined for even cubes
		}
		centerLayer := size / 2

		if clockwise {
			// S moves like F
			for i := 0; i < size; i++ {
				temp[i] = c.Faces[Up][size-1-centerLayer][i]
			}
			for i := 0; i < size; i++ {
				c.Faces[Up][size-1-centerLayer][i] = c.Faces[Left][size-1-i][centerLayer]
			}
			for i := 0; i < size; i++ {
				c.Faces[Left][size-1-i][centerLayer] = c.Faces[Down][centerLayer][size-1-i]
			}
			for i := 0; i < size; i++ {
				c.Faces[Down][centerLayer][i] = c.Faces[Right][i][size-1-centerLayer]
			}
			for i := 0; i < size; i++ {
				c.Faces[Right][i][size-1-centerLayer] = temp[i]
			}
		} else {
			// S' moves opposite
			for i := 0; i < size; i++ {
				temp[i] = c.Faces[Up][size-1-centerLayer][i]
			}
			for i := 0; i < size; i++ {
				c.Faces[Up][size-1-centerLayer][i] = c.Faces[Right][i][size-1-centerLayer]
			}
			for i := 0; i < size; i++ {
				c.Faces[Right][i][size-1-centerLayer] = c.Faces[Down][centerLayer][size-1-i]
			}
			for i := 0; i < size; i++ {
				c.Faces[Down][centerLayer][size-1-i] = c.Faces[Left][size-1-i][centerLayer]
			}
			for i := 0; i < size; i++ {
				c.Faces[Left][i][centerLayer] = temp[size-1-i]
			}
		}
	}
}

// applyWideMove applies wide moves (Rw, Fw, etc.)
func (c *Cube) applyWideMove(move Move) {
	// Wide move = face move + adjacent inner layers
	// For Rw: apply R + inner layer moves
	layers := move.WideDepth
	if layers <= 0 {
		layers = 2 // Default to 2 layers
	}

	for layer := 0; layer < layers; layer++ {
		if layer == 0 {
			// Outermost layer includes face rotation
			if move.Double {
				c.rotateFace(move.Face, move.Clockwise)
				c.rotateFace(move.Face, move.Clockwise)
			} else {
				c.rotateFace(move.Face, move.Clockwise)
			}
		} else {
			// Inner layers - just edge rotation
			if move.Double {
				c.rotateAdjacentEdgesLayer(move.Face, move.Clockwise, layer)
				c.rotateAdjacentEdgesLayer(move.Face, move.Clockwise, layer)
			} else {
				c.rotateAdjacentEdgesLayer(move.Face, move.Clockwise, layer)
			}
		}
	}
}

// applyLayerMove applies layer-specific moves (2R, 3L, etc.)
func (c *Cube) applyLayerMove(move Move) {
	// Layer move affects only the specified inner layer (no face rotation)
	if move.Double {
		c.rotateAdjacentEdgesLayer(move.Face, move.Clockwise, move.Layer)
		c.rotateAdjacentEdgesLayer(move.Face, move.Clockwise, move.Layer)
	} else {
		c.rotateAdjacentEdgesLayer(move.Face, move.Clockwise, move.Layer)
	}
}

// applyCubeRotation applies whole cube rotations (x, y, z)
func (c *Cube) applyCubeRotation(move Move) {
	if move.Double {
		c.executeCubeRotation(move.Rotation, move.Clockwise)
		c.executeCubeRotation(move.Rotation, move.Clockwise)
	} else {
		c.executeCubeRotation(move.Rotation, move.Clockwise)
	}
}

// executeCubeRotation performs the actual cube rotation
func (c *Cube) executeCubeRotation(rotation RotationType, clockwise bool) {
	switch rotation {
	case X_Rotation:
		// x rotation: rotate around R face axis
		// This reorients the entire cube - faces change positions
		if clockwise {
			// F→D, D→B, B→U, U→F, L rotates CCW, R rotates CW
			c.swapFaces(Front, Down, Back, Up)
			c.rotateFaceMatrix(int(Left), false) // Left face rotates CCW
			c.rotateFaceMatrix(int(Right), true) // Right face rotates CW
		} else {
			// F→U, U→B, B→D, D→F, L rotates CW, R rotates CCW
			c.swapFaces(Front, Up, Back, Down)
			c.rotateFaceMatrix(int(Left), true)   // Left face rotates CW
			c.rotateFaceMatrix(int(Right), false) // Right face rotates CCW
		}

	case Y_Rotation:
		// y rotation: rotate around U face axis
		if clockwise {
			// F→L, L→B, B→R, R→F, U rotates CW, D rotates CCW
			c.swapFaces(Front, Left, Back, Right)
			c.rotateFaceMatrix(int(Up), true)    // Up face rotates CW
			c.rotateFaceMatrix(int(Down), false) // Down face rotates CCW
		} else {
			// F→R, R→B, B→L, L→F, U rotates CCW, D rotates CW
			c.swapFaces(Front, Right, Back, Left)
			c.rotateFaceMatrix(int(Up), false)  // Up face rotates CCW
			c.rotateFaceMatrix(int(Down), true) // Down face rotates CW
		}

	case Z_Rotation:
		// z rotation: rotate around F face axis
		if clockwise {
			// U→L, L→D, D→R, R→U, F rotates CW, B rotates CCW
			c.swapFaces(Up, Left, Down, Right)
			c.rotateFaceMatrix(int(Front), true) // Front face rotates CW
			c.rotateFaceMatrix(int(Back), false) // Back face rotates CCW
		} else {
			// U→R, R→D, D→L, L→U, F rotates CCW, B rotates CW
			c.swapFaces(Up, Right, Down, Left)
			c.rotateFaceMatrix(int(Front), false) // Front face rotates CCW
			c.rotateFaceMatrix(int(Back), true)   // Back face rotates CW
		}
	}
}

// swapFaces swaps the positions of four faces in a cycle
func (c *Cube) swapFaces(f1, f2, f3, f4 Face) {
	size := c.Size
	temp := make([][]Color, size)
	for i := range temp {
		temp[i] = make([]Color, size)
		for j := range temp[i] {
			temp[i][j] = c.Faces[f1][i][j]
		}
	}

	// f1 ← f4
	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			c.Faces[f1][i][j] = c.Faces[f4][i][j]
		}
	}
	// f4 ← f3
	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			c.Faces[f4][i][j] = c.Faces[f3][i][j]
		}
	}
	// f3 ← f2
	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			c.Faces[f3][i][j] = c.Faces[f2][i][j]
		}
	}
	// f2 ← f1 (saved)
	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			c.Faces[f2][i][j] = temp[i][j]
		}
	}
}
