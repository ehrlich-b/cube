#!/usr/bin/env python3
"""Verify full beginner solutions and every printed checkpoint independently.

Run after building: python3 test/full_lesson_oracle.py
Uses only Python's standard library and the adjacent independent sticker oracle.
Physical moves are simulated; this is not evidence of a human physical-cube trial.
"""

import argparse
import itertools
import json
from pathlib import Path
import random
import re
import shlex
import statistics
import subprocess
import sys
import threading
import time

# Keep this optional standalone check from adding generated files to the repo.
sys.dont_write_bytecode = True
import first_layer_oracle as physical

TOP_EDGES = [g for g in physical.EDGES if any(p[0] == 4 for p in g)]
TOP_CORNERS = [g for g in physical.CORNERS if any(p[0] == 4 for p in g)]
BELT_EDGES = [g for g in physical.EDGES if all(p[0] not in (4, 5) for p in g)]



def run(binary, args, text=None):
    # The imported first-layer runner intentionally fixes its own goal. Keep
    # full-goal subprocess calls here so the two test contracts stay separate.
    return subprocess.run([str(binary)] + args, input=text, text=True,
                          capture_output=True, timeout=3)

def solved(state):
    return all(value == face[1][1] for face in state for row in face for value in row)


def two_layers(state):
    return (physical.white_layer(state)
            and all(state[face][1][col] == state[face][1][1]
                    for face in range(4) for col in range(3)))


def edge_fixture(groups, permutation, flips):
    state = physical.fresh()
    for slot, source, flip in zip(groups, permutation, flips):
        values = [physical.color(physical.SOLVED, p) for p in groups[source]]
        for i, address in enumerate(slot):
            physical.put(state, address, values[i ^ flip])
    if physical.parity(permutation):
        # Match edge parity using two upper corners; preserve their handedness.
        a, b = TOP_CORNERS[:2]
        va = [physical.color(state, p) for p in a]
        vb = [physical.color(state, p) for p in b]
        for p, value in zip(a, vb):
            physical.put(state, p, value)
        for p, value in zip(b, va):
            physical.put(state, p, value)
    return state


def corner_fixture(permutation, twists):
    state = physical.fresh()
    for slot, source, twist in zip(TOP_CORNERS, permutation, twists):
        values = [physical.color(physical.SOLVED, p) for p in TOP_CORNERS[source]]
        for i, address in enumerate(slot):
            physical.put(state, address, values[(i + twist) % 3])
    return state


def fixtures(count):
    rng, orientations = random.Random(20261004), physical.orientation_paths()
    for i in range(count):
        yield "uniform", physical.physical_sequence(
            physical.legal_fixture(rng), orientations[i % 24])
    for mask in range(16):
        if mask.bit_count() % 2 == 0:
            yield "yellow_cross", edge_fixture(
                TOP_EDGES, range(4), [(mask >> i) & 1 for i in range(4)])
    for permutation in itertools.permutations(range(4)):
        yield "yellow_edge_permutation", edge_fixture(TOP_EDGES, permutation, [0] * 4)
    for permutation in itertools.permutations(range(4)):
        if physical.parity(permutation) == 0:
            yield "yellow_corner_permutation", corner_fixture(permutation, [0] * 4)
    for a, b, c in itertools.product(range(3), repeat=3):
        yield "yellow_corner_twists", corner_fixture(range(4), [a, b, c, (-a-b-c) % 3])
    for slot in range(4):
        for flip in range(2):
            yield "middle_edge_from_top", edge_fixture(
                [BELT_EDGES[slot], TOP_EDGES[0]], [1, 0], [flip, flip])
    for a, b in itertools.combinations(range(4), 2):
        for flip in range(2):
            yield "misplaced_middle_edges", edge_fixture(
                [BELT_EDGES[a], BELT_EDGES[b]], [1, 0], [flip, flip])


def verify_printed_lesson(output, initial):
    state, moves, checkpoints = initial, [], 0
    pending, title = False, ""
    for line in output.splitlines():
        if line.startswith("Checkpoint "):
            assert not pending, "Checkpoint has no resulting CFEN"
            pending, moves = True, []
            title = line.lower()
        elif line.startswith("  Do: "):
            assert pending, "Move group is outside a checkpoint"
            moves.append(line[len("  Do: "):])
        elif line.startswith("After this checkpoint: "):
            assert pending, "Resulting CFEN is outside a checkpoint"
            state = physical.physical_sequence(state, " ".join(moves))
            expected = physical.decode(line[len("After this checkpoint: "):])
            assert state == expected, (checkpoints, physical.cfen(state), line)
            if "middle edge" in title:
                assert physical.white_layer(state), title
            if "yellow" in title:
                assert two_layers(state), title
            if "make the yellow cross" in title:
                assert all(physical.color(state, group[0]) == "Y" for group in TOP_EDGES), title
            if "match the yellow cross" in title:
                assert all(physical.color(state, p) == state[p[0]][1][1]
                           for group in TOP_EDGES for p in group), title
            if "yellow corners in their home positions" in title:
                assert all(sorted(physical.color(state, p) for p in group)
                           == sorted(state[p[0]][1][1] for p in group)
                           for group in TOP_CORNERS), title
            pending, checkpoints = False, checkpoints + 1
    assert not pending, "Unfinished checkpoint"
    saved = re.findall(r"^Saved state: (YB\|[^\n]+)$", output, re.MULTILINE)
    assert len(saved) == 1 and state == physical.decode(saved[0]), output
    assert solved(state), physical.cfen(state)
    assert "this lesson stops at the first layer" not in output, output
    return checkpoints


def check_solution(binary, state):
    start = physical.cfen(state)
    begun = time.perf_counter()
    result = run(binary, ["solve", "", "--algorithm", "beginner",
                                  "--goal", "full", "--start", start, "--cfen"])
    duration = time.perf_counter() - begun
    assert result.returncode == 0, result.stderr
    final = physical.decode(result.stdout)
    assert solved(final), result.stdout
    result = run(binary, ["solve", "", "--algorithm", "beginner",
                                  "--goal", "full", "--start", start, "--headless"])
    assert result.returncode == 0, result.stderr
    replay = physical.physical_sequence(state, result.stdout)
    move_count = len(result.stdout.split())
    assert replay == final, (result.stdout, physical.cfen(replay))
    result = run(binary, ["learn", "--goal", "full", "--start", start])
    assert result.returncode == 0, result.stderr
    checkpoints = verify_printed_lesson(result.stdout, state)
    return duration, checkpoints, move_count


def check_fixture_solutions(binary, count):
    counts, durations, move_counts, checkpoints = {}, [], [], 0
    for index, (kind, state) in enumerate(fixtures(count)):
        if kind.startswith("yellow"):
            # Odd edge permutations were balanced using only TOP corners:
            # direct last-layer fixtures must leave the entire F2L intact.
            assert two_layers(state), (kind, physical.cfen(state))
        elif "middle" in kind:
            assert physical.white_layer(state), (kind, physical.cfen(state))
        try:
            duration, steps, moves = check_solution(binary, state)
        except AssertionError as error:
            raise AssertionError((index, kind, physical.cfen(state), str(error))) from error
        counts[kind] = counts.get(kind, 0) + 1
        if kind == "uniform":
            durations.append(duration)
            move_counts.append(moves)
        checkpoints += steps
    return {"fixture_counts": counts, "printed_checkpoints_replayed": checkpoints,
            "uniform_process_latency_ms": {
                "median": round(statistics.median(durations) * 1000, 2),
                "max": round(max(durations) * 1000, 2)},
            "uniform_solution_moves": {
                "mean": round(statistics.mean(move_counts), 1), "max": max(move_counts)}}


def prompt_state(output):
    match = re.search(r"Saved state: (YB\|[^\n]+)", output)
    assert match, output
    return physical.decode(match.group(1))


def next_title(output):
    match = re.search(r"Next checkpoint: ([^\n]+)", output)
    return match.group(1) if match else None


def do_moves(output):
    return " ".join(re.findall(r"  Do: ([^\n]+)", output))


def check_resume(binary, state):
    result = run(binary, ["learn", "--goal", "full", "--start", physical.cfen(state)])
    assert result.returncode == 0, result.stderr
    return verify_printed_lesson(result.stdout, state)


def check_returned_resume_commands(binary):
    state = physical.physical_sequence(physical.fresh(), "R U F2 L' B")
    checked = 0
    for goal in ["full", "first-layer"]:
        for termination in ["quit\n", "quit", ""]:
            result = run(binary, ["learn", "--goal", goal, "--start", physical.cfen(state),
                                  "--interactive"], termination)
            assert result.returncode == 0, result.stderr
            match = re.search(r"Resume: (cube learn[^\n]+)", result.stdout)
            assert match, result.stdout
            command = shlex.split(match.group(1))
            assert command[0] == "cube" and command[1] == "learn", command
            assert command[command.index("--goal") + 1] == goal, command
            resumed = run(binary, command[1:], "quit\n")
            assert resumed.returncode == 0, resumed.stderr
            expected_intro = ("White first-layer lesson:" if goal == "first-layer"
                              else "Complete beginner 3x3 lesson:")
            assert expected_intro in resumed.stdout, resumed.stdout
            assert all(current == state for current in physical.displayed_states(resumed.stdout))
            checked += 1
    return checked


def check_invalid_full_inputs(binary):
    impossible = edge_fixture(TOP_EDGES, range(4), [1, 0, 0, 0])
    wrong_centers = physical.fresh()
    wrong_centers[0], wrong_centers[3] = wrong_centers[3], wrong_centers[0]
    invalid = [
        ["solve", "R3", "--headless"], ["solve", "R", "--dimension", "8"],
        ["solve", "R", "--algorithm", "missing"], ["solve", "R", "--method", "invalid"],
        ["solve", "R", "--goal", "typo"], ["learn", "R", "--goal", "typo"],
        ["solve", "--start", "YB|Y999999999/R9/B9/W9/O9/G9", "--cfen"],
        ["solve", "--start", "YB|?9/R9/B9/W9/O9/G9", "--cfen"],
        ["solve", "--start", physical.cfen(impossible), "--cfen"],
        # Every face is uniform, but this center arrangement is not a rigid
        # orientation of the standard scheme. IsSolved alone would accept it.
        ["solve", "--start", physical.cfen(wrong_centers), "--cfen"],
        ["solve", "R " * 5000, "--headless"],
    ]
    for args in invalid:
        result = run(binary, args)
        assert result.returncode != 0 and not result.stdout, (args, result.stdout)
        assert result.stderr.count("Error:") == 1, result.stderr
    result = run(binary, ["solve", "R", "--dimension", "8"])
    assert "first-layer" not in result.stderr, result.stderr
    return len(invalid)


def check_solved_grips(binary):
    checked = 0
    for grip in ["", "x", "y'", "z2 x"]:
        state = physical.physical_sequence(physical.fresh(), grip)
        start = physical.cfen(state)
        result = run(binary, ["solve", "--start", start, "--headless"])
        assert result.returncode == 0 and result.stdout == "", result
        result = run(binary, ["solve", "--start", start, "--cfen"])
        assert result.returncode == 0 and physical.decode(result.stdout) == state
        result = run(binary, ["learn", "--start", start, "--interactive"],
                     "next\nnext\nnext\nstate\nquit\n")
        assert result.returncode == 0, result.stderr
        assert result.stdout.count("no moves applied") == 3, result.stdout
        assert "Next checkpoint:" not in result.stdout, result.stdout
        assert all(current == state for current in physical.displayed_states(result.stdout))
        checked += 1
    return checked


def check_live_recovery(binary):
    initial = physical.physical_sequence(physical.legal_fixture(random.Random(4723)), "z x2")
    process = subprocess.Popen(
        [str(binary), "learn", "--goal", "full", "--start", physical.cfen(initial), "--interactive"],
        stdin=subprocess.PIPE, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, text=True)
    deadline = threading.Timer(20, process.kill)
    deadline.daemon = True
    deadline.start()
    state, checkpoint_count, resume_count = initial, 0, 0
    interrupted_last_layer, deviated_last_layer = False, False
    try:
        output = physical.read_prompt(process)
        assert prompt_state(output) == state
        while next_title(output) is not None:
            title = next_title(output).lower()
            moves = do_moves(output)
            assert moves, output
            if "yellow" in title and not deviated_last_layer:
                before = state
                state = physical.physical_sequence(state, "x y' U R")
                output, state = physical.checked_command(process, "moves x y' U R", state)
                resume_count += check_resume(binary, state)
                output, state = physical.checked_command(process, "undo", state,
                                                        "Undo on your physical cube")
                assert state == before
                moves = do_moves(output)
                assert moves, output
                deviated_last_layer = True
            # Interrupt the complete final orientation sweep after only a prefix.
            # Actual-move recording must recover even if the lower layers changed.
            if ("corner" in title and any(word in title for word in ("orient", "twist", "upright"))
                    and not interrupted_last_layer):
                prefix = " ".join(moves.split()[:4])
                state = physical.physical_sequence(state, prefix)
                output, state = physical.checked_command(process, "moves " + prefix, state)
                resume_count += check_resume(binary, state)
                output, state = physical.checked_command(process, "undo", state,
                                                        "Undo on your physical cube")
                interrupted_last_layer = True
                moves = do_moves(output)
                assert moves, output
            before = state
            state = physical.physical_sequence(state, moves)
            output, state = physical.checked_command(process, "next", state)
            checkpoint_count += 1
            assert checkpoint_count < 30
            # Replay physical inverses through later checkpoints, then repeat next.
            if "yellow" in title:
                output, state = physical.checked_command(process, "undo", state,
                                                        "Undo on your physical cube")
                assert state == before
                state = physical.physical_sequence(state, do_moves(output))
                output, state = physical.checked_command(process, "next", state)
        assert solved(state), physical.cfen(state)
        assert interrupted_last_layer, "Fixture failed to exercise final corner orientation"
        assert deviated_last_layer, "Fixture failed to exercise changed grip during last layer"
        for _ in range(3):
            output, state = physical.checked_command(process, "next", state)
            assert "no moves applied" in output
        output, state = physical.checked_command(process, "reset", state,
                                                "Reset on your physical cube")
        assert state == initial
        process.stdin.write("quit\n")
        process.stdin.flush()
        process.communicate(timeout=3)
        assert process.returncode == 0
    finally:
        deadline.cancel()
        if process.poll() is None:
            process.kill()
            process.communicate()
    return {"confirmed_checkpoints": checkpoint_count,
            "resumed_checkpoints_replayed": resume_count,
            "partial_final_sweep_record_undo_reset": "passed",
            "last_layer_grip_change_and_wrong_turn": "passed"}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--binary", type=Path,
                        default=Path(__file__).resolve().parents[1] / "dist" / "cube")
    parser.add_argument("--cases", type=int, default=120)
    args = parser.parse_args()
    if args.cases < 1:
        parser.error("--cases must be positive")
    binary = args.binary.resolve()
    if not binary.is_file():
        parser.error("Cube binary missing; run make build or pass --binary")
    results = check_fixture_solutions(binary, args.cases)
    results["invalid_full_inputs"] = check_invalid_full_inputs(binary)
    results["returned_resume_commands_replayed"] = check_returned_resume_commands(binary)
    results["solved_grips_preserved"] = check_solved_grips(binary)
    results["live_recovery"] = check_live_recovery(binary)
    results["ux_evidence"] = "Independent simulated physical moves; no human physical trial"
    print(json.dumps(results, indent=2))


if __name__ == "__main__":
    main()
