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

const port = randomInt(49152, 65536);
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

try {
  if (server) await new Promise((resolve, reject) => {
    const timer = setTimeout(() => reject(new Error(`HTTP server did not start: ${serverLog}`)), 10000);
    server.stdout.on("data", chunk => { if (String(chunk).includes("Serving HTTP")) { clearTimeout(timer); resolve(); } });
    server.on("error", error => { clearTimeout(timer); reject(error); });
    server.on("exit", code => { clearTimeout(timer); reject(new Error(`HTTP server exited ${code}: ${serverLog}`)); });
  });
  context = await chromium.launchPersistentContext(profile, { executablePath: await cachedChromium(), headless: true, viewport: { width: 1280, height: 800 }, args: ["--disable-gpu"], reducedMotion: "no-preference" });
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
  await page.screenshot({ path: path.join(screens, "cube-1280x800.png") });

  await page.locator("#scramble").click();
  await idle(page);
  assert.equal(await page.locator("#cube").getAttribute("data-solved"), "false");
  assert.equal((await page.locator("#scramble-text").textContent()).trim().split(/\s+/).length, 20);
  await page.locator("#solve").click();
  await page.locator("#playback").waitFor({ state: "visible" });
  await idle(page);
  const count = Number(await page.locator("#scrubber").getAttribute("max"));
  assert.ok(count > 0);
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
  // An impossible all-yellow target makes a long search deterministic.
  await page.locator("#target").fill("YB|Y9/Y9/Y9/Y9/Y9/Y9");
  await page.locator("#depth").fill("8");
  await page.locator("#find").click();
  await page.locator("#cancel-search").waitFor({ state: "visible" });
  // These controls run while the worker is searching, proving the UI responds.
  await page.locator("#view-net").click();
  assert.equal(await page.locator("#net").isVisible(), true);
  await page.locator("#cancel-search").click();
  await idle(page);
  assert.match(await page.locator("#notice").textContent(), /canceled/);
  await page.locator("#view-3d").click();
  await page.locator("#target").fill("YB|?9/?9/?9/?9/?9/?9");
  await page.locator("#depth").fill("0");
  await page.locator("#find").click();
  await page.waitForFunction(() => document.getElementById("search-result").textContent.includes("already matches"));
  console.log("PASS browser: worker search, result replay, wildcard target and responsive cancellation");

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
  console.log("PASS browser: relative-path hosting and shareable scramble/algorithm hash");

  await page.goto(`http://127.0.0.1:${port}/web/`);
  await idle(page);
  await page.setViewportSize({ width: 390, height: 844 });
  assert.equal(await page.evaluate(() => document.documentElement.scrollWidth > innerWidth), false);
  await page.screenshot({ path: path.join(screens, "cube-390x844.png") });
  await page.locator('[data-move="F"]').click();
  await idle(page);
  assert.equal(await page.locator("#cube").getAttribute("data-solved"), "false");
  assert.deepEqual(errors, []);
  console.log("PASS browser: 390px phone layout, touch-sized controls and no browser errors");
  console.log(`Screenshots: ${path.relative(root, screens)}/cube-1280x800.png and cube-390x844.png`);
} finally {
  if (context) await context.close();
  if (server && server.exitCode === null) { server.kill("SIGTERM"); await once(server, "exit"); }
  await rm(profile, { recursive: true, force: true });
}
