#!/usr/bin/env python3
"""Independently construct 200 physical states and replay Kociemba solutions.

No Go move engine or cubie coordinates are used as the replay oracle.
Run after building: python3 test/kociemba_oracle.py
"""

import argparse
from collections import Counter
import os
from pathlib import Path
import random
import statistics
import subprocess
import sys
import time

sys.dont_write_bytecode = True
import first_layer_oracle as physical


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--cases", type=int, default=200)
    parser.add_argument("--binary", type=Path,
                        default=Path(__file__).resolve().parents[1] / "dist" / "cube")
    args = parser.parse_args()
    if args.cases < 200:
        parser.error("at least 200 cases are required")
    rng = random.Random(2026100706)
    lengths, durations = [], []
    env = dict(os.environ)
    env.setdefault("CUBE_CACHE_DIR", str(Path(__file__).resolve().parents[1]
                                        / ".scratch" / "cube-cache"))
    orientations = physical.orientation_paths()
    for index in range(args.cases):
        state = physical.legal_fixture(rng)
        # Most cases use the canonical grip so solution lengths are face turns.
        # The remaining cases cover every rigid orientation; grip rotations
        # precede the face-turn solution and are counted separately.
        if index < 24:
            state = physical.physical_sequence(state, orientations[index])
        begun = time.perf_counter()
        result = subprocess.run([str(args.binary.resolve()), "solve", "--method", "kociemba",
                                 "--start", physical.cfen(state), "--headless"],
                                text=True, capture_output=True, timeout=3, env=env)
        duration = time.perf_counter() - begun
        assert result.returncode == 0, (index, physical.cfen(state), result.stderr)
        tokens = result.stdout.split()
        turns = [token for token in tokens if token[0] not in "xyz"]
        assert 0 < len(turns) <= 20, (index, result.stdout)
        replay = physical.physical_sequence(state, result.stdout)
        assert all(color == face[1][1] for face in replay for row in face for color in row), (
            index, physical.cfen(state), result.stdout, physical.cfen(replay))
        lengths.append(len(turns))
        durations.append(duration * 1000)
    mean = statistics.mean(lengths)
    at_target = sum(length <= 20 for length in lengths) / args.cases
    assert mean <= 20, ("mean length regressed", mean, Counter(lengths))
    assert at_target == 1, ("<=20 share regressed", at_target, Counter(lengths))
    print(f"{args.cases} independent uniform physical states solved; "
          f"mean {statistics.mean(lengths):.3f}, max {max(lengths)} face turns; "
          f"<=20 {at_target:.1%}; "
          f"process mean {statistics.mean(durations):.2f} ms, "
          f"max {max(durations):.2f} ms; "
          f"histogram {dict(sorted(Counter(lengths).items()))}")


if __name__ == "__main__":
    main()
