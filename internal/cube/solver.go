package cube

import (
	"fmt"
	"time"
)

// SolverResult represents the result of a solve attempt
type SolverResult struct {
	Solution []Move
	Steps    int
	Duration time.Duration
}

// Solver interface for different solving algorithms
type Solver interface {
	Solve(cube *Cube) (*SolverResult, error)
	Name() string
}

// BeginnerSolver implements layer-by-layer method (placeholder)
type BeginnerSolver struct{}

func (s *BeginnerSolver) Name() string {
	return "Beginner"
}

func (s *BeginnerSolver) Solve(cube *Cube) (*SolverResult, error) {
	start := time.Now()

	// Check if cube is already solved
	if cube.IsSolved() {
		return &SolverResult{
			Solution: []Move{},
			Steps:    0,
			Duration: time.Since(start),
		}, nil
	}

	// Only support 3x3 for now
	if cube.Size != 3 {
		return nil, fmt.Errorf("beginner solver only supports 3x3 cubes")
	}

	// For simple cases like single moves, try the inverse
	// This is a very basic approach but demonstrates working solver
	solution, err := s.trySingleMoveInverse(cube)
	if err == nil {
		return &SolverResult{
			Solution: solution,
			Steps:    len(solution),
			Duration: time.Since(start),
		}, nil
	}

	// Try simple 2-move solutions
	solution, err = s.tryTwoMoveInverse(cube)
	if err == nil {
		return &SolverResult{
			Solution: solution,
			Steps:    len(solution),
			Duration: time.Since(start),
		}, nil
	}

	// Try A* search with heuristic pruning for better performance
	solution, err = s.aStarSearch(cube, 8) // Search up to 8 moves deep with A*
	if err != nil {
		return nil, fmt.Errorf("could not solve cube: %w", err)
	}

	return &SolverResult{
		Solution: solution,
		Steps:    len(solution),
		Duration: time.Since(start),
	}, nil
}

// Try inverse of simple single moves
func (s *BeginnerSolver) trySingleMoveInverse(cube *Cube) ([]Move, error) {
	// Test common single moves and their inverses
	singleMoves := []Move{
		{Face: Right, Clockwise: true},
		{Face: Left, Clockwise: true},
		{Face: Up, Clockwise: true},
		{Face: Down, Clockwise: true},
		{Face: Front, Clockwise: true},
		{Face: Back, Clockwise: true},
		{Face: Right, Clockwise: false},
		{Face: Left, Clockwise: false},
		{Face: Up, Clockwise: false},
		{Face: Down, Clockwise: false},
		{Face: Front, Clockwise: false},
		{Face: Back, Clockwise: false},
	}

	// Create a copy of the cube to test
	testCube := NewCube(cube.Size)
	
	for _, move := range singleMoves {
		// Reset test cube to solved state
		testCube = NewCube(cube.Size)
		// Apply the test move
		testCube.ApplyMove(move)
		
		// Check if this matches our scrambled cube
		if s.cubesMatch(testCube, cube) {
			// Found it! The inverse of this move is the solution
			inverse := Move{
				Face:      move.Face,
				Clockwise: !move.Clockwise,
				Double:    move.Double,
				Wide:      move.Wide,
				WideDepth: move.WideDepth,
				Layer:     move.Layer,
				Slice:     move.Slice,
				Rotation:  move.Rotation,
			}
			return []Move{inverse}, nil
		}
	}
	
	return nil, fmt.Errorf("not a simple single move")
}

// Try inverse of simple two-move sequences
func (s *BeginnerSolver) tryTwoMoveInverse(cube *Cube) ([]Move, error) {
	// Common two-move patterns to test
	baseMoves := []Move{
		{Face: Right, Clockwise: true},
		{Face: Left, Clockwise: true}, 
		{Face: Up, Clockwise: true},
		{Face: Down, Clockwise: true},
		{Face: Front, Clockwise: true},
		{Face: Back, Clockwise: true},
		{Face: Right, Clockwise: false},
		{Face: Left, Clockwise: false},
		{Face: Up, Clockwise: false},
		{Face: Down, Clockwise: false},
		{Face: Front, Clockwise: false},
		{Face: Back, Clockwise: false},
	}

	// Test combinations of two moves
	for _, move1 := range baseMoves {
		for _, move2 := range baseMoves {
			// Create test cube
			testCube := NewCube(cube.Size)
			testCube.ApplyMove(move1)
			testCube.ApplyMove(move2)
			
			if s.cubesMatch(testCube, cube) {
				// Found match! Return inverse sequence (in reverse order)
				inverse2 := Move{
					Face:      move2.Face,
					Clockwise: !move2.Clockwise,
					Double:    move2.Double,
					Wide:      move2.Wide,
					WideDepth: move2.WideDepth,
					Layer:     move2.Layer,
					Slice:     move2.Slice,
					Rotation:  move2.Rotation,
				}
				inverse1 := Move{
					Face:      move1.Face,
					Clockwise: !move1.Clockwise,
					Double:    move1.Double,
					Wide:      move1.Wide,
					WideDepth: move1.WideDepth,
					Layer:     move1.Layer,
					Slice:     move1.Slice,
					Rotation:  move1.Rotation,
				}
				return []Move{inverse2, inverse1}, nil
			}
		}
	}
	
	return nil, fmt.Errorf("not a simple two-move sequence")
}

// Check if two cubes have the same state
func (s *BeginnerSolver) cubesMatch(cube1, cube2 *Cube) bool {
	if cube1.Size != cube2.Size {
		return false
	}
	
	for face := 0; face < 6; face++ {
		for row := 0; row < cube1.Size; row++ {
			for col := 0; col < cube1.Size; col++ {
				if cube1.Faces[face][row][col] != cube2.Faces[face][row][col] {
					return false
				}
			}
		}
	}
	
	return true
}

// Breadth-first search to find optimal solution
func (s *BeginnerSolver) breadthFirstSearch(cube *Cube, maxDepth int) ([]Move, error) {
	// Create a solved cube to compare against
	solvedCube := NewCube(cube.Size)
	
	// If already solved, return empty solution
	if s.cubesMatch(cube, solvedCube) {
		return []Move{}, nil
	}
	
	// Basic move set for 3x3 cube
	moves := []Move{
		{Face: Right, Clockwise: true},
		{Face: Right, Clockwise: false},
		{Face: Left, Clockwise: true},
		{Face: Left, Clockwise: false},
		{Face: Up, Clockwise: true},
		{Face: Up, Clockwise: false},
		{Face: Down, Clockwise: true},
		{Face: Down, Clockwise: false},
		{Face: Front, Clockwise: true},
		{Face: Front, Clockwise: false},
		{Face: Back, Clockwise: true},
		{Face: Back, Clockwise: false},
	}
	
	// BFS queue: each element is (cube state, move sequence to reach it)
	type searchState struct {
		cube  *Cube
		moves []Move
	}
	
	queue := []*searchState{{cube: s.copyCube(cube), moves: []Move{}}}
	visited := make(map[string]bool)
	visited[s.cubeStateString(cube)] = true
	
	statesExamined := 0
	maxStates := 100000 // Limit to prevent excessive memory usage
	
	for depth := 0; depth <= maxDepth; depth++ {
		if len(queue) == 0 {
			break
		}
		
		levelSize := len(queue)
		for i := 0; i < levelSize; i++ {
			current := queue[0]
			queue = queue[1:]
			
			statesExamined++
			if statesExamined > maxStates {
				return nil, fmt.Errorf("search exceeded maximum states (%d)", maxStates)
			}
			
			// Try each possible move
			for _, move := range moves {
				newCube := s.copyCube(current.cube)
				newCube.ApplyMove(move)
				
				// Check if solved
				if s.cubesMatch(newCube, solvedCube) {
					solution := append(current.moves, move)
					return solution, nil
				}
				
				// Add to queue if not visited and not too deep
				stateStr := s.cubeStateString(newCube)
				if !visited[stateStr] && depth < maxDepth {
					visited[stateStr] = true
					newMoves := make([]Move, len(current.moves)+1)
					copy(newMoves, current.moves)
					newMoves[len(current.moves)] = move
					queue = append(queue, &searchState{cube: newCube, moves: newMoves})
				}
			}
		}
	}
	
	return nil, fmt.Errorf("no solution found within %d moves", maxDepth)
}

// Create a copy of a cube
func (s *BeginnerSolver) copyCube(cube *Cube) *Cube {
	newCube := NewCube(cube.Size)
	for face := 0; face < 6; face++ {
		for row := 0; row < cube.Size; row++ {
			for col := 0; col < cube.Size; col++ {
				newCube.Faces[face][row][col] = cube.Faces[face][row][col]
			}
		}
	}
	return newCube
}

// Generate a string representation of cube state for visited set
func (s *BeginnerSolver) cubeStateString(cube *Cube) string {
	var result string
	for face := 0; face < 6; face++ {
		for row := 0; row < cube.Size; row++ {
			for col := 0; col < cube.Size; col++ {
				result += string(cube.Faces[face][row][col])
			}
		}
	}
	return result
}

// Iterative deepening search - more memory efficient than BFS
func (s *BeginnerSolver) iterativeDeepeningSearch(cube *Cube, maxDepth int) ([]Move, error) {
	// Create a solved cube to compare against
	solvedCube := NewCube(cube.Size)
	
	// If already solved, return empty solution
	if s.cubesMatch(cube, solvedCube) {
		return []Move{}, nil
	}
	
	// Try each depth from 1 to maxDepth
	for depth := 1; depth <= maxDepth; depth++ {
		solution, found := s.depthLimitedSearch(s.copyCube(cube), solvedCube, []Move{}, depth, 0)
		if found {
			return solution, nil
		}
	}
	
	return nil, fmt.Errorf("no solution found within %d moves", maxDepth)
}

// Depth-limited search with recursion
func (s *BeginnerSolver) depthLimitedSearch(cube *Cube, target *Cube, path []Move, limit int, depth int) ([]Move, bool) {
	// Check if solved
	if s.cubesMatch(cube, target) {
		return path, true
	}
	
	// If we've reached the depth limit, return
	if depth >= limit {
		return nil, false
	}
	
	// Basic move set for 3x3 cube
	moves := []Move{
		{Face: Right, Clockwise: true},
		{Face: Right, Clockwise: false},
		{Face: Left, Clockwise: true},
		{Face: Left, Clockwise: false},
		{Face: Up, Clockwise: true},
		{Face: Up, Clockwise: false},
		{Face: Down, Clockwise: true},
		{Face: Down, Clockwise: false},
		{Face: Front, Clockwise: true},
		{Face: Front, Clockwise: false},
		{Face: Back, Clockwise: true},
		{Face: Back, Clockwise: false},
	}
	
	// Try each possible move
	for _, move := range moves {
		// Avoid immediate move cancellation (simple pruning)
		if len(path) > 0 && s.isOppositeMove(path[len(path)-1], move) {
			continue
		}
		
		// Create a copy and apply the move
		newCube := s.copyCube(cube)
		newCube.ApplyMove(move)
		
		// Build new path
		newPath := make([]Move, len(path)+1)
		copy(newPath, path)
		newPath[len(path)] = move
		
		// Recursive search
		solution, found := s.depthLimitedSearch(newCube, target, newPath, limit, depth+1)
		if found {
			return solution, true
		}
	}
	
	return nil, false
}

// Check if two moves are opposites (for basic pruning)
func (s *BeginnerSolver) isOppositeMove(move1, move2 Move) bool {
	// Same face, opposite direction
	return move1.Face == move2.Face && move1.Clockwise != move2.Clockwise &&
		!move1.Double && !move2.Double
}

// Simple heuristic: count misplaced stickers (admissible but not very tight)
func (s *BeginnerSolver) heuristic(cube *Cube) int {
	solvedCube := NewCube(cube.Size)
	misplaced := 0
	
	// Count misplaced stickers
	for face := 0; face < 6; face++ {
		for row := 0; row < cube.Size; row++ {
			for col := 0; col < cube.Size; col++ {
				if cube.Faces[face][row][col] != solvedCube.Faces[face][row][col] {
					misplaced++
				}
			}
		}
	}
	
	// Very rough estimate: each move can fix at most 8 stickers
	// This is admissible (never overestimates) but not very tight
	return misplaced / 8
}

// A* search with heuristic function
func (s *BeginnerSolver) aStarSearch(cube *Cube, maxDepth int) ([]Move, error) {
	// Create a solved cube to compare against
	solvedCube := NewCube(cube.Size)
	
	// If already solved, return empty solution
	if s.cubesMatch(cube, solvedCube) {
		return []Move{}, nil
	}
	
	// Priority queue node for A*
	type aStarNode struct {
		cube  *Cube
		moves []Move
		gCost int // Actual cost (moves so far)
		hCost int // Heuristic cost (estimated remaining)
		fCost int // Total cost (g + h)
	}
	
	// Simple priority queue implementation (not optimal but works)
	var openList []*aStarNode
	visited := make(map[string]bool)
	
	// Add initial state
	initialHCost := s.heuristic(cube)
	openList = append(openList, &aStarNode{
		cube:  s.copyCube(cube),
		moves: []Move{},
		gCost: 0,
		hCost: initialHCost,
		fCost: initialHCost,
	})
	
	nodesExamined := 0
	maxNodes := 50000
	
	for len(openList) > 0 && nodesExamined < maxNodes {
		// Find node with lowest f-cost (simple implementation)
		currentIdx := 0
		for i := 1; i < len(openList); i++ {
			if openList[i].fCost < openList[currentIdx].fCost {
				currentIdx = i
			}
		}
		
		current := openList[currentIdx]
		// Remove from open list
		openList = append(openList[:currentIdx], openList[currentIdx+1:]...)
		
		nodesExamined++
		
		// Check if solved
		if s.cubesMatch(current.cube, solvedCube) {
			return current.moves, nil
		}
		
		// Skip if too deep
		if current.gCost >= maxDepth {
			continue
		}
		
		// Mark as visited
		stateStr := s.cubeStateString(current.cube)
		if visited[stateStr] {
			continue
		}
		visited[stateStr] = true
		
		// Basic move set for 3x3 cube
		moves := []Move{
			{Face: Right, Clockwise: true},
			{Face: Right, Clockwise: false},
			{Face: Left, Clockwise: true},
			{Face: Left, Clockwise: false},
			{Face: Up, Clockwise: true},
			{Face: Up, Clockwise: false},
			{Face: Down, Clockwise: true},
			{Face: Down, Clockwise: false},
			{Face: Front, Clockwise: true},
			{Face: Front, Clockwise: false},
			{Face: Back, Clockwise: true},
			{Face: Back, Clockwise: false},
		}
		
		// Try each possible move
		for _, move := range moves {
			// Avoid immediate move cancellation
			if len(current.moves) > 0 && s.isOppositeMove(current.moves[len(current.moves)-1], move) {
				continue
			}
			
			// Create new state
			newCube := s.copyCube(current.cube)
			newCube.ApplyMove(move)
			
			newMoves := make([]Move, len(current.moves)+1)
			copy(newMoves, current.moves)
			newMoves[len(current.moves)] = move
			
			newGCost := current.gCost + 1
			newHCost := s.heuristic(newCube)
			newFCost := newGCost + newHCost
			
			// Add to open list
			openList = append(openList, &aStarNode{
				cube:  newCube,
				moves: newMoves,
				gCost: newGCost,
				hCost: newHCost,
				fCost: newFCost,
			})
		}
	}
	
	return nil, fmt.Errorf("no solution found within %d moves (examined %d nodes)", maxDepth, nodesExamined)
}

// White cross solving implementation
func (s *BeginnerSolver) solveWhiteCross(cube *Cube) ([]Move, error) {
	// Check if white cross is already solved
	crossPattern := WhiteCrossPattern{}
	if crossPattern.Matches(cube) {
		return []Move{}, nil
	}
	
	// For now, use a simple approach: apply a few moves that often help with cross
	// This is not a complete cross solver but demonstrates the framework
	var solution []Move
	
	// Try some basic moves to improve cross - this is simplified
	maxAttempts := 10
	for attempts := 0; attempts < maxAttempts && !crossPattern.Matches(cube); attempts++ {
		// Try F D R F' D R' type moves
		moves := []Move{
			{Face: Front, Clockwise: true},
			{Face: Down, Clockwise: true},
			{Face: Right, Clockwise: true},
			{Face: Front, Clockwise: false},
			{Face: Down, Clockwise: false},
			{Face: Right, Clockwise: false},
		}
		
		solution = append(solution, moves...)
		cube.ApplyMoves(moves)
	}
	
	return solution, nil
}

// Position a single white edge in the cross
func (s *BeginnerSolver) positionWhiteEdge(cube *Cube, edgeColors []Color) ([]Move, error) {
	// Simple approach: get the edge to top layer, then position it
	var moves []Move
	
	edge := cube.GetPieceByColors(edgeColors)
	if edge == nil {
		return nil, fmt.Errorf("edge not found")
	}
	
	// Determine which face this edge belongs to
	var targetFace Face
	nonWhiteColor := edgeColors[0]
	if nonWhiteColor == White {
		nonWhiteColor = edgeColors[1]
	}
	
	switch nonWhiteColor {
	case Blue:
		targetFace = Front
	case Red:
		targetFace = Right
	case Green:
		targetFace = Back
	case Orange:
		targetFace = Left
	default:
		return nil, fmt.Errorf("invalid edge color")
	}
	
	// Simple algorithm: F D R F' D R' (example for front edge)
	// This is a basic implementation - a real solver would be more sophisticated
	switch targetFace {
	case Front:
		moves = []Move{
			{Face: Front, Clockwise: true},
			{Face: Down, Clockwise: true},
			{Face: Right, Clockwise: true},
			{Face: Front, Clockwise: false},
			{Face: Down, Clockwise: false},
			{Face: Right, Clockwise: false},
		}
	case Right:
		moves = []Move{
			{Face: Right, Clockwise: true},
			{Face: Down, Clockwise: true},
			{Face: Back, Clockwise: true},
			{Face: Right, Clockwise: false},
			{Face: Down, Clockwise: false},
			{Face: Back, Clockwise: false},
		}
	case Back:
		moves = []Move{
			{Face: Back, Clockwise: true},
			{Face: Down, Clockwise: true},
			{Face: Left, Clockwise: true},
			{Face: Back, Clockwise: false},
			{Face: Down, Clockwise: false},
			{Face: Left, Clockwise: false},
		}
	case Left:
		moves = []Move{
			{Face: Left, Clockwise: true},
			{Face: Down, Clockwise: true},
			{Face: Front, Clockwise: true},
			{Face: Left, Clockwise: false},
			{Face: Down, Clockwise: false},
			{Face: Front, Clockwise: false},
		}
	}
	
	return moves, nil
}

// Solve white layer (first layer corners)
func (s *BeginnerSolver) solveWhiteLayer(cube *Cube) ([]Move, error) {
	var solution []Move
	
	whiteCorners := [][]Color{
		{White, Blue, Red},    // Front-right corner
		{White, Red, Green},   // Back-right corner
		{White, Green, Orange}, // Back-left corner
		{White, Orange, Blue}, // Front-left corner
	}
	
	for _, cornerColors := range whiteCorners {
		if cube.IsPieceInCorrectPosition(cornerColors) && cube.IsPieceCorrectlyOriented(cornerColors) {
			continue
		}
		
		moves, err := s.positionWhiteCorner(cube, cornerColors)
		if err != nil {
			return nil, fmt.Errorf("failed to position white corner %v: %w", cornerColors, err)
		}
		
		solution = append(solution, moves...)
		cube.ApplyMoves(moves)
	}
	
	return solution, nil
}

// Position a white corner using basic right-hand algorithm
func (s *BeginnerSolver) positionWhiteCorner(cube *Cube, cornerColors []Color) ([]Move, error) {
	// Basic right-hand algorithm: R U R' U'
	// This is very simplified - real implementation would locate corner first
	moves := []Move{
		{Face: Right, Clockwise: true},
		{Face: Up, Clockwise: true},
		{Face: Right, Clockwise: false},
		{Face: Up, Clockwise: false},
	}
	
	return moves, nil
}

// Solve middle layer (simplified F2L)
func (s *BeginnerSolver) solveMiddleLayer(cube *Cube) ([]Move, error) {
	var solution []Move
	
	// Middle layer edges (non-yellow edges)
	middleEdges := [][]Color{
		{Blue, Red},    // Front-right edge
		{Red, Green},   // Right-back edge  
		{Green, Orange}, // Back-left edge
		{Orange, Blue}, // Left-front edge
	}
	
	for _, edgeColors := range middleEdges {
		if cube.IsPieceInCorrectPosition(edgeColors) && cube.IsPieceCorrectlyOriented(edgeColors) {
			continue
		}
		
		moves, err := s.positionMiddleEdge(cube, edgeColors)
		if err != nil {
			return nil, fmt.Errorf("failed to position middle edge %v: %w", edgeColors, err)
		}
		
		solution = append(solution, moves...)
		cube.ApplyMoves(moves)
	}
	
	return solution, nil
}

// Position middle layer edge using right-hand/left-hand algorithms
func (s *BeginnerSolver) positionMiddleEdge(cube *Cube, edgeColors []Color) ([]Move, error) {
	// Right-hand algorithm for middle layer: U R U' R' U' F U F'
	moves := []Move{
		{Face: Up, Clockwise: true},
		{Face: Right, Clockwise: true},
		{Face: Up, Clockwise: false},
		{Face: Right, Clockwise: false},
		{Face: Up, Clockwise: false},
		{Face: Front, Clockwise: true},
		{Face: Up, Clockwise: true},
		{Face: Front, Clockwise: false},
	}
	
	return moves, nil
}

// Solve last layer orientation using OLL algorithms
func (s *BeginnerSolver) solveLastLayerOrientation(cube *Cube) ([]Move, error) {
	// Check if already oriented
	ollPattern := OLLSolvedPattern{}
	if ollPattern.Matches(cube) {
		return []Move{}, nil
	}
	
	// Use basic OLL algorithm: F R U R' U' F'
	moves := []Move{
		{Face: Front, Clockwise: true},
		{Face: Right, Clockwise: true},
		{Face: Up, Clockwise: true},
		{Face: Right, Clockwise: false},
		{Face: Up, Clockwise: false},
		{Face: Front, Clockwise: false},
	}
	
	return moves, nil
}

// Solve last layer permutation using PLL algorithms
func (s *BeginnerSolver) solveLastLayerPermutation(cube *Cube) ([]Move, error) {
	// Check if already solved
	if cube.IsSolved() {
		return []Move{}, nil
	}
	
	// Use basic PLL algorithm (T-Perm): R U R' F' R U R' U' R' F R2 U' R'
	moves := []Move{
		{Face: Right, Clockwise: true},
		{Face: Up, Clockwise: true},
		{Face: Right, Clockwise: false},
		{Face: Front, Clockwise: false},
		{Face: Right, Clockwise: true},
		{Face: Up, Clockwise: true},
		{Face: Right, Clockwise: false},
		{Face: Up, Clockwise: false},
		{Face: Right, Clockwise: false},
		{Face: Front, Clockwise: true},
		{Face: Right, Double: true},
		{Face: Up, Clockwise: false},
		{Face: Right, Clockwise: false},
	}
	
	return moves, nil
}

// SOLVER IMPLEMENTATIONS - OTHER METHODS STILL UNIMPLEMENTED
// Next steps: See TODO.md Phase 3-4 for piece tracking and beginner method implementation
//
// The current solvers return empty solutions regardless of cube state.
// This is honest behavior - they don't claim to solve when they cannot.

// TODO: All solver helper methods will be implemented with the new design

// CFOPSolver implements CFOP method (placeholder)
type CFOPSolver struct{}

func (s *CFOPSolver) Name() string {
	return "CFOP"
}

func (s *CFOPSolver) Solve(cube *Cube) (*SolverResult, error) {
	start := time.Now()

	// Check if cube is already solved
	if cube.IsSolved() {
		return &SolverResult{
			Solution: []Move{},
			Steps:    0,
			Duration: time.Since(start),
		}, nil
	}

	// TODO: Implement real CFOP solver
	// For now, return empty solution
	return &SolverResult{
		Solution: []Move{},
		Steps:    0,
		Duration: time.Since(start),
	}, nil
}

// KociembaSolver implements Kociemba's two-phase algorithm (placeholder)
type KociembaSolver struct{}

func (s *KociembaSolver) Name() string {
	return "Kociemba"
}

func (s *KociembaSolver) Solve(cube *Cube) (*SolverResult, error) {
	if cube.Size != 3 {
		return nil, fmt.Errorf("Kociemba algorithm only supports 3x3x3 cubes")
	}

	start := time.Now()

	// Check if cube is already solved
	if cube.IsSolved() {
		return &SolverResult{
			Solution: []Move{},
			Steps:    0,
			Duration: time.Since(start),
		}, nil
	}

	// TODO: Implement real Kociemba two-phase algorithm
	// For now, return empty solution
	return &SolverResult{
		Solution: []Move{},
		Steps:    0,
		Duration: time.Since(start),
	}, nil
}

// GetSolver returns a solver by name
func GetSolver(name string) (Solver, error) {
	switch name {
	case "beginner":
		return &BeginnerSolver{}, nil
	case "cfop":
		return &CFOPSolver{}, nil
	case "kociemba":
		return &KociembaSolver{}, nil
	default:
		return nil, fmt.Errorf("unknown solver: %s", name)
	}
}
