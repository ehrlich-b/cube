# Moves Refactor Design Document

## Overview

This document outlines the refactoring of `internal/cube/moves.go` from a branch-heavy, hard-coded implementation to a data-driven permutation-based system that scales elegantly to N×N×N cubes.

## Current Problems

1. **Brittle Geometry**: Each face has hand-coded `size-1-i` mirror math prone to off-by-one errors
2. **Duplicated Logic**: Every move has its own branch with copy-pasted rotation logic
3. **Poor Scalability**: Adding support for new cube sizes requires extensive code duplication
4. **Testing Nightmare**: Hard-coded edge logic makes comprehensive testing difficult

## Core Concepts

### Sticker Index Space
- Flatten all stickers into a single index space [0, 6*N*N)
- Mapping: `(face, row, col) → index = face*N² + row*N + col`
- Enables treating moves as permutations on indices

### Permutation Tables
- Each move is a permutation of sticker indices
- Generate tables on-demand based on:
  - Cube size (N)
  - Move type (R, L, U, D, F, B, M, E, S, x, y, z)
  - Layer depth (0 = outer, 1 = second layer, etc.)
  - Quarter turns (1, 2, or 3)

### Ring Coordinates
- Moves affect "rings" of stickers around cube edges
- Example for R move: Up→Back→Down→Front cycle
- Each ring is a sequence of (face, row, col) coordinates

## Implementation Plan

### Phase 1: Core Infrastructure

#### 1.1 Index Mapping Functions
```go
// Convert (face, row, col) to flat index
func stickerIndex(face Face, row, col, N int) int {
    return int(face)*N*N + row*N + col
}

// Convert flat index back to (face, row, col)
func indexToCoord(idx, N int) (Face, int, int) {
    face := Face(idx / (N * N))
    remainder := idx % (N * N)
    row := remainder / N
    col := remainder % N
    return face, row, col
}
```

#### 1.2 Coordinate Type
```go
type Coord struct {
    Face Face
    Row  int
    Col  int
}
```

### Phase 2: Ring Generators

#### 2.1 Face Ring Generators
Each face move affects stickers in a specific pattern:

```go
// Generate ring coordinates for R move at layer k
func ringR(N, k int) []Coord {
    var ring []Coord
    // Up face: column N-1-k, rows 0 to N-1
    for r := 0; r < N; r++ {
        ring = append(ring, Coord{Up, r, N-1-k})
    }
    // Back face: column k, rows N-1 to 0 (reversed)
    for r := N-1; r >= 0; r-- {
        ring = append(ring, Coord{Back, r, k})
    }
    // Down face: column N-1-k, rows N-1 to 0
    for r := N-1; r >= 0; r-- {
        ring = append(ring, Coord{Down, r, N-1-k})
    }
    // Front face: column N-1-k, rows 0 to N-1
    for r := 0; r < N; r++ {
        ring = append(ring, Coord{Front, r, N-1-k})
    }
    return ring
}

// Similar generators for L, U, D, F, B moves
func ringL(N, k int) []Coord { /* ... */ }
func ringU(N, k int) []Coord { /* ... */ }
func ringD(N, k int) []Coord { /* ... */ }
func ringF(N, k int) []Coord { /* ... */ }
func ringB(N, k int) []Coord { /* ... */ }
```

#### 2.2 Slice Ring Generators
For M, E, S moves:
```go
func ringM(N, k int) []Coord { /* Similar to L but for middle layers */ }
func ringE(N, k int) []Coord { /* Similar to D but for equatorial layers */ }
func ringS(N, k int) []Coord { /* Similar to F but for standing layers */ }
```

#### 2.3 Face Rotation Generator
When turning a face, the stickers on that face rotate:
```go
func faceRotationRing(face Face, N, layer int) [][]Coord {
    // Returns concentric rings from outer edge to center
    // Each ring rotates independently
}
```

### Phase 3: Permutation Generation

#### 3.1 Permutation Builder
```go
type Permutation []int // Maps source index to destination index

func generatePermutation(N int, moveType MoveType, layer int, quarterTurns int) Permutation {
    perm := make(Permutation, 6*N*N)
    // Initialize identity permutation
    for i := range perm {
        perm[i] = i
    }
    
    // Get ring coordinates based on move type
    var ring []Coord
    switch moveType {
    case MoveR: ring = ringR(N, layer)
    case MoveL: ring = ringL(N, layer)
    // ... etc
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
        // Add face rotation permutation
    }
    
    return perm
}

func rotateSlice(slice []int, quarterTurns int) []int {
    n := len(slice)
    shift := (quarterTurns * n / 4) % n
    result := make([]int, n)
    for i := range slice {
        result[i] = slice[(i+shift)%n]
    }
    return result
}
```

### Phase 4: Permutation Cache

#### 4.1 Cache Key
```go
type PermKey struct {
    N            int
    MoveType     MoveType
    Layer        int
    QuarterTurns int
}
```

#### 4.2 Cache Implementation
```go
var permCache = make(map[PermKey]Permutation)
var permCacheMu sync.RWMutex

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
```

### Phase 5: Apply Permutation

#### 5.1 Simple Copy Implementation
```go
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
```

#### 5.2 In-Place Cycle Implementation (Advanced)
```go
func applyPermutationInPlace(cube *Cube, perm Permutation) {
    // Track visited indices
    visited := make([]bool, len(perm))
    
    for start := range perm {
        if visited[start] || perm[start] == start {
            continue
        }
        
        // Follow the cycle
        current := start
        startColor := getColor(cube, current)
        
        for {
            next := perm[current]
            if next == start {
                setColor(cube, current, startColor)
                break
            }
            setColor(cube, current, getColor(cube, next))
            visited[current] = true
            current = next
        }
    }
}
```

### Phase 6: Move Execution

#### 6.1 Refactored ApplyMove
```go
func (c *Cube) ApplyMove(move Move) error {
    // Determine layers affected
    layers := getAffectedLayers(move, c.Size)
    
    for _, layer := range layers {
        // Get permutation from cache
        perm := getPermutation(c.Size, move.Type, layer, move.Prime)
        
        // Apply permutation
        applyPermutation(c, perm)
    }
    
    return nil
}

func getAffectedLayers(move Move, N int) []int {
    switch move.Type {
    case MoveR, MoveL, MoveU, MoveD, MoveF, MoveB:
        if move.Wide {
            return []int{0, 1} // Outer two layers for wide moves
        }
        return []int{0} // Just outer layer
    case MoveM:
        return []int{N/2} // Middle layer(s)
    case MoveE:
        return []int{N/2} // Equatorial layer(s)
    case MoveS:
        return []int{N/2} // Standing layer(s)
    case Movex, Movey, Movez:
        // Cube rotations affect all layers
        layers := make([]int, N)
        for i := range layers {
            layers[i] = i
        }
        return layers
    default:
        return []int{move.Layer}
    }
}
```

## Testing Strategy

### Integration with Existing Fuzzing Framework

The cube repository already has a **comprehensive fuzzing framework** that has successfully identified critical bugs in the current move system. We will leverage and extend this existing framework rather than creating a new one.

#### Existing Fuzzing Components:

1. **Unit-Level Fuzzing** (`internal/cube/moves_test.go`):
   - `TestMoveInverses`: 50 random scrambles with deterministic seed
   - `TestCircuitFuzzing`: Validates finite cycle order property
   - `TestMoveInversesAggressiveFuzzing`: 1000 tests with longer sequences
   - Uses `rand.New(rand.NewSource(42))` for reproducibility

2. **E2E Shell Fuzzing** (`test/e2e_test.sh`):
   - `fuzz_test_move_inverses`: Tests scramble + inverse = solved
   - Currently **blocks solver fuzzing** due to discovered move bugs
   - Uses `RANDOM=42` for deterministic bash testing

#### Why Keep the Existing Framework:

1. **It Works**: Has already found fundamental bugs in the move system
2. **Comprehensive**: Tests from unit level to full CLI integration
3. **Deterministic**: Reproducible failures for debugging
4. **Property-Based**: Tests mathematical invariants (inverses, cycles)
5. **Performance-Aware**: Balanced test counts avoid excessive runtime

### Testing Plan for Refactored System

#### Phase 1: Maintain Existing Tests
```go
// Keep all existing tests in moves_test.go
// They should pass with the new implementation
func TestMoveInverses(t *testing.T) {
    // Existing 50-test fuzzing with same seed
}

func TestCircuitFuzzing(t *testing.T) {
    // Existing cycle order validation
}

func TestMoveInversesAggressiveFuzzing(t *testing.T) {
    // Existing 1000-test suite
}
```

#### Phase 2: Add Permutation-Specific Tests
```go
// New tests for permutation system
func TestPermutationGeneration(t *testing.T) {
    // Verify permutations are valid (bijective)
    // Test all move types generate correct permutations
}

func TestPermutationCache(t *testing.T) {
    // Verify cache returns consistent results
    // Test concurrent access safety
}

func TestPermutationProperties(t *testing.T) {
    // Test M · M' = Identity permutation
    // Test M⁴ = Identity for all moves
    // Test permutation composition
}
```

#### Phase 3: Extend Fuzzing Coverage
```go
// Add support for advanced moves to existing fuzzer
func TestAdvancedMovesFuzzing(t *testing.T) {
    rng := rand.New(rand.NewSource(42))
    
    // Include M, E, S, Rw, Fw, x, y, z moves
    moves := []string{
        "R", "R'", "R2", "L", "L'", "L2", 
        "U", "U'", "U2", "D", "D'", "D2",
        "F", "F'", "F2", "B", "B'", "B2",
        "M", "M'", "M2", "E", "E'", "E2",
        "S", "S'", "S2", "Rw", "Rw'", "Rw2",
        "x", "x'", "x2", "y", "y'", "y2"
    }
    
    // Run same inverse test pattern
    for i := 0; i < 100; i++ {
        // Generate scramble with advanced moves
        // Test scramble + inverse = solved
    }
}
```

#### Phase 4: Re-enable Solver Fuzzing
Once the refactored move system passes all tests:
```bash
# Uncomment solver fuzzing in e2e_test.sh
# Lines 467-489 can be re-enabled
```

### Success Criteria

1. **All existing tests pass**: No regression from current functionality
2. **E2E fuzzing passes**: 100% success on `fuzz_test_move_inverses`
3. **Extended move support**: Advanced notation fuzzes correctly
4. **Solver fuzzing enabled**: Can uncomment and run solver tests
5. **Performance maintained**: Similar or better than current system

### Debugging Support

Keep the existing debugging features:
- Detailed failure output with cube state
- Reproducible test sequences with fixed seeds
- Progress reporting during long test runs
- First-failure analysis with verbose output

## Migration Path

1. Implement new system alongside old
2. Add feature flag to switch implementations
3. Run both in parallel, compare results
4. Gradually migrate tests to new system
5. Remove old implementation

## Performance Considerations

- Permutation generation: O(N²) but cached
- Move application: O(N²) per move
- Memory: O(N²) per cached permutation
- Cache size: ~1000 entries for typical usage

## Future Extensions

1. **Optimal move sequences**: Permutations compose naturally
2. **Pattern databases**: Store permutations as compressed integers
3. **GPU acceleration**: Permutations are embarrassingly parallel
4. **Group theory analysis**: Permutation representation enables mathematical analysis