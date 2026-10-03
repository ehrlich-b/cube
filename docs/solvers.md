# Solver status and contracts

Updated 2026-10-03.

## Complete beginner goal

`cube solve [scramble]` and `cube learn [scramble]` now finish all six faces of
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

CFOP and Kociemba remain empty API stubs; the CLI rejects them instead of
emitting misleading empty solutions. The engine supports larger cubes, but
beginner solving and lessons support 3×3 only. `solving_db.go` remains
experimental unwired code; the algorithm database has five verification
patterns. Optimal solutions, generic NxN solving and human usability work
remain future milestones.

## Verification

```sh
make build-all-local test-all test-first-layer test-beginner
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
