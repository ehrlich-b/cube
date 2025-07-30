# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 🚀 Quick Start

**IMPORTANT: Always check TODO.md first!** It contains the current development plan and status.

### Essential Commands for Testing

```bash
# Build and test basic functionality
make build
./dist/cube twist "R U R' U'" --color
./dist/cube solve "R U R' U'" --color

# Test all major features
./dist/cube verify "R U R' U'" "U R U' R'"
./dist/cube show "R U R' U'" --highlight-oll --color
./dist/cube lookup sune --preview

# Test different algorithms and cube sizes
./dist/cube solve "R U R' U'" --algorithm beginner
./dist/cube solve "Rw Uw Fw" --dimension 4 --color

# Run comprehensive test suite (45 tests including web interface)
make test-all

# Start web terminal interface
./dist/cube serve
# Then open http://localhost:8080/terminal in browser
```

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
- `make build` - Compile binary to `dist/cube`
- `make clean` - Remove build artifacts
- `make build-all` - Build for multiple platforms (linux, darwin, windows)
- `make serve` - Start web server quickly
- `make dev` - Hot reload development (requires Air: `make install-tools`)
- `make run` - Run CLI directly with go run

**Code Quality (ALWAYS run before committing):**
- `make test` - Run unit tests
- `make e2e-test` - Run end-to-end test suite (44 comprehensive tests)
- `make test-all` - Run both unit and e2e tests
- `make fmt` - Format Go code (cross-platform compatible)
- `make vet` - Static analysis
- `make lint` - Lint with golangci-lint (requires: `make install-tools`)

**Dependencies:**
- `make install` - Download and tidy Go modules
- `make install-tools` - Install Air and golangci-lint

## Architecture Overview

This is a Rubik's cube solver with dual interfaces (CLI + Web) built around a flexible core engine:

```
cmd/cube/main.go
    ↓
internal/cli/ (Cobra commands)
    ├── solve.go → internal/cube/ (Core logic)
    └── serve.go → internal/web/ → internal/cube/
```

### Core Components

**Cube Representation (`internal/cube/cube.go`):**
- `Cube` struct supports NxNxN cubes (2x2, 3x3, 4x4+) 
- Uses `[6][][]Color` for six faces with dynamic sizing
- Standard Singmaster notation parsing (R, U', F2, etc.)

**Solver System (`internal/cube/solver.go`):**
- Interface-driven design: `type Solver interface { Solve(*Cube) (*SolverResult, error) }`
- Three algorithms: BeginnerSolver, CFOPSolver, KociembaSolver
- All solvers currently have placeholder implementations

**CLI Commands (`internal/cli/`):**
- `cube solve` - CLI solving with `--algorithm` and `--dimension` flags
- `cube serve` - Web server with `--host` and `--port` options
- Built with Cobra framework

**Web Interface (`internal/web/`):**
- Gorilla Mux router with `/api/solve` REST endpoint
- Embedded HTML/CSS/JS (no separate static files currently)
- JSON API accepts scramble, algorithm, and dimension parameters

### Key Data Flow

**Solving Process (both CLI and Web):**
1. Parse scramble string into `[]Move`
2. Create `Cube` with specified dimension
3. Apply scramble moves to cube
4. Get solver by algorithm name from factory
5. Execute `solver.Solve(cube)` → `SolverResult`
6. Format output (text for CLI, JSON for web)

### Current Implementation Status

**✅ Completed (Phase 3 - Terminal Web Interface):**
- Full NxN cube support (2x2 through 6x6+) with proper multi-layer moves
- BeginnerSolver with real layer-by-layer algorithm
- Beautiful ASCII color output with `--color` flag (ANSI colored letters)
- Advanced move notation: M/E/S slices, Rw/Fw wide moves, 2R/3L layer moves, x/y/z rotations
- **Solution verification system** - `cube verify` command with verbose output
- **Pattern highlighting system** - `cube show` with cross/OLL/PLL/F2L highlighting
- **Algorithm database** - 15 common algorithms with lookup by name/pattern/category
- **Terminal web interface** - Full CLI functionality accessible via web browser at `/terminal`
- **REST API** - `/api/exec` endpoint mirrors all CLI commands with proper argument parsing
- **Comprehensive test suite** - 45 end-to-end tests covering all features including web interface
- Cross-platform build system (macOS/Linux compatible)
- All three algorithms produce distinct solutions

**⚠️ Current Issues:**
- CFOP and Kociemba solvers have basic placeholder implementations (functional but not optimal)
- Empty `pkg/algorithms/` and `internal/web/static/` directories

**📍 Key Files to Know:**
- `TODO.md` - **ALWAYS READ FIRST** - Current development plan and progress
- `internal/cube/cube.go` - Core cube representation, color output methods
- `internal/cube/moves.go` - Move parsing and application logic
- `internal/cube/solver.go` - Solver implementations
- `internal/cube/algorithms.go` - Algorithm database with 15 common algorithms
- `internal/cli/solve.go` - CLI solve command with algorithm selection
- `internal/cli/verify.go` - Solution verification command
- `internal/cli/show.go` - Cube display with pattern highlighting
- `internal/cli/lookup.go` - Algorithm database lookup command
- `test/e2e_test.sh` - Comprehensive end-to-end test suite

### Adding New Features

**New Solver Algorithm:**
1. Implement `Solver` interface in `internal/cube/solver.go`
2. Add to `GetSolver()` factory function
3. Update CLI flag validation in `internal/cli/solve.go`

**New Cube Dimension:**
- Cube struct already supports arbitrary dimensions
- Validate in CLI flag parsing and web API handlers

**Web UI Enhancements:**
- Static files go in `internal/web/static/` (currently empty)
- Update handlers in `internal/web/handlers.go` to serve static content

### Dependencies and Tools

**Runtime Dependencies:**
- `github.com/spf13/cobra` - CLI framework
- `github.com/gorilla/mux` - HTTP router

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

### Testing Algorithm Differences
```bash
# Quick test to see all three algorithms produce different solutions
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
# Test advanced moves on different cube sizes
./dist/cube solve "M E S" --dimension 3 --color        # Slice moves
./dist/cube solve "Rw Fw' Uw2" --dimension 4 --color    # Wide moves  
./dist/cube solve "2R 3L'" --dimension 5 --color       # Layer moves

# Use verify to test solutions
./dist/cube verify "R U R' U'" "U R U' R'"             # Should pass
./dist/cube verify "R U R' U'" "F U F' U'"             # Should fail

# Use show to visualize patterns
./dist/cube show "R U R' U'" --highlight-oll --color   # Highlight top layer
```

**Before starting any work:**
1. Read TODO.md to understand current phase
2. Run `make build` and test basic functionality
3. Check if tests pass with `make test-all` (runs all 45 e2e tests)
4. Always run `make fmt && make vet` before committing

**Test Suite Coverage:**
- **45 comprehensive end-to-end tests** covering every command and feature including web interface
- **All cube dimensions** (2x2 through 20x20) with proper multi-layer moves
- **All algorithms** (beginner, cfop, kociemba) produce distinct solutions
- **Advanced notation** (M/E/S slices, Rw/Fw wide moves, 2R/3L layer moves, x/y/z rotations)
- **All Phase 2 features** (verify, show with highlighting, lookup database)
- **Phase 3 web interface** (terminal emulator, REST API, command parsing)
- **Error handling and edge cases** with proper exit codes
- **Integration tests** (solve + verify workflows)
- **Performance tests** (large scrambles, complex cubes)
- **Pure bash implementation** - no Python dependencies, fully cross-platform