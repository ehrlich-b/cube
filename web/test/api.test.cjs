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

  for (const scramble of ["R U R' U'", "R U F2 L' B", "x y R U F2 L' B D2 R2 U' F L2 B'", "R2 U F' D B2 L' U2 F R' D2 L B' U R2 F2 D' L2 U' B R"]) {
    const start = call({ op: "twist", moves: scramble }).state;
    const solution = call({ op: "solve", cfen: start.cfen });
    assert.ok(solution.moves.length > 0);
    assert.equal(solution.state.solved, true);
    assert.equal(call({ op: "twist", cfen: start.cfen, moves: solution.moves.join(" ") }).state.solved, true);
    assert.equal(call({ op: "state", cfen: start.cfen }).state.cfen, start.cfen);
  }
  assert.equal(call({ op: "solve" }).state.solved, true);
  console.log("PASS solve: replay four scrambles, rotated grips, full solution invariant");

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
  console.log("PASS find: exact and wildcard targets, shortest path, depth exhaustion, no mutation");

  rejects({ op: "twist", moves: "R garbage U" }, /invalid/);
  rejects({ op: "twist", moves: "999Rw" }, /invalid/);
  rejects({ op: "twist", moves: "R ".repeat(5000) }, /exceeds/);
  rejects({ op: "state", cfen: "YB|Y999999999/R9/B9/W9/O9/G9" }, /runs 1–9/);
  rejects({ op: "state", cfen: "YB|?9/R9/B9/W9/O9/G9" }, /only allowed/);
  rejects({ op: "state", cfen: "YB|W9/R9/B9/W9/O9/G9" }, /color|sticker|center/i);
  rejects({ op: "state", cfen: "WB|W9/R9/B9/Y9/O9/G9" }, /YB/);
  rejects({ op: "find", target: solved.cfen, maxDepth: 9 }, /depth/);
  rejects({ op: "find", target: solved.cfen, maxDepth: -1 }, /depth/);
  rejects({ op: "find", target: "YB|?4/?4/?4/?4/?4/?4", maxDepth: 0 }, /3x3/);
  rejects({ op: "solve", moves: "R", method: "cfop" }, /does not solve/);
  rejects({ op: "missing" }, /unknown/);
  assert.equal(JSON.parse(globalThis.cubeAPI("{")).ok, false);
  assert.equal(JSON.parse(globalThis.cubeAPI()).ok, false);
  assert.equal(JSON.parse(globalThis.cubeAPI(42)).ok, false);
  console.log("PASS API bounds: invalid moves, unsafe runs, impossible states and unavailable solvers");
  process.exit(0);
})().catch(error => { console.error(error); process.exit(1); });
