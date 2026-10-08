import assert from "node:assert/strict";
import { existsSync } from "node:fs";
import { lstat, mkdir, mkdtemp, readFile, readdir, rm, writeFile } from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { brotliCompressSync, constants, gzipSync } from "node:zlib";
import { chromium } from "playwright";
import { exportPages, readRuntimeAssets } from "../build-pages.mjs";

const root = fileURLToPath(new URL("../../", import.meta.url));
const scratch = path.join(root, ".scratch");
await mkdir(scratch, { recursive: true });
process.env.TMPDIR = process.env.TMP = process.env.TEMP = scratch;
const temporary = await mkdtemp(path.join(scratch, "pages-test-"));
const site = "https://ehrlich-b.github.io/cube/";
const mime = { ".html": "text/html", ".css": "text/css", ".js": "text/javascript", ".json": "application/json", ".wasm": "application/wasm", ".svg": "image/svg+xml", ".gz": "application/octet-stream" };

async function executable() {
  if (process.platform !== "darwin" && existsSync(chromium.executablePath())) return chromium.executablePath();
  const cache = path.join(os.homedir(), "Library/Caches/ms-playwright");
  const versions = (await readdir(cache)).filter(name => /^chromium(_headless_shell)?-\d+$/.test(name)).sort((a, b) => Number(b.includes("headless_shell")) - Number(a.includes("headless_shell")) || Number(b.match(/\d+$/)[0]) - Number(a.match(/\d+$/)[0]));
  for (const version of versions) for (const binary of ["chrome-headless-shell-mac-arm64/chrome-headless-shell", "chrome-headless-shell-mac-arm64/headless_shell", "chrome-mac-arm64/Google Chrome for Testing.app/Contents/MacOS/Google Chrome for Testing", "chrome-mac/Chromium.app/Contents/MacOS/Chromium"]) {
    const candidate = path.join(cache, version, binary);
    if (existsSync(candidate)) return candidate;
  }
  throw new Error("No cached Playwright Chromium; provision it before running test-pages.");
}

async function artifact(directory) {
  const files = new Map();
  for (const name of await readdir(directory)) {
    assert.ok((await lstat(path.join(directory, name))).isFile(), `Pages contains only regular runtime files: ${name}`);
    files.set(name, await readFile(path.join(directory, name)));
  }
  assert.ok(files.has("index.html"), "Pages entry point exists");
  return files;
}

function checkPaths(files) {
  const references = new Set(["index.html", ".nojekyll"]);
  for (const [name, bytes] of files) {
    if (!/\.(html|css|js|svg)$/.test(name)) continue;
    const text = bytes.toString("utf8");
    const paths = name.endsWith("html") ? [...text.matchAll(/\b(?:src|href)="([^"]+)"/g)].map(match => match[1]) :
      [...text.matchAll(/^\s*import\s+(?:[^"'\n]+\s+from\s+)?["']([^"']+)["']/gm), ...text.matchAll(/\bnew URL\(\s*["']([^"']+)["']/g), ...text.matchAll(/url\(\s*["']?([^\s)'";]+)/g)].map(match => match[1]);
    for (const reference of paths) {
      assert.ok(reference.startsWith("./"), `${name}: relative asset/link ${reference}`);
      if (reference === "./") continue;
      const url = new URL(reference, site);
      assert.equal(url.origin, new URL(site).origin);
      assert.ok(url.pathname.startsWith("/cube/"), `${reference} stays in the project subpath`);
      const asset = url.pathname.slice("/cube/".length);
      assert.ok(files.has(asset), `${name} references exported asset ${asset}`);
      references.add(asset);
    }
    assert.doesNotMatch(text, /SharedArrayBuffer|crossOriginIsolated|Atomics\./, `${name} has no shared-memory/isolation dependency`);
  }
  assert.deepEqual([...files.keys()].sort(), [...references].sort(), "Export contains exactly the linked runtime graph and .nojekyll");
  const { version } = JSON.parse(files.get("version.json"));
  assert.match(version, /^[a-f0-9]{16}$/, "Deployment metadata contains the runtime fingerprint");
  assert.ok(files.has(`app.${version}.js`), "Deployment metadata matches the running app");
}

function sizes(files, optimized) {
  const result = { total: { raw: 0, gzip: 0, brotli: 0 }, initial: { raw: 0, gzip: 0, brotli: 0 }, lazy: { raw: 0, gzip: 0, brotli: 0 }, lazyAssets: {} };
  for (const [name, body] of files) {
    if (name === ".nojekyll") continue;
    const size = { raw: body.length, gzip: gzipSync(body, { level: 9 }).length,
      brotli: brotliCompressSync(body, { params: { [constants.BROTLI_PARAM_QUALITY]: 11 } }).length };
    for (const kind of Object.keys(size)) {
      result.total[kind] += size[kind];
      result[name.endsWith(".gz") ? "lazy" : "initial"][kind] += size[kind];
    }
    if (name.endsWith(".gz")) result.lazyAssets[name.replace(/\.[a-f0-9]{16}(?=\.)/, "")] = size;
    if (name.endsWith(".wasm")) result.wasm = size;
  }
  assert.ok(result.initial.raw < (optimized ? 4_800_000 : 5_150_000), "Initial Pages bytes stay within the optimized/portable build budget");
  assert.ok(result.initial.gzip < 1_500_000, "Initial Pages gzip stays under 1.5 MB");
  assert.ok(result.initial.brotli < 1_150_000, "Initial Pages Brotli stays under 1.15 MB");
  assert.ok(result.wasm.raw < (optimized ? 4_700_000 : 5_050_000), "WASM stays within the optimized/portable build budget");
  assert.ok(result.lazy.raw < 4_600_000, "All lazy solver assets stay under 4.6 MB");
  assert.ok(result.total.raw < (optimized ? 9_400_000 : 9_750_000), "Total Pages artifact stays within its build budget");
  return result;
}

async function idle(page, size = 3) {
  await page.waitForFunction(size => !document.getElementById("reset")?.disabled && document.querySelectorAll(".sticker").length === 6 * size * size, size, { timeout: size === 3 ? 60000 : 120000 });
}

// Small, dependency-free accessibility gate: rendered controls only, computed
// colors composited through ancestors, and real keyboard/focus checks below.
async function accessibility(page) {
  return page.evaluate(() => {
    const visible = node => node.checkVisibility() && !node.closest("[inert]");
    const controls = [...document.querySelectorAll("button,a,input,textarea,select,summary,[tabindex]")].filter(node => visible(node) && !node.disabled);
    const color = value => value.match(/[\d.]+/g)?.map(Number) || [0, 0, 0, 0];
    const blend = (front, back) => {
      const alpha = front[3] ?? 1;
      return front.slice(0, 3).map((channel, index) => channel * alpha + back[index] * (1 - alpha));
    };
    const background = node => {
      const ancestors = [];
      for (let parent = node; parent; parent = parent.parentElement) ancestors.unshift(parent);
      return ancestors.reduce((back, parent) => blend(color(getComputedStyle(parent).backgroundColor), back), [255, 255, 255]);
    };
    const luminance = rgb => rgb.map(channel => channel / 255).map(channel => channel <= .04045 ? channel / 12.92 : ((channel + .055) / 1.055) ** 2.4).reduce((sum, channel, index) => sum + channel * [.2126, .7152, .0722][index], 0);
    const ratio = (a, b) => (Math.max(luminance(a), luminance(b)) + .05) / (Math.min(luminance(a), luminance(b)) + .05);
    const name = node => node.getAttribute("aria-label") || node.getAttribute("aria-labelledby")?.split(/\s+/).map(id => document.getElementById(id)?.textContent || "").join(" ") || [...(node.labels || [])].map(label => label.textContent).join(" ") || (node.matches("button,a,summary") ? node.textContent : "");
    const failures = [];
    let minimum = Infinity;
    for (const node of controls) {
      const selector = node.id ? `#${node.id}` : `${node.tagName.toLowerCase()} ${node.textContent.trim().slice(0, 30)}`;
      if (!name(node)?.trim()) failures.push(`${selector}: accessible name missing`);
      if (node.tabIndex > 0) failures.push(`${selector}: positive tabindex disrupts DOM focus order`);
      const style = getComputedStyle(node), back = background(node);
      const contrast = ratio(blend(color(style.color), back), back);
      minimum = Math.min(minimum, contrast);
      const large = parseFloat(style.fontSize) >= 24 || (parseFloat(style.fontSize) >= 18.66 && parseInt(style.fontWeight) >= 700);
      if (contrast < (large ? 3 : 4.5)) failures.push(`${selector}: text contrast ${contrast.toFixed(2)}:1`);
      // The 3px outline offset puts the ring on the surrounding card, outside
      // the control's own fill (especially the dark green primary buttons).
      if (node === document.activeElement && (style.outlineStyle === "none" || parseFloat(style.outlineWidth) < 2 || ratio(color(style.outlineColor), background(node.parentElement)) < 3)) failures.push(`${selector}: keyboard focus outline needs 3:1 contrast and 2px width`);
    }
    return { failures, minimumContrast: Number(minimum.toFixed(2)), controls: controls.length };
  });
}

let context;
try {
  let files = await artifact(path.join(root, "dist/web"));
  checkPaths(files);
  const originalFiles = files;
  const requests = [], failures = [], errors = [], responseTypes = new Map();
  let assetFault, workerFault, versionFault, blockSolverAssets = false;
  context = await chromium.launchPersistentContext(path.join(temporary, "profile"), { executablePath: await executable(), headless: true, viewport: { width: 390, height: 844 }, args: ["--disable-gpu"], serviceWorkers: "block" });
  await context.addInitScript(() => {
    Object.defineProperty(navigator, "clipboard", { value: { writeText: async text => { globalThis.copiedText = text; } } });
    const instantiate = WebAssembly.instantiate;
    WebAssembly.instantiate = async (...args) => {
      const result = await instantiate(...args);
      const instance = result.instance || result;
      globalThis.pagesPrivateMemory = Object.values(instance.exports).filter(value => value instanceof WebAssembly.Memory).map(memory => memory.buffer instanceof ArrayBuffer);
      return result;
    };
    new MutationObserver((_, observer) => {
      const ready = document.getElementById("reset");
      if (ready && !ready.disabled && document.querySelectorAll(".sticker").length === 54) {
        requestAnimationFrame(() => requestAnimationFrame(() => { globalThis.pagesInteractiveMs ??= performance.now(); }));
        observer.disconnect();
      }
    }).observe(document, { subtree: true, childList: true, attributes: true, attributeFilter: ["disabled"] });
  });
  // Every URL is intercepted. There is no network server, API, route fallback,
  // custom header, COOP, COEP, or service worker in this Pages simulation.
  await context.route("**/*", async route => {
    const request = route.request(), url = new URL(request.url());
    const asset = url.pathname === "/cube/" ? "index.html" : url.pathname.slice("/cube/".length);
    const permitted = url.origin === new URL(site).origin && url.pathname.startsWith("/cube/") && !asset.includes("/") && files.has(asset) && asset !== ".nojekyll";
    requests.push({ url: url.href, asset, method: request.method() });
    if (!permitted || request.method() !== "GET") {
      failures.push(url.href);
      return route.fulfill({ status: 404, contentType: "text/plain", body: "Not found" });
    }
    const contentType = mime[path.extname(asset)];
    responseTypes.set(asset, contentType);
    if (blockSolverAssets && asset.endsWith(".gz")) return route.abort("internetdisconnected");
    if (workerFault && /^worker\./.test(asset)) return route.abort("internetdisconnected");
    if (versionFault && asset === "version.json") return route.abort("internetdisconnected");
    if (assetFault && /^coordinates-web-v1\.bin\./.test(asset)) {
      if (assetFault === "missing") return route.fulfill({ status: 503, contentType: "text/plain", body: "Unavailable" });
      if (assetFault === "offline") return route.abort("internetdisconnected");
      const corrupt = Buffer.from(files.get(asset));
      corrupt[corrupt.length - 1] ^= 1;
      return route.fulfill({ status: 200, contentType, body: corrupt });
    }
    await route.fulfill({ status: 200, contentType, body: files.get(asset) });
  });
  const page = context.pages()[0];
  page.on("pageerror", error => errors.push(error.message));
  page.on("console", message => { if (message.type() === "error") errors.push(message.text()); });
  const cdp = await context.newCDPSession(page);
  await cdp.send("Emulation.setCPUThrottlingRate", { rate: 4 });
  await cdp.send("Network.setCacheDisabled", { cacheDisabled: true });
  await page.goto(site);
  await idle(page);
  await page.waitForFunction(() => Number.isFinite(globalThis.pagesInteractiveMs));
  const load = await page.evaluate(() => ({ interactiveMs: Math.round(globalThis.pagesInteractiveMs), isolated: crossOriginIsolated, sharedArrayBuffer: typeof SharedArrayBuffer, privateWasmMemory: globalThis.pagesPrivateMemory, width: innerWidth, height: innerHeight }));
  assert.equal(load.isolated, false);
  assert.equal(load.sharedArrayBuffer, "undefined");
  assert.deepEqual(load.privateWasmMemory, [true], "The engine exports one ordinary ArrayBuffer memory");
  assert.equal(load.width, 390);
  assert.equal(load.height, 844);
  assert.ok(load.interactiveMs < 15000, `First load under 4x CPU throttle: ${load.interactiveMs}ms < 15s`);
  console.log(`First load: ${load.interactiveMs}ms to interactive, 390x844, 4x CPU throttle; local in-memory transfer (no network throttling)`);
  await cdp.send("Emulation.setCPUThrottlingRate", { rate: 1 });
  const firstLoadAssets = [...new Set(requests.map(request => request.asset))];
  assert.equal(firstLoadAssets.some(name => name.endsWith(".gz")), false, "First load fetches no solver table assets");
  const firstLoadRaw = firstLoadAssets.reduce((sum, name) => sum + files.get(name).length, 0);
  assert.deepEqual(errors, [], "First-load console and runtime errors");
  const headers = await page.evaluate(async () => Object.fromEntries((await fetch("./index.html")).headers));
  for (const header of ["cross-origin-opener-policy", "cross-origin-embedder-policy", "cross-origin-resource-policy"]) assert.equal(headers[header], undefined);
  console.log("PASS Pages: exact exported files, project-subpath paths, static MIME types, no shared memory or isolation headers");

  // Tab through the actual DOM order, including the cube's keyboard surface.
  await page.evaluate(() => document.activeElement.blur());
  const focusOrder = [];
  for (const expected of [".brand", "#help-toggle", "#share", "#size", "#view-3d", "#view-net", "#stage", "#home-view", "#scramble", "#solve-method", "#solve", "#reset", "#tab-practice"]) {
    await page.keyboard.press("Tab");
    assert.ok(await page.locator(expected).evaluate(node => node === document.activeElement), `Keyboard focus follows DOM order: ${expected}`);
    focusOrder.push(expected);
    const result = await accessibility(page);
    assert.deepEqual(result.failures, [], "Basic accessibility: names, control text contrast and visible focus");
  }
  for (const [key, mode] of [["ArrowRight", "lesson"], ["End", "search"], ["Home", "practice"]]) {
    await page.keyboard.press(key);
    assert.equal(await page.locator(`#tab-${mode}`).getAttribute("aria-selected"), "true");
    assert.ok(await page.locator(`#tab-${mode}`).evaluate(node => node === document.activeElement));
    assert.equal(await page.locator(`#${mode}`).isVisible(), true);
    assert.deepEqual((await accessibility(page)).failures, [], `${mode} controls accessible`);
  }
  await page.locator("#stage").focus();
  const solved = await page.locator("#cfen").inputValue();
  await page.keyboard.press("r");
  await idle(page);
  const turned = await page.locator("#cfen").inputValue();
  assert.notEqual(turned, solved, "Keyboard turns the cube");
  await page.keyboard.press("Shift+r");
  await idle(page);
  assert.equal(await page.locator("#cfen").inputValue(), solved, "Prime keyboard turn restores the cube");
  await page.locator("#help-toggle").focus();
  await page.keyboard.press("Enter");
  assert.equal(await page.locator("#keyboard-help").evaluate(node => node.matches(":popover-open")), true);
  assert.deepEqual((await accessibility(page)).failures, [], "Keyboard help controls accessible");
  await page.keyboard.press("Escape");
  assert.equal(await page.locator("#keyboard-help").evaluate(node => node.matches(":popover-open")), false);
  console.log(`PASS accessibility: labels, ${focusOrder.length} tab stops, visible focus, control contrast, mode tabs, help and face/prime keys`);

  // A copied hash deep link must restore from the real static index.html.
  await page.goto(`${site}index.html#${new URLSearchParams({ scramble: "R U", alg: "U' R'", index: "1" })}`);
  await idle(page);
  assert.equal(await page.locator("#scrubber").inputValue(), "1");
  assert.deepEqual(await page.locator("#sequence-moves button").allTextContents(), ["U'", "R'"]);
  const checkpoint = await page.locator("#cfen").inputValue();
  await page.locator("#share").click();
  const shared = await page.evaluate(() => globalThis.copiedText);
  assert.equal(new URL(shared).pathname, "/cube/index.html");
  assert.ok(new URL(shared).hash.length > 1);
  await page.goto(shared);
  await idle(page);
  assert.equal(await page.locator("#cfen").inputValue(), checkpoint);
  assert.equal(await page.locator("#scrubber").inputValue(), "1");
  const rootLink = new URL(shared);
  rootLink.pathname = "/cube/";
  await page.goto(rootLink.href);
  await idle(page);
  assert.equal(await page.locator("#cfen").inputValue(), checkpoint, "The directory index restores the same deep link");
  assert.equal(await page.locator("#scrubber").inputValue(), "1");
  assert.deepEqual((await accessibility(page)).failures, [], "Playback controls accessible");
  await page.locator("#step").focus();
  await page.keyboard.press("Enter");
  await idle(page);
  assert.equal(await page.locator("#cfen").inputValue(), solved, "Keyboard completes the restored sequence");
  console.log("PASS Pages: direct index.html deep link, copied hash restoration and keyboard playback");

  // Exercise both module-worker entry points without COOP/COEP or SAB.
  await page.locator("#stage").focus();
  await page.keyboard.press("r");
  await idle(page);
  const firstSolveStarted = Date.now();
  await page.locator("#solve").click();
  await idle(page);
  const first3x3Ms = Date.now() - firstSolveStarted;
  assert.equal(await page.locator("#sequence-title").textContent(), "Your path to solved · kociemba");
  await page.locator("#tab-search").click();
  await page.locator("#depth").fill("1");
  await page.locator("#find").click();
  await idle(page);
  assert.match(await page.locator("#search-result").textContent(), /Found 1 moves/);
  assert.deepEqual((await accessibility(page)).failures, [], "Dynamic search controls accessible");
  const workerRequests = requests.filter(request => /^worker(?:\.|$)/.test(request.asset));
  assert.ok(workerRequests.length >= 2, "Solve and search load the exported module worker");
  for (const [asset, contentType] of responseTypes) {
    if (asset.endsWith(".wasm")) assert.equal(contentType, "application/wasm");
    if (asset.endsWith(".js")) assert.equal(contentType, "text/javascript");
  }
  assert.deepEqual(failures, [], "No missing assets, external requests or API requests");
  assert.deepEqual(errors, [], "No console or runtime errors throughout Pages use");
  console.log("PASS Pages: cold solve/search module workers without cross-origin isolation; no console errors");

  await page.locator("#size").selectOption("7");
  await idle(page, 7);
  const sevenStart = await page.evaluate(() => JSON.parse(globalThis.cubeAPI(JSON.stringify({ op: "twist", size: 7, moves: "2R U 2F' Rw D2 L B'" }))).data.state.cfen);
  await page.locator("#cfen").fill(sevenStart);
  await page.locator("#import").click();
  const sevenStarted = Date.now();
  await page.locator("#solve").click();
  await idle(page, 7);
  const first7x7 = { readyMs: Date.now() - sevenStarted, workerMs: Number(await page.locator("#cube").getAttribute("data-solve-wall-ms")), solveMs: Number(await page.locator("#cube").getAttribute("data-solve-ms")), moves: Number(await page.locator("#scrubber").getAttribute("max")) };
  const sevenMoves = await page.locator("#sequence-moves button").allTextContents();
  assert.equal(await page.evaluate(({ cfen, moves }) => JSON.parse(globalThis.cubeAPI(JSON.stringify({ op: "twist", cfen, moves }))).data.state.solved, { cfen: sevenStart, moves: sevenMoves.join(" ") }), true, "First 7x7 solution replays to solved");
  assert.ok(requests.some(request => /^coordinates-web-v1\.bin\./.test(request.asset)), "First solver use fetches coordinates lazily");
  assert.ok(requests.some(request => /^nxn-7-v1\.bin\./.test(request.asset)), "First 7x7 solve fetches its own table");
  assert.equal(requests.some(request => /^nxn-[456]-v1\.bin\./.test(request.asset)), false, "Unselected NxN dimensions remain unfetched");
  console.log(`First solver use: 3x3 ${first3x3Ms}ms; 7x7 ${JSON.stringify(first7x7)} (including asset loading and playback preparation)`);

  // Every case starts a fresh page/runtime, with no table bytes in memory.
  const solvedPage = await context.newPage();
  const solvedErrors = [];
  solvedPage.on("pageerror", error => solvedErrors.push(error.message));
  for (const cubeSize of [2, 3, 4, 5, 6, 7]) for (const grip of ["", "x"]) {
    const params = new URLSearchParams({ size: String(cubeSize), scramble: grip });
    const before = requests.length;
    await solvedPage.goto(`${site}#${params}`);
    await idle(solvedPage, cubeSize);
    const cfen = await solvedPage.locator("#cfen").inputValue();
    assert.equal(await solvedPage.locator("#cube").getAttribute("data-solved"), "true");
    await solvedPage.locator("#solve").click();
    await idle(solvedPage, cubeSize);
    assert.equal(await solvedPage.locator("#notice").evaluate(node => node.classList.contains("error")), false, `Cold solved ${cubeSize}x${cubeSize} ${grip}`);
    assert.equal(await solvedPage.locator("#sequence-title").textContent(), `Your path to solved · ${cubeSize === 3 ? "kociemba" : "reduction"}`);
    const moves = await solvedPage.locator("#sequence-moves button").allTextContents();
    assert.ok(moves.every(move => /^[xyz](?:2|')?$/.test(move)), "Solved grips return only rotations");
    assert.equal(await solvedPage.locator("#cfen").inputValue(), cfen, "Solved input remains unchanged");
    assert.equal(await solvedPage.evaluate(({ cfen, moves }) => JSON.parse(globalThis.cubeAPI(JSON.stringify({ op: "twist", cfen, moves }))).data.state.solved, { cfen, moves: moves.join(" ") }), true);
    assert.equal(requests.slice(before).some(request => request.asset.endsWith(".gz")), false, "Solved grips need no solver assets");
  }
  assert.deepEqual(solvedErrors, [], "Cold solved grips have no uncaught errors");
  await solvedPage.close();
  console.log("PASS Pages: cold solved and rotated 2x2–7x7 return only rotations without table downloads");

  const cachePage = await context.newPage();
  const cacheErrors = [];
  cachePage.on("pageerror", error => cacheErrors.push(error.message));
  await cachePage.addInitScript(() => {
    const NativeWorker = globalThis.Worker;
    globalThis.pagesWorkerStarts = globalThis.pagesWorkerStops = 0;
    globalThis.Worker = class extends NativeWorker {
      constructor(...args) {
        super(...args);
        globalThis.pagesWorkerStarts++;
        this.addEventListener("message", ({ data }) => {
          if (data.type === "progress") globalThis.pagesWorkerPhase = data.phase;
        });
      }
      terminate() { globalThis.pagesWorkerStops++; return super.terminate(); }
    };
  });
  await cachePage.goto(site);
  await idle(cachePage);
  const assertCachedSolve = async cubeSize => {
    const cfen = await cachePage.locator("#cfen").inputValue();
    const before = requests.length;
    await cachePage.locator("#solve").click();
    await idle(cachePage, cubeSize);
    assert.equal(await cachePage.locator("#notice").evaluate(node => node.classList.contains("error")), false, `Cached ${cubeSize}x${cubeSize} solve succeeds with downloads blocked`);
    const moves = await cachePage.locator("#sequence-moves button").allTextContents();
    assert.ok(moves.length > 0);
    assert.equal(await cachePage.evaluate(({ cfen, moves }) => JSON.parse(globalThis.cubeAPI(JSON.stringify({ op: "twist", cfen, moves }))).data.state.solved, { cfen, moves: moves.join(" ") }), true);
    assert.equal(await cachePage.locator("#cfen").inputValue(), cfen, "Offline solve preserves the input");
    assert.equal(requests.slice(before).some(request => request.asset.endsWith(".gz")), false, "Verified table bytes are reused without a fetch");
  };
  for (const cubeSize of [3, 2, 4, 5, 6, 7]) {
    await cachePage.locator("#size").selectOption(String(cubeSize));
    await idle(cachePage, cubeSize);
    const cfen = await cachePage.evaluate(size => JSON.parse(globalThis.cubeAPI(JSON.stringify({ op: "twist", size, moves: "R U" }))).data.state.cfen, cubeSize);
    await cachePage.locator("#cfen").fill(cfen);
    await cachePage.locator("#import").click();
    await cachePage.locator("#solve").click();
    await idle(cachePage, cubeSize);
    assert.equal(await cachePage.locator("#notice").evaluate(node => node.classList.contains("error")), false, "Online solve installs verified bytes");
    const starts = await cachePage.evaluate(() => pagesWorkerStarts);
    blockSolverAssets = true;
    await assertCachedSolve(cubeSize);
    if (cubeSize === 3) assert.equal(await cachePage.evaluate(() => pagesWorkerStarts), starts + 1, "3x3 reuse survives runtime disposal");
    blockSolverAssets = false;
  }
  blockSolverAssets = true;
  for (const cubeSize of [4, 7, 3]) {
    await cachePage.locator("#size").selectOption(String(cubeSize));
    await idle(cachePage, cubeSize);
    const cfen = await cachePage.evaluate(size => JSON.parse(globalThis.cubeAPI(JSON.stringify({ op: "twist", size, moves: "R U" }))).data.state.cfen, cubeSize);
    await cachePage.locator("#cfen").fill(cfen);
    await cachePage.locator("#import").click();
    await assertCachedSolve(cubeSize);
  }

  // Terminate an actual 3x3 search, then create a new runtime using cached bytes.
  const hard = await cachePage.evaluate(() => JSON.parse(globalThis.cubeAPI(JSON.stringify({ op: "twist", moves: "R2 U F' D B2 L' U2 F R' D2 L B' U R2 F2 D' L2 U' B R" }))).data.state.cfen);
  await cachePage.locator("#cfen").fill(hard);
  await cachePage.locator("#import").click();
  await cachePage.locator("#tab-search").click();
  await cachePage.locator("#target").fill(solved);
  await cachePage.locator("#depth").fill("10");
  const stops = await cachePage.evaluate(() => { pagesWorkerPhase = ""; return pagesWorkerStops; });
  const beforeCancel = requests.length;
  await cachePage.locator("#find").click();
  await cachePage.waitForFunction(() => pagesWorkerPhase === "solving");
  await cachePage.locator("#view-net").click();
  assert.equal(await cachePage.locator("#net").isVisible(), true, "Search leaves view controls responsive");
  await cachePage.locator("#cancel-search").click();
  await idle(cachePage);
  assert.match(await cachePage.locator("#notice").textContent(), /Search canceled/);
  assert.equal(await cachePage.evaluate(() => pagesWorkerStops), stops + 1, "Cancellation terminates synchronous search");
  assert.equal(await cachePage.locator("#cfen").inputValue(), hard);
  assert.equal(await cachePage.locator("#search-result").textContent(), "");
  assert.equal(requests.slice(beforeCancel).some(request => request.asset.endsWith(".gz")), false);
  await cachePage.locator("#reset").click();
  const retryCFEN = await cachePage.evaluate(() => JSON.parse(globalThis.cubeAPI(JSON.stringify({ op: "twist", moves: "R U" }))).data.state.cfen);
  await cachePage.locator("#cfen").fill(retryCFEN);
  await cachePage.locator("#import").click();
  await assertCachedSolve(3);
  blockSolverAssets = false;
  assert.deepEqual(cacheErrors, [], "Asset cache and cancellation have no uncaught errors");
  await cachePage.close();
  console.log("PASS Pages: offline repeat solves for 2x2–7x7, size-switch reuse, 3x3 search termination and offline retry");

  const recoveryPage = await context.newPage();
  const recoveryErrors = [];
  recoveryPage.on("pageerror", error => recoveryErrors.push(error.message));
  for (const fault of ["missing", "corrupt", "offline", "worker-offline"]) {
    await recoveryPage.goto(site);
    await idle(recoveryPage);
    await recoveryPage.locator("#stage").focus();
    await recoveryPage.keyboard.press("r");
    await idle(recoveryPage);
    const mixed = await recoveryPage.locator("#cfen").inputValue();
    assetFault = fault === "worker-offline" ? null : fault;
    workerFault = fault === "worker-offline";
    versionFault = fault === "offline";
    await recoveryPage.locator("#solve").click();
    await recoveryPage.locator("#notice.error").waitFor();
    await idle(recoveryPage);
    assert.match(await recoveryPage.locator("#notice").textContent(), fault === "worker-offline" ? /worker could not start/ : /Solver data.*(?:missing|verified)/);
    // Wait for the failure-only version check before asserting no update prompt.
    await recoveryPage.waitForFunction(() => performance.getEntriesByType("resource").some(entry => entry.name.endsWith("/version.json")));
    assert.equal(await recoveryPage.locator("#update-notice").isVisible(), false, `${fault} on the current release remains a loading/network error`);
    assert.equal(await recoveryPage.locator("#cfen").inputValue(), mixed, `${fault} solver data preserves the input`);
    assert.equal(await recoveryPage.locator("#cancel-task").isVisible(), false, "Failed loading releases the worker and controls");
    assetFault = null;
    workerFault = false;
    versionFault = false;
    await recoveryPage.locator("#solve").click();
    await idle(recoveryPage);
    assert.equal(await recoveryPage.locator("#sequence-title").textContent(), "Your path to solved · kociemba", `${fault} data recovers on retry`);
  }
  assert.deepEqual(recoveryErrors, [], "Asset failure recovery has no uncaught exceptions");
  await recoveryPage.close();
  console.log("PASS Pages: missing/corrupt/offline lazy data and offline worker preserve the cube, keep their own errors and recover on retry");

  // Change WASM alone, then JS alone: the entire graph must get new URLs.
  // Seed the old URL namespace with poisoned binaries to simulate a stale cache.
  assert.ok(files.has(".nojekyll"), "Jekyll processing is disabled in the export");
  const stableFiles = ["index.html", ".nojekyll", "version.json"];
  for (const name of files.keys()) if (!stableFiles.includes(name)) assert.match(name, /\.[a-f0-9]{16}\./, `Fingerprint ${name}`);
  const source = await readRuntimeAssets(path.join(root, "web"));
  const baseline = path.join(temporary, "baseline");
  await exportPages(source, baseline);
  assert.deepEqual(await artifact(baseline), files, "Readiness serves exactly make web-pages output");
  const oldFiles = files;
  const changedWasm = new Map(source);
  changedWasm.set("cube.wasm", Buffer.concat([source.get("cube.wasm"), Buffer.from([0, 1, 0])])); // valid empty custom section
  const wasmRelease = path.join(temporary, "wasm-release");
  await exportPages(changedWasm, wasmRelease);
  files = await artifact(wasmRelease);
  checkPaths(files);
  for (const name of oldFiles.keys()) if (!stableFiles.includes(name)) {
    assert.equal(files.has(name), false, `WASM change invalidates ${name}`);
    files.set(name, Buffer.from("stale cached asset must never be requested"));
  }
  const beforeRedeploy = requests.length;
  await page.goto(site);
  await idle(page);
  await page.locator("#stage").focus();
  await page.keyboard.press("r");
  await idle(page);
  await page.locator("#solve").click();
  await idle(page);
  for (const request of requests.slice(beforeRedeploy)) if (!stableFiles.includes(request.asset)) assert.equal(oldFiles.has(request.asset), false, `Redeploy avoids cached ${request.asset}`);
  const changedJS = new Map(changedWasm);
  changedJS.set("app.js", Buffer.concat([source.get("app.js"), Buffer.from("\n// next release\n")]));
  const jsRelease = path.join(temporary, "js-release");
  await exportPages(changedJS, jsRelease);
  const nextFiles = await artifact(jsRelease);
  for (const name of nextFiles.keys()) if (!stableFiles.includes(name)) assert.equal(files.has(name), false, `JS change invalidates ${name}`);
  files = nextFiles;
  await page.goto(site);
  await idle(page);
  assert.deepEqual(errors, [], "Redeploy never pairs stale WASM/runtime with fresh JS");
  assert.deepEqual(failures, [], "Redeploy references only available runtime files");
  console.log("PASS Pages: WASM-only and JS-only redeploys invalidate the whole asset graph; poisoned old assets are unused");

  // Replace the same served root while A is open. Exercise both an unloaded
  // worker and an already running worker whose 7x7 tables are still lazy.
  const staleErrors = [];
  for (const cubeSize of [3, 7]) {
    files = oldFiles;
    const stalePage = await context.newPage();
    stalePage.on("pageerror", error => staleErrors.push(error.message));
    await stalePage.goto(`${site}#${new URLSearchParams({ size: String(cubeSize) })}`);
    await idle(stalePage, cubeSize);
    if (cubeSize === 7) {
      const beforeWarmup = requests.length;
      const workerAsset = [...oldFiles.keys()].find(name => /^worker\./.test(name));
      await stalePage.evaluate(async asset => {
        const NativeWorker = Worker, url = new URL(asset, location.href);
        const worker = new NativeWorker(url, { type: "module" });
        // Load A's real engine using its ordinary state operation, which needs
        // no solver data. Give this live worker to the app's first computation.
        await new Promise((resolve, reject) => {
          const timer = setTimeout(() => { worker.terminate(); reject(new Error("Worker warmup timed out")); }, 30000);
          worker.onerror = event => { clearTimeout(timer); reject(new Error(event.message || "Worker warmup failed")); };
          worker.onmessage = ({ data }) => {
            if (data.ok === undefined) return;
            clearTimeout(timer);
            if (!data.ok) reject(new Error(data.error));
            else resolve();
          };
          worker.postMessage({ request: { op: "state", size: 7 } });
        });
        worker.onmessage = worker.onerror = null;
        globalThis.Worker = function (requestedURL, options) {
          globalThis.Worker = NativeWorker;
          if (String(requestedURL) !== url.href || options.type !== "module") throw new Error("Unexpected warmed worker URL");
          return worker;
        };
      }, workerAsset);
      assert.ok(requests.slice(beforeWarmup).some(request => /^worker\./.test(request.asset)), "Warm a retained 7x7 worker on A");
      assert.equal(requests.slice(beforeWarmup).some(request => request.asset.endsWith(".gz")), false, "State-only worker warmup leaves all tables lazy");
    }
    const params = new URLSearchParams({ size: String(cubeSize), scramble: cubeSize === 7 ? "2R U 2F'" : "R U", alg: "F R' U2", index: "1", draft: "R U unfinished" });
    await stalePage.evaluate(hash => { location.hash = hash; }, `#${params}`);
    await stalePage.waitForFunction(() => document.getElementById("scrubber").value === "1" && document.getElementById("algorithm").value === "R U unfinished");
    await idle(stalePage, cubeSize);
    const mixed = await stalePage.locator("#cfen").inputValue();
    const savedHash = await stalePage.evaluate(() => location.hash);
    const moves = await stalePage.locator("#sequence-moves button").allTextContents();
    files = nextFiles;
    const failedBefore = failures.length, requestsBefore = requests.length;
    await stalePage.locator("#solve").click();
    await stalePage.locator("#update-notice").waitFor();
    await idle(stalePage, cubeSize);
    assert.equal(await stalePage.locator("#update-notice span").textContent(), "A new version of the cube is available, reload to continue.");
    assert.equal(await stalePage.locator("#cfen").inputValue(), mixed, "Stale solve preserves every sticker");
    assert.equal(await stalePage.evaluate(() => location.hash), savedHash, "Failed solve keeps the entire saved sequence and playhead");
    assert.deepEqual(await stalePage.locator("#sequence-moves button").allTextContents(), moves);
    assert.equal(await stalePage.locator("#scrubber").inputValue(), "1");
    assert.equal(await stalePage.locator("#cancel-task").isVisible(), false);
    assert.equal(await stalePage.locator("#reload").isEnabled(), true);
    const failedAssets = [...oldFiles.keys()].filter(name => cubeSize === 3 ? /^worker\./.test(name) : /^(coordinates-web|nxn-7)-v1\.bin\./.test(name));
    assert.deepEqual(failures.slice(failedBefore).sort(), failedAssets.map(name => new URL(name, site).href).sort(), "Only A's deleted worker or lazy tables fail");
    assert.ok(requests.slice(requestsBefore).some(request => request.asset === "version.json"), "Load failure checks the current deployment version");
    if (cubeSize === 7) assert.equal(requests.slice(requestsBefore).some(request => /^worker\./.test(request.asset)), false, "The lazy-asset repro uses the existing worker");
    assert.deepEqual((await accessibility(stalePage)).failures, [], "The update prompt and Reload button are accessible");
    const prompt = await stalePage.locator("#update-notice").boundingBox();
    assert.ok(prompt && prompt.y >= 0 && prompt.y + prompt.height <= stalePage.viewportSize().height, "The update prompt stays visible at the scrolled solve controls");
    if (cubeSize === 7) await stalePage.screenshot({ path: path.join(scratch, "pages-update-390x844.png") });

    // The prompt stays available while normal playback and editing continue.
    await stalePage.locator("#step").click();
    await idle(stalePage, cubeSize);
    await stalePage.locator("#algorithm").fill("U R edited after the update");
    assert.equal(await stalePage.locator("#update-notice").isVisible(), true);
    const latest = await stalePage.locator("#cfen").inputValue();
    const latestHash = await stalePage.evaluate(() => location.hash);
    await Promise.all([stalePage.waitForEvent("load"), stalePage.locator("#reload").click()]);
    await stalePage.waitForFunction(() => document.getElementById("scrubber").value === "2");
    await idle(stalePage, cubeSize);
    assert.equal(await stalePage.locator("#cfen").inputValue(), latest, "Reload restores the latest cube, including edits after the prompt");
    assert.equal(await stalePage.evaluate(() => location.hash), latestHash);
    assert.equal(await stalePage.locator("#size").inputValue(), String(cubeSize));
    assert.equal(await stalePage.locator("#algorithm").inputValue(), "U R edited after the update");
    assert.deepEqual(await stalePage.locator("#sequence-moves button").allTextContents(), moves);
    assert.equal(await stalePage.locator("#update-notice").isVisible(), false);
    await stalePage.locator("#solve").click();
    await stalePage.waitForFunction(() => document.getElementById("notice").textContent.startsWith("Solution ready"), null, { timeout: 120000 });
    await idle(stalePage, cubeSize);
    const solution = await stalePage.locator("#sequence-moves button").allTextContents();
    assert.equal(await stalePage.evaluate(({ cfen, moves }) => JSON.parse(globalThis.cubeAPI(JSON.stringify({ op: "twist", cfen, moves }))).data.state.solved, { cfen: latest, moves: solution.join(" ") }), true, "B solves the restored cube to completion");
    assert.deepEqual(failures.slice(failedBefore).sort(), failedAssets.map(name => new URL(name, site).href).sort(), "Recovery requests only B's available assets");
    await stalePage.close();
  }
  assert.deepEqual(staleErrors, [], "Stale-tab recovery has no uncaught exceptions");
  console.log("PASS Pages: old worker and lazy 7x7 asset failures detect B, offer a nonblocking reload, preserve cube/sequence/playhead/size/draft and solve after recovery");
  const report = { browser: context.browser().version(), viewport: "390x844", cpuThrottle: 4, network: "in-memory, unthrottled", ...load, firstLoadRaw, first3x3Ms, first7x7, accessibility: await accessibility(page) };
  await page.screenshot({ path: path.join(scratch, "pages-390x844.png"), fullPage: true });
  await context.close();
  context = null;
  const wasmBuild = JSON.parse(await readFile(path.join(scratch, "wasm-build.json"), "utf8"));
  report.sizes = sizes(originalFiles, wasmBuild.optimized);
  assert.equal(report.sizes.wasm.raw, wasmBuild.raw, "The exported wasm matches the measured build");
  report.wasmBuild = wasmBuild;
  console.log(`Bytes: ${JSON.stringify({ ...report.sizes, firstLoadRaw })}; gzip level 9 / Brotli quality 11 estimates, per file`);
  await writeFile(path.join(scratch, "pages-readiness.json"), JSON.stringify(report, null, 2) + "\n");
  console.log("PASS test-pages; report .scratch/pages-readiness.json, screenshot .scratch/pages-390x844.png");
} finally {
  if (context) await context.close();
  await rm(temporary, { recursive: true, force: true });
}
