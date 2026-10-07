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
half turns. Pairwise exact-distance tables supply admissible bounds; searches
try multiple phase-one endpoints. They prefer at most 22 face turns for one
second, then permit the complete 12+18 two-phase bound of 30 turns on hard cases.
Kociemba does not promise a globally shortest sequence.

Tables are generated deterministically and cached in the user's cache directory
under `cube/coordinates-v1.gob`; `CUBE_CACHE_DIR` overrides it. The cache has a
versioned filename, checksum and shape checks. Missing, unreadable or corrupt
caches rebuild in memory; a read-only filesystem does not prevent solving.
Writes use a temporary file in the destination directory and atomic rename.
The generated cache is about 7.6 MiB, with no generated table files committed.
WASM can use the same public entry points and retain tables in memory when
filesystem caching is unavailable.

Two Go oracles check 200 seeded scrambles and 200 uniformly generated legal
cubie states, rejecting answers longer than 22 face turns in those samples.
The separate Python oracle constructs 200 uniform physical states and replays
every returned move through independent geometry, including all 24 grips.
It measured mean/max **21.73/22 turns** and mean/max process latency
**71.5/365.7 ms** on this Mac under background QoS. The Go scramble sample measured
**21.68/22 turns**, **13.5/113.6 ms** with loaded tables. Generation measured
**0.67 s**, and warm cache loading **26 ms**. Thirty paired scrambles averaged
**21.70 Kociemba moves vs 209.57 beginner moves**. These are reproducible samples,
not latency guarantees. Grip normalization uses rotation notation and its moves
are separate from face-turn distance.

## Shortest sequence and pattern search

`cube find sequence`, `cube find pattern` and `cube find --target CFEN` use IDA*
for representable 3×3 states. `--start` takes a concrete CFEN, `--from` retains
scramble input and also accepts CFEN, and `--moves` selects an alphabet. Exact
targets are reduced to a relative cubie state. The heuristic is the maximum of
orientation/slice bounds, corner-permutation distance, and three four-edge
pattern databases (12P4×16 states each). Their transition table is shared and
generated in memory on first use. Targets with wildcard stickers use
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
Ten exact targets at each depth 8/9/10 averaged **0.29/1.58/9.40 ms** with tables
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
The engine supports larger cubes, but full solving and lessons support 3×3 only.
`solving_db.go` remains an unwired historical experiment. All 131 database
entries carry inverse-to-solved verification patterns; 15 raw rows are quarantined.
Generic NxN solving, stronger deep optimal search and human usability work remain
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
