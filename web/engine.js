import "./wasm_exec.js";

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
  return call;
}
