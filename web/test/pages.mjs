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
const mime = { ".html": "text/html", ".css": "text/css", ".js": "text/javascript", ".wasm": "application/wasm", ".svg": "image/svg+xml" };

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
}

function sizes(files) {
  const result = { total: { raw: 0, gzip: 0, brotli: 0 } };
  for (const [name, body] of files) {
    if (name === ".nojekyll") continue;
    const size = { raw: body.length, gzip: gzipSync(body, { level: 9 }).length,
      brotli: brotliCompressSync(body, { params: { [constants.BROTLI_PARAM_QUALITY]: 11 } }).length };
    for (const kind of Object.keys(size)) result.total[kind] += size[kind];
    if (name.endsWith(".wasm")) result.wasm = size;
  }
  assert.ok(result.total.raw < 20_000_000, "Uncompressed Pages artifact stays under 20 MB");
  assert.ok(result.wasm.raw < 15_000_000, "WASM stays under 15 MB");
  return result;
}

async function idle(page) {
  await page.waitForFunction(() => !document.getElementById("reset")?.disabled && document.querySelectorAll(".sticker").length === 54, null, { timeout: 60000 });
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
  await page.locator("#solve").click();
  await idle(page);
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

  // Change WASM alone, then JS alone: the entire graph must get new URLs.
  // Seed the old URL namespace with poisoned binaries to simulate a stale cache.
  assert.ok(files.has(".nojekyll"), "Jekyll processing is disabled in the export");
  for (const name of files.keys()) if (!["index.html", ".nojekyll"].includes(name)) assert.match(name, /\.[a-f0-9]{16}\./, `Fingerprint ${name}`);
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
  for (const name of oldFiles.keys()) if (!["index.html", ".nojekyll"].includes(name)) {
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
  for (const request of requests.slice(beforeRedeploy)) if (request.asset !== "index.html") assert.equal(oldFiles.has(request.asset), false, `Redeploy avoids cached ${request.asset}`);
  const changedJS = new Map(changedWasm);
  changedJS.set("app.js", Buffer.concat([source.get("app.js"), Buffer.from("\n// next release\n")]));
  const jsRelease = path.join(temporary, "js-release");
  await exportPages(changedJS, jsRelease);
  const nextFiles = await artifact(jsRelease);
  for (const name of nextFiles.keys()) if (!["index.html", ".nojekyll"].includes(name)) assert.equal(files.has(name), false, `JS change invalidates ${name}`);
  files = nextFiles;
  await page.goto(site);
  await idle(page);
  assert.deepEqual(errors, [], "Redeploy never pairs stale WASM/runtime with fresh JS");
  assert.deepEqual(failures, [], "Redeploy references only available runtime files");
  console.log("PASS Pages: WASM-only and JS-only redeploys invalidate the whole asset graph; poisoned old assets are unused");
  const report = { browser: context.browser().version(), viewport: "390x844", cpuThrottle: 4, network: "in-memory, unthrottled", ...load, firstLoadRaw, accessibility: await accessibility(page) };
  await page.screenshot({ path: path.join(scratch, "pages-390x844.png"), fullPage: true });
  await context.close();
  context = null;
  report.sizes = sizes(originalFiles);
  console.log(`Bytes: ${JSON.stringify({ ...report.sizes, firstLoadRaw })}; gzip level 9 / Brotli quality 11 estimates, per file`);
  await writeFile(path.join(scratch, "pages-readiness.json"), JSON.stringify(report, null, 2) + "\n");
  console.log("PASS test-pages; report .scratch/pages-readiness.json, screenshot .scratch/pages-390x844.png");
} finally {
  if (context) await context.close();
  await rm(temporary, { recursive: true, force: true });
}
