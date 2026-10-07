# CLAUDE.md

## Current milestone (2026-10-07)

Read TODO.md first. `cube solve` defaults to Kociemba; `--method beginner` and
`cube learn` retain the complete beginner method. `--goal first-layer` retains
the white cross/corners sublesson. Interactive lessons support actual moves,
undo/reset, rotated grips and saved CFEN recovery. `full_lesson.go` extends the
verified first-layer checkpoints; only complete solutions implement `Solver`.
Do not weaken the full-solver invariant. CFOP now solves Cross, F2L, OLL and PLL
from the imported database; unsupported dimensions are rejected. Kociemba has 200-scramble and 200-uniform-state Go
oracles plus an independent 200-state Python physical replay oracle.
`cube find` uses shortest coordinate IDA* for exact/wildcard 3×3 patterns with
restricted alphabets, retaining sticker BFS for other dimensions/move types.
`cube solve --optimal --time-limit 1s` errors if shortestness is not proved in time.
See README.md and docs/solvers.md for measured performance and table caching.

Run `make test-all`, `make test-first-layer`, `make test-beginner` (independent
Python stdlib geometry/cubie oracles), and `make fmt && make vet` before
committing. Run `make test-kociemba` too. E2E checks actively replay full solves.
Rotation, slice and far-layer permutations are checked by an independent 3D
sticker oracle. x/y/z follow R/U/F and E follows D. CFEN lesson input uses YB
storage order; actual centers determine orientation, never the prefix alone.
See examples/beginner.md and docs/verification-full-2026-10-03.md. Human physical
usability is not yet verified.

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 🚀 Quick Start

**IMPORTANT: Always check TODO.md first!** It contains the current development plan and status.

### Essential Commands for Testing

```bash
# Build main CLI and database tools
make build              # Main CLI only
make build-tools        # Database tools only
make build-all-local    # Everything locally

# Test basic cube functionality
./dist/cube twist "R U R' U'" --color
./dist/cube solve "R U R' U'" --color   # Complete 3×3 beginner solution

# Test enhanced verification system
./dist/cube verify "R U R' U'" --start "YB|Y9/R9/B9/W9/O9/G9" --target "YB|Y9/R9/B9/W9/O9/G9" --verbose
./dist/cube show "R U R' U'" --highlight-oll --color
./dist/cube lookup sune --preview

# Test database tools (separate utilities)
./dist/tools/verify-algorithm "T-Perm" --verbose
./dist/tools/verify-database --category OLL

# Cube sizes / notation (twist supports larger dimensions; beginner solving supports 3×3)
./dist/cube twist "R U R' U'" --dimension 3
./dist/cube twist "Rw Uw Fw" --dimension 4 --color

# Run comprehensive test suite 
make test-all
```

## 🛡️ Invariants & Guardrails (read before touching engine/CFEN code)

These are the load-bearing tests. **If any go red, stop and fix that first** — a red invariant means the move engine is corrupting state, the CFEN backbone is lying, or a "solver" is emitting solutions that don't actually solve. They are designed to stay green and trustworthy while everything else changes.

- `internal/cube/invariants_test.go`
  - **Sticker conservation** — every move is a permutation; each color stays at exactly N² across sizes 2–6.
  - **Scramble + inverse = solved** — the core "moves are consistent permutations" check.
  - **Determinism** — same sequence ⇒ same state.
  - **Solver contract (the big one)** — *if a solver returns a non-empty solution, applying it MUST solve the cube.* Stubs return empty, so the test SKIPs (never a false pass); the moment a solver emits real moves it is held to actually solving. **Never weaken this test to make a solver "pass."**
- `internal/cfen/cfen_test.go` — canonical solved CFEN, cube<->CFEN round-trip (all 4 orientations), wildcard matching, verify semantics.
- `internal/cli/commands_test.go` — no command is registered twice.

Run with `make test` (Go unit tests) or `make test-all` (adds the E2E suite). All must be green before committing.

## Cube Orientation

**Canonical Starting Orientation:**
- 🟨 **Yellow** on top (Up face)
- ⬜ **White** on bottom (Down face)  
- 🟦 **Blue** facing front (Front face)
- 🟩 **Green** facing back (Back face)
- 🟧 **Orange** on left (Left face)
- 🟥 **Red** on right (Right face)

**Customizing Orientation:**
If you prefer a different orientation, use cube rotations before your scramble:
```bash
# Standard orientation
./dist/cube solve "R U R' U'" --color

# Rotate to different orientation first  
./dist/cube solve "x y R U R' U'" --color  # x = pitch, y = yaw
./dist/cube solve "z R U R' U'" --color    # z = roll
```

Available rotations: `x`, `y`, `z` (with `'` for counter-clockwise, `2` for 180°)

## Development Commands

**All commands are cross-platform (macOS/Linux compatible)**

**Build and Run:**
- `make build` - Compile main CLI binary to `dist/cube`
- `make build-tools` - Compile database tools to `dist/tools/`
- `make build-all-local` - Build main CLI + database tools locally
- `make build-all` - Build for multiple platforms (linux, darwin, windows)
- `make clean` - Remove build artifacts
- `make dev` - Hot reload development (requires Air: `make install-tools`)
- `make run` - Run CLI directly with go run

**Code Quality (ALWAYS run before committing):**
- `make test` - Run unit tests
- `make e2e-test` - Run end-to-end test suite
- `make test-all` - Run both unit and e2e tests
- `make fmt` - Format Go code (cross-platform compatible)
- `make vet` - Static analysis
- `make lint` - Lint with golangci-lint (requires: `make install-tools`)

**Dependencies:**
- `make install` - Download and tidy Go modules
- `make install-tools` - Install Air and golangci-lint

## Architecture Overview

This is a Rubik's cube solver CLI tool with a clean architecture separating end-user functionality from database curation tools:

```
cmd/cube/main.go (Clean CLI)
    ↓
internal/cli/ (Cobra commands) → internal/cube/ (Core logic)
                                       ↓
                              internal/cfen/ (CFEN verification)
                                       ↑
tools/ (Database utilities) ────────────┘
├── verify-algorithm/
└── verify-database/
```

### Core Components

**Cube Representation (`internal/cube/cube.go`):**
- `Cube` struct supports NxNxN cubes (2x2, 3x3, 4x4+) 
- Uses `[6][][]Color` for six faces with dynamic sizing
- Standard Singmaster notation parsing (R, U', F2, etc.)

**Algorithm Database (`internal/cube/algorithms.go`):**
- `Algorithm` struct: `Name`, `CaseID`, `Category`, `Moves`, `Pattern` (concrete inverse-to-solved CFEN), `Recognition`, `Inverse`, `Mirror`
- 131 unique algorithms, all with inverse-to-solved CFEN patterns (117 for 3×3)
- 144 CSV rows accepted, 18 merged across categories, 15 quarantined with reasons
- `tools/import-algorithms` reproducibly generates the embedded JSON and reports

**CFEN Verification System (`internal/cfen/`):**
- CFEN parsing and generation with wildcard (`?`) support
- Cube<->CFEN conversion; round-trip is verified for all 4 supported orientations (YB/WB/YG/WG)
- `MatchesCube()` wildcard pattern matching — this is the engine behind `cube verify`
- **The canonical orientation is YB (yellow-up, blue-front) and is by far the most exercised path.** Non-YB orientations round-trip but their rotation semantics are not otherwise tested.

**Solver System (`internal/cube/solver.go`):**
- Interface-driven design: `type Solver interface { Solve(*Cube) (*SolverResult, error); Name() string }`
- Three registered solvers: BeginnerSolver, CFOPSolver, KociembaSolver
- BeginnerSolver completes a validated 3×3 via `PlanBeginner`; KociembaSolver uses two-phase coordinate search. CFOP uses optimal cross search, paired F2L recognition/search and complete database OLL/PLL tables.
- `internal/cube/solving_db.go` sketches a 4-look pattern-match solver but is **dead code** (unwired to any command)

**Main CLI Commands (`internal/cli/`):** (with honest status)
- `cube twist` - apply moves, display result (+ `--cfen`) — **works**
- `cube solve` - Kociemba by default; beginner selectable; unavailable algorithms/dimensions error
- `cube learn` - complete beginner checkpoints and recovery; explicit `--goal first-layer` supported
- `cube verify <alg> --start <cfen> --target <cfen>` - CFEN verification — **works** (YB). Note: takes ONE positional arg (the algorithm), not two.
- `cube show` - display with cross/OLL/PLL/F2L highlighting — **works**
- `cube lookup` - algorithm database lookup — **works**
- `cube optimize` - move cancellation (`R R R`->`R'`) — **works**
- `cube find` - shortest coordinate IDA* for 3×3 patterns and restricted moves, sticker BFS fallback
- `cube identify` / `show-alg` - **partial** (recognition logic is a TODO)
- `cube parse-cfen` / `generate-cfen` / `verify-cfen` / `match-cfen` - CFEN utilities — **work**
- Built with Cobra framework

**Database Tools (`tools/`):**
- `verify-algorithm` - Single algorithm verification using cube package as library
- `verify-database` - Batch verification of all algorithms with CFEN patterns
- Separate binaries for specialized database curation workflows

### Key Data Flow

**Solving Process:**
1. Parse scramble string into `[]Move`
2. Create `Cube` with specified dimension
3. Apply scramble moves to cube
4. Get solver by algorithm name from factory
5. Execute `solver.Solve(cube)` → `SolverResult` (BeginnerSolver solves the complete 3×3 without mutating input)
6. Replay and verify the full solution, then format output for CLI display

### Current Implementation Status

**✅ Completed Features:**
- Full NxN cube support (2x2 through 6x6+) with proper multi-layer moves
- Beautiful ASCII color output with `--color` flag (ANSI colored letters)
- Advanced move notation: M/E/S slices, Rw/Fw wide moves, 2R/3L layer moves, x/y/z rotations
- **Enhanced verification system** - `cube verify` command with flexible CFEN start/target support
- **CFEN infrastructure** - Complete parsing, generation, and wildcard matching
- **Pattern highlighting system** - `cube show` with cross/OLL/PLL/F2L highlighting
- **Algorithm database** - 131 unique entries with verified recognition patterns
- **Verified algorithm collection** - all 131 entries pass `verify-database` (inverse pattern → solved)
- **Clean architecture** - Separate database tools from main CLI
- **Database verification tools** - Standalone utilities for algorithm curation
- **Comprehensive test suite** - End-to-end tests, Go unit tests and independent Python physical replay
- **Invariant guardrail suite** - load-bearing engine/solver/CFEN invariants (see "Invariants & Guardrails" near the top)
- Cross-platform build system (macOS/Linux compatible)

**⚠️ Current Issues / Known Gaps:**
- Larger-cube solving remains unimplemented; CFOP color neutrality and look-ahead remain future work.
- Beginner physical usability remains unverified by a human trial.
- Incorrect raw CSV cases remain in quarantine; other-size parity descriptions need independent validation.
- `internal/cube/solving_db.go` is a dead-code 4-look pattern-matcher — wire it up or delete it
- `internal/cube/cubie.go` supplies piece addresses, ranges and 3×3 selector aliases
- All nine CSV algorithm dumps in `alg_dumps/` are imported reproducibly, with rejected rows quarantined.

**📍 Key Files to Know:**
- `TODO.md` - **ALWAYS READ FIRST** - Current development plan and progress
- `internal/cube/cube.go` - Core cube representation, color output methods
- `internal/cube/moves.go` - Move parsing and application logic
- `internal/cube/solver.go` / `kociemba.go` - Full beginner, Kociemba and CFOP implementations (`cfop.go`)
- `internal/cube/algorithms.go` - Embedded algorithm database (131 entries, all with CFEN recognition patterns)
- `internal/cfen/` - Complete CFEN parsing, generation, and verification system
- `internal/cli/verify.go` - Enhanced verification command with CFEN support
- `internal/cli/solve.go` - CLI solve command with algorithm selection
- `internal/cli/show.go` - Cube display with pattern highlighting
- `internal/cli/lookup.go` - Algorithm database lookup command
- `tools/verify-algorithm/` - Standalone algorithm verification tool
- `tools/verify-database/` - Standalone database verification tool
- `tools/README.md` - Documentation for database tools
- `test/e2e_test.sh` - Comprehensive end-to-end test suite
- **Project Documentation:**
  - `/docs/move_db_refactor.md` - Algorithm database refactor design
  - `/docs/move_visualization.md` - Enhanced last-layer visualization
  - `/docs/solvers.md` - Solver analysis and implementation roadmap

### Adding New Features

**New Solver Algorithm:**
1. Implement `Solver` interface in `internal/cube/solver.go`
2. Add to `GetSolver()` factory function
3. Update CLI flag validation in `internal/cli/solve.go`

**New Cube Dimension:**
- Cube struct already supports arbitrary dimensions
- Validate in CLI flag parsing

### Dependencies and Tools

**Runtime Dependencies:**
- `github.com/spf13/cobra` - CLI framework

**Development Tools (optional):**
- `cosmtrek/air` - Hot reload for development
- `golangci/golangci-lint` - Advanced linting

The codebase uses minimal dependencies and follows standard Go project layout conventions.

## 🔧 Troubleshooting & Common Tasks

### Display Format Options
- **Default (no --color)**: Black and white letters in clean unfolded cross layout
- **Unicode blocks (--color)**: Colorful emoji blocks in unfolded cross layout (recommended)
- **Colored letters (--color --letters)**: ANSI colored letters in unfolded cross layout
- The unfolded cross layout shows all faces in a traditional cube net format for easy visualization

**Example outputs:**
```
# Default: Clean and readable
    BBR       <- Up face (aligned with Front) 
    BBW
    BBW

YRR WWG OOB YOO <- Left | Front | Right | Back
RRR WWB YOO YYY
RRR WWW BOO YYY

    GGO       <- Down face (aligned with Front)
    GGG
    GGG

# Unicode blocks: Visual and intuitive  
    🟦🟦🟥      <- Up face (aligned with Front)
    🟦🟦⬜
    🟦🟦⬜

🟨🟥🟥 ⬜⬜🟩 🟧🟧🟦 🟨🟧🟧 <- Left | Front | Right | Back
🟥🟥🟥 ⬜⬜🟦 🟨🟧🟧 🟨🟨🟨
🟥🟥🟥 ⬜⬜⬜ 🟦🟧🟧 🟨🟨🟨

    🟩🟩🟧      <- Down face (aligned with Front)
    🟩🟩🟩
    🟩🟩🟩
```

### Solver status (IMPORTANT)
The default Kociemba solver completes a valid 3×3 and its returned moves are
checked against all six faces. Use `--method beginner` for the beginner solver
or `--goal first-layer` for the partial goal. `--method cfop` returns named, verified stage checkpoints.
```bash
./dist/cube solve "R U R' U'" --headless
./dist/cube learn "R U F2 L' B" --interactive
```

### Common Development Patterns

**Adding a new CLI command:**
1. Create command file in `internal/cli/` (e.g., `mycommand.go`)
2. Register command in `internal/cli/root.go` with `rootCmd.AddCommand(myCmd)`
3. Add tests to `test/e2e_test.sh` for the new command
4. Run `make test-all` to verify everything works

**Adding a new CLI flag:**
1. Add flag in `init()` function of relevant command file
2. Retrieve with `cmd.Flags().GetType("flag-name")`
3. Pass to core logic functions
4. Add test cases for the new flag

**Testing move notation:**
```bash
# Test advanced moves on different cube sizes (beginner solving supports 3×3 only)
./dist/cube twist "M E S" --dimension 3 --color        # Slice moves
./dist/cube twist "Rw Fw' Uw2" --dimension 4 --color    # Wide moves  
./dist/cube twist "2R 3L'" --dimension 5 --color       # Layer moves

# Use verify to test an algorithm against CFEN start/target states (ONE positional arg + flags)
./dist/cube verify "R U R' U'" --start "YB|Y9/R9/B9/W9/O9/G9" --target "YB|Y9/R9/B9/W9/O9/G9"  # fails (sexy move is not the identity)
./dist/cube verify "R U R' U' R U R' U' R U R' U' R U R' U' R U R' U' R U R' U'" --start "YB|Y9/R9/B9/W9/O9/G9" --target "YB|Y9/R9/B9/W9/O9/G9"  # passes (sexy x6 = identity)

# Use show to visualize patterns
./dist/cube show "R U R' U'" --highlight-oll --color   # Highlight top layer
```

### Database Tools and Algorithm Curation

**Working with the Algorithm Database:**
```bash
# Build database tools
make build-tools

# List all algorithms with CFEN patterns
./dist/tools/verify-algorithm --list

# Verify a specific algorithm
./dist/tools/verify-algorithm "T-Perm" --verbose

# Verify all algorithms in the database
./dist/tools/verify-database

# Verify only specific categories
./dist/tools/verify-database --category OLL --verbose
```

**Adding New Verified Algorithms:**
1. Add a CSV algorithm row and run `make import-algorithms` to regenerate patterns
2. Use `./dist/tools/verify-algorithm` to test the algorithm
3. Use `./dist/tools/verify-database` to ensure database consistency
4. Review the generated import report and any quarantined rows

**CFEN Pattern Development:**
```bash
# Generate CFEN from a cube state
./dist/cube twist "R U R' U'" --cfen

# Test verification with specific patterns
./dist/cube verify "R U R' U'" --start "YB|Y9/R9/B9/W9/O9/G9" --target "YB|scrambled_pattern" --verbose
```

**Before starting any work:**
1. Read TODO.md to understand current phase
2. Run `make build-all-local` to build CLI + tools
3. Check if tests pass with `make test-all` (runs Go and E2E tests)
4. Test database tools with `./dist/tools/verify-database`
5. Always run `make fmt && make vet` before committing

**Test Suite Coverage:**
- **Comprehensive end-to-end tests** covering every CLI command and feature
- **All cube dimensions** (2x2 through 20x20) with proper multi-layer moves
- **Advanced notation** (M/E/S slices, Rw/Fw wide moves, 2R/3L layer moves, x/y/z rotations)
- **Enhanced verification system** (CFEN patterns, wildcard matching, database verification)
- **Error handling and edge cases** with proper exit codes
- **Integration tests** (solve + verify workflows)
- **Performance tests** (large scrambles, complex cubes)
- **CFEN parsing and generation tests** (cube state conversion, pattern matching)
- **Bash E2E harness** plus optional independent Python stdlib physical replay oracles
