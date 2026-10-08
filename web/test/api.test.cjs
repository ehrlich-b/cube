const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const vm = require("node:vm");
const { webcrypto } = require("node:crypto");

globalThis.crypto ??= webcrypto;
vm.runInThisContext(fs.readFileSync(path.join(__dirname, "../wasm_exec.js"), "utf8"));

function call(request) {
  const response = JSON.parse(globalThis.cubeAPI(JSON.stringify(request)));
  assert.equal(response.ok, true, response.error);
  return response.data;
}

function rejects(request, pattern) {
  const response = JSON.parse(globalThis.cubeAPI(JSON.stringify(request)));
  assert.equal(response.ok, false);
  assert.match(response.error, pattern);
}

(async () => {
  const go = new Go();
  const { instance } = await WebAssembly.instantiate(fs.readFileSync(path.join(__dirname, "../cube.wasm")), go.importObject);
  go.run(instance).catch(error => { console.error(error); process.exit(1); });
  const solved = call({ op: "state" }).state;
  assert.equal(solved.cfen, "YB|Y9/R9/B9/W9/O9/G9");
  assert.equal(solved.solved, true);
  assert.deepEqual(solved.faces.F, Array(9).fill("B"));

  const inverse = move => move.endsWith("2") ? move : move.endsWith("'") ? move.slice(0, -1) : move + "'";
  for (const token of ["R", "L", "U", "D", "F", "B", "M", "E", "S", "x", "y", "z", "Rw", "Lw", "Uw", "Dw", "Fw", "Bw"]) {
    for (const suffix of ["", "'", "2"]) {
      const move = token + suffix;
      const changed = call({ op: "twist", cfen: solved.cfen, moves: move });
      assert.deepEqual(changed.moves, [move]);
      const restored = call({ op: "twist", cfen: changed.state.cfen, moves: inverse(move) }).state;
      assert.equal(restored.cfen, solved.cfen, `${move} inverse`);
      assert.deepEqual(call({ op: "state", cfen: changed.state.cfen }).state, changed.state);
    }
  }
  console.log("PASS twist: faces, slices, rotations, wide turns, inverses and CFEN round trips");

  for (const size of [2, 3, 4, 5, 6, 7]) {
    const home = call({ op: "state", size }).state;
    assert.equal(home.size, size);
    assert.equal(home.solved, true);
    for (const face of Object.values(home.faces)) assert.equal(face.length, size * size);
    const tokens = ["R", "L", "U", "D", "F", "B", "x", "y", "z", "Rw", "Uw", "Fw"];
    for (let layer = 2; layer <= size; layer++) for (const face of "RLUDFB") tokens.push(`${layer}${face}`, `${layer}${face}w`);
    if (size % 2) tokens.push("M", "E", "S");
    for (const token of tokens) for (const suffix of ["", "'", "2"]) {
      const move = token + suffix;
      const next = call({ op: "twist", size, cfen: home.cfen, moves: move }).state;
      assert.deepEqual(call({ op: "state", cfen: next.cfen }).state, next);
      assert.equal(call({ op: "twist", cfen: next.cfen, moves: inverse(move) }).state.cfen, home.cfen, `${size}x${size} ${move} inverse`);
    }
    const scramble = size > 3 ? "2R U 2F' Rw D2 L B'" : "R U F2 L' B";
    const start = call({ op: "twist", size, moves: scramble }).state;
    const solution = call({ op: "solve", cfen: start.cfen });
    assert.equal(solution.method, size === 3 ? "kociemba" : "reduction");
    assert.ok(solution.moves.length > 0);
    assert.equal(solution.state.solved, true);
    assert.equal(call({ op: "twist", cfen: start.cfen, moves: solution.moves.join(" ") }).state.solved, true);
    assert.deepEqual(call({ op: "state", cfen: start.cfen }).state, start, "solver preserves the input");
    if (size !== 3) {
      rejects({ op: "solve", size, method: "cfop" }, /3x3-only/);
      rejects({ op: "learn", size }, /3x3-only/);
      rejects({ op: "find", size, target: home.cfen }, /3x3-only/);
    }
    rejects({ op: "twist", size, moves: `${size + 1}R` }, /invalid|outside/);
    if (size % 2 === 0) rejects({ op: "twist", size, moves: "M" }, /even/);
    console.log(`PASS ${size}x${size}: every layer, wide depth, inverse, CFEN and solution replay (${solution.moves.length} moves, ${solution.solveMs.toFixed(1)} ms)`);
  }

  for (const scramble of ["R U R' U'", "R U F2 L' B", "x y R U F2 L' B D2 R2 U' F L2 B'", "R2 U F' D B2 L' U2 F R' D2 L B' U R2 F2 D' L2 U' B R"]) {
    const start = call({ op: "twist", moves: scramble }).state;
    const solution = call({ op: "solve", cfen: start.cfen });
    assert.equal(solution.method, "kociemba");
    assert.ok(solution.moves.length > 0);
    assert.equal(solution.state.solved, true);
    assert.equal(call({ op: "twist", cfen: start.cfen, moves: solution.moves.join(" ") }).state.solved, true);
    assert.equal(call({ op: "state", cfen: start.cfen }).state.cfen, start.cfen);
  }
  assert.equal(call({ op: "solve" }).state.solved, true);
  // Superflip exercises the default search budget and the incumbent returned
  // when a 20-turn target is not reached in time. Setup is separate from it.
  const superflip = "YB|YGYOYRYBY/RYRBRGRWR/BYBOBRBWB/WBWOWRWGW/OYOGOBOWO/GYGRGOGWG";
  const budgetStarted = performance.now();
  const budgetSolution = call({ op: "solve", cfen: superflip });
  assert.ok(performance.now() - budgetStarted < 3000, "warm WASM solve must finish near its one-second budget");
  assert.ok(budgetSolution.moves.length > 0 && budgetSolution.moves.length <= 30);
  assert.equal(call({ op: "twist", cfen: superflip, moves: budgetSolution.moves.join(" ") }).state.solved, true);
  for (const method of ["auto", "kociemba", "beginner", "cfop"]) {
    const start = call({ op: "twist", moves: "R U F2 L' B" }).state;
    const solution = call({ op: "solve", cfen: start.cfen, method });
    assert.equal(solution.method, method === "auto" ? "kociemba" : method);
    assert.ok(solution.moves.length > 0);
    assert.equal(solution.state.solved, true);
    assert.equal(call({ op: "twist", cfen: start.cfen, moves: solution.moves.join(" ") }).state.solved, true);
  }
  for (const scramble of ["R U R' U'", "R U F2 L' B", "x y R U F2 L' B D2 R2 U' F L2 B'", "R2 U F' D B2 L' U2 F R' D2 L B' U R2 F2 D' L2 U' B R"]) {
    const start = call({ op: "twist", moves: scramble }).state;
    const solution = call({ op: "solve", cfen: start.cfen, method: "cfop" });
    assert.deepEqual(solution.stages.map(stage => stage.name), ["Cross", "F2L 1", "F2L 2", "F2L 3", "F2L 4", "OLL", "PLL"]);
    assert.deepEqual(solution.stages.flatMap(stage => stage.moves), solution.moves);
    let current = start;
    for (const stage of solution.stages) {
      assert.ok(stage.cases.length > 0);
      assert.equal(stage.turns, stage.moves.filter(move => !/^[xyz]/.test(move)).length);
      current = call({ op: "twist", cfen: current.cfen, moves: stage.moves.join(" ") }).state;
      assert.equal(current.cfen, stage.after.cfen, stage.name);
      // The complete white cross stays solved at every checkpoint.
      for (const i of [1, 3, 5, 7]) assert.equal(current.faces.D[i], "W");
      for (const face of ["F", "R", "B", "L"]) assert.equal(current.faces[face][7], current.faces[face][4]);
      if (["F2L 4", "OLL", "PLL"].includes(stage.name)) {
        assert.deepEqual(current.faces.D, Array(9).fill("W"));
        for (const face of ["F", "R", "B", "L"]) assert.deepEqual(current.faces[face].slice(3), Array(6).fill(current.faces[face][4]));
      }
      if (["OLL", "PLL"].includes(stage.name)) assert.deepEqual(current.faces.U, Array(9).fill("Y"));
    }
    assert.equal(current.solved, true);
  }
  const cfopSkip = call({ op: "solve", method: "cfop" });
  assert.equal(cfopSkip.stages.length, 7);
  assert.deepEqual(cfopSkip.moves, []);
  console.log("PASS CFOP: grouped stage metadata, case names, checkpoint replay, rotated inputs and skips");
  console.log("PASS solve: Kociemba default/auto, all method selections, replay and full solution invariant");

  const start = call({ op: "twist", moves: "R U F2 L' B" }).state;
  const lesson = call({ op: "learn", cfen: start.cfen });
  assert.ok(lesson.steps.length > 1);
  let current = start;
  for (const step of lesson.steps) {
    assert.ok(step.title && step.check);
    assert.deepEqual(step.actions.flatMap(action => action.moves), step.moves);
    current = call({ op: "twist", cfen: current.cfen, moves: step.moves.join(" ") }).state;
    assert.equal(current.cfen, step.after.cfen, step.title);
  }
  assert.equal(current.solved, true);
  // Replanning follows actual moves, including recovery from a wrong turn.
  current = call({ op: "twist", cfen: start.cfen, moves: "x R'" }).state;
  const recovery = call({ op: "learn", cfen: current.cfen });
  assert.equal(call({ op: "twist", cfen: current.cfen, moves: recovery.moves.join(" ") }).state.solved, true);
  console.log("PASS learn: replay every checkpoint and replan after actual moves");

  const near = call({ op: "twist", moves: "R U" }).state;
  const result = call({ op: "find", cfen: near.cfen, target: solved.cfen, maxDepth: 2 });
  assert.equal(result.found, true);
  assert.equal(result.moves.length, 2);
  assert.equal(call({ op: "twist", cfen: near.cfen, moves: result.moves.join(" ") }).state.solved, true);
  assert.equal(call({ op: "find", cfen: near.cfen, target: solved.cfen, maxDepth: 1 }).found, false);
  assert.deepEqual(call({ op: "find", target: "YB|?9/?9/?9/?9/?9/?9", maxDepth: 0 }).moves, []);
  // This concrete target is intentionally unsatisfiable at depth zero.
  assert.equal(call({ op: "find", target: near.cfen, maxDepth: 0 }).found, false);
  const target = call({ op: "twist", moves: "F" }).state.cfen;
  assert.equal(call({ op: "find", target, maxDepth: 1 }).state.cfen, target);
  const wildcard = "YB|Y9/?9/?9/?9/?9/?9";
  const partial = call({ op: "find", cfen: target, target: wildcard, maxDepth: 1 });
  assert.equal(partial.found, true);
  assert.equal(partial.moves.length, 1);
  assert.deepEqual(call({ op: "twist", cfen: target, moves: partial.moves.join(" ") }).state.faces.U, Array(9).fill("Y"));
  assert.equal(call({ op: "find", cfen: target, target: wildcard, maxDepth: 0 }).found, false);
  // Same proven ten-move fixture as the CLI's e2e test; depth nine cannot reach it.
  const deep = call({ op: "twist", moves: "R U F2 L' B D2 R' F U2 L" }).state;
  const deepResult = call({ op: "find", cfen: deep.cfen, target: solved.cfen, maxDepth: 10 });
  assert.equal(deepResult.found, true);
  assert.equal(deepResult.moves.length, 10);
  assert.equal(call({ op: "twist", cfen: deep.cfen, moves: deepResult.moves.join(" ") }).state.cfen, solved.cfen);
  const exhausted = call({ op: "find", cfen: deep.cfen, target: solved.cfen, maxDepth: 9 });
  assert.equal(exhausted.found, false);
  assert.deepEqual(exhausted.moves, []);
  assert.equal(exhausted.state.cfen, deep.cfen);
  assert.equal(call({ op: "state", cfen: deep.cfen }).state.cfen, deep.cfen);
  // Face turns keep centers fixed, even for otherwise unrestricted wildcards.
  assert.equal(call({ op: "find", target: "YB|W9/?9/?9/?9/?9/?9", maxDepth: 10 }).found, false);
  console.log("PASS find: exact/wildcard shortest paths through depth ten, depth exhaustion, fixed centers, no mutation");

  rejects({ op: "twist", moves: "R garbage U" }, /invalid/);
  rejects({ op: "twist", moves: "999Rw" }, /invalid/);
  rejects({ op: "twist", moves: "R ".repeat(32769) }, /exceeds/);
  rejects({ op: "state", cfen: "YB|Y999999999/R9/B9/W9/O9/G9" }, /runs 1–49/);
  rejects({ op: "state", cfen: "YB|Y99/R49/B49/W49/O49/G49" }, /at most 49/);
  rejects({ op: "state", cfen: "YB|" + Array(6).fill("Y49Y49").join("/") }, /at most 49/);
  rejects({ op: "state", size: 8 }, /between 2 and 7/);
  rejects({ op: "state", size: 4, cfen: solved.cfen }, /does not match/);
  rejects({ op: "state", cfen: "YB|?9/R9/B9/W9/O9/G9" }, /only allowed/);
  rejects({ op: "state", cfen: "YB|W9/R9/B9/W9/O9/G9" }, /color|sticker|center/i);
  rejects({ op: "state", cfen: "WB|W9/R9/B9/Y9/O9/G9" }, /YB/);
  rejects({ op: "find", target: solved.cfen, maxDepth: 11 }, /depth/);
  rejects({ op: "find", target: solved.cfen, maxDepth: -1 }, /depth/);
  rejects({ op: "find", target: "YB|?4/?4/?4/?4/?4/?4", maxDepth: 0 }, /3x3/);
  rejects({ op: "solve", moves: "R", method: "missing" }, /unknown/);
  rejects({ op: "missing" }, /unknown/);
  assert.equal(JSON.parse(globalThis.cubeAPI("{")).ok, false);
  assert.equal(JSON.parse(globalThis.cubeAPI()).ok, false);
  assert.equal(JSON.parse(globalThis.cubeAPI(42)).ok, false);
  console.log("PASS API bounds: invalid moves, unsafe runs, impossible states and unknown solvers");
  process.exit(0);
})().catch(error => { console.error(error); process.exit(1); });
