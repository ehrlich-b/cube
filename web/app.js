import { loadEngine } from "./engine.js";
import { CubeView } from "./cube-view.js";
import { solverAssets } from "./solver-assets.js";

const $ = id => document.getElementById(id);
const view = new CubeView($("cube"), $("net"), $("camera"), $("stage"));
let engine, state, initial, history = [], historyIndex = 0, busy = false, running = false, sequence = null, job = null;
let scramble = "", baseCFEN = "", currentMode = "practice";
let restoring = true;
let size = 3, saved3x3Method = "kociemba", scrambledCache;
let computeWorker;
// At most one compressed copy of each manifest asset (<4.6 MB total). Expanded
// tables stay in the active worker and are released on size changes/cancel.
const solverAssetBytes = new Map();
const pendingHash = location.hash.slice(1);
const buildVersion = new URL(import.meta.url).pathname.match(/\/app\.([a-f0-9]{16})\.js$/)?.[1];
const primaryControls = ["scramble", "solve", "solve-method", "size", "turn-layer", "turn-width", "reset", "run-algorithm", "start-lesson", "find", "import", "export"];

function notice(message, error = false) {
  $("notice").textContent = message;
  $("notice").classList.toggle("error", error);
}

function updateControls() {
  const locked = !engine || busy || !!job;
  for (const id of primaryControls) $(id).disabled = locked;
  for (const button of document.querySelectorAll("[data-move]")) button.disabled = locked || (size % 2 === 0 && /^[MES]/.test(button.dataset.move));
  for (const id of ["tab-lesson", "tab-search", "start-lesson", "find"]) $(id).disabled = locked || size !== 3;
  $("undo").disabled = locked || historyIndex === 0;
  $("redo").disabled = locked || historyIndex === history.length - 1;
  $("back-step").disabled = locked || !sequence || sequence.index === 0;
  $("step").disabled = locked || !sequence || sequence.index === sequence.moves.length;
  $("play").disabled = !engine || !!job || !sequence || (busy && !running) || (!running && sequence.index === sequence.moves.length);
  $("scrubber").disabled = locked;
  $("sequence-moves").inert = locked;
  $("play").textContent = running ? "Pause" : "Play";
  $("cancel-task").hidden = !job;
  $("reload").disabled = busy || !!job || (restoring && !!state);
  if (engine && state && !busy && !job && !restoring) persistHash();
}

function showState(next) {
  if (size !== next.size) configureSize(next.size);
  state = next;
  view.render(state);
  $("state-label").textContent = state.solved ? "● Solved" : "● Mixed";
  $("state-label").classList.toggle("mixed", !state.solved);
  $("cfen").value = state.cfen;
  updateControls();
}

function configureSize(nextSize) {
  // Reduction tables are sizable. Reuse them for this size, without keeping
  // several dimensions' caches alive when a phone switches cube sizes.
  if (nextSize !== size && computeWorker) {
    computeWorker.terminate();
    computeWorker = null;
  }
  if (size === 3) saved3x3Method = $("solve-method").value;
  size = nextSize;
  initial = engine({ op: "state", size }).state;
  $("size").value = size;
  $("size-label").textContent = `${size} × ${size}`;
  $("layer-max").textContent = size;
  $("size-note").hidden = size === 3;
  for (const option of $("solve-method").options) {
    option.hidden = option.disabled = (option.value === "reduction") === (size === 3);
  }
  $("solve-method").value = size === 3 ? saved3x3Method : "reduction";
  $("turn-layer").replaceChildren(...Array.from({ length: size }, (_, index) => {
    const option = document.createElement("option");
    option.value = index + 1;
    option.textContent = index === 0 ? "1 · outer" : String(index + 1);
    return option;
  }));
  $("turn-width").value = "single";
  updateMoveLabels();
  if (size !== 3) setMode("practice");
}

function selectedMove(token) {
  if (!/^[URFDLB]/.test(token)) return token;
  const depth = Number($("turn-layer").value);
  const wide = $("turn-width").value === "wide";
  return (depth === 1 ? "" : depth) + token[0] + (wide ? (depth === 1 ? "" : "w") : "") + token.slice(1);
}

function updateMoveLabels() {
  for (const button of $("move-buttons").children) {
    const token = selectedMove(button.dataset.move);
    button.textContent = token.replace("'", "′");
    button.setAttribute("aria-label", `Turn ${token}`);
  }
  updateViewHint();
}

function updateViewHint() {
  $("view-hint").textContent = $("stage").hidden ? "U above · L F R B across · D below" :
    `Drags use Layer ${$("turn-layer").value} · ${$("turn-width").value === "wide" ? "Wide" : "Single"}; background orbits`;
}

function record(next) {
  history = history.slice(0, historyIndex + 1);
  history.push(next);
  historyIndex++;
}

function clearSequence() {
  running = false;
  sequence = null;
  $("playback").hidden = true;
  $("lesson-content").replaceChildren();
  $("search-result").replaceChildren();
}

function freshState(next, newScramble = "", newBase = "") {
  clearSequence();
  history = [next];
  historyIndex = 0;
  scramble = newScramble;
  baseCFEN = newBase;
  $("scramble-text").textContent = scramble || (baseCFEN ? "Imported CFEN state" : "Start solved, or mix things up.");
  showState(next);
}

async function turn(token, duration = Number($("speed").value), settle = null, next = null) {
  next ??= engine({ op: "twist", cfen: state.cfen, moves: token }).state;
  await (settle ? settle() : view.animate(token, duration));
  record(next);
  showState(next);
}

async function manualTurn(token, settle = null) {
  if (!settle && (!engine || busy || job)) return;
  busy = true;
  updateControls();
  try {
    if (token) {
      // Validate before discarding playback or changing undoable state.
      const next = engine({ op: "twist", cfen: state.cfen, moves: token }).state;
      clearSequence();
      await turn(token, Number($("speed").value), settle, next);
      notice(`Turned ${token}.`);
    } else { await settle(); view.render(state); }
  }
  catch (error) { notice(error.message, true); view.render(state); }
  finally { busy = false; updateControls(); }
}

view.selectMove = selectedMove;
view.onTurnStart = () => {
  if (!engine || busy || job || running) return false;
  busy = true;
  updateControls();
  return true;
};
view.onTurnEnd = (token, settle) => safe(() => manualTurn(token, settle));

function refreshPlayback() {
  if (!sequence) return;
  $("progress").textContent = `${sequence.index} / ${sequence.moves.length}`;
  $("scrubber").value = sequence.index;
  const first = sequence.highlighted < 0 ? 0 : Math.min(sequence.highlighted, sequence.index);
  const last = sequence.highlighted < 0 ? sequence.moves.length - 1 : Math.max(sequence.highlighted, sequence.index);
  for (let index = first; index <= last && index < sequence.moves.length; index++) {
    const button = sequence.buttons[index];
    button.classList.toggle("done", index < sequence.index);
    button.classList.toggle("current", index === sequence.index);
    if (index === sequence.index) button.setAttribute("aria-current", "step");
    else button.removeAttribute("aria-current");
  }
  sequence.highlighted = sequence.index;
  updateControls();
}

async function prepareSequence(moves, kind, title, expected, stages = [], preparedFrames) {
  const wasBusy = busy;
  busy = true;
  updateControls();
  try {
    const frames = preparedFrames || [state];
    if (!preparedFrames) for (const [index, move] of moves.entries()) {
      frames.push(engine({ op: "twist", cfen: frames.at(-1).cfen, moves: move }).state);
      if (index % 32 === 31) await new Promise(requestAnimationFrame);
    }
    if (expected && frames.at(-1).cfen !== expected) throw new Error("Sequence did not reach its verified checkpoint.");
    sequence = { moves, frames, index: 0, historyStart: historyIndex, kind, title, stages, buttons: [], highlighted: -1 };
    $("sequence-kind").textContent = kind;
    $("sequence-title").textContent = title;
    $("sequence-count").textContent = `${moves.length} moves`;
    $("scrubber").max = moves.length;
    $("sequence-moves").replaceChildren();
    $("sequence-moves").classList.toggle("grouped", stages.length > 0);
    const appendMove = (parent, move, index) => {
      const button = document.createElement("button");
      button.textContent = move;
      button.title = `Jump to after move ${index + 1}`;
      button.addEventListener("click", () => jump(index + 1));
      sequence.buttons.push(button);
      parent.append(button);
    };
    if (stages.length) {
      if (stages.flatMap(stage => stage.moves).join(" ") !== moves.join(" ")) throw new Error("Stage moves differ from the verified solution.");
      let index = 0;
      for (const stage of stages) {
        const group = document.createElement("section");
        group.className = "sequence-stage";
        group.dataset.stage = stage.name;
        const heading = document.createElement("h3");
        heading.textContent = `${stage.name} · ${stage.turns} turns`;
        const cases = document.createElement("p");
        cases.className = "case-name";
        cases.textContent = stage.cases.join(" + ");
        const buttons = document.createElement("div");
        group.append(heading, cases, buttons);
        for (const move of stage.moves) appendMove(buttons, move, index++);
        $("sequence-moves").append(group);
      }
    } else {
      const fragment = document.createDocumentFragment();
      for (const [index, move] of moves.entries()) {
        appendMove(fragment, move, index);
        if (size !== 3 && index % 64 === 63) await new Promise(requestAnimationFrame);
      }
      $("sequence-moves").append(fragment);
    }
    $("playback").hidden = false;
    refreshPlayback();
  } finally { busy = wasBusy; updateControls(); }
}

function jump(index) {
  if (!sequence || busy || job) return;
  running = false;
  index = Math.max(0, Math.min(sequence.moves.length, index));
  // A scrub is one undoable state change, rather than invented physical turns.
  record(sequence.frames[index]);
  sequence.index = index;
  showState(sequence.frames[index]);
  refreshPlayback();
  finishSequence();
}

function finishSequence() {
  if (!sequence || sequence.index !== sequence.moves.length) return;
  running = false;
  if (sequence.kind === "CHECKPOINT") notice("Checkpoint reached. Compare the check above, then ask for your next hint.");
  else notice(state.solved ? "Cube complete. All six faces are solved." : "Sequence complete.");
}

async function stepSequence() {
  if (!sequence || busy || job || sequence.index >= sequence.moves.length) return;
  busy = true;
  updateControls();
  try {
    await turn(sequence.moves[sequence.index], Number($("speed").value), null, sequence.frames[sequence.index + 1]);
    sequence.index++;
    finishSequence();
    refreshPlayback();
  } catch (error) { running = false; notice(error.message, true); view.render(state); }
  finally { busy = false; updateControls(); }
}

async function playSequence() {
  if (running) { running = false; updateControls(); return; }
  if (!sequence || busy || job) return;
  running = true;
  updateControls();
  while (running && sequence && sequence.index < sequence.moves.length) await stepSequence();
  running = false;
  refreshPlayback();
}

function compute(request, label) {
  if (job) throw new Error("Another task is already running.");
  running = false;
  const retainWorker = size !== 3;
  const worker = retainWorker ? (computeWorker ??= new Worker(new URL("./worker.js", import.meta.url), { type: "module" })) :
    new Worker(new URL("./worker.js", import.meta.url), { type: "module" });
  const frames = request.prepareFrames ? [state] : null;
  notice(label);
  return new Promise((resolve, reject) => {
    let finished = false, phase = "loading", completed = 0, total = 0, computedWallMs;
    const started = performance.now();
    const progress = () => {
      if (request.op !== "solve" || size === 3) return;
      const seconds = ((performance.now() - started) / 1000).toFixed(1);
      notice(phase === "replay" ? `Preparing playback · ${completed} / ${total} moves. Cancel anytime.` :
        `${phase === "loading" ? "Loading the cube engine" : `Solving ${size} × ${size}`} · ${seconds}s elapsed. Cancel anytime.`);
    };
    const ticker = setInterval(progress, 250);
    const finish = (error, result) => {
      if (finished) return;
      finished = true;
      clearTimeout(timer);
      clearInterval(ticker);
      // Termination interrupts synchronous WASM search without waiting for it to yield.
      if (error || !retainWorker) {
        worker.terminate();
        if (worker === computeWorker) computeWorker = null;
      }
      worker.onmessage = worker.onerror = null;
      job = null;
      $("cancel-search").hidden = true;
      $("find").hidden = false;
      updateControls();
      if (error) reject(error); else resolve({ ...result, ...(frames ? { frames } : {}), computedWallMs: computedWallMs ?? performance.now() - started });
    };
    const timeout = request.op === "solve" && size !== 3 ? 120000 : 30000;
    const timer = setTimeout(() => finish(new Error(request.op === "find" ? "The search took more than 30 seconds. Try a smaller depth or a more flexible target." : "The solver reached its time limit. Try again or use a simpler state.")), timeout);
    job = { cancel: () => finish(new Error(`${request.op === "find" ? "Search" : "Computation"} canceled.`)) };
    updateControls();
    worker.onmessage = ({ data }) => {
      if (data.type === "asset") {
        if (Object.hasOwn(solverAssets, data.key) && data.bytes instanceof Uint8Array && data.bytes.length === solverAssets[data.key].bytes) {
          solverAssetBytes.set(data.key, data.bytes);
        }
      } else if (data.type === "progress" || data.type === "frames") {
        phase = data.type === "frames" ? "replay" : data.phase;
        if (phase === "replay" && computedWallMs === undefined) computedWallMs = performance.now() - started;
        completed = data.completed || 0; total = data.total || 0;
        if (data.frames) frames.push(...data.frames);
        progress();
      } else finish(data.ok ? null : Object.assign(new Error(data.error), { code: data.code }), data.data);
    };
    worker.onerror = event => {
      event.preventDefault();
      finish(Object.assign(new Error(event.message || "The cube worker could not start. Check your connection and try again."), { code: "asset-load" }));
    };
    const assets = [...solverAssetBytes].filter(([key]) => key === "coordinates" || key === `nxn-${size}`);
    worker.postMessage({ request, module: engine.module, assets });
  });
}

function setMode(mode) {
  if (size !== 3 && mode !== "practice") return;
  currentMode = mode;
  for (const name of ["practice", "lesson", "search"]) {
    $(name).hidden = name !== mode;
    $(`tab-${name}`).setAttribute("aria-selected", String(name === mode));
    $(`tab-${name}`).tabIndex = name === mode ? 0 : -1;
  }
}

async function computeSequence(request, label) {
  try {
    const result = await compute(request, label);
    clearSequence();
    updateControls();
    return result;
  } catch (error) {
    // Deployment failures retain playback for a safe reload; cancellation and
    // other computation failures keep the existing behavior of clearing it.
    if (error.code !== "asset-load") { clearSequence(); updateControls(); }
    throw error;
  }
}

async function solve() {
  const started = performance.now();
  const result = await computeSequence({ op: "solve", cfen: state.cfen, method: $("solve-method").value, prepareFrames: size !== 3, profile: $("cube").hasAttribute("data-profile-solve") }, "Finding a verified solution…");
  $("cube").dataset.solveMs = result.solveMs;
  $("cube").dataset.solveWallMs = result.computedWallMs ?? performance.now() - started;
  if (result.runtime) $("cube").dataset.solveRuntime = JSON.stringify(result.runtime);
  await prepareSequence(result.moves, "SOLUTION", `Your path to solved · ${result.method}`, result.state.cfen, result.stages, result.frames);
  notice(result.moves.length ? "Solution ready. Play, step through, or drag the slider to explore." : "Your cube is already solved.");
}

async function hint() {
  const result = await computeSequence({ op: "learn", cfen: state.cfen }, "Planning your next beginner checkpoint…");
  const step = result.steps[0];
  if (!step) { $("lesson-content").textContent = "All six faces are solved. You’ve reached the final checkpoint."; notice("Lesson complete."); return; }
  showLesson(step);
  await prepareSequence(step.moves, "CHECKPOINT", step.title, step.after.cfen);
  notice("Follow the moves, check your cube, then get the next hint. Your own turns are always recoverable.");
}

function showLesson(step) {
  $("lesson-content").replaceChildren();
  const title = document.createElement("h3");
  title.textContent = step.title;
  $("lesson-content").append(title);
  for (const action of step.actions) {
    const box = document.createElement("div");
    box.className = "lesson-action";
    box.textContent = action.instruction;
    const moves = document.createElement("code");
    moves.textContent = action.moves.join(" ");
    box.append(moves);
    $("lesson-content").append(box);
  }
  const check = document.createElement("p");
  check.className = "checkpoint-check";
  check.textContent = `Check: ${step.check}`;
  $("lesson-content").append(check);
  const next = document.createElement("button");
  next.className = "secondary";
  next.textContent = "Next hint →";
  next.addEventListener("click", () => { if (!busy && !job) safe(hint); });
  $("lesson-content").append(next);
}

async function find() {
  const startCFEN = state.cfen;
  const target = $("target").value.trim();
  const maxDepth = Number($("depth").value);
  if (!Number.isInteger(maxDepth) || maxDepth < 0 || maxDepth > 10) throw new Error("Choose a whole-number depth from 0 to 10.");
  const pending = computeSequence({ op: "find", cfen: startCFEN, target, maxDepth }, "Searching from your current cube… You can cancel anytime.");
  $("find").hidden = true;
  $("cancel-search").hidden = false;
  const result = await pending;
  if (!result.found) { $("search-result").textContent = `No path found within ${maxDepth} moves.`; notice("Search finished. Try a larger depth or allow more wildcard stickers."); return; }
  $("search-result").textContent = result.moves.length ? `Found ${result.moves.length} moves: ${result.moves.join(" ")}` : "Your cube already matches this target.";
  await prepareSequence(result.moves, "SEARCH RESULT", "Your path to the pattern", result.state.cfen);
  notice("Search complete. Play the result to reach your target.");
}

function randomScramble() {
  const faces = "RULDFB", suffixes = ["", "'", "2"], moves = [];
  let last;
  for (let i = 0; i < 20; i++) {
    let face;
    do { face = faces[crypto.getRandomValues(new Uint32Array(1))[0] % 6]; } while (face === last);
    const depth = size > 3 ? 1 + crypto.getRandomValues(new Uint32Array(1))[0] % Math.floor(size / 2) : 1;
    moves.push((depth === 1 ? "" : depth) + face + suffixes[crypto.getRandomValues(new Uint32Array(1))[0] % 3]);
    last = face;
  }
  return moves;
}

async function mix() {
  const moves = randomScramble();
  freshState(initial, moves.join(" "));
  busy = true;
  updateControls();
  try {
    for (const move of moves) {
      await turn(move, 65);
      // Save each committed frame even while the next scramble turn animates.
      persistHash();
    }
    notice(size === 3 ? "Fresh scramble. Try a sequence, a lesson, or a full solution." : "Fresh scramble. Explore layers or play a full solution.");
  }
  finally { busy = false; updateControls(); }
}

async function algorithm() {
  const moves = engine({ op: "twist", size, moves: $("algorithm").value }).moves;
  if (!moves.length) throw new Error("Type an algorithm first, for example R U R' U'.");
  clearSequence();
  await prepareSequence(moves, "ALGORITHM", "Explore your sequence");
  await playSequence();
}

async function restoreHash(text) {
  const params = new URLSearchParams(text);
  const newBase = params.get("state") || "";
  const newScramble = params.get("scramble") || "";
  const alg = params.get("alg") || "";
  const sizeText = params.get("size") || "3";
  if (!/^[2-7]$/.test(sizeText)) throw new Error("Invalid cube size in this link.");
  const nextSize = Number(sizeText);
  // Validate the entire link before changing any visible state.
  const next = engine({ op: "state", size: nextSize, cfen: newBase, moves: newScramble }).state;
  const moves = engine({ op: "twist", size: nextSize, moves: alg }).moves;
  const indexText = params.get("index") || "0";
  if (!/^\d+$/.test(indexText) || Number(indexText) > moves.length) throw new Error("Invalid playback position in this link.");
  const index = Number(indexText);
  const current = engine({ op: "twist", cfen: next.cfen, moves: moves.slice(0, index).join(" ") }).state;
  if (params.has("current") && engine({ op: "state", size: nextSize, cfen: params.get("current") }).state.cfen !== current.cfen) {
    throw new Error("The saved cube does not match its playback position.");
  }
  const kind = params.get("kind") || "ALGORITHM";
  if (nextSize !== 3 && ["CHECKPOINT", "SEARCH RESULT"].includes(kind)) throw new Error("Lessons and search are 3x3-only.");
  if (!["ALGORITHM", "SOLUTION", "CHECKPOINT", "SEARCH RESULT"].includes(kind)) throw new Error("Invalid sequence kind in this link.");
  const title = params.get("title") || "Shared sequence";
  const stages = JSON.parse(params.get("stages") || "[]");
  if (!Array.isArray(stages) || stages.some(stage => !stage || typeof stage.name !== "string" || !Number.isInteger(stage.turns) || !Array.isArray(stage.cases) || stage.cases.some(name => typeof name !== "string") || !Array.isArray(stage.moves) || stage.moves.some(move => typeof move !== "string")) || (stages.length && stages.flatMap(stage => stage.moves).join(" ") !== moves.join(" "))) {
    throw new Error("Invalid playback stages in this link.");
  }
  restoring = true;
  try {
    // Rebuild teaching text from the sequence's starting cube, including links
    // saved before lessons carried their instructions through a reload.
    let lesson;
    if (kind === "CHECKPOINT") {
      const result = await compute({ op: "learn", cfen: next.cfen }, "Restoring your beginner checkpoint…");
      lesson = result.steps[0];
      if (!lesson || lesson.moves.join(" ") !== moves.join(" ")) throw new Error("The saved moves do not match this beginner checkpoint.");
      const after = engine({ op: "twist", cfen: next.cfen, moves: moves.join(" ") }).state;
      if (after.cfen !== lesson.after.cfen) throw new Error("The saved lesson did not reach its verified checkpoint.");
    }
    freshState(next, newScramble, newBase);
    $("algorithm").value = params.get("draft") ?? alg;
    if (lesson) { showLesson(lesson); setMode("lesson"); }
    if (moves.length || params.has("index")) {
      await prepareSequence(moves, kind, title, null, stages);
      history = sequence.frames.slice(0, index + 1);
      historyIndex = index;
      sequence.index = index;
      showState(current);
      refreshPlayback();
    }
  } finally { restoring = false; }
  updateControls();
  notice(moves.length ? "Saved cube loaded. Press Play to continue its sequence." : "Saved cube loaded.");
}

function persistHash() {
  const params = new URLSearchParams();
  const key = `${size}:${baseCFEN}:${scramble}`;
  if (scrambledCache?.key !== key) scrambledCache = { key, state: engine({ op: "state", size, cfen: baseCFEN, moves: scramble }).state };
  const scrambled = scrambledCache.state;
  if (size !== 3) params.set("size", size);
  const start = sequence ? sequence.frames[0] : state;
  if (start.cfen !== scrambled.cfen) params.set("state", start.cfen);
  else { if (baseCFEN) params.set("state", baseCFEN); if (scramble) params.set("scramble", scramble); }
  params.set("current", state.cfen);
  // The editor is independent of prepared playback. Keep even empty or
  // incomplete text verbatim without treating it as executable notation.
  params.set("draft", $("algorithm").value);
  if (sequence) {
    params.set("alg", sequence.moves.join(" "));
    params.set("index", sequence.index);
    params.set("kind", sequence.kind);
    params.set("title", sequence.title);
    if (sequence.stages.length) params.set("stages", JSON.stringify(sequence.stages));
  }
  const hash = `#${params}`;
  if (location.hash !== hash) historyReplace(hash);
}

async function copy(text, message) {
  try { await navigator.clipboard.writeText(text); notice(message); }
  catch {
    notice("Clipboard unavailable. The text is selected so you can copy it.");
    const field = document.createElement("input");
    field.readOnly = true;
    field.value = text;
    field.setAttribute("aria-label", "Text to copy");
    field.style.width = "100%";
    field.style.marginTop = "8px";
    $("notice").append(field);
    field.select();
  }
}

async function safe(action) {
  try { await action(); }
  catch (error) {
    notice(error.message, true);
    if (error.code !== "asset-load" || !buildVersion) return;
    try {
      const response = await fetch(new URL("./version.json", import.meta.url), { cache: "no-store", signal: AbortSignal.timeout(5000) });
      if (!response.ok) return;
      const { version } = await response.json();
      if (/^[a-f0-9]{16}$/.test(version) && version !== buildVersion) {
        $("update-notice").hidden = false;
        if ($("notice").textContent === error.message) notice("Your cube and playback will be kept when you reload.");
      }
    } catch { /* An unavailable version check leaves the original network error visible. */ }
  }
}

for (const prime of [false, true]) for (const face of ["U", "D", "L", "R", "F", "B"]) {
  const button = document.createElement("button");
  const token = face + (prime ? "'" : "");
  button.textContent = face + (prime ? "′" : "");
  button.dataset.move = token;
  button.disabled = true;
  button.classList.toggle("prime", prime);
  button.setAttribute("aria-label", `Turn ${token}`);
  button.addEventListener("click", () => safe(() => manualTurn(selectedMove(token))));
  $("move-buttons").append(button);
}
for (const token of ["M", "E", "S", "x", "y", "z", "Rw", "Lw", "Uw", "Dw", "Fw", "Bw"]) {
  const button = document.createElement("button");
  button.dataset.move = token;
  button.textContent = token;
  button.disabled = true;
  button.addEventListener("click", event => safe(() => manualTurn(token + (event.shiftKey ? "'" : ""))));
  $("extra-moves").append(button);
}
for (const mode of ["practice", "lesson", "search"]) $(`tab-${mode}`).addEventListener("click", () => setMode(mode));
document.querySelector(".tabs").addEventListener("keydown", event => {
  const modes = size === 3 ? ["practice", "lesson", "search"] : ["practice"];
  if (!["ArrowLeft", "ArrowRight", "Home", "End"].includes(event.key)) return;
  event.preventDefault();
  const index = event.key === "Home" ? 0 : event.key === "End" ? modes.length - 1 : (modes.indexOf(currentMode) + (event.key === "ArrowRight" ? 1 : modes.length - 1)) % modes.length;
  setMode(modes[index]);
  $(`tab-${modes[index]}`).focus();
});
$("scramble").addEventListener("click", () => safe(mix));
$("solve").addEventListener("click", () => safe(solve));
$("run-algorithm").addEventListener("click", () => safe(algorithm));
$("algorithm").addEventListener("input", () => { if (engine && state && !restoring) persistHash(); });
$("start-lesson").addEventListener("click", () => safe(hint));
$("find").addEventListener("click", () => safe(find));
$("cancel-search").addEventListener("click", () => job?.cancel());
$("cancel-task").addEventListener("click", () => job?.cancel());
$("size").addEventListener("change", () => { configureSize(Number($("size").value)); freshState(initial); notice(`${size} × ${size} cube ready.`); });
for (const id of ["turn-layer", "turn-width"]) $(id).addEventListener("change", updateMoveLabels);
$("play").addEventListener("click", () => safe(playSequence));
$("step").addEventListener("click", () => safe(stepSequence));
$("back-step").addEventListener("click", () => jump(sequence.index - 1));
$("scrubber").addEventListener("input", event => jump(Number(event.target.value)));
$("reset").addEventListener("click", () => { freshState(initial); notice("Cube reset to solved."); });
$("undo").addEventListener("click", () => { clearSequence(); showState(history[--historyIndex]); notice("Last move undone."); });
$("redo").addEventListener("click", () => { clearSequence(); showState(history[++historyIndex]); notice("Move redone."); });
$("home-view").addEventListener("click", () => view.center());
for (const mode of ["3d", "net"]) $(`view-${mode}`).addEventListener("click", () => {
  view.cancelDrag();
  $("stage").hidden = mode !== "3d";
  $("net").hidden = mode !== "net";
  if (mode === "3d") view.orient();
  $("view-3d").setAttribute("aria-pressed", String(mode === "3d"));
  $("view-net").setAttribute("aria-pressed", String(mode === "net"));
  updateViewHint();
});
$("import").addEventListener("click", () => safe(() => {
  const cfen = $("cfen").value.trim();
  if (!cfen) throw new Error("Enter a CFEN cube state before importing.");
  const next = engine({ op: "state", cfen }).state;
  freshState(next, "", next.cfen);
  notice("Physical cube state imported.");
}));
$("export").addEventListener("click", () => { $("cfen").value = state.cfen; $("cfen").select(); safe(() => copy(state.cfen, "CFEN copied. Paste it here later to restore your cube.")); });
$("share").addEventListener("click", () => safe(async () => {
  if (!engine || busy || job) return;
  persistHash();
  await copy(location.href, "Link copied. It includes the current cube, algorithm draft, sequence and playback position.");
}));
function historyReplace(hash) { window.history.replaceState(null, "", hash || location.pathname); }
$("reload").addEventListener("click", () => safe(() => {
  if (busy || job || (restoring && state)) return;
  if (engine && state) persistHash();
  location.reload();
}));
window.addEventListener("hashchange", () => { if (engine && !busy && !job) safe(() => restoreHash(location.hash.slice(1))); });
window.addEventListener("keydown", event => {
  if (event.ctrlKey || event.metaKey || event.altKey || event.repeat || event.target.closest("input,textarea,select,[contenteditable]")) return;
  if (event.key === "?") { event.preventDefault(); $("keyboard-help").togglePopover(); return; }
  if ($("keyboard-help").matches(":popover-open")) return;
  const key = event.key.toLowerCase();
  if (/^[1-7]$/.test(key) && Number(key) <= size && !busy && !job) {
    event.preventDefault(); $("turn-layer").value = key; updateMoveLabels(); return;
  }
  const token = "xyz".includes(key) && key.length === 1 ? key : "urfdlbmes".includes(key) && key.length === 1 ? key.toUpperCase() : "";
  if (!token) return;
  event.preventDefault();
  safe(() => manualTurn(selectedMove(token + (event.shiftKey ? "'" : ""))));
});

safe(async () => {
  engine = await loadEngine();
  configureSize(3);
  freshState(initial);
  try {
    if (pendingHash) await restoreHash(pendingHash);
    else notice("Ready when you are. Turn a face or start with a scramble.");
  } finally { restoring = false; }
  updateControls();
});
