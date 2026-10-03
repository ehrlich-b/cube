# Finish a 3×3 with the beginner lesson

Build and start a complete lesson:

```sh
make build
./dist/cube learn "R U F2 L' B" --interactive --color
```

For this example, start with a solved cube held white down, yellow up, blue
front and red right, then perform the quoted scramble. The program tracks those
exact moves. For an already scrambled cube, supply its known move history or a
complete observed CFEN state using `--start`; a random scramble describes a
different cube.

The lesson finishes all six faces using these stages:

| Stage | What to check at its endpoint |
| --- | --- |
| White cross | White down; each edge's side color matches its center |
| White corners | The white face and bottom row on every side are solved |
| Middle edges | The bottom two rows on every side are solved |
| Yellow cross | All four top edge stickers are yellow |
| Yellow edge alignment | The cross also matches all four side centers |
| Yellow corner placement | Each corner has the three colors of its surrounding centers |
| Yellow corner orientation | All six faces are uniform and match their centers |

Already satisfied stages are skipped. Each checkpoint names its piece or goal,
prints exact move groups, explains the grip, and describes the resulting state.
Use the printed whole-cube rotations when requested; subsequent face letters
refer to your new grip. Setup rotations return to white down, blue front before
the checkpoint finishes.

## Follow a checkpoint

Perform every `Do:` group in order, then inspect the printed `Check:`. Type
`next` only after completing that checkpoint. `next` records all its moves and
plans the next checkpoint. `hint` repeats the current instruction and `show`
shows the tracked state; neither changes it.

Face letters mean front/back/right/left/up/down. A plain face letter is a
clockwise quarter turn **looking directly at that face**, including B and D.
An apostrophe reverses it, and `2` means a half turn. `x`, `y`, and `z` rotate
the whole cube in the direction of R, U, and F: x brings the front to the top,
and y brings the right to the front. Face moves and whole-cube rotations are
different actions.

## Finish the entire final corner sweep

The final checkpoint repeats `R' D' R D` to turn the corner at top front-right,
then uses U to bring the next corner there. **The lower layers temporarily look
scrambled.** Keep the same grip and perform the complete printed sweep; the
lower layers return at its endpoint.

Follow the exact printed repetition counts. Always finish the D at the end of
each group, even if yellow appears upright earlier. Perform the printed U moves
for corners that were already upright too, including the final U that restores
top alignment. Check all six faces only after the whole checkpoint, then type
`next`. There is one checkpoint for the entire four-corner sweep.

## Recover from actual moves

If you did a different sequence, or stopped partway through a checkpoint, use
`moves` instead of `next`:

```text
moves R U R'
```

Record **all actual moves since the last recorded state**, including any grip
rotations. The program applies them and replans from there. An interrupted final
sweep may need to restore earlier layers; this is expected recovery from the
recorded state. If the move history is unknown, restart using a fresh observed
CFEN state.

`undo` and `reset` print exact inverse sequences. Perform those inverses on your
physical cube as well: the program cannot move it for you. `undo` returns one
recorded batch; `reset` returns to this session's initial state. `state` prints
the tracked CFEN. `quit` or end of input prints a ready-to-run resume command
that preserves both the state and the selected goal. Repeated `next` after
completion leaves the solved state and grip unchanged.

## Start from a saved or observed state

```sh
./dist/cube learn --start "YB|Y9/R9/B9/W9/O9/G9" --interactive
./dist/cube solve --start "YB|Y9/R9/B9/W9/O9/G9" --headless
```

Lesson input uses `YB|` storage order: U / R / F / D / L / B, each face's nine
stickers read in rows while looking directly at that face. Repeated colors can
use counts, as in `Y9`. The actual center stickers determine the grip; the YB
prefix alone does not imply that an observed cube is held yellow up, blue front.
Lessons require a concrete state without wildcard stickers and reject other
storage prefixes. A positional move sequence with `--start` is applied after
that supplied state.

## Print a lesson or obtain a solution

```sh
# Print every checkpoint without recording interactive progress
./dist/cube learn "x y R U F2 L' B"

# Return moves that solve the complete cube
./dist/cube solve "R U F2 L' B" --headless

# Return the final solved CFEN
./dist/cube solve "R U F2 L' B" --cfen --headless

# Stop at the white first layer
./dist/cube learn "R U F2 L' B" --goal first-layer --interactive
./dist/cube solve "R U F2 L' B" --goal first-layer --headless
```

The full beginner path supports standard 3×3 cubes. It favors familiar move
groups over short solutions. Other dimensions, unavailable CFOP/Kociemba
algorithms, malformed notation, and impossible physical states produce errors.
Color counts, center relationships, cubie identities, flips, twists and parity
are checked before planning.

The tests independently replay physical stickers and interactive recovery.
**A human physical-cube trial has not occurred**, so ease of reading and following
the lesson by hand remains unverified. See the
[verification receipt](../docs/verification-full-2026-10-03.md).
