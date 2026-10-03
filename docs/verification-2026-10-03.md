# First-layer milestone verification — 2026-10-03

Environment: Bryan's Mac, Darwin arm64, Apple M4. All runs were local and
bounded with `GOMAXPROCS=2`. Build/module caches were under `/tmp` to respect
the workspace sandbox. Only the existing public Go modules were downloaded;
no paid model service, WSL, push, deploy or credential change was used.

Base: `f34d3e5` on clean `main`. Work is isolated on
`codex/cube-first-layer-20261003` in
`/Users/ehrlich/repos/cube-first-layer-20261003`. The original checkout and
canonical board were not edited.

## Delivered behavior

- `cube learn [scramble]`: white cross and four corners, named pieces, move
  groups, physical sticker checks and CFEN after each checkpoint.
- `--interactive`: next/hint/show/state, actual-move recovery, exact inverse
  undo/reset, saved resume command, safe repeated steps and oversized-input recovery.
- `cube solve <scramble> --goal first-layer`: normal, headless moves and CFEN.
- Physical 3×3 legality validation before planning; rotated input supported.
- Rigid rotation, E-direction and far-layer engine fixes backed by geometry.

Partial results remain separate from the full solver interface. No invariant
was weakened; the incorrect legacy x expectation was corrected using independent
physical geometry. Full-cube solving remains unimplemented.

## Checks and outcomes

| Check | Outcome |
|---|---|
| Baseline `make build-all-local test-all`, database verification | Green before changes |
| Focused first-layer, session and CLI Go tests | Green |
| First-layer replay oracle | 500/500 deterministic 25-turn scrambles, with additional rotations/slices/wide turns |
| Corner setup cases | 96/96 |
| Checkpoint recovery | Empty, extra-turn, interrupted-trigger, rotation and slice deviations all solved |
| State validation | Rejects wrong shape/counts/centers, wildcards, duplicate/mirrored pieces, flip/twist/parity defects |
| Independent 3D move geometry | 696 cases, sizes 2–6 (slices: 3 and 5) |
| Aggregate `make build-all-local test-all test-first-layer vet` | Exit 0 |
| E2E harness | 114 cases, zero failures; includes five legacy full-solver skips counted as passes by the existing harness |
| Full-solver contract | Preserved; its 15 empty-stub subcases still skip, not evidence of a working full solver |
| Database verification | 5/5 patterned algorithms |
| `make fmt vet` and `git diff --check` | Green |

The independent reviewer constructed physical cubie states from 3D sticker
adjacency, balanced twists/flips and matching parity, rather than generating
all inputs with the Go move engine. The reusable oracle is
`test/first_layer_oracle.py`; run it with `make test-first-layer`.

Its final aggregate run passed:

- 120 legal cubie states across all 24 rigid orientations.
- Four impossible-state fixtures and nine invalid CLI cases.
- Seven scripted interactive sessions.
- Three prompt-driven sessions that physically replayed printed instructions
  with a separate model: 6, 7 and 9 checkpoints (22 total).
- A regression check for rotated solved-state orientation wording.

The independent review found and verified fixes for an orientation instruction
that could be obeyed twice, incorrect bottom-row completion wording on a rotated
cube, duplicate error output and an oversized line that originally ended the
session. Final review found no remaining correctness blocker.

## Performance and UX limits

The 500-scramble corpus averaged **79.6 moves**, maximum **119**, including
whole-cube rotations. The lesson prioritizes simple triggers and restored
checkpoints, not optimal solutions.

`BenchmarkFirstLayer` on one fixed rotated scramble, three one-second runs:
**0.325–0.337 ms/op**, approximately **1.02 MB/op** and **5,232 allocations/op**.
This is one fixture, not a worst-case bound or a memory optimization claim.

The final 120-state oracle measured separate CLI processes at median
**6.09 ms**, maximum **17.91 ms**. Independent review's preceding run measured
median **6.02 ms**, maximum **20.40 ms**; an earlier run under concurrent load
reached **142.87 ms**. These measurements include process startup and are
observations on this Mac, not general latency guarantees.

UX evidence is simulated physical replay, including recovery, undo/reset and
every printed checkpoint. **No human beginner trial with a physical cube was
performed.** Recovery requires a known actual move sequence or an observed
complete CFEN state. Lesson input supports 3×3 only, the standard project color
scheme and YB storage order; other CFEN frames and wildcard patterns are rejected.
The middle and last layers remain the next implementation gap.

## Reproduce

```sh
GOCACHE=/tmp/cube-go-cache GOMODCACHE=/tmp/cube-go-mod GOMAXPROCS=2 \
  make build-all-local test-all test-first-layer vet
GOCACHE=/tmp/cube-go-cache GOMODCACHE=/tmp/cube-go-mod GOMAXPROCS=2 \
  go test ./internal/cube -run '^$' -bench '^BenchmarkFirstLayer$' -benchmem -benchtime=1s -count=3
./dist/cube learn "R U F2 L' B" --interactive --color
```

Aggregate command output from this session is retained locally at
`/tmp/cube-aggregate-20261003.log` (temporary, not required to reproduce).
