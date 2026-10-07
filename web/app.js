import { loadEngine } from "./engine.js";
import { CubeView } from "./cube-view.js";

const $ = id => document.getElementById(id);
const view = new CubeView($("cube"), $("net"), $("camera"), $("stage"));
let engine, state, initial, history = [], historyIndex = 0, busy = false, running = false, sequence = null, job = null;
let scramble = "", baseCFEN = "", currentMode = "practice";
const pendingHash = location.hash.slice(1);
const primaryControls = ["scramble", "solve", "reset", "run-algorithm", "start-lesson", "find", "import", "export"];

function notice(message, error = false) {
  $("notice").textContent = message;
  $("notice").classList.toggle("error", error);
}

function updateControls() {
  const locked = !engine || busy || !!job;
  for (const id of primaryControls) $(id).disabled = locked;
  for (const button of document.querySelectorAll("[data-move]")) button.disabled = locked;
  $("undo").disabled = locked || historyIndex === 0;
  $("redo").disabled = locked || historyIndex === history.length - 1;
  $("back-step").disabled = locked || !sequence || sequence.index === 0;
  $("step").disabled = locked || !sequence || sequence.index === sequence.moves.length;
  $("play").disabled = !engine || !!job || !sequence || (busy && !running) || (!running && sequence.index === sequence.moves.length);
  $("scrubber").disabled = locked;
  for (const button of $("sequence-moves").querySelectorAll("button")) button.disabled = locked;
  $("play").textContent = running ? "Pause" : "Play";
}

function showState(next) {
  state = next;
  view.render(state);
  $("state-label").textContent = state.solved ? "● Solved" : "● Mixed";
  $("state-label").classList.toggle("mixed", !state.solved);
  $("cfen").value = state.cfen;
  updateControls();
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

async function turn(token, duration = Number($("speed").value)) {
  const next = engine({ op: "twist", cfen: state.cfen, moves: token }).state;
  await view.animate(token, duration);
  record(next);
  showState(next);
}

async function manualTurn(token) {
  if (!engine || busy || job) return;
  clearSequence();
  busy = true;
  updateControls();
  try { await turn(token); notice(`Turned ${token}.`); }
  catch (error) { notice(error.message, true); view.render(state); }
  finally { busy = false; updateControls(); }
}

function refreshPlayback() {
  if (!sequence) return;
  $("progress").textContent = `${sequence.index} / ${sequence.moves.length}`;
  $("scrubber").value = sequence.index;
  [...$("sequence-moves").children].forEach((button, index) => {
    button.classList.toggle("done", index < sequence.index);
    button.classList.toggle("current", index === sequence.index);
    if (index === sequence.index) button.setAttribute("aria-current", "step");
    else button.removeAttribute("aria-current");
  });
  updateControls();
}

function prepareSequence(moves, kind, title, expected) {
  const frames = [state];
  for (const move of moves) frames.push(engine({ op: "twist", cfen: frames.at(-1).cfen, moves: move }).state);
  if (expected && frames.at(-1).cfen !== expected) throw new Error("Sequence did not reach its verified checkpoint.");
  sequence = { moves, frames, index: 0, historyStart: historyIndex, kind };
  $("sequence-kind").textContent = kind;
  $("sequence-title").textContent = title;
  $("sequence-count").textContent = `${moves.length} moves`;
  $("scrubber").max = moves.length;
  $("sequence-moves").replaceChildren();
  moves.forEach((move, index) => {
    const button = document.createElement("button");
    button.textContent = move;
    button.title = `Jump to after move ${index + 1}`;
    button.addEventListener("click", () => jump(index + 1));
    $("sequence-moves").append(button);
  });
  $("playback").hidden = false;
  refreshPlayback();
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
    await turn(sequence.moves[sequence.index]);
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
  const worker = new Worker(new URL("./worker.js", import.meta.url), { type: "module" });
  notice(label);
  return new Promise((resolve, reject) => {
    const finish = (error, result) => {
      clearTimeout(timer);
      worker.terminate();
      job = null;
      $("cancel-search").hidden = true;
      $("find").hidden = false;
      updateControls();
      if (error) reject(error); else resolve(result);
    };
    const timer = setTimeout(() => finish(new Error("The search took more than 30 seconds. Try a smaller depth or a more flexible target.")), 30000);
    job = { cancel: () => finish(new Error("Search canceled.")) };
    updateControls();
    worker.onmessage = ({ data }) => finish(data.ok ? null : new Error(data.error), data.data);
    worker.onerror = event => finish(new Error(event.message || "The cube worker could not start."));
    worker.postMessage(request);
  });
}

function setMode(mode) {
  currentMode = mode;
  for (const name of ["practice", "lesson", "search"]) {
    $(name).hidden = name !== mode;
    $(`tab-${name}`).setAttribute("aria-selected", String(name === mode));
    $(`tab-${name}`).tabIndex = name === mode ? 0 : -1;
  }
}

async function solve() {
  clearSequence();
  const result = await compute({ op: "solve", cfen: state.cfen }, "Finding a verified solution…");
  prepareSequence(result.moves, "SOLUTION", `Your path to solved · ${result.method}`, result.state.cfen);
  notice(result.moves.length ? "Solution ready. Play, step through, or drag the slider to explore." : "Your cube is already solved.");
}

async function hint() {
  clearSequence();
  const result = await compute({ op: "learn", cfen: state.cfen }, "Planning your next beginner checkpoint…");
  const step = result.steps[0];
  if (!step) { $("lesson-content").textContent = "All six faces are solved. You’ve reached the final checkpoint."; notice("Lesson complete."); return; }
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
  prepareSequence(step.moves, "CHECKPOINT", step.title, step.after.cfen);
  notice("Follow the moves, check your cube, then get the next hint. Your own turns are always recoverable.");
}

async function find() {
  clearSequence();
  const startCFEN = state.cfen;
  const target = $("target").value.trim();
  const maxDepth = Number($("depth").value);
  if (!Number.isInteger(maxDepth) || maxDepth < 0 || maxDepth > 8) throw new Error("Choose a whole-number depth from 0 to 8.");
  const pending = compute({ op: "find", cfen: startCFEN, target, maxDepth }, "Searching from your current cube… You can cancel anytime.");
  $("find").hidden = true;
  $("cancel-search").hidden = false;
  const result = await pending;
  if (!result.found) { $("search-result").textContent = `No path found within ${maxDepth} moves.`; notice("Search finished. Try a larger depth or allow more wildcard stickers."); return; }
  $("search-result").textContent = result.moves.length ? `Found ${result.moves.length} moves: ${result.moves.join(" ")}` : "Your cube already matches this target.";
  prepareSequence(result.moves, "SEARCH RESULT", "Your path to the pattern", result.state.cfen);
  notice("Search complete. Play the result to reach your target.");
}

function randomScramble() {
  const faces = "RULDFB", suffixes = ["", "'", "2"], moves = [];
  let last;
  for (let i = 0; i < 20; i++) {
    let face;
    do { face = faces[crypto.getRandomValues(new Uint32Array(1))[0] % 6]; } while (face === last);
    moves.push(face + suffixes[crypto.getRandomValues(new Uint32Array(1))[0] % 3]);
    last = face;
  }
  return moves;
}

async function mix() {
  const moves = randomScramble();
  freshState(initial, moves.join(" "));
  busy = true;
  updateControls();
  try { for (const move of moves) await turn(move, 65); notice("Fresh scramble. Try a sequence, a lesson, or a full solution."); }
  finally { busy = false; updateControls(); }
}

async function algorithm() {
  const moves = engine({ op: "twist", moves: $("algorithm").value }).moves;
  if (!moves.length) throw new Error("Type an algorithm first, for example R U R' U'.");
  clearSequence();
  prepareSequence(moves, "ALGORITHM", "Explore your sequence");
  await playSequence();
}

function restoreHash(text) {
  const params = new URLSearchParams(text);
  const newBase = params.get("state") || "";
  const newScramble = params.get("scramble") || "";
  const alg = params.get("alg") || "";
  // Validate the entire link before changing any visible state.
  const next = engine({ op: "state", cfen: newBase, moves: newScramble }).state;
  const moves = engine({ op: "twist", moves: alg }).moves;
  freshState(next, newScramble, newBase);
  $("algorithm").value = alg;
  if (moves.length) prepareSequence(moves, "ALGORITHM", "Shared sequence");
  notice("Shared cube loaded. Press Play to explore its algorithm.");
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
  catch (error) { notice(error.message, true); }
}

for (const prime of [false, true]) for (const face of ["U", "D", "L", "R", "F", "B"]) {
  const button = document.createElement("button");
  const token = face + (prime ? "'" : "");
  button.textContent = face + (prime ? "′" : "");
  button.dataset.move = token;
  button.disabled = true;
  button.classList.toggle("prime", prime);
  button.setAttribute("aria-label", `Turn ${token}`);
  button.addEventListener("click", () => safe(() => manualTurn(token)));
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
  const modes = ["practice", "lesson", "search"];
  if (!["ArrowLeft", "ArrowRight", "Home", "End"].includes(event.key)) return;
  event.preventDefault();
  const index = event.key === "Home" ? 0 : event.key === "End" ? 2 : (modes.indexOf(currentMode) + (event.key === "ArrowRight" ? 1 : 2)) % 3;
  setMode(modes[index]);
  $(`tab-${modes[index]}`).focus();
});
$("scramble").addEventListener("click", () => safe(mix));
$("solve").addEventListener("click", () => safe(solve));
$("run-algorithm").addEventListener("click", () => safe(algorithm));
$("start-lesson").addEventListener("click", () => safe(hint));
$("find").addEventListener("click", () => safe(find));
$("cancel-search").addEventListener("click", () => job?.cancel());
$("play").addEventListener("click", () => safe(playSequence));
$("step").addEventListener("click", () => safe(stepSequence));
$("back-step").addEventListener("click", () => jump(sequence.index - 1));
$("scrubber").addEventListener("input", event => jump(Number(event.target.value)));
$("reset").addEventListener("click", () => { freshState(initial); notice("Cube reset to solved."); });
$("undo").addEventListener("click", () => { clearSequence(); showState(history[--historyIndex]); notice("Last move undone."); });
$("redo").addEventListener("click", () => { clearSequence(); showState(history[++historyIndex]); notice("Move redone."); });
$("home-view").addEventListener("click", () => view.center());
for (const mode of ["3d", "net"]) $(`view-${mode}`).addEventListener("click", () => {
  $("stage").hidden = mode !== "3d";
  $("net").hidden = mode !== "net";
  $("view-3d").setAttribute("aria-pressed", String(mode === "3d"));
  $("view-net").setAttribute("aria-pressed", String(mode === "net"));
  $("view-hint").textContent = mode === "3d" ? "Drag to look around" : "U above · L F R B across · D below";
});
$("import").addEventListener("click", () => safe(() => {
  const next = engine({ op: "state", cfen: $("cfen").value.trim() }).state;
  freshState(next, "", next.cfen);
  notice("Physical cube state imported.");
}));
$("export").addEventListener("click", () => { $("cfen").value = state.cfen; $("cfen").select(); safe(() => copy(state.cfen, "CFEN copied. Paste it here later to restore your cube.")); });
$("share").addEventListener("click", () => safe(async () => {
  if (!engine || busy || job) return;
  const params = new URLSearchParams();
  // A fresh scramble is reproducible; preserve any later manual moves with CFEN.
  const scrambled = engine({ op: "state", cfen: baseCFEN, moves: scramble }).state;
  const start = sequence ? sequence.frames[0] : state;
  if (start.cfen !== scrambled.cfen) params.set("state", start.cfen);
  else { if (baseCFEN) params.set("state", baseCFEN); if (scramble) params.set("scramble", scramble); }
  const alg = sequence ? sequence.moves.join(" ") : $("algorithm").value;
  if (alg) { engine({ op: "twist", moves: alg }); params.set("alg", alg); }
  const url = new URL(location.href);
  url.hash = params.toString();
  historyReplace(url.hash);
  await copy(url.href, "Link copied. It includes the starting cube and algorithm.");
}));
function historyReplace(hash) { window.history.replaceState(null, "", hash || location.pathname); }
window.addEventListener("hashchange", () => { if (engine && !busy && !job) safe(() => restoreHash(location.hash.slice(1))); });
window.addEventListener("keydown", event => {
  if (event.ctrlKey || event.metaKey || event.altKey || event.repeat || event.target.closest("input,textarea,select,[contenteditable]")) return;
  const key = event.key.toLowerCase();
  const token = "xyz".includes(key) && key.length === 1 ? key : "urfdlbmes".includes(key) && key.length === 1 ? key.toUpperCase() : "";
  if (!token) return;
  event.preventDefault();
  safe(() => manualTurn(token + (event.shiftKey ? "'" : "")));
});

safe(async () => {
  engine = await loadEngine();
  initial = engine({ op: "state" }).state;
  freshState(initial);
  if (pendingHash) restoreHash(pendingHash);
  else notice("Ready when you are. Turn a face or start with a scramble.");
});
