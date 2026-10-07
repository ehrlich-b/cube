# TODO.md - Rubik's Cube Solver Project

## 🔍 Project Status: Reality Check

This project has built a correct, well-tested engine for cube manipulation, verification,
optimization, and search. **The complete 3×3 beginner solver and lesson work as of 2026-10-03.**
`cube solve` defaults to Kociemba as of 2026-10-07; `--method beginner` and
`cube learn` retain the complete beginner sequence with
recoverable playback. The white-layer sublesson remains explicit via `--goal first-layer`. This TODO is a pragmatic path forward.

**Verified reality (current):**
- Move engine: correct, NxN, covered by fuzz + invariant tests ✅
- CFEN verify / optimize: working ✅
- Find: coordinate IDA* for exact/wildcard 3×3 targets, restricted moves and BFS fallback ✅
- Algorithm DB: 63 entries, **5** with verification patterns (Sune, Anti-Sune, Cross OLL, T-Perm, Sexy Move) ✅
- White first layer: cross + four corners, rotations, saved CFEN and recovery ✅
- Physical 3×3 input validation and independent cubie/geometry oracles ✅
- Full beginner solve: middle edges, yellow cross/alignment, corner placement/orientation ✅
- Kociemba two-phase: complete 3×3 default; deterministic lazy cached tables ✅
- Optimal face-turn IDA*: explicit time limit and no answer on timeout ✅
- CFOP remains unimplemented; CLI rejects it ❌
- `solving_db.go`: a 4-look pattern-matcher that is **dead code** (unwired) ⚠️
- Tests: active full-solver contract, Go tests, binary E2E cases, and independent
  physical-cubie/checkpoint/interaction oracles for full and partial goals.

**Measured 2026-10-07, Mac background QoS:** fresh depth 8/9/10 CLI searches
0.72/0.46/0.52 s; loaded-table means across ten targets each 0.29/1.58/9.40 ms.
200 independently constructed Python physical states: Kociemba mean 21.73,
max 22 face turns; process mean 71.5 ms, max 365.7 ms. Thirty paired scrambles:
Kociemba 21.70 vs beginner 209.57 moves. Table generation 0.67 s, cached load
26 ms; search table generation 0.49 s. The 22-turn preference has a one-second
budget; difficult cases can use the complete 30-turn two-phase bound.

**Guardrails (do not let these go red):** `internal/cube/invariants_test.go`,
`internal/cfen/cfen_test.go`, `internal/cli/commands_test.go`. The solver-contract test is the
acceptance gate for Phase 4 — a non-empty solution must actually solve the cube, or it fails.

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

### Guardrails & Truth Pass (done 2026-06-01)
- [x] Add load-bearing invariant suite: engine (sticker conservation, scramble+inverse, determinism), solver contract, CFEN round-trip/wildcards, no-duplicate-commands
- [x] Fix WG CFEN reverse-mapping bug (round-trip was broken for the WG orientation)
- [x] Fix duplicate command registration (`find`, `optimize` were registered twice)
- [x] Fix stale e2e assertion (`verify-algorithm --list` now emits "HAS PATTERN", not "VERIFIED")
- [x] `go mod tidy` — drop `gorilla/mux` left over from the removed web interface
- [x] Deep-update anchor docs (CLAUDE.md, README.md, docs/solvers.md, this file) to match reality

---

## 🗃️ Phase 1: Algorithm Database Modernization
*Goal: Replace current 67-algorithm database with comprehensive, well-structured system*

### 1.1 Refactor Core Structure ✅ COMPLETE
- [x] Implement new Algorithm struct per `/docs/move_db_refactor.md`:
  - [~] Remove `Verified`, `TestedOn`, `StartCFEN`, `TargetCFEN` fields — PARTIAL: `Verified` bool + `MarkVerified`/`GetVerifiedAlgorithms` still present in algorithms.go
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
- [x] Concrete 3×3 edge/corner tracking in `pieces.go` (`FindEdge`, `FindCorner`,
  center-relative solved predicates); generic NxN tracking below remains future work.
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
  - [x] White cross predicate (4 white edges in correct positions)
  - [x] White first-layer predicate (cross + 4 corners)
  - [ ] F2LSlotPattern (corner-edge pair in position)
  - [ ] OLLSolvedPattern (all yellow stickers on top)
  - [ ] PLLSolvedPattern (last layer permuted correctly)
- [ ] Connect patterns to CFEN system for verification

---

## 🚀 Phase 4: First Working Solver
*Goal: Implement beginner method solver that actually solves cubes*

### 4.1 White Cross Solver
- [x] Cross piece finding and fixed insertion order (not a globally optimal cross)
- [x] Bounded move generation restores previously placed edges at checkpoints
- [x] Connect cross solving to CLI with piece checkpoints

### 4.1b White Corners / First-Layer Lesson (2026-10-03)
- [x] Insert four white corners with U setups and repeated right-hand triggers
- [x] Normalize all 24 rigid orientations and keep physical/model frames aligned
- [x] Preserve the cross and earlier corners at each corner checkpoint
- [x] Reject incomplete/unreachable sticker states before searching
- [x] Interactive actual-move recovery, undo, reset, save/resume and repeated-step safety
- [x] Independent physical geometry and cubie-state replay tests; bounded Mac performance measurements

### 4.2 First Two Layers (F2L)
- [x] Beginner middle-layer insertion (four edges); preserve white and earlier belt edges
- [ ] Implement intuitive F2L (not advanced algorithms)
- [ ] Find corner-edge pairs and position above slots
- [ ] Insert using basic algorithms and track completed slots

### 4.3 Last Layer
- [x] Beginner yellow cross, edge alignment and corner placement via named algorithm groups
- [x] Atomic four-corner orientation sweep restores all six solved faces
- [ ] OLL recognition from pattern database and algorithm application
- [ ] PLL recognition from piece positions and algorithm application
- [x] Verification that cube is fully solved

### 4.4 Integration & Testing
- [x] Generate random scrambles and solve end-to-end
- [x] Verify all solutions actually solve the cube, including independently generated cubie states
- [x] Add real full solving tests to e2e suite; remove beginner placeholder skips
- [x] Benchmark solving performance on the Mac

---

## 🔍 Phase 5: Search & Optimization
*Goal: Add search-based solving for better solutions*

### 5.1 Basic Search Implementation
- [x] Retain correct sticker BFS with independent path storage as fallback
- [x] Add coordinate IDA* with depth limits and shortest-sequence guarantees
- [x] Create duplicate detection and solution extraction for BFS

### 5.2 Heuristic Search
- [ ] Implement A* search with heuristic functions
- [x] Create pattern databases (corner/edge orientation, slice, permutation, four-edge groups)
- [x] Build admissible pruning tables for exact and wildcard pattern search
- [x] Match retained BFS on 40 cases through depth four and independent sticker
  BFS on 60 random start/target pairs through depth six
- [x] Verify exact depths 8–10, restricted move alphabets and non-3×3 fallback
- [x] Add `cube solve --optimal --time-limit` with explicit timeout errors

---

## 🎓 Phase 6: Advanced Methods
*Goal: Implement CFOP and Kociemba solvers*

### 6.1 CFOP Implementation
- [ ] Cross optimization (extended cross, color neutrality)
- [ ] Advanced F2L with look-ahead
- [ ] Algorithm-based OLL/PLL from database

### 6.2 Kociemba Two-Phase
- [x] Phase 1: Reduce to &lt;U,D,R2,L2,F2,B2&gt; subgroup
- [x] Phase 2: Search subgroup with exact-depth IDA* (combined result is not globally optimal)
- [x] Generate deterministic lazy cached pruning tables and coordinate systems
- [x] Keep `KociembaSolver.Solve` / `GetSolver` signatures and beginner lesson intact
- [x] Default to Kociemba after measured paired length comparison
- [x] 200 seeded scramble oracle, 200 uniform cubie-state oracle and 200 independent
  Python physical-state replays; unchanged full-solver/CFEN/command guardrails

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
- [x] Add first-layer explanations and checkpoint playback (`cube learn`)
- [x] Extend explanation/playback to a full beginner solver
- [ ] Human beginner physical-cube trial (simulated replay is not a human trial)
- [ ] Create difficulty settings and solving statistics
- [ ] Implement progress tracking and hints

### 7.3 Integration
- [x] Static interactive website backed by the Go WASM engine (3D/net, playback,
      lessons, worker pattern search, CFEN and URL sharing; `make web`)
- [x] Node WASM API regression tests and headless Chromium smoke (`make test-web`,
      `make test-web-smoke`)
- [ ] Configure and publish GitHub Pages (manual; documented in README)
- [ ] Web API for solving service (the static website does not require a service)
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
