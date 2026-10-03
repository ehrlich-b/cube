package cube

// PERMUTATION-BASED MOVE SYSTEM - ALTERNATIVE IMPLEMENTATION
//
// This file implements an alternative move system using permutation matrices instead of
// the current coordinate-based approach. It pre-computes permutations for all moves and
// applies them via array indexing for potentially better performance.
//
// Status: Complete implementation but not used (current system uses rings.go)
// Activation: Set environment variable CUBE_USE_PERMUTATION_MOVES=true
// Performance: Potentially faster for repeated moves, more memory usage
//
// This system could be valuable for:
// - High-performance solving algorithms
// - Batch move application
// - Algorithm verification at scale

import "sync"

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
		// E follows D; ringE is ordered in the U direction.
		quarterTurns = (4 - quarterTurns) % 4
	case MoveS:
		ring = ringS(N, layer)
	case MoveX:
		return generateCubeRotationPermutation(N, MoveX, quarterTurns)
	case MoveY:
		return generateCubeRotationPermutation(N, MoveY, quarterTurns)
	case MoveZ:
		return generateCubeRotationPermutation(N, MoveZ, quarterTurns)
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

	// The far layer is also an outer face, viewed from the opposite side.
	// This matters for full-width moves and numbered turns through the cube.
	if layer == N-1 {
		opposite := map[MoveType]MoveType{
			MoveR: MoveL, MoveL: MoveR, MoveU: MoveD,
			MoveD: MoveU, MoveF: MoveB, MoveB: MoveF,
		}
		if other, ok := opposite[moveType]; ok {
			faceRotationPerm := generateFaceRotationPermutation(N, other, (4-quarterTurns)%4)
			for i, dst := range faceRotationPerm {
				if dst != i {
					perm[i] = dst
				}
			}
		}
	}

	return perm
}

// generateCubeRotationPermutation creates a rigid whole-cube permutation.
// x/y/z follow R/U/F respectively. Moving a face also changes its grid's
// orientation; copying row/column unchanged would split edges and corners.
func generateCubeRotationPermutation(N int, rotationType MoveType, quarterTurns int) Permutation {
	perm := make(Permutation, 6*N*N)
	quarterTurns = (quarterTurns%4 + 4) % 4
	last := N - 1
	for index := range perm {
		face, row, col := indexToCoord(index, N)
		for turn := 0; turn < quarterTurns; turn++ {
			switch rotationType {
			case MoveX:
				switch face {
				case Front:
					face = Up
				case Up:
					face, row, col = Back, last-row, last-col
				case Back:
					face, row, col = Down, last-row, last-col
				case Down:
					face = Front
				case Right:
					row, col = col, last-row
				case Left:
					row, col = last-col, row
				}
			case MoveY:
				switch face {
				case Front:
					face = Left
				case Left:
					face = Back
				case Back:
					face = Right
				case Right:
					face = Front
				case Up:
					row, col = col, last-row
				case Down:
					row, col = last-col, row
				}
			case MoveZ:
				switch face {
				case Up:
					face, row, col = Right, col, last-row
				case Right:
					face, row, col = Down, col, last-row
				case Down:
					face, row, col = Left, col, last-row
				case Left:
					face, row, col = Up, col, last-row
				case Front:
					row, col = col, last-row
				case Back:
					row, col = last-col, row
				}
			}
		}
		perm[index] = stickerIndex(face, row, col, N)
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

// rotateSlice rotates a slice of indices by quarterTurns
func rotateSlice(slice []int, quarterTurns int) []int {
	n := len(slice)
	if n == 0 {
		return slice
	}
	// Normalize quarterTurns to 0-3 range
	quarterTurns = quarterTurns % 4
	shift := (quarterTurns * n / 4) % n
	result := make([]int, n)
	for i := range slice {
		result[i] = slice[(i+shift)%n]
	}
	return result
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
