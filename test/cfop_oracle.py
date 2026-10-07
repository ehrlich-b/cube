#!/usr/bin/env python3
"""Construct 200 independent physical states and replay every CFOP checkpoint."""

import os
from pathlib import Path
import random
import re
import statistics
import subprocess
import sys
import time

sys.dont_write_bytecode = True
import first_layer_oracle as physical


def main():
    root = Path(__file__).resolve().parents[1]
    rng = random.Random(2026100713)
    names = ["Cross", "F2L 1", "F2L 2", "F2L 3", "F2L 4", "OLL", "PLL"]
    lengths = {name: [] for name in names + ["Total"]}
    durations = []
    env = dict(os.environ)
    env.setdefault("CUBE_CACHE_DIR", str(root / ".scratch" / "cube-cache"))
    orientations = physical.orientation_paths()
    # Derive paired slots from independent 3D adjacency and home colors.
    slots = []
    for corner in physical.CORNERS:
        colors = {physical.color(physical.SOLVED, address) for address in corner}
        if "W" in colors:
            edge = next(edge for edge in physical.EDGES
                        if {physical.color(physical.SOLVED, address) for address in edge}
                        == colors - {"W"})
            slots.append(corner + edge)
    assert len(slots) == 4
    for index in range(200):
        state = physical.legal_fixture(rng)
        if index < 24:
            state = physical.physical_sequence(state, orientations[index])
        begun = time.perf_counter()
        result = subprocess.run([str(root / "dist" / "cube"), "solve", "--method", "cfop",
                                 "--start", physical.cfen(state)],
                                text=True, capture_output=True, timeout=10, env=env)
        durations.append((time.perf_counter() - begun) * 1000)
        assert result.returncode == 0, (index, physical.cfen(state), result.stderr)
        solution = re.search(r"^Solution: (.*)$", result.stdout, re.M).group(1)
        stages = re.findall(r"^(Cross|F2L [1-4]|OLL|PLL): (.*) \((\d+) turns\)\n  ([^\n]*)",
                            result.stdout, re.M)
        assert [stage[0] for stage in stages] == names, (index, result.stdout)
        assert " ".join(move for stage in stages for move in stage[3].split()) == solution
        protected = set()
        for name, case, count, moves in stages:
            assert case, (index, name)
            turns = sum(token[0] not in "xyz" for token in moves.split())
            assert turns == int(count), (index, name, count, moves)
            lengths[name].append(turns)
            state = physical.physical_sequence(state, moves)
            # The physical oracle stores F/B/L/R/U/D and derives all moves
            # from 3D geometry, without the Go engine or coordinate tables.
            assert all(state[5][r][c] == "W" for r, c in [(0, 1), (1, 0), (1, 2), (2, 1)])
            assert all(state[f][2][1] == state[f][1][1] for f in [0, 1, 2, 3])
            solved = set()
            for slot, coords in enumerate(slots):
                if all(state[f][r][c] == state[f][1][1] for f, r, c in coords):
                    solved.add(slot)
            assert protected <= solved, (index, name, protected, solved)
            protected = solved
            if name in ["F2L 4", "OLL", "PLL"]:
                assert len(solved) == 4, (index, name)
            if name in ["OLL", "PLL"]:
                assert all(color == "Y" for row in state[4] for color in row), (index, name)
        assert all(color == face[1][1] for face in state for row in face for color in row), (
            index, physical.cfen(state), solution)
        lengths["Total"].append(sum(token[0] not in "xyz" for token in solution.split()))
    for name, counts in lengths.items():
        print(f"{name}: mean {statistics.mean(counts):.3f}, max {max(counts)} turns")
    print(f"200 independent uniform physical states and every checkpoint solved; "
          f"process mean {statistics.mean(durations):.2f} ms, max {max(durations):.2f} ms; "
          f"total {sum(durations) / 1000:.2f} s (each process includes table setup)")


if __name__ == "__main__":
    main()
