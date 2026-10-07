import assert from "node:assert/strict";
import { spawn } from "node:child_process";
import { randomInt } from "node:crypto";
import { once } from "node:events";
import { existsSync } from "node:fs";
import { mkdir, mkdtemp, readFile, readdir, rm } from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { chromium } from "playwright";

const root = fileURLToPath(new URL("../../", import.meta.url));
const scratch = path.join(root, ".scratch");
const screens = path.join(scratch, "screens");
await mkdir(screens, { recursive: true });
// Playwright's throwaway profile, downloads and temporary artifacts stay here.
process.env.TMPDIR = scratch;
process.env.TMP = scratch;
process.env.TEMP = scratch;
const profile = await mkdtemp(path.join(scratch, "playwright-profile-"));

async function cachedChromium() {
  if (process.platform !== "darwin" && existsSync(chromium.executablePath())) return chromium.executablePath();
  const cache = path.join(os.homedir(), "Library/Caches/ms-playwright");
  const versions = (await readdir(cache)).filter(name => /^chromium(_headless_shell)?-\d+$/.test(name)).sort((a, b) => Number(b.includes("headless_shell")) - Number(a.includes("headless_shell")) || Number(b.match(/\d+$/)[0]) - Number(a.match(/\d+$/)[0]));
  for (const version of versions) {
    for (const executable of ["chrome-headless-shell-mac-arm64/chrome-headless-shell", "chrome-headless-shell-mac-arm64/headless_shell", "chrome-mac-arm64/Google Chrome for Testing.app/Contents/MacOS/Google Chrome for Testing", "chrome-mac/Chromium.app/Contents/MacOS/Chromium"]) {
      const candidate = path.join(cache, version, executable);
      if (existsSync(candidate)) return candidate;
    }
  }
  throw new Error("No cached Playwright Chromium found. Install it with npx playwright install chromium.");
}

let port = process.argv.includes("--in-memory") ? randomInt(49152, 65536) : 0;
// Optional local-file transport can exercise the same browser assertions when
// an execution sandbox denies loopback sockets. The default still tests Python.
const inMemory = process.argv.includes("--in-memory");
const server = inMemory ? null : spawn("taskpolicy", ["-b", "nice", "-n", "15", "python3", "-m", "http.server", "--bind", "127.0.0.1", String(port)], { cwd: root, env: { ...process.env, PYTHONUNBUFFERED: "1" }, stdio: ["ignore", "pipe", "pipe"] });
let serverLog = "", context;
server?.stdout.on("data", chunk => { serverLog += chunk; });
server?.stderr.on("data", chunk => { serverLog += chunk; });

async function idle(page) {
  await page.locator("#reset").waitFor({ state: "visible" });
  await page.waitForFunction(() => !document.getElementById("reset").disabled);
}

async function runAlgorithm(page, moves) {
  await page.locator("#tab-practice").click();
  await page.locator("#algorithm").fill(moves);
  await page.locator("#run-algorithm").click();
  await idle(page);
  await page.waitForFunction(() => document.getElementById("play").textContent === "Play");
}

async function stickerPoint(page, face, index) {
  const sticker = page.locator(`#cube .sticker[data-face="${face}"][data-index="${index}"]`);
  const box = await sticker.boundingBox();
  assert.ok(box, `${face}${index} has a projected box`);
  const point = { x: box.x + box.width / 2, y: box.y + box.height / 2 };
  assert.equal(await page.evaluate(({ x, y, face, index }) => {
    const hit = document.elementFromPoint(x, y)?.closest(".sticker");
    return hit?.dataset.face === face && Number(hit.dataset.index) === index;
  }, { ...point, face, index }), true, `${face}${index} is the touched sticker`);
  return point;
}

async function dragSticker(page, { face, index, dx, dy, move, screenshot, cancel = false, reverse = false }) {
  const before = await page.locator("#cfen").inputValue();
  const expected = await page.evaluate(({ cfen, moves }) => JSON.parse(globalThis.cubeAPI(JSON.stringify({ op: "twist", cfen, moves }))).data.state,
    { cfen: before, moves: move || "" });
  const camera = await page.locator("#camera").getAttribute("style");
  const point = await stickerPoint(page, face, index);
  await page.mouse.move(point.x, point.y);
  await page.mouse.down();
  await page.mouse.move(point.x + dx / 2, point.y + dy / 2, { steps: 8 });
  assert.equal(await page.locator("#stage").getAttribute("data-dragging"), "turn");
  assert.equal(await page.locator("#reset").isDisabled(), true);
  assert.equal(await page.locator("#cube .layer .cubie").count(), 9);
  assert.notEqual(await page.locator("#cube .layer").evaluate(layer => getComputedStyle(layer).transform), "none");
  assert.equal(await page.locator("#cfen").inputValue(), before, "preview does not mutate the engine");
  if (screenshot) await page.screenshot({ path: path.join(screens, screenshot) });
  await page.mouse.move(point.x + dx, point.y + dy, { steps: 8 });
  if (reverse) await page.mouse.move(point.x, point.y, { steps: 8 });
  if (cancel) await page.keyboard.press("Escape");
  await page.mouse.up();
  await idle(page);
  assert.equal(await page.locator("#cfen").inputValue(), expected.cfen, `${move || "cancelled drag"} engine state`);
  for (const [face, colors] of Object.entries(expected.faces)) {
    assert.deepEqual(await page.locator(`#cube .sticker[data-face="${face}"]`).evaluateAll(stickers =>
      stickers.sort((a, b) => Number(a.dataset.index) - Number(b.dataset.index)).map(sticker => sticker.dataset.color)), colors, `${move || "cancel"} ${face} stickers`);
  }
  assert.equal(await page.locator("#cube .layer").count(), 0, "temporary layer is cleaned up");
  assert.equal(await page.locator("#camera").getAttribute("style"), camera, "sticker turns preserve the view");
  if (move) {
    assert.equal(await page.locator("#notice").textContent(), `Turned ${move}.`);
    await page.locator("#undo").click();
    assert.equal(await page.locator("#cfen").inputValue(), before, "drag is one undoable move");
    await page.locator("#redo").click();
    assert.equal(await page.locator("#cfen").inputValue(), expected.cfen, "drag can be redone");
  }
}

async function touchSticker(page, { face, index, dx, dy, move, screenshot, cancel = false }) {
  const before = await page.locator("#cfen").inputValue();
  const expected = await page.evaluate(({ cfen, moves }) => JSON.parse(globalThis.cubeAPI(JSON.stringify({ op: "twist", cfen, moves }))).data.state.cfen,
    { cfen: before, moves: move || "" });
  const point = await stickerPoint(page, face, index);
  const session = await page.context().newCDPSession(page);
  try {
    await session.send("Input.dispatchTouchEvent", { type: "touchStart", touchPoints: [{ ...point, id: 1 }] });
    for (let step = 1; step <= 8; step++) {
      await session.send("Input.dispatchTouchEvent", { type: "touchMove", touchPoints: [{ x: point.x + dx * step / 8, y: point.y + dy * step / 8, id: 1 }] });
      if (step === 4 && screenshot) await page.screenshot({ path: path.join(screens, screenshot) });
    }
    assert.equal(await page.locator("#cfen").inputValue(), before);
    assert.equal(await page.locator("#cube .layer .cubie").count(), 9);
    await session.send("Input.dispatchTouchEvent", { type: cancel ? "touchCancel" : "touchEnd", touchPoints: [] });
    await idle(page);
    assert.equal(await page.locator("#cfen").inputValue(), expected, `${move || "cancelled touch"} state`);
    assert.equal(await page.locator("#cube .layer").count(), 0);
  } finally { await session.detach(); }
}

try {
  if (server) await new Promise((resolve, reject) => {
    const timer = setTimeout(() => reject(new Error(`HTTP server did not start: ${serverLog}`)), 10000);
    server.stdout.on("data", chunk => {
      const match = String(chunk).match(/Serving HTTP .* port (\d+)/);
      if (match) { port = Number(match[1]); clearTimeout(timer); resolve(); }
    });
    server.on("error", error => { clearTimeout(timer); reject(error); });
    server.on("exit", code => { clearTimeout(timer); reject(new Error(`HTTP server exited ${code}: ${serverLog}`)); });
  });
  context = await chromium.launchPersistentContext(profile, { executablePath: await cachedChromium(), headless: true, viewport: { width: 1280, height: 800 }, args: ["--disable-gpu"], reducedMotion: "no-preference" });
  // Exercise copy actions without changing Bryan's desktop clipboard.
  await context.addInitScript(() => {
    Object.defineProperty(navigator, "clipboard", { value: { writeText: async text => { globalThis.copiedText = text; } } });
  });
  if (inMemory) {
    const types = { ".html": "text/html", ".css": "text/css", ".js": "text/javascript", ".wasm": "application/wasm", ".svg": "image/svg+xml" };
    await context.route(`http://127.0.0.1:${port}/**`, async route => {
      const pathname = new URL(route.request().url()).pathname;
      const asset = pathname === "/web/" ? "index.html" : pathname.slice("/web/".length);
      if (!/^\/web\//.test(pathname) || asset.includes("/") || asset.includes("..")) return route.fulfill({ status: 404, body: "Not found" });
      try { await route.fulfill({ status: 200, contentType: types[path.extname(asset)] || "application/octet-stream", body: await readFile(path.join(root, "web", asset)) }); }
      catch { await route.fulfill({ status: 404, body: "Not found" }); }
    });
    console.log("Transport: local files via Playwright routing; Python loopback server is not exercised.");
  }
  const page = context.pages()[0];
  const errors = [];
  page.on("pageerror", error => errors.push(error.message));
  page.on("console", message => { if (message.type() === "error") errors.push(message.text()); });
  await page.goto(`http://127.0.0.1:${port}/web/`);
  await idle(page);
  const solved = await page.locator("#cfen").inputValue();
  assert.equal(await page.locator("#cube").getAttribute("data-solved"), "true");
  assert.equal(await page.locator(".sticker").count(), 54);
  assert.equal(await page.locator("#solve-method").inputValue(), "kociemba");
  await page.screenshot({ path: path.join(screens, "cube-1280x800.png") });

  await page.locator("#help-toggle").click();
  assert.equal(await page.locator("#keyboard-help").evaluate(help => help.matches(":popover-open")), true);
  assert.match(await page.locator("#keyboard-help").textContent(), /Middle slices/);
  await page.keyboard.press("Escape");
  await page.locator("#stage").focus();
  await page.keyboard.press("?");
  assert.equal(await page.locator("#keyboard-help").evaluate(help => help.matches(":popover-open")), true);
  await page.keyboard.press("?");
  assert.equal(await page.locator("#keyboard-help").evaluate(help => help.matches(":popover-open")), false);
  for (const drag of [
    { face: "F", index: 5, dx: 0, dy: -110, move: "R", screenshot: "drag-1280x800.png" },
    { face: "F", index: 5, dx: 0, dy: 110, move: "R'" },
    { face: "F", index: 1, dx: -110, dy: 0, move: "U" },
    { face: "U", index: 7, dx: 240, dy: 0, move: "F2" },
    { face: "F", index: 3, dx: 0, dy: 110, move: "L" },
    { face: "F", index: 7, dx: 110, dy: 0, move: "D" },
    { face: "R", index: 5, dx: 0, dy: -110, move: "B" },
    { face: "F", index: 4, dx: 0, dy: 110, move: "M" },
    { face: "F", index: 4, dx: 110, dy: 0, move: "E" },
    { face: "R", index: 4, dx: 0, dy: 110, move: "S" },
    { face: "F", index: 5, dx: 0, dy: -18 },
    { face: "F", index: 5, dx: 0, dy: -110, cancel: true },
    { face: "F", index: 5, dx: 0, dy: -110, reverse: true }
  ]) {
    await page.locator("#reset").click();
    await page.locator("#home-view").click();
    await dragSticker(page, drag);
    if (!drag.move) assert.equal(await page.locator("#undo").isDisabled(), true, "cancellation adds no history");
  }
  const stage = await page.locator("#stage").boundingBox();
  const background = { x: stage.x + 12, y: stage.y + 12 };
  assert.equal(await page.evaluate(({ x, y }) => document.elementFromPoint(x, y).id, background), "stage");
  const beforeOrbit = await page.locator("#camera").getAttribute("style");
  await page.mouse.move(background.x, background.y);
  await page.mouse.down();
  await page.mouse.move(background.x + 80, background.y + 30, { steps: 10 });
  await page.mouse.up();
  assert.notEqual(await page.locator("#camera").getAttribute("style"), beforeOrbit);
  assert.equal(await page.locator("#cfen").inputValue(), solved, "background drag changes only the view");
  assert.equal(await page.locator("#undo").isDisabled(), true);
  await page.locator("#home-view").click();
  // Turn the cube around, then twist R from its back sticker. The screen
  // direction reverses while the move still follows the touched face.
  await page.mouse.move(background.x, background.y);
  await page.mouse.down();
  await page.mouse.move(background.x + 360, background.y, { steps: 15 });
  await page.mouse.up();
  await dragSticker(page, { face: "B", index: 3, dx: 0, dy: 110, move: "R" });
  await page.locator("#reset").click();
  await page.locator("#home-view").click();
  console.log("PASS browser: all face/slice sticker drags, prime/half turns, live preview, undo/redo, threshold/reversal/cancellation, background orbit and keyboard help");

  await page.locator("#scramble").click();
  await idle(page);
  assert.equal(await page.locator("#cube").getAttribute("data-solved"), "false");
  assert.equal((await page.locator("#scramble-text").textContent()).trim().split(/\s+/).length, 20);
  await page.locator("#solve").click();
  await page.locator("#playback").waitFor({ state: "visible" });
  await idle(page);
  assert.match(await page.locator("#sequence-title").textContent(), /kociemba$/);
  const count = Number(await page.locator("#scrubber").getAttribute("max"));
  assert.ok(count > 0 && count <= 30);
  await page.locator("#speed").selectOption("70");
  await page.locator("#step").click();
  await idle(page);
  assert.equal(await page.locator("#scrubber").inputValue(), "1");
  await page.locator("#play").click();
  await page.waitForFunction(() => document.getElementById("progress").textContent.split(" / ")[0] >= 2);
  await page.locator("#play").click();
  await idle(page);
  const paused = Number(await page.locator("#scrubber").inputValue());
  assert.ok(paused >= 2 && paused < count);
  await page.locator("#play").click();
  await page.waitForFunction(() => document.getElementById("cube").dataset.solved === "true" && !document.getElementById("reset").disabled, null, { timeout: 60000 });
  assert.match(await page.locator("#notice").textContent(), /All six faces are solved/);
  assert.equal(await page.locator("#scrubber").inputValue(), String(count));
  assert.equal(await page.locator("#cfen").inputValue(), solved);
  await page.locator("#scrubber").fill("0");
  assert.equal(await page.locator("#cube").getAttribute("data-solved"), "false");
  await page.locator("#scrubber").fill(String(count));
  assert.equal(await page.locator("#cube").getAttribute("data-solved"), "true");
  console.log("PASS browser: scramble → solve → step/play/pause/scrub → all six faces solved");

  for (const method of ["beginner", "kociemba", "cfop"]) {
    await page.locator("#reset").click();
    await runAlgorithm(page, method === "cfop" ? "R2 U F' D B2 L' U2 F R' D2 L B' U R2 F2 D' L2 U' B R" : "R U F2 L' B");
    await page.locator("#solve-method").selectOption(method);
    await page.locator("#solve").click();
    await idle(page);
    assert.ok((await page.locator("#sequence-moves button").count()) > 0);
    assert.match(await page.locator("#sequence-title").textContent(), new RegExp(`${method}$`));
    if (method === "cfop") {
      assert.deepEqual(await page.locator(".sequence-stage").evaluateAll(groups => groups.map(group => group.dataset.stage)), ["Cross", "F2L 1", "F2L 2", "F2L 3", "F2L 4", "OLL", "PLL"]);
      assert.equal(await page.locator(".sequence-stage .case-name").count(), 7);
      assert.match(await page.locator('[data-stage="OLL"] .case-name').textContent(), /OLL-/);
      assert.match(await page.locator('[data-stage="PLL"] .case-name').textContent(), /PLL-/);
      const buttons = page.locator("#sequence-moves button");
      await buttons.first().click();
      assert.equal(await page.locator("#scrubber").inputValue(), "1");
      assert.equal(await buttons.first().getAttribute("class"), "done");
      await page.locator("#back-step").click();
      assert.equal(await page.locator("#scrubber").inputValue(), "0");
      await page.locator("#playback").evaluate(element => element.scrollIntoView({ block: "end" }));
      await page.screenshot({ path: path.join(screens, "cfop-1280x800.png") });
      await page.setViewportSize({ width: 390, height: 844 });
      assert.equal(await page.evaluate(() => document.documentElement.scrollWidth > innerWidth), false);
      await page.locator("#playback").evaluate(element => element.scrollIntoView({ block: "start" }));
      await page.screenshot({ path: path.join(screens, "cfop-390x844.png") });
      await page.setViewportSize({ width: 1280, height: 800 });
    }
    await page.locator("#scrubber").fill(await page.locator("#scrubber").getAttribute("max"));
    assert.equal(await page.locator("#cfen").inputValue(), solved);
  }
  console.log("PASS browser: method selector, all solver headings and verified solution replay");

  // Force the default anytime budget on a hard physical state in a fresh
  // worker. Animation frames must continue throughout setup and synchronous
  // WASM search, and Solve must return a sequence that completes playback.
  await page.locator("#reset").click();
  await page.locator("#cfen").fill("YB|YGYOYRYBY/RYRBRGRWR/BYBOBRBWB/WBWOWRWGW/OYOGOBOWO/GYGRGOGWG");
  await page.locator("#import").click();
  await idle(page);
  await page.locator("#solve-method").selectOption("kociemba");
  await page.evaluate(() => {
    globalThis.solveFrames = 0;
    globalThis.countSolveFrames = true;
    const tick = () => { globalThis.solveFrames++; if (globalThis.countSolveFrames) requestAnimationFrame(tick); };
    requestAnimationFrame(tick);
  });
  const budgetStarted = Date.now();
  await page.locator("#solve").click();
  await idle(page);
  assert.ok(Date.now() - budgetStarted < 20000, "cold worker Solve must finish before the UI's 30-second timeout");
  const frames = await page.evaluate(() => { globalThis.countSolveFrames = false; return globalThis.solveFrames; });
  assert.ok(frames >= 5, `UI stalled while Solve ran: ${frames} animation frames`);
  assert.ok((await page.locator("#sequence-moves button").count()) > 0);
  await page.locator("#scrubber").fill(await page.locator("#scrubber").getAttribute("max"));
  assert.equal(await page.locator("#cfen").inputValue(), solved);
  console.log(`PASS browser: cold-worker budget solve and responsive UI (${frames} animation frames)`);

  await page.locator("#reset").click();
  await page.locator("#stage").focus();
  await page.keyboard.press("r");
  await idle(page);
  assert.equal(await page.locator("#cube").getAttribute("data-solved"), "false");
  await page.keyboard.press("Shift+R");
  await idle(page);
  assert.equal(await page.locator("#cfen").inputValue(), solved);
  await runAlgorithm(page, "M E S x y z Rw Uw' Fw2");
  assert.equal(await page.locator("#cube").getAttribute("data-solved"), "false");
  const after = await page.locator("#cfen").inputValue();
  await page.locator("#undo").click();
  assert.notEqual(await page.locator("#cfen").inputValue(), after);
  await page.locator("#redo").click();
  assert.equal(await page.locator("#cfen").inputValue(), after);
  await page.locator("#view-net").click();
  assert.equal(await page.locator(".net-face").count(), 6);
  assert.equal(await page.locator("#net span").count(), 54);
  await page.locator("#view-3d").click();
  const beforeImport = await page.locator("#cfen").inputValue();
  await page.locator("#cfen").fill("YB|W9/R9/B9/W9/O9/G9");
  await page.locator("#import").click();
  assert.equal(await page.locator("#cube").getAttribute("data-solved"), "false");
  assert.match(await page.locator("#notice").textContent(), /color|sticker|center/i);
  await page.locator("#cfen").fill(beforeImport);
  await page.locator("#import").click();
  assert.equal(await page.locator("#cfen").inputValue(), beforeImport);
  await page.locator("#export").click();
  assert.equal(await page.evaluate(() => globalThis.copiedText), beforeImport);
  console.log("PASS browser: keyboard/prime, advanced animated turns, undo/redo, net, CFEN import rejection/recovery");

  await page.locator("#reset").click();
  await runAlgorithm(page, "R U F2 L' B");
  await page.locator("#tab-lesson").click();
  await page.locator("#start-lesson").click();
  await page.locator("#lesson-content h3").waitFor();
  assert.ok((await page.locator("#lesson-content .lesson-action").count()) > 0);
  assert.match(await page.locator(".checkpoint-check").textContent(), /^Check:/);
  await page.locator("#play").click();
  await idle(page);
  await page.waitForFunction(() => document.getElementById("play").textContent === "Play");
  assert.match(await page.locator("#notice").textContent(), /Checkpoint reached/);
  await page.locator("#start-lesson").click();
  await page.locator("#lesson-content h3").waitFor();
  console.log("PASS browser: beginner lesson hints, checkpoint playback and next-state replanning");

  await page.locator("#reset").click();
  await runAlgorithm(page, "R U");
  await page.locator("#tab-search").click();
  await page.locator("#depth").fill("2");
  await page.locator("#find").click();
  await page.waitForFunction(() => document.getElementById("search-result").textContent.startsWith("Found"));
  assert.match(await page.locator("#search-result").textContent(), /Found 2 moves/);
  await page.locator("#play").click();
  await idle(page);
  await page.waitForFunction(() => document.getElementById("cube").dataset.solved === "true");

  await page.locator("#reset").click();
  await runAlgorithm(page, "R U F2 L' B D2 R' F U2 L");
  const deepStart = await page.locator("#cfen").inputValue();
  await page.locator("#tab-search").click();
  assert.equal(await page.locator("#depth").getAttribute("max"), "10");
  await page.locator("#depth").fill("9");
  await page.locator("#find").click();
  await idle(page);
  assert.match(await page.locator("#search-result").textContent(), /No path found within 9 moves/);
  assert.equal(await page.locator("#cfen").inputValue(), deepStart);
  await page.locator("#depth").fill("10");
  await page.locator("#find").click();
  await idle(page);
  assert.match(await page.locator("#search-result").textContent(), /Found 10 moves/);
  assert.equal(await page.locator("#sequence-kind").textContent(), "SEARCH RESULT");
  assert.equal(await page.locator("#cfen").inputValue(), deepStart);
  await page.screenshot({ path: path.join(screens, "search-1280x800.png") });
  await page.setViewportSize({ width: 390, height: 844 });
  assert.equal(await page.evaluate(() => document.documentElement.scrollWidth > innerWidth), false);
  await page.locator(".control-card").evaluate(element => element.scrollIntoView({ block: "start" }));
  await page.screenshot({ path: path.join(screens, "search-390x844.png") });
  await page.setViewportSize({ width: 1280, height: 800 });
  await page.locator("#play").click();
  await idle(page);
  await page.waitForFunction(() => document.getElementById("cube").dataset.solved === "true");
  assert.equal(await page.locator("#cfen").inputValue(), solved);
  console.log("PASS browser: shortest ten-move worker search, depth-nine exhaustion and result replay");

  // A physical scramble and reachable center frame keep the worker busy;
  // an impossible center target would now return immediately through IDA*.
  await page.locator("#reset").click();
  await runAlgorithm(page, "R2 U F' D B2 L' U2 F R' D2 L B' U R2 F2 D' L2 U' B R");
  const cancelStart = await page.locator("#cfen").inputValue();
  await page.locator("#tab-search").click();
  await page.locator("#find").click();
  await page.locator("#cancel-search").waitFor({ state: "visible" });
  // These controls run while the worker is searching, proving the UI responds.
  await page.locator("#view-net").click();
  assert.equal(await page.locator("#net").isVisible(), true);
  assert.equal(await page.locator("#solve-method").isDisabled(), true);
  await page.locator("#cancel-search").click();
  await idle(page);
  assert.match(await page.locator("#notice").textContent(), /canceled/);
  assert.equal(await page.locator("#cfen").inputValue(), cancelStart);
  assert.equal(await page.locator("#search-result").textContent(), "");
  assert.equal(await page.locator("#playback").isVisible(), false);
  assert.equal(await page.locator("#solve-method").isEnabled(), true);
  await page.locator("#view-3d").click();
  await page.locator("#depth").fill("11");
  await page.locator("#find").click();
  assert.match(await page.locator("#notice").textContent(), /0 to 10/);
  await page.locator("#target").fill("YB|?9/?9/?9/?9/?9/?9");
  await page.locator("#depth").fill("0");
  await page.locator("#find").click();
  await page.waitForFunction(() => document.getElementById("search-result").textContent.includes("already matches"));
  // Restart after cancellation and exercise a wildcard that needs a face turn.
  await page.locator("#reset").click();
  await runAlgorithm(page, "F");
  await page.locator("#tab-search").click();
  await page.locator("#target").fill("YB|Y9/?9/?9/?9/?9/?9");
  await page.locator("#depth").fill("1");
  await page.locator("#find").click();
  await idle(page);
  assert.match(await page.locator("#search-result").textContent(), /Found 1 moves/);
  await page.locator("#play").click();
  await idle(page);
  await page.waitForFunction(() => document.getElementById("cube").dataset.solved === "true");
  console.log("PASS browser: wildcard paths, responsive cancellation, unchanged state and fresh worker recovery");

  await page.goto(`http://127.0.0.1:${port}/web/#scramble=R+U&alg=U%27+R%27`);
  await idle(page);
  assert.equal(await page.locator("#algorithm").inputValue(), "U' R'");
  assert.equal(await page.locator("#cube").getAttribute("data-solved"), "false");
  await page.locator("#play").click();
  await idle(page);
  await page.waitForFunction(() => document.getElementById("cube").dataset.solved === "true");
  await page.locator("#share").click();
  assert.match(page.url(), /scramble=R\+U/);
  assert.match(page.url(), /alg=/);
  assert.equal(await page.evaluate(() => globalThis.copiedText), page.url());
  console.log("PASS browser: relative-path hosting and shareable scramble/algorithm hash");

  await page.goto(`http://127.0.0.1:${port}/web/`);
  await idle(page);
  await page.setViewportSize({ width: 390, height: 844 });
  assert.equal(await page.evaluate(() => document.documentElement.scrollWidth > innerWidth), false);
  assert.equal(await page.locator("#solve-method").inputValue(), "kociemba");
  await page.locator("#solve-method").selectOption("beginner");
  assert.equal(await page.locator("#solve-method").inputValue(), "beginner");
  await page.screenshot({ path: path.join(screens, "cube-390x844.png") });
  await dragSticker(page, { face: "F", index: 5, dx: 0, dy: -110, move: "R", screenshot: "drag-390x844.png" });
  await page.locator("#reset").click();
  await touchSticker(page, { face: "F", index: 5, dx: 0, dy: 110, move: "R'", screenshot: "touch-390x844.png" });
  await page.locator("#reset").click();
  await touchSticker(page, { face: "F", index: 5, dx: 0, dy: -110, cancel: true });
  assert.equal(await page.locator("#undo").isDisabled(), true, "touch cancellation adds no history");
  await page.locator('[data-move="F"]').click();
  await idle(page);
  assert.equal(await page.locator("#cube").getAttribute("data-solved"), "false");
  assert.deepEqual(errors, []);
  console.log("PASS browser: 390px phone layout, native touch turn/cancellation and no browser errors");
  console.log(`Screenshots: ${path.relative(root, screens)}/{cube,drag,search,cfop}-{1280x800,390x844}.png, touch-390x844.png`);
} finally {
  if (context) await context.close();
  if (server && server.exitCode === null) { server.kill("SIGTERM"); await once(server, "exit"); }
  await rm(profile, { recursive: true, force: true });
}
