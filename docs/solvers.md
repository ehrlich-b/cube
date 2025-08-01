# Cube Solver Design Document

## Overview

This document outlines the design and implementation strategy for real cube solvers in our system, with a focus on supporting:
- Layer-by-layer solving with CFEN verification
- Wildcard/partial state handling (grey squares)
- Support for all cube sizes (2x2 through NxN)
- Multiple solving algorithms (Beginner, CFOP, Kociemba)

## The Core Challenges

### 1. Even-Numbered Cubes Have No Fixed Centers

For odd-numbered cubes (3x3, 5x5, 7x7), the center pieces are fixed and define the color scheme. However, even-numbered cubes (2x2, 4x4, 6x6) have no fixed centers:

```
3x3 cube:                    4x4 cube:
    Y Y Y                        ? ? ? ?
    Y Y Y  <- center is Y        ? ? ? ?  <- no fixed center!
    Y Y Y                        ? ? ? ?
                                 ? ? ? ?
```

**Implications:**
- Cannot determine "correct" color placement without additional context
- Multiple valid solved states exist
- Need heuristics or user input to determine intended color scheme

**Solution Approaches:**
1. **Reference corners**: Use corner pieces to infer color scheme (corners have 3 colors each)
2. **CFEN orientation**: Use the orientation field (e.g., `YB|...`) to define expected colors
3. **Solver hints**: Allow user to specify expected color positions

### 2. Pattern Matching with Wildcards

When CFEN contains grey squares (`?`), we need to match partial patterns:

```
Target: YB|?Y?YYY?Y?/?9/?9/?9/?9/?9  (white cross)

Current state might be:
    R Y B
    Y Y Y  <- How do we know if R and B are "correct" when target is ?
    G Y O
```

**Challenges:**
- Cannot use exact position matching
- Need to identify "movable" vs "fixed" pieces
- Must track which pieces contribute to the pattern

**Solution: Semantic Pattern Recognition**

Instead of position-based matching, use semantic patterns:
```
WhiteCrossPattern {
    Requirements: [
        EdgePiece{colors: [White, Red], position: TopEdge},
        EdgePiece{colors: [White, Blue], position: RightEdge},
        EdgePiece{colors: [White, Orange], position: BottomEdge},
        EdgePiece{colors: [White, Green], position: LeftEdge}
    ]
}
```

### 3. Larger Cube Complexities

#### Piece Types by Cube Size
- **2x2**: Only corners (8 pieces)
- **3x3**: Corners (8), Edges (12), Centers (6 fixed)
- **4x4**: Corners (8), Edges (24), Centers (24 movable)
- **5x5**: Corners (8), Edges (36), Centers (24 fixed + 24 movable)
- **NxN**: Pattern continues...

#### Parity Issues
4x4+ cubes can have parity errors impossible on 3x3:
- OLL parity: Single edge flipped
- PLL parity: Two edges swapped

### 4. Algorithm Complexity Growth

Layer-by-layer steps for different cube sizes:

```
3x3 Layer-by-Layer:
1. White cross (4 edges)
2. White corners (4 corners)
3. Middle layer (4 edges)
4. Yellow cross (4 edges)
5. Yellow corners (4 corners)
6. Final positioning

4x4 Layer-by-Layer:
1. Center building (6 centers, 4 pieces each)
2. Edge pairing (12 edge pairs, 2 pieces each)
3. Reduce to 3x3
4. Solve as 3x3 with parity algorithms

5x5 Layer-by-Layer:
1. Center building (6 centers, 9 pieces each)
2. Edge pairing (12 edge triplets, 3 pieces each)
3. Reduce to 3x3
4. Solve as 3x3
```

## Prior Art Research

### 1. Kociemba's Algorithm (Two-Phase)
- **Phase 1**: Reduce to <U,D,R2,L2,F2,B2> group (orientation)
- **Phase 2**: Solve within this group (permutation)
- **Limitation**: Only works for 3x3 cubes
- **Strength**: Near-optimal solutions (typically 20-26 moves)

### 2. Thistlethwaite's Algorithm (Four-Phase)
- Progressively restricts move groups
- Foundation for Kociemba's algorithm
- Also 3x3 only

### 3. IDA* with Pattern Databases
- Used in optimal solvers
- Precomputed heuristics for subproblems
- Memory intensive but very effective

### 4. Reduction Method (4x4+)
- Build centers first
- Pair edges to form composite edges
- Solve as 3x3 with parity handling
- Standard approach for big cubes

### 5. Commutator-Based Methods
- Use sequences like [A, B] = A B A' B'
- Powerful for specific piece cycles
- Used in blindfolded solving

## Proposed Implementation Architecture

### 1. Piece Identification System

```go
type Piece interface {
    GetColors() []Color
    GetType() PieceType  // Corner, Edge, Center
    GetPosition() Position
}

type PieceTracker struct {
    pieces map[PieceID]*TrackedPiece
    cube   *Cube
}

// Identify pieces by their color combinations
func (pt *PieceTracker) FindPiece(colors []Color) *TrackedPiece
func (pt *PieceTracker) GetPiecesAtLayer(layer int) []*TrackedPiece
```

### 2. Pattern Recognition Engine

```go
type Pattern interface {
    Name() string
    Matches(cube *Cube, wildcards bool) bool
    GetTargetCFEN() string
    GetSolvingAlgorithm() []Move
}

type PatternMatcher struct {
    patterns []Pattern
    
    // Match with wildcards support
    func Match(cube *Cube, targetCFEN string) []Pattern
}
```

### 3. Layer-by-Layer Solver Framework

```go
type SolverPhase interface {
    GetName() string
    GetTargetPattern() string  // CFEN pattern
    IsComplete(cube *Cube) bool
    GetSubgoals() []Subgoal
    Execute(cube *Cube) ([]Move, error)
}

type LayerByLayerSolver struct {
    phases []SolverPhase
    
    func Solve(cube *Cube) (*SolverResult, error)
}
```

### 4. Algorithm Database

```go
type AlgorithmDB struct {
    algorithms map[string]Algorithm
}

type Algorithm struct {
    Name     string
    Moves    []Move
    Pattern  string  // What it solves
    PreCond  string  // Required state
    PostCond string  // Result state
}
```

## Implementation Plan

### Phase 1: 3x3 Beginner Method
1. **Piece Tracking**
   - Implement piece identification by color
   - Track piece positions throughout solve
   
2. **White Cross**
   - Find white edges
   - Calculate optimal insertion order
   - Use semantic pattern matching
   
3. **F2L (First Two Layers)**
   - Pair corners with edges
   - Track solved slots
   
4. **Last Layer**
   - OLL: 57 cases with recognition
   - PLL: 21 cases with recognition

### Phase 2: 4x4+ Support
1. **Center Building**
   - Implement center piece tracking
   - Commutator-based center solving
   
2. **Edge Pairing**
   - Slice flip-flop algorithm
   - Track edge parity
   
3. **Parity Algorithms**
   - OLL parity: r2 B2 U2 l U2 r' U2 r U2 F2 r F2 l' B2 r2
   - PLL parity: r2 U2 r2 Uw2 r2 Uw2

### Phase 3: Advanced Algorithms
1. **CFOP Implementation**
   - Cross optimization
   - F2L pair recognition
   - Full OLL/PLL algorithms
   
2. **Kociemba Two-Phase**
   - Implement coordinate system
   - Pruning tables
   - Phase 1 & 2 search

## Handling Wildcards in CFEN

### Strategy 1: Piece-Centric Validation
Instead of validating positions, validate pieces:
```
Target: YB|?Y?YYY?Y?/?9/?9/?9/?9/?9

Validation:
1. Find all white edges
2. Check if they're in cross positions
3. Ignore non-white stickers in those positions
```

### Strategy 2: Progressive Refinement
```
Step 1: YB|?Y?YYY?Y?/?9/?9/?9/?9/?9  (white cross)
Step 2: YB|YYYYYYYYY/?9/?9/?9/?9/?9    (white face)
Step 3: YB|YYYYYYYYY/RRR???RRR/...     (F2L)
```

Each step refines the pattern, replacing wildcards with concrete requirements.

### Strategy 3: Constraint Satisfaction
Model as constraint satisfaction problem:
- Variables: Piece positions
- Constraints: CFEN patterns
- Solve using backtracking/propagation

## Testing Strategy

### 1. Unit Tests
- Piece identification
- Pattern matching with wildcards
- Individual algorithm correctness

### 2. Integration Tests
- Full solves from various scrambles
- CFEN verification at each step
- Parity handling for 4x4+

### 3. Property-Based Tests
- Scramble + Solve = Solved state
- No algorithm breaks solved pieces
- Commutator properties: [A,B] = ABA'B'

### 4. Performance Tests
- Solve times for various cube sizes
- Memory usage for pattern databases
- Algorithm efficiency metrics

## Open Questions

1. **Color Scheme Detection**: How do we determine intended color scheme for even cubes?
   - Option A: First valid scheme found
   - Option B: User preference
   - Option C: Most common scheme heuristic

2. **Optimal vs Practical**: Do we optimize for:
   - Fewest moves? (Kociemba)
   - Easiest to understand? (Layer-by-layer)
   - Fastest execution? (CFOP)

3. **Wildcard Semantics**: What does `?` mean exactly?
   - "Any color acceptable"
   - "Don't know current color"
   - "Will be determined by other constraints"

4. **Large Cube Limits**: Where do we draw the line?
   - Memory constraints for 10x10+
   - Algorithm complexity for 20x20+
   - Practical solving time limits

## Conclusion

Building a robust cube solver requires:
1. Strong piece identification and tracking
2. Flexible pattern matching with wildcard support
3. Cube-size-aware algorithm selection
4. Careful handling of edge cases (parity, color schemes)

The implementation should start with 3x3 layer-by-layer, establish the pattern matching framework, then expand to other sizes and methods.