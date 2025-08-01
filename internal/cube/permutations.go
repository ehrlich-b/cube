package cube

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
	case MoveS:
		ring = ringS(N, layer)
	case MoveX:
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
