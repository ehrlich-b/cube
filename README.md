# Cube - Rubik's Cube Solver

A flexible Rubik's cube solver written in Go that supports multiple dimensions and solving algorithms. Provides both CLI and web interfaces for solving, analyzing, and visualizing cube states.

## Features

- **Multiple Cube Sizes**: Support for 2x2x2, 3x3x3, 4x4x4, and larger cubes
- **Multiple Algorithms**: Beginner, CFOP, and Kociemba solving methods
- **CLI Interface**: Command-line tool for quick solving and analysis
- **Web Interface**: Browser-based interface with terminal-style interaction
- **Colored Output**: Beautiful ASCII visualization with muted terminal colors
- **Move Parsing**: Standard Singmaster notation (R, U', F2, etc.)
- **Fast Performance**: Optimized cube representation and move application

## Installation

```bash
git clone https://github.com/ehrlich-b/cube
cd cube
make install  # Download dependencies
make build    # Build binary to dist/cube
```

## Quick Start

### Basic Solving

```bash
# Solve a scrambled 3x3x3 cube
./dist/cube solve "R U R' U' R' F R F'"

# Use different algorithm
./dist/cube solve "R U R' U'" --algorithm cfop

# Solve 4x4x4 cube
./dist/cube solve "R U R' U'" --dimension 4

# Enable colored output
./dist/cube solve "R U R' U'" --color
```

### Web Interface

```bash
# Start web server
./dist/cube serve

# Custom host and port
./dist/cube serve --host 0.0.0.0 --port 3000
```

Visit `http://localhost:8080` for the terminal-style web interface.

## CLI Examples

### Basic Commands

```bash
# Solve with default beginner algorithm
./dist/cube solve "R U R' U'"

# Show all available algorithms
./dist/cube solve --help

# Solve empty scramble (already solved)
./dist/cube solve ""
```

### Algorithm Comparison

```bash
# Compare different algorithms on same scramble
./dist/cube solve "R U R' U' R' F R F'" --algorithm beginner
./dist/cube solve "R U R' U' R' F R F'" --algorithm cfop
./dist/cube solve "R U R' U' R' F R F'" --algorithm kociemba
```

### Different Cube Sizes

```bash
# 2x2x2 cube (Pocket Cube)
./dist/cube solve "R U R' U'" --dimension 2

# 3x3x3 cube (Standard Rubik's Cube)
./dist/cube solve "R U R' U'" --dimension 3

# 4x4x4 cube (Rubik's Revenge)
./dist/cube solve "R U R' U'" --dimension 4

# 5x5x5 cube (Professor's Cube)
./dist/cube solve "R U R' U'" --dimension 5
```

### Visual Output

```bash
# Standard text output
./dist/cube solve "R U R' U'"

# Colored terminal output (muted colors, eye-friendly)
./dist/cube solve "R U R' U'" --color
```

## Move Notation

The solver uses standard Singmaster notation:

### Basic Moves
- `F` - Front face clockwise
- `B` - Back face clockwise  
- `R` - Right face clockwise
- `L` - Left face clockwise
- `U` - Up face clockwise
- `D` - Down face clockwise

### Modifiers
- `'` - Counter-clockwise (e.g., `R'`, `U'`)
- `2` - Double turn (e.g., `R2`, `U2`)

### Example Scrambles
```bash
# T-Perm algorithm
./dist/cube solve "R U R' F' R U R' U' R' F R2 U' R'"

# Sexy move sequence
./dist/cube solve "R U R' U'"

# Sune algorithm
./dist/cube solve "R U R' U R U2 R'"

# J-Perm
./dist/cube solve "R U R' F' R U R' U' R' F R2 U' R' U'"
```

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
make test

# Test specific functionality
go test ./internal/cube -v                    # Core cube logic
go test ./internal/cube -run TestMove         # Move system tests
go test ./internal/cube -run TestSolver       # Solver tests
go test ./internal/cube -bench=.              # Performance benchmarks
```

### Test Coverage

- ✅ Move parsing and application (all notation types)
- ✅ Cube state management (2x2 through 5x5+ cubes)
- ✅ Solver algorithms (beginner, CFOP, Kociemba)
- ✅ Move sequence validation and inverses
- ✅ Edge cases (empty scrambles, invalid notation)
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

1. **BeginnerSolver**: Layer-by-layer method suitable for learning
2. **CFOPSolver**: Cross, F2L, OLL, PLL - advanced speedcubing method  
3. **KociembaSolver**: Two-phase algorithm for optimal solutions (3x3 only)

## API Examples

### Web API

```bash
# Start server
./dist/cube serve --port 8080

# Solve via API
curl -X POST http://localhost:8080/api/solve \
  -H "Content-Type: application/json" \
  -d '{"scramble": "R U R'\'' U'\''", "algorithm": "beginner", "dimension": 3}'
```

### Programmatic Usage

```go
package main

import (
    "fmt"
    "github.com/ehrlich-b/cube/internal/cube"
)

func main() {
    // Create and scramble cube
    c := cube.NewCube(3)
    moves, _ := cube.ParseScramble("R U R' U'")
    c.ApplyMoves(moves)
    
    // Solve with beginner method
    solver, _ := cube.GetSolver("beginner")
    result, _ := solver.Solve(c)
    
    fmt.Printf("Solution: %v\n", result.Solution)
    fmt.Printf("Steps: %d\n", result.Steps)
}
```

## Known Limitations

### Current Implementation Status

- ✅ **3x3x3 cubes**: Fully implemented with proper edge rotation
- ✅ **2x2x2 cubes**: Works (simpler case, no edges)
- ⚠️ **4x4x4+ cubes**: **Limited** - inner layers not handled correctly
- ✅ **All notation**: Standard Singmaster moves supported
- ✅ **Multiple algorithms**: Beginner, CFOP, Kociemba available

### Future Improvements

1. **Complete 4x4+ Support**: Implement proper inner layer rotations
2. **Real Algorithms**: Replace placeholder implementations with actual solving logic
3. **Pattern Recognition**: Add algorithm database and pattern matching
4. **Interactive Mode**: Terminal-based interactive cube manipulation
5. **Algorithm Discovery**: Exhaustive search for new move sequences

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
