# Solving a 3×3

Run from the repository root after `make build`. All full solvers support
**3×3 only**. The quoted moves describe a scramble applied to a solved cube;
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

## Inspect other sizes

The move engine supports other dimensions. Full solving and lessons do not.

```sh
./dist/cube twist "R U R' U'" --dimension 2 --color
./dist/cube twist "Rw U2 Rw'" --dimension 4 --color
./dist/cube twist "2R 3L' Uw2" --dimension 5 --color
```

See [advanced examples](./advanced.md) for CFEN verification and bounded search,
and [algorithm lookup](./algorithms.md) for stored case IDs.
