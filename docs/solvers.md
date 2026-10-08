# Solver status and contracts

Updated 2026-10-07.

## Fast full-cube solving

`cube solve` defaults to Kociemba; both `--method kociemba` and the existing
`--algorithm kociemba` select it. `--method beginner` keeps the full beginner
solver. `cube learn` and `--goal first-layer` continue to use the beginner path.
`Solver`, `SolverResult`, `GetSolver`, and the existing solver signatures are
unchanged. Solving accepts only validated physical 3×3 states. Every nonempty
result is replayed and checked on all six faces before it is returned, and the
CLI checks it again before printing. Planning does not mutate the input.

The implementation follows [Kociemba's two-phase coordinate construction](https://kociemba.org/math/twophase.htm).
Ordered facelets are shared with physical validation. Cubie moves are extracted
from the corrected sticker engine, then tested against 6,000 sticker
transitions. Phase one tracks corner twist (2187), edge flip (2048) and the slice
edge subset (495). Phase two tracks corner permutation (40320), U/D edge
permutation (40320) and slice permutation (24), using U/D turns and side-face
half turns. Phase-one pruning takes the maximum of exact twist/slice, flip/slice
and twist/flip distances. Three reduction axes and their inverses search with
resumable DFS stacks, sharing a single budget without restarting at each switch.
Phase-one depth increases up to 12; phase two uses at most 18 turns. Every new
incumbent tightens the total bound to one less than its length.
`--target-length` defaults to 20 face turns and `--time-limit` to one second.
Search stops on reaching the target or returns its best verified solution when
the budget expires; expiry before any solution is an explicit error. Setup is
separate from this budget so cold WASM workers can finish initialization.
`--optimal` retains its distinct timeout contract, including setup and rejecting
unproved answers. Kociemba does not promise a globally shortest sequence or a
20-turn answer for every state. These flags are rejected for beginner/CFOP and
the partial first-layer goal rather than silently ignored.

The native and WASM executables embed the engine-owned compact asset
`internal/cube/tables/coordinates-v5.bin.gz` (**3,337,787 bytes**). An untimed
first use with an empty cache copies those exact bytes to the user's cache
directory under `cube/`; `CUBE_CACHE_DIR` overrides that directory. Ordinary
Kociemba does not generate large tables. The cache files, including their
checksums and move-fingerprint headers, are:

| File | Bytes | Used by |
|---|---:|---|
| `coordinates-v5.bin.gz` | 3,337,787 | Embedded compact tables for Kociemba and coordinate search |
| `edges-v1.bin` | 14,256,064 | Four-edge databases for CFOP and exact-state search, generated on first use |
| `phase1-sym8-v1.bin` | 147,502,144 | Optional large phase-one table (140.67 MiB) |
| `optimal-v1.bin` | 134,568,064 | Full-corner and two six-edge databases for deep optimal search (128.33 MiB) |
| `optimal-phase1-sorted-sym16-cap11-v1.bin` | 1,021,870,144 | Saturated sorted phase-one database for deep optimal search (974.53 MiB) |

Versioned filenames, checksums, dimensions and move fingerprints guard cached
data. Loaders reject non-regular files before opening them, including FIFOs and
symlinks to FIFOs. Missing, unreadable or corrupt compact caches use the embedded
asset; other requested databases rebuild in memory. A read-only filesystem does
not prevent solving. Writes use temporary files in the destination directory and
atomic rename, and failures leave any existing complete cache intact. WASM keeps
tables in memory when filesystem caching is unavailable.

`cube tables build --large` explicitly builds and persists the optional phase-one
file, allowing a few minutes and about 1.6 GiB of temporary memory. The command
reports success only after persistence and exits nonzero with a diagnostic on
failure. `CUBE_LARGE_TABLES=1` selects that table for Kociemba; ordinary solves
ignore it even when it exists. Deep `--optimal` search generates `optimal-v1.bin`
lazily (about one minute in the recorded run) and the saturated sorted table
`optimal-phase1-sorted-sym16-cap11-v1.bin` (23m08.27s in the recorded run).
The latter fixes the four slice edges in place, uses sixteen symmetries and stores
two-bit modulo-three distances through ten, with an admissible bound of eleven
for remaining states. Its three oriented bounds, inverse root selection and two
search workers strengthen deep optimal search. The two optimal files total
**1,156,438,208 bytes**; ordinary solves use neither. Optimal initialization,
including cache I/O and generation, counts toward `--time-limit`; timed compact/edge setup skips
optional writes, and expired large-database writes never publish partial caches.

Two Go oracles check 200 seeded scrambles and 200 uniformly generated legal
cubie states, requiring mean length ≤20, maximum 21 and ≥95% at ≤20 turns.
The separate Python oracle constructs 200 uniform physical states and replays
every returned move through independent geometry, including all 24 grips.
It measured mean/max **19.775/21 turns**, **98.5% at ≤20**, and mean/max process
latency **162.53/1069.96 ms** on this Mac under background QoS (before:
**21.730/22**, **2.5% at ≤20**, **68.51/256.60 ms**). The loaded-table Go uniform
sample averaged **19.790 turns**, maximum **21**, at **89.47/1000.19 ms** mean/max
(before: **21.675/22**, **18.47/267.90 ms**). Expanded table generation took
**1.83 s** in that earlier runtime-generated implementation; the current default
loads the embedded compact asset described above. These are reproducible
historical samples, not latency guarantees; current compact-table measurements
are in [README](../README.md).
Grip normalization uses rotation notation and its moves
are separate from face-turn distance.

## Shortest sequence and pattern search

`cube find sequence`, `cube find pattern` and `cube find --target CFEN` use IDA*
for representable 3×3 states. `--start` takes a concrete CFEN, `--from` retains
scramble input and also accepts CFEN, and `--moves` selects an alphabet. Exact
targets are reduced to a relative cubie state. The heuristic is the maximum of
orientation/slice bounds, corner-permutation distance, and three four-edge
pattern databases (12P4×16 states each). Their transition table is shared and
generated on first use and optionally cached as `edges-v1.bin`. Targets with wildcard stickers use
multi-source single-piece distance tables and are checked exactly at endpoints.
Centers stay in the starting frame: search cannot insert grip rotations into a
face-turn-only alphabet.

Iterative deepening proves shortestness in the supplied alphabet. Same-face
pruning only removes adjacent moves when their composition is identity or an
allowed single move; this preserves answers such as `R R` when only `R` is
allowed. Opposite faces commute and are searched in one canonical order.
The result is one shortest sequence. Non-3×3 states and alphabets with slices,
wide moves or rotations fall back to sticker BFS. The original CLI BFS and its
correctness harness remain available for differential tests.

Forty random cases match the retained BFS through depth four. Sixty random
start/target pairs match two independent radius-three sticker BFS frontiers,
which prove the same forward BFS distances through depth six without the huge
forward frontier. Another 100 randomized wildcard cases match sticker BFS.
Earlier measurements of ten exact targets at each depth 8/9/10 averaged **0.29/1.58/9.40 ms** with tables
loaded; fresh CLI processes took **0.72/0.46/0.52 s**, including setup.

`cube solve --optimal --time-limit 1s` uses the same admissible exact-state
search, counting each half turn as one move. The limit includes table loading
and generation. A timeout returns an explicit error with no unproved answer.
Deep optimal searches can time out; use Kociemba for general solving.

## Complete beginner goal

`cube solve [scramble] --method beginner` and `cube learn [scramble]` finish all six faces of
a validated physical 3×3. `--start` accepts a complete YB-storage CFEN; its real
centers determine the grip. `--goal first-layer` keeps the prior explicit
white-layer sublesson. See [the beginner guide](../examples/beginner.md).

The complete planner follows this path:

1. Validate shape, colors, rigid center frame, unique cubies, edge flips,
   corner twists/handedness and permutation parity. An already solved cube
   needs no moves, including when held in a rotated grip.
2. Normalize unsolved input to white down, blue front. Solve the white cross
   using incremental bounded search, then four corners with setup turns and
   repeated `R U R' U'`. Verify earlier pieces at each endpoint.
3. Insert the four middle edges using left/right eight-turn insertions.
   Eject misplaced edges first. Verify white and earlier middle edges.
4. Make the yellow cross with U setups and `F R U R' U' F'`.
5. Match yellow edges to their side centers with U setups and
   `R U R' U R U2 R'`.
6. Put yellow corners into their home slots using y-conjugated
   `U R U' L' U R' U' L`, ignoring their twists at this stage.
7. Orient all yellow corners in one checkpoint using counted `R' D' R D`
   repetitions and U advances, including the final U. The lower layers are
   temporarily mixed inside this sweep and restored at its endpoint.
8. Verify all six faces are solved before returning.

Last-layer planning searches only named algorithm groups over tiny finite
stage keys (8 cross patterns, 24 edge permutations, 12 corner permutations),
with explicit depths and a 64-state limit. It does not search arbitrary face
turns on a full scramble. The complete plan has a 600-move safety bound;
[measured counts and timings](./verification-full-2026-10-03.md) are smaller,
but are samples rather than a performance guarantee.

`BeginnerLesson` reuses action groups and checkpoint snapshots from the partial
lesson. Planning never mutates input. `BeginnerSolver` flattens only complete
plans into the full-cube `Solver` interface. The unchanged full-solver contract
now actively checks all five beginner cases, with no empty-result skips.
The CLI also independently reapplies the returned full solution before output.

## Recoverable interaction

A `FirstLayerSession` supports both goals via `NewFirstLayerSession` and
`NewBeginnerSession`. It replans after each checkpoint or recorded actual-move
batch. Undo stores snapshots and exact physical inverses; reset returns the
initial state. Repeated next after completion preserves both state and grip.
Quit and EOF print executable resume commands retaining the selected goal.
Input errors and bounded oversized lines leave the current state unchanged.

The final corner sweep is atomic as a lesson checkpoint. A user who interrupts
it can record the true prefix with `moves` and replan; intermediate lower-layer
scrambling is a physical state, not an assumed completed checkpoint. The
independent live oracle exercises this path, wrong turns and grip changes.

## Engine corrections retained from the first-layer milestone

Scramble/inverse tests alone could pass when both used the same incorrect
mapping. The independent 3D oracle exposed grid-orientation errors, repeated
whole-cube rotations, reversed x/E directions and missing far-face turns.
The first-layer commit repaired these mappings. Unique-label tests check
rotations, faces, wide widths and numbered layers for sizes 2–6, slices for
sizes 3 and 5, and standard rotation/slice identities. x/y/z follow R/U/F and
E follows D. Load-bearing conservation and inverse invariants remain intact.

## Remaining work

CFOP now solves a complete 3×3 using `FindPattern` for an optimal white cross,
paired database F2L insertions with search fallback, and OLL/PLL pattern matching
with AUF. The imported database supplies all last-layer moves; reverse Dijkstra
composes valid algorithms into complete 216-orientation and 288-permutation
recognition tables when the raw cases are incomplete. Cross/F2L search uses
admissible fixed four-edge and corner-edge pair bounds. Every checkpoint and
full solution are replayed, and the site exposes the same seven named stages.
The 200-state Go oracle measures 57.405 mean and 73 maximum turns; see README
for stage lengths and measured timings.

CFOP color neutrality, extended cross and F2L look-ahead remain future work.
Full CLI solving supports 2×2 through 7×7: the 2×2 follows a 3×3 corners path,
and 4–7 solve centers and pair wings before calling the public Kociemba solver.
Reduction corrects odd wing permutations before restoring centers and chooses
a reduced edge state whose permutation parity matches the corners. This handles
OLL/PLL parity, while odd fixed centers define the orientation. Pure center and
wing three-cycles preserve completed pieces. The returned sequence is replayed
to check every sticker against the solved center frame. See
[NxN examples and measurements](../examples/solving.md#solve-other-sizes).
Beginner, CFOP, optimal search and lessons still support 3×3 only.
`solving_db.go` remains an unwired historical experiment. All 131 database
entries carry inverse-to-solved verification patterns; 15 raw rows are quarantined.
Shorter big-cube solutions, dimensions above 7, stronger deep optimal search and human usability work remain
future milestones.

## Verification

```sh
make build-all-local test-all test-first-layer test-beginner
make test-kociemba test-cfop
make fmt vet
go test ./internal/cube -run '^$' -bench '^BenchmarkBeginnerFull$' -benchmem
```

Go tests include 500 deterministic full-solve scrambles, checkpoint replay,
recovery, repeated steps and illegal states. The separate Python oracle
constructs legal physical cubies without the Go engine, then independently
replays headless moves, printed checkpoints and interactive commands. Its
last-layer cases cover every orientation/permutation class. A human holding a
physical cube has not tested the lesson. Historical first-layer evidence is
preserved in [the earlier receipt](./verification-2026-10-03.md).
