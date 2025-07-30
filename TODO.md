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

## 🏗️ **THE REBUILD: Phase ∞ - Make Solvers Actually Work**

**Goal**: Build the world's best command-line cube solver with REAL algorithms that actually solve cubes.

### 🎯 **PRIORITY 1: Fix BeginnerSolver (Layer-by-Layer Method)**

**Real Beginner Method Steps:**
1. **White Cross** - Get white edges in correct positions
2. **White Corners** - Complete white face with corners
3. **Middle Layer** - Insert edge pieces into middle layer
4. **Yellow Cross** - Create cross on top (various cases)
5. **Yellow Corners** - Orient all yellow corners
6. **Corner Permutation** - Position yellow corners correctly
7. **Edge Permutation** - Position final edges correctly

**🔧 CRITICAL INFRASTRUCTURE REQUIREMENT: Grey Square System**

Layer-by-layer solving requires the ability to **ignore already-solved cubies** during pattern matching and verification. We need:

**Cubie Addressing System:**
```
Reading cube faces like a book (left-to-right, top-to-bottom):
Face positions: 1,2,3 / 4,5,6 / 7,8,9
  1 2 3
  4 5 6  <- 5 is center (doesn't move)
  7 8 9

3D cube numbering:
- Up face: 1-9
- Left face: 10-18  
- Front face: 19-27 (where 23 = front center)
- Right face: 28-36
- Back face: 37-45  
- Down face: 46-54
```

**Grey Square CLI Interface:**
```bash
# Verify white cross (only care about white edges + centers)
cube verify-layer "scramble" "solution" --care "2,4,6,8,23,14,32,41,50"

# Verify white corners (white face complete, ignore rest)  
cube verify-layer "scramble" "solution" --care "1-9" --dontcare "10-54"

# Check middle layer edges (ignore top/bottom layers)
cube verify-layer "scramble" "solution" --care "12,16,21,25,30,34,39,43" --dontcare "TL,BL"

# Using layer aliases for convenience
cube verify-layer "scramble" "solution" --care "WC" --dontcare "YL,ML"  # White cross, ignore yellow+middle
cube verify-layer "scramble" "solution" --dontcare "TE,BE"  # Don't care about top/bottom edges
```

**Layer and Piece Aliases:**
```bash
# Layer aliases
TL = "1-9"     # Top Layer (all 9 positions)
ML = "12,16,21,25,30,34,39,43"  # Middle Layer (edge positions)
BL = "46-54"   # Bottom Layer (all 9 positions)

# Piece type aliases  
TC = "1,3,7,9"     # Top Corners
TE = "2,4,6,8"     # Top Edges
MC = "11,13,17,19,29,31,35,37"  # Middle Corners (middle layer corners)
ME = "12,16,21,25,30,34,39,43"  # Middle Edges
BC = "46,48,52,54" # Bottom Corners
BE = "47,49,51,53" # Bottom Edges

# Face aliases
UF = "1-9"    # Up Face
LF = "10-18"  # Left Face  
FF = "19-27"  # Front Face
RF = "28-36"  # Right Face
BF = "37-45"  # Back Face
DF = "46-54"  # Down Face

# Common combinations
WC = "2,4,6,8,50"     # White Cross (white edges + white center)
WF = "1-9"            # White Face (complete white layer)
YC = "47,49,51,53,5"  # Yellow Cross (yellow edges + yellow center)
YF = "46-54"          # Yellow Face (complete yellow layer)
```

**Pattern Matching with Grey:**
- [ ] **Grey-aware pattern recognition** - Match patterns ignoring specified cubies
- [ ] **Partial cube state verification** - Check only relevant cubies for each layer
- [ ] **CLI cubie specification** - `--care` (whitelist) and `--dontcare` (blacklist) flags
- [ ] **Layer aliases** - TL, ML, BL, TC, TE, etc. for user convenience
- [ ] **3D cubie addressing** - Systematic numbering for all cube positions

**Implementation Requirements:**
- [ ] **Cubie numbering system** - Map 3D positions to linear addresses (1-54 for 3x3)
- [ ] **CLI cubie specification** - `--care "1,2,3"` and `--dontcare "4,5,6"` parsing
- [ ] **Alias expansion system** - Convert "TL,ML,BC" to actual cubie numbers
- [ ] **Range parsing** - Handle "1-9" and "46-54" syntax
- [ ] **Pattern matching engine** - Ignore grey squares during comparisons (internal grey logic)
- [ ] **Layer verification** - Check partial cube states during solving
- [ ] **White Cross Algorithm** - Systematic edge positioning with grey matching
- [ ] **White Corner Insertion** - Right-hand and left-hand algorithms with partial verification  
- [ ] **Middle Layer Edges** - F2L-style insertion with `--dontcare "TL,BL"`
- [ ] **Top Cross Formation** - F R U R' U' F' variations with `--dontcare "ML,BL"`
- [ ] **Corner Orientation** - Sune/Anti-Sune with `--dontcare "TE,ME,BE"`
- [ ] **Corner Permutation** - A-perms with `--dontcare "TE,ME,BE"`  
- [ ] **Edge Permutation** - U-perms with `--dontcare "TC,MC,BC"`

**Success Criteria:**
- [ ] **3D thinking in code** - Proper cubie addressing and grey square logic
- [ ] **Layer-by-layer verification** - Each step verifies only relevant cubies
- [ ] Solves ANY valid 3x3 scramble in under 80 moves
- [ ] Passes 100% of fuzzing tests (25+ random scrambles)
- [ ] Each step reduces cube to known state with proper grey square masking
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

### 🎯 **PRIORITY 4: Fix Verification System**

Currently verify() might accept incorrect solutions by accident, and it lacks grey square support for layer-by-layer verification.

**Core Verification Fixes:**
- [ ] **Strict Validation** - Actually apply scramble + solution and check solved state
- [ ] **Edge Case Testing** - Verify incorrect solutions properly fail
- [ ] **Performance** - Fast verification for fuzzing tests
- [ ] **Multi-dimensional** - Proper verification for all cube sizes

**Grey Square Verification System:**
- [ ] **verify-layer command** - New CLI command for partial cube verification
- [ ] **Cubie selection parsing** - Parse `--care "WC"` and `--dontcare "TL,ML"` syntax  
- [ ] **Alias expansion** - Convert layer aliases to cubie numbers
- [ ] **Partial state comparison** - Compare only specified cubies, ignore grey ones
- [ ] **Layer-by-layer testing** - Verify each step of beginner method independently
- [ ] **3D position mapping** - Convert cubie numbers to actual cube positions
- [ ] **Pattern matching integration** - Use grey squares in algorithm recognition

---

## 🧪 **THE NEW TESTING PHILOSOPHY**

### **Fuzzing-First Development**
- [ ] **Build solver incrementally** - Each step must pass fuzz tests
- [ ] **Test on real scrambles** - No more "simple case only" testing
- [ ] **Self-verification** - Every solution must verify correctly
- [ ] **Debug on failure** - When fuzzing finds a bug, fix it immediately
- [ ] **Layer-by-layer fuzzing** - Test each solving step with grey square verification
- [ ] **Grey square validation** - Fuzz test partial cube state matching

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

### **Phase 1: Fix BeginnerSolver**
1. **Study real layer-by-layer method** - Research proper algorithms
2. **Implement white cross solving** - First step of beginner method
3. **Add fuzzing for each step** - Verify each layer works before moving on
4. **Build complete beginner method** - All 7 steps implemented
5. **Achieve 100% fuzz test success** - No failures on random scrambles

### **Phase 2: Fix CFOPSolver** 
1. **Implement cross solving** - Efficient bottom cross construction
2. **Integrate F2L algorithms** - Use our existing 10 F2L cases
3. **Implement OLL recognition** - Pattern match our 22 OLL algorithms  
4. **Implement PLL recognition** - Pattern match our 21 PLL algorithms
5. **Optimize and fuzz test** - Achieve 100% success rate

### **Phase 3: Fix Multi-dimensional**
1. **Implement 2x2 solver** - Simplified method for corner-only cube
2. **Implement 4x4+ reduction** - Centers, edge pairing, 3x3 solve
3. **Add parity handling** - 4x4 specific algorithms
4. **Comprehensive fuzz testing** - All dimensions must pass

### **Phase 4: Polish and Optimize**
1. **Move count optimization** - Reduce average solution length
2. **Performance optimization** - Faster solving algorithms  
3. **Advanced features** - Competition mode, scramble generation
4. **Documentation perfection** - Every example verified to work

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
We have built world-class infrastructure and testing. Now we implement world-class algorithms. The fuzzing system will guide us - when it passes 100% of tests on all solvers, we'll have the best cube tool in the world.

**Next Action**: Start implementing real BeginnerSolver with proper layer-by-layer method.

---

## 💪 **MOTIVATION: Why We're Building This**

We're not building another toy cube solver. We're building the **definitive command-line cube tool** that:

- **Actually solves any cube** - No more placeholder algorithms
- **Teaches proper methods** - Real beginner and CFOP techniques  
- **Handles any size** - 2x2 through 10x10+ flawlessly
- **Has perfect reliability** - Fuzzing ensures it never fails
- **Provides amazing UX** - Beautiful output, web interface, comprehensive docs

**This will be the cube solver that becomes the standard.** Nothing less than perfection.