# Solver status and contracts

Updated 2026-10-03.

## Working first-layer goal

`cube learn [scramble]` and `cube solve <scramble> --goal first-layer` solve
the white cross and four corners on a valid 3×3 cube. The middle and last
layers may remain scrambled. See [the user guide](../examples/first-layer.md).

The implementation follows this path:

1. Validate the complete physical state: shape, colors, center frame, unique
   cubies, edge flips, corner twists/handedness and permutation parity.
2. Find a rigid rotation to white down and blue front by examining centers.
   Include those rotations in the returned moves and lesson instructions.
3. Solve four white edges with the existing incremental bounded search,
   restoring earlier placed edges at each endpoint. This is a fixed order,
   not a globally optimal cross.
4. Lift an unsolved bottom corner through its own slot when necessary. Bring
   its destination to bottom front-right, line it up with U, and repeat
   `R U R' U'` until solved (at most six repetitions). Return to blue front.
5. Verify the cross, new corner and previously solved corners at every corner
   checkpoint, then verify the whole first layer before returning.

`FirstLayerLesson` has its own result type, checkpoint snapshots and flattened
moves. It never implements the full-cube `Solver` interface. Planning does not
mutate input. The lesson rejects results beyond 260 moves; observed counts
are recorded in [the verification receipt](./verification-2026-10-03.md).

`FirstLayerSession` replans from the current state after every checkpoint or
actual-move batch. Undo history stores snapshots and exact inverse moves.
Repeated next after success is a no-op; reset returns the initial state.

## Engine corrections needed for the lesson

Existing scramble/inverse tests could pass even if a move and its inverse
both used the same incorrect physical mapping. The new 3D oracle exposed:

- Whole-cube rotations copied grids without all required reversals/rotations.
- The same whole-cube permutation ran once per layer, repeating N times.
- x used the opposite direction to R; E used the opposite direction to D.
- A turn at the far outer layer omitted rotation of the opposite face.

These mappings now preserve physical cubie adjacency. Unique-label geometry
tests cover rotations, faces, every valid wide width and numbered layer for
sizes 2–6, slices for sizes 3 and 5, and standard rotation/slice identities.
The legacy x-center expectation was corrected against this independent oracle;
the load-bearing solver, conservation and inverse invariants were preserved.

## Full-cube solving remains future work

`BeginnerSolver`, `CFOPSolver` and `KociembaSolver` still return empty results
for unsolved inputs. `solve --goal full` remains the default for compatibility
and explains the limitation in normal output. Legacy headless full solving can
still emit an empty string. Callers must explicitly choose `--goal first-layer`
to obtain the working partial solution.

The full-solver invariant remains unchanged: any nonempty full solution must
solve the entire cube. Empty stub results are skipped in the contract tests;
they are not evidence of working full solvers.

Next coherent milestones are beginner middle-layer edge insertion, then last
layer orientation and permutation. `solving_db.go` remains experimental,
unwired code; the algorithm database still has only five verification patterns.
Generic NxN piece solving, globally optimal search, CFOP and Kociemba remain
future work.

## Verification

```sh
make build-all-local test-all
make test-first-layer
make fmt vet
go test ./internal/cube -run '^$' -bench '^BenchmarkFirstLayer$' -benchmem
```

Go tests cover 500 deterministic scrambles, 96 corner setup cases, recovery
at every checkpoint, repeated steps and invalid physical states. The separate
Python oracle generates legal cubies without the Go engine and replays the
printed interactive instructions through its own geometry model. Human physical
usability has not been tested.
