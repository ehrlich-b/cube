# Cube — a CLI Rubik's Cube Toolkit

A Go command-line toolkit for Rubik's cubes: a correct NxNxN move engine, full WCA
move notation, a CFEN state/pattern language with verification, move optimization,
and a breadth-first algorithm search.

> **Honest status:** the move engine, verification, optimization, and search all work
> and are covered by tests. **A complete 3×3 beginner solver and lesson now work.**
> See [the beginner guide](./examples/beginner.md). CFOP and Kociemba remain unimplemented.

## Status

**Works today**
- `cube learn` — a complete layer-by-layer 3×3 solve, with checkpoints, orientation
  guidance and interactive recovery (`next`, `moves`, `undo`, `reset`, saved CFEN)
- `cube solve` — complete 3×3 beginner solution, including headless moves
- `--goal first-layer` retains the explicit white-layer sublesson
- Physical 3×3 state validation: color counts, centers, cubie identity, flips, twists, parity
- NxNxN move engine (2x2 through large N), all WCA notation, whole-cube rotations
- `cube twist` — apply moves, render the cube (ASCII / colored / Unicode)
- `cube verify` — check an algorithm against CFEN start/target states (wildcards supported)
- `cube optimize` — cancel/merge moves (`R R R` → `R'`)
- `cube find` — BFS search for move sequences that reach a target pattern
- `cube lookup`, `cube show`, the CFEN utility commands, and the `verify-*` database tools

**Not implemented yet**
- **CFOP and Kociemba remain empty API stubs**; the CLI rejects these unavailable algorithms.
- Beginner solving supports 3×3 only; other dimensions are rejected, not reported as solved.
- Algorithm database has only 5 entries with verification patterns (of 63 defined)
- `cube find` is correct but exponential — practical only to ~6 moves; it is not a general scramble solver

## Quick Start

```bash
git clone https://github.com/ehrlich-b/cube
cd cube
make build            # builds dist/cube
make build-tools      # builds dist/tools/verify-algorithm and verify-database

# Apply moves and see the result
./dist/cube twist "R U R' U'" --color

# Finish a scrambled 3x3, one beginner checkpoint at a time
./dist/cube learn "R U F2 L' B" --interactive --color

# Return moves that solve all six faces
./dist/cube solve "R U F2 L' B" --headless

# Return moves that solve the first layer (other layers may remain scrambled)
./dist/cube solve "R U F2 L' B" --goal first-layer --headless

# Verify an algorithm against CFEN states (sexy move x6 = identity → solved)
./dist/cube verify "R U R' U' R U R' U' R U R' U' R U R' U' R U R' U' R U R' U'" \
    --start "YB|Y9/R9/B9/W9/O9/G9" --target "YB|Y9/R9/B9/W9/O9/G9"

# Optimize a sequence
./dist/cube optimize "R R R"            # → R'

# Search for a short sequence that reaches a pattern
./dist/cube find sequence "R U"         # → U' R'
```

See [examples/](./examples/) for tutorials and pattern walkthroughs.

## Command Overview

| Command | Purpose | Status |
|---------|---------|--------|
| `twist` | Apply moves and render the cube | works |
| `verify` | Check an algorithm vs. CFEN start/target (`--start`/`--target`) | works |
| `show` | Render a cube with cross/OLL/PLL/F2L highlighting | works |
| `lookup` | Search the algorithm database | works |
| `optimize` | Cancel/merge a move sequence | works |
| `find` | BFS search for sequences reaching a pattern | works (exponential) |
| `parse-cfen` / `generate-cfen` / `verify-cfen` / `match-cfen` | CFEN utilities | works |
| `identify` / `show-alg` | Pattern identify / algorithm display | partial |
| `learn` | Teach a complete beginner solve, with recovery and checkpoints | works (3×3) |
| `solve --goal first-layer` | Solve the white first layer | works (3×3 beginner) |
| `solve` / `solve --goal full` | Solve all three layers with beginner method | works (3×3) |

Note: `verify` takes a single positional argument — the algorithm — plus `--start`/`--target` flags.

## Move Notation

Full WCA (World Cube Association) notation:

| Type | Syntax | Description | Cube Sizes |
|------|--------|-------------|------------|
| Basic | `R`, `U'`, `F2` | Face moves (F/B/R/L/U/D) | Any |
| Slice | `M`, `E'`, `S2` | Middle-layer moves | Odd only (3x3, 5x5, …) |
| Wide | `Rw`, `Fw'`, `Uw2` | Multiple outer layers | 3x3+ |
| Layer | `2R`, `3L'`, `4U2` | Specific inner layers | 4x4+ |
| Rotation | `x`, `y'`, `z2` | Whole-cube rotations | Any |

Modifiers: `'` (counter-clockwise), `2` (double turn).

## Cube Orientation (canonical)

Yellow up, White down, Blue front, Green back, Orange left, Red right. The default CFEN
orientation is `YB` (yellow-up, blue-front). Apply `x`/`y`/`z` rotations before a sequence
to work from a different orientation.

## Testing & Invariants

```bash
make test        # Go unit tests, including the invariant suite
make e2e-test    # 122 end-to-end CLI tests
make test-all    # both
make test-first-layer  # independent partial-goal regression oracle (Python 3)
make test-beginner     # independently replay full solutions and last-layer recovery
make fmt && make vet   # before committing
```

The **invariant suite** is the project's safety net — it must stay green:

- `internal/cube/invariants_test.go` — sticker conservation, scramble+inverse = solved,
  determinism, and the **solver contract** (any non-empty solution must actually solve the cube;
  stubs SKIP rather than fake a pass).
- `internal/cfen/cfen_test.go` — canonical solved CFEN, cube↔CFEN round-trip (all orientations),
  wildcard matching, verify semantics.
- `internal/cli/commands_test.go` — no command registered twice.

## Architecture

```
cmd/cube/main.go                 # CLI entry point
internal/cli/                    # Cobra commands (twist, verify, solve, find, optimize, ...)
internal/cube/                   # Core engine
  cube.go                        # NxNxN representation, IsSolved, rendering
  moves.go / move_parser.go      # move parsing + application
  ring_generators.go / permutations.go  # the permutation engine
  algorithms.go                  # algorithm database
  solver.go                      # full beginner solver + CFOP/Kociemba stubs
  first_layer.go / full_lesson.go / lesson_session.go  # beginner checkpoints and recovery
  state_validation.go            # physical 3x3 legality checks
  solving_db.go                  # experimental 4-look pattern matcher (currently unwired)
  cubie.go                       # piece-addressing scaffold for future piece tracking (unused)
internal/cfen/                   # CFEN parsing, generation, conversion, matching
tools/                           # verify-algorithm, verify-database, generate-patterns
```

## Programmatic Usage

```go
package main

import (
	"fmt"

	"github.com/ehrlich-b/cube/internal/cfen"
	"github.com/ehrlich-b/cube/internal/cube"
)

func main() {
	c := cube.NewCube(3)
	moves, _ := cube.ParseMoves("R U R' U'")
	c.ApplyMoves(moves)

	fmt.Println(c.String())               // ASCII unfolded layout
	fmt.Println(c.StringWithColor(true))  // colored
	fmt.Println("solved:", c.IsSolved())

	cfenStr, _ := cfen.GenerateCFEN(c)    // YB|... state string
	fmt.Println(cfenStr)

	for _, alg := range cube.LookupAlgorithm("Sune") {
		fmt.Printf("%s (%s): %s\n", alg.Name, alg.CaseID, alg.Moves)
	}
}
```

## Roadmap

The complete beginner path now finishes all six faces. The unchanged full-solver
contract actively checks it. Partial first-layer results retain a separate result
type and explicit goal. Next improvements should follow a human physical-cube
trial; CFOP, Kociemba and larger-cube solving remain future work.
See [TODO.md](./TODO.md) and [docs/solvers.md](./docs/solvers.md).

## License

MIT License — see [LICENSE](LICENSE).
