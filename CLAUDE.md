# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

**Build and Run:**
- `make build` - Compile binary to `dist/cube`
- `make serve` - Start web server quickly
- `make dev` - Hot reload development (requires Air: `make install-tools`)
- `./dist/cube solve "R U R' U'"` - CLI solve with scramble
- `./dist/cube serve --host localhost --port 8080` - Web server with custom host/port

**Code Quality:**
- `make test` - Run tests (currently no tests exist)
- `make fmt` - Format Go code
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

**Completed:**
- Basic cube data structures and move parsing
- CLI command structure with proper flag handling  
- Web server with API endpoints and basic UI
- Cross-platform build system

**Incomplete (TODOs in code):**
- All three solver algorithms are placeholder implementations
- Cube move application has incomplete edge rotation logic (`internal/cube/moves.go`)
- No test coverage
- Empty `pkg/algorithms/` and `internal/web/static/` directories

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