import { createHash } from "node:crypto";
import { mkdir, readFile, rm, writeFile } from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { solverAssetNames } from "./build-solver-assets.mjs";

const root = fileURLToPath(new URL("../", import.meta.url));
export const runtimeAssets = ["index.html", "style.css", "icon.svg", "app.js", "cube-view.js", "engine.js", "worker.js", "cube.wasm", "wasm_exec.js", "solver-assets.js", ...solverAssetNames];

export async function readRuntimeAssets(directory) {
  const assets = new Map();
  for (const name of runtimeAssets) assets.set(name, await readFile(path.join(directory, name)));
  return assets;
}

// One fingerprint covers the whole module graph, including Go's matching runtime.
// An unchanged worker cannot accidentally import a newer engine or WASM binary.
export async function exportPages(assets, destination) {
  const hash = createHash("sha256");
  for (const name of runtimeAssets) hash.update(name).update("\0").update(assets.get(name)).update("\0");
  const version = hash.digest("hex").slice(0, 16);
  const names = new Map(runtimeAssets.map(name => [name, name === "index.html" ? name : `${path.parse(name).name}.${version}${path.extname(name)}`]));
  await rm(destination, { recursive: true, force: true });
  await mkdir(destination, { recursive: true });
  for (const [name, bytes] of assets) {
    let body = bytes;
    if (/\.(html|css|js|svg)$/.test(name)) {
      body = bytes.toString("utf8").replace(/\.\/([\w.-]+\.(?:html|css|js|svg|wasm|gz))\b/g, (reference, asset) => {
        if (!names.has(asset)) throw new Error(`Unknown runtime asset ${asset} in ${name}`);
        return `./${names.get(asset)}`;
      });
    }
    await writeFile(path.join(destination, names.get(name)), body);
  }
  await writeFile(path.join(destination, ".nojekyll"), "");
  return version;
}

if (process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  const version = await exportPages(await readRuntimeAssets(path.join(root, "web")), path.join(root, "dist/web"));
  console.log(`Pages artifact: dist/web (release ${version})`);
}
