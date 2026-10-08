#!/usr/bin/env python3
"""Gate a first-ever plain solve at 1 second and 10 MB of cache.

Run under taskpolicy -b nice -n 15 on macOS, after make build. The temporary
directory is inside this clone's .scratch, and is removed after verification.
"""

import os
from pathlib import Path
import re
import subprocess
import sys
import tempfile
import time

sys.dont_write_bytecode = True
import first_layer_oracle as physical


def main():
    root = Path(__file__).resolve().parents[1]
    scratch = root / ".scratch"
    scratch.mkdir(exist_ok=True)
    scramble = "R U F2 D' L B2 U R' F D2 L' B U2"
    with tempfile.TemporaryDirectory(prefix="cold-start-", dir=scratch) as cache:
        env = dict(os.environ, CUBE_CACHE_DIR=cache)
        env.pop("CUBE_LARGE_TABLES", None)
        assert not list(Path(cache).iterdir()), "cold cache must start empty"
        begun = time.perf_counter()
        result = subprocess.run([str(root / "dist" / "cube"), "solve", scramble],
                                env=env, capture_output=True, text=True, timeout=2)
        elapsed = time.perf_counter() - begun
        assert result.returncode == 0, result.stderr
        match = re.search(r"^Solution:\s*(.+)$", result.stdout, re.MULTILINE)
        assert match, result.stdout
        moves = match.group(1)
        assert 0 < len(moves.split()) <= 20, moves
        state = physical.physical_sequence(physical.fresh(), scramble + " " + moves)
        assert all(color == face[1][1] for face in state for row in face for color in row), moves
        files = [p for p in Path(cache).rglob("*") if p.is_file()]
        size = sum(p.stat().st_size for p in files)
        assert {p.name for p in files} <= {"coordinates-v5.bin.gz"}, files
        print(f"Cold first-ever plain solve: {elapsed * 1000:.2f} ms; "
              f"cache {size:,} bytes ({size / 1_000_000:.3f} MB); "
              f"{len(moves.split())} turns; budgets 1,000 ms / 10,000,000 bytes")
        assert elapsed <= 1.0, ("cold-start time budget exceeded", elapsed)
        assert size <= 10_000_000, ("default cache budget exceeded", size)


if __name__ == "__main__":
    main()
