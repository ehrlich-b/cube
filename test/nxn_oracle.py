#!/usr/bin/env python3
"""Independent NxN oracle: uniform legal states AND uniform-move scrambles.

Uses Python's stdlib and integer 3D geometry, never the Go move engine to create
or replay fixtures. Each solve runs in a fresh process. Uniform states in the
canonical frame shuffle
each center color orbit and each ordered wing orbit independently; corner twists
balance, and odd-cube midge flips/parity satisfy the ordinary 3x3 constraints.
Indistinguishable centers absorb labeled-center parity. A second batch draws
every legal single-layer turn uniformly to exercise actual scramble inputs.
"""

import argparse
from collections import defaultdict
from functools import lru_cache
import os
from pathlib import Path
import random
import re
import statistics
import subprocess
import time

COLORS = "BGORYW"
CFEN_ORDER = (4, 3, 0, 5, 2, 1)
OLL = "2R2 B2 U2 2L U2 2R' U2 2R U2 F2 2R F2 2L' B2 2R2"
PLL = "2R2 U2 2R2 Uw2 2R2 Uw2"


def parse_documented_metrics(text, name):
    rows = re.findall(
        r"^\| ([24567])×\1 \| (\d+\.\d+) / (\d+) \| "
        r"(\d+\.\d+) / (\d+) \| (\d+\.\d+) / (\d+\.\d+) ms \|$",
        text, re.M)
    assert len(rows) == 5, (name, "expected one measurement row for every size")
    metrics = {int(row[0]): tuple(float(value) for value in row[1:]) for row in rows}
    assert set(metrics) == {2, 4, 5, 6, 7}, (name, "missing or repeated size")
    return metrics


def documented_metrics():
    root = Path(__file__).resolve().parents[1]
    metrics = [parse_documented_metrics((root / name).read_text(), name)
               for name in ("README.md", "examples/solving.md")]
    assert metrics[0] == metrics[1], "README and solving guide measurements disagree"
    return metrics[0]


def check_move_mean(n, batch, lengths, documented):
    mean = statistics.mean(lengths)
    assert mean <= documented * 1.05 + 1e-9, (
        f"{n}x{n} {batch} mean moves regressed: {mean:.2f} exceeds "
        f"documented {documented:.2f} by more than 5%")


def quarter(vector, axis):
    x, y, z = vector
    return ((x, z, -y), (-z, y, x), (y, -x, z))[axis]


def determinant(a, b, c):
    return (a[0] * (b[1] * c[2] - b[2] * c[1])
            - a[1] * (b[0] * c[2] - b[2] * c[0])
            + a[2] * (b[0] * c[1] - b[1] * c[0]))


def parity(p):
    return sum(p[i] > p[j] for i in range(len(p))
               for j in range(i + 1, len(p))) % 2


class Geometry:
    def __init__(self, n):
        self.n = n
        last = n - 1
        self.addresses = []
        for f in range(6):
            for r in range(n):
                for c in range(n):
                    a, b = 2 * c - last, last - 2 * r
                    self.addresses.append((
                        ((a, b, last), (0, 0, 1)),
                        ((-a, b, -last), (0, 0, -1)),
                        ((-last, b, a), (-1, 0, 0)),
                        ((last, b, -a), (1, 0, 0)),
                        ((a, last, -b), (0, 1, 0)),
                        ((a, -last, b), (0, -1, 0)),
                    )[f])
        self.lookup = {p: i for i, p in enumerate(self.addresses)}
        self.home = [COLORS[i // (n * n)] for i in range(6 * n * n)]
        groups = defaultdict(list)
        for i, (position, normal) in enumerate(self.addresses):
            groups[position].append((normal, i))
        self.corners, self.midges, self.partners = [], [], {}
        for position, stickers in sorted(groups.items()):
            if len(stickers) == 3:
                stickers.sort(key=lambda s: -abs(s[0][1]))
                if determinant(*(s[0] for s in stickers)) != -1:
                    stickers[1], stickers[2] = stickers[2], stickers[1]
                self.corners.append([i for _, i in stickers])
            elif len(stickers) == 2:
                (normal_a, a), (normal_b, b) = stickers
                self.partners[a], self.partners[b] = b, a
                if 0 in position:
                    stickers.sort(key=lambda s: (-abs(s[0][1]), -abs(s[0][2])))
                    self.midges.append([i for _, i in stickers])
        assert len(self.corners) == 8
        assert len(self.midges) == (12 if n % 2 else 0)
        self.centers, self.wings = self.orbits()

    @lru_cache(maxsize=None)
    def permutation(self, token):
        match = re.fullmatch(r"([1-9][0-9]*)?([RULDFB]w?|[MESxyz])(['2]?)", token)
        assert match, ("invalid solution token", token)
        depth, base, modifier = match.groups()
        turns = {"": 1, "2": 2, "'": 3}[modifier]
        last = self.n - 1
        if base in "xyz":
            assert depth is None
            axis, sign = "xyz".index(base), 1
            selected = lambda p: True
        elif base in "MES":
            assert depth is None and self.n % 2, token
            axis, sign = {"M": (0, -1), "E": (1, -1), "S": (2, 1)}[base]
            selected = lambda p: p[axis] == 0
        else:
            axis, sign = {"R": (0, 1), "L": (0, -1), "U": (1, 1),
                          "D": (1, -1), "F": (2, 1), "B": (2, -1)}[base[0]]
            depth = int(depth or (2 if base.endswith("w") else 1))
            assert 1 <= depth <= self.n, token
            boundary = last - 2 * (depth - 1)
            if base.endswith("w"):
                selected = lambda p: sign * p[axis] >= boundary
            else:
                selected = lambda p: sign * p[axis] == boundary
        result = []
        for position, normal in self.addresses:
            if selected(position):
                for _ in range(sign * turns % 4):
                    position, normal = quarter(position, axis), quarter(normal, axis)
            result.append(self.lookup[position, normal])
        return result

    def replay(self, state, sequence):
        state = state[:]
        for token in sequence.split():
            after = state[:]
            for source, destination in enumerate(self.permutation(token)):
                after[destination] = state[source]
            state = after
        return state

    def orbits(self):
        generators = [self.permutation(f"{k}{f}")
                      for f in "RUF" for k in range(1, self.n + 1)]
        remaining = set(range(len(self.home)))
        centers, wings = [], []
        while remaining:
            root = min(remaining)
            queue, seen = [root], {root}
            for index in queue:
                for g in generators:
                    dst = g[index]
                    if dst not in seen:
                        seen.add(dst)
                        queue.append(dst)
            remaining -= seen
            position = self.addresses[root][0]
            boundaries = sum(abs(x) == self.n - 1 for x in position)
            if boundaries == 1 and len(queue) == 24:
                centers.append(queue)
            if boundaries == 2 and 0 not in position:
                # Process each paired sticker orbit once (one sticker/wing).
                if not any(self.partners[root] in o for o in wings):
                    assert len(queue) == 24
                    wings.append(queue)
        assert len(centers) == ((self.n - 2)**2 * 6 - (6 if self.n % 2 else 0)) // 24
        assert len(wings) == (self.n - 2) // 2
        return centers, wings

    def uniform_state(self, rng):
        state = self.home[:]
        corners = list(range(8))
        rng.shuffle(corners)
        twists = [rng.randrange(3) for _ in range(7)]
        twists.append(-sum(twists) % 3)
        for slot, source, twist in zip(self.corners, corners, twists):
            values = [self.home[i] for i in self.corners[source]]
            for i, dst in enumerate(slot):
                state[dst] = values[(i + twist) % 3]
        if self.midges:
            edges = list(range(12))
            rng.shuffle(edges)
            if parity(edges) != parity(corners):
                edges[0], edges[1] = edges[1], edges[0]
            flips = [rng.randrange(2) for _ in range(11)]
            flips.append(sum(flips) % 2)
            for slot, source, flip in zip(self.midges, edges, flips):
                values = [self.home[i] for i in self.midges[source]]
                for i, dst in enumerate(slot):
                    state[dst] = values[i ^ flip]
        for orbit in self.centers:
            values = [self.home[i] for i in orbit]
            rng.shuffle(values)
            for dst, color in zip(orbit, values):
                state[dst] = color
        for orbit in self.wings:
            permutation = orbit[:]
            rng.shuffle(permutation)
            for dst, source in zip(orbit, permutation):
                state[dst] = self.home[source]
                state[self.partners[dst]] = self.home[self.partners[source]]
        return state

    def scramble(self, rng):
        tokens = []
        # Uniform independent single-layer moves, including every inner/far
        # layer. This random walk is not claimed to be a uniform state sampler.
        for _ in range(40 * self.n):
            depth = rng.randrange(1, self.n + 1)
            tokens.append((str(depth) if depth > 1 else "")
                          + rng.choice("RULDFB") + rng.choice(("", "2", "'")))
        return " ".join(tokens)

    def cfen(self, state):
        area = self.n * self.n
        return "YB|" + "/".join("".join(state[f*area:(f+1)*area]) for f in CFEN_ORDER)

    def solved(self, state):
        # Stronger than face uniformity: every face matches the canonical
        # center frame, including odd fixed centers and even solved centers.
        return state == self.home

    def orientations(self):
        queue, seen = [(self.home, "")], {tuple(self.home)}
        for state, path in queue:
            for axis in "xyz":
                after = self.replay(state, axis)
                key = tuple(after)
                if key not in seen:
                    seen.add(key)
                    queue.append((after, (path + " " + axis).strip()))
        assert len(queue) == 24
        return [path for _, path in queue]


def run(binary, geometry, state, extra=(), headless=True):
    command = [str(binary), "solve", "--dimension", str(geometry.n),
               "--start", geometry.cfen(state), *extra]
    if headless:
        command.append("--headless")
    begun = time.perf_counter()
    result = subprocess.run(command, capture_output=True, text=True, timeout=10)
    duration = (time.perf_counter() - begun) * 1000
    assert result.returncode == 0, (geometry.n, geometry.cfen(state), result.stderr)
    sequence = result.stdout
    if not headless:
        sequence = re.search(r"^Solution: (.*)$", result.stdout, re.M)[1]
        steps = int(re.search(r"^Steps: (\d+)$", result.stdout, re.M)[1])
        assert steps == len(sequence.split()), (steps, sequence)
        assert f"Solving {geometry.n}x{geometry.n}x{geometry.n} cube" in result.stdout
    return sequence, duration


def parity_checks(binary):
    g = Geometry(4)
    for name, algorithm in (("OLL", OLL), ("PLL", PLL), ("both", OLL + " " + PLL)):
        state = g.replay(g.home, algorithm)
        for orbit in g.centers:
            assert all(state[i] == g.home[i] for i in orbit), name
        assert not g.solved(state), name
        sequence, _ = run(binary, g, state)
        assert g.solved(g.replay(state, sequence)), (name, sequence)
    print("isolated OLL, PLL and combined 4x4 parities independently replayed", flush=True)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--binary", type=Path, default=Path("dist/cube"))
    parser.add_argument("--cases", type=int, default=100)
    parser.add_argument("--sizes", type=int, nargs="+", default=[2, 4, 5, 6, 7])
    args = parser.parse_args()
    assert args.cases > 0
    binary = args.binary.resolve()
    scratch = Path(".scratch").resolve()
    scratch.mkdir(exist_ok=True)
    os.environ.setdefault("CUBE_CACHE_DIR", str(scratch / "cube-cache"))
    documented = documented_metrics()
    parity_checks(binary)
    for n in args.sizes:
        assert n in (2, 4, 5, 6, 7)
        g = Geometry(n)
        orientations = g.orientations()
        rng = random.Random(20261007 + n)
        lengths, durations = [], []
        for batch in ("uniform", "scramble"):
            batch_lengths, batch_durations = [], []
            for index in range(args.cases):
                state = (g.uniform_state(rng) if batch == "uniform"
                         else g.replay(g.home, g.scramble(rng)))
                # Cover saved states plus appended moves, rigid orientations,
                # wide turns, odd middle slices, and numbered far-layer turns.
                extra = ()
                if index < 24:
                    sequence = orientations[index]
                    extra = (sequence,) if sequence else ()
                sequence, duration = run(binary, g, state, extra)
                source = g.replay(state, extra[0]) if extra else state
                replay = g.replay(source, sequence)
                assert sequence.strip() or g.solved(source), (n, batch, index, "empty answer")
                assert g.solved(replay), (n, batch, index, g.cfen(source), sequence, g.cfen(replay))
                assert sorted(source) == sorted(replay), (n, "sticker conservation")
                lengths.append(len(sequence.split()))
                durations.append(duration)
                batch_lengths.append(len(sequence.split()))
                batch_durations.append(duration)
            print(f"{n}x{n} {batch}: {args.cases} fresh-process solutions replayed; "
                  f"moves mean {statistics.mean(batch_lengths):.2f}, max {max(batch_lengths)}; "
                  f"process ms mean {statistics.mean(batch_durations):.2f}, "
                  f"max {max(batch_durations):.2f}", flush=True)
            if batch == "uniform":
                check_move_mean(n, batch, batch_lengths, documented[n][2])
        # Full CLI output must agree with the solver contract; --cfen must
        # preserve the dimension and yield the exact center-matched state.
        state = g.uniform_state(rng)
        sequence, _ = run(binary, g, state, headless=False)
        assert g.solved(g.replay(state, sequence))
        result = subprocess.run([str(binary), "solve", "--dimension", str(n),
                                 "--start", g.cfen(state), "--cfen", "--headless"],
                                capture_output=True, text=True, timeout=10)
        expected = "YB|" + "/".join(COLORS[f] + str(n*n) for f in CFEN_ORDER)
        assert result.returncode == 0 and result.stdout == expected, (n, result)
        print(f"{n}x{n}: {len(lengths)} cases; moves mean {statistics.mean(lengths):.2f}, "
              f"max {max(lengths)}; process ms mean {statistics.mean(durations):.2f}, "
              f"max {max(durations):.2f}", flush=True)
        check_move_mean(n, "combined", lengths, documented[n][0])


if __name__ == "__main__":
    main()
