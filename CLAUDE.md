# CLAUDE.md

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
./dist/cube solve "R U R' U'" --color   # NOTE: solve is a STUB — prints the cube but an EMPTY solution

# Test enhanced verification system
./dist/cube verify "R U R' U'" --start "YB|Y9/R9/B9/W9/O9/G9" --target "YB|Y9/R9/B9/W9/O9/G9" --verbose
./dist/cube show "R U R' U'" --highlight-oll --color
./dist/cube lookup sune --preview

# Test database tools (separate utilities)
./dist/tools/verify-algorithm "T-Perm" --verbose
./dist/tools/verify-database --category OLL

# Cube sizes / notation (solve still emits no solution — these just show the scrambled state)
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

Run with `make test` (Go unit tests) or `make test-all` (adds the 98 e2e tests). All must be green before committing.

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
- `make e2e-test` - Run end-to-end test suite (98 comprehensive tests)
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
- `Algorithm` struct: `Name`, `CaseID`, `Category`, `Moves`, `Pattern` (masked CFEN), `Recognition`, `Inverse`, `Mirror`
- 63 algorithms defined (OLL, PLL, F2L, triggers)
- **Only 5 carry a verification `Pattern`:** Sune, Anti-Sune, Cross OLL, T-Perm, Sexy Move
- A legacy `Verified` bool still lingers on a few entries — the db refactor was only partial

**CFEN Verification System (`internal/cfen/`):**
- CFEN parsing and generation with wildcard (`?`) support
- Cube<->CFEN conversion; round-trip is verified for all 4 supported orientations (YB/WB/YG/WG)
- `MatchesCube()` wildcard pattern matching — this is the engine behind `cube verify`
- **The canonical orientation is YB (yellow-up, blue-front) and is by far the most exercised path.** Non-YB orientations round-trip but their rotation semantics are not otherwise tested.

**Solver System (`internal/cube/solver.go`):**
- Interface-driven design: `type Solver interface { Solve(*Cube) (*SolverResult, error); Name() string }`
- Three registered solvers: BeginnerSolver, CFOPSolver, KociembaSolver
- **All three are empty stubs that return an empty solution for any unsolved cube.** This is the project's central gap.
- `internal/cube/solving_db.go` sketches a 4-look pattern-match solver but is **dead code** (unwired to any command)

**Main CLI Commands (`internal/cli/`):** (with honest status)
- `cube twist` - apply moves, display result (+ `--cfen`) — **works**
- `cube solve` - `--algorithm`/`--dimension` — **STUB: prints the scramble then an EMPTY solution**
- `cube verify <alg> --start <cfen> --target <cfen>` - CFEN verification — **works** (YB). Note: takes ONE positional arg (the algorithm), not two.
- `cube show` - display with cross/OLL/PLL/F2L highlighting — **works**
- `cube lookup` - algorithm database lookup — **works**
- `cube optimize` - move cancellation (`R R R`->`R'`) — **works**
- `cube find` - BFS search for sequences reaching a pattern — **works, but exponential (practical to ~6 moves; deeper times out)**
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
5. Execute `solver.Solve(cube)` → `SolverResult` (**currently returns an empty solution — solvers are stubs**)
6. Format output for CLI display

### Current Implementation Status

**✅ Completed Features:**
- Full NxN cube support (2x2 through 6x6+) with proper multi-layer moves
- Beautiful ASCII color output with `--color` flag (ANSI colored letters)
- Advanced move notation: M/E/S slices, Rw/Fw wide moves, 2R/3L layer moves, x/y/z rotations
- **Enhanced verification system** - `cube verify` command with flexible CFEN start/target support
- **CFEN infrastructure** - Complete parsing, generation, and wildcard matching
- **Pattern highlighting system** - `cube show` with cross/OLL/PLL/F2L highlighting
- **Algorithm database** - 63 algorithms defined; 5 carry verification patterns
- **Verified algorithm collection** - 5 algorithms with real CFEN patterns (Sune, Anti-Sune, Cross OLL, T-Perm, Sexy Move) — all pass `verify-database`
- **Clean architecture** - Separate database tools from main CLI
- **Database verification tools** - Standalone utilities for algorithm curation
- **Comprehensive test suite** - 98 end-to-end tests + Go unit tests, all green
- **Invariant guardrail suite** - load-bearing engine/solver/CFEN invariants (see "Invariants & Guardrails" near the top)
- Cross-platform build system (macOS/Linux compatible)

**⚠️ Current Issues / Known Gaps:**
- **All solvers are unimplemented stubs** — `cube solve` emits nothing. This is the headline gap.
- Algorithm database has only 5 verification patterns (of 63 entries)
- `internal/cube/solving_db.go` is a dead-code 4-look pattern-matcher — wire it up or delete it
- `internal/cube/cubie.go` is an unused piece-addressing stub reserved for future piece tracking
- CSV algorithm dumps ready for import in `/alg_dumps/` (9 files, 100+ algorithms)

**📍 Key Files to Know:**
- `TODO.md` - **ALWAYS READ FIRST** - Current development plan and progress
- `internal/cube/cube.go` - Core cube representation, color output methods
- `internal/cube/moves.go` - Move parsing and application logic
- `internal/cube/solver.go` - Solver implementations (currently unimplemented stubs)
- `internal/cube/algorithms.go` - Algorithm database (63 entries, 5 with CFEN patterns)
- `internal/cfen/` - Complete CFEN parsing, generation, and verification system
- `internal/cli/verify.go` - Enhanced verification command with CFEN support
- `internal/cli/solve.go` - CLI solve command with algorithm selection
- `internal/cli/show.go` - Cube display with pattern highlighting
- `internal/cli/lookup.go` - Algorithm database lookup command
- `tools/verify-algorithm/` - Standalone algorithm verification tool
- `tools/verify-database/` - Standalone database verification tool
- `tools/README.md` - Documentation for database tools
- `test/e2e_test.sh` - Comprehensive end-to-end test suite (98 tests)
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
The `solve` command is a **stub**. All three algorithms (`beginner`, `cfop`, `kociemba`)
return an empty solution for any unsolved cube, so the loop below currently prints three
empty `Solution:` lines — it does NOT demonstrate working solvers.
```bash
# All three currently emit an EMPTY solution (no solver implemented yet):
for algo in beginner cfop kociemba; do
    echo "=== $algo ==="
    ./dist/cube solve "R U R' U'" --algorithm $algo | grep "Solution:"
done
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
# Test advanced moves on different cube sizes (use twist; solve is a stub)
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
1. Add algorithm to `internal/cube/algorithms.go` with proper CFEN patterns
2. Use `./dist/tools/verify-algorithm` to test the algorithm
3. Use `./dist/tools/verify-database` to ensure database consistency
4. Update move count with `algorithm.UpdateMoveCount()`

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
3. Check if tests pass with `make test-all` (runs all 98 e2e tests)
4. Test database tools with `./dist/tools/verify-database`
5. Always run `make fmt && make vet` before committing

**Test Suite Coverage:**
- **Comprehensive end-to-end tests** covering every CLI command and feature (98 tests)
- **All cube dimensions** (2x2 through 20x20) with proper multi-layer moves
- **Advanced notation** (M/E/S slices, Rw/Fw wide moves, 2R/3L layer moves, x/y/z rotations)
- **Enhanced verification system** (CFEN patterns, wildcard matching, database verification)
- **Error handling and edge cases** with proper exit codes
- **Integration tests** (solve + verify workflows)
- **Performance tests** (large scrambles, complex cubes)
- **CFEN parsing and generation tests** (cube state conversion, pattern matching)
- **Pure bash implementation** - no Python dependencies, fully cross-platform