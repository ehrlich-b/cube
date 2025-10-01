# Cube - Rubik's Cube Solver

A comprehensive Rubik's cube solver written in Go supporting NxNxN cubes (2x2 through 10x10+), multiple solving algorithms, advanced move notation, and both CLI/web interfaces.

## ⚡ Quick Start

```bash
git clone https://github.com/ehrlich-b/cube
cd cube
make build

# Start by exploring moves and patterns
./dist/cube twist "R U R' U'" --color

# Then try solving scrambled cubes
./dist/cube solve "R U R' U'" --color
```

📖 **New to cubing?** Start with the [User Guide & Examples](./examples/) for step-by-step tutorials and spectacular demonstrations.

## 🔧 Core Features

- **NxNxN Cube Support**: 2x2x2, 3x3x3, 4x4x4, 5x5x5, and larger dimensions
- **Multiple Algorithms**: BeginnerSolver, CFOPSolver, KociembaSolver with distinct solutions
- **Advanced Notation**: Full WCA notation including M/E/S slices, Rw/Fw wide moves, 2R/3L layer moves, x/y/z rotations
- **Power User Tools**: Move optimization (`cube optimize`), algorithm discovery (`cube find`)
- **Solution Verification**: Built-in solution checking with `cube verify`
- **Pattern Recognition**: Algorithm database with lookup functionality
- **Dual Interfaces**: CLI tool and web terminal interface
- **Visual Output**: Unicode blocks and ANSI colored ASCII with clean unfolded cross layout

## 🚀 Command Overview

| Command | Purpose | Example |
|---------|---------|---------|
| **`twist`** | **Apply moves and see results** | **`cube twist "R U R' U'" --color`** |
| `solve` | Solve scrambled cubes | `cube solve "R U R' U'" --algorithm cfop --color` |
| `verify` | Check if solution works | `cube verify "R U" "U' R'" --verbose` |
| `show` | Display cube state with pattern highlighting | `cube show "R U R' U'" --highlight-oll --color` |
| `lookup` | Search algorithm database | `cube lookup sune --preview` |
| `optimize` | Minimize move sequences | `cube optimize "R R R"` → `R'` |
| `find` | Discover new algorithms | `cube find pattern solved --max-moves 4` |
| `serve` | Start web interface | `cube serve --port 8080` |

## 📖 Documentation

- **[User Guide & Examples](./examples/)** - Complete learning path from basics to advanced techniques
- **[CLAUDE.md](./CLAUDE.md)** - Development guidance and project instructions  
- **[TODO.md](./TODO.md)** - Current development status and roadmap

## 🔧 Installation & Development

```bash
# Basic setup
make install && make build

# Run comprehensive test suite (55 tests)
make test-all

# Code quality (always run before commits)
make fmt && make vet

# Start development server with hot reload
make dev
```

## ⚡ Quick Examples

```bash
# Basic solving with different algorithms
./dist/cube solve "R U R' U'" --algorithm beginner --color
./dist/cube solve "R U R' U'" --algorithm cfop --color 
./dist/cube solve "R U R' U'" --algorithm kociemba --color

# Advanced notation on larger cubes
./dist/cube solve "M E S" --dimension 3 --color        # Slice moves
./dist/cube solve "Rw Fw Uw" --dimension 4 --color     # Wide moves
./dist/cube solve "2R 3L 2F" --dimension 5 --color     # Layer moves

# Power user tools
./dist/cube optimize "R R R"                           # → R' (1 move)
./dist/cube find pattern solved --max-moves 4          # Algorithm discovery
./dist/cube verify "R U" "U' R'" --verbose            # Solution verification
```

**See [examples/](./examples/) for 400+ comprehensive examples, tutorials, and patterns.**

## 🔤 Move Notation

Full WCA (World Cube Association) standard notation support:

| Type | Syntax | Description | Cube Sizes |
|------|--------|-------------|------------|
| **Basic** | `R`, `U'`, `F2` | Standard face moves (F/B/R/L/U/D) | Any |
| **Slice** | `M`, `E'`, `S2` | Middle layer moves | Odd only (3x3, 5x5, 7x7...) |
| **Wide** | `Rw`, `Fw'`, `Uw2` | Multiple outer layers | 4x4+ |
| **Layer** | `2R`, `3L'`, `4U2` | Specific inner layers | 4x4+ |
| **Rotation** | `x`, `y'`, `z2` | Whole cube rotations | Any |

**Modifiers**: `'` (counter-clockwise), `2` (double turn)  
**Examples**: `R U R' U'` (sexy move), `M E S` (all slice moves), `Rw Uw Fw` (4x4 wide moves)

## Development

### Commands

```bash
# Build and run
make build          # Compile binary to dist/cube
make serve          # Start web server quickly
make dev            # Hot reload development (requires Air)

# Code quality
make test           # Run comprehensive test suite
make fmt            # Format code and clean whitespace
make vet            # Static analysis
make lint           # Lint with golangci-lint

# Dependencies
make install        # Download and tidy Go modules
make install-tools  # Install Air and golangci-lint

# Multi-platform builds
make build-all      # Build for Linux, macOS, Windows
```

### Development Workflow

```bash
# 1. Install development tools
make install-tools

# 2. Start hot-reload development
make dev

# 3. Run tests frequently
make test

# 4. Format and lint before commits
make fmt && make vet && make lint
```

## Testing

Comprehensive test suite covering:

```bash
# Run all tests
make test           # Unit tests
make e2e-test       # End-to-end CLI tests (109 tests)
make fuzz-solvers   # Fuzz test solvers with random scrambles
make test-all       # Everything

# Test specific functionality
go test ./internal/cube -v                    # Core cube logic
go test ./internal/cube -run TestMove         # Move system tests
go test ./internal/cube -run TestSolver       # Solver tests
go test ./internal/cube -bench=.              # Performance benchmarks
```

### Test Coverage

- ✅ Move parsing and application (all notation types)
- ✅ Cube state management (2x2 through 5x5+ cubes)
- ✅ Solver algorithms (Beginner 100%, Kociemba 100%, CFOP 95%)
- ✅ Fuzz testing: 59/60 pass on 1-3 move scrambles (98% overall success rate)
- ✅ Move sequence validation and inverses
- ✅ Edge cases (empty scrambles, invalid notation, identity moves)
- ✅ Performance benchmarks

## Architecture

### Core Components

```
cmd/cube/main.go                    # CLI entry point
├── internal/cli/                   # Cobra command definitions
│   ├── solve.go                   # Solve command
│   └── serve.go                   # Web server command
├── internal/cube/                  # Core cube logic
│   ├── cube.go                    # Cube representation
│   ├── moves.go                   # Move parsing and application
│   └── solver.go                  # Solving algorithms
└── internal/web/                   # Web interface
    ├── server.go                  # HTTP server
    └── handlers.go                # API endpoints
```

### Cube Representation

- **NxNxN Support**: `[6][][]Color` structure supports any cube dimension
- **Standard Colors**: White, Yellow, Red, Orange, Blue, Green faces
- **Efficient Storage**: Minimal memory footprint with direct array access
- **Fast Operations**: Optimized face rotations and edge movements

### Solving Algorithms

1. **BeginnerSolver**: Layer-by-layer method using A* search
   - **Status**: ✅ Works reliably on 1-3 move scrambles
   - **Performance**: <10s for most simple scrambles
   - **Fuzz Test**: 20/20 pass (100%)

2. **KociembaSolver**: Two-phase algorithm (3x3 only)
   - **Status**: ✅ Works reliably, slower than Beginner
   - **Performance**: 1-3 moves <10s, 6 moves ~53s
   - **Fuzz Test**: 20/20 pass (100%)

3. **CFOPSolver**: Cross, F2L, OLL, PLL speedcubing method (Hybrid)
   - **Status**: ✅ Works with Beginner fallback
   - **Performance**: <15s for 1-3 move scrambles
   - **Fuzz Test**: 19/20 pass (95%) - falls back to Beginner on failure
   - **Note**: Nearly as reliable as pure Beginner/Kociemba

## API Examples

### Current Programmatic Usage

```go
package main

import (
    "fmt"
    "github.com/ehrlich-b/cube/internal/cube"
)

func main() {
    // Create and manipulate cube
    c := cube.NewCube(3)
    moves, _ := cube.ParseScramble("R U R' U'")
    c.ApplyMoves(moves)

    // Display cube state
    fmt.Println(c.String())  // ASCII representation
    fmt.Println(c.ColorString())  // With colors

    // Verify algorithms
    alg := cube.GetAlgorithm("Sune")
    verified := cube.VerifyAlgorithm(alg)
    fmt.Printf("Algorithm verified: %v\n", verified)
}
```

## ⚙️ Technical Details

### Current Implementation Status

- ✅ **NxNxN cubes**: Support for 2x2 through 10x10+ with proper layer handling
- ✅ **Working solvers**:
  - **BeginnerSolver**: 100% reliable (20/20 fuzz tests), <10s solve time
  - **KociembaSolver**: 100% reliable (20/20 fuzz tests), <10s on 1-3 moves, 53s on 6 moves
  - **CFOPSolver**: 95% reliable (19/20 fuzz tests), hybrid with Beginner fallback
- ✅ **Advanced notation**: M/E/S slices, Rw/Fw wide moves, 2R/3L layer moves, x/y/z rotations
- ✅ **Algorithm database**: 140 algorithms across all categories with pattern generation
- ✅ **Power user tools**: Move optimization and algorithm discovery via BFS
- ✅ **Comprehensive testing**: 109 e2e tests + fuzz testing (40/40 pass on 1-3 move scrambles)

### 🚧 Future Enhancements

- **Optimal solving**: Implement true optimal solvers for each algorithm type
- **Interactive mode**: Terminal-based live cube manipulation interface  
- **3D visualization**: ASCII 3D cube rendering and step-by-step solving
- **Custom patterns**: User-defined pattern recognition and generation
- **Performance profiling**: Detailed timing analysis and optimization metrics

## Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Make changes and add tests
4. Format and test (`make fmt && make test`)
5. Commit changes (`git commit -m 'Add amazing feature'`)
6. Push to branch (`git push origin feature/amazing-feature`)
7. Open Pull Request

### Development Guidelines

- Write tests for all new functionality
- Follow existing code style and conventions
- Run `make fmt` before committing
- Ensure all tests pass with `make test`
- Add examples to README for new features

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Acknowledgments

- **Singmaster Notation**: Standard cube move notation system
- **Kociemba Algorithm**: Herbert Kociemba's two-phase solving method
- **CFOP Method**: Cross, F2L, OLL, PLL speedcubing approach
- **Go Community**: Excellent tooling and library ecosystem
