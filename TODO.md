# TODO.md - Rubik's Cube Verified Algorithm Database

## 🎯 **NEW FOCUS: Building a Verified Algorithm Database**

We're pivoting from solver implementation to building a robust foundation of verified algorithms. This database will become the bedrock for future solver development.

---

## 🏗️ **PHASE 2: Enhanced Verification System**

### **Current State**
We have a basic `verify` command that checks if a solution solves a scramble. We need to dramatically enhance this to support verifying any algorithm with flexible start/end states.

### **The Vision: Universal Algorithm Verifier**

```bash
# Current (limited):
cube verify "R U R' U'" "U R U' R'"  # Does solution solve scramble?

# New (flexible):
cube verify "R U R' U R U2 R'" --start <cfen> --target <cfen>  # Does algorithm transform start to target?

# Examples:
# Verify Sune algorithm (OLL)
cube verify "R U R' U R U2 R'" \
  --start "YB|Y9/R3G3R3/B3W3B3/W9/O3Y3O3/G3R3G3" \  # OLL case
  --target "YB|Y9/?9/?9/?9/?9/?9"                    # Yellow face complete (rest wildcarded)

# Verify T-Perm (PLL)  
cube verify "R U R' U' R' F R2 U' R' U' R U R' F'" \
  --start "YB|Y9/?9/?9/W9/?9/?9" \                   # OLL complete
  --target "YB|Y9/R9/B9/W9/O9/G9"                    # Fully solved

# Verify F2L pair insertion
cube verify "U R U' R'" \
  --start "YB|?Y?YYY?Y?/?9/?9/W9/?9/?9" \           # White cross only
  --target "YB|?Y?YYY?Y?/??R??R??R/??B??B??B/W9/??O??O??O/??G??G??G"  # One F2L pair solved
```

### 🎯 **Verification Enhancement Tasks**

**1. Redesign `verify` Command**
- [ ] Rename current `verify` to `verify-solve` (backwards compatibility)
- [ ] Create new `verify` command with `--start` and `--target` CFEN support
- [ ] Default `--start` to solved cube if not specified
- [ ] Default `--target` to solved cube if not specified
- [ ] Support algorithm chains (multiple algorithms in sequence)
- [ ] Add `--verbose` mode showing intermediate states

**2. Wildcard Semantics**
- [ ] Document precise wildcard behavior: `?` means "don't care about this sticker"
- [ ] Ensure wildcards work correctly for:
  - OLL: Only yellow face orientation matters
  - PLL: Last layer permutation matters, first two layers must stay solved
  - F2L: Specific slots matter, others can be wildcarded
  - Cross: Only 4 edge pieces matter

**3. Algorithm Categories**
- [ ] Define standard CFEN patterns for common algorithm types:
  - Cross patterns
  - F2L slot patterns  
  - OLL patterns (57 standard cases)
  - PLL patterns (21 standard cases)
  - Special patterns (parities, etc.)

---

## 📚 **PHASE 3: Algorithm Database Schema**

### **Core Data Structure**

```go
type VerifiedAlgorithm struct {
    ID          string      // Unique identifier (e.g., "oll-27", "pll-t", "f2l-1")
    Name        string      // Human-readable name ("Sune", "T-Perm", etc.)
    Category    string      // "OLL", "PLL", "F2L", "CROSS", etc.
    Moves       string      // Algorithm notation ("R U R' U R U2 R'")
    StartCFEN   string      // Required starting state (with wildcards)
    TargetCFEN  string      // Expected ending state (with wildcards)
    
    // Metadata
    MoveCount   int         // Number of moves
    Recognition string      // Visual pattern description
    Probability float64     // Chance of occurring in solve
    Variants    []string    // Alternative algorithms for same case
    
    // Verification
    Verified    bool        // Has this been verified?
    TestedOn    []int       // Cube sizes tested (e.g., [3, 4, 5])
}
```

### 🎯 **Database Implementation Tasks**

**1. Storage Format**
- [ ] Design JSON schema for algorithm database
- [ ] Create `data/algorithms/` directory structure:
  - `data/algorithms/oll/`
  - `data/algorithms/pll/`
  - `data/algorithms/f2l/`
  - `data/algorithms/cross/`
  - `data/algorithms/special/`

**2. Algorithm Loading**
- [ ] Create `AlgorithmDB` type to load and manage algorithms
- [ ] Implement lazy loading from JSON files
- [ ] Add caching for frequently used algorithms
- [ ] Support hot-reloading during development

**3. Batch Verification**
- [ ] Create `verify-db` command to verify all algorithms in database
- [ ] Support parallel verification for speed
- [ ] Generate verification report
- [ ] Flag any algorithms that fail verification

---

## 🔬 **PHASE 4: Algorithm Collection & Curation**

### **Standard Algorithm Sets**

**1. OLL Algorithms (57 cases)**
- [ ] All edges oriented correctly (Cross cases)
- [ ] T shapes
- [ ] Pi shapes
- [ ] U shapes
- [ ] H shape
- [ ] L shapes
- [ ] Fish shapes
- [ ] Knight shapes
- [ ] Awkward shapes
- [ ] Dot cases

**2. PLL Algorithms (21 cases)**
- [ ] Edge permutations (U perms)
- [ ] Corner permutations (A perms)
- [ ] Adjacent swaps (J, T, F perms)
- [ ] Diagonal swaps (Y, V, N perms)
- [ ] Special cases (G perms, R perms)

**3. F2L Algorithms (41 standard cases)**
- [ ] Both pieces in top layer
- [ ] Corner in slot, edge in top
- [ ] Edge in slot, corner in top
- [ ] Both pieces in slots

**4. Advanced Algorithms**
- [ ] COLL (42 cases)
- [ ] ZBLL (493 cases - future)
- [ ] Winter Variation (27 cases)
- [ ] VLS (216 cases - future)

### 🎯 **Collection Tasks**

**1. Initial Algorithm Set**
- [ ] Start with basic OLL/PLL algorithms
- [ ] Source from well-known algorithm databases
- [ ] Ensure each algorithm has proper attribution
- [ ] Test on multiple cube sizes where applicable

**2. CFEN Pattern Generation**
- [ ] Create tool to generate StartCFEN from case description
- [ ] Create tool to generate TargetCFEN from algorithm effect
- [ ] Validate patterns are minimal (only specify necessary stickers)

**3. Quality Assurance**
- [ ] Each algorithm must be verified before inclusion
- [ ] Check for move cancellations and optimize
- [ ] Ensure consistent notation (no Rw vs r confusion)
- [ ] Add execution notes (finger tricks, etc.)

---

## 🛠️ **PHASE 5: Verification Infrastructure**

### **Testing Framework**

**1. Unit Tests**
- [ ] Test verifier with known algorithm/state combinations
- [ ] Test wildcard matching edge cases
- [ ] Test multi-algorithm chains
- [ ] Test on different cube sizes

**2. Integration Tests**
- [ ] Full database verification test
- [ ] Performance benchmarks
- [ ] Memory usage analysis
- [ ] Cross-platform compatibility

**3. Continuous Verification**
- [ ] GitHub Action to verify algorithms on PR
- [ ] Nightly full database verification
- [ ] Algorithm performance tracking

### **Visualization Tools**

**1. Algorithm Viewer**
- [ ] `cube show-alg <algorithm-id>` command
- [ ] Display start state, algorithm, end state
- [ ] Show intermediate states for learning
- [ ] Support `--animate` flag for step-by-step

**2. Pattern Recognition**
- [ ] `cube identify <cfen>` to find matching OLL/PLL cases
- [ ] Suggest algorithms based on current state
- [ ] Show multiple algorithm options

---

## 📊 **Success Metrics**

1. **Coverage**: 
   - ✅ All 57 OLL cases verified
   - ✅ All 21 PLL cases verified  
   - ✅ Core F2L cases verified
   - ✅ 4x4 parity algorithms verified

2. **Reliability**:
   - ✅ 100% of algorithms pass verification
   - ✅ Verification works on 3x3 through 7x7
   - ✅ No false positives/negatives

3. **Performance**:
   - ✅ Full database verification < 10 seconds
   - ✅ Single algorithm verification < 100ms
   - ✅ Memory usage < 100MB

---

## 🚀 **Why This Matters**

Building a verified algorithm database gives us:

1. **Foundation for Solvers**: Can't build reliable solvers without verified algorithms
2. **Learning Tool**: Users can explore and understand algorithms
3. **Consistency**: Single source of truth for algorithm notation
4. **Extensibility**: Easy to add new algorithms and methods
5. **Quality**: Every algorithm is tested and proven to work

This is the critical infrastructure piece we need before attempting to build world-class solvers.

---

## 📋 **Immediate Next Steps**

1. **Enhance `verify` command** to support flexible start/target states
2. **Create algorithm database schema** and storage format
3. **Implement first 10 OLL algorithms** as proof of concept
4. **Build verification test suite** 
5. **Document wildcard semantics** precisely

Let's build the world's most reliable cube algorithm database! 🎯