import { access, readFile, realpath, rename, rm, stat, writeFile } from "node:fs/promises";
import { spawn } from "node:child_process";
import path from "node:path";
import { fileURLToPath } from "node:url";

const root = fileURLToPath(new URL("../", import.meta.url));
const binary = path.join(root, "web/cube.wasm");
async function available(file) {
  try { await access(file); return true; } catch { return false; }
}
async function optimizer() {
  if (process.env.WASM_OPT === "off") return null;
  if (process.env.WASM_OPT) return process.env.WASM_OPT;
  for (const directory of (process.env.PATH || "").split(path.delimiter)) {
    const candidate = path.join(directory, "wasm-opt");
    if (await available(candidate)) return candidate;
    // Homebrew's Emscripten includes Binaryen, although wasm-opt is not on PATH.
    const emcc = path.join(directory, "emcc");
    if (await available(emcc)) {
      const resolved = await realpath(emcc);
      for (const relative of ["../libexec/binaryen/bin/wasm-opt", "binaryen/bin/wasm-opt"]) {
        const bundled = path.resolve(path.dirname(resolved), relative);
        if (await available(bundled)) return bundled;
      }
    }
  }
  return null;
}

const raw = (await stat(binary)).size;
const tool = await optimizer();
const report = { optimized: false, raw, unoptimizedRaw: raw };
if (tool) {
  const output = path.join(root, ".scratch/optimized-cube.wasm");
  // These features already occur in Go's output. Enable no new wasm proposals.
  const args = ["-Oz", "--enable-bulk-memory", "--enable-nontrapping-float-to-int", "--enable-sign-ext", binary, "-o", output];
  await new Promise((resolve, reject) => {
    const child = spawn(tool, args, { stdio: "inherit", env: { ...process.env, BINARYEN_CORES: "2" } });
    child.on("error", reject);
    child.on("exit", code => code === 0 ? resolve() : reject(new Error(`wasm-opt exited ${code}`)));
  });
  if (!WebAssembly.validate(await readFile(output))) throw new Error("wasm-opt produced invalid wasm");
  const optimizedRaw = (await stat(output)).size;
  if (optimizedRaw < raw) {
    await rename(output, binary);
    report.optimized = true;
    report.raw = optimizedRaw;
  } else await rm(output);
  console.log(`WASM: ${raw} → ${report.raw} bytes with optional wasm-opt -Oz`);
} else console.log(`WASM: ${raw} bytes; optional wasm-opt unavailable/disabled`);
await writeFile(path.join(root, ".scratch/wasm-build.json"), JSON.stringify(report, null, 2) + "\n");
