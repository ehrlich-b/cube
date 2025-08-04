# TODO.md - Rubik's Cube Solver Project

## 🔍 Project Status: Reality Check

This project has built excellent infrastructure for cube manipulation and algorithm verification, but **the actual solving functionality is completely unimplemented**. This TODO represents an honest assessment and a pragmatic path forward.

---

## 📋 Phase 0: Foundation & Cleanup
*Goal: Clean house and establish solid foundations*

### Documentation Truth Reconciliation
- [x] Update CLAUDE.md to reflect actual counts:
  - [x] Change "60+ algorithms" → "67 algorithms"
  - [x] Change "79 end-to-end tests" → "98 end-to-end tests"
  - [x] Remove claim of "basic placeholder implementations" for solvers
  - [x] Add note that solvers are completely unimplemented
  - [x] Add note about CSV algorithm dumps ready for import
- [x] Update README.md to be transparent about solver status
- [x] Add links to `/docs/` project documentation

### Code Organization
- [x] Document unused `cubie.go` addressing system (future piece tracking)
- [x] Document `permutations.go` alternative move system (performance option)
- [x] Add comments to `solver.go` clarifying unimplemented status
- [x] Document `solving_db.go` experimental pattern-matching approach

---

## 🗃️ Phase 1: Algorithm Database Modernization
*Goal: Replace current 67-algorithm database with comprehensive, well-structured system*

### 1.1 Refactor Core Structure ✅ COMPLETE
- [x] Implement new Algorithm struct per `/docs/move_db_refactor.md`:
  - [x] Remove `Verified`, `TestedOn`, `StartCFEN`, `TargetCFEN` fields
  - [x] Add `CaseID`, `Pattern`, `Recognition`, `Inverse`, `Mirror` fields
  - [x] Update all existing code references
- [x] Build pattern generation tool:
  - [x] Apply algorithm to solved YB cube
  - [x] Generate CFEN patterns automatically (`tools/generate-patterns/`)
  - [x] Updated 5 key algorithms with generated patterns (Sune, Anti-Sune, Cross OLL, T-Perm, Sexy Move)
- [x] Update CLI commands and database tools to work with new structure
- [x] Fix e2e tests - all 98 tests now passing ✅

### 1.2 Import Comprehensive Dataset
- [ ] Create CSV import system for `/alg_dumps/` (9 files, 100+ algorithms)
- [ ] Handle data quality issues (inconsistent formats, references)
- [ ] Merge duplicates across CSV files
- [ ] Auto-generate patterns for all imported algorithms
- [ ] Support multi-dimensional algorithms (2x2, 4x4+, parity cases)

### 1.3 Database Enhancement
- [ ] Identify inverse and mirror relationships automatically
- [ ] Add algorithm lookup/search improvements
- [ ] Create database validation tools
- [ ] Update CLI commands to work with new structure

---

## 🎨 Phase 2: Enhanced Visualization
*Goal: Better algorithm display and pattern recognition*

### 2.1 Context-Aware Display
- [ ] Implement last layer mode per `/docs/move_visualization.md`
- [ ] Auto-detect when algorithms only affect top 2 layers
- [ ] Create 5x5 grid view (top face + surrounding edges)
- [ ] Keep full cube view for F2L and multi-layer algorithms

### 2.2 CLI Integration
- [ ] Add view mode flags: `--view=last`, `--view=full`, `--view=both`
- [ ] Update `cube show-alg` command
- [ ] Test with OLL/PLL algorithms for clarity

---

## 🧩 Phase 3: Core Solving Infrastructure
*Goal: Build the foundation needed for ANY solving algorithm*

### 3.1 Piece Tracking System
- [ ] Implement piece identification:
  - [ ] Define PieceType (Corner, Edge, Center)
  - [ ] Track 8 corners (3 colors each)
  - [ ] Track 12 edges (2 colors each)
  - [ ] Handle centers (fixed on odd cubes, mobile on even)
- [ ] Build piece location mapping:
  - [ ] `GetPieceByColors(colors []Color) *Piece`
  - [ ] `GetPieceLocation(piece *Piece) Position`
  - [ ] `IsPieceInCorrectPosition(piece *Piece) bool`
  - [ ] `IsPieceCorrectlyOriented(piece *Piece) bool`

### 3.2 Semantic Pattern Recognition
- [ ] Define pattern interface for cube states
- [ ] Implement concrete patterns:
  - [ ] WhiteCrossPattern (4 white edges in correct positions)
  - [ ] WhiteLayerPattern (cross + 4 corners)
  - [ ] F2LSlotPattern (corner-edge pair in position)
  - [ ] OLLSolvedPattern (all yellow stickers on top)
  - [ ] PLLSolvedPattern (last layer permuted correctly)
- [ ] Connect patterns to CFEN system for verification

---

## 🚀 Phase 4: First Working Solver
*Goal: Implement beginner method solver that actually solves cubes*

### 4.1 White Cross Solver
- [ ] Implement cross piece finding and optimal insertion order
- [ ] Build move generation to position edges without disturbing placed pieces
- [ ] Create solving logic with progress tracking

### 4.2 First Two Layers (F2L)
- [ ] Implement intuitive F2L (not advanced algorithms)
- [ ] Find corner-edge pairs and position above slots
- [ ] Insert using basic algorithms and track completed slots

### 4.3 Last Layer
- [ ] OLL recognition from pattern database and algorithm application
- [ ] PLL recognition from piece positions and algorithm application
- [ ] Verification that cube is fully solved

### 4.4 Integration & Testing
- [ ] Generate random scrambles and solve end-to-end
- [ ] Verify all solutions actually solve the cube
- [ ] Add solving tests to e2e suite
- [ ] Benchmark solving performance

---

## 🔍 Phase 5: Search & Optimization
*Goal: Add search-based solving for better solutions*

### 5.1 Basic Search Implementation
- [ ] Implement breadth-first search with state representation
- [ ] Add iterative deepening with depth limits
- [ ] Create duplicate detection and solution extraction

### 5.2 Heuristic Search
- [ ] Implement A* search with heuristic functions
- [ ] Create pattern databases (corner/edge orientation)
- [ ] Build pruning tables for search optimization

---

## 🎓 Phase 6: Advanced Methods
*Goal: Implement CFOP and Kociemba solvers*

### 6.1 CFOP Implementation
- [ ] Cross optimization (extended cross, color neutrality)
- [ ] Advanced F2L with look-ahead
- [ ] Algorithm-based OLL/PLL from database

### 6.2 Kociemba Two-Phase
- [ ] Phase 1: Reduce to &lt;U,D,R2,L2,F2,B2&gt; subgroup
- [ ] Phase 2: Solve within subgroup optimally
- [ ] Generate pruning tables and coordinate systems

### 6.3 Big Cube Support
- [ ] 4x4 reduction method (centers, edges, parity)
- [ ] 5x5+ support with generalized algorithms

---

## 📊 Phase 7: Polish & Performance
*Goal: Production-ready solver with great UX*

### 7.1 Optimization
- [ ] Profile and optimize hot paths
- [ ] Implement move cancellation and solution compression
- [ ] Add caching for common patterns

### 7.2 User Experience
- [ ] Add solve explanation mode and step-by-step playback
- [ ] Create difficulty settings and solving statistics
- [ ] Implement progress tracking and hints

### 7.3 Integration
- [ ] Web API for solving service
- [ ] Export solutions in standard notation
- [ ] Competition timer integration

---

## 🎯 Success Criteria

- **Phase 1**: Comprehensive algorithm database with 100+ algorithms and auto-generated patterns
- **Phase 2**: Clean last-layer visualization for OLL/PLL algorithms
- **Phase 3**: Working piece tracking and pattern recognition systems
- **Phase 4**: Beginner method that solves any valid 3x3 scramble
- **Phase 5**: Sub-second solving with search optimization
- **Phase 6**: Multiple solving methods (CFOP, Kociemba) with &lt;20 move average
- **Phase 7**: Production-ready solver with &lt;100ms response time

---

## 🔗 References

- **Algorithm Database Design**: `/docs/move_db_refactor.md`
- **Visualization Design**: `/docs/move_visualization.md`
- **Solver Analysis**: `/docs/solvers.md`
- **Raw Algorithm Data**: `/alg_dumps/` (9 CSV files)

---

## 💡 Development Philosophy

1. **Iterative Progress**: Each phase should produce working improvements
2. **Quality First**: Write tests before implementation to ensure correctness
3. **Honest Assessment**: Update progress based on reality, not aspirations
4. **Practical Focus**: The best solver is one that actually solves cubes

**Remember**: Perfect is the enemy of good. A working beginner solver beats a perfect plan.