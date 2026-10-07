import "./wasm_exec.js";

export async function loadEngine() {
  const go = new globalThis.Go();
  const response = await fetch(new URL("./cube.wasm", import.meta.url));
  if (!response.ok) throw new Error("Cube engine missing. Build the site with make web.");
  const { instance } = await WebAssembly.instantiate(await response.arrayBuffer(), go.importObject);
  go.run(instance).catch(error => console.error("Cube engine stopped", error));
  if (!globalThis.cubeAPI) throw new Error("Cube engine did not initialize.");
  return request => {
    const result = JSON.parse(globalThis.cubeAPI(JSON.stringify(request)));
    if (!result.ok) throw new Error(result.error);
    return result.data;
  };
}
