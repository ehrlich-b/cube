# TODO.md - Rubik's Cube Solver Development Plan

## 🚨 **CRITICAL RESET: Moves Refactor Before Everything Else (January 2025)**

**NEW REALITY**: The current moves.go implementation is a branch-heavy, hard-coded nightmare that makes everything else harder. We MUST refactor to a data-driven permutation system BEFORE touching any solver code.

---

## 🔥 **PHASE 0: MOVES.GO REFACTOR (MANDATORY FIRST STEP)**

### **The Problem**
- **500+ lines of nested if/switch statements** with hand-coded `size-1-i` geometry
- **No reuse across cube sizes** - copy-pasted logic for each dimension
- **Impossible to test properly** - one-off errors hide in the branch jungle
- **Blocks all progress** - can't build reliable solvers on unreliable moves

### **The Solution: Permutation-Based Move System**

**Core Concepts:**
1. **Flatten stickers to indices**: `(face,row,col) → index = face*N² + row*N + col`
2. **Moves as permutations**: Each move is just a mapping of indices
3. **Data-driven generation**: Compute permutations once, cache forever
4. **Zero branches**: Apply moves by copying through permutation tables

### 🎯 **Refactor Implementation Tasks**

**Infrastructure:**
- [ ] Create index mapping functions: `stickerIndex()`, `indexToCoord()`
- [ ] Define `Coord` struct for (face, row, col) tuples
- [ ] Create `Permutation` type as `[]int` mapping

**Ring Generators:**
- [ ] Implement `ringR(N, k)` - generates coordinates for R move at layer k
- [ ] Implement `ringL(N, k)` - generates coordinates for L move at layer k
- [ ] Implement `ringU(N, k)` - generates coordinates for U move at layer k
- [ ] Implement `ringD(N, k)` - generates coordinates for D move at layer k
- [ ] Implement `ringF(N, k)` - generates coordinates for F move at layer k
- [ ] Implement `ringB(N, k)` - generates coordinates for B move at layer k
- [ ] Implement slice ring generators: `ringM()`, `ringE()`, `ringS()`
- [ ] Implement face rotation generator for rotating stickers on turned face

**Permutation System:**
- [ ] Implement `generatePermutation(N, moveType, layer, quarterTurns)`
- [ ] Create permutation cache with `PermKey` struct
- [ ] Implement thread-safe cache with `sync.RWMutex`
- [ ] Add `getPermutation()` with cache lookup

**Application:**
- [ ] Implement `applyPermutation()` with simple copy approach
- [ ] Implement `applyPermutationInPlace()` with cycle detection (advanced)
- [ ] Refactor `ApplyMove()` to use permutation system
- [ ] Implement `getAffectedLayers()` for wide/slice moves

**Testing:**
- [ ] Unit test each ring generator
- [ ] Test permutation generation for all move types
- [ ] Test permutation caching behavior
- [ ] Implement fuzz test: scramble + inverse = solved
- [ ] Property tests: M·M' = I, M⁴ = I
- [ ] Benchmark old vs new implementation

**Migration:**
- [ ] Add feature flag to switch implementations
- [ ] Run both in parallel, compare results
- [ ] Gradually migrate existing tests
- [ ] Remove old implementation once verified

**Success Criteria:**
- [ ] All existing tests pass with new implementation
- [ ] Fuzz test passes 10,000+ random scrambles
- [ ] Performance equal or better than old system
- [ ] Code reduced from 500+ lines to <200 lines
- [ ] Zero hand-coded geometry or branches

---

## 🎯 **THEN: PHASE 1 - CFEN Foundation**

Only after moves.go is refactored properly can we proceed with CFEN implementation...

[Rest of original TODO.md content follows below, but DO NOT START until moves refactor is complete]

---

## 🎯 **CRITICAL DISCOVERY: We Need CFEN First (January 2025)**

**BREAKTHROUGH**: We discovered CFEN 1.0 (Cube Forsyth-Edwards Notation) - a standardized format for representing cube states with wildcard support. This is the missing foundation piece we need before building any solvers.

**Why CFEN Changes Everything:**
- ✅ **FEN-equivalent for cubes**: Single-line, diff-friendly, unambiguous format
- ✅ **Orientation-aware**: Explicit Up/Front specification (`WG|...`)
- ✅ **Wildcard support**: `?` for "don't care" positions (perfect for layer-by-layer verification)
- ✅ **NxN scalable**: Works for 2x2 through 17x17+ cubes automatically
- ✅ **Partial states**: Can express "white cross solved, rest unknown"
- ✅ **Human-readable**: `WG|?W?WWW?W?/?9/?9/?9/?9/?9` (white cross only)
- ✅ **Compact**: Run-length encoding (`W9` = 9 white stickers)

**Examples:**
```
WG|WWWWWWWWW/RRRRRRRRR/GGGGGGGGG/YYYYYYYYY/OOOOOOOOO/BBBBBBBBB  # Solved 3x3
WG|?W?WWW?W?/?9/?9/?9/?9/?9                                      # White cross only
WG|W16/R16/G16/Y16/O16/B16                                       # Solved 4x4  
WG|Y25/?25/?25/?25/?25/?25                                       # 5x5 OLL drill
```

**CFEN Implementation Plan:**

### 🎯 **PHASE 1: CFEN Foundation (After Moves Refactor)**

**1. CFEN Data Structures**
- [ ] Create `CFENState` struct with orientation + 6 faces
- [ ] Create `CFENOrientation` struct (Up color, Front color)
- [ ] Create `CFENFace` struct with sticker array and wildcard support

**2. CFEN Parser (Core)**
- [ ] Implement orientation field parser: `WG` → Up=White, Front=Green
- [ ] Implement run-length decoder: `W9` → `WWWWWWWWW`, `?5` → `?????`
- [ ] Add face parsing with validation (6 faces, equal counts, perfect square)
- [ ] Add cube dimension detection: sticker count → N×N verification

**3. CFEN ↔ Cube Conversion**
- [ ] Implement CFEN → internal `Cube` conversion (map positions to `Faces[6][][]Color`)
- [ ] Implement `Cube` → CFEN generation (convert back to CFEN string)
- [ ] Handle wildcard (`?`) to `Grey` color mapping

**4. CFEN Wildcard Matching System**
- [ ] Implement partial cube state verification (ignore `?` positions during comparison)
- [ ] Add `cube verify-cfen` command: verify solution reaches target CFEN state
- [ ] Support layer-by-layer verification: white cross, white face, middle layer, etc.

**5. CFEN CLI Integration**
- [ ] Add `cube parse-cfen <cfen-string>` command (parse and display cube state)
- [ ] Add `cube generate-cfen <scramble>` command (apply moves and output CFEN)
- [ ] Add `cube verify-cfen <scramble> <solution> --target <cfen>` command
- [ ] Add `cube match-cfen <current-cfen> <target-cfen>` command (show differences)
- [ ] Add `--cfen` output flag to ALL commands (twist, solve, show, etc.) for CFEN output instead of visual
- [ ] Add `--start <cfen-string>` input flag to ALL commands to specify starting cube state
- [ ] Add CFEN dimension validation: if `--dimension` provided, verify CFEN matches that size
- [ ] Add CFEN auto-dimension detection: if no `--dimension` provided, infer from CFEN string
- [ ] Update existing commands to support CFEN input: `cube solve --start "WG|?W?WWW?W?/?9/?9/?9/?9/?9" "R U R'"`
- [ ] Update existing commands to support CFEN output: `cube twist "R U R' U'" --cfen` → CFEN string

**6. CFEN Test Suite**
- [ ] Test 3x3, 4x4, 5x5+ cube parsing and generation
- [ ] Test wildcard matching and partial state verification
- [ ] Test run-length encoding/decoding edge cases
- [ ] Test invalid CFEN strings (wrong face counts, non-square, bad tokens)
- [ ] Test orientation field validation and conversion
- [ ] Test `--cfen` output flag on all commands (twist, solve, show, verify, lookup)
- [ ] Test `--start <cfen>` input flag on all commands with various cube states
- [ ] Test CFEN dimension validation and auto-detection
- [ ] Test CFEN input/output integration with existing move parsing and application
- [ ] Test complex workflows: `cube solve --start <partial-cfen> --cfen` (CFEN in, CFEN out)

---

## 🏗️ **PHASE 2: Real Solvers Using CFEN**

**Goal**: Build the world's best command-line cube solver with REAL algorithms that actually solve cubes.

### 🎯 **PRIORITY 1: Fix BeginnerSolver Using CFEN (Layer-by-Layer Method)**

**Real Beginner Method Steps (Now With CFEN Verification):**
1. **White Cross** - Get white edges in correct positions
   - Target CFEN: `WG|?W?WWW?W?/?9/?9/?9/?9/?9` (only cross positions matter)
2. **White Corners** - Complete white face with corners  
   - Target CFEN: `WG|W9/?9/?9/?9/?9/?9` (entire white face solved)
3. **Middle Layer** - Insert edge pieces into middle layer
   - Target CFEN: `WG|W9/?W?W?W?W?/?9/?9/?9/?9` (white + middle edges)
4. **Yellow Cross** - Create cross on top (various cases)
   - Target CFEN: `WG|W9/?W?W?W?W?/?Y?YYY?Y?/?9/?9` (+ yellow cross)
5. **Yellow Corners** - Orient all yellow corners
   - Target CFEN: `WG|W9/?W?W?W?W?/Y9/?9/?9` (entire yellow face)
6. **Corner Permutation** - Position yellow corners correctly
7. **Edge Permutation** - Position final edges correctly
   - Target CFEN: `WG|W9/R9/G9/Y9/O9/B9` (completely solved)

[Rest of original TODO.md solver content...]

---

## 📊 **CURRENT IMPLEMENTATION STATUS**

### ✅ **Infrastructure (World-Class)**
- **Testing Framework**: Comprehensive fuzzing with 76 tests
- **Cube Engine**: Perfect NxN representation (but moves need refactor)
- **Visualization**: Beautiful terminal and web output
- **Algorithm Database**: 61 real speedcubing algorithms
- **User Interface**: CLI + web terminal with headless mode

### 🔥 **Critical Issues**
- **Moves Implementation**: Branch-heavy nightmare blocking everything
- **Solvers**: All completely broken placeholders
- **CFEN**: Not implemented yet

### 🎯 **The Path Forward**
1. **FIRST**: Refactor moves.go to permutation-based system
2. **THEN**: Implement CFEN for verification
3. **FINALLY**: Build real solvers with CFEN verification

**Next Action**: Start moves.go refactor immediately. See `/docs/moves-refactor-design.md` for detailed design.

---

## 💪 **MOTIVATION: Why We're Building This**

We're not building another toy cube solver. We're building the **definitive command-line cube tool** that:

- **Actually solves any cube** - No more placeholder algorithms
- **Has clean, maintainable code** - Data-driven, not branch-heavy
- **Teaches proper methods** - Real beginner and CFOP techniques  
- **Handles any size** - 2x2 through 10x10+ flawlessly
- **Has perfect reliability** - Fuzzing ensures it never fails
- **Provides amazing UX** - Beautiful output, web interface, comprehensive docs

**This will be the cube solver that becomes the standard.** Nothing less than perfection.