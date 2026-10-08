import assert from "node:assert/strict";
import { spawn } from "node:child_process";
import { randomInt } from "node:crypto";
import { once } from "node:events";
import { existsSync } from "node:fs";
import { mkdir, mkdtemp, readFile, readdir, rm, writeFile } from "node:fs/promises";
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
  await page.waitForFunction(() => !document.getElementById("reset").disabled, null, { timeout: 120000 });
}

async function runAlgorithm(page, moves) {
  await page.locator("#tab-practice").click();
  await page.locator("#algorithm").fill(moves);
  await page.locator("#run-algorithm").click();
  await idle(page);
  await page.waitForFunction(() => document.getElementById("play").textContent === "Play");
}

async function savedStateRegressions(page, baseURL, solved) {
  const failures = [];
  const stickers = () => page.locator("#cube .sticker").evaluateAll(nodes => nodes.map(node =>
    `${node.dataset.face}:${node.dataset.index}:${node.dataset.color}`).sort());
  const checks = [
    ["blank/invalid CFEN preserves the cube and history", async () => {
      await page.goto(`${baseURL}?import`);
      await idle(page);
      await page.locator("#speed").selectOption("70", { force: true });
      await runAlgorithm(page, "R U");
      const mixed = await page.locator("#cfen").inputValue();
      const visible = await stickers();
      const savedURL = page.url();
      for (const input of ["", "   ", "not a CFEN", "YB|W9/R9/B9/W9/O9/G9"]) {
        await page.locator("#cfen").fill(input);
        await page.locator("#import").click();
        assert.equal(await page.locator("#notice").getAttribute("class"), "notice error", `reject ${JSON.stringify(input)}`);
        assert.ok((await page.locator("#notice").textContent()).trim());
        assert.deepEqual(await stickers(), visible, "failed import preserves visible stickers");
        assert.equal(page.url(), savedURL, "failed import preserves saved state");
        assert.equal(await page.locator("#progress").textContent(), "2 / 2");
        assert.equal(await page.locator("#undo").isEnabled(), true);
        assert.equal(await page.locator("#redo").isEnabled(), false);
      }
      await page.locator("#undo").click();
      const afterR = await page.evaluate(() => JSON.parse(globalThis.cubeAPI(JSON.stringify({ op: "twist", moves: "R" }))).data.state.cfen);
      assert.equal(await page.locator("#cfen").inputValue(), afterR);
      await page.locator("#undo").click();
      assert.equal(await page.locator("#cfen").inputValue(), solved);
      await page.locator("#redo").click();
      await page.locator("#redo").click();
      assert.equal(await page.locator("#cfen").inputValue(), mixed, "both turns remain redoable");
    }],
    ["edited algorithm drafts survive share/reload beside prepared playback", async () => {
      await page.goto(`${baseURL}?draft`);
      await idle(page);
      await page.locator("#speed").selectOption("70", { force: true });
      await runAlgorithm(page, "R U");
      const mixed = await page.locator("#cfen").inputValue();
      for (const draft of [" F\n", "R U (", ""]) {
        await page.locator("#algorithm").fill(draft);
        await page.locator("#share").click();
        const shared = await page.evaluate(() => globalThis.copiedText);
        assert.equal(new URLSearchParams(new URL(shared).hash.slice(1)).get("draft"), draft, "share includes the exact visible draft");
        await page.reload();
        await idle(page);
        assert.equal(await page.locator("#algorithm").inputValue(), draft);
        assert.equal(await page.locator("#cfen").inputValue(), mixed);
        assert.equal(await page.locator("#progress").textContent(), "2 / 2");
        assert.deepEqual(await page.locator("#sequence-moves button").allTextContents(), ["R", "U"]);
        await page.goto(`${baseURL}?fresh`);
        await idle(page);
        await page.goto(shared);
        await idle(page);
        assert.equal(await page.locator("#algorithm").inputValue(), draft, "a shared link restores the draft too");
        assert.equal(await page.locator("#cfen").inputValue(), mixed);
      }
      await page.locator("#cube").evaluate(root => root.removeAttribute("data-measure-frames"));
      await page.locator("#speed").selectOption("70");
      await runAlgorithm(page, "F");
      const afterF = await page.evaluate(cfen => JSON.parse(globalThis.cubeAPI(JSON.stringify({ op: "twist", cfen, moves: "F" }))).data.state.cfen, mixed);
      assert.equal(await page.locator("#cfen").inputValue(), afterF, "Play above the draft runs the edited text");
      await page.goto(baseURL);
      await idle(page);
      await page.locator("#algorithm").fill("R U (");
      await page.reload();
      await idle(page);
      assert.equal(await page.locator("#algorithm").inputValue(), "R U (", "an unprepared incomplete draft also survives");
      assert.equal(await page.locator("#cfen").inputValue(), solved);
      assert.equal(await page.locator("#playback").isVisible(), false);
    }],
    ["reload restores the first committed scramble turn exactly", async () => {
      for (const width of [1280, 390]) {
        await page.setViewportSize({ width, height: width === 1280 ? 800 : 844 });
        await page.goto(`${baseURL}?scramble=${width}`);
        await idle(page);
        // Capture the first committed render and stop at the next animation.
        // This avoids racing a later turn between reading CFEN and reloading.
        await page.evaluate(() => {
          const target = document.getElementById("cube");
          const observer = new MutationObserver(() => {
            if (target.dataset.solved !== "false") return;
            observer.disconnect();
            Element.prototype.animate = () => ({ finished: new Promise(() => {}) });
            // Allow the completed turn's continuation to save its frame first.
            queueMicrotask(() => {
              globalThis.partialScramble = {
                cfen: document.getElementById("cfen").value,
                url: location.href,
                stickers: [...target.querySelectorAll(".sticker")].map(node => `${node.dataset.face}:${node.dataset.index}:${node.dataset.color}`).sort()
              };
            });
          });
          observer.observe(target, { attributes: true, childList: true });
        });
        await page.locator("#scramble").click();
        await page.waitForFunction(() => !!globalThis.partialScramble);
        const partial = await page.evaluate(() => globalThis.partialScramble);
        assert.notEqual(partial.cfen, solved);
        const expected = await page.evaluate(() => JSON.parse(globalThis.cubeAPI(JSON.stringify({ op: "twist", moves: document.getElementById("scramble-text").textContent.split(" ")[0] }))).data.state.cfen);
        assert.equal(partial.cfen, expected, "snapshot is exactly the first scramble turn");
        assert.equal(new URLSearchParams(new URL(partial.url).hash.slice(1)).get("current"), partial.cfen, "the committed frame is saved while Scramble is busy");
        await page.reload();
        await idle(page);
        assert.equal(await page.locator("#cfen").inputValue(), partial.cfen);
        assert.deepEqual(await stickers(), partial.stickers);
      }
      await page.setViewportSize({ width: 1280, height: 800 });
    }],
    ["restored lessons retain instructions/checks and can continue", async () => {
      await page.setViewportSize({ width: 1280, height: 800 });
      await page.goto(`${baseURL}?lesson#scramble=F+R+U+B+L+D`);
      await idle(page);
      await page.locator("#tab-lesson").click();
      await page.locator("#start-lesson").click();
      await idle(page);
      const lesson = await page.locator("#lesson-content").textContent();
      const moves = await page.locator("#sequence-moves button").allTextContents();
      assert.ok(moves.length > 1);
      await page.locator("#step").click();
      await idle(page);
      const checkpoint = await page.locator("#cfen").inputValue();
      const shared = page.url();
      await page.reload();
      await idle(page);
      assert.equal(await page.locator("#cfen").inputValue(), checkpoint);
      assert.equal(await page.locator("#progress").textContent(), `1 / ${moves.length}`);
      assert.equal(await page.locator("#lesson-content").textContent(), lesson);
      assert.equal(await page.locator("#tab-lesson").getAttribute("aria-selected"), "true");
      assert.equal(await page.locator(".checkpoint-check").isVisible(), true);
      assert.equal(await page.locator("#lesson-content .lesson-action").count() > 0, true);
      await page.locator(".control-card").evaluate(node => node.scrollIntoView({ block: "start" }));
      await page.screenshot({ path: path.join(screens, "lesson-restored-1280x800.png") });
      await page.setViewportSize({ width: 390, height: 844 });
      await page.reload();
      await idle(page);
      assert.equal(await page.locator("#lesson-content").textContent(), lesson);
      assert.equal(await page.locator("#cfen").inputValue(), checkpoint);
      assert.equal(await page.evaluate(() => document.documentElement.scrollWidth > innerWidth), false);
      await page.locator("#lesson").evaluate(node => node.scrollIntoView({ block: "start" }));
      await page.screenshot({ path: path.join(screens, "lesson-restored-390x844.png") });
      await page.setViewportSize({ width: 1280, height: 800 });
      await page.locator("#cube").evaluate(root => root.removeAttribute("data-measure-frames"));
      await page.locator("#speed").selectOption("70");
      await page.locator("#play").click();
      await idle(page);
      await page.waitForFunction(() => document.getElementById("play").textContent === "Play");
      assert.match(await page.locator("#notice").textContent(), /Compare the check above/);
      assert.equal(await page.locator(".checkpoint-check").isVisible(), true);
      await page.locator("#lesson-content button").click();
      await idle(page);
      assert.equal(await page.locator(".checkpoint-check").isVisible(), true, "restored Next hint replans from the new cube");
      await page.goto(`${baseURL}?fresh`);
      await idle(page);
      await page.goto(shared);
      await idle(page);
      assert.equal(await page.locator("#lesson-content").textContent(), lesson, "shared checkpoints retain their lesson too");
      assert.equal(await page.locator("#cfen").inputValue(), checkpoint);
    }]
  ];
  for (const [name, check] of checks) {
    try { await check(); console.log(`PASS browser: ${name}`); }
    catch (error) { failures.push(new Error(name, { cause: error })); console.error(`FAIL browser: ${name}: ${error.message}`); }
  }
  if (failures.length) throw new AggregateError(failures, "Saved-state browser regressions failed");
  await page.setViewportSize({ width: 1280, height: 800 });
  await page.goto(baseURL);
  await idle(page);
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

async function blackInterior(page) {
  assert.equal(await page.locator("#cube .face:empty").evaluateAll(faces =>
    faces.length > 0 && faces.every(face => getComputedStyle(face).backgroundColor === "rgb(17, 17, 17)")), true,
  "exposed interior surfaces are black");
  assert.equal(await page.locator("#cube .sticker").evaluateAll(stickers =>
    stickers.every(sticker => getComputedStyle(sticker).backfaceVisibility === "hidden")), true,
  "colored sticker backs do not show through the turning layer");
}

async function dragSticker(page, { face, index, dx, dy, move, screenshot, cancel = false, reverse = false, tiles = null }) {
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
  if (tiles === null) assert.equal(await page.locator("#cube .layer .cubie").count(), 9);
  else assert.equal(await page.locator("#cube .svg-tile[data-turning]").count(), tiles);
  if (tiles === null) assert.notEqual(await page.locator("#cube .layer").evaluate(layer => getComputedStyle(layer).transform), "none");
  else assert.notEqual(await page.locator("#cube svg").getAttribute("data-turn-angle"), "0");
  await blackInterior(page);
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
  assert.equal(await page.locator("#cube [data-turning]").count(), 0, "turn preview is cleaned up");
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

async function nxnRegressions(page) {
  const metrics = [];
  for (const size of [2, 3, 4, 5, 6, 7]) {
    if (size === 7) await page.setViewportSize({ width: 390, height: 844 });
    await page.locator("#size").selectOption(String(size));
    await idle(page);
    const home = await page.locator("#cfen").inputValue();
    assert.equal(await page.locator("#cube").getAttribute("data-size"), String(size));
    assert.equal(await page.locator("#cube .sticker").count(), 6 * size * size);
    await page.locator("#stage").evaluate(node => node.scrollIntoView({ block: "center" }));
    await stickerPoint(page, "U", Math.floor((size - 1) / 2) * size + Math.floor((size - 1) / 2));
    await page.locator("#view-net").click();
    assert.equal(await page.locator("#net span").count(), 6 * size * size);
    assert.equal(await page.locator(".net-face.F").evaluate(node => getComputedStyle(node).gridTemplateColumns.split(" ").length), size);
    await page.locator("#view-3d").click();
    assert.equal(await page.locator("#tab-lesson").isDisabled(), size !== 3);
    assert.equal(await page.locator("#tab-search").isDisabled(), size !== 3);
    assert.equal(await page.locator("#size-note").isVisible(), size !== 3);
    assert.equal(await page.locator('[data-move="M"]').isDisabled(), size % 2 === 0);
    await page.locator("#stage").focus();
    await page.keyboard.press("2");
    await page.keyboard.press("r");
    await idle(page);
    const afterInner = await page.evaluate(({ cfen }) => JSON.parse(globalThis.cubeAPI(JSON.stringify({ op: "twist", cfen, moves: "2R" }))).data.state.cfen, { cfen: home });
    assert.equal(await page.locator("#cfen").inputValue(), afterInner, `${size}x${size} numbered keyboard turn`);
    await page.keyboard.press("Shift+R");
    await idle(page);
    assert.equal(await page.locator("#cfen").inputValue(), home);
    await page.locator("#turn-width").selectOption("wide");
    await page.locator('[data-move="F"]').click();
    await idle(page);
    const afterWide = await page.evaluate(cfen => JSON.parse(globalThis.cubeAPI(JSON.stringify({ op: "twist", cfen, moves: "2Fw" }))).data.state.cfen, home);
    assert.equal(await page.locator("#cfen").inputValue(), afterWide, `${size}x${size} wide button turn`);
    await page.locator("#turn-width").selectOption("single");
    await page.locator("#turn-layer").selectOption("1");
    await page.locator("#reset").click();
    if (size === 5) {
      await dragSticker(page, { face: "F", index: 13, dx: 0, dy: -110, move: "2R", tiles: 20, screenshot: "5x5-inner-drag-1280x800.png" });
      await page.locator("#reset").click();
    }
    await page.locator("#scramble").click();
    await idle(page);
    assert.equal(await page.locator("#cube").getAttribute("data-solved"), "false");
    const scrambled = await page.locator("#cfen").inputValue();
    if (size === 3) await page.locator("#solve-method").selectOption("kociemba");
    await page.evaluate(() => {
      globalThis.nxnSolveFrames = 0;
      globalThis.nxnSolveCounting = true;
      const tick = () => { globalThis.nxnSolveFrames++; if (globalThis.nxnSolveCounting) requestAnimationFrame(tick); };
      requestAnimationFrame(tick);
    });
    const started = Date.now();
    await page.locator("#solve").click();
    await idle(page);
    const frames = await page.evaluate(() => { globalThis.nxnSolveCounting = false; return globalThis.nxnSolveFrames; });
    const count = Number(await page.locator("#scrubber").getAttribute("max"));
    assert.ok(frames >= 5, `${size}x${size} UI keeps updating during solve`);
    assert.ok(count > 0);
    assert.match(await page.locator("#sequence-title").textContent(), size === 3 ? /kociemba$/ : /reduction$/);
    metrics.push({ size, viewport: page.viewportSize(), moves: count, wasmSolveMs: Number(await page.locator("#cube").getAttribute("data-solve-ms")), coldWorkerMs: Number(await page.locator("#cube").getAttribute("data-solve-wall-ms")), readyMs: Date.now() - started, responsiveFrames: frames });
    await page.locator("#speed").selectOption("70");
    await page.locator("#step").click();
    await idle(page);
    assert.equal(await page.locator("#scrubber").inputValue(), "1");
    await page.locator("#back-step").click();
    assert.equal(await page.locator("#cfen").inputValue(), scrambled, "back-step restores the solution's starting cube");
    await page.locator("#scrubber").fill(String(Math.floor(count / 2)));
    await page.locator("#view-net").click();
    await page.locator("#scrubber").fill(String(Math.floor(count / 2) + 1));
    await page.locator("#view-3d").click();
    await page.locator("#stage").evaluate(node => node.scrollIntoView({ block: "center" }));
    await stickerPoint(page, "U", Math.floor((size - 1) / 2) * size + Math.floor((size - 1) / 2));
    if (size === 5 || size === 7) {
      for (const [width, height] of [[1280, 800], [390, 844]]) {
        await page.setViewportSize({ width, height });
        await page.locator(".size-picker").evaluate(node => node.scrollIntoView({ block: "start" }));
        assert.equal(await page.evaluate(() => document.documentElement.scrollWidth > innerWidth), false, `${size}x${size} ${width}px layout`);
        await page.screenshot({ path: path.join(screens, `${size}x${size}-mid-solve-${width}x${height}.png`) });
      }
      await page.setViewportSize({ width: 1280, height: 800 });
    }
    if (size === 7) {
      await page.setViewportSize({ width: 390, height: 844 });
      await page.locator("#speed").selectOption("460");
      await page.locator("#stage").evaluate(node => node.scrollIntoView({ block: "center" }));
      const baseline = await page.evaluate(() => new Promise(resolve => {
        const times = []; let last;
        const tick = now => { if (last) times.push(now - last); last = now; if (times.length < 20) requestAnimationFrame(tick); else resolve(times); };
        requestAnimationFrame(tick);
      }));
      metrics.at(-1).staticFrameMs = baseline.reduce((a, b) => a + b) / baseline.length;
      await page.locator("#cube").evaluate(root => root.dataset.measureFrames = "true");
      const intervals = [], renderTimes = [];
      // Sample several turns; a background-priority headless process can skip
      // compositor ticks even on a static cube. Keep that pacing separate from
      // the actual projection and SVG layout cost, and report both.
      for (let turn = 0; turn < 5; turn++) {
        await page.locator("#step").evaluate(button => button.click());
        await idle(page);
        const samples = await page.locator("#cube svg").evaluate(svg => ({
          render: JSON.parse(svg.dataset.frameTimes), intervals: JSON.parse(svg.dataset.frameIntervals)
        }));
        renderTimes.push(...samples.render);
        intervals.push(...samples.intervals);
      }
      const summary = times => {
        const sorted = [...times].sort((a, b) => a - b);
        return { samples: times.length, meanMs: times.reduce((a, b) => a + b) / times.length,
          p95Ms: sorted[Math.floor(sorted.length * .95)], maxMs: sorted.at(-1) };
      };
      assert.ok(intervals.length >= 10, "7x7 animation continues across phone-width turns");
      metrics.at(-1).phoneFrameInterval = summary(intervals);
      metrics.at(-1).phoneRenderTime = summary(renderTimes);
      assert.ok(metrics.at(-1).phoneRenderTime.meanMs < 40, `7x7 frame work averaged ${metrics.at(-1).phoneRenderTime.meanMs.toFixed(1)} ms`);
      await page.locator("#cube").evaluate(root => root.removeAttribute("data-measure-frames"));
      await page.locator("#speed").selectOption("70");
      await page.setViewportSize({ width: 1280, height: 800 });
    }
    await page.locator("#scrubber").fill(String(count - Math.min(3, count)));
    await page.locator("#play").click();
    await page.waitForFunction(() => document.getElementById("cube").dataset.solved === "true" && !document.getElementById("reset").disabled);
    assert.equal(await page.locator("#scrubber").inputValue(), String(count));
    assert.equal(await page.locator("#cube .sticker").evaluateAll(nodes => new Set(nodes.filter(node => node.dataset.face === "F").map(node => node.dataset.color)).size), 1);
    console.log(`PASS browser ${size}x${size}: picker, net, keyboard/layer/wide, scramble, solve, step/back/scrub/play (${count} moves)`);
  }

  await page.locator("#size").selectOption("4");
  await page.locator("#speed").selectOption("460", { force: true });
  const moves = "2R U 2F Lw D B 2U Rw F2 L 2B D2";
  await page.locator("#algorithm").fill(moves);
  await page.locator("#run-algorithm").click();
  await page.waitForFunction(() => document.getElementById("progress").textContent === "2 / 12");
  const snapshot = await page.evaluate(() => ({ cfen: document.getElementById("cfen").value, url: location.href,
    stickers: [...document.querySelectorAll("#cube .sticker")].map(node => `${node.dataset.face}:${node.dataset.index}:${node.dataset.color}`).sort() }));
  await page.reload();
  await idle(page);
  assert.equal(await page.locator("#size").inputValue(), "4");
  assert.equal(await page.locator("#cfen").inputValue(), snapshot.cfen);
  assert.equal(await page.locator("#progress").textContent(), "2 / 12");
  assert.equal(await page.locator("#play").textContent(), "Play");
  assert.deepEqual(await page.locator("#cube .sticker").evaluateAll(nodes => nodes.map(node => `${node.dataset.face}:${node.dataset.index}:${node.dataset.color}`).sort()), snapshot.stickers);
  await page.locator("#share").click();
  assert.equal(await page.evaluate(() => globalThis.copiedText), snapshot.url);
  await page.locator("#size").selectOption("2");
  await page.goto(snapshot.url);
  await idle(page);
  assert.equal(await page.locator("#size").inputValue(), "4");
  assert.equal(await page.locator("#cfen").inputValue(), snapshot.cfen);
  assert.deepEqual(await page.locator("#sequence-moves button").allTextContents(), moves.split(" "));
  assert.equal(await page.locator("#scrubber").inputValue(), "2");
  await page.locator("#step").click();
  await idle(page);
  const third = await page.evaluate(cfen => JSON.parse(globalThis.cubeAPI(JSON.stringify({ op: "twist", cfen, moves: "2F" }))).data.state.cfen, snapshot.cfen);
  assert.equal(await page.locator("#cfen").inputValue(), third);
  console.log("PASS browser 4x4: active-playback share/reload restores size, stickers, sequence, playhead and continuation");
  console.log(`NxN performance: ${JSON.stringify(metrics)}`);
  await writeFile(path.join(scratch, "nxn-browser-metrics.json"), JSON.stringify(metrics, null, 2) + "\n");
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

  await savedStateRegressions(page, `http://127.0.0.1:${port}/web/`, solved);

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
      const cfopTitle = await page.locator("#sequence-title").textContent();
      await page.reload();
      await idle(page);
      assert.equal(await page.locator("#sequence-title").textContent(), cfopTitle);
      assert.equal(await page.locator("#sequence-kind").textContent(), "SOLUTION");
      assert.equal(await page.locator(".sequence-stage").count(), 7, "CFOP stages survive reload");
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

  await page.locator("#speed").selectOption("460");
  await page.locator("#reset").click();
  await page.locator("#tab-practice").click();
  const playbackMoves = "R U F L D B R U F L D B";
  await page.locator("#algorithm").fill(playbackMoves);
  await page.locator("#run-algorithm").click();
  await page.waitForFunction(() => document.getElementById("progress").textContent === "2 / 12");
  const snapshot = await page.evaluate(() => ({
    cfen: document.getElementById("cfen").value,
    stickers: [...document.querySelectorAll("#cube .sticker")].map(sticker => `${sticker.dataset.face}:${sticker.dataset.index}:${sticker.dataset.color}`).sort(),
    url: location.href,
    progress: document.getElementById("progress").textContent,
    playing: document.getElementById("play").textContent
  }));
  await page.reload();
  await idle(page);
  assert.equal(snapshot.progress, "2 / 12");
  assert.equal(snapshot.playing, "Pause", "reload interrupts active playback");
  const { cfen: playbackState, stickers: playbackStickers, url: savedPlaybackURL } = snapshot;
  const savedParams = new URL(savedPlaybackURL).hash.slice(1);
  assert.equal(new URLSearchParams(savedParams).get("current"), playbackState);
  assert.equal(new URLSearchParams(savedParams).get("index"), "2");
  assert.equal(await page.locator("#cfen").inputValue(), playbackState, "reload restores the current cube");
  assert.deepEqual(await page.locator("#cube .sticker").evaluateAll(stickers => stickers.map(sticker => `${sticker.dataset.face}:${sticker.dataset.index}:${sticker.dataset.color}`).sort()), playbackStickers);
  assert.equal(await page.locator("#algorithm").inputValue(), playbackMoves);
  assert.deepEqual(await page.locator("#sequence-moves button").allTextContents(), playbackMoves.split(" "));
  assert.equal(await page.locator("#progress").textContent(), "2 / 12");
  assert.equal(await page.locator("#scrubber").inputValue(), "2");
  assert.equal(await page.locator("#play").textContent(), "Play", "restored playback starts paused");
  await page.locator("#share").click();
  assert.equal(await page.evaluate(() => globalThis.copiedText), savedPlaybackURL);
  await page.locator("#reset").click();
  await page.goto(savedPlaybackURL);
  await idle(page);
  assert.equal(await page.locator("#cfen").inputValue(), playbackState, "share restores the current cube");
  assert.equal(await page.locator("#progress").textContent(), "2 / 12");
  const expectedPlayback = await page.evaluate(cfen => JSON.parse(globalThis.cubeAPI(JSON.stringify({ op: "twist", cfen, moves: "R U" }))).data.state.cfen, solved);
  assert.equal(playbackState, expectedPlayback, "the restored cube is exactly R U from solved");
  await page.locator("#scrubber").fill("0");
  assert.equal(await page.locator("#cfen").inputValue(), solved, "restored sequence retains its starting cube");
  await page.locator("#scrubber").fill("2");
  await page.locator("#step").click();
  await idle(page);
  const thirdFrame = await page.evaluate(cfen => JSON.parse(globalThis.cubeAPI(JSON.stringify({ op: "twist", cfen, moves: "F" }))).data.state.cfen, playbackState);
  assert.equal(await page.locator("#cfen").inputValue(), thirdFrame, "restored playback continues with the third move");
  assert.equal(await page.locator("#progress").textContent(), "3 / 12");
  await page.locator("#speed").selectOption("70");
  console.log("PASS browser: mid-playback reload/share restores cube stickers, sequence, playhead and continuation");

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
  await page.setViewportSize({ width: 1280, height: 800 });
  await nxnRegressions(page);
  assert.deepEqual(errors, []);
  console.log(`Screenshots: ${path.relative(root, screens)}/{cube,drag,search,cfop,lesson-restored}-{1280x800,390x844}.png, touch-390x844.png`);
} finally {
  if (context) await context.close();
  if (server && server.exitCode === null) { server.kill("SIGTERM"); await once(server, "exit"); }
  await rm(profile, { recursive: true, force: true });
}
