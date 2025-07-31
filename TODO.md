# TODO.md - Rubik's Cube Solver Development Plan

## 🚨 **REALITY CHECK: FUZZING EXPOSED THE TRUTH (January 2025)**

**CRITICAL DISCOVERY**: Our comprehensive fuzzing test suite has revealed that the "working solvers" we thought we built are sophisticated placeholders that fail on real scrambles. This TODO.md was living in a fantasy world. Time to build the BEST cube solver tool in the world properly.

---

## 🎯 **THE TRUTH: What Actually Works vs. What's Broken**

### ✅ **SOLID FOUNDATION - These Work Perfectly**
- **Cube Representation**: Full NxN support (2x2 through 20x20+) with proper multi-layer moves
- **Move System**: WCA standard notation (M/E/S slices, Rw/Fw wide moves, 2R/3L layers, x/y/z rotations)
- **Visualization**: Beautiful ASCII/Unicode output with color support
- **Web Interface**: Full terminal emulator at `/terminal` with REST API
- **Algorithm Database**: 61 real speedcubing algorithms with lookup system
- **Pattern Highlighting**: Cross, OLL, PLL, F2L detection and visualization
- **Testing Infrastructure**: Comprehensive 76-test suite including fuzzing
- **Headless Mode**: Programmatic output for `solve` and `verify` commands
- **All Infrastructure Commands**: twist, show, lookup, optimize, find all work perfectly

### 🔥 **BROKEN CORE - The Solvers Are Placeholders**

**BeginnerSolver**: 
- ❌ Only handles 5 hardcoded inverse patterns (`R` → `R'`, etc.)
- ❌ For complex scrambles, randomly applies T-perm/U-perm up to 15 times hoping something works
- ❌ Produces 117-move "solutions" that don't actually solve the cube
- ❌ **FUZZ RESULT**: Fails on first complex scramble tested

**CFOPSolver**:
- ❌ **FUZZ RESULT**: 0/25 random scrambles solved (100% failure rate)
- ❌ Likely just calls BeginnerSolver or returns empty moves

**KociembaSolver**:
- ❌ **FUZZ RESULT**: 0/25 random scrambles solved (100% failure rate)  
- ❌ Likely just calls BeginnerSolver or returns empty moves

**Multi-dimensional Solving**:
- ❌ **FUZZ RESULT**: 2x2 cubes: 0/10 solved, 4x4 cubes: 0/5 solved
- ❌ Uses same broken algorithms on larger cubes

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

### 🎯 **PHASE 1: CFEN Foundation (CRITICAL - Do This First)**

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

**🎯 CFEN MAKES THIS TRIVIAL:**

Instead of complex "grey square system" with cubie addressing, we use CFEN:

```bash
# Verify each step using CFEN targets
cube verify-cfen "R U R' U'" "solution" --target "WG|?W?WWW?W?/?9/?9/?9/?9/?9"  # White cross
cube verify-cfen "R U R' U'" "solution" --target "WG|W9/?9/?9/?9/?9/?9"        # White face  
cube verify-cfen "R U R' U'" "solution" --target "WG|W9/?W?W?W?W?/?9/?9/?9/?9" # Middle layer
```

**No more complex cubie addressing needed!** CFEN wildcards (`?`) handle "don't care" positions automatically.

**BeginnerSolver Implementation (CFEN-Based):**
- [ ] **White Cross Algorithm** - Find and position white edges, verify with `WG|?W?WWW?W?/?9/?9/?9/?9/?9`
- [ ] **White Corner Insertion** - Use right-hand/left-hand algorithms, verify with `WG|W9/?9/?9/?9/?9/?9`
- [ ] **Middle Layer Edges** - F2L-style edge insertion, verify with `WG|W9/?W?W?W?W?/?9/?9/?9/?9`
- [ ] **Yellow Cross Formation** - F R U R' U' F' and variations, verify with target CFEN
- [ ] **Corner Orientation** - Sune/Anti-Sune algorithms, verify with `WG|W9/?W?W?W?W?/Y9/?9/?9`
- [ ] **Corner Permutation** - A-perms and T-perms for corner positioning
- [ ] **Edge Permutation** - U-perms for final edge positioning, verify solved with `WG|W9/R9/G9/Y9/O9/B9`

**Success Criteria:**
- [ ] Solves ANY valid 3x3 scramble in under 80 moves
- [ ] Passes 100% of fuzzing tests (25+ random scrambles)  
- [ ] Each step verifies correct using CFEN pattern matching
- [ ] No random algorithm application - every move has purpose

### 🎯 **PRIORITY 2: Fix CFOPSolver (Advanced Method)**

**Real CFOP Method Steps:**
1. **Cross** - Create cross on bottom in under 8 moves
2. **F2L** - First Two Layers using our 61-algorithm database
3. **OLL** - Orient Last Layer using our OLL algorithms (22 implemented)  
4. **PLL** - Permute Last Layer using our PLL algorithms (21 implemented - complete set!)

**Implementation Requirements:**
- [ ] **Cross Solver** - Efficient cross construction algorithms
- [ ] **F2L Integration** - Use existing F2L algorithms from database
- [ ] **OLL Recognition** - Pattern matching for our 22 OLL cases
- [ ] **PLL Recognition** - Pattern matching for our 21 PLL cases (complete!)
- [ ] **Look-ahead** - Basic optimization to reduce move count

**Success Criteria:**
- [ ] Solves ANY valid 3x3 scramble in under 60 moves  
- [ ] Passes 100% of fuzzing tests
- [ ] Uses actual CFOP methodology, not random algorithms
- [ ] Leverages our comprehensive algorithm database

### 🎯 **PRIORITY 3: Fix Multi-Dimensional Support**

**2x2x2 Cubes:**
- [ ] **Simplified Beginner** - Only corners, no edge pieces
- [ ] **R U R' F R F'** method implementation
- [ ] **CLL (Corners of Last Layer)** algorithms

**4x4x4+ Cubes:**
- [ ] **Reduction Method** - Solve centers, pair edges, then solve as 3x3
- [ ] **Center Solving** - Systematic center piece algorithms
- [ ] **Edge Pairing** - Pair up double edges
- [ ] **Parity Handling** - 4x4 parity algorithms when needed

**Success Criteria:**
- [ ] 2x2: Solves ANY scramble in under 20 moves
- [ ] 4x4: Solves ANY scramble in under 120 moves  
- [ ] Passes multi-dimensional fuzzing tests

### 🎯 **PRIORITY 4: Enhanced Verification Using CFEN**

Replace the current broken verification system with CFEN-powered verification.

**CFEN Verification System:**
- [ ] **CFEN-based validation** - Apply scramble + solution, generate CFEN, compare with target
- [ ] **Wildcard matching** - Ignore `?` positions during CFEN comparison
- [ ] **Edge case testing** - Verify incorrect solutions properly fail CFEN matching
- [ ] **Performance** - Fast CFEN parsing and comparison for fuzzing tests
- [ ] **Multi-dimensional** - CFEN works for all cube sizes automatically

---

## 🧪 **THE NEW TESTING PHILOSOPHY**

### **CFEN-Powered Fuzzing Development**
- [ ] **Build solver incrementally** - Each step must pass CFEN verification
- [ ] **Test on real scrambles** - No more "simple case only" testing
- [ ] **CFEN self-verification** - Every solution must match target CFEN state
- [ ] **Debug on failure** - When fuzzing finds CFEN mismatch, fix immediately
- [ ] **Layer-by-layer fuzzing** - Test each solving step with CFEN wildcard verification
- [ ] **CFEN validation** - Fuzz test CFEN parsing, generation, and wildcard matching

### **Comprehensive Coverage**
- [ ] **All algorithms fuzzed** - BeginnerSolver, CFOPSolver, KociembaSolver
- [ ] **All dimensions fuzzed** - 2x2, 3x3, 4x4, 5x5+ 
- [ ] **Edge cases fuzzed** - Empty scrambles, single moves, long scrambles
- [ ] **Performance testing** - Solutions under move count limits

### **No More Placeholder Acceptance**
- [ ] **Every algorithm must actually work** - No "gets lucky sometimes"
- [ ] **Every test must test real functionality** - No testing placeholder happy paths  
- [ ] **Documentation must reflect reality** - No claiming things work when they don't

---

## 🏆 **SUCCESS METRICS: Best Cube Tool in the World**

### **Solving Performance**
- [ ] **BeginnerSolver**: 100% success rate, <80 moves average
- [ ] **CFOPSolver**: 100% success rate, <60 moves average  
- [ ] **2x2 Solver**: 100% success rate, <20 moves average
- [ ] **4x4 Solver**: 100% success rate, <120 moves average

### **Reliability** 
- [ ] **10,000 fuzz tests pass** - Extensive random scramble testing
- [ ] **All WCA scrambles solvable** - Handle competition-style scrambles
- [ ] **Zero false positives** - verify() never accepts wrong solutions
- [ ] **Sub-second solving** - Even complex scrambles solve quickly

### **User Experience** 
- [ ] **Every documented example works** - No broken examples in guides
- [ ] **Error messages are helpful** - Clear feedback on what went wrong
- [ ] **Headless mode perfect** - Reliable programmatic interface
- [ ] **Web interface seamless** - All CLI functionality available online

---

## 🚀 **IMPLEMENTATION STRATEGY**

### **Phase 1: CFEN Foundation (DO THIS FIRST)**
1. **Implement CFEN parser and generator** - Core data structures and validation
2. **Add CFEN CLI commands** - parse-cfen, generate-cfen, verify-cfen 
3. **Test CFEN extensively** - All cube sizes, wildcards, edge cases
4. **Integrate CFEN with existing cube system** - Bidirectional conversion

### **Phase 2: CFEN-Powered BeginnerSolver**
1. **Rewrite BeginnerSolver using CFEN verification** - Each step has target CFEN
2. **Implement layer-by-layer with CFEN** - White cross, corners, middle, yellow
3. **Add CFEN fuzzing for each step** - Verify wildcard matching works
4. **Achieve 100% success rate** - Real algorithms, CFEN-verified

### **Phase 3: CFEN-Powered CFOPSolver**  
1. **Implement CFOP with CFEN targets** - Cross, F2L, OLL, PLL verification
2. **Integrate existing algorithm database** - Use our 61 algorithms with CFEN
3. **Add pattern recognition using CFEN** - Match current state to known patterns
4. **Optimize and fuzz test with CFEN** - 100% success rate

### **Phase 4: Multi-dimensional CFEN Support**
1. **Extend CFEN to 2x2, 4x4+ cubes** - Already scales, just add solvers
2. **Implement size-specific methods** - 2x2 corner-only, 4x4+ reduction
3. **CFEN parity handling** - 4x4 specific algorithms with CFEN targets
4. **Comprehensive multi-dimensional fuzzing** - All sizes pass

---

## 📊 **CURRENT IMPLEMENTATION STATUS**

### ✅ **Infrastructure (World-Class)**
- **Testing Framework**: Comprehensive fuzzing with 76 tests
- **Cube Engine**: Perfect NxN representation and moves
- **Visualization**: Beautiful terminal and web output
- **Algorithm Database**: 61 real speedcubing algorithms
- **User Interface**: CLI + web terminal with headless mode

### 🔥 **Solvers (Completely Broken)**
- **BeginnerSolver**: Sophisticated placeholder (fails on real scrambles)
- **CFOPSolver**: Placeholder (100% failure rate)
- **KociembaSolver**: Placeholder (100% failure rate)
- **Multi-dimensional**: All broken

### 🎯 **The Path to Greatness**
We have built world-class infrastructure and testing. Now we implement CFEN (the missing foundation) then world-class algorithms. CFEN will enable proper layer-by-layer verification, and the fuzzing system will guide us.

**Next Action**: Implement CFEN 1.0 parser and generator system (Phase 1).

---

## 💪 **MOTIVATION: Why We're Building This**

We're not building another toy cube solver. We're building the **definitive command-line cube tool** that:

- **Actually solves any cube** - No more placeholder algorithms
- **Teaches proper methods** - Real beginner and CFOP techniques  
- **Handles any size** - 2x2 through 10x10+ flawlessly
- **Has perfect reliability** - Fuzzing ensures it never fails
- **Provides amazing UX** - Beautiful output, web interface, comprehensive docs

**This will be the cube solver that becomes the standard.** Nothing less than perfection.