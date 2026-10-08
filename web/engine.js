import "./wasm_exec.js";
import { solverAssets } from "./solver-assets.js";

export async function loadEngine({ module, worker = false } = {}) {
  const go = new globalThis.Go();
  if (worker) {
    // The constructive NxN setup allocates many short-lived buffers. Give its
    // worker heap more headroom, with a soft limit for memory on phones.
    go.env.GOGC = "200";
    go.env.GOMEMLIMIT = "128MiB";
  }
  let instance;
  if (!module) {
    const response = await fetch(new URL("./cube.wasm", import.meta.url));
    if (!response.ok) throw new Error("Cube engine missing. Build the site with make web.");
    ({ instance, module } = await WebAssembly.instantiate(await response.arrayBuffer(), go.importObject));
  } else instance = await WebAssembly.instantiate(module, go.importObject);
  go.run(instance).catch(error => console.error("Cube engine stopped", error));
  if (!globalThis.cubeAPI) throw new Error("Cube engine did not initialize.");
  const call = request => {
    const result = JSON.parse(globalThis.cubeAPI(JSON.stringify(request)));
    if (!result.ok) throw new Error(result.error);
    return result.data;
  };
  // Structured cloning shares compiled code, never the Go runtime or memory.
  call.module = module;
  const assets = new Map();
  call.prepare = async request => {
    if (!["solve", "find"].includes(request.op)) return;
    const state = call({ op: "state", size: request.size, cfen: request.cfen, moves: request.moves }).state;
    if (request.op === "solve" && state.solved) return;
    const size = state.size;
    const coordinates = request.op === "find" || size !== 3 || !request.method || ["auto", "kociemba"].includes(request.method);
    const keys = [...(coordinates ? ["coordinates"] : []), ...(size > 3 ? [`nxn-${size}`] : [])];
    await Promise.all(keys.map(key => {
      if (!assets.has(key)) {
        const asset = solverAssets[key];
        const loading = (async () => {
          let response;
          try { response = await fetch(asset.url, { integrity: asset.integrity }); }
          catch { throw new Error(`Solver data could not be verified (${key}). Please reload and try again.`); }
          if (!response.ok) throw new Error(`Solver data missing (${key}). Please reload and try again.`);
          const bytes = new Uint8Array(await response.arrayBuffer());
          if (bytes.length !== asset.bytes) throw new Error(`Solver data length mismatch (${key}).`);
          const error = globalThis.cubeLoadSolverAsset(key, bytes, asset.digest);
          if (error) throw new Error(error);
        })();
        assets.set(key, loading);
        loading.catch(() => assets.delete(key));
      }
      return assets.get(key);
    }));
  };
  return call;
}
