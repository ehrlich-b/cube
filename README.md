# Cube — a CLI Rubik's Cube Toolkit

A Go command-line toolkit for Rubik's cubes: a correct NxNxN move engine, full WCA
move notation, a CFEN state/pattern language with verification, move optimization,
and shortest-sequence pattern search.

> **Honest status:** the move engine, verification, optimization, and search all work
> and are covered by tests. **Kociemba solves a complete 3×3 by default.** The
> complete beginner solver and [lesson](./examples/beginner.md) remain available.
> CFOP solves a complete 3×3 with named stage playback. **Dimensions 2–7 now solve
> completely:** 4–7 use center/edge reduction with parity correction, then Kociemba.

## Status

**Works today**
- `cube learn` — a complete layer-by-layer 3×3 solve, with checkpoints, orientation
  guidance and interactive recovery (`next`, `moves`, `undo`, `reset`, saved CFEN)
- `cube solve` — fast complete 3×3 Kociemba solution; `--method beginner` retains
  the beginner solver; `--method cfop` solves Cross, F2L, OLL and PLL with case names.
  `--algorithm` remains a supported alias
- `cube solve --dimension N` — complete 2×2 through 7×7 solutions; 2×2 uses the
  3×3 corners path, 4–7 reduce centers and edges, including both even-cube parities
- `cube solve --optimal --time-limit 1s` — shortest face-turn solution if proved
  before the deadline; an explicit error if the search times out
- `--goal first-layer` retains the explicit white-layer sublesson
- Physical 3×3 state validation: color counts, centers, cubie identity, flips, twists, parity
- NxNxN move engine (2x2 through large N), all WCA notation, whole-cube rotations
- `cube twist` — apply moves, render the cube (ASCII / colored / Unicode)
- `cube verify` — check an algorithm against CFEN start/target states (wildcards supported)
- `cube optimize` — cancel/merge moves (`R R R` → `R'`)
- `cube find` — IDA* with cubie coordinates and admissible pruning tables for
  exact/wildcard 3×3 targets; sticker BFS for other dimensions or move alphabets
- CFEN search targets and concrete start states, restricted `--moves`, and
  checkerboard/cross patterns; returns one shortest sequence
- `cube lookup`, `cube show`, the CFEN utility commands, and the `verify-*` database tools

**Not implemented yet**
- CFOP color neutrality, extended cross and F2L look-ahead remain future work.
- Full solving above dimension 7 is unsupported. Big-cube solutions prioritize
  correctness and currently use hundreds or thousands of slice turns.
- The dump contains incorrect descriptions and missing cases; quarantined rows need
  curation, and some OLL/PLL cases use multiple verified database algorithms.
- Optimal search is exponential on deep states; wildcard heuristics are weaker
  than exact-state heuristics. Use Kociemba for general scramble solving.

Native measurements use **1,000 uniform legal states in one process**, seed
`2026100709`, one search thread, Go 1.26.2 and Mac background QoS
(`taskpolicy -b nice -n 15`). Every answer is replayed against the original
stickers and the input is checked for mutation. `make test-kociemba` gates
**100% at ≤20 turns, warm mean <50 ms and p99 <250 ms**, runs a first-ever plain
solve in a fresh process with an empty cache (**≤1 second and ≤10,000,000 cache
bytes**), and independently constructs/replays 200 Python physical states,
including all 24 grips. `make bench-kociemba` runs the warm sample alone;
`make test-kociemba-cold` runs the cold gate alone.

The compact default measured **19.889 mean / 20 max turns, 100% at ≤20**, with
**12.29 ms warm mean / 106.08 ms p99 / 475.02 ms max**. The first-ever plain solve
of the cold-gate scramble took **214.45 ms**, including setup, and wrote
**3,337,787 bytes (3.338 MB)**. A fresh cached Go subprocess including a uniform
solve took **156.81 ms**. The independent 200-state CLI oracle measured **19.935
mean / 20 max turns**, **192.92 ms process mean / 821.07 ms max**, including
cached table setup in each process. These are seeded measurements under the
policy above, rather than worst-case bounds.

Default tables quotient phase-one twist/slice and slice/flip coordinates under
**16 U/D-axis symmetries (eight rotations and their mirrors)**, giving **168
twist classes and 45 slice classes**. Twist/flip uses the eight symmetries that
preserve flip independently of slice
(**324 twist classes**), with four-bit distances in two frames. An exact
**7,646-entry symmetry-reduced radius-five frontier** quickly rejects shallow
dead ends; tests check every descending witness and closure of all layers below
the frontier. Phase two has **2,768 corner and edge permutation classes**,
paired with slice permutation and the other group's four-piece combination
plus parity; inverse bounds preserve the existing correlation pruning.
Three axes, inverse views and up to three pre-moves share a resumable total-length
budget. A turn and its inverse share phase-one work; both phase-two endings are
checked. Reductions advance together by phase-one plus pre-move length.

The engine-owned **3,337,787-byte** compressed compact asset ships inside the
native and WASM executables. First use copies it to
`cube/coordinates-v5.bin.gz` in the user's cache; `CUBE_CACHE_DIR` overrides the
directory. A missing or corrupt cache uses the embedded asset without generating
large tables. `make generate-tables` rebuilds the asset from this engine's moves and exhaustively compares every
symmetry-reduced pair distance with unquotiented BFS. Checksums, dimensions and
move fingerprints guard the cache. The committed generated file is under 5 MB.

The previous large-table solver measured **19.932 mean / 20 max turns, 100% at
≤20**, **5.68 ms warm mean / 57.22 ms p99 / 187.94 ms max**, with a **625.88 ms
fresh cached Go subprocess**. Its CLI oracle measured **19.935 mean / 20 max
turns**, **558.76 ms process mean / 890.13 ms max**. Its phase-one table had
**295,004,160 states / 147,502,080 packed bytes**; generation measured **120.99 s**
under background QoS with roughly **1.6 GiB** of temporary memory. A first-ever
plain solve regressed to about **37 s on performance cores / 161 MB of cache**.
That table is now optional: **`cube tables build --large`** explicitly builds it,
and **`CUBE_LARGE_TABLES=1 cube solve ...`** explicitly selects it. Ordinary solves
ignore it even if it already exists. `make test-phase1-tables` checks the optional
table's admissibility, consistency and descending witnesses (maximum distance 12).
For comparison, c862379 with CPU profiling measured **19.801 mean / 21 max turns,
98.8% at ≤20**, **88.80 ms mean / 1000.08 ms p99 / 1000.52 ms max**.

The pruning/search strategy is informed by
[min2phase](https://github.com/cs0x7f/min2phase). A paired private reference run
used [min2phaseCXX](https://github.com/lilborgo/min2phaseCXX), built with installed
Clang (`-std=c++14 -O3`) inside ignored `.scratch/`, on the identical 1,000 fixtures
and process policy. It returned **19.746 mean / 20 max turns, 100% at ≤20**, with
**5.28 ms warm mean / 53.22 ms p99 / 316.64 ms max**, **991,712 cache bytes**,
**156.14 ms generation** and **24.73 ms fresh cached process**. Every answer
passed independent Python physical replay. Reference code and outputs are not
committed. `make export-reference-fixtures` writes the same 1,000 two-phase and
ten optimal URFDLB inputs into `.scratch/` without fetching code or dependencies.
Cube Explorer and nissy/vcube paired timings remain outstanding.

Optimal search retains the small databases for states within ten turns. Deep
`--optimal` states lazily build full-corner and two disjoint six-edge databases:
**88,179,840 corner states** and **42,577,920 states per edge group**, packed into
**44,089,920 + 21,288,960 + 21,288,960 bytes**. Factorized six-edge transitions
use **47,900,160 bytes**. Their checksummed `optimal-v1.bin` is
**134,568,064 bytes (128.33 MiB)**; earlier generation took **67.53 seconds**.

The stronger heuristic follows [Kociemba’s huge optimal coordinate](https://kociemba.org/math/optimal.htm):
it fixes the four slice edges in place as well as orienting all cubies. Sixteen U/D-axis symmetries reduce corner twists to **168 classes**,
with **4,087,480,320 combined entries**. The [modulo-three encoding](https://kociemba.org/math/pruning.htm) uses two bits
for exact distances through ten turns; remaining entries give the admissible lower bound
**eleven**. Saturation retains consistency across moves. The parent's bound
and a small lookup recover each child's bound, avoiding division in the search.
The cache is **1,021,870,144 bytes (974.53 MiB)**, including its checksum and move
fingerprint; the two optimal caches total **1,156,438,208 bytes**. Metadata and
moves are reconstructed from native coordinates and occupy about 1.2 MB.
Generation took **23m08.27s**, with **2,822,651,904 bytes peak RSS** and
**5,500,473,664 bytes peak memory footprint**, measured directly on the test
executable under Mac background QoS. A fresh cached setup took **1.75 s**;
a profiled warm search peaked at **1.20 GB RSS**.

Search takes the maximum of the corner, six-edge and three oriented sorted
phase-one bounds. Three equal positive phase-one bounds imply one additional
turn. Forward and inverse root bounds select the search direction. At depth
thirteen and above, two workers partition the eighteen root moves; both must
finish to disprove a depth. Tables are shared and `GOMAXPROCS=2` bounds CPU work.
Compared with the previous phase-one heuristic, completed depth-fifteen searches
on the first two uniform fixtures visited **126,450 / 141,599 nodes**, versus
**957,974 / 1,077,891**: about **7.6 times fewer**. The two-worker measurements
were **0.306 / 0.432 s**, versus **2.78 / 3.78 s** for the smaller heuristic in
one worker. An eight-edge pattern-table experiment reduced nodes less and
increased elapsed time, so it is not part of the solver.

Large generation is opt-in through deep `--optimal` calls, with a one-time
stderr message. Ordinary solves retain the compact default tables and cache.
Optimal setup is included in `--time-limit`; interrupted builds never publish
partial tables. `make test-optimal` checks short distances against the separate
IDA* finder, independently proves three fixed distances by physical-sticker
BFS, checks deeper finder fixtures, and replays superflip in every grip.

The uniform-state benchmark (seed `2026100714`, **twenty states**, two search
workers, background QoS, no CPU profiling, **180-second limit per state**)
proved **9/20** and timed out on **11/20**. Proven solve times, in increasing
order, were **13.42, 13.74, 18.05, 19.04, 57.60, 60.67, 71.41, 76.94 and
176.67 seconds**. Eight answers had optimal length **17**, and one had length
**18**; every answer passed full-cube replay and input-immutability checks.
**Five of twenty finished within 60 seconds. The full-sample median exceeds
180 seconds**, because eleven observations are right-censored above that limit.
No solved-only median is presented as the uniform-state median. On the original
first ten fixtures, **5/10** now finish at the same budget, compared with the
previous **0/10** profiled result. Warm benchmark RSS peaked at
**1,193,705,472 bytes (1.19 GB)**, and cached setup took **2.01 seconds**.

**The practical optimal target is not met.** The median-under-60-second target
fails in this sample; completion within ten minutes for all twenty states was
not established by the three-minute trials. `make bench-optimal` reproduces
the twenty fixtures and budgets. Stronger inverse-state heuristics or further
search improvements remain necessary. The rejected eight-edge experiment and
all generated native caches stay outside Git.

Superflip is recognized by exact cubie coordinates and uses
[Reid's published 20-turn lower bound](https://www.math.rwth-aachen.de/~Martin.Schoenert/Cube-Lovers/michael_reid__superflip_requires_20_face_turns.html)
plus a replayed 20-turn witness. General states exhaust every shorter IDA* depth
before an answer is returned; a timeout remains an explicit error.

NxN reduction measurements on the same Mac under background QoS (2026-10-07):
`make test-nxn` independently replayed **1,000 solutions**, with **100 uniform
legal states and 100 uniformly sampled single-layer scrambles per size**. States
are generated with independent integer 3D geometry; scrambles use 40×N turns.
The oracle covers all 24 grips, isolated and combined 4×4 OLL/PLL parity,
printed step counts and final CFEN. Every solve runs in a fresh CLI process.
Mean/max and times cover all 200 cases; the uniform column isolates the
100 uniform-state cases used for the move-length targets.

| Size | Mean / max moves | Uniform mean / max | Fresh CLI mean / max |
|---|---:|---:|---:|
| 2×2 | 18.89 / 20 | 18.78 / 20 | 76.86 / 271.37 ms |
| 4×4 | 92.59 / 128 | 92.09 / 115 | 290.48 / 1230.35 ms |
| 5×5 | 192.58 / 253 | 193.16 / 253 | 682.90 / 1745.57 ms |
| 6×6 | 457.26 / 522 | 452.01 / 509 | 707.85 / 1773.85 ms |
| 7×7 | 638.47 / 712 | 640.78 / 712 | 1217.23 / 2678.78 ms |

Outer, numbered-slice, wide and half turns, and grip rotations each count
once. Times include process startup, 3×3 cache loading and fresh reduction
tables; the 3×3 disk cache was already populated. These are sample
measurements, not worst-case bounds. The oracle fails if either combined
or uniform mean moves exceeds the documented value by more than 5%, or
if this table disagrees with the solving guide. The former per-piece
reduction averaged 371.87, 532.82, 1058.29 and 1378.36 moves on sizes 4–7.

Reduction uses center block searches on 4×4/5×5, batched bar commutators on
6×6/7×7, and slice-based edge pairing with short parity corrections.
Adjacent turns on one axis are canceled and packed into wide blocks.
All answers are verified, with no optimality guarantee. `--time-limit`
applies to the final Kociemba search.
See [the solving guide](./examples/solving.md#solve-other-sizes) for the reduction
method, parity examples and saved-state usage.

## Algorithm database and CFOP

The nine CSV files contain **159 rows** (including an empty `zz_misc.csv`). The
reproducible importer accepts **115**, merges **12** duplicate move sequences
across files/categories, and quarantines **44** with original rows and reasons
in [quarantine.json](./alg_dumps/quarantine.json). Together with the five existing
algorithms this produces **108 unique entries: 94 for 3×3, 14 for other sizes**.
Every entry has a concrete inverse-to-solved CFEN recognition pattern and passes
CFEN verification. The importer also checks that 3×3 OLL/PLL preserve F2L,
PLL preserves orientation, and F2L preserves the cross. Independent named
fixtures validate the 3×3 F2L/OLL/PLL case IDs. F2L descriptions and recognition
text derive corner/edge positions and sticker directions from those fixtures
in the standard FR frame before yaw/AUF. The importer does not independently
validate the claimed 2×2/big-cube case or parity description.

| Category | Entries |
|---|---:|
| F2L / OLL / PLL | 23 / 39 / 18 |
| Trigger / Advanced | 8 / 7 |
| Roux CMLL / LSE | 2 / 2 |
| 2×2 CLL / EG1 / EG2 / OLL / PBL | 6 / 1 / 1 / 1 / 2 |
| 4×4 / 5×5 / 6×6 parity | 2 / 1 / 1 |

Counts are category memberships; merged entries preserve aliases, categories,
and file/row provenance. **22 inverse pairs and 7 mirror pairs** are detected
from exact sticker permutations (including self-inverse/self-mirror cases).
Unknown relationships stay empty. Mirror means reflection in the left/right
plane; this does not claim every familiar case mirror matches without AUF.
Counts come from [import-report.json](./alg_dumps/import-report.json); a Go test
checks this section and TODO.md against the report so stale counts fail validation.

```sh
CUBE_CACHE_DIR=$PWD/.scratch/cube-cache taskpolicy -b nice -n 15 make import-algorithms
./dist/tools/verify-database
./dist/cube lookup --category F2L
./dist/cube solve "R U F2 L' B" --method cfop
```

CFOP uses `FindPattern` for a shortest white cross (at most eight face turns),
then recognizes safe one-slot F2L algorithms from the imported CFEN patterns,
with U setups and rotated slots.
When no database case matches, `FindPattern` inserts a pair while preserving the
cross and every solved slot. Last-layer recognition uses yellow-sticker masks
for OLL and full permutations for PLL. The valid database algorithms plus AUF
are composed into complete tables of **216 OLL orientations and 288 PLL
permutations**; incomplete raw case sets do not require a full-solver fallback.
Every checkpoint and the complete solution are replayed and verified.

The seeded 200-uniform-state Go oracle measured these real lengths on this Mac
under background QoS on 2026-10-07:

| Stage | Mean turns | Max turns |
|---|---:|---:|
| Cross | 5.790 | 8 |
| F2L 1 | 6.820 | 12 |
| F2L 2 | 6.715 | 12 |
| F2L 3 | 7.025 | 12 |
| F2L 4 | 7.350 | 12 |
| OLL | 10.635 | 18 |
| PLL | 13.070 | 20 |
| Total | **57.405** | **73** |

The same Go sample took **8.81 s total, 44.04 ms mean, 835.87 ms maximum**,
including first-use setup. A separate Python 3D geometry oracle replayed every
checkpoint on **200 independent uniform physical states**, including all 24
rigid grips: **58.275 mean, 75 maximum turns**. Fresh CLI processes averaged
**835.14 ms**, maximum **1796.34 ms** (167.03 s total, including table setup in
every process).
A face, wide or slice turn counts as one, including half turns; grip rotations
are excluded. Stage maxima can occur on different states. These samples are
not worst-case runtime or solution-length guarantees. This is paired F2L with
search fallback, without look-ahead or color neutrality.

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
./dist/cube solve "R U F2 L' B" --target-length 19 --time-limit 500ms --headless
./dist/cube solve "R U F2 L' B" --method beginner --headless
./dist/cube solve "R U F2 L' B" --method cfop --headless

# Return moves that solve the first layer (other layers may remain scrambled)
./dist/cube solve "R U F2 L' B" --goal first-layer --headless

# Verify an algorithm against CFEN states (sexy move x6 = identity → solved)
./dist/cube verify "R U R' U' R U R' U' R U R' U' R U R' U' R U R' U' R U R' U'" \
    --start "YB|Y9/R9/B9/W9/O9/G9" --target "YB|Y9/R9/B9/W9/O9/G9"

# Optimize a sequence
./dist/cube optimize "R R R"            # → R'

# Search for a short sequence that reaches a pattern
./dist/cube find sequence "R U"         # → U' R'
./dist/cube find sequence "R U F2 L' B D2 R' F U2 L" --max-moves 10
./dist/cube find --target 'YB|Y9/?9/?9/?9/?9/?9' --start 'YB|Y9/R9/B9/W9/O9/G9'
./dist/cube find sequence R2 --moves R --max-moves 2  # → R R
```

See [examples/](./examples/) for checked CLI commands, the beginner lesson,
algorithm IDs and short verification/search workflows. Run `make test-docs`
to check the runnable CLI examples against the built binary.

## Interactive Website

The static [cube playground](./web/) uses the real Go engine compiled to
WebAssembly. There is no npm build step, server API, account, or external CDN.

```bash
make web
python3 -m http.server --bind 127.0.0.1 8080
# Open http://127.0.0.1:8080/web/
```

Turn the 3D cube with WCA keys (`R U F L D B`, `M E S`, `x y z`); hold Shift
for prime turns. Drag to orbit the view, or use arrow keys while the cube is
focused. The buttons, algorithm box, undo/redo, reset, scramble, and 2D net
work on phones too. Solving produces a verified sequence with play/pause,
step forward/back, clickable moves, a position slider, and adjustable speed.
The selector beside Solve defaults to Kociemba and also offers Beginner and
CFOP. CFOP playback groups moves into Cross, F2L 1–4, OLL and PLL with case names,
turn counts and skips; every move retains click, step and scrub playback.
The solution heading identifies the method that ran.
Kociemba uses the same 20-turn goal and one-second search budget as the CLI;
each worker initializes its own tables before starting that budget. Hard states
return the best solution found at the deadline. Browser smoke tests check that
Solve finishes and animation frames keep running during a cold-worker solve.
Lesson mode gives the beginner method's actual instructions and checks,
replanning the next hint from your current state after your own turns.

Search accepts a 3×3 CFEN target with `?` wildcards and a depth of 0–10. It runs
in a cancellable Web Worker with a 30-second limit; start with short searches.
Solver and lesson planning also run in workers. The WASM binding calls
`ParseMoves` / `ApplyMoves`, `GetSolver`, `PlanBeginner`, `FindPattern`, and
the existing CFEN and physical validation APIs. Search uses the CLI's shortest
IDA* path, including its sticker BFS fallback, without a separate JavaScript
solver or searcher.

Import/export uses concrete CFEN in YB storage order (`U/R/F/D/L/B`); actual
center colors preserve rotated grips. Share copies a URL hash containing the
scramble and algorithm, or an imported/manual starting state. Opening a link
restores the starting cube and prepares its algorithm for playback.

```bash
taskpolicy -b nice -n 15 make test-web
cd web && npm install && cd ..   # Playwright is only a web devDependency
taskpolicy -b nice -n 15 make test-web-smoke
```

The smoke test uses headless Playwright Chromium with a disposable profile,
starts Python on a random loopback port, stops it on exit, and saves
`.scratch/screens/{cube,search,cfop}-{1280x800,390x844}.png`. It uses the cached
Chromium on macOS; elsewhere install Chromium with `cd web && npx playwright
install chromium`. All test temporary files stay in `.scratch/`.

If an execution sandbox denies loopback sockets, `taskpolicy -b nice -n 15
node web/test/smoke.mjs --in-memory` runs the same browser assertions using
local-file request routing. This alternate transport does not check Python
serving; the normal smoke target continues to require the Python server. The
2026-10-07 validation used this transport because the sandbox refused loopback
bind; WASM API checks and the full desktop/mobile browser assertions passed,
including hard-state Solve completion and 55 animation frames during its cold
worker's setup/search. Go tests, vet, all 133 CLI E2E cases and both 200-state
independent Kociemba/CFOP oracles also passed under background QoS.

To publish, run this **single command yourself** from the repository root:

```bash
taskpolicy -b nice -n 15 make web-pages && npx --yes gh-pages --dist dist/web --nojekyll
```

This builds and publishes only runtime assets to the `gh-pages` branch of the
current repository's origin. GitHub Pages must separately be configured to
serve that branch's root. All asset paths are relative, including worker and
WASM paths, so repository subpaths work. Publishing and Pages configuration
are intentionally not performed by this implementation task.

## Command Overview

| Command | Purpose | Status |
|---------|---------|--------|
| `twist` | Apply moves and render the cube | works |
| `verify` | Check an algorithm vs. CFEN start/target (`--start`/`--target`) | works |
| `show` | Render a cube with cross/OLL/PLL/F2L highlighting | works |
| `lookup` | Search the algorithm database | works |
| `optimize` | Cancel/merge a move sequence | works |
| `find` | Shortest exact/wildcard pattern search | works (3×3 IDA*, BFS fallback) |
| `parse-cfen` / `generate-cfen` / `verify-cfen` / `match-cfen` | CFEN utilities | works |
| `identify` / `show-alg` | Pattern identify / algorithm display | partial |
| `learn` | Teach a complete beginner solve, with recovery and checkpoints | works (3×3) |
| `solve --goal first-layer` | Solve the white first layer | works (3×3 beginner) |
| `solve` / `solve --goal full` | Kociemba; reduction for 4–7, corners for 2; beginner/CFOP on 3 | works (2×2–7×7) |
| `solve --optimal` | Prove a shortest face-turn solution within a time limit | works (deep states may time out) |

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
make e2e-test    # end-to-end CLI tests
make test-all    # both
make test-first-layer  # independent partial-goal regression oracle (Python 3)
make test-beginner     # independently replay full solutions and last-layer recovery
make test-kociemba     # 200 independent uniform physical states and move replay
make test-cfop         # 200 independent physical states, every CFOP checkpoint
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
  solver.go / kociemba.go / cfop.go # full beginner, Kociemba and CFOP solvers
  nxn*.go                        # 2x2 corners, center blocks and wing pairing on 4x4-7x7
  coordinates.go / coordinate_tables.go  # cubie moves and deterministic pruning tables
  pattern_search.go / optimal_search.go  # shortest wildcard/exact search + BFS fallback
  first_layer.go / full_lesson.go / lesson_session.go  # beginner checkpoints and recovery
  state_validation.go            # physical 3x3 legality checks
  solving_db.go                  # experimental 4-look pattern matcher (currently unwired)
  cubie.go                       # piece addresses, ranges and 3×3 selector aliases
internal/cfen/                   # CFEN parsing, generation, conversion, matching
tools/                           # import-algorithms, verify-*, generate-patterns
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
trial. Database CFOP now complements Kociemba and bounded optimal search;
color neutrality, look-ahead, shorter big-cube solutions and dimensions above 7 remain future work.
See [TODO.md](./TODO.md) and [docs/solvers.md](./docs/solvers.md).

## License

MIT License — see [LICENSE](LICENSE).
