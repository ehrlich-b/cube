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
- Algorithm DB: **108 unique entries, all with verified inverse-to-solved patterns**;
  115/159 CSV rows accepted, 12 merged, 44 quarantined with reasons ✅
  Counts checked against [import-report.json](./alg_dumps/import-report.json) by Go tests.
- White first layer: cross + four corners, rotations, saved CFEN and recovery ✅
- Physical 3×3 input validation and independent cubie/geometry oracles ✅
- Full beginner solve: middle edges, yellow cross/alignment, corner placement/orientation ✅
- Kociemba two-phase: complete 3×3 default; deterministic lazy cached tables ✅
- Optimal face-turn IDA*: explicit time limit and no answer on timeout ✅
- CFOP: optimal white cross, database paired F2L + search fallback, complete database
  OLL/PLL recognition and AUF; CLI and staged site playback ✅
- `solving_db.go`: a 4-look pattern-matcher that is **dead code** (unwired) ⚠️
- Tests: active full-solver contract, Go tests, binary E2E cases, and independent
  physical-cubie/checkpoint/interaction oracles for full and partial goals.

**Measured 2026-10-07, native Mac background QoS:** 1,000 uniform legal states,
seed `2026100709`, in one process with tables loaded: **19.932 mean / 20 max
face turns, 100% at ≤20**, **5.68 ms warm mean / 57.22 ms p99 / 187.94 ms max**.
Fresh cached subprocess including its first solve: **625.88 ms**. The native
phase-one symmetry database has **295,004,160 states / 147,502,080 packed bytes**;
generation took **120.99 s**. WASM retains the smaller pair databases and is
outside these native latency measurements. `make bench-kociemba` reproduces and
gates the 1,000-state length and warm latency targets.

Optimal search adds full-corner and two six-edge databases: **44,089,920 +
21,288,960 + 21,288,960 bytes**, with **47,900,160 bytes** of factorized transitions.
Generation took **67.53 s**, measured peak RSS **665 MB**; cache **134,568,064 bytes**.
Two 12-turn states were independently cross-checked with the IDA* finder and
proved in **6.96 / 8.07 ms** using the larger tables. Superflip uses Reid's published
20-turn lower bound and a physically replayed witness, including all 24 grips.
The uniform-state benchmark (seed `2026100714`, one thread, CPU profiling enabled)
returned **0/10 proved and 10/10 timeouts at 180 s**; return times **180.002–180.209 s**.
The few-minute target remains **unmet**. `make bench-optimal` reproduces all ten
states and three-minute budgets; timeouts stay in the reported distribution. The paired
min2phaseCXX reference on the identical 1,000 states measured **5.28 ms mean /
53.22 ms p99**, **19.746 mean / 20 max turns**, with every answer physically
replayed. Its **991,712-byte cache / 24.73 ms cold cached process** beat the
engine's large-cache/cold-start costs. Source/data hashes and configurations are
in [the receipt](./test/solver_benchmark_20261007.json). Cube Explorer and
nissy/vcube paired timing remains outstanding.

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
*Goal: Import a comprehensive, reproducible algorithm database (only five original entries were live)*

### 1.1 Refactor Core Structure ✅ COMPLETE
- [x] Implement new Algorithm struct per `/docs/move_db_refactor.md`:
  - [x] Remove obsolete verification fields and commented legacy database
  - [x] Add `CaseID`, `Pattern`, `Recognition`, `Inverse`, `Mirror` fields
  - [x] Update all existing code references
- [x] Build pattern generation tool:
  - [x] Apply inverse algorithm to solved YB cube to generate recognition state
  - [x] Generate CFEN patterns automatically (`tools/generate-patterns/`)
  - [x] Updated 5 key algorithms with generated patterns (Sune, Anti-Sune, Cross OLL, T-Perm, Sexy Move)
- [x] Update CLI commands and database tools to work with new structure
- [x] Fix e2e tests - all 98 tests now passing ✅

### 1.2 Import Comprehensive Dataset
- [x] Create CSV import system for `/alg_dumps/` (9 files, 100+ algorithms)
- [x] Handle data quality issues (inconsistent formats, references)
- [x] Merge duplicates across CSV files
- [x] Auto-generate patterns for all imported algorithms
- [x] Support multi-dimensional algorithms (2x2, 4x4+, parity cases)

Imported 144 rows from 159; 18 duplicate sequences merged across categories;
15 quarantined (invalid references or broken claimed OLL/PLL stage behavior).
131 unique entries verify: 117 for 3×3, 14 for other dimensions. Category
memberships: F2L 41, OLL 46, PLL 18, Trigger 8, Advanced 7, Roux CMLL/LSE 2/2,
2×2 CLL/EG1/EG2/OLL/PBL 6/1/1/1/2, 4×4/5×5/6×6 parity 2/1/1.
Case descriptions for other dimensions remain unvalidated beyond inverse replay.

### 1.3 Database Enhancement
- [x] Identify inverse and mirror relationships automatically
- [x] Add algorithm lookup/search improvements
- [x] Create database validation tools
- [x] Update CLI commands to work with new structure

21 exact inverse pairs and 12 exact left/right mirror pairs (self-pairs included).
Aliases and all category memberships survive merging. Reports are reproducible.

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
  - [x] F2L slot predicate and masked corner-edge recognition patterns
  - [x] OLL solved predicate and yellow-sticker recognition patterns
  - [x] PLL full-state recognition patterns and full solved predicate
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
- [x] OLL recognition from pattern database and algorithm application
- [x] PLL recognition from piece positions and algorithm application
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
- [x] Shortest white cross via `FindPattern` (eight-turn bound)
- [x] Paired F2L database recognition, AUF/slot rotations, preserving solved slots
- [x] `FindPattern` F2L fallback when the database has no matching case
- [x] Complete OLL/PLL recognition: 216/288 states through database compositions
- [x] 200 uniform-state Go oracle, full-solver invariant, rotated-frame coverage
- [x] Site method picker and Cross/F2L 1–4/OLL/PLL case-name playback
- [ ] Cross optimization (extended cross, color neutrality)
- [ ] Advanced F2L with look-ahead
- [x] Algorithm-based OLL/PLL from database

200-state Go lengths (mean/max turns): Cross 5.790/8; F2L 1–4
6.820/12, 6.715/12, 7.025/12, 7.350/12; OLL 10.635/18; PLL 13.070/20;
total **57.405/73**. Face/wide/slice turns count one, rotations are separate.
The same Go sample took **8.81 s total, 44.04 ms mean, 835.87 ms maximum**,
including first-use setup. A separate Python 3D geometry oracle replayed every
checkpoint on **200 independent uniform physical states**, including all 24
rigid grips: **58.275 mean, 75 maximum turns**. Fresh CLI processes averaged
**835.14 ms**, maximum **1796.34 ms** (167.03 s total, including table setup in
every process).
The raw OLL/PLL sets are incomplete, so some stages use multiple named algorithms.
No look-ahead, extended cross, or color neutrality is claimed.

### 6.2 Kociemba Two-Phase
- [x] Phase 1: Reduce to &lt;U,D,R2,L2,F2,B2&gt; subgroup
- [x] Phase 2: Search subgroup with bounded DFS and admissible pruning (combined result is not globally optimal)
- [x] Generate deterministic lazy cached pruning tables and coordinate systems
- [x] Keep `KociembaSolver.Solve` / `GetSolver` signatures and beginner lesson intact
- [x] Default to Kociemba after measured paired length comparison
- [x] 200 seeded scramble oracle, 200 uniform cubie-state oracle and 200 independent
  Python physical-state replays; unchanged full-solver/CFEN/command guardrails
- [x] Continue after the first solution; tighten total length while increasing phase-one depth
- [x] Three reduction axes and inverse search, with resumable DFS sharing one budget
- [x] Combined twist/slice, flip/slice and twist/flip admissible pruning with conjugate checks; version-4 cache
- [x] CLI `--target-length` (20) and shared `--time-limit` (1s); best incumbent on expiry
- [x] Oracle gates: all sampled solutions at ≤20 turns; 1,000-state warm mean <50 ms and p99 <250 ms
- [x] WASM uses the same default budget; cold-worker completion and responsive UI regression
- [x] Exact symmetry-reduced phase-one pruning (291 twist classes × flip/slice)
- [x] Stabilizer closure and reverse-fill generation, checksummed four-bit native cache
- [x] One pre-move, early phase-two pruning and inverse edge/corner-combination bounds
- [x] Native 1,000-state sample: 100% at ≤20 turns, 5.68 ms mean and 57.22 ms p99
- [x] Full-corner and disjoint six-edge optimal databases, lazy cache generation
- [x] Optimal distance oracle: short/deep finder cross-checks and superflip = 20
- [ ] Reach the few-minute proven-optimal target: measured 0/10 at 180 s, with all ten observations censored
- [x] Paired min2phaseCXX comparison on the exact 1,000 fixtures, with physical replay and source hashes
- [x] Export seeded URFDLB reference datasets with `make export-reference-fixtures`
- [ ] Reduce the 147.5 MB native phase-one cache and cold-start cost toward the compact reference
- [ ] Run paired native reference benchmarks against Cube Explorer and nissy/vcube

Final validation on 2026-10-07 completed before 21:55 local: `go test -p 2 ./...`,
`go vet -p 2 ./...`, **133/133 CLI E2E cases**, `make test-kociemba`,
`make test-cfop`, `make test-beginner`, `make test-first-layer`, `make test-web`,
`make test-phase1-tables`, `make test-optimal`, `make bench-kociemba` and
`make export-reference-fixtures` passed. The full ten-state optimal benchmark
completed with CPU profiling; all ten observations remained 180-second timeouts.
The three load-bearing test files were left unchanged, and every new solver
answer was checked against the full-cube contract. The paired min2phaseCXX
run passed 1,000 independent physical replays and a cached-process replay.

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
- **Phase 6**: Complete, independently verified CFOP and Kociemba; report measured
  lengths honestly (CFOP ~57 turns; Kociemba ~20, with no 20-turn guarantee)
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
