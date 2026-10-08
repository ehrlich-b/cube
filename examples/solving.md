# Solving 2×2 through 7×7

Run from the repository root after `make build`. The default full solver supports
**2×2 through 7×7**; beginner, CFOP, optimal search and lessons support **3×3 only**.
The quoted moves describe a scramble applied to a solved cube;
for an already scrambled physical cube, provide its observed state with
`--start`. A new random scramble describes a different cube.

## Inspect moves, then solve

```sh
./dist/cube twist "R U R' U'" --color
./dist/cube solve "R U R' U'" --headless
```

Face turns are clockwise when looking directly at that face. An apostrophe
reverses the turn; `2` means a half turn. Keep spaces between moves.

The default method is Kociemba, a two-phase solver. It returns verified moves;
it does not promise a shortest solution. Beginner uses layer-by-layer move
groups. CFOP prints Cross, F2L, OLL and PLL checkpoints with database case names.

```sh
./dist/cube solve "R U F2 L' B" --method kociemba --headless
./dist/cube solve "R U F2 L' B" --method beginner --headless
./dist/cube solve "R U F2 L' B" --method cfop
```

Return the final solved state instead of moves:

```sh
./dist/cube solve "R U F2 L' B" --cfen --headless
```

```text cube-output
YB|Y9/R9/B9/W9/O9/G9
```

## Learn or stop after the white layer

For a complete physical-cube walkthrough, follow
[the beginner lesson](./beginner.md). Its interactive mode waits for your
checkpoint confirmations and can record actual moves and recover mistakes.

```sh
./dist/cube learn "R U F2 L' B" --interactive --color
./dist/cube solve "R U F2 L' B" --goal first-layer --headless
```

`--goal first-layer` solves only the white cross and white corners; the other
layers can remain scrambled. See [the white-layer guide](./first-layer.md).

## Solve other sizes

For 4×4 through 7×7, reduction solves centers and pairs every edge orbit, then
finishes through the existing 3×3 Kociemba solver. Odd fixed centers determine
the cube's orientation. Before solving centers, one inner turn corrects any odd
wing permutation (OLL parity); pairing to an edge state with the corners' parity
also handles PLL parity. Pure three-piece cycles preserve completed centers and
other edge orbits. The 2×2 uses the same 3×3 solver's corners path.

These are verified complete solutions, with correctness ahead of move count.
They do not promise short or optimal sequences. `--time-limit` and
`--target-length` control the final 3×3 search; the reduction happens first.

```sh
./dist/cube solve "R U R' U'" --dimension 2 --headless
./dist/cube solve "Rw U2 Rw'" --dimension 4 --headless
./dist/cube solve "2R 3L' Uw2" --dimension 5 --headless
```

```sh
./dist/cube solve "2R 3U Fw" --dimension 6 --cfen
```

```text cube-output
YB|Y36/R36/B36/W36/O36/G36
```

```sh
./dist/cube solve "M E S x Rw" --dimension 7 --cfen
```

```text cube-output
YB|Y49/R49/B49/W49/O49/G49
```

The two even-cube parity cases also solve completely:

```sh
./dist/cube solve "2R2 B2 U2 2L U2 2R' U2 2R U2 F2 2R F2 2L' B2 2R2" --dimension 4 --headless
./dist/cube solve "2R2 U2 2R2 Uw2 2R2 Uw2" --dimension 4 --headless
```

Use `--dimension N --start "YB|..."` for an observed NxN state. The CFEN must
contain exactly N² concrete stickers per face; wildcards are search patterns.
An optional scramble is applied after that saved state, as for 3×3 solving.

`make test-nxn` independently samples and replays 100 uniform legal states and
100 uniform-move scrambles per size in fresh CLI processes. It also checks all
24 grips, isolated/combined OLL and PLL parity, printed step counts and final
CFEN. Background-QoS measurements on this Mac (2026-10-07), over all 200 cases
per size, were:

| Size | Mean / max moves | Fresh process mean / max |
|---|---:|---:|
| 2×2 | 19.50 / 20 | 85.24 / 231.06 ms |
| 4×4 | 371.87 / 444 | 95.79 / 235.50 ms |
| 5×5 | 532.82 / 611 | 182.99 / 1131.25 ms |
| 6×6 | 1058.29 / 1169 | 145.78 / 713.70 ms |
| 7×7 | 1378.36 / 1510 | 261.22 / 1189.32 ms |

Moves count each numbered slice turn, half turn and grip rotation once. Times
include startup, fresh reduction setups and loading an already populated 3×3
disk cache. These are sample results, not worst-case guarantees.

See [advanced examples](./advanced.md) for CFEN verification and bounded search,
and [algorithm lookup](./algorithms.md) for stored case IDs.
