#!/usr/bin/env python3
"""Independent physical replay, short IDA* cross-checks and superflip = 20."""

import argparse
import random
import re
import subprocess
import sys
from pathlib import Path

sys.dont_write_bytecode = True
import first_layer_oracle as physical


def solved(state):
    return all(color == face[1][1] for face in state for row in face for color in row)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--binary", type=Path,
                        default=Path(__file__).resolve().parents[1] / "dist" / "cube")
    args = parser.parse_args()
    binary = str(args.binary.resolve())
    rng = random.Random(2026100717)
    cases = 0
    for depth in range(1, 9):
        for _ in range(4):
            scramble = " ".join(rng.choice("URFDLB") + rng.choice(["", "2", "'"])
                                for _ in range(depth))
            state = physical.physical_sequence(physical.fresh(), scramble)
            answer = subprocess.run([binary, "solve", "--optimal", "--time-limit", "5s",
                                     "--start", physical.cfen(state), "--headless"],
                                    capture_output=True, text=True, timeout=10)
            assert answer.returncode == 0, (scramble, answer.stderr)
            assert solved(physical.physical_sequence(state, answer.stdout)), (scramble, answer.stdout)
            turns = len(answer.stdout.split())
            assert turns <= depth, (scramble, answer.stdout)
            finder = subprocess.run([binary, "find", "pattern", "solved", "--start",
                                     physical.cfen(state), "--max-moves", str(depth)],
                                    capture_output=True, text=True, timeout=10)
            assert finder.returncode == 0, finder.stderr
            if solved(state):
                assert turns == 0
            else:
                match = re.search(r"^1\. (.*?) \((\d+) moves", finder.stdout, re.M)
                assert match, finder.stdout
                assert turns == int(match.group(2)), (scramble, answer.stdout, finder.stdout)
                assert solved(physical.physical_sequence(state, match.group(1)))
                shorter = subprocess.run([binary, "find", "pattern", "solved", "--start",
                                          physical.cfen(state), "--max-moves", str(turns-1)],
                                         capture_output=True, text=True, timeout=10)
                assert "No sequences found" in shorter.stdout, shorter.stdout
            cases += 1
    # Construct superflip by swapping the two stickers of each physical edge,
    # independently of the solver's cubie layout and its certificate sequence.
    state = physical.fresh()
    for a, b in physical.EDGES:
        fa, ra, ca = a
        fb, rb, cb = b
        state[fa][ra][ca], state[fb][rb][cb] = state[fb][rb][cb], state[fa][ra][ca]
    for grip in physical.orientation_paths():
        oriented = physical.physical_sequence(state, grip)
        answer = subprocess.run([binary, "solve", "--optimal", "--time-limit", "1s",
                                 "--start", physical.cfen(oriented), "--headless"],
                                capture_output=True, text=True, timeout=3)
        assert answer.returncode == 0, answer.stderr
        turns = [m for m in answer.stdout.split() if m[0] not in "xyz"]
        assert len(turns) == 20, answer.stdout
        assert solved(physical.physical_sequence(oriented, answer.stdout)), answer.stdout
        cases += 1
    print(f"{cases} independent optimal checks passed: 32 short states cross-checked "
          "with IDA* finder, superflip = 20 in all 24 grips (published lower bound).")


if __name__ == "__main__":
    main()
